// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

// CredentialExchangerRegistry registry for credential exchanger instances.
//
// # Experimental
//
// This feature is experimental and may change or be removed in future versions without notice. It may
// introduce breaking changes at any time.
type CredentialExchangerRegistry struct {
	exchangers map[AuthCredentialTypes]CredentialExchanger
}

// NewCredentialExchangerRegistry returns the new [CredentialExchangerRegistry].
func NewCredentialExchangerRegistry() *CredentialExchangerRegistry {
	return &CredentialExchangerRegistry{
		exchangers: make(map[AuthCredentialTypes]CredentialExchanger),
	}
}

// Register registry for credential exchanger instances.
func (e *CredentialExchangerRegistry) Register(credentialType AuthCredentialTypes, exchanger CredentialExchanger) {
	e.exchangers[credentialType] = exchanger
}

// GetExchanger get the exchanger for a credential type.
func (e *CredentialExchangerRegistry) GetExchanger(credentialType AuthCredentialTypes) (CredentialExchanger, bool) {
	exchanger, ok := e.exchangers[credentialType]
	return exchanger, ok
}
