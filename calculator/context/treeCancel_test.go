package test_context

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

	if errors.Is(childCtx.Err(),context.Canceled){
		t.Errorf("Child context is cancelled")
	}

	// Parent cancellation propagates DOWN to Child and Grandchild instantly
	if !errors.Is(childCtx.Err(), context.Canceled) {
		t.Errorf("Expected child context to be canceled, got %v", childCtx.Err())
	}
	if !errors.Is(grandchildCtx.Err(), context.Canceled) {
		t.Errorf("Expected grandchild context to be canceled, got %v", grandchildCtx.Err())
	}

	t.Log("Tree Cancellation Verified: Parent cancel propagated down to all descendants!")
}

