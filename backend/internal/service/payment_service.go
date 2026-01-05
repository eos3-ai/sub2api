package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/infrastructure/errors"
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
	bonusService     *BonusService
	promotionService *PromotionService
	referralService  *ReferralService
	dingtalkService  *DingtalkService
}

func NewPaymentService(
	cfg *config.Config,
	orderRepo PaymentOrderRepository,
	paymentCache PaymentCache,
	balanceService *BalanceService,
	bonusService *BonusService,
	promotionService *PromotionService,
	referralService *ReferralService,
	dingtalkService *DingtalkService,
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
		bonusService:     bonusService,
		promotionService: promotionService,
		referralService:  referralService,
		dingtalkService:  dingtalkService,
	}
}

// CreateOrder 创建支付订单
func (s *PaymentService) CreateOrder(ctx context.Context, req *CreatePaymentOrderRequest) (*PaymentOrder, error) {
	if s == nil || s.cfg == nil || !s.cfg.Enabled {
		return nil, infraerrors.ServiceUnavailable("PAYMENT_DISABLED", "Payment is not enabled. Please contact the administrator.")
	}
	if req == nil {
		return nil, infraerrors.BadRequest("PAYMENT_INVALID_REQUEST", "Invalid request.")
	}
	if req.AmountCNY < s.cfg.MinAmount || req.AmountCNY > s.cfg.MaxAmount {
		return nil, infraerrors.BadRequest(
			"PAYMENT_INVALID_AMOUNT",
			fmt.Sprintf("Amount must be between %.2f and %.2f CNY.", s.cfg.MinAmount, s.cfg.MaxAmount),
		)
	}

	if s.paymentCache != nil && s.cfg.MaxOrdersPerMinute > 0 {
		count, err := s.paymentCache.IncrementUserOrderCounter(ctx, req.UserID, time.Minute)
		if err == nil && count > s.cfg.MaxOrdersPerMinute {
			return nil, infraerrors.TooManyRequests("PAYMENT_RATE_LIMITED", "Too many orders. Please try again later.")
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
		return nil, infraerrors.ServiceUnavailable("PAYMENT_CREATE_ORDER_FAILED", "Failed to create order. Please try again later.").WithCause(err)
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
		return infraerrors.BadRequest("PAYMENT_ORDER_REQUIRED", "Order is required.")
	}
	return s.orderRepo.Update(ctx, order)
}

// MarkOrderFailed marks a pending order as failed and stores the error message.
func (s *PaymentService) MarkOrderFailed(ctx context.Context, orderNo string, reason string) (*PaymentOrder, error) {
	if s == nil {
		return nil, nil
	}
	log.Printf("[Payment Service] MarkOrderFailed called: order_no=%s, reason=%s", orderNo, reason)
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	log.Printf("[Payment Service] Order retrieved: order_no=%s, current_status=%s, user_id=%d, amount_usd=%.2f",
		order.OrderNo, order.Status, order.UserID, order.TotalUSD)
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

	log.Printf("[Payment Service] Order marked as failed: order_no=%s, user_id=%d, status=%s, callback_at=%v",
		order.OrderNo, order.UserID, order.Status, order.CallbackAt)
	s.notifyAsync("Payment Failed", fmt.Sprintf("**Order**: %s\n\n**Provider**: %s\n\n**Amount(CNY)**: %.2f\n\n**Reason**: %s", order.OrderNo, order.Provider, order.AmountCNY, reason))
	return order, nil
}

// MarkOrderCancelled marks a pending order as cancelled and stores the reason/message.
func (s *PaymentService) MarkOrderCancelled(ctx context.Context, orderNo string, reason string) (*PaymentOrder, error) {
	if s == nil {
		return nil, nil
	}
	log.Printf("[Payment Service] MarkOrderCancelled called: order_no=%s, reason=%s", orderNo, reason)
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	log.Printf("[Payment Service] Order retrieved: order_no=%s, current_status=%s, user_id=%d, amount_usd=%.2f",
		order.OrderNo, order.Status, order.UserID, order.TotalUSD)
	if order.Status != PaymentStatusPending {
		return order, nil
	}

	now := timePtr(time.Now())
	order.Status = PaymentStatusCancelled
	order.CallbackAt = now
	if reason != "" {
		order.CallbackData = reason
	}
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}

	log.Printf("[Payment Service] Order marked as cancelled: order_no=%s, user_id=%d, status=%s, callback_at=%v",
		order.OrderNo, order.UserID, order.Status, order.CallbackAt)
	s.notifyAsync("Payment Cancelled", fmt.Sprintf("**Order**: %s\n\n**Provider**: %s\n\n**Amount(CNY)**: %.2f\n\n**Reason**: %s", order.OrderNo, order.Provider, order.AmountCNY, reason))
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
	log.Printf("[Payment Service] MarkOrderPaid called: order_no=%s, trade_no=%s", orderNo, tradeNo)
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	log.Printf("[Payment Service] Order retrieved: order_no=%s, current_status=%s, user_id=%d, amount_usd=%.2f",
		order.OrderNo, order.Status, order.UserID, order.TotalUSD)
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

	// 使用BonusService处理赠送逻辑
	log.Printf("[Payment Service] Checking bonusService availability: bonusService_is_nil=%v", s.bonusService == nil)
	if s.bonusService != nil {
		log.Printf("[Payment Service] BonusService available, preparing bonus request for order_no=%s, user_id=%d, amount_usd=%.2f",
			order.OrderNo, order.UserID, order.AmountUSD)

		bonusReq := &BonusRequest{
			OrderID:   order.ID,
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			Username:  order.Username,
			AmountUSD: order.AmountUSD,
			AmountCNY: order.AmountCNY,
			Provider:  order.Provider,
		}

		log.Printf("[Payment Service] Calling BonusService.ProcessOrderBonus with req=%+v", bonusReq)
		bonusResult, err := s.bonusService.ProcessOrderBonus(ctx, bonusReq)
		log.Printf("[Payment Service] BonusService.ProcessOrderBonus returned: result=%+v, err=%v", bonusResult, err)
		if err != nil {
			log.Printf("[Payment Service] ERROR: Bonus processing failed: order_no=%s, user_id=%d, error=%v",
				order.OrderNo, order.UserID, err)
			// 发送DingTalk告警
			s.notifyAsync("Bonus Processing Failed", fmt.Sprintf(
				"**警告: 首充赠送发放失败**\n\n"+
					"**Order**: %s\n\n"+
					"**User ID**: %d\n\n"+
					"**Username**: %s\n\n"+
					"**Amount(USD)**: %.2f\n\n"+
					"**Error**: %v\n\n"+
					"**Action**: 需要管理员手动补发",
				order.OrderNo, order.UserID, order.Username, order.AmountUSD, err,
			))
		} else if bonusResult != nil && bonusResult.Applied {
			// 更新订单bonus信息
			order.BonusUSD = bonusResult.BonusAmount
			order.PromotionTier = &bonusResult.Tier
			order.PromotionUsed = true

			log.Printf("[Payment Service] Bonus applied successfully: order_no=%s, user_id=%d, amount=%.2f, tier=%d",
				order.OrderNo, order.UserID, bonusResult.BonusAmount, bonusResult.Tier)
		}
	} else {
		log.Printf("[Payment Service] WARNING: BonusService is nil, skipping bonus processing for order_no=%s", order.OrderNo)
	}

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

	// 处理邀请返利
	if s.referralService != nil {
		reward, err := s.referralService.ProcessInviteeRecharge(ctx, &ReferralRechargeRequest{
			InviteeID:          order.UserID,
			RechargeAmountUSD:  order.AmountUSD,
			RechargeRecordType: RechargeTypePayment,
		})
		if err == nil && reward != nil && reward.ShouldIssue {
			if s.balanceService != nil {
				activityOrder, _ := s.createActivityRechargeOrder(ctx, reward.ReferrerID, reward.RewardAmountUSD, "邀请奖励")
				relatedID := &order.OrderNo
				if activityOrder != nil {
					relatedID = &activityOrder.OrderNo
				}
				_, _ = s.balanceService.ApplyChange(ctx, BalanceChangeRequest{
					UserID:    reward.ReferrerID,
					Amount:    reward.RewardAmountUSD,
					Type:      RechargeTypeReferral,
					Operator:  "system",
					Remark:    "邀请奖励",
					RelatedID: relatedID,
				})
			}
			_ = s.referralService.MarkRewardIssued(ctx, order.UserID, reward.RewardAmountUSD)
		}
	}

	log.Printf("[Payment Service] Order payment completed: order_no=%s, user_id=%d, username=%s, amount_usd=%.2f, bonus_usd=%.2f, paid_at=%v, promotion_used=%v",
		order.OrderNo, order.UserID, order.Username, order.AmountUSD, order.BonusUSD, order.PaidAt, order.PromotionUsed)
	s.notifyAsync("Payment Paid", fmt.Sprintf("**Order**: %s\n\n**Provider**: %s\n\n**Amount(CNY)**: %.2f\n\n**Credits(USD)**: %.2f\n\n**Bonus(USD)**: %.2f\n\n**Total(USD)**: %.2f", order.OrderNo, order.Provider, order.AmountCNY, order.AmountUSD, order.BonusUSD, order.TotalUSD))
	return order, nil
}

