// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// Base represents a base implementation of a Large Language Model.
// It's an equivalent of the Python ADK BaseLlm class.
type Base struct {
	// Model represents the specific LLM model name.
	model string

	// GenerationConfig contains configuration for generation.
	generationConfig *genai.GenerationConfig

	// SafetySettings contains safety settings for content generation.
	safetySettings []*genai.SafetySetting
}

var _ Model = (*Base)(nil)

// NewBase creates a new BaseLLM instance.
func NewBase(model string) *Base {
	return &Base{
		model: model,
	}
}

// Name returns the name of the model.
func (m *Base) Name() string {
	return m.model
}

// SupportedModels returns a list of supported models.
// This method should be overridden by specific LLM implementations.
func (m *Base) SupportedModels() []string {
	return []string{}
}

// WithGenerationConfig returns a new model with the specified generation config.
func (m *Base) WithGenerationConfig(config *genai.GenerationConfig) *Base {
	clone := *m
	clone.generationConfig = config
	return &clone
}

// WithSafetySettings returns a new model with the specified safety settings.
func (m *Base) WithSafetySettings(settings []*genai.SafetySetting) *Base {
	clone := *m
	clone.safetySettings = settings
	return &clone
}

// Connect creates a live connection to the LLM.
// This method should be overridden by specific LLM implementations.
func (m *Base) Connect() (BaseLLMConnection, error) {
	return nil, fmt.Errorf("Connect not implemented for BaseLLM")
}

// Generate generates content from the model.
func (m *Base) Generate(ctx context.Context, request GenerateRequest) (*GenerateResponse, error) {
	return nil, fmt.Errorf("Generate not implemented for BaseLLM")
}

// GenerateContent generates content from the model.
func (m *Base) GenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	return nil, fmt.Errorf("GenerateContent not implemented for BaseLLM")
}

// StreamGenerate streams generated content from the model.
func (m *Base) StreamGenerate(ctx context.Context, request GenerateRequest) (StreamGenerateResponse, error) {
	return nil, fmt.Errorf("StreamGenerate not implemented for BaseLLM")
}

// StreamGenerateContent streams generated content from the model.
func (m *Base) StreamGenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (StreamGenerateResponse, error) {
	return nil, fmt.Errorf("StreamGenerateContent not implemented for BaseLLM")
}
