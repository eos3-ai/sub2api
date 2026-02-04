package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	anthropicAPIKeyMonitorLeaderLockKey = "gateway:scheduling:anthropic_apikey_monitor:leader"
)

const (
	anthropicAPIKeyMonitorExtraAutoDisabledKey  = "anthropic_apikey_monitor_auto_disabled"
	anthropicAPIKeyMonitorExtraDisabledAtKey    = "anthropic_apikey_monitor_disabled_at"
	anthropicAPIKeyMonitorExtraDisabledReasonKey = "anthropic_apikey_monitor_disabled_reason"
	anthropicAPIKeyMonitorExtraRecoveredAtKey   = "anthropic_apikey_monitor_recovered_at"
	anthropicAPIKeyMonitorExtraRecoveredReasonKey = "anthropic_apikey_monitor_recovered_reason"
)

type AnthropicAPIKeyMonitorService struct {
	accountRepo  AccountRepository
	httpUpstream HTTPUpstream
	redisClient  *redis.Client
	cfg          *config.Config

	instanceID        string
	distributedLockOn bool
	warnNoRedisOnce   sync.Once

	startOnce sync.Once
	stopOnce  sync.Once
	stopCtx   context.Context
	stop      context.CancelFunc
	wg        sync.WaitGroup

	leader bool

	state map[int64]*anthropicAPIKeyMonitorState

	dingtalk *DingtalkService
}

type anthropicAPIKeyMonitorState struct {
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	LastError            string
	LastCheckedAt        time.Time
}

type anthropicAPIKeyMonitorResult struct {
	AccountID int64
	Account   Account
	Success   bool
	Message   string
	Latency   time.Duration
}

func NewAnthropicAPIKeyMonitorService(
	accountRepo AccountRepository,
	httpUpstream HTTPUpstream,
	redisClient *redis.Client,
	cfg *config.Config,
) *AnthropicAPIKeyMonitorService {
	lockOn := cfg == nil || strings.TrimSpace(cfg.RunMode) != config.RunModeSimple
	return &AnthropicAPIKeyMonitorService{
		accountRepo:        accountRepo,
		httpUpstream:       httpUpstream,
		redisClient:        redisClient,
		cfg:               cfg,
		instanceID:         uuid.NewString(),
		distributedLockOn:  lockOn,
		warnNoRedisOnce:    sync.Once{},
		startOnce:          sync.Once{},
		stopOnce:           sync.Once{},
		stopCtx:            nil,
		stop:               nil,
		wg:                 sync.WaitGroup{},
		leader:             false,
		state:              map[int64]*anthropicAPIKeyMonitorState{},
		dingtalk:           NewDingtalkService(cfg),
	}
}

func (s *AnthropicAPIKeyMonitorService) Start() {
	s.StartWithContext(context.Background())
}

func (s *AnthropicAPIKeyMonitorService) StartWithContext(ctx context.Context) {
	if s == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if s.cfg == nil {
		slog.Warn("anthropic_apikey_monitor_config_missing")
		return
	}
	if !s.cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.Enabled {
		return
	}
	if s.accountRepo == nil || s.httpUpstream == nil {
		slog.Warn("anthropic_apikey_monitor_missing_deps")
		return
	}

	s.startOnce.Do(func() {
		s.stopCtx, s.stop = context.WithCancel(ctx)
		s.wg.Add(1)
		go s.run()
		slog.Info(
			"anthropic_apikey_monitor_started",
			"interval", s.effectiveInterval().String(),
			"failure_threshold", s.effectiveFailureThreshold(),
			"success_threshold", s.effectiveSuccessThreshold(),
			"request_timeout", s.effectiveRequestTimeout().String(),
			"max_concurrency", s.effectiveMaxConcurrency(),
		)
	})
}

func (s *AnthropicAPIKeyMonitorService) Stop() {
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
	slog.Info("anthropic_apikey_monitor_stopped")
}

