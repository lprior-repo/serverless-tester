// Package vasdeference provides functional programming utilities for the core testing infrastructure.
//
// This module provides pure functional programming patterns for:
//   - Configuration management with immutable structures
//   - Error handling with monadic patterns
//   - Retry mechanisms with functional composition
//   - Performance monitoring with immutable metrics
//   - Resilience patterns using functional error handling
//   - Test context management with immutable state
//
// All operations follow strict functional programming principles with no mutations,
// using samber/lo for functional utilities and samber/mo for monadic types.
package vasdeference

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// Default values for functional core operations
const (
	FunctionalCoreDefaultMaxRetries     = 3
	FunctionalCoreDefaultRetryDelay     = 200 * time.Millisecond
	FunctionalCoreDefaultTimeout        = 30 * time.Second
	FunctionalCoreDefaultRegion         = "us-east-1"
	FunctionalCoreMaxNamespaceLength    = 50
	FunctionalCoreMaxRetryAttempts      = 10
)

// FunctionalCoreConfig represents immutable core configuration using functional options
type FunctionalCoreConfig struct {
	region           string
	timeout          time.Duration
	maxRetries       int
	retryDelay       time.Duration
	namespace        mo.Option[string]
	awsProfile       mo.Option[string]
	enableLogging    bool
	enableMetrics    bool
	enableResilience bool
	metadata         map[string]interface{}
}

// FunctionalCoreConfigOption represents a functional option for core configuration
type FunctionalCoreConfigOption func(*FunctionalCoreConfig)

// NewFunctionalCoreConfig creates a new immutable core configuration with functional options
func NewFunctionalCoreConfig(opts ...FunctionalCoreConfigOption) FunctionalCoreConfig {
	config := FunctionalCoreConfig{
		region:           FunctionalCoreDefaultRegion,
		timeout:          FunctionalCoreDefaultTimeout,
		maxRetries:       FunctionalCoreDefaultMaxRetries,
		retryDelay:       FunctionalCoreDefaultRetryDelay,
		namespace:        mo.None[string](),
		awsProfile:       mo.None[string](),
		enableLogging:    true,
		enableMetrics:    false,
		enableResilience: false,
		metadata:         make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(&config)
	}

	return config
}

// Functional options for core configuration
func WithCoreRegion(region string) FunctionalCoreConfigOption {
	return func(c *FunctionalCoreConfig) {
		c.region = region
	}
}

func WithCoreTimeout(timeout time.Duration) FunctionalCoreConfigOption {
	return func(c *FunctionalCoreConfig) {
		c.timeout = timeout
	}
}

func WithCoreRetryConfig(maxRetries int, retryDelay time.Duration) FunctionalCoreConfigOption {
	return func(c *FunctionalCoreConfig) {
		c.maxRetries = maxRetries
		c.retryDelay = retryDelay
	}
}

func WithCoreNamespace(namespace string) FunctionalCoreConfigOption {
	return func(c *FunctionalCoreConfig) {
		c.namespace = mo.Some(namespace)
	}
}

func WithCoreAWSProfile(profile string) FunctionalCoreConfigOption {
	return func(c *FunctionalCoreConfig) {
		c.awsProfile = mo.Some(profile)
	}
}

func WithCoreLogging(enabled bool) FunctionalCoreConfigOption {
	return func(c *FunctionalCoreConfig) {
		c.enableLogging = enabled
	}
}

func WithCoreMetrics(enabled bool) FunctionalCoreConfigOption {
	return func(c *FunctionalCoreConfig) {
		c.enableMetrics = enabled
	}
}

func WithCoreResilience(enabled bool) FunctionalCoreConfigOption {
	return func(c *FunctionalCoreConfig) {
		c.enableResilience = enabled
	}
}

func WithCoreMetadata(key string, value interface{}) FunctionalCoreConfigOption {
	return func(c *FunctionalCoreConfig) {
		c.metadata[key] = value
	}
}

