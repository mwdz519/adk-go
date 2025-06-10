// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"google.golang.org/genai"
)

// ReadOnlyContext provides read-only access to agent context.
type ReadOnlyContext struct {
	InvocationContext *InvocationContext
}

// NewReadOnlyContext creates a new read-only context.
func NewReadOnlyContext(ictx *InvocationContext) *ReadOnlyContext {
	return &ReadOnlyContext{
		InvocationContext: ictx,
	}
}

// UserContent returns the user content that started this invocation. READONLY field.
func (rc *ReadOnlyContext) UserContent() *genai.Content {
	return rc.InvocationContext.UserContent
}

// InvocationContextID returns the current invocation id.
func (rc *ReadOnlyContext) InvocationContextID() string {
	return rc.InvocationContext.InvocationID
}

// AgentName returns the name of the agent that is currently running.
func (rc *ReadOnlyContext) AgentName() string {
	return rc.InvocationContext.Agent.Name()
}

// State returns the state of the current session. READONLY field.
func (rc *ReadOnlyContext) State() map[string]any {
	return rc.InvocationContext.Session.State()
}
