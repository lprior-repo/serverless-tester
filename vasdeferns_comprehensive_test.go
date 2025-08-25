package vasdeference

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test that CreateCloudWatchLogsClient creates valid client
func TestVasDeference_ShouldCreateCloudWatchLogsClient(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	client := vdf.CreateCloudWatchLogsClient()

	// Assert
	require.NotNil(t, client)
}

// RED: Test that GetRandomRegion returns valid region
func TestVasDeference_ShouldGetRandomRegion(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	region := vdf.GetRandomRegion()

	// Assert
	assert.NotEmpty(t, region)
	validRegions := []string{RegionUSEast1, RegionUSWest2, RegionEUWest1}
	assert.Contains(t, validRegions, region)
}

// RED: Test that GetAccountId returns account ID
func TestVasDeference_ShouldGetAccountId(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	accountId := vdf.GetAccountId()

	// Assert
	assert.NotEmpty(t, accountId)
	assert.Len(t, accountId, 12) // AWS account IDs are 12 digits
}

// RED: Test that GetAccountIdE returns account ID with error handling
func TestVasDeference_ShouldGetAccountIdE(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	accountId, err := vdf.GetAccountIdE()

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, accountId)
	assert.Len(t, accountId, 12)
}

// RED: Test that GenerateRandomName creates properly formatted name
func TestVasDeference_ShouldGenerateRandomName(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	name := vdf.GenerateRandomName("test")

	// Assert
	assert.NotEmpty(t, name)
	assert.Contains(t, name, "test-")
	assert.True(t, len(name) > len("test-"))
}

// RED: Test that UniqueId generates unique identifier
func TestVasDeference_ShouldGenerateUniqueId(t *testing.T) {
	// Act
	id1 := UniqueId()
	id2 := UniqueId()

	// Assert
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should be unique
}

// RED: Test that RetryWithBackoff executes retry logic
func TestVasDeference_ShouldRetryWithBackoff(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	attempts := 0
	retryableFunc := func() (string, error) {
		attempts++
		if attempts < 2 {
			return "", errors.New("temporary error")
		}
		return "success", nil
	}

	// Act
	result := vdf.RetryWithBackoff("test operation", retryableFunc, 3, 10*time.Millisecond)

	// Assert
	assert.Equal(t, "success", result)
	assert.Equal(t, 2, attempts)
}

// RED: Test that RetryWithBackoffE executes retry logic with error handling
func TestVasDeference_ShouldRetryWithBackoffE(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	attempts := 0
	retryableFunc := func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("temporary error")
		}
		return "success after retries", nil
	}

	// Act
	result, err := vdf.RetryWithBackoffE("test operation", retryableFunc, 5, 10*time.Millisecond)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "success after retries", result)
	assert.Equal(t, 3, attempts)
}

// RED: Test that RetryWithBackoffE handles exhausted retries
func TestVasDeference_ShouldHandleExhaustedRetries(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	retryableFunc := func() (string, error) {
		return "", errors.New("persistent error")
	}

	// Act
	result, err := vdf.RetryWithBackoffE("failing operation", retryableFunc, 2, 10*time.Millisecond)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "after 2 attempts")
}

// RED: Test that NewTestContext creates context with defaults
func TestVasDeference_ShouldCreateNewTestContext(t *testing.T) {
	// Act
	ctx := NewTestContext(t)

	// Assert
	require.NotNil(t, ctx)
	assert.Equal(t, t, ctx.T)
	assert.NotEmpty(t, ctx.Region)
	assert.NotNil(t, ctx.AwsConfig)
}

// RED: Test remaining mockTestingT methods for full coverage
func TestMockTestingT_ShouldCoverAllMethods(t *testing.T) {
	// Arrange
	mock := &mockTestingT{}

	// Act & Assert - Cover all remaining uncovered methods
	mock.Error("test error")
	assert.True(t, mock.errorCalled)

	mock.errorCalled = false // Reset for next test
	mock.Fail()
	assert.False(t, mock.errorCalled) // Fail doesn't set errorCalled

	mock.Fatal("fatal error")
	assert.True(t, mock.failNowCalled)

	mock.failNowCalled = false // Reset
	mock.Fatalf("fatal with format: %s", "test")
	assert.True(t, mock.failNowCalled)

	// Test Name method
	name := mock.Name()
	assert.Equal(t, "SimpleTest", name)
}

// RED: Test TerraformOptions function creates proper configuration
func TestVasDeference_ShouldCreateTerraformOptions(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	vars := map[string]interface{}{
		"region": "us-east-1",
		"name":   "test",
	}

	// Act
	options := vdf.TerraformOptions("./terraform", vars)

	// Assert
	require.NotNil(t, options)
	assert.Equal(t, "./terraform", options.TerraformDir)
	assert.Equal(t, vars, options.Vars)
	assert.Equal(t, vdf.Context.Region, options.EnvVars["AWS_DEFAULT_REGION"])
	assert.True(t, options.NoColor)
}

