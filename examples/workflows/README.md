# Advanced Workflow Examples

Comprehensive workflow orchestration testing using the enhanced VasDeference framework with parallel Step Functions execution, complex state machines, and error handling patterns.

## Overview

These examples demonstrate advanced workflow testing patterns for complex serverless orchestrations:

- **Parallel Workflow Execution**: Concurrent Step Functions with synchronization
- **Complex State Machines**: Multi-branch workflows with conditional logic  
- **Error Handling & Recovery**: Comprehensive failure scenarios and recovery testing
- **Workflow Chaining**: Sequential and parallel workflow composition
- **Long-Running Workflows**: Extended duration workflow testing
- **Event-Driven Workflows**: EventBridge integration with Step Functions

## Examples

### 1. Parallel Workflow Orchestration  
**File**: `parallel_workflow_test.go`

Advanced parallel workflow testing:
- Concurrent Step Functions execution
- Workflow synchronization patterns
- Resource contention handling
- Performance optimization
- Batch workflow processing

### 2. Complex State Machine Testing
**File**: `complex_statemachine_test.go`

Comprehensive state machine validation:
- Multi-branch conditional workflows
- Nested state machine testing
- Choice state validation
- Parallel state coordination
- Map state processing

### 3. Error Handling & Recovery
**File**: `error_recovery_test.go`

Robust error handling patterns:
- Retry mechanism testing
- Circuit breaker patterns
- Dead letter queue handling
- Workflow rollback scenarios
- Partial failure recovery

### 4. Event-Driven Workflow Testing
**File**: `event_driven_workflow_test.go`

Event integration patterns:
- EventBridge trigger validation
- Event pattern matching
- Cross-region event handling
- Event ordering and correlation
- Event replay testing

### 5. Long-Running Workflow Testing
**File**: `long_running_workflow_test.go`

Extended workflow validation:
- Multi-hour workflow testing
- State persistence validation
- Timeout handling
- Resource optimization
- Cost optimization patterns

## Key Features

### Parallel Workflow Execution
```go
// Advanced parallel workflow orchestration
executions := []stepfunctions.ParallelExecution{
    {
        StateMachineArn: orderProcessingWorkflow,
        ExecutionName:   "order-processing-1",
        Input:          orderData1,
    },
    {
        StateMachineArn: inventoryWorkflow,
        ExecutionName:   "inventory-update-1", 
        Input:          inventoryData1,
    },
    {
        StateMachineArn: paymentWorkflow,
        ExecutionName:   "payment-processing-1",
        Input:          paymentData1,
    },
}

results := stepfunctions.ParallelExecutions(ctx, executions, &stepfunctions.ParallelExecutionConfig{
    MaxConcurrency:    10,
    WaitForCompletion: true,
    FailFast:         false,
    Timeout:          300 * time.Second,
})

// Validate all executions completed successfully
for _, result := range results {
    assert.Equal(t, stepfunctions.ExecutionStatusSucceeded, result.Status)
}
```

### Workflow Chain Testing
```go
// Sequential workflow chain execution
chain := stepfunctions.NewWorkflowChain(ctx).
    AddWorkflow("data-ingestion", dataIngestionSM, ingestionInput).
    AddWorkflow("data-processing", dataProcessingSM, nil). // Uses output from previous
    AddWorkflow("data-validation", dataValidationSM, nil).
    AddWorkflow("data-publishing", dataPublishingSM, publishConfig)

result := chain.Execute(stepfunctions.WorkflowChainConfig{
    StopOnError:    true,
    MaxRetries:     3,
    RetryDelay:     30 * time.Second,
})

assert.True(t, result.Success)
assert.Equal(t, 4, len(result.CompletedWorkflows))
```

