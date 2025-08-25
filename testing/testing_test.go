package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ========== TDD: RED-GREEN-REFACTOR TESTS ==========

// RED: Test that fails - Core functionality that needs implementation
func TestGenerateRandomId_ShouldUseTerratestPattern(t *testing.T) {
	utils := NewTestUtils()
	
	id := utils.GenerateRandomId()
	
	assert.NotEmpty(t, id, "Random ID should not be empty")
	assert.True(t, len(id) > 0, "Random ID should have length > 0")
}

// RED: Test TerratestAdapter delegation
func TestTerratestAdapter_ShouldDelegateToTestingT(t *testing.T) {
	mockT := NewMockTestingT("test-adapter")
	adapter := &TerratestAdapter{t: mockT}
	
	// Test Error method
	adapter.Error("test error")
	assert.True(t, mockT.WasErrorfCalled(), "Error should be called")
	
	// Test Errorf method  
	adapter.Errorf("test error: %s", "param")
	assert.Equal(t, 2, mockT.GetErrorfCallCount(), "Errorf should be called twice")
	
	// Test Fatal method
	adapter.Fatal("fatal error")
	assert.True(t, mockT.WasFailNowCalled(), "FailNow should be called")
	
	// Test Fatalf method
	adapter.Fatalf("fatal error: %s", "param")
	assert.Equal(t, 2, mockT.GetFailNowCallCount(), "FailNow should be called twice")
	
	// Test Name method
	name := adapter.Name()
	assert.Equal(t, "test-adapter", name, "Name should match")
	
	// Test Fail method
	adapter.Fail()
	assert.True(t, mockT.failCallCount > 0, "Fail should be called")
	
	// Test FailNow method
	adapter.FailNow()
	assert.Equal(t, 3, mockT.GetFailNowCallCount(), "FailNow should be called three times")
}

// ========== MOCK TESTING COVERAGE ==========

func TestMockTestingT_ShouldImplementTestingTInterface(t *testing.T) {
	mock := NewMockTestingT("test-mock")
	
	// Verify interface compliance
	var _ TestingT = mock
	
	assert.Equal(t, "test-mock", mock.Name(), "Name should match")
}

func TestMockTestingT_ShouldTrackAllMethodCalls(t *testing.T) {
	mock := NewMockTestingT("test")
	
	mock.Helper()
	mock.Errorf("error %d", 1)
	mock.Error("direct error")
	mock.Logf("log %s", "message")
	mock.Fail()
	mock.FailNow()
	
	assert.True(t, mock.WasHelperCalled(), "Helper should be tracked")
	assert.True(t, mock.WasErrorfCalled(), "Errorf should be tracked")
	assert.True(t, mock.WasLogfCalled(), "Logf should be tracked")
	assert.True(t, mock.WasFailNowCalled(), "FailNow should be tracked")
	
	assert.Equal(t, 1, mock.GetHelperCallCount(), "Helper count should be correct")
	assert.Equal(t, 2, mock.GetErrorfCallCount(), "Errorf count should be correct")
	assert.Equal(t, 1, mock.GetLogfCallCount(), "Logf count should be correct")
	assert.Equal(t, 1, mock.GetFailNowCallCount(), "FailNow count should be correct")
	assert.Equal(t, 1, mock.failCallCount, "Fail count should be correct")
}

func TestMockTestingT_ShouldCaptureMessages(t *testing.T) {
	mock := NewMockTestingT("test")
	
	mock.Errorf("error %d", 123)
	mock.Error("direct error")
	mock.Logf("log message")
	
	errorMsgs := mock.GetErrorMessages()
	assert.Len(t, errorMsgs, 2, "Should capture both error messages")
	assert.Contains(t, errorMsgs, "error 123", "Should contain formatted error")
	assert.Contains(t, errorMsgs, "direct error", "Should contain direct error")
	
	logMsgs := mock.GetLogMessages()
	assert.Len(t, logMsgs, 1, "Should capture log message")
	assert.Contains(t, logMsgs, "log message", "Should contain log message")
}

func TestMockTestingT_ShouldAllowReset(t *testing.T) {
	mock := NewMockTestingT("test")
	
	mock.Errorf("error")
	mock.Logf("log")
	mock.Helper()
	mock.FailNow()
	
	assert.True(t, mock.WasErrorfCalled(), "Should have errors before reset")
	
	mock.Reset()
	
	assert.False(t, mock.WasErrorfCalled(), "Should not have errors after reset")
	assert.False(t, mock.WasLogfCalled(), "Should not have logs after reset")
	assert.False(t, mock.WasHelperCalled(), "Should not have helper calls after reset")
	assert.False(t, mock.WasFailNowCalled(), "Should not have fail calls after reset")
	assert.Equal(t, 0, mock.failCallCount, "Fail count should be reset")
}

