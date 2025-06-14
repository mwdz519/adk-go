// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNewService(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		projectID string
		location  string
		wantErr   bool
	}{
		{
			name:      "valid_parameters",
			projectID: "test-project",
			location:  "us-central1",
			wantErr:   false,
		},
		{
			name:      "empty_project_id",
			projectID: "",
			location:  "us-central1",
			wantErr:   true,
		},
		{
			name:      "empty_location",
			projectID: "test-project",
			location:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewService(ctx, tt.projectID, tt.location)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewService() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewService() unexpected error: %v", err)
				return
			}

			if service == nil {
				t.Errorf("NewService() returned nil service")
				return
			}

			// Verify service properties
			if service.GetProjectID() != tt.projectID {
				t.Errorf("GetProjectID() = %v, want %v", service.GetProjectID(), tt.projectID)
			}

			if service.GetLocation() != tt.location {
				t.Errorf("GetLocation() = %v, want %v", service.GetLocation(), tt.location)
			}

			if !service.IsInitialized() {
				t.Errorf("IsInitialized() = false, want true")
			}

			// Test close
			if err := service.Close(); err != nil {
				t.Errorf("Close() unexpected error: %v", err)
			}

			if service.IsInitialized() {
				t.Errorf("IsInitialized() = true after close, want false")
			}
		})
	}
}

func TestServiceWithOptions(t *testing.T) {
	ctx := context.Background()

	// Test with custom cache expiry
	cacheExpiry := 10 * time.Minute
	service, err := NewService(ctx, "test-project", "us-central1", WithCacheExpiry(cacheExpiry))
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	if service.cacheExpiry != cacheExpiry {
		t.Errorf("cacheExpiry = %v, want %v", service.cacheExpiry, cacheExpiry)
	}
}

func TestCreatePrompt(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name    string
		req     *CreatePromptRequest
		wantErr bool
	}{
		{
			name: "valid_prompt",
			req: &CreatePromptRequest{
				Prompt: &Prompt{
					Name:        "test-prompt",
					DisplayName: "Test Prompt",
					Description: "A test prompt",
					Template:    "Hello {name}!",
					Variables:   []string{"name"},
				},
				ValidateTemplate: true,
			},
			wantErr: false,
		},
		{
			name:    "nil_request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "nil_prompt",
			req: &CreatePromptRequest{
				Prompt: nil,
			},
			wantErr: true,
		},
		{
			name: "empty_template",
			req: &CreatePromptRequest{
				Prompt: &Prompt{
					Name:     "test-prompt",
					Template: "",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_template",
			req: &CreatePromptRequest{
				Prompt: &Prompt{
					Name:     "test-prompt",
					Template: "Hello {unclosed",
				},
				ValidateTemplate: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := service.CreatePrompt(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreatePrompt() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreatePrompt() unexpected error: %v", err)
				return
			}

			if prompt == nil {
				t.Errorf("CreatePrompt() returned nil prompt")
				return
			}

			// Verify prompt properties
			if prompt.ID == "" {
				t.Errorf("CreatePrompt() prompt ID is empty")
			}

			if prompt.ProjectID != service.GetProjectID() {
				t.Errorf("prompt.ProjectID = %v, want %v", prompt.ProjectID, service.GetProjectID())
			}

			if prompt.Location != service.GetLocation() {
				t.Errorf("prompt.Location = %v, want %v", prompt.Location, service.GetLocation())
			}

			if prompt.CreatedAt.IsZero() {
				t.Errorf("prompt.CreatedAt is zero")
			}

			if prompt.UpdatedAt.IsZero() {
				t.Errorf("prompt.UpdatedAt is zero")
			}
		})
	}
}

func TestGetPrompt(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	// Create a test prompt first
	createReq := &CreatePromptRequest{
		Prompt: &Prompt{
			Name:        "test-prompt",
			DisplayName: "Test Prompt",
			Template:    "Hello {name}!",
			Variables:   []string{"name"},
		},
	}

	createdPrompt, err := service.CreatePrompt(ctx, createReq)
	if err != nil {
		t.Fatalf("CreatePrompt() unexpected error: %v", err)
	}

	tests := []struct {
		name    string
		req     *GetPromptRequest
		wantErr bool
	}{
		{
			name: "get_by_id",
			req: &GetPromptRequest{
				PromptID: createdPrompt.ID,
			},
			wantErr: false,
		},
		{
			name: "get_by_name",
			req: &GetPromptRequest{
				Name: createdPrompt.Name,
			},
			wantErr: false,
		},
		{
			name:    "empty_request",
			req:     &GetPromptRequest{},
			wantErr: true,
		},
		{
			name: "nonexistent_id",
			req: &GetPromptRequest{
				PromptID: "nonexistent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := service.GetPrompt(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetPrompt() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetPrompt() unexpected error: %v", err)
				return
			}

			if prompt == nil {
				t.Errorf("GetPrompt() returned nil prompt")
				return
			}

			// Verify we got the right prompt
			if prompt.ID != createdPrompt.ID {
				t.Errorf("prompt.ID = %v, want %v", prompt.ID, createdPrompt.ID)
			}
		})
	}
}

