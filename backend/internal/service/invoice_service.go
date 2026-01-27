package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const invoiceRequestNotifyEmailsEnv = "INVOICE_REQUEST_NOTIFY_EMAILS"

type CreateInvoiceRequestInput struct {
	OrderNos []string

	InvoiceType  string
	BuyerType    string
	InvoiceTitle string
	TaxNo        string

	BuyerAddress     string
	BuyerPhone       string
	BuyerBankName    string
	BuyerBankAccount string

	ReceiverEmail string
	ReceiverPhone string

	InvoiceItemName string
	Remark          string
}

type AdminIssueInvoiceInput struct {
	InvoiceNumber string
	InvoiceDate   *time.Time
	InvoicePDFURL string
}

type InvoiceService struct {
	repo           InvoiceRepository
	settingService *SettingService
	emailService   *EmailService
	entClient      *dbent.Client
}

func NewInvoiceService(
	repo InvoiceRepository,
	settingService *SettingService,
	emailService *EmailService,
	entClient *dbent.Client,
) *InvoiceService {
	return &InvoiceService{
		repo:           repo,
		settingService: settingService,
		emailService:   emailService,
		entClient:      entClient,
	}
}

func (s *InvoiceService) ListEligibleOrders(ctx context.Context, userID int64, params pagination.PaginationParams, from, to *time.Time) ([]PaymentOrder, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, nil
	}
	return s.repo.ListEligibleOrders(ctx, userID, params, from, to)
}

func (s *InvoiceService) GetInvoiceProfile(ctx context.Context, userID int64) (*InvoiceProfile, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	profile, err := s.repo.GetInvoiceProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile != nil {
		if strings.TrimSpace(profile.InvoiceItemName) == "" {
			profile.InvoiceItemName = s.getInvoiceDefaultItemName(ctx)
		}
		return profile, nil
	}

	return &InvoiceProfile{
		UserID:           userID,
		InvoiceType:      InvoiceTypeNormal,
		BuyerType:        InvoiceBuyerTypeCompany,
		InvoiceTitle:     "",
		TaxNo:            "",
		BuyerAddress:     "",
		BuyerPhone:       "",
		BuyerBankName:    "",
		BuyerBankAccount: "",
		ReceiverEmail:    "",
		ReceiverPhone:    "",
		InvoiceItemName:  s.getInvoiceDefaultItemName(ctx),
		Remark:           "",
	}, nil
}

