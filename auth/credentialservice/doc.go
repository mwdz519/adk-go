// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package credentialservice provides storage and management of authentication credentials for tools and agents.
//
// The credentialservice package implements the types.CredentialService interface for secure storage
// and retrieval of authentication credentials used by tools that require external API access.
// Credentials are organized hierarchically by application and user for proper isolation.
//
// # Credential Organization
//
// Credentials are organized in a three-tier hierarchy:
//
//	{appName} -> {userID} -> {credentialKey} -> AuthCredential
//
// This structure ensures:
//   - Application isolation: Each app has separate credential storage
//   - User isolation: Each user's credentials are kept separate
//   - Credential type isolation: Multiple credential types per user (API keys, OAuth tokens, etc.)
//
// # Supported Backends
//
// Currently provides:
//   - InMemory: Fast in-memory storage for development and testing
//
// Additional backends (database, secure vault, etc.) can be implemented by satisfying
// the types.CredentialService interface.
//
// # Basic Usage
//
// Creating a credential service:
//
//	service := credentialservice.NewInMemory()
//
// The service is typically used through tools via the ToolContext:
//
//	func MyAPITool(ctx context.Context, toolCtx *types.ToolContext) error {
//		// Request authentication if needed
//		authConfig := &types.AuthConfig{
//			Type: types.AuthTypeAPIKey,
//			Key:  "api_key",
//			// ... other config
//		}
//
//		// Tool framework automatically loads/saves credentials
//		toolCtx.RequestCredential("my_api_key", authConfig)
//
//		// Use authenticated client...
//		return nil
//	}
//
// # Credential Flow
//
// The typical credential flow:
//
//  1. Tool requests credentials via ToolContext.RequestCredential()
//  2. Service attempts to load existing credentials via LoadCredential()
//  3. If not found, authentication flow is initiated
//  4. New credentials are saved via SaveCredential()
//  5. Tool receives authenticated client/credentials
//
// # Security Model
//
// The credential service provides:
//   - Application isolation: Apps cannot access each other's credentials
//   - User isolation: Users cannot access each other's credentials
//   - Credential type isolation: Different auth types are stored separately
//   - Secure credential storage: Credentials are stored securely based on backend
//
// # Experimental Status
//
// This package is marked as experimental and may undergo breaking changes.
// The API and storage format may change in future versions.
//
// # Integration with Authentication
//
// The service works with the broader authentication system:
//
//	// AuthConfig defines how to authenticate
//	authConfig := &types.AuthConfig{
//		Type:           types.AuthTypeOAuth2,
//		CredentialKey:  "google_oauth",
//		ClientID:       "your-client-id",
//		ClientSecret:   "your-client-secret",
//		Scopes:         []string{"scope1", "scope2"},
//		RedirectURI:    "http://localhost:8080/callback",
//	}
//
//	// Service handles loading/saving automatically
//	credential, err := service.LoadCredential(ctx, authConfig, toolCtx)
//
// # Thread Safety
//
// The InMemory implementation uses lazy initialization and is safe for concurrent
// access across multiple goroutines. All operations are atomic at the credential level.
//
// # Context Integration
//
// All operations accept context.Context for:
//   - Cancellation support
//   - Timeout handling
//   - Request-scoped values
//   - Distributed tracing
//
// The ToolContext provides access to:
//   - Application name for credential isolation
//   - User ID for user-specific credentials
//   - Session information for audit trails
package credentialservice

