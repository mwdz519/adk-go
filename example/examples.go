// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package example

import (
	"google.golang.org/genai"
)

// Example represents a few-shot example.
type Example struct {
	Input  *genai.Content
	Output []*genai.Content
}
