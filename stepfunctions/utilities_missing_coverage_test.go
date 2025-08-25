package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for 0% coverage functions from coverage analysis

func TestDefaultExecutionOptions(t *testing.T) {
	// Act
	options := defaultExecutionOptions()

	// Assert
	assert.NotNil(t, options)
	assert.Empty(t, options.TraceHeader)
	assert.Equal(t, 5*time.Minute, options.Timeout) // DefaultTimeout
	assert.Equal(t, 3, options.MaxRetries)          // DefaultRetryAttempts
	assert.Equal(t, 1*time.Second, options.RetryDelay) // DefaultRetryDelay
}

func TestDefaultWaitOptions(t *testing.T) {
	// Act
	options := defaultWaitOptions()

	// Assert
	assert.NotNil(t, options)
	assert.Equal(t, 5*time.Minute, options.Timeout)       // DefaultTimeout
	assert.Equal(t, 5*time.Second, options.PollingInterval) // DefaultPollingInterval
	assert.Equal(t, 60, options.MaxAttempts)              // DefaultTimeout / DefaultPollingInterval
}

func TestDefaultRetryConfig(t *testing.T) {
	// Act
	config := defaultRetryConfig()

	// Assert
	assert.NotNil(t, config)
	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, 1*time.Second, config.BaseDelay)
	assert.Equal(t, 2.0, config.Multiplier)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
}

func TestMergeExecutionOptions(t *testing.T) {
	tests := []struct {
		name        string
		options     *ExecutionOptions
		expected    *ExecutionOptions
		description string
	}{
		{
			name:     "NilOptions",
			options:  nil,
			expected: &ExecutionOptions{
				Timeout:    5 * time.Minute, // DefaultTimeout
				MaxRetries: 3,               // DefaultRetryAttempts
				RetryDelay: 1 * time.Second, // DefaultRetryDelay
			},
			description: "Nil options should return default options",
		},
		{
			name: "PartialOptions",
			options: &ExecutionOptions{
				TraceHeader: "trace-test",
			},
			expected: &ExecutionOptions{
				TraceHeader: "trace-test",
				Timeout:     5 * time.Minute, // Default filled in
				MaxRetries:  3,               // Default filled in
				RetryDelay:  1 * time.Second, // Default filled in
			},
			description: "Partial options should be merged with defaults",
		},
		{
			name: "CompleteOptions",
			options: &ExecutionOptions{
				TraceHeader: "trace-123",
				Timeout:     5 * time.Minute,
				MaxRetries:  3,
				RetryDelay:  2 * time.Second,
			},
			expected: &ExecutionOptions{
				TraceHeader: "trace-123",
				Timeout:     5 * time.Minute,
				MaxRetries:  3,
				RetryDelay:  2 * time.Second,
			},
			description: "Complete options should be preserved",
		},
		{
			name: "EmptyOptions",
			options: &ExecutionOptions{},
			expected: &ExecutionOptions{
				Timeout:    5 * time.Minute, // Default filled in
				MaxRetries: 3,               // Default filled in
				RetryDelay: 1 * time.Second, // Default filled in
			},
			description: "Empty options should be handled correctly",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := mergeExecutionOptions(tc.options)

			// Assert
			require.NotNil(t, result)
			assert.Equal(t, tc.expected.TraceHeader, result.TraceHeader)
			assert.Equal(t, tc.expected.Timeout, result.Timeout)
			assert.Equal(t, tc.expected.MaxRetries, result.MaxRetries)
			assert.Equal(t, tc.expected.RetryDelay, result.RetryDelay)
		})
	}
}

