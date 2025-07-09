// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"

	"google.golang.org/genai"

	adk "github.com/go-a2a/adk-go"
	"github.com/go-a2a/adk-go/types"
)

const (
	// GeminiLLMDefaultModel is the default model name for [Gemini].
	GeminiLLMDefaultModel = "gemini-1.5-pro"

	// EnvGoogleAPIKey is the environment variable name for the Google AI API key.
	EnvGoogleAPIKey = "GOOGLE_API_KEY"
)

// Gemini represents a Google Gemini Large Language Model.
type Gemini struct {
	*BaseLLM

	genAIClient *genai.Client
}

var _ types.Model = (*Gemini)(nil)

// NewGemini creates a new [Gemini] instance.
func NewGemini(ctx context.Context, apiKey, modelName string, opts ...Option) (*Gemini, error) {
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

	clintConfig := &genai.ClientConfig{
		APIKey: apiKey,
		HTTPOptions: genai.HTTPOptions{
			Headers: make(http.Header),
		},
	}

	frameworkLabel := fmt.Sprintf("go-a2a/adk-go/%s", adk.Version)
	languageLabel := fmt.Sprintf("go/%s", runtime.Version())
	versionHeaderValue := frameworkLabel + " " + languageLabel
	clintConfig.HTTPOptions.Headers.Set(`x-goog-api-client`, versionHeaderValue)
	clintConfig.HTTPOptions.Headers.Set(`user-agent`, versionHeaderValue)

	// Create GenAI client
	genAIClient, err := genai.NewClient(ctx, clintConfig)
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}

	gemini := &Gemini{
		BaseLLM:     NewBaseLLM(modelName),
		genAIClient: genAIClient,
	}
	for _, opt := range opts {
		gemini.Config = opt.apply(gemini.Config)
	}

	return gemini, nil
}

// Name returns the name of the [Gemini] model.
func (m *Gemini) Name() string {
	return m.modelName
}

// SupportedModels returns a list of supported models in the [Gemini].
//
// See https://ai.google.dev/gemini-api/docs/models.
func (m *Gemini) SupportedModels() []string {
	return []string{
		"gemini-2.5-pro",
		"gemini-2.5-flash",
		"gemini-2.5-flash-lite-preview-06-17",
		"gemini-2.5-flash-preview-native-audio-dialog",
		"gemini-2.5-flash-exp-native-audio-thinking-dialog",
		"gemini-2.5-flash-preview-tts",
		"gemini-2.5-pro-preview-tts",
		"gemini-2.0-flash",
		"gemini-2.0-flash-preview-image-generation",
		"gemini-2.0-flash-lite",
		"gemini-1.5-flash",
		"gemini-1.5-flash-8b",
		"gemini-1.5-pro",
		"imagen-4.0-generate-preview-06-06",
		"imagen-4.0-ultra-generate-preview-06-06",
		"imagen-3.0-generate-002",
		"veo-2.0-generate-001",
		"gemini-live-2.5-flash-preview",
		"gemini-2.0-flash-live-001",
	}
}

// Connect creates a live connection to the Gemini LLM.
func (m *Gemini) Connect(ctx context.Context, _ *types.LLMRequest) (types.ModelConnection, error) {
	// Create and return a new connection
	return newGeminiConnection(ctx, m.modelName, m.genAIClient), nil
}

// GenerateContent generates content from the model.
func (m *Gemini) GenerateContent(ctx context.Context, request *types.LLMRequest) (*types.LLMResponse, error) {
	// Ensure the last message is from the user
	request.Contents = m.appendUserContent(request.Contents)

	// Generate content
	response, err := m.genAIClient.Models.GenerateContent(ctx, m.modelName, request.Contents, request.Config)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}
	m.logger.DebugContext(ctx, "response", buildResponseLog(response))

	return types.CreateLLMResponse(response), nil
}

// StreamGenerateContent streams generated content from the model.
func (m *Gemini) StreamGenerateContent(ctx context.Context, request *types.LLMRequest) iter.Seq2[*types.LLMResponse, error] {
	return func(yield func(*types.LLMResponse, error) bool) {
		// Ensure the last message is from the user
		contents := m.appendUserContent(request.Contents)

		// Stream generate content
		stream := m.genAIClient.Models.GenerateContentStream(ctx, m.modelName, contents, request.Config)

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
			}

			if ctx.Err() != nil || resp == nil {
				return
			}

			lastResp = resp
			llmResp := types.CreateLLMResponse(resp)

			switch {
			case containsText(llmResp):
				buf.WriteString(llmResp.Content.Parts[0].Text)
				llmResp.WithPartial(true)

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

func newAggregateText(s string) *types.LLMResponse {
	return &types.LLMResponse{
		Content: &genai.Content{
			Role:  RoleModel,
			Parts: []*genai.Part{genai.NewPartFromText(s)},
		},
	}
}

// containsText returns true when the first part has a non-empty Text field.
func containsText(r *types.LLMResponse) bool {
	return r.Content != nil && len(r.Content.Parts) > 0 && r.Content.Parts[0].Text != ""
}

// isAudio returns true when InlineData is present (optionally mime-typed audio/*).
func isAudio(r *types.LLMResponse) bool {
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
