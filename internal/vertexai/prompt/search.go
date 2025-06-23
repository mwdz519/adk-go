// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"
)

// Search and batch operations for the prompts service

// SearchPrompts searches for prompts based on various criteria.
func (s *service) SearchPrompts(ctx context.Context, req *SearchPromptsRequest) (*SearchPromptsResponse, error) {
	if req.Query == "" {
		return nil, NewInvalidRequestError("query", "search query cannot be empty")
	}

	// Start performance tracking
	tracker := s.metrics.StartOperation("search_prompts")
	defer tracker.Finish()

	startTime := time.Now()

	// Load prompts based on filters
	prompts, err := s.loadPromptsForSearch(ctx, req)
	if err != nil {
		tracker.FinishWithError("cloud")
		return nil, fmt.Errorf("failed to load prompts for search: %w", err)
	}

	// Perform search
	searchResults := s.performSearch(prompts, req)

	// Sort results by score
	sort.Slice(searchResults, func(i, j int) bool {
		if req.OrderDesc {
			return searchResults[i].Score > searchResults[j].Score
		}
		return searchResults[i].Score < searchResults[j].Score
	})

	// Apply pagination
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	startIndex := 0
	if req.PageToken != "" {
		startIndex = s.parsePageToken(req.PageToken)
	}

	endIndex := min(startIndex+int(pageSize), len(searchResults))

	var paginatedResults []*SearchResult
	var nextPageToken string

	if startIndex < len(searchResults) {
		paginatedResults = searchResults[startIndex:endIndex]

		if endIndex < len(searchResults) {
			nextPageToken = s.generatePageToken(endIndex)
		}
	}

	searchTime := time.Since(startTime)

	response := &SearchPromptsResponse{
		Results:       paginatedResults,
		NextPageToken: nextPageToken,
		TotalSize:     int32(len(searchResults)),
		SearchTime:    searchTime,
	}

	s.logger.InfoContext(ctx, "Prompt search completed",
		slog.String("query", req.Query),
		slog.Int("total_results", len(searchResults)),
		slog.Int("returned_results", len(paginatedResults)),
		slog.Duration("search_time", searchTime),
	)

	return response, nil
}

// Batch operations

// BatchCreatePrompts creates multiple prompts in a single operation.
func (s *service) BatchCreatePrompts(ctx context.Context, req *BatchCreatePromptsRequest) (*BatchCreatePromptsResponse, error) {
	if len(req.Prompts) == 0 {
		return nil, NewInvalidRequestError("prompts", "cannot be empty")
	}

	results := make([]*BatchOperationResult, len(req.Prompts))
	succeeded := int32(0)
	failed := int32(0)

	for i, prompt := range req.Prompts {
		result := &BatchOperationResult{
			Index: int32(i),
		}

		// Validate prompt if requested
		if req.ValidateAll {
			if err := s.validatePromptTemplate(prompt); err != nil {
				result.Error = err.Error()
				result.Success = false
				failed++
				results[i] = result

				if !req.ContinueOnError {
					break
				}
				continue
			}
		}

		// Create the prompt
		createReq := &CreatePromptRequest{
			Prompt:           prompt,
			CreateVersion:    req.CreateVersions,
			ValidateTemplate: req.ValidateAll,
		}

		createdPrompt, err := s.CreatePrompt(ctx, createReq)
		if err != nil {
			result.Error = err.Error()
			result.Success = false
			failed++

			if !req.ContinueOnError {
				results[i] = result
				break
			}
		} else {
			result.Prompt = createdPrompt
			result.Success = true
			succeeded++
		}

		results[i] = result
	}

	response := &BatchCreatePromptsResponse{
		Results:   results,
		Succeeded: succeeded,
		Failed:    failed,
	}

	s.logger.InfoContext(ctx, "Batch prompt creation completed",
		slog.Int("total_prompts", len(req.Prompts)),
		slog.Int("succeeded", int(succeeded)),
		slog.Int("failed", int(failed)),
	)

	return response, nil
}

