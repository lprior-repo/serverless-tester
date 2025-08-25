package lambda

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Setup helper for function tests
func setupMockFunctionTest() (*TestContext, *MockLambdaClient) {
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

// RED: Test GetFunctionE with existing function using mocks
func TestGetFunctionE_WithExistingFunction_UsingMocks_ShouldReturnConfiguration(t *testing.T) {
	// Given: Mock setup with existing function
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	expectedConfig := &types.FunctionConfiguration{
		FunctionName: aws.String("test-function"),
		Runtime:      types.RuntimeNodejs18x,
		Handler:      aws.String("index.handler"),
		Timeout:      aws.Int32(30),
		MemorySize:   aws.Int32(256),
		State:        types.StateActive,
	}
	
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: expectedConfig,
	}
	mockClient.GetFunctionError = nil
	
	// When: GetFunctionE is called
	config, err := GetFunctionE(ctx, "test-function")
	
	// Then: Should return configuration without error
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-function", config.FunctionName)
	assert.Equal(t, types.RuntimeNodejs18x, config.Runtime)
	assert.Equal(t, "index.handler", config.Handler)
	assert.Equal(t, int32(30), config.Timeout)
	assert.Equal(t, int32(256), config.MemorySize)
	assert.Equal(t, types.StateActive, config.State)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test GetFunctionE with non-existing function using mocks
func TestGetFunctionE_WithNonExistingFunction_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with non-existing function
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.GetFunctionResponse = nil
	mockClient.GetFunctionError = createResourceNotFoundError("non-existing-function")
	
	// When: GetFunctionE is called
	config, err := GetFunctionE(ctx, "non-existing-function")
	
	// Then: Should return error and nil config
	require.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "ResourceNotFoundException")
	assert.Contains(t, mockClient.GetFunctionCalls, "non-existing-function")
}

// RED: Test GetFunctionE with invalid function name using mocks
func TestGetFunctionE_WithInvalidFunctionName_UsingMocks_ShouldReturnValidationError(t *testing.T) {
	// Given: Mock setup
	ctx, _ := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	// When: GetFunctionE is called with empty function name
	config, err := GetFunctionE(ctx, "")
	
	// Then: Should return validation error without making API call
	require.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "invalid function name")
}

// RED: Test FunctionExistsE with existing function using mocks
func TestFunctionExistsE_WithExistingFunction_UsingMocks_ShouldReturnTrue(t *testing.T) {
	// Given: Mock setup with existing function
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("existing-function"),
			Runtime:      types.RuntimeNodejs18x,
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: FunctionExistsE is called
	exists, err := FunctionExistsE(ctx, "existing-function")
	
	// Then: Should return true without error
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Contains(t, mockClient.GetFunctionCalls, "existing-function")
}

// RED: Test FunctionExistsE with non-existing function using mocks
func TestFunctionExistsE_WithNonExistingFunction_UsingMocks_ShouldReturnFalse(t *testing.T) {
	// Given: Mock setup with non-existing function
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.GetFunctionResponse = nil
	mockClient.GetFunctionError = createResourceNotFoundError("non-existing-function")
	
	// When: FunctionExistsE is called
	exists, err := FunctionExistsE(ctx, "non-existing-function")
	
	// Then: Should return false without error (ResourceNotFoundException is expected)
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Contains(t, mockClient.GetFunctionCalls, "non-existing-function")
}

