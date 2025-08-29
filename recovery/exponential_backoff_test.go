package recovery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestExponentialBackoff tests exponential backoff retry pattern
func TestExponentialBackoff(t *testing.T) {
	t.Run("calculates exponential backoff delays correctly", func(t *testing.T) {
		config := ExponentialBackoffConfig{
			BaseDelay:    100 * time.Millisecond,
			MaxDelay:     10 * time.Second,
			Multiplier:   2.0,
			MaxAttempts:  5,
			Jitter:       false, // Disable jitter for predictable testing
		}

		backoff := NewExponentialBackoff(config)

		delays := []time.Duration{
			100 * time.Millisecond, // attempt 1
			200 * time.Millisecond, // attempt 2
			400 * time.Millisecond, // attempt 3
			800 * time.Millisecond, // attempt 4
			1600 * time.Millisecond, // attempt 5
		}

		for i, expectedDelay := range delays {
			delay := backoff.NextDelay(i + 1)
			assert.Equal(t, expectedDelay, delay, "Incorrect delay for attempt %d", i+1)
		}
	})

	t.Run("caps delay at maximum", func(t *testing.T) {
		config := ExponentialBackoffConfig{
			BaseDelay:   1 * time.Second,
			MaxDelay:    5 * time.Second,
			Multiplier:  2.0,
			MaxAttempts: 10,
			Jitter:      false,
		}

		backoff := NewExponentialBackoff(config)

		// After a few attempts, delay should be capped
		delay := backoff.NextDelay(5) // Should be 16s without cap, but capped at 5s
		assert.Equal(t, 5*time.Second, delay)
	})

	t.Run("applies jitter when enabled", func(t *testing.T) {
		config := ExponentialBackoffConfig{
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    10 * time.Second,
			Multiplier:  2.0,
			MaxAttempts: 3,
			Jitter:      true,
		}

		backoff := NewExponentialBackoff(config)

		// Run multiple times to ensure jitter varies
		delays := make([]time.Duration, 5)
		for i := 0; i < 5; i++ {
			delays[i] = backoff.NextDelay(2) // Second attempt
		}

		// Check that at least some delays are different (due to jitter)
		allSame := true
		for i := 1; i < len(delays); i++ {
			if delays[i] != delays[0] {
				allSame = false
				break
			}
		}
		assert.False(t, allSame, "Expected jitter to create variation in delays")

		// All delays should be within reasonable bounds
		expectedBase := 200 * time.Millisecond // Base delay for attempt 2
		for _, delay := range delays {
			assert.True(t, delay >= expectedBase/2, "Delay too small with jitter")
			assert.True(t, delay <= expectedBase*3/2, "Delay too large with jitter")
		}
	})

	t.Run("resets after successful operation", func(t *testing.T) {
		config := ExponentialBackoffConfig{
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    10 * time.Second,
			Multiplier:  2.0,
			MaxAttempts: 5,
			Jitter:      false,
		}

		backoff := NewExponentialBackoff(config)

		// Simulate multiple failed attempts
		delay1 := backoff.NextDelay(3) // 400ms
		assert.Equal(t, 400*time.Millisecond, delay1)

		// Reset after success
		backoff.Reset()

		// Next delay should be back to base delay
		delay2 := backoff.NextDelay(1) // 100ms
		assert.Equal(t, 100*time.Millisecond, delay2)
	})
}

