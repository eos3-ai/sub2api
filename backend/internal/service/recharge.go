package service

import "time"

// Recharge record types
const (
	RechargeTypeAdmin    = "admin"
	RechargeTypePayment  = "payment"
	RechargeTypeRedeem   = "redeem"
	RechargeTypeReferral = "referral"
	RechargeTypeDeduct   = "deduct"
	RechargeTypeRefund   = "refund"
)

// RechargeRecord 表示余额变动流水
type RechargeRecord struct {
	ID            int64
	UserID        int64
	Amount        float64
	Type          string
	Operator      string
	Remark        string
	RelatedID     *string
	BalanceBefore float64
	BalanceAfter  float64
	CreatedAt     time.Time
}
