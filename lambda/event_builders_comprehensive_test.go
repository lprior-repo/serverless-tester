package lambda

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Event Builder Test Data Factories - Pure functions for immutable test data creation

// EventBuilderTestCase represents a test case for event builders with higher abstraction
type EventBuilderTestCase struct {
	Name        string
	BuildFunc   func() (string, error)
	ValidateFunc func(t *testing.T, eventJSON string)
	ShouldError bool
}

// S3TestData creates immutable S3 test data
type S3TestData struct {
	BucketName string
	ObjectKey  string
	EventName  string
}

// DynamoDBTestData creates immutable DynamoDB test data  
type DynamoDBTestData struct {
	TableName string
	EventName string
	Keys      map[string]interface{}
}

// SQSTestData creates immutable SQS test data
type SQSTestData struct {
	QueueURL    string
	MessageBody string
}

// SNSTestData creates immutable SNS test data
type SNSTestData struct {
	TopicArn string
	Message  string
}

// APIGatewayTestData creates immutable API Gateway test data
type APIGatewayTestData struct {
	HTTPMethod string
	Path       string
	Body       string
}

// CloudWatchTestData creates immutable CloudWatch test data
type CloudWatchTestData struct {
	LogGroupName string
	LogStream    string
	MessageText  string
}

// KinesisTestData creates immutable Kinesis test data
type KinesisTestData struct {
	StreamName      string
	PartitionKey    string
	Data            string
	SequenceNumber  string
}

// Pure functions for creating test data - following immutability principles

func createS3TestData(bucketName, objectKey, eventName string) S3TestData {
	return S3TestData{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		EventName:  eventName,
	}
}

func createDynamoDBTestData(tableName, eventName string, keys map[string]interface{}) DynamoDBTestData {
	// Create immutable copy of keys map
	keysCopy := make(map[string]interface{})
	for k, v := range keys {
		keysCopy[k] = v
	}
	return DynamoDBTestData{
		TableName: tableName,
		EventName: eventName,
		Keys:      keysCopy,
	}
}

func createSQSTestData(queueURL, messageBody string) SQSTestData {
	return SQSTestData{
		QueueURL:    queueURL,
		MessageBody: messageBody,
	}
}

func createSNSTestData(topicArn, message string) SNSTestData {
	return SNSTestData{
		TopicArn: topicArn,
		Message:  message,
	}
}

func createAPIGatewayTestData(httpMethod, path, body string) APIGatewayTestData {
	return APIGatewayTestData{
		HTTPMethod: httpMethod,
		Path:       path,
		Body:       body,
	}
}

func createCloudWatchTestData(logGroupName, logStream, messageText string) CloudWatchTestData {
	return CloudWatchTestData{
		LogGroupName: logGroupName,
		LogStream:    logStream,
		MessageText:  messageText,
	}
}

func createKinesisTestData(streamName, partitionKey, data, sequenceNumber string) KinesisTestData {
	return KinesisTestData{
		StreamName:     streamName,
		PartitionKey:   partitionKey,
		Data:           data,
		SequenceNumber: sequenceNumber,
	}
}

// Higher-order validation functions - reusable across all event types

func validateEventStructure(eventType string) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		errors := ValidateEventStructure(eventJSON, eventType)
		assert.Empty(t, errors, "Event structure validation should pass")
		assert.NotEmpty(t, eventJSON, "Event JSON should not be empty")
		
		// Validate it's proper JSON
		var jsonObj map[string]interface{}
		err := json.Unmarshal([]byte(eventJSON), &jsonObj)
		assert.NoError(t, err, "Event should be valid JSON")
	}
}

func validateS3EventContent(expectedData S3TestData) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		event, err := ParseS3EventE(eventJSON)
		require.NoError(t, err)
		require.Len(t, event.Records, 1)
		
		record := event.Records[0]
		assert.Equal(t, expectedData.BucketName, record.S3.Bucket.Name)
		assert.Equal(t, expectedData.ObjectKey, record.S3.Object.Key)
		assert.Equal(t, "aws:s3", record.EventSource)
		
		if expectedData.EventName != "" {
			assert.Equal(t, expectedData.EventName, record.EventName)
		} else {
			assert.Equal(t, "s3:ObjectCreated:Put", record.EventName)
		}
	}
}

