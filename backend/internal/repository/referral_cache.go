package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	referralCodeCachePrefix = "referral:code:"
	referralCodeCacheTTL    = 24 * time.Hour
)

func referralCodeCacheKey(code string) string {
	return referralCodeCachePrefix + code
}

type referralCache struct {
	rdb *redis.Client
}

func NewReferralCache(rdb *redis.Client) service.ReferralCache {
	return &referralCache{rdb: rdb}
}

func (c *referralCache) GetUserIDByCode(ctx context.Context, code string) (int64, error) {
	key := referralCodeCacheKey(code)
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

func (c *referralCache) SetUserIDByCode(ctx context.Context, code string, userID int64, ttl time.Duration) error {
	key := referralCodeCacheKey(code)
	if ttl <= 0 {
		ttl = referralCodeCacheTTL
	}
	return c.rdb.Set(ctx, key, fmt.Sprintf("%d", userID), ttl).Err()
}

func (c *referralCache) InvalidateCode(ctx context.Context, code string) error {
	key := referralCodeCacheKey(code)
	return c.rdb.Del(ctx, key).Err()
}
