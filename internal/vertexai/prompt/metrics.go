// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"sync"
	"time"
)

// MetricsCollector collects and tracks usage metrics for the prompts service.
type MetricsCollector struct {
	// Operation counters
	promptsCreated   int64
	promptsRetrieved int64
	promptsUpdated   int64
	promptsDeleted   int64
	promptsListed    int64

	// Template operations
	templatesApplied int64
	variablesApplied int64

	// Version operations
	versionsCreated  int64
	versionsRestored int64

	// Error counters
	validationErrors int64
	templateErrors   int64
	cloudErrors      int64

	// Performance metrics
	totalLatency   time.Duration
	operationCount int64

	// Cache metrics
	cacheHits   int64
	cacheMisses int64

	// Timestamp tracking
	startTime time.Time
	lastReset time.Time

	// Thread safety
	mu sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	now := time.Now()
	return &MetricsCollector{
		startTime: now,
		lastReset: now,
	}
}

// Operation metrics

// IncrementPromptCreated increments the prompt created counter.
func (mc *MetricsCollector) IncrementPromptCreated() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.promptsCreated++
}

// IncrementPromptRetrieved increments the prompt retrieved counter.
func (mc *MetricsCollector) IncrementPromptRetrieved() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.promptsRetrieved++
}

// IncrementPromptUpdated increments the prompt updated counter.
func (mc *MetricsCollector) IncrementPromptUpdated() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.promptsUpdated++
}

// IncrementPromptDeleted increments the prompt deleted counter.
func (mc *MetricsCollector) IncrementPromptDeleted() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.promptsDeleted++
}

// IncrementPromptsListed increments the prompts listed counter.
func (mc *MetricsCollector) IncrementPromptsListed(count int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.promptsListed += count
}

// IncrementTemplateApplied increments the template applied counter.
func (mc *MetricsCollector) IncrementTemplateApplied() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.templatesApplied++
}

// IncrementVariablesApplied increments the variables applied counter.
func (mc *MetricsCollector) IncrementVariablesApplied(count int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.variablesApplied += count
}

// IncrementVersionCreated increments the version created counter.
func (mc *MetricsCollector) IncrementVersionCreated() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.versionsCreated++
}

// IncrementVersionRestored increments the version restored counter.
func (mc *MetricsCollector) IncrementVersionRestored() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.versionsRestored++
}

// Error metrics

// IncrementValidationError increments the validation error counter.
func (mc *MetricsCollector) IncrementValidationError() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.validationErrors++
}

// IncrementTemplateError increments the template error counter.
func (mc *MetricsCollector) IncrementTemplateError() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.templateErrors++
}

// IncrementCloudError increments the cloud error counter.
func (mc *MetricsCollector) IncrementCloudError() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cloudErrors++
}

// Performance metrics

// RecordOperationLatency records the latency of an operation.
func (mc *MetricsCollector) RecordOperationLatency(duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.totalLatency += duration
	mc.operationCount++
}

// GetAverageLatency returns the average operation latency.
func (mc *MetricsCollector) GetAverageLatency() time.Duration {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	if mc.operationCount == 0 {
		return 0
	}
	return mc.totalLatency / time.Duration(mc.operationCount)
}

// Cache metrics

// IncrementCacheHit increments the cache hit counter.
func (mc *MetricsCollector) IncrementCacheHit() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cacheHits++
}

// IncrementCacheMiss increments the cache miss counter.
func (mc *MetricsCollector) IncrementCacheMiss() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cacheMisses++
}

// GetCacheHitRatio returns the cache hit ratio.
func (mc *MetricsCollector) GetCacheHitRatio() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	total := mc.cacheHits + mc.cacheMisses
	if total == 0 {
		return 0.0
	}
	return float64(mc.cacheHits) / float64(total)
}

// Metrics retrieval

// GetOperationMetrics returns operation-related metrics.
func (mc *MetricsCollector) GetOperationMetrics() map[string]int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]int64{
		"prompts_created":   mc.promptsCreated,
		"prompts_retrieved": mc.promptsRetrieved,
		"prompts_updated":   mc.promptsUpdated,
		"prompts_deleted":   mc.promptsDeleted,
		"prompts_listed":    mc.promptsListed,
		"templates_applied": mc.templatesApplied,
		"variables_applied": mc.variablesApplied,
		"versions_created":  mc.versionsCreated,
		"versions_restored": mc.versionsRestored,
	}
}

// GetErrorMetrics returns error-related metrics.
func (mc *MetricsCollector) GetErrorMetrics() map[string]int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]int64{
		"validation_errors": mc.validationErrors,
		"template_errors":   mc.templateErrors,
		"cloud_errors":      mc.cloudErrors,
	}
}

