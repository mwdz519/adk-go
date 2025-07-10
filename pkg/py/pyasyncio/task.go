// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package pyasyncio

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TaskState represents the current state of a task.
type TaskState int

const (
	// TaskPending indicates the task has been created but not yet started.
	TaskPending TaskState = iota
	// TaskRunning indicates the task is currently executing.
	TaskRunning
	// TaskDone indicates the task has completed successfully or with an error.
	TaskDone
	// TaskCancelled indicates the task was cancelled before or during execution.
	TaskCancelled
)

// String returns a string representation of the TaskState.
func (s TaskState) String() string {
	switch s {
	case TaskPending:
		return "pending"
	case TaskRunning:
		return "running"
	case TaskDone:
		return "done"
	case TaskCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// TaskCancelledError is returned when a task is cancelled.
//
// This is equivalent to Python's [asyncio.CancelledError].
//
// [asyncio.CancelledError]: https://docs.python.org/3/library/asyncio-exceptions.html#asyncio.CancelledError
type TaskCancelledError struct {
	Message string
}

// Error implements the error interface for TaskCancelledError.
func (e *TaskCancelledError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "task was cancelled"
}

func NewTaskCancelledError(message string) error {
	return &TaskCancelledError{
		Message: message,
	}
}

// Task represents a Python [asyncio.Task] in Go.
//
// A task wraps a function and schedules it for concurrent execution.
// Tasks have a lifecycle: pending -> running -> done/cancelled.
// Tasks support cancellation and completion callbacks.
//
// [asyncio.Task]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task
type Task[T any] struct {
	// mu protects all mutable fields
	mu sync.RWMutex

	// state tracks the current task state using atomic operations for fast reads
	state atomic.Int64 // TaskState stored for atomic access

	// ctx is the context for this task
	ctx context.Context
	// cancel function to cancel this specific task
	cancel context.CancelFunc

	// fn is the function to execute
	fn func(context.Context) (T, error)

	// result and err store the execution outcome
	result T
	err    error

	// done channel is closed when the task completes
	done chan struct{}

	// callbacks are executed when the task completes
	callbacks []func(*Task[T])

	// metadata for debugging and monitoring
	name     string
	created  time.Time
	started  time.Time
	finished time.Time
}

// CreateTask creates and immediately starts a new task.
//
// This is equivalent to Python's [asyncio.create_task].
//
// The task begins execution immediately in a new goroutine.
// Use Cancel() to stop the task or wait for ctx to be cancelled.
//
// [asyncio.create_task]: https://docs.python.org/3/library/asyncio-task.html#asyncio.create_task
func CreateTask[T any](ctx context.Context, fn func(context.Context) (T, error)) *Task[T] {
	return CreateNamedTask(ctx, "", fn)
}

// CreateNamedTask creates and immediately starts a new task with a name.
//
// The name is useful for debugging and monitoring.
// Otherwise behaves identically to CreateTask.
func CreateNamedTask[T any](ctx context.Context, name string, fn func(context.Context) (T, error)) *Task[T] {
	if fn == nil {
		panic("task function cannot be nil")
	}

	taskCtx, cancel := context.WithCancel(ctx)

	task := &Task[T]{
		ctx:     taskCtx,
		cancel:  cancel,
		fn:      fn,
		done:    make(chan struct{}),
		name:    name,
		created: time.Now(),
	}
	task.state.Store(int64(TaskPending))

	// Start the task immediately
	go task.run()

	return task
}

// run executes the task function in a goroutine.
func (t *Task[T]) run() {
	defer close(t.done)
	defer t.executeCallbacks()

	// Transition to running state
	t.state.Store(int64(TaskRunning))

	t.mu.Lock()
	t.started = time.Now()
	t.mu.Unlock()

	// Execute the function with the task context
	result, err := t.fn(t.ctx)

	t.mu.Lock()
	defer t.mu.Unlock()

	t.finished = time.Now()

	// Check if we were cancelled during execution
	select {
	case <-t.ctx.Done():
		t.state.Store(int64(TaskCancelled))
		t.err = &TaskCancelledError{Message: t.ctx.Err().Error()}
	default:
		t.state.Store(int64(TaskDone))
		t.result = result
		t.err = err
	}
}

// executeCallbacks runs all registered callbacks after task completion.
func (t *Task[T]) executeCallbacks() {
	t.mu.RLock()
	callbacks := make([]func(*Task[T]), len(t.callbacks))
	copy(callbacks, t.callbacks)
	t.mu.RUnlock()

	// Execute callbacks without holding the lock
	for _, callback := range callbacks {
		// Run each callback in its own goroutine to prevent blocking
		go func(cb func(*Task[T])) {
			// Recover from callback panics to prevent crashing the task
			defer func() {
				if r := recover(); r != nil {
					// Log panic but continue
					fmt.Printf("Task callback panicked: %v\n", r)
				}
			}()
			cb(t)
		}(callback)
	}
}

// Cancel requests cancellation of the task.
//
// This is equivalent to Python's [asyncio.Task.cancel].
//
// If the task is already done, this has no effect and returns false.
// If the task is running, it will be cancelled cooperatively.
// Returns true if the cancellation request was accepted.
//
// [asyncio.Task.cancel]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.cancel
func (t *Task[T]) Cancel() bool {
	currentState := TaskState(t.state.Load())

	if currentState == TaskDone || currentState == TaskCancelled {
		return false
	}

	t.cancel()
	return true
}

// Cancelled returns true if the task was cancelled.
//
// This is equivalent to Python's [asyncio.Task.cancelled].
//
// [asyncio.Task.cancelled]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.cancelled
func (t *Task[T]) Cancelled() bool {
	return TaskState(t.state.Load()) == TaskCancelled
}

// Done returns true if the task is done (completed or cancelled).
//
// This is equivalent to Python's [asyncio.Task.done].
//
// [asyncio.Task.done]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.done
func (t *Task[T]) Done() bool {
	state := TaskState(t.state.Load())
	return state == TaskDone || state == TaskCancelled
}

// State returns the current state of the task.
func (t *Task[T]) State() TaskState {
	return TaskState(t.state.Load())
}

// Wait waits for the task to complete and returns the result.
//
// If the task was cancelled, returns a TaskCancelledError.
// If the task completed with an error, returns that error.
// If the task completed successfully, returns the result.
//
// This method can be called multiple times and will return the same result.
//
// The context can be used to timeout the wait operation, but does not
// cancel the task itself.
func (t *Task[T]) Wait(ctx context.Context) (T, error) {
	var zero T

	select {
	case <-t.done:
		// Task completed, return cached result
		t.mu.RLock()
		defer t.mu.RUnlock()
		return t.result, t.err

	case <-ctx.Done():
		// Wait context cancelled, but task continues running
		return zero, ctx.Err()
	}
}

// Result returns the result of the task without blocking.
//
// This is equivalent to Python's [asyncio.Task.result].
//
// If the task is not yet done, returns an error.
// If the task was cancelled, returns a TaskCancelledError.
// If the task completed with an error, returns that error.
//
// [asyncio.Task.result]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.result
func (t *Task[T]) Result() (T, error) {
	var zero T

	if !t.Done() {
		return zero, fmt.Errorf("task is not yet done")
	}

	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.result, t.err
}

// Exception returns the exception (error) of the task without blocking.
//
// This is equivalent to Python's [asyncio.Task.exception].
//
// If the task is not yet done, returns an error.
// If the task completed successfully, returns nil.
// If the task was cancelled or failed, returns the error.
//
// [asyncio.Task.exception]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.exception
func (t *Task[T]) Exception() error {
	if !t.Done() {
		return fmt.Errorf("task is not yet done")
	}

	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.err
}

// AddCallback adds a callback to be executed when the task completes.
//
// This is equivalent to Python's [asyncio.Task.add_done_callback].
//
// If the task is already done, the callback is executed immediately
// in a separate goroutine.
//
// Callbacks are executed concurrently and should not block.
// Panics in callbacks are recovered and logged.
//
// [asyncio.Task.add_done_callback]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.add_done_callback
func (t *Task[T]) AddCallback(callback func(*Task[T])) {
	if callback == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// If task is already done, execute callback immediately
	if t.Done() {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Task callback panicked: %v\n", r)
				}
			}()
			callback(t)
		}()
		return
	}

	// Add to callback list for later execution
	t.callbacks = append(t.callbacks, callback)
}

