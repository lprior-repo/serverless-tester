# EventBridge

Event routing, rules, and custom events for EventBridge testing.

## Quick Start

```bash
# Run EventBridge examples
go test -v ./examples/eventbridge/

# View patterns
go run ./examples/eventbridge/examples.go
```

## Basic Usage

### Create Event Bus
```go
func TestCreateEventBus(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    busName := vdf.PrefixResourceName("custom-bus")
    
    eventBus := eventbridge.CreateEventBus(vdf.Context, &eventbridge.EventBusConfig{
        Name: busName,
        Tags: map[string]string{
            "Environment": "test",
            "Project":     "my-app",
        },
    })
    
    vdf.RegisterCleanup(func() error {
        return eventBridge.DeleteEventBus()
    })
    
    eventbridge.AssertEventBusExists(vdf.Context, busName)
}
```

### Create Rule with Event Pattern
```go
func TestCreateRule(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    busName := vdf.PrefixResourceName("app-bus")
    ruleName := vdf.PrefixResourceName("user-events")
    
    rule := eventbridge.CreateRule(vdf.Context, &eventbridge.RuleConfig{
        Name:         ruleName,
        EventBusName: busName,
        EventPattern: map[string]interface{}{
            "source":      []string{"user.service"},
            "detail-type": []string{"User Created", "User Updated"},
            "detail": map[string]interface{}{
                "status": []string{"active"},
            },
        },
        State:       "ENABLED",
        Description: "Route user events to processing functions",
    })
    
    vdf.RegisterCleanup(func() error {
        return rule.Delete()
    })
    
    eventbridge.AssertRuleExists(vdf.Context, ruleName, busName)
    eventbridge.AssertRuleEnabled(vdf.Context, ruleName, busName)
}
```

### Add Targets to Rule
```go
func TestAddTargets(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    busName := vdf.PrefixResourceName("app-bus")
    ruleName := vdf.PrefixResourceName("order-events")
    
    targets := []eventbridge.TargetConfig{
        {
            ID:  "lambda-processor",
            Arn: "arn:aws:lambda:us-east-1:123456789012:function:process-orders",
            InputTransformer: &eventbridge.InputTransformer{
                InputPathsMap: map[string]string{
                    "orderId":    "$.detail.orderId",
                    "customerId": "$.detail.customerId",
                },
                InputTemplate: `{"orderId": "<orderId>", "customerId": "<customerId>"}`,
            },
        },
        {
            ID:         "sqs-queue",
            Arn:        "arn:aws:sqs:us-east-1:123456789012:order-processing-queue",
            SqsParameters: &eventbridge.SqsParameters{
                MessageGroupId: "order-processing",
            },
        },
    }
    
    err := eventbridge.PutTargets(vdf.Context, ruleName, busName, targets)
    assert.NoError(t, err)
    
    eventbridge.AssertTargetCount(vdf.Context, ruleName, busName, 2)
}
```

## Event Publishing

### Publish Single Event
```go
func TestPublishEvent(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    busName := vdf.PrefixResourceName("app-bus")
    
    event := eventbridge.BuildCustomEvent("user.service", "User Created", map[string]interface{}{
        "userId":    "user-12345",
        "email":     "john@example.com",
        "createdAt": time.Now().Format(time.RFC3339),
        "plan":      "premium",
    })
    
    result := eventbridge.PutEvent(vdf.Context, busName, event)
    eventbridge.AssertEventSent(vdf.Context, result)
}
```

### Publish Batch Events
```go
func TestPublishBatchEvents(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    busName := vdf.PrefixResourceName("app-bus")
    
    events := []eventbridge.CustomEvent{
        eventbridge.BuildCustomEvent("user.service", "User Created", map[string]interface{}{
            "userId": "user-001", "email": "alice@example.com",
        }),
        eventbridge.BuildCustomEvent("user.service", "User Created", map[string]interface{}{
            "userId": "user-002", "email": "bob@example.com",
        }),
        eventbridge.BuildCustomEvent("order.service", "Order Placed", map[string]interface{}{
            "orderId": "order-123", "amount": 99.99,
        }),
    }
    
    result := eventbridge.PutEvents(vdf.Context, busName, events)
    eventbridge.AssertEventBatchSent(vdf.Context, result)
    assert.Equal(t, 3, result.SuccessfulCount)
}
```

