package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type scheduledTestPlanRepoRecorder struct {
	createdPlans []*ScheduledTestPlan
	createErr    error
}

func (s *scheduledTestPlanRepoRecorder) Create(_ context.Context, plan *ScheduledTestPlan) (*ScheduledTestPlan, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	cp := *plan
	s.createdPlans = append(s.createdPlans, &cp)
	return &cp, nil
}

func (s *scheduledTestPlanRepoRecorder) GetByID(_ context.Context, _ int64) (*ScheduledTestPlan, error) {
	panic("unexpected GetByID call")
}

func (s *scheduledTestPlanRepoRecorder) ListByAccountID(_ context.Context, _ int64) ([]*ScheduledTestPlan, error) {
	panic("unexpected ListByAccountID call")
}

func (s *scheduledTestPlanRepoRecorder) ListDue(_ context.Context, _ time.Time) ([]*ScheduledTestPlan, error) {
	panic("unexpected ListDue call")
}

func (s *scheduledTestPlanRepoRecorder) Update(_ context.Context, _ *ScheduledTestPlan) (*ScheduledTestPlan, error) {
	panic("unexpected Update call")
}

func (s *scheduledTestPlanRepoRecorder) Delete(_ context.Context, _ int64) error {
	panic("unexpected Delete call")
}

func (s *scheduledTestPlanRepoRecorder) UpdateAfterRun(_ context.Context, _ int64, _ time.Time, _ time.Time) error {
	panic("unexpected UpdateAfterRun call")
}

func TestBuildDefaultScheduledTestPlan_TargetPlatforms(t *testing.T) {
	now := time.Date(2026, 4, 2, 12, 34, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		platform string
		modelID  string
	}{
		{
			name:     "anthropic default model",
			platform: PlatformAnthropic,
			modelID:  defaultScheduledTestModelAnthropic,
		},
		{
			name:     "openai default model",
			platform: PlatformOpenAI,
			modelID:  defaultScheduledTestModelOpenAI,
		},
		{
			name:     "gemini default model",
			platform: PlatformGemini,
			modelID:  defaultScheduledTestModelGemini,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := buildDefaultScheduledTestPlan(101, tc.platform, now)
			require.NoError(t, err)
			require.NotNil(t, plan)
			require.Equal(t, int64(101), plan.AccountID)
			require.Equal(t, tc.modelID, plan.ModelID)
			require.Equal(t, defaultScheduledTestCronExpression, plan.CronExpression)
			require.True(t, plan.Enabled)
			require.True(t, plan.AutoRecover)
			require.Equal(t, defaultScheduledTestMaxResults, plan.MaxResults)
			require.NotNil(t, plan.NextRunAt)
			require.Equal(t, time.Date(2026, 4, 2, 12, 35, 0, 0, time.UTC), *plan.NextRunAt)
		})
	}
}

func TestBuildDefaultScheduledTestPlan_UnsupportedPlatform(t *testing.T) {
	plan, err := buildDefaultScheduledTestPlan(101, PlatformSora, time.Now())
	require.NoError(t, err)
	require.Nil(t, plan)
}

func TestCreateDefaultScheduledTestPlan_CreatesExpectedPlan(t *testing.T) {
	repo := &scheduledTestPlanRepoRecorder{}
	svc := &adminServiceImpl{
		scheduledTestPlanRepo: repo,
	}

	err := svc.createDefaultScheduledTestPlan(context.Background(), &Account{
		ID:       1001,
		Platform: PlatformOpenAI,
	})
	require.NoError(t, err)
	require.Len(t, repo.createdPlans, 1)

	plan := repo.createdPlans[0]
	require.Equal(t, int64(1001), plan.AccountID)
	require.Equal(t, defaultScheduledTestModelOpenAI, plan.ModelID)
	require.Equal(t, defaultScheduledTestCronExpression, plan.CronExpression)
	require.True(t, plan.Enabled)
	require.True(t, plan.AutoRecover)
}

func TestCreateDefaultScheduledTestPlan_PropagatesCreateError(t *testing.T) {
	repo := &scheduledTestPlanRepoRecorder{
		createErr: errors.New("insert failed"),
	}
	svc := &adminServiceImpl{
		scheduledTestPlanRepo: repo,
	}

	err := svc.createDefaultScheduledTestPlan(context.Background(), &Account{
		ID:       1001,
		Platform: PlatformGemini,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "insert failed")
}
