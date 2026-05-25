package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/payment"
)

const (
	PaymentTypeAdminAdjustment   = "admin_adjustment"
	PaymentTypeRedeemCode        = "redeem_code"
	PaymentTypePromoCode         = "promo_code"
	PaymentTypeAffiliateTransfer = "affiliate_transfer"
	PaymentTypeOAuthFirstBind    = "oauth_first_bind"
	PaymentTypeSignupGrant       = "signup_grant"
	PaymentTypeRefundRollback    = "refund_rollback"
)

var internalBalanceOrderPaymentTypes = map[string]struct{}{
	PaymentTypeAdminAdjustment:   {},
	PaymentTypeRedeemCode:        {},
	PaymentTypePromoCode:         {},
	PaymentTypeAffiliateTransfer: {},
	PaymentTypeOAuthFirstBind:    {},
	PaymentTypeSignupGrant:       {},
	PaymentTypeRefundRollback:    {},
}

type ctxKeySkipInternalBalanceOrder struct{}

func ContextSkipInternalBalanceOrder(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeySkipInternalBalanceOrder{}, true)
}

func internalBalanceOrderSkipped(ctx context.Context) bool {
	skipped, _ := ctx.Value(ctxKeySkipInternalBalanceOrder{}).(bool)
	return skipped
}

func IsInternalBalanceOrderPaymentType(paymentType string) bool {
	_, ok := internalBalanceOrderPaymentTypes[strings.TrimSpace(paymentType)]
	return ok
}

type InternalBalanceOrderInput struct {
	UserID      int64
	Amount      float64
	SourceType  string
	SourceRef   string
	PaymentType string
	Notes       string
	Operator    string
	CreatedAt   *time.Time
	Metadata    map[string]any
}

type InternalBalanceOrderRecorder interface {
	RecordInternalBalanceOrder(ctx context.Context, input InternalBalanceOrderInput) (*dbent.PaymentOrder, error)
}

type internalBalanceOrderRecorder struct {
	entClient *dbent.Client
	userRepo  UserRepository
}

func NewInternalBalanceOrderRecorder(entClient *dbent.Client, userRepo UserRepository) InternalBalanceOrderRecorder {
	return &internalBalanceOrderRecorder{entClient: entClient, userRepo: userRepo}
}

func (s *PaymentService) RecordInternalBalanceOrder(ctx context.Context, input InternalBalanceOrderInput) (*dbent.PaymentOrder, error) {
	if s == nil {
		return nil, nil
	}
	recorder := &internalBalanceOrderRecorder{entClient: s.entClient, userRepo: s.userRepo}
	return recorder.RecordInternalBalanceOrder(ctx, input)
}

