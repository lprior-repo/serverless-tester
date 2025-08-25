package stepfunctions

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

func TestAssertExecutionOutputAdvanced(t *testing.T) {
	tests := []struct {
		name          string
		result        *ExecutionResult
		expectedOutput string
		shouldPanic   bool
	}{
		{
			name: "MatchingOutput",
			result: &ExecutionResult{
				Output: `{"message": "success", "data": {"processed": true}}`,
			},
			expectedOutput: "success",
			shouldPanic:    false,
		},
		{
			name: "NonMatchingOutput",
			result: &ExecutionResult{
				Output: `{"message": "failure", "data": {"processed": false}}`,
			},
			expectedOutput: "success",
			shouldPanic:    true,
		},
		{
			name: "EmptyExpectedOutput",
			result: &ExecutionResult{
				Output: `{"message": "any output"}`,
			},
			expectedOutput: "",
			shouldPanic:    false, // Empty expected should pass
		},
		{
			name: "NilResult",
			result: nil,
			expectedOutput: "success",
			shouldPanic:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &MockT{}
			ctx := &TestContext{T: mockT}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertExecutionOutput(ctx.T, tt.result, tt.expectedOutput)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertExecutionOutput(ctx.T, tt.result, tt.expectedOutput)
				})
			}
		})
	}
}

func TestAssertExecutionOutputEAdvanced(t *testing.T) {
	tests := []struct {
		name          string
		result        *ExecutionResult
		expectedOutput string
		expectError   bool
	}{
		{
			name: "MatchingPartialJSON",
			result: &ExecutionResult{
				Output: `{"status": "completed", "result": {"count": 42, "items": ["a", "b"]}}`,
			},
			expectedOutput: "completed",
			expectError:    false,
		},
		{
			name: "NonMatchingOutput",
			result: &ExecutionResult{
				Output: `{"status": "failed", "error": "validation error"}`,
			},
			expectedOutput: "completed",
			expectError:    true,
		},
		{
			name: "EmptyExpectedOutput",
			result: &ExecutionResult{
				Output: `{"any": "content"}`,
			},
			expectedOutput: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertExecutionOutputE(tt.result, tt.expectedOutput)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssertExecutionOutputJSONAdvanced(t *testing.T) {
	tests := []struct {
		name           string
		result         *ExecutionResult
		expectedJSON   interface{}
		shouldPanic    bool
	}{
		{
			name: "MatchingJSONStructure",
			result: &ExecutionResult{
				Output: `{"status": "success", "data": {"processed": true, "count": 42}}`,
			},
			expectedJSON: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"processed": true,
					"count":     42,
				},
			},
			shouldPanic: false,
		},
		{
			name: "PartialJSONMatch",
			result: &ExecutionResult{
				Output: `{"status": "success", "data": {"processed": true, "count": 42}, "extra": "ignored"}`,
			},
			expectedJSON: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"processed": true,
				},
			},
			shouldPanic: false,
		},
		{
			name: "NonMatchingJSON",
			result: &ExecutionResult{
				Output: `{"status": "failed", "error": "validation error"}`,
			},
			expectedJSON: map[string]interface{}{
				"status": "success",
			},
			shouldPanic: true,
		},
		{
			name: "InvalidJSONOutput",
			result: &ExecutionResult{
				Output: `invalid json content`,
			},
			expectedJSON: map[string]interface{}{
				"status": "success",
			},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &MockT{}
			ctx := &TestContext{T: mockT}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertExecutionOutputJSON(ctx, tt.result, tt.expectedJSON)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertExecutionOutputJSON(ctx, tt.result, tt.expectedJSON)
				})
			}
		})
	}
}

