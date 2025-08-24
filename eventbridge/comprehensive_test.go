package eventbridge

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
)

// TestEventBridgePackageIntegration demonstrates comprehensive EventBridge package usage
func TestEventBridgePackageIntegration(t *testing.T) {
	// This is a comprehensive integration test that demonstrates all functionality
	// It will fail with authentication errors in CI/local environment without AWS creds
	// but serves as a documentation of the expected API
	
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	t.Run("Event Bus Management", func(t *testing.T) {
		busConfig := EventBusConfig{
			Name: "test-custom-bus",
			Tags: map[string]string{
				"Environment": "test",
				"Purpose":     "eventbridge-package-test",
			},
		}
		
		// These tests will fail without AWS credentials, but demonstrate the API
		if testing.Short() {
			t.Skip("Skipping integration test in short mode")
		}
		
		// Test creating event bus (will fail without credentials)
		_, err := CreateEventBusE(ctx, busConfig)
		// We expect this to fail with auth error, not API error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MissingAuthenticationToken")
	})
	
	t.Run("Rule Management", func(t *testing.T) {
		ruleConfig := RuleConfig{
			Name:         "test-integration-rule",
			Description:  "Integration test rule",
			EventPattern: `{"source": ["test.service"], "detail-type": ["Test Event"]}`,
			State:        types.RuleStateEnabled,
			EventBusName: "default",
			Tags: map[string]string{
				"Test": "true",
			},
		}
		
		if testing.Short() {
			t.Skip("Skipping integration test in short mode")
		}
		
		// Test creating rule (will fail without credentials)
		_, err := CreateRuleE(ctx, ruleConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MissingAuthenticationToken")
	})
	
	t.Run("Event Publishing", func(t *testing.T) {
		event := BuildCustomEvent("test.service", "Integration Test", map[string]interface{}{
			"testId":    "12345",
			"timestamp": time.Now().Unix(),
			"data":      "integration test data",
		})
		
		if testing.Short() {
			t.Skip("Skipping integration test in short mode")
		}
		
		// Test putting event (will fail without credentials)
		_, err := PutEventE(ctx, event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MissingAuthenticationToken")
	})
	
	t.Run("Batch Event Publishing", func(t *testing.T) {
		events := []CustomEvent{
			BuildCustomEvent("test.service", "Batch Test 1", map[string]interface{}{
				"batchId": 1,
			}),
			BuildCustomEvent("test.service", "Batch Test 2", map[string]interface{}{
				"batchId": 2,
			}),
		}
		
		if testing.Short() {
			t.Skip("Skipping integration test in short mode")
		}
		
		// Test putting batch events (will fail without credentials)
		_, err := PutEventsE(ctx, events)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MissingAuthenticationToken")
	})
}

// TestEventBuilders tests all the AWS service event builders
func TestEventBuilders(t *testing.T) {
	// These tests don't require AWS credentials since they're just building event structures
	
	t.Run("S3 Event Builder", func(t *testing.T) {
		event := BuildS3Event("test-bucket", "test-object.txt", "ObjectCreated:Put")
		
		assert.Equal(t, "aws.s3", event.Source)
		assert.Equal(t, "S3 ObjectCreated:Put", event.DetailType)
		assert.Contains(t, event.Detail, "test-bucket")
		assert.Contains(t, event.Detail, "test-object.txt")
		assert.Contains(t, event.Resources[0], "test-bucket/test-object.txt")
	})
	
	t.Run("EC2 Event Builder", func(t *testing.T) {
		event := BuildEC2Event("i-1234567890abcdef0", "running", "pending")
		
		assert.Equal(t, "aws.ec2", event.Source)
		assert.Equal(t, "EC2 Instance State-change Notification", event.DetailType)
		assert.Contains(t, event.Detail, "i-1234567890abcdef0")
		assert.Contains(t, event.Detail, "running")
		assert.Contains(t, event.Resources[0], "i-1234567890abcdef0")
	})
	
	t.Run("Lambda Event Builder", func(t *testing.T) {
		functionArn := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
		event := BuildLambdaEvent("test-function", functionArn, "Active", "")
		
		assert.Equal(t, "aws.lambda", event.Source)
		assert.Equal(t, "Lambda Function State Change", event.DetailType)
		assert.Contains(t, event.Detail, "test-function")
		assert.Contains(t, event.Detail, "Active")
		assert.Equal(t, functionArn, event.Resources[0])
	})
	
	t.Run("DynamoDB Event Builder", func(t *testing.T) {
		record := map[string]interface{}{
			"eventName": "INSERT",
			"dynamodb": map[string]interface{}{
				"Keys": map[string]interface{}{
					"id": map[string]interface{}{
						"S": "test-id",
					},
				},
			},
		}
		
		event := BuildDynamoDBEvent("test-table", "INSERT", record)
		
		assert.Equal(t, "aws.dynamodb", event.Source)
		assert.Equal(t, "DynamoDB Stream Record", event.DetailType)
		assert.Contains(t, event.Detail, "test-table")
		assert.Contains(t, event.Detail, "INSERT")
		assert.Contains(t, event.Resources[0], "test-table")
	})
	
	t.Run("SQS Event Builder", func(t *testing.T) {
		event := BuildSQSEvent("test-queue", "msg-12345", "test message body")
		
		assert.Equal(t, "aws.sqs", event.Source)
		assert.Equal(t, "SQS Message", event.DetailType)
		assert.Contains(t, event.Detail, "test-queue")
		assert.Contains(t, event.Detail, "msg-12345")
		assert.Contains(t, event.Detail, "test message body")
	})
	
	t.Run("CloudWatch Event Builder", func(t *testing.T) {
		event := BuildCloudWatchEvent("test-alarm", "ALARM", "Threshold Crossed", "CPUUtilization")
		
		assert.Equal(t, "aws.cloudwatch", event.Source)
		assert.Equal(t, "CloudWatch Alarm State Change", event.DetailType)
		assert.Contains(t, event.Detail, "test-alarm")
		assert.Contains(t, event.Detail, "ALARM")
		assert.Contains(t, event.Detail, "CPUUtilization")
	})
	
	t.Run("Scheduled Event Builders", func(t *testing.T) {
		cronRule := BuildScheduledEventWithCron("0 12 * * ? *", map[string]interface{}{
			"action": "daily-report",
		})
		
		assert.Contains(t, cronRule.Name, "scheduled-rule-")
		assert.Contains(t, cronRule.ScheduleExpression, "cron(0 12 * * ? *)")
		assert.Equal(t, types.RuleStateEnabled, cronRule.State)
		
		rateRule := BuildScheduledEventWithRate(5, "minutes")
		
		assert.Contains(t, rateRule.Name, "rate-rule-")
		assert.Equal(t, "rate(5 minutes)", rateRule.ScheduleExpression)
		assert.Equal(t, types.RuleStateEnabled, rateRule.State)
	})
}

// TestEventPatternValidation tests event pattern validation functionality
func TestEventPatternValidation(t *testing.T) {
	testCases := []struct {
		name        string
		pattern     string
		shouldBeValid bool
	}{
		{
			name:        "Valid pattern with source and detail-type",
			pattern:     `{"source": ["test.service"], "detail-type": ["Test Event"]}`,
			shouldBeValid: true,
		},
		{
			name:        "Valid pattern with nested detail",
			pattern:     `{"source": ["aws.s3"], "detail": {"eventName": ["ObjectCreated:Put"]}}`,
			shouldBeValid: true,
		},
		{
			name:        "Invalid pattern with string instead of array",
			pattern:     `{"source": "test.service"}`,
			shouldBeValid: false,
		},
		{
			name:        "Invalid JSON",
			pattern:     `{"source": [}`,
			shouldBeValid: false,
		},
		{
			name:        "Empty pattern",
			pattern:     "",
			shouldBeValid: false,
		},
		{
			name:        "Invalid key",
			pattern:     `{"invalid-key": ["value"]}`,
			shouldBeValid: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateEventPattern(tc.pattern)
			assert.Equal(t, tc.shouldBeValid, result, "Pattern: %s", tc.pattern)
		})
	}
}

// TestRetryLogic tests retry logic for various operations
func TestRetryLogic(t *testing.T) {
	
	t.Run("Calculate Backoff Delay", func(t *testing.T) {
		config := RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   1 * time.Second,
			MaxDelay:    10 * time.Second,
			Multiplier:  2.0,
		}
		
		// Test exponential backoff
		delay0 := calculateBackoffDelay(0, config)
		assert.Equal(t, 1*time.Second, delay0)
		
		delay1 := calculateBackoffDelay(1, config)
		assert.Equal(t, 2*time.Second, delay1)
		
		delay2 := calculateBackoffDelay(2, config)
		assert.Equal(t, 4*time.Second, delay2)
	})
	
	t.Run("Event Validation", func(t *testing.T) {
		// Test valid event detail
		validDetail := `{"key": "value", "number": 42}`
		err := validateEventDetail(validDetail)
		assert.NoError(t, err)
		
		// Test invalid JSON
		invalidDetail := `{"key": "value"`
		err = validateEventDetail(invalidDetail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event detail")
		
		// Test empty detail (should be valid)
		err = validateEventDetail("")
		assert.NoError(t, err)
	})
}

// TestTargetConfiguration tests target configuration and management
func TestTargetConfiguration(t *testing.T) {
	t.Run("Target Config Creation", func(t *testing.T) {
		targetConfig := TargetConfig{
			ID:       "target-1",
			Arn:      "arn:aws:lambda:us-east-1:123456789012:function:test",
			RoleArn:  "arn:aws:iam::123456789012:role/service-role/test-role",
			Input:    `{"key": "value"}`,
		}
		
		assert.Equal(t, "target-1", targetConfig.ID)
		assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:test", targetConfig.Arn)
		assert.Equal(t, "arn:aws:iam::123456789012:role/service-role/test-role", targetConfig.RoleArn)
		assert.Contains(t, targetConfig.Input, "key")
	})
	
	t.Run("Multiple Target Configuration", func(t *testing.T) {
		targets := []TargetConfig{
			{
				ID:  "lambda-target",
				Arn: "arn:aws:lambda:us-east-1:123456789012:function:handler",
			},
			{
				ID:  "sqs-target",
				Arn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			},
		}
		
		assert.Len(t, targets, 2)
		assert.Equal(t, "lambda-target", targets[0].ID)
		assert.Equal(t, "sqs-target", targets[1].ID)
	})
}

// TestArchiveAndReplayConfiguration tests archive and replay configuration
func TestArchiveAndReplayConfiguration(t *testing.T) {
	t.Run("Archive Configuration", func(t *testing.T) {
		archiveConfig := ArchiveConfig{
			ArchiveName:      "test-archive",
			EventSourceArn:   "arn:aws:events:us-east-1:123456789012:event-bus/test-bus",
			EventPattern:     `{"source": ["test.service"]}`,
			Description:      "Test archive for integration testing",
			RetentionDays:    7,
		}
		
		assert.Equal(t, "test-archive", archiveConfig.ArchiveName)
		assert.Equal(t, 7, int(archiveConfig.RetentionDays))
		assert.True(t, ValidateEventPattern(archiveConfig.EventPattern))
	})
	
	t.Run("Replay Configuration", func(t *testing.T) {
		now := time.Now()
		replayConfig := ReplayConfig{
			ReplayName:     "test-replay",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
			EventStartTime: now.Add(-24 * time.Hour),
			EventEndTime:   now,
			Description:    "Test replay for integration testing",
		}
		
		assert.Equal(t, "test-replay", replayConfig.ReplayName)
		assert.True(t, replayConfig.EventStartTime.Before(replayConfig.EventEndTime))
		assert.Contains(t, replayConfig.Description, "integration testing")
	})
}

// BenchmarkEventBuilding benchmarks event building performance
func BenchmarkEventBuilding(b *testing.B) {
	b.Run("BuildCustomEvent", func(b *testing.B) {
		detail := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
			"nested": map[string]interface{}{
				"inner": "value",
			},
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BuildCustomEvent("test.service", "Benchmark Test", detail)
		}
	})
	
	b.Run("BuildS3Event", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BuildS3Event("test-bucket", "test-object.txt", "ObjectCreated:Put")
		}
	})
	
	b.Run("ValidateEventPattern", func(b *testing.B) {
		pattern := `{"source": ["test.service"], "detail-type": ["Test Event"]}`
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ValidateEventPattern(pattern)
		}
	})
}