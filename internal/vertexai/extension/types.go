// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"strings"
	"time"

	"cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb"
)

// Extension represents a Vertex AI extension with its configuration and metadata.
//
// Extensions enable models to connect to external APIs for real-time data processing
// and performing real-world actions. Each extension is defined by its manifest which
// includes API specifications, authentication configuration, and runtime settings.
//
// This type wraps the protobuf Extension type and adds additional fields for state
// management and error handling that are specific to this ADK implementation.
type Extension struct {
	// Embed the protobuf Extension type to get all the standard fields
	*aiplatformpb.Extension

	// State indicates the current state of the extension (ADK-specific field).
	State ExtensionState `json:"state"`

	// Error contains error information if the extension is in an error state (ADK-specific field).
	Error *ExtensionError `json:"error,omitempty"`
}

// GetID extracts the extension ID from the resource name.
// Resource name format: projects/{project}/locations/{location}/extensions/{extension_id}
func (e *Extension) GetID() string {
	if e.Extension == nil || e.Extension.Name == "" {
		return ""
	}
	parts := strings.Split(e.Extension.Name, "/")
	if len(parts) >= 6 && parts[4] == "extensions" {
		return parts[5]
	}
	return ""
}

// GetCreateTimeAsTime returns the create time as a time.Time.
func (e *Extension) GetCreateTimeAsTime() time.Time {
	if e.Extension == nil || e.Extension.CreateTime == nil {
		return time.Time{}
	}
	return e.Extension.CreateTime.AsTime()
}

// GetUpdateTimeAsTime returns the update time as a time.Time.
func (e *Extension) GetUpdateTimeAsTime() time.Time {
	if e.Extension == nil || e.Extension.UpdateTime == nil {
		return time.Time{}
	}
	return e.Extension.UpdateTime.AsTime()
}

// ExtensionState represents the current state of an extension.
type ExtensionState string

const (
	// ExtensionStateUnspecified indicates an unspecified state.
	ExtensionStateUnspecified ExtensionState = "EXTENSION_STATE_UNSPECIFIED"

	// ExtensionStateActive indicates the extension is active and ready for use.
	ExtensionStateActive ExtensionState = "ACTIVE"

	// ExtensionStateCreating indicates the extension is being created.
	ExtensionStateCreating ExtensionState = "CREATING"

	// ExtensionStateDeleting indicates the extension is being deleted.
	ExtensionStateDeleting ExtensionState = "DELETING"

	// ExtensionStateError indicates the extension is in an error state.
	ExtensionStateError ExtensionState = "ERROR"
)

// ExtensionError contains error information for extensions in error state.
type ExtensionError struct {
	// Code is the error code.
	Code int32 `json:"code"`

	// Message is the error message.
	Message string `json:"message"`

	// Details contains additional error details.
	Details []any `json:"details,omitempty"`
}

// Type aliases for protobuf types to provide a cleaner API
type (
	// ExtensionManifest represents the manifest spec of an extension.
	ExtensionManifest = aiplatformpb.ExtensionManifest

	// AuthConfig specifies authentication configuration for extensions.
	AuthConfig = aiplatformpb.AuthConfig

	// AuthType represents the authentication type for extensions.
	AuthType = aiplatformpb.AuthType

	// RuntimeConfig contains runtime-specific configuration for extensions.
	RuntimeConfig = aiplatformpb.RuntimeConfig
)

