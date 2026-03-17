package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	openaiAPIKeyMonitorLeaderLockKey = "gateway:scheduling:openai_apikey_monitor:leader"
)

const (
	openaiAPIKeyMonitorExtraAutoDisabledKey    = "openai_apikey_monitor_auto_disabled"
	openaiAPIKeyMonitorExtraDisabledAtKey      = "openai_apikey_monitor_disabled_at"
	openaiAPIKeyMonitorExtraDisabledReasonKey  = "openai_apikey_monitor_disabled_reason"
	openaiAPIKeyMonitorExtraRecoveredAtKey     = "openai_apikey_monitor_recovered_at"
	openaiAPIKeyMonitorExtraRecoveredReasonKey = "openai_apikey_monitor_recovered_reason"
)

type OpenAIAPIKeyMonitorService struct {
	accountRepo  AccountRepository
	httpUpstream HTTPUpstream
	redisClient  *redis.Client
	cfg          *config.Config
	testService  *AccountTestService

	instanceID        string
	distributedLockOn bool
	warnNoRedisOnce   sync.Once

	startOnce sync.Once
	stopOnce  sync.Once
	stopCtx   context.Context
	stop      context.CancelFunc
	wg        sync.WaitGroup

	leader bool

	state map[int64]*openaiAPIKeyMonitorState

	lastAvailableAlertAt time.Time

	dingtalk *DingtalkService
}

type openaiAPIKeyMonitorState struct {
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	LastError            string
	LastCheckedAt        time.Time
}

type openaiAPIKeyMonitorResult struct {
	AccountID int64
	Account   Account
	Success   bool
	Message   string
	Latency   time.Duration
}

func NewOpenAIAPIKeyMonitorService(
	accountRepo AccountRepository,
	httpUpstream HTTPUpstream,
	redisClient *redis.Client,
	cfg *config.Config,
) *OpenAIAPIKeyMonitorService {
	lockOn := cfg == nil || strings.TrimSpace(cfg.RunMode) != config.RunModeSimple
	return &OpenAIAPIKeyMonitorService{
		accountRepo:       accountRepo,
		httpUpstream:      httpUpstream,
		redisClient:       redisClient,
		cfg:               cfg,
		testService:       NewAccountTestService(accountRepo, nil, nil, httpUpstream, cfg),
		instanceID:        uuid.NewString(),
		distributedLockOn: lockOn,
		warnNoRedisOnce:   sync.Once{},
		startOnce:         sync.Once{},
		stopOnce:          sync.Once{},
		stopCtx:           nil,
		stop:              nil,
		wg:                sync.WaitGroup{},
		leader:            false,
		state:             map[int64]*openaiAPIKeyMonitorState{},
		dingtalk:          NewDingtalkService(cfg),
	}
}

func (s *OpenAIAPIKeyMonitorService) Start() {
	s.StartWithContext(context.Background())
}

func (s *OpenAIAPIKeyMonitorService) StartWithContext(ctx context.Context) {
	if s == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if s.cfg == nil {
		slog.Warn("openai_apikey_monitor_config_missing")
		return
	}
	if !s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.Enabled {
		return
	}
	if s.accountRepo == nil || s.httpUpstream == nil {
		slog.Warn("openai_apikey_monitor_missing_deps")
		return
	}

	s.startOnce.Do(func() {
		s.stopCtx, s.stop = context.WithCancel(ctx)
		s.wg.Add(1)
		go s.run()
		slog.Info(
			"openai_apikey_monitor_started",
			"interval", s.effectiveInterval().String(),
			"failure_threshold", s.effectiveFailureThreshold(),
			"success_threshold", s.effectiveSuccessThreshold(),
			"request_timeout", s.effectiveRequestTimeout().String(),
			"max_concurrency", s.effectiveMaxConcurrency(),
			"include_account_ids", s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.IncludeAccountIDs,
			"exclude_account_ids", s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.ExcludeAccountIDs,
		)
	})
}

func (s *OpenAIAPIKeyMonitorService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.stop != nil {
			s.stop()
		}
	})
	s.wg.Wait()
	s.releaseLeaderLockBestEffort()
	slog.Info("openai_apikey_monitor_stopped")
}

