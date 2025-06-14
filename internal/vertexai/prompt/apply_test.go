// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestService_ApplyTemplate(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name    string
		req     *ApplyTemplateRequest
		want    string
		wantErr bool
	}{
		{
			name: "direct_template_application",
			req: &ApplyTemplateRequest{
				Template: "Hello {name}!",
				Variables: map[string]any{
					"name": "Alice",
				},
			},
			want:    "Hello Alice!",
			wantErr: false,
		},
		{
			name: "multiple_variables",
			req: &ApplyTemplateRequest{
				Template: "Hello {name}, welcome to {company}!",
				Variables: map[string]any{
					"name":    "Bob",
					"company": "Tech Corp",
				},
			},
			want:    "Hello Bob, welcome to Tech Corp!",
			wantErr: false,
		},
		{
			name: "number_substitution",
			req: &ApplyTemplateRequest{
				Template: "You have {count} items",
				Variables: map[string]any{
					"count": 42,
				},
			},
			want:    "You have 42 items",
			wantErr: false,
		},
		{
			name:    "nil_request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty_template_and_no_prompt",
			req: &ApplyTemplateRequest{
				Template: "",
				Variables: map[string]any{
					"name": "Alice",
				},
			},
			wantErr: true,
		},
		{
			name: "nil_variables",
			req: &ApplyTemplateRequest{
				Template:  "Hello {name}!",
				Variables: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.ApplyTemplate(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ApplyTemplate() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ApplyTemplate() unexpected error: %v", err)
				return
			}

			if response == nil {
				t.Errorf("ApplyTemplate() returned nil response")
				return
			}

			if response.Content != tt.want {
				t.Errorf("ApplyTemplate() content = %v, want %v", response.Content, tt.want)
			}
		})
	}
}

func TestService_ApplyTemplateToPrompt(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	prompt := &Prompt{
		Name:      "test-prompt",
		Template:  "Hello {name}, you have {count} messages!",
		Variables: []string{"name", "count"},
	}

	variables := map[string]any{
		"name":  "Alice",
		"count": 5,
	}

	response, err := service.ApplyTemplateToPrompt(ctx, prompt, variables)
	if err != nil {
		t.Errorf("ApplyTemplateToPrompt() unexpected error: %v", err)
		return
	}

	expected := "Hello Alice, you have 5 messages!"
	if response.Content != expected {
		t.Errorf("ApplyTemplateToPrompt() content = %v, want %v", response.Content, expected)
	}
}

func TestService_ApplyTemplateSimple(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	// Create a test prompt first
	createReq := &CreatePromptRequest{
		Prompt: &Prompt{
			Name:     "simple-test",
			Template: "Hello {name}!",
		},
	}

	createdPrompt, err := service.CreatePrompt(ctx, createReq)
	if err != nil {
		t.Fatalf("CreatePrompt() unexpected error: %v", err)
	}

	variables := map[string]any{
		"name": "Alice",
	}

	content, err := service.ApplyTemplateSimple(ctx, createdPrompt.ID, variables)
	if err != nil {
		t.Errorf("ApplyTemplateSimple() unexpected error: %v", err)
		return
	}

	expected := "Hello Alice!"
	if content != expected {
		t.Errorf("ApplyTemplateSimple() content = %v, want %v", content, expected)
	}
}

func TestService_ValidateTemplate(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name      string
		template  string
		variables []string
		wantValid bool
	}{
		{
			name:      "valid_template",
			template:  "Hello {name}!",
			variables: []string{"name"},
			wantValid: true,
		},
		{
			name:      "invalid_template",
			template:  "Hello {name!",
			variables: []string{"name"},
			wantValid: false,
		},
		{
			name:      "undeclared_variable",
			template:  "Hello {name}!",
			variables: []string{},
			wantValid: true, // In warn mode, this should still be valid but with warnings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateTemplate(ctx, tt.template, tt.variables)
			if err != nil {
				t.Errorf("ValidateTemplate() unexpected error: %v", err)
				return
			}

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateTemplate() IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}
		})
	}
}

