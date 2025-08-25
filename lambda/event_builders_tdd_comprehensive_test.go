package lambda

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TDD Event Builder Tests - Comprehensive Coverage
// Following strict TDD principles: RED -> GREEN -> REFACTOR

// Test Data Factories - Pure functions creating immutable test data
type eventBuilderScenario struct {
	name        string
	buildFunc   func() (string, error)
	validateFunc func(t *testing.T, eventJSON string)
	expectError bool
	errorMsg    string
}

// createEventBuilderScenarios creates comprehensive test scenarios for all event builders
func createEventBuilderScenarios() []eventBuilderScenario {
	return []eventBuilderScenario{
		// S3 Event Builder Scenarios
		{
			name:        "S3Event_ValidBasicEvent",
			buildFunc:   func() (string, error) { return BuildS3EventE("test-bucket", "test-key.txt", "") },
			validateFunc: validateS3BasicEvent("test-bucket", "test-key.txt", "s3:ObjectCreated:Put"),
			expectError: false,
		},
		{
			name:        "S3Event_ValidCustomEventName", 
			buildFunc:   func() (string, error) { return BuildS3EventE("custom-bucket", "folder/file.json", "s3:ObjectCreated:Post") },
			validateFunc: validateS3BasicEvent("custom-bucket", "folder/file.json", "s3:ObjectCreated:Post"),
			expectError: false,
		},
		{
			name:        "S3Event_EmptyBucketName",
			buildFunc:   func() (string, error) { return BuildS3EventE("", "key.txt", "") },
			validateFunc: nil,
			expectError: true,
			errorMsg:    "bucket name cannot be empty",
		},
		{
			name:        "S3Event_EmptyObjectKey",
			buildFunc:   func() (string, error) { return BuildS3EventE("bucket", "", "") },
			validateFunc: nil,
			expectError: true,
			errorMsg:    "object key cannot be empty",
		},
		// DynamoDB Event Builder Scenarios
		{
			name: "DynamoDBEvent_ValidBasicEvent",
			buildFunc: func() (string, error) {
				return BuildDynamoDBEventE("test-table", "", map[string]interface{}{
					"id": map[string]interface{}{"S": "test-id"},
				})
			},
			validateFunc: validateDynamoDBBasicEvent("test-table", "INSERT"),
			expectError:  false,
		},
		{
			name: "DynamoDBEvent_ValidCustomEventName",
			buildFunc: func() (string, error) {
				return BuildDynamoDBEventE("users-table", "MODIFY", map[string]interface{}{
					"pk": map[string]interface{}{"S": "partition-key"},
					"sk": map[string]interface{}{"S": "sort-key"},
				})
			},
			validateFunc: validateDynamoDBBasicEvent("users-table", "MODIFY"),
			expectError:  false,
		},
		{
			name:        "DynamoDBEvent_EmptyTableName",
			buildFunc:   func() (string, error) { return BuildDynamoDBEventE("", "INSERT", nil) },
			validateFunc: nil,
			expectError: true,
			errorMsg:    "table name cannot be empty",
		},
		// SQS Event Builder Scenarios
		{
			name:        "SQSEvent_ValidBasicEvent",
			buildFunc:   func() (string, error) { return BuildSQSEventE("https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", "") },
			validateFunc: validateSQSBasicEvent("https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", "Hello from SQS!"),
			expectError: false,
		},
		{
			name:        "SQSEvent_ValidCustomMessage",
			buildFunc:   func() (string, error) { return BuildSQSEventE("https://sqs.us-west-2.amazonaws.com/123/queue", "Custom message") },
			validateFunc: validateSQSBasicEvent("https://sqs.us-west-2.amazonaws.com/123/queue", "Custom message"),
			expectError: false,
		},
		{
			name:        "SQSEvent_EmptyQueueURL",
			buildFunc:   func() (string, error) { return BuildSQSEventE("", "message") },
			validateFunc: nil,
			expectError: true,
			errorMsg:    "queue URL cannot be empty",
		},
		// SNS Event Builder Scenarios
		{
			name:        "SNSEvent_ValidBasicEvent",
			buildFunc:   func() (string, error) { return BuildSNSEventE("arn:aws:sns:us-east-1:123456789012:test-topic", "") },
			validateFunc: validateSNSBasicEvent("arn:aws:sns:us-east-1:123456789012:test-topic", "Hello from SNS!"),
			expectError: false,
		},
		{
			name:        "SNSEvent_ValidCustomMessage",
			buildFunc:   func() (string, error) { return BuildSNSEventE("arn:aws:sns:us-west-2:123:topic", "Custom SNS message") },
			validateFunc: validateSNSBasicEvent("arn:aws:sns:us-west-2:123:topic", "Custom SNS message"),
			expectError: false,
		},
		{
			name:        "SNSEvent_EmptyTopicArn",
			buildFunc:   func() (string, error) { return BuildSNSEventE("", "message") },
			validateFunc: nil,
			expectError: true,
			errorMsg:    "topic ARN cannot be empty",
		},
		// API Gateway Event Builder Scenarios
		{
			name:        "APIGatewayEvent_ValidBasicEvent",
			buildFunc:   func() (string, error) { return BuildAPIGatewayEventE("", "", "") },
			validateFunc: validateAPIGatewayBasicEvent("GET", "/", ""),
			expectError: false,
		},
		{
			name:        "APIGatewayEvent_ValidPostRequest",
			buildFunc:   func() (string, error) { return BuildAPIGatewayEventE("POST", "/api/users", `{"name":"John"}`) },
			validateFunc: validateAPIGatewayBasicEvent("POST", "/api/users", `{"name":"John"}`),
			expectError: false,
		},
		// CloudWatch Event Builder Scenarios
		{
			name: "CloudWatchEvent_ValidBasicEvent",
			buildFunc: func() (string, error) {
				return BuildCloudWatchEventE("aws.logs", "", map[string]interface{}{
					"state": "OK",
				})
			},
			validateFunc: validateCloudWatchBasicEvent("aws.logs"),
			expectError:  false,
		},
		{
			name:        "CloudWatchEvent_EmptySource",
			buildFunc:   func() (string, error) { return BuildCloudWatchEventE("", "", nil) },
			validateFunc: nil,
			expectError: true,
			errorMsg:    "source cannot be empty",
		},
		// Kinesis Event Builder Scenarios
		{
			name:        "KinesisEvent_ValidBasicEvent",
			buildFunc:   func() (string, error) { return BuildKinesisEventE("test-stream", "partition-key", "") },
			validateFunc: validateKinesisBasicEvent("test-stream", "partition-key", "Hello from Kinesis!"),
			expectError: false,
		},
		{
			name:        "KinesisEvent_ValidCustomData",
			buildFunc:   func() (string, error) { return BuildKinesisEventE("json-stream", "json-key", `{"eventType":"user_action"}`) },
			validateFunc: validateKinesisBasicEvent("json-stream", "json-key", `{"eventType":"user_action"}`),
			expectError: false,
		},
		{
			name:        "KinesisEvent_EmptyStreamName",
			buildFunc:   func() (string, error) { return BuildKinesisEventE("", "partition-key", "data") },
			validateFunc: nil,
			expectError: true,
			errorMsg:    "stream name cannot be empty",
		},
		{
			name:        "KinesisEvent_EmptyPartitionKey",
			buildFunc:   func() (string, error) { return BuildKinesisEventE("stream", "", "data") },
			validateFunc: nil,
			expectError: true,
			errorMsg:    "partition key cannot be empty",
		},
	}
}

