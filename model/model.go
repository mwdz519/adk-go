// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"

	"google.golang.org/genai"
)

// Role represents the role of a participant in a conversation.
type Role = string

const (
	// RoleSystem is the role of the system.
	RoleSystem Role = "system"

	// RoleAssistant is the role of the assistant.
	RoleAssistant Role = "assistant"

	// RoleUser is the role of the user.
	RoleUser Role = genai.RoleUser

	// RoleModel is the role of the model.
	RoleModel Role = genai.RoleModel
)

// Model represents a generative AI model.
type Model interface {
	// Name returns the name of the model.
	Name() string

	// Connect creates a live connection to the model.
	Connect() (BaseLLMConnection, error)

	// Generate generates content from the model.
	Generate(ctx context.Context, request *LLMRequest) (*LLMResponse, error)

	// GenerateContent generates content from the model.
	GenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (*LLMResponse, error)
}

// GenerateResponse represents a response from generating content.
type GenerateResponse struct {
	// Content is the generated content.
	Content *genai.GenerateContentResponse
}

// GenerativeModel represents a generative AI model.
type GenerativeModel interface {
	Model

	// StreamGenerate streams generated content from the model.
	StreamGenerate(ctx context.Context, request *LLMRequest) (StreamGenerateResponse, error)

	// StreamGenerateContent streams generated content from the model.
	StreamGenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (StreamGenerateResponse, error)

	// WithGenerationConfig returns a new model with the specified generation config.
	WithGenerationConfig(config *genai.GenerationConfig) GenerativeModel

	// WithSafetySettings returns a new model with the specified safety settings.
	WithSafetySettings(settings []*genai.SafetySetting) GenerativeModel
}

// StreamGenerateResponse represents a stream of generated content.
type StreamGenerateResponse interface {
	// Next returns the next response in the stream.
	Next() (*genai.GenerateContentResponse, error)
}

// BaseGenerativeModel provides a base implementation of GenerativeModel.
type BaseGenerativeModel struct {
	*Base
}

var _ GenerativeModel = (*BaseGenerativeModel)(nil)

// NewBaseGenerativeModel creates a new base generative model.
func NewBaseGenerativeModel(name string) *BaseGenerativeModel {
	return &BaseGenerativeModel{
		Base: NewBase(name),
	}
}

// WithGenerationConfig returns a new model with the specified generation config.
func (m *BaseGenerativeModel) WithGenerationConfig(config *genai.GenerationConfig) GenerativeModel {
	clone := *m
	clone.Base = m.Base.WithGenerationConfig(config)
	return &clone
}

// WithSafetySettings returns a new model with the specified safety settings.
func (m *BaseGenerativeModel) WithSafetySettings(settings []*genai.SafetySetting) GenerativeModel {
	clone := *m
	clone.Base = m.Base.WithSafetySettings(settings)
	return &clone
}

// Generate generates content from the model.
func (m *BaseGenerativeModel) Generate(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	return m.Base.Generate(ctx, request)
}

// GenerateContent generates content from the model.
func (m *BaseGenerativeModel) GenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (*LLMResponse, error) {
	return m.Base.GenerateContent(ctx, contents, config)
}

// StreamGenerate streams generated content from the model.
func (m *BaseGenerativeModel) StreamGenerate(ctx context.Context, request *LLMRequest) (StreamGenerateResponse, error) {
	return m.Base.StreamGenerate(ctx, request)
}

// StreamGenerateContent streams generated content from the model.
func (m *BaseGenerativeModel) StreamGenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (StreamGenerateResponse, error) {
	return m.Base.StreamGenerateContent(ctx, contents, config)
}
