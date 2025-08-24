package lambda

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test suite focusing on pure functions and utilities that don't require AWS clients

// Test function name validation
func TestValidateFunctionName_Comprehensive(t *testing.T) {
	testCases := []struct {
		name         string
		functionName string
		expectError  bool
		errorSubstring string
	}{
		{"valid simple name", "test-function", false, ""},
		{"valid with numbers", "test123", false, ""},
		{"valid with underscores", "test_function", false, ""},
		{"valid with dashes", "test-function-name", false, ""},
		{"valid mixed characters", "Test_Function-123", false, ""},
		{"valid exactly 64 chars", strings.Repeat("a", 64), false, ""},
		{"empty name", "", true, "invalid function name"},
		{"too long", strings.Repeat("a", 65), true, "name too long"},
		{"invalid characters", "test@function", true, "invalid character"},
		{"spaces", "test function", true, "invalid character"},
		{"unicode", "test函数", true, "invalid character"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Validating function name
			err := validateFunctionName(tc.functionName)
			
			// Then: Should match expectation
			if tc.expectError {
				assert.Error(t, err, "Expected error for function name: %s", tc.functionName)
				if tc.errorSubstring != "" {
					assert.Contains(t, err.Error(), tc.errorSubstring)
				}
			} else {
				assert.NoError(t, err, "Expected no error for function name: %s", tc.functionName)
			}
		})
	}
}

// Test payload validation
func TestValidatePayload_Comprehensive(t *testing.T) {
	testCases := []struct {
		name        string
		payload     string
		expectError bool
	}{
		{"empty payload", "", false},
		{"valid JSON object", `{"key": "value"}`, false},
		{"valid JSON array", `[1, 2, 3]`, false},
		{"valid JSON string", `"hello"`, false},
		{"valid JSON number", `42`, false},
		{"valid JSON boolean", `true`, false},
		{"valid JSON null", `null`, false},
		{"valid nested object", `{"user": {"name": "John", "age": 30}}`, false},
		{"large valid JSON", `{"key":"` + strings.Repeat("x", 1000) + `"}`, false},
		{"invalid JSON missing brace", `{"key": "value"`, true},
		{"malformed JSON", `{key: value}`, true},
		{"single quotes", `{'key': 'value'}`, true},
		{"trailing comma", `{"key": "value",}`, true},
		{"binary data", string([]byte{0, 1, 2, 3, 4, 5}), true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Validating payload
			err := validatePayload(tc.payload)
			
			// Then: Should match expectation
			if tc.expectError {
				assert.Error(t, err, "Expected error for payload: %s", tc.name)
			} else {
				assert.NoError(t, err, "Expected no error for payload: %s", tc.name)
			}
		})
	}
}

// Test payload marshaling
func TestMarshalPayloadE_Comprehensive(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello", `"hello"`},
		{"number", 42, "42"},
		{"boolean true", true, "true"},
		{"boolean false", false, "false"},
		{"object", map[string]interface{}{"key": "value"}, `{"key":"value"}`},
		{"array", []int{1, 2, 3}, "[1,2,3]"},
		{"nested object", map[string]interface{}{
			"user": map[string]interface{}{
				"name": "John",
				"age":  30,
			},
		}, `{"user":{"age":30,"name":"John"}}`},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Marshaling payload
			result, err := MarshalPayloadE(tc.input)
			
			// Then: Should succeed and match expected output
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Test payload parsing
func TestParseInvokeOutputE_Comprehensive(t *testing.T) {
	testCases := []struct {
		name        string
		result      *InvokeResult
		target      interface{}
		expectError bool
		validate    func(t *testing.T, target interface{})
	}{
		{
			"parse into map",
			&InvokeResult{StatusCode: 200, Payload: `{"message": "success", "count": 42}`},
			&map[string]interface{}{},
			false,
			func(t *testing.T, target interface{}) {
				result := target.(*map[string]interface{})
				assert.Equal(t, "success", (*result)["message"])
				assert.Equal(t, float64(42), (*result)["count"]) // JSON numbers are float64
			},
		},
		{
			"parse into struct",
			&InvokeResult{StatusCode: 200, Payload: `{"message": "success", "count": 42}`},
			&struct {
				Message string `json:"message"`
				Count   int    `json:"count"`
			}{},
			false,
			func(t *testing.T, target interface{}) {
				result := target.(*struct {
					Message string `json:"message"`
					Count   int    `json:"count"`
				})
				assert.Equal(t, "success", result.Message)
				assert.Equal(t, 42, result.Count)
			},
		},
		{
			"nil result",
			nil,
			&map[string]interface{}{},
			true,
			nil,
		},
		{
			"empty payload",
			&InvokeResult{StatusCode: 200, Payload: ""},
			&map[string]interface{}{},
			true,
			nil,
		},
		{
			"invalid JSON",
			&InvokeResult{StatusCode: 200, Payload: `{"invalid": json}`},
			&map[string]interface{}{},
			true,
			nil,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Parsing invoke output
			err := ParseInvokeOutputE(tc.result, tc.target)
			
			// Then: Should match expectation
			if tc.expectError {
				assert.Error(t, err, "Expected error for: %s", tc.name)
			} else {
				assert.NoError(t, err, "Expected no error for: %s", tc.name)
				if tc.validate != nil {
					tc.validate(t, tc.target)
				}
			}
		})
	}
}