func TestMergeWaitOptions(t *testing.T) {
	tests := []struct {
		name        string
		options     *WaitOptions
		expected    *WaitOptions
		description string
	}{
		{
			name:     "NilOptions",
			options:  nil,
			expected: &WaitOptions{
				Timeout:         5 * time.Minute,
				PollingInterval: 5 * time.Second,
				MaxAttempts:     60,
			},
			description: "Nil options should return default options",
		},
		{
			name: "PartialOptions",
			options: &WaitOptions{
				Timeout: 10 * time.Minute,
			},
			expected: &WaitOptions{
				Timeout:         10 * time.Minute,
				PollingInterval: 5 * time.Second, // default
				MaxAttempts:     60,               // default
			},
			description: "Partial options should be merged with defaults",
		},
		{
			name: "CompleteOptions",
			options: &WaitOptions{
				Timeout:         3 * time.Minute,
				PollingInterval: 5 * time.Second,
				MaxAttempts:     20,
			},
			expected: &WaitOptions{
				Timeout:         3 * time.Minute,
				PollingInterval: 5 * time.Second,
				MaxAttempts:     20,
			},
			description: "Complete options should be preserved",
		},
		{
			name: "ZeroValues",
			options: &WaitOptions{
				Timeout:         0,
				PollingInterval: 0,
				MaxAttempts:     0,
			},
			expected: &WaitOptions{
				Timeout:         5 * time.Minute, // default
				PollingInterval: 5 * time.Second, // default
				MaxAttempts:     60,               // default
			},
			description: "Zero values should be replaced with defaults",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := mergeWaitOptions(tc.options)

			// Assert
			require.NotNil(t, result)
			assert.Equal(t, tc.expected.Timeout, result.Timeout)
			assert.Equal(t, tc.expected.PollingInterval, result.PollingInterval)
			assert.Equal(t, tc.expected.MaxAttempts, result.MaxAttempts)
		})
	}
}