// RED: Test FunctionExistsE with API access error using mocks
func TestFunctionExistsE_WithAPIAccessError_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with access denied error
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.GetFunctionResponse = nil
	mockClient.GetFunctionError = createAccessDeniedError("GetFunction")
	
	// When: FunctionExistsE is called
	exists, err := FunctionExistsE(ctx, "test-function")
	
	// Then: Should return error (access denied should be propagated)
	require.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "AccessDeniedException")
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test UpdateFunctionE with valid configuration using mocks
func TestUpdateFunctionE_WithValidConfiguration_UsingMocks_ShouldUpdateAndReturn(t *testing.T) {
	// Given: Mock setup with successful update
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	updatedConfig := &lambda.UpdateFunctionConfigurationOutput{
		FunctionName: aws.String("test-function"),
		Runtime:      types.RuntimePython39,
		Handler:      aws.String("app.handler"),
		Timeout:      aws.Int32(60),
		MemorySize:   aws.Int32(512),
	}
	
	mockClient.UpdateFunctionConfigurationResponse = updatedConfig
	mockClient.UpdateFunctionConfigurationError = nil
	
	updateConfig := UpdateFunctionConfig{
		FunctionName: "test-function",
		Runtime:      types.RuntimePython39,
		Handler:      "app.handler",
		Timeout:      60,
		MemorySize:   512,
	}
	
	// When: UpdateFunctionE is called
	result, err := UpdateFunctionE(ctx, updateConfig)
	
	// Then: Should return updated configuration without error
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-function", result.FunctionName)
	assert.Equal(t, types.RuntimePython39, result.Runtime)
	assert.Equal(t, "app.handler", result.Handler)
	assert.Equal(t, int32(60), result.Timeout)
	assert.Equal(t, int32(512), result.MemorySize)
	assert.Contains(t, mockClient.UpdateFunctionConfigurationCalls, "test-function")
}

// RED: Test UpdateFunctionE with invalid function name using mocks
func TestUpdateFunctionE_WithInvalidFunctionName_UsingMocks_ShouldReturnValidationError(t *testing.T) {
	// Given: Mock setup
	ctx, _ := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	updateConfig := UpdateFunctionConfig{
		FunctionName: "", // Invalid empty name
		Runtime:      types.RuntimeNodejs18x,
	}
	
	// When: UpdateFunctionE is called with invalid function name
	result, err := UpdateFunctionE(ctx, updateConfig)
	
	// Then: Should return validation error without making API call
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid function name")
}

// RED: Test ListFunctionsE with successful response using mocks
func TestListFunctionsE_WithSuccessfulResponse_UsingMocks_ShouldReturnFunctionList(t *testing.T) {
	// Given: Mock setup with list of functions
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	functions := []types.FunctionConfiguration{
		{
			FunctionName: aws.String("function-1"),
			Runtime:      types.RuntimeNodejs18x,
			Handler:      aws.String("index.handler"),
		},
		{
			FunctionName: aws.String("function-2"),
			Runtime:      types.RuntimePython39,
			Handler:      aws.String("app.handler"),
		},
	}
	
	mockClient.ListFunctionsResponse = &lambda.ListFunctionsOutput{
		Functions:  functions,
		NextMarker: nil, // No more functions
	}
	mockClient.ListFunctionsError = nil
	
	// When: ListFunctionsE is called
	result, err := ListFunctionsE(ctx)
	
	// Then: Should return list of functions without error
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "function-1", result[0].FunctionName)
	assert.Equal(t, "function-2", result[1].FunctionName)
	assert.Equal(t, 1, mockClient.ListFunctionsCalls)
}

// RED: Test ListFunctionsE with API error using mocks
func TestListFunctionsE_WithAPIError_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with API error
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.ListFunctionsResponse = nil
	mockClient.ListFunctionsError = createAccessDeniedError("ListFunctions")
	
	// When: ListFunctionsE is called
	result, err := ListFunctionsE(ctx)
	
	// Then: Should return error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "AccessDeniedException")
	assert.Equal(t, 1, mockClient.ListFunctionsCalls)
}

