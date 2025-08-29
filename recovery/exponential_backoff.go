package recovery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sync"
	"time"
)

// ExponentialBackoffConfig defines the configuration for exponential backoff
type ExponentialBackoffConfig struct {
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	MaxAttempts int
	Jitter      bool
	JitterRange float64 // 0.0 to 1.0, defaults to 0.5
}

// AdaptiveBackoffConfig extends exponential backoff with error-type-based adaptation
type AdaptiveBackoffConfig struct {
	BaseConfig           ExponentialBackoffConfig
	ErrorTypeMultipliers map[ErrorType]float64
}

// ExponentialBackoff implements exponential backoff with optional jitter
type ExponentialBackoff struct {
	config       ExponentialBackoffConfig
	currentDelay time.Duration
	mutex        sync.Mutex
}

// AdaptiveBackoff implements adaptive exponential backoff based on error types
type AdaptiveBackoff struct {
	config AdaptiveBackoffConfig
	base   *ExponentialBackoff
	mutex  sync.Mutex
}

// NewExponentialBackoff creates a new exponential backoff calculator
func NewExponentialBackoff(config ExponentialBackoffConfig) *ExponentialBackoff {
	// Set default values if not specified
	if config.BaseDelay == 0 {
		config.BaseDelay = 100 * time.Millisecond
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 30 * time.Second
	}
	if config.Multiplier == 0 {
		config.Multiplier = 2.0
	}
	if config.MaxAttempts == 0 {
		config.MaxAttempts = 5
	}
	if config.JitterRange == 0 {
		config.JitterRange = 0.5
	}

	return &ExponentialBackoff{
		config:       config,
		currentDelay: config.BaseDelay,
	}
}

// NextDelay calculates the delay for the given attempt number
func (eb *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if attempt <= 0 {
		return eb.config.BaseDelay
	}

	// Calculate exponential delay
	delay := eb.config.BaseDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * eb.config.Multiplier)
	}

	// Cap at max delay
	if delay > eb.config.MaxDelay {
		delay = eb.config.MaxDelay
	}

	// Apply jitter if enabled
	if eb.config.Jitter {
		delay = eb.applyJitter(delay)
	}

	eb.currentDelay = delay
	return delay
}

// Reset resets the backoff calculator to initial state
func (eb *ExponentialBackoff) Reset() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	eb.currentDelay = eb.config.BaseDelay
}

// applyJitter applies jitter to the delay to avoid thundering herd problems
func (eb *ExponentialBackoff) applyJitter(delay time.Duration) time.Duration {
	if eb.config.JitterRange <= 0 {
		return delay
	}

	// Use full jitter: delay = random(0, delay)
	jitterAmount := rand.Float64() * eb.config.JitterRange
	minDelay := float64(delay) * (1.0 - jitterAmount)
	maxDelay := float64(delay) * (1.0 + jitterAmount)
	
	jitteredDelay := minDelay + rand.Float64()*(maxDelay-minDelay)
	return time.Duration(math.Max(jitteredDelay, float64(eb.config.BaseDelay)))
}

// NewAdaptiveBackoff creates a new adaptive backoff calculator
func NewAdaptiveBackoff(config AdaptiveBackoffConfig) *AdaptiveBackoff {
	// Set default error type multipliers if not provided
	if config.ErrorTypeMultipliers == nil {
		config.ErrorTypeMultipliers = map[ErrorType]float64{
			ErrorTypeThrottling:         3.0,
			ErrorTypeServiceUnavailable: 2.5,
			ErrorTypeTimeout:            1.5,
			ErrorTypeNetwork:            2.0,
			ErrorTypeInternal:           1.0,
		}
	}

	return &AdaptiveBackoff{
		config: config,
		base:   NewExponentialBackoff(config.BaseConfig),
	}
}

// NextDelayForError calculates delay adapted for the specific error type
func (ab *AdaptiveBackoff) NextDelayForError(attempt int, err error) time.Duration {
	ab.mutex.Lock()
	defer ab.mutex.Unlock()

	baseDelay := ab.base.NextDelay(attempt)
	
	if err == nil {
		return baseDelay
	}

	// Classify the error and apply multiplier
	classification := ClassifyError(err)
	multiplier := ab.config.ErrorTypeMultipliers[classification.Type]
	if multiplier == 0 {
		multiplier = 1.0 // Default multiplier
	}

	adaptedDelay := time.Duration(float64(baseDelay) * multiplier)
	
	// Ensure we don't exceed max delay
	if adaptedDelay > ab.config.BaseConfig.MaxDelay {
		adaptedDelay = ab.config.BaseConfig.MaxDelay
	}

	slog.Debug("Adaptive backoff calculated delay",
		"attempt", attempt,
		"errorType", classification.Type.String(),
		"baseDelay", baseDelay,
		"multiplier", multiplier,
		"adaptedDelay", adaptedDelay,
	)

	return adaptedDelay
}