// ExportPrompts exports prompts to various formats.
func (s *service) ExportPrompts(ctx context.Context, req *ExportPromptsRequest) (*ExportPromptsResponse, error) {
	// Determine which prompts to export
	var prompts []*Prompt
	var err error

	if len(req.PromptIDs) > 0 {
		prompts, err = s.getPromptsByIDs(ctx, req.PromptIDs, req.IncludeVersions)
	} else if len(req.Names) > 0 {
		prompts, err = s.getPromptsByNames(ctx, req.Names, req.IncludeVersions)
	} else {
		// Export by category/tags
		listReq := &ListPromptsRequest{
			Category:        req.Category,
			Tags:            req.Tags,
			IncludeVersions: req.IncludeVersions,
			PageSize:        1000, // Large page size for export
		}
		listResp, listErr := s.ListPrompts(ctx, listReq)
		if listErr != nil {
			return nil, fmt.Errorf("failed to list prompts for export: %w", listErr)
		}
		prompts = listResp.Prompts
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve prompts for export: %w", err)
	}

	// Export to specified format
	format := req.Format
	if format == "" {
		format = "json"
	}

	data, err := s.exportToFormat(prompts, format)
	if err != nil {
		return nil, fmt.Errorf("failed to export prompts to format %s: %w", format, err)
	}

	response := &ExportPromptsResponse{
		Data:    data,
		Format:  format,
		Prompts: prompts,
		Count:   int32(len(prompts)),
		Metadata: map[string]any{
			"exported_at":      time.Now().Format(time.RFC3339),
			"exported_by":      "prompts-service",
			"total_prompts":    len(prompts),
			"include_versions": req.IncludeVersions,
		},
	}

	s.logger.InfoContext(ctx, "Prompts exported successfully",
		slog.Int("prompt_count", len(prompts)),
		slog.String("format", format),
		slog.Bool("include_versions", req.IncludeVersions),
	)

	return response, nil
}

