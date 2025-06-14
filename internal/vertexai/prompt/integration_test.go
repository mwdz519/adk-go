// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package prompt

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// Integration tests require actual Vertex AI API credentials
// Run with: go test -tags=integration ./...

func TestIntegration_FullWorkflow(t *testing.T) {
	// Skip if no API key provided
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GCLOUD_PROJECT") == "" {
		t.Skip("Skipping integration test: no Google Cloud credentials found")
	}

	ctx := context.Background()
	projectID := os.Getenv("GCLOUD_PROJECT")
	if projectID == "" {
		projectID = "test-project" // fallback for local testing
	}

	service, err := NewService(ctx, projectID, "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	// Test complete workflow: Create -> Get -> Update -> Version -> Apply -> Delete
	t.Run("complete_workflow", func(t *testing.T) {
		// 1. Create a prompt
		createReq := &CreatePromptRequest{
			Prompt: &Prompt{
				Name:        "integration-test-prompt",
				DisplayName: "Integration Test Prompt",
				Description: "A prompt for integration testing",
				Template:    "Hello {name}, welcome to {company}!",
				Variables:   []string{"name", "company"},
				Category:    "test",
				Tags:        []string{"integration", "test"},
			},
			CreateVersion:    true,
			ValidateTemplate: true,
		}

		createdPrompt, err := service.CreatePrompt(ctx, createReq)
		if err != nil {
			t.Fatalf("CreatePrompt() unexpected error: %v", err)
		}

		if createdPrompt.ID == "" {
			t.Errorf("CreatePrompt() returned prompt with empty ID")
		}

		// 2. Get the prompt
		getReq := &GetPromptRequest{
			PromptID:        createdPrompt.ID,
			IncludeVersions: true,
		}

		retrievedPrompt, err := service.GetPrompt(ctx, getReq)
		if err != nil {
			t.Fatalf("GetPrompt() unexpected error: %v", err)
		}

		if retrievedPrompt.ID != createdPrompt.ID {
			t.Errorf("GetPrompt() ID = %v, want %v", retrievedPrompt.ID, createdPrompt.ID)
		}

		// 3. Apply template
		applyReq := &ApplyTemplateRequest{
			PromptID: createdPrompt.ID,
			Variables: map[string]any{
				"name":    "Alice",
				"company": "Tech Corp",
			},
			ValidateVariables: true,
		}

		applyResp, err := service.ApplyTemplate(ctx, applyReq)
		if err != nil {
			t.Fatalf("ApplyTemplate() unexpected error: %v", err)
		}

		expectedContent := "Hello Alice, welcome to Tech Corp!"
		if applyResp.Content != expectedContent {
			t.Errorf("ApplyTemplate() content = %v, want %v", applyResp.Content, expectedContent)
		}

		// 4. Update the prompt
		updateReq := &UpdatePromptRequest{
			Prompt: &Prompt{
				ID:          createdPrompt.ID,
				Name:        createdPrompt.Name,
				Template:    "Hi {name}, thanks for joining {company}!",
				Variables:   []string{"name", "company"},
				Description: "Updated description",
			},
			CreateNewVersion: true,
			VersionName:      "v2",
			Changelog:        "Updated greeting message",
			ValidateTemplate: true,
		}

		updatedPrompt, err := service.UpdatePrompt(ctx, updateReq)
		if err != nil {
			t.Fatalf("UpdatePrompt() unexpected error: %v", err)
		}

		if updatedPrompt.Template == createdPrompt.Template {
			t.Errorf("UpdatePrompt() template not updated")
		}

		// 5. List versions
		listVersionsReq := &ListVersionsRequest{
			PromptID: createdPrompt.ID,
		}

		versionsResp, err := service.ListVersions(ctx, listVersionsReq)
		if err != nil {
			t.Fatalf("ListVersions() unexpected error: %v", err)
		}

		if len(versionsResp.Versions) < 2 {
			t.Errorf("ListVersions() returned %d versions, expected at least 2", len(versionsResp.Versions))
		}

		// 6. Apply updated template
		applyReq.Variables = map[string]any{
			"name":    "Bob",
			"company": "Innovation Inc",
		}

		applyResp, err = service.ApplyTemplate(ctx, applyReq)
		if err != nil {
			t.Fatalf("ApplyTemplate() on updated prompt unexpected error: %v", err)
		}

		expectedUpdatedContent := "Hi Bob, thanks for joining Innovation Inc!"
		if applyResp.Content != expectedUpdatedContent {
			t.Errorf("ApplyTemplate() updated content = %v, want %v", applyResp.Content, expectedUpdatedContent)
		}

		// 7. Restore previous version
		if len(versionsResp.Versions) >= 2 {
			restoreReq := &RestoreVersionRequest{
				PromptID:       createdPrompt.ID,
				VersionID:      versionsResp.Versions[0].VersionID,
				NewVersionName: "restored-v1",
				Changelog:      "Restored original version",
			}

			_, err = service.RestoreVersion(ctx, restoreReq)
			if err != nil {
				t.Fatalf("RestoreVersion() unexpected error: %v", err)
			}
		}

		// 8. Delete the prompt
		deleteReq := &DeletePromptRequest{
			PromptID:       createdPrompt.ID,
			DeleteVersions: true,
		}

		err = service.DeletePrompt(ctx, deleteReq)
		if err != nil {
			t.Fatalf("DeletePrompt() unexpected error: %v", err)
		}

		// 9. Verify deletion
		_, err = service.GetPrompt(ctx, &GetPromptRequest{PromptID: createdPrompt.ID})
		if !IsNotFound(err) {
			t.Errorf("GetPrompt() after deletion should return NotFound error, got %v", err)
		}
	})
}

