package stepfunctions

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExecutionHistory(t *testing.T) {
	tests := []struct {
		name          string
		executionArn  string
		mockSetup     func(*MockStepFunctionsClient)
		expectedError bool
		expectedPanic bool
	}{
		{
			name:         "SuccessfulHistoryRetrieval",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.GetExecutionHistoryFunc = func(ctx context.Context, input *sfn.GetExecutionHistoryInput) (*sfn.GetExecutionHistoryOutput, error) {
					return &sfn.GetExecutionHistoryOutput{
						Events: []types.HistoryEvent{
							createTestHistoryEvent(1, types.HistoryEventTypeExecutionStarted, time.Now()),
							createTestHistoryEvent(2, types.HistoryEventTypeExecutionSucceeded, time.Now().Add(time.Minute)),
						},
					}, nil
				}
			},
			expectedError: false,
			expectedPanic: false,
		},
		{
			name:         "InvalidExecutionArnShouldPanic",
			executionArn: "invalid-arn",
			mockSetup: func(client *MockStepFunctionsClient) {
				// No setup needed as validation should fail first
			},
			expectedError: true,
			expectedPanic: true,
		},
		{
			name:         "EmptyExecutionArnShouldPanic",
			executionArn: "",
			mockSetup: func(client *MockStepFunctionsClient) {
				// No setup needed as validation should fail first
			},
			expectedError: true,
			expectedPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockStepFunctionsClient{}
			tt.mockSetup(mockClient)

			mockT := &MockT{}
			ctx := &TestContext{
				T:                mockT,
				StepFunctionsClient: mockClient,
			}

			if tt.expectedPanic {
				assert.Panics(t, func() {
					GetExecutionHistory(ctx, tt.executionArn)
				})
			} else {
				result := GetExecutionHistory(ctx, tt.executionArn)
				assert.NotNil(t, result)
				assert.True(t, len(result) >= 0)
			}
		})
	}
}

