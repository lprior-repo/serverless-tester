package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

// Test data factories for assertion testing

func createTestExecutionPattern() *ExecutionPattern {
	status := types.ExecutionStatusSucceeded
	minTime := 1 * time.Second
	maxTime := 10 * time.Second
	
	return &ExecutionPattern{
		Status:           &status,
		OutputContains:   "success",
		OutputJSON:       map[string]interface{}{"result": "success", "message": "Hello, World!"},
		ErrorContains:    "",
		MinExecutionTime: &minTime,
		MaxExecutionTime: &maxTime,
	}
}

func createSuccessfulExecutionResult() *ExecutionResult {
	return &ExecutionResult{
		ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
		Name:            "test-execution",
		Status:          types.ExecutionStatusSucceeded,
		StartDate:       time.Now().Add(-5 * time.Minute),
		StopDate:        time.Now(),
		Input:           `{"message": "Hello, World!"}`,
		Output:          `{"result": "success", "message": "Hello, World!"}`,
		ExecutionTime:   5 * time.Second,
	}
}

func createFailedExecutionResult() *ExecutionResult {
	return &ExecutionResult{
		ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
		Name:            "test-execution",
		Status:          types.ExecutionStatusFailed,
		StartDate:       time.Now().Add(-2 * time.Minute),
		StopDate:        time.Now(),
		Input:           `{"message": "Hello, World!"}`,
		Output:          "",
		Error:           "States.TaskFailed",
		Cause:           "Lambda function failed",
		ExecutionTime:   2 * time.Minute,
	}
}

func createTimedOutExecutionResult() *ExecutionResult {
	return &ExecutionResult{
		ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
		Name:            "test-execution",
		Status:          types.ExecutionStatusTimedOut,
		StartDate:       time.Now().Add(-15 * time.Minute),
		StopDate:        time.Now(),
		Input:           `{"message": "Hello, World!"}`,
		Output:          "",
		Error:           "States.Timeout",
		Cause:           "Execution timed out",
		ExecutionTime:   15 * time.Minute,
	}
}

// Unit tests for execution status assertions

