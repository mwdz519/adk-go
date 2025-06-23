// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package artifact provides storage services for managing agent artifacts with versioning support.
//
// The artifact package implements the types.ArtifactService interface with multiple storage backends
// for managing files, data, and content generated or used by agents. Artifacts are organized in a
// hierarchical structure by application, user, and session for proper isolation and management.
//
// # Supported Backends
//
// The package provides two storage implementations:
//
//   - InMemoryService: Fast in-memory storage for development and testing
//   - GCSService: Google Cloud Storage backend for production scalability
//
// # Artifact Organization
//
// Artifacts are organized hierarchically:
//
//	{appName}/{userID}/{sessionID}/{filename}  // Session-scoped artifacts
//	{appName}/{userID}/user/{filename}         // User-scoped artifacts (user: prefix)
//
// This structure provides proper isolation between applications, users, and sessions
// while supporting both session-specific and user-persistent artifacts.
//
// # Versioning
//
// All artifacts support automatic versioning:
//   - Each save operation creates a new version
//   - Versions are identified by incremental integers
//   - List and load operations support version-specific access
//   - Version history can be retrieved for any artifact
//
// # Basic Usage
//
// Creating a service:
//
//	// In-memory for development
//	service := artifact.NewInMemoryService()
//
//	// Google Cloud Storage for production
//	service, err := artifact.NewGCSService(ctx, "my-bucket")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer service.Close()
//
// Saving and loading artifacts:
//
//	// Save a text artifact
//	content := &genai.Part{Text: "Generated report content"}
//	version, err := service.SaveArtifact(ctx, "myapp", "user123", "session456", "report.txt", content)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Load the latest version
//	artifact, err := service.LoadArtifact(ctx, "myapp", "user123", "session456", "report.txt", version)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # User-Scoped Artifacts
//
// Use the "user:" prefix for artifacts that persist across sessions:
//
//	// Save user preference (persists across sessions)
//	preference := &genai.Part{Text: "dark_mode"}
//	_, err := service.SaveArtifact(ctx, "myapp", "user123", "session456", "user:theme", preference)
//
// # Listing and Management
//
// The service provides comprehensive artifact management:
//
//	// List all artifacts in a session
//	filenames, err := service.ListArtifactKey(ctx, "myapp", "user123", "session456")
//
//	// List all versions of a specific artifact
//	versions, err := service.ListVersions(ctx, "myapp", "user123", "session456", "report.txt")
//
//	// Delete an artifact (all versions)
//	err := service.DeleteArtifact(ctx, "myapp", "user123", "session456", "report.txt")
//
// # Content Types
//
// Artifacts support all genai.Part types:
//   - Text content
//   - Binary data
//   - Images
//   - Audio
//   - Video
//   - Files
//
// # Thread Safety
//
// All service implementations are safe for concurrent use across multiple goroutines.
// The in-memory service uses internal locking, while GCS provides atomic operations.
//
// # Integration with Tools
//
// The artifact service integrates seamlessly with agent tools through the ToolContext:
//
//	func MyTool(ctx context.Context, toolCtx *types.ToolContext) error {
//		// Access artifact service from tool context
//		artifactService := toolCtx.GetArtifactService()
//
//		// Save generated content
//		content := &genai.Part{Text: "Tool output"}
//		_, err := artifactService.SaveArtifact(ctx,
//			toolCtx.AppName(), toolCtx.UserID(), toolCtx.SessionID(),
//			"tool_output.txt", content)
//		return err
//	}
//
// This enables tools to persist their outputs and share data across agent executions.
package artifact
