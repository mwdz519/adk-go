// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"iter"

	"google.golang.org/genai"
)

// AgentCallback represents a callback function that can be invoked before or after an agent runs.
type AgentCallback func(cctx *CallbackContext) (*genai.Content, error)

// Agent represents an all agents in Agent Development Kit.
type Agent interface {
	// Name returns the agent's name.
	//
	// Agent name must be a Python identifier and unique within the agent tree.
	// Agent name cannot be "user", since it's reserved for end-user's input.
	Name() string

	// Description returns the description about the agent's capability.
	//
	// The model uses this to determine whether to delegate control to the agent.
	// One-line description is enough and preferred.
	Description() string

	// ParentAgent is the parent agent of this agent.
	//
	// Note that an agent can ONLY be added as sub-agent once.
	//
	// If you want to add one agent twice as sub-agent, consider to create two agent
	// instances with identical config, but with different name and add them to the
	// agent tree.
	ParentAgent() Agent

	// SubAgents returns the sub-agents of this agent.
	SubAgents() []Agent

	// BeforeAgentCallbacks returns the list of [AgentCallback] to be invoked before the agent run.
	//
	// When a list of callbacks is provided, the callbacks will be called in the
	// order they are listed until a callback does not return None.
	BeforeAgentCallbacks() []AgentCallback

	// AfterAgentCallbacks returns the list of [AgentCallback] to be invoked after the agent run.
	//
	// When a list of callbacks is provided, the callbacks will be called in the
	// order they are listed until a callback does not return None.
	AfterAgentCallbacks() []AgentCallback

	// Execute is the core logic to run this agent via text-based conversation.
	//
	// Execute is equivalent to [agents._run_async_impl] in adk-python.
	Execute(ctx context.Context, ictx *InvocationContext) iter.Seq2[*Event, error]

	// ExecuteLive is the core logic to run this agent via video/audio-based conversation.
	//
	// ExecuteLive is equivalent to [agents._run_live_impl] in adk-python.
	ExecuteLive(ctx context.Context, ictx *InvocationContext) iter.Seq2[*Event, error]

	// Run is the entry method to run an agent via text-based conversation.
	//
	// Run is equivalent to [agents.run_async] in adk-python.
	Run(ctx context.Context, parentContext *InvocationContext) iter.Seq2[*Event, error]

	// RunLive is the entry method to run an agent via video/audio-based conversation.
	//
	// RunLive is equivalent to [agents.run_live] in adk-python.
	RunLive(ctx context.Context, parentContext *InvocationContext) iter.Seq2[*Event, error]

	// Gets the root agent of this agent.
	RootAgent() Agent

	// FindAgent finds the agent with the given name in this agent and its descendants.
	FindAgent(name string) Agent

	// FindSubAgent finds the agent with the given name in this agent's descendants.
	FindSubAgent(name string) Agent

	// AsLLMAgent reports whether this agent is an [LLMAgent].
	AsLLMAgent() (LLMAgent, bool)
}

// InstructionProvider is a function that provides instructions based on context.
type InstructionProvider func(rctx *ReadOnlyContext) string

// BeforeModelCallback is called before sending a request to the model.
type BeforeModelCallback func(cctx *CallbackContext, request *LLMRequest) (*LLMResponse, error)

// AfterModelCallback is called after receiving a response from the model.
type AfterModelCallback func(cctx *CallbackContext, response *LLMResponse) (*LLMResponse, error)

// BeforeToolCallback is called before executing a tool.
type BeforeToolCallback func(tool Tool, args map[string]any, toolCtx *ToolContext) (map[string]any, error)

// AfterToolCallback is called after executing a tool.
type AfterToolCallback func(tool Tool, args map[string]any, toolCtx *ToolContext, toolResponse map[string]any) (map[string]any, error)

// IncludeContents whether to include contents in the model request.
type IncludeContents string

const (
	IncludeContentsDefault IncludeContents = "default"
	IncludeContentsNone    IncludeContents = "none"
)

// LLMAgent is an interface for agents that are specifically designed to work with LLMs (Large Language Models).
type LLMAgent interface {
	Agent

	// CanonicalModel returns the resolved model field as [model.Model].
	//
	// This method is only for use by Agent Development Kit.
	CanonicalModel(ctx context.Context) (Model, error)

	// CanonicalInstructions returns the resolved self.instruction field to construct instruction for this agent.
	//
	// This method is only for use by Agent Development Kit.
	CanonicalInstructions(rctx *ReadOnlyContext) string

	// CanonicalGlobalInstruction returns the resolved self.instruction field to construct global instruction.
	//
	// This method is only for use by Agent Development Kit.
	CanonicalGlobalInstruction(rctx *ReadOnlyContext) (string, bool)

	// CanonicalTool returns the resolved tools field as a list of [Tool] based on the context.
	//
	// This method is only for use by Agent Development Kit.
	CanonicalTool(rctx *ReadOnlyContext) []Tool

	// GenerateContentConfig returns the [*genai.GenerateContentConfig] for [LLMAgent] agent.
	GenerateContentConfig() *genai.GenerateContentConfig

	// DisallowTransferToParent reports whether teh disallows LLM-controlled transferring to the parent agent.
	DisallowTransferToParent() bool

	// DisallowTransferToPeers reports whether teh disallows LLM-controlled transferring to the peer agents.
	DisallowTransferToPeers() bool

	// IncludeContents returns the mode of include contents in the model request.
	IncludeContents() IncludeContents

	// InputSchema returns the structured input.
	InputSchema() *genai.Schema

	// OutputSchema returns the structured output.
	OutputSchema() *genai.Schema

	// OutputKey returns the key in session state to store the output of the agent.
	OutputKey() string

	// Planner returns the instructs the agent to make a plan and execute it step by step.
	Planner() Planner

	// CodeExecutor returns the code executor for the agent.
	CodeExecutor() CodeExecutor

	// BeforeModelCallbacks returns the resolved self.before_model_callback field as a list of _SingleBeforeModelCallback.
	//
	// This method is only for use by Agent Development Kit.
	BeforeModelCallbacks() []BeforeModelCallback

	// AfterModelCallbacks returns the resolved self.before_tool_callback field as a list of BeforeToolCallback.
	//
	// This method is only for use by Agent Development Kit.
	AfterModelCallbacks() []AfterModelCallback

	// BeforeToolCallbacks returns the resolved self.before_tool_callback field as a list of BeforeToolCallback.
	//
	// This method is only for use by Agent Development Kit.
	BeforeToolCallback() []BeforeToolCallback

	// AfterToolCallbacks returns the resolved self.after_tool_callback field as a list of AfterToolCallback.
	//
	// This method is only for use by Agent Development Kit.
	AfterToolCallbacks() []AfterToolCallback
}
