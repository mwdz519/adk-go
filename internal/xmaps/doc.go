// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package xmaps provides extended utility functions for working with maps, complementing the standard maps package.
//
// The xmaps package extends Go's standard maps package with additional utility functions
// that are commonly needed when working with maps in real-world applications. It leverages
// Go's generics support to provide type-safe operations while maintaining performance and
// ease of use.
//
// # Overview
//
// This package provides utility functions that are not available in the standard maps package
// but are frequently needed in applications. All functions are designed to work with generic
// map types and maintain type safety through Go's type constraints.
//
// # Current Functionality
//
// The package currently provides:
//
//   - Contains: Check if a key exists in a map using ordered comparison
//   - Future functions will be added as needed by the ADK framework
//
// # Contains Function
//
// The Contains function provides a way to check if a key exists in a map:
//
//	func Contains[Map ~map[K]V, K cmp.Ordered, V any](m Map, key K) bool
//
// This function works with any map type where the key type implements cmp.Ordered.
//
// ## Basic Usage
//
//	// String keys
//	userMap := map[string]int{
//		"alice": 25,
//		"bob":   30,
//		"charlie": 35,
//	}
//
//	exists := xmaps.Contains(userMap, "alice")  // true
//	missing := xmaps.Contains(userMap, "david") // false
//
//	// Integer keys
//	scoreMap := map[int]string{
//		100: "excellent",
//		85:  "good",
//		70:  "average",
//	}
//
//	hasScore := xmaps.Contains(scoreMap, 85)  // true
//	hasScore = xmaps.Contains(scoreMap, 95)   // false
//
// ## Type Safety
//
// The function is fully type-safe and works with custom map types:
//
//	// Custom map type
//	type UserScores map[string]int
//
//	scores := UserScores{
//		"alice": 95,
//		"bob":   87,
//	}
//
//	// Type-safe usage
//	hasAlice := xmaps.Contains(scores, "alice") // true
//
// ## Ordered Key Constraint
//
// The function requires keys to implement cmp.Ordered, which includes:
//
//   - All integer types (int, int8, int16, int32, int64)
//
//   - All unsigned integer types (uint, uint8, uint16, uint32, uint64, uintptr)
//
//   - All floating-point types (float32, float64)
//
//   - string
//
//   - Custom types based on the above
//
//     // These all work
//     intMap := map[int]string{1: "one", 2: "two"}
//     xmaps.Contains(intMap, 1) // true
//
//     floatMap := map[float64]bool{3.14: true, 2.71: false}
//     xmaps.Contains(floatMap, 3.14) // true
//
//     stringMap := map[string]int{"key": 42}
//     xmaps.Contains(stringMap, "key") // true
//
// # Performance Characteristics
//
// ## Time Complexity
//
// The Contains function has O(n log n) time complexity where n is the number of keys in the map:
//
//   - O(n) to extract all keys using maps.Keys()
//   - O(n log n) to sort the keys using slices.Sorted()
//   - O(log n) to search for the target key using slices.Contains() on sorted slice
//
// ## Space Complexity
//
// The function has O(n) space complexity as it creates a sorted slice of all keys.
//
// ## Alternative for Frequent Lookups
//
// For frequent key existence checks, the standard map lookup is more efficient:
//
//	// For frequent checks, use standard map lookup
//	_, exists := userMap["alice"] // O(1) average case
//
//	// Use xmaps.Contains when you need ordered comparison semantics
//	// or when working with generic code that needs this specific behavior
//
// # Use Cases
//
// ## Generic Map Processing
//
// When writing generic functions that work with any ordered map:
//
//	func ProcessMapWithKey[K cmp.Ordered, V any](m map[K]V, key K) V {
//		if xmaps.Contains(m, key) {
//			return m[key]
//		}
//		var zero V
//		return zero
//	}
//
//	// Works with any ordered key type
//	result1 := ProcessMapWithKey(map[string]int{"a": 1}, "a")     // 1
//	result2 := ProcessMapWithKey(map[int]string{42: "answer"}, 42) // "answer"
//
// ## Validation and Safety
//
// When you need explicit validation before map operations:
//
//	func SafeMapAccess[K cmp.Ordered, V any](m map[K]V, key K) (V, bool) {
//		if !xmaps.Contains(m, key) {
//			var zero V
//			return zero, false
//		}
//		return m[key], true
//	}
//
// ## Integration with Other Operations
//
// When combining with other map operations in a pipeline:
//
//	func FilterMapByKeys[K cmp.Ordered, V any](m map[K]V, validKeys []K) map[K]V {
//		result := make(map[K]V)
//		for _, key := range validKeys {
//			if xmaps.Contains(m, key) {
//				result[key] = m[key]
//			}
//		}
//		return result
//	}
//
// # Best Practices
//
//  1. Use standard map lookup (_, exists := m[key]) for simple existence checks
//  2. Use xmaps.Contains() for generic functions requiring ordered key constraints
//  3. Consider performance implications for large maps with frequent lookups
//  4. Leverage type safety - the compiler will catch type mismatches
//  5. Use appropriate key types that implement cmp.Ordered
//
// # Common Patterns
//
// ## Safe Map Operations
//
//	// Pattern: Safe access with default value
//	func GetOrDefault[K cmp.Ordered, V any](m map[K]V, key K, defaultValue V) V {
//		if xmaps.Contains(m, key) {
//			return m[key]
//		}
//		return defaultValue
//	}
//
//	// Usage
//	score := GetOrDefault(scoreMap, "unknown_user", 0)
//
// ## Batch Key Validation
//
//	// Pattern: Check multiple keys
//	func AllKeysExist[K cmp.Ordered, V any](m map[K]V, keys []K) bool {
//		for _, key := range keys {
//			if !xmaps.Contains(m, key) {
//				return false
//			}
//		}
//		return true
//	}
//
//	// Usage
//	requiredKeys := []string{"name", "email", "age"}
//	valid := AllKeysExist(userDataMap, requiredKeys)
//
// ## Map Intersection
//
//	// Pattern: Find common keys between maps
//	func MapIntersection[K cmp.Ordered, V any](m1, m2 map[K]V) map[K]V {
//		result := make(map[K]V)
//		for k, v := range m1 {
//			if xmaps.Contains(m2, k) {
//				result[k] = v
//			}
//		}
//		return result
//	}
//
// # Error Handling
//
// The Contains function is designed to be safe and never panic:
//
//	// Safe with nil maps
//	var nilMap map[string]int
//	exists := xmaps.Contains(nilMap, "key") // false, no panic
//
//	// Safe with empty maps
//	emptyMap := make(map[string]int)
//	exists = xmaps.Contains(emptyMap, "key") // false
//
// # Integration with ADK Framework
//
// The xmaps package is used throughout the ADK framework for:
//
// ## Configuration Validation
//
//	// Validating required configuration keys
//	func ValidateConfig(config map[string]interface{}) error {
//		required := []string{"api_key", "model", "temperature"}
//		for _, key := range required {
//			if !xmaps.Contains(config, key) {
//				return fmt.Errorf("missing required config key: %s", key)
//			}
//		}
//		return nil
//	}
//
// ## State Management
//
//	// Checking for state keys in agent contexts
//	func HasUserPreference(state map[string]any, key string) bool {
//		prefixedKey := "user:" + key
//		return xmaps.Contains(state, prefixedKey)
//	}
//
// ## Tool Parameter Validation
//
//	// Validating tool parameters
//	func ValidateToolParams[K cmp.Ordered](params map[K]any, required []K) error {
//		for _, param := range required {
//			if !xmaps.Contains(params, param) {
//				return fmt.Errorf("missing required parameter: %v", param)
//			}
//		}
//		return nil
//	}
//
// # Future Extensions
//
// The package is designed to be extended with additional utility functions as needed:
//
//   - Map merging utilities
//   - Key transformation functions
//   - Value filtering operations
//   - Map comparison utilities
//   - Thread-safe map operations
//
// # Thread Safety
//
// The Contains function is read-only and safe for concurrent use on read-only maps.
// However, it's not safe to use concurrently with map modifications:
//
//	// Safe: Concurrent reads
//	go func() { exists := xmaps.Contains(readOnlyMap, "key1") }()
//	go func() { exists := xmaps.Contains(readOnlyMap, "key2") }()
//
//	// Unsafe: Concurrent read/write
//	go func() { readWriteMap["new"] = "value" }()           // Write
//	go func() { exists := xmaps.Contains(readWriteMap, "key") }() // Read - UNSAFE
//
// For concurrent access to maps being modified, use appropriate synchronization:
//
//	var mu sync.RWMutex
//	var sharedMap = make(map[string]int)
//
//	// Safe concurrent access
//	func SafeContains(key string) bool {
//		mu.RLock()
//		defer mu.RUnlock()
//		return xmaps.Contains(sharedMap, key)
//	}
//
// # Limitations
//
//  1. Performance: O(n log n) complexity may be suboptimal for large maps
//  2. Memory: Creates temporary sorted slice of all keys
//  3. Key constraint: Only works with cmp.Ordered types
//  4. Not optimized for frequent lookups on the same map
//
// # When Not to Use
//
//   - For simple key existence checks: use `_, exists := m[key]`
//   - For frequent lookups on large maps: standard map lookup is faster
//   - For non-ordered key types: use standard map operations
//   - For performance-critical hot paths: consider caching or alternative approaches
//
// The xmaps package provides essential utility functions for working with maps in
// generic, type-safe Go code while complementing the standard library's map operations.
package xmaps
