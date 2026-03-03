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
	"github.com/Wei-Shaw/sub2api/internal/config"
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
	cfg            *config.Config
	nuonuoService  *NuonuoInvoiceService
}

func NewInvoiceService(
	repo InvoiceRepository,
	settingService *SettingService,
	emailService *EmailService,
	entClient *dbent.Client,
	cfg *config.Config,
	nuonuoService *NuonuoInvoiceService,
) *InvoiceService {
	return &InvoiceService{
		repo:           repo,
		settingService: settingService,
		emailService:   emailService,
		entClient:      entClient,
		cfg:            cfg,
		nuonuoService:  nuonuoService,
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
		req.Provider = InvoiceProviderManual
		if s.shouldAutoIssue() {
			req.Provider = InvoiceProviderNuonuo
		}

		if err := s.repo.UpdateInvoiceRequest(txCtx, req); err != nil {
			return err
		}
		updated = req
		return nil
	}); err != nil {
		return nil, err
	}

	// Trigger async auto-issue after successful approval.
	if s.shouldAutoIssue() {
		go s.autoIssueInvoiceAsync(context.Background(), id)
	}

	return updated, nil
}

func (s *InvoiceService) shouldAutoIssue() bool {
	if s == nil || s.cfg == nil || s.nuonuoService == nil {
		return false
	}
	return strings.ToLower(strings.TrimSpace(s.cfg.Payment.Invoice.Provider)) == InvoiceProviderNuonuo &&
		s.cfg.Payment.Invoice.Nuonuo.Enabled
}

func (s *InvoiceService) autoIssueInvoiceAsync(ctx context.Context, invoiceRequestID int64) {
	// 1. Transition to "issuing".
	now := time.Now()
	var req *InvoiceRequest
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		r, err := s.repo.GetInvoiceRequestByID(txCtx, invoiceRequestID)
		if err != nil {
			return err
		}
		if r == nil || r.Status != InvoiceStatusApproved {
			return fmt.Errorf("invoice %d not in approved state", invoiceRequestID)
		}
		r.Status = InvoiceStatusIssuing
		r.IssuingStartedAt = &now
		r.ProviderError = ""
		if err := s.repo.UpdateInvoiceRequest(txCtx, r); err != nil {
			return err
		}
		req = r
		return nil
	}); err != nil {
		log.Printf("[Invoice] auto-issue: set issuing failed: invoice_id=%d err=%v", invoiceRequestID, err)
		return
	}

	// 2. Fetch order items for amount info.
	items, err := s.repo.ListInvoiceOrderItems(ctx, invoiceRequestID)
	if err != nil {
		log.Printf("[Invoice] auto-issue: list items failed: invoice_id=%d err=%v", invoiceRequestID, err)
		s.revertIssuingToApproved(ctx, invoiceRequestID, fmt.Sprintf("获取订单明细失败: %v", err))
		return
	}

	// 3. Call Nuonuo API.
	issueReq := &NuonuoIssueRequest{
		InvoiceType:      req.InvoiceType,
		OrderNo:          req.InvoiceRequestNo,
		BuyerName:        req.InvoiceTitle,
		BuyerTaxNo:       req.TaxNo,
		BuyerAddress:     req.BuyerAddress,
		BuyerPhone:       req.BuyerPhone,
		BuyerBankName:    req.BuyerBankName,
		BuyerBankAccount: req.BuyerBankAccount,
		ReceiverEmail:    req.ReceiverEmail,
		ItemName:         req.InvoiceItemName,
		TotalAmountCNY:   req.AmountCNYTotal,
	}
	_ = items // items used for amount total which is already on the request

	nuonuoResp, err := s.nuonuoService.IssueInvoice(ctx, issueReq)
	if err != nil {
		log.Printf("[Invoice] auto-issue: nuonuo API failed: invoice_id=%d err=%v", invoiceRequestID, err)
		s.revertIssuingToApproved(ctx, invoiceRequestID, err.Error())
		return
	}

	// 4. Save provider_invoice_id; status stays "issuing" until webhook arrives.
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		r, err := s.repo.GetInvoiceRequestByID(txCtx, invoiceRequestID)
		if err != nil {
			return err
		}
		if r == nil {
			return fmt.Errorf("invoice %d not found", invoiceRequestID)
		}
		r.ProviderInvoiceID = nuonuoResp.SerialNo
		r.ProviderError = ""
		return s.repo.UpdateInvoiceRequest(txCtx, r)
	}); err != nil {
		log.Printf("[Invoice] auto-issue: save serial_no failed: invoice_id=%d err=%v", invoiceRequestID, err)
		s.revertIssuingToApproved(ctx, invoiceRequestID, fmt.Sprintf("保存流水号失败: %v", err))
	}
}

func (s *InvoiceService) revertIssuingToApproved(ctx context.Context, invoiceRequestID int64, errMsg string) {
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		r, err := s.repo.GetInvoiceRequestByID(txCtx, invoiceRequestID)
		if err != nil {
			return err
		}
		if r == nil {
			return nil
		}
		r.Status = InvoiceStatusApproved
		r.ProviderError = errMsg
		return s.repo.UpdateInvoiceRequest(txCtx, r)
	}); err != nil {
		log.Printf("[Invoice] auto-issue: revert to approved failed: invoice_id=%d err=%v", invoiceRequestID, err)
	}
}

