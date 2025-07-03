// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package pool provides strongly-typed object pooling with generic support and predefined pools for common types.
//
// The pool package implements a type-safe wrapper around Go's sync.Pool, providing performance
// optimizations through object reuse while maintaining compile-time type safety. It includes
// predefined pools for commonly used types like bytes.Buffer and strings.Builder, and supports
// creating custom pools for any type.
//
// # Object Pooling Benefits
//
// Object pooling provides significant performance benefits by:
//
//   - Reducing garbage collection pressure
//   - Minimizing memory allocations
//   - Improving throughput in high-frequency operations
//   - Reusing expensive-to-create objects
//   - Maintaining memory locality for frequently used objects
//
// # Generic Pool Implementation
//
// The core Pool[T] type provides strongly-typed object pooling:
//
//	type Pool[T any] struct {
//		// Internal sync.Pool implementation
//	}
//
// This design ensures compile-time type safety while providing the performance
// benefits of sync.Pool underneath.
//
// # Basic Usage
//
// ## Creating Custom Pools
//
// Create pools for any type with a constructor function:
//
//	// Pool for custom structs
//	type MyStruct struct {
//		Data []byte
//		ID   int
//	}
//
//	myPool := pool.New(func() *MyStruct {
//		return &MyStruct{
//			Data: make([]byte, 0, 1024), // Pre-allocate capacity
//		}
//	})
//
//	// Pool for slices
//	slicePool := pool.New(func() []string {
//		return make([]string, 0, 10) // Pre-allocate capacity
//	})
//
//	// Pool for maps
//	mapPool := pool.New(func() map[string]int {
//		return make(map[string]int)
//	})
//
// ## Using Pools
//
// Get and return objects using Get() and Put():
//
//	// Get object from pool
//	obj := myPool.Get()
//
//	// Use the object
//	obj.ID = 42
//	obj.Data = append(obj.Data, []byte("hello")...)
//
//	// Reset object state before returning to pool
//	obj.ID = 0
//	obj.Data = obj.Data[:0] // Reset slice but keep capacity
//
//	// Return to pool for reuse
//	myPool.Put(obj)
//
// # Predefined Pools
//
// The package provides ready-to-use pools for common types:
//
// ## Buffer Pool
//
// For bytes.Buffer operations:
//
//	// Get a buffer from the pool
//	buf := pool.Buffer.Get()
//
//	// Use the buffer
//	buf.WriteString("Hello, ")
//	buf.WriteString("World!")
//	result := buf.String()
//
//	// Reset and return to pool
//	buf.Reset()
//	pool.Buffer.Put(buf)
//
// ## String Builder Pool
//
// For strings.Builder operations:
//
//	// Get a string builder from the pool
//	sb := pool.String.Get()
//
//	// Use the builder
//	sb.WriteString("Processing ")
//	sb.WriteString("data...")
//	result := sb.String()
//
//	// Reset and return to pool
//	sb.Reset()
//	pool.String.Put(sb)
//
// # Advanced Usage Patterns
//
// ## With Defer for Automatic Cleanup
//
// Ensure objects are always returned to the pool:
//
//	func processData(data []byte) string {
//		buf := pool.Buffer.Get()
//		defer func() {
//			buf.Reset()
//			pool.Buffer.Put(buf)
//		}()
//
//		// Process data with buffer
//		buf.Write(data)
//		buf.WriteString(" processed")
//		return buf.String()
//	}
//
// ## Custom Reset Logic
//
// For complex objects that need custom cleanup:
//
//	type Connection struct {
//		conn   net.Conn
//		buffer []byte
//		state  int
//	}
//
//	func (c *Connection) Reset() {
//		if c.conn != nil {
//			c.conn.Close()
//			c.conn = nil
//		}
//		c.buffer = c.buffer[:0] // Reset slice but keep capacity
//		c.state = 0
//	}
//
//	connPool := pool.New(func() *Connection {
//		return &Connection{
//			buffer: make([]byte, 0, 4096),
//		}
//	})
//
//	// Usage with custom reset
//	conn := connPool.Get()
//	defer func() {
//		conn.Reset()
//		connPool.Put(conn)
//	}()
//
// # Performance Considerations
//
// ## Optimal Constructor Functions
//
// Constructor functions should create objects with appropriate initial capacity:
//
//	// Good: Pre-allocate reasonable capacity
//	slicePool := pool.New(func() []string {
//		return make([]string, 0, 100) // Expect ~100 elements
//	})
//
//	// Good: Pre-allocate map capacity
//	mapPool := pool.New(func() map[string]int {
//		return make(map[string]int, 50) // Expect ~50 key-value pairs
//	})
//
//	// Avoid: No initial capacity (causes reallocations)
//	badPool := pool.New(func() []string {
//		return []string{} // Will grow from zero capacity
//	})
//
// ## Reset Strategies
//
// Proper reset logic is crucial for pool effectiveness:
//
//	// Good: Reset slice but preserve capacity
//	slice = slice[:0]
//
//	// Good: Clear map but preserve capacity (Go 1.11+)
//	for k := range m {
//		delete(m, k)
//	}
//
//	// Good: Reset buffer
//	buffer.Reset()
//
//	// Avoid: Creating new objects
//	slice = make([]string, 0) // Loses capacity benefit
//
// # Integration with ADK Components
//
// The pool package is used throughout the ADK for performance optimization:
//
// ## In LLM Request Processing
//
//	// Used for building request content
//	func buildLLMRequest(parts []string) *types.LLMRequest {
//		buf := pool.Buffer.Get()
//		defer func() {
//			buf.Reset()
//			pool.Buffer.Put(buf)
//		}()
//
//		for _, part := range parts {
//			buf.WriteString(part)
//			buf.WriteString("\n")
//		}
//
//		return &types.LLMRequest{
//			Content: buf.String(),
//		}
//	}
//
// ## In Response Processing
//
//	// Used for accumulating streaming responses
//	func processStreamingResponse(events <-chan *types.Event) string {
//		sb := pool.String.Get()
//		defer func() {
//			sb.Reset()
//			pool.String.Put(sb)
//		}()
//
//		for event := range events {
//			if event.TextDelta != "" {
//				sb.WriteString(event.TextDelta)
//			}
//		}
//
//		return sb.String()
//	}
//
// ## In JSON Processing
//
//	// Used for building JSON responses
//	func marshalResponse(data interface{}) ([]byte, error) {
//		buf := pool.Buffer.Get()
//		defer func() {
//			buf.Reset()
//			pool.Buffer.Put(buf)
//		}()
//
//		encoder := json.NewEncoder(buf)
//		if err := encoder.Encode(data); err != nil {
//			return nil, err
//		}
//
//		// Return copy since buffer will be reset
//		result := make([]byte, buf.Len())
//		copy(result, buf.Bytes())
//		return result, nil
//	}
//
// # Thread Safety
//
// All pool operations are thread-safe and can be used concurrently:
//
//	// Safe to use from multiple goroutines
//	go func() {
//		buf := pool.Buffer.Get()
//		defer func() {
//			buf.Reset()
//			pool.Buffer.Put(buf)
//		}()
//		// Use buffer...
//	}()
//
//	go func() {
//		sb := pool.String.Get()
//		defer func() {
//			sb.Reset()
//			pool.String.Put(sb)
//		}()
//		// Use string builder...
//	}()
//
// # Best Practices
//
//  1. Always reset objects before returning them to the pool
//  2. Use defer statements to ensure objects are returned even on panic
//  3. Pre-allocate appropriate capacity in constructor functions
//  4. Don't hold references to pooled objects after putting them back
//  5. Reset slices/maps by clearing contents, not reallocating
//  6. Consider object lifecycle - don't pool very short-lived objects
//  7. Profile your application to verify pooling benefits
//  8. Use pools for frequently allocated objects with measurable overhead
//
// # Common Pitfalls
//
// ## Holding References After Put
//
//	// WRONG: Using object after returning to pool
//	buf := pool.Buffer.Get()
//	buf.WriteString("data")
//	pool.Buffer.Put(buf)
//	result := buf.String() // BUG: buf may be reused by another goroutine
//
//	// CORRECT: Extract data before returning to pool
//	buf := pool.Buffer.Get()
//	buf.WriteString("data")
//	result := buf.String()
//	buf.Reset()
//	pool.Buffer.Put(buf)
//
// ## Incomplete Reset
//
//	// WRONG: Not resetting object state
//	obj := myPool.Get()
//	obj.SomeField = "data"
//	myPool.Put(obj) // Next user gets dirty object
//
//	// CORRECT: Reset all fields
//	obj := myPool.Get()
//	obj.SomeField = "data"
//	obj.SomeField = "" // Reset state
//	myPool.Put(obj)
//
// ## Pooling Everything
//
//	// WRONG: Pooling very simple, short-lived objects
//	intPool := pool.New(func() int { return 0 }) // Overhead > benefit
//
//	// CORRECT: Pool complex or frequently allocated objects
//	complexPool := pool.New(func() *ComplexStruct {
//		return &ComplexStruct{
//			LargeSlice: make([]byte, 0, 4096),
//			// ... other expensive fields
//		}
//	})
//
// # Performance Impact
//
// The pool package provides measurable performance improvements:
//
//   - Reduced allocation rate in high-throughput scenarios
//   - Lower garbage collection overhead
//   - Improved memory locality for frequently used objects
//   - Better cache utilization through object reuse
//   - Reduced time spent in memory allocation/deallocation
//
// # Benchmarking
//
// Always benchmark to verify pooling benefits:
//
//	func BenchmarkWithPool(b *testing.B) {
//		for i := 0; i < b.N; i++ {
//			buf := pool.Buffer.Get()
//			buf.WriteString("test data")
//			result := buf.String()
//			buf.Reset()
//			pool.Buffer.Put(buf)
//			_ = result
//		}
//	}
//
//	func BenchmarkWithoutPool(b *testing.B) {
//		for i := 0; i < b.N; i++ {
//			var buf bytes.Buffer
//			buf.WriteString("test data")
//			result := buf.String()
//			_ = result
//		}
//	}
//
// # When to Use Pools
//
// Use object pools when:
//
//   - Objects are expensive to create (large allocations, complex initialization)
//   - Objects are allocated frequently in hot paths
//   - Allocation pressure is causing GC performance issues
//   - Objects have significant setup/teardown costs
//   - Memory locality benefits are important
//
// Don't use pools when:
//
//   - Objects are very simple (primitive types, small structs)
//   - Allocation frequency is low
//   - Object lifetime is very short
//   - Reset logic is complex or error-prone
//   - Memory usage patterns don't benefit from pooling
//
// The pool package provides essential performance optimization tools for building
// high-performance Go applications with efficient memory management.
package pool
