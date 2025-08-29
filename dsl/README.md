# SFX Functional AWS Testing DSL

A comprehensive Domain-Specific Language for AWS serverless testing built on pure functional programming principles using **samber/lo** and **samber/mo** for immutable data structures and monadic error handling.

## Overview

This functional DSL provides maximum abstraction for AWS testing while maintaining mathematical precision and type safety. It follows these core functional programming principles:

- **Immutable configurations** with functional options pattern
- **Monadic error handling** using `mo.Option[T]` and `mo.Result[T]`
- **Function composition** with `lo.Map`, `lo.Filter`, and `lo.Reduce`
- **Zero mutations** - all operations return new immutable structures
- **Type-safe operations** leveraging Go generics for compile-time safety

## Basic Usage

### Functional Lambda Function Test

```go
func TestFunctionalLambdaInvocation(t *testing.T) {
    // Create immutable Lambda configuration using functional options
    lambdaConfig := NewFunctionalLambdaConfig(
        WithFunctionName("my-function"),
        WithJSONPayload(map[string]interface{}{"key": "value"}),
        WithSynchronousInvocation(),
        WithLogRetrieval(true),
        WithTimeout(5 * time.Second),
    )
    
    // Execute Lambda function with monadic error handling
    invocationResult := InvokeFunctionalLambda(lambdaConfig).
        Map(func(result LambdaInvocationResult) LambdaInvocationResult {
            return ValidateLambdaResponse(result)
        })
    
    // Assert success using monadic error handling
    invocationResult.GetError().
        Map(func(err error) error {
            t.Errorf("Lambda invocation failed: %v", err)
            return err
        }).
        OrElse(func() {
            t.Log("✓ Lambda invocation successful")
        })
}
```

### Functional DynamoDB Operations

```go
func TestFunctionalDynamoDBOperations(t *testing.T) {
    // Create immutable table configuration
    tableConfig := NewFunctionalTableConfig(
        WithTableName("users"),
        WithHashKey("userId", "S"),
        WithRangeKey("timestamp", "N"),
        WithOnDemandBilling(),
    )
    
    // Create table with monadic error handling
    tableCreationResult := CreateFunctionalTable(tableConfig).
        Map(func(result TableCreationResult) TableCreationResult {
            return ValidateTableCreation(result)
        })
    
    // Create immutable item for insertion
    userItem := NewFunctionalDynamoDBItem(
        WithAttribute("userId", "user123"),
        WithAttribute("name", "John Doe"),
        WithAttribute("timestamp", time.Now().Unix()),
    )
    
    // Put item functionally
    putResult := tableCreationResult.FlatMap(func(_ TableCreationResult) mo.Result[PutItemResult] {
        return PutFunctionalItem("users", userItem).
            Map(func(result PutItemResult) PutItemResult {
                return ValidateItemInsertion(result)
            })
    })
    
    // Create immutable query configuration
    queryConfig := NewFunctionalQueryConfig(
        WithTableName("users"),
        WithKeyCondition("userId", "user123"),
        WithLimit(10),
        WithConsistentRead(true),
    )
    
    // Execute query with function composition
    queryResult := putResult.FlatMap(func(_ PutItemResult) mo.Result[QueryResult] {
        return ExecuteFunctionalQuery(queryConfig).
            Map(func(result QueryResult) QueryResult {
                return ValidateQueryResults(result, "user123")
            })
    })
    
    // Assert all operations succeeded
    queryResult.GetError().
        Map(func(err error) error {
            t.Errorf("DynamoDB operations failed: %v", err)
            return err
        }).
        OrElse(func() {
            t.Log("✓ All DynamoDB operations successful")
        })
}
```

### Functional S3 Bucket Management

