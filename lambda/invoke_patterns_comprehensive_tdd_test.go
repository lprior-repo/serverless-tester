package lambda

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TDD RED: Write failing tests first for invocation patterns

func TestInvokeWithOptionsE_TDD_ComprehensiveScenarios(t *testing.T) {
	testScenarios := createInvokePatternScenarios()
	
	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := createTestContextTDD(t)
			
			// Setup mock factory for this scenario
			oldFactory := globalClientFactory
			defer func() { globalClientFactory = oldFactory }()
			
			mockFactory := &MockClientFactory{}
			globalClientFactory = mockFactory
			mockClient := &MockLambdaClient{}
			mockFactory.LambdaClient = mockClient
			
			// Configure mock response
			scenario.setupMock(mockClient)
			
			// Execute test
			result, err := InvokeWithOptionsE(ctx, scenario.functionName, scenario.payload, scenario.options)
			
			// Validate result
			scenario.validateResult(t, result, err)
		})
	}
}

type invokeScenario struct {
	name         string
	functionName string
	payload      string
	options      *InvokeOptions
	setupMock    func(*MockLambdaClient)
	validateResult func(*testing.T, *InvokeResult, error)
}

func createInvokePatternScenarios() []invokeScenario {
	return []invokeScenario{
		{
			name:         "ValidSyncInvocation_Success",
			functionName: "test-function",
			payload:      `{"key": "value"}`,
			options:      nil, // Use defaults
			setupMock: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode:      200,
					Payload:        []byte(`{"result": "success"}`),
					ExecutedVersion: aws.String("1"),
				}
				mock.InvokeError = nil
			},
			validateResult: func(t *testing.T, result *InvokeResult, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, int32(200), result.StatusCode)
				assert.Contains(t, result.Payload, "success")
				assert.Equal(t, "1", result.ExecutedVersion)
				assert.True(t, result.ExecutionTime > 0)
			},
		},
		{
			name:         "AsyncInvocation_Success",
			functionName: "async-function",
			payload:      `{"message": "async"}`,
			options: &InvokeOptions{
				InvocationType: types.InvocationTypeEvent,
				LogType:        LogTypeNone,
			},
			setupMock: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode: 202,
					Payload:    []byte(""),
				}
				mock.InvokeError = nil
			},
			validateResult: func(t *testing.T, result *InvokeResult, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, int32(202), result.StatusCode)
			},
		},
		{
			name:         "DryRunInvocation_Success",
			functionName: "dry-run-function",
			payload:      `{"test": true}`,
			options: &InvokeOptions{
				InvocationType: types.InvocationTypeDryRun,
			},
			setupMock: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode: 204,
					Payload:    []byte(""),
				}
				mock.InvokeError = nil
			},
			validateResult: func(t *testing.T, result *InvokeResult, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, int32(204), result.StatusCode)
			},
		},
		{
			name:         "InvocationWithQualifier_Success",
			functionName: "versioned-function",
			payload:      `{"version": "2"}`,
			options: &InvokeOptions{
				Qualifier: "2",
			},
			setupMock: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode:      200,
					Payload:        []byte(`{"version": "executed"}`),
					ExecutedVersion: aws.String("2"),
				}
				mock.InvokeError = nil
			},
			validateResult: func(t *testing.T, result *InvokeResult, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, "2", result.ExecutedVersion)
			},
		},
		{
			name:         "InvocationWithClientContext_Success",
			functionName: "context-function",
			payload:      `{}`,
			options: &InvokeOptions{
				ClientContext: "custom-context",
			},
			setupMock: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode: 200,
					Payload:    []byte(`{"context": "received"}`),
				}
				mock.InvokeError = nil
			},
			validateResult: func(t *testing.T, result *InvokeResult, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
			},
		},
		{
			name:         "FunctionError_HandledCorrectly",
			functionName: "error-function",
			payload:      `{"cause_error": true}`,
			options:      nil,
			setupMock: func(mock *MockLambdaClient) {
				mock.InvokeResponse = &lambda.InvokeOutput{
					StatusCode:    200,
					Payload:       []byte(`{"errorMessage": "Function failed"}`),
					FunctionError: aws.String("Unhandled"),
				}
				mock.InvokeError = nil
			},
			validateResult: func(t *testing.T, result *InvokeResult, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "Unhandled")
				require.NotNil(t, result)
				assert.Equal(t, "Unhandled", result.FunctionError)
			},
		},
		{
			name:         "InvalidFunctionName_ReturnsError",
			functionName: "",
			payload:      `{}`,
			options:      nil,
			setupMock:    func(mock *MockLambdaClient) {},
			validateResult: func(t *testing.T, result *InvokeResult, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid function name")
				assert.Nil(t, result)
			},
		},
		{
			name:         "InvalidPayload_ReturnsError",
			functionName: "test-function",
			payload:      `{"invalid": json}`,
			options:      nil,
			setupMock:    func(mock *MockLambdaClient) {},
			validateResult: func(t *testing.T, result *InvokeResult, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid payload")
				assert.Nil(t, result)
			},
		},
		{
			name:         "LargePayload_ReturnsError",
			functionName: "test-function",
			payload:      createLargePayloadTDD(),
			options:      nil,
			setupMock:    func(mock *MockLambdaClient) {},
			validateResult: func(t *testing.T, result *InvokeResult, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "payload size exceeds maximum")
				assert.Nil(t, result)
			},
		},
	}
}