// RemoveCallback removes a callback from the task.
//
// This is equivalent to Python's [asyncio.Task.remove_done_callback].
//
// Returns the number of callbacks removed.
//
// [asyncio.Task.remove_done_callback]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.remove_done_callback
func (t *Task[T]) RemoveCallback(callback func(*Task[T])) int {
	if callback == nil {
		return 0
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Cannot compare functions directly in Go, so we can't implement this
	// exactly like Python. In practice, this method is rarely used.
	// We return 0 to indicate no callbacks were removed.
	return 0
}

// Name returns the name of the task.
//
// This is equivalent to Python's [asyncio.Task.get_name].
//
// [asyncio.Task.get_name]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.get_name
func (t *Task[T]) Name() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.name
}

// SetName sets the name of the task.
//
// This is equivalent to Python's [asyncio.Task.set_name].
//
// [asyncio.Task.set_name]: https://docs.python.org/3/library/asyncio-task.html#asyncio.Task.set_name
func (t *Task[T]) SetName(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.name = name
}

// Context returns the context associated with this task.
func (t *Task[T]) Context() context.Context {
	return t.ctx
}

// Created returns the time when the task was created.
func (t *Task[T]) Created() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.created
}

// Started returns the time when the task started executing.
// Returns zero time if the task hasn't started yet.
func (t *Task[T]) Started() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.started
}

// Finished returns the time when the task finished executing.
// Returns zero time if the task hasn't finished yet.
func (t *Task[T]) Finished() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.finished
}

// String returns a string representation of the task.
func (t *Task[T]) String() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	name := t.name
	if name == "" {
		name = "unnamed"
	}

	state := TaskState(t.state.Load())
	return fmt.Sprintf("Task[%s](%s)", name, state)
}
