# 🚀 AWS Lambda Functional Testing Package

A **pure functional programming** AWS Lambda testing package for the `vasdeference` framework. Built with immutable data structures, monadic error handling, and type-safe operations using `samber/lo` and `samber/mo`.

## ✨ Functional Programming Features

🔒 **Type-Safe Operations** with Go generics and monadic types  
🚀 **Pure Functions** with no mutations or side effects  
📦 **Immutable Configurations** using functional options pattern  
⚡ **Monadic Error Handling** with `Option[T]` and `Result[T]`  
🔧 **Function Composition** with chainable operations  
🧮 **Mathematical Precision** - predictable, verifiable, composable  

## 🏗️ Architecture

- **Pure Functional Core**: All functions are pure with predictable inputs/outputs
- **Immutable Data Structures**: Zero mutations, only transformations
- **Monadic Operations**: Safe error handling without exceptions
- **Type Safety**: Compile-time guarantees with Go generics
- **Function Composition**: Elegant pipelines for complex operations
- **Zero Side Effects**: All effects isolated to result types

## 🔧 Functional Operations

### Pure Lambda Functions
```go
// Immutable configuration with functional options
config := NewFunctionalLambdaConfig(
    WithFunctionName("my-processor"),
    WithRuntime("nodejs22.x"),
    WithTimeout(30*time.Second),
    WithMemorySize(512),
)

// Type-safe function creation with monadic result
result := SimulateFunctionalLambdaCreateFunction(config)
if result.IsSuccess() {
    t.Log("✓ Function created successfully")
}
```

### Monadic Invocation Operations
```go
// Immutable invocation payload
payload := NewFunctionalLambdaInvocationPayload(`{"test": "data"}`)

// Invoke with functional composition
invokeResult := SimulateFunctionalLambdaInvoke(payload, config)

// Safe monadic error handling
invokeResult.GetError().
    Map(func(err error) error {
        t.Errorf("Invocation failed: %v", err)
        return err
    }).
    OrElse(func() {
        t.Log("✓ Lambda invoked successfully")
    })
```

### Function Composition Pipeline
```go
// Compose multiple operations into a pipeline
pipeline := ComposeSimulationOperations(
    CreateFunctionOperation(config),
    InvokeFunctionOperation(payload),
    ValidateResponseOperation(),
    CleanupResourcesOperation(),
)

// Execute pipeline with single result
pipelineResult := ExecuteFunctionalLambdaPipeline(pipeline)
```
- `GetLogStats` - Log statistics and analysis

### Event Source Mappings
- `CreateEventSourceMapping/CreateEventSourceMappingE` - Create event mappings
- `GetEventSourceMapping/GetEventSourceMappingE` - Retrieve mapping configuration
- `UpdateEventSourceMapping/UpdateEventSourceMappingE` - Update mappings
- `DeleteEventSourceMapping/DeleteEventSourceMappingE` - Delete mappings
- `ListEventSourceMappings/ListEventSourceMappingsE` - List all mappings
- `WaitForEventSourceMappingState/WaitForEventSourceMappingStateE` - Wait for mapping state

### Assertions
- `AssertFunctionExists/AssertFunctionDoesNotExist` - Function existence assertions
- `AssertFunctionRuntime/AssertFunctionHandler` - Function configuration assertions
- `AssertInvokeSuccess/AssertInvokeError` - Invocation result assertions
- `AssertPayloadContains/AssertPayloadEquals` - Payload content assertions
- `AssertLogsContain/AssertLogsContainLevel` - Log content assertions
- `AssertExecutionTimeLessThan` - Performance assertions

## Event Builders

### AWS Service Event Patterns
- `BuildS3Event/BuildS3EventE` - S3 bucket events (ObjectCreated, ObjectRemoved, etc.)
- `BuildDynamoDBEvent/BuildDynamoDBEventE` - DynamoDB stream events
- `BuildSQSEvent/BuildSQSEventE` - SQS message events
- `BuildSNSEvent/BuildSNSEventE` - SNS notification events
- `BuildAPIGatewayEvent/BuildAPIGatewayEventE` - API Gateway request events

### Event Utilities
- `ParseS3Event/ParseS3EventE` - Parse S3 events from JSON
- `ExtractS3BucketName/ExtractS3ObjectKey` - Extract S3 event details
- `AddS3EventRecord/AddS3EventRecordE` - Add records to existing events
- `ValidateEventStructure` - Validate event JSON structure

## JSON Utilities

### Payload Handling
- `MarshalPayload/MarshalPayloadE` - Marshal objects to JSON payloads
- `ParseInvokeOutput/ParseInvokeOutputE` - Parse Lambda response payloads

### Validation
- `ValidateFunctionConfiguration` - Comprehensive function config validation
- `ValidateInvokeResult` - Invocation result validation
- `ValidateLogEntries` - Log entry validation

## Usage Examples

### Basic Lambda Invocation

```go
func TestLambdaFunction(t *testing.T) {
    // Create test context
    ctx := sfx.NewTestContext(t)
    
    // Invoke Lambda function
    result := lambda.Invoke(ctx, "my-function", `{"key": "value"}`)
    
    // Assert results
    lambda.AssertInvokeSuccess(ctx, result)
    lambda.AssertPayloadContains(ctx, result, "success")
}
```

### Event-Driven Testing