// Test S3 event building
func TestBuildS3EventE_Comprehensive(t *testing.T) {
	// Given: Valid S3 event parameters
	bucketName := "test-bucket"
	objectKey := "test-object.txt"
	eventName := "s3:ObjectCreated:Put"
	
	// When: Building S3 event
	eventJSON, err := BuildS3EventE(bucketName, objectKey, eventName)
	
	// Then: Should create valid S3 event
	require.NoError(t, err)
	assert.NotEmpty(t, eventJSON)
	
	// Validate event structure
	var event S3Event
	err = json.Unmarshal([]byte(eventJSON), &event)
	require.NoError(t, err)
	
	assert.Len(t, event.Records, 1)
	record := event.Records[0]
	assert.Equal(t, eventName, record.EventName)
	assert.Equal(t, "aws:s3", record.EventSource)
	assert.Equal(t, "2.1", record.EventVersion)
	assert.Equal(t, bucketName, record.S3.Bucket.Name)
	assert.Equal(t, objectKey, record.S3.Object.Key)
	assert.Equal(t, int64(1024), record.S3.Object.Size)
	assert.NotEmpty(t, record.EventTime)
	
	// Validate it's valid JSON
	assert.NoError(t, validatePayload(eventJSON))
}

// Test DynamoDB event building
func TestBuildDynamoDBEventE_Comprehensive(t *testing.T) {
	// Given: Valid DynamoDB event parameters
	tableName := "test-table"
	eventName := "INSERT"
	keys := map[string]interface{}{
		"id": map[string]interface{}{"S": "test-id"},
	}
	
	// When: Building DynamoDB event
	eventJSON, err := BuildDynamoDBEventE(tableName, eventName, keys)
	
	// Then: Should create valid DynamoDB event
	require.NoError(t, err)
	assert.NotEmpty(t, eventJSON)
	
	// Validate event structure
	var event DynamoDBEvent
	err = json.Unmarshal([]byte(eventJSON), &event)
	require.NoError(t, err)
	
	assert.Len(t, event.Records, 1)
	record := event.Records[0]
	assert.Equal(t, eventName, record.EventName)
	assert.Equal(t, "aws:dynamodb", record.EventSource)
	assert.Equal(t, "1.1", record.EventVersion)
	assert.Equal(t, "us-east-1", record.AwsRegion)
	assert.NotNil(t, record.Dynamodb.Keys)
	assert.Equal(t, "111", record.Dynamodb.SequenceNumber)
	assert.Equal(t, int64(26), record.Dynamodb.SizeBytes)
	assert.Equal(t, "KEYS_ONLY", record.Dynamodb.StreamViewType)
	
	// Validate it's valid JSON
	assert.NoError(t, validatePayload(eventJSON))
}

