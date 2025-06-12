// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package tools

// LongRunningFunctionTool represents a function tool that returns the result asynchronously.
//
// This tool is used for long-running operations that may take a significant
// amount of time to complete. The framework will call the function. Once the
// function returns, the response will be returned asynchronously to the
// framework which is identified by the function_call_id.
//
// Example:
//
//	tool = LongRunningFunctionTool(a_long_running_function)
type LongRunningFunctionTool struct {
	*FunctionTool
}

// NewLongRunningFunctionTool returns the new [LongRunningFunctionTool] with the given function.
func NewLongRunningFunctionTool(fn Function) *LongRunningFunctionTool {
	t := &LongRunningFunctionTool{
		FunctionTool: NewFunctionTool(fn),
	}
	t.FunctionTool.Tool.SetLongRunning(true)
	return t
}
