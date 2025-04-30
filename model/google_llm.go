// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
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
	logger          *slog.Logger
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
		logger:          slog.Default(),
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

// SupportedModels returns a list of supported Gemini models.
//
// See https://ai.google.dev/gemini-api/docs/models.
func (m *Gemini) SupportedModels() []string {
	return []string{
		"gemini-2.5-flash-preview-04-17",
		"gemini-2.5-pro-preview-03-25",
		"gemini-2.0-flash",
		"gemini-2.0-flash-lite",
		"gemini-1.5-flash,",
		"gemini-1.5-flash-8b",
		"gemini-1.5-pro",
	}
}

// Connect creates a live connection to the Gemini LLM.
func (m *Gemini) Connect() (BaseLLMConnection, error) {
	// Ensure we have an API client
	// Create and return a new connection
	return newGeminiLLMConnection(m.model, m.genAIClient), nil
}

// appendUserContent checks if the last message is from the user and if not, appends an empty user message.
func (m *Gemini) appendUserContent(contents []*genai.Content) []*genai.Content {
	switch {
	case len(contents) == 0:
		return append(contents, &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				genai.NewPartFromText(`Handle the requests as specified in the System Instruction.`),
			},
		})

	case strings.ToLower(contents[len(contents)-1].Role) != genai.RoleUser:
		return append(contents, &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				genai.NewPartFromText(`Continue processing previous requests as instructed. Exit or provide a summary if no more outputs are needed.`),
			},
		})

	default:
		return contents
	}
}

// Generate generates content from the model.
func (m *Gemini) Generate(ctx context.Context, request *LLMRequest) (*LLMResponse, error) {
	// Ensure the last message is from the user
	request.Contents = m.appendUserContent(request.Contents)

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
	if len(request.SafetySettings) > 0 {
		config.SafetySettings = request.SafetySettings
	}

	// Apply tool if provided
	if len(request.Tools) > 0 {
		config.Tools = request.Tools
	}

	// Generate content
	response, err := m.genAIClient.Models.GenerateContent(ctx, m.model, request.Contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}
	m.logger.DebugContext(ctx, "response", buildResponseLog(response))

	return CreateLLMResponse(response), nil
}

// GenerateContent generates content from the model.
func (m *Gemini) GenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (*LLMResponse, error) {
	// Ensure the last message is from the user
	contents = m.appendUserContent(contents)

	// Create generate content config
	genConfig := &genai.GenerationConfig{}
	// Apply generation config if provided
	if config != nil {
		genConfig.MaxOutputTokens = config.MaxOutputTokens
		genConfig.Temperature = config.Temperature
		genConfig.TopP = config.TopP
		genConfig.TopK = config.TopK
	}

	request := &LLMRequest{
		Contents: contents,
		Config:   genConfig,
	}

	return m.Generate(ctx, request)
}

// StreamGenerate streams generated content from the model.
func (m *Gemini) StreamGenerate(ctx context.Context, request *LLMRequest) iter.Seq2[*LLMResponse, error] {
	return func(yield func(*LLMResponse, error) bool) {
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
		if len(request.SafetySettings) > 0 {
			config.SafetySettings = request.SafetySettings
		}

		// Apply tool if provided
		if len(request.Tools) > 0 {
			config.Tools = request.Tools
		}

		// Ensure the last message is from the user
		contents := m.appendUserContent(request.Contents)

		// Stream generate content
		stream := m.genAIClient.Models.GenerateContentStream(ctx, m.model, contents, config)

		var (
			buf      strings.Builder
			lastResp *genai.GenerateContentResponse
		)
		for resp, err := range stream {
			// catch error first
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}

			if ctx.Err() != nil || resp == nil {
				return
			}

			lastResp = resp
			llmResp := CreateLLMResponse(resp)

			switch {
			case containsText(llmResp):
				buf.WriteString(llmResp.Content.Parts[0].Text)
				llmResp.Partial = true

			case buf.Len() > 0 && !isAudio(llmResp):
				if !yield(newAggregateText(buf.String()), nil) {
					return
				}
				buf.Reset()
			}

			if !yield(llmResp, nil) {
				return
			}
		}

		if buf.Len() > 0 && lastResp != nil && finishStop(lastResp) {
			yield(newAggregateText(buf.String()), nil)
		}
	}
}

