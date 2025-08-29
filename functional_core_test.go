package vasdeference

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
)

// TestFunctionalCoreScenario tests a complete functional core scenario
func TestFunctionalCoreScenario(t *testing.T) {
	// Create core configuration
	config := NewFunctionalCoreConfig(
		WithCoreRegion("us-east-1"),
		WithCoreTimeout(30*time.Second),
		WithCoreRetryConfig(3, 200*time.Millisecond),
		WithCoreNamespace("test-namespace"),
		WithCoreAWSProfile("test-profile"),
		WithCoreLogging(true),
		WithCoreMetrics(true),
		WithCoreResilience(true),
		WithCoreMetadata("environment", "test"),
		WithCoreMetadata("scenario", "full_workflow"),
	)

	// Step 1: Create test context
	ctx := NewFunctionalTestContext(config)

	// Verify test context properties
	if ctx.GetConfig().GetRegion() != "us-east-1" {
		t.Errorf("Expected region to be 'us-east-1', got %s", ctx.GetConfig().GetRegion())
	}

	if ctx.GetContextID() == "" {
		t.Error("Expected context ID to be generated")
	}

	if ctx.GetStartTime().IsZero() {
		t.Error("Expected start time to be set")
	}

	// Step 2: Execute core operation
	operationResult := SimulateFunctionalCoreOperation("test-operation", config)

	if !operationResult.IsSuccess() {
		t.Errorf("Expected core operation to succeed, got error: %v", operationResult.GetError().MustGet())
	}

	// Verify operation result
	if operationResult.GetResult().IsPresent() {
		result, ok := operationResult.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected core operation result to be map[string]interface{}")
		} else {
			if result["operation"] != "test-operation" {
				t.Errorf("Expected operation name to be 'test-operation', got %v", result["operation"])
			}
			if result["success"] != true {
				t.Error("Expected operation to be successful")
			}
			if result["region"] != "us-east-1" {
				t.Errorf("Expected region to be 'us-east-1', got %v", result["region"])
			}
		}
	}

	// Step 3: Execute health check operation
	services := []string{"lambda", "dynamodb", "s3", "eventbridge", "stepfunctions"}
	healthCheckResult := SimulateFunctionalCoreHealthCheck(services, config)

	if !healthCheckResult.IsSuccess() {
		t.Errorf("Expected health check to succeed, got error: %v", healthCheckResult.GetError().MustGet())
	}

	// Verify health check result
	if healthCheckResult.GetResult().IsPresent() {
		result, ok := healthCheckResult.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected health check result to be map[string]interface{}")
		} else {
			if result["total_services"] != len(services) {
				t.Errorf("Expected total services to be %d, got %v", len(services), result["total_services"])
			}
			
			servicesResult, ok := result["services"].([]map[string]interface{})
			if !ok {
				t.Error("Expected services result to be []map[string]interface{}")
			} else if len(servicesResult) != len(services) {
				t.Errorf("Expected %d service results, got %d", len(services), len(servicesResult))
			}
		}
	}

	// Step 4: Verify metadata accumulation across operations
	allResults := []FunctionalCoreOperationResult{
		operationResult,
		healthCheckResult,
	}

	for i, result := range allResults {
		metadata := result.GetMetadata()
		if metadata["environment"] != "test" {
			t.Errorf("Operation %d missing environment metadata, metadata: %+v", i, metadata)
		}
		if result.GetDuration() == 0 {
			t.Errorf("Operation %d should have recorded duration", i)
		}
		if metadata["phase"] != "completed" {
			t.Errorf("Operation %d should have completed phase, got %v", i, metadata["phase"])
		}
	}
}

