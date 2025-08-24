package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example tests demonstrating execution history analysis capabilities
// These examples show how to use the history analysis and debugging utilities

// TestHistoryAnalysisExample demonstrates comprehensive execution analysis
func TestHistoryAnalysisExample(t *testing.T) {
	t.Run("complete_execution_analysis", func(t *testing.T) {
		// Create sample execution history (simulated)
		history := []HistoryEvent{
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
					Name: "ValidateOrder",
				},
			},
			{
				ID:        3,
				Type:      types.HistoryEventTypeTaskSucceeded,
				Timestamp: time.Now().Add(-8 * time.Minute),
				TaskSucceededEventDetails: &TaskSucceededEventDetails{
					Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ValidateOrder",
					ResourceType: "lambda",
					Output:       `{"valid": true}`,
				},
			},
			{
				ID:        4,
				Type:      types.HistoryEventTypeTaskStateEntered,
				Timestamp: time.Now().Add(-7 * time.Minute),
				StateEnteredEventDetails: &StateEnteredEventDetails{
					Name: "ProcessPayment",
				},
			},
			{
				ID:        5,
				Type:      types.HistoryEventTypeTaskFailed,
				Timestamp: time.Now().Add(-6 * time.Minute),
				TaskFailedEventDetails: &TaskFailedEventDetails{
					Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessPayment",
					ResourceType: "lambda",
					Error:        "States.TaskFailed",
					Cause:        "Insufficient funds",
				},
			},
			{
				ID:        6,
				Type:      types.HistoryEventTypeExecutionFailed,
				Timestamp: time.Now().Add(-5 * time.Minute),
				ExecutionFailedEventDetails: &ExecutionFailedEventDetails{
					Error: "States.TaskFailed",
					Cause: "ProcessPayment step failed",
				},
			},
		}

		// Analyze execution history
		analysis, err := AnalyzeExecutionHistoryE(history)
		require.NoError(t, err)
		require.NotNil(t, analysis)

		// Verify analysis results
		assert.Equal(t, 6, analysis.TotalSteps)
		assert.Equal(t, 1, analysis.CompletedSteps) // ValidateOrder succeeded
		assert.Equal(t, 1, analysis.FailedSteps)    // ProcessPayment failed
		assert.Equal(t, 0, analysis.RetryAttempts)  // No retries
		assert.Equal(t, types.ExecutionStatusFailed, analysis.ExecutionStatus)
		assert.Equal(t, "States.TaskFailed", analysis.FailureReason)
		assert.Equal(t, "ProcessPayment step failed", analysis.FailureCause)

		// Check that we have timing and resource information
		assert.NotEmpty(t, analysis.StepTimings)
		assert.NotEmpty(t, analysis.ResourceUsage)

		t.Logf("Analysis: %+v", analysis)
	})
}

// TestFailedStepsAnalysisExample demonstrates failed step analysis
func TestFailedStepsAnalysisExample(t *testing.T) {
	t.Run("identify_failed_steps", func(t *testing.T) {
		// Create history with multiple failed steps
		history := []HistoryEvent{
			{
				ID:        1,
				Type:      types.HistoryEventTypeTaskFailed,
				Timestamp: time.Now().Add(-10 * time.Minute),
				TaskFailedEventDetails: &TaskFailedEventDetails{
					Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
					ResourceType: "lambda",
					Error:        "States.TaskFailed",
					Cause:        "Validation error",
				},
			},
			{
				ID:        2,
				Type:      types.HistoryEventTypeLambdaFunctionFailed,
				Timestamp: time.Now().Add(-5 * time.Minute),
				LambdaFunctionFailedEventDetails: &LambdaFunctionFailedEventDetails{
					Error: "Runtime.HandlerNotFound",
					Cause: "Handler function not found",
				},
			},
		}

		// Find all failed steps
		failedSteps, err := FindFailedStepsE(history)
		require.NoError(t, err)
		assert.Len(t, failedSteps, 2)

		// Verify first failed step
		assert.Equal(t, "ProcessOrder", failedSteps[0].StepName)
		assert.Equal(t, "States.TaskFailed", failedSteps[0].Error)
		assert.Equal(t, "Validation error", failedSteps[0].Cause)

		t.Logf("Failed steps: %+v", failedSteps)
	})
}

// TestExecutionTimelineExample demonstrates execution timeline creation
func TestExecutionTimelineExample(t *testing.T) {
	t.Run("create_execution_timeline", func(t *testing.T) {
		// Create sample history events
		baseTime := time.Now().Add(-10 * time.Minute)
		history := []HistoryEvent{
			{
				ID:        1,
				Type:      types.HistoryEventTypeExecutionStarted,
				Timestamp: baseTime,
			},
			{
				ID:        2,
				Type:      types.HistoryEventTypeTaskStateEntered,
				Timestamp: baseTime.Add(1 * time.Minute),
				StateEnteredEventDetails: &StateEnteredEventDetails{
					Name: "ProcessOrder",
				},
			},
			{
				ID:        3,
				Type:      types.HistoryEventTypeTaskSucceeded,
				Timestamp: baseTime.Add(5 * time.Minute),
			},
			{
				ID:        4,
				Type:      types.HistoryEventTypeExecutionSucceeded,
				Timestamp: baseTime.Add(6 * time.Minute),
			},
		}

		// Create execution timeline
		timeline, err := GetExecutionTimelineE(history)
		require.NoError(t, err)
		assert.Len(t, timeline, 4)

		// Verify timeline is chronologically ordered
		for i := 1; i < len(timeline); i++ {
			assert.True(t, timeline[i-1].Timestamp.Before(timeline[i].Timestamp) ||
				timeline[i-1].Timestamp.Equal(timeline[i].Timestamp))
		}

		// Verify durations are calculated
		assert.Equal(t, time.Duration(0), timeline[0].Duration) // First event has no duration
		assert.Equal(t, 1*time.Minute, timeline[1].Duration)   // Second event duration

		t.Logf("Timeline: %+v", timeline)
	})
}

