package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Config contains Redis connection and pool tuning.
type Config struct {
	Addr            string
	Password        string
	DB              int
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	ConnMaxIdleTime time.Duration
}

func defaultConfig(cfg Config) Config {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:6379"
	}
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 128
	}
	if cfg.MinIdleConns < 0 {
		cfg.MinIdleConns = 0
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 2
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 2 * time.Second
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 500 * time.Millisecond
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 500 * time.Millisecond
	}
	if cfg.PoolTimeout <= 0 {
		cfg.PoolTimeout = 1 * time.Second
	}
	if cfg.ConnMaxIdleTime <= 0 {
		cfg.ConnMaxIdleTime = 5 * time.Minute
	}
	return cfg
}

// NewClient creates a go-redis client with connection pooling and built-in retries.
func NewClient(cfg Config) *goredis.Client {
	cfg = defaultConfig(cfg)
	return goredis.NewClient(&goredis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	})
}

// Ping validates connectivity.
func Ping(ctx context.Context, client *goredis.Client) error {
	if client == nil {
		return fmt.Errorf("nil redis client")
	}
	return client.Ping(ctx).Err()
}