func TestExecuteInvoke_TDD_CoreFunctionality(t *testing.T) {
	t.Run("ExecuteInvoke_ValidCall_Success", func(t *testing.T) {
		mockClient := &MockLambdaClient{}
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode: 200,
			Payload:    []byte(`{"success": true}`),
		}
		
		options := defaultInvokeOptions()
		result, err := executeInvoke(mockClient, "test-function", `{}`, options)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int32(200), result.StatusCode)
		assert.Contains(t, result.Payload, "success")
	})
	
	t.Run("ExecuteInvoke_ClientError_ReturnsError", func(t *testing.T) {
		mockClient := &MockLambdaClient{}
		mockClient.InvokeError = fmt.Errorf("API error")
		
		options := defaultInvokeOptions()
		result, err := executeInvoke(mockClient, "test-function", `{}`, options)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lambda invoke API call failed")
		assert.Nil(t, result)
	})
	
	t.Run("ExecuteInvoke_BadStatusCode_ReturnsError", func(t *testing.T) {
		mockClient := &MockLambdaClient{}
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode: 500,
			Payload:    []byte(`{"error": "internal"}`),
		}
		
		options := defaultInvokeOptions()
		result, err := executeInvoke(mockClient, "test-function", `{}`, options)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status code: 500")
		require.NotNil(t, result)
		assert.Equal(t, int32(500), result.StatusCode)
	})
}

func TestIsNonRetryableError_TDD(t *testing.T) {
	retryErrorScenarios := []struct {
		name        string
		err         error
		shouldRetry bool
	}{
		{
			name:        "NilError_Retryable",
			err:         nil,
			shouldRetry: true,
		},
		{
			name:        "InvalidParameterValueException_NonRetryable",
			err:         fmt.Errorf("InvalidParameterValueException: invalid parameter"),
			shouldRetry: false,
		},
		{
			name:        "ResourceNotFoundException_NonRetryable",
			err:         fmt.Errorf("ResourceNotFoundException: function not found"),
			shouldRetry: false,
		},
		{
			name:        "InvalidRequestContentException_NonRetryable",
			err:         fmt.Errorf("InvalidRequestContentException: bad request"),
			shouldRetry: false,
		},
		{
			name:        "RequestTooLargeException_NonRetryable",
			err:         fmt.Errorf("RequestTooLargeException: payload too large"),
			shouldRetry: false,
		},
		{
			name:        "UnsupportedMediaTypeException_NonRetryable",
			err:         fmt.Errorf("UnsupportedMediaTypeException: unsupported type"),
			shouldRetry: false,
		},
		{
			name:        "ThrottlingException_Retryable",
			err:         fmt.Errorf("ThrottlingException: rate exceeded"),
			shouldRetry: true,
		},
		{
			name:        "InternalError_Retryable",
			err:         fmt.Errorf("InternalError: service unavailable"),
			shouldRetry: true,
		},
	}
	
	for _, scenario := range retryErrorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			isNonRetryable := isNonRetryableError(scenario.err)
			
			if scenario.shouldRetry {
				assert.False(t, isNonRetryable, "Error should be retryable")
			} else {
				assert.True(t, isNonRetryable, "Error should not be retryable")
			}
		})
	}
}

