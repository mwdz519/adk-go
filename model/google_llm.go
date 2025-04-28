// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"fmt"
	"iter"
	"os"
	"strings"

	"google.golang.org/genai"
)

const (
	// GeminiLLMDefaultModel is the default model name for [GeminiLLM].
	GeminiLLMDefaultModel = "gemini-1.5-pro"

	// EnvGoogleAPIKey is the environment variable name for the Google AI API key.
	EnvGoogleAPIKey = "GOOGLE_API_KEY"
)

// Gemini represents a Google Gemini Large Language Model.
type Gemini struct {
	*Base

	genAIClient     *genai.Client
	trackingHeaders map[string]string
}

var _ GenerativeModel = (*Gemini)(nil)

// NewGemini creates a new [Gemini] instance.
func NewGemini(ctx context.Context, apiKey string, modelName string) (*Gemini, error) {
	// Use default model if none provided
	if modelName == "" {
		modelName = GeminiLLMDefaultModel
	}

	// Check API key and use [EnvGoogleAPIKey] environment variable if not provided
	if apiKey == "" {
		envApiKey := os.Getenv(EnvGoogleAPIKey)
		if envApiKey == "" {
			return nil, fmt.Errorf("either apiKey arg or %q environment variable must bu set", EnvGoogleAPIKey)
		}
		apiKey = envApiKey
	}

	// Create GenAI client for BaseLLM
	genAIClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &Gemini{
		Base:            NewBase(modelName),
		genAIClient:     genAIClient,
		trackingHeaders: make(map[string]string),
	}, nil
}

// SupportedModels returns a list of supported Gemini models.
func (m *Gemini) SupportedModels() []string {
	return []string{
		"gemini-1.0-pro",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
	}
}

// Connect creates a live connection to the Gemini LLM.
func (m *Gemini) Connect() (BaseLLMConnection, error) {
	// Ensure we have an API client
	// Create and return a new connection
	return newGeminiLLMConnection(m.model, m.genAIClient), nil
}

// maybeAppendUserContent checks if the last message is from the user and if not, appends an empty user message.
func (m *Gemini) maybeAppendUserContent(contents []*genai.Content) []*genai.Content {
	if len(contents) == 0 {
		return []*genai.Content{{
			Role:  "user",
			Parts: []*genai.Part{},
		}}
	}

	lastContent := contents[len(contents)-1]
	if strings.ToLower(lastContent.Role) != "user" {
		return append(contents, &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{},
		})
	}

	return contents
}

// Generate generates content from the model.
func (m *Gemini) Generate(ctx context.Context, request *LLMRequest) (*GenerateResponse, error) {
	// Get access to the Models service
	models := m.genAIClient.Models

	// Create config for generate content
	config := &genai.GenerateContentConfig{}

	// Apply generation config if provided
	if request.Config != nil {
		config.Temperature = request.Config.Temperature
		config.MaxOutputTokens = request.Config.MaxOutputTokens
		config.TopP = request.Config.TopP
		config.TopK = request.Config.TopK
	}

	// Apply safety settings if provided
	if request.SafetySettings != nil && len(request.SafetySettings) > 0 {
		config.SafetySettings = request.SafetySettings
	}

	// Ensure the last message is from the user
	contents := m.maybeAppendUserContent(request.Contents)

	// Generate content
	resp, err := models.GenerateContent(ctx, m.model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	return &GenerateResponse{
		Content: resp,
	}, nil
}

// GenerateContent generates content from the model.
func (m *Gemini) GenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	// Get access to the Models service
	models := m.genAIClient.Models

	// Create generate content config
	genConfig := &genai.GenerateContentConfig{}

	// Apply generation config if provided
	if config != nil {
		genConfig.MaxOutputTokens = config.MaxOutputTokens
		genConfig.Temperature = config.Temperature
		genConfig.TopP = config.TopP
		genConfig.TopK = config.TopK
	}

	// Ensure the last message is from the user
	contents = m.maybeAppendUserContent(contents)

	// Generate content
	return models.GenerateContent(ctx, m.model, contents, genConfig)
}

// StreamGenerate streams generated content from the model.
func (m *Gemini) StreamGenerate(ctx context.Context, request *LLMRequest) (StreamGenerateResponse, error) {
	// Get access to the Models service
	models := m.genAIClient.Models

	// Create config for generate content
	config := &genai.GenerateContentConfig{}

	// Apply generation config if provided
	if request.Config != nil {
		config.Temperature = request.Config.Temperature
		config.MaxOutputTokens = request.Config.MaxOutputTokens
		config.TopP = request.Config.TopP
		config.TopK = request.Config.TopK
	}

	// Apply safety settings if provided
	if request.SafetySettings != nil && len(request.SafetySettings) > 0 {
		config.SafetySettings = request.SafetySettings
	}

	// Ensure the last message is from the user
	contents := m.maybeAppendUserContent(request.Contents)

	// Stream generate content
	stream := models.GenerateContentStream(ctx, m.model, contents, config)

	return &geminiStreamResponse{
		stream: stream,
	}, nil
}

// StreamGenerateContent streams generated content from the model.
func (m *Gemini) StreamGenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (StreamGenerateResponse, error) {
	// Get access to the Models service
	models := m.genAIClient.Models

	// Create generate content config
	genConfig := &genai.GenerateContentConfig{}

	// Apply generation config if provided
	if config != nil {
		genConfig.MaxOutputTokens = config.MaxOutputTokens
		genConfig.Temperature = config.Temperature
		genConfig.TopP = config.TopP
		genConfig.TopK = config.TopK
	}

	// Ensure the last message is from the user
	contents = m.maybeAppendUserContent(contents)

	// Stream generate content
	stream := models.GenerateContentStream(ctx, m.model, contents, genConfig)

	return &geminiStreamResponse{
		stream: stream,
	}, nil
}

// WithGenerationConfig returns a new model with the specified generation config.
func (m *Gemini) WithGenerationConfig(config *genai.GenerationConfig) GenerativeModel {
	// Create a new instance to avoid modifying the original
	clone := &Gemini{
		Base:            m.Base.WithGenerationConfig(config),
		genAIClient:     m.genAIClient,
		trackingHeaders: m.trackingHeaders,
	}
	return clone
}

// WithSafetySettings returns a new model with the specified safety settings.
func (m *Gemini) WithSafetySettings(settings []*genai.SafetySetting) GenerativeModel {
	// Create a new instance to avoid modifying the original
	clone := &Gemini{
		Base:            m.Base.WithSafetySettings(settings),
		genAIClient:     m.genAIClient,
		trackingHeaders: m.trackingHeaders,
	}
	return clone
}

// geminiStreamResponse implements [StreamGenerateResponse] for [Gemini].
type geminiStreamResponse struct {
	stream    iter.Seq2[*genai.GenerateContentResponse, error]
	streamIdx int
}

var _ StreamGenerateResponse = (*geminiStreamResponse)(nil)

// Next returns the next response in the stream.
func (s *geminiStreamResponse) Next() (*genai.GenerateContentResponse, error) {
	// With [iter.Seq2], we need to use a for loop with a single iteration
	// to get the next value and error
	for resp, err := range s.stream {
		return resp, err
	}

	// If the iterator is empty, return nil
	return nil, nil
}
