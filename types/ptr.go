// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

// ToPtr returns a pointer to the given value.
func ToPtr[T any](v T) *T {
	return &v
}

// Deref dereferences ptr and returns the value it points to if no nil, or else returns def.
func Deref[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
