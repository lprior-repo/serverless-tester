package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Setup helper function that configures mock clients for a test
func setupMockAssertionTest() (*TestContext, *MockLambdaClient, *MockClientFactory) {
	mockLambdaClient := &MockLambdaClient{}
	mockCloudWatchClient := &MockCloudWatchLogsClient{}
	
	mockFactory := &MockClientFactory{
		LambdaClient:     mockLambdaClient,
		CloudWatchClient: mockCloudWatchClient,
	}
	
	// Temporarily replace the global client factory
	SetClientFactory(mockFactory)
	
	// Create test context with mock TestingT
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	
	return ctx, mockLambdaClient, mockFactory
}

// Teardown helper function to restore original factory
func teardownMockAssertionTest() {
	SetClientFactory(&DefaultClientFactory{})
}

// RED: Test AssertFunctionExists with existing function using mocks
func TestAssertFunctionExists_WithExistingFunction_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with existing function
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return successful function response
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("existing-function"),
			Runtime:      types.RuntimeNodejs18x,
			Handler:      aws.String("index.handler"),
			State:        types.StateActive,
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionExists is called
	AssertFunctionExists(ctx, "existing-function")
	
	// Then: Should not fail and should have called GetFunction
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "existing-function")
}

// RED: Test AssertFunctionExists with non-existing function using mocks  
func TestAssertFunctionExists_WithNonExistingFunction_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with non-existing function
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return ResourceNotFoundException
	mockClient.GetFunctionResponse = nil
	mockClient.GetFunctionError = createResourceNotFoundError("non-existing-function")
	
	// When: AssertFunctionExists is called
	AssertFunctionExists(ctx, "non-existing-function")
	
	// Then: Should fail with appropriate error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected Lambda function 'non-existing-function' to exist, but it does not")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "non-existing-function")
}

// RED: Test AssertFunctionExists with API error using mocks
func TestAssertFunctionExists_WithAPIError_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with API error
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return access denied error
	mockClient.GetFunctionResponse = nil
	mockClient.GetFunctionError = createAccessDeniedError("GetFunction")
	
	// When: AssertFunctionExists is called
	AssertFunctionExists(ctx, "test-function")
	
	// Then: Should fail with API error (expecting 2 errors since both the API call fails and the function doesn't exist)
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 2)
	assert.Contains(t, mockT.errors[0], "Failed to check if function exists")
	assert.Contains(t, mockT.errors[1], "Expected Lambda function 'test-function' to exist, but it does not")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionDoesNotExist with non-existing function using mocks
func TestAssertFunctionDoesNotExist_WithNonExistingFunction_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with non-existing function
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return ResourceNotFoundException
	mockClient.GetFunctionResponse = nil
	mockClient.GetFunctionError = createResourceNotFoundError("non-existing-function")
	
	// When: AssertFunctionDoesNotExist is called
	AssertFunctionDoesNotExist(ctx, "non-existing-function")
	
	// Then: Should not fail (ResourceNotFoundException is expected)
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "non-existing-function")
}

// RED: Test AssertFunctionDoesNotExist with existing function using mocks
func TestAssertFunctionDoesNotExist_WithExistingFunction_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with existing function
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return successful function response
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("existing-function"),
			Runtime:      types.RuntimeNodejs18x,
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionDoesNotExist is called
	AssertFunctionDoesNotExist(ctx, "existing-function")
	
	// Then: Should fail with appropriate error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected Lambda function 'existing-function' not to exist, but it does")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "existing-function")
}

// RED: Test AssertFunctionRuntime with matching runtime using mocks
func TestAssertFunctionRuntime_WithMatchingRuntime_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with function having expected runtime
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with nodejs18.x runtime
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Runtime:      types.RuntimeNodejs18x,
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionRuntime is called
	AssertFunctionRuntime(ctx, "test-function", types.RuntimeNodejs18x)
	
	// Then: Should not fail
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionRuntime with different runtime using mocks
func TestAssertFunctionRuntime_WithDifferentRuntime_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with function having different runtime
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with python3.9 runtime
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Runtime:      types.RuntimePython39,
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionRuntime is called expecting nodejs18.x
	AssertFunctionRuntime(ctx, "test-function", types.RuntimeNodejs18x)
	
	// Then: Should fail with runtime mismatch error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have runtime 'nodejs18.x', but got 'python3.9'")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionHandler with matching handler using mocks
func TestAssertFunctionHandler_WithMatchingHandler_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with function having expected handler
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with expected handler
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Handler:      aws.String("index.handler"),
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionHandler is called
	AssertFunctionHandler(ctx, "test-function", "index.handler")
	
	// Then: Should not fail
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionHandler with different handler using mocks
func TestAssertFunctionHandler_WithDifferentHandler_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with function having different handler
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with different handler
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Handler:      aws.String("app.main"),
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionHandler is called expecting different handler
	AssertFunctionHandler(ctx, "test-function", "index.handler")
	
	// Then: Should fail with handler mismatch error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have handler 'index.handler', but got 'app.main'")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionTimeout with matching timeout using mocks
