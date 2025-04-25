// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
	"google.golang.org/genai"
)

const (
	// ClaudeLLMDefaultModel is the default model name for [ClaudeLLM].
	//
	// This model is only available on Google Cloud Platform (GCP) Vertex AI.
	// If you want to use the Anthropic official model, pass any model name that is defined in the
	// anthropic-sdk-go package's constants to [NewClaudeLLM].
	ClaudeLLMDefaultModel = "claude-3-5-sonnet-v2@20241022"

	// EnvAnthropicAPIKey is the environment variable name for the Anthropic API key.
	EnvAnthropicAPIKey = "ANTHROPIC_API_KEY"
)

// Claude represents a Claude Large Language Model.
type Claude struct {
	*Base

	anthropicClient anthropic.Client
}

var _ GenerativeModel = (*Claude)(nil)

// NewClaude creates a new Claude LLM instance.
func NewClaude(ctx context.Context, apiKey string, modelName string) (*Claude, error) {
	// Check API key and use [EnvAnthropicAPIKey] environment variable if not provided
	if apiKey == "" {
		envApiKey := os.Getenv(EnvAnthropicAPIKey)
		if envApiKey == "" {
			return nil, fmt.Errorf("either apiKey arg or %q environment variable must bu set", EnvAnthropicAPIKey)
		}
		apiKey = envApiKey
	}

	// Use default model if none provided
	if modelName == "" {
		modelName = ClaudeLLMDefaultModel
	}

	anthropicClient := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &Claude{
		Base:            NewBase(modelName),
		anthropicClient: anthropicClient,
	}, nil
}

// SupportedModels returns a list of supported Claude models.
func (m *Claude) SupportedModels() []string {
	return []string{
		// GCP Vertex AI
		"claude-3-7-sonnet@20250219",
		"claude-3-5-haiku@20241022",
		"claude-3-5-sonnet-v2@20241022",
		"claude-3-opus@20240229",
		"claude-3-sonnet@20240229",
		"claude-3-haiku@20240307",

		// Anthropic API
		anthropic.ModelClaude3_7SonnetLatest,
		anthropic.ModelClaude3_7Sonnet20250219,
		anthropic.ModelClaude3_5HaikuLatest,
		anthropic.ModelClaude3_5Haiku20241022,
		anthropic.ModelClaude3_5SonnetLatest,
		anthropic.ModelClaude3_5Sonnet20241022,
		anthropic.ModelClaude_3_5_Sonnet_20240620,
		anthropic.ModelClaude3OpusLatest,
		anthropic.ModelClaude_3_Opus_20240229,
	}
}

// Connect creates a live connection to the Claude LLM.
//
// TODO(zchee): implements.
func (m *Claude) Connect() (BaseLLMConnection, error) {
	// Ensure we can get an Anthropic client
	_ = m.anthropicClient

	// For now, this is a placeholder as we haven't implemented ClaudeLLMConnection yet
	// In a real implementation, we would return a proper ClaudeLLMConnection
	return nil, fmt.Errorf("ClaudeLLMConnection not implemented yet")
}

// extractSystemPrompt extracts system prompt text from the first message if it's a system message
func extractSystemPrompt(messages []*genai.Content) (string, bool) {
	if len(messages) == 0 || messages[0].Role != RoleSystem {
		return "", false
	}

	systemText := ""
	for _, part := range messages[0].Parts {
		if part != nil && part.Text != "" {
			systemText += part.Text
		}
	}
	return systemText, systemText != ""
}

// extractFunctionDeclarations extracts function declarations from content parts
func extractFunctionDeclarations(contents []*genai.Content) []anthropic.ToolUnionParam {
	var tools []anthropic.ToolUnionParam

	for _, content := range contents {
		if content.Parts == nil {
			continue
		}

		for _, part := range content.Parts {
			if part != nil && part.FunctionCall != nil {
				toolSchema := anthropic.ToolInputSchemaParam{
					Type:       "object",
					Properties: make(map[string]any),
				}

				// Create a tool from function
				tool := anthropic.ToolUnionParamOfTool(toolSchema, part.FunctionCall.Name)
				tools = append(tools, tool)
			}
		}
	}

	return tools
}