```go
func TestFunctionalS3Operations(t *testing.T) {
    // Create immutable bucket configuration
    bucketConfig := NewFunctionalS3BucketConfig(
        WithBucketName("test-bucket-12345"),
        WithRegion("us-west-2"),
        WithVersioning(true),
        WithEncryption(true),
        WithPublicAccessBlocking(true),
    )
    
    // Create bucket with monadic error handling
    bucketCreationResult := CreateFunctionalS3Bucket(bucketConfig).
        Map(func(result S3BucketCreationResult) S3BucketCreationResult {
            return ValidateBucketCreation(result)
        })
    
    // Create immutable object configuration
    objectConfig := NewFunctionalS3ObjectConfig(
        WithBucketName("test-bucket-12345"),
        WithObjectKey("test/file.json"),
        WithBody(`{"message": "Hello World"}`),
        WithContentType("application/json"),
        WithMetadata(map[string]string{
            "created_by": "functional_test",
            "version":    "1.0",
        }),
    )
    
    // Upload object with function composition
    uploadResult := bucketCreationResult.FlatMap(func(_ S3BucketCreationResult) mo.Result[S3UploadResult] {
        return UploadFunctionalS3Object(objectConfig).
            Map(func(result S3UploadResult) S3UploadResult {
                return ValidateObjectUpload(result)
            })
    })
    
    // Verify operations with monadic error handling
    uploadResult.GetError().
        Map(func(err error) error {
            t.Errorf("S3 operations failed: %v", err)
            return err
        }).
        OrElse(func() {
            t.Log("✓ S3 bucket and object operations successful")
        })
}
```

## Supported AWS Services

### Functional AWS Services

- **Lambda** - Immutable function configurations, monadic invocation results, type-safe payload handling
- **DynamoDB** - Functional table operations, immutable item structures, composable query pipelines
- **S3** - Immutable bucket configurations, type-safe object operations, functional metadata handling
- **SES** - Functional email configurations, monadic sending results, immutable template management
- **EventBridge** - Immutable event structures, functional rule configurations, composable event pipelines
- **Step Functions** - Immutable state machine definitions, functional execution management, monadic result handling
- **API Gateway** - Functional API configurations, immutable resource definitions, composable integration patterns

### Service Features

#### Functional Lambda
- Immutable invocation configurations with functional options
- Type-safe JSON payload handling using Go generics
- Monadic result handling for synchronous/asynchronous operations
- Functional CloudWatch log integration with `mo.Option[T]`
- Zero-mutation configuration updates using functional transformations

#### Functional DynamoDB  
- Immutable table configurations with functional options pattern
- Type-safe CRUD operations using `lo.Map` and `lo.Filter`
- Functional conditional expressions with monadic validation
- Consistent read configurations using `mo.Option[T]`
- Composable parallel scanning with `lo.Reduce` aggregation

#### Functional S3
- Immutable bucket configurations with functional options
- Type-safe object operations using Go generics
- Functional versioning and lifecycle management
- Immutable CORS configuration structures
- Monadic presigned URL generation with error handling

#### SES
- Email sending with templates
- Identity verification
- Configuration sets
- Bounce/complaint handling
- Send quota monitoring

#### EventBridge
- Custom event buses
- Rule creation with event patterns
- Event publishing
- Archive and replay functionality

#### Step Functions
- State machine definition
- Execution management
- Activity polling
- History retrieval

#### API Gateway
- REST and HTTP API management
- Resource and method creation
- Lambda and HTTP integrations
- Stage management
- Test invocation

## High-Level Testing Features

### Scenarios

Build complex testing scenarios with multiple steps:

```go
func TestComplexScenario(t *testing.T) {
    aws := NewAWS(t)
    
    scenario := aws.Scenario().
        Step("Create infrastructure").
        Describe("Set up S3 bucket and DynamoDB table").
        Do(func() *ExecutionResult {
            // Infrastructure setup logic
            return setupInfrastructure(aws)
        }).
        Done().
        Step("Test workflow").
        Describe("Execute end-to-end workflow").
        Do(func() *ExecutionResult {
            // Workflow testing logic
            return testWorkflow(aws)
        }).
        Done()
    
    scenario.Execute().
        Should().
        Succeed().
        And().
        CompleteWithin(30 * time.Second)
}
```