func (s *InvoiceService) UpdateInvoiceProfile(ctx context.Context, userID int64, in *InvoiceProfile) (*InvoiceProfile, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	if in == nil {
		return nil, infraerrors.BadRequest("INVOICE_PROFILE_REQUIRED", "Invoice profile is required.")
	}

	profile := &InvoiceProfile{
		UserID:           userID,
		InvoiceType:      strings.ToLower(strings.TrimSpace(in.InvoiceType)),
		BuyerType:        strings.ToLower(strings.TrimSpace(in.BuyerType)),
		InvoiceTitle:     strings.TrimSpace(in.InvoiceTitle),
		TaxNo:            strings.TrimSpace(in.TaxNo),
		BuyerAddress:     strings.TrimSpace(in.BuyerAddress),
		BuyerPhone:       strings.TrimSpace(in.BuyerPhone),
		BuyerBankName:    strings.TrimSpace(in.BuyerBankName),
		BuyerBankAccount: strings.TrimSpace(in.BuyerBankAccount),
		ReceiverEmail:    strings.TrimSpace(in.ReceiverEmail),
		ReceiverPhone:    strings.TrimSpace(in.ReceiverPhone),
		InvoiceItemName:  strings.TrimSpace(in.InvoiceItemName),
		Remark:           strings.TrimSpace(in.Remark),
	}
	if profile.InvoiceItemName == "" {
		profile.InvoiceItemName = s.getInvoiceDefaultItemName(ctx)
	}

	if err := validateInvoiceInfo(profile.InvoiceType, profile.BuyerType, profile.InvoiceTitle, profile.TaxNo, profile.BuyerAddress, profile.BuyerPhone, profile.BuyerBankName, profile.BuyerBankAccount, profile.ReceiverEmail); err != nil {
		return nil, err
	}

	if err := s.repo.UpsertInvoiceProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *InvoiceService) CreateInvoiceRequest(ctx context.Context, userID int64, in *CreateInvoiceRequestInput) (*InvoiceRequest, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	if in == nil {
		return nil, infraerrors.BadRequest("INVOICE_INVALID_REQUEST", "Invalid request.")
	}

	orderNos := normalizeOrderNos(in.OrderNos)
	if len(orderNos) == 0 {
		return nil, infraerrors.BadRequest("INVOICE_ORDER_NOS_REQUIRED", "order_nos is required.")
	}
	if len(orderNos) > 5 {
		return nil, infraerrors.BadRequest("INVOICE_TOO_MANY_ORDERS", "At most 5 orders can be invoiced in one request.")
	}

	invoiceType := strings.ToLower(strings.TrimSpace(in.InvoiceType))
	buyerType := strings.ToLower(strings.TrimSpace(in.BuyerType))
	invoiceTitle := strings.TrimSpace(in.InvoiceTitle)
	taxNo := strings.TrimSpace(in.TaxNo)
	buyerAddress := strings.TrimSpace(in.BuyerAddress)
	buyerPhone := strings.TrimSpace(in.BuyerPhone)
	buyerBankName := strings.TrimSpace(in.BuyerBankName)
	buyerBankAccount := strings.TrimSpace(in.BuyerBankAccount)
	receiverEmail := strings.TrimSpace(in.ReceiverEmail)
	receiverPhone := strings.TrimSpace(in.ReceiverPhone)
	itemName := strings.TrimSpace(in.InvoiceItemName)
	remark := strings.TrimSpace(in.Remark)

	if itemName == "" {
		itemName = s.getInvoiceDefaultItemName(ctx)
	}

	if err := validateInvoiceInfo(invoiceType, buyerType, invoiceTitle, taxNo, buyerAddress, buyerPhone, buyerBankName, buyerBankAccount, receiverEmail); err != nil {
		return nil, err
	}

	req := &InvoiceRequest{
		InvoiceRequestNo: s.generateInvoiceRequestNo(),
		UserID:           userID,
		Status:           InvoiceStatusSubmitted,
		InvoiceType:      invoiceType,
		BuyerType:        buyerType,
		InvoiceTitle:     invoiceTitle,
		TaxNo:            taxNo,
		BuyerAddress:     buyerAddress,
		BuyerPhone:       buyerPhone,
		BuyerBankName:    buyerBankName,
		BuyerBankAccount: buyerBankAccount,
		ReceiverEmail:    receiverEmail,
		ReceiverPhone:    receiverPhone,
		InvoiceItemName:  itemName,
		Remark:           remark,
	}

	err := s.withTx(ctx, func(txCtx context.Context) error {
		orders, err := s.repo.GetEligibleOrdersByOrderNos(txCtx, userID, orderNos)
		if err != nil {
			return err
		}
		if len(orders) != len(orderNos) {
			return infraerrors.BadRequest("INVOICE_ORDER_NOT_ELIGIBLE", "Some orders are not eligible for invoicing.")
		}

		var amountCNYTotal float64
		var totalUSDTotal float64
		for i := range orders {
			amountCNYTotal += orders[i].AmountCNY
			totalUSDTotal += orders[i].TotalUSD
		}
		req.AmountCNYTotal = amountCNYTotal
		req.TotalUSDTotal = totalUSDTotal

		if err := s.repo.CreateInvoiceRequest(txCtx, req, orders); err != nil {
			if infraerrors.IsConflict(err) {
				return err
			}
			// Unique conflict on invoice_order_items also indicates orders already invoiced.
			if infraerrors.Code(err) == 0 && errors.Is(err, infraerrors.Conflict("", "")) {
				return err
			}
			return err
		}

		// Auto-save as default profile (single default, UNIQUE(user_id)).
		profile := &InvoiceProfile{
			UserID:           userID,
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
		}
		if err := s.repo.UpsertInvoiceProfile(txCtx, profile); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		// Map unique conflicts to a stable 409 for clients.
		if infraerrors.IsConflict(err) {
			return nil, err
		}
		if isUniqueConstraintError(err) {
			return nil, infraerrors.Conflict("INVOICE_ORDER_ALREADY_INVOICED", "Some orders are already invoiced or in another request.").WithCause(err)
		}
		return nil, err
	}

	s.sendInvoiceRequestSubmittedNotification(ctx, req)
	return req, nil
}

func (s *InvoiceService) ListInvoiceRequests(ctx context.Context, userID int64, params pagination.PaginationParams) ([]InvoiceRequest, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, nil
	}
	filter := InvoiceRequestListFilter{UserID: &userID}
	return s.repo.ListInvoiceRequests(ctx, params, filter)
}

