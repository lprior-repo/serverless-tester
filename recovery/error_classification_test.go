package recovery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestErrorClassification tests automatic error classification and handling
func TestErrorClassification(t *testing.T) {
	t.Run("classifies AWS throttling errors", func(t *testing.T) {
		err := errors.New("ThrottlingException: Rate exceeded")
		
		classification := ClassifyError(err)
		
		assert.Equal(t, ErrorTypeThrottling, classification.Type)
		assert.True(t, classification.Retryable)
		assert.Equal(t, RetryStrategyExponentialBackoff, classification.RecommendedStrategy)
	})

	t.Run("classifies AWS service unavailable errors", func(t *testing.T) {
		err := errors.New("ServiceUnavailableException: Service is temporarily unavailable")
		
		classification := ClassifyError(err)
		
		assert.Equal(t, ErrorTypeServiceUnavailable, classification.Type)
		assert.True(t, classification.Retryable)
		assert.Equal(t, RetryStrategyCircuitBreaker, classification.RecommendedStrategy)
	})

	t.Run("classifies network timeout errors", func(t *testing.T) {
		err := context.DeadlineExceeded
		
		classification := ClassifyError(err)
		
		assert.Equal(t, ErrorTypeTimeout, classification.Type)
		assert.True(t, classification.Retryable)
		assert.Equal(t, RetryStrategyAdaptive, classification.RecommendedStrategy)
	})

	t.Run("classifies authorization errors as non-retryable", func(t *testing.T) {
		err := errors.New("UnauthorizedOperation: You do not have permission")
		
		classification := ClassifyError(err)
		
		assert.Equal(t, ErrorTypeAuthorization, classification.Type)
		assert.False(t, classification.Retryable)
		assert.Equal(t, RetryStrategyNone, classification.RecommendedStrategy)
	})

	t.Run("classifies validation errors as non-retryable", func(t *testing.T) {
		err := errors.New("ValidationException: Invalid parameter value")
		
		classification := ClassifyError(err)
		
		assert.Equal(t, ErrorTypeValidation, classification.Type)
		assert.False(t, classification.Retryable)
		assert.Equal(t, RetryStrategyNone, classification.RecommendedStrategy)
	})
}

// TestErrorAggregation tests error aggregation and correlation
func TestErrorAggregation(t *testing.T) {
	t.Run("aggregates multiple errors from same source", func(t *testing.T) {
		aggregator := NewErrorAggregator("test-operation")
		
		err1 := errors.New("first error")
		err2 := errors.New("second error")
		err3 := errors.New("third error")
		
		aggregator.AddError("step1", err1)
		aggregator.AddError("step2", err2)
		aggregator.AddError("step3", err3)
		
		summary := aggregator.GetSummary()
		
		assert.Equal(t, 3, summary.TotalErrors)
		assert.Len(t, summary.ErrorsByStep, 3)
		assert.Contains(t, summary.ErrorsByStep, "step1")
		assert.Contains(t, summary.ErrorsByStep, "step2")
		assert.Contains(t, summary.ErrorsByStep, "step3")
	})

	t.Run("correlates errors with common patterns", func(t *testing.T) {
		aggregator := NewErrorAggregator("test-operation")
		
		// Multiple throttling errors
		throttleErr1 := errors.New("ThrottlingException: Rate exceeded for service A")
		throttleErr2 := errors.New("ThrottlingException: Rate exceeded for service B")
		timeoutErr := context.DeadlineExceeded
		
		aggregator.AddError("serviceA", throttleErr1)
		aggregator.AddError("serviceB", throttleErr2)
		aggregator.AddError("serviceC", timeoutErr)
		
		patterns := aggregator.GetPatterns()
		
		assert.Len(t, patterns, 2)
		assert.Contains(t, patterns, ErrorPattern{
			Type:     ErrorTypeThrottling,
			Count:    2,
			Services: []string{"serviceA", "serviceB"},
			Severity: 2,
		})
		assert.Contains(t, patterns, ErrorPattern{
			Type:     ErrorTypeTimeout,
			Count:    1,
			Services: []string{"serviceC"},
			Severity: 3,
		})
	})
}

// TestRecoveryStrategySelection tests recovery strategy selection based on error context
func TestRecoveryStrategySelection(t *testing.T) {
	t.Run("selects exponential backoff for throttling errors", func(t *testing.T) {
		ctx := context.Background()
		err := errors.New("ThrottlingException: Rate exceeded")
		context := RecoveryContext{
			OperationName: "test-operation",
			Attempt:      1,
			MaxAttempts:  5,
		}
		
		strategy := SelectRecoveryStrategy(ctx, err, context)
		
		assert.Equal(t, RetryStrategyExponentialBackoff, strategy.Type)
		assert.True(t, strategy.ShouldRetry)
		assert.Greater(t, strategy.Delay, time.Duration(0))
	})

	t.Run("selects circuit breaker for service unavailable", func(t *testing.T) {
		ctx := context.Background()
		err := errors.New("ServiceUnavailableException: Service temporarily unavailable")
		context := RecoveryContext{
			OperationName: "test-operation",
			Attempt:      3,
			MaxAttempts:  5,
		}
		
		strategy := SelectRecoveryStrategy(ctx, err, context)
		
		assert.Equal(t, RetryStrategyCircuitBreaker, strategy.Type)
		assert.True(t, strategy.ShouldRetry)
	})

	t.Run("selects no retry for authorization errors", func(t *testing.T) {
		ctx := context.Background()
		err := errors.New("UnauthorizedOperation: You do not have permission")
		context := RecoveryContext{
			OperationName: "test-operation",
			Attempt:      1,
			MaxAttempts:  5,
		}
		
		strategy := SelectRecoveryStrategy(ctx, err, context)
		
		assert.Equal(t, RetryStrategyNone, strategy.Type)
		assert.False(t, strategy.ShouldRetry)
	})

	t.Run("escalates strategy after multiple failures", func(t *testing.T) {
		ctx := context.Background()
		err := errors.New("InternalServerError: Service error")
		context := RecoveryContext{
			OperationName: "test-operation",
			Attempt:      4,
			MaxAttempts:  5,
		}
		
		strategy := SelectRecoveryStrategy(ctx, err, context)
		
		// Should escalate to more aggressive strategy near max attempts
		assert.True(t, strategy.Type == RetryStrategyCircuitBreaker || strategy.Type == RetryStrategyAdaptive)
	})
}