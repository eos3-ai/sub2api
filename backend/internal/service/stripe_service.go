package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/url"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

// StripeService 封装 Stripe 渠道
type StripeService struct {
	cfg            *config.StripeConfig
	paymentBaseURL string
}

type StripeWebhookInfo struct {
	EventID        string
	OrderNo        string
	TradeNo        string
	EventType      string
	Amount         int64
	Currency       string
	FailureMessage string
}

func NewStripeService(cfg *config.Config) *StripeService {
	var stripeCfg *config.StripeConfig
	var paymentBaseURL string
	if cfg != nil {
		stripeCfg = &cfg.Payment.Stripe
		paymentBaseURL = cfg.Payment.BaseURL
	}
	return &StripeService{cfg: stripeCfg, paymentBaseURL: paymentBaseURL}
}

// CreatePayment creates a Stripe PaymentIntent for the selected method and returns pay_url/qr_url.
func (s *StripeService) CreatePayment(ctx context.Context, order *PaymentOrder, channel string) (payURL string, qrURL string, err error) {
	if s.cfg == nil || !s.cfg.Enabled {
		return "", "", errors.New("stripe is disabled")
	}
	if order == nil {
		return "", "", errors.New("order is required")
	}
	if strings.TrimSpace(s.cfg.APIKey) == "" {
		return "", "", errors.New("stripe api_key is required")
	}

	stripe.Key = strings.TrimSpace(s.cfg.APIKey)

	paymentMethodTypes := parseCommaList(s.cfg.PaymentMethods)
	if len(paymentMethodTypes) == 0 {
		paymentMethodTypes = []string{"wechat_pay"}
	}
	selectedMethod, err := resolveStripeSelectedPaymentMethod(channel, paymentMethodTypes)
	if err != nil {
		return "", "", err
	}

	currency := strings.ToLower(strings.TrimSpace(s.cfg.Currency))
	if currency == "" {
		currency = "cny"
	}

	amountFen := int64(math.Round(order.AmountCNY * 100))
	if amountFen <= 0 {
		return "", "", errors.New("invalid amount")
	}

	params := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(amountFen),
		Currency:           stripe.String(currency),
		PaymentMethodTypes: stripe.StringSlice([]string{selectedMethod}),
		Description:        stripe.String(fmt.Sprintf("Recharge %s", order.OrderNo)),
		Metadata: map[string]string{
			"order_no": order.OrderNo,
			// Backward/compat keys: some deployments use camelCase keys.
			"orderId":  order.OrderNo,
			"orderNo":  order.OrderNo,
			"provider": order.Provider,
			"channel":  channel,
			"method":   selectedMethod,
		},
		// 关键参数：自动确认 PaymentIntent 以触发 QR 码生成
		Confirm: stripe.Bool(true),
		PaymentMethodData: &stripe.PaymentIntentPaymentMethodDataParams{
			Type: stripe.String(selectedMethod),
		},
	}

	if selectedMethod == "alipay" {
		if returnURL, err := s.resolveReturnURL(order); err != nil {
			return "", "", err
		} else if returnURL != "" {
			params.ReturnURL = stripe.String(returnURL)
		}
	}

	if selectedMethod == "wechat_pay" {
		// 设置 WeChat Pay 选项
		wechatClient := strings.TrimSpace(s.cfg.WechatClient)
		if wechatClient == "" {
			wechatClient = "web" // 默认使用 web 客户端
		}

		params.PaymentMethodOptions = &stripe.PaymentIntentPaymentMethodOptionsParams{
			WeChatPay: &stripe.PaymentIntentPaymentMethodOptionsWeChatPayParams{
				Client: stripe.String(wechatClient),
			},
		}

		if appID := strings.TrimSpace(s.cfg.WechatAppID); appID != "" {
			params.PaymentMethodOptions.WeChatPay.AppID = stripe.String(appID)
		}
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return "", "", fmt.Errorf("create payment_intent: %w", err)
	}

	if intent.NextAction != nil && intent.NextAction.WeChatPayDisplayQRCode != nil {
		payURL = intent.NextAction.WeChatPayDisplayQRCode.HostedInstructionsURL
		// Prefer hosted image if available; frontend can also use `data` to render QR code.
		if intent.NextAction.WeChatPayDisplayQRCode.ImageURLPNG != "" {
			qrURL = intent.NextAction.WeChatPayDisplayQRCode.ImageURLPNG
		} else if intent.NextAction.WeChatPayDisplayQRCode.ImageDataURL != "" {
			qrURL = intent.NextAction.WeChatPayDisplayQRCode.ImageDataURL
		} else {
			qrURL = intent.NextAction.WeChatPayDisplayQRCode.Data
		}
	}
	if intent.NextAction != nil && intent.NextAction.RedirectToURL != nil {
		payURL = strings.TrimSpace(intent.NextAction.RedirectToURL.URL)
	}

	return payURL, qrURL, nil
}

