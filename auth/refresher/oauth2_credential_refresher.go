// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package refresher

import (
	"context"
	"fmt"
	"time"

	"github.com/go-a2a/adk-go/types"
	"golang.org/x/oauth2"
)

// OAuth2CredentialRefresher Refreshes OAuth2 credentials including Google OAuth2 JSON credentials.
type OAuth2CredentialRefresher struct{}

var _ types.CredentialRefresher = (*OAuth2CredentialRefresher)(nil)

// IsRefreshNeeded implements [types.CredentialRefresher].
func (r *OAuth2CredentialRefresher) IsRefreshNeeded(ctx context.Context, authCredential *types.AuthCredential, authScheme types.AuthScheme) bool {
	if authCredential.OAuth2 != nil {
		tok := &oauth2.Token{
			AccessToken: authCredential.OAuth2.AccessToken,
			Expiry:      authCredential.OAuth2.ExpiresAt,
			ExpiresIn:   authCredential.OAuth2.ExpiresIn,
		}
		return !tok.Valid()
	}
	return false
}

// Refresh implements [types.CredentialRefresher].
func (r *OAuth2CredentialRefresher) Refresh(ctx context.Context, authCredential *types.AuthCredential, authScheme types.AuthScheme) (*types.AuthCredential, error) {
	if authCredential.OAuth2 != nil && authScheme != nil {
		if r.IsRefreshNeeded(ctx, authCredential, authScheme) {
			client := types.CreateOAuth2Session(ctx, authScheme, authCredential)
			if client == nil {
				return authCredential, nil
			}

			// Create a token with the refresh token
			currentToken := &oauth2.Token{
				RefreshToken: authCredential.OAuth2.RefreshToken,
				// Set expiry to past time to force refresh
				Expiry: time.Now().Add(-time.Hour),
			}

			// Create token source and get fresh token
			tokenSource := client.TokenSource(ctx, currentToken)
			newToken, err := tokenSource.Token()
			if err != nil {
				return nil, fmt.Errorf("refresh token: %w", err)
			}

			newAuthCredential := types.UpdateCredentialWithTokens(authCredential, newToken)
			return newAuthCredential, nil
		}
	}

	return authCredential, nil
}
