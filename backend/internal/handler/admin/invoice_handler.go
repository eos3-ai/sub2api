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
	return &InvoiceHandler{
		invoiceService: invoiceService,
		userRepo:       userRepo,
	}
}

// Export exports filtered invoice requests as CSV.
// GET /api/v1/admin/invoices/export?status=submitted|approved|rejected|issued|cancelled&user_email=...&from=...&to=...
func (h *InvoiceHandler) Export(c *gin.Context) {
	filter, err := h.parseFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	items, err := h.fetchAll(ctx, filter, 200)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	emailByUserID := h.resolveUserEmails(ctx, items)

	ids := make([]int64, 0, len(items))
	for i := range items {
		if items[i].ID > 0 {
			ids = append(ids, items[i].ID)
		}
	}
	orderNosByReqID, err := h.invoiceService.ListInvoiceOrderNosByRequestIDs(ctx, ids)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"invoice_request_no",
		"status",
		"user_email",
		"invoice_type",
		"buyer_type",
		"invoice_title",
		"tax_no",
		"receiver_email",
		"amount_cny_total",
		"total_usd_total",
		"invoice_item_name",
		"order_nos",
		"created_at",
		"reviewed_at",
		"reject_reason",
		"issued_at",
		"invoice_number",
		"invoice_date",
		"invoice_pdf_url",
	})

	for i := range items {
		it := items[i]
		userEmail := emailByUserID[it.UserID]
		createdAt := it.CreatedAt.Format(time.RFC3339)

		reviewedAt := ""
		if it.ReviewedAt != nil {
			reviewedAt = it.ReviewedAt.Format(time.RFC3339)
		}
		issuedAt := ""
		if it.IssuedAt != nil {
			issuedAt = it.IssuedAt.Format(time.RFC3339)
		}
		invoiceDate := ""
		if it.InvoiceDate != nil {
			invoiceDate = it.InvoiceDate.Format("2006-01-02")
		}

		orderNos := strings.Join(orderNosByReqID[it.ID], "|")

		_ = w.Write([]string{
			it.InvoiceRequestNo,
			it.Status,
			sanitizeCSVCell(userEmail),
			it.InvoiceType,
			it.BuyerType,
			sanitizeCSVCell(it.InvoiceTitle),
			sanitizeCSVCell(it.TaxNo),
			sanitizeCSVCell(it.ReceiverEmail),
			fmt.Sprintf("%.2f", it.AmountCNYTotal),
			fmt.Sprintf("%.8f", it.TotalUSDTotal),
			sanitizeCSVCell(it.InvoiceItemName),
			sanitizeCSVCell(orderNos),
			createdAt,
			reviewedAt,
			sanitizeCSVCell(it.RejectReason),
			issuedAt,
			sanitizeCSVCell(it.InvoiceNumber),
			invoiceDate,
			sanitizeCSVCell(it.InvoicePDFURL),
		})
	}
	w.Flush()

	filename := "invoice_requests.csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.String(http.StatusOK, buf.String())
}

// List lists invoice requests for admins.
// GET /api/v1/admin/invoices?page=1&page_size=20&status=submitted|approved|rejected|issued|cancelled&user_email=...&from=...&to=...
func (h *InvoiceHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	filter, err := h.parseFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	items, result, err := h.invoiceService.AdminListInvoiceRequests(c.Request.Context(), params, filter)
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

// GetByID returns invoice request detail for admins.
// GET /api/v1/admin/invoices/:id
func (h *InvoiceHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
		return
	}

	req, items, err := h.invoiceService.AdminGetInvoiceRequestDetail(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	outItems := make([]dto.InvoiceOrderItem, 0, len(items))
	for i := range items {
		item := dto.InvoiceOrderItemFromService(&items[i])
		if item != nil {
			outItems = append(outItems, *item)
		}
	}

	userEmails, _ := h.userRepo.GetEmailsByIDs(c.Request.Context(), []int64{req.UserID})
	outReq := dto.InvoiceRequestFromService(req)
	if outReq != nil {
		outReq.UserEmail = userEmails[req.UserID]
	}

	response.Success(c, gin.H{
		"invoice": outReq,
		"items":   outItems,
	})
}

// Approve approves a submitted invoice request.
// POST /api/v1/admin/invoices/:id/approve
func (h *InvoiceHandler) Approve(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
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

// Reject rejects a submitted invoice request.
// POST /api/v1/admin/invoices/:id/reject
func (h *InvoiceHandler) Reject(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
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
	InvoiceDate   string `json:"invoice_date"` // YYYY-MM-DD, optional
	InvoicePDFURL string `json:"invoice_pdf_url"`
}

// Issue marks an approved invoice request as issued and stores invoice metadata.
// POST /api/v1/admin/invoices/:id/issue
func (h *InvoiceHandler) Issue(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
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

	userEmail := strings.TrimSpace(c.Query("user_email"))
	if userEmail != "" {
		if h.userRepo == nil {
			return filter, fmt.Errorf("user repository is required for user_email filter")
		}
		user, err := h.userRepo.GetByEmail(c.Request.Context(), userEmail)
		if err != nil {
			return filter, err
		}
		if user == nil {
			// No such user -> empty result set.
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
	if h.userRepo == nil || len(items) == 0 {
		return out
	}
	ids := make([]int64, 0, len(items))
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
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return out
	}
	m, err := h.userRepo.GetEmailsByIDs(ctx, ids)
	if err != nil {
		return out
	}
	return m
}

func (h *InvoiceHandler) fetchAll(ctx context.Context, filter service.InvoiceRequestListFilter, pageSize int) ([]service.InvoiceRequest, error) {
	page := 1
	out := make([]service.InvoiceRequest, 0, pageSize)
	for {
		params := pagination.PaginationParams{Page: page, PageSize: pageSize}
		items, result, err := h.invoiceService.AdminListInvoiceRequests(ctx, params, filter)
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
