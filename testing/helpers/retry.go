package helpers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
)

// TestingT defines the interface for test objects that can be used with retry operations
// This interface is compatible with both testing.T and Terratest's testing interface
type TestingT interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	FailNow()
	Fail()
	Helper()
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Name() string
}

// DoWithRetry executes an operation with retry logic using Terratest's retry module
// This provides compatibility with existing code while using Terratest internally
func DoWithRetry(description string, maxRetries int, sleepBetweenRetries time.Duration, operation func() error) error {
	// Create a no-op test that doesn't fail - just collects errors
	noopTest := &noopTesting{}
	
	// Use Terratest's DoWithRetry function
	retry.DoWithRetry(noopTest, description, maxRetries, sleepBetweenRetries, func() (string, error) {
		err := operation()
		if err != nil {
			return "", err
		}
		return "success", nil
	})
	
	// Return the last error if any occurred
	return noopTest.lastErr
}

// DoWithRetryE executes an operation that returns a value with retry logic using Terratest
// Following Terratest patterns for operations that return values
func DoWithRetryE[T any](description string, maxRetries int, sleepBetweenRetries time.Duration, operation func() (T, error)) (T, error) {
	var result T
	
	// Create a no-op test that doesn't fail - just collects errors
	noopTest := &noopTesting{}
	
	// Use Terratest's DoWithRetry function
	retry.DoWithRetry(noopTest, description, maxRetries, sleepBetweenRetries, func() (string, error) {
		r, err := operation()
		if err != nil {
			return "", err
		}
		result = r
		return "success", nil
	})
	
	if noopTest.lastErr != nil {
		var zeroValue T
		return zeroValue, noopTest.lastErr
	}
	
	return result, nil
}

// DoWithRetryWithTest executes operation with Terratest retry using a real test object
func DoWithRetryWithTest(t TestingT, description string, maxRetries int, sleepBetweenRetries time.Duration, operation func() error) {
	retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
		err := operation()
		if err != nil {
			return "", err
		}
		return "success", nil
	})
}

// DoWithRetryEWithTest executes operation with Terratest retry using a real test object
func DoWithRetryEWithTest[T any](t TestingT, description string, maxRetries int, sleepBetweenRetries time.Duration, operation func() (T, error)) T {
	var result T
	
	retry.DoWithRetry(t, description, maxRetries, sleepBetweenRetries, func() (string, error) {
		r, err := operation()
		if err != nil {
			return "", err
		}
		result = r
		return "success", nil
	})
	
	return result
}

// noopTesting is a testing.T implementation that collects errors without failing
type noopTesting struct {
	lastErr error
}

func (nt *noopTesting) Errorf(format string, args ...interface{}) {
	nt.lastErr = fmt.Errorf(format, args...)
}

func (nt *noopTesting) Error(args ...interface{}) {
	nt.lastErr = fmt.Errorf("%v", args...)
}

func (nt *noopTesting) FailNow() {
	// No-op: don't actually fail, just collect errors
}

func (nt *noopTesting) Fail() {
	// No-op: don't actually fail, just collect errors
}

func (nt *noopTesting) Helper() {
	// No-op: marker method for test helpers
}

func (nt *noopTesting) Fatal(args ...interface{}) {
	nt.lastErr = fmt.Errorf("%v", args...)
}

func (nt *noopTesting) Fatalf(format string, args ...interface{}) {
	nt.lastErr = fmt.Errorf(format, args...)
}

func (nt *noopTesting) Name() string {
	return "noopTest"
}

// DoWithRetryWithContext executes an operation with retry logic and context support
// This maintains context cancellation while leveraging Terratest retry patterns
func DoWithRetryWithContext(ctx context.Context, description string, maxRetries int, sleepBetweenRetries time.Duration, operation func() error) error {
	_, err := DoWithRetryEWithContext(ctx, description, maxRetries, sleepBetweenRetries, func() (interface{}, error) {
		return nil, operation()
	})
	return err
}

