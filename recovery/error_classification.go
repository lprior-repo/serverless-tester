// Package recovery provides sophisticated error recovery and retry mechanisms for AWS services.
//
// This package implements advanced retry patterns including exponential backoff with jitter,
// circuit breaker patterns, bulkhead isolation, and adaptive retry based on error types.
// It follows pure functional programming principles and provides deterministic error recovery.
//
// Key Features:
//   - Automatic error classification and handling
//   - Recovery strategy selection based on error context
//   - Rollback mechanisms for failed operations
//   - Error aggregation and correlation
//   - Circuit breaker patterns for AWS services
//   - Adaptive retry with progressive escalation
//   - Observability for error tracking and analytics
//
// Example usage:
//
//   func ExampleWithRecovery(ctx context.Context, operation func() error) error {
//       return ExecuteWithRecovery(ctx, RecoveryConfig{
//           OperationName: "example-operation",
//           MaxAttempts:   5,
//           BaseDelay:     100 * time.Millisecond,
//       }, operation)
//   }
package recovery

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// ErrorType represents the classification of an error
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeThrottling
	ErrorTypeServiceUnavailable
	ErrorTypeTimeout
	ErrorTypeAuthorization
	ErrorTypeValidation
	ErrorTypeNetwork
	ErrorTypeResource
	ErrorTypeInternal
)

// String returns the string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeThrottling:
		return "throttling"
	case ErrorTypeServiceUnavailable:
		return "service_unavailable"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeAuthorization:
		return "authorization"
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeResource:
		return "resource"
	case ErrorTypeInternal:
		return "internal"
	default:
		return "unknown"
	}
}

// RetryStrategy represents different retry approaches
type RetryStrategy int

const (
	RetryStrategyNone RetryStrategy = iota
	RetryStrategyLinear
	RetryStrategyExponentialBackoff
	RetryStrategyCircuitBreaker
	RetryStrategyAdaptive
	RetryStrategyBulkhead
)

// String returns the string representation of RetryStrategy
func (rs RetryStrategy) String() string {
	switch rs {
	case RetryStrategyNone:
		return "none"
	case RetryStrategyLinear:
		return "linear"
	case RetryStrategyExponentialBackoff:
		return "exponential_backoff"
	case RetryStrategyCircuitBreaker:
		return "circuit_breaker"
	case RetryStrategyAdaptive:
		return "adaptive"
	case RetryStrategyBulkhead:
		return "bulkhead"
	default:
		return "unknown"
	}
}

// ErrorClassification represents the analysis result of an error
type ErrorClassification struct {
	Type                ErrorType
	Retryable           bool
	RecommendedStrategy RetryStrategy
	Severity            int // 1-5, where 5 is most severe
	ExpectedDuration    time.Duration
	RequiresEscalation  bool
}

// RecoveryStrategy represents the selected recovery approach
type RecoveryStrategy struct {
	Type        RetryStrategy
	ShouldRetry bool
	Delay       time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Jitter      bool
	Metadata    map[string]interface{}
}

// RecoveryContext provides context for recovery decisions
type RecoveryContext struct {
	OperationName string
	Attempt       int
	MaxAttempts   int
	StartTime     time.Time
	LastError     error
	ServiceName   string
}

// ErrorAggregator collects and analyzes errors for pattern detection
type ErrorAggregator struct {
	operationName string
	errors        []AggregatedError
	mutex         sync.RWMutex
	startTime     time.Time
}

// AggregatedError represents an error with context
type AggregatedError struct {
	Step      string
	Error     error
	Timestamp time.Time
	Context   map[string]interface{}
}

// ErrorSummary provides aggregated error statistics
type ErrorSummary struct {
	OperationName string
	TotalErrors   int
	ErrorsByType  map[ErrorType]int
	ErrorsByStep  map[string][]error
	Duration      time.Duration
	Patterns      []ErrorPattern
}

// ErrorPattern represents a detected error pattern
type ErrorPattern struct {
	Type     ErrorType
	Count    int
	Services []string
	Severity int
}

