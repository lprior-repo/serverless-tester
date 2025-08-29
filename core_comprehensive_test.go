package vasdeference

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === TDD Test Suite for Core VasDeference Framework ===
// Following strict Red-Green-Refactor methodology for comprehensive coverage

// Test VasDeference Constructor Functions

func TestVasDeference_ConstructorValidation(t *testing.T) {
	t.Parallel()
	
	// Red: Test nil TestingT
	t.Run("FailsWithNilTestingT", func(t *testing.T) {
		assert.Panics(t, func() {
			New(nil)
		}, "Should panic with nil TestingT")
	})
	
	// Green: Test successful creation with default options
	t.Run("CreatesWithDefaultOptions", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		assert.NotNil(t, vdf)
		assert.NotNil(t, vdf.Context)
		assert.NotNil(t, vdf.Arrange)
		assert.Equal(t, RegionUSEast1, vdf.Context.Region)
		assert.False(t, mockT.errorCalled)
	})
	
	// Green: Test creation with custom options
	t.Run("CreatesWithCustomOptions", func(t *testing.T) {
		mockT := &mockTestingT{}
		opts := &VasDeferenceOptions{
			Region:     RegionUSWest2,
			Namespace:  "custom-test",
			MaxRetries: 5,
			RetryDelay: 2 * time.Second,
		}
		
		vdf := NewWithOptions(mockT, opts)
		
		assert.NotNil(t, vdf)
		assert.Equal(t, RegionUSWest2, vdf.Context.Region)
		assert.Equal(t, "custom-test", vdf.Arrange.Namespace)
		assert.Equal(t, 5, vdf.options.MaxRetries)
		assert.Equal(t, 2*time.Second, vdf.options.RetryDelay)
	})
}

func TestVasDeference_ConstructorE_ComprehensiveValidation(t *testing.T) {
	t.Parallel()
	
	// Red: Test nil TestingT
	t.Run("FailsWithNilTestingT", func(t *testing.T) {
		_, err := NewWithOptionsE(nil, &VasDeferenceOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TestingT cannot be nil")
	})
	
	// Green: Test nil options handling (should use defaults)
	t.Run("HandlesNilOptionsWithDefaults", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf, err := NewWithOptionsE(mockT, nil)
		
		require.NoError(t, err)
		assert.NotNil(t, vdf)
		assert.Equal(t, RegionUSEast1, vdf.Context.Region)
		assert.Equal(t, DefaultMaxRetries, vdf.options.MaxRetries)
		assert.Equal(t, DefaultRetryDelay, vdf.options.RetryDelay)
	})
	
	// Green: Test default value assignment
	t.Run("AssignsDefaultValues", func(t *testing.T) {
		mockT := &mockTestingT{}
		opts := &VasDeferenceOptions{} // Empty options
		
		vdf, err := NewWithOptionsE(mockT, opts)
		
		require.NoError(t, err)
		assert.Equal(t, RegionUSEast1, vdf.options.Region)
		assert.Equal(t, DefaultMaxRetries, vdf.options.MaxRetries)
		assert.Equal(t, DefaultRetryDelay, vdf.options.RetryDelay)
	})
	
	// Green: Test custom namespace handling
	t.Run("HandlesCustomNamespace", func(t *testing.T) {
		mockT := &mockTestingT{}
		opts := &VasDeferenceOptions{
			Namespace: "pr-123-custom_test/branch",
		}
		
		vdf, err := NewWithOptionsE(mockT, opts)
		
		require.NoError(t, err)
		// Namespace should be sanitized
		assert.Equal(t, "pr-123-custom-test-branch", vdf.Arrange.Namespace)
	})
}

// Test TestContext Creation and Management

