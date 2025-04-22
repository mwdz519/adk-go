// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"fmt"
	"iter"
	"regexp"

	"google.golang.org/genai"
)

// GoogleLLM represents a Google Large Language Model (Gemini).
// It's an equivalent of the Python ADK Gemini class in google_llm.py.
type GoogleLLM struct {
	*BaseLLM

	// Default model is 'gemini-1.5-flash'
	modelName string
}

// NewGoogleLLM creates a new Google LLM instance.
func NewGoogleLLM(ctx context.Context, apiKey string, modelName string) (*GoogleLLM, error) {
	// If model name is not provided, use the default
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	// Create the client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Genai client: %w", err)
	}

	return &GoogleLLM{
		BaseLLM:   NewBaseLLM(modelName, client),
		modelName: modelName,
	}, nil
}

// SupportedModels returns a list of supported Google LLM models.
// This matches the patterns in the Python implementation.
func (m *GoogleLLM) SupportedModels() []string {
	return []string{
		// Gemini models
		`gemini-.*`,
		// Fine-tuned vertex endpoint pattern
		`projects\/.*\/locations\/.*\/endpoints\/.*`,
		// Vertex gemini long name
		`projects\/.*\/locations\/.*\/publishers\/google\/models\/gemini-.*`,
	}
}

// IsSupported checks if the given model name is supported.
func (m *GoogleLLM) IsSupported(modelName string) bool {
	for _, pattern := range m.SupportedModels() {
		match, err := regexp.MatchString(pattern, modelName)
		if err == nil && match {
			return true
		}
	}
	return false
}

// Connect creates a live connection to the Google LLM.
func (m *GoogleLLM) Connect(llmRequest *LLMRequest) (BaseLLMConnection, error) {
	llmRequest.LiveConnectConfig.SystemInstruction = &genai.Content{
		Role:  "system",
		Parts: []*genai.Part{
			// genai.NewPartFromText(llmRequest.CountTokensConfig.SystemInstruction),
		},
	}

	// Create a new GeminiLLMConnection
	conn, err := NewGeminiLLMConnection(m)
	if err != nil {
		return nil, fmt.Errorf("failed to create GeminiLLMConnection: %w", err)
	}

	return conn, nil
}

// Generate generates content from the model.
func (m *GoogleLLM) Generate(ctx context.Context, request GenerateRequest) (*GenerateResponse, error) {
	config := &genai.GenerateContentConfig{}

	// Apply generation config if provided
	if request.GenerationConfig != nil {
		config.Temperature = request.GenerationConfig.Temperature
		config.MaxOutputTokens = request.GenerationConfig.MaxOutputTokens
		config.TopP = request.GenerationConfig.TopP
		config.TopK = request.GenerationConfig.TopK
	}

	// Apply safety settings if provided
	if request.SafetySettings != nil {
		config.SafetySettings = request.SafetySettings
	}

	// Generate content
	resp, err := m.client.Models.GenerateContent(ctx, m.modelName, request.Content, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return &GenerateResponse{
		Content: resp,
	}, nil
}

// GenerateContent generates content from the model.
func (m *GoogleLLM) GenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	return m.client.Models.GenerateContent(ctx, m.modelName, contents, config)
}

// StreamGenerate streams generated content from the model.
func (m *GoogleLLM) StreamGenerate(ctx context.Context, request GenerateRequest) (GenerateStreamResponse, error) {
	config := &genai.GenerateContentConfig{}

	// Apply generation config if provided
	if request.GenerationConfig != nil {
		config.Temperature = request.GenerationConfig.Temperature
		config.MaxOutputTokens = request.GenerationConfig.MaxOutputTokens
		config.TopP = request.GenerationConfig.TopP
		config.TopK = request.GenerationConfig.TopK
	}

	// Apply safety settings if provided
	if request.SafetySettings != nil {
		config.SafetySettings = request.SafetySettings
	}

	// Generate content stream
	stream := m.client.Models.GenerateContentStream(ctx, m.modelName, request.Content, config)

	return &googleStreamResponse{
		stream: stream,
		ctx:    ctx,
	}, nil
}

// StreamGenerateContent streams generated content from the model.
func (m *GoogleLLM) StreamGenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (GenerateStreamResponse, error) {
	stream := m.client.Models.GenerateContentStream(ctx, m.modelName, contents, config)

	return &googleStreamResponse{
		stream: stream,
		ctx:    ctx,
	}, nil
}

// WithGenerationConfig returns a new model with the specified generation config.
func (m *GoogleLLM) WithGenerationConfig(config *genai.GenerationConfig) GenerativeModel {
	clone := *m
	clone.BaseLLM = m.BaseLLM.WithGenerationConfig(config)
	return &clone
}

// WithSafetySettings returns a new model with the specified safety settings.
func (m *GoogleLLM) WithSafetySettings(settings []*genai.SafetySetting) GenerativeModel {
	clone := *m
	clone.BaseLLM = m.BaseLLM.WithSafetySettings(settings)
	return &clone
}

// googleStreamResponse implements GenerateStreamResponse for Google LLM models.
type googleStreamResponse struct {
	stream  iter.Seq2[*genai.GenerateContentResponse, error]
	ctx     context.Context
	nextVal *genai.GenerateContentResponse
	nextErr error
	done    bool
}

// Next returns the next response in the stream.
func (s *googleStreamResponse) Next() (*genai.GenerateContentResponse, error) {
	if s.done {
		return nil, s.nextErr
	}

	for val, err := range s.stream {
		if err != nil {
			s.done = true
			s.nextErr = err
			return nil, err
		}
		return val, nil
	}

	s.done = true
	return nil, nil // End of stream
}

// Ensure GoogleLLM implements GenerativeModel interface.
var _ GenerativeModel = (*GoogleLLM)(nil)
