package limiter

import (
	"hash/fnv"
	"sync"
	"time"
)

type fixedWindowEntry struct {
	count       int
	windowStart int64
}

type fixedWindowShard struct {
	mu     sync.Mutex
	values map[string]fixedWindowEntry
}

// FixedWindowLimiter is a lock-sharded in-memory fixed-window limiter.
type FixedWindowLimiter struct {
	limit      int
	windowSize time.Duration
	shards     []fixedWindowShard
}

func NewFixedWindowLimiter(limit int, window time.Duration, shards int) *FixedWindowLimiter {
	if shards <= 0 {
		shards = 256
	}
	l := &FixedWindowLimiter{limit: limit, windowSize: window, shards: make([]fixedWindowShard, shards)}
	for i := range l.shards {
		l.shards[i].values = make(map[string]fixedWindowEntry)
	}
	return l
}

func (l *FixedWindowLimiter) Allow(key string) (bool, int) {
	idx := l.shardIndex(key)
	s := &l.shards[idx]
	now := time.Now().UnixNano()
	windowStart := (now / l.windowSize.Nanoseconds()) * l.windowSize.Nanoseconds()

	s.mu.Lock()
	defer s.mu.Unlock()

	e := s.values[key]
	if e.windowStart != windowStart {
		e.windowStart = windowStart
		e.count = 0
	}
	if e.count >= l.limit {
		s.values[key] = e
		return false, 0
	}
	e.count++
	s.values[key] = e
	return true, l.limit - e.count
}

func (l *FixedWindowLimiter) shardIndex(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32() % uint32(len(l.shards))
}
