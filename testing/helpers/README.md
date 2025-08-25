# Terratest-style Test Helpers

This package provides reusable test helpers following Terratest patterns for comprehensive testing of AWS resources and infrastructure. All helpers are implemented using strict Test-Driven Development (TDD) principles with 100% test coverage.

## Features

- **Random ID Generation** - Unique identifiers and resource names
- **Retry Mechanisms** - Robust retry logic with exponential backoff
- **AWS Resource Validation** - Comprehensive AWS resource assertions
- **Cleanup Patterns** - Proper resource cleanup following LIFO order
- **Test Fixtures** - Terratest-style configuration management
- **Error Handling & Assertions** - Rich assertion library with logging

## Core Modules

### 1. Random Generation (`random.go`)

Generates unique identifiers and resource names for testing:

```go
// Generate unique UUID
id := UniqueID()

// Generate random strings
randomStr := RandomString(10)
lowerStr := RandomStringLower(8)

// Generate AWS-compliant resource names
resourceName := RandomAWSResourceName("test", "lambda")

// Deterministic testing
SeedRandom(12345)
```

### 2. Retry Mechanisms (`retry.go`)

Robust retry logic following Terratest patterns:

```go
// Simple retry
err := DoWithRetry("operation", 3, time.Second, func() error {
    return someOperation()
})

// Retry with return value
result, err := DoWithRetryE("fetch-data", 5, time.Second, func() (string, error) {
    return fetchData()
})

// Context-aware retry
err = DoWithRetryWithContext(ctx, "operation", 3, time.Second, func() error {
    return someOperation()
})

// Exponential backoff
err = DoWithRetryExponentialBackoff("operation", 4, 100*time.Millisecond, 2.0, func() error {
    return someOperation()
})
```

### 3. AWS Resource Validation (`aws.go`)

Comprehensive AWS resource validation and assertions:

```go
// Validate AWS resource names and ARNs
valid := ValidateAWSResourceName("my-lambda-function")
valid = ValidateAWSArn("arn:aws:lambda:us-east-1:123456789012:function:my-function")

// Assert resource existence
err := AssertDynamoDBTableExists(ctx, dynamoClient, "my-table")
err = AssertLambdaFunctionExists(ctx, lambdaClient, "my-function")

// With retry logic
err = AssertDynamoDBTableExistsE(ctx, dynamoClient, "my-table", 3)
err = AssertLambdaFunctionExistsE(ctx, lambdaClient, "my-function", 5)

// Extract information from ARNs
resourceType := GetAWSResourceType(arn)  // "table", "function", etc.
resourceName := GetAWSResourceName(arn)  // actual resource name
```

### 4. Cleanup Management (`cleanup.go`)

Proper cleanup patterns following LIFO (Last In, First Out) order:

```go
// Create cleanup manager
cleanup := NewCleanupManager()

// Register cleanup functions
cleanup.Register("database", func() error {
    return database.Close()
})

cleanup.Register("server", func() error {
    return server.Stop()
})

// Execute all cleanups in reverse order (server first, then database)
err := cleanup.ExecuteAll()

// Defer cleanup with test framework
DeferCleanup(t, "resource-cleanup", func() error {
    return cleanupResource()
})

// Terraform cleanup wrapper
terraformCleanup := TerraformCleanupWrapper("my-infrastructure", func() {
    terraform.Destroy(t, options)
})

// AWS resource cleanup wrapper
awsCleanup := AWSResourceCleanupWrapper("my-table", "dynamodb", func() error {
    return deleteTable()
})
```

### 5. Test Fixtures (`fixtures.go`)

Terratest-style configuration and fixture management:

```go
// Create Terraform options
options := CreateTerraformOptions("/path/to/terraform", map[string]interface{}{
    "region": "us-west-2",
    "name":   "test-resource",
})

// With backend configuration
options = CreateTerraformOptionsWithBackend(
    "/path/to/terraform",
    vars,
    map[string]interface{}{
        "bucket": "my-terraform-state",
        "key":    "test.tfstate",
    },
)

// AWS test configuration
awsOptions := AWSTestOptions("us-east-1", "default")

// Generate test resource names
resourceName := GenerateTestResourceName("integration", "lambda")

// Test configuration management
config := TestConfig{
    AWSRegion:      "us-west-2",
    AWSProfile:     "default",
    Environment:    "test",
    ResourcePrefix: "terratest",
}

config = ApplyTestConfigDefaults(config)
err := ValidateTestConfig(config)

// Test tags
tags := CreateTestTags("TestMyFunction", "integration")
merged := MergeTestTags(baseTags, additionalTags)
```

### 6. Assertions (`assertions.go`)

Rich assertion library with structured logging:

```go
// Error assertions
AssertNoError(t, err, "operation should succeed")
AssertError(t, err, "operation should fail")
AssertErrorContains(t, err, "not found", "should contain specific text")

// String assertions
AssertStringNotEmpty(t, value, "value should not be empty")
AssertStringContains(t, str, "expected", "should contain text")

// Numeric assertions
AssertGreaterThan(t, 10, 5, "value should be greater")
AssertLessThan(t, 5, 10, "value should be less")

// Equality assertions
AssertEqual(t, expected, actual, "values should be equal")
AssertNotEqual(t, notExpected, actual, "values should be different")

// Boolean assertions
AssertTrue(t, condition, "condition should be true")
AssertFalse(t, condition, "condition should be false")
```

## Usage Patterns

### Basic Resource Testing

```go
func TestLambdaFunction(t *testing.T) {
    // Generate unique resource name
    functionName := RandomAWSResourceName("test", "lambda")
    
    // Setup cleanup
    cleanup := NewCleanupManager()
    defer func() {
        if err := cleanup.ExecuteAll(); err != nil {
            t.Logf("Cleanup failed: %v", err)
        }
    }()
    
    // Create resource with retry
    err := DoWithRetry("create-function", 3, time.Second, func() error {
        return createLambdaFunction(functionName)
    })
    AssertNoError(t, err, "function creation should succeed")
    
    // Register cleanup
    cleanup.Register("lambda-function", AWSResourceCleanupWrapper(
        functionName, "lambda", func() error {
            return deleteLambdaFunction(functionName)
        },
    ))
    
    // Assert function exists with retry
    err = AssertLambdaFunctionExistsE(ctx, lambdaClient, functionName, 5)
    AssertNoError(t, err, "function should exist")
    
    // Test function behavior...
}
```

### Integration Testing with Terraform

```go
func TestTerraformInfrastructure(t *testing.T) {
    // Create Terraform options
    options := CreateTerraformOptions("./terraform", map[string]interface{}{
        "environment": "test",
        "region":      "us-west-2",
    })
    
    // Setup cleanup
    cleanup := NewCleanupManager()
    defer func() {
        cleanup.ExecuteAll()
    }()
    
    // Deploy infrastructure
    terraform.InitAndApply(t, options)
    cleanup.Register("terraform", TerraformCleanupWrapper("infrastructure", func() {
        terraform.Destroy(t, options)
    }))
    
    // Get outputs
    tableName := terraform.Output(t, options, "table_name")
    functionName := terraform.Output(t, options, "function_name")
    
    // Assert resources exist
    AssertDynamoDBTableExistsE(ctx, dynamoClient, tableName, 5)
    AssertLambdaFunctionExistsE(ctx, lambdaClient, functionName, 5)
    
    // Test functionality...
}
```

### Configuration-Driven Testing

```go
func TestWithConfiguration(t *testing.T) {
    // Load test configuration
    config := TestConfig{
        AWSRegion: "us-west-2",
        Environment: "test",
    }
    config = ApplyTestConfigDefaults(config)
    
    err := ValidateTestConfig(config)
    AssertNoError(t, err, "configuration should be valid")
    
    // Create AWS options
    awsOptions := AWSTestOptions(config.AWSRegion, config.AWSProfile)
    
    // Generate resource name
    resourceName := GenerateTestResourceName(config.ResourcePrefix, "dynamodb")
    
    // Create tags
    tags := CreateTestTags(t.Name(), config.Environment)
    
    // Use configuration for testing...
}
```

## Key Design Principles

1. **Test-Driven Development**: All code written following strict TDD cycles
2. **Function/FunctionE Pattern**: Consistent error handling patterns from Terratest
3. **LIFO Cleanup**: Resources cleaned up in reverse order of creation
4. **Structured Logging**: All operations logged with structured context using slog
5. **Context Awareness**: Support for cancellation and timeouts
6. **Type Safety**: Strong typing with generic support where appropriate
7. **Composability**: Small, focused functions that compose well

## Testing

All helpers have comprehensive test coverage:

```bash
# Run all helper tests
go test ./testing/helpers -v

# Run specific test groups
go test ./testing/helpers -run TestRetry -v
go test ./testing/helpers -run TestAWS -v
go test ./testing/helpers -run TestCleanup -v
```

## Dependencies

- **Terratest**: Core retry and random utilities
- **AWS SDK v2**: AWS service clients
- **slog**: Structured logging
- **testify**: Test assertions
- **uuid**: Unique identifier generation

## Integration

These helpers integrate seamlessly with existing testing frameworks and can be used alongside:

- Standard Go testing
- Terratest
- Custom testing frameworks
- CI/CD pipelines

The helpers are designed to be framework-agnostic while following Terratest conventions for maximum compatibility.