// TestFunctionalCoreErrorHandling tests comprehensive error handling
func TestFunctionalCoreErrorHandling(t *testing.T) {
	// Test invalid region
	invalidRegionConfig := NewFunctionalCoreConfig(
		WithCoreRegion("invalid-region"),
	)
	
	result := SimulateFunctionalCoreOperation("test-operation", invalidRegionConfig)

	if result.IsSuccess() {
		t.Error("Expected operation to fail with invalid region")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	// Test invalid timeout
	invalidTimeoutConfig := NewFunctionalCoreConfig(
		WithCoreTimeout(-1*time.Second),
	)
	
	result2 := SimulateFunctionalCoreOperation("test-operation", invalidTimeoutConfig)

	if result2.IsSuccess() {
		t.Error("Expected operation to fail with invalid timeout")
	}

	// Test invalid retry config
	invalidRetryConfig := NewFunctionalCoreConfig(
		WithCoreRetryConfig(-1, 100*time.Millisecond),
	)
	
	result3 := SimulateFunctionalCoreOperation("test-operation", invalidRetryConfig)

	if result3.IsSuccess() {
		t.Error("Expected operation to fail with invalid retry config")
	}

	// Test invalid namespace
	invalidNamespaceConfig := NewFunctionalCoreConfig(
		WithCoreNamespace("invalid@namespace"),
	)
	
	result4 := SimulateFunctionalCoreOperation("test-operation", invalidNamespaceConfig)

	if result4.IsSuccess() {
		t.Error("Expected operation to fail with invalid namespace")
	}

	// Verify error metadata consistency
	errorResults := []FunctionalCoreOperationResult{result, result2, result3, result4}
	for i, errorResult := range errorResults {
		metadata := errorResult.GetMetadata()
		if metadata["phase"] != "validation" {
			t.Errorf("Error result %d should have validation phase, got %v", i, metadata["phase"])
		}
	}
}

// TestFunctionalCoreConfigurationCombinations tests various config combinations
func TestFunctionalCoreConfigurationCombinations(t *testing.T) {
	// Test minimal configuration
	minimalConfig := NewFunctionalCoreConfig()
	
	if minimalConfig.GetRegion() != FunctionalCoreDefaultRegion {
		t.Error("Expected minimal config to use default region")
	}
	
	if minimalConfig.GetTimeout() != FunctionalCoreDefaultTimeout {
		t.Error("Expected minimal config to use default timeout")
	}

	if minimalConfig.GetMaxRetries() != FunctionalCoreDefaultMaxRetries {
		t.Error("Expected minimal config to use default max retries")
	}

	// Test maximal configuration
	maximalConfig := NewFunctionalCoreConfig(
		WithCoreRegion("eu-west-1"),
		WithCoreTimeout(60*time.Minute),
		WithCoreRetryConfig(10, 5*time.Second),
		WithCoreNamespace("production-namespace"),
		WithCoreAWSProfile("production-profile"),
		WithCoreLogging(false),
		WithCoreMetrics(true),
		WithCoreResilience(true),
		WithCoreMetadata("service", "core-service"),
		WithCoreMetadata("environment", "production"),
		WithCoreMetadata("version", "v3.0.0"),
	)

	// Verify all options were applied
	if maximalConfig.GetRegion() != "eu-west-1" {
		t.Error("Expected maximal config region")
	}

	if maximalConfig.GetTimeout() != 60*time.Minute {
		t.Error("Expected maximal config timeout")
	}

	if maximalConfig.GetMaxRetries() != 10 {
		t.Error("Expected maximal config retries")
	}

	if maximalConfig.GetRetryDelay() != 5*time.Second {
		t.Error("Expected maximal config retry delay")
	}

	if maximalConfig.GetNamespace().IsAbsent() {
		t.Error("Expected maximal config namespace")
	}

	if maximalConfig.GetAWSProfile().IsAbsent() {
		t.Error("Expected maximal config AWS profile")
	}

	if maximalConfig.IsLoggingEnabled() {
		t.Error("Expected maximal config logging to be disabled")
	}

	if !maximalConfig.IsMetricsEnabled() {
		t.Error("Expected maximal config metrics to be enabled")
	}

	if !maximalConfig.IsResilienceEnabled() {
		t.Error("Expected maximal config resilience to be enabled")
	}

	if len(maximalConfig.GetMetadata()) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(maximalConfig.GetMetadata()))
	}
}

