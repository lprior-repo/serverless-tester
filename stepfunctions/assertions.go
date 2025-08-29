// Package stepfunctions assertions - Comprehensive validation functions for Step Functions testing
// Following strict Terratest patterns with detailed assertions

package stepfunctions

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

// Execution Assertions

// AssertExecutionSucceeded asserts that an execution completed successfully
// This is a Terratest-style function that fails the test on assertion failure
func AssertExecutionSucceeded(t TestingT, result *ExecutionResult) {
	if !assertExecutionSucceededE(result) {
		t.Errorf("Expected execution to succeed but got status: %s", result.Status)
		t.FailNow()
	}
}

// AssertExecutionSucceededE asserts that an execution completed successfully with boolean return
func AssertExecutionSucceededE(result *ExecutionResult) bool {
	return assertExecutionSucceededE(result)
}

// assertExecutionSucceededE is the internal implementation
func assertExecutionSucceededE(result *ExecutionResult) bool {
	if result == nil {
		return false
	}
	return result.Status == types.ExecutionStatusSucceeded
}

// AssertExecutionFailed asserts that an execution failed
func AssertExecutionFailed(t TestingT, result *ExecutionResult) {
	if !assertExecutionFailedE(result) {
		t.Errorf("Expected execution to fail but got status: %s", result.Status)
		t.FailNow()
	}
}

// AssertExecutionFailedE asserts that an execution failed with boolean return
func AssertExecutionFailedE(result *ExecutionResult) bool {
	return assertExecutionFailedE(result)
}

// assertExecutionFailedE is the internal implementation
func assertExecutionFailedE(result *ExecutionResult) bool {
	if result == nil {
		return false
	}
	return result.Status == types.ExecutionStatusFailed ||
		result.Status == types.ExecutionStatusTimedOut ||
		result.Status == types.ExecutionStatusAborted
}

// AssertExecutionTimedOut asserts that an execution timed out
func AssertExecutionTimedOut(t TestingT, result *ExecutionResult) {
	if !assertExecutionTimedOutE(result) {
		t.Errorf("Expected execution to timeout but got status: %s", result.Status)
		t.FailNow()
	}
}

// AssertExecutionTimedOutE asserts that an execution timed out with boolean return
func AssertExecutionTimedOutE(result *ExecutionResult) bool {
	return assertExecutionTimedOutE(result)
}

// assertExecutionTimedOutE is the internal implementation
func assertExecutionTimedOutE(result *ExecutionResult) bool {
	if result == nil {
		return false
	}
	return result.Status == types.ExecutionStatusTimedOut
}

// AssertExecutionAborted asserts that an execution was aborted
func AssertExecutionAborted(t TestingT, result *ExecutionResult) {
	if !assertExecutionAbortedE(result) {
		t.Errorf("Expected execution to be aborted but got status: %s", result.Status)
		t.FailNow()
	}
}

// AssertExecutionAbortedE asserts that an execution was aborted with boolean return
func AssertExecutionAbortedE(result *ExecutionResult) bool {
	return assertExecutionAbortedE(result)
}

// assertExecutionAbortedE is the internal implementation
func assertExecutionAbortedE(result *ExecutionResult) bool {
	if result == nil {
		return false
	}
	return result.Status == types.ExecutionStatusAborted
}

// AssertExecutionOutput asserts that execution output contains expected content
func AssertExecutionOutput(t TestingT, result *ExecutionResult, expectedOutput string) {
	if !assertExecutionOutputE(result, expectedOutput) {
		t.Errorf("Expected execution output to contain '%s' but got: %s", expectedOutput, result.Output)
		t.FailNow()
	}
}

// AssertExecutionOutputE asserts that execution output contains expected content with boolean return
func AssertExecutionOutputE(result *ExecutionResult, expectedOutput string) bool {
	return assertExecutionOutputE(result, expectedOutput)
}

// assertExecutionOutputE is the internal implementation
func assertExecutionOutputE(result *ExecutionResult, expectedOutput string) bool {
	if result == nil {
		return false
	}
	return strings.Contains(result.Output, expectedOutput)
}