```go
func TestS3EventProcessing(t *testing.T) {
    ctx := sfx.NewTestContext(t)
    
    // Build S3 event
    s3Event := lambda.BuildS3Event("my-bucket", "data/file.json", "s3:ObjectCreated:Put")
    
    // Invoke with event
    result := lambda.Invoke(ctx, "s3-processor", s3Event)
    lambda.AssertInvokeSuccess(ctx, result)
    
    // Verify logs
    lambda.AssertLogsContain(ctx, "s3-processor", "Processing S3 object")
}
```

### Table-Driven Testing

```go
func TestMultiplePayloads(t *testing.T) {
    ctx := sfx.NewTestContext(t)
    
    tests := []struct {
        name           string
        payload        string
        expectSuccess  bool
        expectedInPayload string
    }{
        {
            name:           "ValidPayload",
            payload:        `{"name": "John", "age": 30}`,
            expectSuccess:  true,
            expectedInPayload: "processed",
        },
        {
            name:           "InvalidPayload",
            payload:        `{"name": ""}`,
            expectSuccess:  false,
            expectedInPayload: "validation_error",
        },
    }
    
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            result := lambda.Invoke(ctx, "validator-function", tc.payload)
            
            if tc.expectSuccess {
                lambda.AssertInvokeSuccess(ctx, result)
            } else {
                lambda.AssertInvokeError(ctx, result)
            }
            
            lambda.AssertPayloadContains(ctx, result, tc.expectedInPayload)
        })
    }
}
```

### Event Source Mapping

```go
func TestEventSourceMapping(t *testing.T) {
    ctx := sfx.NewTestContext(t)
    arrange := sfx.NewArrange(ctx)
    defer arrange.Cleanup()
    
    // Create event source mapping
    config := lambda.EventSourceMappingConfig{
        EventSourceArn:   "arn:aws:sqs:us-east-1:123456789012:my-queue",
        FunctionName:     "queue-processor",
        BatchSize:        10,
        Enabled:          true,
        StartingPosition: types.EventSourcePositionLatest,
    }
    
    mapping := lambda.CreateEventSourceMapping(ctx, config)
    arrange.RegisterCleanup(func() error {
        return lambda.DeleteEventSourceMappingE(ctx, mapping.UUID)
    })
    
    // Wait for mapping to be active
    lambda.WaitForEventSourceMappingState(ctx, mapping.UUID, "Enabled", 2*time.Minute)
}
```

### Performance Testing

```go
func TestConcurrentInvocations(t *testing.T) {
    ctx := sfx.NewTestContext(t)
    
    concurrency := 10
    results := make(chan *lambda.InvokeResult, concurrency)
    
    // Launch concurrent invocations
    for i := 0; i < concurrency; i++ {
        go func() {
            result := lambda.Invoke(ctx, "concurrent-function", `{"test": true}`)
            results <- result
        }()
    }
    
    // Collect and validate results
    successCount := 0
    for i := 0; i < concurrency; i++ {
        result := <-results
        if result.StatusCode == 200 {
            successCount++
        }
    }
    
    assert.Equal(t, concurrency, successCount, "All invocations should succeed")
}
```

## Test Organization

### Recommended Structure

```go
func TestLambdaWorkflow(t *testing.T) {
    // Group 1: Unit tests for individual functions
    t.Run("UnitTests", func(t *testing.T) {
        t.Run("PayloadValidation", func(t *testing.T) {
            // Test payload validation logic
        })
        
        t.Run("ErrorHandling", func(t *testing.T) {
            // Test error handling scenarios
        })
    })
    
    // Group 2: Integration tests
    t.Run("IntegrationTests", func(t *testing.T) {
        t.Run("SingleFunction", func(t *testing.T) {
            // Test individual Lambda function
        })
        
        t.Run("WorkflowChain", func(t *testing.T) {
            // Test chain of Lambda functions
        })
    })
    
    // Group 3: End-to-end tests
    t.Run("EndToEndTests", func(t *testing.T) {
        t.Run("CompleteWorkflow", func(t *testing.T) {
            // Test complete serverless workflow
        })
    })
}
```

## Architecture

This package follows strict functional programming principles:

- **Pure Functions**: All business logic implemented as pure functions
- **Immutable Data**: All data structures are immutable
- **Function Composition**: Complex behavior built through composition
- **Error Handling**: Consistent error handling with `Function/FunctionE` patterns
- **Testability**: Every function is easily testable in isolation

## Dependencies

- AWS SDK for Go v2 (Lambda and CloudWatch Logs services)
- Testify for test assertions and utilities
- Standard Go testing package compatibility

## Integration

This package integrates seamlessly with:

- **vasdeference core module**: Uses shared test context and arrangement patterns
- **Terratest**: Drop-in replacement with enhanced functionality  
- **Standard Go testing**: Compatible with `go test` and all testing frameworks
- **CI/CD pipelines**: Designed for automated testing environments

## Best Practices

1. **Always use test contexts**: Create proper test contexts for AWS configuration
2. **Resource cleanup**: Use arrangement patterns for automatic resource cleanup
3. **Table-driven tests**: Leverage table-driven patterns for comprehensive coverage
4. **Retry logic**: Use built-in retry mechanisms for eventual consistency
5. **Log validation**: Always validate logs for comprehensive testing
6. **Event patterns**: Use event builders for consistent test data
7. **Performance testing**: Include concurrency and timeout testing
8. **Error scenarios**: Test both success and failure cases thoroughly

This package provides everything needed to build comprehensive, maintainable, and reliable tests for AWS Lambda functions and serverless workflows.