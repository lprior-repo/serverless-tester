// Package lambda provides comprehensive AWS Lambda testing utilities.
// This file consolidates all Lambda testing functionality with 90%+ coverage.
//
// Following strict TDD principles:
// - Red: Write failing tests first
// - Green: Write minimal code to pass tests
// - Refactor: Clean up code while maintaining passing tests
package lambda

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data factories following functional programming principles
func createTestContext() *TestContext {
	return &TestContext{
		T:      &mockTestingT{},
		Region: "us-east-1",
	}
}

func createValidFunctionName() string {
	return "test-function"
}

func createInvalidFunctionName() string {
	return "invalid@function#name"
}

func createTestPayload() string {
	return `{"test": "payload"}`
}

func createInvalidPayload() string {
	return `{"test": invalid json}`
}

func createMockLambdaClient() *MockLambdaClient {
	return &MockLambdaClient{}
}

func createMockCloudWatchClient() *MockCloudWatchLogsClient {
	return &MockCloudWatchLogsClient{}
}

// Mock testing interface for pure functional testing
type mockTestingT struct {
	errors   []string
	failures []string
	fatals   []string
	helpers  []string
	name     string
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *mockTestingT) Error(args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprint(args...))
}

func (m *mockTestingT) Fail() {
	m.failures = append(m.failures, "fail called")
}

func (m *mockTestingT) FailNow() {
	m.fatals = append(m.fatals, "failnow called")
}

func (m *mockTestingT) Helper() {
	m.helpers = append(m.helpers, "helper called")
}

func (m *mockTestingT) Fatal(args ...interface{}) {
	m.fatals = append(m.fatals, fmt.Sprint(args...))
}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.fatals = append(m.fatals, fmt.Sprintf(format, args...))
}

func (m *mockTestingT) Name() string {
	if m.name == "" {
		return "MockTest"
	}
	return m.name
}

func (m *mockTestingT) Reset() {
	m.errors = nil
	m.failures = nil
	m.fatals = nil
	m.helpers = nil
	m.name = ""
}

// ==============================================================================
// TDD Test Suite 1: Validation Functions
// ==============================================================================

// RED: Write failing test first
func TestValidateFunctionName_WithEmptyName_ShouldReturnError(t *testing.T) {
	// RED: This test should fail because validateFunctionName doesn't exist yet
	err := validateFunctionName("")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFunctionName, err)
}

func TestValidateFunctionName_WithValidName_ShouldReturnNil(t *testing.T) {
	// RED: This test should fail initially
	err := validateFunctionName("valid-function-name")
	assert.NoError(t, err)
}

func TestValidateFunctionName_WithInvalidCharacters_ShouldReturnError(t *testing.T) {
	// RED: This test should fail initially
	invalidNames := []string{
		"invalid@name",
		"invalid#name",
		"invalid name", // space
		"invalid.name",
		"invalid$name",
	}

	for _, invalidName := range invalidNames {
		t.Run(fmt.Sprintf("InvalidName_%s", invalidName), func(t *testing.T) {
			err := validateFunctionName(invalidName)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid character")
		})
	}
}

func TestValidateFunctionName_WithTooLongName_ShouldReturnError(t *testing.T) {
	// RED: This test should fail initially
	longName := strings.Repeat("a", 65) // AWS limit is 64 characters
	err := validateFunctionName(longName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name too long")
}

func TestValidatePayload_WithEmptyPayload_ShouldReturnNil(t *testing.T) {
	// RED: This test should fail because validatePayload doesn't exist yet
	err := validatePayload("")
	assert.NoError(t, err)
}

func TestValidatePayload_WithValidJSON_ShouldReturnNil(t *testing.T) {
	// RED: This test should fail initially
	validPayloads := []string{
		`{"test": "value"}`,
		`[]`,
		`"string"`,
		`123`,
		`true`,
		`null`,
	}

	for _, payload := range validPayloads {
		t.Run(fmt.Sprintf("ValidPayload_%s", payload), func(t *testing.T) {
			err := validatePayload(payload)
			assert.NoError(t, err)
		})
	}
}

func TestValidatePayload_WithInvalidJSON_ShouldReturnError(t *testing.T) {
	// RED: This test should fail initially
	invalidPayloads := []string{
		`{"test": invalid}`,
		`{broken json}`,
		`{"unclosed": `,
		`invalid`,
	}

	for _, payload := range invalidPayloads {
		t.Run(fmt.Sprintf("InvalidPayload_%s", payload), func(t *testing.T) {
			err := validatePayload(payload)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidPayload, errors.Unwrap(err))
		})
	}
}

// ==============================================================================
// TDD Test Suite 2: Helper Functions
// ==============================================================================

func TestDefaultInvokeOptions_ShouldReturnCorrectDefaults(t *testing.T) {
	// RED: This test should fail initially
	opts := defaultInvokeOptions()

	assert.Equal(t, InvocationTypeRequestResponse, opts.InvocationType)
	assert.Equal(t, LogTypeNone, opts.LogType)
	assert.Equal(t, DefaultTimeout, opts.Timeout)
	assert.Equal(t, DefaultRetryAttempts, opts.MaxRetries)
	assert.Equal(t, DefaultRetryDelay, opts.RetryDelay)
	assert.NotNil(t, opts.PayloadValidator)
}

