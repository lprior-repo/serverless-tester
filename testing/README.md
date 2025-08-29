# SFX Functional Testing Infrastructure

This directory contains comprehensive functional testing infrastructure for the SFX framework, built using pure functional programming principles with **samber/lo** and **samber/mo** to solve the critical challenge of testing a testing framework without breaking the tests themselves.

## Functional Testing Overview

The functional testing infrastructure provides:

1. **Immutable MockTestingT**: Functional mock implementation with zero mutations and monadic state management
2. **Functional Test Utilities**: Pure functional utilities for AWS client mocking, immutable test data generation, and composable validation
3. **Immutable Test Fixtures**: Type-safe, reusable test data and AWS event samples using functional options pattern
4. **Functional Framework Validation**: Monadic validation tests that ensure framework correctness without test interference

## Key Features

### Functional MockTestingT (mock.go)

The `FunctionalMockTestingT` struct provides an immutable, thread-safe mock implementation using functional programming principles:

```go
// Create immutable mock configuration
mockConfig := NewFunctionalMockConfig(
    WithTestName("test-name"),
    WithThreadSafety(true),
    WithStateTracking(true),
)

// Create functional mock instance
mockT := CreateFunctionalMockTestingT(mockConfig)

// Use with immutable state transitions
updatedMock := mockT.
    CallHelper().
    CallErrorf("Test error: %s", "message").
    CallFailNow()

// Verify behavior functionally with monadic queries
helperCalled := QueryFunctionalMockState(updatedMock, WasHelperCalled)
errorfCount := QueryFunctionalMockState(updatedMock, GetErrorfCallCount)
```

**Functional Key Features:**
- `CreateFunctionalMockTestingT(config)` - Immutable mock instance creation
- `QueryFunctionalMockState(mock, query)` - Pure function state queries with `mo.Option[T]`
- `GetFunctionalMessages(mock)` - Immutable message retrieval using `lo.Map`
- `ResetFunctionalMock(mock)` - Returns new mock instance with cleared state
- Thread-safe through immutability and functional state management

### Functional Test Utilities (utils.go)

Comprehensive functional testing utilities using immutable patterns:

**Functional AWS Client Mocking:**
```go
// Create immutable utility configuration
utilsConfig := NewFunctionalTestUtilsConfig(
    WithAWSMockingEnabled(true),
    WithDataGeneration(true),
    WithAssertionHelpers(true),
)

// Create immutable AWS client mocks
clientMocks := CreateFunctionalAWSMocks(utilsConfig, []AWSService{
    LambdaService, DynamoDBService, S3Service,
})

// Access mocks functionally
lambdaMock := GetFunctionalMock(clientMocks, LambdaService)
dynamoMock := GetFunctionalMock(clientMocks, DynamoDBService)
```

**Functional Test Data Generation:**
```go
// Generate immutable test data using functional options
testDataConfig := NewFunctionalDataGenerationConfig(
    WithStringLength(10),
    WithNumberRange(1, 100),
    WithUUIDGeneration(true),
)

generatedData := GenerateFunctionalTestData(testDataConfig).
    Map(func(data TestData) TestData {
        return ValidateTestData(data)
    })
```

**Functional Assertion Helpers:**
```go
// Create immutable assertion configuration
assertionConfig := NewFunctionalAssertionConfig(
    WithStringValidation(),
    WithJSONValidation(),
    WithAWSArnValidation(),
)

// Execute functional assertions with monadic results
validationResults := []mo.Result[AssertionResult]{
    AssertFunctionalStringNotEmpty(assertionConfig, value, "should not be empty"),
    AssertFunctionalJSONValid(assertionConfig, jsonString, "should be valid JSON"),
    AssertFunctionalAWSArn(assertionConfig, arn, "should be valid ARN"),
}

// Aggregate assertion results
aggregatedResult := AggregateFunctionalAssertions(validationResults)
```

**Functional Environment Management:**
```go
// Create immutable environment configuration
envConfig := NewFunctionalEnvironmentConfig(
    WithVariable("TEST_VAR", "value"),
    WithVariable("AWS_REGION", "us-east-1"),
    WithRestoreOnCleanup(true),
)

// Apply environment changes functionally
environmentResult := ApplyFunctionalEnvironment(envConfig).
    Map(func(env FunctionalEnvironment) FunctionalEnvironment {
        return ValidateEnvironmentState(env)
    })
```

**Functional Retry Utilities:**
```go
// Create immutable retry configuration
retryConfig := NewFunctionalRetryConfig(
    WithMaxAttempts(3),
    WithBackoffDelay(100*time.Millisecond),
    WithExponentialBackoff(true),
)

// Execute operation with functional retry logic
result := ExecuteFunctionalRetry(retryConfig, func() mo.Result[OperationResult] {
    return ExecutePotentiallyFlakOperation()
}).Map(func(result OperationResult) OperationResult {
    return ValidateOperationResult(result)
})
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