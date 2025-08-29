package lambda

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// TestFunctionalLambdaConfig tests functional Lambda configuration
func TestFunctionalLambdaConfig(t *testing.T) {
	config := NewFunctionalLambdaConfig("test-function",
		WithInvocationType(types.InvocationTypeEvent),
		WithLogType(types.LogTypeTail),
		WithTimeout(45*time.Second),
		WithRetryConfig(5, 2*time.Second),
		WithQualifier("v1"),
		WithLambdaMetadata("environment", "test"),
		WithLambdaMetadata("version", "1.0"),
	)

	if config.functionName != "test-function" {
		t.Errorf("Expected function name 'test-function', got '%s'", config.functionName)
	}

	if config.invocationType != types.InvocationTypeEvent {
		t.Errorf("Expected invocation type Event, got %v", config.invocationType)
	}

	if config.logType != types.LogTypeTail {
		t.Errorf("Expected log type Tail, got %v", config.logType)
	}

	if config.timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", config.timeout)
	}

	if config.maxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", config.maxRetries)
	}

	if config.retryDelay != 2*time.Second {
		t.Errorf("Expected retry delay 2s, got %v", config.retryDelay)
	}

	if config.qualifier != "v1" {
		t.Errorf("Expected qualifier 'v1', got '%s'", config.qualifier)
	}

	if config.metadata["environment"] != "test" {
		t.Errorf("Expected environment 'test', got %v", config.metadata["environment"])
	}

	if config.metadata["version"] != "1.0" {
		t.Errorf("Expected version '1.0', got %v", config.metadata["version"])
	}
}

// TestFunctionalLambdaConfigPayloadValidator tests payload validation
func TestFunctionalLambdaConfigPayloadValidator(t *testing.T) {
	validator := func(payload string) error {
		if len(payload) == 0 {
			return fmt.Errorf("payload cannot be empty")
		}
		return nil
	}

	config := NewFunctionalLambdaConfig("test-function",
		WithPayloadValidator(validator),
	)

	if config.payloadValidator.IsAbsent() {
		t.Error("Expected payload validator to be present")
	}

	// Test that validator works
	if config.payloadValidator.IsPresent() {
		validatorFunc := config.payloadValidator.MustGet()
		err := validatorFunc("")
		if err == nil {
			t.Error("Expected validator to return error for empty payload")
		}

		err = validatorFunc("valid payload")
		if err != nil {
			t.Errorf("Expected validator to accept valid payload, got error: %v", err)
		}
	}
}

// TestFunctionalInvokeResult tests invoke result functionality
func TestFunctionalInvokeResult(t *testing.T) {
	// Test successful result
	invokeResult := InvokeResult{
		StatusCode:      200,
		Payload:         `{"result": "success"}`,
		LogResult:       "logs",
		ExecutedVersion: "1",
		FunctionError:   "",
		ExecutionTime:   100 * time.Millisecond,
	}

	result := NewSuccessFunctionalInvokeResult(
		invokeResult,
		200*time.Millisecond,
		map[string]interface{}{"test": "metadata"},
	)

	if !result.IsSuccess() {
		t.Error("Expected result to be successful")
	}

	if result.IsError() {
		t.Error("Expected result not to be error")
	}

	if result.GetResult().IsAbsent() {
		t.Error("Expected result to be present")
	}

	if result.GetError().IsPresent() {
		t.Error("Expected error to be absent")
	}

	if result.GetDuration() != 200*time.Millisecond {
		t.Errorf("Expected duration 200ms, got %v", result.GetDuration())
	}

	metadata := result.GetMetadata()
	if metadata["test"] != "metadata" {
		t.Errorf("Expected metadata test='metadata', got %v", metadata["test"])
	}
}

