# Topic I: Context (Lifecycle & Scope Management)

In Go, `context.Context` serves two distinct purposes across API boundaries and goroutine trees:

1. **Cancellation Propagation**: Signaling to downstream operations that work should be abandoned (timeouts, user cancellations, disconnects).
2. **Request Scoping**: Transporting request-scoped metadata (Trace IDs, Auth Tokens, tenant details) down a execution call stack.

---

## Under the Hood: How Context Works

A Go `Context` is an **immutable node in a tree structure**.

```text
               [ context.Background() ]  (Root)
                          │
             [ context.WithValue(TraceID) ]
                          │
             [ context.WithCancel() ]  (Parent)
                ┌─────────┴─────────┐
                ▼                   ▼
     [ context.WithTimeout(2s) ]  [ Child 2 ]

```

### 1. The Context Interface

The standard library defines `context.Context` as a 4-method interface:

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}

```

- **`Done() <-chan struct{}`**: Returns a channel that is **closed** when the context is canceled or times out. Receiving from a closed channel unblocks instantly.
- **`Err() error`**: Returns `nil` while the context is active. After cancellation, it returns `context.Canceled` or `context.DeadlineExceeded`.
- **`Deadline()`**: Returns the point in time when this context will automatically be canceled.
- **`Value(key)`**: Traverses **upward** toward the root node searching for a matching key.

### 2. Golden Rules of Context

- **Pass Context as the First Parameter**: `func DoWork(ctx context.Context, arg string) error`.
- **Never Store Contexts in Structs**: Always pass them explicitly through the function call chain.
- **Never Pass `nil**`: Use `context.Background()`for root entry points or`context.TODO()` when refactoring legacy code.
- **Request-Scoped Data Only**: Use `context.WithValue` _only_ for operational metadata (telemetry, trace IDs, security tokens), never for optional functional arguments.

---

## Assignments & Implementation Guide

### I.1 Timeout Wrapper

**Goal**: Run a slow operation inside a goroutine. If the operation exceeds the context timeout, return an error immediately without hanging.

#### Mechanics

You wrap an unbuffered channel completion signal inside a `select` statement that listens on `<-ctx.Done()`.

```go
package context_demo

import (
	"context"
	"errors"
	"testing"
	"time"
)

func SlowOperation(ctx context.Context) (string, error) {
	// Create a channel to catch the background result
	resultCh := make(chan string, 1)

	go func() {
		// Simulate a 5-second heavy operation
		time.Sleep(5 * time.Second)
		resultCh <- "Work Complete!"
	}()

	select {
	case res := <-resultCh:
		return res, nil
	case <-ctx.Done():
		// Context timed out or was canceled
		return "", ctx.Err() // Returns context.DeadlineExceeded
	}
}

func TestTimeoutWrapper(t *testing.T) {
	// Set a 2-second timeout on the context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // Always call cancel to release timer resources!

	_, err := SlowOperation(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected DeadlineExceeded error, got: %v", err)
	}
	t.Log("Successfully aborted slow operation after 2s timeout!")
}

```

---

### I.2 HTTP Request with Context

**Goal**: Cancel an outgoing HTTP request if the remote server takes too long to respond.

#### Mechanics

Standard library constructs like `http.Client`, `database/sql`, and `os/exec` natively listen to `ctx.Done()`. When the context expires, `http.Client` instantly terminates the underlying TCP socket connection.

```go
package context_demo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPRequestWithContext(t *testing.T) {
	// 1. Create a slow test HTTP server (delays response by 3 seconds)
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	// 2. Create a context with a strict 1-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// 3. Bind the context to the HTTP Request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, slowServer.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// 4. Execute the HTTP Request
	client := &http.Client{}
	_, err = client.Do(req)

	if err == nil {
		t.Fatal("Expected HTTP request to fail due to timeout, but it succeeded.")
	}
	t.Logf("HTTP Request canceled successfully as expected: %v", err)
}