func TestAssertFunctionTimeout_WithMatchingTimeout_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with function having expected timeout
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with 30 second timeout
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Timeout:      aws.Int32(30),
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionTimeout is called
	AssertFunctionTimeout(ctx, "test-function", 30)
	
	// Then: Should not fail
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionTimeout with different timeout using mocks
func TestAssertFunctionTimeout_WithDifferentTimeout_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with function having different timeout
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with 60 second timeout
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Timeout:      aws.Int32(60),
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionTimeout is called expecting 30 seconds
	AssertFunctionTimeout(ctx, "test-function", 30)
	
	// Then: Should fail with timeout mismatch error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have timeout 30 seconds, but got 60")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionMemorySize with matching memory size using mocks
func TestAssertFunctionMemorySize_WithMatchingMemorySize_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with function having expected memory size
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with 256 MB memory
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			MemorySize:   aws.Int32(256),
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionMemorySize is called
	AssertFunctionMemorySize(ctx, "test-function", 256)
	
	// Then: Should not fail
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionMemorySize with different memory size using mocks
func TestAssertFunctionMemorySize_WithDifferentMemorySize_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with function having different memory size
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with 512 MB memory
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			MemorySize:   aws.Int32(512),
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionMemorySize is called expecting 256 MB
	AssertFunctionMemorySize(ctx, "test-function", 256)
	
	// Then: Should fail with memory size mismatch error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to have memory size 256 MB, but got 512")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionState with matching state using mocks
func TestAssertFunctionState_WithMatchingState_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with function in expected state
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function in Active state
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			State:        types.StateActive,
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionState is called
	AssertFunctionState(ctx, "test-function", types.StateActive)
	
	// Then: Should not fail
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertFunctionState with different state using mocks
func TestAssertFunctionState_WithDifferentState_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with function in different state
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function in Pending state
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			State:        types.StatePending,
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertFunctionState is called expecting Active state
	AssertFunctionState(ctx, "test-function", types.StateActive)
	
	// Then: Should fail with state mismatch error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected function 'test-function' to be in state 'Active', but got 'Pending'")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertEnvVarEquals with matching environment variable using mocks
func TestAssertEnvVarEquals_WithMatchingEnvVar_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with function having expected environment variable
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with environment variables
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"NODE_ENV": "production",
					"DEBUG":    "false",
				},
			},
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertEnvVarEquals is called
	AssertEnvVarEquals(ctx, "test-function", "NODE_ENV", "production")
	
	// Then: Should not fail
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertEnvVarEquals with different environment variable value using mocks
func TestAssertEnvVarEquals_WithDifferentEnvVarValue_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with function having different environment variable value
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with environment variables
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"NODE_ENV": "development", // Different from expected
				},
			},
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertEnvVarEquals is called expecting production
	AssertEnvVarEquals(ctx, "test-function", "NODE_ENV", "production")
	
	// Then: Should fail with environment variable mismatch error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected environment variable 'NODE_ENV' in function 'test-function' to be 'production', but got 'development'")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertEnvVarExists with existing environment variable using mocks
func TestAssertEnvVarExists_WithExistingEnvVar_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with function having environment variable
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with environment variables
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"API_KEY": "secret-key",
				},
			},
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertEnvVarExists is called
	AssertEnvVarExists(ctx, "test-function", "API_KEY")
	
	// Then: Should not fail
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertEnvVarExists with missing environment variable using mocks
func TestAssertEnvVarExists_WithMissingEnvVar_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with function without the expected environment variable
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function without environment variables
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"OTHER_VAR": "other-value",
				},
			},
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertEnvVarExists is called for missing variable
	AssertEnvVarExists(ctx, "test-function", "MISSING_VAR")
	
	// Then: Should fail with environment variable missing error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected environment variable 'MISSING_VAR' to exist in function 'test-function', but it does not")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertEnvVarDoesNotExist with non-existing environment variable using mocks
func TestAssertEnvVarDoesNotExist_WithNonExistingEnvVar_UsingMocks_ShouldPass(t *testing.T) {
	// Given: Mock setup with function without the environment variable
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function without target environment variable
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"OTHER_VAR": "other-value",
				},
			},
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertEnvVarDoesNotExist is called
	AssertEnvVarDoesNotExist(ctx, "test-function", "SHOULD_NOT_EXIST")
	
	// Then: Should not fail
	mockT := ctx.T.(*mockAssertionTestingT)
	assert.Empty(t, mockT.errors)
	assert.False(t, mockT.failed)
	assert.False(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test AssertEnvVarDoesNotExist with existing environment variable using mocks
func TestAssertEnvVarDoesNotExist_WithExistingEnvVar_UsingMocks_ShouldFail(t *testing.T) {
	// Given: Mock setup with function having the environment variable
	ctx, mockClient, _ := setupMockAssertionTest()
	defer teardownMockAssertionTest()
	
	// Configure mock to return function with the environment variable
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"SECRET_KEY": "should-not-exist",
				},
			},
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: AssertEnvVarDoesNotExist is called for existing variable
	AssertEnvVarDoesNotExist(ctx, "test-function", "SECRET_KEY")
	
	// Then: Should fail with environment variable exists error
	mockT := ctx.T.(*mockAssertionTestingT)
	require.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "Expected environment variable 'SECRET_KEY' not to exist in function 'test-function', but it does")
	assert.True(t, mockT.failed)
	assert.True(t, mockT.failNow)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}