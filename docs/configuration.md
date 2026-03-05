# Configuration

GoShield reads `config.yaml` (search paths: repository root and `deploy/`) and then applies environment variable overrides.

## YAML keys
- `port`
- `limit`
- `window`
- `algorithm`
- `burst`
- `enable_redis`
- `redis_addr`
- `redis_password`
- `redis_db`
- `redis_pool_size`
- `redis_min_idle`
- `trusted_proxies`

## Environment overrides
Prefix all keys with `GOSHIELD_`.

Examples:
- `GOSHIELD_PORT=8081`
- `GOSHIELD_LIMIT=5000`
- `GOSHIELD_ALGORITHM=token_bucket`
- `GOSHIELD_ENABLE_REDIS=true`
