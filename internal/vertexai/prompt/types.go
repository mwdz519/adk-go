// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"time"

	"google.golang.org/genai"
)

// Prompt represents a prompt template with metadata and versioning information.
//
// This type mirrors the Python vertexai.prompts.Prompt class and provides
// comprehensive prompt management capabilities including versioning,
// variable substitution, and cloud storage integration.
type Prompt struct {
	// ID is the unique prompt identifier
	ID string `json:"id,omitempty"`

	// Name is the prompt name (used as identifier for operations)
	Name string `json:"name"`

	// DisplayName is a human-readable name for the prompt
	DisplayName string `json:"display_name,omitempty"`

	// Description provides context about the prompt's purpose
	Description string `json:"description,omitempty"`

	// Template is the prompt template text with variables in {variable} format
	Template string `json:"template"`

	// Variables is a list of variable names used in the template
	Variables []string `json:"variables,omitempty"`

	// Category organizes prompts by use case
	Category string `json:"category,omitempty"`

	// Tags provide additional metadata for organization and search
	Tags []string `json:"tags,omitempty"`

	// Version information
	VersionID   string `json:"version_id,omitempty"`
	VersionName string `json:"version_name,omitempty"`

	// Generation configuration for use with models
	GenerationConfig *genai.GenerationConfig `json:"generation_config,omitempty"`

	// Safety settings for content generation
	SafetySettings []*genai.SafetySetting `json:"safety_settings,omitempty"`

	// System instruction for model behavior
	SystemInstruction string `json:"system_instruction,omitempty"`

	// Cloud resource information
	ResourceName string `json:"resource_name,omitempty"`
	ProjectID    string `json:"project_id,omitempty"`
	Location     string `json:"location,omitempty"`

	// Metadata
	CreatedAt   time.Time         `json:"created_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
	CreatedBy   string            `json:"created_by,omitempty"`
	UpdatedBy   string            `json:"updated_by,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	// Schema validation
	InputSchema  *VariableSchema `json:"input_schema,omitempty"`
	OutputSchema *genai.Schema   `json:"output_schema,omitempty"`

	// Access control
	IsPublic    bool     `json:"is_public,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// VariableSchema defines the schema for template variables.
type VariableSchema struct {
	// Variables maps variable names to their schema definitions
	Variables map[string]*VariableDefinition `json:"variables,omitempty"`

	// Required lists variables that must be provided
	Required []string `json:"required,omitempty"`

	// Description of the overall variable schema
	Description string `json:"description,omitempty"`
}

// VariableDefinition defines the schema for a single template variable.
type VariableDefinition struct {
	// Type of the variable (string, number, boolean, object, array)
	Type string `json:"type"`

	// Description of what this variable represents
	Description string `json:"description,omitempty"`

	// Default value if not provided
	Default any `json:"default,omitempty"`

	// Format for string types (email, url, date-time, etc.)
	Format string `json:"format,omitempty"`

	// Constraints
	MinLength int    `json:"min_length,omitempty"`
	MaxLength int    `json:"max_length,omitempty"`
	Pattern   string `json:"pattern,omitempty"`

	// Enum values for categorical variables
	Enum []any `json:"enum,omitempty"`

	// For array types
	Items *VariableDefinition `json:"items,omitempty"`

	// For object types
	Properties map[string]*VariableDefinition `json:"properties,omitempty"`
}

// PromptVersion represents a specific version of a prompt.
type PromptVersion struct {
	// Version identification
	VersionID   string `json:"version_id"`
	VersionName string `json:"version_name,omitempty"`
	PromptID    string `json:"prompt_id"`

	// Version content
	Template          string                  `json:"template"`
	Variables         []string                `json:"variables,omitempty"`
	GenerationConfig  *genai.GenerationConfig `json:"generation_config,omitempty"`
	SafetySettings    []*genai.SafetySetting  `json:"safety_settings,omitempty"`
	SystemInstruction string                  `json:"system_instruction,omitempty"`

	// Version metadata
	Description string         `json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	CreatedBy   string         `json:"created_by,omitempty"`
	IsActive    bool           `json:"is_active"`
	Changelog   string         `json:"changelog,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`

	// Relationships
	ParentVersionID string `json:"parent_version_id,omitempty"`
	BranchName      string `json:"branch_name,omitempty"`

	// Performance metrics
	Usage VersionUsage `json:"usage,omitempty"`
}

