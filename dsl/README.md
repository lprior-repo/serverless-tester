# AWS Testing DSL

A comprehensive Domain-Specific Language for AWS serverless testing that replaces the Function/FunctionE patterns with fluent, expressive interfaces.

## Overview

This DSL provides maximum abstraction for AWS testing, making test writing feel like natural language descriptions. It follows the key design principles:

- **Fluent interfaces** for maximum readability
- **Method chaining** for natural test flow  
- **Implicit error handling** with descriptive failures
- **High-level abstractions** that hide AWS complexity
- **Compositional builders** for complex scenarios

## Basic Usage

### Simple Lambda Function Test

```go
func TestLambdaInvocation(t *testing.T) {
    aws := NewAWS(t)
    
    result := aws.Lambda().
        Function("my-function").
        WithJSONPayload(map[string]interface{}{"key": "value"}).
        Sync().
        WithLogs().
        Invoke()
    
    result.Should().
        Succeed().
        And().
        CompleteWithin(5 * time.Second).
        And().
        HaveMetadata("function_name", "my-function")
}
```

### DynamoDB Operations

```go
func TestDynamoDBOperations(t *testing.T) {
    aws := NewAWS(t)
    
    // Create table
    aws.DynamoDB().
        CreateTable().
        Named("users").
        WithHashKey("userId", "S").
        WithRangeKey("timestamp", "N").
        WithOnDemandBilling().
        Create().
        Should().
        Succeed()
    
    // Put item
    aws.DynamoDB().
        Table("users").
        PutItem().
        WithAttribute("userId", "user123").
        WithAttribute("name", "John Doe").
        Execute().
        Should().
        Succeed()
    
    // Query items
    aws.DynamoDB().
        Table("users").
        Query().
        WhereKey("userId").
        Equals("user123").
        Limit(10).
        Execute().
        Should().
        Succeed().
        And().
        ContainValue("user123")
}
```

### S3 Bucket Management

```go
func TestS3Operations(t *testing.T) {
    aws := NewAWS(t)
    
    // Create bucket
    aws.S3().
        CreateBucket().
        Named("test-bucket-12345").
        InRegion("us-west-2").
        WithVersioning().
        WithEncryption().
        Create().
        Should().
        Succeed()
    
    // Upload object
    aws.S3().
        Bucket("test-bucket-12345").
        UploadObject().
        WithKey("test/file.json").
        WithBody(`{"message": "Hello World"}`).
        WithContentType("application/json").
        Execute().
        Should().
        Succeed()
}
```

## Supported AWS Services

### Core Services

- **Lambda** - Function invocation, configuration management, log retrieval
- **DynamoDB** - Table operations, item management, queries and scans
- **S3** - Bucket management, object operations, versioning, CORS
- **SES** - Email sending, identity verification, template management
- **EventBridge** - Event buses, rules, event publishing, archives
- **Step Functions** - State machines, executions, activities
- **API Gateway** - REST APIs, resources, methods, deployments

### Service Features

#### Lambda
- Synchronous and asynchronous invocations
- JSON payload handling
- Version and alias management
- CloudWatch log integration
- Configuration updates

#### DynamoDB
- Table creation and management
- Put, get, query, scan operations
- Conditional expressions
- Consistent reads
- Parallel scanning

#### S3
- Bucket creation and configuration
- Object upload/download
- Versioning and lifecycle management
- CORS configuration
- Presigned URL generation

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