func TestTestContext_CreationAndValidation(t *testing.T) {
	t.Parallel()
	
	// Red: Test nil TestingT
	t.Run("FailsWithNilTestingT", func(t *testing.T) {
		_, err := NewTestContextE(nil, &Options{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TestingT cannot be nil")
	})
	
	// Green: Test successful creation with defaults
	t.Run("CreatesWithDefaultOptions", func(t *testing.T) {
		mockT := &mockTestingT{}
		ctx, err := NewTestContextE(mockT, nil)
		
		require.NoError(t, err)
		assert.NotNil(t, ctx)
		assert.Equal(t, mockT, ctx.T)
		assert.Equal(t, RegionUSEast1, ctx.Region)
		assert.NotNil(t, ctx.AwsConfig)
	})
	
	// Green: Test custom region configuration
	t.Run("ConfiguresCustomRegion", func(t *testing.T) {
		mockT := &mockTestingT{}
		opts := &Options{
			Region:     RegionEUWest1,
			MaxRetries: 5,
			RetryDelay: 3 * time.Second,
		}
		
		ctx, err := NewTestContextE(mockT, opts)
		
		require.NoError(t, err)
		assert.Equal(t, RegionEUWest1, ctx.Region)
		assert.Equal(t, RegionEUWest1, ctx.AwsConfig.Region)
	})
	
	// Green: Test wrapper function
	t.Run("WrapperFunctionCallsE", func(t *testing.T) {
		mockT := &mockTestingT{}
		
		ctx := NewTestContext(mockT)
		
		assert.NotNil(t, ctx)
		assert.False(t, mockT.errorCalled) // Should succeed
	})
}

// Test Arrange Pattern and Cleanup Management

func TestArrange_CreationAndCleanup(t *testing.T) {
	t.Parallel()
	
	// Green: Test automatic namespace detection
	t.Run("CreatesWithAutomaticNamespace", func(t *testing.T) {
		mockT := &mockTestingT{}
		ctx := &TestContext{T: mockT, Region: RegionUSEast1}
		
		arrange := NewArrange(ctx)
		
		assert.NotNil(t, arrange)
		assert.NotEmpty(t, arrange.Namespace)
		assert.NotNil(t, arrange.cleanups)
	})
	
	// Green: Test custom namespace
	t.Run("CreatesWithCustomNamespace", func(t *testing.T) {
		mockT := &mockTestingT{}
		ctx := &TestContext{T: mockT, Region: RegionUSEast1}
		customNamespace := "test-custom-123"
		
		arrange := NewArrangeWithNamespace(ctx, customNamespace)
		
		assert.NotNil(t, arrange)
		assert.Equal(t, "test-custom-123", arrange.Namespace)
	})
	
	// Green: Test cleanup registration and execution
	t.Run("HandlesCleanupRegistrationAndExecution", func(t *testing.T) {
		mockT := &mockTestingT{}
		ctx := &TestContext{T: mockT, Region: RegionUSEast1}
		arrange := NewArrangeWithNamespace(ctx, "test")
		
		// Track cleanup execution order
		var executionOrder []int
		
		// Register cleanup functions
		arrange.RegisterCleanup(func() error {
			executionOrder = append(executionOrder, 1)
			return nil
		})
		arrange.RegisterCleanup(func() error {
			executionOrder = append(executionOrder, 2)
			return nil
		})
		arrange.RegisterCleanup(func() error {
			executionOrder = append(executionOrder, 3)
			return nil
		})
		
		// Execute cleanup
		arrange.Cleanup()
		
		// Should execute in LIFO order (3, 2, 1)
		assert.Equal(t, []int{3, 2, 1}, executionOrder)
		assert.False(t, mockT.errorCalled)
	})
	
	// Red: Test cleanup error handling
	t.Run("HandlesCleanupErrors", func(t *testing.T) {
		mockT := &mockTestingT{}
		ctx := &TestContext{T: mockT, Region: RegionUSEast1}
		arrange := NewArrangeWithNamespace(ctx, "test")
		
		// Register cleanup that will fail
		arrange.RegisterCleanup(func() error {
			return fmt.Errorf("cleanup failed")
		})
		
		// Execute cleanup
		arrange.Cleanup()
		
		// Should have logged error
		assert.True(t, mockT.errorCalled)
	})
}

// Test VasDeference Cleanup Integration

func TestVasDeference_CleanupIntegration(t *testing.T) {
	t.Parallel()
	
	// Green: Test cleanup registration through VasDeference
	t.Run("RegistersAndExecutesCleanupThroughVdf", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		var cleanupExecuted bool
		
		vdf.RegisterCleanup(func() error {
			cleanupExecuted = true
			return nil
		})
		
		vdf.Cleanup()
		
		assert.True(t, cleanupExecuted)
		assert.False(t, mockT.errorCalled)
	})
	
	// Red: Test cleanup error propagation
	t.Run("PropagatesCleanupErrors", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		vdf.RegisterCleanup(func() error {
			return fmt.Errorf("VasDeference cleanup failed")
		})
		
		vdf.Cleanup()
		
		assert.True(t, mockT.errorCalled)
	})
}