// VerifyWebhook 校验 Stripe Webhook
func (s *StripeService) VerifyWebhook(ctx context.Context, payload []byte, signature string) (*StripeWebhookInfo, error) {
	if s.cfg == nil || !s.cfg.Enabled {
		return nil, errors.New("stripe is disabled")
	}
	if strings.TrimSpace(s.cfg.WebhookSecret) == "" {
		return nil, errors.New("stripe webhook_secret is required")
	}
	log.Printf("[Stripe Webhook] Starting verification: payload_length=%d, has_signature=%v",
		len(payload), signature != "")
	event, err := webhook.ConstructEventWithOptions(
		payload,
		signature,
		strings.TrimSpace(s.cfg.WebhookSecret),
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		return nil, fmt.Errorf("verify webhook: %w", err)
	}

	info := &StripeWebhookInfo{EventType: string(event.Type)}
	info.EventID = event.ID
	log.Printf("[Stripe Webhook] Event parsed successfully: event_id=%s, type=%s, api_version=%s",
		event.ID, event.Type, event.APIVersion)

	switch info.EventType {
	case "payment_intent.succeeded":
		fallthrough
	case "payment_intent.payment_failed":
		fallthrough
	case "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, fmt.Errorf("parse payment_intent: %w", err)
		}
		log.Printf("[Stripe Webhook] PaymentIntent details: id=%s, amount=%d, currency=%s, status=%s",
			pi.ID, pi.Amount, pi.Currency, pi.Status)
		info.OrderNo = stripeFirstNonEmpty(
			pi.Metadata["order_no"],
			pi.Metadata["orderId"],
			pi.Metadata["order_id"],
			pi.Metadata["orderNo"],
		)
		info.TradeNo = pi.ID
		info.Amount = pi.Amount
		info.Currency = strings.ToLower(string(pi.Currency))
		if pi.LastPaymentError != nil && strings.TrimSpace(pi.LastPaymentError.Msg) != "" {
			info.FailureMessage = strings.TrimSpace(pi.LastPaymentError.Msg)
		} else if pi.CancellationReason != "" {
			info.FailureMessage = string(pi.CancellationReason)
		}
		log.Printf("[Stripe Webhook] Extracted metadata: order_no=%s, failure_message=%s",
			info.OrderNo, info.FailureMessage)
		return info, nil
	default:
		return info, nil
	}
}

func stripeFirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func parseCommaList(value string) []string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func resolveStripeSelectedPaymentMethod(channel string, enabledMethods []string) (string, error) {
	selected := stripePaymentMethodFromChannel(channel)
	if selected == "" {
		selected = firstUserVisibleStripeMethod(enabledMethods)
	}
	if selected == "" {
		return "", errors.New("stripe payment method is not supported")
	}
	for _, method := range enabledMethods {
		if normalizeStripePaymentMethod(method) == selected {
			return selected, nil
		}
	}
	return "", fmt.Errorf("stripe payment method %s is disabled", selected)
}

func stripePaymentMethodFromChannel(channel string) string {
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case "alipay", "stripe_alipay":
		return "alipay"
	case "wechat", "wxpay", "wechat_pay", "stripe", "stripe_wechat", "stripe_wxpay":
		return "wechat_pay"
	default:
		return ""
	}
}

func firstUserVisibleStripeMethod(methods []string) string {
	for _, method := range methods {
		normalized := normalizeStripePaymentMethod(method)
		if normalized == "alipay" || normalized == "wechat_pay" {
			return normalized
		}
	}
	return ""
}

func normalizeStripePaymentMethod(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "alipay":
		return "alipay"
	case "wechat", "wxpay", "wechat_pay":
		return "wechat_pay"
	default:
		return ""
	}
}

func (s *StripeService) resolveReturnURL(order *PaymentOrder) (string, error) {
	if s == nil || s.cfg == nil || order == nil {
		return "", nil
	}
	raw := strings.TrimSpace(s.cfg.SuccessURL)
	if raw == "" {
		return "", nil
	}
	raw = strings.ReplaceAll(raw, "{ORDER_ID}", url.QueryEscape(order.OrderNo))
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw, nil
	}
	base := strings.TrimRight(strings.TrimSpace(s.paymentBaseURL), "/")
	if base == "" {
		return "", errors.New("payment.base_url is required for relative stripe success_url")
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	return base + raw, nil
}
