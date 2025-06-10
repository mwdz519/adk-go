// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package llmprocessor

import (
	"context"
	"iter"
	"log/slog"
	_ "unsafe" // for go:linkname

	"github.com/go-a2a/adk-go/types"
)

type LLMFlow struct {
	_ []types.LLMRequestProcessor
	_ []types.LLMResponseProcessor
	_ *slog.Logger
}

var _ types.Flow = (*LLMFlow)(nil)

func (f *LLMFlow) Run(ctx context.Context, ictx *types.InvocationContext) iter.Seq2[*types.Event, error] {
	return Run(f, ctx, ictx)
}

func (f *LLMFlow) RunLive(ctx context.Context, ictx *types.InvocationContext) iter.Seq2[*types.Event, error] {
	return RunLive(f, ctx, ictx)
}

// SingleFlow is the LLM flows that handles tools calls.
//
// A single flow only consider an agent itself and tools.
// No sub-agents are allowed for single flow.
type SingleFlow struct {
	*LLMFlow
}

//go:linkname NewLLMFlow github.com/go-a2a/adk-go/flow/llmflow.NewLLMFlow
func NewLLMFlow() *LLMFlow

//go:linkname WithRequestProcessors github.com/go-a2a/adk-go/flow/llmflow.(*LLMFlow).WithRequestProcessors
func WithRequestProcessors(f *LLMFlow, processors ...types.LLMRequestProcessor) *LLMFlow

//go:linkname WithResponseProcessors github.com/go-a2a/adk-go/flow/llmflow.(*LLMFlow).WithResponseProcessors
func WithResponseProcessors(f *LLMFlow, processors ...types.LLMResponseProcessor) *LLMFlow

//go:linkname Run github.com/go-a2a/adk-go/flow/llmflow.(*LLMFlow).Run
func Run(f *LLMFlow, ctx context.Context, ictx *types.InvocationContext) iter.Seq2[*types.Event, error]

//go:linkname RunLive github.com/go-a2a/adk-go/flow/llmflow.(*LLMFlow).RunLive
func RunLive(f *LLMFlow, ctx context.Context, ictx *types.InvocationContext) iter.Seq2[*types.Event, error]

//go:linkname SingleRequestProcessor github.com/go-a2a/adk-go/flow/llmflow.SingleRequestProcessor
func SingleRequestProcessor() []types.LLMRequestProcessor

//go:linkname SingleResponseProcessor github.com/go-a2a/adk-go/flow/llmflow.SingleResponseProcessor
func SingleResponseProcessor() []types.LLMResponseProcessor

//go:linkname AutoRequestProcessor github.com/go-a2a/adk-go/flow/llmflow.AutoRequestProcessor
func AutoRequestProcessor() []types.LLMRequestProcessor

//go:linkname AutoResponseProcessor github.com/go-a2a/adk-go/flow/llmflow.AutoResponseProcessor
func AutoResponseProcessor() []types.LLMResponseProcessor

// NewSingleFlow creates a new [SingleFlow] with the default [types.LLMRequestProcessor] and [types.LLMResponseProcessor].
func NewSingleFlow() *SingleFlow {
	flow := &SingleFlow{
		LLMFlow: NewLLMFlow(),
	}
	flow.LLMFlow = WithRequestProcessors(flow.LLMFlow, SingleRequestProcessor()...)
	flow.LLMFlow = WithResponseProcessors(flow.LLMFlow, SingleResponseProcessor()...)

	return flow
}

// AutoFlow is [SingleFlow] with agent transfer capability.
//
// Agent transfer is allowed in the following direction:
//
//  1. from parent to sub-agent;
//  2. from sub-agent to parent;
//  3. from sub-agent to its peer agents;
//
// For peer-agent transfers, it's only enabled when all below conditions are met:
//
//   - The parent agent is also of AutoFlow;
//   - `disallow_transfer_to_peer` option of this agent is False (default).
//
// Depending on the target agent flow type, the transfer may be automatically
// reversed. The condition is as below:
//
//   - If the flow type of the tranferee agent is also auto, transfee agent will
//     remain as the active agent. The transfee agent will respond to the user's
//     next message directly.
//   - If the flow type of the transfere agent is not auto, the active agent will
//     be reversed back to previous agent.
//
// TODO(adk-python): allow user to config auto-reverse function.
type AutoFlow struct {
	*LLMFlow
}

// NewAutoFlow creates a new [AutoFlow] with the default [types.LLMRequestProcessor] and [types.LLMResponseProcessor].
func NewAutoFlow() *AutoFlow {
	flow := &AutoFlow{
		LLMFlow: NewLLMFlow(),
	}
	flow.LLMFlow = WithRequestProcessors(flow.LLMFlow, AutoRequestProcessor()...)
	flow.LLMFlow = WithResponseProcessors(flow.LLMFlow, AutoResponseProcessor()...)

	return flow
}