// Accessor methods for FunctionalCoreConfig
func (c FunctionalCoreConfig) GetRegion() string                    { return c.region }
func (c FunctionalCoreConfig) GetTimeout() time.Duration           { return c.timeout }
func (c FunctionalCoreConfig) GetMaxRetries() int                  { return c.maxRetries }
func (c FunctionalCoreConfig) GetRetryDelay() time.Duration        { return c.retryDelay }
func (c FunctionalCoreConfig) GetNamespace() mo.Option[string]     { return c.namespace }
func (c FunctionalCoreConfig) GetAWSProfile() mo.Option[string]    { return c.awsProfile }
func (c FunctionalCoreConfig) IsLoggingEnabled() bool              { return c.enableLogging }
func (c FunctionalCoreConfig) IsMetricsEnabled() bool              { return c.enableMetrics }
func (c FunctionalCoreConfig) IsResilienceEnabled() bool           { return c.enableResilience }
func (c FunctionalCoreConfig) GetMetadata() map[string]interface{} { return c.metadata }

// FunctionalTestContext represents an immutable test context
type FunctionalTestContext struct {
	config        FunctionalCoreConfig
	awsConfig     mo.Option[aws.Config]
	contextID     string
	startTime     time.Time
	capabilities  []string
	environment   map[string]string
}

// NewFunctionalTestContext creates a new immutable test context
func NewFunctionalTestContext(config FunctionalCoreConfig) FunctionalTestContext {
	return FunctionalTestContext{
		config:        config,
		awsConfig:     mo.None[aws.Config](),
		contextID:     generateContextID(),
		startTime:     time.Now(),
		capabilities:  []string{},
		environment:   make(map[string]string),
	}
}

// Accessor methods for FunctionalTestContext
func (ctx FunctionalTestContext) GetConfig() FunctionalCoreConfig        { return ctx.config }
func (ctx FunctionalTestContext) GetAWSConfig() mo.Option[aws.Config]    { return ctx.awsConfig }
func (ctx FunctionalTestContext) GetContextID() string                   { return ctx.contextID }
func (ctx FunctionalTestContext) GetStartTime() time.Time               { return ctx.startTime }
func (ctx FunctionalTestContext) GetCapabilities() []string             { return ctx.capabilities }
func (ctx FunctionalTestContext) GetEnvironment() map[string]string     { return ctx.environment }

// Functional options for test context
func (ctx FunctionalTestContext) WithAWSConfig(awsConfig aws.Config) FunctionalTestContext {
	return FunctionalTestContext{
		config:        ctx.config,
		awsConfig:     mo.Some(awsConfig),
		contextID:     ctx.contextID,
		startTime:     ctx.startTime,
		capabilities:  ctx.capabilities,
		environment:   ctx.environment,
	}
}

func (ctx FunctionalTestContext) WithCapabilities(capabilities []string) FunctionalTestContext {
	return FunctionalTestContext{
		config:        ctx.config,
		awsConfig:     ctx.awsConfig,
		contextID:     ctx.contextID,
		startTime:     ctx.startTime,
		capabilities:  capabilities,
		environment:   ctx.environment,
	}
}

func (ctx FunctionalTestContext) WithEnvironment(environment map[string]string) FunctionalTestContext {
	return FunctionalTestContext{
		config:        ctx.config,
		awsConfig:     ctx.awsConfig,
		contextID:     ctx.contextID,
		startTime:     ctx.startTime,
		capabilities:  ctx.capabilities,
		environment:   environment,
	}
}

// FunctionalPerformanceMetrics represents immutable performance metrics
type FunctionalPerformanceMetrics struct {
	executionTime     time.Duration
	memoryUsage       uint64
	allocationsCount  uint64
	gcCycles          uint32
	cpuUsage          float64
	throughputOps     float64
	startTime         time.Time
	endTime           time.Time
	operationName     string
	tags              map[string]string
}

