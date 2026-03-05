package limiter

import (
	"hash/fnv"
	"math"
	"sync"
	"time"
)

type swcEntry struct {
	currentWindow int64
	currentCount  int
	prevWindow    int64
	prevCount     int
}

type swcShard struct {
	mu     sync.Mutex
	values map[string]swcEntry
}

// SlidingWindowCounterLimiter approximates rolling windows using weighted buckets.
type SlidingWindowCounterLimiter struct {
	limit  int
	window time.Duration
	shards []swcShard
}

func NewSlidingWindowCounterLimiter(limit int, window time.Duration, shards int) *SlidingWindowCounterLimiter {
	if shards <= 0 {
		shards = 256
	}
	l := &SlidingWindowCounterLimiter{limit: limit, window: window, shards: make([]swcShard, shards)}
	for i := range l.shards {
		l.shards[i].values = make(map[string]swcEntry)
	}
	return l
}

func (l *SlidingWindowCounterLimiter) Allow(key string) (bool, int) {
	now := time.Now().UnixNano()
	windowNs := l.window.Nanoseconds()
	currWindow := (now / windowNs) * windowNs
	s := &l.shards[l.shardIndex(key)]
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.values[key]
	if e.currentWindow != currWindow {
		e.prevWindow, e.prevCount = e.currentWindow, e.currentCount
		e.currentWindow, e.currentCount = currWindow, 0
	}
	weight := float64(currWindow+windowNs-now) / float64(windowNs)
	est := float64(e.currentCount) + float64(e.prevCount)*math.Max(weight, 0)
	if int(est) >= l.limit {
		s.values[key] = e
		return false, 0
	}
	e.currentCount++
	s.values[key] = e
	remaining := l.limit - (e.currentCount + int(float64(e.prevCount)*math.Max(weight, 0)))
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining
}

func (l *SlidingWindowCounterLimiter) shardIndex(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32() % uint32(len(l.shards))
}
