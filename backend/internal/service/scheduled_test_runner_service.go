package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/robfig/cron/v3"
)

const scheduledTestDefaultMaxWorkers = 10

const (
	scheduledTestExtraAutoDisabledKey    = "scheduled_test_auto_disabled"
	scheduledTestExtraDisabledAtKey      = "scheduled_test_disabled_at"
	scheduledTestExtraDisabledReasonKey  = "scheduled_test_disabled_reason"
	scheduledTestExtraRecoveredAtKey     = "scheduled_test_recovered_at"
	scheduledTestExtraRecoveredReasonKey = "scheduled_test_recovered_reason"
)

// ScheduledTestRunnerService periodically scans due test plans and executes them.
type ScheduledTestRunnerService struct {
	planRepo       ScheduledTestPlanRepository
	scheduledSvc   *ScheduledTestService
	accountTestSvc *AccountTestService
	rateLimitSvc   *RateLimitService
	cfg            *config.Config

	cron      *cron.Cron
	startOnce sync.Once
	stopOnce  sync.Once
}

// NewScheduledTestRunnerService creates a new runner.
func NewScheduledTestRunnerService(
	planRepo ScheduledTestPlanRepository,
	scheduledSvc *ScheduledTestService,
	accountTestSvc *AccountTestService,
	rateLimitSvc *RateLimitService,
	cfg *config.Config,
) *ScheduledTestRunnerService {
	return &ScheduledTestRunnerService{
		planRepo:       planRepo,
		scheduledSvc:   scheduledSvc,
		accountTestSvc: accountTestSvc,
		rateLimitSvc:   rateLimitSvc,
		cfg:            cfg,
	}
}

// Start begins the cron ticker (every minute).
func (s *ScheduledTestRunnerService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		loc := time.Local
		if s.cfg != nil {
			if parsed, err := time.LoadLocation(s.cfg.Timezone); err == nil && parsed != nil {
				loc = parsed
			}
		}

		c := cron.New(cron.WithParser(scheduledTestCronParser), cron.WithLocation(loc))
		_, err := c.AddFunc("* * * * *", func() { s.runScheduled() })
		if err != nil {
			logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] not started (invalid schedule): %v", err)
			return
		}
		s.cron = c
		s.cron.Start()
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] started (tick=every minute)")
	})
}

// Stop gracefully shuts down the cron scheduler.
func (s *ScheduledTestRunnerService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.cron != nil {
			ctx := s.cron.Stop()
			select {
			case <-ctx.Done():
			case <-time.After(3 * time.Second):
				logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] cron stop timed out")
			}
		}
	})
}

func (s *ScheduledTestRunnerService) runScheduled() {
	// Delay 10s so execution lands at ~:10 of each minute instead of :00.
	time.Sleep(10 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	now := time.Now()
	plans, err := s.planRepo.ListDue(ctx, now)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] ListDue error: %v", err)
		return
	}
	if len(plans) == 0 {
		return
	}

	logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] found %d due plans", len(plans))

	sem := make(chan struct{}, scheduledTestDefaultMaxWorkers)
	var wg sync.WaitGroup
	var resultMu sync.Mutex
	failedAccountIDs := make(map[int64]struct{})
	failedReasons := make(map[int64]string)
	succeededAccountIDs := make(map[int64]struct{})

	for _, plan := range plans {
		sem <- struct{}{}
		wg.Add(1)
		go func(p *ScheduledTestPlan) {
			defer wg.Done()
			defer func() { <-sem }()
			abnormal, reason := s.runOnePlan(ctx, p)
			resultMu.Lock()
			defer resultMu.Unlock()
			if abnormal {
				failedAccountIDs[p.AccountID] = struct{}{}
				if reason != "" {
					failedReasons[p.AccountID] = reason
				}
				return
			}
			succeededAccountIDs[p.AccountID] = struct{}{}
		}(plan)
	}

	wg.Wait()
	s.applyScheduledTestSchedulableState(ctx, failedAccountIDs, failedReasons, succeededAccountIDs, time.Now())
}

func (s *ScheduledTestRunnerService) runOnePlan(ctx context.Context, plan *ScheduledTestPlan) (bool, string) {
	result, err := s.accountTestSvc.RunTestBackground(ctx, plan.AccountID, plan.ModelID)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d RunTestBackground error: %v", plan.ID, err)
		return true, strings.TrimSpace(err.Error())
	}
	if result == nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d RunTestBackground returned nil result", plan.ID)
		return true, "scheduled test returned nil result"
	}

	if err := s.scheduledSvc.SaveResult(ctx, plan.ID, plan.MaxResults, result); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d SaveResult error: %v", plan.ID, err)
	}
	isAbnormal := !strings.EqualFold(result.Status, "success")

	// Auto-recover account if test succeeded and auto_recover is enabled.
	if !isAbnormal && plan.AutoRecover {
		s.tryRecoverAccount(ctx, plan.AccountID, plan.ID)
	}

	nextRun, err := computeNextRun(plan.CronExpression, time.Now())
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d computeNextRun error: %v", plan.ID, err)
		if isAbnormal {
			return true, scheduledTestFailureReason(result)
		}
		return false, ""
	}

	if err := s.planRepo.UpdateAfterRun(ctx, plan.ID, time.Now(), nextRun); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d UpdateAfterRun error: %v", plan.ID, err)
	}
	if isAbnormal {
		return true, scheduledTestFailureReason(result)
	}
	return false, ""
}

