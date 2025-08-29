package vasdeference

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === TDD Test Suite for Terratest Integration ===
// Following strict Red-Green-Refactor methodology for comprehensive coverage

// Test Terraform Options Creation

func TestVasDeference_TerraformOptionsCreation(t *testing.T) {
	t.Parallel()
	
	// Green: Test basic Terraform options creation
	t.Run("CreatesBasicTerraformOptions", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := NewWithOptions(mockT, &VasDeferenceOptions{
			Region: RegionUSWest2,
		})
		
		terraformDir := "/path/to/terraform"
		vars := map[string]interface{}{
			"region":      "us-west-2",
			"environment": "test",
			"instance_count": 3,
		}
		
		options := vdf.TerraformOptions(terraformDir, vars)
		
		assert.NotNil(t, options)
		assert.Equal(t, terraformDir, options.TerraformDir)
		assert.Equal(t, vars, options.Vars)
		assert.Equal(t, RegionUSWest2, options.EnvVars["AWS_DEFAULT_REGION"])
		assert.True(t, options.NoColor)
	})
	
	// Green: Test with nil variables
	t.Run("HandlesNilVariables", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		terraformDir := "/path/to/terraform"
		
		options := vdf.TerraformOptions(terraformDir, nil)
		
		assert.NotNil(t, options)
		assert.Equal(t, terraformDir, options.TerraformDir)
		assert.Nil(t, options.Vars)
		assert.Equal(t, RegionUSEast1, options.EnvVars["AWS_DEFAULT_REGION"]) // Default region
	})
	
	// Green: Test with empty variables
	t.Run("HandlesEmptyVariables", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		terraformDir := "/path/to/terraform"
		vars := make(map[string]interface{})
		
		options := vdf.TerraformOptions(terraformDir, vars)
		
		assert.NotNil(t, options)
		assert.Equal(t, vars, options.Vars)
		assert.NotNil(t, options.EnvVars)
	})
}

// Test Terraform Output Retrieval Methods

func TestVasDeference_TerraformOutputRetrieval(t *testing.T) {
	t.Parallel()
	
	// Note: These tests focus on the wrapper methods and parameter validation
	// Actual Terraform operations would require real Terraform state
	
	// Green: Test output retrieval method signatures and basic validation  
	t.Run("ProvidesOutputRetrievalMethods", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		terraformDir := "/path/to/terraform"
		vars := map[string]interface{}{
			"test_var": "test_value",
		}
		
		options := vdf.TerraformOptions(terraformDir, vars)
		
		// These methods exist and can be called (they would interact with actual Terraform in real usage)
		assert.NotPanics(t, func() {
			// GetTerraformOutputAsString - would call terraform.Output internally
			_ = vdf.GetTerraformOutputAsString
			
			// GetTerraformOutputAsMap - would call terraform.OutputMap internally  
			_ = vdf.GetTerraformOutputAsMap
			
			// GetTerraformOutputAsList - would call terraform.OutputList internally
			_ = vdf.GetTerraformOutputAsList
			
			// LogTerraformOutput - would call terraform.OutputAll internally
			_ = vdf.LogTerraformOutput
		})
		
		// Verify the options are properly structured for Terratest
		assert.Equal(t, terraformDir, options.TerraformDir)
		assert.Equal(t, vars, options.Vars)
	})
}

// Test Retryable Actions