func TestAssertExecutionSucceededE(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		expected    bool
		description string
	}{
		{
			name:        "SuccessfulExecution",
			result:      createSuccessfulExecutionResult(),
			expected:    true,
			description: "Successful execution should pass assertion",
		},
		{
			name:        "FailedExecution",
			result:      createFailedExecutionResult(),
			expected:    false,
			description: "Failed execution should fail assertion",
		},
		{
			name:        "TimedOutExecution",
			result:      createTimedOutExecutionResult(),
			expected:    false,
			description: "Timed out execution should fail assertion",
		},
		{
			name:        "NilResult",
			result:      nil,
			expected:    false,
			description: "Nil result should fail assertion",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := AssertExecutionSucceededE(tc.result)
			
			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestAssertExecutionFailedE(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		expected    bool
		description string
	}{
		{
			name:        "FailedExecution",
			result:      createFailedExecutionResult(),
			expected:    true,
			description: "Failed execution should pass assertion",
		},
		{
			name:        "TimedOutExecution",
			result:      createTimedOutExecutionResult(),
			expected:    true,
			description: "Timed out execution should pass assertion",
		},
		{
			name:        "SuccessfulExecution",
			result:      createSuccessfulExecutionResult(),
			expected:    false,
			description: "Successful execution should fail assertion",
		},
		{
			name:        "NilResult",
			result:      nil,
			expected:    false,
			description: "Nil result should fail assertion",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := AssertExecutionFailedE(tc.result)
			
			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestAssertExecutionTimedOutE(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		expected    bool
		description string
	}{
		{
			name:        "TimedOutExecution",
			result:      createTimedOutExecutionResult(),
			expected:    true,
			description: "Timed out execution should pass assertion",
		},
		{
			name:        "FailedExecution",
			result:      createFailedExecutionResult(),
			expected:    false,
			description: "Failed execution should fail assertion",
		},
		{
			name:        "SuccessfulExecution",
			result:      createSuccessfulExecutionResult(),
			expected:    false,
			description: "Successful execution should fail assertion",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := AssertExecutionTimedOutE(tc.result)
			
			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

// Unit tests for output assertions

func TestAssertExecutionOutputE(t *testing.T) {
	result := createSuccessfulExecutionResult()
	
	tests := []struct {
		name           string
		expectedOutput string
		expected       bool
		description    string
	}{
		{
			name:           "ContainsExpectedText",
			expectedOutput: "success",
			expected:       true,
			description:    "Output containing expected text should pass assertion",
		},
		{
			name:           "ContainsPartialJSON",
			expectedOutput: "Hello, World!",
			expected:       true,
			description:    "Output containing partial JSON should pass assertion",
		},
		{
			name:           "DoesNotContainText",
			expectedOutput: "failure",
			expected:       false,
			description:    "Output not containing expected text should fail assertion",
		},
		{
			name:           "EmptyExpectedOutput",
			expectedOutput: "",
			expected:       true,
			description:    "Empty expected output should always pass",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			actualResult := AssertExecutionOutputE(result, tc.expectedOutput)
			
			// Assert
			assert.Equal(t, tc.expected, actualResult, tc.description)
		})
	}
}

func TestAssertExecutionOutputJSONE(t *testing.T) {
	result := createSuccessfulExecutionResult()
	
	tests := []struct {
		name         string
		expectedJSON interface{}
		expected     bool
		description  string
	}{
		{
			name: "MatchingJSONStructure",
			expectedJSON: map[string]interface{}{
				"result":  "success",
				"message": "Hello, World!",
			},
			expected:    true,
			description: "Matching JSON structure should pass assertion",
		},
		{
			name: "PartialJSONMatch",
			expectedJSON: map[string]interface{}{
				"result": "success",
			},
			expected:    false,
			description: "Partial JSON match should fail assertion (requires exact match)",
		},
		{
			name: "NonMatchingJSON",
			expectedJSON: map[string]interface{}{
				"status": "complete",
			},
			expected:    false,
			description: "Non-matching JSON should fail assertion",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			actualResult := AssertExecutionOutputJSONE(result, tc.expectedJSON)
			
			// Assert
			assert.Equal(t, tc.expected, actualResult, tc.description)
		})
	}
}

func TestAssertExecutionTimeE(t *testing.T) {
	result := createSuccessfulExecutionResult()
	
	tests := []struct {
		name        string
		minTime     time.Duration
		maxTime     time.Duration
		expected    bool
		description string
	}{
		{
			name:        "WithinRange",
			minTime:     1 * time.Second,
			maxTime:     10 * time.Second,
			expected:    true,
			description: "Execution time within range should pass assertion",
		},
		{
			name:        "BelowMinTime",
			minTime:     10 * time.Second,
			maxTime:     20 * time.Second,
			expected:    false,
			description: "Execution time below minimum should fail assertion",
		},
		{
			name:        "AboveMaxTime",
			minTime:     1 * time.Second,
			maxTime:     3 * time.Second,
			expected:    false,
			description: "Execution time above maximum should fail assertion",
		},
		{
			name:        "ExactMatch",
			minTime:     5 * time.Second,
			maxTime:     5 * time.Second,
			expected:    true,
			description: "Exact execution time match should pass assertion",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			actualResult := AssertExecutionTimeE(result, tc.minTime, tc.maxTime)
			
			// Assert
			assert.Equal(t, tc.expected, actualResult, tc.description)
		})
	}
}

func TestAssertExecutionErrorE(t *testing.T) {
	result := createFailedExecutionResult()
	
	tests := []struct {
		name          string
		expectedError string
		expected      bool
		description   string
	}{
		{
			name:          "ContainsExpectedError",
			expectedError: "States.TaskFailed",
			expected:      true,
			description:   "Error containing expected text should pass assertion",
		},
		{
			name:          "ContainsPartialError",
			expectedError: "TaskFailed",
			expected:      true,
			description:   "Error containing partial text should pass assertion",
		},
		{
			name:          "DoesNotContainError",
			expectedError: "States.Timeout",
			expected:      false,
			description:   "Error not containing expected text should fail assertion",
		},
		{
			name:          "EmptyExpectedError",
			expectedError: "",
			expected:      true,
			description:   "Empty expected error should always pass",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			actualResult := AssertExecutionErrorE(result, tc.expectedError)
			
			// Assert
			assert.Equal(t, tc.expected, actualResult, tc.description)
		})
	}
}

// Unit tests for execution pattern assertions

func TestAssertExecutionPatternE(t *testing.T) {
	successResult := createSuccessfulExecutionResult()
	failedResult := createFailedExecutionResult()
	
	tests := []struct {
		name        string
		result      *ExecutionResult
		pattern     *ExecutionPattern
		expected    bool
		description string
	}{
		{
			name:   "MatchingPattern",
			result: successResult,
			pattern: &ExecutionPattern{
				Status:         &[]types.ExecutionStatus{types.ExecutionStatusSucceeded}[0],
				OutputContains: "success",
				MinExecutionTime: &[]time.Duration{1 * time.Second}[0],
				MaxExecutionTime: &[]time.Duration{10 * time.Second}[0],
			},
			expected:    true,
			description: "Matching pattern should pass assertion",
		},
		{
			name:   "NonMatchingStatus",
			result: successResult,
			pattern: &ExecutionPattern{
				Status: &[]types.ExecutionStatus{types.ExecutionStatusFailed}[0],
			},
			expected:    false,
			description: "Non-matching status should fail assertion",
		},
		{
			name:   "NonMatchingOutput",
			result: successResult,
			pattern: &ExecutionPattern{
				OutputContains: "failure",
			},
			expected:    false,
			description: "Non-matching output should fail assertion",
		},
		{
			name:   "FailedExecutionPattern",
			result: failedResult,
			pattern: &ExecutionPattern{
				Status:        &[]types.ExecutionStatus{types.ExecutionStatusFailed}[0],
				ErrorContains: "TaskFailed",
			},
			expected:    true,
			description: "Failed execution pattern should match",
		},
		{
			name:        "NilPattern",
			result:      successResult,
			pattern:     nil,
			expected:    false,
			description: "Nil pattern should fail assertion",
		},
		{
			name:        "NilResult",
			result:      nil,
			pattern:     createTestExecutionPattern(),
			expected:    false,
			description: "Nil result should fail assertion",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			actualResult := AssertExecutionPatternE(tc.result, tc.pattern)
			
			// Assert
			assert.Equal(t, tc.expected, actualResult, tc.description)
		})
	}
}

// Tests for non-E variants (panic on error assertion functions)

func TestAssertExecutionSucceeded(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		shouldPanic bool
		description string
	}{
		{
			name:        "SuccessfulExecutionShouldNotPanic",
			result:      createSuccessfulExecutionResult(),
			shouldPanic: false,
			description: "Successful execution should not cause panic",
		},
		{
			name:        "FailedExecutionShouldPanic",
			result:      createFailedExecutionResult(),
			shouldPanic: true,
			description: "Failed execution should cause panic",
		},
		{
			name:        "NilResultShouldPanic",
			result:      nil,
			shouldPanic: true,
			description: "Nil result should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionSucceeded to panic for %s, but it did not", tc.description)
					}
				}()
				
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionSucceeded(mockT, tc.result)
			} else {
				// For successful cases, test that no panic occurs
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionSucceeded(mockT, tc.result)
			}
		})
	}
}

