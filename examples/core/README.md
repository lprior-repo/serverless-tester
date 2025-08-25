# Core Framework

Basic VasDeference setup, configuration, and namespace management.

## Quick Start

```bash
# Run core examples
go test -v ./examples/core/

# View patterns
go run ./examples/core/examples.go
```

## Basic Setup

### Simple Initialization
```go
func TestBasicSetup(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Framework is ready to use
    functionName := vdf.PrefixResourceName("my-function")
    // Result: "pr-123-my-function" (with namespace)
}
```

### With Custom Options
```go
func TestCustomSetup(t *testing.T) {
    vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDefernOptions{
        Region:     "eu-west-1",
        Namespace:  "my-test",
        MaxRetries: 5,
        RetryDelay: 200 * time.Millisecond,
    })
    defer vdf.Cleanup()
    
    // Use custom configuration...
}
```

## AWS Client Creation

### Service Clients
```go
func TestAWSClients(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Create AWS service clients
    lambdaClient := vdf.CreateLambdaClient()
    dynamoClient := vdf.CreateDynamoDBClient()
    eventBridgeClient := vdf.CreateEventBridgeClient()
    sfnClient := vdf.CreateStepFunctionsClient()
    logsClient := vdf.CreateCloudWatchLogsClient()
    
    // All clients are configured and ready to use
}
```

### Multi-Region Clients
```go
func TestMultiRegion(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Default region client
    lambdaEast := vdf.CreateLambdaClient()
    
    // Specific region clients
    lambdaWest := vdf.CreateLambdaClientWithRegion("us-west-2")
    lambdaEU := vdf.CreateLambdaClientWithRegion("eu-west-1")
}
```

## Resource Management

### Resource Naming
```go
func TestResourceNaming(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Generate namespaced resource names
    functionName := vdf.PrefixResourceName("processor")
    tableName := vdf.PrefixResourceName("users")
    bucketName := vdf.PrefixResourceName("data")
    
    // All names include namespace for isolation
}
```

### Cleanup Registration
```go
func TestCleanupManagement(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Create resources
    resource1 := createMyResource()
    resource2 := createAnotherResource()
    
    // Register cleanup (LIFO order)
    vdf.RegisterCleanup(func() error {
        return resource2.Delete()
    })
    vdf.RegisterCleanup(func() error {
        return resource1.Delete()
    })
    
    // Cleanup happens automatically on defer
}
```

## Terraform Integration

### Extract Outputs
```go
func TestTerraformOutputs(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    outputs := map[string]interface{}{
        "lambda_function_name": "my-function-abc123",
        "api_gateway_url":     "https://api.example.com",
        "dynamodb_table_name": "users-table",
    }
    
    // Extract values
    functionName := vdf.GetTerraformOutput(outputs, "lambda_function_name")
    apiUrl := vdf.GetTerraformOutput(outputs, "api_gateway_url") 
    tableName := vdf.GetTerraformOutput(outputs, "dynamodb_table_name")
    
    // Use extracted values in tests...
}
```

### With Error Handling
```go
func TestTerraformErrorHandling(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    outputs := map[string]interface{}{
        "required_field": "value",
    }
    
    // Required field (panics if not found)
    required := vdf.GetTerraformOutput(outputs, "required_field")
    
    // Optional field (returns error if not found)
    optional, err := vdf.GetTerraformOutputE(outputs, "optional_field")
    if err != nil {
        t.Log("Optional field not found, using default")
        optional = "default_value"
    }
}
```

## Environment Utilities

### Environment Variables
```go
func TestEnvironmentHandling(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Get environment variables with defaults
    region := vdf.GetEnvVarWithDefault("AWS_REGION", "us-east-1")
    profile := vdf.GetEnvVarWithDefault("AWS_PROFILE", "default")
    
    // CI environment detection
    if vdf.IsCIEnvironment() {
        t.Log("Running in CI/CD environment")
    }
}
```

### Credential Validation
```go
func TestCredentialValidation(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Validate AWS credentials
    if !vdf.ValidateAWSCredentials() {
        t.Skip("AWS credentials not configured")
    }
    
    // Continue with tests requiring AWS access...
}
```

## Utilities

