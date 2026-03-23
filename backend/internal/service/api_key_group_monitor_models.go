package service

import "time"

// PublicGroupMonitorQuery controls user-facing public group monitor aggregation.
type PublicGroupMonitorQuery struct {
	// Window controls 1h metrics aggregation horizon.
	Window time.Duration
	// SampleSize limits latest merged group samples.
	SampleSize int
	// BucketSeconds defines legacy fallback bucket width when round_id is missing.
	BucketSeconds int
}

// PublicGroupMonitorSample represents one merged sample point for a group.
type PublicGroupMonitorSample struct {
	StartedAt time.Time `json:"started_at"`
	Status    string    `json:"status"`
	Model     string    `json:"model"`
	LatencyMs int64     `json:"latency_ms"`
}

// PublicGroupMonitorItem is user-facing group monitor summary.
type PublicGroupMonitorItem struct {
	GroupID   int64  `json:"group_id"`
	GroupName string `json:"group_name"`
	Platform  string `json:"platform"`

	CurrentStatus string `json:"current_status"` // normal | abnormal | unknown

	TotalRequests1h   int64 `json:"total_requests_1h"`
	SuccessRequests1h int64 `json:"success_requests_1h"`
	FailureRequests1h int64 `json:"failure_requests_1h"`

	Samples []*PublicGroupMonitorSample `json:"samples"`
}

// PublicGroupMonitorResponse is returned to user-side monitor page.
type PublicGroupMonitorResponse struct {
	GeneratedAt    time.Time `json:"generated_at"`
	WindowSeconds  int64     `json:"window_seconds"`
	SampleSize     int       `json:"sample_size"`
	BucketSeconds  int       `json:"bucket_seconds"`
	PublicGroupNum int       `json:"public_group_num"`

	Items []*PublicGroupMonitorItem `json:"items"`
}

// PublicGroupMonitorAggregate contains repository aggregation payload for one group.
type PublicGroupMonitorAggregate struct {
	CurrentStatus string

	TotalRequests1h   int64
	SuccessRequests1h int64
	FailureRequests1h int64

	Samples []*PublicGroupMonitorSample
}
