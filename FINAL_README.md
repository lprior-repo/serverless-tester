# Vas Deference - Comprehensive AWS Serverless Testing Framework

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![AWS SDK](https://img.shields.io/badge/AWS%20SDK-v2-FF9900?style=flat&logo=amazon-aws)](https://github.com/aws/aws-sdk-go-v2)
[![Terratest Compatible](https://img.shields.io/badge/Terratest-Compatible-green?style=flat)](https://terratest.gruntwork.io/)
[![Test Coverage](https://img.shields.io/badge/Coverage-95%25-brightgreen?style=flat)](./coverage)

**Vas Deference** is a comprehensive Go testing framework for AWS Serverless architectures, inspired by [Gruntwork's Terratest](https://terratest.gruntwork.io/). It provides Playwright-like syntax sugar with comprehensive AWS SDK v2 wrappers, making serverless testing intuitive, reliable, and maintainable.

> *"Defer to the experts, defer to the testing."* - The philosophy behind comprehensive AWS serverless testing.

## 🎯 Features

- **Terratest-Style API**: Every function follows the `Function/FunctionE` pattern for consistent error handling
- **Complete AWS SDK v2 Coverage**: Lambda, DynamoDB, EventBridge, Step Functions, CloudWatch Logs
- **Playwright-Like Syntax**: Intuitive APIs with intelligent abstractions and helper functions
- **Parallel Test Execution**: Built-in concurrency with resource isolation using worker pools
- **Snapshot Testing**: Golden file testing with intelligent sanitization for dynamic data
- **Table-Driven Tests**: First-class support for comprehensive test scenarios
- **Terraform Integration**: Seamless integration with Terraform-deployed infrastructure
- **Namespace Management**: Automatic PR-based resource isolation for CI/CD workflows
- **Comprehensive Assertions**: Rich assertion libraries for all AWS services
- **Pure Functional Design**: Immutable data structures and composable functions

## 🚀 Quick Start

```bash
go get vasdeference
```

### Basic Usage

```go
package main

import (
    "testing"
    "vasdeference"
    "vasdeference/lambda"
    "vasdeference/dynamodb"
)

func TestMyServerlessApp(t *testing.T) {
    // Initialize with automatic namespace detection (PR-based)
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()

    // Test Lambda function
    result := lambda.Invoke(t, vdf.GetTestContext(), "my-function", map[string]interface{}{
        "action": "process",
        "data":   "test-data",
    })
    lambda.AssertInvokeSuccess(t, result)

    // Test DynamoDB operations
    testItem := map[string]interface{}{
        "PK": "USER#123",
        "SK": "PROFILE",
        "Name": "John Doe",
    }
    dynamodb.PutItem(t, vdf.GetTestContext(), "users-table", testItem)
    
    item := dynamodb.GetItem(t, vdf.GetTestContext(), "users-table", 
        dynamodb.CreateKey("USER#123", "PROFILE"))
    dynamodb.AssertItemExists(t, item)
    dynamodb.AssertAttributeEquals(t, item, "Name", "John Doe")
}
```

## 📦 Package Structure

```
vasdeference/
├── vasdeference.go           # Core framework with TestContext and Arrange patterns
├── lambda/                   # Lambda testing utilities
├── dynamodb/                # DynamoDB testing helpers
├── eventbridge/             # EventBridge testing tools  
├── stepfunctions/           # Step Functions testing framework
├── parallel/                # Parallel test execution
├── snapshot/                # Snapshot testing with regression detection
└── integration_example_test.go # Comprehensive integration examples
```

## 🏗️ Core Packages

### 🔷 Core Framework (`vasdeference`)

The main package providing foundational testing infrastructure:

```go
// Initialize with automatic namespace detection
vdf := vasdeference.New(t)

// Or with custom options
vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDeferenceOptions{
    Region:     "eu-west-1",
    Namespace:  "integration-test",
    MaxRetries: 5,
    RetryDelay: 200 * time.Millisecond,
})

// Get AWS service clients
lambdaClient := vdf.GetLambdaClient()
dynamoClient := vdf.GetDynamoDBClient()

// Terraform integration
functionName := vdf.GetTerraformOutput(outputs, "lambda_function_name")
tableName := vdf.PrefixResourceName("users")  // → "pr-123-users"

// Resource cleanup (automatic via defer)
vdf.RegisterCleanup(func() error {
    return cleanupResource()
})
```

**Key Features:**
- Automatic PR namespace detection from CI/CD environments
- AWS client management with region support
- Terraform output integration
- LIFO cleanup registration
- Random string and unique ID generation

### 🚀 Lambda Testing (`vasdeference/lambda`)

Comprehensive Lambda function testing with event builders and log analysis:

```go
import "vasdeference/lambda"

// Basic invocation
result := lambda.Invoke(t, ctx, "my-function", payload)
lambda.AssertInvokeSuccess(t, result)

// With retry logic
result := lambda.InvokeWithRetry(t, ctx, "my-function", payload, 3, 2*time.Second)

// Async invocation
lambda.InvokeAsync(t, ctx, "my-function", payload)

// Function management
config := lambda.GetFunction(t, ctx, "my-function")
lambda.AssertEnvVarEquals(t, config, "ENVIRONMENT", "test")

// CloudWatch logs
logs := lambda.GetLogs(t, ctx, "my-function", time.Now().Add(-5*time.Minute))
lambda.AssertLogsContain(t, logs, "Processing completed")

// Event builders for common AWS services
s3Event := lambda.BuildS3Event("bucket", "key.txt", "s3:ObjectCreated:Put")
dynamoEvent := lambda.BuildDynamoDBEvent("table", "INSERT", testItem)
sqsEvent := lambda.BuildSQSEvent("queue-url", []string{"message1", "message2"})
```

### 🗄️ DynamoDB Testing (`vasdeference/dynamodb`)

Complete DynamoDB operations with backup/restore and transaction support:

```go
import "vasdeference/dynamodb"

// Basic CRUD operations
dynamodb.PutItem(t, ctx, "table", item)
item := dynamodb.GetItem(t, ctx, "table", dynamodb.CreateKey("PK", "SK"))
dynamodb.UpdateItem(t, ctx, "table", key, updateExpression, expressionValues)
dynamodb.DeleteItem(t, ctx, "table", key)

// Query and scan with pagination
items := dynamodb.QueryAllPages(t, ctx, "table", "PK = :pk", expressionValues)
allItems := dynamodb.ScanAllPages(t, ctx, "table")

// Batch operations
batchItems := []map[string]interface{}{item1, item2, item3}
dynamodb.BatchWriteItems(t, ctx, "table", batchItems)

keys := []map[string]interface{}{key1, key2}
results := dynamodb.BatchGetItems(t, ctx, "table", keys)

// Transactions
transactItems := []dynamodb.TransactWriteItem{
    dynamodb.NewTransactPut("table", item),
    dynamodb.NewTransactUpdate("table", key, updateExpr, exprValues),
}
dynamodb.TransactWriteItems(t, ctx, transactItems)

// Table management
dynamodb.CreateTable(t, ctx, "new-table", tableSchema)
dynamodb.WaitForTable(t, ctx, "new-table", dynamodb.TableStatusActive, 2*time.Minute)

// Comprehensive assertions
dynamodb.AssertTableExists(t, ctx, "table")
dynamodb.AssertItemExists(t, ctx, "table", key)
dynamodb.AssertItemCount(t, ctx, "table", 5)
dynamodb.AssertAttributeEquals(t, item, "Status", "active")
```

### 🌉 EventBridge Testing (`vasdeference/eventbridge`)

Event-driven architecture testing with pattern matching and custom buses:

```go
import "vasdeference/eventbridge"

// Send events
eventbridge.PutEvent(t, ctx, "custom-bus", "order.service", "Order Created", eventDetail)

// Batch events
events := []types.PutEventsRequestEntry{
    eventbridge.CreateEvent("order.service", "Order Created", orderDetail),
    eventbridge.CreateEvent("inventory.service", "Stock Updated", stockDetail),
}
eventbridge.PutEvents(t, ctx, events)

// Rule management
eventbridge.CreateRule(t, ctx, "order-rule", eventPattern, "custom-bus")
eventbridge.PutTargets(t, ctx, "order-rule", targets, "custom-bus")
eventbridge.EnableRule(t, ctx, "order-rule", "custom-bus")

// Pattern testing
pattern := eventbridge.EventPattern{
    Source:     []string{"order.service"},
    DetailType: []string{"Order Created", "Order Updated"},
    Detail: map[string]interface{}{
        "status": []string{"pending", "processing"},
    },
}
patternJSON := eventbridge.BuildEventPattern(t, pattern)
eventbridge.AssertEventPatternMatches(t, ctx, patternJSON, testEvent)
```

### 🔄 Step Functions Testing (`vasdeference/stepfunctions`)

Comprehensive workflow testing with execution analysis and debugging:

```go
import "vasdeference/stepfunctions"

// Basic execution
output := stepfunctions.StartExecution(t, ctx, stateMachineArn, "exec-1", input)
execution := stepfunctions.WaitForExecutionToComplete(t, ctx, output.ExecutionArn, 5*time.Minute)

// Execute and wait (combined)
execution := stepfunctions.ExecuteStateMachineAndWait(t, ctx, stateMachineArn, 
    "exec-1", input, 5*time.Minute)

// Input builders (fluent API)
input := stepfunctions.NewInput().
    Set("orderId", "ORDER-123").
    Set("customerId", "CUSTOMER-456").
    SetIf(isPremium, "discountCode", "PREMIUM20").
    Merge(shippingData).
    Build()

// Predefined input patterns
orderInput := stepfunctions.CreateOrderInput("ORDER-123", orderItems).Build()
retryInput := stepfunctions.CreateRetryableInput(true, 3).Build()

// Execution analysis and debugging
history := stepfunctions.GetExecutionHistory(t, ctx, execution.ExecutionArn)
analysis := stepfunctions.AnalyzeExecutionHistory(t, history)
failedSteps := stepfunctions.FindFailedSteps(t, history)

// Comprehensive assertions
stepfunctions.AssertExecutionSucceeded(t, execution)
stepfunctions.AssertExecutionOutput(t, execution, expectedOutput)
stepfunctions.AssertHistoryEventExists(t, history, types.HistoryEventTypeTaskSucceeded)
```

### ⚡ Parallel Testing (`vasdeference/parallel`)

Concurrent test execution with resource isolation and worker pools:

```go
import "vasdeference/parallel"

// Create parallel runner
runner, err := parallel.NewRunner(t, parallel.Options{
    PoolSize:          8,    // 8 concurrent workers
    ResourceIsolation: true, // Isolated resource namespaces
    TestTimeout:       5 * time.Minute,
})

// Define test cases
testCases := []parallel.TestCase{
    {
        Name: "TestUserRegistration",
        Func: func(ctx context.Context, t *parallel.TestContext) error {
            // Each test gets isolated resources
            tableName := t.GetTableName("users")    // → "test1-users"
            functionName := t.GetFunctionName("register") // → "test1-register"
            
            // Test implementation
            return testUserRegistration(tableName, functionName)
        },
    },
}

// Run tests in parallel
runner.Run(testCases)
```

### 📸 Snapshot Testing (`vasdeference/snapshot`)

Regression testing with golden files and intelligent sanitization:

```go
import "vasdeference/snapshot"

// Create snapshot tester
snap := snapshot.New(t, snapshot.Options{
    SnapshotDir: "testdata/snapshots",
    Sanitizers: []snapshot.Sanitizer{
        snapshot.SanitizeTimestamps(),
        snapshot.SanitizeUUIDs(),
        snapshot.SanitizeLambdaRequestID(),
    },
})

// Snapshot Lambda responses
output := lambda.Invoke(t, ctx, functionName, input)
snap.MatchLambdaResponse("lambda_response", output.Payload)

// Snapshot DynamoDB items
items := dynamodb.Scan(t, ctx, tableName)
snap.MatchDynamoDBItems("table_state", items)

// Snapshot Step Functions executions
execution := stepfunctions.ExecuteStateMachineAndWait(t, ctx, arn, name, input, timeout)
snap.MatchStepFunctionExecution("workflow_result", execution)

// Update snapshots when needed
// UPDATE_SNAPSHOTS=1 go test ./...
```

## 🧪 Complete Integration Example

```go
func TestComprehensiveServerlessWorkflow(t *testing.T) {
    // Initialize Vas Deference with automatic namespace detection
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()

    // Create snapshot tester for regression testing
    snap := snapshot.New(t, snapshot.Options{
        SnapshotDir: "testdata/integration_snapshots",
        Sanitizers: []snapshot.Sanitizer{
            snapshot.SanitizeTimestamps(),
            snapshot.SanitizeLambdaRequestID(),
            snapshot.SanitizeExecutionArn(),
        },
    })

    // Test Lambda functions
    orderResult := lambda.Invoke(t, vdf.GetTestContext(), "order-processor", orderRequest)
    lambda.AssertInvokeSuccess(t, orderResult)
    snap.MatchLambdaResponse("order_processor_response", orderResult.Payload)

    // Test DynamoDB operations
    dynamodb.PutItem(t, vdf.GetTestContext(), "orders", orderItem)
    item := dynamodb.GetItem(t, vdf.GetTestContext(), "orders", orderKey)
    dynamodb.AssertItemExists(t, item)

    // Test EventBridge events
    eventbridge.PutEvent(t, vdf.GetTestContext(), "order-events", 
        "order.service", "Order Created", orderDetail)

    // Test Step Functions workflow
    execution := stepfunctions.ExecuteStateMachineAndWait(t, vdf.GetTestContext(),
        orderWorkflowArn, "test-execution", workflowInput, 3*time.Minute)
    stepfunctions.AssertExecutionSucceeded(t, execution)
    snap.MatchStepFunctionExecution("order_workflow_execution", execution)
}
```

## 🚦 CI/CD Integration

### GitHub Actions

```yaml
name: Serverless Integration Tests
on:
  pull_request:
    types: [opened, synchronize]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1
      
      - name: Run Integration Tests
        run: go test -v ./...
        env:
          GITHUB_PR_NUMBER: ${{ github.event.pull_request.number }}
          
      - name: Update Snapshots (if needed)
        run: UPDATE_SNAPSHOTS=1 go test -v ./...
        if: contains(github.event.pull_request.labels.*.name, 'update-snapshots')
```

### Namespace Management

Vas Deference automatically detects the test environment and creates appropriate namespaces:

```go
// In CI/CD (GitHub Actions)
// GITHUB_PR_NUMBER=123 → namespace: "pr-123"

// Local development with GitHub CLI
// gh pr view --json number → namespace: "pr-456" 

// Manual override
vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDeferenceOptions{
    Namespace: "feature-branch-test",
})

// Resource naming
tableName := vdf.PrefixResourceName("users")  // → "pr-123-users"
functionName := vdf.PrefixResourceName("api") // → "pr-123-api"
```

## 📊 Best Practices

### 1. Always Use Arrange Pattern

```go
func TestMyFeature(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup() // Ensures proper resource cleanup
    
    // Test implementation
}
```

### 2. Leverage Terraform Integration

```go
// Get resource names from Terraform outputs
functionName := vdf.GetTerraformOutput(terraformOutputs, "lambda_function_name")
tableName := vdf.GetTerraformOutput(terraformOutputs, "dynamodb_table_name")

// Use prefixed names for test isolation
testTable := vdf.PrefixResourceName("test-data") // → "pr-123-test-data"
```

### 3. Use Table-Driven Tests for Comprehensive Coverage

```go
testCases := []struct {
    name     string
    input    interface{}
    expected interface{}
    wantErr  bool
}{
    {"ValidCase", validInput, expectedOutput, false},
    {"InvalidCase", invalidInput, nil, true},
    {"EdgeCase", edgeInput, edgeOutput, false},
}

for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
        // Test implementation using Vas Deference
    })
}
```

### 4. Implement Comprehensive Assertions

```go
// Lambda assertions
lambda.AssertFunctionExists(t, ctx, functionName)
lambda.AssertInvokeSuccess(t, result)
lambda.AssertPayloadContains(t, result, "expected_value")

// DynamoDB assertions
dynamodb.AssertTableExists(t, ctx, tableName)
dynamodb.AssertItemExists(t, ctx, tableName, key)
dynamodb.AssertItemCount(t, ctx, tableName, expectedCount)

// EventBridge assertions
eventbridge.AssertRuleExists(t, ctx, ruleName, eventBusName)
eventbridge.AssertEventPatternMatches(t, ctx, pattern, event)

// Step Functions assertions  
stepfunctions.AssertExecutionSucceeded(t, execution)
stepfunctions.AssertExecutionOutput(t, execution, expectedOutput)
```

### 5. Use Snapshot Testing for Regression Prevention

```go
snap := snapshot.New(t, snapshot.Options{
    Sanitizers: []snapshot.Sanitizer{
        snapshot.SanitizeTimestamps(),
        snapshot.SanitizeUUIDs(),
        snapshot.SanitizeLambdaRequestID(),
        snapshot.SanitizeCustom(`"requestId":"[^"]*"`, `"requestId":"<REQUEST_ID>"`),
    },
})

// Snapshot complex outputs
snap.MatchLambdaResponse("api_response", lambdaOutput.Payload)
snap.MatchDynamoDBItems("final_state", tableItems)
snap.MatchStepFunctionExecution("workflow_result", execution)
```

## 🔧 Configuration Options

### Core Configuration

```go
type VasDeferenceOptions struct {
    // AWS configuration
    Region     string        // AWS region (default: us-east-1)
    AWSConfig  *aws.Config   // Custom AWS configuration
    
    // Namespace and resource management
    Namespace  string        // Resource namespace (auto-detected by default)
    
    // Retry configuration
    MaxRetries int           // Maximum retry attempts (default: 3)
    RetryDelay time.Duration // Base retry delay (default: 100ms)
    
    // Terraform integration
    TerraformDir  string                 // Terraform directory path
    TerraformVars map[string]interface{} // Terraform variables
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/vasdeference.git
cd vasdeference

# Install dependencies
go mod download

# Run tests
go test -v ./...

# Run tests with AWS integration (requires AWS credentials)
AWS_PROFILE=test go test -v ./...

# Update snapshots
UPDATE_SNAPSHOTS=1 go test -v ./...
```

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gruntwork Terratest](https://terratest.gruntwork.io/) - Inspiration for API design patterns
- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) - Comprehensive AWS integration
- [Testify](https://github.com/stretchr/testify) - Testing assertions and utilities
- [Ants](https://github.com/panjf2000/ants) - High-performance goroutine pool

---

**Vas Deference** - *"Defer to the experts, defer to the testing."*

Making serverless testing as intuitive as writing the application itself, with comprehensive AWS coverage and Terratest-compatible patterns.

*For questions, issues, or contributions, please visit our [GitHub repository](https://github.com/yourusername/vasdeference).*