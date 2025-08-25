package lambda

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetFunctionE follows TDD principles - comprehensive coverage
func TestGetFunctionE(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		mockResponse   *lambda.GetFunctionOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, config *FunctionConfiguration)
	}{
		{
			name:         "should_retrieve_function_configuration_successfully",
			functionName: "test-function",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimeNodejs18x,
					Handler:      aws.String("index.handler"),
					Description:  aws.String("Test function"),
					Timeout:      aws.Int32(30),
					MemorySize:   aws.Int32(128),
					LastModified: aws.String("2023-01-01T00:00:00.000+0000"),
					Role:         aws.String("arn:aws:iam::123456789012:role/lambda-role"),
					State:        types.StateActive,
					StateReason:  aws.String("The function is active"),
					Version:      aws.String("$LATEST"),
					Environment: &types.EnvironmentResponse{
						Variables: map[string]string{
							"NODE_ENV": "production",
							"DEBUG":    "false",
						},
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, config *FunctionConfiguration) {
				assert.Equal(t, "test-function", config.FunctionName)
				assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:test-function", config.FunctionArn)
				assert.Equal(t, types.RuntimeNodejs18x, config.Runtime)
				assert.Equal(t, "index.handler", config.Handler)
				assert.Equal(t, "Test function", config.Description)
				assert.Equal(t, int32(30), config.Timeout)
				assert.Equal(t, int32(128), config.MemorySize)
				assert.Equal(t, "2023-01-01T00:00:00.000+0000", config.LastModified)
				assert.Equal(t, "arn:aws:iam::123456789012:role/lambda-role", config.Role)
				assert.Equal(t, types.StateActive, config.State)
				assert.Equal(t, "The function is active", config.StateReason)
				assert.Equal(t, "$LATEST", config.Version)
				assert.Equal(t, "production", config.Environment["NODE_ENV"])
				assert.Equal(t, "false", config.Environment["DEBUG"])
			},
		},
		{
			name:          "should_return_error_for_empty_function_name",
			functionName:  "",
			expectedError: true,
			errorContains: "invalid function name",
		},
		{
			name:          "should_return_error_for_invalid_function_name",
			functionName:  "invalid@function",
			expectedError: true,
			errorContains: "invalid character '@'",
		},
		{
			name:          "should_return_error_when_aws_api_fails",
			functionName:  "test-function",
			mockError:     errors.New("ResourceNotFoundException: Function not found"),
			expectedError: true,
			errorContains: "failed to get function configuration",
		},
		{
			name:         "should_handle_function_with_no_environment_variables",
			functionName: "test-function",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimePython39,
					Handler:      aws.String("lambda_function.lambda_handler"),
					Timeout:      aws.Int32(60),
					MemorySize:   aws.Int32(256),
					State:        types.StateActive,
					// No Environment field - should handle gracefully
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, config *FunctionConfiguration) {
				assert.Equal(t, "test-function", config.FunctionName)
				assert.Equal(t, types.RuntimePython39, config.Runtime)
				assert.Equal(t, "lambda_function.lambda_handler", config.Handler)
				assert.Equal(t, int32(60), config.Timeout)
				assert.Equal(t, int32(256), config.MemorySize)
				assert.Nil(t, config.Environment) // Should be nil when not present
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetFunctionResponse: tc.mockResponse,
				GetFunctionError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{
				T: &mockTestingT{},
			}

			// Act
			result, err := GetFunctionE(ctx, tc.functionName)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}

			// Verify mock interactions
			if !tc.expectedError && tc.functionName != "" && !strings.Contains(tc.functionName, "@") {
				assert.Contains(t, mockClient.GetFunctionCalls, tc.functionName)
			}
		})
	}
}

