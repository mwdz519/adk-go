// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"errors"
)

// ToolContext represents a context of the tool.
//
// This type provides the context for a tool invocation, including access to
// the invocation context, function call ID, event actions, and authentication
// response. It also provides methods for requesting credentials, retrieving
// authentication responses, listing artifacts, and searching memory.
type ToolContext struct {
	*CallbackContext

	invocationContext *InvocationContext
	functionCallID    string
	eventActions      *EventActions
}

// WithFunctionCallID sets the function call ID for the [*ToolContext].
func (tc *ToolContext) WithFunctionCallID(funcCallID string) *ToolContext {
	tc.functionCallID = funcCallID
	return tc
}

// WithEventActions sets the [*EventActions] for the [*ToolContext].
func (tc *ToolContext) WithEventActions(eventActions *EventActions) *ToolContext {
	tc.eventActions = eventActions
	tc.CallbackContext.eventActions = eventActions
	return tc
}

// NewToolContext creates a new [ToolContext] with the given invocation context.
func NewToolContext(ictx *InvocationContext) *ToolContext {
	return &ToolContext{
		CallbackContext: &CallbackContext{
			ReadOnlyContext: NewReadOnlyContext(ictx),
		},
		invocationContext: ictx,
	}
}

// InvocationContext returns the invocation context for the tool context.
func (tc *ToolContext) InvocationContext() *InvocationContext {
	return tc.invocationContext
}

// FunctionCallID returns the function call ID for the tool context.
func (tc *ToolContext) FunctionCallID() string {
	return tc.functionCallID
}

// Actions returns the event actions for the tool context.
func (tc *ToolContext) Actions() *EventActions {
	return tc.eventActions
}

func (tc *ToolContext) RequestCredential(ts *AuthConfig) error {
	if tc.functionCallID == "" {
		return errors.New("functionCallID is not set")
	}

	tc.eventActions.RequestedAuthConfigs[tc.functionCallID] = ts

	return nil
}

// GetAuthResponse returns the authentication credential for the given authConfig.
func (tc *ToolContext) GetAuthResponse(authConfig *AuthConfig) *AuthCredential {
	return NewAuthHandler(authConfig).GetAuthResponse(tc.state)
}

// ListArtifacts lists the filenames of the artifacts attached to the current session.
func (tc *ToolContext) ListArtifacts(ctx context.Context) ([]string, error) {
	artifactSvc := tc.invocationContext.ArtifactService
	if artifactSvc == nil {
		return nil, errors.New("artifact service is not initialized")
	}

	return artifactSvc.ListArtifactKey(ctx, tc.InvocationContext().AppName(), tc.InvocationContext().UserID(), tc.InvocationContext().Session.ID())
}

// SearchMemory searches the memory of the current user.
func (tc *ToolContext) SearchMemory(ctx context.Context, query string) (*SearchMemoryResponse, error) {
	memorySvc := tc.invocationContext.MemoryService
	if memorySvc == nil {
		return nil, errors.New("memory service is not available")
	}

	return memorySvc.SearchMemory(ctx, tc.InvocationContext().AppName(), tc.InvocationContext().UserID(), query)
}