// NewFunctionalPerformanceMetrics creates immutable performance metrics
func NewFunctionalPerformanceMetrics(
	executionTime time.Duration,
	memoryUsage uint64,
	allocationsCount uint64,
	gcCycles uint32,
	operationName string,
) FunctionalPerformanceMetrics {
	return FunctionalPerformanceMetrics{
		executionTime:     executionTime,
		memoryUsage:       memoryUsage,
		allocationsCount:  allocationsCount,
		gcCycles:          gcCycles,
		cpuUsage:          0.0,
		throughputOps:     0.0,
		startTime:         time.Now().Add(-executionTime),
		endTime:           time.Now(),
		operationName:     operationName,
		tags:              make(map[string]string),
	}
}

// Accessor methods for FunctionalPerformanceMetrics
func (m FunctionalPerformanceMetrics) GetExecutionTime() time.Duration     { return m.executionTime }
func (m FunctionalPerformanceMetrics) GetMemoryUsage() uint64              { return m.memoryUsage }
func (m FunctionalPerformanceMetrics) GetAllocationsCount() uint64         { return m.allocationsCount }
func (m FunctionalPerformanceMetrics) GetGCCycles() uint32                 { return m.gcCycles }
func (m FunctionalPerformanceMetrics) GetCPUUsage() float64                { return m.cpuUsage }
func (m FunctionalPerformanceMetrics) GetThroughputOps() float64           { return m.throughputOps }
func (m FunctionalPerformanceMetrics) GetStartTime() time.Time             { return m.startTime }
func (m FunctionalPerformanceMetrics) GetEndTime() time.Time               { return m.endTime }
func (m FunctionalPerformanceMetrics) GetOperationName() string            { return m.operationName }
func (m FunctionalPerformanceMetrics) GetTags() map[string]string          { return m.tags }

// FunctionalRetryStrategy represents immutable retry strategy
type FunctionalRetryStrategy struct {
	maxRetries        int
	baseDelay         time.Duration
	maxDelay          time.Duration
	multiplier        float64
	jitterEnabled     bool
	exponentialBackoff bool
	retryableErrors   []string
}

// NewFunctionalRetryStrategy creates a new immutable retry strategy
func NewFunctionalRetryStrategy(maxRetries int, baseDelay time.Duration) FunctionalRetryStrategy {
	return FunctionalRetryStrategy{
		maxRetries:        maxRetries,
		baseDelay:         baseDelay,
		maxDelay:          30 * time.Second,
		multiplier:        2.0,
		jitterEnabled:     true,
		exponentialBackoff: true,
		retryableErrors:   []string{"timeout", "throttling", "service unavailable", "internal server error"},
	}
}

// Accessor methods for FunctionalRetryStrategy
func (s FunctionalRetryStrategy) GetMaxRetries() int             { return s.maxRetries }
func (s FunctionalRetryStrategy) GetBaseDelay() time.Duration    { return s.baseDelay }
func (s FunctionalRetryStrategy) GetMaxDelay() time.Duration     { return s.maxDelay }
func (s FunctionalRetryStrategy) GetMultiplier() float64         { return s.multiplier }
func (s FunctionalRetryStrategy) IsJitterEnabled() bool         { return s.jitterEnabled }
func (s FunctionalRetryStrategy) IsExponentialBackoff() bool    { return s.exponentialBackoff }
func (s FunctionalRetryStrategy) GetRetryableErrors() []string  { return s.retryableErrors }

// Functional options for retry strategy
func (s FunctionalRetryStrategy) WithMaxDelay(maxDelay time.Duration) FunctionalRetryStrategy {
	return FunctionalRetryStrategy{
		maxRetries:        s.maxRetries,
		baseDelay:         s.baseDelay,
		maxDelay:          maxDelay,
		multiplier:        s.multiplier,
		jitterEnabled:     s.jitterEnabled,
		exponentialBackoff: s.exponentialBackoff,
		retryableErrors:   s.retryableErrors,
	}
}