func TestIntegration_BatchOperations(t *testing.T) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GCLOUD_PROJECT") == "" {
		t.Skip("Skipping integration test: no Google Cloud credentials found")
	}

	ctx := context.Background()
	projectID := os.Getenv("GCLOUD_PROJECT")
	if projectID == "" {
		projectID = "test-project"
	}

	service, err := NewService(ctx, projectID, "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	t.Run("batch_create_and_apply", func(t *testing.T) {
		// Create multiple prompts
		prompts := []*Prompt{
			{
				Name:      "batch-test-1",
				Template:  "Hello {name}!",
				Variables: []string{"name"},
				Category:  "greeting",
				Tags:      []string{"batch", "test"},
			},
			{
				Name:      "batch-test-2",
				Template:  "Welcome {name} to {company}!",
				Variables: []string{"name", "company"},
				Category:  "greeting",
				Tags:      []string{"batch", "test"},
			},
			{
				Name:      "batch-test-3",
				Template:  "Thank you {name} for your order #{order_id}!",
				Variables: []string{"name", "order_id"},
				Category:  "notification",
				Tags:      []string{"batch", "test"},
			},
		}

		batchCreateReq := &BatchCreatePromptsRequest{
			Prompts:         prompts,
			CreateVersions:  true,
			ValidateAll:     true,
			ContinueOnError: true,
		}

		batchResp, err := service.BatchCreatePrompts(ctx, batchCreateReq)
		if err != nil {
			t.Fatalf("BatchCreatePrompts() unexpected error: %v", err)
		}

		if batchResp.Succeeded != int32(len(prompts)) {
			t.Errorf("BatchCreatePrompts() succeeded = %d, want %d", batchResp.Succeeded, len(prompts))
		}

		// Batch apply templates
		applyRequests := []*ApplyTemplateRequest{
			{
				PromptID: batchResp.Results[0].Prompt.ID,
				Variables: map[string]any{
					"name": "Alice",
				},
			},
			{
				PromptID: batchResp.Results[1].Prompt.ID,
				Variables: map[string]any{
					"name":    "Bob",
					"company": "Tech Corp",
				},
			},
			{
				PromptID: batchResp.Results[2].Prompt.ID,
				Variables: map[string]any{
					"name":     "Charlie",
					"order_id": "12345",
				},
			},
		}

		applyResults, err := service.BatchApplyTemplates(ctx, applyRequests)
		if err != nil {
			t.Fatalf("BatchApplyTemplates() unexpected error: %v", err)
		}

		expectedContents := []string{
			"Hello Alice!",
			"Welcome Bob to Tech Corp!",
			"Thank you Charlie for your order #12345!",
		}

		for i, result := range applyResults {
			if !result.Success {
				t.Errorf("BatchApplyTemplates() result[%d] failed: %s", i, result.Error)
				continue
			}

			if result.Response.Content != expectedContents[i] {
				t.Errorf("BatchApplyTemplates() result[%d] content = %v, want %v",
					i, result.Response.Content, expectedContents[i])
			}
		}

		// Clean up
		for _, result := range batchResp.Results {
			if result.Success {
				_ = service.DeletePrompt(ctx, &DeletePromptRequest{
					PromptID: result.Prompt.ID,
					Force:    true,
				})
			}
		}
	})
}

