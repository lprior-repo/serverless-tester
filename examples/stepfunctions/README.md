# Step Functions

Workflow execution and state machine testing patterns.

## Quick Start

```bash
# Run Step Functions examples
go test -v ./examples/stepfunctions/

# View patterns
go run ./examples/stepfunctions/examples.go
```

## Basic Usage

### Create State Machine
```go
func TestCreateStateMachine(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    stateMachineName := vdf.PrefixResourceName("hello-world")
    
    definition := `{
        "Comment": "A Hello World example",
        "StartAt": "HelloWorld",
        "States": {
            "HelloWorld": {
                "Type": "Pass",
                "Result": "Hello World!",
                "End": true
            }
        }
    }`
    
    stateMachine := stepfunctions.CreateStateMachine(vdf.Context, &stepfunctions.StateMachineConfig{
        Name:       stateMachineName,
        Definition: definition,
        RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
        Type:       "STANDARD",
    })
    
    vdf.RegisterCleanup(func() error {
        return stateMachine.Delete()
    })
    
    stepfunctions.AssertStateMachineExists(vdf.Context, stateMachine.Arn)
}
```

### Execute Workflow
```go
func TestExecuteWorkflow(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:MyWorkflow"
    
    input := stepfunctions.NewInput().
        Set("orderId", "order-123").
        Set("customerId", "customer-456").
        Set("amount", 99.99)
    
    execution := stepfunctions.StartExecution(vdf.Context, stateMachineArn, input)
    result := stepfunctions.WaitForCompletion(vdf.Context, execution.Arn, 5*time.Minute)
    
    stepfunctions.AssertExecutionSucceeded(vdf.Context, result)
    stepfunctions.AssertExecutionOutput(vdf.Context, result, "success")
}
```

### Execute and Wait Pattern
```go
func TestExecuteAndWait(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:DataProcessor"
    
    input := map[string]interface{}{
        "dataset": "customer-data",
        "format":  "json",
    }
    
    result := stepfunctions.ExecuteAndWait(vdf.Context, stateMachineArn, input, &stepfunctions.WaitOptions{
        MaxWaitTime: 10 * time.Minute,
        PollInterval: 5 * time.Second,
    })
    
    stepfunctions.AssertExecutionSucceeded(vdf.Context, result)
    
    var output map[string]interface{}
    stepfunctions.ParseExecutionOutput(result, &output)
    assert.Equal(t, "processed", output["status"])
}
```

## Input Management

### Fluent Input Builder
```go
func TestFluentInputBuilder(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Build input dynamically
    isPriority := true
    hasDiscount := false
    
    input := stepfunctions.NewInput().
        Set("orderId", "order-789").
        Set("customerId", "customer-123").
        Set("items", []string{"item1", "item2", "item3"}).
        SetIf(isPriority, "priority", "high").
        SetIf(hasDiscount, "discount", 10.0).
        SetTimestamp("createdAt", time.Now())
    
    // Convert to JSON for execution
    jsonStr, err := input.ToJSON()
    assert.NoError(t, err)
    assert.Contains(t, jsonStr, "priority")
    assert.NotContains(t, jsonStr, "discount")
}
```

### Input Patterns
```go
func TestInputPatterns(t *testing.T) {
    // Order processing input
    orderInput := stepfunctions.CreateOrderInput(&stepfunctions.OrderData{
        OrderId:    "order-123",
        CustomerId: "customer-456",
        Items: []stepfunctions.OrderItem{
            {SKU: "item-001", Quantity: 2, Price: 29.99},
            {SKU: "item-002", Quantity: 1, Price: 49.99},
        },
        Total: 109.97,
    })
    
    // User workflow input
    userInput := stepfunctions.CreateUserWorkflowInput(&stepfunctions.UserData{
        UserId: "user-789",
        Email:  "user@example.com",
        Plan:   "premium",
        Metadata: map[string]interface{}{
            "source":      "web",
            "campaign":    "spring-2024",
            "referralId":  "ref-123",
        },
    })
    
    // Retry configuration input
    retryInput := stepfunctions.CreateRetryableInput(&stepfunctions.RetryConfig{
        Operation:   "data-processing",
        MaxAttempts: 3,
        BackoffRate: 2.0,
        Payload: map[string]interface{}{
            "bucket": "data-bucket",
            "key":    "batch-001.csv",
        },
    })
}
```

## Execution Analysis

### History Analysis
```go
func TestExecutionHistory(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    executionArn := "arn:aws:states:us-east-1:123456789012:execution:MyWorkflow:exec-123"
    
    // Get execution history
    history := stepfunctions.GetExecutionHistory(vdf.Context, executionArn)
    
    // Analyze execution
    analysis := stepfunctions.AnalyzeExecutionHistory(history)
    
    assert.Greater(t, analysis.TotalSteps, 0)
    assert.Equal(t, "SUCCEEDED", analysis.ExecutionStatus)
    assert.True(t, analysis.ExecutionTime > 0)
    
    // Find specific events
    startEvents := stepfunctions.FindEventsByType(history, "ExecutionStarted")
    assert.Len(t, startEvents, 1)
    
    taskEvents := stepfunctions.FindEventsByType(history, "TaskStateExited")
    assert.Greater(t, len(taskEvents), 0)
}
```

