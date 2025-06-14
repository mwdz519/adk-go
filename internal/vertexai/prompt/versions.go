// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Version management methods for the prompts service

// CreateVersion creates a new version of an existing prompt.
func (s *Service) CreateVersion(ctx context.Context, req *CreateVersionRequest) (*PromptVersion, error) {
	if req.PromptID == "" {
		return nil, NewInvalidRequestError("prompt_id", "cannot be empty")
	}
	if req.Prompt == nil {
		return nil, NewInvalidRequestError("prompt", "cannot be nil")
	}

	// Validate the prompt exists
	_, err := s.GetPrompt(ctx, &GetPromptRequest{
		PromptID: req.PromptID,
	})
	if err != nil {
		return nil, err
	}

	// Validate the template
	if err := s.validatePromptTemplate(req.Prompt); err != nil {
		return nil, err
	}

	// Create the version
	version, err := s.createPromptVersion(ctx, req.PromptID, req.Prompt, req.VersionName, req.Changelog)
	if err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "Prompt version created successfully",
		slog.String("prompt_id", req.PromptID),
		slog.String("version_id", version.VersionID),
		slog.String("version_name", version.VersionName),
	)

	return version, nil
}

// GetVersion retrieves a specific version of a prompt.
func (s *Service) GetVersion(ctx context.Context, promptID, versionID string) (*PromptVersion, error) {
	if promptID == "" {
		return nil, NewInvalidRequestError("prompt_id", "cannot be empty")
	}
	if versionID == "" {
		return nil, NewInvalidRequestError("version_id", "cannot be empty")
	}

	version, err := s.getPromptVersion(ctx, promptID, versionID)
	if err != nil {
		return nil, err
	}

	return version, nil
}

// ListVersions lists all versions of a prompt.
func (s *Service) ListVersions(ctx context.Context, req *ListVersionsRequest) (*ListVersionsResponse, error) {
	if req.PromptID == "" {
		return nil, NewInvalidRequestError("prompt_id", "cannot be empty")
	}

	versions, err := s.listPromptVersions(ctx, req.PromptID)
	if err != nil {
		return nil, err
	}

	// Apply filtering if specified
	filteredVersions := s.filterVersions(versions, req)

	// Apply pagination
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	startIndex := 0
	if req.PageToken != "" {
		// Parse page token to get start index
		// In a real implementation, this would be a proper pagination token
		startIndex = s.parsePageToken(req.PageToken)
	}

	endIndex := startIndex + int(pageSize)
	if endIndex > len(filteredVersions) {
		endIndex = len(filteredVersions)
	}

	var resultVersions []*PromptVersion
	var nextPageToken string

	if startIndex < len(filteredVersions) {
		resultVersions = filteredVersions[startIndex:endIndex]

		if endIndex < len(filteredVersions) {
			nextPageToken = s.generatePageToken(endIndex)
		}
	}

	return &ListVersionsResponse{
		Versions:      resultVersions,
		NextPageToken: nextPageToken,
		TotalSize:     int32(len(filteredVersions)),
	}, nil
}

