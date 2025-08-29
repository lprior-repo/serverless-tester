package recovery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRecoveryFramework tests the complete recovery framework integration
func TestRecoveryFramework(t *testing.T) {
	t.Run("creates recovery framework with default configuration", func(t *testing.T) {
		framework := NewRecoveryFramework(RecoveryFrameworkConfig{
			Name: "test-framework",
		})

		assert.NotNil(t, framework)
		assert.Equal(t, "test-framework", framework.Name())
	})

	t.Run("executes operation with recovery strategies", func(t *testing.T) {
		config := RecoveryFrameworkConfig{
			Name: "test-framework",
			DefaultRetryConfig: ExponentialBackoffConfig{
				BaseDelay:   10 * time.Millisecond,
				MaxDelay:    100 * time.Millisecond,
				Multiplier:  2.0,
				MaxAttempts: 3,
				Jitter:      false,
			},
			CircuitBreakerConfig: CircuitBreakerConfig{
				Name:             "test-circuit",
				FailureThreshold: 5,
				RecoveryTimeout:  1 * time.Second,
			},
			BulkheadConfig: BulkheadConfig{
				Name:         "test-bulkhead",
				MaxConcurrent: 10,
				QueueSize:    20,
				Timeout:      1 * time.Second,
			},
		}

		framework := NewRecoveryFramework(config)
		ctx := context.Background()

		attemptCount := 0
		result, err := framework.Execute(ctx, "test-operation", func() (string, error) {
			attemptCount++
			if attemptCount < 2 {
				return "", errors.New("temporary failure")
			}
			return "success after retry", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success after retry", result)
		assert.Equal(t, 2, attemptCount)
	})

	t.Run("selects appropriate recovery strategy based on error", func(t *testing.T) {
		config := RecoveryFrameworkConfig{
			Name: "test-framework",
			DefaultRetryConfig: ExponentialBackoffConfig{
				BaseDelay:   10 * time.Millisecond,
				MaxDelay:    100 * time.Millisecond,
				Multiplier:  2.0,
				MaxAttempts: 3,
				Jitter:      false,
			},
			EnableAdaptiveRetry: true,
		}

		framework := NewRecoveryFramework(config)
		ctx := context.Background()

		// Test with throttling error - should use exponential backoff
		attemptCount := 0
		result, err := framework.Execute(ctx, "throttling-test", func() (string, error) {
			attemptCount++
			if attemptCount < 2 {
				return "", errors.New("ThrottlingException: Rate exceeded")
			}
			return "success after throttling", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success after throttling", result)
	})

	t.Run("handles non-retryable errors correctly", func(t *testing.T) {
		config := RecoveryFrameworkConfig{
			Name: "test-framework",
			DefaultRetryConfig: ExponentialBackoffConfig{
				BaseDelay:   10 * time.Millisecond,
				MaxDelay:    100 * time.Millisecond,
				Multiplier:  2.0,
				MaxAttempts: 3,
				Jitter:      false,
			},
		}

		framework := NewRecoveryFramework(config)
		ctx := context.Background()

		attemptCount := 0
		result, err := framework.Execute(ctx, "auth-test", func() (string, error) {
			attemptCount++
			return "", errors.New("UnauthorizedOperation: Access denied")
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, 1, attemptCount) // Should not retry authorization errors
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		config := RecoveryFrameworkConfig{
			Name: "test-framework",
			DefaultRetryConfig: ExponentialBackoffConfig{
				BaseDelay:   100 * time.Millisecond,
				MaxDelay:    1 * time.Second,
				Multiplier:  2.0,
				MaxAttempts: 5,
				Jitter:      false,
			},
		}

		framework := NewRecoveryFramework(config)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		result, err := framework.Execute(ctx, "timeout-test", func() (string, error) {
			return "", errors.New("will be cancelled")
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "context")
	})

	t.Run("tracks recovery metrics", func(t *testing.T) {
		config := RecoveryFrameworkConfig{
			Name: "test-framework",
			DefaultRetryConfig: ExponentialBackoffConfig{
				BaseDelay:   10 * time.Millisecond,
				MaxDelay:    100 * time.Millisecond,
				Multiplier:  2.0,
				MaxAttempts: 3,
				Jitter:      false,
			},
		}

		framework := NewRecoveryFramework(config)
		ctx := context.Background()

		// Execute some operations
		framework.Execute(ctx, "success-test", func() (string, error) {
			return "success", nil
		})

		framework.Execute(ctx, "failure-test", func() (string, error) {
			return "", errors.New("persistent failure")
		})

		metrics := framework.GetMetrics()

		assert.Equal(t, "test-framework", metrics.Name)
		assert.Equal(t, int64(2), metrics.TotalOperations)
		assert.Equal(t, int64(1), metrics.TotalSuccesses)
		assert.GreaterOrEqual(t, metrics.TotalFailures, int64(1))
	})
}

// TestRecoveryFrameworkWithServiceBulkheads tests service-specific isolation
func TestRecoveryFrameworkWithServiceBulkheads(t *testing.T) {
	t.Run("uses service-specific bulkheads", func(t *testing.T) {
		config := RecoveryFrameworkConfig{
			Name: "service-framework",
			ServiceBulkheads: map[string]BulkheadConfig{
				"lambda": {
					Name:         "lambda-bulkhead",
					MaxConcurrent: 2,
					QueueSize:    1,
					Timeout:      1 * time.Second,
				},
				"dynamodb": {
					Name:         "dynamo-bulkhead",
					MaxConcurrent: 1,
					QueueSize:    1,
					Timeout:      1 * time.Second,
				},
			},
		}

		framework := NewRecoveryFramework(config)
		ctx := context.Background()

		// Test lambda service
		lambdaResult, err := framework.ExecuteWithService(ctx, "lambda", "lambda-op", func() (string, error) {
			return "lambda-success", nil
		})

		// Test DynamoDB service
		dynamoResult, err2 := framework.ExecuteWithService(ctx, "dynamodb", "dynamo-op", func() (string, error) {
			return "dynamo-success", nil
		})

		assert.NoError(t, err)
		assert.NoError(t, err2)
		assert.Equal(t, "lambda-success", lambdaResult)
		assert.Equal(t, "dynamo-success", dynamoResult)
	})

	t.Run("isolates failures between services", func(t *testing.T) {
		config := RecoveryFrameworkConfig{
			Name: "isolation-framework",
			ServiceBulkheads: map[string]BulkheadConfig{
				"service1": {
					Name:         "service1-bulkhead",
					MaxConcurrent: 1,
					QueueSize:    0,
					Timeout:      100 * time.Millisecond,
				},
				"service2": {
					Name:         "service2-bulkhead",
					MaxConcurrent: 1,
					QueueSize:    0,
					Timeout:      100 * time.Millisecond,
				},
			},
		}

		framework := NewRecoveryFramework(config)
		ctx := context.Background()

		// Fill service1 with long-running operation
		go func() {
			framework.ExecuteWithService(ctx, "service1", "slow-op", func() (string, error) {
				time.Sleep(200 * time.Millisecond)
				return "slow-result", nil
			})
		}()

		time.Sleep(10 * time.Millisecond) // Give time for operation to start

		// Service1 should be busy
		service1Result, err1 := framework.ExecuteWithService(ctx, "service1", "quick-op", func() (string, error) {
			return "should-fail", nil
		})

		// Service2 should still work
		service2Result, err2 := framework.ExecuteWithService(ctx, "service2", "quick-op", func() (string, error) {
			return "should-work", nil
		})

		assert.Error(t, err1)
		assert.Empty(t, service1Result)
		assert.NoError(t, err2)
		assert.Equal(t, "should-work", service2Result)
	})
}

// TestRecoveryFrameworkIntegration tests full integration with all patterns
func TestRecoveryFrameworkIntegration(t *testing.T) {
	t.Run("combines circuit breaker, exponential backoff, and bulkhead", func(t *testing.T) {
		config := RecoveryFrameworkConfig{
			Name: "integration-framework",
			DefaultRetryConfig: ExponentialBackoffConfig{
				BaseDelay:   10 * time.Millisecond,
				MaxDelay:    100 * time.Millisecond,
				Multiplier:  2.0,
				MaxAttempts: 3,
				Jitter:      false,
			},
			CircuitBreakerConfig: CircuitBreakerConfig{
				Name:             "integration-circuit",
				FailureThreshold: 3,
				RecoveryTimeout:  500 * time.Millisecond,
			},
			BulkheadConfig: BulkheadConfig{
				Name:         "integration-bulkhead",
				MaxConcurrent: 5,
				QueueSize:    10,
				Timeout:      1 * time.Second,
			},
			EnableCircuitBreaker: true,
			EnableBulkhead:       true,
			EnableAdaptiveRetry:  true,
		}

		framework := NewRecoveryFramework(config)
		ctx := context.Background()

		// Test successful operation
		result, err := framework.Execute(ctx, "integration-test", func() (string, error) {
			return "integration-success", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "integration-success", result)

		// Verify all components are working
		metrics := framework.GetMetrics()
		assert.NotNil(t, metrics.CircuitBreakerMetrics)
		assert.NotNil(t, metrics.BulkheadMetrics)
		assert.Equal(t, int64(1), metrics.TotalSuccesses)
	})
}