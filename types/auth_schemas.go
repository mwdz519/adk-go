// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

// OAuthGrantType represents the OAuth2 flow (or grant type).
type OAuthGrantType string

const (
	ClientCredentialsGrant OAuthGrantType = "client_credentials"
	AuthorizationCodeGrant OAuthGrantType = "authorization_code"
	ImplicitGrant          OAuthGrantType = "implicit"
	PasswordGrant          OAuthGrantType = "password"
)

// AuthScheme represents either a [SecurityScheme] or [OpenIDConnectWithConfig].
type AuthScheme interface {
	isAuthScheme()
	AuthType() AuthCredentialTypes
}

type OpenIDConnectWithConfig struct {
	Type                              AuthCredentialTypes `json:"type"`
	AuthorizationEndpoint             string              `json:"authorization_endpoint"`
	TokenEndpoint                     string              `json:"token_endpoint"`
	UserinfoEndpoint                  string              `json:"userinfo_endpoint,omitzero"`
	RevocationEndpoint                string              `json:"revocation_endpoint,omitzero"`
	TokenEndpointAuthMethodsSupported string              `json:"token_endpoint_auth_methods_supported,omitzero"`
	GrantTypesSupported               []string            `json:"grant_types_supported,omitzero"`
	Scopes                            []string            `json:"scopes,omitzero"`
}

var _ AuthScheme = (*OpenIDConnectWithConfig)(nil)

func (*OpenIDConnectWithConfig) isAuthScheme() {}

// AuthType returns a string representation of the [OpenIDConnectWithConfig] type.
func (a *OpenIDConnectWithConfig) AuthType() AuthCredentialTypes {
	return a.Type
}

type SecurityScheme interface {
	isAuthScheme()
	isSecurityScheme()
}

type APIKeySecurityScheme struct {
	Type AuthCredentialTypes `json:"type"`
	In   string              `json:"in,omitzero"`
	Name string              `json:"name,omitzero"`
}

var (
	_ SecurityScheme = (*APIKeySecurityScheme)(nil)
	_ AuthScheme     = (*APIKeySecurityScheme)(nil)
)

func (*APIKeySecurityScheme) isSecurityScheme() {}
func (*APIKeySecurityScheme) isAuthScheme()     {}

// AuthType returns a string representation of the [APIKeySecurityScheme] type.
func (a *APIKeySecurityScheme) AuthType() AuthCredentialTypes {
	return a.Type
}

type HTTPBaseSecurityScheme struct {
	Type   AuthCredentialTypes `json:"type"`
	Scheme string              `json:"scheme,omitzero"`
}

var (
	_ SecurityScheme = (*HTTPBaseSecurityScheme)(nil)
	_ AuthScheme     = (*HTTPBaseSecurityScheme)(nil)
)

func (*HTTPBaseSecurityScheme) isSecurityScheme() {}
func (*HTTPBaseSecurityScheme) isAuthScheme()     {}

// AuthType returns a string representation of the [HTTPBaseSecurityScheme] type.
func (a *HTTPBaseSecurityScheme) AuthType() AuthCredentialTypes {
	return a.Type
}

type OAuth2SecurityScheme struct {
	Type  AuthCredentialTypes `json:"type"`
	Flows *OAuthFlows         `json:"flows"`
}

var (
	_ SecurityScheme = (*OAuth2SecurityScheme)(nil)
	_ AuthScheme     = (*OAuth2SecurityScheme)(nil)
)

func (*OAuth2SecurityScheme) isSecurityScheme() {}
func (*OAuth2SecurityScheme) isAuthScheme()     {}

// AuthType returns a string representation of the [OAuth2SecurityScheme] type.
func (a *OAuth2SecurityScheme) AuthType() AuthCredentialTypes {
	return a.Type
}

// OAuthFlow represents an OAuth2 flow configuration.
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitzero"`
	TokenURL         string            `json:"tokenUrl,omitzero"`
	RefreshURL       string            `json:"refreshUrl,omitzero"`
	Scopes           map[string]string `json:"scopes,omitzero"`
}

// OAuthFlows represents an OAuth2 flow configurations.
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitzero"`
	Password          *OAuthFlow `json:"password,omitzero"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitzero"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitzero"`
}

type OpenIdConnectSecurityScheme struct {
	Type             AuthCredentialTypes `json:"type"`
	OpenIDConnectURL string              `json:"openIdConnectUrl"`
}

var (
	_ SecurityScheme = (*OpenIdConnectSecurityScheme)(nil)
	_ AuthScheme     = (*OpenIdConnectSecurityScheme)(nil)
)

func (*OpenIdConnectSecurityScheme) isSecurityScheme() {}
func (*OpenIdConnectSecurityScheme) isAuthScheme()     {}

// AuthType returns a string representation of the [OpenIdConnectSecurityScheme] type.
func (a *OpenIdConnectSecurityScheme) AuthType() AuthCredentialTypes {
	return a.Type
}

// GetAuthSchemeType returns the scheme type from an AuthScheme type.
func GetAuthSchemeType(scheme AuthScheme) AuthCredentialTypes {
	switch s := scheme.(type) {
	case *APIKeySecurityScheme:
		return s.Type
	case *HTTPBaseSecurityScheme:
		return s.Type
	case *OAuth2SecurityScheme:
		return s.Type
	case *OpenIdConnectSecurityScheme:
		return s.Type
	case *OpenIDConnectWithConfig:
		return s.Type
	default:
		return ""
	}
}

// FromOAuthFlows determines the grant type from OAuthFlows.
func FromOAuthFlows(flows *OAuthFlows) OAuthGrantType {
	if flows == nil {
		return ""
	}

	switch {
	case flows.ClientCredentials != nil:
		return ClientCredentialsGrant
	case flows.AuthorizationCode != nil:
		return AuthorizationCodeGrant
	case flows.Implicit != nil:
		return ImplicitGrant
	case flows.Password != nil:
		return PasswordGrant
	default:
		return ""
	}
}
