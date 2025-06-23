// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"fmt"
	"log/slog"
)

// Template application methods for the prompts service

// ApplyTemplate applies variables to a prompt template and returns the resulting content.
func (s *service) ApplyTemplate(ctx context.Context, req *ApplyTemplateRequest) (*ApplyTemplateResponse, error) {
	tracker := s.metrics.StartOperation("apply_template")
	defer tracker.Finish()

	if err := s.validateApplyTemplateRequest(req); err != nil {
		tracker.FinishWithError("validation")
		return nil, err
	}

	var template string
	var declaredVariables []string

	// Get the template content
	if req.Template != "" {
		// Use provided template directly
		template = req.Template
	} else {
		// Load template from stored prompt
		prompt, err := s.getPromptForTemplate(ctx, req)
		if err != nil {
			tracker.FinishWithError("cloud")
			return nil, err
		}
		template = prompt.Template
		declaredVariables = prompt.Variables
	}

	// Validate variables if requested
	if req.ValidateVariables {
		if err := s.validateTemplateVariables(template, declaredVariables, req.Variables, req.StrictMode); err != nil {
			tracker.FinishWithError("validation")
			return nil, err
		}
	}

	// Apply the variables to the template
	response, err := s.templateEngine.ApplyVariables(template, req.Variables)
	if err != nil {
		tracker.FinishWithError("template")
		return nil, fmt.Errorf("failed to apply template variables: %w", err)
	}

	// Track metrics
	s.metrics.IncrementTemplateApplied()
	s.metrics.IncrementVariablesApplied(int64(len(req.Variables)))

	s.logger.InfoContext(ctx, "Template applied successfully",
		slog.String("prompt_id", req.PromptID),
		slog.String("name", req.Name),
		slog.Int("variables_count", len(req.Variables)),
		slog.Int("content_length", len(response.Content)),
	)

	return response, nil
}

// ApplyTemplateToPrompt is a convenience method that applies variables to a prompt object.
func (s *service) ApplyTemplateToPrompt(ctx context.Context, prompt *Prompt, variables map[string]any) (*ApplyTemplateResponse, error) {
	return s.ApplyTemplate(ctx, &ApplyTemplateRequest{
		Template:          prompt.Template,
		Variables:         variables,
		ValidateVariables: true,
		StrictMode:        false,
	})
}