func TestIntegration_SearchAndFilter(t *testing.T) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GCLOUD_PROJECT") == "" {
		t.Skip("Skipping integration test: no Google Cloud credentials found")
	}

	ctx := context.Background()
	projectID := os.Getenv("GCLOUD_PROJECT")
	if projectID == "" {
		projectID = "test-project"
	}

	service, err := NewService(ctx, projectID, "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	t.Run("search_and_filter", func(t *testing.T) {
		// Create test prompts with different categories and tags
		testPrompts := []*Prompt{
			{
				Name:        "customer-greeting",
				DisplayName: "Customer Greeting",
				Template:    "Hello {customer_name}, welcome!",
				Category:    "customer-service",
				Tags:        []string{"greeting", "customer"},
			},
			{
				Name:        "support-ticket",
				DisplayName: "Support Ticket Response",
				Template:    "Thank you for contacting support, {customer_name}.",
				Category:    "customer-service",
				Tags:        []string{"support", "ticket"},
			},
			{
				Name:        "marketing-email",
				DisplayName: "Marketing Email",
				Template:    "Special offer for {customer_name}!",
				Category:    "marketing",
				Tags:        []string{"email", "promotion"},
			},
		}

		var createdPrompts []*Prompt
		for _, prompt := range testPrompts {
			createReq := &CreatePromptRequest{
				Prompt: prompt,
			}

			created, err := service.CreatePrompt(ctx, createReq)
			if err != nil {
				t.Fatalf("CreatePrompt() unexpected error: %v", err)
			}
			createdPrompts = append(createdPrompts, created)
		}

		// Test search by query
		searchReq := &SearchPromptsRequest{
			Query:    "customer",
			PageSize: 10,
		}

		searchResp, err := service.SearchPrompts(ctx, searchReq)
		if err != nil {
			t.Fatalf("SearchPrompts() unexpected error: %v", err)
		}

		// Should find prompts containing "customer" in name or template
		if len(searchResp.Results) < 2 {
			t.Errorf("SearchPrompts() found %d results, expected at least 2", len(searchResp.Results))
		}

		// Test filter by category
		listReq := &ListPromptsRequest{
			Category: "customer-service",
			PageSize: 10,
		}

		listResp, err := service.ListPrompts(ctx, listReq)
		if err != nil {
			t.Fatalf("ListPrompts() with category filter unexpected error: %v", err)
		}

		// Should find 2 customer-service prompts
		expectedCount := 2
		if len(listResp.Prompts) != expectedCount {
			t.Errorf("ListPrompts() with category filter found %d prompts, want %d",
				len(listResp.Prompts), expectedCount)
		}

		// Test filter by tags
		listReq.Category = ""
		listReq.Tags = []string{"greeting"}

		listResp, err = service.ListPrompts(ctx, listReq)
		if err != nil {
			t.Fatalf("ListPrompts() with tags filter unexpected error: %v", err)
		}

		// Should find 1 greeting prompt
		if len(listResp.Prompts) != 1 {
			t.Errorf("ListPrompts() with tags filter found %d prompts, want 1", len(listResp.Prompts))
		}

		// Clean up
		for _, prompt := range createdPrompts {
			_ = service.DeletePrompt(ctx, &DeletePromptRequest{
				PromptID: prompt.ID,
				Force:    true,
			})
		}
	})
}

func TestIntegration_Performance(t *testing.T) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GCLOUD_PROJECT") == "" {
		t.Skip("Skipping integration test: no Google Cloud credentials found")
	}

	ctx := context.Background()
	projectID := os.Getenv("GCLOUD_PROJECT")
	if projectID == "" {
		projectID = "test-project"
	}

	service, err := NewService(ctx, projectID, "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	t.Run("template_application_performance", func(t *testing.T) {
		// Create a prompt for performance testing
		createReq := &CreatePromptRequest{
			Prompt: &Prompt{
				Name:      "performance-test",
				Template:  "Hello {name}, welcome to {company}! Your user ID is {user_id} and role is {role}.",
				Variables: []string{"name", "company", "user_id", "role"},
			},
		}

		prompt, err := service.CreatePrompt(ctx, createReq)
		if err != nil {
			t.Fatalf("CreatePrompt() unexpected error: %v", err)
		}
		defer service.DeletePrompt(ctx, &DeletePromptRequest{PromptID: prompt.ID, Force: true})

		// Test template application performance
		variables := map[string]any{
			"name":    "Alice",
			"company": "Tech Corp",
			"user_id": "12345",
			"role":    "admin",
		}

		// Warm up
		for i := 0; i < 10; i++ {
			_, err := service.ApplyTemplateSimple(ctx, prompt.ID, variables)
			if err != nil {
				t.Fatalf("ApplyTemplateSimple() warmup error: %v", err)
			}
		}

		// Measure performance
		iterations := 100
		start := time.Now()

		for i := 0; i < iterations; i++ {
			_, err := service.ApplyTemplateSimple(ctx, prompt.ID, variables)
			if err != nil {
				t.Fatalf("ApplyTemplateSimple() performance test error: %v", err)
			}
		}

		duration := time.Since(start)
		avgLatency := duration / time.Duration(iterations)

		t.Logf("Performance test: %d iterations in %v (avg: %v per operation)",
			iterations, duration, avgLatency)

		// Performance should be reasonable (less than 100ms per operation on average)
		if avgLatency > 100*time.Millisecond {
			t.Errorf("Average latency %v exceeds expected threshold of 100ms", avgLatency)
		}

		// Test metrics
		metrics := service.metrics.GetAllMetrics()
		t.Logf("Service metrics: %+v", metrics)
	})
}