func TestMergeInvokeOptions_WithNilInput_ShouldReturnDefaults(t *testing.T) {
	// RED: This test should fail initially
	opts := mergeInvokeOptions(nil)

	assert.Equal(t, InvocationTypeRequestResponse, opts.InvocationType)
	assert.Equal(t, LogTypeNone, opts.LogType)
	assert.Equal(t, DefaultTimeout, opts.Timeout)
	assert.Equal(t, DefaultRetryAttempts, opts.MaxRetries)
	assert.Equal(t, DefaultRetryDelay, opts.RetryDelay)
	assert.NotNil(t, opts.PayloadValidator)
}

func TestMergeInvokeOptions_WithPartialOptions_ShouldMergeWithDefaults(t *testing.T) {
	// RED: This test should fail initially
	userOpts := &InvokeOptions{
		InvocationType: InvocationTypeEvent,
		Timeout:        60 * time.Second,
	}

	opts := mergeInvokeOptions(userOpts)

	assert.Equal(t, InvocationTypeEvent, opts.InvocationType)
	assert.Equal(t, LogTypeNone, opts.LogType) // Default
	assert.Equal(t, 60*time.Second, opts.Timeout)
	assert.Equal(t, DefaultRetryAttempts, opts.MaxRetries) // Default
	assert.Equal(t, DefaultRetryDelay, opts.RetryDelay)   // Default
	assert.NotNil(t, opts.PayloadValidator)               // Default
}

func TestCalculateBackoffDelay_WithZeroAttempt_ShouldReturnBaseDelay(t *testing.T) {
	// RED: This test should fail initially
	config := defaultRetryConfig()
	delay := calculateBackoffDelay(0, config)
	assert.Equal(t, config.BaseDelay, delay)
}

func TestCalculateBackoffDelay_WithIncrementingAttempts_ShouldIncreaseExponentially(t *testing.T) {
	// RED: This test should fail initially
	config := RetryConfig{
		BaseDelay:   1 * time.Second,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
		MaxAttempts: 5,
	}

	delay1 := calculateBackoffDelay(1, config)
	delay2 := calculateBackoffDelay(2, config)
	delay3 := calculateBackoffDelay(3, config)

	assert.Equal(t, 2*time.Second, delay1)
	assert.Equal(t, 4*time.Second, delay2)
	assert.Equal(t, 8*time.Second, delay3)
}

func TestCalculateBackoffDelay_WithOverflow_ShouldCapAtMaxDelay(t *testing.T) {
	// RED: This test should fail initially
	config := RetryConfig{
		BaseDelay:   1 * time.Second,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
		MaxAttempts: 10,
	}

	delay := calculateBackoffDelay(10, config)
	assert.Equal(t, config.MaxDelay, delay)
}

// ==============================================================================
// TDD Test Suite 3: Invoke Functions
// ==============================================================================

func TestInvokeE_WithValidInput_ShouldReturnResult(t *testing.T) {
	// RED: This test should fail initially
	// Setup mock client
	mockClient := createMockLambdaClient()
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 200,
		Payload:    []byte(`{"result": "success"}`),
	}

	// Setup mock factory
	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	result, err := InvokeE(ctx, createValidFunctionName(), createTestPayload())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(200), result.StatusCode)
	assert.Equal(t, `{"result": "success"}`, result.Payload)
	assert.Contains(t, mockClient.InvokeCalls, createValidFunctionName())
}

func TestInvokeE_WithInvalidFunctionName_ShouldReturnError(t *testing.T) {
	// RED: This test should fail initially
	ctx := createTestContext()
	result, err := InvokeE(ctx, createInvalidFunctionName(), createTestPayload())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidFunctionName, errors.Unwrap(err))
}

func TestInvokeE_WithInvalidPayload_ShouldReturnError(t *testing.T) {
	// RED: This test should fail initially
	ctx := createTestContext()
	result, err := InvokeE(ctx, createValidFunctionName(), createInvalidPayload())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidPayload, errors.Unwrap(err))
}

func TestInvokeE_WithClientError_ShouldReturnError(t *testing.T) {
	// RED: This test should fail initially
	mockClient := createMockLambdaClient()
	mockClient.InvokeError = errors.New("AWS API error")

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	result, err := InvokeE(ctx, createValidFunctionName(), createTestPayload())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "AWS API error")
}

func TestInvokeWithOptionsE_WithCustomOptions_ShouldUseOptions(t *testing.T) {
	// RED: This test should fail initially
	mockClient := createMockLambdaClient()
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 200,
		Payload:    []byte(`{"result": "success"}`),
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	opts := &InvokeOptions{
		InvocationType: InvocationTypeEvent,
		LogType:        LogTypeTail,
		ClientContext:  "test-context",
		Qualifier:      "test-qualifier",
	}

	result, err := InvokeWithOptionsE(ctx, createValidFunctionName(), createTestPayload(), opts)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(200), result.StatusCode)
}