// ApplyTemplateSimple is a simplified method for quick template application.
func (s *service) ApplyTemplateSimple(ctx context.Context, promptID string, variables map[string]any) (string, error) {
	response, err := s.ApplyTemplate(ctx, &ApplyTemplateRequest{
		PromptID:          promptID,
		Variables:         variables,
		ValidateVariables: false,
	})
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// ValidateTemplate validates a template without applying variables.
func (s *service) ValidateTemplate(ctx context.Context, template string, variables []string) (*TemplateValidationResult, error) {
	return s.templateEngine.ValidateTemplateDetailed(template, variables), nil
}

// PreviewTemplate previews how a template would look with sample variables.
func (s *service) PreviewTemplate(ctx context.Context, template string, sampleVariables map[string]any) (*ApplyTemplateResponse, error) {
	// Create a copy of the template processor in loose mode for preview
	previewProcessor := NewTemplateProcessorWithOptions(s.templateEngine.engine, ValidationModeLoose)

	response, err := previewProcessor.ApplyVariables(template, sampleVariables)
	if err != nil {
		return nil, fmt.Errorf("failed to preview template: %w", err)
	}

	return response, nil
}

// ExtractVariables extracts all variables from a template.
func (s *service) ExtractVariables(ctx context.Context, template string) ([]string, error) {
	variables := s.templateEngine.ExtractVariables(template)

	s.logger.InfoContext(ctx, "Variables extracted from template",
		slog.Int("variables_count", len(variables)),
		slog.Any("variables", variables),
	)

	return variables, nil
}

// Batch template operations

// BatchApplyTemplates applies variables to multiple templates in a single operation.
func (s *service) BatchApplyTemplates(ctx context.Context, requests []*ApplyTemplateRequest) ([]*BatchTemplateResult, error) {
	if len(requests) == 0 {
		return nil, NewInvalidRequestError("requests", "cannot be empty")
	}

	results := make([]*BatchTemplateResult, len(requests))

	for i, req := range requests {
		result := &BatchTemplateResult{
			Index: int32(i),
		}

		response, err := s.ApplyTemplate(ctx, req)
		if err != nil {
			result.Error = err.Error()
			result.Success = false
		} else {
			result.Response = response
			result.Success = true
		}

		results[i] = result
	}

	return results, nil
}

// CompileTemplate compiles a template for efficient repeated use.
func (s *service) CompileTemplate(ctx context.Context, template string) (*CompiledTemplate, error) {
	compiler := NewTemplateCompiler(s.templateEngine)

	compiled, err := compiler.Compile(template)
	if err != nil {
		return nil, fmt.Errorf("failed to compile template: %w", err)
	}

	s.logger.InfoContext(ctx, "Template compiled successfully",
		slog.Int("variables_count", len(compiled.GetVariables())),
		slog.String("engine", string(compiled.engine)),
	)

	return compiled, nil
}

// Helper methods

// validateApplyTemplateRequest validates an apply template request.
func (s *service) validateApplyTemplateRequest(req *ApplyTemplateRequest) error {
	if req == nil {
		return NewInvalidRequestError("request", "cannot be nil")
	}

	// Must have either template content or prompt identifier
	if req.Template == "" && req.PromptID == "" && req.Name == "" {
		return NewInvalidRequestError("template_or_prompt", "must specify either template content or prompt identifier")
	}

	if req.Variables == nil {
		return NewInvalidRequestError("variables", "cannot be nil")
	}

	return nil
}

// getPromptForTemplate retrieves a prompt for template application.
func (s *service) getPromptForTemplate(ctx context.Context, req *ApplyTemplateRequest) (*Prompt, error) {
	getReq := &GetPromptRequest{}

	if req.PromptID != "" {
		getReq.PromptID = req.PromptID
	} else if req.Name != "" {
		getReq.Name = req.Name
	}

	if req.VersionID != "" {
		getReq.VersionID = req.VersionID
	}

	prompt, err := s.GetPrompt(ctx, getReq)
	if err != nil {
		return nil, err
	}

	return prompt, nil
}

// validateTemplateVariables validates template variables.
func (s *service) validateTemplateVariables(template string, declaredVars []string, providedVars map[string]any, strictMode bool) error {
	// Extract variables from template
	templateVars := s.templateEngine.ExtractVariables(template)

	// Check for missing required variables
	var missingVars []string
	for _, templateVar := range templateVars {
		if _, exists := providedVars[templateVar]; !exists {
			missingVars = append(missingVars, templateVar)
		}
	}

	if len(missingVars) > 0 {
		if strictMode {
			return NewMissingVariablesError(missingVars)
		}
		// In non-strict mode, just log warnings
		s.logger.Warn("Missing template variables",
			slog.Any("missing_variables", missingVars))
	}

	// Check for undeclared variables in strict mode
	if strictMode && len(declaredVars) > 0 {
		declaredMap := make(map[string]bool)
		for _, declared := range declaredVars {
			declaredMap[declared] = true
		}

		var undeclaredVars []string
		for _, templateVar := range templateVars {
			if !declaredMap[templateVar] {
				undeclaredVars = append(undeclaredVars, templateVar)
			}
		}

		if len(undeclaredVars) > 0 {
			return NewInvalidVariableError("undeclared", fmt.Sprintf("undeclared variables: %v", undeclaredVars))
		}
	}

	return nil
}

// Additional types for batch operations

// BatchTemplateResult represents the result of a single template application in a batch.
type BatchTemplateResult struct {
	Index    int32                  `json:"index"`
	Response *ApplyTemplateResponse `json:"response,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Success  bool                   `json:"success"`
}

// TemplatePreview provides a preview of template rendering.
type TemplatePreview struct {
	Template          string                    `json:"template"`
	SampleVariables   map[string]any            `json:"sample_variables"`
	PreviewContent    string                    `json:"preview_content"`
	DetectedVariables []string                  `json:"detected_variables"`
	ValidationResult  *TemplateValidationResult `json:"validation_result"`
}

// GenerateTemplatePreview generates a comprehensive template preview.
func (s *service) GenerateTemplatePreview(ctx context.Context, template string, sampleVariables map[string]any) (*TemplatePreview, error) {
	// Extract variables
	detectedVars := s.templateEngine.ExtractVariables(template)

	// Validate template
	validationResult := s.templateEngine.ValidateTemplateDetailed(template, detectedVars)

	// Generate preview content
	previewResponse, err := s.PreviewTemplate(ctx, template, sampleVariables)
	if err != nil {
		return nil, fmt.Errorf("failed to generate template preview: %w", err)
	}

	preview := &TemplatePreview{
		Template:          template,
		SampleVariables:   sampleVariables,
		PreviewContent:    previewResponse.Content,
		DetectedVariables: detectedVars,
		ValidationResult:  validationResult,
	}

	return preview, nil
}

// TemplateAnalyzer provides analysis of template usage patterns.
type TemplateAnalyzer struct {
	service *service
}

// NewTemplateAnalyzer creates a new template analyzer.
func (s *service) NewTemplateAnalyzer() *TemplateAnalyzer {
	return &TemplateAnalyzer{
		service: s,
	}
}

// AnalyzeTemplate provides detailed analysis of a template.
func (ta *TemplateAnalyzer) AnalyzeTemplate(ctx context.Context, template string) (*TemplateAnalysis, error) {
	// Extract variables
	variables := ta.service.templateEngine.ExtractVariables(template)

	// Validate template
	validation := ta.service.templateEngine.ValidateTemplateDetailed(template, variables)

	// Calculate complexity metrics
	complexity := ta.calculateComplexity(template, variables)

	analysis := &TemplateAnalysis{
		Template:         template,
		Variables:        variables,
		ValidationResult: validation,
		Complexity:       complexity,
		Recommendations:  ta.generateRecommendations(template, variables, validation),
	}

	return analysis, nil
}

// TemplateAnalysis represents a comprehensive template analysis.
type TemplateAnalysis struct {
	Template         string                    `json:"template"`
	Variables        []string                  `json:"variables"`
	ValidationResult *TemplateValidationResult `json:"validation_result"`
	Complexity       *TemplateComplexity       `json:"complexity"`
	Recommendations  []string                  `json:"recommendations"`
}

// TemplateComplexity represents template complexity metrics.
type TemplateComplexity struct {
	CharacterCount   int     `json:"character_count"`
	VariableCount    int     `json:"variable_count"`
	UniqueVariables  int     `json:"unique_variables"`
	ComplexityScore  float64 `json:"complexity_score"`
	ReadabilityScore float64 `json:"readability_score"`
}

// calculateComplexity calculates template complexity metrics.
func (ta *TemplateAnalyzer) calculateComplexity(template string, variables []string) *TemplateComplexity {
	charCount := len(template)
	varCount := len(variables)

	// Calculate unique variables
	uniqueVars := make(map[string]bool)
	for _, v := range variables {
		uniqueVars[v] = true
	}
	uniqueVarCount := len(uniqueVars)

	// Simple complexity scoring (can be enhanced with more sophisticated algorithms)
	complexityScore := float64(charCount)/100.0 + float64(varCount)*2.0

	// Simple readability score (inverse of complexity, normalized)
	readabilityScore := 100.0 / (1.0 + complexityScore/10.0)

	return &TemplateComplexity{
		CharacterCount:   charCount,
		VariableCount:    varCount,
		UniqueVariables:  uniqueVarCount,
		ComplexityScore:  complexityScore,
		ReadabilityScore: readabilityScore,
	}
}

// generateRecommendations generates improvement recommendations for a template.
func (ta *TemplateAnalyzer) generateRecommendations(template string, variables []string, validation *TemplateValidationResult) []string {
	var recommendations []string

	if len(variables) > 10 {
		recommendations = append(recommendations, "Consider reducing the number of variables for better maintainability")
	}

	if len(template) > 1000 {
		recommendations = append(recommendations, "Template is quite long; consider breaking it into smaller, reusable components")
	}

	if len(validation.UndeclaredVars) > 0 {
		recommendations = append(recommendations, "Declare all template variables for better documentation and validation")
	}

	if len(validation.UnusedVars) > 0 {
		recommendations = append(recommendations, "Remove unused declared variables to keep the template clean")
	}

	return recommendations
}
