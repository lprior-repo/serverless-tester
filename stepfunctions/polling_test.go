package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

// TestValidatePollConfig tests poll configuration validation
func TestValidatePollConfig(t *testing.T) {
	validArn := "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution"

	tests := []struct {
		name          string
		executionArn  string
		config        *PollConfig
		expectedError bool
		errorContains string
		description   string
	}{
		{
			name:         "ValidConfig",
			executionArn: validArn,
			config: &PollConfig{
				MaxAttempts:        10,
				Interval:           5 * time.Second,
				Timeout:            5 * time.Minute,
				ExponentialBackoff: true,
				BackoffMultiplier:  1.5,
				MaxInterval:        30 * time.Second,
			},
			expectedError: false,
			description:   "Valid poll config should pass validation",
		},
		{
			name:          "EmptyExecutionArn",
			executionArn:  "",
			config:        nil,
			expectedError: true,
			errorContains: "execution ARN is required",
			description:   "Empty execution ARN should fail validation",
		},
		{
			name:         "NilConfigAllowed",
			executionArn: validArn,
			config:       nil,
			expectedError: false,
			description:   "Nil config should be allowed (defaults will be used)",
		},
		{
			name:         "NegativeMaxAttempts",
			executionArn: validArn,
			config: &PollConfig{
				MaxAttempts: -1,
			},
			expectedError: true,
			errorContains: "max attempts cannot be negative",
			description:   "Negative max attempts should fail validation",
		},
		{
			name:         "NegativeInterval",
			executionArn: validArn,
			config: &PollConfig{
				Interval: -1 * time.Second,
			},
			expectedError: true,
			errorContains: "interval cannot be negative",
			description:   "Negative interval should fail validation",
		},
		{
			name:         "NegativeTimeout",
			executionArn: validArn,
			config: &PollConfig{
				Timeout: -1 * time.Minute,
			},
			expectedError: true,
			errorContains: "timeout cannot be negative",
			description:   "Negative timeout should fail validation",
		},
		{
			name:         "InvalidBackoffMultiplier",
			executionArn: validArn,
			config: &PollConfig{
				BackoffMultiplier: 0.5, // Must be >= 1.0
			},
			expectedError: true,
			errorContains: "backoff multiplier must be >= 1.0",
			description:   "Backoff multiplier < 1.0 should fail validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validatePollConfig(tc.executionArn, tc.config)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// TestCalculateNextPollInterval tests polling interval calculation
func TestCalculateNextPollInterval(t *testing.T) {
	tests := []struct {
		name            string
		currentAttempt  int
		config          *PollConfig
		expectedResult  time.Duration
		description     string
	}{
		{
			name:           "NilConfigUsesDefault",
			currentAttempt: 1,
			config:         nil,
			expectedResult: DefaultPollingInterval,
			description:    "Nil config should use default polling interval",
		},
		{
			name:           "NoExponentialBackoff",
			currentAttempt: 3,
			config: &PollConfig{
				Interval:           10 * time.Second,
				ExponentialBackoff: false,
			},
			expectedResult: 10 * time.Second,
			description:    "No exponential backoff should return fixed interval",
		},
		{
			name:           "ExponentialBackoffFirstAttempt",
			currentAttempt: 1,
			config: &PollConfig{
				Interval:           5 * time.Second,
				ExponentialBackoff: true,
				BackoffMultiplier:  2.0,
			},
			expectedResult: 5 * time.Second,
			description:    "First attempt should return base interval",
		},
		{
			name:           "ExponentialBackoffSecondAttempt",
			currentAttempt: 2,
			config: &PollConfig{
				Interval:           5 * time.Second,
				ExponentialBackoff: true,
				BackoffMultiplier:  2.0,
			},
			expectedResult: 10 * time.Second,
			description:    "Second attempt should double the interval",
		},
		{
			name:           "ExponentialBackoffThirdAttempt",
			currentAttempt: 3,
			config: &PollConfig{
				Interval:           5 * time.Second,
				ExponentialBackoff: true,
				BackoffMultiplier:  2.0,
			},
			expectedResult: 20 * time.Second,
			description:    "Third attempt should quadruple the interval",
		},
		{
			name:           "ExponentialBackoffCappedAtMax",
			currentAttempt: 10,
			config: &PollConfig{
				Interval:           1 * time.Second,
				ExponentialBackoff: true,
				BackoffMultiplier:  2.0,
				MaxInterval:        30 * time.Second,
			},
			expectedResult: 30 * time.Second,
			description:    "High attempt should be capped at max interval",
		},
		{
			name:           "CustomBackoffMultiplier",
			currentAttempt: 2,
			config: &PollConfig{
				Interval:           4 * time.Second,
				ExponentialBackoff: true,
				BackoffMultiplier:  1.5,
			},
			expectedResult: 6 * time.Second,
			description:    "Custom multiplier should be applied correctly",
		},
		{
			name:           "ZeroIntervalUsesDefault",
			currentAttempt: 1,
			config: &PollConfig{
				Interval:           0, // Should use default
				ExponentialBackoff: true,
				BackoffMultiplier:  2.0,
			},
			expectedResult: DefaultPollingInterval,
			description:    "Zero interval should use default",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := calculateNextPollInterval(tc.currentAttempt, tc.config)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
		})
	}
}

// TestShouldContinuePolling tests polling continuation logic
func TestShouldContinuePolling(t *testing.T) {
	tests := []struct {
		name           string
		attempt        int
		elapsed        time.Duration
		config         *PollConfig
		expectedResult bool
		description    string
	}{
		{
			name:    "NilConfigUsesDefaults",
			attempt: 10,
			elapsed: 30 * time.Second,
			config:  nil,
			expectedResult: true, // Should continue with defaults
			description:    "Nil config should use default limits",
		},
		{
			name:    "WithinMaxAttempts",
			attempt: 5,
			elapsed: 30 * time.Second,
			config: &PollConfig{
				MaxAttempts: 10,
				Timeout:     5 * time.Minute,
			},
			expectedResult: true,
			description:    "Should continue when within max attempts",
		},
		{
			name:    "ExceedsMaxAttempts",
			attempt: 15,
			elapsed: 30 * time.Second,
			config: &PollConfig{
				MaxAttempts: 10,
				Timeout:     5 * time.Minute,
			},
			expectedResult: false,
			description:    "Should stop when exceeding max attempts",
		},
		{
			name:    "WithinTimeout",
			attempt: 5,
			elapsed: 2 * time.Minute,
			config: &PollConfig{
				MaxAttempts: 100,
				Timeout:     5 * time.Minute,
			},
			expectedResult: true,
			description:    "Should continue when within timeout",
		},
		{
			name:    "ExceedsTimeout",
			attempt: 5,
			elapsed: 6 * time.Minute,
			config: &PollConfig{
				MaxAttempts: 100,
				Timeout:     5 * time.Minute,
			},
			expectedResult: false,
			description:    "Should stop when exceeding timeout",
		},
		{
			name:    "BothLimitsExceeded",
			attempt: 15,
			elapsed: 6 * time.Minute,
			config: &PollConfig{
				MaxAttempts: 10,
				Timeout:     5 * time.Minute,
			},
			expectedResult: false,
			description:    "Should stop when both limits exceeded",
		},
		{
			name:    "ZeroMaxAttemptsIgnored",
			attempt: 100,
			elapsed: 1 * time.Minute,
			config: &PollConfig{
				MaxAttempts: 0, // Should be ignored
				Timeout:     5 * time.Minute,
			},
			expectedResult: true,
			description:    "Zero max attempts should be ignored",
		},
		{
			name:    "ZeroTimeoutIgnored",
			attempt: 5,
			elapsed: 10 * time.Minute,
			config: &PollConfig{
				MaxAttempts: 100,
				Timeout:     0, // Should be ignored
			},
			expectedResult: true,
			description:    "Zero timeout should be ignored",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := shouldContinuePolling(tc.attempt, tc.elapsed, tc.config)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
		})
	}
}

// TestDefaultPollConfig tests default polling configuration
func TestDefaultPollConfig(t *testing.T) {
	// Act
	config := defaultPollConfig()

	// Assert
	assert.NotNil(t, config, "Default config should not be nil")
	assert.Equal(t, int(DefaultTimeout/DefaultPollingInterval), config.MaxAttempts, "MaxAttempts should be calculated from defaults")
	assert.Equal(t, DefaultPollingInterval, config.Interval, "Interval should be default")
	assert.Equal(t, DefaultTimeout, config.Timeout, "Timeout should be default")
	assert.False(t, config.ExponentialBackoff, "ExponentialBackoff should be false by default")
	assert.Equal(t, 1.5, config.BackoffMultiplier, "BackoffMultiplier should have sensible default")
	assert.Equal(t, 30*time.Second, config.MaxInterval, "MaxInterval should have sensible default")
}

// TestMergePollConfig tests poll configuration merging
func TestMergePollConfig(t *testing.T) {
	tests := []struct {
		name        string
		userConfig  *PollConfig
		expected    *PollConfig
		description string
	}{
		{
			name:       "NilConfigUsesDefaults",
			userConfig: nil,
			expected:   defaultPollConfig(),
			description: "Nil config should return defaults",
		},
		{
			name: "PartialConfigMergesWithDefaults",
			userConfig: &PollConfig{
				MaxAttempts: 20, // Custom value
				// Other fields should use defaults
			},
			expected: &PollConfig{
				MaxAttempts:        20,
				Interval:           DefaultPollingInterval,
				Timeout:            DefaultTimeout,
				ExponentialBackoff: false,
				BackoffMultiplier:  1.5,
				MaxInterval:        30 * time.Second,
			},
			description: "Partial config should merge with defaults",
		},
		{
			name: "CompleteConfigPreserved",
			userConfig: &PollConfig{
				MaxAttempts:        50,
				Interval:           10 * time.Second,
				Timeout:            10 * time.Minute,
				ExponentialBackoff: true,
				BackoffMultiplier:  2.5,
				MaxInterval:        60 * time.Second,
			},
			expected: &PollConfig{
				MaxAttempts:        50,
				Interval:           10 * time.Second,
				Timeout:            10 * time.Minute,
				ExponentialBackoff: true,
				BackoffMultiplier:  2.5,
				MaxInterval:        60 * time.Second,
			},
			description: "Complete config should be preserved",
		},
		{
			name: "ZeroValuesReplacedWithDefaults",
			userConfig: &PollConfig{
				MaxAttempts:       0, // Should use default
				Interval:          0, // Should use default
				Timeout:           0, // Should use default
				BackoffMultiplier: 0, // Should use default
				MaxInterval:       0, // Should use default
			},
			expected: &PollConfig{
				MaxAttempts:        int(DefaultTimeout / DefaultPollingInterval),
				Interval:           DefaultPollingInterval,
				Timeout:            DefaultTimeout,
				ExponentialBackoff: false,
				BackoffMultiplier:  1.5,
				MaxInterval:        30 * time.Second,
			},
			description: "Zero values should be replaced with defaults",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := mergePollConfig(tc.userConfig)

			// Assert
			assert.NotNil(t, result, tc.description)
			assert.Equal(t, tc.expected.MaxAttempts, result.MaxAttempts, "MaxAttempts should match")
			assert.Equal(t, tc.expected.Interval, result.Interval, "Interval should match")
			assert.Equal(t, tc.expected.Timeout, result.Timeout, "Timeout should match")
			assert.Equal(t, tc.expected.ExponentialBackoff, result.ExponentialBackoff, "ExponentialBackoff should match")
			assert.Equal(t, tc.expected.BackoffMultiplier, result.BackoffMultiplier, "BackoffMultiplier should match")
			assert.Equal(t, tc.expected.MaxInterval, result.MaxInterval, "MaxInterval should match")
		})
	}
}

