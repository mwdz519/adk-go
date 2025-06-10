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
}
