package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// PromotionApplyResult 表示活动计算结果
type PromotionApplyResult struct {
	Applied      bool
	Tier         int
	BonusPercent float64
	Bonus        float64
}

// PromotionService 管理新用户优惠活动
type PromotionService struct {
	cfg   *config.PromotionConfig
	repo  PromotionRepository
	cache PromotionCache
}

func NewPromotionService(cfg *config.Config, repo PromotionRepository, cache PromotionCache) *PromotionService {
	var promotionCfg *config.PromotionConfig
	if cfg != nil {
		promotionCfg = &cfg.Promotion
	}
	return &PromotionService{
		cfg:   promotionCfg,
		repo:  repo,
		cache: cache,
	}
}

// EnsureUserPromotion ensures a user has an active promotion record when:
// - promotion is enabled
// - user was created within the promotion window (based on registration time)
// This is used to support "old users" created shortly before the feature was enabled.
func (s *PromotionService) EnsureUserPromotion(ctx context.Context, userID int64) (*UserPromotion, error) {
	if s == nil || s.cfg == nil || !s.cfg.Enabled || s.repo == nil {
		return nil, nil
	}
	if userID <= 0 {
		return nil, nil
	}

	existing, err := s.GetUserPromotion(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	meta, err := s.repo.GetUserMetaForPromotion(ctx, userID)
	if err != nil || meta == nil || meta.CreatedAt.IsZero() {
		return nil, err
	}

	expire := meta.CreatedAt.Add(time.Duration(s.cfg.DurationHours) * time.Hour)
	now := time.Now()
	if now.After(expire) {
		return nil, nil
	}

	promotion := &UserPromotion{
		UserID:      userID,
		Username:    meta.Username,
		ActivatedAt: meta.CreatedAt,
		ExpireAt:    expire,
		Status:      PromotionStatusActive,
	}
	if err := s.repo.Create(ctx, promotion); err != nil {
		return nil, fmt.Errorf("create promotion: %w", err)
	}
	if s.cache != nil {
		ttl := time.Until(expire)
		if ttl > 0 {
			_ = s.cache.SetUserPromotion(ctx, promotion, ttl)
		}
	}
	return promotion, nil
}

// InitUserPromotion 初始化用户活动资格
func (s *PromotionService) InitUserPromotion(ctx context.Context, userID int64, username string) error {
	if s == nil || s.cfg == nil || !s.cfg.Enabled {
		return nil
	}

	existing, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get promotion: %w", err)
	}
	if existing != nil {
		return nil
	}

	now := time.Now()
	activatedAt := now
	if meta, err := s.repo.GetUserMetaForPromotion(ctx, userID); err == nil && meta != nil && !meta.CreatedAt.IsZero() {
		activatedAt = meta.CreatedAt
		if username == "" {
			username = meta.Username
		}
	}
	expire := activatedAt.Add(time.Duration(s.cfg.DurationHours) * time.Hour)
	promotion := &UserPromotion{
		UserID:      userID,
		Username:    username,
		ActivatedAt: activatedAt,
		ExpireAt:    expire,
		Status:      PromotionStatusActive,
	}

	if err := s.repo.Create(ctx, promotion); err != nil {
		return fmt.Errorf("create promotion: %w", err)
	}
	if s.cache != nil {
		_ = s.cache.SetUserPromotion(ctx, promotion, expire.Sub(now))
	}
	return nil
}

// GetUserPromotion 获取活动状态
func (s *PromotionService) GetUserPromotion(ctx context.Context, userID int64) (*UserPromotion, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	if s.cache != nil {
		if cached, err := s.cache.GetUserPromotion(ctx, userID); err == nil && cached != nil {
			return cached, nil
		}
	}

	promotion, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if promotion != nil && s.cache != nil {
		ttl := time.Until(promotion.ExpireAt)
		if ttl > 0 {
			_ = s.cache.SetUserPromotion(ctx, promotion, ttl)
		}
	}
	return promotion, nil
}

// ApplyPromotion 评估是否可以使用活动
func (s *PromotionService) ApplyPromotion(ctx context.Context, userID int64, amountUSD float64) (*PromotionApplyResult, error) {
	if s == nil || s.cfg == nil || !s.cfg.Enabled {
		return nil, nil
	}
	if amountUSD < s.cfg.MinAmount {
		return &PromotionApplyResult{Applied: false}, nil
	}

	promotion, err := s.GetUserPromotion(ctx, userID)
	if err != nil {
		return nil, err
	}
	if promotion == nil {
		promotion, err = s.EnsureUserPromotion(ctx, userID)
		if err != nil {
			return nil, err
		}
		if promotion == nil {
			return &PromotionApplyResult{Applied: false}, nil
		}
	}
	if promotion.Status != PromotionStatusActive {
		return &PromotionApplyResult{Applied: false}, nil
	}
	if time.Now().After(promotion.ExpireAt) {
		return &PromotionApplyResult{Applied: false}, nil
	}

	tier, percent := s.calculateTier(promotion.ActivatedAt)
	if tier < 0 || percent <= 0 {
		return &PromotionApplyResult{Applied: false}, nil
	}

	bonus := amountUSD * percent / 100.0
	return &PromotionApplyResult{
		Applied:      true,
		Tier:         tier,
		BonusPercent: percent,
		Bonus:        bonus,
	}, nil
}

func (s *PromotionService) calculateTier(activatedAt time.Time) (int, float64) {
	if s.cfg == nil {
		return -1, 0
	}
	elapsed := time.Since(activatedAt)
	elapsedHours := elapsed.Hours()
	for i, tier := range s.cfg.Tiers {
		if elapsedHours <= float64(tier.Hours) {
			return i, tier.BonusPercent
		}
	}
	return -1, 0
}

// MarkPromotionUsed 标记活动已使用并记录
func (s *PromotionService) MarkPromotionUsed(ctx context.Context, userID int64, tier int, amountUSD, bonusUSD float64) error {
	if s == nil || s.repo == nil {
		return nil
	}

	promotion, err := s.repo.GetByUserID(ctx, userID)
	if err != nil || promotion == nil {
		return err
	}

	now := time.Now()
	promotion.Status = PromotionStatusUsed
	promotion.UsedTier = &tier
	promotion.UsedAt = &now
	promotion.UsedAmount = &amountUSD
	promotion.BonusAmount = &bonusUSD

	if err := s.repo.Update(ctx, promotion); err != nil {
		return fmt.Errorf("update promotion: %w", err)
	}

	record := &PromotionRecord{
		UserID:   userID,
		Username: promotion.Username,
		Tier:     tier,
		Amount:   amountUSD,
		Bonus:    bonusUSD,
	}
	if err := s.repo.CreateRecord(ctx, record); err != nil {
		return fmt.Errorf("create promotion record: %w", err)
	}

	if s.cache != nil {
		_ = s.cache.InvalidateUserPromotion(ctx, userID)
	}
	return nil
}