func (s *OpenAIAPIKeyMonitorService) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.effectiveInterval())
	defer ticker.Stop()

	// Run once on startup.
	s.runOnce()

	for {
		select {
		case <-ticker.C:
			s.runOnce()
		case <-s.stopCtx.Done():
			return
		}
	}
}

func (s *OpenAIAPIKeyMonitorService) runOnce() {
	if s == nil || s.cfg == nil || s.accountRepo == nil || s.httpUpstream == nil {
		return
	}

	// Ensure leadership is stable; consecutive counters are in-memory.
	leader := s.acquireOrRefreshLeaderLock()
	if !leader {
		if s.leader {
			s.leader = false
			s.resetState()
			slog.Info("openai_apikey_monitor_leader_lost")
		}
		return
	}
	if !s.leader {
		s.leader = true
		s.resetState()
		slog.Info("openai_apikey_monitor_leader_acquired")
	}

	ctx := s.stopCtx
	accounts, err := s.accountRepo.ListByPlatformForMonitor(ctx, PlatformOpenAI)
	if err != nil {
		slog.Warn("openai_apikey_monitor_list_accounts_failed", "error", err)
		return
	}

	targets := s.selectTargets(accounts)
	if len(targets) == 0 {
		return
	}

	s.extendLeaderLockTTL(len(targets))

	now := time.Now().UTC()

	maxConc := s.effectiveMaxConcurrency()
	sem := make(chan struct{}, maxConc)
	results := make(chan openaiAPIKeyMonitorResult, len(targets))

	var wg sync.WaitGroup
	for i := range targets {
		acc := targets[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ok, msg, latency := s.testOpenAIAPIKeyAccount(ctx, &acc)
			results <- openaiAPIKeyMonitorResult{
				AccountID: acc.ID,
				Account:   acc,
				Success:   ok,
				Message:   msg,
				Latency:   latency,
			}
		}()
	}

	wg.Wait()
	close(results)

	// Prune stale entries (deleted accounts, type changes, etc.).
	seen := make(map[int64]struct{}, len(targets))
	for i := range targets {
		seen[targets[i].ID] = struct{}{}
	}
	for id := range s.state {
		if _, ok := seen[id]; !ok {
			delete(s.state, id)
		}
	}

	// Count schedulable accounts before applying results.
	// Exclude error accounts: even if schedulable=true, they are not usable.
	schedulable := 0
	for i := range targets {
		if targets[i].Status == StatusActive && targets[i].Schedulable {
			schedulable++
		}
	}

	for res := range results {
		schedulable += s.applyResult(ctx, now, res)
	}

	// Alert when available schedulable accounts fall to or below the configured threshold.
	alertThreshold := s.effectiveAvailableAccountAlertThreshold()
	if alertThreshold > 0 && schedulable <= alertThreshold {
		const availableAlertCooldown = 5 * time.Minute
		if now.Sub(s.lastAvailableAlertAt) >= availableAlertCooldown {
			s.lastAvailableAlertAt = now
			s.sendLowAvailableAccountAlert(schedulable, alertThreshold, now)
		}
	}
}

func (s *OpenAIAPIKeyMonitorService) resetState() {
	s.state = map[int64]*openaiAPIKeyMonitorState{}
}

func (s *OpenAIAPIKeyMonitorService) selectTargets(accounts []Account) []Account {
	if s == nil {
		return nil
	}

	includeIDs := map[int64]struct{}{}
	excludeIDs := map[int64]struct{}{}
	if s.cfg != nil {
		for _, id := range s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.IncludeAccountIDs {
			includeIDs[id] = struct{}{}
		}
		for _, id := range s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.ExcludeAccountIDs {
			excludeIDs[id] = struct{}{}
		}
	}
	hasInclude := len(includeIDs) > 0

	targets := make([]Account, 0, len(accounts))
	for i := range accounts {
		acc := accounts[i]
		if acc.Platform != PlatformOpenAI {
			continue
		}
		if acc.Type != AccountTypeAPIKey {
			continue
		}
		// Active and error accounts; error accounts may be restored by this monitor.
		if acc.Status != StatusActive && acc.Status != StatusError {
			continue
		}
		if hasInclude {
			if _, ok := includeIDs[acc.ID]; !ok {
				continue
			}
		}
		if _, ok := excludeIDs[acc.ID]; ok {
			continue
		}
		targets = append(targets, acc)
	}
	return targets
}

