// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package py provides Go implementations of Python's core data structures and patterns for seamless interoperability.
//
// The py package bridges the gap between Python and Go by providing native Go implementations
// of common Python data structures and patterns. This enables code that needs to maintain
// Python-like semantics while benefiting from Go's performance and type safety.
//
// # Supported Python Patterns
//
// The package currently implements:
//   - Set[T]: Python-style sets with comprehensive set operations
//   - Additional Python patterns via subpackages (pyasyncio)
//
// # Set Implementation
//
// The primary feature of this package is a memory-efficient, type-safe implementation
// of Python's set data structure:
//
//	type Set[T comparable] map[T]Empty
//
// This design uses Go's map implementation for O(1) average-case operations while
// maintaining minimal memory overhead through the Empty struct.
//
// # Basic Set Usage
//
// ## Creating Sets
//
// Create sets from values or other data structures:
//
//	// Create set from values
//	numbers := py.NewSet(1, 2, 3, 4, 5)
//	names := py.NewSet("alice", "bob", "charlie")
//
//	// Create empty set (explicit type required)
//	empty := py.NewSet[string]()
//
//	// Create set from map keys
//	userRoles := map[string]string{
//		"alice": "admin",
//		"bob":   "user",
//		"charlie": "guest",
//	}
//	users := py.KeySet(userRoles) // Set[string]{"alice", "bob", "charlie"}
//
// ## Set Operations
//
// Comprehensive set manipulation operations:
//
//	set1 := py.NewSet(1, 2, 3, 4)
//	set2 := py.NewSet(3, 4, 5, 6)
//
//	// Basic operations
//	set1.Insert(7, 8)              // Add elements
//	set1.Delete(1)                 // Remove elements
//	set1.Clear()                   // Remove all elements
//
//	// Membership testing
//	exists := set1.Has(3)          // true if 3 is in set
//	hasAll := set1.HasAll(2, 3)    // true if all elements are in set
//	hasAny := set1.HasAny(5, 6)    // true if any element is in set
//
//	// Set math operations
//	union := set1.Union(set2)           // {1, 2, 3, 4, 5, 6}
//	intersection := set1.Intersection(set2)  // {3, 4}
//	difference := set1.Difference(set2)      // {1, 2}
//	symmetric := set1.SymmetricDifference(set2) // {1, 2, 5, 6}
//
// ## Set Relationships
//
// Test relationships between sets:
//
//	smallSet := py.NewSet(2, 3)
//	largeSet := py.NewSet(1, 2, 3, 4, 5)
//
//	// Test if sets are equal
//	equal := smallSet.Equal(largeSet) // false
//
//	// Test subset/superset relationships
//	isSuperset := largeSet.IsSuperset(smallSet) // true
//	isSubset := smallSet.IsSuperset(largeSet)   // false (reverse test)
//
// # Data Conversion
//
// ## Converting to Slices
//
// Extract set contents as slices:
//
//	numbers := py.NewSet(3, 1, 4, 1, 5, 9, 2, 6)
//
//	// Sorted slice (requires comparable elements)
//	sorted := py.List(numbers) // []int{1, 2, 3, 4, 5, 6, 9}
//
//	// Unsorted slice (random order)
//	unsorted := numbers.UnsortedList() // []int{3, 1, 4, 5, 9, 2, 6} (order varies)
//
// ## Pop Operations
//
// Remove and return arbitrary elements:
//
//	element, ok := numbers.PopAny()
//	if ok {
//		fmt.Printf("Removed: %d\n", element)
//		fmt.Printf("Remaining size: %d\n", numbers.Len())
//	}
//
// # Memory Efficiency
//
// The set implementation is optimized for memory efficiency:
//
//	// Empty struct uses zero bytes
//	type Empty struct{}
//
//	// Set uses map with empty values for minimal memory
//	type Set[T comparable] map[T]Empty
//
// This design provides:
//   - Zero additional memory per element beyond the key
//   - Fast O(1) average-case operations (insert, delete, lookup)
//   - Excellent cache locality for small sets
//   - No boxing/unboxing overhead with generics
//
// # Thread Safety
//
// Sets are NOT thread-safe by default. For concurrent access, use external synchronization:
//
//	var mu sync.RWMutex
//	var sharedSet = py.NewSet[string]()
//
//	// Safe concurrent read
//	func safeRead(key string) bool {
//		mu.RLock()
//		defer mu.RUnlock()
//		return sharedSet.Has(key)
//	}
//
//	// Safe concurrent write
//	func safeWrite(key string) {
//		mu.Lock()
//		defer mu.Unlock()
//		sharedSet.Insert(key)
//	}
//
// # Performance Characteristics
//
// ## Time Complexity
//
//   - Insert/Delete/Has: O(1) average case, O(n) worst case
//   - Union/Intersection/Difference: O(n + m) where n, m are set sizes
//   - Equal/IsSuperset: O(min(n, m))
//   - Clone: O(n)
//   - List (sorted): O(n log n)
//   - UnsortedList: O(n)
//
// ## Space Complexity
//
//   - Storage: O(n) where n is number of elements
//   - Additional overhead: ~24 bytes per map + 0 bytes per element
//
// # Integration with ADK Framework
//
// Sets are used throughout the ADK for efficient membership testing and deduplication:
//
// ## Event Processing
//
//	// Track processed event IDs to avoid duplicates
//	processedEvents := py.NewSet[string]()
//
//	for event, err := range agent.Run(ctx, ictx) {
//		if err != nil {
//			continue
//		}
//
//		eventID := event.ID
//		if processedEvents.Has(eventID) {
//			continue // Skip duplicate
//		}
//		processedEvents.Insert(eventID)
//
//		// Process event
//		handleEvent(event)
//	}
//
// ## Tool Validation
//
//	// Required parameters for tool execution
//	requiredParams := py.NewSet("model", "prompt", "temperature")
//	providedParams := py.KeySet(toolArgs)
//
//	// Check if all required parameters are provided
//	if !providedParams.IsSuperset(requiredParams) {
//		missing := requiredParams.Difference(providedParams)
//		return fmt.Errorf("missing required parameters: %v", missing.UnsortedList())
//	}
//
// ## Agent Coordination
//
//	// Track which agents have completed their tasks
//	completedAgents := py.NewSet[string]()
//	requiredAgents := py.NewSet("researcher", "analyzer", "reporter")
//
//	for _, agent := range agentResults {
//		if agent.Success {
//			completedAgents.Insert(agent.Name)
//		}
//	}
//
//	// Check if all required agents completed successfully
//	allCompleted := completedAgents.Equal(requiredAgents)
//
// # Python Set Compatibility
//
// The implementation maintains compatibility with Python set semantics:
//
//	# Python set operations
//	set1 = {1, 2, 3, 4}
//	set2 = {3, 4, 5, 6}
//
//	union = set1 | set2           # {1, 2, 3, 4, 5, 6}
//	intersection = set1 & set2    # {3, 4}
//	difference = set1 - set2      # {1, 2}
//	symmetric = set1 ^ set2       # {1, 2, 5, 6}
//
//	// Go equivalent
//	set1 := py.NewSet(1, 2, 3, 4)
//	set2 := py.NewSet(3, 4, 5, 6)
//
//	union := set1.Union(set2)               // {1, 2, 3, 4, 5, 6}
//	intersection := set1.Intersection(set2) // {3, 4}
//	difference := set1.Difference(set2)     // {1, 2}
//	symmetric := set1.SymmetricDifference(set2) // {1, 2, 5, 6}
//
// # Best Practices
//
//  1. Use sets for membership testing and deduplication
//  2. Prefer sets over slices when order doesn't matter and uniqueness is important
//  3. Use KeySet() to extract unique keys from maps
//  4. Consider memory usage for very large sets (map overhead)
//  5. Use external synchronization for concurrent access
//  6. Use List() for sorted output, UnsortedList() for faster iteration
//  7. Clone sets before modification when sharing between functions
//
// # Common Patterns
//
// ## Unique Elements from Slice
//
//	// Remove duplicates from slice
//	func unique[T comparable](slice []T) []T {
//		set := py.NewSet(slice...)
//		return set.UnsortedList()
//	}
//
//	// Sorted unique elements
//	func sortedUnique[T cmp.Ordered](slice []T) []T {
//		set := py.NewSet(slice...)
//		return py.List(set)
//	}
//
// ## Set-based Filtering
//
//	// Filter slice using set membership
//	func filterBySet[T comparable](slice []T, allowedValues py.Set[T]) []T {
//		var result []T
//		for _, item := range slice {
//			if allowedValues.Has(item) {
//				result = append(result, item)
//			}
//		}
//		return result
//	}
//
// ## Batch Operations
//
//	// Process items in batches, avoiding duplicates
//	func processBatch[T comparable](items []T, batchSize int) {
//		seen := py.NewSet[T]()
//		var batch []T
//
//		for _, item := range items {
//			if seen.Has(item) {
//				continue // Skip duplicates
//			}
//			seen.Insert(item)
//			batch = append(batch, item)
//
//			if len(batch) >= batchSize {
//				processBatchItems(batch)
//				batch = batch[:0] // Reset batch
//			}
//		}
//
//		// Process remaining items
//		if len(batch) > 0 {
//			processBatchItems(batch)
//		}
//	}
//
// # Attribution
//
// The Set implementation is adapted from Kubernetes' utility library
// (https://github.com/kubernetes/kubernetes/tree/master/staging/src/k8s.io/apimachinery/pkg/util/sets)
// and is used under the Apache 2.0 license. The original copyright is:
//
//	Copyright 2022 The Kubernetes Authors
//
// # Future Extensions
//
// The package is designed for extensibility with additional Python patterns:
//   - dict-like structures with Python semantics
//   - list/tuple equivalents with Python behavior
//   - Additional container types (deque, defaultdict, etc.)
//   - Python-style iteration patterns
//
// The py package provides essential Python compatibility while maintaining Go's
// performance characteristics and type safety.
package py
