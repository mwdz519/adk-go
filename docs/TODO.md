# TODO: Missing Implementations from Python ADK

Based on comprehensive comparison with Python ADK (google/adk-python), the following components are missing in Go ADK:

## Critical Priority

### 1. A2A Protocol Support ⚠️
- [ ] Complete A2A (Agent-to-Agent) protocol implementation (`a2a/` package)
- [ ] Protocol converters for events, parts, and requests
- [ ] A2A agent executor for remote agent execution
- [ ] Task result aggregator for distributed agent systems
- [ ] A2A logging utilities
- [ ] Remote agent communication handlers
- [ ] Serialization/deserialization for network transport
- [ ] Security and authentication for remote agents

### 2. Evaluation System ✅ (Partially Implemented)
- [ ] Core evaluation framework (`eval/` package)
- [ ] EvalCase and EvalSet types
- [ ] Response evaluators (exact match, pattern, contains)
- [ ] Trajectory evaluators (tool usage, order, arguments)
- [ ] Metrics system (latency, tokens, errors, completeness)
- [ ] Session-to-eval conversion utilities
- [ ] Eval runner with parallel execution support
- [ ] History tracking and comparison
- [ ] Fix compilation errors with session types
- [ ] Integration with CLI commands (`adk eval`)
- [ ] JSON file format support (.evalset.json)
- [ ] Google Cloud Storage (GCS) evaluation managers
- [ ] Evaluation generator for creating test cases

## High Priority

### 3. Tool Integrations
- [ ] BigQuery toolset (`tool/tools/bigquery/`)
- [ ] Google API tools (`tool/tools/google_api_tool/`)
- [ ] API Hub integration (`tool/tools/apihub_tool/`)
- [ ] Enterprise Search tool
- [ ] OpenAPI specification tools
- [ ] MCP (Model Context Protocol) tools
- [ ] Application Integration tools
- [ ] LangGraph agent integration
- [ ] CrewAI tool integration


| Name                                | Status | Description |
|:------------------------------------|:------:|:-----------:|
| apihub_tool/                         |       |             |
| application_integration_tool/        |       |             |
| bigquery/                            |       |             |
| google_api_tool/                     |       |             |
| mcp_tool/                            |       |             |
| openapi_tool/                        |       |             |
| retrieval/                           |       |             |
| _automatic_function_calling_util.py  |  ✔️   |             |
| _forwarding_artifact_service.py      |  ✔️   |             |
| _function_parameter_parse_util.py    |       |             |
| _gemini_schema_util.py               |       |             |
| _memory_entry_utils.py               |       |             |
| agent_tool.py                        |  ✔️   |             |
| base_tool.py                         |  ✔️   |             |
| base_toolset.py                      |  ✔️   |             |
| crewai_tool.py                       |       |             |
| enterprise_search_tool.py            |       |             |
| example_tool.py                      |  ✔️   |             |
| exit_loop_tool.py                    |  ✔️   |             |
| function_tool.py                     |  ✔️   |             |
| get_user_choice_tool.py              |       |             |
| google_search_tool.py                |       |             |
| langchain_tool.py                    |       |             |
| load_artifacts_tool.py               |       |             |
| load_memory_tool.py                  |       |             |
| load_web_page.py                     |       |             |
| long_running_tool.py                 |       |             |
| preload_memory_tool.py               |       |             |
| tool_context.py                      |       |             |
| toolbox_toolset.py                   |       |             |
| transfer_to_agent_tool.py            |       |             |
| url_context_tool.py                  |       |             |
| vertex_ai_search_tool.py             |       |             |


### 4. Advanced Vertex AI Integration
- [ ] Vertex AI Memory Bank Service
- [ ] Vertex AI RAG Memory Service (beyond current placeholder)
- [ ] Vertex AI Session Service
- [ ] Vertex AI Example Store
- [ ] Vertex AI RAG Retrieval
- [ ] Vertex AI-specific code executor
- [ ] Agent deployment helpers for Vertex AI
- [ ] Integration with Vertex AI Model Garden