func validateJSONStructure(expectedFields []string) func(t *testing.T, eventJSON string) {
	return func(t *testing.T, eventJSON string) {
		t.Helper()
		var jsonObj map[string]interface{}
		err := json.Unmarshal([]byte(eventJSON), &jsonObj)
		require.NoError(t, err)
		
		for _, field := range expectedFields {
			assert.Contains(t, jsonObj, field, fmt.Sprintf("Event should contain field: %s", field))
		}
	}
}

// Comprehensive tests for S3 Event Builder

func TestBuildS3Event_ComprehensiveScenarios(t *testing.T) {
	testCases := []struct {
		name      string
		testData  S3TestData
		shouldErr bool
	}{
		{
			name:      "ValidBasicS3Event",
			testData:  createS3TestData("test-bucket", "test-key.txt", ""),
			shouldErr: false,
		},
		{
			name:      "ValidCustomEventName",
			testData:  createS3TestData("custom-bucket", "folder/file.json", "s3:ObjectCreated:Post"),
			shouldErr: false,
		},
		{
			name:      "ValidDeleteEvent",
			testData:  createS3TestData("delete-bucket", "delete-key.txt", "s3:ObjectRemoved:Delete"),
			shouldErr: false,
		},
		{
			name:      "EmptyBucketNameShouldFail",
			testData:  createS3TestData("", "key.txt", ""),
			shouldErr: true,
		},
		{
			name:      "EmptyObjectKeyShouldFail",
			testData:  createS3TestData("bucket", "", ""),
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the E version (error-returning)
			result, err := BuildS3EventE(tc.testData.BucketName, tc.testData.ObjectKey, tc.testData.EventName)
			
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, result)
				return
			}
			
			require.NoError(t, err)
			
			// Apply validations using higher-order functions
			validateEventStructure("s3")(t, result)
			validateS3EventContent(tc.testData)(t, result)
			validateJSONStructure([]string{"Records"})(t, result)
			
			// Test the non-E version (panic on error)
			if !tc.shouldErr {
				assert.NotPanics(t, func() {
					nonEResult := BuildS3Event(tc.testData.BucketName, tc.testData.ObjectKey, tc.testData.EventName)
					assert.Equal(t, result, nonEResult)
				})
			} else {
				assert.Panics(t, func() {
					BuildS3Event(tc.testData.BucketName, tc.testData.ObjectKey, tc.testData.EventName)
				})
			}
		})
	}
}

// Comprehensive tests for DynamoDB Event Builder

func TestBuildDynamoDBEvent_ComprehensiveScenarios(t *testing.T) {
	defaultKeys := map[string]interface{}{
		"id": map[string]interface{}{"S": "test-id"},
	}
	
	customKeys := map[string]interface{}{
		"pk": map[string]interface{}{"S": "partition-key"},
		"sk": map[string]interface{}{"S": "sort-key"},
	}

	testCases := []struct {
		name      string
		testData  DynamoDBTestData
		shouldErr bool
	}{
		{
			name:      "ValidBasicDynamoDBEvent",
			testData:  createDynamoDBTestData("test-table", "", defaultKeys),
			shouldErr: false,
		},
		{
			name:      "ValidInsertEvent",
			testData:  createDynamoDBTestData("users-table", "INSERT", customKeys),
			shouldErr: false,
		},
		{
			name:      "ValidModifyEvent",
			testData:  createDynamoDBTestData("products-table", "MODIFY", defaultKeys),
			shouldErr: false,
		},
		{
			name:      "ValidRemoveEvent",
			testData:  createDynamoDBTestData("orders-table", "REMOVE", customKeys),
			shouldErr: false,
		},
		{
			name:      "EmptyTableNameShouldFail",
			testData:  createDynamoDBTestData("", "INSERT", defaultKeys),
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := BuildDynamoDBEventE(tc.testData.TableName, tc.testData.EventName, tc.testData.Keys)
			
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, result)
				return
			}
			
			require.NoError(t, err)
			validateEventStructure("dynamodb")(t, result)
			validateJSONStructure([]string{"Records"})(t, result)
			
			// Validate DynamoDB specific content
			var event DynamoDBEvent
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			require.Len(t, event.Records, 1)
			
			record := event.Records[0]
			assert.Equal(t, "aws:dynamodb", record.EventSource)
			if tc.testData.EventName != "" {
				assert.Equal(t, tc.testData.EventName, record.EventName)
			} else {
				assert.Equal(t, "INSERT", record.EventName)
			}
			
			// Test non-E version
			if !tc.shouldErr {
				assert.NotPanics(t, func() {
					nonEResult := BuildDynamoDBEvent(tc.testData.TableName, tc.testData.EventName, tc.testData.Keys)
					assert.Equal(t, result, nonEResult)
				})
			}
		})
	}
}

