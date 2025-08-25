package lambda

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data factories for invoke assertions testing

func createSuccessfulInvokeResult() *InvokeResult {
	return &InvokeResult{
		StatusCode:      200,
		Payload:         `{"message": "success", "data": {"key": "value"}}`,
		LogResult:       "Log output from function",
		ExecutedVersion: "$LATEST",
		FunctionError:   "",
		ExecutionTime:   150 * time.Millisecond,
	}
}

func createInvokeErrorResult() *InvokeResult {
	return &InvokeResult{
		StatusCode:      200,
		Payload:         `{"errorMessage": "Something went wrong", "errorType": "RuntimeError"}`,
		LogResult:       "Error log output",
		ExecutedVersion: "$LATEST",
		FunctionError:   "Unhandled",
		ExecutionTime:   250 * time.Millisecond,
	}
}

func createSlowInvokeResult() *InvokeResult {
	return &InvokeResult{
		StatusCode:      200,
		Payload:         `{"message": "success"}`,
		LogResult:       "Slow function log",
		ExecutedVersion: "$LATEST",
		FunctionError:   "",
		ExecutionTime:   5 * time.Second,
	}
}

// Tests for AssertInvokeSuccess

func TestAssertInvokeSuccess_WithSuccessfulResult_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createSuccessfulInvokeResult()
	
	// Act
	AssertInvokeSuccess(ctx, result)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for successful invoke")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertInvokeSuccess_WithErrorResult_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createInvokeErrorResult()
	
	// Act
	AssertInvokeSuccess(ctx, result)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for failed invoke")
	assert.Contains(t, mockT.errors[0], "Expected Lambda invocation to succeed")
	assert.Contains(t, mockT.errors[0], "Unhandled")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertInvokeSuccess_WithNilResult_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Act
	AssertInvokeSuccess(ctx, nil)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for nil result")
	assert.Contains(t, mockT.errors[0], "InvokeResult cannot be nil")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertInvokeError

func TestAssertInvokeError_WithErrorResult_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createInvokeErrorResult()
	
	// Act
	AssertInvokeError(ctx, result)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors when expecting error")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertInvokeError_WithSuccessfulResult_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createSuccessfulInvokeResult()
	
	// Act
	AssertInvokeError(ctx, result)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors when expecting error but got success")
	assert.Contains(t, mockT.errors[0], "Expected Lambda invocation to fail")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertInvokeError_WithNilResult_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Act
	AssertInvokeError(ctx, nil)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for nil result")
	assert.Contains(t, mockT.errors[0], "InvokeResult cannot be nil")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertPayloadContains

func TestAssertPayloadContains_WithMatchingText_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createSuccessfulInvokeResult()
	
	// Act
	AssertPayloadContains(ctx, result, "success")
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors when payload contains expected text")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertPayloadContains_WithNonMatchingText_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createSuccessfulInvokeResult()
	
	// Act
	AssertPayloadContains(ctx, result, "failure")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors when payload doesn't contain expected text")
	assert.Contains(t, mockT.errors[0], "Expected payload to contain")
	assert.Contains(t, mockT.errors[0], "failure")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertPayloadContains_WithNilResult_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Act
	AssertPayloadContains(ctx, nil, "test")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for nil result")
	assert.Contains(t, mockT.errors[0], "InvokeResult cannot be nil")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertPayloadEquals

func TestAssertPayloadEquals_WithMatchingPayload_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createSuccessfulInvokeResult()
	expectedPayload := `{"message": "success", "data": {"key": "value"}}`
	
	// Act
	AssertPayloadEquals(ctx, result, expectedPayload)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors when payloads match")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertPayloadEquals_WithDifferentPayload_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createSuccessfulInvokeResult()
	expectedPayload := `{"message": "different"}`
	
	// Act
	AssertPayloadEquals(ctx, result, expectedPayload)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors when payloads don't match")
	assert.Contains(t, mockT.errors[0], "Expected payload to equal")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertPayloadEquals_WithNilResult_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Act
	AssertPayloadEquals(ctx, nil, "test")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for nil result")
	assert.Contains(t, mockT.errors[0], "InvokeResult cannot be nil")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertExecutionTimeLessThan

func TestAssertExecutionTimeLessThan_WithFastExecution_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createSuccessfulInvokeResult() // 150ms execution time
	
	// Act
	AssertExecutionTimeLessThan(ctx, result, 1*time.Second)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors when execution time is less than threshold")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertExecutionTimeLessThan_WithSlowExecution_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	result := createSlowInvokeResult() // 5s execution time
	
	// Act
	AssertExecutionTimeLessThan(ctx, result, 1*time.Second)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors when execution time exceeds threshold")
	assert.Contains(t, mockT.errors[0], "Expected execution time")
	assert.Contains(t, mockT.errors[0], "to be less than")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertExecutionTimeLessThan_WithNilResult_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, _, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Act
	AssertExecutionTimeLessThan(ctx, nil, 1*time.Second)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for nil result")
	assert.Contains(t, mockT.errors[0], "InvokeResult cannot be nil")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}