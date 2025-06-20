// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package session provides stateful conversation tracking and state management for agent interactions.
//
// The session package implements the types.SessionService interface for managing user sessions,
// conversation history, and state persistence across agent interactions. Sessions are organized
// hierarchically by application and user for proper isolation and management.
//
// # Core Components
//
// The package provides:
//
//   - Session: Tracks events, state, and metadata for a single conversation
//   - SessionService: CRUD operations and event management interface
//   - InMemoryService: Reference implementation with thread safety
//   - Three-tier state management (app, user, session)
//
// # Session Organization
//
// Sessions are organized hierarchically:
//
//	{appName} -> {userID} -> {sessionID} -> Session
//
// This structure provides:
//   - Application isolation: Each app has separate session storage
//   - User isolation: Each user's sessions are kept separate
//   - Session isolation: Individual conversations are tracked separately
//
// # State Management
//
// The package supports three tiers of state:
//
//   - App State: Shared across all users of an application
//   - User State: Specific to a user across all their sessions
//   - Session State: Specific to a single conversation session
//
// # Basic Usage
//
// Creating a session service:
//
//	service := session.NewInMemoryService()
//
// Creating and managing sessions:
//
//	// Create a new session
//	session, err := service.CreateSession(ctx, "myapp", "user123", "session456", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Retrieve an existing session
//	session, err = service.GetSession(ctx, "myapp", "user123", "session456", &types.GetSessionConfig{
//		NumRecentEvents: 10, // Get last 10 events
//	})
//
//	// List all sessions for a user
//	sessions, err := service.ListSessions(ctx, "myapp", "user123")
//
// # Event Management
//
// Sessions track events throughout the conversation:
//
//	// Create an event
//	event := &types.Event{
//		Type: types.EventTypeTextDelta,
//		TextDelta: "Hello, how can I help you?",
//		Timestamp: time.Now(),
//		Actions: &types.EventActions{
//			StateDelta: map[string]any{
//				"user:preference": "friendly_tone",
//				"temp:last_topic": "greetings",
//			},
//		},
//	}
//
//	// Append event to session
//	savedEvent, err := service.AppendEvent(ctx, session, event)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # State Management Patterns
//
// Use prefixed keys for different state scopes:
//
//	// App-level state (shared across all users)
//	event.Actions.StateDelta["app:config"] = "production"
//
//	// User-level state (shared across user's sessions)
//	event.Actions.StateDelta["user:theme"] = "dark_mode"
//	event.Actions.StateDelta["user:language"] = "en"
//
//	// Session-level state (specific to this conversation)
//	event.Actions.StateDelta["temp:context"] = "discussing_weather"
//	event.Actions.StateDelta["temp:step"] = 3
//
// # Integration with Agents
//
// Sessions integrate seamlessly with the agent system:
//
//	// Create session service
//	sessionService := session.NewInMemoryService()
//
//	// Create or get session
//	session, err := sessionService.CreateSession(ctx, appName, userID, sessionID, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create invocation context with session
//	ictx := types.NewInvocationContext(session, sessionService, nil, nil)
//
//	// Run agent with session context
//	for event, err := range agent.Run(ctx, ictx) {
//		if err != nil {
//			log.Printf("Agent error: %v", err)
//			continue
//		}
//
//		// Events are automatically stored in the session
//		fmt.Printf("Event: %+v\n", event)
//	}
//
// # Event Filtering and Pagination
//
// Retrieve events with filtering options:
//
//	// Get recent events
//	config := &types.GetSessionConfig{
//		NumRecentEvents: 20,
//		AfterTimestamp:  time.Now().Add(-1 * time.Hour),
//	}
//
//	session, err := service.GetSession(ctx, appName, userID, sessionID, config)
//
//	// List events directly
//	since := time.Now().Add(-30 * time.Minute)
//	events, err := service.ListEvents(ctx, appName, userID, sessionID, 50, &since)
//
// # Session Lifecycle
//
// Typical session lifecycle:
//
//  1. CreateSession: Initialize new conversation
//  2. AppendEvent: Add user messages, agent responses, tool calls
//  3. GetSession: Retrieve conversation history for context
//  4. State updates: Track conversation progress and preferences
//  5. DeleteSession: Clean up completed conversations
//
// # Persistence and Scaling
//
// The InMemoryService is suitable for development and small deployments.
// For production use, implement SessionService with persistent storage:
//
//	type DatabaseSessionService struct {
//		db *sql.DB
//	}
//
//	func (s *DatabaseSessionService) CreateSession(ctx context.Context, appName, userID, sessionID string, state map[string]any) (types.Session, error) {
//		// Implement database storage
//		return session, nil
//	}
//
// # State Delta Processing
//
// Events can contain state changes that are automatically applied:
//
//	event := &types.Event{
//		Actions: &types.EventActions{
//			StateDelta: map[string]any{
//				"user:name":        "Alice",
//				"user:timezone":    "UTC-8",
//				"temp:calculation": 42,
//				"app:version":      "1.2.0",
//			},
//		},
//	}
//
//	// State changes are applied when event is appended
//	service.AppendEvent(ctx, session, event)
//
// # Thread Safety
//
// The InMemoryService implementation is safe for concurrent use across multiple
// goroutines. All operations use appropriate locking to ensure data consistency.
//
// # Error Handling
//
// The package provides specific errors for common scenarios:
//
//	session, err := service.GetSession(ctx, appName, userID, sessionID, nil)
//	if err != nil {
//		if errors.Is(err, types.ErrSessionNotFound) {
//			// Create new session
//			session, err = service.CreateSession(ctx, appName, userID, sessionID, nil)
//		} else {
//			log.Fatal(err)
//		}
//	}
//
// # Best Practices
//
//  1. Use meaningful session IDs for debugging and tracing
//  2. Clean up old sessions periodically to manage memory/storage
//  3. Use appropriate state prefixes (app:, user:, temp:)
//  4. Include relevant context in events for replay/debugging
//  5. Handle session not found errors gracefully
//  6. Consider session TTL for automatic cleanup
//
// Sessions provide the foundation for stateful agent interactions, enabling
// context-aware conversations and persistent user experiences.
package session
