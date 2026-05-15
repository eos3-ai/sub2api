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
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	invoiceService *service.InvoiceService
	userRepo       service.UserRepository
}

func NewInvoiceHandler(invoiceService *service.InvoiceService, userRepo service.UserRepository) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService, userRepo: userRepo}
}

func (h *InvoiceHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	filter, err := h.parseFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	items, result, err := h.invoiceService.AdminListInvoiceRequests(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize}, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	emailByUserID := h.resolveUserEmails(c.Request.Context(), items)
	out := make([]dto.InvoiceRequest, 0, len(items))
	for i := range items {
		item := dto.InvoiceRequestFromService(&items[i])
		if item != nil {
			item.UserEmail = emailByUserID[item.UserID]
			out = append(out, *item)
		}
	}
	total := int64(len(out))
	if result != nil {
		total = result.Total
	}
	response.Paginated(c, out, total, page, pageSize)
}

func (h *InvoiceHandler) Export(c *gin.Context) {
	filter, err := h.parseFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	items, err := h.fetchAll(c.Request.Context(), filter, 200)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	emailByUserID := h.resolveUserEmails(c.Request.Context(), items)
	ids := make([]int64, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}
	orderNosByReqID, err := h.invoiceService.ListInvoiceOrderNosByRequestIDs(c.Request.Context(), ids)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"invoice_request_no", "status", "user_email", "invoice_type", "buyer_type",
		"invoice_title", "tax_no", "receiver_email", "amount_cny_total", "total_usd_total",
		"invoice_item_name", "order_nos", "created_at", "reviewed_at", "reject_reason",
		"issued_at", "invoice_number", "invoice_date", "invoice_pdf_url",
	})
	for i := range items {
		it := items[i]
		_ = w.Write([]string{
			invoiceCSVSafe(it.InvoiceRequestNo),
			invoiceCSVSafe(it.Status),
			invoiceCSVSafe(emailByUserID[it.UserID]),
			invoiceCSVSafe(it.InvoiceType),
			invoiceCSVSafe(it.BuyerType),
			invoiceCSVSafe(it.InvoiceTitle),
			invoiceCSVSafe(it.TaxNo),
			invoiceCSVSafe(it.ReceiverEmail),
			fmt.Sprintf("%.2f", it.AmountCNYTotal),
			fmt.Sprintf("%.8f", it.TotalUSDTotal),
			invoiceCSVSafe(it.InvoiceItemName),
			invoiceCSVSafe(strings.Join(orderNosByReqID[it.ID], "|")),
			it.CreatedAt.Format(time.RFC3339),
			invoiceCSVTime(it.ReviewedAt),
			invoiceCSVSafe(it.RejectReason),
			invoiceCSVTime(it.IssuedAt),
			invoiceCSVSafe(it.InvoiceNumber),
			invoiceCSVDate(it.InvoiceDate),
			invoiceCSVSafe(it.InvoicePDFURL),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		response.InternalError(c, "failed to export invoices")
		return
	}

	filename := "invoice_requests_" + time.Now().UTC().Format("20060102_150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

func (h *InvoiceHandler) GetByID(c *gin.Context) {
	id, ok := parseAdminInvoiceID(c)
	if !ok {
		return
	}
	req, items, err := h.invoiceService.AdminGetInvoiceRequestDetail(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	outReq := dto.InvoiceRequestFromService(req)
	if outReq != nil {
		outReq.UserEmail = h.userEmail(c.Request.Context(), req.UserID)
	}
	response.Success(c, gin.H{"invoice": outReq, "items": invoiceAdminOrderItemsDTO(items)})
}

func (h *InvoiceHandler) Approve(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, ok := parseAdminInvoiceID(c)
	if !ok {
		return
	}
	updated, err := h.invoiceService.AdminApproveInvoiceRequest(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.InvoiceRequestFromService(updated))
}

type rejectInvoiceRequest struct {
	RejectReason string `json:"reject_reason"`
}

func (h *InvoiceHandler) Reject(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, ok := parseAdminInvoiceID(c)
	if !ok {
		return
	}
	var req rejectInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	updated, err := h.invoiceService.AdminRejectInvoiceRequest(c.Request.Context(), subject.UserID, id, req.RejectReason)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.InvoiceRequestFromService(updated))
}

type issueInvoiceRequest struct {
	InvoiceNumber string `json:"invoice_number"`
	InvoiceDate   string `json:"invoice_date"`
	InvoicePDFURL string `json:"invoice_pdf_url"`
}

func (h *InvoiceHandler) Issue(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	id, ok := parseAdminInvoiceID(c)
	if !ok {
		return
	}
	var req issueInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	var invoiceDate *time.Time
	if strings.TrimSpace(req.InvoiceDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(req.InvoiceDate))
		if err != nil {
			response.BadRequest(c, "Invalid invoice_date (use YYYY-MM-DD)")
			return
		}
		tt := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		invoiceDate = &tt
	}
	updated, err := h.invoiceService.AdminIssueInvoiceRequest(c.Request.Context(), subject.UserID, id, &service.AdminIssueInvoiceInput{
		InvoiceNumber: req.InvoiceNumber,
		InvoiceDate:   invoiceDate,
		InvoicePDFURL: req.InvoicePDFURL,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.InvoiceRequestFromService(updated))
}

func (h *InvoiceHandler) parseFilter(c *gin.Context) (service.InvoiceRequestListFilter, error) {
	var filter service.InvoiceRequestListFilter
	if s := strings.TrimSpace(c.Query("status")); s != "" {
		filter.Status = strings.ToLower(s)
	}
	if s := strings.TrimSpace(c.Query("from")); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return filter, err
		}
		filter.From = &t
	}
	if s := strings.TrimSpace(c.Query("to")); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return filter, err
		}
		filter.To = &t
	}
	if userEmail := strings.TrimSpace(c.Query("user_email")); userEmail != "" {
		if h.userRepo == nil {
			return filter, fmt.Errorf("user repository is required for user_email filter")
		}
		user, err := h.userRepo.GetByEmail(c.Request.Context(), userEmail)
		if err != nil {
			return filter, err
		}
		if user == nil {
			zero := int64(0)
			filter.UserID = &zero
		} else {
			filter.UserID = &user.ID
		}
	}
	return filter, nil
}