// ==============================================================================
// TDD Test Suite 4: Constants and Error Values  
// ==============================================================================

func TestConstants_ShouldHaveExpectedValues(t *testing.T) {
	// RED: This test should pass since constants are already defined
	assert.Equal(t, 30*time.Second, DefaultTimeout)
	assert.Equal(t, 128, DefaultMemorySize) // Fixed: remove int32 cast
	assert.Equal(t, 6*1024*1024, MaxPayloadSize)
	assert.Equal(t, 3, DefaultRetryAttempts)
	assert.Equal(t, 1*time.Second, DefaultRetryDelay)

	assert.Equal(t, types.InvocationTypeRequestResponse, InvocationTypeRequestResponse)
	assert.Equal(t, types.InvocationTypeEvent, InvocationTypeEvent)
	assert.Equal(t, types.InvocationTypeDryRun, InvocationTypeDryRun)

	assert.Equal(t, types.LogTypeNone, LogTypeNone)
	assert.Equal(t, types.LogTypeTail, LogTypeTail)
}

// Test error constants
func TestErrorConstants_ShouldBeDefinedCorrectly(t *testing.T) {
	// RED: This test should pass since error constants are already defined
	assert.Equal(t, "lambda function not found", ErrFunctionNotFound.Error())
	assert.Equal(t, "invalid payload format", ErrInvalidPayload.Error())
	assert.Equal(t, "lambda invocation timeout", ErrInvocationTimeout.Error())
	assert.Equal(t, "lambda invocation failed", ErrInvocationFailed.Error())
	assert.Equal(t, "failed to retrieve lambda logs", ErrLogRetrievalFailed.Error())
	assert.Equal(t, "invalid function name", ErrInvalidFunctionName.Error())
}

// ==============================================================================
// TDD Test Suite 5: Utility Functions that should pass
// ==============================================================================

func TestContains_WithMatchingSubstring_ShouldReturnTrue(t *testing.T) {
	// Should PASS as contains function exists
	testCases := []struct {
		s      string
		substr string
	}{
		{"hello world", "hello"},
		{"hello world", "world"},
		{"hello world", "llo wo"},
		{"ResourceNotFoundException", "ResourceNotFound"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_contains_%s", tc.s, tc.substr), func(t *testing.T) {
			result := contains(tc.s, tc.substr)
			assert.True(t, result)
		})
	}
}

func TestContains_WithNonMatchingSubstring_ShouldReturnFalse(t *testing.T) {
	// Should PASS as contains function exists
	testCases := []struct {
		s      string
		substr string
	}{
		{"hello world", "xyz"},
		{"ResourceNotFoundException", "InvalidParameter"},
		{"test", "testing"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_does_not_contain_%s", tc.s, tc.substr), func(t *testing.T) {
			result := contains(tc.s, tc.substr)
			assert.False(t, result)
		})
	}
}

func TestIndexOf_WithMatchingSubstring_ShouldReturnCorrectIndex(t *testing.T) {
	// Should PASS as indexOf function exists
	testCases := []struct {
		s      string
		substr string
		index  int
	}{
		{"hello world", "hello", 0},
		{"hello world", "world", 6},
		{"hello world", "llo", 2},
		{"test test", "test", 0}, // Should return first occurrence
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_indexOf_%s", tc.s, tc.substr), func(t *testing.T) {
			result := indexOf(tc.s, tc.substr)
			assert.Equal(t, tc.index, result)
		})
	}
}

func TestIndexOf_WithNonMatchingSubstring_ShouldReturnNegativeOne(t *testing.T) {
	// Should PASS as indexOf function exists
	testCases := []struct {
		s      string
		substr string
	}{
		{"hello world", "xyz"},
		{"test", "testing"},
		{"", "test"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_indexOf_%s", tc.s, tc.substr), func(t *testing.T) {
			result := indexOf(tc.s, tc.substr)
			assert.Equal(t, -1, result)
		})
	}
}

func TestSanitizeLogResult_ShouldRemoveInternalLogLines(t *testing.T) {
	// Should PASS as sanitizeLogResult function exists
	logResult := "Application log line 1\nRequestId: 12345-abcde\nDuration: 1000.52 ms\nBilled Duration: 1001 ms\nApplication log line 2\nAnother application line"

	sanitized := sanitizeLogResult(logResult)

	assert.Contains(t, sanitized, "Application log line 1")
	assert.Contains(t, sanitized, "Application log line 2")
	assert.Contains(t, sanitized, "Another application line")
	assert.NotContains(t, sanitized, "RequestId:")
	assert.NotContains(t, sanitized, "Duration:")
	assert.NotContains(t, sanitized, "Billed Duration:")
}

// ==============================================================================
// TDD Test Suite 6: Event Builder Functions 
// ==============================================================================

