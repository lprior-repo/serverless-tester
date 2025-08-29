# 📡 EventBridge Functional Testing Package

A **pure functional programming** EventBridge testing package built with immutable data structures, monadic error handling, and type-safe operations using `samber/lo` and `samber/mo`.

## ✨ Functional Programming Features

🔒 **Type-Safe Event Operations** with Go generics for event patterns  
🚀 **Immutable Event Structures** with functional builders  
📦 **Pure Function Pipeline** for event publishing and rule management  
⚡ **Monadic Error Handling** with safe operations and pattern validation  
🔧 **Functional Composition** for complex event routing and filtering  
🧮 **Mathematical Precision** - predictable, verifiable event operations  

## 🏗️ Pure Functional Architecture

### Zero-Mutation Event System
- **Immutable Events**: All event structures are immutable with functional builders
- **Pure Functions**: Event validation, pattern matching, and routing are pure
- **Monadic Results**: Safe error handling with `Option[T]` and `Result[T]` types
- **Type Safety**: Compile-time guarantees for event patterns and routing

### Functional Event Operations
- **Event Builders**: Immutable event creation with method chaining
- **Pattern Validation**: Pure functions for event pattern verification
- **Rule Composition**: Functional composition for complex event routing
- **Archive & Replay**: Immutable operations with monadic error handling

## 🚀 Functional Usage Examples

### Immutable Event Creation & Publishing

```go
import (
    "vasdeference/eventbridge"
    "github.com/samber/mo"
)

func TestFunctionalEventBridge(t *testing.T) {
    // Immutable event bus configuration
    config := NewFunctionalEventBridgeConfig("orders-bus",
        WithEventBridgeTimeout(30*time.Second),
        WithEventPattern(`{"source": ["ecommerce"]}`),
        WithRuleState(types.RuleStateEnabled),
    )
    
    // Build immutable event with functional composition
    orderEvent := NewFunctionalEventBridgeEvent(
        "ecommerce.orders",
        "Order Created", 
        `{"orderId": "12345", "amount": 99.99}`,
    ).WithEventBusName("orders-bus").
      WithResources([]string{"arn:aws:lambda:...:function:process-order"})
    
    // Publish with monadic error handling
    publishResult := SimulateFunctionalEventBridgePutEvents(
        []FunctionalEventBridgeEvent{orderEvent}, 
        config,
    )
    
    // Safe error handling without exceptions
    publishResult.GetError().
        Map(func(err error) error {
            t.Errorf("Publishing failed: %v", err)
            return err
        }).
        OrElse(func() {
            t.Log("✓ Events published successfully")
        })
}
```
func TestEventPublishing(t *testing.T) {
    ctx := &eventbridge.TestContext{
        T:         t,
        AwsConfig: aws.Config{Region: "us-east-1"},
        Region:    "us-east-1",
    }
    
    // Build and send custom event
    event := eventbridge.BuildCustomEvent("user.service", "User Created", map[string]interface{}{
        "userId": "user-123",
        "email":  "test@example.com",
    })
    
    result := eventbridge.PutEvent(ctx, event)
    eventbridge.AssertEventSent(ctx, result)
}
```

### Rule and Target Management

```go
func TestRuleManagement(t *testing.T) {
    ctx := &eventbridge.TestContext{
        T:         t,
        AwsConfig: aws.Config{Region: "us-east-1"},
        Region:    "us-east-1",
    }
    
    // Create rule with event pattern
    ruleConfig := eventbridge.RuleConfig{
        Name:         "user-events-rule",
        Description:  "Rule for user service events",
        EventPattern: `{"source": ["user.service"], "detail-type": ["User Created"]}`,
        State:        eventbridge.RuleStateEnabled,
        EventBusName: "default",
    }
    
    rule := eventbridge.CreateRule(ctx, ruleConfig)
    
    // Add Lambda target
    targets := []eventbridge.TargetConfig{{
        ID:      "process-user-events",
        Arn:     "arn:aws:lambda:us-east-1:123456789012:function:process-user",
        RoleArn: "arn:aws:iam::123456789012:role/EventBridgeExecutionRole",
    }}
    
    eventbridge.PutTargets(ctx, rule.Name, "default", targets)
    
    // Verify setup
    eventbridge.AssertRuleExists(ctx, "user-events-rule", "default")
    eventbridge.AssertRuleEnabled(ctx, "user-events-rule", "default")
    eventbridge.AssertTargetExists(ctx, "user-events-rule", "default", "process-user-events")
}
```

### Archive and Replay

```go
func TestArchiveAndReplay(t *testing.T) {
    ctx := &eventbridge.TestContext{
        T:         t,
        AwsConfig: aws.Config{Region: "us-east-1"},
        Region:    "us-east-1",
    }
    
    // Create custom event bus
    busConfig := eventbridge.EventBusConfig{
        Name: "audit-events",
        Tags: map[string]string{
            "Purpose": "auditing",
        },
    }
    eventBus := eventbridge.CreateEventBus(ctx, busConfig)
    
    // Create archive
    archiveConfig := eventbridge.ArchiveConfig{
        ArchiveName:    "audit-archive",
        EventSourceArn: eventBus.Arn,
        EventPattern:   `{"source": ["audit.service"]}`,
        RetentionDays:  30,
    }
    archive := eventbridge.CreateArchive(ctx, archiveConfig)
    
    // Start replay
    replayConfig := eventbridge.ReplayConfig{
        ReplayName:     "audit-replay",
        EventSourceArn: archive.ArchiveArn,
        EventStartTime: time.Now().Add(-24 * time.Hour),
        EventEndTime:   time.Now(),
    }
    
    replay := eventbridge.StartReplay(ctx, replayConfig)
    eventbridge.WaitForReplayCompletion(ctx, replay.ReplayName, 10*time.Minute)
}
```

### Batch Event Processing

```go
func TestBatchEvents(t *testing.T) {
    ctx := &eventbridge.TestContext{
        T:         t,
        AwsConfig: aws.Config{Region: "us-east-1"},
        Region:    "us-east-1",
    }
    
    // Create events from different AWS services
    events := []eventbridge.CustomEvent{
        eventbridge.BuildS3Event("data-bucket", "file.csv", "ObjectCreated:Put"),
        eventbridge.BuildEC2Event("i-1234567890abcdef0", "running", "pending"),
        eventbridge.BuildLambdaEvent("processor", "arn:aws:lambda:...", "Active", ""),
    }
    
    // Send batch
    result := eventbridge.PutEvents(ctx, events)
    eventbridge.AssertEventBatchSent(ctx, result)
}
```

### Scheduled Events

```go
func TestScheduledEvents(t *testing.T) {
    ctx := &eventbridge.TestContext{
        T:         t,
        AwsConfig: aws.Config{Region: "us-east-1"},
        Region:    "us-east-1",
    }
    
    // Daily report at 9 AM UTC
    dailyRule := eventbridge.BuildScheduledEventWithCron("0 9 * * ? *", map[string]interface{}{
        "reportType": "daily",
    })
    dailyRule.Name = "daily-reports"
    
    rule := eventbridge.CreateRule(ctx, dailyRule)
    
    // Health check every 5 minutes
    healthRule := eventbridge.BuildScheduledEventWithRate(5, "minutes")
    healthRule.Name = "health-checks"
    
    eventbridge.CreateRule(ctx, healthRule)
}
```

## Testing

The package includes comprehensive tests:

```bash
# Run all tests (requires AWS credentials)
go test ./eventbridge -v

