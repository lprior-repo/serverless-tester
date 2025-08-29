package recovery

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// RecoveryFrameworkConfig defines the configuration for the recovery framework
type RecoveryFrameworkConfig struct {
	Name                   string
	DefaultRetryConfig     ExponentialBackoffConfig
	CircuitBreakerConfig   CircuitBreakerConfig
	BulkheadConfig         BulkheadConfig
	ServiceBulkheads       map[string]BulkheadConfig
	EnableCircuitBreaker   bool
	EnableBulkhead         bool
	EnableAdaptiveRetry    bool
	ErrorTypeMultipliers   map[ErrorType]float64
	FallbackStrategies     map[ErrorType]func() (interface{}, error)
	HealthCheckInterval    time.Duration
	MetricsRetentionPeriod time.Duration
}

// RecoveryFrameworkMetrics provides comprehensive metrics for the recovery framework
type RecoveryFrameworkMetrics struct {
	Name                   string
	TotalOperations        int64
	TotalSuccesses         int64
	TotalFailures          int64
	TotalRetries           int64
	AverageResponseTime    time.Duration
	SuccessRate            float64
	ErrorsByType           map[ErrorType]int64
	CircuitBreakerMetrics  *CircuitBreakerMetrics
	BulkheadMetrics        *BulkheadMetrics
	ServiceBulkheadMetrics map[string]BulkheadMetrics
	LastUpdated            time.Time
}

// RecoveryFramework provides a comprehensive error recovery system
type RecoveryFramework struct {
	config              RecoveryFrameworkConfig
	circuitBreaker      *CircuitBreaker
	bulkhead           *Bulkhead
	bulkheadManager    *BulkheadManager
	adaptiveBackoff    *AdaptiveBackoff
	errorAggregator    *ErrorAggregator
	mutex              sync.RWMutex
	totalOperations    int64
	totalSuccesses     int64
	totalFailures      int64
	totalRetries       int64
	errorsByType       map[ErrorType]int64
	responseTimeTotal  time.Duration
	startTime          time.Time
}

// NewRecoveryFramework creates a new recovery framework with the given configuration
func NewRecoveryFramework(config RecoveryFrameworkConfig) *RecoveryFramework {
	// Set default values
	if config.Name == "" {
		config.Name = "default-recovery-framework"
	}
	
	if config.DefaultRetryConfig.BaseDelay == 0 {
		config.DefaultRetryConfig = ExponentialBackoffConfig{
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    30 * time.Second,
			Multiplier:  2.0,
			MaxAttempts: 5,
			Jitter:      true,
		}
	}

	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 30 * time.Second
	}

	if config.MetricsRetentionPeriod == 0 {
		config.MetricsRetentionPeriod = 24 * time.Hour
	}

	framework := &RecoveryFramework{
		config:           config,
		errorAggregator:  NewErrorAggregator(config.Name),
		errorsByType:     make(map[ErrorType]int64),
		startTime:        time.Now(),
		bulkheadManager:  NewBulkheadManager(),
	}

	// Initialize circuit breaker if enabled
	if config.EnableCircuitBreaker {
		framework.circuitBreaker = NewCircuitBreaker(config.CircuitBreakerConfig)
		slog.Info("Circuit breaker enabled for recovery framework",
			"frameworkName", config.Name,
			"circuitName", config.CircuitBreakerConfig.Name,
		)
	}

	// Initialize bulkhead if enabled
	if config.EnableBulkhead {
		framework.bulkhead = NewBulkhead(config.BulkheadConfig)
		slog.Info("Bulkhead enabled for recovery framework",
			"frameworkName", config.Name,
			"bulkheadName", config.BulkheadConfig.Name,
		)
	}

	// Initialize service-specific bulkheads
	if config.ServiceBulkheads != nil {
		for serviceType, bulkheadConfig := range config.ServiceBulkheads {
			framework.bulkheadManager.AddBulkhead(serviceType, bulkheadConfig)
		}
		slog.Info("Service bulkheads configured",
			"frameworkName", config.Name,
			"serviceCount", len(config.ServiceBulkheads),
		)
	}

	// Initialize adaptive backoff if enabled
	if config.EnableAdaptiveRetry {
		adaptiveConfig := AdaptiveBackoffConfig{
			BaseConfig:           config.DefaultRetryConfig,
			ErrorTypeMultipliers: config.ErrorTypeMultipliers,
		}
		framework.adaptiveBackoff = NewAdaptiveBackoff(adaptiveConfig)
		slog.Info("Adaptive retry enabled for recovery framework",
			"frameworkName", config.Name,
		)
	}

	slog.Info("Recovery framework initialized",
		"name", config.Name,
		"circuitBreakerEnabled", config.EnableCircuitBreaker,
		"bulkheadEnabled", config.EnableBulkhead,
		"adaptiveRetryEnabled", config.EnableAdaptiveRetry,
	)

	return framework
}

