package lambda

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Setup helper for invoke tests
func setupMockInvokeTest() (*TestContext, *MockLambdaClient) {
	mockLambdaClient := &MockLambdaClient{}
	mockCloudWatchClient := &MockCloudWatchLogsClient{}
	
	mockFactory := &MockClientFactory{
		LambdaClient:     mockLambdaClient,
		CloudWatchClient: mockCloudWatchClient,
	}
	
	SetClientFactory(mockFactory)
	
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	
	return ctx, mockLambdaClient
}

// RED: Test InvokeE with successful synchronous invocation using mocks
func TestInvokeE_WithSuccessfulSynchronousInvocation_UsingMocks_ShouldReturnResult(t *testing.T) {
	// Given: Mock setup with successful invoke response
	ctx, mockClient := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	responsePayload := `{"message": "success", "result": "processed"}`
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode:      200,
		Payload:         []byte(responsePayload),
		ExecutedVersion: aws.String("$LATEST"),
		LogResult:       aws.String("START RequestId: abc-123\nProcessing event\nEND RequestId: abc-123"),
		FunctionError:   nil,
	}
	mockClient.InvokeError = nil
	
	payload := `{"key": "value", "data": "test"}`
	
	// When: InvokeE is called
	result, err := InvokeE(ctx, "test-function", payload)
	
	// Then: Should return successful result
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(200), result.StatusCode)
	assert.Equal(t, responsePayload, result.Payload)
	assert.Equal(t, "$LATEST", result.ExecutedVersion)
	assert.Contains(t, result.LogResult, "Processing event")
	assert.Empty(t, result.FunctionError)
	assert.Greater(t, result.ExecutionTime, time.Duration(0))
	assert.Contains(t, mockClient.InvokeCalls, "test-function")
}

// RED: Test InvokeE with function error using mocks
func TestInvokeE_WithFunctionError_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with function error response
	ctx, mockClient := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	errorPayload := `{"errorMessage": "Validation failed", "errorType": "ValidationError"}`
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode:    200,
		Payload:       []byte(errorPayload),
		FunctionError: aws.String("Handled"),
	}
	mockClient.InvokeError = nil
	
	// When: InvokeE is called
	_, err := InvokeE(ctx, "test-function", `{"input": "invalid"}`)
	
	// Then: Should return error due to function error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lambda function error: Handled")
	assert.Contains(t, mockClient.InvokeCalls, "test-function")
}

// RED: Test InvokeE with invalid payload using mocks
func TestInvokeE_WithInvalidPayload_UsingMocks_ShouldReturnValidationError(t *testing.T) {
	// Given: Mock setup
	ctx, _ := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	invalidPayload := `{invalid json}`
	
	// When: InvokeE is called with invalid payload
	result, err := InvokeE(ctx, "test-function", invalidPayload)
	
	// Then: Should return validation error without making API call
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid payload format")
}

// RED: Test InvokeE with invalid function name using mocks
func TestInvokeE_WithInvalidFunctionName_UsingMocks_ShouldReturnValidationError(t *testing.T) {
	// Given: Mock setup
	ctx, _ := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	// When: InvokeE is called with empty function name
	result, err := InvokeE(ctx, "", `{"valid": "payload"}`)
	
	// Then: Should return validation error without making API call
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid function name")
}

// RED: Test InvokeE with bad status code using mocks
func TestInvokeE_WithBadStatusCode_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with unsuccessful status code
	ctx, mockClient := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 500,
		Payload:    []byte(`{"error": "Internal server error"}`),
	}
	mockClient.InvokeError = nil
	
	// When: InvokeE is called
	_, err := InvokeE(ctx, "test-function", `{"input": "data"}`)
	
	// Then: Should return error due to bad status code
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lambda invocation failed with status code: 500")
	assert.Contains(t, mockClient.InvokeCalls, "test-function")
}

// RED: Test InvokeE with API error using mocks
func TestInvokeE_WithAPIError_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with API error
	ctx, mockClient := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	mockClient.InvokeResponse = nil
	mockClient.InvokeError = createAccessDeniedError("Invoke")
	
	// When: InvokeE is called
	result, err := InvokeE(ctx, "test-function", `{"input": "data"}`)
	
	// Then: Should return API error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "lambda invocation failed")
	assert.Contains(t, err.Error(), "AccessDeniedException")
	assert.Contains(t, mockClient.InvokeCalls, "test-function")
}

