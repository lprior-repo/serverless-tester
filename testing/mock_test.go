package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test that MockTestingT implements the TestingT interface
func TestMockTestingT_ShouldImplementTestingTInterface(t *testing.T) {
	// Arrange
	var mockT TestingT

	// Act
	mockT = NewMockTestingT("test-mock")

	// Assert
	require.NotNil(t, mockT)
	assert.Equal(t, "test-mock", mockT.Name())
}

// RED: Test that MockTestingT tracks all method calls
func TestMockTestingT_ShouldTrackAllMethodCalls(t *testing.T) {
	// Arrange
	mockT := NewMockTestingT("test-tracking")

	// Act
	mockT.Helper()
	mockT.Errorf("Error message: %s", "test error")
	mockT.Logf("Log message: %s", "test log")
	mockT.FailNow()

	// Assert
	assert.True(t, mockT.WasHelperCalled())
	assert.True(t, mockT.WasErrorfCalled())
	assert.True(t, mockT.WasLogfCalled())
	assert.True(t, mockT.WasFailNowCalled())
	
	// Verify messages were captured
	errors := mockT.GetErrorMessages()
	require.Len(t, errors, 1)
	assert.Equal(t, "Error message: test error", errors[0])
	
	logs := mockT.GetLogMessages()
	require.Len(t, logs, 1)
	assert.Equal(t, "Log message: test log", logs[0])
}

// RED: Test that MockTestingT can be reset for reuse
func TestMockTestingT_ShouldAllowReset(t *testing.T) {
	// Arrange
	mockT := NewMockTestingT("test-reset")
	
	// Setup initial state
	mockT.Helper()
	mockT.Errorf("Initial error")
	mockT.FailNow()

	// Act - Reset the mock
	mockT.Reset()

	// Assert - All state should be cleared
	assert.False(t, mockT.WasHelperCalled())
	assert.False(t, mockT.WasErrorfCalled())
	assert.False(t, mockT.WasLogfCalled())
	assert.False(t, mockT.WasFailNowCalled())
	assert.Empty(t, mockT.GetErrorMessages())
	assert.Empty(t, mockT.GetLogMessages())
}

// RED: Test multiple mock instances for parallel testing
func TestMockTestingT_ShouldSupportMultipleInstances(t *testing.T) {
	// Arrange
	mockT1 := NewMockTestingT("test-1")
	mockT2 := NewMockTestingT("test-2")

	// Act
	mockT1.Errorf("Error from mock 1")
	mockT2.Errorf("Error from mock 2")

	// Assert - Each mock should track its own state
	errors1 := mockT1.GetErrorMessages()
	errors2 := mockT2.GetErrorMessages()
	
	require.Len(t, errors1, 1)
	require.Len(t, errors2, 1)
	assert.Equal(t, "Error from mock 1", errors1[0])
	assert.Equal(t, "Error from mock 2", errors2[0])
	
	// Names should be different
	assert.Equal(t, "test-1", mockT1.Name())
	assert.Equal(t, "test-2", mockT2.Name())
}

// RED: Test that MockTestingT can verify call counts
func TestMockTestingT_ShouldTrackCallCounts(t *testing.T) {
	// Arrange
	mockT := NewMockTestingT("test-counts")

	// Act
	mockT.Helper()
	mockT.Helper()
	mockT.Errorf("Error 1")
	mockT.Errorf("Error 2")
	mockT.Errorf("Error 3")
	mockT.Logf("Log 1")

	// Assert
	assert.Equal(t, 2, mockT.GetHelperCallCount())
	assert.Equal(t, 3, mockT.GetErrorfCallCount())
	assert.Equal(t, 1, mockT.GetLogfCallCount())
	assert.Equal(t, 0, mockT.GetFailNowCallCount())
}

// RED: Test compatibility with *testing.T interface
func TestMockTestingT_ShouldBeCompatibleWithRealTestingT(t *testing.T) {
	// This test verifies that we can use MockTestingT anywhere TestingT is expected
	
	// Arrange
	mockT := NewMockTestingT("compatibility-test")
	
	// Act - Use in a function that expects TestingT
	useMockInFramework(mockT)
	
	// Assert - The mock should have been called
	assert.True(t, mockT.WasHelperCalled())
	assert.True(t, mockT.WasErrorfCalled())
}

// Helper function to test TestingT interface compatibility
func useMockInFramework(t TestingT) {
	t.Helper()
	t.Errorf("Framework called mock")
}