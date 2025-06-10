// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package xiter

import (
	"iter"
)

// Error returns an iterator that yields an error at the end of the iteration.
func Error[T any](err error) iter.Seq2[*T, error] {
	return func(yield func(*T, error) bool) {
		yield(nil, err)
	}
}

// EndError returns an iterator that yields an error at the end of the iteration.
func EndError[T any](err error) iter.Seq2[*T, error] {
	return func(yield func(*T, error) bool) {
		if !yield(nil, err) {
			return
		}
	}
}