// Generate generates content from the model.
func (m *Claude) Generate(ctx context.Context, request GenerateRequest) (*GenerateResponse, error) {
	// Convert messages to Anthropic format
	messageParams := make([]anthropic.MessageParam, len(request.Content))
	for i, content := range request.Content {
		messageParams[i] = contentToClaudeMessageParam(content)
	}

	// Prepare parameters
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(m.model),
		Messages:  messageParams,
		MaxTokens: 4096,
	}

	// Apply generation config if provided
	if request.GenerationConfig != nil {
		if request.GenerationConfig.Temperature != nil {
			params.Temperature = anthropic.Float(float64(*request.GenerationConfig.Temperature))
		}

		// MaxOutputTokens is an int32 directly, not a pointer
		if request.GenerationConfig.MaxOutputTokens > 0 {
			params.MaxTokens = int64(request.GenerationConfig.MaxOutputTokens)
		}

		if request.GenerationConfig.TopP != nil {
			params.TopP = anthropic.Float(float64(*request.GenerationConfig.TopP))
		}
	}

	// Apply system prompt if it exists in first content
	systemText, hasSystem := extractSystemPrompt(request.Content)
	if hasSystem {
		// For System, we need to create TextBlockParam
		var systemTextBlocks []anthropic.TextBlockParam
		systemTextBlocks = append(systemTextBlocks, anthropic.TextBlockParam{
			Text: systemText,
			Type: constant.ValueOf[constant.Text]().Default(),
		})
		params.System = systemTextBlocks

		// Remove system message from the message list since it's set separately
		// Only if there are more than one message, otherwise we keep the empty list
		if len(messageParams) > 1 {
			messageParams = messageParams[1:]
			params.Messages = messageParams
		}
	}

	// Add tools if provided
	toolDeclarations := extractFunctionDeclarations(request.Content)
	if len(toolDeclarations) > 0 {
		params.Tools = toolDeclarations
	}

	// Make API call
	message, err := m.anthropicClient.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	// Convert response to GenAI format
	content := anthropicMessageToGenAIContent(message)

	// Create response
	response := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: content,
			},
		},
	}

	return &GenerateResponse{
		Content: response,
	}, nil
}

// anthropicMessageToGenAIContent converts an Anthropic message to GenAI content
func anthropicMessageToGenAIContent(message *anthropic.Message) *genai.Content {
	var parts []*genai.Part

	// Convert content blocks to parts
	for _, block := range message.Content {
		if block.Type == "text" {
			// Handle text content - Text is a string
			parts = append(parts, genai.NewPartFromText(block.Text))
		}
	}

	// Create a new content with "model" (in anthropic, called "assistant") role
	return &genai.Content{
		Role:  RoleModel,
		Parts: parts,
	}
}

// GenerateContent generates content from the model.
func (m *Claude) GenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	request := GenerateRequest{
		Content: contents,
	}

	if config != nil {
		genConfig := &genai.GenerationConfig{}
		if config.Temperature != nil {
			genConfig.Temperature = config.Temperature
		}

		// MaxOutputTokens is an int32 directly, not a pointer
		genConfig.MaxOutputTokens = config.MaxOutputTokens

		if config.TopP != nil {
			genConfig.TopP = config.TopP
		}
		if config.TopK != nil {
			genConfig.TopK = config.TopK
		}
		request.GenerationConfig = genConfig
		request.SafetySettings = config.SafetySettings
	}

	resp, err := m.Generate(ctx, request)
	if err != nil {
		return nil, err
	}

	return resp.Content, nil
}

