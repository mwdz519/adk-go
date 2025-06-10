// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

// Toolset represents a base for toolset.
//
// A toolset is a collection of tools that can be used by an agent.
type Toolset interface {
	// GetTools returns the all tools in the toolset based on the provided context.
	GetTools(rctx *ReadOnlyContext) []Tool

	// Close performs cleanup and releases resources held by the toolset.
	//
	// NOTE: This method is invoked, for example, at the end of an agent server's
	// lifecycle or when the toolset is no longer needed. Implementations
	// should ensure that any open connections, files, or other managed
	// resources are properly released to prevent leaks.
	Close()
}
