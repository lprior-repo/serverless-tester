package lambda

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock testing interface for assertions testing
type mockAssertionTestingT struct {
	errors   []string
	failed   bool
	failNow  bool
}

func (m *mockAssertionTestingT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
	m.failed = true
}

func (m *mockAssertionTestingT) Error(args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprint(args...))
	m.failed = true
}

func (m *mockAssertionTestingT) Fail() {
	m.failed = true
}

func (m *mockAssertionTestingT) FailNow() {
	m.failNow = true
}

func (m *mockAssertionTestingT) Fatal(args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprint(args...))
	m.failed = true
	m.failNow = true
}

func (m *mockAssertionTestingT) Fatalf(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
	m.failed = true
	m.failNow = true
}

func (m *mockAssertionTestingT) Helper() {
	// No-op for mock
}

func (m *mockAssertionTestingT) Name() string {
	return "mockAssertionTestingT"
}

func (m *mockAssertionTestingT) reset() {
	m.errors = nil
	m.failed = false
	m.failNow = false
}

// Test data factories for assertions testing

func createValidInvokeResult() *InvokeResult {
	return &InvokeResult{
		StatusCode:     200,
		Payload:        `{"message": "success", "data": "test"}`,
		LogResult:      "Log output",
		ExecutedVersion: "$LATEST",
		FunctionError:  "",
		ExecutionTime:  100 * time.Millisecond,
	}
}

func createErrorInvokeResult() *InvokeResult {
	return &InvokeResult{
		StatusCode:     200,
		Payload:        `{"errorMessage": "Function error"}`,
		LogResult:      "Error log output",
		ExecutedVersion: "$LATEST",
		FunctionError:  "Handled",
		ExecutionTime:  50 * time.Millisecond,
	}
}

func createTestLogEntriesForAssertions() []LogEntry {
	return []LogEntry{
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			Message:   "Starting function execution",
			Level:     "INFO",
			RequestID: "req-123",
		},
		{
			Timestamp: time.Now().Add(-4 * time.Minute),
			Message:   "Processing data",
			Level:     "DEBUG",
			RequestID: "req-123",
		},
		{
			Timestamp: time.Now().Add(-3 * time.Minute),
			Message:   "Error occurred: validation failed",
			Level:     "ERROR",
			RequestID: "req-123",
		},
	}
}

