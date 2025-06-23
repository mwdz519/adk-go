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

// Helper functions for creating auth configs

// NewGoogleServiceAccountConfig creates a new Google Service Account auth config.
func NewGoogleServiceAccountConfig(serviceAccount string) *aiplatformpb.AuthConfig {
	return &aiplatformpb.AuthConfig{
		AuthType: aiplatformpb.AuthType_GOOGLE_SERVICE_ACCOUNT_AUTH,
		AuthConfig: &aiplatformpb.AuthConfig_GoogleServiceAccountConfig_{
			GoogleServiceAccountConfig: &aiplatformpb.AuthConfig_GoogleServiceAccountConfig{
				ServiceAccount: serviceAccount,
			},
		},
	}
}

// NewAPIKeyConfig creates a new API key auth config.
func NewAPIKeyConfig(secretName, header string) *aiplatformpb.AuthConfig {
	return &aiplatformpb.AuthConfig{
		AuthType: aiplatformpb.AuthType_API_KEY_AUTH,
		AuthConfig: &aiplatformpb.AuthConfig_ApiKeyConfig_{
			ApiKeyConfig: &aiplatformpb.AuthConfig_ApiKeyConfig{
				ApiKeySecret: secretName,
				Name:         header,
			},
		},
	}
}

// NewHTTPBasicAuthConfig creates a new HTTP Basic auth config.
func NewHTTPBasicAuthConfig(credentialSecret string) *aiplatformpb.AuthConfig {
	return &aiplatformpb.AuthConfig{
		AuthType: aiplatformpb.AuthType_HTTP_BASIC_AUTH,
		AuthConfig: &aiplatformpb.AuthConfig_HttpBasicAuthConfig_{
			HttpBasicAuthConfig: &aiplatformpb.AuthConfig_HttpBasicAuthConfig{
				CredentialSecret: credentialSecret,
			},
		},
	}
}

// NewOAuthConfigWithAccessToken creates a new OAuth auth config with access token.
func NewOAuthConfigWithAccessToken(accessToken string) *aiplatformpb.AuthConfig {
	return &aiplatformpb.AuthConfig{
		AuthType: aiplatformpb.AuthType_OAUTH,
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
func NewOAuthConfigWithServiceAccount(serviceAccount string) *aiplatformpb.AuthConfig {
	return &aiplatformpb.AuthConfig{
		AuthType: aiplatformpb.AuthType_OAUTH,
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
func NewCodeInterpreterRuntimeConfig(inputBucket, outputBucket string) *aiplatformpb.RuntimeConfig {
	return &aiplatformpb.RuntimeConfig{
		GoogleFirstPartyExtensionConfig: &aiplatformpb.RuntimeConfig_CodeInterpreterRuntimeConfig_{
			CodeInterpreterRuntimeConfig: &aiplatformpb.RuntimeConfig_CodeInterpreterRuntimeConfig{
				FileInputGcsBucket:  inputBucket,
				FileOutputGcsBucket: outputBucket,
			},
		},
	}
}

// NewVertexAISearchRuntimeConfig creates a new Vertex AI Search runtime config.
func NewVertexAISearchRuntimeConfig(servingConfigName, engineID string) *aiplatformpb.RuntimeConfig {
	return &aiplatformpb.RuntimeConfig{
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
func NewExtensionManifest(name, description, openAPIGCSURI string, authConfig *aiplatformpb.AuthConfig) *aiplatformpb.ExtensionManifest {
	return &aiplatformpb.ExtensionManifest{
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
func NewExtensionManifestWithYAML(name, description, openAPIYAML string, authConfig *aiplatformpb.AuthConfig) *aiplatformpb.ExtensionManifest {
	return &aiplatformpb.ExtensionManifest{
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

// Helper functions for creating requests

// NewImportExtensionRequest creates a new import extension request.
func NewImportExtensionRequest(parent, displayName, description string, manifest *aiplatformpb.ExtensionManifest, runtimeConfig *aiplatformpb.RuntimeConfig) *aiplatformpb.ImportExtensionRequest {
	ext := &aiplatformpb.Extension{
		DisplayName:   displayName,
		Description:   description,
		Manifest:      manifest,
		RuntimeConfig: runtimeConfig,
	}

	return &aiplatformpb.ImportExtensionRequest{
		Parent:    parent,
		Extension: ext,
	}
}

// NewListExtensionsRequest creates a new list extensions request.
func NewListExtensionsRequest(parent string, pageSize int32, pageToken, filter, orderBy string) *aiplatformpb.ListExtensionsRequest {
	return &aiplatformpb.ListExtensionsRequest{
		Parent:    parent,
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
		OrderBy:   orderBy,
	}
}

// NewGetExtensionRequest creates a new get extension request.
func NewGetExtensionRequest(name string) *aiplatformpb.GetExtensionRequest {
	return &aiplatformpb.GetExtensionRequest{
		Name: name,
	}
}

// NewDeleteExtensionRequest creates a new delete extension request.
func NewDeleteExtensionRequest(name string) *aiplatformpb.DeleteExtensionRequest {
	return &aiplatformpb.DeleteExtensionRequest{
		Name: name,
	}
}

// Prebuilt Extension Types

// PrebuiltExtensionType represents the type of prebuilt extension.
type PrebuiltExtensionType string

const (
	// PrebuiltExtensionCodeInterpreter is the code interpreter extension.
	PrebuiltExtensionCodeInterpreter PrebuiltExtensionType = "code_interpreter"

	// PrebuiltExtensionVertexAISearch is the Vertex AI Search extension.
	PrebuiltExtensionVertexAISearch PrebuiltExtensionType = "vertex_ai_search"

	// PrebuiltExtensionWebpageBrowser is the webpage browser extension.
	PrebuiltExtensionWebpageBrowser PrebuiltExtensionType = "webpage_browser"
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