### Test Orchestration

Orchestrate complete tests with setup, scenarios, and teardown:

```go
func TestFullOrchestration(t *testing.T) {
    aws := NewAWS(t)
    
    setupScenario := aws.Scenario().
        Step("Create infrastructure").
        Do(setupInfrastructure).
        Done()
    
    testScenario := aws.Scenario().
        Step("Run tests").
        Do(runMainTests).
        Done()
    
    aws.Test().
        Named("Full Integration Test").
        WithSetup(func() *ExecutionResult {
            return setupScenario.Execute()
        }).
        AddScenario(testScenario).
        WithTeardown(func() *ExecutionResult {
            return cleanupResources()
        }).
        Run().
        Should().
        Succeed().
        And().
        HaveMetadata("test_name", "Full Integration Test")
}
```

### Asynchronous Operations

Handle long-running operations asynchronously:

```go
func TestAsyncOperations(t *testing.T) {
    aws := NewAWS(t)
    
    // Start async operation
    asyncResult := aws.Lambda().
        Function("long-running-task").
        WithJSONPayload(map[string]interface{}{"task": "process-data"}).
        Async().
        InvokeAsync()
    
    // Do other work while operation runs
    aws.S3().Bucket("temp-bucket").Object("status.json").Download()
    
    // Wait for completion
    asyncResult.
        WaitWithTimeout(10 * time.Second).
        Should().
        Succeed()
}
```

## Assertion Framework

The DSL includes a comprehensive assertion framework:

### Basic Assertions

- `.Should().Succeed()` - Assert operation succeeded
- `.Should().Fail()` - Assert operation failed
- `.Should().ReturnValue(expected)` - Assert specific return value
- `.Should().ContainValue(expected)` - Assert result contains value

### Metadata Assertions

- `.Should().HaveMetadata(key, value)` - Assert specific metadata exists

### Performance Assertions

- `.Should().CompleteWithin(duration)` - Assert operation completed within timeframe

### Chained Assertions

All assertions can be chained with `.And()`:

```go
result.Should().
    Succeed().
    And().
    CompleteWithin(5 * time.Second).
    And().
    HaveMetadata("status", "completed").
    And().
    ReturnValue("expected-result")
```

## Architecture

The DSL is built on several key architectural patterns:

### Builder Pattern
Each AWS service uses the builder pattern for fluent method chaining.

### Context Management
All operations share a common `TestContext` that manages AWS configuration, timeouts, and metadata.

### Result Handling
Operations return `ExecutionResult` objects that can be chained with assertions.

### Scenario Composition
Complex testing workflows are composed using scenario builders that can be nested and orchestrated.

## Migration from Function/FunctionE

The DSL replaces the old Function/FunctionE patterns:

### Before (Function/FunctionE)
```go
// Old pattern
result := lambda.InvokeFunction(t, awsRegion, functionName, payload)
require.NoError(t, result.Error)
assert.Equal(t, expectedOutput, result.Payload)
```

### After (DSL)
```go
// New DSL pattern
aws.Lambda().
    Function(functionName).
    WithJSONPayload(payload).
    Invoke().
    Should().
    Succeed().
    And().
    ReturnValue(expectedOutput)
```

The DSL provides:
- **Better readability** - Tests read like natural language
- **Consistent API** - All AWS services use the same patterns
- **Rich assertions** - Built-in assertion framework
- **Error handling** - Implicit error handling with descriptive messages
- **Composability** - Easy to build complex scenarios

## Best Practices

1. **Use descriptive test names** that explain the business scenario
2. **Leverage scenarios** for multi-step workflows
3. **Chain assertions** for comprehensive validation
4. **Use async operations** for long-running tasks
5. **Implement proper cleanup** in teardown functions
6. **Test timeouts** to avoid hanging tests
7. **Validate metadata** for detailed assertions

## Examples

See the `examples.go` file for comprehensive examples of all DSL features and patterns.