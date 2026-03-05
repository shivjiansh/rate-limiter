package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goshield/rate-limiter/internal/metrics"
	"github.com/goshield/rate-limiter/internal/service"
)

// Config controls HTTP middleware behavior and response headers.
type Config struct {
	// Limit is the configured max requests in a window.
	Limit int
	// Window controls reset time reported via X-RateLimit-Reset.
	Window time.Duration
	// Algorithm labels rate-limit metrics with the active algorithm name.
	Algorithm string
}

// NewRateLimitMiddleware creates an HTTP middleware that:
//   - extracts client identity (JWT subject, API key, or IP)
//   - performs a rate-limit decision through RateLimitService
//   - rejects over-limit calls with HTTP 429
//   - emits standard rate-limit headers
func NewRateLimitMiddleware(svc *service.RateLimitService, cfg Config) func(http.Handler) http.Handler {
	if cfg.Limit <= 0 {
		cfg.Limit = 1
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identifier := clientIdentifier(r)
			key := fmt.Sprintf("%s|%s", identifier, r.URL.Path)

			allowed, remaining, err := svc.Allow(r.Context(), key)
			if err != nil {
				http.Error(w, "rate limit evaluation failed", http.StatusInternalServerError)
				return
			}

			setRateLimitHeaders(w, cfg.Limit, remaining, time.Now().Add(cfg.Window).Unix())

			if !allowed {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewMetricsMiddleware records Prometheus metrics for each request passing through the middleware chain.
func NewMetricsMiddleware(reg *metrics.Registry, algorithm string) func(http.Handler) http.Handler {
	if algorithm == "" {
		algorithm = "unknown"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecordingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			status := strconv.Itoa(rw.statusCode)
			blocked := rw.statusCode == http.StatusTooManyRequests
			reg.ObserveRateLimit(r.URL.Path, status, algorithm, blocked, time.Since(start))
		})
	}
}

// RateLimitMiddleware keeps compatibility with previous placeholder usage.
// For production wiring use NewRateLimitMiddleware.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return next
}

type statusRecordingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusRecordingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func setRateLimitHeaders(w http.ResponseWriter, limit, remaining int, resetUnix int64) {
	if remaining < 0 {
		remaining = 0
	}
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetUnix))
}

func clientIdentifier(r *http.Request) string {
	if sub := jwtSubject(r); sub != "" {
		return "sub:" + sub
	}
	if apiKey := strings.TrimSpace(r.Header.Get("X-API-Key")); apiKey != "" {
		return "api_key:" + apiKey
	}
	if ip := clientIP(r); ip != "" {
		return "ip:" + ip
	}
	return "ip:unknown"
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func jwtSubject(r *http.Request) string {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return ""
	}
	token := strings.TrimSpace(auth[7:])
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	sub, _ := claims["sub"].(string)
	return strings.TrimSpace(sub)
}

// compile-time guard: service must satisfy expected contract.
var _ interface {
	Allow(context.Context, string) (bool, int, error)
} = (*service.RateLimitService)(nil)