func TestCalculateBackoffDelay(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Second,
		Multiplier:  2.0,
		MaxDelay:    30 * time.Second,
	}

	tests := []struct {
		name         string
		attemptCount int
		expected     time.Duration
		description  string
	}{
		{
			name:         "FirstAttempt",
			attemptCount: 1,
			expected:     2 * time.Second, // BaseDelay * Multiplier^1 = 1 * 2^1
			description:  "First attempt should apply backoff",
		},
		{
			name:         "SecondAttempt",
			attemptCount: 2,
			expected:     4 * time.Second, // BaseDelay * Multiplier^2 = 1 * 2^2
			description:  "Second attempt should apply exponential backoff",
		},
		{
			name:         "ThirdAttempt",
			attemptCount: 3,
			expected:     8 * time.Second, // BaseDelay * Multiplier^3 = 1 * 2^3
			description:  "Third attempt should apply exponential backoff",
		},
		{
			name:         "MaxDelayReached",
			attemptCount: 10,
			expected:     30 * time.Second, // Capped at MaxDelay
			description:  "Very high attempts should be capped at max delay",
		},
		{
			name:         "ZeroAttempt",
			attemptCount: 0,
			expected:     1 * time.Second, // Should return BaseDelay for attempt <= 0
			description:  "Zero attempt should return base delay",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := calculateBackoffDelay(tc.attemptCount, *config)

			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

// TestPollUntilComplete removed to avoid duplication with waiting_test.go

// Tests for history processing functions with missing coverage

func TestConvertAwsEventToHistoryEvent_TaskStateEntered(t *testing.T) {
	// Test the TaskStateEntered case that shows 0% coverage
	awsEvent := types.HistoryEvent{
		Id:        1,
		Timestamp: timePtr(time.Now()),
		Type:      types.HistoryEventTypeTaskStateEntered,
		StateEnteredEventDetails: &types.StateEnteredEventDetails{
			Name:  stringPtr("TestState"),
			Input: stringPtr(`{"test": "value"}`),
		},
	}

	// Act
	result := convertAwsEventToHistoryEvent(awsEvent)

	// Assert
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, types.HistoryEventTypeTaskStateEntered, result.Type)
	require.NotNil(t, result.StateEnteredEventDetails)
	assert.Equal(t, "TestState", result.StateEnteredEventDetails.Name)
	assert.Equal(t, `{"test": "value"}`, result.StateEnteredEventDetails.Input)
}

func TestConvertAwsEventToHistoryEvent_ExecutionFailed(t *testing.T) {
	// Test the ExecutionFailed case that shows 0% coverage
	awsEvent := types.HistoryEvent{
		Id:        2,
		Timestamp: timePtr(time.Now()),
		Type:      types.HistoryEventTypeExecutionFailed,
		ExecutionFailedEventDetails: &types.ExecutionFailedEventDetails{
			Error: stringPtr("States.TaskFailed"),
			Cause: stringPtr("Task execution failed"),
		},
	}

	// Act
	result := convertAwsEventToHistoryEvent(awsEvent)

	// Assert
	assert.Equal(t, int64(2), result.ID)
	assert.Equal(t, types.HistoryEventTypeExecutionFailed, result.Type)
	require.NotNil(t, result.ExecutionFailedEventDetails)
	assert.Equal(t, "States.TaskFailed", result.ExecutionFailedEventDetails.Error)
	assert.Equal(t, "Task execution failed", result.ExecutionFailedEventDetails.Cause)
}

func TestExtractStepNameFromEvent_TaskStateEnteredEventDetails(t *testing.T) {
	// Test the TaskStateEnteredEventDetails branch that shows 0% coverage
	event := HistoryEvent{
		Type: types.HistoryEventTypeTaskStateEntered,
		TaskStateEnteredEventDetails: &TaskStateEnteredEventDetails{
			Name: "TestTaskState",
		},
	}

	// Act
	stepName := extractStepNameFromEvent(event)

	// Assert
	assert.Equal(t, "TestTaskState", stepName)
}

func TestFormatEventDescription_TaskStateEntered_NoDetails(t *testing.T) {
	// Test the fallback case when StateEnteredEventDetails is nil
	event := HistoryEvent{
		Type: types.HistoryEventTypeTaskStateEntered,
		StateEnteredEventDetails: nil, // This should trigger the fallback
	}

	// Act
	description := formatEventDescription(event)

	// Assert
	assert.Equal(t, "Task state entered", description)
}

func TestFormatEventDescription_TaskTimedOut(t *testing.T) {
	// Test the TaskTimedOut case that shows 0% coverage
	event := HistoryEvent{
		Type: types.HistoryEventTypeTaskTimedOut,
	}

	// Act
	description := formatEventDescription(event)

	// Assert
	assert.Equal(t, "Task timed out", description)
}

// TestGetEventDetails removed - function is not public

// Tests for validation functions edge cases

func TestValidateStartExecutionRequest_StateMachineArnValidation(t *testing.T) {
	// Test the validateStateMachineArn call that shows 0% coverage
	request := &StartExecutionRequest{
		StateMachineArn: "invalid-arn-format", // This should trigger ARN validation
		Name:            "valid-execution-name",
		Input:           "{}",
	}

	// Act
	err := validateStartExecutionRequest(request)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ARN")
}

func TestValidateExecuteAndWaitPattern_StateMachineArnValidation(t *testing.T) {
	// Test the validateStateMachineArn call that shows 0% coverage
	err := validateExecuteAndWaitPattern(
		"invalid-arn-format", // This should trigger ARN validation
		"valid-execution-name",
		nil, // Use nil input to simplify
		5*time.Minute,
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ARN")
}

func TestValidateExecuteAndWaitPattern_ExecutionNameValidation(t *testing.T) {
	// Test the validateExecutionName call that shows 0% coverage
	err := validateExecuteAndWaitPattern(
		"arn:aws:states:us-east-1:123456789012:stateMachine:valid-machine",
		"", // Empty name should trigger validation
		nil, // Use nil input to simplify
		5*time.Minute,
	)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution name is required")
}

// Tests for execution result processing functions - simplified to avoid duplicates

// Helper functions for tests

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int64) *int64 {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Benchmark tests for utility functions

func BenchmarkMergeExecutionOptions(b *testing.B) {
	options := &ExecutionOptions{
		TraceHeader: "trace-123",
		Timeout:     5 * time.Minute,
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mergeExecutionOptions(options)
	}
}

func BenchmarkMergeWaitOptions(b *testing.B) {
	options := &WaitOptions{
		Timeout:         3 * time.Minute,
		PollingInterval: 5 * time.Second,
		MaxAttempts:     20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mergeWaitOptions(options)
	}
}

func BenchmarkCalculateBackoffDelay(b *testing.B) {
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Second,
		Multiplier:  2.0,
		MaxDelay:    30 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateBackoffDelay(5, config)
	}
}

// Tests for specific Input methods and validation functions with 0% coverage

func TestInputSetIfCoverage(t *testing.T) {
	input := NewInput()
	
	// Test SetIf with true condition (correct parameter order)
	result := input.SetIf(true, "key1", "value1")
	assert.Equal(t, input, result, "SetIf should return same instance")
	value, exists := input.data["key1"]
	assert.True(t, exists)
	assert.Equal(t, "value1", value)
	
	// Test SetIf with false condition
	input.SetIf(false, "key2", "value2")
	_, exists = input.data["key2"]
	assert.False(t, exists, "Key should not be set when condition is false")
}

func TestInputMergeCoverage(t *testing.T) {
	input1 := NewInput().Set("key1", "value1")
	input2 := NewInput().Set("key2", "value2").Set("key1", "overwritten")
	
	result := input1.Merge(input2)
	assert.Equal(t, input1, result, "Merge should return same instance")
	
	// Check merged values
	assert.Equal(t, "overwritten", input1.data["key1"], "key1 should be overwritten")
	assert.Equal(t, "value2", input1.data["key2"], "key2 should be merged")
}

func TestInputGettersCoverage(t *testing.T) {
	input := NewInput().
		Set("string_key", "test_value").
		Set("int_key", 42).
		Set("bool_key", true)
	
	// Test Get (returns interface{}, bool)
	value, exists := input.Get("string_key")
	assert.True(t, exists)
	assert.Equal(t, "test_value", value)
	
	// Test GetString (returns string, bool)
	strValue, exists := input.GetString("string_key")
	assert.True(t, exists)
	assert.Equal(t, "test_value", strValue)
	
	// Test GetInt (returns int, bool)
	intValue, exists := input.GetInt("int_key")
	assert.True(t, exists)
	assert.Equal(t, 42, intValue)
	
	// Test GetBool (returns bool, bool)
	boolValue, exists := input.GetBool("bool_key")
	assert.True(t, exists)
	assert.True(t, boolValue)
	
	// Test non-existent key
	_, exists = input.Get("non_existent")
	assert.False(t, exists)
}

func TestInputFactoryFunctionsCoverage(t *testing.T) {
	// Test CreateOrderInput - func CreateOrderInput(orderID, customerID string, items []string) *Input
	orderInput := CreateOrderInput("12345", "customer123", []string{"item1", "item2"})
	json, err := orderInput.ToJSON()
	assert.NoError(t, err)
	assert.Contains(t, json, "12345")
	assert.Contains(t, json, "customer123")
	assert.Contains(t, json, "item1")
	
	// Test CreateWorkflowInput - func CreateWorkflowInput(workflowName string, parameters map[string]interface{}) *Input
	workflowInput := CreateWorkflowInput("workflow1", map[string]interface{}{"key": "value"})
	json, err = workflowInput.ToJSON()
	assert.NoError(t, err)
	assert.Contains(t, json, "workflow1")
	assert.Contains(t, json, "value")
	
	// Test CreateRetryableInput - func CreateRetryableInput(operation string, maxRetries int) *Input
	retryableInput := CreateRetryableInput("task1", 3)
	json, err = retryableInput.ToJSON()
	assert.NoError(t, err)
	assert.Contains(t, json, "task1")
	assert.Contains(t, json, "3")
}

// Tests for validation functions (focusing on functions not covered elsewhere)

func TestValidateStateMachineNameCoverage(t *testing.T) {
	// Test valid name
	err := validateStateMachineName("ValidStateMachine")
	assert.NoError(t, err, "Valid state machine name should pass")
	
	// Test empty name
	err = validateStateMachineName("")
	assert.Error(t, err, "Empty name should fail validation")
	
	// Test too long name
	longName := "ThisIsAVeryLongStateMachineNameThatExceedsTheMaximumAllowedLengthForAStateMachineName"
	err = validateStateMachineName(longName)
	assert.Error(t, err, "Names over 80 characters should fail")
}

func TestValidateStateMachineDefinitionCoverage(t *testing.T) {
	// Test valid JSON
	validDefinition := `{"Comment":"A Hello World example","StartAt":"HelloWorld","States":{"HelloWorld":{"Type":"Pass","Result":"Hello World!","End":true}}}`
	err := validateStateMachineDefinition(validDefinition)
	assert.NoError(t, err, "Valid JSON definition should pass")
	
	// Test empty definition
	err = validateStateMachineDefinition("")
	assert.Error(t, err, "Empty definition should fail")
	
	// Test invalid JSON
	err = validateStateMachineDefinition(`{"Invalid":"JSON"`)
	assert.Error(t, err, "Invalid JSON should fail")
}

// Tests for more validation functions with 0% coverage

func TestValidateRoleArnCoverage(t *testing.T) {
	// Test valid role ARN
	err := validateRoleArn("arn:aws:iam::123456789012:role/StepFunctionsRole")
	assert.NoError(t, err, "Valid role ARN should pass")
	
	// Test empty role ARN
	err = validateRoleArn("")
	assert.Error(t, err, "Empty role ARN should fail")
	
	// Test invalid role ARN
	err = validateRoleArn("invalid-arn")
	assert.Error(t, err, "Invalid role ARN should fail")
}

func TestValidateMaxResultsCoverage(t *testing.T) {
	// Test valid max results
	err := validateMaxResults(100)
	assert.NoError(t, err, "Valid max results should pass")
	
	// Test zero max results (should be allowed in this signature)
	err = validateMaxResults(0)
	assert.NoError(t, err, "Zero max results should pass")
	
	// Test negative max results
	err = validateMaxResults(-1)
	assert.Error(t, err, "Negative max results should fail")
	
	// Test too high max results
	err = validateMaxResults(1001)
	assert.Error(t, err, "Max results above 1000 should fail")
}

func TestValidateExecutionArnCoverage(t *testing.T) {
	// Test valid execution ARN
	err := validateExecutionArn("arn:aws:states:us-east-1:123456789012:execution:test-machine:test-exec")
	assert.NoError(t, err, "Valid execution ARN should pass")
	
	// Test empty execution ARN
	err = validateExecutionArn("")
	assert.Error(t, err, "Empty execution ARN should fail")
	
	// Test invalid execution ARN
	err = validateExecutionArn("invalid-arn")
	assert.Error(t, err, "Invalid execution ARN should fail")
}

func TestParseExecutionArnCoverage(t *testing.T) {
	// Test valid execution ARN
	stateMachine, executionName, err := parseExecutionArn("arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution")
	assert.NoError(t, err)
	assert.Equal(t, "test-machine", stateMachine)
	assert.Equal(t, "test-execution", executionName)
	
	// Test empty ARN
	_, _, err = parseExecutionArn("")
	assert.Error(t, err, "Empty ARN should fail")
	
	// Test invalid ARN format
	_, _, err = parseExecutionArn("invalid-arn-format")
	assert.Error(t, err, "Invalid ARN format should fail")
}

func TestIsExecutionCompleteCoverage(t *testing.T) {
	// Test completed statuses
	assert.True(t, isExecutionComplete(types.ExecutionStatusSucceeded))
	assert.True(t, isExecutionComplete(types.ExecutionStatusFailed))
	assert.True(t, isExecutionComplete(types.ExecutionStatusTimedOut))
	assert.True(t, isExecutionComplete(types.ExecutionStatusAborted))
	
	// Test non-completed status
	assert.False(t, isExecutionComplete(types.ExecutionStatusRunning))
}

func TestCalculateExecutionTimeCoverage(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)
	
	// Test with both start and end times
	duration := calculateExecutionTime(startTime, endTime)
	assert.Equal(t, 5*time.Minute, duration)
	
	// Test with zero end time
	duration = calculateExecutionTime(startTime, time.Time{})
	assert.Equal(t, time.Duration(0), duration)
}

