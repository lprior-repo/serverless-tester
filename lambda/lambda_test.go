package lambda

import (
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock testing context for unit tests
type mockTestingT struct {
	errorMessages []string
	failed        bool
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

func (m *mockTestingT) FailNow() {
	m.failed = true
}

func newMockTestingT() *mockTestingT {
	return &mockTestingT{
		errorMessages: make([]string, 0),
		failed:        false,
	}
}

// Test data factories following functional programming principles

func createTestFunctionConfiguration() *FunctionConfiguration {
	return &FunctionConfiguration{
		FunctionName:    "test-function",
		FunctionArn:     "arn:aws:lambda:us-east-1:123456789012:function:test-function",
		Runtime:         types.RuntimeNodejs18x,
		Handler:         "index.handler",
		Description:     "Test Lambda function",
		Timeout:         30,
		MemorySize:      256,
		LastModified:    "2024-01-01T00:00:00.000+0000",
		Environment: map[string]string{
			"NODE_ENV": "test",
			"DEBUG":    "true",
		},
		Role:        "arn:aws:iam::123456789012:role/lambda-role",
		State:       types.StateActive,
		StateReason: "",
		Version:     "$LATEST",
	}
}

func createTestInvokeResult(statusCode int32, payload string, functionError string) *InvokeResult {
	return &InvokeResult{
		StatusCode:      statusCode,
		Payload:         payload,
		LogResult:       "",
		ExecutedVersion: "$LATEST",
		FunctionError:   functionError,
		ExecutionTime:   100 * time.Millisecond,
	}
}

func createTestLogEntries() []LogEntry {
	now := time.Now()
	return []LogEntry{
		{
			Timestamp: now.Add(-2 * time.Minute),
			Message:   "START RequestId: abc123",
			Level:     "INFO",
			RequestID: "abc123",
		},
		{
			Timestamp: now.Add(-90 * time.Second),
			Message:   "Processing event",
			Level:     "INFO",
			RequestID: "abc123",
		},
		{
			Timestamp: now.Add(-80 * time.Second),
			Message:   "Error occurred during processing",
			Level:     "ERROR",
			RequestID: "abc123",
		},
		{
			Timestamp: now.Add(-70 * time.Second),
			Message:   "END RequestId: abc123",
			Level:     "INFO",
			RequestID: "abc123",
		},
	}
}

// Table-driven test structures

type validateFunctionNameTestCase struct {
	name          string
	functionName  string
	expectedError bool
	description   string
}

type invokeOptionsTestCase struct {
	name              string
	options           *InvokeOptions
	expectedDefaults  InvokeOptions
	description       string
}

type eventBuilderTestCase struct {
	name            string
	builderFunc     func() (string, error)
	expectedFields  []string
	shouldError     bool
	description     string
}

type assertionTestCase struct {
	name           string
	setupFunc      func() interface{}
	assertionFunc  func(interface{}, *mockTestingT)
	shouldFail     bool
	description    string
}

// Unit tests for core validation functions

func TestValidateFunctionName(t *testing.T) {
	tests := []validateFunctionNameTestCase{
		{
			name:          "ValidFunctionName",
			functionName:  "my-lambda-function",
			expectedError: false,
			description:   "Valid function name with hyphens should pass validation",
		},
		{
			name:          "ValidFunctionNameWithUnderscores",
			functionName:  "my_lambda_function",
			expectedError: false,
			description:   "Valid function name with underscores should pass validation",
		},
		{
			name:          "ValidFunctionNameWithNumbers",
			functionName:  "myFunction123",
			expectedError: false,
			description:   "Valid function name with numbers should pass validation",
		},
		{
			name:          "EmptyFunctionName",
			functionName:  "",
			expectedError: true,
			description:   "Empty function name should fail validation",
		},
		{
			name:          "TooLongFunctionName",
			functionName:  "this-function-name-is-way-too-long-and-exceeds-the-maximum-allowed-length-for-aws-lambda-functions",
			expectedError: true,
			description:   "Function name exceeding 64 characters should fail validation",
		},
		{
			name:          "InvalidCharacters",
			functionName:  "my-function@invalid",
			expectedError: true,
			description:   "Function name with invalid characters should fail validation",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateFunctionName(tc.functionName)
			
			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestValidatePayload(t *testing.T) {
	tests := []struct {
		name          string
		payload       string
		expectedError bool
		description   string
	}{
		{
			name:          "ValidJSONPayload",
			payload:       `{"key": "value"}`,
			expectedError: false,
			description:   "Valid JSON payload should pass validation",
		},
		{
			name:          "EmptyPayload",
			payload:       "",
			expectedError: false,
			description:   "Empty payload should be allowed",
		},
		{
			name:          "ValidArrayPayload",
			payload:       `[1, 2, 3]`,
			expectedError: false,
			description:   "Valid JSON array payload should pass validation",
		},
		{
			name:          "ValidStringPayload",
			payload:       `"simple string"`,
			expectedError: false,
			description:   "Valid JSON string payload should pass validation",
		},
		{
			name:          "InvalidJSONPayload",
			payload:       `{"key": "value"`,
			expectedError: true,
			description:   "Invalid JSON payload should fail validation",
		},
		{
			name:          "InvalidStructure",
			payload:       `{key: value}`,
			expectedError: true,
			description:   "JSON with unquoted keys should fail validation",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validatePayload(tc.payload)
			
			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestMergeInvokeOptions(t *testing.T) {
	tests := []invokeOptionsTestCase{
		{
			name:    "NilOptions",
			options: nil,
			expectedDefaults: InvokeOptions{
				InvocationType: InvocationTypeRequestResponse,
				LogType:        LogTypeNone,
				Timeout:        DefaultTimeout,
				MaxRetries:     DefaultRetryAttempts,
				RetryDelay:     DefaultRetryDelay,
			},
			description: "Nil options should return all defaults",
		},
		{
			name: "PartialOptions",
			options: &InvokeOptions{
				InvocationType: types.InvocationTypeEvent,
				Timeout:        60 * time.Second,
			},
			expectedDefaults: InvokeOptions{
				InvocationType: types.InvocationTypeEvent,
				LogType:        LogTypeNone,
				Timeout:        60 * time.Second,
				MaxRetries:     DefaultRetryAttempts,
				RetryDelay:     DefaultRetryDelay,
			},
			description: "Partial options should merge with defaults",
		},
		{
			name: "FullOptions",
			options: &InvokeOptions{
				InvocationType: types.InvocationTypeDryRun,
				LogType:        LogTypeTail,
				Timeout:        90 * time.Second,
				MaxRetries:     5,
				RetryDelay:     2 * time.Second,
			},
			expectedDefaults: InvokeOptions{
				InvocationType: types.InvocationTypeDryRun,
				LogType:        LogTypeTail,
				Timeout:        90 * time.Second,
				MaxRetries:     5,
				RetryDelay:     2 * time.Second,
			},
			description: "Full options should override all defaults",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := mergeInvokeOptions(tc.options)
			
			// Assert
			assert.Equal(t, tc.expectedDefaults.InvocationType, result.InvocationType, "InvocationType should match")
			assert.Equal(t, tc.expectedDefaults.LogType, result.LogType, "LogType should match")
			assert.Equal(t, tc.expectedDefaults.Timeout, result.Timeout, "Timeout should match")
			assert.Equal(t, tc.expectedDefaults.MaxRetries, result.MaxRetries, "MaxRetries should match")
			assert.Equal(t, tc.expectedDefaults.RetryDelay, result.RetryDelay, "RetryDelay should match")
			assert.NotNil(t, result.PayloadValidator, "PayloadValidator should not be nil")
		})
	}
}

func TestCalculateBackoffDelay(t *testing.T) {
	config := defaultRetryConfig()
	
	tests := []struct {
		name           string
		attempt        int
		expectedMin    time.Duration
		expectedMax    time.Duration
		description    string
	}{
		{
			name:        "FirstAttempt",
			attempt:     0,
			expectedMin: config.BaseDelay,
			expectedMax: config.BaseDelay,
			description: "First attempt should return base delay",
		},
		{
			name:        "SecondAttempt", 
			attempt:     1,
			expectedMin: config.BaseDelay * 2, // BaseDelay * Multiplier^1 = BaseDelay * 2
			expectedMax: config.BaseDelay * 2,
			description: "Second attempt should be base delay times multiplier",
		},
		{
			name:        "HighAttempt",
			attempt:     10,
			expectedMin: config.BaseDelay,
			expectedMax: config.MaxDelay,
			description: "High attempt should be capped at max delay",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			delay := calculateBackoffDelay(tc.attempt, config)
			
			// Assert
			assert.GreaterOrEqual(t, delay, tc.expectedMin, tc.description)
			assert.LessOrEqual(t, delay, tc.expectedMax, tc.description)
		})
	}
}

// Tests for payload marshaling/unmarshaling

func TestMarshalPayloadE(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		expectedResult string
		shouldError    bool
		description    string
	}{
		{
			name:           "SimpleMap",
			payload:        map[string]interface{}{"key": "value"},
			expectedResult: `{"key":"value"}`,
			shouldError:    false,
			description:    "Simple map should marshal correctly",
		},
		{
			name:           "NilPayload",
			payload:        nil,
			expectedResult: "",
			shouldError:    false,
			description:    "Nil payload should return empty string",
		},
		{
			name:           "StringPayload",
			payload:        "simple string",
			expectedResult: `"simple string"`,
			shouldError:    false,
			description:    "String payload should marshal with quotes",
		},
		{
			name:        "InvalidPayload",
			payload:     make(chan int), // channels can't be marshaled
			shouldError: true,
			description: "Invalid payload should return error",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := MarshalPayloadE(tc.payload)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				assert.Equal(t, tc.expectedResult, result, tc.description)
			}
		})
	}
}

func TestParseInvokeOutputE(t *testing.T) {
	tests := []struct {
		name        string
		result      *InvokeResult
		target      interface{}
		shouldError bool
		description string
	}{
		{
			name: "ValidJSONParsing",
			result: &InvokeResult{
				Payload: `{"message": "success", "count": 42}`,
			},
			target:      &map[string]interface{}{},
			shouldError: false,
			description: "Valid JSON should parse correctly",
		},
		{
			name:        "NilResult",
			result:      nil,
			target:      &map[string]interface{}{},
			shouldError: true,
			description: "Nil result should return error",
		},
		{
			name: "EmptyPayload",
			result: &InvokeResult{
				Payload: "",
			},
			target:      &map[string]interface{}{},
			shouldError: true,
			description: "Empty payload should return error",
		},
		{
			name: "InvalidJSON",
			result: &InvokeResult{
				Payload: `{"invalid": json}`,
			},
			target:      &map[string]interface{}{},
			shouldError: true,
			description: "Invalid JSON should return error",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := ParseInvokeOutputE(tc.result, tc.target)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				// Verify the target was populated for valid cases
				if tc.result != nil && tc.result.Payload != "" {
					targetMap := tc.target.(*map[string]interface{})
					assert.NotEmpty(t, *targetMap, "Target should be populated")
				}
			}
		})
	}
}

// Tests for event builders

func TestBuildS3EventE(t *testing.T) {
	tests := []eventBuilderTestCase{
		{
			name: "ValidS3Event",
			builderFunc: func() (string, error) {
				return BuildS3EventE("test-bucket", "test/key.json", "s3:ObjectCreated:Put")
			},
			expectedFields: []string{"Records", "eventSource", "s3", "bucket", "object"},
			shouldError:    false,
			description:    "Valid S3 event should build correctly",
		},
		{
			name: "EmptyBucketName",
			builderFunc: func() (string, error) {
				return BuildS3EventE("", "test/key.json", "s3:ObjectCreated:Put")
			},
			shouldError: true,
			description: "Empty bucket name should return error",
		},
		{
			name: "EmptyObjectKey",
			builderFunc: func() (string, error) {
				return BuildS3EventE("test-bucket", "", "s3:ObjectCreated:Put")
			},
			shouldError: true,
			description: "Empty object key should return error",
		},
		{
			name: "DefaultEventName",
			builderFunc: func() (string, error) {
				return BuildS3EventE("test-bucket", "test/key.json", "")
			},
			expectedFields: []string{"Records", "eventSource", "s3:ObjectCreated:Put"},
			shouldError:    false,
			description:    "Empty event name should use default",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := tc.builderFunc()
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				
				// Verify expected fields are present
				for _, field := range tc.expectedFields {
					assert.Contains(t, result, field, "Expected field %s should be present", field)
				}
				
				// Verify it's valid JSON
				var parsed interface{}
				err := ParseInvokeOutputE(&InvokeResult{Payload: result}, &parsed)
				assert.NoError(t, err, "Result should be valid JSON")
			}
		})
	}
}