// tryRecoverAccount attempts to recover an account from recoverable runtime state.
func (s *ScheduledTestRunnerService) tryRecoverAccount(ctx context.Context, accountID int64, planID int64) {
	if s.rateLimitSvc == nil {
		return
	}

	recovery, err := s.rateLimitSvc.RecoverAccountAfterSuccessfulTest(ctx, accountID)
	if err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d auto-recover failed: %v", planID, err)
		return
	}
	if recovery == nil {
		return
	}

	if recovery.ClearedError {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d auto-recover: account=%d recovered from error status", planID, accountID)
	}
	if recovery.ClearedRateLimit {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] plan=%d auto-recover: account=%d cleared rate-limit/runtime state", planID, accountID)
	}
}

func (s *ScheduledTestRunnerService) applyScheduledTestSchedulableState(
	ctx context.Context,
	failedAccountIDs map[int64]struct{},
	failedReasons map[int64]string,
	succeededAccountIDs map[int64]struct{},
	now time.Time,
) {
	if s == nil || s.accountTestSvc == nil || s.accountTestSvc.accountRepo == nil {
		return
	}
	for accountID := range failedAccountIDs {
		s.tryAutoDisableSchedulableAccount(ctx, accountID, failedReasons[accountID], now)
	}
	for accountID := range succeededAccountIDs {
		if _, failed := failedAccountIDs[accountID]; failed {
			continue
		}
		s.tryAutoResumeSchedulableAccount(ctx, accountID, now)
	}
}

func (s *ScheduledTestRunnerService) tryAutoDisableSchedulableAccount(ctx context.Context, accountID int64, reason string, now time.Time) {
	account, err := s.accountTestSvc.accountRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto-disable get account failed: account=%d err=%v", accountID, err)
		return
	}
	if !account.Schedulable {
		return
	}

	if err := s.accountTestSvc.accountRepo.SetSchedulable(ctx, accountID, false); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto-disable schedulable failed: account=%d err=%v", accountID, err)
		return
	}
	updates := map[string]any{
		scheduledTestExtraAutoDisabledKey:    true,
		scheduledTestExtraDisabledAtKey:      now.Format(time.RFC3339),
		scheduledTestExtraDisabledReasonKey:  scheduledTestDisableReason(reason),
		scheduledTestExtraRecoveredAtKey:     nil,
		scheduledTestExtraRecoveredReasonKey: nil,
	}
	if err := s.accountTestSvc.accountRepo.UpdateExtra(ctx, accountID, updates); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto-disable update extra failed: account=%d err=%v", accountID, err)
	}
	logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto-disabled account scheduling: account=%d", accountID)
}

func (s *ScheduledTestRunnerService) tryAutoResumeSchedulableAccount(ctx context.Context, accountID int64, now time.Time) {
	account, err := s.accountTestSvc.accountRepo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto-resume get account failed: account=%d err=%v", accountID, err)
		return
	}
	if account.Schedulable || !account.getExtraBool(scheduledTestExtraAutoDisabledKey) {
		return
	}

	if err := s.accountTestSvc.accountRepo.SetSchedulable(ctx, accountID, true); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto-resume schedulable failed: account=%d err=%v", accountID, err)
		return
	}
	updates := map[string]any{
		scheduledTestExtraAutoDisabledKey:    false,
		scheduledTestExtraRecoveredAtKey:     now.Format(time.RFC3339),
		scheduledTestExtraRecoveredReasonKey: "scheduled test recovered",
	}
	if err := s.accountTestSvc.accountRepo.UpdateExtra(ctx, accountID, updates); err != nil {
		logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto-resume update extra failed: account=%d err=%v", accountID, err)
	}
	logger.LegacyPrintf("service.scheduled_test_runner", "[ScheduledTestRunner] auto-resumed account scheduling: account=%d", accountID)
}

func scheduledTestFailureReason(result *ScheduledTestResult) string {
	if result == nil {
		return "scheduled test failed"
	}
	if msg := strings.TrimSpace(result.ErrorMessage); msg != "" {
		return msg
	}
	if status := strings.TrimSpace(result.Status); !strings.EqualFold(status, "success") && status != "" {
		return "scheduled test status=" + status
	}
	return "scheduled test failed"
}

func scheduledTestDisableReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "scheduled test failed"
	}
	return truncateString(reason, 1500)
}
