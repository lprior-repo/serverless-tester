package lambda

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TDD RED: Write failing tests first for core utility functions

func TestDefaultConfigurations_TDD(t *testing.T) {
	t.Run("DefaultRetryConfig_ReturnsCorrectValues", func(t *testing.T) {
		config := defaultRetryConfig()
		
		assert.Equal(t, DefaultRetryAttempts, config.MaxAttempts)
		assert.Equal(t, DefaultRetryDelay, config.BaseDelay)
		assert.Equal(t, 30*time.Second, config.MaxDelay)
		assert.Equal(t, 2.0, config.Multiplier)
	})
	
	t.Run("DefaultInvokeOptions_ReturnsCorrectValues", func(t *testing.T) {
		options := defaultInvokeOptions()
		
		assert.Equal(t, InvocationTypeRequestResponse, options.InvocationType)
		assert.Equal(t, LogTypeNone, options.LogType)
		assert.Equal(t, DefaultTimeout, options.Timeout)
		assert.Equal(t, DefaultRetryAttempts, options.MaxRetries)
		assert.Equal(t, DefaultRetryDelay, options.RetryDelay)
		assert.NotNil(t, options.PayloadValidator)
		
		// Test default payload validator
		err := options.PayloadValidator("small payload")
		assert.NoError(t, err)
		
		// Test payload size limit
		largePayload := make([]byte, MaxPayloadSize+1)
		err = options.PayloadValidator(string(largePayload))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "payload size exceeds maximum")
	})
}

func TestBackoffCalculation_TDD(t *testing.T) {
	backoffScenarios := []struct {
		name           string
		attempt        int
		config         RetryConfig
		expectedDelay  time.Duration
		maxDelay       time.Duration
	}{
		{
			name:          "ZeroAttempt_ReturnsBaseDelay",
			attempt:       0,
			config:        RetryConfig{BaseDelay: 100 * time.Millisecond, Multiplier: 2.0, MaxDelay: 10 * time.Second},
			expectedDelay: 100 * time.Millisecond,
		},
		{
			name:          "FirstAttempt_ReturnsBaseDelayTimeMultiplier",
			attempt:       1,
			config:        RetryConfig{BaseDelay: 100 * time.Millisecond, Multiplier: 2.0, MaxDelay: 10 * time.Second},
			expectedDelay: 200 * time.Millisecond,
		},
		{
			name:          "SecondAttempt_ReturnsExponentialDelay",
			attempt:       2,
			config:        RetryConfig{BaseDelay: 100 * time.Millisecond, Multiplier: 2.0, MaxDelay: 10 * time.Second},
			expectedDelay: 400 * time.Millisecond,
		},
		{
			name:          "HighAttempt_CapsAtMaxDelay",
			attempt:       10,
			config:        RetryConfig{BaseDelay: 1 * time.Second, Multiplier: 2.0, MaxDelay: 5 * time.Second},
			expectedDelay: 5 * time.Second,
		},
		{
			name:          "NegativeAttempt_ReturnsBaseDelay",
			attempt:       -1,
			config:        RetryConfig{BaseDelay: 500 * time.Millisecond, Multiplier: 1.5, MaxDelay: 5 * time.Second},
			expectedDelay: 500 * time.Millisecond,
		},
		{
			name:          "OverflowPrevention_CapsAtMaxDelay",
			attempt:       20, // This would cause overflow without protection
			config:        RetryConfig{BaseDelay: 1 * time.Second, Multiplier: 2.0, MaxDelay: 30 * time.Second},
			expectedDelay: 30 * time.Second,
		},
	}
	
	for _, scenario := range backoffScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := calculateBackoffDelay(scenario.attempt, scenario.config)
			
			// For capped scenarios, check exact match
			if scenario.attempt >= 5 {
				assert.Equal(t, scenario.expectedDelay, result,
					"High attempt should be capped at MaxDelay")
			} else {
				assert.Equal(t, scenario.expectedDelay, result,
					"Attempt %d should produce delay %v, got %v", scenario.attempt, scenario.expectedDelay, result)
			}
			
			// Ensure result never exceeds MaxDelay
			assert.LessOrEqual(t, result, scenario.config.MaxDelay,
				"Result should never exceed MaxDelay")
			
			// Ensure result is never negative
			assert.GreaterOrEqual(t, result, time.Duration(0),
				"Result should never be negative")
		})
	}
}

