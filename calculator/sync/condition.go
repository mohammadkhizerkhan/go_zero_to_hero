package sync

import (
	"sync"
)

// ReadySignal uses sync.Cond to coordinate multiple workers waiting for a shared condition.
type ReadySignal struct {
	mu    sync.Mutex
	cond  *sync.Cond
	ready bool
}

func NewReadySignal() *ReadySignal {
	r := &ReadySignal{}
	r.cond = sync.NewCond(&r.mu)
	return r
}

// Wait blocks the calling goroutine until SetReady is broadcast.
func (r *ReadySignal) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Always use a for loop to guard against spurious wakeups
	for !r.ready {
		r.cond.Wait()
	}
}

// SetReady marks the state as ready and wakes ALL waiting goroutines.
func (r *ReadySignal) SetReady() {
	r.mu.Lock()
	r.ready = true
	r.mu.Unlock()

	// Broadcast wakes all goroutines currently suspended in cond.Wait()
	r.cond.Broadcast()
}