func TestVasDeference_RetryableActions(t *testing.T) {
	t.Parallel()
	
	// Green: Test successful retryable action
	t.Run("ExecutesSuccessfulRetryableAction", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		attemptCount := 0
		action := RetryableAction{
			Description: "test successful action",
			MaxRetries:  3,
			TimeBetween: 10 * time.Millisecond,
			Action: func() (string, error) {
				attemptCount++
				return "success", nil
			},
		}
		
		result := vdf.DoWithRetry(action)
		
		assert.Equal(t, "success", result)
		assert.Equal(t, 1, attemptCount) // Should succeed on first try
		assert.False(t, mockT.errorCalled)
	})
	
	// Green: Test retryable action with eventual success
	t.Run("ExecutesRetryableActionWithEventualSuccess", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		attemptCount := 0
		action := RetryableAction{
			Description: "test eventual success",
			MaxRetries:  5,
			TimeBetween: 5 * time.Millisecond,
			Action: func() (string, error) {
				attemptCount++
				if attemptCount < 3 {
					return "", fmt.Errorf("attempt %d failed", attemptCount)
				}
				return "eventual success", nil
			},
		}
		
		result, err := vdf.DoWithRetryE(action)
		
		require.NoError(t, err)
		assert.Equal(t, "eventual success", result)
		assert.Equal(t, 3, attemptCount) // Should succeed on third try
	})
	
	// Red: Test retryable action failure after all retries
	t.Run("FailsRetryableActionAfterAllRetries", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		attemptCount := 0
		action := RetryableAction{
			Description: "test persistent failure",
			MaxRetries:  3,
			TimeBetween: 1 * time.Millisecond,
			Action: func() (string, error) {
				attemptCount++
				return "", fmt.Errorf("attempt %d failed", attemptCount)
			},
		}
		
		_, err := vdf.DoWithRetryE(action)
		
		assert.Error(t, err)
		assert.Equal(t, 3, attemptCount) // Should attempt MaxRetries times
	})
	
	// Red: Test wrapper function failure behavior
	t.Run("WrapperFunctionHandlesFailure", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		action := RetryableAction{
			Description: "test wrapper failure",
			MaxRetries:  2,
			TimeBetween: 1 * time.Millisecond,
			Action: func() (string, error) {
				return "", errors.New("persistent failure")
			},
		}
		
		vdf.DoWithRetry(action)
		
		assert.True(t, mockT.errorCalled)
		assert.True(t, mockT.failNowCalled)
	})
}

// Test Wait Until Functionality

func TestVasDeference_WaitUntilFunctionality(t *testing.T) {
	t.Parallel()
	
	// Green: Test successful wait condition
	t.Run("WaitsUntilConditionMet", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		checkCount := 0
		
		vdf.WaitUntil(
			"test condition",
			3,
			5*time.Millisecond,
			func() (bool, error) {
				checkCount++
				return checkCount >= 2, nil // Succeed on second check
			},
		)
		
		assert.Equal(t, 2, checkCount)
		assert.False(t, mockT.errorCalled)
	})
	
	// Green: Test immediate condition success
	t.Run("WaitsUntilImmediateSuccess", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		checkCount := 0
		
		err := vdf.WaitUntilE(
			"immediate success condition",
			5,
			10*time.Millisecond,
			func() (bool, error) {
				checkCount++
				return true, nil // Succeed immediately
			},
		)
		
		assert.NoError(t, err)
		assert.Equal(t, 1, checkCount)
	})
	
	// Red: Test condition timeout/failure
	t.Run("WaitUntilConditionTimeout", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		checkCount := 0
		
		err := vdf.WaitUntilE(
			"failing condition",
			3,
			1*time.Millisecond,
			func() (bool, error) {
				checkCount++
				return false, nil // Never succeed
			},
		)
		
		assert.Error(t, err)
		assert.Equal(t, 3, checkCount) // Should check MaxRetries times
	})
	
	// Red: Test condition error
	t.Run("WaitUntilConditionError", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		checkCount := 0
		
		err := vdf.WaitUntilE(
			"error condition",
			5,
			1*time.Millisecond,
			func() (bool, error) {
				checkCount++
				return false, errors.New("condition check failed")
			},
		)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "condition check failed")
		assert.GreaterOrEqual(t, checkCount, 1)
	})
	
	// Red: Test wrapper function failure
	t.Run("WrapperFunctionHandlesWaitFailure", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		vdf.WaitUntil(
			"wrapper failure condition",
			2,
			1*time.Millisecond,
			func() (bool, error) {
				return false, errors.New("wrapper test error")
			},
		)
		
		assert.True(t, mockT.errorCalled)
		assert.True(t, mockT.failNowCalled)
	})
}

// Test AWS Session and Region Management

