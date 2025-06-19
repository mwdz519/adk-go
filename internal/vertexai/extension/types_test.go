// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"testing"
	"time"

	"cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestExtension_GetID(t *testing.T) {
	// Test the GetID method extracts ID from resource name
	extensionName := "projects/test-project/locations/us-central1/extensions/ext_123"
	ext := &Extension{
		Extension: &aiplatformpb.Extension{
			Name:        extensionName,
			DisplayName: "Test Extension",
			Description: "A test extension",
			Manifest: NewExtensionManifest(
				"test_extension",
				"Test extension manifest",
				"gs://test-bucket/openapi.yaml",
				NewGoogleServiceAccountConfig(""),
			),
			RuntimeConfig: NewCodeInterpreterRuntimeConfig("", ""),
			CreateTime:    timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			UpdateTime:    timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		State: ExtensionStateActive,
	}

	// Test GetID extraction
	expectedID := "ext_123"
	if ext.GetID() != expectedID {
		t.Errorf("Extension.GetID() = %v, want %v", ext.GetID(), expectedID)
	}

	// Test GetCreateTimeAsTime and GetUpdateTimeAsTime
	expectedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !ext.GetCreateTimeAsTime().Equal(expectedTime) {
		t.Errorf("Extension.GetCreateTimeAsTime() = %v, want %v", ext.GetCreateTimeAsTime(), expectedTime)
	}
	if !ext.GetUpdateTimeAsTime().Equal(expectedTime) {
		t.Errorf("Extension.GetUpdateTimeAsTime() = %v, want %v", ext.GetUpdateTimeAsTime(), expectedTime)
	}
}

func TestManifest_Creation(t *testing.T) {
	tests := []struct {
		name     string
		manifest *ExtensionManifest
		isValid  bool
	}{
		{
			name: "valid manifest with GCS URI",
			manifest: NewExtensionManifest(
				"test_extension",
				"Test extension",
				"gs://test-bucket/openapi.yaml",
				NewGoogleServiceAccountConfig(""),
			),
			isValid: true,
		},
		{
			name: "manifest with OAuth access token auth",
			manifest: NewExtensionManifest(
				"oauth_extension",
				"Extension with OAuth",
				"gs://test-bucket/openapi.yaml",
				NewOAuthConfigWithAccessToken("access-token-123"),
			),
			isValid: true,
		},
		{
			name: "manifest with API key auth",
			manifest: NewExtensionManifest(
				"apikey_extension",
				"Extension with API key",
				"gs://test-bucket/openapi.yaml",
				NewAPIKeyConfig("api-key-secret", "X-API-Key"),
			),
			isValid: true,
		},
		{
			name: "manifest with YAML specification",
			manifest: NewExtensionManifestWithYAML(
				"yaml_extension",
				"Extension with inline YAML",
				"openapi: 3.0.0\ninfo:\n  title: Test API\n  version: 1.0.0",
				NewGoogleServiceAccountConfig(""),
			),
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify manifest is created correctly
			if tt.manifest == nil {
				t.Errorf("Expected manifest to be created, got nil")
				return
			}

			// Check required fields
			if tt.manifest.Name == "" {
				t.Errorf("Expected manifest name to be set")
			}
			if tt.manifest.Description == "" {
				t.Errorf("Expected manifest description to be set")
			}
			if tt.manifest.ApiSpec == nil {
				t.Errorf("Expected manifest API spec to be set")
			}
			if tt.manifest.AuthConfig == nil {
				t.Errorf("Expected manifest auth config to be set")
			}
		})
	}
}

func TestRuntimeConfig_Creation(t *testing.T) {
	tests := []struct {
		name   string
		config *RuntimeConfig
	}{
		{
			name:   "code interpreter config",
			config: NewCodeInterpreterRuntimeConfig("input-bucket", "output-bucket"),
		},
		{
			name: "vertex ai search config",
			config: NewVertexAISearchRuntimeConfig(
				"projects/test/locations/us-central1/collections/default/engines/search/servingConfigs/default",
				"engine-123",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify config is created correctly
			if tt.config == nil {
				t.Errorf("Expected runtime config to be created, got nil")
				return
			}

			// Check that the GoogleFirstPartyExtensionConfig is set
			if tt.config.GoogleFirstPartyExtensionConfig == nil {
				t.Errorf("Expected GoogleFirstPartyExtensionConfig to be set")
			}
		})
	}
}

