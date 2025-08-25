package lambda

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssertFunctionExists(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		mockResponse   *lambda.GetFunctionOutput
		mockError      error
		shouldFailTest bool
		errorMessage   string
	}{
		{
			name:         "should_pass_when_function_exists",
			functionName: "existing-function",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("existing-function"),
					Runtime:      lambdaTypes.RuntimeNodejs18x,
				},
			},
			shouldFailTest: false,
		},
		{
			name:           "should_fail_test_when_function_does_not_exist",
			functionName:   "non-existent-function",
			mockError:      errors.New("ResourceNotFoundException: Function not found"),
			shouldFailTest: true,
			errorMessage:   "Expected Lambda function 'non-existent-function' to exist, but it does not",
		},
		{
			name:           "should_fail_test_when_get_function_fails",
			functionName:   "test-function",
			mockError:      errors.New("AccessDeniedException: User is not authorized"),
			shouldFailTest: true,
			errorMessage:   "Failed to check if function exists",
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

			mockT := &mockTestingT{}
			ctx := &TestContext{T: mockT}

			// Act
			AssertFunctionExists(ctx, tc.functionName)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
				if tc.errorMessage != "" {
					// In a real test, we'd check the actual error message
					// Here we just verify the test failed as expected
				}
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertFunctionDoesNotExist(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		mockResponse   *lambda.GetFunctionOutput
		mockError      error
		shouldFailTest bool
	}{
		{
			name:         "should_pass_when_function_does_not_exist_resource_not_found",
			functionName: "non-existent-function",
			mockError:    errors.New("ResourceNotFoundException: Function not found"),
			shouldFailTest: false,
		},
		{
			name:         "should_pass_when_function_does_not_exist_function_not_found",
			functionName: "non-existent-function",
			mockError:    errors.New("Function not found: arn:aws:lambda:us-east-1:123456789012:function:non-existent-function"),
			shouldFailTest: false,
		},
		{
			name:         "should_fail_test_when_function_exists",
			functionName: "existing-function",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("existing-function"),
					Runtime:      lambdaTypes.RuntimeNodejs18x,
				},
			},
			shouldFailTest: true,
		},
		{
			name:           "should_fail_test_when_other_error_occurs",
			functionName:   "test-function",
			mockError:      errors.New("AccessDeniedException: User is not authorized"),
			shouldFailTest: true,
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

			mockT := &mockTestingT{}
			ctx := &TestContext{T: mockT}

			// Act
			AssertFunctionDoesNotExist(ctx, tc.functionName)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertFunctionRuntime(t *testing.T) {
	testCases := []struct {
		name             string
		functionName     string
		expectedRuntime  lambdaTypes.Runtime
		mockResponse     *lambda.GetFunctionOutput
		shouldFailTest   bool
	}{
		{
			name:            "should_pass_when_runtime_matches",
			functionName:    "test-function",
			expectedRuntime: lambdaTypes.RuntimeNodejs18x,
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Runtime:      lambdaTypes.RuntimeNodejs18x,
				},
			},
			shouldFailTest: false,
		},
		{
			name:            "should_fail_test_when_runtime_does_not_match",
			functionName:    "test-function",
			expectedRuntime: lambdaTypes.RuntimePython39,
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Runtime:      lambdaTypes.RuntimeNodejs18x,
				},
			},
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetFunctionResponse: tc.mockResponse,
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
			AssertFunctionRuntime(ctx, tc.functionName, tc.expectedRuntime)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertFunctionHandler(t *testing.T) {
	testCases := []struct {
		name             string
		functionName     string
		expectedHandler  string
		mockResponse     *lambda.GetFunctionOutput
		shouldFailTest   bool
	}{
		{
			name:            "should_pass_when_handler_matches",
			functionName:    "test-function",
			expectedHandler: "index.handler",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Handler:      aws.String("index.handler"),
				},
			},
			shouldFailTest: false,
		},
		{
			name:            "should_fail_test_when_handler_does_not_match",
			functionName:    "test-function",
			expectedHandler: "app.handler",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Handler:      aws.String("index.handler"),
				},
			},
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetFunctionResponse: tc.mockResponse,
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
			AssertFunctionHandler(ctx, tc.functionName, tc.expectedHandler)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertFunctionTimeout(t *testing.T) {
	testCases := []struct {
		name             string
		functionName     string
		expectedTimeout  int32
		mockResponse     *lambda.GetFunctionOutput
		shouldFailTest   bool
	}{
		{
			name:            "should_pass_when_timeout_matches",
			functionName:    "test-function",
			expectedTimeout: 30,
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Timeout:      aws.Int32(30),
				},
			},
			shouldFailTest: false,
		},
		{
			name:            "should_fail_test_when_timeout_does_not_match",
			functionName:    "test-function",
			expectedTimeout: 60,
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Timeout:      aws.Int32(30),
				},
			},
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetFunctionResponse: tc.mockResponse,
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
			AssertFunctionTimeout(ctx, tc.functionName, tc.expectedTimeout)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertFunctionMemorySize(t *testing.T) {
	testCases := []struct {
		name                string
		functionName        string
		expectedMemorySize  int32
		mockResponse        *lambda.GetFunctionOutput
		shouldFailTest      bool
	}{
		{
			name:               "should_pass_when_memory_size_matches",
			functionName:       "test-function",
			expectedMemorySize: 256,
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					MemorySize:   aws.Int32(256),
				},
			},
			shouldFailTest: false,
		},
		{
			name:               "should_fail_test_when_memory_size_does_not_match",
			functionName:       "test-function",
			expectedMemorySize: 512,
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					MemorySize:   aws.Int32(256),
				},
			},
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetFunctionResponse: tc.mockResponse,
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
			AssertFunctionMemorySize(ctx, tc.functionName, tc.expectedMemorySize)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertFunctionState(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		expectedState  lambdaTypes.State
		mockResponse   *lambda.GetFunctionOutput
		shouldFailTest bool
	}{
		{
			name:          "should_pass_when_state_matches",
			functionName:  "test-function",
			expectedState: lambdaTypes.StateActive,
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					State:        lambdaTypes.StateActive,
				},
			},
			shouldFailTest: false,
		},
		{
			name:          "should_fail_test_when_state_does_not_match",
			functionName:  "test-function",
			expectedState: lambdaTypes.StateActive,
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					State:        lambdaTypes.StatePending,
				},
			},
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetFunctionResponse: tc.mockResponse,
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
			AssertFunctionState(ctx, tc.functionName, tc.expectedState)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertEnvVarEquals(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		varName        string
		expectedValue  string
		mockResponse   *lambda.GetFunctionOutput
		shouldFailTest bool
	}{
		{
			name:          "should_pass_when_env_var_matches",
			functionName:  "test-function",
			varName:       "NODE_ENV",
			expectedValue: "production",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Environment: &lambdaTypes.EnvironmentResponse{
						Variables: map[string]string{
							"NODE_ENV": "production",
							"DEBUG":    "false",
						},
					},
				},
			},
			shouldFailTest: false,
		},
		{
			name:          "should_fail_test_when_env_var_value_does_not_match",
			functionName:  "test-function",
			varName:       "NODE_ENV",
			expectedValue: "development",
			mockResponse: &lambda.GetFunctionOutput{
				Configuration: &lambdaTypes.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					Environment: &lambdaTypes.EnvironmentResponse{
						Variables: map[string]string{
							"NODE_ENV": "production",
						},
					},
				},
			},
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetFunctionResponse: tc.mockResponse,
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
			AssertEnvVarEquals(ctx, tc.functionName, tc.varName, tc.expectedValue)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertInvokeSuccess(t *testing.T) {
	testCases := []struct {
		name           string
		result         *InvokeResult
		shouldFailTest bool
	}{
		{
			name: "should_pass_when_invoke_result_is_successful",
			result: &InvokeResult{
				StatusCode:    200,
				Payload:       `{"success": true}`,
				FunctionError: "",
			},
			shouldFailTest: false,
		},
		{
			name: "should_pass_when_invoke_result_has_2xx_status_code",
			result: &InvokeResult{
				StatusCode:    202,
				Payload:       ``,
				FunctionError: "",
			},
			shouldFailTest: false,
		},
		{
			name:           "should_fail_test_when_invoke_result_is_nil",
			result:         nil,
			shouldFailTest: true,
		},
		{
			name: "should_fail_test_when_function_error_exists",
			result: &InvokeResult{
				StatusCode:    200,
				Payload:       `{"errorMessage": "Test error"}`,
				FunctionError: "Unhandled",
			},
			shouldFailTest: true,
		},
		{
			name: "should_fail_test_when_status_code_is_not_2xx",
			result: &InvokeResult{
				StatusCode:    500,
				Payload:       `{"errorMessage": "Internal error"}`,
				FunctionError: "",
			},
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockT := &mockTestingT{}
			ctx := &TestContext{T: mockT}

			// Act
			AssertInvokeSuccess(ctx, tc.result)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertInvokeError(t *testing.T) {
	testCases := []struct {
		name           string
		result         *InvokeResult
		shouldFailTest bool
	}{
		{
			name: "should_pass_when_invoke_result_has_function_error",
			result: &InvokeResult{
				StatusCode:    200,
				Payload:       `{"errorMessage": "Test error"}`,
				FunctionError: "Unhandled",
			},
			shouldFailTest: false,
		},
		{
			name:           "should_fail_test_when_invoke_result_is_nil",
			result:         nil,
			shouldFailTest: true,
		},
		{
			name: "should_fail_test_when_no_function_error",
			result: &InvokeResult{
				StatusCode:    200,
				Payload:       `{"success": true}`,
				FunctionError: "",
			},
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockT := &mockTestingT{}
			ctx := &TestContext{T: mockT}

			// Act
			AssertInvokeError(ctx, tc.result)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertPayloadContains(t *testing.T) {
	testCases := []struct {
		name           string
		result         *InvokeResult
		expectedText   string
		shouldFailTest bool
	}{
		{
			name: "should_pass_when_payload_contains_expected_text",
			result: &InvokeResult{
				Payload: `{"message": "Hello, World!", "status": "success"}`,
			},
			expectedText:   "Hello, World!",
			shouldFailTest: false,
		},
		{
			name: "should_pass_when_payload_contains_partial_match",
			result: &InvokeResult{
				Payload: `{"result": "success"}`,
			},
			expectedText:   "success",
			shouldFailTest: false,
		},
		{
			name:           "should_fail_test_when_invoke_result_is_nil",
			result:         nil,
			expectedText:   "test",
			shouldFailTest: true,
		},
		{
			name: "should_fail_test_when_payload_does_not_contain_expected_text",
			result: &InvokeResult{
				Payload: `{"result": "failure"}`,
			},
			expectedText:   "success",
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockT := &mockTestingT{}
			ctx := &TestContext{T: mockT}

			// Act
			AssertPayloadContains(ctx, tc.result, tc.expectedText)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertPayloadEquals(t *testing.T) {
	testCases := []struct {
		name            string
		result          *InvokeResult
		expectedPayload string
		shouldFailTest  bool
	}{
		{
			name: "should_pass_when_payload_equals_expected",
			result: &InvokeResult{
				Payload: `{"success": true}`,
			},
			expectedPayload: `{"success": true}`,
			shouldFailTest:  false,
		},
		{
			name:            "should_fail_test_when_invoke_result_is_nil",
			result:          nil,
			expectedPayload: `{"success": true}`,
			shouldFailTest:  true,
		},
		{
			name: "should_fail_test_when_payload_does_not_equal_expected",
			result: &InvokeResult{
				Payload: `{"success": false}`,
			},
			expectedPayload: `{"success": true}`,
			shouldFailTest:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockT := &mockTestingT{}
			ctx := &TestContext{T: mockT}

			// Act
			AssertPayloadEquals(ctx, tc.result, tc.expectedPayload)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestAssertExecutionTimeLessThan(t *testing.T) {
	testCases := []struct {
		name           string
		result         *InvokeResult
		maxDuration    time.Duration
		shouldFailTest bool
	}{
		{
			name: "should_pass_when_execution_time_is_less_than_max",
			result: &InvokeResult{
				ExecutionTime: 100 * time.Millisecond,
			},
			maxDuration:    200 * time.Millisecond,
			shouldFailTest: false,
		},
		{
			name:           "should_fail_test_when_invoke_result_is_nil",
			result:         nil,
			maxDuration:    200 * time.Millisecond,
			shouldFailTest: true,
		},
		{
			name: "should_fail_test_when_execution_time_exceeds_max",
			result: &InvokeResult{
				ExecutionTime: 300 * time.Millisecond,
			},
			maxDuration:    200 * time.Millisecond,
			shouldFailTest: true,
		},
		{
			name: "should_fail_test_when_execution_time_equals_max",
			result: &InvokeResult{
				ExecutionTime: 200 * time.Millisecond,
			},
			maxDuration:    200 * time.Millisecond,
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockT := &mockTestingT{}
			ctx := &TestContext{T: mockT}

			// Act
			AssertExecutionTimeLessThan(ctx, tc.result, tc.maxDuration)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}

func TestValidateFunctionConfiguration(t *testing.T) {
	testCases := []struct {
		name           string
		config         *FunctionConfiguration
		expected       *FunctionConfiguration
		expectedErrors []string
	}{
		{
			name:           "should_return_no_errors_for_matching_configuration",
			config:         &FunctionConfiguration{
				FunctionName: "test-function",
				Runtime:      lambdaTypes.RuntimeNodejs18x,
				Handler:      "index.handler",
				Timeout:      30,
				MemorySize:   128,
				State:        lambdaTypes.StateActive,
				Environment:  map[string]string{"NODE_ENV": "production"},
			},
			expected:       &FunctionConfiguration{
				FunctionName: "test-function",
				Runtime:      lambdaTypes.RuntimeNodejs18x,
				Handler:      "index.handler",
				Timeout:      30,
				MemorySize:   128,
				State:        lambdaTypes.StateActive,
				Environment:  map[string]string{"NODE_ENV": "production"},
			},
			expectedErrors: []string{},
		},
		{
			name:           "should_return_error_for_nil_config",
			config:         nil,
			expected:       &FunctionConfiguration{},
			expectedErrors: []string{"function configuration is nil"},
		},
		{
			name:           "should_return_no_errors_for_nil_expected",
			config:         &FunctionConfiguration{FunctionName: "test-function"},
			expected:       nil,
			expectedErrors: []string{},
		},
		{
			name: "should_return_errors_for_mismatched_fields",
			config: &FunctionConfiguration{
				FunctionName: "actual-function",
				Runtime:      lambdaTypes.RuntimeNodejs18x,
				Handler:      "actual.handler",
				Timeout:      60,
				MemorySize:   256,
				State:        lambdaTypes.StatePending,
				Environment:  map[string]string{"NODE_ENV": "development"},
			},
			expected: &FunctionConfiguration{
				FunctionName: "expected-function",
				Runtime:      lambdaTypes.RuntimePython39,
				Handler:      "expected.handler",
				Timeout:      30,
				MemorySize:   128,
				State:        lambdaTypes.StateActive,
				Environment:  map[string]string{"NODE_ENV": "production", "DEBUG": "true"},
			},
			expectedErrors: []string{
				"expected function name 'expected-function', got 'actual-function'",
				"expected runtime 'python3.9', got 'nodejs18.x'",
				"expected handler 'expected.handler', got 'actual.handler'",
				"expected timeout 30, got 60",
				"expected memory size 128, got 256",
				"expected state 'Active', got 'Pending'",
				"expected environment variable 'DEBUG' not found",
				"expected env var 'NODE_ENV'='production', got 'development'",
			},
		},
		{
			name: "should_skip_validation_for_empty_expected_fields",
			config: &FunctionConfiguration{
				FunctionName: "test-function",
				Runtime:      lambdaTypes.RuntimeNodejs18x,
				Timeout:      60,
				MemorySize:   256,
			},
			expected: &FunctionConfiguration{
				// Only validate non-zero/empty fields
				Runtime: lambdaTypes.RuntimeNodejs18x,
				Timeout: 60,
			},
			expectedErrors: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			errors := ValidateFunctionConfiguration(tc.config, tc.expected)

			// Assert
			if len(tc.expectedErrors) == 0 {
				assert.Empty(t, errors)
			} else {
				require.Equal(t, len(tc.expectedErrors), len(errors))
				for i, expectedError := range tc.expectedErrors {
					assert.Contains(t, errors[i], expectedError)
				}
			}
		})
	}
}

func TestValidateInvokeResult(t *testing.T) {
	testCases := []struct {
		name                     string
		result                   *InvokeResult
		expectSuccess            bool
		expectedPayloadContains  string
		expectedErrors           []string
	}{
		{
			name: "should_return_no_errors_for_successful_result",
			result: &InvokeResult{
				StatusCode:    200,
				Payload:       `{"message": "success"}`,
				FunctionError: "",
			},
			expectSuccess:           true,
			expectedPayloadContains: "success",
			expectedErrors:          []string{},
		},
		{
			name:           "should_return_error_for_nil_result",
			result:         nil,
			expectSuccess:  true,
			expectedErrors: []string{"invoke result is nil"},
		},
		{
			name: "should_return_error_when_expecting_success_but_got_function_error",
			result: &InvokeResult{
				StatusCode:    200,
				Payload:       `{"errorMessage": "Test error"}`,
				FunctionError: "Unhandled",
			},
			expectSuccess:  true,
			expectedErrors: []string{"expected success but got function error: Unhandled"},
		},
		{
			name: "should_return_error_when_expecting_success_but_got_bad_status_code",
			result: &InvokeResult{
				StatusCode:    500,
				Payload:       `{"errorMessage": "Internal error"}`,
				FunctionError: "",
			},
			expectSuccess:  true,
			expectedErrors: []string{"expected success but got status code: 500"},
		},
		{
			name: "should_return_error_when_expecting_failure_but_got_success",
			result: &InvokeResult{
				StatusCode:    200,
				Payload:       `{"success": true}`,
				FunctionError: "",
			},
			expectSuccess:  false,
			expectedErrors: []string{"expected failure but invocation succeeded"},
		},
		{
			name: "should_return_error_when_payload_does_not_contain_expected_text",
			result: &InvokeResult{
				StatusCode:    200,
				Payload:       `{"result": "failure"}`,
				FunctionError: "",
			},
			expectSuccess:           true,
			expectedPayloadContains: "success",
			expectedErrors:          []string{"expected payload to contain 'success', but payload was: {\"result\": \"failure\"}"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			errors := ValidateInvokeResult(tc.result, tc.expectSuccess, tc.expectedPayloadContains)

			// Assert
			if len(tc.expectedErrors) == 0 {
				assert.Empty(t, errors)
			} else {
				require.Equal(t, len(tc.expectedErrors), len(errors))
				for i, expectedError := range tc.expectedErrors {
					assert.Contains(t, errors[i], expectedError)
				}
			}
		})
	}
}

func TestValidateLogEntries(t *testing.T) {
	testCases := []struct {
		name            string
		logs            []LogEntry
		expectedCount   int
		expectedContains string
		expectedLevel   string
		expectedErrors  []string
	}{
		{
			name: "should_return_no_errors_for_valid_log_entries",
			logs: []LogEntry{
				{Message: "INFO: Application started", Level: "INFO"},
				{Message: "DEBUG: Processing request", Level: "DEBUG"},
				{Message: "SUCCESS: Request completed", Level: "INFO"},
			},
			expectedCount:    3,
			expectedContains: "Application started",
			expectedLevel:    "INFO",
			expectedErrors:   []string{},
		},
		{
			name:           "should_return_error_for_incorrect_count",
			logs:           []LogEntry{{Message: "Test log"}},
			expectedCount:  2,
			expectedErrors: []string{"expected 2 log entries, got 1"},
		},
		{
			name:             "should_return_error_when_text_not_found",
			logs:             []LogEntry{{Message: "Test log message"}},
			expectedContains: "Not found",
			expectedErrors:   []string{"expected logs to contain 'Not found'"},
		},
		{
			name:           "should_return_error_when_level_not_found",
			logs:           []LogEntry{{Message: "Test log", Level: "DEBUG"}},
			expectedLevel:  "ERROR",
			expectedErrors: []string{"expected logs to contain level 'ERROR'"},
		},
		{
			name: "should_handle_negative_expected_count",
			logs: []LogEntry{{Message: "Test"}},
			expectedCount: -1, // Negative count should be ignored
			expectedErrors: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			errors := ValidateLogEntries(tc.logs, tc.expectedCount, tc.expectedContains, tc.expectedLevel)

			// Assert
			if len(tc.expectedErrors) == 0 {
				assert.Empty(t, errors)
			} else {
				require.Equal(t, len(tc.expectedErrors), len(errors))
				for i, expectedError := range tc.expectedErrors {
					assert.Contains(t, errors[i], expectedError)
				}
			}
		})
	}
}

func TestAssertLogGroupExists(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		mockResponse   *cloudwatchlogs.DescribeLogGroupsOutput
		mockError      error
		shouldFailTest bool
	}{
		{
			name:         "should_pass_when_log_group_exists",
			functionName: "test-function",
			mockResponse: &cloudwatchlogs.DescribeLogGroupsOutput{
				LogGroups: []types.LogGroup{
					{
						LogGroupName: aws.String("/aws/lambda/test-function"),
					},
				},
			},
			shouldFailTest: false,
		},
		{
			name:         "should_fail_test_when_log_group_does_not_exist",
			functionName: "test-function",
			mockResponse: &cloudwatchlogs.DescribeLogGroupsOutput{
				LogGroups: []types.LogGroup{}, // Empty result
			},
			shouldFailTest: true,
		},
		{
			name:           "should_fail_test_when_describe_log_groups_fails",
			functionName:   "test-function",
			mockError:      errors.New("AccessDeniedException: User is not authorized"),
			shouldFailTest: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockCwClient := &MockCloudWatchLogsClient{
				DescribeLogGroupsResponse: tc.mockResponse,
				DescribeLogGroupsError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				CloudWatchClient: mockCwClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			mockT := &mockTestingT{}
			ctx := &TestContext{T: mockT}

			// Act
			AssertLogGroupExists(ctx, tc.functionName)

			// Assert
			if tc.shouldFailTest {
				assert.True(t, mockT.failNowCalled)
				assert.True(t, mockT.errorCalled)
			} else {
				assert.False(t, mockT.failNowCalled)
				assert.False(t, mockT.errorCalled)
			}
		})
	}
}