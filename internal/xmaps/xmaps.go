// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package xmaps

import (
	"cmp"
	"maps"
	"slices"
)

// Contains reports whether key is present in m.
func Contains[Map ~map[K]V, K cmp.Ordered, V any](m Map, key K) bool {
	return slices.Contains(slices.Sorted(maps.Keys(m)), key)
}
