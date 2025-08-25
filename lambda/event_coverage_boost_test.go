package lambda

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddS3EventRecordE_CoverageBoost tests the AddS3EventRecord functionality
func TestAddS3EventRecordE_CoverageBoost(t *testing.T) {
	t.Run("AddS3EventRecord_ValidInputs", func(t *testing.T) {
		initialEvent := BuildS3Event("bucket1", "key1.txt", "s3:ObjectCreated:Put")
		
		result, err := AddS3EventRecordE(initialEvent, "bucket2", "key2.txt", "s3:ObjectCreated:Post")
		
		require.NoError(t, err)
		
		var event S3Event
		err = json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		
		assert.Len(t, event.Records, 2)
		assert.Equal(t, "bucket1", event.Records[0].S3.Bucket.Name)
		assert.Equal(t, "bucket2", event.Records[1].S3.Bucket.Name)
	})
	
	t.Run("AddS3EventRecord_ErrorCases", func(t *testing.T) {
		validEvent := BuildS3Event("bucket", "key", "event")
		
		// Test empty bucket name
		_, err := AddS3EventRecordE(validEvent, "", "key", "event")
		assert.Error(t, err)
		
		// Test empty object key
		_, err = AddS3EventRecordE(validEvent, "bucket", "", "event")
		assert.Error(t, err)
	})
}

// TestAddS3EventRecord_CoverageBoost tests the non-E version
func TestAddS3EventRecord_CoverageBoost(t *testing.T) {
	initialEvent := BuildS3Event("bucket1", "key1.txt", "s3:ObjectCreated:Put")
	
	result := AddS3EventRecord(initialEvent, "bucket2", "key2.txt", "s3:ObjectCreated:Post")
	
	var event S3Event
	err := json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	
	assert.Len(t, event.Records, 2)
}

