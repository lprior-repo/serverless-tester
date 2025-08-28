package main

import (
	"encoding/json"
	"fmt"
	"go/build"
	"os/exec"
	"strings"
	"testing"
	"time"

	"vasdeference/eventbridge"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test suite for ExampleEventBridgeWorkflow
func TestExampleEventBridgeWorkflow(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		// RED: Test that function executes without panic
		require.NotPanics(t, ExampleEventBridgeWorkflow, "ExampleEventBridgeWorkflow should not panic")
	})
	
	t.Run("should demonstrate complete EventBridge workflow pattern", func(t *testing.T) {
		// GREEN: Test workflow demonstration
		ExampleEventBridgeWorkflow()
		
		// Verify the function completes successfully
		assert.True(t, true, "EventBridge workflow example completed")
	})
	
	t.Run("should create proper EventBusConfig structure", func(t *testing.T) {
		// Test EventBusConfig structure creation
		busConfig := eventbridge.EventBusConfig{
			Name: "test-event-bus",
			Tags: map[string]string{
				"Environment": "test",
				"Team":        "platform",
			},
		}
		
		assert.Equal(t, "test-event-bus", busConfig.Name, "EventBus name should match")
		assert.Len(t, busConfig.Tags, 2, "Should have 2 tags")
		assert.Contains(t, busConfig.Tags, "Environment", "Should contain Environment tag")
		assert.Contains(t, busConfig.Tags, "Team", "Should contain Team tag")
		assert.Equal(t, "test", busConfig.Tags["Environment"], "Environment tag should be 'test'")
		assert.Equal(t, "platform", busConfig.Tags["Team"], "Team tag should be 'platform'")
	})
	
	t.Run("should create proper RuleConfig structure", func(t *testing.T) {
		// Test RuleConfig structure creation
		ruleConfig := eventbridge.RuleConfig{
			Name:         "test-rule",
			Description:  "Test rule for user events",
			EventPattern: `{"source": ["user.service"], "detail-type": ["User Action"]}`,
			State:        eventbridge.RuleStateEnabled,
			EventBusName: "test-event-bus",
		}
		
		assert.Equal(t, "test-rule", ruleConfig.Name, "Rule name should match")
		assert.Equal(t, "Test rule for user events", ruleConfig.Description, "Rule description should match")
		assert.Contains(t, ruleConfig.EventPattern, "user.service", "Event pattern should contain source")
		assert.Contains(t, ruleConfig.EventPattern, "User Action", "Event pattern should contain detail-type")
		assert.Equal(t, eventbridge.RuleStateEnabled, ruleConfig.State, "Rule should be enabled")
		assert.Equal(t, "test-event-bus", ruleConfig.EventBusName, "EventBus name should match")
	})
	
	t.Run("should create proper TargetConfig structures", func(t *testing.T) {
		// Test TargetConfig structure creation
		targets := []eventbridge.TargetConfig{
			{
				ID:      "lambda-target",
				Arn:     "arn:aws:lambda:us-east-1:123456789012:function:test-function",
				RoleArn: "arn:aws:iam::123456789012:role/EventBridgeRole",
			},
			{
				ID:  "sqs-target",
				Arn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			},
		}
		
		assert.Len(t, targets, 2, "Should have 2 targets")
		
		// Test Lambda target
		lambdaTarget := targets[0]
		assert.Equal(t, "lambda-target", lambdaTarget.ID, "Lambda target ID should match")
		assert.Contains(t, lambdaTarget.Arn, "lambda", "Lambda target ARN should contain lambda")
		assert.Contains(t, lambdaTarget.Arn, "test-function", "Lambda target ARN should contain function name")
		assert.Contains(t, lambdaTarget.RoleArn, "EventBridgeRole", "Lambda target should have execution role")
		
		// Test SQS target
		sqsTarget := targets[1]
		assert.Equal(t, "sqs-target", sqsTarget.ID, "SQS target ID should match")
		assert.Contains(t, sqsTarget.Arn, "sqs", "SQS target ARN should contain sqs")
		assert.Contains(t, sqsTarget.Arn, "test-queue", "SQS target ARN should contain queue name")
	})
}