// AssertExecutionOutputJSON asserts that execution output matches expected JSON structure
func AssertExecutionOutputJSON(t TestingT, result *ExecutionResult, expectedJSON interface{}) {
	if !assertExecutionOutputJSONE(result, expectedJSON) {
		t.Errorf("Expected execution output to match JSON structure but validation failed")
		t.FailNow()
	}
}

// AssertExecutionOutputJSONE asserts that execution output matches expected JSON structure with boolean return
func AssertExecutionOutputJSONE(result *ExecutionResult, expectedJSON interface{}) bool {
	return assertExecutionOutputJSONE(result, expectedJSON)
}

// assertExecutionOutputJSONE is the internal implementation
func assertExecutionOutputJSONE(result *ExecutionResult, expectedJSON interface{}) bool {
	if result == nil || result.Output == "" {
		return false
	}
	
	// Parse actual output
	var actualJSON interface{}
	if err := json.Unmarshal([]byte(result.Output), &actualJSON); err != nil {
		return false
	}
	
	// Marshal expected JSON to compare structure
	expectedBytes, err := json.Marshal(expectedJSON)
	if err != nil {
		return false
	}
	
	actualBytes, err := json.Marshal(actualJSON)
	if err != nil {
		return false
	}
	
	return string(expectedBytes) == string(actualBytes)
}

// AssertExecutionTime asserts that execution time is within expected bounds
func AssertExecutionTime(t TestingT, result *ExecutionResult, minTime, maxTime time.Duration) {
	if !assertExecutionTimeE(result, minTime, maxTime) {
		t.Errorf("Expected execution time to be between %v and %v but got: %v", 
			minTime, maxTime, result.ExecutionTime)
		t.FailNow()
	}
}

// AssertExecutionTimeE asserts that execution time is within expected bounds with boolean return
func AssertExecutionTimeE(result *ExecutionResult, minTime, maxTime time.Duration) bool {
	return assertExecutionTimeE(result, minTime, maxTime)
}

// assertExecutionTimeE is the internal implementation
func assertExecutionTimeE(result *ExecutionResult, minTime, maxTime time.Duration) bool {
	if result == nil {
		return false
	}
	return result.ExecutionTime >= minTime && result.ExecutionTime <= maxTime
}

// AssertExecutionError asserts that execution failed with specific error
func AssertExecutionError(t TestingT, result *ExecutionResult, expectedError string) {
	if !assertExecutionErrorE(result, expectedError) {
		t.Errorf("Expected execution error to contain '%s' but got: %s", expectedError, result.Error)
		t.FailNow()
	}
}

// AssertExecutionErrorE asserts that execution failed with specific error with boolean return
func AssertExecutionErrorE(result *ExecutionResult, expectedError string) bool {
	return assertExecutionErrorE(result, expectedError)
}

// assertExecutionErrorE is the internal implementation
func assertExecutionErrorE(result *ExecutionResult, expectedError string) bool {
	if result == nil {
		return false
	}
	return strings.Contains(result.Error, expectedError)
}

// State Machine Assertions

// AssertStateMachineExists asserts that a state machine exists and is active
func AssertStateMachineExists(t TestingT, ctx *TestContext, stateMachineArn string) {
	if !assertStateMachineExistsE(ctx, stateMachineArn) {
		t.Errorf("Expected state machine to exist: %s", stateMachineArn)
		t.FailNow()
	}
}

// AssertStateMachineExistsE asserts that a state machine exists and is active with boolean return
func AssertStateMachineExistsE(ctx *TestContext, stateMachineArn string) bool {
	return assertStateMachineExistsE(ctx, stateMachineArn)
}

// assertStateMachineExistsE is the internal implementation
func assertStateMachineExistsE(ctx *TestContext, stateMachineArn string) bool {
	info, err := DescribeStateMachineE(ctx, stateMachineArn)
	if err != nil {
		return false
	}
	return info.Status == types.StateMachineStatusActive
}

