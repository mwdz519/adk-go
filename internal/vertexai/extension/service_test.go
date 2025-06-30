// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"fmt"
	"strings"
	"testing"

	"cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		location    string
		wantErr     bool
		expectedErr string
	}{
		{
			name:      "valid us-central1",
			projectID: "test-project",
			location:  "us-central1",
			wantErr:   false,
		},
		{
			name:        "empty project ID",
			projectID:   "",
			location:    "us-central1",
			wantErr:     true,
			expectedErr: "projectID is required",
		},
		{
			name:        "empty location",
			projectID:   "test-project",
			location:    "",
			wantErr:     true,
			expectedErr: "location is required",
		},
		{
			name:        "invalid region",
			projectID:   "test-project",
			location:    "us-east1",
			wantErr:     true,
			expectedErr: "extension API is not supported in region 'us-east1'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			service, err := NewService(ctx, tt.projectID, tt.location)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewService() expected error but got none")
					return
				}
				if tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("NewService() error = %v, want error containing %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewService() unexpected error = %v", err)
				return
			}

			if service == nil {
				t.Errorf("NewService() returned nil service")
				return
			}

			// Verify service configuration
			if service.GetProjectID() != tt.projectID {
				t.Errorf("NewService() projectID = %v, want %v", service.GetProjectID(), tt.projectID)
			}
			if service.GetLocation() != tt.location {
				t.Errorf("NewService() location = %v, want %v", service.GetLocation(), tt.location)
			}

			// Clean up
			service.Close()
		})
	}
}