func (s FunctionalRetryStrategy) WithMultiplier(multiplier float64) FunctionalRetryStrategy {
	return FunctionalRetryStrategy{
		maxRetries:        s.maxRetries,
		baseDelay:         s.baseDelay,
		maxDelay:          s.maxDelay,
		multiplier:        multiplier,
		jitterEnabled:     s.jitterEnabled,
		exponentialBackoff: s.exponentialBackoff,
		retryableErrors:   s.retryableErrors,
	}
}

func (s FunctionalRetryStrategy) WithJitter(enabled bool) FunctionalRetryStrategy {
	return FunctionalRetryStrategy{
		maxRetries:        s.maxRetries,
		baseDelay:         s.baseDelay,
		maxDelay:          s.maxDelay,
		multiplier:        s.multiplier,
		jitterEnabled:     enabled,
		exponentialBackoff: s.exponentialBackoff,
		retryableErrors:   s.retryableErrors,
	}
}

func (s FunctionalRetryStrategy) WithRetryableErrors(errors []string) FunctionalRetryStrategy {
	return FunctionalRetryStrategy{
		maxRetries:        s.maxRetries,
		baseDelay:         s.baseDelay,
		maxDelay:          s.maxDelay,
		multiplier:        s.multiplier,
		jitterEnabled:     s.jitterEnabled,
		exponentialBackoff: s.exponentialBackoff,
		retryableErrors:   errors,
	}
}

// FunctionalCoreOperationResult represents the result of a core operation using monads
type FunctionalCoreOperationResult struct {
	result         mo.Option[interface{}]
	err            mo.Option[error]
	duration       time.Duration
	metrics        mo.Option[FunctionalPerformanceMetrics]
	retryAttempts  int
	metadata       map[string]interface{}
}

// NewFunctionalCoreOperationResult creates a new operation result
func NewFunctionalCoreOperationResult(
	result mo.Option[interface{}],
	err mo.Option[error],
	duration time.Duration,
	metadata map[string]interface{},
) FunctionalCoreOperationResult {
	return FunctionalCoreOperationResult{
		result:         result,
		err:            err,
		duration:       duration,
		metrics:        mo.None[FunctionalPerformanceMetrics](),
		retryAttempts:  0,
		metadata:       metadata,
	}
}

// IsSuccess returns true if the operation was successful
func (r FunctionalCoreOperationResult) IsSuccess() bool {
	return r.err.IsAbsent()
}

// GetResult returns the operation result
func (r FunctionalCoreOperationResult) GetResult() mo.Option[interface{}] {
	return r.result
}

// GetError returns the operation error
func (r FunctionalCoreOperationResult) GetError() mo.Option[error] {
	return r.err
}

// GetDuration returns the operation duration
func (r FunctionalCoreOperationResult) GetDuration() time.Duration {
	return r.duration
}

// GetMetrics returns the performance metrics
func (r FunctionalCoreOperationResult) GetMetrics() mo.Option[FunctionalPerformanceMetrics] {
	return r.metrics
}

// GetRetryAttempts returns the number of retry attempts
func (r FunctionalCoreOperationResult) GetRetryAttempts() int {
	return r.retryAttempts
}

// GetMetadata returns the operation metadata
func (r FunctionalCoreOperationResult) GetMetadata() map[string]interface{} {
	return r.metadata
}

// Functional options for operation result
func (r FunctionalCoreOperationResult) WithMetrics(metrics FunctionalPerformanceMetrics) FunctionalCoreOperationResult {
	return FunctionalCoreOperationResult{
		result:         r.result,
		err:            r.err,
		duration:       r.duration,
		metrics:        mo.Some(metrics),
		retryAttempts:  r.retryAttempts,
		metadata:       r.metadata,
	}
}