// RED: Test GetEnvVarE with existing environment variable using mocks
func TestGetEnvVarE_WithExistingEnvVar_UsingMocks_ShouldReturnValue(t *testing.T) {
	// Given: Mock setup with function having environment variables
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"NODE_ENV":  "production",
					"LOG_LEVEL": "info",
				},
			},
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: GetEnvVarE is called
	value, err := GetEnvVarE(ctx, "test-function", "NODE_ENV")
	
	// Then: Should return environment variable value without error
	require.NoError(t, err)
	assert.Equal(t, "production", value)
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test GetEnvVarE with missing environment variable using mocks
func TestGetEnvVarE_WithMissingEnvVar_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with function having different environment variables
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
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
	
	// When: GetEnvVarE is called for missing variable
	value, err := GetEnvVarE(ctx, "test-function", "MISSING_VAR")
	
	// Then: Should return error
	require.Error(t, err)
	assert.Empty(t, value)
	assert.Contains(t, err.Error(), "environment variable MISSING_VAR not found")
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test GetEnvVarE with function having no environment variables using mocks
func TestGetEnvVarE_WithNoEnvVars_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with function having no environment variables
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.GetFunctionResponse = &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			Environment:  nil, // No environment variables
		},
	}
	mockClient.GetFunctionError = nil
	
	// When: GetEnvVarE is called
	value, err := GetEnvVarE(ctx, "test-function", "ANY_VAR")
	
	// Then: Should return error about no environment variables
	require.Error(t, err)
	assert.Empty(t, value)
	assert.Contains(t, err.Error(), "function test-function has no environment variables")
	assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
}

// RED: Test CreateFunctionE with valid input using mocks
func TestCreateFunctionE_WithValidInput_UsingMocks_ShouldCreateAndReturnConfig(t *testing.T) {
	// Given: Mock setup with successful creation
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	createdFunction := &lambda.CreateFunctionOutput{
		FunctionName: aws.String("new-function"),
		Runtime:      types.RuntimeNodejs18x,
		Handler:      aws.String("index.handler"),
		State:        types.StateActive,
	}
	
	mockClient.CreateFunctionResponse = createdFunction
	mockClient.CreateFunctionError = nil
	
	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String("new-function"),
		Runtime:      types.RuntimeNodejs18x,
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
		Role: aws.String("arn:aws:iam::123456789012:role/lambda-role"),
	}
	
	// When: CreateFunctionE is called
	config, err := CreateFunctionE(ctx, input)
	
	// Then: Should return function configuration without error
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "new-function", config.FunctionName)
	assert.Equal(t, types.RuntimeNodejs18x, config.Runtime)
	assert.Equal(t, "index.handler", config.Handler)
	assert.Equal(t, types.StateActive, config.State)
	assert.Contains(t, mockClient.CreateFunctionCalls, "new-function")
}

// RED: Test CreateFunctionE with invalid function name using mocks
func TestCreateFunctionE_WithInvalidFunctionName_UsingMocks_ShouldReturnValidationError(t *testing.T) {
	// Given: Mock setup
	ctx, _ := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String(""), // Invalid empty name
		Runtime:      types.RuntimeNodejs18x,
	}
	
	// When: CreateFunctionE is called with invalid function name
	config, err := CreateFunctionE(ctx, input)
	
	// Then: Should return validation error without making API call
	require.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "invalid function name")
}

// RED: Test DeleteFunctionE with existing function using mocks
func TestDeleteFunctionE_WithExistingFunction_UsingMocks_ShouldSucceed(t *testing.T) {
	// Given: Mock setup with successful deletion
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.DeleteFunctionResponse = &lambda.DeleteFunctionOutput{}
	mockClient.DeleteFunctionError = nil
	
	// When: DeleteFunctionE is called
	err := DeleteFunctionE(ctx, "function-to-delete")
	
	// Then: Should succeed without error
	require.NoError(t, err)
	assert.Contains(t, mockClient.DeleteFunctionCalls, "function-to-delete")
}

// RED: Test DeleteFunctionE with non-existing function using mocks
func TestDeleteFunctionE_WithNonExistingFunction_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with ResourceNotFoundException
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.DeleteFunctionResponse = nil
	mockClient.DeleteFunctionError = createResourceNotFoundError("non-existing-function")
	
	// When: DeleteFunctionE is called
	err := DeleteFunctionE(ctx, "non-existing-function")
	
	// Then: Should return error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ResourceNotFoundException")
	assert.Contains(t, mockClient.DeleteFunctionCalls, "non-existing-function")
}

