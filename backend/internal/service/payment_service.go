package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// CreatePaymentOrderRequest 定义创建订单请求
type CreatePaymentOrderRequest struct {
	UserID        int64
	Username      string
	AmountCNY     float64
	Provider      string
	PaymentMethod string
	ClientIP      string
	UserAgent     string
}

// PaymentService 支付核心服务
type PaymentService struct {
	cfg              *config.PaymentConfig
	orderRepo        PaymentOrderRepository
	paymentCache     PaymentCache
	balanceService   *BalanceService
	promotionService *PromotionService
	referralService  *ReferralService
}

func NewPaymentService(
	cfg *config.Config,
	orderRepo PaymentOrderRepository,
	paymentCache PaymentCache,
	balanceService *BalanceService,
	promotionService *PromotionService,
	referralService *ReferralService,
) *PaymentService {
	var paymentCfg *config.PaymentConfig
	if cfg != nil {
		paymentCfg = &cfg.Payment
	}
	return &PaymentService{
		cfg:              paymentCfg,
		orderRepo:        orderRepo,
		paymentCache:     paymentCache,
		balanceService:   balanceService,
		promotionService: promotionService,
		referralService:  referralService,
	}
}

// CreateOrder 创建支付订单
func (s *PaymentService) CreateOrder(ctx context.Context, req *CreatePaymentOrderRequest) (*PaymentOrder, error) {
	if s == nil || s.cfg == nil || !s.cfg.Enabled {
		return nil, errors.New("payment is disabled")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}
	if req.AmountCNY < s.cfg.MinAmount || req.AmountCNY > s.cfg.MaxAmount {
		return nil, fmt.Errorf("amount must be between %.2f and %.2f", s.cfg.MinAmount, s.cfg.MaxAmount)
	}

	if s.paymentCache != nil && s.cfg.MaxOrdersPerMinute > 0 {
		count, err := s.paymentCache.IncrementUserOrderCounter(ctx, req.UserID, time.Minute)
		if err == nil && count > s.cfg.MaxOrdersPerMinute {
			return nil, fmt.Errorf("too many orders, please try later")
		}
	}

	amountUSD := req.AmountCNY / s.cfg.ExchangeRate
	discount := normalizedDiscountRate(s.cfg.DiscountRate)
	creditsUSD := amountUSD
	if discount > 0 && discount < 1 {
		// amountUSD is the payable USD because AmountCNY is computed from:
		// creditsUSD * discount * exchangeRate in the handler.
		creditsUSD = amountUSD / discount
	}
	order := &PaymentOrder{
		OrderNo:       s.generateOrderNo(),
		UserID:        req.UserID,
		Username:      req.Username,
		AmountCNY:     req.AmountCNY,
		AmountUSD:     amountUSD,
		TotalUSD:      creditsUSD,
		ExchangeRate:  s.cfg.ExchangeRate,
		DiscountRate:  discount,
		Provider:      strings.ToLower(req.Provider),
		PaymentMethod: req.PaymentMethod,
		Status:        PaymentStatusPending,
		ExpireAt:      time.Now().Add(time.Duration(s.cfg.OrderExpireMinutes) * time.Minute),
		ClientIP:      req.ClientIP,
		UserAgent:     req.UserAgent,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return order, nil
}

// GetOrderByOrderNo 获取订单
func (s *PaymentService) GetOrderByOrderNo(ctx context.Context, orderNo string) (*PaymentOrder, error) {
	if s == nil {
		return nil, nil
	}
	return s.orderRepo.GetByOrderNo(ctx, orderNo)
}

// UpdateOrder persists order fields (e.g. PaymentURL) to the repository.
func (s *PaymentService) UpdateOrder(ctx context.Context, order *PaymentOrder) error {
	if s == nil {
		return nil
	}
	if order == nil {
		return errors.New("order is required")
	}
	return s.orderRepo.Update(ctx, order)
}

// MarkOrderFailed marks a pending order as failed and stores the error message.
func (s *PaymentService) MarkOrderFailed(ctx context.Context, orderNo string, reason string) (*PaymentOrder, error) {
	if s == nil {
		return nil, nil
	}
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != PaymentStatusPending {
		return order, nil
	}

	now := timePtr(time.Now())
	order.Status = PaymentStatusFailed
	order.CallbackAt = now
	if reason != "" {
		order.CallbackData = reason
	}
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}
	return order, nil
}