func (s *OpenAIAPIKeyMonitorService) effectiveInterval() time.Duration {
	if s == nil || s.cfg == nil {
		return 10 * time.Second
	}
	if d := s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.Interval; d > 0 {
		return d
	}
	return 10 * time.Second
}

func (s *OpenAIAPIKeyMonitorService) effectiveFailureThreshold() int {
	if s == nil || s.cfg == nil {
		return 6
	}
	if n := s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.FailureThreshold; n > 0 {
		return n
	}
	return 6
}

func (s *OpenAIAPIKeyMonitorService) effectiveSuccessThreshold() int {
	if s == nil || s.cfg == nil {
		return 6
	}
	if n := s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.SuccessThreshold; n > 0 {
		return n
	}
	return 6
}

func (s *OpenAIAPIKeyMonitorService) effectiveRequestTimeout() time.Duration {
	if s == nil || s.cfg == nil {
		return 8 * time.Second
	}
	if d := s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.RequestTimeout; d > 0 {
		return d
	}
	return 8 * time.Second
}

func (s *OpenAIAPIKeyMonitorService) effectiveMaxConcurrency() int {
	if s == nil || s.cfg == nil {
		return 4
	}
	n := s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.MaxConcurrency
	if n <= 0 {
		return 4
	}
	if n > 64 {
		return 64
	}
	return n
}

func (s *OpenAIAPIKeyMonitorService) effectiveAvailableAccountAlertThreshold() int {
	if s == nil || s.cfg == nil {
		return 0
	}
	return s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.AvailableAccountAlertThreshold
}

func (s *OpenAIAPIKeyMonitorService) effectiveModelID() string {
	if s == nil || s.cfg == nil {
		return ""
	}
	return strings.TrimSpace(s.cfg.Gateway.Scheduling.OpenAIAPIKeyMonitor.ModelID)
}

func (s *OpenAIAPIKeyMonitorService) testOpenAIAPIKeyAccount(ctx context.Context, account *Account) (bool, string, time.Duration) {
	if s == nil || account == nil {
		return false, "nil account", 0
	}

	startedAt := time.Now()

	// Use per-request timeout to avoid piling up goroutines.
	reqTimeout := s.effectiveRequestTimeout()
	reqCtx := ctx
	var cancel context.CancelFunc
	if reqTimeout > 0 {
		reqCtx, cancel = context.WithTimeout(ctx, reqTimeout)
		defer cancel()
	}

	modelID := s.effectiveModelID()
	if s.testService == nil {
		s.testService = NewAccountTestService(s.accountRepo, nil, nil, s.httpUpstream, s.cfg)
	}

	// Reuse the same logic as the admin endpoint:
	//   POST /api/v1/admin/accounts/:id/test
	// but make it lightweight for monitoring by overriding max_output_tokens=1.
	gin.SetMode(gin.TestMode)
	w := newLimitedResponseWriter(8 * 1024)
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequestWithContext(reqCtx, http.MethodPost, fmt.Sprintf("http://localhost/api/v1/admin/accounts/%d/test", account.ID), nil)
	req.Header.Set("content-type", "application/json")
	c.Request = req

	if err := s.testService.testOpenAIAccountConnection(c, account, modelID, 1); err != nil {
		msg := strings.TrimSpace(err.Error())
		if msg == "" {
			msg = "test failed"
		}
		return false, msg, time.Since(startedAt)
	}

	return true, "", time.Since(startedAt)
}

