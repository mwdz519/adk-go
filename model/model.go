// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"iter"

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
	Connect() (BaseConnection, error)

	// GenerateContent generates content from the model.
	GenerateContent(ctx context.Context, request *LLMRequest) (*LLMResponse, error)
}

// GenerativeModel represents a generative AI model.
type GenerativeModel interface {
	Model

	// StreamGenerateContent streams generated content from the model.
	StreamGenerateContent(ctx context.Context, request *LLMRequest) iter.Seq2[*LLMResponse, error]
}

// StreamGenerateResponse represents a stream of generated content.
type StreamGenerateResponse interface {
	// Next returns the next response in the stream.
	Next(context.Context) iter.Seq2[*LLMResponse, error]
}
