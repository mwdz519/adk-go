// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
)

// CredentialService represents an abstract class for Service that loads / saves tool credentials from / to
// the backend credential store.
//
// # Experimental
//
// This feature is experimental and may change or be removed in future versions without notice. It may
// introduce breaking changes at any time.
type CredentialService interface {
	// LoadCredential loads the credential by auth config and current tool context from the
	// backend credential store.
	LoadCredential(ctx context.Context, authConfig *AuthConfig, toolCtx *ToolContext) (*AuthCredential, error)

	// SaveCredential saves the exchanged_auth_credential in auth config to the backend credential
	// store.
	SaveCredential(ctx context.Context, authConfig *AuthConfig, toolCtx *ToolContext) error
}