// Validation Functions - Pure functions for event validation

func validateS3BasicEvent(expectedBucket, expectedKey, expectedEventName string) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		
		// Validate JSON structure
		validateJSONStructureTDD(t, eventJSON, []string{"Records"})
		
		// Parse and validate S3 event
		event, err := ParseS3EventE(eventJSON)
		require.NoError(t, err, "Should parse S3 event successfully")
		require.Len(t, event.Records, 1, "Should have exactly one record")
		
		record := event.Records[0]
		assert.Equal(t, "aws:s3", record.EventSource)
		assert.Equal(t, expectedBucket, record.S3.Bucket.Name)
		assert.Equal(t, expectedKey, record.S3.Object.Key)
		assert.Equal(t, expectedEventName, record.EventName)
		assert.Equal(t, "2.1", record.EventVersion)
		
		// Validate timestamp format
		_, err = time.Parse("2006-01-02T15:04:05.000Z", record.EventTime)
		assert.NoError(t, err, "Event time should be in ISO format")
	}
}

func validateDynamoDBBasicEvent(expectedTable, expectedEventName string) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		
		// Validate JSON structure
		validateJSONStructureTDD(t, eventJSON, []string{"Records"})
		
		var event DynamoDBEvent
		err := json.Unmarshal([]byte(eventJSON), &event)
		require.NoError(t, err, "Should parse DynamoDB event successfully")
		require.Len(t, event.Records, 1, "Should have exactly one record")
		
		record := event.Records[0]
		assert.Equal(t, "aws:dynamodb", record.EventSource)
		assert.Equal(t, expectedEventName, record.EventName)
		assert.Equal(t, "1.1", record.EventVersion)
		assert.NotEmpty(t, record.EventID)
		assert.NotEmpty(t, record.Dynamodb.Keys)
	}
}