func (s *AnthropicAPIKeyMonitorService) run() {
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

func (s *AnthropicAPIKeyMonitorService) runOnce() {
	if s == nil || s.cfg == nil || s.accountRepo == nil || s.httpUpstream == nil {
		return
	}

	// Ensure leadership is stable; consecutive counters are in-memory.
	leader := s.acquireOrRefreshLeaderLock()
	if !leader {
		if s.leader {
			s.leader = false
			s.resetState()
			slog.Info("anthropic_apikey_monitor_leader_lost")
		}
		return
	}
	if !s.leader {
		s.leader = true
		s.resetState()
		slog.Info("anthropic_apikey_monitor_leader_acquired")
	}

	ctx := s.stopCtx
	accounts, err := s.accountRepo.ListByPlatform(ctx, PlatformAnthropic)
	if err != nil {
		slog.Warn("anthropic_apikey_monitor_list_accounts_failed", "error", err)
		return
	}

	targets := make([]Account, 0, len(accounts))
	for i := range accounts {
		acc := accounts[i]
		if acc.Platform != PlatformAnthropic {
			continue
		}
		if acc.Type != AccountTypeAPIKey {
			continue
		}
		// Only active accounts; schedulable may be toggled by this monitor.
		if acc.Status != StatusActive {
			continue
		}
		targets = append(targets, acc)
	}
	if len(targets) == 0 {
		return
	}

	s.extendLeaderLockTTL(len(targets))

	now := time.Now().UTC()

	maxConc := s.effectiveMaxConcurrency()
	sem := make(chan struct{}, maxConc)
	results := make(chan anthropicAPIKeyMonitorResult, len(targets))

	var wg sync.WaitGroup
	for i := range targets {
		acc := targets[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ok, msg, latency := s.testAnthropicAPIKeyAccount(ctx, &acc)
			results <- anthropicAPIKeyMonitorResult{
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

	for res := range results {
		s.applyResult(ctx, now, res)
	}
}

func (s *AnthropicAPIKeyMonitorService) resetState() {
	s.state = map[int64]*anthropicAPIKeyMonitorState{}
}

func (s *AnthropicAPIKeyMonitorService) effectiveInterval() time.Duration {
	if s == nil || s.cfg == nil {
		return 10 * time.Second
	}
	if d := s.cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.Interval; d > 0 {
		return d
	}
	return 10 * time.Second
}

func (s *AnthropicAPIKeyMonitorService) effectiveFailureThreshold() int {
	if s == nil || s.cfg == nil {
		return 6
	}
	if n := s.cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.FailureThreshold; n > 0 {
		return n
	}
	return 6
}

func (s *AnthropicAPIKeyMonitorService) effectiveSuccessThreshold() int {
	if s == nil || s.cfg == nil {
		return 6
	}
	if n := s.cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.SuccessThreshold; n > 0 {
		return n
	}
	return 6
}

func (s *AnthropicAPIKeyMonitorService) effectiveRequestTimeout() time.Duration {
	if s == nil || s.cfg == nil {
		return 8 * time.Second
	}
	if d := s.cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.RequestTimeout; d > 0 {
		return d
	}
	return 8 * time.Second
}

func (s *AnthropicAPIKeyMonitorService) effectiveMaxConcurrency() int {
	if s == nil || s.cfg == nil {
		return 4
	}
	n := s.cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.MaxConcurrency
	if n <= 0 {
		return 4
	}
	if n > 64 {
		return 64
	}
	return n
}

func (s *AnthropicAPIKeyMonitorService) effectiveModelID() string {
	if s == nil || s.cfg == nil {
		return ""
	}
	return strings.TrimSpace(s.cfg.Gateway.Scheduling.AnthropicAPIKeyMonitor.ModelID)
}

func (s *AnthropicAPIKeyMonitorService) validateUpstreamBaseURL(raw string) (string, error) {
	if s == nil || s.cfg == nil {
		return "", fmt.Errorf("config is not available")
	}
	if !s.cfg.Security.URLAllowlist.Enabled {
		return urlvalidator.ValidateURLFormat(raw, s.cfg.Security.URLAllowlist.AllowInsecureHTTP)
	}
	normalized, err := urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{
		AllowedHosts:     s.cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     s.cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
	if err != nil {
		return "", err
	}
	return normalized, nil
}

func (s *AnthropicAPIKeyMonitorService) testAnthropicAPIKeyAccount(ctx context.Context, account *Account) (bool, string, time.Duration) {
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
	if modelID == "" {
		modelID = claude.DefaultMonitorModel
	}
	// Apply account-level model mapping when configured.
	modelID = account.GetMappedModel(modelID)

	apiKey := strings.TrimSpace(account.GetCredential("api_key"))
	if apiKey == "" {
		return false, "No API key available", time.Since(startedAt)
	}

	baseURL := strings.TrimSpace(account.GetBaseURL())
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return false, fmt.Sprintf("Invalid base URL: %s", err.Error()), time.Since(startedAt)
	}

	apiURL := strings.TrimSuffix(normalizedBaseURL, "/") + "/v1/messages"

	payload, err := createTestPayload(modelID)
	if err != nil {
		return false, "Failed to create test payload", time.Since(startedAt)
	}
	// Make the automated monitor lightweight: we only need to validate connectivity + auth.
	payload["max_tokens"] = 1
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(reqCtx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return false, "Failed to create request", time.Since(startedAt)
	}

	// Match the admin "test account" headers (Claude Code style).
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-beta", claude.DefaultBetaHeader)
	for key, value := range claude.DefaultHeaders {
		req.Header.Set(key, value)
	}
	req.Header.Set("x-api-key", apiKey)

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	enableTLSFingerprint := false
	if s.cfg != nil {
		enableTLSFingerprint = s.cfg.Gateway.TLSFingerprint.Enabled
	}

	resp, err := s.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, enableTLSFingerprint)
	if err != nil {
		return false, fmt.Sprintf("Request failed: %s", err.Error()), time.Since(startedAt)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = "(empty response)"
		}
		return false, fmt.Sprintf("API returned %d: %s", resp.StatusCode, msg), time.Since(startedAt)
	}

	ok, msg := consumeClaudeTestStream(resp.Body)
	return ok, msg, time.Since(startedAt)
}

func consumeClaudeTestStream(body io.Reader) (bool, string) {
	reader := bufio.NewReader(body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return true, ""
			}
			return false, fmt.Sprintf("Stream read error: %s", err.Error())
		}

		line = strings.TrimSpace(line)
		if line == "" || !sseDataPrefix.MatchString(line) {
			continue
		}

		jsonStr := sseDataPrefix.ReplaceAllString(line, "")
		if jsonStr == "[DONE]" {
			return true, ""
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			continue
		}

		eventType, _ := data["type"].(string)
		switch eventType {
		case "content_block_delta":
			if delta, ok := data["delta"].(map[string]any); ok {
				if text, ok := delta["text"].(string); ok && strings.TrimSpace(text) != "" {
					// Success fast-path: any streamed content implies auth + model routing is working.
					return true, ""
				}
			}
		case "message_stop":
			return true, ""
		case "error":
			errorMsg := "Unknown error"
			if errData, ok := data["error"].(map[string]any); ok {
				if msg, ok := errData["message"].(string); ok && strings.TrimSpace(msg) != "" {
					errorMsg = strings.TrimSpace(msg)
				}
			}
			return false, errorMsg
		}
	}
}

