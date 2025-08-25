package lambda

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddS3EventRecordE_WithValidInputs_ShouldAddRecord tests adding records to existing S3 events
func TestAddS3EventRecordE_WithValidInputs_ShouldAddRecord(t *testing.T) {
	// Create initial event
	initialEvent := BuildS3Event("initial-bucket", "initial-key.txt", "s3:ObjectCreated:Put")
	
	// Add another record
	result, err := AddS3EventRecordE(initialEvent, "second-bucket", "second-key.txt", "s3:ObjectCreated:Post")
	
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	
	var event S3Event
	err = json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	
	// Should have 2 records now
	assert.Len(t, event.Records, 2)
	
	// Verify first record
	assert.Equal(t, "initial-bucket", event.Records[0].S3.Bucket.Name)
	assert.Equal(t, "initial-key.txt", event.Records[0].S3.Object.Key)
	assert.Equal(t, "s3:ObjectCreated:Put", event.Records[0].EventName)
	
	// Verify second record
	assert.Equal(t, "second-bucket", event.Records[1].S3.Bucket.Name)
	assert.Equal(t, "second-key.txt", event.Records[1].S3.Object.Key)
	assert.Equal(t, "s3:ObjectCreated:Post", event.Records[1].EventName)
}

// TestAddS3EventRecordE_WithInvalidJSON_ShouldReturnError tests error handling for invalid JSON
func TestAddS3EventRecordE_WithInvalidJSON_ShouldReturnError(t *testing.T) {
	invalidJSON := `{"Records": [invalid json}`
	
	result, err := AddS3EventRecordE(invalidJSON, "bucket", "key", "s3:ObjectCreated:Put")
	
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to unmarshal existing S3 event")
}

// TestAddS3EventRecord_WithValidInputs_ShouldAddRecord tests panic-free variant
func TestAddS3EventRecord_WithValidInputs_ShouldAddRecord(t *testing.T) {
	initialEvent := BuildS3Event("initial-bucket", "initial-key.txt", "s3:ObjectCreated:Put")
	
	result := AddS3EventRecord(initialEvent, "second-bucket", "second-key.txt", "s3:ObjectCreated:Post")
	
	assert.NotEmpty(t, result)
	
	var event S3Event
	err := json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	assert.Len(t, event.Records, 2)
}

// TestAddS3EventRecord_WithInvalidInputs_ShouldPanic tests panic behavior
func TestAddS3EventRecord_WithInvalidInputs_ShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		AddS3EventRecord("invalid json", "bucket", "key", "s3:ObjectCreated:Put")
	})
}

// TestParseS3EventE_WithValidJSON_ShouldParseCorrectly tests S3 event parsing
func TestParseS3EventE_WithValidJSON_ShouldParseCorrectly(t *testing.T) {
	validJSON := BuildS3Event("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
	
	event, err := ParseS3EventE(validJSON)
	
	require.NoError(t, err)
	assert.Len(t, event.Records, 1)
	assert.Equal(t, "test-bucket", event.Records[0].S3.Bucket.Name)
	assert.Equal(t, "test-key.txt", event.Records[0].S3.Object.Key)
	assert.Equal(t, "s3:ObjectCreated:Put", event.Records[0].EventName)
	assert.Equal(t, "aws:s3", event.Records[0].EventSource)
}

// TestParseS3EventE_WithInvalidJSON_ShouldReturnError tests error handling for invalid JSON
func TestParseS3EventE_WithInvalidJSON_ShouldReturnError(t *testing.T) {
	tests := []struct {
		name        string
		invalidJSON string
	}{
		{
			name:        "completely invalid JSON",
			invalidJSON: `{invalid json`,
		},
		{
			name:        "empty string",
			invalidJSON: "",
		},
		{
			name:        "null JSON",
			invalidJSON: "null",
		},
		{
			name:        "wrong structure",
			invalidJSON: `{"wrongField": "value"}`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseS3EventE(tt.invalidJSON)
			
			assert.Error(t, err)
			assert.Equal(t, S3Event{}, event)
			assert.Contains(t, err.Error(), "failed to parse S3 event")
		})
	}
}

// TestParseS3Event_WithValidJSON_ShouldParseCorrectly tests panic-free variant
func TestParseS3Event_WithValidJSON_ShouldParseCorrectly(t *testing.T) {
	validJSON := BuildS3Event("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
	
	event := ParseS3Event(validJSON)
	
	assert.Len(t, event.Records, 1)
	assert.Equal(t, "test-bucket", event.Records[0].S3.Bucket.Name)
}

// TestParseS3Event_WithInvalidJSON_ShouldPanic tests panic behavior
func TestParseS3Event_WithInvalidJSON_ShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		ParseS3Event(`{invalid json`)
	})
}

