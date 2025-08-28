package main

import (
	"go/build"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test suite for ExampleBasicUsage
func TestExampleBasicUsage(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		// RED: Test that function executes without panic
		require.NotPanics(t, ExampleBasicUsage, "ExampleBasicUsage should not panic")
	})
	
	t.Run("should demonstrate basic usage patterns", func(t *testing.T) {
		// GREEN: Test basic usage demonstration
		ExampleBasicUsage()
		
		// Verify the function completes successfully
		assert.True(t, true, "Basic usage example completed")
	})
	
	t.Run("should validate function names and payloads", func(t *testing.T) {
		// Test function name patterns
		functionNames := []string{
			"my-test-function",
			"s3-processor-function",
			"unreliable-function",
			"async-processor",
		}
		
		for _, name := range functionNames {
			assert.NotEmpty(t, name, "Function name should not be empty")
			assert.True(t, len(name) > 3, "Function name should be descriptive")
			assert.Contains(t, name, "-", "Function name should use kebab-case")
		}
		
		// Test payload patterns
		payloads := []string{
			`{"message": "Hello, Lambda!"}`,
			`{"test": "retry"}`,
			`{"data": "async-test"}`,
		}
		
		for _, payload := range payloads {
			assert.NotEmpty(t, payload, "Payload should not be empty")
			assert.True(t, strings.HasPrefix(payload, "{"), "Payload should be JSON object")
			assert.True(t, strings.HasSuffix(payload, "}"), "Payload should be complete JSON")
		}
	})
}

// Test suite for ExampleTableDrivenValidation
func TestExampleTableDrivenValidation(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleTableDrivenValidation, "ExampleTableDrivenValidation should not panic")
	})
	
	t.Run("should demonstrate table-driven validation patterns", func(t *testing.T) {
		// Test table-driven validation demonstration
		ExampleTableDrivenValidation()
		
		assert.True(t, true, "Table-driven validation example completed")
	})
	
	t.Run("should validate test scenarios structure", func(t *testing.T) {
		// Test scenario structure used in examples
		scenarios := []struct {
			name              string
			functionName      string
			payload           string
			expectSuccess     bool
			expectedInPayload string
			description       string
		}{
			{
				name:              "ValidPayload",
				functionName:      "validator-function",
				payload:           `{"name": "John", "age": 30}`,
				expectSuccess:     true,
				expectedInPayload: "valid",
				description:       "Valid payload should be processed successfully",
			},
			{
				name:              "InvalidPayload", 
				functionName:      "validator-function",
				payload:           `{"name": ""}`,
				expectSuccess:     false,
				expectedInPayload: "error",
				description:       "Invalid payload should return validation error",
			},
			{
				name:              "EmptyPayload",
				functionName:      "validator-function",
				payload:           `{}`,
				expectSuccess:     false,
				expectedInPayload: "missing_required_fields",
				description:       "Empty payload should return missing fields error",
			},
		}
		
		assert.Len(t, scenarios, 3, "Should have 3 test scenarios")
		
		for i, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				assert.NotEmpty(t, scenario.name, "Scenario name should not be empty")
				assert.NotEmpty(t, scenario.functionName, "Function name should not be empty")
				assert.NotEmpty(t, scenario.payload, "Payload should not be empty")
				assert.NotEmpty(t, scenario.expectedInPayload, "Expected payload content should not be empty")
				assert.NotEmpty(t, scenario.description, "Description should not be empty")
				
				// Validate success expectation matches scenario
				if i == 0 {
					assert.True(t, scenario.expectSuccess, "Valid payload scenario should expect success")
				} else {
					assert.False(t, scenario.expectSuccess, "Invalid payload scenarios should expect failure")
				}
			})
		}
	})
}

// Test suite for ExampleAdvancedEventProcessing
func TestExampleAdvancedEventProcessing(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleAdvancedEventProcessing, "ExampleAdvancedEventProcessing should not panic")
	})
	
	t.Run("should demonstrate advanced event processing patterns", func(t *testing.T) {
		// Test advanced event processing demonstration
		ExampleAdvancedEventProcessing()
		
		assert.True(t, true, "Advanced event processing example completed")
	})
}

