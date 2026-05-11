package handler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type PaymentHandler struct {
	cfg            *config.Config
	paymentService *service.PaymentService
	zpayService    *service.ZpayService
	stripeService  *service.StripeService
}

func NewPaymentHandler(
	cfg *config.Config,
	paymentService *service.PaymentService,
	zpayService *service.ZpayService,
	stripeService *service.StripeService,
) *PaymentHandler {
	return &PaymentHandler{
		cfg:            cfg,
		paymentService: paymentService,
		zpayService:    zpayService,
		stripeService:  stripeService,
	}
}

type createPaymentOrderRequest struct {
	PlanID              string   `json:"plan_id"`
	AmountUSD           *float64 `json:"amount_usd,omitempty"`
	SubscriptionGroupID *int64   `json:"subscription_group_id,omitempty"`
	Channel             string   `json:"channel" binding:"required"`
}

// GetPlans returns configured payment packages as "plans".
// GET /api/v1/payment/plans
func (h *PaymentHandler) GetPlans(c *gin.Context) {
	if h.cfg == nil {
		response.Success(c, []dto.PaymentPlan{})
		return
	}

	paymentCfg := h.cfg.Payment
	availableOptions := h.availableUserPaymentChannelOptions()
	availableChannels := availableChannelsFromOptions(availableOptions)
	discount := normalizedDiscountRate(paymentCfg.DiscountRate)
	plans := make([]dto.PaymentPlan, 0, len(paymentCfg.Packages))
	for i, pkg := range paymentCfg.Packages {
		planID := fmt.Sprintf("pkg_%d", i)
		amountUSD := pkg.AmountUSD
		if amountUSD <= 0 && paymentCfg.ExchangeRate > 0 && pkg.AmountCNY > 0 {
			amountUSD = pkg.AmountCNY / paymentCfg.ExchangeRate
		}
		creditsUSD := amountUSD
		payUSD := creditsUSD * discount
		plans = append(plans, dto.PaymentPlan{
			ID:                planID,
			Name:              pkg.Label,
			AmountUSD:         amountUSD,
			PayUSD:            payUSD,
			CreditsUSD:        creditsUSD,
			ExchangeRate:      paymentCfg.ExchangeRate,
			DiscountRate:      discount,
			Enabled:           paymentCfg.Enabled,
			AvailableChannels: append([]string(nil), availableChannels...),
			AvailableOptions:  append([]dto.PaymentChannelOption(nil), availableOptions...),
		})
	}
	response.Success(c, plans)
}

