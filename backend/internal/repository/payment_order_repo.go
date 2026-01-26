package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentOrderRepository struct {
	db *sql.DB
}

func NewPaymentOrderRepository(_ *ent.Client, sqlDB *sql.DB) service.PaymentOrderRepository {
	return &paymentOrderRepository{db: sqlDB}
}

func (r *paymentOrderRepository) Create(ctx context.Context, order *service.PaymentOrder) error {
	if order == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	var tradeNo sql.NullString
	if order.TradeNo != nil && *order.TradeNo != "" {
		tradeNo = sql.NullString{String: *order.TradeNo, Valid: true}
	}
	var paidAt sql.NullTime
	if order.PaidAt != nil {
		paidAt = sql.NullTime{Time: *order.PaidAt, Valid: true}
	}
	var promotionTier sql.NullInt64
	if order.PromotionTier != nil {
		promotionTier = sql.NullInt64{Int64: int64(*order.PromotionTier), Valid: true}
	}
	var callbackAt sql.NullTime
	if order.CallbackAt != nil {
		callbackAt = sql.NullTime{Time: *order.CallbackAt, Valid: true}
	}

	err := scanSingleRow(
		ctx,
		exec,
		`
INSERT INTO payment_orders (
  order_no, trade_no, user_id, username, remark,
  amount_cny, amount_usd, bonus_usd, total_usd, exchange_rate,
  provider, channel, payment_method, payment_url,
  status, paid_at, expire_at,
  promotion_tier, promotion_used,
  callback_data, callback_at,
  client_ip, user_agent
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
RETURNING id, created_at, updated_at
`,
		[]any{
			order.OrderNo,
			tradeNo,
			order.UserID,
			nullIfEmpty(order.Username),
			order.Remark,
			order.AmountCNY,
			order.AmountUSD,
			order.BonusUSD,
			order.TotalUSD,
			order.ExchangeRate,
			order.Provider,
			nullIfEmpty(order.Channel),
			nullIfEmpty(order.PaymentMethod),
			nullIfEmpty(order.PaymentURL),
			order.Status,
			paidAt,
			order.ExpireAt,
			promotionTier,
			order.PromotionUsed,
			nullIfEmpty(order.CallbackData),
			callbackAt,
			nullIfEmpty(order.ClientIP),
			nullIfEmpty(order.UserAgent),
		},
		&order.ID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	return err
}

func (r *paymentOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.PaymentOrder, error) {
	exec := sqlExecutorFromContext(ctx, r.db)
	return r.getOne(ctx, exec, `WHERE order_no = $1`, []any{orderNo})
}

func (r *paymentOrderRepository) GetByOrderNoForUpdate(ctx context.Context, orderNo string) (*service.PaymentOrder, error) {
	exec := sqlExecutorFromContext(ctx, r.db)
	return r.getOne(ctx, exec, `WHERE order_no = $1 FOR UPDATE`, []any{orderNo})
}

func (r *paymentOrderRepository) GetByTradeNo(ctx context.Context, tradeNo string) (*service.PaymentOrder, error) {
	exec := sqlExecutorFromContext(ctx, r.db)
	return r.getOne(ctx, exec, `WHERE trade_no = $1`, []any{tradeNo})
}

func (r *paymentOrderRepository) Update(ctx context.Context, order *service.PaymentOrder) error {
	if order == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	var tradeNo sql.NullString
	if order.TradeNo != nil && *order.TradeNo != "" {
		tradeNo = sql.NullString{String: *order.TradeNo, Valid: true}
	}
	var paidAt sql.NullTime
	if order.PaidAt != nil {
		paidAt = sql.NullTime{Time: *order.PaidAt, Valid: true}
	}
	var promotionTier sql.NullInt64
	if order.PromotionTier != nil {
		promotionTier = sql.NullInt64{Int64: int64(*order.PromotionTier), Valid: true}
	}
	var callbackAt sql.NullTime
	if order.CallbackAt != nil {
		callbackAt = sql.NullTime{Time: *order.CallbackAt, Valid: true}
	}

	err := scanSingleRow(
		ctx,
		exec,
		`
UPDATE payment_orders
SET trade_no=$2,
    username=$3,
    remark=$4,
    amount_cny=$5,
    amount_usd=$6,
    bonus_usd=$7,
    total_usd=$8,
    exchange_rate=$9,
    provider=$10,
    channel=$11,
    payment_method=$12,
    payment_url=$13,
    status=$14,
    paid_at=$15,
    expire_at=$16,
    promotion_tier=$17,
    promotion_used=$18,
    callback_data=$19,
    callback_at=$20,
    client_ip=$21,
    user_agent=$22,
    updated_at=NOW()
WHERE order_no=$1
RETURNING updated_at
`,
		[]any{
			order.OrderNo,
			tradeNo,
			nullIfEmpty(order.Username),
			order.Remark,
			order.AmountCNY,
			order.AmountUSD,
			order.BonusUSD,
			order.TotalUSD,
			order.ExchangeRate,
			order.Provider,
			nullIfEmpty(order.Channel),
			nullIfEmpty(order.PaymentMethod),
			nullIfEmpty(order.PaymentURL),
			order.Status,
			paidAt,
			order.ExpireAt,
			promotionTier,
			order.PromotionUsed,
			nullIfEmpty(order.CallbackData),
			callbackAt,
			nullIfEmpty(order.ClientIP),
			nullIfEmpty(order.UserAgent),
		},
		&order.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func (r *paymentOrderRepository) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	filter := service.PaymentOrderFilter{UserID: &userID, Status: status}
	return r.List(ctx, params, filter)
}

func (r *paymentOrderRepository) List(ctx context.Context, params pagination.PaginationParams, filter service.PaymentOrderFilter) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	where := "WHERE 1=1"
	args := make([]any, 0, 8)
	add := func(cond string, v any) {
		args = append(args, v)
		where += fmt.Sprintf(" AND %s = $%d", cond, len(args))
	}

	if filter.UserID != nil {
		add("user_id", *filter.UserID)
	}
	if filter.Status != "" {
		add("status", filter.Status)
	}
	if filter.OrderType != "" {
		switch filter.OrderType {
		case "admin_recharge":
			add("provider", "admin")
		case "activity_recharge":
			add("provider", "activity")
		case "online_recharge":
			where += " AND provider NOT IN ('admin','activity')"
		default:
			return nil, nil, fmt.Errorf("unsupported order_type: %s", filter.OrderType)
		}
	}
	if filter.Provider != "" {
		add("provider", filter.Provider)
	}
	if filter.From != nil {
		args = append(args, *filter.From)
		where += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if filter.To != nil {
		args = append(args, *filter.To)
		where += fmt.Sprintf(" AND created_at <= $%d", len(args))
	}

	var total int64
	if err := scanSingleRow(ctx, exec, `SELECT COUNT(*) FROM payment_orders `+where, args, &total); err != nil {
		return nil, nil, err
	}

	offsetPos := len(args) + 1
	limitPos := len(args) + 2
	query := fmt.Sprintf(
		`
SELECT id, order_no, trade_no, user_id, username, remark,
       amount_cny, amount_usd, bonus_usd, total_usd, exchange_rate,
       provider, channel, payment_method, payment_url,
       status, paid_at, expire_at,
       promotion_tier, promotion_used,
       callback_data, callback_at,
       client_ip, user_agent,
       created_at, updated_at
FROM payment_orders
%s
ORDER BY created_at DESC
OFFSET $%d LIMIT $%d
`,
		where,
		offsetPos,
		limitPos,
	)
	queryArgs := append(args, params.Offset(), params.Limit())
	rows, err := exec.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	out := make([]service.PaymentOrder, 0, params.Limit())
	for rows.Next() {
		item, err := scanPaymentOrder(rows)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return out, paginationResultFromTotal(total, params), nil
}

func (r *paymentOrderRepository) Summary(ctx context.Context, filter service.PaymentOrderFilter) (totalUSD float64, amountCNY float64, err error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	where := "WHERE 1=1"
	args := make([]any, 0, 8)
	add := func(cond string, v any) {
		args = append(args, v)
		where += fmt.Sprintf(" AND %s = $%d", cond, len(args))
	}

	if filter.UserID != nil {
		add("user_id", *filter.UserID)
	}
	if filter.Status != "" {
		add("status", filter.Status)
	}
	if filter.OrderType != "" {
		switch filter.OrderType {
		case "admin_recharge":
			add("provider", "admin")
		case "activity_recharge":
			add("provider", "activity")
		case "online_recharge":
			where += " AND provider NOT IN ('admin','activity')"
		default:
			return 0, 0, fmt.Errorf("unsupported order_type: %s", filter.OrderType)
		}
	}
	if filter.Provider != "" {
		add("provider", filter.Provider)
	}
	if filter.From != nil {
		args = append(args, *filter.From)
		where += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if filter.To != nil {
		args = append(args, *filter.To)
		where += fmt.Sprintf(" AND created_at <= $%d", len(args))
	}

	query := `SELECT COALESCE(SUM(total_usd), 0), COALESCE(SUM(amount_cny), 0) FROM payment_orders ` + where
	if err := scanSingleRow(ctx, exec, query, args, &totalUSD, &amountCNY); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return totalUSD, amountCNY, nil
}

func (r *paymentOrderRepository) MarkExpired(ctx context.Context, now time.Time) (int64, error) {
	exec := sqlExecutorFromContext(ctx, r.db)
	res, err := exec.ExecContext(
		ctx,
		`UPDATE payment_orders SET status=$1, updated_at=$2 WHERE status=$3 AND expire_at < $4`,
		service.PaymentStatusExpired,
		now,
		service.PaymentStatusPending,
		now,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *paymentOrderRepository) getOne(ctx context.Context, exec sqlExecutor, where string, args []any) (*service.PaymentOrder, error) {
	query := `
SELECT id, order_no, trade_no, user_id, username, remark,
       amount_cny, amount_usd, bonus_usd, total_usd, exchange_rate,
       provider, channel, payment_method, payment_url,
       status, paid_at, expire_at,
       promotion_tier, promotion_used,
       callback_data, callback_at,
       client_ip, user_agent,
       created_at, updated_at
FROM payment_orders
` + where

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	item, err := scanPaymentOrder(rows)
	if err != nil {
		return nil, err
	}
	return item, nil
}

type sqlRowScanner interface {
	Scan(dest ...any) error
}

func scanPaymentOrder(s sqlRowScanner) (*service.PaymentOrder, error) {
	var o service.PaymentOrder
	var tradeNo sql.NullString
	var username sql.NullString
	var channel sql.NullString
	var paymentMethod sql.NullString
	var paymentURL sql.NullString
	var paidAt sql.NullTime
	var promotionTier sql.NullInt64
	var callbackData sql.NullString
	var callbackAt sql.NullTime
	var clientIP sql.NullString
	var userAgent sql.NullString

	if err := s.Scan(
		&o.ID,
		&o.OrderNo,
		&tradeNo,
		&o.UserID,
		&username,
		&o.Remark,
		&o.AmountCNY,
		&o.AmountUSD,
		&o.BonusUSD,
		&o.TotalUSD,
		&o.ExchangeRate,
		&o.Provider,
		&channel,
		&paymentMethod,
		&paymentURL,
		&o.Status,
		&paidAt,
		&o.ExpireAt,
		&promotionTier,
		&o.PromotionUsed,
		&callbackData,
		&callbackAt,
		&clientIP,
		&userAgent,
		&o.CreatedAt,
		&o.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if tradeNo.Valid {
		v := tradeNo.String
		o.TradeNo = &v
	}
	o.Username = username.String
	o.Channel = channel.String
	o.PaymentMethod = paymentMethod.String
	o.PaymentURL = paymentURL.String
	if paidAt.Valid {
		t := paidAt.Time
		o.PaidAt = &t
	}
	if promotionTier.Valid {
		v := int(promotionTier.Int64)
		o.PromotionTier = &v
	}
	o.CallbackData = callbackData.String
	if callbackAt.Valid {
		t := callbackAt.Time
		o.CallbackAt = &t
	}
	o.ClientIP = clientIP.String
	o.UserAgent = userAgent.String

	o.DiscountRate = inferredDiscountRate(o.AmountUSD, o.TotalUSD)
	return &o, nil
}

func inferredDiscountRate(amountUSD, totalUSD float64) float64 {
	if totalUSD <= 0 {
		return 1
	}
	v := amountUSD / totalUSD
	if !isFinite(v) || v <= 0 {
		return 1
	}
	if v > 1 {
		return 1
	}
	return v
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
