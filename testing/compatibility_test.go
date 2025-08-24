package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// RED: Test that *testing.T implements our TestingT interface
func TestCompatibility_ShouldWorkWithRealTestingT(t *testing.T) {
	// This test validates that a real *testing.T can be used with our framework
	
	// Act - Pass real *testing.T to a function expecting our interface
	result := useTestingTInterface(t)
	
	// Assert
	assert.True(t, result)
}

// RED: Test that the framework works with both mock and real testing interfaces
func TestCompatibility_ShouldWorkWithBothMockAndReal(t *testing.T) {
	// Test with real *testing.T
	realResult := useTestingTInterface(t)
	
	// Test with mock
	mockT := NewMockTestingT("mock-compatibility")
	mockResult := useTestingTInterface(mockT)
	
	// Both should work
	assert.True(t, realResult)
	assert.True(t, mockResult)
	
	// Verify mock captured the call
	assert.True(t, mockT.WasHelperCalled())
}

// RED: Test interface method compatibility
func TestCompatibility_ShouldHaveIdenticalMethods(t *testing.T) {
	// This test ensures all required methods exist and work correctly
	
	mockT := NewMockTestingT("method-test")
	
	// Test Helper method
	mockT.Helper()
	assert.True(t, mockT.WasHelperCalled())
	
	// Test Errorf method
	mockT.Errorf("Test error: %s", "message")
	assert.True(t, mockT.WasErrorfCalled())
	assert.Contains(t, mockT.GetErrorMessages()[0], "Test error: message")
	
	// Test Logf method
	mockT.Logf("Test log: %s", "message")
	assert.True(t, mockT.WasLogfCalled())
	assert.Contains(t, mockT.GetLogMessages()[0], "Test log: message")
	
	// Test FailNow method
	mockT.FailNow()
	assert.True(t, mockT.WasFailNowCalled())
	
	// Test Name method
	assert.Equal(t, "method-test", mockT.Name())
}

// RED: Test that framework error scenarios don't break the test runner
func TestCompatibility_ShouldIsolateFrameworkErrors(t *testing.T) {
	// This is critical: framework errors should not cause the test runner to fail
	
	// Arrange
	mockT := NewMockTestingT("error-isolation-test")
	
	// Act - Simulate framework encountering various error conditions
	simulateFrameworkFailureScenarios(mockT)
	
	// Assert - Test runner should still be working (we're here!)
	// The mock should have captured all the errors
	assert.True(t, mockT.WasErrorfCalled())
	assert.True(t, mockT.WasFailNowCalled())
	
	// We should be able to examine what errors occurred
	errors := mockT.GetErrorMessages()
	assert.NotEmpty(t, errors)
	
	// The test itself should not have failed
	assert.True(t, true) // We made it here, so test isolation works
}

// RED: Test concurrent usage doesn't cause race conditions
func TestCompatibility_ShouldSupportConcurrentUsage(t *testing.T) {
	// Test that multiple goroutines can use different mock instances safely
	
	numGoroutines := 10
	results := make([]bool, numGoroutines)
	done := make(chan bool, numGoroutines)
	
	// Launch concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer func() { done <- true }()
			
			mockT := NewMockTestingT("concurrent-test-" + string(rune(index+'0')))
			mockT.Helper()
			mockT.Errorf("Error from goroutine %d", index)
			mockT.Logf("Log from goroutine %d", index)
			
			results[index] = mockT.WasHelperCalled() && 
							mockT.WasErrorfCalled() && 
							mockT.WasLogfCalled()
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Assert - All should have succeeded
	for i, result := range results {
		assert.True(t, result, "Goroutine %d failed", i)
	}
}

// RED: Test that framework behavior is predictable and testable
func TestCompatibility_ShouldProvideTestableFrameworkBehavior(t *testing.T) {
	// Test that we can predictably test framework behavior patterns
	
	testCases := []struct {
		name           string
		operation      func(TestingT)
		expectedHelper bool
		expectedError  bool
		expectedLog    bool
		expectedFail   bool
	}{
		{
			name: "successful operation",
			operation: func(tt TestingT) {
				tt.Helper()
			},
			expectedHelper: true,
			expectedError:  false,
			expectedLog:    false,
			expectedFail:   false,
		},
		{
			name: "error operation",
			operation: func(tt TestingT) {
				tt.Helper()
				tt.Errorf("Error occurred")
			},
			expectedHelper: true,
			expectedError:  true,
			expectedLog:    false,
			expectedFail:   false,
		},
		{
			name: "failure operation",
			operation: func(tt TestingT) {
				tt.Helper()
				tt.Errorf("Critical error")
				tt.FailNow()
			},
			expectedHelper: true,
			expectedError:  true,
			expectedLog:    false,
			expectedFail:   true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockT := NewMockTestingT("predictable-test")
			
			// Act
			tc.operation(mockT)
			
			// Assert
			assert.Equal(t, tc.expectedHelper, mockT.WasHelperCalled(), "Helper call mismatch")
			assert.Equal(t, tc.expectedError, mockT.WasErrorfCalled(), "Error call mismatch")
			assert.Equal(t, tc.expectedLog, mockT.WasLogfCalled(), "Log call mismatch")
			assert.Equal(t, tc.expectedFail, mockT.WasFailNowCalled(), "Fail call mismatch")
		})
	}
}

// Helper functions

// useTestingTInterface demonstrates using our TestingT interface
func useTestingTInterface(t TestingT) bool {
	t.Helper()
	
	// Perform some testing operations
	name := t.Name()
	if name == "" {
		t.Errorf("Test name should not be empty")
		return false
	}
	
	t.Logf("Testing with: %s", name)
	return true
}

// simulateFrameworkFailureScenarios tests various error conditions
func simulateFrameworkFailureScenarios(t TestingT) {
	t.Helper()
	
	// Simulate various types of framework errors
	t.Errorf("Simulated configuration error")
	t.Errorf("Simulated AWS connection error")
	t.Errorf("Simulated resource creation error")
	
	// Simulate a critical failure
	t.FailNow()
}