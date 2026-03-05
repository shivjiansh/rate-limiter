package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/example/go-rate-limiter/api/handlers"
	"github.com/example/go-rate-limiter/internal/config"
	"github.com/example/go-rate-limiter/internal/limiter"
	"github.com/example/go-rate-limiter/internal/logging"
	"github.com/example/go-rate-limiter/internal/metrics"
	"github.com/example/go-rate-limiter/internal/middleware"
	"github.com/example/go-rate-limiter/internal/redis"
	"github.com/example/go-rate-limiter/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()
	logger, err := logging.New()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	local, err := limiter.NewFromPolicy(limiter.Policy{Algorithm: cfg.Algorithm, Limit: cfg.Limit, Window: cfg.Window, ShardCount: 512, Burst: cfg.Limit})
	if err != nil {
		panic(err)
	}

	var redisLimiter *redis.DistributedLimiter
	if cfg.EnableRedis {
		client := redis.NewClient(redis.Config{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB, PoolSize: cfg.RedisPoolSize, MinIdleConns: cfg.RedisMinIdle})
		redisLimiter = redis.NewDistributedLimiter(client)
	}

	metricsRegistry := metrics.New()
	svc := service.NewRateLimiterService(local, redisLimiter, cfg.Limit, cfg.Window, metricsRegistry)

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", handlers.Health)
	r.GET("/v1/echo", middleware.RateLimit(svc, logger), handlers.Echo)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