func TestIntegration_ErrorHandling(t *testing.T) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GCLOUD_PROJECT") == "" {
		t.Skip("Skipping integration test: no Google Cloud credentials found")
	}

	ctx := context.Background()
	projectID := os.Getenv("GCLOUD_PROJECT")
	if projectID == "" {
		projectID = "test-project"
	}

	service, err := NewService(ctx, projectID, "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	t.Run("error_handling", func(t *testing.T) {
		// Test getting non-existent prompt
		_, err := service.GetPrompt(ctx, &GetPromptRequest{
			PromptID: "non-existent-prompt",
		})
		if !IsNotFound(err) {
			t.Errorf("GetPrompt() for non-existent prompt should return NotFound error, got %v", err)
		}

		// Test creating prompt with invalid template
		invalidCreateReq := &CreatePromptRequest{
			Prompt: &Prompt{
				Name:     "invalid-template",
				Template: "Hello {unclosed_brace",
			},
			ValidateTemplate: true,
		}

		_, err = service.CreatePrompt(ctx, invalidCreateReq)
		if !IsInvalidTemplate(err) {
			t.Errorf("CreatePrompt() with invalid template should return InvalidTemplate error, got %v", err)
		}

		// Test applying template with missing variables in strict mode
		strictProcessor := NewTemplateProcessorWithOptions(TemplateEngineSimple, ValidationModeStrict)
		serviceWithStrict, err := NewService(ctx, projectID, "us-central1",
			WithTemplateEngine(strictProcessor))
		if err != nil {
			t.Fatalf("NewService() with strict processor unexpected error: %v", err)
		}
		defer serviceWithStrict.Close()

		applyReq := &ApplyTemplateRequest{
			Template:          "Hello {name}!",
			Variables:         map[string]any{}, // Missing variable
			ValidateVariables: true,
			StrictMode:        true,
		}

		_, err = serviceWithStrict.ApplyTemplate(ctx, applyReq)
		if !IsMissingVariables(err) {
			t.Errorf("ApplyTemplate() with missing variables in strict mode should return MissingVariables error, got %v", err)
		}
	})
}

func TestIntegration_Concurrency(t *testing.T) {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv("GCLOUD_PROJECT") == "" {
		t.Skip("Skipping integration test: no Google Cloud credentials found")
	}

	ctx := context.Background()
	projectID := os.Getenv("GCLOUD_PROJECT")
	if projectID == "" {
		projectID = "test-project"
	}

	service, err := NewService(ctx, projectID, "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	t.Run("concurrent_operations", func(t *testing.T) {
		// Create a base prompt
		createReq := &CreatePromptRequest{
			Prompt: &Prompt{
				Name:      "concurrency-test",
				Template:  "Hello {name}!",
				Variables: []string{"name"},
			},
		}

		prompt, err := service.CreatePrompt(ctx, createReq)
		if err != nil {
			t.Fatalf("CreatePrompt() unexpected error: %v", err)
		}
		defer service.DeletePrompt(ctx, &DeletePromptRequest{PromptID: prompt.ID, Force: true})

		// Run concurrent template applications
		const numGoroutines = 10
		const opsPerGoroutine = 10

		errChan := make(chan error, numGoroutines*opsPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				for j := 0; j < opsPerGoroutine; j++ {
					variables := map[string]any{
						"name": fmt.Sprintf("User_%d_%d", goroutineID, j),
					}

					_, err := service.ApplyTemplateSimple(ctx, prompt.ID, variables)
					if err != nil {
						errChan <- fmt.Errorf("goroutine %d, op %d: %w", goroutineID, j, err)
						return
					}
				}
			}(i)
		}

		// Wait for all operations to complete
		time.Sleep(5 * time.Second)
		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Errorf("Concurrent operation error: %v", err)
		}

		// Verify metrics
		metrics := service.metrics.GetOperationMetrics()
		expectedApplications := int64(numGoroutines * opsPerGoroutine)
		if metrics["templates_applied"] < expectedApplications {
			t.Errorf("Expected at least %d template applications, got %d",
				expectedApplications, metrics["templates_applied"])
		}
	})
}