// RED: Test DeleteFunctionE with invalid function name using mocks
func TestDeleteFunctionE_WithInvalidFunctionName_UsingMocks_ShouldReturnValidationError(t *testing.T) {
	// Given: Mock setup
	ctx, _ := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	// When: DeleteFunctionE is called with invalid function name
	err := DeleteFunctionE(ctx, "")
	
	// Then: Should return validation error without making API call
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid function name")
}

// RED: Test validateFunctionName with valid names
func TestValidateFunctionName_WithValidNames_ShouldReturnNil(t *testing.T) {
	validNames := []string{
		"valid-function",
		"ValidFunction",
		"function123",
		"my_function",
		"a",
		"1234567890123456789012345678901234567890123456789012345678901234", // 64 chars
	}
	
	for _, name := range validNames {
		t.Run(fmt.Sprintf("ValidName_%s", name), func(t *testing.T) {
			err := validateFunctionName(name)
			assert.NoError(t, err, "Function name should be valid: %s", name)
		})
	}
}

// RED: Test validateFunctionName with invalid names
func TestValidateFunctionName_WithInvalidNames_ShouldReturnError(t *testing.T) {
	invalidNames := map[string]string{
		"":                  "empty name",
		"invalid.function":  "contains dot",
		"invalid function":  "contains space",
		"invalid@function":  "contains at symbol",
		"12345678901234567890123456789012345678901234567890123456789012345": "too long (65 chars)",
	}
	
	for name, reason := range invalidNames {
		t.Run(fmt.Sprintf("InvalidName_%s", reason), func(t *testing.T) {
			err := validateFunctionName(name)
			assert.Error(t, err, "Function name should be invalid: %s (%s)", name, reason)
		})
	}
}

// RED: Test validatePayload with valid payloads
func TestValidatePayload_WithValidPayloads_ShouldReturnNil(t *testing.T) {
	validPayloads := []string{
		"",                           // Empty payload is valid
		"{}",                         // Empty JSON object
		`{"key": "value"}`,          // Simple JSON
		`{"nested": {"key": "value"}}`, // Nested JSON
		`[1, 2, 3]`,                 // JSON array
		`"simple string"`,           // JSON string
		`123`,                       // JSON number
		`true`,                      // JSON boolean
	}
	
	for _, payload := range validPayloads {
		t.Run(fmt.Sprintf("ValidPayload_%s", payload), func(t *testing.T) {
			err := validatePayload(payload)
			assert.NoError(t, err, "Payload should be valid: %s", payload)
		})
	}
}

// RED: Test validatePayload with invalid payloads
func TestValidatePayload_WithInvalidPayloads_ShouldReturnError(t *testing.T) {
	invalidPayloads := []string{
		`{invalid json}`,
		`{"unclosed": "object"`,
		`{"trailing": "comma",}`,
		`undefined`,
		`{broken json with "quotes"}`,
	}
	
	for _, payload := range invalidPayloads {
		t.Run(fmt.Sprintf("InvalidPayload_%s", payload), func(t *testing.T) {
			err := validatePayload(payload)
			assert.Error(t, err, "Payload should be invalid: %s", payload)
		})
	}
}

// RED: Test CreateFunctionE with environment variables using mocks
func TestCreateFunctionE_WithEnvironmentVariables_UsingMocks_ShouldCreateWithEnvVars(t *testing.T) {
	// Given: Mock setup with function creation including environment variables
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	createdFunction := &lambda.CreateFunctionOutput{
		FunctionName: aws.String("env-function"),
		Runtime:      types.RuntimeNodejs18x,
		Handler:      aws.String("index.handler"),
		State:        types.StateActive,
		Environment: &types.EnvironmentResponse{
			Variables: map[string]string{
				"NODE_ENV":  "production",
				"LOG_LEVEL": "info",
			},
		},
	}
	
	mockClient.CreateFunctionResponse = createdFunction
	mockClient.CreateFunctionError = nil
	
	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String("env-function"),
		Runtime:      types.RuntimeNodejs18x,
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
		Role: aws.String("arn:aws:iam::123456789012:role/lambda-role"),
		Environment: &types.Environment{
			Variables: map[string]string{
				"NODE_ENV":  "production",
				"LOG_LEVEL": "info",
			},
		},
	}
	
	// When: CreateFunctionE is called
	config, err := CreateFunctionE(ctx, input)
	
	// Then: Should return function configuration with environment variables
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "env-function", config.FunctionName)
	assert.NotNil(t, config.Environment)
	assert.Equal(t, "production", config.Environment["NODE_ENV"])
	assert.Equal(t, "info", config.Environment["LOG_LEVEL"])
	assert.Contains(t, mockClient.CreateFunctionCalls, "env-function")
}