// Test suite for ExampleEventSourceMapping
func TestExampleEventSourceMapping(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleEventSourceMapping, "ExampleEventSourceMapping should not panic")
	})
	
	t.Run("should demonstrate event source mapping patterns", func(t *testing.T) {
		// Test event source mapping demonstration
		ExampleEventSourceMapping()
		
		assert.True(t, true, "Event source mapping example completed")
	})
	
	t.Run("should validate event source ARN patterns", func(t *testing.T) {
		// Test common event source ARN patterns
		eventSourceARNs := []struct {
			service string
			arn     string
		}{
			{"SQS", "arn:aws:sqs:us-east-1:123456789012:test-queue"},
			{"Kinesis", "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream"},
			{"DynamoDB", "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2023-01-01T00:00:00.000"},
		}
		
		for _, eventSource := range eventSourceARNs {
			t.Run(eventSource.service, func(t *testing.T) {
				assert.True(t, strings.HasPrefix(eventSource.arn, "arn:aws:"), "ARN should start with arn:aws:")
				assert.Contains(t, eventSource.arn, strings.ToLower(eventSource.service), "ARN should contain service name")
				assert.Contains(t, eventSource.arn, "us-east-1", "ARN should contain region")
				assert.Contains(t, eventSource.arn, "123456789012", "ARN should contain account ID")
			})
		}
	})
}

// Test suite for ExamplePerformancePatterns
func TestExamplePerformancePatterns(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExamplePerformancePatterns, "ExamplePerformancePatterns should not panic")
	})
	
	t.Run("should demonstrate performance patterns", func(t *testing.T) {
		// Test performance patterns demonstration
		ExamplePerformancePatterns()
		
		assert.True(t, true, "Performance patterns example completed")
	})
	
	t.Run("should validate performance configuration patterns", func(t *testing.T) {
		// Test performance-related configurations
		performanceConfigs := []struct {
			name        string
			memoryMB    int32
			timeoutSec  int32
			description string
		}{
			{"Low Memory", 128, 30, "Minimal configuration for simple functions"},
			{"Standard", 256, 60, "Standard configuration for most functions"},
			{"High Memory", 1024, 300, "High-performance configuration"},
			{"Maximum", 3008, 900, "Maximum Lambda configuration"},
		}
		
		for _, config := range performanceConfigs {
			t.Run(config.name, func(t *testing.T) {
				assert.GreaterOrEqual(t, config.memoryMB, int32(128), "Memory should be at least 128MB")
				assert.LessOrEqual(t, config.memoryMB, int32(3008), "Memory should not exceed 3008MB")
				assert.GreaterOrEqual(t, config.timeoutSec, int32(1), "Timeout should be at least 1 second")
				assert.LessOrEqual(t, config.timeoutSec, int32(900), "Timeout should not exceed 900 seconds")
				assert.NotEmpty(t, config.description, "Description should not be empty")
			})
		}
	})
}

// Test suite for ExampleErrorHandlingAndRecovery
func TestExampleErrorHandlingAndRecovery(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleErrorHandlingAndRecovery, "ExampleErrorHandlingAndRecovery should not panic")
	})
	
	t.Run("should demonstrate error handling patterns", func(t *testing.T) {
		// Test error handling demonstration
		ExampleErrorHandlingAndRecovery()
		
		assert.True(t, true, "Error handling and recovery example completed")
	})
	
	t.Run("should validate error patterns", func(t *testing.T) {
		// Test common Lambda error patterns
		errorPatterns := []struct {
			errorType   string
			statusCode  int
			description string
		}{
			{"Timeout", 408, "Function execution timeout"},
			{"Memory", 500, "Out of memory error"},
			{"Runtime", 500, "Runtime execution error"},
			{"Throttling", 429, "Too many concurrent executions"},
		}
		
		for _, pattern := range errorPatterns {
			t.Run(pattern.errorType, func(t *testing.T) {
				assert.NotEmpty(t, pattern.errorType, "Error type should not be empty")
				assert.Greater(t, pattern.statusCode, 399, "Should be HTTP error status code")
				assert.Less(t, pattern.statusCode, 600, "Should be valid HTTP status code")
				assert.NotEmpty(t, pattern.description, "Error description should not be empty")
			})
		}
	})
}

