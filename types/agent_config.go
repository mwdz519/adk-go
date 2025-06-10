// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"log/slog"
)

// Config represents the configuration for an [types.Agent].
type Config struct {
	// The agent's Name.
	//
	// Agent Name must be a Go identifier and unique within the agent tree.
	// Agent Name cannot be "user", since it's reserved for end-user's input.
	Name string

	// Description about the agent's capability.
	//
	// The model uses this to determine whether to delegate control to the agent.
	// One-line Description is enough and preferred.
	Description string

	// The parent agent of this agent.
	//
	// Note that an agent can ONLY be added as sub-agent once.
	//
	// If you want to add one agent twice as sub-agent, consider to create two agent
	// instances with identical config, but with different name and add them to the
	// agent tree.
	parentAgent Agent

	// The sub-agents of this agent.
	subAgents []Agent

	// callback signature that is invoked before the agent run.
	beforeAgentCallbacks []AgentCallback

	// callback signature that is invoked after the agent run.
	afterAgentCallbacks []AgentCallback

	logger *slog.Logger
}

// Option configures a [Config].
type Option interface {
	apply(*Config)
}

type optionFunc func(*Config)

func (o optionFunc) apply(c *Config) { o(c) }

// WithParentAgent sets the parentAgent for the [Config].
func WithParentAgent(parentAgent Agent) Option {
	return optionFunc(func(c *Config) {
		c.parentAgent = parentAgent
	})
}

// WithSubAgents adds sub-agents for the [Config].
func WithSubAgents(agents ...Agent) Option {
	return optionFunc(func(c *Config) {
		c.subAgents = append(c.subAgents, agents...)
	})
}

func WithBeforeAgentCallbacks(callbacks ...AgentCallback) Option {
	return optionFunc(func(c *Config) {
		c.beforeAgentCallbacks = append(c.beforeAgentCallbacks, callbacks...)
	})
}

func WithAfterAgentCallbacks(callbacks ...AgentCallback) Option {
	return optionFunc(func(c *Config) {
		c.afterAgentCallbacks = append(c.afterAgentCallbacks, callbacks...)
	})
}

// WithLogger sets the logger for the [Config].
func WithLogger(logger *slog.Logger) Option {
	return optionFunc(func(c *Config) {
		c.logger = logger
	})
}

// NewConfig creates a new agent configuration with the given name.
func NewConfig(name string, opts ...Option) *Config {
	c := &Config{
		Name:   name,
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}

// AfterAgentCallbacks returns the callbacks that are invoked after the agent run.
func (c *Config) Logger() *slog.Logger {
	return c.logger
}