### AWS Service Events
```go
func TestAWSServiceEvents(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // S3 event
    s3Event := eventbridge.BuildS3Event("my-bucket", "data/file.json", "ObjectCreated:Put")
    
    // EC2 event
    ec2Event := eventbridge.BuildEC2Event("i-1234567890abcdef0", "running", "pending")
    
    // DynamoDB event
    dynamoEvent := eventbridge.BuildDynamoDBEvent("my-table", "INSERT", map[string]interface{}{
        "id": "item-123",
    })
    
    events := []eventbridge.CustomEvent{s3Event, ec2Event, dynamoEvent}
    result := eventbridge.PutEvents(vdf.Context, "default", events)
    
    assert.Equal(t, 3, result.SuccessfulCount)
}
```

## Scheduled Events

### Cron Schedule
```go
func TestCronSchedule(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ruleName := vdf.PrefixResourceName("daily-report")
    
    rule := eventbridge.CreateScheduledRule(vdf.Context, &eventbridge.ScheduledRuleConfig{
        Name:               ruleName,
        ScheduleExpression: "cron(0 9 * * ? *)", // Daily at 9 AM UTC
        Input: map[string]interface{}{
            "reportType": "daily",
            "format":     "pdf",
        },
        Targets: []eventbridge.TargetConfig{
            {
                ID:  "report-generator",
                Arn: "arn:aws:lambda:us-east-1:123456789012:function:generate-report",
            },
        },
    })
    
    vdf.RegisterCleanup(func() error {
        return rule.Delete()
    })
    
    eventbridge.AssertRuleExists(vdf.Context, ruleName, "default")
}
```

### Rate Schedule
```go
func TestRateSchedule(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ruleName := vdf.PrefixResourceName("health-check")
    
    rule := eventbridge.CreateScheduledRule(vdf.Context, &eventbridge.ScheduledRuleConfig{
        Name:               ruleName,
        ScheduleExpression: "rate(5 minutes)",
        Input: map[string]interface{}{
            "checkType": "health",
            "timeout":   30,
        },
        Targets: []eventbridge.TargetConfig{
            {
                ID:  "health-checker",
                Arn: "arn:aws:lambda:us-east-1:123456789012:function:check-health",
            },
        },
    })
    
    vdf.RegisterCleanup(func() error {
        return rule.Delete()
    })
}
```

## Advanced Patterns

### Event Pattern Testing
```go
func TestEventPatterns(t *testing.T) {
    testCases := []struct {
        name    string
        pattern map[string]interface{}
        event   map[string]interface{}
        matches bool
    }{
        {
            name: "UserCreatedMatches",
            pattern: map[string]interface{}{
                "source":      []string{"user.service"},
                "detail-type": []string{"User Created"},
            },
            event: map[string]interface{}{
                "source":      "user.service",
                "detail-type": "User Created",
                "detail":      map[string]interface{}{"userId": "123"},
            },
            matches: true,
        },
        {
            name: "DifferentSourceNoMatch",
            pattern: map[string]interface{}{
                "source": []string{"order.service"},
            },
            event: map[string]interface{}{
                "source": "user.service",
            },
            matches: false,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            matches := eventbridge.TestEventPattern(tc.pattern, tc.event)
            assert.Equal(t, tc.matches, matches)
        })
    }
}
```

### Archive and Replay
```go
func TestArchiveAndReplay(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    busName := vdf.PrefixResourceName("audit-bus")
    archiveName := vdf.PrefixResourceName("audit-archive")
    
    // Create archive
    archive := eventbridge.CreateArchive(vdf.Context, &eventbridge.ArchiveConfig{
        ArchiveName:    archiveName,
        EventBusName:   busName,
        EventPattern: map[string]interface{}{
            "source": []string{"audit.service"},
        },
        RetentionDays: 7,
        Description:   "Archive audit events for compliance",
    })
    
    vdf.RegisterCleanup(func() error {
        return archive.Delete()
    })
    
    // Publish some events to archive
    events := []eventbridge.CustomEvent{
        eventbridge.BuildCustomEvent("audit.service", "Login Attempt", map[string]interface{}{
            "userId": "user-123", "success": true,
        }),
        eventbridge.BuildCustomEvent("audit.service", "Data Access", map[string]interface{}{
            "userId": "user-123", "resource": "/api/users",
        }),
    }
    
    eventbridge.PutEvents(vdf.Context, busName, events)
    
    // Start replay (after some time has passed)
    replayName := vdf.PrefixResourceName("audit-replay")
    replay := eventbridge.StartReplay(vdf.Context, &eventbridge.ReplayConfig{
        ReplayName:     replayName,
        ArchiveName:    archiveName,
        EventStartTime: time.Now().Add(-1 * time.Hour),
        EventEndTime:   time.Now(),
        Destination: eventbridge.ReplayDestination{
            Arn: "arn:aws:events:us-east-1:123456789012:event-bus/replay-bus",
        },
    })
    
    vdf.RegisterCleanup(func() error {
        return replay.Cancel()
    })
    
    eventbridge.AssertArchiveExists(vdf.Context, archiveName)
    eventbridge.AssertReplayStarted(vdf.Context, replayName)
}
```

