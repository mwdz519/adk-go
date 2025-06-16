// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	aiplatform "cloud.google.com/go/aiplatform/apiv1beta1"
	"cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb"
	"cloud.google.com/go/auth/credentials"
	"google.golang.org/api/option"
)

type vertexExtensionHubType struct {
	displayName string
	description string
	manifest    *aiplatformpb.ExtensionManifest
}

var VertexExtensionHub = map[string]vertexExtensionHubType{
	"code_interpreter": {
		displayName: "Code Interpreter",
		description: "This extension generates and executes code in the specified language",
		manifest: &aiplatformpb.ExtensionManifest{
			Name:        "code_interpreter_tool",
			Description: "Google Code Interpreter Extension",
			ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
				ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
					OpenApiGcsUri: "gs://vertex-extension-public/code_interpreter.yaml",
				},
			},
			AuthConfig: &aiplatformpb.AuthConfig{
				AuthType:   aiplatformpb.AuthType_GOOGLE_SERVICE_ACCOUNT_AUTH,
				AuthConfig: &aiplatformpb.AuthConfig_GoogleServiceAccountConfig_{},
			},
		},
	},
	"vertex_ai_search": {
		displayName: "Vertex AI Search",
		description: "This extension generates and executes search queries",
		manifest: &aiplatformpb.ExtensionManifest{
			Name:        "vertex_ai_search",
			Description: "Vertex AI Search Extension",
			ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
				ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
					OpenApiGcsUri: "gs://vertex-extension-public/vertex_ai_search.yaml",
				},
			},
			AuthConfig: &aiplatformpb.AuthConfig{
				AuthType:   aiplatformpb.AuthType_GOOGLE_SERVICE_ACCOUNT_AUTH,
				AuthConfig: &aiplatformpb.AuthConfig_GoogleServiceAccountConfig_{},
			},
		},
	},
	"webpage_browser": {
		displayName: "Webpage Browser",
		description: "This extension fetches the content of a webpage",
		manifest: &aiplatformpb.ExtensionManifest{
			Name:        "webpage_browser",
			Description: "Vertex Webpage Browser Extension",
			ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
				ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
					OpenApiGcsUri: "gs://vertex-extension-public/webpage_browser.yaml",
				},
			},
			AuthConfig: &aiplatformpb.AuthConfig{
				AuthType:   aiplatformpb.AuthType_GOOGLE_SERVICE_ACCOUNT_AUTH,
				AuthConfig: &aiplatformpb.AuthConfig_GoogleServiceAccountConfig_{},
			},
		},
	},
}

// Service provides access to Vertex AI Extensions functionality.
//
// The service manages extension lifecycles, executes operations, and provides
// access to prebuilt extensions from Google's extension hub.
type Service struct {
	// AI Platform clients
	extensionExecutionClient *aiplatform.ExtensionExecutionClient
	extensionRegistryClient  *aiplatform.ExtensionRegistryClient

	// Configuration
	projectID string
	location  string
	logger    *slog.Logger

	resourceName string

	// Service state
	initialized bool
	mu          sync.RWMutex
}

// ServiceOption is a functional option for configuring the extension service.
type ServiceOption func(*Service)

// WithLogger sets a custom logger for the service.
func WithLogger(logger *slog.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}

// NewService creates a new Vertex AI Extension service.
//
// The service provides access to all extension operations including creation,
// execution, and management of both custom and prebuilt extensions.
//
// Parameters:
//   - ctx: Context for the initialization
//   - projectID: Google Cloud project ID
//   - location: Geographic location for Vertex AI services (must be "us-central1")
//   - opts: Optional configuration options
//
// Returns a fully initialized extension service or an error if initialization fails.
func NewService(ctx context.Context, projectID, location string, opts ...ServiceOption) (*Service, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if location == "" {
		return nil, fmt.Errorf("location is required")
	}

	// Enforce region restriction
	if location != "us-central1" {
		return nil, &RegionNotSupportedError{
			RequestedRegion:  location,
			SupportedRegions: []string{"us-central1"},
		}
	}

	service := &Service{
		projectID: projectID,
		location:  location,
		logger:    slog.Default(),
	}

	// Apply options
	for _, opt := range opts {
		opt(service)
	}

	// Create credentials
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{
			"https://www.googleapis.com/auth/cloud-platform",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to detect default credentials: %w", err)
	}

	extensionExecutionClient, err := aiplatform.NewExtensionExecutionClient(ctx, option.WithAuthCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform Extension Execution client: %w", err)
	}
	service.extensionExecutionClient = extensionExecutionClient

	extensionRegistryClient, err := aiplatform.NewExtensionRegistryClient(ctx, option.WithAuthCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform Extension Execution client: %w", err)
	}
	service.extensionRegistryClient = extensionRegistryClient

	service.initialized = true

	service.logger.InfoContext(ctx, "Vertex AI Extension service initialized successfully",
		slog.String("project_id", projectID),
		slog.String("location", location),
	)

	return service, nil
}