// TestValidateExecuteAndWaitPattern tests execute-and-wait pattern validation
func TestValidateExecuteAndWaitPattern(t *testing.T) {
	validArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine"

	tests := []struct {
		name            string
		stateMachineArn string
		executionName   string
		input           *Input
		timeout         time.Duration
		expectedError   bool
		errorContains   string
		description     string
	}{
		{
			name:            "ValidPattern",
			stateMachineArn: validArn,
			executionName:   "test-execution",
			input:           NewInput().Set("message", "test"),
			timeout:         5 * time.Minute,
			expectedError:   false,
			description:     "Valid execute-and-wait pattern should pass validation",
		},
		{
			name:            "EmptyStateMachineArn",
			stateMachineArn: "",
			executionName:   "test-execution",
			input:           NewInput(),
			timeout:         5 * time.Minute,
			expectedError:   true,
			errorContains:   "state machine ARN is required",
			description:     "Empty state machine ARN should fail validation",
		},
		{
			name:            "EmptyExecutionName",
			stateMachineArn: validArn,
			executionName:   "",
			input:           NewInput(),
			timeout:         5 * time.Minute,
			expectedError:   true,
			errorContains:   "execution name is required",
			description:     "Empty execution name should fail validation",
		},
		{
			name:            "ZeroTimeout",
			stateMachineArn: validArn,
			executionName:   "test-execution",
			input:           NewInput(),
			timeout:         0,
			expectedError:   true,
			errorContains:   "timeout must be positive",
			description:     "Zero timeout should fail validation",
		},
		{
			name:            "NegativeTimeout",
			stateMachineArn: validArn,
			executionName:   "test-execution",
			input:           NewInput(),
			timeout:         -1 * time.Minute,
			expectedError:   true,
			errorContains:   "timeout must be positive",
			description:     "Negative timeout should fail validation",
		},
		{
			name:            "NilInputAllowed",
			stateMachineArn: validArn,
			executionName:   "test-execution",
			input:           nil,
			timeout:         5 * time.Minute,
			expectedError:   false,
			description:     "Nil input should be allowed",
		},
		{
			name:            "EmptyInputAllowed",
			stateMachineArn: validArn,
			executionName:   "test-execution",
			input:           NewInput(), // Empty but not nil
			timeout:         5 * time.Minute,
			expectedError:   false,
			description:     "Empty input should be allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateExecuteAndWaitPattern(tc.stateMachineArn, tc.executionName, tc.input, tc.timeout)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// TestIsExecutionComplete tests execution completion checking
func TestIsExecutionCompletePolling(t *testing.T) {
	tests := []struct {
		name           string
		status         types.ExecutionStatus
		expectedResult bool
		description    string
	}{
		{
			name:           "RunningNotComplete",
			status:         types.ExecutionStatusRunning,
			expectedResult: false,
			description:    "Running execution should not be complete",
		},
		{
			name:           "SucceededIsComplete",
			status:         types.ExecutionStatusSucceeded,
			expectedResult: true,
			description:    "Succeeded execution should be complete",
		},
		{
			name:           "FailedIsComplete",
			status:         types.ExecutionStatusFailed,
			expectedResult: true,
			description:    "Failed execution should be complete",
		},
		{
			name:           "TimedOutIsComplete",
			status:         types.ExecutionStatusTimedOut,
			expectedResult: true,
			description:    "Timed out execution should be complete",
		},
		{
			name:           "AbortedIsComplete",
			status:         types.ExecutionStatusAborted,
			expectedResult: true,
			description:    "Aborted execution should be complete",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := isExecutionComplete(tc.status)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
		})
	}
}

// TestParseExecutionArn tests execution ARN parsing
func TestParseExecutionArnPolling(t *testing.T) {
	tests := []struct {
		name                    string
		executionArn            string
		expectedStateMachine    string
		expectedExecution       string
		expectedError           bool
		errorContains           string
		description             string
	}{
		{
			name:                 "ValidExecutionArn",
			executionArn:         "arn:aws:states:us-east-1:123456789012:execution:test-state-machine:test-execution",
			expectedStateMachine: "test-state-machine",
			expectedExecution:    "test-execution",
			expectedError:        false,
			description:          "Valid execution ARN should parse correctly",
		},
		{
			name:          "EmptyArn",
			executionArn:  "",
			expectedError: true,
			errorContains: "execution ARN is required",
			description:   "Empty ARN should fail parsing",
		},
		{
			name:          "InvalidArnFormat",
			executionArn:  "invalid-arn",
			expectedError: true,
			errorContains: "invalid execution ARN format",
			description:   "Invalid ARN format should fail parsing",
		},
		{
			name:          "WrongService",
			executionArn:  "arn:aws:lambda:us-east-1:123456789012:function:test",
			expectedError: true,
			errorContains: "invalid execution ARN format",
			description:   "Wrong service ARN should fail parsing",
		},
		{
			name:          "InsufficientParts",
			executionArn:  "arn:aws:states:us-east-1:123456789012",
			expectedError: true,
			errorContains: "invalid execution ARN format",
			description:   "ARN with insufficient parts should fail parsing",
		},
		{
			name:                 "ValidArnWithHyphens",
			executionArn:         "arn:aws:states:us-west-2:987654321098:execution:my-state-machine-test:my-execution-name-123",
			expectedStateMachine: "my-state-machine-test",
			expectedExecution:    "my-execution-name-123",
			expectedError:        false,
			description:          "Valid ARN with hyphens should parse correctly",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			stateMachine, execution, err := parseExecutionArn(tc.executionArn)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, tc.description)
				}
				assert.Empty(t, stateMachine, "State machine should be empty on error")
				assert.Empty(t, execution, "Execution should be empty on error")
			} else {
				assert.NoError(t, err, tc.description)
				assert.Equal(t, tc.expectedStateMachine, stateMachine, "State machine name should match")
				assert.Equal(t, tc.expectedExecution, execution, "Execution name should match")
			}
		})
	}
}

