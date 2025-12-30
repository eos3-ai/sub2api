package service

import "time"

// ReferralCode 表示用户的邀请码
type ReferralCode struct {
	ID        int64
	UserID    int64
	Code      string
	CreatedAt time.Time
}

// ReferralInvite 表示邀请关系
type ReferralInvite struct {
	ID               int64
	InviteeID        int64
	InviteeUsername  string
	ReferrerID       int64
	ReferrerUsername string
	TotalRechargeUSD float64
	IsQualified      bool
	QualifiedAt      *time.Time
	RewardIssued     bool
	RewardIssuedAt   *time.Time
	RewardAmountUSD  float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
