// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"google.golang.org/genai"
)

// TranscriptionEntry represents a store the data that can be used for transcription.
type TranscriptionEntry struct {
	// The role that created this data, typically "user" or "model".
	Role string

	// The data that can be used for transcription.
	Data any
}

// NewTranscriptionEntry creates a new [TranscriptionEntry].
func NewTranscriptionEntry[T *genai.Blob | *genai.Content](role string, data T) *TranscriptionEntry {
	return &TranscriptionEntry{
		Role: role,
		Data: data,
	}
}
