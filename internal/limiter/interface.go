package limiter

// Limiter is implemented by all rate limiter algorithms.
// The returned int is the remaining quota for current decision context.
type Limiter interface {
	Allow(key string) (bool, int)
}