### Complex Error Handling
```go
// Comprehensive error handling and recovery testing
errorScenarios := []ErrorScenario{
    {
        Name:         "lambda_timeout",
        InjectAt:     "data-processing",
        ErrorType:    "States.Timeout",
        ExpectedRetry: 3,
    },
    {
        Name:         "service_unavailable",
        InjectAt:     "external-api-call",
        ErrorType:    "States.TaskFailed",
        ExpectedFallback: "fallback-handler",
    },
    {
        Name:         "invalid_input",
        InjectAt:     "input-validation",
        ErrorType:    "States.InputPathFailed",
        ExpectedTermination: true,
    },
}

for _, scenario := range errorScenarios {
    t.Run(scenario.Name, func(t *testing.T) {
        // Inject error and validate recovery behavior
        result := executeWorkflowWithErrorInjection(t, ctx, workflow, scenario)
        validateErrorRecovery(t, result, scenario)
    })
}
```

### Event-Driven Workflow Integration
```go
// Event-driven workflow testing with EventBridge
func TestEventDrivenWorkflowIntegration(t *testing.T) {
    // Setup event capture
    eventCapture := eventbridge.CreateEventCapture(t, eventBusName)
    defer eventCapture.Stop()
    
    // Publish trigger event
    triggerEvent := eventbridge.BuildCustomEvent("workflow.trigger", "WorkflowStartRequested", triggerData)
    err := eventbridge.PutEvent(ctx, triggerEvent)
    require.NoError(t, err)
    
    // Wait for workflow execution to start
    execution := waitForWorkflowTriggered(t, ctx, workflowSM, 30*time.Second)
    require.NotNil(t, execution)
    
    // Wait for completion and validate
    result := stepfunctions.WaitForCompletion(ctx, execution.ExecutionArn, 120*time.Second)
    assert.Equal(t, stepfunctions.ExecutionStatusSucceeded, result.Status)
    
    // Validate event correlation
    events := eventCapture.GetEvents()
    validateEventCorrelation(t, triggerEvent, events)
}
```

## Running Workflow Tests

### Prerequisites
```bash
# Set workflow test configuration
export WORKFLOW_TEST_TIMEOUT=1800s
export MAX_CONCURRENT_WORKFLOWS=20
export ENABLE_LONG_RUNNING_TESTS=true

# Configure Step Functions permissions
export STEP_FUNCTIONS_ROLE_ARN=arn:aws:iam::123456789012:role/StepFunctionsWorkflowRole
export WORKFLOW_LOG_LEVEL=ALL

# Set error injection configuration
export ENABLE_ERROR_INJECTION=true
export ERROR_INJECTION_RATE=0.1
```

### Execute Tests
```bash
# Run all workflow tests
go test -v -timeout=3600s ./examples/workflows/

# Run parallel workflow tests
go test -v ./examples/workflows/ -run TestParallelWorkflow

# Run error handling tests
go test -v ./examples/workflows/ -run TestErrorHandling

# Run long-running workflow tests
LONG_RUNNING=true go test -v -timeout=7200s ./examples/workflows/ -run TestLongRunning

# Run with error injection
ERROR_INJECTION=true go test -v ./examples/workflows/ -run TestErrorRecovery
```

### Performance Testing
```bash
# Workflow performance benchmarks
go test -v -bench=BenchmarkWorkflow ./examples/workflows/

# High-concurrency workflow testing
CONCURRENCY=50 go test -v ./examples/workflows/ -run TestWorkflowConcurrency

# Memory usage profiling
go test -v -benchmem -cpuprofile=workflow.prof ./examples/workflows/
```

## Workflow Patterns

### Map State Testing
```go
// Test Map state for batch processing
func TestMapStateProcessing(t *testing.T) {
    // Create batch of items to process
    batchItems := generateBatchItems(100)
    
    mapInput := stepfunctions.MapStateInput{
        Items: batchItems,
        MaxConcurrency: 10,
        ItemsPath: "$.items",
        ItemProcessor: itemProcessorStateMachine,
    }
    
    execution := stepfunctions.StartExecution(ctx, mapStateMachine, mapInput)
    result := stepfunctions.WaitForCompletion(ctx, execution.ExecutionArn, 300*time.Second)
    
    // Validate all items were processed
    assert.Equal(t, stepfunctions.ExecutionStatusSucceeded, result.Status)
    
    output := parseMapStateOutput(result.Output)
    assert.Equal(t, len(batchItems), len(output.ProcessedItems))
    
    // Validate individual item processing
    for i, processedItem := range output.ProcessedItems {
        validateItemProcessing(t, batchItems[i], processedItem)
    }
}
```

