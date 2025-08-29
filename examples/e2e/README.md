# End-to-End Testing Examples

Complete serverless application testing scenarios using the enhanced VasDeference framework with table-driven testing, parallel execution, and bulletproof database snapshots.

## Overview

These examples demonstrate real-world, production-ready testing patterns for complex serverless applications. Each example includes:

- **Full TDD Implementation**: Red-Green-Refactor cycles with comprehensive test coverage
- **Table-Driven Testing**: Advanced parallel execution with concurrency control
- **Database Snapshots**: Compressed and encrypted state management
- **Cross-Service Integration**: Lambda, DynamoDB, Step Functions, and EventBridge
- **Error Handling**: Comprehensive failure scenarios and recovery testing
- **Performance Validation**: Load testing and benchmarking patterns

## Examples

### 1. E-Commerce Order Processing Pipeline
**File**: `ecommerce_pipeline_test.go`

Complete order processing workflow testing:
- Order validation Lambda functions
- DynamoDB inventory management
- EventBridge order status events
- Step Functions order fulfillment workflow
- Payment processing integration
- Inventory snapshots with rollback
- Performance benchmarking

### 2. Financial Transaction Processing
**File**: `financial_processing_test.go`

Financial transaction system testing:
- Multi-region transaction processing
- ACID compliance validation
- Fraud detection workflows
- Real-time reporting
- Compliance audit trails
- Disaster recovery scenarios

### 3. IoT Data Processing Pipeline
**File**: `iot_pipeline_test.go`

IoT data ingestion and processing:
- Streaming data validation
- Real-time analytics
- Device management
- Data archival workflows
- Alert processing
- Load testing with parallel execution

### 4. ML Model Deployment Pipeline
**File**: `ml_deployment_test.go`

Machine learning model lifecycle testing:
- Model training validation
- A/B testing workflows
- Performance monitoring
- Model rollback scenarios
- Feature store integration

## Key Features Demonstrated

### Advanced Table Testing
```go
// Parallel execution with concurrency control
TableTest[OrderRequest, OrderResult](t, "Order Processing Pipeline").
    Case("valid_order", validOrder, expectedResult).
    Case("invalid_payment", invalidPayment, paymentError).
    Case("insufficient_inventory", lowStock, stockError).
    Parallel().
    WithMaxWorkers(10).
    Repeat(5).
    Run(processOrder, validateOrderResult)
```

### Database Snapshots with Compression
```go
// Bulletproof database state management
snapshot := dynamodb.CreateSnapshot(t, tableName, dynamodb.SnapshotOptions{
    Compression: true,
    Encryption:  true,
    Incremental: true,
})
defer snapshot.Restore()
```

### Parallel Step Functions Orchestration
```go
// Concurrent workflow execution
executions := []stepfunctions.ParallelExecution{
    {StateMachineArn: orderSM, ExecutionName: "order-1", Input: order1},
    {StateMachineArn: orderSM, ExecutionName: "order-2", Input: order2},
    {StateMachineArn: paymentSM, ExecutionName: "payment-1", Input: payment1},
}

results := stepfunctions.ParallelExecutions(ctx, executions, &stepfunctions.ParallelExecutionConfig{
    MaxConcurrency:    5,
    WaitForCompletion: true,
    FailFast:         false,
    Timeout:          30 * time.Second,
})
```

### Cross-Service Integration Testing
```go
// Complete workflow validation
func TestCompleteOrderWorkflow(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Setup test infrastructure
    ctx := createTestContext(vdf)
    snapshot := createDatabaseSnapshot(ctx)
    defer snapshot.Restore()
    
    // Execute complete workflow
    result := executeOrderWorkflow(ctx, testOrder)
    
    // Validate all components
    validateLambdaExecution(ctx, result.LambdaInvocations)
    validateDatabaseChanges(ctx, result.DatabaseOperations)
    validateEventBridgeEvents(ctx, result.Events)
    validateStepFunctionExecution(ctx, result.WorkflowExecution)
}
```

## Running the Examples

### Prerequisites
```bash
# Set required environment variables
export AWS_REGION=us-east-1
export AWS_PROFILE=your-test-profile
export TEST_NAMESPACE=e2e-tests

# Deploy test infrastructure (optional)
terraform init && terraform apply -var="namespace=${TEST_NAMESPACE}"
```

### Execute Examples
```bash
# Run all E2E tests
go test -v ./examples/e2e/

# Run specific scenarios
go test -v ./examples/e2e/ -run TestECommerceOrderPipeline
go test -v ./examples/e2e/ -run TestFinancialProcessing

# Run with parallel execution
go test -v -parallel 4 ./examples/e2e/

# Run performance benchmarks
go test -v -bench=. ./examples/e2e/
```

### Load Testing
```bash
# High-concurrency testing
TEST_CONCURRENCY=50 go test -v ./examples/e2e/ -run TestLoadTesting

# Extended duration testing
TEST_DURATION=300s go test -v ./examples/e2e/ -timeout 600s -run TestEndurance
```

## Best Practices Demonstrated

### 1. Resource Isolation
Each test uses namespace isolation to prevent conflicts:
```go
vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDeferenceOptions{
    Namespace: fmt.Sprintf("e2e-test-%s", random.UniqueId()),
})
```

### 2. Comprehensive Cleanup
Automatic resource cleanup with proper error handling:
```go
vdf.RegisterCleanup(func() error {
    return cleanupTestResources(ctx, resources)
})
```

### 3. Snapshot-Based Testing
Database state validation with snapshots:
```go
snapshot.MatchDynamoDBTable(t, "order_state", tableName)
```

### 4. Performance Validation
Built-in benchmarking and performance assertions:
```go
benchmark := TableTest[OrderRequest, OrderResult](t, "Performance Test").
    Benchmark(processOrder, 1000)
assert.True(t, benchmark.AverageTime < 100*time.Millisecond)
```

## Error Scenarios

Each example includes comprehensive error testing:

- Network failures and timeouts
- Service unavailability
- Data corruption scenarios
- Race condition testing
- Resource limit testing
- Security validation

## Monitoring and Observability

Examples include observability testing:

- CloudWatch metrics validation
- X-Ray trace validation
- Custom metrics verification
- Alert testing
- Dashboard validation

---

**Next Steps**: Choose an example to explore detailed implementation patterns and advanced testing techniques.