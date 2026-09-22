package test_context

import (
	"context"
	"testing"
	"time"
)

func ProcessBatchItems(ctx context.Context, items []int) (int, error) {
	processedCount := 0

	for _, item := range items {
		// Polling for cancellation at the start of each iteration
		select {
		case <-ctx.Done():
			// Abort processing early and return current status
			return processedCount, ctx.Err()
		default:
			// No cancellation signal; proceed with work
		}

		// Simulate 10ms of work per item
		time.Sleep(10 * time.Millisecond)
		_ = item * 2
		processedCount++
	}

	return processedCount, nil
}

//  go test -v -run TestDBLoopEarlyAbort ./context/...
func TestDBLoopEarlyAbort(t *testing.T) {
	// Set up 100 items (would normally take 1000ms / 1 second)
	items := make([]int, 100)

	// Set context to time out after 150ms
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	processed, err := ProcessBatchItems(ctx, items)

	if processed >= 100 {
		t.Fatalf("Expected loop to abort early, but processed all %d items", processed)
	}

	t.Logf("Loop successfully aborted early! Processed %d/100 items before timeout error: %v", processed, err)
}

