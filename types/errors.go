// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

// NotImplementedError is the error type for unimplemented behaiviour.
type NotImplementedError string

// Error returns a string representation of the [NotImplementedError].
func (e NotImplementedError) Error() string {
	return string(e)
}
