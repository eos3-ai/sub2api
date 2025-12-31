package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type ReferralStats struct {
	TotalInvites     int64   `json:"total_invites"`
	QualifiedInvites int64   `json:"qualified_invites"`
	RewardedInvites  int64   `json:"rewarded_invites"`
	RewardedUSD      float64 `json:"rewarded_usd"`
}

type ReferralInfoResponse struct {
	Enabled               bool          `json:"enabled"`
	Code                  string        `json:"code,omitempty"`
	InviteLink            string        `json:"invite_link,omitempty"`
	RewardUSD             float64       `json:"reward_usd,omitempty"`
	QualifiedRechargeCNY  float64       `json:"qualified_recharge_cny,omitempty"`
	QualifiedRechargeUSD  float64       `json:"qualified_recharge_usd,omitempty"`
	QualifiedRechargeTypes []string      `json:"qualified_recharge_types,omitempty"`
	MaxInviteesPerUser    int           `json:"max_invitees_per_user,omitempty"`
	Stats                 *ReferralStats `json:"stats,omitempty"`
}

type ReferralInvite struct {
	ID               int64      `json:"id"`
	InviteeID        int64      `json:"invitee_id"`
	InviteeUsername  string     `json:"invitee_username"`
	TotalRechargeUSD float64    `json:"total_recharge_usd"`
	IsQualified      bool       `json:"is_qualified"`
	QualifiedAt      *time.Time `json:"qualified_at,omitempty"`
	RewardIssued     bool       `json:"reward_issued"`
	RewardIssuedAt   *time.Time `json:"reward_issued_at,omitempty"`
	RewardAmountUSD  float64    `json:"reward_amount_usd"`
	CreatedAt        time.Time  `json:"created_at"`
}

func ReferralStatsFromService(s *service.ReferralStats) *ReferralStats {
	if s == nil {
		return nil
	}
	return &ReferralStats{
		TotalInvites:     s.TotalInvites,
		QualifiedInvites: s.QualifiedInvites,
		RewardedInvites:  s.RewardedInvites,
		RewardedUSD:      s.RewardedUSD,
	}
}

func ReferralInviteFromService(i *service.ReferralInvite) *ReferralInvite {
	if i == nil {
		return nil
	}
	return &ReferralInvite{
		ID:               i.ID,
		InviteeID:        i.InviteeID,
		InviteeUsername:  i.InviteeUsername,
		TotalRechargeUSD: i.TotalRechargeUSD,
		IsQualified:      i.IsQualified,
		QualifiedAt:      i.QualifiedAt,
		RewardIssued:     i.RewardIssued,
		RewardIssuedAt:   i.RewardIssuedAt,
		RewardAmountUSD:  i.RewardAmountUSD,
		CreatedAt:        i.CreatedAt,
	}
}