// HandleNuonuoWebhook processes an incoming Nuonuo callback.
func (s *InvoiceService) HandleNuonuoWebhook(ctx context.Context, payload *NuonuoWebhookPayload) error {
	if s == nil || payload == nil {
		return nil
	}

	// Verify signature.
	if s.nuonuoService != nil && !s.nuonuoService.VerifyWebhookSignature(payload) {
		return infraerrors.BadRequest("INVOICE_WEBHOOK_SIGNATURE_INVALID", "Invalid webhook signature.")
	}

	if payload.SerialNo == "" {
		return infraerrors.BadRequest("INVOICE_WEBHOOK_MISSING_SERIAL", "Missing serialNo in webhook payload.")
	}

	req, err := s.repo.GetInvoiceRequestByProviderInvoiceID(ctx, payload.SerialNo)
	if err != nil {
		return err
	}
	if req == nil {
		// Not found — could be from a different system. Return 200 to prevent Nuonuo retries.
		log.Printf("[Invoice] webhook: unknown serial_no=%s, ignoring", payload.SerialNo)
		return nil
	}

	// Idempotency: already issued.
	if req.Status == InvoiceStatusIssued {
		return nil
	}

	// Invoice failed on Nuonuo side.
	if payload.InvoiceStatus == "4" {
		errMsg := payload.ErrMsg
		if errMsg == "" {
			errMsg = "诺诺开票失败"
		}
		s.revertIssuingToApproved(ctx, req.ID, errMsg)
		return nil
	}

	// Invoice succeeded — update to issued.
	var receiverEmail string
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		r, err := s.repo.GetInvoiceRequestByID(txCtx, req.ID)
		if err != nil {
			return err
		}
		if r == nil {
			return nil
		}
		now := time.Now()
		r.Status = InvoiceStatusIssued
		r.IssuedAt = &now
		r.InvoiceNumber = payload.InvoiceNo
		r.InvoicePDFURL = payload.PDFURL
		r.ProviderError = ""
		if err := s.repo.UpdateInvoiceRequest(txCtx, r); err != nil {
			return err
		}
		receiverEmail = r.ReceiverEmail
		return nil
	}); err != nil {
		return err
	}

	log.Printf("[Invoice] webhook: issued invoice_id=%d serial_no=%s invoice_no=%s", req.ID, payload.SerialNo, payload.InvoiceNo)

	// Send email notification (best-effort).
	if s.emailService != nil && receiverEmail != "" && payload.PDFURL != "" {
		if err := s.emailService.SendEmail(ctx, receiverEmail, "发票已开具", fmt.Sprintf(
			"您的发票已开具，下载链接：<a href=\"%s\">%s</a>", payload.PDFURL, payload.PDFURL,
		)); err != nil && !errors.Is(err, ErrEmailNotConfigured) {
			log.Printf("[Invoice] webhook: send email failed: invoice_id=%d to=%s err=%v", req.ID, receiverEmail, err)
		}
	}
	return nil
}

// AdminRetryAutoIssue triggers auto-issue for a previously failed approved invoice.
func (s *InvoiceService) AdminRetryAutoIssue(ctx context.Context, id int64) error {
	if s == nil || s.repo == nil {
		return nil
	}
	if !s.shouldAutoIssue() {
		return infraerrors.BadRequest("INVOICE_AUTO_ISSUE_DISABLED", "Auto-issue is not enabled.")
	}

	req, err := s.repo.GetInvoiceRequestByID(ctx, id)
	if err != nil {
		return err
	}
	if req == nil {
		return infraerrors.NotFound("INVOICE_NOT_FOUND", "Invoice request not found.")
	}
	if req.Status != InvoiceStatusApproved {
		return infraerrors.BadRequest("INVOICE_CANNOT_RETRY", "Only approved invoice requests can be retried.")
	}

	go s.autoIssueInvoiceAsync(context.Background(), id)
	return nil
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
	if s == nil || req == nil {
		return
	}

	recipients := parseEmailListFromEnv(invoiceRequestNotifyEmailsEnv)
	if len(recipients) == 0 {
		log.Printf("[Invoice] submit notify skipped: invoice_request_no=%s env_%s empty", req.InvoiceRequestNo, invoiceRequestNotifyEmailsEnv)
		return
	}
	if s.emailService == nil {
		log.Printf("[Invoice] submit notify skipped: invoice_request_no=%s email_service=nil", req.InvoiceRequestNo)
		return
	}

	taxNo := strings.TrimSpace(req.TaxNo)
	if taxNo == "" {
		taxNo = "-"
	}

	subject := fmt.Sprintf("开票申请已提交 [%s]", strings.TrimSpace(req.InvoiceRequestNo))
	body := fmt.Sprintf(
		"## 新的开票申请已提交\n\n- 抬头：%s\n- 税号：%s\n- 申请开票金额（CNY）：%.2f\n- 收票邮箱：%s\n",
		strings.TrimSpace(req.InvoiceTitle),
		taxNo,
		req.AmountCNYTotal,
		strings.TrimSpace(req.ReceiverEmail),
	)

	for _, to := range recipients {
		log.Printf("[Invoice] submit notify sending: invoice_request_no=%s to=%s", req.InvoiceRequestNo, to)
		if err := s.emailService.SendPlainTextEmail(ctx, to, subject, body); err != nil {
			if errors.Is(err, ErrEmailNotConfigured) {
				log.Printf("[Invoice] submit notify skipped: invoice_request_no=%s email not configured", req.InvoiceRequestNo)
				return
			}
			log.Printf("[Invoice] send submit notify email failed: invoice_request_no=%s to=%s err=%v", req.InvoiceRequestNo, to, err)
			continue
		}
		log.Printf("[Invoice] submit notify sent: invoice_request_no=%s to=%s", req.InvoiceRequestNo, to)
	}
}

func parseEmailListFromEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '，', ';', '；', ' ', '\t', '\n', '\r':
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
