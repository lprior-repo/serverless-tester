# VasDeference Examples

Get started quickly with copy-paste ready examples for testing AWS serverless applications.

## Quick Start

```bash
# Run all examples
go test -v ./examples/...

# Run specific service examples
go test -v ./examples/lambda/
go test -v ./examples/dynamodb/
```

## Services

### [🚀 Core Framework](./core/)
Basic framework usage, client setup, and namespace management.

```go
vdf := vasdeference.New(t)
defer vdf.Cleanup()
functionName := vdf.PrefixResourceName("my-function")
```

### [⚡ Lambda Functions](./lambda/)
Test AWS Lambda functions, events, and configurations.

```go
result := lambda.Invoke(ctx, functionName, `{"test": "data"}`)
lambda.AssertInvokeSuccess(ctx, result)
```

### [🗄️ DynamoDB](./dynamodb/)
CRUD operations, transactions, and query patterns.

```go
err := dynamodb.PutItem(ctx, tableName, map[string]interface{}{
    "id": "user1", "name": "John Doe",
})
```

### [📡 EventBridge](./eventbridge/)
Event routing, rules, and custom events.

```go
event := eventbridge.BuildCustomEvent("app.orders", "Order Created", data)
err := eventbridge.PutEvent(ctx, event)
```

### [🔄 Step Functions](./stepfunctions/)
Workflow execution and state machine testing.

```go
execution := stepfunctions.StartExecution(ctx, stateMachine, input)
result := stepfunctions.WaitForCompletion(ctx, execution.Arn, 30*time.Second)
```

## Common Patterns

### Basic Test Setup
```go
func TestMyFunction(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Your test code here...
}
```

### With Custom Configuration
```go
vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDefernOptions{
    Region:    "eu-west-1",
    Namespace: "my-test",
})
```

### Resource Management
```go
resource := createResource()
vdf.RegisterCleanup(func() error {
    return resource.Delete()
})
```

## Run Examples

Each service directory contains runnable examples:

```bash
# Basic usage patterns
go run ./examples/core/examples.go

# Lambda function patterns  
go run ./examples/lambda/examples.go

# DynamoDB patterns
go run ./examples/dynamodb/examples.go

# EventBridge patterns
go run ./examples/eventbridge/examples.go

# Step Functions patterns
go run ./examples/stepfunctions/examples.go
```

## Features

- **Copy-paste ready** code snippets
- **Production patterns** for real applications  
- **Comprehensive examples** covering edge cases
- **Best practices** following TDD principles
- **Error handling** patterns
- **Performance testing** examples

---

**Next:** Choose a service above to see detailed examples and patterns.