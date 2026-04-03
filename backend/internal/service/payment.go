package service

import (
	"strings"
	"time"
)

const (
	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusFailed    = "failed"
	PaymentStatusExpired   = "expired"
	PaymentStatusCancelled = "cancelled"
	PaymentStatusRefunded  = "refunded"
)

const (
	PaymentBizTypeOnlineRecharge       = "online_recharge"
	PaymentBizTypeSubscriptionPurchase = "subscription_purchase"
)

// PaymentOrder 表示支付订单
type PaymentOrder struct {
	ID              int64
	OrderNo         string
	TradeNo         *string
	UserID          int64
	Username        string
	Remark          string
	AmountCNY       float64
	AmountUSD       float64
	BonusUSD        float64
	TotalUSD        float64
	ExchangeRate    float64
	DiscountRate    float64
	BizType         string
	BizGroupID      *int64
	BizValidityDays *int

	Provider      string
	Channel       string // 支付渠道(alipay/wechat/zpay/stripe)
	PaymentMethod string
	PaymentURL    string

	Status   string
	PaidAt   *time.Time
	ExpireAt time.Time

	PromotionTier *int
	PromotionUsed bool

	CallbackData string
	CallbackAt   *time.Time

	ClientIP  string
	UserAgent string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (o *PaymentOrder) NormalizedBizType() string {
	if o == nil {
		return PaymentBizTypeOnlineRecharge
	}
	v := strings.ToLower(strings.TrimSpace(o.BizType))
	switch v {
	case PaymentBizTypeSubscriptionPurchase:
		return PaymentBizTypeSubscriptionPurchase
	default:
		return PaymentBizTypeOnlineRecharge
	}
}

func (o *PaymentOrder) IsSubscriptionPurchase() bool {
	return o.NormalizedBizType() == PaymentBizTypeSubscriptionPurchase
}
