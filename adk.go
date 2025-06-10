// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package adk an open-source, code-first Go toolkit for building, evaluating, and deploying sophisticated AI agents with flexibility and control.
package adk

import (
	// for raw string prompt constants
	_ "github.com/MakeNowJust/heredoc/v2"
	// for prompt templating
	_ "github.com/google/dotprompt/go/dotprompt"
)

// Version is the version of the Agent Development Kit.
var Version = "v0.0.0"
