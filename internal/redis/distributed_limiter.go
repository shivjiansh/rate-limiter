package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v9"
)

type DistributedLimiter struct {
	client *goredis.Client
}

func NewDistributedLimiter(client *goredis.Client) *DistributedLimiter {
	return &DistributedLimiter{client: client}
}

func (d *DistributedLimiter) AllowFixedWindow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	res, err := d.client.Eval(ctx, fixedWindowLua, []string{key}, limit, window.Milliseconds()).Result()
	if err != nil {
		return false, 0, fmt.Errorf("eval fixed window: %w", err)
	}
	return parseAllowResult(res)
}

func (d *DistributedLimiter) AllowTokenBucket(ctx context.Context, key string, capacity int, refillPerMs float64) (bool, int, error) {
	now := float64(time.Now().UnixMilli())
	res, err := d.client.Eval(ctx, tokenBucketLua, []string{key}, capacity, refillPerMs, now).Result()
	if err != nil {
		return false, 0, fmt.Errorf("eval token bucket: %w", err)
	}
	return parseAllowResult(res)
}

func parseAllowResult(res interface{}) (bool, int, error) {
	vals, ok := res.([]interface{})
	if !ok || len(vals) != 2 {
		return false, 0, fmt.Errorf("unexpected eval response: %#v", res)
	}
	allowed, ok1 := vals[0].(int64)
	remaining, ok2 := vals[1].(int64)
	if !ok1 || !ok2 {
		return false, 0, fmt.Errorf("unexpected eval response types: %#v", vals)
	}
	return allowed == 1, int(remaining), nil
}
