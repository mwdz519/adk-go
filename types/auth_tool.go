// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
)

// AuthConfig response an auth config sent by tool asking client to collect auth credentials and
// adk and client will help to fill in the response.
type AuthConfig struct {
	// The auth scheme used to collect credentials
	AuthScheme AuthScheme

	// The raw auth credential used to collect credentials. The raw auth
	// credentials are used in some auth scheme that needs to exchange auth
	// credentials. e.g. OAuth2 and OIDC. For other auth scheme, it could be None.
	RawAuthCredential *AuthCredential

	// The exchanged auth credential used to collect credentials. adk and client
	// will work together to fill it. For those auth scheme that doesn't need to
	// exchange auth credentials, e.g. API key, service account etc. It's filled by
	// client directly. For those auth scheme that need to exchange auth credentials,
	// e.g. OAuth2 and OIDC, it's first filled by adk. If the raw credentials
	// passed by tool only has client id and client credential, adk will help to
	// generate the corresponding authorization uri and state and store the processed
	// credential in this field. If the raw credentials passed by tool already has
	// authorization uri, state, etc. then it's copied to this field. Client will use
	// this field to guide the user through the OAuth2 flow and fill auth response in
	// this field.
	ExchangedAuthCredential *AuthCredential

	// A user specified key used to load and save this credential in a credential
	// service.
	credentialKey string
}

// CredentialKey builds a hash key based on auth_scheme and raw_auth_credential used to
// save / load this credential to / from a credentials service.
func (ac *AuthConfig) CredentialKey() string {
	return NewAuthHandler(ac).GetCredentialKey()
}

// AuthToolArguments response an arguments for the special long running function tool that is used to
// request end user credentials.
type AuthToolArguments struct {
	// FunctionCallID is the ID of the function call requesting authentication.
	FunctionCallID string `json:"function_call_id"`

	// AuthConfig is the authentication configuration requested.
	AuthConfig *AuthConfig `json:"auth_config"`
}

// TODO(zchee): implements correctly
func ConvertToAuthConfig(data map[string]any, config *AuthConfig) (*AuthConfig, error) {
	// Process auth scheme
	schemeData, ok := data["auth_scheme"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("auth_scheme not found or invalid")
	}

	var scheme AuthScheme
	schemeTypeStr, ok := schemeData["type"].(string)
	if !ok {
		return nil, fmt.Errorf("auth scheme type not found")
	}

	schemeType := AuthCredentialTypes(schemeTypeStr)

	// Process based on scheme type
	switch schemeType {
	case OpenIDConnectCredentialTypes:
		oidcScheme := &OpenIDConnectWithConfig{
			Type: schemeType,
		}
		if endpoint, ok := schemeData["authorization_endpoint"].(string); ok {
			oidcScheme.AuthorizationEndpoint = endpoint
		}
		if endpoint, ok := schemeData["token_endpoint"].(string); ok {
			oidcScheme.TokenEndpoint = endpoint
		}
		scheme = oidcScheme
	case APIKeyCredentialTypes:
		securityScheme := &APIKeySecurityScheme{
			Type: schemeType,
		}
		scheme = securityScheme
	case HTTPCredentialTypes:
		securityScheme := &HTTPBaseSecurityScheme{
			Type: schemeType,
		}
		scheme = securityScheme
	case OAuth2CredentialTypes:
		securityScheme := &OAuth2SecurityScheme{
			Type: schemeType,
		}
		scheme = securityScheme
	}

	// Process credentials
	var rawCred, exchangedCred *AuthCredential

	if rawCredData, ok := data["raw_auth_credential"].(map[string]any); ok {
		rawCred = &AuthCredential{}
		if authTypeStr, ok := rawCredData["auth_type"].(string); ok {
			rawCred.AuthType = AuthCredentialTypes(authTypeStr)
		}
		// Process other credential fields as needed
	}

	if exchangedCredData, ok := data["exchanged_auth_credential"].(map[string]any); ok {
		exchangedCred = &AuthCredential{}
		if authTypeStr, ok := exchangedCredData["auth_type"].(string); ok {
			exchangedCred.AuthType = AuthCredentialTypes(authTypeStr)
		}
		// Process other credential fields as needed
	}

	// Set the config fields
	config.AuthScheme = scheme
	config.RawAuthCredential = rawCred
	config.ExchangedAuthCredential = exchangedCred

	return config, nil
}
