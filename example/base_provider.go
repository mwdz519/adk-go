// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package example

import (
	"context"
)

// Provider represents a base interface for example providers.
//
// This type defines the interface for providing examples for a given query.
type Provider interface {
	GetExamples(ctx context.Context, query string) ([]*Example, error)
}
