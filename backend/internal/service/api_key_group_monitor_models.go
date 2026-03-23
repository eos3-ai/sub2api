package service

import "time"

// PublicGroupMonitorItem is user-facing group monitor summary.
type PublicGroupMonitorItem struct {
	GroupID   int64  `json:"group_id"`
	GroupName string `json:"group_name"`
	Platform  string `json:"platform"`

	CurrentStatus string `json:"current_status"` // normal | abnormal | unknown
}

// PublicGroupMonitorResponse is returned to user-side monitor page.
type PublicGroupMonitorResponse struct {
	GeneratedAt    time.Time `json:"generated_at"`
	PublicGroupNum int       `json:"public_group_num"`

	Items []*PublicGroupMonitorItem `json:"items"`
}

// PublicGroupMonitorAggregate contains repository aggregation payload for one group.
type PublicGroupMonitorAggregate struct {
	CurrentStatus string
}