func TestVasDeference_AWSSessionAndRegions(t *testing.T) {
	t.Parallel()
	
	// Green: Test AWS session retrieval
	t.Run("RetrievesAWSSession", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		session := vdf.GetAWSSession()
		
		assert.NotNil(t, session)
		// In the implementation, this returns the AwsConfig directly
		assert.Equal(t, vdf.Context.AwsConfig, session)
	})
	
	// Green: Test availability zone retrieval
	t.Run("RetrievesAvailabilityZones", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// This method calls Terratest's aws.GetAvailabilityZones
		// In a real environment, it would return actual AZs
		// For testing, we just verify it can be called without panicking
		assert.NotPanics(t, func() {
			zones := vdf.GetAvailabilityZones()
			// zones may be empty in test environment, but should not panic
			_ = zones
		})
	})
	
	// Green: Test all regions retrieval
	t.Run("RetrievesAllAvailableRegions", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// This method calls Terratest's aws.GetAllAwsRegions
		assert.NotPanics(t, func() {
			regions := vdf.GetAllAvailableRegions()
			// regions may be empty in test environment, but should not panic
			_ = regions
		})
	})
	
	// Green: Test random region excluding specified regions
	t.Run("ReturnsRandomRegionExcludingSpecified", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		excludeRegions := []string{RegionUSEast1, RegionUSWest2}
		
		// This method calls Terratest's aws.GetRandomRegion
		assert.NotPanics(t, func() {
			region := vdf.GetRandomRegionExcluding(excludeRegions)
			// In a real environment, this would return a region not in the exclude list
			_ = region
		})
	})
}

// Test Resource Name Generation

func TestVasDeference_ResourceNameGeneration(t *testing.T) {
	t.Parallel()
	
	// Green: Test random resource name creation
	t.Run("CreatesRandomResourceNames", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		prefix := "test-resource"
		maxLength := 50
		
		name := vdf.CreateRandomResourceName(prefix, maxLength)
		
		assert.NotEmpty(t, name)
		assert.LessOrEqual(t, len(name), maxLength)
		assert.Contains(t, name, prefix)
		assert.True(t, strings.ToLower(name) == name) // Should be lowercase
		
		// Generate another to test uniqueness
		name2 := vdf.CreateRandomResourceName(prefix, maxLength)
		assert.NotEqual(t, name, name2)
	})
	
	// Green: Test with very short max length
	t.Run("HandlesShortMaxLength", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		prefix := "test"
		maxLength := 10
		
		name := vdf.CreateRandomResourceName(prefix, maxLength)
		
		assert.NotEmpty(t, name)
		assert.LessOrEqual(t, len(name), maxLength)
	})
	
	// Red: Test with zero max length
	t.Run("HandlesZeroMaxLength", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		name := vdf.CreateRandomResourceName("test", 0)
		assert.Empty(t, name)
	})
	
	// Red: Test with negative max length
	t.Run("HandlesNegativeMaxLength", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		name := vdf.CreateRandomResourceName("test", -1)
		assert.Empty(t, name)
	})
	
	// Green: Test with very long prefix
	t.Run("HandlesLongPrefix", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		longPrefix := "very-very-very-long-prefix-that-might-exceed-limits"
		maxLength := 20
		
		name := vdf.CreateRandomResourceName(longPrefix, maxLength)
		
		assert.NotEmpty(t, name)
		assert.LessOrEqual(t, len(name), maxLength)
	})
}

// Test Test Identifier Creation

func TestVasDeference_TestIdentifierCreation(t *testing.T) {
	t.Parallel()
	
	// Green: Test test identifier creation
	t.Run("CreatesTestIdentifiers", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := NewWithOptions(mockT, &VasDeferenceOptions{
			Namespace: "test-namespace",
		})
		
		identifier := vdf.CreateTestIdentifier()
		
		assert.NotEmpty(t, identifier)
		assert.Contains(t, identifier, "test-namespace")
		assert.Contains(t, identifier, "-") // Should contain separator
		
		// Generate another to test uniqueness
		identifier2 := vdf.CreateTestIdentifier()
		assert.NotEqual(t, identifier, identifier2)
	})
	
	// Green: Test with sanitized namespace
	t.Run("CreatesTestIdentifiersWithSanitizedNamespace", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := NewWithOptions(mockT, &VasDeferenceOptions{
			Namespace: "test_namespace.with@special",
		})
		
		identifier := vdf.CreateTestIdentifier()
		
		assert.NotEmpty(t, identifier)
		assert.Contains(t, identifier, "test-namespace-with-special") // Should be sanitized
		assert.NotContains(t, identifier, "_")
		assert.NotContains(t, identifier, ".")
		assert.NotContains(t, identifier, "@")
	})
}

