package helpers

import (
	"fmt"
	"log/slog"
	"strings"
)

// TestingTInterface represents the minimal testing interface for assertions
// Following Terratest patterns for test framework integration
type TestingTInterface interface {
	Helper()
	Errorf(format string, args ...interface{})
}

// AssertNoError asserts that the error is nil
// Following Terratest patterns for error assertions
func AssertNoError(t TestingTInterface, err error, msg string) {
	t.Helper()
	
	if err != nil {
		slog.Error("Assertion failed: expected no error",
			"error", err.Error(),
			"message", msg,
		)
		t.Errorf("%s: expected no error but got: %v", msg, err)
	}
}

// AssertError asserts that an error is present
// Following Terratest patterns for error assertions
func AssertError(t TestingTInterface, err error, msg string) {
	t.Helper()
	
	if err == nil {
		slog.Error("Assertion failed: expected error but got nil",
			"message", msg,
		)
		t.Errorf("%s: expected error but got nil", msg)
	}
}

// AssertErrorContains asserts that an error is present and contains specific text
// Following Terratest patterns for error content assertions
func AssertErrorContains(t TestingTInterface, err error, expectedText, msg string) {
	t.Helper()
	
	if err == nil {
		slog.Error("Assertion failed: expected error but got nil",
			"expectedText", expectedText,
			"message", msg,
		)
		t.Errorf("%s: expected error containing '%s' but got nil", msg, expectedText)
		return
	}
	
	if !strings.Contains(err.Error(), expectedText) {
		slog.Error("Assertion failed: error does not contain expected text",
			"error", err.Error(),
			"expectedText", expectedText,
			"message", msg,
		)
		t.Errorf("%s: expected error to contain '%s' but got: %v", msg, expectedText, err)
	}
}

// AssertStringNotEmpty asserts that a string is not empty
// Following Terratest patterns for string assertions
func AssertStringNotEmpty(t TestingTInterface, value, msg string) {
	t.Helper()
	
	if value == "" {
		slog.Error("Assertion failed: string is empty",
			"message", msg,
		)
		t.Errorf("%s: string is empty", msg)
	}
}

// AssertStringContains asserts that a string contains expected text
// Following Terratest patterns for string content assertions
func AssertStringContains(t TestingTInterface, str, expectedText, msg string) {
	t.Helper()
	
	if !strings.Contains(str, expectedText) {
		slog.Error("Assertion failed: string does not contain expected text",
			"string", str,
			"expectedText", expectedText,
			"message", msg,
		)
		t.Errorf("%s: expected string to contain '%s' but got: %s", msg, expectedText, str)
	}
}

// AssertGreaterThan asserts that a numeric value is greater than expected
// Following Terratest patterns for numeric assertions
func AssertGreaterThan(t TestingTInterface, actual, expected int, msg string) {
	t.Helper()
	
	if actual <= expected {
		slog.Error("Assertion failed: value is not greater than expected",
			"actual", actual,
			"expected", expected,
			"message", msg,
		)
		t.Errorf("%s: expected %v to be greater than %v", msg, actual, expected)
	}
}

// AssertLessThan asserts that a numeric value is less than expected
// Following Terratest patterns for numeric assertions
func AssertLessThan(t TestingTInterface, actual, expected int, msg string) {
	t.Helper()
	
	if actual >= expected {
		slog.Error("Assertion failed: value is not less than expected",
			"actual", actual,
			"expected", expected,
			"message", msg,
		)
		t.Errorf("%s: expected %v to be less than %v", msg, actual, expected)
	}
}

// AssertEqual asserts that two values are equal
// Following Terratest patterns for equality assertions
func AssertEqual[T comparable](t TestingTInterface, expected, actual T, msg string) {
	t.Helper()
	
	if expected != actual {
		slog.Error("Assertion failed: values are not equal",
			"expected", fmt.Sprintf("%v", expected),
			"actual", fmt.Sprintf("%v", actual),
			"message", msg,
		)
		t.Errorf("%s: expected %v but got %v", msg, expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal
// Following Terratest patterns for inequality assertions
func AssertNotEqual[T comparable](t TestingTInterface, notExpected, actual T, msg string) {
	t.Helper()
	
	if notExpected == actual {
		slog.Error("Assertion failed: values are equal but should be different",
			"notExpected", fmt.Sprintf("%v", notExpected),
			"actual", fmt.Sprintf("%v", actual),
			"message", msg,
		)
		t.Errorf("%s: expected values to be different but both were %v", msg, actual)
	}
}

// AssertTrue asserts that a condition is true
// Following Terratest patterns for boolean assertions
func AssertTrue(t TestingTInterface, condition bool, msg string) {
	t.Helper()
	
	if !condition {
		slog.Error("Assertion failed: condition is false",
			"message", msg,
		)
		t.Errorf("%s: expected condition to be true but it was false", msg)
	}
}

// AssertFalse asserts that a condition is false
// Following Terratest patterns for boolean assertions
func AssertFalse(t TestingTInterface, condition bool, msg string) {
	t.Helper()
	
	if condition {
		slog.Error("Assertion failed: condition is true",
			"message", msg,
		)
		t.Errorf("%s: expected condition to be false but it was true", msg)
	}
}