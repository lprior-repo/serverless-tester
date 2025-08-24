package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test that AWS client mocking utilities work
func TestUtils_ShouldProvideAWSClientMocking(t *testing.T) {
	// Arrange
	utils := NewTestUtils()
	
	// Act
	lambdaClientStub := utils.CreateMockLambdaClient()
	dynamoClientStub := utils.CreateMockDynamoDBClient()
	
	// Assert
	require.NotNil(t, lambdaClientStub)
	require.NotNil(t, dynamoClientStub)
}

// RED: Test context creation helpers
func TestUtils_ShouldProvideContextCreationHelpers(t *testing.T) {
	// Arrange
	utils := NewTestUtils()
	
	// Act
	bgCtx := utils.CreateBackgroundContext()
	timeoutCtx := utils.CreateTimeoutContext(5 * time.Second)
	cancelCtx, cancel := utils.CreateCancelableContext()
	defer cancel()
	
	// Assert
	require.NotNil(t, bgCtx)
	require.NotNil(t, timeoutCtx)
	require.NotNil(t, cancelCtx)
	
	// Timeout context should have deadline
	deadline, hasDeadline := timeoutCtx.Deadline()
	assert.True(t, hasDeadline)
	assert.True(t, deadline.After(time.Now()))
}

// RED: Test that utils provide test data generation
func TestUtils_ShouldProvideTestDataGeneration(t *testing.T) {
	// Arrange
	utils := NewTestUtils()
	
	// Act
	randomString := utils.GenerateRandomString(10)
	randomNumber := utils.GenerateRandomNumber(1, 100)
	uuid := utils.GenerateUUID()
	
	// Assert
	assert.Len(t, randomString, 10)
	assert.True(t, randomNumber >= 1 && randomNumber <= 100)
	assert.Len(t, uuid, 36) // Standard UUID length with hyphens
}

// RED: Test assertion helpers for common validations
func TestUtils_ShouldProvideAssertionHelpers(t *testing.T) {
	// Arrange
	utils := NewTestUtils()
	mockT := NewMockTestingT("assertion-test")
	
	// Act - Test various assertion helpers
	utils.AssertStringNotEmpty(mockT, "test string", "string should not be empty")
	utils.AssertJSONValid(mockT, `{"key": "value"}`, "should be valid JSON")
	utils.AssertAWSArn(mockT, "arn:aws:lambda:us-east-1:123456789012:function:test-func", "should be valid ARN")
	
	// Assert - No errors should have occurred
	assert.False(t, mockT.WasErrorfCalled())
	assert.False(t, mockT.WasFailNowCalled())
}

// RED: Test assertion helpers with invalid data
func TestUtils_ShouldDetectInvalidDataInAssertions(t *testing.T) {
	// Arrange
	utils := NewTestUtils()
	mockT := NewMockTestingT("assertion-failure-test")
	
	// Act - Test with invalid data
	utils.AssertStringNotEmpty(mockT, "", "empty string should fail")
	utils.AssertJSONValid(mockT, `{invalid json}`, "invalid JSON should fail")
	utils.AssertAWSArn(mockT, "not-an-arn", "invalid ARN should fail")
	
	// Assert - Errors should have occurred
	assert.True(t, mockT.WasErrorfCalled())
	assert.Equal(t, 3, mockT.GetErrorfCallCount())
	
	errors := mockT.GetErrorMessages()
	assert.Contains(t, errors[0], "empty string should fail")
	assert.Contains(t, errors[1], "invalid JSON should fail")
	assert.Contains(t, errors[2], "invalid ARN should fail")
}

// RED: Test environment variable management
func TestUtils_ShouldProvideEnvironmentVariableManagement(t *testing.T) {
	// Arrange
	utils := NewTestUtils()
	
	// Act
	envManager := utils.CreateEnvironmentManager()
	originalValue := envManager.GetEnv("TEST_VAR", "default")
	
	// Set test environment variable
	envManager.SetEnv("TEST_VAR", "test-value")
	newValue := envManager.GetEnv("TEST_VAR", "default")
	
	// Restore original value
	envManager.RestoreEnv("TEST_VAR", originalValue)
	restoredValue := envManager.GetEnv("TEST_VAR", "default")
	
	// Assert
	assert.Equal(t, "test-value", newValue)
	assert.Equal(t, originalValue, restoredValue)
}

// RED: Test performance benchmarking helpers
func TestUtils_ShouldProvidePerformanceBenchmarking(t *testing.T) {
	// Arrange
	utils := NewTestUtils()
	
	// Act
	benchmark := utils.CreateBenchmark("test-operation")
	
	// Simulate some work
	time.Sleep(10 * time.Millisecond)
	duration := benchmark.Stop()
	
	// Assert
	assert.True(t, duration > 0)
	assert.True(t, duration >= 10*time.Millisecond)
}

// RED: Test retry utilities for flaky operations
func TestUtils_ShouldProvideRetryUtilities(t *testing.T) {
	// Arrange
	utils := NewTestUtils()
	attempts := 0
	
	// Act - Function that succeeds on third attempt
	err := utils.RetryOperation(func() error {
		attempts++
		if attempts < 3 {
			return assert.AnError
		}
		return nil
	}, 5, 10*time.Millisecond)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

// RED: Test retry utilities with permanent failure
func TestUtils_ShouldHandleRetryFailures(t *testing.T) {
	// Arrange
	utils := NewTestUtils()
	attempts := 0
	
	// Act - Function that always fails
	err := utils.RetryOperation(func() error {
		attempts++
		return assert.AnError
	}, 3, 1*time.Millisecond)
	
	// Assert
	assert.Error(t, err)
	assert.Equal(t, 3, attempts)
}