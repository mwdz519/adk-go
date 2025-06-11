// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"

	"google.golang.org/genai"
)

// Tool defines the interface that all tools must implement.
type Tool interface {
	// Name returns the name of the tool.
	Name() string

	// Description returns the description of the tool.
	Description() string

	// IsLongRunning whether the tool is a long running operation, which typically returns a
	// resource id first and finishes the operation later.
	IsLongRunning() bool

	// GetDeclaration gets the OpenAPI specification of this tool in the form of a [*genai.FunctionDeclaration].
	GetDeclaration() *genai.FunctionDeclaration

	// Run runs the tool with the given arguments and context.
	Run(ctx context.Context, args map[string]any, toolCtx *ToolContext) (any, error)

	// ProcessLLMRequest processes the outgoing LLM request for this tool.
	ProcessLLMRequest(ctx context.Context, toolCtx *ToolContext, llmRequest *LLMRequest) error
}
