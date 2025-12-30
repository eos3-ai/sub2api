package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"gorm.io/gorm"
)

type rechargeRecordRepository struct {
	db *gorm.DB
}

func NewRechargeRecordRepository(db *gorm.DB) service.RechargeRecordRepository {
	return &rechargeRecordRepository{db: db}
}

func (r *rechargeRecordRepository) Create(ctx context.Context, record *service.RechargeRecord) error {
	m := rechargeRecordModelFromService(record)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return err
	}
	applyRechargeRecordModelToService(record, m)
	return nil
}

func (r *rechargeRecordRepository) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.RechargeRecord, *pagination.PaginationResult, error) {
	var records []rechargeRecordModel
	var total int64

	db := r.db.WithContext(ctx).Model(&rechargeRecordModel{}).
		Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	if err := db.Order("created_at DESC").Offset(params.Offset()).Limit(params.Limit()).Find(&records).Error; err != nil {
		return nil, nil, err
	}

	outRecords := make([]service.RechargeRecord, 0, len(records))
	for i := range records {
		outRecords = append(outRecords, *rechargeRecordModelToService(&records[i]))
	}
	return outRecords, paginationResultFromTotal(total, params), nil
}

func (r *rechargeRecordRepository) ListAll(ctx context.Context, params pagination.PaginationParams, filter service.RechargeRecordListFilter) ([]service.RechargeRecord, *pagination.PaginationResult, error) {
	var records []rechargeRecordModel
	var total int64

	db := r.db.WithContext(ctx).Model(&rechargeRecordModel{})

	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if filter.Type != "" {
		db = db.Where("type = ?", filter.Type)
	}
	if filter.RelatedID != "" {
		db = db.Where("related_id = ?", filter.RelatedID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	if err := db.Order("created_at DESC").Offset(params.Offset()).Limit(params.Limit()).Find(&records).Error; err != nil {
		return nil, nil, err
	}

	outRecords := make([]service.RechargeRecord, 0, len(records))
	for i := range records {
		outRecords = append(outRecords, *rechargeRecordModelToService(&records[i]))
	}
	return outRecords, paginationResultFromTotal(total, params), nil
}

type rechargeRecordModel struct {
	ID            int64      `gorm:"primaryKey"`
	UserID        int64      `gorm:"index;not null"`
	Amount        float64    `gorm:"type:decimal(20,8);not null"`
	Type          string     `gorm:"size:20;not null"`
	Operator      string     `gorm:"size:100"`
	Remark        string     `gorm:"size:500"`
	RelatedID     *string    `gorm:"size:100"`
	BalanceBefore float64    `gorm:"type:decimal(20,8);not null"`
	BalanceAfter  float64    `gorm:"type:decimal(20,8);not null"`
	CreatedAt     time.Time  `gorm:"not null"`
	User          *userModel `gorm:"foreignKey:UserID"`
}

func (rechargeRecordModel) TableName() string { return "recharge_records" }

func rechargeRecordModelToService(m *rechargeRecordModel) *service.RechargeRecord {
	if m == nil {
		return nil
	}
	return &service.RechargeRecord{
		ID:            m.ID,
		UserID:        m.UserID,
		Amount:        m.Amount,
		Type:          m.Type,
		Operator:      m.Operator,
		Remark:        m.Remark,
		RelatedID:     m.RelatedID,
		BalanceBefore: m.BalanceBefore,
		BalanceAfter:  m.BalanceAfter,
		CreatedAt:     m.CreatedAt,
	}
}

func rechargeRecordModelFromService(r *service.RechargeRecord) *rechargeRecordModel {
	if r == nil {
		return nil
	}
	return &rechargeRecordModel{
		ID:            r.ID,
		UserID:        r.UserID,
		Amount:        r.Amount,
		Type:          r.Type,
		Operator:      r.Operator,
		Remark:        r.Remark,
		RelatedID:     r.RelatedID,
		BalanceBefore: r.BalanceBefore,
		BalanceAfter:  r.BalanceAfter,
		CreatedAt:     r.CreatedAt,
	}
}

func applyRechargeRecordModelToService(dst *service.RechargeRecord, src *rechargeRecordModel) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
}
