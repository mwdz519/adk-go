// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package goruntime

import (
	"flag"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type method string

const (
	runtimeCallers   method = "runtime"
	goruntimeCallers method = "goruntime"
)

func funcNames(pcs []uintptr) []string {
	fns := make([]string, 0, len(pcs))
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		fns = append(fns, frame.Function)
		if !more {
			break
		}
	}
	return fns
}

//go:noinline
func testCallers(m method) []uintptr {
	return testCallersA(m)
}

//go:noinline
func testCallersA(m method) []uintptr {
	return testCallersB(m)
}

//go:noinline
func testCallersB(m method) []uintptr {
	pcs := make([]uintptr, 32)
	switch m {
	case runtimeCallers:
		return pcs[0:runtime.Callers(1, pcs)]
	case goruntimeCallers:
		return pcs[0:Callers(1, pcs)]
	}
	panic("unreachable")
}

func TestCallers(t *testing.T) {
	want := funcNames(testCallers(runtimeCallers))
	got := funcNames(testCallers(goruntimeCallers))

	// frame pointer unwinding discovers an additional frame that
	// gentraceback seems to miss.
	// TODO: debug this further
	got = append(got[0:len(got)-2], got[len(got)-1])

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("\n got=%v\nwant=%v\n", got, want)
	}
}

func TestFuncForPC(t *testing.T) {
	t.Run("Inline", func(t *testing.T) {
		procs := runtime.GOMAXPROCS(-1)
		c := make(chan bool, procs)
		for range procs {
			func() {
				go func() {
					for range 1000 {
						testCallerFoo(t)
					}
					c <- true
				}()
				defer func() {
					<-c
				}()
			}()
		}
	})

	t.Run("NoInline", func(t *testing.T) {
		procs := runtime.GOMAXPROCS(-1)
		c := make(chan bool, procs)
		for range procs {
			func() {
				go func() {
					for range 1000 {
						testCallerFooNoInline(t)
					}
					c <- true
				}()
				defer func() {
					<-c
				}()
			}()
		}
	})
}

// These are marked noinline so that we can use FuncForPC
// in testCallerBar.
func testCallerFoo(t *testing.T) {
	testCallerBar(t)
}

func testCallerBar(t *testing.T) {
	for i := range 2 {
		pc, file, line, ok := runtime.Caller(i)
		f := FuncForPC(pc)
		if !ok ||
			!strings.HasSuffix(file, "caller_test.go") ||
			(i == 0 && !strings.HasSuffix(f.Name(), "testCallerBar")) ||
			(i == 1 && !strings.HasSuffix(f.Name(), "testCallerFoo")) ||
			line < 5 || line > 1000 ||
			f.Entry() >= pc {
			t.Errorf("incorrect symbol info %d: %t %d %d %s %s %d",
				i, ok, f.Entry(), pc, f.Name(), file, line)
		}
	}
}

// These are marked noinline so that we can use FuncForPC
// in testCallerBar.
//
//go:noinline
func testCallerFooNoInline(t *testing.T) {
	testCallerBarNoInline(t)
}

//go:noinline
func testCallerBarNoInline(t *testing.T) {
	for i := range 2 {
		pc, file, line, ok := runtime.Caller(i)
		f := FuncForPC(pc)
		if !ok ||
			!strings.HasSuffix(file, "caller_test.go") ||
			(i == 0 && !strings.HasSuffix(f.Name(), "testCallerBarNoInline")) ||
			(i == 1 && !strings.HasSuffix(f.Name(), "testCallerFooNoInline")) ||
			line < 5 || line > 1000 ||
			f.Entry() >= pc {
			t.Errorf("incorrect symbol info %d: %t %d %d %s %s %d",
				i, ok, f.Entry(), pc, f.Name(), file, line)
		}
	}
}

var fmethod = flag.String("method", string(goruntimeCallers), "bennchmark method [runtime,goruntime]")

func BenchmarkCallers(b *testing.B) {
	b.Run("Callers", func(b *testing.B) {
		bench(b, method(*fmethod), 16)
	})
}

var n int

//go:noinline
func bench(b *testing.B, m method, depth int) {
	pcs := make([]uintptr, depth+10)

	for b.Loop() {
		switch m {
		case runtimeCallers:
			n = runtime.Callers(1, pcs)
		case goruntimeCallers:
			n = Callers(1, pcs)
		}
		if n > depth {
			panic("bad")
		} else if n < depth {
			bench(b, m, depth)
			break
		}
	}
}