// GetPerformanceMetrics returns performance-related metrics.
func (mc *MetricsCollector) GetPerformanceMetrics() map[string]any {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]any{
		"total_operations":   mc.operationCount,
		"total_latency_ms":   mc.totalLatency.Milliseconds(),
		"average_latency_ms": mc.GetAverageLatency().Milliseconds(),
		"cache_hits":         mc.cacheHits,
		"cache_misses":       mc.cacheMisses,
		"cache_hit_ratio":    mc.GetCacheHitRatio(),
	}
}

// GetAllMetrics returns all collected metrics.
func (mc *MetricsCollector) GetAllMetrics() map[string]any {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	uptime := time.Since(mc.startTime)
	timeSinceReset := time.Since(mc.lastReset)

	metrics := map[string]any{
		"uptime_seconds":           uptime.Seconds(),
		"time_since_reset_seconds": timeSinceReset.Seconds(),
		"operations":               mc.GetOperationMetrics(),
		"errors":                   mc.GetErrorMetrics(),
		"performance":              mc.GetPerformanceMetrics(),
	}

	return metrics
}

// Reset resets all metrics counters.
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Reset operation counters
	mc.promptsCreated = 0
	mc.promptsRetrieved = 0
	mc.promptsUpdated = 0
	mc.promptsDeleted = 0
	mc.promptsListed = 0
	mc.templatesApplied = 0
	mc.variablesApplied = 0
	mc.versionsCreated = 0
	mc.versionsRestored = 0

	// Reset error counters
	mc.validationErrors = 0
	mc.templateErrors = 0
	mc.cloudErrors = 0

	// Reset performance metrics
	mc.totalLatency = 0
	mc.operationCount = 0

	// Reset cache metrics
	mc.cacheHits = 0
	mc.cacheMisses = 0

	// Update reset timestamp
	mc.lastReset = time.Now()
}

// GetUptime returns the service uptime.
func (mc *MetricsCollector) GetUptime() time.Duration {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return time.Since(mc.startTime)
}

// GetTimeSinceReset returns the time since the last metrics reset.
func (mc *MetricsCollector) GetTimeSinceReset() time.Duration {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return time.Since(mc.lastReset)
}

// GetOperationsPerSecond returns the average operations per second.
func (mc *MetricsCollector) GetOperationsPerSecond() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	timeSinceReset := time.Since(mc.lastReset).Seconds()
	if timeSinceReset == 0 {
		return 0.0
	}

	totalOps := mc.promptsCreated + mc.promptsRetrieved + mc.promptsUpdated +
		mc.promptsDeleted + mc.templatesApplied + mc.versionsCreated + mc.versionsRestored

	return float64(totalOps) / timeSinceReset
}

// GetErrorRate returns the error rate as a percentage.
func (mc *MetricsCollector) GetErrorRate() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	totalErrors := mc.validationErrors + mc.templateErrors + mc.cloudErrors
	totalOps := mc.operationCount

	if totalOps == 0 {
		return 0.0
	}

	return float64(totalErrors) / float64(totalOps) * 100.0
}

// MetricsSnapshot represents a point-in-time snapshot of metrics.
type MetricsSnapshot struct {
	Timestamp time.Time      `json:"timestamp"`
	Uptime    time.Duration  `json:"uptime"`
	Metrics   map[string]any `json:"metrics"`
}

// GetSnapshot returns a snapshot of current metrics.
func (mc *MetricsCollector) GetSnapshot() *MetricsSnapshot {
	return &MetricsSnapshot{
		Timestamp: time.Now(),
		Uptime:    mc.GetUptime(),
		Metrics:   mc.GetAllMetrics(),
	}
}

// PerformanceTracker helps track operation performance.
type PerformanceTracker struct {
	metrics   *MetricsCollector
	startTime time.Time
	operation string
}

// StartOperation starts tracking performance for an operation.
func (mc *MetricsCollector) StartOperation(operation string) *PerformanceTracker {
	return &PerformanceTracker{
		metrics:   mc,
		startTime: time.Now(),
		operation: operation,
	}
}

// Finish completes the performance tracking and records the latency.
func (pt *PerformanceTracker) Finish() {
	duration := time.Since(pt.startTime)
	pt.metrics.RecordOperationLatency(duration)
}

// FinishWithError completes the performance tracking and records an error.
func (pt *PerformanceTracker) FinishWithError(errorType string) {
	duration := time.Since(pt.startTime)
	pt.metrics.RecordOperationLatency(duration)

	switch errorType {
	case "validation":
		pt.metrics.IncrementValidationError()
	case "template":
		pt.metrics.IncrementTemplateError()
	case "cloud":
		pt.metrics.IncrementCloudError()
	}
}
