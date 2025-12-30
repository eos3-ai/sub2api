package service

import "time"

const (
	PromotionStatusActive  = "active"
	PromotionStatusUsed    = "used"
	PromotionStatusExpired = "expired"
)

// UserPromotion 表示用户的活动资格
type UserPromotion struct {
	ID          int64
	UserID      int64
	Username    string
	ActivatedAt time.Time
	ExpireAt    time.Time
	Status      string

	UsedTier    *int
	UsedAt      *time.Time
	UsedAmount  *float64
	BonusAmount *float64

	CreatedAt time.Time
	UpdatedAt time.Time
}

// PromotionRecord 记录活动使用历史
type PromotionRecord struct {
	ID        int64
	UserID    int64
	Username  string
	Tier      int
	Amount    float64
	Bonus     float64
	CreatedAt time.Time
}
