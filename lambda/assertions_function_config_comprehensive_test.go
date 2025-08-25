package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data factories for function configuration assertions

func createTestFunctionConfiguration() *types.FunctionConfiguration {
	return &types.FunctionConfiguration{
		FunctionName:  aws.String("test-function"),
		FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
		Runtime:       types.RuntimeNodejs18x,
		Handler:       aws.String("index.handler"),
		Description:   aws.String("Test function description"),
		Timeout:       aws.Int32(30),
		MemorySize:    aws.Int32(256),
		LastModified:  aws.String("2024-01-01T00:00:00.000+0000"),
		Role:          aws.String("arn:aws:iam::123456789012:role/test-role"),
		State:         types.StateActive,
		StateReason:   aws.String("Function is active"),
		Version:       aws.String("$LATEST"),
		Environment: &types.Environment{
			Variables: map[string]string{
				"NODE_ENV": "production",
				"API_KEY":  "test-api-key",
				"TIMEOUT":  "30",
			},
		},
	}
}

func createTestFunctionConfigurationInactive() *types.FunctionConfiguration {
	config := createTestFunctionConfiguration()
	config.State = types.StateInactive
	config.StateReason = aws.String("Function is inactive")
	return config
}

func createTestFunctionConfigurationPending() *types.FunctionConfiguration {
	config := createTestFunctionConfiguration()
	config.State = types.StatePending
	config.StateReason = aws.String("Function deployment is pending")
	return config
}

func createTestFunctionConfigurationFailed() *types.FunctionConfiguration {
	config := createTestFunctionConfiguration()
	config.State = types.StateFailed
	config.StateReason = aws.String("Function deployment failed")
	return config
}

// Tests for AssertFunctionExists

func TestAssertFunctionExists_WithExistingFunction_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response for existing function
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act
	AssertFunctionExists(ctx, "test-function")
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for existing function")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
	assert.Contains(t, mockLambda.GetFunctionCalls, "test-function", "Should have called GetFunction")
}

func TestAssertFunctionExists_WithNonExistingFunction_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response for non-existing function
	mockLambda.GetFunctionError = createResourceNotFoundError("non-existing-function")
	
	// Act
	AssertFunctionExists(ctx, "non-existing-function")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for non-existing function")
	assert.Contains(t, mockT.errors[0], "Expected Lambda function 'non-existing-function' to exist, but it does not")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertFunctionExists_WithAccessDeniedError_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response for access denied
	mockLambda.GetFunctionError = createAccessDeniedError("GetFunction")
	
	// Act
	AssertFunctionExists(ctx, "test-function")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for access denied")
	assert.Contains(t, mockT.errors[0], "Failed to check if function exists")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertFunctionDoesNotExist

func TestAssertFunctionDoesNotExist_WithNonExistingFunction_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response for non-existing function
	mockLambda.GetFunctionError = createResourceNotFoundError("non-existing-function")
	
	// Act
	AssertFunctionDoesNotExist(ctx, "non-existing-function")
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for non-existing function")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertFunctionDoesNotExist_WithExistingFunction_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response for existing function
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act
	AssertFunctionDoesNotExist(ctx, "test-function")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for existing function")
	assert.Contains(t, mockT.errors[0], "Expected Lambda function 'test-function' not to exist, but it does")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertFunctionDoesNotExist_WithAccessDeniedError_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response for access denied (not a ResourceNotFoundException)
	mockLambda.GetFunctionError = createAccessDeniedError("GetFunction")
	
	// Act
	AssertFunctionDoesNotExist(ctx, "test-function")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for access denied")
	assert.Contains(t, mockT.errors[0], "Failed to check if function exists")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertFunctionRuntime

func TestAssertFunctionRuntime_WithMatchingRuntime_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act
	AssertFunctionRuntime(ctx, "test-function", types.RuntimeNodejs18x)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for matching runtime")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertFunctionRuntime_WithDifferentRuntime_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with nodejs18.x runtime
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act - expect python3.9 but function has nodejs18.x
	AssertFunctionRuntime(ctx, "test-function", types.RuntimePython39)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for different runtime")
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have runtime 'python3.9', but got 'nodejs18.x'")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertFunctionRuntime_WithGetFunctionError_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with error
	mockLambda.GetFunctionError = createResourceNotFoundError("test-function")
	
	// Act
	AssertFunctionRuntime(ctx, "test-function", types.RuntimePython39)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for GetFunction failure")
	assert.Contains(t, mockT.errors[0], "Failed to get Lambda function configuration")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertFunctionHandler

