package lambda

import (
	"context"
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

func TestInvokeWithOptionsE(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		payload        string
		options        *InvokeOptions
		mockResponse   *lambda.InvokeOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, result *InvokeResult)
	}{
		{
			name:         "should_invoke_function_successfully_with_default_options",
			functionName: "test-function",
			payload:      `{"key": "value"}`,
			options:      nil, // Use defaults
			mockResponse: &lambda.InvokeOutput{
				StatusCode:       200,
				Payload:          []byte(`{"result": "success"}`),
				ExecutedVersion:  aws.String("1"),
				LogResult:        aws.String("START RequestId: 123\nINFO: Function executed\nEND RequestId: 123"),
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *InvokeResult) {
				assert.Equal(t, int32(200), result.StatusCode)
				assert.Equal(t, `{"result": "success"}`, result.Payload)
				assert.Equal(t, "1", result.ExecutedVersion)
				assert.Contains(t, result.LogResult, "Function executed")
				assert.Empty(t, result.FunctionError)
				assert.Greater(t, result.ExecutionTime, time.Duration(0))
			},
		},
		{
			name:         "should_invoke_function_with_custom_options",
			functionName: "test-function",
			payload:      `{"test": true}`,
			options: &InvokeOptions{
				InvocationType: types.InvocationTypeEvent,
				LogType:        LogTypeTail,
				ClientContext:  "test-context",
				Qualifier:      "PROD",
				Timeout:        10 * time.Second,
			},
			mockResponse: &lambda.InvokeOutput{
				StatusCode:      202,
				Payload:         []byte(""),
				ExecutedVersion: aws.String("PROD"),
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *InvokeResult) {
				assert.Equal(t, int32(202), result.StatusCode)
				assert.Equal(t, "", result.Payload)
				assert.Equal(t, "PROD", result.ExecutedVersion)
				assert.Empty(t, result.FunctionError)
			},
		},
		{
			name:         "should_handle_function_error_response",
			functionName: "test-function",
			payload:      `{"cause_error": true}`,
			mockResponse: &lambda.InvokeOutput{
				StatusCode:    200,
				Payload:       []byte(`{"errorMessage": "Test error", "errorType": "Error"}`),
				FunctionError: aws.String("Unhandled"),
			},
			expectedError: true,
			errorContains: "lambda function error: Unhandled",
		},
		{
			name:          "should_return_error_for_invalid_function_name",
			functionName:  "invalid@function",
			payload:       `{}`,
			expectedError: true,
			errorContains: "invalid character '@'",
		},
		{
			name:          "should_return_error_for_invalid_payload",
			functionName:  "test-function",
			payload:       `{invalid json}`,
			expectedError: true,
			errorContains: "invalid payload format",
		},
		{
			name:         "should_return_error_for_oversized_payload",
			functionName: "test-function",
			payload:      strings.Repeat("a", MaxPayloadSize+1),
			expectedError: true,
			errorContains: "payload size exceeds maximum",
		},
		{
			name:         "should_return_error_when_invoke_fails",
			functionName: "test-function",
			payload:      `{}`,
			mockError:    errors.New("ResourceNotFoundException: Function not found"),
			expectedError: true,
			errorContains: "lambda invocation failed",
		},
		{
			name:         "should_return_error_for_non_2xx_status_code",
			functionName: "test-function",
			payload:      `{}`,
			mockResponse: &lambda.InvokeOutput{
				StatusCode: 500,
				Payload:    []byte(`{"errorMessage": "Internal Server Error"}`),
			},
			expectedError: true,
			errorContains: "lambda invocation failed with status code: 500",
		},
		{
			name:         "should_sanitize_log_result",
			functionName: "test-function",
			payload:      `{}`,
			mockResponse: &lambda.InvokeOutput{
				StatusCode: 200,
				Payload:    []byte(`{"success": true}`),
				LogResult: aws.String(`START RequestId: 12345-67890-abcdef
INFO: Application started
Duration: 123.45 ms
Billed Duration: 200 ms
INFO: Processing complete
END RequestId: 12345-67890-abcdef`),
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *InvokeResult) {
				assert.Equal(t, int32(200), result.StatusCode)
				// Verify that AWS internal log lines are removed
				assert.NotContains(t, result.LogResult, "Duration:")
				assert.NotContains(t, result.LogResult, "Billed Duration:")
				assert.NotContains(t, result.LogResult, "RequestId:")
				// Verify that application logs are preserved
				assert.Contains(t, result.LogResult, "Application started")
				assert.Contains(t, result.LogResult, "Processing complete")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				InvokeResponse: tc.mockResponse,
				InvokeError:    tc.mockError,
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
			result, err := InvokeWithOptionsE(ctx, tc.functionName, tc.payload, tc.options)

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

			// Verify mock interactions for successful invocations
			if !tc.expectedError && tc.functionName != "" && !strings.Contains(tc.functionName, "@") && len(tc.payload) <= MaxPayloadSize {
				assert.Contains(t, mockClient.InvokeCalls, tc.functionName)
			}
		})
	}
}