// Name returns the name of the recovery framework
func (rf *RecoveryFramework) Name() string {
	return rf.config.Name
}

// Execute executes an operation with full recovery capabilities
func (rf *RecoveryFramework) Execute(ctx context.Context, operationName string, operation func() (string, error)) (string, error) {
	return rf.executeWithRecovery(ctx, operationName, "", operation)
}

// ExecuteWithService executes an operation using service-specific bulkhead isolation
func (rf *RecoveryFramework) ExecuteWithService(ctx context.Context, serviceType, operationName string, operation func() (string, error)) (string, error) {
	return rf.executeWithRecovery(ctx, operationName, serviceType, operation)
}

// executeWithRecovery implements the core recovery logic
func (rf *RecoveryFramework) executeWithRecovery(ctx context.Context, operationName, serviceType string, operation func() (string, error)) (string, error) {
	startTime := time.Now()
	
	rf.mutex.Lock()
	rf.totalOperations++
	rf.mutex.Unlock()

	slog.Info("Executing operation with recovery framework",
		"framework", rf.config.Name,
		"operation", operationName,
		"serviceType", serviceType,
	)

	// Wrap operation with recovery layers
	wrappedOperation := rf.wrapOperation(ctx, operationName, serviceType, operation)

	// Execute with appropriate retry strategy
	result, err := rf.executeWithRetry(ctx, operationName, wrappedOperation)

	// Update metrics
	duration := time.Since(startTime)
	rf.updateMetrics(operationName, err, duration)

	if err != nil {
		slog.Error("Operation failed after recovery attempts",
			"framework", rf.config.Name,
			"operation", operationName,
			"error", err.Error(),
			"duration", duration,
		)
		return "", err
	}

	slog.Info("Operation succeeded",
		"framework", rf.config.Name,
		"operation", operationName,
		"duration", duration,
	)

	return result, nil
}

// wrapOperation wraps the operation with circuit breaker and bulkhead protection
func (rf *RecoveryFramework) wrapOperation(ctx context.Context, operationName, serviceType string, operation func() (string, error)) func() (string, error) {
	wrappedOp := operation

	// Apply service-specific bulkhead if specified
	if serviceType != "" {
		if bulkhead := rf.bulkheadManager.GetBulkhead(serviceType); bulkhead != nil {
			originalOp := wrappedOp
			wrappedOp = func() (string, error) {
				return ExecuteWithBulkhead(ctx, bulkhead, originalOp)
			}
		}
	} else if rf.config.EnableBulkhead && rf.bulkhead != nil {
		// Apply general bulkhead
		originalOp := wrappedOp
		wrappedOp = func() (string, error) {
			return ExecuteWithBulkhead(ctx, rf.bulkhead, originalOp)
		}
	}

	// Apply circuit breaker
	if rf.config.EnableCircuitBreaker && rf.circuitBreaker != nil {
		originalOp := wrappedOp
		wrappedOp = func() (string, error) {
			return ExecuteWithCircuitBreaker(ctx, rf.circuitBreaker, originalOp)
		}
	}

	return wrappedOp
}