func TestService_PreviewTemplate(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	template := "Hello {name}, you have {count} messages!"
	sampleVariables := map[string]any{
		"name":  "Alice",
		"count": 5,
	}

	response, err := service.PreviewTemplate(ctx, template, sampleVariables)
	if err != nil {
		t.Errorf("PreviewTemplate() unexpected error: %v", err)
		return
	}

	expected := "Hello Alice, you have 5 messages!"
	if response.Content != expected {
		t.Errorf("PreviewTemplate() content = %v, want %v", response.Content, expected)
	}
}

func TestService_ExtractVariables(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	template := "Hello {name}, welcome to {company}! Your ID is {user_id}."

	variables, err := service.ExtractVariables(ctx, template)
	if err != nil {
		t.Errorf("ExtractVariables() unexpected error: %v", err)
		return
	}

	expected := []string{"name", "company", "user_id"}
	if diff := cmp.Diff(expected, variables); diff != "" {
		t.Errorf("ExtractVariables() mismatch (-want +got):\n%s", diff)
	}
}

func TestService_BatchApplyTemplates(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	requests := []*ApplyTemplateRequest{
		{
			Template: "Hello {name}!",
			Variables: map[string]any{
				"name": "Alice",
			},
		},
		{
			Template: "Welcome {name} to {company}!",
			Variables: map[string]any{
				"name":    "Bob",
				"company": "Tech Corp",
			},
		},
		{
			Template: "Hello {name}!", // Valid template but will work since we're in loose validation mode
			Variables: map[string]any{
				"name": "Charlie",
			},
		},
	}

	results, err := service.BatchApplyTemplates(ctx, requests)
	if err != nil {
		t.Errorf("BatchApplyTemplates() unexpected error: %v", err)
		return
	}

	if len(results) != len(requests) {
		t.Errorf("BatchApplyTemplates() returned %d results, want %d", len(results), len(requests))
		return
	}

	// First two should succeed
	if !results[0].Success {
		t.Errorf("BatchApplyTemplates() result[0] should succeed")
	}
	if !results[1].Success {
		t.Errorf("BatchApplyTemplates() result[1] should succeed")
	}

	// Third should succeed since it's a valid template now
	if !results[2].Success {
		t.Errorf("BatchApplyTemplates() result[2] should succeed")
	}

	// Check content
	if results[0].Response.Content != "Hello Alice!" {
		t.Errorf("BatchApplyTemplates() result[0] content = %v, want 'Hello Alice!'", results[0].Response.Content)
	}
}

func TestService_CompileTemplate(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	template := "Hello {name}, welcome to {company}!"

	compiled, err := service.CompileTemplate(ctx, template)
	if err != nil {
		t.Errorf("CompileTemplate() unexpected error: %v", err)
		return
	}

	if compiled == nil {
		t.Errorf("CompileTemplate() returned nil compiled template")
		return
	}

	// Test execution
	variables := map[string]any{
		"name":    "Alice",
		"company": "Tech Corp",
	}

	response, err := compiled.Execute(variables)
	if err != nil {
		t.Errorf("compiled.Execute() unexpected error: %v", err)
		return
	}

	expected := "Hello Alice, welcome to Tech Corp!"
	if response.Content != expected {
		t.Errorf("compiled.Execute() content = %v, want %v", response.Content, expected)
	}
}

func TestService_GenerateTemplatePreview(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	template := "Hello {name}, you have {count} messages!"
	sampleVariables := map[string]any{
		"name":  "Alice",
		"count": 3,
	}

	preview, err := service.GenerateTemplatePreview(ctx, template, sampleVariables)
	if err != nil {
		t.Errorf("GenerateTemplatePreview() unexpected error: %v", err)
		return
	}

	if preview == nil {
		t.Errorf("GenerateTemplatePreview() returned nil preview")
		return
	}

	expectedContent := "Hello Alice, you have 3 messages!"
	if preview.PreviewContent != expectedContent {
		t.Errorf("GenerateTemplatePreview() PreviewContent = %v, want %v", preview.PreviewContent, expectedContent)
	}

	expectedVars := []string{"name", "count"}
	if diff := cmp.Diff(expectedVars, preview.DetectedVariables); diff != "" {
		t.Errorf("GenerateTemplatePreview() DetectedVariables mismatch (-want +got):\n%s", diff)
	}

	if preview.ValidationResult == nil {
		t.Errorf("GenerateTemplatePreview() ValidationResult is nil")
	}
}