func TestCalculateNextPollIntervalCoverage(t *testing.T) {
	config := &PollConfig{
		Interval:           2 * time.Second,
		ExponentialBackoff: false,
	}
	
	// Test without exponential backoff
	interval := calculateNextPollInterval(1, config)
	assert.Equal(t, 2*time.Second, interval)
	
	// Test with exponential backoff
	config.ExponentialBackoff = true
	config.BackoffMultiplier = 2.0
	config.MaxInterval = 60 * time.Second
	interval = calculateNextPollInterval(2, config)
	assert.Equal(t, 4*time.Second, interval) // 2 * 2.0
}

func TestShouldContinuePollingCoverage(t *testing.T) {
	config := &PollConfig{
		MaxAttempts: 10,
		Timeout:     60 * time.Second,
	}
	
	// Test within limits
	result := shouldContinuePolling(5, 30*time.Second, config)
	assert.True(t, result)
	
	// Test exceeded max attempts
	result = shouldContinuePolling(11, 30*time.Second, config)
	assert.False(t, result)
	
	// Test exceeded timeout
	result = shouldContinuePolling(5, 65*time.Second, config)
	assert.False(t, result)
}

func TestValidatePollConfigCoverage(t *testing.T) {
	// Test valid config
	err := validatePollConfig("arn:aws:states:us-east-1:123456789012:execution:test:exec", &PollConfig{
		MaxAttempts:        10,
		Interval:           5 * time.Second,
		Timeout:            60 * time.Second,
		ExponentialBackoff: true,
		BackoffMultiplier:  2.0,
		MaxInterval:        30 * time.Second,
	})
	assert.NoError(t, err)
	
	// Test empty execution ARN
	err = validatePollConfig("", &PollConfig{
		MaxAttempts:        10,
		Interval:           5 * time.Second,
		Timeout:            60 * time.Second,
		ExponentialBackoff: false,
		BackoffMultiplier:  1.0,
		MaxInterval:        30 * time.Second,
	})
	assert.Error(t, err)
	
	// Test nil config (should pass - uses defaults)
	err = validatePollConfig("arn:aws:states:us-east-1:123456789012:execution:test:exec", nil)
	assert.NoError(t, err)
}

