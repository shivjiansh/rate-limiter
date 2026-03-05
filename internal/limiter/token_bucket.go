package limiter

import (
	"hash/fnv"
	"sync"
	"time"
)

type tokenBucketEntry struct {
	tokens     float64
	lastRefill time.Time
}

type tokenBucketShard struct {
	mu     sync.Mutex
	values map[string]tokenBucketEntry
}

// TokenBucketLimiter supports burst traffic and steady refill.
type TokenBucketLimiter struct {
	capacity   float64
	refillRate float64
	shards     []tokenBucketShard
}

func NewTokenBucketLimiter(capacity int, refillPerSec float64, shards int) *TokenBucketLimiter {
	if shards <= 0 {
		shards = 256
	}
	l := &TokenBucketLimiter{capacity: float64(capacity), refillRate: refillPerSec, shards: make([]tokenBucketShard, shards)}
	for i := range l.shards {
		l.shards[i].values = make(map[string]tokenBucketEntry)
	}
	return l
}

func (l *TokenBucketLimiter) Allow(key string) (bool, int) {
	now := time.Now()
	s := &l.shards[l.shardIndex(key)]
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.values[key]
	if !ok {
		e = tokenBucketEntry{tokens: l.capacity, lastRefill: now}
	}
	elapsed := now.Sub(e.lastRefill).Seconds()
	e.tokens = minF(l.capacity, e.tokens+elapsed*l.refillRate)
	e.lastRefill = now
	if e.tokens < 1 {
		s.values[key] = e
		return false, int(e.tokens)
	}
	e.tokens -= 1
	s.values[key] = e
	return true, int(e.tokens)
}

func (l *TokenBucketLimiter) shardIndex(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32() % uint32(len(l.shards))
}

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