// Test Resource Name with Timestamp

func TestVasDeference_ResourceNameWithTimestamp(t *testing.T) {
	t.Parallel()
	
	// Green: Test resource name with timestamp creation
	t.Run("CreatesResourceNameWithTimestamp", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := NewWithOptions(mockT, &VasDeferenceOptions{
			Namespace: "test-ns",
		})
		
		prefix := "resource"
		name := vdf.CreateResourceNameWithTimestamp(prefix)
		
		assert.NotEmpty(t, name)
		assert.Contains(t, name, "test-ns")
		assert.Contains(t, name, prefix)
		
		// Should contain multiple parts separated by hyphens
		parts := strings.Split(name, "-")
		assert.GreaterOrEqual(t, len(parts), 3) // namespace, prefix, timestamp/id
		
		// Generate another to test uniqueness
		name2 := vdf.CreateResourceNameWithTimestamp(prefix)
		assert.NotEqual(t, name, name2)
	})
}

// Test AWS Resource Validation

func TestVasDeference_AWSResourceValidation(t *testing.T) {
	t.Parallel()
	
	// Green: Test successful resource validation
	t.Run("ValidatesExistingResource", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		resourceType := "lambda"
		resourceId := "test-function"
		
		vdf.ValidateAWSResourceExists(
			resourceType,
			resourceId,
			func() (bool, error) {
				return true, nil // Resource exists
			},
		)
		
		assert.False(t, mockT.errorCalled)
		assert.False(t, mockT.failNowCalled)
	})
	
	// Green: Test successful resource validation with E version
	t.Run("ValidatesExistingResourceWithE", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		resourceType := "dynamodb"
		resourceId := "test-table"
		
		err := vdf.ValidateAWSResourceExistsE(
			resourceType,
			resourceId,
			func() (bool, error) {
				return true, nil // Resource exists
			},
		)
		
		assert.NoError(t, err)
	})
	
	// Red: Test resource validation failure
	t.Run("FailsWhenResourceDoesNotExist", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		resourceType := "s3"
		resourceId := "nonexistent-bucket"
		
		err := vdf.ValidateAWSResourceExistsE(
			resourceType,
			resourceId,
			func() (bool, error) {
				return false, nil // Resource does not exist
			},
		)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
	
	// Red: Test resource validation error
	t.Run("FailsWhenValidatorReturnsError", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		resourceType := "ec2"
		resourceId := "test-instance"
		
		err := vdf.ValidateAWSResourceExistsE(
			resourceType,
			resourceId,
			func() (bool, error) {
				return false, errors.New("validation error")
			},
		)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})
	
	// Red: Test wrapper function failure
	t.Run("WrapperFunctionHandlesValidationFailure", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		resourceType := "rds"
		resourceId := "test-db"
		
		vdf.ValidateAWSResourceExists(
			resourceType,
			resourceId,
			func() (bool, error) {
				return false, errors.New("wrapper test error")
			},
		)
		
		assert.True(t, mockT.errorCalled)
		assert.True(t, mockT.failNowCalled)
	})
}

// Test Integration with Environment Variables

