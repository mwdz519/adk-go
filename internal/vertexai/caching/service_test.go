// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package caching

import (
	"log/slog"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/option"
	"google.golang.org/genai"
)

func TestNewService(t *testing.T) {
	ctx := t.Context()

	tests := []struct {
		name      string
		projectID string
		location  string
		opts      []option.ClientOption
		wantErr   bool
	}{
		{
			name:      "valid configuration",
			projectID: "test-project",
			location:  "us-central1",
			opts:      nil,
			wantErr:   false,
		},
		{
			name:      "with custom logger",
			projectID: "test-project",
			location:  "us-central1",
			opts:      []option.ClientOption{option.WithLogger(slog.Default())},
			wantErr:   false,
		},
		{
			name:      "empty project ID",
			projectID: "",
			location:  "us-central1",
			opts:      nil,
			wantErr:   true,
		},
		{
			name:      "empty location",
			projectID: "test-project",
			location:  "",
			opts:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewService(ctx, tt.projectID, tt.location, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if service == nil {
					t.Error("NewService() returned nil service")
					return
				}

				// Verify service configuration
				if got := service.GetProjectID(); got != tt.projectID {
					t.Errorf("GetProjectID() = %v, want %v", got, tt.projectID)
				}

				if got := service.GetLocation(); got != tt.location {
					t.Errorf("GetLocation() = %v, want %v", got, tt.location)
				}

				// Clean up
				if err := service.Close(); err != nil {
					t.Errorf("Close() error = %v", err)
				}
			}
		})
	}
}

func TestService_CreateCache(t *testing.T) {
	ctx := t.Context()

	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	content := &genai.Content{
		Parts: []*genai.Part{
			{Text: "This is a large document that should be cached for optimal token usage."},
		},
		Role: "user",
	}

	tests := []struct {
		name    string
		content *genai.Content
		config  *CacheConfig
		wantErr bool
	}{
		{
			name:    "valid cache creation",
			content: content,
			config: &CacheConfig{
				DisplayName: "Test Cache",
				Model:       ModelGemini20Flash001,
				TTL:         time.Hour * 24,
			},
			wantErr: false,
		},
		{
			name:    "nil content",
			content: nil,
			config: &CacheConfig{
				DisplayName: "Test Cache",
				Model:       ModelGemini20Flash001,
				TTL:         time.Hour * 24,
			},
			wantErr: true,
		},
		{
			name:    "nil config",
			content: content,
			config:  nil,
			wantErr: true,
		},
		{
			name:    "unsupported model",
			content: content,
			config: &CacheConfig{
				DisplayName: "Test Cache",
				Model:       "unsupported-model",
				TTL:         time.Hour * 24,
			},
			wantErr: true,
		},
		{
			name:    "zero TTL",
			content: content,
			config: &CacheConfig{
				DisplayName: "Test Cache",
				Model:       ModelGemini20Flash001,
				TTL:         0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := service.CreateCache(ctx, tt.content, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cache == nil {
					t.Error("CreateCache() returned nil cache")
					return
				}

				// Verify cache properties
				if cache.DisplayName != tt.config.DisplayName {
					t.Errorf("Cache DisplayName = %v, want %v", cache.DisplayName, tt.config.DisplayName)
				}

				if cache.Model != tt.config.Model {
					t.Errorf("Cache Model = %v, want %v", cache.Model, tt.config.Model)
				}

				if cache.State != CacheStateActive {
					t.Errorf("Cache State = %v, want %v", cache.State, CacheStateActive)
				}

				if cache.Name == "" {
					t.Error("Cache Name is empty")
				}

				if len(cache.Contents) != 1 {
					t.Errorf("Cache Contents length = %v, want 1", len(cache.Contents))
				}

				// Verify TTL was applied
				expectedExpire := time.Now().Add(tt.config.TTL)
				if cache.ExpireTime.Before(expectedExpire.Add(-time.Minute)) ||
					cache.ExpireTime.After(expectedExpire.Add(time.Minute)) {
					t.Errorf("Cache ExpireTime not within expected range")
				}
			}
		})
	}
}

func TestService_GetCache(t *testing.T) {
	ctx := t.Context()

	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name      string
		cacheName string
		wantErr   bool
	}{
		{
			name:      "valid cache name",
			cacheName: "projects/test-project/locations/us-central1/cachedContents/test-cache",
			wantErr:   false,
		},
		{
			name:      "empty cache name",
			cacheName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := service.GetCache(ctx, tt.cacheName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cache == nil {
					t.Error("GetCache() returned nil cache")
					return
				}

				if cache.Name != tt.cacheName {
					t.Errorf("Cache Name = %v, want %v", cache.Name, tt.cacheName)
				}

				if cache.State != CacheStateActive {
					t.Errorf("Cache State = %v, want %v", cache.State, CacheStateActive)
				}
			}
		})
	}
}

func TestService_ListCaches(t *testing.T) {
	ctx := t.Context()

	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name string
		opts *ListCacheOptions
	}{
		{
			name: "default options",
			opts: nil,
		},
		{
			name: "custom page size",
			opts: &ListCacheOptions{PageSize: 10},
		},
		{
			name: "with page token",
			opts: &ListCacheOptions{PageSize: 50, PageToken: "test-token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.ListCaches(ctx, tt.opts)
			if err != nil {
				t.Errorf("ListCaches() error = %v", err)
				return
			}

			if response == nil {
				t.Error("ListCaches() returned nil response")
				return
			}

			// Verify response structure
			if len(response.CachedContents) == 0 {
				t.Error("ListCaches() returned empty cache list")
			}

			// Verify each cache in the response
			for _, cache := range response.CachedContents {
				if cache.Name == "" {
					t.Error("Cache Name is empty")
				}

				if cache.State == "" {
					t.Error("Cache State is empty")
				}
			}
		})
	}
}

