package unit

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goshield/rate-limiter/internal/metrics"
	"github.com/goshield/rate-limiter/internal/middleware"
)

func TestMetricsMiddlewareRecordsRequestsAndBlocked(t *testing.T) {
	reg := metrics.New()
	mw := middleware.NewMetricsMiddleware(reg, "token_bucket")

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	mreq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mrr := httptest.NewRecorder()
	reg.Handler().ServeHTTP(mrr, mreq)
	body := mrr.Body.String()

	if !strings.Contains(body, "rate_limiter_requests_total") {
		t.Fatalf("expected requests metric in output")
	}
	if !strings.Contains(body, "rate_limiter_blocked_total") {
		t.Fatalf("expected blocked metric in output")
	}
	if !strings.Contains(body, `endpoint="/v1/test",status="429",algorithm="token_bucket"`) {
		t.Fatalf("expected labels in metrics output, got: %s", body)
	}
}

func TestLatencyMetricOutput(t *testing.T) {
	reg := metrics.New()
	reg.ObserveRateLimit("/v1/latency", "200", "token_bucket", false, 25*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	reg.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()

	if !strings.Contains(body, "rate_limiter_latency_seconds_sum") || !strings.Contains(body, "rate_limiter_latency_seconds_count") {
		t.Fatalf("expected latency summary series in output")
	}
}
