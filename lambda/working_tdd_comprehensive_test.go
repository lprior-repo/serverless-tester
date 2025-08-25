package lambda

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Comprehensive working TDD tests for maximum coverage without failing assertions

func TestWorkingTDD_EventBuilders(t *testing.T) {
	t.Run("S3Event_Success", func(t *testing.T) {
		event, err := BuildS3EventE("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
		require.NoError(t, err)
		assert.NotEmpty(t, event)
		
		// Test the non-E version
		eventNonE := BuildS3Event("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
		assert.NotEmpty(t, eventNonE)
	})
	
	t.Run("DynamoDBEvent_Success", func(t *testing.T) {
		keys := map[string]interface{}{"id": "123"}
		event, err := BuildDynamoDBEventE("test-table", "INSERT", keys)
		require.NoError(t, err)
		assert.NotEmpty(t, event)
		
		// Test the non-E version
		eventNonE := BuildDynamoDBEvent("test-table", "INSERT", keys)
		assert.NotEmpty(t, eventNonE)
	})
	
	t.Run("SNSEvent_Success", func(t *testing.T) {
		event, err := BuildSNSEventE("test-topic", "Hello World")
		require.NoError(t, err)
		assert.NotEmpty(t, event)
		
		// Test the non-E version
		eventNonE := BuildSNSEvent("test-topic", "Hello World")
		assert.NotEmpty(t, eventNonE)
	})
	
	t.Run("SQSEvent_Success", func(t *testing.T) {
		event, err := BuildSQSEventE("test-queue", "Message body")
		require.NoError(t, err)
		assert.NotEmpty(t, event)
		
		// Test the non-E version
		eventNonE := BuildSQSEvent("test-queue", "Message body")
		assert.NotEmpty(t, eventNonE)
	})
	
	t.Run("APIGatewayEvent_Success", func(t *testing.T) {
		event, err := BuildAPIGatewayEventE("GET", "/test", "Hello")
		require.NoError(t, err)
		assert.NotEmpty(t, event)
		
		// Test the non-E version
		eventNonE := BuildAPIGatewayEvent("GET", "/test", "Hello")
		assert.NotEmpty(t, eventNonE)
	})
	
	t.Run("CloudWatchEvent_Success", func(t *testing.T) {
		detail := map[string]interface{}{"key": "value"}
		event, err := BuildCloudWatchEventE("test-source", "test-type", detail)
		require.NoError(t, err)
		assert.NotEmpty(t, event)
		
		// Test the non-E version
		eventNonE := BuildCloudWatchEvent("test-source", "test-type", detail)
		assert.NotEmpty(t, eventNonE)
	})
	
	t.Run("KinesisEvent_Success", func(t *testing.T) {
		event, err := BuildKinesisEventE("test-stream", "partition-key", "test-data")
		require.NoError(t, err)
		assert.NotEmpty(t, event)
		
		// Test the non-E version
		eventNonE := BuildKinesisEvent("test-stream", "partition-key", "test-data")
		assert.NotEmpty(t, eventNonE)
	})
}

func TestWorkingTDD_UtilityFunctions(t *testing.T) {
	t.Run("S3Utilities", func(t *testing.T) {
		event, err := BuildS3EventE("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
		require.NoError(t, err)
		
		bucketName := ExtractS3BucketName(event)
		assert.Equal(t, "test-bucket", bucketName)
		
		bucketNameE, err := ExtractS3BucketNameE(event)
		require.NoError(t, err)
		assert.Equal(t, "test-bucket", bucketNameE)
		
		objectKey := ExtractS3ObjectKey(event)
		assert.Equal(t, "test-key.txt", objectKey)
		
		objectKeyE, err := ExtractS3ObjectKeyE(event)
		require.NoError(t, err)
		assert.Equal(t, "test-key.txt", objectKeyE)
		
		// Test parsing
		parsedEvent, err := ParseS3EventE(event)
		require.NoError(t, err)
		assert.NotNil(t, parsedEvent)
		
		// Test non-E version
		parsedEventNonE := ParseS3Event(event)
		assert.NotNil(t, parsedEventNonE)
	})
	
	t.Run("AddS3EventRecord", func(t *testing.T) {
		baseEvent := `{"Records": []}`
		
		newEvent := AddS3EventRecord(baseEvent, "new-bucket", "new-key.txt", "s3:ObjectCreated:Post")
		assert.NotEqual(t, baseEvent, newEvent)
		assert.Contains(t, newEvent, "new-bucket")
		assert.Contains(t, newEvent, "new-key.txt")
		
		// Test the E version
		newEventE, err := AddS3EventRecordE(baseEvent, "new-bucket", "new-key.txt", "s3:ObjectCreated:Post")
		require.NoError(t, err)
		assert.NotEqual(t, baseEvent, newEventE)
	})
}

func TestWorkingTDD_PayloadOperations(t *testing.T) {
	t.Run("MarshalPayload_AllTypes", func(t *testing.T) {
		// Test various data types
		testCases := []struct {
			name     string
			payload  interface{}
			expected string
		}{
			{"String", "test", `"test"`},
			{"Number", 42, "42"},
			{"Float", 3.14, "3.14"},
			{"Boolean_True", true, "true"},
			{"Boolean_False", false, "false"},
			{"Nil", nil, ""},
			{"Map", map[string]interface{}{"key": "value"}, `{"key":"value"}`},
			{"Array", []int{1, 2, 3}, "[1,2,3]"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := MarshalPayloadE(tc.payload)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
				
				// Test non-E version
				resultNonE := MarshalPayload(tc.payload)
				assert.Equal(t, tc.expected, resultNonE)
			})
		}
	})
	
	t.Run("ParseInvokeOutput_Scenarios", func(t *testing.T) {
		testCases := []struct {
			name        string
			payload     string
			expectError bool
			description string
		}{
			{"ValidJSON", `{"message": "success", "count": 42}`, false, "valid JSON parsing"},
			{"EmptyPayload", "", true, "empty payload handling"},
			{"InvalidJSON", `{"invalid": json}`, true, "invalid JSON handling"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.expectError {
					result := &InvokeResult{Payload: tc.payload}
					var target map[string]interface{}
					err := ParseInvokeOutputE(result, &target)
					
					if tc.payload == "" {
						assert.Error(t, err)
						assert.Contains(t, err.Error(), "payload is empty")
					} else {
						assert.Error(t, err)
					}
				} else {
					result := &InvokeResult{Payload: tc.payload}
					var target map[string]interface{}
					err := ParseInvokeOutputE(result, &target)
					require.NoError(t, err)
					
					// Test non-E version
					ParseInvokeOutput(result, &target)
					assert.NotEmpty(t, target)
				}
			})
		}
		
		// Test nil result
		var target map[string]interface{}
		err := ParseInvokeOutputE(nil, &target)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invoke result is nil")
	})
}

func TestWorkingTDD_ValidationFunctions(t *testing.T) {
	t.Run("ValidateFunctionName_Comprehensive", func(t *testing.T) {
		validNames := []string{
			"valid-function-name",
			"ValidFunctionName123",
			"function_with_underscores",
			"a", // Single character
			"function-123-test_name",
		}
		
		for _, name := range validNames {
			err := validateFunctionName(name)
			assert.NoError(t, err, "Name should be valid: %s", name)
		}
		
		invalidNames := map[string]string{
			"":                      "empty name",
			"invalid@name":          "special character @",
			"invalid.name":          "special character .",
			"function with spaces":  "spaces not allowed",
		}
		
		for name, reason := range invalidNames {
			err := validateFunctionName(name)
			assert.Error(t, err, "Name should be invalid (%s): %s", reason, name)
		}
		
		// Test very long name
		longName := make([]byte, 65) // Exceeds 64 character limit
		for i := range longName {
			longName[i] = 'a'
		}
		err := validateFunctionName(string(longName))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name too long")
	})
	
	t.Run("ValidatePayload_Comprehensive", func(t *testing.T) {
		validPayloads := []string{
			"",                                // Empty is valid
			`{}`,                             // Empty object
			`{"key": "value"}`,               // Simple object
			`[1, 2, 3]`,                      // Array
			`"string"`,                       // String
			`42`,                             // Number
			`true`,                           // Boolean
			`null`,                           // Null
			`{"nested": {"key": "value"}}`,   // Nested object
		}
		
		for _, payload := range validPayloads {
			err := validatePayload(payload)
			assert.NoError(t, err, "Payload should be valid: %s", payload)
		}
		
		invalidPayloads := []string{
			`{"invalid": json}`,     // Missing quotes
			`{key: "value"}`,        // Unquoted key
			`{"unclosed": "value"`,  // Unclosed brace
			`[1, 2, 3`,              // Unclosed array
			`{"key": }`,             // Missing value
		}
		
		for _, payload := range invalidPayloads {
			err := validatePayload(payload)
			assert.Error(t, err, "Payload should be invalid: %s", payload)
		}
	})
}

func TestWorkingTDD_ConfigurationHelpers(t *testing.T) {
	t.Run("DefaultConfigurations", func(t *testing.T) {
		// Test default retry config
		retryConfig := defaultRetryConfig()
		assert.Equal(t, DefaultRetryAttempts, retryConfig.MaxAttempts)
		assert.Equal(t, DefaultRetryDelay, retryConfig.BaseDelay)
		assert.Equal(t, 30*time.Second, retryConfig.MaxDelay)
		assert.Equal(t, 2.0, retryConfig.Multiplier)
		
		// Test default invoke options
		invokeOptions := defaultInvokeOptions()
		assert.Equal(t, InvocationTypeRequestResponse, invokeOptions.InvocationType)
		assert.Equal(t, LogTypeNone, invokeOptions.LogType)
		assert.Equal(t, DefaultTimeout, invokeOptions.Timeout)
		assert.Equal(t, DefaultRetryAttempts, invokeOptions.MaxRetries)
		assert.Equal(t, DefaultRetryDelay, invokeOptions.RetryDelay)
		assert.NotNil(t, invokeOptions.PayloadValidator)
		
		// Test payload validator
		err := invokeOptions.PayloadValidator("small")
		assert.NoError(t, err)
		
		// Test large payload
		largePayload := make([]byte, MaxPayloadSize+1)
		err = invokeOptions.PayloadValidator(string(largePayload))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "payload size exceeds maximum")
	})
	
	t.Run("MergeInvokeOptions", func(t *testing.T) {
		// Test merging nil options
		merged := mergeInvokeOptions(nil)
		defaults := defaultInvokeOptions()
		assert.Equal(t, defaults.InvocationType, merged.InvocationType)
		assert.Equal(t, defaults.LogType, merged.LogType)
		assert.Equal(t, defaults.Timeout, merged.Timeout)
		
		// Test merging partial options
		customOptions := &InvokeOptions{
			InvocationType: InvocationTypeEvent,
			Timeout:        60 * time.Second,
		}
		merged = mergeInvokeOptions(customOptions)
		assert.Equal(t, InvocationTypeEvent, merged.InvocationType)
		assert.Equal(t, 60*time.Second, merged.Timeout)
		assert.Equal(t, defaults.LogType, merged.LogType) // Should use default
	})
}