// TestFunctionalInvokeResultError tests error result functionality
func TestFunctionalInvokeResultError(t *testing.T) {
	err := fmt.Errorf("invocation failed")
	result := NewErrorFunctionalInvokeResult(
		err,
		100*time.Millisecond,
		map[string]interface{}{"phase": "validation"},
	)

	if result.IsSuccess() {
		t.Error("Expected result to be unsuccessful")
	}

	if !result.IsError() {
		t.Error("Expected result to be error")
	}

	if result.GetResult().IsPresent() {
		t.Error("Expected result to be absent")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	if result.GetError().MustGet().Error() != "invocation failed" {
		t.Errorf("Expected error 'invocation failed', got %v", result.GetError().MustGet())
	}
}

// TestFunctionalInvokeResultMap tests result mapping
func TestFunctionalInvokeResultMap(t *testing.T) {
	invokeResult := InvokeResult{
		StatusCode:    200,
		Payload:       `{"result": "success"}`,
		ExecutionTime: 100 * time.Millisecond,
	}

	result := NewSuccessFunctionalInvokeResult(
		invokeResult,
		200*time.Millisecond,
		map[string]interface{}{},
	)

	// Test mapping successful result
	mappedResult := result.MapResult(func(r InvokeResult) InvokeResult {
		return InvokeResult{
			StatusCode:    r.StatusCode,
			Payload:       "transformed: " + r.Payload,
			LogResult:     r.LogResult,
			ExecutedVersion: r.ExecutedVersion,
			FunctionError: r.FunctionError,
			ExecutionTime: r.ExecutionTime,
		}
	})

	if !mappedResult.IsSuccess() {
		t.Error("Expected mapped result to be successful")
	}

	if mappedResult.GetResult().IsAbsent() {
		t.Error("Expected mapped result to be present")
	}

	transformedPayload := mappedResult.GetResult().MustGet().Payload
	expectedPayload := `transformed: {"result": "success"}`
	if transformedPayload != expectedPayload {
		t.Errorf("Expected payload '%s', got '%s'", expectedPayload, transformedPayload)
	}

	// Test mapping error result - should not transform
	errorResult := NewErrorFunctionalInvokeResult(
		fmt.Errorf("test error"),
		100*time.Millisecond,
		map[string]interface{}{},
	)

	mappedErrorResult := errorResult.MapResult(func(r InvokeResult) InvokeResult {
		t.Error("Map function should not be called on error result")
		return r
	})

	if !mappedErrorResult.IsError() {
		t.Error("Expected mapped error result to still be error")
	}
}

// TestFunctionalPipeline tests the functional pipeline for validation
func TestFunctionalPipeline(t *testing.T) {
	// Test successful pipeline
	pipeline := ExecuteFunctionalPipeline[string]("hello",
		func(s string) (string, error) {
			if len(s) == 0 {
				return s, fmt.Errorf("string cannot be empty")
			}
			return s, nil
		},
		func(s string) (string, error) {
			return s + " world", nil
		},
		func(s string) (string, error) {
			if len(s) < 5 {
				return s, fmt.Errorf("string too short")
			}
			return s, nil
		},
	)

	if pipeline.IsError() {
		t.Errorf("Expected pipeline to succeed, got error: %v", pipeline.GetError().MustGet())
	}

	if pipeline.GetValue() != "hello world" {
		t.Errorf("Expected value 'hello world', got '%s'", pipeline.GetValue())
	}

	// Test failing pipeline
	failingPipeline := ExecuteFunctionalPipeline[string]("",
		func(s string) (string, error) {
			if len(s) == 0 {
				return s, fmt.Errorf("string cannot be empty")
			}
			return s, nil
		},
		func(s string) (string, error) {
			t.Error("This function should not be called after error")
			return s, nil
		},
	)

	if !failingPipeline.IsError() {
		t.Error("Expected pipeline to fail")
	}

	if failingPipeline.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	if failingPipeline.GetError().MustGet().Error() != "string cannot be empty" {
		t.Errorf("Expected error 'string cannot be empty', got %v", failingPipeline.GetError().MustGet())
	}
}

// TestFunctionalRetryWithBackoff tests retry functionality
func TestFunctionalRetryWithBackoff(t *testing.T) {
	// Test successful operation on first try
	attempt := 0
	successResult := FunctionalRetryWithBackoff(
		func() (string, error) {
			attempt++
			return "success", nil
		},
		3,
		10*time.Millisecond,
	)

	if successResult.IsError() {
		t.Error("Expected retry to succeed")
	}

	if successResult.GetResult().IsAbsent() {
		t.Error("Expected result to be present")
	}

	if successResult.GetResult().MustGet() != "success" {
		t.Errorf("Expected result 'success', got '%s'", successResult.GetResult().MustGet())
	}

	if attempt != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempt)
	}

	// Test operation that succeeds after retries
	attempt = 0
	retryResult := FunctionalRetryWithBackoff(
		func() (string, error) {
			attempt++
			if attempt < 3 {
				return "", fmt.Errorf("attempt %d failed", attempt)
			}
			return "success after retries", nil
		},
		5,
		5*time.Millisecond,
	)

	if retryResult.IsError() {
		t.Error("Expected retry to eventually succeed")
	}

	if retryResult.GetResult().MustGet() != "success after retries" {
		t.Errorf("Expected result 'success after retries', got '%s'", retryResult.GetResult().MustGet())
	}

	if attempt != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempt)
	}

	// Test operation that fails all retries
	attempt = 0
	failResult := FunctionalRetryWithBackoff(
		func() (string, error) {
			attempt++
			return "", fmt.Errorf("attempt %d failed", attempt)
		},
		3,
		1*time.Millisecond,
	)

	if !failResult.IsError() {
		t.Error("Expected retry to fail after all attempts")
	}

	if failResult.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	// Should attempt maxRetries times, plus one final attempt to capture error
	if attempt < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attempt)
	}
}