func TestAssertFunctionHandler_WithMatchingHandler_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act
	AssertFunctionHandler(ctx, "test-function", "index.handler")
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for matching handler")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertFunctionHandler_WithDifferentHandler_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with index.handler
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act - expect app.handler but function has index.handler
	AssertFunctionHandler(ctx, "test-function", "app.handler")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for different handler")
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have handler 'app.handler', but got 'index.handler'")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertFunctionTimeout

func TestAssertFunctionTimeout_WithMatchingTimeout_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act
	AssertFunctionTimeout(ctx, "test-function", 30)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for matching timeout")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertFunctionTimeout_WithDifferentTimeout_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with 30 second timeout
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act - expect 60 seconds but function has 30 seconds
	AssertFunctionTimeout(ctx, "test-function", 60)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for different timeout")
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have timeout 60 seconds, but got 30")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertFunctionMemorySize

func TestAssertFunctionMemorySize_WithMatchingMemorySize_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act
	AssertFunctionMemorySize(ctx, "test-function", 256)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for matching memory size")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertFunctionMemorySize_WithDifferentMemorySize_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with 256MB
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act - expect 512MB but function has 256MB
	AssertFunctionMemorySize(ctx, "test-function", 512)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for different memory size")
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have memory size 512 MB, but got 256")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for AssertFunctionState - covering all possible Lambda function states

func TestAssertFunctionState_WithActiveState_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(), // Active by default
	}
	
	// Act
	AssertFunctionState(ctx, "test-function", types.StateActive)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for active state")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertFunctionState_WithInactiveState_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with inactive state
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfigurationInactive(),
	}
	
	// Act
	AssertFunctionState(ctx, "test-function", types.StateInactive)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for inactive state")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertFunctionState_WithPendingState_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with pending state
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfigurationPending(),
	}
	
	// Act
	AssertFunctionState(ctx, "test-function", types.StatePending)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for pending state")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertFunctionState_WithFailedState_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with failed state
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfigurationFailed(),
	}
	
	// Act
	AssertFunctionState(ctx, "test-function", types.StateFailed)
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for failed state")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertFunctionState_WithWrongState_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with active state
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(), // Active by default
	}
	
	// Act - expect inactive but function is active
	AssertFunctionState(ctx, "test-function", types.StateInactive)
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for wrong state")
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to be in state 'Inactive', but got 'Active'")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

// Tests for environment variable assertions

func TestAssertEnvVarEquals_WithMatchingEnvVar_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act
	AssertEnvVarEquals(ctx, "test-function", "NODE_ENV", "production")
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for matching env var")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertEnvVarEquals_WithDifferentEnvVar_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response with NODE_ENV=production
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act - expect development but env var is production
	AssertEnvVarEquals(ctx, "test-function", "NODE_ENV", "development")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for different env var value")
	assert.Contains(t, mockT.errors[0], "Expected environment variable 'NODE_ENV' in function 'test-function' to be 'development', but got 'production'")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertEnvVarExists_WithExistingEnvVar_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act
	AssertEnvVarExists(ctx, "test-function", "API_KEY")
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for existing env var")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertEnvVarExists_WithNonExistingEnvVar_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act - check for non-existing env var
	AssertEnvVarExists(ctx, "test-function", "NON_EXISTING_VAR")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for non-existing env var")
	assert.Contains(t, mockT.errors[0], "Expected environment variable 'NON_EXISTING_VAR' to exist in function 'test-function', but it does not")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}

func TestAssertEnvVarDoesNotExist_WithNonExistingEnvVar_UsingMocks_ShouldPass(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act
	AssertEnvVarDoesNotExist(ctx, "test-function", "NON_EXISTING_VAR")
	
	// Assert
	assert.Empty(t, mockT.errors, "Should not have any errors for non-existing env var")
	assert.False(t, mockT.failed, "Test should not have failed")
	assert.False(t, mockT.failNow, "Should not have called FailNow")
}

func TestAssertEnvVarDoesNotExist_WithExistingEnvVar_UsingMocks_ShouldFail(t *testing.T) {
	// Arrange
	ctx, mockLambda, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	mockT := &mockAssertionTestingT{}
	ctx.T = mockT
	
	// Set up mock response
	mockLambda.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: createTestFunctionConfiguration(),
	}
	
	// Act - check that existing env var doesn't exist
	AssertEnvVarDoesNotExist(ctx, "test-function", "API_KEY")
	
	// Assert
	require.NotEmpty(t, mockT.errors, "Should have errors for existing env var")
	assert.Contains(t, mockT.errors[0], "Expected environment variable 'API_KEY' not to exist in function 'test-function', but it does")
	assert.True(t, mockT.failed, "Test should have failed")
	assert.True(t, mockT.failNow, "Should have called FailNow")
}