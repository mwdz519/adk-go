// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
	"github.com/bytedance/sonic"
	"google.golang.org/genai"
)

const (
	// ClaudeDefaultModel is the default model name for [Claude].
	//
	// This model is only available on Google Cloud Platform (GCP) Vertex AI.
	// If you want to use the Anthropic official model, pass any model name that is defined in the
	// anthropic-sdk-go package's constants to [NewClaude].
	ClaudeDefaultModel = "claude-3-5-sonnet-v2@20241022"

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
		modelName = ClaudeDefaultModel
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

		// GCP Vertex AI
		"claude-3-7-sonnet@20250219",
		"claude-3-5-haiku@20241022",
		"claude-3-5-sonnet-v2@20241022",
		"claude-3-opus@20240229",
		"claude-3-sonnet@20240229",
		"claude-3-haiku@20240307",

		// AWS Bedrock
		"anthropic.claude-3-7-sonnet-20250219-v1:0",
		"anthropic.claude-3-5-haiku-20241022-v1:0",
		"anthropic.claude-3-5-sonnet-20241022-v2:0",
		"anthropic.claude-3-5-sonnet-20240620-v1:0",
		"anthropic.claude-3-opus-20240229-v1:0",
		"anthropic.claude-3-sonnet-20240229-v1:0",
		"anthropic.claude-3-haiku-20240307-v1:0",
	}
}

// Connect creates a live connection to the Claude LLM.
//
// TODO(zchee): implements.
func (m *Claude) Connect() (BaseLLMConnection, error) {
	// Ensure we can get an Anthropic client
	_ = m.anthropicClient

	// For now, this is a placeholder as we haven't implemented ClaudeConnection yet
	// In a real implementation, we would return a proper ClaudeConnection
	return nil, fmt.Errorf("ClaudeConnection not implemented yet")
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

func functionDeclarationToToolParam(funcDeclaration *genai.FunctionDeclaration) (toolUnion anthropic.ToolUnionParam, err error) {
	if funcDeclaration.Name == "" {
		return toolUnion, errors.New("functionDeclaration name is empty")
	}

	inputSchemaProps := make(map[string]*genai.Schema)
	if params := funcDeclaration.Parameters; params != nil && params.Properties != nil {
		maps.Insert(inputSchemaProps, maps.All(params.Properties))
	}
	inputSchema := anthropic.ToolInputSchemaParam{
		Type:       constant.ValueOf[constant.Object]().Default(),
		Properties: inputSchemaProps,
	}

	toolUnion = anthropic.ToolUnionParamOfTool(inputSchema, funcDeclaration.Name)
	toolUnion.OfTool.Description = param.NewOpt(funcDeclaration.Description)

	return toolUnion, nil
}

// Generate generates content from the model.
func (m *Claude) Generate(ctx context.Context, request *LLMRequest) (*GenerateResponse, error) {
	// Convert messages to Anthropic format
	messages := make([]anthropic.MessageParam, len(request.Contents))
	for i, content := range request.Contents {
		messages[i] = contentToClaudeMessageParam(content)
	}

	// Prepare parameters
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(m.model),
		Messages:  messages,
		MaxTokens: 4096,
	}

	// Apply generation config if provided
	if config := request.Config; config != nil {
		// MaxOutputTokens is an int32 directly, not a pointer
		if config.MaxOutputTokens > 0 {
			params.MaxTokens = int64(config.MaxOutputTokens)
		}

		if config.Temperature != nil {
			params.Temperature = anthropic.Float(float64(*config.Temperature))
		}

		if config.TopK != nil {
			params.TopK = anthropic.Int(int64(*config.TopK))
		}

		if config.TopP != nil {
			params.TopP = anthropic.Float(float64(*config.TopP))
		}

		// Add tools if provided
		var tools []anthropic.ToolUnionParam
		if len(request.Tools) > 0 && request.Tools[0].FunctionDeclarations != nil {
			tools = slices.Grow(tools, len(request.Tools[0].FunctionDeclarations))
			for _, funcDeclarations := range request.Tools[0].FunctionDeclarations {
				toolUnion, err := functionDeclarationToToolParam(funcDeclarations)
				if err != nil {
					return nil, err
				}
				tools = append(tools, toolUnion)
			}
		}
		params.Tools = tools
	}

	// Apply system prompt if it exists in first content
	systemText, hasSystem := extractSystemPrompt(request.Contents)
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
		if len(messages) > 1 {
			messages = messages[1:]
			params.Messages = messages
		}
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
	request := &LLMRequest{
		Contents: contents,
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
		request.Config = genConfig
		request.SafetySettings = config.SafetySettings
	}

	resp, err := m.Generate(ctx, request)
	if err != nil {
		return nil, err
	}

	return resp.Content, nil
}