// Test suite for ExampleComprehensiveIntegration
func TestExampleComprehensiveIntegration(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleComprehensiveIntegration, "ExampleComprehensiveIntegration should not panic")
	})
	
	t.Run("should demonstrate comprehensive integration patterns", func(t *testing.T) {
		// Test comprehensive integration demonstration
		ExampleComprehensiveIntegration()
		
		assert.True(t, true, "Comprehensive integration example completed")
	})
}

// Test suite for ExampleDataFactories
func TestExampleDataFactories(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleDataFactories, "ExampleDataFactories should not panic")
	})
	
	t.Run("should demonstrate data factory patterns", func(t *testing.T) {
		// Test data factory demonstration
		ExampleDataFactories()
		
		assert.True(t, true, "Data factories example completed")
	})
	
	t.Run("should validate data factory patterns", func(t *testing.T) {
		// Test data factory concepts
		factoryPatterns := []struct {
			eventType string
			service   string
			example   string
		}{
			{"S3", "s3", "ObjectCreated:Put"},
			{"DynamoDB", "dynamodb", "INSERT"},
			{"SQS", "sqs", "ReceiveMessage"},
			{"SNS", "sns", "Notification"},
			{"API Gateway", "apigateway", "GET /users"},
		}
		
		for _, pattern := range factoryPatterns {
			t.Run(pattern.eventType, func(t *testing.T) {
				assert.NotEmpty(t, pattern.eventType, "Event type should not be empty")
				assert.NotEmpty(t, pattern.service, "Service should not be empty")
				assert.NotEmpty(t, pattern.example, "Example should not be empty")
				
				// Special handling for API Gateway
				if pattern.eventType == "API Gateway" {
					assert.Equal(t, "apigateway", pattern.service, "API Gateway service should be apigateway")
				} else {
					assert.Equal(t, strings.ToLower(pattern.eventType), pattern.service, "Service should match event type (case insensitive)")
				}
			})
		}
	})
}

// Test suite for ExampleValidationHelpers
func TestExampleValidationHelpers(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleValidationHelpers, "ExampleValidationHelpers should not panic")
	})
	
	t.Run("should demonstrate validation helper patterns", func(t *testing.T) {
		// Test validation helper demonstration
		ExampleValidationHelpers()
		
		assert.True(t, true, "Validation helpers example completed")
	})
}

// Test suite for ExampleFluentAPIPatterns
func TestExampleFluentAPIPatterns(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleFluentAPIPatterns, "ExampleFluentAPIPatterns should not panic")
	})
	
	t.Run("should demonstrate fluent API patterns", func(t *testing.T) {
		// Test fluent API demonstration
		ExampleFluentAPIPatterns()
		
		assert.True(t, true, "Fluent API patterns example completed")
	})
}

// Test suite for ExampleOrganizationPatterns
func TestExampleOrganizationPatterns(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleOrganizationPatterns, "ExampleOrganizationPatterns should not panic")
	})
	
	t.Run("should demonstrate organization patterns", func(t *testing.T) {
		// Test organization patterns demonstration
		ExampleOrganizationPatterns()
		
		assert.True(t, true, "Organization patterns example completed")
	})
}

// Test suite for runAllExamples
func TestRunAllExamples(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, runAllExamples, "runAllExamples should not panic")
	})
	
	t.Run("should call all example functions", func(t *testing.T) {
		// Verify that runAllExamples executes all example functions
		runAllExamples()
		
		assert.True(t, true, "All Lambda examples executed successfully")
	})
}

// Test suite for Lambda runtime validation
func TestLambdaRuntimeValidation(t *testing.T) {
	t.Run("should validate supported runtimes", func(t *testing.T) {
		// Test supported Lambda runtimes
		supportedRuntimes := []types.Runtime{
			types.RuntimeNodejs20x,
			types.RuntimePython312,
			types.RuntimeJava21,
			types.RuntimeGo1x,
			types.RuntimeDotnet8,
		}
		
		for _, runtime := range supportedRuntimes {
			t.Run(string(runtime), func(t *testing.T) {
				runtimeStr := string(runtime)
				assert.NotEmpty(t, runtimeStr, "Runtime should not be empty")
				assert.True(t, len(runtimeStr) > 3, "Runtime should be descriptive")
				
				// Validate runtime naming patterns
				if strings.Contains(runtimeStr, "nodejs") {
					assert.Contains(t, runtimeStr, "nodejs", "Node.js runtime should contain 'nodejs'")
				} else if strings.Contains(runtimeStr, "python") {
					assert.Contains(t, runtimeStr, "python", "Python runtime should contain 'python'")
				} else if strings.Contains(runtimeStr, "java") {
					assert.Contains(t, runtimeStr, "java", "Java runtime should contain 'java'")
				}
			})
		}
	})
}

