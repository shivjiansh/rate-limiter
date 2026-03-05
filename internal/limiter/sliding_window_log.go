package limiter

import (
	"hash/fnv"
	"sync"
	"time"
)

type slidingWindowLogShard struct {
	mu     sync.Mutex
	values map[string][]int64
}

// SlidingWindowLogLimiter stores timestamps for exact rolling windows.
type SlidingWindowLogLimiter struct {
	limit  int
	window time.Duration
	shards []slidingWindowLogShard
}

func NewSlidingWindowLogLimiter(limit int, window time.Duration, shards int) *SlidingWindowLogLimiter {
	if shards <= 0 {
		shards = 256
	}
	l := &SlidingWindowLogLimiter{limit: limit, window: window, shards: make([]slidingWindowLogShard, shards)}
	for i := range l.shards {
		l.shards[i].values = make(map[string][]int64)
	}
	return l
}

func (l *SlidingWindowLogLimiter) Allow(key string) (bool, int) {
	now := time.Now().UnixNano()
	cutoff := now - l.window.Nanoseconds()
	s := &l.shards[l.shardIndex(key)]
	s.mu.Lock()
	defer s.mu.Unlock()
	arr := s.values[key]
	idx := 0
	for idx < len(arr) && arr[idx] <= cutoff {
		idx++
	}
	arr = arr[idx:]
	if len(arr) >= l.limit {
		s.values[key] = arr
		return false, 0
	}
	arr = append(arr, now)
	s.values[key] = arr
	return true, l.limit - len(arr)
}

func (l *SlidingWindowLogLimiter) shardIndex(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32() % uint32(len(l.shards))
}