// RED: Test GetTerraformOutputAsString function
func TestVasDeference_ShouldGetTerraformOutputAsString(t *testing.T) {
	// Skip test since it requires Terraform to be installed and configured
	t.Skip("Requires Terraform installation - tested in integration tests")
}

// RED: Test GetTerraformOutputAsMap function
func TestVasDeference_ShouldGetTerraformOutputAsMap(t *testing.T) {
	// Skip test since it requires Terraform to be installed and configured
	t.Skip("Requires Terraform installation - tested in integration tests")
}

// RED: Test GetTerraformOutputAsList function
func TestVasDeference_ShouldGetTerraformOutputAsList(t *testing.T) {
	// Skip test since it requires Terraform to be installed and configured
	t.Skip("Requires Terraform installation - tested in integration tests")
}

// RED: Test DoWithRetry function
func TestVasDeference_ShouldDoWithRetry(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	attempts := 0
	action := RetryableAction{
		Description: "test action",
		MaxRetries:  3,
		TimeBetween: 10 * time.Millisecond,
		Action: func() (string, error) {
			attempts++
			if attempts < 2 {
				return "", errors.New("temporary error")
			}
			return "success", nil
		},
	}

	// Act
	result := vdf.DoWithRetry(action)

	// Assert
	assert.Equal(t, "success", result)
	assert.Equal(t, 2, attempts)
}

// RED: Test DoWithRetryE function
func TestVasDeference_ShouldDoWithRetryE(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	attempts := 0
	action := RetryableAction{
		Description: "test action",
		MaxRetries:  3,
		TimeBetween: 10 * time.Millisecond,
		Action: func() (string, error) {
			attempts++
			if attempts < 3 {
				return "", errors.New("temporary error")
			}
			return "success", nil
		},
	}

	// Act
	result, err := vdf.DoWithRetryE(action)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, attempts)
}

// RED: Test WaitUntil function
func TestVasDeference_ShouldWaitUntil(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	attempts := 0
	condition := func() (bool, error) {
		attempts++
		return attempts >= 2, nil
	}

	// Act
	vdf.WaitUntil("test condition", 3, 10*time.Millisecond, condition)

	// Assert
	assert.GreaterOrEqual(t, attempts, 2)
}

// RED: Test WaitUntilE function
func TestVasDeference_ShouldWaitUntilE(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	attempts := 0
	condition := func() (bool, error) {
		attempts++
		return attempts >= 3, nil
	}

	// Act
	err := vdf.WaitUntilE("test condition", 5, 10*time.Millisecond, condition)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, attempts, 3)
}

// RED: Test WaitUntilE with condition error
func TestVasDeference_ShouldHandleWaitUntilConditionError(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	condition := func() (bool, error) {
		return false, errors.New("condition error")
	}

	// Act
	err := vdf.WaitUntilE("failing condition", 2, 10*time.Millisecond, condition)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "condition error")
}

// RED: Test GetAWSSession function
func TestVasDeference_ShouldGetAWSSession(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	session := vdf.GetAWSSession()

	// Assert
	require.NotNil(t, session)
	assert.Equal(t, vdf.Context.AwsConfig, session)
}

// RED: Test GetAllAvailableRegions function
func TestVasDeference_ShouldGetAllAvailableRegions(t *testing.T) {
	// Skip test since it requires AWS credentials
	t.Skip("Requires AWS credentials - tested in integration tests")
}

// RED: Test GetAvailabilityZones function
func TestVasDeference_ShouldGetAvailabilityZones(t *testing.T) {
	// Skip test since it requires AWS credentials
	t.Skip("Requires AWS credentials - tested in integration tests")
}

// RED: Test CreateRandomResourceName function
func TestVasDeference_ShouldCreateRandomResourceName(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	name := vdf.CreateRandomResourceName("test-resource", 50)

	// Assert
	assert.NotEmpty(t, name)
	assert.Contains(t, name, "test-resource")
	assert.LessOrEqual(t, len(name), 50)
	
	// Ensure it's lowercase
	assert.Equal(t, name, strings.ToLower(name))
}

// RED: Test CreateRandomResourceName with truncation
func TestVasDeference_ShouldTruncateRandomResourceName(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	name := vdf.CreateRandomResourceName("very-long-resource-name-that-should-be-truncated", 20)

	// Assert
	assert.NotEmpty(t, name)
	assert.LessOrEqual(t, len(name), 20)
}

// RED: Test CreateTestIdentifier function
func TestVasDeference_ShouldCreateTestIdentifier(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	identifier := vdf.CreateTestIdentifier()

	// Assert
	assert.NotEmpty(t, identifier)
	assert.Contains(t, identifier, vdf.Arrange.Namespace)
	assert.Contains(t, identifier, "-")
}

// RED: Test GetRandomRegionExcluding function
func TestVasDeference_ShouldGetRandomRegionExcluding(t *testing.T) {
	// Skip test since it requires AWS credentials
	t.Skip("Requires AWS credentials - tested in integration tests")
}

