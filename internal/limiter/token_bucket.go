package limiter

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

// tokenBucketState holds per-key mutable token bucket data.
//
// tokens is fractional to preserve refill precision between requests.
// lastRefill tracks when we last replenished this key.
type tokenBucketState struct {
	tokens     float64
	lastRefill time.Time
}

// tokenBucketShard owns a subset of keys.
//
// Sharding reduces lock contention at high throughput: requests for different keys
// can proceed in parallel as long as they hash to different shards.
type tokenBucketShard struct {
	mu      sync.Mutex
	buckets map[string]tokenBucketState
}

// TokenBucket is a high-performance, sharded, per-key token bucket limiter.
//
// Algorithm behavior:
//   - Each key has an independent bucket with capacity tokens.
//   - Tokens are refilled continuously over time at refillRate tokens/second.
//   - A request consumes one token if available; otherwise it is rejected.
//
// Burst handling:
//   - Capacity controls maximum burst: a key can consume up to capacity immediate
//     requests after being idle long enough to refill.
//
// Complexity:
//   - Allow is O(1): hash, shard lock, single map lookup/update, constant-time math.
//
// Scaling characteristics:
//   - Lock sharding (hash(key) % shard_count) minimizes global contention.
//   - Per-key state is isolated, making the limiter suitable for very high-cardinality
//     identities (IP/user/API-key) in multi-core services.
type TokenBucket struct {
	capacity   float64
	refillRate float64
	shards     []tokenBucketShard
}

// NewTokenBucket creates a limiter using practical defaults.
//
// Defaults:
//   - capacity: 100 tokens
//   - refillRate: 100 tokens/second
//   - shardCount: 256
func NewTokenBucket() *TokenBucket {
	l, _ := NewTokenBucketWithConfig(100, 100, 256)
	return l
}

// NewTokenBucketWithConfig creates a sharded token bucket limiter.
func NewTokenBucketWithConfig(capacity int, refillRate float64, shardCount int) (*TokenBucket, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be > 0")
	}
	if refillRate <= 0 {
		return nil, fmt.Errorf("refillRate must be > 0")
	}
	if shardCount <= 0 {
		return nil, fmt.Errorf("shardCount must be > 0")
	}

	tb := &TokenBucket{
		capacity:   float64(capacity),
		refillRate: refillRate,
		shards:     make([]tokenBucketShard, shardCount),
	}
	for i := range tb.shards {
		tb.shards[i].buckets = make(map[string]tokenBucketState)
	}
	return tb, nil
}

// Allow applies token-bucket admission control for a single key.
func (t *TokenBucket) Allow(ctx context.Context, key string) (bool, int, error) {
	if err := ctx.Err(); err != nil {
		return false, 0, err
	}
	if key == "" {
		return false, 0, fmt.Errorf("key must not be empty")
	}

	now := time.Now()
	shard := &t.shards[t.shardIndex(key)]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	state, ok := shard.buckets[key]
	if !ok {
		// New keys start full, enabling immediate bursts up to capacity.
		state = tokenBucketState{tokens: t.capacity, lastRefill: now}
	}

	// Refill based on elapsed wall-clock time since last decision.
	elapsedSeconds := now.Sub(state.lastRefill).Seconds()
	state.tokens = minFloat(t.capacity, state.tokens+(elapsedSeconds*t.refillRate))
	state.lastRefill = now

	if state.tokens < 1 {
		shard.buckets[key] = state
		return false, int(state.tokens), nil
	}

	state.tokens -= 1
	shard.buckets[key] = state
	return true, int(state.tokens), nil
}

// shardIndex implements the required sharding strategy: hash(key) % shard_count.
func (t *TokenBucket) shardIndex(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32() % uint32(len(t.shards))
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
