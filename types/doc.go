// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package types provides core interfaces and contracts for the Agent Development Kit (ADK).
//
// The types package defines the fundamental interfaces, structures, and contracts that all
// components of the ADK system follow. It serves as the central definition of how agents,
// models, tools, sessions, and other components interact with each other.
//
// # Core Interfaces
//
// The package defines several key interfaces that form the foundation of the ADK:
//
//   - Agent: Defines execution, hierarchy navigation, and lifecycle methods
//   - Model: Unified LLM abstraction for content generation and streaming
//   - Tool/Toolset: Extensible tool system with function declarations
//   - Session/SessionService: Conversation and state management abstractions
//   - Flow: Pipeline architecture for LLM interaction processing
//   - ArtifactService: Storage and retrieval of agent-generated content
//   - CredentialService: Secure authentication credential management
//
// # Agent Interface
//
// The Agent interface defines the contract for all agent types:
//
//	type Agent interface {
//		Name() string
//		Description() string
//		ParentAgent() Agent
//		SubAgents() []Agent
//		BeforeAgentCallbacks() []AgentCallback
//		AfterAgentCallbacks() []AgentCallback
//		Execute(ctx context.Context, ictx *InvocationContext) iter.Seq2[*Event, error]
//		Run(ctx context.Context, ictx *InvocationContext) iter.Seq2[*Event, error]
//		RunLive(ctx context.Context, ictx *InvocationContext) iter.Seq2[*Event, error]
//		// ... additional methods
//	}
//
// All agent implementations (LLMAgent, SequentialAgent, ParallelAgent, LoopAgent)
// implement this interface, enabling polymorphic usage and hierarchical composition.
//
// # Model Interface
//
// The Model interface provides unified LLM abstraction:
//
//	type Model interface {
//		Name() string
//		SupportedModels() []string
//		Connect(ctx context.Context, request *LLMRequest) (ModelConnection, error)
//		GenerateContent(ctx context.Context, request *LLMRequest) (*LLMResponse, error)
//		StreamGenerateContent(ctx context.Context, request *LLMRequest) iter.Seq2[*LLMResponse, error]
//	}
//
// This abstraction supports multiple providers (Google Gemini, Anthropic Claude) with
// consistent interfaces for both synchronous and streaming generation.
//
// # Tool System
//
// The tool system is built around flexible interfaces:
//
//	type Tool interface {
//		Name() string
//		Description() string
//		IsLongRunning() bool
//		GetDeclaration() *genai.FunctionDeclaration
//		Run(ctx context.Context, args map[string]any, toolCtx *ToolContext) (any, error)
//		ProcessLLMRequest(ctx context.Context, toolCtx *ToolContext, request *LLMRequest) error
//	}
//
//	type Toolset interface {
//		Name() string
//		Description() string
//		Tools() []Tool
//	}
//
// Tools can be individual functions or grouped into toolsets for organized functionality.
//
// # Event System
//
// Events represent all interactions in the system:
//
//	type Event struct {
//		*LLMResponse
//		InvocationID       string
//		Author             string
//		Actions            *EventActions
//		LongRunningToolIDs py.Set[string]
//		Branch             string
//		ID                 string
//		Timestamp          time.Time
//	}
//
//	type EventActions struct {
//		StateDelta         map[string]any
//		AgentTransfer      *AgentTransfer
//		FunctionCalls      []*genai.FunctionCall
//		FunctionResponses  []*genai.FunctionResponse
//	}
//
// Events carry rich metadata and enable comprehensive conversation tracking.
//
// # Session Management
//
// Sessions provide stateful conversation tracking:
//
//	type Session interface {
//		ID() string
//		AppName() string
//		UserID() string
//		State() map[string]any
//		Events() []*Event
//		LastUpdateTime() time.Time
//		AddEvent(events ...*Event)
//		// ... additional methods
//	}
//
//	type SessionService interface {
//		CreateSession(ctx context.Context, appName, userID, sessionID string, state map[string]any) (Session, error)
//		GetSession(ctx context.Context, appName, userID, sessionID string, config *GetSessionConfig) (Session, error)
//		ListSessions(ctx context.Context, appName, userID string) ([]Session, error)
//		DeleteSession(ctx context.Context, appName, userID, sessionID string) error
//		AppendEvent(ctx context.Context, ses Session, event *Event) (*Event, error)
//		ListEvents(ctx context.Context, appName, userID, sessionID string, maxEvents int, since *time.Time) ([]Event, error)
//	}
//
// # State Management
//
// The ADK supports three-tier state management:
//
//	// App-level state (shared across all users)
//	StateDelta["app:config"] = "production"
//
//	// User-level state (shared across user's sessions)
//	StateDelta["user:preferences"] = userPrefs
//
//	// Session-level state (specific to conversation)
//	StateDelta["temp:context"] = "current_topic"
//
// State changes are applied through EventActions with automatic propagation.
//
// # Context System
//
// Rich context flows through all operations:
//
//	type InvocationContext struct {
//		// Session and state management
//		Session() Session
//		SessionService() SessionService
//		State() map[string]any
//
//		// Agent hierarchy
//		Agent() Agent
//		ParentAgent() Agent
//		RootAgent() Agent
//
//		// Services and configuration
//		ArtifactService() ArtifactService
//		CredentialService() CredentialService
//		MemoryService() MemoryService
//		CodeExecutor() CodeExecutor
//
//		// Execution tracking
//		AppName() string
//		UserID() string
//		SessionID() string
//		InvocationID() string
//		Branch() string
//	}
//
// # Flow System
//
// Flows provide pipeline architecture for processing:
//
//	type Flow interface {
//		Run(ctx context.Context, ictx *InvocationContext) iter.Seq2[*Event, error]
//		RunLive(ctx context.Context, ictx *InvocationContext) iter.Seq2[*Event, error]
//	}
//
// Flows can be composed into pipelines with processors for request/response handling.
//
// # Code Execution
//
// Secure code execution is abstracted through interfaces:
//
//	type CodeExecutor interface {
//		OptimizeDataFile() bool
//		IsLongRunning() bool
//		IsStateful() bool
//		ErrorRetryAttempts() int
//		CodeBlockDelimiters() []DelimiterPair
//		ExecutionResultDelimiters() DelimiterPair
//		ExecuteCode(ctx context.Context, ictx *InvocationContext, input *CodeExecutionInput) (*CodeExecutionResult, error)
//		Close() error
//	}
//
// # Authentication System
//
// Authentication is handled through flexible interfaces:
//
//	type CredentialService interface {
//		LoadCredential(ctx context.Context, authConfig *AuthConfig, toolCtx *ToolContext) (*AuthCredential, error)
//		SaveCredential(ctx context.Context, authConfig *AuthConfig, toolCtx *ToolContext) error
//	}
//
//	type AuthConfig struct {
//		Type                   AuthType
//		CredentialKey          string
//		ClientID               string
//		ClientSecret           string
//		Scopes                 []string
//		ExchangedAuthCredential *AuthCredential
//		// ... additional fields
//	}
//
// # Memory System
//
// Long-term memory is provided through:
//
//	type MemoryService interface {
//		AddMemories(ctx context.Context, appName, userID string, memories []*MemoryEntry) error
//		SearchMemories(ctx context.Context, appName, userID, query string) (*SearchMemoryResponse, error)
//	}
//
// # Artifact Management
//
// Artifacts provide persistent storage:
//
//	type ArtifactService interface {
//		SaveArtifact(ctx context.Context, appName, userID, sessionID, filename string, artifact *genai.Part) (int, error)
//		LoadArtifact(ctx context.Context, appName, userID, sessionID, filename string, version int) (*genai.Part, error)
//		ListArtifactKey(ctx context.Context, appName, userID, sessionID string) ([]string, error)
//		DeleteArtifact(ctx context.Context, appName, userID, sessionID, filename string) error
//		ListVersions(ctx context.Context, appName, userID, sessionID, filename string) ([]int, error)
//		Close() error
//	}
//
// # Planning System
//
// Strategic planning is defined through:
//
//	type Planner interface {
//		BuildPlanningInstruction(ctx context.Context, rctx *ReadOnlyContext, request *LLMRequest) string
//		ProcessPlanningResponse(ctx context.Context, cctx *CallbackContext, responseParts []*genai.Part) []*genai.Part
//	}
//
// # Python Compatibility
//
// The types/py subpackage provides Go implementations of Python patterns:
//
//	// Python-style sets
//	set := py.NewSet[string]()
//	set.Add("item1", "item2")
//
//	// Python asyncio patterns
//	task := pyasyncio.CreateTask(ctx, asyncFunction)
//	result, err := task.Result(ctx)
//
// # Iterator Patterns
//
// The ADK extensively uses Go 1.23+ iterators for streaming:
//
//	for event, err := range agent.Run(ctx, ictx) {
//		if err != nil {
//			// Handle error
//			continue
//		}
//		// Process event
//	}
//
// # Error Handling
//
// The package defines specific error types for different scenarios:
//
//	type ExecutionError struct {
//		Message   string
//		Attempts  int
//		LastError error
//	}
//
//	type AuthenticationError struct {
//		Type    AuthType
//		Message string
//	}
//
// # Thread Safety
//
// All interfaces are designed to be safe for concurrent use. Implementations
// should provide appropriate synchronization where needed.
//
// # Integration Patterns
//
// The types package enables flexible composition patterns:
//
//	// Agent hierarchies
//	coordinator := NewSequentialAgent("coordinator",
//		researcher, analyzer, reporter)
//
//	// Tool composition
//	toolset := NewToolset("data_tools",
//		searchTool, analysisTool, visualizationTool)
//
//	// Service injection
//	ictx := NewInvocationContext(session, sessionService, artifactService, credentialService)
//
// # Best Practices
//
// When implementing these interfaces:
//
//  1. Follow interface contracts exactly
//  2. Handle context cancellation appropriately
//  3. Provide meaningful error messages
//  4. Ensure thread safety where required
//  5. Clean up resources in Close() methods
//  6. Use appropriate state prefixes (app:, user:, temp:)
//  7. Validate inputs early and thoroughly
//
// The types package provides the foundation for building sophisticated AI agents
// with strong typing, clear contracts, and flexible composition patterns.
package types
