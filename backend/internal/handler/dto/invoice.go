package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type InvoiceRequest struct {
	ID               int64      `json:"id"`
	InvoiceRequestNo string     `json:"invoice_request_no"`
	UserID           int64      `json:"user_id"`
	UserEmail        string     `json:"user_email,omitempty"` // admin list convenience
	Status           string     `json:"status"`
	InvoiceType      string     `json:"invoice_type"`
	BuyerType        string     `json:"buyer_type"`
	InvoiceTitle     string     `json:"invoice_title"`
	TaxNo            string     `json:"tax_no"`
	BuyerAddress     string     `json:"buyer_address"`
	BuyerPhone       string     `json:"buyer_phone"`
	BuyerBankName    string     `json:"buyer_bank_name"`
	BuyerBankAccount string     `json:"buyer_bank_account"`
	ReceiverEmail    string     `json:"receiver_email"`
	ReceiverPhone    string     `json:"receiver_phone"`
	InvoiceItemName  string     `json:"invoice_item_name"`
	Remark           string     `json:"remark"`
	AmountCNYTotal   float64    `json:"amount_cny_total"`
	TotalUSDTotal    float64    `json:"total_usd_total"`
	ReviewedBy       *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
	RejectReason     string     `json:"reject_reason,omitempty"`
	IssuedBy         *int64     `json:"issued_by,omitempty"`
	IssuedAt         *time.Time `json:"issued_at,omitempty"`
	InvoiceNumber    string     `json:"invoice_number,omitempty"`
	InvoiceDate      *time.Time `json:"invoice_date,omitempty"`
	InvoicePDFURL    string     `json:"invoice_pdf_url,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func InvoiceRequestFromService(r *service.InvoiceRequest) *InvoiceRequest {
	if r == nil {
		return nil
	}
	return &InvoiceRequest{
		ID:               r.ID,
		InvoiceRequestNo: r.InvoiceRequestNo,
		UserID:           r.UserID,
		Status:           r.Status,
		InvoiceType:      r.InvoiceType,
		BuyerType:        r.BuyerType,
		InvoiceTitle:     r.InvoiceTitle,
		TaxNo:            r.TaxNo,
		BuyerAddress:     r.BuyerAddress,
		BuyerPhone:       r.BuyerPhone,
		BuyerBankName:    r.BuyerBankName,
		BuyerBankAccount: r.BuyerBankAccount,
		ReceiverEmail:    r.ReceiverEmail,
		ReceiverPhone:    r.ReceiverPhone,
		InvoiceItemName:  r.InvoiceItemName,
		Remark:           r.Remark,
		AmountCNYTotal:   r.AmountCNYTotal,
		TotalUSDTotal:    r.TotalUSDTotal,
		ReviewedBy:       r.ReviewedBy,
		ReviewedAt:       r.ReviewedAt,
		RejectReason:     r.RejectReason,
		IssuedBy:         r.IssuedBy,
		IssuedAt:         r.IssuedAt,
		InvoiceNumber:    r.InvoiceNumber,
		InvoiceDate:      r.InvoiceDate,
		InvoicePDFURL:    r.InvoicePDFURL,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

type InvoiceOrderItem struct {
	ID             int64      `json:"id"`
	PaymentOrderID int64      `json:"payment_order_id"`
	OrderNo        string     `json:"order_no"`
	AmountCNY      float64    `json:"amount_cny"`
	TotalUSD       float64    `json:"total_usd"`
	Active         bool       `json:"active"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func InvoiceOrderItemFromService(i *service.InvoiceOrderItem) *InvoiceOrderItem {
	if i == nil {
		return nil
	}
	return &InvoiceOrderItem{
		ID:             i.ID,
		PaymentOrderID: i.PaymentOrderID,
		OrderNo:        i.OrderNo,
		AmountCNY:      i.AmountCNY,
		TotalUSD:       i.TotalUSD,
		Active:         i.Active,
		PaidAt:         i.PaidAt,
		CreatedAt:      i.CreatedAt,
	}
}

type InvoiceProfile struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	InvoiceType      string    `json:"invoice_type"`
	BuyerType        string    `json:"buyer_type"`
	InvoiceTitle     string    `json:"invoice_title"`
	TaxNo            string    `json:"tax_no"`
	BuyerAddress     string    `json:"buyer_address"`
	BuyerPhone       string    `json:"buyer_phone"`
	BuyerBankName    string    `json:"buyer_bank_name"`
	BuyerBankAccount string    `json:"buyer_bank_account"`
	ReceiverEmail    string    `json:"receiver_email"`
	ReceiverPhone    string    `json:"receiver_phone"`
	InvoiceItemName  string    `json:"invoice_item_name"`
	Remark           string    `json:"remark"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func InvoiceProfileFromService(p *service.InvoiceProfile) *InvoiceProfile {
	if p == nil {
		return nil
	}
	return &InvoiceProfile{
		ID:               p.ID,
		UserID:           p.UserID,
		InvoiceType:      p.InvoiceType,
		BuyerType:        p.BuyerType,
		InvoiceTitle:     p.InvoiceTitle,
		TaxNo:            p.TaxNo,
		BuyerAddress:     p.BuyerAddress,
		BuyerPhone:       p.BuyerPhone,
		BuyerBankName:    p.BuyerBankName,
		BuyerBankAccount: p.BuyerBankAccount,
		ReceiverEmail:    p.ReceiverEmail,
		ReceiverPhone:    p.ReceiverPhone,
		InvoiceItemName:  p.InvoiceItemName,
		Remark:           p.Remark,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}
