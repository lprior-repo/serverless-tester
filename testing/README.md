# Vas Deference Testing Infrastructure

This directory contains comprehensive testing infrastructure for the Vas Deference framework, designed to solve the critical challenge of testing a testing framework without breaking the tests themselves.

## Overview

The testing infrastructure provides:

1. **MockTestingT**: A mock implementation of the TestingT interface for testing framework behavior
2. **Test Utilities**: Comprehensive utilities for AWS client mocking, test data generation, and validation
3. **Test Fixtures**: Reusable test data and AWS event samples
4. **Framework Validation**: Tests that validate the framework works correctly without interfering with test execution

## Key Features

### MockTestingT (mock.go)

The `MockTestingT` struct provides a thread-safe mock implementation of the `TestingT` interface:

```go
mockT := testutils.NewMockTestingT("test-name")

// Use like any TestingT
mockT.Helper()
mockT.Errorf("Test error: %s", "message")
mockT.FailNow()

// Verify behavior without breaking tests
assert.True(t, mockT.WasErrorfCalled())
assert.Equal(t, 2, mockT.GetErrorfCallCount())
```

**Key Methods:**
- `NewMockTestingT(name)` - Create new mock instance
- `WasHelperCalled()`, `WasErrorfCalled()`, etc. - Check if methods were called
- `GetErrorMessages()`, `GetLogMessages()` - Retrieve captured messages
- `Reset()` - Clear all state for reuse
- Thread-safe for concurrent testing

### Test Utilities (utils.go)

Comprehensive testing utilities including:

**AWS Client Mocking:**
```go
utils := testutils.NewTestUtils()
mockLambda := utils.CreateMockLambdaClient()
mockDynamoDB := utils.CreateMockDynamoDBClient()
```

**Test Data Generation:**
```go
randomString := utils.GenerateRandomString(10)
uuid := utils.GenerateUUID()
number := utils.GenerateRandomNumber(1, 100)
```

**Assertion Helpers:**
```go
utils.AssertStringNotEmpty(t, value, "should not be empty")
utils.AssertJSONValid(t, jsonString, "should be valid JSON")
utils.AssertAWSArn(t, arn, "should be valid ARN")
```

**Environment Management:**
```go
envManager := utils.CreateEnvironmentManager()
envManager.SetEnv("TEST_VAR", "value")
envManager.RestoreAll() // Restore original values
```

**Retry Utilities:**
```go
err := utils.RetryOperation(func() error {
    // Your potentially flaky operation
    return doSomething()
}, 3, 100*time.Millisecond)
```

### Test Fixtures (fixtures.go)

Reusable test data and AWS event samples:

```go
fixtures := testutils.NewTestFixtures()
// or with custom namespace
fixtures := testutils.NewTestFixturesWithNamespace("custom-ns")

// Get sample data
lambdaNames := fixtures.GetLambdaFunctionNames()
s3Event := fixtures.GetSampleS3Event()
terraformOutputs := fixtures.GetSampleTerraformOutputs()
```

**Available Fixtures:**
- AWS Lambda function names
- DynamoDB table names
- EventBridge rule names
- Step Function names
- Sample AWS events (S3, SQS, API Gateway)
- Terraform outputs structure
- Environment variables
- Lambda context data

### Framework Testing (framework_test.go)

Tests that validate the framework behavior without breaking test execution:

- **Error Scenario Testing**: Test framework error handling without failing tests
- **Namespace Sanitization**: Validate namespace detection and cleaning
- **Cleanup Lifecycle**: Test LIFO cleanup execution
- **Terraform Output Parsing**: Validate output parsing and error handling
- **Parallel Testing**: Ensure thread-safe operation

### Compatibility Testing (compatibility_test.go)

Critical tests that ensure the framework works with real `*testing.T`:

- **Interface Compatibility**: Verify `*testing.T` implements `TestingT`
- **Error Isolation**: Ensure framework errors don't break test runner
- **Concurrent Usage**: Validate thread safety
- **Predictable Behavior**: Test framework behavior patterns

## Usage Examples

### Testing Framework Error Handling

```go
func TestMyFrameworkFunction(t *testing.T) {
    // Use mock to test error scenarios without failing the test
    mockT := testutils.NewMockTestingT("error-test")
    
    // Call framework function that might error
    myFrameworkFunction(mockT, invalidConfig)
    
    // Verify error was handled correctly
    assert.True(t, mockT.WasErrorfCalled())
    assert.Contains(t, mockT.GetErrorMessages()[0], "expected error message")
}
```

### Testing with Real Testing Interface

```go
func TestFrameworkWithRealT(t *testing.T) {
    // Your framework should work with real *testing.T
    result := myFrameworkFunction(t, validConfig)
    assert.True(t, result)
}
```

### Comprehensive Test Setup

```go
func TestComplexScenario(t *testing.T) {
    // Setup test environment
    utils := testutils.NewTestUtils()
    fixtures := testutils.NewTestFixturesWithNamespace("integration-test")
    envManager := utils.CreateEnvironmentManager()
    
    // Configure test environment
    envManager.SetEnv("AWS_REGION", "us-east-1")
    defer envManager.RestoreAll()
    
    // Get test data
    terraformOutputs := fixtures.GetSampleTerraformOutputs()
    
    // Test the framework
    mockT := testutils.NewMockTestingT("complex-test")
    result := frameworkOperation(mockT, terraformOutputs)
    
    // Verify behavior
    assert.NoError(t, result)
    assert.False(t, mockT.WasErrorfCalled())
}
```

## Testing Philosophy

This infrastructure follows these key principles:

1. **Test Isolation**: Framework errors should never break the test runner
2. **TDD Compliance**: All functionality was built using strict Red-Green-Refactor cycles
3. **Thread Safety**: All components support concurrent usage
4. **Interface Compatibility**: Full compatibility with Go's `testing.T` interface
5. **Comprehensive Coverage**: Every aspect of framework behavior is testable

## File Structure

```
/home/family/sfx/testing/
├── README.md              # This documentation
├── mock.go                # MockTestingT implementation
├── mock_test.go           # MockTestingT tests
├── utils.go               # Testing utilities
├── utils_test.go          # Utilities tests
├── fixtures.go            # Test data and fixtures
├── fixtures_test.go       # Fixtures tests
├── framework_test.go      # Framework behavior tests
└── compatibility_test.go  # Interface compatibility tests
```

## Integration with Main Framework

The testing infrastructure integrates seamlessly with the main Vas Deference framework:

- All tests validate that the framework can be tested without breaking other tests
- MockTestingT implements the same TestingT interface used by the framework
- Test utilities provide mocks for AWS services used by the framework
- Fixtures provide realistic test data matching framework expectations

This ensures that the Vas Deference framework can be thoroughly validated and that users can test their own code that uses the framework without any interference with their test execution.