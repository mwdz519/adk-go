// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package tool provides the base infrastructure for creating and managing tools that extend agent capabilities.
//
// The tool package implements the foundation for the sophisticated tool framework that enables
// agents to interact with external systems, execute code, process data, and perform complex
// operations beyond text generation.
//
// # Core Components
//
// The package provides:
//
//   - Tool: Base class for all tool implementations
//   - Tool interface: Contract that all tools must implement
//   - Function declaration generation for LLM integration
//   - Context management for tool execution
//
// # Basic Tool Implementation
//
// Creating a custom tool:
//
//	type WeatherTool struct {
//		*tool.Tool
//		apiKey string
//	}
//
//	func NewWeatherTool(apiKey string) *WeatherTool {
//		return &WeatherTool{
//			Tool: tool.NewTool(
//				"get_weather",
//				"Get current weather information for a location",
//				false, // not long-running
//			),
//			apiKey: apiKey,
//		}
//	}
//
//	func (t *WeatherTool) GetDeclaration() *genai.FunctionDeclaration {
//		return &genai.FunctionDeclaration{
//			Name:        t.Name(),
//			Description: t.Description(),
//			Parameters: &genai.Schema{
//				Type: genai.TypeObject,
//				Properties: map[string]*genai.Schema{
//					"location": {
//						Type: genai.TypeString,
//						Description: "The city and country, e.g. 'Paris, France'",
//					},
//				},
//				Required: []string{"location"},
//			},
//		}
//	}
//
//	func (t *WeatherTool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		location := args["location"].(string)
//		// Implement weather API call
//		return weatherData, nil
//	}
//
// # Tool Integration with Agents
//
// Tools integrate seamlessly with the agent system:
//
//	weatherTool := NewWeatherTool(apiKey)
//
//	agent := agent.NewLLMAgent(ctx, "weather_assistant",
//		agent.WithTools(weatherTool),
//		agent.WithInstruction("You can check weather information using the available tools"),
//	)
//
//	// Agent automatically calls tools when needed
//	for event, err := range agent.Run(ctx, ictx) {
//		// Handle events including tool calls and results
//	}
//
// # Tool Context
//
// Tools receive rich context during execution:
//
//	func (t *MyTool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		// Access session information
//		sessionID := toolCtx.InvocationContext().SessionID()
//		userID := toolCtx.InvocationContext().UserID()
//
//		// Access state
//		state := toolCtx.GetState()
//		userPrefs := state["user:preferences"]
//
//		// Access artifact service
//		artifactService := toolCtx.GetArtifactService()
//
//		// Request credentials if needed
//		toolCtx.RequestCredential("api_key", &types.AuthConfig{...})
//
//		// Implementation
//		return result, nil
//	}
//
// # Long-Running Operations
//
// Tools can be marked as long-running for operations that take extended time:
//
//	type DataProcessingTool struct {
//		*tool.Tool
//	}
//
//	func NewDataProcessingTool() *DataProcessingTool {
//		return &DataProcessingTool{
//			Tool: tool.NewTool(
//				"process_large_dataset",
//				"Process large datasets with ML algorithms",
//				true, // long-running operation
//			),
//		}
//	}
//
//	func (t *DataProcessingTool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		// Start processing job
//		jobID := startProcessingJob(args)
//
//		// Return job ID immediately for long-running operations
//		return map[string]any{
//			"job_id": jobID,
//			"status": "started",
//			"message": "Processing started, check status with job ID",
//		}, nil
//	}
//
// # Tool Declaration Generation
//
// Tools automatically generate OpenAPI specifications for LLM consumption:
//
//	declaration := tool.GetDeclaration()
//	// Results in:
//	// {
//	//   "name": "get_weather",
//	//   "description": "Get current weather information for a location",
//	//   "parameters": {
//	//     "type": "object",
//	//     "properties": {
//	//       "location": {
//	//         "type": "string",
//	//         "description": "The city and country"
//	//       }
//	//     },
//	//     "required": ["location"]
//	//   }
//	// }
//
// # Error Handling
//
// Tools should provide meaningful error messages:
//
//	func (t *MyTool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		location, ok := args["location"].(string)
//		if !ok {
//			return nil, fmt.Errorf("location parameter must be a string, got %T", args["location"])
//		}
//
//		if location == "" {
//			return nil, fmt.Errorf("location parameter cannot be empty")
//		}
//
//		// Call external API
//		data, err := t.callWeatherAPI(location)
//		if err != nil {
//			return nil, fmt.Errorf("failed to fetch weather data: %w", err)
//		}
//
//		return data, nil
//	}
//
// # Request Processing
//
// Tools can modify outgoing LLM requests:
//
//	func (t *MyTool) ProcessLLMRequest(ctx context.Context, toolCtx *types.ToolContext, request *types.LLMRequest) error {
//		// Add tool-specific context or modify request
//		if t.needsSpecialHandling() {
//			// Modify request parameters
//			request.GenerationConfig.Temperature = 0.1
//		}
//		return nil
//	}
//
// # Tool Composition
//
// Multiple tools can be combined for complex workflows:
//
//	tools := []types.Tool{
//		NewWebSearchTool(),
//		NewDataAnalysisTool(),
//		NewVisualizationTool(),
//		NewEmailTool(),
//	}
//
//	agent := agent.NewLLMAgent(ctx, "research_assistant",
//		agent.WithTools(tools...),
//		agent.WithInstruction("Use the available tools to research topics and create reports"),
//	)
//
// # Best Practices
//
//  1. Provide clear, descriptive tool names and descriptions
//  2. Define comprehensive parameter schemas with validation
//  3. Return structured data that models can easily interpret
//  4. Handle errors gracefully with informative messages
//  5. Use long-running flag for operations taking >30 seconds
//  6. Validate input parameters before processing
//  7. Consider security implications of tool actions
//  8. Document expected input/output formats
//
// # Security Considerations
//
// Tools should implement appropriate security measures:
//   - Input validation and sanitization
//   - Authentication and authorization checks
//   - Rate limiting for external API calls
//   - Secure credential management
//   - Audit logging for sensitive operations
//
// # Performance Optimization
//
// Optimize tool performance through:
//   - Connection pooling for HTTP clients
//   - Caching frequently accessed data
//   - Async processing for long operations
//   - Proper timeout handling
//   - Resource cleanup in defer statements
//
// The tool package provides the foundation for extending agent capabilities
// with external integrations and custom functionality.
package tool
