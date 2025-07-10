// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package pyasyncio

import (
	"context"
	"time"
)

// TimeoutError is raised when an operation times out.
//
// This is equivalent to Python's [asyncio.TimeoutError].
//
// [asyncio.TimeoutError]: https://docs.python.org/3/library/asyncio-exceptions.html#asyncio.TimeoutError
type TimeoutError struct {
	// Message provides additional context about the timeout
	Message string
	// Timeout is the duration that was exceeded
	Timeout time.Duration
}

// Error implements the error interface for TimeoutError.
func (e *TimeoutError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Timeout > 0 {
		return "operation timed out after " + e.Timeout.String()
	}
	return "operation timed out"
}

// NewTimeoutError creates a new TimeoutError with the specified timeout duration.
func NewTimeoutError(timeout time.Duration) error {
	return &TimeoutError{
		Timeout: timeout,
	}
}

// NewTimeoutErrorWithMessage creates a new TimeoutError with a custom message and timeout.
func NewTimeoutErrorWithMessage(message string, timeout time.Duration) error {
	return &TimeoutError{
		Message: message,
		Timeout: timeout,
	}
}

// WaitFor waits for a function to complete within the specified timeout.
//
// This is equivalent to Python's [asyncio.wait_for].
//
// If the function completes within the timeout, its result and error are returned.
// If the timeout elapses before completion, the function is cancelled (via context)
// and a TimeoutError is returned.
//
// The function receives a context that will be cancelled if the timeout elapses,
// allowing it to cooperatively terminate early.
//
// Example:
//
//	result, err := pyasyncio.WaitFor(ctx, 5*time.Second, func(ctx context.Context) (string, error) {
//		// Your async work here
//		select {
//		case <-time.After(2*time.Second):
//			return "completed", nil
//		case <-ctx.Done():
//			return "", ctx.Err()
//		}
//	})
//
// [asyncio.wait_for]: https://docs.python.org/3/library/asyncio-task.html#asyncio.wait_for
func WaitFor[T any](ctx context.Context, timeout time.Duration, fn func(context.Context) (T, error)) (T, error) {
	if fn == nil {
		var zero T
		return zero, &TimeoutError{Message: "function cannot be nil"}
	}

	if timeout <= 0 {
		var zero T
		return zero, &TimeoutError{Message: "timeout must be positive"}
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create a task to execute the function
	task := CreateTask(timeoutCtx, fn)

	// Wait for task completion with the timeout context
	result, err := task.Wait(timeoutCtx)
	// Check what kind of error we got
	if err != nil {
		// If timeout context was cancelled due to deadline, return TimeoutError
		if timeoutCtx.Err() == context.DeadlineExceeded {
			task.Cancel()
			var zero T
			return zero, NewTimeoutError(timeout)
		}
		// Otherwise, return the original error (could be parent context cancellation)
		var zero T
		return zero, err
	}

	// Task completed successfully
	return result, nil
}

// WaitForTask waits for an existing task to complete within the specified timeout.
//
// This is a convenience function for applying timeout semantics to an already-created task.
// Unlike WaitFor, this function does not create a new task but applies timeout to an existing one.
//
// If the task completes within the timeout, its result and error are returned.
// If the timeout elapses before completion, the task is cancelled and a TimeoutError is returned.
//
// Note: The task's original context is not modified. Instead, a new timeout context
// is used only for the wait operation. The task itself may continue running after
// timeout unless it's explicitly cancelled.
//
// Example:
//
//	task := pyasyncio.CreateTask(ctx, myFunction)
//	result, err := pyasyncio.WaitForTask(ctx, 5*time.Second, task)
func WaitForTask[T any](ctx context.Context, timeout time.Duration, task *Task[T]) (T, error) {
	if task == nil {
		var zero T
		return zero, &TimeoutError{Message: "task cannot be nil"}
	}

	if timeout <= 0 {
		var zero T
		return zero, &TimeoutError{Message: "timeout must be positive"}
	}

	// Create a timeout context for the wait operation
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Wait for task completion or timeout
	result, err := task.Wait(timeoutCtx)
	// Check what kind of error we got
	if err != nil {
		// If timeout context was cancelled due to deadline, return TimeoutError
		if timeoutCtx.Err() == context.DeadlineExceeded {
			task.Cancel()
			var zero T
			return zero, NewTimeoutError(timeout)
		}
		// Otherwise, return the original error
		var zero T
		return zero, err
	}

	// Task completed successfully
	return result, nil
}

// WaitForAny waits for any of the provided functions to complete within the timeout.
//
// This function runs all provided functions concurrently and returns the result
// of the first one to complete successfully. If all functions fail or timeout,
// it returns an appropriate error.
//
// This is similar to Python's [asyncio.wait_for] combined with [asyncio.wait] with
// return_when=FIRST_COMPLETED, but simplified for common use cases.
//
// All functions receive the same timeout context and will be cancelled if the
// timeout elapses or if any function completes successfully.
//
// Example:
//
//	result, err := pyasyncio.WaitForAny(ctx, 5*time.Second,
//		func(ctx context.Context) (string, error) { return "fast", nil },
//		func(ctx context.Context) (string, error) {
//			time.Sleep(10*time.Second)
//			return "slow", nil
//		},
//	)
//
// [asyncio.wait_for]: https://docs.python.org/3/library/asyncio-task.html#asyncio.wait_for
// [asyncio.wait]: https://docs.python.org/3/library/asyncio-task.html#asyncio.wait
func WaitForAny[T any](ctx context.Context, timeout time.Duration, fns ...func(context.Context) (T, error)) (T, error) {
	if len(fns) == 0 {
		var zero T
		return zero, &TimeoutError{Message: "at least one function must be provided"}
	}

	if timeout <= 0 {
		var zero T
		return zero, &TimeoutError{Message: "timeout must be positive"}
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create a task group to manage all functions
	tg := NewTaskGroup[T](timeoutCtx)
	defer tg.Close()

	// Create tasks for all functions
	for i, fn := range fns {
		if fn != nil {
			_, err := tg.CreateNamedTask("wait-for-any-"+string(rune('0'+i)), fn)
			if err != nil {
				var zero T
				return zero, err
			}
		}
	}

	// Use a separate goroutine to wait for the first successful completion
	type result struct {
		value T
		err   error
	}

	resultCh := make(chan result, 1)

	go func() {
		// Monitor tasks for first success
		tasks := tg.Tasks()
		for len(tasks) > 0 && !tg.Cancelled() {
			// Check each task
			for _, task := range tasks {
				if task.Done() {
					if !task.Cancelled() {
						// Task completed (successfully or with error)
						val, err := task.Result()
						if err == nil {
							// First successful result
							resultCh <- result{value: val, err: nil}
							return
						}
					}
				}
			}

			// Brief sleep to avoid busy waiting
			time.Sleep(1 * time.Millisecond)

			// Get updated task list
			tasks = tg.Tasks()
		}

		// All tasks completed without success, or group was cancelled
		var zero T
		if timeoutCtx.Err() != nil {
			resultCh <- result{value: zero, err: NewTimeoutError(timeout)}
		} else {
			resultCh <- result{value: zero, err: &TimeoutError{Message: "all functions failed"}}
		}
	}()

	// Wait for first result or timeout
	select {
	case res := <-resultCh:
		return res.value, res.err
	case <-timeoutCtx.Done():
		var zero T
		return zero, NewTimeoutError(timeout)
	}
}

// WaitForAll waits for all provided functions to complete within the timeout.
//
// This function runs all provided functions concurrently and waits for all
// to complete successfully. If any function fails or the timeout elapses,
// all functions are cancelled and an error is returned.
//
// This is similar to Python's [asyncio.wait_for] combined with [asyncio.gather].
//
// All functions receive the same timeout context and will be cancelled if the
// timeout elapses or if any function fails.
//
// Example:
//
//	results, err := pyasyncio.WaitForAll(ctx, 5*time.Second,
//		func(ctx context.Context) (string, error) { return "first", nil },
//		func(ctx context.Context) (string, error) { return "second", nil },
//	)
//
// [asyncio.wait_for]: https://docs.python.org/3/library/asyncio-task.html#asyncio.wait_for
// [asyncio.gather]: https://docs.python.org/3/library/asyncio-task.html#asyncio.gather
func WaitForAll[T any](ctx context.Context, timeout time.Duration, fns ...func(context.Context) (T, error)) ([]T, error) {
	if len(fns) == 0 {
		return nil, nil
	}

	if timeout <= 0 {
		return nil, &TimeoutError{Message: "timeout must be positive"}
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Use Gather with the timeout context
	results, err := Gather(timeoutCtx, fns...)

	// Check if we timed out
	if timeoutCtx.Err() != nil {
		return nil, NewTimeoutError(timeout)
	}

	return results, err
}

// Shield protects a function from cancellation while still respecting timeouts.
//
// This is equivalent to Python's [asyncio.shield].
//
// The function continues to run even if the calling context is cancelled,
// but it can still be subject to timeouts. This is useful when you want to
// ensure a critical operation completes even if the caller gives up waiting.
//
// Note: The shielded function should still check its context for timeout
// cancellation to respect timeout semantics.
//
// Example:
//
//	result, err := pyasyncio.Shield(ctx, func(ctx context.Context) (string, error) {
//		// This continues running even if parent context is cancelled
//		// but still respects timeout from WaitFor
//		return criticalOperation(ctx)
//	})
//
// [asyncio.shield]: https://docs.python.org/3/library/asyncio-task.html#asyncio.shield
func Shield[T any](ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	if fn == nil {
		var zero T
		return zero, &TimeoutError{Message: "function cannot be nil"}
	}

	// Create a new context that is not cancelled when parent is cancelled
	// but still inherits deadline/timeout behavior
	shieldCtx := context.Background()

	// If parent has a deadline, apply it to shield context
	if deadline, ok := ctx.Deadline(); ok {
		var cancel context.CancelFunc
		shieldCtx, cancel = context.WithDeadline(shieldCtx, deadline)
		defer cancel()
	}

	// Execute function with shielded context
	return fn(shieldCtx)
}