// Test SQS event building
func TestBuildSQSEventE_Comprehensive(t *testing.T) {
	// Given: Valid SQS event parameters
	queueUrl := "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
	messageBody := "Hello from SQS"
	
	// When: Building SQS event
	eventJSON, err := BuildSQSEventE(queueUrl, messageBody)
	
	// Then: Should create valid SQS event
	require.NoError(t, err)
	assert.NotEmpty(t, eventJSON)
	
	// Validate event structure
	var event SQSEvent
	err = json.Unmarshal([]byte(eventJSON), &event)
	require.NoError(t, err)
	
	assert.Len(t, event.Records, 1)
	record := event.Records[0]
	assert.Equal(t, messageBody, record.Body)
	assert.Equal(t, "aws:sqs", record.EventSource)
	assert.Equal(t, queueUrl, record.EventSourceARN)
	assert.Equal(t, "us-east-1", record.AwsRegion)
	assert.NotEmpty(t, record.MessageId)
	assert.NotEmpty(t, record.Md5OfBody)
	assert.NotNil(t, record.Attributes)
	assert.NotNil(t, record.MessageAttributes)
	
	// Validate it's valid JSON
	assert.NoError(t, validatePayload(eventJSON))
}

// Test SNS event building
func TestBuildSNSEventE_Comprehensive(t *testing.T) {
	// Given: Valid SNS event parameters
	topicArn := "arn:aws:sns:us-east-1:123456789012:test-topic"
	message := "Hello from SNS"
	
	// When: Building SNS event
	eventJSON, err := BuildSNSEventE(topicArn, message)
	
	// Then: Should create valid SNS event
	require.NoError(t, err)
	assert.NotEmpty(t, eventJSON)
	
	// Validate event structure
	var event SNSEvent
	err = json.Unmarshal([]byte(eventJSON), &event)
	require.NoError(t, err)
	
	assert.Len(t, event.Records, 1)
	record := event.Records[0]
	assert.Equal(t, message, record.Sns.Message)
	assert.Equal(t, "aws:sns", record.EventSource)
	assert.Equal(t, topicArn, record.Sns.TopicArn)
	assert.Equal(t, "1.0", record.EventVersion)
	assert.Equal(t, "Notification", record.Sns.Type)
	assert.NotEmpty(t, record.Sns.MessageId)
	assert.NotEmpty(t, record.Sns.Timestamp)
	assert.NotNil(t, record.Sns.MessageAttributes)
	
	// Validate it's valid JSON
	assert.NoError(t, validatePayload(eventJSON))
}

// Test API Gateway event building
func TestBuildAPIGatewayEventE_Comprehensive(t *testing.T) {
	// Given: Valid API Gateway event parameters
	httpMethod := "POST"
	path := "/api/test"
	body := `{"key": "value"}`
	
	// When: Building API Gateway event
	eventJSON, err := BuildAPIGatewayEventE(httpMethod, path, body)
	
	// Then: Should create valid API Gateway event
	require.NoError(t, err)
	assert.NotEmpty(t, eventJSON)
	
	// Validate event structure
	var event APIGatewayEvent
	err = json.Unmarshal([]byte(eventJSON), &event)
	require.NoError(t, err)
	
	assert.Equal(t, httpMethod, event.HttpMethod)
	assert.Equal(t, path, event.Path)
	assert.Equal(t, body, event.Body)
	assert.Equal(t, "/{proxy+}", event.Resource)
	assert.False(t, event.IsBase64Encoded)
	assert.NotNil(t, event.Headers)
	assert.NotNil(t, event.RequestContext)
	assert.Equal(t, httpMethod, event.RequestContext.HttpMethod)
	assert.NotEmpty(t, event.RequestContext.RequestId)
	assert.NotEmpty(t, event.RequestContext.AccountId)
	
	// Validate it's valid JSON
	assert.NoError(t, validatePayload(eventJSON))
}

// Test event structure validation
func TestValidateEventStructure_Comprehensive(t *testing.T) {
	testCases := []struct {
		name         string
		eventJSON    string
		eventType    string
		expectErrors bool
	}{
		{
			"valid S3 event",
			BuildS3Event("test-bucket", "test.txt", "s3:ObjectCreated:Put"),
			"s3",
			false,
		},
		{
			"valid DynamoDB event", 
			BuildDynamoDBEvent("test-table", "INSERT", map[string]interface{}{"id": map[string]interface{}{"S": "123"}}),
			"dynamodb",
			false,
		},
		{
			"valid SQS event",
			BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123456789012/queue", "message"),
			"sqs",
			false,
		},
		{
			"valid SNS event",
			BuildSNSEvent("arn:aws:sns:us-east-1:123456789012:topic", "message"),
			"sns",
			false,
		},
		{
			"valid API Gateway event",
			BuildAPIGatewayEvent("GET", "/test", ""),
			"apigateway",
			false,
		},
		{
			"invalid S3 event - no records",
			`{"Records": []}`,
			"s3",
			true,
		},
		{
			"invalid DynamoDB event - wrong source",
			`{"Records": [{"eventSource": "aws:sqs"}]}`,
			"dynamodb",
			true,
		},
		{
			"unknown event type",
			`{"Records": []}`,
			"unknown",
			true,
		},
		{
			"invalid JSON",
			`{invalid json}`,
			"s3",
			true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Validating event structure
			errors := ValidateEventStructure(tc.eventJSON, tc.eventType)
			
			// Then: Should match expectation
			if tc.expectErrors {
				assert.NotEmpty(t, errors, "Expected validation errors for %s", tc.name)
			} else {
				assert.Empty(t, errors, "Expected no validation errors for %s", tc.name)
			}
		})
	}
}

