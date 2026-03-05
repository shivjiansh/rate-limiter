package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Registry stores in-process counters/histograms exposed in Prometheus text format.
// Metric names follow Prometheus naming conventions required by GoShield.
type Registry struct {
	mu sync.RWMutex

	requests map[string]uint64
	blocked  map[string]uint64
	latency  map[string]latencySample
}

type latencySample struct {
	sum   float64
	count uint64
}

func New() *Registry {
	return &Registry{
		requests: make(map[string]uint64),
		blocked:  make(map[string]uint64),
		latency:  make(map[string]latencySample),
	}
}

func labelsKey(endpoint, status, algorithm string) string {
	return fmt.Sprintf("endpoint=%q,status=%q,algorithm=%q", endpoint, status, algorithm)
}

// ObserveRateLimit records metrics for one request decision.
func (r *Registry) ObserveRateLimit(endpoint, status, algorithm string, blocked bool, duration time.Duration) {
	k := labelsKey(endpoint, status, algorithm)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.requests[k]++
	if blocked {
		r.blocked[k]++
	}

	s := r.latency[k]
	s.sum += duration.Seconds()
	s.count++
	r.latency[k] = s
}

// Handler returns an HTTP handler that exposes metrics on /metrics.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(r.renderPrometheusText()))
	})
}

func (r *Registry) renderPrometheusText() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var b strings.Builder

	b.WriteString("# HELP rate_limiter_requests_total Total number of requests seen by the rate limiter middleware.\n")
	b.WriteString("# TYPE rate_limiter_requests_total counter\n")
	writeSeries(&b, "rate_limiter_requests_total", r.requests)

	b.WriteString("# HELP rate_limiter_blocked_total Total number of blocked requests.\n")
	b.WriteString("# TYPE rate_limiter_blocked_total counter\n")
	writeSeries(&b, "rate_limiter_blocked_total", r.blocked)

	b.WriteString("# HELP rate_limiter_latency_seconds Rate limiter middleware latency in seconds.\n")
	b.WriteString("# TYPE rate_limiter_latency_seconds summary\n")
	writeLatencySeries(&b, "rate_limiter_latency_seconds", r.latency)

	return b.String()
}

func writeSeries(b *strings.Builder, metric string, values map[string]uint64) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("%s{%s} %d\n", metric, k, values[k]))
	}
}

func writeLatencySeries(b *strings.Builder, metric string, values map[string]latencySample) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s := values[k]
		b.WriteString(fmt.Sprintf("%s_sum{%s} %.9f\n", metric, k, s.sum))
		b.WriteString(fmt.Sprintf("%s_count{%s} %d\n", metric, k, s.count))
	}
}
