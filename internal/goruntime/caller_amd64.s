// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

#include "textflag.h"

// func regfp() uintptr
TEXT ·regfp(SB), NOSPLIT|NOFRAME, $0-8
	MOVQ BP, ret+0(FP)
	RET
