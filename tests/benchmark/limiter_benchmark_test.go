package benchmark

import (
	"strconv"
	"testing"
	"time"

	"github.com/example/go-rate-limiter/internal/limiter"
)

func BenchmarkFixedWindowAllow(b *testing.B) {
	l := limiter.NewFixedWindowLimiter(1000000, time.Second, 1024)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			l.Allow("k" + strconv.Itoa(i%10000))
			i++
		}
	})
}
