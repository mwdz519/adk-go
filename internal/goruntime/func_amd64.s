// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

//go:build go1.21 && amd64

#include "textflag.h"
#include "funcdata.h"

// func Func() string
TEXT ·Func(SB), NOSPLIT|NOFRAME, $24-16
	NO_LOCAL_POINTERS

	MOVQ addr-8(FP), AX
	MOVQ AX, 0(SP)
	CALL ·Name(SB)
	MOVQ 16(SP), AX
	MOVQ 8(SP), CX
	MOVQ CX, ret_base+0(FP)
	MOVQ AX, ret_len+8(FP)
	RET
