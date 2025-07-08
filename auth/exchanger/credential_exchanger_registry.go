// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package exchanger

import (
	"github.com/go-a2a/adk-go/types"
)

// CredentialExchangerRegistry registry for credential exchanger instances.
//
// # Experimental
//
// This feature is experimental and may change or be removed in future versions without notice. It may
// introduce breaking changes at any time.
type CredentialExchangerRegistry struct {
	exchangers map[types.AuthCredentialTypes]types.CredentialExchanger
}

// NewCredentialExchangerRegistry returns the new [CredentialExchangerRegistry].
func NewCredentialExchangerRegistry() *CredentialExchangerRegistry {
	return &CredentialExchangerRegistry{
		exchangers: make(map[types.AuthCredentialTypes]types.CredentialExchanger),
	}
}

// Register registry for credential exchanger instances.
func (e *CredentialExchangerRegistry) Register(credentialType types.AuthCredentialTypes, exchanger types.CredentialExchanger) {
	e.exchangers[credentialType] = exchanger
}

// GetExchanger get the exchanger for a credential type.
func (e *CredentialExchangerRegistry) GetExchanger(credentialType types.AuthCredentialTypes) (types.CredentialExchanger, bool) {
	exchanger, ok := e.exchangers[credentialType]
	return exchanger, ok
}