// TestExtractS3BucketNameE_WithValidEvent_ShouldExtractName tests bucket name extraction
func TestExtractS3BucketNameE_WithValidEvent_ShouldExtractName(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectKey  string
	}{
		{
			name:       "simple bucket name",
			bucketName: "test-bucket",
			objectKey:  "test-key.txt",
		},
		{
			name:       "bucket with dashes and numbers",
			bucketName: "my-test-bucket-123",
			objectKey:  "folder/subfolder/file.json",
		},
		{
			name:       "bucket with dots",
			bucketName: "my.test.bucket",
			objectKey:  "image.png",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventJSON := BuildS3Event(tt.bucketName, tt.objectKey, "s3:ObjectCreated:Put")
			
			bucketName, err := ExtractS3BucketNameE(eventJSON)
			
			require.NoError(t, err)
			assert.Equal(t, tt.bucketName, bucketName)
		})
	}
}

// TestExtractS3BucketNameE_WithInvalidEvent_ShouldReturnError tests error cases
func TestExtractS3BucketNameE_WithInvalidEvent_ShouldReturnError(t *testing.T) {
	tests := []struct {
		name      string
		eventJSON string
		errMsg    string
	}{
		{
			name:      "invalid JSON",
			eventJSON: `{invalid json`,
			errMsg:    "failed to parse S3 event",
		},
		{
			name:      "empty records",
			eventJSON: `{"Records": []}`,
			errMsg:    "no S3 records found in event",
		},
		{
			name:      "missing records field",
			eventJSON: `{}`,
			errMsg:    "no S3 records found in event",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucketName, err := ExtractS3BucketNameE(tt.eventJSON)
			
			assert.Error(t, err)
			assert.Empty(t, bucketName)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// TestExtractS3BucketName_WithValidEvent_ShouldExtractName tests panic-free variant
func TestExtractS3BucketName_WithValidEvent_ShouldExtractName(t *testing.T) {
	eventJSON := BuildS3Event("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
	
	bucketName := ExtractS3BucketName(eventJSON)
	
	assert.Equal(t, "test-bucket", bucketName)
}

// TestExtractS3BucketName_WithInvalidEvent_ShouldPanic tests panic behavior
func TestExtractS3BucketName_WithInvalidEvent_ShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		ExtractS3BucketName(`{invalid json`)
	})
}

// TestExtractS3ObjectKeyE_WithValidEvent_ShouldExtractKey tests object key extraction
func TestExtractS3ObjectKeyE_WithValidEvent_ShouldExtractKey(t *testing.T) {
	tests := []struct {
		name      string
		objectKey string
	}{
		{
			name:      "simple file name",
			objectKey: "test-file.txt",
		},
		{
			name:      "file with path",
			objectKey: "folder/subfolder/file.json",
		},
		{
			name:      "file with special characters",
			objectKey: "file with spaces & symbols.txt",
		},
		{
			name:      "file with unicode",
			objectKey: "файл-тест.txt",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventJSON := BuildS3Event("test-bucket", tt.objectKey, "s3:ObjectCreated:Put")
			
			objectKey, err := ExtractS3ObjectKeyE(eventJSON)
			
			require.NoError(t, err)
			assert.Equal(t, tt.objectKey, objectKey)
		})
	}
}