// Authentication type constants mapped from protobuf
const (
	// AuthTypeUnspecified indicates unspecified authentication.
	AuthTypeUnspecified = aiplatformpb.AuthType_AUTH_TYPE_UNSPECIFIED

	// AuthTypeNoAuth indicates no authentication.
	AuthTypeNoAuth = aiplatformpb.AuthType_NO_AUTH

	// AuthTypeAPIKey uses API key authentication.
	AuthTypeAPIKey = aiplatformpb.AuthType_API_KEY_AUTH

	// AuthTypeHTTPBasic uses HTTP Basic authentication.
	AuthTypeHTTPBasic = aiplatformpb.AuthType_HTTP_BASIC_AUTH

	// AuthTypeGoogleServiceAccount uses Google Service Account authentication.
	AuthTypeGoogleServiceAccount = aiplatformpb.AuthType_GOOGLE_SERVICE_ACCOUNT_AUTH

	// AuthTypeOAuth uses OAuth authentication.
	AuthTypeOAuth = aiplatformpb.AuthType_OAUTH

	// AuthTypeOIDC uses OpenID Connect authentication.
	AuthTypeOIDC = aiplatformpb.AuthType_OIDC_AUTH
)

// Helper functions for creating auth configs

// NewGoogleServiceAccountConfig creates a new Google Service Account auth config.
func NewGoogleServiceAccountConfig(serviceAccount string) *AuthConfig {
	return &AuthConfig{
		AuthType: AuthTypeGoogleServiceAccount,
		AuthConfig: &aiplatformpb.AuthConfig_GoogleServiceAccountConfig_{
			GoogleServiceAccountConfig: &aiplatformpb.AuthConfig_GoogleServiceAccountConfig{
				ServiceAccount: serviceAccount,
			},
		},
	}
}

// NewAPIKeyConfig creates a new API key auth config.
func NewAPIKeyConfig(secretName, header string) *AuthConfig {
	return &AuthConfig{
		AuthType: AuthTypeAPIKey,
		AuthConfig: &aiplatformpb.AuthConfig_ApiKeyConfig_{
			ApiKeyConfig: &aiplatformpb.AuthConfig_ApiKeyConfig{
				ApiKeySecret: secretName,
				Name:         header,
			},
		},
	}
}

// NewHTTPBasicAuthConfig creates a new HTTP Basic auth config.
func NewHTTPBasicAuthConfig(credentialSecret string) *AuthConfig {
	return &AuthConfig{
		AuthType: AuthTypeHTTPBasic,
		AuthConfig: &aiplatformpb.AuthConfig_HttpBasicAuthConfig_{
			HttpBasicAuthConfig: &aiplatformpb.AuthConfig_HttpBasicAuthConfig{
				CredentialSecret: credentialSecret,
			},
		},
	}
}

// NewOAuthConfigWithAccessToken creates a new OAuth auth config with access token.
func NewOAuthConfigWithAccessToken(accessToken string) *AuthConfig {
	return &AuthConfig{
		AuthType: AuthTypeOAuth,
		AuthConfig: &aiplatformpb.AuthConfig_OauthConfig_{
			OauthConfig: &aiplatformpb.AuthConfig_OauthConfig{
				OauthConfig: &aiplatformpb.AuthConfig_OauthConfig_AccessToken{
					AccessToken: accessToken,
				},
			},
		},
	}
}

// NewOAuthConfigWithServiceAccount creates a new OAuth auth config with service account.
func NewOAuthConfigWithServiceAccount(serviceAccount string) *AuthConfig {
	return &AuthConfig{
		AuthType: AuthTypeOAuth,
		AuthConfig: &aiplatformpb.AuthConfig_OauthConfig_{
			OauthConfig: &aiplatformpb.AuthConfig_OauthConfig{
				OauthConfig: &aiplatformpb.AuthConfig_OauthConfig_ServiceAccount{
					ServiceAccount: serviceAccount,
				},
			},
		},
	}
}

// Helper functions for creating runtime configs

// NewCodeInterpreterRuntimeConfig creates a new code interpreter runtime config.
func NewCodeInterpreterRuntimeConfig(inputBucket, outputBucket string) *RuntimeConfig {
	return &RuntimeConfig{
		GoogleFirstPartyExtensionConfig: &aiplatformpb.RuntimeConfig_CodeInterpreterRuntimeConfig_{
			CodeInterpreterRuntimeConfig: &aiplatformpb.RuntimeConfig_CodeInterpreterRuntimeConfig{
				FileInputGcsBucket:  inputBucket,
				FileOutputGcsBucket: outputBucket,
			},
		},
	}
}

