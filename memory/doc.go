// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package memory provides long-term knowledge storage and retrieval capabilities for persistent agent memory across sessions.
//
// The memory package implements the types.MemoryService interface to enable agents to store,
// search, and retrieve information from past conversations and interactions. This allows agents
// to build knowledge over time and provide context-aware responses based on historical data.
//
// # Memory Services
//
// The package provides two distinct memory service implementations:
//
//   - InMemoryService: Simple keyword-based search for development and prototyping
//   - VertexAIRagService: Production-ready semantic search using Google Cloud Vertex AI RAG
//
// # Architecture Overview
//
// Memory services follow a consistent interface for storing and retrieving agent memories:
//
//	┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
//	│    Sessions     │───▶│   Memory         │───▶│   Search        │
//	│   (Events &     │    │   Service        │    │   Results       │
//	│    Content)     │    │                  │    │                 │
//	└─────────────────┘    └──────────────────┘    └─────────────────┘
//
// Sessions containing agent interactions are stored in memory services, which can then
// be searched using natural language queries to retrieve relevant historical context.
//
// # Basic Usage
//
// ## In-Memory Service
//
// For development and prototyping:
//
//	// Create in-memory service
//	memoryService := memory.NewInMemoryService()
//
//	// Store a session in memory
//	err := memoryService.AddSessionToMemory(ctx, session)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Search for relevant memories
//	response, err := memoryService.SearchMemory(ctx, "myapp", "user123", "tell me about weather data")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use retrieved memories
//	for _, memory := range response.Memories {
//		fmt.Printf("Found: %s by %s at %v\n",
//			memory.Content.Parts[0].Text, memory.Author, memory.Timestamp)
//	}
//
// ## Vertex AI RAG Service
//
// For production deployments with semantic search:
//
//	// Create Vertex AI RAG service
//	ragService, err := memory.NewVertexAIRagService(ctx,
//		"my-project", "us-central1", "my-rag-corpus",
//		memory.WithSimilarityTopK(10),
//		memory.WithVectorDistanceThreshold(0.7),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer ragService.Close()
//
//	// Store session in vector database
//	err = ragService.AddSessionToMemory(ctx, session)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Perform semantic search
//	response, err := ragService.SearchMemory(ctx, "myapp", "user123",
//		"what did we discuss about machine learning models?")
//
// # Memory Storage Model
//
// ## Session-Based Storage
//
// Memories are organized by sessions, which contain events with rich content:
//
//	session := &types.Session{
//		AppName:  "data_analysis_app",
//		UserID:   "analyst_123",
//		ID:       "session_456",
//		Events: []*types.Event{
//			{
//				Author:    "user",
//				Content:   &genai.Content{Parts: []genai.Part{genai.Text("Analyze sales data")}},
//				Timestamp: time.Now(),
//			},
//			{
//				Author:    "agent",
//				Content:   &genai.Content{Parts: []genai.Part{genai.Text("Sales increased 15% this quarter")}},
//				Timestamp: time.Now(),
//			},
//		},
//	}
//
//	// Store entire session
//	memoryService.AddSessionToMemory(ctx, session)
//
// ## Memory Entry Structure
//
// Retrieved memories are returned as MemoryEntry objects:
//
//	type MemoryEntry struct {
//		Content   *genai.Content  // The actual text/data content
//		Author    string          // Who created this content (user, agent, tool)
//		Timestamp time.Time       // When this memory was created
//	}
//
// # In-Memory Service Implementation
//
// ## Features
//
// The InMemoryService provides:
//   - Simple keyword-based matching
//   - Fast setup for development and testing
//   - No external dependencies
//   - Thread-safe concurrent access
//   - Automatic memory organization by app/user
//
// ## Search Algorithm
//
// Uses basic keyword matching for memory retrieval:
//
//	// Create service with custom logger
//	service := memory.NewInMemoryService().WithLogger(logger)
//
//	// Searches use word-level matching
//	query := "machine learning models"
//	// Will match memories containing "machine", "learning", or "models"
//
// ## Limitations
//
// The InMemoryService has several limitations:
//   - No semantic understanding (only exact word matches)
//   - Memory lost on application restart
//   - Linear search performance (O(n) with number of memories)
//   - No advanced filtering or ranking
//   - Limited to single-node deployments
//
// ## When to Use
//
// Use InMemoryService for:
//   - Development and prototyping
//   - Testing agent behaviors
//   - Small-scale deployments
//   - Applications with simple memory requirements
//
// # Vertex AI RAG Service Implementation
//
// ## Features
//
// The VertexAIRagService provides:
//   - Semantic search using vector embeddings
//   - Production-scale storage and retrieval
//   - Configurable similarity thresholds
//   - Automatic text chunking and processing
//   - Persistent storage across restarts
//   - Multi-user and multi-app isolation
//
// ## Configuration Options
//
// Extensive configuration for production requirements:
//
//	ragService, err := memory.NewVertexAIRagService(ctx,
//		"my-project-id",           // Google Cloud project
//		"us-central1",             // Vertex AI location
//		"my-rag-corpus",           // RAG corpus resource name
//		memory.WithSimilarityTopK(15),              // Return top 15 results
//		memory.WithVectorDistanceThreshold(0.8),    // Higher similarity threshold
//		memory.WithVertexAIRagLogger(customLogger), // Custom logging
//	)
//
// ## Vector Search Capabilities
//
// Advanced semantic search features:
//
//	// Semantic queries work beyond keyword matching
//	queries := []string{
//		"what machine learning techniques were discussed?",   // Matches ML content
//		"tell me about data analysis methods",               // Matches analytics content
//		"show me performance metrics and results",           // Matches numerical data
//	}
//
//	for _, query := range queries {
//		response, err := ragService.SearchMemory(ctx, appName, userID, query)
//		// Returns semantically relevant memories even if exact words don't match
//	}
//
// ## Data Processing Pipeline
//
// The RAG service processes session data through several stages:
//
//  1. Event Extraction: Extract text content from session events
//  2. JSON Serialization: Structure event data with metadata
//  3. File Upload: Upload to Google Cloud Storage
//  4. Vector Indexing: Generate embeddings and build search index
//  5. Search Integration: Enable semantic search across stored content
//
// ## Filtering and Isolation
//
// Built-in filtering ensures proper data isolation:
//
//	// Searches are automatically filtered by app and user
//	response, err := ragService.SearchMemory(ctx, "app1", "user1", query)
//	// Only returns memories from app1/user1, never from other users/apps
//
// # Integration with Agent System
//
// ## Memory-Enabled Agents
//
// Memory services integrate seamlessly with agents:
//
//	// Create agent with memory service
//	agent := agent.NewLLMAgent(ctx, "memory_agent",
//		agent.WithModel("gemini-1.5-pro"),
//		agent.WithMemoryService(ragService),
//		agent.WithInstruction("Use your memory to provide context-aware responses"),
//	)
//
//	// Memory is automatically available during agent execution
//	for event, err := range agent.Run(ctx, ictx) {
//		// Agent can access memories from past sessions
//	}
//
// ## Automatic Memory Tools
//
// The agent system can automatically provide memory access tools:
//
//	// Tools for loading and searching memories
//	memoryTools := []types.Tool{
//		tools.NewLoadMemoryTool(ragService),
//		tools.NewPreloadMemoryTool(ragService),
//	}
//
//	agent := agent.NewLLMAgent(ctx, "assistant",
//		agent.WithTools(memoryTools...),
//		agent.WithMemoryService(ragService),
//	)
//
// ## Context Integration
//
// Memories are accessible through the invocation context:
//
//	func customTool(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
//		// Access memory service from context
//		memoryService := toolCtx.InvocationContext().MemoryService()
//		if memoryService != nil {
//			response, err := memoryService.SearchMemory(ctx,
//				toolCtx.InvocationContext().AppName(),
//				toolCtx.InvocationContext().UserID(),
//				"relevant query")
//		}
//		return result, nil
//	}
//
// # Performance Characteristics
//
// ## In-Memory Service
//
//   - Storage: O(n) memory usage where n = number of events
//   - Search: O(n) time complexity for each search
//   - Concurrency: Thread-safe with read-write mutex
//   - Scalability: Limited to single node, bounded by available RAM
//
// ## Vertex AI RAG Service
//
//   - Storage: Virtually unlimited (Google Cloud Storage)
//   - Search: O(log n) with vector index, sub-second response times
//   - Concurrency: Highly concurrent, production-scale
//   - Scalability: Auto-scaling based on Google Cloud infrastructure
//
// # Error Handling
//
// ## Common Error Scenarios
//
//	// Handle memory service errors appropriately
//	response, err := memoryService.SearchMemory(ctx, appName, userID, query)
//	if err != nil {
//		if errors.Is(err, context.DeadlineExceeded) {
//			// Search timeout - use fallback or cached results
//			handleTimeout()
//		} else if strings.Contains(err.Error(), "permission denied") {
//			// Authentication/authorization issue
//			handleAuthError()
//		} else {
//			// Other errors - log and continue without memory
//			log.Printf("Memory search failed: %v", err)
//		}
//	}
//
//	// Always check for nil/empty results
//	if response == nil || len(response.Memories) == 0 {
//		// No memories found - proceed without context
//		proceedWithoutMemory()
//	}
//
// # Best Practices
//
//  1. Use InMemoryService for development, VertexAIRagService for production
//  2. Store complete sessions rather than individual events for better context
//  3. Include relevant metadata (timestamps, authors) for better search results
//  4. Handle memory service failures gracefully (agents should work without memory)
//  5. Configure appropriate similarity thresholds for your use case
//  6. Monitor memory usage and search performance in production
//  7. Implement proper authentication and authorization for memory access
//  8. Consider data retention policies and privacy requirements
//
// # Security Considerations
//
// ## Data Isolation
//
//   - Memories are automatically filtered by application and user
//   - Cross-user/cross-app data leakage is prevented by design
//   - Search filters ensure proper isolation boundaries
//
// ## Access Control
//
//   - Memory services respect authentication context
//   - Google Cloud IAM controls access to Vertex AI resources
//   - Implement application-level access controls as needed
//
// ## Data Privacy
//
//   - Consider data retention policies for stored memories
//   - Implement data anonymization where required
//   - Use appropriate Google Cloud security features
//   - Monitor access patterns and audit logs
//
// # Migration and Deployment
//
// ## Development to Production
//
//	// Development configuration
//	var memoryService types.MemoryService
//	if isDevelopment {
//		memoryService = memory.NewInMemoryService()
//	} else {
//		// Production configuration
//		memoryService, err = memory.NewVertexAIRagService(ctx,
//			projectID, location, ragCorpus,
//			memory.WithSimilarityTopK(10),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//
// ## Resource Management
//
//	// Always close services to release resources
//	defer memoryService.Close()
//
// # Future Enhancements
//
// The memory package is designed for extensibility:
//   - Additional vector database backends (Pinecone, Weaviate, etc.)
//   - Enhanced filtering and ranking algorithms
//   - Memory compression and archival features
//   - Advanced privacy and security controls
//   - Cross-session memory aggregation and insights
//
// The memory package provides the foundation for building agents with persistent knowledge
// and context-aware capabilities across multiple interaction sessions.
package memory
