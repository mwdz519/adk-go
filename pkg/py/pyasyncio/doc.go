// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package pyasyncio provides Go implementations of Python's asyncio concepts for concurrent programming patterns.
//
// The pyasyncio package brings Python's powerful asyncio primitives to Go, enabling familiar
// concurrent programming patterns while leveraging Go's native concurrency features. It provides
// Task management, Queue operations, and other async patterns that mirror Python's asyncio module.
//
// # Supported Asyncio Patterns
//
// The package implements core asyncio concepts:
//   - Task[T]: Asynchronous task execution with lifecycle management
//   - Queue[T]: Producer-consumer queues with blocking operations
//   - TaskGroup: Coordinated task execution (via task_group.go)
//   - WaitFor: Timeout and cancellation utilities (via wait_for.go)
//
// # Task Implementation
//
// Tasks represent concurrent operations that can be cancelled, monitored, and coordinated:
//
//	type Task[T any] struct {
//		// Lifecycle: pending -> running -> done/cancelled
//		// Supports cancellation, callbacks, and result retrieval
//	}
//
// Tasks map Python's asyncio.Task to Go's goroutines with additional control features.
//
// # Basic Task Usage
//
// ## Creating and Running Tasks
//
// Create tasks that execute immediately in the background:
//
//	// Simple task creation
//	task := pyasyncio.CreateTask(ctx, func(ctx context.Context) (string, error) {
//		// Simulate some work
//		time.Sleep(1 * time.Second)
//		return "task completed", nil
//	})
//
//	// Named task for debugging
//	namedTask := pyasyncio.CreateNamedTask(ctx, "data_processor", func(ctx context.Context) ([]Data, error) {
//		return processLargeDataset(ctx)
//	})
//
//	// Wait for completion
//	result, err := task.Wait(ctx)
//	if err != nil {
//		log.Printf("Task failed: %v", err)
//	} else {
//		fmt.Printf("Result: %s\n", result)
//	}
//
// ## Task Lifecycle Management
//
// Monitor and control task execution:
//
//	// Check task state
//	fmt.Printf("Task state: %s\n", task.State()) // pending, running, done, cancelled
//
//	// Check completion without blocking
//	if task.Done() {
//		result, err := task.Result() // Get result immediately
//		if err != nil {
//			fmt.Printf("Task error: %v\n", err)
//		}
//	}
//
//	// Cancel task if running too long
//	if task.Cancel() {
//		fmt.Println("Task cancellation requested")
//	}
//
//	// Check if task was cancelled
//	if task.Cancelled() {
//		fmt.Println("Task was cancelled")
//	}
//
// ## Task Callbacks
//
// Execute callbacks when tasks complete:
//
//	task.AddCallback(func(t *pyasyncio.Task[string]) {
//		if t.Cancelled() {
//			fmt.Println("Task was cancelled")
//		} else if err := t.Exception(); err != nil {
//			fmt.Printf("Task failed: %v\n", err)
//		} else {
//			result, _ := t.Result()
//			fmt.Printf("Task completed: %s\n", result)
//		}
//	})
//
// # Queue Implementation
//
// Queues provide producer-consumer coordination with blocking operations:
//
//	type Queue[T any] interface {
//		Put(ctx context.Context, item T) error      // Blocking put
//		PutNowait(item T) error                     // Non-blocking put
//		Get(ctx context.Context) (T, error)        // Blocking get
//		GetNowait() (T, error)                      // Non-blocking get
//		TaskDone() error                            // Mark task complete
//		Join(ctx context.Context) error             // Wait for all tasks
//	}
//
// # Basic Queue Usage
//
// ## Producer-Consumer Pattern
//
// Coordinate between producers and consumers:
//
//	// Create queue with maximum size
//	queue := pyasyncio.NewQueue[string](10) // Max 10 items
//
//	// Producer goroutine
//	go func() {
//		defer close(done)
//		for i := 0; i < 100; i++ {
//			item := fmt.Sprintf("item-%d", i)
//
//			// Blocks if queue is full
//			if err := queue.Put(ctx, item); err != nil {
//				log.Printf("Failed to put item: %v", err)
//				return
//			}
//		}
//	}()
//
//	// Consumer goroutine
//	go func() {
//		for {
//			// Blocks if queue is empty
//			item, err := queue.Get(ctx)
//			if err != nil {
//				if errors.Is(err, context.Canceled) {
//					return // Context cancelled
//				}
//				log.Printf("Failed to get item: %v", err)
//				continue
//			}
//
//			// Process item
//			processItem(item)
//
//			// Mark task as done
//			queue.TaskDone()
//		}
//	}()
//
//	// Wait for all items to be processed
//	if err := queue.Join(ctx); err != nil {
//		log.Printf("Queue join failed: %v", err)
//	}
//
// ## Non-blocking Operations
//
// Use non-blocking operations when you can't wait:
//
//	// Try to put without blocking
//	if err := queue.PutNowait("urgent_item"); err != nil {
//		if _, ok := err.(*pyasyncio.ErrQueueFull); ok {
//			fmt.Println("Queue is full, dropping item")
//		}
//	}
//
//	// Try to get without blocking
//	item, err := queue.GetNowait()
//	if err != nil {
//		if _, ok := err.(*pyasyncio.ErrQueueEmpty); ok {
//			fmt.Println("Queue is empty, nothing to process")
//		}
//	} else {
//		processItem(item)
//		queue.TaskDone()
//	}
//
// ## Queue State Monitoring
//
// Monitor queue status and capacity:
//
//	fmt.Printf("Queue size: %d\n", queue.Size())
//	fmt.Printf("Queue empty: %t\n", queue.Empty())
//	fmt.Printf("Queue full: %t\n", queue.Full())
//
// # Advanced Task Patterns
//
// ## Parallel Task Execution
//
// Run multiple tasks concurrently and collect results:
//
//	// Start multiple tasks
//	var tasks []*pyasyncio.Task[int]
//	for i := 0; i < 10; i++ {
//		i := i // Capture loop variable
//		task := pyasyncio.CreateNamedTask(ctx, fmt.Sprintf("worker-%d", i),
//			func(ctx context.Context) (int, error) {
//				return processWorkItem(ctx, i)
//			})
//		tasks = append(tasks, task)
//	}
//
//	// Collect all results
//	var results []int
//	for _, task := range tasks {
//		result, err := task.Wait(ctx)
//		if err != nil {
//			log.Printf("Task %s failed: %v", task.Name(), err)
//			continue
//		}
//		results = append(results, result)
//	}
//
// ## Task Cancellation Patterns
//
// Implement timeout and cancellation logic:
//
//	// Create task with timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	task := pyasyncio.CreateTask(ctx, func(ctx context.Context) (string, error) {
//		// Long-running operation that respects context
//		select {
//		case <-time.After(10 * time.Second):
//			return "completed", nil
//		case <-ctx.Done():
//			return "", ctx.Err()
//		}
//	})
//
//	result, err := task.Wait(ctx)
//	if err != nil {
//		if task.Cancelled() {
//			fmt.Println("Task was cancelled due to timeout")
//		} else {
//			fmt.Printf("Task failed: %v\n", err)
//		}
//	}
//
// ## Graceful Shutdown Pattern
//
// Coordinate clean shutdown of multiple tasks:
//
//	func gracefulShutdown(tasks []*pyasyncio.Task[any]) {
//		// Request cancellation for all tasks
//		for _, task := range tasks {
//			task.Cancel()
//		}
//
//		// Wait for all tasks to finish (with timeout)
//		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//		defer cancel()
//
//		for _, task := range tasks {
//			_, err := task.Wait(shutdownCtx)
//			if err != nil && !task.Cancelled() {
//				log.Printf("Task %s failed during shutdown: %v", task.Name(), err)
//			}
//		}
//	}
//
// # Queue Coordination Patterns
//
// ## Worker Pool Pattern
//
// Implement a worker pool using queues:
//
//	func startWorkerPool[T any](ctx context.Context, numWorkers int, processor func(T) error) (*pyasyncio.Queue[T], error) {
//		queue := pyasyncio.NewQueue[T](100) // Buffer size
//
//		// Start worker goroutines
//		for i := 0; i < numWorkers; i++ {
//			go func(workerID int) {
//				for {
//					select {
//					case <-ctx.Done():
//						return
//					default:
//						item, err := queue.Get(ctx)
//						if err != nil {
//							if errors.Is(err, context.Canceled) {
//								return
//							}
//							log.Printf("Worker %d: get failed: %v", workerID, err)
//							continue
//						}
//
//						if err := processor(item); err != nil {
//							log.Printf("Worker %d: processing failed: %v", workerID, err)
//						}
//
//						queue.TaskDone()
//					}
//				}
//			}(i)
//		}
//
//		return queue, nil
//	}
//
// ## Pipeline Pattern
//
// Chain multiple processing stages:
//
//	func createPipeline[T, U, V any](
//		stage1 func(T) U,
//		stage2 func(U) V,
//	) (input *pyasyncio.Queue[T], output *pyasyncio.Queue[V]) {
//
//		input = pyasyncio.NewQueue[T](10)
//		intermediate := pyasyncio.NewQueue[U](10)
//		output = pyasyncio.NewQueue[V](10)
//
//		// Stage 1: T -> U
//		go func() {
//			for {
//				item, err := input.Get(context.Background())
//				if err != nil {
//					break
//				}
//
//				result := stage1(item)
//				intermediate.Put(context.Background(), result)
//				input.TaskDone()
//			}
//		}()
//
//		// Stage 2: U -> V
//		go func() {
//			for {
//				item, err := intermediate.Get(context.Background())
//				if err != nil {
//					break
//				}
//
//				result := stage2(item)
//				output.Put(context.Background(), result)
//				intermediate.TaskDone()
//			}
//		}()
//
//		return input, output
//	}
//
// # Integration with ADK Framework
//
// ## Agent Task Coordination
//
// Use tasks for coordinating agent operations:
//
//	// Parallel agent execution
//	func runAgentsInParallel(ctx context.Context, agents []types.Agent, ictx *types.InvocationContext) map[string][]types.Event {
//		var tasks []*pyasyncio.Task[[]types.Event]
//
//		for _, agent := range agents {
//			agent := agent // Capture loop variable
//			task := pyasyncio.CreateNamedTask(ctx, agent.Name(),
//				func(ctx context.Context) ([]types.Event, error) {
//					var events []types.Event
//					for event, err := range agent.Run(ctx, ictx) {
//						if err != nil {
//							return nil, err
//						}
//						events = append(events, *event)
//					}
//					return events, nil
//				})
//			tasks = append(tasks, task)
//		}
//
//		// Collect results
//		results := make(map[string][]types.Event)
//		for i, task := range tasks {
//			events, err := task.Wait(ctx)
//			if err != nil {
//				log.Printf("Agent %s failed: %v", agents[i].Name(), err)
//				continue
//			}
//			results[agents[i].Name()] = events
//		}
//
//		return results
//	}
//
// ## Tool Execution Queuing
//
// Queue tool execution requests for controlled processing:
//
//	type ToolRequest struct {
//		Tool    types.Tool
//		Args    map[string]any
//		Context *types.ToolContext
//		Result  chan ToolResult
//	}
//
//	type ToolResult struct {
//		Value any
//		Error error
//	}
//
//	func startToolExecutionService(ctx context.Context) *pyasyncio.Queue[ToolRequest] {
//		queue := pyasyncio.NewQueue[ToolRequest](50)
//
//		// Start tool execution workers
//		for i := 0; i < 5; i++ {
//			go func() {
//				for {
//					req, err := queue.Get(ctx)
//					if err != nil {
//						return
//					}
//
//					// Execute tool
//					result, err := req.Tool.Run(ctx, req.Args, req.Context)
//
//					// Send result back
//					req.Result <- ToolResult{Value: result, Error: err}
//					close(req.Result)
//
//					queue.TaskDone()
//				}
//			}()
//		}
//
//		return queue
//	}
//
// # Error Handling and Recovery
//
// ## Task Error Recovery
//
// Implement retry logic for failed tasks:
//
//	func retryTask[T any](ctx context.Context, maxRetries int, taskFn func(context.Context) (T, error)) (T, error) {
//		var lastErr error
//		var zero T
//
//		for attempt := 0; attempt <= maxRetries; attempt++ {
//			task := pyasyncio.CreateTask(ctx, taskFn)
//			result, err := task.Wait(ctx)
//
//			if err == nil {
//				return result, nil
//			}
//
//			lastErr = err
//			if attempt < maxRetries {
//				// Exponential backoff
//				delay := time.Duration(1<<attempt) * time.Second
//				time.Sleep(delay)
//			}
//		}
//
//		return zero, fmt.Errorf("task failed after %d retries: %w", maxRetries, lastErr)
//	}
//
// ## Queue Error Handling
//
// Handle queue errors gracefully:
//
//	func robustQueueOperation[T any](queue *pyasyncio.Queue[T], item T, timeout time.Duration) error {
//		ctx, cancel := context.WithTimeout(context.Background(), timeout)
//		defer cancel()
//
//		// Try non-blocking first
//		if err := queue.PutNowait(item); err == nil {
//			return nil
//		}
//
//		// Fall back to blocking with timeout
//		return queue.Put(ctx, item)
//	}
//
// # Performance Considerations
//
// ## Task Performance
//
//   - Task creation: O(1) with minimal overhead
//   - Task switching: Leverages Go's efficient goroutine scheduler
//   - Memory usage: ~2KB per goroutine (Go's stack)
//   - Cancellation: Cooperative, requires context checking
//
// ## Queue Performance
//
//   - Put/Get operations: O(1) average case
//   - Memory usage: O(n) where n is queue size
//   - Blocking behavior: Uses Go's condition variables
//   - Throughput: High with proper buffer sizing
//
// # Thread Safety
//
// All pyasyncio types are safe for concurrent use:
//   - Tasks can be safely accessed from multiple goroutines
//   - Queues handle concurrent producers and consumers
//   - State changes are protected by internal synchronization
//   - Callbacks execute concurrently without blocking the task
//
// # Python Asyncio Compatibility
//
// The implementation maintains API compatibility with Python's asyncio:
//
//	# Python asyncio
//	import asyncio
//
//	async def my_task():
//		await asyncio.sleep(1)
//		return "done"
//
//	task = asyncio.create_task(my_task())
//	result = await task
//
//	# Go equivalent
//	task := pyasyncio.CreateTask(ctx, func(ctx context.Context) (string, error) {
//		time.Sleep(1 * time.Second)
//		return "done", nil
//	})
//	result, err := task.Wait(ctx)
//
// # Best Practices
//
//  1. Always check context cancellation in long-running task functions
//  2. Use named tasks for better debugging and monitoring
//  3. Set appropriate queue sizes to balance memory and blocking behavior
//  4. Call TaskDone() for every Get() operation on queues
//  5. Handle cancellation gracefully in task functions
//  6. Use timeouts to prevent indefinite blocking
//  7. Monitor task states for debugging and observability
//  8. Clean up resources in task completion callbacks
//
// # Debugging and Monitoring
//
// Tasks provide metadata for debugging:
//
//	fmt.Printf("Task: %s\n", task.Name())
//	fmt.Printf("Created: %v\n", task.Created())
//	fmt.Printf("Started: %v\n", task.Started())
//	fmt.Printf("Finished: %v\n", task.Finished())
//	fmt.Printf("State: %s\n", task.State())
//
// # Future Extensions
//
// The package is designed for extensibility:
//   - Additional asyncio primitives (Event, Condition, etc.)
//   - More sophisticated task scheduling
//   - Task priority and resource management
//   - Integration with distributed systems patterns
//   - Enhanced observability and metrics
//
// The pyasyncio package provides powerful concurrent programming primitives that
// bridge Python's asyncio patterns with Go's native concurrency features.
package pyasyncio
