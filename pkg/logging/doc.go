// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

// Package logging provides context-based structured logging utilities using Go's standard slog package.
//
// The logging package implements a context-based logging pattern that allows loggers to be stored
// in and retrieved from context.Context values. This enables consistent logging throughout the
// application stack with automatic logger propagation.
//
// # Basic Usage
//
// Creating a logger context:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
//		Level: slog.LevelInfo,
//	}))
//
//	ctx := logging.NewContext(ctx, logger)
//
// Retrieving logger from context:
//
//	logger := logging.FromContext(ctx)
//	logger.Info("Operation completed", "duration", duration, "status", "success")
//
// # Integration with Agent System
//
// The logging package integrates with the agent system for consistent logging:
//
//	func (a *MyAgent) Execute(ctx context.Context, ictx *types.InvocationContext) iter.Seq2[*types.Event, error] {
//		logger := logging.FromContext(ctx)
//		logger.Info("Agent execution started", "agent", a.Name())
//
//		// ... execution logic
//
//		logger.Info("Agent execution completed")
//	}
//
// # Default Behavior
//
// When no logger is found in the context, FromContext returns a default JSON logger
// that writes to stdout with INFO level logging. This ensures logging always works
// even when no explicit logger is configured.
//
// # Structured Logging
//
// The package leverages Go's slog for structured logging with key-value pairs:
//
//	logger := logging.FromContext(ctx)
//	logger.Info("Request processed",
//		"method", "POST",
//		"path", "/api/v1/agents",
//		"duration", duration,
//		"status_code", 200,
//	)
//
// # Logger Configuration
//
// Configure loggers with different handlers and options:
//
//	// JSON handler for production
//	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
//		Level: slog.LevelInfo,
//		AddSource: true,
//	})
//
//	// Text handler for development
//	textHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
//		Level: slog.LevelDebug,
//	})
//
//	logger := slog.New(jsonHandler)
//	ctx := logging.NewContext(ctx, logger)
//
// # Context Propagation
//
// Loggers automatically propagate through context chains:
//
//	func parentFunction(ctx context.Context) {
//		logger := slog.New(handler)
//		ctx = logging.NewContext(ctx, logger)
//
//		// Logger is available in child functions
//		childFunction(ctx)
//	}
//
//	func childFunction(ctx context.Context) {
//		logger := logging.FromContext(ctx) // Same logger from parent
//		logger.Debug("Child function called")
//	}
//
// # Best Practices
//
//  1. Set up logging context early in request/operation lifecycle
//  2. Use structured logging with consistent key names
//  3. Include relevant context (user ID, session ID, request ID) in log messages
//  4. Use appropriate log levels (Debug, Info, Warn, Error)
//  5. Don't log sensitive information (passwords, tokens, PII)
//
// # Thread Safety
//
// The logging package is safe for concurrent use. Multiple goroutines can safely
// access loggers from context without additional synchronization.
package logging