func TestAssertExecutionOutputJSONEAdvanced(t *testing.T) {
	tests := []struct {
		name         string
		result       *ExecutionResult
		expectedJSON interface{}
		expectError  bool
	}{
		{
			name: "MatchingJSONArray",
			result: &ExecutionResult{
				Output: `[{"id": 1, "name": "test"}, {"id": 2, "name": "test2"}]`,
			},
			expectedJSON: []interface{}{
				map[string]interface{}{"id": float64(1), "name": "test"},
				map[string]interface{}{"id": float64(2), "name": "test2"},
			},
			expectError: false,
		},
		{
			name: "PartialArrayMatch",
			result: &ExecutionResult{
				Output: `[{"id": 1, "name": "test", "extra": "data"}, {"id": 2, "name": "test2"}]`,
			},
			expectedJSON: []interface{}{
				map[string]interface{}{"id": float64(1), "name": "test"},
			},
			expectError: false,
		},
		{
			name: "NonMatchingJSONStructure",
			result: &ExecutionResult{
				Output: `{"different": "structure"}`,
			},
			expectedJSON: map[string]interface{}{
				"expected": "structure",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertExecutionOutputJSONE(tt.result, tt.expectedJSON)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssertExecutionTimeAdvanced(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		minTime     time.Duration
		maxTime     time.Duration
		shouldPanic bool
	}{
		{
			name: "WithinTimeRange",
			result: &ExecutionResult{
				ExecutionTime: 3 * time.Second,
			},
			minTime:     2 * time.Second,
			maxTime:     5 * time.Second,
			shouldPanic: false,
		},
		{
			name: "BelowMinTime",
			result: &ExecutionResult{
				ExecutionTime: 1 * time.Second,
			},
			minTime:     2 * time.Second,
			maxTime:     5 * time.Second,
			shouldPanic: true,
		},
		{
			name: "AboveMaxTime",
			result: &ExecutionResult{
				ExecutionTime: 6 * time.Second,
			},
			minTime:     2 * time.Second,
			maxTime:     5 * time.Second,
			shouldPanic: true,
		},
		{
			name: "ExactMinTime",
			result: &ExecutionResult{
				ExecutionTime: 2 * time.Second,
			},
			minTime:     2 * time.Second,
			maxTime:     5 * time.Second,
			shouldPanic: false,
		},
		{
			name: "ExactMaxTime",
			result: &ExecutionResult{
				ExecutionTime: 5 * time.Second,
			},
			minTime:     2 * time.Second,
			maxTime:     5 * time.Second,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &MockT{}
			ctx := &TestContext{T: mockT}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertExecutionTime(ctx, tt.result, tt.minTime, tt.maxTime)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertExecutionTime(ctx, tt.result, tt.minTime, tt.maxTime)
				})
			}
		})
	}
}

func TestAssertExecutionTimeEAdvanced(t *testing.T) {
	tests := []struct {
		name        string
		result      *ExecutionResult
		minTime     time.Duration
		maxTime     time.Duration
		expectError bool
	}{
		{
			name: "ValidTimeRange",
			result: &ExecutionResult{
				ExecutionTime: 3 * time.Second,
			},
			minTime:     1 * time.Second,
			maxTime:     5 * time.Second,
			expectError: false,
		},
		{
			name: "NilResult",
			result:      nil,
			minTime:     1 * time.Second,
			maxTime:     5 * time.Second,
			expectError: true,
		},
		{
			name: "InvalidTimeRange",
			result: &ExecutionResult{
				ExecutionTime: 10 * time.Second,
			},
			minTime:     1 * time.Second,
			maxTime:     5 * time.Second,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertExecutionTimeE(tt.result, tt.minTime, tt.maxTime)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssertExecutionErrorAdvanced(t *testing.T) {
	tests := []struct {
		name          string
		result        *ExecutionResult
		expectedError string
		shouldPanic   bool
	}{
		{
			name: "ContainsExpectedError",
			result: &ExecutionResult{
				Status: types.ExecutionStatusFailed,
				Error:  "ValidationError: Input validation failed for field 'name'",
			},
			expectedError: "ValidationError",
			shouldPanic:   false,
		},
		{
			name: "DoesNotContainError",
			result: &ExecutionResult{
				Status: types.ExecutionStatusFailed,
				Error:  "TimeoutError: Execution timed out",
			},
			expectedError: "ValidationError",
			shouldPanic:   true,
		},
		{
			name: "EmptyExpectedError",
			result: &ExecutionResult{
				Status: types.ExecutionStatusFailed,
				Error:  "Any error message",
			},
			expectedError: "",
			shouldPanic:   false,
		},
		{
			name: "SucceededExecutionShouldFail",
			result: &ExecutionResult{
				Status: types.ExecutionStatusSucceeded,
			},
			expectedError: "ValidationError",
			shouldPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &MockT{}
			ctx := &TestContext{T: mockT}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertExecutionError(ctx, tt.result, tt.expectedError)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertExecutionError(ctx, tt.result, tt.expectedError)
				})
			}
		})
	}
}

