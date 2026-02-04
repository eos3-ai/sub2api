package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type stubHTTPUpstream struct {
	calls                int
	enableTLSFingerprint bool
	userAgent            string
}

func (s *stubHTTPUpstream) Do(req *http.Request, proxyURL string, accountID int64, accountConcurrency int) (*http.Response, error) {
	return nil, errors.New("unexpected Do call")
}

func (s *stubHTTPUpstream) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, enableTLSFingerprint bool) (*http.Response, error) {
	s.calls++
	s.enableTLSFingerprint = enableTLSFingerprint
	if req != nil {
		s.userAgent = req.Header.Get("User-Agent")
	}

	// Minimal Claude SSE stream response that resolves quickly.
	body := io.NopCloser(strings.NewReader("data: {\"type\":\"message_stop\"}\n\n"))
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       body,
	}, nil
}

func newMonitorTestConfig() *config.Config {
	cfg := &config.Config{}
	// Keep URL validation permissive in unit tests.
	cfg.Security.URLAllowlist.Enabled = false
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	return cfg
}

func newMonitorTestAccount() *Account {
	return &Account{
		ID:          1,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":   "sk-test",
			"base_url":  "https://api.anthropic.com",
		},
	}
}

func TestAnthropicAPIKeyMonitor_UsesTLSFingerprintWhenEnabled(t *testing.T) {
	cfg := newMonitorTestConfig()
	cfg.Gateway.TLSFingerprint.Enabled = true

	upstream := &stubHTTPUpstream{}
	svc := NewAnthropicAPIKeyMonitorService(nil, upstream, nil, cfg)

	ok, msg, _ := svc.testAnthropicAPIKeyAccount(context.Background(), newMonitorTestAccount())
	if !ok {
		t.Fatalf("expected ok=true, got ok=false msg=%q", msg)
	}
	if upstream.calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", upstream.calls)
	}
	if !upstream.enableTLSFingerprint {
		t.Fatalf("expected enableTLSFingerprint=true, got false")
	}
	if !strings.HasPrefix(upstream.userAgent, "claude-cli/") {
		t.Fatalf("expected User-Agent to start with %q, got %q", "claude-cli/", upstream.userAgent)
	}
}

func TestAnthropicAPIKeyMonitor_DisablesTLSFingerprintWhenDisabled(t *testing.T) {
	cfg := newMonitorTestConfig()
	cfg.Gateway.TLSFingerprint.Enabled = false

	upstream := &stubHTTPUpstream{}
	svc := NewAnthropicAPIKeyMonitorService(nil, upstream, nil, cfg)

	ok, msg, _ := svc.testAnthropicAPIKeyAccount(context.Background(), newMonitorTestAccount())
	if !ok {
		t.Fatalf("expected ok=true, got ok=false msg=%q", msg)
	}
	if upstream.calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", upstream.calls)
	}
	if upstream.enableTLSFingerprint {
		t.Fatalf("expected enableTLSFingerprint=false, got true")
	}
}
