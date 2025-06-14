// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTemplateProcessor_ExtractVariables(t *testing.T) {
	processor := NewTemplateProcessor()

	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "simple_variable",
			template: "Hello {name}!",
			want:     []string{"name"},
		},
		{
			name:     "multiple_variables",
			template: "Hello {name}, welcome to {company}!",
			want:     []string{"name", "company"},
		},
		{
			name:     "duplicate_variables",
			template: "Hello {name}, {name} is a great name!",
			want:     []string{"name"},
		},
		{
			name:     "no_variables",
			template: "Hello world!",
			want:     nil,
		},
		{
			name:     "empty_template",
			template: "",
			want:     nil,
		},
		{
			name:     "complex_variables",
			template: "User {user_id} with email {email_address} and role {user_role}",
			want:     []string{"user_id", "email_address", "user_role"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processor.ExtractVariables(tt.template)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ExtractVariables() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTemplateProcessor_ValidateTemplate(t *testing.T) {
	processor := NewTemplateProcessor()

	tests := []struct {
		name         string
		template     string
		declaredVars []string
		wantErr      bool
	}{
		{
			name:         "valid_template",
			template:     "Hello {name}!",
			declaredVars: []string{"name"},
			wantErr:      false,
		},
		{
			name:         "unmatched_opening_brace",
			template:     "Hello {name!",
			declaredVars: []string{"name"},
			wantErr:      true,
		},
		{
			name:         "unmatched_closing_brace",
			template:     "Hello name}!",
			declaredVars: []string{"name"},
			wantErr:      true,
		},
		{
			name:         "empty_variable_name",
			template:     "Hello {}!",
			declaredVars: []string{},
			wantErr:      true,
		},
		{
			name:         "invalid_variable_name",
			template:     "Hello {123invalid}!",
			declaredVars: []string{},
			wantErr:      true,
		},
		{
			name:         "valid_underscore_variable",
			template:     "Hello {user_name}!",
			declaredVars: []string{"user_name"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ValidateTemplate(tt.template, tt.declaredVars)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateProcessor_ValidateTemplateDetailed(t *testing.T) {
	processor := NewTemplateProcessorWithOptions(TemplateEngineSimple, ValidationModeStrict)

	tests := []struct {
		name           string
		template       string
		declaredVars   []string
		wantValid      bool
		wantUndeclared int
		wantUnused     int
	}{
		{
			name:           "all_variables_declared_and_used",
			template:       "Hello {name}!",
			declaredVars:   []string{"name"},
			wantValid:      true,
			wantUndeclared: 0,
			wantUnused:     0,
		},
		{
			name:           "undeclared_variable",
			template:       "Hello {name}!",
			declaredVars:   []string{},
			wantValid:      false,
			wantUndeclared: 1,
			wantUnused:     0,
		},
		{
			name:           "unused_declared_variable",
			template:       "Hello world!",
			declaredVars:   []string{"name"},
			wantValid:      true,
			wantUndeclared: 0,
			wantUnused:     1,
		},
		{
			name:           "mixed_scenario",
			template:       "Hello {name}, welcome to {company}!",
			declaredVars:   []string{"name", "unused_var"},
			wantValid:      false,
			wantUndeclared: 1,
			wantUnused:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.ValidateTemplateDetailed(tt.template, tt.declaredVars)

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateTemplateDetailed() IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}

			if len(result.UndeclaredVars) != tt.wantUndeclared {
				t.Errorf("ValidateTemplateDetailed() UndeclaredVars count = %d, want %d", len(result.UndeclaredVars), tt.wantUndeclared)
			}

			if len(result.UnusedVars) != tt.wantUnused {
				t.Errorf("ValidateTemplateDetailed() UnusedVars count = %d, want %d", len(result.UnusedVars), tt.wantUnused)
			}
		})
	}
}

func TestTemplateProcessor_ApplyVariables(t *testing.T) {
	processor := NewTemplateProcessor()

	tests := []struct {
		name      string
		template  string
		variables map[string]any
		want      string
		wantErr   bool
	}{
		{
			name:     "simple_substitution",
			template: "Hello {name}!",
			variables: map[string]any{
				"name": "Alice",
			},
			want:    "Hello Alice!",
			wantErr: false,
		},
		{
			name:     "multiple_substitutions",
			template: "Hello {name}, welcome to {company}!",
			variables: map[string]any{
				"name":    "Bob",
				"company": "Acme Corp",
			},
			want:    "Hello Bob, welcome to Acme Corp!",
			wantErr: false,
		},
		{
			name:     "number_substitution",
			template: "You have {count} items",
			variables: map[string]any{
				"count": 42,
			},
			want:    "You have 42 items",
			wantErr: false,
		},
		{
			name:     "boolean_substitution",
			template: "Status: {active}",
			variables: map[string]any{
				"active": true,
			},
			want:    "Status: true",
			wantErr: false,
		},
		{
			name:      "missing_variable_loose_mode",
			template:  "Hello {name}!",
			variables: map[string]any{},
			want:      "Hello {name}!",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := processor.ApplyVariables(tt.template, tt.variables)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if response.Content != tt.want {
					t.Errorf("ApplyVariables() content = %v, want %v", response.Content, tt.want)
				}
			}
		})
	}
}

func TestTemplateProcessor_StrictMode(t *testing.T) {
	processor := NewTemplateProcessorWithOptions(TemplateEngineSimple, ValidationModeStrict)

	template := "Hello {name}!"
	variables := map[string]any{} // Missing required variable

	response, err := processor.ApplyVariables(template, variables)

	// Should return error in strict mode for missing variables
	if err == nil {
		t.Errorf("ApplyVariables() expected error in strict mode for missing variables")
	}

	// Response should still contain information about missing variables
	if response == nil {
		t.Errorf("ApplyVariables() response should not be nil even with error")
	} else if len(response.MissingVariables) != 1 {
		t.Errorf("ApplyVariables() MissingVariables count = %d, want 1", len(response.MissingVariables))
	}
}

func TestTemplateCompiler(t *testing.T) {
	processor := NewTemplateProcessor()
	compiler := NewTemplateCompiler(processor)

	template := "Hello {name}, welcome to {company}!"

	// Test compilation
	compiled, err := compiler.Compile(template)
	if err != nil {
		t.Fatalf("Compile() unexpected error: %v", err)
	}

	if compiled == nil {
		t.Fatalf("Compile() returned nil compiled template")
	}

	// Test variable extraction
	variables := compiled.GetVariables()
	expectedVars := []string{"name", "company"}
	if diff := cmp.Diff(expectedVars, variables); diff != "" {
		t.Errorf("GetVariables() mismatch (-want +got):\n%s", diff)
	}

	// Test execution
	testVars := map[string]any{
		"name":    "Alice",
		"company": "Tech Corp",
	}

	response, err := compiled.Execute(testVars)
	if err != nil {
		t.Errorf("Execute() unexpected error: %v", err)
	}

	expectedContent := "Hello Alice, welcome to Tech Corp!"
	if response.Content != expectedContent {
		t.Errorf("Execute() content = %v, want %v", response.Content, expectedContent)
	}

	// Test caching - compile same template again
	compiled2, err := compiler.Compile(template)
	if err != nil {
		t.Errorf("Compile() second call unexpected error: %v", err)
	}

	// Should return the same cached instance
	if compiled != compiled2 {
		t.Errorf("Compile() should return cached instance for same template")
	}
}

func TestAdvancedTemplateEngine(t *testing.T) {
	processor := NewTemplateProcessorWithOptions(TemplateEngineAdvanced, ValidationModeWarn)

	// Test Go template syntax
	template := "Hello {{.Name}}, you have {{.Count}} messages"
	variables := map[string]any{
		"Name":  "Alice",
		"Count": 5,
	}

	response, err := processor.ApplyVariables(template, variables)
	if err != nil {
		t.Errorf("ApplyVariables() with Go template unexpected error: %v", err)
	}

	expected := "Hello Alice, you have 5 messages"
	if response.Content != expected {
		t.Errorf("ApplyVariables() content = %v, want %v", response.Content, expected)
	}
}

func TestTemplateValidationModes(t *testing.T) {
	template := "Hello {name}!"
	declaredVars := []string{} // No declared variables

	tests := []struct {
		name     string
		mode     ValidationMode
		wantErr  bool
		wantWarn bool
	}{
		{
			name:     "strict_mode",
			mode:     ValidationModeStrict,
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "warn_mode",
			mode:     ValidationModeWarn,
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "loose_mode",
			mode:     ValidationModeLoose,
			wantErr:  false,
			wantWarn: false,
		},
		{
			name:     "none_mode",
			mode:     ValidationModeNone,
			wantErr:  false,
			wantWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewTemplateProcessorWithOptions(TemplateEngineSimple, tt.mode)
			result := processor.ValidateTemplateDetailed(template, declaredVars)

			if tt.wantErr && result.IsValid {
				t.Errorf("ValidateTemplateDetailed() expected invalid result in %s mode", tt.mode)
			}

			if !tt.wantErr && !result.IsValid {
				t.Errorf("ValidateTemplateDetailed() expected valid result in %s mode", tt.mode)
			}

			if tt.wantWarn && len(result.Warnings) == 0 {
				t.Errorf("ValidateTemplateDetailed() expected warnings in %s mode", tt.mode)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	processor := NewTemplateProcessor()

	tests := []struct {
		name      string
		template  string
		variables map[string]any
		wantErr   bool
	}{
		{
			name:      "empty_template",
			template:  "",
			variables: map[string]any{},
			wantErr:   false,
		},
		{
			name:      "only_braces",
			template:  "{}",
			variables: map[string]any{},
			wantErr:   false, // Empty braces are treated as literal text, not a validation error
		},
		{
			name:      "nested_braces",
			template:  "Hello {{name}}!",
			variables: map[string]any{},
			wantErr:   false, // Should be treated as literal text
		},
		{
			name:     "special_characters_in_variables",
			template: "Hello {name}!",
			variables: map[string]any{
				"name": "Alice & Bob",
			},
			wantErr: false,
		},
		{
			name:     "unicode_in_variables",
			template: "Hello {name}!",
			variables: map[string]any{
				"name": "Alice 🚀",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.ApplyVariables(tt.template, tt.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyVariables() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPerformance(t *testing.T) {
	processor := NewTemplateProcessor()
	template := "Hello {name}, welcome to {company}! Your user ID is {user_id}."
	variables := map[string]any{
		"name":    "Alice",
		"company": "Tech Corp",
		"user_id": "12345",
	}

	// Test without compilation (should be slower for repeated use)
	for i := 0; i < 100; i++ {
		_, err := processor.ApplyVariables(template, variables)
		if err != nil {
			t.Fatalf("ApplyVariables() unexpected error: %v", err)
		}
	}

	// Test with compilation (should be faster for repeated use)
	compiler := NewTemplateCompiler(processor)
	compiled, err := compiler.Compile(template)
	if err != nil {
		t.Fatalf("Compile() unexpected error: %v", err)
	}

	for i := 0; i < 100; i++ {
		_, err := compiled.Execute(variables)
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
	}
}

func BenchmarkTemplateApplication(b *testing.B) {
	processor := NewTemplateProcessor()
	template := "Hello {name}, welcome to {company}!"
	variables := map[string]any{
		"name":    "Alice",
		"company": "Tech Corp",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ApplyVariables(template, variables)
		if err != nil {
			b.Fatalf("ApplyVariables() unexpected error: %v", err)
		}
	}
}

func BenchmarkCompiledTemplateExecution(b *testing.B) {
	processor := NewTemplateProcessor()
	compiler := NewTemplateCompiler(processor)
	template := "Hello {name}, welcome to {company}!"
	variables := map[string]any{
		"name":    "Alice",
		"company": "Tech Corp",
	}

	compiled, err := compiler.Compile(template)
	if err != nil {
		b.Fatalf("Compile() unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compiled.Execute(variables)
		if err != nil {
			b.Fatalf("Execute() unexpected error: %v", err)
		}
	}
}