func TestAssertExecutionErrorEAdvanced(t *testing.T) {
	tests := []struct {
		name          string
		result        *ExecutionResult
		expectedError string
		expectError   bool
	}{
		{
			name: "ContainsPartialError",
			result: &ExecutionResult{
				Status: types.ExecutionStatusFailed,
				Error:  "ValidationError: Multiple validation issues found",
			},
			expectedError: "validation",
			expectError:   false,
		},
		{
			name: "CaseInsensitiveMatch",
			result: &ExecutionResult{
				Status: types.ExecutionStatusFailed,
				Error:  "TIMEOUT_ERROR: Request timed out",
			},
			expectedError: "timeout",
			expectError:   false,
		},
		{
			name: "NoMatchingError",
			result: &ExecutionResult{
				Status: types.ExecutionStatusFailed,
				Error:  "NetworkError: Connection failed",
			},
			expectedError: "validation",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertExecutionErrorE(tt.result, tt.expectedError)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssertStateMachineStatusAdvanced(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(*MockStepFunctionsClient)
		stateMachineArn string
		expectedStatus  types.StateMachineStatus
		shouldPanic     bool
	}{
		{
			name: "ActiveStateMachine",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.DescribeStateMachineFunc = func(ctx context.Context, input *sfn.DescribeStateMachineInput) (*sfn.DescribeStateMachineOutput, error) {
					return &sfn.DescribeStateMachineOutput{
						Status: types.StateMachineStatusActive,
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedStatus:  types.StateMachineStatusActive,
			shouldPanic:     false,
		},
		{
			name: "StatusMismatch",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.DescribeStateMachineFunc = func(ctx context.Context, input *sfn.DescribeStateMachineInput) (*sfn.DescribeStateMachineOutput, error) {
					return &sfn.DescribeStateMachineOutput{
						Status: types.StateMachineStatusDeleting,
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedStatus:  types.StateMachineStatusActive,
			shouldPanic:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockStepFunctionsClient{}
			tt.mockSetup(mockClient)

			mockT := &MockT{}
			ctx := &TestContext{
				T:                   mockT,
				StepFunctionsClient: mockClient,
			}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertStateMachineStatus(ctx, tt.stateMachineArn, tt.expectedStatus)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertStateMachineStatus(ctx, tt.stateMachineArn, tt.expectedStatus)
				})
			}
		})
	}
}

func TestAssertStateMachineTypeAdvanced(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(*MockStepFunctionsClient)
		stateMachineArn string
		expectedType    types.StateMachineType
		shouldPanic     bool
	}{
		{
			name: "StandardStateMachine",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.DescribeStateMachineFunc = func(ctx context.Context, input *sfn.DescribeStateMachineInput) (*sfn.DescribeStateMachineOutput, error) {
					return &sfn.DescribeStateMachineOutput{
						Type: types.StateMachineTypeStandard,
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedType:    types.StateMachineTypeStandard,
			shouldPanic:     false,
		},
		{
			name: "ExpressStateMachine",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.DescribeStateMachineFunc = func(ctx context.Context, input *sfn.DescribeStateMachineInput) (*sfn.DescribeStateMachineOutput, error) {
					return &sfn.DescribeStateMachineOutput{
						Type: types.StateMachineTypeExpress,
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedType:    types.StateMachineTypeExpress,
			shouldPanic:     false,
		},
		{
			name: "TypeMismatch",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.DescribeStateMachineFunc = func(ctx context.Context, input *sfn.DescribeStateMachineInput) (*sfn.DescribeStateMachineOutput, error) {
					return &sfn.DescribeStateMachineOutput{
						Type: types.StateMachineTypeExpress,
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedType:    types.StateMachineTypeStandard,
			shouldPanic:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockStepFunctionsClient{}
			tt.mockSetup(mockClient)

			mockT := &MockT{}
			ctx := &TestContext{
				T:                   mockT,
				StepFunctionsClient: mockClient,
			}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertStateMachineType(ctx, tt.stateMachineArn, tt.expectedType)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertStateMachineType(ctx, tt.stateMachineArn, tt.expectedType)
				})
			}
		})
	}
}

func TestAssertExecutionCountAdvanced(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(*MockStepFunctionsClient)
		stateMachineArn string
		expectedCount   int
		shouldPanic     bool
	}{
		{
			name: "MatchingExecutionCount",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.ListExecutionsFunc = func(ctx context.Context, input *sfn.ListExecutionsInput) (*sfn.ListExecutionsOutput, error) {
					return &sfn.ListExecutionsOutput{
						Executions: []types.ExecutionListItem{
							{ExecutionArn: aws.String("exec1")},
							{ExecutionArn: aws.String("exec2")},
							{ExecutionArn: aws.String("exec3")},
						},
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedCount:   3,
			shouldPanic:     false,
		},
		{
			name: "NonMatchingExecutionCount",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.ListExecutionsFunc = func(ctx context.Context, input *sfn.ListExecutionsInput) (*sfn.ListExecutionsOutput, error) {
					return &sfn.ListExecutionsOutput{
						Executions: []types.ExecutionListItem{
							{ExecutionArn: aws.String("exec1")},
						},
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedCount:   3,
			shouldPanic:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockStepFunctionsClient{}
			tt.mockSetup(mockClient)

			mockT := &MockT{}
			ctx := &TestContext{
				T:                   mockT,
				StepFunctionsClient: mockClient,
			}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertExecutionCount(ctx, tt.stateMachineArn, tt.expectedCount)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertExecutionCount(ctx, tt.stateMachineArn, tt.expectedCount)
				})
			}
		})
	}
}

func TestAssertExecutionCountByStatusAdvanced(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(*MockStepFunctionsClient)
		stateMachineArn string
		status          types.ExecutionStatus
		expectedCount   int
		shouldPanic     bool
	}{
		{
			name: "MatchingSucceededCount",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.ListExecutionsFunc = func(ctx context.Context, input *sfn.ListExecutionsInput) (*sfn.ListExecutionsOutput, error) {
					return &sfn.ListExecutionsOutput{
						Executions: []types.ExecutionListItem{
							{Status: types.ExecutionStatusSucceeded},
							{Status: types.ExecutionStatusSucceeded},
							{Status: types.ExecutionStatusFailed},
						},
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			status:          types.ExecutionStatusSucceeded,
			expectedCount:   2,
			shouldPanic:     false,
		},
		{
			name: "NonMatchingFailedCount",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.ListExecutionsFunc = func(ctx context.Context, input *sfn.ListExecutionsInput) (*sfn.ListExecutionsOutput, error) {
					return &sfn.ListExecutionsOutput{
						Executions: []types.ExecutionListItem{
							{Status: types.ExecutionStatusSucceeded},
							{Status: types.ExecutionStatusFailed},
						},
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			status:          types.ExecutionStatusFailed,
			expectedCount:   2, // Expected 2 but only 1 failed
			shouldPanic:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockStepFunctionsClient{}
			tt.mockSetup(mockClient)

			mockT := &MockT{}
			ctx := &TestContext{
				T:                   mockT,
				StepFunctionsClient: mockClient,
			}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertExecutionCountByStatus(ctx, tt.stateMachineArn, tt.status, tt.expectedCount)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertExecutionCountByStatus(ctx, tt.stateMachineArn, tt.status, tt.expectedCount)
				})
			}
		})
	}
}

func TestAssertHistoryEventExistsAdvanced(t *testing.T) {
	baseTime := time.Now()
	history := []HistoryEvent{
		{ID: 1, Type: types.HistoryEventTypeExecutionStarted, Timestamp: baseTime},
		{ID: 2, Type: types.HistoryEventTypeTaskSucceeded, Timestamp: baseTime.Add(time.Second)},
		{ID: 3, Type: types.HistoryEventTypeExecutionSucceeded, Timestamp: baseTime.Add(2 * time.Second)},
	}

	tests := []struct {
		name        string
		history     []HistoryEvent
		eventType   types.HistoryEventType
		shouldPanic bool
	}{
		{
			name:        "ExistingEventType",
			history:     history,
			eventType:   types.HistoryEventTypeTaskSucceeded,
			shouldPanic: false,
		},
		{
			name:        "NonExistingEventType",
			history:     history,
			eventType:   types.HistoryEventTypeTaskFailed,
			shouldPanic: true,
		},
		{
			name:        "EmptyHistory",
			history:     []HistoryEvent{},
			eventType:   types.HistoryEventTypeExecutionStarted,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &MockT{}
			ctx := &TestContext{T: mockT}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertHistoryEventExists(ctx, tt.history, tt.eventType)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertHistoryEventExists(ctx, tt.history, tt.eventType)
				})
			}
		})
	}
}

