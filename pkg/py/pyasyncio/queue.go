// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package pyasyncio

import (
	"context"
	"fmt"
	"sync"
)

// ErrQueueEmpty is raised when a non-blocking get operation is performed on an empty queue.
//
// This is equivalent to Python's [asyncio.QueueEmpty].
//
// [asyncio.QueueEmpty]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.QueueEmpty
type ErrQueueEmpty struct{}

// Error implements the error interface for ErrQueueEmpty.
func (e *ErrQueueEmpty) Error() string {
	return "queue is empty"
}

// ErrQueueFull is raised when a non-blocking put operation is performed on a full queue.
//
// This is equivalent to Python's [asyncio.QueueFull].
//
// [asyncio.QueueFull]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.QueueFull
type ErrQueueFull struct{}

// Error implements the error interface for ErrQueueFull.
func (e *ErrQueueFull) Error() string {
	return "queue is full"
}

// Queue defines the interface for asyncio-style queue operations.
//
// This interface matches Python's [asyncio.Queue] API.
//
// [asyncio.Queue]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue
type Queue[T any] interface {
	// Put adds an item to the queue, blocking if full.
	Put(ctx context.Context, item T) error

	// PutNowait adds an item to the queue without blocking.
	PutNowait(item T) error

	// Get removes and returns an item from the queue, blocking if empty.
	Get(ctx context.Context) (T, error)

	// GetNowait removes and returns an item from the queue without blocking.
	GetNowait() (T, error)

	// Size returns the current number of items in the queue.
	Size() int

	// Empty returns true if the queue is empty.
	Empty() bool

	// Full returns true if the queue is full (maxsize > 0 and queue is at capacity).
	Full() bool

	// TaskDone marks a task as done. Used with Join().
	TaskDone() error

	// Join waits until all tasks are done.
	Join(ctx context.Context) error
}

// queue represents a Python [asyncio.queue] in Go.
//
// A queue, useful for coordinating producer and consumer coroutines.
// If maxsize is less than or equal to zero, the queue size is infinite.
// Otherwise, put() blocks when the queue reaches maxsize until an item is removed by get().
//
// [asyncio.queue]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.queue
type queue[T any] struct {
	// mu protects all queue state
	mu sync.Mutex

	// notEmpty signals when items are available
	notEmpty *sync.Cond

	// notFull signals when space is available
	notFull *sync.Cond

	// allTasksDone signals when all tasks are complete
	allTasksDone *sync.Cond

	// maxsize is the maximum number of items allowed in the queue.
	// Zero or negative means unlimited.
	maxsize int

	// items stores the actual queue data
	items []T

	// unfinished tracks the number of tasks not yet marked as done
	unfinished int

	// closed indicates if the queue has been closed
	closed bool
}

var _ Queue[struct{}] = (*queue[struct{}])(nil)

// NewQueue creates a new Queue with the specified maximum size.
//
// If maxsize is less than or equal to zero, the queue size is infinite.
// Otherwise, put() blocks when the queue reaches maxsize until an item is removed by get().
//
// This is equivalent to Python's [asyncio.Queue] constructor.
//
// [asyncio.Queue]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue
func NewQueue[T any](maxsize int) *queue[T] {
	q := &queue[T]{
		maxsize: maxsize,
		items:   make([]T, 0),
	}

	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)
	q.allTasksDone = sync.NewCond(&q.mu)

	return q
}

// Size returns the number of items in the queue.
//
// This is equivalent to Python's [asyncio.Queue.qsize] method.
//
// [asyncio.Queue.qsize]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue.qsize
func (q *queue[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}

// Empty returns true if the queue is empty, false otherwise.
//
// This is equivalent to Python's [asyncio.Queue.empty] method.
//
// [asyncio.Queue.empty]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue.empty
func (q *queue[T]) Empty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items) == 0
}

// Full returns true if there are maxsize items in the queue.
//
// If the queue was initialized with maxsize=0 (the default), then Full() never returns true.
//
// This is equivalent to Python's [asyncio.Queue.full] method.
//
// [asyncio.Queue.full]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue.full
func (q *queue[T]) Full() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.maxsize <= 0 {
		return false
	}
	return len(q.items) >= q.maxsize
}

// putItem adds an item to the queue. Must be called with mutex held.
func (q *queue[T]) putItem(item T) {
	q.items = append(q.items, item)
	q.unfinished++
	q.notEmpty.Signal() // Wake up any waiting getters
}

// getItem removes and returns an item from the queue. Must be called with mutex held.
func (q *queue[T]) getItem() T {
	item := q.items[0]
	q.items = q.items[1:]
	q.notFull.Signal() // Wake up any waiting putters
	return item
}

// PutNowait puts an item into the queue without blocking.
//
// If no free slot is immediately available, raise ErrQueueFull.
//
// This is equivalent to Python's [asyncio.Queue.put_nowait] method.
//
// [asyncio.Queue.put_nowait]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue.put_nowait
func (q *queue[T]) PutNowait(item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	// Check if queue is full
	if q.maxsize > 0 && len(q.items) >= q.maxsize {
		return &ErrQueueFull{}
	}

	q.putItem(item)
	return nil
}