// StreamGenerate streams generated content from the model.
func (m *Claude) StreamGenerate(ctx context.Context, request *LLMRequest) (StreamGenerateResponse, error) {
	// Convert to Anthropic format
	messages := make([]anthropic.MessageParam, len(request.Contents))
	for i, content := range request.Contents {
		messages[i] = contentToClaudeMessageParam(content)
	}

	// Prepare parameters
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(m.model),
		Messages:  messages,
		MaxTokens: 4096,
	}

	// Apply generation config if provided
	if config := request.Config; config != nil {
		// MaxOutputTokens is an int32 directly, not a pointer
		if config.MaxOutputTokens > 0 {
			params.MaxTokens = int64(config.MaxOutputTokens)
		}

		if config.Temperature != nil {
			params.Temperature = anthropic.Float(float64(*config.Temperature))
		}

		if config.TopK != nil {
			params.TopK = anthropic.Int(int64(*config.TopK))
		}

		if config.TopP != nil {
			params.TopP = anthropic.Float(float64(*config.TopP))
		}

		// Add tools if provided
		var tools []anthropic.ToolUnionParam
		if len(request.Tools) > 0 && request.Tools[0].FunctionDeclarations != nil {
			tools = slices.Grow(tools, len(request.Tools[0].FunctionDeclarations))
			for _, funcDeclarations := range request.Tools[0].FunctionDeclarations {
				toolUnion, err := functionDeclarationToToolParam(funcDeclarations)
				if err != nil {
					return nil, err
				}
				tools = append(tools, toolUnion)
			}
		}
		params.Tools = tools
	}

	// Apply system prompt if it exists in first content
	systemText, hasSystem := extractSystemPrompt(request.Contents)
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
		if len(messages) > 1 {
			messages = messages[1:]
			params.Messages = messages
		}
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
	request := &LLMRequest{
		Contents: contents,
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
		request.Config = genConfig
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

var claudeStopReasons = []anthropic.StopReason{
	anthropic.StopReasonEndTurn,
	anthropic.StopReasonStopSequence,
	anthropic.StopReasonToolUse,
}

func asClaudeToFinishReason(stopReason anthropic.StopReason) genai.FinishReason {
	if slices.Contains(claudeStopReasons, stopReason) {
		return genai.FinishReasonStop
	}

	if stopReason == anthropic.StopReasonMaxTokens {
		return genai.FinishReasonMaxTokens
	}

	return genai.FinishReasonUnspecified
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
		return params, nil
	}

	if part.FunctionResponse != nil {
		funcResp := part.FunctionResponse
		if result, ok := funcResp.Response["result"]; ok {
			params := anthropic.NewToolResultBlock(funcResp.ID, fmt.Sprintf("%s", result), false)
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

func claudeContentBlockToPart(contentBlock anthropic.ContentBlockUnion) (*genai.Part, error) {
	switch cBlock := contentBlock.AsAny().(type) {
	case anthropic.TextBlock:
		return genai.NewPartFromText(cBlock.Text), nil

	case anthropic.ToolUseBlock:
		if cBlock.Input == nil {
			return nil, fmt.Errorf("input field must be non-nil: %#v", cBlock)
		}
		var args map[string]any
		if err := sonic.ConfigFastest.Unmarshal(cBlock.Input, &args); err != nil {
			return nil, fmt.Errorf("unmarshal ToolUseBlock input: %w", err)
		}
		part := genai.NewPartFromFunctionCall(cBlock.Name, args)
		part.FunctionCall.ID = cBlock.ID
		return part, nil

	case anthropic.ThinkingBlock, anthropic.RedactedThinkingBlock:
		return nil, fmt.Errorf("not supported yet converts %T content block", cBlock)
	}

	return nil, fmt.Errorf("unreachable: no variant present")
}

func claudeMessageToGenerateContentResponse(message anthropic.Message) *LLMResponse {
	parts := make([]*genai.Part, 0, len(message.Content))
	for _, mcontent := range message.Content {
		part, err := claudeContentBlockToPart(mcontent)
		if err != nil {
			continue
		}
		parts = append(parts, part)
	}

	return &LLMResponse{
		Content: &genai.Content{
			Role:  RoleModel,
			Parts: parts,
		},
		// TODO: Deal with these later.
		// finish_reason=to_google_genai_finish_reason(message.stop_reason),
		// usage_metadata=types.GenerateContentResponseUsageMetadata(
		//     prompt_token_count=message.usage.input_tokens,
		//     candidates_token_count=message.usage.output_tokens,
		//     total_token_count=(
		//         message.usage.input_tokens + message.usage.output_tokens
		//     ),
		// ),
	}
}
