//go:build unit

package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type scheduledTestAccountRepoStub struct {
	accountRepoStub
	account            *Account
	schedulableUpdates []bool
	extraUpdates       []map[string]any
}

func (s *scheduledTestAccountRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	return s.account, nil
}

func (s *scheduledTestAccountRepoStub) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	s.schedulableUpdates = append(s.schedulableUpdates, schedulable)
	if s.account != nil {
		s.account.Schedulable = schedulable
	}
	return nil
}

func (s *scheduledTestAccountRepoStub) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	copied := make(map[string]any, len(updates))
	for key, value := range updates {
		copied[key] = value
	}
	s.extraUpdates = append(s.extraUpdates, copied)
	if s.account != nil {
		if s.account.Extra == nil {
			s.account.Extra = map[string]any{}
		}
		for key, value := range updates {
			s.account.Extra[key] = value
		}
	}
	return nil
}

func TestScheduledTestRunnerAutoDisablesSchedulableAccountOnFailure(t *testing.T) {
	repo := &scheduledTestAccountRepoStub{
		account: &Account{ID: 42, Status: StatusActive, Schedulable: true},
	}
	svc := &ScheduledTestRunnerService{
		accountTestSvc: &AccountTestService{accountRepo: repo},
	}

	now := time.Date(2026, 5, 14, 8, 0, 0, 0, time.UTC)
	svc.tryAutoDisableSchedulableAccount(context.Background(), 42, " upstream timeout ", now)

	require.Equal(t, []bool{false}, repo.schedulableUpdates)
	require.Len(t, repo.extraUpdates, 1)
	require.Equal(t, true, repo.extraUpdates[0][scheduledTestExtraAutoDisabledKey])
	require.Equal(t, now.Format(time.RFC3339), repo.extraUpdates[0][scheduledTestExtraDisabledAtKey])
	require.Equal(t, "upstream timeout", repo.extraUpdates[0][scheduledTestExtraDisabledReasonKey])
}

func TestScheduledTestRunnerAutoResumesOnlyAutoDisabledAccount(t *testing.T) {
	repo := &scheduledTestAccountRepoStub{
		account: &Account{
			ID:          42,
			Status:      StatusActive,
			Schedulable: false,
			Extra: map[string]any{
				scheduledTestExtraAutoDisabledKey: true,
			},
		},
	}
	svc := &ScheduledTestRunnerService{
		accountTestSvc: &AccountTestService{accountRepo: repo},
	}

	now := time.Date(2026, 5, 14, 8, 5, 0, 0, time.UTC)
	svc.tryAutoResumeSchedulableAccount(context.Background(), 42, now)

	require.Equal(t, []bool{true}, repo.schedulableUpdates)
	require.Len(t, repo.extraUpdates, 1)
	require.Equal(t, false, repo.extraUpdates[0][scheduledTestExtraAutoDisabledKey])
	require.Equal(t, now.Format(time.RFC3339), repo.extraUpdates[0][scheduledTestExtraRecoveredAtKey])
}

func TestScheduledTestRunnerDoesNotResumeManuallyDisabledAccount(t *testing.T) {
	repo := &scheduledTestAccountRepoStub{
		account: &Account{ID: 42, Status: StatusActive, Schedulable: false},
	}
	svc := &ScheduledTestRunnerService{
		accountTestSvc: &AccountTestService{accountRepo: repo},
	}

	svc.tryAutoResumeSchedulableAccount(context.Background(), 42, time.Now())

	require.Empty(t, repo.schedulableUpdates)
	require.Empty(t, repo.extraUpdates)
}

func TestScheduledTestRunnerFailureReasonDefaultAndTruncate(t *testing.T) {
	require.Equal(t, "scheduled test failed", scheduledTestDisableReason("   "))

	long := strings.Repeat("x", 2000)
	got := scheduledTestDisableReason(long)
	require.Len(t, got, 1500)
}

func TestScheduledTestRunnerFailureReason(t *testing.T) {
	require.Equal(t, "scheduled test failed", scheduledTestFailureReason(nil))
	require.Equal(t, "upstream timeout", scheduledTestFailureReason(&ScheduledTestResult{
		Status:       "failed",
		ErrorMessage: " upstream timeout ",
	}))
	require.Equal(t, "scheduled test status=partial_success", scheduledTestFailureReason(&ScheduledTestResult{
		Status: "partial_success",
	}))
}
