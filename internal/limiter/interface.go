package limiter

import "context"

// Limiter is the contract implemented by all rate limiting algorithms in GoShield.
//
// The interface is intentionally small so different algorithm strategies (for example,
// token bucket, fixed window, and sliding window variants) can be swapped behind a
// common API with minimal call-site overhead and latency.
//
// Implementations are expected to be concurrency-safe and optimized for high-throughput
// per-key decisions in multi-goroutine services.
//
// Allow evaluates whether the request identified by key should be accepted.
//
// Parameters:
//   - ctx: request-scoped context for cancellation/deadline propagation.
//   - key: the logical subject being rate limited (user, IP, API key, endpoint, etc.).
//
// Returns:
//   - allowed: true when the request is permitted.
//   - remaining: remaining quota/tokens for the key after the decision.
//   - err: evaluation/storage error; callers should treat non-nil as an indeterminate result.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, int, error)
}
