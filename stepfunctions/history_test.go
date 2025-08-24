package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data factories for history analysis testing

func createTestHistoryEvents() []HistoryEvent {
	return []HistoryEvent{
		{
			ID:        1,
			Type:      types.HistoryEventTypeExecutionStarted,
			Timestamp: time.Now().Add(-5 * time.Minute),
		},
		{
			ID:        2,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: time.Now().Add(-4 * time.Minute),
			StateEnteredEventDetails: &StateEnteredEventDetails{
				Name: "ProcessOrder",
			},
		},
		{
			ID:        3,
			Type:      types.HistoryEventTypeTaskFailed,
			Timestamp: time.Now().Add(-3 * time.Minute),
			TaskFailedEventDetails: &TaskFailedEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
				ResourceType: "lambda",
				Error:        "States.TaskFailed",
				Cause:        "Lambda function returned error: Insufficient funds",
			},
		},
		{
			ID:        4,
			Type:      types.HistoryEventTypeExecutionFailed,
			Timestamp: time.Now().Add(-2 * time.Minute),
			ExecutionFailedEventDetails: &ExecutionFailedEventDetails{
				Error: "States.TaskFailed",
				Cause: "Task ProcessOrder failed",
			},
		},
	}
}

func createSuccessHistoryEvents() []HistoryEvent {
	return []HistoryEvent{
		{
			ID:        1,
			Type:      types.HistoryEventTypeExecutionStarted,
			Timestamp: time.Now().Add(-3 * time.Minute),
		},
		{
			ID:        2,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: time.Now().Add(-2 * time.Minute),
			StateEnteredEventDetails: &StateEnteredEventDetails{
				Name: "ProcessOrder",
			},
		},
		{
			ID:        3,
			Type:      types.HistoryEventTypeTaskSucceeded,
			Timestamp: time.Now().Add(-1 * time.Minute),
			TaskSucceededEventDetails: &TaskSucceededEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
				ResourceType: "lambda",
			},
		},
		{
			ID:        4,
			Type:      types.HistoryEventTypeExecutionSucceeded,
			Timestamp: time.Now(),
		},
	}
}

func createRetryHistoryEvents() []HistoryEvent {
	return []HistoryEvent{
		{
			ID:        1,
			Type:      types.HistoryEventTypeExecutionStarted,
			Timestamp: time.Now().Add(-10 * time.Minute),
		},
		{
			ID:        2,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: time.Now().Add(-9 * time.Minute),
			StateEnteredEventDetails: &StateEnteredEventDetails{
				Name: "RetryableTask",
			},
		},
		{
			ID:        3,
			Type:      types.HistoryEventTypeTaskFailed,
			Timestamp: time.Now().Add(-8 * time.Minute),
			TaskFailedEventDetails: &TaskFailedEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:RetryableTask",
				ResourceType: "lambda",
				Error:        "States.TaskFailed",
				Cause:        "Temporary network error",
			},
		},
		{
			ID:        4,
			Type:      types.HistoryEventTypeTaskScheduled,
			Timestamp: time.Now().Add(-7 * time.Minute),
		},
		{
			ID:        5,
			Type:      types.HistoryEventTypeTaskSucceeded,
			Timestamp: time.Now().Add(-6 * time.Minute),
			TaskSucceededEventDetails: &TaskSucceededEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:RetryableTask",
				ResourceType: "lambda",
			},
		},
		{
			ID:        6,
			Type:      types.HistoryEventTypeExecutionSucceeded,
			Timestamp: time.Now().Add(-5 * time.Minute),
		},
	}
}

// Unit tests for execution history retrieval

func TestGetExecutionHistoryE(t *testing.T) {
	tests := []struct {
		name            string
		executionArn    string
		expectedError   bool
		errorContains   string
		expectedMinSize int
		description     string
	}{
		{
			name:            "ValidExecutionArn",
			executionArn:    "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			expectedError:   false,
			expectedMinSize: 0,
			description:     "Valid execution ARN should return history without error",
		},
		{
			name:          "InvalidExecutionArn",
			executionArn:  "invalid-arn",
			expectedError: true,
			errorContains: "invalid ARN",
			description:   "Invalid execution ARN should return error",
		},
		{
			name:          "EmptyExecutionArn",
			executionArn:  "",
			expectedError: true,
			errorContains: "execution ARN is required",
			description:   "Empty execution ARN should return error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test context (mock)
			ctx := &TestContext{T: t}

			// Act
			history, err := GetExecutionHistoryE(ctx, tc.executionArn)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
				assert.Contains(t, err.Error(), tc.errorContains)
				assert.Nil(t, history)
			} else {
				assert.NoError(t, err, tc.description)
				assert.NotNil(t, history)
				assert.GreaterOrEqual(t, len(history), tc.expectedMinSize)
			}
		})
	}
}

