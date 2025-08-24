// Package stepfunctions history - Execution history analysis and debugging utilities
// Following strict Terratest patterns with Function/FunctionE variants

package stepfunctions

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

// History Retrieval Functions

// GetExecutionHistory gets the complete execution history for a Step Functions execution
func GetExecutionHistory(ctx *TestContext, executionArn string) []HistoryEvent {
	history, err := GetExecutionHistoryE(ctx, executionArn)
	if err != nil {
		ctx.T.FailNow()
	}
	return history
}

// GetExecutionHistoryE gets the complete execution history for a Step Functions execution with error return
func GetExecutionHistoryE(ctx *TestContext, executionArn string) ([]HistoryEvent, error) {
	start := time.Now()
	logOperation("GetExecutionHistory", map[string]interface{}{
		"execution": executionArn,
	})

	// Validate input parameters
	if err := validateExecutionArn(executionArn); err != nil {
		logResult("GetExecutionHistory", false, time.Since(start), err)
		return nil, err
	}

	client := createStepFunctionsClient(ctx)

	var allEvents []HistoryEvent
	var nextToken *string

	for {
		input := &sfn.GetExecutionHistoryInput{
			ExecutionArn: aws.String(executionArn),
			MaxResults:   1000, // Maximum allowed by AWS
			ReverseOrder: false, // Get events in chronological order
		}

		if nextToken != nil {
			input.NextToken = nextToken
		}

		result, err := client.GetExecutionHistory(context.TODO(), input)
		if err != nil {
			logResult("GetExecutionHistory", false, time.Since(start), err)
			return nil, fmt.Errorf("failed to get execution history: %w", err)
		}

		// Convert AWS events to our format
		for _, awsEvent := range result.Events {
			event := convertAwsEventToHistoryEvent(awsEvent)
			allEvents = append(allEvents, event)
		}

		// Check if there are more events to fetch
		nextToken = result.NextToken
		if nextToken == nil {
			break
		}
	}

	logResult("GetExecutionHistory", true, time.Since(start), nil)
	return allEvents, nil
}

// History Analysis Functions

// AnalyzeExecutionHistory analyzes execution history to provide comprehensive insights
func AnalyzeExecutionHistory(history []HistoryEvent) *ExecutionAnalysis {
	analysis, err := AnalyzeExecutionHistoryE(history)
	if err != nil {
		return nil
	}
	return analysis
}

// AnalyzeExecutionHistoryE analyzes execution history with error return
func AnalyzeExecutionHistoryE(history []HistoryEvent) (*ExecutionAnalysis, error) {
	if len(history) == 0 {
		return nil, fmt.Errorf("cannot analyze empty history")
	}

	analysis := &ExecutionAnalysis{
		StepTimings:   make(map[string]time.Duration),
		ResourceUsage: make(map[string]int),
	}

	// Sort events by timestamp to ensure correct order
	sortedHistory := make([]HistoryEvent, len(history))
	copy(sortedHistory, history)
	sort.Slice(sortedHistory, func(i, j int) bool {
		return sortedHistory[i].Timestamp.Before(sortedHistory[j].Timestamp)
	})

	// Analyze execution overview
	analysis.TotalSteps = len(sortedHistory)
	analysis.TotalDuration = calculateTotalDuration(sortedHistory)
	analysis.ExecutionStatus = determineExecutionStatus(sortedHistory)

	// Count different types of steps
	analysis.CompletedSteps = countEventsByType(sortedHistory, []types.HistoryEventType{
		types.HistoryEventTypeTaskSucceeded,
		types.HistoryEventTypeLambdaFunctionSucceeded,
	})

	analysis.FailedSteps = countEventsByType(sortedHistory, []types.HistoryEventType{
		types.HistoryEventTypeTaskFailed,
		types.HistoryEventTypeLambdaFunctionFailed,
		types.HistoryEventTypeTaskTimedOut,
	})

	// For retry attempts, count unique step failures instead of all failures
	analysis.RetryAttempts = countRetryAttempts(sortedHistory)

	// Find failure details if execution failed
	if analysis.ExecutionStatus == types.ExecutionStatusFailed {
		analysis.FailureReason, analysis.FailureCause = extractFailureDetails(sortedHistory)
	}

	// Extract key events
	analysis.KeyEvents = extractKeyEvents(sortedHistory)

	// Calculate step timings
	analysis.StepTimings = calculateStepTimings(sortedHistory)

	// Track resource usage
	analysis.ResourceUsage = calculateResourceUsage(sortedHistory)

	return analysis, nil
}

