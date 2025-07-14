// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

//go:build go1.13

package goruntime

import (
	"runtime"
	"strings"
	"unsafe"
)

// Callers is a drop-in replacement for runtime.Callers that uses frame
// pointers for fast and simple stack unwinding.
//
// Based by: https://github.com/golang/go/blob/go1.23.0/src/runtime/extern.go#L324-L332
//
//go:noinline
func Callers(skip int, pc []uintptr) int {
	// runtime.callers uses pc.array==nil as a signal
	// to print a stack trace. Pick off 0-length pc here
	// so that we don't let a nil pc slice get to it.
	if len(pc) == 0 {
		return 0
	}

	return callers(skip, pc)
}

type callerFrame struct {
	pointer *callerFrame
	retpc   uintptr
}

func regfp() unsafe.Pointer

//go:noinline
//go:nosplit
func callers(skip int, pc []uintptr) int {
	i := 0
	frame := (*callerFrame)(regfp())

	for i < len(pc) {
		if skip == 0 {
			pc[i] = frame.retpc
			i++
		} else {
			skip--
		}
		if frame.pointer == nil {
			break
		}
		frame = frame.pointer
	}

	return i
}

// link to https://github.com/golang/go/blob/go1.23.0/src/runtime/symtab.go#L857-L899
//
//go:linkname findfunc runtime.findfunc
func findfunc(pc uintptr) funcInfo

// Token from https://github.com/golang/go/blob/go1.23.0/src/runtime/symtab.go#L823-L826
type funcInfo struct {
	*_func
	datap *moduledata
}

func (f funcInfo) valid() bool {
	return f._func != nil
}

func (f funcInfo) entry() uintptr {
	return f.datap.textAddr(f.entryOff)
}

func (f funcInfo) _Func() *runtime.Func {
	return (*runtime.Func)(unsafe.Pointer(f._func))
}

func (f funcInfo) srcFunc() srcFunc {
	if !f.valid() {
		return srcFunc{}
	}
	return srcFunc{f.datap, f.nameOff, f.startLine, f.funcID}
}

// Token from https://github.com/golang/go/blob/go1.23.0/src/runtime/symtab.go#L901-L909
type srcFunc struct {
	datap     *moduledata
	nameOff   int32
	startLine int32
	funcID    uint8
}

// name should be an internal detail.
func (s srcFunc) name() string {
	if s.datap == nil {
		return ""
	}
	return s.datap.funcName(s.nameOff)
}

type functab struct {
	entryoff uint32 // relative to runtime.text
	funcoff  uint32
}

type textsect struct {
	vaddr    uintptr // prelinked section vaddr
	end      uintptr // vaddr + section length
	baseaddr uintptr // relocated section address
}

// Token from https://github.com/golang/go/blob/go1.23.0/src/runtime/symtab.go#L383-L436
type moduledata struct {
	pcHeader     unsafe.Pointer
	funcnametab  []byte
	cutab        []uint32
	filetab      []byte
	pctab        []byte
	pclntable    []byte
	ftab         []functab
	findfunctab  uintptr
	minpc, maxpc uintptr

	text, etext           uintptr
	noptrdata, enoptrdata uintptr
	data, edata           uintptr
	bss, ebss             uintptr
	noptrbss, enoptrbss   uintptr
	covctrs, ecovctrs     uintptr
	end, gcdata, gcbss    uintptr
	types, etypes         uintptr
	rodata                uintptr
	gofunc                uintptr // go.func.*

	textsectmap []textsect

	// omitted
}

// The compiler knows that a print of a value of this type
// should use printhex instead of printuint (decimal).
type hex uint64

//go:linkname throw runtime.throw
//go:nosplit
func throw(s string)

func (md *moduledata) textAddr(off32 uint32) uintptr {
	off := uintptr(off32)
	res := md.text + off
	if len(md.textsectmap) > 1 {
		for i, sect := range md.textsectmap {
			// For the last section, include the end address (etext), as it is included in the functab.
			if off >= sect.vaddr && off < sect.end || (i == len(md.textsectmap)-1 && off == sect.end) {
				res = sect.baseaddr + off - sect.vaddr
				break
			}
		}
		if res > md.etext && runtime.GOARCH != "wasm" { // on wasm, functions do not live in the same address space as the linear memory
			println("runtime: textAddr", hex(res), "out of range", hex(md.text), "-", hex(md.etext))
			throw("runtime: text offset out of range")
		}
	}
	return res
}

//go:nosplit
func gostringnocopy(str *byte) string {
	ss := stringStruct{str: unsafe.Pointer(str), len: findnull(str)}
	s := *(*string)(unsafe.Pointer(&ss))
	return s
}

type stringStruct struct {
	str unsafe.Pointer
	len int
}

// Token from: https://github.com/golang/go/blob/go1.24.5/src/runtime/string.go#L543-L585
//
//go:nosplit
func findnull(s *byte) int {
	if s == nil {
		return 0
	}

	// pageSize is the unit we scan at a time looking for NULL.
	// It must be the minimum page size for any architecture Go
	// runs on. It's okay (just a minor performance loss) if the
	// actual system page size is larger than this value.
	const pageSize = 4096

	offset := 0
	ptr := unsafe.Pointer(s)
	// IndexByteString uses wide reads, so we need to be careful
	// with page boundaries. Call IndexByteString on
	// [ptr, endOfPage) interval.
	safeLen := int(pageSize - uintptr(ptr)%pageSize)

	for {
		t := *(*string)(unsafe.Pointer(&stringStruct{ptr, safeLen}))
		// Check one page at a time.
		if i := strings.IndexByte(t, 0); i != -1 {
			return offset + i
		}
		// Move to next page
		ptr = unsafe.Pointer(uintptr(ptr) + uintptr(safeLen))
		offset += safeLen
		safeLen = pageSize
	}
}

