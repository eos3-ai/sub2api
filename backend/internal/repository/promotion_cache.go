package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	promotionCacheKeyPrefix = "promotion:user:"
	promotionCacheTTL       = 30 * time.Minute
)

func promotionCacheKey(userID int64) string {
	return fmt.Sprintf("%s%d", promotionCacheKeyPrefix, userID)
}

type promotionCache struct {
	rdb *redis.Client
}

func NewPromotionCache(rdb *redis.Client) service.PromotionCache {
	return &promotionCache{rdb: rdb}
}

func (c *promotionCache) GetUserPromotion(ctx context.Context, userID int64) (*service.UserPromotion, error) {
	key := promotionCacheKey(userID)
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var promotion service.UserPromotion
	if err := json.Unmarshal(val, &promotion); err != nil {
		return nil, err
	}
	return &promotion, nil
}

func (c *promotionCache) SetUserPromotion(ctx context.Context, promotion *service.UserPromotion, ttl time.Duration) error {
	if promotion == nil {
		return nil
	}
	data, err := json.Marshal(promotion)
	if err != nil {
		return err
	}
	key := promotionCacheKey(promotion.UserID)
	if ttl <= 0 {
		ttl = promotionCacheTTL
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

func (c *promotionCache) InvalidateUserPromotion(ctx context.Context, userID int64) error {
	key := promotionCacheKey(userID)
	return c.rdb.Del(ctx, key).Err()
}