func TestMockTestingT_SetName_ShouldUpdateName(t *testing.T) {
	mock := NewMockTestingT("original")
	
	mock.SetName("updated")
	
	assert.Equal(t, "updated", mock.Name(), "Name should be updated")
}

func TestMockTestingT_ConcurrentAccess_ShouldBeSafe(t *testing.T) {
	mock := NewMockTestingT("concurrent-test")
	
	// Test concurrent access to verify mutex protection
	done := make(chan bool, 2)
	
	go func() {
		for i := 0; i < 50; i++ {
			mock.Errorf("concurrent error %d", i)
		}
		done <- true
	}()
	
	go func() {
		for i := 0; i < 50; i++ {
			mock.Logf("concurrent log %d", i)
		}
		done <- true
	}()
	
	// Wait for goroutines to complete
	<-done
	<-done
	
	// Should not panic and should have captured some messages
	assert.True(t, mock.GetErrorfCallCount() > 0, "Should capture concurrent errors")
	assert.True(t, mock.GetLogfCallCount() > 0, "Should capture concurrent logs")
}

// ========== UTILS TESTING COVERAGE ==========

func TestUtils_ShouldProvideAWSClientMocking(t *testing.T) {
	utils := NewTestUtils()
	
	lambdaClient := utils.CreateMockLambdaClient()
	assert.NotNil(t, lambdaClient, "Lambda client should not be nil")
	assert.Equal(t, 0, len(lambdaClient.InvokeCalls), "Invoke calls should be empty")
	
	ddbClient := utils.CreateMockDynamoDBClient()
	assert.NotNil(t, ddbClient, "DynamoDB client should not be nil")
	assert.Equal(t, 0, len(ddbClient.PutItemCalls), "Put item calls should be empty")
	assert.Equal(t, 0, len(ddbClient.GetItemCalls), "Get item calls should be empty")
}

func TestUtils_ShouldProvideContextCreationHelpers(t *testing.T) {
	utils := NewTestUtils()
	
	bgCtx := utils.CreateBackgroundContext()
	assert.NotNil(t, bgCtx, "Background context should not be nil")
	assert.Equal(t, context.Background(), bgCtx, "Should return background context")
	
	timeoutCtx := utils.CreateTimeoutContext(100 * time.Millisecond)
	assert.NotNil(t, timeoutCtx, "Timeout context should not be nil")
	
	// Test context timeout
	select {
	case <-timeoutCtx.Done():
		assert.Equal(t, context.DeadlineExceeded, timeoutCtx.Err(), "Context should timeout")
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Context should have timed out")
	}
	
	cancelCtx, cancel := utils.CreateCancelableContext()
	assert.NotNil(t, cancelCtx, "Cancelable context should not be nil")
	assert.NotNil(t, cancel, "Cancel function should not be nil")
	
	cancel()
	
	select {
	case <-cancelCtx.Done():
		assert.Equal(t, context.Canceled, cancelCtx.Err(), "Context should be canceled")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should have been canceled")
	}
}

func TestUtils_ShouldProvideTestDataGeneration(t *testing.T) {
	utils := NewTestUtils()
	
	// Test string generation
	str1 := utils.GenerateRandomString(10)
	str2 := utils.GenerateRandomString(10)
	assert.Equal(t, 10, len(str1), "String length should match")
	assert.Equal(t, 10, len(str2), "String length should match")
	assert.NotEqual(t, str1, str2, "Strings should be unique")
	
	// Test empty string generation
	emptyStr := utils.GenerateRandomString(0)
	assert.Empty(t, emptyStr, "Empty string should be returned for length 0")
	
	// Test number generation
	for i := 0; i < 10; i++ {
		num := utils.GenerateRandomNumber(5, 10)
		assert.True(t, num >= 5 && num <= 10, "Number should be in range [5,10]")
	}
	
	// Test UUID generation
	uuid1 := utils.GenerateUUID()
	uuid2 := utils.GenerateUUID()
	assert.NotEmpty(t, uuid1, "UUID should not be empty")
	assert.NotEmpty(t, uuid2, "UUID should not be empty")
	assert.NotEqual(t, uuid1, uuid2, "UUIDs should be unique")
	assert.Contains(t, uuid1, "-", "UUID should contain hyphens")
}