// TestProcessStateMachineInfo tests state machine info processing
func TestProcessStateMachineInfoPolling(t *testing.T) {
	tests := []struct {
		name        string
		input       *StateMachineInfo
		expected    *StateMachineInfo
		description string
	}{
		{
			name:        "NilInfo",
			input:       nil,
			expected:    nil,
			description: "Nil info should return nil",
		},
		{
			name: "InfoWithEmptyDefinition",
			input: &StateMachineInfo{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				Name:            "test",
				Status:          types.StateMachineStatusActive,
				Definition:      "", // Empty definition
				RoleArn:         "arn:aws:iam::123456789012:role/test",
				Type:            types.StateMachineTypeStandard,
			},
			expected: &StateMachineInfo{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				Name:            "test",
				Status:          types.StateMachineStatusActive,
				Definition:      "{}", // Should be set to empty object
				RoleArn:         "arn:aws:iam::123456789012:role/test",
				Type:            types.StateMachineTypeStandard,
			},
			description: "Empty definition should be replaced with empty JSON object",
		},
		{
			name: "InfoWithValidDefinition",
			input: &StateMachineInfo{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				Name:            "test",
				Status:          types.StateMachineStatusActive,
				Definition:      `{"StartAt": "Test", "States": {"Test": {"Type": "Pass", "End": true}}}`,
				RoleArn:         "arn:aws:iam::123456789012:role/test",
				Type:            types.StateMachineTypeStandard,
			},
			expected: &StateMachineInfo{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
				Name:            "test",
				Status:          types.StateMachineStatusActive,
				Definition:      `{"StartAt": "Test", "States": {"Test": {"Type": "Pass", "End": true}}}`,
				RoleArn:         "arn:aws:iam::123456789012:role/test",
				Type:            types.StateMachineTypeStandard,
			},
			description: "Valid definition should be preserved",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := processStateMachineInfo(tc.input)

			// Assert
			if tc.expected == nil {
				assert.Nil(t, result, tc.description)
			} else {
				assert.NotNil(t, result, tc.description)
				assert.Equal(t, tc.expected.StateMachineArn, result.StateMachineArn, "StateMachineArn should match")
				assert.Equal(t, tc.expected.Name, result.Name, "Name should match")
				assert.Equal(t, tc.expected.Status, result.Status, "Status should match")
				assert.Equal(t, tc.expected.Definition, result.Definition, "Definition should match")
				assert.Equal(t, tc.expected.RoleArn, result.RoleArn, "RoleArn should match")
				assert.Equal(t, tc.expected.Type, result.Type, "Type should match")
			}
		})
	}
}