func (s *InvoiceService) GetInvoiceRequestDetail(ctx context.Context, userID int64, id int64) (*InvoiceRequest, []InvoiceOrderItem, error) {
	if s == nil || s.repo == nil {
		return nil, nil, nil
	}
	req, err := s.repo.GetInvoiceRequestByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if req == nil || req.UserID != userID {
		return nil, nil, infraerrors.NotFound("INVOICE_NOT_FOUND", "Invoice request not found.")
	}

	items, err := s.repo.ListInvoiceOrderItems(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return req, items, nil
}

func (s *InvoiceService) CancelInvoiceRequest(ctx context.Context, userID int64, id int64) (*InvoiceRequest, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}

	var updated *InvoiceRequest
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		req, err := s.repo.GetInvoiceRequestByID(txCtx, id)
		if err != nil {
			return err
		}
		if req == nil || req.UserID != userID {
			return infraerrors.NotFound("INVOICE_NOT_FOUND", "Invoice request not found.")
		}
		if req.Status != InvoiceStatusSubmitted {
			return infraerrors.BadRequest("INVOICE_CANNOT_CANCEL", "Only submitted invoice requests can be cancelled.")
		}

		req.Status = InvoiceStatusCancelled
		if err := s.repo.UpdateInvoiceRequest(txCtx, req); err != nil {
			return err
		}
		if err := s.repo.SetInvoiceOrderItemsActive(txCtx, id, false); err != nil {
			return err
		}
		updated = req
		return nil
	}); err != nil {
		return nil, err
	}

	return updated, nil
}

// Admin APIs

func (s *InvoiceService) AdminListInvoiceRequests(ctx context.Context, params pagination.PaginationParams, filter InvoiceRequestListFilter) ([]InvoiceRequest, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, nil
	}
	return s.repo.ListInvoiceRequests(ctx, params, filter)
}

