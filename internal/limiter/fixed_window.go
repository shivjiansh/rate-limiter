package limiter

import "context"

type FixedWindow struct{}

func NewFixedWindow() *FixedWindow { return &FixedWindow{} }

func (f *FixedWindow) Allow(_ context.Context, _ string) (bool, int, error) {
	return true, 0, nil
}
