package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port           string
	Limit          int
	Window         time.Duration
	Algorithm      string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	RedisPoolSize  int
	RedisMinIdle   int
	EnableRedis    bool
	TrustedProxies []string
}

func Load() Config {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("LIMIT", 1000)
	viper.SetDefault("WINDOW", "1s")
	viper.SetDefault("ALGORITHM", "fixed_window")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_POOL_SIZE", 256)
	viper.SetDefault("REDIS_MIN_IDLE", 64)
	viper.SetDefault("ENABLE_REDIS", false)
	viper.AutomaticEnv()
	w, _ := time.ParseDuration(viper.GetString("WINDOW"))
	return Config{
		Port:          viper.GetString("PORT"),
		Limit:         viper.GetInt("LIMIT"),
		Window:        w,
		Algorithm:     viper.GetString("ALGORITHM"),
		RedisAddr:     viper.GetString("REDIS_ADDR"),
		RedisPassword: viper.GetString("REDIS_PASSWORD"),
		RedisDB:       viper.GetInt("REDIS_DB"),
		RedisPoolSize: viper.GetInt("REDIS_POOL_SIZE"),
		RedisMinIdle:  viper.GetInt("REDIS_MIN_IDLE"),
		EnableRedis:   viper.GetBool("ENABLE_REDIS"),
	}
}
