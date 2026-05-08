package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fixedHTTPUpstream struct {
	statusCode int
	headers    http.Header
	body       string
	err        error
	calls      int
}

func (u *fixedHTTPUpstream) buildResponse() *http.Response {
	headers := http.Header{}
	if u.headers != nil {
		headers = u.headers.Clone()
	}
	return &http.Response{
		StatusCode: u.statusCode,
		Header:     headers,
		Body:       io.NopCloser(strings.NewReader(u.body)),
	}
}

func (u *fixedHTTPUpstream) Do(req *http.Request, proxyURL string, accountID int64, accountConcurrency int) (*http.Response, error) {
	u.calls++
	if u.err != nil {
		return nil, u.err
	}
	return u.buildResponse(), nil
}

func (u *fixedHTTPUpstream) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxyURL, accountID, accountConcurrency)
}

func newGatewayServiceForRetryRegression(upstream HTTPUpstream) *GatewayService {
	return &GatewayService{
		cfg: &config.Config{
			Security: config.SecurityConfig{
				URLAllowlist: config.URLAllowlistConfig{Enabled: false},
			},
		},
		httpUpstream:     upstream,
		rateLimitService: &RateLimitService{},
	}
}

func newAnthropicAPIKeyAccountNoPassthrough() *Account {
	return &Account{
		ID:          301,
		Name:        "anthropic-apikey",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key": "upstream-anthropic-key",
		},
		Status:      StatusActive,
		Schedulable: true,
	}
}

func TestGatewayService_Forward_AnthropicAPIKey400DoesNotBecomeRetryExhausted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))

	body := []byte(`{"model":"claude-3-5-sonnet-latest","messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]}`)
	parsed := &ParsedRequest{
		Body:  body,
		Model: "claude-3-5-sonnet-latest",
	}

	upstream := &fixedHTTPUpstream{
		statusCode: http.StatusBadRequest,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
			"x-request-id": []string{"rid-gateway-400"},
		},
		body: `{"type":"error","error":{"type":"invalid_request_error","message":"model not found"}}`,
	}

	svc := newGatewayServiceForRetryRegression(upstream)
	result, err := svc.Forward(context.Background(), c, newAnthropicAPIKeyAccountNoPassthrough(), parsed)

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, 1, upstream.calls, "400 invalid_request should not enter generic retry loop")
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "model not found")
	require.NotContains(t, rec.Body.String(), "Upstream request failed after retries")
	require.NotContains(t, err.Error(), "retries exhausted")
}

func TestGatewayService_Forward_AnthropicPassthrough400DoesNotBecomeRetryExhausted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(nil))

	body := []byte(`{"model":"claude-3-5-sonnet-latest","messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]}`)
	parsed := &ParsedRequest{
		Body:  body,
		Model: "claude-3-5-sonnet-latest",
	}

	upstream := &fixedHTTPUpstream{
		statusCode: http.StatusBadRequest,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
			"x-request-id": []string{"rid-pass-400"},
		},
		body: `{"type":"error","error":{"type":"invalid_request_error","message":"bad tool schema"}}`,
	}

	svc := newGatewayServiceForRetryRegression(upstream)
	result, err := svc.Forward(context.Background(), c, newAnthropicAPIKeyAccountForTest(), parsed)

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, 1, upstream.calls, "passthrough 400 should not enter generic retry loop")
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "bad tool schema")
	require.NotContains(t, rec.Body.String(), "Upstream request failed after retries")
	require.NotContains(t, err.Error(), "retries exhausted")
}
