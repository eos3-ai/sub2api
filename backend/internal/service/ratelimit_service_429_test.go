//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type rateLimitAccountRepo429Stub struct {
	mockAccountRepoForGemini
	setRateLimitedCalls int
	lastAccountID       int64
	lastResetAt         time.Time
}

func (r *rateLimitAccountRepo429Stub) SetRateLimited(ctx context.Context, id int64, resetAt time.Time) error {
	r.setRateLimitedCalls++
	r.lastAccountID = id
	r.lastResetAt = resetAt
	return nil
}

func TestRateLimitService_HandleUpstreamError_429FallbackCooldownFromConfig(t *testing.T) {
	tests := []struct {
		name           string
		resetHeaderVal string
	}{
		{name: "missing reset header"},
		{name: "invalid reset header", resetHeaderVal: "not-a-unix-timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &rateLimitAccountRepo429Stub{}
			cfg := &config.Config{
				RateLimit: config.RateLimitConfig{
					FallbackCooldownSeconds: 7,
				},
			}
			svc := NewRateLimitService(repo, nil, cfg, nil, nil)
			account := &Account{
				ID:       200,
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
			}

			headers := http.Header{}
			if tt.resetHeaderVal != "" {
				headers.Set("anthropic-ratelimit-unified-reset", tt.resetHeaderVal)
			}

			start := time.Now()
			shouldDisable := svc.HandleUpstreamError(context.Background(), account, 429, headers, []byte(`{"error":"rate limit"}`))

			require.False(t, shouldDisable)
			require.Equal(t, 1, repo.setRateLimitedCalls)
			require.Equal(t, int64(200), repo.lastAccountID)
			require.WithinDuration(t, start.Add(7*time.Second), repo.lastResetAt, 2*time.Second)
		})
	}
}

func TestRateLimitService_HandleUpstreamError_429FallbackCooldownMinutesBackwardCompatible(t *testing.T) {
	repo := &rateLimitAccountRepo429Stub{}
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			FallbackCooldownMinutes: 1,
		},
	}
	svc := NewRateLimitService(repo, nil, cfg, nil, nil)
	account := &Account{
		ID:       201,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
	}

	start := time.Now()
	shouldDisable := svc.HandleUpstreamError(context.Background(), account, 429, http.Header{}, []byte(`{"error":"rate limit"}`))

	require.False(t, shouldDisable)
	require.Equal(t, 1, repo.setRateLimitedCalls)
	require.Equal(t, int64(201), repo.lastAccountID)
	require.WithinDuration(t, start.Add(1*time.Minute), repo.lastResetAt, 2*time.Second)
}
