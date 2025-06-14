// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// TemplateProcessor handles prompt template processing, validation, and variable substitution.
type TemplateProcessor struct {
	engine TemplateEngine
	mode   ValidationMode
}

// NewTemplateProcessor creates a new template processor with default settings.
func NewTemplateProcessor() *TemplateProcessor {
	return &TemplateProcessor{
		engine: TemplateEngineSimple,
		mode:   ValidationModeWarn,
	}
}

// NewTemplateProcessorWithOptions creates a template processor with specific settings.
func NewTemplateProcessorWithOptions(engine TemplateEngine, mode ValidationMode) *TemplateProcessor {
	return &TemplateProcessor{
		engine: engine,
		mode:   mode,
	}
}

// ValidateTemplate validates a prompt template and returns any errors.
func (tp *TemplateProcessor) ValidateTemplate(templateText string, declaredVars []string) error {
	result := tp.ValidateTemplateDetailed(templateText, declaredVars)
	if !result.IsValid {
		return NewInvalidTemplateError(templateText, result.Errors)
	}
	return nil
}

// ValidateTemplateDetailed performs detailed template validation.
func (tp *TemplateProcessor) ValidateTemplateDetailed(templateText string, declaredVars []string) *TemplateValidationResult {
	result := &TemplateValidationResult{
		IsValid: true,
	}

	// Extract variables from template
	detectedVars := tp.ExtractVariables(templateText)
	result.DetectedVars = detectedVars

	// Create maps for efficient lookup
	declaredMap := make(map[string]bool)
	for _, v := range declaredVars {
		declaredMap[v] = true
	}

	detectedMap := make(map[string]bool)
	for _, v := range detectedVars {
		detectedMap[v] = true
	}

	// Check for undeclared variables
	for _, detected := range detectedVars {
		if !declaredMap[detected] {
			result.UndeclaredVars = append(result.UndeclaredVars, detected)
			if tp.mode == ValidationModeStrict {
				result.Errors = append(result.Errors, fmt.Sprintf("undeclared variable: %s", detected))
				result.IsValid = false
			} else if tp.mode == ValidationModeWarn {
				result.Warnings = append(result.Warnings, fmt.Sprintf("undeclared variable: %s", detected))
			}
		}
	}

	// Check for unused declared variables
	for _, declared := range declaredVars {
		if !detectedMap[declared] {
			result.UnusedVars = append(result.UnusedVars, declared)
			if tp.mode == ValidationModeStrict {
				result.Warnings = append(result.Warnings, fmt.Sprintf("unused declared variable: %s", declared))
			}
		}
	}

	// Validate template syntax based on engine
	switch tp.engine {
	case TemplateEngineSimple:
		if err := tp.validateSimpleTemplate(templateText); err != nil {
			result.Errors = append(result.Errors, err.Error())
			result.IsValid = false
		}
	case TemplateEngineAdvanced:
		if err := tp.validateGoTemplate(templateText); err != nil {
			result.Errors = append(result.Errors, err.Error())
			result.IsValid = false
		}
	}

	return result
}

// ExtractVariables extracts all variable names from a template.
func (tp *TemplateProcessor) ExtractVariables(templateText string) []string {
	switch tp.engine {
	case TemplateEngineSimple:
		return tp.extractSimpleVariables(templateText)
	case TemplateEngineAdvanced:
		return tp.extractGoTemplateVariables(templateText)
	default:
		return tp.extractSimpleVariables(templateText)
	}
}

// ApplyVariables applies variables to a template and returns the result.
func (tp *TemplateProcessor) ApplyVariables(templateText string, variables map[string]any) (*ApplyTemplateResponse, error) {
	switch tp.engine {
	case TemplateEngineSimple:
		return tp.applySimpleVariables(templateText, variables)
	case TemplateEngineAdvanced:
		return tp.applyGoTemplateVariables(templateText, variables)
	default:
		return tp.applySimpleVariables(templateText, variables)
	}
}

// Simple template engine implementation (Python-style {variable} substitution)

// extractSimpleVariables extracts variables in {variable} format.
func (tp *TemplateProcessor) extractSimpleVariables(templateText string) []string {
	// Regular expression to match {variable_name} patterns
	re := regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
	matches := re.FindAllStringSubmatch(templateText, -1)

	variableSet := make(map[string]bool)
	var variables []string

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if !variableSet[varName] {
				variableSet[varName] = true
				variables = append(variables, varName)
			}
		}
	}

	return variables
}

// validateSimpleTemplate validates simple template syntax.
func (tp *TemplateProcessor) validateSimpleTemplate(templateText string) error {
	// Check for balanced braces
	braceCount := 0
	for i, char := range templateText {
		switch char {
		case '{':
			braceCount++
		case '}':
			braceCount--
			if braceCount < 0 {
				return fmt.Errorf("unmatched closing brace at position %d", i)
			}
		}
	}

	if braceCount > 0 {
		return fmt.Errorf("unmatched opening brace(s)")
	}

	// Check for invalid variable names
	re := regexp.MustCompile(`\{([^}]*)\}`)
	matches := re.FindAllStringSubmatch(templateText, -1)

	varNamePattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if varName == "" {
				return fmt.Errorf("empty variable name in template")
			}
			if !varNamePattern.MatchString(varName) {
				return fmt.Errorf("invalid variable name: %s", varName)
			}
		}
	}

	return nil
}

