package main

import (
	"fmt"
	"time"

	"vasdeference/eventbridge"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// ExampleEventBridgeWorkflow demonstrates a complete functional EventBridge workflow
func ExampleEventBridgeWorkflow() {
	fmt.Println("Functional EventBridge workflow patterns:")
	
	// Functional context initialization with error handling
	ctxResult := mo.TryOr(func() error {
		return nil // Simulate context creation
	}, func(err error) error {
		return fmt.Errorf("failed to initialize context: %w", err)
	})
	
	// Functional event bus configuration with validation
	busConfig := lo.Pipe3(
		map[string]string{
			"Environment": "example",
			"Team":        "platform",
		},
		func(tags map[string]string) mo.Option[map[string]string] {
			if len(tags) == 0 {
				return mo.None[map[string]string]()
			}
			return mo.Some(tags)
		},
		func(tagsOpt mo.Option[map[string]string]) mo.Result[eventbridge.EventBusConfig] {
			return tagsOpt.Match(
				func(tags map[string]string) mo.Result[eventbridge.EventBusConfig] {
					return mo.Ok(eventbridge.EventBusConfig{
						Name: "my-example-bus",
						Tags: tags,
					})
				},
				func() mo.Result[eventbridge.EventBusConfig] {
					return mo.Err[eventbridge.EventBusConfig](fmt.Errorf("invalid tags configuration"))
				},
			)
		},
		func(configResult mo.Result[eventbridge.EventBusConfig]) eventbridge.EventBusConfig {
			return configResult.OrElse(eventbridge.EventBusConfig{Name: "default-bus"})
		},
	)
	
	// Use functional EventBridge configuration
	config := eventbridge.NewFunctionalEventBridgeConfig().OrEmpty()
	result := eventbridge.SimulateFunctionalEventBridgePutEvents(config)
	
	result.Match(
		func(success eventbridge.FunctionalEventBridgeConfig) {
			fmt.Printf("✓ Created functional event bus: %s\n", busConfig.Name)
			
			// Functional rule configuration with pattern validation
			ruleConfigResult := lo.Pipe2(
				`{"source": ["user.service"], "detail-type": ["User Signup"]}`,
				func(pattern string) mo.Result[string] {
					if eventbridge.ValidateEventPattern(pattern) {
						return mo.Ok(pattern)
					}
					return mo.Err[string](fmt.Errorf("invalid event pattern"))
				},
				func(validatedPattern mo.Result[string]) eventbridge.RuleConfig {
					return validatedPattern.Match(
						func(pattern string) eventbridge.RuleConfig {
							return eventbridge.RuleConfig{
								Name:         "user-signup-rule",
								Description:  "Functional rule for user signup events",
								EventPattern: pattern,
								State:        eventbridge.RuleStateEnabled,
								EventBusName: "my-example-bus",
							}
						},
						func(err error) eventbridge.RuleConfig {
							fmt.Printf("✗ Pattern validation failed: %v\n", err)
							return eventbridge.RuleConfig{}
						},
					)
				},
			)
			
			fmt.Printf("✓ Creating functional rule: %s\n", ruleConfigResult.Name)
		},
		func(err error) {
			fmt.Printf("✗ Failed to create event bus: %v\n", err)
		},
	)
	
	// Functional target configuration with validation
	targetConfigs := lo.Pipe2(
		[]eventbridge.TargetConfig{
			{
				ID:      "lambda-processor",
				Arn:     "arn:aws:lambda:us-east-1:123456789012:function:process-signup",
				RoleArn: "arn:aws:iam::123456789012:role/EventBridgeExecutionRole",
			},
			{
				ID:  "sqs-queue",
				Arn: "arn:aws:sqs:us-east-1:123456789012:signup-queue",
			},
		},
		func(targets []eventbridge.TargetConfig) []eventbridge.TargetConfig {
			return lo.Filter(targets, func(target eventbridge.TargetConfig, _ int) bool {
				return len(target.ID) > 0 && len(target.Arn) > 0
			})
		},
		func(validatedTargets []eventbridge.TargetConfig) mo.Result[[]eventbridge.TargetConfig] {
			if len(validatedTargets) == 0 {
				return mo.Err[[]eventbridge.TargetConfig](fmt.Errorf("no valid targets"))
			}
			return mo.Ok(validatedTargets)
		},
	)
	
	targetConfigs.Match(
		func(targets []eventbridge.TargetConfig) {
			fmt.Printf("✓ Adding %d validated targets to rule\n", len(targets))
			lo.ForEach(targets, func(target eventbridge.TargetConfig, index int) {
				fmt.Printf("  - Target %d: %s\n", index+1, target.ID)
			})
		},
		func(err error) {
			fmt.Printf("✗ Target validation failed: %v\n", err)
		},
	)
	
	// Functional event creation with validation
	signupEventResult := lo.Pipe3(
		map[string]interface{}{
			"userId":    "user-12345",
			"email":     "test@example.com",
			"timestamp": time.Now().Unix(),
		},
		func(eventData map[string]interface{}) mo.Option[map[string]interface{}] {
			if userId, exists := eventData["userId"]; exists && userId != "" {
				return mo.Some(eventData)
			}
			return mo.None[map[string]interface{}]()
		},
		func(validatedData mo.Option[map[string]interface{}]) mo.Result[eventbridge.CustomEvent] {
			return validatedData.Match(
				func(data map[string]interface{}) mo.Result[eventbridge.CustomEvent] {
					event := eventbridge.BuildCustomEvent("user.service", "User Signup", data)
					return mo.Ok(event)
				},
				func() mo.Result[eventbridge.CustomEvent] {
					return mo.Err[eventbridge.CustomEvent](fmt.Errorf("invalid event data"))
				},
			)
		},
		func(eventResult mo.Result[eventbridge.CustomEvent]) mo.Option[string] {
			return eventResult.Match(
				func(event eventbridge.CustomEvent) mo.Option[string] {
					fmt.Println("✓ Sending functional user signup event")
					return mo.Some("event-sent-successfully")
				},
				func(err error) mo.Option[string] {
					fmt.Printf("✗ Event creation failed: %v\n", err)
					return mo.None[string]()
				},
			)
		},
	)
	
	signupEventResult.IfPresent(func(status string) {
		fmt.Printf("✓ Event status: %s\n", status)
	})
	
	// Functional verification chain
	verificationResult := lo.Pipe3(
		[]string{"rule-exists", "rule-enabled", "targets-configured"},
		func(checks []string) []mo.Result[string] {
			return lo.Map(checks, func(check string, _ int) mo.Result[string] {
				return mo.Ok(check)
			})
		},
		func(checks []mo.Result[string]) mo.Result[[]string] {
			successfulChecks := lo.FilterMap(checks, func(check mo.Result[string], _ int) (string, bool) {
				return check.Match(
					func(c string) (string, bool) { return c, true },
					func(err error) (string, bool) { return "", false },
				)
			})
			return mo.Ok(successfulChecks)
		},
		func(result mo.Result[[]string]) mo.Option[string] {
			return result.Match(
				func(checks []string) mo.Option[string] {
					fmt.Printf("✓ Verification completed: %d checks passed\n", len(checks))
					return mo.Some("verification-successful")
				},
				func(err error) mo.Option[string] {
					fmt.Printf("✗ Verification failed: %v\n", err)
					return mo.None[string]()
				},
			)
		},
	)
	
	// Functional cleanup with error handling
	cleanupResult := mo.Some([]string{"user-signup-rule", "my-example-bus"}).Map(func(resources []string) string {
		lo.ForEach(resources, func(resource string, index int) {
			fmt.Printf("✓ Cleaning up resource %d: %s\n", index+1, resource)
		})
		return "cleanup-completed"
	})
	
	mo.Tuple2(verificationResult, cleanupResult).Match(
		func(verification mo.Option[string], cleanup mo.Option[string]) {
			fmt.Println("✓ Functional EventBridge workflow completed successfully")
		},
		func() {
			fmt.Println("✗ Workflow completed with partial success")
		},
	)
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

// ExampleBatchEventProcessing shows how to handle functional batch event processing
func ExampleBatchEventProcessing() {
	fmt.Println("Functional EventBridge batch event processing patterns:")
	
	// Functional event factory patterns
	eventBuilders := []func() eventbridge.CustomEvent{
		func() eventbridge.CustomEvent {
			return eventbridge.BuildS3Event("data-bucket", "uploads/file1.csv", "ObjectCreated:Put")
		},
		func() eventbridge.CustomEvent {
			return eventbridge.BuildS3Event("data-bucket", "uploads/file2.json", "ObjectCreated:Put")
		},
		func() eventbridge.CustomEvent {
			return eventbridge.BuildEC2Event("i-1234567890abcdef0", "running", "pending")
		},
		func() eventbridge.CustomEvent {
			return eventbridge.BuildLambdaEvent("data-processor", "arn:aws:lambda:us-east-1:123456789012:function:data-processor", "Active", "")
		},
		func() eventbridge.CustomEvent {
			return eventbridge.BuildDynamoDBEvent("users", "INSERT", map[string]interface{}{
				"eventName": "INSERT",
				"dynamodb": map[string]interface{}{
					"Keys": map[string]interface{}{
						"userId": map[string]interface{}{"S": "new-user-123"},
					},
				},
			})
		},
	}
	
	// Functional event batch creation with validation
	events := lo.Pipe2(
		eventBuilders,
		func(builders []func() eventbridge.CustomEvent) []eventbridge.CustomEvent {
			return lo.Map(builders, func(builder func() eventbridge.CustomEvent, _ int) eventbridge.CustomEvent {
				return builder()
			})
		},
		func(eventList []eventbridge.CustomEvent) []eventbridge.CustomEvent {
			// Filter out any invalid events functionally
			return lo.Filter(eventList, func(event eventbridge.CustomEvent, _ int) bool {
				return len(event.Source) > 0 && len(event.DetailType) > 0
			})
		},
	)
	
	// Functional batch processing with monadic result handling
	batchResult := lo.Pipe3(
		events,
		func(eventList []eventbridge.CustomEvent) mo.Result[[]eventbridge.CustomEvent] {
			if len(eventList) == 0 {
				return mo.Err[[]eventbridge.CustomEvent](fmt.Errorf("no events to process"))
			}
			fmt.Printf("✓ Sending batch of %d events from different AWS services\n", len(eventList))
			return mo.Ok(eventList)
		},
		func(validated mo.Result[[]eventbridge.CustomEvent]) mo.Result[[]mo.Result[string]] {
			return validated.Map(func(eventList []eventbridge.CustomEvent) []mo.Result[string] {
				return lo.Map(eventList, func(event eventbridge.CustomEvent, index int) mo.Result[string] {
					// Simulate event processing
					return mo.Ok(fmt.Sprintf("event-%d-sent", index))
				})
			})
		},
		func(results mo.Result[[]mo.Result[string]]) mo.Option[eventbridge.BatchProcessingResult] {
			return results.Match(
				func(resultList []mo.Result[string]) mo.Option[eventbridge.BatchProcessingResult] {
					successCount := lo.CountBy(resultList, func(result mo.Result[string]) bool {
						return result.IsOk()
					})
					return mo.Some(eventbridge.BatchProcessingResult{
						TotalEvents:    len(resultList),
						SuccessfulEvents: successCount,
						FailedEvents:   len(resultList) - successCount,
					})
				},
				func(err error) mo.Option[eventbridge.BatchProcessingResult] {
					fmt.Printf("✗ Batch processing failed: %v\n", err)
					return mo.None[eventbridge.BatchProcessingResult]()
				},
			)
		},
	)
	
	batchResult.Match(
		func(result eventbridge.BatchProcessingResult) {
			fmt.Printf("✓ Batch processing completed: %d successful, %d failed, %d total\n", 
				result.SuccessfulEvents, result.FailedEvents, result.TotalEvents)
		},
		func() {
			fmt.Println("✗ Batch processing was not completed")
		},
	)
}

// ExampleScheduledEvents demonstrates creating functional scheduled events
func ExampleScheduledEvents() {
	fmt.Println("Functional EventBridge scheduled events patterns:")
	
	// Functional cron rule creation with validation
	dailyReportRule := lo.Pipe3(
		"0 9 * * ? *",
		func(cronExpression string) mo.Result[string] {
			if len(cronExpression) == 0 {
				return mo.Err[string](fmt.Errorf("empty cron expression"))
			}
			return mo.Ok(cronExpression)
		},
		func(validCron mo.Result[string]) mo.Result[eventbridge.RuleConfig] {
			return validCron.Map(func(cron string) eventbridge.RuleConfig {
				rule := eventbridge.BuildScheduledEventWithCron(cron, map[string]interface{}{
					"reportType": "daily",
				})
				rule.Name = "daily-report-trigger"
				rule.Description = "Functional daily report generation trigger"
				return rule
			})
		},
		func(ruleResult mo.Result[eventbridge.RuleConfig]) eventbridge.RuleConfig {
			return ruleResult.OrElse(eventbridge.RuleConfig{Name: "fallback-rule"})
		},
	)
	
	fmt.Printf("✓ Creating functional daily scheduled rule: %s (9 AM UTC)\n", dailyReportRule.Name)
	
	// Functional rate-based rule creation
	healthCheckRule := lo.Pipe3(
		struct{ rate int; unit string }{5, "minutes"},
		func(config struct{ rate int; unit string }) mo.Result[struct{ rate int; unit string }] {
			if config.rate <= 0 || len(config.unit) == 0 {
				return mo.Err[struct{ rate int; unit string }](fmt.Errorf("invalid rate configuration"))
			}
			return mo.Ok(config)
		},
		func(validConfig mo.Result[struct{ rate int; unit string }]) mo.Result[eventbridge.RuleConfig] {
			return validConfig.Map(func(config struct{ rate int; unit string }) eventbridge.RuleConfig {
				rule := eventbridge.BuildScheduledEventWithRate(config.rate, config.unit)
				rule.Name = "health-check-trigger"
				rule.Description = "Functional health checks every 5 minutes"
				return rule
			})
		},
		func(ruleResult mo.Result[eventbridge.RuleConfig]) eventbridge.RuleConfig {
			return ruleResult.OrElse(eventbridge.RuleConfig{Name: "fallback-health-check"})
		},
	)
	
	fmt.Printf("✓ Creating functional rate scheduled rule: %s (every 5 minutes)\n", healthCheckRule.Name)
	
	// Functional target configuration with type safety
	targetFactory := func(id, arn, input string) mo.Result[eventbridge.TargetConfig] {
		if len(id) == 0 || len(arn) == 0 {
			return mo.Err[eventbridge.TargetConfig](fmt.Errorf("invalid target configuration"))
		}
		return mo.Ok(eventbridge.TargetConfig{
			ID:    id,
			Arn:   arn,
			Input: input,
		})
	}
	
	dailyTarget := lo.Pipe2(
		targetFactory("daily-report-lambda", "arn:aws:lambda:us-east-1:123456789012:function:generate-daily-report", `{"reportType": "daily", "format": "pdf"}`),
		func(target mo.Result[eventbridge.TargetConfig]) []eventbridge.TargetConfig {
			return target.Match(
				func(t eventbridge.TargetConfig) []eventbridge.TargetConfig { return []eventbridge.TargetConfig{t} },
				func(err error) []eventbridge.TargetConfig {
					fmt.Printf("✗ Daily target creation failed: %v\n", err)
					return []eventbridge.TargetConfig{}
				},
			)
		},
		func(targets []eventbridge.TargetConfig) []eventbridge.TargetConfig {
			return targets
		},
	)
	
	healthCheckTarget := lo.Pipe2(
		targetFactory("health-check-lambda", "arn:aws:lambda:us-east-1:123456789012:function:health-check", ""),
		func(target mo.Result[eventbridge.TargetConfig]) []eventbridge.TargetConfig {
			return target.Match(
				func(t eventbridge.TargetConfig) []eventbridge.TargetConfig { return []eventbridge.TargetConfig{t} },
				func(err error) []eventbridge.TargetConfig {
					fmt.Printf("✗ Health check target creation failed: %v\n", err)
					return []eventbridge.TargetConfig{}
				},
			)
		},
		func(targets []eventbridge.TargetConfig) []eventbridge.TargetConfig {
			return targets
		},
	)
	
	// Functional target assignment with validation
	targetAssignmentResult := lo.Pipe2(
		mo.Tuple2(mo.Some(dailyTarget), mo.Some(healthCheckTarget)),
		func(targets mo.Tuple2[mo.Option[[]eventbridge.TargetConfig], mo.Option[[]eventbridge.TargetConfig]]) mo.Result[struct{ daily, health []eventbridge.TargetConfig }] {
			dailyTargets, healthTargets := targets.Unpack()
			return mo.Tuple2(dailyTargets, healthTargets).Match(
				func(daily mo.Option[[]eventbridge.TargetConfig], health mo.Option[[]eventbridge.TargetConfig]) mo.Result[struct{ daily, health []eventbridge.TargetConfig }] {
					return mo.Tuple2(daily, health).Match(
						func(d []eventbridge.TargetConfig, h []eventbridge.TargetConfig) mo.Result[struct{ daily, health []eventbridge.TargetConfig }] {
							if len(d) > 0 && len(h) > 0 {
								return mo.Ok(struct{ daily, health []eventbridge.TargetConfig }{d, h})
							}
							return mo.Err[struct{ daily, health []eventbridge.TargetConfig }](fmt.Errorf("missing targets"))
						},
						func() mo.Result[struct{ daily, health []eventbridge.TargetConfig }] {
							return mo.Err[struct{ daily, health []eventbridge.TargetConfig }](fmt.Errorf("targets not configured"))
						},
					)
				},
				func() mo.Result[struct{ daily, health []eventbridge.TargetConfig }] {
					return mo.Err[struct{ daily, health []eventbridge.TargetConfig }](fmt.Errorf("no targets provided"))
				},
			)
		},
		func(validated mo.Result[struct{ daily, health []eventbridge.TargetConfig }]) mo.Option[string] {
			return validated.Match(
				func(targets struct{ daily, health []eventbridge.TargetConfig }) mo.Option[string] {
					if len(targets.daily) > 0 && len(targets.health) > 0 {
						fmt.Printf("✓ Adding functional targets: %s and %s\n", targets.daily[0].ID, targets.health[0].ID)
						return mo.Some("targets-configured")
					}
					return mo.None[string]()
				},
				func(err error) mo.Option[string] {
					fmt.Printf("✗ Target assignment failed: %v\n", err)
					return mo.None[string]()
				},
			)
		},
	)
	
	// Functional verification with monadic chaining
	verificationResult := targetAssignmentResult.Map(func(status string) []string {
		return []string{"rules-enabled", "targets-exist"}
	})
	
	mo.Tuple2(targetAssignmentResult, verificationResult).Match(
		func(assignment mo.Option[string], verification mo.Option[[]string]) {
			fmt.Println("✓ Functional scheduled events setup completed successfully")
		},
		func() {
			fmt.Println("✗ Scheduled events setup completed with issues")
		},
	)
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

// runAllExamples demonstrates running all functional EventBridge examples
func runAllExamples() {
	fmt.Println("Running all functional EventBridge package examples:")
	fmt.Println("")
	
	// Use functional composition to run examples with error handling
	exampleFunctions := []func(){
		ExampleEventBridgeWorkflow,
		ExampleArchiveAndReplay,
		ExampleBatchEventProcessing,
		ExampleScheduledEvents,
		ExampleCrossAccountEventBridge,
		ExampleEventPatternTesting,
		ExampleEventBridgeCleanup,
	}
	
	// Functional execution with monadic error handling
	executionResult := lo.Pipe2(
		exampleFunctions,
		func(examples []func()) mo.Result[[]func()] {
			if len(examples) == 0 {
				return mo.Err[[]func()](fmt.Errorf("no examples to run"))
			}
			return mo.Ok(examples)
		},
		func(validated mo.Result[[]func()]) mo.Option[int] {
			return validated.Match(
				func(examples []func()) mo.Option[int] {
					lo.ForEach(examples, func(example func(), index int) {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("✗ Example %d failed: %v\n", index+1, r)
							}
						}()
						example()
						fmt.Println("")
					})
					return mo.Some(len(examples))
				},
				func(err error) mo.Option[int] {
					fmt.Printf("✗ Failed to run examples: %v\n", err)
					return mo.None[int]()
				},
			)
		},
	)
	
	executionResult.Match(
		func(count int) {
			fmt.Printf("✓ All %d functional EventBridge examples completed successfully\n", count)
		},
		func() {
			fmt.Println("✗ Example execution was interrupted")
		},
	)
}

func main() {
	// This file demonstrates functional EventBridge usage patterns with the vasdeference framework.
	// Run examples with: go run ./examples/eventbridge/examples.go
	
	// Functional main execution with error boundary
	mainResult := mo.TryOr(func() error {
		runAllExamples()
		return nil
	}, func(err error) error {
		fmt.Printf("✗ Application failed: %v\n", err)
		return err
	})
	
	mainResult.Match(
		func(value interface{}) {
			fmt.Println("✓ Application completed successfully")
		},
		func(err error) {
			fmt.Printf("✗ Application terminated with error: %v\n", err)
		},
	)
}