func TestLogOperationsAndResults_TDD(t *testing.T) {
	t.Run("LogOperation_WithDetails", func(t *testing.T) {
		// This test verifies the function doesn't panic and handles parameters correctly
		details := map[string]interface{}{
			"payload_size": 100,
			"retry_count":  3,
		}
		
		// Should not panic
		assert.NotPanics(t, func() {
			logOperation("test_operation", "test-function", details)
		})
	})
	
	t.Run("LogOperation_WithEmptyDetails", func(t *testing.T) {
		// Should handle empty details map gracefully
		assert.NotPanics(t, func() {
			logOperation("test_operation", "test-function", map[string]interface{}{})
		})
	})
	
	t.Run("LogOperation_WithNilDetails", func(t *testing.T) {
		// Should handle nil details map gracefully
		assert.NotPanics(t, func() {
			logOperation("test_operation", "test-function", nil)
		})
	})
	
	t.Run("LogResult_Success", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logResult("test_operation", "test-function", true, 100*time.Millisecond, nil)
		})
	})
	
	t.Run("LogResult_WithError", func(t *testing.T) {
		testErr := fmt.Errorf("test error")
		assert.NotPanics(t, func() {
			logResult("test_operation", "test-function", false, 200*time.Millisecond, testErr)
		})
	})
}

func TestSanitizeLogResult_TDD_ComprehensiveScenarios(t *testing.T) {
	sanitizeScenarios := []struct {
		name     string
		input    string
		expected []string // Lines that should be present
		excluded []string // Lines that should be excluded
	}{
		{
			name: "StandardLambdaLogs_FiltersPlatformData",
			input: `2023-01-01T12:00:00.000Z	INFO	Application started
RequestId: abc-123-def-456
2023-01-01T12:00:01.000Z	INFO	Processing request
Duration: 1500.00 ms	Billed Duration: 1500 ms	Memory Size: 128 MB	Max Memory Used: 64 MB
2023-01-01T12:00:01.500Z	INFO	Request completed successfully`,
			expected: []string{
				"2023-01-01T12:00:00.000Z	INFO	Application started",
				"2023-01-01T12:00:01.000Z	INFO	Processing request",
				"2023-01-01T12:00:01.500Z	INFO	Request completed successfully",
			},
			excluded: []string{
				"RequestId:",
				"Duration:",
				"Billed Duration:",
			},
		},
		{
			name: "OnlyPlatformLogs_ReturnsEmpty",
			input: `RequestId: abc-123-def-456
Duration: 1000.00 ms
Billed Duration: 1000 ms`,
			expected: []string{},
			excluded: []string{
				"RequestId:",
				"Duration:",
				"Billed Duration:",
			},
		},
		{
			name: "EmptyInput_ReturnsEmpty",
			input:    "",
			expected: []string{},
			excluded: []string{},
		},
		{
			name: "OnlyApplicationLogs_PreservesAll",
			input: `INFO: Starting application
ERROR: Something went wrong
DEBUG: Processing complete`,
			expected: []string{
				"INFO: Starting application",
				"ERROR: Something went wrong",
				"DEBUG: Processing complete",
			},
			excluded: []string{},
		},
		{
			name: "MixedLogTypes_FiltersCorrectly",
			input: `START RequestId: 123
INFO: Application log 1
RequestId: abc-123
INFO: Application log 2
Duration: 500.00 ms
INFO: Application log 3
Billed Duration: 500 ms
END RequestId: 123`,
			expected: []string{
				"START RequestId: 123", // This doesn't match our filter patterns
				"INFO: Application log 1",
				"INFO: Application log 2",
				"INFO: Application log 3",
				"END RequestId: 123", // This doesn't match our filter patterns
			},
			excluded: []string{
				"RequestId: abc-123", // Only the exact RequestId: pattern is filtered
			},
		},
	}
	
	for _, scenario := range sanitizeScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := sanitizeLogResult(scenario.input)
			
			// Check that expected lines are present
			for _, expectedLine := range scenario.expected {
				assert.Contains(t, result, expectedLine,
					"Sanitized result should contain: %q", expectedLine)
			}
			
			// Check that excluded patterns are not present
			for _, excludedPattern := range scenario.excluded {
				assert.NotContains(t, result, excludedPattern,
					"Sanitized result should not contain: %q", excludedPattern)
			}
		})
	}
}

