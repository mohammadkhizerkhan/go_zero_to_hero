package test_context

import (
	"context"
	"errors"
	"testing"
	"time"
)

// go test -v ./context/...
func TestTimeoutWrapper_TimesOut(t *testing.T) {
	// Create a context with 2-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // Best practice: release timer resources

	start := time.Now()
	_, err := SlowOperation(ctx)
	elapsed := time.Since(start)

	// Check if error is context.DeadlineExceeded
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got: %v", err)
	}

	// Verify that it returned immediately around 2 seconds, NOT waiting for 5 seconds
	if elapsed >= 3*time.Second {
		t.Fatalf("operation took too long (%v), expected ~2s timeout", elapsed)
	}

	t.Logf("Successfully timed out in %v with error: %v", elapsed, err)
}
