// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package llmprocessor provides a public API facade for LLM flow processing with simplified interfaces for common use cases.
//
// The llmprocessor package implements a facade pattern over the internal flow/llmflow package,
// providing clean, easy-to-use interfaces for building LLM-powered agent workflows. It offers
// two main flow types optimized for different use cases and hides the complexity of the
// underlying processor pipeline architecture.
//
// # Design Philosophy
//
// This package follows the facade design pattern, using Go's go:linkname directive to provide
// a stable public API while linking to internal implementation details. This approach offers:
//
//   - Clean separation between public API and internal complexity
//   - Simplified interfaces for common use cases
//   - Stable API that can evolve independently of internal implementation
//   - Better encapsulation and reduced coupling
//
// # Core Flow Types
//
// The package provides two primary flow types optimized for different scenarios:
//
// ## SingleFlow
//
// For straightforward LLM interactions with tool calling but without agent transfers:
//
//	flow := llmprocessor.NewSingleFlow()
//
//	// Includes optimized processors for:
//	// - Basic LLM interaction
//	// - Authentication handling
//	// - System instructions
//	// - Content processing
//	// - Natural language planning
//	// - Code execution
//	// - Tool/function calling
//
// ## AutoFlow
//
// For complex agent hierarchies with full agent transfer capabilities:
//
//	flow := llmprocessor.NewAutoFlow()
//
//	// Includes all SingleFlow capabilities plus:
//	// - Parent/child agent transfers
//	// - Peer agent transfers
//	// - Sophisticated delegation workflows
//	// - Hierarchy navigation
//
// # Basic Usage
//
// ## Simple LLM Agent
//
// For basic LLM interactions with tool support:
//
//	// Create a single flow for straightforward interactions
//	flow := llmprocessor.NewSingleFlow()
//
//	// Use with an LLM agent
//	agent := agent.NewLLMAgent(ctx, "assistant",
//		agent.WithModel("gemini-1.5-pro"),
//		agent.WithInstruction("You are a helpful assistant"),
//		agent.WithTools(searchTool, calculatorTool),
//		agent.WithFlow(flow),
//	)
//
//	// Execute the agent
//	for event, err := range agent.Run(ctx, ictx) {
//		if err != nil {
//			log.Printf("Error: %v", err)
//			continue
//		}
//
//		// Handle streaming events
//		if event.TextDelta != "" {
//			fmt.Print(event.TextDelta)
//		}
//	}
//
// ## Hierarchical Agent System
//
// For complex workflows with agent transfers:
//
//	// Create auto flow for transfer capabilities
//	coordinatorFlow := llmprocessor.NewAutoFlow()
//	analystFlow := llmprocessor.NewAutoFlow()
//	reporterFlow := llmprocessor.NewSingleFlow()
//
//	// Create agent hierarchy
//	coordinator := agent.NewLLMAgent(ctx, "coordinator",
//		agent.WithFlow(coordinatorFlow),
//		agent.WithInstruction("Coordinate complex analysis tasks"),
//	)
//
//	analyst := agent.NewLLMAgent(ctx, "data_analyst",
//		agent.WithFlow(analystFlow),
//		agent.WithInstruction("Perform detailed data analysis"),
//		agent.WithTools(dataTool, visualizationTool),
//	)
//
//	reporter := agent.NewLLMAgent(ctx, "reporter",
//		agent.WithFlow(reporterFlow),
//		agent.WithInstruction("Generate comprehensive reports"),
//		agent.WithTools(documentTool),
//	)
//
//	// Set up hierarchy
//	coordinator.WithAgents(analyst, reporter)
//
//	// Execute with automatic transfers
//	for event, err := range coordinator.Run(ctx, ictx) {
//		// Agent transfers happen automatically based on context
//	}
//
// # Agent Transfer Mechanics
//
// AutoFlow provides sophisticated agent transfer capabilities:
//
// ## Transfer Directions
//
// The flow supports transfers in multiple directions:
//
//	// 1. Parent to child agent
//	coordinator → analyst
//
//	// 2. Child to parent agent
//	analyst → coordinator
//
//	// 3. Peer to peer transfers (when conditions are met)
//	analyst → reporter
//
// ## Transfer Conditions
//
// Peer agent transfers are enabled when:
//
//   - The parent agent uses AutoFlow
//   - The transferring agent has transfer permissions enabled
//   - Target peer agent is available and appropriate
//
// ## Automatic Reversal
//
// Transfer behavior depends on target agent flow type:
//
//	// AutoFlow target: Transfer persists
//	analyst (AutoFlow) → coordinator (AutoFlow)
//	// Coordinator remains active for next interaction
//
//	// SingleFlow target: Transfer reverses
//	analyst (AutoFlow) → reporter (SingleFlow)
//	// Control returns to analyst after reporter completes
//
// # Integration Patterns
//
// ## With Agent Framework
//
// The flows integrate seamlessly with the agent system:
//
//	agent := agent.NewLLMAgent(ctx, "agent_name",
//		agent.WithFlow(llmprocessor.NewAutoFlow()),
//		// ... other configurations
//	)
//
//	// Flow is automatically used for all agent interactions
//
// ## Custom Flow Configuration
//
// While the package provides optimized defaults, you can customize flows:
//
//	// Start with base flow
//	flow := llmprocessor.NewSingleFlow()
//
//	// Access underlying LLMFlow for customization
//	llmFlow := flow.LLMFlow
//	llmFlow = llmprocessor.WithRequestProcessors(llmFlow, customProcessor)
//	llmFlow = llmprocessor.WithResponseProcessors(llmFlow, customResponseProcessor)
//
// # Function Calling Support
//
// Both flow types include comprehensive function calling:
//
//	// Tools are automatically executed when called by the LLM
//	searchTool := tools.NewFunctionTool("web_search", searchFunction,
//		tools.WithDescription("Search the web for information"),
//	)
//
//	flow := llmprocessor.NewSingleFlow()
//	agent := agent.NewLLMAgent(ctx, "searcher",
//		agent.WithFlow(flow),
//		agent.WithTools(searchTool),
//	)
//
//	// Function calls are handled automatically:
//	// 1. LLM decides to call function
//	// 2. Flow executes function in parallel if multiple calls
//	// 3. Results are integrated back into conversation
//	// 4. LLM continues with function results
//
// # Code Execution Integration
//
// Both flows support secure code execution:
//
//	// Code blocks in responses are automatically detected and executed
//	response := "Let me analyze the data:\n```python\nimport pandas as pd\ndata.describe()\n```"
//
//	// Flow automatically:
//	// 1. Detects code block
//	// 2. Sets up secure execution environment
//	// 3. Executes code safely
//	// 4. Captures output
//	// 5. Integrates results into conversation
//
// # Authentication Handling
//
// Authentication is transparently managed:
//
//	// Tools can request authentication
//	func (t *GitHubTool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		// Request GitHub credentials
//		toolCtx.RequestCredential("github_token", &types.AuthConfig{
//			Type: types.AuthTypeOAuth2,
//			ClientID: "your-client-id",
//			Scopes: []string{"repo"},
//		})
//
//		// Flow handles OAuth flow automatically
//		// Credentials are securely stored and provided
//	}
//
// # Error Handling Patterns
//
// Flows provide comprehensive error handling:
//
//	for event, err := range flow.Run(ctx, ictx) {
//		if err != nil {
//			// Handle specific error types
//			switch e := err.(type) {
//			case *types.RateLimitError:
//				// Wait and retry
//				time.Sleep(e.RetryAfter)
//				continue
//
//			case *types.AuthenticationError:
//				// Handle auth failure
//				log.Printf("Authentication failed: %v", e)
//
//			case *types.ExecutionError:
//				// Handle code execution failure
//				log.Printf("Code execution failed after %d attempts: %v",
//					e.Attempts, e.LastError)
//
//			default:
//				log.Printf("Flow error: %v", err)
//			}
//			continue
//		}
//
//		// Process successful events
//		processEvent(event)
//	}
//
// # Live Connection Support
//
// Both flows support live connections for real-time interactions:
//
//	// For models that support live connections (e.g., Gemini 2.0)
//	for event, err := range flow.RunLive(ctx, ictx) {
//		if err != nil {
//			log.Printf("Live error: %v", err)
//			continue
//		}
//
//		// Handle real-time events including audio/video
//		switch event.Type {
//		case types.EventTypeAudioDelta:
//			// Process audio stream
//		case types.EventTypeTextDelta:
//			// Process text stream
//		}
//	}
//
// # Performance Characteristics
//
// The flows are optimized for performance:
//
//   - Parallel function execution when multiple tools are called
//   - Streaming response processing for real-time interactions
//   - Connection pooling for HTTP requests
//   - Efficient memory management with object pooling
//   - Minimal overhead facade pattern implementation
//   - Optimized processor pipelines for common use cases
//
// # When to Use Each Flow Type
//
// ## Use SingleFlow When:
//
//   - Building standalone LLM agents
//   - Tool calling is primary interaction pattern
//   - No need for agent transfers or hierarchies
//   - Simpler workflows with direct user interaction
//   - Performance is critical and transfers not needed
//
// ## Use AutoFlow When:
//
//   - Building hierarchical agent systems
//   - Need delegation and escalation capabilities
//   - Complex workflows spanning multiple specialized agents
//   - Peer collaboration between agents is required
//   - Dynamic task routing based on expertise
//
// # Migration and Compatibility
//
// The package is designed for easy migration and compatibility:
//
//	// Existing flow/llmflow code can be easily migrated
//	// Old approach:
//	flow := &llmflow.LLMFlow{
//		RequestProcessors: llmflow.SingleRequestProcessor(),
//		ResponseProcessors: llmflow.SingleResponseProcessor(),
//	}
//
//	// New simplified approach:
//	flow := llmprocessor.NewSingleFlow()
//
// # Advanced Customization
//
// For advanced use cases requiring custom processors:
//
//	// Create base flow
//	flow := llmprocessor.NewAutoFlow()
//
//	// Add custom processors via the underlying LLMFlow
//	customRequestProcessor := &MyCustomRequestProcessor{}
//	customResponseProcessor := &MyCustomResponseProcessor{}
//
//	flow.LLMFlow = llmprocessor.WithRequestProcessors(flow.LLMFlow, customRequestProcessor)
//	flow.LLMFlow = llmprocessor.WithResponseProcessors(flow.LLMFlow, customResponseProcessor)
//
//	// Use customized flow
//	agent := agent.NewLLMAgent(ctx, "custom_agent",
//		agent.WithFlow(flow),
//	)
//
// # Thread Safety
//
// All flow types are safe for concurrent use across multiple goroutines.
// Each execution context is properly isolated.
//
// # Best Practices
//
//  1. Use SingleFlow for simple, direct LLM interactions
//  2. Use AutoFlow for complex hierarchical workflows
//  3. Handle errors appropriately for different error types
//  4. Process streaming events promptly to avoid blocking
//  5. Consider performance implications of flow choice
//  6. Test transfer scenarios thoroughly in hierarchical setups
//  7. Use appropriate authentication configurations for tools
//  8. Monitor resource usage in high-throughput scenarios
//
// # Integration Examples
//
// ## Data Analysis Workflow
//
//	analyzerFlow := llmprocessor.NewSingleFlow()
//	analyzer := agent.NewLLMAgent(ctx, "data_analyzer",
//		agent.WithFlow(analyzerFlow),
//		agent.WithTools(dataTool, plotTool),
//	)
//
// ## Customer Service Hierarchy
//
//	// Tier 1 support (SingleFlow)
//	tier1 := agent.NewLLMAgent(ctx, "tier1_support",
//		agent.WithFlow(llmprocessor.NewSingleFlow()),
//		agent.WithTools(kbTool, ticketTool),
//	)
//
//	// Specialist support (AutoFlow for escalation)
//	specialist := agent.NewLLMAgent(ctx, "specialist",
//		agent.WithFlow(llmprocessor.NewAutoFlow()),
//		agent.WithTools(advancedTools...),
//	)
//
//	// Coordinator (AutoFlow for routing)
//	coordinator := agent.NewLLMAgent(ctx, "coordinator",
//		agent.WithFlow(llmprocessor.NewAutoFlow()),
//	)
//	coordinator.WithAgents(tier1, specialist)
//
// The llmprocessor package provides the essential building blocks for creating
// sophisticated LLM-powered agent workflows with clean, maintainable APIs.
package llmprocessor
