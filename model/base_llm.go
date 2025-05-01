// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"

	"google.golang.org/genai"
)

// Option is a function that modifies the Base model.
type Option func(*Base)

// WithGenerationConfig sets the generation configuration for the Base model.
func (m *Base) WithGenerationConfig(config *genai.GenerationConfig) Option {
	return func(base *Base) {
		base.generationConfig = config
	}
}

// WithSafetySettings sets the safety settings for the Base model.
func (m *Base) WithSafetySettings(settings []*genai.SafetySetting) Option {
	return func(base *Base) {
		base.safetySettings = settings
	}
}

// WithLogger sets the logger for the Base model.
func (m *Base) WithLogger(logger *slog.Logger) Option {
	return func(base *Base) {
		base.logger = logger
	}
}

// Base represents a base implementation of a Large Language Model.
// It's an equivalent of the Python ADK BaseLlm class.
type Base struct {
	// model represents the specific LLM model name.
	model string

	// generationConfig contains configuration for generation.
	generationConfig *genai.GenerationConfig

	// safetySettings contains safety settings for content generation.
	safetySettings []*genai.SafetySetting

	// logger is the logger used for logging.
	logger *slog.Logger
}

var _ Model = (*Base)(nil)

// NewBase creates a new [Base] instance.
func NewBase(model string, opts ...Option) *Base {
	base := &Base{
		model:  model,
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(base)
	}

	return base
}

// Name returns the name of the model.
func (m *Base) Name() string {
	return m.model
}

// SupportedModels returns a list of supported models.
func (m *Base) SupportedModels() []string {
	return []string{}
}

// Connect creates a live connection to the LLM.
func (m *Base) Connect() (BaseConnection, error) {
	return nil, fmt.Errorf("Connect not implemented for BaseLLM")
}

// GenerateContent generates content from the model.
func (m *Base) GenerateContent(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	return nil, fmt.Errorf("Generate not implemented for BaseLLM")
}

// StreamGenerateContent streams generated content from the model.
func (m *Base) StreamGenerateContent(ctx context.Context, request *LLMRequest) iter.Seq2[*LLMResponse, error] {
	return func(yield func(*LLMResponse, error) bool) {
		if !yield(nil, errors.New("Base: not implemented yet")) {
			return
		}
	}
}
