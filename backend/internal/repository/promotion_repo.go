package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type promotionRepository struct {
	client *dbent.Client
	db     *sql.DB
}

func NewPromotionRepository(client *dbent.Client, sqlDB *sql.DB) service.PromotionRepository {
	return &promotionRepository{client: client, db: sqlDB}
}

func (r *promotionRepository) Create(ctx context.Context, promotion *service.UserPromotion) error {
	if promotion == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	var usedTier sql.NullInt64
	if promotion.UsedTier != nil {
		usedTier = sql.NullInt64{Int64: int64(*promotion.UsedTier), Valid: true}
	}
	var usedAt sql.NullTime
	if promotion.UsedAt != nil {
		usedAt = sql.NullTime{Time: *promotion.UsedAt, Valid: true}
	}
	var usedAmount sql.NullFloat64
	if promotion.UsedAmount != nil {
		usedAmount = sql.NullFloat64{Float64: *promotion.UsedAmount, Valid: true}
	}
	var bonusAmount sql.NullFloat64
	if promotion.BonusAmount != nil {
		bonusAmount = sql.NullFloat64{Float64: *promotion.BonusAmount, Valid: true}
	}

	err := scanSingleRow(
		ctx,
		exec,
		`
INSERT INTO user_promotions (
  user_id, username, activated_at, expire_at, status,
  used_tier, used_at, used_amount, bonus_amount
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING id, created_at, updated_at
`,
		[]any{
			promotion.UserID,
			nullIfEmpty(promotion.Username),
			promotion.ActivatedAt,
			promotion.ExpireAt,
			promotion.Status,
			usedTier,
			usedAt,
			usedAmount,
			bonusAmount,
		},
		&promotion.ID,
		&promotion.CreatedAt,
		&promotion.UpdatedAt,
	)
	return err
}

func (r *promotionRepository) GetByUserID(ctx context.Context, userID int64) (*service.UserPromotion, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	var out service.UserPromotion
	var username sql.NullString
	var usedTier sql.NullInt64
	var usedAt sql.NullTime
	var usedAmount sql.NullFloat64
	var bonusAmount sql.NullFloat64
	err := scanSingleRow(
		ctx,
		exec,
		`
SELECT id, user_id, username, activated_at, expire_at, status, used_tier, used_at, used_amount, bonus_amount, created_at, updated_at
FROM user_promotions
WHERE user_id = $1
`,
		[]any{userID},
		&out.ID,
		&out.UserID,
		&username,
		&out.ActivatedAt,
		&out.ExpireAt,
		&out.Status,
		&usedTier,
		&usedAt,
		&usedAmount,
		&bonusAmount,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out.Username = username.String
	if usedTier.Valid {
		v := int(usedTier.Int64)
		out.UsedTier = &v
	}
	if usedAt.Valid {
		t := usedAt.Time
		out.UsedAt = &t
	}
	if usedAmount.Valid {
		v := usedAmount.Float64
		out.UsedAmount = &v
	}
	if bonusAmount.Valid {
		v := bonusAmount.Float64
		out.BonusAmount = &v
	}
	return &out, nil
}

func (r *promotionRepository) Update(ctx context.Context, promotion *service.UserPromotion) error {
	if promotion == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	var usedTier sql.NullInt64
	if promotion.UsedTier != nil {
		usedTier = sql.NullInt64{Int64: int64(*promotion.UsedTier), Valid: true}
	}
	var usedAt sql.NullTime
	if promotion.UsedAt != nil {
		usedAt = sql.NullTime{Time: *promotion.UsedAt, Valid: true}
	}
	var usedAmount sql.NullFloat64
	if promotion.UsedAmount != nil {
		usedAmount = sql.NullFloat64{Float64: *promotion.UsedAmount, Valid: true}
	}
	var bonusAmount sql.NullFloat64
	if promotion.BonusAmount != nil {
		bonusAmount = sql.NullFloat64{Float64: *promotion.BonusAmount, Valid: true}
	}

	err := scanSingleRow(
		ctx,
		exec,
		`
UPDATE user_promotions
SET username=$2, activated_at=$3, expire_at=$4, status=$5,
    used_tier=$6, used_at=$7, used_amount=$8, bonus_amount=$9,
    updated_at=NOW()
WHERE user_id=$1
RETURNING updated_at
`,
		[]any{
			promotion.UserID,
			nullIfEmpty(promotion.Username),
			promotion.ActivatedAt,
			promotion.ExpireAt,
			promotion.Status,
			usedTier,
			usedAt,
			usedAmount,
			bonusAmount,
		},
		&promotion.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func (r *promotionRepository) CreateRecord(ctx context.Context, record *service.PromotionRecord) error {
	if record == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	err := scanSingleRow(
		ctx,
		exec,
		`
INSERT INTO promotion_records (user_id, username, tier, amount, bonus)
VALUES ($1,$2,$3,$4,$5)
RETURNING id, created_at
`,
		[]any{
			record.UserID,
			nullIfEmpty(record.Username),
			record.Tier,
			record.Amount,
			record.Bonus,
		},
		&record.ID,
		&record.CreatedAt,
	)
	return err
}

func (r *promotionRepository) GetUserMetaForPromotion(ctx context.Context, userID int64) (*service.PromotionUserMeta, error) {
	if userID <= 0 {
		return nil, nil
	}
	client := clientFromContext(ctx, r.client)
	u, err := client.User.Query().
		Select(dbuser.FieldID, dbuser.FieldEmail, dbuser.FieldUsername, dbuser.FieldCreatedAt).
		Where(dbuser.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	name := strings.TrimSpace(u.Username)
	if name == "" {
		name = strings.TrimSpace(u.Email)
	}

	return &service.PromotionUserMeta{
		UserID:    u.ID,
		Username:  name,
		CreatedAt: u.CreatedAt,
	}, nil
}