# Run tests without AWS integration (recommended for CI)
go test ./eventbridge -v -short

# Run specific test patterns
go test ./eventbridge -run TestEventBuilders -v

# Run benchmarks
go test ./eventbridge -bench=. -v
```

## Package Structure

```
eventbridge/
├── eventbridge.go       # Core types and configurations
├── events.go           # Event publishing and validation
├── rules.go            # Rule management functions
├── targets.go          # Target management functions
├── eventbus.go         # Event bus operations
├── archive.go          # Archive and replay functionality
├── builders.go         # AWS service event builders
├── assertions.go       # Testing assertion utilities
├── eventbridge_test.go # Basic functionality tests
├── comprehensive_test.go # Integration tests
├── examples_test.go    # Usage examples
└── README.md           # This documentation
```

## Error Handling

All functions follow the Terratest pattern:
- `Function` variants panic on error (for test assertions)
- `FunctionE` variants return errors (for conditional logic)

Example:
```go
// Panics on error
result := eventbridge.PutEvent(ctx, event)

// Returns error
result, err := eventbridge.PutEventE(ctx, event)
if err != nil {
    // Handle error
}
```

## Retry Logic

All AWS operations include automatic retry with exponential backoff:
- Default: 3 attempts with 1s base delay
- Configurable via `RetryConfig`
- Proper backoff calculation with jitter

## Observability

The package provides structured logging using `log/slog`:
- Operation start/end logging
- Duration tracking
- Error context
- Request/response details (sanitized)

## Best Practices

1. **Use TestContext**: Always create proper test context with AWS config
2. **Follow TDD**: Write tests first, then implementation
3. **Clean Up**: Use defer for resource cleanup in tests
4. **Validate Patterns**: Always validate event patterns before using
5. **Use Assertions**: Leverage assertion functions for reliable tests
6. **Handle Errors**: Use `E` variants for conditional logic

## Dependencies

- `github.com/aws/aws-sdk-go-v2/service/eventbridge`
- `github.com/stretchr/testify` (for testing)

## Integration with Terratest

This package is designed to integrate seamlessly with Terratest workflows:

```go
func TestEventBridgeInfrastructure(t *testing.T) {
    // Deploy infrastructure with Terraform/Terragrunt
    terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
        TerraformDir: "../terraform",
    })
    defer terraform.Destroy(t, terraformOptions)
    terraform.InitAndApply(t, terraformOptions)
    
    // Get outputs
    eventBusArn := terraform.Output(t, terraformOptions, "event_bus_arn")
    
    // Test EventBridge functionality
    ctx := &eventbridge.TestContext{
        T:         t,
        AwsConfig: aws.Config{Region: "us-east-1"},
        Region:    "us-east-1",
    }
    
    // Your EventBridge tests here...
}
```