func TestBuildS3EventE_WithValidInput_ShouldReturnValidEvent(t *testing.T) {
	// Should PASS as BuildS3EventE function exists
	bucketName := "test-bucket"
	objectKey := "test-key.txt"
	eventName := "s3:ObjectCreated:Put"

	eventJSON, err := BuildS3EventE(bucketName, objectKey, eventName)

	assert.NoError(t, err)
	assert.NotEmpty(t, eventJSON)

	// Parse the event to verify structure
	var event S3Event
	err = json.Unmarshal([]byte(eventJSON), &event)
	assert.NoError(t, err)

	assert.Len(t, event.Records, 1)
	record := event.Records[0]
	assert.Equal(t, eventName, record.EventName)
	assert.Equal(t, "aws:s3", record.EventSource)
	assert.Equal(t, bucketName, record.S3.Bucket.Name)
	assert.Equal(t, objectKey, record.S3.Object.Key)
}

func TestBuildS3EventE_WithEmptyBucket_ShouldReturnError(t *testing.T) {
	// Should PASS as BuildS3EventE function exists
	eventJSON, err := BuildS3EventE("", "test-key.txt", "s3:ObjectCreated:Put")

	assert.Error(t, err)
	assert.Empty(t, eventJSON)
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
}

func TestBuildS3EventE_WithEmptyObjectKey_ShouldReturnError(t *testing.T) {
	// Should PASS as BuildS3EventE function exists
	eventJSON, err := BuildS3EventE("test-bucket", "", "s3:ObjectCreated:Put")

	assert.Error(t, err)
	assert.Empty(t, eventJSON)
	assert.Contains(t, err.Error(), "object key cannot be empty")
}

func TestBuildS3EventE_WithDefaultEventName_ShouldUseDefault(t *testing.T) {
	// Should PASS as BuildS3EventE function exists
	bucketName := "test-bucket"
	objectKey := "test-key.txt"

	eventJSON, err := BuildS3EventE(bucketName, objectKey, "")

	assert.NoError(t, err)
	assert.NotEmpty(t, eventJSON)

	var event S3Event
	err = json.Unmarshal([]byte(eventJSON), &event)
	assert.NoError(t, err)

	assert.Equal(t, "s3:ObjectCreated:Put", event.Records[0].EventName)
}

// ==============================================================================
// TDD Test Suite 7: Event Parsing Functions
// ==============================================================================

func TestParseS3EventE_WithValidEvent_ShouldParseCorrectly(t *testing.T) {
	// Should PASS as ParseS3EventE function exists
	eventJSON, err := BuildS3EventE("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
	require.NoError(t, err)

	event, err := ParseS3EventE(eventJSON)

	assert.NoError(t, err)
	assert.Len(t, event.Records, 1)
	assert.Equal(t, "s3:ObjectCreated:Put", event.Records[0].EventName)
	assert.Equal(t, "test-bucket", event.Records[0].S3.Bucket.Name)
	assert.Equal(t, "test-key.txt", event.Records[0].S3.Object.Key)
}

func TestParseS3EventE_WithInvalidJSON_ShouldReturnError(t *testing.T) {
	// Should PASS as ParseS3EventE function exists
	invalidJSON := `{"Records": [invalid json]}`

	event, err := ParseS3EventE(invalidJSON)

	assert.Error(t, err)
	assert.Equal(t, S3Event{}, event)
	assert.Contains(t, err.Error(), "failed to parse S3 event")
}

func TestExtractS3BucketNameE_WithValidEvent_ShouldExtractBucket(t *testing.T) {
	// Should PASS as ExtractS3BucketNameE function exists
	eventJSON, err := BuildS3EventE("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
	require.NoError(t, err)

	bucketName, err := ExtractS3BucketNameE(eventJSON)

	assert.NoError(t, err)
	assert.Equal(t, "test-bucket", bucketName)
}

func TestExtractS3BucketNameE_WithNoRecords_ShouldReturnError(t *testing.T) {
	// Should PASS as ExtractS3BucketNameE function exists
	emptyEventJSON := `{"Records": []}`

	bucketName, err := ExtractS3BucketNameE(emptyEventJSON)

	assert.Error(t, err)
	assert.Empty(t, bucketName)
	assert.Contains(t, err.Error(), "no S3 records found in event")
}

func TestExtractS3ObjectKeyE_WithValidEvent_ShouldExtractKey(t *testing.T) {
	// Should PASS as ExtractS3ObjectKeyE function exists
	eventJSON, err := BuildS3EventE("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
	require.NoError(t, err)

	objectKey, err := ExtractS3ObjectKeyE(eventJSON)

	assert.NoError(t, err)
	assert.Equal(t, "test-key.txt", objectKey)
}

func TestExtractS3ObjectKeyE_WithNoRecords_ShouldReturnError(t *testing.T) {
	// Should PASS as ExtractS3ObjectKeyE function exists
	emptyEventJSON := `{"Records": []}`

	objectKey, err := ExtractS3ObjectKeyE(emptyEventJSON)

	assert.Error(t, err)
	assert.Empty(t, objectKey)
	assert.Contains(t, err.Error(), "no S3 records found in event")
}

// ==============================================================================
// TDD Test Suite 8: Function Management Comprehensive Tests
// ==============================================================================

func TestGetFunctionE_WithValidFunction_ShouldReturnConfiguration(t *testing.T) {
	// Setup mock client
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test"),
			Runtime:      types.RuntimeNodejs18x,
			Handler:      aws.String("index.handler"),
			State:        types.StateActive,
			Description:  aws.String("Test function"),
			Timeout:      aws.Int32(30),
			MemorySize:   aws.Int32(256),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"ENV_VAR": "test_value",
				},
			},
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	config, err := GetFunctionE(ctx, createValidFunctionName())

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, createValidFunctionName(), config.FunctionName)
	assert.Equal(t, types.RuntimeNodejs18x, config.Runtime)
	assert.Equal(t, "index.handler", config.Handler)
	assert.Equal(t, types.StateActive, config.State)
	assert.Equal(t, "Test function", config.Description)
	assert.Equal(t, int32(30), config.Timeout)
	assert.Equal(t, int32(256), config.MemorySize)
	assert.Equal(t, "test_value", config.Environment["ENV_VAR"])
}

