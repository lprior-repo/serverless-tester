package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

// Test data factories

func createTestContext() *TestContext {
	return &TestContext{
		T:         newMockTestingT(),
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
}

func createValidStateMachineDefinition() *StateMachineDefinition {
	return &StateMachineDefinition{
		Name:       "test-state-machine",
		Definition: `{"Comment": "Test state machine", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "Result": "Hello World!", "End": true}}}`,
		RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
		Type:       types.StateMachineTypeStandard,
		Tags: map[string]string{
			"Environment": "test",
			"Purpose":     "testing",
		},
	}
}

func createValidStateMachineOptions() *StateMachineOptions {
	return &StateMachineOptions{
		LoggingConfiguration: &types.LoggingConfiguration{
			Level: types.LogLevelAll,
			Destinations: []types.LogDestination{
				{
					CloudWatchLogsLogGroup: &types.CloudWatchLogsLogGroup{
						LogGroupArn: aws.String("arn:aws:logs:us-east-1:123456789012:log-group:test-log-group"),
					},
				},
			},
		},
		TracingConfiguration: &types.TracingConfiguration{
			Enabled: true,
		},
		Tags: map[string]string{
			"Project": "testing",
		},
	}
}

// Test validation functions

func TestValidateStateMachineCreation(t *testing.T) {
	tests := []struct {
		name          string
		definition    *StateMachineDefinition
		options       *StateMachineOptions
		expectedError bool
		errorContains string
		description   string
	}{
		{
			name:          "ValidDefinitionAndOptions",
			definition:    createValidStateMachineDefinition(),
			options:       createValidStateMachineOptions(),
			expectedError: false,
			description:   "Valid definition and options should pass validation",
		},
		{
			name:          "NilDefinition",
			definition:    nil,
			options:       nil,
			expectedError: true,
			errorContains: "state machine definition is required",
			description:   "Nil definition should fail validation",
		},
		{
			name: "InvalidRoleArn",
			definition: &StateMachineDefinition{
				Name:       "test-machine",
				Definition: `{"StartAt": "Test", "States": {"Test": {"Type": "Pass", "End": true}}}`,
				RoleArn:    "invalid-arn", // Invalid ARN format
				Type:       types.StateMachineTypeStandard,
			},
			options:       nil,
			expectedError: true,
			errorContains: "must start with 'arn:aws:iam::'",
			description:   "Invalid role ARN format should fail validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStateMachineCreation(tc.definition, tc.options)

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

func TestValidateStateMachineArn(t *testing.T) {
	tests := []struct {
		name          string
		arn           string
		expectedError bool
		errorContains string
		description   string
	}{
		{
			name:          "ValidStateMachineArn",
			arn:           "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			expectedError: false,
			description:   "Valid state machine ARN should pass validation",
		},
		{
			name:          "EmptyArn",
			arn:           "",
			expectedError: true,
			errorContains: "state machine ARN is required",
			description:   "Empty ARN should fail validation",
		},
		{
			name:          "InvalidArnPrefix",
			arn:           "arn:aws:lambda:us-east-1:123456789012:function:test",
			expectedError: true,
			errorContains: "must start with 'arn:aws:states:'",
			description:   "ARN with wrong service should fail validation",
		},
		{
			name:          "MissingStateMachine",
			arn:           "arn:aws:states:us-east-1:123456789012:execution:test",
			expectedError: true,
			errorContains: "must contain ':stateMachine:'",
			description:   "ARN without stateMachine resource should fail validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStateMachineArn(tc.arn)

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

func TestValidateExecutionArn(t *testing.T) {
	tests := []struct {
		name          string
		arn           string
		expectedError bool
		errorContains string
		description   string
	}{
		{
			name:          "ValidExecutionArn",
			arn:           "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			expectedError: false,
			description:   "Valid execution ARN should pass validation",
		},
		{
			name:          "EmptyArn",
			arn:           "",
			expectedError: true,
			errorContains: "execution ARN is required",
			description:   "Empty ARN should fail validation",
		},
		{
			name:          "InvalidArnPrefix",
			arn:           "arn:aws:lambda:us-east-1:123456789012:function:test",
			expectedError: true,
			errorContains: "must start with 'arn:aws:states:'",
			description:   "ARN with wrong service should fail validation",
		},
		{
			name:          "MissingExecution",
			arn:           "arn:aws:states:us-east-1:123456789012:stateMachine:test",
			expectedError: true,
			errorContains: "must contain ':execution:'",
			description:   "ARN without execution resource should fail validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateExecutionArn(tc.arn)

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

func TestValidateRoleArn(t *testing.T) {
	tests := []struct {
		name          string
		arn           string
		expectedError bool
		errorContains string
		description   string
	}{
		{
			name:          "ValidRoleArn",
			arn:           "arn:aws:iam::123456789012:role/StepFunctionsRole",
			expectedError: false,
			description:   "Valid role ARN should pass validation",
		},
		{
			name:          "EmptyArn",
			arn:           "",
			expectedError: true,
			errorContains: "role ARN is required",
			description:   "Empty ARN should fail validation",
		},
		{
			name:          "InvalidArnPrefix",
			arn:           "arn:aws:lambda::123456789012:function:test",
			expectedError: true,
			errorContains: "must start with 'arn:aws:iam::'",
			description:   "ARN with wrong service should fail validation",
		},
		{
			name:          "MissingRole",
			arn:           "arn:aws:iam::123456789012:user/test",
			expectedError: true,
			errorContains: "must contain ':role/'",
			description:   "ARN without role resource should fail validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateRoleArn(tc.arn)

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

func TestValidateMaxResults(t *testing.T) {
	tests := []struct {
		name          string
		maxResults    int32
		expectedError bool
		errorContains string
		description   string
	}{
		{
			name:          "ValidMaxResults",
			maxResults:    10,
			expectedError: false,
			description:   "Valid max results should pass validation",
		},
		{
			name:          "ZeroMaxResults",
			maxResults:    0,
			expectedError: false,
			description:   "Zero max results should pass validation",
		},
		{
			name:          "NegativeMaxResults",
			maxResults:    -1,
			expectedError: true,
			errorContains: "max results cannot be negative",
			description:   "Negative max results should fail validation",
		},
		{
			name:          "TooLargeMaxResults",
			maxResults:    1001,
			expectedError: true,
			errorContains: "max results cannot exceed 1000",
			description:   "Max results over 1000 should fail validation",
		},
		{
			name:          "ExactlyMaxResults",
			maxResults:    1000,
			expectedError: false,
			description:   "Max results exactly 1000 should pass validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateMaxResults(tc.maxResults)

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

// Test execution validation functions

func TestValidateStartExecutionRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       *StartExecutionRequest
		expectedError bool
		errorContains string
		description   string
	}{
		{
			name: "ValidRequest",
			request: &StartExecutionRequest{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
				Name:            "test-execution",
				Input:           `{"message": "Hello, World!"}`,
			},
			expectedError: false,
			description:   "Valid request should pass validation",
		},
		{
			name:          "NilRequest",
			request:       nil,
			expectedError: true,
			errorContains: "start execution request is required",
			description:   "Nil request should fail validation",
		},
		{
			name: "EmptyStateMachineArn",
			request: &StartExecutionRequest{
				StateMachineArn: "",
				Name:            "test-execution",
				Input:           `{}`,
			},
			expectedError: true,
			errorContains: "state machine ARN is required",
			description:   "Empty state machine ARN should fail validation",
		},
		{
			name: "EmptyExecutionName",
			request: &StartExecutionRequest{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
				Name:            "",
				Input:           `{}`,
			},
			expectedError: true,
			errorContains: "execution name is required",
			description:   "Empty execution name should fail validation",
		},
		{
			name: "InvalidInputJSON",
			request: &StartExecutionRequest{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
				Name:            "test-execution",
				Input:           `{"invalid": json}`, // Invalid JSON
			},
			expectedError: true,
			errorContains: "invalid JSON",
			description:   "Invalid JSON input should fail validation",
		},
		{
			name: "EmptyInputAllowed",
			request: &StartExecutionRequest{
				StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
				Name:            "test-execution",
				Input:           "", // Empty input should be allowed
			},
			expectedError: false,
			description:   "Empty input should be allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStartExecutionRequest(tc.request)

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

func TestValidateStopExecutionRequest(t *testing.T) {
	tests := []struct {
		name          string
		executionArn  string
		error         string
		cause         string
		expectedError bool
		errorContains string
		description   string
	}{
		{
			name:          "ValidRequest",
			executionArn:  "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			error:         "User requested stop",
			cause:         "Test stop",
			expectedError: false,
			description:   "Valid stop request should pass validation",
		},
		{
			name:          "EmptyExecutionArn",
			executionArn:  "",
			error:         "Error",
			cause:         "Cause",
			expectedError: true,
			errorContains: "execution ARN is required",
			description:   "Empty execution ARN should fail validation",
		},
		{
			name:          "EmptyErrorAndCauseAllowed",
			executionArn:  "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			error:         "",
			cause:         "",
			expectedError: false,
			description:   "Empty error and cause should be allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStopExecutionRequest(tc.executionArn, tc.error, tc.cause)

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

func TestValidateWaitForExecutionRequest(t *testing.T) {
	tests := []struct {
		name          string
		executionArn  string
		options       *WaitOptions
		expectedError bool
		errorContains string
		description   string
	}{
		{
			name:         "ValidRequest",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			options: &WaitOptions{
				Timeout:         5 * time.Minute,
				PollingInterval: 5 * time.Second,
				MaxAttempts:     60,
			},
			expectedError: false,
			description:   "Valid wait request should pass validation",
		},
		{
			name:          "EmptyExecutionArn",
			executionArn:  "",
			options:       nil,
			expectedError: true,
			errorContains: "execution ARN is required",
			description:   "Empty execution ARN should fail validation",
		},
		{
			name:          "NilOptionsAllowed",
			executionArn:  "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			options:       nil,
			expectedError: false,
			description:   "Nil options should be allowed (defaults will be used)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateWaitForExecutionRequest(tc.executionArn, tc.options)

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


// Utility functions are tested in utility_functions_test.go to avoid duplication

// Option merging and utility functions are tested in utility_functions_test.go