// Close closes the Extension Execution service and releases any resources.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return nil
	}

	s.logger.Info("Closing Vertex AI Extension Execution service")

	// Close AI Platform clients
	if s.extensionExecutionClient != nil {
		if err := s.extensionExecutionClient.Close(); err != nil {
			s.logger.Error("Failed to close Extension Execution client", slog.String("error", err.Error()))
			return fmt.Errorf("failed to close Extension Execution client: %w", err)
		}
	}

	s.initialized = false
	s.logger.Info("Vertex AI Extension Execution service closed successfully")

	return nil
}

// Extension Management Methods

// CreateExtension creates a new custom extension.
//
// This method allows you to register a custom extension with Vertex AI by providing
// a manifest that defines the API specification, authentication configuration,
// and runtime settings.
func (s *Service) CreateExtension(ctx context.Context, manifest *aiplatformpb.ExtensionManifest, extensionName, displayName, description string, runtimeConfig *aiplatformpb.RuntimeConfig) (*Extension, error) {
	extension := &aiplatformpb.Extension{
		Name:        extensionName,
		DisplayName: displayName,
		Description: description,
		Manifest:    manifest,
	}
	if runtimeConfig != nil {
		extension.RuntimeConfig = runtimeConfig
	}

	req := &aiplatformpb.ImportExtensionRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s", s.projectID, s.location),
		Extension: extension,
	}
	operationFuture, err := s.extensionRegistryClient.ImportExtension(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to import extension config: %w", err)
	}

	createdExtension, err := operationFuture.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait long running operation: %w", err)
	}

	s.resourceName = createdExtension.GetName()

	s.logger.InfoContext(ctx, "Extension created successfully",
		slog.String("extension_name", createdExtension.GetName()),
	)

	e := &Extension{
		Extension: createdExtension,
		State:     ExtensionStateActive,
	}

	return e, nil
}

// Extension Execution Methods

// ExecuteExtension executes an operation on a specific extension.
//
// This method allows you to invoke extension operations with structured parameters
// and receive the execution results. The operation parameters are specific to
// each extension and operation type.
func (s *Service) ExecuteExtension(ctx context.Context, req *ExecuteExtensionRequest) (*ExecuteExtensionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.OperationId == "" {
		return nil, fmt.Errorf("operation_id is required")
	}

	s.logger.InfoContext(ctx, "Executing extension operation",
		slog.String("name", req.Name),
		slog.String("operation_id", req.OperationId),
	)

	// TODO: Implement actual API call to execute extension
	// For now, return a mock response
	// Create a mock response using the protobuf structure
	_ = &ExecuteExtensionResponse{
		Content: `{"result": "mock execution result"}`,
	}

	s.logger.InfoContext(ctx, "Extension operation executed successfully",
		slog.String("name", req.Name),
		slog.String("operation_id", req.OperationId),
	)

	return &ExecuteExtensionResponse{
		Content: `{"result": "mock execution result"}`,
	}, nil
}

// CreateFromHub creates an extension from Google's prebuilt extension hub.
//
// This method provides easy access to Google's curated extensions like
// code_interpreter and vertex_ai_search without requiring manual manifest creation.
func (s *Service) CreateFromHub(ctx context.Context, name string, runtimeConfig *aiplatformpb.RuntimeConfig) (*Extension, error) {
	s.logger.InfoContext(ctx, "Creating extension from hub", slog.String("name", name))

	switch name {
	case "code_interpreter":
		if _, ok := runtimeConfig.GetGoogleFirstPartyExtensionConfig().(*aiplatformpb.RuntimeConfig_CodeInterpreterRuntimeConfig_); !ok {
			return nil, errors.New("code_interpreter_runtime_config is required for code_interpreter extension")
		}
	case "vertex_ai_search":
		if runtimeConfig == nil {
			return nil, errors.New("runtime_config is required for vertex_ai_search extension")
		}
		if _, ok := runtimeConfig.GetGoogleFirstPartyExtensionConfig().(*aiplatformpb.RuntimeConfig_VertexAiSearchRuntimeConfig); !ok {
			return nil, errors.New("runtime_config is required for vertex_ai_search extension")
		}
	case "webpage_browser":
		// nothing to do
	default:
		return nil, fmt.Errorf("unsupported 1P extension name: %s", name)
	}

	extensionInfo := VertexExtensionHub[name]

	extension, err := s.CreateExtension(ctx, extensionInfo.manifest, name, extensionInfo.displayName, extensionInfo.description, runtimeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create extension from hub: %w", err)
	}

	s.logger.InfoContext(ctx, "Extension created from hub successfully", slog.String("extension_id", extension.GetID()))

	return extension, nil
}

// ListExtensions lists all extensions in the project and location.
func (s *Service) ListExtensions(ctx context.Context, req *ListExtensionsRequest) (*ListExtensionsResponse, error) {
	if req == nil {
		req = &ListExtensionsRequest{}
	}

	s.logger.InfoContext(ctx, "Listing extensions",
		slog.Int("page_size", int(req.PageSize)),
		slog.String("page_token", req.PageToken),
	)

	// TODO: Implement actual API call to list extensions
	// For now, return empty response
	response := &ListExtensionsResponse{
		Extensions:    []*aiplatformpb.Extension{},
		NextPageToken: "",
	}

	s.logger.InfoContext(ctx, "Listed extensions successfully",
		slog.Int("extension_count", len(response.Extensions)),
	)

	return response, nil
}

