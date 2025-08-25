package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test AddS3EventRecord with valid existing event
func TestAddS3EventRecord_WithValidExistingEvent_ShouldAddRecord(t *testing.T) {
	// Given: Existing S3 event JSON and new record details
	existingEvent := `{
		"Records": [
			{
				"eventVersion": "2.1",
				"eventSource": "aws:s3",
				"eventName": "s3:ObjectCreated:Put",
				"eventTime": "2023-01-01T00:00:00.000Z",
				"s3": {
					"bucket": {"name": "existing-bucket"},
					"object": {"key": "existing-key"}
				}
			}
		]
	}`
	
	// When: AddS3EventRecord is called
	result, err := AddS3EventRecordE(existingEvent, "new-bucket", "new-key", "s3:ObjectCreated:Post")
	
	// Then: Should add new record successfully
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	
	// Parse the result to verify it contains both records
	event, parseErr := ParseS3EventE(result)
	require.NoError(t, parseErr)
	require.Len(t, event.Records, 2)
	
	// Verify original record is preserved
	assert.Equal(t, "existing-bucket", event.Records[0].S3.Bucket.Name)
	assert.Equal(t, "existing-key", event.Records[0].S3.Object.Key)
	
	// Verify new record is added
	assert.Equal(t, "new-bucket", event.Records[1].S3.Bucket.Name)
	assert.Equal(t, "new-key", event.Records[1].S3.Object.Key)
	assert.Equal(t, "s3:ObjectCreated:Post", event.Records[1].EventName)
}

// RED: Test AddS3EventRecord with invalid JSON
func TestAddS3EventRecord_WithInvalidJSON_ShouldReturnError(t *testing.T) {
	// Given: Invalid JSON string
	invalidEvent := `{"invalid": "json" missing brace`
	
	// When: AddS3EventRecord is called with invalid JSON
	result, err := AddS3EventRecordE(invalidEvent, "bucket", "key", "eventName")
	
	// Then: Should return parsing error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal existing S3 event")
	assert.Empty(t, result)
}

// RED: Test AddS3EventRecord with empty bucket name
func TestAddS3EventRecord_WithEmptyBucketName_ShouldReturnError(t *testing.T) {
	// Given: Valid existing event but empty bucket name for new record
	existingEvent := `{"Records": []}`
	
	// When: AddS3EventRecord is called with empty bucket name
	result, err := AddS3EventRecordE(existingEvent, "", "key", "eventName")
	
	// Then: Should return bucket name error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
	assert.Empty(t, result)
}

// RED: Test non-error version of AddS3EventRecord
func TestAddS3EventRecord_NonErrorVersion_ShouldCallErrorVersion(t *testing.T) {
	// Given: Valid existing event
	existingEvent := `{"Records": []}`
	
	// When: AddS3EventRecord (non-error version) is called
	result := AddS3EventRecord(existingEvent, "test-bucket", "test-key", "s3:ObjectCreated:Put")
	
	// Then: Should return valid result without panicking
	assert.NotEmpty(t, result)
	
	// Verify the result is valid
	event, err := ParseS3EventE(result)
	require.NoError(t, err)
	require.Len(t, event.Records, 1)
	assert.Equal(t, "test-bucket", event.Records[0].S3.Bucket.Name)
}

// RED: Test ParseS3Event non-error version
func TestParseS3Event_NonErrorVersion_ShouldCallErrorVersion(t *testing.T) {
	// Given: Valid S3 event JSON
	eventJSON, err := BuildS3EventE("test-bucket", "test-key", "s3:ObjectCreated:Put")
	require.NoError(t, err)
	
	// When: ParseS3Event (non-error version) is called
	event := ParseS3Event(eventJSON)
	
	// Then: Should return parsed event without panicking
	require.Len(t, event.Records, 1)
	assert.Equal(t, "test-bucket", event.Records[0].S3.Bucket.Name)
	assert.Equal(t, "test-key", event.Records[0].S3.Object.Key)
}

// RED: Test ExtractS3BucketName with valid event
func TestExtractS3BucketName_WithValidEvent_ShouldReturnBucketName(t *testing.T) {
	// Given: Valid S3 event JSON
	eventJSON, err := BuildS3EventE("my-bucket", "my-key", "s3:ObjectCreated:Put")
	require.NoError(t, err)
	
	// When: ExtractS3BucketName is called
	bucketName, extractErr := ExtractS3BucketNameE(eventJSON)
	
	// Then: Should return correct bucket name
	require.NoError(t, extractErr)
	assert.Equal(t, "my-bucket", bucketName)
}

// RED: Test ExtractS3BucketName with empty records
func TestExtractS3BucketName_WithEmptyRecords_ShouldReturnError(t *testing.T) {
	// Given: S3 event with no records
	eventJSON := `{"Records": []}`
	
	// When: ExtractS3BucketName is called
	bucketName, err := ExtractS3BucketNameE(eventJSON)
	
	// Then: Should return error about no records
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no S3 records found in event")
	assert.Empty(t, bucketName)
}

// RED: Test ExtractS3BucketName with invalid JSON
func TestExtractS3BucketName_WithInvalidJSON_ShouldReturnError(t *testing.T) {
	// Given: Invalid JSON
	eventJSON := `{"invalid": "json"`
	
	// When: ExtractS3BucketName is called
	bucketName, err := ExtractS3BucketNameE(eventJSON)
	
	// Then: Should return parsing error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse S3 event")
	assert.Empty(t, bucketName)
}

// RED: Test non-error version of ExtractS3BucketName
func TestExtractS3BucketName_NonErrorVersion_ShouldCallErrorVersion(t *testing.T) {
	// Given: Valid S3 event JSON
	eventJSON, err := BuildS3EventE("test-bucket", "test-key", "s3:ObjectCreated:Put")
	require.NoError(t, err)
	
	// When: ExtractS3BucketName (non-error version) is called
	bucketName := ExtractS3BucketName(eventJSON)
	
	// Then: Should return bucket name without panicking
	assert.Equal(t, "test-bucket", bucketName)
}