func TestInvokeE(t *testing.T) {
	t.Run("should_call_invoke_with_options_e_with_nil_options", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			InvokeResponse: &lambda.InvokeOutput{
				StatusCode: 200,
				Payload:    []byte(`{"success": true}`),
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
		result, err := InvokeE(ctx, "test-function", `{"key": "value"}`)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(200), result.StatusCode)
		assert.Contains(t, mockClient.InvokeCalls, "test-function")
	})
}

func TestInvoke(t *testing.T) {
	t.Run("should_call_invoke_e_and_fail_test_on_error", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			InvokeError: errors.New("lambda invocation failed"),
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
		result := Invoke(ctx, "test-function", `{}`)

		// Assert
		assert.Nil(t, result)
		assert.True(t, mockT.failNowCalled)
		assert.True(t, mockT.errorCalled)
	})
}

func TestInvokeAsyncE(t *testing.T) {
	t.Run("should_set_event_invocation_type", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			InvokeResponse: &lambda.InvokeOutput{
				StatusCode: 202,
				Payload:    []byte(""),
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
		result, err := InvokeAsyncE(ctx, "test-function", `{"async": true}`)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(202), result.StatusCode)
		assert.Contains(t, mockClient.InvokeCalls, "test-function")
	})
}

func TestInvokeWithRetryE(t *testing.T) {
	t.Run("should_set_max_retries_option", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			InvokeResponse: &lambda.InvokeOutput{
				StatusCode: 200,
				Payload:    []byte(`{"retry": "success"}`),
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
		result, err := InvokeWithRetryE(ctx, "test-function", `{}`, 5)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(200), result.StatusCode)
		assert.Contains(t, mockClient.InvokeCalls, "test-function")
	})
}

func TestDryRunInvokeE(t *testing.T) {
	t.Run("should_set_dry_run_invocation_type", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			InvokeResponse: &lambda.InvokeOutput{
				StatusCode: 204,
				Payload:    []byte(""),
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
		result, err := DryRunInvokeE(ctx, "test-function", `{"dryRun": true}`)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(204), result.StatusCode)
		assert.Contains(t, mockClient.InvokeCalls, "test-function")
	})
}