// VersionUsage tracks usage statistics for a prompt version.
type VersionUsage struct {
	TotalCalls      int64     `json:"total_calls"`
	UniqueUsers     int64     `json:"unique_users"`
	LastUsed        time.Time `json:"last_used,omitempty"`
	AverageLatency  float64   `json:"average_latency"`
	ErrorRate       float64   `json:"error_rate"`
	TokensGenerated int64     `json:"tokens_generated"`
}

// CreatePromptRequest represents a request to create a new prompt.
type CreatePromptRequest struct {
	// Prompt data
	Prompt *Prompt `json:"prompt"`

	// Options
	CreateVersion bool   `json:"create_version,omitempty"`
	VersionName   string `json:"version_name,omitempty"`

	// Validation options
	ValidateTemplate bool `json:"validate_template,omitempty"`
	DryRun           bool `json:"dry_run,omitempty"`
}

// UpdatePromptRequest represents a request to update an existing prompt.
type UpdatePromptRequest struct {
	// Prompt data
	Prompt *Prompt `json:"prompt"`

	// Update options
	CreateNewVersion bool   `json:"create_new_version,omitempty"`
	VersionName      string `json:"version_name,omitempty"`
	Changelog        string `json:"changelog,omitempty"`

	// Validation
	ValidateTemplate bool `json:"validate_template,omitempty"`

	// Concurrency control
	IfMatchETag string `json:"if_match_etag,omitempty"`
}

// GetPromptRequest represents a request to retrieve a prompt.
type GetPromptRequest struct {
	// Identification
	PromptID string `json:"prompt_id,omitempty"`
	Name     string `json:"name,omitempty"`

	// Version selection
	VersionID string `json:"version_id,omitempty"`
	Latest    bool   `json:"latest,omitempty"`

	// Response options
	IncludeVersions bool `json:"include_versions,omitempty"`
	IncludeUsage    bool `json:"include_usage,omitempty"`
}

