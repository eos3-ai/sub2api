package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"gorm.io/gorm"
)

type referralRepository struct {
	db *gorm.DB
}

func NewReferralRepository(db *gorm.DB) service.ReferralRepository {
	return &referralRepository{db: db}
}

func (r *referralRepository) CreateCode(ctx context.Context, code *service.ReferralCode) error {
	m := referralCodeModelFromService(code)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return err
	}
	applyReferralCodeModelToService(code, m)
	return nil
}

func (r *referralRepository) GetCodeByUserID(ctx context.Context, userID int64) (*service.ReferralCode, error) {
	var m referralCodeModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return referralCodeModelToService(&m), nil
}

func (r *referralRepository) GetCodeByCode(ctx context.Context, code string) (*service.ReferralCode, error) {
	var m referralCodeModel
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return referralCodeModelToService(&m), nil
}

func (r *referralRepository) CreateInvite(ctx context.Context, invite *service.ReferralInvite) error {
	m := referralInviteModelFromService(invite)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return err
	}
	applyReferralInviteModelToService(invite, m)
	return nil
}

func (r *referralRepository) GetInviteByInviteeID(ctx context.Context, inviteeID int64) (*service.ReferralInvite, error) {
	var m referralInviteModel
	err := r.db.WithContext(ctx).Where("invitee_id = ?", inviteeID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return referralInviteModelToService(&m), nil
}

func (r *referralRepository) ListInvitesByReferrer(ctx context.Context, referrerID int64, params pagination.PaginationParams) ([]service.ReferralInvite, *pagination.PaginationResult, error) {
	var invites []referralInviteModel
	var total int64

	db := r.db.WithContext(ctx).Model(&referralInviteModel{}).Where("referrer_id = ?", referrerID)

	if err := db.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	if err := db.Order("created_at DESC").Offset(params.Offset()).Limit(params.Limit()).Find(&invites).Error; err != nil {
		return nil, nil, err
	}

	outInvites := make([]service.ReferralInvite, 0, len(invites))
	for i := range invites {
		outInvites = append(outInvites, *referralInviteModelToService(&invites[i]))
	}
	return outInvites, paginationResultFromTotal(total, params), nil
}

func (r *referralRepository) UpdateInvite(ctx context.Context, invite *service.ReferralInvite) error {
	m := referralInviteModelFromService(invite)
	err := r.db.WithContext(ctx).Save(m).Error
	if err != nil {
		return err
	}
	applyReferralInviteModelToService(invite, m)
	return nil
}

func (r *referralRepository) CountInvitesByReferrer(ctx context.Context, referrerID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&referralInviteModel{}).Where("referrer_id = ?", referrerID).Count(&count).Error
	return count, err
}

func (r *referralRepository) GetReferrerStats(ctx context.Context, referrerID int64) (*service.ReferralStats, error) {
	type row struct {
		TotalInvites     int64   `gorm:"column:total_invites"`
		QualifiedInvites int64   `gorm:"column:qualified_invites"`
		RewardedInvites  int64   `gorm:"column:rewarded_invites"`
		RewardedUSD      float64 `gorm:"column:rewarded_usd"`
	}
	var out row
	err := r.db.WithContext(ctx).Model(&referralInviteModel{}).
		Select(`
			COUNT(*) as total_invites,
			COUNT(CASE WHEN is_qualified = true THEN 1 END) as qualified_invites,
			COUNT(CASE WHEN reward_issued = true THEN 1 END) as rewarded_invites,
			COALESCE(SUM(CASE WHEN reward_issued = true THEN reward_amount_usd ELSE 0 END), 0) as rewarded_usd
		`).
		Where("referrer_id = ?", referrerID).
		Scan(&out).Error
	if err != nil {
		return nil, err
	}
	return &service.ReferralStats{
		TotalInvites:     out.TotalInvites,
		QualifiedInvites: out.QualifiedInvites,
		RewardedInvites:  out.RewardedInvites,
		RewardedUSD:      out.RewardedUSD,
	}, nil
}

type referralCodeModel struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"uniqueIndex;not null"`
	Code      string    `gorm:"size:20;uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"not null"`
}

func (referralCodeModel) TableName() string { return "referral_codes" }

type referralInviteModel struct {
	ID               int64   `gorm:"primaryKey"`
	InviteeID        int64   `gorm:"uniqueIndex;not null"`
	InviteeUsername  string  `gorm:"size:100"`
	ReferrerID       int64   `gorm:"index;not null"`
	ReferrerUsername string  `gorm:"size:100"`
	TotalRechargeUSD float64 `gorm:"type:decimal(20,8);not null;default:0"`
	IsQualified      bool    `gorm:"not null;default:false"`
	QualifiedAt      *time.Time
	RewardIssued     bool `gorm:"not null;default:false"`
	RewardIssuedAt   *time.Time
	RewardAmountUSD  float64   `gorm:"type:decimal(20,8);not null;default:0"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
}

func (referralInviteModel) TableName() string { return "referral_invites" }

func referralCodeModelToService(m *referralCodeModel) *service.ReferralCode {
	if m == nil {
		return nil
	}
	return &service.ReferralCode{
		ID:        m.ID,
		UserID:    m.UserID,
		Code:      m.Code,
		CreatedAt: m.CreatedAt,
	}
}

func referralCodeModelFromService(c *service.ReferralCode) *referralCodeModel {
	if c == nil {
		return nil
	}
	return &referralCodeModel{
		ID:        c.ID,
		UserID:    c.UserID,
		Code:      c.Code,
		CreatedAt: c.CreatedAt,
	}
}

func applyReferralCodeModelToService(dst *service.ReferralCode, src *referralCodeModel) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
}

func referralInviteModelToService(m *referralInviteModel) *service.ReferralInvite {
	if m == nil {
		return nil
	}
	return &service.ReferralInvite{
		ID:               m.ID,
		InviteeID:        m.InviteeID,
		InviteeUsername:  m.InviteeUsername,
		ReferrerID:       m.ReferrerID,
		ReferrerUsername: m.ReferrerUsername,
		TotalRechargeUSD: m.TotalRechargeUSD,
		IsQualified:      m.IsQualified,
		QualifiedAt:      m.QualifiedAt,
		RewardIssued:     m.RewardIssued,
		RewardIssuedAt:   m.RewardIssuedAt,
		RewardAmountUSD:  m.RewardAmountUSD,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func referralInviteModelFromService(r *service.ReferralInvite) *referralInviteModel {
	if r == nil {
		return nil
	}
	return &referralInviteModel{
		ID:               r.ID,
		InviteeID:        r.InviteeID,
		InviteeUsername:  r.InviteeUsername,
		ReferrerID:       r.ReferrerID,
		ReferrerUsername: r.ReferrerUsername,
		TotalRechargeUSD: r.TotalRechargeUSD,
		IsQualified:      r.IsQualified,
		QualifiedAt:      r.QualifiedAt,
		RewardIssued:     r.RewardIssued,
		RewardIssuedAt:   r.RewardIssuedAt,
		RewardAmountUSD:  r.RewardAmountUSD,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func applyReferralInviteModelToService(dst *service.ReferralInvite, src *referralInviteModel) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}