// applySimpleVariables applies variables using simple string replacement.
func (tp *TemplateProcessor) applySimpleVariables(templateText string, variables map[string]any) (*ApplyTemplateResponse, error) {
	response := &ApplyTemplateResponse{
		AppliedVariables: make(map[string]any),
	}

	// Extract variables from template
	templateVars := tp.extractSimpleVariables(templateText)

	// Track missing and unused variables
	var missingVars []string
	usedVars := make(map[string]bool)

	result := templateText

	// Replace each variable
	for _, varName := range templateVars {
		placeholder := fmt.Sprintf("{%s}", varName)

		if value, exists := variables[varName]; exists {
			// Convert value to string
			stringValue := fmt.Sprintf("%v", value)
			result = strings.ReplaceAll(result, placeholder, stringValue)
			response.AppliedVariables[varName] = value
			usedVars[varName] = true
		} else {
			missingVars = append(missingVars, varName)
		}
	}

	// Check for unused variables
	for varName := range variables {
		if !usedVars[varName] {
			response.UnusedVariables = append(response.UnusedVariables, varName)
		}
	}

	response.Content = result
	response.MissingVariables = missingVars

	// Return error if there are missing variables in strict mode
	if len(missingVars) > 0 && tp.mode == ValidationModeStrict {
		return response, NewMissingVariablesError(missingVars)
	}

	return response, nil
}

// Advanced template engine implementation (Go text/template)

// extractGoTemplateVariables extracts variables from Go template syntax.
func (tp *TemplateProcessor) extractGoTemplateVariables(templateText string) []string {
	// This is a simplified extraction for Go templates
	// In a real implementation, you'd parse the template AST
	re := regexp.MustCompile(`\{\{\s*\.([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)
	matches := re.FindAllStringSubmatch(templateText, -1)

	variableSet := make(map[string]bool)
	var variables []string

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if !variableSet[varName] {
				variableSet[varName] = true
				variables = append(variables, varName)
			}
		}
	}

	return variables
}

// validateGoTemplate validates Go template syntax.
func (tp *TemplateProcessor) validateGoTemplate(templateText string) error {
	_, err := template.New("validation").Parse(templateText)
	return err
}

// applyGoTemplateVariables applies variables using Go text/template.
func (tp *TemplateProcessor) applyGoTemplateVariables(templateText string, variables map[string]any) (*ApplyTemplateResponse, error) {
	response := &ApplyTemplateResponse{
		AppliedVariables: variables,
	}

	// Parse the template
	tmpl, err := template.New("prompt").Parse(templateText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	response.Content = buf.String()
	return response, nil
}

// TemplateCompiler compiles templates for better performance with repeated use.
type TemplateCompiler struct {
	cache     map[string]*CompiledTemplate
	processor *TemplateProcessor
}

// CompiledTemplate represents a pre-compiled template for efficient variable substitution.
type CompiledTemplate struct {
	originalTemplate string
	variables        []string
	compiledTemplate *template.Template
	engine           TemplateEngine
	compiledAt       int64
}

// NewTemplateCompiler creates a new template compiler.
func NewTemplateCompiler(processor *TemplateProcessor) *TemplateCompiler {
	return &TemplateCompiler{
		cache:     make(map[string]*CompiledTemplate),
		processor: processor,
	}
}

// Compile compiles a template for efficient repeated use.
func (tc *TemplateCompiler) Compile(templateText string) (*CompiledTemplate, error) {
	// Check cache first
	if compiled, exists := tc.cache[templateText]; exists {
		return compiled, nil
	}

	compiled := &CompiledTemplate{
		originalTemplate: templateText,
		variables:        tc.processor.ExtractVariables(templateText),
		engine:           tc.processor.engine,
		compiledAt:       time.Now().Unix(),
	}

	// Compile based on engine
	switch tc.processor.engine {
	case TemplateEngineAdvanced:
		tmpl, err := template.New("compiled").Parse(templateText)
		if err != nil {
			return nil, fmt.Errorf("failed to compile template: %w", err)
		}
		compiled.compiledTemplate = tmpl
	}

	// Cache the compiled template
	tc.cache[templateText] = compiled

	return compiled, nil
}

// Execute executes a compiled template with the given variables.
func (ct *CompiledTemplate) Execute(variables map[string]any) (*ApplyTemplateResponse, error) {
	switch ct.engine {
	case TemplateEngineSimple:
		// Use simple string replacement
		processor := NewTemplateProcessorWithOptions(TemplateEngineSimple, ValidationModeWarn)
		return processor.applySimpleVariables(ct.originalTemplate, variables)
	case TemplateEngineAdvanced:
		// Use compiled Go template
		if ct.compiledTemplate == nil {
			return nil, fmt.Errorf("template not properly compiled")
		}

		var buf bytes.Buffer
		if err := ct.compiledTemplate.Execute(&buf, variables); err != nil {
			return nil, fmt.Errorf("failed to execute compiled template: %w", err)
		}

		return &ApplyTemplateResponse{
			Content:          buf.String(),
			AppliedVariables: variables,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported template engine: %s", ct.engine)
	}
}

// GetVariables returns the variables used in the compiled template.
func (ct *CompiledTemplate) GetVariables() []string {
	return ct.variables
}

// GetOriginalTemplate returns the original template text.
func (ct *CompiledTemplate) GetOriginalTemplate() string {
	return ct.originalTemplate
}