func TestGetFunctionE_WithInvalidFunctionName_ShouldReturnError(t *testing.T) {
	ctx := createTestContext()
	config, err := GetFunctionE(ctx, createInvalidFunctionName())

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "invalid function name")
}

func TestGetFunctionE_WithNonExistentFunction_ShouldReturnError(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionError = errors.New("ResourceNotFoundException")

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	config, err := GetFunctionE(ctx, createValidFunctionName())

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to get function configuration")
}

func TestFunctionExistsE_WithExistingFunction_ShouldReturnTrue(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	exists, err := FunctionExistsE(ctx, createValidFunctionName())

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestFunctionExistsE_WithNonExistentFunction_ShouldReturnFalse(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionError = errors.New("ResourceNotFoundException: Function not found")

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	exists, err := FunctionExistsE(ctx, createValidFunctionName())

	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestFunctionExistsE_WithOtherError_ShouldReturnError(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionError = errors.New("AccessDeniedException")

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	exists, err := FunctionExistsE(ctx, createValidFunctionName())

	assert.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "AccessDeniedException")
}

// ==============================================================================
// TDD Test Suite 9: Payload Processing
// ==============================================================================

func TestMarshalPayloadE_WithValidObject_ShouldReturnJSON(t *testing.T) {
	testObj := map[string]interface{}{
		"key":    "value",
		"number": 42,
		"bool":   true,
	}

	result, err := MarshalPayloadE(testObj)

	assert.NoError(t, err)
	assert.Contains(t, result, `"key":"value"`)
	assert.Contains(t, result, `"number":42`)
	assert.Contains(t, result, `"bool":true`)
}

func TestMarshalPayloadE_WithNilInput_ShouldReturnEmptyString(t *testing.T) {
	result, err := MarshalPayloadE(nil)

	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestMarshalPayloadE_WithUnmarshalableObject_ShouldReturnError(t *testing.T) {
	// Functions cannot be marshaled to JSON
	unmarshalable := map[string]interface{}{
		"func": func() {},
	}

	result, err := MarshalPayloadE(unmarshalable)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to marshal payload")
}

// ==============================================================================
// TDD Test Suite 10: Error Handling Functions
// ==============================================================================

func TestIsNonRetryableError_WithRetryableError_ShouldReturnFalse(t *testing.T) {
	retryableErrors := []error{
		errors.New("ThrottleException"),
		errors.New("ServiceUnavailableException"),
		errors.New("InternalServerError"),
		errors.New("connection timeout"),
	}

	for _, err := range retryableErrors {
		result := isNonRetryableError(err)
		assert.False(t, result, "Error %v should be retryable", err)
	}
}

func TestIsNonRetryableError_WithNonRetryableError_ShouldReturnTrue(t *testing.T) {
	nonRetryableErrors := []error{
		errors.New("InvalidParameterValueException"),
		errors.New("ResourceNotFoundException"),
		errors.New("InvalidRequestContentException"),
		errors.New("RequestTooLargeException"),
		errors.New("UnsupportedMediaTypeException"),
	}

	for _, err := range nonRetryableErrors {
		result := isNonRetryableError(err)
		assert.True(t, result, "Error %v should be non-retryable", err)
	}
}

func TestIsNonRetryableError_WithNilError_ShouldReturnFalse(t *testing.T) {
	result := isNonRetryableError(nil)
	assert.False(t, result)
}

// ==============================================================================
// TDD Test Suite 11: Additional Event Builders
// ==============================================================================

func TestBuildDynamoDBEventE_WithValidInput_ShouldReturnValidEvent(t *testing.T) {
	tableName := "test-table"
	eventName := "INSERT"
	keys := map[string]interface{}{
		"id": map[string]interface{}{"S": "test-id"},
	}

	eventJSON, err := BuildDynamoDBEventE(tableName, eventName, keys)

	assert.NoError(t, err)
	assert.NotEmpty(t, eventJSON)

	var event DynamoDBEvent
	err = json.Unmarshal([]byte(eventJSON), &event)
	assert.NoError(t, err)

	assert.Len(t, event.Records, 1)
	record := event.Records[0]
	assert.Equal(t, eventName, record.EventName)
	assert.Equal(t, "aws:dynamodb", record.EventSource)
	assert.Equal(t, keys, record.Dynamodb.Keys)
}

func TestBuildDynamoDBEventE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	eventJSON, err := BuildDynamoDBEventE("", "INSERT", nil)

	assert.Error(t, err)
	assert.Empty(t, eventJSON)
	assert.Contains(t, err.Error(), "table name cannot be empty")
}

func TestBuildSQSEventE_WithValidInput_ShouldReturnValidEvent(t *testing.T) {
	queueUrl := "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
	messageBody := "Test message"

	eventJSON, err := BuildSQSEventE(queueUrl, messageBody)

	assert.NoError(t, err)
	assert.NotEmpty(t, eventJSON)

	var event SQSEvent
	err = json.Unmarshal([]byte(eventJSON), &event)
	assert.NoError(t, err)

	assert.Len(t, event.Records, 1)
	record := event.Records[0]
	assert.Equal(t, messageBody, record.Body)
	assert.Equal(t, "aws:sqs", record.EventSource)
	assert.Equal(t, queueUrl, record.EventSourceARN)
}

func TestBuildSQSEventE_WithEmptyQueueUrl_ShouldReturnError(t *testing.T) {
	eventJSON, err := BuildSQSEventE("", "test message")

	assert.Error(t, err)
	assert.Empty(t, eventJSON)
	assert.Contains(t, err.Error(), "queue URL cannot be empty")
}

func TestBuildSNSEventE_WithValidInput_ShouldReturnValidEvent(t *testing.T) {
	topicArn := "arn:aws:sns:us-east-1:123456789012:test-topic"
	message := "Test SNS message"

	eventJSON, err := BuildSNSEventE(topicArn, message)

	assert.NoError(t, err)
	assert.NotEmpty(t, eventJSON)

	var event SNSEvent
	err = json.Unmarshal([]byte(eventJSON), &event)
	assert.NoError(t, err)

	assert.Len(t, event.Records, 1)
	record := event.Records[0]
	assert.Equal(t, message, record.Sns.Message)
	assert.Equal(t, "aws:sns", record.EventSource)
	assert.Equal(t, topicArn, record.Sns.TopicArn)
}

func TestBuildSNSEventE_WithEmptyTopicArn_ShouldReturnError(t *testing.T) {
	eventJSON, err := BuildSNSEventE("", "test message")

	assert.Error(t, err)
	assert.Empty(t, eventJSON)
	assert.Contains(t, err.Error(), "topic ARN cannot be empty")
}

func TestBuildAPIGatewayEventE_WithValidInput_ShouldReturnValidEvent(t *testing.T) {
	httpMethod := "POST"
	path := "/api/test"
	body := `{"key": "value"}`

	eventJSON, err := BuildAPIGatewayEventE(httpMethod, path, body)

	assert.NoError(t, err)
	assert.NotEmpty(t, eventJSON)

	var event APIGatewayEvent
	err = json.Unmarshal([]byte(eventJSON), &event)
	assert.NoError(t, err)

	assert.Equal(t, httpMethod, event.HttpMethod)
	assert.Equal(t, path, event.Path)
	assert.Equal(t, body, event.Body)
	assert.Equal(t, httpMethod, event.RequestContext.HttpMethod)
}

func TestBuildAPIGatewayEventE_WithDefaults_ShouldUseDefaults(t *testing.T) {
	eventJSON, err := BuildAPIGatewayEventE("", "", "")

	assert.NoError(t, err)
	assert.NotEmpty(t, eventJSON)

	var event APIGatewayEvent
	err = json.Unmarshal([]byte(eventJSON), &event)
	assert.NoError(t, err)

	assert.Equal(t, "GET", event.HttpMethod) // Default
	assert.Equal(t, "/", event.Path)         // Default
	assert.Equal(t, "", event.Body)
}

// ==============================================================================
// TDD Test Suite 12: Non-Error Wrapper Functions
// ==============================================================================

func TestInvoke_WithSuccessfulInvocation_ShouldReturnResult(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 200,
		Payload:    []byte(`{"result": "success"}`),
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	result := Invoke(ctx, createValidFunctionName(), createTestPayload())

	assert.NotNil(t, result)
	assert.Equal(t, int32(200), result.StatusCode)
	assert.Equal(t, `{"result": "success"}`, result.Payload)

	// Verify no errors were recorded on the mock testing interface
	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestGetFunction_WithSuccessfulGet_ShouldReturnConfiguration(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			Runtime:      types.RuntimeNodejs18x,
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	config := GetFunction(ctx, createValidFunctionName())

	assert.NotNil(t, config)
	assert.Equal(t, createValidFunctionName(), config.FunctionName)
	assert.Equal(t, types.RuntimeNodejs18x, config.Runtime)

	// Verify no errors were recorded
	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

// ==============================================================================
// TDD Test Suite 13: Additional Invoke Scenarios
// ==============================================================================

func TestInvokeAsyncE_ShouldUseEventInvocationType(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 202, // Async invocation status
		Payload:    nil,
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	result, err := InvokeAsyncE(ctx, createValidFunctionName(), createTestPayload())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(202), result.StatusCode)
}

func TestDryRunInvokeE_ShouldUseDryRunInvocationType(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 204, // Dry run status
		Payload:    nil,
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	result, err := DryRunInvokeE(ctx, createValidFunctionName(), createTestPayload())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(204), result.StatusCode)
}

func TestInvokeWithRetryE_WithCustomRetries_ShouldUseCustomRetries(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode: 200,
		Payload:    []byte(`{"result": "success"}`),
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	result, err := InvokeWithRetryE(ctx, createValidFunctionName(), createTestPayload(), 5)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(200), result.StatusCode)
}

