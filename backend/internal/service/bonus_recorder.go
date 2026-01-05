package service

import (
	"context"
	"fmt"
	"log"
)

// PromotionBonusRecorder Promotion记录实现
type PromotionBonusRecorder struct {
	promotionService *PromotionService
}

// NewPromotionBonusRecorder 创建Promotion记录器
func NewPromotionBonusRecorder(promotionService *PromotionService) BonusRecorder {
	return &PromotionBonusRecorder{
		promotionService: promotionService,
	}
}

// Record 记录赠送使用情况
func (r *PromotionBonusRecorder) Record(ctx context.Context, req *RecordRequest) error {
	log.Printf("[BonusRecorder] Record ENTRY: user_id=%d, tier=%d, amount_usd=%.2f, bonus_usd=%.2f, promotionService_is_nil=%v",
		req.UserID, req.Tier, req.AmountUSD, req.BonusUSD, r.promotionService == nil)

	if r.promotionService == nil {
		return fmt.Errorf("promotion service not available")
	}

	log.Printf("[BonusRecorder] Recording promotion usage: user_id=%d, tier=%d, amount_usd=%.2f, bonus_usd=%.2f",
		req.UserID, req.Tier, req.AmountUSD, req.BonusUSD)

	err := r.promotionService.MarkPromotionUsed(
		ctx,
		req.UserID,
		req.Tier,
		req.AmountUSD,
		req.BonusUSD,
	)

	if err != nil {
		log.Printf("[BonusRecorder] ERROR: Failed to record promotion usage: user_id=%d, tier=%d, error=%v",
			req.UserID, req.Tier, err)
		return fmt.Errorf("mark promotion used: %w", err)
	}

	log.Printf("[BonusRecorder] Successfully recorded promotion usage: user_id=%d, tier=%d",
		req.UserID, req.Tier)

	return nil
}
