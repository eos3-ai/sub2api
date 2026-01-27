package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type invoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(_ *ent.Client, sqlDB *sql.DB) service.InvoiceRepository {
	return &invoiceRepository{db: sqlDB}
}

func (r *invoiceRepository) ListEligibleOrders(ctx context.Context, userID int64, params pagination.PaginationParams, from, to *time.Time) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	where := "WHERE po.user_id = $1 AND po.status = $2 AND po.amount_cny > 0 AND ioi.id IS NULL"
	args := []any{userID, service.PaymentStatusPaid}
	if from != nil {
		args = append(args, *from)
		where += fmt.Sprintf(" AND po.created_at >= $%d", len(args))
	}
	if to != nil {
		args = append(args, *to)
		where += fmt.Sprintf(" AND po.created_at <= $%d", len(args))
	}

	var total int64
	countQuery := `
SELECT COUNT(*)
FROM payment_orders po
LEFT JOIN invoice_order_items ioi
  ON ioi.payment_order_id = po.id AND ioi.active = TRUE
` + where
	if err := scanSingleRow(ctx, exec, countQuery, args, &total); err != nil {
		return nil, nil, err
	}

	offsetPos := len(args) + 1
	limitPos := len(args) + 2
	listQuery := fmt.Sprintf(
		`
SELECT po.id, po.order_no, po.trade_no, po.user_id, po.username, po.remark,
       po.amount_cny, po.amount_usd, po.bonus_usd, po.total_usd, po.exchange_rate,
       po.provider, po.channel, po.payment_method, po.payment_url,
       po.status, po.paid_at, po.expire_at,
       po.promotion_tier, po.promotion_used,
       po.callback_data, po.callback_at,
       po.client_ip, po.user_agent,
       po.created_at, po.updated_at
FROM payment_orders po
LEFT JOIN invoice_order_items ioi
  ON ioi.payment_order_id = po.id AND ioi.active = TRUE
%s
ORDER BY po.created_at DESC
OFFSET $%d LIMIT $%d
`,
		where,
		offsetPos,
		limitPos,
	)

	queryArgs := append(args, params.Offset(), params.Limit())
	rows, err := exec.QueryContext(ctx, listQuery, queryArgs...)
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