// RED: Test AssertInvokeSuccess with successful result
func TestAssertInvokeSuccess_WithSuccessfulResult_ShouldPass(t *testing.T) {
	// Given: Mock testing context and successful invoke result
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createValidInvokeResult()
	
	// When: AssertInvokeSuccess is called
	AssertInvokeSuccess(ctx, result)
	
	// Then: No errors should be recorded
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertInvokeSuccess with nil result
func TestAssertInvokeSuccess_WithNilResult_ShouldFail(t *testing.T) {
	// Given: Mock testing context and nil result
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	
	// When: AssertInvokeSuccess is called with nil result
	AssertInvokeSuccess(ctx, nil)
	
	// Then: Should record error and fail
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected invocation result to be non-nil")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertInvokeSuccess with function error
func TestAssertInvokeSuccess_WithFunctionError_ShouldFail(t *testing.T) {
	// Given: Mock testing context and result with function error
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createErrorInvokeResult()
	
	// When: AssertInvokeSuccess is called
	AssertInvokeSuccess(ctx, result)
	
	// Then: Should record error and fail
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected successful invocation, but got function error: Handled")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertInvokeSuccess with bad status code
func TestAssertInvokeSuccess_WithBadStatusCode_ShouldFail(t *testing.T) {
	// Given: Mock testing context and result with bad status code
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := &InvokeResult{
		StatusCode:    500,
		Payload:       `{"error": "Internal error"}`,
		FunctionError: "",
	}
	
	// When: AssertInvokeSuccess is called
	AssertInvokeSuccess(ctx, result)
	
	// Then: Should record error and fail
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected successful status code (2xx), but got 500")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertInvokeError with error result
func TestAssertInvokeError_WithErrorResult_ShouldPass(t *testing.T) {
	// Given: Mock testing context and error result
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createErrorInvokeResult()
	
	// When: AssertInvokeError is called
	AssertInvokeError(ctx, result)
	
	// Then: No errors should be recorded
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertInvokeError with successful result
func TestAssertInvokeError_WithSuccessfulResult_ShouldFail(t *testing.T) {
	// Given: Mock testing context and successful result
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createValidInvokeResult()
	
	// When: AssertInvokeError is called
	AssertInvokeError(ctx, result)
	
	// Then: Should record error and fail
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected invocation to result in function error, but got none")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertPayloadContains with matching content
func TestAssertPayloadContains_WithMatchingContent_ShouldPass(t *testing.T) {
	// Given: Mock testing context and result with expected content
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createValidInvokeResult()
	
	// When: AssertPayloadContains is called with matching text
	AssertPayloadContains(ctx, result, "success")
	
	// Then: No errors should be recorded
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertPayloadContains with non-matching content
func TestAssertPayloadContains_WithNonMatchingContent_ShouldFail(t *testing.T) {
	// Given: Mock testing context and result without expected content
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createValidInvokeResult()
	
	// When: AssertPayloadContains is called with non-matching text
	AssertPayloadContains(ctx, result, "failure")
	
	// Then: Should record error and fail
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected payload to contain 'failure'")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertPayloadEquals with matching payload
func TestAssertPayloadEquals_WithMatchingPayload_ShouldPass(t *testing.T) {
	// Given: Mock testing context and result with specific payload
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createValidInvokeResult()
	
	// When: AssertPayloadEquals is called with matching payload
	AssertPayloadEquals(ctx, result, `{"message": "success", "data": "test"}`)
	
	// Then: No errors should be recorded
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertPayloadEquals with different payload
func TestAssertPayloadEquals_WithDifferentPayload_ShouldFail(t *testing.T) {
	// Given: Mock testing context and result with different payload
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createValidInvokeResult()
	
	// When: AssertPayloadEquals is called with different payload
	AssertPayloadEquals(ctx, result, `{"message": "failure"}`)
	
	// Then: Should record error and fail
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected payload to be")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertExecutionTimeLessThan with acceptable time
func TestAssertExecutionTimeLessThan_WithAcceptableTime_ShouldPass(t *testing.T) {
	// Given: Mock testing context and result with short execution time
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createValidInvokeResult()
	
	// When: AssertExecutionTimeLessThan is called with longer max duration
	AssertExecutionTimeLessThan(ctx, result, 200*time.Millisecond)
	
	// Then: No errors should be recorded
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertExecutionTimeLessThan with excessive time
func TestAssertExecutionTimeLessThan_WithExcessiveTime_ShouldFail(t *testing.T) {
	// Given: Mock testing context and result with long execution time
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	result := createValidInvokeResult()
	
	// When: AssertExecutionTimeLessThan is called with shorter max duration
	AssertExecutionTimeLessThan(ctx, result, 50*time.Millisecond)
	
	// Then: Should record error and fail
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected execution time to be less than")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test ValidateFunctionConfiguration with valid matching configuration
func TestValidateFunctionConfiguration_WithValidMatchingConfig_ShouldReturnNoErrors(t *testing.T) {
	// Given: Actual and expected configurations that match
	actual := &FunctionConfiguration{
		FunctionName: "test-function",
		Runtime:      types.RuntimeNodejs18x,
		Handler:      "index.handler",
		Timeout:      30,
		MemorySize:   256,
		State:        types.StateActive,
		Environment: map[string]string{
			"NODE_ENV": "production",
			"DEBUG":    "false",
		},
	}
	expected := &FunctionConfiguration{
		FunctionName: "test-function",
		Runtime:      types.RuntimeNodejs18x,
		Handler:      "index.handler",
		Timeout:      30,
		MemorySize:   256,
		State:        types.StateActive,
		Environment: map[string]string{
			"NODE_ENV": "production",
			"DEBUG":    "false",
		},
	}
	
	// When: ValidateFunctionConfiguration is called
	errors := ValidateFunctionConfiguration(actual, expected)
	
	// Then: Should return no errors
	assert.Empty(t, errors)
}

// RED: Test ValidateFunctionConfiguration with nil actual configuration
func TestValidateFunctionConfiguration_WithNilActualConfig_ShouldReturnError(t *testing.T) {
	// Given: Nil actual configuration and valid expected configuration
	expected := &FunctionConfiguration{
		FunctionName: "test-function",
		Runtime:      types.RuntimeNodejs18x,
	}
	
	// When: ValidateFunctionConfiguration is called
	errors := ValidateFunctionConfiguration(nil, expected)
	
	// Then: Should return error about nil configuration
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], "function configuration is nil")
}

// RED: Test ValidateFunctionConfiguration with mismatched runtime
func TestValidateFunctionConfiguration_WithMismatchedRuntime_ShouldReturnError(t *testing.T) {
	// Given: Configurations with different runtimes
	actual := &FunctionConfiguration{
		Runtime: types.RuntimePython39,
	}
	expected := &FunctionConfiguration{
		Runtime: types.RuntimeNodejs18x,
	}
	
	// When: ValidateFunctionConfiguration is called
	errors := ValidateFunctionConfiguration(actual, expected)
	
	// Then: Should return runtime mismatch error
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected runtime 'nodejs18.x', got 'python3.9'")
}

// RED: Test ValidateFunctionConfiguration with missing environment variable
func TestValidateFunctionConfiguration_WithMissingEnvVar_ShouldReturnError(t *testing.T) {
	// Given: Configurations with missing environment variable
	actual := &FunctionConfiguration{
		Environment: map[string]string{
			"NODE_ENV": "production",
		},
	}
	expected := &FunctionConfiguration{
		Environment: map[string]string{
			"NODE_ENV":  "production",
			"LOG_LEVEL": "info",
		},
	}
	
	// When: ValidateFunctionConfiguration is called
	errors := ValidateFunctionConfiguration(actual, expected)
	
	// Then: Should return missing environment variable error
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected environment variable 'LOG_LEVEL' not found")
}

// RED: Test ValidateInvokeResult with successful result expectation
func TestValidateInvokeResult_WithSuccessfulResult_ShouldReturnNoErrors(t *testing.T) {
	// Given: Successful invoke result and expectation of success
	result := createValidInvokeResult()
	
	// When: ValidateInvokeResult is called expecting success
	errors := ValidateInvokeResult(result, true, "success")
	
	// Then: Should return no errors
	assert.Empty(t, errors)
}

// RED: Test ValidateInvokeResult with nil result
func TestValidateInvokeResult_WithNilResult_ShouldReturnError(t *testing.T) {
	// When: ValidateInvokeResult is called with nil result
	errors := ValidateInvokeResult(nil, true, "")
	
	// Then: Should return nil result error
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], "invoke result is nil")
}

// RED: Test ValidateInvokeResult expecting success but got failure
func TestValidateInvokeResult_ExpectingSuccessButGotFailure_ShouldReturnError(t *testing.T) {
	// Given: Error result but expectation of success
	result := createErrorInvokeResult()
	
	// When: ValidateInvokeResult is called expecting success
	errors := ValidateInvokeResult(result, true, "")
	
	// Then: Should return error about unexpected failure
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected success but got function error: Handled")
}

// RED: Test ValidateInvokeResult with payload content validation
func TestValidateInvokeResult_WithPayloadContentValidation_ShouldReturnErrorWhenMissing(t *testing.T) {
	// Given: Result without expected payload content
	result := createValidInvokeResult()
	
	// When: ValidateInvokeResult is called with content expectation
	errors := ValidateInvokeResult(result, true, "failure")
	
	// Then: Should return payload content error
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected payload to contain 'failure'")
}

// RED: Test ValidateLogEntries with valid log entries
func TestValidateLogEntries_WithValidLogEntries_ShouldReturnNoErrors(t *testing.T) {
	// Given: Valid log entries
	logs := createTestLogEntriesForAssertions()
	
	// When: ValidateLogEntries is called with matching criteria
	errors := ValidateLogEntries(logs, 3, "Starting function", "INFO")
	
	// Then: Should return no errors
	assert.Empty(t, errors)
}

// RED: Test ValidateLogEntries with wrong count
func TestValidateLogEntries_WithWrongCount_ShouldReturnError(t *testing.T) {
	// Given: Log entries with count mismatch expectation
	logs := createTestLogEntriesForAssertions()
	
	// When: ValidateLogEntries is called with wrong count
	errors := ValidateLogEntries(logs, 5, "", "")
	
	// Then: Should return count mismatch error
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected 5 log entries, got 3")
}

// RED: Test ValidateLogEntries with missing content
func TestValidateLogEntries_WithMissingContent_ShouldReturnError(t *testing.T) {
	// Given: Log entries without expected content
	logs := createTestLogEntriesForAssertions()
	
	// When: ValidateLogEntries is called with content that doesn't exist
	errors := ValidateLogEntries(logs, -1, "nonexistent content", "")
	
	// Then: Should return content missing error
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected logs to contain 'nonexistent content'")
}

// RED: Test ValidateLogEntries with missing level
func TestValidateLogEntries_WithMissingLevel_ShouldReturnError(t *testing.T) {
	// Given: Log entries without expected level
	logs := createTestLogEntriesForAssertions()
	
	// When: ValidateLogEntries is called with level that doesn't exist
	errors := ValidateLogEntries(logs, -1, "", "CRITICAL")
	
	// Then: Should return level missing error
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected logs to contain level 'CRITICAL'")
}

// Mock function existence checker for testing assertions that depend on function operations
type mockFunctionChecker struct {
	functionExists     bool
	functionConfig     *FunctionConfiguration
	shouldReturnError  bool
	errorToReturn      error
}

func (m *mockFunctionChecker) checkExists() (bool, error) {
	if m.shouldReturnError {
		return false, m.errorToReturn
	}
	return m.functionExists, nil
}

func (m *mockFunctionChecker) getConfig() *FunctionConfiguration {
	return m.functionConfig
}

// Mock TestContext implementation for testing assertions
type mockFunctionTestContext struct {
	*mockAssertionTestingT
	functionChecker *mockFunctionChecker
}

func newMockFunctionTestContext() *mockFunctionTestContext {
	return &mockFunctionTestContext{
		mockAssertionTestingT: &mockAssertionTestingT{},
		functionChecker:       &mockFunctionChecker{},
	}
}

// RED: Test AssertFunctionExists with existing function
func TestAssertFunctionExists_WithExistingFunction_ShouldPass(t *testing.T) {
	// Given: Mock context with existing function
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionExists = true
	mockT.functionChecker.shouldReturnError = false
	
	// Mock the FunctionExistsE call by temporarily replacing the global behavior
	originalFunctionExistsE := functionExistsEFunc
	functionExistsEFunc = func(ctx *TestContext, functionName string) (bool, error) {
		return mockT.functionChecker.checkExists()
	}
	defer func() {
		functionExistsEFunc = originalFunctionExistsE
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionExists is called
	AssertFunctionExists(ctx, "existing-function")
	
	// Then: Should not fail
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertFunctionExists with non-existing function
func TestAssertFunctionExists_WithNonExistingFunction_ShouldFail(t *testing.T) {
	// Given: Mock context with non-existing function
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionExists = false
	mockT.functionChecker.shouldReturnError = false
	
	// Mock the FunctionExistsE call
	originalFunctionExistsE := functionExistsEFunc
	functionExistsEFunc = func(ctx *TestContext, functionName string) (bool, error) {
		return mockT.functionChecker.checkExists()
	}
	defer func() {
		functionExistsEFunc = originalFunctionExistsE
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionExists is called
	AssertFunctionExists(ctx, "non-existing-function")
	
	// Then: Should fail with appropriate error
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected Lambda function 'non-existing-function' to exist, but it does not")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertFunctionExists with API error
func TestAssertFunctionExists_WithAPIError_ShouldFail(t *testing.T) {
	// Given: Mock context with API error
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.shouldReturnError = true
	mockT.functionChecker.errorToReturn = fmt.Errorf("API error: access denied")
	
	// Mock the FunctionExistsE call
	originalFunctionExistsE := functionExistsEFunc
	functionExistsEFunc = func(ctx *TestContext, functionName string) (bool, error) {
		return mockT.functionChecker.checkExists()
	}
	defer func() {
		functionExistsEFunc = originalFunctionExistsE
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionExists is called
	AssertFunctionExists(ctx, "test-function")
	
	// Then: Should fail with API error
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Failed to check if function exists")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertFunctionDoesNotExist with non-existing function
func TestAssertFunctionDoesNotExist_WithNonExistingFunction_ShouldPass(t *testing.T) {
	// Given: Mock context with non-existing function
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionExists = false
	mockT.functionChecker.shouldReturnError = false
	
	// Mock the FunctionExistsE call
	originalFunctionExistsE := functionExistsEFunc
	functionExistsEFunc = func(ctx *TestContext, functionName string) (bool, error) {
		return mockT.functionChecker.checkExists()
	}
	defer func() {
		functionExistsEFunc = originalFunctionExistsE
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionDoesNotExist is called
	AssertFunctionDoesNotExist(ctx, "non-existing-function")
	
	// Then: Should not fail
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertFunctionDoesNotExist with existing function
func TestAssertFunctionDoesNotExist_WithExistingFunction_ShouldFail(t *testing.T) {
	// Given: Mock context with existing function
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionExists = true
	mockT.functionChecker.shouldReturnError = false
	
	// Mock the FunctionExistsE call
	originalFunctionExistsE := functionExistsEFunc
	functionExistsEFunc = func(ctx *TestContext, functionName string) (bool, error) {
		return mockT.functionChecker.checkExists()
	}
	defer func() {
		functionExistsEFunc = originalFunctionExistsE
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionDoesNotExist is called
	AssertFunctionDoesNotExist(ctx, "existing-function")
	
	// Then: Should fail with appropriate error
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected Lambda function 'existing-function' not to exist, but it does")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertFunctionDoesNotExist with ResourceNotFoundException
func TestAssertFunctionDoesNotExist_WithResourceNotFoundException_ShouldPass(t *testing.T) {
	// Given: Mock context with ResourceNotFoundException
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.shouldReturnError = true
	mockT.functionChecker.errorToReturn = fmt.Errorf("ResourceNotFoundException: Function not found")
	
	// Mock the FunctionExistsE call
	originalFunctionExistsE := functionExistsEFunc
	functionExistsEFunc = func(ctx *TestContext, functionName string) (bool, error) {
		return mockT.functionChecker.checkExists()
	}
	defer func() {
		functionExistsEFunc = originalFunctionExistsE
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionDoesNotExist is called
	AssertFunctionDoesNotExist(ctx, "non-existing-function")
	
	// Then: Should not fail (ResourceNotFoundException is expected)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertFunctionRuntime with matching runtime
func TestAssertFunctionRuntime_WithMatchingRuntime_ShouldPass(t *testing.T) {
	// Given: Mock context with function having expected runtime
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		Runtime: types.RuntimeNodejs18x,
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionRuntime is called
	AssertFunctionRuntime(ctx, "test-function", types.RuntimeNodejs18x)
	
	// Then: Should not fail
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertFunctionRuntime with different runtime
func TestAssertFunctionRuntime_WithDifferentRuntime_ShouldFail(t *testing.T) {
	// Given: Mock context with function having different runtime
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		Runtime: types.RuntimePython39,
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionRuntime is called
	AssertFunctionRuntime(ctx, "test-function", types.RuntimeNodejs18x)
	
	// Then: Should fail with runtime mismatch error
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have runtime 'nodejs18.x', but got 'python3.9'")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertFunctionHandler with matching handler
func TestAssertFunctionHandler_WithMatchingHandler_ShouldPass(t *testing.T) {
	// Given: Mock context with function having expected handler
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		Handler: "index.handler",
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionHandler is called
	AssertFunctionHandler(ctx, "test-function", "index.handler")
	
	// Then: Should not fail
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertFunctionHandler with different handler
func TestAssertFunctionHandler_WithDifferentHandler_ShouldFail(t *testing.T) {
	// Given: Mock context with function having different handler
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		Handler: "app.main",
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionHandler is called
	AssertFunctionHandler(ctx, "test-function", "index.handler")
	
	// Then: Should fail with handler mismatch error
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have handler 'index.handler', but got 'app.main'")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertFunctionTimeout with matching timeout
func TestAssertFunctionTimeout_WithMatchingTimeout_ShouldPass(t *testing.T) {
	// Given: Mock context with function having expected timeout
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		Timeout: 30,
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionTimeout is called
	AssertFunctionTimeout(ctx, "test-function", 30)
	
	// Then: Should not fail
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertFunctionTimeout with different timeout
func TestAssertFunctionTimeout_WithDifferentTimeout_ShouldFail(t *testing.T) {
	// Given: Mock context with function having different timeout
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		Timeout: 60,
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionTimeout is called
	AssertFunctionTimeout(ctx, "test-function", 30)
	
	// Then: Should fail with timeout mismatch error
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have timeout 30 seconds, but got 60")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertFunctionMemorySize with matching memory size
func TestAssertFunctionMemorySize_WithMatchingMemorySize_ShouldPass(t *testing.T) {
	// Given: Mock context with function having expected memory size
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		MemorySize: 256,
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionMemorySize is called
	AssertFunctionMemorySize(ctx, "test-function", 256)
	
	// Then: Should not fail
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertFunctionMemorySize with different memory size
func TestAssertFunctionMemorySize_WithDifferentMemorySize_ShouldFail(t *testing.T) {
	// Given: Mock context with function having different memory size
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		MemorySize: 512,
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionMemorySize is called
	AssertFunctionMemorySize(ctx, "test-function", 256)
	
	// Then: Should fail with memory size mismatch error
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have memory size 256 MB, but got 512")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}

// RED: Test AssertFunctionState with matching state
func TestAssertFunctionState_WithMatchingState_ShouldPass(t *testing.T) {
	// Given: Mock context with function in expected state
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		State: types.StateActive,
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionState is called
	AssertFunctionState(ctx, "test-function", types.StateActive)
	
	// Then: Should not fail
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
}

// RED: Test AssertFunctionState with different state
func TestAssertFunctionState_WithDifferentState_ShouldFail(t *testing.T) {
	// Given: Mock context with function in different state
	mockT := newMockFunctionTestContext()
	mockT.functionChecker.functionConfig = &FunctionConfiguration{
		State: types.StatePending,
	}
	
	// Mock the GetFunction call
	originalGetFunction := getFunctionFunc
	getFunctionFunc = func(ctx *TestContext, functionName string) *FunctionConfiguration {
		return mockT.functionChecker.getConfig()
	}
	defer func() {
		getFunctionFunc = originalGetFunction
	}()
	
	ctx := &TestContext{T: mockT.mockAssertionTestingT}
	
	// When: AssertFunctionState is called
	AssertFunctionState(ctx, "test-function", types.StateActive)
	
	// Then: Should fail with state mismatch error
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to be in state 'Active', but got 'Pending'")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
}