func (s *InvoiceService) AdminGetInvoiceRequestDetail(ctx context.Context, id int64) (*InvoiceRequest, []InvoiceOrderItem, error) {
	if s == nil || s.repo == nil {
		return nil, nil, nil
	}
	req, err := s.repo.GetInvoiceRequestByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if req == nil {
		return nil, nil, infraerrors.NotFound("INVOICE_NOT_FOUND", "Invoice request not found.")
	}
	items, err := s.repo.ListInvoiceOrderItems(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return req, items, nil
}

func (s *InvoiceService) AdminApproveInvoiceRequest(ctx context.Context, adminID int64, id int64) (*InvoiceRequest, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	var updated *InvoiceRequest
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		req, err := s.repo.GetInvoiceRequestByID(txCtx, id)
		if err != nil {
			return err
		}
		if req == nil {
			return infraerrors.NotFound("INVOICE_NOT_FOUND", "Invoice request not found.")
		}
		if req.Status != InvoiceStatusSubmitted {
			return infraerrors.BadRequest("INVOICE_CANNOT_APPROVE", "Only submitted invoice requests can be approved.")
		}
		now := time.Now()
		req.Status = InvoiceStatusApproved
		req.RejectReason = ""
		req.ReviewedBy = &adminID
		req.ReviewedAt = &now

		if err := s.repo.UpdateInvoiceRequest(txCtx, req); err != nil {
			return err
		}
		updated = req
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *InvoiceService) AdminRejectInvoiceRequest(ctx context.Context, adminID int64, id int64, reason string) (*InvoiceRequest, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, infraerrors.BadRequest("INVOICE_REJECT_REASON_REQUIRED", "reject_reason is required.")
	}

	var updated *InvoiceRequest
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		req, err := s.repo.GetInvoiceRequestByID(txCtx, id)
		if err != nil {
			return err
		}
		if req == nil {
			return infraerrors.NotFound("INVOICE_NOT_FOUND", "Invoice request not found.")
		}
		if req.Status != InvoiceStatusSubmitted {
			return infraerrors.BadRequest("INVOICE_CANNOT_REJECT", "Only submitted invoice requests can be rejected.")
		}
		now := time.Now()
		req.Status = InvoiceStatusRejected
		req.RejectReason = reason
		req.ReviewedBy = &adminID
		req.ReviewedAt = &now

		if err := s.repo.UpdateInvoiceRequest(txCtx, req); err != nil {
			return err
		}
		// Release the locked payment orders so user can re-submit.
		if err := s.repo.SetInvoiceOrderItemsActive(txCtx, id, false); err != nil {
			return err
		}
		updated = req
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *InvoiceService) AdminIssueInvoiceRequest(ctx context.Context, adminID int64, id int64, in *AdminIssueInvoiceInput) (*InvoiceRequest, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	if in == nil {
		return nil, infraerrors.BadRequest("INVOICE_INVALID_REQUEST", "Invalid request.")
	}
	invoiceNumber := strings.TrimSpace(in.InvoiceNumber)
	pdfURL := strings.TrimSpace(in.InvoicePDFURL)

	var updated *InvoiceRequest
	var receiverEmail string
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		req, err := s.repo.GetInvoiceRequestByID(txCtx, id)
		if err != nil {
			return err
		}
		if req == nil {
			return infraerrors.NotFound("INVOICE_NOT_FOUND", "Invoice request not found.")
		}
		if req.Status != InvoiceStatusApproved {
			return infraerrors.BadRequest("INVOICE_CANNOT_ISSUE", "Only approved invoice requests can be issued.")
		}
		now := time.Now()
		req.Status = InvoiceStatusIssued
		req.IssuedBy = &adminID
		req.IssuedAt = &now
		req.InvoiceNumber = invoiceNumber
		req.InvoiceDate = in.InvoiceDate
		req.InvoicePDFURL = pdfURL

		if err := s.repo.UpdateInvoiceRequest(txCtx, req); err != nil {
			return err
		}
		updated = req
		receiverEmail = req.ReceiverEmail
		return nil
	}); err != nil {
		return nil, err
	}

	// Best-effort email notification.
	if s.emailService != nil && receiverEmail != "" && strings.TrimSpace(updated.InvoicePDFURL) != "" {
		if err := s.emailService.SendEmail(ctx, receiverEmail, "发票已开具", fmt.Sprintf("您的发票已开具，下载链接：<a href=\"%s\">%s</a>", updated.InvoicePDFURL, updated.InvoicePDFURL)); err != nil {
			if !errors.Is(err, ErrEmailNotConfigured) {
				log.Printf("[Invoice] send email failed: invoice_request_id=%d to=%s err=%v", id, receiverEmail, err)
			}
		}
	}

	return updated, nil
}

func (s *InvoiceService) ListInvoiceOrderNosByRequestIDs(ctx context.Context, invoiceRequestIDs []int64) (map[int64][]string, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	return s.repo.ListInvoiceOrderNosByRequestIDs(ctx, invoiceRequestIDs)
}

func (s *InvoiceService) withTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	if s == nil || s.entClient == nil {
		return fn(ctx)
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return err
	}
	txCtx := dbent.NewTxContext(ctx, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}

func (s *InvoiceService) generateInvoiceRequestNo() string {
	now := time.Now().UTC()
	random := cryptoRandInt64(1000000)
	return "IR" + now.Format("20060102150405") + fmt.Sprintf("%09d", now.Nanosecond()) + fmt.Sprintf("%06d", random)
}

