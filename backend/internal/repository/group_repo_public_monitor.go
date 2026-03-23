package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

const (
	publicGroupMonitorHistoryMaxWindow        = 30 * 24 * time.Hour
	publicGroupMonitorLegacyBucketSeconds int = 15
)

// GetPublicGroupMonitorOverview aggregates scheduled-test results for public groups.
//
// Aggregation rules:
// - "same round" is merged by round_id (legacy data falls back to fixed time bucket)
// - within one round: any success => success, all failed => failed
// - current_status uses the latest round result only
func (r *groupRepository) GetPublicGroupMonitorOverview(
	ctx context.Context,
	groupIDs []int64,
	now time.Time,
) (map[int64]*service.PublicGroupMonitorAggregate, error) {
	result := make(map[int64]*service.PublicGroupMonitorAggregate, len(groupIDs))
	normalizedGroupIDs := normalizeInt64IDs(groupIDs)
	if len(normalizedGroupIDs) == 0 {
		return result, nil
	}

	now = now.UTC()
	historyStart := now.Add(-publicGroupMonitorHistoryMaxWindow)

	for _, groupID := range normalizedGroupIDs {
		result[groupID] = &service.PublicGroupMonitorAggregate{
			CurrentStatus: "unknown",
		}
	}

	if err := r.loadPublicGroupMonitorCurrentStatus(ctx, result, normalizedGroupIDs, historyStart); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *groupRepository) loadPublicGroupMonitorCurrentStatus(
	ctx context.Context,
	out map[int64]*service.PublicGroupMonitorAggregate,
	groupIDs []int64,
	historyStart time.Time,
) error {
	const q = `
WITH raw AS (
  SELECT
    ag.group_id,
    r.started_at,
    COALESCE(
      NULLIF(TRIM(COALESCE(r.round_id, '')), ''),
      'legacy:' || ((EXTRACT(EPOCH FROM r.started_at)::BIGINT / $3::BIGINT) * $3::BIGINT)::text
    ) AS round_key,
    CASE
      WHEN LOWER(COALESCE(NULLIF(r.status, ''), 'failed')) = 'success' THEN 'success'
      ELSE 'failed'
    END AS status
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
latest AS (
  SELECT DISTINCT ON (group_id)
    group_id,
    CASE WHEN has_success = 1 THEN 'success' ELSE 'failed' END AS round_status
  FROM round_eval
  ORDER BY group_id, round_started_at DESC, round_key DESC
)
SELECT
  gids.group_id,
  COALESCE(latest.round_status, 'unknown') AS current_status
FROM unnest($1::bigint[]) AS gids(group_id)
LEFT JOIN latest ON latest.group_id = gids.group_id
ORDER BY gids.group_id ASC
`

	rows, err := r.sql.QueryContext(
		ctx,
		q,
		pq.Array(groupIDs),
		historyStart,
		publicGroupMonitorLegacyBucketSeconds,
	)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			groupID       int64
			currentStatus string
		)
		if err := rows.Scan(&groupID, &currentStatus); err != nil {
			return err
		}
		agg := out[groupID]
		if agg == nil {
			continue
		}
		agg.CurrentStatus = normalizePublicGroupCurrentStatus(currentStatus)
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
