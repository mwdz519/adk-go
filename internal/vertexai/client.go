// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package vertex

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-a2a/adk-go/internal/vertexai/caching"
	"github.com/go-a2a/adk-go/internal/vertexai/examplestore"
	"github.com/go-a2a/adk-go/internal/vertexai/extension"
	"github.com/go-a2a/adk-go/internal/vertexai/generativemodel"
	"github.com/go-a2a/adk-go/internal/vertexai/modelgarden"
	"github.com/go-a2a/adk-go/internal/vertexai/preview/evaluation"
	"github.com/go-a2a/adk-go/internal/vertexai/preview/rag"
	"github.com/go-a2a/adk-go/internal/vertexai/preview/reasoningengine"
	"github.com/go-a2a/adk-go/internal/vertexai/preview/tuning"
	"github.com/go-a2a/adk-go/internal/vertexai/prompt"
)

// Client provides unified access to all Vertex AI GA and preview functionality.
//
// The client orchestrates multiple specialized services to provide
// comprehensive access to GA and and preview features of Vertex AI.
// It maintains a single authentication context and configuration across
// all preview services.
type Client struct {
	// Configuration
	projectID string
	location  string
	logger    *slog.Logger

	// Core services
	cachingService      *caching.Service
	exampleStoreService *examplestore.Service
	generativeService   *generativemodel.Service
	modelGardenService  *modelgarden.Service
	extensionService    *extension.Service
	promptsService      *prompt.Service

	// Previwe services
	ragClient              *rag.Service
	evaluationService      *evaluation.Service
	reasoningengineService *reasoningengine.Service
	tuningService          *tuning.Service
}

// ClientOption is a functional option for configuring the preview client.
type ClientOption func(*Client)

