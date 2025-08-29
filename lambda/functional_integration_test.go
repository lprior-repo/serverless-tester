package lambda

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// TestSimulateFunctionalLambdaInvocation tests the simulated Lambda invocation
func TestSimulateFunctionalLambdaInvocation(t *testing.T) {
	config := NewFunctionalLambdaConfig("test-function",
		WithTimeout(30*time.Second),
		WithRetryConfig(3, 100*time.Millisecond),
		WithLambdaMetadata("environment", "test"),
	)

	payload := `{"message": "Hello from functional programming"}`

	result := SimulateFunctionalLambdaInvocation(payload, config)

	if !result.IsSuccess() {
		t.Errorf("Expected invocation to succeed, got error: %v", result.GetError().MustGet())
	}

	if result.GetResult().IsAbsent() {
		t.Error("Expected result to be present")
	}

	// Verify result structure
	if result.IsSuccess() && result.GetResult().IsPresent() {
		invokeResult := result.GetResult().MustGet()
		
		if invokeResult.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", invokeResult.StatusCode)
		}

		if len(invokeResult.Payload) == 0 {
			t.Error("Expected payload to be present")
		}

		if invokeResult.ExecutedVersion != "1" {
			t.Errorf("Expected executed version '1', got '%s'", invokeResult.ExecutedVersion)
		}
	}

	// Verify metadata
	metadata := result.GetMetadata()
	if metadata["function_name"] != "test-function" {
		t.Errorf("Expected function_name metadata, got %v", metadata["function_name"])
	}

	if metadata["phase"] != "success" {
		t.Errorf("Expected phase metadata to be 'success', got %v", metadata["phase"])
	}
}

// TestSimulateFunctionalLambdaInvocationValidationError tests validation error handling
func TestSimulateFunctionalLambdaInvocationValidationError(t *testing.T) {
	// Test with empty function name
	config := NewFunctionalLambdaConfig("",
		WithTimeout(30*time.Second),
	)

	payload := `{"message": "test"}`

	result := SimulateFunctionalLambdaInvocation(payload, config)

	if result.IsSuccess() {
		t.Error("Expected invocation to fail due to empty function name")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	// Verify error metadata
	metadata := result.GetMetadata()
	if metadata["phase"] != "validation" {
		t.Errorf("Expected phase metadata to be 'validation', got %v", metadata["phase"])
	}
}

// TestSimulateFunctionalLambdaInvocationWithPayloadValidator tests custom payload validation
func TestSimulateFunctionalLambdaInvocationWithPayloadValidator(t *testing.T) {
	validator := func(payload string) error {
		if len(payload) > 100 {
			return fmt.Errorf("payload too large for test")
		}
		return nil
	}

	config := NewFunctionalLambdaConfig("test-function",
		WithPayloadValidator(validator),
		WithLambdaMetadata("validation", "custom"),
	)

	// Test with acceptable payload
	smallPayload := `{"test": "small"}`
	result := SimulateFunctionalLambdaInvocation(smallPayload, config)

	if !result.IsSuccess() {
		t.Errorf("Expected invocation to succeed with small payload, got error: %v", result.GetError().MustGet())
	}

	// Test with payload that fails validation
	largePayload := `{"test": "` + string(make([]byte, 150)) + `"}`
	resultLarge := SimulateFunctionalLambdaInvocation(largePayload, config)

	if resultLarge.IsSuccess() {
		t.Error("Expected invocation to fail with large payload")
	}

	if resultLarge.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}
}

// TestFunctionalLambdaInvocationResultMapping tests result transformation
func TestFunctionalLambdaInvocationResultMapping(t *testing.T) {
	config := NewFunctionalLambdaConfig("mapping-function")
	payload := `{"input": "test"}`

	result := SimulateFunctionalLambdaInvocation(payload, config)

	// Test mapping successful result
	mappedResult := result.MapResult(func(r FunctionalInvokeResult) FunctionalInvokeResult {
		return FunctionalInvokeResult{
			StatusCode:      r.StatusCode,
			Payload:         "transformed: " + r.Payload,
			LogResult:       r.LogResult + " [processed]",
			ExecutedVersion: r.ExecutedVersion,
			FunctionError:   r.FunctionError,
			ExecutionTime:   r.ExecutionTime,
		}
	})

	if !mappedResult.IsSuccess() {
		t.Error("Expected mapped result to be successful")
	}

	if mappedResult.GetResult().IsAbsent() {
		t.Error("Expected mapped result to be present")
	}

	// Verify transformation
	transformedResult := mappedResult.GetResult().MustGet()
	if !contains(transformedResult.Payload, "transformed:") {
		t.Error("Expected payload to be transformed")
	}

	if !contains(transformedResult.LogResult, "[processed]") {
		t.Error("Expected log result to be processed")
	}
}