func TestUtils_ShouldProvideAssertionHelpers(t *testing.T) {
	utils := NewTestUtils()
	mockT := NewMockTestingT("test")
	
	// Test valid assertions (should not trigger errors)
	utils.AssertStringNotEmpty(mockT, "not empty", "should accept non-empty string")
	utils.AssertJSONValid(mockT, `{"valid": "json"}`, "should accept valid JSON")
	utils.AssertAWSArn(mockT, "arn:aws:lambda:us-east-1:123456789012:function:test", "should accept valid ARN")
	
	assert.False(t, mockT.WasErrorfCalled(), "Should not report errors for valid inputs")
	
	// Test invalid assertions (should trigger errors)
	mockT.Reset()
	utils.AssertStringNotEmpty(mockT, "", "should reject empty string")
	utils.AssertJSONValid(mockT, "invalid json", "should reject invalid JSON")
	utils.AssertAWSArn(mockT, "invalid-arn", "should reject invalid ARN")
	
	assert.True(t, mockT.WasErrorfCalled(), "Should report errors for invalid inputs")
	assert.Equal(t, 3, mockT.GetErrorfCallCount(), "Should report three errors")
}

func TestUtils_ShouldProvideEnvironmentVariableManagement(t *testing.T) {
	utils := NewTestUtils()
	em := utils.CreateEnvironmentManager()
	
	// Test setting and getting environment variables
	em.SetEnv("TEST_VAR", "test_value")
	value := em.GetEnv("TEST_VAR", "default")
	assert.Equal(t, "test_value", value, "Should return set value")
	
	// Test getting non-existent variable with default
	value = em.GetEnv("NON_EXISTENT", "default_value")
	assert.Equal(t, "default_value", value, "Should return default for missing var")
	
	// Test restoring environment
	em.RestoreAll()
	
	// Test setting same key multiple times (should store original only once)
	em.SetEnv("REPEAT_VAR", "value1")
	em.SetEnv("REPEAT_VAR", "value2")
	assert.Equal(t, 1, len(em.originalValues), "Should store original value only once")
	
	// Test restore with empty original value
	em.RestoreEnv("TEST_EMPTY", "")
}

func TestUtils_ShouldProvidePerformanceBenchmarking(t *testing.T) {
	utils := NewTestUtils()
	
	benchmark := utils.CreateBenchmark("test-benchmark")
	assert.Equal(t, "test-benchmark", benchmark.GetName(), "Should return correct name")
	
	time.Sleep(10 * time.Millisecond)
	duration := benchmark.Stop()
	
	assert.True(t, duration > 0, "Duration should be positive")
	assert.True(t, duration >= 10*time.Millisecond, "Duration should be at least sleep time")
}

func TestUtils_ShouldProvideRetryUtilities(t *testing.T) {
	utils := NewTestUtils()
	
	// Test successful retry
	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 2 {
			return fmt.Errorf("temporary failure")
		}
		return nil
	}
	
	err := utils.RetryOperation(operation, 3, 10*time.Millisecond)
	assert.NoError(t, err, "Should succeed on retry")
	assert.Equal(t, 2, attempts, "Should succeed on second attempt")
	
	// Test failure after max attempts
	attempts = 0
	failOperation := func() error {
		attempts++
		return fmt.Errorf("operation failed on attempt %d", attempts)
	}
	
	err = utils.RetryOperation(failOperation, 3, 5*time.Millisecond)
	assert.Error(t, err, "Should return error after max attempts")
	assert.Contains(t, err.Error(), "operation failed after 3 attempts", "Should include attempt count")
	assert.Equal(t, 3, attempts, "Should attempt exactly 3 times")
}

// ========== VALIDATION UTILITIES COVERAGE ==========

func TestUtils_ValidateEmail_ComprehensiveCases(t *testing.T) {
	utils := NewTestUtils()
	
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"test+tag@example.org",
		"123@numbers.net",
	}
	
	invalidEmails := []string{
		"",
		"not-an-email",
		"@domain.com",
		"user@",
		"user@domain",
		"user.domain.com",
	}
	
	for _, email := range validEmails {
		t.Run("valid_"+email, func(t *testing.T) {
			assert.True(t, utils.ValidateEmail(email), "Should validate valid email")
		})
	}
	
	for _, email := range invalidEmails {
		t.Run("invalid_"+email, func(t *testing.T) {
			assert.False(t, utils.ValidateEmail(email), "Should reject invalid email")
		})
	}
}

