// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"google.golang.org/genai"
)

// DefaultMaxLLMCalls is the default limit on the total number of llm calls.
const DefaultMaxLLMCalls = 500

// StreamingMode is the streaming mode.
type StreamingMode int

const (
	StreamingModeNone StreamingMode = iota
	StreamingModeSSE
	StreamingModeBidi
)

// String returns a string representation of the StreamingMode.
func (mode StreamingMode) String() string {
	switch mode {
	case StreamingModeNone:
		return "None"
	case StreamingModeSSE:
		return "sse"
	case StreamingModeBidi:
		return "bidi"
	}
	return ""
}

// RunConfig represents a configs for runtime behavior of agents.
type RunConfig struct {
	// Speech configuration for the live agent.
	SpeechConfig *genai.SpeechConfig

	// The output modalities. If not set, it's default to AUDIO.
	ResponseModalities []genai.Modality

	// Whether or not to save the input blobs as artifacts.
	SaveInputBlobsAsArtifacts bool

	// Whether to support CFC (Compositional Function Calling). Only applicable for
	// StreamingMode.SSE. If it's true. the LIVE API will be invoked. Since only LIVE
	// API supports CFC
	SupportCFC bool

	// Streaming mode.
	StreamingMode StreamingMode

	// Output transcription for live agents with audio response.
	OutputAudioTranscription *genai.AudioTranscriptionConfig

	// Input transcription for live agents with audio input from user.
	InputAudioTranscription *genai.AudioTranscriptionConfig

	// A limit on the total number of llm calls for a given run.
	MaxLLMCalls int
}