// ListPromptsRequest represents a request to list prompts.
type ListPromptsRequest struct {
	// Pagination
	PageSize  int32  `json:"page_size,omitempty"`
	PageToken string `json:"page_token,omitempty"`

	// Filtering
	Category      string    `json:"category,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	CreatedBy     string    `json:"created_by,omitempty"`
	CreatedAfter  time.Time `json:"created_after,omitempty"`
	CreatedBefore time.Time `json:"created_before,omitempty"`
	IsPublic      *bool     `json:"is_public,omitempty"`

	// Search
	Query        string   `json:"query,omitempty"`
	SearchFields []string `json:"search_fields,omitempty"`

	// Sorting
	OrderBy   string `json:"order_by,omitempty"`
	OrderDesc bool   `json:"order_desc,omitempty"`

	// Response options
	IncludeVersions bool `json:"include_versions,omitempty"`
}

// ListPromptsResponse represents a response from listing prompts.
type ListPromptsResponse struct {
	Prompts       []*Prompt `json:"prompts"`
	NextPageToken string    `json:"next_page_token,omitempty"`
	TotalSize     int32     `json:"total_size,omitempty"`
}

// DeletePromptRequest represents a request to delete a prompt.
type DeletePromptRequest struct {
	// Identification
	PromptID string `json:"prompt_id,omitempty"`
	Name     string `json:"name,omitempty"`

	// Options
	Force          bool `json:"force,omitempty"`
	DeleteVersions bool `json:"delete_versions,omitempty"`

	// Concurrency control
	IfMatchETag string `json:"if_match_etag,omitempty"`
}

// ApplyTemplateRequest represents a request to apply variables to a template.
type ApplyTemplateRequest struct {
	// Template identification
	PromptID  string `json:"prompt_id,omitempty"`
	Name      string `json:"name,omitempty"`
	VersionID string `json:"version_id,omitempty"`
	Template  string `json:"template,omitempty"`

	// Variables to substitute
	Variables map[string]any `json:"variables"`

	// Options
	ValidateVariables bool `json:"validate_variables,omitempty"`
	StrictMode        bool `json:"strict_mode,omitempty"`
}

// ApplyTemplateResponse represents the result of template variable application.
type ApplyTemplateResponse struct {
	// Result
	Content string `json:"content"`

	// Applied variables
	AppliedVariables map[string]any `json:"applied_variables"`

	// Validation results
	MissingVariables []string `json:"missing_variables,omitempty"`
	UnusedVariables  []string `json:"unused_variables,omitempty"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

// CreateVersionRequest represents a request to create a new prompt version.
type CreateVersionRequest struct {
	PromptID    string  `json:"prompt_id"`
	Prompt      *Prompt `json:"prompt"`
	VersionName string  `json:"version_name,omitempty"`
	Changelog   string  `json:"changelog,omitempty"`
	BranchName  string  `json:"branch_name,omitempty"`
}

// ListVersionsRequest represents a request to list prompt versions.
type ListVersionsRequest struct {
	PromptID      string    `json:"prompt_id"`
	PageSize      int32     `json:"page_size,omitempty"`
	PageToken     string    `json:"page_token,omitempty"`
	IncludeUsage  bool      `json:"include_usage,omitempty"`
	CreatedAfter  time.Time `json:"created_after,omitempty"`
	CreatedBefore time.Time `json:"created_before,omitempty"`
	BranchName    string    `json:"branch_name,omitempty"`
}

// ListVersionsResponse represents a response from listing prompt versions.
type ListVersionsResponse struct {
	Versions      []*PromptVersion `json:"versions"`
	NextPageToken string           `json:"next_page_token,omitempty"`
	TotalSize     int32            `json:"total_size,omitempty"`
}

// RestoreVersionRequest represents a request to restore a prompt version.
type RestoreVersionRequest struct {
	PromptID       string `json:"prompt_id"`
	VersionID      string `json:"version_id"`
	NewVersionName string `json:"new_version_name,omitempty"`
	Changelog      string `json:"changelog,omitempty"`
}

// BatchCreatePromptsRequest represents a request to create multiple prompts.
type BatchCreatePromptsRequest struct {
	Prompts         []*Prompt `json:"prompts"`
	CreateVersions  bool      `json:"create_versions,omitempty"`
	ValidateAll     bool      `json:"validate_all,omitempty"`
	ContinueOnError bool      `json:"continue_on_error,omitempty"`
}

// BatchCreatePromptsResponse represents a response from batch prompt creation.
type BatchCreatePromptsResponse struct {
	Results   []*BatchOperationResult `json:"results"`
	Succeeded int32                   `json:"succeeded"`
	Failed    int32                   `json:"failed"`
}

// BatchOperationResult represents the result of a single batch operation.
type BatchOperationResult struct {
	Index   int32   `json:"index"`
	Prompt  *Prompt `json:"prompt,omitempty"`
	Error   string  `json:"error,omitempty"`
	Success bool    `json:"success"`
}

// ExportPromptsRequest represents a request to export prompts.
type ExportPromptsRequest struct {
	PromptIDs       []string `json:"prompt_ids,omitempty"`
	Names           []string `json:"names,omitempty"`
	Category        string   `json:"category,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	IncludeVersions bool     `json:"include_versions,omitempty"`
	Format          string   `json:"format,omitempty"` // json, yaml, csv
}

// ExportPromptsResponse represents a response from exporting prompts.
type ExportPromptsResponse struct {
	Data     []byte         `json:"data"`
	Format   string         `json:"format"`
	Prompts  []*Prompt      `json:"prompts,omitempty"`
	Count    int32          `json:"count"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ImportPromptsRequest represents a request to import prompts.
type ImportPromptsRequest struct {
	Data            []byte `json:"data"`
	Format          string `json:"format,omitempty"` // json, yaml, csv
	Overwrite       bool   `json:"overwrite,omitempty"`
	CreateVersions  bool   `json:"create_versions,omitempty"`
	ValidateAll     bool   `json:"validate_all,omitempty"`
	ContinueOnError bool   `json:"continue_on_error,omitempty"`
}

// SearchPromptsRequest represents a request to search prompts.
type SearchPromptsRequest struct {
	// Search query
	Query        string   `json:"query"`
	SearchFields []string `json:"search_fields,omitempty"` // name, description, template, tags

	// Filters
	Category      string    `json:"category,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	CreatedBy     string    `json:"created_by,omitempty"`
	CreatedAfter  time.Time `json:"created_after,omitempty"`
	CreatedBefore time.Time `json:"created_before,omitempty"`
	IsPublic      *bool     `json:"is_public,omitempty"`

	// Search options
	FuzzySearch   bool    `json:"fuzzy_search,omitempty"`
	CaseSensitive bool    `json:"case_sensitive,omitempty"`
	MinScore      float64 `json:"min_score,omitempty"`

	// Pagination and sorting
	PageSize  int32  `json:"page_size,omitempty"`
	PageToken string `json:"page_token,omitempty"`
	OrderBy   string `json:"order_by,omitempty"`
	OrderDesc bool   `json:"order_desc,omitempty"`
}

// SearchPromptsResponse represents a response from searching prompts.
type SearchPromptsResponse struct {
	Results       []*SearchResult `json:"results"`
	NextPageToken string          `json:"next_page_token,omitempty"`
	TotalSize     int32           `json:"total_size,omitempty"`
	SearchTime    time.Duration   `json:"search_time,omitempty"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	Prompt      *Prompt             `json:"prompt"`
	Score       float64             `json:"score"`
	Highlights  map[string][]string `json:"highlights,omitempty"`
	MatchFields []string            `json:"match_fields,omitempty"`
}

// TemplateValidationResult represents the result of template validation.
type TemplateValidationResult struct {
	IsValid        bool     `json:"is_valid"`
	Errors         []string `json:"errors,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	DetectedVars   []string `json:"detected_variables,omitempty"`
	UndeclaredVars []string `json:"undeclared_variables,omitempty"`
	UnusedVars     []string `json:"unused_variables,omitempty"`
}

// PromptMetrics represents performance and usage metrics for a prompt.
type PromptMetrics struct {
	PromptID        string    `json:"prompt_id"`
	TotalCalls      int64     `json:"total_calls"`
	UniqueUsers     int64     `json:"unique_users"`
	LastUsed        time.Time `json:"last_used,omitempty"`
	AverageLatency  float64   `json:"average_latency"`
	ErrorRate       float64   `json:"error_rate"`
	TokensGenerated int64     `json:"tokens_generated"`

	// Time-series data
	DailyCalls   []int64 `json:"daily_calls,omitempty"`
	WeeklyCalls  []int64 `json:"weekly_calls,omitempty"`
	MonthlyCalls []int64 `json:"monthly_calls,omitempty"`

	// Version breakdown
	VersionMetrics map[string]*VersionUsage `json:"version_metrics,omitempty"`
}

// PromptAnalytics represents detailed analytics for prompt usage.
type PromptAnalytics struct {
	PromptID       string           `json:"prompt_id"`
	TimeRange      TimeRange        `json:"time_range"`
	Metrics        *PromptMetrics   `json:"metrics"`
	TopVariables   []VariableUsage  `json:"top_variables,omitempty"`
	ErrorBreakdown map[string]int64 `json:"error_breakdown,omitempty"`
	UserSegments   []UserSegment    `json:"user_segments,omitempty"`
}

// TimeRange represents a time range for analytics.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// VariableUsage represents usage statistics for a template variable.
type VariableUsage struct {
	Name         string           `json:"name"`
	Frequency    int64            `json:"frequency"`
	UniqueValues int64            `json:"unique_values"`
	TopValues    map[string]int64 `json:"top_values,omitempty"`
}

// UserSegment represents usage data for a user segment.
type UserSegment struct {
	Segment    string  `json:"segment"`
	UserCount  int64   `json:"user_count"`
	CallCount  int64   `json:"call_count"`
	ErrorRate  float64 `json:"error_rate"`
	AvgLatency float64 `json:"avg_latency"`
}

// PromptCollection represents a curated collection of related prompts.
type PromptCollection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Description string    `json:"description,omitempty"`
	PromptIDs   []string  `json:"prompt_ids"`
	Category    string    `json:"category,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	IsPublic    bool      `json:"is_public,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by,omitempty"`
}

// ValidationMode specifies how strict template validation should be.
type ValidationMode string

const (
	ValidationModeStrict ValidationMode = "strict" // All variables must be declared and used
	ValidationModeWarn   ValidationMode = "warn"   // Warnings for undeclared/unused variables
	ValidationModeLoose  ValidationMode = "loose"  // Allow undeclared variables
	ValidationModeNone   ValidationMode = "none"   // No validation
)

// TemplateEngine specifies which template engine to use.
type TemplateEngine string

const (
	TemplateEngineSimple   TemplateEngine = "simple"   // Simple {variable} substitution
	TemplateEngineAdvanced TemplateEngine = "advanced" // Advanced templating with conditionals
	TemplateEngineJinja    TemplateEngine = "jinja"    // Jinja2-style templating
)

// PromptStatus represents the current status of a prompt.
type PromptStatus string

const (
	PromptStatusDraft    PromptStatus = "draft"
	PromptStatusActive   PromptStatus = "active"
	PromptStatusArchived PromptStatus = "archived"
	PromptStatusDeleted  PromptStatus = "deleted"
)

// CloudResourceInfo contains information about the cloud storage of a prompt.
type CloudResourceInfo struct {
	ResourceName string            `json:"resource_name"`
	ProjectID    string            `json:"project_id"`
	Location     string            `json:"location"`
	ResourceURI  string            `json:"resource_uri,omitempty"`
	Etag         string            `json:"etag,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
}
