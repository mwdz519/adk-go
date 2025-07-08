// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

// CredentialRefresherRegistry registry for credential refresher instances.
//
// # Experimental
//
// This feature is experimental and may change or be removed in future versions without notice. It may
// introduce breaking changes at any time.
type CredentialRefresherRegistry struct {
	refreshers map[AuthCredentialTypes]CredentialRefresher
}

// NewCredentialRefresherRegistry returns the new [CredentialRefresherRegistry].
func NewCredentialRefresherRegistry() *CredentialRefresherRegistry {
	return &CredentialRefresherRegistry{
		refreshers: make(map[AuthCredentialTypes]CredentialRefresher),
	}
}

// Register register a refresher instance for a credential type.
func (r *CredentialRefresherRegistry) Register(credentialType AuthCredentialTypes, refresher CredentialRefresher) {
	r.refreshers[credentialType] = refresher
}

// GetRefresher get the refresher instance for a credential type.
func (r *CredentialRefresherRegistry) GetRefresher(credentialType AuthCredentialTypes) (CredentialRefresher, bool) {
	refresher, ok := r.refreshers[credentialType]
	return refresher, ok
}