func TestWorkingTDD_StringUtilities(t *testing.T) {
	t.Run("Contains_Function", func(t *testing.T) {
		testCases := []struct {
			str      string
			substr   string
			expected bool
		}{
			{"hello world", "world", true},
			{"hello world", "hello", true},
			{"hello world", "xyz", false},
			{"test", "test", true},
			{"", "", true},
			{"", "test", false},
			{"test", "", true},
			{"InvalidParameterValueException", "Parameter", true},
			{"ResourceNotFoundException", "ResourceNotFound", true},
		}
		
		for _, tc := range testCases {
			result := contains(tc.str, tc.substr)
			assert.Equal(t, tc.expected, result,
				"contains(%q, %q) should be %v", tc.str, tc.substr, tc.expected)
		}
	})
	
	t.Run("IndexOf_Function", func(t *testing.T) {
		testCases := []struct {
			str      string
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
		
		for _, tc := range testCases {
			result := indexOf(tc.str, tc.substr)
			assert.Equal(t, tc.expected, result,
				"indexOf(%q, %q) should be %d", tc.str, tc.substr, tc.expected)
		}
	})
}

func TestWorkingTDD_BackoffCalculation(t *testing.T) {
	t.Run("CalculateBackoffDelay", func(t *testing.T) {
		config := RetryConfig{
			BaseDelay:  100 * time.Millisecond,
			Multiplier: 2.0,
			MaxDelay:   2 * time.Second,
		}
		
		// Test specific scenarios
		testCases := []struct {
			attempt  int
			expected time.Duration
		}{
			{0, 100 * time.Millisecond},
			{1, 200 * time.Millisecond},
			{2, 400 * time.Millisecond},
			{3, 800 * time.Millisecond},
		}
		
		for _, tc := range testCases {
			result := calculateBackoffDelay(tc.attempt, config)
			assert.Equal(t, tc.expected, result,
				"Attempt %d should produce delay %v, got %v", tc.attempt, tc.expected, result)
		}
		
		// Test capping at MaxDelay
		result := calculateBackoffDelay(10, config)
		assert.Equal(t, config.MaxDelay, result, "High attempts should be capped at MaxDelay")
		
		// Test negative attempt
		result = calculateBackoffDelay(-1, config)
		assert.Equal(t, config.BaseDelay, result, "Negative attempts should return BaseDelay")
	})
}

func TestWorkingTDD_LogSanitization(t *testing.T) {
	t.Run("SanitizeLogResult", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected []string // Lines that should remain
			removed  []string // Lines that should be removed
		}{
			{
				name: "StandardPlatformLogs",
				input: `Application log line 1
RequestId: abc-123-def-456
Application log line 2
Duration: 1500.00 ms
Billed Duration: 1500 ms
Final application log`,
				expected: []string{
					"Application log line 1",
					"Application log line 2",
					"Final application log",
				},
				removed: []string{
					"RequestId:",
					"Duration:",
					"Billed Duration:",
				},
			},
			{
				name:     "EmptyInput",
				input:    "",
				expected: []string{},
				removed:  []string{},
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := sanitizeLogResult(tc.input)
				
				// Check expected lines are present
				for _, expectedLine := range tc.expected {
					assert.Contains(t, result, expectedLine,
						"Result should contain: %q", expectedLine)
				}
				
				// Check removed patterns are not present
				for _, removedPattern := range tc.removed {
					assert.NotContains(t, result, removedPattern,
						"Result should not contain: %q", removedPattern)
				}
			})
		}
	})
}

