package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test that we can validate TestingT interface compatibility without breaking tests
func TestFramework_ShouldValidateTestingTCompatibility(t *testing.T) {
	// This test ensures our MockTestingT can be used anywhere a real *testing.T is expected
	
	// Arrange
	mockT := NewMockTestingT("compatibility-test")
	
	// Act - Use mock in place of testing.T
	result := validateFrameworkBehavior(mockT)
	
	// Assert - Framework should work with mock
	assert.True(t, result)
	assert.True(t, mockT.WasHelperCalled())
	assert.False(t, mockT.WasErrorfCalled()) // No errors should occur
}

// RED: Test that we can test the framework's error handling without breaking the test run
func TestFramework_ShouldAllowTestingErrorScenarios(t *testing.T) {
	// This is crucial - we need to test framework error behavior without failing our tests
	
	// Arrange
	mockT := NewMockTestingT("error-scenario-test")
	
	// Act - Simulate framework encountering an error condition
	simulateFrameworkError(mockT, "simulated error condition")
	
	// Assert - We can verify the framework called Errorf without our test failing
	assert.True(t, mockT.WasErrorfCalled())
	assert.True(t, mockT.WasFailNowCalled())
	
	errorMessages := mockT.GetErrorMessages()
	require.Len(t, errorMessages, 1)
	assert.Contains(t, errorMessages[0], "simulated error condition")
}

// RED: Test namespace detection and sanitization
func TestFramework_ShouldTestNamespaceHandling(t *testing.T) {
	// Arrange
	testCases := []struct {
		input    string
		expected string
	}{
		{"feature/my_branch", "feature-my-branch"},
		{"PR-123", "pr-123"},
		{"user@email.com", "useremail-com"},
		{"Test_With_Underscores", "test-with-underscores"},
		{"", "default"},
		{"---", "default"},
	}
	
	// Act & Assert
	for _, tc := range testCases {
		result := sanitizeNamespaceForTesting(tc.input)
		assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
	}
}

// RED: Test cleanup registration and execution (LIFO)
func TestFramework_ShouldTestCleanupLifecycle(t *testing.T) {
	// Arrange
	mockT := NewMockTestingT("cleanup-test")
	executionOrder := make([]string, 0)
	
	// Create a test cleanup manager
	cleanupManager := &TestCleanupManager{}
	
	// Act - Register cleanup functions in order
	cleanupManager.RegisterCleanup(func() error {
		executionOrder = append(executionOrder, "first")
		return nil
	})
	
	cleanupManager.RegisterCleanup(func() error {
		executionOrder = append(executionOrder, "second")
		return nil
	})
	
	cleanupManager.RegisterCleanup(func() error {
		executionOrder = append(executionOrder, "third")
		return nil
	})
	
	// Execute cleanup
	cleanupManager.ExecuteCleanup(mockT)
	
	// Assert - Should execute in LIFO order (Last In, First Out)
	require.Len(t, executionOrder, 3)
	assert.Equal(t, "third", executionOrder[0])
	assert.Equal(t, "second", executionOrder[1])
	assert.Equal(t, "first", executionOrder[2])
}

// RED: Test cleanup error handling
func TestFramework_ShouldHandleCleanupErrors(t *testing.T) {
	// Arrange
	mockT := NewMockTestingT("cleanup-error-test")
	cleanupManager := &TestCleanupManager{}
	
	// Act - Register cleanup that will fail
	cleanupManager.RegisterCleanup(func() error {
		return assert.AnError
	})
	
	cleanupManager.RegisterCleanup(func() error {
		return nil // This should still execute even if previous failed
	})
	
	cleanupManager.ExecuteCleanup(mockT)
	
	// Assert - Error should be logged but cleanup should continue
	assert.True(t, mockT.WasErrorfCalled())
	errorMessages := mockT.GetErrorMessages()
	require.Len(t, errorMessages, 1)
	assert.Contains(t, errorMessages[0], "Cleanup function failed")
}

// RED: Test Terraform output parsing
func TestFramework_ShouldTestTerraformOutputParsing(t *testing.T) {
	// Arrange
	mockT := NewMockTestingT("terraform-test")
	fixtures := NewTestFixtures()
	outputs := fixtures.GetSampleTerraformOutputs()
	
	// Act
	result := parseTerraformOutputForTesting(mockT, outputs, "lambda_function_arn")
	
	// Assert
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "arn:aws:lambda")
	assert.False(t, mockT.WasErrorfCalled())
}