// TestFunctionalCoreRetryStrategy tests retry strategy functionality
func TestFunctionalCoreRetryStrategy(t *testing.T) {
	// Test basic retry strategy
	basicStrategy := NewFunctionalRetryStrategy(3, 100*time.Millisecond)

	if basicStrategy.GetMaxRetries() != 3 {
		t.Error("Expected basic strategy max retries to be 3")
	}

	if basicStrategy.GetBaseDelay() != 100*time.Millisecond {
		t.Error("Expected basic strategy base delay to be 100ms")
	}

	if !basicStrategy.IsJitterEnabled() {
		t.Error("Expected basic strategy to have jitter enabled by default")
	}

	if !basicStrategy.IsExponentialBackoff() {
		t.Error("Expected basic strategy to have exponential backoff by default")
	}

	// Test retry strategy with custom options
	customStrategy := basicStrategy.
		WithMaxDelay(10*time.Second).
		WithMultiplier(3.0).
		WithJitter(false).
		WithRetryableErrors([]string{"custom-error", "another-error"})

	if customStrategy.GetMaxDelay() != 10*time.Second {
		t.Error("Expected custom strategy max delay")
	}

	if customStrategy.GetMultiplier() != 3.0 {
		t.Error("Expected custom strategy multiplier")
	}

	if customStrategy.IsJitterEnabled() {
		t.Error("Expected custom strategy jitter to be disabled")
	}

	if len(customStrategy.GetRetryableErrors()) != 2 {
		t.Error("Expected custom strategy to have 2 retryable errors")
	}
}

// TestFunctionalCorePerformanceMetrics tests performance metrics functionality
func TestFunctionalCorePerformanceMetrics(t *testing.T) {
	// Create performance metrics
	executionTime := 250 * time.Millisecond
	memoryUsage := uint64(1024 * 1024) // 1MB
	allocationsCount := uint64(100)
	gcCycles := uint32(2)
	operationName := "test-performance"

	metrics := NewFunctionalPerformanceMetrics(
		executionTime,
		memoryUsage,
		allocationsCount,
		gcCycles,
		operationName,
	)

	// Verify metrics properties
	if metrics.GetExecutionTime() != executionTime {
		t.Error("Expected metrics execution time to match")
	}

	if metrics.GetMemoryUsage() != memoryUsage {
		t.Error("Expected metrics memory usage to match")
	}

	if metrics.GetAllocationsCount() != allocationsCount {
		t.Error("Expected metrics allocations count to match")
	}

	if metrics.GetGCCycles() != gcCycles {
		t.Error("Expected metrics GC cycles to match")
	}

	if metrics.GetOperationName() != operationName {
		t.Error("Expected metrics operation name to match")
	}

	if metrics.GetStartTime().IsZero() {
		t.Error("Expected metrics start time to be set")
	}

	if metrics.GetEndTime().IsZero() {
		t.Error("Expected metrics end time to be set")
	}
}

// TestFunctionalCoreTestContext tests test context functionality
func TestFunctionalCoreTestContext(t *testing.T) {
	config := NewFunctionalCoreConfig(
		WithCoreRegion("us-west-2"),
		WithCoreNamespace("context-test"),
	)

	// Create test context
	ctx := NewFunctionalTestContext(config)

	// Verify initial state
	if ctx.GetConfig().GetRegion() != "us-west-2" {
		t.Error("Expected context config region to match")
	}

	if ctx.GetContextID() == "" {
		t.Error("Expected context ID to be generated")
	}

	if ctx.GetAWSConfig().IsPresent() {
		t.Error("Expected AWS config to be absent initially")
	}

	if len(ctx.GetCapabilities()) != 0 {
		t.Error("Expected initial capabilities to be empty")
	}

	// Test context modifications
	capabilities := []string{"lambda", "dynamodb", "s3"}
	environment := map[string]string{
		"ENV": "test",
		"REGION": "us-west-2",
	}

	modifiedCtx := ctx.
		WithCapabilities(capabilities).
		WithEnvironment(environment)

	if len(modifiedCtx.GetCapabilities()) != 3 {
		t.Error("Expected modified context to have 3 capabilities")
	}

	if len(modifiedCtx.GetEnvironment()) != 2 {
		t.Error("Expected modified context to have 2 environment variables")
	}

	// Original context should remain unchanged (immutability)
	if len(ctx.GetCapabilities()) != 0 {
		t.Error("Original context should remain unchanged")
	}
}