func (s *PaymentService) createActivityRechargeOrder(ctx context.Context, userID int64, amountUSD float64, remark string) (*PaymentOrder, error) {
	if s == nil || s.orderRepo == nil {
		return nil, nil
	}
	if userID <= 0 || !(amountUSD > 0) {
		return nil, nil
	}
	exchangeRate := 1.0
	if s.cfg != nil && s.cfg.ExchangeRate > 0 {
		exchangeRate = s.cfg.ExchangeRate
	}
	now := time.Now()
	order := &PaymentOrder{
		OrderNo:       s.generateOrderNo(),
		UserID:        userID,
		Remark:        remark,
		AmountCNY:     0,
		AmountUSD:     amountUSD,
		BonusUSD:      0,
		TotalUSD:      amountUSD,
		ExchangeRate:  exchangeRate,
		DiscountRate:  1.0,
		Provider:      "activity",
		PaymentMethod: "system",
		Status:        PaymentStatusPaid,
		PaidAt:        timePtr(now),
		ExpireAt:      now,
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *PaymentService) notifyAsync(title string, text string) {
	if s == nil || s.dingtalkService == nil || !s.dingtalkService.Enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = s.dingtalkService.SendMarkdown(ctx, title, text)
		cancel()
	}()
}

// CancelExpiredOrders 更新过期订单
func (s *PaymentService) CancelExpiredOrders(ctx context.Context) (int64, error) {
	if s == nil {
		return 0, nil
	}
	return s.orderRepo.MarkExpired(ctx, time.Now())
}

func (s *PaymentService) generateOrderNo() string {
	prefix := "PO" // 默认前缀
	if s.cfg != nil && s.cfg.OrderPrefix != "" {
		prefix = s.cfg.OrderPrefix
	}
	now := time.Now().UTC()
	random := rand.Intn(1000000)
	return prefix + now.Format("20060102150405") + fmt.Sprintf("%06d", random)
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