// StreamGenerate streams generated content from the model.
func (m *Claude) StreamGenerate(ctx context.Context, request GenerateRequest) (StreamGenerateResponse, error) {
	// Convert to Anthropic format
	msgParams := make([]anthropic.MessageParam, len(request.Content))
	for i, content := range request.Content {
		msgParams[i] = contentToClaudeMessageParam(content)
	}

	// Prepare parameters
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(m.model),
		Messages:  msgParams,
		MaxTokens: 4096,
	}

	// Apply generation config if provided
	if request.GenerationConfig != nil {
		if request.GenerationConfig.Temperature != nil {
			params.Temperature = anthropic.Float(float64(*request.GenerationConfig.Temperature))
		}

		// MaxOutputTokens is an int32 directly, not a pointer
		if request.GenerationConfig.MaxOutputTokens > 0 {
			params.MaxTokens = int64(request.GenerationConfig.MaxOutputTokens)
		}

		if request.GenerationConfig.TopP != nil {
			params.TopP = anthropic.Float(float64(*request.GenerationConfig.TopP))
		}
	}

	// Apply system prompt if it exists in first content
	systemText, hasSystem := extractSystemPrompt(request.Content)
	if hasSystem {
		// For System, we need to create TextBlockParam
		var systemTextBlocks []anthropic.TextBlockParam
		systemTextBlocks = append(systemTextBlocks, anthropic.TextBlockParam{
			Text: systemText,
			Type: constant.Text("text"),
		})
		params.System = systemTextBlocks

		// Remove system message from the message list since it's set separately
		// Only if there are more than one message, otherwise we keep the empty list
		if len(msgParams) > 1 {
			msgParams = msgParams[1:]
			params.Messages = msgParams
		}
	}

	// Add tools if provided
	toolDeclarations := extractFunctionDeclarations(request.Content)
	if len(toolDeclarations) > 0 {
		params.Tools = toolDeclarations
	}

	// Make streaming API call - stream parameter is added by the method
	stream := m.anthropicClient.Messages.NewStreaming(ctx, params)

	return &claudeStreamResponse{
		stream: stream,
		ctx:    ctx,
	}, nil
}

// StreamGenerateContent streams generated content from the model.
func (m *Claude) StreamGenerateContent(ctx context.Context, contents []*genai.Content, config *genai.GenerateContentConfig) (StreamGenerateResponse, error) {
	request := GenerateRequest{
		Content: contents,
	}

	if config != nil {
		genConfig := &genai.GenerationConfig{}
		if config.Temperature != nil {
			genConfig.Temperature = config.Temperature
		}

		// MaxOutputTokens is an int32 directly, not a pointer
		genConfig.MaxOutputTokens = config.MaxOutputTokens

		if config.TopP != nil {
			genConfig.TopP = config.TopP
		}
		if config.TopK != nil {
			genConfig.TopK = config.TopK
		}
		request.GenerationConfig = genConfig
		request.SafetySettings = config.SafetySettings
	}

	return m.StreamGenerate(ctx, request)
}

// WithGenerationConfig returns a new model with the specified generation config.
func (m *Claude) WithGenerationConfig(config *genai.GenerationConfig) GenerativeModel {
	// Create a new instance to avoid copying sync.Once
	return &Claude{
		Base:            m.Base.WithGenerationConfig(config),
		anthropicClient: m.anthropicClient,
	}
}

// WithSafetySettings returns a new model with the specified safety settings.
func (m *Claude) WithSafetySettings(settings []*genai.SafetySetting) GenerativeModel {
	// Create a new instance to avoid copying sync.Once
	return &Claude{
		Base:            m.Base.WithSafetySettings(settings),
		anthropicClient: m.anthropicClient,
	}
}

// ClaudeRequest contains the request parameters for Claude models.
type ClaudeRequest struct {
	SystemInstruction string
	Messages          []*genai.Content
	Tools             []*genai.Tool
}

