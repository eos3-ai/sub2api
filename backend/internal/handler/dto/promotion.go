package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type PromotionTier struct {
	Hours        int     `json:"hours"`
	BonusPercent float64 `json:"bonus_percent"`
}

type PromotionCurrentTier struct {
	Tier         int     `json:"tier"`
	Hours        int     `json:"hours"`
	BonusPercent float64 `json:"bonus_percent"`
}

type UserPromotion struct {
	UserID      int64      `json:"user_id"`
	Status      string     `json:"status"`
	ActivatedAt time.Time  `json:"activated_at"`
	ExpireAt    time.Time  `json:"expire_at"`
	UsedTier    *int       `json:"used_tier,omitempty"`
	UsedAt      *time.Time `json:"used_at,omitempty"`
	UsedAmount  *float64   `json:"used_amount,omitempty"`
	BonusAmount *float64   `json:"bonus_amount,omitempty"`
}

type PromotionStatusResponse struct {
	Enabled           bool                  `json:"enabled"`
	Status            string                `json:"status"` // none/active/used/expired
	Promotion         *UserPromotion        `json:"promotion,omitempty"`
	CurrentTier       *PromotionCurrentTier `json:"current_tier,omitempty"`
	CurrentTierEndsAt *time.Time            `json:"current_tier_ends_at,omitempty"`
	// CurrentTierRemainingSeconds is the remaining seconds until the current tier ends (i.e. next tier boundary).
	CurrentTierRemainingSeconds int64           `json:"current_tier_remaining_seconds,omitempty"`
	RemainingSeconds            int64           `json:"remaining_seconds,omitempty"`
	Tiers                       []PromotionTier `json:"tiers,omitempty"`
	DurationHours               int             `json:"duration_hours,omitempty"`
	MinAmountUSD                float64         `json:"min_amount_usd,omitempty"`
}

func UserPromotionFromService(p *service.UserPromotion) *UserPromotion {
	if p == nil {
		return nil
	}
	return &UserPromotion{
		UserID:      p.UserID,
		Status:      p.Status,
		ActivatedAt: p.ActivatedAt,
		ExpireAt:    p.ExpireAt,
		UsedTier:    p.UsedTier,
		UsedAt:      p.UsedAt,
		UsedAmount:  p.UsedAmount,
		BonusAmount: p.BonusAmount,
	}
}

func PromotionTiersFromConfig(tiers []config.PromotionTier) []PromotionTier {
	out := make([]PromotionTier, 0, len(tiers))
	for _, tier := range tiers {
		out = append(out, PromotionTier{
			Hours:        tier.Hours,
			BonusPercent: tier.BonusPercent,
		})
	}
	return out
}

func CurrentPromotionTier(tiers []config.PromotionTier, activatedAt time.Time, expireAt time.Time, now time.Time) *PromotionCurrentTier {
	if activatedAt.IsZero() || expireAt.IsZero() {
		return nil
	}
	if now.After(expireAt) {
		return nil
	}
	elapsedHours := now.Sub(activatedAt).Hours()
	for i, tier := range tiers {
		if tier.Hours <= 0 || tier.BonusPercent <= 0 {
			continue
		}
		if elapsedHours <= float64(tier.Hours) {
			return &PromotionCurrentTier{
				Tier:         i,
				Hours:        tier.Hours,
				BonusPercent: tier.BonusPercent,
			}
		}
	}
	return nil
}
