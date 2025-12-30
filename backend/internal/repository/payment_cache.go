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
	paymentOrderCounterPrefix = "payment:counter:"
	paymentURLPrefix          = "payment:url:"
	defaultOrderCounterTTL    = time.Minute
	defaultPaymentURLTTL      = 30 * time.Minute
)

func paymentCounterKey(userID int64) string {
	return fmt.Sprintf("%s%d", paymentOrderCounterPrefix, userID)
}

func paymentURLKey(orderNo string) string {
	return paymentURLPrefix + orderNo
}

type paymentCache struct {
	rdb *redis.Client
}

func NewPaymentCache(rdb *redis.Client) service.PaymentCache {
	return &paymentCache{rdb: rdb}
}

func (c *paymentCache) IncrementUserOrderCounter(ctx context.Context, userID int64, ttl time.Duration) (int, error) {
	key := paymentCounterKey(userID)
	result, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if ttl <= 0 {
		ttl = defaultOrderCounterTTL
	}
	// Always refresh TTL to enforce rolling window
	c.rdb.Expire(ctx, key, ttl)
	return int(result), nil
}

func (c *paymentCache) GetUserOrderCounter(ctx context.Context, userID int64) (int, error) {
	key := paymentCounterKey(userID)
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	count, err := strconv.ParseInt(val, 10, 64)
	return int(count), err
}

func (c *paymentCache) ResetUserOrderCounter(ctx context.Context, userID int64) error {
	key := paymentCounterKey(userID)
	return c.rdb.Del(ctx, key).Err()
}

func (c *paymentCache) SetPaymentURL(ctx context.Context, orderNo, url string, ttl time.Duration) error {
	key := paymentURLKey(orderNo)
	if ttl <= 0 {
		ttl = defaultPaymentURLTTL
	}
	return c.rdb.Set(ctx, key, url, ttl).Err()
}

func (c *paymentCache) GetPaymentURL(ctx context.Context, orderNo string) (string, error) {
	key := paymentURLKey(orderNo)
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

func (c *paymentCache) DeletePaymentURL(ctx context.Context, orderNo string) error {
	key := paymentURLKey(orderNo)
	return c.rdb.Del(ctx, key).Err()
}
