// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package llmflow

import (
	"github.com/go-a2a/adk-go/types"
)

func SingleRequestProcessor() []types.LLMRequestProcessor {
	return []types.LLMRequestProcessor{
		&BasicLlmRequestProcessor{},
		&AuthLLMRequestProcessor{},
		&InstructionsLlmRequestProcessor{},
		&IdentityLlmRequestProcessor{},
		&ContentLLMRequestProcessor{},
		// Some implementations of NL Planning mark planning contents as thoughts
		// in the post processor. Since these need to be unmarked, NL Planning
		// should be after contents.
		&NLPlanningRequestProcessor{},
		// Code execution should be after the contents as it mutates the contents
		// to optimize data files.
		&CodeExecutionRequestProcessor{},
	}
}

func SingleResponseProcessor() []types.LLMResponseProcessor {
	return []types.LLMResponseProcessor{
		&NLPlanningResponseProcessor{},
		&CodeExecutionResponseProcessor{},
	}
}
