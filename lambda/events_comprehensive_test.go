package lambda

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildS3EventE_WithValidInputs_ShouldReturnValidEvent tests S3 event building with various valid inputs
func TestBuildS3EventE_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	tests := []struct {
		name        string
		bucketName  string
		objectKey   string
		eventName   string
		expectError bool
	}{
		{
			name:        "valid inputs with default event name",
			bucketName:  "test-bucket",
			objectKey:   "test-key.txt",
			eventName:   "",
			expectError: false,
		},
		{
			name:        "valid inputs with custom event name",
			bucketName:  "my-bucket",
			objectKey:   "folder/file.json",
			eventName:   "s3:ObjectCreated:Post",
			expectError: false,
		},
		{
			name:        "valid inputs with delete event",
			bucketName:  "delete-bucket",
			objectKey:   "to-delete.txt",
			eventName:   "s3:ObjectRemoved:Delete",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildS3EventE(tt.bucketName, tt.objectKey, tt.eventName)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			
			// Parse the result to validate structure
			var event S3Event
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			
			assert.Len(t, event.Records, 1)
			record := event.Records[0]
			
			assert.Equal(t, "2.1", record.EventVersion)
			assert.Equal(t, "aws:s3", record.EventSource)
			assert.Equal(t, tt.bucketName, record.S3.Bucket.Name)
			assert.Equal(t, tt.objectKey, record.S3.Object.Key)
			
			if tt.eventName == "" {
				assert.Equal(t, "s3:ObjectCreated:Put", record.EventName)
			} else {
				assert.Equal(t, tt.eventName, record.EventName)
			}
			
			// Validate ARN format
			expectedArn := "arn:aws:s3:::" + tt.bucketName
			assert.Equal(t, expectedArn, record.S3.Bucket.Arn)
			
			// Validate timestamp format
			assert.NotEmpty(t, record.EventTime)
			_, err = time.Parse("2006-01-02T15:04:05.000Z", record.EventTime)
			assert.NoError(t, err)
		})
	}
}

// TestBuildS3EventE_WithInvalidInputs_ShouldReturnError tests S3 event building error cases
func TestBuildS3EventE_WithInvalidInputs_ShouldReturnError(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectKey  string
		eventName  string
	}{
		{
			name:       "empty bucket name",
			bucketName: "",
			objectKey:  "test-key.txt",
			eventName:  "s3:ObjectCreated:Put",
		},
		{
			name:       "empty object key",
			bucketName: "test-bucket",
			objectKey:  "",
			eventName:  "s3:ObjectCreated:Put",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildS3EventE(tt.bucketName, tt.objectKey, tt.eventName)
			
			assert.Error(t, err)
			assert.Empty(t, result)
		})
	}
}

// TestBuildS3Event_WithValidInputs_ShouldReturnValidEvent tests panic-free variant
func TestBuildS3Event_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	result := BuildS3Event("test-bucket", "test-key.txt", "s3:ObjectCreated:Put")
	
	assert.NotEmpty(t, result)
	
	var event S3Event
	err := json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	assert.Len(t, event.Records, 1)
}

// TestBuildS3Event_WithInvalidInputs_ShouldPanic tests panic behavior
func TestBuildS3Event_WithInvalidInputs_ShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		BuildS3Event("", "test-key.txt", "s3:ObjectCreated:Put")
	})
}

