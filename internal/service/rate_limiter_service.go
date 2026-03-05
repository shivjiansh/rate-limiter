package service

import (
	"context"
	"fmt"
	"time"

	"github.com/example/go-rate-limiter/internal/limiter"
	"github.com/example/go-rate-limiter/internal/metrics"
	"github.com/example/go-rate-limiter/internal/redis"
)

type Scope string

const (
	ScopeIP       Scope = "ip"
	ScopeUser     Scope = "user"
	ScopeAPIKey   Scope = "api_key"
	ScopeEndpoint Scope = "endpoint"
	ScopeGlobal   Scope = "global"
)

type Request struct {
	Scope      Scope
	Identifier string
	Endpoint   string
}

type RateLimiterService struct {
	local   limiter.Limiter
	redis   *redis.DistributedLimiter
	limit   int
	window  time.Duration
	metrics *metrics.Registry
}

func NewRateLimiterService(local limiter.Limiter, redis *redis.DistributedLimiter, limit int, window time.Duration, m *metrics.Registry) *RateLimiterService {
	return &RateLimiterService{local: local, redis: redis, limit: limit, window: window, metrics: m}
}

func (s *RateLimiterService) Allow(ctx context.Context, req Request) (bool, int, error) {
	start := time.Now()
	key := s.key(req)
	allowed, remaining := s.local.Allow(key)
	if !allowed && s.redis != nil {
		ra, rr, err := s.redis.AllowFixedWindow(ctx, key, s.limit, s.window)
		if err != nil {
			return false, 0, err
		}
		allowed, remaining = ra, rr
	}
	s.metrics.ObserveRequest(string(req.Scope), req.Endpoint, allowed, time.Since(start))
	return allowed, remaining, nil
}

func (s *RateLimiterService) key(req Request) string {
	return fmt.Sprintf("rl:%s:%s:%s", req.Scope, req.Endpoint, req.Identifier)
}
