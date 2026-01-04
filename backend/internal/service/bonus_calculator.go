package service

import (
	"context"
	"fmt"
)

// PromotionCalculator Promotion策略的计算器实现
type PromotionCalculator struct {
	promotionService *PromotionService
}

// NewPromotionCalculator 创建Promotion计算器
func NewPromotionCalculator(promotionService *PromotionService) BonusCalculator {
	return &PromotionCalculator{
		promotionService: promotionService,
	}
}

// Calculate 计算promotion赠送金额
func (c *PromotionCalculator) Calculate(ctx context.Context, userID int64, amountUSD float64) (*CalculationResult, error) {
	if c.promotionService == nil {
		return &CalculationResult{
			ShouldGrant: false,
			Reason:      "promotion service not available",
		}, nil
	}

	// 调用现有的 PromotionService.ApplyPromotion()
	result, err := c.promotionService.ApplyPromotion(ctx, userID, amountUSD)
	if err != nil {
		return nil, fmt.Errorf("apply promotion: %w", err)
	}

	if result == nil || !result.Applied {
		return &CalculationResult{
			ShouldGrant: false,
			Reason:      "promotion not applicable",
		}, nil
	}

	return &CalculationResult{
		ShouldGrant: true,
		Amount:      result.Bonus,
		Percent:     result.BonusPercent,
		Tier:        result.Tier,
		Reason:      fmt.Sprintf("新用户首充阶梯%d奖励", result.Tier),
	}, nil
}