func TestRequestResponse_Creation(t *testing.T) {
	// Test CreateExtensionRequest (backward compatibility)
	createReq := &CreateExtensionRequest{
		DisplayName: "Test Extension",
		Description: "A test extension",
		Manifest: NewExtensionManifest(
			"test_extension",
			"Test extension",
			"gs://test-bucket/openapi.yaml",
			NewGoogleServiceAccountConfig(""),
		),
	}

	// Test conversion to ImportExtensionRequest
	importReq := createReq.ToImportRequest("projects/test-project/locations/us-central1")
	if importReq == nil {
		t.Errorf("Expected ToImportRequest to return non-nil request")
	}
	if importReq.Parent != "projects/test-project/locations/us-central1" {
		t.Errorf("ImportExtensionRequest.Parent = %v, want projects/test-project/locations/us-central1", importReq.Parent)
	}
	if importReq.Extension.DisplayName != "Test Extension" {
		t.Errorf("ImportExtensionRequest.Extension.DisplayName = %v, want Test Extension", importReq.Extension.DisplayName)
	}

	// Test helper function for creating requests
	listReq := NewListExtensionsRequest(
		"projects/test-project/locations/us-central1",
		10,
		"token123",
		"state=ACTIVE",
		"create_time desc",
	)
	if listReq.Parent != "projects/test-project/locations/us-central1" {
		t.Errorf("ListExtensionsRequest.Parent = %v, want projects/test-project/locations/us-central1", listReq.Parent)
	}
	if listReq.PageSize != 10 {
		t.Errorf("ListExtensionsRequest.PageSize = %v, want 10", listReq.PageSize)
	}
	if listReq.PageToken != "token123" {
		t.Errorf("ListExtensionsRequest.PageToken = %v, want token123", listReq.PageToken)
	}

	// Test GetExtensionRequest
	getReq := NewGetExtensionRequest("projects/test-project/locations/us-central1/extensions/ext_123")
	if getReq.Name != "projects/test-project/locations/us-central1/extensions/ext_123" {
		t.Errorf("GetExtensionRequest.Name = %v, want projects/test-project/locations/us-central1/extensions/ext_123", getReq.Name)
	}

	// Test DeleteExtensionRequest
	deleteReq := NewDeleteExtensionRequest("projects/test-project/locations/us-central1/extensions/ext_123")
	if deleteReq.Name != "projects/test-project/locations/us-central1/extensions/ext_123" {
		t.Errorf("DeleteExtensionRequest.Name = %v, want projects/test-project/locations/us-central1/extensions/ext_123", deleteReq.Name)
	}
}

func TestPrebuiltExtensionTypes(t *testing.T) {
	tests := []struct {
		name          string
		extensionType PrebuiltExtensionType
		expected      string
	}{
		{
			name:          "code interpreter",
			extensionType: PrebuiltExtensionCodeInterpreter,
			expected:      "code_interpreter",
		},
		{
			name:          "vertex ai search",
			extensionType: PrebuiltExtensionVertexAISearch,
			expected:      "vertex_ai_search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.extensionType) != tt.expected {
				t.Errorf("PrebuiltExtensionType %v = %v, want %v", tt.name, string(tt.extensionType), tt.expected)
			}
		})
	}
}

func TestExtensionState(t *testing.T) {
	tests := []struct {
		name     string
		state    ExtensionState
		expected string
	}{
		{
			name:     "unspecified",
			state:    ExtensionStateUnspecified,
			expected: "EXTENSION_STATE_UNSPECIFIED",
		},
		{
			name:     "active",
			state:    ExtensionStateActive,
			expected: "ACTIVE",
		},
		{
			name:     "creating",
			state:    ExtensionStateCreating,
			expected: "CREATING",
		},
		{
			name:     "deleting",
			state:    ExtensionStateDeleting,
			expected: "DELETING",
		},
		{
			name:     "error",
			state:    ExtensionStateError,
			expected: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.state) != tt.expected {
				t.Errorf("ExtensionState %v = %v, want %v", tt.name, string(tt.state), tt.expected)
			}
		})
	}
}