func (h *InvoiceHandler) resolveUserEmails(ctx context.Context, items []service.InvoiceRequest) map[int64]string {
	out := map[int64]string{}
	seen := map[int64]struct{}{}
	for i := range items {
		id := items[i].UserID
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out[id] = h.userEmail(ctx, id)
	}
	return out
}

func (h *InvoiceHandler) userEmail(ctx context.Context, userID int64) string {
	if h.userRepo == nil || userID <= 0 {
		return ""
	}
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return ""
	}
	return user.Email
}

func (h *InvoiceHandler) fetchAll(ctx context.Context, filter service.InvoiceRequestListFilter, pageSize int) ([]service.InvoiceRequest, error) {
	page := 1
	out := make([]service.InvoiceRequest, 0, pageSize)
	for {
		items, result, err := h.invoiceService.AdminListInvoiceRequests(ctx, pagination.PaginationParams{Page: page, PageSize: pageSize}, filter)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if result == nil || int64(len(out)) >= result.Total || len(items) == 0 {
			return out, nil
		}
		page++
		if page > 10000 {
			return out, nil
		}
	}
}

func parseAdminInvoiceID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
		return 0, false
	}
	return id, true
}

func invoiceAdminOrderItemsDTO(items []service.InvoiceOrderItem) []dto.InvoiceOrderItem {
	out := make([]dto.InvoiceOrderItem, 0, len(items))
	for i := range items {
		item := dto.InvoiceOrderItemFromService(&items[i])
		if item != nil {
			out = append(out, *item)
		}
	}
	return out
}

func invoiceCSVSafe(value string) string {
	if value == "" {
		return ""
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	default:
		return value
	}
}

func invoiceCSVTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func invoiceCSVDate(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02")
}
