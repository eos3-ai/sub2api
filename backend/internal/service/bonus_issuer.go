package service

import (
	"context"
	"fmt"
	"log"
)

// BalanceBonusIssuer 余额发放实现
type BalanceBonusIssuer struct {
	balanceService *BalanceService
	paymentService *PaymentService // 用于创建activity order
}

// NewBalanceBonusIssuer 创建余额发放器
func NewBalanceBonusIssuer(balanceService *BalanceService, paymentService *PaymentService) BonusIssuer {
	return &BalanceBonusIssuer{
		balanceService: balanceService,
		paymentService: paymentService,
	}
}

// Issue 发放赠送到用户余额
func (i *BalanceBonusIssuer) Issue(ctx context.Context, req *IssueRequest) error {
	if i.balanceService == nil {
		return fmt.Errorf("balance service not available")
	}

	relatedID := req.RelatedID

	// 1. 创建活动订单（用于记录）
	if i.paymentService != nil && req.ActivityOrderNo == nil {
		activityOrder, err := i.paymentService.createActivityRechargeOrder(ctx, req.UserID, req.Amount, req.Remark)
		if err != nil {
			log.Printf("[BonusIssuer] WARNING: Failed to create activity order: user_id=%d, amount=%.2f, error=%v",
				req.UserID, req.Amount, err)
			// 继续执行，activity order失败不阻断余额发放
		} else if activityOrder != nil {
			orderNo := activityOrder.OrderNo
			relatedID = &orderNo
		}
	}

	// 2. 发放余额
	log.Printf("[BonusIssuer] Issuing bonus: user_id=%d, amount=%.2f, type=%s, remark=%s",
		req.UserID, req.Amount, req.Type, req.Remark)

	_, err := i.balanceService.ApplyChange(ctx, BalanceChangeRequest{
		UserID:    req.UserID,
		Amount:    req.Amount,
		Type:      req.Type,
		Operator:  "system",
		Remark:    req.Remark,
		RelatedID: relatedID,
	})

	if err != nil {
		log.Printf("[BonusIssuer] ERROR: Failed to issue bonus: user_id=%d, amount=%.2f, type=%s, error=%v",
			req.UserID, req.Amount, req.Type, err)
		return fmt.Errorf("apply balance change: %w", err)
	}

	log.Printf("[BonusIssuer] Successfully issued bonus: user_id=%d, amount=%.2f, type=%s",
		req.UserID, req.Amount, req.Type)

	return nil
}