func (r FunctionalCoreOperationResult) WithRetryAttempts(attempts int) FunctionalCoreOperationResult {
	return FunctionalCoreOperationResult{
		result:         r.result,
		err:            r.err,
		duration:       r.duration,
		metrics:        r.metrics,
		retryAttempts:  attempts,
		metadata:       r.metadata,
	}
}

// Validation functions using functional programming patterns
type FunctionalCoreValidationRule[T any] func(T) mo.Option[error]

// validateRegion validates AWS region format
func validateRegion(region string) mo.Option[error] {
	if region == "" {
		return mo.Some[error](fmt.Errorf("region cannot be empty"))
	}

	// AWS regions follow pattern: us-east-1, eu-west-1, etc.
	validRegions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
		"ap-south-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2",
		"ca-central-1", "sa-east-1", "af-south-1", "me-south-1",
	}

	if !lo.Contains(validRegions, region) {
		return mo.Some[error](fmt.Errorf("invalid AWS region: %s", region))
	}

	return mo.None[error]()
}

// validateTimeout validates timeout duration
func validateTimeout(timeout time.Duration) mo.Option[error] {
	if timeout <= 0 {
		return mo.Some[error](fmt.Errorf("timeout must be positive"))
	}

	maxTimeout := 60 * time.Minute
	if timeout > maxTimeout {
		return mo.Some[error](fmt.Errorf("timeout cannot exceed %v", maxTimeout))
	}

	return mo.None[error]()
}

// validateRetryConfig validates retry configuration
func validateRetryConfig(maxRetries int, retryDelay time.Duration) mo.Option[error] {
	if maxRetries < 0 {
		return mo.Some[error](fmt.Errorf("max retries cannot be negative"))
	}

	if maxRetries > FunctionalCoreMaxRetryAttempts {
		return mo.Some[error](fmt.Errorf("max retries cannot exceed %d", FunctionalCoreMaxRetryAttempts))
	}

	if retryDelay < 0 {
		return mo.Some[error](fmt.Errorf("retry delay cannot be negative"))
	}

	return mo.None[error]()
}

// validateNamespace validates namespace format
func validateNamespace(namespace string) mo.Option[error] {
	if namespace == "" {
		return mo.None[error]() // Empty namespace is valid
	}

	if len(namespace) > FunctionalCoreMaxNamespaceLength {
		return mo.Some[error](fmt.Errorf("namespace cannot exceed %d characters", FunctionalCoreMaxNamespaceLength))
	}

	// Namespace should contain only alphanumeric characters, hyphens, and underscores
	for _, char := range namespace {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return mo.Some[error](fmt.Errorf("namespace contains invalid character: %c", char))
		}
	}

	return mo.None[error]()
}

// Functional validation pipeline
func validateFunctionalCoreConfig(config FunctionalCoreConfig) mo.Option[error] {
	validators := []func() mo.Option[error]{
		func() mo.Option[error] { return validateRegion(config.region) },
		func() mo.Option[error] { return validateTimeout(config.timeout) },
		func() mo.Option[error] { return validateRetryConfig(config.maxRetries, config.retryDelay) },
		func() mo.Option[error] {
			return config.namespace.
				Map(validateNamespace).
				OrElse(mo.None[error]())
		},
	}

	for _, validator := range validators {
		if err := validator(); err.IsPresent() {
			return err
		}
	}

	return mo.None[error]()
}

// Retry logic with exponential backoff and jitter using generics
func withFunctionalCoreRetry[T any](
	operation func() (T, error), 
	strategy FunctionalRetryStrategy,
) (T, error, int) {
	var result T
	var lastErr error

	for attempt := 0; attempt < strategy.maxRetries; attempt++ {
		if attempt > 0 {
			delay := calculateRetryDelay(attempt-1, strategy)
			time.Sleep(delay)
		}

		result, lastErr = operation()
		if lastErr == nil {
			return result, nil, attempt
		}

		// Check if error is retryable
		if !isRetryableError(lastErr, strategy.retryableErrors) {
			break
		}
	}

	return result, lastErr, strategy.maxRetries
}

