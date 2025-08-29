package recovery

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitStateClosed CircuitState = iota
	CircuitStateOpen
	CircuitStateHalfOpen
)

// String returns the string representation of CircuitState
func (cs CircuitState) String() string {
	switch cs {
	case CircuitStateClosed:
		return "closed"
	case CircuitStateOpen:
		return "open"
	case CircuitStateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig defines the configuration for a circuit breaker
type CircuitBreakerConfig struct {
	Name                string
	FailureThreshold    int
	RecoveryTimeout     time.Duration
	SuccessThreshold    int
	MaxConcurrentCalls  int
	FailureRateWindow   time.Duration
	MinRequestThreshold int
}

// CircuitBreaker implements the circuit breaker pattern for AWS services
type CircuitBreaker struct {
	config            CircuitBreakerConfig
	state             int32 // atomic access to CircuitState
	failureCount      int64
	successCount      int64
	totalRequests     int64
	lastFailureTime   int64 // unix timestamp
	lastStateChange   int64 // unix timestamp
	concurrentCalls   int32
	mutex             sync.RWMutex
	failures          []time.Time // failure window for rate calculation
}

// CircuitBreakerMetrics provides metrics about circuit breaker state
type CircuitBreakerMetrics struct {
	Name            string
	State           CircuitState
	TotalFailures   int64
	TotalSuccesses  int64
	TotalRequests   int64
	ConcurrentCalls int32
	FailureRate     float64
	LastStateChange time.Time
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	// Set default values if not specified
	if config.RecoveryTimeout == 0 {
		config.RecoveryTimeout = 60 * time.Second
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 1
	}
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.MaxConcurrentCalls == 0 {
		config.MaxConcurrentCalls = 100
	}
	if config.FailureRateWindow == 0 {
		config.FailureRateWindow = 60 * time.Second
	}
	if config.MinRequestThreshold == 0 {
		config.MinRequestThreshold = 10
	}

	cb := &CircuitBreaker{
		config:          config,
		state:           int32(CircuitStateClosed),
		lastStateChange: time.Now().UnixNano(),
		failures:        make([]time.Time, 0),
	}

	slog.Info("Circuit breaker created",
		"name", config.Name,
		"failureThreshold", config.FailureThreshold,
		"recoveryTimeout", config.RecoveryTimeout,
		"successThreshold", config.SuccessThreshold,
	)

	return cb
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	return CircuitState(atomic.LoadInt32(&cb.state))
}

// CanExecute determines if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	state := cb.GetState()

	switch state {
	case CircuitStateClosed:
		return true

	case CircuitStateOpen:
		// Check if recovery timeout has passed
		lastFailure := time.Unix(0, atomic.LoadInt64(&cb.lastFailureTime))
		if time.Since(lastFailure) >= cb.config.RecoveryTimeout {
			// Transition to half-open state
			if atomic.CompareAndSwapInt32(&cb.state, int32(CircuitStateOpen), int32(CircuitStateHalfOpen)) {
				atomic.StoreInt64(&cb.lastStateChange, time.Now().UnixNano())
				slog.Info("Circuit breaker transitioned to half-open",
					"name", cb.config.Name,
					"recoveryTimeout", cb.config.RecoveryTimeout,
				)
			}
			return true
		}
		return false

	case CircuitStateHalfOpen:
		// Allow limited requests in half-open state
		return atomic.LoadInt32(&cb.concurrentCalls) == 0

	default:
		return false
	}
}

// AcquireCall attempts to acquire a call slot for concurrent execution
func (cb *CircuitBreaker) AcquireCall() bool {
	if !cb.CanExecute() {
		return false
	}

	currentCalls := atomic.LoadInt32(&cb.concurrentCalls)
	if currentCalls >= int32(cb.config.MaxConcurrentCalls) {
		return false
	}

	atomic.AddInt32(&cb.concurrentCalls, 1)
	atomic.AddInt64(&cb.totalRequests, 1)
	return true
}

// ReleaseCall releases a call slot after execution
func (cb *CircuitBreaker) ReleaseCall() {
	atomic.AddInt32(&cb.concurrentCalls, -1)
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	atomic.AddInt64(&cb.successCount, 1)

	state := cb.GetState()
	if state == CircuitStateHalfOpen {
		// Check if we have enough successes to close the circuit
		successCount := atomic.LoadInt64(&cb.successCount)
		if successCount >= int64(cb.config.SuccessThreshold) {
			if atomic.CompareAndSwapInt32(&cb.state, int32(CircuitStateHalfOpen), int32(CircuitStateClosed)) {
				atomic.StoreInt64(&cb.lastStateChange, time.Now().UnixNano())
				// Reset counters on successful close
				atomic.StoreInt64(&cb.failureCount, 0)
				atomic.StoreInt64(&cb.successCount, 0)
				
				slog.Info("Circuit breaker closed after successful recovery",
					"name", cb.config.Name,
					"successThreshold", cb.config.SuccessThreshold,
				)
			}
		}
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure(err error) {
	atomic.AddInt64(&cb.failureCount, 1)
	atomic.StoreInt64(&cb.lastFailureTime, time.Now().UnixNano())

	cb.mutex.Lock()
	cb.failures = append(cb.failures, time.Now())
	cb.mutex.Unlock()

	state := cb.GetState()

	slog.Warn("Circuit breaker recorded failure",
		"name", cb.config.Name,
		"error", err.Error(),
		"currentState", state.String(),
		"failureCount", atomic.LoadInt64(&cb.failureCount),
	)

	switch state {
	case CircuitStateClosed:
		failureCount := atomic.LoadInt64(&cb.failureCount)
		if failureCount >= int64(cb.config.FailureThreshold) {
			if atomic.CompareAndSwapInt32(&cb.state, int32(CircuitStateClosed), int32(CircuitStateOpen)) {
				atomic.StoreInt64(&cb.lastStateChange, time.Now().UnixNano())
				slog.Error("Circuit breaker opened due to failure threshold",
					"name", cb.config.Name,
					"failureThreshold", cb.config.FailureThreshold,
					"failureCount", failureCount,
				)
			}
		}

	case CircuitStateHalfOpen:
		// Any failure in half-open state should open the circuit
		if atomic.CompareAndSwapInt32(&cb.state, int32(CircuitStateHalfOpen), int32(CircuitStateOpen)) {
			atomic.StoreInt64(&cb.lastStateChange, time.Now().UnixNano())
			atomic.StoreInt64(&cb.successCount, 0) // Reset success count
			slog.Error("Circuit breaker opened from half-open due to failure",
				"name", cb.config.Name,
			)
		}
	}
}

// GetMetrics returns current metrics for the circuit breaker
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return CircuitBreakerMetrics{
		Name:            cb.config.Name,
		State:           cb.GetState(),
		TotalFailures:   atomic.LoadInt64(&cb.failureCount),
		TotalSuccesses:  atomic.LoadInt64(&cb.successCount),
		TotalRequests:   atomic.LoadInt64(&cb.totalRequests),
		ConcurrentCalls: atomic.LoadInt32(&cb.concurrentCalls),
		FailureRate:     cb.calculateFailureRate(),
		LastStateChange: time.Unix(0, atomic.LoadInt64(&cb.lastStateChange)),
	}
}

// calculateFailureRate calculates the current failure rate within the window
func (cb *CircuitBreaker) calculateFailureRate() float64 {
	now := time.Now()
	windowStart := now.Add(-cb.config.FailureRateWindow)

	// Count failures within the window
	failuresInWindow := 0
	for _, failureTime := range cb.failures {
		if failureTime.After(windowStart) {
			failuresInWindow++
		}
	}

	totalRequests := atomic.LoadInt64(&cb.totalRequests)
	if totalRequests == 0 {
		return 0.0
	}

	return float64(failuresInWindow) / float64(totalRequests)
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func ExecuteWithCircuitBreaker[T any](ctx context.Context, cb *CircuitBreaker, operation func() (T, error)) (T, error) {
	var zeroValue T

	// Check if circuit breaker allows execution
	if !cb.AcquireCall() {
		return zeroValue, errors.New("circuit breaker is open or max concurrent calls exceeded")
	}

	defer cb.ReleaseCall()

	// Execute the operation
	result, err := operation()

	if err != nil {
		cb.RecordFailure(err)
		return zeroValue, err
	}

	cb.RecordSuccess()
	return result, nil
}

// ExecuteWithCircuitBreakerAndTimeout executes a function with circuit breaker protection and timeout
func ExecuteWithCircuitBreakerAndTimeout[T any](ctx context.Context, cb *CircuitBreaker, timeout time.Duration, operation func() (T, error)) (T, error) {
	var zeroValue T

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Check if circuit breaker allows execution
	if !cb.AcquireCall() {
		return zeroValue, errors.New("circuit breaker is open or max concurrent calls exceeded")
	}

	defer cb.ReleaseCall()

	// Execute operation in a goroutine to handle timeout
	type operationResult struct {
		result T
		err    error
	}

	resultChan := make(chan operationResult, 1)
	go func() {
		result, err := operation()
		resultChan <- operationResult{result: result, err: err}
	}()

	// Wait for either completion or timeout
	select {
	case <-timeoutCtx.Done():
		timeoutErr := timeoutCtx.Err()
		cb.RecordFailure(timeoutErr)
		return zeroValue, timeoutErr

	case opResult := <-resultChan:
		if opResult.err != nil {
			cb.RecordFailure(opResult.err)
			return zeroValue, opResult.err
		}

		cb.RecordSuccess()
		return opResult.result, nil
	}
}