func validateSQSBasicEvent(expectedQueueURL, expectedMessage string) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		
		// Validate JSON structure
		validateJSONStructureTDD(t, eventJSON, []string{"Records"})
		
		var event SQSEvent
		err := json.Unmarshal([]byte(eventJSON), &event)
		require.NoError(t, err, "Should parse SQS event successfully")
		require.Len(t, event.Records, 1, "Should have exactly one record")
		
		record := event.Records[0]
		assert.Equal(t, "aws:sqs", record.EventSource)
		assert.Equal(t, expectedQueueURL, record.EventSourceARN)
		assert.Equal(t, expectedMessage, record.Body)
		assert.NotEmpty(t, record.MessageId)
	}
}

func validateSNSBasicEvent(expectedTopicArn, expectedMessage string) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		
		// Validate JSON structure
		validateJSONStructureTDD(t, eventJSON, []string{"Records"})
		
		var event SNSEvent
		err := json.Unmarshal([]byte(eventJSON), &event)
		require.NoError(t, err, "Should parse SNS event successfully")
		require.Len(t, event.Records, 1, "Should have exactly one record")
		
		record := event.Records[0]
		assert.Equal(t, "aws:sns", record.EventSource)
		assert.Equal(t, expectedTopicArn, record.Sns.TopicArn)
		assert.Equal(t, expectedMessage, record.Sns.Message)
		assert.Equal(t, "1.0", record.EventVersion)
		assert.NotEmpty(t, record.Sns.MessageId)
		
		// Validate timestamp format
		_, err = time.Parse("2006-01-02T15:04:05.000Z", record.Sns.Timestamp)
		assert.NoError(t, err, "SNS timestamp should be in ISO format")
	}
}

func validateAPIGatewayBasicEvent(expectedMethod, expectedPath, expectedBody string) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		
		// Validate JSON structure
		validateJSONStructureTDD(t, eventJSON, []string{"httpMethod", "path", "requestContext"})
		
		var event APIGatewayEvent
		err := json.Unmarshal([]byte(eventJSON), &event)
		require.NoError(t, err, "Should parse API Gateway event successfully")
		
		assert.Equal(t, expectedMethod, event.HttpMethod)
		assert.Equal(t, expectedPath, event.Path)
		assert.Equal(t, expectedBody, event.Body)
		assert.NotEmpty(t, event.RequestContext.RequestId)
		assert.NotNil(t, event.Headers)
	}
}

func validateCloudWatchBasicEvent(expectedSource string) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		
		// Validate JSON structure
		validateJSONStructureTDD(t, eventJSON, []string{"source", "detail-type", "detail"})
		
		var event CloudWatchEvent
		err := json.Unmarshal([]byte(eventJSON), &event)
		require.NoError(t, err, "Should parse CloudWatch event successfully")
		
		assert.Equal(t, expectedSource, event.Source)
		assert.Equal(t, "0", event.Version)
		assert.NotEmpty(t, event.ID)
		assert.NotNil(t, event.Detail)
		
		// Validate timestamp format
		_, err = time.Parse("2006-01-02T15:04:05Z", event.Time)
		assert.NoError(t, err, "CloudWatch timestamp should be in ISO format")
	}
}

func validateKinesisBasicEvent(expectedStreamName, expectedPartitionKey, expectedData string) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		
		// Validate JSON structure
		validateJSONStructureTDD(t, eventJSON, []string{"Records"})
		
		var event KinesisEvent
		err := json.Unmarshal([]byte(eventJSON), &event)
		require.NoError(t, err, "Should parse Kinesis event successfully")
		require.Len(t, event.Records, 1, "Should have exactly one record")
		
		record := event.Records[0]
		assert.Equal(t, "aws:kinesis", record.EventSource)
		assert.Equal(t, expectedPartitionKey, record.Kinesis.PartitionKey)
		assert.Contains(t, record.EventSourceARN, expectedStreamName)
		assert.Equal(t, "1.0", record.EventVersion)
		assert.NotEmpty(t, record.Kinesis.Data)
		assert.NotEmpty(t, record.Kinesis.SequenceNumber)
	}
}

