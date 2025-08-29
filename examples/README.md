# ⚡ Vas Deference - Functional Programming Examples

Comprehensive real-world testing examples demonstrating the full power of the **pure functional programming** Vas Deference framework with immutable data structures, monadic error handling, and mathematical precision.

## 🚀 Functional Programming Overview

This examples directory showcases the complete capabilities of the **functional Vas Deference** framework:

- **🏗️ Functional Core**: Immutable configurations, monadic operations, and pure functions
- **⚡ Functional Lambda**: Type-safe Lambda operations with monadic results  
- **🗄️ Functional DynamoDB**: Immutable table operations with validation pipelines
- **📡 Functional EventBridge**: Event-driven patterns with immutable events
- **🔄 Functional Step Functions**: Workflow orchestration with immutable state machines

### **Pure Functional Programming Benefits:**
- **✅ Type Safety** - Compile-time error prevention with Go generics
- **✅ Immutability** - Zero mutations, only functional transformations
- **✅ Monadic Operations** - Safe error handling without exceptions
- **✅ Function Composition** - Elegant data transformation pipelines
- **✅ Mathematical Precision** - Predictable, composable, verifiable operations

## Advanced Examples

### [🚀 End-to-End Testing](./e2e/)
**Complete serverless application testing scenarios**

Production-ready E2E testing with the full VasDeference feature set:
- **Table-driven testing** with parallel execution and concurrency control
- **Database snapshots** with compression, encryption, and incremental backup
- **Cross-service integration** testing (Lambda + DynamoDB + Step Functions + EventBridge)
- **Performance validation** and load testing patterns
- **Error handling** and recovery testing

```go
// Advanced E2E testing with parallel execution
vasdeference.TableTest[OrderRequest, OrderResult](t, "E-Commerce Pipeline").
    Case("standard_order", validOrder, expectedResult).
    Case("bulk_order", bulkOrder, bulkResult).
    Case("international_order", intlOrder, intlResult).
    Parallel().
    WithMaxWorkers(10).
    Repeat(5).
    Timeout(120*time.Second).
    Run(processCompleteOrder, validateOrderResult)
```

### [⚡ Performance Testing](./performance/)
**Advanced benchmarking and load testing patterns**

High-performance testing with statistical analysis:
- **Lambda performance** benchmarking with cold/warm start analysis
- **DynamoDB throughput** testing with batch operations
- **Step Functions parallel** execution performance
- **Memory profiling** and leak detection
- **Regression testing** with baseline comparison

```go
// Statistical performance benchmarking
benchmark := vasdeference.TableTest[TestInput, TestOutput](t, "Performance").
    Case("baseline", input, expected).
    Benchmark(testFunction, 10000)

assert.Less(t, benchmark.AverageTime, 100*time.Millisecond)
assert.Less(t, benchmark.P95Time, 200*time.Millisecond)
```

### [🏢 Enterprise Integration](./enterprise/)
**Multi-environment, cross-region, and compliance testing**

Enterprise-grade testing patterns:
- **Multi-environment** testing (dev, staging, production)
- **Cross-region** disaster recovery scenarios
- **Security & compliance** validation (GDPR, HIPAA, SOX)
- **Configuration drift** detection
- **Access control** and audit trail validation

```go
// Multi-environment promotion pipeline
stages := []PromotionStage{
    {From: "dev", To: "staging", RequiredTests: []string{"unit", "integration", "security"}},
    {From: "staging", To: "prod", RequiredTests: []string{"e2e", "performance", "compliance"}},
}

for _, stage := range stages {
    result := executePromotionPipeline(t, ctx, stage)
    validatePromotionSuccess(t, result)
}
```

### [🔄 Advanced Workflows](./workflows/)
**Complex workflow orchestration and parallel execution**

Sophisticated workflow testing:
- **Parallel Step Functions** execution with synchronization
- **Complex state machines** with conditional logic and error handling
- **Workflow chaining** for sequential and parallel composition
- **Long-running workflows** with timeout and retry patterns
- **Event-driven workflows** with EventBridge integration

```go
// Parallel workflow orchestration
executions := []stepfunctions.ParallelExecution{
    {StateMachineArn: orderSM, ExecutionName: "order-1", Input: order1},
    {StateMachineArn: paymentSM, ExecutionName: "payment-1", Input: payment1},
    {StateMachineArn: inventorySM, ExecutionName: "inventory-1", Input: inventory1},
}

results := stepfunctions.ParallelExecutions(ctx, executions, &stepfunctions.ParallelExecutionConfig{
    MaxConcurrency: 15, WaitForCompletion: true, FailFast: false, Timeout: 300*time.Second,
})
```

