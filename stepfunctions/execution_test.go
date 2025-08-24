package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

// Test data factories for execution operations

func createTestExecutionOptions() *ExecutionOptions {
	return &ExecutionOptions{
		TraceHeader: "test-trace-header",
		Timeout:     10 * time.Minute,
		MaxRetries:  5,
		RetryDelay:  2 * time.Second,
	}
}

func createTestWaitOptions() *WaitOptions {
	return &WaitOptions{
		Timeout:         5 * time.Minute,
		PollingInterval: 10 * time.Second,
		MaxAttempts:     30,
	}
}

func createTestStartExecutionRequest() *StartExecutionRequest {
	return &StartExecutionRequest{
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
		Name:            "test-execution",
		Input:           `{"message": "Hello, World!"}`,
		TraceHeader:     "test-trace-header",
	}
}

// Table-driven test structures

type startExecutionTestCase struct {
	name                 string
	request              *StartExecutionRequest
	shouldError          bool
	expectedErrorMessage string
	description          string
}

type waitForExecutionTestCase struct {
	name                 string
	executionArn         string
	waitOptions          *WaitOptions
	shouldError          bool
	expectedErrorMessage string
	description          string
}

type stopExecutionTestCase struct {
	name                 string
	executionArn         string
	error                string
	cause                string
	shouldError          bool
	expectedErrorMessage string
	description          string
}

// Unit tests for execution operations