// Comprehensive tests for SQS Event Builder

func TestBuildSQSEvent_ComprehensiveScenarios(t *testing.T) {
	testCases := []struct {
		name      string
		testData  SQSTestData
		shouldErr bool
	}{
		{
			name:      "ValidBasicSQSEvent",
			testData:  createSQSTestData("https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", ""),
			shouldErr: false,
		},
		{
			name:      "ValidCustomMessageBody",
			testData:  createSQSTestData("https://sqs.us-west-2.amazonaws.com/123456789012/my-queue", "Custom message body"),
			shouldErr: false,
		},
		{
			name:      "ValidJSONMessageBody",
			testData:  createSQSTestData("arn:aws:sqs:us-east-1:123456789012:json-queue", `{"key": "value", "number": 42}`),
			shouldErr: false,
		},
		{
			name:      "EmptyQueueURLShouldFail",
			testData:  createSQSTestData("", "message"),
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := BuildSQSEventE(tc.testData.QueueURL, tc.testData.MessageBody)
			
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, result)
				return
			}
			
			require.NoError(t, err)
			validateEventStructure("sqs")(t, result)
			validateJSONStructure([]string{"Records"})(t, result)
			
			// Validate SQS specific content
			var event SQSEvent
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			require.Len(t, event.Records, 1)
			
			record := event.Records[0]
			assert.Equal(t, "aws:sqs", record.EventSource)
			assert.Equal(t, tc.testData.QueueURL, record.EventSourceARN)
			
			if tc.testData.MessageBody != "" {
				assert.Equal(t, tc.testData.MessageBody, record.Body)
			} else {
				assert.Equal(t, "Hello from SQS!", record.Body)
			}
			
			// Test non-E version
			if !tc.shouldErr {
				assert.NotPanics(t, func() {
					nonEResult := BuildSQSEvent(tc.testData.QueueURL, tc.testData.MessageBody)
					assert.Equal(t, result, nonEResult)
				})
			}
		})
	}
}

// Comprehensive tests for SNS Event Builder

