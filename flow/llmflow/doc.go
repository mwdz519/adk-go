// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package llmflow provides pipeline architecture for processing LLM interactions with configurable request and response processors.
//
// The llmflow package implements the core processing pipeline that handles LLM requests and responses
// through a series of configurable processors. It provides the foundation for sophisticated agent
// workflows by enabling modular processing of authentication, content transformation, function calling,
// agent transfers, code execution, and more.
//
// # Architecture Overview
//
// The package follows a pipeline architecture where LLM requests and responses flow through
// a series of processors:
//
//	┌─────────────┐    ┌──────────────────┐    ┌─────────────┐    ┌───────────────────┐
//	│   Request   │───▶│ Request          │───▶│     LLM     │───▶│ Response          │
//	│             │    │ Processors       │    │    Call     │    │ Processors        │
//	└─────────────┘    └──────────────────┘    └─────────────┘    └───────────────────┘
//
// # Core Components
//
// The package provides several key components:
//
//   - LLMFlow: Base flow that orchestrates the entire pipeline
//   - Request Processors: Process and modify requests before sending to LLM
//   - Response Processors: Process and transform responses after receiving from LLM
//   - Predefined Pipelines: Common processor configurations for different use cases
//
// # LLMFlow Base Class
//
// The LLMFlow struct provides the foundation for all LLM processing:
//
//	flow := &llmflow.LLMFlow{
//		RequestProcessors: []types.LLMRequestProcessor{
//			&BasicLlmRequestProcessor{},
//			&AuthLLMRequestProcessor{},
//			&InstructionsLlmRequestProcessor{},
//		},
//		ResponseProcessors: []types.LLMResponseProcessor{
//			&CodeExecutionResponseProcessor{},
//		},
//	}
//
//	// Configure flow
//	flow.WithLogger(logger).
//		WithRequestProcessors(customProcessor).
//		WithResponseProcessors(customResponseProcessor)
//
// # Predefined Pipelines
//
// The package provides two main predefined processor pipelines:
//
// ## Single Flow Pipeline
//
// For simple LLM interactions without agent transfers:
//
//	requestProcessors := llmflow.SingleRequestProcessor()
//	responseProcessors := llmflow.SingleResponseProcessor()
//
//	// Includes:
//	// - BasicLlmRequestProcessor: Core LLM interaction
//	// - AuthLLMRequestProcessor: Authentication handling
//	// - InstructionsLlmRequestProcessor: System instructions
//	// - IdentityLlmRequestProcessor: Identity management
//	// - ContentLLMRequestProcessor: Content processing
//	// - NLPlanningRequestProcessor: Natural language planning
//	// - CodeExecutionRequestProcessor: Code execution support
//
// ## Auto Flow Pipeline
//
// For complex workflows with agent transfer capabilities:
//
//	requestProcessors := llmflow.AutoRequestProcessor()
//	responseProcessors := llmflow.AutoResponseProcessor()
//
//	// Includes all Single Flow processors plus:
//	// - AgentTransferLlmRequestProcessor: Agent transfer support
//
// # Request Processors
//
// Request processors modify and enhance LLM requests before sending:
//
// ## BasicLlmRequestProcessor
//
// Handles core LLM interaction and model management:
//
//	processor := &BasicLlmRequestProcessor{}
//	// Automatically manages model creation, request formatting, and basic error handling
//
// ## AuthLLMRequestProcessor
//
// Processes authentication requirements for tools:
//
//	processor := &AuthLLMRequestProcessor{}
//	// Handles credential requests, OAuth flows, API key management
//
// ## InstructionsLlmRequestProcessor
//
// Manages system instructions and context:
//
//	processor := &InstructionsLlmRequestProcessor{}
//	// Applies agent instructions, context-specific prompts, and system messages
//
// ## ContentLLMRequestProcessor
//
// Processes and transforms content before sending to LLM:
//
//	processor := &ContentLLMRequestProcessor{}
//	// Handles content optimization, artifact management, and context preparation
//
// ## CodeExecutionRequestProcessor
//
// Prepares code execution context and optimizes data files:
//
//	processor := &CodeExecutionRequestProcessor{}
//	// Handles code block detection, execution environment setup, and data file optimization
//
// ## AgentTransferLlmRequestProcessor
//
// Enables agent transfer capabilities:
//
//	processor := &AgentTransferLlmRequestProcessor{}
//	// Manages parent/peer agent transfers, hierarchy navigation, and delegation
//
// # Response Processors
//
// Response processors handle LLM outputs and execute actions:
//
// ## CodeExecutionResponseProcessor
//
// Executes code blocks and handles results:
//
//	processor := &CodeExecutionResponseProcessor{}
//	// Detects code blocks, executes them securely, and integrates results
//
// ## NLPlanningResponseProcessor
//
// Processes natural language planning responses:
//
//	processor := &NLPlanningResponseProcessor{}
//	// Handles planning markup, thought processing, and structured reasoning
//
// # Function Calling Integration
//
// The pipeline includes sophisticated function calling support:
//
//	// Function calls are automatically handled through the processor pipeline
//	// with parallel execution, proper error handling, and result integration
//
//	for event, err := range flow.Run(ctx, ictx) {
//		if err != nil {
//			log.Printf("Flow error: %v", err)
//			continue
//		}
//
//		if event.Actions != nil && len(event.Actions.FunctionCalls) > 0 {
//			// Function calls are automatically executed in parallel
//			// Results are integrated back into the conversation
//		}
//	}
//
// # Authentication Flow
//
// Authentication is seamlessly integrated through the auth processor:
//
//	// Tools can request credentials during execution
//	tool.RequestCredential("github_token", &types.AuthConfig{
//		Type: types.AuthTypeOAuth2,
//		ClientID: "your-client-id",
//		Scopes: []string{"repo", "user"},
//	})
//
//	// Auth processor handles the flow automatically:
//	// 1. Detects credential requests
//	// 2. Initiates appropriate auth flow (OAuth2, API Key, etc.)
//	// 3. Stores and manages credentials securely
//	// 4. Provides credentials to tools when needed
//
// # Agent Transfer Support
//
// The auto flow supports sophisticated agent transfer:
//
//	// Agents can transfer to parent or peer agents
//	event := &types.Event{
//		Actions: &types.EventActions{
//			AgentTransfer: &types.AgentTransfer{
//				Target: "parent",  // or specific agent name
//				Reason: "Escalation needed for complex analysis",
//			},
//		},
//	}
//
//	// Transfer processor handles:
//	// - Target agent resolution
//	// - Context preservation
//	// - Conversation continuity
//	// - Proper delegation workflow
//
// # Code Execution Pipeline
//
// Code execution is integrated throughout the pipeline:
//
//	// Request processor optimizes data files and prepares execution context
//	// Response processor detects and executes code blocks
//
//	// Example: Python code execution
//	response := `Here's the analysis:
//	'''python
//	import pandas as pd
//	data = pd.read_csv('data.csv')
//	print(data.describe())
//	'''`
//
//	// Processor automatically:
//	// 1. Detects code block
//	// 2. Sets up secure execution environment
//	// 3. Executes code with proper isolation
//	// 4. Captures output and integrates into conversation
//
// # Error Handling and Retry Logic
//
// The pipeline includes comprehensive error handling:
//
//	for event, err := range flow.Run(ctx, ictx) {
//		if err != nil {
//			// Handle different error types
//			if rateLimitErr, ok := err.(*types.RateLimitError); ok {
//				// Wait and retry
//				time.Sleep(rateLimitErr.RetryAfter)
//				continue
//			}
//
//			if execErr, ok := err.(*types.ExecutionError); ok {
//				// Code execution failed
//				log.Printf("Execution failed after %d attempts: %v", execErr.Attempts, execErr.LastError)
//			}
//
//			// Other error handling
//		}
//
//		// Process successful events
//	}
//
// # Streaming and Real-time Processing
//
// All processors support streaming for real-time interactions:
//
//	for event, err := range flow.Run(ctx, ictx) {
//		if err != nil {
//			continue
//		}
//
//		// Stream text deltas in real-time
//		if event.TextDelta != "" {
//			fmt.Print(event.TextDelta)
//		}
//
//		// Handle function calls as they occur
//		if event.Actions != nil && len(event.Actions.FunctionCalls) > 0 {
//			// Process function calls immediately
//		}
//	}
//
// # Custom Processor Development
//
// Create custom processors for specialized workflows:
//
//	type CustomRequestProcessor struct{}
//
//	func (p *CustomRequestProcessor) Run(ctx context.Context, ictx *types.InvocationContext, request *types.LLMRequest) iter.Seq2[*types.Event, error] {
//		return func(yield func(*types.Event, error) bool) {
//			// Custom request processing logic
//
//			// Modify request
//			request.GenerationConfig.Temperature = 0.1
//
//			// Add custom system instructions
//			if request.SystemInstruction == nil {
//				request.SystemInstruction = &genai.Content{}
//			}
//			// ... custom logic
//		}
//	}
//
//	// Integrate into pipeline
//	flow.WithRequestProcessors(&CustomRequestProcessor{})
//
// # Integration with Agent System
//
// The flow seamlessly integrates with the agent framework:
//
//	agent := agent.NewLLMAgent(ctx, "assistant",
//		agent.WithModel("gemini-1.5-pro"),
//		agent.WithInstruction("You are a helpful assistant"),
//		agent.WithTools(tool1, tool2),
//	)
//
//	// Agent automatically uses appropriate flow based on configuration
//	// SingleFlow for simple agents, AutoFlow for complex hierarchical agents
//
// # Performance Optimization
//
// The pipeline includes several performance optimizations:
//
//   - Parallel function execution with proper synchronization
//   - Connection pooling for model requests
//   - Content caching and optimization
//   - Streaming response processing
//   - Efficient memory management with object pooling
//   - Request batching where supported
//
// # Security Considerations
//
// The pipeline implements security best practices:
//
//   - Secure credential storage and management
//   - Code execution sandboxing and isolation
//   - Input validation and sanitization
//   - Rate limiting and quota management
//   - Audit logging for sensitive operations
//   - Proper authentication and authorization
//
// # Thread Safety
//
// All processors are designed to be safe for concurrent use across multiple goroutines.
// The pipeline can handle multiple concurrent requests with proper isolation.
//
// # Best Practices
//
// When working with the flow pipeline:
//
//  1. Use predefined pipelines (Single/Auto) for standard use cases
//  2. Add custom processors only when specific functionality is needed
//  3. Order processors carefully - some depend on others (e.g., content before planning)
//  4. Handle streaming events promptly to avoid blocking
//  5. Implement proper error handling for different error types
//  6. Use appropriate flow type based on agent hierarchy needs
//  7. Consider performance implications of processor ordering
//  8. Test custom processors thoroughly with various input scenarios
//
// # Configuration Examples
//
// ## Basic LLM Agent Flow
//
//	flow := &llmflow.LLMFlow{
//		RequestProcessors: llmflow.SingleRequestProcessor(),
//		ResponseProcessors: llmflow.SingleResponseProcessor(),
//	}
//
// ## Advanced Agent with Custom Processing
//
//	customProcessor := &MyCustomProcessor{}
//
//	flow := &llmflow.LLMFlow{
//		RequestProcessors: append(llmflow.AutoRequestProcessor(), customProcessor),
//		ResponseProcessors: llmflow.AutoResponseProcessor(),
//	}
//
// ## Code-Heavy Workflow
//
//	flow := &llmflow.LLMFlow{
//		RequestProcessors: []types.LLMRequestProcessor{
//			&BasicLlmRequestProcessor{},
//			&ContentLLMRequestProcessor{},
//			&CodeExecutionRequestProcessor{},
//		},
//		ResponseProcessors: []types.LLMResponseProcessor{
//			&CodeExecutionResponseProcessor{},
//		},
//	}
//
// The llmflow package provides the essential pipeline infrastructure for building
// sophisticated AI agent workflows with comprehensive LLM integration capabilities.
package llmflow
