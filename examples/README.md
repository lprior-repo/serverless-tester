# ⚡ SFX Framework - Pure Functional Programming Examples

Comprehensive real-world testing examples demonstrating the full power of the **pure functional programming** SFX framework with immutable data structures, monadic error handling, and mathematical precision using **samber/lo** and **samber/mo**.

## 🚀 Functional Programming Overview

This examples directory showcases the complete capabilities of the **functional SFX** framework:

- **🏗️ Functional Core**: Immutable configurations with functional options pattern
- **⚡ Functional Lambda**: Type-safe Lambda operations with `mo.Option[T]` and `mo.Result[T]`  
- **🗄️ Functional DynamoDB**: Immutable table operations with `lo.Reduce` pipelines
- **📡 Functional EventBridge**: Event-driven patterns with immutable event structures
- **🔄 Functional Step Functions**: Workflow orchestration with pure function composition

### **Pure Functional Programming Benefits:**
- **✅ Type Safety** - Compile-time error prevention with Go generics and monads
- **✅ Zero Mutations** - All data structures are immutable by design
- **✅ Monadic Operations** - Safe error handling with `mo.Option[T]` and `mo.Result[T]`
- **✅ Function Composition** - Elegant data pipelines using `lo.Reduce`, `lo.Map`, `lo.Filter`
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
// Functional E2E testing with immutable configuration and monadic operations
config := NewFunctionalTestConfig(
    WithParallelExecution(true),
    WithMaxWorkers(10),
    WithTimeout(120*time.Second),
    WithRetryPolicy(5),
)

orderResults := lo.Map(testCases, func(testCase OrderTestCase, _ int) mo.Result[OrderResult] {
    return ProcessFunctionalOrder(testCase.Input).
        Map(func(result OrderResult) OrderResult {
            return validateOrderResult(result)
        })
})

// Collect results with monadic error handling
finalResult := lo.Reduce(orderResults, func(acc mo.Result[[]OrderResult], result mo.Result[OrderResult], _ int) mo.Result[[]OrderResult] {
    return acc.FlatMap(func(results []OrderResult) mo.Result[[]OrderResult] {
        return result.Map(func(r OrderResult) []OrderResult {
            return append(results, r)
        })
    })
}, mo.Some([]OrderResult{}))
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
// Functional performance benchmarking with immutable metrics
benchmarkConfig := NewFunctionalBenchmarkConfig(
    WithIterations(10000),
    WithWarmupIterations(1000),
    WithMetricsCollection(true),
)

performanceResult := RunFunctionalBenchmark(benchmarkConfig, testFunction).
    Map(func(metrics BenchmarkMetrics) BenchmarkMetrics {
        return validatePerformanceThresholds(metrics)
    })

performanceResult.GetError().
    Map(func(err error) error {
        t.Errorf("Performance benchmark failed: %v", err)
        return err
    }).
    OrElse(func() {
        t.Log("✓ Performance benchmarks passed")
    })
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
// Functional multi-environment promotion pipeline with immutable configuration
promotionConfig := NewFunctionalPromotionConfig(
    WithStage("dev", "staging", []string{"unit", "integration", "security"}),
    WithStage("staging", "prod", []string{"e2e", "performance", "compliance"}),
    WithValidationPolicy(StrictValidation),
)

promotionResults := lo.Map(promotionConfig.Stages, func(stage PromotionStage, _ int) mo.Result[PromotionResult] {
    return ExecuteFunctionalPromotion(stage).
        Map(func(result PromotionResult) PromotionResult {
            return ValidatePromotionSuccess(result)
        })
})

// Aggregate results with early failure detection
finalResult := lo.Reduce(promotionResults, func(acc mo.Result[bool], result mo.Result[PromotionResult], _ int) mo.Result[bool] {
    return acc.FlatMap(func(success bool) mo.Result[bool] {
        return result.Map(func(_ PromotionResult) bool { return success })
    })
}, mo.Some(true))
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
// Functional parallel workflow orchestration with immutable configuration
workflowConfig := NewFunctionalWorkflowConfig(
    WithMaxConcurrency(15),
    WithTimeout(300*time.Second),
    WithFailFastPolicy(false),
    WithCompletionWait(true),
)

executionSpecs := []WorkflowExecutionSpec{
    NewExecutionSpec(orderSM, "order-1", order1),
    NewExecutionSpec(paymentSM, "payment-1", payment1),
    NewExecutionSpec(inventorySM, "inventory-1", inventory1),
}

// Execute all workflows functionally with monadic error handling
workflowResults := lo.Map(executionSpecs, func(spec WorkflowExecutionSpec, _ int) mo.Result[WorkflowResult] {
    return ExecuteFunctionalWorkflow(workflowConfig, spec).
        Map(func(result WorkflowResult) WorkflowResult {
            return ValidateWorkflowCompletion(result)
        })
})