func TestService_UpdateCache(t *testing.T) {
	ctx := t.Context()

	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name          string
		cachedContent *CachedContent
		updateMask    []string
		wantErr       bool
	}{
		{
			name: "valid update",
			cachedContent: &CachedContent{
				Name:        "projects/test-project/locations/us-central1/cachedContents/test-cache",
				DisplayName: "Updated Cache",
			},
			updateMask: []string{"display_name"},
			wantErr:    false,
		},
		{
			name:          "nil cached content",
			cachedContent: nil,
			updateMask:    []string{"display_name"},
			wantErr:       true,
		},
		{
			name: "empty cache name",
			cachedContent: &CachedContent{
				Name:        "",
				DisplayName: "Updated Cache",
			},
			updateMask: []string{"display_name"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := service.UpdateCache(ctx, tt.cachedContent, tt.updateMask)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cache == nil {
					t.Error("UpdateCache() returned nil cache")
					return
				}

				// Verify update timestamp was set
				if cache.UpdateTime.IsZero() {
					t.Error("UpdateTime was not set")
				}
			}
		})
	}
}

func TestService_DeleteCache(t *testing.T) {
	ctx := t.Context()

	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	tests := []struct {
		name      string
		cacheName string
		wantErr   bool
	}{
		{
			name:      "valid cache name",
			cacheName: "projects/test-project/locations/us-central1/cachedContents/test-cache",
			wantErr:   false,
		},
		{
			name:      "empty cache name",
			cacheName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteCache(ctx, tt.cacheName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CreateCacheWithTTL(t *testing.T) {
	ctx := t.Context()

	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	content := &genai.Content{
		Parts: []*genai.Part{{Text: "Test content"}},
		Role:  "user",
	}

	cache, err := service.CreateCacheWithTTL(ctx, content, ModelGemini20Flash001, "Test Cache", time.Hour*24)
	if err != nil {
		t.Errorf("CreateCacheWithTTL() error = %v", err)
		return
	}

	if cache == nil {
		t.Error("CreateCacheWithTTL() returned nil cache")
		return
	}

	// Verify cache properties
	if cache.DisplayName != "Test Cache" {
		t.Errorf("Cache DisplayName = %v, want Test Cache", cache.DisplayName)
	}

	if cache.Model != ModelGemini20Flash001 {
		t.Errorf("Cache Model = %v, want %v", cache.Model, ModelGemini20Flash001)
	}
}

func TestService_CreateCacheForModel(t *testing.T) {
	ctx := t.Context()

	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	content := &genai.Content{
		Parts: []*genai.Part{{Text: "Test content"}},
		Role:  "user",
	}

	cache, err := service.CreateCacheForModel(ctx, content, ModelGemini20Pro001)
	if err != nil {
		t.Errorf("CreateCacheForModel() error = %v", err)
		return
	}

	if cache == nil {
		t.Error("CreateCacheForModel() returned nil cache")
		return
	}

	// Verify cache properties
	expectedDisplayName := "Cache for " + ModelGemini20Pro001
	if cache.DisplayName != expectedDisplayName {
		t.Errorf("Cache DisplayName = %v, want %v", cache.DisplayName, expectedDisplayName)
	}

	if cache.Model != ModelGemini20Pro001 {
		t.Errorf("Cache Model = %v, want %v", cache.Model, ModelGemini20Pro001)
	}
}

func TestIsSupportedModel(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		want      bool
	}{
		{
			name:      "Gemini 2.0 Flash 001",
			modelName: ModelGemini20Flash001,
			want:      true,
		},
		{
			name:      "Gemini 2.0 Pro 001",
			modelName: ModelGemini20Pro001,
			want:      true,
		},
		{
			name:      "unsupported model",
			modelName: "gemini-1.5-pro",
			want:      false,
		},
		{
			name:      "empty model name",
			modelName: "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSupportedModel(tt.modelName); got != tt.want {
				t.Errorf("IsSupportedModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSupportedModels(t *testing.T) {
	models := GetSupportedModels()

	if len(models) == 0 {
		t.Error("GetSupportedModels() returned empty list")
	}

	expectedModels := []string{ModelGemini20Flash001, ModelGemini20Pro001}
	if diff := cmp.Diff(expectedModels, models); diff != "" {
		t.Errorf("GetSupportedModels() mismatch (-want +got):\n%s", diff)
	}
}

// Benchmark tests
func BenchmarkService_CreateCache(b *testing.B) {
	ctx := b.Context()

	service, err := NewService(ctx, "test-project", "us-central1")
	if err != nil {
		b.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	content := &genai.Content{
		Parts: []*genai.Part{{Text: "Benchmark content"}},
		Role:  "user",
	}

	config := &CacheConfig{
		DisplayName: "Benchmark Cache",
		Model:       ModelGemini20Flash001,
		TTL:         time.Hour * 24,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := service.CreateCache(ctx, content, config)
		if err != nil {
			b.Fatalf("CreateCache() error = %v", err)
		}
	}
}
