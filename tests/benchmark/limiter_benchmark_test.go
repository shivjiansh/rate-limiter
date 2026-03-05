package benchmark

import (
	"context"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/goshield/rate-limiter/internal/limiter"
)

const (
	simulatedUsers = 10_000
	targetRPS      = 100_000.0
)

// BenchmarkTokenBucket_10KUsers_100KRPS runs a high-cardinality parallel benchmark
// that simulates 10k concurrent users by spreading requests across 10k unique keys.
//
// Reported custom metrics:
//   - rps: achieved requests/second throughput
//   - reject_rate: blocked_requests / total_requests
//   - avg_latency_ms: average per-request decision latency
func BenchmarkTokenBucket_10KUsers_100KRPS(b *testing.B) {
	tb, err := limiter.NewTokenBucketWithConfig(200_000, 200_000, 1024)
	if err != nil {
		b.Fatalf("new token bucket: %v", err)
	}

	var requestID uint64
	var rejected uint64
	var latencyTotalNs uint64

	start := time.Now()
	b.ResetTimer()
	b.ReportAllocs()
	// Increase worker fan-out to stress lock sharding paths.
	b.SetParallelism(runtime.GOMAXPROCS(0) * 16)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			id := atomic.AddUint64(&requestID, 1)
			userKey := "user-" + strconv.FormatUint(id%simulatedUsers, 10)

			t0 := time.Now()
			allowed, _, err := tb.Allow(ctx, userKey)
			latency := time.Since(t0)
			if err != nil {
				b.Fatalf("allow error: %v", err)
			}
			atomic.AddUint64(&latencyTotalNs, uint64(latency.Nanoseconds()))
			if !allowed {
				atomic.AddUint64(&rejected, 1)
			}
		}
	})
	elapsed := time.Since(start)

	total := uint64(b.N)
	rps := float64(total) / elapsed.Seconds()
	rejectRate := float64(rejected) / float64(total)
	avgLatencyMs := float64(latencyTotalNs) / float64(total) / 1e6

	b.ReportMetric(rps, "rps")
	b.ReportMetric(targetRPS, "target_rps")
	b.ReportMetric(rejectRate, "reject_rate")
	b.ReportMetric(avgLatencyMs, "avg_latency_ms")
}

// BenchmarkTokenBucket_BurstTraffic models burst traffic by sending large request
// volumes to a tiny key-set with bucket capacity intentionally lower than burst size.
// This makes rejection behavior measurable under contention.
func BenchmarkTokenBucket_BurstTraffic(b *testing.B) {
	tb, err := limiter.NewTokenBucketWithConfig(100, 10, 64)
	if err != nil {
		b.Fatalf("new token bucket: %v", err)
	}

	burstKeys := []string{"burst-a", "burst-b", "burst-c", "burst-d"}

	var rejected uint64
	var latencyTotalNs uint64
	start := time.Now()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		localI := 0
		for pb.Next() {
			key := burstKeys[localI%len(burstKeys)]
			localI++

			t0 := time.Now()
			allowed, _, err := tb.Allow(ctx, key)
			latency := time.Since(t0)
			if err != nil {
				b.Fatalf("allow error: %v", err)
			}
			atomic.AddUint64(&latencyTotalNs, uint64(latency.Nanoseconds()))
			if !allowed {
				atomic.AddUint64(&rejected, 1)
			}
		}
	})

	elapsed := time.Since(start)
	total := uint64(b.N)

	b.ReportMetric(float64(total)/elapsed.Seconds(), "rps")
	b.ReportMetric(float64(rejected)/float64(total), "reject_rate")
	b.ReportMetric(float64(latencyTotalNs)/float64(total)/1e6, "avg_latency_ms")
}