func (r *internalBalanceOrderRecorder) RecordInternalBalanceOrder(ctx context.Context, input InternalBalanceOrderInput) (*dbent.PaymentOrder, error) {
	if internalBalanceOrderSkipped(ctx) {
		return nil, nil
	}
	if r == nil || r.entClient == nil || r.userRepo == nil {
		return nil, nil
	}
	if input.UserID <= 0 || input.Amount <= 0 || math.IsNaN(input.Amount) || math.IsInf(input.Amount, 0) {
		return nil, nil
	}

	sourceType := normalizeInternalOrderToken(input.SourceType)
	paymentType := normalizeInternalOrderToken(input.PaymentType)
	if paymentType == "" {
		paymentType = sourceType
	}
	if sourceType == "" || input.SourceRef == "" || !IsInternalBalanceOrderPaymentType(paymentType) {
		return nil, fmt.Errorf("invalid internal balance order source: type=%q payment_type=%q ref=%q", input.SourceType, input.PaymentType, input.SourceRef)
	}

	outTradeNo := internalBalanceOrderOutTradeNo(sourceType, input.SourceRef)
	client := r.entClient
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}

	existing, err := client.PaymentOrder.Query().Where(paymentorder.OutTradeNo(outTradeNo)).Only(ctx)
	if err == nil {
		return existing, nil
	}
	if err != nil && !dbent.IsNotFound(err) {
		return nil, fmt.Errorf("query internal balance order: %w", err)
	}

	user, err := r.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get internal balance order user: %w", err)
	}

	now := time.Now()
	if input.CreatedAt != nil && !input.CreatedAt.IsZero() {
		now = *input.CreatedAt
	}
	operator := strings.TrimSpace(input.Operator)
	if operator == "" {
		operator = "system"
	}

	snapshot := map[string]any{
		"schema_version": 1,
		"provider_key":   "internal",
		"source_type":    sourceType,
		"source_ref":     input.SourceRef,
		"payment_type":   paymentType,
	}
	if notes := strings.TrimSpace(input.Notes); notes != "" {
		snapshot["notes"] = notes
	}
	if operator != "" {
		snapshot["operator"] = operator
	}
	for k, v := range input.Metadata {
		if strings.TrimSpace(k) != "" {
			snapshot[k] = v
		}
	}

	order, err := client.PaymentOrder.Create().
		SetUserID(input.UserID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetNillableUserNotes(psNilIfEmpty(user.Notes)).
		SetAmount(roundInternalOrderAmount(input.Amount)).
		SetPayAmount(0).
		SetFeeRate(0).
		SetRechargeCode(internalBalanceOrderRechargeCode(sourceType, input.SourceRef)).
		SetOutTradeNo(outTradeNo).
		SetPaymentType(paymentType).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetPaidAt(now).
		SetCompletedAt(now).
		SetExpiresAt(now).
		SetClientIP("internal").
		SetSrcHost("internal").
		SetProviderSnapshot(snapshot).
		SetProviderKey("internal").
		Save(ctx)
	if err != nil {
		if dbent.IsConstraintError(err) {
			existing, lookupErr := client.PaymentOrder.Query().Where(paymentorder.OutTradeNo(outTradeNo)).Only(ctx)
			if lookupErr == nil {
				return existing, nil
			}
		}
		return nil, fmt.Errorf("create internal balance order: %w", err)
	}

	detail := map[string]any{
		"creditedAmount": order.Amount,
		"sourceType":     sourceType,
		"sourceRef":      input.SourceRef,
		"paymentType":    paymentType,
		"payAmount":      order.PayAmount,
	}
	if input.Notes != "" {
		detail["notes"] = input.Notes
	}
	if err := writePaymentAuditLog(ctx, client, order.ID, "INTERNAL_BALANCE_ORDER_CREATED", operator, detail); err != nil {
		return order, fmt.Errorf("create internal balance order audit log: %w", err)
	}

	return order, nil
}

func normalizeInternalOrderToken(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func internalBalanceOrderOutTradeNo(sourceType, sourceRef string) string {
	return "int_" + internalBalanceOrderDigest(sourceType, sourceRef)
}

func internalBalanceOrderRechargeCode(sourceType, sourceRef string) string {
	return "INT-" + strings.ToUpper(internalBalanceOrderDigest(sourceType, sourceRef)[:20])
}

func internalBalanceOrderDigest(sourceType, sourceRef string) string {
	sum := sha1.Sum([]byte(sourceType + ":" + sourceRef))
	return hex.EncodeToString(sum[:])
}

func roundInternalOrderAmount(v float64) float64 {
	rounded, err := strconv.ParseFloat(fmt.Sprintf("%.8f", v), 64)
	if err != nil {
		return v
	}
	return rounded
}

func writePaymentAuditLog(ctx context.Context, client *dbent.Client, orderID int64, action, operator string, detail map[string]any) error {
	if client == nil {
		return nil
	}
	dj, _ := json.Marshal(detail)
	_, err := client.PaymentAuditLog.Create().
		SetOrderID(strconv.FormatInt(orderID, 10)).
		SetAction(action).
		SetDetail(string(dj)).
		SetOperator(operator).
		Save(ctx)
	return err
}
