package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

const publicGroupMonitorHistoryMaxWindow = 30 * 24 * time.Hour

// GetPublicGroupMonitorOverview aggregates scheduled-test results for public groups.
//
// Aggregation rules:
// - "same round" is merged by round_id (legacy data falls back to fixed time bucket)
// - within one round: any success => success, all failed => failed
// - representative sample:
//   - success round: choose the success record with shortest latency
//   - failure round: choose one deterministic failed record (latest by started_at/id)
func (r *groupRepository) GetPublicGroupMonitorOverview(
	ctx context.Context,
	groupIDs []int64,
	bucketSeconds int,
	sampleSize int,
	now time.Time,
	window time.Duration,
) (map[int64]*service.PublicGroupMonitorAggregate, error) {
	result := make(map[int64]*service.PublicGroupMonitorAggregate, len(groupIDs))
	normalizedGroupIDs := normalizeInt64IDs(groupIDs)
	if len(normalizedGroupIDs) == 0 {
		return result, nil
	}

	if bucketSeconds <= 0 {
		bucketSeconds = 15
	}
	if bucketSeconds > 300 {
		bucketSeconds = 300
	}
	if sampleSize <= 0 {
		sampleSize = 30
	}
	if sampleSize > 30 {
		sampleSize = 30
	}
	if window <= 0 {
		window = time.Hour
	}
	if window > 24*time.Hour {
		window = 24 * time.Hour
	}

	now = now.UTC()
	historyStart := now.Add(-publicGroupMonitorHistoryMaxWindow)
	windowStart := now.Add(-window)

	for _, groupID := range normalizedGroupIDs {
		result[groupID] = &service.PublicGroupMonitorAggregate{
			CurrentStatus: "unknown",
			Samples:       []*service.PublicGroupMonitorSample{},
		}
	}

	if err := r.loadPublicGroupMonitorStats(ctx, result, normalizedGroupIDs, bucketSeconds, historyStart, windowStart); err != nil {
		return nil, err
	}
	if err := r.loadPublicGroupMonitorSamples(ctx, result, normalizedGroupIDs, bucketSeconds, sampleSize, historyStart); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *groupRepository) loadPublicGroupMonitorStats(
	ctx context.Context,
	out map[int64]*service.PublicGroupMonitorAggregate,
	groupIDs []int64,
	bucketSeconds int,
	historyStart time.Time,
	windowStart time.Time,
) error {
	const q = `
WITH raw AS (
  SELECT
    ag.group_id,
    r.id AS result_id,
    r.started_at,
    COALESCE(
      NULLIF(TRIM(COALESCE(r.round_id, '')), ''),
      'legacy:' || ((EXTRACT(EPOCH FROM r.started_at)::BIGINT / $3::BIGINT) * $3::BIGINT)::text
    ) AS round_key,
    CASE
      WHEN LOWER(COALESCE(NULLIF(r.status, ''), 'failed')) = 'success' THEN 'success'
      ELSE 'failed'
    END AS status,
    COALESCE(r.latency_ms, 0) AS latency_ms
  FROM scheduled_test_results r
  JOIN scheduled_test_plans p ON p.id = r.plan_id
  JOIN account_groups ag ON ag.account_id = p.account_id
  WHERE ag.group_id = ANY($1)
    AND r.started_at >= $2
),
round_eval AS (
  SELECT
    group_id,
    round_key,
    MAX(started_at) AS round_started_at,
    MAX(CASE WHEN status = 'success' THEN 1 ELSE 0 END) AS has_success
  FROM raw
  GROUP BY group_id, round_key
),
round_rows AS (
  SELECT
    ranked.group_id,
    ranked.round_key,
    ranked.round_started_at,
    CASE WHEN ranked.has_success = 1 THEN 'success' ELSE 'failed' END AS round_status
  FROM (
    SELECT
      r.group_id,
      r.round_key,
      r.started_at,
      r.result_id,
      r.status,
      r.latency_ms,
      re.round_started_at,
      re.has_success,
      ROW_NUMBER() OVER (
        PARTITION BY r.group_id, r.round_key
        ORDER BY
          CASE
            WHEN re.has_success = 1 AND r.status = 'success' THEN 0
            WHEN re.has_success = 1 THEN 1
            ELSE 0
          END ASC,
          CASE
            WHEN re.has_success = 1 AND r.status = 'success' THEN r.latency_ms
            ELSE 0
          END ASC,
          r.started_at DESC,
          r.result_id DESC
      ) AS rn
    FROM raw r
    JOIN round_eval re ON re.group_id = r.group_id AND re.round_key = r.round_key
  ) ranked
  WHERE ranked.rn = 1
),
latest AS (
  SELECT DISTINCT ON (group_id)
    group_id,
    round_status
  FROM round_rows
  ORDER BY group_id, round_started_at DESC, round_key DESC
),
window_stats AS (
  SELECT
    group_id,
    COUNT(*) FILTER (WHERE round_started_at >= $4) AS total_requests_1h,
    COUNT(*) FILTER (WHERE round_started_at >= $4 AND round_status = 'success') AS success_requests_1h
  FROM round_rows
  GROUP BY group_id
)
SELECT
  gids.group_id,
  COALESCE(latest.round_status, 'unknown') AS current_status,
  COALESCE(window_stats.total_requests_1h, 0) AS total_requests_1h,
  COALESCE(window_stats.success_requests_1h, 0) AS success_requests_1h
FROM unnest($1::bigint[]) AS gids(group_id)
LEFT JOIN latest ON latest.group_id = gids.group_id
LEFT JOIN window_stats ON window_stats.group_id = gids.group_id
ORDER BY gids.group_id ASC
`

	rows, err := r.sql.QueryContext(
		ctx,
		q,
		pq.Array(groupIDs),
		historyStart,
		bucketSeconds,
		windowStart,
	)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			groupID         int64
			currentStatus   string
			totalRequests1h int64
			success1h       int64
		)
		if err := rows.Scan(&groupID, &currentStatus, &totalRequests1h, &success1h); err != nil {
			return err
		}
		agg := out[groupID]
		if agg == nil {
			continue
		}
		agg.CurrentStatus = normalizePublicGroupCurrentStatus(currentStatus)
		agg.TotalRequests1h = totalRequests1h
		agg.SuccessRequests1h = success1h
		agg.FailureRequests1h = totalRequests1h - success1h
	}

	return rows.Err()
}