// executeWithRetry executes the operation with the appropriate retry strategy
func (rf *RecoveryFramework) executeWithRetry(ctx context.Context, operationName string, operation func() (string, error)) (string, error) {
	// Execute once first to check if error is retryable
	result, err := operation()
	if err != nil {
		rf.errorAggregator.AddError(operationName, err)
		classification := ClassifyError(err)
		
		// If error is not retryable, return immediately
		if !classification.Retryable {
			slog.Info("Error is not retryable, not attempting retry",
				"framework", rf.config.Name,
				"operation", operationName,
				"errorType", classification.Type.String(),
			)
			return result, err
		}
	} else {
		// Success on first try
		return result, nil
	}

	// Error is retryable, proceed with retry logic
	if rf.config.EnableAdaptiveRetry && rf.adaptiveBackoff != nil {
		// Use adaptive backoff
		adaptiveConfig := AdaptiveBackoffConfig{
			BaseConfig:           rf.config.DefaultRetryConfig,
			ErrorTypeMultipliers: rf.config.ErrorTypeMultipliers,
		}
		
		return ExecuteWithAdaptiveBackoff(ctx, adaptiveConfig, func() (string, error) {
			result, err := operation()
			if err != nil {
				rf.errorAggregator.AddError(operationName, err)
				rf.mutex.Lock()
				rf.totalRetries++
				rf.mutex.Unlock()
			}
			return result, err
		})
	}

	// Use standard exponential backoff
	return ExecuteWithExponentialBackoff(ctx, rf.config.DefaultRetryConfig, func() (string, error) {
		result, err := operation()
		if err != nil {
			rf.errorAggregator.AddError(operationName, err)
			rf.mutex.Lock()
			rf.totalRetries++
			rf.mutex.Unlock()
		}
		return result, err
	})
}

// updateMetrics updates the framework metrics
func (rf *RecoveryFramework) updateMetrics(operationName string, err error, duration time.Duration) {
	rf.mutex.Lock()
	defer rf.mutex.Unlock()

	rf.responseTimeTotal += duration

	if err != nil {
		rf.totalFailures++
		classification := ClassifyError(err)
		rf.errorsByType[classification.Type]++
	} else {
		rf.totalSuccesses++
	}
}

// GetMetrics returns comprehensive metrics for the recovery framework
func (rf *RecoveryFramework) GetMetrics() RecoveryFrameworkMetrics {
	rf.mutex.RLock()
	defer rf.mutex.RUnlock()

	metrics := RecoveryFrameworkMetrics{
		Name:            rf.config.Name,
		TotalOperations: rf.totalOperations,
		TotalSuccesses:  rf.totalSuccesses,
		TotalFailures:   rf.totalFailures,
		TotalRetries:    rf.totalRetries,
		ErrorsByType:    make(map[ErrorType]int64),
		LastUpdated:     time.Now(),
	}

	// Copy error counts
	for errorType, count := range rf.errorsByType {
		metrics.ErrorsByType[errorType] = count
	}

	// Calculate derived metrics
	if rf.totalOperations > 0 {
		metrics.SuccessRate = float64(rf.totalSuccesses) / float64(rf.totalOperations)
		metrics.AverageResponseTime = rf.responseTimeTotal / time.Duration(rf.totalOperations)
	}

	// Add circuit breaker metrics if enabled
	if rf.circuitBreaker != nil {
		cbMetrics := rf.circuitBreaker.GetMetrics()
		metrics.CircuitBreakerMetrics = &cbMetrics
	}

	// Add bulkhead metrics if enabled
	if rf.bulkhead != nil {
		bhMetrics := rf.bulkhead.GetMetrics()
		metrics.BulkheadMetrics = &bhMetrics
	}

	// Add service bulkhead metrics
	if rf.bulkheadManager != nil {
		metrics.ServiceBulkheadMetrics = rf.bulkheadManager.GetAllMetrics()
	}

	return metrics
}

