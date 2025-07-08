// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
)

// CredentialRefresherError is the credential refresh errors.
type CredentialRefresherError string

// Error returns a string representation of the [CredentialRefresherError].
func (e CredentialRefresherError) Error() string {
	return string(e)
}

// CredentialRefresher represents an interface for credential refreshers.
//
// Credential refreshers are responsible for checking if a credential is expired
// or needs to be refreshed, and for refreshing it if necessary.
type CredentialRefresher interface {
	// IsRefreshNeeded checks if a credential needs to be refreshed.
	IsRefreshNeeded(ctx context.Context, authCredential *AuthCredential, authScheme AuthScheme) bool

	// Refresh refreshes a credential if needed.
	Refresh(ctx context.Context, authCredential *AuthCredential, authScheme AuthScheme) (*AuthCredential, error)
}