// NewVertexAISearchRuntimeConfig creates a new Vertex AI Search runtime config.
func NewVertexAISearchRuntimeConfig(servingConfigName, engineID string) *RuntimeConfig {
	return &RuntimeConfig{
		GoogleFirstPartyExtensionConfig: &aiplatformpb.RuntimeConfig_VertexAiSearchRuntimeConfig{
			VertexAiSearchRuntimeConfig: &aiplatformpb.RuntimeConfig_VertexAISearchRuntimeConfig{
				ServingConfigName: servingConfigName,
				EngineId:          engineID,
			},
		},
	}
}

// Helper functions for creating extension manifests

// NewExtensionManifest creates a new extension manifest.
func NewExtensionManifest(name, description, openAPIGCSURI string, authConfig *AuthConfig) *ExtensionManifest {
	return &ExtensionManifest{
		Name:        name,
		Description: description,
		ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
			ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
				OpenApiGcsUri: openAPIGCSURI,
			},
		},
		AuthConfig: authConfig,
	}
}

// NewExtensionManifestWithYAML creates a new extension manifest with inline YAML.
func NewExtensionManifestWithYAML(name, description, openAPIYAML string, authConfig *AuthConfig) *ExtensionManifest {
	return &ExtensionManifest{
		Name:        name,
		Description: description,
		ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
			ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiYaml{
				OpenApiYaml: openAPIYAML,
			},
		},
		AuthConfig: authConfig,
	}
}

// Request and Response Types - using protobuf types
type (
	// ImportExtensionRequest contains parameters for importing/creating a new extension.
	// This replaces the custom CreateExtensionRequest to align with the protobuf API.
	ImportExtensionRequest = aiplatformpb.ImportExtensionRequest

	// ListExtensionsRequest contains parameters for listing extensions.
	ListExtensionsRequest = aiplatformpb.ListExtensionsRequest

	// ListExtensionsResponse contains the response from listing extensions.
	ListExtensionsResponse = aiplatformpb.ListExtensionsResponse

	// GetExtensionRequest contains parameters for getting a specific extension.
	GetExtensionRequest = aiplatformpb.GetExtensionRequest

	// DeleteExtensionRequest contains parameters for deleting an extension.
	DeleteExtensionRequest = aiplatformpb.DeleteExtensionRequest

	// UpdateExtensionRequest contains parameters for updating an extension.
	UpdateExtensionRequest = aiplatformpb.UpdateExtensionRequest

	// ExecuteExtensionRequest contains parameters for executing an extension operation.
	ExecuteExtensionRequest = aiplatformpb.ExecuteExtensionRequest

	// ExecuteExtensionResponse contains the result of executing an extension operation.
	ExecuteExtensionResponse = aiplatformpb.ExecuteExtensionResponse

	// QueryExtensionRequest contains parameters for querying extension capabilities.
	QueryExtensionRequest = aiplatformpb.QueryExtensionRequest

	// QueryExtensionResponse contains the response from querying extension capabilities.
	QueryExtensionResponse = aiplatformpb.QueryExtensionResponse
)

// Helper functions for creating requests

// NewImportExtensionRequest creates a new import extension request.
func NewImportExtensionRequest(parent, displayName, description string, manifest *ExtensionManifest, runtimeConfig *RuntimeConfig) *ImportExtensionRequest {
	ext := &aiplatformpb.Extension{
		DisplayName:   displayName,
		Description:   description,
		Manifest:      manifest,
		RuntimeConfig: runtimeConfig,
	}

	return &ImportExtensionRequest{
		Parent:    parent,
		Extension: ext,
	}
}

// NewListExtensionsRequest creates a new list extensions request.
func NewListExtensionsRequest(parent string, pageSize int32, pageToken, filter, orderBy string) *ListExtensionsRequest {
	return &ListExtensionsRequest{
		Parent:    parent,
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
		OrderBy:   orderBy,
	}
}

