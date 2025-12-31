package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// ReferralRewardResult 定义返利处理结果
type ReferralRewardResult struct {
	ShouldIssue     bool
	ReferrerID      int64
	RewardAmountUSD float64
}

// RecordInvitationRequest 定义邀请记录请求
type RecordInvitationRequest struct {
	InviteeID        int64
	InviteeUsername  string
	ReferrerID       int64
	ReferrerUsername string
}

// ReferralRechargeRequest 定义邀请用户充值请求
type ReferralRechargeRequest struct {
	InviteeID          int64
	RechargeAmountUSD  float64
	RechargeRecordType string
}

// ReferralService 管理邀请码和返利
type ReferralService struct {
	cfg        *config.ReferralConfig
	paymentCfg *config.PaymentConfig
	repo       ReferralRepository
	cache      ReferralCache
}

func NewReferralService(cfg *config.Config, repo ReferralRepository, cache ReferralCache) *ReferralService {
	var referralCfg *config.ReferralConfig
	var paymentCfg *config.PaymentConfig
	if cfg != nil {
		referralCfg = &cfg.Referral
		paymentCfg = &cfg.Payment
	}
	return &ReferralService{
		cfg:        referralCfg,
		paymentCfg: paymentCfg,
		repo:       repo,
		cache:      cache,
	}
}

// GetOrCreateUserCode 返回用户的邀请码，如不存在则自动生成
func (s *ReferralService) GetOrCreateUserCode(ctx context.Context, userID int64) (*ReferralCode, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	code, err := s.repo.GetCodeByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get referral code: %w", err)
	}
	if code != nil {
		return code, nil
	}

	newCode, err := s.generateCode()
	if err != nil {
		return nil, err
	}
	record := &ReferralCode{
		UserID: userID,
		Code:   newCode,
	}
	if err := s.repo.CreateCode(ctx, record); err != nil {
		return nil, fmt.Errorf("create referral code: %w", err)
	}
	return record, nil
}

func (s *ReferralService) generateCode() (string, error) {
	length := 8
	if s.cfg != nil && s.cfg.CodeLength > 0 {
		length = s.cfg.CodeLength
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate code: %w", err)
	}
	encoded := strings.ToUpper(hex.EncodeToString(bytes))
	if len(encoded) > length {
		encoded = encoded[:length]
	}
	return encoded, nil
}

// FindUserIDByCode 根据邀请码查找用户ID
func (s *ReferralService) FindUserIDByCode(ctx context.Context, code string) (int64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return 0, nil
	}

	if s.cache != nil {
		if uid, err := s.cache.GetUserIDByCode(ctx, code); err == nil && uid > 0 {
			return uid, nil
		}
	}

	record, err := s.repo.GetCodeByCode(ctx, code)
	if err != nil || record == nil {
		return 0, err
	}
	if s.cache != nil {
		_ = s.cache.SetUserIDByCode(ctx, code, record.UserID, 0)
	}
	return record.UserID, nil
}

func (s *ReferralService) GetReferrerStats(ctx context.Context, referrerID int64) (*ReferralStats, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	return s.repo.GetReferrerStats(ctx, referrerID)
}

// RecordInvitation 保存邀请关系
func (s *ReferralService) RecordInvitation(ctx context.Context, req RecordInvitationRequest) error {
	if s == nil || s.repo == nil {
		return nil
	}
	if s.cfg == nil || !s.cfg.Enabled {
		return nil
	}
	if req.InviteeID == 0 || req.ReferrerID == 0 {
		return fmt.Errorf("invalid invitation request")
	}
	if req.InviteeID == req.ReferrerID {
		return fmt.Errorf("cannot invite yourself")
	}

	if s.cfg.MaxInviteesPerUser > 0 {
		count, err := s.repo.CountInvitesByReferrer(ctx, req.ReferrerID)
		if err != nil {
			return err
		}
		if count >= int64(s.cfg.MaxInviteesPerUser) {
			return fmt.Errorf("invitees limit reached")
		}
	}

	existing, err := s.repo.GetInviteByInviteeID(ctx, req.InviteeID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	invite := &ReferralInvite{
		InviteeID:        req.InviteeID,
		InviteeUsername:  req.InviteeUsername,
		ReferrerID:       req.ReferrerID,
		ReferrerUsername: req.ReferrerUsername,
	}
	return s.repo.CreateInvite(ctx, invite)
}

// ProcessInviteeRecharge 处理被邀请用户充值逻辑
func (s *ReferralService) ProcessInviteeRecharge(ctx context.Context, req *ReferralRechargeRequest) (*ReferralRewardResult, error) {
	if s == nil || s.repo == nil || s.cfg == nil || !s.cfg.Enabled {
		return nil, nil
	}
	if req == nil || req.InviteeID == 0 {
		return nil, nil
	}

	if len(s.cfg.QualifiedRechargeTypes) > 0 {
		rt := strings.ToLower(strings.TrimSpace(req.RechargeRecordType))
		allowed := false
		for _, t := range s.cfg.QualifiedRechargeTypes {
			if strings.ToLower(strings.TrimSpace(t)) == rt {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, nil
		}
	}

	invite, err := s.repo.GetInviteByInviteeID(ctx, req.InviteeID)
	if err != nil || invite == nil {
		return nil, err
	}

	invite.TotalRechargeUSD += req.RechargeAmountUSD

	rewardResult := &ReferralRewardResult{
		ReferrerID: invite.ReferrerID,
	}

	if !invite.IsQualified && s.hasReachedThreshold(invite.TotalRechargeUSD) {
		now := time.Now()
		invite.IsQualified = true
		invite.QualifiedAt = &now
	}

	if invite.IsQualified && !invite.RewardIssued && s.cfg.RewardUSD > 0 {
		invite.RewardAmountUSD = s.cfg.RewardUSD
		rewardResult.ShouldIssue = true
		rewardResult.RewardAmountUSD = s.cfg.RewardUSD
	}

	if err := s.repo.UpdateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("update invite: %w", err)
	}

	if rewardResult.ShouldIssue {
		return rewardResult, nil
	}
	return nil, nil
}

func (s *ReferralService) hasReachedThreshold(totalRechargeUSD float64) bool {
	if s.cfg == nil {
		return false
	}
	thresholdUSD := s.cfg.QualifiedRechargeUSD
	if thresholdUSD <= 0 && s.paymentCfg != nil && s.paymentCfg.ExchangeRate > 0 {
		thresholdUSD = s.cfg.QualifiedRechargeCNY / s.paymentCfg.ExchangeRate
	}
	if thresholdUSD <= 0 {
		return false
	}
	return totalRechargeUSD >= thresholdUSD
}

// MarkRewardIssued 更新返利发放状态
func (s *ReferralService) MarkRewardIssued(ctx context.Context, inviteeID int64, rewardAmount float64) error {
	if s == nil || s.repo == nil {
		return nil
	}
	invite, err := s.repo.GetInviteByInviteeID(ctx, inviteeID)
	if err != nil || invite == nil {
		return err
	}
	if invite.RewardIssued {
		return nil
	}
	invite.RewardIssued = true
	now := time.Now()
	invite.RewardIssuedAt = &now
	invite.RewardAmountUSD = rewardAmount
	return s.repo.UpdateInvite(ctx, invite)
}

// ListInvites 查询邀请列表
func (s *ReferralService) ListInvites(ctx context.Context, referrerID int64, params pagination.PaginationParams) ([]ReferralInvite, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, nil
	}
	return s.repo.ListInvitesByReferrer(ctx, referrerID, params)
}
