# ADK-Go: Agent Development Kit for Go

<div align="center">

An open-source, code-first Go toolkit for building, evaluating, and deploying sophisticated AI agents with flexibility and control.

[![Go Reference](https://pkg.go.dev/badge/github.com/go-a2a/adk-go.svg)](https://pkg.go.dev/github.com/go-a2a/adk-go)
[![Go](https://github.com/go-a2a/adk-go/actions/workflows/go.yml/badge.svg)](https://github.com/go-a2a/adk-go/actions/workflows/test.yml)

[Features](#install-adk-go) •
[Installation](#install-adk-go) •
[Quick Start](#%EF%B8%8F-architecture) •
[Architecture](#%EF%B8%8F-architecture) •
[Examples](#install-adk-go) •

</div>

> [!IMPORTANT]
> This project is in the alpha stage.
>
> Flags, configuration, behavior, and design may change significantly.

---

## 🌟 Features

- **⚡️ Fully compatible with the official SDK**: Fully emulates the Python implementation of [adk-python](https://github.com/google/adk-python).
- **🤖 Multi-Agent Architecture**: Build hierarchical agent systems with LLM, Sequential, Parallel, and Loop agents
- **🔗 Multi-Provider Support**: Unified interface for Google Gemini, Anthropic Claude, and more via `google.golang.org/genai`
- **🛠️ Extensible Tools**: Rich ecosystem of tools with automatic function calling and authentication
- **💾 Memory Systems**: Long-term knowledge storage and retrieval with vector-based search
- **🔒 Secure Code Execution**: Multiple backends (built-in, container, local) with resource limits
- **🌊 Streaming First**: Real-time event streaming with Go 1.23+ iterators
- **📊 Session Management**: Stateful conversation tracking with three-tier state management
- **🎯 Smart Planning**: Strategic planning with built-in and ReAct planners
- **🔐 Authentication**: Multi-scheme auth support (OAuth2, API Key, Basic, Bearer)
- **🎬 Live Mode**: Video/audio-based conversations for supported models

## 📦 Installation

### Prerequisites

- Go 1.24 or higher
- API keys for your chosen LLM providers

### Install ADK Go

```bash
go mod init your-project
go get github.com/go-a2a/adk-go
```

### Environment Setup

```bash
# For Google Gemini
export GEMINI_API_KEY="your-gemini-api-key"

# For Anthropic Claude
export ANTHROPIC_API_KEY="your-anthropic-api-key"

# Optional: For Google Cloud AI Platform
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account.json"
```

## 🚀 Quick Start

### Simple LLM Agent

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/go-a2a/adk-go/agent"
    "github.com/go-a2a/adk-go/model"
    "github.com/go-a2a/adk-go/session"
    "github.com/go-a2a/adk-go/types"
)

func main() {
    ctx := context.Background()

    // Create a model
    m, err := model.NewGoogleModel("gemini-2.0-flash-exp")
    if err != nil {
        log.Fatal(err)
    }
    defer m.Close()

    // Create an LLM agent
    llmAgent := agent.NewLLMAgent(ctx, "assistant",
        agent.WithModel(m),
        agent.WithInstruction("You are a helpful AI assistant."),
    )

    // Create session
    sessionService := session.NewInMemoryService()
    sess, _ := sessionService.CreateSession(ctx, "myapp", "user123", "session456", nil)

    // Create invocation context
    ictx := types.NewInvocationContext(sess, sessionService, nil, nil)

    // Run the agent
    for event, err := range llmAgent.Run(ctx, ictx) {
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        // Handle events
        if event.Message != nil {
            fmt.Println("Agent:", event.Message.Text)
        }
    }
}
```

### Agent with Tools

```go
package main

import (
    "context"
    "fmt"
    "math/rand"

    "github.com/go-a2a/adk-go/agent"
    "github.com/go-a2a/adk-go/model"
    "github.com/go-a2a/adk-go/tool/tools"
    "github.com/go-a2a/adk-go/types"
)

// Simple dice rolling function
func rollDice(ctx context.Context, sides int) (int, error) {
    if sides <= 0 {
        return 0, fmt.Errorf("dice must have at least 1 side")
    }
    return rand.Intn(sides) + 1, nil
}

func main() {
    ctx := context.Background()

    // Create model
    m, _ := model.NewGoogleModel("gemini-2.0-flash-exp")
    defer m.Close()

    // Create function tool
    diceTool := tools.NewFunctionTool("roll_dice", rollDice,
        tools.WithDescription("Roll a dice with specified number of sides"),
        tools.WithParameterDescription("sides", "Number of sides on the dice"),
    )

    // Create agent with tools
    agent := agent.NewLLMAgent(ctx, "game_master",
        agent.WithModel(m),
        agent.WithInstruction("You are a game master. Help users with dice rolls and games."),
        agent.WithTools(diceTool),
    )

    // ... run agent similar to above
}
```

### Multi-Agent System

```go
package main

import (
    "context"

    "github.com/go-a2a/adk-go/agent"
    "github.com/go-a2a/adk-go/model"
)

func main() {
    ctx := context.Background()
    m, _ := model.NewGoogleModel("gemini-2.0-flash-exp")
    defer m.Close()

    // Create specialized agents
    researcher := agent.NewLLMAgent(ctx, "researcher",
        agent.WithModel(m),
        agent.WithInstruction("You are a research specialist. Gather and analyze information."),
    )

    writer := agent.NewLLMAgent(ctx, "writer",
        agent.WithModel(m),
        agent.WithInstruction("You are a content writer. Create compelling content based on research."),
    )

    reviewer := agent.NewLLMAgent(ctx, "reviewer",
        agent.WithModel(m),
        agent.WithInstruction("You are a content reviewer. Ensure quality and accuracy."),
    )

    // Create sequential workflow
    workflow := agent.NewSequentialAgent("content_pipeline",
        agent.WithSubAgents(researcher, writer, reviewer),
        agent.WithDescription("Complete content creation pipeline"),
    )

    // ... run workflow
}
```

## 🏗️ Architecture

ADK Go follows a hierarchical, event-driven architecture with strong type safety and extensibility:

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    Agent        │    │     Model       │    │     Tools       │
│   System        │◄──►│     Layer       │◄──►│   Ecosystem     │
│                 │    │                 │    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ • LLMAgent      │    │ • Google Gemini │    │ • Function Tools│
│ • Sequential    │    │ • Anthropic     │    │ • Agent Tools   │
│ • Parallel      │    │ • Multi-provider│    │ • Auth Tools    │
│ • Loop          │    │ • Streaming     │    │ • Toolsets      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Flow        │    │     Event       │    │    Session      │
│  Management     │    │    System       │    │  Management     │
│                 │    │                 │    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ • LLMFlow       │    │ • Streaming     │    │ • State Mgmt    │
│ • AutoFlow      │    │ • Real-time     │    │ • Memory        │
│ • SingleFlow    │    │ • Event Actions │    │ • Persistence   │
│ • Processors    │    │ • Deltas        │    │ • Three-tier    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Agent Types

- **LLMAgent**: Full-featured agents powered by language models with tools, instructions, callbacks, planners, and code execution
- **SequentialAgent**: Executes sub-agents one after another, supports live mode with `taskCompleted()` flow control
- **ParallelAgent**: Runs sub-agents concurrently in isolated branches, merges event streams
- **LoopAgent**: Repeatedly executes sub-agents until escalation or max iterations

### Key Patterns

1. **Event-Driven Streaming**: All operations use `iter.Seq2[*Event, error]` for real-time processing
2. **Hierarchical Composition**: Agents form trees with parent/child relationships  
3. **Interface-Driven Design**: Core abstractions in `types/` enable extensibility
4. **Functional Options**: Configuration via `WithXxx()` functions
5. **Context Propagation**: Rich context flows through all operations
6. **Type Safety with Flexibility**: Strong typing while supporting dynamic LLM interactions
7. **Resource Management**: Proper cleanup with Close() methods throughout

## 🔧 Core Components

### Agent System (`agent/`)

```go
// Create different agent types
llmAgent := agent.NewLLMAgent(ctx, "assistant", ...)
seqAgent := agent.NewSequentialAgent("workflow", ...)
parAgent := agent.NewParallelAgent("concurrent", ...)
loopAgent := agent.NewLoopAgent("repeater", ...)
```

### Model Layer (`model/`)

```go
// Multi-provider support
gemini, _ := model.NewGoogleModel("gemini-2.0-flash-exp")
claude, _ := model.NewAnthropicModel("claude-3-5-sonnet-20241022")

// Registry pattern
model.RegisterModel("custom-model", customModelFactory)
m, _ := model.GetModel("custom-model")
```

### Tool System (`tool/`)

```go
// Function tools with automatic declaration generation
tool := tools.NewFunctionTool("my_function", myFunc,
    tools.WithDescription("Description of the function"),
    tools.WithParameterDescription("param", "Parameter description"),
)

// Custom tools
type CustomTool struct {
    *tool.Tool
}

func (t *CustomTool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
    // Tool implementation
    return result, nil
}
```

### Memory System (`memory/`)

```go
// In-memory storage
memService := memory.NewInMemoryService()

// Vertex AI RAG (future)
ragService := memory.NewVertexAIRAGService(projectID, location)

// Store and retrieve memories
memService.AddSession(ctx, sessionID, "Important information")
memories, _ := memService.SearchMemories(ctx, "search query")
```

### Session Management (`session/`)

```go
// Three-tier state management
state := map[string]any{
    "app:theme":        "dark",      // Application-wide
    "user:preference":  "verbose",   // User-specific
    "temp:calculation": 42,          // Session-temporary
}

sessionService := session.NewInMemoryService()
sess, _ := sessionService.CreateSession(ctx, appName, userID, sessionID, state)
```

## 📚 Examples

### Code Execution Agent

```go
codeAgent := agent.NewLLMAgent(ctx, "coder",
    agent.WithModel(model),
    agent.WithInstruction("You are a coding assistant. Write and execute code to solve problems."),
    agent.WithCodeExecutor(codeexecutor.NewBuiltInExecutor()), // Use model's native execution
)
```

### Authentication-Enabled Tool

```go
type APITool struct {
    *tool.Tool
}

func (t *APITool) Run(ctx context.Context, args map[string]any, toolCtx *types.ToolContext) (any, error) {
    // Request API key if not available
    if !toolCtx.HasCredential("api_key") {
        toolCtx.RequestCredential("api_key", &types.AuthConfig{
            Type: types.AuthTypeAPIKey,
            Description: "API key for external service",
        })
        return "Please provide your API key", nil
    }
    
    // Use the API key
    apiKey := toolCtx.GetCredential("api_key")
    // ... make API call
}
```

### Streaming with Error Handling

```go
for event, err := range agent.Run(ctx, ictx) {
    if err != nil {
        if errors.Is(err, context.Canceled) {
            log.Println("Operation canceled")
            break
        }
        log.Printf("Error: %v", err)
        continue
    }
    
    // Process different event types
    switch {
    case event.Message != nil:
        fmt.Printf("Message: %s\n", event.Message.Text)
    case event.ToolCall != nil:
        fmt.Printf("Tool call: %s\n", event.ToolCall.Name)
    case event.Actions != nil && event.Actions.StateDelta != nil:
        fmt.Printf("State update: %+v\n", event.Actions.StateDelta)
    }
}
```

## 🧪 Testing

Run tests with API keys:

```bash
# Set API keys
export GEMINI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"

# Run all tests
go test ./...

# Run specific tests
go test ./agent -run TestLLMAgent

# With coverage
go test -cover ./...
```

## 🔨 Build Commands

```bash
# Build
go build ./...

# Lint
go vet ./...

# Format
gofmt -w .
```

### API Reference

- **Agents**: Core agent interfaces and implementations (`agent/`)
- **Flow**: Request/response processing pipelines (`flow/`)
- **Memory**: Long-term storage and retrieval systems (`memory/`)
- **Models**: LLM provider integrations and abstractions (`model/`)
- **Session**: Conversation and state management (`session/`)
- **Tools**: Extensible tool system with function declarations (`tool/`)
- **Types**: Core interfaces and type definitions (`types/`)

## 🛠️ Development

### Project Structure

```
adk-go/
├── agent/           # Agent implementations (LLM, Sequential, Parallel, Loop)
├── artifact/        # Artifact storage services (GCS, in-memory)
├── codeexecutor/    # Code execution backends (built-in, container, local)
├── example/         # Example implementations and utilities
├── flow/            # LLM processing pipelines and flows
├── memory/          # Memory storage systems (in-memory, Vertex AI RAG)
├── model/           # LLM provider integrations (Gemini, Claude, registry)
├── planner/         # Strategic planning components (built-in, ReAct)
├── session/         # Session management and state tracking
├── tool/            # Tool framework and implementations
├── types/           # Core interfaces and type definitions
├── internal/        # Internal utilities (pool, iterators, maps)
└── third_party/     # Python ADK reference implementation
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and linting
6. Submit a pull request

### Code Style

- Use Go 1.24+ features including generics
- Follow standard Go conventions
- Use `any` instead of `interface{}`
- Include copyright headers in all files:
  ```go
  // Copyright 2025 The Go A2A Authors
  // SPDX-License-Identifier: Apache-2.0
  ```
- Write comprehensive tests with `github.com/google/go-cmp`

## 🤝 Community & Support

- **Issues**: [GitHub Issues](https://github.com/go-a2a/adk-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/go-a2a/adk-go/discussions)

## 🔗 Related Projects

This is a Go implementation of the [Agent Development Kit (ADK)](https://github.com/google/adk-python), a toolkit for building, evaluating, and deploying sophisticated AI agents. ADK-Go follows the same architectural principles as the Python implementation, but with Go's strengths of type safety, performance, and concurrency.

## 🙏 Acknowledgments

- Built on top of [google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai) for unified LLM integration
- Inspired by the [Agent Development Kit for Python](https://github.com/googleapis/agent-development-kit)

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