// TestExponentialBackoffWithRetry tests exponential backoff with actual retry execution
func TestExponentialBackoffWithRetry(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		config := ExponentialBackoffConfig{
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    1 * time.Second,
			Multiplier:  2.0,
			MaxAttempts: 3,
			Jitter:      false,
		}

		ctx := context.Background()
		result, err := ExecuteWithExponentialBackoff(ctx, config, func() (string, error) {
			return "success", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("succeeds after retries", func(t *testing.T) {
		config := ExponentialBackoffConfig{
			BaseDelay:   10 * time.Millisecond, // Short delay for testing
			MaxDelay:    100 * time.Millisecond,
			Multiplier:  2.0,
			MaxAttempts: 3,
			Jitter:      false,
		}

		attemptCount := 0
		ctx := context.Background()

		result, err := ExecuteWithExponentialBackoff(ctx, config, func() (string, error) {
			attemptCount++
			if attemptCount < 3 {
				return "", errors.New("temporary failure")
			}
			return "success on third attempt", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success on third attempt", result)
		assert.Equal(t, 3, attemptCount)
	})

	t.Run("fails after max attempts", func(t *testing.T) {
		config := ExponentialBackoffConfig{
			BaseDelay:   10 * time.Millisecond,
			MaxDelay:    100 * time.Millisecond,
			Multiplier:  2.0,
			MaxAttempts: 2,
			Jitter:      false,
		}

		attemptCount := 0
		ctx := context.Background()

		result, err := ExecuteWithExponentialBackoff(ctx, config, func() (string, error) {
			attemptCount++
			return "", errors.New("persistent failure")
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, 2, attemptCount)
		assert.Contains(t, err.Error(), "after 2 attempts")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		config := ExponentialBackoffConfig{
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    1 * time.Second,
			Multiplier:  2.0,
			MaxAttempts: 5,
			Jitter:      false,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		attemptCount := 0
		result, err := ExecuteWithExponentialBackoff(ctx, config, func() (string, error) {
			attemptCount++
			return "", errors.New("will be cancelled")
		})

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "context")
		// Should have made at least one attempt before cancellation
		assert.GreaterOrEqual(t, attemptCount, 1)
	})
}

// TestAdaptiveExponentialBackoff tests adaptive backoff based on error types
func TestAdaptiveExponentialBackoff(t *testing.T) {
	t.Run("adapts delay based on error classification", func(t *testing.T) {
		config := AdaptiveBackoffConfig{
			BaseConfig: ExponentialBackoffConfig{
				BaseDelay:   100 * time.Millisecond,
				MaxDelay:    10 * time.Second,
				Multiplier:  2.0,
				MaxAttempts: 5,
				Jitter:      false,
			},
			ErrorTypeMultipliers: map[ErrorType]float64{
				ErrorTypeThrottling:         3.0, // Longer delay for throttling
				ErrorTypeServiceUnavailable: 2.0, // Moderate delay
				ErrorTypeTimeout:            1.5, // Slight increase
			},
		}

		backoff := NewAdaptiveBackoff(config)

		// Test throttling error - should have longer delay
		throttleErr := errors.New("ThrottlingException: Rate exceeded")
		delay1 := backoff.NextDelayForError(2, throttleErr)
		
		// Test timeout error - should have shorter delay
		timeoutErr := context.DeadlineExceeded
		delay2 := backoff.NextDelayForError(2, timeoutErr)
		
		// Test unknown error - should use base delay
		unknownErr := errors.New("unknown error")
		delay3 := backoff.NextDelayForError(2, unknownErr)

		// Throttling should have longest delay
		assert.True(t, delay1 > delay2, "Throttling delay should be longer than timeout delay")
		assert.True(t, delay1 > delay3, "Throttling delay should be longer than unknown delay")
		assert.True(t, delay2 > delay3, "Timeout delay should be longer than unknown delay")
	})

	t.Run("executes with adaptive backoff", func(t *testing.T) {
		config := AdaptiveBackoffConfig{
			BaseConfig: ExponentialBackoffConfig{
				BaseDelay:   10 * time.Millisecond,
				MaxDelay:    100 * time.Millisecond,
				Multiplier:  2.0,
				MaxAttempts: 3,
				Jitter:      false,
			},
			ErrorTypeMultipliers: map[ErrorType]float64{
				ErrorTypeThrottling: 2.0,
			},
		}

		attemptCount := 0
		ctx := context.Background()

		result, err := ExecuteWithAdaptiveBackoff(ctx, config, func() (string, error) {
			attemptCount++
			if attemptCount < 3 {
				// Return throttling error for first two attempts
				return "", errors.New("ThrottlingException: Rate exceeded")
			}
			return "success", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 3, attemptCount)
	})
}