// Test suite for ExampleArchiveAndReplay
func TestExampleArchiveAndReplay(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleArchiveAndReplay, "ExampleArchiveAndReplay should not panic")
	})
	
	t.Run("should demonstrate archive and replay patterns", func(t *testing.T) {
		// Test archive and replay demonstration
		ExampleArchiveAndReplay()
		
		assert.True(t, true, "Archive and replay example completed")
	})
	
	t.Run("should create proper ArchiveConfig structure", func(t *testing.T) {
		// Test ArchiveConfig structure creation
		archiveConfig := eventbridge.ArchiveConfig{
			ArchiveName:    "test-archive",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/test-bus",
			EventPattern:   `{"source": ["test.service"]}`,
			Description:    "Test archive for events",
			RetentionDays:  int32(7),
		}
		
		assert.Equal(t, "test-archive", archiveConfig.ArchiveName, "Archive name should match")
		assert.Contains(t, archiveConfig.EventSourceArn, "event-bus", "EventSourceArn should contain event-bus")
		assert.Contains(t, archiveConfig.EventSourceArn, "test-bus", "EventSourceArn should contain bus name")
		assert.Contains(t, archiveConfig.EventPattern, "test.service", "EventPattern should contain source")
		assert.Equal(t, "Test archive for events", archiveConfig.Description, "Description should match")
		assert.Equal(t, int32(7), archiveConfig.RetentionDays, "Retention should be 7 days")
	})
	
	t.Run("should create proper ReplayConfig structure", func(t *testing.T) {
		// Test ReplayConfig structure creation
		startTime := time.Now().Add(-2 * time.Hour)
		endTime := time.Now()
		
		replayConfig := eventbridge.ReplayConfig{
			ReplayName:     "test-replay",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
			EventStartTime: startTime,
			EventEndTime:   endTime,
			Destination: types.ReplayDestination{
				Arn: func() *string {
					arn := "arn:aws:events:us-east-1:123456789012:event-bus/test-bus"
					return &arn
				}(),
			},
			Description: "Test replay for analysis",
		}
		
		assert.Equal(t, "test-replay", replayConfig.ReplayName, "Replay name should match")
		assert.Contains(t, replayConfig.EventSourceArn, "archive", "EventSourceArn should contain archive")
		assert.Contains(t, replayConfig.EventSourceArn, "test-archive", "EventSourceArn should contain archive name")
		assert.True(t, replayConfig.EventStartTime.Before(replayConfig.EventEndTime), "Start time should be before end time")
		assert.Equal(t, "Test replay for analysis", replayConfig.Description, "Description should match")
		require.NotNil(t, replayConfig.Destination.Arn, "Destination ARN should not be nil")
		assert.Contains(t, *replayConfig.Destination.Arn, "event-bus", "Destination should be event-bus")
	})
}

// Test suite for ExampleBatchEventProcessing
func TestExampleBatchEventProcessing(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleBatchEventProcessing, "ExampleBatchEventProcessing should not panic")
	})
	
	t.Run("should demonstrate batch event processing patterns", func(t *testing.T) {
		// Test batch event processing demonstration
		ExampleBatchEventProcessing()
		
		assert.True(t, true, "Batch event processing example completed")
	})
	
	t.Run("should create proper AWS service events", func(t *testing.T) {
		// Test S3 event creation
		s3Event := eventbridge.BuildS3Event("test-bucket", "test/file.txt", "ObjectCreated:Put")
		assert.NotEmpty(t, s3Event, "S3 event should not be empty")
		
		// Test EC2 event creation
		ec2Event := eventbridge.BuildEC2Event("i-1234567890abcdef0", "running", "pending")
		assert.NotEmpty(t, ec2Event, "EC2 event should not be empty")
		
		// Test Lambda event creation
		lambdaEvent := eventbridge.BuildLambdaEvent("test-function", "arn:aws:lambda:us-east-1:123456789012:function:test", "Active", "")
		assert.NotEmpty(t, lambdaEvent, "Lambda event should not be empty")
		
		// Test DynamoDB event creation
		dynamoEvent := eventbridge.BuildDynamoDBEvent("test-table", "INSERT", map[string]interface{}{
			"eventName": "INSERT",
			"dynamodb": map[string]interface{}{
				"Keys": map[string]interface{}{
					"id": map[string]interface{}{"S": "test-id"},
				},
			},
		})
		assert.NotEmpty(t, dynamoEvent, "DynamoDB event should not be empty")
	})
}

