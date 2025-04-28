// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"google.golang.org/genai"
)

// Blob represents binary data with a MIME type.
type Blob struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// FileRef represents a reference to a file.
type FileRef struct {
	MimeType string `json:"mime_type"`
	FileURI  string `json:"file_uri"`
}

// LLMRequest represents a request to a language model.
type LLMRequest struct {
	Model              string                   `json:"model,omitempty"`
	Contents           []*genai.Content         `json:"contents"`
	Config             *genai.GenerationConfig  `json:"config,omitempty"`
	LiveConnectConfig  *genai.LiveConnectConfig `json:"live_connect_config,omitempty"`
	CountTokensConfig  *genai.CountTokensConfig `json:"count_tokens_config,omitempty"`
	SystemInstructions []string                 `json:"system_instructions,omitempty"`
	Tools              []*genai.Tool            `json:"tools,omitempty"`
	ToolMap            map[string]*genai.Tool   `json:"tool_map,omitempty"`
	SafetySettings     []*genai.SafetySetting   `json:"safety_settings,omitempty"`
	OutputSchema       map[string]any           `json:"output_schema,omitempty"`
}

// NewLLMRequest creates a new LLMRequest.
func NewLLMRequest(contents []*genai.Content) *LLMRequest {
	return &LLMRequest{
		Contents: contents,
	}
}

// UserContent creates a new user content.
func UserContent(parts ...string) *genai.Content {
	contentParts := make([]*genai.Part, 0, len(parts))
	for _, part := range parts {
		contentParts = append(contentParts, &genai.Part{Text: part})
	}
	return &genai.Content{
		Role:  "user",
		Parts: contentParts,
	}
}

// ModelContent creates a new model content.
func ModelContent(parts ...string) *genai.Content {
	contentParts := make([]*genai.Part, 0, len(parts))
	for _, part := range parts {
		contentParts = append(contentParts, &genai.Part{Text: part})
	}
	return &genai.Content{
		Role:  "model",
		Parts: contentParts,
	}
}

// AppendInstructions adds system instructions to the request.
func (r *LLMRequest) AppendInstructions(instructions ...string) *LLMRequest {
	if r.CountTokensConfig.SystemInstruction == nil {
		r.CountTokensConfig.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{
					Text: strings.Join(instructions, "\n"),
				},
			},
		}
		return r
	}

	r.CountTokensConfig.SystemInstruction.Parts = append(r.CountTokensConfig.SystemInstruction.Parts, &genai.Part{
		Text: strings.Join(instructions, "\n"),
	})
	return r
}

// AppendTools adds tools to the request.
func (r *LLMRequest) AppendTools(tools ...*genai.Tool) *LLMRequest {
	r.Tools = append(r.Tools, tools...)
	return r
}

// WithGenerationConfig sets the generation configuration.
func (r *LLMRequest) WithGenerationConfig(config *genai.GenerationConfig) *LLMRequest {
	r.Config = config
	return r
}

// WithSafetySettings sets the safety settings.
func (r *LLMRequest) WithSafetySettings(settings ...*genai.SafetySetting) *LLMRequest {
	r.SafetySettings = append(r.SafetySettings, settings...)
	return r
}

// WithModelName sets the model name.
func (r *LLMRequest) WithModelName(name string) *LLMRequest {
	r.Model = name
	return r
}

// SetOutputSchema configures the expected response format.
func (r *LLMRequest) SetOutputSchema(schema map[string]any, mimeType string) *LLMRequest {
	r.OutputSchema = schema
	if r.Config == nil {
		r.Config = &genai.GenerationConfig{}
	}
	r.Config.ResponseMIMEType = mimeType
	return r
}

// ToJSON converts the request to a JSON string.
func (r *LLMRequest) ToJSON() (string, error) {
	bytes, err := sonic.ConfigFastest.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LLMRequest to JSON: %w", err)
	}
	return string(bytes), nil
}

// ToGenerateContentConfig converts the LLMRequest to a genai.GenerateContentConfig.
func (r *LLMRequest) ToGenerateContentConfig() *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{}

	if r.Config != nil {
		// Add required fields
		config.MaxOutputTokens = int32(r.Config.MaxOutputTokens)
		config.StopSequences = r.Config.StopSequences

		// Add optional fields
		if *r.Config.Temperature > 0 {
			config.Temperature = r.Config.Temperature
		}

		if *r.Config.TopK > 0 {
			config.TopK = r.Config.TopK
		}

		if *r.Config.TopP > 0 {
			config.TopP = r.Config.TopP
		}

		if r.Config.CandidateCount > 0 {
			config.CandidateCount = int32(r.Config.CandidateCount)
		}
	}

	// Skip safety settings conversion for now due to type mismatch

	// Note: SystemInstructions might not be directly supported in the genai package
	// We might need to add them as a special content part instead

	return config
}

// ToGenaiContents converts the LLMRequest contents to genai.Content slice.
func (r *LLMRequest) ToGenaiContents() []*genai.Content {
	genaiContents := make([]*genai.Content, 0, len(r.Contents))
	for _, content := range r.Contents {
		// Create a genai.Content with text parts
		genContent := &genai.Content{
			Role:  content.Role,
			Parts: []*genai.Part{},
		}

		// Add text parts
		for _, part := range content.Parts {
			if part.Text != "" {
				genContent.Parts = append(genContent.Parts, genai.NewPartFromText(part.Text))
			}
			// Note: For simplicity, we're only handling text parts for now
		}

		genaiContents = append(genaiContents, genContent)
	}
	return genaiContents
}