### [🌍 Real-World Scenarios](./scenarios/)
**Complete business use case implementations**

Production-tested business scenarios:
- **E-Commerce Platform**: Complete order processing, inventory, and fulfillment
- **Financial Services**: Transaction processing, fraud detection, and compliance
- **IoT Data Processing**: Device data ingestion, processing, and analytics
- **ML Model Pipeline**: Training, deployment, and inference workflows
- **Healthcare Systems**: Patient data processing with HIPAA compliance

```go
// Complete e-commerce order processing
func TestCompleteOrderProcessing(t *testing.T) {
    ctx := setupECommerceInfrastructure(t)
    defer ctx.VDF.Cleanup()
    
    // Process order through complete business pipeline
    result := processCompleteOrder(t, ctx, testOrder)
    
    // Validate all business aspects
    validateOrderCreation(t, result.OrderCreated)
    validateInventoryReservation(t, result.InventoryReserved)  
    validatePaymentProcessing(t, result.PaymentProcessed)
    validateShippingArrangement(t, result.ShippingArranged)
    validateCustomerNotifications(t, result.NotificationsSent)
    validateAnalyticsEvents(t, result.AnalyticsRecorded)
}
```

## Key Enhanced Features

### Advanced Table Testing
```go
// Sophisticated table-driven testing with parallel execution
TableTest[Input, Output](t, "Test Name").
    Case("case1", input1, expected1).
    Case("case2", input2, expected2).
    Case("case3", input3, expected3).
    Parallel().                    // Enable parallel execution
    WithMaxWorkers(10).           // Control concurrency
    Repeat(5).                    // Repeat each case 5 times
    Timeout(60*time.Second).      // Per-case timeout
    Focus([]string{"case2"}).     // Run only specific cases
    Skip(func(name string) bool { // Skip cases conditionally
        return strings.Contains(name, "slow")
    }).
    Run(testFunction, assertFunction)
```

### Database Snapshots with Advanced Features
```go
// Bulletproof database state management
snapshot := dynamodb.CreateSnapshot(t, []string{tableName}, dynamodb.SnapshotOptions{
    Compression: true,             // Compress snapshots
    Encryption:  true,             // Encrypt sensitive data
    Incremental: true,             // Only store changes
    Metadata: map[string]interface{}{
        "test_suite": "e2e_tests",
        "version":    "v1.2.3",
    },
})
defer snapshot.Restore()

// Validate database state with snapshots
snapshot.MatchDynamoDBTable(t, "expected_state", tableName)
```

### Performance Benchmarking
```go
// Comprehensive performance analysis
benchmark := TableTest[Request, Response](t, "Performance Test").
    Case("performance_case", request, expected).
    Benchmark(performanceFunction, 1000)

// Detailed performance metrics
assert.Less(t, benchmark.AverageTime, 50*time.Millisecond)
assert.Less(t, benchmark.P50Time, 30*time.Millisecond)
assert.Less(t, benchmark.P95Time, 100*time.Millisecond)
assert.Less(t, benchmark.P99Time, 200*time.Millisecond)
assert.Greater(t, benchmark.ThroughputPerSecond, 1000.0)
```

### Snapshot Testing
```go
// Advanced snapshot testing with sanitization
snapshotTester := snapshot.New(t, snapshot.Options{
    SnapshotDir: "testdata/snapshots",
    JSONIndent:  true,
    Sanitizers: []snapshot.Sanitizer{
        snapshot.SanitizeTimestamps(),
        snapshot.SanitizeUUIDs(),
        snapshot.SanitizeExecutionArn(),
        snapshot.SanitizeCustom(`"orderId":"[^"]*"`, `"orderId":"<ORDER_ID>"`),
    },
})

// Validate complex data structures
snapshotTester.MatchJSON("business_result", businessResult)
snapshotTester.MatchDynamoDBItems("order_records", orderRecords)
snapshotTester.MatchStepFunctionExecution("workflow_output", workflowOutput)
```

## Quick Start

### Basic Usage
```bash
# Run all examples
go test -v ./examples/...

