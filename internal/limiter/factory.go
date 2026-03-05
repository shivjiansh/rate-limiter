package limiter

import "fmt"

func NewFromPolicy(p Policy) (Limiter, error) {
	switch p.Algorithm {
	case "fixed_window":
		return NewFixedWindowLimiter(p.Limit, p.Window, p.ShardCount), nil
	case "token_bucket":
		refill := float64(p.Limit) / p.Window.Seconds()
		if p.Burst > 0 {
			return NewTokenBucketLimiter(p.Burst, refill, p.ShardCount), nil
		}
		return NewTokenBucketLimiter(p.Limit, refill, p.ShardCount), nil
	case "sliding_window_log":
		return NewSlidingWindowLogLimiter(p.Limit, p.Window, p.ShardCount), nil
	case "sliding_window_counter":
		return NewSlidingWindowCounterLimiter(p.Limit, p.Window, p.ShardCount), nil
	case "leaky_bucket":
		leak := p.LeakRate
		if leak == 0 {
			leak = float64(p.Limit) / p.Window.Seconds()
		}
		cap := p.Burst
		if cap == 0 {
			cap = p.Limit
		}
		return NewLeakyBucketLimiter(cap, leak, p.ShardCount), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", p.Algorithm)
	}
}