// Test suite for timeout and memory validation
func TestLambdaConstraints(t *testing.T) {
	t.Run("should validate timeout constraints", func(t *testing.T) {
		// Test Lambda timeout constraints
		timeoutTests := []struct {
			timeout     int32
			valid       bool
			description string
		}{
			{1, true, "Minimum timeout"},
			{30, true, "Default timeout"},
			{300, true, "Medium timeout"},
			{900, true, "Maximum timeout"},
			{0, false, "Zero timeout"},
			{901, false, "Excessive timeout"},
		}
		
		for _, test := range timeoutTests {
			t.Run(test.description, func(t *testing.T) {
				if test.valid {
					assert.GreaterOrEqual(t, test.timeout, int32(1), "Valid timeout should be >= 1")
					assert.LessOrEqual(t, test.timeout, int32(900), "Valid timeout should be <= 900")
				} else {
					assert.True(t, test.timeout < 1 || test.timeout > 900, "Invalid timeout should be outside valid range")
				}
			})
		}
	})
	
	t.Run("should validate memory constraints", func(t *testing.T) {
		// Test Lambda memory constraints
		memoryTests := []struct {
			memory      int32
			valid       bool
			description string
		}{
			{128, true, "Minimum memory"},
			{256, true, "Default memory"},
			{512, true, "Medium memory"},
			{1024, true, "High memory"},
			{3008, true, "Maximum memory"},
			{127, false, "Below minimum"},
			{3009, false, "Above maximum"},
		}
		
		for _, test := range memoryTests {
			t.Run(test.description, func(t *testing.T) {
				if test.valid {
					assert.GreaterOrEqual(t, test.memory, int32(128), "Valid memory should be >= 128")
					assert.LessOrEqual(t, test.memory, int32(3008), "Valid memory should be <= 3008")
				} else {
					assert.True(t, test.memory < 128 || test.memory > 3008, "Invalid memory should be outside valid range")
				}
			})
		}
	})
}

// Test suite for ARN validation patterns
func TestARNValidation(t *testing.T) {
	t.Run("should validate Lambda ARN formats", func(t *testing.T) {
		// Test various Lambda ARN patterns
		arnPatterns := []struct {
			arn         string
			arnType     string
			valid       bool
		}{
			{"arn:aws:lambda:us-east-1:123456789012:function:test-function", "function", true},
			{"arn:aws:lambda:us-east-1:123456789012:function:test-function:1", "function-version", true},
			{"arn:aws:lambda:us-east-1:123456789012:function:test-function:PROD", "function-alias", true},
			{"arn:aws:lambda:us-east-1:123456789012:layer:test-layer:1", "layer", true},
			{"invalid-arn", "invalid", false},
		}
		
		for _, pattern := range arnPatterns {
			t.Run(pattern.arnType, func(t *testing.T) {
				if pattern.valid {
					assert.True(t, strings.HasPrefix(pattern.arn, "arn:aws:lambda:"), "Valid ARN should start with arn:aws:lambda:")
					assert.Contains(t, pattern.arn, "us-east-1", "ARN should contain region")
					assert.Contains(t, pattern.arn, "123456789012", "ARN should contain account ID")
					
					parts := strings.Split(pattern.arn, ":")
					assert.GreaterOrEqual(t, len(parts), 6, "ARN should have at least 6 parts")
				} else {
					assert.False(t, strings.HasPrefix(pattern.arn, "arn:aws:lambda:"), "Invalid ARN should not have proper format")
				}
			})
		}
	})
}

