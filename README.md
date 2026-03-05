# Go Distributed Rate Limiter

Production-oriented distributed rate limiter service in Go with in-memory and Redis-backed enforcement.

## Features
- Algorithms: Token Bucket, Sliding Window Log, Sliding Window Counter, Fixed Window, Leaky Bucket
- Scopes: per IP, per user, per API key, per endpoint, global
- Gin HTTP middleware and gRPC unary interceptor
- Prometheus metrics (`rate_limiter_requests_total`, `rate_limiter_blocked_total`, `rate_limiter_latency`)
- Structured logging (Zap) with `request_id` and `user_id`
- Redis atomic Lua scripts for distributed decisions
- Docker, Kubernetes, Terraform examples

## Architecture
```text
Client Request
  -> HTTP Middleware
  -> Key Extractor
  -> Rate Limit Service
  -> Local Sharded Limiter
  -> Redis Distributed Limiter (fallback/global)
```

See [docs/architecture.md](docs/architecture.md).

## Run locally
```bash
go run ./cmd/server
```

## Docker compose
```bash
docker compose -f deploy/docker/docker-compose.yml up --build
```

## Benchmarks
```bash
go test -bench=. ./tests/benchmark/...
./tests/benchmark/vegeta.sh
```