## Medium Priority

### 5. Runners Module
- [ ] Centralized Runner class for agent execution management
- [ ] InMemoryRunner for lightweight testing/development
- [ ] Agent selection logic for session continuation
- [ ] Streaming tool context management
- Note: Go uses distributed execution model which works well, but Python's centralized approach has benefits for certain use cases

### 6. Live Mode and Streaming Features
- [ ] Active streaming tool support
- [ ] Live request queue management
- [ ] Transcription entry support for audio
- [ ] Enhanced audio/video streaming capabilities

### 7. Example Management System
- [ ] Example provider framework
- [ ] Vertex AI example store integration
- [ ] Dynamic example loading and management

## Low Priority

### 8. Deployment and API Server Support
- [ ] FastAPI-style wrapper for agents (library support only)
- [ ] HTTP API server scaffolding utilities
- [ ] Cloud Run deployment helper libraries
- [ ] Container build utilities (without CLI dependencies)
- [ ] Agent graph visualization library
- Note: Excluded CLI tools as requested, but library support could be useful

### 9. Development Utilities
- [ ] Interactive session recording utilities
- [ ] Debug mode with detailed tracing helpers
- [ ] Performance profiling helper libraries
- [ ] Agent visualization tools (library components)
- [ ] Dynamic agent loader utilities
- [ ] Environment management helpers

### 10. Authentication Enhancements
- [ ] OAuth2 utility helpers
- [ ] Credential exchanger mechanisms
- [ ] Token refresher implementations
- [ ] Additional auth scheme support

### 11. Multi-Provider Support
- [ ] LiteLLM integration for multi-provider support
- [ ] Additional LLM provider integrations

## Architectural Differences (No Action Needed)

These are intentional design differences between Python and Go implementations:

1. **Execution Model**: Python uses centralized runners, Go uses distributed execution
2. **Event System**: Python has separate events package, Go integrates into types
3. **Type Safety**: Go provides stronger type safety with generics
4. **Agent Types**: Both have similar agent types (Sequential, Parallel, Loop)
5. **Concurrency**: Go uses native goroutines vs Python's threading/asyncio
6. **Package Structure**: Go has cleaner separation of concerns

## Go ADK Advantages

Go ADK actually has unique features and advantages:

1. **Better Type Safety**: Full use of Go generics and strong typing
2. **Performance**: Native compilation and superior concurrency
3. **Cleaner Architecture**: Better separation of concerns in package structure
4. **Python Compatibility**: types/py/ package for Python pattern support
5. **AI Platform Conversions**: types/aiconv/ for platform-specific conversions
6. **Internal Utilities**: Advanced pool, iterator, and map utilities
7. **More Comprehensive Tools**: Some tools like get_user_choice, url_context exist only in Go

## Implementation Priority Recommendations

1. **Immediate** (blocking for production use):
   - A2A Protocol Support (critical for distributed systems)
   - Fix Evaluation System compilation issues
   - BigQuery and Google API tool integrations

2. **Short Term** (significant feature gaps):
   - Enhanced Vertex AI integration features
   - Enterprise Search and OpenAPI tools
   - Active streaming and live mode features

3. **Medium Term** (nice to have):
   - Centralized runner option (alongside distributed model)
   - Example management system
   - Development and debugging utilities

4. **Long Term** (evaluate based on demand):
   - LangGraph and CrewAI integrations
   - Deployment helper libraries
   - Additional LLM provider support via LiteLLM

## Next Steps

1. Fix compilation issues in eval package (session type references)
2. Design and implement A2A protocol support for distributed agent systems
3. Prioritize tool integrations based on user requirements
4. Enhance Vertex AI integration to match Python capabilities
5. Consider implementing optional centralized runner for compatibility
