// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

//go:build go1.21 && arm64

package goruntime

// Func returns the package/function name of the caller.
func Func() string {
	var pcbuf [1]uintptr
	callers(1, pcbuf[:])
	return Name(pcbuf[0])
}

// FuncN returns the package/function n levels below the caller.
func FuncN(n int) string {
	var pcbuf [1]uintptr
	callers(1+n, pcbuf[:])
	return Name(pcbuf[0])
}