// calculateRetryDelay calculates the delay for a retry attempt with optional jitter
func calculateRetryDelay(attempt int, strategy FunctionalRetryStrategy) time.Duration {
	if !strategy.exponentialBackoff {
		delay := strategy.baseDelay
		if strategy.jitterEnabled {
			jitter := time.Duration(lo.RandomInt(0, int(delay.Milliseconds()/4))) * time.Millisecond
			delay += jitter
		}
		return delay
	}

	// Exponential backoff: baseDelay * (multiplier ^ attempt)
	delay := time.Duration(float64(strategy.baseDelay) * (strategy.multiplier * float64(attempt+1)))
	
	// Cap at maximum delay
	if delay > strategy.maxDelay {
		delay = strategy.maxDelay
	}

	// Add jitter if enabled (up to 25% of delay)
	if strategy.jitterEnabled {
		maxJitter := int(delay.Milliseconds() / 4)
		if maxJitter > 0 {
			jitter := time.Duration(lo.RandomInt(0, maxJitter)) * time.Millisecond
			delay += jitter
		}
	}

	return delay
}

// isRetryableError determines if an error should be retried
func isRetryableError(err error, retryableErrors []string) bool {
	if err == nil {
		return false
	}

	errorMessage := strings.ToLower(err.Error())
	
	return lo.Some(retryableErrors, func(retryableErr string) bool {
		return strings.Contains(errorMessage, strings.ToLower(retryableErr))
	})
}

// measurePerformance executes an operation and returns performance metrics
func measurePerformance[T any](
	operation func() (T, error),
	operationName string,
) (T, error, FunctionalPerformanceMetrics) {
	var memStatsBefore, memStatsAfter runtime.MemStats
	
	// Capture initial memory stats
	runtime.ReadMemStats(&memStatsBefore)
	runtime.GC() // Force GC for accurate baseline
	
	startTime := time.Now()
	result, err := operation()
	endTime := time.Now()
	
	// Capture final memory stats
	runtime.ReadMemStats(&memStatsAfter)
	
	metrics := NewFunctionalPerformanceMetrics(
		endTime.Sub(startTime),
		memStatsAfter.Alloc - memStatsBefore.Alloc,
		memStatsAfter.Mallocs - memStatsBefore.Mallocs,
		memStatsAfter.NumGC - memStatsBefore.NumGC,
		operationName,
	)
	
	return result, err, metrics
}

// generateContextID generates a unique context identifier
func generateContextID() string {
	return fmt.Sprintf("ctx-%d-%d", time.Now().Unix(), lo.RandomInt(1000, 9999))
}

// combineConfigs merges multiple configurations using functional composition
func combineConfigs(configs ...FunctionalCoreConfig) FunctionalCoreConfig {
	if len(configs) == 0 {
		return NewFunctionalCoreConfig()
	}

	baseConfig := configs[0]
	
	// Apply each subsequent config's non-default values
	for _, config := range configs[1:] {
		if config.region != FunctionalCoreDefaultRegion {
			baseConfig.region = config.region
		}
		if config.timeout != FunctionalCoreDefaultTimeout {
			baseConfig.timeout = config.timeout
		}
		if config.maxRetries != FunctionalCoreDefaultMaxRetries {
			baseConfig.maxRetries = config.maxRetries
		}
		if config.retryDelay != FunctionalCoreDefaultRetryDelay {
			baseConfig.retryDelay = config.retryDelay
		}
		if config.namespace.IsPresent() {
			baseConfig.namespace = config.namespace
		}
		if config.awsProfile.IsPresent() {
			baseConfig.awsProfile = config.awsProfile
		}
		
		// Merge metadata
		for k, v := range config.metadata {
			baseConfig.metadata[k] = v
		}
	}

	return baseConfig
}

