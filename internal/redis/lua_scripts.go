package redis

// SlidingWindowScript enforces a rolling window limit using a Redis sorted set.
//
// Responsibilities:
// 1) remove expired entries from the sorted set
// 2) count current requests in the window
// 3) reject if limit is already reached
// 4) add the new request atomically
//
// ARGV:
//   1: now in milliseconds
//   2: window size in milliseconds
//   3: max requests allowed in window
//   4: unique member id for current request
const SlidingWindowScript = `
local key = KEYS[1]
local now_ms = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local max_requests = tonumber(ARGV[3])
local member = ARGV[4]

local window_start = now_ms - window_ms

-- 1) remove expired entries
redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

-- 2) count existing requests
local current = redis.call('ZCARD', key)

-- 3) reject if limit exceeded
if current >= max_requests then
  return {0, max_requests - current}
end

-- 4) add new request
redis.call('ZADD', key, now_ms, member)

-- keep key bounded in memory
redis.call('PEXPIRE', key, window_ms)

return {1, max_requests - current - 1}
`