func TestGetExecutionHistoryE(t *testing.T) {
	tests := []struct {
		name          string
		executionArn  string
		mockSetup     func(*MockStepFunctionsClient)
		expectedError bool
	}{
		{
			name:         "SuccessfulHistoryRetrievalWithPagination",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			mockSetup: func(client *MockStepFunctionsClient) {
				callCount := 0
				client.GetExecutionHistoryFunc = func(ctx context.Context, input *sfn.GetExecutionHistoryInput) (*sfn.GetExecutionHistoryOutput, error) {
					callCount++
					if callCount == 1 {
						// First call with next token
						return &sfn.GetExecutionHistoryOutput{
							Events: []types.HistoryEvent{
								createTestHistoryEvent(1, types.HistoryEventTypeExecutionStarted, time.Now()),
							},
							NextToken: aws.String("next-token"),
						}, nil
					}
					// Second call without next token
					return &sfn.GetExecutionHistoryOutput{
						Events: []types.HistoryEvent{
							createTestHistoryEvent(2, types.HistoryEventTypeExecutionSucceeded, time.Now().Add(time.Minute)),
						},
					}, nil
				}
			},
			expectedError: false,
		},
		{
			name:         "InvalidExecutionArn",
			executionArn: "invalid-arn",
			mockSetup: func(client *MockStepFunctionsClient) {
				// No setup needed as validation should fail first
			},
			expectedError: true,
		},
		{
			name:         "AWSServiceError",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			mockSetup: func(client *MockStepFunctionsClient) {
				client.GetExecutionHistoryFunc = func(ctx context.Context, input *sfn.GetExecutionHistoryInput) (*sfn.GetExecutionHistoryOutput, error) {
					return nil, assert.AnError
				}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockStepFunctionsClient{}
			tt.mockSetup(mockClient)

			ctx := &TestContext{
				T:                &MockT{},
				StepFunctionsClient: mockClient,
			}

			result, err := GetExecutionHistoryE(ctx, tt.executionArn)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAnalyzeExecutionHistory(t *testing.T) {
	tests := []struct {
		name     string
		history  []HistoryEvent
		expected bool // whether analysis should succeed
	}{
		{
			name:     "EmptyHistoryShouldReturnEmptyAnalysis",
			history:  []HistoryEvent{},
			expected: true,
		},
		{
			name: "SuccessfulExecutionAnalysis",
			history: []HistoryEvent{
				createTestHistoryEvent(1, types.HistoryEventTypeExecutionStarted, time.Now()),
				createTestHistoryEvent(2, types.HistoryEventTypeTaskSucceeded, time.Now().Add(time.Second)),
				createTestHistoryEvent(3, types.HistoryEventTypeExecutionSucceeded, time.Now().Add(2*time.Second)),
			},
			expected: true,
		},
		{
			name: "FailedExecutionAnalysis",
			history: []HistoryEvent{
				createTestHistoryEvent(1, types.HistoryEventTypeExecutionStarted, time.Now()),
				createTestHistoryEvent(2, types.HistoryEventTypeTaskFailed, time.Now().Add(time.Second)),
				createTestHistoryEvent(3, types.HistoryEventTypeExecutionFailed, time.Now().Add(2*time.Second)),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeExecutionHistory(tt.history)

			if tt.expected {
				assert.NotNil(t, result)
				assert.NotNil(t, result.StepTimings)
				assert.NotNil(t, result.ResourceUsage)
				assert.NotNil(t, result.KeyEvents)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestAnalyzeExecutionHistoryE(t *testing.T) {
	baseTime := time.Now()
	
	tests := []struct {
		name     string
		history  []HistoryEvent
		validate func(*testing.T, *ExecutionAnalysis, error)
	}{
		{
			name:    "EmptyHistoryAnalysis",
			history: []HistoryEvent{},
			validate: func(t *testing.T, analysis *ExecutionAnalysis, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, analysis)
				assert.Equal(t, 0, analysis.TotalSteps)
				assert.Equal(t, 0, analysis.CompletedSteps)
				assert.Equal(t, 0, analysis.FailedSteps)
				assert.Equal(t, 0, analysis.RetryAttempts)
				assert.Equal(t, time.Duration(0), analysis.TotalDuration)
				assert.Equal(t, types.ExecutionStatusRunning, analysis.ExecutionStatus)
			},
		},
		{
			name: "ComplexExecutionWithRetries",
			history: []HistoryEvent{
				createTestHistoryEventWithDetails(1, types.HistoryEventTypeExecutionStarted, baseTime),
				createTestHistoryEventWithTaskEntered(2, types.HistoryEventTypeTaskStateEntered, baseTime.Add(100*time.Millisecond), "Step1"),
				createTestHistoryEventWithTaskFailure(3, types.HistoryEventTypeTaskFailed, baseTime.Add(200*time.Millisecond), "lambda", "ValidationError", "Input validation failed"),
				createTestHistoryEventWithTaskFailure(4, types.HistoryEventTypeTaskFailed, baseTime.Add(300*time.Millisecond), "lambda", "ValidationError", "Input validation failed"),
				createTestHistoryEventWithTaskSuccess(5, types.HistoryEventTypeTaskSucceeded, baseTime.Add(400*time.Millisecond), "lambda", "success output"),
				createTestHistoryEventWithExecutionFailure(6, types.HistoryEventTypeExecutionFailed, baseTime.Add(500*time.Millisecond), "WorkflowFailed", "Step failed after retries"),
			},
			validate: func(t *testing.T, analysis *ExecutionAnalysis, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, analysis)
				assert.Equal(t, 6, analysis.TotalSteps)
				assert.Equal(t, 1, analysis.CompletedSteps) // One TaskSucceeded
				assert.Equal(t, 2, analysis.FailedSteps)   // Two TaskFailed
				assert.Equal(t, 1, analysis.RetryAttempts) // 2 failures of same step = 1 retry
				assert.Equal(t, types.ExecutionStatusFailed, analysis.ExecutionStatus)
				assert.Equal(t, "WorkflowFailed", analysis.FailureReason)
				assert.Equal(t, "Step failed after retries", analysis.FailureCause)
				assert.True(t, analysis.TotalDuration > 0)
			},
		},
		{
			name: "SuccessfulExecutionWithTimings",
			history: []HistoryEvent{
				createTestHistoryEventWithDetails(1, types.HistoryEventTypeExecutionStarted, baseTime),
				createTestHistoryEventWithTaskEntered(2, types.HistoryEventTypeTaskStateEntered, baseTime.Add(100*time.Millisecond), "ProcessData"),
				createTestHistoryEventWithTaskSuccess(3, types.HistoryEventTypeTaskSucceeded, baseTime.Add(500*time.Millisecond), "arn:aws:lambda:us-east-1:123456789012:function:ProcessData", "processed"),
				createTestHistoryEventWithDetails(4, types.HistoryEventTypeExecutionSucceeded, baseTime.Add(600*time.Millisecond)),
			},
			validate: func(t *testing.T, analysis *ExecutionAnalysis, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, analysis)
				assert.Equal(t, types.ExecutionStatusSucceeded, analysis.ExecutionStatus)
				assert.Equal(t, 1, analysis.CompletedSteps)
				assert.Equal(t, 0, analysis.FailedSteps)
				assert.Contains(t, analysis.StepTimings, "ProcessData")
				assert.True(t, analysis.TotalDuration > 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := AnalyzeExecutionHistoryE(tt.history)
			tt.validate(t, analysis, err)
		})
	}
}

func TestFindFailedSteps(t *testing.T) {
	baseTime := time.Now()
	
	tests := []struct {
		name     string
		history  []HistoryEvent
		expected int
	}{
		{
			name:     "NoFailedSteps",
			history:  []HistoryEvent{},
			expected: 0,
		},
		{
			name: "SingleTaskFailure",
			history: []HistoryEvent{
				createTestHistoryEventWithTaskFailure(1, types.HistoryEventTypeTaskFailed, baseTime, "lambda", "Error", "Task failed"),
			},
			expected: 1,
		},
		{
			name: "MultipleDifferentFailures",
			history: []HistoryEvent{
				createTestHistoryEventWithTaskFailure(1, types.HistoryEventTypeTaskFailed, baseTime, "lambda", "Error1", "First failure"),
				createTestHistoryEventWithLambdaFailure(2, types.HistoryEventTypeLambdaFunctionFailed, baseTime.Add(time.Second), "Error2", "Lambda failure"),
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindFailedSteps(tt.history)
			assert.Len(t, result, tt.expected)
			
			// Test the E version as well
			resultE, err := FindFailedStepsE(tt.history)
			assert.NoError(t, err)
			assert.Len(t, resultE, tt.expected)
			assert.Equal(t, result, resultE)
		})
	}
}

func TestGetRetryAttempts(t *testing.T) {
	baseTime := time.Now()
	
	tests := []struct {
		name     string
		history  []HistoryEvent
		expected int
	}{
		{
			name:     "NoRetries",
			history:  []HistoryEvent{},
			expected: 0,
		},
		{
			name: "SingleStepWithRetries",
			history: []HistoryEvent{
				createTestHistoryEventWithTaskEntered(1, types.HistoryEventTypeTaskStateEntered, baseTime, "Step1"),
				createTestHistoryEventWithTaskFailure(2, types.HistoryEventTypeTaskFailed, baseTime.Add(100*time.Millisecond), "lambda", "Error", "First attempt failed"),
				createTestHistoryEventWithTaskFailure(3, types.HistoryEventTypeTaskFailed, baseTime.Add(200*time.Millisecond), "lambda", "Error", "Second attempt failed"),
				createTestHistoryEventWithTaskSuccess(4, types.HistoryEventTypeTaskSucceeded, baseTime.Add(300*time.Millisecond), "lambda", "Success on third attempt"),
			},
			expected: 2, // 2 retries for Step1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRetryAttempts(tt.history)
			assert.Len(t, result, tt.expected)
			
			// Test the E version as well
			resultE, err := GetRetryAttemptsE(tt.history)
			assert.NoError(t, err)
			assert.Len(t, resultE, tt.expected)
		})
	}
}

func TestGetExecutionTimeline(t *testing.T) {
	baseTime := time.Now()
	
	tests := []struct {
		name     string
		history  []HistoryEvent
		expected int
	}{
		{
			name:     "EmptyHistory",
			history:  []HistoryEvent{},
			expected: 0,
		},
		{
			name: "OrderedEvents",
			history: []HistoryEvent{
				createTestHistoryEventWithDetails(1, types.HistoryEventTypeExecutionStarted, baseTime),
				createTestHistoryEventWithTaskEntered(2, types.HistoryEventTypeTaskStateEntered, baseTime.Add(100*time.Millisecond), "Step1"),
				createTestHistoryEventWithTaskSuccess(3, types.HistoryEventTypeTaskSucceeded, baseTime.Add(500*time.Millisecond), "lambda", "success"),
				createTestHistoryEventWithDetails(4, types.HistoryEventTypeExecutionSucceeded, baseTime.Add(600*time.Millisecond)),
			},
			expected: 4,
		},
		{
			name: "UnorderedEvents",
			history: []HistoryEvent{
				createTestHistoryEventWithDetails(4, types.HistoryEventTypeExecutionSucceeded, baseTime.Add(600*time.Millisecond)),
				createTestHistoryEventWithDetails(1, types.HistoryEventTypeExecutionStarted, baseTime),
				createTestHistoryEventWithTaskEntered(2, types.HistoryEventTypeTaskStateEntered, baseTime.Add(100*time.Millisecond), "Step1"),
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExecutionTimeline(tt.history)
			assert.Len(t, result, tt.expected)
			
			// Test the E version as well
			resultE, err := GetExecutionTimelineE(tt.history)
			assert.NoError(t, err)
			assert.Len(t, resultE, tt.expected)
			
			// Verify chronological ordering
			if len(resultE) > 1 {
				for i := 1; i < len(resultE); i++ {
					assert.True(t, resultE[i].Timestamp.After(resultE[i-1].Timestamp) || resultE[i].Timestamp.Equal(resultE[i-1].Timestamp),
						"Events should be in chronological order")
				}
			}
		})
	}
}

func TestFormatExecutionSummary(t *testing.T) {
	tests := []struct {
		name      string
		analysis  *ExecutionAnalysis
		shouldErr bool
	}{
		{
			name:      "NilAnalysisShouldReturnError",
			analysis:  nil,
			shouldErr: true,
		},
		{
			name: "BasicSummaryFormatting",
			analysis: &ExecutionAnalysis{
				TotalSteps:      5,
				CompletedSteps:  4,
				FailedSteps:     1,
				RetryAttempts:   2,
				TotalDuration:   5 * time.Second,
				ExecutionStatus: types.ExecutionStatusSucceeded,
				StepTimings: map[string]time.Duration{
					"Step1": 1 * time.Second,
					"Step2": 2 * time.Second,
				},
				ResourceUsage: map[string]int{
					"lambda": 3,
					"sns":    1,
				},
			},
			shouldErr: false,
		},
		{
			name: "FailedExecutionSummary",
			analysis: &ExecutionAnalysis{
				TotalSteps:      3,
				CompletedSteps:  1,
				FailedSteps:     2,
				RetryAttempts:   1,
				TotalDuration:   3 * time.Second,
				ExecutionStatus: types.ExecutionStatusFailed,
				FailureReason:   "ValidationError",
				FailureCause:    "Input validation failed",
				StepTimings:     make(map[string]time.Duration),
				ResourceUsage:   make(map[string]int),
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldErr {
				// Test panic version
				result := FormatExecutionSummary(tt.analysis)
				assert.Equal(t, "Error formatting summary", result)
				
				// Test error version
				_, err := FormatExecutionSummaryE(tt.analysis)
				assert.Error(t, err)
			} else {
				// Test panic version
				result := FormatExecutionSummary(tt.analysis)
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "Step Functions Execution Summary")
				
				// Test error version
				resultE, err := FormatExecutionSummaryE(tt.analysis)
				assert.NoError(t, err)
				assert.Equal(t, result, resultE)
			}
		})
	}
}

// Helper functions to create test data

func createTestHistoryEvent(id int32, eventType types.HistoryEventType, timestamp time.Time) HistoryEvent {
	return HistoryEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: timestamp,
	}
}

func createTestHistoryEventWithDetails(id int32, eventType types.HistoryEventType, timestamp time.Time) HistoryEvent {
	return HistoryEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: timestamp,
	}
}

func createTestHistoryEventWithTaskEntered(id int32, eventType types.HistoryEventType, timestamp time.Time, stateName string) HistoryEvent {
	return HistoryEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: timestamp,
		StateEnteredEventDetails: &StateEnteredEventDetails{
			Name:  stateName,
			Input: `{"test": "input"}`,
		},
	}
}

func createTestHistoryEventWithTaskFailure(id int32, eventType types.HistoryEventType, timestamp time.Time, resourceType, error, cause string) HistoryEvent {
	return HistoryEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: timestamp,
		TaskFailedEventDetails: &TaskFailedEventDetails{
			ResourceType: resourceType,
			Error:        error,
			Cause:        cause,
			Resource:     "arn:aws:lambda:us-east-1:123456789012:function:test",
		},
	}
}

func createTestHistoryEventWithTaskSuccess(id int32, eventType types.HistoryEventType, timestamp time.Time, resource, output string) HistoryEvent {
	return HistoryEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: timestamp,
		TaskSucceededEventDetails: &TaskSucceededEventDetails{
			Resource:     resource,
			ResourceType: "lambda",
			Output:       output,
		},
	}
}

func createTestHistoryEventWithLambdaFailure(id int32, eventType types.HistoryEventType, timestamp time.Time, error, cause string) HistoryEvent {
	return HistoryEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: timestamp,
		TaskFailedEventDetails: &TaskFailedEventDetails{
			ResourceType: "lambda",
			Error:        error,
			Cause:        cause,
			Resource:     "arn:aws:lambda:us-east-1:123456789012:function:test",
		},
	}
}

func createTestHistoryEventWithExecutionFailure(id int32, eventType types.HistoryEventType, timestamp time.Time, error, cause string) HistoryEvent {
	return HistoryEvent{
		ID:        id,
		Type:      eventType,
		Timestamp: timestamp,
		ExecutionFailedEventDetails: &ExecutionFailedEventDetails{
			Error: error,
			Cause: cause,
		},
	}
}