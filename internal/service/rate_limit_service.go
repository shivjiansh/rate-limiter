package service

import (
	"context"

	"github.com/goshield/rate-limiter/internal/limiter"
)

type RateLimitService struct {
	Limiter limiter.Limiter
}

func NewRateLimitService(l limiter.Limiter) *RateLimitService {
	return &RateLimitService{Limiter: l}
}

func (s *RateLimitService) Allow(ctx context.Context, key string) (bool, int, error) {
	return s.Limiter.Allow(ctx, key)
}
