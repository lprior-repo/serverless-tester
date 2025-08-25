# 🚀 Vas Deference Framework - Deployment Guide

## Quick Start

### Prerequisites
```bash
# Install required tools
go install go.uber.org/mock/mockgen@latest
go install github.com/gruntwork-io/terratest@latest

# Verify Go version
go version  # Requires Go 1.24+
```

### Installation
```bash
# Add to your go.mod
go get github.com/your-org/vasdeference

# Or clone and install locally
git clone https://github.com/your-org/vasdeference.git
cd vasdeference
go mod tidy
```

## 📦 Package Overview

### 🎯 Production-Ready Packages (90%+ Coverage)

#### **Snapshot Package (95.2%)**
```go
import "github.com/your-org/vasdeference/snapshot"

// Compare AWS resource states
snapshot := snapshot.New(t)
snapshot.MatchJSON("lambda_config", actualResponse)
```

#### **Parallel Package (91.9%)**
```go
import "github.com/your-org/vasdeference/parallel"

// Run tests in parallel with resource isolation
runner := parallel.NewRunner(t, parallel.WithPoolSize(10))
runner.RunLambdaTests(testCases)
```

### 🔧 Foundation Packages (Ready for Use)

#### **Testing Package (83.5%)**
```go
import "github.com/your-org/vasdeference/testing"

// Environment management and utilities
env := testing.NewEnvironmentManager()
env.Set("AWS_REGION", "us-east-1")
defer env.RestoreAll()
```

#### **Core Package (56.0%)**
```go
import "github.com/your-org/vasdeference"

// Main framework initialization
vdf := vasdeference.New(t)
defer vdf.Cleanup()

client := vdf.CreateLambdaClient()
```

#### **AWS Service Packages**
```go
// Lambda testing
import "github.com/your-org/vasdeference/lambda"
lambda.AssertFunctionExists(t, "my-function")

// DynamoDB testing  
import "github.com/your-org/vasdeference/dynamodb"
dynamodb.AssertTableExists(t, "my-table")

// EventBridge testing
import "github.com/your-org/vasdeference/eventbridge" 
eventbridge.AssertRuleExists(t, "my-rule")

// Step Functions testing
import "github.com/your-org/vasdeference/stepfunctions"
stepfunctions.AssertStateMachineExists(t, "my-state-machine")
```

## 🛠️ Configuration

### Environment Variables
```bash
# Required
export AWS_REGION=us-east-1
export AWS_PROFILE=your-profile

# Optional  
export VDF_NAMESPACE=test-$(date +%s)
export VDF_CLEANUP_ENABLED=true
export VDF_PARALLEL_WORKERS=10
```

### AWS Credentials
```bash
# Option 1: AWS CLI
aws configure

# Option 2: Environment variables
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret

# Option 3: IAM roles (recommended for CI/CD)
# Attach appropriate IAM policies
```