// GetErrorSummary returns aggregated error information
func (rf *RecoveryFramework) GetErrorSummary() ErrorSummary {
	return rf.errorAggregator.GetSummary()
}

// Reset resets the recovery framework state (useful for testing)
func (rf *RecoveryFramework) Reset() {
	rf.mutex.Lock()
	defer rf.mutex.Unlock()

	rf.totalOperations = 0
	rf.totalSuccesses = 0
	rf.totalFailures = 0
	rf.totalRetries = 0
	rf.responseTimeTotal = 0
	rf.errorsByType = make(map[ErrorType]int64)
	rf.startTime = time.Now()

	// Reset components
	if rf.adaptiveBackoff != nil {
		rf.adaptiveBackoff.Reset()
	}

	rf.errorAggregator = NewErrorAggregator(rf.config.Name)

	slog.Info("Recovery framework reset",
		"framework", rf.config.Name,
	)
}

// Shutdown gracefully shuts down the recovery framework
func (rf *RecoveryFramework) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down recovery framework",
		"framework", rf.config.Name,
	)

	// Log final metrics
	finalMetrics := rf.GetMetrics()
	slog.Info("Final recovery framework metrics",
		"framework", rf.config.Name,
		"totalOperations", finalMetrics.TotalOperations,
		"successRate", finalMetrics.SuccessRate,
		"totalRetries", finalMetrics.TotalRetries,
	)

	return nil
}

// DefaultRecoveryFramework creates a recovery framework with sensible defaults for AWS services
func DefaultRecoveryFramework(name string) *RecoveryFramework {
	return NewRecoveryFramework(RecoveryFrameworkConfig{
		Name: name,
		DefaultRetryConfig: ExponentialBackoffConfig{
			BaseDelay:   100 * time.Millisecond,
			MaxDelay:    30 * time.Second,
			Multiplier:  2.0,
			MaxAttempts: 5,
			Jitter:      true,
		},
		CircuitBreakerConfig: CircuitBreakerConfig{
			Name:             name + "-circuit",
			FailureThreshold: 5,
			RecoveryTimeout:  60 * time.Second,
			SuccessThreshold: 3,
		},
		BulkheadConfig: BulkheadConfig{
			Name:         name + "-bulkhead",
			MaxConcurrent: 100,
			QueueSize:    200,
			Timeout:      30 * time.Second,
		},
		ServiceBulkheads: map[string]BulkheadConfig{
			"lambda": {
				Name:         name + "-lambda-bulkhead",
				MaxConcurrent: 50,
				QueueSize:    100,
				Timeout:      30 * time.Second,
			},
			"dynamodb": {
				Name:         name + "-dynamodb-bulkhead",
				MaxConcurrent: 25,
				QueueSize:    50,
				Timeout:      10 * time.Second,
			},
			"s3": {
				Name:         name + "-s3-bulkhead",
				MaxConcurrent: 20,
				QueueSize:    40,
				Timeout:      60 * time.Second,
			},
			"eventbridge": {
				Name:         name + "-eventbridge-bulkhead",
				MaxConcurrent: 30,
				QueueSize:    60,
				Timeout:      15 * time.Second,
			},
		},
		EnableCircuitBreaker: true,
		EnableBulkhead:       true,
		EnableAdaptiveRetry:  true,
		ErrorTypeMultipliers: map[ErrorType]float64{
			ErrorTypeThrottling:         3.0,
			ErrorTypeServiceUnavailable: 2.5,
			ErrorTypeTimeout:            1.5,
			ErrorTypeNetwork:            2.0,
			ErrorTypeInternal:           1.2,
		},
	})
}