func TestStartExecutionE(t *testing.T) {
	tests := []startExecutionTestCase{
		{
			name:        "ValidExecutionStart",
			request:     createTestStartExecutionRequest(),
			shouldError: false,
			description: "Valid execution request should pass validation",
		},
		{
			name: "MissingStateMachineArn",
			request: &StartExecutionRequest{
				StateMachineArn: "",
				Name:            "test-execution",
				Input:           `{"message": "Hello"}`,
			},
			shouldError:          true,
			expectedErrorMessage: "state machine ARN is required",
			description:          "Missing state machine ARN should fail",
		},
		{
			name: "InvalidExecutionName",
			request: &StartExecutionRequest{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
				Name:            "",
				Input:           `{"message": "Hello"}`,
			},
			shouldError:          true,
			expectedErrorMessage: "execution name is required",
			description:          "Empty execution name should fail",
		},
		{
			name: "InvalidInputJSON",
			request: &StartExecutionRequest{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
				Name:            "test-execution",
				Input:           `{"invalid": json}`,
			},
			shouldError:          true,
			expectedErrorMessage: "invalid JSON",
			description:          "Invalid JSON input should fail",
		},
		{
			name: "EmptyInputAllowed",
			request: &StartExecutionRequest{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
				Name:            "test-execution",
				Input:           "",
			},
			shouldError: false,
			description: "Empty input should be allowed",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStartExecutionRequest(tc.request)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
				if tc.expectedErrorMessage != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorMessage, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestStopExecutionE(t *testing.T) {
	tests := []stopExecutionTestCase{
		{
			name:         "ValidExecutionStop",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			error:        "User requested stop",
			cause:        "Testing stop functionality",
			shouldError:  false,
			description:  "Valid execution stop should pass validation",
		},
		{
			name:                 "EmptyExecutionArn",
			executionArn:         "",
			error:                "User requested stop",
			cause:                "Testing stop functionality",
			shouldError:          true,
			expectedErrorMessage: "execution ARN is required",
			description:          "Empty execution ARN should fail",
		},
		{
			name:                 "InvalidExecutionArn",
			executionArn:         "invalid-arn",
			error:                "User requested stop",
			cause:                "Testing stop functionality",
			shouldError:          true,
			expectedErrorMessage: "invalid ARN",
			description:          "Invalid execution ARN should fail",
		},
		{
			name:         "EmptyErrorMessage",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			error:        "",
			cause:        "Testing stop functionality",
			shouldError:  false,
			description:  "Empty error message should be allowed",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStopExecutionRequest(tc.executionArn, tc.error, tc.cause)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
				if tc.expectedErrorMessage != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorMessage, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestWaitForExecutionE(t *testing.T) {
	tests := []waitForExecutionTestCase{
		{
			name:         "ValidWaitRequest",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			waitOptions:  createTestWaitOptions(),
			shouldError:  false,
			description:  "Valid wait request should pass validation",
		},
		{
			name:                 "EmptyExecutionArn",
			executionArn:         "",
			waitOptions:          createTestWaitOptions(),
			shouldError:          true,
			expectedErrorMessage: "execution ARN is required",
			description:          "Empty execution ARN should fail",
		},
		{
			name:                 "InvalidExecutionArn",
			executionArn:         "invalid-arn",
			waitOptions:          createTestWaitOptions(),
			shouldError:          true,
			expectedErrorMessage: "invalid ARN",
			description:          "Invalid execution ARN should fail",
		},
		{
			name:         "NilWaitOptions",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			waitOptions:  nil,
			shouldError:  false,
			description:  "Nil wait options should use defaults",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateWaitForExecutionRequest(tc.executionArn, tc.waitOptions)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
				if tc.expectedErrorMessage != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorMessage, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestDescribeExecutionE(t *testing.T) {
	tests := []struct {
		name                 string
		executionArn         string
		shouldError          bool
		expectedErrorMessage string
		description          string
	}{
		{
			name:         "ValidExecutionArn",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			shouldError:  false,
			description:  "Valid execution ARN should pass validation",
		},
		{
			name:                 "EmptyExecutionArn",
			executionArn:         "",
			shouldError:          true,
			expectedErrorMessage: "execution ARN is required",
			description:          "Empty execution ARN should fail",
		},
		{
			name:                 "InvalidExecutionArn",
			executionArn:         "not-an-arn",
			shouldError:          true,
			expectedErrorMessage: "invalid ARN",
			description:          "Invalid execution ARN should fail",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateExecutionArn(tc.executionArn)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
				if tc.expectedErrorMessage != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorMessage, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestListExecutionsE(t *testing.T) {
	tests := []struct {
		name                 string
		stateMachineArn      string
		statusFilter         types.ExecutionStatus
		maxResults           int32
		shouldError          bool
		expectedErrorMessage string
		description          string
	}{
		{
			name:            "ValidListRequest",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			statusFilter:    types.ExecutionStatusRunning,
			maxResults:      100,
			shouldError:     false,
			description:     "Valid list request should pass validation",
		},
		{
			name:                 "EmptyStateMachineArn",
			stateMachineArn:      "",
			statusFilter:         types.ExecutionStatusRunning,
			maxResults:           100,
			shouldError:          true,
			expectedErrorMessage: "state machine ARN is required",
			description:          "Empty state machine ARN should fail",
		},
		{
			name:                 "InvalidMaxResults",
			stateMachineArn:      "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			statusFilter:         types.ExecutionStatusRunning,
			maxResults:           -1,
			shouldError:          true,
			expectedErrorMessage: "max results cannot be negative",
			description:          "Negative max results should fail",
		},
		{
			name:            "DefaultStatusFilter",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			statusFilter:    "",
			maxResults:      100,
			shouldError:     false,
			description:     "Empty status filter should be allowed",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateListExecutionsRequest(tc.stateMachineArn, tc.statusFilter, tc.maxResults)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
				if tc.expectedErrorMessage != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorMessage, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// Tests for execution result processing

func TestProcessExecutionResult(t *testing.T) {
	// Arrange
	result := createTestExecutionResult(types.ExecutionStatusSucceeded, `{"result": "success"}`)
	
	// Act
	processed := processExecutionResult(result)
	
	// Assert
	assert.NotNil(t, processed, "Processed result should not be nil")
	assert.Equal(t, result.ExecutionArn, processed.ExecutionArn, "ARN should be preserved")
	assert.Equal(t, result.Status, processed.Status, "Status should be preserved")
	assert.Equal(t, result.Output, processed.Output, "Output should be preserved")
}

func TestCalculateExecutionTime(t *testing.T) {
	// Arrange
	startTime := time.Now()
	stopTime := startTime.Add(5 * time.Minute)
	
	// Act
	duration := calculateExecutionTime(startTime, stopTime)
	
	// Assert
	assert.Equal(t, 5*time.Minute, duration, "Should calculate correct execution time")
}

func TestIsExecutionComplete(t *testing.T) {
	tests := []struct {
		name        string
		status      types.ExecutionStatus
		expected    bool
		description string
	}{
		{
			name:        "RunningExecution",
			status:      types.ExecutionStatusRunning,
			expected:    false,
			description: "Running execution should not be complete",
		},
		{
			name:        "SucceededExecution",
			status:      types.ExecutionStatusSucceeded,
			expected:    true,
			description: "Succeeded execution should be complete",
		},
		{
			name:        "FailedExecution",
			status:      types.ExecutionStatusFailed,
			expected:    true,
			description: "Failed execution should be complete",
		},
		{
			name:        "TimedOutExecution",
			status:      types.ExecutionStatusTimedOut,
			expected:    true,
			description: "Timed out execution should be complete",
		},
		{
			name:        "AbortedExecution",
			status:      types.ExecutionStatusAborted,
			expected:    true,
			description: "Aborted execution should be complete",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isExecutionComplete(tc.status)
			
			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestParseExecutionArn(t *testing.T) {
	tests := []struct {
		name                   string
		executionArn           string
		expectedStateMachine   string
		expectedExecutionName  string
		shouldError            bool
		description            string
	}{
		{
			name:                  "ValidExecutionArn",
			executionArn:          "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			expectedStateMachine:  "test-state-machine",
			expectedExecutionName: "test-execution",
			shouldError:           false,
			description:           "Valid execution ARN should parse correctly",
		},
		{
			name:         "InvalidExecutionArn",
			executionArn: "invalid-arn",
			shouldError:  true,
			description:  "Invalid execution ARN should fail parsing",
		},
		{
			name:         "EmptyExecutionArn",
			executionArn: "",
			shouldError:  true,
			description:  "Empty execution ARN should fail parsing",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			stateMachine, executionName, err := parseExecutionArn(tc.executionArn)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				assert.Equal(t, tc.expectedStateMachine, stateMachine, "State machine name should match")
				assert.Equal(t, tc.expectedExecutionName, executionName, "Execution name should match")
			}
		})
	}
}

// Benchmark tests for execution operations

func BenchmarkValidateStartExecutionRequest(b *testing.B) {
	request := createTestStartExecutionRequest()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateStartExecutionRequest(request)
	}
}

func BenchmarkIsExecutionComplete(b *testing.B) {
	status := types.ExecutionStatusSucceeded
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isExecutionComplete(status)
	}
}

func BenchmarkParseExecutionArn(b *testing.B) {
	executionArn := "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parseExecutionArn(executionArn)
	}
}