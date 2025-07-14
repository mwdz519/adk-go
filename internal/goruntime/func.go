// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package goruntime

import (
	"unsafe"
)

// Name returns the function name for the given pc.
func Name(pc uintptr) string {
	f := findfunc(pc)
	if f._func == nil {
		return ""
	}

	str := &f.datap.funcnametab[f.nameOff]
	ss := stringStruct{str: unsafe.Pointer(str), len: findnull(str)}
	return *(*string)(unsafe.Pointer(&ss))
}
