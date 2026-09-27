package benchmark_demo

import (
	"fmt"
	"testing"
)

// go test -bench=. -benchmem ./benchmark

func linearSearch(data []int, target int) bool {
	for _, v := range data {
		if v == target {
			return true
		}
	}
	return false
}

func BenchmarkLinearSearch_Table(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}
			target := size - 1 // Worst-case search at end of slice

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = linearSearch(data, target)
			}
		})
	}
}