func TestInvokeOptionsValidation_TDD(t *testing.T) {
	t.Run("MergeInvokeOptions_PartialOverride", func(t *testing.T) {
		partialOptions := &InvokeOptions{
			InvocationType: types.InvocationTypeEvent,
			Timeout:        60 * time.Second,
			// Other fields left as zero values
		}
		
		merged := mergeInvokeOptions(partialOptions)
		
		// Explicitly set values should be preserved
		assert.Equal(t, types.InvocationTypeEvent, merged.InvocationType)
		assert.Equal(t, 60*time.Second, merged.Timeout)
		
		// Zero values should use defaults
		assert.Equal(t, LogTypeNone, merged.LogType) // Should use default
		assert.Equal(t, DefaultRetryAttempts, merged.MaxRetries) // Should use default
		assert.Equal(t, DefaultRetryDelay, merged.RetryDelay) // Should use default
		assert.NotNil(t, merged.PayloadValidator) // Should use default
	})
	
	t.Run("MergeInvokeOptions_CompleteOverride", func(t *testing.T) {
		customValidator := func(payload string) error {
			if len(payload) > 1000 {
				return fmt.Errorf("custom validator: payload too large")
			}
			return nil
		}
		
		customOptions := &InvokeOptions{
			InvocationType:   types.InvocationTypeDryRun,
			LogType:          LogTypeTail,
			ClientContext:    "custom-context",
			Qualifier:        "LATEST",
			Timeout:          45 * time.Second,
			MaxRetries:       5,
			RetryDelay:       2 * time.Second,
			PayloadValidator: customValidator,
		}
		
		merged := mergeInvokeOptions(customOptions)
		
		assert.Equal(t, types.InvocationTypeDryRun, merged.InvocationType)
		assert.Equal(t, LogTypeTail, merged.LogType)
		assert.Equal(t, "custom-context", merged.ClientContext)
		assert.Equal(t, "LATEST", merged.Qualifier)
		assert.Equal(t, 45*time.Second, merged.Timeout)
		assert.Equal(t, 5, merged.MaxRetries)
		assert.Equal(t, 2*time.Second, merged.RetryDelay)
		
		// Test custom validator was preserved
		err := merged.PayloadValidator(string(make([]byte, 2000)))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom validator")
	})
}

func TestConstants_TDD_VerifyValues(t *testing.T) {
	t.Run("Constants_HaveExpectedValues", func(t *testing.T) {
		// Verify Lambda configuration constants
		assert.Equal(t, 30*time.Second, DefaultTimeout)
		assert.Equal(t, 128, DefaultMemorySize)
		assert.Equal(t, 6*1024*1024, MaxPayloadSize) // 6MB
		assert.Equal(t, 3, DefaultRetryAttempts)
		assert.Equal(t, 1*time.Second, DefaultRetryDelay)
		
		// Verify invocation type constants
		assert.Equal(t, types.InvocationTypeRequestResponse, InvocationTypeRequestResponse)
		assert.Equal(t, types.InvocationTypeEvent, InvocationTypeEvent)
		assert.Equal(t, types.InvocationTypeDryRun, InvocationTypeDryRun)
		
		// Verify log type constants
		assert.Equal(t, types.LogTypeNone, LogTypeNone)
		assert.Equal(t, types.LogTypeTail, LogTypeTail)
	})
}

