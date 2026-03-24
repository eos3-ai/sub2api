package service

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestComputeAvailableSchedulableAPIKeyCount(t *testing.T) {
	now := time.Now()
	future := now.Add(10 * time.Minute)

	accounts := []Account{
		{ID: 1, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},                         // available
		{ID: 2, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: false},                        // not schedulable
		{ID: 3, Type: AccountTypeAPIKey, Status: StatusError, Schedulable: true},                          // status error
		{ID: 4, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true},                          // not apikey
		{ID: 5, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, OverloadUntil: &future}, // overloaded
		{ID: 6, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},                         // available
	}

	available, deduction := computeAvailableSchedulableAPIKeyCount(accounts, nil)
	require.Equal(t, 2, available)
	require.Equal(t, 0, deduction)
}

func TestComputeAvailableSchedulableAPIKeyCount_DeductScheduledAbnormal(t *testing.T) {
	accounts := []Account{
		{ID: 10, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},
		{ID: 11, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},
		{ID: 12, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true},
	}
	abnormal := map[int64]struct{}{
		11: {},
		99: {}, // nonexistent id should be ignored
	}

	available, deduction := computeAvailableSchedulableAPIKeyCount(accounts, abnormal)
	require.Equal(t, 2, available)
	require.Equal(t, 1, deduction)
}

func TestScheduledTestDisableReason_DefaultAndTruncate(t *testing.T) {
	require.Equal(t, "scheduled test failed", scheduledTestDisableReason("   "))

	long := strings.Repeat("x", 2000)
	got := scheduledTestDisableReason(long)
	require.Len(t, got, 1500)
}

func TestScheduledTestFailureReason(t *testing.T) {
	require.Equal(t, "scheduled test failed", scheduledTestFailureReason(nil))

	require.Equal(t, "upstream timeout", scheduledTestFailureReason(&ScheduledTestResult{
		Status:       "failed",
		ErrorMessage: " upstream timeout ",
	}))

	require.Equal(t, "scheduled test status=partial_success", scheduledTestFailureReason(&ScheduledTestResult{
		Status: "partial_success",
	}))
}

func TestShouldSendLowAvailableAlertProgressiveBackoff(t *testing.T) {
	svc := &ScheduledTestRunnerService{
		lowAvailableAlertThrottle: map[string]scheduledTestAlertThrottleState{},
	}
	now := time.Date(2026, 3, 24, 4, 0, 0, 0, time.UTC)

	require.True(t, svc.shouldSendLowAvailableAlert(PlatformAnthropic, now))
	require.False(t, svc.shouldSendLowAvailableAlert(PlatformAnthropic, now.Add(1*time.Minute)))

	second := now.Add(scheduledTestLowAvailableAlertBaseCooldown + time.Second)
	require.True(t, svc.shouldSendLowAvailableAlert(PlatformAnthropic, second))

	// next required window should be 10 minutes.
	require.False(t, svc.shouldSendLowAvailableAlert(PlatformAnthropic, second.Add(9*time.Minute)))
	require.True(t, svc.shouldSendLowAvailableAlert(PlatformAnthropic, second.Add(10*time.Minute+time.Second)))

	// different platform should not be throttled by anthropic state.
	require.True(t, svc.shouldSendLowAvailableAlert(PlatformOpenAI, second.Add(10*time.Minute+time.Second)))
}

func TestShouldSendLowAvailableAlertCapsAtMax(t *testing.T) {
	svc := &ScheduledTestRunnerService{
		lowAvailableAlertThrottle: map[string]scheduledTestAlertThrottleState{},
	}
	now := time.Date(2026, 3, 24, 4, 0, 0, 0, time.UTC)

	require.True(t, svc.shouldSendLowAvailableAlert(PlatformGemini, now))
	st := svc.lowAvailableAlertThrottle[PlatformGemini]
	require.Equal(t, scheduledTestLowAvailableAlertBaseCooldown, st.cooldown)

	current := now
	for i := 0; i < 8; i++ {
		current = current.Add(st.cooldown + time.Second)
		require.True(t, svc.shouldSendLowAvailableAlert(PlatformGemini, current))
		st = svc.lowAvailableAlertThrottle[PlatformGemini]
	}
	require.Equal(t, scheduledTestLowAvailableAlertMaxCooldown, st.cooldown)

	require.False(t, svc.shouldSendLowAvailableAlert(PlatformGemini, current.Add(scheduledTestLowAvailableAlertMaxCooldown-time.Second)))
	require.True(t, svc.shouldSendLowAvailableAlert(PlatformGemini, current.Add(scheduledTestLowAvailableAlertMaxCooldown+time.Second)))
	require.Equal(t, scheduledTestLowAvailableAlertMaxCooldown, svc.lowAvailableAlertThrottle[PlatformGemini].cooldown)
}

func TestShouldSendLowAvailableAlertResetsAfterQuietPeriod(t *testing.T) {
	svc := &ScheduledTestRunnerService{
		lowAvailableAlertThrottle: map[string]scheduledTestAlertThrottleState{},
	}
	now := time.Date(2026, 3, 24, 4, 0, 0, 0, time.UTC)

	require.True(t, svc.shouldSendLowAvailableAlert(PlatformOpenAI, now))
	second := now.Add(scheduledTestLowAvailableAlertBaseCooldown + time.Second)
	require.True(t, svc.shouldSendLowAvailableAlert(PlatformOpenAI, second))
	require.Equal(t, 2*scheduledTestLowAvailableAlertBaseCooldown, svc.lowAvailableAlertThrottle[PlatformOpenAI].cooldown)

	afterQuiet := second.Add(scheduledTestLowAvailableAlertResetAfter + time.Second)
	require.True(t, svc.shouldSendLowAvailableAlert(PlatformOpenAI, afterQuiet))
	require.Equal(t, scheduledTestLowAvailableAlertBaseCooldown, svc.lowAvailableAlertThrottle[PlatformOpenAI].cooldown)
}

func TestShouldSendLowAvailableAlertRejectsUnknownPlatform(t *testing.T) {
	svc := &ScheduledTestRunnerService{
		lowAvailableAlertThrottle: map[string]scheduledTestAlertThrottleState{},
	}
	now := time.Date(2026, 3, 24, 4, 0, 0, 0, time.UTC)
	require.False(t, svc.shouldSendLowAvailableAlert("unknown", now))
}