func TestUtils_ValidateAWSResourceName_ComprehensiveCases(t *testing.T) {
	utils := NewTestUtils()
	
	validNames := []string{
		"test-resource",
		"my-lambda-function-123",
		"table_name",
		"resource.with.dots",
		"UPPERCASE-resource",
		"resource123",
	}
	
	invalidNames := []string{
		"",                          // Empty string
		"resource with spaces",      // Spaces
		"resource@with@symbols",     // Invalid symbols
		strings.Repeat("a", 256),    // Too long
	}
	
	for _, name := range validNames {
		t.Run("valid_"+name, func(t *testing.T) {
			assert.True(t, utils.ValidateAWSResourceName(name), "Should validate valid name")
		})
	}
	
	for _, name := range invalidNames {
		t.Run("invalid_"+name, func(t *testing.T) {
			assert.False(t, utils.ValidateAWSResourceName(name), "Should reject invalid name")
		})
	}
}

// ========== STRING UTILITIES COVERAGE ==========

func TestUtils_SanitizeForAWS_ShouldHandleEdgeCases(t *testing.T) {
	utils := NewTestUtils()
	
	testCases := []struct {
		input    string
		expected string
	}{
		{"", "default"},
		{"---", "default"},
		{"TEST_/with.SYMBOLS", "test--with-symbols"},
		{"123-valid", "123-valid"},
		{"UPPERCASE", "uppercase"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := utils.SanitizeForAWS(tc.input)
			assert.Equal(t, tc.expected, result, "Should sanitize correctly")
		})
	}
}

func TestUtils_TruncateString_EdgeCases(t *testing.T) {
	utils := NewTestUtils()
	
	// Test normal truncation
	result := utils.TruncateString("hello world", 8)
	assert.Equal(t, "hello...", result, "Should truncate with ellipsis")
	
	// Test no truncation needed
	result = utils.TruncateString("short", 10)
	assert.Equal(t, "short", result, "Should return unchanged for short strings")
	
	// Test maxLength <= 3
	result = utils.TruncateString("test", 3)
	assert.Equal(t, "...", result, "Should return ellipsis for maxLength 3")
	
	result = utils.TruncateString("test", 2)
	assert.Equal(t, "...", result, "Should return ellipsis for maxLength 2")
	
	result = utils.TruncateString("test", 1)
	assert.Equal(t, "...", result, "Should return ellipsis for maxLength 1")
}

// ========== FIXTURES TESTING COVERAGE ==========

func TestFixtures_ShouldProvideCommonTestData(t *testing.T) {
	fixtures := NewTestFixtures()
	
	lambdaNames := fixtures.GetLambdaFunctionNames()
	assert.True(t, len(lambdaNames) > 0, "Should return Lambda function names")
	
	ddbNames := fixtures.GetDynamoDBTableNames()
	assert.True(t, len(ddbNames) > 0, "Should return DynamoDB table names")
	
	ruleNames := fixtures.GetEventBridgeRuleNames()
	assert.True(t, len(ruleNames) > 0, "Should return EventBridge rule names")
	
	sfnNames := fixtures.GetStepFunctionNames()
	assert.True(t, len(sfnNames) > 0, "Should return Step Function names")
}

func TestFixtures_ShouldProvideSampleAWSEvents(t *testing.T) {
	fixtures := NewTestFixtures()
	
	events := []string{
		fixtures.GetSampleS3Event(),
		fixtures.GetSampleSQSEvent(),
		fixtures.GetSampleAPIGatewayEvent(),
	}
	
	for i, event := range events {
		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			var result interface{}
			err := json.Unmarshal([]byte(event), &result)
			assert.NoError(t, err, "Event should be valid JSON")
			assert.NotEmpty(t, event, "Event should not be empty")
		})
	}
}

func TestFixtures_ShouldProvideTerraformOutputs(t *testing.T) {
	fixtures := NewTestFixtures()
	
	outputs := fixtures.GetSampleTerraformOutputs()
	assert.True(t, len(outputs) > 0, "Should return Terraform outputs")
	
	expectedKeys := []string{
		"lambda_function_arn",
		"lambda_function_name",
		"api_gateway_url",
		"dynamodb_table_name",
		"dynamodb_table_arn",
		"eventbridge_rule_name",
		"step_function_arn",
	}
	
	for _, key := range expectedKeys {
		assert.Contains(t, outputs, key, "Should contain expected key: %s", key)
	}
}

func TestFixtures_ShouldSupportCustomNamespace(t *testing.T) {
	fixtures := NewTestFixturesWithNamespace("custom-ns")
	
	lambdaNames := fixtures.GetLambdaFunctionNames()
	for _, name := range lambdaNames {
		assert.Contains(t, name, "custom-ns", "Names should contain custom namespace")
	}
}

