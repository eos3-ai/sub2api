package service

import (
	"context"
	"fmt"
	"log"
)

// BonusService 统一的赠送处理服务
type BonusService struct {
	calculator BonusCalculator
	issuer     BonusIssuer
	recorder   BonusRecorder
}

// NewBonusService 创建BonusService实例
func NewBonusService(calculator BonusCalculator, issuer BonusIssuer, recorder BonusRecorder) *BonusService {
	return &BonusService{
		calculator: calculator,
		issuer:     issuer,
		recorder:   recorder,
	}
}

// ProcessOrderBonus 处理订单的赠送逻辑（计算+发放+记录）
func (s *BonusService) ProcessOrderBonus(ctx context.Context, req *BonusRequest) (*BonusResult, error) {
	if s == nil {
		log.Printf("[BonusService] ERROR: BonusService instance is nil")
		return nil, fmt.Errorf("bonus service is nil")
	}

	log.Printf("[BonusService] ProcessOrderBonus ENTRY: order_no=%s, user_id=%d, amount_usd=%.2f, provider=%s, calculator_is_nil=%v, issuer_is_nil=%v, recorder_is_nil=%v",
		req.OrderNo, req.UserID, req.AmountUSD, req.Provider, s.calculator == nil, s.issuer == nil, s.recorder == nil)

	result := &BonusResult{
		Applied: false,
	}

	// 1. 计算赠送金额
	if s.calculator == nil {
		log.Printf("[BonusService] Calculator not available, skipping bonus calculation")
		return result, nil
	}

	log.Printf("[BonusService] Calling Calculator.Calculate for user_id=%d, amount_usd=%.2f", req.UserID, req.AmountUSD)
	calculation, err := s.calculator.Calculate(ctx, req.UserID, req.AmountUSD)
	log.Printf("[BonusService] Calculator.Calculate returned: calculation=%+v, err=%v", calculation, err)
	if err != nil {
		log.Printf("[BonusService] ERROR: Calculation failed: order_no=%s, user_id=%d, error=%v",
			req.OrderNo, req.UserID, err)
		result.Error = fmt.Errorf("calculate bonus: %w", err)
		return result, result.Error
	}

	if calculation == nil || !calculation.ShouldGrant {
		log.Printf("[BonusService] Bonus not applicable: order_no=%s, user_id=%d, reason=%s",
			req.OrderNo, req.UserID, calculation.Reason)
		return result, nil
	}

	log.Printf("[BonusService] Bonus calculated: order_no=%s, user_id=%d, amount=%.2f, percent=%.2f%%, tier=%d",
		req.OrderNo, req.UserID, calculation.Amount, calculation.Percent, calculation.Tier)

	// 2. 发放赠送
	if s.issuer == nil {
		err := fmt.Errorf("bonus issuer not available")
		log.Printf("[BonusService] ERROR: %v", err)
		result.Error = err
		return result, err
	}

	issueReq := &IssueRequest{
		UserID:    req.UserID,
		Amount:    calculation.Amount,
		Type:      RechargeTypePromotion,
		Remark:    "新用户首充奖励",
		RelatedID: &req.OrderNo,
	}

	if err := s.issuer.Issue(ctx, issueReq); err != nil {
		log.Printf("[BonusService] ERROR: Issue failed: order_no=%s, user_id=%d, amount=%.2f, error=%v",
			req.OrderNo, req.UserID, calculation.Amount, err)
		result.Error = fmt.Errorf("issue bonus: %w", err)
		// 注意：即使发放失败，也继续执行（不return），确保记录promotion已使用
	}

	// 3. 记录赠送使用（无论发放是否成功，都要记录已使用，防止重复使用）
	if s.recorder != nil {
		recordReq := &RecordRequest{
			UserID:    req.UserID,
			Tier:      calculation.Tier,
			AmountUSD: req.AmountUSD,
			BonusUSD:  calculation.Amount,
		}

		if recordErr := s.recorder.Record(ctx, recordReq); recordErr != nil {
			log.Printf("[BonusService] ERROR: Record failed: order_no=%s, user_id=%d, tier=%d, error=%v",
				req.OrderNo, req.UserID, calculation.Tier, recordErr)
			// 记录失败：如果之前发放成功，这是数据一致性问题
			if result.Error == nil {
				result.Error = fmt.Errorf("record bonus usage: %w", recordErr)
			}
		} else {
			log.Printf("[BonusService] Promotion marked as used: order_no=%s, user_id=%d, tier=%d",
				req.OrderNo, req.UserID, calculation.Tier)
		}
	}

	// 如果发放失败，返回错误
	if result.Error != nil {
		log.Printf("[BonusService] Bonus processing failed: order_no=%s, user_id=%d, error=%v",
			req.OrderNo, req.UserID, result.Error)
		return result, result.Error
	}

	// 成功：填充结果
	result.Applied = true
	result.BonusAmount = calculation.Amount
	result.Tier = calculation.Tier
	result.BonusPercent = calculation.Percent

	log.Printf("[BonusService] Bonus processing completed successfully: order_no=%s, user_id=%d, amount=%.2f, tier=%d",
		req.OrderNo, req.UserID, result.BonusAmount, result.Tier)

	return result, nil
}
