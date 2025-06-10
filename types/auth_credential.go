// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

type HTTPCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// HTTPAuth represents a credentials and metadata for HTTP authentication.
//
// The name of the HTTP Authorization scheme to be used in the Authorization
// header as defined in RFC7235. The values used SHOULD be registered in the
// IANA Authentication Scheme registry.
// Examples: 'basic', 'bearer'
type HTTPAuth struct {
	Scheme      string          `json:"scheme"`
	Credentials HTTPCredentials `json:"Credentials"`
}

// OAuth2Auth represents credential value and its metadata for a OAuth2 credential.
type OAuth2Auth struct {
	ClientID     string `json:"client_id,omitzero"`
	ClientSecret string `json:"client_secret,omitzero"`
	// tool or adk can generate the auth_uri with the state info thus client
	// can verify the state
	AuthURI string `json:"auth_uri,omitzero"`
	State   string `json:"state,omitzero"`
	// tool or adk can decide the redirect_uri if they don't want client to decide
	RedirectURI     string `json:"redirect_uri,omitzero"`
	AuthResponseURI string `json:"auth_response_uri,omitzero"`
	AuthCode        string `json:"auth_code,omitzero"`
	AccessToken     string `json:"access_token,omitzero"`
	RefreshToken    string `json:"refresh_token,omitzero"`
}

// ServiceAccountCredential represents Google Service Account configuration.
type ServiceAccountCredential struct {
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

// ServiceAccount represents Google Service Account configuration.
type ServiceAccount struct {
	ServiceAccountCredential ServiceAccountCredential `json:"service_account_credential,omitzero"`
	Scopes                   []string                 `json:"scopes"`
	UseDefaultCredential     bool                     `json:"use_default_credential,omitzero"`
}

// AuthCredentialTypes represents the type of authentication credential.
type AuthCredentialTypes string

const (
	// # API Key credential:
	// # https://swagger.io/docs/specification/v3_0/authentication/api-keys/
	APIKeyCredentialTypes AuthCredentialTypes = "apiKey"

	// Credentials for HTTP Auth schemes:
	// https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml
	HTTPCredentialTypes AuthCredentialTypes = "http"

	// OAuth2 credentials:
	// https://swagger.io/docs/specification/v3_0/authentication/oauth2/
	OAuth2CredentialTypes AuthCredentialTypes = "oauth2"

	// OpenID Connect credentials:
	// https://swagger.io/docs/specification/v3_0/authentication/openid-connect-discovery/
	OpenIDConnectCredentialTypes AuthCredentialTypes = "openIdConnect"

	// Service Account credentials:
	// https://cloud.google.com/iam/docs/service-account-creds
	ServiceAccountCredentialTypes AuthCredentialTypes = "serviceAccount"
)

// AuthCredential represents a data class representing an authentication credential.
//
// To exchange for the actual credential, please use CredentialExchanger.exchange_credential().
type AuthCredential struct {
	AuthType AuthCredentialTypes `json:"auth_type,omitzero"`

	// Resource reference for the credential.
	// This will be supported in the future.
	ResourceRef string `json:"resource_ref,omitzero"`

	APIKey         string          `json:"api_key,omitzero"`
	HTTP           *HTTPAuth       `json:"http,omitzero"`
	ServiceAccount *ServiceAccount `json:"service_account,omitzero"`
	OAuth2         *OAuth2Auth     `json:"oauth2,omitzero"`
}
