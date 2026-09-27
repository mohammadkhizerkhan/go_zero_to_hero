package benchmark_demo

import "testing"

const sliceSize = 10000

var globalSliceSink []int
// go test -bench=. -benchmem ./benchmark

// 1. Dynamic slice growth (triggers repeated doubling & GC copies)
func BenchmarkSliceDynamic(b *testing.B) {
	b.ReportAllocs()
	var res []int
	for i := 0; i < b.N; i++ {
		res = []int{}
		for j := 0; j < sliceSize; j++ {
			res = append(res, j)
		}
	}
	globalSliceSink = res
}

// 2. Pre-allocated capacity (1 heap allocation total)
func BenchmarkSlicePreAllocated(b *testing.B) {
	b.ReportAllocs()
	var res []int
	for i := 0; i < b.N; i++ {
		res = make([]int, 0, sliceSize)
		for j := 0; j < sliceSize; j++ {
			res = append(res, j)
		}
	}
	globalSliceSink = res
}