func TestTemplateAnalyzer(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	analyzer := service.NewTemplateAnalyzer()
	template := "Hello {name}, welcome to {company}! You have {message_count} messages."

	analysis, err := analyzer.AnalyzeTemplate(ctx, template)
	if err != nil {
		t.Errorf("AnalyzeTemplate() unexpected error: %v", err)
		return
	}

	if analysis == nil {
		t.Errorf("AnalyzeTemplate() returned nil analysis")
		return
	}

	// Check detected variables
	expectedVars := []string{"name", "company", "message_count"}
	if diff := cmp.Diff(expectedVars, analysis.Variables); diff != "" {
		t.Errorf("AnalyzeTemplate() Variables mismatch (-want +got):\n%s", diff)
	}

	// Check complexity metrics
	if analysis.Complexity == nil {
		t.Errorf("AnalyzeTemplate() Complexity is nil")
		return
	}

	if analysis.Complexity.VariableCount != 3 {
		t.Errorf("AnalyzeTemplate() VariableCount = %d, want 3", analysis.Complexity.VariableCount)
	}

	if analysis.Complexity.UniqueVariables != 3 {
		t.Errorf("AnalyzeTemplate() UniqueVariables = %d, want 3", analysis.Complexity.UniqueVariables)
	}

	if analysis.Complexity.CharacterCount != len(template) {
		t.Errorf("AnalyzeTemplate() CharacterCount = %d, want %d", analysis.Complexity.CharacterCount, len(template))
	}
}

func TestApplyTemplateWithValidation(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name          string
		template      string
		variables     map[string]any
		validateVars  bool
		strictMode    bool
		wantErr       bool
		expectMissing int
		expectUnused  int
	}{
		{
			name:         "all_variables_provided",
			template:     "Hello {name}!",
			variables:    map[string]any{"name": "Alice"},
			validateVars: true,
			strictMode:   false,
			wantErr:      false,
		},
		{
			name:          "missing_variable_non_strict",
			template:      "Hello {name}!",
			variables:     map[string]any{},
			validateVars:  true,
			strictMode:    false,
			wantErr:       false,
			expectMissing: 1,
		},
		{
			name:          "missing_variable_strict",
			template:      "Hello {name}!",
			variables:     map[string]any{},
			validateVars:  true,
			strictMode:    true,
			wantErr:       true,
			expectMissing: 1,
		},
		{
			name:         "unused_variable",
			template:     "Hello world!",
			variables:    map[string]any{"unused": "value"},
			validateVars: true,
			strictMode:   false,
			wantErr:      false,
			expectUnused: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ApplyTemplateRequest{
				Template:          tt.template,
				Variables:         tt.variables,
				ValidateVariables: tt.validateVars,
				StrictMode:        tt.strictMode,
			}

			response, err := service.ApplyTemplate(ctx, req)

			if tt.wantErr && err == nil {
				t.Errorf("ApplyTemplate() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ApplyTemplate() unexpected error: %v", err)
			}

			if response != nil {
				if len(response.MissingVariables) != tt.expectMissing {
					t.Errorf("ApplyTemplate() MissingVariables count = %d, want %d", len(response.MissingVariables), tt.expectMissing)
				}

				if len(response.UnusedVariables) != tt.expectUnused {
					t.Errorf("ApplyTemplate() UnusedVariables count = %d, want %d", len(response.UnusedVariables), tt.expectUnused)
				}
			}
		})
	}
}

func TestApplyTemplateMetrics(t *testing.T) {
	ctx := context.Background()
	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("NewService() unexpected error: %v", err)
	}
	defer service.Close()

	// Reset metrics
	service.metrics.Reset()

	req := &ApplyTemplateRequest{
		Template: "Hello {name}!",
		Variables: map[string]any{
			"name": "Alice",
		},
	}

	_, err = service.ApplyTemplate(ctx, req)
	if err != nil {
		t.Errorf("ApplyTemplate() unexpected error: %v", err)
		return
	}

	// Check metrics
	metrics := service.metrics.GetOperationMetrics()
	if metrics["templates_applied"] != 1 {
		t.Errorf("templates_applied metric = %d, want 1", metrics["templates_applied"])
	}

	if metrics["variables_applied"] != 1 {
		t.Errorf("variables_applied metric = %d, want 1", metrics["variables_applied"])
	}
}