func TestAssertExecutionFailed(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		shouldPanic bool
		description string
	}{
		{
			name:        "FailedExecutionShouldNotPanic",
			result:      createFailedExecutionResult(),
			shouldPanic: false,
			description: "Failed execution should not cause panic",
		},
		{
			name:        "SuccessfulExecutionShouldPanic",
			result:      createSuccessfulExecutionResult(),
			shouldPanic: true,
			description: "Successful execution should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionFailed to panic for %s, but it did not", tc.description)
					}
				}()
				
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionFailed(mockT, tc.result)
			} else {
				// For expected failure cases, test that no panic occurs
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionFailed(mockT, tc.result)
			}
		})
	}
}

func TestAssertExecutionTimedOut(t *testing.T) {
	timedOutResult := &ExecutionResult{
		ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
		Name:            "test-execution",
		Status:          types.ExecutionStatusTimedOut,
		StartDate:       time.Now().Add(-10 * time.Minute),
		StopDate:        time.Now(),
		Input:           `{"message": "Hello, World!"}`,
		ExecutionTime:   10 * time.Minute,
	}

	tests := []struct {
		name        string
		result      *ExecutionResult
		shouldPanic bool
		description string
	}{
		{
			name:        "TimedOutExecutionShouldNotPanic",
			result:      timedOutResult,
			shouldPanic: false,
			description: "Timed out execution should not cause panic",
		},
		{
			name:        "SuccessfulExecutionShouldPanic",
			result:      createSuccessfulExecutionResult(),
			shouldPanic: true,
			description: "Successful execution should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionTimedOut to panic for %s, but it did not", tc.description)
					}
				}()
				
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionTimedOut(mockT, tc.result)
			} else {
				// For expected timeout cases, test that no panic occurs
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionTimedOut(mockT, tc.result)
			}
		})
	}
}

