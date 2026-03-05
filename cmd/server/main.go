package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/goshield/rate-limiter/internal/limiter"
	"github.com/goshield/rate-limiter/internal/metrics"
	"github.com/goshield/rate-limiter/internal/middleware"
	"github.com/goshield/rate-limiter/internal/service"
)

func main() {
	localLimiter := limiter.NewTokenBucket()
	rateLimitService := service.NewRateLimitService(localLimiter)
	metricsRegistry := metrics.New()

	rateLimitMW := middleware.NewRateLimitMiddleware(rateLimitService, middleware.Config{
		Limit:     100,
		Window:    time.Minute,
		Algorithm: "token_bucket",
	})
	metricsMW := middleware.NewMetricsMiddleware(metricsRegistry, "token_bucket")

	mux := http.NewServeMux()
	mux.Handle("/metrics", metricsRegistry.Handler())
	mux.Handle("/health", metricsMW(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})))

	protectedHandler := metricsMW(rateLimitMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(fmt.Sprintf("request accepted: %s", r.URL.Path)))
	})))
	mux.Handle("/", protectedHandler)

	addr := ":8080"
	log.Printf("GoShield server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
