package admin

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentOrdersHandler struct {
	paymentService *service.PaymentService
	userRepo       service.UserRepository
}

func NewPaymentOrdersHandler(paymentService *service.PaymentService, userRepo service.UserRepository) *PaymentOrdersHandler {
	return &PaymentOrdersHandler{paymentService: paymentService, userRepo: userRepo}
}

// List lists all payment orders for admins.
// GET /api/v1/admin/payment/orders?page=1&page_size=20&provider=zpay|stripe&user_id=123
func (h *PaymentOrdersHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	filter, err := h.parsePaymentOrderFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	orders, result, err := h.paymentService.ListOrders(c.Request.Context(), params, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.PaymentOrder, 0, len(orders))
	for i := range orders {
		out = append(out, *dto.PaymentOrderFromService(&orders[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// Export exports filtered orders as CSV.
// GET /api/v1/admin/payment/orders/export?provider=zpay|stripe&user_id=123
func (h *PaymentOrdersHandler) Export(c *gin.Context) {
	filter, err := h.parsePaymentOrderFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	orders, err := h.fetchAll(ctx, filter, 100)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"order_no",
		"order_type",
		"user_id",
		"provider",
		"amount_cny",
		"amount_usd",
		"total_usd",
		"status",
		"created_at",
		"paid_at",
	})
	for i := range orders {
		o := orders[i]
		paidAt := ""
		if o.PaidAt != nil {
			paidAt = o.PaidAt.Format(time.RFC3339)
		}
		orderType := "online_recharge"
		if strings.EqualFold(o.Provider, "admin") {
			orderType = "admin_recharge"
		}
		_ = w.Write([]string{
			o.OrderNo,
			orderType,
			strconv.FormatInt(o.UserID, 10),
			o.Provider,
			fmt.Sprintf("%.2f", o.AmountCNY),
			fmt.Sprintf("%.8f", o.AmountUSD),
			fmt.Sprintf("%.8f", o.TotalUSD),
			o.Status,
			o.CreatedAt.Format(time.RFC3339),
			paidAt,
		})
	}
	w.Flush()

	filename := "payment_orders.csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.String(http.StatusOK, buf.String())
}

func (h *PaymentOrdersHandler) fetchAll(ctx context.Context, filter service.PaymentOrderFilter, pageSize int) ([]service.PaymentOrder, error) {
	page := 1
	out := make([]service.PaymentOrder, 0, pageSize)
	for {
		params := pagination.PaginationParams{Page: page, PageSize: pageSize}
		items, result, err := h.paymentService.ListOrders(ctx, params, filter)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if result == nil || int64(len(out)) >= result.Total || len(items) == 0 {
			return out, nil
		}
		page++
		// Safety guard to prevent infinite loops if total is inconsistent.
		if page > 10000 {
			return out, nil
		}
	}
}

func parsePaymentOrderFilter(c *gin.Context) (service.PaymentOrderFilter, error) {
	var filter service.PaymentOrderFilter

	method := strings.TrimSpace(c.Query("method"))
	provider := strings.TrimSpace(c.Query("provider"))
	if provider == "" && method != "" {
		switch strings.ToLower(method) {
		case "alipay":
			provider = "zpay"
		case "wechat":
			provider = "stripe"
		default:
			provider = method
		}
	}
	if provider != "" {
		filter.Provider = strings.ToLower(provider)
	}

	if s := strings.TrimSpace(c.Query("status")); s != "" {
		filter.Status = strings.ToLower(s)
	}

	userQuery := strings.TrimSpace(c.Query("user_id"))
	if userQuery == "" {
		userQuery = strings.TrimSpace(c.Query("user"))
	}
	if userQuery != "" {
		// Support both user ID and email.
		if id, err := strconv.ParseInt(userQuery, 10, 64); err == nil && id > 0 {
			filter.UserID = &id
		} else if strings.Contains(userQuery, "@") {
			return filter, fmt.Errorf("user email filter requires user repository")
		} else {
			return filter, fmt.Errorf("invalid user (use user_id or email)")
		}
	}

	if s := strings.TrimSpace(c.Query("from")); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return filter, fmt.Errorf("invalid from (use RFC3339)")
		}
		filter.From = &t
	}
	if s := strings.TrimSpace(c.Query("to")); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return filter, fmt.Errorf("invalid to (use RFC3339)")
		}
		filter.To = &t
	}

	return filter, nil
}

func (h *PaymentOrdersHandler) parsePaymentOrderFilter(c *gin.Context) (service.PaymentOrderFilter, error) {
	filter, err := parsePaymentOrderFilter(c)
	if err == nil && filter.UserID != nil {
		return filter, nil
	}
	if err != nil && strings.Contains(err.Error(), "requires user repository") {
		// Fall through to resolve email.
	} else if err != nil {
		return filter, err
	}

	userQuery := strings.TrimSpace(c.Query("user"))
	if userQuery == "" {
		userQuery = strings.TrimSpace(c.Query("user_id"))
	}
	if userQuery == "" || !strings.Contains(userQuery, "@") {
		return filter, nil
	}
	if h == nil || h.userRepo == nil {
		return filter, fmt.Errorf("user repository is not configured")
	}

	user, uerr := h.userRepo.GetByEmail(c.Request.Context(), userQuery)
	if uerr != nil {
		return filter, fmt.Errorf("lookup user: %w", uerr)
	}
	if user == nil {
		return filter, fmt.Errorf("user not found")
	}
	id := user.ID
	filter.UserID = &id
	return filter, nil
}