// funcName returns the string at nameOff in the function name table.
func (md *moduledata) funcName(nameOff int32) string {
	if nameOff == 0 {
		return ""
	}
	return gostringnocopy(&md.funcnametab[nameOff])
}

// Token from https://github.com/golang/go/blob/go1.23.0/src/runtime/runtime2.go#L908-L953
type _func struct {
	entryOff uint32 // start pc, as offset from moduledata.text/pcHeader.textStart
	nameOff  int32  // function name, as index into moduledata.funcnametab.

	args        int32  // in/out args size
	deferreturn uint32 // offset of start of a deferreturn call instruction from entry, if any.

	pcsp      uint32
	pcfile    uint32
	pcln      uint32
	npcdata   uint32
	cuOffset  uint32 // runtime.cutab offset of this function's CU
	startLine int32  // line number of start of function (func keyword/TEXT directive)
	funcID    uint8  // set for certain special runtime functions
	flag      uint8
	_         [1]byte // pad
	nfuncdata uint8   // must be last, must end on a uint32-aligned boundary
}

// Pseudo-Func that is returned for PCs that occur in inlined code.
// A *Func can be either a *_func or a *funcinl, and they are distinguished
// by the first uintptr.
//
// TODO(austin): Can we merge this with inlinedCall?
type funcinl struct {
	ones      uint32  // set to ^0 to distinguish from _func
	entry     uintptr // entry of the real (the "outermost") frame
	name      string
	file      string
	line      int32
	startLine int32
}

// inlinedCall is the encoding of entries in the FUNCDATA_InlTree table.
type inlinedCall struct {
	funcID    uint8 // type of the called function
	_         [3]byte
	nameOff   int32 // offset into pclntab for name of called function
	parentPc  int32 // position of an instruction whose source position is the call site (offset from entry)
	startLine int32 // line number of start of function (func keyword/TEXT directive)
}

// An inlineUnwinder iterates over the stack of inlined calls at a PC by
// decoding the inline table. The last step of iteration is always the frame of
// the physical function, so there's always at least one frame.
//
// This is typically used as:
//
//	for u, uf := newInlineUnwinder(...); uf.valid(); uf = u.next(uf) { ... }
//
// Implementation note: This is used in contexts that disallow write barriers.
// Hence, the constructor returns this by value and pointer receiver methods
// must not mutate pointer fields. Also, we keep the mutable state in a separate
// struct mostly to keep both structs SSA-able, which generates much better
// code.
type inlineUnwinder struct {
	f       funcInfo
	inlTree *[1 << 20]inlinedCall
}

// isInlined returns whether uf is an inlined frame.
func (u *inlineUnwinder) isInlined(uf inlineFrame) bool {
	return uf.index >= 0
}

//go:linkname funcline1 runtime.funcline1
func funcline1(f funcInfo, targetpc uintptr, strict bool) (file string, line int32)

// fileLine returns the file name and line number of the call within the given
// frame. As a convenience, for the innermost frame, it returns the file and
// line of the PC this unwinder was started at (often this is a call to another
// physical function).
//
// It returns "?", 0 if something goes wrong.
func (u *inlineUnwinder) fileLine(uf inlineFrame) (file string, line int) {
	file, line32 := funcline1(u.f, uf.pc, false)
	return file, int(line32)
}

// srcFunc returns the srcFunc representing the given frame.
func (u *inlineUnwinder) srcFunc(uf inlineFrame) srcFunc {
	if uf.index < 0 {
		return u.f.srcFunc()
	}
	t := &u.inlTree[uf.index]
	return srcFunc{
		u.f.datap,
		t.nameOff,
		t.startLine,
		t.funcID,
	}
}

// An inlineFrame is a position in an inlineUnwinder.
type inlineFrame struct {
	// pc is the PC giving the file/line metadata of the current frame. This is
	// always a "call PC" (not a "return PC"). This is 0 when the iterator is
	// exhausted.
	pc uintptr

	// index is the index of the current record in inlTree, or -1 if we are in
	// the outermost function.
	index int32
}

//go:linkname newInlineUnwinder runtime.newInlineUnwinder
func newInlineUnwinder(f funcInfo, pc uintptr) (inlineUnwinder, inlineFrame)

// FuncForPC is a drop-in replacement for [runtime.FuncForPC].
//
//go:nosplit
func FuncForPC(pc uintptr) *runtime.Func {
	f := findfunc(pc)
	if !f.valid() {
		return nil
	}
	// This must interpret PC non-strictly so bad PCs (those between functions) don't crash the runtime.
	// We just report the preceding function in that situation. See issue 29735.
	// TODO: Perhaps we should report no function at all in that case.
	// The runtime currently doesn't have function end info, alas.
	u, uf := newInlineUnwinder(f, pc)
	if !u.isInlined(uf) {
		return f._Func()
	}
	sf := u.srcFunc(uf)
	file, line := u.fileLine(uf)
	fi := &funcinl{
		ones:      ^uint32(0),
		entry:     f.entry(), // entry of the real (the outermost) function.
		name:      sf.name(),
		file:      file,
		line:      int32(line),
		startLine: sf.startLine,
	}
	return (*runtime.Func)(unsafe.Pointer(fi))
}
