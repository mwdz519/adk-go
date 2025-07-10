// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

// AuthScheme represents either a [SecurityScheme] or [OpenIDConnectWithConfig].
type AuthScheme interface {
	AuthType() AuthCredentialTypes
}

// SecurityScheme represents a security scheme that implements the [AuthScheme] interface.
type SecurityScheme interface {
	AuthScheme

	isSecurityScheme()
}

// OpenIDConnectWithConfig represents an OpenID Connect configuration with additional endpoints and methods.
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

// AuthType returns a string representation of the [OpenIDConnectWithConfig] type.
func (a *OpenIDConnectWithConfig) AuthType() AuthCredentialTypes {
	return a.Type
}

// APIKeyIn represents the location of an API key in a request.
type APIKeyIn string

const (
	// InQuery indicates the API key is passed in the url query parameters.
	InQuery APIKeyIn = "query"
	// InHeader indicates the API key is passed in the request header.
	InHeader APIKeyIn = "header"
	// InCookie indicates the API key is passed in a cookie.
	InCookie APIKeyIn = "cookie"
)

// APIKeySecurityScheme represents an API key security scheme.
type APIKeySecurityScheme struct {
	Type AuthCredentialTypes `json:"type"`
	In   APIKeyIn            `json:"in,omitzero"`
	Name string              `json:"name,omitzero"`
}

var _ SecurityScheme = (*APIKeySecurityScheme)(nil)

func (*APIKeySecurityScheme) isSecurityScheme() {}

// AuthType returns a string representation of the [APIKeySecurityScheme] type.
func (a *APIKeySecurityScheme) AuthType() AuthCredentialTypes {
	return a.Type
}

// HTTPBaseSecurityScheme represents a base HTTP security scheme.
type HTTPBaseSecurityScheme struct {
	Type   AuthCredentialTypes `json:"type"`
	Scheme string              `json:"scheme,omitzero"`
}

var _ SecurityScheme = (*HTTPBaseSecurityScheme)(nil)

func (*HTTPBaseSecurityScheme) isSecurityScheme() {}

// AuthType returns a string representation of the [HTTPBaseSecurityScheme] type.
func (a *HTTPBaseSecurityScheme) AuthType() AuthCredentialTypes {
	return a.Type
}

// OAuth2SecurityScheme represents an OAuth2 security scheme with various flows.
type OAuth2SecurityScheme struct {
	Type  AuthCredentialTypes `json:"type"`
	Flows *OAuthFlows         `json:"flows"`
}

var _ SecurityScheme = (*OAuth2SecurityScheme)(nil)

func (*OAuth2SecurityScheme) isSecurityScheme() {}

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

// OpenIdConnectSecurityScheme represents an OpenID Connect security scheme.
type OpenIdConnectSecurityScheme struct {
	Type             AuthCredentialTypes `json:"type"`
	OpenIDConnectURL string              `json:"openIdConnectUrl"`
}

var _ SecurityScheme = (*OpenIdConnectSecurityScheme)(nil)

func (*OpenIdConnectSecurityScheme) isSecurityScheme() {}

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

// OAuthGrantType represents the OAuth2 flow or grant type.
type OAuthGrantType string

const (
	// ClientCredentialsGrant represents the client credentials grant type.
	// See RFC 6749 Section 4.4.
	ClientCredentialsGrant OAuthGrantType = "client_credentials"

	// AuthorizationCodeGrant represents the authorization code grant type.
	// See RFC 6749 Section 4.1.
	AuthorizationCodeGrant OAuthGrantType = "authorization_code"

	// ImplicitGrant represents the implicit grant type.
	// See RFC 6749 Section 4.2.
	ImplicitGrant OAuthGrantType = "implicit"

	// PasswordGrant represents the password grant type.
	// See RFC 6749 Section 4.3.
	PasswordGrant OAuthGrantType = "password"
)

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