// Unit tests for execution analysis functions

func TestAnalyzeExecutionHistoryE(t *testing.T) {
	tests := []struct {
		name        string
		history     []HistoryEvent
		expected    *ExecutionAnalysis
		description string
	}{
		{
			name:    "FailedExecution",
			history: createTestHistoryEvents(),
			expected: &ExecutionAnalysis{
				TotalSteps:      4,
				CompletedSteps:  0, // No TaskSucceeded events
				FailedSteps:     1, // One TaskFailed event
				RetryAttempts:   0, // No retries
				TotalDuration:   3 * time.Minute,
				ExecutionStatus: types.ExecutionStatusFailed,
				FailureReason:   "States.TaskFailed",
				FailureCause:    "Task ProcessOrder failed",
			},
			description: "Failed execution should be analyzed correctly",
		},
		{
			name:    "SuccessfulExecution",
			history: createSuccessHistoryEvents(),
			expected: &ExecutionAnalysis{
				TotalSteps:      4,
				CompletedSteps:  1, // One TaskSucceeded event
				FailedSteps:     0, // No failed tasks
				RetryAttempts:   0, // No retries
				TotalDuration:   3 * time.Minute,
				ExecutionStatus: types.ExecutionStatusSucceeded,
			},
			description: "Successful execution should be analyzed correctly",
		},
		{
			name:    "ExecutionWithRetries",
			history: createRetryHistoryEvents(),
			expected: &ExecutionAnalysis{
				TotalSteps:      6,
				CompletedSteps:  1, // One TaskSucceeded event
				FailedSteps:     1, // One TaskFailed event
				RetryAttempts:   0, // No actual retries (would need multiple failures of same step)
				TotalDuration:   5 * time.Minute,
				ExecutionStatus: types.ExecutionStatusSucceeded,
			},
			description: "Execution with retries should be analyzed correctly",
		},
		{
			name:        "EmptyHistory",
			history:     []HistoryEvent{},
			expected:    nil,
			description: "Empty history should return nil analysis",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := AnalyzeExecutionHistoryE(tc.history)

			// Assert
			if tc.expected == nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err, tc.description)
				require.NotNil(t, result)
				
				assert.Equal(t, tc.expected.TotalSteps, result.TotalSteps)
				assert.Equal(t, tc.expected.CompletedSteps, result.CompletedSteps)
				assert.Equal(t, tc.expected.FailedSteps, result.FailedSteps)
				assert.Equal(t, tc.expected.RetryAttempts, result.RetryAttempts)
				assert.Equal(t, tc.expected.ExecutionStatus, result.ExecutionStatus)
				assert.Equal(t, tc.expected.FailureReason, result.FailureReason)
				assert.Equal(t, tc.expected.FailureCause, result.FailureCause)
				
				// Check duration is approximately correct (within 1 second tolerance)
				assert.InDelta(t, tc.expected.TotalDuration.Seconds(), result.TotalDuration.Seconds(), 1.0)
				
				// Verify we have key events
				assert.NotNil(t, result.KeyEvents)
				assert.NotNil(t, result.StepTimings)
				assert.NotNil(t, result.ResourceUsage)
			}
		})
	}
}

func TestFindFailedStepsE(t *testing.T) {
	tests := []struct {
		name          string
		history       []HistoryEvent
		expectedCount int
		expectedNames []string
		description   string
	}{
		{
			name:          "WithFailedSteps",
			history:       createTestHistoryEvents(),
			expectedCount: 1,
			expectedNames: []string{"ProcessOrder"},
			description:   "History with failed steps should return failed step information",
		},
		{
			name:          "NoFailedSteps",
			history:       createSuccessHistoryEvents(),
			expectedCount: 0,
			expectedNames: []string{},
			description:   "History without failed steps should return empty list",
		},
		{
			name:          "EmptyHistory",
			history:       []HistoryEvent{},
			expectedCount: 0,
			expectedNames: []string{},
			description:   "Empty history should return empty list",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := FindFailedStepsE(tc.history)

			// Assert
			require.NoError(t, err, tc.description)
			assert.Len(t, result, tc.expectedCount)

			if tc.expectedCount > 0 {
				stepNames := make([]string, len(result))
				for i, step := range result {
					stepNames[i] = step.StepName
				}
				assert.ElementsMatch(t, tc.expectedNames, stepNames)
			}
		})
	}
}