// TestValidateStateMachineStatus tests state machine status validation
func TestValidateStateMachineStatusPolling(t *testing.T) {
	tests := []struct {
		name           string
		actual         types.StateMachineStatus
		expected       types.StateMachineStatus
		expectedResult bool
		description    string
	}{
		{
			name:           "ActiveMatchesActive",
			actual:         types.StateMachineStatusActive,
			expected:       types.StateMachineStatusActive,
			expectedResult: true,
			description:    "Active status should match active",
		},
		{
			name:           "DeletingMatchesDeleting",
			actual:         types.StateMachineStatusDeleting,
			expected:       types.StateMachineStatusDeleting,
			expectedResult: true,
			description:    "Deleting status should match deleting",
		},
		{
			name:           "ActiveDoesNotMatchDeleting",
			actual:         types.StateMachineStatusActive,
			expected:       types.StateMachineStatusDeleting,
			expectedResult: false,
			description:    "Active status should not match deleting",
		},
		{
			name:           "DeletingDoesNotMatchActive",
			actual:         types.StateMachineStatusDeleting,
			expected:       types.StateMachineStatusActive,
			expectedResult: false,
			description:    "Deleting status should not match active",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := validateStateMachineStatus(tc.actual, tc.expected)

			// Assert
			assert.Equal(t, tc.expectedResult, result, tc.description)
		})
	}
}

