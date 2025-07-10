// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package pyasyncio

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAsyncExitStack_Basic(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	// Test basic enter/exit
	result, err := stack.AEnter(ctx)
	if err != nil {
		t.Fatalf("AEnter failed: %v", err)
	}
	if result != stack {
		t.Error("AEnter should return self")
	}

	suppressed, err := stack.AExit(ctx, nil)
	if err != nil {
		t.Fatalf("AExit failed: %v", err)
	}
	if suppressed {
		t.Error("AExit should not suppress when no error")
	}
}

func TestAsyncExitStack_AlreadyEntered(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	_, _ = stack.AEnter(ctx)
	_, err := stack.AEnter(ctx)
	if err == nil {
		t.Error("Second AEnter should fail")
	}
}

func TestAsyncExitStack_EnterAsyncContext(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	var entered bool
	var exited bool
	var exitErr error

	cm := NewAsyncContextManager(
		func(ctx context.Context) (any, error) {
			entered = true
			return "test", nil
		},
		func(ctx context.Context, exc error) (bool, error) {
			exited = true
			exitErr = exc
			return false, nil
		},
	)

	result, err := stack.EnterAsyncContext(ctx, cm)
	if err != nil {
		t.Fatalf("EnterAsyncContext failed: %v", err)
	}
	if result != "test" {
		t.Errorf("Expected result 'test', got %v", result)
	}
	if !entered {
		t.Error("Context manager was not entered")
	}

	stack.AClose(ctx)
	if !exited {
		t.Error("Context manager was not exited")
	}
	if exitErr != nil {
		t.Errorf("Exit received unexpected error: %v", exitErr)
	}
}

func TestAsyncExitStack_EnterContext(t *testing.T) {
	ctx := context.Background()
	stack := NewAsyncExitStack()

	var entered bool
	var exited bool
	var exitErr error

	cm := NewContextManager(
		func() (any, error) {
			entered = true
			return "sync", nil
		},
		func(exc error) (bool, error) {
			exited = true
			exitErr = exc
			return false, nil
		},
	)

	result, err := stack.EnterContext(cm)
	if err != nil {
		t.Fatalf("EnterContext failed: %v", err)
	}
	if result != "sync" {
		t.Errorf("Expected result 'sync', got %v", result)
	}
	if !entered {
		t.Error("Context manager was not entered")
	}

	stack.AClose(ctx)
	if !exited {
		t.Error("Context manager was not exited")
	}
	if exitErr != nil {
		t.Errorf("Exit received unexpected error: %v", exitErr)
	}
}

func TestAsyncExitStack_PushAsyncExit(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	var called bool
	var receivedErr error

	exitFunc := func(ctx context.Context, exc error) (bool, error) {
		called = true
		receivedErr = exc
		return false, nil
	}

	returnedFunc := stack.PushAsyncExit(exitFunc)
	// Verify function was returned (can't compare functions directly in Go)
	if returnedFunc == nil {
		t.Error("PushAsyncExit should return the function")
	}

	testErr := errors.New("test error")
	stack.unwindStack(ctx, testErr)

	if !called {
		t.Error("Exit function was not called")
	}
	if receivedErr != testErr {
		t.Errorf("Exit function received wrong error: %v", receivedErr)
	}
}

func TestAsyncExitStack_Push(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	var called bool
	var receivedErr error

	exitFunc := func(exc error) (bool, error) {
		called = true
		receivedErr = exc
		return false, nil
	}

	returnedFunc := stack.Push(exitFunc)
	// Verify function was returned (can't compare functions directly in Go)
	if returnedFunc == nil {
		t.Error("Push should return the function")
	}

	testErr := errors.New("test error")
	stack.unwindStack(ctx, testErr)

	if !called {
		t.Error("Exit function was not called")
	}
	if receivedErr != testErr {
		t.Errorf("Exit function received wrong error: %v", receivedErr)
	}
}

func TestAsyncExitStack_Callbacks(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	var asyncCalled bool
	var syncCalled bool

	asyncCallback := func(ctx context.Context) error {
		asyncCalled = true
		return nil
	}

	syncCallback := func() error {
		syncCalled = true
		return nil
	}

	returnedAsync := stack.PushAsyncCallback(asyncCallback)
	// Verify function was returned (can't compare functions directly in Go)
	if returnedAsync == nil {
		t.Error("PushAsyncCallback should return the function")
	}

	returnedSync := stack.Callback(syncCallback)
	// Verify function was returned (can't compare functions directly in Go)
	if returnedSync == nil {
		t.Error("Callback should return the function")
	}

	stack.AClose(ctx)

	if !asyncCalled {
		t.Error("Async callback was not called")
	}
	if !syncCalled {
		t.Error("Sync callback was not called")
	}
}

