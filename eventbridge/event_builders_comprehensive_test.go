package eventbridge

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCustomEvent_BasicFunctionality tests the BuildCustomEvent function
func TestBuildCustomEvent_BasicFunctionality(t *testing.T) {
	// RED: Test basic custom event building
	source := "test.service"
	detailType := "Test Event"
	detail := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}

	event := BuildCustomEvent(source, detailType, detail)

	assert.Equal(t, source, event.Source)
	assert.Equal(t, detailType, event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Verify detail is valid JSON
	var parsedDetail map[string]interface{}
	err := json.Unmarshal([]byte(event.Detail), &parsedDetail)
	require.NoError(t, err)
	assert.Equal(t, "value", parsedDetail["key"])
	assert.Equal(t, float64(42), parsedDetail["count"]) // JSON numbers are floats
}

func TestBuildCustomEvent_InvalidDetail(t *testing.T) {
	// RED: Test custom event with invalid detail (should handle gracefully)
	source := "test.service"
	detailType := "Test Event"
	detail := map[string]interface{}{
		"invalid": make(chan int), // Channels cannot be marshaled to JSON
	}

	event := BuildCustomEvent(source, detailType, detail)

	// Should fall back to empty JSON object
	assert.Equal(t, "{}", event.Detail)
}

func TestBuildScheduledEvent_BasicFunctionality(t *testing.T) {
	// RED: Test scheduled event building
	detailType := "Scheduled Task"
	detail := map[string]interface{}{
		"taskName": "cleanup",
		"interval": "daily",
	}

	event := BuildScheduledEvent(detailType, detail)

	assert.Equal(t, "aws.events", event.Source)
	assert.Equal(t, detailType, event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Verify detail content
	var parsedDetail map[string]interface{}
	err := json.Unmarshal([]byte(event.Detail), &parsedDetail)
	require.NoError(t, err)
	assert.Equal(t, "cleanup", parsedDetail["taskName"])
	assert.Equal(t, "daily", parsedDetail["interval"])
}

func TestBuildS3Event_Comprehensive(t *testing.T) {
	// RED: Test S3 event building with various parameters
	testCases := []struct {
		name       string
		bucketName string
		objectKey  string
		eventName  string
	}{
		{
			name:       "Basic S3 event",
			bucketName: "test-bucket",
			objectKey:  "test-file.txt",
			eventName:  "s3:ObjectCreated:Put",
		},
		{
			name:       "S3 event with path-like key",
			bucketName: "my-bucket",
			objectKey:  "folder/subfolder/file.json",
			eventName:  "s3:ObjectRemoved:Delete",
		},
		{
			name:       "S3 event with special characters",
			bucketName: "bucket-with-dashes",
			objectKey:  "file with spaces.txt",
			eventName:  "s3:ObjectCreated:CompleteMultipartUpload",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := BuildS3Event(tc.bucketName, tc.objectKey, tc.eventName)

			assert.Equal(t, "aws.s3", event.Source)
			assert.Contains(t, event.DetailType, "S3")
			assert.Contains(t, event.DetailType, tc.eventName)
			assert.Equal(t, DefaultEventBusName, event.EventBusName)

			// Verify resource ARN
			expectedArn := "arn:aws:s3:::" + tc.bucketName + "/" + tc.objectKey
			assert.Contains(t, event.Resources, expectedArn)

			// Verify detail structure
			var detail S3EventDetail
			err := json.Unmarshal([]byte(event.Detail), &detail)
			require.NoError(t, err)
			assert.Equal(t, "2.1", detail.EventVersion)
			assert.Equal(t, "aws:s3", detail.EventSource)
			assert.Equal(t, tc.eventName, detail.EventName)
			assert.Equal(t, tc.bucketName, detail.S3.Bucket.Name)
			assert.Equal(t, tc.objectKey, detail.S3.Object.Key)
		})
	}
}

func TestBuildEC2Event_Comprehensive(t *testing.T) {
	// RED: Test EC2 event building with various states
	testCases := []struct {
		name          string
		instanceID    string
		state         string
		previousState string
	}{
		{
			name:          "Instance starting",
			instanceID:    "i-1234567890abcdef0",
			state:         "running",
			previousState: "pending",
		},
		{
			name:          "Instance stopping",
			instanceID:    "i-abcdef1234567890",
			state:         "stopped",
			previousState: "running",
		},
		{
			name:          "Instance terminating",
			instanceID:    "i-fedcba0987654321",
			state:         "terminated",
			previousState: "stopping",
		},
		{
			name:          "Initial state (no previous)",
			instanceID:    "i-123456789abcdef0",
			state:         "pending",
			previousState: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := BuildEC2Event(tc.instanceID, tc.state, tc.previousState)

			assert.Equal(t, "aws.ec2", event.Source)
			assert.Equal(t, "EC2 Instance State-change Notification", event.DetailType)
			assert.Equal(t, DefaultEventBusName, event.EventBusName)

			// Verify resource ARN
			expectedArn := "arn:aws:ec2:us-east-1:123456789012:instance/" + tc.instanceID
			assert.Contains(t, event.Resources, expectedArn)

			// Verify detail structure
			var detail EC2EventDetail
			err := json.Unmarshal([]byte(event.Detail), &detail)
			require.NoError(t, err)
			assert.Equal(t, tc.instanceID, detail.InstanceID)
			assert.Equal(t, tc.state, detail.State)
			if tc.previousState != "" {
				assert.Equal(t, tc.previousState, detail.PreviousState)
			}
		})
	}
}

