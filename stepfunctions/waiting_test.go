package stepfunctions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test data factories for waiting operations

func createTestPollConfig() *PollConfig {
	return &PollConfig{
		MaxAttempts:     30,
		Interval:        2 * time.Second,
		Timeout:         60 * time.Second,
		ExponentialBackoff: true,
		BackoffMultiplier:  1.5,
		MaxInterval:        30 * time.Second,
	}
}

// Table-driven test structures for waiting operations

type waitForExecutionCompletionTestCase struct {
	name                 string
	executionArn         string
	waitOptions          *WaitOptions
	shouldError          bool
	expectedErrorMessage string
	description          string
}

type pollUntilCompleteTestCase struct {
	name                 string
	executionArn         string
	config               *PollConfig
	shouldError          bool
	expectedErrorMessage string
	description          string
}

// Unit tests for waiting functions

func TestWaitForExecutionCompletion(t *testing.T) {
	tests := []waitForExecutionCompletionTestCase{
		{
			name:         "ValidWaitForCompletion",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			waitOptions:  createTestWaitOptions(),
			shouldError:  false,
			description:  "Valid wait for execution completion should pass validation",
		},
		{
			name:                 "EmptyExecutionArn",
			executionArn:         "",
			waitOptions:          createTestWaitOptions(),
			shouldError:          true,
			expectedErrorMessage: "execution ARN is required",
			description:          "Empty execution ARN should fail validation",
		},
		{
			name:                 "InvalidExecutionArn",
			executionArn:         "invalid-arn",
			waitOptions:          createTestWaitOptions(),
			shouldError:          true,
			expectedErrorMessage: "invalid ARN",
			description:          "Invalid execution ARN should fail validation",
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

func TestPollUntilComplete(t *testing.T) {
	tests := []pollUntilCompleteTestCase{
		{
			name:         "ValidPollRequest",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			config:       createTestPollConfig(),
			shouldError:  false,
			description:  "Valid poll request should pass validation",
		},
		{
			name:                 "EmptyExecutionArn",
			executionArn:         "",
			config:               createTestPollConfig(),
			shouldError:          true,
			expectedErrorMessage: "execution ARN is required",
			description:          "Empty execution ARN should fail validation",
		},
		{
			name:                 "InvalidExecutionArn",
			executionArn:         "not-valid-arn",
			config:               createTestPollConfig(),
			shouldError:          true,
			expectedErrorMessage: "invalid ARN",
			description:          "Invalid execution ARN should fail validation",
		},
		{
			name:         "NilConfig",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			config:       nil,
			shouldError:  false,
			description:  "Nil config should use defaults",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validatePollConfig(tc.executionArn, tc.config)
			
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

// Tests for poll configuration and utilities

func TestCalculateNextPollInterval(t *testing.T) {
	tests := []struct {
		name              string
		currentAttempt    int
		config            *PollConfig
		expectedMin       time.Duration
		expectedMax       time.Duration
		description       string
	}{
		{
			name:           "FirstAttempt",
			currentAttempt: 1,
			config:         createTestPollConfig(),
			expectedMin:    2 * time.Second,
			expectedMax:    2 * time.Second,
			description:    "First attempt should use base interval",
		},
		{
			name:           "SecondAttemptWithBackoff",
			currentAttempt: 2,
			config: &PollConfig{
				Interval:           2 * time.Second,
				ExponentialBackoff: true,
				BackoffMultiplier:  2.0,
				MaxInterval:        60 * time.Second,
			},
			expectedMin: 4 * time.Second, // 2 * 2.0
			expectedMax: 4 * time.Second,
			description: "Second attempt with exponential backoff should double",
		},
		{
			name:           "MaxIntervalCap",
			currentAttempt: 10,
			config: &PollConfig{
				Interval:           2 * time.Second,
				ExponentialBackoff: true,
				BackoffMultiplier:  2.0,
				MaxInterval:        10 * time.Second, // Low max to test capping
			},
			expectedMin: 0 * time.Second, // Will be capped
			expectedMax: 10 * time.Second, // Capped at max
			description: "High attempt count should be capped at max interval",
		},
		{
			name:           "NoExponentialBackoff",
			currentAttempt: 5,
			config: &PollConfig{
				Interval:           3 * time.Second,
				ExponentialBackoff: false,
				BackoffMultiplier:  2.0,
				MaxInterval:        60 * time.Second,
			},
			expectedMin: 3 * time.Second,
			expectedMax: 3 * time.Second,
			description: "Without exponential backoff should use constant interval",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			interval := calculateNextPollInterval(tc.currentAttempt, tc.config)
			
			// Assert
			assert.GreaterOrEqual(t, interval, tc.expectedMin, tc.description)
			assert.LessOrEqual(t, interval, tc.expectedMax, tc.description)
		})
	}
}

func TestShouldContinuePolling(t *testing.T) {
	tests := []struct {
		name           string
		attempt        int
		elapsed        time.Duration
		config         *PollConfig
		expected       bool
		description    string
	}{
		{
			name:    "WithinLimits",
			attempt: 5,
			elapsed: 30 * time.Second,
			config: &PollConfig{
				MaxAttempts: 10,
				Timeout:     60 * time.Second,
			},
			expected:    true,
			description: "Should continue when within both attempt and time limits",
		},
		{
			name:    "ExceededMaxAttempts",
			attempt: 11,
			elapsed: 30 * time.Second,
			config: &PollConfig{
				MaxAttempts: 10,
				Timeout:     60 * time.Second,
			},
			expected:    false,
			description: "Should stop when max attempts exceeded",
		},
		{
			name:    "ExceededTimeout",
			attempt: 5,
			elapsed: 65 * time.Second,
			config: &PollConfig{
				MaxAttempts: 10,
				Timeout:     60 * time.Second,
			},
			expected:    false,
			description: "Should stop when timeout exceeded",
		},
		{
			name:    "ExceededBothLimits",
			attempt: 11,
			elapsed: 65 * time.Second,
			config: &PollConfig{
				MaxAttempts: 10,
				Timeout:     60 * time.Second,
			},
			expected:    false,
			description: "Should stop when both limits exceeded",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := shouldContinuePolling(tc.attempt, tc.elapsed, tc.config)
			
			// Assert
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestExecuteStateMachineAndWaitPattern(t *testing.T) {
	// Arrange
	stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine"
	executionName := "test-execution"
	input := NewInput().Set("message", "test")
	timeout := 60 * time.Second
	
	// Act - This tests the pattern validation, not actual execution
	err := validateExecuteAndWaitPattern(stateMachineArn, executionName, input, timeout)
	
	// Assert
	assert.NoError(t, err, "Valid execute and wait pattern should pass validation")
}

func TestExecuteStateMachineAndWaitPatternInvalid(t *testing.T) {
	tests := []struct {
		name                 string
		stateMachineArn      string
		executionName        string
		input                *Input
		timeout              time.Duration
		expectedErrorMessage string
		description          string
	}{
		{
			name:                 "EmptyStateMachineArn",
			stateMachineArn:      "",
			executionName:        "test-execution",
			input:                NewInput(),
			timeout:              60 * time.Second,
			expectedErrorMessage: "state machine ARN is required",
			description:          "Empty state machine ARN should fail",
		},
		{
			name:                 "EmptyExecutionName",
			stateMachineArn:      "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			executionName:        "",
			input:                NewInput(),
			timeout:              60 * time.Second,
			expectedErrorMessage: "execution name is required",
			description:          "Empty execution name should fail",
		},
		{
			name:                 "ZeroTimeout",
			stateMachineArn:      "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			executionName:        "test-execution",
			input:                NewInput(),
			timeout:              0,
			expectedErrorMessage: "timeout must be positive",
			description:          "Zero timeout should fail",
		},
		{
			name:                 "NegativeTimeout",
			stateMachineArn:      "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			executionName:        "test-execution",
			input:                NewInput(),
			timeout:              -60 * time.Second,
			expectedErrorMessage: "timeout must be positive",
			description:          "Negative timeout should fail",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateExecuteAndWaitPattern(tc.stateMachineArn, tc.executionName, tc.input, tc.timeout)
			
			// Assert
			assert.Error(t, err, tc.description)
			assert.Contains(t, err.Error(), tc.expectedErrorMessage, tc.description)
		})
	}
}

// Benchmark tests for polling and waiting operations

func BenchmarkCalculateNextPollInterval(b *testing.B) {
	config := createTestPollConfig()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateNextPollInterval(5, config)
	}
}

func BenchmarkShouldContinuePolling(b *testing.B) {
	config := createTestPollConfig()
	elapsed := 30 * time.Second
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = shouldContinuePolling(5, elapsed, config)
	}
}