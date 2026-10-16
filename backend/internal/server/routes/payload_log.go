package routes

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/repository/payloadlog"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func registerPayloadLogRoutes(admin *gin.RouterGroup, svc *service.PayloadLogService) {
	if svc == nil || !svc.FeatureEnabled() {
		return
	}
	g := admin.Group("/payload-logs")
	g.GET("/settings", func(c *gin.Context) { success(c, svc.Status()) })
	g.PUT("/settings", func(c *gin.Context) {
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			failure(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		if err := svc.SetEnabled(c.Request.Context(), req.Enabled); err != nil {
			payloadError(c, err)
			return
		}
		success(c, svc.Status())
	})
	g.GET("", func(c *gin.Context) {
		f, p, err := payloadListParams(c)
		if err != nil {
			failure(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		items, total, err := svc.List(c.Request.Context(), f, p)
		if err != nil {
			payloadError(c, err)
			return
		}
		pages := int64(0)
		if p.PageSize > 0 {
			pages = int64(math.Ceil(float64(total) / float64(p.PageSize)))
		}
		success(c, gin.H{"items": payloadMetaJSON(items), "total": total, "page": p.Page, "page_size": p.PageSize, "pages": pages})
	})
	g.GET("/:request_id", func(c *gin.Context) {
		tr, err := payloadTimeRange(c, false)
		if err != nil {
			failure(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		items, err := svc.GetByRequestID(c.Request.Context(), c.Param("request_id"), tr)
		if err != nil {
			payloadError(c, err)
			return
		}
		out := make([]gin.H, 0, len(items))
		for _, x := range items {
			m := payloadLogJSON(x)
			m["request_body"] = x.RequestBody
			m["response_body"] = x.ResponseBody
			out = append(out, m)
		}
		success(c, out)
	})
}
func success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": data})
}
func failure(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"code": code, "message": msg})
}
func payloadError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrPayloadLogStoreUnavailable) {
		failure(c, http.StatusServiceUnavailable, "PAYLOAD_LOG_STORE_UNAVAILABLE", err.Error())
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		failure(c, http.StatusGatewayTimeout, "PAYLOAD_LOG_QUERY_TIMEOUT", err.Error())
		return
	}
	failure(c, http.StatusInternalServerError, "PAYLOAD_LOG_ERROR", err.Error())
}
func payloadListParams(c *gin.Context) (payloadlog.PayloadLogFilter, payloadlog.Pagination, error) {
	page := queryInt(c, "page", queryInt(c, "p", 1))
	size := queryInt(c, "page_size", 20)
	if page < 1 || size < 1 || size > 100 {
		return payloadlog.PayloadLogFilter{}, payloadlog.Pagination{}, errors.New("invalid pagination")
	}
	tr, err := payloadTimeRange(c, true)
	if err != nil {
		return payloadlog.PayloadLogFilter{}, payloadlog.Pagination{}, err
	}
	return payloadlog.PayloadLogFilter{Username: c.Query("username"), APIKeyName: firstQuery(c, "api_key_name", "token_name"), ModelName: c.Query("model_name"), Start: tr.Start, End: tr.End}, payloadlog.Pagination{Page: page, PageSize: size}, nil
}
func payloadTimeRange(c *gin.Context, defaults bool) (*payloadlog.TimeRange, error) {
	now := time.Now()
	var start, end *time.Time
	if s := c.Query("start_timestamp"); s != "" {
		n, e := strconv.ParseInt(s, 10, 64)
		if e != nil {
			return nil, e
		}
		t := time.Unix(n, 0)
		start = &t
	}
	if s := c.Query("end_timestamp"); s != "" {
		n, e := strconv.ParseInt(s, 10, 64)
		if e != nil {
			return nil, e
		}
		t := time.Unix(n, 0)
		end = &t
	}
	if defaults && start == nil {
		t := now.Add(-24 * time.Hour)
		start = &t
	}
	if defaults && end == nil {
		end = &now
	}
	if start != nil && end != nil {
		if end.Before(*start) {
			return nil, errors.New("end_timestamp is before start_timestamp")
		}
		if end.Sub(*start) > 31*24*time.Hour {
			return nil, errors.New("time range exceeds 31 days")
		}
	}
	return &payloadlog.TimeRange{Start: start, End: end}, nil
}
func queryInt(c *gin.Context, key string, d int) int {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		return d
	}
	n, e := strconv.Atoi(v)
	if e != nil {
		return -1
	}
	return n
}
func firstQuery(c *gin.Context, keys ...string) string {
	for _, k := range keys {
		if v := c.Query(k); v != "" {
			return v
		}
	}
	return ""
}
func payloadMetaJSON(items []payloadlog.PayloadLogMeta) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, x := range items {
		out = append(out, gin.H{"id": x.ID, "request_id": x.RequestID, "created_at": x.CreatedAt.Unix(), "username": x.Username, "api_key_id": x.APIKeyID, "api_key_name": x.APIKeyName, "account_id": x.AccountID, "model_name": x.ModelName, "method": x.Method, "path": x.Path, "status_code": x.StatusCode, "is_stream": x.IsStream, "request_content_type": x.RequestContentType, "response_content_type": x.ResponseContentType, "request_bytes": x.RequestBytes, "response_bytes": x.ResponseBytes})
	}
	return out
}
func payloadLogJSON(x *payloadlog.PayloadLog) gin.H {
	return gin.H{"id": x.ID, "request_id": x.RequestID, "created_at": x.CreatedAt.Unix(), "username": x.Username, "api_key_id": x.APIKeyID, "api_key_name": x.APIKeyName, "account_id": x.AccountID, "model_name": x.ModelName, "method": x.Method, "path": x.Path, "status_code": x.StatusCode, "is_stream": x.IsStream, "request_content_type": x.RequestContentType, "response_content_type": x.ResponseContentType, "request_bytes": x.RequestBytes, "response_bytes": x.ResponseBytes}
}
