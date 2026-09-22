package test_context

import (
	"context"
	"fmt"
	"testing"
)

// 1. Define custom types for context keys to prevent package collisions
type contextKey string

const (
	traceIDKey  contextKey = "trace_id"
	userNameKey contextKey = "user_name"
)

// Layer 3: Database/Repository Layer
func RepositoryLayer(ctx context.Context) (string, error) {
	// Extract values from context
	traceID, ok := ctx.Value(traceIDKey).(string)
	if !ok {
		return "", fmt.Errorf("trace ID missing from context")
	}

	userName, ok := ctx.Value(userNameKey).(string)
	if !ok {
		return "", fmt.Errorf("user name missing from context")
	}

	return fmt.Sprintf("Query executed by %s under TraceID: %s", userName, traceID), nil
}

// Layer 2: Business Logic Layer
func ServiceLayer(ctx context.Context) (string, error) {
	return RepositoryLayer(ctx)
}

// Layer 1: HTTP Controller Layer
func ControllerLayer(ctx context.Context) (string, error) {
	return ServiceLayer(ctx)
}

// go test -v -run TestValueTransport ./context/...
func TestValueTransport(t *testing.T) {
	// Start with a root context
	rootCtx := context.Background()

	// 2. Wrap/derive context with the first value (TraceID)
	ctxWithTrace := context.WithValue(rootCtx, traceIDKey, "tx-99201-abc")

	// 3. Wrap/derive with a new custom value (UserName)
	customCtx := context.WithValue(ctxWithTrace, userNameKey, "Khizer Khan")

	// Pass the single derived context containing both values down the call stack
	result, err := ControllerLayer(customCtx)
	if err != nil {
		t.Fatalf("Value transport failed: %v", err)
	}

	t.Log(result) // Prints: "Query executed by Khizer Khan under TraceID: tx-99201-abc"
}
