package sync

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

// TestCacheConcurrency verifies concurrency safety under multiple readers and writers.
func TestCacheConcurrency(t *testing.T) {
	cache := NewCache()
	cache.Set("env", "production")

	var wg sync.WaitGroup

	// 100 Concurrent Readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, ok := cache.Get("env")
			if !ok && val == "" {
				t.Errorf("expected valid value, got empty")
			}
		}()
	}

	// 1 Writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.Set("env", "staging")
	}()

	wg.Wait()
}

// -------------------------------------------------------------
// BENCHMARKS: Read-Heavy Concurrency (RLock vs Lock vs No Lock)
// -------------------------------------------------------------

// Benchmark 1: RWMutex RLock under parallel concurrent readers
func BenchmarkCache_WithRLock_Parallel(b *testing.B) {
	cache := NewCache()
	cache.Set("key", "value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.Get("key")
		}
	})
}

// Benchmark 2: Standard Mutex Lock under parallel concurrent readers
func BenchmarkCache_WithMutex_Parallel(b *testing.B) {
	cache := NewMutexCache()
	cache.Set("key", "value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.Get("key")
		}
	})
}

// Benchmark 3: Unsynchronized reads (Baseline single goroutine, zero locking overhead)
func BenchmarkCache_WithoutLock_Single(b *testing.B) {
	cache := NewNoLockCache()
	cache.Set("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("key")
	}
}

// -------------------------------------------------------------
// BENCHMARKS: Mixed Workload (90% Reads, 10% Writes)
// -------------------------------------------------------------

func BenchmarkMixed_RWMutex(b *testing.B) {
	cache := NewCache()
	cache.Set("k0", "val")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(42))
		for pb.Next() {
			if r.Intn(10) == 0 { // 10% writes
				cache.Set("k"+strconv.Itoa(r.Intn(100)), "new_val")
			} else { // 90% reads
				_, _ = cache.Get("k" + strconv.Itoa(r.Intn(100)))
			}
		}
	})
}

func BenchmarkMixed_Mutex(b *testing.B) {
	cache := NewMutexCache()
	cache.Set("k0", "val")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(42))
		for pb.Next() {
			if r.Intn(10) == 0 { // 10% writes
				cache.Set("k"+strconv.Itoa(r.Intn(100)), "new_val")
			} else { // 90% reads
				_, _ = cache.Get("k" + strconv.Itoa(r.Intn(100)))
			}
		}
	})
}