### Required IAM Permissions
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:*",
        "dynamodb:*", 
        "events:*",
        "states:*",
        "logs:*"
      ],
      "Resource": "*"
    }
  ]
}
```

## 📋 Usage Patterns

### Basic Testing Pattern
```go
func TestMyServerlessApp(t *testing.T) {
    // Initialize framework
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Test Lambda function
    lambda.AssertFunctionExists(t, "my-function")
    result := lambda.Invoke(t, "my-function", `{"test": "data"}`)
    
    // Test DynamoDB table
    dynamodb.AssertTableExists(t, "my-table")
    dynamodb.AssertItemExists(t, "my-table", map[string]interface{}{
        "id": "test-id",
    })
    
    // Test EventBridge rules
    eventbridge.AssertRuleExists(t, "my-rule")
}
```

### Parallel Testing Pattern
```go
func TestParallelServerlessOperations(t *testing.T) {
    runner := parallel.NewRunner(t, 
        parallel.WithPoolSize(5),
        parallel.WithNamespace("load-test"),
    )
    
    testCases := []parallel.LambdaTestCase{
        {Name: "test-1", Payload: `{"id": 1}`},
        {Name: "test-2", Payload: `{"id": 2}`},
        {Name: "test-3", Payload: `{"id": 3}`},
    }
    
    results := runner.RunLambdaTests("my-function", testCases)
    assert.Len(t, results, 3)
}
```

### Snapshot Testing Pattern
```go
func TestLambdaConfigSnapshot(t *testing.T) {
    snap := snapshot.New(t)
    
    // Capture current Lambda configuration
    config := getLambdaConfiguration("my-function")
    
    // Compare against stored snapshot
    snap.MatchJSON("lambda_config", config)
    
    // Update snapshot if needed with -update flag
    // go test -update
}
```

### Advanced Workflow Testing
```go
func TestCompleteWorkflow(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // 1. Test EventBridge event publishing
    event := eventbridge.BuildS3Event("my-bucket", "test-key")
    eventbridge.PutEvent(t, "my-bus", event)
    
    // 2. Verify Lambda function triggered
    time.Sleep(2 * time.Second)
    logs := lambda.GetLogsByRequestID(t, "my-function", requestID)
    assert.Contains(t, logs, "Processing S3 event")
    
    // 3. Verify DynamoDB item created
    dynamodb.AssertItemExists(t, "my-table", map[string]interface{}{
        "bucket": "my-bucket",
        "key": "test-key", 
    })
    
    // 4. Test Step Functions execution
    input := stepfunctions.BuildInput().
        Add("bucket", "my-bucket").
        Add("key", "test-key").
        Build()
        
    execution := stepfunctions.StartExecution(t, "my-state-machine", input)
    stepfunctions.WaitForExecution(t, execution.ExecutionArn, 30*time.Second)
    stepfunctions.AssertExecutionSucceeded(t, execution.ExecutionArn)
}
```

## 🎯 Testing Strategies

### Unit Testing
```go
// Test individual functions with mocks
func TestLambdaFunction(t *testing.T) {
    mockClient := lambda.NewMockClient()
    mockClient.On("Invoke", mock.Anything).Return(validResponse, nil)
    
    result := lambda.InvokeWithClient(t, mockClient, "function", payload)
    assert.NotNil(t, result)
}
```

### Integration Testing
```go
// Test real AWS resources
func TestIntegrationWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Test against real AWS services
    lambda.AssertFunctionExists(t, "my-function")
}
```

### Performance Testing
```go
// Load testing with parallel execution
func TestLoadTesting(t *testing.T) {
    runner := parallel.NewRunner(t, parallel.WithPoolSize(20))
    
    var testCases []parallel.TestCase
    for i := 0; i < 100; i++ {
        testCases = append(testCases, parallel.TestCase{
            Name: fmt.Sprintf("load-test-%d", i),
            // Test logic
        })
    }
    
    results := runner.RunTests(testCases)
    
    // Analyze performance results
    avgDuration := calculateAverageDuration(results)
    assert.Less(t, avgDuration, 500*time.Millisecond)
}
```

## 🔧 CI/CD Integration

### GitHub Actions
```yaml
name: Serverless Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.24'
        
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v1
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1
        
    - name: Run tests
      run: |
        go test -v ./...
        
    - name: Run integration tests
      run: |
        go test -v -tags=integration ./...
        
    - name: Generate coverage
      run: |
        go test -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
```

### GitLab CI
```yaml
stages:
  - test
  - integration

test:
  stage: test
  image: golang:1.24
  script:
    - go mod download
    - go test -short -v ./...
    
integration:
  stage: integration
  image: golang:1.24
  script:
    - export AWS_REGION=us-east-1
    - go test -v ./...
  only:
    - main
```

## 📊 Monitoring and Observability

### Test Metrics Collection
```go
func TestWithMetrics(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Enable metrics collection
    vdf.EnableMetrics()
    
    // Run tests...
    
    // Report metrics
    metrics := vdf.GetMetrics()
    t.Logf("Test duration: %v", metrics.Duration)
    t.Logf("AWS API calls: %d", metrics.APICallCount)
}
```

### Log Analysis
```go
func TestWithLogging(t *testing.T) {
    // Enable structured logging
    vdf := vasdeference.New(t, vasdeference.WithLogging(slog.LevelDebug))
    defer vdf.Cleanup()
    
    // All operations will be logged with context
}
```

## 🚨 Troubleshooting

### Common Issues

#### **AWS Credentials**
```bash
# Check credentials
aws sts get-caller-identity

# Check region
echo $AWS_REGION
```

#### **Resource Cleanup**
```go
// Ensure proper cleanup
defer func() {
    if r := recover(); r != nil {
        vdf.Cleanup() // Force cleanup on panic
        panic(r)
    }
}()
```

#### **Test Timeouts**
```go
// Increase timeout for slow operations
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result := lambda.InvokeWithContext(ctx, t, "function", payload)
```

#### **Parallel Test Conflicts**
```go
// Use unique namespaces
vdf := vasdeference.New(t, vasdeference.WithNamespace(
    fmt.Sprintf("test-%d", time.Now().Unix()),
))
```

## 📚 API Reference

See the complete API documentation in the `/examples` directory:
- [Core Framework Examples](./examples/core/README.md)
- [Lambda Testing Examples](./examples/lambda/README.md) 
- [DynamoDB Testing Examples](./examples/dynamodb/README.md)
- [EventBridge Testing Examples](./examples/eventbridge/README.md)
- [Step Functions Testing Examples](./examples/stepfunctions/README.md)

## 🤝 Support

- **Documentation**: [examples/](./examples/)
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions

---

*Ready to build reliable serverless applications with confidence!*