func TestErrorConstants_TDD_VerifyMessages(t *testing.T) {
	errorScenarios := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrFunctionNotFound",
			err:      ErrFunctionNotFound,
			expected: "lambda function not found",
		},
		{
			name:     "ErrInvalidPayload",
			err:      ErrInvalidPayload,
			expected: "invalid payload format",
		},
		{
			name:     "ErrInvocationTimeout",
			err:      ErrInvocationTimeout,
			expected: "lambda invocation timeout",
		},
		{
			name:     "ErrInvocationFailed",
			err:      ErrInvocationFailed,
			expected: "lambda invocation failed",
		},
		{
			name:     "ErrLogRetrievalFailed",
			err:      ErrLogRetrievalFailed,
			expected: "failed to retrieve lambda logs",
		},
		{
			name:     "ErrInvalidFunctionName",
			err:      ErrInvalidFunctionName,
			expected: "invalid function name",
		},
	}
	
	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			assert.Equal(t, scenario.expected, scenario.err.Error())
		})
	}
}

func TestStructInitialization_TDD(t *testing.T) {
	t.Run("TestContext_CanBeInitialized", func(t *testing.T) {
		ctx := &TestContext{
			T:      &MockTestingT{},
			Region: "us-west-2",
		}
		
		assert.NotNil(t, ctx.T)
		assert.Equal(t, "us-west-2", ctx.Region)
	})
	
	t.Run("InvokeResult_CanBeInitialized", func(t *testing.T) {
		result := &InvokeResult{
			StatusCode:      200,
			Payload:         `{"success": true}`,
			LogResult:       "LOG INFO: Success",
			ExecutedVersion: "$LATEST",
			FunctionError:   "",
			ExecutionTime:   100 * time.Millisecond,
		}
		
		assert.Equal(t, int32(200), result.StatusCode)
		assert.Contains(t, result.Payload, "success")
		assert.Equal(t, "$LATEST", result.ExecutedVersion)
		assert.Equal(t, 100*time.Millisecond, result.ExecutionTime)
	})
	
	t.Run("FunctionConfiguration_CanBeInitialized", func(t *testing.T) {
		config := &FunctionConfiguration{
			FunctionName:    "test-function",
			FunctionArn:     "arn:aws:lambda:us-east-1:123456789012:function:test-function",
			Runtime:         types.RuntimeNodejs18x,
			Handler:         "index.handler",
			Description:     "Test function",
			Timeout:         30,
			MemorySize:      128,
			LastModified:    "2023-01-01T12:00:00.000Z",
			Environment:     map[string]string{"ENV": "test"},
			Role:            "arn:aws:iam::123456789012:role/lambda-role",
			State:           types.StateActive,
			StateReason:     "",
			Version:         "$LATEST",
			CodeSha256:      "abc123",
			PackageType:     types.PackageTypeZip,
		}
		
		assert.Equal(t, "test-function", config.FunctionName)
		assert.Equal(t, types.RuntimeNodejs18x, config.Runtime)
		assert.Equal(t, int32(128), config.MemorySize)
		assert.Equal(t, "test", config.Environment["ENV"])
	})
}

func TestPayloadValidation_TDD_EdgeCases(t *testing.T) {
	t.Run("DefaultPayloadValidator_EdgeCases", func(t *testing.T) {
		validator := defaultInvokeOptions().PayloadValidator
		
		// Test boundary conditions
		exactMaxSize := make([]byte, MaxPayloadSize)
		for i := range exactMaxSize {
			exactMaxSize[i] = 'a'
		}
		
		// Should pass at exact limit
		err := validator(string(exactMaxSize))
		assert.NoError(t, err)
		
		// Should fail at limit + 1
		overMaxSize := make([]byte, MaxPayloadSize+1)
		for i := range overMaxSize {
			overMaxSize[i] = 'a'
		}
		
		err = validator(string(overMaxSize))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "payload size exceeds maximum")
		assert.Contains(t, err.Error(), fmt.Sprintf("%d bytes", MaxPayloadSize+1))
	})
}

// Mock TestingT implementation for struct initialization tests
type MockTestingT struct{}

func (m *MockTestingT) Errorf(format string, args ...interface{}) {}
func (m *MockTestingT) Error(args ...interface{})                 {}
func (m *MockTestingT) Fail()                                     {}
func (m *MockTestingT) FailNow()                                  {}
func (m *MockTestingT) Helper()                                   {}
func (m *MockTestingT) Fatal(args ...interface{})                {}
func (m *MockTestingT) Fatalf(format string, args ...interface{}) {}
func (m *MockTestingT) Name() string                              { return "MockTest" }