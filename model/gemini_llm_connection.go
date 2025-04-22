// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"fmt"
	"iter"
	"sync"
	"time"

	"google.golang.org/genai"
)

// GeminiLLMConnection implements BaseLLMConnection interface for Gemini models.
type GeminiLLMConnection struct {
	model       *GoogleLLM
	responsesCh chan *LLMResponse
	client      *genai.Client
	stopped     bool
	mutex       sync.Mutex
}

var _ BaseLLMConnection = (*GeminiLLMConnection)(nil)

// NewGeminiLLMConnection creates a new GeminiLLMConnection.
func NewGeminiLLMConnection(model *GoogleLLM) (*GeminiLLMConnection, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	conn := &GeminiLLMConnection{
		model:       model,
		responsesCh: make(chan *LLMResponse, 100), // Buffer size to prevent blocking
		client:      model.client,
		stopped:     false,
	}

	return conn, nil
}

// Helper function to get generation config from the model
func getGenerationConfig(model *GoogleLLM) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{}

	if model.generationConfig != nil {
		config = &genai.GenerateContentConfig{
			Temperature:     model.generationConfig.Temperature,
			MaxOutputTokens: model.generationConfig.MaxOutputTokens,
			TopK:            model.generationConfig.TopK,
			TopP:            model.generationConfig.TopP,
		}
	}

	if len(model.safetySettings) > 0 {
		config.SafetySettings = model.safetySettings
	}

	return config
}

// SendHistory sends the conversation history to the model.
// The model will respond if the last content is from user, otherwise it will
// wait for new user input before responding.
func (c *GeminiLLMConnection) SendHistory(ctx context.Context, history []*genai.Content) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.stopped {
		return fmt.Errorf("connection is closed")
	}

	// Determine if the last message is from the user to decide if we should generate a response
	shouldRespond := false
	if len(history) > 0 {
		lastMsg := history[len(history)-1]
		if lastMsg.Role == "user" {
			shouldRespond = true
		}
	}

	// If the last message is from the user, generate a response
	if shouldRespond {
		go func() {
			// Start a streaming session for responses
			config := getGenerationConfig(c.model)
			stream := c.client.Models.GenerateContentStream(ctx, c.model.modelName, history, config)

			// Process the stream responses
			c.processStream(ctx, stream)
		}()
	}

	return nil
}

// SendContent sends a user content to the model.
// The model will respond immediately upon receiving the content.
func (c *GeminiLLMConnection) SendContent(ctx context.Context, content *genai.Content) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.stopped {
		return fmt.Errorf("connection is closed")
	}

	// Start a streaming session for responses
	go func() {
		// Send a single content message to the model
		config := getGenerationConfig(c.model)
		stream := c.client.Models.GenerateContentStream(ctx, c.model.modelName, []*genai.Content{content}, config)

		// Process the stream responses
		c.processStream(ctx, stream)
	}()

	return nil
}

// SendRealtime sends a chunk of audio or a frame of video to the model in realtime.
// The model may not respond immediately upon receiving the blob.
func (c *GeminiLLMConnection) SendRealtime(ctx context.Context, blob []byte, mimeType string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.stopped {
		return fmt.Errorf("connection is closed")
	}

	// Create a part with media blob data
	part := &genai.Part{
		InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     blob,
		},
	}

	// Create content with the blob
	content := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{part},
	}

	// Start a streaming session for responses
	go func() {
		config := getGenerationConfig(c.model)
		stream := c.client.Models.GenerateContentStream(ctx, c.model.modelName, []*genai.Content{content}, config)
		c.processStream(ctx, stream)
	}()

	return nil
}

// Receive returns a channel that yields model responses.
// It should be called after SendHistory, SendContent, or SendRealtime.
func (c *GeminiLLMConnection) Receive(ctx context.Context) (<-chan *LLMResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.stopped {
		return nil, fmt.Errorf("connection is closed")
	}

	// Return the channel that will receive responses
	return c.responsesCh, nil
}

// Close terminates the connection to the model.
// The connection object should not be used after this call.
func (c *GeminiLLMConnection) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.stopped {
		return nil // Already closed
	}

	c.stopped = true

	// Close the response channel after ensuring all in-flight writes are complete
	// Use a deferred close to allow goroutines to finish writing
	go func() {
		// Give pending goroutines time to terminate
		time.Sleep(100 * time.Millisecond)
		close(c.responsesCh)
	}()

	return nil
}

// processStream handles the processing of response streams and sends results to the responsesCh.
func (c *GeminiLLMConnection) processStream(ctx context.Context, stream iter.Seq2[*genai.GenerateContentResponse, error]) {
	var fullResponseText string
	isFirstChunk := true

	// For each response in the stream
	for response, err := range stream {
		// Check if we've stopped
		c.mutex.Lock()
		if c.stopped {
			c.mutex.Unlock()
			return
		}
		c.mutex.Unlock()

		// Check for context cancellation
		select {
		case <-ctx.Done():
			// Send an interruption response
			c.responsesCh <- &LLMResponse{
				Interrupted:  true,
				ErrorCode:    "CANCELLED",
				ErrorMessage: "Request was cancelled",
			}
			return
		default:
			// Continue processing
		}

		// Check for errors
		if err != nil {
			c.responsesCh <- &LLMResponse{
				ErrorCode:    "STREAM_ERROR",
				ErrorMessage: err.Error(),
			}
			return
		}

		// Create LLMResponse from the stream response
		llmResponse := Create(response)

		// For text responses, track partial/complete status
		if len(response.Candidates) > 0 && response.Candidates[0].Content != nil {
			// Handle partial flag for streaming text responses
			currentText := llmResponse.GetText()

			if isFirstChunk {
				fullResponseText = currentText
				isFirstChunk = false
				// First chunk is partial unless it's empty
				llmResponse.Partial = (currentText != "")
			} else {
				// Subsequent chunks are partial
				fullResponseText += currentText
				llmResponse.Partial = true
			}

			// Check if this is the last chunk (finish reason is present)
			if response.Candidates[0].FinishReason != genai.FinishReasonUnspecified {
				llmResponse.Partial = false
				llmResponse.TurnComplete = true
			}
		}

		// Send the response
		select {
		case c.responsesCh <- llmResponse:
			// Response sent successfully
		case <-ctx.Done():
			// Context cancelled while sending
			return
		}
	}

	// Send a final turn complete message
	c.responsesCh <- &LLMResponse{
		TurnComplete: true,
	}
}