// Reset resets the adaptive backoff calculator
func (ab *AdaptiveBackoff) Reset() {
	ab.mutex.Lock()
	defer ab.mutex.Unlock()

	ab.base.Reset()
}

// ExecuteWithExponentialBackoff executes a function with exponential backoff retry
func ExecuteWithExponentialBackoff[T any](ctx context.Context, config ExponentialBackoffConfig, operation func() (T, error)) (T, error) {
	var zeroValue T
	var lastErr error

	backoff := NewExponentialBackoff(config)

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		slog.Info("Attempting operation with exponential backoff",
			"attempt", attempt,
			"maxAttempts", config.MaxAttempts,
		)

		// Check context cancellation before attempt
		select {
		case <-ctx.Done():
			return zeroValue, ctx.Err()
		default:
		}

		// Execute operation
		result, err := operation()
		if err == nil {
			slog.Info("Operation succeeded",
				"attempt", attempt,
			)
			return result, nil
		}

		lastErr = err
		slog.Warn("Operation failed, will retry",
			"attempt", attempt,
			"error", err.Error(),
		)

		// Don't sleep after the last attempt
		if attempt < config.MaxAttempts {
			delay := backoff.NextDelay(attempt + 1)
			
			slog.Info("Waiting before retry",
				"delay", delay,
				"nextAttempt", attempt+1,
			)

			// Wait with context cancellation support
			select {
			case <-ctx.Done():
				return zeroValue, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	return zeroValue, errors.New("operation failed after " + 
		fmt.Sprintf("%d", config.MaxAttempts) + " attempts: " + lastErr.Error())
}

// ExecuteWithAdaptiveBackoff executes a function with adaptive exponential backoff
func ExecuteWithAdaptiveBackoff[T any](ctx context.Context, config AdaptiveBackoffConfig, operation func() (T, error)) (T, error) {
	var zeroValue T
	var lastErr error

	backoff := NewAdaptiveBackoff(config)

	for attempt := 1; attempt <= config.BaseConfig.MaxAttempts; attempt++ {
		slog.Info("Attempting operation with adaptive backoff",
			"attempt", attempt,
			"maxAttempts", config.BaseConfig.MaxAttempts,
		)

		// Check context cancellation before attempt
		select {
		case <-ctx.Done():
			return zeroValue, ctx.Err()
		default:
		}

		// Execute operation
		result, err := operation()
		if err == nil {
			slog.Info("Operation succeeded with adaptive backoff",
				"attempt", attempt,
			)
			return result, nil
		}

		lastErr = err
		classification := ClassifyError(err)
		
		slog.Warn("Operation failed, will retry with adaptive delay",
			"attempt", attempt,
			"error", err.Error(),
			"errorType", classification.Type.String(),
		)

		// Don't sleep after the last attempt
		if attempt < config.BaseConfig.MaxAttempts {
			delay := backoff.NextDelayForError(attempt + 1, err)
			
			slog.Info("Waiting before retry with adaptive delay",
				"delay", delay,
				"nextAttempt", attempt+1,
				"errorType", classification.Type.String(),
			)

			// Wait with context cancellation support
			select {
			case <-ctx.Done():
				return zeroValue, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	return zeroValue, errors.New("adaptive operation failed after " + 
		fmt.Sprintf("%d", config.BaseConfig.MaxAttempts) + " attempts: " + lastErr.Error())
}

// ExecuteWithExponentialBackoffAndCircuitBreaker combines exponential backoff with circuit breaker
func ExecuteWithExponentialBackoffAndCircuitBreaker[T any](ctx context.Context, config ExponentialBackoffConfig, cb *CircuitBreaker, operation func() (T, error)) (T, error) {
	return ExecuteWithExponentialBackoff(ctx, config, func() (T, error) {
		return ExecuteWithCircuitBreaker(ctx, cb, operation)
	})
}

// CalculateJitteredDelay calculates a jittered delay for the given base delay
func CalculateJitteredDelay(baseDelay time.Duration, jitterRange float64) time.Duration {
	if jitterRange <= 0 || jitterRange > 1 {
		return baseDelay
	}

	jitterAmount := rand.Float64() * jitterRange
	minDelay := float64(baseDelay) * (1.0 - jitterAmount)
	maxDelay := float64(baseDelay) * (1.0 + jitterAmount)
	
	jitteredDelay := minDelay + rand.Float64()*(maxDelay-minDelay)
	return time.Duration(math.Max(jitteredDelay, 0))
}