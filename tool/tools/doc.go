// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package tools provides a comprehensive collection of pre-built tools for extending agent capabilities.
//
// The tools package implements ready-to-use tool implementations that cover common use cases
// for agent interactions, from code execution and web search to memory management and
// artifact handling. These tools can be used directly or serve as examples for custom implementations.
//
// # Available Tools
//
// The package provides several categories of tools:
//
// ## Function Tools
//   - FunctionTool: Wraps arbitrary Go functions as tools with automatic schema generation
//   - Automatic function calling utilities for reflection-based tool creation
//
// ## Agent Tools
//   - Agent: Wraps other agents as tools for hierarchical agent composition
//   - ExitLoopTool: Provides loop termination control for LoopAgent
//   - GetUserChoiceTool: Interactive user input for decision-making
//
// ## Code Execution Tools
//   - CodeExecutionTool: Secure code execution with multiple backend support
//   - Integration with codeexecutor package for sandboxed execution
//
// ## Search and Data Tools
//   - GoogleSearchTool: Built-in Google Search integration (Gemini 2.0+ models)
//   - LoadWebPageTool: Web content fetching and processing
//   - URLContextTool: Extract and analyze web page context
//
// ## Memory and Artifact Tools
//   - LoadMemoryTool: Retrieve relevant memories for context
//   - PreloadMemoryTool: Proactively load memories for agent sessions
//   - LoadArtifactsTool: Access stored artifacts and files
//   - ForwardingArtifactService: Artifact management delegation
//
// ## Utility Tools
//   - LongRunningTool: Base class for asynchronous operations
//   - ExampleTool: Demonstration tool for learning and testing
//
// # Basic Usage
//
// Using pre-built tools with agents:
//
//	agent := agent.NewLLMAgent(ctx, "assistant",
//		agent.WithTools(
//			tools.NewGoogleSearchTool(),
//			tools.NewCodeExecutionTool(executor),
//			tools.NewLoadMemoryTool(),
//		),
//		agent.WithInstruction("Use available tools to help users with research and analysis"),
//	)
//
// # Function Tool Creation
//
// Create tools from Go functions with automatic schema generation:
//
//	// Define a function
//	func CalculateArea(ctx context.Context, length, width float64) (float64, error) {
//		return length * width, nil
//	}
//
//	// Convert to tool automatically
//	calculatorTool := tools.NewFunctionTool(CalculateArea,
//		tools.WithDescription("Calculate the area of a rectangle"),
//		tools.WithParameterDescription("length", "Length of the rectangle in meters"),
//		tools.WithParameterDescription("width", "Width of the rectangle in meters"),
//	)
//
//	// Use with agent
//	agent := agent.NewLLMAgent(ctx, "calculator",
//		agent.WithTools(calculatorTool),
//	)
//
// # Agent as Tool Pattern
//
// Use agents as tools for hierarchical composition:
//
//	// Create specialized sub-agents
//	researchAgent := agent.NewLLMAgent(ctx, "researcher",
//		agent.WithTools(tools.NewGoogleSearchTool()),
//		agent.WithInstruction("Research topics thoroughly using web search"),
//	)
//
//	analysisAgent := agent.NewLLMAgent(ctx, "analyst",
//		agent.WithTools(tools.NewCodeExecutionTool(executor)),
//		agent.WithInstruction("Analyze data using statistical methods"),
//	)
//
//	// Create tool wrappers
//	researchTool := tools.NewAgent("research", "Research topics using web search")
//	analysisTool := tools.NewAgent("analyze", "Analyze data statistically")
//
//	// Compose into coordinator agent
//	coordinator := agent.NewLLMAgent(ctx, "coordinator",
//		agent.WithTools(researchTool, analysisTool),
//		agent.WithInstruction("Coordinate research and analysis tasks"),
//	)
//
// # Code Execution Tool
//
// Provide secure code execution capabilities:
//
//	// Create executor with desired backend
//	executor := codeexecutor.NewContainerExecutor(
//		codeexecutor.WithImage("python:3.11-slim"),
//		codeexecutor.WithTimeout(30*time.Second),
//	)
//
//	// Create code execution tool
//	codeTool := tools.NewCodeExecutionTool(executor,
//		tools.WithLanguageSupport("python", "javascript", "bash"),
//		tools.WithMaxRetries(2),
//	)
//
//	agent := agent.NewLLMAgent(ctx, "developer",
//		agent.WithTools(codeTool),
//		agent.WithInstruction("You can execute code to solve problems"),
//	)
//
// # Memory Integration
//
// Enable agents to access and store memories:
//
//	memoryService := memory.NewInMemoryService()
//
//	agent := agent.NewLLMAgent(ctx, "assistant",
//		agent.WithTools(
//			tools.NewLoadMemoryTool(),
//			tools.NewPreloadMemoryTool(),
//		),
//		agent.WithMemoryService(memoryService),
//		agent.WithInstruction("Remember important information from our conversations"),
//	)
//
// # Web Search and Content Tools
//
// Enable web research capabilities:
//
//	agent := agent.NewLLMAgent(ctx, "researcher",
//		agent.WithTools(
//			tools.NewGoogleSearchTool(),
//			tools.NewLoadWebPageTool(),
//			tools.NewURLContextTool(),
//		),
//		agent.WithInstruction("Research topics using web search and content analysis"),
//	)
//
// # Long-Running Operations
//
// Handle asynchronous operations with long-running tools:
//
//	type DataProcessingTool struct {
//		*tools.LongRunningTool
//	}
//
//	func (t *DataProcessingTool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		// Start background processing
//		jobID := t.startProcessingJob(args)
//
//		// Return immediately with job tracking info
//		return map[string]any{
//			"job_id": jobID,
//			"status": "processing",
//			"check_url": fmt.Sprintf("/jobs/%s/status", jobID),
//		}, nil
//	}
//
// # Interactive Tools
//
// Create tools that interact with users:
//
//	agent := agent.NewLLMAgent(ctx, "interactive_assistant",
//		agent.WithTools(
//			tools.NewGetUserChoiceTool(),
//		),
//		agent.WithInstruction("Ask users for clarification when needed"),
//	)
//
//	// Agent can now ask users to choose between options
//	// The GetUserChoiceTool will present choices and collect user input
//
// # Artifact Management
//
// Tools can save and load artifacts:
//
//	func (t *ReportGeneratorTool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		report := t.generateReport(args)
//
//		// Save report as artifact
//		artifactService := toolCtx.GetArtifactService()
//		_, err := artifactService.SaveArtifact(ctx,
//			toolCtx.AppName(), toolCtx.UserID(), toolCtx.SessionID(),
//			"report.pdf", &genai.Part{Bytes: reportPDF})
//
//		return map[string]any{
//			"status": "completed",
//			"artifact": "report.pdf",
//		}, nil
//	}
//
// # Error Handling Best Practices
//
// Tools should provide clear error messages:
//
//	func (t *MyTool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		// Validate required parameters
//		query, ok := args["query"].(string)
//		if !ok || query == "" {
//			return nil, fmt.Errorf("query parameter is required and must be a non-empty string")
//		}
//
//		// Handle context cancellation
//		select {
//		case <-ctx.Done():
//			return nil, ctx.Err()
//		default:
//		}
//
//		// Call external service with proper error handling
//		result, err := t.callExternalAPI(ctx, query)
//		if err != nil {
//			return nil, fmt.Errorf("external API call failed: %w", err)
//		}
//
//		return result, nil
//	}
//
// # Authentication Integration
//
// Tools can request authentication credentials:
//
//	func (t *APITool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		// Request API key authentication
//		authConfig := &types.AuthConfig{
//			Type: types.AuthTypeAPIKey,
//			CredentialKey: "external_api_key",
//		}
//
//		toolCtx.RequestCredential("api_access", authConfig)
//
//		// Tool framework handles credential flow
//		// Tool receives authenticated client or credentials
//		return t.callAuthenticatedAPI(args)
//	}
//
// # Performance Considerations
//
//  1. Cache expensive computations and API calls
//  2. Use appropriate timeouts for external services
//  3. Implement connection pooling for HTTP clients
//  4. Consider async processing for long operations
//  5. Validate inputs early to avoid wasted processing
//
// # Security Guidelines
//
//  1. Validate and sanitize all input parameters
//  2. Use secure credential management
//  3. Implement rate limiting for external APIs
//  4. Log tool usage for audit trails
//  5. Follow principle of least privilege
//
// The tools package provides a solid foundation for agent capabilities while
// serving as examples for implementing custom tools.
package tools