// FindFailedSteps finds all failed steps in the execution history
func FindFailedSteps(history []HistoryEvent) []FailedStep {
	steps, err := FindFailedStepsE(history)
	if err != nil {
		return []FailedStep{}
	}
	return steps
}

// FindFailedStepsE finds all failed steps in the execution history with error return
func FindFailedStepsE(history []HistoryEvent) ([]FailedStep, error) {
	var failedSteps []FailedStep

	for _, event := range history {
		switch event.Type {
		case types.HistoryEventTypeTaskFailed:
			if event.TaskFailedEventDetails != nil {
				failedStep := FailedStep{
					StepName:     extractStepNameFromEvent(event),
					EventID:      event.ID,
					FailureTime:  event.Timestamp,
					Error:        event.TaskFailedEventDetails.Error,
					Cause:        event.TaskFailedEventDetails.Cause,
					ResourceType: event.TaskFailedEventDetails.ResourceType,
					ResourceArn:  event.TaskFailedEventDetails.Resource,
				}
				failedSteps = append(failedSteps, failedStep)
			}

		case types.HistoryEventTypeLambdaFunctionFailed:
			if event.LambdaFunctionFailedEventDetails != nil {
				failedStep := FailedStep{
					StepName:     extractStepNameFromEvent(event),
					EventID:      event.ID,
					FailureTime:  event.Timestamp,
					Error:        event.LambdaFunctionFailedEventDetails.Error,
					Cause:        event.LambdaFunctionFailedEventDetails.Cause,
					ResourceType: "lambda",
				}
				failedSteps = append(failedSteps, failedStep)
			}
		}
	}

	return failedSteps, nil
}

// GetRetryAttempts gets all retry attempts from execution history
func GetRetryAttempts(history []HistoryEvent) []RetryAttempt {
	attempts, err := GetRetryAttemptsE(history)
	if err != nil {
		return []RetryAttempt{}
	}
	return attempts
}

// GetRetryAttemptsE gets all retry attempts from execution history with error return
func GetRetryAttemptsE(history []HistoryEvent) ([]RetryAttempt, error) {
	var retryAttempts []RetryAttempt
	attemptCounter := make(map[string]int)

	for _, event := range history {
		if event.Type == types.HistoryEventTypeTaskFailed {
			stepName := extractStepNameFromEvent(event)
			attemptCounter[stepName]++

			if event.TaskFailedEventDetails != nil {
				retry := RetryAttempt{
					StepName:      stepName,
					AttemptNumber: attemptCounter[stepName],
					RetryTime:     event.Timestamp,
					Error:         event.TaskFailedEventDetails.Error,
					Cause:         event.TaskFailedEventDetails.Cause,
					ResourceType:  event.TaskFailedEventDetails.ResourceType,
					ResourceArn:   event.TaskFailedEventDetails.Resource,
				}
				retryAttempts = append(retryAttempts, retry)
			}
		}
	}

	return retryAttempts, nil
}

// Debugging Utilities

// GetExecutionTimeline creates a chronological timeline of execution events
func GetExecutionTimeline(history []HistoryEvent) []TimelineEvent {
	timeline, err := GetExecutionTimelineE(history)
	if err != nil {
		return []TimelineEvent{}
	}
	return timeline
}

// GetExecutionTimelineE creates a chronological timeline of execution events with error return
func GetExecutionTimelineE(history []HistoryEvent) ([]TimelineEvent, error) {
	var timeline []TimelineEvent

	// Sort by timestamp
	sortedHistory := make([]HistoryEvent, len(history))
	copy(sortedHistory, history)
	sort.Slice(sortedHistory, func(i, j int) bool {
		return sortedHistory[i].Timestamp.Before(sortedHistory[j].Timestamp)
	})

	for i, event := range sortedHistory {
		timelineEvent := TimelineEvent{
			EventID:     event.ID,
			EventType:   event.Type,
			Timestamp:   event.Timestamp,
			StepName:    extractStepNameFromEvent(event),
			Description: formatEventDescription(event),
			Details:     extractEventDetails(event),
		}

		// Calculate duration from previous event
		if i > 0 {
			timelineEvent.Duration = event.Timestamp.Sub(sortedHistory[i-1].Timestamp)
		}

		timeline = append(timeline, timelineEvent)
	}

	return timeline, nil
}