# Run specific categories
go test -v ./examples/e2e/          # End-to-end tests
go test -v ./examples/performance/  # Performance tests  
go test -v ./examples/enterprise/   # Enterprise tests
go test -v ./examples/workflows/    # Workflow tests
go test -v ./examples/scenarios/    # Business scenarios

# Run with specific configurations
PERFORMANCE_TEST_DURATION=300s go test -v ./examples/performance/
ENTERPRISE_TEST_MODE=true go test -v ./examples/enterprise/
SCENARIO_TEST_MODE=true go test -v ./examples/scenarios/
```

### Advanced Usage
```bash
# Performance testing with profiling
go test -v -bench=. -cpuprofile=cpu.prof ./examples/performance/
go test -v -bench=. -memprofile=mem.prof ./examples/performance/

# Load testing with high concurrency
CONCURRENCY=50 go test -v ./examples/e2e/ -run TestLoadTesting

# Enterprise compliance testing
COMPLIANCE_MODE=GDPR go test -v ./examples/enterprise/ -run TestCompliance

# Long-running scenario tests
go test -v -timeout=3600s ./examples/scenarios/ -run TestProductionScale
```

## Core Framework Services

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

### Advanced Test Setup with Options
```go
vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDeferenceOptions{
    Region:    "eu-west-1",
    Namespace: "my-test",
    MaxRetries: 5,
    RetryDelay: 2*time.Second,
})
```

### Resource Management with Cleanup
```go
resource := createResource()
vdf.RegisterCleanup(func() error {
    return resource.Delete()
})
```

### Parallel Test Execution
```go
// Table-driven testing with parallel execution
vasdeference.TableTest[TestInput, TestOutput](t, "Parallel Test").
    Case("case1", input1, expected1).
    Case("case2", input2, expected2).
    Case("case3", input3, expected3).
    Parallel().
    WithMaxWorkers(5).
    Run(testFunction, assertFunction)
```

## Testing Methodologies

### Test-Driven Development (TDD)
All examples follow strict TDD methodology:
1. **Red Phase**: Write failing test first
2. **Green Phase**: Write minimal code to pass
3. **Refactor Phase**: Clean up while keeping tests green

### Production Testing Patterns
- **Comprehensive error handling** with recovery scenarios
- **Performance validation** with SLA enforcement
- **Security testing** with compliance validation
- **Integration testing** with external services
- **Chaos engineering** with failure injection

### Business Validation
- **End-to-end workflows** matching real business processes
- **Data consistency** validation across services
- **Performance benchmarking** against business SLAs
- **Compliance testing** for regulatory requirements
- **Cost optimization** validation

## Run Examples

### Individual Service Examples
```bash
# Core framework patterns
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

### Advanced Example Categories
```bash
# End-to-end testing examples
go test -v ./examples/e2e/

# Performance testing examples
go test -v ./examples/performance/

# Enterprise integration examples  
go test -v ./examples/enterprise/

# Advanced workflow examples
go test -v ./examples/workflows/

# Real-world scenario examples
go test -v ./examples/scenarios/
```

## Key Features

### Enhanced Testing Capabilities
- **Advanced table testing** with parallel execution and concurrency control
- **Database snapshots** with compression, encryption, and incremental backup
- **Performance benchmarking** with statistical analysis and regression detection
- **Cross-service integration** testing with event correlation
- **Comprehensive error handling** with recovery validation

### Production-Ready Patterns
- **Multi-environment** testing and promotion pipelines
- **Cross-region** disaster recovery testing
- **Security and compliance** validation (GDPR, HIPAA, SOX, PCI-DSS)
- **Load testing** with realistic traffic patterns
- **Business logic validation** with complete workflows

### Developer Experience
- **Copy-paste ready** code snippets for immediate use
- **Comprehensive documentation** with real-world examples
- **TDD methodology** with red-green-refactor cycles
- **Snapshot testing** for complex data validation
- **Performance monitoring** with detailed metrics

---

**Next Steps:**
1. **Start with [Core Framework](./core/)** for basic patterns
2. **Explore [E2E Testing](./e2e/)** for comprehensive application testing
3. **Review [Performance Testing](./performance/)** for benchmarking patterns
4. **Check [Enterprise Integration](./enterprise/)** for multi-environment testing
5. **Study [Real-World Scenarios](./scenarios/)** for complete business use cases

**Choose the example category that best matches your testing needs and business domain.**