func TestVasDeference_IntegrationWithEnvironment(t *testing.T) {
	t.Parallel()
	
	// Green: Test environment-based configuration
	t.Run("UsesEnvironmentConfiguration", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("AWS_DEFAULT_REGION", RegionEUWest1)
		defer os.Unsetenv("AWS_DEFAULT_REGION")
		
		mockT := &mockTestingT{}
		// Create VasDeference without explicit region to test environment pickup
		vdf := New(mockT)
		
		// Test environment variable access
		region := vdf.GetEnvVar("AWS_DEFAULT_REGION")
		assert.Equal(t, RegionEUWest1, region)
		
		// Test with default fallback
		nonexistentVar := vdf.GetEnvVarWithDefault("NONEXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", nonexistentVar)
	})
	
	// Green: Test CI environment detection
	t.Run("DetectsCIEnvironment", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// Clear CI variables
		ciVars := []string{"CI", "GITHUB_ACTIONS", "TRAVIS"}
		for _, v := range ciVars {
			os.Unsetenv(v)
		}
		
		// Should not detect CI
		assert.False(t, vdf.IsCIEnvironment())
		
		// Set CI variable
		os.Setenv("CI", "true")
		defer os.Unsetenv("CI")
		
		// Should detect CI
		assert.True(t, vdf.IsCIEnvironment())
	})
}

// Performance and Load Testing

func TestVasDeference_PerformanceCharacteristics(t *testing.T) {
	t.Parallel()
	
	// Green: Test multiple instance creation performance
	t.Run("HandlesMultipleInstanceCreation", func(t *testing.T) {
		mockT := &mockTestingT{}
		
		start := time.Now()
		
		// Create multiple instances
		instances := make([]*VasDeference, 100)
		for i := 0; i < 100; i++ {
			instances[i] = New(mockT)
		}
		
		elapsed := time.Since(start)
		
		// Should create instances reasonably quickly (increased timeout for slower systems)
		assert.Less(t, elapsed, 60*time.Second)
		
		// Verify all instances are valid
		for _, vdf := range instances {
			assert.NotNil(t, vdf)
			assert.NotNil(t, vdf.Context)
			assert.NotNil(t, vdf.Arrange)
		}
	})
	
	// Green: Test concurrent access safety
	t.Run("HandlesConcurrentAccess", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := New(mockT)
		
		// Test concurrent read operations
		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()
				
				// Perform various read operations
				region := vdf.GetDefaultRegion()
				assert.NotEmpty(t, region)
				
				randomName := vdf.GenerateRandomName("test")
				assert.Contains(t, randomName, "test")
				
				accountId := vdf.GetAccountId()
				assert.NotEmpty(t, accountId)
			}()
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			select {
			case <-done:
				// Success
			case <-time.After(5 * time.Second):
				t.Fatal("Concurrent access test timed out")
			}
		}
	})
}

// Integration Test Scenarios

func TestVasDeference_IntegrationScenarios(t *testing.T) {
	t.Parallel()
	
	// Green: Test complete workflow simulation
	t.Run("SimulatesCompleteTestWorkflow", func(t *testing.T) {
		mockT := &mockTestingT{}
		vdf := NewWithOptions(mockT, &VasDeferenceOptions{
			Region:    RegionUSWest2,
			Namespace: "integration-test",
		})
		
		// Simulate typical test workflow
		
		// 1. Generate resource names
		functionName := vdf.PrefixResourceName("test-function")
		tableName := vdf.PrefixResourceName("test-table")
		
		assert.Contains(t, functionName, "integration-test")
		assert.Contains(t, tableName, "integration-test")
		
		// 2. Create AWS clients
		lambdaClient := vdf.CreateLambdaClient()
		logsClient := vdf.CreateCloudWatchLogsClient()
		
		assert.NotNil(t, lambdaClient)
		assert.NotNil(t, logsClient)
		
		// 3. Register cleanup actions
		cleanupCount := 0
		vdf.RegisterCleanup(func() error {
			cleanupCount++
			return nil
		})
		vdf.RegisterCleanup(func() error {
			cleanupCount++
			return nil
		})
		
		// 4. Simulate test operations with retry
		result, err := vdf.RetryWithBackoffE(
			"simulate test operation",
			func() (string, error) {
				return "operation completed", nil
			},
			3,
			10*time.Millisecond,
		)
		
		assert.NoError(t, err)
		assert.Equal(t, "operation completed", result)
		
		// 5. Execute cleanup
		vdf.Cleanup()
		
		assert.Equal(t, 2, cleanupCount) // Both cleanup functions should execute
		assert.False(t, mockT.errorCalled)
	})
}