// TestFunctionalCoreValidationRules tests validation rules
func TestFunctionalCoreValidationRules(t *testing.T) {
	// Test region validation
	validRegions := []string{
		"us-east-1", "us-west-2", "eu-west-1", "ap-northeast-1",
	}

	invalidRegions := []string{
		"", "invalid-region", "us-invalid-1", "europe-1",
	}

	for _, region := range validRegions {
		if validateRegion(region).IsPresent() {
			t.Errorf("Expected region %s to be valid", region)
		}
	}

	for _, region := range invalidRegions {
		if validateRegion(region).IsAbsent() {
			t.Errorf("Expected region %s to be invalid", region)
		}
	}

	// Test timeout validation
	validTimeouts := []time.Duration{
		1 * time.Second, 30 * time.Second, 5 * time.Minute,
	}

	invalidTimeouts := []time.Duration{
		0, -1 * time.Second, 61 * time.Minute,
	}

	for _, timeout := range validTimeouts {
		if validateTimeout(timeout).IsPresent() {
			t.Errorf("Expected timeout %v to be valid", timeout)
		}
	}

	for _, timeout := range invalidTimeouts {
		if validateTimeout(timeout).IsAbsent() {
			t.Errorf("Expected timeout %v to be invalid", timeout)
		}
	}

	// Test namespace validation
	validNamespaces := []string{
		"", "test", "test-namespace", "test_namespace", "TestNamespace123",
	}

	invalidNamespaces := []string{
		"invalid@namespace",
		"namespace with spaces",
		strings.Repeat("a", FunctionalCoreMaxNamespaceLength+1),
	}

	for _, namespace := range validNamespaces {
		if validateNamespace(namespace).IsPresent() {
			t.Errorf("Expected namespace %s to be valid", namespace)
		}
	}

	for _, namespace := range invalidNamespaces {
		if validateNamespace(namespace).IsAbsent() {
			t.Errorf("Expected namespace %s to be invalid", namespace)
		}
	}
}

// TestFunctionalCoreOperationResult tests operation result functionality
func TestFunctionalCoreOperationResult(t *testing.T) {
	// Test successful result
	successResult := NewFunctionalCoreOperationResult(
		mo.Some[interface{}](map[string]interface{}{"success": true}),
		mo.None[error](),
		100*time.Millisecond,
		map[string]interface{}{"operation": "test"},
	)

	if !successResult.IsSuccess() {
		t.Error("Expected result to be successful")
	}

	if successResult.GetResult().IsAbsent() {
		t.Error("Expected successful result to have result data")
	}

	if successResult.GetError().IsPresent() {
		t.Error("Expected successful result to have no error")
	}

	// Test failed result
	failureResult := NewFunctionalCoreOperationResult(
		mo.None[interface{}](),
		mo.Some[error](fmt.Errorf("test error")),
		50*time.Millisecond,
		map[string]interface{}{"operation": "test"},
	)

	if failureResult.IsSuccess() {
		t.Error("Expected result to be failed")
	}

	if failureResult.GetResult().IsPresent() {
		t.Error("Expected failed result to have no result data")
	}

	if failureResult.GetError().IsAbsent() {
		t.Error("Expected failed result to have error")
	}

	// Test result modification
	metrics := NewFunctionalPerformanceMetrics(
		100*time.Millisecond, 1024, 10, 1, "test-op",
	)

	modifiedResult := successResult.
		WithMetrics(metrics).
		WithRetryAttempts(2)

	if modifiedResult.GetMetrics().IsAbsent() {
		t.Error("Expected modified result to have metrics")
	}

	if modifiedResult.GetRetryAttempts() != 2 {
		t.Error("Expected modified result to have 2 retry attempts")
	}

	// Original result should remain unchanged (immutability)
	if successResult.GetMetrics().IsPresent() {
		t.Error("Original result should remain unchanged")
	}
}

