package config

func Load() Config {
	return Config{Port: "8080", RedisAddr: "localhost:6379"}
}