// TestBuildDynamoDBEventE_WithValidInputs_ShouldReturnValidEvent tests DynamoDB event building
func TestBuildDynamoDBEventE_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		eventName string
		keys      map[string]interface{}
	}{
		{
			name:      "valid inputs with default values",
			tableName: "test-table",
			eventName: "",
			keys:      nil,
		},
		{
			name:      "valid inputs with custom event name",
			tableName: "my-table",
			eventName: "MODIFY",
			keys:      map[string]interface{}{"id": map[string]interface{}{"S": "custom-id"}},
		},
		{
			name:      "valid inputs with remove event",
			tableName: "delete-table",
			eventName: "REMOVE",
			keys:      map[string]interface{}{"pk": map[string]interface{}{"S": "delete-me"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildDynamoDBEventE(tt.tableName, tt.eventName, tt.keys)
			
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			
			var event DynamoDBEvent
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			
			assert.Len(t, event.Records, 1)
			record := event.Records[0]
			
			assert.Equal(t, "1.1", record.EventVersion)
			assert.Equal(t, "aws:dynamodb", record.EventSource)
			assert.Equal(t, "us-east-1", record.AwsRegion)
			
			if tt.eventName == "" {
				assert.Equal(t, "INSERT", record.EventName)
			} else {
				assert.Equal(t, tt.eventName, record.EventName)
			}
			
			// Validate keys
			if tt.keys == nil {
				expectedKeys := map[string]interface{}{"id": map[string]interface{}{"S": "example-id"}}
				assert.Equal(t, expectedKeys, record.Dynamodb.Keys)
			} else {
				assert.Equal(t, tt.keys, record.Dynamodb.Keys)
			}
			
			// Validate DynamoDB stream record structure
			assert.Equal(t, "111", record.Dynamodb.SequenceNumber)
			assert.Equal(t, int64(26), record.Dynamodb.SizeBytes)
			assert.Equal(t, "KEYS_ONLY", record.Dynamodb.StreamViewType)
		})
	}
}

// TestBuildDynamoDBEventE_WithInvalidInputs_ShouldReturnError tests DynamoDB event building errors
func TestBuildDynamoDBEventE_WithInvalidInputs_ShouldReturnError(t *testing.T) {
	result, err := BuildDynamoDBEventE("", "INSERT", nil)
	
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "table name cannot be empty")
}

// TestBuildDynamoDBEvent_WithValidInputs_ShouldReturnValidEvent tests panic-free variant
func TestBuildDynamoDBEvent_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	result := BuildDynamoDBEvent("test-table", "INSERT", nil)
	
	assert.NotEmpty(t, result)
	
	var event DynamoDBEvent
	err := json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	assert.Len(t, event.Records, 1)
}

// TestBuildDynamoDBEvent_WithInvalidInputs_ShouldPanic tests panic behavior
func TestBuildDynamoDBEvent_WithInvalidInputs_ShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		BuildDynamoDBEvent("", "INSERT", nil)
	})
}

// TestBuildSQSEventE_WithValidInputs_ShouldReturnValidEvent tests SQS event building
func TestBuildSQSEventE_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	tests := []struct {
		name        string
		queueUrl    string
		messageBody string
	}{
		{
			name:        "valid inputs with default message",
			queueUrl:    "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			messageBody: "",
		},
		{
			name:        "valid inputs with custom message",
			queueUrl:    "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue",
			messageBody: "Custom SQS message content",
		},
		{
			name:        "valid inputs with JSON message",
			queueUrl:    "https://sqs.us-east-1.amazonaws.com/123456789012/json-queue",
			messageBody: `{"type": "order", "id": 12345}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildSQSEventE(tt.queueUrl, tt.messageBody)
			
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			
			var event SQSEvent
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			
			assert.Len(t, event.Records, 1)
			record := event.Records[0]
			
			assert.Equal(t, "aws:sqs", record.EventSource)
			assert.Equal(t, tt.queueUrl, record.EventSourceARN)
			assert.Equal(t, "us-east-1", record.AwsRegion)
			
			if tt.messageBody == "" {
				assert.Equal(t, "Hello from SQS!", record.Body)
			} else {
				assert.Equal(t, tt.messageBody, record.Body)
			}
			
			// Validate required fields
			assert.NotEmpty(t, record.MessageId)
			assert.NotEmpty(t, record.ReceiptHandle)
			assert.NotEmpty(t, record.Md5OfBody)
			assert.NotNil(t, record.Attributes)
			assert.NotNil(t, record.MessageAttributes)
		})
	}
}

// TestBuildSQSEventE_WithInvalidInputs_ShouldReturnError tests SQS event building errors
func TestBuildSQSEventE_WithInvalidInputs_ShouldReturnError(t *testing.T) {
	result, err := BuildSQSEventE("", "test message")
	
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "queue URL cannot be empty")
}

// TestBuildSQSEvent_WithValidInputs_ShouldReturnValidEvent tests panic-free variant
func TestBuildSQSEvent_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	result := BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", "test message")
	
	assert.NotEmpty(t, result)
	
	var event SQSEvent
	err := json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	assert.Len(t, event.Records, 1)
}

// TestBuildSQSEvent_WithInvalidInputs_ShouldPanic tests panic behavior
func TestBuildSQSEvent_WithInvalidInputs_ShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		BuildSQSEvent("", "test message")
	})
}

// TestBuildSNSEventE_WithValidInputs_ShouldReturnValidEvent tests SNS event building
func TestBuildSNSEventE_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	tests := []struct {
		name     string
		topicArn string
		message  string
	}{
		{
			name:     "valid inputs with default message",
			topicArn: "arn:aws:sns:us-east-1:123456789012:test-topic",
			message:  "",
		},
		{
			name:     "valid inputs with custom message",
			topicArn: "arn:aws:sns:us-east-1:123456789012:my-topic",
			message:  "Custom SNS message content",
		},
		{
			name:     "valid inputs with JSON message",
			topicArn: "arn:aws:sns:us-east-1:123456789012:json-topic",
			message:  `{"alert": "system down", "severity": "critical"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildSNSEventE(tt.topicArn, tt.message)
			
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			
			var event SNSEvent
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			
			assert.Len(t, event.Records, 1)
			record := event.Records[0]
			
			assert.Equal(t, "1.0", record.EventVersion)
			assert.Equal(t, "aws:sns", record.EventSource)
			assert.Equal(t, tt.topicArn, record.Sns.TopicArn)
			
			if tt.message == "" {
				assert.Equal(t, "Hello from SNS!", record.Sns.Message)
			} else {
				assert.Equal(t, tt.message, record.Sns.Message)
			}
			
			// Validate subscription ARN format
			expectedPrefix := tt.topicArn + ":"
			assert.True(t, strings.HasPrefix(record.EventSubscriptionArn, expectedPrefix))
			
			// Validate required fields
			assert.Equal(t, "1", record.Sns.SignatureVersion)
			assert.Equal(t, "Notification", record.Sns.Type)
			assert.Equal(t, "TestInvoke", record.Sns.Subject)
			assert.NotEmpty(t, record.Sns.MessageId)
			assert.NotEmpty(t, record.Sns.Timestamp)
			assert.NotNil(t, record.Sns.MessageAttributes)
		})
	}
}

