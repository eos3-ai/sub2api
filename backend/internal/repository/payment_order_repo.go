package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"gorm.io/gorm"
)

type paymentOrderRepository struct {
	db *gorm.DB
}

func NewPaymentOrderRepository(db *gorm.DB) service.PaymentOrderRepository {
	return &paymentOrderRepository{db: db}
}

func (r *paymentOrderRepository) Create(ctx context.Context, order *service.PaymentOrder) error {
	m := paymentOrderModelFromService(order)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return err
	}
	applyPaymentOrderModelToService(order, m)
	return nil
}

func (r *paymentOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.PaymentOrder, error) {
	var m paymentOrderModel
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return paymentOrderModelToService(&m), nil
}

func (r *paymentOrderRepository) GetByTradeNo(ctx context.Context, tradeNo string) (*service.PaymentOrder, error) {
	var m paymentOrderModel
	err := r.db.WithContext(ctx).Where("trade_no = ?", tradeNo).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return paymentOrderModelToService(&m), nil
}

func (r *paymentOrderRepository) Update(ctx context.Context, order *service.PaymentOrder) error {
	m := paymentOrderModelFromService(order)
	err := r.db.WithContext(ctx).Save(m).Error
	if err != nil {
		return err
	}
	applyPaymentOrderModelToService(order, m)
	return nil
}

func (r *paymentOrderRepository) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	var orders []paymentOrderModel
	var total int64

	db := r.db.WithContext(ctx).Model(&paymentOrderModel{}).Where("user_id = ?", userID)
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	if err := db.Order("created_at DESC").Offset(params.Offset()).Limit(params.Limit()).Find(&orders).Error; err != nil {
		return nil, nil, err
	}

	outOrders := make([]service.PaymentOrder, 0, len(orders))
	for i := range orders {
		outOrders = append(outOrders, *paymentOrderModelToService(&orders[i]))
	}
	return outOrders, paginationResultFromTotal(total, params), nil
}

func (r *paymentOrderRepository) List(ctx context.Context, params pagination.PaginationParams, filter service.PaymentOrderFilter) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	var orders []paymentOrderModel
	var total int64

	db := r.db.WithContext(ctx).Model(&paymentOrderModel{})

	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.Provider != "" {
		db = db.Where("provider = ?", filter.Provider)
	}
	if filter.From != nil {
		db = db.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		db = db.Where("created_at <= ?", *filter.To)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	if err := db.Order("created_at DESC").Offset(params.Offset()).Limit(params.Limit()).Find(&orders).Error; err != nil {
		return nil, nil, err
	}

	outOrders := make([]service.PaymentOrder, 0, len(orders))
	for i := range orders {
		outOrders = append(outOrders, *paymentOrderModelToService(&orders[i]))
	}
	return outOrders, paginationResultFromTotal(total, params), nil
}

func (r *paymentOrderRepository) MarkExpired(ctx context.Context, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Model(&paymentOrderModel{}).
		Where("status = ? AND expire_at < ?", service.PaymentStatusPending, now).
		Updates(map[string]any{
			"status":     service.PaymentStatusExpired,
			"updated_at": now,
		})
	return result.RowsAffected, result.Error
}

type paymentOrderModel struct {
	ID            int64   `gorm:"primaryKey"`
	OrderNo       string  `gorm:"size:50;uniqueIndex;not null"`
	TradeNo       *string `gorm:"size:100;uniqueIndex"`
	UserID        int64   `gorm:"index;not null"`
	Username      string  `gorm:"size:100"`
	AmountCNY     float64 `gorm:"type:decimal(20,2);not null"`
	AmountUSD     float64 `gorm:"type:decimal(20,8);not null"`
	BonusUSD      float64 `gorm:"type:decimal(20,8);not null;default:0"`
	TotalUSD      float64 `gorm:"type:decimal(20,8);not null"`
	ExchangeRate  float64 `gorm:"type:decimal(10,4);not null"`
	DiscountRate  float64 `gorm:"type:decimal(10,4);not null;default:1"`
	Provider      string  `gorm:"size:20;not null"`
	PaymentMethod string  `gorm:"size:50"`
	PaymentURL    string  `gorm:"size:1000"`
	Status        string  `gorm:"size:20;not null"`
	PaidAt        *time.Time
	ExpireAt      time.Time `gorm:"not null"`
	PromotionTier *int
	PromotionUsed bool
	CallbackData  string `gorm:"type:text"`
	CallbackAt    *time.Time
	ClientIP      string    `gorm:"size:50"`
	UserAgent     string    `gorm:"size:500"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (paymentOrderModel) TableName() string { return "payment_orders" }

func paymentOrderModelToService(m *paymentOrderModel) *service.PaymentOrder {
	if m == nil {
		return nil
	}
	return &service.PaymentOrder{
		ID:            m.ID,
		OrderNo:       m.OrderNo,
		TradeNo:       m.TradeNo,
		UserID:        m.UserID,
		Username:      m.Username,
		AmountCNY:     m.AmountCNY,
		AmountUSD:     m.AmountUSD,
		BonusUSD:      m.BonusUSD,
		TotalUSD:      m.TotalUSD,
		ExchangeRate:  m.ExchangeRate,
		DiscountRate:  m.DiscountRate,
		Provider:      m.Provider,
		PaymentMethod: m.PaymentMethod,
		PaymentURL:    m.PaymentURL,
		Status:        m.Status,
		PaidAt:        m.PaidAt,
		ExpireAt:      m.ExpireAt,
		PromotionTier: m.PromotionTier,
		PromotionUsed: m.PromotionUsed,
		CallbackData:  m.CallbackData,
		CallbackAt:    m.CallbackAt,
		ClientIP:      m.ClientIP,
		UserAgent:     m.UserAgent,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func paymentOrderModelFromService(o *service.PaymentOrder) *paymentOrderModel {
	if o == nil {
		return nil
	}
	return &paymentOrderModel{
		ID:            o.ID,
		OrderNo:       o.OrderNo,
		TradeNo:       o.TradeNo,
		UserID:        o.UserID,
		Username:      o.Username,
		AmountCNY:     o.AmountCNY,
		AmountUSD:     o.AmountUSD,
		BonusUSD:      o.BonusUSD,
		TotalUSD:      o.TotalUSD,
		ExchangeRate:  o.ExchangeRate,
		DiscountRate:  o.DiscountRate,
		Provider:      o.Provider,
		PaymentMethod: o.PaymentMethod,
		PaymentURL:    o.PaymentURL,
		Status:        o.Status,
		PaidAt:        o.PaidAt,
		ExpireAt:      o.ExpireAt,
		PromotionTier: o.PromotionTier,
		PromotionUsed: o.PromotionUsed,
		CallbackData:  o.CallbackData,
		CallbackAt:    o.CallbackAt,
		ClientIP:      o.ClientIP,
		UserAgent:     o.UserAgent,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
	}
}

func applyPaymentOrderModelToService(dst *service.PaymentOrder, src *paymentOrderModel) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
	dst.DiscountRate = src.DiscountRate
}
