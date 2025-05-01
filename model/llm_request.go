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
	Model             string                       `json:"model,omitempty"`
	Contents          []*genai.Content             `json:"contents"`
	Config            *genai.GenerateContentConfig `json:"config,omitempty"`
	LiveConnectConfig *genai.LiveConnectConfig     `json:"live_connect_config,omitempty"`
	ToolMap           map[string]*genai.Tool       `json:"tool_map,omitempty"`
	OutputSchema      map[string]any               `json:"output_schema,omitempty"`
}

// NewLLMRequest creates a new LLMRequest.
func NewLLMRequest(contents []*genai.Content) *LLMRequest {
	return &LLMRequest{
		Contents: contents,
	}
}

// UserContent creates a new user content.
func UserContent(texts ...string) *genai.Content {
	contentParts := make([]*genai.Part, len(texts))
	for i, part := range texts {
		contentParts[i] = &genai.Part{Text: part}
	}
	return &genai.Content{
		Role:  "user",
		Parts: contentParts,
	}
}

// ModelContent creates a new model content.
func ModelContent(texts ...string) *genai.Content {
	contentParts := make([]*genai.Part, len(texts))
	for i, part := range texts {
		contentParts[i] = &genai.Part{Text: part}
	}
	return &genai.Content{
		Role:  "model",
		Parts: contentParts,
	}
}

// AppendInstructions adds system instructions to the request.
func (r *LLMRequest) AppendInstructions(instructions ...string) *LLMRequest {
	if r.Config == nil {
		r.Config = &genai.GenerateContentConfig{}
	}
	if r.Config.SystemInstruction == nil {
		r.Config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{
					Text: strings.Join(instructions, "\n"),
				},
			},
		}
		return r
	}

	r.Config.SystemInstruction.Parts = append(r.Config.SystemInstruction.Parts, &genai.Part{
		Text: strings.Join(instructions, "\n"),
	})

	return r
}

// AppendTools adds tools to the request.
func (r *LLMRequest) AppendTools(tools ...*genai.Tool) *LLMRequest {
	if r.Config == nil {
		r.Config = &genai.GenerateContentConfig{}
	}
	r.Config.Tools = append(r.Config.Tools, tools...)
	return r
}

// WithGenerationConfig sets the generation configuration.
func (r *LLMRequest) WithGenerationConfig(config *genai.GenerateContentConfig) *LLMRequest {
	r.Config = config
	return r
}

// WithSafetySettings sets the safety settings.
func (r *LLMRequest) WithSafetySettings(settings ...*genai.SafetySetting) *LLMRequest {
	if r.Config == nil {
		r.Config = &genai.GenerateContentConfig{}
	}
	r.Config.SafetySettings = append(r.Config.SafetySettings, settings...)
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
		r.Config = &genai.GenerateContentConfig{}
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