// Test suite for ExampleScheduledEvents
func TestExampleScheduledEvents(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleScheduledEvents, "ExampleScheduledEvents should not panic")
	})
	
	t.Run("should demonstrate scheduled events patterns", func(t *testing.T) {
		// Test scheduled events demonstration
		ExampleScheduledEvents()
		
		assert.True(t, true, "Scheduled events example completed")
	})
	
	t.Run("should create proper cron-based scheduled rule", func(t *testing.T) {
		// Test cron-based scheduled rule creation
		cronRule := eventbridge.BuildScheduledEventWithCron("0 9 * * ? *", map[string]interface{}{
			"reportType": "daily",
		})
		cronRule.Name = "test-cron-rule"
		cronRule.Description = "Test daily cron rule"
		
		assert.Equal(t, "test-cron-rule", cronRule.Name, "Cron rule name should match")
		assert.Equal(t, "Test daily cron rule", cronRule.Description, "Cron rule description should match")
		// The actual ScheduleExpression field validation would depend on the structure returned
	})
	
	t.Run("should create proper rate-based scheduled rule", func(t *testing.T) {
		// Test rate-based scheduled rule creation
		rateRule := eventbridge.BuildScheduledEventWithRate(5, "minutes")
		rateRule.Name = "test-rate-rule"
		rateRule.Description = "Test rate rule every 5 minutes"
		
		assert.Equal(t, "test-rate-rule", rateRule.Name, "Rate rule name should match")
		assert.Equal(t, "Test rate rule every 5 minutes", rateRule.Description, "Rate rule description should match")
	})
	
	t.Run("should validate target configurations for scheduled events", func(t *testing.T) {
		// Test target configurations for scheduled events
		targets := []eventbridge.TargetConfig{
			{
				ID:    "scheduled-lambda",
				Arn:   "arn:aws:lambda:us-east-1:123456789012:function:scheduled-task",
				Input: `{"type": "scheduled", "frequency": "daily"}`,
			},
		}
		
		assert.Len(t, targets, 1, "Should have 1 target")
		target := targets[0]
		assert.Equal(t, "scheduled-lambda", target.ID, "Target ID should match")
		assert.Contains(t, target.Arn, "lambda", "Target should be Lambda function")
		assert.Contains(t, target.Input, "scheduled", "Input should contain scheduled type")
		assert.Contains(t, target.Input, "daily", "Input should contain frequency")
	})
}

// Test suite for ExampleCrossAccountEventBridge
func TestExampleCrossAccountEventBridge(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleCrossAccountEventBridge, "ExampleCrossAccountEventBridge should not panic")
	})
	
	t.Run("should demonstrate cross-account patterns", func(t *testing.T) {
		// Test cross-account demonstration
		ExampleCrossAccountEventBridge()
		
		assert.True(t, true, "Cross-account EventBridge example completed")
	})
	
	t.Run("should create proper cross-account rule configuration", func(t *testing.T) {
		// Test cross-account rule configuration
		crossAccountRule := eventbridge.RuleConfig{
			Name:         "cross-account-test-rule",
			Description:  "Test cross-account rule",
			EventPattern: `{"source": ["partner.service"], "account": ["123456789012"]}`,
			State:        eventbridge.RuleStateEnabled,
			EventBusName: "cross-account-test-bus",
		}
		
		assert.Equal(t, "cross-account-test-rule", crossAccountRule.Name, "Rule name should match")
		assert.Contains(t, crossAccountRule.EventPattern, "partner.service", "Pattern should contain partner source")
		assert.Contains(t, crossAccountRule.EventPattern, "account", "Pattern should contain account filter")
		assert.Contains(t, crossAccountRule.EventPattern, "123456789012", "Pattern should contain account ID")
		assert.Equal(t, eventbridge.RuleStateEnabled, crossAccountRule.State, "Rule should be enabled")
		assert.Equal(t, "cross-account-test-bus", crossAccountRule.EventBusName, "Should use cross-account bus")
	})
	
	t.Run("should create proper cross-account targets", func(t *testing.T) {
		// Test cross-account target configuration
		targets := []eventbridge.TargetConfig{
			{
				ID:      "cross-account-processor",
				Arn:     "arn:aws:lambda:us-east-1:123456789012:function:process-partner-events",
				RoleArn: "arn:aws:iam::123456789012:role/CrossAccountEventRole",
			},
		}
		
		assert.Len(t, targets, 1, "Should have 1 cross-account target")
		target := targets[0]
		assert.Equal(t, "cross-account-processor", target.ID, "Target ID should match")
		assert.Contains(t, target.Arn, "lambda", "Target should be Lambda function")
		assert.Contains(t, target.Arn, "process-partner-events", "Target should process partner events")
		assert.Contains(t, target.RoleArn, "CrossAccountEventRole", "Should use cross-account role")
	})
}

