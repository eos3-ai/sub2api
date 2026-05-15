//go:build unit

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestPaymentOrderLegacyCompat_StatusFilterAndResponse(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderCompatTestClient(t)

	user := client.User.Create().
		SetEmail("compat@example.com").
		SetPasswordHash("hash").
		SetUsername("compat").
		SaveX(ctx)

	createCompatOrder(t, client, user.ID, "legacy-paid", "paid", time.Now().Add(-2*time.Hour))
	createCompatOrder(t, client, user.ID, "new-paid", "PAID", time.Now().Add(-1*time.Hour))
	createCompatOrder(t, client, user.ID, "pending", "pending", time.Now())

	svc := &PaymentService{entClient: client}
	orders, total, err := svc.GetUserOrders(ctx, user.ID, OrderListParams{
		Page:     1,
		PageSize: 10,
		Status:   "PAID",
	})
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, orders, 2)
	require.Equal(t, "PAID", orders[0].Status)
	require.Equal(t, "PAID", orders[1].Status)
	require.ElementsMatch(t, []string{"legacy-paid", "new-paid"}, []string{orders[0].OutTradeNo, orders[1].OutTradeNo})

	order, err := svc.GetOrder(ctx, orders[0].ID, user.ID)
	require.NoError(t, err)
	require.Equal(t, "PAID", order.Status)
}

func TestPaymentOrderAdminSummary_StatusVariantsAndTotals(t *testing.T) {
	ctx := context.Background()
	client := newPaymentOrderCompatTestClient(t)

	user := client.User.Create().
		SetEmail("summary@example.com").
		SetPasswordHash("hash").
		SetUsername("summary").
		SaveX(ctx)

	createCompatOrder(t, client, user.ID, "legacy-paid", "paid", time.Now().Add(-4*time.Hour))
	createCompatOrder(t, client, user.ID, "completed", "COMPLETED", time.Now().Add(-3*time.Hour))
	refunded := createCompatOrder(t, client, user.ID, "refunded", "refunded", time.Now().Add(-2*time.Hour))
	client.PaymentOrder.UpdateOneID(refunded.ID).SetRefundAmount(25).SaveX(ctx)
	createCompatOrder(t, client, user.ID, "pending", "PENDING", time.Now().Add(-1*time.Hour))

	svc := &PaymentService{entClient: client}
	summary, err := svc.AdminOrderSummary(ctx, user.ID, OrderListParams{})
	require.NoError(t, err)
	require.Equal(t, 4, summary.TotalOrders)
	require.Equal(t, 3, summary.PaidOrders)
	require.Equal(t, 1, summary.RefundedOrders)
	require.Equal(t, 400.0, summary.AmountTotal)
	require.Equal(t, 400.0, summary.PayAmountTotal)
	require.Equal(t, 25.0, summary.RefundAmountTotal)
	require.Equal(t, 375.0, summary.NetPayAmountTotal)
}

func newPaymentOrderCompatTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:payment_order_legacy_compat?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func createCompatOrder(t *testing.T, client *dbent.Client, userID int64, outTradeNo, status string, createdAt time.Time) *dbent.PaymentOrder {
	t.Helper()

	return client.PaymentOrder.Create().
		SetUserID(userID).
		SetUserEmail("compat@example.com").
		SetUserName("compat").
		SetAmount(100).
		SetPayAmount(100).
		SetRechargeCode(outTradeNo).
		SetOutTradeNo(outTradeNo).
		SetPaymentType("alipay").
		SetPaymentTradeNo("").
		SetStatus(status).
		SetExpiresAt(createdAt.Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("test.local").
		SetCreatedAt(createdAt).
		SaveX(context.Background())
}