### Choice State Logic
```go
// Test complex Choice state logic
func TestChoiceStateLogic(t *testing.T) {
    choiceTestCases := []struct {
        name           string
        input          map[string]interface{}
        expectedPath   string
        expectedOutput interface{}
    }{
        {
            name:  "high_priority_order",
            input: map[string]interface{}{"priority": "high", "amount": 1000},
            expectedPath: "HighPriorityProcessing",
            expectedOutput: "priority_processed",
        },
        {
            name:  "bulk_order",
            input: map[string]interface{}{"type": "bulk", "quantity": 500},
            expectedPath: "BulkProcessing", 
            expectedOutput: "bulk_processed",
        },
        {
            name:  "standard_order",
            input: map[string]interface{}{"type": "standard", "amount": 100},
            expectedPath: "StandardProcessing",
            expectedOutput: "standard_processed",
        },
    }
    
    for _, tc := range choiceTestCases {
        t.Run(tc.name, func(t *testing.T) {
            execution := stepfunctions.StartExecution(ctx, choiceStateMachine, tc.input)
            result := stepfunctions.WaitForCompletion(ctx, execution.ExecutionArn, 60*time.Second)
            
            // Validate correct path was taken
            assert.Equal(t, stepfunctions.ExecutionStatusSucceeded, result.Status)
            
            // Validate execution history shows correct path
            history := stepfunctions.GetExecutionHistory(ctx, execution.ExecutionArn)
            validateChoiceStatePath(t, history, tc.expectedPath)
            
            // Validate output
            validateWorkflowOutput(t, result.Output, tc.expectedOutput)
        })
    }
}
```

### Parallel State Coordination
```go
// Test Parallel state with multiple branches
func TestParallelStateBranches(t *testing.T) {
    parallelInput := map[string]interface{}{
        "orderId": "order-123",
        "items": []map[string]interface{}{
            {"productId": "prod-1", "quantity": 2},
            {"productId": "prod-2", "quantity": 1},
        },
        "customer": map[string]interface{}{
            "customerId": "cust-456",
            "email": "customer@example.com",
        },
    }
    
    execution := stepfunctions.StartExecution(ctx, parallelStateMachine, parallelInput)
    result := stepfunctions.WaitForCompletion(ctx, execution.ExecutionArn, 180*time.Second)
    
    assert.Equal(t, stepfunctions.ExecutionStatusSucceeded, result.Status)
    
    // Validate all parallel branches completed
    history := stepfunctions.GetExecutionHistory(ctx, execution.ExecutionArn)
    branches := extractParallelBranches(history)
    
    expectedBranches := []string{
        "InventoryReservation",
        "PaymentProcessing", 
        "CustomerNotification",
        "ShippingPreparation",
    }
    
    assert.Equal(t, len(expectedBranches), len(branches))
    
    for _, expectedBranch := range expectedBranches {
        found := false
        for _, branch := range branches {
            if branch.Name == expectedBranch {
                assert.Equal(t, "SUCCEEDED", branch.Status)
                found = true
                break
            }
        }
        assert.True(t, found, "Branch %s should be executed", expectedBranch)
    }
}
```

