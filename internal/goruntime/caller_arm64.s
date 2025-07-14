// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

#include "textflag.h"

// func regfp() uintptr
TEXT ·regfp(SB), NOSPLIT, $8-0
	MOVD (R29), R0
	MOVD R0, ret+0(FP)
	RET