func TestAssertExecutionAborted(t *testing.T) {
	abortedResult := &ExecutionResult{
		ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
		Name:            "test-execution",
		Status:          types.ExecutionStatusAborted,
		StartDate:       time.Now().Add(-2 * time.Minute),
		StopDate:        time.Now(),
		Input:           `{"message": "Hello, World!"}`,
		ExecutionTime:   1 * time.Minute,
	}

	tests := []struct {
		name        string
		result      *ExecutionResult
		shouldPanic bool
		description string
	}{
		{
			name:        "AbortedExecutionShouldNotPanic",
			result:      abortedResult,
			shouldPanic: false,
			description: "Aborted execution should not cause panic",
		},
		{
			name:        "SuccessfulExecutionShouldPanic",
			result:      createSuccessfulExecutionResult(),
			shouldPanic: true,
			description: "Successful execution should cause panic when expecting aborted",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionAborted to panic for %s, but it did not", tc.description)
					}
				}()
				
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionAborted(mockT, tc.result)
			} else {
				// For expected aborted cases, test that no panic occurs
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionAborted(mockT, tc.result)
			}
		})
	}
}

func TestAssertExecutionOutput(t *testing.T) {
	tests := []struct {
		name           string
		result         *ExecutionResult
		expectedOutput string
		shouldPanic    bool
		description    string
	}{
		{
			name:           "MatchingOutputShouldNotPanic",
			result:         createSuccessfulExecutionResult(),
			expectedOutput: "success",
			shouldPanic:    false,
			description:    "Matching output should not cause panic",
		},
		{
			name:           "NonMatchingOutputShouldPanic",
			result:         createSuccessfulExecutionResult(),
			expectedOutput: "failure",
			shouldPanic:    true,
			description:    "Non-matching output should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionOutput to panic for %s, but it did not", tc.description)
					}
				}()
				
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionOutput(mockT, tc.result, tc.expectedOutput)
			} else {
				// For matching cases, test that no panic occurs
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionOutput(mockT, tc.result, tc.expectedOutput)
			}
		})
	}
}

func TestAssertExecutionOutputJSON(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		expected    map[string]interface{}
		shouldPanic bool
		description string
	}{
		{
			name:        "MatchingJSONShouldNotPanic",
			result:      createSuccessfulExecutionResult(),
			expected:    map[string]interface{}{"result": "success", "message": "Hello, World!"},
			shouldPanic: false,
			description: "Matching JSON should not cause panic",
		},
		{
			name:        "NonMatchingJSONShouldPanic",
			result:      createSuccessfulExecutionResult(),
			expected:    map[string]interface{}{"result": "failure"},
			shouldPanic: true,
			description: "Non-matching JSON should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionOutputJSON to panic for %s, but it did not", tc.description)
					}
				}()
				
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionOutputJSON(mockT, tc.result, tc.expected)
			} else {
				// For matching cases, test that no panic occurs
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionOutputJSON(mockT, tc.result, tc.expected)
			}
		})
	}
}