// Simulated core operations for testing
func SimulateFunctionalCoreOperation(operationName string, config FunctionalCoreConfig) FunctionalCoreOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":    operationName,
		"context_id":   generateContextID(),
		"region":       config.region,
		"phase":        "validation",
		"timestamp":    startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate configuration
	if configErr := validateFunctionalCoreConfig(config); configErr.IsPresent() {
		return NewFunctionalCoreOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	metadata["phase"] = "execution"

	// Create retry strategy
	retryStrategy := NewFunctionalRetryStrategy(config.maxRetries, config.retryDelay)

	// Execute operation with retry, performance measurement
	result, err, retryAttempts := withFunctionalCoreRetry(func() (map[string]interface{}, error) {
		// Measure performance of the operation
		operationResult, operationErr, metrics := measurePerformance(func() (map[string]interface{}, error) {
			// Simulate core operation
			time.Sleep(time.Duration(lo.RandomInt(10, 100)) * time.Millisecond)
			
			return map[string]interface{}{
				"operation":     operationName,
				"success":       true,
				"context_id":    metadata["context_id"],
				"region":        config.region,
				"namespace":     config.namespace.OrEmpty(),
				"execution_time": time.Since(startTime).Milliseconds(),
			}, nil
		}, operationName)

		// Store metrics in metadata
		metadata["performance"] = map[string]interface{}{
			"execution_time_ms": metrics.GetExecutionTime().Milliseconds(),
			"memory_usage":      metrics.GetMemoryUsage(),
			"allocations":       metrics.GetAllocationsCount(),
			"gc_cycles":         metrics.GetGCCycles(),
		}

		return operationResult, operationErr
	}, retryStrategy)

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()
	metadata["retry_attempts"] = retryAttempts

	if err != nil {
		return NewFunctionalCoreOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			duration,
			metadata,
		).WithRetryAttempts(retryAttempts)
	}

	return NewFunctionalCoreOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	).WithRetryAttempts(retryAttempts)
}

// SimulateFunctionalCoreHealthCheck simulates a health check operation
func SimulateFunctionalCoreHealthCheck(services []string, config FunctionalCoreConfig) FunctionalCoreOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":      "health_check",
		"services_count": len(services),
		"phase":          "validation",
		"timestamp":      startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate configuration
	if configErr := validateFunctionalCoreConfig(config); configErr.IsPresent() {
		return NewFunctionalCoreOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	metadata["phase"] = "execution"

	// Simulate health checks for each service
	serviceResults := lo.Map(services, func(service string, index int) map[string]interface{} {
		// Simulate health check with some random outcomes
		healthy := lo.RandomInt(1, 10) > 2 // 80% chance of being healthy
		responseTime := time.Duration(lo.RandomInt(10, 200)) * time.Millisecond
		
		result := map[string]interface{}{
			"service":       service,
			"healthy":       healthy,
			"response_time": responseTime.Milliseconds(),
			"checked_at":    time.Now().Format(time.RFC3339),
		}

		if !healthy {
			result["error"] = fmt.Sprintf("Service %s is unhealthy", service)
		}

		return result
	})

	// Calculate overall health
	healthyCount := lo.CountBy(serviceResults, func(result map[string]interface{}) bool {
		return result["healthy"].(bool)
	})

	overallHealthy := healthyCount == len(services)
	healthPercentage := float64(healthyCount) / float64(len(services)) * 100

	result := map[string]interface{}{
		"overall_healthy":   overallHealthy,
		"health_percentage": healthPercentage,
		"services":          serviceResults,
		"total_services":    len(services),
		"healthy_services":  healthyCount,
		"unhealthy_services": len(services) - healthyCount,
		"check_duration":    time.Since(startTime).Milliseconds(),
	}

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()

	return NewFunctionalCoreOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	)
}