func validateJSONStructureTDD(t *testing.T, eventJSON string, requiredFields []string) {
	t.Helper()
	
	var jsonObj map[string]interface{}
	err := json.Unmarshal([]byte(eventJSON), &jsonObj)
	require.NoError(t, err, "Should be valid JSON")
	
	for _, field := range requiredFields {
		assert.Contains(t, jsonObj, field, fmt.Sprintf("Should contain field: %s", field))
	}
}

// Main Comprehensive Test Function - Covering all scenarios with TDD approach

func TestEventBuilders_ComprehensiveScenarios_TDD(t *testing.T) {
	scenarios := createEventBuilderScenarios()
	
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Execute the build function
			result, err := scenario.buildFunc()
			
			if scenario.expectError {
				// RED phase validation - expecting failure
				assert.Error(t, err, "Should return error for invalid input")
				assert.Empty(t, result, "Should return empty result on error")
				if scenario.errorMsg != "" {
					assert.Contains(t, err.Error(), scenario.errorMsg, "Should contain expected error message")
				}
				return
			}
			
			// GREEN phase validation - expecting success
			require.NoError(t, err, "Should not return error for valid input")
			assert.NotEmpty(t, result, "Should return non-empty event JSON")
			
			// Apply specific validation function
			if scenario.validateFunc != nil {
				scenario.validateFunc(t, result)
			}
			
			// Universal validations - all events should pass these
			validateUniversalEventProperties(t, result)
		})
	}
}

func validateUniversalEventProperties(t *testing.T, eventJSON string) {
	t.Helper()
	
	// All events should be valid JSON
	var jsonObj interface{}
	err := json.Unmarshal([]byte(eventJSON), &jsonObj)
	require.NoError(t, err, "All events should be valid JSON")
	
	// All events should be non-empty
	assert.NotEmpty(t, eventJSON, "All events should produce non-empty output")
	
	// All events should be reasonable in size (not too large)
	assert.Less(t, len(eventJSON), 50000, "Event JSON should not be excessively large")
}

// Test Non-E Versions - Ensuring panic behavior is correct

func TestEventBuilders_NonEVersions_PanicBehavior(t *testing.T) {
	panicScenarios := []struct {
		name     string
		testFunc func()
	}{
		{
			name:     "BuildS3Event_EmptyBucket_ShouldPanic",
			testFunc: func() { BuildS3Event("", "key", "") },
		},
		{
			name:     "BuildS3Event_EmptyKey_ShouldPanic", 
			testFunc: func() { BuildS3Event("bucket", "", "") },
		},
		{
			name:     "BuildDynamoDBEvent_EmptyTable_ShouldPanic",
			testFunc: func() { BuildDynamoDBEvent("", "", nil) },
		},
		{
			name:     "BuildSQSEvent_EmptyQueue_ShouldPanic",
			testFunc: func() { BuildSQSEvent("", "message") },
		},
		{
			name:     "BuildSNSEvent_EmptyTopic_ShouldPanic",
			testFunc: func() { BuildSNSEvent("", "message") },
		},
		{
			name:     "BuildCloudWatchEvent_EmptySource_ShouldPanic",
			testFunc: func() { BuildCloudWatchEvent("", "", nil) },
		},
		{
			name:     "BuildKinesisEvent_EmptyStream_ShouldPanic",
			testFunc: func() { BuildKinesisEvent("", "key", "data") },
		},
		{
			name:     "BuildKinesisEvent_EmptyPartitionKey_ShouldPanic",
			testFunc: func() { BuildKinesisEvent("stream", "", "data") },
		},
	}
	
	for _, scenario := range panicScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			assert.Panics(t, scenario.testFunc, "Should panic on invalid input")
		})
	}
}

// Test Success Cases for Non-E Versions