func TestContainsAndIndexOf_TDD_StringUtilities(t *testing.T) {
	t.Run("Contains_VariousScenarios", func(t *testing.T) {
		containsScenarios := []struct {
			s        string
			substr   string
			expected bool
		}{
			{"hello world", "world", true},
			{"hello world", "xyz", false},
			{"test", "test", true},
			{"", "", true},
			{"", "test", false},
			{"test", "", true},
			{"ResourceNotFoundException", "ResourceNotFound", true},
		}
		
		for _, scenario := range containsScenarios {
			result := contains(scenario.s, scenario.substr)
			assert.Equal(t, scenario.expected, result,
				"contains(%q, %q) should be %v", scenario.s, scenario.substr, scenario.expected)
		}
	})
	
	t.Run("IndexOf_VariousScenarios", func(t *testing.T) {
		indexScenarios := []struct {
			s        string
			substr   string
			expected int
		}{
			{"hello world", "world", 6},
			{"hello world", "hello", 0},
			{"hello world", "xyz", -1},
			{"test", "test", 0},
			{"", "", 0},
			{"", "test", -1},
			{"InvalidParameterValueException", "Parameter", 7},
		}
		
		for _, scenario := range indexScenarios {
			result := indexOf(scenario.s, scenario.substr)
			assert.Equal(t, scenario.expected, result,
				"indexOf(%q, %q) should be %d", scenario.s, scenario.substr, scenario.expected)
		}
	})
}

func TestMarshalAndParsePayload_TDD_JSONOperations(t *testing.T) {
	t.Run("MarshalPayloadE_ValidObjects", func(t *testing.T) {
		payloadScenarios := []struct {
			name     string
			payload  interface{}
			expected string
		}{
			{
				name:     "SimpleMap",
				payload:  map[string]interface{}{"key": "value"},
				expected: `{"key":"value"}`,
			},
			{
				name:     "NilPayload",
				payload:  nil,
				expected: "",
			},
			{
				name:     "StringPayload",
				payload:  "simple string",
				expected: `"simple string"`,
			},
			{
				name:     "NumberPayload",
				payload:  42,
				expected: "42",
			},
			{
				name:     "BooleanPayload",
				payload:  true,
				expected: "true",
			},
		}
		
		for _, scenario := range payloadScenarios {
			t.Run(scenario.name, func(t *testing.T) {
				result, err := MarshalPayloadE(scenario.payload)
				require.NoError(t, err)
				assert.Equal(t, scenario.expected, result)
			})
		}
	})
	
	t.Run("ParseInvokeOutputE_ValidPayloads", func(t *testing.T) {
		t.Run("ParseValidJSON", func(t *testing.T) {
			result := &InvokeResult{
				Payload: `{"message": "success", "code": 200}`,
			}
			
			var target map[string]interface{}
			err := ParseInvokeOutputE(result, &target)
			
			require.NoError(t, err)
			assert.Equal(t, "success", target["message"])
			assert.Equal(t, float64(200), target["code"]) // JSON numbers are float64
		})
		
		t.Run("ParseNilResult_ReturnsError", func(t *testing.T) {
			var target map[string]interface{}
			err := ParseInvokeOutputE(nil, &target)
			
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invoke result is nil")
		})
		
		t.Run("ParseEmptyPayload_ReturnsError", func(t *testing.T) {
			result := &InvokeResult{Payload: ""}
			var target map[string]interface{}
			err := ParseInvokeOutputE(result, &target)
			
			require.Error(t, err)
			assert.Contains(t, err.Error(), "payload is empty")
		})
		
		t.Run("ParseInvalidJSON_ReturnsError", func(t *testing.T) {
			result := &InvokeResult{Payload: `{"invalid": json}`}
			var target map[string]interface{}
			err := ParseInvokeOutputE(result, &target)
			
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to unmarshal payload")
		})
	})
}

func TestTerratestPatternFunctions_TDD(t *testing.T) {
	t.Run("Invoke_Success", func(t *testing.T) {
		ctx := createTestContextTDD(t)
		
		// Setup mock
		oldFactory := globalClientFactory
		defer func() { globalClientFactory = oldFactory }()
		
		mockFactory := &MockClientFactory{}
		globalClientFactory = mockFactory
		mockClient := &MockLambdaClient{}
		mockFactory.LambdaClient = mockClient
		
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode: 200,
			Payload:    []byte(`{"result": "success"}`),
		}
		
		result := Invoke(ctx, "test-function", `{}`)
		
		require.NotNil(t, result)
		assert.Equal(t, int32(200), result.StatusCode)
	})
	
	t.Run("InvokeAsync_Success", func(t *testing.T) {
		ctx := createTestContextTDD(t)
		
		oldFactory := globalClientFactory
		defer func() { globalClientFactory = oldFactory }()
		
		mockFactory := &MockClientFactory{}
		globalClientFactory = mockFactory
		mockClient := &MockLambdaClient{}
		mockFactory.LambdaClient = mockClient
		
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode: 202,
			Payload:    []byte(""),
		}
		
		result := InvokeAsync(ctx, "async-function", `{}`)
		
		require.NotNil(t, result)
		assert.Equal(t, int32(202), result.StatusCode)
	})
	
	t.Run("DryRunInvoke_Success", func(t *testing.T) {
		ctx := createTestContextTDD(t)
		
		oldFactory := globalClientFactory
		defer func() { globalClientFactory = oldFactory }()
		
		mockFactory := &MockClientFactory{}
		globalClientFactory = mockFactory
		mockClient := &MockLambdaClient{}
		mockFactory.LambdaClient = mockClient
		
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode: 204,
			Payload:    []byte(""),
		}
		
		result := DryRunInvoke(ctx, "dry-run-function", `{}`)
		
		require.NotNil(t, result)
		assert.Equal(t, int32(204), result.StatusCode)
	})
}

