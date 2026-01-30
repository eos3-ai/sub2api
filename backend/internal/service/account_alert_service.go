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

	// accountAlertSendTimeout bounds the DingTalk webhook call.
	accountAlertSendTimeout = 5 * time.Second
)

// AccountAlertService sends notifications when an account becomes abnormal (e.g. status=error).
//
// It is intentionally best-effort:
// - never blocks the critical path (async send)
// - skips when DingTalk isn't enabled/configured
// - applies a small in-process cooldown per account ID to reduce duplicate alerts
type AccountAlertService struct {
	dingtalk *DingtalkService

	mu       sync.Mutex
	lastSent map[int64]time.Time
}

func NewAccountAlertService(cfg *config.Config) *AccountAlertService {
	return &AccountAlertService{
		dingtalk:  NewDingtalkService(cfg),
		lastSent: map[int64]time.Time{},
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

	if s.lastSent == nil {
		s.lastSent = map[int64]time.Time{}
	}

	if last, ok := s.lastSent[accountID]; ok && !last.IsZero() {
		if now.Sub(last) < accountAlertCooldown {
			return false
		}
	}
	s.lastSent[accountID] = now
	return true
}

func buildAccountErrorDingtalkMessage(account *Account, source string, reason string, fields map[string]string, now time.Time) (title string, text string) {
	if account == nil {
		return "Account Alert", "account is nil"
	}

	source = strings.TrimSpace(source)
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = strings.TrimSpace(account.ErrorMessage)
	}

	// Avoid breaking markdown code block formatting.
	reason = strings.ReplaceAll(reason, "```", "'''")
	reason = truncateString(reason, 1500)

	name := strings.TrimSpace(account.Name)
	if name == "" {
		name = "(unnamed)"
	}

	title = fmt.Sprintf("Account Alert: %s (#%d)", name, account.ID)

	sb := strings.Builder{}
	sb.WriteString("### 账号状态异常\n\n")
	sb.WriteString(fmt.Sprintf("- Time: %s\n", now.Format(time.RFC3339)))
	if source != "" {
		sb.WriteString(fmt.Sprintf("- Source: %s\n", escapeInlineMarkdown(source)))
	}
	sb.WriteString(fmt.Sprintf("- AccountID: %d\n", account.ID))
	sb.WriteString(fmt.Sprintf("- Name: %s\n", escapeInlineMarkdown(name)))
	if strings.TrimSpace(account.Platform) != "" {
		sb.WriteString(fmt.Sprintf("- Platform: %s\n", escapeInlineMarkdown(account.Platform)))
	}
	if strings.TrimSpace(account.Type) != "" {
		sb.WriteString(fmt.Sprintf("- Type: %s\n", escapeInlineMarkdown(account.Type)))
	}
	if strings.TrimSpace(account.Status) != "" {
		sb.WriteString(fmt.Sprintf("- Status: %s\n", escapeInlineMarkdown(account.Status)))
	}

	if len(fields) > 0 {
		for k, v := range fields {
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			if k == "" || v == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf("- %s: %s\n", escapeInlineMarkdown(k), escapeInlineMarkdown(v)))
		}
	}

	if reason != "" {
		sb.WriteString("\n**Reason**:\n")
		sb.WriteString("```text\n")
		sb.WriteString(reason)
		sb.WriteString("\n```\n")
	}

	return title, sb.String()
}

// escapeInlineMarkdown does minimal escaping for DingTalk markdown list items.
func escapeInlineMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"\r", " ",
		"\n", " ",
		"|", "\\|",
	)
	return replacer.Replace(s)
}
