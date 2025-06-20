// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package planner provides strategic planning capabilities for guiding agent execution and decision-making.
//
// The planner package implements the types.Planner interface to enable agents to generate structured
// plans for complex queries, improving their problem-solving capabilities through systematic thinking
// and step-by-step execution strategies.
//
// # Planning Strategies
//
// The package provides two main planning implementations:
//
//   - BuiltInPlanner: Leverages model's native thinking capabilities (Claude, Gemini 2.0+)
//   - PlanReActPlanner: Structured planning/reasoning/action framework with explicit tags
//
// # Built-In Planning
//
// BuiltInPlanner uses the model's built-in thinking features for strategic planning:
//
//	thinkingConfig := &genai.ThinkingConfig{
//		IncludeThinkingInResponse: false, // Hide thinking from final response
//	}
//
//	planner := planner.NewBuiltInPlanner(thinkingConfig)
//
//	agent := agent.NewLLMAgent(ctx, "strategic_agent",
//		agent.WithPlanner(planner),
//		agent.WithModel("claude-3-5-sonnet-20241022"),
//	)
//
// This approach leverages the model's internal reasoning capabilities for planning
// without requiring explicit planning prompts or structured formats.
//
// # ReAct Planning Framework
//
// PlanReActPlanner implements a structured Reasoning and Acting (ReAct) framework:
//
//	planner := planner.NewPlanReActPlanner()
//
//	agent := agent.NewLLMAgent(ctx, "reasoning_agent",
//		agent.WithPlanner(planner),
//		agent.WithTools(tools...),
//	)
//
// The ReAct framework follows this structured approach:
//
//  1. Planning: Generate a high-level plan to solve the problem
//  2. Action: Execute tools based on the plan
//  3. Reasoning: Analyze results and determine next steps
//  4. Iteration: Repeat action/reasoning until goal is achieved
//  5. Final Answer: Provide the complete solution
//
// # Planning Format Tags
//
// PlanReActPlanner uses structured tags to organize the planning process:
//
//	/*PLANNING*/
//	1. Search for current weather data in Paris
//	2. Look up any weather alerts or warnings
//	3. Provide a comprehensive weather summary
//	/*PLANNING*/
//
//	/*ACTION*/
//	weather_tool(city="Paris", country="France")
//	/*ACTION*/
//
//	/*REASONING*/
//	I retrieved the current weather for Paris. The temperature is 18°C with light rain.
//	I should now check for any weather alerts to provide a complete summary.
//	/*REASONING*/
//
//	/*ACTION*/
//	weather_alerts(location="Paris, France")
//	/*ACTION*/
//
//	/*FINAL_ANSWER*/
//	The weather in Paris is currently 18°C with light rain. There are no active weather alerts.
//	/*FINAL_ANSWER*/
//
// # Planning Process
//
// Planners integrate with the agent execution flow:
//
//  1. BuildPlanningInstruction: Adds planning directives to the LLM request
//  2. Model generates response with planning structure
//  3. ProcessPlanningResponse: Processes and extracts planned actions
//  4. Agent executes tools based on the plan
//  5. Results are fed back for iterative planning
//
// # Custom Planner Implementation
//
// Implement the Planner interface for custom planning strategies:
//
//	type CustomPlanner struct{}
//
//	func (p *CustomPlanner) BuildPlanningInstruction(ctx context.Context, rctx *types.ReadOnlyContext, request *types.LLMRequest) string {
//		return `Create a detailed step-by-step plan using available tools.
//		Consider dependencies between steps and potential failure modes.`
//	}
//
//	func (p *CustomPlanner) ProcessPlanningResponse(ctx context.Context, cctx *types.CallbackContext, responseParts []*genai.Part) []*genai.Part {
//		// Process and potentially modify the response parts
//		return responseParts
//	}
//
// # Planning Principles
//
// Effective planning follows these principles:
//
//  1. Decomposition: Break complex problems into manageable steps
//  2. Tool Awareness: Plan only uses available tools and capabilities
//  3. Dependency Management: Sequence steps based on dependencies
//  4. Error Handling: Include contingency plans for failures
//  5. Iterative Refinement: Adapt plans based on execution results
//
// # Integration with Tools
//
// Planners work closely with the tool system:
//
//	planner := planner.NewPlanReActPlanner()
//
//	agent := agent.NewLLMAgent(ctx, "research_agent",
//		agent.WithPlanner(planner),
//		agent.WithTools(
//			tools.NewWebSearchTool(),
//			tools.NewCalculatorTool(),
//			tools.NewDataAnalysisTool(),
//		),
//		agent.WithInstruction("Use tools systematically to research and analyze data"),
//	)
//
// The planner guides tool selection and sequencing for optimal problem solving.
//
// # Model Compatibility
//
// Planning strategies have different model requirements:
//
//   - BuiltInPlanner: Requires models with native thinking support (Claude 3.5+, Gemini 2.0+)
//   - PlanReActPlanner: Works with any model that can follow structured prompts
//
// Choose the appropriate planner based on your model capabilities and requirements.
//
// # Performance Considerations
//
// Planning impacts performance in several ways:
//   - Token Usage: Planning instructions increase prompt size
//   - Latency: Complex planning may increase response time
//   - Quality: Good planning significantly improves task success rates
//   - Iterative Overhead: ReAct framework may require multiple model calls
//
// Balance planning complexity with performance requirements for your use case.
//
// # Best Practices
//
//  1. Use BuiltInPlanner when supported by the model for efficiency
//  2. Use PlanReActPlanner for models without native thinking capabilities
//  3. Provide clear tool descriptions to enable better planning
//  4. Monitor planning effectiveness and adjust strategies as needed
//  5. Consider planning overhead vs. task complexity
//
// # Error Handling
//
// Planners should handle planning failures gracefully:
//
//	// Planning may fail if tools are insufficient or query is unclear
//	if planningFailed {
//		// Fall back to direct tool execution or request clarification
//		fallbackStrategy()
//	}
//
// The agent system provides error recovery mechanisms when planning encounters issues.
package planner