func TestGetRetryAttemptsE(t *testing.T) {
	tests := []struct {
		name            string
		history         []HistoryEvent
		expectedCount   int
		expectedRetries []RetryAttempt
		description     string
	}{
		{
			name:          "WithRetryAttempts",
			history:       createRetryHistoryEvents(),
			expectedCount: 1,
			expectedRetries: []RetryAttempt{
				{
					StepName:     "RetryableTask",
					AttemptNumber: 1,
					Error:        "States.TaskFailed",
					Cause:        "Temporary network error",
				},
			},
			description: "History with retry attempts should return retry information",
		},
		{
			name:            "NoRetryAttempts",
			history:         createSuccessHistoryEvents(), // Use success history which has no failures
			expectedCount:   0,
			expectedRetries: []RetryAttempt{},
			description:     "History without retry attempts should return empty list",
		},
		{
			name:            "EmptyHistory",
			history:         []HistoryEvent{},
			expectedCount:   0,
			expectedRetries: []RetryAttempt{},
			description:     "Empty history should return empty list",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := GetRetryAttemptsE(tc.history)

			// Assert
			require.NoError(t, err, tc.description)
			assert.Len(t, result, tc.expectedCount)

			if tc.expectedCount > 0 {
				for i, expected := range tc.expectedRetries {
					assert.Equal(t, expected.StepName, result[i].StepName)
					assert.Equal(t, expected.Error, result[i].Error)
					assert.Equal(t, expected.Cause, result[i].Cause)
				}
			}
		})
	}
}

// Unit tests for debugging utility functions

func TestGetExecutionTimelineE(t *testing.T) {
	tests := []struct {
		name          string
		history       []HistoryEvent
		expectedCount int
		description   string
	}{
		{
			name:          "ValidHistory",
			history:       createTestHistoryEvents(),
			expectedCount: 4,
			description:   "Valid history should return timeline with all events",
		},
		{
			name:          "EmptyHistory",
			history:       []HistoryEvent{},
			expectedCount: 0,
			description:   "Empty history should return empty timeline",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := GetExecutionTimelineE(tc.history)

			// Assert
			require.NoError(t, err, tc.description)
			assert.Len(t, result, tc.expectedCount)

			// Verify timeline is sorted by timestamp
			for i := 1; i < len(result); i++ {
				assert.True(t, result[i-1].Timestamp.Before(result[i].Timestamp) ||
					result[i-1].Timestamp.Equal(result[i].Timestamp),
					"Timeline should be sorted by timestamp")
			}
		})
	}
}

func TestFormatExecutionSummaryE(t *testing.T) {
	analysis := &ExecutionAnalysis{
		TotalSteps:      4,
		CompletedSteps:  1,
		FailedSteps:     1,
		RetryAttempts:   0,
		TotalDuration:   3 * time.Minute,
		ExecutionStatus: types.ExecutionStatusFailed,
		FailureReason:   "States.TaskFailed",
		FailureCause:    "Task ProcessOrder failed",
	}

	// Act
	result, err := FormatExecutionSummaryE(analysis)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, result, "Total Steps: 4")
	assert.Contains(t, result, "Completed: 1")
	assert.Contains(t, result, "Failed: 1")
	assert.Contains(t, result, "Status: FAILED")
	assert.Contains(t, result, "Duration: 3m0s")
	assert.Contains(t, result, "Error: States.TaskFailed")
}

// Benchmark tests for history analysis performance

func BenchmarkAnalyzeExecutionHistory(b *testing.B) {
	history := createTestHistoryEvents()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AnalyzeExecutionHistoryE(history)
	}
}

func BenchmarkFindFailedSteps(b *testing.B) {
	history := createTestHistoryEvents()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FindFailedStepsE(history)
	}
}

func BenchmarkGetRetryAttempts(b *testing.B) {
	history := createRetryHistoryEvents()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetRetryAttemptsE(history)
	}
}