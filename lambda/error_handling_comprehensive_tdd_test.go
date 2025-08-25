package lambda

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TDD RED: Write failing tests first for error handling scenarios

func TestInvocationErrorHandling_TDD_ComprehensiveScenarios(t *testing.T) {
	errorScenarios := []struct {
		name               string
		setupMockError     func(*MockLambdaClient)
		functionName       string
		payload            string
		options            *InvokeOptions
		expectedErrorType  string
		expectedErrorMsg   string
		shouldHaveResult   bool
		validateResult     func(*testing.T, *InvokeResult)
	}{
		{
			name:         "ResourceNotFound_HandledCorrectly",
			functionName: "non-existent-function",
			payload:      `{}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeError = &types.ResourceNotFoundException{
					Message: aws.String("Function not found"),
				}
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "Function not found",
			shouldHaveResult:  false,
		},
		{
			name:         "InvalidParameterValue_HandledCorrectly",
			functionName: "test-function",
			payload:      `{}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeError = &types.InvalidParameterValueException{
					Message: aws.String("Invalid parameter value"),
				}
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "Invalid parameter value",
			shouldHaveResult:  false,
		},
		{
			name:         "RequestTooLarge_HandledCorrectly",
			functionName: "test-function",
			payload:      `{}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeError = &types.RequestTooLargeException{
					Message: aws.String("Request payload too large"),
				}
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "Request payload too large",
			shouldHaveResult:  false,
		},
		{
			name:         "ServiceException_HandledCorrectly",
			functionName: "test-function",
			payload:      `{}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeError = &types.ServiceException{
					Message: aws.String("Internal service error"),
				}
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "Internal service error",
			shouldHaveResult:  false,
		},
		{
			name:         "TooManyRequestsException_HandledCorrectly",
			functionName: "test-function",
			payload:      `{}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeError = &types.TooManyRequestsException{
					Message: aws.String("Rate exceeded"),
				}
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "Rate exceeded",
			shouldHaveResult:  false,
		},
		{
			name:         "UnsupportedMediaType_HandledCorrectly",
			functionName: "test-function",
			payload:      `{}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeError = &types.UnsupportedMediaTypeException{
					Message: aws.String("Unsupported media type"),
				}
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "Unsupported media type",
			shouldHaveResult:  false,
		},
		{
			name:         "FunctionError_UnhandledException",
			functionName: "error-function",
			payload:      `{"cause_error": true}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode:    200,
					Payload:       []byte(`{"errorMessage": "Runtime error occurred"}`),
					FunctionError: aws.String("Unhandled"),
				}
				mock.InvokeError = nil
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "lambda function error: Unhandled",
			shouldHaveResult:  true,
			validateResult: func(t *testing.T, result *InvokeResult) {
				assert.Equal(t, "Unhandled", result.FunctionError)
				assert.Contains(t, result.Payload, "Runtime error occurred")
			},
		},
		{
			name:         "FunctionError_HandledException",
			functionName: "error-function",
			payload:      `{"cause_handled_error": true}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode:    200,
					Payload:       []byte(`{"errorMessage": "Handled error", "errorType": "CustomError"}`),
					FunctionError: aws.String("Handled"),
				}
				mock.InvokeError = nil
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "lambda function error: Handled",
			shouldHaveResult:  true,
			validateResult: func(t *testing.T, result *InvokeResult) {
				assert.Equal(t, "Handled", result.FunctionError)
				assert.Contains(t, result.Payload, "CustomError")
			},
		},
		{
			name:         "HTTPStatusError_4xx",
			functionName: "test-function",
			payload:      `{}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode: 400,
					Payload:    []byte(`{"error": "Bad request"}`),
				}
				mock.InvokeError = nil
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "lambda invocation failed with status code: 400",
			shouldHaveResult:  true,
			validateResult: func(t *testing.T, result *InvokeResult) {
				assert.Equal(t, int32(400), result.StatusCode)
				assert.Contains(t, result.Payload, "Bad request")
			},
		},
		{
			name:         "HTTPStatusError_5xx",
			functionName: "test-function",
			payload:      `{}`,
			options:      nil,
			setupMockError: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode: 500,
					Payload:    []byte(`{"error": "Internal server error"}`),
				}
				mock.InvokeError = nil
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "lambda invocation failed with status code: 500",
			shouldHaveResult:  true,
			validateResult: func(t *testing.T, result *InvokeResult) {
				assert.Equal(t, int32(500), result.StatusCode)
				assert.Contains(t, result.Payload, "Internal server error")
			},
		},
		{
			name:         "ContextTimeout_HandledCorrectly",
			functionName: "slow-function",
			payload:      `{}`,
			options: &InvokeOptions{
				Timeout: 1 * time.Nanosecond, // Extremely short timeout
			},
			setupMockError: func(mock *MockLambdaClient) {
				// Simulate slow response by introducing delay in mock
				mock.InvokeError = context.DeadlineExceeded
			},
			expectedErrorType: "lambda invocation failed",
			expectedErrorMsg:  "context deadline exceeded",
			shouldHaveResult:  false,
		},
	}
	
	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := createTestContextTDD(t)
			
			// Setup mock
			oldFactory := globalClientFactory
			defer func() { globalClientFactory = oldFactory }()
			
			mockFactory := &MockClientFactory{}
			globalClientFactory = mockFactory
			mockClient := &MockLambdaClient{}
			mockFactory.LambdaClient = mockClient
			
			scenario.setupMockError(mockClient)
			
			// Execute test
			result, err := InvokeWithOptionsE(ctx, scenario.functionName, scenario.payload, scenario.options)
			
			// Validate error
			require.Error(t, err)
			assert.Contains(t, err.Error(), scenario.expectedErrorType)
			if scenario.expectedErrorMsg != "" {
				assert.Contains(t, err.Error(), scenario.expectedErrorMsg)
			}
			
			// Validate result presence
			if scenario.shouldHaveResult {
				require.NotNil(t, result, "Expected result to be present for error scenario")
				if scenario.validateResult != nil {
					scenario.validateResult(t, result)
				}
			} else {
				assert.Nil(t, result, "Expected result to be nil for error scenario")
			}
		})
	}
}

func TestValidationErrors_TDD_ComprehensiveScenarios(t *testing.T) {
	validationScenarios := []struct {
		name         string
		functionName string
		payload      string
		options      *InvokeOptions
		expectedErr  string
	}{
		{
			name:         "EmptyFunctionName_ValidationError",
			functionName: "",
			payload:      `{}`,
			options:      nil,
			expectedErr:  "invalid function name",
		},
		{
			name:         "FunctionNameTooLong_ValidationError",
			functionName: "this-function-name-is-way-too-long-and-exceeds-the-64-character-limit-for-aws-lambda-functions-by-design",
			payload:      `{}`,
			options:      nil,
			expectedErr:  "name too long",
		},
		{
			name:         "InvalidFunctionNameCharacters_ValidationError",
			functionName: "invalid@function#name",
			payload:      `{}`,
			options:      nil,
			expectedErr:  "invalid character",
		},
		{
			name:         "InvalidJSONPayload_ValidationError",
			functionName: "test-function",
			payload:      `{"invalid": json, "missing": quotes}`,
			options:      nil,
			expectedErr:  "invalid payload format",
		},
		{
			name:         "PayloadTooLarge_ValidationError",
			functionName: "test-function",
			payload:      createOversizedPayload(),
			options:      nil,
			expectedErr:  "payload size exceeds maximum",
		},
		{
			name:         "CustomValidatorError_ValidationError",
			functionName: "test-function",
			payload:      `{"test": "data"}`,
			options: &InvokeOptions{
				PayloadValidator: func(payload string) error {
					if len(payload) > 10 {
						return fmt.Errorf("custom validator failed: payload too long")
					}
					return nil
				},
			},
			expectedErr: "custom validator failed",
		},
	}
	
	for _, scenario := range validationScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := createTestContextTDD(t)
			
			// Setup mock (shouldn't be called due to validation failure)
			oldFactory := globalClientFactory
			defer func() { globalClientFactory = oldFactory }()
			
			mockFactory := &MockClientFactory{}
			globalClientFactory = mockFactory
			mockClient := &MockLambdaClient{}
			mockFactory.LambdaClient = mockClient
			
			// Execute test
			result, err := InvokeWithOptionsE(ctx, scenario.functionName, scenario.payload, scenario.options)
			
			// Validate error
			require.Error(t, err)
			assert.Contains(t, err.Error(), scenario.expectedErr)
			assert.Nil(t, result)
		})
	}
}

func TestNonRetryableErrorClassification_TDD(t *testing.T) {
	errorClassificationScenarios := []struct {
		name                 string
		error                error
		expectedNonRetryable bool
	}{
		{
			name:                 "NilError_IsRetryable",
			error:                nil,
			expectedNonRetryable: false,
		},
		{
			name:                 "InvalidParameterValueException_NonRetryable",
			error:                fmt.Errorf("InvalidParameterValueException: The parameter value is invalid"),
			expectedNonRetryable: true,
		},
		{
			name:                 "ResourceNotFoundException_NonRetryable",
			error:                fmt.Errorf("ResourceNotFoundException: The resource was not found"),
			expectedNonRetryable: true,
		},
		{
			name:                 "InvalidRequestContentException_NonRetryable",
			error:                fmt.Errorf("InvalidRequestContentException: The request content is invalid"),
			expectedNonRetryable: true,
		},
		{
			name:                 "RequestTooLargeException_NonRetryable",
			error:                fmt.Errorf("RequestTooLargeException: The request payload is too large"),
			expectedNonRetryable: true,
		},
		{
			name:                 "UnsupportedMediaTypeException_NonRetryable",
			error:                fmt.Errorf("UnsupportedMediaTypeException: The media type is not supported"),
			expectedNonRetryable: true,
		},
		{
			name:                 "ThrottlingException_Retryable",
			error:                fmt.Errorf("ThrottlingException: Rate limit exceeded"),
			expectedNonRetryable: false,
		},
		{
			name:                 "ServiceException_Retryable",
			error:                fmt.Errorf("ServiceException: Internal service error"),
			expectedNonRetryable: false,
		},
		{
			name:                 "NetworkError_Retryable",
			error:                fmt.Errorf("RequestError: network connection failed"),
			expectedNonRetryable: false,
		},
		{
			name:                 "TimeoutError_Retryable",
			error:                fmt.Errorf("RequestTimeout: request timed out"),
			expectedNonRetryable: false,
		},
		{
			name:                 "GenericError_Retryable",
			error:                fmt.Errorf("Unknown error occurred"),
			expectedNonRetryable: false,
		},
	}
	
	for _, scenario := range errorClassificationScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := isNonRetryableError(scenario.error)
			assert.Equal(t, scenario.expectedNonRetryable, result,
				"Error classification mismatch for: %v", scenario.error)
		})
	}
}

func TestEdgeCaseHandling_TDD(t *testing.T) {
	t.Run("EmptyPayload_HandledCorrectly", func(t *testing.T) {
		ctx := createTestContextTDD(t)
		
		oldFactory := globalClientFactory
		defer func() { globalClientFactory = oldFactory }()
		
		mockFactory := &MockClientFactory{}
		globalClientFactory = mockFactory
		mockClient := &MockLambdaClient{}
		mockFactory.LambdaClient = mockClient
		
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode: 200,
			Payload:    []byte(`{}`),
		}
		
		result, err := InvokeWithOptionsE(ctx, "test-function", "", nil)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int32(200), result.StatusCode)
	})
	
	t.Run("NullPayloadResponse_HandledCorrectly", func(t *testing.T) {
		ctx := createTestContextTDD(t)
		
		oldFactory := globalClientFactory
		defer func() { globalClientFactory = oldFactory }()
		
		mockFactory := &MockClientFactory{}
		globalClientFactory = mockFactory
		mockClient := &MockLambdaClient{}
		mockFactory.LambdaClient = mockClient
		
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode: 204,
			Payload:    nil, // Nil payload
		}
		
		result, err := InvokeWithOptionsE(ctx, "test-function", `{}`, nil)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int32(204), result.StatusCode)
		assert.Equal(t, "", result.Payload) // Should handle nil gracefully
	})
	
	t.Run("EmptyPayloadResponse_HandledCorrectly", func(t *testing.T) {
		ctx := createTestContextTDD(t)
		
		oldFactory := globalClientFactory
		defer func() { globalClientFactory = oldFactory }()
		
		mockFactory := &MockClientFactory{}
		globalClientFactory = mockFactory
		mockClient := &MockLambdaClient{}
		mockFactory.LambdaClient = mockClient
		
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode: 202,
			Payload:    []byte{}, // Empty payload
		}
		
		result, err := InvokeWithOptionsE(ctx, "test-function", `{}`, nil)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int32(202), result.StatusCode)
		assert.Equal(t, "", result.Payload)
	})
	
	t.Run("SpecialCharactersInFunctionName_HandledCorrectly", func(t *testing.T) {
		validNames := []string{
			"function-with-dashes",
			"function_with_underscores",
			"FunctionWithCamelCase",
			"function123with456numbers",
			"a", // Single character
		}
		
		for _, name := range validNames {
			err := validateFunctionName(name)
			assert.NoError(t, err, "Function name should be valid: %s", name)
		}
		
		invalidNames := []string{
			"function with spaces",
			"function@with.special",
			"function$with$dollars",
			"function%with%percents",
		}
		
		for _, name := range invalidNames {
			err := validateFunctionName(name)
			assert.Error(t, err, "Function name should be invalid: %s", name)
		}
	})
}

func TestPanicRecovery_TDD(t *testing.T) {
	t.Run("MarshalPayload_PanicsOnError", func(t *testing.T) {
		// Create an unmarshalable object
		unmarshalable := make(chan int)
		
		assert.Panics(t, func() {
			MarshalPayload(unmarshalable)
		}, "MarshalPayload should panic on unmarshalable objects")
	})
	
	t.Run("ParseInvokeOutput_PanicsOnError", func(t *testing.T) {
		result := &InvokeResult{
			Payload: `{"invalid": json}`,
		}
		
		var target map[string]interface{}
		assert.Panics(t, func() {
			ParseInvokeOutput(result, &target)
		}, "ParseInvokeOutput should panic on invalid JSON")
	})
}

func TestComplexPayloadHandling_TDD(t *testing.T) {
	t.Run("DeepNestedPayload_HandledCorrectly", func(t *testing.T) {
		complexPayload := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"level4": []interface{}{
							map[string]interface{}{
								"item": "value",
								"number": 42,
								"boolean": true,
								"null": nil,
							},
						},
					},
				},
			},
		}
		
		marshaled, err := MarshalPayloadE(complexPayload)
		require.NoError(t, err)
		assert.Contains(t, marshaled, "level4")
		assert.Contains(t, marshaled, "item")
		
		// Validate it's proper JSON
		err = validatePayload(marshaled)
		assert.NoError(t, err)
	})
	
	t.Run("UnicodePayload_HandledCorrectly", func(t *testing.T) {
		unicodePayload := map[string]interface{}{
			"message": "Hello, 世界! 🌍",
			"emoji":   "🚀🎉",
			"chinese": "你好世界",
			"arabic":  "مرحبا بالعالم",
		}
		
		marshaled, err := MarshalPayloadE(unicodePayload)
		require.NoError(t, err)
		
		err = validatePayload(marshaled)
		assert.NoError(t, err)
		
		// Should preserve unicode characters
		assert.Contains(t, marshaled, "世界")
		assert.Contains(t, marshaled, "🌍")
	})
}

// Helper function to create oversized payload
func createOversizedPayload() string {
	// Create payload larger than MaxPayloadSize (6MB)
	largeData := make([]byte, MaxPayloadSize+1000)
	for i := range largeData {
		largeData[i] = 'X'
	}
	
	payload := map[string]string{
		"large_field": string(largeData),
	}
	
	marshaled, _ := MarshalPayloadE(payload)
	return marshaled
}