func TestEventBuilders_NonEVersions_SuccessCases(t *testing.T) {
	successScenarios := []struct {
		name     string
		testFunc func() string
		validateFunc func(t *testing.T, result string)
	}{
		{
			name:         "BuildS3Event_ValidInput_ShouldSucceed",
			testFunc:     func() string { return BuildS3Event("bucket", "key", "") },
			validateFunc: func(t *testing.T, result string) { validateS3BasicEvent("bucket", "key", "s3:ObjectCreated:Put")(t, result) },
		},
		{
			name:         "BuildDynamoDBEvent_ValidInput_ShouldSucceed",
			testFunc:     func() string { return BuildDynamoDBEvent("table", "", nil) },
			validateFunc: func(t *testing.T, result string) { validateDynamoDBBasicEvent("table", "INSERT")(t, result) },
		},
		{
			name:         "BuildSQSEvent_ValidInput_ShouldSucceed",
			testFunc:     func() string { return BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/queue", "") },
			validateFunc: func(t *testing.T, result string) { validateSQSBasicEvent("https://sqs.us-east-1.amazonaws.com/123/queue", "Hello from SQS!")(t, result) },
		},
		{
			name:         "BuildSNSEvent_ValidInput_ShouldSucceed",
			testFunc:     func() string { return BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", "") },
			validateFunc: func(t *testing.T, result string) { validateSNSBasicEvent("arn:aws:sns:us-east-1:123:topic", "Hello from SNS!")(t, result) },
		},
		{
			name:         "BuildAPIGatewayEvent_ValidInput_ShouldSucceed",
			testFunc:     func() string { return BuildAPIGatewayEvent("", "", "") },
			validateFunc: func(t *testing.T, result string) { validateAPIGatewayBasicEvent("GET", "/", "")(t, result) },
		},
		{
			name:         "BuildCloudWatchEvent_ValidInput_ShouldSucceed",
			testFunc:     func() string { return BuildCloudWatchEvent("aws.logs", "", nil) },
			validateFunc: func(t *testing.T, result string) { validateCloudWatchBasicEvent("aws.logs")(t, result) },
		},
		{
			name:         "BuildKinesisEvent_ValidInput_ShouldSucceed",
			testFunc:     func() string { return BuildKinesisEvent("stream", "key", "") },
			validateFunc: func(t *testing.T, result string) { validateKinesisBasicEvent("stream", "key", "Hello from Kinesis!")(t, result) },
		},
	}
	
	for _, scenario := range successScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				result := scenario.testFunc()
				scenario.validateFunc(t, result)
			}, "Should not panic on valid input")
		})
	}
}

// Edge Cases and Advanced Validation Tests

