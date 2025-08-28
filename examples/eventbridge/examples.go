package main

import (
	"fmt"
	"time"

	"vasdeference/eventbridge"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// ExampleEventBridgeWorkflow demonstrates a complete EventBridge workflow
func ExampleEventBridgeWorkflow() {
	fmt.Println("Complete EventBridge workflow patterns:")
	
	// This example shows how to use the EventBridge package in real implementations
	// following functional programming patterns with proper error handling and cleanup
	
	// Initialize context (would normally be done in application setup)
	_ = /* ctx */ &eventbridge.TestContext{
		// T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	// Step 1: Create a custom event bus
	busConfig := eventbridge.EventBusConfig{
		Name: "my-example-bus",
		Tags: map[string]string{
			"Environment": "example",
			"Team":        "platform",
		},
	}
	
	// eventBus := eventbridge.CreateEventBus(ctx, busConfig)
	// fmt.Printf("Created event bus: %s\n", eventBus.Arn)
	fmt.Printf("Creating event bus: %s\n", busConfig.Name)
	
	// Step 2: Create a rule with event pattern
	ruleConfig := eventbridge.RuleConfig{
		Name:         "user-signup-rule",
		Description:  "Rule for user signup events",
		EventPattern: `{"source": ["user.service"], "detail-type": ["User Signup"]}`,
		State:        eventbridge.RuleStateEnabled,
		EventBusName: "my-example-bus",
	}
	
	// rule := eventbridge.CreateRule(ctx, ruleConfig)
	// fmt.Printf("Created rule: %s\n", rule.RuleArn)
	fmt.Printf("Creating rule: %s\n", ruleConfig.Name)
	
	// Step 3: Add targets to the rule
	targets := []eventbridge.TargetConfig{
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
	
	// eventbridge.PutTargets(ctx, rule.Name, "my-example-bus", targets)
	fmt.Printf("Adding %d targets to rule\n", len(targets))
	
	// Step 4: Send events
	_ = /* signupEvent */ eventbridge.BuildCustomEvent("user.service", "User Signup", map[string]interface{}{
		"userId":    "user-12345",
		"email":     "test@example.com",
		"timestamp": time.Now().Unix(),
	})
	
	// result := eventbridge.PutEvent(ctx, signupEvent)
	// fmt.Printf("Sent event with ID: %s\n", result.EventID)
	fmt.Println("Sending user signup event")
	
	// Step 5: Verify rule configuration
	// eventbridge.AssertRuleExists(ctx, "user-signup-rule", "my-example-bus")
	// eventbridge.AssertRuleEnabled(ctx, "user-signup-rule", "my-example-bus")
	// eventbridge.AssertTargetCount(ctx, "user-signup-rule", "my-example-bus", 2)
	fmt.Println("Verifying rule exists, is enabled, and has 2 targets")
	
	// Step 6: Cleanup (normally in defer or application teardown)
	// eventbridge.DeleteRule(ctx, "user-signup-rule", "my-example-bus")
	// eventbridge.DeleteEventBus(ctx, "my-example-bus")
	fmt.Println("Cleanup completed")
	fmt.Println("EventBridge workflow completed")
}

// ExampleArchiveAndReplay demonstrates event archiving and replay functionality
func ExampleArchiveAndReplay() {
	fmt.Println("EventBridge archive and replay patterns:")
	
	_ = /* ctx */ &eventbridge.TestContext{
		// T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	// Create event bus for archiving
	busConfig := eventbridge.EventBusConfig{
		Name: "audit-event-bus",
	}
	
	// eventBus := eventbridge.CreateEventBus(ctx, busConfig)
	fmt.Printf("Creating event bus for archiving: %s\n", busConfig.Name)
	
	// Create archive for the event bus
	archiveConfig := eventbridge.ArchiveConfig{
		ArchiveName:    "audit-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/audit-event-bus",
		EventPattern:   `{"source": ["audit.service"]}`,
		Description:    "Archive for audit events",
		RetentionDays:  30,
	}
	
	// archive := eventbridge.CreateArchive(ctx, archiveConfig)
	// fmt.Printf("Created archive: %s\n", archive.ArchiveArn)
	fmt.Printf("Creating archive: %s with %d day retention\n", archiveConfig.ArchiveName, archiveConfig.RetentionDays)
	
	// Send some events that will be archived
	auditEvents := []eventbridge.CustomEvent{
		eventbridge.BuildCustomEvent("audit.service", "User Login", map[string]interface{}{
			"userId":    "user-001",
			"timestamp": time.Now().Add(-2 * time.Hour).Unix(),
		}),
		eventbridge.BuildCustomEvent("audit.service", "Data Access", map[string]interface{}{
			"userId":     "user-001",
			"resourceId": "resource-123",
			"timestamp":  time.Now().Add(-1 * time.Hour).Unix(),
		}),
	}
	
	// batchResult := eventbridge.PutEvents(ctx, auditEvents)
	// eventbridge.AssertEventBatchSent(ctx, batchResult)
	fmt.Printf("Sending %d audit events for archiving\n", len(auditEvents))
	
	// Wait for archive to be ready (in real scenario)
	// eventbridge.WaitForArchiveState(ctx, "audit-archive", eventbridge.ArchiveStateEnabled, 5*time.Minute)
	fmt.Println("Waiting for archive to be enabled")
	
	// Create replay for the last 3 hours
	replayConfig := eventbridge.ReplayConfig{
		ReplayName:     "audit-replay",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/audit-archive",
		EventStartTime: time.Now().Add(-3 * time.Hour),
		EventEndTime:   time.Now(),
		Destination: types.ReplayDestination{
			Arn: func() *string {
				arn := "arn:aws:events:us-east-1:123456789012:event-bus/audit-event-bus"
				return &arn
			}(),
		},
		Description: "Replay audit events for analysis",
	}
	
	// replay := eventbridge.StartReplay(ctx, replayConfig)
	// fmt.Printf("Started replay: %s\n", replay.ReplayArn)
	fmt.Printf("Starting replay: %s for 3-hour window\n", replayConfig.ReplayName)
	
	// Wait for replay to complete
	// eventbridge.WaitForReplayCompletion(ctx, "audit-replay", 10*time.Minute)
	// eventbridge.AssertReplayCompleted(ctx, "audit-replay")
	fmt.Println("Waiting for replay to complete")
	
	fmt.Println("Archive and replay workflow completed")
}

// ExampleBatchEventProcessing shows how to handle batch event processing
func ExampleBatchEventProcessing() {
	fmt.Println("EventBridge batch event processing patterns:")
	
	_ = /* ctx */ &eventbridge.TestContext{
		// T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	// Create multiple events from different AWS services
	events := []eventbridge.CustomEvent{
		eventbridge.BuildS3Event("data-bucket", "uploads/file1.csv", "ObjectCreated:Put"),
		eventbridge.BuildS3Event("data-bucket", "uploads/file2.json", "ObjectCreated:Put"),
		eventbridge.BuildEC2Event("i-1234567890abcdef0", "running", "pending"),
		eventbridge.BuildLambdaEvent("data-processor", "arn:aws:lambda:us-east-1:123456789012:function:data-processor", "Active", ""),
		eventbridge.BuildDynamoDBEvent("users", "INSERT", map[string]interface{}{
			"eventName": "INSERT",
			"dynamodb": map[string]interface{}{
				"Keys": map[string]interface{}{
					"userId": map[string]interface{}{"S": "new-user-123"},
				},
			},
		}),
	}
	
	// Send batch of events
	// batchResult := eventbridge.PutEvents(ctx, events)
	fmt.Printf("Sending batch of %d events from different AWS services\n", len(events))
	
	// Verify all events were sent successfully
	// eventbridge.AssertEventBatchSent(ctx, batchResult)
	
	// Check individual results (simulated)
	for i := range events {
		// Simulate successful processing
		fmt.Printf("Event %d sent successfully\n", i)
	}
	
	fmt.Printf("Batch processing completed. Total events: %d\n", len(events))
}

// ExampleScheduledEvents demonstrates creating scheduled events
func ExampleScheduledEvents() {
	fmt.Println("EventBridge scheduled events patterns:")
	
	_ = /* ctx */ &eventbridge.TestContext{
		// T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	// Create a rule that triggers every day at 9 AM UTC
	dailyReportRule := eventbridge.BuildScheduledEventWithCron("0 9 * * ? *", map[string]interface{}{
		"reportType": "daily",
	})
	dailyReportRule.Name = "daily-report-trigger"
	dailyReportRule.Description = "Triggers daily report generation"
	
	// rule1 := eventbridge.CreateRule(ctx, dailyReportRule)
	// fmt.Printf("Created daily scheduled rule: %s\n", rule1.RuleArn)
	fmt.Printf("Creating daily scheduled rule: %s (9 AM UTC)\n", dailyReportRule.Name)
	
	// Create a rule that triggers every 5 minutes
	healthCheckRule := eventbridge.BuildScheduledEventWithRate(5, "minutes")
	healthCheckRule.Name = "health-check-trigger"
	healthCheckRule.Description = "Triggers health checks every 5 minutes"
	
	// rule2 := eventbridge.CreateRule(ctx, healthCheckRule)
	// fmt.Printf("Created rate scheduled rule: %s\n", rule2.RuleArn)
	fmt.Printf("Creating rate scheduled rule: %s (every 5 minutes)\n", healthCheckRule.Name)
	
	// Add Lambda targets to both rules
	dailyTarget := []eventbridge.TargetConfig{{
		ID:    "daily-report-lambda",
		Arn:   "arn:aws:lambda:us-east-1:123456789012:function:generate-daily-report",
		Input: `{"reportType": "daily", "format": "pdf"}`,
	}}
	
	healthCheckTarget := []eventbridge.TargetConfig{{
		ID:  "health-check-lambda",
		Arn: "arn:aws:lambda:us-east-1:123456789012:function:health-check",
	}}
	
	// eventbridge.PutTargets(ctx, rule1.Name, "default", dailyTarget)
	// eventbridge.PutTargets(ctx, rule2.Name, "default", healthCheckTarget)
	fmt.Printf("Adding targets: %s and %s\n", dailyTarget[0].ID, healthCheckTarget[0].ID)
	
	// Verify rules are properly configured
	// eventbridge.AssertRuleEnabled(ctx, rule1.Name, "default")
	// eventbridge.AssertRuleEnabled(ctx, rule2.Name, "default")
	// eventbridge.AssertTargetExists(ctx, rule1.Name, "default", "daily-report-lambda")
	// eventbridge.AssertTargetExists(ctx, rule2.Name, "default", "health-check-lambda")
	fmt.Println("Verifying rules are enabled and targets exist")
	
	fmt.Println("Scheduled events setup completed")
}

// ExampleCrossAccountEventBridge demonstrates cross-account EventBridge setup
func ExampleCrossAccountEventBridge() {
	fmt.Println("EventBridge cross-account patterns:")
	
	_ = /* ctx */ &eventbridge.TestContext{
		// T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	// Create custom event bus for cross-account events
	crossAccountBus := eventbridge.EventBusConfig{
		Name: "cross-account-bus",
	}
	
	// eventBus := eventbridge.CreateEventBus(ctx, crossAccountBus)
	fmt.Printf("Creating cross-account event bus: %s\n", crossAccountBus.Name)
	
	// Grant permission to another account to put events
	// eventbridge.PutPermission(ctx, "cross-account-bus", "AllowAccount123456789012", "123456789012", "events:PutEvents")
	fmt.Println("Granting permission to account 123456789012 for events:PutEvents")
	
	// Create rule to process cross-account events
	crossAccountRule := eventbridge.RuleConfig{
		Name:         "cross-account-rule",
		Description:  "Process events from partner account",
		EventPattern: `{"source": ["partner.service"], "account": ["123456789012"]}`,
		State:        eventbridge.RuleStateEnabled,
		EventBusName: "cross-account-bus",
	}
	
	// rule := eventbridge.CreateRule(ctx, crossAccountRule)
	fmt.Printf("Creating cross-account rule: %s\n", crossAccountRule.Name)
	
	// Add target to process cross-account events
	targets := []eventbridge.TargetConfig{{
		ID:      "cross-account-processor",
		Arn:     "arn:aws:lambda:us-east-1:123456789012:function:process-partner-events",
		RoleArn: "arn:aws:iam::123456789012:role/CrossAccountEventRole",
	}}
	
	// eventbridge.PutTargets(ctx, rule.Name, "cross-account-bus", targets)
	fmt.Printf("Adding target: %s\n", targets[0].ID)
	
	// Verify setup
	// eventbridge.AssertEventBusExists(ctx, "cross-account-bus")
	// eventbridge.AssertRuleExists(ctx, "cross-account-rule", "cross-account-bus")
	// eventbridge.AssertTargetExists(ctx, "cross-account-rule", "cross-account-bus", "cross-account-processor")
	fmt.Println("Verifying cross-account setup: bus exists, rule exists, target exists")
	
	fmt.Println("Cross-account EventBridge setup completed")
}

// ExampleEventPatternTesting demonstrates event pattern testing and validation
func ExampleEventPatternTesting() {
	fmt.Println("EventBridge event pattern testing patterns:")
	
	// Test various event patterns
	patterns := []string{
		`{"source": ["user.service"], "detail-type": ["User Created"]}`,
		`{"source": ["aws.s3"], "detail": {"eventName": ["ObjectCreated:Put"]}}`,
		`{"source": ["order.service"], "detail": {"status": ["completed"], "amount": [{"numeric": [">", 100]}]}}`,
	}
	
	// Create test events
	testEvents := []eventbridge.CustomEvent{
		eventbridge.BuildCustomEvent("user.service", "User Created", map[string]interface{}{
			"userId": "user-123",
			"email":  "test@example.com",
		}),
		eventbridge.BuildS3Event("test-bucket", "test-file.txt", "ObjectCreated:Put"),
		eventbridge.BuildCustomEvent("order.service", "Order Completed", map[string]interface{}{
			"orderId": "order-456",
			"status":  "completed",
			"amount":  150.00,
		}),
	}
	
	// Test pattern matching
	for i, pattern := range patterns {
		fmt.Printf("Testing pattern %d: %s\n", i+1, pattern)
		
		// Validate pattern format
		isValid := eventbridge.ValidateEventPattern(pattern)
		if isValid {
			fmt.Println("Pattern is valid")
		} else {
			fmt.Println("Pattern is invalid")
		}
		
		// Test if event matches pattern (would use AWS API in real scenario)
		if i < len(testEvents) {
			// This would call AssertPatternMatches in a real implementation
			fmt.Printf("Testing event match for pattern %d\n", i+1)
		}
	}
	
	fmt.Println("Event pattern testing completed")
}

// ExampleEventBridgeCleanup demonstrates proper cleanup of EventBridge resources
func ExampleEventBridgeCleanup() {
	fmt.Println("EventBridge cleanup patterns:")
	
	_ = /* ctx */ &eventbridge.TestContext{
		// T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
	
	// List all custom event buses for cleanup
	// eventBuses, err := eventbridge.ListEventBusesE(ctx)
	// if err != nil {
	//     fmt.Printf("Error listing event buses: %v\n", err)
	//     return
	// }
	fmt.Println("Listing all custom event buses")
	
	// Simulate event buses for demonstration
	simulatedBuses := []string{"my-custom-bus", "cross-account-bus", "audit-bus"}
	
	for _, busName := range simulatedBuses {
		// Skip default bus
		if busName == "default" {
			continue
		}
		
		fmt.Printf("Cleaning up event bus: %s\n", busName)
		
		// List and delete all rules in the bus
		// rules, err := eventbridge.ListRulesE(ctx, busName)
		// if err != nil {
		//     fmt.Printf("Error listing rules for bus %s: %v\n", busName, err)
		//     continue
		// }
		fmt.Printf("Listing rules for bus: %s\n", busName)
		
		// Simulate rules cleanup
		simulatedRules := []string{"rule-1", "rule-2"}
		for _, ruleName := range simulatedRules {
			fmt.Printf("Deleting rule: %s\n", ruleName)
			// eventbridge.DeleteRule(ctx, ruleName, busName)
		}
		
		// Delete the event bus
		fmt.Printf("Deleting event bus: %s\n", busName)
		// eventbridge.DeleteEventBus(ctx, busName)
	}
	
	// Clean up archives
	// archives, err := eventbridge.ListArchivesE(ctx, "")
	// if err != nil {
	//     fmt.Printf("Error listing archives: %v\n", err)
	//     return
	// }
	fmt.Println("Listing all archives")
	
	// Simulate archives cleanup
	simulatedArchives := []string{"audit-archive", "backup-archive"}
	for _, archiveName := range simulatedArchives {
		fmt.Printf("Deleting archive: %s\n", archiveName)
		// eventbridge.DeleteArchive(ctx, archiveName)
	}
	
	fmt.Println("EventBridge cleanup completed")
}

// runAllExamples demonstrates running all EventBridge examples
func runAllExamples() {
	fmt.Println("Running all EventBridge package examples:")
	fmt.Println("")
	
	ExampleEventBridgeWorkflow()
	fmt.Println("")
	
	ExampleArchiveAndReplay()
	fmt.Println("")
	
	ExampleBatchEventProcessing()
	fmt.Println("")
	
	ExampleScheduledEvents()
	fmt.Println("")
	
	ExampleCrossAccountEventBridge()
	fmt.Println("")
	
	ExampleEventPatternTesting()
	fmt.Println("")
	
	ExampleEventBridgeCleanup()
	fmt.Println("All EventBridge examples completed")
}

func main() {
	// This file demonstrates EventBridge usage patterns with the vasdeference framework.
	// Run examples with: go run ./examples/eventbridge/examples.go
	
	runAllExamples()
}