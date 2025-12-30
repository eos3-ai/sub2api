package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPaymentCallbackRoutes registers payment provider callbacks (no auth).
func RegisterPaymentCallbackRoutes(r *gin.Engine, h *handler.Handlers) {
	if r == nil || h == nil || h.Payment == nil {
		return
	}

	// ZPay callbacks
	r.Any("/api/payment/webhook/zpay", h.Payment.ZpayNotify)
	r.Any("/payment/zpay/notify", h.Payment.ZpayNotify)
	r.GET("/payment/zpay/return", h.Payment.PaymentReturn)

	// Stripe callbacks
	r.POST("/api/payment/webhook/stripe", h.Payment.StripeWebhook)
	r.POST("/payment/stripe/webhook", h.Payment.StripeWebhook)
	// Lightweight return endpoints (optional, for external redirects).
	r.GET("/payment/return/stripe", h.Payment.PaymentReturn)
	r.GET("/payment/success", h.Payment.PaymentReturn)
	r.GET("/payment/cancel", h.Payment.PaymentReturn)
}

