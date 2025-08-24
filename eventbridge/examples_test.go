package eventbridge

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
)

// TestExampleEventBridgeWorkflow demonstrates a complete EventBridge workflow using Terratest patterns
func TestExampleEventBridgeWorkflow(t *testing.T) {
	// This example shows how to use the EventBridge package in a real testing scenario
	// following Terratest patterns with proper error handling and cleanup
	
	// Initialize test context (would normally be done in test setup)
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	if testing.Short() {
		t.Skip("Skipping example test in short mode")
	}
	
	// Step 1: Create a custom event bus
	busConfig := EventBusConfig{
		Name: "my-test-bus",
		Tags: map[string]string{
			"Environment": "test",
			"Team":        "platform",
		},
	}
	
	// This will panic on error (Terratest pattern)
	eventBus := CreateEventBus(ctx, busConfig)
	fmt.Printf("Created event bus: %s\n", eventBus.Arn)
	
	// Step 2: Create a rule with event pattern
	ruleConfig := RuleConfig{
		Name:         "user-signup-rule",
		Description:  "Rule for user signup events",
		EventPattern: `{"source": ["user.service"], "detail-type": ["User Signup"]}`,
		State:        RuleStateEnabled,
		EventBusName: "my-test-bus",
	}
	
	rule := CreateRule(ctx, ruleConfig)
	fmt.Printf("Created rule: %s\n", rule.RuleArn)
	
	// Step 3: Add targets to the rule
	targets := []TargetConfig{
		{
			ID:      "lambda-processor",
			Arn:     "arn:aws:lambda:us-east-1:123456789012:function:process-signup",
			RoleArn: "arn:aws:iam::123456789012:role/EventBridgeExecutionRole",
		},
		{
			ID:  "sqs-queue",
			Arn: "arn:aws:sqs:us-east-1:123456789012:signup-queue",
		},
	}
	
	PutTargets(ctx, rule.Name, "my-test-bus", targets)
	fmt.Println("Added targets to rule")
	
	// Step 4: Send test events
	signupEvent := BuildCustomEvent("user.service", "User Signup", map[string]interface{}{
		"userId":    "user-12345",
		"email":     "test@example.com",
		"timestamp": time.Now().Unix(),
	})
	
	result := PutEvent(ctx, signupEvent)
	fmt.Printf("Sent event with ID: %s\n", result.EventID)
	
	// Step 5: Verify rule is working with assertions
	AssertRuleExists(ctx, "user-signup-rule", "my-test-bus")
	AssertRuleEnabled(ctx, "user-signup-rule", "my-test-bus")
	AssertTargetCount(ctx, "user-signup-rule", "my-test-bus", 2)
	
	// Step 6: Cleanup (normally in defer or test teardown)
	DeleteRule(ctx, "user-signup-rule", "my-test-bus")
	DeleteEventBus(ctx, "my-test-bus")
	
	fmt.Println("Cleanup completed")
}

// TestExampleArchiveAndReplay demonstrates event archiving and replay functionality
func TestExampleArchiveAndReplay(t *testing.T) {
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	if testing.Short() {
		t.Skip("Skipping example test in short mode")
	}
	
	// Create event bus for archiving
	busConfig := EventBusConfig{
		Name: "audit-event-bus",
	}
	
	eventBus := CreateEventBus(ctx, busConfig)
	
	// Create archive for the event bus
	archiveConfig := ArchiveConfig{
		ArchiveName:    "audit-archive",
		EventSourceArn: eventBus.Arn,
		EventPattern:   `{"source": ["audit.service"]}`,
		Description:    "Archive for audit events",
		RetentionDays:  30,
	}
	
	archive := CreateArchive(ctx, archiveConfig)
	fmt.Printf("Created archive: %s\n", archive.ArchiveArn)
	
	// Send some events that will be archived
	auditEvents := []CustomEvent{
		BuildCustomEvent("audit.service", "User Login", map[string]interface{}{
			"userId":    "user-001",
			"timestamp": time.Now().Add(-2*time.Hour).Unix(),
		}),
		BuildCustomEvent("audit.service", "Data Access", map[string]interface{}{
			"userId":     "user-001",
			"resourceId": "resource-123",
			"timestamp":  time.Now().Add(-1*time.Hour).Unix(),
		}),
	}
	
	batchResult := PutEvents(ctx, auditEvents)
	AssertEventBatchSent(ctx, batchResult)
	
	// Wait for archive to be ready (in real scenario)
	WaitForArchiveState(ctx, "audit-archive", ArchiveStateEnabled, 5*time.Minute)
	
	// Create replay for the last 3 hours
	replayConfig := ReplayConfig{
		ReplayName:     "audit-replay",
		EventSourceArn: archive.ArchiveArn,
		EventStartTime: time.Now().Add(-3 * time.Hour),
		EventEndTime:   time.Now(),
		Destination: types.ReplayDestination{
			Arn: &eventBus.Arn,
		},
		Description: "Replay audit events for analysis",
	}
	
	replay := StartReplay(ctx, replayConfig)
	fmt.Printf("Started replay: %s\n", replay.ReplayArn)
	
	// Wait for replay to complete
	WaitForReplayCompletion(ctx, "audit-replay", 10*time.Minute)
	AssertReplayCompleted(ctx, "audit-replay")
	
	fmt.Println("Archive and replay workflow completed")
}

