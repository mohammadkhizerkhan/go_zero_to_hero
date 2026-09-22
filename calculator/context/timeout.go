package test_context

import (
	"context"
	"time"
)

// SlowOperation simulates a 5-second heavy task.
// If the provided context times out or is canceled before 5s,
// it immediately returns ctx.Err() (e.g. context.DeadlineExceeded).
func SlowOperation(ctx context.Context) (string, error) {
	// Buffered channel of size 1 ensures the goroutine won't leak
	// if it finishes after the function has already returned on timeout.
	resultChan := make(chan string, 1)

	go func() {
		time.Sleep(5 * time.Second)
		resultChan <- "Task completed successfully"
	}()

	select {
	case res := <-resultChan:
		return res, nil
	case <-ctx.Done():
		// Returns context.DeadlineExceeded (or context.Canceled) immediately
		return "", ctx.Err()
	}
}