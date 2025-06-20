// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package model provides multi-provider LLM integration with unified interfaces and automatic model resolution.
//
// The model package implements the types.Model interface for various Large Language Model providers,
// using google.golang.org/genai as the primary abstraction layer. It provides consistent content format,
// streaming patterns, and provider-specific conversions while supporting both synchronous and streaming generation.
//
// # Supported Providers
//
// The package supports multiple LLM providers:
//
//   - Google Gemini: Direct integration with full streaming and live connection support
//   - Anthropic Claude: Support for direct API, Vertex AI, and AWS Bedrock deployments
//   - Registry-based extensibility for additional providers
//
// # Model Registry
//
// Models are automatically resolved using regex pattern matching:
//
//	// Gemini models
//	gemini-1.5-pro
//	gemini-2.0-flash-exp
//	projects/my-project/locations/us-central1/publishers/google/models/gemini-pro
//
//	// Claude models
//	claude-3-5-sonnet-20241022
//	claude-3-haiku-20240307
//
// # Basic Usage
//
// Creating models using the factory pattern:
//
//	factory := model.NewDefaultModelFactory("your-api-key")
//	model, err := factory.CreateModel(ctx, "gemini-1.5-pro")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer model.Close()
//
// Direct model creation:
//
//	// Google Gemini
//	gemini, err := model.NewGemini(ctx, apiKey, "gemini-1.5-pro")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Anthropic Claude
//	claude, err := model.NewClaude(ctx, "claude-3-5-sonnet-20241022", model.ClaudeModeAnthropic)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Content Generation
//
// Synchronous generation:
//
//	request := &types.LLMRequest{
//		Contents: []*genai.Content{{
//			Parts: []genai.Part{genai.Text("What is the capital of France?")},
//		}},
//		GenerationConfig: &genai.GenerationConfig{
//			Temperature: 0.7,
//			MaxOutputTokens: 1000,
//		},
//	}
//
//	response, err := model.GenerateContent(ctx, request)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println(response.Candidates[0].Content.Parts[0])
//
// Streaming generation:
//
//	for event, err := range model.GenerateContentStream(ctx, request) {
//		if err != nil {
//			log.Printf("Stream error: %v", err)
//			continue
//		}
//
//		if event.TextDelta != "" {
//			fmt.Print(event.TextDelta)
//		}
//	}
//
// # Live Connections
//
// Some providers support stateful live connections for real-time interactions:
//
//	if liveModel, ok := model.(types.LiveModel); ok {
//		conn, err := liveModel.StartLiveConnection(ctx, config)
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer conn.Close()
//
//		// Real-time bidirectional communication
//		for event, err := range conn.ReceiveEvents(ctx) {
//			// Handle real-time events
//		}
//	}
//
// # Model Configuration
//
// Models support extensive configuration options:
//
//	model, err := model.NewGemini(ctx, apiKey, "gemini-1.5-pro",
//		model.WithTemperature(0.7),
//		model.WithMaxOutputTokens(2048),
//		model.WithTopP(0.9),
//		model.WithTopK(40),
//		model.WithSafetySettings(safetySettings),
//		model.WithSystemInstruction("You are a helpful assistant"),
//	)
//
// # Claude Integration
//
// Claude models support multiple deployment modes:
//
//	// Direct Anthropic API
//	claude, err := model.NewClaude(ctx, "claude-3-5-sonnet-20241022", model.ClaudeModeAnthropic)
//
//	// Vertex AI deployment
//	claude, err := model.NewClaude(ctx, "claude-3-5-sonnet@20241022", model.ClaudeModeVertexAI)
//
//	// AWS Bedrock deployment
//	claude, err := model.NewClaude(ctx, "anthropic.claude-3-5-sonnet-20241022-v2:0", model.ClaudeModeBedrock)
//
// # Custom Model Registration
//
// Register custom model implementations:
//
//	model.RegisterLLMType(
//		[]string{`my-custom-model-.*`},
//		func(ctx context.Context, apiKey, modelName string) (types.Model, error) {
//			return NewCustomModel(ctx, apiKey, modelName)
//		},
//	)
//
//	// Now factory can create custom models
//	customModel, err := factory.CreateModel(ctx, "my-custom-model-v1")
//
// # Error Handling
//
// The package provides detailed error information:
//
//	response, err := model.GenerateContent(ctx, request)
//	if err != nil {
//		if rateLimitErr, ok := err.(*types.RateLimitError); ok {
//			fmt.Printf("Rate limited, retry after: %v\n", rateLimitErr.RetryAfter)
//		} else if quotaErr, ok := err.(*types.QuotaExceededError); ok {
//			fmt.Printf("Quota exceeded: %s\n", quotaErr.Message)
//		} else {
//			fmt.Printf("Generation error: %v\n", err)
//		}
//	}
//
// # Function Calling
//
// Models support function calling for tool integration:
//
//	tools := []*genai.Tool{{
//		FunctionDeclarations: []*genai.FunctionDeclaration{{
//			Name: "get_weather",
//			Description: "Get current weather for a location",
//			Parameters: weatherSchema,
//		}},
//	}}
//
//	request := &types.LLMRequest{
//		Contents: contents,
//		Tools: tools,
//		ToolConfig: &genai.ToolConfig{
//			FunctionCallingConfig: &genai.FunctionCallingConfig{
//				Mode: genai.FunctionCallingAuto,
//			},
//		},
//	}
//
// # Content Caching
//
// Support for content caching to optimize token usage:
//
//	request := &types.LLMRequest{
//		Contents: contents,
//		CachedContent: "projects/my-project/locations/us-central1/cachedContents/my-cache",
//	}
//
// # Thread Safety
//
// All model implementations are safe for concurrent use across multiple goroutines.
// Each request is handled independently with proper context propagation.
//
// # Performance Optimization
//
// The package provides several performance optimizations:
//   - Connection pooling for HTTP requests
//   - Automatic retry with exponential backoff
//   - Request batching where supported
//   - Efficient streaming with minimal buffering
//   - Content caching for large contexts
//
// # Environment Variables
//
// Models can be configured via environment variables:
//
//	GOOGLE_API_KEY        - Google AI API key
//	ANTHROPIC_API_KEY     - Anthropic API key
//	VERTEX_AI_PROJECT     - Google Cloud project for Vertex AI
//	VERTEX_AI_LOCATION    - Google Cloud location for Vertex AI
//
// # Integration with Agent System
//
// Models integrate seamlessly with the agent framework:
//
//	agent := agent.NewLLMAgent(ctx, "assistant",
//		agent.WithModel("gemini-1.5-pro"),
//		agent.WithInstruction("You are a helpful assistant"),
//	)
//
// The agent system automatically handles model creation, request formatting,
// response processing, and error handling.
package model