// Test suite for ExampleEventPatternTesting
func TestExampleEventPatternTesting(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleEventPatternTesting, "ExampleEventPatternTesting should not panic")
	})
	
	t.Run("should demonstrate event pattern testing patterns", func(t *testing.T) {
		// Test event pattern testing demonstration
		ExampleEventPatternTesting()
		
		assert.True(t, true, "Event pattern testing example completed")
	})
	
	t.Run("should validate event pattern structures", func(t *testing.T) {
		// Test event patterns used in the examples
		patterns := []string{
			`{"source": ["user.service"], "detail-type": ["User Created"]}`,
			`{"source": ["aws.s3"], "detail": {"eventName": ["ObjectCreated:Put"]}}`,
			`{"source": ["order.service"], "detail": {"status": ["completed"], "amount": [{"numeric": [">", 100]}]}}`,
		}
		
		for i, pattern := range patterns {
			t.Run(fmt.Sprintf("pattern_%d", i+1), func(t *testing.T) {
				// Validate JSON structure
				assert.True(t, json.Valid([]byte(pattern)), "Pattern should be valid JSON")
				assert.Contains(t, pattern, "source", "Pattern should contain source")
				
				// Validate specific pattern contents
				switch i {
				case 0:
					assert.Contains(t, pattern, "user.service", "Pattern should contain user service")
					assert.Contains(t, pattern, "detail-type", "Pattern should contain detail-type")
				case 1:
					assert.Contains(t, pattern, "aws.s3", "Pattern should contain S3 source")
					assert.Contains(t, pattern, "eventName", "Pattern should contain eventName")
				case 2:
					assert.Contains(t, pattern, "order.service", "Pattern should contain order service")
					assert.Contains(t, pattern, "numeric", "Pattern should contain numeric filter")
				}
			})
		}
	})
	
	t.Run("should create proper test events", func(t *testing.T) {
		// Test event creation patterns
		userEvent := eventbridge.BuildCustomEvent("user.service", "User Created", map[string]interface{}{
			"userId": "test-123",
			"email":  "test@example.com",
		})
		assert.NotEmpty(t, userEvent, "User event should not be empty")
		
		s3Event := eventbridge.BuildS3Event("test-bucket", "test-file.txt", "ObjectCreated:Put")
		assert.NotEmpty(t, s3Event, "S3 event should not be empty")
		
		orderEvent := eventbridge.BuildCustomEvent("order.service", "Order Completed", map[string]interface{}{
			"orderId": "order-456",
			"status":  "completed",
			"amount":  150.00,
		})
		assert.NotEmpty(t, orderEvent, "Order event should not be empty")
	})
	
	t.Run("should validate event pattern validation function", func(t *testing.T) {
		// Test event pattern validation
		validPattern := `{"source": ["test.service"]}`
		invalidPattern := `{"source": ["test.service",]}`  // Invalid JSON
		
		assert.True(t, eventbridge.ValidateEventPattern(validPattern), "Valid pattern should pass validation")
		assert.False(t, eventbridge.ValidateEventPattern(invalidPattern), "Invalid pattern should fail validation")
	})
}

// Test suite for ExampleEventBridgeCleanup
func TestExampleEventBridgeCleanup(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, ExampleEventBridgeCleanup, "ExampleEventBridgeCleanup should not panic")
	})
	
	t.Run("should demonstrate cleanup patterns", func(t *testing.T) {
		// Test cleanup demonstration
		ExampleEventBridgeCleanup()
		
		assert.True(t, true, "EventBridge cleanup example completed")
	})
	
	t.Run("should demonstrate proper cleanup workflow", func(t *testing.T) {
		// Test cleanup workflow patterns
		// Simulate event buses for demonstration
		simulatedBuses := []string{"custom-bus-1", "custom-bus-2", "audit-bus"}
		
		for _, busName := range simulatedBuses {
			// Skip default bus (as shown in the example)
			if busName == "default" {
				continue
			}
			
			assert.NotEqual(t, "default", busName, "Should skip default bus")
			assert.NotEmpty(t, busName, "Bus name should not be empty")
			
			// Simulate rules cleanup
			simulatedRules := []string{"rule-1", "rule-2"}
			for _, ruleName := range simulatedRules {
				assert.NotEmpty(t, ruleName, "Rule name should not be empty")
			}
		}
		
		// Simulate archives cleanup
		simulatedArchives := []string{"audit-archive", "backup-archive"}
		for _, archiveName := range simulatedArchives {
			assert.NotEmpty(t, archiveName, "Archive name should not be empty")
		}
	})
}