func TestUpdatePrompt(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	// Create a test prompt first
	createReq := &CreatePromptRequest{
		Prompt: &Prompt{
			Name:        "test-prompt",
			DisplayName: "Test Prompt",
			Template:    "Hello {name}!",
			Variables:   []string{"name"},
		},
	}

	createdPrompt, err := service.CreatePrompt(ctx, createReq)
	if err != nil {
		t.Fatalf("CreatePrompt() unexpected error: %v", err)
	}

	// Update the prompt
	updatedTemplate := "Hi {name}, how are you?"
	updateReq := &UpdatePromptRequest{
		Prompt: &Prompt{
			ID:        createdPrompt.ID,
			Name:      createdPrompt.Name,
			Template:  updatedTemplate,
			Variables: []string{"name"},
		},
		ValidateTemplate: true,
	}

	updatedPrompt, err := service.UpdatePrompt(ctx, updateReq)
	if err != nil {
		t.Errorf("UpdatePrompt() unexpected error: %v", err)
		return
	}

	if updatedPrompt.Template != updatedTemplate {
		t.Errorf("updatedPrompt.Template = %v, want %v", updatedPrompt.Template, updatedTemplate)
	}

	if !updatedPrompt.UpdatedAt.After(createdPrompt.UpdatedAt) {
		t.Errorf("updatedPrompt.UpdatedAt should be after createdPrompt.UpdatedAt")
	}
}

func TestDeletePrompt(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	// Create a test prompt first
	createReq := &CreatePromptRequest{
		Prompt: &Prompt{
			Name:        "test-prompt-delete",
			DisplayName: "Test Prompt for Deletion",
			Template:    "Hello {name}!",
			Variables:   []string{"name"},
		},
	}

	createdPrompt, err := service.CreatePrompt(ctx, createReq)
	if err != nil {
		t.Fatalf("CreatePrompt() unexpected error: %v", err)
	}

	// Delete the prompt
	deleteReq := &DeletePromptRequest{
		PromptID: createdPrompt.ID,
	}

	err = service.DeletePrompt(ctx, deleteReq)
	if err != nil {
		t.Errorf("DeletePrompt() unexpected error: %v", err)
		return
	}

	// Verify prompt is deleted (should not be in cache)
	cachedPrompt := service.getCachedPrompt(createdPrompt.ID)
	if cachedPrompt != nil {
		t.Errorf("Prompt still in cache after deletion")
	}
}

func TestListPrompts(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	// Test listing (will be empty since we're not actually persisting to cloud)
	req := &ListPromptsRequest{
		PageSize: 10,
	}

	response, err := service.ListPrompts(ctx, req)
	if err != nil {
		t.Errorf("ListPrompts() unexpected error: %v", err)
		return
	}

	if response == nil {
		t.Errorf("ListPrompts() returned nil response")
		return
	}

	// Since we're not actually persisting to cloud, we expect empty results
	if len(response.Prompts) > 0 {
		t.Errorf("ListPrompts() returned %d prompts, expected 0", len(response.Prompts))
	}
}