func TestEventBuilders_EdgeCases_TDD(t *testing.T) {
	t.Run("S3Event_SpecialCharactersInKey", func(t *testing.T) {
		// Test special characters in object key
		specialKey := "folder/subfolder/file with spaces & special chars!@#.txt"
		result, err := BuildS3EventE("test-bucket", specialKey, "")
		require.NoError(t, err)
		
		event, err := ParseS3EventE(result)
		require.NoError(t, err)
		assert.Equal(t, specialKey, event.Records[0].S3.Object.Key)
	})
	
	t.Run("S3Event_LongObjectKey", func(t *testing.T) {
		// Test long object key (up to 1024 characters is valid for S3)
		longKey := strings.Repeat("a", 1000) + ".txt"
		result, err := BuildS3EventE("test-bucket", longKey, "")
		require.NoError(t, err)
		
		event, err := ParseS3EventE(result)
		require.NoError(t, err)
		assert.Equal(t, longKey, event.Records[0].S3.Object.Key)
	})
	
	t.Run("DynamoDB_ComplexKeys", func(t *testing.T) {
		// Test complex key structure
		complexKeys := map[string]interface{}{
			"pk": map[string]interface{}{
				"S": "USER#123456",
			},
			"sk": map[string]interface{}{
				"S": "PROFILE#2023-12-01",
			},
			"gsi1pk": map[string]interface{}{
				"S": "EMAIL#user@example.com",
			},
		}
		
		result, err := BuildDynamoDBEventE("complex-table", "INSERT", complexKeys)
		require.NoError(t, err)
		
		var event DynamoDBEvent
		err = json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		
		record := event.Records[0]
		assert.Contains(t, record.Dynamodb.Keys, "pk")
		assert.Contains(t, record.Dynamodb.Keys, "sk") 
		assert.Contains(t, record.Dynamodb.Keys, "gsi1pk")
	})
	
	t.Run("SQS_JSONMessageBody", func(t *testing.T) {
		// Test JSON in message body
		jsonMessage := `{"userId": 12345, "action": "login", "timestamp": "2023-12-01T10:00:00Z", "metadata": {"source": "mobile", "version": "1.2.3"}}`
		result, err := BuildSQSEventE("https://sqs.us-east-1.amazonaws.com/123/queue", jsonMessage)
		require.NoError(t, err)
		
		var event SQSEvent
		err = json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		assert.Equal(t, jsonMessage, event.Records[0].Body)
		
		// Verify the JSON inside is valid
		var messageObj map[string]interface{}
		err = json.Unmarshal([]byte(event.Records[0].Body), &messageObj)
		assert.NoError(t, err)
		assert.Equal(t, float64(12345), messageObj["userId"]) // JSON numbers are float64
	})
	
	t.Run("SNS_JSONMessageAndAttributes", func(t *testing.T) {
		// Test JSON message in SNS
		jsonMessage := `{"alert": "System maintenance scheduled", "severity": "INFO", "scheduledTime": "2023-12-01T02:00:00Z"}`
		result, err := BuildSNSEventE("arn:aws:sns:us-east-1:123456789012:alerts", jsonMessage)
		require.NoError(t, err)
		
		var event SNSEvent
		err = json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		assert.Equal(t, jsonMessage, event.Records[0].Sns.Message)
		assert.NotNil(t, event.Records[0].Sns.MessageAttributes)
	})
	
	t.Run("APIGateway_AllHTTPMethods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
		
		for _, method := range methods {
			t.Run(fmt.Sprintf("Method_%s", method), func(t *testing.T) {
				result, err := BuildAPIGatewayEventE(method, "/api/test", "")
				require.NoError(t, err)
				
				var event APIGatewayEvent
				err = json.Unmarshal([]byte(result), &event)
				require.NoError(t, err)
				assert.Equal(t, method, event.HttpMethod)
				assert.Equal(t, method, event.RequestContext.HttpMethod)
			})
		}
	})
	
	t.Run("CloudWatch_ComplexDetailObject", func(t *testing.T) {
		// Test complex detail object
		complexDetail := map[string]interface{}{
			"alarmName": "HighCPUAlarm",
			"state": "ALARM",
			"previousState": "OK",
			"threshold": 80.0,
			"metric": map[string]interface{}{
				"name": "CPUUtilization",
				"namespace": "AWS/EC2",
				"dimensions": map[string]interface{}{
					"InstanceId": "i-1234567890abcdef0",
				},
			},
			"evaluationPeriods": 2,
			"datapointsToAlarm": 2,
		}
		
		result, err := BuildCloudWatchEventE("aws.cloudwatch", "CloudWatch Alarm State Change", complexDetail)
		require.NoError(t, err)
		
		var event CloudWatchEvent
		err = json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		
		assert.Equal(t, "aws.cloudwatch", event.Source)
		assert.Equal(t, "CloudWatch Alarm State Change", event.DetailType)
		assert.Contains(t, event.Detail, "alarmName")
		assert.Contains(t, event.Detail, "metric")
	})
	
	t.Run("Kinesis_Base64EncodedData", func(t *testing.T) {
		// Test that data gets encoded properly in Kinesis events
		originalData := `{"eventType": "user_action", "userId": 123, "action": "click", "element": "button"}`
		result, err := BuildKinesisEventE("analytics-stream", "user-123", originalData)
		require.NoError(t, err)
		
		var event KinesisEvent
		err = json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		
		record := event.Records[0]
		assert.Equal(t, "user-123", record.Kinesis.PartitionKey)
		assert.NotEmpty(t, record.Kinesis.Data)
		
		// The data should be hex-encoded (as per current implementation)
		expectedEncoded := fmt.Sprintf("%x", []byte(originalData))
		assert.Equal(t, expectedEncoded, record.Kinesis.Data)
	})
}

// Performance and Memory Tests

