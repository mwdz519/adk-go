// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"errors"

	"google.golang.org/genai"
)

// CallbackContext provides the context of various callbacks within an agent run.
type CallbackContext struct {
	*ReadOnlyContext

	eventActions *EventActions

	state *State
}

func (cc *CallbackContext) WithEventActions(eventActions *EventActions) *CallbackContext {
	cc.eventActions = eventActions
	return cc
}

// NewCallbackContext creates a new [*CallbackContext] with the given args.
func NewCallbackContext(iccx *InvocationContext) *CallbackContext {
	cc := &CallbackContext{
		ReadOnlyContext: NewReadOnlyContext(iccx),
		// TODO(adk-python: weisun): make this public for Agent Development Kit, but private for users.
		eventActions: new(EventActions),
	}

	cc.state = NewState(iccx.Session.State(), cc.eventActions.StateDelta)

	return cc
}

// EventActions returns the event actions of the current session.
func (cc *CallbackContext) EventActions() *EventActions {
	return cc.eventActions
}

// State returns the delta-aware state of the current session.
//
// For any state change, you can mutate this object directly,
// e.g. `ctx.state['foo'] = 'bar'`
func (cc *CallbackContext) State() *State {
	return cc.state
}

// LoadArtifact loads an artifact attached to the current session.
func (cc *CallbackContext) LoadArtifact(ctx context.Context, filename string, version int) (*genai.Part, error) {
	artifactSvc := cc.InvocationContext.ArtifactService
	if artifactSvc == nil {
		return nil, errors.New("artifact service is not initialized")
	}

	return artifactSvc.LoadArtifact(ctx,
		cc.InvocationContext.AppName(),
		cc.InvocationContext.UserID(),
		cc.InvocationContext.Session.ID(),
		filename,
		version,
	)
}

// SaveArtifact saves an artifact and records it as delta for the current session.
func (cc *CallbackContext) SaveArtifact(ctx context.Context, filename string, artifact *genai.Part) (int, error) {
	artifactSvc := cc.InvocationContext.ArtifactService
	if artifactSvc == nil {
		return 0, errors.New("artifact service is not initialized")
	}

	version, err := artifactSvc.SaveArtifact(
		ctx,
		cc.InvocationContext.AppName(),
		cc.InvocationContext.UserID(),
		cc.InvocationContext.Session.ID(),
		filename,
		artifact,
	)
	if err != nil {
		return 0, err
	}

	cc.eventActions.ArtifactDelta[filename] = version
	return version, nil
}