func TestBuildLambdaEvent_Comprehensive(t *testing.T) {
	// RED: Test Lambda event building with various states
	testCases := []struct {
		name         string
		functionName string
		functionArn  string
		state        string
		stateReason  string
	}{
		{
			name:         "Function active",
			functionName: "my-function",
			functionArn:  "arn:aws:lambda:us-east-1:123456789012:function:my-function",
			state:        "Active",
			stateReason:  "Function created successfully",
		},
		{
			name:         "Function failed",
			functionName: "failing-function",
			functionArn:  "arn:aws:lambda:us-east-1:123456789012:function:failing-function",
			state:        "Failed",
			stateReason:  "Function creation failed due to invalid code",
		},
		{
			name:         "Function pending",
			functionName: "new-function",
			functionArn:  "arn:aws:lambda:us-east-1:123456789012:function:new-function",
			state:        "Pending",
			stateReason:  "Function creation in progress",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := BuildLambdaEvent(tc.functionName, tc.functionArn, tc.state, tc.stateReason)

			assert.Equal(t, "aws.lambda", event.Source)
			assert.Equal(t, "Lambda Function State Change", event.DetailType)
			assert.Equal(t, DefaultEventBusName, event.EventBusName)

			// Verify resource ARN
			assert.Contains(t, event.Resources, tc.functionArn)

			// Verify detail structure
			var detail LambdaEventDetail
			err := json.Unmarshal([]byte(event.Detail), &detail)
			require.NoError(t, err)
			assert.Equal(t, tc.functionName, detail.FunctionName)
			assert.Equal(t, tc.functionArn, detail.FunctionArn)
			assert.Equal(t, tc.state, detail.State)
			assert.Equal(t, tc.stateReason, detail.StateReason)
			assert.Equal(t, "python3.9", detail.Runtime) // Default runtime
		})
	}
}