func TestValidationFunctions_TDD_Comprehensive(t *testing.T) {
	t.Run("ValidateFunctionName_TDD", func(t *testing.T) {
		validationScenarios := []struct {
			name         string
			functionName string
			expectError  bool
		}{
			{
				name:         "ValidName",
				functionName: "my-test-function_123",
				expectError:  false,
			},
			{
				name:         "EmptyName",
				functionName: "",
				expectError:  true,
			},
			{
				name:         "TooLongName",
				functionName: "this-is-a-very-long-function-name-that-exceeds-the-64-character-limit-for-aws-lambda",
				expectError:  true,
			},
			{
				name:         "InvalidCharacters",
				functionName: "invalid@function.name",
				expectError:  true,
			},
		}
		
		for _, scenario := range validationScenarios {
			t.Run(scenario.name, func(t *testing.T) {
				err := validateFunctionName(scenario.functionName)
				
				if scenario.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
	
	t.Run("ValidatePayload_TDD", func(t *testing.T) {
		payloadScenarios := []struct {
			name        string
			payload     string
			expectError bool
		}{
			{
				name:        "ValidJSON",
				payload:     `{"key": "value"}`,
				expectError: false,
			},
			{
				name:        "EmptyPayload",
				payload:     "",
				expectError: false,
			},
			{
				name:        "InvalidJSON",
				payload:     `{"invalid": json}`,
				expectError: true,
			},
		}
		
		for _, scenario := range payloadScenarios {
			t.Run(scenario.name, func(t *testing.T) {
				err := validatePayload(scenario.payload)
				
				if scenario.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestInvokeOptionsHandling_TDD(t *testing.T) {
	t.Run("MergeInvokeOptions_WithNilOptions", func(t *testing.T) {
		merged := mergeInvokeOptions(nil)
		
		assert.Equal(t, InvocationTypeRequestResponse, merged.InvocationType)
		assert.Equal(t, LogTypeNone, merged.LogType)
		assert.Equal(t, DefaultTimeout, merged.Timeout)
		assert.Equal(t, DefaultRetryAttempts, merged.MaxRetries)
		assert.Equal(t, DefaultRetryDelay, merged.RetryDelay)
		assert.NotNil(t, merged.PayloadValidator)
	})
	
	t.Run("MergeInvokeOptions_WithCustomOptions", func(t *testing.T) {
		customOpts := &InvokeOptions{
			InvocationType: types.InvocationTypeEvent,
			LogType:        LogTypeTail,
			Timeout:        60 * time.Second,
		}
		
		merged := mergeInvokeOptions(customOpts)
		
		assert.Equal(t, types.InvocationTypeEvent, merged.InvocationType)
		assert.Equal(t, LogTypeTail, merged.LogType)
		assert.Equal(t, 60*time.Second, merged.Timeout)
		// Other fields should use defaults
		assert.Equal(t, DefaultRetryAttempts, merged.MaxRetries)
	})
}

func TestLogSanitization_TDD(t *testing.T) {
	t.Run("SanitizeLogResult_RemovesSensitiveInfo", func(t *testing.T) {
		logResult := `Application log line 1
RequestId: abc-123-def
Application log line 2
Duration: 1000.00 ms
Billed Duration: 1000 ms
Final log line`
		
		sanitized := sanitizeLogResult(logResult)
		
		assert.NotContains(t, sanitized, "RequestId:")
		assert.NotContains(t, sanitized, "Duration:")
		assert.NotContains(t, sanitized, "Billed Duration:")
		assert.Contains(t, sanitized, "Application log line 1")
		assert.Contains(t, sanitized, "Application log line 2")
		assert.Contains(t, sanitized, "Final log line")
	})
}

// Helper functions

func createTestContextTDD(t *testing.T) *TestContext {
	return &TestContext{
		T:      t,
		Region: "us-east-1",
	}
}

func createLargePayloadTDD() string {
	// Create a payload larger than 6MB
	data := make(map[string]string)
	largeValue := make([]byte, 7*1024*1024) // 7MB of data
	for i := range largeValue {
		largeValue[i] = 'a'
	}
	data["large_field"] = string(largeValue)
	
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}