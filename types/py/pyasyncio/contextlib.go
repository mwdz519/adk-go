// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package pyasyncio

import (
	"context"
	"fmt"
	"sync"
)

// AsyncContextManager represents an asynchronous context manager similar to Python's async context managers.
// It provides AEnter and AExit methods that can be used with async operations.
//
// This is equivalent to Python's async context manager protocol (__aenter__ and __aexit__).
type AsyncContextManager interface {
	// AEnter is called when entering the async context.
	// It returns the value that should be used in the context.
	AEnter(ctx context.Context) (any, error)

	// AExit is called when exiting the async context.
	// The exc parameter contains any error that occurred in the context.
	// Returns true if the error should be suppressed, false otherwise.
	AExit(ctx context.Context, exc error) (bool, error)
}

// ContextManager represents a synchronous context manager.
// This is for compatibility with sync context managers in AsyncExitStack.
type ContextManager interface {
	// Enter is called when entering the context.
	Enter() (any, error)

	// Exit is called when exiting the context.
	// The exc parameter contains any error that occurred in the context.
	// Returns true if the error should be suppressed, false otherwise.
	Exit(exc error) (bool, error)
}

// exitCallback represents a callback to be executed on exit.
type exitCallback struct {
	// isAsync indicates whether this is an async callback
	isAsync bool

	// asyncFn is the async function to call (if isAsync is true)
	asyncFn func(context.Context) error

	// syncFn is the sync function to call (if isAsync is false)
	syncFn func() error

	// asyncExitFn is the async exit function (for context managers)
	asyncExitFn func(context.Context, error) (bool, error)

	// syncExitFn is the sync exit function (for context managers)
	syncExitFn func(error) (bool, error)

	// isExitFn indicates this is an exit function (not a callback)
	isExitFn bool

	// args holds arguments for callbacks with args
	args []any

	// asyncFnWithArgs is the async function with args to call
	asyncFnWithArgs func(context.Context, ...any) error

	// syncFnWithArgs is the sync function with args to call
	syncFnWithArgs func(...any) error
}

// AsyncExitStack is a context manager for dynamic management of async and sync context managers.
// It enables programmatic construction of async with statements.
//
// This is equivalent to Python's [contextlib.AsyncExitStack].
//
// [contextlib.AsyncExitStack]: https://docs.python.org/3/library/contextlib.html#contextlib.AsyncExitStack
type AsyncExitStack struct {
	mu        sync.Mutex
	callbacks []exitCallback
	entered   bool
	closed    bool
}

// NewAsyncExitStack creates a new AsyncExitStack.
func NewAsyncExitStack() *AsyncExitStack {
	return &AsyncExitStack{
		callbacks: make([]exitCallback, 0),
	}
}

// AEnter implements the AsyncContextManager interface.
// This allows AsyncExitStack to be used as an async context manager itself.
func (s *AsyncExitStack) AEnter(ctx context.Context) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.entered {
		return nil, fmt.Errorf("AsyncExitStack already entered")
	}
	s.entered = true
	return s, nil
}

// AExit implements the AsyncContextManager interface.
// It unwinds the callback stack, calling all registered exit callbacks.
func (s *AsyncExitStack) AExit(ctx context.Context, exc error) (bool, error) {
	return s.unwindStack(ctx, exc)
}

// EnterAsyncContext enters an async context manager and schedules its exit.
// Returns the result of the context manager's AEnter method.
func (s *AsyncExitStack) EnterAsyncContext(ctx context.Context, cm AsyncContextManager) (any, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, fmt.Errorf("AsyncExitStack is closed")
	}
	s.mu.Unlock()

	// Enter the context
	result, err := cm.AEnter(ctx)
	if err != nil {
		return nil, err
	}

	// Schedule the exit
	s.mu.Lock()
	s.callbacks = append(s.callbacks, exitCallback{
		isAsync:     true,
		asyncExitFn: cm.AExit,
		isExitFn:    true,
	})
	s.mu.Unlock()

	return result, nil
}

// EnterContext enters a sync context manager and schedules its exit.
// Returns the result of the context manager's Enter method.
func (s *AsyncExitStack) EnterContext(cm ContextManager) (any, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, fmt.Errorf("AsyncExitStack is closed")
	}
	s.mu.Unlock()

	// Enter the context
	result, err := cm.Enter()
	if err != nil {
		return nil, err
	}

	// Schedule the exit
	s.mu.Lock()
	s.callbacks = append(s.callbacks, exitCallback{
		isAsync:    false,
		syncExitFn: cm.Exit,
		isExitFn:   true,
	})
	s.mu.Unlock()

	return result, nil
}

// PushAsyncExit adds an async exit callback to the stack.
// The callback should have the same signature as AsyncContextManager.AExit.
// Returns the exit function to enable decorator pattern usage.
func (s *AsyncExitStack) PushAsyncExit(exit func(context.Context, error) (bool, error)) func(context.Context, error) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return exit
	}

	s.callbacks = append(s.callbacks, exitCallback{
		isAsync:     true,
		asyncExitFn: exit,
		isExitFn:    true,
	})

	return exit
}

// PushAsyncCallback adds an async callback to be called on exit.
// The callback is called with the provided context.
// Returns the callback to enable decorator pattern usage.
func (s *AsyncExitStack) PushAsyncCallback(callback func(context.Context) error) func(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return callback
	}

	s.callbacks = append(s.callbacks, exitCallback{
		isAsync: true,
		asyncFn: callback,
	})

	return callback
}

