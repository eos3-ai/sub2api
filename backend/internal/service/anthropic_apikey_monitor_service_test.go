package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
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
			"api_key":  "sk-test",
			"base_url": "https://api.anthropic.com",
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

func TestAnthropicAPIKeyMonitor_SelectTargets_Default(t *testing.T) {
	cfg := newMonitorTestConfig()
	svc := NewAnthropicAPIKeyMonitorService(nil, nil, nil, cfg)

	accounts := []Account{
		{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive},
		{ID: 2, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive},
		{ID: 3, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive},
		{ID: 4, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusDisabled},
		{ID: 5, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive},
		{ID: 6, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusError},
	}

	targets := svc.selectTargets(accounts)
	got := extractAccountIDs(targets)
	want := []int64{1, 5, 6}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("selectTargets IDs = %v, want %v", got, want)
	}
}

func TestAnthropicAPIKeyMonitor_SelectTargets_IncludesErrorAccounts(t *testing.T) {
	cfg := newMonitorTestConfig()
	svc := NewAnthropicAPIKeyMonitorService(nil, nil, nil, cfg)

	accounts := []Account{
		{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusError},
		{ID: 2, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusDisabled},
		{ID: 3, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive},
	}

	targets := svc.selectTargets(accounts)
	got := extractAccountIDs(targets)
	want := []int64{1, 3}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("selectTargets IDs = %v, want %v", got, want)
	}
}

func TestAnthropicAPIKeyMonitor_SelectTargets_IncludeIDs(t *testing.T) {
	cfg := newMonitorTestConfig()
	cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.IncludeAccountIDs = []int64{2, 5, 999}
	svc := NewAnthropicAPIKeyMonitorService(nil, nil, nil, cfg)

	accounts := []Account{
		{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive},
		{ID: 2, Platform: PlatformAnthropic, Type: AccountTypeOAuth, Status: StatusActive},
		{ID: 5, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive},
	}

	targets := svc.selectTargets(accounts)
	got := extractAccountIDs(targets)
	want := []int64{5}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("selectTargets IDs = %v, want %v", got, want)
	}
}

func TestAnthropicAPIKeyMonitor_SelectTargets_ExcludeIDs(t *testing.T) {
	cfg := newMonitorTestConfig()
	cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.ExcludeAccountIDs = []int64{1}
	svc := NewAnthropicAPIKeyMonitorService(nil, nil, nil, cfg)

	accounts := []Account{
		{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive},
		{ID: 5, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive},
	}

	targets := svc.selectTargets(accounts)
	got := extractAccountIDs(targets)
	want := []int64{5}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("selectTargets IDs = %v, want %v", got, want)
	}
}

func TestAnthropicAPIKeyMonitor_SelectTargets_IncludeAndExclude(t *testing.T) {
	cfg := newMonitorTestConfig()
	cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.IncludeAccountIDs = []int64{1, 5}
	cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.ExcludeAccountIDs = []int64{5}
	svc := NewAnthropicAPIKeyMonitorService(nil, nil, nil, cfg)

	accounts := []Account{
		{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive},
		{ID: 5, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive},
	}

	targets := svc.selectTargets(accounts)
	got := extractAccountIDs(targets)
	want := []int64{1}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("selectTargets IDs = %v, want %v", got, want)
	}
}

func extractAccountIDs(accounts []Account) []int64 {
	if len(accounts) == 0 {
		return nil
	}
	out := make([]int64, 0, len(accounts))
	for i := range accounts {
		out = append(out, accounts[i].ID)
	}
	return out
}

// monitorRepoStub is a minimal AccountRepository stub for monitor applyResult tests.
// Only ClearError, UpdateExtra, and SetSchedulable are functional; all others panic.
type monitorRepoStub struct {
	clearErrorCalled    bool
	clearErrorID        int64
	clearErrorErr       error
	setSchedulableCalls []struct {
		id          int64
		schedulable bool
	}
	setSchedulableErr error
	updateExtraErr    error
}

func (m *monitorRepoStub) ClearError(ctx context.Context, id int64) error {
	m.clearErrorCalled = true
	m.clearErrorID = id
	return m.clearErrorErr
}

func (m *monitorRepoStub) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	return m.updateExtraErr
}

