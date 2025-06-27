// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package aiconv provides comprehensive bidirectional type conversion between unified genai types and provider-specific AI platform types.
//
// The aiconv package serves as the critical bridge between [google.golang.org/genai]'s unified interface
// and [cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb]'s provider-specific protobuf types. It enables seamless
// interoperability when working with Google AI Platform services while maintaining the consistency
// of the genai abstraction layer throughout the ADK system.
//
// # Core Conversion Categories
//
// The package provides exhaustive conversion support across all major AI interaction types:
//
//   - Content & Parts: Text, inline data, file data, function calls, responses, and video metadata
//   - Tools & Functions: Function declarations, code execution, and search retrieval capabilities
//   - Schema Definitions: JSON schemas with full type validation and constraints
//   - Generation Config: Temperature, top-p, top-k, tokens, stop sequences, and response formatting
//   - Safety & Security: Harm categories, block thresholds, and safety ratings
//   - Response Data: Candidates, usage metadata, prompt feedback, and finish reasons
//   - Advanced Features: Grounding metadata, citations, logprobs, and URL context
//
// # Utility Functions
//
// Essential pointer manipulation utilities for working with optional fields:
//
//	// Create pointers from values
//	tempPtr := aiconv.ToPtr(0.7)
//	maxTokensPtr := aiconv.ToPtr(1024)
//
//	// Safely dereference with defaults
//	temp := aiconv.Deref(tempPtr, 0.0) // Returns 0.7
//	tokens := aiconv.Deref(nil, 100)   // Returns 100 (default)
//
// # Basic Content Conversion
//
// Converting content between type systems:
//
//	// genai.Content to aiplatformpb.Content
//	genaiContent := &genai.Content{
//		Role: "user",
//		Parts: []genai.Part{
//			genai.Text("What is machine learning?"),
//		},
//	}
//	platformContent := aiconv.ToAIPlatformContent(genaiContent)
//
//	// aiplatformpb.Content to genai.Content
//	backToGenai := aiconv.FromAIPlatformContent(platformContent)
//
// # Function Call Conversion
//
// Converting function calls and responses:
//
//	// Function call conversion
//	genaiCall := &genai.FunctionCall{
//		Name: "get_weather",
//		Args: map[string]any{
//			"location": "San Francisco",
//			"units": "celsius",
//		},
//	}
//	platformCall := aiconv.ToAIPlatformFunctionCall(genaiCall)
//
//	// Function response conversion
//	genaiResponse := &genai.FunctionResponse{
//		Name: "get_weather",
//		Response: map[string]any{
//			"temperature": 20,
//			"condition": "sunny",
//		},
//	}
//	platformResponse := aiconv.ToAIPlatformFunctionResponse(genaiResponse)
//
// # Tool Configuration Conversion
//
// Converting complex tool configurations:
//
//	// Tool with function declarations
//	genaiTool := &genai.Tool{
//		FunctionDeclarations: []*genai.FunctionDeclaration{{
//			Name: "search_web",
//			Description: "Search the internet for information",
//			Parameters: &genai.Schema{
//				Type: genai.TypeObject,
//				Properties: map[string]*genai.Schema{
//					"query": {
//						Type: genai.TypeString,
//						Description: "Search query",
//					},
//				},
//				Required: []string{"query"},
//			},
//		}},
//		CodeExecution: &genai.ToolCodeExecution{},
//	}
//	platformTool := aiconv.ToAIPlatformTool(genaiTool)
//
// # Generation Configuration Conversion
//
// Converting generation parameters and constraints:
//
//	// Generation config with safety settings
//	genaiConfig := &genai.GenerationConfig{
//		Temperature:     aiconv.ToPtr(0.7),
//		TopP:           aiconv.ToPtr(0.9),
//		TopK:           aiconv.ToPtr(40),
//		MaxOutputTokens: 2048,
//		StopSequences:  []string{"END", "STOP"},
//		ResponseMIMEType: "application/json",
//		ResponseSchema: &genai.Schema{
//			Type: genai.TypeObject,
//			Properties: map[string]*genai.Schema{
//				"result": {Type: genai.TypeString},
//			},
//		},
//	}
//	platformConfig := aiconv.ToAIPlatformGenerationConfig(genaiConfig)
//
// # Safety Settings Conversion
//
// Converting safety configurations and ratings:
//
//	// Safety settings
//	genaiSafety := []*genai.SafetySetting{{
//		Category:  genai.HarmCategoryHarassment,
//		Threshold: genai.HarmBlockThresholdBlockMediumAndAbove,
//	}, {
//		Category:  genai.HarmCategoryDangerousContent,
//		Threshold: genai.HarmBlockThresholdBlockOnlyHigh,
//	}}
//	platformSafety := aiconv.ToAIPlatformSafetySettings(genaiSafety)
//
// # Response Processing
//
// Converting complete response objects:
//
//	// Full response conversion
//	genaiResponse := &genai.GenerateContentResponse{
//		Candidates: []*genai.Candidate{{
//			Index: 0,
//			Content: &genai.Content{
//				Parts: []genai.Part{genai.Text("Machine learning is...")},
//			},
//			FinishReason: genai.FinishReasonStop,
//			SafetyRatings: []*genai.SafetyRating{{
//				Category:    genai.HarmCategoryHarassment,
//				Probability: genai.HarmProbabilityNegligible,
//				Blocked:     false,
//			}},
//		}},
//		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
//			PromptTokenCount:     15,
//			CandidatesTokenCount: 42,
//			TotalTokenCount:      57,
//		},
//	}
//	platformResponse := aiconv.ToAIPlatformGenerateContentResponse(genaiResponse)
//
// # Batch Conversion
//
// Converting arrays and slices:
//
//	// Multiple contents
//	genaiContents := []*genai.Content{content1, content2, content3}
//	platformContents := aiconv.ToAIPlatformContents(genaiContents)
//
//	// Multiple tools
//	genaiTools := []*genai.Tool{tool1, tool2}
//	platformTools := aiconv.ToAIPlatformTools(genaiTools)
//
//	// Multiple candidates
//	genaiCandidates := []*genai.Candidate{candidate1, candidate2}
//	platformCandidates := aiconv.ToAIPlatformCandidates(genaiCandidates)
//
// # Advanced Features
//
// Converting advanced AI features like grounding and citations:
//
//	// Grounding metadata conversion
//	genaiGrounding := &genai.GroundingMetadata{
//		GroundingSupports: []*genai.GroundingSupport{{
//			Segment: &genai.Segment{
//				PartIndex:  0,
//				StartIndex: 10,
//				EndIndex:   25,
//				Text:       "machine learning",
//			},
//			ConfidenceScores: []float32{0.95, 0.87},
//		}},
//		WebSearchQueries: []string{"machine learning basics"},
//	}
//	platformGrounding := aiconv.ToAIPlatformGroundingMetadata(genaiGrounding)
//
//	// Citation metadata conversion
//	genaiCitations := &genai.CitationMetadata{
//		Citations: []*genai.Citation{{
//			StartIndex: 15,
//			EndIndex:   45,
//			URI:        "https://example.com/ml-guide",
//			Title:      "Machine Learning Guide",
//			License:    "CC-BY-4.0",
//		}},
//	}
//	platformCitations := aiconv.ToAIPlatformCitationMetadata(genaiCitations)
//
// # Error Handling
//
// All conversion functions handle nil inputs gracefully:
//
//	// Safe nil handling
//	var nilContent *genai.Content
//	result := aiconv.ToAIPlatformContent(nilContent) // Returns nil, no panic
//
//	// Invalid data handling
//	defer func() {
//		if r := recover(); r != nil {
//			log.Printf("Conversion error: %v", r)
//		}
//	}()
//
// The package panics on unsupported enum values or invalid data structures,
// indicating programming errors that should be caught during development.
//
// # Integration with ADK
//
// The aiconv package is primarily used internally by the model implementations
// to handle provider-specific conversions:
//
//	// In Vertex AI model implementation
//	func (m *vertexModel) GenerateContent(ctx context.Context, req *types.LLMRequest) (*types.LLMResponse, error) {
//		// Convert request to AI Platform format
//		platformContents := aiconv.ToAIPlatformContents(req.Contents)
//		platformTools := aiconv.ToAIPlatformTools(req.Tools)
//		platformConfig := aiconv.ToAIPlatformGenerationConfig(req.GenerationConfig)
//
//		// Make API call to Vertex AI
//		resp, err := m.client.GenerateContent(ctx, &aiplatformpb.GenerateContentRequest{
//			Contents:         platformContents,
//			Tools:           platformTools,
//			GenerationConfig: platformConfig,
//		})
//		if err != nil {
//			return nil, err
//		}
//
//		// Convert response back to unified format
//		genaiResp := aiconv.FromAIPlatformGenerateContentResponse(resp)
//		return &types.LLMResponse{GenerateContentResponse: genaiResp}, nil
//	}
//
// # Performance Characteristics
//
// Conversion functions are designed for efficiency:
//   - Zero-copy conversions where possible
//   - Minimal memory allocation for simple types
//   - Lazy conversion of complex nested structures
//   - Direct protobuf field mapping without intermediate serialization
//
// # Type Safety
//
// The package ensures type safety through:
//   - Compile-time type checking for all conversions
//   - Runtime validation of enum values with panics on invalid input
//   - Nil-safe operations that return nil for nil inputs
//   - Consistent bidirectional conversion behavior
//
// # Bidirectional Consistency
//
// All conversion functions maintain round-trip consistency:
//
//	// Round-trip conversion should preserve data
//	original := &genai.Content{...}
//	converted := aiconv.ToAIPlatformContent(original)
//	roundTrip := aiconv.FromAIPlatformContent(converted)
//	// roundTrip should equal original (with pointer differences)
//
// # Thread Safety
//
// All conversion functions are stateless and safe for concurrent use.
// No synchronization is required when calling conversion functions from
// multiple goroutines.
//
// The aiconv package provides the essential type system bridge that enables
// the ADK to maintain a clean, unified interface while supporting the rich
// functionality of provider-specific AI platform APIs.
package aiconv
