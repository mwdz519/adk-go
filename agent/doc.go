// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package agent provides hierarchical agent implementations for building sophisticated AI agents.
//
// The agent package implements a hierarchical, event-driven agent architecture with four core agent types:
//
//   - LLMAgent: Full-featured agents powered by language models with tools, instructions, callbacks, planners, and code execution
//   - SequentialAgent: Executes sub-agents one after another, supports live mode with taskCompleted() flow control
//   - ParallelAgent: Runs sub-agents concurrently in isolated branches, merges event streams
//   - LoopAgent: Repeatedly executes sub-agents until escalation or max iterations
//
// All agents embed types.BaseAgent for common functionality and use event streaming via
// iter.Seq2[*Event, error] iterators for real-time processing. The rich InvocationContext
// tracks execution state, session, and hierarchy with before/after callbacks for customizing behavior.
//
// # Basic Usage
//
// Creating an LLM agent:
//
//	agent := agent.NewLLMAgent(ctx, "my_agent",
//		agent.WithModel("gemini-2.0-flash-exp"),
//		agent.WithInstruction("You are a helpful assistant"),
//		agent.WithTools(tool1, tool2),
//	)
//
// Creating a sequential agent:
//
//	sequential := agent.NewSequentialAgent("coordinator").
//		WithAgents(subAgent1, subAgent2, subAgent3)
//
// Running an agent:
//
//	for event, err := range agent.Run(ctx, invocationContext) {
//		if err != nil {
//			log.Fatal(err)
//		}
//		// Process event
//	}
//
// # Agent Types
//
// LLMAgent provides comprehensive LLM integration with features like:
//   - Model abstraction supporting multiple providers
//   - Tool execution with parallel processing
//   - Custom instructions and dynamic context
//   - Before/after callbacks for customization
//   - Planning and reasoning capabilities
//   - Code execution support
//
// SequentialAgent executes child agents in order:
//   - Useful for multi-step workflows
//   - Supports live mode for real-time interactions
//   - Maintains conversation flow between agents
//
// ParallelAgent runs multiple agents concurrently:
//   - Isolated execution branches
//   - Event stream merging
//   - Useful for multi-perspective analysis
//
// LoopAgent provides iterative execution:
//   - Configurable maximum iterations
//   - Escalation-based termination
//   - Useful for refinement workflows
//
// # Event-Driven Architecture
//
// All agents use Go 1.23+ iterators for streaming results:
//
//	for event, err := range agent.Run(ctx, ictx) {
//		if err != nil {
//			// Handle error
//			continue
//		}
//
//		switch event.Type {
//		case types.EventTypeTextDelta:
//			// Handle streaming text
//		case types.EventTypeFunctionCall:
//			// Handle tool execution
//		}
//	}
//
// # Callbacks and Customization
//
// Agents support before/after callbacks for customization:
//
//	agent.WithBeforeModelCallback(func(cctx *types.CallbackContext, req *types.LLMRequest) (*types.LLMResponse, error) {
//		// Modify request before sending to model
//		return nil, nil
//	})
//
//	agent.WithAfterModelCallback(func(cctx *types.CallbackContext, resp *types.LLMResponse) (*types.LLMResponse, error) {
//		// Process response after receiving from model
//		return resp, nil
//	})
//
// # Hierarchical Composition
//
// Agents form trees with parent/child relationships:
//
//	coordinator := agent.NewSequentialAgent("coordinator")
//	analyzer := agent.NewLLMAgent(ctx, "analyzer", ...)
//	reporter := agent.NewLLMAgent(ctx, "reporter", ...)
//
//	coordinator.WithAgents(analyzer, reporter)
//
// The agent hierarchy enables complex workflows with proper context propagation
// and state management throughout the execution tree.
package agent