func TestBuildAPIGatewayEventE(t *testing.T) {
	tests := []struct {
		name           string
		httpMethod     string
		path           string
		body           string
		expectedMethod string
		expectedPath   string
		description    string
	}{
		{
			name:           "ValidGETRequest",
			httpMethod:     "GET",
			path:           "/api/users",
			body:           "",
			expectedMethod: "GET",
			expectedPath:   "/api/users",
			description:    "Valid GET request should build correctly",
		},
		{
			name:           "ValidPOSTRequest",
			httpMethod:     "POST",
			path:           "/api/users",
			body:           `{"name": "John Doe"}`,
			expectedMethod: "POST",
			expectedPath:   "/api/users",
			description:    "Valid POST request with body should build correctly",
		},
		{
			name:           "DefaultValues",
			httpMethod:     "",
			path:           "",
			body:           "",
			expectedMethod: "GET",
			expectedPath:   "/",
			description:    "Empty values should use defaults",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := BuildAPIGatewayEventE(tc.httpMethod, tc.path, tc.body)
			
			// Assert
			require.NoError(t, err, tc.description)
			assert.Contains(t, result, tc.expectedMethod, "Should contain expected HTTP method")
			assert.Contains(t, result, tc.expectedPath, "Should contain expected path")
			
			if tc.body != "" {
				// The body gets JSON-encoded when marshalled, so we need to check for the escaped version
				expectedBodyInJSON := `"` + strings.ReplaceAll(strings.ReplaceAll(tc.body, `\`, `\\`), `"`, `\"`) + `"`
				assert.Contains(t, result, expectedBodyInJSON, "Should contain expected body in JSON-encoded form")
			}
		})
	}
}

// Tests for validation functions

func TestValidateFunctionConfiguration(t *testing.T) {
	validConfig := createTestFunctionConfiguration()
	
	tests := []struct {
		name            string
		config          *FunctionConfiguration
		expected        *FunctionConfiguration
		expectedErrors  int
		description     string
	}{
		{
			name:           "ValidConfiguration",
			config:         validConfig,
			expected:       validConfig,
			expectedErrors: 0,
			description:    "Identical configurations should have no validation errors",
		},
		{
			name:   "NilConfiguration",
			config: nil,
			expected: validConfig,
			expectedErrors: 1,
			description:    "Nil configuration should return one error",
		},
		{
			name: "MismatchedRuntime",
			config: func() *FunctionConfiguration {
				config := *validConfig
				config.Runtime = types.RuntimePython39
				return &config
			}(),
			expected:       validConfig,
			expectedErrors: 1,
			description:    "Mismatched runtime should return one error",
		},
		{
			name: "MissingEnvironmentVariable",
			config: func() *FunctionConfiguration {
				config := *validConfig
				config.Environment = map[string]string{"NODE_ENV": "test"}
				return &config
			}(),
			expected:       validConfig,
			expectedErrors: 1,
			description:    "Missing environment variable should return one error",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			errors := ValidateFunctionConfiguration(tc.config, tc.expected)
			
			// Assert
			assert.Len(t, errors, tc.expectedErrors, tc.description)
		})
	}
}

func TestValidateInvokeResult(t *testing.T) {
	successResult := createTestInvokeResult(200, `{"status": "success"}`, "")
	errorResult := createTestInvokeResult(200, `{"status": "error"}`, "Handled")
	
	tests := []struct {
		name                     string
		result                   *InvokeResult
		expectSuccess            bool
		expectedPayloadContains  string
		expectedErrors           int
		description              string
	}{
		{
			name:          "ValidSuccessResult",
			result:        successResult,
			expectSuccess: true,
			expectedPayloadContains: "success",
			expectedErrors: 0,
			description:   "Valid success result should have no errors",
		},
		{
			name:          "ValidErrorResult", 
			result:        errorResult,
			expectSuccess: false,
			expectedPayloadContains: "error",
			expectedErrors: 0,
			description:   "Valid error result should have no errors",
		},
		{
			name:           "NilResult",
			result:         nil,
			expectSuccess:  true,
			expectedErrors: 1,
			description:    "Nil result should return one error",
		},
		{
			name:          "UnexpectedSuccess",
			result:        successResult,
			expectSuccess: false,
			expectedErrors: 1,
			description:   "Unexpected success should return one error",
		},
		{
			name:                    "MissingPayloadContent",
			result:                  successResult,
			expectSuccess:           true,
			expectedPayloadContains: "missing",
			expectedErrors:          1,
			description:             "Missing payload content should return one error",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			errors := ValidateInvokeResult(tc.result, tc.expectSuccess, tc.expectedPayloadContains)
			
			// Assert
			assert.Len(t, errors, tc.expectedErrors, tc.description)
		})
	}
}

// Tests for log utility functions

func TestLogsContain(t *testing.T) {
	logs := createTestLogEntries()
	
	tests := []struct {
		name        string
		logs        []LogEntry
		searchText  string
		expected    bool
		description string
	}{
		{
			name:        "TextExists",
			logs:        logs,
			searchText:  "Processing event",
			expected:    true,
			description: "Existing text should be found",
		},
		{
			name:        "TextDoesNotExist",
			logs:        logs,
			searchText:  "nonexistent text",
			expected:    false,
			description: "Non-existing text should not be found",
		},
		{
			name:        "EmptyLogs",
			logs:        []LogEntry{},
			searchText:  "any text",
			expected:    false,
			description: "Empty logs should return false",
		},
		{
			name:        "EmptySearchText",
			logs:        logs,
			searchText:  "",
			expected:    true,
			description: "Empty search text should always be found",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := LogsContain(tc.logs, tc.searchText)
			
			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestLogsContainLevel(t *testing.T) {
	logs := createTestLogEntries()
	
	tests := []struct {
		name        string
		logs        []LogEntry
		level       string
		expected    bool
		description string
	}{
		{
			name:        "ErrorLevelExists",
			logs:        logs,
			level:       "ERROR",
			expected:    true,
			description: "ERROR level should be found",
		},
		{
			name:        "InfoLevelExists",
			logs:        logs,
			level:       "INFO",
			expected:    true,
			description: "INFO level should be found",
		},
		{
			name:        "WarnLevelDoesNotExist",
			logs:        logs,
			level:       "WARN",
			expected:    false,
			description: "WARN level should not be found",
		},
		{
			name:        "CaseInsensitive",
			logs:        logs,
			level:       "error",
			expected:    true,
			description: "Search should be case insensitive",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := LogsContainLevel(tc.logs, tc.level)
			
			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestGetLogStats(t *testing.T) {
	logs := createTestLogEntries()
	
	// Act
	stats := GetLogStats(logs)
	
	// Assert
	assert.Equal(t, len(logs), stats.TotalEntries, "Total entries should match")
	assert.Equal(t, 3, stats.LevelCounts["INFO"], "Should have 3 INFO entries")
	assert.Equal(t, 1, stats.LevelCounts["ERROR"], "Should have 1 ERROR entry")
	assert.True(t, stats.EarliestEntry.Before(stats.LatestEntry), "Earliest should be before latest")
	assert.True(t, stats.TimeSpan > 0, "Time span should be positive")
}

func TestGetLogsByRequestID(t *testing.T) {
	logs := createTestLogEntries()
	
	// Act
	filteredLogs := GetLogsByRequestID(logs, "abc123")
	
	// Assert
	assert.Equal(t, len(logs), len(filteredLogs), "All logs should have the same request ID")
	
	for _, log := range filteredLogs {
		assert.Equal(t, "abc123", log.RequestID, "All filtered logs should have the expected request ID")
	}
}

// Tests for event parsing and validation

func TestParseS3EventE(t *testing.T) {
	// Arrange
	validEventJSON, err := BuildS3EventE("test-bucket", "test/key.json", "s3:ObjectCreated:Put")
	require.NoError(t, err)
	
	tests := []struct {
		name        string
		eventJSON   string
		shouldError bool
		description string
	}{
		{
			name:        "ValidS3Event",
			eventJSON:   validEventJSON,
			shouldError: false,
			description: "Valid S3 event JSON should parse successfully",
		},
		{
			name:        "InvalidJSON",
			eventJSON:   `{"invalid": json}`,
			shouldError: true,
			description: "Invalid JSON should return error",
		},
		{
			name:        "EmptyJSON",
			eventJSON:   "",
			shouldError: true,
			description: "Empty JSON should return error",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			event, err := ParseS3EventE(tc.eventJSON)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				assert.NotEmpty(t, event.Records, "Parsed event should have records")
			}
		})
	}
}

func TestExtractS3BucketNameE(t *testing.T) {
	// Arrange
	bucketName := "my-test-bucket"
	eventJSON, err := BuildS3EventE(bucketName, "test/key.json", "s3:ObjectCreated:Put")
	require.NoError(t, err)
	
	// Act
	extractedName, err := ExtractS3BucketNameE(eventJSON)
	
	// Assert
	assert.NoError(t, err, "Should extract bucket name without error")
	assert.Equal(t, bucketName, extractedName, "Extracted name should match original")
}

func TestExtractS3ObjectKeyE(t *testing.T) {
	// Arrange
	objectKey := "test/path/to/file.json"
	eventJSON, err := BuildS3EventE("test-bucket", objectKey, "s3:ObjectCreated:Put")
	require.NoError(t, err)
	
	// Act
	extractedKey, err := ExtractS3ObjectKeyE(eventJSON)
	
	// Assert
	assert.NoError(t, err, "Should extract object key without error")
	assert.Equal(t, objectKey, extractedKey, "Extracted key should match original")
}

func TestValidateEventStructure(t *testing.T) {
	// Arrange
	s3EventJSON, _ := BuildS3EventE("test-bucket", "test/key.json", "s3:ObjectCreated:Put")
	apiGatewayEventJSON, _ := BuildAPIGatewayEventE("GET", "/test", "")
	
	tests := []struct {
		name           string
		eventJSON      string
		eventType      string
		expectedErrors int
		description    string
	}{
		{
			name:           "ValidS3Event",
			eventJSON:      s3EventJSON,
			eventType:      "s3",
			expectedErrors: 0,
			description:    "Valid S3 event should have no validation errors",
		},
		{
			name:           "ValidAPIGatewayEvent",
			eventJSON:      apiGatewayEventJSON,
			eventType:      "apigateway",
			expectedErrors: 0,
			description:    "Valid API Gateway event should have no validation errors",
		},
		{
			name:           "InvalidEventType",
			eventJSON:      s3EventJSON,
			eventType:      "unknown",
			expectedErrors: 1,
			description:    "Unknown event type should return one error",
		},
		{
			name:           "WrongStructureForType",
			eventJSON:      s3EventJSON,
			eventType:      "dynamodb",
			expectedErrors: 1,
			description:    "Wrong structure for event type should return one error",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			errors := ValidateEventStructure(tc.eventJSON, tc.eventType)
			
			// Assert
			assert.Len(t, errors, tc.expectedErrors, tc.description)
		})
	}
}

// Benchmark tests for performance-critical functions

func BenchmarkMarshalPayloadE(b *testing.B) {
	payload := map[string]interface{}{
		"message": "Hello, World!",
		"count":   42,
		"data":    []string{"a", "b", "c"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := MarshalPayloadE(payload)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildS3EventE(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildS3EventE("test-bucket", "test/key.json", "s3:ObjectCreated:Put")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateFunctionName(b *testing.B) {
	functionName := "my-test-lambda-function"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateFunctionName(functionName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLogsContain(b *testing.B) {
	logs := createTestLogEntries()
	searchText := "Processing event"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LogsContain(logs, searchText)
	}
}

// Integration test examples (these would require actual AWS resources)

// Note: These integration tests are provided as examples but commented out
// since they require actual AWS Lambda functions to exist.

/*
func TestIntegrationInvokeE(t *testing.T) {
	// This test would require a real AWS Lambda function
	t.Skip("Integration test - requires AWS resources")
	
	// Example usage:
	// ctx := &TestContext{
	//     T:         t,
	//     AwsConfig: awsConfig, // real AWS config
	//     Region:    "us-east-1",
	// }
	// 
	// result, err := InvokeE(ctx, "my-real-function", `{"test": "data"}`)
	// assert.NoError(t, err)
	// assert.Equal(t, int32(200), result.StatusCode)
}

func TestIntegrationGetFunctionE(t *testing.T) {
	t.Skip("Integration test - requires AWS resources")
	
	// Example usage:
	// config, err := GetFunctionE(ctx, "my-real-function")
	// assert.NoError(t, err)
	// assert.Equal(t, "my-real-function", config.FunctionName)
}
*/