// TestParseS3Event_CoverageBoost tests the S3 event parsing functionality
func TestParseS3Event_CoverageBoost(t *testing.T) {
	t.Run("ParseS3EventE_ValidEvent", func(t *testing.T) {
		eventJSON := BuildS3Event("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
		
		event, err := ParseS3EventE(eventJSON)
		
		require.NoError(t, err)
		assert.Len(t, event.Records, 1)
		assert.Equal(t, "test-bucket", event.Records[0].S3.Bucket.Name)
	})
	
	t.Run("ParseS3Event_ValidEvent", func(t *testing.T) {
		eventJSON := BuildS3Event("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
		
		event := ParseS3Event(eventJSON)
		
		assert.Len(t, event.Records, 1)
		assert.Equal(t, "test-bucket", event.Records[0].S3.Bucket.Name)
	})
}

// TestExtractS3BucketName_CoverageBoost tests bucket name extraction
func TestExtractS3BucketName_CoverageBoost(t *testing.T) {
	t.Run("ExtractS3BucketNameE_ValidEvent", func(t *testing.T) {
		eventJSON := BuildS3Event("my-test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
		
		bucketName, err := ExtractS3BucketNameE(eventJSON)
		
		require.NoError(t, err)
		assert.Equal(t, "my-test-bucket", bucketName)
	})
	
	t.Run("ExtractS3BucketName_ValidEvent", func(t *testing.T) {
		eventJSON := BuildS3Event("my-test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
		
		bucketName := ExtractS3BucketName(eventJSON)
		
		assert.Equal(t, "my-test-bucket", bucketName)
	})
	
	t.Run("ExtractS3BucketNameE_EmptyRecords", func(t *testing.T) {
		emptyEventJSON := `{"Records": []}`
		
		_, err := ExtractS3BucketNameE(emptyEventJSON)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no S3 records found")
	})
}

// TestExtractS3ObjectKey_CoverageBoost tests object key extraction
func TestExtractS3ObjectKey_CoverageBoost(t *testing.T) {
	t.Run("ExtractS3ObjectKeyE_ValidEvent", func(t *testing.T) {
		eventJSON := BuildS3Event("test-bucket", "my-test-key.txt", "s3:ObjectCreated:Put")
		
		objectKey, err := ExtractS3ObjectKeyE(eventJSON)
		
		require.NoError(t, err)
		assert.Equal(t, "my-test-key.txt", objectKey)
	})
	
	t.Run("ExtractS3ObjectKey_ValidEvent", func(t *testing.T) {
		eventJSON := BuildS3Event("test-bucket", "my-test-key.txt", "s3:ObjectCreated:Put")
		
		objectKey := ExtractS3ObjectKey(eventJSON)
		
		assert.Equal(t, "my-test-key.txt", objectKey)
	})
	
	t.Run("ExtractS3ObjectKeyE_EmptyRecords", func(t *testing.T) {
		emptyEventJSON := `{"Records": []}`
		
		_, err := ExtractS3ObjectKeyE(emptyEventJSON)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no S3 records found")
	})
}

// TestValidateEventStructure_CoverageBoost tests event structure validation
func TestValidateEventStructure_CoverageBoost(t *testing.T) {
	t.Run("ValidateEventStructure_S3", func(t *testing.T) {
		validS3Event := BuildS3Event("bucket", "key", "s3:ObjectCreated:Put")
		errors := ValidateEventStructure(validS3Event, "s3")
		assert.Empty(t, errors)
		
		// Test invalid structure
		invalidS3Event := `{"Records": []}`
		errors = ValidateEventStructure(invalidS3Event, "s3")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "S3 event must have at least one record")
	})
	
	t.Run("ValidateEventStructure_DynamoDB", func(t *testing.T) {
		validDynamoDBEvent := BuildDynamoDBEvent("table", "INSERT", nil)
		errors := ValidateEventStructure(validDynamoDBEvent, "dynamodb")
		assert.Empty(t, errors)
		
		// Test invalid structure
		invalidDynamoDBEvent := `{"Records": []}`
		errors = ValidateEventStructure(invalidDynamoDBEvent, "dynamodb")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "DynamoDB event must have at least one record")
	})
	
	t.Run("ValidateEventStructure_SQS", func(t *testing.T) {
		validSQSEvent := BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/queue", "message")
		errors := ValidateEventStructure(validSQSEvent, "sqs")
		assert.Empty(t, errors)
		
		// Test invalid structure
		invalidSQSEvent := `{"Records": []}`
		errors = ValidateEventStructure(invalidSQSEvent, "sqs")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "SQS event must have at least one record")
	})
	
	t.Run("ValidateEventStructure_SNS", func(t *testing.T) {
		validSNSEvent := BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", "message")
		errors := ValidateEventStructure(validSNSEvent, "sns")
		assert.Empty(t, errors)
		
		// Test invalid structure
		invalidSNSEvent := `{"Records": []}`
		errors = ValidateEventStructure(invalidSNSEvent, "sns")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "SNS event must have at least one record")
	})
	
	t.Run("ValidateEventStructure_APIGateway", func(t *testing.T) {
		validAPIGWEvent := BuildAPIGatewayEvent("GET", "/test", "")
		errors := ValidateEventStructure(validAPIGWEvent, "apigateway")
		assert.Empty(t, errors)
		
		// Test missing HTTP method
		missingMethodEvent := `{"Path": "/test"}`
		errors = ValidateEventStructure(missingMethodEvent, "apigateway")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "API Gateway event must have an HTTP method")
		
		// Test missing path
		missingPathEvent := `{"HttpMethod": "GET"}`
		errors = ValidateEventStructure(missingPathEvent, "apigateway")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "API Gateway event must have a path")
	})
	
	t.Run("ValidateEventStructure_UnknownType", func(t *testing.T) {
		errors := ValidateEventStructure(`{}`, "unknown")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "unknown event type: unknown")
	})
	
	t.Run("ValidateEventStructure_WrongEventSources", func(t *testing.T) {
		// Test S3 with wrong event source
		wrongS3Source := `{
			"Records": [
				{"EventSource": "aws:sqs", "S3": {"Bucket": {"Name": "test"}, "Object": {"Key": "key"}}}
			]
		}`
		errors := ValidateEventStructure(wrongS3Source, "s3")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "S3 record 0 has wrong event source: aws:sqs")
		
		// Test DynamoDB with wrong event source
		wrongDynamoSource := `{
			"Records": [
				{"EventSource": "aws:s3"}
			]
		}`
		errors = ValidateEventStructure(wrongDynamoSource, "dynamodb")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "DynamoDB record 0 has wrong event source: aws:s3")
		
		// Test SQS with wrong event source
		wrongSQSSource := `{
			"Records": [
				{"EventSource": "aws:sns"}
			]
		}`
		errors = ValidateEventStructure(wrongSQSSource, "sqs")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "SQS record 0 has wrong event source: aws:sns")
		
		// Test SNS with wrong event source
		wrongSNSSource := `{
			"Records": [
				{"EventSource": "aws:sqs"}
			]
		}`
		errors = ValidateEventStructure(wrongSNSSource, "sns")
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0], "SNS record 0 has wrong event source: aws:sqs")
	})
}

// TestMarshalPayloadE_CoverageBoost tests the marshaling utility
func TestMarshalPayloadE_CoverageBoost(t *testing.T) {
	t.Run("MarshalPayloadE_ValidStruct", func(t *testing.T) {
		event := S3Event{
			Records: []S3Record{
				{
					EventVersion: "2.1",
					EventSource:  "aws:s3",
					EventName:    "s3:ObjectCreated:Put",
				},
			},
		}
		
		result, err := MarshalPayloadE(event)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
		
		// Verify it's valid JSON
		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(result), &parsed)
		assert.NoError(t, err)
	})
}