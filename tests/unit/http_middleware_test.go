package unit

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goshield/rate-limiter/internal/limiter"
	"github.com/goshield/rate-limiter/internal/middleware"
	"github.com/goshield/rate-limiter/internal/service"
)

type stubLimiter struct {
	allowByKey map[string]bool
}

func (s *stubLimiter) Allow(_ context.Context, key string) (bool, int, error) {
	allowed, ok := s.allowByKey[key]
	if !ok {
		return true, 9, nil
	}
	if allowed {
		return true, 1, nil
	}
	return false, 0, nil
}

var _ limiter.Limiter = (*stubLimiter)(nil)

func TestHTTPMiddlewareRejectsWith429AndHeaders(t *testing.T) {
	svc := service.NewRateLimitService(&stubLimiter{allowByKey: map[string]bool{"api_key:k1|/v1/resource": false}})
	mw := middleware.NewRateLimitMiddleware(svc, middleware.Config{Limit: 10, Window: time.Second})

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/resource", nil)
	req.Header.Set("X-API-Key", "k1")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
	if rr.Header().Get("X-RateLimit-Limit") != "10" {
		t.Fatalf("missing/invalid X-RateLimit-Limit")
	}
	if rr.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("missing/invalid X-RateLimit-Remaining")
	}
	if rr.Header().Get("X-RateLimit-Reset") == "" {
		t.Fatalf("missing X-RateLimit-Reset")
	}
}

func TestHTTPMiddlewareUsesJWTSubjectIdentifier(t *testing.T) {
	sub := "user-42"
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"sub":"%s"}`, sub)))
	token := strings.Join([]string{"header", payload, "sig"}, ".")

	svc := service.NewRateLimitService(&stubLimiter{allowByKey: map[string]bool{"sub:user-42|/v1/me": true}})
	mw := middleware.NewRateLimitMiddleware(svc, middleware.Config{Limit: 5, Window: time.Second})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-API-Key", "ignored")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