func TestAssertExecutionTime(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		minTime     time.Duration
		maxTime     time.Duration
		shouldPanic bool
		description string
	}{
		{
			name:        "WithinTimeLimitsShouldNotPanic",
			result:      createSuccessfulExecutionResult(),
			minTime:     1 * time.Second,
			maxTime:     10 * time.Second,
			shouldPanic: false,
			description: "Execution time within limits should not cause panic",
		},
		{
			name:        "ExceedsMaxTimeShouldPanic",
			result:      createSuccessfulExecutionResult(),
			minTime:     1 * time.Second,
			maxTime:     1 * time.Second, // Less than actual execution time
			shouldPanic: true,
			description: "Execution time exceeding max should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionTime to panic for %s, but it did not", tc.description)
					}
				}()
				
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionTime(mockT, tc.result, tc.minTime, tc.maxTime)
			} else {
				// For valid time cases, test that no panic occurs
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionTime(mockT, tc.result, tc.minTime, tc.maxTime)
			}
		})
	}
}

func TestAssertExecutionError(t *testing.T) {
	tests := []struct {
		name          string
		result        *ExecutionResult
		expectedError string
		shouldPanic   bool
		description   string
	}{
		{
			name:          "MatchingErrorShouldNotPanic",
			result:        createFailedExecutionResult(),
			expectedError: "TaskFailed",
			shouldPanic:   false,
			description:   "Matching error should not cause panic",
		},
		{
			name:          "NonMatchingErrorShouldPanic",
			result:        createFailedExecutionResult(),
			expectedError: "UnknownError",
			shouldPanic:   true,
			description:   "Non-matching error should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionError to panic for %s, but it did not", tc.description)
					}
				}()
				
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionError(mockT, tc.result, tc.expectedError)
			} else {
				// For matching error cases, test that no panic occurs
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionError(mockT, tc.result, tc.expectedError)
			}
		})
	}
}

func TestAssertExecutionPattern(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		pattern     *ExecutionPattern
		shouldPanic bool
		description string
	}{
		{
			name:        "MatchingPatternShouldNotPanic",
			result:      createSuccessfulExecutionResult(),
			pattern:     createTestExecutionPattern(),
			shouldPanic: false,
			description: "Matching pattern should not cause panic",
		},
		{
			name:   "NonMatchingPatternShouldPanic",
			result: createSuccessfulExecutionResult(),
			pattern: &ExecutionPattern{
				Status: &[]types.ExecutionStatus{types.ExecutionStatusFailed}[0],
			},
			shouldPanic: true,
			description: "Non-matching pattern should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected AssertExecutionPattern to panic for %s, but it did not", tc.description)
					}
				}()
				
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionPattern(mockT, tc.result, tc.pattern)
			} else {
				// For matching pattern cases, test that no panic occurs
				mockT := &MockT{errorMessages: make([]string, 0)}
				AssertExecutionPattern(mockT, tc.result, tc.pattern)
			}
		})
	}
}

// Benchmark tests for assertion performance

func BenchmarkAssertExecutionSucceededE(b *testing.B) {
	result := createSuccessfulExecutionResult()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AssertExecutionSucceededE(result)
	}
}

func BenchmarkAssertExecutionOutputE(b *testing.B) {
	result := createSuccessfulExecutionResult()
	expectedOutput := "success"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AssertExecutionOutputE(result, expectedOutput)
	}
}

func BenchmarkAssertExecutionPatternE(b *testing.B) {
	result := createSuccessfulExecutionResult()
	pattern := createTestExecutionPattern()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AssertExecutionPatternE(result, pattern)
	}
}