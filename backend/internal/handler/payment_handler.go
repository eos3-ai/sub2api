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
	PlanID    string   `json:"plan_id"`
	AmountUSD *float64 `json:"amount_usd,omitempty"`
	Channel   string   `json:"channel" binding:"required,oneof=zpay stripe alipay wechat"`
}

// GetPlans returns configured payment packages as "plans".
// GET /api/v1/payment/plans
func (h *PaymentHandler) GetPlans(c *gin.Context) {
	if h.cfg == nil {
		response.Success(c, []dto.PaymentPlan{})
		return
	}

	paymentCfg := h.cfg.Payment
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
			ID:         planID,
			Name:       pkg.Label,
			AmountUSD:  amountUSD,
			PayUSD:     payUSD,
			CreditsUSD: creditsUSD,
			ExchangeRate: paymentCfg.ExchangeRate,
			DiscountRate: discount,
			Enabled:    paymentCfg.Enabled,
		})
	}
	response.Success(c, plans)
}

// CreateOrder creates an order record using existing PaymentService.
// POST /api/v1/payment/orders
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req createPaymentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if h.cfg == nil {
		response.Error(c, http.StatusBadRequest, "payment config is missing")
		return
	}

	var amountCNY float64
	if req.PlanID != "" {
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
	} else {
		response.BadRequest(c, "either plan_id or amount_usd is required")
		return
	}

	provider := normalizePaymentProvider(req.Channel)
	// Business rule: WeChat Pay (Stripe) minimum payable is ¥100.
	// Frontend also enforces it, but backend must validate to prevent bypass.
	if provider == "stripe" && amountCNY > 0 && amountCNY < 100 {
		response.BadRequest(c, "wechat pay minimum amount is 100 CNY")
		return
	}

	order, err := h.paymentService.CreateOrder(c.Request.Context(), &service.CreatePaymentOrderRequest{
		UserID:        subject.UserID,
		Username:      "",
		AmountCNY:     amountCNY,
		Provider:      provider,
		PaymentMethod: "web",
		ClientIP:      c.ClientIP(),
		UserAgent:     c.GetHeader("User-Agent"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var payURL string
	var qrURL string
	switch strings.ToLower(strings.TrimSpace(order.Provider)) {
	case "zpay":
		payURL, qrURL, err = h.zpayService.CreatePayment(c.Request.Context(), order, req.Channel)
	case "stripe":
		payURL, qrURL, err = h.stripeService.CreatePayment(c.Request.Context(), order, req.Channel)
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
		"order": dto.PaymentOrderFromService(order),
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

func normalizePaymentProvider(channel string) string {
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case "alipay":
		return "zpay"
	case "wechat":
		return "stripe"
	default:
		return strings.ToLower(strings.TrimSpace(channel))
	}
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

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	signature := c.GetHeader("Stripe-Signature")
	info, err := h.stripeService.VerifyWebhook(c.Request.Context(), payload, signature)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if info == nil || strings.TrimSpace(info.OrderNo) == "" {
		c.Status(http.StatusOK)
		return
	}

	order, _ := h.paymentService.GetOrderByOrderNo(c.Request.Context(), info.OrderNo)
	if order == nil || !strings.EqualFold(order.Provider, "stripe") {
		c.Status(http.StatusOK)
		return
	}
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

	switch info.EventType {
	case "payment_intent.succeeded":
		_, err := h.paymentService.MarkOrderPaid(c.Request.Context(), info.OrderNo, info.TradeNo, gin.H{
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
	case "payment_intent.payment_failed":
		reason := info.EventType
		if strings.TrimSpace(info.FailureMessage) != "" {
			reason = reason + ": " + strings.TrimSpace(info.FailureMessage)
		}
		_, err := h.paymentService.MarkOrderFailed(c.Request.Context(), info.OrderNo, reason)
		if err != nil {
			log.Printf("[Stripe Webhook] Failed to mark order as failed: order_no=%s, reason=%s, error=%v",
				info.OrderNo, reason, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process payment failure",
			})
			return
		}
	case "payment_intent.canceled":
		reason := info.EventType
		if strings.TrimSpace(info.FailureMessage) != "" {
			reason = reason + ": " + strings.TrimSpace(info.FailureMessage)
		}
		_, err := h.paymentService.MarkOrderCancelled(c.Request.Context(), info.OrderNo, reason)
		if err != nil {
			log.Printf("[Stripe Webhook] Failed to mark order as cancelled: order_no=%s, reason=%s, error=%v",
				info.OrderNo, reason, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process payment cancellation",
			})
			return
		}
	default:
	}
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
