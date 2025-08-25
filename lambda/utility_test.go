package lambda

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test sanitizeLogResult with various log entries
func TestSanitizeLogResult_WithVariousLogEntries_ShouldRemoveSensitiveInfo(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "log with RequestId should be filtered",
			input:    "Log message\nRequestId: 12345\nAnother message",
			expected: "Log message\nAnother message",
		},
		{
			name:     "log with Duration should be filtered",
			input:    "Log message\nDuration: 1000ms\nAnother message",
			expected: "Log message\nAnother message",
		},
		{
			name:     "log with Billed Duration should be filtered",
			input:    "Log message\nBilled Duration: 1000ms\nAnother message",
			expected: "Log message\nAnother message",
		},
		{
			name:     "log without sensitive info should remain unchanged",
			input:    "Log message\nAnother message\nThird message",
			expected: "Log message\nAnother message\nThird message",
		},
		{
			name:     "empty log should return empty",
			input:    "",
			expected: "",
		},
		{
			name:     "log with multiple sensitive entries should filter all",
			input:    "Message 1\nRequestId: 123\nMessage 2\nDuration: 100ms\nMessage 3\nBilled Duration: 150ms\nMessage 4",
			expected: "Message 1\nMessage 2\nMessage 3\nMessage 4",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: sanitizeLogResult is called
			result := sanitizeLogResult(tc.input)
			
			// Then: Should return expected sanitized result
			assert.Equal(t, tc.expected, result)
		})
	}
}

// RED: Test logOperation function
func TestLogOperation_WithValidParameters_ShouldNotPanic(t *testing.T) {
	// Given: Valid operation parameters
	operation := "test_operation"
	functionName := "test-function"
	details := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	
	// When: logOperation is called
	// Then: Should not panic
	assert.NotPanics(t, func() {
		logOperation(operation, functionName, details)
	})
}

// RED: Test logOperation with nil details
func TestLogOperation_WithNilDetails_ShouldNotPanic(t *testing.T) {
	// Given: Valid operation parameters with nil details
	operation := "test_operation"
	functionName := "test-function"
	
	// When: logOperation is called with nil details
	// Then: Should not panic
	assert.NotPanics(t, func() {
		logOperation(operation, functionName, nil)
	})
}

// RED: Test logOperation with empty details
func TestLogOperation_WithEmptyDetails_ShouldNotPanic(t *testing.T) {
	// Given: Valid operation parameters with empty details
	operation := "test_operation"
	functionName := "test-function"
	details := make(map[string]interface{})
	
	// When: logOperation is called with empty details
	// Then: Should not panic
	assert.NotPanics(t, func() {
		logOperation(operation, functionName, details)
	})
}

// RED: Test logResult function with success
func TestLogResult_WithSuccess_ShouldNotPanic(t *testing.T) {
	// Given: Valid result parameters for success
	operation := "test_operation"
	functionName := "test-function"
	success := true
	duration := 100 * time.Millisecond
	
	// When: logResult is called with success
	// Then: Should not panic
	assert.NotPanics(t, func() {
		logResult(operation, functionName, success, duration, nil)
	})
}

// RED: Test logResult function with failure
func TestLogResult_WithFailure_ShouldNotPanic(t *testing.T) {
	// Given: Valid result parameters for failure
	operation := "test_operation"
	functionName := "test-function"
	success := false
	duration := 50 * time.Millisecond
	err := assert.AnError
	
	// When: logResult is called with failure
	// Then: Should not panic
	assert.NotPanics(t, func() {
		logResult(operation, functionName, success, duration, err)
	})
}

// RED: Test createLambdaClient function
func TestCreateLambdaClient_WithValidContext_ShouldReturnClient(t *testing.T) {
	// Given: Valid test context with AWS config
	ctx := &TestContext{
		AwsConfig: aws.Config{Region: "us-east-1"},
	}
	
	// When: createLambdaClient is called
	client := createLambdaClient(ctx)
	
	// Then: Should return non-nil client
	assert.NotNil(t, client)
}

// RED: Test createCloudWatchLogsClient function
func TestCreateCloudWatchLogsClient_WithValidContext_ShouldReturnClient(t *testing.T) {
	// Given: Valid test context with AWS config
	ctx := &TestContext{
		AwsConfig: aws.Config{Region: "us-east-1"},
	}
	
	// When: createCloudWatchLogsClient is called
	client := createCloudWatchLogsClient(ctx)
	
	// Then: Should return non-nil client
	assert.NotNil(t, client)
}

// RED: Test defaultRetryConfig function
func TestDefaultRetryConfig_ShouldReturnValidConfig(t *testing.T) {
	// When: defaultRetryConfig is called
	config := defaultRetryConfig()
	
	// Then: Should return config with expected values
	assert.Equal(t, DefaultRetryAttempts, config.MaxAttempts)
	assert.Equal(t, DefaultRetryDelay, config.BaseDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

// RED: Test defaultInvokeOptions function
func TestDefaultInvokeOptions_ShouldReturnValidOptions(t *testing.T) {
	// When: defaultInvokeOptions is called
	options := defaultInvokeOptions()
	
	// Then: Should return options with expected defaults
	assert.Equal(t, InvocationTypeRequestResponse, options.InvocationType)
	assert.Equal(t, LogTypeNone, options.LogType)
	assert.Equal(t, DefaultTimeout, options.Timeout)
	assert.Equal(t, DefaultRetryAttempts, options.MaxRetries)
	assert.Equal(t, DefaultRetryDelay, options.RetryDelay)
	assert.NotNil(t, options.PayloadValidator)
}

// RED: Test default payload validator
func TestDefaultPayloadValidator_WithVariousPayloads_ShouldValidateCorrectly(t *testing.T) {
	// Given: Default invoke options with payload validator
	options := defaultInvokeOptions()
	validator := options.PayloadValidator
	require.NotNil(t, validator)
	
	testCases := []struct {
		name        string
		payload     string
		shouldError bool
	}{
		{
			name:        "empty payload should be valid",
			payload:     "",
			shouldError: false,
		},
		{
			name:        "small payload should be valid",
			payload:     `{"key": "value"}`,
			shouldError: false,
		},
		{
			name:        "normal size payload should be valid",
			payload:     createPayload(1024), // 1KB
			shouldError: false,
		},
		{
			name:        "maximum size payload should be valid",
			payload:     createPayload(MaxPayloadSize - 100), // Just under 6MB
			shouldError: false,
		},
		{
			name:        "oversized payload should be invalid",
			payload:     createPayload(MaxPayloadSize + 1000), // Just over 6MB by a safe margin
			shouldError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: payload validator is called
			err := validator(tc.payload)
			
			// Then: Should validate according to expectations
			if tc.shouldError {
				require.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), "payload size exceeds maximum")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create payloads of specific sizes
func createPayload(size int) string {
	if size <= 0 {
		return ""
	}
	
	// For small payloads, create simple JSON
	if size < 20 {
		return `{"test":"data"}`
	}
	
	// Create a JSON object with repeated content to reach the desired size
	contentSize := size - 20 // Leave room for JSON structure: {"data": "..."}
	if contentSize <= 0 {
		return `{"data": ""}`
	}
	
	content := make([]byte, contentSize)
	for i := range content {
		content[i] = 'a'
	}
	
	return `{"data": "` + string(content) + `"}`
}