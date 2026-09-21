package sync

import (
	"sync"
	"sync/atomic"
)

//  go test -bench="BenchmarkCounter" -benchmem ./sync

// MutexCounter uses a sync.Mutex to protect an integer counter.
type MutexCounter struct {
	mu    sync.Mutex
	count int64
}

func (c *MutexCounter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *MutexCounter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// AtomicCounter uses hardware-level atomic operations.
type AtomicCounter struct {
	count int64
}

func (c *AtomicCounter) Inc() {
	atomic.AddInt64(&c.count, 1)
}

func (c *AtomicCounter) Value() int64 {
	return atomic.LoadInt64(&c.count)
}
