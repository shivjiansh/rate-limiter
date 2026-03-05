package limiter

import "context"

type SlidingWindow struct{}

func NewSlidingWindow() *SlidingWindow { return &SlidingWindow{} }

func (s *SlidingWindow) Allow(_ context.Context, _ string) (bool, int, error) {
	return true, 0, nil
}