// NewGetExtensionRequest creates a new get extension request.
func NewGetExtensionRequest(name string) *GetExtensionRequest {
	return &GetExtensionRequest{
		Name: name,
	}
}

// NewDeleteExtensionRequest creates a new delete extension request.
func NewDeleteExtensionRequest(name string) *DeleteExtensionRequest {
	return &DeleteExtensionRequest{
		Name: name,
	}
}

// // Backward compatibility types
//
// // CreateExtensionRequest provides backward compatibility with the old API.
// // Deprecated: Use ImportExtensionRequest instead.
// type CreateExtensionRequest struct {
// 	// DisplayName is the human-readable name for the extension.
// 	DisplayName string `json:"display_name"`
//
// 	// Description provides details about the extension's purpose.
// 	Description string `json:"description"`
//
// 	// Manifest defines the extension's configuration.
// 	Manifest *ExtensionManifest `json:"manifest"`
//
// 	// RuntimeConfig contains runtime-specific configuration.
// 	RuntimeConfig *RuntimeConfig `json:"runtime_config,omitempty"`
// }
//
// // ToImportRequest converts a CreateExtensionRequest to an ImportExtensionRequest.
// func (r *CreateExtensionRequest) ToImportRequest(parent string) *ImportExtensionRequest {
// 	return NewImportExtensionRequest(parent, r.DisplayName, r.Description, r.Manifest, r.RuntimeConfig)
// }

// Prebuilt Extension Types

// PrebuiltExtensionType represents the type of prebuilt extension.
type PrebuiltExtensionType string

const (
	// PrebuiltExtensionCodeInterpreter is the code interpreter extension.
	PrebuiltExtensionCodeInterpreter PrebuiltExtensionType = "code_interpreter"

	// PrebuiltExtensionVertexAISearch is the Vertex AI Search extension.
	PrebuiltExtensionVertexAISearch PrebuiltExtensionType = "vertex_ai_search"
)

// CodeInterpreterExecutionRequest contains parameters for code interpreter execution.
type CodeInterpreterExecutionRequest struct {
	// Query is the query or task for the code interpreter.
	Query string `json:"query"`

	// Files are optional input files for the execution.
	Files []string `json:"files,omitempty"`
}

// CodeInterpreterExecutionResponse contains the result of code interpreter execution.
type CodeInterpreterExecutionResponse struct {
	// GeneratedCode is the code generated by the interpreter.
	GeneratedCode string `json:"generated_code"`

	// ExecutionResult is the result of executing the code.
	ExecutionResult string `json:"execution_result"`

	// ExecutionError contains any execution errors.
	ExecutionError string `json:"execution_error"`

	// OutputFiles are any files generated during execution.
	OutputFiles []string `json:"output_files"`
}

// VertexAISearchExecutionRequest contains parameters for Vertex AI Search execution.
type VertexAISearchExecutionRequest struct {
	// Query is the search query.
	Query string `json:"query"`

	// MaxResults is the maximum number of results to return.
	MaxResults int32 `json:"max_results,omitempty"`
}

// VertexAISearchExecutionResponse contains the result of Vertex AI Search execution.
type VertexAISearchExecutionResponse struct {
	// Results are the search results.
	Results []SearchResult `json:"results"`

	// NextPageToken is the token for the next page of results.
	NextPageToken string `json:"next_page_token,omitempty"`
}

// SearchResult represents a single search result.
// This remains as a custom type since it's used for convenience functions
// and doesn't have a direct protobuf equivalent in the extension context.
type SearchResult struct {
	// ID is the unique identifier for the result.
	ID string `json:"id"`

	// Title is the title of the result.
	Title string `json:"title"`

	// Content is the content of the result.
	Content string `json:"content"`

	// URI is the URI of the source document.
	URI string `json:"uri"`

	// Score is the relevance score.
	Score float64 `json:"score"`

	// Metadata contains additional metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}