### Unique ID Generation
```go
func TestUniqueIDs(t *testing.T) {
    // Generate unique identifiers
    id1 := vasdeference.UniqueId()
    id2 := vasdeference.UniqueId()
    
    assert.NotEqual(t, id1, id2)
    
    // Random strings
    randomStr := vasdeference.RandomString(8)
    assert.Len(t, randomStr, 8)
}
```

### Namespace Sanitization
```go
func TestNamespaceSanitization(t *testing.T) {
    // Sanitize names for AWS resources
    clean1 := vasdeference.SanitizeNamespace("test_with_underscores")
    assert.Equal(t, "test-with-underscores", clean1)
    
    clean2 := vasdeference.SanitizeNamespace("feature/branch-name")
    assert.Equal(t, "feature-branch-name", clean2)
}
```

### Region Management
```go
func TestRegionUtilities(t *testing.T) {
    // Select random region for testing
    region := vasdeference.SelectRegion()
    assert.Contains(t, []string{"us-east-1", "us-west-2", "eu-west-1"}, region)
    
    // Use predefined constants
    assert.Equal(t, "us-east-1", vasdeference.RegionUSEast1)
    assert.Equal(t, "us-west-2", vasdeference.RegionUSWest2)
    assert.Equal(t, "eu-west-1", vasdeference.RegionEUWest1)
}
```

## Advanced Patterns

### Complete Integration Test
```go
func TestCompleteIntegration(t *testing.T) {
    vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDefernOptions{
        Region:     vasdeference.SelectRegion(),
        Namespace:  "integration-" + vasdeference.UniqueId(),
        MaxRetries: 3,
    })
    defer vdf.Cleanup()
    
    // Skip if no AWS credentials
    if !vdf.ValidateAWSCredentials() {
        t.Skip("AWS credentials required")
    }
    
    // Create all required resources
    functionName := vdf.PrefixResourceName("processor")
    tableName := vdf.PrefixResourceName("data")
    
    // Set up infrastructure
    // ... deploy and test
    
    // Cleanup is automatic
}
```

### Performance Test Setup
```go
func TestPerformanceSetup(t *testing.T) {
    vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDefernOptions{
        Region:     "us-east-1", // Fixed region for performance tests
        Namespace:  "perf-" + vasdeference.UniqueId(),
        MaxRetries: 1, // Fast failure for performance tests
        RetryDelay: 50 * time.Millisecond,
    })
    defer vdf.Cleanup()
    
    // Performance test setup...
}
```

## Error Scenarios

### Framework Initialization Errors
```go
func TestInitializationErrors(t *testing.T) {
    // Test nil parameter
    vdf, err := vasdeference.NewWithOptionsE(nil, &vasdeference.VasDefernOptions{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "TestingT cannot be nil")
}
```

### Configuration Errors  
```go
func TestConfigurationErrors(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    // Test invalid Terraform outputs
    _, err := vdf.GetTerraformOutputE(nil, "key")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "outputs map cannot be nil")
    
    // Test missing key
    outputs := map[string]interface{}{"other": "value"}
    _, err = vdf.GetTerraformOutputE(outputs, "missing")
    assert.Error(t, err)
}
```

## Run Examples

```bash
# View all patterns
go run ./examples/core/examples.go

# Run as tests
go test -v ./examples/core/

# Test specific patterns
go test -v ./examples/core/ -run TestBasicSetup
```

## API Reference

### Initialization
- `vasdeference.New(t)` - Basic setup
- `vasdeference.NewWithOptions(t, opts)` - Custom configuration
- `vasdeference.NewWithOptionsE(t, opts)` - With error handling

### AWS Clients
- `vdf.CreateLambdaClient()` - Lambda client
- `vdf.CreateDynamoDBClient()` - DynamoDB client  
- `vdf.CreateEventBridgeClient()` - EventBridge client
- `vdf.CreateStepFunctionsClient()` - Step Functions client
- `vdf.CreateCloudWatchLogsClient()` - CloudWatch Logs client

### Resource Management
- `vdf.PrefixResourceName(name)` - Namespace prefixing
- `vdf.RegisterCleanup(func)` - Register cleanup function
- `vdf.Cleanup()` - Execute all cleanups

### Terraform Integration
- `vdf.GetTerraformOutput(outputs, key)` - Extract value (panics on error)
- `vdf.GetTerraformOutputE(outputs, key)` - Extract value (returns error)

---

**Next:** [Lambda Examples](../lambda/) | [DynamoDB Examples](../dynamodb/)