// Test log filtering functionality
func TestLogFiltering(t *testing.T) {
	// Given: Log entries with different content and levels
	logs := []LogEntry{
		{Message: "This is an error message", Level: "ERROR", Timestamp: time.Now()},
		{Message: "This is an info message", Level: "INFO", Timestamp: time.Now().Add(1 * time.Second)},
		{Message: "This is another error", Level: "ERROR", Timestamp: time.Now().Add(2 * time.Second)},
		{Message: "Debug information here", Level: "DEBUG", Timestamp: time.Now().Add(3 * time.Second)},
		{Message: "Processing user data", Level: "INFO", Timestamp: time.Now().Add(4 * time.Second)},
	}
	
	t.Run("FilterLogs", func(t *testing.T) {
		// When: Filtering logs by pattern
		errorLogs := FilterLogs(logs, "error")
		
		// Then: Should return matching entries
		assert.Len(t, errorLogs, 2)
		assert.Contains(t, errorLogs[0].Message, "error")
		assert.Contains(t, errorLogs[1].Message, "error")
		
		// Test case-insensitive filtering
		userLogs := FilterLogs(logs, "user")
		assert.Len(t, userLogs, 1)
		assert.Contains(t, userLogs[0].Message, "user")
	})
	
	t.Run("FilterLogsByLevel", func(t *testing.T) {
		// When: Filtering logs by level
		errorLogs := FilterLogsByLevel(logs, "ERROR")
		infoLogs := FilterLogsByLevel(logs, "INFO")
		
		// Then: Should return entries with matching level
		assert.Len(t, errorLogs, 2)
		assert.Len(t, infoLogs, 2)
		
		for _, log := range errorLogs {
			assert.Equal(t, "ERROR", log.Level)
		}
		
		for _, log := range infoLogs {
			assert.Equal(t, "INFO", log.Level)
		}
	})
	
	t.Run("LogsContain", func(t *testing.T) {
		// When: Checking if logs contain text
		hasError := LogsContain(logs, "error")
		hasWarning := LogsContain(logs, "warning")
		
		// Then: Should return correct boolean
		assert.True(t, hasError)
		assert.False(t, hasWarning)
	})
	
	t.Run("LogsContainLevel", func(t *testing.T) {
		// When: Checking if logs contain level
		hasError := LogsContainLevel(logs, "ERROR")
		hasWarning := LogsContainLevel(logs, "WARNING")
		
		// Then: Should return correct boolean
		assert.True(t, hasError)
		assert.False(t, hasWarning)
	})
}

// Test log statistics
func TestLogStatistics(t *testing.T) {
	// Given: Log entries with various levels and timestamps
	baseTime := time.Now()
	logs := []LogEntry{
		{Timestamp: baseTime, Message: "Error 1", Level: "ERROR"},
		{Timestamp: baseTime.Add(1 * time.Minute), Message: "Info 1", Level: "INFO"},
		{Timestamp: baseTime.Add(2 * time.Minute), Message: "Error 2", Level: "ERROR"},
		{Timestamp: baseTime.Add(3 * time.Minute), Message: "Debug 1", Level: "DEBUG"},
		{Timestamp: baseTime.Add(4 * time.Minute), Message: "Info 2", Level: "INFO"},
	}
	
	// When: Getting log statistics
	stats := GetLogStats(logs)
	
	// Then: Should provide correct statistics
	assert.Equal(t, 5, stats.TotalEntries)
	assert.Equal(t, baseTime, stats.EarliestEntry)
	assert.Equal(t, baseTime.Add(4*time.Minute), stats.LatestEntry)
	assert.Equal(t, 4*time.Minute, stats.TimeSpan)
	assert.Equal(t, 2, stats.LevelCounts["ERROR"])
	assert.Equal(t, 2, stats.LevelCounts["INFO"])
	assert.Equal(t, 1, stats.LevelCounts["DEBUG"])
}