// GetSubscriptionPlans returns user-visible subscription purchase packages.
// GET /api/v1/payment/subscription-plans
func (h *PaymentHandler) GetSubscriptionPlans(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if h.cfg == nil {
		response.Success(c, []dto.PaymentSubscriptionPlan{})
		return
	}

	plans, err := h.paymentService.GetSubscriptionPlans(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// Subscription billing should be "displayed USD amount = paid CNY amount".
	// Keep exchange rate as 1 for subscription plan presentation.
	exchangeRate := 1.0
	availableOptions := h.availableUserPaymentChannelOptions()
	availableChannels := availableChannelsFromOptions(availableOptions)
	out := make([]dto.PaymentSubscriptionPlan, 0, len(plans))
	for i := range plans {
		item := plans[i]
		out = append(out, dto.PaymentSubscriptionPlan{
			GroupID:               item.GroupID,
			GroupName:             item.GroupName,
			Description:           item.Description,
			Platform:              item.Platform,
			PriceUSD:              item.PriceUSD,
			ValidityDays:          item.ValidityDays,
			DailyLimitUSD:         item.DailyLimitUSD,
			WeeklyLimitUSD:        item.WeeklyLimitUSD,
			MonthlyLimitUSD:       item.MonthlyLimitUSD,
			ExchangeRate:          exchangeRate,
			AvailableChannels:     append([]string(nil), availableChannels...),
			AvailableOptions:      append([]dto.PaymentChannelOption(nil), availableOptions...),
			HasActiveSubscription: item.HasActiveSubscription,
		})
	}

	response.Success(c, out)
}

// CreateOrder creates an order record using existing PaymentService.
// POST /api/v1/payment/orders
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	var compatReq createOrderV119Request
	if err := c.ShouldBindBodyWith(&compatReq, binding.JSON); err == nil && compatReq.looksLikeCompatRequest() {
		h.handleCreateOrderV119(c, compatReq)
		return
	}

	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req createPaymentOrderRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if h.cfg == nil {
		response.Error(c, http.StatusBadRequest, "payment config is missing")
		return
	}
	channelSelection, ok := h.resolveCreateChannel(req.Channel)
	if !ok {
		response.BadRequest(c, "payment channel is disabled")
		return
	}

	provider := channelSelection.Provider
	channel := channelSelection.Method

	var amountCNY float64
	bizType := service.PaymentBizTypeOnlineRecharge
	var bizGroupID *int64
	var bizValidityDays *int
	selectionCount := 0
	if req.PlanID != "" {
		selectionCount++
	}
	if req.AmountUSD != nil && *req.AmountUSD > 0 {
		selectionCount++
	}
	if req.SubscriptionGroupID != nil {
		selectionCount++
	}
	if selectionCount == 0 {
		response.BadRequest(c, "one of plan_id, amount_usd or subscription_group_id is required")
		return
	}
	if selectionCount > 1 {
		response.BadRequest(c, "plan_id, amount_usd and subscription_group_id are mutually exclusive")
		return
	}

	if req.SubscriptionGroupID != nil {
		if *req.SubscriptionGroupID <= 0 {
			response.BadRequest(c, "subscription_group_id must be positive")
			return
		}
		plan, err := h.paymentService.GetSubscriptionPlanByGroupID(c.Request.Context(), subject.UserID, *req.SubscriptionGroupID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		// 套餐计费采用 1:1：显示多少 USD，实付多少 CNY（数值一致）。
		amountCNY = plan.PriceUSD
		bizType = service.PaymentBizTypeSubscriptionPurchase
		groupID := plan.GroupID
		bizGroupID = &groupID
		validityDays := plan.ValidityDays
		bizValidityDays = &validityDays
	} else if req.PlanID != "" {
		v, err := h.amountCNYFromPlanID(req.PlanID)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		amountCNY = v
	} else if req.AmountUSD != nil && *req.AmountUSD > 0 {
		discount := normalizedDiscountRate(h.cfg.Payment.DiscountRate)
		payUSD := (*req.AmountUSD) * discount
		amountCNY = payUSD * h.cfg.Payment.ExchangeRate
	}

	order, err := h.paymentService.CreateOrder(c.Request.Context(), &service.CreatePaymentOrderRequest{
		UserID:          subject.UserID,
		Username:        "",
		AmountCNY:       amountCNY,
		BizType:         bizType,
		BizGroupID:      bizGroupID,
		BizValidityDays: bizValidityDays,
		Provider:        provider,
		Channel:         channel,
		PaymentMethod:   "web",
		ClientIP:        c.ClientIP(),
		UserAgent:       c.GetHeader("User-Agent"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var payURL string
	var qrURL string
	switch strings.ToLower(strings.TrimSpace(order.Provider)) {
	case "zpay":
		payURL, qrURL, err = h.zpayService.CreatePayment(c.Request.Context(), order, channel)
	case "stripe":
		payURL, qrURL, err = h.stripeService.CreatePayment(c.Request.Context(), order, channel)
	default:
		// Provider integration is still WIP.
	}
	if err != nil {
		_, _ = h.paymentService.MarkOrderFailed(c.Request.Context(), order.OrderNo, err.Error())
		// Most errors here are configuration issues (disabled/missing keys/base_url).
		// Return 400 so frontend can surface a clear actionable message.
		response.BadRequest(c, err.Error())
		return
	}
	if payURL != "" {
		order.PaymentURL = payURL
		_ = h.paymentService.UpdateOrder(c.Request.Context(), order)
	}

	response.Created(c, gin.H{
		"order":   dto.PaymentOrderFromService(order),
		"pay_url": payURL,
		"qr_url":  qrURL,
	})
}

// ListMyOrders lists current user's payment orders.
// GET /api/v1/payment/orders?page=1&page_size=20
func (h *PaymentHandler) ListMyOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	status := c.Query("status")

	orders, result, err := h.paymentService.ListUserOrders(c.Request.Context(), subject.UserID, params, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.PaymentOrder, 0, len(orders))
	for i := range orders {
		out = append(out, *dto.PaymentOrderFromService(&orders[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// GetMyOrder returns current user's order by order_no.
// GET /api/v1/payment/orders/:orderNo
func (h *PaymentHandler) GetMyOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	orderNo := strings.TrimSpace(c.Param("orderNo"))
	if orderNo == "" {
		response.BadRequest(c, "order_no is required")
		return
	}

	if orderID, err := strconv.ParseInt(orderNo, 10, 64); err == nil && orderID > 0 {
		order, err := h.paymentService.GetOrder(c.Request.Context(), orderID, subject.UserID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, sanitizePaymentOrderForResponse(order))
		return
	}

	order, err := h.paymentService.GetOrderByOrderNo(c.Request.Context(), orderNo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if order == nil || order.UserID != subject.UserID {
		response.NotFound(c, "order not found")
		return
	}

	response.Success(c, dto.PaymentOrderFromService(order))
}

func (h *PaymentHandler) amountCNYFromPlanID(planID string) (float64, error) {
	if h.cfg == nil {
		return 0, errors.New("payment config is missing")
	}
	if !strings.HasPrefix(planID, "pkg_") {
		return 0, errors.New("invalid plan_id")
	}
	indexStr := strings.TrimPrefix(planID, "pkg_")
	i, err := strconv.Atoi(indexStr)
	if err != nil {
		return 0, errors.New("invalid plan_id")
	}
	if i < 0 || i >= len(h.cfg.Payment.Packages) {
		return 0, errors.New("plan_id not found")
	}
	pkg := h.cfg.Payment.Packages[i]
	if pkg.AmountCNY > 0 {
		return pkg.AmountCNY, nil
	}
	if pkg.AmountUSD > 0 && h.cfg.Payment.ExchangeRate > 0 {
		discount := normalizedDiscountRate(h.cfg.Payment.DiscountRate)
		payUSD := pkg.AmountUSD * discount
		return payUSD * h.cfg.Payment.ExchangeRate, nil
	}
	return 0, errors.New("invalid package config: amount_cny or amount_usd is required")
}

func normalizedDiscountRate(discountRate float64) float64 {
	// discountRate is a payable multiplier in (0,1], e.g. 0.15 means "pay 15%".
	// Compatibility: historical default was 1.0 (pay full).
	if discountRate <= 0 {
		return 1.0
	}
	if discountRate > 1 {
		return 1.0
	}
	return discountRate
}

type createPaymentChannelSelection struct {
	Provider string
	Method   string
}

func (h *PaymentHandler) availableUserPaymentChannelOptions() []dto.PaymentChannelOption {
	if h == nil || h.cfg == nil {
		return nil
	}
	paymentCfg := h.cfg.Payment
	if !paymentCfg.Enabled {
		return nil
	}

	out := make([]dto.PaymentChannelOption, 0, 4)
	if paymentCfg.Zpay.Enabled {
		out = appendPaymentChannelOptions(out, "zpay", paymentCfg.Zpay.PaymentMethods, []string{"alipay", "wechat"})
	}
	if paymentCfg.Stripe.Enabled {
		out = appendPaymentChannelOptions(out, "stripe", paymentCfg.Stripe.PaymentMethods, []string{"wechat"})
	}
	return out
}

func appendPaymentChannelOptions(out []dto.PaymentChannelOption, provider string, rawMethods string, fallbackMethods []string) []dto.PaymentChannelOption {
	hasExplicitMethods := strings.TrimSpace(rawMethods) != ""
	enabled := normalizePaymentMethodSet(rawMethods)
	if len(enabled) == 0 && !hasExplicitMethods {
		enabled = make(map[string]struct{}, len(fallbackMethods))
		for _, method := range fallbackMethods {
			if normalized := normalizeUserPaymentMethod(method); normalized != "" {
				enabled[normalized] = struct{}{}
			}
		}
	}

	for _, method := range []string{"alipay", "wechat"} {
		if _, ok := enabled[method]; !ok {
			continue
		}
		out = append(out, dto.PaymentChannelOption{
			Provider: provider,
			Method:   method,
			Channel:  provider + "_" + method,
		})
	}
	return out
}

func availableChannelsFromOptions(options []dto.PaymentChannelOption) []string {
	enabled := make(map[string]struct{}, 2)
	for _, option := range options {
		if option.Method == "alipay" || option.Method == "wechat" {
			enabled[option.Method] = struct{}{}
		}
	}

	out := make([]string, 0, 2)
	for _, method := range []string{"alipay", "wechat"} {
		if _, ok := enabled[method]; ok {
			out = append(out, method)
		}
	}
	return out
}

func normalizePaymentMethodSet(rawMethods string) map[string]struct{} {
	methods := parseCommaListLower(rawMethods)
	if len(methods) == 0 {
		return nil
	}

	out := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		if normalized := normalizeUserPaymentMethod(method); normalized != "" {
			out[normalized] = struct{}{}
		}
	}
	return out
}

func normalizeUserPaymentMethod(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "alipay", "zpay":
		return "alipay"
	case "wechat", "wxpay", "wechat_pay":
		return "wechat"
	default:
		return ""
	}
}

func (h *PaymentHandler) isCreateChannelEnabled(channel string) bool {
	_, ok := h.resolveCreateChannel(channel)
	return ok
}

func (h *PaymentHandler) resolveCreateChannel(channel string) (createPaymentChannelSelection, bool) {
	normalized := strings.ToLower(strings.TrimSpace(channel))
	if normalized == "" {
		return createPaymentChannelSelection{}, false
	}

	options := h.availableUserPaymentChannelOptions()
	if len(options) == 0 {
		return createPaymentChannelSelection{}, false
	}

	if normalized == "zpay" || normalized == "stripe" {
		for _, option := range options {
			if option.Provider == normalized {
				return createPaymentChannelSelection{Provider: option.Provider, Method: option.Method}, true
			}
		}
		return createPaymentChannelSelection{}, false
	}

	if method := normalizeUserPaymentMethod(normalized); method != "" {
		// Preserve legacy alipay/wechat payloads as ZPay selections when available,
		// then fall back to another enabled provider for old clients reading
		// available_channels from a Stripe-only configuration.
		if selection, ok := findPaymentChannelOption(options, "zpay", method); ok {
			return selection, true
		}
		return findFirstPaymentChannelOptionByMethod(options, method)
	}

	provider, method, ok := splitProviderMethodChannel(normalized)
	if !ok {
		return createPaymentChannelSelection{}, false
	}
	return findPaymentChannelOption(options, provider, method)
}

func findPaymentChannelOption(options []dto.PaymentChannelOption, provider string, method string) (createPaymentChannelSelection, bool) {
	for _, option := range options {
		if option.Provider == provider && option.Method == method {
			return createPaymentChannelSelection{Provider: option.Provider, Method: option.Method}, true
		}
	}
	return createPaymentChannelSelection{}, false
}

func findFirstPaymentChannelOptionByMethod(options []dto.PaymentChannelOption, method string) (createPaymentChannelSelection, bool) {
	for _, option := range options {
		if option.Method == method {
			return createPaymentChannelSelection{Provider: option.Provider, Method: option.Method}, true
		}
	}
	return createPaymentChannelSelection{}, false
}

func splitProviderMethodChannel(channel string) (string, string, bool) {
	parts := strings.FieldsFunc(channel, func(r rune) bool {
		return r == '_' || r == ':' || r == '-'
	})
	if len(parts) != 2 {
		return "", "", false
	}
	provider := strings.ToLower(strings.TrimSpace(parts[0]))
	if provider != "zpay" && provider != "stripe" {
		return "", "", false
	}
	method := normalizeUserPaymentMethod(parts[1])
	if method == "" {
		return "", "", false
	}
	return provider, method, true
}

func parseCommaListLower(raw string) []string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

// ZpayNotify handles ZPay notify callback.
// ZPay expects plain "success" on successful processing.
func (h *PaymentHandler) ZpayNotify(c *gin.Context) {
	if h.cfg == nil || h.paymentService == nil || h.zpayService == nil {
		c.String(http.StatusOK, "fail")
		return
	}
	if !h.cfg.Payment.Zpay.Enabled {
		c.String(http.StatusOK, "fail")
		return
	}

	clientIP := c.ClientIP()
	if !ipAllowed(clientIP, h.cfg.Payment.Zpay.IPWhitelist) {
		c.String(http.StatusOK, "fail")
		return
	}

	data := collectCallbackParams(c)
	orderNo, tradeNo, err := h.zpayService.VerifyCallback(c.Request.Context(), data)
	if err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	order, _ := h.paymentService.GetOrderByOrderNo(c.Request.Context(), orderNo)
	if order == nil || !strings.EqualFold(order.Provider, "zpay") {
		c.String(http.StatusOK, "fail")
		return
	}

	status := strings.ToUpper(strings.TrimSpace(data["trade_status"]))
	if status == "" {
		status = strings.ToUpper(strings.TrimSpace(data["status"]))
	}
	if status != "" && status != "TRADE_SUCCESS" && status != "SUCCESS" {
		c.String(http.StatusOK, "success")
		return
	}

	if moneyStr := strings.TrimSpace(data["money"]); moneyStr != "" {
		if money, err := strconv.ParseFloat(moneyStr, 64); err == nil {
			if !approxEqual(order.AmountCNY, money, 0.02) {
				c.String(http.StatusOK, "fail")
				return
			}
		}
	}

	if tradeNo == "" {
		tradeNo = "zpay:" + orderNo
	}
	if _, err := h.paymentService.MarkOrderPaid(c.Request.Context(), orderNo, tradeNo, data); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

// StripeWebhook handles Stripe webhook callback.
func (h *PaymentHandler) StripeWebhook(c *gin.Context) {
	if h.cfg == nil || h.paymentService == nil || h.stripeService == nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if !h.cfg.Payment.Stripe.Enabled {
		c.Status(http.StatusBadRequest)
		return
	}

	log.Printf("[Stripe Webhook] Request received: remote_addr=%s, content_length=%d",
		c.Request.RemoteAddr, c.Request.ContentLength)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	signature := c.GetHeader("Stripe-Signature")
	log.Printf("[Stripe Webhook] Stripe-Signature present: %v", signature != "")
	info, err := h.stripeService.VerifyWebhook(c.Request.Context(), payload, signature)
	if err != nil {
		log.Printf("[Stripe Webhook] verify failed: error=%v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	if info == nil || strings.TrimSpace(info.OrderNo) == "" {
		if info != nil {
			log.Printf("[Stripe Webhook] missing order_no: event_id=%s, type=%s, trade_no=%s",
				info.EventID, info.EventType, info.TradeNo)
		}
		c.Status(http.StatusOK)
		return
	}

	order, _ := h.paymentService.GetOrderByOrderNo(c.Request.Context(), info.OrderNo)
	if order == nil || !strings.EqualFold(order.Provider, "stripe") {
		log.Printf("[Stripe Webhook] order not found or provider mismatch: event_id=%s, type=%s, order_no=%s",
			info.EventID, info.EventType, info.OrderNo)
		c.Status(http.StatusOK)
		return
	}
	log.Printf("[Stripe Webhook] Order found: order_no=%s, user_id=%d, status=%s, amount_usd=%.2f, provider=%s",
		order.OrderNo, order.UserID, order.Status, order.TotalUSD, order.Provider)

	if info.Amount > 0 {
		money := float64(info.Amount) / 100
		if !approxEqual(order.AmountCNY, money, 0.02) {
			c.Status(http.StatusBadRequest)
			return
		}
	}
	if info.Currency != "" {
		expected := strings.ToLower(strings.TrimSpace(h.cfg.Payment.Stripe.Currency))
		if expected != "" && strings.ToLower(expected) != strings.ToLower(info.Currency) {
			c.Status(http.StatusBadRequest)
			return
		}
	}

	log.Printf("[Stripe Webhook] Processing event type: %s for order_no=%s", info.EventType, info.OrderNo)

	switch info.EventType {
	case "payment_intent.succeeded":
		updatedOrder, err := h.paymentService.MarkOrderPaid(c.Request.Context(), info.OrderNo, info.TradeNo, gin.H{
			"type": info.EventType,
		})
		if err != nil {
			log.Printf("[Stripe Webhook] Failed to mark order as paid: order_no=%s, trade_no=%s, error=%v",
				info.OrderNo, info.TradeNo, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process payment",
			})
			return
		}
		if updatedOrder != nil {
			log.Printf("[Stripe Webhook] Order marked as PAID successfully: order_no=%s, trade_no=%s, user_id=%d, amount_usd=%.2f, paid_at=%v",
				updatedOrder.OrderNo, *updatedOrder.TradeNo, updatedOrder.UserID, updatedOrder.TotalUSD, updatedOrder.PaidAt)
		}
	case "payment_intent.payment_failed":
		reason := info.EventType
		if strings.TrimSpace(info.FailureMessage) != "" {
			reason = reason + ": " + strings.TrimSpace(info.FailureMessage)
		}
		updatedOrder, err := h.paymentService.MarkOrderFailed(c.Request.Context(), info.OrderNo, reason)
		if err != nil {
			log.Printf("[Stripe Webhook] Failed to mark order as failed: order_no=%s, reason=%s, error=%v",
				info.OrderNo, reason, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process payment failure",
			})
			return
		}
		if updatedOrder != nil {
			log.Printf("[Stripe Webhook] Order marked as FAILED successfully: order_no=%s, user_id=%d, reason=%s, callback_at=%v",
				updatedOrder.OrderNo, updatedOrder.UserID, reason, updatedOrder.CallbackAt)
		}
	case "payment_intent.canceled":
		reason := info.EventType
		if strings.TrimSpace(info.FailureMessage) != "" {
			reason = reason + ": " + strings.TrimSpace(info.FailureMessage)
		}
		updatedOrder, err := h.paymentService.MarkOrderCancelled(c.Request.Context(), info.OrderNo, reason)
		if err != nil {
			log.Printf("[Stripe Webhook] Failed to mark order as cancelled: order_no=%s, reason=%s, error=%v",
				info.OrderNo, reason, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process payment cancellation",
			})
			return
		}
		if updatedOrder != nil {
			log.Printf("[Stripe Webhook] Order marked as CANCELLED successfully: order_no=%s, user_id=%d, reason=%s, callback_at=%v",
				updatedOrder.OrderNo, updatedOrder.UserID, reason, updatedOrder.CallbackAt)
		}
	default:
	}
	log.Printf("[Stripe Webhook] Request processing completed successfully: event_id=%s, type=%s, order_no=%s",
		info.EventID, info.EventType, info.OrderNo)
	c.Status(http.StatusOK)
}

// PaymentReturn provides a lightweight return endpoint for payment providers.
func (h *PaymentHandler) PaymentReturn(c *gin.Context) {
	// Some deployments mistakenly configure Stripe webhook URL to point to the return endpoint.
	// If this looks like a Stripe webhook call, handle it as webhook (no redirect).
	if strings.TrimSpace(c.GetHeader("Stripe-Signature")) != "" {
		h.StripeWebhook(c)
		return
	}

	// Return endpoints are for external redirects; keep it lightweight and safe.
	// If order_no is provided by the channel, redirect to the SPA result page to show status.
	orderNo := strings.TrimSpace(c.Query("order"))
	if orderNo == "" {
		orderNo = strings.TrimSpace(c.Query("order_no"))
	}
	if orderNo == "" {
		// ZPay commonly uses out_trade_no.
		orderNo = strings.TrimSpace(c.Query("out_trade_no"))
	}

	status := strings.TrimSpace(c.Query("status"))
	if status == "" {
		// Some providers use trade_status.
		status = strings.TrimSpace(c.Query("trade_status"))
	}
	if status == "" {
		// Fallback based on path for legacy /payment/success|/payment/cancel.
		p := strings.ToLower(strings.TrimSpace(c.Request.URL.Path))
		if strings.Contains(p, "success") {
			status = "success"
		} else if strings.Contains(p, "cancel") {
			status = "cancel"
		}
	}

	if orderNo != "" {
		target := fmt.Sprintf("/payment/result?order=%s", orderNo)
		if status != "" {
			target = target + "&status=" + status
		}
		c.Redirect(http.StatusFound, target)
		return
	}

	c.Redirect(http.StatusFound, "/payment")
}

func collectCallbackParams(c *gin.Context) map[string]string {
	_ = c.Request.ParseForm()
	out := map[string]string{}
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out
}

func approxEqual(a float64, b float64, eps float64) bool {
	if a > b {
		return a-b <= eps
	}
	return b-a <= eps
}

func ipAllowed(clientIP string, whitelist string) bool {
	raw := strings.TrimSpace(whitelist)
	if raw == "" {
		return true
	}
	items := strings.Split(raw, ",")
	for _, item := range items {
		if strings.TrimSpace(item) == clientIP {
			return true
		}
	}
	return false
}
