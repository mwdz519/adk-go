// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

// EventActions represents the actions attached to an event.
type EventActions struct {
	// SkipSummarization if true, it won't call model to summarize function response.
	//
	// Only used for functionResponse event.
	SkipSummarization bool

	// StateDelta indicates that the event is updating the state with the given delta.
	StateDelta map[string]any

	// ArtifactDelta indicates that the event is updating an artifact. key is the filename, value is the version.
	ArtifactDelta map[string]int

	// TransferToAgent if set, the event transfers to the specified agent.
	TransferToAgent string

	// Escalate is the agent is escalating to a higher level agent.
	Escalate bool

	// RequestedAuthConfigs authentication configurations requested by tool responses.
	//
	// This field will only be set by a tool response event indicating tool request
	// auth credential.
	//
	// Keys:
	// The function call id. Since one function response event could contain
	// multiple function responses that correspond to multiple function calls. Each
	// function call could request different auth configs. This id is used to
	// identify the function call.
	//
	// Values:
	// The requested auth config.
	RequestedAuthConfigs map[string]*AuthConfig
}

// WithSkipSummarization configures the skipSummarization to the [EventActions].
func (ea *EventActions) WithSkipSummarization(skipSummarization bool) *EventActions {
	ea.SkipSummarization = skipSummarization
	return ea
}

// WithStateDelta configures the stateDelta to the [EventActions].
func (ea *EventActions) WithStateDelta(stateDelta map[string]any) *EventActions {
	ea.StateDelta = stateDelta
	return ea
}

// WithArtifactDelta configures the artifactDelta to the [EventActions].
func (ea *EventActions) WithArtifactDelta(artifactDelta map[string]int) *EventActions {
	ea.ArtifactDelta = artifactDelta
	return ea
}

// WithTransferToAgent configures the transferToAgent to the [EventActions].
func (ea *EventActions) WithTransferToAgent(transferToAgent string) *EventActions {
	ea.TransferToAgent = transferToAgent
	return ea
}

// WithEscalate configures the escalate to the [EventActions].
func (ea *EventActions) WithEscalate(escalate bool) *EventActions {
	ea.Escalate = escalate
	return ea
}

// WithRequestedAuthConfigs configures the requestedAuthConfigs to the [EventActions].
func (ea *EventActions) WithRequestedAuthConfigs(requestedAuthConfigs map[string]*AuthConfig) *EventActions {
	ea.RequestedAuthConfigs = requestedAuthConfigs
	return ea
}

// NewEventActions creates a new [EventActions] instance with default values.
func NewEventActions() *EventActions {
	return &EventActions{
		StateDelta:           make(map[string]any),
		ArtifactDelta:        make(map[string]int),
		RequestedAuthConfigs: make(map[string]*AuthConfig),
	}
}