### Failed Step Analysis
```go
func TestFailedStepAnalysis(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    executionArn := "arn:aws:states:us-east-1:123456789012:execution:FailedWorkflow:exec-456"
    
    history := stepfunctions.GetExecutionHistory(vdf.Context, executionArn)
    failedSteps := stepfunctions.FindFailedSteps(history)
    
    if len(failedSteps) > 0 {
        for _, step := range failedSteps {
            t.Logf("Failed Step: %s", step.StepName)
            t.Logf("Error: %s", step.Error)
            t.Logf("Cause: %s", step.Cause)
            
            // Analyze failure patterns
            if stepfunctions.IsTimeoutError(step) {
                t.Log("Failure was due to timeout")
            } else if stepfunctions.IsResourceError(step) {
                t.Log("Failure was due to resource unavailability")
            }
        }
    }
    
    // Create failure report
    report := stepfunctions.CreateFailureReport(failedSteps)
    t.Logf("Failure Report:\n%s", report)
}
```

### Execution Timeline
```go
func TestExecutionTimeline(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    executionArn := "arn:aws:states:us-east-1:123456789012:execution:TimingWorkflow:exec-789"
    
    history := stepfunctions.GetExecutionHistory(vdf.Context, executionArn)
    timeline := stepfunctions.CreateExecutionTimeline(history)
    
    for _, event := range timeline {
        t.Logf("%s: %s - Duration: %v",
            event.Timestamp.Format(time.RFC3339),
            event.StateType,
            event.Duration,
        )
    }
    
    // Performance analysis
    totalDuration := stepfunctions.CalculateTotalDuration(timeline)
    longestStep := stepfunctions.FindLongestStep(timeline)
    
    assert.True(t, totalDuration > 0)
    assert.NotNil(t, longestStep)
    
    t.Logf("Total execution time: %v", totalDuration)
    t.Logf("Longest step: %s (%v)", longestStep.StateName, longestStep.Duration)
}
```

## Advanced Patterns

### Complex Workflow
```go
func TestComplexWorkflow(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    stateMachineName := vdf.PrefixResourceName("order-processing")
    
    definition := `{
        "Comment": "Order processing workflow",
        "StartAt": "ValidateOrder",
        "States": {
            "ValidateOrder": {
                "Type": "Task",
                "Resource": "arn:aws:lambda:us-east-1:123456789012:function:ValidateOrder",
                "Next": "CheckInventory",
                "Catch": [
                    {
                        "ErrorEquals": ["ValidationError"],
                        "Next": "OrderValidationFailed"
                    }
                ],
                "Retry": [
                    {
                        "ErrorEquals": ["Lambda.ServiceException"],
                        "IntervalSeconds": 2,
                        "MaxAttempts": 3,
                        "BackoffRate": 2.0
                    }
                ]
            },
            "CheckInventory": {
                "Type": "Task", 
                "Resource": "arn:aws:lambda:us-east-1:123456789012:function:CheckInventory",
                "Next": "ProcessPayment"
            },
            "ProcessPayment": {
                "Type": "Task",
                "Resource": "arn:aws:lambda:us-east-1:123456789012:function:ProcessPayment", 
                "End": true
            },
            "OrderValidationFailed": {
                "Type": "Fail",
                "Error": "ValidationError",
                "Cause": "Order validation failed"
            }
        }
    }`
    
    stateMachine := stepfunctions.CreateStateMachine(vdf.Context, &stepfunctions.StateMachineConfig{
        Name:       stateMachineName,
        Definition: definition,
        RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
    })
    
    vdf.RegisterCleanup(func() error {
        return stateMachine.Delete()
    })
    
    // Test successful execution
    input := stepfunctions.CreateOrderInput(&stepfunctions.OrderData{
        OrderId: "order-123",
        Items: []stepfunctions.OrderItem{
            {SKU: "item-001", Quantity: 1, Price: 99.99},
        },
    })
    
    result := stepfunctions.ExecuteAndWait(vdf.Context, stateMachine.Arn, input, &stepfunctions.WaitOptions{
        MaxWaitTime: 5 * time.Minute,
    })
    
    stepfunctions.AssertExecutionSucceeded(vdf.Context, result)
}
```