// TestExampleBatchEventProcessing shows how to handle batch event processing
func TestExampleBatchEventProcessing(t *testing.T) {
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	if testing.Short() {
		t.Skip("Skipping example test in short mode")
	}
	
	// Create multiple events from different AWS services
	events := []CustomEvent{
		BuildS3Event("data-bucket", "uploads/file1.csv", "ObjectCreated:Put"),
		BuildS3Event("data-bucket", "uploads/file2.json", "ObjectCreated:Put"),
		BuildEC2Event("i-1234567890abcdef0", "running", "pending"),
		BuildLambdaEvent("data-processor", "arn:aws:lambda:us-east-1:123456789012:function:data-processor", "Active", ""),
		BuildDynamoDBEvent("users", "INSERT", map[string]interface{}{
			"eventName": "INSERT",
			"dynamodb": map[string]interface{}{
				"Keys": map[string]interface{}{
					"userId": map[string]interface{}{"S": "new-user-123"},
				},
			},
		}),
	}
	
	// Send batch of events
	batchResult := PutEvents(ctx, events)
	
	// Verify all events were sent successfully
	AssertEventBatchSent(ctx, batchResult)
	
	// Check individual results
	for i, result := range batchResult.Entries {
		if result.Success {
			fmt.Printf("Event %d sent successfully with ID: %s\n", i, result.EventID)
		} else {
			fmt.Printf("Event %d failed: %s - %s\n", i, result.ErrorCode, result.ErrorMessage)
		}
	}
	
	fmt.Printf("Batch processing completed. Failed entries: %d\n", batchResult.FailedEntryCount)
}

// TestExampleScheduledEvents demonstrates creating scheduled events
func TestExampleScheduledEvents(t *testing.T) {
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	if testing.Short() {
		t.Skip("Skipping example test in short mode")
	}
	
	// Create a rule that triggers every day at 9 AM UTC
	dailyReportRule := BuildScheduledEventWithCron("0 9 * * ? *", map[string]interface{}{
		"reportType": "daily",
	})
	dailyReportRule.Name = "daily-report-trigger"
	dailyReportRule.Description = "Triggers daily report generation"
	
	rule1 := CreateRule(ctx, dailyReportRule)
	fmt.Printf("Created daily scheduled rule: %s\n", rule1.RuleArn)
	
	// Create a rule that triggers every 5 minutes
	healthCheckRule := BuildScheduledEventWithRate(5, "minutes")
	healthCheckRule.Name = "health-check-trigger"
	healthCheckRule.Description = "Triggers health checks every 5 minutes"
	
	rule2 := CreateRule(ctx, healthCheckRule)
	fmt.Printf("Created rate scheduled rule: %s\n", rule2.RuleArn)
	
	// Add Lambda targets to both rules
	dailyTarget := []TargetConfig{{
		ID:  "daily-report-lambda",
		Arn: "arn:aws:lambda:us-east-1:123456789012:function:generate-daily-report",
		Input: `{"reportType": "daily", "format": "pdf"}`,
	}}
	
	healthCheckTarget := []TargetConfig{{
		ID:  "health-check-lambda",
		Arn: "arn:aws:lambda:us-east-1:123456789012:function:health-check",
	}}
	
	PutTargets(ctx, rule1.Name, "default", dailyTarget)
	PutTargets(ctx, rule2.Name, "default", healthCheckTarget)
	
	// Verify rules are properly configured
	AssertRuleEnabled(ctx, rule1.Name, "default")
	AssertRuleEnabled(ctx, rule2.Name, "default")
	AssertTargetExists(ctx, rule1.Name, "default", "daily-report-lambda")
	AssertTargetExists(ctx, rule2.Name, "default", "health-check-lambda")
	
	fmt.Println("Scheduled events setup completed")
}

