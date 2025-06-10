// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"iter"
)

// Flow represents the basic interface that all flows must implement.
type Flow interface {
	// Run runs the flow with the given invocation context and returns a sequence of events.
	Run(ctx context.Context, ictx *InvocationContext) iter.Seq2[*Event, error]

	// RunLive runs the flow with the given invocation context in a live mode and returns a sequence of events.
	RunLive(ctx context.Context, ictx *InvocationContext) iter.Seq2[*Event, error]
}
