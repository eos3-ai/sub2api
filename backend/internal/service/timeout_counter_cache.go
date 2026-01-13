package service

import (
	"context"
	"time"
)

// TimeoutCounterCache tracks per-account timeout counters in a time window.
// It is used by RateLimitService for stream-timeout handling.
type TimeoutCounterCache interface {
	IncrementTimeoutCount(ctx context.Context, accountID int64, windowMinutes int) (int64, error)
	GetTimeoutCount(ctx context.Context, accountID int64) (int64, error)
	ResetTimeoutCount(ctx context.Context, accountID int64) error
	GetTimeoutCountTTL(ctx context.Context, accountID int64) (time.Duration, error)
}

