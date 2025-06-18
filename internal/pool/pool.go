// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package pool provides generic type pooling, and provides [*bytes.Buffer] and [*strings.Builder] pooling objects.
package pool

import (
	"bytes"
	"strings"
	"sync"
)

// Pool is a generics wrapper around [syncx.Pool] to provide strongly-typed object pooling.
type Pool[T any] struct {
	pool sync.Pool
}

// New returns a new [Pool] for T, and will use fn to construct new T's when the pool is empty.
func New[T any](fn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return fn()
			},
		},
	}
}

// Get gets a T from the pool, or creates a new one if the pool is empty.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns x into the pool.
func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}

// Buffer provides the [*bytes.Buffer] pooling objects.
var Buffer = New(func() *bytes.Buffer {
	return &bytes.Buffer{}
})

// String provides the [*strings.Builder] pooling objects.
var String = New(func() *strings.Builder {
	return &strings.Builder{}
})