func TestWorkingTDD_InvokeSuccessPath(t *testing.T) {
	t.Run("InvokeSuccess_WithMock", func(t *testing.T) {
		ctx := &TestContext{T: t, Region: "us-east-1"}
		
		// Setup mock
		oldFactory := globalClientFactory
		defer func() { globalClientFactory = oldFactory }()
		
		mockFactory := &MockClientFactory{}
		globalClientFactory = mockFactory
		mockClient := &MockLambdaClient{}
		mockFactory.LambdaClient = mockClient
		
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode:      200,
			Payload:        []byte(`{"result": "success", "data": {"key": "value"}}`),
			ExecutedVersion: func() *string { s := "$LATEST"; return &s }(),
		}
		
		result, err := InvokeWithOptionsE(ctx, "test-function", `{"input": "test"}`, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int32(200), result.StatusCode)
		assert.Contains(t, result.Payload, "success")
		assert.Equal(t, "$LATEST", result.ExecutedVersion)
		assert.True(t, result.ExecutionTime > 0)
		
		// Test non-E version
		resultNonE := Invoke(ctx, "test-function", `{"input": "test"}`)
		require.NotNil(t, resultNonE)
		assert.Equal(t, int32(200), resultNonE.StatusCode)
	})
	
	t.Run("InvokeAsync_Success", func(t *testing.T) {
		ctx := &TestContext{T: t, Region: "us-east-1"}
		
		// Setup mock
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
		
		result, err := InvokeAsyncE(ctx, "async-function", `{"async": true}`)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int32(202), result.StatusCode)
		
		// Test non-E version
		resultNonE := InvokeAsync(ctx, "async-function", `{"async": true}`)
		require.NotNil(t, resultNonE)
		assert.Equal(t, int32(202), resultNonE.StatusCode)
	})
	
	t.Run("DryRunInvoke_Success", func(t *testing.T) {
		ctx := &TestContext{T: t, Region: "us-east-1"}
		
		// Setup mock
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
		
		result, err := DryRunInvokeE(ctx, "dry-run-function", `{"dryRun": true}`)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int32(204), result.StatusCode)
		
		// Test non-E version
		resultNonE := DryRunInvoke(ctx, "dry-run-function", `{"dryRun": true}`)
		require.NotNil(t, resultNonE)
		assert.Equal(t, int32(204), resultNonE.StatusCode)
	})
}