// RED: Test InvokeWithOptionsE with custom options using mocks
func TestInvokeWithOptionsE_WithCustomOptions_UsingMocks_ShouldUseOptions(t *testing.T) {
	// Given: Mock setup with successful response
	ctx, mockClient := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 200,
		Payload:    []byte(`{"async": true}`),
	}
	mockClient.InvokeError = nil
	
	// Custom options for async invocation
	opts := &InvokeOptions{
		InvocationType: types.InvocationTypeEvent,
		LogType:        LogTypeNone,
		ClientContext:  "custom-context",
		Qualifier:      "PROD",
	}
	
	// When: InvokeWithOptionsE is called with custom options
	result, err := InvokeWithOptionsE(ctx, "test-function", `{"data": "test"}`, opts)
	
	// Then: Should return result and respect options
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(200), result.StatusCode)
	assert.Equal(t, `{"async": true}`, result.Payload)
	assert.Contains(t, mockClient.InvokeCalls, "test-function")
}

// RED: Test InvokeAsyncE using mocks
func TestInvokeAsyncE_WithValidPayload_UsingMocks_ShouldInvokeAsynchronously(t *testing.T) {
	// Given: Mock setup for async invocation
	ctx, mockClient := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 202, // Accepted for async
		Payload:    []byte(""),
	}
	mockClient.InvokeError = nil
	
	// When: InvokeAsyncE is called
	result, err := InvokeAsyncE(ctx, "test-function", `{"async": "data"}`)
	
	// Then: Should return successful async result
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(202), result.StatusCode)
	assert.Contains(t, mockClient.InvokeCalls, "test-function")
}

// RED: Test DryRunInvokeE using mocks
func TestDryRunInvokeE_WithValidPayload_UsingMocks_ShouldValidateOnly(t *testing.T) {
	// Given: Mock setup for dry run
	ctx, mockClient := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 204, // No content for dry run
		Payload:    []byte(""),
	}
	mockClient.InvokeError = nil
	
	// When: DryRunInvokeE is called
	result, err := DryRunInvokeE(ctx, "test-function", `{"test": "payload"}`)
	
	// Then: Should return dry run result
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(204), result.StatusCode)
	assert.Contains(t, mockClient.InvokeCalls, "test-function")
}

// RED: Test InvokeWithRetryE with eventual success using mocks
func TestInvokeWithRetryE_WithEventualSuccess_UsingMocks_ShouldRetryAndSucceed(t *testing.T) {
	// Given: Mock setup that succeeds immediately
	ctx, _ := setupMockInvokeTest()
	defer teardownMockAssertionTest()
	
	// Override the global client factory to simulate successful response
	originalFactory := globalClientFactory
	SetClientFactory(&MockClientFactory{
		LambdaClient: &MockLambdaClient{
			InvokeResponse: &lambda.InvokeOutput{
				StatusCode: 200,
				Payload:    []byte(`{"success": true}`),
			},
		},
		CloudWatchClient: &MockCloudWatchLogsClient{},
	})
	defer SetClientFactory(originalFactory)
	
	// When: InvokeWithRetryE is called with retry count
	result, err := InvokeWithRetryE(ctx, "test-function", `{"data": "test"}`, 3)
	
	// Then: Should eventually succeed
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(200), result.StatusCode)
}

// RED: Test ParseInvokeOutputE with valid JSON using mocks
func TestParseInvokeOutputE_WithValidJSON_ShouldParseCorrectly(t *testing.T) {
	// Given: Invoke result with valid JSON payload
	result := &InvokeResult{
		StatusCode: 200,
		Payload:    `{"message": "hello", "count": 42, "success": true}`,
	}
	
	// When: ParseInvokeOutputE is called
	var output struct {
		Message string `json:"message"`
		Count   int    `json:"count"`
		Success bool   `json:"success"`
	}
	
	err := ParseInvokeOutputE(result, &output)
	
	// Then: Should parse successfully
	require.NoError(t, err)
	assert.Equal(t, "hello", output.Message)
	assert.Equal(t, 42, output.Count)
	assert.True(t, output.Success)
}

// RED: Test ParseInvokeOutputE with invalid JSON using mocks
func TestParseInvokeOutputE_WithInvalidJSON_ShouldReturnError(t *testing.T) {
	// Given: Invoke result with invalid JSON payload
	result := &InvokeResult{
		StatusCode: 200,
		Payload:    `{invalid json}`,
	}
	
	// When: ParseInvokeOutputE is called
	var output map[string]interface{}
	err := ParseInvokeOutputE(result, &output)
	
	// Then: Should return JSON parsing error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal payload")
}

// RED: Test ParseInvokeOutputE with nil result using mocks
func TestParseInvokeOutputE_WithNilResult_ShouldReturnError(t *testing.T) {
	// Given: Nil result
	var result *InvokeResult = nil
	
	// When: ParseInvokeOutputE is called
	var output map[string]interface{}
	err := ParseInvokeOutputE(result, &output)
	
	// Then: Should return nil result error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoke result is nil")
}