// RED: Test CreateFunctionE with timeout and memory configuration using mocks
func TestCreateFunctionE_WithTimeoutAndMemoryConfig_UsingMocks_ShouldCreateWithCustomConfig(t *testing.T) {
	// Given: Mock setup with custom timeout and memory settings
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	createdFunction := &lambda.CreateFunctionOutput{
		FunctionName: aws.String("custom-config-function"),
		Runtime:      types.RuntimePython39,
		Handler:      aws.String("app.handler"),
		Timeout:      aws.Int32(300),
		MemorySize:   aws.Int32(1024),
		State:        types.StateActive,
	}
	
	mockClient.CreateFunctionResponse = createdFunction
	mockClient.CreateFunctionError = nil
	
	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String("custom-config-function"),
		Runtime:      types.RuntimePython39,
		Handler:      aws.String("app.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
		Role:       aws.String("arn:aws:iam::123456789012:role/lambda-role"),
		Timeout:    aws.Int32(300),
		MemorySize: aws.Int32(1024),
	}
	
	// When: CreateFunctionE is called
	config, err := CreateFunctionE(ctx, input)
	
	// Then: Should return function configuration with custom settings
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "custom-config-function", config.FunctionName)
	assert.Equal(t, types.RuntimePython39, config.Runtime)
	assert.Equal(t, int32(300), config.Timeout)
	assert.Equal(t, int32(1024), config.MemorySize)
	assert.Contains(t, mockClient.CreateFunctionCalls, "custom-config-function")
}

// RED: Test CreateFunctionE with dead letter queue configuration using mocks
func TestCreateFunctionE_WithDeadLetterQueue_UsingMocks_ShouldCreateWithDLQ(t *testing.T) {
	// Given: Mock setup with dead letter queue configuration
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	createdFunction := &lambda.CreateFunctionOutput{
		FunctionName: aws.String("dlq-function"),
		Runtime:      types.RuntimeJava11,
		Handler:      aws.String("com.example.Handler"),
		State:        types.StateActive,
		DeadLetterConfig: &types.DeadLetterConfig{
			TargetArn: aws.String("arn:aws:sqs:us-east-1:123456789012:dlq-queue"),
		},
	}
	
	mockClient.CreateFunctionResponse = createdFunction
	mockClient.CreateFunctionError = nil
	
	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String("dlq-function"),
		Runtime:      types.RuntimeJava11,
		Handler:      aws.String("com.example.Handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
		Role: aws.String("arn:aws:iam::123456789012:role/lambda-role"),
		DeadLetterConfig: &types.DeadLetterConfig{
			TargetArn: aws.String("arn:aws:sqs:us-east-1:123456789012:dlq-queue"),
		},
	}
	
	// When: CreateFunctionE is called
	config, err := CreateFunctionE(ctx, input)
	
	// Then: Should return function configuration with dead letter queue
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "dlq-function", config.FunctionName)
	assert.Equal(t, types.RuntimeJava11, config.Runtime)
	assert.Contains(t, mockClient.CreateFunctionCalls, "dlq-function")
}

// RED: Test CreateFunctionE with API error using mocks
func TestCreateFunctionE_WithAPIError_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with API error
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.CreateFunctionResponse = nil
	mockClient.CreateFunctionError = createAccessDeniedError("CreateFunction")
	
	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String("error-function"),
		Runtime:      types.RuntimeNodejs18x,
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: []byte("fake-zip-content"),
		},
		Role: aws.String("arn:aws:iam::123456789012:role/lambda-role"),
	}
	
	// When: CreateFunctionE is called
	config, err := CreateFunctionE(ctx, input)
	
	// Then: Should return error
	require.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "AccessDeniedException")
	assert.Contains(t, mockClient.CreateFunctionCalls, "error-function")
}