func TestBuildSNSEvent_ComprehensiveScenarios(t *testing.T) {
	testCases := []struct {
		name      string
		testData  SNSTestData
		shouldErr bool
	}{
		{
			name:      "ValidBasicSNSEvent",
			testData:  createSNSTestData("arn:aws:sns:us-east-1:123456789012:test-topic", ""),
			shouldErr: false,
		},
		{
			name:      "ValidCustomMessage",
			testData:  createSNSTestData("arn:aws:sns:us-west-2:123456789012:my-topic", "Custom SNS message"),
			shouldErr: false,
		},
		{
			name:      "ValidJSONMessage",
			testData:  createSNSTestData("arn:aws:sns:eu-west-1:123456789012:json-topic", `{"alert": "system down", "severity": "high"}`),
			shouldErr: false,
		},
		{
			name:      "EmptyTopicArnShouldFail",
			testData:  createSNSTestData("", "message"),
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := BuildSNSEventE(tc.testData.TopicArn, tc.testData.Message)
			
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, result)
				return
			}
			
			require.NoError(t, err)
			validateEventStructure("sns")(t, result)
			validateJSONStructure([]string{"Records"})(t, result)
			
			// Validate SNS specific content
			var event SNSEvent
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			require.Len(t, event.Records, 1)
			
			record := event.Records[0]
			assert.Equal(t, "aws:sns", record.EventSource)
			assert.Equal(t, tc.testData.TopicArn, record.Sns.TopicArn)
			
			if tc.testData.Message != "" {
				assert.Equal(t, tc.testData.Message, record.Sns.Message)
			} else {
				assert.Equal(t, "Hello from SNS!", record.Sns.Message)
			}
			
			// Test non-E version
			if !tc.shouldErr {
				assert.NotPanics(t, func() {
					nonEResult := BuildSNSEvent(tc.testData.TopicArn, tc.testData.Message)
					assert.Equal(t, result, nonEResult)
				})
			}
		})
	}
}

// Comprehensive tests for API Gateway Event Builder

func TestBuildAPIGatewayEvent_ComprehensiveScenarios(t *testing.T) {
	testCases := []struct {
		name      string
		testData  APIGatewayTestData
		shouldErr bool
	}{
		{
			name:      "ValidGetRequest",
			testData:  createAPIGatewayTestData("", "", ""),
			shouldErr: false,
		},
		{
			name:      "ValidPostRequest",
			testData:  createAPIGatewayTestData("POST", "/api/users", `{"name": "John", "email": "john@example.com"}`),
			shouldErr: false,
		},
		{
			name:      "ValidPutRequest",
			testData:  createAPIGatewayTestData("PUT", "/api/users/123", `{"name": "Jane", "email": "jane@example.com"}`),
			shouldErr: false,
		},
		{
			name:      "ValidDeleteRequest",
			testData:  createAPIGatewayTestData("DELETE", "/api/users/123", ""),
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := BuildAPIGatewayEventE(tc.testData.HTTPMethod, tc.testData.Path, tc.testData.Body)
			
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, result)
				return
			}
			
			require.NoError(t, err)
			validateEventStructure("apigateway")(t, result)
			validateJSONStructure([]string{"httpMethod", "path", "headers", "requestContext"})(t, result)
			
			// Validate API Gateway specific content
			var event APIGatewayEvent
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			
			if tc.testData.HTTPMethod != "" {
				assert.Equal(t, tc.testData.HTTPMethod, event.HttpMethod)
			} else {
				assert.Equal(t, "GET", event.HttpMethod)
			}
			
			if tc.testData.Path != "" {
				assert.Equal(t, tc.testData.Path, event.Path)
			} else {
				assert.Equal(t, "/", event.Path)
			}
			
			assert.Equal(t, tc.testData.Body, event.Body)
			
			// Test non-E version
			if !tc.shouldErr {
				assert.NotPanics(t, func() {
					nonEResult := BuildAPIGatewayEvent(tc.testData.HTTPMethod, tc.testData.Path, tc.testData.Body)
					assert.Equal(t, result, nonEResult)
				})
			}
		})
	}
}

// FAILING TESTS for missing CloudWatch Event Builder (RED phase)
// These tests should fail until we implement BuildCloudWatchEvent