func TestService_CreateExtension(t *testing.T) {
	t.Skip("requires Google Cloud credentials and existing corpus")

	ctx := t.Context()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name    string
		req     *aiplatformpb.ImportExtensionRequest
		wantErr bool
	}{
		{
			name: "valid extension",
			req: &aiplatformpb.ImportExtensionRequest{
				Parent: fmt.Sprintf("projects/%s/locations/%s", service.GetProjectID(), service.GetLocation()),
				Extension: &aiplatformpb.Extension{
					Name:        "test-extension",
					DisplayName: "Test Extension",
					Description: "A test extension",
					Manifest: &aiplatformpb.ExtensionManifest{
						Name:        "test_extension",
						Description: "Test extension for validation",
						ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
							ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
								OpenApiGcsUri: "gs://test-bucket/openapi.yaml",
							},
						},
						AuthConfig: &aiplatformpb.AuthConfig{
							AuthConfig: &aiplatformpb.AuthConfig_GoogleServiceAccountConfig_{},
							AuthType:   aiplatformpb.AuthType_GOOGLE_SERVICE_ACCOUNT_AUTH,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil req",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty display name",
			req: &aiplatformpb.ImportExtensionRequest{
				Parent: fmt.Sprintf("projects/%s/locations/%s", service.GetProjectID(), service.GetLocation()),
				Extension: &aiplatformpb.Extension{
					Name:        "test-extension",
					DisplayName: "",
					Description: "A test extension",
					Manifest: &aiplatformpb.ExtensionManifest{
						Name:        "test_extension",
						Description: "Test extension for validation",
						ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
							ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
								OpenApiGcsUri: "gs://test-bucket/openapi.yaml",
							},
						},
						AuthConfig: NewGoogleServiceAccountConfig(""),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "nil manifest",
			req: &aiplatformpb.ImportExtensionRequest{
				Parent: fmt.Sprintf("projects/%s/locations/%s", service.GetProjectID(), service.GetLocation()),
				Extension: &aiplatformpb.Extension{
					Name:        "test-extension",
					DisplayName: "Test Extension",
					Description: "A test extension",
					Manifest:    nil,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid manifest - missing name",
			req: &aiplatformpb.ImportExtensionRequest{
				Parent: fmt.Sprintf("projects/%s/locations/%s", service.GetProjectID(), service.GetLocation()),
				Extension: &aiplatformpb.Extension{
					Name:        "",
					DisplayName: "Test Extension",
					Description: "Test extension for validation",
					Manifest: &aiplatformpb.ExtensionManifest{
						Name:        "test_extension",
						Description: "Test extension",
						ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
							ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
								OpenApiGcsUri: "gs://test-bucket/openapi.yaml",
							},
						},
						AuthConfig: NewGoogleServiceAccountConfig(""),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid manifest - invalid GCS URI",
			req: &aiplatformpb.ImportExtensionRequest{
				Parent: fmt.Sprintf("projects/%s/locations/%s", service.GetProjectID(), service.GetLocation()),
				Extension: &aiplatformpb.Extension{
					Name:        "test-extension",
					DisplayName: "Test Extension",
					Description: "Test extension for validation",
					Manifest: &aiplatformpb.ExtensionManifest{
						Name:        "test_extension",
						Description: "Test extension",
						ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
							ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
								OpenApiGcsUri: "invalid-uri",
							},
						},
						AuthConfig: NewGoogleServiceAccountConfig(""),
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := service.CreateExtension(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateExtension() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateExtension() unexpected error = %v", err)
				return
			}

			if ext == nil {
				t.Errorf("CreateExtension() returned nil extension")
				return
			}

			// Verify extension properties
			if ext.DisplayName != tt.req.GetExtension().GetDisplayName() {
				t.Errorf("CreateExtension() displayName = %v, want %v", ext.DisplayName, tt.req.GetExtension().GetDisplayName())
			}
			if ext.Description != tt.req.GetExtension().GetDescription() {
				t.Errorf("CreateExtension() description = %v, want %v", ext.Description, tt.req.GetExtension().GetDescription())
			}
			if ext.State != ExtensionStateActive {
				t.Errorf("CreateExtension() state = %v, want %v", ext.State, ExtensionStateActive)
			}
		})
	}
}

func TestService_CreateFromHub(t *testing.T) {
	t.Skip("requires Google Cloud credentials and existing corpus")

	ctx := t.Context()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name          string
		extensionType PrebuiltExtensionType
		runtimeConfig *aiplatformpb.RuntimeConfig
		wantErr       bool
	}{
		{
			name:          "code interpreter",
			extensionType: PrebuiltExtensionCodeInterpreter,
			runtimeConfig: &aiplatformpb.RuntimeConfig{
				GoogleFirstPartyExtensionConfig: &aiplatformpb.RuntimeConfig_CodeInterpreterRuntimeConfig_{
					CodeInterpreterRuntimeConfig: &aiplatformpb.RuntimeConfig_CodeInterpreterRuntimeConfig{
						FileInputGcsBucket:  "test-input-bucket",
						FileOutputGcsBucket: "test-output-bucket",
					},
				},
			},
			wantErr: false,
		},
		{
			name:          "vertex ai search",
			extensionType: PrebuiltExtensionVertexAISearch,
			wantErr:       false,
		},
		{
			name:          "webpage browser",
			extensionType: PrebuiltExtensionWebpageBrowser,
			wantErr:       false,
		},
		{
			name:          "invalid extension type",
			extensionType: "invalid_type",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := service.CreateFromHub(ctx, tt.extensionType, tt.runtimeConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateFromHub() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateFromHub() unexpected error = %v", err)
				return
			}

			if ext == nil {
				t.Errorf("CreateFromHub() returned nil extension")
				return
			}

			// Verify extension was created with expected properties
			expectedDisplayName := service.getPrebuiltDisplayName(tt.extensionType)
			if ext.DisplayName != expectedDisplayName {
				t.Errorf("CreateFromHub() displayName = %v, want %v", ext.DisplayName, expectedDisplayName)
			}
		})
	}
}

func TestService_ExecuteExtension(t *testing.T) {
	ctx := t.Context()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name    string
		req     func(t *testing.T) *aiplatformpb.ExecuteExtensionRequest
		wantErr bool
	}{
		{
			name: "valid execution",
			req: func(t *testing.T) *aiplatformpb.ExecuteExtensionRequest {
				operationParams, err := structpb.NewStruct(map[string]any{
					"param1": "value1",
					"param2": 42,
				})
				if err != nil {
					t.Fatal(err)
				}
				return &aiplatformpb.ExecuteExtensionRequest{
					Name:            "projects/test-project/locations/us-central1/extensions/ext_123",
					OperationId:     "test_operation",
					OperationParams: operationParams,
				}
			},
			wantErr: false,
		},
		{
			name: "nil request",
			req: func(t *testing.T) *aiplatformpb.ExecuteExtensionRequest {
				return nil
			},
			wantErr: true,
		},
		{
			name: "empty name",
			req: func(t *testing.T) *aiplatformpb.ExecuteExtensionRequest {
				return &aiplatformpb.ExecuteExtensionRequest{
					Name:        "",
					OperationId: "test_operation",
				}
			},
			wantErr: true,
		},
		{
			name: "empty operation ID",
			req: func(t *testing.T) *aiplatformpb.ExecuteExtensionRequest {
				return &aiplatformpb.ExecuteExtensionRequest{
					Name:        "projects/test-project/locations/us-central1/extensions/ext_123",
					OperationId: "",
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.ExecuteExtension(ctx, tt.req(t))

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExecuteExtension() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ExecuteExtension() unexpected error = %v", err)
				return
			}

			if resp == nil {
				t.Errorf("ExecuteExtension() returned nil response")
				return
			}

			// Verify response has content
			if len(resp.Content) == 0 {
				t.Errorf("ExecuteExtension() response content is empty")
			}
		})
	}
}

func TestGetPrebuiltExtensionConfig(t *testing.T) {
	service := &service{}

	tests := []struct {
		name          string
		extensionType PrebuiltExtensionType
		wantErr       bool
	}{
		{
			name:          "code interpreter",
			extensionType: PrebuiltExtensionCodeInterpreter,
			wantErr:       false,
		},
		{
			name:          "vertex ai search",
			extensionType: PrebuiltExtensionVertexAISearch,
			wantErr:       false,
		},
		{
			name:          "unknown type",
			extensionType: "unknown_type",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, runtimeConfig, err := service.getPrebuiltExtensionConfig(tt.extensionType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("getPrebuiltExtensionConfig() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("getPrebuiltExtensionConfig() unexpected error = %v", err)
				return
			}

			if manifest == nil {
				t.Errorf("getPrebuiltExtensionConfig() manifest is nil")
				return
			}

			if runtimeConfig == nil {
				t.Errorf("getPrebuiltExtensionConfig() runtimeConfig is nil")
				return
			}

			// Verify manifest structure
			if manifest.Name == "" {
				t.Errorf("getPrebuiltExtensionConfig() manifest name is empty")
			}
			if manifest.Description == "" {
				t.Errorf("getPrebuiltExtensionConfig() manifest description is empty")
			}
			if manifest.GetApiSpec() == nil || manifest.GetApiSpec().GetOpenApiGcsUri() == "" {
				t.Errorf("getPrebuiltExtensionConfig() manifest API spec is invalid")
			}
			if manifest.AuthConfig == nil || manifest.AuthConfig.AuthType != aiplatformpb.AuthType_GOOGLE_SERVICE_ACCOUNT_AUTH {
				t.Errorf("getPrebuiltExtensionConfig() manifest auth config is invalid")
			}
		})
	}
}

func TestGetSupportedPrebuiltExtensions(t *testing.T) {
	service := &service{}

	extensions := service.GetSupportedPrebuiltExtensions()

	if len(extensions) == 0 {
		t.Errorf("GetSupportedPrebuiltExtensions() returned empty list")
		return
	}

	expected := []PrebuiltExtensionType{
		PrebuiltExtensionCodeInterpreter,
		PrebuiltExtensionVertexAISearch,
	}

	if diff := cmp.Diff(expected, extensions); diff != "" {
		t.Errorf("GetSupportedPrebuiltExtensions() mismatch (-want +got):\n%s", diff)
	}
}

func TestValidatePrebuiltExtensionType(t *testing.T) {
	service := &service{}

	// Test valid types
	validTypes := []PrebuiltExtensionType{
		PrebuiltExtensionCodeInterpreter,
		PrebuiltExtensionVertexAISearch,
	}

	for _, extensionType := range validTypes {
		err := service.ValidatePrebuiltExtensionType(extensionType)
		if err != nil {
			t.Errorf("ValidatePrebuiltExtensionType(%v) unexpected error = %v", extensionType, err)
		}
	}

	// Test invalid type
	err := service.ValidatePrebuiltExtensionType("invalid_type")
	if err == nil {
		t.Errorf("ValidatePrebuiltExtensionType() expected error for invalid type but got none")
	}
}
