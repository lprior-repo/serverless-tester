package helpers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockTestingT for assertion testing
type MockTestingT struct {
	errorCalled bool
	errorMsg    string
}

func (m *MockTestingT) Helper() {}
func (m *MockTestingT) Errorf(format string, args ...interface{}) {
	m.errorCalled = true
	m.errorMsg = fmt.Sprintf(format, args...)
}
func (m *MockTestingT) FailNow() {}
func (m *MockTestingT) Logf(format string, args ...interface{}) {}
func (m *MockTestingT) Name() string { return "mock-test" }

// RED: Test that AssertNoError succeeds when error is nil
func TestAssertNoError_ShouldSucceedWhenErrorIsNil(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertNoError(mockT, nil, "operation should succeed")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when error is nil")
}

// RED: Test that AssertNoError fails when error is present
func TestAssertNoError_ShouldFailWhenErrorIsPresent(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	testError := errors.New("test error")
	
	// Act
	AssertNoError(mockT, testError, "operation should succeed")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when error is present")
	assert.Contains(t, mockT.errorMsg, "operation should succeed", "Error message should contain context")
}

// RED: Test that AssertError succeeds when error is present
func TestAssertError_ShouldSucceedWhenErrorIsPresent(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	testError := errors.New("expected error")
	
	// Act
	AssertError(mockT, testError, "operation should fail")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when error is present as expected")
}

// RED: Test that AssertError fails when error is nil
func TestAssertError_ShouldFailWhenErrorIsNil(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertError(mockT, nil, "operation should fail")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when error is nil but expected")
}

// RED: Test that AssertErrorContains succeeds when error contains expected text
func TestAssertErrorContains_ShouldSucceedWhenErrorContainsText(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	testError := errors.New("resource not found")
	
	// Act
	AssertErrorContains(mockT, testError, "not found", "error should contain specific text")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when error contains expected text")
}

// RED: Test that AssertErrorContains fails when error doesn't contain expected text
func TestAssertErrorContains_ShouldFailWhenErrorDoesNotContainText(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	testError := errors.New("access denied")
	
	// Act
	AssertErrorContains(mockT, testError, "not found", "error should contain specific text")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when error doesn't contain expected text")
}

// RED: Test that AssertStringNotEmpty succeeds when string is not empty
func TestAssertStringNotEmpty_ShouldSucceedWhenStringIsNotEmpty(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertStringNotEmpty(mockT, "test value", "string should not be empty")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when string is not empty")
}

// RED: Test that AssertStringNotEmpty fails when string is empty
func TestAssertStringNotEmpty_ShouldFailWhenStringIsEmpty(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertStringNotEmpty(mockT, "", "string should not be empty")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when string is empty")
}

// RED: Test that AssertStringContains succeeds when string contains expected text
func TestAssertStringContains_ShouldSucceedWhenStringContainsText(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertStringContains(mockT, "hello world", "world", "string should contain expected text")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when string contains expected text")
}

// RED: Test that AssertStringContains fails when string doesn't contain expected text
func TestAssertStringContains_ShouldFailWhenStringDoesNotContainText(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertStringContains(mockT, "hello world", "goodbye", "string should contain expected text")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when string doesn't contain expected text")
}

// RED: Test that AssertGreaterThan succeeds when value is greater
func TestAssertGreaterThan_ShouldSucceedWhenValueIsGreater(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertGreaterThan(mockT, 10, 5, "value should be greater")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when value is greater")
}

// RED: Test that AssertGreaterThan fails when value is not greater
func TestAssertGreaterThan_ShouldFailWhenValueIsNotGreater(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertGreaterThan(mockT, 5, 10, "value should be greater")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when value is not greater")
}

// RED: Test that AssertLessThan succeeds when value is less
func TestAssertLessThan_ShouldSucceedWhenValueIsLess(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertLessThan(mockT, 5, 10, "value should be less")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when value is less")
}

// RED: Test that AssertLessThan fails when value is not less
func TestAssertLessThan_ShouldFailWhenValueIsNotLess(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertLessThan(mockT, 10, 5, "value should be less")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when value is not less")
}

// RED: Test that AssertEqual succeeds when values are equal
func TestAssertEqual_ShouldSucceedWhenValuesAreEqual(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertEqual(mockT, "expected", "expected", "values should be equal")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when values are equal")
}

// RED: Test that AssertEqual fails when values are not equal
func TestAssertEqual_ShouldFailWhenValuesAreNotEqual(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertEqual(mockT, "expected", "actual", "values should be equal")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when values are not equal")
}

// RED: Test that AssertNotEqual succeeds when values are different
func TestAssertNotEqual_ShouldSucceedWhenValuesAreDifferent(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertNotEqual(mockT, "expected", "different", "values should be different")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when values are different")
}

// RED: Test that AssertNotEqual fails when values are equal
func TestAssertNotEqual_ShouldFailWhenValuesAreEqual(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertNotEqual(mockT, "same", "same", "values should be different")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when values are equal but should be different")
}

// RED: Test that AssertTrue succeeds when condition is true
func TestAssertTrue_ShouldSucceedWhenConditionIsTrue(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertTrue(mockT, true, "condition should be true")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when condition is true")
}

// RED: Test that AssertTrue fails when condition is false
func TestAssertTrue_ShouldFailWhenConditionIsFalse(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertTrue(mockT, false, "condition should be true")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when condition is false")
}

// RED: Test that AssertFalse succeeds when condition is false
func TestAssertFalse_ShouldSucceedWhenConditionIsFalse(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertFalse(mockT, false, "condition should be false")
	
	// Assert
	assert.False(t, mockT.errorCalled, "Should not call Errorf when condition is false")
}

// RED: Test that AssertFalse fails when condition is true
func TestAssertFalse_ShouldFailWhenConditionIsTrue(t *testing.T) {
	// Arrange
	mockT := &MockTestingT{}
	
	// Act
	AssertFalse(mockT, true, "condition should be false")
	
	// Assert
	assert.True(t, mockT.errorCalled, "Should call Errorf when condition is true but should be false")
}