### Error Handling
```go
func TestErrorHandling(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:ErrorTest"
    
    // Test with invalid input that should cause failure
    invalidInput := map[string]interface{}{
        "amount": -100, // Invalid negative amount
    }
    
    result := stepfunctions.ExecuteAndWait(vdf.Context, stateMachineArn, invalidInput, &stepfunctions.WaitOptions{
        MaxWaitTime: 2 * time.Minute,
    })
    
    stepfunctions.AssertExecutionFailed(vdf.Context, result)
    stepfunctions.AssertExecutionError(vdf.Context, result, "ValidationError")
    
    // Analyze the failure
    history := stepfunctions.GetExecutionHistory(vdf.Context, result.ExecutionArn)
    failedSteps := stepfunctions.FindFailedSteps(history)
    
    assert.Greater(t, len(failedSteps), 0)
    assert.Contains(t, failedSteps[0].Error, "ValidationError")
}
```

### Custom Polling Configuration
```go
func TestCustomPolling(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:LongRunning"
    
    input := map[string]interface{}{
        "task": "heavy-computation",
    }
    
    // Use exponential backoff for long-running workflows
    waitOptions := &stepfunctions.WaitOptions{
        MaxWaitTime:        20 * time.Minute,
        InitialInterval:    5 * time.Second,
        MaxInterval:        60 * time.Second,
        ExponentialBackoff: true,
        BackoffMultiplier:  1.5,
    }
    
    result := stepfunctions.ExecuteAndWait(vdf.Context, stateMachineArn, input, waitOptions)
    stepfunctions.AssertExecutionSucceeded(vdf.Context, result)
}
```

## Assertions

```go
func TestStepFunctionsAssertions(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:TestWorkflow"
    executionArn := "arn:aws:states:us-east-1:123456789012:execution:TestWorkflow:exec-123"
    
    // State machine assertions
    stepfunctions.AssertStateMachineExists(vdf.Context, stateMachineArn)
    stepfunctions.AssertStateMachineStatus(vdf.Context, stateMachineArn, "ACTIVE")
    stepfunctions.AssertStateMachineType(vdf.Context, stateMachineArn, "STANDARD")
    
    // Execution assertions
    result := stepfunctions.GetExecutionResult(vdf.Context, executionArn)
    stepfunctions.AssertExecutionSucceeded(vdf.Context, result)
    stepfunctions.AssertExecutionTime(vdf.Context, result, time.Second, 5*time.Minute)
    
    // Output assertions
    stepfunctions.AssertExecutionOutput(vdf.Context, result, "completed")
    stepfunctions.AssertExecutionOutputContains(vdf.Context, result, "success")
    
    // JSON output assertions
    expectedOutput := map[string]interface{}{
        "status":         "completed",
        "itemsProcessed": 100.0,
    }
    stepfunctions.AssertExecutionOutputJSON(vdf.Context, result, expectedOutput)
    
    // Pattern-based assertions
    pattern := &stepfunctions.ExecutionPattern{
        Status:         "SUCCEEDED",
        OutputContains: "completed",
        MinDuration:    time.Second,
        MaxDuration:    time.Minute,
    }
    stepfunctions.AssertExecutionPattern(vdf.Context, result, pattern)
}
```

## Run Examples

```bash
# View all patterns
go run ./examples/stepfunctions/examples.go

# Run as tests  
go test -v ./examples/stepfunctions/

# Test specific patterns
go test -v ./examples/stepfunctions/ -run TestExecuteWorkflow
go test -v ./examples/stepfunctions/ -run TestExecutionHistory
```

## API Reference

### State Machine Management
- `stepfunctions.CreateStateMachine(ctx, config)` - Create state machine
- `stepfunctions.DeleteStateMachine(ctx, arn)` - Delete state machine
- `stepfunctions.AssertStateMachineExists(ctx, arn)` - Assert state machine exists

### Execution Management
- `stepfunctions.StartExecution(ctx, stateMachineArn, input)` - Start execution
- `stepfunctions.WaitForCompletion(ctx, executionArn, timeout)` - Wait for completion
- `stepfunctions.ExecuteAndWait(ctx, stateMachineArn, input, options)` - Execute and wait

### Input Management
- `stepfunctions.NewInput()` - Create fluent input builder
- `stepfunctions.CreateOrderInput(orderData)` - Create order processing input
- `stepfunctions.CreateUserWorkflowInput(userData)` - Create user workflow input

### History Analysis
- `stepfunctions.GetExecutionHistory(ctx, executionArn)` - Get execution history
- `stepfunctions.AnalyzeExecutionHistory(history)` - Analyze execution
- `stepfunctions.FindFailedSteps(history)` - Find failed steps
- `stepfunctions.CreateExecutionTimeline(history)` - Create timeline

### Assertions
- `stepfunctions.AssertExecutionSucceeded(ctx, result)` - Assert success
- `stepfunctions.AssertExecutionFailed(ctx, result)` - Assert failure
- `stepfunctions.AssertExecutionOutput(ctx, result, expected)` - Assert output
- `stepfunctions.AssertExecutionPattern(ctx, result, pattern)` - Assert pattern

---

**Next:** [Lambda Examples](../lambda/) | [EventBridge Examples](../eventbridge/)