// RED: Test UpdateFunctionE with environment variables update using mocks
func TestUpdateFunctionE_WithEnvironmentVariablesUpdate_UsingMocks_ShouldUpdateEnvVars(t *testing.T) {
	// Given: Mock setup with environment variables update
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	updatedConfig := &lambda.UpdateFunctionConfigurationOutput{
		FunctionName: aws.String("env-update-function"),
		Runtime:      types.RuntimeNodejs18x,
		Handler:      aws.String("index.handler"),
		Environment: &types.EnvironmentResponse{
			Variables: map[string]string{
				"NODE_ENV":     "staging",
				"DEBUG_MODE":   "true",
				"API_ENDPOINT": "https://api-staging.example.com",
			},
		},
	}
	
	mockClient.UpdateFunctionConfigurationResponse = updatedConfig
	mockClient.UpdateFunctionConfigurationError = nil
	
	updateConfig := UpdateFunctionConfig{
		FunctionName: "env-update-function",
		Environment: map[string]string{
			"NODE_ENV":     "staging",
			"DEBUG_MODE":   "true",
			"API_ENDPOINT": "https://api-staging.example.com",
		},
	}
	
	// When: UpdateFunctionE is called
	result, err := UpdateFunctionE(ctx, updateConfig)
	
	// Then: Should return updated configuration with new environment variables
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "env-update-function", result.FunctionName)
	assert.NotNil(t, result.Environment)
	assert.Equal(t, "staging", result.Environment["NODE_ENV"])
	assert.Equal(t, "true", result.Environment["DEBUG_MODE"])
	assert.Equal(t, "https://api-staging.example.com", result.Environment["API_ENDPOINT"])
	assert.Contains(t, mockClient.UpdateFunctionConfigurationCalls, "env-update-function")
}

// RED: Test UpdateFunctionE with partial configuration update using mocks
func TestUpdateFunctionE_WithPartialUpdate_UsingMocks_ShouldUpdateOnlySpecifiedFields(t *testing.T) {
	// Given: Mock setup with partial update (only timeout)
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	updatedConfig := &lambda.UpdateFunctionConfigurationOutput{
		FunctionName: aws.String("partial-update-function"),
		Runtime:      types.RuntimeNodejs18x, // Existing runtime unchanged
		Handler:      aws.String("index.handler"), // Existing handler unchanged
		Timeout:      aws.Int32(120), // Only timeout updated
	}
	
	mockClient.UpdateFunctionConfigurationResponse = updatedConfig
	mockClient.UpdateFunctionConfigurationError = nil
	
	updateConfig := UpdateFunctionConfig{
		FunctionName: "partial-update-function",
		Timeout:      120, // Only update timeout
		// Other fields left empty/zero to test partial updates
	}
	
	// When: UpdateFunctionE is called
	result, err := UpdateFunctionE(ctx, updateConfig)
	
	// Then: Should return updated configuration with timeout changed
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "partial-update-function", result.FunctionName)
	assert.Equal(t, int32(120), result.Timeout)
	assert.Contains(t, mockClient.UpdateFunctionConfigurationCalls, "partial-update-function")
}

// RED: Test UpdateFunctionE with API error using mocks
func TestUpdateFunctionE_WithAPIError_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with API error
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.UpdateFunctionConfigurationResponse = nil
	mockClient.UpdateFunctionConfigurationError = createResourceNotFoundError("non-existing-function")
	
	updateConfig := UpdateFunctionConfig{
		FunctionName: "non-existing-function",
		Timeout:      60,
	}
	
	// When: UpdateFunctionE is called
	result, err := UpdateFunctionE(ctx, updateConfig)
	
	// Then: Should return error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ResourceNotFoundException")
	assert.Contains(t, mockClient.UpdateFunctionConfigurationCalls, "non-existing-function")
}