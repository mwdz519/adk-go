// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"

	"google.golang.org/genai"
)

// Planner represents an abstract base class for all planners.
//
// The planner allows the agent to generate plans for the queries to guide its
// action.
type Planner interface {
	// BuildPlanningInstruction builds the system instruction to be appended to the LLM request for planning.
	BuildPlanningInstruction(ctx context.Context, rctx *ReadOnlyContext, request *LLMRequest) string

	// ProcessPlanningResponse Processes the LLM response for planning.
	ProcessPlanningResponse(ctx context.Context, cctx *CallbackContext, responseParts []*genai.Part) []*genai.Part
}