// Put an item into the queue.
//
// If the queue is full, wait until a free slot is available before adding the item.
// The operation can be cancelled through the context.
//
// This is equivalent to Python's [asyncio.Queue.put] method.
//
// [asyncio.Queue.put]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue.put
func (q *queue[T]) Put(ctx context.Context, item T) error {
	// Use a channel to coordinate between the waiting goroutine and context cancellation
	done := make(chan error, 1)
	cancelled := make(chan struct{})

	go func() {
		q.mu.Lock()
		defer q.mu.Unlock()

		if q.closed {
			done <- fmt.Errorf("queue is closed")
			return
		}

		// Wait until there's space in the queue or context is cancelled
		for q.maxsize > 0 && len(q.items) >= q.maxsize {
			// Release lock temporarily to check for cancellation
			q.mu.Unlock()
			select {
			case <-cancelled:
				q.mu.Lock()
				done <- ctx.Err()
				return
			default:
			}
			q.mu.Lock()

			if q.closed {
				done <- fmt.Errorf("queue is closed")
				return
			}

			// If still full, wait on condition
			if q.maxsize > 0 && len(q.items) >= q.maxsize {
				q.notFull.Wait()
			}
		}

		q.putItem(item)
		done <- nil
	}()

	select {
	case err := <-done:
		return err

	case <-ctx.Done():
		close(cancelled)
		q.notFull.Broadcast()
		// Wait for goroutine to finish
		<-done
		return ctx.Err()
	}
}

// GetNowait removes and returns an item if one is immediately available.
//
// If no item is immediately available, raise ErrQueueEmpty.
//
// This is equivalent to Python's [asyncio.Queue.get_nowait] method.
//
// [asyncio.Queue.get_nowait]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue.get_nowait
func (q *queue[T]) GetNowait() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var zero T

	if len(q.items) == 0 {
		return zero, &ErrQueueEmpty{}
	}

	return q.getItem(), nil
}

// Get removes and returns an item from the queue.
//
// If queue is empty, wait until an item is available.
// The operation can be cancelled through the context.
//
// This is equivalent to Python's [asyncio.Queue.get] method.
//
// [asyncio.Queue.get]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue.get
func (q *queue[T]) Get(ctx context.Context) (T, error) {
	// Use a channel to coordinate between the waiting goroutine and context cancellation
	type result struct {
		item T
		err  error
	}
	done := make(chan result, 1)
	cancelled := make(chan struct{})

	go func() {
		q.mu.Lock()
		defer q.mu.Unlock()

		var zero T

		// Wait until there's an item in the queue or context is cancelled
		for len(q.items) == 0 {
			// Release lock temporarily to check for cancellation
			q.mu.Unlock()
			select {
			case <-cancelled:
				q.mu.Lock()
				done <- result{item: zero, err: ctx.Err()}
				return
			default:
			}
			q.mu.Lock()

			// If still empty, wait on condition
			if len(q.items) == 0 {
				q.notEmpty.Wait()
			}
		}

		item := q.getItem()
		done <- result{item: item, err: nil}
	}()

	select {
	case res := <-done:
		return res.item, res.err

	case <-ctx.Done():
		close(cancelled)
		q.notEmpty.Broadcast()
		// Wait for goroutine to finish
		res := <-done
		return res.item, res.err
	}
}

// TaskDone indicates that a formerly enqueued task is complete.
//
// Used by queue consumer threads. For each Get() used to fetch a task,
// a subsequent call to TaskDone() tells the queue that the processing on the task is complete.
//
// If a Join() is currently blocking, it will resume when all items have been processed
// (meaning that a TaskDone() call was received for every item that had been Put() into the queue).
//
// This is equivalent to Python's [asyncio.Queue.task_done] method.
//
// [asyncio.Queue.task_done]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue.task_done
func (q *queue[T]) TaskDone() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.unfinished <= 0 {
		return fmt.Errorf("task_done() called too many times")
	}

	q.unfinished--
	if q.unfinished == 0 {
		q.allTasksDone.Broadcast() // Wake up all waiting Join() calls
	}

	return nil
}

// Join blocks until all items in the queue have been gotten and processed.
//
// The count of unfinished tasks goes up whenever an item is added to the queue.
// The count goes down whenever a consumer calls TaskDone() to indicate that
// the item was retrieved and all work on it is complete.
// When the count of unfinished tasks drops to zero, Join() unblocks.
//
// This is equivalent to Python's [asyncio.Queue.join] method.
//
// [asyncio.Queue.join]: https://docs.python.org/3/library/asyncio-queue.html#asyncio.Queue.join
func (q *queue[T]) Join(ctx context.Context) error {
	// Use a channel to coordinate between the waiting goroutine and context cancellation
	done := make(chan error, 1)
	cancelled := make(chan struct{})

	go func() {
		q.mu.Lock()
		defer q.mu.Unlock()

		// Wait until all tasks are done or context is cancelled
		for q.unfinished > 0 {
			// Release lock temporarily to check for cancellation
			q.mu.Unlock()
			select {
			case <-cancelled:
				q.mu.Lock()
				done <- ctx.Err()
				return
			default:
			}
			q.mu.Lock()

			// If still unfinished tasks, wait on condition
			if q.unfinished > 0 {
				q.allTasksDone.Wait()
			}
		}

		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		close(cancelled)
		q.allTasksDone.Broadcast()
		// Wait for goroutine to finish
		<-done
		return ctx.Err()
	}
}

// Close closes the queue, preventing further operations.
// Any blocked Put/Get operations will be unblocked and return an error.
func (q *queue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.closed = true
	q.notEmpty.Broadcast()
	q.notFull.Broadcast()
	q.allTasksDone.Broadcast()
}
