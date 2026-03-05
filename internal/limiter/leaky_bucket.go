package limiter

import (
	"hash/fnv"
	"sync"
	"time"
)

type leakyEntry struct {
	water    float64
	lastLeak time.Time
}

type leakyShard struct {
	mu     sync.Mutex
	values map[string]leakyEntry
}

// LeakyBucketLimiter smooths burst traffic with constant outflow.
type LeakyBucketLimiter struct {
	capacity float64
	leakRate float64
	shards   []leakyShard
}

func NewLeakyBucketLimiter(capacity int, leakRate float64, shards int) *LeakyBucketLimiter {
	if shards <= 0 {
		shards = 256
	}
	l := &LeakyBucketLimiter{capacity: float64(capacity), leakRate: leakRate, shards: make([]leakyShard, shards)}
	for i := range l.shards {
		l.shards[i].values = make(map[string]leakyEntry)
	}
	return l
}

func (l *LeakyBucketLimiter) Allow(key string) (bool, int) {
	now := time.Now()
	s := &l.shards[l.shardIndex(key)]
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.values[key]
	if !ok {
		e = leakyEntry{lastLeak: now}
	}
	leaked := now.Sub(e.lastLeak).Seconds() * l.leakRate
	e.water = maxF(0, e.water-leaked)
	e.lastLeak = now
	if e.water+1 > l.capacity {
		s.values[key] = e
		return false, int(l.capacity - e.water)
	}
	e.water += 1
	s.values[key] = e
	return true, int(l.capacity - e.water)
}

func (l *LeakyBucketLimiter) shardIndex(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32() % uint32(len(l.shards))
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
