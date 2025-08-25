package lambda

import (
	"encoding/json"
	"fmt"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple validation tests for the event builders to ensure they work
// These focus only on basic functionality without dependencies on other files

func TestEventBuildersBasicValidation(t *testing.T) {
	t.Run("BuildS3Event_BasicValidation", func(t *testing.T) {
		result := BuildS3Event("test-bucket", "test-key.txt", "")
		
		var event S3Event
		err := json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		assert.Len(t, event.Records, 1)
		assert.Equal(t, "test-bucket", event.Records[0].S3.Bucket.Name)
		assert.Equal(t, "test-key.txt", event.Records[0].S3.Object.Key)
	})

	t.Run("BuildDynamoDBEvent_BasicValidation", func(t *testing.T) {
		keys := map[string]interface{}{
			"id": map[string]interface{}{"S": "test-id"},
		}
		result := BuildDynamoDBEvent("test-table", "INSERT", keys)
		
		var event DynamoDBEvent
		err := json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		assert.Len(t, event.Records, 1)
		assert.Equal(t, "INSERT", event.Records[0].EventName)
	})

	t.Run("BuildSQSEvent_BasicValidation", func(t *testing.T) {
		result := BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", "test message")
		
		var event SQSEvent
		err := json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		assert.Len(t, event.Records, 1)
		assert.Equal(t, "test message", event.Records[0].Body)
	})

	t.Run("BuildSNSEvent_BasicValidation", func(t *testing.T) {
		result := BuildSNSEvent("arn:aws:sns:us-east-1:123456789012:test-topic", "test message")
		
		var event SNSEvent
		err := json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		assert.Len(t, event.Records, 1)
		assert.Equal(t, "test message", event.Records[0].Sns.Message)
	})

	t.Run("BuildAPIGatewayEvent_BasicValidation", func(t *testing.T) {
		result := BuildAPIGatewayEvent("POST", "/api/test", `{"test": true}`)
		
		var event APIGatewayEvent
		err := json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		assert.Equal(t, "POST", event.HttpMethod)
		assert.Equal(t, "/api/test", event.Path)
		assert.Equal(t, `{"test": true}`, event.Body)
	})

	t.Run("BuildCloudWatchEvent_BasicValidation", func(t *testing.T) {
		detail := map[string]interface{}{
			"state": "ALARM",
			"description": "Test alarm",
		}
		result := BuildCloudWatchEvent("aws.cloudwatch", "CloudWatch Alarm State Change", detail)
		
		var event CloudWatchEvent
		err := json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		assert.Equal(t, "aws.cloudwatch", event.Source)
		assert.Equal(t, "CloudWatch Alarm State Change", event.DetailType)
		assert.NotNil(t, event.Detail)
	})

	t.Run("BuildKinesisEvent_BasicValidation", func(t *testing.T) {
		result := BuildKinesisEvent("test-stream", "partition-key-1", "test data payload")
		
		var event KinesisEvent
		err := json.Unmarshal([]byte(result), &event)
		require.NoError(t, err)
		assert.Len(t, event.Records, 1)
		assert.Equal(t, "partition-key-1", event.Records[0].Kinesis.PartitionKey)
		assert.Equal(t, "aws:kinesis", event.Records[0].EventSource)
		assert.Contains(t, event.Records[0].EventSourceARN, "test-stream")
	})
}

func TestEventBuildersErrorHandling(t *testing.T) {
	t.Run("BuildS3EventE_EmptyBucket_ReturnsError", func(t *testing.T) {
		result, err := BuildS3EventE("", "key", "")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "bucket name cannot be empty")
	})

	t.Run("BuildS3EventE_EmptyKey_ReturnsError", func(t *testing.T) {
		result, err := BuildS3EventE("bucket", "", "")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "object key cannot be empty")
	})

	t.Run("BuildCloudWatchEventE_EmptySource_ReturnsError", func(t *testing.T) {
		result, err := BuildCloudWatchEventE("", "detail-type", nil)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "source cannot be empty")
	})

	t.Run("BuildKinesisEventE_EmptyStreamName_ReturnsError", func(t *testing.T) {
		result, err := BuildKinesisEventE("", "partition-key", "data")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "stream name cannot be empty")
	})

	t.Run("BuildKinesisEventE_EmptyPartitionKey_ReturnsError", func(t *testing.T) {
		result, err := BuildKinesisEventE("stream", "", "data")
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "partition key cannot be empty")
	})
}

func TestEventBuildersDefaultValues(t *testing.T) {
	t.Run("BuildS3Event_DefaultEventName", func(t *testing.T) {
		result := BuildS3Event("bucket", "key", "")
		var event S3Event
		json.Unmarshal([]byte(result), &event)
		assert.Equal(t, "s3:ObjectCreated:Put", event.Records[0].EventName)
	})

	t.Run("BuildDynamoDBEvent_DefaultEventName", func(t *testing.T) {
		result := BuildDynamoDBEvent("table", "", nil)
		var event DynamoDBEvent
		json.Unmarshal([]byte(result), &event)
		assert.Equal(t, "INSERT", event.Records[0].EventName)
	})

	t.Run("BuildSQSEvent_DefaultMessageBody", func(t *testing.T) {
		result := BuildSQSEvent("queue-url", "")
		var event SQSEvent
		json.Unmarshal([]byte(result), &event)
		assert.Equal(t, "Hello from SQS!", event.Records[0].Body)
	})

	t.Run("BuildSNSEvent_DefaultMessage", func(t *testing.T) {
		result := BuildSNSEvent("topic-arn", "")
		var event SNSEvent
		json.Unmarshal([]byte(result), &event)
		assert.Equal(t, "Hello from SNS!", event.Records[0].Sns.Message)
	})

	t.Run("BuildKinesisEvent_DefaultData", func(t *testing.T) {
		result := BuildKinesisEvent("stream", "key", "")
		var event KinesisEvent
		json.Unmarshal([]byte(result), &event)
		// Should have hex encoded "Hello from Kinesis!"
		assert.NotEmpty(t, event.Records[0].Kinesis.Data)
	})

	t.Run("BuildCloudWatchEvent_DefaultDetailType", func(t *testing.T) {
		result := BuildCloudWatchEvent("aws.logs", "", nil)
		var event CloudWatchEvent
		json.Unmarshal([]byte(result), &event)
		assert.Equal(t, "CloudWatch Event", event.DetailType)
	})
}

func TestEventBuildersJSONValidity(t *testing.T) {
	builders := []func() string{
		func() string { return BuildS3Event("bucket", "key", "") },
		func() string { return BuildDynamoDBEvent("table", "", nil) },
		func() string { return BuildSQSEvent("queue-url", "") },
		func() string { return BuildSNSEvent("topic-arn", "") },
		func() string { return BuildAPIGatewayEvent("", "", "") },
		func() string { return BuildCloudWatchEvent("aws.logs", "", nil) },
		func() string { return BuildKinesisEvent("stream", "key", "") },
	}

	for i, builder := range builders {
		t.Run(fmt.Sprintf("EventBuilder_%d_ProducesValidJSON", i), func(t *testing.T) {
			result := builder()
			
			var jsonObj map[string]interface{}
			err := json.Unmarshal([]byte(result), &jsonObj)
			assert.NoError(t, err, "All event builders should produce valid JSON")
			assert.NotEmpty(t, result, "All event builders should produce non-empty results")
		})
	}
}