func TestAssertHistoryEventCountAdvanced(t *testing.T) {
	baseTime := time.Now()
	history := []HistoryEvent{
		{ID: 1, Type: types.HistoryEventTypeExecutionStarted, Timestamp: baseTime},
		{ID: 2, Type: types.HistoryEventTypeTaskSucceeded, Timestamp: baseTime.Add(time.Second)},
		{ID: 3, Type: types.HistoryEventTypeTaskSucceeded, Timestamp: baseTime.Add(2 * time.Second)},
		{ID: 4, Type: types.HistoryEventTypeExecutionSucceeded, Timestamp: baseTime.Add(3 * time.Second)},
	}

	tests := []struct {
		name          string
		history       []HistoryEvent
		eventType     types.HistoryEventType
		expectedCount int
		shouldPanic   bool
	}{
		{
			name:          "CorrectTaskSucceededCount",
			history:       history,
			eventType:     types.HistoryEventTypeTaskSucceeded,
			expectedCount: 2,
			shouldPanic:   false,
		},
		{
			name:          "IncorrectExecutionStartedCount",
			history:       history,
			eventType:     types.HistoryEventTypeExecutionStarted,
			expectedCount: 2, // Expected 2 but only 1 exists
			shouldPanic:   true,
		},
		{
			name:          "ZeroCountForNonExistentType",
			history:       history,
			eventType:     types.HistoryEventTypeTaskFailed,
			expectedCount: 0,
			shouldPanic:   false,
		},
		{
			name:          "NonZeroCountForNonExistentType",
			history:       history,
			eventType:     types.HistoryEventTypeTaskFailed,
			expectedCount: 1, // Expected 1 but 0 exists
			shouldPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &MockT{}
			ctx := &TestContext{T: mockT}

			if tt.shouldPanic {
				assert.Panics(t, func() {
					AssertHistoryEventCount(ctx, tt.history, tt.eventType, tt.expectedCount)
				})
			} else {
				assert.NotPanics(t, func() {
					AssertHistoryEventCount(ctx, tt.history, tt.eventType, tt.expectedCount)
				})
			}
		})
	}
}