// TestFunctionalLambdaOptions tests functional option combinations
func TestFunctionalLambdaOptions(t *testing.T) {
	// Test combining multiple options
	config := NewFunctionalLambdaConfig("complex-function",
		WithInvocationType(types.InvocationTypeEvent),
		WithLogType(types.LogTypeTail),
		WithTimeout(60*time.Second),
		WithRetryConfig(10, 5*time.Second),
		WithQualifier("LATEST"),
		WithLambdaMetadata("service", "test-service"),
		WithLambdaMetadata("environment", "staging"),
		WithPayloadValidator(func(payload string) error {
			if len(payload) > 1000 {
				return fmt.Errorf("payload too large")
			}
			return nil
		}),
	)

	// Verify all options applied correctly
	if config.functionName != "complex-function" {
		t.Errorf("Expected function name 'complex-function', got '%s'", config.functionName)
	}

	if config.invocationType != types.InvocationTypeEvent {
		t.Error("Expected invocation type Event")
	}

	if config.logType != types.LogTypeTail {
		t.Error("Expected log type Tail")
	}

	if config.timeout != 60*time.Second {
		t.Error("Expected timeout 60s")
	}

	if config.maxRetries != 10 {
		t.Error("Expected max retries 10")
	}

	if config.retryDelay != 5*time.Second {
		t.Error("Expected retry delay 5s")
	}

	if config.qualifier != "LATEST" {
		t.Error("Expected qualifier 'LATEST'")
	}

	if config.metadata["service"] != "test-service" {
		t.Error("Expected service metadata to be 'test-service'")
	}

	if config.metadata["environment"] != "staging" {
		t.Error("Expected environment metadata to be 'staging'")
	}

	if config.payloadValidator.IsAbsent() {
		t.Error("Expected payload validator to be present")
	}
}

// TestFunctionalLambdaUtilities tests functional utilities used in Lambda operations
func TestFunctionalLambdaUtilities(t *testing.T) {
	// Test lo.Reduce for options application
	options := []LambdaConfigOption{
		WithTimeout(30 * time.Second),
		WithRetryConfig(5, 2*time.Second),
		WithLambdaMetadata("test", "value"),
	}

	baseConfig := NewFunctionalLambdaConfig("test-function")
	
	finalConfig := lo.Reduce(options, func(config FunctionalLambdaConfig, opt LambdaConfigOption, _ int) FunctionalLambdaConfig {
		return opt(config)
	}, baseConfig)

	if finalConfig.timeout != 30*time.Second {
		t.Error("Expected timeout to be applied via reduce")
	}

	if finalConfig.maxRetries != 5 {
		t.Error("Expected max retries to be applied via reduce")
	}

	if finalConfig.retryDelay != 2*time.Second {
		t.Error("Expected retry delay to be applied via reduce")
	}

	if finalConfig.metadata["test"] != "value" {
		t.Error("Expected metadata to be applied via reduce")
	}

	// Test lo.Assign for metadata merging
	originalMetadata := map[string]interface{}{"key1": "value1"}
	newMetadata := map[string]interface{}{"key2": "value2"}
	
	mergedMetadata := lo.Assign(originalMetadata, newMetadata)
	
	if mergedMetadata["key1"] != "value1" {
		t.Error("Expected original metadata to be preserved")
	}
	
	if mergedMetadata["key2"] != "value2" {
		t.Error("Expected new metadata to be merged")
	}
}

// BenchmarkFunctionalLambdaConfig benchmarks functional config creation
func BenchmarkFunctionalLambdaConfig(b *testing.B) {
	options := []LambdaConfigOption{
		WithInvocationType(types.InvocationTypeEvent),
		WithLogType(types.LogTypeTail),
		WithTimeout(30 * time.Second),
		WithRetryConfig(5, 2*time.Second),
		WithQualifier("v1"),
		WithLambdaMetadata("environment", "test"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFunctionalLambdaConfig("benchmark-function", options...)
	}
}

// BenchmarkFunctionalPipeline benchmarks functional pipeline execution
func BenchmarkFunctionalPipeline(b *testing.B) {
	operations := []func(string) (string, error){
		func(s string) (string, error) {
			if len(s) == 0 {
				return s, fmt.Errorf("empty string")
			}
			return s, nil
		},
		func(s string) (string, error) {
			return s + "_processed", nil
		},
		func(s string) (string, error) {
			if len(s) > 100 {
				return s, fmt.Errorf("string too long")
			}
			return s, nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExecuteFunctionalPipeline[string]("test_input", operations...)
	}
}