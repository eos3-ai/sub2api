package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type InvoiceRequestListFilter struct {
	UserID *int64
	Status string
	From   *time.Time
	To     *time.Time
}

type InvoiceRepository interface {
	ListEligibleOrders(ctx context.Context, userID int64, params pagination.PaginationParams, from, to *time.Time) ([]InvoiceOrderSnapshot, *pagination.PaginationResult, error)
	GetEligibleOrdersByOrderNos(ctx context.Context, userID int64, orderNos []string) ([]InvoiceOrderSnapshot, error)

	CreateInvoiceRequest(ctx context.Context, req *InvoiceRequest, orders []InvoiceOrderSnapshot) error
	UpdateInvoiceRequest(ctx context.Context, req *InvoiceRequest) error
	ListInvoiceRequests(ctx context.Context, params pagination.PaginationParams, filter InvoiceRequestListFilter) ([]InvoiceRequest, *pagination.PaginationResult, error)
	GetInvoiceRequestByID(ctx context.Context, id int64) (*InvoiceRequest, error)
	ListInvoiceOrderItems(ctx context.Context, invoiceRequestID int64) ([]InvoiceOrderItem, error)
	SetInvoiceOrderItemsActive(ctx context.Context, invoiceRequestID int64, active bool) error
	ListInvoiceOrderNosByRequestIDs(ctx context.Context, invoiceRequestIDs []int64) (map[int64][]string, error)

	GetInvoiceProfile(ctx context.Context, userID int64) (*InvoiceProfile, error)
	UpsertInvoiceProfile(ctx context.Context, profile *InvoiceProfile) error
}