// AssertStateMachineStatus asserts that a state machine has expected status
func AssertStateMachineStatus(t TestingT, ctx *TestContext, stateMachineArn string, expectedStatus types.StateMachineStatus) {
	if !assertStateMachineStatusE(ctx, stateMachineArn, expectedStatus) {
		t.Errorf("Expected state machine status to be %s", expectedStatus)
		t.FailNow()
	}
}

// AssertStateMachineStatusE asserts that a state machine has expected status with boolean return
func AssertStateMachineStatusE(ctx *TestContext, stateMachineArn string, expectedStatus types.StateMachineStatus) bool {
	return assertStateMachineStatusE(ctx, stateMachineArn, expectedStatus)
}

// assertStateMachineStatusE is the internal implementation
func assertStateMachineStatusE(ctx *TestContext, stateMachineArn string, expectedStatus types.StateMachineStatus) bool {
	info, err := DescribeStateMachineE(ctx, stateMachineArn)
	if err != nil {
		return false
	}
	return info.Status == expectedStatus
}

// AssertStateMachineType asserts that a state machine is of expected type
func AssertStateMachineType(t TestingT, ctx *TestContext, stateMachineArn string, expectedType types.StateMachineType) {
	if !assertStateMachineTypeE(ctx, stateMachineArn, expectedType) {
		t.Errorf("Expected state machine type to be %s", expectedType)
		t.FailNow()
	}
}

// AssertStateMachineTypeE asserts that a state machine is of expected type with boolean return
func AssertStateMachineTypeE(ctx *TestContext, stateMachineArn string, expectedType types.StateMachineType) bool {
	return assertStateMachineTypeE(ctx, stateMachineArn, expectedType)
}

// assertStateMachineTypeE is the internal implementation
func assertStateMachineTypeE(ctx *TestContext, stateMachineArn string, expectedType types.StateMachineType) bool {
	info, err := DescribeStateMachineE(ctx, stateMachineArn)
	if err != nil {
		return false
	}
	return info.Type == expectedType
}

// Execution Count Assertions

// AssertExecutionCount asserts that a state machine has expected number of executions
func AssertExecutionCount(t TestingT, ctx *TestContext, stateMachineArn string, expectedCount int) {
	if !assertExecutionCountE(ctx, stateMachineArn, expectedCount) {
		t.Errorf("Expected execution count to be %d", expectedCount)
		t.FailNow()
	}
}

// AssertExecutionCountE asserts that a state machine has expected number of executions with boolean return
func AssertExecutionCountE(ctx *TestContext, stateMachineArn string, expectedCount int) bool {
	return assertExecutionCountE(ctx, stateMachineArn, expectedCount)
}

// assertExecutionCountE is the internal implementation
func assertExecutionCountE(ctx *TestContext, stateMachineArn string, expectedCount int) bool {
	executions, err := ListExecutionsE(ctx, stateMachineArn, "", 1000)
	if err != nil {
		return false
	}
	return len(executions) == expectedCount
}

// AssertExecutionCountByStatus asserts execution count for specific status
func AssertExecutionCountByStatus(t TestingT, ctx *TestContext, stateMachineArn string, status types.ExecutionStatus, expectedCount int) {
	if !assertExecutionCountByStatusE(ctx, stateMachineArn, status, expectedCount) {
		t.Errorf("Expected execution count with status %s to be %d", status, expectedCount)
		t.FailNow()
	}
}

// AssertExecutionCountByStatusE asserts execution count for specific status with boolean return
func AssertExecutionCountByStatusE(ctx *TestContext, stateMachineArn string, status types.ExecutionStatus, expectedCount int) bool {
	return assertExecutionCountByStatusE(ctx, stateMachineArn, status, expectedCount)
}

// assertExecutionCountByStatusE is the internal implementation
func assertExecutionCountByStatusE(ctx *TestContext, stateMachineArn string, status types.ExecutionStatus, expectedCount int) bool {
	executions, err := ListExecutionsE(ctx, stateMachineArn, status, 1000)
	if err != nil {
		return false
	}
	return len(executions) == expectedCount
}

// Combined Assertions

// AssertExecutionPattern asserts multiple execution properties at once
func AssertExecutionPattern(t TestingT, result *ExecutionResult, pattern *ExecutionPattern) {
	if !assertExecutionPatternE(result, pattern) {
		t.Errorf("Execution did not match expected pattern")
		t.FailNow()
	}
}

