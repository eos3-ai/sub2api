package service

import (
	"context"
	"log/slog"
)

func (s *RateLimitService) ResetOpenAI403Counter(ctx context.Context, accountID int64) {
	if s == nil || s.openAI403CounterCache == nil || accountID <= 0 {
		return
	}
	if err := s.openAI403CounterCache.ResetOpenAI403Count(ctx, accountID); err != nil {
		slog.Warn("openai_403_reset_failed", "account_id", accountID, "error", err)
	}
}