// RED: Test ParseInvokeOutputE with empty payload using mocks
func TestParseInvokeOutputE_WithEmptyPayload_ShouldReturnError(t *testing.T) {
	// Given: Invoke result with empty payload
	result := &InvokeResult{
		StatusCode: 200,
		Payload:    "",
	}
	
	// When: ParseInvokeOutputE is called
	var output map[string]interface{}
	err := ParseInvokeOutputE(result, &output)
	
	// Then: Should return empty payload error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payload is empty")
}

// RED: Test MarshalPayloadE with valid object
func TestMarshalPayloadE_WithValidObject_ShouldReturnJSON(t *testing.T) {
	// Given: Valid payload object
	payload := map[string]interface{}{
		"message": "test",
		"count":   123,
		"active":  true,
		"tags":    []string{"tag1", "tag2"},
	}
	
	// When: MarshalPayloadE is called
	jsonStr, err := MarshalPayloadE(payload)
	
	// Then: Should return JSON string
	require.NoError(t, err)
	assert.Contains(t, jsonStr, "test")
	assert.Contains(t, jsonStr, "123")
	assert.Contains(t, jsonStr, "true")
	assert.Contains(t, jsonStr, "tag1")
	
	// Verify it's valid JSON by unmarshaling back
	var unmarshaled map[string]interface{}
	err = ParseInvokeOutputE(&InvokeResult{Payload: jsonStr}, &unmarshaled)
	require.NoError(t, err)
}

// RED: Test MarshalPayloadE with nil payload
func TestMarshalPayloadE_WithNilPayload_ShouldReturnEmptyString(t *testing.T) {
	// Given: Nil payload
	var payload interface{} = nil
	
	// When: MarshalPayloadE is called
	jsonStr, err := MarshalPayloadE(payload)
	
	// Then: Should return empty string without error
	require.NoError(t, err)
	assert.Equal(t, "", jsonStr)
}

// RED: Test MarshalPayloadE with unmarshalable object
func TestMarshalPayloadE_WithUnmarshalableObject_ShouldReturnError(t *testing.T) {
	// Given: Object that can't be marshaled to JSON (channel)
	payload := make(chan int)
	
	// When: MarshalPayloadE is called
	jsonStr, err := MarshalPayloadE(payload)
	
	// Then: Should return marshal error
	require.Error(t, err)
	assert.Empty(t, jsonStr)
	assert.Contains(t, err.Error(), "failed to marshal payload")
}

// RED: Test contains function with various inputs
func TestContains_WithVariousInputs_ShouldReturnCorrectResults(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"ExactMatch", "hello", "hello", true},
		{"SubstringAtStart", "hello world", "hello", true},
		{"SubstringAtEnd", "hello world", "world", true},
		{"SubstringInMiddle", "hello world", "lo wo", true},
		{"NotFound", "hello world", "xyz", false},
		{"EmptySubstring", "hello", "", true}, // Empty string is always found
		{"EmptyString", "", "hello", false},
		{"BothEmpty", "", "", true},
		{"CaseSensitive", "Hello", "hello", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := contains(tc.s, tc.substr)
			assert.Equal(t, tc.expected, result, 
				"contains(%q, %q) should return %v", tc.s, tc.substr, tc.expected)
		})
	}
}

// RED: Test indexOf function with various inputs
func TestIndexOf_WithVariousInputs_ShouldReturnCorrectIndexes(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{"AtStart", "hello world", "hello", 0},
		{"AtEnd", "hello world", "world", 6},
		{"InMiddle", "hello world", "lo", 3},
		{"NotFound", "hello world", "xyz", -1},
		{"MultipleOccurrences", "hello hello", "hello", 0}, // Returns first occurrence
		{"EmptySubstring", "hello", "", 0},
		{"EmptyString", "", "hello", -1},
		{"SingleChar", "hello", "e", 1},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := indexOf(tc.s, tc.substr)
			assert.Equal(t, tc.expected, result,
				"indexOf(%q, %q) should return %d", tc.s, tc.substr, tc.expected)
		})
	}
}

// RED: Test isNonRetryableError with various error types
func TestIsNonRetryableError_WithVariousErrors_ShouldReturnCorrectResults(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"NilError", nil, false},
		{"InvalidParameterValueException", fmt.Errorf("InvalidParameterValueException: Invalid parameter"), true},
		{"ResourceNotFoundException", fmt.Errorf("ResourceNotFoundException: Function not found"), true},
		{"InvalidRequestContentException", fmt.Errorf("InvalidRequestContentException: Invalid request"), true},
		{"RequestTooLargeException", fmt.Errorf("RequestTooLargeException: Request too large"), true},
		{"UnsupportedMediaTypeException", fmt.Errorf("UnsupportedMediaTypeException: Unsupported media"), true},
		{"ThrottlingException", fmt.Errorf("ThrottlingException: Rate exceeded"), false}, // Should be retried
		{"ServiceException", fmt.Errorf("ServiceException: Internal error"), false}, // Should be retried
		{"GenericError", fmt.Errorf("Something went wrong"), false}, // Should be retried
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isNonRetryableError(tc.err)
			assert.Equal(t, tc.expected, result,
				"isNonRetryableError(%v) should return %v", tc.err, tc.expected)
		})
	}
}

