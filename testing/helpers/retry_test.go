package helpers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// RED: Test that DoWithRetry executes operation successfully on first try
func TestDoWithRetry_ShouldExecuteSuccessfullyOnFirstTry(t *testing.T) {
	// Arrange
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}
	
	// Act
	err := DoWithRetry("test-operation", 3, 100*time.Millisecond, operation)
	
	// Assert
	assert.NoError(t, err, "DoWithRetry should succeed on first try")
	assert.Equal(t, 1, callCount, "Operation should be called exactly once")
}

// RED: Test that DoWithRetry retries on failure and eventually succeeds
func TestDoWithRetry_ShouldRetryOnFailureAndEventuallySucceed(t *testing.T) {
	// Arrange
	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary failure")
		}
		return nil
	}
	
	// Act
	err := DoWithRetry("test-operation", 3, 10*time.Millisecond, operation)
	
	// Assert
	assert.NoError(t, err, "DoWithRetry should eventually succeed")
	assert.Equal(t, 3, callCount, "Operation should be called 3 times")
}

// RED: Test that DoWithRetry fails after max attempts
func TestDoWithRetry_ShouldFailAfterMaxAttempts(t *testing.T) {
	// Arrange
	callCount := 0
	operation := func() error {
		callCount++
		return errors.New("persistent failure")
	}
	
	// Act
	err := DoWithRetry("test-operation", 3, 10*time.Millisecond, operation)
	
	// Assert
	assert.Error(t, err, "DoWithRetry should fail after max attempts")
	assert.Contains(t, err.Error(), "after 3 retries", "Error should mention number of retries")
	assert.Equal(t, 4, callCount, "Operation should be called exactly max attempts times (initial + retries)")
}

// RED: Test that DoWithRetryE executes operation successfully on first try
func TestDoWithRetryE_ShouldExecuteSuccessfullyOnFirstTry(t *testing.T) {
	// Arrange
	expectedResult := "success"
	callCount := 0
	operation := func() (string, error) {
		callCount++
		return expectedResult, nil
	}
	
	// Act
	result, err := DoWithRetryE("test-operation", 3, 100*time.Millisecond, operation)
	
	// Assert
	assert.NoError(t, err, "DoWithRetryE should succeed on first try")
	assert.Equal(t, expectedResult, result, "DoWithRetryE should return expected result")
	assert.Equal(t, 1, callCount, "Operation should be called exactly once")
}

// RED: Test that DoWithRetryE retries on failure and eventually succeeds
func TestDoWithRetryE_ShouldRetryOnFailureAndEventuallySucceed(t *testing.T) {
	// Arrange
	expectedResult := "success"
	callCount := 0
	operation := func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", errors.New("temporary failure")
		}
		return expectedResult, nil
	}
	
	// Act
	result, err := DoWithRetryE("test-operation", 3, 10*time.Millisecond, operation)
	
	// Assert
	assert.NoError(t, err, "DoWithRetryE should eventually succeed")
	assert.Equal(t, expectedResult, result, "DoWithRetryE should return expected result")
	assert.Equal(t, 3, callCount, "Operation should be called 3 times")
}

// RED: Test that DoWithRetryE fails after max attempts
func TestDoWithRetryE_ShouldFailAfterMaxAttempts(t *testing.T) {
	// Arrange
	callCount := 0
	operation := func() (string, error) {
		callCount++
		return "", errors.New("persistent failure")
	}
	
	// Act
	result, err := DoWithRetryE("test-operation", 3, 10*time.Millisecond, operation)
	
	// Assert
	assert.Error(t, err, "DoWithRetryE should fail after max attempts")
	assert.Empty(t, result, "DoWithRetryE should return zero value on failure")
	assert.Contains(t, err.Error(), "after 3 retries", "Error should mention number of retries")
	assert.Equal(t, 4, callCount, "Operation should be called exactly max attempts times (initial + retries)")
}

// RED: Test that DoWithRetryWithContext respects context cancellation
func TestDoWithRetryWithContext_ShouldRespectContextCancellation(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	callCount := 0
	operation := func() error {
		callCount++
		// First call should detect context cancellation during operation
		time.Sleep(100 * time.Millisecond) // Longer than context timeout
		return errors.New("should not reach here due to context cancellation")
	}
	
	// Act
	err := DoWithRetryWithContext(ctx, "test-operation", 5, 10*time.Millisecond, operation)
	
	// Assert
	assert.Error(t, err, "DoWithRetryWithContext should fail when context is cancelled")
	assert.Contains(t, err.Error(), "context", "Error should mention context cancellation")
}

// RED: Test that DoWithRetryWithContext succeeds within context deadline
func TestDoWithRetryWithContext_ShouldSucceedWithinContextDeadline(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary failure")
		}
		return nil
	}
	
	// Act
	err := DoWithRetryWithContext(ctx, "test-operation", 3, 10*time.Millisecond, operation)
	
	// Assert
	assert.NoError(t, err, "DoWithRetryWithContext should succeed within context deadline")
	assert.Equal(t, 2, callCount, "Operation should be called 2 times")
}

// RED: Test that DoWithRetryEWithContext respects context cancellation
func TestDoWithRetryEWithContext_ShouldRespectContextCancellation(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	callCount := 0
	operation := func() (string, error) {
		callCount++
		time.Sleep(100 * time.Millisecond) // Longer than context timeout
		return "result", nil
	}
	
	// Act
	result, err := DoWithRetryEWithContext(ctx, "test-operation", 5, 10*time.Millisecond, operation)
	
	// Assert
	assert.Error(t, err, "DoWithRetryEWithContext should fail when context is cancelled")
	assert.Empty(t, result, "DoWithRetryEWithContext should return zero value on context cancellation")
	assert.Contains(t, err.Error(), "context", "Error should mention context cancellation")
}

// RED: Test exponential backoff behavior
func TestDoWithRetryExponentialBackoff_ShouldUseExponentialBackoff(t *testing.T) {
	// Arrange
	callCount := 0
	startTime := time.Now()
	delays := make([]time.Duration, 0)
	
	operation := func() error {
		callCount++
		if callCount > 1 {
			delays = append(delays, time.Since(startTime))
		}
		startTime = time.Now()
		
		if callCount < 4 {
			return errors.New("temporary failure")
		}
		return nil
	}
	
	// Act
	err := DoWithRetryExponentialBackoff("test-operation", 4, 20*time.Millisecond, 2.0, operation)
	
	// Assert
	assert.NoError(t, err, "DoWithRetryExponentialBackoff should eventually succeed")
	assert.Equal(t, 4, callCount, "Operation should be called 4 times")
	assert.Len(t, delays, 3, "Should have 3 delay measurements")
	
	// Check that delays increase exponentially (with some tolerance for timing variations)
	if len(delays) >= 2 {
		assert.Greater(t, delays[1], delays[0], "Second delay should be greater than first")
	}
}

// RED: Test that DoWithRetryExponentialBackoff fails after max attempts
func TestDoWithRetryExponentialBackoff_ShouldFailAfterMaxAttempts(t *testing.T) {
	// Arrange
	callCount := 0
	operation := func() error {
		callCount++
		return errors.New("persistent failure")
	}
	
	// Act
	err := DoWithRetryExponentialBackoff("test-operation", 3, 10*time.Millisecond, 2.0, operation)
	
	// Assert
	assert.Error(t, err, "DoWithRetryExponentialBackoff should fail after max attempts")
	assert.Contains(t, err.Error(), "after 3 attempts", "Error should mention number of attempts")
	assert.Equal(t, 3, callCount, "Operation should be called exactly max attempts times")
}