// Test Terraform Output Handling

func TestVasDeference_TerraformOutputHandling(t *testing.T) {
	t.Parallel()
	
	// Red: Test nil outputs map
	t.Run("FailsWithNilOutputsMap", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		_, err := vdf.GetTerraformOutputE(nil, "test-key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "outputs map cannot be nil")
	})
	
	// Red: Test missing key
	t.Run("FailsWithMissingKey", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		outputs := map[string]interface{}{
			"existing_key": "existing_value",
		}
		
		_, err := vdf.GetTerraformOutputE(outputs, "missing_key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key 'missing_key' not found")
	})
	
	// Red: Test non-string value
	t.Run("FailsWithNonStringValue", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		outputs := map[string]interface{}{
			"int_key": 12345,
		}
		
		_, err := vdf.GetTerraformOutputE(outputs, "int_key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a string")
	})
	
	// Green: Test successful value retrieval
	t.Run("RetrievesStringValueSuccessfully", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		outputs := map[string]interface{}{
			"string_key": "expected_value",
			"other_key":  "other_value",
		}
		
		value, err := vdf.GetTerraformOutputE(outputs, "string_key")
		require.NoError(t, err)
		assert.Equal(t, "expected_value", value)
	})
	
	// Green: Test wrapper function behavior
	t.Run("WrapperFunctionCallsE", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		outputs := map[string]interface{}{
			"test_key": "test_value",
		}
		
		value := vdf.GetTerraformOutput(outputs, "test_key")
		assert.Equal(t, "test_value", value)
		assert.False(t, mockT.errorCalled)
	})
	
	// Red: Test wrapper function failure
	t.Run("WrapperFunctionFailsAndCallsFailNow", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// This will cause the wrapper to fail
		vdf.GetTerraformOutput(nil, "test_key")
		
		assert.True(t, mockT.errorCalled)
		assert.True(t, mockT.failNowCalled)
	})
}

// Test AWS Credentials Validation

func TestVasDeference_AWSCredentialsValidation(t *testing.T) {
	t.Parallel()
	
	// Green: Test credentials validation success (with existing config)
	t.Run("ValidatesCredentialsWhenPresent", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// This may pass or fail depending on environment
		result := vdf.ValidateAWSCredentials()
		
		// Test the E version as well
		err := vdf.ValidateAWSCredentialsE()
		
		// Consistency check
		assert.Equal(t, err == nil, result)
	})
	
	// Test that we can call validation functions without panicking
	t.Run("CallsValidationFunctionsWithoutPanic", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// These should not panic regardless of credential state
		assert.NotPanics(t, func() {
			vdf.ValidateAWSCredentials()
		})
		
		assert.NotPanics(t, func() {
			vdf.ValidateAWSCredentialsE()
		})
	})
}

// Test AWS Client Creation

func TestVasDeference_AWSClientCreation(t *testing.T) {
	t.Parallel()
	
	// Green: Test Lambda client creation
	t.Run("CreatesLambdaClient", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		client := vdf.CreateLambdaClient()
		
		assert.NotNil(t, client)
		assert.IsType(t, &lambda.Client{}, client)
	})
	
	// Green: Test Lambda client with custom region
	t.Run("CreatesLambdaClientWithCustomRegion", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		client := vdf.CreateLambdaClientWithRegion(RegionUSWest2)
		
		assert.NotNil(t, client)
		assert.IsType(t, &lambda.Client{}, client)
	})
	
	// Green: Test other client creation (placeholder implementations)
	t.Run("CreatesOtherClients", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// These are placeholder implementations
		dynamoClient := vdf.CreateDynamoDBClient()
		assert.NotNil(t, dynamoClient)
		
		eventBridgeClient := vdf.CreateEventBridgeClient()
		assert.NotNil(t, eventBridgeClient)
		
		stepFunctionsClient := vdf.CreateStepFunctionsClient()
		assert.NotNil(t, stepFunctionsClient)
		
		logsClient := vdf.CreateCloudWatchLogsClient()
		assert.NotNil(t, logsClient)
	})
}

// Test Resource Naming and Prefixing