// RestoreVersion restores a previous version as the current version.
func (s *Service) RestoreVersion(ctx context.Context, req *RestoreVersionRequest) (*PromptVersion, error) {
	if req.PromptID == "" {
		return nil, NewInvalidRequestError("prompt_id", "cannot be empty")
	}
	if req.VersionID == "" {
		return nil, NewInvalidRequestError("version_id", "cannot be empty")
	}

	// Get the version to restore
	versionToRestore, err := s.getPromptVersion(ctx, req.PromptID, req.VersionID)
	if err != nil {
		return nil, err
	}

	// Create a new prompt based on the version to restore
	restoredPrompt := &Prompt{
		ID:                versionToRestore.PromptID,
		Template:          versionToRestore.Template,
		Variables:         versionToRestore.Variables,
		GenerationConfig:  versionToRestore.GenerationConfig,
		SafetySettings:    versionToRestore.SafetySettings,
		SystemInstruction: versionToRestore.SystemInstruction,
		UpdatedAt:         time.Now(),
	}

	// Create a new version from the restored content
	newVersionName := req.NewVersionName
	if newVersionName == "" {
		newVersionName = fmt.Sprintf("restored-from-%s", req.VersionID)
	}

	changelog := req.Changelog
	if changelog == "" {
		changelog = fmt.Sprintf("Restored from version %s", req.VersionID)
	}

	newVersion, err := s.createPromptVersion(ctx, req.PromptID, restoredPrompt, newVersionName, changelog)
	if err != nil {
		return nil, fmt.Errorf("failed to create restored version: %w", err)
	}

	// Update the prompt to use the new version
	_, err = s.UpdatePrompt(ctx, &UpdatePromptRequest{
		Prompt: restoredPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update prompt with restored content: %w", err)
	}

	s.metrics.IncrementVersionRestored()

	s.logger.InfoContext(ctx, "Prompt version restored successfully",
		slog.String("prompt_id", req.PromptID),
		slog.String("restored_from_version", req.VersionID),
		slog.String("new_version_id", newVersion.VersionID),
	)

	return newVersion, nil
}

// DeleteVersion deletes a specific version of a prompt.
func (s *Service) DeleteVersion(ctx context.Context, promptID, versionID string) error {
	if promptID == "" {
		return NewInvalidRequestError("prompt_id", "cannot be empty")
	}
	if versionID == "" {
		return NewInvalidRequestError("version_id", "cannot be empty")
	}

	// Check if version exists
	_, err := s.getPromptVersion(ctx, promptID, versionID)
	if err != nil {
		return err
	}

	// Delete the version
	if err := s.deletePromptVersion(ctx, promptID, versionID); err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	s.logger.InfoContext(ctx, "Prompt version deleted successfully",
		slog.String("prompt_id", promptID),
		slog.String("version_id", versionID),
	)

	return nil
}

// Internal version management methods

// createPromptVersion creates a new version of a prompt.
func (s *Service) createPromptVersion(ctx context.Context, promptID string, prompt *Prompt, versionName, changelog string) (*PromptVersion, error) {
	now := time.Now()

	version := &PromptVersion{
		VersionID:         s.generateVersionID(),
		VersionName:       versionName,
		PromptID:          promptID,
		Template:          prompt.Template,
		Variables:         prompt.Variables,
		GenerationConfig:  prompt.GenerationConfig,
		SafetySettings:    prompt.SafetySettings,
		SystemInstruction: prompt.SystemInstruction,
		Description:       prompt.Description,
		CreatedAt:         now,
		IsActive:          true,
		Changelog:         changelog,
	}

	// Save version to cloud storage (simulated)
	if err := s.saveVersionToCloud(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to save version to cloud: %w", err)
	}

	// Cache the version
	s.cacheVersion(version)

	s.metrics.IncrementVersionCreated()

	return version, nil
}

// getPromptVersion retrieves a specific version.
func (s *Service) getPromptVersion(ctx context.Context, promptID, versionID string) (*PromptVersion, error) {
	// Try cache first
	if version := s.getCachedVersion(promptID, versionID); version != nil {
		return version, nil
	}

	// Load from cloud storage (simulated)
	version, err := s.loadVersionFromCloud(ctx, promptID, versionID)
	if err != nil {
		return nil, NewVersionNotFoundError(promptID, versionID)
	}

	// Cache the version
	s.cacheVersion(version)

	return version, nil
}

// listPromptVersions lists all versions of a prompt.
func (s *Service) listPromptVersions(ctx context.Context, promptID string) ([]*PromptVersion, error) {
	// Load from cloud storage (simulated)
	versions, err := s.listVersionsFromCloud(ctx, promptID)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions from cloud: %w", err)
	}

	// Cache the versions
	for _, version := range versions {
		s.cacheVersion(version)
	}

	return versions, nil
}

// deletePromptVersion deletes a version.
func (s *Service) deletePromptVersion(ctx context.Context, promptID, versionID string) error {
	// Delete from cloud storage (simulated)
	if err := s.deleteVersionFromCloud(ctx, promptID, versionID); err != nil {
		return err
	}

	// Remove from cache
	s.removeCachedVersion(promptID, versionID)

	return nil
}

// Version caching methods

// cacheVersion caches a prompt version.
func (s *Service) cacheVersion(version *PromptVersion) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	if s.versionCache[version.PromptID] == nil {
		s.versionCache[version.PromptID] = make([]*PromptVersion, 0)
	}

	// Check if version already exists in cache
	for i, cached := range s.versionCache[version.PromptID] {
		if cached.VersionID == version.VersionID {
			s.versionCache[version.PromptID][i] = version
			return
		}
	}

	// Add new version to cache
	s.versionCache[version.PromptID] = append(s.versionCache[version.PromptID], version)
}