// Test retry configuration and backoff calculation
func TestRetryLogic(t *testing.T) {
	t.Run("DefaultRetryConfig", func(t *testing.T) {
		// When: Getting default retry config
		config := defaultRetryConfig()
		
		// Then: Should have sensible defaults
		assert.Equal(t, DefaultRetryAttempts, config.MaxAttempts)
		assert.Equal(t, DefaultRetryDelay, config.BaseDelay)
		assert.Equal(t, 30*time.Second, config.MaxDelay)
		assert.Equal(t, 2.0, config.Multiplier)
	})
	
	t.Run("CalculateBackoffDelay", func(t *testing.T) {
		config := defaultRetryConfig()
		
		// Test exponential backoff
		delay0 := calculateBackoffDelay(0, config)
		delay1 := calculateBackoffDelay(1, config)
		delay2 := calculateBackoffDelay(2, config)
		delay3 := calculateBackoffDelay(3, config)
		
		// Should increase exponentially
		assert.Equal(t, config.BaseDelay, delay0)
		assert.Equal(t, config.BaseDelay*time.Duration(config.Multiplier), delay1)
		assert.True(t, delay2 > delay1)
		assert.True(t, delay3 > delay2)
		
		// Test max delay cap
		veryHighAttempt := calculateBackoffDelay(100, config)
		assert.Equal(t, config.MaxDelay, veryHighAttempt)
		
		// Test negative attempt handling
		negativeDelay := calculateBackoffDelay(-1, config)
		assert.Equal(t, config.BaseDelay, negativeDelay)
	})
}

// Test string utility functions
func TestStringUtilities(t *testing.T) {
	t.Run("Contains", func(t *testing.T) {
		testCases := []struct {
			s        string
			substr   string
			expected bool
		}{
			{"", "", true},
			{"", "a", false},
			{"a", "", true},
			{"hello", "hello", true},
			{"hello", "hell", true},
			{"hello", "ello", true},
			{"hello", "ell", true},
			{"hello", "Hello", false}, // Case sensitive
			{"hello", "world", false},
			{"a", "aa", false},
		}
		
		for _, tc := range testCases {
			result := contains(tc.s, tc.substr)
			assert.Equal(t, tc.expected, result,
				"contains(%q, %q) = %v, expected %v", tc.s, tc.substr, result, tc.expected)
		}
	})
	
	t.Run("IndexOf", func(t *testing.T) {
		testCases := []struct {
			s        string
			substr   string
			expected int
		}{
			{"", "", 0},
			{"", "a", -1},
			{"a", "", 0},
			{"hello", "h", 0},
			{"hello", "o", 4},
			{"hello", "ll", 2},
			{"hello", "Hello", -1}, // Case sensitive
			{"hello", "world", -1},
			{"aaaa", "aa", 0}, // Returns first occurrence
		}
		
		for _, tc := range testCases {
			result := indexOf(tc.s, tc.substr)
			assert.Equal(t, tc.expected, result,
				"indexOf(%q, %q) = %v, expected %v", tc.s, tc.substr, result, tc.expected)
		}
	})
}

