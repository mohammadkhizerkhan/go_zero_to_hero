package sync

import (
	"sync"
	"testing"
	"time"
)

func TestCondBroadcast(t *testing.T) {
	readySignal := NewReadySignal()
	var wg sync.WaitGroup

	numWorkers := 5

	// Start 5 workers that will wait for the broadcast signal
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			t.Logf("Worker %d: Waiting for broadcast signal...", workerID)
			readySignal.Wait()
			t.Logf("Worker %d: Woke up! Starting work.", workerID)
		}(i)
	}

	// Simulate preparation / initialization work
	time.Sleep(100 * time.Millisecond)
	t.Log("Coordinator: Initialization complete. Broadcasting to all workers!")

	// Wake up all 5 waiting goroutines simultaneously
	readySignal.SetReady()

	wg.Wait()
	t.Log("All workers finished successfully.")
}
