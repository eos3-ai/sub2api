package service

import (
	"context"
	"fmt"
	"math"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// BalanceChangeRequest 定义余额变动请求
type BalanceChangeRequest struct {
	UserID    int64
	Amount    float64
	Type      string
	Operator  string
	Remark    string
	RelatedID *string
}

// BalanceService 封装余额操作和流水记录
type BalanceService struct {
	userRepo            UserRepository
	rechargeRecordRepo  RechargeRecordRepository
	billingCacheService *BillingCacheService
}

func NewBalanceService(
	userRepo UserRepository,
	rechargeRecordRepo RechargeRecordRepository,
	billingCacheService *BillingCacheService,
) *BalanceService {
	return &BalanceService{
		userRepo:            userRepo,
		rechargeRecordRepo:  rechargeRecordRepo,
		billingCacheService: billingCacheService,
	}
}

// ApplyChange 调整用户余额并记录流水
func (s *BalanceService) ApplyChange(ctx context.Context, req BalanceChangeRequest) (*RechargeRecord, error) {
	if req.Amount == 0 {
		return nil, fmt.Errorf("amount must not be zero")
	}
	if req.UserID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("record type is required")
	}

	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	balanceBefore := user.Balance
	if req.Amount > 0 {
		if err := s.userRepo.UpdateBalance(ctx, req.UserID, req.Amount); err != nil {
			return nil, fmt.Errorf("increase balance: %w", err)
		}
	} else {
		if err := s.userRepo.DeductBalance(ctx, req.UserID, math.Abs(req.Amount)); err != nil {
			return nil, fmt.Errorf("deduct balance: %w", err)
		}
	}

	balanceAfter := balanceBefore + req.Amount
	record := &RechargeRecord{
		UserID:        req.UserID,
		Amount:        req.Amount,
		Type:          req.Type,
		Operator:      req.Operator,
		Remark:        req.Remark,
		RelatedID:     req.RelatedID,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
	}

	if err := s.rechargeRecordRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("create recharge record: %w", err)
	}

	// 余额变动后失效缓存
	if s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateUserBalance(ctx, req.UserID)
	}

	return record, nil
}

// ListUserRecords 获取用户的充值记录
func (s *BalanceService) ListUserRecords(ctx context.Context, userID int64, params pagination.PaginationParams) ([]RechargeRecord, *pagination.PaginationResult, error) {
	records, paginationResult, err := s.rechargeRecordRepo.ListByUser(ctx, userID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list recharge records: %w", err)
	}
	return records, paginationResult, nil
}

// ListRecords 管理员查询
func (s *BalanceService) ListRecords(ctx context.Context, params pagination.PaginationParams, filter RechargeRecordListFilter) ([]RechargeRecord, *pagination.PaginationResult, error) {
	records, paginationResult, err := s.rechargeRecordRepo.ListAll(ctx, params, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("list recharge records: %w", err)
	}
	return records, paginationResult, nil
}