func TestCacheOperations(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	// Test cache operations
	testPrompt := &Prompt{
		ID:       "test-cache-id",
		Name:     "test-cache",
		Template: "Hello {name}!",
	}

	// Test caching
	service.cachePrompt(testPrompt)

	// Test retrieval from cache
	cachedPrompt := service.getCachedPrompt(testPrompt.ID)
	if cachedPrompt == nil {
		t.Errorf("getCachedPrompt() returned nil for cached prompt")
		return
	}

	if diff := cmp.Diff(testPrompt, cachedPrompt); diff != "" {
		t.Errorf("cached prompt mismatch (-want +got):\n%s", diff)
	}

	// Test cache removal
	service.removeCachedPrompt(testPrompt.ID)
	cachedPrompt = service.getCachedPrompt(testPrompt.ID)
	if cachedPrompt != nil {
		t.Errorf("getCachedPrompt() returned prompt after removal")
	}

	// Test cache stats
	stats := service.GetCacheStats()
	if stats == nil {
		t.Errorf("GetCacheStats() returned nil")
	}

	// Test cache clear
	service.cachePrompt(testPrompt)
	service.ClearCache()
	cachedPrompt = service.getCachedPrompt(testPrompt.ID)
	if cachedPrompt != nil {
		t.Errorf("getCachedPrompt() returned prompt after cache clear")
	}
}

func TestServiceOptions(t *testing.T) {
	ctx := context.Background()

	// Test with custom template engine
	customEngine := NewTemplateProcessorWithOptions(TemplateEngineAdvanced, ValidationModeStrict)
	service, err := NewService(ctx, "test-project", "us-central1",
		WithTemplateEngine(customEngine),
		WithCacheExpiry(5*time.Minute))
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	if service.templateEngine != customEngine {
		t.Errorf("templateEngine not set correctly")
	}

	if service.cacheExpiry != 5*time.Minute {
		t.Errorf("cacheExpiry = %v, want %v", service.cacheExpiry, 5*time.Minute)
	}
}

func TestPromptValidation(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name     string
		prompt   *Prompt
		wantErr  bool
		errorMsg string
	}{
		{
			name: "valid_prompt",
			prompt: &Prompt{
				Name:      "valid",
				Template:  "Hello {name}!",
				Variables: []string{"name"},
			},
			wantErr: false,
		},
		{
			name: "unmatched_braces",
			prompt: &Prompt{
				Name:     "invalid",
				Template: "Hello {name!",
			},
			wantErr: true,
		},
		{
			name: "empty_variable",
			prompt: &Prompt{
				Name:     "invalid",
				Template: "Hello {}!",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePromptTemplate(tt.prompt)

			if tt.wantErr && err == nil {
				t.Errorf("validatePromptTemplate() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("validatePromptTemplate() unexpected error: %v", err)
			}
		})
	}
}

func TestMetricsIntegration(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	// Create a prompt to trigger metrics
	createReq := &CreatePromptRequest{
		Prompt: &Prompt{
			Name:     "metrics-test",
			Template: "Hello {name}!",
		},
	}

	_, err = service.CreatePrompt(ctx, createReq)
	if err != nil {
		t.Fatalf("CreatePrompt() unexpected error: %v", err)
	}

	// Check metrics
	metrics := service.metrics.GetOperationMetrics()
	if metrics["prompts_created"] != 1 {
		t.Errorf("prompts_created metric = %d, want 1", metrics["prompts_created"])
	}

	// Test metrics reset
	service.metrics.Reset()
	metrics = service.metrics.GetOperationMetrics()
	if metrics["prompts_created"] != 0 {
		t.Errorf("prompts_created metric after reset = %d, want 0", metrics["prompts_created"])
	}
}
