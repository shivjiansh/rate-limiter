# GoShield — Distributed Rate Limiter in Go

GoShield is a production-grade, distributed rate limiting service for high-scale microservice environments.

## Highlights
- Five pluggable algorithms: Fixed Window, Token Bucket, Sliding Window Log, Sliding Window Counter, and Leaky Bucket.
- Concurrency-safe sharded in-memory implementations to reduce lock contention.
- Redis distributed enforcement using atomic Lua scripts.
- Gin middleware + gRPC unary interceptor integrations.
- Prometheus metrics and Zap structured logging.
- YAML-based config with environment variable overrides.

## Limiter Contract
```go
type Limiter interface {
    Allow(ctx context.Context, key string) (bool, int, error)
}
```

## Project Structure
```text
cmd/
  server/
  cli/
internal/
  limiter/
  redis/
  middleware/
  service/
  config/
  metrics/
  logging/
  utils/
api/
  proto/
  handlers/
deploy/
  docker/
  kubernetes/
tests/
  unit/
  integration/
  benchmark/
examples/
docs/
```

## Run
```bash
go run ./cmd/server
```

## Benchmark
```bash
go test -bench=. ./tests/benchmark/...
./tests/benchmark/vegeta.sh
```

## Configuration
- Base YAML: `deploy/config.yaml`
- Env override prefix: `GOSHIELD_` (e.g., `GOSHIELD_LIMIT=5000`)

See `docs/configuration.md` and `docs/architecture.md` for details.