func (m *monitorRepoStub) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	m.setSchedulableCalls = append(m.setSchedulableCalls, struct {
		id          int64
		schedulable bool
	}{id, schedulable})
	return m.setSchedulableErr
}

func (m *monitorRepoStub) Create(ctx context.Context, account *Account) error {
	panic("unexpected Create")
}
func (m *monitorRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	panic("unexpected GetByID")
}
func (m *monitorRepoStub) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	panic("unexpected GetByIDs")
}
func (m *monitorRepoStub) ExistsByID(ctx context.Context, id int64) (bool, error) {
	panic("unexpected ExistsByID")
}
func (m *monitorRepoStub) GetByCRSAccountID(ctx context.Context, crsAccountID string) (*Account, error) {
	panic("unexpected GetByCRSAccountID")
}
func (m *monitorRepoStub) FindByExtraField(ctx context.Context, key string, value any) ([]Account, error) {
	panic("unexpected FindByExtraField")
}
func (m *monitorRepoStub) ListCRSAccountIDs(ctx context.Context) (map[string]int64, error) {
	panic("unexpected ListCRSAccountIDs")
}
func (m *monitorRepoStub) Update(ctx context.Context, account *Account) error {
	panic("unexpected Update")
}
func (m *monitorRepoStub) Delete(ctx context.Context, id int64) error {
	panic("unexpected Delete")
}
func (m *monitorRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	panic("unexpected List")
}
func (m *monitorRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, accountType, status, search string, groupID int64) ([]Account, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters")
}
func (m *monitorRepoStub) ListByGroup(ctx context.Context, groupID int64) ([]Account, error) {
	panic("unexpected ListByGroup")
}
func (m *monitorRepoStub) ListActive(ctx context.Context) ([]Account, error) {
	panic("unexpected ListActive")
}
func (m *monitorRepoStub) ListByPlatform(ctx context.Context, platform string) ([]Account, error) {
	panic("unexpected ListByPlatform")
}
func (m *monitorRepoStub) ListByPlatformForMonitor(ctx context.Context, platform string) ([]Account, error) {
	panic("unexpected ListByPlatformForMonitor")
}
func (m *monitorRepoStub) UpdateLastUsed(ctx context.Context, id int64) error {
	panic("unexpected UpdateLastUsed")
}
func (m *monitorRepoStub) BatchUpdateLastUsed(ctx context.Context, updates map[int64]time.Time) error {
	panic("unexpected BatchUpdateLastUsed")
}
func (m *monitorRepoStub) SetError(ctx context.Context, id int64, errorMsg string) error {
	panic("unexpected SetError")
}
func (m *monitorRepoStub) AutoPauseExpiredAccounts(ctx context.Context, now time.Time) (int64, error) {
	panic("unexpected AutoPauseExpiredAccounts")
}
func (m *monitorRepoStub) BindGroups(ctx context.Context, accountID int64, groupIDs []int64) error {
	panic("unexpected BindGroups")
}
func (m *monitorRepoStub) ListSchedulable(ctx context.Context) ([]Account, error) {
	panic("unexpected ListSchedulable")
}
func (m *monitorRepoStub) ListSchedulableByGroupID(ctx context.Context, groupID int64) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupID")
}
func (m *monitorRepoStub) ListSchedulableByPlatform(ctx context.Context, platform string) ([]Account, error) {
	panic("unexpected ListSchedulableByPlatform")
}
func (m *monitorRepoStub) ListSchedulableByGroupIDAndPlatform(ctx context.Context, groupID int64, platform string) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupIDAndPlatform")
}
func (m *monitorRepoStub) ListSchedulableByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	panic("unexpected ListSchedulableByPlatforms")
}
func (m *monitorRepoStub) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, groupID int64, platforms []string) ([]Account, error) {
	panic("unexpected ListSchedulableByGroupIDAndPlatforms")
}
func (m *monitorRepoStub) SetRateLimited(ctx context.Context, id int64, resetAt time.Time) error {
	panic("unexpected SetRateLimited")
}
func (m *monitorRepoStub) SetModelRateLimit(ctx context.Context, id int64, scope string, resetAt time.Time) error {
	panic("unexpected SetModelRateLimit")
}
func (m *monitorRepoStub) SetOverloaded(ctx context.Context, id int64, until time.Time) error {
	panic("unexpected SetOverloaded")
}
func (m *monitorRepoStub) SetTempUnschedulable(ctx context.Context, id int64, until time.Time, reason string) error {
	panic("unexpected SetTempUnschedulable")
}
func (m *monitorRepoStub) ClearTempUnschedulable(ctx context.Context, id int64) error {
	panic("unexpected ClearTempUnschedulable")
}
func (m *monitorRepoStub) ClearRateLimit(ctx context.Context, id int64) error {
	panic("unexpected ClearRateLimit")
}
func (m *monitorRepoStub) ClearAntigravityQuotaScopes(ctx context.Context, id int64) error {
	panic("unexpected ClearAntigravityQuotaScopes")
}
func (m *monitorRepoStub) ClearModelRateLimits(ctx context.Context, id int64) error {
	panic("unexpected ClearModelRateLimits")
}
func (m *monitorRepoStub) UpdateSessionWindow(ctx context.Context, id int64, start, end *time.Time, status string) error {
	panic("unexpected UpdateSessionWindow")
}
func (m *monitorRepoStub) BulkUpdate(ctx context.Context, ids []int64, updates AccountBulkUpdate) (int64, error) {
	panic("unexpected BulkUpdate")
}