func TestListExecutionsEAdvanced(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(*MockStepFunctionsClient)
		stateMachineArn string
		expectError     bool
		expectedCount   int
	}{
		{
			name: "SuccessfulListing",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.ListExecutionsFunc = func(ctx context.Context, input *sfn.ListExecutionsInput) (*sfn.ListExecutionsOutput, error) {
					return &sfn.ListExecutionsOutput{
						Executions: []types.ExecutionListItem{
							{
								ExecutionArn: aws.String("arn:aws:states:us-east-1:123456789012:execution:test-machine:exec1"),
								Status:       types.ExecutionStatusSucceeded,
							},
							{
								ExecutionArn: aws.String("arn:aws:states:us-east-1:123456789012:execution:test-machine:exec2"),
								Status:       types.ExecutionStatusFailed,
							},
						},
					}, nil
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectError:     false,
			expectedCount:   2,
		},
		{
			name: "InvalidStateMachineArn",
			mockSetup: func(client *MockStepFunctionsClient) {
				// No setup needed as validation should fail first
			},
			stateMachineArn: "invalid-arn",
			expectError:     true,
			expectedCount:   0,
		},
		{
			name: "AWSServiceError",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.ListExecutionsFunc = func(ctx context.Context, input *sfn.ListExecutionsInput) (*sfn.ListExecutionsOutput, error) {
					return nil, assert.AnError
				}
			},
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectError:     true,
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockStepFunctionsClient{}
			tt.mockSetup(mockClient)

			ctx := &TestContext{
				T:                   &MockT{},
				StepFunctionsClient: mockClient,
			}

			result, err := ListExecutionsE(ctx, tt.stateMachineArn)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}