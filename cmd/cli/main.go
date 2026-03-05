package main

import (
	"fmt"
	"time"

	"github.com/example/go-rate-limiter/internal/limiter"
)

func main() {
	l := limiter.NewFixedWindowLimiter(5, time.Second, 16)
	for i := 0; i < 10; i++ {
		ok, rem := l.Allow("demo-user")
		fmt.Printf("request=%d allowed=%v remaining=%d\n", i+1, ok, rem)
	}
}
