// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"maps"
	"slices"

	"golang.org/x/oauth2"
)

// OAuth2Session represents a new OAuth 2 client requests session.
type OAuth2Session struct {
	*oauth2.Config
	State string
}

// CreateOAuth2Session create an OAuth2 session for token operations.
func CreateOAuth2Session(ctx context.Context, authScheme AuthScheme, authCredential *AuthCredential) *OAuth2Session {
	var (
		tokenEndpoint string
		scopes        []string
	)

	switch authSchema := authScheme.(type) {
	case *OpenIDConnectWithConfig:
		if authSchema.TokenEndpoint == "" {
			return nil
		}
		tokenEndpoint = authSchema.TokenEndpoint
		scopes = authSchema.Scopes

	case *OAuth2SecurityScheme:
		if authSchema.Flows.AuthorizationCode == nil || authSchema.Flows.AuthorizationCode.TokenURL == "" {
			return nil
		}
		tokenEndpoint = authSchema.Flows.AuthorizationCode.TokenURL
		scopes = slices.Sorted(maps.Keys(authSchema.Flows.AuthorizationCode.Scopes))

	default:
		return nil
	}

	if authCredential == nil || authCredential.OAuth2 == nil || authCredential.OAuth2.ClientID == "" || authCredential.OAuth2.ClientSecret == "" {
		return nil
	}

	return &OAuth2Session{
		Config: &oauth2.Config{
			ClientID:     authCredential.OAuth2.ClientID,
			ClientSecret: authCredential.OAuth2.ClientSecret,
			Endpoint: oauth2.Endpoint{
				TokenURL: tokenEndpoint,
			},
			Scopes:      scopes,
			RedirectURL: authCredential.OAuth2.RedirectURI,
		},
		State: authCredential.OAuth2.State,
	}
}