// FormatExecutionSummary formats execution analysis into readable summary
func FormatExecutionSummary(analysis *ExecutionAnalysis) string {
	summary, err := FormatExecutionSummaryE(analysis)
	if err != nil {
		return "Error formatting summary"
	}
	return summary
}

// FormatExecutionSummaryE formats execution analysis into readable summary with error return
func FormatExecutionSummaryE(analysis *ExecutionAnalysis) (string, error) {
	if analysis == nil {
		return "", fmt.Errorf("analysis cannot be nil")
	}

	var summary strings.Builder

	// Header
	summary.WriteString("=== Step Functions Execution Summary ===\n\n")

	// Basic metrics
	summary.WriteString(fmt.Sprintf("Total Steps: %d\n", analysis.TotalSteps))
	summary.WriteString(fmt.Sprintf("Completed: %d\n", analysis.CompletedSteps))
	summary.WriteString(fmt.Sprintf("Failed: %d\n", analysis.FailedSteps))
	summary.WriteString(fmt.Sprintf("Retry Attempts: %d\n", analysis.RetryAttempts))
	summary.WriteString(fmt.Sprintf("Duration: %s\n", analysis.TotalDuration))
	summary.WriteString(fmt.Sprintf("Status: %s\n", analysis.ExecutionStatus))

	// Failure details if applicable
	if analysis.ExecutionStatus == types.ExecutionStatusFailed {
		summary.WriteString("\n=== Failure Details ===\n")
		if analysis.FailureReason != "" {
			summary.WriteString(fmt.Sprintf("Error: %s\n", analysis.FailureReason))
		}
		if analysis.FailureCause != "" {
			summary.WriteString(fmt.Sprintf("Cause: %s\n", analysis.FailureCause))
		}
	}

	// Step timings
	if len(analysis.StepTimings) > 0 {
		summary.WriteString("\n=== Step Timings ===\n")
		for stepName, duration := range analysis.StepTimings {
			summary.WriteString(fmt.Sprintf("%s: %s\n", stepName, duration))
		}
	}

	// Resource usage
	if len(analysis.ResourceUsage) > 0 {
		summary.WriteString("\n=== Resource Usage ===\n")
		for resource, count := range analysis.ResourceUsage {
			summary.WriteString(fmt.Sprintf("%s: %d invocations\n", resource, count))
		}
	}

	return summary.String(), nil
}

// Internal helper functions for history analysis

// convertAwsEventToHistoryEvent converts AWS SDK event to our internal format
func convertAwsEventToHistoryEvent(awsEvent types.HistoryEvent) HistoryEvent {
	event := HistoryEvent{
		ID:              awsEvent.Id,
		Type:            awsEvent.Type,
		Timestamp:       aws.ToTime(awsEvent.Timestamp),
		PreviousEventID: awsEvent.PreviousEventId,
	}

	// Convert event details based on type
	switch awsEvent.Type {
	case types.HistoryEventTypeTaskFailed:
		if awsEvent.TaskFailedEventDetails != nil {
			event.TaskFailedEventDetails = &TaskFailedEventDetails{
				Resource:     aws.ToString(awsEvent.TaskFailedEventDetails.Resource),
				ResourceType: aws.ToString(awsEvent.TaskFailedEventDetails.ResourceType),
				Error:        aws.ToString(awsEvent.TaskFailedEventDetails.Error),
				Cause:        aws.ToString(awsEvent.TaskFailedEventDetails.Cause),
			}
		}

	case types.HistoryEventTypeTaskSucceeded:
		if awsEvent.TaskSucceededEventDetails != nil {
			event.TaskSucceededEventDetails = &TaskSucceededEventDetails{
				Resource:     aws.ToString(awsEvent.TaskSucceededEventDetails.Resource),
				ResourceType: aws.ToString(awsEvent.TaskSucceededEventDetails.ResourceType),
				Output:       aws.ToString(awsEvent.TaskSucceededEventDetails.Output),
			}
		}

	// Note: TaskRetry event type handling removed as it's not available in AWS SDK

	case types.HistoryEventTypeTaskStateEntered:
		if awsEvent.StateEnteredEventDetails != nil {
			event.StateEnteredEventDetails = &StateEnteredEventDetails{
				Name:  aws.ToString(awsEvent.StateEnteredEventDetails.Name),
				Input: aws.ToString(awsEvent.StateEnteredEventDetails.Input),
			}
		}

	case types.HistoryEventTypeExecutionFailed:
		if awsEvent.ExecutionFailedEventDetails != nil {
			event.ExecutionFailedEventDetails = &ExecutionFailedEventDetails{
				Error: aws.ToString(awsEvent.ExecutionFailedEventDetails.Error),
				Cause: aws.ToString(awsEvent.ExecutionFailedEventDetails.Cause),
			}
		}
	}

	return event
}