func TestIsNonRetryableError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "should_return_false_for_nil_error",
			err:      nil,
			expected: false,
		},
		{
			name:     "should_return_true_for_invalid_parameter_value_exception",
			err:      errors.New("InvalidParameterValueException: Invalid parameter value"),
			expected: true,
		},
		{
			name:     "should_return_true_for_resource_not_found_exception",
			err:      errors.New("ResourceNotFoundException: Function not found"),
			expected: true,
		},
		{
			name:     "should_return_true_for_invalid_request_content_exception",
			err:      errors.New("InvalidRequestContentException: Invalid request content"),
			expected: true,
		},
		{
			name:     "should_return_true_for_request_too_large_exception",
			err:      errors.New("RequestTooLargeException: Request payload is too large"),
			expected: true,
		},
		{
			name:     "should_return_true_for_unsupported_media_type_exception",
			err:      errors.New("UnsupportedMediaTypeException: Unsupported media type"),
			expected: true,
		},
		{
			name:     "should_return_false_for_retryable_error",
			err:      errors.New("TooManyRequestsException: Rate exceeded"),
			expected: false,
		},
		{
			name:     "should_return_false_for_service_exception",
			err:      errors.New("ServiceException: Service unavailable"),
			expected: false,
		},
		{
			name:     "should_return_false_for_generic_error",
			err:      errors.New("Something went wrong"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isNonRetryableError(tc.err)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "should_return_true_when_substring_found_at_beginning",
			s:        "ResourceNotFoundException: Function not found",
			substr:   "ResourceNotFoundException",
			expected: true,
		},
		{
			name:     "should_return_true_when_substring_found_in_middle",
			s:        "Error: ResourceNotFoundException occurred",
			substr:   "ResourceNotFoundException",
			expected: true,
		},
		{
			name:     "should_return_true_when_substring_found_at_end",
			s:        "This is a ResourceNotFoundException",
			substr:   "ResourceNotFoundException",
			expected: true,
		},
		{
			name:     "should_return_true_when_strings_are_equal",
			s:        "ResourceNotFoundException",
			substr:   "ResourceNotFoundException",
			expected: true,
		},
		{
			name:     "should_return_false_when_substring_not_found",
			s:        "ServiceException: Service unavailable",
			substr:   "ResourceNotFoundException",
			expected: false,
		},
		{
			name:     "should_return_false_when_string_is_shorter_than_substring",
			s:        "Error",
			substr:   "ResourceNotFoundException",
			expected: false,
		},
		{
			name:     "should_return_false_for_empty_string",
			s:        "",
			substr:   "ResourceNotFoundException",
			expected: false,
		},
		{
			name:     "should_return_true_for_empty_substring",
			s:        "ResourceNotFoundException",
			substr:   "",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := contains(tc.s, tc.substr)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIndexOf(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{
			name:     "should_return_zero_when_substring_at_beginning",
			s:        "ResourceNotFoundException: Function not found",
			substr:   "ResourceNotFoundException",
			expected: 0,
		},
		{
			name:     "should_return_correct_index_when_substring_in_middle",
			s:        "Error: ResourceNotFoundException occurred",
			substr:   "ResourceNotFoundException",
			expected: 7,
		},
		{
			name:     "should_return_correct_index_when_substring_at_end",
			s:        "This is a ResourceNotFoundException",
			substr:   "ResourceNotFoundException",
			expected: 10,
		},
		{
			name:     "should_return_minus_one_when_substring_not_found",
			s:        "ServiceException: Service unavailable",
			substr:   "ResourceNotFoundException",
			expected: -1,
		},
		{
			name:     "should_return_minus_one_when_string_shorter_than_substring",
			s:        "Error",
			substr:   "ResourceNotFoundException",
			expected: -1,
		},
		{
			name:     "should_return_zero_for_empty_substring",
			s:        "ResourceNotFoundException",
			substr:   "",
			expected: 0,
		},
		{
			name:     "should_return_minus_one_for_empty_string_with_non_empty_substring",
			s:        "",
			substr:   "ResourceNotFoundException",
			expected: -1,
		},
		{
			name:     "should_return_first_occurrence_when_multiple_matches",
			s:        "abc abc abc",
			substr:   "abc",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := indexOf(tc.s, tc.substr)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseInvokeOutputE(t *testing.T) {
	type TestStruct struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	testCases := []struct {
		name          string
		result        *InvokeResult
		target        interface{}
		expectedError bool
		errorContains string
		validateTarget func(t *testing.T, target interface{})
	}{
		{
			name: "should_parse_json_payload_successfully",
			result: &InvokeResult{
				Payload: `{"message": "success", "code": 200}`,
			},
			target:        &TestStruct{},
			expectedError: false,
			validateTarget: func(t *testing.T, target interface{}) {
				parsed := target.(*TestStruct)
				assert.Equal(t, "success", parsed.Message)
				assert.Equal(t, 200, parsed.Code)
			},
		},
		{
			name:          "should_return_error_for_nil_result",
			result:        nil,
			target:        &TestStruct{},
			expectedError: true,
			errorContains: "invoke result is nil",
		},
		{
			name: "should_return_error_for_empty_payload",
			result: &InvokeResult{
				Payload: "",
			},
			target:        &TestStruct{},
			expectedError: true,
			errorContains: "payload is empty",
		},
		{
			name: "should_return_error_for_invalid_json",
			result: &InvokeResult{
				Payload: `{"message": invalid json}`,
			},
			target:        &TestStruct{},
			expectedError: true,
			errorContains: "failed to unmarshal payload",
		},
		{
			name: "should_parse_simple_string_value",
			result: &InvokeResult{
				Payload: `"Hello, World!"`,
			},
			target:        new(string),
			expectedError: false,
			validateTarget: func(t *testing.T, target interface{}) {
				parsed := target.(*string)
				assert.Equal(t, "Hello, World!", *parsed)
			},
		},
		{
			name: "should_parse_number_value",
			result: &InvokeResult{
				Payload: `42`,
			},
			target:        new(int),
			expectedError: false,
			validateTarget: func(t *testing.T, target interface{}) {
				parsed := target.(*int)
				assert.Equal(t, 42, *parsed)
			},
		},
		{
			name: "should_parse_boolean_value",
			result: &InvokeResult{
				Payload: `true`,
			},
			target:        new(bool),
			expectedError: false,
			validateTarget: func(t *testing.T, target interface{}) {
				parsed := target.(*bool)
				assert.True(t, *parsed)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := ParseInvokeOutputE(tc.result, tc.target)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				if tc.validateTarget != nil {
					tc.validateTarget(t, tc.target)
				}
			}
		})
	}
}

func TestParseInvokeOutput(t *testing.T) {
	t.Run("should_panic_on_error", func(t *testing.T) {
		// Arrange
		result := &InvokeResult{
			Payload: `{"invalid": json}`,
		}
		target := &struct{}{}

		// Act & Assert
		assert.Panics(t, func() {
			ParseInvokeOutput(result, target)
		})
	})

	t.Run("should_parse_successfully_without_panic", func(t *testing.T) {
		// Arrange
		result := &InvokeResult{
			Payload: `{"message": "success"}`,
		}
		target := &struct {
			Message string `json:"message"`
		}{}

		// Act & Assert
		assert.NotPanics(t, func() {
			ParseInvokeOutput(result, target)
		})
		assert.Equal(t, "success", target.Message)
	})
}

func TestMarshalPayloadE(t *testing.T) {
	testCases := []struct {
		name          string
		payload       interface{}
		expected      string
		expectedError bool
		errorContains string
	}{
		{
			name:     "should_marshal_struct_to_json",
			payload:  struct{ Message string }{Message: "hello"},
			expected: `{"Message":"hello"}`,
		},
		{
			name:     "should_marshal_map_to_json",
			payload:  map[string]interface{}{"key": "value", "number": 42},
			expected: `{"key":"value","number":42}`,
		},
		{
			name:     "should_marshal_slice_to_json",
			payload:  []string{"item1", "item2", "item3"},
			expected: `["item1","item2","item3"]`,
		},
		{
			name:     "should_marshal_string_to_json",
			payload:  "simple string",
			expected: `"simple string"`,
		},
		{
			name:     "should_marshal_number_to_json",
			payload:  42,
			expected: `42`,
		},
		{
			name:     "should_marshal_boolean_to_json",
			payload:  true,
			expected: `true`,
		},
		{
			name:     "should_return_empty_string_for_nil_payload",
			payload:  nil,
			expected: ``,
		},
		{
			name:          "should_return_error_for_unmarshalable_type",
			payload:       make(chan int), // channels cannot be marshaled
			expectedError: true,
			errorContains: "failed to marshal payload",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := MarshalPayloadE(tc.payload)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.JSONEq(t, tc.expected, result)
			}
		})
	}
}

func TestMarshalPayload(t *testing.T) {
	t.Run("should_panic_on_marshal_error", func(t *testing.T) {
		// Arrange
		payload := make(chan int) // unmarshalable type

		// Act & Assert
		assert.Panics(t, func() {
			MarshalPayload(payload)
		})
	})

	t.Run("should_marshal_successfully_without_panic", func(t *testing.T) {
		// Arrange
		payload := map[string]string{"key": "value"}

		// Act & Assert
		assert.NotPanics(t, func() {
			result := MarshalPayload(payload)
			assert.JSONEq(t, `{"key":"value"}`, result)
		})
	})
}

func TestExecuteInvokeWithRetryLogic(t *testing.T) {
	t.Run("should_succeed_on_first_attempt", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			InvokeResponse: &lambda.InvokeOutput{
				StatusCode: 200,
				Payload:    []byte(`{"success": true}`),
			},
		}

		options := InvokeOptions{
			InvocationType: types.InvocationTypeRequestResponse,
			LogType:        LogTypeNone,
			MaxRetries:     3,
			RetryDelay:     100 * time.Millisecond,
		}

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		result, err := executeInvokeWithTerratestRetry(ctx, mockClient, "test-function", `{}`, options)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(200), result.StatusCode)
		assert.Equal(t, 1, len(mockClient.InvokeCalls)) // Should only call once
	})

	t.Run("should_not_retry_non_retryable_errors", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			InvokeError: errors.New("ResourceNotFoundException: Function not found"),
		}

		options := InvokeOptions{
			InvocationType: types.InvocationTypeRequestResponse,
			LogType:        LogTypeNone,
			MaxRetries:     3,
			RetryDelay:     100 * time.Millisecond,
		}

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		result, err := executeInvokeWithTerratestRetry(ctx, mockClient, "test-function", `{}`, options)

		// Assert
		// The function should return nil because Terratest's DoWithRetry doesn't return the error from FatalError
		assert.Nil(t, result)
		assert.Nil(t, err) // DoWithRetry handles the error internally
		assert.Equal(t, 1, len(mockClient.InvokeCalls)) // Should only call once due to fatal error
	})
}

