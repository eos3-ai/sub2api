package payloadlog

import (
	"context"
	"time"
)

type PayloadLog struct {
	ID                  uint64    `json:"id"`
	RequestID           string    `json:"request_id"`
	CreatedAt           time.Time `json:"-"`
	Username            string    `json:"username"`
	APIKeyID            uint64    `json:"api_key_id"`
	APIKeyName          string    `json:"api_key_name"`
	AccountID           uint64    `json:"account_id"`
	ModelName           string    `json:"model_name"`
	Method              string    `json:"method"`
	Path                string    `json:"path"`
	StatusCode          uint16    `json:"status_code"`
	IsStream            bool      `json:"is_stream"`
	RequestContentType  string    `json:"request_content_type"`
	ResponseContentType string    `json:"response_content_type"`
	RequestBytes        uint64    `json:"request_bytes"`
	ResponseBytes       uint64    `json:"response_bytes"`
	RequestBody         string    `json:"request_body"`
	ResponseBody        string    `json:"response_body"`
}

type PayloadLogMeta struct {
	ID                  uint64
	RequestID           string
	CreatedAt           time.Time
	Username            string
	APIKeyID            uint64
	APIKeyName          string
	AccountID           uint64
	ModelName           string
	Method              string
	Path                string
	StatusCode          uint16
	IsStream            bool
	RequestContentType  string
	ResponseContentType string
	RequestBytes        uint64
	ResponseBytes       uint64
}

type PayloadLogFilter struct {
	Username   string
	APIKeyName string
	ModelName  string
	Start      *time.Time
	End        *time.Time
}

type Pagination struct{ Page, PageSize int }
type TimeRange struct{ Start, End *time.Time }

type Store interface {
	InsertBatch(context.Context, []*PayloadLog) error
	List(context.Context, PayloadLogFilter, Pagination) ([]PayloadLogMeta, int64, error)
	GetByRequestID(context.Context, string, *TimeRange) ([]*PayloadLog, error)
	Ping(context.Context) error
	Close() error
}