func TestAuthType(t *testing.T) {
	tests := []struct {
		name     string
		authType AuthType
		expected string
	}{
		{
			name:     "unspecified",
			authType: AuthTypeUnspecified,
			expected: "AUTH_TYPE_UNSPECIFIED",
		},
		{
			name:     "no auth",
			authType: AuthTypeNoAuth,
			expected: "NO_AUTH",
		},
		{
			name:     "api key",
			authType: AuthTypeAPIKey,
			expected: "API_KEY_AUTH",
		},
		{
			name:     "http basic",
			authType: AuthTypeHTTPBasic,
			expected: "HTTP_BASIC_AUTH",
		},
		{
			name:     "google service account",
			authType: AuthTypeGoogleServiceAccount,
			expected: "GOOGLE_SERVICE_ACCOUNT_AUTH",
		},
		{
			name:     "oauth",
			authType: AuthTypeOAuth,
			expected: "OAUTH",
		},
		{
			name:     "oidc",
			authType: AuthTypeOIDC,
			expected: "OIDC_AUTH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.authType.String() != tt.expected {
				t.Errorf("AuthType %v = %v, want %v", tt.name, tt.authType.String(), tt.expected)
			}
		})
	}
}

func TestCodeInterpreterExecutionResponse_Fields(t *testing.T) {
	response := &CodeInterpreterExecutionResponse{
		GeneratedCode:   "import pandas as pd\ndf = pd.DataFrame([1,2,3,4,5])\nprint(df.max())",
		ExecutionResult: "5",
		ExecutionError:  "",
		OutputFiles:     []string{"output.csv", "chart.png"},
	}

	// Test that the response fields are accessible
	if response.GeneratedCode != "import pandas as pd\ndf = pd.DataFrame([1,2,3,4,5])\nprint(df.max())" {
		t.Errorf("GeneratedCode mismatch")
	}
	if response.ExecutionResult != "5" {
		t.Errorf("ExecutionResult mismatch")
	}
	if len(response.OutputFiles) != 2 {
		t.Errorf("OutputFiles length = %v, want 2", len(response.OutputFiles))
	}
}

func TestVertexAISearchExecutionResponse_Fields(t *testing.T) {
	response := &VertexAISearchExecutionResponse{
		Results: []SearchResult{
			{
				ID:      "result1",
				Title:   "Test Document 1",
				Content: "This is the content of the first document",
				URI:     "gs://bucket/doc1.pdf",
				Score:   0.95,
				Metadata: map[string]any{
					"author": "John Doe",
					"year":   float64(2024), // Use float64 to match JSON unmarshaling behavior
				},
			},
			{
				ID:      "result2",
				Title:   "Test Document 2",
				Content: "This is the content of the second document",
				URI:     "gs://bucket/doc2.pdf",
				Score:   0.87,
				Metadata: map[string]any{
					"author": "Jane Smith",
					"year":   float64(2023), // Use float64 to match JSON unmarshaling behavior
				},
			},
		},
		NextPageToken: "next_page_123",
	}

	// Test that the response fields are accessible
	if len(response.Results) != 2 {
		t.Errorf("Results length = %v, want 2", len(response.Results))
	}
	if response.NextPageToken != "next_page_123" {
		t.Errorf("NextPageToken = %v, want next_page_123", response.NextPageToken)
	}

	// Test first result
	firstResult := response.Results[0]
	if firstResult.ID != "result1" {
		t.Errorf("First result ID = %v, want result1", firstResult.ID)
	}
	if firstResult.Score != 0.95 {
		t.Errorf("First result Score = %v, want 0.95", firstResult.Score)
	}
}