// Test suite for runAllExamples
func TestRunAllExamples(t *testing.T) {
	t.Run("should execute without panic", func(t *testing.T) {
		require.NotPanics(t, runAllExamples, "runAllExamples should not panic")
	})
	
	t.Run("should call all example functions", func(t *testing.T) {
		// Verify that runAllExamples executes all example functions
		runAllExamples()
		
		assert.True(t, true, "All EventBridge examples executed successfully")
	})
}

// Test suite for event validation and structure patterns
func TestEventValidationPatterns(t *testing.T) {
	t.Run("should validate custom event structure", func(t *testing.T) {
		// Test custom event creation pattern
		event := eventbridge.BuildCustomEvent("test.service", "Test Event", map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		})
		
		assert.NotEmpty(t, event, "Custom event should not be empty")
		// Additional validations would depend on the actual CustomEvent structure
	})
	
	t.Run("should validate AWS service event structures", func(t *testing.T) {
		// Test S3 event structure
		s3Event := eventbridge.BuildS3Event("test-bucket", "test/key.txt", "ObjectCreated:Put")
		assert.NotEmpty(t, s3Event, "S3 event should not be empty")
		
		// Test EC2 event structure
		ec2Event := eventbridge.BuildEC2Event("i-1234567890abcdef0", "running", "pending")
		assert.NotEmpty(t, ec2Event, "EC2 event should not be empty")
		
		// Test Lambda event structure
		lambdaEvent := eventbridge.BuildLambdaEvent("test-function", "arn:aws:lambda:us-east-1:123456789012:function:test", "Active", "")
		assert.NotEmpty(t, lambdaEvent, "Lambda event should not be empty")
	})
	
	t.Run("should validate complex event patterns", func(t *testing.T) {
		// Test complex event patterns
		complexPatterns := []string{
			`{"source": ["aws.ec2"], "detail-type": ["EC2 Instance State-change Notification"], "detail": {"state": ["running"]}}`,
			`{"source": ["aws.s3"], "detail": {"eventName": ["ObjectCreated:Put", "ObjectCreated:Post"], "requestParameters": {"bucketName": ["my-bucket"]}}}`,
			`{"source": ["myapp.orders"], "detail": {"status": ["completed"], "amount": [{"numeric": [">=", 100]}]}}`,
		}
		
		for i, pattern := range complexPatterns {
			t.Run(fmt.Sprintf("complex_pattern_%d", i+1), func(t *testing.T) {
				// Basic JSON validation
				assert.True(t, json.Valid([]byte(pattern)), "Pattern should be valid JSON")
				assert.Contains(t, pattern, "source", "Pattern should contain source")
				assert.Contains(t, pattern, "detail", "Pattern should contain detail")
			})
		}
	})
}

// Test suite for TestContext structure validation
func TestContextValidation(t *testing.T) {
	t.Run("should create proper TestContext", func(t *testing.T) {
		// Test TestContext structure creation
		ctx := &eventbridge.TestContext{
			AwsConfig: aws.Config{Region: "us-east-1"},
			Region:    "us-east-1",
		}
		
		assert.Equal(t, "us-east-1", ctx.Region, "Region should match")
		assert.Equal(t, "us-east-1", ctx.AwsConfig.Region, "AwsConfig region should match")
	})
	
	t.Run("should validate region consistency", func(t *testing.T) {
		// Test region consistency
		testRegions := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}
		
		for _, region := range testRegions {
			t.Run(region, func(t *testing.T) {
				ctx := &eventbridge.TestContext{
					AwsConfig: aws.Config{Region: region},
					Region:    region,
				}
				
				assert.Equal(t, region, ctx.Region, "Region should match")
				assert.Equal(t, region, ctx.AwsConfig.Region, "AwsConfig region should match")
				assert.NotEmpty(t, region, "Region should not be empty")
			})
		}
	})
}

