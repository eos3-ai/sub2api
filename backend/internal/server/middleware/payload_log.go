package middleware

import (
	"bufio"
	"bytes"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/repository/payloadlog"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

var payloadLogSequence atomic.Uint64

func PayloadLog(svc *service.PayloadLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if svc == nil || !svc.ShouldCapture() || isWebSocketUpgrade(c.Request) {
			c.Next()
			return
		}
		started := time.Now().UTC()
		contentType := c.GetHeader("Content-Type")
		captureRequest := isPayloadJSON(contentType)
		tee := &payloadTeeReadCloser{rc: c.Request.Body, capture: captureRequest}
		c.Request.Body = tee
		writer := &payloadLogWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()
		_ = tee.drainRemainder()
		requestID, _ := c.Request.Context().Value(ctxkey.ClientRequestID).(string)
		entry := &payloadlog.PayloadLog{ID: nextPayloadLogID(started), RequestID: requestID, CreatedAt: started, Method: c.Request.Method, Path: c.Request.URL.Path, StatusCode: uint16(writer.Status()), RequestContentType: contentType, ResponseContentType: writer.Header().Get("Content-Type"), RequestBytes: uint64(tee.count), ResponseBytes: uint64(writer.body.Len())}
		entry.IsStream = strings.Contains(strings.ToLower(entry.ResponseContentType), "text/event-stream")
		if v, ok := c.Get("ops_model"); ok {
			stringOK(v, &entry.ModelName)
		}
		if v, ok := c.Get("ops_stream"); ok {
			if b, ok := v.(bool); ok {
				entry.IsStream = b
			}
		}
		if v, ok := c.Get("ops_account_id"); ok {
			entry.AccountID = uint64Value(v)
		}
		if v, ok := c.Get(string(ContextKeyAPIKey)); ok {
			if key, ok := v.(*service.APIKey); ok && key != nil {
				entry.APIKeyID = uint64(key.ID)
				entry.APIKeyName = key.Name
				if key.User != nil {
					entry.Username = key.User.Username
					if entry.Username == "" {
						entry.Username = key.User.Email
					}
				}
			}
		}
		req := tee.body.Bytes()
		if !captureRequest {
			req = nil
		}
		svc.TryEnqueue(service.PayloadLogJob{Log: entry, RequestRaw: append([]byte(nil), req...), ResponseRaw: append([]byte(nil), writer.body.Bytes()...)})
	}
}

func nextPayloadLogID(t time.Time) uint64 {
	return uint64(t.UnixMilli())<<20 | payloadLogSequence.Add(1)&((1<<20)-1)
}
func stringOK(v any, dst *string) {
	if s, ok := v.(string); ok {
		*dst = s
	}
}
func uint64Value(v any) uint64 {
	switch x := v.(type) {
	case int:
		return uint64(x)
	case int64:
		return uint64(x)
	case uint64:
		return x
	case uint:
		return uint64(x)
	}
	return 0
}
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket")
}
func isPayloadJSON(value string) bool {
	if value == "" {
		return true
	}
	m, _, err := mime.ParseMediaType(value)
	if err != nil {
		m = strings.TrimSpace(strings.SplitN(value, ";", 2)[0])
	}
	return m == "application/json" || strings.HasSuffix(m, "+json")
}

type payloadTeeReadCloser struct {
	rc      io.ReadCloser
	body    bytes.Buffer
	count   int64
	capture bool
}

func (t *payloadTeeReadCloser) Read(p []byte) (int, error) {
	n, err := t.rc.Read(p)
	t.count += int64(n)
	if t.capture && n > 0 {
		_, _ = t.body.Write(p[:n])
	}
	return n, err
}
func (t *payloadTeeReadCloser) Close() error          { return t.rc.Close() }
func (t *payloadTeeReadCloser) drainRemainder() error { _, err := io.Copy(io.Discard, t); return err }

type payloadLogWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *payloadLogWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	if n > 0 {
		_, _ = w.body.Write(p[:n])
	}
	return n, err
}
func (w *payloadLogWriter) WriteString(s string) (int, error) {
	n, err := w.ResponseWriter.WriteString(s)
	if n > 0 {
		_, _ = w.body.WriteString(s[:n])
	}
	return n, err
}
func (w *payloadLogWriter) Flush() { w.ResponseWriter.Flush() }
func (w *payloadLogWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.Hijack()
}
func (w *payloadLogWriter) CloseNotify() <-chan bool    { return w.ResponseWriter.CloseNotify() }
func (w *payloadLogWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

var _ interface{ Unwrap() http.ResponseWriter } = (*payloadLogWriter)(nil)
var _ http.Flusher = (*payloadLogWriter)(nil)
var _ http.Hijacker = (*payloadLogWriter)(nil)
