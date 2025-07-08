// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
)

// CredentialExchangError is the credential exchange errors.
type CredentialExchangError string

// Error returns a string representation of the [CredentialExchangError].
func (e CredentialExchangError) Error() string {
	return string(e)
}

// CredentialExchanger represents an interface for credential exchangers.
//
// Credential exchangers are responsible for exchanging credentials from
// one format or scheme to another.
type CredentialExchanger interface {
	// Exchange exchange credential if needed.
	Exchange(ctx context.Context, authCredential *AuthCredential, authScheme AuthScheme) (*AuthCredential, error)
}
