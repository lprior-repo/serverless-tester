package recovery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCircuitBreaker tests circuit breaker functionality for AWS services
func TestCircuitBreaker(t *testing.T) {
	t.Run("starts in closed state", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:                "test-circuit",
			FailureThreshold:    3,
			RecoveryTimeout:     5 * time.Second,
			SuccessThreshold:    2,
			MaxConcurrentCalls:  10,
		})

		assert.Equal(t, CircuitStateClosed, cb.GetState())
		assert.True(t, cb.CanExecute())
	})

	t.Run("transitions to open state after threshold failures", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "test-circuit",
			FailureThreshold: 3,
			RecoveryTimeout:  5 * time.Second,
		})

		// Record failures up to threshold
		for i := 0; i < 3; i++ {
			cb.RecordFailure(errors.New("service failure"))
		}

		assert.Equal(t, CircuitStateOpen, cb.GetState())
		assert.False(t, cb.CanExecute())
	})

	t.Run("transitions to half-open after recovery timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "test-circuit",
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
		})

		// Force circuit to open
		cb.RecordFailure(errors.New("failure 1"))
		cb.RecordFailure(errors.New("failure 2"))
		assert.Equal(t, CircuitStateOpen, cb.GetState())

		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Circuit should now allow limited requests
		assert.True(t, cb.CanExecute())
		assert.Equal(t, CircuitStateHalfOpen, cb.GetState())
	})

	t.Run("closes after successful executions in half-open state", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "test-circuit",
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
			SuccessThreshold: 2,
		})

		// Force to half-open state
		cb.RecordFailure(errors.New("failure 1"))
		cb.RecordFailure(errors.New("failure 2"))
		time.Sleep(150 * time.Millisecond)
		cb.CanExecute() // Transition to half-open

		// Record successful executions
		cb.RecordSuccess()
		cb.RecordSuccess()

		assert.Equal(t, CircuitStateClosed, cb.GetState())
	})

	t.Run("returns to open state on failure in half-open", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "test-circuit",
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
		})

		// Force to half-open state
		cb.RecordFailure(errors.New("failure 1"))
		cb.RecordFailure(errors.New("failure 2"))
		time.Sleep(150 * time.Millisecond)
		cb.CanExecute() // Transition to half-open

		// Record failure in half-open state
		cb.RecordFailure(errors.New("half-open failure"))

		assert.Equal(t, CircuitStateOpen, cb.GetState())
		assert.False(t, cb.CanExecute())
	})

	t.Run("enforces max concurrent calls", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:               "test-circuit",
			MaxConcurrentCalls: 2,
		})

		// Acquire max concurrent calls
		cb.AcquireCall()
		cb.AcquireCall()

		// Third call should be rejected
		assert.False(t, cb.AcquireCall())

		// Release one call
		cb.ReleaseCall()

		// Now should be able to acquire again
		assert.True(t, cb.AcquireCall())
	})
}

// TestCircuitBreakerExecution tests circuit breaker with actual operation execution
func TestCircuitBreakerExecution(t *testing.T) {
	t.Run("executes operation when closed", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "test-circuit",
			FailureThreshold: 3,
		})

		ctx := context.Background()
		result, err := ExecuteWithCircuitBreaker(ctx, cb, func() (string, error) {
			return "success", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, CircuitStateClosed, cb.GetState())
	})

	t.Run("rejects execution when open", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "test-circuit",
			FailureThreshold: 1,
		})

		ctx := context.Background()

		// Force circuit open
		cb.RecordFailure(errors.New("initial failure"))

		result, err := ExecuteWithCircuitBreaker(ctx, cb, func() (string, error) {
			return "should not execute", nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker is open")
		assert.Empty(t, result)
	})

	t.Run("handles operation failures correctly", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "test-circuit",
			FailureThreshold: 2,
		})

		ctx := context.Background()
		operationError := errors.New("operation failed")

		result, err := ExecuteWithCircuitBreaker(ctx, cb, func() (string, error) {
			return "", operationError
		})

		assert.Error(t, err)
		assert.Equal(t, operationError, err)
		assert.Empty(t, result)

		// Circuit should still be closed after one failure
		assert.Equal(t, CircuitStateClosed, cb.GetState())
	})
}

// TestCircuitBreakerMetrics tests circuit breaker metrics collection
func TestCircuitBreakerMetrics(t *testing.T) {
	t.Run("tracks failure and success counts", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "test-circuit",
			FailureThreshold: 5,
		})

		// Record some failures and successes
		cb.RecordFailure(errors.New("failure 1"))
		cb.RecordFailure(errors.New("failure 2"))
		cb.RecordSuccess()
		cb.RecordSuccess()
		cb.RecordSuccess()

		metrics := cb.GetMetrics()

		assert.Equal(t, int64(2), metrics.TotalFailures)
		assert.Equal(t, int64(3), metrics.TotalSuccesses)
		assert.Equal(t, CircuitStateClosed, metrics.State)
	})

	t.Run("tracks state transition times", func(t *testing.T) {
		cb := NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "test-circuit",
			FailureThreshold: 2,
		})

		startTime := time.Now()

		// Force state transition
		cb.RecordFailure(errors.New("failure 1"))
		cb.RecordFailure(errors.New("failure 2"))

		metrics := cb.GetMetrics()

		assert.True(t, metrics.LastStateChange.After(startTime))
		assert.Equal(t, CircuitStateOpen, metrics.State)
	})
}