package benchmark_demo

import (
	"sync"
	"testing"
)

type SafeCounter struct {
	mu  sync.RWMutex
	val int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	c.val++
	c.mu.Unlock()
}

func (c *SafeCounter) Read() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}


//  go test -bench=BenchmarkParallelCounter -benchmem ./benchmark
func BenchmarkParallelCounter(b *testing.B) {
	counter := &SafeCounter{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc()
		}
	})
}