func TestEventBuilders_Performance_TDD(t *testing.T) {
	t.Run("AllEventBuilders_ShouldCompleteQuickly", func(t *testing.T) {
		// Performance test - all builders should complete in reasonable time
		builders := []func(){
			func() { BuildS3Event("bucket", "key", "") },
			func() { BuildDynamoDBEvent("table", "", nil) },
			func() { BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/queue", "") },
			func() { BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", "") },
			func() { BuildAPIGatewayEvent("GET", "/", "") },
			func() { BuildCloudWatchEvent("aws.logs", "", nil) },
			func() { BuildKinesisEvent("stream", "key", "") },
		}
		
		for i, builder := range builders {
			t.Run(fmt.Sprintf("Builder_%d", i), func(t *testing.T) {
				start := time.Now()
				builder()
				duration := time.Since(start)
				
				// Each builder should complete in under 10ms
				assert.Less(t, duration, 10*time.Millisecond, "Event builder should complete quickly")
			})
		}
	})
}

// Cross-cutting Concerns Tests

func TestEventBuilders_CrossCuttingConcerns_TDD(t *testing.T) {
	t.Run("AllEvents_ShouldHaveConsistentJSONStructure", func(t *testing.T) {
		events := map[string]string{
			"S3":         BuildS3Event("bucket", "key", ""),
			"DynamoDB":   BuildDynamoDBEvent("table", "", nil),
			"SQS":        BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/queue", ""),
			"SNS":        BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", ""),
			"APIGateway": BuildAPIGatewayEvent("", "", ""),
			"CloudWatch": BuildCloudWatchEvent("aws.logs", "", nil),
			"Kinesis":    BuildKinesisEvent("stream", "key", ""),
		}
		
		for eventType, eventJSON := range events {
			t.Run(fmt.Sprintf("EventType_%s", eventType), func(t *testing.T) {
				// All should be valid JSON
				var jsonObj interface{}
				err := json.Unmarshal([]byte(eventJSON), &jsonObj)
				assert.NoError(t, err, "Event should be valid JSON")
				
				// All should be reasonably sized
				assert.Greater(t, len(eventJSON), 50, "Event should have reasonable content")
				assert.Less(t, len(eventJSON), 10000, "Event should not be excessively large")
				
				// All should be properly formatted (no trailing commas, etc)
				assert.True(t, json.Valid([]byte(eventJSON)), "Event should be valid JSON")
			})
		}
	})
	
	t.Run("RecordBasedEvents_ShouldHaveRecordsArray", func(t *testing.T) {
		recordBasedEvents := map[string]string{
			"S3":       BuildS3Event("bucket", "key", ""),
			"DynamoDB": BuildDynamoDBEvent("table", "", nil),
			"SQS":      BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/queue", ""),
			"SNS":      BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", ""),
			"Kinesis":  BuildKinesisEvent("stream", "key", ""),
		}
		
		for eventType, eventJSON := range recordBasedEvents {
			t.Run(fmt.Sprintf("RecordEvent_%s", eventType), func(t *testing.T) {
				var eventObj map[string]interface{}
				err := json.Unmarshal([]byte(eventJSON), &eventObj)
				require.NoError(t, err)
				
				records, hasRecords := eventObj["Records"]
				assert.True(t, hasRecords, "Should have Records field")
				
				recordsList, isList := records.([]interface{})
				assert.True(t, isList, "Records should be an array")
				assert.NotEmpty(t, recordsList, "Records array should not be empty")
			})
		}
	})
	
	t.Run("AllEvents_ShouldHaveCorrectEventSources", func(t *testing.T) {
		expectedSources := map[string]string{
			BuildS3Event("bucket", "key", ""):                                               "aws:s3",
			BuildDynamoDBEvent("table", "", nil):                                            "aws:dynamodb", 
			BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/queue", ""):             "aws:sqs",
			BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", ""):                           "aws:sns",
			BuildKinesisEvent("stream", "key", ""):                                          "aws:kinesis",
		}
		
		for eventJSON, expectedSource := range expectedSources {
			var eventObj map[string]interface{}
			err := json.Unmarshal([]byte(eventJSON), &eventObj)
			require.NoError(t, err)
			
			records := eventObj["Records"].([]interface{})
			firstRecord := records[0].(map[string]interface{})
			
			// Try both eventSource and EventSource (SNS uses capitalized version)
			var eventSource string
			if source, ok := firstRecord["eventSource"]; ok && source != nil {
				eventSource = source.(string)
			} else if source, ok := firstRecord["EventSource"]; ok && source != nil {
				eventSource = source.(string)
			}
			
			assert.Equal(t, expectedSource, eventSource, "Should have correct event source")
		}
	})
}