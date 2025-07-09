// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
)

// AuthenticatedTool represents a base tool class that handles authentication before the actual tool logic
// gets called. Functions can accept a special `credential` argument which is the
// credential ready for use.(Experimental)
type AuthenticatedTool interface {
	Tool

	// Execute executes the tool logic with the provided arguments, tool context, and authentication credential.
	Execute(ctx context.Context, args map[string]any, toolCtx *ToolContext, credential *AuthCredential) (any, error)
}