// Callback adds a sync callback to be called on exit.
// Returns the callback to enable decorator pattern usage.
func (s *AsyncExitStack) Callback(callback func() error) func() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return callback
	}

	s.callbacks = append(s.callbacks, exitCallback{
		isAsync: false,
		syncFn:  callback,
	})

	return callback
}

// Push adds a sync exit callback to the stack.
// The callback should have the same signature as ContextManager.Exit.
// Returns the exit function to enable decorator pattern usage.
func (s *AsyncExitStack) Push(exit func(error) (bool, error)) func(error) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return exit
	}

	s.callbacks = append(s.callbacks, exitCallback{
		isAsync:    false,
		syncExitFn: exit,
		isExitFn:   true,
	})

	return exit
}

// PushAsyncCallbackArgs adds an async callback with arguments to be called on exit.
// The callback is called with the provided context and arguments.
// Returns the callback to enable decorator pattern usage.
func (s *AsyncExitStack) PushAsyncCallbackArgs(callback func(context.Context, ...any) error, args ...any) func(context.Context, ...any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return callback
	}

	s.callbacks = append(s.callbacks, exitCallback{
		isAsync:         true,
		asyncFnWithArgs: callback,
		args:            args,
	})

	return callback
}

// CallbackArgs adds a sync callback with arguments to be called on exit.
// Returns the callback to enable decorator pattern usage.
func (s *AsyncExitStack) CallbackArgs(callback func(...any) error, args ...any) func(...any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return callback
	}

	s.callbacks = append(s.callbacks, exitCallback{
		isAsync:        false,
		syncFnWithArgs: callback,
		args:           args,
	})

	return callback
}

// PopAll transfers all callbacks to a new AsyncExitStack.
// This clears the current stack and returns a new stack with all the callbacks.
func (s *AsyncExitStack) PopAll() *AsyncExitStack {
	s.mu.Lock()
	defer s.mu.Unlock()

	newStack := NewAsyncExitStack()
	newStack.callbacks = s.callbacks
	s.callbacks = make([]exitCallback, 0)

	return newStack
}

// AClose immediately unwinds the callback stack.
// This is equivalent to exiting the async context manager.
func (s *AsyncExitStack) AClose(ctx context.Context) error {
	_, err := s.unwindStack(ctx, nil)
	return err
}

// unwindStack calls all callbacks in reverse order (LIFO).
// It properly handles both sync and async callbacks and error propagation.
func (s *AsyncExitStack) unwindStack(ctx context.Context, exc error) (bool, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return false, nil
	}
	s.closed = true

	// Copy callbacks and clear the stack
	callbacks := s.callbacks
	s.callbacks = nil
	s.mu.Unlock()

	// Track if any exit handler suppressed the exception
	suppressed := false
	pendingExc := exc

	// Process callbacks in reverse order (LIFO)
	for i := len(callbacks) - 1; i >= 0; i-- {
		cb := callbacks[i]

		if cb.isExitFn {
			// This is an exit function from a context manager
			var suppress bool
			var err error

			if cb.isAsync && cb.asyncExitFn != nil {
				suppress, err = cb.asyncExitFn(ctx, pendingExc)
			} else if !cb.isAsync && cb.syncExitFn != nil {
				suppress, err = cb.syncExitFn(pendingExc)
			}

			if err != nil {
				// Exit function failed, this becomes the new pending exception
				pendingExc = err
			} else if suppress && pendingExc != nil {
				// Exception was suppressed
				suppressed = true
				pendingExc = nil
			}
		} else {
			// This is a regular callback
			var err error

			if cb.isAsync {
				if cb.asyncFnWithArgs != nil {
					err = cb.asyncFnWithArgs(ctx, cb.args...)
				} else if cb.asyncFn != nil {
					err = cb.asyncFn(ctx)
				}
			} else {
				if cb.syncFnWithArgs != nil {
					err = cb.syncFnWithArgs(cb.args...)
				} else if cb.syncFn != nil {
					err = cb.syncFn()
				}
			}

			if err != nil && pendingExc == nil {
				// Callback failed and no pending exception
				pendingExc = err
			}
		}
	}

	// Return whether the original exception was suppressed and any pending exception
	return suppressed, pendingExc
}

// asyncContextManagerFunc is a helper to create AsyncContextManager from functions
type asyncContextManagerFunc struct {
	enter func(context.Context) (any, error)
	exit  func(context.Context, error) (bool, error)
}

func (f *asyncContextManagerFunc) AEnter(ctx context.Context) (any, error) {
	return f.enter(ctx)
}

func (f *asyncContextManagerFunc) AExit(ctx context.Context, exc error) (bool, error) {
	return f.exit(ctx, exc)
}

// NewAsyncContextManager creates an AsyncContextManager from enter and exit functions.
// This is a convenience function for creating simple async context managers.
func NewAsyncContextManager(enter func(context.Context) (any, error), exit func(context.Context, error) (bool, error)) AsyncContextManager {
	return &asyncContextManagerFunc{
		enter: enter,
		exit:  exit,
	}
}

// contextManagerFunc is a helper to create ContextManager from functions
type contextManagerFunc struct {
	enter func() (any, error)
	exit  func(error) (bool, error)
}

func (f *contextManagerFunc) Enter() (any, error) {
	return f.enter()
}

func (f *contextManagerFunc) Exit(exc error) (bool, error) {
	return f.exit(exc)
}

// NewContextManager creates a ContextManager from enter and exit functions.
// This is a convenience function for creating simple context managers.
func NewContextManager(enter func() (any, error), exit func(error) (bool, error)) ContextManager {
	return &contextManagerFunc{
		enter: enter,
		exit:  exit,
	}
}
