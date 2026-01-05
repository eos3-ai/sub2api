package service

import (
	"context"
	"fmt"
	"log"
)

// BalanceBonusIssuer 余额发放实现
type BalanceBonusIssuer struct {
	balanceService *BalanceService
}

// NewBalanceBonusIssuer 创建余额发放器
func NewBalanceBonusIssuer(balanceService *BalanceService) BonusIssuer {
	return &BalanceBonusIssuer{
		balanceService: balanceService,
	}
}

// Issue 发放赠送到用户余额
func (i *BalanceBonusIssuer) Issue(ctx context.Context, req *IssueRequest) error {
	log.Printf("[BonusIssuer] Issue ENTRY: user_id=%d, amount=%.2f, type=%s, remark=%s, balanceService_is_nil=%v",
		req.UserID, req.Amount, req.Type, req.Remark, i.balanceService == nil)

	if i.balanceService == nil {
		return fmt.Errorf("balance service not available")
	}

	// 直接使用传入的RelatedID（订单号）
	relatedID := req.RelatedID

	// 发放余额
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
