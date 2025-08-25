# Lambda Functions

Test AWS Lambda functions with events, configurations, and performance patterns.

## Quick Start

```bash
# Run Lambda examples
go test -v ./examples/lambda/

# Run specific patterns  
go test -v ./examples/lambda/ -run TestExampleUsage
```

## Basic Usage

### Simple Function Invocation
```go
func TestLambdaInvocation(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("my-function")
    
    result := lambda.Invoke(ctx, functionName, `{"message": "hello"}`)
    lambda.AssertInvokeSuccess(ctx, result)
    lambda.AssertPayloadContains(ctx, result, "success")
}
```

### Function Configuration
```go
func TestLambdaConfig(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("config-test")
    
    lambda.AssertFunctionRuntime(ctx, functionName, types.RuntimeNodejs18x)
    lambda.AssertFunctionTimeout(ctx, functionName, 30)
    lambda.AssertEnvVarEquals(ctx, functionName, "NODE_ENV", "production")
}
```

## Event Processing

### S3 Events
```go
func TestS3EventProcessing(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("s3-processor")
    
    s3Event := lambda.BuildS3Event("my-bucket", "data/file.json", "s3:ObjectCreated:Put")
    result := lambda.Invoke(ctx, functionName, s3Event)
    
    lambda.AssertInvokeSuccess(ctx, result)
    lambda.AssertPayloadContains(ctx, result, "processed")
}
```

### DynamoDB Events
```go
func TestDynamoDBEvents(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("dynamo-processor")
    
    keys := map[string]interface{}{
        "id": map[string]interface{}{"S": "item-123"},
    }
    
    dynamoEvent := lambda.BuildDynamoDBEvent("users", "INSERT", keys)
    result := lambda.Invoke(ctx, functionName, dynamoEvent)
    
    lambda.AssertInvokeSuccess(ctx, result)
}
```

### API Gateway Events
```go
func TestAPIGatewayEvents(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("api-handler")
    
    apiEvent := lambda.BuildAPIGatewayEvent("POST", "/users", `{"name": "John"}`)
    result := lambda.Invoke(ctx, functionName, apiEvent)
    
    lambda.AssertInvokeSuccess(ctx, result)
    
    var response map[string]interface{}
    lambda.ParseInvokeOutput(result, &response)
    assert.Equal(t, 200, response["statusCode"])
}
```

## Advanced Patterns

### Async Invocation
```go
func TestAsyncInvocation(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("async-processor")
    
    result := lambda.InvokeAsync(ctx, functionName, `{"data": "async-test"}`)
    assert.Equal(t, 202, result.StatusCode) // Async returns 202
    
    // Wait for logs
    logs := lambda.WaitForLogs(ctx, functionName, "Processing completed", 30*time.Second)
    assert.NotEmpty(t, logs)
}
```

### Retry Logic
```go
func TestRetryLogic(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("retry-function")
    
    result := lambda.InvokeWithRetry(ctx, functionName, `{"test": "retry"}`, 3)
    lambda.AssertInvokeSuccess(ctx, result)
}
```

### Performance Testing
```go
func TestConcurrency(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("load-test")
    
    concurrency := 10
    results := make(chan *lambda.InvokeResult, concurrency)
    
    for i := 0; i < concurrency; i++ {
        go func() {
            result := lambda.Invoke(ctx, functionName, `{"load": true}`)
            results <- result
        }()
    }
    
    successCount := 0
    for i := 0; i < concurrency; i++ {
        result := <-results
        if result.StatusCode == 200 {
            successCount++
        }
    }
    
    assert.Equal(t, concurrency, successCount)
}
```

## Table-Driven Testing

```go
func TestTableDriven(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("validator")
    
    testCases := []struct {
        name          string
        payload       string
        expectSuccess bool
        expectedText  string
    }{
        {
            name:          "ValidPayload",
            payload:       `{"name": "John", "age": 30}`,
            expectSuccess: true,
            expectedText:  "valid",
        },
        {
            name:          "InvalidPayload", 
            payload:       `{"name": ""}`,
            expectSuccess: false,
            expectedText:  "error",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := lambda.Invoke(ctx, functionName, tc.payload)
            
            if tc.expectSuccess {
                lambda.AssertInvokeSuccess(ctx, result)
            } else {
                lambda.AssertInvokeError(ctx, result)
            }
            
            lambda.AssertPayloadContains(ctx, result, tc.expectedText)
        })
    }
}
```

## Error Handling

```go
func TestErrorScenarios(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    ctx := &lambda.TestContext{T: t}
    functionName := vdf.PrefixResourceName("error-test")
    
    // Test function error
    result := lambda.Invoke(ctx, functionName, `{"trigger_error": true}`)
    lambda.AssertInvokeError(ctx, result)
    
    // Verify error logs
    lambda.AssertLogsContainLevel(ctx, functionName, "ERROR")
    lambda.AssertLogsContain(ctx, functionName, "triggered error")
}
```

## Run Examples

```bash
# View all patterns
go run ./examples/lambda/examples.go

# Run as tests
go test -v ./examples/lambda/

# Test specific functions  
go test -v ./examples/lambda/ -run TestExampleUsage
go test -v ./examples/lambda/ -run TestExampleTableDrivenTesting
```

## API Reference

### Invoke Functions
- `lambda.Invoke(ctx, functionName, payload)` - Synchronous invoke
- `lambda.InvokeAsync(ctx, functionName, payload)` - Asynchronous invoke  
- `lambda.InvokeWithRetry(ctx, functionName, payload, maxAttempts)` - With retry

### Event Builders
- `lambda.BuildS3Event(bucket, key, eventName)` - S3 events
- `lambda.BuildDynamoDBEvent(table, eventName, keys)` - DynamoDB events
- `lambda.BuildAPIGatewayEvent(method, path, body)` - API Gateway events
- `lambda.BuildSQSEvent(queueUrl, messageBody)` - SQS events

### Assertions
- `lambda.AssertInvokeSuccess(ctx, result)` - Success validation
- `lambda.AssertInvokeError(ctx, result)` - Error validation
- `lambda.AssertPayloadContains(ctx, result, text)` - Payload content
- `lambda.AssertLogsContain(ctx, functionName, text)` - Log content

---

**Next:** [DynamoDB Examples](../dynamodb/) | [EventBridge Examples](../eventbridge/)