// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"log/slog"

	"google.golang.org/genai"
)

// Config represents a base implementation of a Large Language Model.
// It's an equivalent of the Python ADK BaseLlm class.
type Config struct {
	// generationConfig contains configuration for generation.
	generationConfig *genai.GenerationConfig

	// safetySettings contains safety settings for content generation.
	safetySettings []*genai.SafetySetting

	// logger is the logger used for logging.
	logger *slog.Logger
}

func newConfig() Config {
	return Config{
		logger: slog.Default(),
	}
}

// Option is a function that modifies the [Config] model.
type Option interface {
	apply(base Config) Config
}

type generationConfigOption struct{ *genai.GenerationConfig }

func (o generationConfigOption) apply(base Config) Config {
	base.generationConfig = o.GenerationConfig
	return base
}

// WithGenerationConfig sets the generation configuration for the Base model.
func WithGenerationConfig(config *genai.GenerationConfig) Option {
	return generationConfigOption{config}
}

type safetySettingOption []*genai.SafetySetting

func (o safetySettingOption) apply(base Config) Config {
	base.safetySettings = append(base.safetySettings, o...)
	return base
}

// WithSafetySettings sets the safety settings for the Base model.
func WithSafetySettings(settings []*genai.SafetySetting) Option {
	return safetySettingOption(settings)
}

type loggerOption struct{ *slog.Logger }

func (o loggerOption) apply(base Config) Config {
	base.logger = o.Logger
	return base
}

// WithLogger sets the logger for the Base model.
func WithLogger(logger *slog.Logger) Option {
	return loggerOption{logger}
}
