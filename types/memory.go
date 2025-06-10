// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"time"

	"google.golang.org/genai"
)

// MemoryService defines the interface for memory services.
//
// A session may be added multiple times during its lifetime.
type MemoryService interface {
	// AddSessionToMemory adds the contents of a session to memory.
	AddSessionToMemory(ctx context.Context, session Session) error

	// SearchMemory searches for sessions that match the query.
	SearchMemory(ctx context.Context, appName, userID, query string) (*SearchMemoryResponse, error)
}

// MemoryEntry represents an one memory entry.
type MemoryEntry struct {
	// The main content of the memory.
	Content *genai.Content

	// The author of the memory.
	Author string

	// The timestamp when the original content of this memory happened.
	//
	// This string will be forwarded to LLM. Preferred format is ISO 8601 format.
	Timestamp time.Time
}

// SearchMemoryResponse represents the response from a memory search.
type SearchMemoryResponse struct {
	// Results are the memory items matching the search.
	Memories []*MemoryEntry `json:"memories"`
}