func TestGetFunction(t *testing.T) {
	t.Run("should_call_get_function_e_and_fail_test_on_error", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			GetFunctionError: errors.New("function not found"),
		}
		mockFactory := &MockClientFactory{
			LambdaClient: mockClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		mockT := &mockTestingT{}
		ctx := &TestContext{T: mockT}

		// Act
		result := GetFunction(ctx, "test-function")

		// Assert
		assert.Nil(t, result)
		assert.True(t, mockT.failNowCalled)
		assert.True(t, mockT.errorCalled)
	})
}

func TestFunctionExistsE(t *testing.T) {
	testCases := []struct {
		name         string
		functionName string
		mockResponse *lambda.GetFunctionOutput
		mockError    error
		expected     bool
		expectError  bool
	}{
		{
			name:         "should_return_true_when_function_exists",
			functionName: "existing-function",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("existing-function"),
					Runtime:      types.RuntimeNodejs18x,
				},
			},
			expected: true,
		},
		{
			name:         "should_return_false_when_function_not_found",
			functionName: "non-existent-function",
			mockError:    errors.New("ResourceNotFoundException: Function not found"),
			expected:     false,
		},
		{
			name:         "should_return_false_for_function_not_found_error",
			functionName: "non-existent-function",
			mockError:    errors.New("Function not found: arn:aws:lambda:us-east-1:123456789012:function:non-existent-function"),
			expected:     false,
		},
		{
			name:         "should_return_error_for_other_aws_errors",
			functionName: "test-function",
			mockError:    errors.New("AccessDeniedException: User is not authorized"),
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetFunctionResponse: tc.mockResponse,
				GetFunctionError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := FunctionExistsE(ctx, tc.functionName)

			// Assert
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestFunctionExists(t *testing.T) {
	t.Run("should_call_function_exists_e_and_fail_test_on_error", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			GetFunctionError: errors.New("AccessDeniedException: User is not authorized"),
		}
		mockFactory := &MockClientFactory{
			LambdaClient: mockClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		mockT := &mockTestingT{}
		ctx := &TestContext{T: mockT}

		// Act
		result := FunctionExists(ctx, "test-function")

		// Assert
		assert.False(t, result)
		assert.True(t, mockT.failNowCalled)
		assert.True(t, mockT.errorCalled)
	})
}

func TestWaitForFunctionActiveE(t *testing.T) {
	t.Run("should_return_immediately_when_function_is_active", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			GetFunctionResponse: &lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					State:        types.StateActive,
				},
			},
		}
		mockFactory := &MockClientFactory{
			LambdaClient: mockClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		err := WaitForFunctionActiveE(ctx, "test-function", 5*time.Second)

		// Assert
		assert.NoError(t, err)
		assert.Contains(t, mockClient.GetFunctionCalls, "test-function")
	})

	t.Run("should_timeout_when_function_never_becomes_active", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			GetFunctionResponse: &lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					State:        types.StatePending,
				},
			},
		}
		mockFactory := &MockClientFactory{
			LambdaClient: mockClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		start := time.Now()
		err := WaitForFunctionActiveE(ctx, "test-function", 100*time.Millisecond)
		duration := time.Since(start)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout waiting for function")
		assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
		assert.Less(t, duration, 500*time.Millisecond) // Should not take too long
	})

	t.Run("should_return_error_when_get_function_fails", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			GetFunctionError: errors.New("AccessDeniedException: User is not authorized"),
		}
		mockFactory := &MockClientFactory{
			LambdaClient: mockClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		err := WaitForFunctionActiveE(ctx, "test-function", 5*time.Second)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AccessDeniedException")
	})
}