// applyResult processes a single monitor result and returns the change in schedulable count:
// -1 if the account was just disabled, +1 if re-enabled, 0 if unchanged.
func (s *OpenAIAPIKeyMonitorService) applyResult(ctx context.Context, now time.Time, res openaiAPIKeyMonitorResult) int {
	if res.AccountID <= 0 {
		return 0
	}

	st := s.state[res.AccountID]
	if st == nil {
		st = &openaiAPIKeyMonitorState{}
		s.state[res.AccountID] = st
	}
	st.LastCheckedAt = now

	if res.Success {
		st.ConsecutiveSuccesses++
		st.ConsecutiveFailures = 0
		st.LastError = ""

		threshold := s.effectiveSuccessThreshold()
		if st.ConsecutiveSuccesses < threshold {
			return 0
		}

		// Case 1: account is in error state — call ClearError to restore it to active.
		if res.Account.Status == StatusError {
			if err := s.accountRepo.ClearError(ctx, res.AccountID); err != nil {
				slog.Warn("openai_apikey_monitor_clear_error_failed", "account_id", res.AccountID, "error", err)
				return 0
			}
			updates := map[string]any{
				openaiAPIKeyMonitorExtraRecoveredAtKey:     now.Format(time.RFC3339),
				openaiAPIKeyMonitorExtraRecoveredReasonKey: fmt.Sprintf("consecutive_successes=%d", threshold),
			}
			if err := s.accountRepo.UpdateExtra(ctx, res.AccountID, updates); err != nil {
				slog.Warn("openai_apikey_monitor_update_extra_on_error_recovery_failed", "account_id", res.AccountID, "error", err)
			}
			slog.Info("openai_apikey_monitor_error_cleared", "account_id", res.AccountID, "consecutive_successes", threshold)
			st.ConsecutiveSuccesses = 0
			return +1
		}

		// Case 2: recovery — only auto-resume accounts that were auto-disabled by this monitor.
		if res.Account.Schedulable {
			return 0
		}
		if !getExtraBool(res.Account.Extra, openaiAPIKeyMonitorExtraAutoDisabledKey) {
			return 0
		}

		if err := s.accountRepo.SetSchedulable(ctx, res.AccountID, true); err != nil {
			slog.Warn("openai_apikey_monitor_enable_schedulable_failed", "account_id", res.AccountID, "error", err)
			return 0
		}
		updates := map[string]any{
			openaiAPIKeyMonitorExtraAutoDisabledKey:    false,
			openaiAPIKeyMonitorExtraRecoveredAtKey:     now.Format(time.RFC3339),
			openaiAPIKeyMonitorExtraRecoveredReasonKey: fmt.Sprintf("consecutive_successes=%d", threshold),
		}
		if err := s.accountRepo.UpdateExtra(ctx, res.AccountID, updates); err != nil {
			slog.Warn("openai_apikey_monitor_update_extra_on_recovery_failed", "account_id", res.AccountID, "error", err)
		}

		// Reset counters to avoid immediate flip-flop on transient next failures.
		st.ConsecutiveSuccesses = 0
		return +1
	}

	// Failure path.
	st.ConsecutiveFailures++
	st.ConsecutiveSuccesses = 0
	st.LastError = strings.TrimSpace(res.Message)

	threshold := s.effectiveFailureThreshold()
	if st.ConsecutiveFailures < threshold {
		return 0
	}

	// Only stop scheduling if currently schedulable and not already in error state.
	if !res.Account.Schedulable || res.Account.Status == StatusError {
		return 0
	}

	if err := s.accountRepo.SetSchedulable(ctx, res.AccountID, false); err != nil {
		slog.Warn("openai_apikey_monitor_disable_schedulable_failed", "account_id", res.AccountID, "error", err)
		return 0
	}

	reason := st.LastError
	if reason == "" {
		reason = fmt.Sprintf("consecutive_failures=%d", threshold)
	}
	reason = truncateString(reason, 1500)

	updates := map[string]any{
		openaiAPIKeyMonitorExtraAutoDisabledKey:    true,
		openaiAPIKeyMonitorExtraDisabledAtKey:      now.Format(time.RFC3339),
		openaiAPIKeyMonitorExtraDisabledReasonKey:  reason,
		openaiAPIKeyMonitorExtraRecoveredAtKey:     nil,
		openaiAPIKeyMonitorExtraRecoveredReasonKey: nil,
	}
	if err := s.accountRepo.UpdateExtra(ctx, res.AccountID, updates); err != nil {
		slog.Warn("openai_apikey_monitor_update_extra_on_failure_failed", "account_id", res.AccountID, "error", err)
	}

	st.ConsecutiveFailures = 0
	return -1
}

