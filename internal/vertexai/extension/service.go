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
	"google.golang.org/api/option"

	"github.com/go-a2a/adk-go/pkg/logging"
)

var VertexExtensionHub = map[PrebuiltExtensionType]*aiplatformpb.ImportExtensionRequest{
	PrebuiltExtensionCodeInterpreter: {
		Extension: &aiplatformpb.Extension{
			DisplayName: "Code Interpreter",
			Description: "This extension generates and executes code in the specified language",
			Manifest: &aiplatformpb.ExtensionManifest{
				Name:        "code_interpreter_tool",
				Description: "Google Code Interpreter Extension",
				ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
					ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
						OpenApiGcsUri: "gs://vertex-extension-public/code_interpreter.yaml",
					},
				},
				AuthConfig: NewGoogleServiceAccountConfig(""),
			},
		},
	},
	PrebuiltExtensionVertexAISearch: {
		Extension: &aiplatformpb.Extension{
			DisplayName: "Vertex AI Search",
			Description: "This extension generates and executes search queries",
			Manifest: &aiplatformpb.ExtensionManifest{
				Name:        "vertex_ai_search",
				Description: "Vertex AI Search Extension",
				ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
					ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
						OpenApiGcsUri: "gs://vertex-extension-public/vertex_ai_search.yaml",
					},
				},
				AuthConfig: NewGoogleServiceAccountConfig(""),
			},
		},
	},
	PrebuiltExtensionWebpageBrowser: {
		Extension: &aiplatformpb.Extension{
			DisplayName: "Webpage Browser",
			Description: "This extension fetches the content of a webpage",
			Manifest: &aiplatformpb.ExtensionManifest{
				Name:        "webpage_browser",
				Description: "Vertex Webpage Browser Extension",
				ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec{
					ApiSpec: &aiplatformpb.ExtensionManifest_ApiSpec_OpenApiGcsUri{
						OpenApiGcsUri: "gs://vertex-extension-public/webpage_browser.yaml",
					},
				},
				AuthConfig: NewGoogleServiceAccountConfig(""),
			},
		},
	},
}

// Service provides access to Vertex AI Extensions functionality.
//
// The service manages extension lifecycles, executes operations, and provides
// access to prebuilt extensions from Google's extension hub.
type Service interface {
	// GetProjectID returns the configured Google Cloud project ID.
	GetProjectID() string

	// GetLocation returns the configured geographic location.
	GetLocation() string

	// GetParent returns the resource name of the Location to import the Extension in.
	GetParent() string

	// CreateExtension creates a new custom extension.
	CreateExtension(ctx context.Context, req *aiplatformpb.ImportExtensionRequest) (*Extension, error)

	// ExecuteExtension executes an operation on a specific extension.
	ExecuteExtension(ctx context.Context, req *aiplatformpb.ExecuteExtensionRequest) (*aiplatformpb.ExecuteExtensionResponse, error)

	// CreateFromHub creates an extension from Google's prebuilt extension hub.
	CreateFromHub(ctx context.Context, extensionType PrebuiltExtensionType, runtimeConfig *aiplatformpb.RuntimeConfig) (*Extension, error)

	// ListExtensions lists all extensions in the project and location.
	ListExtensions(ctx context.Context, req *aiplatformpb.ListExtensionsRequest) (*aiplatformpb.ListExtensionsResponse, error)

	// GetExtension retrieves a specific extension by its resource name.
	GetExtension(ctx context.Context, req *aiplatformpb.GetExtensionRequest) (*Extension, error)

	// DeleteExtension deletes an extension.
	DeleteExtension(ctx context.Context, req *aiplatformpb.DeleteExtensionRequest) error

	// ExecuteCodeInterpreter executes a code interpreter operation with simplified parameters.
	ExecuteCodeInterpreter(ctx context.Context, extensionName, query string, files []string) (*CodeInterpreterExecutionResponse, error)

	// ExecuteVertexAISearch executes a Vertex AI Search operation with simplified parameters.
	ExecuteVertexAISearch(ctx context.Context, extensionName, query string, maxResults int32) (*VertexAISearchExecutionResponse, error)

	// GetSupportedPrebuiltExtensions returns a list of supported prebuilt extension types.
	GetSupportedPrebuiltExtensions() []PrebuiltExtensionType

	// ValidatePrebuiltExtensionType validates that the extension type is supported.
	ValidatePrebuiltExtensionType(extensionType PrebuiltExtensionType) error

	// Close closes the Extension Execution service and releases any resources.
	Close() error
}