func TestBuildDynamoDBEvent_Comprehensive(t *testing.T) {
	// RED: Test DynamoDB event building
	testCases := []struct {
		name      string
		tableName string
		eventName string
		record    map[string]interface{}
	}{
		{
			name:      "Insert event",
			tableName: "Users",
			eventName: "INSERT",
			record: map[string]interface{}{
				"Keys": map[string]interface{}{
					"userId": map[string]interface{}{
						"S": "user123",
					},
				},
				"NewImage": map[string]interface{}{
					"userId": map[string]interface{}{"S": "user123"},
					"name":   map[string]interface{}{"S": "John Doe"},
				},
			},
		},
		{
			name:      "Remove event",
			tableName: "Products",
			eventName: "REMOVE",
			record: map[string]interface{}{
				"Keys": map[string]interface{}{
					"productId": map[string]interface{}{
						"S": "prod456",
					},
				},
				"OldImage": map[string]interface{}{
					"productId": map[string]interface{}{"S": "prod456"},
					"name":      map[string]interface{}{"S": "Old Product"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := BuildDynamoDBEvent(tc.tableName, tc.eventName, tc.record)

			assert.Equal(t, "aws.dynamodb", event.Source)
			assert.Equal(t, "DynamoDB Stream Record", event.DetailType)
			assert.Equal(t, DefaultEventBusName, event.EventBusName)

			// Verify resource ARN
			expectedArn := "arn:aws:dynamodb:us-east-1:123456789012:table/" + tc.tableName
			assert.Contains(t, event.Resources, expectedArn)

			// Verify detail structure
			var detail DynamoDBEventDetail
			err := json.Unmarshal([]byte(event.Detail), &detail)
			require.NoError(t, err)
			assert.Equal(t, tc.tableName, detail.TableName)
			assert.Equal(t, tc.eventName, detail.EventName)
			assert.Equal(t, "aws:dynamodb", detail.EventSource)
			assert.Equal(t, expectedArn, detail.EventSourceArn)
			assert.Equal(t, tc.record, detail.DynamoDBRecord)
		})
	}
}

func TestBuildScheduledEventWithCron_Comprehensive(t *testing.T) {
	// RED: Test scheduled event building with cron expressions
	testCases := []struct {
		name           string
		cronExpression string
		input          map[string]interface{}
	}{
		{
			name:           "Daily at midnight",
			cronExpression: "0 0 * * *",
			input:          map[string]interface{}{"task": "daily_cleanup"},
		},
		{
			name:           "Weekly on Monday",
			cronExpression: "0 9 * * MON",
			input:          map[string]interface{}{"task": "weekly_report"},
		},
		{
			name:           "Every 15 minutes",
			cronExpression: "*/15 * * * *",
			input:          map[string]interface{}{"task": "health_check"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rule := BuildScheduledEventWithCron(tc.cronExpression, tc.input)

			assert.Contains(t, rule.Name, "scheduled-rule-")
			assert.Contains(t, rule.Description, tc.cronExpression)
			assert.Equal(t, "cron("+tc.cronExpression+")", rule.ScheduleExpression)
			assert.Equal(t, RuleStateEnabled, rule.State)
			assert.Equal(t, DefaultEventBusName, rule.EventBusName)
		})
	}
}

func TestBuildScheduledEventWithRate_Comprehensive(t *testing.T) {
	// RED: Test scheduled event building with rate expressions
	testCases := []struct {
		name      string
		rateValue int
		rateUnit  string
	}{
		{
			name:      "Every 5 minutes",
			rateValue: 5,
			rateUnit:  "minutes",
		},
		{
			name:      "Every hour",
			rateValue: 1,
			rateUnit:  "hour",
		},
		{
			name:      "Every 3 days",
			rateValue: 3,
			rateUnit:  "days",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rule := BuildScheduledEventWithRate(tc.rateValue, tc.rateUnit)

			assert.Contains(t, rule.Name, "rate-rule-")
			assert.Contains(t, rule.Description, tc.rateUnit)
			expectedSchedule := "rate(" + string(rune('0'+tc.rateValue)) + " " + tc.rateUnit + ")"
			// For values > 9, we need a different approach
			if tc.rateValue <= 9 {
				assert.Equal(t, expectedSchedule, rule.ScheduleExpression)
			} else {
				assert.Contains(t, rule.ScheduleExpression, tc.rateUnit)
			}
			assert.Equal(t, RuleStateEnabled, rule.State)
			assert.Equal(t, DefaultEventBusName, rule.EventBusName)
		})
	}
}

func TestValidateEventPattern_EventPublishingFocus(t *testing.T) {
	// RED: Test comprehensive event pattern validation
	testCases := []struct {
		name    string
		pattern string
		valid   bool
	}{
		{
			name:    "Valid simple pattern",
			pattern: `{"source": ["aws.s3"]}`,
			valid:   true,
		},
		{
			name:    "Valid complex pattern",
			pattern: `{"source": ["aws.ec2"], "detail-type": ["EC2 Instance State-change Notification"], "detail": {"state": ["running", "stopped"]}}`,
			valid:   true,
		},
		{
			name:    "Valid pattern with multiple sources",
			pattern: `{"source": ["aws.s3", "aws.dynamodb"], "region": ["us-east-1", "us-west-2"]}`,
			valid:   true,
		},
		{
			name:    "Invalid - empty pattern",
			pattern: "",
			valid:   false,
		},
		{
			name:    "Invalid - malformed JSON",
			pattern: `{"source": aws.s3}`,
			valid:   false,
		},
		{
			name:    "Invalid - source not array",
			pattern: `{"source": "aws.s3"}`,
			valid:   false,
		},
		{
			name:    "Invalid - unknown field",
			pattern: `{"unknown-field": ["value"]}`,
			valid:   false,
		},
		{
			name:    "Invalid - detail-type not array",
			pattern: `{"detail-type": "EC2 Instance State-change Notification"}`,
			valid:   false,
		},
		{
			name:    "Valid pattern with detail object",
			pattern: `{"detail": {"state": ["running"]}}`,
			valid:   true,
		},
		{
			name:    "Invalid - detail not object",
			pattern: `{"detail": ["invalid"]}`,
			valid:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateEventPattern(tc.pattern)
			assert.Equal(t, tc.valid, result, "Pattern validation result mismatch for: %s", tc.pattern)
		})
	}
}

func TestEventBuilders_TimeHandling(t *testing.T) {
	// RED: Test time handling in event builders
	t.Run("S3 event time format", func(t *testing.T) {
		event := BuildS3Event("test-bucket", "test-key", "s3:ObjectCreated:Put")
		
		var detail S3EventDetail
		err := json.Unmarshal([]byte(event.Detail), &detail)
		require.NoError(t, err)
		
		// Verify time format is RFC3339
		eventTime, err := time.Parse(time.RFC3339, detail.EventTime)
		require.NoError(t, err)
		
		// Time should be recent (within last minute)
		assert.WithinDuration(t, time.Now(), eventTime, time.Minute)
	})
	
	t.Run("CloudWatch event time format", func(t *testing.T) {
		event := BuildCloudWatchEvent("test-alarm", "ALARM", "Threshold crossed", "CPUUtilization")
		
		var detail map[string]interface{}
		err := json.Unmarshal([]byte(event.Detail), &detail)
		require.NoError(t, err)
		
		// Verify state timestamp
		state := detail["state"].(map[string]interface{})
		timestamp := state["timestamp"].(string)
		
		eventTime, err := time.Parse(time.RFC3339, timestamp)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), eventTime, time.Minute)
		
		// Verify previous state timestamp
		previousState := detail["previousState"].(map[string]interface{})
		prevTimestamp := previousState["timestamp"].(string)
		
		prevTime, err := time.Parse(time.RFC3339, prevTimestamp)
		require.NoError(t, err)
		assert.True(t, prevTime.Before(eventTime), "Previous state should be before current state")
	})
}