// getCachedVersion retrieves a cached version.
func (s *Service) getCachedVersion(promptID, versionID string) *PromptVersion {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	versions, exists := s.versionCache[promptID]
	if !exists {
		return nil
	}

	for _, version := range versions {
		if version.VersionID == versionID {
			return version
		}
	}

	return nil
}

// removeCachedVersion removes a version from cache.
func (s *Service) removeCachedVersion(promptID, versionID string) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	versions, exists := s.versionCache[promptID]
	if !exists {
		return
	}

	for i, version := range versions {
		if version.VersionID == versionID {
			s.versionCache[promptID] = append(versions[:i], versions[i+1:]...)
			break
		}
	}
}

// Utility methods

// generateVersionID generates a unique version ID.
func (s *Service) generateVersionID() string {
	return fmt.Sprintf("version_%d", time.Now().UnixNano())
}

// promptFromVersion creates a prompt from a version.
func (s *Service) promptFromVersion(basePrompt *Prompt, version *PromptVersion) *Prompt {
	prompt := *basePrompt
	prompt.Template = version.Template
	prompt.Variables = version.Variables
	prompt.GenerationConfig = version.GenerationConfig
	prompt.SafetySettings = version.SafetySettings
	prompt.SystemInstruction = version.SystemInstruction
	prompt.VersionID = version.VersionID
	prompt.UpdatedAt = version.CreatedAt
	return &prompt
}

// filterVersions applies filtering to a list of versions.
func (s *Service) filterVersions(versions []*PromptVersion, req *ListVersionsRequest) []*PromptVersion {
	var filtered []*PromptVersion

	for _, version := range versions {
		// Apply filters
		if !req.CreatedAfter.IsZero() && version.CreatedAt.Before(req.CreatedAfter) {
			continue
		}
		if !req.CreatedBefore.IsZero() && version.CreatedAt.After(req.CreatedBefore) {
			continue
		}
		if req.BranchName != "" && version.BranchName != req.BranchName {
			continue
		}

		filtered = append(filtered, version)
	}

	return filtered
}

// parsePageToken parses a pagination token.
func (s *Service) parsePageToken(token string) int {
	// In a real implementation, this would properly decode the token
	return 0
}

// generatePageToken generates a pagination token.
func (s *Service) generatePageToken(offset int) string {
	// In a real implementation, this would properly encode the token
	return fmt.Sprintf("offset_%d", offset)
}

// getPromptMetrics retrieves usage metrics for a prompt (placeholder).
func (s *Service) getPromptMetrics(ctx context.Context, promptID string) (*PromptMetrics, error) {
	// This would implement actual metrics retrieval
	return &PromptMetrics{
		PromptID:    promptID,
		TotalCalls:  0,
		UniqueUsers: 0,
	}, nil
}

// Placeholder methods for cloud operations (to be implemented with actual Vertex AI APIs)

// saveVersionToCloud saves a version to cloud storage.
func (s *Service) saveVersionToCloud(ctx context.Context, version *PromptVersion) error {
	s.logger.InfoContext(ctx, "Saving version to cloud storage",
		slog.String("prompt_id", version.PromptID),
		slog.String("version_id", version.VersionID))
	return nil
}

// loadVersionFromCloud loads a version from cloud storage.
func (s *Service) loadVersionFromCloud(ctx context.Context, promptID, versionID string) (*PromptVersion, error) {
	s.logger.InfoContext(ctx, "Loading version from cloud storage",
		slog.String("prompt_id", promptID),
		slog.String("version_id", versionID))
	return nil, ErrVersionNotFound
}

// listVersionsFromCloud lists versions from cloud storage.
func (s *Service) listVersionsFromCloud(ctx context.Context, promptID string) ([]*PromptVersion, error) {
	s.logger.InfoContext(ctx, "Listing versions from cloud storage",
		slog.String("prompt_id", promptID))
	return []*PromptVersion{}, nil
}

// deleteVersionFromCloud deletes a version from cloud storage.
func (s *Service) deleteVersionFromCloud(ctx context.Context, promptID, versionID string) error {
	s.logger.InfoContext(ctx, "Deleting version from cloud storage",
		slog.String("prompt_id", promptID),
		slog.String("version_id", versionID))
	return nil
}
