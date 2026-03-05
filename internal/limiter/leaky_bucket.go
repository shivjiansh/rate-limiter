package limiter

import "context"

type LeakyBucket struct{}

func NewLeakyBucket() *LeakyBucket { return &LeakyBucket{} }

func (l *LeakyBucket) Allow(_ context.Context, _ string) (bool, int, error) {
	return true, 0, nil
}