func TestVasDeference_ResourceNamingAndPrefixing(t *testing.T) {
	t.Parallel()
	
	// Green: Test resource name prefixing
	t.Run("PrefixesResourceNamesCorrectly", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := NewWithOptions(mockT, &VasDeferenceOptions{
			Namespace: "test-namespace",
		})
		
		prefixed := vdf.PrefixResourceName("my-resource")
		
		assert.Contains(t, prefixed, "test-namespace")
		assert.Contains(t, prefixed, "my-resource")
		assert.Contains(t, prefixed, "-") // Should be joined with hyphen
	})
	
	// Green: Test namespace sanitization in prefixing
	t.Run("SanitizesNamespaceInPrefixing", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := NewWithOptions(mockT, &VasDeferenceOptions{
			Namespace: "test_namespace/with.special@chars",
		})
		
		prefixed := vdf.PrefixResourceName("my_resource.test")
		
		// Should sanitize both namespace and resource name
		assert.NotContains(t, prefixed, "_")
		assert.NotContains(t, prefixed, "/")
		assert.NotContains(t, prefixed, ".")
		assert.NotContains(t, prefixed, "@")
		assert.Contains(t, prefixed, "-") // Should use hyphens instead
	})
}

// Test Environment Variable Handling

func TestVasDeference_EnvironmentVariables(t *testing.T) {
	t.Parallel()
	
	// Green: Test environment variable retrieval
	t.Run("RetrievesEnvironmentVariables", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// Set a test environment variable
		testKey := "VDF_TEST_VAR"
		testValue := "test_value"
		os.Setenv(testKey, testValue)
		defer os.Unsetenv(testKey)
		
		value := vdf.GetEnvVar(testKey)
		assert.Equal(t, testValue, value)
	})
	
	// Green: Test environment variable with default
	t.Run("RetrievesEnvironmentVariableWithDefault", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// Test with existing variable
		testKey := "VDF_TEST_WITH_DEFAULT"
		testValue := "actual_value"
		os.Setenv(testKey, testValue)
		defer os.Unsetenv(testKey)
		
		value := vdf.GetEnvVarWithDefault(testKey, "default_value")
		assert.Equal(t, testValue, value)
		
		// Test with non-existent variable
		value = vdf.GetEnvVarWithDefault("VDF_NONEXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", value)
	})
}

// Test CI Environment Detection

func TestVasDeference_CIEnvironmentDetection(t *testing.T) {
	t.Parallel()
	
	// Green: Test CI environment detection with various CI variables
	t.Run("DetectsCIEnvironmentVariables", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		ciVars := []string{
			"CI",
			"CONTINUOUS_INTEGRATION",
			"GITHUB_ACTIONS",
			"CIRCLE_CI",
			"JENKINS_URL",
			"TRAVIS",
			"BUILDKITE",
		}
		
		// Clear all CI variables first
		for _, envVar := range ciVars {
			os.Unsetenv(envVar)
		}
		
		// Should not detect CI when no variables are set
		assert.False(t, vdf.IsCIEnvironment())
		
		// Test each CI variable
		for _, envVar := range ciVars {
			os.Setenv(envVar, "true")
			assert.True(t, vdf.IsCIEnvironment(), "Should detect CI with %s set", envVar)
			os.Unsetenv(envVar)
		}
	})
}

// Test Context Creation

func TestVasDeference_ContextCreation(t *testing.T) {
	t.Parallel()
	
	// Green: Test context creation with timeout
	t.Run("CreatesContextWithTimeout", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		timeout := 30 * time.Second
		ctx := vdf.CreateContextWithTimeout(timeout)
		
		assert.NotNil(t, ctx)
		
		// Verify context has deadline
		deadline, ok := ctx.Deadline()
		assert.True(t, ok, "Context should have a deadline")
		assert.True(t, deadline.After(time.Now()), "Deadline should be in the future")
		assert.True(t, deadline.Before(time.Now().Add(35*time.Second)), "Deadline should be within expected range")
	})
}

// Test Region Management

func TestVasDeference_RegionManagement(t *testing.T) {
	t.Parallel()
	
	// Green: Test default region retrieval
	t.Run("ReturnsDefaultRegion", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		region := vdf.GetDefaultRegion()
		assert.Equal(t, RegionUSEast1, region) // Default is us-east-1
	})
	
	// Green: Test custom default region
	t.Run("ReturnsCustomDefaultRegion", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := NewWithOptions(mockT, &VasDeferenceOptions{
			Region: RegionEUWest1,
		})
		
		region := vdf.GetDefaultRegion()
		assert.Equal(t, RegionEUWest1, region)
	})
	
	// Green: Test random region selection
	t.Run("ReturnsRandomRegion", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		region := vdf.GetRandomRegion()
		
		// Should return one of the predefined regions
		validRegions := []string{RegionUSEast1, RegionUSWest2, RegionEUWest1}
		assert.Contains(t, validRegions, region)
	})
}

