// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"errors"
	"fmt"
)

// Error types for prompt operations.
var (
	// ErrPromptNotFound indicates that a requested prompt was not found.
	ErrPromptNotFound = errors.New("prompt not found")

	// ErrPromptAlreadyExists indicates that a prompt with the given name/ID already exists.
	ErrPromptAlreadyExists = errors.New("prompt already exists")

	// ErrInvalidTemplate indicates that a prompt template is invalid.
	ErrInvalidTemplate = errors.New("invalid template")

	// ErrVersionNotFound indicates that a requested prompt version was not found.
	ErrVersionNotFound = errors.New("prompt version not found")

	// ErrVersionConflict indicates a version conflict during updates.
	ErrVersionConflict = errors.New("version conflict")

	// ErrInvalidVariable indicates that a template variable is invalid.
	ErrInvalidVariable = errors.New("invalid template variable")

	// ErrMissingVariables indicates that required template variables are missing.
	ErrMissingVariables = errors.New("missing required variables")

	// ErrUnauthorized indicates insufficient permissions for the operation.
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrQuotaExceeded indicates that a quota or limit has been exceeded.
	ErrQuotaExceeded = errors.New("quota exceeded")

	// ErrInvalidRequest indicates that the request parameters are invalid.
	ErrInvalidRequest = errors.New("invalid request")

	// ErrServiceUnavailable indicates that the service is temporarily unavailable.
	ErrServiceUnavailable = errors.New("service unavailable")
)

