package benchmark

import "testing"

func BenchmarkSkeleton(b *testing.B) {
	for i := 0; i < b.N; i++ {
	}
}