// TestExtractS3ObjectKeyE_WithInvalidEvent_ShouldReturnError tests error cases
func TestExtractS3ObjectKeyE_WithInvalidEvent_ShouldReturnError(t *testing.T) {
	tests := []struct {
		name      string
		eventJSON string
		errMsg    string
	}{
		{
			name:      "invalid JSON",
			eventJSON: `{invalid json`,
			errMsg:    "failed to parse S3 event",
		},
		{
			name:      "empty records",
			eventJSON: `{"Records": []}`,
			errMsg:    "no S3 records found in event",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objectKey, err := ExtractS3ObjectKeyE(tt.eventJSON)
			
			assert.Error(t, err)
			assert.Empty(t, objectKey)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// TestExtractS3ObjectKey_WithValidEvent_ShouldExtractKey tests panic-free variant
func TestExtractS3ObjectKey_WithValidEvent_ShouldExtractKey(t *testing.T) {
	eventJSON := BuildS3Event("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
	
	objectKey := ExtractS3ObjectKey(eventJSON)
	
	assert.Equal(t, "test-key.txt", objectKey)
}

// TestExtractS3ObjectKey_WithInvalidEvent_ShouldPanic tests panic behavior
func TestExtractS3ObjectKey_WithInvalidEvent_ShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		ExtractS3ObjectKey(`{invalid json`)
	})
}

// TestValidateEventStructure_WithValidS3Event_ShouldReturnNoErrors tests S3 event validation
func TestValidateEventStructure_WithValidS3Event_ShouldReturnNoErrors(t *testing.T) {
	validEvent := BuildS3Event("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
	
	errors := ValidateEventStructure(validEvent, "s3")
	
	assert.Empty(t, errors)
}

// TestValidateEventStructure_WithInvalidS3Event_ShouldReturnErrors tests S3 event validation errors
func TestValidateEventStructure_WithInvalidS3Event_ShouldReturnErrors(t *testing.T) {
	tests := []struct {
		name      string
		eventJSON string
		errMsg    string
	}{
		{
			name:      "invalid JSON",
			eventJSON: `{invalid json`,
			errMsg:    "invalid S3 event structure",
		},
		{
			name:      "empty records",
			eventJSON: `{"Records": []}`,
			errMsg:    "S3 event must have at least one record",
		},
		{
			name:      "wrong event source",
			eventJSON: `{"Records": [{"EventSource": "aws:dynamodb"}]}`,
			errMsg:    "S3 record 0 has wrong event source: aws:dynamodb",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateEventStructure(tt.eventJSON, "s3")
			
			assert.NotEmpty(t, errors)
			assert.Contains(t, strings.Join(errors, " "), tt.errMsg)
		})
	}
}

// TestValidateEventStructure_WithValidDynamoDBEvent_ShouldReturnNoErrors tests DynamoDB event validation
func TestValidateEventStructure_WithValidDynamoDBEvent_ShouldReturnNoErrors(t *testing.T) {
	validEvent := BuildDynamoDBEvent("test-table", "INSERT", nil)
	
	errors := ValidateEventStructure(validEvent, "dynamodb")
	
	assert.Empty(t, errors)
}

// TestValidateEventStructure_WithInvalidDynamoDBEvent_ShouldReturnErrors tests DynamoDB event validation errors
func TestValidateEventStructure_WithInvalidDynamoDBEvent_ShouldReturnErrors(t *testing.T) {
	tests := []struct {
		name      string
		eventJSON string
		errMsg    string
	}{
		{
			name:      "invalid JSON",
			eventJSON: `{invalid json`,
			errMsg:    "invalid DynamoDB event structure",
		},
		{
			name:      "empty records",
			eventJSON: `{"Records": []}`,
			errMsg:    "DynamoDB event must have at least one record",
		},
		{
			name:      "wrong event source",
			eventJSON: `{"Records": [{"eventSource": "aws:s3"}]}`,
			errMsg:    "DynamoDB record 0 has wrong event source: aws:s3",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateEventStructure(tt.eventJSON, "dynamodb")
			
			assert.NotEmpty(t, errors)
			assert.Contains(t, strings.Join(errors, " "), tt.errMsg)
		})
	}
}

// TestValidateEventStructure_WithValidSQSEvent_ShouldReturnNoErrors tests SQS event validation
func TestValidateEventStructure_WithValidSQSEvent_ShouldReturnNoErrors(t *testing.T) {
	validEvent := BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", "test message")
	
	errors := ValidateEventStructure(validEvent, "sqs")
	
	assert.Empty(t, errors)
}

// TestValidateEventStructure_WithInvalidSQSEvent_ShouldReturnErrors tests SQS event validation errors
func TestValidateEventStructure_WithInvalidSQSEvent_ShouldReturnErrors(t *testing.T) {
	tests := []struct {
		name      string
		eventJSON string
		errMsg    string
	}{
		{
			name:      "invalid JSON",
			eventJSON: `{invalid json`,
			errMsg:    "invalid SQS event structure",
		},
		{
			name:      "empty records",
			eventJSON: `{"Records": []}`,
			errMsg:    "SQS event must have at least one record",
		},
		{
			name:      "wrong event source",
			eventJSON: `{"Records": [{"eventSource": "aws:sns"}]}`,
			errMsg:    "SQS record 0 has wrong event source: aws:sns",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateEventStructure(tt.eventJSON, "sqs")
			
			assert.NotEmpty(t, errors)
			assert.Contains(t, strings.Join(errors, " "), tt.errMsg)
		})
	}
}

// TestValidateEventStructure_WithValidSNSEvent_ShouldReturnNoErrors tests SNS event validation
func TestValidateEventStructure_WithValidSNSEvent_ShouldReturnNoErrors(t *testing.T) {
	validEvent := BuildSNSEvent("arn:aws:sns:us-east-1:123456789012:test-topic", "test message")
	
	errors := ValidateEventStructure(validEvent, "sns")
	
	assert.Empty(t, errors)
}

