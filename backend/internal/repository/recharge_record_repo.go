package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type rechargeRecordRepository struct {
	db *sql.DB
}

func NewRechargeRecordRepository(_ *ent.Client, sqlDB *sql.DB) service.RechargeRecordRepository {
	return &rechargeRecordRepository{db: sqlDB}
}

func (r *rechargeRecordRepository) Create(ctx context.Context, record *service.RechargeRecord) error {
	if record == nil {
		return nil
	}

	exec := sqlExecutorFromContext(ctx, r.db)
	var relatedID sql.NullString
	if record.RelatedID != nil && *record.RelatedID != "" {
		relatedID = sql.NullString{String: *record.RelatedID, Valid: true}
	}

	err := scanSingleRow(
		ctx,
		exec,
		`
INSERT INTO recharge_records (user_id, amount, type, operator, remark, related_id, balance_before, balance_after)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING id, created_at
`,
		[]any{
			record.UserID,
			record.Amount,
			record.Type,
			nullIfEmpty(record.Operator),
			nullIfEmpty(record.Remark),
			relatedID,
			record.BalanceBefore,
			record.BalanceAfter,
		},
		&record.ID,
		&record.CreatedAt,
	)
	return err
}

func (r *rechargeRecordRepository) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.RechargeRecord, *pagination.PaginationResult, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	var total int64
	if err := scanSingleRow(ctx, exec, `SELECT COUNT(*) FROM recharge_records WHERE user_id = $1`, []any{userID}, &total); err != nil {
		return nil, nil, err
	}

	rows, err := exec.QueryContext(
		ctx,
		`
SELECT id, user_id, amount, type, operator, remark, related_id, balance_before, balance_after, created_at
FROM recharge_records
WHERE user_id = $1
ORDER BY created_at DESC
OFFSET $2 LIMIT $3
`,
		userID,
		params.Offset(),
		params.Limit(),
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	out := make([]service.RechargeRecord, 0, params.Limit())
	for rows.Next() {
		var item service.RechargeRecord
		var operator sql.NullString
		var remark sql.NullString
		var related sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Amount,
			&item.Type,
			&operator,
			&remark,
			&related,
			&item.BalanceBefore,
			&item.BalanceAfter,
			&item.CreatedAt,
		); err != nil {
			return nil, nil, err
		}
		item.Operator = operator.String
		item.Remark = remark.String
		if related.Valid {
			v := related.String
			item.RelatedID = &v
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return out, paginationResultFromTotal(total, params), nil
}

func (r *rechargeRecordRepository) ListAll(ctx context.Context, params pagination.PaginationParams, filter service.RechargeRecordListFilter) ([]service.RechargeRecord, *pagination.PaginationResult, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	where := "WHERE 1=1"
	args := make([]any, 0, 6)
	if filter.UserID != nil {
		args = append(args, *filter.UserID)
		where += fmt.Sprintf(" AND user_id = $%d", len(args))
	}
	if filter.Type != "" {
		args = append(args, filter.Type)
		where += fmt.Sprintf(" AND type = $%d", len(args))
	}
	if filter.RelatedID != "" {
		args = append(args, filter.RelatedID)
		where += fmt.Sprintf(" AND related_id = $%d", len(args))
	}

	var total int64
	if err := scanSingleRow(ctx, exec, `SELECT COUNT(*) FROM recharge_records `+where, args, &total); err != nil {
		return nil, nil, err
	}

	offsetPos := len(args) + 1
	limitPos := len(args) + 2
	query := fmt.Sprintf(
		`
SELECT id, user_id, amount, type, operator, remark, related_id, balance_before, balance_after, created_at
FROM recharge_records
%s
ORDER BY created_at DESC
OFFSET $%d LIMIT $%d
`,
		where,
		offsetPos,
		limitPos,
	)
	queryArgs := append(args, params.Offset(), params.Limit())
	rows, err := exec.QueryContext(
		ctx,
		query,
		queryArgs...,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	out := make([]service.RechargeRecord, 0, params.Limit())
	for rows.Next() {
		var item service.RechargeRecord
		var operator sql.NullString
		var remark sql.NullString
		var related sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Amount,
			&item.Type,
			&operator,
			&remark,
			&related,
			&item.BalanceBefore,
			&item.BalanceAfter,
			&item.CreatedAt,
		); err != nil {
			return nil, nil, err
		}
		item.Operator = operator.String
		item.Remark = remark.String
		if related.Valid {
			v := related.String
			item.RelatedID = &v
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return out, paginationResultFromTotal(total, params), nil
}

func nullIfEmpty(v string) sql.NullString {
	if v == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: v, Valid: true}
}