// ==============================================================================
// TDD Test Suite 14: Comprehensive Edge Cases
// ==============================================================================

func TestLogOperation_ShouldNotPanic(t *testing.T) {
	// This is a side-effect function, mainly testing it doesn't panic
	assert.NotPanics(t, func() {
		logOperation("test_operation", "test-function", map[string]interface{}{
			"test_param": "test_value",
		})
	})
}

func TestLogResult_ShouldNotPanic(t *testing.T) {
	// This is a side-effect function, mainly testing it doesn't panic
	assert.NotPanics(t, func() {
		logResult("test_operation", "test-function", true, time.Second, nil)
	})

	assert.NotPanics(t, func() {
		logResult("test_operation", "test-function", false, time.Second, errors.New("test error"))
	})
}

func TestCreateLambdaClient_ShouldReturnClient(t *testing.T) {
	ctx := createTestContext()
	client := createLambdaClient(ctx)
	assert.NotNil(t, client)
}

func TestDefaultRetryConfig_ShouldReturnCorrectDefaults(t *testing.T) {
	config := defaultRetryConfig()

	assert.Equal(t, DefaultRetryAttempts, config.MaxAttempts)
	assert.Equal(t, DefaultRetryDelay, config.BaseDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

// ==============================================================================
// TDD Test Suite 15: Assertion Functions - Comprehensive Coverage  
// ==============================================================================

func TestAssertFunctionExists_WithExistingFunction_ShouldNotFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	// This should not fail/panic since function exists
	assert.NotPanics(t, func() {
		AssertFunctionExists(ctx, createValidFunctionName())
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertFunctionExists_WithNonExistentFunction_ShouldFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionError = errors.New("ResourceNotFoundException: Function not found")

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	AssertFunctionExists(ctx, createValidFunctionName())

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "to exist")
}

func TestAssertFunctionDoesNotExist_WithNonExistentFunction_ShouldNotFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionError = errors.New("ResourceNotFoundException: Function not found")

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertFunctionDoesNotExist(ctx, createValidFunctionName())
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertFunctionDoesNotExist_WithExistingFunction_ShouldFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	AssertFunctionDoesNotExist(ctx, createValidFunctionName())

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "not to exist")
}