func TestUpdateFunctionE(t *testing.T) {
	testCases := []struct {
		name           string
		config         UpdateFunctionConfig
		mockResponse   *lambda.UpdateFunctionConfigurationOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, config *FunctionConfiguration)
	}{
		{
			name: "should_update_function_configuration_successfully",
			config: UpdateFunctionConfig{
				FunctionName: "test-function",
				Runtime:      types.RuntimePython39,
				Handler:      "new_handler.handler",
				Description:  "Updated description",
				Timeout:      60,
				MemorySize:   256,
				Environment: map[string]string{
					"NEW_VAR": "new_value",
				},
			},
			mockResponse: &lambda.UpdateFunctionConfigurationOutput{
				FunctionName: aws.String("test-function"),
				Runtime:      types.RuntimePython39,
				Handler:      aws.String("new_handler.handler"),
				Description:  aws.String("Updated description"),
				Timeout:      aws.Int32(60),
				MemorySize:   aws.Int32(256),
				Environment: &types.EnvironmentResponse{
					Variables: map[string]string{
						"NEW_VAR": "new_value",
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, config *FunctionConfiguration) {
				assert.Equal(t, "test-function", config.FunctionName)
				assert.Equal(t, types.RuntimePython39, config.Runtime)
				assert.Equal(t, "new_handler.handler", config.Handler)
				assert.Equal(t, "Updated description", config.Description)
				assert.Equal(t, int32(60), config.Timeout)
				assert.Equal(t, int32(256), config.MemorySize)
				assert.Equal(t, "new_value", config.Environment["NEW_VAR"])
			},
		},
		{
			name: "should_return_error_for_invalid_function_name",
			config: UpdateFunctionConfig{
				FunctionName: "invalid@name",
			},
			expectedError: true,
			errorContains: "invalid character '@'",
		},
		{
			name: "should_return_error_when_aws_update_fails",
			config: UpdateFunctionConfig{
				FunctionName: "test-function",
				Runtime:      types.RuntimeNodejs18x,
			},
			mockError:     errors.New("InvalidParameterValueException: Invalid runtime"),
			expectedError: true,
			errorContains: "failed to update function configuration",
		},
		{
			name: "should_handle_partial_update_configuration",
			config: UpdateFunctionConfig{
				FunctionName: "test-function",
				Timeout:      90, // Only updating timeout
			},
			mockResponse: &lambda.UpdateFunctionConfigurationOutput{
				FunctionName: aws.String("test-function"),
				Timeout:      aws.Int32(90),
				MemorySize:   aws.Int32(128), // Original values preserved
			},
			expectedError: false,
			validateResult: func(t *testing.T, config *FunctionConfiguration) {
				assert.Equal(t, "test-function", config.FunctionName)
				assert.Equal(t, int32(90), config.Timeout)
				assert.Equal(t, int32(128), config.MemorySize)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				UpdateFunctionConfigurationResponse: tc.mockResponse,
				UpdateFunctionConfigurationError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := UpdateFunctionE(ctx, tc.config)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}

			// Verify mock interactions
			if !tc.expectedError && tc.config.FunctionName != "" && !strings.Contains(tc.config.FunctionName, "@") {
				assert.Contains(t, mockClient.UpdateFunctionConfigurationCalls, tc.config.FunctionName)
			}
		})
	}
}

func TestGetEnvVarE(t *testing.T) {
	testCases := []struct {
		name          string
		functionName  string
		varName       string
		mockResponse  *lambda.GetFunctionOutput
		mockError     error
		expectedValue string
		expectedError bool
		errorContains string
	}{
		{
			name:         "should_retrieve_environment_variable_successfully",
			functionName: "test-function",
			varName:      "NODE_ENV",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Environment: &types.EnvironmentResponse{
						Variables: map[string]string{
							"NODE_ENV": "production",
							"DEBUG":    "false",
						},
					},
				},
			},
			expectedValue: "production",
			expectedError: false,
		},
		{
			name:         "should_return_error_when_variable_not_found",
			functionName: "test-function",
			varName:      "MISSING_VAR",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Environment: &types.EnvironmentResponse{
						Variables: map[string]string{
							"NODE_ENV": "production",
						},
					},
				},
			},
			expectedError: true,
			errorContains: "environment variable MISSING_VAR not found",
		},
		{
			name:         "should_return_error_when_no_environment_variables",
			functionName: "test-function",
			varName:      "ANY_VAR",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					// No Environment field
				},
			},
			expectedError: true,
			errorContains: "function test-function has no environment variables",
		},
		{
			name:          "should_return_error_when_get_function_fails",
			functionName:  "test-function",
			varName:       "NODE_ENV",
			mockError:     errors.New("ResourceNotFoundException: Function not found"),
			expectedError: true,
			errorContains: "ResourceNotFoundException",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetFunctionResponse: tc.mockResponse,
				GetFunctionError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := GetEnvVarE(ctx, tc.functionName, tc.varName)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedValue, result)
			}
		})
	}
}