// calculateTotalDuration calculates the total execution duration from history
func calculateTotalDuration(history []HistoryEvent) time.Duration {
	if len(history) == 0 {
		return 0
	}

	var startTime, endTime time.Time

	for _, event := range history {
		switch event.Type {
		case types.HistoryEventTypeExecutionStarted:
			if startTime.IsZero() || event.Timestamp.Before(startTime) {
				startTime = event.Timestamp
			}
		case types.HistoryEventTypeExecutionSucceeded,
			types.HistoryEventTypeExecutionFailed,
			types.HistoryEventTypeExecutionTimedOut,
			types.HistoryEventTypeExecutionAborted:
			if endTime.IsZero() || event.Timestamp.After(endTime) {
				endTime = event.Timestamp
			}
		}
	}

	if startTime.IsZero() || endTime.IsZero() {
		return 0
	}

	return endTime.Sub(startTime)
}

// determineExecutionStatus determines the final execution status from history
func determineExecutionStatus(history []HistoryEvent) types.ExecutionStatus {
	for _, event := range history {
		switch event.Type {
		case types.HistoryEventTypeExecutionSucceeded:
			return types.ExecutionStatusSucceeded
		case types.HistoryEventTypeExecutionFailed:
			return types.ExecutionStatusFailed
		case types.HistoryEventTypeExecutionTimedOut:
			return types.ExecutionStatusTimedOut
		case types.HistoryEventTypeExecutionAborted:
			return types.ExecutionStatusAborted
		}
	}
	return types.ExecutionStatusRunning // Default if no terminal event found
}

// countEventsByType counts events of specified types
func countEventsByType(history []HistoryEvent, eventTypes []types.HistoryEventType) int {
	count := 0
	for _, event := range history {
		for _, eventType := range eventTypes {
			if event.Type == eventType {
				count++
				break
			}
		}
	}
	return count
}

// countRetryAttempts counts the number of retry attempts by counting repeated failures per step
func countRetryAttempts(history []HistoryEvent) int {
	stepFailures := make(map[string]int)
	
	for _, event := range history {
		if event.Type == types.HistoryEventTypeTaskFailed {
			stepName := extractStepNameFromEvent(event)
			if stepName != "" {
				stepFailures[stepName]++
			}
		}
	}
	
	// Count total retry attempts (failures beyond the first for each step)
	retryCount := 0
	for _, failures := range stepFailures {
		if failures > 1 {
			retryCount += failures - 1 // First failure is not a retry
		}
	}
	
	return retryCount
}

// extractFailureDetails extracts failure reason and cause from history
func extractFailureDetails(history []HistoryEvent) (string, string) {
	for _, event := range history {
		if event.Type == types.HistoryEventTypeExecutionFailed {
			if event.ExecutionFailedEventDetails != nil {
				return event.ExecutionFailedEventDetails.Error, event.ExecutionFailedEventDetails.Cause
			}
		}
	}
	return "", ""
}

// extractKeyEvents extracts important events from history
func extractKeyEvents(history []HistoryEvent) []HistoryEvent {
	var keyEvents []HistoryEvent

	keyEventTypes := map[types.HistoryEventType]bool{
		types.HistoryEventTypeExecutionStarted:   true,
		types.HistoryEventTypeExecutionSucceeded: true,
		types.HistoryEventTypeExecutionFailed:    true,
		types.HistoryEventTypeExecutionTimedOut:  true,
		types.HistoryEventTypeExecutionAborted:   true,
		types.HistoryEventTypeTaskFailed:         true,
		types.HistoryEventTypeTaskSucceeded:      true,
	}

	for _, event := range history {
		if keyEventTypes[event.Type] {
			keyEvents = append(keyEvents, event)
		}
	}

	return keyEvents
}