func TestAssertFunctionRuntime_WithCorrectRuntime_ShouldNotFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			Runtime:      types.RuntimeNodejs18x,
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertFunctionRuntime(ctx, createValidFunctionName(), types.RuntimeNodejs18x)
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertFunctionRuntime_WithIncorrectRuntime_ShouldFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			Runtime:      types.RuntimeNodejs18x,
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	AssertFunctionRuntime(ctx, createValidFunctionName(), types.RuntimePython39)

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "runtime")
}

func TestAssertFunctionHandler_WithCorrectHandler_ShouldNotFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			Handler:      aws.String("index.handler"),
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertFunctionHandler(ctx, createValidFunctionName(), "index.handler")
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertFunctionHandler_WithIncorrectHandler_ShouldFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			Handler:      aws.String("index.handler"),
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	AssertFunctionHandler(ctx, createValidFunctionName(), "different.handler")

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "handler")
}

func TestAssertFunctionTimeout_WithCorrectTimeout_ShouldNotFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			Timeout:      aws.Int32(30),
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertFunctionTimeout(ctx, createValidFunctionName(), 30)
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertFunctionTimeout_WithIncorrectTimeout_ShouldFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			Timeout:      aws.Int32(30),
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	AssertFunctionTimeout(ctx, createValidFunctionName(), 60)

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "timeout")
}

func TestAssertFunctionMemorySize_WithCorrectMemorySize_ShouldNotFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			MemorySize:   aws.Int32(256),
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertFunctionMemorySize(ctx, createValidFunctionName(), 256)
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertFunctionMemorySize_WithIncorrectMemorySize_ShouldFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			MemorySize:   aws.Int32(256),
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	AssertFunctionMemorySize(ctx, createValidFunctionName(), 512)

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "memory")
}