func TestAsyncExitStack_CallbacksWithArgs(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	var asyncReceived []any
	var syncReceived []any

	asyncCallback := func(ctx context.Context, args ...any) error {
		asyncReceived = args
		return nil
	}

	syncCallback := func(args ...any) error {
		syncReceived = args
		return nil
	}

	stack.PushAsyncCallbackArgs(asyncCallback, "async", 42, true)
	stack.CallbackArgs(syncCallback, "sync", 24, false)

	stack.AClose(ctx)

	expectedAsync := []any{"async", 42, true}
	if diff := cmp.Diff(expectedAsync, asyncReceived); diff != "" {
		t.Errorf("Async callback args mismatch (-want +got):\n%s", diff)
	}

	expectedSync := []any{"sync", 24, false}
	if diff := cmp.Diff(expectedSync, syncReceived); diff != "" {
		t.Errorf("Sync callback args mismatch (-want +got):\n%s", diff)
	}
}

func TestAsyncExitStack_ExceptionSuppression(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	// Add an exit handler that suppresses errors
	stack.PushAsyncExit(func(ctx context.Context, exc error) (bool, error) {
		if exc != nil && exc.Error() == "suppress me" {
			return true, nil // Suppress the error
		}
		return false, nil
	})

	// Test suppression
	testErr := errors.New("suppress me")
	suppressed, err := stack.unwindStack(ctx, testErr)

	if !suppressed {
		t.Error("Error should have been suppressed")
	}
	if err != nil {
		t.Errorf("No error should be returned when suppressed: %v", err)
	}
}

func TestAsyncExitStack_CallbackOrder(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	var order []int

	// Add callbacks in order 1, 2, 3
	stack.Callback(func() error {
		order = append(order, 1)
		return nil
	})
	stack.PushAsyncCallback(func(ctx context.Context) error {
		order = append(order, 2)
		return nil
	})
	stack.Push(func(exc error) (bool, error) {
		order = append(order, 3)
		return false, nil
	})

	stack.AClose(ctx)

	// Should be called in reverse order: 3, 2, 1
	expected := []int{3, 2, 1}
	if diff := cmp.Diff(expected, order); diff != "" {
		t.Errorf("Callback order mismatch (-want +got):\n%s", diff)
	}
}

func TestAsyncExitStack_PopAll(t *testing.T) {
	ctx := t.Context()
	stack1 := NewAsyncExitStack()

	var called1, called2 bool

	stack1.Callback(func() error {
		called1 = true
		return nil
	})
	stack1.Callback(func() error {
		called2 = true
		return nil
	})

	// Transfer callbacks to new stack
	stack2 := stack1.PopAll()

	// Close original stack - nothing should happen
	stack1.AClose(ctx)
	if called1 || called2 {
		t.Error("Callbacks should not be called on original stack after PopAll")
	}

	// Close new stack - callbacks should be called
	stack2.AClose(ctx)
	if !called1 || !called2 {
		t.Error("Callbacks should be called on new stack")
	}
}

func TestAsyncExitStack_ErrorPropagation(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	expectedErr := errors.New("callback error")

	// Add a callback that returns an error
	stack.Callback(func() error {
		return expectedErr
	})

	// Add another callback that should still be called
	var secondCalled bool
	stack.Callback(func() error {
		secondCalled = true
		return nil
	})

	err := stack.AClose(ctx)

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if !secondCalled {
		t.Error("Second callback should still be called despite first error")
	}
}

func TestAsyncExitStack_ClosedStack(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	// Close the stack
	stack.AClose(ctx)

	// Try to use closed stack
	var called bool
	callback := func() error {
		called = true
		return nil
	}

	returned := stack.Callback(callback)
	// We can't directly compare functions in Go, but we can verify it returns something
	if returned == nil {
		t.Error("Callback should still return the function even when closed")
	}

	// Verify callback wasn't added
	stack.AClose(ctx) // Close again
	if called {
		t.Error("Callback should not be called when added to closed stack")
	}
}

func TestAsyncExitStack_ComplexScenario(t *testing.T) {
	ctx := t.Context()
	stack := NewAsyncExitStack()

	var events []string

	// Enter multiple context managers
	cm1 := NewAsyncContextManager(
		func(ctx context.Context) (any, error) {
			events = append(events, "enter-async-1")
			return "async1", nil
		},
		func(ctx context.Context, exc error) (bool, error) {
			events = append(events, "exit-async-1")
			return false, nil
		},
	)

	cm2 := NewContextManager(
		func() (any, error) {
			events = append(events, "enter-sync-2")
			return "sync2", nil
		},
		func(exc error) (bool, error) {
			events = append(events, "exit-sync-2")
			return false, nil
		},
	)

	stack.EnterAsyncContext(ctx, cm1)
	stack.EnterContext(cm2)

	// Add callbacks
	stack.PushAsyncCallback(func(ctx context.Context) error {
		events = append(events, "async-callback")
		return nil
	})

	stack.Callback(func() error {
		events = append(events, "sync-callback")
		return nil
	})

	// Close and verify order
	stack.AClose(ctx)

	expected := []string{
		"enter-async-1",
		"enter-sync-2",
		"sync-callback",
		"async-callback",
		"exit-sync-2",
		"exit-async-1",
	}

	if diff := cmp.Diff(expected, events); diff != "" {
		t.Errorf("Event order mismatch (-want +got):\n%s", diff)
	}
}