// TestFunctionalLambdaConfigurationOptions tests various configuration combinations
func TestFunctionalLambdaConfigurationOptions(t *testing.T) {
	config := NewFunctionalLambdaConfig("complex-function",
		WithInvocationType(types.InvocationTypeEvent),
		WithLogType(types.LogTypeTail),
		WithTimeout(60*time.Second),
		WithRetryConfig(5, 500*time.Millisecond),
		WithQualifier("LATEST"),
		WithLambdaMetadata("service", "integration-test"),
		WithLambdaMetadata("environment", "development"),
	)

	payload := `{"complex": "configuration", "test": true}`

	result := SimulateFunctionalLambdaInvocation(payload, config)

	if !result.IsSuccess() {
		t.Errorf("Expected complex configuration to work, got error: %v", result.GetError().MustGet())
	}

	// Verify that all metadata is preserved
	metadata := result.GetMetadata()
	if metadata["service"] != "integration-test" {
		t.Error("Expected service metadata to be preserved")
	}

	if metadata["environment"] != "development" {
		t.Error("Expected environment metadata to be preserved")
	}
}

// TestFunctionalLambdaValidation tests the validation utilities
func TestFunctionalLambdaValidation(t *testing.T) {
	validator := FunctionalLambdaValidation()

	// Test function name validation
	err := validator.ValidateFunctionName("")
	if err == nil {
		t.Error("Expected empty function name to be invalid")
	}

	err = validator.ValidateFunctionName("valid-function-name")
	if err != nil {
		t.Errorf("Expected valid function name to pass validation, got: %v", err)
	}

	longName := string(make([]byte, 141)) // Longer than 140 characters
	err = validator.ValidateFunctionName(longName)
	if err == nil {
		t.Error("Expected long function name to be invalid")
	}

	// Test payload validation
	err = validator.ValidatePayload("")
	if err == nil {
		t.Error("Expected empty payload to be invalid")
	}

	err = validator.ValidatePayload(`{"valid": "payload"}`)
	if err != nil {
		t.Errorf("Expected valid payload to pass validation, got: %v", err)
	}

	largePayload := string(make([]byte, 7*1024*1024)) // Larger than 6MB
	err = validator.ValidatePayload(largePayload)
	if err == nil {
		t.Error("Expected large payload to be invalid")
	}
}

// BenchmarkSimulateFunctionalLambdaInvocation benchmarks the simulation
func BenchmarkSimulateFunctionalLambdaInvocation(b *testing.B) {
	config := NewFunctionalLambdaConfig("benchmark-function",
		WithTimeout(5*time.Second),
		WithRetryConfig(1, 10*time.Millisecond),
	)

	payload := `{"benchmark": "test", "data": [1, 2, 3, 4, 5]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := SimulateFunctionalLambdaInvocation(payload, config)
		if !result.IsSuccess() {
			b.Errorf("Benchmark iteration failed: %v", result.GetError().MustGet())
		}
	}
}

// BenchmarkFunctionalLambdaConfig benchmarks configuration creation
func BenchmarkFunctionalLambdaConfig(b *testing.B) {
	options := []LambdaConfigOption{
		WithInvocationType(types.InvocationTypeEvent),
		WithLogType(types.LogTypeTail),
		WithTimeout(30 * time.Second),
		WithRetryConfig(3, 100*time.Millisecond),
		WithQualifier("v1"),
		WithLambdaMetadata("environment", "benchmark"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFunctionalLambdaConfig("benchmark-function", options...)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsInner(s[1:len(s)-1], substr))))
}

func containsInner(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}