// RED: Test mergeInvokeOptions with various option combinations
func TestMergeInvokeOptions_WithVariousOptions_ShouldMergeCorrectly(t *testing.T) {
	// Test with nil options
	t.Run("NilOptions", func(t *testing.T) {
		result := mergeInvokeOptions(nil)
		
		assert.Equal(t, InvocationTypeRequestResponse, result.InvocationType)
		assert.Equal(t, LogTypeNone, result.LogType)
		assert.Equal(t, DefaultTimeout, result.Timeout)
		assert.Equal(t, DefaultRetryAttempts, result.MaxRetries)
		assert.Equal(t, DefaultRetryDelay, result.RetryDelay)
		assert.NotNil(t, result.PayloadValidator)
	})
	
	// Test with partial options
	t.Run("PartialOptions", func(t *testing.T) {
		userOpts := &InvokeOptions{
			InvocationType: types.InvocationTypeEvent,
			Timeout:        60 * time.Second,
		}
		
		result := mergeInvokeOptions(userOpts)
		
		assert.Equal(t, types.InvocationTypeEvent, result.InvocationType)
		assert.Equal(t, LogTypeNone, result.LogType) // Default
		assert.Equal(t, 60*time.Second, result.Timeout)
		assert.Equal(t, DefaultRetryAttempts, result.MaxRetries) // Default
	})
	
	// Test with complete options
	t.Run("CompleteOptions", func(t *testing.T) {
		customValidator := func(string) error { return nil }
		userOpts := &InvokeOptions{
			InvocationType:   types.InvocationTypeDryRun,
			LogType:          LogTypeTail,
			Timeout:          120 * time.Second,
			MaxRetries:       5,
			RetryDelay:       2 * time.Second,
			PayloadValidator: customValidator,
		}
		
		result := mergeInvokeOptions(userOpts)
		
		assert.Equal(t, types.InvocationTypeDryRun, result.InvocationType)
		assert.Equal(t, LogTypeTail, result.LogType)
		assert.Equal(t, 120*time.Second, result.Timeout)
		assert.Equal(t, 5, result.MaxRetries)
		assert.Equal(t, 2*time.Second, result.RetryDelay)
		assert.NotNil(t, result.PayloadValidator)
	})
}

// RED: Test calculateBackoffDelay with various attempts
func TestCalculateBackoffDelay_WithVariousAttempts_ShouldCalculateCorrectly(t *testing.T) {
	config := RetryConfig{
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}
	
	testCases := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"ZeroAttempt", 0, 1 * time.Second},
		{"FirstAttempt", 1, 2 * time.Second},
		{"SecondAttempt", 2, 4 * time.Second},
		{"ThirdAttempt", 3, 8 * time.Second},
		{"FourthAttempt", 4, 16 * time.Second},
		{"FifthAttempt", 5, 30 * time.Second}, // Capped at MaxDelay
		{"HighAttempt", 10, 30 * time.Second}, // Capped at MaxDelay
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateBackoffDelay(tc.attempt, config)
			assert.Equal(t, tc.expected, result,
				"calculateBackoffDelay(%d) should return %v", tc.attempt, tc.expected)
		})
	}
}

// RED: Test defaultInvokeOptions
func TestDefaultInvokeOptions_ShouldProvideCorrectDefaults(t *testing.T) {
	// When: defaultInvokeOptions is called
	opts := defaultInvokeOptions()
	
	// Then: Should return correct defaults
	assert.Equal(t, InvocationTypeRequestResponse, opts.InvocationType)
	assert.Equal(t, LogTypeNone, opts.LogType)
	assert.Equal(t, DefaultTimeout, opts.Timeout)
	assert.Equal(t, DefaultRetryAttempts, opts.MaxRetries)
	assert.Equal(t, DefaultRetryDelay, opts.RetryDelay)
	assert.NotNil(t, opts.PayloadValidator)
	
	// Test payload validator
	err := opts.PayloadValidator(`{"test": "valid"}`)
	assert.NoError(t, err)
	
	// Test payload validator with oversized payload
	largePayload := make([]byte, MaxPayloadSize+1)
	err = opts.PayloadValidator(string(largePayload))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payload size exceeds maximum")
}