### Wait State Testing
```go
// Test Wait state with different timing strategies
func TestWaitStateStrategies(t *testing.T) {
    waitStrategies := []struct {
        name        string
        waitType    string
        waitValue   interface{}
        expectedMin time.Duration
        expectedMax time.Duration
    }{
        {
            name:        "seconds_wait",
            waitType:    "Seconds",
            waitValue:   5,
            expectedMin: 5 * time.Second,
            expectedMax: 6 * time.Second,
        },
        {
            name:        "timestamp_wait", 
            waitType:    "Timestamp",
            waitValue:   time.Now().Add(10 * time.Second).Format(time.RFC3339),
            expectedMin: 9 * time.Second,
            expectedMax: 11 * time.Second,
        },
        {
            name:        "seconds_path_wait",
            waitType:    "SecondsPath",
            waitValue:   "$.waitTime",
            expectedMin: 3 * time.Second,
            expectedMax: 4 * time.Second,
        },
    }
    
    for _, strategy := range waitStrategies {
        t.Run(strategy.name, func(t *testing.T) {
            input := map[string]interface{}{
                "waitType":  strategy.waitType,
                "waitValue": strategy.waitValue,
                "waitTime":  3, // For SecondsPath test
            }
            
            startTime := time.Now()
            execution := stepfunctions.StartExecution(ctx, waitStateMachine, input)
            result := stepfunctions.WaitForCompletion(ctx, execution.ExecutionArn, 30*time.Second)
            actualDuration := time.Since(startTime)
            
            assert.Equal(t, stepfunctions.ExecutionStatusSucceeded, result.Status)
            assert.True(t, actualDuration >= strategy.expectedMin, 
                "Wait should be at least %v, was %v", strategy.expectedMin, actualDuration)
            assert.True(t, actualDuration <= strategy.expectedMax,
                "Wait should be at most %v, was %v", strategy.expectedMax, actualDuration)
        })
    }
}
```

## Monitoring and Observability

### Workflow Metrics
```go
// Comprehensive workflow metrics collection
type WorkflowMetrics struct {
    ExecutionCount      int64         `json:"executionCount"`
    SuccessRate        float64       `json:"successRate"`
    AverageExecutionTime time.Duration `json:"averageExecutionTime"`
    P95ExecutionTime    time.Duration `json:"p95ExecutionTime"`
    ErrorRate          float64       `json:"errorRate"`
    RetryRate          float64       `json:"retryRate"`
    CostPerExecution   float64       `json:"costPerExecution"`
    StateTransitions   map[string]int `json:"stateTransitions"`
}

func collectWorkflowMetrics(t *testing.T, ctx *WorkflowTestContext) WorkflowMetrics {
    // Implementation would collect metrics from CloudWatch, X-Ray, etc.
    return WorkflowMetrics{
        ExecutionCount:      100,
        SuccessRate:        0.98,
        AverageExecutionTime: 45 * time.Second,
        P95ExecutionTime:    120 * time.Second,
        ErrorRate:          0.02,
        RetryRate:          0.05,
        CostPerExecution:   0.025,
    }
}
```

### X-Ray Tracing Validation
```go
// Validate X-Ray tracing for workflow execution
func TestWorkflowTracingValidation(t *testing.T) {
    // Enable X-Ray tracing for the execution
    execution := stepfunctions.StartExecutionWithTracing(ctx, tracedWorkflow, input)
    result := stepfunctions.WaitForCompletion(ctx, execution.ExecutionArn, 60*time.Second)
    
    assert.Equal(t, stepfunctions.ExecutionStatusSucceeded, result.Status)
    
    // Retrieve and validate trace data
    traceId := extractTraceId(result)
    require.NotEmpty(t, traceId)
    
    trace := xray.GetTrace(ctx, traceId)
    validateTraceCompleteness(t, trace, execution)
    validateTracePerformance(t, trace, 30*time.Second)
    validateTraceErrorHandling(t, trace)
}
```

### CloudWatch Logs Analysis
```go
// Validate CloudWatch logs for workflow execution
func TestWorkflowLogsValidation(t *testing.T) {
    execution := stepfunctions.StartExecution(ctx, loggedWorkflow, input)
    result := stepfunctions.WaitForCompletion(ctx, execution.ExecutionArn, 60*time.Second)
    
    // Retrieve logs for the execution
    logGroups := []string{
        "/aws/stepfunctions/MyWorkflow",
        "/aws/lambda/workflow-processor",
    }
    
    logs := cloudwatch.GetExecutionLogs(ctx, logGroups, execution.ExecutionArn)
    
    // Validate log completeness
    validateLogCompleteness(t, logs, execution)
    validateLogStructure(t, logs)
    validateLogCorrelation(t, logs, execution.ExecutionArn)
    
    // Validate no sensitive data in logs
    validateLogDataPrivacy(t, logs)
}
```

---

**Next**: Explore specific workflow patterns that match your serverless architecture requirements.