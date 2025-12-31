package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

// StripeService 封装 Stripe 渠道
type StripeService struct {
	cfg *config.StripeConfig
}

type StripeWebhookInfo struct {
	OrderNo        string
	TradeNo        string
	EventType      string
	Amount         int64
	Currency       string
	FailureMessage string
}

func NewStripeService(cfg *config.Config) *StripeService {
	var stripeCfg *config.StripeConfig
	if cfg != nil {
		stripeCfg = &cfg.Payment.Stripe
	}
	return &StripeService{cfg: stripeCfg}
}

// CreatePayment creates a Stripe PaymentIntent for WeChat Pay and returns pay_url/qr_url.
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
		PaymentMethodTypes: stripe.StringSlice(paymentMethodTypes),
		Description:        stripe.String(fmt.Sprintf("Recharge %s", order.OrderNo)),
		Metadata: map[string]string{
			"order_no": order.OrderNo,
			"provider": order.Provider,
			"channel":  channel,
		},
		// 关键参数：自动确认 PaymentIntent 以触发 QR 码生成
		Confirm: stripe.Bool(true),
		// 创建 WeChat Pay 支付方式
		PaymentMethodData: &stripe.PaymentIntentPaymentMethodDataParams{
			Type: stripe.String("wechat_pay"),
		},
	}

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
		info.OrderNo = pi.Metadata["order_no"]
		info.TradeNo = pi.ID
		info.Amount = pi.Amount
		info.Currency = strings.ToLower(string(pi.Currency))
		if pi.LastPaymentError != nil && strings.TrimSpace(pi.LastPaymentError.Msg) != "" {
			info.FailureMessage = strings.TrimSpace(pi.LastPaymentError.Msg)
		} else if pi.CancellationReason != "" {
			info.FailureMessage = string(pi.CancellationReason)
		}
		return info, nil
	default:
		return info, nil
	}
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