// ListUserOrders 用户订单列表
func (s *PaymentService) ListUserOrders(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]PaymentOrder, *pagination.PaginationResult, error) {
	if s == nil {
		return nil, nil, nil
	}
	return s.orderRepo.ListByUser(ctx, userID, params, status)
}

// ListOrders 管理员订单列表
func (s *PaymentService) ListOrders(ctx context.Context, params pagination.PaginationParams, filter PaymentOrderFilter) ([]PaymentOrder, *pagination.PaginationResult, error) {
	if s == nil {
		return nil, nil, nil
	}
	return s.orderRepo.List(ctx, params, filter)
}

// MarkOrderPaid 处理支付成功逻辑
func (s *PaymentService) MarkOrderPaid(ctx context.Context, orderNo, tradeNo string, callbackPayload any) (*PaymentOrder, error) {
	if s == nil {
		return nil, nil
	}
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != PaymentStatusPending {
		return order, nil
	}

	var callbackData string
	if callbackPayload != nil {
		if bytes, err := json.Marshal(callbackPayload); err == nil {
			callbackData = string(bytes)
		}
	}

	order.Status = PaymentStatusPaid
	now := timePtr(time.Now())
	order.PaidAt = now
	order.TradeNo = &tradeNo
	order.CallbackData = callbackData
	order.CallbackAt = now

	var promotionTier *int
	var bonusAmount float64
	if s.promotionService != nil {
		result, err := s.promotionService.ApplyPromotion(ctx, order.UserID, order.AmountUSD)
		if err == nil && result != nil && result.Applied {
			bonusAmount = result.Bonus
			// Credits base is already computed on order creation (order.TotalUSD).
			order.TotalUSD = order.TotalUSD + bonusAmount
			promotionTier = &result.Tier
		}
	}
	if promotionTier != nil {
		order.PromotionTier = promotionTier
		order.PromotionUsed = true
	}
	order.BonusUSD = bonusAmount

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}

	// 充值余额
	if s.balanceService != nil {
		_, err := s.balanceService.ApplyChange(ctx, BalanceChangeRequest{
			UserID:    order.UserID,
			Amount:    order.TotalUSD,
			Type:      RechargeTypePayment,
			Operator:  "system",
			Remark:    fmt.Sprintf("payment %s", order.Provider),
			RelatedID: &order.OrderNo,
		})
		if err != nil {
			return nil, fmt.Errorf("apply balance: %w", err)
		}
	}

	if promotionTier != nil && s.promotionService != nil {
		_ = s.promotionService.MarkPromotionUsed(ctx, order.UserID, *promotionTier, order.AmountUSD, bonusAmount)
	}

	if s.referralService != nil {
		reward, err := s.referralService.ProcessInviteeRecharge(ctx, &ReferralRechargeRequest{
			InviteeID:         order.UserID,
			RechargeAmountUSD: order.AmountUSD,
		})
		if err == nil && reward != nil && reward.ShouldIssue {
			_, _ = s.balanceService.ApplyChange(ctx, BalanceChangeRequest{
				UserID:    reward.ReferrerID,
				Amount:    reward.RewardAmountUSD,
				Type:      RechargeTypeReferral,
				Operator:  "system",
				Remark:    fmt.Sprintf("referral reward for %d", order.UserID),
				RelatedID: &order.OrderNo,
			})
			_ = s.referralService.MarkRewardIssued(ctx, order.UserID, reward.RewardAmountUSD)
		}
	}

	return order, nil
}

// CancelExpiredOrders 更新过期订单
func (s *PaymentService) CancelExpiredOrders(ctx context.Context) (int64, error) {
	if s == nil {
		return 0, nil
	}
	return s.orderRepo.MarkExpired(ctx, time.Now())
}

func (s *PaymentService) generateOrderNo() string {
	now := time.Now().UTC()
	random := rand.Intn(1000000)
	return "PO" + now.Format("20060102150405") + fmt.Sprintf("%06d", random)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func normalizedDiscountRate(discountRate float64) float64 {
	// discountRate is a payable multiplier in (0,1], e.g. 0.15 means "pay 15%".
	// Compatibility: historical default was 1.0 (pay full).
	if discountRate <= 0 {
		return 1.0
	}
	if discountRate > 1 {
		return 1.0
	}
	return discountRate
}