// TestValidateEventStructure_WithInvalidSNSEvent_ShouldReturnErrors tests SNS event validation errors
func TestValidateEventStructure_WithInvalidSNSEvent_ShouldReturnErrors(t *testing.T) {
	tests := []struct {
		name      string
		eventJSON string
		errMsg    string
	}{
		{
			name:      "invalid JSON",
			eventJSON: `{invalid json`,
			errMsg:    "invalid SNS event structure",
		},
		{
			name:      "empty records",
			eventJSON: `{"Records": []}`,
			errMsg:    "SNS event must have at least one record",
		},
		{
			name:      "wrong event source",
			eventJSON: `{"Records": [{"EventSource": "aws:sqs"}]}`,
			errMsg:    "SNS record 0 has wrong event source: aws:sqs",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateEventStructure(tt.eventJSON, "sns")
			
			assert.NotEmpty(t, errors)
			assert.Contains(t, strings.Join(errors, " "), tt.errMsg)
		})
	}
}

// TestValidateEventStructure_WithValidAPIGatewayEvent_ShouldReturnNoErrors tests API Gateway event validation
func TestValidateEventStructure_WithValidAPIGatewayEvent_ShouldReturnNoErrors(t *testing.T) {
	validEvent := BuildAPIGatewayEvent("GET", "/api/test", "")
	
	errors := ValidateEventStructure(validEvent, "apigateway")
	
	assert.Empty(t, errors)
}

// TestValidateEventStructure_WithInvalidAPIGatewayEvent_ShouldReturnErrors tests API Gateway event validation errors
func TestValidateEventStructure_WithInvalidAPIGatewayEvent_ShouldReturnErrors(t *testing.T) {
	tests := []struct {
		name      string
		eventJSON string
		errMsg    string
	}{
		{
			name:      "invalid JSON",
			eventJSON: `{invalid json`,
			errMsg:    "invalid API Gateway event structure",
		},
		{
			name:      "missing HTTP method",
			eventJSON: `{"path": "/test"}`,
			errMsg:    "API Gateway event must have an HTTP method",
		},
		{
			name:      "missing path",
			eventJSON: `{"httpMethod": "GET"}`,
			errMsg:    "API Gateway event must have a path",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateEventStructure(tt.eventJSON, "apigateway")
			
			assert.NotEmpty(t, errors)
			assert.Contains(t, strings.Join(errors, " "), tt.errMsg)
		})
	}
}

// TestValidateEventStructure_WithUnknownEventType_ShouldReturnError tests unknown event type handling
func TestValidateEventStructure_WithUnknownEventType_ShouldReturnError(t *testing.T) {
	errors := ValidateEventStructure(`{"test": "data"}`, "unknown-type")
	
	assert.NotEmpty(t, errors)
	assert.Contains(t, strings.Join(errors, " "), "unknown event type: unknown-type")
}

// TestValidateEventStructure_WithComplexValidEvents_ShouldReturnNoErrors tests complex valid events
func TestValidateEventStructure_WithComplexValidEvents_ShouldReturnNoErrors(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		builder   func() string
	}{
		{
			name:      "S3 event with multiple records",
			eventType: "s3",
			builder: func() string {
				initial := BuildS3Event("bucket1", "key1.txt", "s3:ObjectCreated:Put")
				return AddS3EventRecord(initial, "bucket2", "key2.txt", "s3:ObjectCreated:Post")
			},
		},
		{
			name:      "DynamoDB event with MODIFY",
			eventType: "dynamodb",
			builder: func() string {
				keys := map[string]interface{}{
					"pk": map[string]interface{}{"S": "test-pk"},
					"sk": map[string]interface{}{"S": "test-sk"},
				}
				return BuildDynamoDBEvent("test-table", "MODIFY", keys)
			},
		},
		{
			name:      "SQS event with JSON body",
			eventType: "sqs",
			builder: func() string {
				return BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123456789012/json-queue", `{"order": 123, "items": ["a", "b"]}`)
			},
		},
		{
			name:      "SNS event with large message",
			eventType: "sns",
			builder: func() string {
				largeMessage := strings.Repeat("Large message content. ", 100)
				return BuildSNSEvent("arn:aws:sns:us-east-1:123456789012:large-topic", largeMessage)
			},
		},
		{
			name:      "API Gateway event with POST and body",
			eventType: "apigateway",
			builder: func() string {
				return BuildAPIGatewayEvent("POST", "/api/users", `{"name": "John", "email": "john@example.com"}`)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventJSON := tt.builder()
			errors := ValidateEventStructure(eventJSON, tt.eventType)
			
			assert.Empty(t, errors, "Expected no validation errors for %s", tt.name)
		})
	}
}