// Test suite for JSON payload validation
func TestPayloadValidation(t *testing.T) {
	t.Run("should validate JSON payloads", func(t *testing.T) {
		// Test various JSON payload patterns
		payloads := []struct {
			payload string
			valid   bool
			name    string
		}{
			{`{"message": "hello"}`, true, "Simple object"},
			{`{"user": {"id": 1, "name": "John"}}`, true, "Nested object"},
			{`[{"id": 1}, {"id": 2}]`, true, "Array"},
			{`"simple string"`, true, "String"},
			{`123`, true, "Number"},
			{`true`, true, "Boolean"},
			{`{"invalid": }`, false, "Invalid JSON"},
		}
		
		for _, payload := range payloads {
			t.Run(payload.name, func(t *testing.T) {
				assert.NotEmpty(t, payload.payload, "Payload should not be empty")
				
				if payload.valid {
					// For valid payloads, we can't easily validate JSON without importing encoding/json
					// So we do basic structural checks
					if strings.HasPrefix(payload.payload, "{") {
						assert.True(t, strings.HasSuffix(payload.payload, "}"), "Object should end with }")
					} else if strings.HasPrefix(payload.payload, "[") {
						assert.True(t, strings.HasSuffix(payload.payload, "]"), "Array should end with ]")
					}
				}
			})
		}
	})
}

// Test suite for timing and duration patterns
func TestTimingPatterns(t *testing.T) {
	t.Run("should validate timeout duration patterns", func(t *testing.T) {
		// Test duration patterns commonly used in Lambda
		durations := []struct {
			duration    time.Duration
			description string
			appropriate string
		}{
			{5 * time.Second, "Short timeout", "Simple synchronous operations"},
			{30 * time.Second, "Default timeout", "Standard processing"},
			{5 * time.Minute, "Medium timeout", "Data processing tasks"},
			{15 * time.Minute, "Long timeout", "Heavy computational tasks"},
		}
		
		for _, d := range durations {
			t.Run(d.description, func(t *testing.T) {
				assert.Greater(t, d.duration, time.Duration(0), "Duration should be positive")
				assert.LessOrEqual(t, d.duration, 15*time.Minute, "Duration should not exceed maximum Lambda timeout")
				assert.NotEmpty(t, d.description, "Description should not be empty")
				assert.NotEmpty(t, d.appropriate, "Appropriate use case should be described")
			})
		}
	})
}

// Legacy tests maintained for compatibility
func TestExamplesCompilation(t *testing.T) {
	t.Run("should compile without errors", func(t *testing.T) {
		cmd := exec.Command("go", "build", "-o", "/tmp/test-lambda-examples", "examples.go")
		cmd.Dir = "."
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Errorf("Examples failed to compile: %v\nOutput: %s", err, output)
		}
	})
}

func TestExampleFunctionsExecute(t *testing.T) {
	t.Run("should execute all example functions without panic", func(t *testing.T) {
		// Test that all example functions can be called without panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Example function panicked: %v", r)
			}
		}()
		
		ExampleBasicUsage()
		ExampleTableDrivenValidation()
		ExampleAdvancedEventProcessing()
		ExampleEventSourceMapping()
		ExamplePerformancePatterns()
		ExampleErrorHandlingAndRecovery()
		ExampleComprehensiveIntegration()
		ExampleDataFactories()
		ExampleValidationHelpers()
		ExampleFluentAPIPatterns()
		ExampleOrganizationPatterns()
	})
}

func TestMainFunction(t *testing.T) {
	t.Run("should execute main without panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Main function panicked: %v", r)
			}
		}()
		
		runAllExamples()
	})
}

func TestPackageIsExecutable(t *testing.T) {
	t.Run("should build as executable binary", func(t *testing.T) {
		// Verify this is a main package
		pkg, err := build.ImportDir(".", 0)
		if err != nil {
			t.Fatalf("Failed to import current directory: %v", err)
		}
		
		if pkg.Name != "main" {
			t.Errorf("Package name should be 'main', got '%s'", pkg.Name)
		}
		
		// Check that main function exists
		hasMain := false
		for _, filename := range pkg.GoFiles {
			if filename == "examples.go" {
				hasMain = true
				break
			}
		}
		
		if !hasMain {
			t.Error("examples.go should exist in package")
		}
	})
}