func TestListFunctionsE(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   *lambda.ListFunctionsOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, functions []FunctionConfiguration)
	}{
		{
			name: "should_list_functions_successfully",
			mockResponse: &lambda.ListFunctionsOutput{
				Functions: []types.FunctionConfiguration{
					{
						FunctionName: aws.String("function1"),
						Runtime:      types.RuntimeNodejs18x,
						Handler:      aws.String("index.handler"),
						MemorySize:   aws.Int32(128),
					},
					{
						FunctionName: aws.String("function2"),
						Runtime:      types.RuntimePython39,
						Handler:      aws.String("lambda_function.lambda_handler"),
						MemorySize:   aws.Int32(256),
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, functions []FunctionConfiguration) {
				assert.Len(t, functions, 2)
				assert.Equal(t, "function1", functions[0].FunctionName)
				assert.Equal(t, types.RuntimeNodejs18x, functions[0].Runtime)
				assert.Equal(t, "function2", functions[1].FunctionName)
				assert.Equal(t, types.RuntimePython39, functions[1].Runtime)
			},
		},
		{
			name: "should_return_empty_list_when_no_functions",
			mockResponse: &lambda.ListFunctionsOutput{
				Functions: []types.FunctionConfiguration{},
			},
			expectedError: false,
			validateResult: func(t *testing.T, functions []FunctionConfiguration) {
				assert.Len(t, functions, 0)
			},
		},
		{
			name:          "should_return_error_when_list_functions_fails",
			mockError:     errors.New("AccessDeniedException: User is not authorized"),
			expectedError: true,
			errorContains: "failed to list functions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				ListFunctionsResponse: tc.mockResponse,
				ListFunctionsError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := ListFunctionsE(ctx)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}

			// Verify mock interactions
			if !tc.expectedError {
				assert.Equal(t, 1, mockClient.ListFunctionsCalls)
			}
		})
	}
}

