//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildAccountErrorDingtalkMessage(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 31, 12, 34, 56, 0, time.UTC)
	acc := &Account{
		ID:           123,
		Name:         "acc-openai-1",
		Platform:     PlatformOpenAI,
		Type:         AccountTypeOAuth,
		Status:       StatusError,
		ErrorMessage: "Authentication failed (401): invalid or expired credentials",
	}

	title, text := buildAccountErrorDingtalkMessage(acc, "ratelimit", acc.ErrorMessage, map[string]string{
		"category":    "auth_error",
		"status_code": "401",
	}, now)

	require.Contains(t, title, "账号告警:")
	require.Contains(t, title, "acc-openai-1")
	require.Contains(t, title, "#123")

	require.Contains(t, text, "【账号告警】账号状态异常")
	require.Contains(t, text, "**账号**：`acc-openai-1`")
	require.Contains(t, text, "(#123)")
	require.Contains(t, text, "**状态**：`error`")
	require.Contains(t, text, "**平台**：`openai`")
	require.Contains(t, text, "**类型**：`oauth`")
	require.Contains(t, text, "**时间**：`2026-01-31T12:34:56Z`")
	require.Contains(t, text, "Authentication failed (401)")
	require.Contains(t, text, "**原因**")
	require.Contains(t, text, "```text")
	require.NotContains(t, text, "ratelimit")
	require.NotContains(t, text, "category")
}

func TestAccountAlertServiceAllowCooldown(t *testing.T) {
	t.Parallel()

	svc := &AccountAlertService{
		throttle: map[int64]accountAlertThrottleState{},
	}
	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	require.False(t, svc.allow(0, now))
	require.True(t, svc.allow(1, now))
	require.False(t, svc.allow(1, now.Add(1*time.Minute)))
	second := now.Add(accountAlertCooldown + time.Second)
	require.True(t, svc.allow(1, second))

	// progressive cooldown: 5m -> 10m
	require.False(t, svc.allow(1, second.Add(9*time.Minute)))
	require.True(t, svc.allow(1, second.Add(10*time.Minute+time.Second)))
}

func TestAccountAlertServiceAllowCooldownCapsAtMax(t *testing.T) {
	t.Parallel()

	svc := &AccountAlertService{
		throttle: map[int64]accountAlertThrottleState{},
	}
	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	// first alert starts with base cooldown
	require.True(t, svc.allow(42, now))
	st := svc.throttle[42]
	require.Equal(t, accountAlertCooldown, st.cooldown)

	// keep triggering at each allowed moment until reaching max cooldown
	current := now
	for i := 0; i < 8; i++ {
		current = current.Add(st.cooldown + time.Second)
		require.True(t, svc.allow(42, current))
		st = svc.throttle[42]
	}
	require.Equal(t, accountAlertMaxCooldown, st.cooldown)

	// once capped, still enforces max cooldown and no longer increases.
	require.False(t, svc.allow(42, current.Add(accountAlertMaxCooldown-time.Second)))
	require.True(t, svc.allow(42, current.Add(accountAlertMaxCooldown+time.Second)))
	require.Equal(t, accountAlertMaxCooldown, svc.throttle[42].cooldown)
}

func TestAccountAlertServiceAllowCooldownResetsAfterQuietPeriod(t *testing.T) {
	t.Parallel()

	svc := &AccountAlertService{
		throttle: map[int64]accountAlertThrottleState{},
	}
	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	require.True(t, svc.allow(7, now))
	second := now.Add(accountAlertCooldown + time.Second)
	require.True(t, svc.allow(7, second))
	require.Equal(t, 2*accountAlertCooldown, svc.throttle[7].cooldown)

	afterQuiet := second.Add(accountAlertBackoffResetAfter + time.Second)
	require.True(t, svc.allow(7, afterQuiet))
	require.Equal(t, accountAlertCooldown, svc.throttle[7].cooldown)
}