// DoWithRetryEWithContext executes an operation that returns a value with retry logic and context support
// This maintains context cancellation while leveraging Terratest retry patterns internally
func DoWithRetryEWithContext[T any](ctx context.Context, description string, maxRetries int, sleepBetweenRetries time.Duration, operation func() (T, error)) (T, error) {
	var zeroValue T
	var lastErr error
	
	// Use Terratest's DoWithRetry function with context awareness
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Check if context is cancelled before starting attempt
		select {
		case <-ctx.Done():
			return zeroValue, fmt.Errorf("%s failed due to context cancellation before attempt %d: %w", description, attempt, ctx.Err())
		default:
		}
		
		slog.Info("Executing operation with retry (context-aware)",
			"description", description,
			"attempt", attempt,
			"maxRetries", maxRetries,
		)
		
		// Run operation in a goroutine to allow context cancellation during execution
		type operationResult struct {
			result T
			err    error
		}
		
		resultChan := make(chan operationResult, 1)
		go func() {
			result, err := operation()
			resultChan <- operationResult{result: result, err: err}
		}()
		
		// Wait for either operation completion or context cancellation
		select {
		case <-ctx.Done():
			return zeroValue, fmt.Errorf("%s failed due to context cancellation during attempt %d: %w", description, attempt, ctx.Err())
		case opResult := <-resultChan:
			if opResult.err == nil {
				slog.Info("Operation succeeded",
					"description", description,
					"attempt", attempt,
				)
				return opResult.result, nil
			}
			
			lastErr = opResult.err
			slog.Warn("Operation failed, will retry",
				"description", description,
				"attempt", attempt,
				"error", opResult.err.Error(),
				"sleepDuration", sleepBetweenRetries,
			)
		}
		
		// Sleep between attempts if not the last attempt
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return zeroValue, fmt.Errorf("%s failed due to context cancellation after attempt %d: %w", description, attempt, ctx.Err())
			case <-time.After(sleepBetweenRetries):
				// Continue to next attempt
			}
		}
	}
	
	return zeroValue, fmt.Errorf("%s failed after %d attempts: %w", description, maxRetries, lastErr)
}

// DoWithRetryExponentialBackoff executes an operation with exponential backoff retry logic
// This implements exponential backoff while maintaining compatibility with existing patterns
func DoWithRetryExponentialBackoff(description string, maxRetries int, initialDelay time.Duration, multiplier float64, operation func() error) error {
	_, err := DoWithRetryExponentialBackoffE(description, maxRetries, initialDelay, multiplier, func() (interface{}, error) {
		return nil, operation()
	})
	return err
}

// DoWithRetryExponentialBackoffE executes an operation that returns a value with exponential backoff retry logic
// This implements exponential backoff pattern commonly used in Terratest-based testing
func DoWithRetryExponentialBackoffE[T any](description string, maxRetries int, initialDelay time.Duration, multiplier float64, operation func() (T, error)) (T, error) {
	var zeroValue T
	var lastErr error
	delay := initialDelay
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		slog.Info("Executing operation with exponential backoff retry",
			"description", description,
			"attempt", attempt,
			"maxRetries", maxRetries,
			"currentDelay", delay,
		)
		
		result, err := operation()
		if err == nil {
			slog.Info("Operation succeeded",
				"description", description,
				"attempt", attempt,
			)
			return result, nil
		}
		
		lastErr = err
		slog.Warn("Operation failed, will retry with exponential backoff",
			"description", description,
			"attempt", attempt,
			"error", err.Error(),
			"nextDelay", delay,
		)
		
		if attempt < maxRetries {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * multiplier)
		}
	}
	
	return zeroValue, fmt.Errorf("%s failed after %d attempts: %w", description, maxRetries, lastErr)
}

// DoWithExponentialBackoffE provides a pure Terratest-compatible exponential backoff
// This directly uses Terratest patterns for exponential backoff
func DoWithExponentialBackoffE[T any](t TestingT, description string, maxRetries int, initialDelay time.Duration, multiplier float64, operation func() (T, error)) T {
	var result T
	delay := initialDelay
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		r, err := operation()
		if err == nil {
			result = r
			return result
		}
		
		if attempt == maxRetries {
			t.Errorf("Operation %s failed after %d attempts with exponential backoff: %v", description, maxRetries, err)
			t.FailNow()
		}
		
		time.Sleep(delay)
		delay = time.Duration(float64(delay) * multiplier)
	}
	
	return result
}