// TestExecutionSummaryFormattingExample demonstrates summary formatting
func TestExecutionSummaryFormattingExample(t *testing.T) {
	t.Run("format_execution_summary", func(t *testing.T) {
		// Create analysis data
		analysis := &ExecutionAnalysis{
			TotalSteps:      10,
			CompletedSteps:  8,
			FailedSteps:     2,
			RetryAttempts:   1,
			TotalDuration:   5 * time.Minute,
			ExecutionStatus: types.ExecutionStatusFailed,
			FailureReason:   "States.TaskFailed",
			FailureCause:    "Payment processing failed",
			StepTimings: map[string]time.Duration{
				"ValidateOrder":  30 * time.Second,
				"ProcessPayment": 2 * time.Minute,
			},
			ResourceUsage: map[string]int{
				"arn:aws:lambda:us-east-1:123456789012:function:ValidateOrder":  1,
				"arn:aws:lambda:us-east-1:123456789012:function:ProcessPayment": 2,
			},
		}

		// Format summary
		summary, err := FormatExecutionSummaryE(analysis)
		require.NoError(t, err)
		assert.NotEmpty(t, summary)

		// Verify summary contains key information
		assert.Contains(t, summary, "Total Steps: 10")
		assert.Contains(t, summary, "Completed: 8")
		assert.Contains(t, summary, "Failed: 2")
		assert.Contains(t, summary, "Status: FAILED")
		assert.Contains(t, summary, "Error: States.TaskFailed")
		assert.Contains(t, summary, "ValidateOrder: 30s")

		t.Logf("Summary:\n%s", summary)
	})
}

// TestHistoryAnalysisPatterns demonstrates common analysis patterns
func TestHistoryAnalysisPatterns(t *testing.T) {
	t.Run("complete_diagnostic_workflow", func(t *testing.T) {
		// Simulate a real-world scenario where execution fails
		// and we need to diagnose the issue

		// Step 1: Get execution history (mocked here)
		history := createComplexExecutionHistory()

		// Step 2: Perform comprehensive analysis
		analysis, err := AnalyzeExecutionHistoryE(history)
		require.NoError(t, err)

		// Step 3: Find specific failed steps for detailed investigation
		failedSteps, err := FindFailedStepsE(history)
		require.NoError(t, err)

		// Step 4: Check for retry patterns
		retryAttempts, err := GetRetryAttemptsE(history)
		require.NoError(t, err)

		// Step 5: Create timeline for step-by-step analysis
		timeline, err := GetExecutionTimelineE(history)
		require.NoError(t, err)

		// Step 6: Generate human-readable summary
		summary, err := FormatExecutionSummaryE(analysis)
		require.NoError(t, err)

		// Verify we have comprehensive diagnostic information
		assert.NotNil(t, analysis)
		assert.NotEmpty(t, failedSteps)
		assert.NotNil(t, retryAttempts)
		assert.NotEmpty(t, timeline)
		assert.NotEmpty(t, summary)

		// Log diagnostic information
		t.Logf("=== EXECUTION DIAGNOSTIC REPORT ===")
		t.Logf("Analysis: %+v", analysis)
		t.Logf("Failed Steps: %+v", failedSteps)
		t.Logf("Retry Attempts: %+v", retryAttempts)
		t.Logf("Timeline Events: %d", len(timeline))
		t.Logf("Summary:\n%s", summary)
	})
}

// Helper function to create complex execution history for testing
func createComplexExecutionHistory() []HistoryEvent {
	baseTime := time.Now().Add(-20 * time.Minute)
	
	return []HistoryEvent{
		{
			ID:        1,
			Type:      types.HistoryEventTypeExecutionStarted,
			Timestamp: baseTime,
		},
		{
			ID:        2,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: baseTime.Add(30 * time.Second),
			StateEnteredEventDetails: &StateEnteredEventDetails{
				Name: "ValidateInput",
			},
		},
		{
			ID:        3,
			Type:      types.HistoryEventTypeTaskSucceeded,
			Timestamp: baseTime.Add(45 * time.Second),
			TaskSucceededEventDetails: &TaskSucceededEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ValidateInput",
				ResourceType: "lambda",
			},
		},
		{
			ID:        4,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: baseTime.Add(1 * time.Minute),
			StateEnteredEventDetails: &StateEnteredEventDetails{
				Name: "ProcessOrder",
			},
		},
		{
			ID:        5,
			Type:      types.HistoryEventTypeTaskFailed,
			Timestamp: baseTime.Add(3 * time.Minute),
			TaskFailedEventDetails: &TaskFailedEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
				ResourceType: "lambda",
				Error:        "States.TaskFailed",
				Cause:        "Database connection timeout",
			},
		},
		{
			ID:        6,
			Type:      types.HistoryEventTypeTaskFailed,
			Timestamp: baseTime.Add(5 * time.Minute),
			TaskFailedEventDetails: &TaskFailedEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
				ResourceType: "lambda",
				Error:        "States.TaskFailed",
				Cause:        "Retry failed - still cannot connect",
			},
		},
		{
			ID:        7,
			Type:      types.HistoryEventTypeExecutionFailed,
			Timestamp: baseTime.Add(6 * time.Minute),
			ExecutionFailedEventDetails: &ExecutionFailedEventDetails{
				Error: "States.TaskFailed",
				Cause: "ProcessOrder step exhausted all retry attempts",
			},
		},
	}
}