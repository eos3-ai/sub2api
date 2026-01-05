package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type referralRepository struct {
	db *sql.DB
}

func NewReferralRepository(_ *ent.Client, sqlDB *sql.DB) service.ReferralRepository {
	return &referralRepository{db: sqlDB}
}

func (r *referralRepository) CreateCode(ctx context.Context, code *service.ReferralCode) error {
	if code == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	err := scanSingleRow(
		ctx,
		exec,
		`INSERT INTO referral_codes (user_id, code) VALUES ($1,$2) RETURNING id, created_at`,
		[]any{code.UserID, code.Code},
		&code.ID,
		&code.CreatedAt,
	)
	return err
}

func (r *referralRepository) GetCodeByUserID(ctx context.Context, userID int64) (*service.ReferralCode, error) {
	exec := sqlExecutorFromContext(ctx, r.db)
	var out service.ReferralCode
	err := scanSingleRow(
		ctx,
		exec,
		`SELECT id, user_id, code, created_at FROM referral_codes WHERE user_id = $1`,
		[]any{userID},
		&out.ID,
		&out.UserID,
		&out.Code,
		&out.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *referralRepository) GetCodeByCode(ctx context.Context, code string) (*service.ReferralCode, error) {
	exec := sqlExecutorFromContext(ctx, r.db)
	var out service.ReferralCode
	err := scanSingleRow(
		ctx,
		exec,
		`SELECT id, user_id, code, created_at FROM referral_codes WHERE code = $1`,
		[]any{code},
		&out.ID,
		&out.UserID,
		&out.Code,
		&out.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *referralRepository) CreateInvite(ctx context.Context, invite *service.ReferralInvite) error {
	if invite == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	var qualifiedAt sql.NullTime
	if invite.QualifiedAt != nil {
		qualifiedAt = sql.NullTime{Time: *invite.QualifiedAt, Valid: true}
	}
	var rewardIssuedAt sql.NullTime
	if invite.RewardIssuedAt != nil {
		rewardIssuedAt = sql.NullTime{Time: *invite.RewardIssuedAt, Valid: true}
	}

	err := scanSingleRow(
		ctx,
		exec,
		`
INSERT INTO referral_invites (
  invitee_id, invitee_username, referrer_id, referrer_username,
  total_recharge_usd, is_qualified, qualified_at, reward_issued, reward_issued_at, reward_amount_usd
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING id, created_at, updated_at
`,
		[]any{
			invite.InviteeID,
			nullIfEmpty(invite.InviteeUsername),
			invite.ReferrerID,
			nullIfEmpty(invite.ReferrerUsername),
			invite.TotalRechargeUSD,
			invite.IsQualified,
			qualifiedAt,
			invite.RewardIssued,
			rewardIssuedAt,
			invite.RewardAmountUSD,
		},
		&invite.ID,
		&invite.CreatedAt,
		&invite.UpdatedAt,
	)
	return err
}

func (r *referralRepository) GetInviteByInviteeID(ctx context.Context, inviteeID int64) (*service.ReferralInvite, error) {
	exec := sqlExecutorFromContext(ctx, r.db)
	var out service.ReferralInvite
	var inviteeUsername sql.NullString
	var referrerUsername sql.NullString
	var qualifiedAt sql.NullTime
	var rewardIssuedAt sql.NullTime

	err := scanSingleRow(
		ctx,
		exec,
		`
SELECT id, invitee_id, invitee_username, referrer_id, referrer_username,
       total_recharge_usd, is_qualified, qualified_at, reward_issued, reward_issued_at, reward_amount_usd,
       created_at, updated_at
FROM referral_invites
WHERE invitee_id = $1
`,
		[]any{inviteeID},
		&out.ID,
		&out.InviteeID,
		&inviteeUsername,
		&out.ReferrerID,
		&referrerUsername,
		&out.TotalRechargeUSD,
		&out.IsQualified,
		&qualifiedAt,
		&out.RewardIssued,
		&rewardIssuedAt,
		&out.RewardAmountUSD,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out.InviteeUsername = inviteeUsername.String
	out.ReferrerUsername = referrerUsername.String
	if qualifiedAt.Valid {
		t := qualifiedAt.Time
		out.QualifiedAt = &t
	}
	if rewardIssuedAt.Valid {
		t := rewardIssuedAt.Time
		out.RewardIssuedAt = &t
	}
	return &out, nil
}

func (r *referralRepository) ListInvitesByReferrer(ctx context.Context, referrerID int64, params pagination.PaginationParams) ([]service.ReferralInvite, *pagination.PaginationResult, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	var total int64
	if err := scanSingleRow(ctx, exec, `SELECT COUNT(*) FROM referral_invites WHERE referrer_id = $1`, []any{referrerID}, &total); err != nil {
		return nil, nil, err
	}

	rows, err := exec.QueryContext(
		ctx,
		`
SELECT id, invitee_id, invitee_username, referrer_id, referrer_username,
       total_recharge_usd, is_qualified, qualified_at, reward_issued, reward_issued_at, reward_amount_usd,
       created_at, updated_at
FROM referral_invites
WHERE referrer_id = $1
ORDER BY created_at DESC
OFFSET $2 LIMIT $3
`,
		referrerID,
		params.Offset(),
		params.Limit(),
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	out := make([]service.ReferralInvite, 0, params.Limit())
	for rows.Next() {
		var item service.ReferralInvite
		var inviteeUsername sql.NullString
		var referrerUsername sql.NullString
		var qualifiedAt sql.NullTime
		var rewardIssuedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.InviteeID,
			&inviteeUsername,
			&item.ReferrerID,
			&referrerUsername,
			&item.TotalRechargeUSD,
			&item.IsQualified,
			&qualifiedAt,
			&item.RewardIssued,
			&rewardIssuedAt,
			&item.RewardAmountUSD,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		item.InviteeUsername = inviteeUsername.String
		item.ReferrerUsername = referrerUsername.String
		if qualifiedAt.Valid {
			t := qualifiedAt.Time
			item.QualifiedAt = &t
		}
		if rewardIssuedAt.Valid {
			t := rewardIssuedAt.Time
			item.RewardIssuedAt = &t
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return out, paginationResultFromTotal(total, params), nil
}

func (r *referralRepository) UpdateInvite(ctx context.Context, invite *service.ReferralInvite) error {
	if invite == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	var qualifiedAt sql.NullTime
	if invite.QualifiedAt != nil {
		qualifiedAt = sql.NullTime{Time: *invite.QualifiedAt, Valid: true}
	}
	var rewardIssuedAt sql.NullTime
	if invite.RewardIssuedAt != nil {
		rewardIssuedAt = sql.NullTime{Time: *invite.RewardIssuedAt, Valid: true}
	}

	err := scanSingleRow(
		ctx,
		exec,
		`
UPDATE referral_invites
SET invitee_username=$2,
    referrer_username=$3,
    total_recharge_usd=$4,
    is_qualified=$5,
    qualified_at=$6,
    reward_issued=$7,
    reward_issued_at=$8,
    reward_amount_usd=$9,
    updated_at=NOW()
WHERE id=$1
RETURNING updated_at
`,
		[]any{
			invite.ID,
			nullIfEmpty(invite.InviteeUsername),
			nullIfEmpty(invite.ReferrerUsername),
			invite.TotalRechargeUSD,
			invite.IsQualified,
			qualifiedAt,
			invite.RewardIssued,
			rewardIssuedAt,
			invite.RewardAmountUSD,
		},
		&invite.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func (r *referralRepository) CountInvitesByReferrer(ctx context.Context, referrerID int64) (int64, error) {
	exec := sqlExecutorFromContext(ctx, r.db)
	var count int64
	if err := scanSingleRow(ctx, exec, `SELECT COUNT(*) FROM referral_invites WHERE referrer_id = $1`, []any{referrerID}, &count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *referralRepository) GetReferrerStats(ctx context.Context, referrerID int64) (*service.ReferralStats, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	var out service.ReferralStats
	err := scanSingleRow(
		ctx,
		exec,
		`
SELECT
  COUNT(*) as total_invites,
  COUNT(CASE WHEN is_qualified = true THEN 1 END) as qualified_invites,
  COUNT(CASE WHEN reward_issued = true THEN 1 END) as rewarded_invites,
  COALESCE(SUM(CASE WHEN reward_issued = true THEN reward_amount_usd ELSE 0 END), 0) as rewarded_usd
FROM referral_invites
WHERE referrer_id = $1
`,
		[]any{referrerID},
		&out.TotalInvites,
		&out.QualifiedInvites,
		&out.RewardedInvites,
		&out.RewardedUSD,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
