package lambda

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateFunctionName follows TDD principles - Red -> Green -> Refactor
func TestValidateFunctionName(t *testing.T) {
	testCases := []struct {
		name         string
		functionName string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "should_accept_valid_function_name",
			functionName: "my-lambda-function",
			expectError:  false,
		},
		{
			name:         "should_accept_function_name_with_underscores",
			functionName: "my_lambda_function_123",
			expectError:  false,
		},
		{
			name:         "should_accept_function_name_with_numbers",
			functionName: "lambda123",
			expectError:  false,
		},
		{
			name:         "should_reject_empty_function_name",
			functionName: "",
			expectError:  true,
			errorMessage: "invalid function name",
		},
		{
			name:         "should_reject_function_name_too_long",
			functionName: strings.Repeat("a", 65), // 65 characters, limit is 64
			expectError:  true,
			errorMessage: "name too long",
		},
		{
			name:         "should_reject_function_name_with_invalid_characters",
			functionName: "my-function@invalid",
			expectError:  true,
			errorMessage: "invalid character '@'",
		},
		{
			name:         "should_reject_function_name_with_spaces",
			functionName: "my function",
			expectError:  true,
			errorMessage: "invalid character ' '",
		},
		{
			name:         "should_accept_single_character_function_name",
			functionName: "f",
			expectError:  false,
		},
		{
			name:         "should_accept_max_length_function_name",
			functionName: strings.Repeat("a", 64), // exactly 64 characters
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateFunctionName(tc.functionName)

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

func TestValidatePayload(t *testing.T) {
	testCases := []struct {
		name        string
		payload     string
		expectError bool
		errorType   error
	}{
		{
			name:        "should_accept_empty_payload",
			payload:     "",
			expectError: false,
		},
		{
			name:        "should_accept_valid_json_object",
			payload:     `{"key": "value", "number": 42}`,
			expectError: false,
		},
		{
			name:        "should_accept_valid_json_array",
			payload:     `["item1", "item2", 123]`,
			expectError: false,
		},
		{
			name:        "should_accept_simple_json_string",
			payload:     `"simple string"`,
			expectError: false,
		},
		{
			name:        "should_accept_json_number",
			payload:     `42`,
			expectError: false,
		},
		{
			name:        "should_accept_json_boolean",
			payload:     `true`,
			expectError: false,
		},
		{
			name:        "should_accept_json_null",
			payload:     `null`,
			expectError: false,
		},
		{
			name:        "should_reject_invalid_json",
			payload:     `{"invalid": json}`,
			expectError: true,
			errorType:   ErrInvalidPayload,
		},
		{
			name:        "should_reject_malformed_json_object",
			payload:     `{"key": "value"`,
			expectError: true,
			errorType:   ErrInvalidPayload,
		},
		{
			name:        "should_reject_unquoted_string",
			payload:     `invalid payload`,
			expectError: true,
			errorType:   ErrInvalidPayload,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validatePayload(tc.payload)

			// Assert
			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.errorType))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	// Act
	config := defaultRetryConfig()

	// Assert
	assert.Equal(t, DefaultRetryAttempts, config.MaxAttempts)
	assert.Equal(t, DefaultRetryDelay, config.BaseDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

func TestDefaultInvokeOptions(t *testing.T) {
	// Act
	opts := defaultInvokeOptions()

	// Assert
	assert.Equal(t, InvocationTypeRequestResponse, opts.InvocationType)
	assert.Equal(t, LogTypeNone, opts.LogType)
	assert.Equal(t, DefaultTimeout, opts.Timeout)
	assert.Equal(t, DefaultRetryAttempts, opts.MaxRetries)
	assert.Equal(t, DefaultRetryDelay, opts.RetryDelay)
	assert.NotNil(t, opts.PayloadValidator)

	// Test payload validator
	err := opts.PayloadValidator("valid json")
	assert.NoError(t, err)

	// Test payload validator with oversized payload
	oversizedPayload := strings.Repeat("a", MaxPayloadSize+1)
	err = opts.PayloadValidator(oversizedPayload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payload size exceeds maximum")
}

func TestMergeInvokeOptions(t *testing.T) {
	testCases := []struct {
		name     string
		userOpts *InvokeOptions
		verify   func(t *testing.T, result InvokeOptions)
	}{
		{
			name:     "should_return_defaults_for_nil_options",
			userOpts: nil,
			verify: func(t *testing.T, result InvokeOptions) {
				defaults := defaultInvokeOptions()
				assert.Equal(t, defaults.InvocationType, result.InvocationType)
				assert.Equal(t, defaults.LogType, result.LogType)
				assert.Equal(t, defaults.Timeout, result.Timeout)
				assert.Equal(t, defaults.MaxRetries, result.MaxRetries)
				assert.Equal(t, defaults.RetryDelay, result.RetryDelay)
				assert.NotNil(t, result.PayloadValidator)
			},
		},
		{
			name: "should_preserve_user_invocation_type",
			userOpts: &InvokeOptions{
				InvocationType: types.InvocationTypeEvent,
			},
			verify: func(t *testing.T, result InvokeOptions) {
				assert.Equal(t, types.InvocationTypeEvent, result.InvocationType)
			},
		},
		{
			name: "should_preserve_user_log_type",
			userOpts: &InvokeOptions{
				LogType: LogTypeTail,
			},
			verify: func(t *testing.T, result InvokeOptions) {
				assert.Equal(t, LogTypeTail, result.LogType)
			},
		},
		{
			name: "should_preserve_user_timeout",
			userOpts: &InvokeOptions{
				Timeout: 60 * time.Second,
			},
			verify: func(t *testing.T, result InvokeOptions) {
				assert.Equal(t, 60*time.Second, result.Timeout)
			},
		},
		{
			name: "should_preserve_user_max_retries",
			userOpts: &InvokeOptions{
				MaxRetries: 5,
			},
			verify: func(t *testing.T, result InvokeOptions) {
				assert.Equal(t, 5, result.MaxRetries)
			},
		},
		{
			name: "should_preserve_user_retry_delay",
			userOpts: &InvokeOptions{
				RetryDelay: 2 * time.Second,
			},
			verify: func(t *testing.T, result InvokeOptions) {
				assert.Equal(t, 2*time.Second, result.RetryDelay)
			},
		},
		{
			name: "should_preserve_user_payload_validator",
			userOpts: &InvokeOptions{
				PayloadValidator: func(payload string) error {
					return errors.New("custom validator")
				},
			},
			verify: func(t *testing.T, result InvokeOptions) {
				err := result.PayloadValidator("test")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "custom validator")
			},
		},
		{
			name: "should_mix_user_and_default_options",
			userOpts: &InvokeOptions{
				InvocationType: types.InvocationTypeEvent,
				Timeout:        45 * time.Second,
				// LogType, MaxRetries, RetryDelay, PayloadValidator should use defaults
			},
			verify: func(t *testing.T, result InvokeOptions) {
				defaults := defaultInvokeOptions()
				assert.Equal(t, types.InvocationTypeEvent, result.InvocationType)
				assert.Equal(t, 45*time.Second, result.Timeout)
				assert.Equal(t, defaults.LogType, result.LogType)
				assert.Equal(t, defaults.MaxRetries, result.MaxRetries)
				assert.Equal(t, defaults.RetryDelay, result.RetryDelay)
				assert.NotNil(t, result.PayloadValidator)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := mergeInvokeOptions(tc.userOpts)

			// Assert
			tc.verify(t, result)
		})
	}
}

func TestCalculateBackoffDelay(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}

	testCases := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "should_return_base_delay_for_zero_attempt",
			attempt:  0,
			expected: time.Second,
		},
		{
			name:     "should_return_base_delay_for_negative_attempt",
			attempt:  -1,
			expected: time.Second,
		},
		{
			name:     "should_calculate_exponential_delay_for_first_attempt",
			attempt:  1,
			expected: 2 * time.Second,
		},
		{
			name:     "should_calculate_exponential_delay_for_second_attempt",
			attempt:  2,
			expected: 4 * time.Second,
		},
		{
			name:     "should_calculate_exponential_delay_for_third_attempt",
			attempt:  3,
			expected: 8 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := calculateBackoffDelay(tc.attempt, config)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalculateBackoffDelayWithMaxDelayLimit(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 10,
		BaseDelay:   time.Second,
		MaxDelay:    5 * time.Second, // Small max delay to test capping
		Multiplier:  2.0,
	}

	// Test that delay is capped at MaxDelay
	result := calculateBackoffDelay(5, config) // Should result in 32 seconds normally
	assert.Equal(t, 5*time.Second, result)    // But capped at 5 seconds
}

func TestCalculateBackoffDelayOverflowProtection(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 100,
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  10.0, // Very high multiplier to test overflow protection
	}

	// Test with a high attempt number that could cause overflow
	result := calculateBackoffDelay(50, config)
	assert.Equal(t, 30*time.Second, result) // Should be capped at MaxDelay
	assert.True(t, result > 0)              // Should not overflow to negative
}

func TestSanitizeLogResult(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "should_return_empty_for_empty_input",
			input:    "",
			expected: "",
		},
		{
			name:     "should_preserve_regular_log_lines",
			input:    "INFO: Application started\nERROR: Something went wrong",
			expected: "INFO: Application started\nERROR: Something went wrong",
		},
		{
			name: "should_remove_request_id_lines",
			input: `INFO: Application started
RequestId: 12345-67890-abcdef
ERROR: Something went wrong`,
			expected: "INFO: Application started\nERROR: Something went wrong",
		},
		{
			name: "should_remove_duration_lines",
			input: `INFO: Application started
Duration: 123.45 ms
ERROR: Something went wrong`,
			expected: "INFO: Application started\nERROR: Something went wrong",
		},
		{
			name: "should_remove_billed_duration_lines",
			input: `INFO: Application started
Billed Duration: 200 ms
ERROR: Something went wrong`,
			expected: "INFO: Application started\nERROR: Something went wrong",
		},
		{
			name: "should_remove_multiple_aws_internal_lines",
			input: `START RequestId: 12345-67890-abcdef
INFO: Application started
Duration: 123.45 ms
Billed Duration: 200 ms
ERROR: Something went wrong
END RequestId: 12345-67890-abcdef`,
			expected: "INFO: Application started\nERROR: Something went wrong",
		},
		{
			name:     "should_handle_single_line_without_newline",
			input:    "INFO: Single line log",
			expected: "INFO: Single line log",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := sanitizeLogResult(tc.input)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestLogOperation(t *testing.T) {
	// This test verifies that logOperation doesn't panic and can handle various inputs
	testCases := []struct {
		name         string
		operation    string
		functionName string
		details      map[string]interface{}
	}{
		{
			name:         "should_handle_basic_operation",
			operation:    "invoke",
			functionName: "test-function",
			details:      map[string]interface{}{"timeout": 30},
		},
		{
			name:         "should_handle_empty_details",
			operation:    "get_function",
			functionName: "test-function",
			details:      map[string]interface{}{},
		},
		{
			name:         "should_handle_nil_details",
			operation:    "list_functions",
			functionName: "test-function",
			details:      nil,
		},
		{
			name:         "should_handle_complex_details",
			operation:    "create_function",
			functionName: "test-function",
			details: map[string]interface{}{
				"runtime":     "nodejs18.x",
				"timeout":     30,
				"memory_size": 512,
				"handler":     "index.handler",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act & Assert - should not panic
			assert.NotPanics(t, func() {
				logOperation(tc.operation, tc.functionName, tc.details)
			})
		})
	}
}

func TestLogResult(t *testing.T) {
	// This test verifies that logResult doesn't panic and can handle various inputs
	testCases := []struct {
		name         string
		operation    string
		functionName string
		success      bool
		duration     time.Duration
		err          error
	}{
		{
			name:         "should_handle_successful_operation",
			operation:    "invoke",
			functionName: "test-function",
			success:      true,
			duration:     100 * time.Millisecond,
			err:          nil,
		},
		{
			name:         "should_handle_failed_operation",
			operation:    "get_function",
			functionName: "test-function",
			success:      false,
			duration:     50 * time.Millisecond,
			err:          errors.New("function not found"),
		},
		{
			name:         "should_handle_zero_duration",
			operation:    "list_functions",
			functionName: "test-function",
			success:      true,
			duration:     0,
			err:          nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act & Assert - should not panic
			assert.NotPanics(t, func() {
				logResult(tc.operation, tc.functionName, tc.success, tc.duration, tc.err)
			})
		})
	}
}

// TestConstants verifies that all constants are defined correctly
func TestConstants(t *testing.T) {
	// Test default values
	assert.Equal(t, 30*time.Second, DefaultTimeout)
	assert.Equal(t, int32(128), DefaultMemorySize)
	assert.Equal(t, 6*1024*1024, MaxPayloadSize)
	assert.Equal(t, 3, DefaultRetryAttempts)
	assert.Equal(t, time.Second, DefaultRetryDelay)

	// Test invocation types
	assert.Equal(t, types.InvocationTypeRequestResponse, InvocationTypeRequestResponse)
	assert.Equal(t, types.InvocationTypeEvent, InvocationTypeEvent)
	assert.Equal(t, types.InvocationTypeDryRun, InvocationTypeDryRun)

	// Test log types
	assert.Equal(t, types.LogTypeNone, LogTypeNone)
	assert.Equal(t, types.LogTypeTail, LogTypeTail)
}

// TestErrors verifies that all error constants are defined correctly
func TestErrors(t *testing.T) {
	assert.Error(t, ErrFunctionNotFound)
	assert.Error(t, ErrInvalidPayload)
	assert.Error(t, ErrInvocationTimeout)
	assert.Error(t, ErrInvocationFailed)
	assert.Error(t, ErrLogRetrievalFailed)
	assert.Error(t, ErrInvalidFunctionName)

	// Test error messages
	assert.Contains(t, ErrFunctionNotFound.Error(), "lambda function not found")
	assert.Contains(t, ErrInvalidPayload.Error(), "invalid payload format")
	assert.Contains(t, ErrInvocationTimeout.Error(), "lambda invocation timeout")
	assert.Contains(t, ErrInvocationFailed.Error(), "lambda invocation failed")
	assert.Contains(t, ErrLogRetrievalFailed.Error(), "failed to retrieve lambda logs")
	assert.Contains(t, ErrInvalidFunctionName.Error(), "invalid function name")
}

func TestPayloadValidatorIntegration(t *testing.T) {
	// Test default payload validator with various payload sizes
	validator := defaultInvokeOptions().PayloadValidator

	testCases := []struct {
		name        string
		payloadSize int
		expectError bool
	}{
		{
			name:        "should_accept_small_payload",
			payloadSize: 1024,
			expectError: false,
		},
		{
			name:        "should_accept_maximum_allowed_payload",
			payloadSize: MaxPayloadSize,
			expectError: false,
		},
		{
			name:        "should_reject_oversized_payload",
			payloadSize: MaxPayloadSize + 1,
			expectError: true,
		},
		{
			name:        "should_accept_empty_payload",
			payloadSize: 0,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			payload := strings.Repeat("a", tc.payloadSize)

			// Act
			err := validator(payload)

			// Assert
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "payload size exceeds maximum")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}