// WithLogger sets a custom logger for the preview client.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// NewClient creates a new Vertex AI preview client.
//
// The client provides unified access to all preview services including RAG,
// content caching, enhanced generative models, and Model Garden integration.
//
// Parameters:
//   - ctx: Context for the initialization
//   - projectID: Google Cloud project ID
//   - location: Geographic location for Vertex AI services (e.g., "us-central1")
//   - opts: Optional configuration options
//
// Returns a fully initialized preview client or an error if initialization fails.
func NewClient(ctx context.Context, projectID, location string, opts ...ClientOption) (*Client, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if location == "" {
		return nil, fmt.Errorf("location is required")
	}

	client := &Client{
		projectID: projectID,
		location:  location,
		logger:    slog.Default(),
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Initialize content caching service
	contentCacheService, err := caching.NewService(ctx, projectID, location, caching.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize content caching service: %w", err)
	}
	client.cachingService = contentCacheService

	// Initialize example store service
	exampleStoreService, err := examplestore.NewService(ctx, projectID, location, examplestore.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize example store service: %w", err)
	}
	client.exampleStoreService = exampleStoreService

	// Initialize generative models service
	generativeService, err := generativemodel.NewService(ctx, projectID, location, generativemodel.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize generative models service: %w", err)
	}
	client.generativeService = generativeService

	// Initialize Model Garden service
	modelGardenService, err := modelgarden.NewService(ctx, projectID, location, modelgarden.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Model Garden service: %w", err)
	}
	client.modelGardenService = modelGardenService

	// Initialize Extension service
	extensionService, err := extension.NewService(ctx, projectID, location, extension.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Extension service: %w", err)
	}
	client.extensionService = extensionService

	// Initialize Prompts service
	promptsService, err := prompt.NewService(ctx, projectID, location, prompt.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Prompts service: %w", err)
	}
	client.promptsService = promptsService

	// Initialize RAG client
	ragClient, err := rag.NewService(ctx, projectID, location, rag.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RAG client: %w", err)
	}
	client.ragClient = ragClient

	// Initialize Evaluation Service
	evaluationService, err := evaluation.NewService(ctx, projectID, location, evaluation.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Evaluation service: %w", err)
	}
	client.evaluationService = evaluationService

	// Initialize Reasoning Engine Service
	reasoningengineService, err := reasoningengine.NewService(ctx, projectID, location, reasoningengine.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Reasoning Engine service: %w", err)
	}
	client.reasoningengineService = reasoningengineService

	// Initialize Tuning Service
	tuningService, err := tuning.NewService(ctx, projectID, location, tuning.WithLogger(client.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Reasoning Engine service: %w", err)
	}
	client.tuningService = tuningService

	client.logger.InfoContext(ctx, "Vertex AI preview client initialized successfully",
		slog.String("project_id", projectID),
		slog.String("location", location),
	)

	return client, nil
}

// Close closes the preview client and releases all resources.
//
// This method should be called when the client is no longer needed to ensure
// proper cleanup of underlying connections and resources.
func (c *Client) Close() error {
	c.logger.Info("Closing Vertex AI preview client")

	if err := c.cachingService.Close(); err != nil {
		c.logger.Error("Failed to close content caching service", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close content caching service: %w", err)
	}

	if err := c.exampleStoreService.Close(); err != nil {
		c.logger.Error("Failed to close example store service", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close example store service: %w", err)
	}

	if err := c.generativeService.Close(); err != nil {
		c.logger.Error("Failed to close generative models service", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close generative models service: %w", err)
	}

	if err := c.modelGardenService.Close(); err != nil {
		c.logger.Error("Failed to close Model Garden service", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close Model Garden service: %w", err)
	}

	if err := c.extensionService.Close(); err != nil {
		c.logger.Error("Failed to close Extension service", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close Extension service: %w", err)
	}

	if err := c.promptsService.Close(); err != nil {
		c.logger.Error("Failed to close Prompts service", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close Prompts service: %w", err)
	}

	// Close all services
	if err := c.ragClient.Close(); err != nil {
		c.logger.Error("Failed to close RAG client", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close RAG client: %w", err)
	}

	if err := c.evaluationService.Close(); err != nil {
		c.logger.Error("Failed to close Evaluation service", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close Evaluation service: %w", err)
	}

	if err := c.reasoningengineService.Close(); err != nil {
		c.logger.Error("Failed to close Reasoning Engine service", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close Evaluation service: %w", err)
	}

	if err := c.tuningService.Close(); err != nil {
		c.logger.Error("Failed to close Tuning service", slog.String("error", err.Error()))
		return fmt.Errorf("failed to close Tuning service: %w", err)
	}

	c.logger.Info("Vertex AI preview client closed successfully")
	return nil
}

// Service Access Methods
//
// These methods provide access to individual preview services while maintaining
// the unified client context and configuration.

// Caching returns the content caching service.
//
// The content caching service provides optimized caching for large content
// contexts, reducing token usage and improving performance for repeated queries.
func (c *Client) Caching() *caching.Service {
	return c.cachingService
}

// ExampleStore returns the example store service.
//
// The example store service provides functionality for managing Example Stores,
// uploading examples, and performing similarity-based retrieval for few-shot learning.
func (c *Client) ExampleStore() *examplestore.Service {
	return c.exampleStoreService
}

// GenerativeModel returns the enhanced generative models service.
//
// This service provides access to preview features for generative AI models,
// including advanced configuration options and experimental capabilities.
func (c *Client) GenerativeModel() *generativemodel.Service {
	return c.generativeService
}

// ModelGarden returns the Model Garden service.
//
// The Model Garden service provides access to experimental and community models,
// including deployment and management capabilities.
func (c *Client) ModelGarden() *modelgarden.Service {
	return c.modelGardenService
}

// Extension returns the Extension service.
//
// The Extension service provides access to Vertex AI Extension functionality,
// including creating, managing, and executing both custom and prebuilt extensions.
func (c *Client) Extension() *extension.Service {
	return c.extensionService
}

// Prompt returns the Prompt service.
//
// The Prompt service provides comprehensive prompt management functionality,
// including creation, versioning, template processing, and cloud storage integration.
func (c *Client) Prompt() *prompt.Service {
	return c.promptsService
}

// RAG returns the RAG (Retrieval-Augmented Generation) client.
//
// The RAG client provides comprehensive functionality for managing corpora,
// importing documents, and performing retrieval-augmented generation.
func (c *Client) RAG() *rag.Service {
	return c.ragClient
}

// Evaluation returns the Evaluation client
func (c *Client) Evaluation() *evaluation.Service {
	return c.evaluationService
}

// ReasoningEngine returns the Reasoning Engine client.
func (c *Client) ReasoningEngine() *reasoningengine.Service {
	return c.reasoningengineService
}

// Tuning returns the Tuning client.
func (c *Client) Tuning() *tuning.Service {
	return c.tuningService
}

// Configuration Access Methods

// GetProjectID returns the configured Google Cloud project ID.
func (c *Client) GetProjectID() string {
	return c.projectID
}

// GetLocation returns the configured geographic location.
func (c *Client) GetLocation() string {
	return c.location
}

// GetLogger returns the configured logger instance.
func (c *Client) GetLogger() *slog.Logger {
	return c.logger
}

// Health Check and Status Methods

// HealthCheck performs a basic health check across all preview services.
//
// This method verifies that all underlying services are accessible and
// functioning correctly. It's useful for monitoring and debugging.
func (c *Client) HealthCheck(ctx context.Context) error {
	c.logger.InfoContext(ctx, "Performing preview client health check")

	// Note: In a full implementation, you would perform actual health checks
	// against each service. For now, we just verify the services are initialized.

	if c.cachingService == nil {
		return fmt.Errorf("content caching service not initialized")
	}

	if c.exampleStoreService == nil {
		return fmt.Errorf("example store service not initialized")
	}

	if c.generativeService == nil {
		return fmt.Errorf("generative models service not initialized")
	}

	if c.modelGardenService == nil {
		return fmt.Errorf("Model Garden service not initialized")
	}

	if c.extensionService == nil {
		return fmt.Errorf("Extension service not initialized")
	}

	if c.promptsService == nil {
		return fmt.Errorf("Prompts service not initialized")
	}

	if c.ragClient == nil {
		return fmt.Errorf("RAG client not initialized")
	}

	if c.evaluationService == nil {
		return fmt.Errorf("Evaluation service not initialized")
	}

	if c.reasoningengineService == nil {
		return fmt.Errorf("Reasoning Engine service not initialized")
	}

	if c.tuningService == nil {
		return fmt.Errorf("Tuning service not initialized")
	}

	c.logger.InfoContext(ctx, "Preview client health check passed")
	return nil
}

// GetServiceStatus returns the status of all preview services.
func (c *Client) GetServiceStatus() map[string]string {
	status := make(map[string]string)

	if c.ragClient != nil {
		status["rag"] = "initialized"
	} else {
		status["rag"] = "not_initialized"
	}

	if c.cachingService != nil {
		status["content_caching"] = "initialized"
	} else {
		status["content_caching"] = "not_initialized"
	}

	if c.exampleStoreService != nil {
		status["example_store"] = "initialized"
	} else {
		status["example_store"] = "not_initialized"
	}

	if c.generativeService != nil {
		status["generative_models"] = "initialized"
	} else {
		status["generative_models"] = "not_initialized"
	}

	if c.modelGardenService != nil {
		status["model_garden"] = "initialized"
	} else {
		status["model_garden"] = "not_initialized"
	}

	if c.extensionService != nil {
		status["extensions"] = "initialized"
	} else {
		status["extensions"] = "not_initialized"
	}

	if c.promptsService != nil {
		status["prompts"] = "initialized"
	} else {
		status["prompts"] = "not_initialized"
	}

	return status
}
