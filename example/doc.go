// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package example provides few-shot example management for improving model performance through demonstration.
//
// The example package implements a system for providing few-shot examples to language models,
// which can significantly improve their performance on specific tasks by showing examples
// of desired input-output patterns.
//
// # Core Components
//
// The package provides:
//
//   - Example: Represents a single few-shot example with input and output
//   - Provider: Interface for retrieving relevant examples based on queries
//   - Formatting utilities for converting examples to model-readable format
//   - Integration with Vertex AI example stores
//
// # Basic Usage
//
// Creating examples:
//
//	example := &example.Example{
//		Input: &genai.Content{
//			Parts: []genai.Part{genai.Text("What is the capital of France?")},
//		},
//		Output: []*genai.Content{{
//			Parts: []genai.Part{genai.Text("The capital of France is Paris.")},
//		}},
//	}
//
// Using with providers:
//
//	provider := example.NewVertexAIProvider(client, "my-example-store")
//
//	examples, err := provider.GetExamples(ctx, "geography questions")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Format examples for model
//	formatted := example.FormatExamples(examples)
//
// # Example Formatting
//
// The package provides standardized formatting for examples:
//
//	<EXAMPLES>
//	Begin few-shot
//	The following are examples of user queries and model responses using the available tools.
//
//	EXAMPLE 1:
//	Begin example
//	[user]
//	What is the weather in Tokyo?
//
//	[model]
//	I'll check the weather in Tokyo for you.
//	```tool_code
//	weather_tool(city="Tokyo")
//	```
//	```tool_outputs
//	{"temperature": "22°C", "condition": "sunny"}
//	```
//	The weather in Tokyo is currently 22°C and sunny.
//	End example
//
//	End few-shot
//	<EXAMPLES>
//
// # Provider Interface
//
// Implement the Provider interface to create custom example sources:
//
//	type CustomProvider struct {
//		examples map[string][]*example.Example
//	}
//
//	func (p *CustomProvider) GetExamples(ctx context.Context, query string) ([]*example.Example, error) {
//		// Retrieve relevant examples based on query
//		return p.examples[query], nil
//	}
//
// # Vertex AI Integration
//
// The package includes integration with Vertex AI example stores:
//
//	provider := example.NewVertexAIProvider(
//		vertexClient,
//		"projects/my-project/locations/us-central1/exampleStores/my-store",
//	)
//
//	// Retrieve examples based on semantic similarity
//	examples, err := provider.GetExamples(ctx, "technical documentation questions")
//
// # Tool Integration
//
// Examples can include tool usage patterns:
//
//	toolExample := &example.Example{
//		Input: &genai.Content{
//			Parts: []genai.Part{genai.Text("Calculate 15% tip on $42.50")},
//		},
//		Output: []*genai.Content{{
//			Parts: []genai.Part{
//				genai.Text("I'll calculate the tip for you."),
//				&genai.FunctionCall{
//					Name: "calculator",
//					Args: map[string]any{
//						"expression": "42.50 * 0.15",
//					},
//				},
//			},
//		}},
//	}
//
// # Dynamic Example Selection
//
// Providers can implement intelligent example selection:
//
//	func (p *SemanticProvider) GetExamples(ctx context.Context, query string) ([]*example.Example, error) {
//		// Use embeddings to find most relevant examples
//		embedding, err := p.embedQuery(ctx, query)
//		if err != nil {
//			return nil, err
//		}
//
//		// Search example store by similarity
//		return p.searchBySimilarity(ctx, embedding, maxExamples)
//	}
//
// # Best Practices
//
//  1. Use 2-5 examples for most tasks (more examples = higher token usage)
//  2. Choose diverse examples that cover different input patterns
//  3. Ensure examples demonstrate the desired output format
//  4. Include both successful and edge case examples
//  5. Update examples based on model performance
//  6. Use semantic similarity for dynamic example selection
//
// # Performance Considerations
//
// Few-shot examples impact:
//   - Token usage: Examples are included in every request
//   - Latency: More examples increase processing time
//   - Cost: Larger prompts cost more to process
//   - Accuracy: Good examples significantly improve performance
//
// Balance the number and quality of examples based on your use case requirements.
//
// # Integration with Agents
//
// Examples integrate with the agent system through instructions:
//
//	examples, _ := provider.GetExamples(ctx, "code review tasks")
//	formatted := example.FormatExamples(examples)
//
//	agent := agent.NewLLMAgent(ctx, "reviewer",
//		agent.WithInstruction(fmt.Sprintf(`
//			You are a code reviewer. Here are examples of good reviews:
//			%s
//
//			Now review the following code...
//		`, formatted)),
//	)
//
// This enables dynamic, context-aware example injection for improved agent performance.
package example