// RED: Test ExtractS3ObjectKey with valid event
func TestExtractS3ObjectKey_WithValidEvent_ShouldReturnObjectKey(t *testing.T) {
	// Given: Valid S3 event JSON
	eventJSON, err := BuildS3EventE("my-bucket", "my-object-key", "s3:ObjectCreated:Put")
	require.NoError(t, err)
	
	// When: ExtractS3ObjectKey is called
	objectKey, extractErr := ExtractS3ObjectKeyE(eventJSON)
	
	// Then: Should return correct object key
	require.NoError(t, extractErr)
	assert.Equal(t, "my-object-key", objectKey)
}

// RED: Test ExtractS3ObjectKey with empty records
func TestExtractS3ObjectKey_WithEmptyRecords_ShouldReturnError(t *testing.T) {
	// Given: S3 event with no records
	eventJSON := `{"Records": []}`
	
	// When: ExtractS3ObjectKey is called
	objectKey, err := ExtractS3ObjectKeyE(eventJSON)
	
	// Then: Should return error about no records
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no S3 records found in event")
	assert.Empty(t, objectKey)
}

// RED: Test ExtractS3ObjectKey with invalid JSON
func TestExtractS3ObjectKey_WithInvalidJSON_ShouldReturnError(t *testing.T) {
	// Given: Invalid JSON
	eventJSON := `{"invalid": "json"`
	
	// When: ExtractS3ObjectKey is called
	objectKey, err := ExtractS3ObjectKeyE(eventJSON)
	
	// Then: Should return parsing error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse S3 event")
	assert.Empty(t, objectKey)
}

// RED: Test non-error version of ExtractS3ObjectKey
func TestExtractS3ObjectKey_NonErrorVersion_ShouldCallErrorVersion(t *testing.T) {
	// Given: Valid S3 event JSON
	eventJSON, err := BuildS3EventE("test-bucket", "test-object-key", "s3:ObjectCreated:Put")
	require.NoError(t, err)
	
	// When: ExtractS3ObjectKey (non-error version) is called
	objectKey := ExtractS3ObjectKey(eventJSON)
	
	// Then: Should return object key without panicking
	assert.Equal(t, "test-object-key", objectKey)
}

// RED: Test BuildDynamoDBEvent with comprehensive options
func TestBuildDynamoDBEvent_WithComprehensiveOptions_ShouldBuildValidEvent(t *testing.T) {
	// Given: DynamoDB event parameters
	tableName := "test-table"
	eventName := "INSERT"
	keys := map[string]interface{}{
		"id":   map[string]interface{}{"S": "user-123"},
		"type": map[string]interface{}{"S": "customer"},
	}
	
	// When: BuildDynamoDBEvent is called
	eventJSON := BuildDynamoDBEvent(tableName, eventName, keys)
	
	// Then: Should return valid DynamoDB event JSON
	assert.NotEmpty(t, eventJSON)
	
	// Parse back to verify structure
	var event DynamoDBEvent
	parseErr := ParseInvokeOutputE(&InvokeResult{Payload: eventJSON}, &event)
	require.NoError(t, parseErr)
	require.Len(t, event.Records, 1)
	
	record := event.Records[0]
	assert.Equal(t, eventName, record.EventName)
	assert.Equal(t, "aws:dynamodb", record.EventSource)
	assert.NotEmpty(t, record.EventID)
	assert.Equal(t, keys, record.Dynamodb.Keys)
}

// RED: Test BuildSQSEvent with comprehensive options
func TestBuildSQSEvent_WithComprehensiveOptions_ShouldBuildValidEvent(t *testing.T) {
	// Given: SQS event parameters
	queueUrl := "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
	messageBody := "Test SQS message body"
	
	// When: BuildSQSEvent is called
	eventJSON := BuildSQSEvent(queueUrl, messageBody)
	
	// Then: Should return valid SQS event JSON
	assert.NotEmpty(t, eventJSON)
	
	// Parse back to verify structure
	var event SQSEvent
	parseErr := ParseInvokeOutputE(&InvokeResult{Payload: eventJSON}, &event)
	require.NoError(t, parseErr)
	require.Len(t, event.Records, 1)
	
	record := event.Records[0]
	assert.Equal(t, messageBody, record.Body)
	assert.Equal(t, "aws:sqs", record.EventSource)
	assert.Equal(t, queueUrl, record.EventSourceARN)
	assert.NotEmpty(t, record.MessageId)
}

// RED: Test BuildSNSEvent with comprehensive options
func TestBuildSNSEvent_WithComprehensiveOptions_ShouldBuildValidEvent(t *testing.T) {
	// Given: SNS event parameters
	topicArn := "arn:aws:sns:us-east-1:123456789012:test-topic"
	message := "Test SNS notification message"
	
	// When: BuildSNSEvent is called
	eventJSON := BuildSNSEvent(topicArn, message)
	
	// Then: Should return valid SNS event JSON
	assert.NotEmpty(t, eventJSON)
	
	// Parse back to verify structure
	var event SNSEvent
	parseErr := ParseInvokeOutputE(&InvokeResult{Payload: eventJSON}, &event)
	require.NoError(t, parseErr)
	require.Len(t, event.Records, 1)
	
	record := event.Records[0]
	assert.Equal(t, message, record.Sns.Message)
	assert.Equal(t, "aws:sns", record.EventSource)
	assert.Equal(t, topicArn, record.Sns.TopicArn)
	assert.NotEmpty(t, record.Sns.MessageId)
}