// TestFunctionalCoreHealthCheck tests health check functionality
func TestFunctionalCoreHealthCheck(t *testing.T) {
	config := NewFunctionalCoreConfig(
		WithCoreMetadata("test_type", "health_check"),
	)

	// Test health check with multiple services
	services := []string{
		"lambda-service",
		"dynamodb-service", 
		"s3-service",
		"eventbridge-service",
		"stepfunctions-service",
	}

	result := SimulateFunctionalCoreHealthCheck(services, config)

	if !result.IsSuccess() {
		t.Errorf("Expected health check to succeed, got error: %v", result.GetError().MustGet())
	}

	// Verify health check result structure
	if result.GetResult().IsPresent() {
		healthResult, ok := result.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected health check result to be map[string]interface{}")
		} else {
			// Verify required fields
			requiredFields := []string{
				"overall_healthy", "health_percentage", "services", 
				"total_services", "healthy_services", "unhealthy_services",
			}

			for _, field := range requiredFields {
				if _, exists := healthResult[field]; !exists {
					t.Errorf("Expected health result to have field: %s", field)
				}
			}

			// Verify service count
			if healthResult["total_services"] != len(services) {
				t.Errorf("Expected total services to be %d, got %v", 
					len(services), healthResult["total_services"])
			}

			// Verify services array
			servicesResult, ok := healthResult["services"].([]map[string]interface{})
			if !ok {
				t.Error("Expected services to be []map[string]interface{}")
			} else {
				if len(servicesResult) != len(services) {
					t.Errorf("Expected %d service results, got %d", 
						len(services), len(servicesResult))
				}

				// Verify each service result structure
				for i, serviceResult := range servicesResult {
					requiredServiceFields := []string{"service", "healthy", "response_time", "checked_at"}
					for _, field := range requiredServiceFields {
						if _, exists := serviceResult[field]; !exists {
							t.Errorf("Service %d missing field: %s", i, field)
						}
					}
				}
			}
		}
	}
}

// TestFunctionalCoreConfigCombination tests configuration combination
func TestFunctionalCoreConfigCombination(t *testing.T) {
	// Create base config
	baseConfig := NewFunctionalCoreConfig(
		WithCoreRegion("us-east-1"),
		WithCoreTimeout(30*time.Second),
		WithCoreMetadata("base", "config"),
	)

	// Create override config
	overrideConfig := NewFunctionalCoreConfig(
		WithCoreRegion("eu-west-1"),
		WithCoreRetryConfig(5, 500*time.Millisecond),
		WithCoreMetadata("override", "config"),
		WithCoreMetadata("additional", "data"),
	)

	// Combine configs
	combinedConfig := combineConfigs(baseConfig, overrideConfig)

	// Verify combination results
	if combinedConfig.GetRegion() != "eu-west-1" {
		t.Error("Expected combined config to use override region")
	}

	if combinedConfig.GetTimeout() != 30*time.Second {
		t.Error("Expected combined config to keep base timeout")
	}

	if combinedConfig.GetMaxRetries() != 5 {
		t.Error("Expected combined config to use override retries")
	}

	metadata := combinedConfig.GetMetadata()
	if len(metadata) != 3 {
		t.Errorf("Expected combined metadata to have 3 entries, got %d", len(metadata))
	}

	if metadata["base"] != "config" {
		t.Error("Expected combined config to have base metadata")
	}

	if metadata["override"] != "config" {
		t.Error("Expected combined config to have override metadata")
	}

	if metadata["additional"] != "data" {
		t.Error("Expected combined config to have additional metadata")
	}
}

// BenchmarkFunctionalCoreScenario benchmarks a complete scenario
func BenchmarkFunctionalCoreScenario(b *testing.B) {
	config := NewFunctionalCoreConfig(
		WithCoreTimeout(5*time.Second),
		WithCoreRetryConfig(1, 10*time.Millisecond),
		WithCoreMetrics(true),
	)

	services := []string{"lambda", "dynamodb", "s3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate full functional core workflow
		ctx := NewFunctionalTestContext(config)
		if ctx.GetContextID() == "" {
			b.Error("Expected context ID to be generated")
		}

		operationResult := SimulateFunctionalCoreOperation(fmt.Sprintf("benchmark-op-%d", i), config)
		if !operationResult.IsSuccess() {
			b.Errorf("Core operation failed: %v", operationResult.GetError().MustGet())
		}

		healthResult := SimulateFunctionalCoreHealthCheck(services, config)
		if !healthResult.IsSuccess() {
			b.Errorf("Health check failed: %v", healthResult.GetError().MustGet())
		}
	}
}