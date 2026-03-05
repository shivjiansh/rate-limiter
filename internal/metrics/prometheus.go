package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Registry struct {
	Requests *prometheus.CounterVec
	Blocked  *prometheus.CounterVec
	Latency  *prometheus.HistogramVec
}

func New() *Registry {
	r := &Registry{
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "rate_limiter_requests_total", Help: "Total rate limiter checks"}, []string{"scope", "endpoint"}),
		Blocked:  prometheus.NewCounterVec(prometheus.CounterOpts{Name: "rate_limiter_blocked_total", Help: "Blocked requests"}, []string{"scope", "endpoint"}),
		Latency:  prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "rate_limiter_latency", Help: "Rate limiter latency", Buckets: prometheus.DefBuckets}, []string{"scope", "endpoint"}),
	}
	prometheus.MustRegister(r.Requests, r.Blocked, r.Latency)
	return r
}

func (r *Registry) ObserveRequest(scope, endpoint string, allowed bool, d time.Duration) {
	r.Requests.WithLabelValues(scope, endpoint).Inc()
	if !allowed {
		r.Blocked.WithLabelValues(scope, endpoint).Inc()
	}
	r.Latency.WithLabelValues(scope, endpoint).Observe(d.Seconds())
}