func normalizeOrderNos(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func (s *InvoiceService) sendInvoiceRequestSubmittedNotification(ctx context.Context, req *InvoiceRequest) {
	if s == nil || s.emailService == nil || req == nil {
		return
	}

	recipients := parseEmailListFromEnv(invoiceRequestNotifyEmailsEnv)
	if len(recipients) == 0 {
		return
	}

	taxNo := strings.TrimSpace(req.TaxNo)
	if taxNo == "" {
		taxNo = "-"
	}

	subject := "开票申请已提交"
	body := fmt.Sprintf(
		"## 新的开票申请已提交\n\n- 抬头：%s\n- 税号：%s\n- 申请开票金额（CNY）：%.2f\n- 收票邮箱：%s\n",
		strings.TrimSpace(req.InvoiceTitle),
		taxNo,
		req.AmountCNYTotal,
		strings.TrimSpace(req.ReceiverEmail),
	)

	for _, to := range recipients {
		if err := s.emailService.SendPlainTextEmail(ctx, to, subject, body); err != nil {
			if errors.Is(err, ErrEmailNotConfigured) {
				return
			}
			log.Printf("[Invoice] send submit notify email failed: invoice_request_no=%s to=%s err=%v", req.InvoiceRequestNo, to, err)
		}
	}
}

func parseEmailListFromEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', ';', ' ', '\t', '\n', '\r':
			return true
		default:
			return false
		}
	})
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		addr := strings.TrimSpace(p)
		if addr == "" {
			continue
		}
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	return out
}

func validateInvoiceInfo(invoiceType, buyerType, invoiceTitle, taxNo, buyerAddress, buyerPhone, buyerBankName, buyerBankAccount, receiverEmail string) error {
	switch invoiceType {
	case InvoiceTypeNormal, InvoiceTypeSpecial:
	default:
		return infraerrors.BadRequest("INVOICE_INVALID_INVOICE_TYPE", "Invalid invoice_type.")
	}
	switch buyerType {
	case InvoiceBuyerTypePersonal, InvoiceBuyerTypeCompany:
	default:
		return infraerrors.BadRequest("INVOICE_INVALID_BUYER_TYPE", "Invalid buyer_type.")
	}
	if strings.TrimSpace(invoiceTitle) == "" {
		return infraerrors.BadRequest("INVOICE_TITLE_REQUIRED", "invoice_title is required.")
	}
	if strings.TrimSpace(receiverEmail) == "" {
		return infraerrors.BadRequest("INVOICE_RECEIVER_EMAIL_REQUIRED", "receiver_email is required.")
	}
	// Lightweight email sanity check (allows non-login email).
	if !strings.Contains(receiverEmail, "@") {
		return infraerrors.BadRequest("INVOICE_INVALID_RECEIVER_EMAIL", "Invalid receiver_email.")
	}

	// 企业：税号必填（普票/专票都适用）
	if buyerType == InvoiceBuyerTypeCompany && strings.TrimSpace(taxNo) == "" {
		return infraerrors.BadRequest("INVOICE_TAX_NO_REQUIRED", "tax_no is required for company invoices.")
	}

	if invoiceType == InvoiceTypeSpecial {
		if buyerType != InvoiceBuyerTypeCompany {
			return infraerrors.BadRequest("INVOICE_SPECIAL_COMPANY_ONLY", "special invoice requires buyer_type=company.")
		}
		if strings.TrimSpace(buyerAddress) == "" ||
			strings.TrimSpace(buyerPhone) == "" ||
			strings.TrimSpace(buyerBankName) == "" ||
			strings.TrimSpace(buyerBankAccount) == "" {
			return infraerrors.BadRequest("INVOICE_SPECIAL_FIELDS_REQUIRED", "special invoice requires address/phone/bank fields.")
		}
	}
	return nil
}

func (s *InvoiceService) getInvoiceDefaultItemName(ctx context.Context) string {
	const fallback = "技术服务费"
	if s == nil || s.settingService == nil {
		return fallback
	}
	v := strings.TrimSpace(s.settingService.GetInvoiceDefaultItemName(ctx))
	if v == "" {
		return fallback
	}
	return v
}

func isUniqueConstraintError(err error) bool {
	// Reuse repository-level detection via message/code, but keep service package decoupled.
	// The infraerrors package doesn't expose a "IsUniqueConstraintViolation" helper.
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicate entry") ||
		strings.Contains(msg, "23505")
}
