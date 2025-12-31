package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"gorm.io/gorm"
)

type promotionRepository struct {
	db *gorm.DB
}

func NewPromotionRepository(db *gorm.DB) service.PromotionRepository {
	return &promotionRepository{db: db}
}

func (r *promotionRepository) Create(ctx context.Context, promotion *service.UserPromotion) error {
	m := userPromotionModelFromService(promotion)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return err
	}
	applyUserPromotionModelToService(promotion, m)
	return nil
}

func (r *promotionRepository) GetByUserID(ctx context.Context, userID int64) (*service.UserPromotion, error) {
	var m userPromotionModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return userPromotionModelToService(&m), nil
}

func (r *promotionRepository) Update(ctx context.Context, promotion *service.UserPromotion) error {
	m := userPromotionModelFromService(promotion)
	err := r.db.WithContext(ctx).Save(m).Error
	if err != nil {
		return err
	}
	applyUserPromotionModelToService(promotion, m)
	return nil
}

func (r *promotionRepository) CreateRecord(ctx context.Context, record *service.PromotionRecord) error {
	m := promotionRecordModelFromService(record)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return err
	}
	record.ID = m.ID
	record.CreatedAt = m.CreatedAt
	return nil
}

func (r *promotionRepository) GetUserMetaForPromotion(ctx context.Context, userID int64) (*service.PromotionUserMeta, error) {
	if r == nil || r.db == nil || userID <= 0 {
		return nil, nil
	}

	type row struct {
		ID        int64     `gorm:"column:id"`
		Email     string    `gorm:"column:email"`
		Username  string    `gorm:"column:username"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var out row
	err := r.db.WithContext(ctx).
		Model(&userModel{}).
		Select("id", "email", "username", "created_at").
		Where("id = ?", userID).
		Take(&out).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	name := strings.TrimSpace(out.Username)
	if name == "" {
		name = strings.TrimSpace(out.Email)
	}

	return &service.PromotionUserMeta{
		UserID:    out.ID,
		Username:  name,
		CreatedAt: out.CreatedAt,
	}, nil
}

type userPromotionModel struct {
	ID          int64     `gorm:"primaryKey"`
	UserID      int64     `gorm:"uniqueIndex;not null"`
	Username    string    `gorm:"size:100"`
	ActivatedAt time.Time `gorm:"not null"`
	ExpireAt    time.Time `gorm:"not null"`
	Status      string    `gorm:"size:20;not null"`
	UsedTier    *int      `gorm:"type:int"`
	UsedAt      *time.Time
	UsedAmount  *float64   `gorm:"type:decimal(20,8)"`
	BonusAmount *float64   `gorm:"type:decimal(20,8)"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
	User        *userModel `gorm:"foreignKey:UserID"`
}

func (userPromotionModel) TableName() string { return "user_promotions" }

type promotionRecordModel struct {
	ID        int64      `gorm:"primaryKey"`
	UserID    int64      `gorm:"index;not null"`
	Username  string     `gorm:"size:100"`
	Tier      int        `gorm:"not null"`
	Amount    float64    `gorm:"type:decimal(20,8);not null"`
	Bonus     float64    `gorm:"type:decimal(20,8);not null"`
	CreatedAt time.Time  `gorm:"not null"`
	User      *userModel `gorm:"foreignKey:UserID"`
}

func (promotionRecordModel) TableName() string { return "promotion_records" }

func userPromotionModelToService(m *userPromotionModel) *service.UserPromotion {
	if m == nil {
		return nil
	}
	return &service.UserPromotion{
		ID:          m.ID,
		UserID:      m.UserID,
		Username:    m.Username,
		ActivatedAt: m.ActivatedAt,
		ExpireAt:    m.ExpireAt,
		Status:      m.Status,
		UsedTier:    m.UsedTier,
		UsedAt:      m.UsedAt,
		UsedAmount:  m.UsedAmount,
		BonusAmount: m.BonusAmount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func userPromotionModelFromService(p *service.UserPromotion) *userPromotionModel {
	if p == nil {
		return nil
	}
	return &userPromotionModel{
		ID:          p.ID,
		UserID:      p.UserID,
		Username:    p.Username,
		ActivatedAt: p.ActivatedAt,
		ExpireAt:    p.ExpireAt,
		Status:      p.Status,
		UsedTier:    p.UsedTier,
		UsedAt:      p.UsedAt,
		UsedAmount:  p.UsedAmount,
		BonusAmount: p.BonusAmount,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func applyUserPromotionModelToService(dst *service.UserPromotion, src *userPromotionModel) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}

func promotionRecordModelFromService(r *service.PromotionRecord) *promotionRecordModel {
	if r == nil {
		return nil
	}
	return &promotionRecordModel{
		ID:       r.ID,
		UserID:   r.UserID,
		Username: r.Username,
		Tier:     r.Tier,
		Amount:   r.Amount,
		Bonus:    r.Bonus,
	}
}

func promotionRecordModelToService(m *promotionRecordModel) *service.PromotionRecord {
	if m == nil {
		return nil
	}
	return &service.PromotionRecord{
		ID:        m.ID,
		UserID:    m.UserID,
		Username:  m.Username,
		Tier:      m.Tier,
		Amount:    m.Amount,
		Bonus:     m.Bonus,
		CreatedAt: m.CreatedAt,
	}
}