// Test invoke options merging
func TestInvokeOptionsMerging(t *testing.T) {
	t.Run("DefaultOptions", func(t *testing.T) {
		// When: Getting default invoke options
		opts := defaultInvokeOptions()
		
		// Then: Should have sensible defaults
		assert.Equal(t, InvocationTypeRequestResponse, opts.InvocationType)
		assert.Equal(t, LogTypeNone, opts.LogType)
		assert.Equal(t, DefaultTimeout, opts.Timeout)
		assert.Equal(t, DefaultRetryAttempts, opts.MaxRetries)
		assert.Equal(t, DefaultRetryDelay, opts.RetryDelay)
		assert.NotNil(t, opts.PayloadValidator)
	})
	
	t.Run("MergeOptions", func(t *testing.T) {
		// Given: User options with some values set
		userOpts := &InvokeOptions{
			InvocationType: lambdaTypes.InvocationTypeEvent,
			Timeout:        5 * time.Second,
			// LogType and others left unset
		}
		
		// When: Merging with defaults
		merged := mergeInvokeOptions(userOpts)
		
		// Then: Should use user values where provided, defaults otherwise
		assert.Equal(t, lambdaTypes.InvocationTypeEvent, merged.InvocationType)
		assert.Equal(t, 5*time.Second, merged.Timeout)
		assert.Equal(t, LogTypeNone, merged.LogType) // Default
		assert.Equal(t, DefaultRetryAttempts, merged.MaxRetries) // Default
		assert.NotNil(t, merged.PayloadValidator) // Default
	})
	
	t.Run("NilOptions", func(t *testing.T) {
		// When: Merging nil options
		merged := mergeInvokeOptions(nil)
		
		// Then: Should return all defaults
		defaults := defaultInvokeOptions()
		assert.Equal(t, defaults.InvocationType, merged.InvocationType)
		assert.Equal(t, defaults.LogType, merged.LogType)
		assert.Equal(t, defaults.Timeout, merged.Timeout)
		assert.Equal(t, defaults.MaxRetries, merged.MaxRetries)
		assert.Equal(t, defaults.RetryDelay, merged.RetryDelay)
	})
}

// Test payload validation in options
func TestPayloadValidatorOptions(t *testing.T) {
	// When: Using default payload validator
	opts := defaultInvokeOptions()
	validator := opts.PayloadValidator
	
	// Then: Should validate payload size
	smallPayload := `{"key": "value"}`
	assert.NoError(t, validator(smallPayload))
	
	largePayload := strings.Repeat("x", MaxPayloadSize+1)
	err := validator(largePayload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payload size exceeds maximum")
}

// Test error classification
func TestErrorClassification(t *testing.T) {
	testCases := []struct {
		name           string
		errorMessage   string
		isNonRetryable bool
	}{
		{"InvalidParameterValue", "InvalidParameterValueException: Invalid parameter", true},
		{"ResourceNotFound", "ResourceNotFoundException: Function not found", true},
		{"InvalidRequestContent", "InvalidRequestContentException: Invalid request", true},
		{"RequestTooLarge", "RequestTooLargeException: Request too large", true},
		{"UnsupportedMediaType", "UnsupportedMediaTypeException: Unsupported media type", true},
		{"ServiceUnavailable", "Service temporarily unavailable", false},
		{"RateExceeded", "Rate exceeded", false},
		{"InternalError", "Internal server error", false},
		{"NetworkTimeout", "Network timeout", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Checking if error is retryable
			err := &testError{message: tc.errorMessage}
			isNonRetryable := isNonRetryableError(err)
			
			// Then: Should match expectation
			assert.Equal(t, tc.isNonRetryable, isNonRetryable,
				"Error '%s' retryable classification mismatch", tc.errorMessage)
		})
	}
}

// Test helper for creating test errors
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// Benchmark tests for performance validation
func BenchmarkValidateFunctionNameUnit(b *testing.B) {
	functionName := "test-function-with-long-name-for-benchmarking"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateFunctionName(functionName)
	}
}

func BenchmarkValidatePayloadUnit(b *testing.B) {
	payload := `{"key": "value", "number": 42, "array": [1, 2, 3], "nested": {"inner": "data"}}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validatePayload(payload)
	}
}

func BenchmarkMarshalPayloadUnit(b *testing.B) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
		"key4": []string{"a", "b", "c"},
		"key5": map[string]interface{}{
			"nested": "data",
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalPayloadE(payload)
	}
}

func BenchmarkFilterLogs(b *testing.B) {
	// Create test logs
	logs := make([]LogEntry, 1000)
	for i := 0; i < 1000; i++ {
		logs[i] = LogEntry{
			Message: fmt.Sprintf("Log entry %d with error", i),
			Level:   "INFO",
		}
		if i%10 == 0 {
			logs[i].Level = "ERROR"
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FilterLogs(logs, "error")
	}
}

func BenchmarkCalculateBackoffDelay(b *testing.B) {
	config := defaultRetryConfig()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateBackoffDelay(i%10, config)
	}
}