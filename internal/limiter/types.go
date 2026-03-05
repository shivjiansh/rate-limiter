package limiter

import "time"

// Policy describes a single rate limit policy.
type Policy struct {
	Algorithm  string
	Limit      int
	Window     time.Duration
	Burst      int
	LeakRate   float64
	ShardCount int
}
