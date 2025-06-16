// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package extension

import (
	"context"
	"fmt"
)

// getPrebuiltExtensionConfig returns the manifest and runtime configuration
// for a prebuilt extension type.
func (s *Service) getPrebuiltExtensionConfig(extensionType PrebuiltExtensionType) (*ExtensionManifest, *RuntimeConfig, error) {
	switch extensionType {
	case PrebuiltExtensionCodeInterpreter:
		return s.getCodeInterpreterConfig()
	case PrebuiltExtensionVertexAISearch:
		return s.getVertexAISearchConfig()
	default:
		return nil, nil, &PrebuiltExtensionError{
			ExtensionType: extensionType,
			Message:       "unknown prebuilt extension type",
		}
	}
}

// getCodeInterpreterConfig returns the configuration for the code interpreter extension.
func (s *Service) getCodeInterpreterConfig() (*ExtensionManifest, *RuntimeConfig, error) {
	manifest := NewExtensionManifest(
		"code_interpreter_tool",
		"Google Code Interpreter Extension",
		"gs://vertex-extension-public/code_interpreter.yaml",
		NewGoogleServiceAccountConfig(""), // Empty string uses default service account
	)

	runtimeConfig := NewCodeInterpreterRuntimeConfig(
		"", // No specific input bucket
		"", // No specific output bucket
	)

	return manifest, runtimeConfig, nil
}

// getVertexAISearchConfig returns the configuration for the Vertex AI Search extension.
func (s *Service) getVertexAISearchConfig() (*ExtensionManifest, *RuntimeConfig, error) {
	manifest := NewExtensionManifest(
		"vertex_ai_search",
		"Google Vertex AI Search Extension",
		"gs://vertex-extension-public/vertex_ai_search.yaml",
		NewGoogleServiceAccountConfig(""), // Empty string uses default service account
	)

	// Note: RuntimeConfig for Vertex AI Search requires serving_config_name
	// which must be provided when creating the extension
	runtimeConfig := NewVertexAISearchRuntimeConfig(
		"", // Serving config name must be set later
		"", // Engine ID not specified
	)

	return manifest, runtimeConfig, nil
}

// getPrebuiltDisplayName returns the display name for a prebuilt extension.
func (s *Service) getPrebuiltDisplayName(extensionType PrebuiltExtensionType) string {
	switch extensionType {
	case PrebuiltExtensionCodeInterpreter:
		return "Code Interpreter"
	case PrebuiltExtensionVertexAISearch:
		return "Vertex AI Search"
	default:
		return string(extensionType)
	}
}

// getPrebuiltDescription returns the description for a prebuilt extension.
func (s *Service) getPrebuiltDescription(extensionType PrebuiltExtensionType) string {
	switch extensionType {
	case PrebuiltExtensionCodeInterpreter:
		return "This extension generates and executes code in the specified language"
	case PrebuiltExtensionVertexAISearch:
		return "This extension searches from provided datastore"
	default:
		return fmt.Sprintf("Prebuilt extension: %s", extensionType)
	}
}

// CreateCodeInterpreterExtension creates a code interpreter extension with default configuration.
func (s *Service) CreateCodeInterpreterExtension(ctx context.Context) (*Extension, error) {
	return s.CreateFromHub(ctx, PrebuiltExtensionCodeInterpreter)
}

// CreateVertexAISearchExtension creates a Vertex AI Search extension with the specified serving config.
func (s *Service) CreateVertexAISearchExtension(ctx context.Context, servingConfigName string) (*Extension, error) {
	if servingConfigName == "" {
		return nil, &PrebuiltExtensionError{
			ExtensionType: PrebuiltExtensionVertexAISearch,
			Message:       "serving_config_name is required for Vertex AI Search extension",
		}
	}

	manifest, runtimeConfig, err := s.getVertexAISearchConfig()
	if err != nil {
		return nil, err
	}

	// Set the serving config name
	if vertexSearchConfig := runtimeConfig.GetVertexAiSearchRuntimeConfig(); vertexSearchConfig != nil {
		vertexSearchConfig.ServingConfigName = servingConfigName
	}

	req := &CreateExtensionRequest{
		DisplayName:   s.getPrebuiltDisplayName(PrebuiltExtensionVertexAISearch),
		Description:   s.getPrebuiltDescription(PrebuiltExtensionVertexAISearch),
		Manifest:      manifest,
		RuntimeConfig: runtimeConfig,
	}

	return s.CreateExtension(ctx, req)
}

// GetSupportedPrebuiltExtensions returns a list of supported prebuilt extension types.
func (s *Service) GetSupportedPrebuiltExtensions() []PrebuiltExtensionType {
	return []PrebuiltExtensionType{
		PrebuiltExtensionCodeInterpreter,
		PrebuiltExtensionVertexAISearch,
	}
}

// ValidatePrebuiltExtensionType validates that the extension type is supported.
func (s *Service) ValidatePrebuiltExtensionType(extensionType PrebuiltExtensionType) error {
	supported := s.GetSupportedPrebuiltExtensions()
	for _, supportedType := range supported {
		if extensionType == supportedType {
			return nil
		}
	}

	return &PrebuiltExtensionError{
		ExtensionType: extensionType,
		Message:       "unsupported prebuilt extension type",
	}
}
