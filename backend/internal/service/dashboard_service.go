package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

const (
	defaultDashboardStatsFreshTTL       = 15 * time.Second
	defaultDashboardStatsCacheTTL       = 30 * time.Second
	defaultDashboardStatsRefreshTimeout = 30 * time.Second
)

// ErrDashboardStatsCacheMiss 标记仪表盘缓存未命中。
var ErrDashboardStatsCacheMiss = errors.New("仪表盘缓存未命中")

// DashboardStatsCache 定义仪表盘统计缓存接口。
type DashboardStatsCache interface {
	GetDashboardStats(ctx context.Context) (string, error)
	SetDashboardStats(ctx context.Context, data string, ttl time.Duration) error
	DeleteDashboardStats(ctx context.Context) error
}

type dashboardStatsRangeFetcher interface {
	GetDashboardStatsWithRange(ctx context.Context, start, end time.Time) (*usagestats.DashboardStats, error)
}

type dashboardStatsCacheEntry struct {
	Stats     *usagestats.DashboardStats `json:"stats"`
	UpdatedAt int64                      `json:"updated_at"`
}

// DashboardService 提供管理员仪表盘统计服务。
type DashboardService struct {
	usageRepo      UsageLogRepository
	aggRepo        DashboardAggregationRepository
	cache          DashboardStatsCache
	cacheFreshTTL  time.Duration
	cacheTTL       time.Duration
	refreshTimeout time.Duration
	refreshing     int32
	aggEnabled     bool
	aggInterval    time.Duration
	aggLookback    time.Duration
	aggUsageDays   int
}

func NewDashboardService(usageRepo UsageLogRepository, aggRepo DashboardAggregationRepository, cache DashboardStatsCache, cfg *config.Config) *DashboardService {
	freshTTL := defaultDashboardStatsFreshTTL
	cacheTTL := defaultDashboardStatsCacheTTL
	refreshTimeout := defaultDashboardStatsRefreshTimeout
	aggEnabled := true
	aggInterval := time.Minute
	aggLookback := 2 * time.Minute
	aggUsageDays := 90
	if cfg != nil {
		if !cfg.Dashboard.Enabled {
			cache = nil
		}
		if cfg.Dashboard.StatsFreshTTLSeconds > 0 {
			freshTTL = time.Duration(cfg.Dashboard.StatsFreshTTLSeconds) * time.Second
		}
		if cfg.Dashboard.StatsTTLSeconds > 0 {
			cacheTTL = time.Duration(cfg.Dashboard.StatsTTLSeconds) * time.Second
		}
		if cfg.Dashboard.StatsRefreshTimeoutSeconds > 0 {
			refreshTimeout = time.Duration(cfg.Dashboard.StatsRefreshTimeoutSeconds) * time.Second
		}
		aggEnabled = cfg.DashboardAgg.Enabled
		if cfg.DashboardAgg.IntervalSeconds > 0 {
			aggInterval = time.Duration(cfg.DashboardAgg.IntervalSeconds) * time.Second
		}
		if cfg.DashboardAgg.LookbackSeconds > 0 {
			aggLookback = time.Duration(cfg.DashboardAgg.LookbackSeconds) * time.Second
		}
		if cfg.DashboardAgg.Retention.UsageLogsDays > 0 {
			aggUsageDays = cfg.DashboardAgg.Retention.UsageLogsDays
		}
	}
	if aggRepo == nil {
		aggEnabled = false
	}
	return &DashboardService{
		usageRepo:      usageRepo,
		aggRepo:        aggRepo,
		cache:          cache,
		cacheFreshTTL:  freshTTL,
		cacheTTL:       cacheTTL,
		refreshTimeout: refreshTimeout,
		aggEnabled:     aggEnabled,
		aggInterval:    aggInterval,
		aggLookback:    aggLookback,
		aggUsageDays:   aggUsageDays,
	}
}

func (s *DashboardService) GetDashboardStats(ctx context.Context) (*usagestats.DashboardStats, error) {
	if s.cache != nil {
		cached, fresh, err := s.getCachedDashboardStats(ctx)
		if err == nil && cached != nil {
			s.refreshAggregationStaleness(cached)
			if !fresh {
				s.refreshDashboardStatsAsync()
			}
			return cached, nil
		}
		if err != nil && !errors.Is(err, ErrDashboardStatsCacheMiss) {
			log.Printf("[Dashboard] 仪表盘缓存读取失败: %v", err)
		}
	}

	stats, err := s.refreshDashboardStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get dashboard stats: %w", err)
	}
	return stats, nil
}

func (s *DashboardService) GetUsageTrendWithFilters(ctx context.Context, startTime, endTime time.Time, granularity string, userID, apiKeyID int64) ([]usagestats.TrendDataPoint, error) {
	trend, err := s.usageRepo.GetUsageTrendWithFilters(ctx, startTime, endTime, granularity, userID, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("get usage trend with filters: %w", err)
	}
	return trend, nil
}

func (s *DashboardService) GetModelStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID int64) ([]usagestats.ModelStat, error) {
	stats, err := s.usageRepo.GetModelStatsWithFilters(ctx, startTime, endTime, userID, apiKeyID, 0)
	if err != nil {
		return nil, fmt.Errorf("get model stats with filters: %w", err)
	}
	return stats, nil
}

func (s *DashboardService) GetAPIKeyUsageTrend(ctx context.Context, startTime, endTime time.Time, granularity string, limit int) ([]usagestats.APIKeyUsageTrendPoint, error) {
	trend, err := s.usageRepo.GetAPIKeyUsageTrend(ctx, startTime, endTime, granularity, limit)
	if err != nil {
		return nil, fmt.Errorf("get api key usage trend: %w", err)
	}
	return trend, nil
}

func (s *DashboardService) GetUserUsageTrend(ctx context.Context, startTime, endTime time.Time, granularity string, limit int) ([]usagestats.UserUsageTrendPoint, error) {
	trend, err := s.usageRepo.GetUserUsageTrend(ctx, startTime, endTime, granularity, limit)
	if err != nil {
		return nil, fmt.Errorf("get user usage trend: %w", err)
	}
	return trend, nil
}

func (s *DashboardService) GetBatchUserUsageStats(ctx context.Context, userIDs []int64) (map[int64]*usagestats.BatchUserUsageStats, error) {
	stats, err := s.usageRepo.GetBatchUserUsageStats(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get batch user usage stats: %w", err)
	}
	return stats, nil
}

func (s *DashboardService) GetBatchAPIKeyUsageStats(ctx context.Context, apiKeyIDs []int64) (map[int64]*usagestats.BatchAPIKeyUsageStats, error) {
	stats, err := s.usageRepo.GetBatchAPIKeyUsageStats(ctx, apiKeyIDs)
	if err != nil {
		return nil, fmt.Errorf("get batch api key usage stats: %w", err)
	}
	return stats, nil
}