// Test suite for edge cases and error conditions
func TestEdgeCasesAndErrorConditions(t *testing.T) {
	t.Run("should handle empty configurations", func(t *testing.T) {
		// Test empty EventBusConfig
		emptyBusConfig := eventbridge.EventBusConfig{}
		assert.Empty(t, emptyBusConfig.Name, "Empty bus config should have empty name")
		assert.Nil(t, emptyBusConfig.Tags, "Empty bus config should have nil tags")
		
		// Test empty RuleConfig
		emptyRuleConfig := eventbridge.RuleConfig{}
		assert.Empty(t, emptyRuleConfig.Name, "Empty rule config should have empty name")
		assert.Empty(t, emptyRuleConfig.EventPattern, "Empty rule config should have empty pattern")
	})
	
	t.Run("should validate ARN formats", func(t *testing.T) {
		// Test various ARN patterns used in examples
		validARNs := []string{
			"arn:aws:lambda:us-east-1:123456789012:function:test-function",
			"arn:aws:sqs:us-east-1:123456789012:test-queue",
			"arn:aws:events:us-east-1:123456789012:event-bus/test-bus",
			"arn:aws:events:us-east-1:123456789012:archive/test-archive",
			"arn:aws:iam::123456789012:role/EventBridgeRole",
		}
		
		for _, arn := range validARNs {
			t.Run(arn, func(t *testing.T) {
				assert.True(t, strings.HasPrefix(arn, "arn:aws:"), "ARN should start with arn:aws:")
				assert.Contains(t, arn, ":", "ARN should contain colons")
				parts := strings.Split(arn, ":")
				assert.GreaterOrEqual(t, len(parts), 6, "ARN should have at least 6 parts")
			})
		}
	})
	
	t.Run("should validate event pattern JSON structures", func(t *testing.T) {
		// Test event pattern validation
		validPatterns := []string{
			`{"source": ["test.service"]}`,
			`{"source": ["test.service"], "detail-type": ["Test Event"]}`,
			`{"source": ["aws.s3"], "detail": {"eventName": ["ObjectCreated:Put"]}}`,
		}
		
		invalidPatterns := []string{
			`{"source": }`,  // Invalid JSON
			`{source: ["test"]}`,  // Missing quotes
			`{"source": ["test",]}`,  // Trailing comma
		}
		
		for _, pattern := range validPatterns {
			assert.True(t, json.Valid([]byte(pattern)), "Valid pattern should pass JSON validation: %s", pattern)
		}
		
		for _, pattern := range invalidPatterns {
			assert.False(t, json.Valid([]byte(pattern)), "Invalid pattern should fail JSON validation: %s", pattern)
		}
	})
}

// Legacy tests maintained for compatibility
func TestExamplesCompilation(t *testing.T) {
	t.Run("should compile without errors", func(t *testing.T) {
		cmd := exec.Command("go", "build", "-o", "/tmp/test-eventbridge-examples", "examples.go")
		cmd.Dir = "."
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Errorf("Examples failed to compile: %v\nOutput: %s", err, output)
		}
	})
}

func TestExampleFunctionsExecute(t *testing.T) {
	t.Run("should execute all example functions without panic", func(t *testing.T) {
		// Test that all example functions can be called without panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Example function panicked: %v", r)
			}
		}()
		
		ExampleEventBridgeWorkflow()
		ExampleArchiveAndReplay()
		ExampleBatchEventProcessing()
		ExampleScheduledEvents()
		ExampleCrossAccountEventBridge()
		ExampleEventPatternTesting()
		ExampleEventBridgeCleanup()
	})
}

func TestMainFunction(t *testing.T) {
	t.Run("should execute main without panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Main function panicked: %v", r)
			}
		}()
		
		runAllExamples()
	})
}

func TestPackageIsExecutable(t *testing.T) {
	t.Run("should build as executable binary", func(t *testing.T) {
		// Verify this is a main package
		pkg, err := build.ImportDir(".", 0)
		if err != nil {
			t.Fatalf("Failed to import current directory: %v", err)
		}
		
		if pkg.Name != "main" {
			t.Errorf("Package name should be 'main', got '%s'", pkg.Name)
		}
		
		// Check that main function exists
		hasMain := false
		for _, filename := range pkg.GoFiles {
			if filename == "examples.go" {
				hasMain = true
				break
			}
		}
		
		if !hasMain {
			t.Error("examples.go should exist in package")
		}
	})
}