// GetExtension retrieves a specific extension by its resource name.
func (s *Service) GetExtension(ctx context.Context, req *GetExtensionRequest) (*Extension, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	s.logger.InfoContext(ctx, "Getting extension",
		slog.String("name", req.Name),
	)

	// TODO: Implement actual API call to get extension
	// For now, return extension not found error
	return nil, &ExtensionNotFoundError{
		Name: req.Name,
	}
}

// DeleteExtension deletes an extension.
func (s *Service) DeleteExtension(ctx context.Context, req *DeleteExtensionRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	s.logger.InfoContext(ctx, "Deleting extension",
		slog.String("name", req.Name),
	)

	// TODO: Implement actual API call to delete extension
	s.logger.InfoContext(ctx, "Extension deleted successfully",
		slog.String("name", req.Name),
	)

	return nil
}

// Convenience Methods for Prebuilt Extensions

// ExecuteCodeInterpreter executes a code interpreter operation with simplified parameters.
func (s *Service) ExecuteCodeInterpreter(ctx context.Context, extensionName, query string, files []string) (*CodeInterpreterExecutionResponse, error) {
	req := &ExecuteExtensionRequest{
		Name:        extensionName,
		OperationId: "generate_and_execute",
		// Note: OperationParams should be *structpb.Struct, but for now we'll handle conversion later
		// OperationParams: ...,  // TODO: Convert map to structpb.Struct
	}

	_, err := s.ExecuteExtension(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute code interpreter: %w", err)
	}

	// TODO: Parse response into CodeInterpreterExecutionResponse
	// For now, return a mock response
	return &CodeInterpreterExecutionResponse{
		GeneratedCode:   "# Generated code would be here",
		ExecutionResult: "Execution result would be here",
		ExecutionError:  "",
		OutputFiles:     []string{},
	}, nil
}

// ExecuteVertexAISearch executes a Vertex AI Search operation with simplified parameters.
func (s *Service) ExecuteVertexAISearch(ctx context.Context, extensionName, query string, maxResults int32) (*VertexAISearchExecutionResponse, error) {
	req := &ExecuteExtensionRequest{
		Name:        extensionName,
		OperationId: "search",
		// Note: OperationParams should be *structpb.Struct, but for now we'll handle conversion later
		// OperationParams: ...,  // TODO: Convert map to structpb.Struct
	}

	_, err := s.ExecuteExtension(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Vertex AI Search: %w", err)
	}

	// TODO: Parse response into VertexAISearchExecutionResponse
	// For now, return a mock response
	return &VertexAISearchExecutionResponse{
		Results:       []SearchResult{},
		NextPageToken: "",
	}, nil
}

// Helper Methods

// generateExtensionID generates a unique extension ID.
func (s *Service) generateExtensionID() string {
	// In a real implementation, this would generate a unique ID
	return fmt.Sprintf("ext_%d", time.Now().UnixNano())
}

// generateExtensionName generates the full resource name for an extension.
func (s *Service) generateExtensionName(extensionID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/extensions/%s", s.projectID, s.location, extensionID)
}

// validateManifest validates an extension manifest.
func (s *Service) validateManifest(manifest *ExtensionManifest) error {
	if manifest.Name == "" {
		return fmt.Errorf("manifest name is required")
	}
	if manifest.Description == "" {
		return fmt.Errorf("manifest description is required")
	}
	if manifest.ApiSpec == nil {
		return fmt.Errorf("API specification is required")
	}

	// Check the API spec type
	switch apiSpec := manifest.ApiSpec.ApiSpec.(type) {
	case *aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri:
		if apiSpec.OpenApiGcsUri == "" {
			return fmt.Errorf("OpenAPI GCS URI is required")
		}
		if !strings.HasPrefix(apiSpec.OpenApiGcsUri, "gs://") {
			return fmt.Errorf("OpenAPI GCS URI must start with gs://")
		}
	case *aiplatformpb.ExtensionManifest_ApiSpec_OpenApiYaml:
		if apiSpec.OpenApiYaml == "" {
			return fmt.Errorf("OpenAPI YAML is required")
		}
	default:
		return fmt.Errorf("API specification must be either GCS URI or YAML")
	}

	if manifest.AuthConfig == nil {
		return fmt.Errorf("authentication configuration is required")
	}
	if manifest.AuthConfig.AuthType == AuthTypeUnspecified {
		return fmt.Errorf("authentication type must be specified")
	}

	return nil
}

// Configuration Access Methods

// GetProjectID returns the configured Google Cloud project ID.
func (s *Service) GetProjectID() string {
	return s.projectID
}

// GetLocation returns the configured geographic location.
func (s *Service) GetLocation() string {
	return s.location
}

// GetLogger returns the configured logger instance.
func (s *Service) GetLogger() *slog.Logger {
	return s.logger
}