func (s *OpenAIAPIKeyMonitorService) sendLowAvailableAccountAlert(schedulable, threshold int, now time.Time) {
	if s == nil || s.dingtalk == nil || !s.dingtalk.Enabled() {
		return
	}

	title := fmt.Sprintf("账号告警: 可用 OpenAI API-key 账号不足 (剩余 %d)", schedulable)

	sb := strings.Builder{}
	sb.WriteString("### 【账号告警】OpenAI API-key 可用调度账号数量不足\n\n")
	sb.WriteString("**当前可用账号数**：`")
	sb.WriteString(fmt.Sprintf("%d", schedulable))
	sb.WriteString("`  \n")
	sb.WriteString("**告警阈值**：`≤ ")
	sb.WriteString(fmt.Sprintf("%d", threshold))
	sb.WriteString("`  \n")
	sb.WriteString("**时间**：`")
	sb.WriteString(escapeInlineCode(now.Format(time.RFC3339)))
	sb.WriteString("`  \n")

	go func(title, text string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.dingtalk.SendMarkdown(ctx, title, text); err != nil {
			slog.Warn("openai_apikey_monitor_low_available_dingtalk_send_failed", "error", err)
		}
	}(title, sb.String())
}

func (s *OpenAIAPIKeyMonitorService) acquireOrRefreshLeaderLock() bool {
	if s == nil || !s.distributedLockOn {
		return true
	}
	if s.redisClient == nil {
		s.warnNoRedisOnce.Do(func() {
			slog.Warn("openai_apikey_monitor_redis_not_configured_running_without_lock")
		})
		return true
	}

	ctx, cancel := context.WithTimeout(s.stopCtx, 2*time.Second)
	defer cancel()

	ttl := s.baseLeaderLockTTL()

	ok, err := s.redisClient.SetNX(ctx, openaiAPIKeyMonitorLeaderLockKey, s.instanceID, ttl).Result()
	if err != nil {
		// Fail-closed to avoid duplicated toggles when Redis is flaky.
		slog.Warn("openai_apikey_monitor_leader_lock_setnx_failed", "error", err)
		return false
	}
	if ok {
		return true
	}

	// If another instance holds the lock, skip. If we already hold it, refresh TTL.
	owner, err := s.redisClient.Get(ctx, openaiAPIKeyMonitorLeaderLockKey).Result()
	if err != nil || strings.TrimSpace(owner) == "" {
		return false
	}
	if owner != s.instanceID {
		return false
	}
	_ = s.redisClient.Expire(ctx, openaiAPIKeyMonitorLeaderLockKey, ttl).Err()
	return true
}

func (s *OpenAIAPIKeyMonitorService) extendLeaderLockTTL(targetCount int) {
	if s == nil || !s.distributedLockOn || s.redisClient == nil || targetCount <= 0 {
		return
	}

	maxConc := s.effectiveMaxConcurrency()
	if maxConc <= 0 {
		maxConc = 1
	}
	batches := (targetCount + maxConc - 1) / maxConc
	estimate := time.Duration(batches) * s.effectiveRequestTimeout()
	ttl := estimate + 30*time.Second
	ttl = maxDuration(ttl, s.baseLeaderLockTTL())

	ctx, cancel := context.WithTimeout(s.stopCtx, 2*time.Second)
	defer cancel()

	owner, err := s.redisClient.Get(ctx, openaiAPIKeyMonitorLeaderLockKey).Result()
	if err != nil || owner != s.instanceID {
		return
	}
	_ = s.redisClient.Expire(ctx, openaiAPIKeyMonitorLeaderLockKey, ttl).Err()
}

func (s *OpenAIAPIKeyMonitorService) baseLeaderLockTTL() time.Duration {
	// Keep a stable leader (counters are in-memory) but avoid overly long failover windows.
	interval := s.effectiveInterval()
	ttl := 12 * interval // e.g. 10s interval -> 120s ttl
	ttl = maxDuration(ttl, 2*time.Minute)
	ttl = maxDuration(ttl, 3*s.effectiveRequestTimeout())
	return ttl
}

func (s *OpenAIAPIKeyMonitorService) releaseLeaderLockBestEffort() {
	if s == nil || !s.distributedLockOn || s.redisClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Best-effort: delete lock only if still owned by this instance.
	val, err := s.redisClient.Get(ctx, openaiAPIKeyMonitorLeaderLockKey).Result()
	if err != nil {
		return
	}
	if val != s.instanceID {
		return
	}
	_, _ = s.redisClient.Del(ctx, openaiAPIKeyMonitorLeaderLockKey).Result()
}