// TestExampleCrossAccountEventBridge demonstrates cross-account EventBridge setup
func TestExampleCrossAccountEventBridge(t *testing.T) {
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	if testing.Short() {
		t.Skip("Skipping example test in short mode")
	}
	
	// Create custom event bus for cross-account events
	crossAccountBus := EventBusConfig{
		Name: "cross-account-bus",
	}
	
	eventBus := CreateEventBus(ctx, crossAccountBus)
	
	// Grant permission to another account to put events
	PutPermission(ctx, "cross-account-bus", "AllowAccount123456789012", "123456789012", "events:PutEvents")
	
	// Create rule to process cross-account events
	crossAccountRule := RuleConfig{
		Name:         "cross-account-rule",
		Description:  "Process events from partner account",
		EventPattern: `{"source": ["partner.service"], "account": ["123456789012"]}`,
		State:        RuleStateEnabled,
		EventBusName: "cross-account-bus",
	}
	
	rule := CreateRule(ctx, crossAccountRule)
	
	// Add target to process cross-account events
	targets := []TargetConfig{{
		ID:      "cross-account-processor",
		Arn:     "arn:aws:lambda:us-east-1:123456789012:function:process-partner-events",
		RoleArn: "arn:aws:iam::123456789012:role/CrossAccountEventRole",
	}}
	
	PutTargets(ctx, rule.Name, "cross-account-bus", targets)
	
	// Verify setup
	AssertEventBusExists(ctx, "cross-account-bus")
	AssertRuleExists(ctx, "cross-account-rule", "cross-account-bus")
	AssertTargetExists(ctx, "cross-account-rule", "cross-account-bus", "cross-account-processor")
	
	fmt.Printf("Cross-account EventBridge setup completed for bus: %s\n", eventBus.Arn)
}

// TestExampleEventPatternTesting demonstrates event pattern testing and validation
func TestExampleEventPatternTesting(t *testing.T) {
	
	// Test various event patterns
	patterns := []string{
		`{"source": ["user.service"], "detail-type": ["User Created"]}`,
		`{"source": ["aws.s3"], "detail": {"eventName": ["ObjectCreated:Put"]}}`,
		`{"source": ["order.service"], "detail": {"status": ["completed"], "amount": [{"numeric": [">", 100]}]}}`,
	}
	
	// Create test events
	testEvents := []CustomEvent{
		BuildCustomEvent("user.service", "User Created", map[string]interface{}{
			"userId": "user-123",
			"email":  "test@example.com",
		}),
		BuildS3Event("test-bucket", "test-file.txt", "ObjectCreated:Put"),
		BuildCustomEvent("order.service", "Order Completed", map[string]interface{}{
			"orderId": "order-456",
			"status":  "completed",
			"amount":  150.00,
		}),
	}
	
	// Test pattern matching
	for i, pattern := range patterns {
		fmt.Printf("Testing pattern %d: %s\n", i+1, pattern)
		
		// Validate pattern format
		isValid := ValidateEventPattern(pattern)
		assert.True(t, isValid, "Pattern should be valid")
		
		// Test if event matches pattern (would use AWS API in real scenario)
		if i < len(testEvents) {
			// This would call AssertPatternMatches in a real test
			fmt.Printf("Testing event match for pattern %d\n", i+1)
		}
	}
	
	fmt.Println("Event pattern testing completed")
}

// TestExampleEventBridgeCleanup demonstrates proper cleanup of EventBridge resources
func TestExampleEventBridgeCleanup(t *testing.T) {
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	if testing.Short() {
		t.Skip("Skipping example test in short mode")
	}
	
	// List all custom event buses for cleanup
	eventBuses, err := ListEventBusesE(ctx)
	if err != nil {
		fmt.Printf("Error listing event buses: %v\n", err)
		return
	}
	
	for _, bus := range eventBuses {
		// Skip default bus
		if bus.Name == "default" {
			continue
		}
		
		fmt.Printf("Cleaning up event bus: %s\n", bus.Name)
		
		// List and delete all rules in the bus
		rules, err := ListRulesE(ctx, bus.Name)
		if err != nil {
			fmt.Printf("Error listing rules for bus %s: %v\n", bus.Name, err)
			continue
		}
		
		for _, rule := range rules {
			fmt.Printf("Deleting rule: %s\n", rule.Name)
			DeleteRule(ctx, rule.Name, bus.Name)
		}
		
		// Delete the event bus
		fmt.Printf("Deleting event bus: %s\n", bus.Name)
		DeleteEventBus(ctx, bus.Name)
	}
	
	// Clean up archives
	archives, err := ListArchivesE(ctx, "")
	if err != nil {
		fmt.Printf("Error listing archives: %v\n", err)
		return
	}
	
	for _, archive := range archives {
		fmt.Printf("Deleting archive: %s\n", archive.ArchiveName)
		DeleteArchive(ctx, archive.ArchiveName)
	}
	
	fmt.Println("EventBridge cleanup completed")
}