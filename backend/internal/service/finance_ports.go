package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// RechargeRecordListFilter 定义充值记录的筛选条件
type RechargeRecordListFilter struct {
	UserID    *int64
	Type      string
	RelatedID string
}

// RechargeRecordRepository 定义充值记录数据访问接口
type RechargeRecordRepository interface {
	Create(ctx context.Context, record *RechargeRecord) error
	ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]RechargeRecord, *pagination.PaginationResult, error)
	ListAll(ctx context.Context, params pagination.PaginationParams, filter RechargeRecordListFilter) ([]RechargeRecord, *pagination.PaginationResult, error)
}

// PromotionRepository 定义活动数据访问接口
type PromotionRepository interface {
	Create(ctx context.Context, promotion *UserPromotion) error
	GetByUserID(ctx context.Context, userID int64) (*UserPromotion, error)
	Update(ctx context.Context, promotion *UserPromotion) error
	CreateRecord(ctx context.Context, record *PromotionRecord) error
	// GetUserMetaForPromotion fetches user registration time and display name for promotion initialization.
	GetUserMetaForPromotion(ctx context.Context, userID int64) (*PromotionUserMeta, error)
}

type PromotionUserMeta struct {
	UserID    int64
	Username  string
	CreatedAt time.Time
}

// PromotionCache 定义活动缓存接口
type PromotionCache interface {
	GetUserPromotion(ctx context.Context, userID int64) (*UserPromotion, error)
	SetUserPromotion(ctx context.Context, promotion *UserPromotion, ttl time.Duration) error
	InvalidateUserPromotion(ctx context.Context, userID int64) error
}

// ReferralRepository 定义邀请码和邀请关系的数据访问接口
type ReferralRepository interface {
	CreateCode(ctx context.Context, code *ReferralCode) error
	GetCodeByUserID(ctx context.Context, userID int64) (*ReferralCode, error)
	GetCodeByCode(ctx context.Context, code string) (*ReferralCode, error)

	CreateInvite(ctx context.Context, invite *ReferralInvite) error
	GetInviteByInviteeID(ctx context.Context, inviteeID int64) (*ReferralInvite, error)
	ListInvitesByReferrer(ctx context.Context, referrerID int64, params pagination.PaginationParams) ([]ReferralInvite, *pagination.PaginationResult, error)
	UpdateInvite(ctx context.Context, invite *ReferralInvite) error

	CountInvitesByReferrer(ctx context.Context, referrerID int64) (int64, error)
	GetReferrerStats(ctx context.Context, referrerID int64) (*ReferralStats, error)
}

type ReferralStats struct {
	TotalInvites     int64   `json:"total_invites"`
	QualifiedInvites int64   `json:"qualified_invites"`
	RewardedInvites  int64   `json:"rewarded_invites"`
	RewardedUSD      float64 `json:"rewarded_usd"`
}

// ReferralCache 定义邀请码缓存接口
type ReferralCache interface {
	GetUserIDByCode(ctx context.Context, code string) (int64, error)
	SetUserIDByCode(ctx context.Context, code string, userID int64, ttl time.Duration) error
	InvalidateCode(ctx context.Context, code string) error
}

// PaymentOrderFilter 定义订单全局查询的筛选条件
type PaymentOrderFilter struct {
	UserID   *int64
	Status   string
	Provider string
	From     *time.Time
	To       *time.Time
}

// PaymentOrderRepository 定义支付订单数据访问接口
type PaymentOrderRepository interface {
	Create(ctx context.Context, order *PaymentOrder) error
	GetByOrderNo(ctx context.Context, orderNo string) (*PaymentOrder, error)
	GetByTradeNo(ctx context.Context, tradeNo string) (*PaymentOrder, error)
	Update(ctx context.Context, order *PaymentOrder) error

	ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]PaymentOrder, *pagination.PaginationResult, error)
	List(ctx context.Context, params pagination.PaginationParams, filter PaymentOrderFilter) ([]PaymentOrder, *pagination.PaginationResult, error)

	MarkExpired(ctx context.Context, now time.Time) (int64, error)
}

// PaymentCache 定义支付相关缓存接口
type PaymentCache interface {
	IncrementUserOrderCounter(ctx context.Context, userID int64, ttl time.Duration) (int, error)
	GetUserOrderCounter(ctx context.Context, userID int64) (int, error)
	ResetUserOrderCounter(ctx context.Context, userID int64) error

	SetPaymentURL(ctx context.Context, orderNo, url string, ttl time.Duration) error
	GetPaymentURL(ctx context.Context, orderNo string) (string, error)
	DeletePaymentURL(ctx context.Context, orderNo string) error
}
