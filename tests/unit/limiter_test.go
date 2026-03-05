package unit

import (
	"testing"
	"time"

	"github.com/example/go-rate-limiter/internal/limiter"
)

func TestFixedWindowLimiter(t *testing.T) {
	l := limiter.NewFixedWindowLimiter(2, time.Second, 4)
	if ok, _ := l.Allow("k"); !ok {
		t.Fatal("first request should pass")
	}
	if ok, _ := l.Allow("k"); !ok {
		t.Fatal("second request should pass")
	}
	if ok, _ := l.Allow("k"); ok {
		t.Fatal("third request should be blocked")
	}
}

func TestTokenBucketLimiterRefill(t *testing.T) {
	l := limiter.NewTokenBucketLimiter(1, 10, 4)
	if ok, _ := l.Allow("k"); !ok {
		t.Fatal("first request should pass")
	}
	if ok, _ := l.Allow("k"); ok {
		t.Fatal("second request should be blocked")
	}
	time.Sleep(110 * time.Millisecond)
	if ok, _ := l.Allow("k"); !ok {
		t.Fatal("request should pass after refill")
	}
}