// ImportPrompts imports prompts from various formats.
func (s *service) ImportPrompts(ctx context.Context, req *ImportPromptsRequest) (*BatchCreatePromptsResponse, error) {
	if len(req.Data) == 0 {
		return nil, NewInvalidRequestError("data", "cannot be empty")
	}

	// Parse data based on format
	format := req.Format
	if format == "" {
		format = "json"
	}

	prompts, err := s.parseImportData(req.Data, format)
	if err != nil {
		return nil, fmt.Errorf("failed to parse import data: %w", err)
	}

	// Create batch request
	batchReq := &BatchCreatePromptsRequest{
		Prompts:         prompts,
		CreateVersions:  req.CreateVersions,
		ValidateAll:     req.ValidateAll,
		ContinueOnError: req.ContinueOnError,
	}

	// Handle overwrites
	if req.Overwrite {
		for _, prompt := range prompts {
			// Check if prompt exists and delete if overwrite is enabled
			existing, err := s.GetPrompt(ctx, &GetPromptRequest{Name: prompt.Name})
			if err == nil && existing != nil {
				if err := s.DeletePrompt(ctx, &DeletePromptRequest{
					PromptID: existing.ID,
					Force:    true,
				}); err != nil {
					s.logger.WarnContext(ctx, "Failed to delete existing prompt for overwrite",
						slog.String("prompt_name", prompt.Name),
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}

	// Perform batch creation
	response, err := s.BatchCreatePrompts(ctx, batchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to import prompts: %w", err)
	}

	s.logger.InfoContext(ctx, "Prompts imported successfully",
		slog.Int("prompt_count", len(prompts)),
		slog.String("format", format),
		slog.Bool("overwrite", req.Overwrite),
		slog.Int("succeeded", int(response.Succeeded)),
		slog.Int("failed", int(response.Failed)),
	)

	return response, nil
}

// Helper methods for search functionality

// loadPromptsForSearch loads prompts that match the basic filters.
func (s *service) loadPromptsForSearch(ctx context.Context, req *SearchPromptsRequest) ([]*Prompt, error) {
	listReq := &ListPromptsRequest{
		Category:      req.Category,
		Tags:          req.Tags,
		CreatedBy:     req.CreatedBy,
		CreatedAfter:  req.CreatedAfter,
		CreatedBefore: req.CreatedBefore,
		IsPublic:      req.IsPublic,
		PageSize:      1000, // Large page size for comprehensive search
	}

	listResp, err := s.ListPrompts(ctx, listReq)
	if err != nil {
		return nil, err
	}

	return listResp.Prompts, nil
}

// performSearch performs the actual search logic.
func (s *service) performSearch(prompts []*Prompt, req *SearchPromptsRequest) []*SearchResult {
	var results []*SearchResult
	query := strings.ToLower(req.Query)

	searchFields := req.SearchFields
	if len(searchFields) == 0 {
		searchFields = []string{"name", "description", "template", "tags"}
	}

	for _, prompt := range prompts {
		score, highlights, matchFields := s.calculateSearchScore(prompt, query, searchFields, req)

		if score >= req.MinScore {
			results = append(results, &SearchResult{
				Prompt:      prompt,
				Score:       score,
				Highlights:  highlights,
				MatchFields: matchFields,
			})
		}
	}

	return results
}

// calculateSearchScore calculates the search score for a prompt.
func (s *service) calculateSearchScore(prompt *Prompt, query string, searchFields []string, req *SearchPromptsRequest) (float64, map[string][]string, []string) {
	var totalScore float64
	highlights := make(map[string][]string)
	var matchFields []string

	for _, field := range searchFields {
		var fieldText string
		var fieldWeight float64 = 1.0

		switch field {
		case "name":
			fieldText = prompt.Name
			fieldWeight = 3.0 // Name matches are more important
		case "display_name":
			fieldText = prompt.DisplayName
			fieldWeight = 2.5
		case "description":
			fieldText = prompt.Description
			fieldWeight = 2.0
		case "template":
			fieldText = prompt.Template
			fieldWeight = 1.5
		case "tags":
			fieldText = strings.Join(prompt.Tags, " ")
			fieldWeight = 2.0
		case "category":
			fieldText = prompt.Category
			fieldWeight = 1.8
		default:
			continue
		}

		if fieldText == "" {
			continue
		}

		fieldScore, fieldHighlights := s.scoreFieldMatch(fieldText, query, req)
		if fieldScore > 0 {
			totalScore += fieldScore * fieldWeight
			if len(fieldHighlights) > 0 {
				highlights[field] = fieldHighlights
				matchFields = append(matchFields, field)
			}
		}
	}

	return totalScore, highlights, matchFields
}

// scoreFieldMatch scores how well a field matches the query.
func (s *service) scoreFieldMatch(fieldText, query string, req *SearchPromptsRequest) (float64, []string) {
	if !req.CaseSensitive {
		fieldText = strings.ToLower(fieldText)
	}

	var score float64
	var highlights []string

	if req.FuzzySearch {
		// Simple fuzzy matching (can be enhanced with more sophisticated algorithms)
		score = s.fuzzyMatchScore(fieldText, query)
		if score > 0 {
			highlights = []string{s.extractHighlight(fieldText, query)}
		}
	} else {
		// Exact substring matching
		if strings.Contains(fieldText, query) {
			score = 1.0
			highlights = []string{s.extractHighlight(fieldText, query)}
		}
	}

	return score, highlights
}

// fuzzyMatchScore calculates a fuzzy match score.
func (s *service) fuzzyMatchScore(text, query string) float64 {
	// Simple fuzzy scoring based on character overlap
	// In a real implementation, you might use more sophisticated algorithms like Levenshtein distance

	if strings.Contains(text, query) {
		return 1.0
	}

	// Check for partial matches
	queryLen := len(query)
	if queryLen < 3 {
		return 0.0 // Too short for fuzzy matching
	}

	matches := 0
	for i := 0; i <= len(query)-3; i++ {
		substr := query[i : i+3]
		if strings.Contains(text, substr) {
			matches++
		}
	}

	return float64(matches) / float64(queryLen-2)
}

// extractHighlight extracts highlighted text around a match.
func (s *service) extractHighlight(text, query string) string {
	index := strings.Index(text, query)
	if index == -1 {
		return ""
	}

	start := index
	end := index + len(query)

	// Expand context around the match
	contextSize := 50
	if start > contextSize {
		start = index - contextSize
	} else {
		start = 0
	}

	if end+contextSize < len(text) {
		end = index + len(query) + contextSize
	} else {
		end = len(text)
	}

	highlight := text[start:end]

	// Add ellipsis if truncated
	if start > 0 {
		highlight = "..." + highlight
	}
	if end < len(text) {
		highlight = highlight + "..."
	}

	return highlight
}

// Helper methods for batch operations

// getPromptsByIDs retrieves multiple prompts by their IDs.
func (s *service) getPromptsByIDs(ctx context.Context, promptIDs []string, includeVersions bool) ([]*Prompt, error) {
	var prompts []*Prompt

	for _, promptID := range promptIDs {
		prompt, err := s.GetPrompt(ctx, &GetPromptRequest{
			PromptID:        promptID,
			IncludeVersions: includeVersions,
		})
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to retrieve prompt by ID",
				slog.String("prompt_id", promptID),
				slog.String("error", err.Error()),
			)
			continue
		}
		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// getPromptsByNames retrieves multiple prompts by their names.
func (s *service) getPromptsByNames(ctx context.Context, names []string, includeVersions bool) ([]*Prompt, error) {
	var prompts []*Prompt

	for _, name := range names {
		prompt, err := s.GetPrompt(ctx, &GetPromptRequest{
			Name:            name,
			IncludeVersions: includeVersions,
		})
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to retrieve prompt by name",
				slog.String("name", name),
				slog.String("error", err.Error()),
			)
			continue
		}
		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// exportToFormat exports prompts to the specified format.
func (s *service) exportToFormat(prompts []*Prompt, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		return s.exportToJSON(prompts)
	case "yaml":
		return s.exportToYAML(prompts)
	case "csv":
		return s.exportToCSV(prompts)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// parseImportData parses import data based on format.
func (s *service) parseImportData(data []byte, format string) ([]*Prompt, error) {
	switch strings.ToLower(format) {
	case "json":
		return s.parseJSONImport(data)
	case "yaml":
		return s.parseYAMLImport(data)
	case "csv":
		return s.parseCSVImport(data)
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}
}

// Placeholder implementations for export/import formats
// These would be implemented with proper JSON/YAML/CSV libraries

func (s *service) exportToJSON(prompts []*Prompt) ([]byte, error) {
	// This would implement JSON marshaling
	return []byte("{}"), nil
}

func (s *service) exportToYAML(prompts []*Prompt) ([]byte, error) {
	// This would implement YAML marshaling
	return []byte(""), nil
}

func (s *service) exportToCSV(prompts []*Prompt) ([]byte, error) {
	// This would implement CSV marshaling
	return []byte(""), nil
}

func (s *service) parseJSONImport(data []byte) ([]*Prompt, error) {
	// This would implement JSON unmarshaling
	return []*Prompt{}, nil
}

func (s *service) parseYAMLImport(data []byte) ([]*Prompt, error) {
	// This would implement YAML unmarshaling
	return []*Prompt{}, nil
}

func (s *service) parseCSVImport(data []byte) ([]*Prompt, error) {
	// This would implement CSV parsing
	return []*Prompt{}, nil
}