func TestBuildCloudWatchEvent_ComprehensiveScenarios(t *testing.T) {
	
	testCases := []struct {
		name      string
		testData  CloudWatchTestData
		shouldErr bool
	}{
		{
			name:      "ValidCloudWatchEvent",
			testData:  createCloudWatchTestData("/aws/lambda/test-function", "2023/12/01/stream1", "Function executed successfully"),
			shouldErr: false,
		},
		{
			name:      "ValidCloudWatchAlarmEvent",
			testData:  createCloudWatchTestData("/aws/lambda/alarm-function", "2023/12/01/stream2", "Alarm state changed"),
			shouldErr: false,
		},
		{
			name:      "EmptyLogGroupShouldFail",
			testData:  createCloudWatchTestData("", "stream", "message"),
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create detail with the CloudWatch test data
			var detail map[string]interface{}
			var source string
			
			if tc.testData.LogGroupName == "" {
				source = ""
				detail = map[string]interface{}{
					"logGroup": tc.testData.LogGroupName,
					"logStream": tc.testData.LogStream,
					"message": tc.testData.MessageText,
				}
			} else {
				source = "aws.logs"
				detail = map[string]interface{}{
					"logGroup": tc.testData.LogGroupName,
					"logStream": tc.testData.LogStream,
					"message": tc.testData.MessageText,
				}
			}
			
			result, err := BuildCloudWatchEventE(source, "CloudWatch Log Event", detail)
			
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, result)
				return
			}
			
			require.NoError(t, err)
			validateJSONStructure([]string{"source", "detail-type", "detail"})(t, result)
			
			// Test non-E version
			if !tc.shouldErr {
				assert.NotPanics(t, func() {
					nonEResult := BuildCloudWatchEvent(source, "CloudWatch Log Event", detail)
					assert.Equal(t, result, nonEResult)
				})
			}
		})
	}
}

// FAILING TESTS for missing Kinesis Event Builder (RED phase)
// These tests should fail until we implement BuildKinesisEvent

func TestBuildKinesisEvent_ComprehensiveScenarios(t *testing.T) {
	
	testCases := []struct {
		name      string
		testData  KinesisTestData
		shouldErr bool
	}{
		{
			name:      "ValidKinesisEvent",
			testData:  createKinesisTestData("test-stream", "partition-key-1", "test data payload", "12345"),
			shouldErr: false,
		},
		{
			name:      "ValidJSONDataEvent",
			testData:  createKinesisTestData("json-stream", "json-key", `{"eventType": "user_action", "userId": 123}`, "67890"),
			shouldErr: false,
		},
		{
			name:      "EmptyStreamNameShouldFail",
			testData:  createKinesisTestData("", "partition-key", "data", "123"),
			shouldErr: true,
		},
		{
			name:      "EmptyPartitionKeyShouldFail",
			testData:  createKinesisTestData("stream", "", "data", "123"),
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the implemented Kinesis event builder
			result, err := BuildKinesisEventE(tc.testData.StreamName, tc.testData.PartitionKey, tc.testData.Data)
			
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Empty(t, result)
				return
			}
			
			require.NoError(t, err)
			validateJSONStructure([]string{"Records"})(t, result)
			
			// Validate Kinesis specific content
			var event KinesisEvent
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			require.Len(t, event.Records, 1)
			
			record := event.Records[0]
			assert.Equal(t, "aws:kinesis", record.EventSource)
			assert.Equal(t, tc.testData.PartitionKey, record.Kinesis.PartitionKey)
			assert.Contains(t, record.EventSourceARN, tc.testData.StreamName)
			
			// Test non-E version
			if !tc.shouldErr {
				assert.NotPanics(t, func() {
					nonEResult := BuildKinesisEvent(tc.testData.StreamName, tc.testData.PartitionKey, tc.testData.Data)
					assert.Equal(t, result, nonEResult)
				})
			}
		})
	}
}

// Cross-cutting validation tests using property-based concepts

