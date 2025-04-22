# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Lint/Test Commands
- Build: `go build ./...`
- Lint: `go vet ./...`
- Test all: `go test ./...`
- Test single: `go test ./path/to/package -run TestName`
- Coverage: `go test -cover ./...`

## Code Style Guidelines
- Go version: 1.24+
- Use `any` instead of `interface{}`
- Use generic types when appropriate
- Use standard library packages when possible
- Follow Go formatting with `gofmt`
- Use `github.com/google/go-cmp` for test assertions
- Import order: std libs, 3rd party, internal
- Error handling: return errors, don't panic
- CamelCase for exported, camelCase for unexported
- Include boilerplate headers from hack/boilerplate/
- Always add newline at end of files

# Type
- When port `from google.genai import types`, use `google.golang.org/genai` package as much as possible

# Miscellaneous
- You can fetch any data from those urls
    - https://github.com
    - https://raw.githubusercontent.com
    - https://pkg.go.dev