func TestCreateFunctionE(t *testing.T) {
	testCases := []struct {
		name           string
		input          *lambda.CreateFunctionInput
		mockResponse   *lambda.CreateFunctionOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, config *FunctionConfiguration)
	}{
		{
			name: "should_create_function_successfully",
			input: &lambda.CreateFunctionInput{
				FunctionName: aws.String("new-function"),
				Runtime:      types.RuntimePython39,
				Handler:      aws.String("lambda_function.lambda_handler"),
				Role:         aws.String("arn:aws:iam::123456789012:role/lambda-role"),
				Code: &types.FunctionCode{
					ZipFile: []byte("dummy zip content"),
				},
			},
			mockResponse: &lambda.CreateFunctionOutput{
				FunctionName: aws.String("new-function"),
				FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:new-function"),
				Runtime:      types.RuntimePython39,
				Handler:      aws.String("lambda_function.lambda_handler"),
				Role:         aws.String("arn:aws:iam::123456789012:role/lambda-role"),
				State:        types.StatePending,
				Version:      aws.String("1"),
			},
			expectedError: false,
			validateResult: func(t *testing.T, config *FunctionConfiguration) {
				assert.Equal(t, "new-function", config.FunctionName)
				assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:new-function", config.FunctionArn)
				assert.Equal(t, types.RuntimePython39, config.Runtime)
				assert.Equal(t, "lambda_function.lambda_handler", config.Handler)
				assert.Equal(t, "arn:aws:iam::123456789012:role/lambda-role", config.Role)
				assert.Equal(t, types.StatePending, config.State)
				assert.Equal(t, "1", config.Version)
			},
		},
		{
			name: "should_return_error_for_invalid_function_name",
			input: &lambda.CreateFunctionInput{
				FunctionName: aws.String("invalid@name"),
			},
			expectedError: true,
			errorContains: "invalid character '@'",
		},
		{
			name: "should_return_error_when_create_function_fails",
			input: &lambda.CreateFunctionInput{
				FunctionName: aws.String("new-function"),
				Runtime:      types.RuntimePython39,
			},
			mockError:     errors.New("InvalidParameterValueException: The role defined for the function cannot be assumed by Lambda"),
			expectedError: true,
			errorContains: "failed to create function",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				CreateFunctionResponse: tc.mockResponse,
				CreateFunctionError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := CreateFunctionE(ctx, tc.input)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}

			// Verify mock interactions
			if !tc.expectedError && tc.input != nil && tc.input.FunctionName != nil && !strings.Contains(*tc.input.FunctionName, "@") {
				assert.Contains(t, mockClient.CreateFunctionCalls, *tc.input.FunctionName)
			}
		})
	}
}

func TestDeleteFunctionE(t *testing.T) {
	testCases := []struct {
		name          string
		functionName  string
		mockError     error
		expectedError bool
		errorContains string
	}{
		{
			name:          "should_delete_function_successfully",
			functionName:  "test-function",
			expectedError: false,
		},
		{
			name:          "should_return_error_for_invalid_function_name",
			functionName:  "invalid@name",
			expectedError: true,
			errorContains: "invalid character '@'",
		},
		{
			name:          "should_return_error_when_delete_function_fails",
			functionName:  "test-function",
			mockError:     errors.New("ResourceNotFoundException: Function not found"),
			expectedError: true,
			errorContains: "failed to delete function",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				DeleteFunctionError: tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			err := DeleteFunctionE(ctx, tc.functionName)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			// Verify mock interactions
			if !tc.expectedError && tc.functionName != "" && !strings.Contains(tc.functionName, "@") {
				assert.Contains(t, mockClient.DeleteFunctionCalls, tc.functionName)
			}
		})
	}
}

func TestValidateLayerName(t *testing.T) {
	testCases := []struct {
		name         string
		layerName    string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "should_accept_valid_layer_name",
			layerName:   "my-layer",
			expectError: false,
		},
		{
			name:        "should_accept_layer_name_with_underscores",
			layerName:   "my_layer_123",
			expectError: false,
		},
		{
			name:         "should_reject_empty_layer_name",
			layerName:    "",
			expectError:  true,
			errorMessage: "empty name",
		},
		{
			name:         "should_reject_layer_name_too_long",
			layerName:    strings.Repeat("a", 65),
			expectError:  true,
			errorMessage: "name too long",
		},
		{
			name:         "should_reject_layer_name_with_invalid_characters",
			layerName:    "my-layer@invalid",
			expectError:  true,
			errorMessage: "invalid character '@'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateLayerName(tc.layerName)

			// Assert
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Mock testing interface for testing purposes
type mockTestingT struct {
	errorCalled   bool
	failNowCalled bool
	messages      []string
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errorCalled = true
}

func (m *mockTestingT) Error(args ...interface{}) {
	m.errorCalled = true
}

func (m *mockTestingT) Fail() {}

func (m *mockTestingT) FailNow() {
	m.failNowCalled = true
}

func (m *mockTestingT) Helper() {}

func (m *mockTestingT) Fatal(args ...interface{}) {}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {}

func (m *mockTestingT) Name() string {
	return "mockTest"
}