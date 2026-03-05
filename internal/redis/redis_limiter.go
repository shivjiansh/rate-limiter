package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Limiter is a Redis-backed distributed sliding window rate limiter.
type Limiter struct {
	client     *goredis.Client
	script     *goredis.Script
	limit      int64
	window     time.Duration
	endpoint   string
	retryCount int
}

// NewLimiter creates a sliding-window limiter.
// endpoint is included in the Redis key format: rate_limit:{key}:{endpoint}.
func NewLimiter(client *goredis.Client, limit int, window time.Duration, endpoint string, retryCount int) (*Limiter, error) {
	if client == nil {
		return nil, fmt.Errorf("nil redis client")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be > 0")
	}
	if window <= 0 {
		return nil, fmt.Errorf("window must be > 0")
	}
	if endpoint == "" {
		endpoint = "global"
	}
	if retryCount < 0 {
		retryCount = 0
	}

	return &Limiter{
		client:     client,
		script:     goredis.NewScript(SlidingWindowScript),
		limit:      int64(limit),
		window:     window,
		endpoint:   endpoint,
		retryCount: retryCount,
	}, nil
}

// Allow evaluates a request for a logical key.
func (l *Limiter) Allow(ctx context.Context, key string) (bool, int, error) {
	if err := ctx.Err(); err != nil {
		return false, 0, err
	}
	if key == "" {
		return false, 0, fmt.Errorf("key must not be empty")
	}

	redisKey := l.redisKey(key, l.endpoint)
	nowMs := time.Now().UnixMilli()
	member := fmt.Sprintf("%d-%d", nowMs, time.Now().UnixNano())

	var res interface{}
	err := l.withRetry(ctx, func() error {
		var runErr error
		res, runErr = l.script.Run(ctx, l.client, []string{redisKey}, nowMs, l.window.Milliseconds(), l.limit, member).Result()
		return runErr
	})
	if err != nil {
		return false, 0, fmt.Errorf("redis sliding-window eval failed: %w", err)
	}

	allowed, remaining, err := parseResult(res)
	if err != nil {
		return false, 0, err
	}
	return allowed, remaining, nil
}

func (l *Limiter) redisKey(key, endpoint string) string {
	return fmt.Sprintf("rate_limit:%s:%s", key, endpoint)
}

func (l *Limiter) withRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= l.retryCount; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if !isRetryable(err) || attempt == l.retryCount {
			break
		}

		backoff := time.Duration(attempt+1) * 10 * time.Millisecond
		t := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
	return lastErr
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if err == goredis.Nil {
		return false
	}
	return true
}

func parseResult(res interface{}) (bool, int, error) {
	vals, ok := res.([]interface{})
	if !ok || len(vals) != 2 {
		return false, 0, fmt.Errorf("unexpected redis script response: %#v", res)
	}
	allowed, okA := vals[0].(int64)
	remaining, okR := vals[1].(int64)
	if !okA || !okR {
		return false, 0, fmt.Errorf("unexpected redis response types: %#v", vals)
	}
	if remaining < 0 {
		remaining = 0
	}
	return allowed == 1, int(remaining), nil
}