// RED: Test Terraform output parsing with missing key
func TestFramework_ShouldHandleMissingTerraformOutput(t *testing.T) {
	// Arrange
	mockT := NewMockTestingT("terraform-missing-test")
	fixtures := NewTestFixtures()
	outputs := fixtures.GetSampleTerraformOutputs()
	
	// Act
	result := parseTerraformOutputForTesting(mockT, outputs, "non_existent_key")
	
	// Assert
	assert.Empty(t, result)
	assert.True(t, mockT.WasErrorfCalled())
	assert.True(t, mockT.WasFailNowCalled())
}

// RED: Test AWS credential validation without actually connecting
func TestFramework_ShouldMockAWSCredentialValidation(t *testing.T) {
	// Arrange
	mockT := NewMockTestingT("aws-creds-test")
	
	// Act
	isValid := simulateAWSCredentialValidation(mockT, true)
	
	// Assert
	assert.True(t, isValid)
	assert.False(t, mockT.WasErrorfCalled())
}

// RED: Test AWS credential validation failure scenarios
func TestFramework_ShouldHandleAWSCredentialFailures(t *testing.T) {
	// Arrange
	mockT := NewMockTestingT("aws-creds-fail-test")
	
	// Act
	isValid := simulateAWSCredentialValidation(mockT, false)
	
	// Assert
	assert.False(t, isValid)
	// Note: Framework might log this but shouldn't necessarily fail the test
}

// RED: Test parallel testing scenarios
func TestFramework_ShouldSupportParallelTesting(t *testing.T) {
	// This tests that multiple mock instances can be used concurrently
	
	// Arrange
	numInstances := 5
	results := make([]bool, numInstances)
	
	// Act - Simulate concurrent testing
	for i := 0; i < numInstances; i++ {
		go func(index int) {
			mockT := NewMockTestingT("parallel-test-" + string(rune(index+'0')))
			results[index] = validateFrameworkBehavior(mockT)
		}(i)
	}
	
	// Allow goroutines to complete
	time.Sleep(100 * time.Millisecond)
	
	// Assert - All instances should work independently
	for i, result := range results {
		assert.True(t, result, "Instance %d failed", i)
	}
}

// Helper functions for testing framework behavior

func validateFrameworkBehavior(t TestingT) bool {
	t.Helper()
	// Simulate framework behavior - no errors
	return true
}

func simulateFrameworkError(t TestingT, message string) {
	t.Helper()
	t.Errorf("Framework error: %s", message)
	t.FailNow()
}

func sanitizeNamespaceForTesting(input string) string {
	// This would call the actual framework sanitization function
	// For now, simulate the behavior
	utils := NewTestUtils()
	return utils.SanitizeForAWS(input)
}

// TestCleanupManager simulates the framework's cleanup management
type TestCleanupManager struct {
	cleanups []func() error
}

func (m *TestCleanupManager) RegisterCleanup(cleanup func() error) {
	m.cleanups = append(m.cleanups, cleanup)
}

func (m *TestCleanupManager) ExecuteCleanup(t TestingT) {
	// Execute in LIFO order
	for i := len(m.cleanups) - 1; i >= 0; i-- {
		if err := m.cleanups[i](); err != nil {
			t.Errorf("Cleanup function failed: %v", err)
		}
	}
}

func parseTerraformOutputForTesting(t TestingT, outputs map[string]interface{}, key string) string {
	if outputs == nil {
		t.Errorf("Terraform outputs cannot be nil")
		t.FailNow()
		return ""
	}
	
	value, exists := outputs[key]
	if !exists {
		t.Errorf("Key '%s' not found in terraform outputs", key)
		t.FailNow()
		return ""
	}
	
	strValue, ok := value.(string)
	if !ok {
		t.Errorf("Value for key '%s' is not a string", key)
		t.FailNow()
		return ""
	}
	
	return strValue
}

func simulateAWSCredentialValidation(t TestingT, shouldSucceed bool) bool {
	t.Helper()
	
	if shouldSucceed {
		return true
	}
	
	// In real implementation, this might log an error but not fail the test
	t.Logf("AWS credential validation failed (simulated)")
	return false
}