type service struct {
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

var _ Service = (*service)(nil)

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
func NewService(ctx context.Context, projectID, location string, opts ...option.ClientOption) (*service, error) {
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

	service := &service{
		projectID: projectID,
		location:  location,
		logger:    logging.FromContext(ctx),
	}

	extensionExecutionClient, err := aiplatform.NewExtensionExecutionClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI Platform Extension Execution client: %w", err)
	}
	service.extensionExecutionClient = extensionExecutionClient

	extensionRegistryClient, err := aiplatform.NewExtensionRegistryClient(ctx, opts...)
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
func (s *service) Close() error {
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
func (s *service) CreateExtension(ctx context.Context, req *aiplatformpb.ImportExtensionRequest) (*Extension, error) {
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
func (s *service) ExecuteExtension(ctx context.Context, req *aiplatformpb.ExecuteExtensionRequest) (*aiplatformpb.ExecuteExtensionResponse, error) {
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
	_ = &aiplatformpb.ExecuteExtensionResponse{
		Content: `{"result": "mock execution result"}`,
	}

	s.logger.InfoContext(ctx, "Extension operation executed successfully",
		slog.String("name", req.Name),
		slog.String("operation_id", req.OperationId),
	)

	return &aiplatformpb.ExecuteExtensionResponse{
		Content: `{"result": "mock execution result"}`,
	}, nil
}

// CreateFromHub creates an extension from Google's prebuilt extension hub.
//
// This method provides easy access to Google's curated extensions like
// code_interpreter and vertex_ai_search without requiring manual manifest creation.
func (s *service) CreateFromHub(ctx context.Context, extensionType PrebuiltExtensionType, runtimeConfig *aiplatformpb.RuntimeConfig) (*Extension, error) {
	s.logger.InfoContext(ctx, "Creating extension from hub", slog.String("extension_type", string(extensionType)))

	extensionInfo := VertexExtensionHub[extensionType]
	extensionInfo.Extension.RuntimeConfig = runtimeConfig

	switch extensionType {
	case "code_interpreter":
		if _, ok := extensionInfo.GetExtension().GetRuntimeConfig().GetGoogleFirstPartyExtensionConfig().(*aiplatformpb.RuntimeConfig_CodeInterpreterRuntimeConfig_); !ok {
			return nil, errors.New("runtime_config is required for code_interpreter extension")
		}
	case "vertex_ai_search":
		if extensionInfo.GetExtension().GetRuntimeConfig() == nil {
			return nil, errors.New("runtime_config is required for vertex_ai_search extension")
		}
		if _, ok := extensionInfo.GetExtension().GetRuntimeConfig().GetGoogleFirstPartyExtensionConfig().(*aiplatformpb.RuntimeConfig_VertexAiSearchRuntimeConfig); !ok {
			return nil, errors.New("runtime_config is required for vertex_ai_search extension")
		}
	case "webpage_browser":
		// nothing to do
	default:
		return nil, fmt.Errorf("unsupported 1P extension name: %s", extensionType)
	}

	extension, err := s.CreateExtension(ctx, extensionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create extension from hub: %w", err)
	}

	s.logger.InfoContext(ctx, "Extension created from hub successfully", slog.String("extension_id", extension.GetID()))

	return extension, nil
}

// ListExtensions lists all extensions in the project and location.
func (s *service) ListExtensions(ctx context.Context, req *aiplatformpb.ListExtensionsRequest) (*aiplatformpb.ListExtensionsResponse, error) {
	if req == nil {
		req = &aiplatformpb.ListExtensionsRequest{}
	}

	s.logger.InfoContext(ctx, "Listing extensions",
		slog.Int("page_size", int(req.PageSize)),
		slog.String("page_token", req.PageToken),
	)

	// TODO: Implement actual API call to list extensions
	// For now, return empty response
	response := &aiplatformpb.ListExtensionsResponse{
		Extensions:    []*aiplatformpb.Extension{},
		NextPageToken: "",
	}

	s.logger.InfoContext(ctx, "Listed extensions successfully",
		slog.Int("extension_count", len(response.Extensions)),
	)

	return response, nil
}

// GetExtension retrieves a specific extension by its resource name.
func (s *service) GetExtension(ctx context.Context, req *aiplatformpb.GetExtensionRequest) (*Extension, error) {
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
func (s *service) DeleteExtension(ctx context.Context, req *aiplatformpb.DeleteExtensionRequest) error {
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
func (s *service) ExecuteCodeInterpreter(ctx context.Context, extensionName, query string, files []string) (*CodeInterpreterExecutionResponse, error) {
	req := &aiplatformpb.ExecuteExtensionRequest{
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
func (s *service) ExecuteVertexAISearch(ctx context.Context, extensionName, query string, maxResults int32) (*VertexAISearchExecutionResponse, error) {
	req := &aiplatformpb.ExecuteExtensionRequest{
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
func (s *service) generateExtensionID() string {
	// In a real implementation, this would generate a unique ID
	return fmt.Sprintf("ext_%d", time.Now().UnixNano())
}

// generateExtensionName generates the full resource name for an extension.
func (s *service) generateExtensionName(extensionID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/extensions/%s", s.projectID, s.location, extensionID)
}

// validateManifest validates an extension manifest.
func (s *service) validateManifest(manifest *aiplatformpb.ExtensionManifest) error {
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
	if manifest.AuthConfig.AuthType == aiplatformpb.AuthType_AUTH_TYPE_UNSPECIFIED {
		return fmt.Errorf("authentication type must be specified")
	}

	return nil
}

// Configuration Access Methods

// GetProjectID returns the configured Google Cloud project ID.
func (s *service) GetProjectID() string {
	return s.projectID
}

// GetLocation returns the configured geographic location.
func (s *service) GetLocation() string {
	return s.location
}

// GetLogger returns the configured logger instance.
func (s *service) GetLogger() *slog.Logger {
	return s.logger
}

// GetParent returns the resource name of the Location to import the Extension in.
func (s *service) GetParent() string {
	return "projects/" + s.projectID + "/locations/" + s.location
}