func (s *AnthropicAPIKeyMonitorService) applyResult(ctx context.Context, now time.Time, res anthropicAPIKeyMonitorResult) {
	if res.AccountID <= 0 {
		return
	}

	st := s.state[res.AccountID]
	if st == nil {
		st = &anthropicAPIKeyMonitorState{}
		s.state[res.AccountID] = st
	}
	st.LastCheckedAt = now

	if res.Success {
		st.ConsecutiveSuccesses++
		st.ConsecutiveFailures = 0
		st.LastError = ""

		threshold := s.effectiveSuccessThreshold()
		if st.ConsecutiveSuccesses < threshold {
			return
		}

		// Recovery: only auto-resume accounts that were auto-disabled by this monitor.
		if res.Account.Schedulable {
			return
		}
		if !getExtraBool(res.Account.Extra, anthropicAPIKeyMonitorExtraAutoDisabledKey) {
			return
		}

		if err := s.accountRepo.SetSchedulable(ctx, res.AccountID, true); err != nil {
			slog.Warn("anthropic_apikey_monitor_enable_schedulable_failed", "account_id", res.AccountID, "error", err)
			return
		}
		updates := map[string]any{
			anthropicAPIKeyMonitorExtraAutoDisabledKey:    false,
			anthropicAPIKeyMonitorExtraRecoveredAtKey:     now.Format(time.RFC3339),
			anthropicAPIKeyMonitorExtraRecoveredReasonKey: fmt.Sprintf("consecutive_successes=%d", threshold),
		}
		if err := s.accountRepo.UpdateExtra(ctx, res.AccountID, updates); err != nil {
			slog.Warn("anthropic_apikey_monitor_update_extra_on_recovery_failed", "account_id", res.AccountID, "error", err)
		}

		s.sendRecoveryAlert(res.Account, threshold, res.Latency, now)
		// Reset counters to avoid immediate flip-flop on transient next failures.
		st.ConsecutiveSuccesses = 0
		return
	}

	// Failure path.
	st.ConsecutiveFailures++
	st.ConsecutiveSuccesses = 0
	st.LastError = strings.TrimSpace(res.Message)

	threshold := s.effectiveFailureThreshold()
	if st.ConsecutiveFailures < threshold {
		return
	}

	// Only stop scheduling if currently schedulable.
	if !res.Account.Schedulable {
		return
	}

	if err := s.accountRepo.SetSchedulable(ctx, res.AccountID, false); err != nil {
		slog.Warn("anthropic_apikey_monitor_disable_schedulable_failed", "account_id", res.AccountID, "error", err)
		return
	}

	reason := st.LastError
	if reason == "" {
		reason = fmt.Sprintf("consecutive_failures=%d", threshold)
	}
	reason = truncateString(reason, 1500)

	updates := map[string]any{
		anthropicAPIKeyMonitorExtraAutoDisabledKey:    true,
		anthropicAPIKeyMonitorExtraDisabledAtKey:      now.Format(time.RFC3339),
		anthropicAPIKeyMonitorExtraDisabledReasonKey:  reason,
		anthropicAPIKeyMonitorExtraRecoveredAtKey:     nil,
		anthropicAPIKeyMonitorExtraRecoveredReasonKey: nil,
	}
	if err := s.accountRepo.UpdateExtra(ctx, res.AccountID, updates); err != nil {
		slog.Warn("anthropic_apikey_monitor_update_extra_on_failure_failed", "account_id", res.AccountID, "error", err)
	}

	s.sendAbnormalAlert(res.Account, threshold, reason, res.Latency, now)
	st.ConsecutiveFailures = 0
}