// Aggregate results with proper error propagation
aggregatedResult := AggregateFunctionalResults(workflowResults)
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
// Complete functional e-commerce order processing with immutable pipeline
func TestFunctionalOrderProcessing(t *testing.T) {
    infrastructureConfig := NewFunctionalInfrastructureConfig(
        WithOrderService("order-lambda"),
        WithPaymentService("payment-lambda"),
        WithInventoryService("inventory-lambda"),
        WithNotificationService("notification-lambda"),
    )
    
    // Create immutable order processing pipeline
    processingPipeline := NewFunctionalOrderPipeline(
        WithOrderValidation(),
        WithInventoryReservation(),
        WithPaymentProcessing(),
        WithShippingArrangement(),
        WithNotificationSending(),
        WithAnalyticsRecording(),
    )
    
    // Execute pipeline with monadic error handling
    orderResult := ProcessFunctionalOrder(processingPipeline, testOrder).
        Map(func(result OrderResult) OrderResult {
            return ValidateCompleteOrderResult(result)
        })
    
    // Assert success with functional error handling
    orderResult.GetError().
        Map(func(err error) error {
            t.Errorf("Order processing failed: %v", err)
            return err
        }).
        OrElse(func() {
            t.Log("✓ Complete order processing successful")
        })
}
```

## Key Enhanced Features

### Functional Table Testing
```go
// Immutable table-driven testing with functional configuration
testConfig := NewFunctionalTableTestConfig(
    WithParallelExecution(true),
    WithMaxWorkers(10),
    WithRepeatCount(5),
    WithTimeout(60*time.Second),
    WithFocusFilter([]string{"case2"}),
    WithSkipPredicate(func(name string) bool {
        return strings.Contains(name, "slow")
    }),
)

testCases := []FunctionalTestCase[Input, Output]{
    NewTestCase("case1", input1, expected1),
    NewTestCase("case2", input2, expected2),
    NewTestCase("case3", input3, expected3),
}

// Execute functional table test with monadic results
testResults := lo.Map(testCases, func(testCase FunctionalTestCase[Input, Output], _ int) mo.Result[TestResult] {
    return ExecuteFunctionalTest(testConfig, testCase, testFunction).
        Map(func(result TestResult) TestResult {
            return assertFunction(result)
        })
})

// Aggregate results with proper error handling
finalResult := AggregateFunctionalTestResults(testResults)
```

### Functional Database Snapshots
```go
// Immutable database state management with functional configuration
snapshotConfig := NewFunctionalSnapshotConfig(
    WithCompression(true),
    WithEncryption(true),
    WithIncrementalBackup(true),
    WithMetadata(map[string]interface{}{
        "test_suite": "e2e_tests",
        "version":    "v1.2.3",
    }),
)

snapshotResult := CreateFunctionalSnapshot([]string{tableName}, snapshotConfig).
    Map(func(snapshot FunctionalSnapshot) FunctionalSnapshot {
        return ValidateSnapshotIntegrity(snapshot)
    })

// Restore database state functionally with monadic error handling
restoreResult := snapshotResult.FlatMap(func(snapshot FunctionalSnapshot) mo.Result[RestoreResult] {
    return RestoreFunctionalSnapshot(snapshot).
        Map(func(result RestoreResult) RestoreResult {
            return ValidateDatabaseState(result, tableName)
        })
})
```

### Functional Performance Benchmarking
```go
// Immutable performance analysis with functional configuration
benchmarkConfig := NewFunctionalBenchmarkConfig(
    WithIterations(1000),
    WithWarmupRounds(100),
    WithMetricsCollection([]MetricType{AverageTime, P50, P95, P99, Throughput}),
    WithThresholds(map[MetricType]time.Duration{
        AverageTime: 50 * time.Millisecond,
        P50:         30 * time.Millisecond,
        P95:         100 * time.Millisecond,
        P99:         200 * time.Millisecond,
    }),
)

benchmarkResult := RunFunctionalBenchmark(benchmarkConfig, performanceFunction).
    Map(func(metrics BenchmarkMetrics) BenchmarkMetrics {
        return ValidatePerformanceThresholds(metrics, benchmarkConfig.Thresholds)
    })

// Validate performance metrics with monadic error handling
benchmarkResult.GetError().
    Map(func(err error) error {
        t.Errorf("Performance benchmark failed: %v", err)
        return err
    }).
    OrElse(func() {
        t.Log("✓ All performance thresholds met")
    })
```

### Functional Snapshot Testing
```go
// Immutable snapshot testing with functional configuration
snapshotConfig := NewFunctionalSnapshotTesterConfig(
    WithSnapshotDirectory("testdata/snapshots"),
    WithJSONFormatting(true),
    WithSanitizers([]FunctionalSanitizer{
        NewTimestampSanitizer(),
        NewUUIDSanitizer(),
        NewExecutionArnSanitizer(),
        NewCustomSanitizer(`"orderId":"[^"]*"`, `"orderId":"<ORDER_ID>"`),
    }),
)

// Create immutable snapshot tester
snapshotTester := CreateFunctionalSnapshotTester(snapshotConfig)

// Validate complex data structures with monadic results
validationResults := []mo.Result[SnapshotValidation]{
    snapshotTester.MatchJSON("business_result", businessResult),
    snapshotTester.MatchDynamoDBItems("order_records", orderRecords),
    snapshotTester.MatchStepFunctionExecution("workflow_output", workflowOutput),
}

// Aggregate validation results
finalValidation := AggregateFunctionalValidations(validationResults)
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