```

---

### I.3 Tree Cancel (Cancellation Propagation)

**Goal**: Verify that canceling a parent context automatically cancels all child contexts derived from it, while canceling a child leaves the parent active.

```go
package context_demo

import (
	"context"
	"errors"
	"testing"
)

func TestTreeCancel(t *testing.T) {
	// Parent Context
	parentCtx, parentCancel := context.WithCancel(context.Background())

	// Child Context derived from Parent
	childCtx, _ := context.WithCancel(parentCtx)

	// Grandchild Context derived from Child
	grandchildCtx, _ := context.WithCancel(childCtx)

	// Before cancellation, all contexts should be active (Err() == nil)
	if parentCtx.Err() != nil || childCtx.Err() != nil || grandchildCtx.Err() != nil {
		t.Fatal("Contexts should not be canceled yet")
	}

	// Cancel the PARENT context
	parentCancel()

	// Parent cancellation propagates DOWN to Child and Grandchild instantly
	if !errors.Is(childCtx.Err(), context.Canceled) {
		t.Errorf("Expected child context to be canceled, got %v", childCtx.Err())
	}
	if !errors.Is(grandchildCtx.Err(), context.Canceled) {
		t.Errorf("Expected grandchild context to be canceled, got %v", grandchildCtx.Err())
	}

	t.Log("Tree Cancellation Verified: Parent cancel propagated down to all descendants!")
}

```

---

### I.4 Value Transport (Request Scoping)

**Goal**: Pass a request metadata item (`TraceID`) through 3 layered function calls (`Controller` $\rightarrow$ `Service` $\rightarrow$ `Repository`) using `context.WithValue`.

#### Type Collision Prevention

_Never_ use plain strings as keys in `context.WithValue("key", "val")`. If two third-party packages use `"trace_id"`, they will overwrite each other's data. Always define a unexported custom key type:

```go
type contextKey string
const traceIDKey contextKey = "trace_id"

```

#### Code Implementation

```go
package context_demo

import (
	"context"
	"fmt"
	"testing"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

// Layer 3: Database/Repository Layer
func RepositoryLayer(ctx context.Context) (string, error) {
	// Extract value from context
	traceID, ok := ctx.Value(traceIDKey).(string)
	if !ok {
		return "", fmt.Errorf("trace ID missing from context")
	}
	return fmt.Sprintf("Query executed under TraceID: %s", traceID), nil
}

// Layer 2: Business Logic Layer
func ServiceLayer(ctx context.Context) (string, error) {
	return RepositoryLayer(ctx)
}

// Layer 1: HTTP Controller Layer
func ControllerLayer(ctx context.Context) (string, error) {
	return ServiceLayer(ctx)
}

func TestValueTransport(t *testing.T) {
	// Inject TraceID into Root Context
	ctx := context.WithValue(context.Background(), traceIDKey, "tx-99201-abc")

	result, err := ControllerLayer(ctx)
	if err != nil {
		t.Fatalf("Value transport failed: %v", err)
	}

	t.Log(result) // Prints: "Query executed under TraceID: tx-99201-abc"
}

```

---

### I.5 DB Loop (Early Abortion)

**Goal**: Run a high-iteration loop doing batch work. Check `ctx.Err()` at the top of every iteration to abort work early if the caller cancels the operation.

#### Mechanics

Instead of completing all 1,000 iterations when a user disconnects or an API request times out, check `ctx.Err()` on every iteration.

```go
package context_demo

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

```

---

## Hands-On Exercise

Try running these context tests locally:

```bash
go test -v -run "TestTimeoutWrapper|TestHTTPRequestWithContext|TestTreeCancel|TestValueTransport|TestDBLoopEarlyAbort"

```

Once you are comfortable with how contexts propagate cancellations down goroutine trees and transport metadata across boundaries, we can move directly into **Module 3: Engineering Rigor — Topic J: Unit Testing & Mocking**!