// Test Account ID Retrieval

func TestVasDeference_AccountIDRetrieval(t *testing.T) {
	t.Parallel()
	
	// Green: Test account ID retrieval (placeholder implementation)
	t.Run("RetrievesAccountID", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		accountId := vdf.GetAccountId()
		assert.NotEmpty(t, accountId)
		assert.Len(t, accountId, 12) // AWS account IDs are 12 digits
	})
	
	// Green: Test account ID retrieval with error handling
	t.Run("RetrievesAccountIDWithErrorHandling", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		accountId, err := vdf.GetAccountIdE()
		assert.NoError(t, err) // Placeholder implementation always succeeds
		assert.NotEmpty(t, accountId)
	})
}

// Test Random Name Generation

func TestVasDeference_RandomNameGeneration(t *testing.T) {
	t.Parallel()
	
	// Green: Test random name generation
	t.Run("GeneratesRandomNames", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		prefix := "test-resource"
		name := vdf.GenerateRandomName(prefix)
		
		assert.Contains(t, name, prefix)
		assert.Greater(t, len(name), len(prefix)) // Should have random suffix
		
		// Generate another name to verify uniqueness
		name2 := vdf.GenerateRandomName(prefix)
		assert.NotEqual(t, name, name2, "Generated names should be unique")
	})
}

// Test Retry with Backoff

func TestVasDeference_RetryWithBackoff(t *testing.T) {
	t.Parallel()
	
	// Green: Test successful retry
	t.Run("RetriesSuccessfully", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		attemptCount := 0
		result := vdf.RetryWithBackoff(
			"test operation",
			func() (string, error) {
				attemptCount++
				if attemptCount < 3 {
					return "", fmt.Errorf("attempt %d failed", attemptCount)
				}
				return "success", nil
			},
			5,                     // maxRetries
			10*time.Millisecond,  // sleepBetweenRetries
		)
		
		assert.Equal(t, "success", result)
		assert.Equal(t, 3, attemptCount)
		assert.False(t, mockT.errorCalled)
	})
	
	// Red: Test retry exhaustion
	t.Run("ExhaustsRetriesAndFails", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		attemptCount := 0
		_, err := vdf.RetryWithBackoffE(
			"failing operation",
			func() (string, error) {
				attemptCount++
				return "", fmt.Errorf("attempt %d failed", attemptCount)
			},
			3,                   // maxRetries
			1*time.Millisecond, // sleepBetweenRetries
		)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "after 3 attempts")
		assert.Equal(t, 3, attemptCount) // Should have attempted exactly maxRetries times
	})
	
	// Green: Test first attempt success
	t.Run("SucceedsOnFirstAttempt", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		attemptCount := 0
		result, err := vdf.RetryWithBackoffE(
			"immediate success",
			func() (string, error) {
				attemptCount++
				return "immediate success", nil
			},
			5,
			10*time.Millisecond,
		)
		
		assert.NoError(t, err)
		assert.Equal(t, "immediate success", result)
		assert.Equal(t, 1, attemptCount) // Should only attempt once
	})
}

// Test Namespace Sanitization

func TestSanitizeNamespace_ComprehensiveScenarios(t *testing.T) {
	t.Parallel()
	
	// Green: Test basic sanitization
	t.Run("SanitizesBasicCharacters", func(t *testing.T) {
		input := "test_namespace.with/special@chars"
		expected := "test-namespace-with-special-chars"
		
		result := SanitizeNamespace(input)
		assert.Equal(t, expected, result)
	})
	
	// Green: Test uppercase conversion
	t.Run("ConvertsToLowercase", func(t *testing.T) {
		input := "Test-Namespace-With-UPPERCASE"
		expected := "test-namespace-with-uppercase"
		
		result := SanitizeNamespace(input)
		assert.Equal(t, expected, result)
	})
	
	// Green: Test special character removal
	t.Run("RemovesInvalidCharacters", func(t *testing.T) {
		input := "test@#$%^&*()namespace"
		expected := "test-namespace"
		
		result := SanitizeNamespace(input)
		assert.Equal(t, expected, result)
	})
	
	// Green: Test hyphen trimming
	t.Run("TrimsLeadingAndTrailingHyphens", func(t *testing.T) {
		input := "---test-namespace---"
		expected := "test-namespace"
		
		result := SanitizeNamespace(input)
		assert.Equal(t, expected, result)
	})
	
	// Green: Test empty string handling
	t.Run("HandlesEmptyString", func(t *testing.T) {
		input := ""
		expected := "default"
		
		result := SanitizeNamespace(input)
		assert.Equal(t, expected, result)
	})
	
	// Green: Test invalid characters only
	t.Run("HandlesInvalidCharactersOnly", func(t *testing.T) {
		input := "@#$%^&*()"
		expected := "default"
		
		result := SanitizeNamespace(input)
		assert.Equal(t, expected, result)
	})
	
	// Green: Test valid alphanumeric with hyphens
	t.Run("PreservesValidCharacters", func(t *testing.T) {
		input := "test-namespace-123"
		expected := "test-namespace-123"
		
		result := SanitizeNamespace(input)
		assert.Equal(t, expected, result)
	})
}