// claudeStreamResponse implements GenerateStreamResponse for Claude models.
type claudeStreamResponse struct {
	stream  *ssestream.Stream[anthropic.MessageStreamEventUnion]
	ctx     context.Context
	message anthropic.Message
	done    bool
	nextErr error
}

// Next returns the next response in the stream.
func (s *claudeStreamResponse) Next() (*genai.GenerateContentResponse, error) {
	if s.done {
		return nil, s.nextErr
	}

	if !s.stream.Next() {
		s.done = true
		if err := s.stream.Err(); err != nil {
			s.nextErr = err
			return nil, err
		}
		// End of stream
		return nil, nil
	}

	// Get the current event
	event := s.stream.Current()

	// Accumulate the response
	if err := s.message.Accumulate(event); err != nil {
		log.Printf("Error accumulating message: %v", err)
	}

	// Create partial response
	var parts []*genai.Part

	// Process based on event type
	switch event.Type {
	case "content_block_delta":
		// Extract delta text from content block delta
		blockDeltaEvent := event.AsContentBlockDeltaEvent()
		if blockDeltaEvent.Delta.Type == "text_delta" {
			parts = append(parts, genai.NewPartFromText(blockDeltaEvent.Delta.Text))
		}
	}

	// Only return a response if we have parts
	if len(parts) > 0 {
		response := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Role:  RoleAssistant,
						Parts: parts,
					},
				},
			},
		}
		return response, nil
	}

	// If no content, get the next chunk
	return s.Next()
}

// Helper functions for conversion between GenAI and Anthropic formats

var genAIRoles = []Role{
	RoleModel,
	RoleAssistant,
}

func asClaudeRole(role string) anthropic.MessageParamRole {
	if slices.Contains(genAIRoles, role) {
		return anthropic.MessageParamRoleAssistant
	}
	return anthropic.MessageParamRoleUser
}

func partToClaudeMessageBlock(part *genai.Part) (anthropic.ContentBlockParamUnion, error) {
	if part.Text != "" {
		params := anthropic.NewTextBlock(part.Text)
		params.OfRequestTextBlock.Type = constant.ValueOf[constant.Text]().Default()
		return params, nil
	}

	if part.FunctionCall != nil {
		funcCall := part.FunctionCall
		// Assert function call name if [genai.Part.FunctionCall] is non-nil
		if funcCall.Name != "" {
			return anthropic.ContentBlockParamUnion{}, errors.New("FunctionCall name is empty")
		}

		params := anthropic.ContentBlockParamOfRequestToolUseBlock(funcCall.ID, funcCall.Args, funcCall.Name)
		params.OfRequestToolUseBlock.Type = constant.ValueOf[constant.ToolUse]().Default()
		return anthropic.ContentBlockParamUnion{}, nil
	}

	if part.FunctionResponse != nil {
		funcResponse := part.FunctionResponse
		if result, ok := funcResponse.Response["result"]; ok {
			params := anthropic.NewToolResultBlock(funcResponse.ID, fmt.Sprintf("%s", result), false)
			params.OfRequestToolResultBlock.Type = constant.ValueOf[constant.ToolResult]().Default()
			return params, nil
		}
	}

	return anthropic.ContentBlockParamUnion{}, fmt.Errorf("not supported yet %T part type", part)
}

// contentToClaudeMessageParam converts [*genai.Content] to [anthropic.MessageParam].
func contentToClaudeMessageParam(content *genai.Content) (msgParam anthropic.MessageParam) {
	// Skip system messages (handled separately in Generate/StreamGenerate)
	if content.Role == RoleSystem {
		return
	}
	msgParam.Role = asClaudeRole(content.Role)

	msgParam.Content = make([]anthropic.ContentBlockParamUnion, 0, len(content.Parts))
	for _, part := range content.Parts {
		msgBlock, err := partToClaudeMessageBlock(part)
		if err != nil {
			continue
		}
		msgParam.Content = append(msgParam.Content, msgBlock)
	}

	return msgParam
}