// PromptError represents a detailed error with additional context.
type PromptError struct {
	// Code is the error code
	Code string `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// Details provides additional error context
	Details map[string]any `json:"details,omitempty"`

	// PromptID is the prompt identifier associated with the error
	PromptID string `json:"prompt_id,omitempty"`

	// VersionID is the version identifier associated with the error
	VersionID string `json:"version_id,omitempty"`

	// Underlying error
	Err error `json:"-"`
}

// Error implements the error interface.
func (e *PromptError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *PromptError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target error.
func (e *PromptError) Is(target error) bool {
	if e.Err != nil {
		return errors.Is(e.Err, target)
	}
	return false
}

// NewPromptError creates a new PromptError.
func NewPromptError(code, message string, err error) *PromptError {
	return &PromptError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewPromptNotFoundError creates a prompt not found error.
func NewPromptNotFoundError(promptID string) *PromptError {
	return &PromptError{
		Code:     "PROMPT_NOT_FOUND",
		Message:  "prompt not found",
		PromptID: promptID,
		Err:      ErrPromptNotFound,
	}
}

// NewPromptAlreadyExistsError creates a prompt already exists error.
func NewPromptAlreadyExistsError(name string) *PromptError {
	return &PromptError{
		Code:    "PROMPT_ALREADY_EXISTS",
		Message: fmt.Sprintf("prompt with name '%s' already exists", name),
		Details: map[string]any{"name": name},
		Err:     ErrPromptAlreadyExists,
	}
}

// NewInvalidTemplateError creates an invalid template error.
func NewInvalidTemplateError(template string, details []string) *PromptError {
	return &PromptError{
		Code:    "INVALID_TEMPLATE",
		Message: "template validation failed",
		Details: map[string]any{
			"template": template,
			"errors":   details,
		},
		Err: ErrInvalidTemplate,
	}
}

// NewVersionNotFoundError creates a version not found error.
func NewVersionNotFoundError(promptID, versionID string) *PromptError {
	return &PromptError{
		Code:      "VERSION_NOT_FOUND",
		Message:   "prompt version not found",
		PromptID:  promptID,
		VersionID: versionID,
		Err:       ErrVersionNotFound,
	}
}

// NewVersionConflictError creates a version conflict error.
func NewVersionConflictError(promptID, expectedVersion, actualVersion string) *PromptError {
	return &PromptError{
		Code:     "VERSION_CONFLICT",
		Message:  "version conflict detected",
		PromptID: promptID,
		Details: map[string]any{
			"expected_version": expectedVersion,
			"actual_version":   actualVersion,
		},
		Err: ErrVersionConflict,
	}
}

// NewMissingVariablesError creates a missing variables error.
func NewMissingVariablesError(missing []string) *PromptError {
	return &PromptError{
		Code:    "MISSING_VARIABLES",
		Message: "required template variables are missing",
		Details: map[string]any{"missing_variables": missing},
		Err:     ErrMissingVariables,
	}
}

// NewInvalidVariableError creates an invalid variable error.
func NewInvalidVariableError(variable, reason string) *PromptError {
	return &PromptError{
		Code:    "INVALID_VARIABLE",
		Message: fmt.Sprintf("invalid variable '%s': %s", variable, reason),
		Details: map[string]any{
			"variable": variable,
			"reason":   reason,
		},
		Err: ErrInvalidVariable,
	}
}

// NewUnauthorizedError creates an unauthorized error.
func NewUnauthorizedError(operation, resource string) *PromptError {
	return &PromptError{
		Code:    "UNAUTHORIZED",
		Message: fmt.Sprintf("unauthorized to perform %s on %s", operation, resource),
		Details: map[string]any{
			"operation": operation,
			"resource":  resource,
		},
		Err: ErrUnauthorized,
	}
}

// NewQuotaExceededError creates a quota exceeded error.
func NewQuotaExceededError(quota string, limit int64) *PromptError {
	return &PromptError{
		Code:    "QUOTA_EXCEEDED",
		Message: fmt.Sprintf("quota exceeded for %s (limit: %d)", quota, limit),
		Details: map[string]any{
			"quota": quota,
			"limit": limit,
		},
		Err: ErrQuotaExceeded,
	}
}

// NewInvalidRequestError creates an invalid request error.
func NewInvalidRequestError(field, reason string) *PromptError {
	return &PromptError{
		Code:    "INVALID_REQUEST",
		Message: fmt.Sprintf("invalid request: %s - %s", field, reason),
		Details: map[string]any{
			"field":  field,
			"reason": reason,
		},
		Err: ErrInvalidRequest,
	}
}

// NewServiceUnavailableError creates a service unavailable error.
func NewServiceUnavailableError(reason string) *PromptError {
	return &PromptError{
		Code:    "SERVICE_UNAVAILABLE",
		Message: "service is temporarily unavailable",
		Details: map[string]any{"reason": reason},
		Err:     ErrServiceUnavailable,
	}
}

// Error checking helper functions

// IsNotFound checks if the error indicates a prompt was not found.
func IsNotFound(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "PROMPT_NOT_FOUND" || promptErr.Code == "VERSION_NOT_FOUND"
	}
	return errors.Is(err, ErrPromptNotFound) || errors.Is(err, ErrVersionNotFound)
}

// IsAlreadyExists checks if the error indicates a prompt already exists.
func IsAlreadyExists(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "PROMPT_ALREADY_EXISTS"
	}
	return errors.Is(err, ErrPromptAlreadyExists)
}

// IsInvalidTemplate checks if the error indicates an invalid template.
func IsInvalidTemplate(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "INVALID_TEMPLATE"
	}
	return errors.Is(err, ErrInvalidTemplate)
}

// IsVersionConflict checks if the error indicates a version conflict.
func IsVersionConflict(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "VERSION_CONFLICT"
	}
	return errors.Is(err, ErrVersionConflict)
}

// IsMissingVariables checks if the error indicates missing variables.
func IsMissingVariables(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "MISSING_VARIABLES"
	}
	return errors.Is(err, ErrMissingVariables)
}

// IsInvalidVariable checks if the error indicates an invalid variable.
func IsInvalidVariable(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "INVALID_VARIABLE"
	}
	return errors.Is(err, ErrInvalidVariable)
}

// IsUnauthorized checks if the error indicates unauthorized access.
func IsUnauthorized(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "UNAUTHORIZED"
	}
	return errors.Is(err, ErrUnauthorized)
}

// IsQuotaExceeded checks if the error indicates quota exceeded.
func IsQuotaExceeded(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "QUOTA_EXCEEDED"
	}
	return errors.Is(err, ErrQuotaExceeded)
}

// IsInvalidRequest checks if the error indicates an invalid request.
func IsInvalidRequest(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "INVALID_REQUEST"
	}
	return errors.Is(err, ErrInvalidRequest)
}

// IsServiceUnavailable checks if the error indicates service unavailability.
func IsServiceUnavailable(err error) bool {
	var promptErr *PromptError
	if errors.As(err, &promptErr) {
		return promptErr.Code == "SERVICE_UNAVAILABLE"
	}
	return errors.Is(err, ErrServiceUnavailable)
}

// ValidationError represents a template validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}
	return fmt.Sprintf("validation failed with %d errors: %s", len(ve), ve[0].Message)
}

// Add adds a validation error.
func (ve *ValidationErrors) Add(field, message string, value any) {
	*ve = append(*ve, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors.
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Fields returns the fields that have validation errors.
func (ve ValidationErrors) Fields() []string {
	fields := make([]string, len(ve))
	for i, err := range ve {
		fields[i] = err.Field
	}
	return fields
}

// Messages returns the validation error messages.
func (ve ValidationErrors) Messages() []string {
	messages := make([]string, len(ve))
	for i, err := range ve {
		messages[i] = err.Message
	}
	return messages
}
