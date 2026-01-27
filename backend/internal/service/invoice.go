package service

import "time"

const (
	InvoiceStatusSubmitted = "submitted"
	InvoiceStatusApproved  = "approved"
	InvoiceStatusRejected  = "rejected"
	InvoiceStatusIssued    = "issued"
	InvoiceStatusCancelled = "cancelled"
)

const (
	InvoiceTypeNormal  = "normal"  // 电子普票
	InvoiceTypeSpecial = "special" // 专票
)

const (
	InvoiceBuyerTypePersonal = "personal"
	InvoiceBuyerTypeCompany  = "company"
)

type InvoiceRequest struct {
	ID               int64
	InvoiceRequestNo string
	UserID           int64
	Status           string

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

	AmountCNYTotal float64
	TotalUSDTotal  float64

	ReviewedBy   *int64
	ReviewedAt   *time.Time
	RejectReason string

	IssuedBy      *int64
	IssuedAt      *time.Time
	InvoiceNumber string
	InvoiceDate   *time.Time
	InvoicePDFURL string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type InvoiceOrderItem struct {
	ID               int64
	InvoiceRequestID int64
	PaymentOrderID   int64
	OrderNo          string
	AmountCNY        float64
	TotalUSD         float64
	Active           bool
	CreatedAt        time.Time

	// Joined from payment_orders for display.
	PaidAt *time.Time
}

type InvoiceProfile struct {
	ID     int64
	UserID int64

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

	CreatedAt time.Time
	UpdatedAt time.Time
}