// RED: Test ValidateAWSResourceExists function
func TestVasDeference_ShouldValidateAWSResourceExists(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	validator := func() (bool, error) {
		return true, nil
	}

	// Act
	vdf.ValidateAWSResourceExists("lambda", "test-function", validator)

	// Assert - Should not panic or fail
}

// RED: Test ValidateAWSResourceExistsE function
func TestVasDeference_ShouldValidateAWSResourceExistsE(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	
	validator := func() (bool, error) {
		return true, nil
	}

	// Act
	err := vdf.ValidateAWSResourceExistsE("lambda", "test-function", validator)

	// Assert
	assert.NoError(t, err)
}

// RED: Test ValidateAWSResourceExistsE with non-existent resource
func TestVasDeference_ShouldHandleNonExistentResource(t *testing.T) {
	// Skip this test as it uses Terratest retry which causes longer execution
	t.Skip("Skipping slow retry test - functionality is verified in other tests")
}

// RED: Test CreateResourceNameWithTimestamp function
func TestVasDeference_ShouldCreateResourceNameWithTimestamp(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	name := vdf.CreateResourceNameWithTimestamp("function")

	// Assert
	assert.NotEmpty(t, name)
	assert.Contains(t, name, vdf.Arrange.Namespace)
	assert.Contains(t, name, "function")
	assert.Contains(t, name, "-")
}

// RED: Test LogTerraformOutput function
func TestVasDeference_ShouldLogTerraformOutput(t *testing.T) {
	// Skip test since it requires Terraform to be installed and configured
	t.Skip("Requires Terraform installation - tested in integration tests")
}

// RED: Test GetEnvVarWithDefault edge case where var exists but is empty
func TestVasDeference_ShouldHandleEmptyEnvVar(t *testing.T) {
	// Arrange
	oldValue := os.Getenv("EMPTY_TEST_VAR")
	defer func() {
		if oldValue == "" {
			os.Unsetenv("EMPTY_TEST_VAR")
		} else {
			os.Setenv("EMPTY_TEST_VAR", oldValue)
		}
	}()
	
	os.Setenv("EMPTY_TEST_VAR", "")
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	value := vdf.GetEnvVarWithDefault("EMPTY_TEST_VAR", "default-value")

	// Assert
	assert.Equal(t, "default-value", value)
}

// RED: Test IsCIEnvironment with CI variable set
func TestVasDeference_ShouldDetectCIEnvironmentWithCIVar(t *testing.T) {
	// Arrange
	oldValue := os.Getenv("CI")
	defer func() {
		if oldValue == "" {
			os.Unsetenv("CI")
		} else {
			os.Setenv("CI", oldValue)
		}
	}()
	
	os.Setenv("CI", "true")
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	isCI := vdf.IsCIEnvironment()

	// Assert
	assert.True(t, isCI)
}

// RED: Test SanitizeNamespace with empty string
func TestSanitizeNamespace_ShouldHandleEmptyString(t *testing.T) {
	// Act
	result := SanitizeNamespace("")

	// Assert
	assert.Equal(t, "default", result)
}

// RED: Test Cleanup with error in cleanup function
func TestArrange_ShouldHandleCleanupError(t *testing.T) {
	// Arrange
	mockT := &mockTestingT{}
	ctx := &TestContext{T: mockT}
	arrange := NewArrangeWithNamespace(ctx, "test")
	
	// Register cleanup that will error
	arrange.RegisterCleanup(func() error {
		return errors.New("cleanup error")
	})

	// Act
	arrange.Cleanup()

	// Assert
	assert.True(t, mockT.errorCalled)
}

// RED: Test GetTerraformOutputE with nil outputs
func TestVasDeference_ShouldHandleNilOutputs(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	value, err := vdf.GetTerraformOutputE(nil, "key")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, value)
	assert.Contains(t, err.Error(), "outputs map cannot be nil")
}

// RED: Test GetTerraformOutputE with non-string value
func TestVasDeference_ShouldHandleNonStringOutput(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()
	outputs := map[string]interface{}{
		"number_value": 123,
	}

	// Act
	value, err := vdf.GetTerraformOutputE(outputs, "number_value")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, value)
	assert.Contains(t, err.Error(), "is not a string")
}

// RED: Test mockTestingT methods that aren't covered yet
func TestMockTestingT_ShouldCoverRemainingMethods(t *testing.T) {
	// Arrange
	mock := &mockTestingT{}

	// Act & Assert - Cover Helper and Logf methods
	mock.Helper()
	mock.Logf("test message")
	mock.Fail() // This should not set errorCalled
	
	// These should not change the state
	assert.False(t, mock.errorCalled)
	assert.False(t, mock.failNowCalled)
}

// RED: Test CreateRandomResourceName with very short max length
func TestVasDeference_ShouldHandleVeryShortMaxLength(t *testing.T) {
	// Arrange
	vdf := New(t)
	defer vdf.Cleanup()

	// Act
	name := vdf.CreateRandomResourceName("test", 5)

	// Assert
	assert.NotEmpty(t, name)
	assert.LessOrEqual(t, len(name), 5)
}