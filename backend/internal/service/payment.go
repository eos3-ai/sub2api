package service

import "time"

const (
	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusFailed    = "failed"
	PaymentStatusExpired   = "expired"
	PaymentStatusCancelled = "cancelled"
	PaymentStatusRefunded  = "refunded"
)

// PaymentOrder 表示支付订单
type PaymentOrder struct {
	ID           int64
	OrderNo      string
	TradeNo      *string
	UserID       int64
	Username     string
	Remark       string
	AmountCNY    float64
	AmountUSD    float64
	BonusUSD     float64
	TotalUSD     float64
	ExchangeRate float64
	DiscountRate float64

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
