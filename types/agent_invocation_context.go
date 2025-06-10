// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

// LLMCallsLimitExceededError represents error thrown when the number of LLM calls exceed the limit.
type LLMCallsLimitExceededError string

// NewLLMCallsLimitExceededError returns the new [LLMCallsLimitExceededError] error.
func NewLLMCallsLimitExceededError(msg string, a ...any) error {
	return LLMCallsLimitExceededError(fmt.Sprintf(msg, a...))
}

// Error returns a string representation of the LLMCallsLimitExceededError.
func (e LLMCallsLimitExceededError) Error() string {
	return string(e)
}

// InvocationCostManager represents a container to keep track of the cost of invocation.
//
// While we don't expected the metrics captured here to be a direct
// representatative of monetary cost incurred in executing the current
// invocation, but they, in someways have an indirect affect.
type InvocationCostManager struct {
	// A counter that keeps track of number of llm calls made.
	llmCalls int
}

// IncrementAndEnforceLLMCallsLimit increments llmCalls and enforces the limit.
func (mgr *InvocationCostManager) IncrementAndEnforceLLMCallsLimit(runConfig *RunConfig) error {
	mgr.llmCalls++
	if runConfig != nil {
		if runConfig.MaxLLMCalls > 0 && mgr.llmCalls > runConfig.MaxLLMCalls {
			return NewLLMCallsLimitExceededError("max number of llm calls limit of %d exceeded", runConfig.MaxLLMCalls)
		}
	}
	return nil
}

// InvocationContext an invocation context represents the data of a single invocation of an agent.
//
// An invocation:
//
//   - Starts with a user message and ends with a final response.
//   - Can contain one or multiple agent calls.
//   - Is handled by runner.run_async().
//
// An invocation runs an agent until it does not request to transfer to another
// agent.
//
// An agent call:
//
//   - Is handled by agent.run().
//   - Ends when agent.run() ends.
//
// An LLM agent call is an agent with a BaseLLMFlow.
//
// An LLM agent call can contain one or multiple steps.
//
// An LLM agent runs steps in a loop until:
//
//   - A final response is generated.
//   - The agent transfers to another agent.
//   - The end_invocation is set to true by any callbacks or tools.
//
// A step:
//
//   - Calls the LLM only once and yields its response.
//   - Calls the tools and yields their responses if requested.
//
// The summarization of the function response is considered another step, since
// it is another llm call.
// A step ends when it's done calling llm and tools, or if the end_invocation
// is set to true at any time.
//
//	┌─────────────────────── invocation ──────────────────────────┐
//	┌──────────── llm_agent_call_1 ────────────┐ ┌─ agent_call_2 ─┐
//	┌──── step_1 ────────┐ ┌───── step_2 ──────┐
//	[call_llm] [call_tool] [call_llm] [transfer]
type InvocationContext struct {
	ArtifactService ArtifactService
	SessionService  SessionService
	MemoryService   MemoryService

	// InvocationID is the id of this invocation context. Readonly.
	InvocationID string

	// The branch of the invocation context.
	//
	// The format is like agent_1.agent_2.agent_3, where agent_1 is the parent of
	// agent_2, and agent_2 is the parent of agent_3.
	//
	// Branch is used when multiple sub-agents shouldn't see their peer agents'
	// conversation history.
	Branch string

	// The current agent of this invocation context. Readonly.
	Agent Agent

	// The user content that started this invocation. Readonly.
	UserContent *genai.Content

	// The current session of this invocation context. Readonly.
	Session Session

	// Whether to end this invocation.
	//
	// Set to True in callbacks or tools to terminate this invocation.
	EndInvocation bool

	// The queue to receive live requests.
	LiveRequestQueue *LiveRequestQueue

	// The running streaming tools of this invocation.
	ActiveStreamingTools map[string]*ActiveStreamingTool[any]

	// Caches necessary, data audio or contents, that are needed by transcription.
	TranscriptionCache []*TranscriptionEntry

	// Configurations for live agents under this invocation.
	RunConfig *RunConfig

	// A container to keep track of different kinds of costs incurred as a part
	// of this invocation.
	invocationCostManager *InvocationCostManager
}

// InvocationContextOption is a function that modifies the [InvocationContext].
type InvocationContextOption func(*InvocationContext)

func WithArtifactService(svc ArtifactService) InvocationContextOption {
	return func(ictx *InvocationContext) {
		ictx.ArtifactService = svc
	}
}

func WithMemoryService(svc MemoryService) InvocationContextOption {
	return func(ictx *InvocationContext) {
		ictx.MemoryService = svc
	}
}

func WithBranch(branch string) InvocationContextOption {
	return func(ictx *InvocationContext) {
		ictx.Branch = branch
	}
}

func WithUserContent(content *genai.Content) InvocationContextOption {
	return func(ictx *InvocationContext) {
		ictx.UserContent = content
	}
}

func WithLiveRequestQueue(liveRequestQueue *LiveRequestQueue) InvocationContextOption {
	return func(ictx *InvocationContext) {
		ictx.LiveRequestQueue = liveRequestQueue
	}
}

func WithActiveStreamingTools[T any](activeStreamingTools map[string]*ActiveStreamingTool[any]) InvocationContextOption {
	return func(ictx *InvocationContext) {
		ictx.ActiveStreamingTools = activeStreamingTools
	}
}

func WithTranscriptionCache(entries ...*TranscriptionEntry) InvocationContextOption {
	return func(ictx *InvocationContext) {
		ictx.TranscriptionCache = entries
	}
}

// NewInvocationContext creates a new [InvocationContext].
func NewInvocationContext(agent Agent, session Session, sessionSvc SessionService, opts ...InvocationContextOption) *InvocationContext {
	ictx := &InvocationContext{
		Agent:                 agent,
		invocationCostManager: &InvocationCostManager{},
		Session:               session,
		SessionService:        sessionSvc,
	}
	for _, opt := range opts {
		opt(ictx)
	}

	return ictx
}

// IncrementLLMCallCount tracks number of llm calls made.
func (ictx *InvocationContext) IncrementLLMCallCount() error {
	return ictx.invocationCostManager.IncrementAndEnforceLLMCallsLimit(ictx.RunConfig)
}

func (ictx *InvocationContext) AppName() string {
	return ictx.Session.AppName()
}

func (ictx *InvocationContext) UserID() string {
	return ictx.Session.UserID()
}

// NewInvocationContextID generates a new invocation context ID.
func NewInvocationContextID() string {
	return `e-` + uuid.NewString()
}
