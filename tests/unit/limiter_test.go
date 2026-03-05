package unit

import (
	"context"
	"testing"
	"time"

	"github.com/goshield/rate-limiter/internal/limiter"
)

func TestTokenBucketBurstAndRefill(t *testing.T) {
	tb, err := limiter.NewTokenBucketWithConfig(2, 10, 8)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	if ok, _, err := tb.Allow(context.Background(), "user-1"); err != nil || !ok {
		t.Fatalf("expected request 1 allowed, ok=%v err=%v", ok, err)
	}
	if ok, _, err := tb.Allow(context.Background(), "user-1"); err != nil || !ok {
		t.Fatalf("expected request 2 allowed, ok=%v err=%v", ok, err)
	}
	if ok, _, err := tb.Allow(context.Background(), "user-1"); err != nil || ok {
		t.Fatalf("expected request 3 blocked, ok=%v err=%v", ok, err)
	}

	time.Sleep(120 * time.Millisecond)
	if ok, _, err := tb.Allow(context.Background(), "user-1"); err != nil || !ok {
		t.Fatalf("expected request allowed after refill, ok=%v err=%v", ok, err)
	}
}

func TestTokenBucketPerKeyIsolation(t *testing.T) {
	tb, err := limiter.NewTokenBucketWithConfig(1, 1, 4)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	okA, _, _ := tb.Allow(context.Background(), "a")
	okB, _, _ := tb.Allow(context.Background(), "b")
	if !okA || !okB {
		t.Fatalf("expected separate keys to have independent buckets")
	}
}