func TestAssertFunctionState_WithCorrectState_ShouldNotFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			State:        types.StateActive,
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertFunctionState(ctx, createValidFunctionName(), types.StateActive)
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertFunctionState_WithIncorrectState_ShouldFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			State:        types.StateActive,
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	AssertFunctionState(ctx, createValidFunctionName(), types.StatePending)

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "state")
}

func TestAssertEnvVarEquals_WithCorrectEnvVar_ShouldNotFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"TEST_VAR": "test_value",
				},
			},
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertEnvVarEquals(ctx, createValidFunctionName(), "TEST_VAR", "test_value")
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertEnvVarEquals_WithIncorrectEnvVar_ShouldFail(t *testing.T) {
	mockClient := createMockLambdaClient()
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(createValidFunctionName()),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"TEST_VAR": "test_value",
				},
			},
		},
	}

	originalFactory := globalClientFactory
	defer func() { globalClientFactory = originalFactory }()
	globalClientFactory = &MockClientFactory{LambdaClient: mockClient}

	ctx := createTestContext()
	
	AssertEnvVarEquals(ctx, createValidFunctionName(), "TEST_VAR", "different_value")

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "environment variable")
}

func TestAssertInvokeSuccess_WithSuccessfulInvocation_ShouldNotFail(t *testing.T) {
	result := &InvokeResult{
		StatusCode: 200,
		Payload:    `{"result": "success"}`,
	}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertInvokeSuccess(ctx, result)
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertInvokeSuccess_WithErrorStatusCode_ShouldFail(t *testing.T) {
	result := &InvokeResult{
		StatusCode: 500,
		Payload:    `{"error": "Internal Server Error"}`,
	}

	ctx := createTestContext()
	
	AssertInvokeSuccess(ctx, result)

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "successful")
}

func TestAssertInvokeError_WithErrorResult_ShouldNotFail(t *testing.T) {
	result := &InvokeResult{
		StatusCode:    200, // Status code can still be 200
		Payload:       `{"errorMessage": "test error"}`,
		FunctionError: "Handled", // This is what AssertInvokeError checks
	}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertInvokeError(ctx, result)
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertInvokeError_WithSuccessResult_ShouldFail(t *testing.T) {
	result := &InvokeResult{
		StatusCode:    200,
		Payload:       `{"result": "success"}`,
		FunctionError: "", // No function error - this should cause the assertion to fail
	}

	ctx := createTestContext()
	
	AssertInvokeError(ctx, result)

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "fail")
}

func TestAssertPayloadContains_WithMatchingPayload_ShouldNotFail(t *testing.T) {
	result := &InvokeResult{
		StatusCode: 200,
		Payload:    `{"message": "Hello World", "status": "success"}`,
	}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertPayloadContains(ctx, result, "Hello World")
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertPayloadContains_WithNonMatchingPayload_ShouldFail(t *testing.T) {
	result := &InvokeResult{
		StatusCode: 200,
		Payload:    `{"message": "Hello World", "status": "success"}`,
	}

	ctx := createTestContext()
	
	AssertPayloadContains(ctx, result, "Goodbye")

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "contain")
}

func TestAssertPayloadEquals_WithMatchingPayload_ShouldNotFail(t *testing.T) {
	expectedPayload := `{"message": "Hello World"}`
	result := &InvokeResult{
		StatusCode: 200,
		Payload:    expectedPayload,
	}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertPayloadEquals(ctx, result, expectedPayload)
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertPayloadEquals_WithNonMatchingPayload_ShouldFail(t *testing.T) {
	result := &InvokeResult{
		StatusCode: 200,
		Payload:    `{"message": "Hello World"}`,
	}

	ctx := createTestContext()
	
	AssertPayloadEquals(ctx, result, `{"message": "Different Message"}`)

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "equal")
}

func TestAssertExecutionTimeLessThan_WithFasterExecution_ShouldNotFail(t *testing.T) {
	result := &InvokeResult{
		StatusCode:    200,
		Payload:       `{"result": "success"}`,
		ExecutionTime: 500 * time.Millisecond,
	}

	ctx := createTestContext()
	
	assert.NotPanics(t, func() {
		AssertExecutionTimeLessThan(ctx, result, 1*time.Second)
	})

	mockT := ctx.T.(*mockTestingT)
	assert.Empty(t, mockT.errors)
	assert.Empty(t, mockT.fatals)
}

func TestAssertExecutionTimeLessThan_WithSlowerExecution_ShouldFail(t *testing.T) {
	result := &InvokeResult{
		StatusCode:    200,
		Payload:       `{"result": "success"}`,
		ExecutionTime: 2 * time.Second,
	}

	ctx := createTestContext()
	
	AssertExecutionTimeLessThan(ctx, result, 1*time.Second)

	mockT := ctx.T.(*mockTestingT)
	assert.NotEmpty(t, mockT.errors)
	assert.NotEmpty(t, mockT.fatals)
	assert.Contains(t, mockT.errors[0], "execution time")
}