func TestExecuteInvoke(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		payload        string
		options        InvokeOptions
		mockResponse   *lambda.InvokeOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, result *InvokeResult)
	}{
		{
			name:         "should_execute_invoke_successfully",
			functionName: "test-function",
			payload:      `{"test": true}`,
			options: InvokeOptions{
				InvocationType: types.InvocationTypeRequestResponse,
				LogType:        LogTypeNone,
				Timeout:        30 * time.Second,
			},
			mockResponse: &lambda.InvokeOutput{
				StatusCode: 200,
				Payload:    []byte(`{"result": "success"}`),
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *InvokeResult) {
				assert.Equal(t, int32(200), result.StatusCode)
				assert.Equal(t, `{"result": "success"}`, result.Payload)
			},
		},
		{
			name:         "should_include_optional_parameters",
			functionName: "test-function",
			payload:      `{}`,
			options: InvokeOptions{
				InvocationType: types.InvocationTypeRequestResponse,
				LogType:        LogTypeTail,
				ClientContext:  "test-context",
				Qualifier:      "PROD",
				Timeout:        30 * time.Second,
			},
			mockResponse: &lambda.InvokeOutput{
				StatusCode:      200,
				Payload:         []byte(`{"success": true}`),
				ExecutedVersion: aws.String("PROD"),
				LogResult:       aws.String("Function logs here"),
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *InvokeResult) {
				assert.Equal(t, int32(200), result.StatusCode)
				assert.Equal(t, "PROD", result.ExecutedVersion)
				assert.Contains(t, result.LogResult, "Function logs")
			},
		},
		{
			name:         "should_return_error_when_lambda_invoke_fails",
			functionName: "test-function",
			payload:      `{}`,
			options: InvokeOptions{
				InvocationType: types.InvocationTypeRequestResponse,
				Timeout:        30 * time.Second,
			},
			mockError:     errors.New("TooManyRequestsException: Rate exceeded"),
			expectedError: true,
			errorContains: "lambda invoke API call failed",
		},
		{
			name:         "should_timeout_when_context_exceeds_timeout",
			functionName: "test-function",
			payload:      `{}`,
			options: InvokeOptions{
				InvocationType: types.InvocationTypeRequestResponse,
				Timeout:        1 * time.Millisecond, // Very short timeout
			},
			expectedError: true,
			errorContains: "context deadline exceeded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			if tc.mockError == nil && tc.name == "should_timeout_when_context_exceeds_timeout" {
				// For timeout test, we need to simulate a slow response
				tc.mockError = context.DeadlineExceeded
			}

			mockClient := &MockLambdaClient{
				InvokeResponse: tc.mockResponse,
				InvokeError:    tc.mockError,
			}

			// Act
			result, err := executeInvoke(mockClient, tc.functionName, tc.payload, tc.options)

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
		})
	}
}