// Test GitHub Integration Functions

func TestGitHubIntegration_ComprehensiveScenarios(t *testing.T) {
	t.Parallel()
	
	// These tests are more challenging to unit test since they depend on external commands
	// We'll test the logic paths we can control
	
	// Green: Test unique ID generation
	t.Run("GeneratesUniqueIDs", func(t *testing.T) {
		id1 := UniqueId()
		id2 := UniqueId()
		
		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2, "Unique IDs should be different")
	})
}

// Test MockTestingT Implementation

func TestMockTestingT_Implementation(t *testing.T) {
	t.Parallel()
	
	// Green: Test mock functionality
	t.Run("TracksCalls", func(t *testing.T) {
		mock := &mockTestingT{}
		
		// Test initial state
		assert.False(t, mock.errorCalled)
		assert.False(t, mock.failNowCalled)
		
		// Test Errorf
		mock.Errorf("test error: %s", "message")
		assert.True(t, mock.errorCalled)
		
		// Test Error
		mock2 := &mockTestingT{}
		mock2.Error("test error")
		assert.True(t, mock2.errorCalled)
		
		// Test FailNow
		mock3 := &mockTestingT{}
		mock3.FailNow()
		assert.True(t, mock3.failNowCalled)
		
		// Test Fatal
		mock4 := &mockTestingT{}
		mock4.Fatal("fatal error")
		assert.True(t, mock4.failNowCalled)
		
		// Test Fatalf
		mock5 := &mockTestingT{}
		mock5.Fatalf("fatal error: %s", "message")
		assert.True(t, mock5.failNowCalled)
		
		// Test Name
		name := mock.Name()
		assert.Equal(t, "SimpleTest", name)
		
		// Test methods that don't change state
		assert.NotPanics(t, func() {
			mock.Helper()
			mock.Fail()
			mock.Logf("test log: %s", "message")
		})
	})
}

// Test Integration with Real Testing Framework

func TestIntegrationWithRealTestingFramework(t *testing.T) {
	t.Parallel()
	
	// Green: Test that VasDeference works with real testing.T
	t.Run("WorksWithRealTestingT", func(t *testing.T) {
		// This test uses the real testing.T to ensure compatibility
		vdf := New(t)
		
		assert.NotNil(t, vdf)
		assert.Equal(t, t, vdf.Context.T)
		
		// Test basic operations
		region := vdf.GetDefaultRegion()
		assert.NotEmpty(t, region)
		
		client := vdf.CreateLambdaClient()
		assert.NotNil(t, client)
		
		randomName := vdf.GenerateRandomName("test")
		assert.Contains(t, randomName, "test")
		
		// Test cleanup registration (but don't execute to avoid side effects)
		vdf.RegisterCleanup(func() error {
			return nil
		})
		
		// Verify cleanup was registered (we can't easily test execution without side effects)
		assert.NotNil(t, vdf.Arrange.cleanups)
		assert.Len(t, vdf.Arrange.cleanups, 1)
	})
}

// Benchmark tests for performance validation

func BenchmarkVasDefererence_Creation(b *testing.B) {
	mockT := &mockTestingT{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vdf := New(mockT)
		_ = vdf
	}
}

func BenchmarkSanitizeNamespace(b *testing.B) {
	input := "Complex_Namespace.With/Special@Characters#And$Symbols%"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := SanitizeNamespace(input)
		_ = result
	}
}

func BenchmarkUniqueIdGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := UniqueId()
		_ = id
	}
}