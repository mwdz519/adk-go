// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"iter"
)

// LLMRequestProcessor represents a base class for LLM request processor.
type LLMRequestProcessor interface {
	// Run runs the processor.
	Run(ctx context.Context, ictx *InvocationContext, request *LLMRequest) iter.Seq2[*Event, error]
}

// LLMResponseProcessor represents a base class for LLM response processor.
type LLMResponseProcessor interface {
	// Run processes the LLM response.
	Run(ctx context.Context, ictx *InvocationContext, response *LLMResponse) iter.Seq2[*Event, error]
}