func TestValidateStateMachineStatusCoverage(t *testing.T) {
	// Test matching statuses
	result := validateStateMachineStatus(types.StateMachineStatusActive, types.StateMachineStatusActive)
	assert.True(t, result)
	
	// Test non-matching statuses
	result = validateStateMachineStatus(types.StateMachineStatusActive, types.StateMachineStatusDeleting)
	assert.False(t, result)
}

func TestProcessStateMachineInfoCoverage(t *testing.T) {
	// Test with nil input
	result := processStateMachineInfo(nil)
	assert.Nil(t, result)
	
	// Test with valid input
	creationDate := time.Now()
	info := &StateMachineInfo{
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
		Name:           "test-machine",
		Definition:     `{"Comment":"Test"}`,
		RoleArn:        "arn:aws:iam::123456789012:role/TestRole",
		Type:          types.StateMachineTypeStandard,
		Status:        types.StateMachineStatusActive,
		CreationDate:  creationDate,
	}
	
	result = processStateMachineInfo(info)
	require.NotNil(t, result)
	assert.Equal(t, info.Name, result.Name)
}

func TestProcessExecutionResultCoverage(t *testing.T) {
	// Test with nil result
	result := processExecutionResult(nil)
	assert.Nil(t, result)
	
	// Test with valid execution result
	startDate := time.Now().Add(-5 * time.Minute)
	stopDate := time.Now()
	
	execResult := &ExecutionResult{
		ExecutionArn:     "arn:aws:states:us-east-1:123456789012:execution:test:exec",
		StateMachineArn:  "arn:aws:states:us-east-1:123456789012:stateMachine:test",
		Name:             "exec",
		Status:           types.ExecutionStatusSucceeded,
		StartDate:        startDate,
		StopDate:         stopDate,
		Input:            `{"test": "value"}`,
		Output:           `{"result": "success"}`,
	}

	result = processExecutionResult(execResult)
	require.NotNil(t, result)
	assert.Equal(t, execResult.Name, result.Name)
}