// calculateStepTimings calculates timing for each step
func calculateStepTimings(history []HistoryEvent) map[string]time.Duration {
	timings := make(map[string]time.Duration)
	stepStarts := make(map[string]time.Time)

	for _, event := range history {
		stepName := extractStepNameFromEvent(event)
		if stepName == "" {
			continue
		}

		switch event.Type {
		case types.HistoryEventTypeTaskStateEntered:
			stepStarts[stepName] = event.Timestamp
		case types.HistoryEventTypeTaskSucceeded, types.HistoryEventTypeTaskFailed:
			if startTime, exists := stepStarts[stepName]; exists {
				timings[stepName] = event.Timestamp.Sub(startTime)
			}
		}
	}

	return timings
}

// calculateResourceUsage tracks resource invocation counts
func calculateResourceUsage(history []HistoryEvent) map[string]int {
	usage := make(map[string]int)

	for _, event := range history {
		var resource string
		switch event.Type {
		case types.HistoryEventTypeTaskSucceeded:
			if event.TaskSucceededEventDetails != nil {
				resource = event.TaskSucceededEventDetails.Resource
			}
		case types.HistoryEventTypeTaskFailed:
			if event.TaskFailedEventDetails != nil {
				resource = event.TaskFailedEventDetails.Resource
			}
		case types.HistoryEventTypeLambdaFunctionSucceeded, types.HistoryEventTypeLambdaFunctionFailed:
			resource = "Lambda"
		}

		if resource != "" {
			usage[resource]++
		}
	}

	return usage
}

// extractStepNameFromEvent extracts step name from event details
func extractStepNameFromEvent(event HistoryEvent) string {
	if event.StateEnteredEventDetails != nil {
		return event.StateEnteredEventDetails.Name
	}
	if event.TaskStateEnteredEventDetails != nil {
		return event.TaskStateEnteredEventDetails.Name
	}
	// For task events, try to infer from resource ARN
	if event.TaskFailedEventDetails != nil {
		return extractStepNameFromResource(event.TaskFailedEventDetails.Resource)
	}
	if event.TaskSucceededEventDetails != nil {
		return extractStepNameFromResource(event.TaskSucceededEventDetails.Resource)
	}
	return ""
}

// extractStepNameFromResource extracts step name from resource ARN
func extractStepNameFromResource(resource string) string {
	// For Lambda functions, extract function name
	if strings.Contains(resource, ":function:") {
		parts := strings.Split(resource, ":")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	// Default fallback
	return "UnknownStep"
}

// formatEventDescription creates human-readable description of event
func formatEventDescription(event HistoryEvent) string {
	switch event.Type {
	case types.HistoryEventTypeExecutionStarted:
		return "Execution started"
	case types.HistoryEventTypeExecutionSucceeded:
		return "Execution completed successfully"
	case types.HistoryEventTypeExecutionFailed:
		return "Execution failed"
	case types.HistoryEventTypeTaskStateEntered:
		if event.StateEnteredEventDetails != nil {
			return fmt.Sprintf("Entered state: %s", event.StateEnteredEventDetails.Name)
		}
		return "Task state entered"
	case types.HistoryEventTypeTaskSucceeded:
		return "Task completed successfully"
	case types.HistoryEventTypeTaskFailed:
		return "Task failed"
	case types.HistoryEventTypeTaskTimedOut:
		return "Task timed out"
	default:
		return string(event.Type)
	}
}

// extractEventDetails extracts relevant details from event
func extractEventDetails(event HistoryEvent) interface{} {
	switch event.Type {
	case types.HistoryEventTypeTaskFailed:
		return event.TaskFailedEventDetails
	case types.HistoryEventTypeTaskSucceeded:
		return event.TaskSucceededEventDetails
	case types.HistoryEventTypeTaskTimedOut:
		return event.TaskFailedEventDetails // Reuse failed details for timeout
	case types.HistoryEventTypeExecutionFailed:
		return event.ExecutionFailedEventDetails
	case types.HistoryEventTypeTaskStateEntered:
		return event.StateEnteredEventDetails
	default:
		return nil
	}
}