// TestValidateListExecutionsRequest tests list executions request validation
func TestValidateListExecutionsRequest(t *testing.T) {
	validArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine"

	tests := []struct {
		name            string
		stateMachineArn string
		statusFilter    types.ExecutionStatus
		maxResults      int32
		expectedError   bool
		errorContains   string
		description     string
	}{
		{
			name:            "ValidRequest",
			stateMachineArn: validArn,
			statusFilter:    types.ExecutionStatusRunning,
			maxResults:      10,
			expectedError:   false,
			description:     "Valid list request should pass validation",
		},
		{
			name:            "EmptyStateMachineArn",
			stateMachineArn: "",
			statusFilter:    types.ExecutionStatusRunning,
			maxResults:      10,
			expectedError:   true,
			errorContains:   "state machine ARN is required",
			description:     "Empty state machine ARN should fail validation",
		},
		{
			name:            "InvalidStateMachineArn",
			stateMachineArn: "invalid-arn",
			statusFilter:    types.ExecutionStatusRunning,
			maxResults:      10,
			expectedError:   true,
			errorContains:   "must start with 'arn:aws:states:'",
			description:     "Invalid state machine ARN should fail validation",
		},
		{
			name:            "NegativeMaxResults",
			stateMachineArn: validArn,
			statusFilter:    types.ExecutionStatusRunning,
			maxResults:      -1,
			expectedError:   true,
			errorContains:   "max results cannot be negative",
			description:     "Negative max results should fail validation",
		},
		{
			name:            "TooLargeMaxResults",
			stateMachineArn: validArn,
			statusFilter:    types.ExecutionStatusRunning,
			maxResults:      1001,
			expectedError:   true,
			errorContains:   "max results cannot exceed 1000",
			description:     "Max results over 1000 should fail validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateListExecutionsRequest(tc.stateMachineArn, tc.statusFilter, tc.maxResults)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}