func TestValidateStopExecutionRequestCoverage(t *testing.T) {
	// Test valid request
	err := validateStopExecutionRequest(
		"arn:aws:states:us-east-1:123456789012:execution:test-machine:test-exec",
		"User requested stop",
		"Manual intervention",
	)
	assert.NoError(t, err)
	
	// Test empty execution ARN
	err = validateStopExecutionRequest("", "", "")
	assert.Error(t, err)
}

func TestValidateWaitForExecutionRequestCoverage(t *testing.T) {
	// Test valid request
	err := validateWaitForExecutionRequest(
		"arn:aws:states:us-east-1:123456789012:execution:test-machine:test-exec",
		&WaitOptions{
			Timeout:         5 * time.Minute,
			PollingInterval: 5 * time.Second,
			MaxAttempts:     60,
		},
	)
	assert.NoError(t, err)
	
	// Test empty execution ARN
	err = validateWaitForExecutionRequest("", nil)
	assert.Error(t, err)
	
	// Test nil options (should pass - uses defaults)
	err = validateWaitForExecutionRequest(
		"arn:aws:states:us-east-1:123456789012:execution:test-machine:test-exec",
		nil,
	)
	assert.NoError(t, err)
}

// Additional utility functions

func int32Ptr(i int32) *int32 {
	return &i
}

// Benchmark removed to avoid duplication