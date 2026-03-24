package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

const (
	// accountAlertCooldown is a best-effort dedup window to avoid spamming DingTalk
	// when multiple requests concurrently mark the same account as error.
	accountAlertCooldown = 5 * time.Minute

	// accountAlertMaxCooldown is the upper bound for progressive cooldown.
	accountAlertMaxCooldown = 60 * time.Minute

	// accountAlertBackoffResetAfter resets cooldown to base after a long quiet period.
	accountAlertBackoffResetAfter = 6 * time.Hour

	// accountAlertSendTimeout bounds the DingTalk webhook call.
	accountAlertSendTimeout = 5 * time.Second
)

// AccountAlertService sends notifications when an account becomes abnormal (e.g. status=error).
//
// It is intentionally best-effort:
// - never blocks the critical path (async send)
// - skips when DingTalk isn't enabled/configured
// - applies a progressive in-process cooldown per account ID when errors keep repeating
type AccountAlertService struct {
	dingtalk *DingtalkService

	mu       sync.Mutex
	throttle map[int64]accountAlertThrottleState
}

type accountAlertThrottleState struct {
	lastSent time.Time
	cooldown time.Duration
}

func NewAccountAlertService(cfg *config.Config) *AccountAlertService {
	return &AccountAlertService{
		dingtalk: NewDingtalkService(cfg),
		throttle: map[int64]accountAlertThrottleState{},
	}
}

func (s *AccountAlertService) NotifyAccountStatusError(account *Account, source string, reason string, fields map[string]string) {
	if s == nil || account == nil {
		return
	}
	if s.dingtalk == nil || !s.dingtalk.Enabled() {
		return
	}

	now := time.Now().UTC()
	if !s.allow(account.ID, now) {
		return
	}

	title, text := buildAccountErrorDingtalkMessage(account, source, reason, fields, now)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), accountAlertSendTimeout)
		defer cancel()

		if err := s.dingtalk.SendMarkdown(ctx, title, text); err != nil {
			slog.Warn("account_alert_dingtalk_send_failed", "account_id", account.ID, "error", err)
		}
	}()
}

func (s *AccountAlertService) allow(accountID int64, now time.Time) bool {
	if accountID <= 0 {
		return false
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.throttle == nil {
		s.throttle = map[int64]accountAlertThrottleState{}
	}

	st, ok := s.throttle[accountID]
	if !ok || st.lastSent.IsZero() || now.Sub(st.lastSent) >= accountAlertBackoffResetAfter {
		s.throttle[accountID] = accountAlertThrottleState{
			lastSent: now,
			cooldown: accountAlertCooldown,
		}
		return true
	}

	if st.cooldown <= 0 {
		st.cooldown = accountAlertCooldown
	}

	if now.Sub(st.lastSent) < st.cooldown {
		return false
	}

	nextCooldown := st.cooldown * 2
	if nextCooldown > accountAlertMaxCooldown {
		nextCooldown = accountAlertMaxCooldown
	}
	if nextCooldown < accountAlertCooldown {
		nextCooldown = accountAlertCooldown
	}

	s.throttle[accountID] = accountAlertThrottleState{
		lastSent: now,
		cooldown: nextCooldown,
	}
	return true
}

func buildAccountErrorDingtalkMessage(account *Account, source string, reason string, fields map[string]string, now time.Time) (title string, text string) {
	if account == nil {
		return "Account Alert", "account is nil"
	}

	// source and fields are intentionally ignored in the output:
	// user requested removing "来源" and "维度" from the alert message.
	_ = source
	_ = fields

	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = strings.TrimSpace(account.ErrorMessage)
	}

	// Avoid breaking markdown code block formatting.
	reason = strings.ReplaceAll(reason, "```", "'''")
	reason = truncateAccountAlertReason(reason, 1500)

	name := strings.TrimSpace(account.Name)
	if name == "" {
		name = "(unnamed)"
	}

	title = fmt.Sprintf("账号告警: %s (#%d)", name, account.ID)

	sb := strings.Builder{}
	sb.WriteString("### 【账号告警】账号状态异常\n\n")

	// Use Markdown hard line-breaks (`two spaces + \n`) to keep the layout vertical
	// across DingTalk clients (some treat single newlines as spaces).
	sb.WriteString("**账号**：`")
	sb.WriteString(escapeInlineCode(name))
	sb.WriteString("` (#")
	sb.WriteString(fmt.Sprintf("%d", account.ID))
	sb.WriteString(")  \n")

	sb.WriteString("**状态**：`")
	sb.WriteString(escapeInlineCode(account.Status))
	sb.WriteString("`  \n")

	sb.WriteString("**平台**：`")
	sb.WriteString(escapeInlineCode(account.Platform))
	sb.WriteString("`  \n")

	sb.WriteString("**类型**：`")
	sb.WriteString(escapeInlineCode(account.Type))
	sb.WriteString("`  \n")

	sb.WriteString("**时间**：`")
	sb.WriteString(escapeInlineCode(now.Format(time.RFC3339)))
	sb.WriteString("`  \n")

	if reason != "" {
		sb.WriteString("\n\n**原因**\n")
		sb.WriteString("```text\n")
		sb.WriteString(reason)
		sb.WriteString("\n```\n")
	}

	return title, sb.String()
}

func escapeInlineCode(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"\r", " ",
		"\n", " ",
		"`", "'",
	)
	return replacer.Replace(s)
}

func truncateAccountAlertReason(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max]
}