func TestFixtures_GetShortId_ShouldHandleShortTimestamps(t *testing.T) {
	fixtures := &TestFixtures{
		namespace: "test",
		timestamp: "short",
	}
	
	shortId := fixtures.getShortId()
	assert.Equal(t, "short", shortId, "Short timestamp should be returned as-is")
	
	// Test with long timestamp
	fixtures.timestamp = "verylongtimestamp"
	longId := fixtures.getShortId()
	assert.Equal(t, "verylongti", longId, "Should return first 10 chars for long timestamp")
}

func TestFixtures_ShouldProvideEnvironmentVariablesAndContext(t *testing.T) {
	fixtures := NewTestFixtures()
	
	envVars := fixtures.GetSampleEnvironmentVariables()
	assert.True(t, len(envVars) > 0, "Should return environment variables")
	assert.Contains(t, envVars, "AWS_REGION", "Should contain AWS_REGION")
	assert.Contains(t, envVars, "ENVIRONMENT", "Should contain ENVIRONMENT")
	
	lambdaContext := fixtures.GetSampleLambdaContext()
	assert.True(t, len(lambdaContext) > 0, "Should return Lambda context")
	assert.Contains(t, lambdaContext, "functionName", "Should contain functionName")
	assert.Contains(t, lambdaContext, "awsRequestId", "Should contain awsRequestId")
}

// ========== COMPREHENSIVE COVERAGE TESTS ==========

// Test functions that might be uncovered to boost coverage to 90%+
func TestComprehensiveCoverage_AllUncoveredFunctions(t *testing.T) {
	t.Run("GetRandomRegion should delegate to Terratest", func(t *testing.T) {
		if os.Getenv("SKIP_INTEGRATION") != "" || os.Getenv("AWS_PROFILE") == "" {
			t.Skip("Skipping integration test - AWS credentials required")
		}
		
		utils := NewTestUtils()
		mockT := NewMockTestingT("test")
		
		// This will test the adapter functionality
		region := utils.GetRandomRegion(mockT)
		assert.NotEmpty(t, region, "Region should not be empty")
	})
	
	t.Run("GetAccountId should delegate to Terratest", func(t *testing.T) {
		if os.Getenv("SKIP_INTEGRATION") != "" || os.Getenv("AWS_PROFILE") == "" {
			t.Skip("Skipping integration test - AWS credentials required")
		}
		
		utils := NewTestUtils()
		mockT := NewMockTestingT("test")
		
		// This will test the adapter and error handling
		accountId := utils.GetAccountId(mockT, "us-east-1")
		assert.NotEmpty(t, accountId, "Account ID should not be empty")
	})
	
	t.Run("Environment Manager edge cases", func(t *testing.T) {
		utils := NewTestUtils()
		em := utils.CreateEnvironmentManager()
		
		// Test GetEnv with existing vs non-existing vars
		em.SetEnv("EXISTING", "value")
		existing := em.GetEnv("EXISTING", "default")
		assert.Equal(t, "value", existing, "Should return existing value")
		
		nonExisting := em.GetEnv("NON_EXISTING", "default")
		assert.Equal(t, "default", nonExisting, "Should return default for non-existing")
		
		// Test RestoreEnv with empty original value
		em.RestoreEnv("TEMP_VAR", "")
	})
	
	t.Run("Benchmark comprehensive test", func(t *testing.T) {
		utils := NewTestUtils()
		
		// Test various benchmark names
		benchmarks := []string{
			"simple-test",
			"complex_benchmark_name",
			"benchmark with spaces",
			"",
		}
		
		for _, name := range benchmarks {
			benchmark := utils.CreateBenchmark(name)
			assert.Equal(t, name, benchmark.GetName(), "Benchmark name should match")
			
			time.Sleep(1 * time.Millisecond)
			duration := benchmark.Stop()
			assert.True(t, duration > 0, "Duration should be positive")
		}
	})
}

// Final coverage boost for any remaining untested branches
func TestFinalCoverageBoost_RemainingBranches(t *testing.T) {
	t.Run("All TerratestAdapter methods", func(t *testing.T) {
		mockT := NewMockTestingT("adapter-test")
		adapter := &TerratestAdapter{t: mockT}
		
		// Test all adapter methods to ensure full coverage  
		adapter.Error("error")
		adapter.Errorf("error %s", "formatted")
		adapter.Fail()
		adapter.FailNow()
		adapter.Fatal("fatal")
		adapter.Fatalf("fatal %s", "formatted")
		name := adapter.Name()
		
		assert.Equal(t, "adapter-test", name, "Name should be correct")
		assert.True(t, mockT.WasErrorfCalled(), "Errorf should be called")
		assert.True(t, mockT.WasFailNowCalled(), "FailNow should be called")
	})
}