// TestBuildSNSEventE_WithInvalidInputs_ShouldReturnError tests SNS event building errors
func TestBuildSNSEventE_WithInvalidInputs_ShouldReturnError(t *testing.T) {
	result, err := BuildSNSEventE("", "test message")
	
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "topic ARN cannot be empty")
}

// TestBuildSNSEvent_WithValidInputs_ShouldReturnValidEvent tests panic-free variant
func TestBuildSNSEvent_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	result := BuildSNSEvent("arn:aws:sns:us-east-1:123456789012:test-topic", "test message")
	
	assert.NotEmpty(t, result)
	
	var event SNSEvent
	err := json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	assert.Len(t, event.Records, 1)
}

// TestBuildSNSEvent_WithInvalidInputs_ShouldPanic tests panic behavior
func TestBuildSNSEvent_WithInvalidInputs_ShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		BuildSNSEvent("", "test message")
	})
}

// TestBuildAPIGatewayEventE_WithValidInputs_ShouldReturnValidEvent tests API Gateway event building
func TestBuildAPIGatewayEventE_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	tests := []struct {
		name       string
		httpMethod string
		path       string
		body       string
	}{
		{
			name:       "valid inputs with defaults",
			httpMethod: "",
			path:       "",
			body:       "",
		},
		{
			name:       "valid POST with JSON body",
			httpMethod: "POST",
			path:       "/api/users",
			body:       `{"name": "John", "email": "john@example.com"}`,
		},
		{
			name:       "valid GET with query parameters path",
			httpMethod: "GET",
			path:       "/api/users/123",
			body:       "",
		},
		{
			name:       "valid PUT with XML body",
			httpMethod: "PUT",
			path:       "/api/orders/456",
			body:       `<order><id>456</id><status>completed</status></order>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildAPIGatewayEventE(tt.httpMethod, tt.path, tt.body)
			
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			
			var event APIGatewayEvent
			err = json.Unmarshal([]byte(result), &event)
			require.NoError(t, err)
			
			expectedMethod := tt.httpMethod
			if expectedMethod == "" {
				expectedMethod = "GET"
			}
			assert.Equal(t, expectedMethod, event.HttpMethod)
			
			expectedPath := tt.path
			if expectedPath == "" {
				expectedPath = "/"
			}
			assert.Equal(t, expectedPath, event.Path)
			
			assert.Equal(t, tt.body, event.Body)
			assert.Equal(t, "/{proxy+}", event.Resource)
			assert.False(t, event.IsBase64Encoded)
			
			// Validate request context
			assert.Equal(t, "123456789012", event.RequestContext.AccountId)
			assert.Equal(t, "1234567890", event.RequestContext.ApiId)
			assert.Equal(t, expectedMethod, event.RequestContext.HttpMethod)
			assert.Equal(t, "prod", event.RequestContext.Stage)
			
			// Validate identity
			assert.Equal(t, "203.0.113.1", event.RequestContext.Identity.SourceIp)
			assert.Equal(t, "PostmanRuntime/2.4.5", event.RequestContext.Identity.UserAgent)
			
			// Validate required maps are initialized
			assert.NotNil(t, event.Headers)
			assert.NotNil(t, event.MultiValueHeaders)
			assert.NotNil(t, event.QueryStringParameters)
			assert.NotNil(t, event.MultiValueQueryStringParameters)
			assert.NotNil(t, event.PathParameters)
			assert.NotNil(t, event.StageVariables)
			
			// Validate some standard headers
			assert.Contains(t, event.Headers, "Host")
			assert.Contains(t, event.Headers, "User-Agent")
		})
	}
}

// TestBuildAPIGatewayEvent_WithValidInputs_ShouldReturnValidEvent tests panic-free variant
func TestBuildAPIGatewayEvent_WithValidInputs_ShouldReturnValidEvent(t *testing.T) {
	result := BuildAPIGatewayEvent("POST", "/api/test", `{"test": true}`)
	
	assert.NotEmpty(t, result)
	
	var event APIGatewayEvent
	err := json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	
	assert.Equal(t, "POST", event.HttpMethod)
	assert.Equal(t, "/api/test", event.Path)
	assert.Equal(t, `{"test": true}`, event.Body)
}

// TestBuildAPIGatewayEventE_WithAllDefaults_ShouldUseCorrectDefaults tests default handling
func TestBuildAPIGatewayEventE_WithAllDefaults_ShouldUseCorrectDefaults(t *testing.T) {
	result, err := BuildAPIGatewayEventE("", "", "")
	
	require.NoError(t, err)
	
	var event APIGatewayEvent
	err = json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	
	assert.Equal(t, "GET", event.HttpMethod)
	assert.Equal(t, "/", event.Path)
	assert.Equal(t, "", event.Body)
}

// TestEventBuildingWithLargePayloads_ShouldHandleCorrectly tests large payload handling
func TestEventBuildingWithLargePayloads_ShouldHandleCorrectly(t *testing.T) {
	largeBody := strings.Repeat("a", 1000)
	
	result, err := BuildAPIGatewayEventE("POST", "/api/large", largeBody)
	require.NoError(t, err)
	
	var event APIGatewayEvent
	err = json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	
	assert.Equal(t, largeBody, event.Body)
	assert.Equal(t, 1000, len(event.Body))
}

// TestEventBuildingWithSpecialCharacters_ShouldEscapeCorrectly tests special character handling
func TestEventBuildingWithSpecialCharacters_ShouldEscapeCorrectly(t *testing.T) {
	specialBody := `{"message": "Test with \"quotes\", \nnewlines, and \ttabs"}`
	
	result, err := BuildSNSEventE("arn:aws:sns:us-east-1:123456789012:test", specialBody)
	require.NoError(t, err)
	
	var event SNSEvent
	err = json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	
	assert.Equal(t, specialBody, event.Records[0].Sns.Message)
}

// TestEventBuildingWithUnicodeCharacters_ShouldHandleCorrectly tests Unicode handling
func TestEventBuildingWithUnicodeCharacters_ShouldHandleCorrectly(t *testing.T) {
	unicodeMessage := "Message with émojis 🚀 and spëcial chärs"
	
	result, err := BuildSQSEventE("https://sqs.us-east-1.amazonaws.com/123456789012/test", unicodeMessage)
	require.NoError(t, err)
	
	var event SQSEvent
	err = json.Unmarshal([]byte(result), &event)
	require.NoError(t, err)
	
	assert.Equal(t, unicodeMessage, event.Records[0].Body)
}