// AssertExecutionPatternE asserts multiple execution properties at once with boolean return
func AssertExecutionPatternE(result *ExecutionResult, pattern *ExecutionPattern) bool {
	return assertExecutionPatternE(result, pattern)
}

// ExecutionPattern defines expected execution characteristics
type ExecutionPattern struct {
	Status          *types.ExecutionStatus
	OutputContains  string
	OutputJSON      interface{}
	ErrorContains   string
	MinExecutionTime *time.Duration
	MaxExecutionTime *time.Duration
}

// assertExecutionPatternE is the internal implementation
func assertExecutionPatternE(result *ExecutionResult, pattern *ExecutionPattern) bool {
	if result == nil || pattern == nil {
		return false
	}
	
	// Check status if specified
	if pattern.Status != nil && result.Status != *pattern.Status {
		return false
	}
	
	// Check output contains if specified
	if pattern.OutputContains != "" && !strings.Contains(result.Output, pattern.OutputContains) {
		return false
	}
	
	// Check JSON output if specified
	if pattern.OutputJSON != nil && !assertExecutionOutputJSONE(result, pattern.OutputJSON) {
		return false
	}
	
	// Check error contains if specified
	if pattern.ErrorContains != "" && !strings.Contains(result.Error, pattern.ErrorContains) {
		return false
	}
	
	// Check execution time bounds if specified
	if pattern.MinExecutionTime != nil && result.ExecutionTime < *pattern.MinExecutionTime {
		return false
	}
	
	if pattern.MaxExecutionTime != nil && result.ExecutionTime > *pattern.MaxExecutionTime {
		return false
	}
	
	return true
}

// History Event Assertions

// AssertHistoryEventExists asserts that a specific event type exists in execution history
func AssertHistoryEventExists(t TestingT, ctx *TestContext, executionArn string, eventType types.HistoryEventType) {
	if !assertHistoryEventExistsE(ctx, executionArn, eventType) {
		t.Errorf("Expected history event type %s to exist", eventType)
		t.FailNow()
	}
}

// AssertHistoryEventExistsE asserts that a specific event type exists in execution history with boolean return
func AssertHistoryEventExistsE(ctx *TestContext, executionArn string, eventType types.HistoryEventType) bool {
	return assertHistoryEventExistsE(ctx, executionArn, eventType)
}

// assertHistoryEventExistsE is the internal implementation
func assertHistoryEventExistsE(ctx *TestContext, executionArn string, eventType types.HistoryEventType) bool {
	history, err := GetExecutionHistoryE(ctx, executionArn)
	if err != nil {
		return false
	}
	
	for _, event := range history {
		if event.Type == eventType {
			return true
		}
	}
	
	return false
}

// AssertHistoryEventCount asserts the count of specific event types
func AssertHistoryEventCount(t TestingT, ctx *TestContext, executionArn string, eventType types.HistoryEventType, expectedCount int) {
	if !assertHistoryEventCountE(ctx, executionArn, eventType, expectedCount) {
		t.Errorf("Expected %d history events of type %s", expectedCount, eventType)
		t.FailNow()
	}
}

// AssertHistoryEventCountE asserts the count of specific event types with boolean return
func AssertHistoryEventCountE(ctx *TestContext, executionArn string, eventType types.HistoryEventType, expectedCount int) bool {
	return assertHistoryEventCountE(ctx, executionArn, eventType, expectedCount)
}

// assertHistoryEventCountE is the internal implementation
func assertHistoryEventCountE(ctx *TestContext, executionArn string, eventType types.HistoryEventType, expectedCount int) bool {
	history, err := GetExecutionHistoryE(ctx, executionArn)
	if err != nil {
		return false
	}
	
	count := 0
	for _, event := range history {
		if event.Type == eventType {
			count++
		}
	}
	
	return count == expectedCount
}

// Utility Functions for Assertions

// ListExecutionsE is implemented in operations.go

// GetExecutionHistory gets execution history (simplified version for assertions)
// Note: Implementation moved to history.go