func (r *groupRepository) loadPublicGroupMonitorSamples(
	ctx context.Context,
	out map[int64]*service.PublicGroupMonitorAggregate,
	groupIDs []int64,
	bucketSeconds int,
	sampleSize int,
	historyStart time.Time,
) error {
	const q = `
WITH raw AS (
  SELECT
    ag.group_id,
    r.id AS result_id,
    r.started_at,
    COALESCE(
      NULLIF(TRIM(COALESCE(r.round_id, '')), ''),
      'legacy:' || ((EXTRACT(EPOCH FROM r.started_at)::BIGINT / $3::BIGINT) * $3::BIGINT)::text
    ) AS round_key,
    CASE
      WHEN LOWER(COALESCE(NULLIF(r.status, ''), 'failed')) = 'success' THEN 'success'
      ELSE 'failed'
    END AS status,
    COALESCE(r.latency_ms, 0) AS latency_ms,
    COALESCE(NULLIF(TRIM(p.model_id), ''), '') AS model
  FROM scheduled_test_results r
  JOIN scheduled_test_plans p ON p.id = r.plan_id
  JOIN account_groups ag ON ag.account_id = p.account_id
  WHERE ag.group_id = ANY($1)
    AND r.started_at >= $2
),
round_eval AS (
  SELECT
    group_id,
    round_key,
    MAX(started_at) AS round_started_at,
    MAX(CASE WHEN status = 'success' THEN 1 ELSE 0 END) AS has_success
  FROM raw
  GROUP BY group_id, round_key
),
round_rows AS (
  SELECT
    ranked.group_id,
    ranked.round_key,
    ranked.round_started_at,
    ranked.started_at,
    CASE WHEN ranked.has_success = 1 THEN 'success' ELSE 'failed' END AS round_status,
    ranked.model,
    ranked.latency_ms
  FROM (
    SELECT
      r.group_id,
      r.round_key,
      r.started_at,
      r.result_id,
      r.status,
      r.model,
      r.latency_ms,
      re.round_started_at,
      re.has_success,
      ROW_NUMBER() OVER (
        PARTITION BY r.group_id, r.round_key
        ORDER BY
          CASE
            WHEN re.has_success = 1 AND r.status = 'success' THEN 0
            WHEN re.has_success = 1 THEN 1
            ELSE 0
          END ASC,
          CASE
            WHEN re.has_success = 1 AND r.status = 'success' THEN r.latency_ms
            ELSE 0
          END ASC,
          r.started_at DESC,
          r.result_id DESC
      ) AS rn
    FROM raw r
    JOIN round_eval re ON re.group_id = r.group_id AND re.round_key = r.round_key
  ) ranked
  WHERE ranked.rn = 1
),
ranked_samples AS (
  SELECT
    group_id,
    started_at,
    round_status,
    model,
    latency_ms,
    ROW_NUMBER() OVER (
      PARTITION BY group_id
      ORDER BY round_started_at DESC, round_key DESC, started_at DESC
    ) AS seq
  FROM round_rows
)
SELECT
  group_id,
  started_at,
  round_status,
  model,
  latency_ms
FROM ranked_samples
WHERE seq <= $4
ORDER BY group_id ASC, seq ASC
`

	rows, err := r.sql.QueryContext(
		ctx,
		q,
		pq.Array(groupIDs),
		historyStart,
		bucketSeconds,
		sampleSize,
	)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			groupID   int64
			startedAt time.Time
			status    string
			model     string
			latencyMs int64
		)
		if err := rows.Scan(&groupID, &startedAt, &status, &model, &latencyMs); err != nil {
			return err
		}
		agg := out[groupID]
		if agg == nil {
			continue
		}
		agg.Samples = append(agg.Samples, &service.PublicGroupMonitorSample{
			StartedAt: startedAt.UTC(),
			Status:    normalizePublicGroupSampleStatus(status),
			Model:     model,
			LatencyMs: latencyMs,
		})
	}

	return rows.Err()
}

func normalizePublicGroupCurrentStatus(status string) string {
	switch status {
	case "success":
		return "normal"
	case "failed":
		return "abnormal"
	default:
		return "unknown"
	}
}

func normalizePublicGroupSampleStatus(status string) string {
	if status == "success" {
		return "success"
	}
	return "failed"
}

func normalizeInt64IDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	out := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

var _ service.PublicGroupMonitorReader = (*groupRepository)(nil)