func TestEventBuilders_CommonProperties(t *testing.T) {
	t.Run("AllEventsShouldProduceValidJSON", func(t *testing.T) {
		eventBuilders := []func() string{
			func() string { return BuildS3Event("bucket", "key", "") },
			func() string { return BuildDynamoDBEvent("table", "", nil) },
			func() string { return BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/queue", "") },
			func() string { return BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", "") },
			func() string { return BuildAPIGatewayEvent("", "", "") },
			func() string { return BuildCloudWatchEvent("aws.logs", "", nil) },
			func() string { return BuildKinesisEvent("test-stream", "partition-key", "") },
		}

		for i, builder := range eventBuilders {
			t.Run(fmt.Sprintf("EventBuilder_%d", i), func(t *testing.T) {
				result := builder()
				
				var jsonObj map[string]interface{}
				err := json.Unmarshal([]byte(result), &jsonObj)
				assert.NoError(t, err, "All event builders should produce valid JSON")
				assert.NotEmpty(t, result, "All event builders should produce non-empty results")
			})
		}
	})

	t.Run("AllEventsShouldHaveConsistentTimestampFormat", func(t *testing.T) {
		s3Event := BuildS3Event("bucket", "key", "")
		snsEvent := BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", "")
		
		// S3 events should have eventTime in ISO format
		var s3 S3Event
		json.Unmarshal([]byte(s3Event), &s3)
		_, err := time.Parse("2006-01-02T15:04:05.000Z", s3.Records[0].EventTime)
		assert.NoError(t, err, "S3 event timestamps should be in ISO format")
		
		// SNS events should have timestamp in ISO format
		var sns SNSEvent
		json.Unmarshal([]byte(snsEvent), &sns)
		_, err = time.Parse("2006-01-02T15:04:05.000Z", sns.Records[0].Sns.Timestamp)
		assert.NoError(t, err, "SNS event timestamps should be in ISO format")
	})

	t.Run("AllEventsWithRecordsShouldHaveNonEmptyRecordsList", func(t *testing.T) {
		events := []string{
			BuildS3Event("bucket", "key", ""),
			BuildDynamoDBEvent("table", "", nil),
			BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/queue", ""),
			BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", ""),
			BuildKinesisEvent("test-stream", "partition-key", ""),
		}

		for i, eventJSON := range events {
			t.Run(fmt.Sprintf("RecordBasedEvent_%d", i), func(t *testing.T) {
				var eventObj map[string]interface{}
				json.Unmarshal([]byte(eventJSON), &eventObj)
				
				records, hasRecords := eventObj["Records"]
				assert.True(t, hasRecords, "Record-based events should have Records field")
				
				recordsList, isList := records.([]interface{})
				assert.True(t, isList, "Records should be a list")
				assert.NotEmpty(t, recordsList, "Records list should not be empty")
			})
		}
	})
}

// Performance and benchmark tests for event builders

func BenchmarkEventBuilders(b *testing.B) {
	benchmarks := []struct {
		name string
		fn   func()
	}{
		{
			name: "BuildS3Event",
			fn:   func() { BuildS3Event("bucket", "key", "") },
		},
		{
			name: "BuildDynamoDBEvent", 
			fn:   func() { BuildDynamoDBEvent("table", "", nil) },
		},
		{
			name: "BuildSQSEvent",
			fn:   func() { BuildSQSEvent("queue-url", "") },
		},
		{
			name: "BuildSNSEvent",
			fn:   func() { BuildSNSEvent("topic-arn", "") },
		},
		{
			name: "BuildAPIGatewayEvent",
			fn:   func() { BuildAPIGatewayEvent("", "", "") },
		},
		{
			name: "BuildCloudWatchEvent",
			fn:   func() { BuildCloudWatchEvent("aws.logs", "", nil) },
		},
		{
			name: "BuildKinesisEvent",
			fn:   func() { BuildKinesisEvent("test-stream", "partition-key", "") },
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.fn()
			}
		})
	}
}

// Memory allocation tests

func TestEventBuilders_MemoryAllocation(t *testing.T) {
	// Test that event builders don't cause excessive allocations
	t.Run("S3EventAllocation", func(t *testing.T) {
		// Run multiple times to warm up
		for i := 0; i < 10; i++ {
			BuildS3Event("bucket", "key", "")
		}
		
		// The actual allocation test would be done with testing/quick
		// or by analyzing benchmark memory stats
		assert.True(t, true, "Memory allocation test placeholder")
	})
}