// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package pyasyncio

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// TaskGroupError aggregates multiple task errors from a TaskGroup.
//
// This is equivalent to Python's [ExceptionGroup] when used with TaskGroup.
//
// [ExceptionGroup]: https://docs.python.org/3/library/exceptions.html#ExceptionGroup
type TaskGroupError struct {
	// Errors contains all the errors from failed tasks in the group
	Errors []error
	// Message provides a summary of the error
	Message string
}

// Error implements the error interface for TaskGroupError.
func (e *TaskGroupError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("task group failed with %d error(s)", len(e.Errors))
}

// Unwrap returns the underlying errors for Go 1.20+ error handling.
func (e *TaskGroupError) Unwrap() []error {
	return e.Errors
}

// TaskGroup provides structured concurrency for a group of tasks.
//
// This is equivalent to Python's [asyncio.TaskGroup].
//
// TaskGroup ensures that if any task in the group fails, all other tasks
// are automatically cancelled. This prevents resource leaks and provides
// fail-fast behavior.
//
// [asyncio.TaskGroup]: https://docs.python.org/3/library/asyncio-task.html#asyncio.TaskGroup
type TaskGroup[T any] struct {
	// mu protects all mutable fields
	mu sync.RWMutex

	// ctx is the context for the entire task group
	ctx context.Context
	// cancel cancels all tasks in the group
	cancel context.CancelFunc

	// tasks contains all tasks in the group
	tasks []*Task[T]

	// completion tracking
	done     chan struct{}
	finished atomic.Int64 // atomic flag indicating completion

	// error tracking
	errors     []error
	firstError error

	// results from successful tasks
	results []T

	// activeCount tracks how many tasks are still running
	activeCount atomic.Int64
}

// NewTaskGroup creates a new TaskGroup.
//
// This is equivalent to Python's [asyncio.TaskGroup] constructor.
//
// The context is used to cancel all tasks in the group if the parent
// context is cancelled.
//
// [asyncio.TaskGroup]: https://docs.python.org/3/library/asyncio-task.html#asyncio.TaskGroup
func NewTaskGroup[T any](ctx context.Context) *TaskGroup[T] {
	groupCtx, cancel := context.WithCancel(ctx)

	return &TaskGroup[T]{
		ctx:    groupCtx,
		cancel: cancel,
		done:   make(chan struct{}),
	}
}

// CreateTask adds a new task to the group and starts it immediately.
//
// This is equivalent to Python's [asyncio.TaskGroup.create_task].
//
// The task will be cancelled if any other task in the group fails,
// or if the group's context is cancelled.
//
// Returns the created task and any error from task creation.
//
// [asyncio.TaskGroup.create_task]: https://docs.python.org/3/library/asyncio-task.html#asyncio.TaskGroup.create_task
func (tg *TaskGroup[T]) CreateTask(fn func(context.Context) (T, error)) (*Task[T], error) {
	return tg.CreateNamedTask("", fn)
}

// CreateNamedTask adds a new named task to the group and starts it immediately.
//
// Similar to CreateTask but allows specifying a name for debugging.
func (tg *TaskGroup[T]) CreateNamedTask(name string, fn func(context.Context) (T, error)) (*Task[T], error) {
	if fn == nil {
		return nil, fmt.Errorf("task function cannot be nil")
	}

	tg.mu.Lock()
	defer tg.mu.Unlock()

	// Check if group is already finished
	if tg.finished.Load() != 0 {
		return nil, fmt.Errorf("cannot add task to finished task group")
	}

	// Create task with group context
	task := CreateNamedTask(tg.ctx, name, fn)

	// Add to group
	tg.tasks = append(tg.tasks, task)
	tg.activeCount.Add(1)

	// Monitor task completion
	go tg.monitorTask(task)

	return task, nil
}

// monitorTask watches a task for completion and handles errors.
func (tg *TaskGroup[T]) monitorTask(task *Task[T]) {
	// Wait for task to complete
	<-task.done

	tg.mu.Lock()
	defer tg.mu.Unlock()

	// Decrease active count
	remaining := tg.activeCount.Add(-1)

	// Check task result
	result, err := task.Result()

	if err != nil {
		// Task failed - record error and cancel group
		tg.errors = append(tg.errors, err)
		if tg.firstError == nil {
			tg.firstError = err
		}

		// Cancel all other tasks (structured concurrency)
		tg.cancel()
	} else {
		// Task succeeded - store result
		tg.results = append(tg.results, result)
	}

	// If this was the last task, signal completion
	if remaining == 0 {
		tg.finished.Store(1)
		close(tg.done)
	}
}

