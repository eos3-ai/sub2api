package handler

import (
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
}

func NewInvoiceHandler(invoiceService *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

// ListEligibleOrders lists current user's eligible paid orders for invoicing.
// GET /api/v1/invoices/eligible-orders?page=1&page_size=20&from=...&to=...
func (h *InvoiceHandler) ListEligibleOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	var from *time.Time
	if s := strings.TrimSpace(c.Query("from")); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			response.BadRequest(c, "Invalid from (use RFC3339)")
			return
		}
		from = &t
	}
	var to *time.Time
	if s := strings.TrimSpace(c.Query("to")); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			response.BadRequest(c, "Invalid to (use RFC3339)")
			return
		}
		to = &t
	}

	orders, result, err := h.invoiceService.ListEligibleOrders(c.Request.Context(), subject.UserID, params, from, to)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.PaymentOrder, 0, len(orders))
	for i := range orders {
		item := dto.PaymentOrderFromService(&orders[i])
		if item != nil {
			out = append(out, *item)
		}
	}
	total := int64(len(out))
	if result != nil {
		total = result.Total
	}
	response.Paginated(c, out, total, page, pageSize)
}

type createInvoiceRequestRequest struct {
	OrderNos []string `json:"order_nos"`

	InvoiceType  string `json:"invoice_type"`
	BuyerType    string `json:"buyer_type"`
	InvoiceTitle string `json:"invoice_title"`
	TaxNo        string `json:"tax_no"`

	BuyerAddress     string `json:"buyer_address"`
	BuyerPhone       string `json:"buyer_phone"`
	BuyerBankName    string `json:"buyer_bank_name"`
	BuyerBankAccount string `json:"buyer_bank_account"`

	ReceiverEmail string `json:"receiver_email"`
	ReceiverPhone string `json:"receiver_phone"`

	InvoiceItemName string `json:"invoice_item_name"`
	Remark          string `json:"remark"`
}

// CreateInvoiceRequest creates an invoice request.
// POST /api/v1/invoices
func (h *InvoiceHandler) CreateInvoiceRequest(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req createInvoiceRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	created, err := h.invoiceService.CreateInvoiceRequest(c.Request.Context(), subject.UserID, &service.CreateInvoiceRequestInput{
		OrderNos:         req.OrderNos,
		InvoiceType:      req.InvoiceType,
		BuyerType:        req.BuyerType,
		InvoiceTitle:     req.InvoiceTitle,
		TaxNo:            req.TaxNo,
		BuyerAddress:     req.BuyerAddress,
		BuyerPhone:       req.BuyerPhone,
		BuyerBankName:    req.BuyerBankName,
		BuyerBankAccount: req.BuyerBankAccount,
		ReceiverEmail:    req.ReceiverEmail,
		ReceiverPhone:    req.ReceiverPhone,
		InvoiceItemName:  req.InvoiceItemName,
		Remark:           req.Remark,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, dto.InvoiceRequestFromService(created))
}

// ListMyInvoiceRequests lists current user's invoice requests.
// GET /api/v1/invoices?page=1&page_size=20
func (h *InvoiceHandler) ListMyInvoiceRequests(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}

	items, result, err := h.invoiceService.ListInvoiceRequests(c.Request.Context(), subject.UserID, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.InvoiceRequest, 0, len(items))
	for i := range items {
		item := dto.InvoiceRequestFromService(&items[i])
		if item != nil {
			out = append(out, *item)
		}
	}
	total := int64(len(out))
	if result != nil {
		total = result.Total
	}
	response.Paginated(c, out, total, page, pageSize)
}

// GetMyInvoiceRequest returns current user's invoice request detail.
// GET /api/v1/invoices/:id
func (h *InvoiceHandler) GetMyInvoiceRequest(c *gin.Context) {
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

	req, items, err := h.invoiceService.GetInvoiceRequestDetail(c.Request.Context(), subject.UserID, id)
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

	response.Success(c, gin.H{
		"invoice": dto.InvoiceRequestFromService(req),
		"items":   outItems,
	})
}

// CancelInvoiceRequest cancels a submitted invoice request.
// POST /api/v1/invoices/:id/cancel
func (h *InvoiceHandler) CancelInvoiceRequest(c *gin.Context) {
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

	updated, err := h.invoiceService.CancelInvoiceRequest(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.InvoiceRequestFromService(updated))
}

// GetProfile returns current user's default invoice profile.
// GET /api/v1/invoices/profile
func (h *InvoiceHandler) GetProfile(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	profile, err := h.invoiceService.GetInvoiceProfile(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.InvoiceProfileFromService(profile))
}

// UpdateProfile updates current user's default invoice profile.
// PUT /api/v1/invoices/profile
func (h *InvoiceHandler) UpdateProfile(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req dto.InvoiceProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	updated, err := h.invoiceService.UpdateInvoiceProfile(c.Request.Context(), subject.UserID, &service.InvoiceProfile{
		InvoiceType:      req.InvoiceType,
		BuyerType:        req.BuyerType,
		InvoiceTitle:     req.InvoiceTitle,
		TaxNo:            req.TaxNo,
		BuyerAddress:     req.BuyerAddress,
		BuyerPhone:       req.BuyerPhone,
		BuyerBankName:    req.BuyerBankName,
		BuyerBankAccount: req.BuyerBankAccount,
		ReceiverEmail:    req.ReceiverEmail,
		ReceiverPhone:    req.ReceiverPhone,
		InvoiceItemName:  req.InvoiceItemName,
		Remark:           req.Remark,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.InvoiceProfileFromService(updated))
}
