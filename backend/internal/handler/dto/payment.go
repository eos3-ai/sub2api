package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// PaymentPlan represents a frontend-friendly payment package definition.
type PaymentPlan struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	AmountUSD    float64 `json:"amount_usd"`
	PayUSD       float64 `json:"pay_usd"`
	CreditsUSD   float64 `json:"credits_usd"`
	ExchangeRate float64 `json:"exchange_rate"`
	DiscountRate float64 `json:"discount_rate"`
	Enabled      bool    `json:"enabled"`
}

// PaymentOrder represents a payment order for API responses.
type PaymentOrder struct {
	ID      int64  `json:"id"`
	OrderNo string `json:"order_no"`
	TradeNo *string `json:"trade_no,omitempty"`

	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	UserEmail string `json:"user_email,omitempty"` // admin list/export convenience

	Remark       string  `json:"remark,omitempty"`
	AmountCNY    float64 `json:"amount_cny"`
	AmountUSD    float64 `json:"amount_usd"`
	BonusUSD     float64 `json:"bonus_usd"`
	TotalUSD     float64 `json:"total_usd"`
	ExchangeRate float64 `json:"exchange_rate"`
	DiscountRate float64 `json:"discount_rate"`

	Provider      string `json:"provider"`
	Channel       string `json:"channel"`
	PaymentMethod string `json:"payment_method"`
	PaymentURL    string `json:"payment_url,omitempty"`

	Status   string     `json:"status"`
	PaidAt   *time.Time `json:"paid_at,omitempty"`
	ExpireAt time.Time  `json:"expire_at"`

	PromotionTier *int `json:"promotion_tier,omitempty"`
	PromotionUsed bool `json:"promotion_used"`

	CallbackData string     `json:"callback_data,omitempty"`
	CallbackAt   *time.Time `json:"callback_at,omitempty"`

	ClientIP  string `json:"client_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func PaymentOrderFromService(o *service.PaymentOrder) *PaymentOrder {
	if o == nil {
		return nil
	}
	return &PaymentOrder{
		ID:           o.ID,
		OrderNo:      o.OrderNo,
		TradeNo:      o.TradeNo,
		UserID:       o.UserID,
		Username:     o.Username,
		Remark:       o.Remark,
		AmountCNY:    o.AmountCNY,
		AmountUSD:    o.AmountUSD,
		BonusUSD:     o.BonusUSD,
		TotalUSD:     o.TotalUSD,
		ExchangeRate: o.ExchangeRate,
		DiscountRate: o.DiscountRate,
		Provider:     o.Provider,
		Channel:      o.Channel,
		PaymentMethod: o.PaymentMethod,
		PaymentURL:   o.PaymentURL,
		Status:       o.Status,
		PaidAt:       o.PaidAt,
		ExpireAt:     o.ExpireAt,
		PromotionTier: o.PromotionTier,
		PromotionUsed: o.PromotionUsed,
		CallbackData: o.CallbackData,
		CallbackAt:   o.CallbackAt,
		ClientIP:     o.ClientIP,
		UserAgent:    o.UserAgent,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}
}

