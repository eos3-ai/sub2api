package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type checkoutMethodLimit struct {
	DailyLimit     float64 `json:"daily_limit"`
	DailyUsed      float64 `json:"daily_used"`
	DailyRemaining float64 `json:"daily_remaining"`
	SingleMin      float64 `json:"single_min"`
	SingleMax      float64 `json:"single_max"`
	FeeRate        float64 `json:"fee_rate"`
	Available      bool    `json:"available"`
}

type checkoutInfoResponse struct {
	Methods                   map[string]checkoutMethodLimit `json:"methods"`
	GlobalMin                 float64                        `json:"global_min"`
	GlobalMax                 float64                        `json:"global_max"`
	Plans                     []checkoutPlan                 `json:"plans"`
	BalanceDisabled           bool                           `json:"balance_disabled"`
	BalanceRechargeMultiplier float64                        `json:"balance_recharge_multiplier"`
	RechargeFeeRate           float64                        `json:"recharge_fee_rate"`
	HelpText                  string                         `json:"help_text"`
	HelpImageURL              string                         `json:"help_image_url"`
	StripePublishableKey      string                         `json:"stripe_publishable_key"`
}

type checkoutPlan struct {
	ID                   int64    `json:"id"`
	GroupID              int64    `json:"group_id"`
	GroupPlatform        string   `json:"group_platform"`
	GroupName            string   `json:"group_name"`
	RateMultiplier       float64  `json:"rate_multiplier"`
	DailyLimitUSD        *float64 `json:"daily_limit_usd"`
	WeeklyLimitUSD       *float64 `json:"weekly_limit_usd"`
	MonthlyLimitUSD      *float64 `json:"monthly_limit_usd"`
	SupportedModelScopes []string `json:"supported_model_scopes"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Price                float64  `json:"price"`
	OriginalPrice        *float64 `json:"original_price,omitempty"`
	ValidityDays         int      `json:"validity_days"`
	ValidityUnit         string   `json:"validity_unit"`
	Features             []string `json:"features"`
	ForSale              bool     `json:"for_sale"`
	SortOrder            int      `json:"sort_order"`
}

type createOrderV119Request struct {
	Amount            float64 `json:"amount"`
	PaymentType       string  `json:"payment_type"`
	OpenID            string  `json:"openid"`
	WechatResumeToken string  `json:"wechat_resume_token"`
	ReturnURL         string  `json:"return_url"`
	PaymentSource     string  `json:"payment_source"`
	OrderType         string  `json:"order_type"`
	PlanID            int64   `json:"plan_id"`
	IsMobile          *bool   `json:"is_mobile,omitempty"`
}

type refundRequestBody struct {
	Reason string `json:"reason"`
}

type verifyOrderRequest struct {
	OutTradeNo string `json:"out_trade_no" binding:"required"`
}

type resolveOrderByResumeTokenRequest struct {
	ResumeToken string `json:"resume_token" binding:"required"`
}

type publicOrderResult struct {
	ID                  int64      `json:"id"`
	OutTradeNo          string     `json:"out_trade_no"`
	Amount              float64    `json:"amount"`
	PayAmount           float64    `json:"pay_amount"`
	FeeRate             float64    `json:"fee_rate"`
	PaymentType         string     `json:"payment_type"`
	OrderType           string     `json:"order_type"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	ExpiresAt           time.Time  `json:"expires_at"`
	PaidAt              *time.Time `json:"paid_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	RefundAmount        float64    `json:"refund_amount"`
	RefundReason        *string    `json:"refund_reason,omitempty"`
	RefundRequestedAt   *time.Time `json:"refund_requested_at,omitempty"`
	RefundRequestedBy   *string    `json:"refund_requested_by,omitempty"`
	RefundRequestReason *string    `json:"refund_request_reason,omitempty"`
	PlanID              *int64     `json:"plan_id,omitempty"`
}

func (r createOrderV119Request) looksLikeCompatRequest() bool {
	return strings.TrimSpace(r.PaymentType) != "" ||
		strings.TrimSpace(r.OrderType) != "" ||
		strings.TrimSpace(r.WechatResumeToken) != "" ||
		strings.TrimSpace(r.ReturnURL) != "" ||
		strings.TrimSpace(r.PaymentSource) != "" ||
		r.PlanID > 0
}

// GetCheckoutInfo returns all payment-page data in a single request.
// GET /api/v1/payment/checkout-info
func (h *PaymentHandler) GetCheckoutInfo(c *gin.Context) {
	configService := h.paymentService.PaymentConfigService()
	if configService == nil {
		response.Error(c, http.StatusServiceUnavailable, "payment config service is unavailable")
		return
	}

	ctx := c.Request.Context()
	limitsResp, err := configService.GetAvailableMethodLimits(ctx)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	cfg, err := configService.GetPaymentConfig(ctx)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	plans, err := configService.ListPlansForSale(ctx)
	if err != nil {
		plans = nil
	}
	groupInfo := configService.GetGroupInfoMap(ctx, plans)
	planList := make([]checkoutPlan, 0, len(plans))
	for _, plan := range plans {
		info := groupInfo[plan.GroupID]
		planList = append(planList, checkoutPlan{
			ID:                   int64(plan.ID),
			GroupID:              plan.GroupID,
			GroupPlatform:        info.Platform,
			GroupName:            info.Name,
			RateMultiplier:       info.RateMultiplier,
			DailyLimitUSD:        info.DailyLimitUSD,
			WeeklyLimitUSD:       info.WeeklyLimitUSD,
			MonthlyLimitUSD:      info.MonthlyLimitUSD,
			SupportedModelScopes: append([]string(nil), info.ModelScopes...),
			Name:                 plan.Name,
			Description:          plan.Description,
			Price:                plan.Price,
			OriginalPrice:        plan.OriginalPrice,
			ValidityDays:         plan.ValidityDays,
			ValidityUnit:         plan.ValidityUnit,
			Features:             parseCheckoutPlanFeatures(plan.Features),
			ForSale:              plan.ForSale,
			SortOrder:            plan.SortOrder,
		})
	}

	methods := make(map[string]checkoutMethodLimit, len(limitsResp.Methods))
	for key, limit := range limitsResp.Methods {
		dailyRemaining := limit.DailyLimit
		if dailyRemaining < 0 {
			dailyRemaining = 0
		}
		methods[key] = checkoutMethodLimit{
			DailyLimit:     limit.DailyLimit,
			DailyUsed:      0,
			DailyRemaining: dailyRemaining,
			SingleMin:      limit.SingleMin,
			SingleMax:      limit.SingleMax,
			FeeRate:        cfg.RechargeFeeRate,
			Available:      true,
		}
	}

	response.Success(c, checkoutInfoResponse{
		Methods:                   methods,
		GlobalMin:                 limitsResp.GlobalMin,
		GlobalMax:                 limitsResp.GlobalMax,
		Plans:                     planList,
		BalanceDisabled:           cfg.BalanceDisabled,
		BalanceRechargeMultiplier: cfg.BalanceRechargeMultiplier,
		RechargeFeeRate:           cfg.RechargeFeeRate,
		HelpText:                  cfg.HelpText,
		HelpImageURL:              cfg.HelpImageURL,
		StripePublishableKey:      cfg.StripePublishableKey,
	})
}

// GetMyOrdersCompat returns the authenticated user's orders in the v0.1.119 contract.
// GET /api/v1/payment/orders/my
func (h *PaymentHandler) GetMyOrdersCompat(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}

	page, pageSize := response.ParsePagination(c)
	orders, total, err := h.paymentService.GetUserOrders(c.Request.Context(), subject.UserID, service.OrderListParams{
		Page:        page,
		PageSize:    pageSize,
		Status:      c.Query("status"),
		OrderType:   c.Query("order_type"),
		PaymentType: c.Query("payment_type"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, sanitizePaymentOrdersForResponse(orders), int64(total), page, pageSize)
}

// CancelOrder cancels a pending order for the authenticated user.
// POST /api/v1/payment/orders/:id/cancel
func (h *PaymentHandler) CancelOrder(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}

	orderID, err := strconv.ParseInt(c.Param("orderNo"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	msg, err := h.paymentService.CancelOrder(c.Request.Context(), orderID, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": msg})
}

// RequestRefund creates a user refund request.
// POST /api/v1/payment/orders/:id/refund
func (h *PaymentHandler) RequestRefund(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}

	orderID, err := strconv.ParseInt(c.Param("orderNo"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var req refundRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.paymentService.RequestRefund(c.Request.Context(), orderID, subject.UserID, req.Reason); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "refund requested"})
}

// GetRefundEligibleProviders returns provider instance IDs that allow user refund.
// GET /api/v1/payment/providers/refund-eligible
func (h *PaymentHandler) GetRefundEligibleProviders(c *gin.Context) {
	configService := h.paymentService.PaymentConfigService()
	if configService == nil {
		response.Error(c, http.StatusServiceUnavailable, "payment config service is unavailable")
		return
	}

	ids, err := configService.GetUserRefundEligibleInstanceIDs(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"provider_instance_ids": ids})
}

// VerifyOrderPublic keeps the anonymous public out_trade_no lookup for compatibility.
// POST /api/v1/payment/public/orders/verify
func (h *PaymentHandler) VerifyOrderPublic(c *gin.Context) {
	var req verifyOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	order, err := h.paymentService.VerifyOrderPublic(c.Request.Context(), req.OutTradeNo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, buildPublicOrderResult(order))
}

// ResolveOrderPublicByResumeToken resolves a public order from a signed resume token.
// POST /api/v1/payment/public/orders/resolve
func (h *PaymentHandler) ResolveOrderPublicByResumeToken(c *gin.Context) {
	var req resolveOrderByResumeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	order, err := h.paymentService.GetPublicOrderByResumeToken(c.Request.Context(), req.ResumeToken)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, buildPublicOrderResult(order))
}

func (h *PaymentHandler) handleCreateOrderV119(c *gin.Context, req createOrderV119Request) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}

	if strings.TrimSpace(req.WechatResumeToken) != "" {
		claims, err := h.paymentService.ParseWeChatPaymentResumeToken(req.WechatResumeToken)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if err := applyWeChatPaymentResumeClaims(&req, claims); err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}

	mobile := isMobile(c)
	if req.IsMobile != nil {
		mobile = *req.IsMobile
	}

	result, err := h.paymentService.CreateOrderV119(c.Request.Context(), service.CreateOrderRequest{
		UserID:          subject.UserID,
		Amount:          req.Amount,
		PaymentType:     req.PaymentType,
		OpenID:          req.OpenID,
		ClientIP:        c.ClientIP(),
		IsMobile:        mobile,
		IsWeChatBrowser: isWeChatBrowser(c),
		SrcHost:         c.Request.Host,
		SrcURL:          c.Request.Referer(),
		ReturnURL:       req.ReturnURL,
		PaymentSource:   req.PaymentSource,
		OrderType:       req.OrderType,
		PlanID:          req.PlanID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}

func applyWeChatPaymentResumeClaims(req *createOrderV119Request, claims *service.WeChatPaymentResumeClaims) error {
	if req == nil || claims == nil {
		return infraerrors.BadRequest("INVALID_WECHAT_PAYMENT_RESUME_TOKEN", "wechat payment resume context is missing")
	}

	openID := strings.TrimSpace(claims.OpenID)
	if openID == "" {
		return infraerrors.BadRequest("INVALID_WECHAT_PAYMENT_RESUME_TOKEN", "wechat payment resume token missing openid")
	}

	paymentType := service.NormalizeVisibleMethod(claims.PaymentType)
	if paymentType == "" {
		paymentType = payment.TypeWxpay
	}
	if req.PaymentType != "" {
		requestPaymentType := service.NormalizeVisibleMethod(req.PaymentType)
		if requestPaymentType != "" && requestPaymentType != paymentType {
			return infraerrors.BadRequest("INVALID_WECHAT_PAYMENT_RESUME_TOKEN", "wechat payment resume token payment type mismatch")
		}
	}

	req.PaymentType = paymentType
	req.OpenID = openID

	if strings.TrimSpace(claims.Amount) != "" {
		amount, err := strconv.ParseFloat(strings.TrimSpace(claims.Amount), 64)
		if err != nil || amount <= 0 {
			return infraerrors.BadRequest("INVALID_WECHAT_PAYMENT_RESUME_TOKEN", "invalid resume amount")
		}
		req.Amount = amount
	}
	if claims.OrderType != "" {
		req.OrderType = claims.OrderType
	}
	if claims.PlanID > 0 {
		req.PlanID = claims.PlanID
	}
	return nil
}

func buildPublicOrderResult(order *dbent.PaymentOrder) publicOrderResult {
	return publicOrderResult{
		ID:                  order.ID,
		OutTradeNo:          order.OutTradeNo,
		Amount:              order.Amount,
		PayAmount:           order.PayAmount,
		FeeRate:             order.FeeRate,
		PaymentType:         order.PaymentType,
		OrderType:           order.OrderType,
		Status:              order.Status,
		CreatedAt:           order.CreatedAt,
		ExpiresAt:           order.ExpiresAt,
		PaidAt:              order.PaidAt,
		CompletedAt:         order.CompletedAt,
		RefundAmount:        order.RefundAmount,
		RefundReason:        order.RefundReason,
		RefundRequestedAt:   order.RefundRequestedAt,
		RefundRequestedBy:   order.RefundRequestedBy,
		RefundRequestReason: order.RefundRequestReason,
		PlanID:              order.PlanID,
	}
}

func requireAuth(c *gin.Context) (middleware2.AuthSubject, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return middleware2.AuthSubject{}, false
	}
	return subject, true
}

func isMobile(c *gin.Context) bool {
	userAgent := strings.ToLower(c.GetHeader("User-Agent"))
	for _, keyword := range []string{"mobile", "android", "iphone", "ipad", "ipod"} {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}
	return false
}

func isWeChatBrowser(c *gin.Context) bool {
	return strings.Contains(strings.ToLower(c.GetHeader("User-Agent")), "micromessenger")
}

func sanitizePaymentOrdersForResponse(orders []*dbent.PaymentOrder) []*dbent.PaymentOrder {
	if len(orders) == 0 {
		return orders
	}

	out := make([]*dbent.PaymentOrder, 0, len(orders))
	for _, order := range orders {
		out = append(out, sanitizePaymentOrderForResponse(order))
	}
	return out
}

func sanitizePaymentOrderForResponse(order *dbent.PaymentOrder) *dbent.PaymentOrder {
	if order == nil {
		return nil
	}

	cloned := *order
	cloned.ProviderSnapshot = nil
	return &cloned
}

func parseCheckoutPlanFeatures(raw string) []string {
	if raw == "" {
		return []string{}
	}

	out := make([]string, 0)
	for _, line := range strings.Split(raw, "\n") {
		item := strings.TrimSpace(line)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
