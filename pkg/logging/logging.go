// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"log/slog"
	"os"
)

// contextKey is how we find [*slog.Logger] in a [context.Context].
type contextKey struct{}

// NewContext returns a new [context.Context], derived from ctx, which carries the provided [*slog.Logger].
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext returns a [slog.Logger] from ctx.
//
// If no [*slog.Logger] is found, this returns a logger with [slog.DiscardHandler].
func FromContext(ctx context.Context) *slog.Logger {
	if v := ctx.Value(contextKey{}); v != nil {
		return v.(*slog.Logger)
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