func (r *invoiceRepository) GetEligibleOrdersByOrderNos(ctx context.Context, userID int64, orderNos []string) ([]service.PaymentOrder, error) {
	if len(orderNos) == 0 {
		return nil, nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	query := `
SELECT po.id, po.order_no, po.trade_no, po.user_id, po.username, po.remark,
       po.amount_cny, po.amount_usd, po.bonus_usd, po.total_usd, po.exchange_rate,
       po.provider, po.channel, po.payment_method, po.payment_url,
       po.status, po.paid_at, po.expire_at,
       po.promotion_tier, po.promotion_used,
       po.callback_data, po.callback_at,
       po.client_ip, po.user_agent,
       po.created_at, po.updated_at
FROM payment_orders po
LEFT JOIN invoice_order_items ioi
  ON ioi.payment_order_id = po.id AND ioi.active = TRUE
WHERE po.user_id = $1
  AND po.order_no = ANY($2)
  AND po.status = $3
  AND po.amount_cny > 0
  AND ioi.id IS NULL
ORDER BY po.created_at DESC
`
	rows, err := exec.QueryContext(ctx, query, userID, pq.Array(orderNos), service.PaymentStatusPaid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]service.PaymentOrder, 0, len(orderNos))
	for rows.Next() {
		item, err := scanPaymentOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *invoiceRepository) CreateInvoiceRequest(ctx context.Context, req *service.InvoiceRequest, orders []service.PaymentOrder) error {
	if req == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	err := scanSingleRow(
		ctx,
		exec,
		`
INSERT INTO invoice_requests (
  invoice_request_no, user_id, status,
  invoice_type, buyer_type, invoice_title, tax_no,
  buyer_address, buyer_phone, buyer_bank_name, buyer_bank_account,
  receiver_email, receiver_phone,
  invoice_item_name, remark,
  amount_cny_total, total_usd_total
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
RETURNING id, created_at, updated_at
`,
		[]any{
			req.InvoiceRequestNo,
			req.UserID,
			req.Status,
			req.InvoiceType,
			req.BuyerType,
			req.InvoiceTitle,
			req.TaxNo,
			req.BuyerAddress,
			req.BuyerPhone,
			req.BuyerBankName,
			req.BuyerBankAccount,
			req.ReceiverEmail,
			req.ReceiverPhone,
			req.InvoiceItemName,
			req.Remark,
			req.AmountCNYTotal,
			req.TotalUSDTotal,
		},
		&req.ID,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return err
	}

	for i := range orders {
		o := orders[i]
		if _, err := exec.ExecContext(
			ctx,
			`INSERT INTO invoice_order_items (invoice_request_id, payment_order_id, order_no, amount_cny, total_usd, active) VALUES ($1,$2,$3,$4,$5,TRUE)`,
			req.ID,
			o.ID,
			o.OrderNo,
			o.AmountCNY,
			o.TotalUSD,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *invoiceRepository) UpdateInvoiceRequest(ctx context.Context, req *service.InvoiceRequest) error {
	if req == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	var reviewedBy sql.NullInt64
	if req.ReviewedBy != nil {
		reviewedBy = sql.NullInt64{Int64: *req.ReviewedBy, Valid: true}
	}
	var reviewedAt sql.NullTime
	if req.ReviewedAt != nil {
		reviewedAt = sql.NullTime{Time: *req.ReviewedAt, Valid: true}
	}
	var issuedBy sql.NullInt64
	if req.IssuedBy != nil {
		issuedBy = sql.NullInt64{Int64: *req.IssuedBy, Valid: true}
	}
	var issuedAt sql.NullTime
	if req.IssuedAt != nil {
		issuedAt = sql.NullTime{Time: *req.IssuedAt, Valid: true}
	}
	var invoiceDate sql.NullTime
	if req.InvoiceDate != nil {
		invoiceDate = sql.NullTime{Time: *req.InvoiceDate, Valid: true}
	}

	err := scanSingleRow(
		ctx,
		exec,
		`
UPDATE invoice_requests
SET status=$2,
    reviewed_by=$3,
    reviewed_at=$4,
    reject_reason=$5,
    issued_by=$6,
    issued_at=$7,
    invoice_number=$8,
    invoice_date=$9,
    invoice_pdf_url=$10,
    updated_at=NOW()
WHERE id=$1
RETURNING updated_at
`,
		[]any{
			req.ID,
			req.Status,
			reviewedBy,
			reviewedAt,
			req.RejectReason,
			issuedBy,
			issuedAt,
			req.InvoiceNumber,
			invoiceDate,
			req.InvoicePDFURL,
		},
		&req.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func (r *invoiceRepository) ListInvoiceRequests(ctx context.Context, params pagination.PaginationParams, filter service.InvoiceRequestListFilter) ([]service.InvoiceRequest, *pagination.PaginationResult, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	where := "WHERE 1=1"
	args := make([]any, 0, 8)
	addEq := func(col string, v any) {
		args = append(args, v)
		where += fmt.Sprintf(" AND %s = $%d", col, len(args))
	}

	if filter.UserID != nil {
		addEq("user_id", *filter.UserID)
	}
	if filter.Status != "" {
		addEq("status", filter.Status)
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
	if err := scanSingleRow(ctx, exec, `SELECT COUNT(*) FROM invoice_requests `+where, args, &total); err != nil {
		return nil, nil, err
	}

	offsetPos := len(args) + 1
	limitPos := len(args) + 2
	query := fmt.Sprintf(
		`
SELECT id, invoice_request_no, user_id, status,
       invoice_type, buyer_type, invoice_title, tax_no,
       buyer_address, buyer_phone, buyer_bank_name, buyer_bank_account,
       receiver_email, receiver_phone,
       invoice_item_name, remark,
       amount_cny_total, total_usd_total,
       reviewed_by, reviewed_at, reject_reason,
       issued_by, issued_at, invoice_number, invoice_date, invoice_pdf_url,
       created_at, updated_at
FROM invoice_requests
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

	out := make([]service.InvoiceRequest, 0, params.Limit())
	for rows.Next() {
		item, err := scanInvoiceRequest(rows)
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

func (r *invoiceRepository) GetInvoiceRequestByID(ctx context.Context, id int64) (*service.InvoiceRequest, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	rows, err := exec.QueryContext(
		ctx,
		`
SELECT id, invoice_request_no, user_id, status,
       invoice_type, buyer_type, invoice_title, tax_no,
       buyer_address, buyer_phone, buyer_bank_name, buyer_bank_account,
       receiver_email, receiver_phone,
       invoice_item_name, remark,
       amount_cny_total, total_usd_total,
       reviewed_by, reviewed_at, reject_reason,
       issued_by, issued_at, invoice_number, invoice_date, invoice_pdf_url,
       created_at, updated_at
FROM invoice_requests
WHERE id = $1
`,
		id,
	)
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
	item, err := scanInvoiceRequest(rows)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *invoiceRepository) ListInvoiceOrderItems(ctx context.Context, invoiceRequestID int64) ([]service.InvoiceOrderItem, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	rows, err := exec.QueryContext(
		ctx,
		`
SELECT i.id, i.invoice_request_id, i.payment_order_id, i.order_no,
       i.amount_cny, i.total_usd, i.active, i.created_at,
       po.paid_at
FROM invoice_order_items i
LEFT JOIN payment_orders po ON po.id = i.payment_order_id
WHERE i.invoice_request_id = $1
ORDER BY i.id ASC
`,
		invoiceRequestID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]service.InvoiceOrderItem, 0, 8)
	for rows.Next() {
		var item service.InvoiceOrderItem
		var paidAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.InvoiceRequestID,
			&item.PaymentOrderID,
			&item.OrderNo,
			&item.AmountCNY,
			&item.TotalUSD,
			&item.Active,
			&item.CreatedAt,
			&paidAt,
		); err != nil {
			return nil, err
		}
		if paidAt.Valid {
			t := paidAt.Time
			item.PaidAt = &t
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *invoiceRepository) SetInvoiceOrderItemsActive(ctx context.Context, invoiceRequestID int64, active bool) error {
	exec := sqlExecutorFromContext(ctx, r.db)
	_, err := exec.ExecContext(ctx, `UPDATE invoice_order_items SET active = $2 WHERE invoice_request_id = $1`, invoiceRequestID, active)
	return err
}

func (r *invoiceRepository) ListInvoiceOrderNosByRequestIDs(ctx context.Context, invoiceRequestIDs []int64) (map[int64][]string, error) {
	out := map[int64][]string{}
	if len(invoiceRequestIDs) == 0 {
		return out, nil
	}

	exec := sqlExecutorFromContext(ctx, r.db)
	rows, err := exec.QueryContext(
		ctx,
		`
SELECT invoice_request_id, order_no
FROM invoice_order_items
WHERE invoice_request_id = ANY($1)
ORDER BY invoice_request_id ASC, id ASC
`,
		pq.Array(invoiceRequestIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reqID int64
		var orderNo string
		if err := rows.Scan(&reqID, &orderNo); err != nil {
			return nil, err
		}
		out[reqID] = append(out[reqID], orderNo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *invoiceRepository) GetInvoiceProfile(ctx context.Context, userID int64) (*service.InvoiceProfile, error) {
	exec := sqlExecutorFromContext(ctx, r.db)

	rows, err := exec.QueryContext(
		ctx,
		`
SELECT id, user_id,
       invoice_type, buyer_type, invoice_title, tax_no,
       buyer_address, buyer_phone, buyer_bank_name, buyer_bank_account,
       receiver_email, receiver_phone,
       invoice_item_name, remark,
       created_at, updated_at
FROM invoice_profiles
WHERE user_id = $1
`,
		userID,
	)
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

	var p service.InvoiceProfile
	if err := rows.Scan(
		&p.ID,
		&p.UserID,
		&p.InvoiceType,
		&p.BuyerType,
		&p.InvoiceTitle,
		&p.TaxNo,
		&p.BuyerAddress,
		&p.BuyerPhone,
		&p.BuyerBankName,
		&p.BuyerBankAccount,
		&p.ReceiverEmail,
		&p.ReceiverPhone,
		&p.InvoiceItemName,
		&p.Remark,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *invoiceRepository) UpsertInvoiceProfile(ctx context.Context, profile *service.InvoiceProfile) error {
	if profile == nil {
		return nil
	}
	exec := sqlExecutorFromContext(ctx, r.db)

	err := scanSingleRow(
		ctx,
		exec,
		`
INSERT INTO invoice_profiles (
  user_id,
  invoice_type, buyer_type, invoice_title, tax_no,
  buyer_address, buyer_phone, buyer_bank_name, buyer_bank_account,
  receiver_email, receiver_phone,
  invoice_item_name, remark
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
ON CONFLICT (user_id) DO UPDATE SET
  invoice_type=EXCLUDED.invoice_type,
  buyer_type=EXCLUDED.buyer_type,
  invoice_title=EXCLUDED.invoice_title,
  tax_no=EXCLUDED.tax_no,
  buyer_address=EXCLUDED.buyer_address,
  buyer_phone=EXCLUDED.buyer_phone,
  buyer_bank_name=EXCLUDED.buyer_bank_name,
  buyer_bank_account=EXCLUDED.buyer_bank_account,
  receiver_email=EXCLUDED.receiver_email,
  receiver_phone=EXCLUDED.receiver_phone,
  invoice_item_name=EXCLUDED.invoice_item_name,
  remark=EXCLUDED.remark,
  updated_at=NOW()
RETURNING id, created_at, updated_at
`,
		[]any{
			profile.UserID,
			profile.InvoiceType,
			profile.BuyerType,
			profile.InvoiceTitle,
			profile.TaxNo,
			profile.BuyerAddress,
			profile.BuyerPhone,
			profile.BuyerBankName,
			profile.BuyerBankAccount,
			profile.ReceiverEmail,
			profile.ReceiverPhone,
			profile.InvoiceItemName,
			profile.Remark,
		},
		&profile.ID,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	return err
}

type sqlInvoiceRequestScanner interface {
	Scan(dest ...any) error
}

func scanInvoiceRequest(s sqlInvoiceRequestScanner) (*service.InvoiceRequest, error) {
	var req service.InvoiceRequest
	var reviewedBy sql.NullInt64
	var reviewedAt sql.NullTime
	var issuedBy sql.NullInt64
	var issuedAt sql.NullTime
	var invoiceDate sql.NullTime

	if err := s.Scan(
		&req.ID,
		&req.InvoiceRequestNo,
		&req.UserID,
		&req.Status,
		&req.InvoiceType,
		&req.BuyerType,
		&req.InvoiceTitle,
		&req.TaxNo,
		&req.BuyerAddress,
		&req.BuyerPhone,
		&req.BuyerBankName,
		&req.BuyerBankAccount,
		&req.ReceiverEmail,
		&req.ReceiverPhone,
		&req.InvoiceItemName,
		&req.Remark,
		&req.AmountCNYTotal,
		&req.TotalUSDTotal,
		&reviewedBy,
		&reviewedAt,
		&req.RejectReason,
		&issuedBy,
		&issuedAt,
		&req.InvoiceNumber,
		&invoiceDate,
		&req.InvoicePDFURL,
		&req.CreatedAt,
		&req.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if reviewedBy.Valid {
		v := reviewedBy.Int64
		req.ReviewedBy = &v
	}
	if reviewedAt.Valid {
		t := reviewedAt.Time
		req.ReviewedAt = &t
	}
	if issuedBy.Valid {
		v := issuedBy.Int64
		req.IssuedBy = &v
	}
	if issuedAt.Valid {
		t := issuedAt.Time
		req.IssuedAt = &t
	}
	if invoiceDate.Valid {
		t := invoiceDate.Time
		req.InvoiceDate = &t
	}

	return &req, nil
}