// TestAnthropicAPIKeyMonitor_ApplyResult_ClearsErrorOnConsecutiveSuccesses verifies that
// applyResult calls ClearError when an error-status account reaches the success threshold.
func TestAnthropicAPIKeyMonitor_ApplyResult_ClearsErrorOnConsecutiveSuccesses(t *testing.T) {
	cfg := newMonitorTestConfig()
	cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.SuccessThreshold = 2

	repo := &monitorRepoStub{}
	svc := NewAnthropicAPIKeyMonitorService(repo, nil, nil, cfg)
	svc.leader = true

	acc := Account{
		ID:       42,
		Platform: PlatformAnthropic,
		Type:     AccountTypeAPIKey,
		Status:   StatusError,
	}
	res := anthropicAPIKeyMonitorResult{
		AccountID: acc.ID,
		Account:   acc,
		Success:   true,
	}

	now := time.Now().UTC()

	// First success — below threshold, no action.
	delta := svc.applyResult(context.Background(), now, res)
	if delta != 0 {
		t.Fatalf("expected delta=0 on first success, got %d", delta)
	}
	if repo.clearErrorCalled {
		t.Fatal("ClearError should not be called before threshold is reached")
	}

	// Second success — reaches threshold, ClearError should be called.
	delta = svc.applyResult(context.Background(), now, res)
	if delta != +1 {
		t.Fatalf("expected delta=+1 on threshold success, got %d", delta)
	}
	if !repo.clearErrorCalled {
		t.Fatal("expected ClearError to be called")
	}
	if repo.clearErrorID != acc.ID {
		t.Fatalf("ClearError called with id=%d, want %d", repo.clearErrorID, acc.ID)
	}
}

// TestAnthropicAPIKeyMonitor_ApplyResult_SkipsSchedulableDisableForErrorAccounts verifies
// that applyResult does NOT call SetSchedulable(false) for an account already in error state.
func TestAnthropicAPIKeyMonitor_ApplyResult_SkipsSchedulableDisableForErrorAccounts(t *testing.T) {
	cfg := newMonitorTestConfig()
	cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.FailureThreshold = 1

	repo := &monitorRepoStub{}
	svc := NewAnthropicAPIKeyMonitorService(repo, nil, nil, cfg)
	svc.leader = true

	acc := Account{
		ID:          99,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeAPIKey,
		Status:      StatusError,
		Schedulable: true, // schedulable=true but status=error — should not be toggled
	}
	res := anthropicAPIKeyMonitorResult{
		AccountID: acc.ID,
		Account:   acc,
		Success:   false,
		Message:   "auth failed",
	}

	now := time.Now().UTC()
	delta := svc.applyResult(context.Background(), now, res)

	if delta != 0 {
		t.Fatalf("expected delta=0 for error account failure, got %d", delta)
	}
	if len(repo.setSchedulableCalls) != 0 {
		t.Fatalf("expected SetSchedulable not to be called for error account, got %d calls", len(repo.setSchedulableCalls))
	}
}