// StreamGenerateContent streams generated content from the model.
func (m *Gemini) StreamGenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) iter.Seq2[*LLMResponse, error] {
	return func(yield func(*LLMResponse, error) bool) {
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
		contents = m.appendUserContent(contents)

		// Stream generate content
		stream := m.genAIClient.Models.GenerateContentStream(ctx, m.model, contents, genConfig)

		var (
			buf      strings.Builder
			lastResp *genai.GenerateContentResponse
		)
		for resp, err := range stream {
			// catch error first
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}

			if ctx.Err() != nil || resp == nil {
				return
			}

			lastResp = resp
			llmResp := CreateLLMResponse(resp)

			switch {
			case containsText(llmResp):
				buf.WriteString(llmResp.Content.Parts[0].Text)
				llmResp.Partial = true

			case buf.Len() > 0 && !isAudio(llmResp):
				if !yield(newAggregateText(buf.String()), nil) {
					return
				}
				buf.Reset()
			}

			if !yield(llmResp, nil) {
				return
			}
		}

		if buf.Len() > 0 && lastResp != nil && finishStop(lastResp) {
			yield(newAggregateText(buf.String()), nil)
		}
	}
}

// geminiStreamResponse implements [StreamGenerateResponse] for [Gemini].
type geminiStreamResponse struct {
	stream iter.Seq2[*genai.GenerateContentResponse, error]
}

var _ StreamGenerateResponse = (*geminiStreamResponse)(nil)

// Next returns the next response in the stream.
func (s *geminiStreamResponse) Next(ctx context.Context) iter.Seq2[*LLMResponse, error] {
	// With [iter.Seq2], we need to use a for loop with a single iteration
	// to get the next value and error
	return func(yield func(*LLMResponse, error) bool) {
		var (
			buf      strings.Builder
			lastResp *genai.GenerateContentResponse
		)
		for resp, err := range s.stream {
			// catch error first
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}

			if ctx.Err() != nil || resp == nil {
				return
			}

			lastResp = resp
			llmResp := CreateLLMResponse(resp)

			switch {
			case containsText(llmResp):
				buf.WriteString(llmResp.Content.Parts[0].Text)
				llmResp.Partial = true

			case buf.Len() > 0 && !isAudio(llmResp):
				if !yield(newAggregateText(buf.String()), nil) {
					return
				}
				buf.Reset()
			}

			if !yield(llmResp, nil) {
				return
			}
		}

		if buf.Len() > 0 && lastResp != nil && finishStop(lastResp) {
			yield(newAggregateText(buf.String()), nil)
		}
	}
}

func newAggregateText(s string) *LLMResponse {
	return &LLMResponse{
		Content: &genai.Content{
			Role:  RoleModel,
			Parts: []*genai.Part{genai.NewPartFromText(s)},
		},
	}
}

// containsText returns true when the first part has a non-empty Text field.
func containsText(r *LLMResponse) bool {
	return r.Content != nil && len(r.Content.Parts) > 0 && r.Content.Parts[0].Text != ""
}

// isAudio returns true when InlineData is present (optionally mime-typed audio/*).
func isAudio(r *LLMResponse) bool {
	if r.Content == nil || len(r.Content.Parts) == 0 {
		return false
	}
	if data := r.Content.Parts[0].InlineData; data != nil {
		if data.MIMEType == "" {
			return true
		}
		return strings.HasPrefix(data.MIMEType, "audio/")
	}
	return false
}

// finishStop reports whether the first candidate finished with STOP.
func finishStop(r *genai.GenerateContentResponse) bool {
	return r != nil && len(r.Candidates) > 0 && r.Candidates[0].FinishReason == genai.FinishReasonStop
}

const repponseLogFmt = `
LLM Response:
-----------------------------------------------------------
Text:
%s
-----------------------------------------------------------
Function calls:
%s
-----------------------------------------------------------
`

func buildResponseLog(resp *genai.GenerateContentResponse) slog.Attr {
	functionCalls := resp.FunctionCalls()
	functionCallsText := make([]string, len(functionCalls))
	for i, funcCall := range functionCalls {
		functionCallsText[i] = fmt.Sprintf("name: %s, args: %s", funcCall.Name, funcCall.Args)
	}

	return slog.String("response", fmt.Sprintf(repponseLogFmt, resp.Text(), strings.Join(functionCallsText, "\n")))
}