// ClassifyError analyzes an error and returns its classification
func ClassifyError(err error) ErrorClassification {
	if err == nil {
		return ErrorClassification{
			Type:      ErrorTypeUnknown,
			Retryable: false,
		}
	}

	errMsg := strings.ToLower(err.Error())

	// Check for context deadline exceeded
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorClassification{
			Type:                ErrorTypeTimeout,
			Retryable:           true,
			RecommendedStrategy: RetryStrategyAdaptive,
			Severity:            3,
			ExpectedDuration:    5 * time.Second,
		}
	}

	// Check for throttling errors
	if containsAny(errMsg, []string{"throttling", "rate exceeded", "too many requests"}) {
		return ErrorClassification{
			Type:                ErrorTypeThrottling,
			Retryable:           true,
			RecommendedStrategy: RetryStrategyExponentialBackoff,
			Severity:            2,
			ExpectedDuration:    10 * time.Second,
		}
	}

	// Check for service unavailable errors
	if containsAny(errMsg, []string{"service unavailable", "temporarily unavailable", "internal server error"}) {
		return ErrorClassification{
			Type:                ErrorTypeServiceUnavailable,
			Retryable:           true,
			RecommendedStrategy: RetryStrategyCircuitBreaker,
			Severity:            4,
			ExpectedDuration:    30 * time.Second,
		}
	}

	// Check for authorization errors
	if containsAny(errMsg, []string{"unauthorized", "access denied", "permission denied", "forbidden"}) {
		return ErrorClassification{
			Type:                ErrorTypeAuthorization,
			Retryable:           false,
			RecommendedStrategy: RetryStrategyNone,
			Severity:            5,
			RequiresEscalation:  true,
		}
	}

	// Check for validation errors
	if containsAny(errMsg, []string{"validation", "invalid parameter", "malformed", "bad request"}) {
		return ErrorClassification{
			Type:                ErrorTypeValidation,
			Retryable:           false,
			RecommendedStrategy: RetryStrategyNone,
			Severity:            3,
		}
	}

	// Check for network errors
	if containsAny(errMsg, []string{"network", "connection", "timeout", "unreachable"}) {
		return ErrorClassification{
			Type:                ErrorTypeNetwork,
			Retryable:           true,
			RecommendedStrategy: RetryStrategyExponentialBackoff,
			Severity:            3,
			ExpectedDuration:    5 * time.Second,
		}
	}

	// Check for resource errors
	if containsAny(errMsg, []string{"not found", "does not exist", "already exists"}) {
		return ErrorClassification{
			Type:                ErrorTypeResource,
			Retryable:           false,
			RecommendedStrategy: RetryStrategyNone,
			Severity:            2,
		}
	}

	// Default classification
	return ErrorClassification{
		Type:                ErrorTypeInternal,
		Retryable:           true,
		RecommendedStrategy: RetryStrategyLinear,
		Severity:            3,
		ExpectedDuration:    5 * time.Second,
	}
}

// SelectRecoveryStrategy selects the appropriate recovery strategy based on error and context
func SelectRecoveryStrategy(ctx context.Context, err error, recoveryCtx RecoveryContext) RecoveryStrategy {
	classification := ClassifyError(err)
	
	// Log strategy selection for observability
	slog.Info("Selecting recovery strategy",
		"operation", recoveryCtx.OperationName,
		"attempt", recoveryCtx.Attempt,
		"maxAttempts", recoveryCtx.MaxAttempts,
		"errorType", classification.Type.String(),
		"retryable", classification.Retryable,
	)

	// Non-retryable errors
	if !classification.Retryable {
		return RecoveryStrategy{
			Type:        RetryStrategyNone,
			ShouldRetry: false,
		}
	}

	// Check if we've exceeded max attempts
	if recoveryCtx.Attempt >= recoveryCtx.MaxAttempts {
		return RecoveryStrategy{
			Type:        RetryStrategyNone,
			ShouldRetry: false,
		}
	}

	// Strategy escalation based on attempt count
	baseStrategy := classification.RecommendedStrategy
	if recoveryCtx.Attempt >= recoveryCtx.MaxAttempts-1 {
		// Near max attempts - use more aggressive strategy
		if baseStrategy == RetryStrategyLinear || baseStrategy == RetryStrategyExponentialBackoff {
			baseStrategy = RetryStrategyCircuitBreaker
		}
	}

	switch baseStrategy {
	case RetryStrategyExponentialBackoff:
		delay := calculateExponentialBackoffDelay(recoveryCtx.Attempt, 100*time.Millisecond, 2.0)
		return RecoveryStrategy{
			Type:        RetryStrategyExponentialBackoff,
			ShouldRetry: true,
			Delay:       delay,
			MaxDelay:    30 * time.Second,
			Multiplier:  2.0,
			Jitter:      true,
		}

	case RetryStrategyCircuitBreaker:
		return RecoveryStrategy{
			Type:        RetryStrategyCircuitBreaker,
			ShouldRetry: true,
			Delay:       5 * time.Second,
			Metadata: map[string]interface{}{
				"circuit_breaker": true,
				"service":         recoveryCtx.ServiceName,
			},
		}

	case RetryStrategyAdaptive:
		delay := calculateAdaptiveDelay(recoveryCtx, classification)
		return RecoveryStrategy{
			Type:        RetryStrategyAdaptive,
			ShouldRetry: true,
			Delay:       delay,
			Jitter:      true,
			Metadata: map[string]interface{}{
				"adaptive": true,
				"severity": classification.Severity,
			},
		}

	default:
		return RecoveryStrategy{
			Type:        RetryStrategyLinear,
			ShouldRetry: true,
			Delay:       1 * time.Second,
		}
	}
}