// Wait waits for all tasks in the group to complete.
//
// This is equivalent to the implicit wait when exiting a Python [asyncio.TaskGroup] context.
//
// Returns all successful results and any aggregated errors.
// If any task failed, returns a TaskGroupError containing all errors.
//
// The context can be used to timeout the wait, but this will not cancel
// the tasks themselves - use Cancel() for that.
//
// [asyncio.TaskGroup]: https://docs.python.org/3/library/asyncio-task.html#asyncio.TaskGroup
func (tg *TaskGroup[T]) Wait(ctx context.Context) ([]T, error) {
	// If no tasks were created, return immediately
	tg.mu.RLock()
	taskCount := len(tg.tasks)
	tg.mu.RUnlock()

	if taskCount == 0 {
		return nil, nil
	}

	// Wait for completion or context cancellation
	select {
	case <-tg.done:
		// All tasks completed
		tg.mu.RLock()
		defer tg.mu.RUnlock()

		// Return results and any errors
		if len(tg.errors) > 0 {
			return tg.results, &TaskGroupError{
				Errors:  tg.errors,
				Message: fmt.Sprintf("task group failed with %d error(s)", len(tg.errors)),
			}
		}
		return tg.results, nil

	case <-ctx.Done():
		// Wait context cancelled
		return nil, ctx.Err()
	}
}

// Cancel cancels all tasks in the group.
//
// This immediately cancels all running tasks and prevents new tasks
// from being added to the group.
func (tg *TaskGroup[T]) Cancel() {
	tg.cancel()
}

// Done returns a channel that is closed when all tasks complete.
//
// This channel is closed regardless of whether tasks succeeded or failed.
func (tg *TaskGroup[T]) Done() <-chan struct{} {
	return tg.done
}

// Cancelled returns true if the group has been cancelled.
func (tg *TaskGroup[T]) Cancelled() bool {
	return tg.ctx.Err() != nil
}

// TaskCount returns the number of tasks in the group.
func (tg *TaskGroup[T]) TaskCount() int {
	tg.mu.RLock()
	defer tg.mu.RUnlock()
	return len(tg.tasks)
}

// ActiveCount returns the number of tasks still running.
func (tg *TaskGroup[T]) ActiveCount() int {
	return int(tg.activeCount.Load())
}

// Tasks returns a copy of all tasks in the group.
func (tg *TaskGroup[T]) Tasks() []*Task[T] {
	tg.mu.RLock()
	defer tg.mu.RUnlock()

	tasks := make([]*Task[T], len(tg.tasks))
	copy(tasks, tg.tasks)
	return tasks
}

// Context returns the context associated with this task group.
func (tg *TaskGroup[T]) Context() context.Context {
	return tg.ctx
}

// Close cancels all tasks and releases resources.
//
// This should be called when the TaskGroup is no longer needed,
// typically in a defer statement.
func (tg *TaskGroup[T]) Close() {
	tg.Cancel()
}

// String returns a string representation of the TaskGroup.
func (tg *TaskGroup[T]) String() string {
	tg.mu.RLock()
	defer tg.mu.RUnlock()

	taskCount := len(tg.tasks)
	activeCount := tg.activeCount.Load()
	errorCount := len(tg.errors)

	return fmt.Sprintf("TaskGroup[%d tasks, %d active, %d errors]",
		taskCount, activeCount, errorCount)
}

// WaitForCompletion is a convenience method that waits for all tasks to complete
// and returns the results, ignoring any context cancellation.
//
// This is useful when you want to wait indefinitely for tasks to complete.
// Use Wait() if you need timeout or cancellation support.
func (tg *TaskGroup[T]) WaitForCompletion() ([]T, error) {
	<-tg.done

	tg.mu.RLock()
	defer tg.mu.RUnlock()

	if len(tg.errors) > 0 {
		return tg.results, &TaskGroupError{
			Errors:  tg.errors,
			Message: fmt.Sprintf("task group failed with %d error(s)", len(tg.errors)),
		}
	}
	return tg.results, nil
}

// Gather creates tasks for all provided functions and waits for completion.
//
// This is a convenience method similar to Python's [asyncio.gather].
//
// Unlike a manual TaskGroup, this method creates all tasks at once
// and waits for completion before returning.
//
// [asyncio.gather]: https://docs.python.org/3/library/asyncio-task.html#asyncio.gather
func Gather[T any](ctx context.Context, fns ...func(context.Context) (T, error)) ([]T, error) {
	if len(fns) == 0 {
		return nil, nil
	}

	tg := NewTaskGroup[T](ctx)
	defer tg.Close()

	// Create all tasks
	for i, fn := range fns {
		name := fmt.Sprintf("gather-task-%d", i)
		if _, err := tg.CreateNamedTask(name, fn); err != nil {
			return nil, fmt.Errorf("failed to create task %d: %w", i, err)
		}
	}

	// Wait for completion
	return tg.Wait(ctx)
}