### Cross-Account Events
```go
func TestCrossAccountEvents(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    busName := vdf.PrefixResourceName("shared-bus")
    
    // Create event bus with cross-account policy
    eventBus := eventbridge.CreateEventBus(vdf.Context, &eventbridge.EventBusConfig{
        Name: busName,
        Policy: eventbridge.CrossAccountPolicy{
            AllowedAccounts: []string{"123456789012", "210987654321"},
            AllowedActions:  []string{"events:PutEvents"},
        },
    })
    
    vdf.RegisterCleanup(func() error {
        return eventBus.Delete()
    })
    
    eventbridge.AssertEventBusPolicy(vdf.Context, busName, "123456789012")
}
```

## Assertions

```go
func TestEventBridgeAssertions(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    busName := vdf.PrefixResourceName("test-bus")
    ruleName := vdf.PrefixResourceName("test-rule")
    
    // Event bus assertions
    eventbridge.AssertEventBusExists(vdf.Context, busName)
    eventbridge.AssertEventBusARN(vdf.Context, busName, "arn:aws:events:us-east-1:123456789012:event-bus/"+busName)
    
    // Rule assertions
    eventbridge.AssertRuleExists(vdf.Context, ruleName, busName)
    eventbridge.AssertRuleEnabled(vdf.Context, ruleName, busName)
    eventbridge.AssertRulePattern(vdf.Context, ruleName, busName, map[string]interface{}{
        "source": []string{"test.service"},
    })
    
    // Target assertions
    eventbridge.AssertTargetCount(vdf.Context, ruleName, busName, 2)
    eventbridge.AssertTargetExists(vdf.Context, ruleName, busName, "lambda-target")
    
    // Event publishing assertions
    result := eventbridge.PutEvent(vdf.Context, busName, testEvent)
    eventbridge.AssertEventSent(vdf.Context, result)
    eventbridge.AssertNoFailures(vdf.Context, result)
}
```

## Run Examples

```bash
# View all patterns
go run ./examples/eventbridge/examples.go

# Run as tests
go test -v ./examples/eventbridge/

# Test specific patterns
go test -v ./examples/eventbridge/ -run TestCreateRule
go test -v ./examples/eventbridge/ -run TestPublishBatchEvents
```

## API Reference

### Event Bus Management
- `eventbridge.CreateEventBus(ctx, config)` - Create custom event bus
- `eventbridge.DeleteEventBus(ctx, name)` - Delete event bus
- `eventbridge.AssertEventBusExists(ctx, name)` - Assert bus exists

### Rule Management  
- `eventbridge.CreateRule(ctx, config)` - Create event rule
- `eventbridge.DeleteRule(ctx, name, busName)` - Delete rule
- `eventbridge.PutTargets(ctx, ruleName, busName, targets)` - Add targets to rule

### Event Publishing
- `eventbridge.PutEvent(ctx, busName, event)` - Publish single event
- `eventbridge.PutEvents(ctx, busName, events)` - Publish batch events
- `eventbridge.BuildCustomEvent(source, detailType, detail)` - Build custom event

### Event Builders
- `eventbridge.BuildS3Event(bucket, key, eventName)` - S3 events
- `eventbridge.BuildEC2Event(instanceId, newState, oldState)` - EC2 events
- `eventbridge.BuildDynamoDBEvent(table, eventName, keys)` - DynamoDB events

### Scheduling
- `eventbridge.CreateScheduledRule(ctx, config)` - Create scheduled rule
- `eventbridge.BuildCronExpression(minute, hour, day, month, year)` - Build cron
- `eventbridge.BuildRateExpression(value, unit)` - Build rate expression

---

**Next:** [Step Functions Examples](../stepfunctions/) | [Core Examples](../core/)