// NewErrorAggregator creates a new error aggregator for an operation
func NewErrorAggregator(operationName string) *ErrorAggregator {
	return &ErrorAggregator{
		operationName: operationName,
		errors:        make([]AggregatedError, 0),
		startTime:     time.Now(),
	}
}

// AddError adds an error to the aggregator
func (ea *ErrorAggregator) AddError(step string, err error) {
	ea.mutex.Lock()
	defer ea.mutex.Unlock()

	ea.errors = append(ea.errors, AggregatedError{
		Step:      step,
		Error:     err,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	})
}

// GetSummary returns an aggregated summary of all errors
func (ea *ErrorAggregator) GetSummary() ErrorSummary {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()

	summary := ErrorSummary{
		OperationName: ea.operationName,
		TotalErrors:   len(ea.errors),
		ErrorsByType:  make(map[ErrorType]int),
		ErrorsByStep:  make(map[string][]error),
		Duration:      time.Since(ea.startTime),
		Patterns:      ea.detectPatterns(),
	}

	for _, aggErr := range ea.errors {
		classification := ClassifyError(aggErr.Error)
		summary.ErrorsByType[classification.Type]++
		
		if summary.ErrorsByStep[aggErr.Step] == nil {
			summary.ErrorsByStep[aggErr.Step] = make([]error, 0)
		}
		summary.ErrorsByStep[aggErr.Step] = append(summary.ErrorsByStep[aggErr.Step], aggErr.Error)
	}

	return summary
}

// GetPatterns returns detected error patterns
func (ea *ErrorAggregator) GetPatterns() []ErrorPattern {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()

	return ea.detectPatterns()
}

// detectPatterns analyzes errors to detect common patterns
func (ea *ErrorAggregator) detectPatterns() []ErrorPattern {
	patterns := make(map[ErrorType]*ErrorPattern)

	for _, aggErr := range ea.errors {
		classification := ClassifyError(aggErr.Error)
		
		if pattern, exists := patterns[classification.Type]; exists {
			pattern.Count++
			// Add service if not already present
			serviceFound := false
			for _, service := range pattern.Services {
				if service == aggErr.Step {
					serviceFound = true
					break
				}
			}
			if !serviceFound {
				pattern.Services = append(pattern.Services, aggErr.Step)
			}
		} else {
			patterns[classification.Type] = &ErrorPattern{
				Type:     classification.Type,
				Count:    1,
				Services: []string{aggErr.Step},
				Severity: classification.Severity,
			}
		}
	}

	result := make([]ErrorPattern, 0, len(patterns))
	for _, pattern := range patterns {
		result = append(result, *pattern)
	}

	return result
}

// Helper functions

// containsAny checks if the string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// calculateExponentialBackoffDelay calculates delay using exponential backoff
func calculateExponentialBackoffDelay(attempt int, baseDelay time.Duration, multiplier float64) time.Duration {
	if attempt <= 0 {
		return baseDelay
	}
	
	delay := baseDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * multiplier)
	}
	
	return delay
}

// calculateAdaptiveDelay calculates delay using adaptive strategy based on error context
func calculateAdaptiveDelay(ctx RecoveryContext, classification ErrorClassification) time.Duration {
	baseDelay := 100 * time.Millisecond
	
	// Adjust based on error severity
	severityMultiplier := float64(classification.Severity)
	
	// Adjust based on attempt count
	attemptMultiplier := float64(ctx.Attempt)
	
	// Calculate adaptive delay
	delay := time.Duration(float64(baseDelay) * severityMultiplier * attemptMultiplier)
	
	// Cap the delay
	maxDelay := 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return delay
}