func getExtraBool(extra map[string]any, key string) bool {
	if extra == nil {
		return false
	}
	v, ok := extra[key]
	if !ok || v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func (s *AnthropicAPIKeyMonitorService) sendAbnormalAlert(account Account, threshold int, reason string, latency time.Duration, now time.Time) {
	if s == nil || s.dingtalk == nil || !s.dingtalk.Enabled() {
		return
	}

	name := strings.TrimSpace(account.Name)
	if name == "" {
		name = "(unnamed)"
	}

	title := fmt.Sprintf("账号告警: 调度已停止 %s (#%d)", name, account.ID)

	sb := strings.Builder{}
	sb.WriteString("### 【账号告警】Anthropic API-key 账号连通性异常，已停止调度\n\n")
	sb.WriteString("**账号**：`")
	sb.WriteString(escapeInlineCode(name))
	sb.WriteString("` (#")
	sb.WriteString(fmt.Sprintf("%d", account.ID))
	sb.WriteString(")  \n")
	sb.WriteString("**平台**：`")
	sb.WriteString(escapeInlineCode(account.Platform))
	sb.WriteString("`  \n")
	sb.WriteString("**类型**：`")
	sb.WriteString(escapeInlineCode(account.Type))
	sb.WriteString("`  \n")
	sb.WriteString("**连续失败**：`")
	sb.WriteString(fmt.Sprintf("%d", threshold))
	sb.WriteString("`  \n")
	sb.WriteString("**耗时**：`")
	sb.WriteString(latency.String())
	sb.WriteString("`  \n")
	sb.WriteString("**时间**：`")
	sb.WriteString(escapeInlineCode(now.Format(time.RFC3339)))
	sb.WriteString("`  \n")

	reason = strings.TrimSpace(reason)
	if reason != "" {
		reason = strings.ReplaceAll(reason, "```", "'''")
		reason = truncateString(reason, 1500)
		sb.WriteString("\n\n**原因**\n")
		sb.WriteString("```text\n")
		sb.WriteString(reason)
		sb.WriteString("\n```\n")
	}

	go func(title, text string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.dingtalk.SendMarkdown(ctx, title, text); err != nil {
			slog.Warn("anthropic_apikey_monitor_dingtalk_send_failed", "account_id", account.ID, "error", err)
		}
	}(title, sb.String())
}

func (s *AnthropicAPIKeyMonitorService) sendRecoveryAlert(account Account, threshold int, latency time.Duration, now time.Time) {
	if s == nil || s.dingtalk == nil || !s.dingtalk.Enabled() {
		return
	}

	name := strings.TrimSpace(account.Name)
	if name == "" {
		name = "(unnamed)"
	}

	title := fmt.Sprintf("账号恢复: 调度已启用 %s (#%d)", name, account.ID)

	sb := strings.Builder{}
	sb.WriteString("### 【账号恢复】Anthropic API-key 账号连通性恢复，已启用调度\n\n")
	sb.WriteString("**账号**：`")
	sb.WriteString(escapeInlineCode(name))
	sb.WriteString("` (#")
	sb.WriteString(fmt.Sprintf("%d", account.ID))
	sb.WriteString(")  \n")
	sb.WriteString("**平台**：`")
	sb.WriteString(escapeInlineCode(account.Platform))
	sb.WriteString("`  \n")
	sb.WriteString("**类型**：`")
	sb.WriteString(escapeInlineCode(account.Type))
	sb.WriteString("`  \n")
	sb.WriteString("**连续成功**：`")
	sb.WriteString(fmt.Sprintf("%d", threshold))
	sb.WriteString("`  \n")
	sb.WriteString("**耗时**：`")
	sb.WriteString(latency.String())
	sb.WriteString("`  \n")
	sb.WriteString("**时间**：`")
	sb.WriteString(escapeInlineCode(now.Format(time.RFC3339)))
	sb.WriteString("`  \n")

	go func(title, text string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.dingtalk.SendMarkdown(ctx, title, text); err != nil {
			slog.Warn("anthropic_apikey_monitor_dingtalk_send_failed", "account_id", account.ID, "error", err)
		}
	}(title, sb.String())
}

func (s *AnthropicAPIKeyMonitorService) acquireOrRefreshLeaderLock() bool {
	if s == nil || !s.distributedLockOn {
		return true
	}
	if s.redisClient == nil {
		s.warnNoRedisOnce.Do(func() {
			slog.Warn("anthropic_apikey_monitor_redis_not_configured_running_without_lock")
		})
		return true
	}

	ctx, cancel := context.WithTimeout(s.stopCtx, 2*time.Second)
	defer cancel()

	ttl := s.baseLeaderLockTTL()

	ok, err := s.redisClient.SetNX(ctx, anthropicAPIKeyMonitorLeaderLockKey, s.instanceID, ttl).Result()
	if err != nil {
		// Fail-closed to avoid duplicated toggles when Redis is flaky.
		slog.Warn("anthropic_apikey_monitor_leader_lock_setnx_failed", "error", err)
		return false
	}
	if ok {
		return true
	}

	// If another instance holds the lock, skip. If we already hold it, refresh TTL.
	owner, err := s.redisClient.Get(ctx, anthropicAPIKeyMonitorLeaderLockKey).Result()
	if err != nil || strings.TrimSpace(owner) == "" {
		return false
	}
	if owner != s.instanceID {
		return false
	}
	_ = s.redisClient.Expire(ctx, anthropicAPIKeyMonitorLeaderLockKey, ttl).Err()
	return true
}

func (s *AnthropicAPIKeyMonitorService) extendLeaderLockTTL(targetCount int) {
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

	owner, err := s.redisClient.Get(ctx, anthropicAPIKeyMonitorLeaderLockKey).Result()
	if err != nil || owner != s.instanceID {
		return
	}
	_ = s.redisClient.Expire(ctx, anthropicAPIKeyMonitorLeaderLockKey, ttl).Err()
}

func (s *AnthropicAPIKeyMonitorService) baseLeaderLockTTL() time.Duration {
	// Keep a stable leader (counters are in-memory) but avoid overly long failover windows.
	interval := s.effectiveInterval()
	ttl := 12 * interval // e.g. 10s interval -> 120s ttl
	ttl = maxDuration(ttl, 2*time.Minute)
	ttl = maxDuration(ttl, 3*s.effectiveRequestTimeout())
	return ttl
}

func (s *AnthropicAPIKeyMonitorService) releaseLeaderLockBestEffort() {
	if s == nil || !s.distributedLockOn || s.redisClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Best-effort: delete lock only if still owned by this instance.
	val, err := s.redisClient.Get(ctx, anthropicAPIKeyMonitorLeaderLockKey).Result()
	if err != nil {
		return
	}
	if val != s.instanceID {
		return
	}
	_, _ = s.redisClient.Del(ctx, anthropicAPIKeyMonitorLeaderLockKey).Result()
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
