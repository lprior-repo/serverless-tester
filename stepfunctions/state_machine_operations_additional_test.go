package stepfunctions

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test additional state machine operations for comprehensive coverage
func TestUpdateStateMachineAdditionalScenarios(t *testing.T) {
	tests := []struct {
		name            string
		stateMachineArn string
		definition      string
		roleArn         string
		options         *StateMachineOptions
		mockSetup       func(*MockSFNClientInterface)
		expectedError   bool
		errorContains   string
		description     string
	}{
		{
			name:            "SuccessfulUpdateWithValidParams",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-update",
			definition:      `{"StartAt": "Updated", "States": {"Updated": {"Type": "Pass", "End": true}}}`,
			roleArn:         "arn:aws:iam::123456789012:role/UpdatedRole",
			options:         nil,
			mockSetup: func(mockClient *MockSFNClientInterface) {
				updateOutput := &sfn.UpdateStateMachineOutput{
					UpdateDate: aws.Time(time.Now()),
				}
				mockClient.On("UpdateStateMachine", mock.Anything, mock.Anything).Return(updateOutput, nil)

				describeOutput := &sfn.DescribeStateMachineOutput{
					StateMachineArn: aws.String("arn:aws:states:us-east-1:123456789012:stateMachine:test-update"),
					Name:            aws.String("test-update"),
					Status:          types.StateMachineStatusActive,
					Definition:      aws.String(`{"StartAt": "Updated", "States": {"Updated": {"Type": "Pass", "End": true}}}`),
					RoleArn:         aws.String("arn:aws:iam::123456789012:role/UpdatedRole"),
					Type:            types.StateMachineTypeStandard,
					CreationDate:    aws.Time(time.Now().Add(-1 * time.Hour)),
				}
				mockClient.On("DescribeStateMachine", mock.Anything, mock.Anything).Return(describeOutput, nil)
			},
			expectedError: false,
			description:   "Valid update parameters should succeed",
		},
		{
			name:            "UpdateFailsWithInvalidArn",
			stateMachineArn: "invalid-arn-format",
			definition:      `{"StartAt": "Test", "States": {"Test": {"Type": "Pass", "End": true}}}`,
			roleArn:         "arn:aws:iam::123456789012:role/TestRole",
			options:         nil,
			mockSetup:       func(mockClient *MockSFNClientInterface) {},
			expectedError:   true,
			errorContains:   "must start with 'arn:aws:states:'",
			description:     "Invalid ARN format should fail validation",
		},
		{
			name:            "UpdateFailsWithEmptyRoleArn",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			definition:      `{"StartAt": "Test", "States": {"Test": {"Type": "Pass", "End": true}}}`,
			roleArn:         "", // Empty role ARN
			options:         nil,
			mockSetup:       func(mockClient *MockSFNClientInterface) {},
			expectedError:   true,
			errorContains:   "role ARN is required",
			description:     "Empty role ARN should fail validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockSFNClientInterface{}
			tc.mockSetup(mockClient)
			ctx := createTestContext()

			// Act
			result, err := updateStateMachineETestHelper(ctx, mockClient, tc.stateMachineArn, tc.definition, tc.roleArn, tc.options)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
				assert.NotNil(t, result, "Result should not be nil for successful update")
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// Test delete state machine additional scenarios
func TestDeleteStateMachineAdditionalScenarios(t *testing.T) {
	tests := []struct {
		name            string
		stateMachineArn string
		mockSetup       func(*MockSFNClientInterface)
		expectedError   bool
		errorContains   string
		description     string
	}{
		{
			name:            "SuccessfulDeletion",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:delete-test",
			mockSetup: func(mockClient *MockSFNClientInterface) {
				deleteOutput := &sfn.DeleteStateMachineOutput{}
				mockClient.On("DeleteStateMachine", mock.Anything, mock.Anything).Return(deleteOutput, nil)
			},
			expectedError: false,
			description:   "Valid deletion should succeed",
		},
		{
			name:            "DeletionFailsWithEmptyArn",
			stateMachineArn: "",
			mockSetup:       func(mockClient *MockSFNClientInterface) {},
			expectedError:   true,
			errorContains:   "state machine ARN is required",
			description:     "Empty ARN should fail validation",
		},
		{
			name:            "DeletionFailsWithInvalidArnFormat",
			stateMachineArn: "not-a-valid-arn",
			mockSetup:       func(mockClient *MockSFNClientInterface) {},
			expectedError:   true,
			errorContains:   "must start with 'arn:aws:states:'",
			description:     "Invalid ARN format should fail validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockSFNClientInterface{}
			tc.mockSetup(mockClient)
			ctx := createTestContext()

			// Act
			err := deleteStateMachineETestHelper(ctx, mockClient, tc.stateMachineArn)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// Test describe state machine additional scenarios
func TestDescribeStateMachineAdditionalScenarios(t *testing.T) {
	tests := []struct {
		name            string
		stateMachineArn string
		mockSetup       func(*MockSFNClientInterface)
		expectedError   bool
		errorContains   string
		validateResult  func(*testing.T, *StateMachineInfo)
		description     string
	}{
		{
			name:            "SuccessfulDescribe",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:describe-test",
			mockSetup: func(mockClient *MockSFNClientInterface) {
				describeOutput := &sfn.DescribeStateMachineOutput{
					StateMachineArn: aws.String("arn:aws:states:us-east-1:123456789012:stateMachine:describe-test"),
					Name:            aws.String("describe-test"),
					Status:          types.StateMachineStatusActive,
					Definition:      aws.String(`{"StartAt": "Test", "States": {"Test": {"Type": "Pass", "End": true}}}`),
					RoleArn:         aws.String("arn:aws:iam::123456789012:role/TestRole"),
					Type:            types.StateMachineTypeStandard,
					CreationDate:    aws.Time(time.Now().Add(-2 * time.Hour)),
				}
				mockClient.On("DescribeStateMachine", mock.Anything, mock.Anything).Return(describeOutput, nil)
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *StateMachineInfo) {
				assert.NotNil(t, result, "Result should not be nil")
				assert.Equal(t, "describe-test", result.Name, "Name should match")
				assert.Equal(t, types.StateMachineStatusActive, result.Status, "Status should be active")
			},
			description: "Valid describe should succeed",
		},
		{
			name:            "DescribeFailsWithEmptyArn",
			stateMachineArn: "",
			mockSetup:       func(mockClient *MockSFNClientInterface) {},
			expectedError:   true,
			errorContains:   "state machine ARN is required",
			validateResult: func(t *testing.T, result *StateMachineInfo) {
				assert.Nil(t, result, "Result should be nil on error")
			},
			description: "Empty ARN should fail validation",
		},
		{
			name:            "DescribeFailsWithInvalidArnFormat",
			stateMachineArn: "arn:aws:lambda:us-east-1:123456789012:function:test",
			mockSetup:       func(mockClient *MockSFNClientInterface) {},
			expectedError:   true,
			errorContains:   "must start with 'arn:aws:states:'",
			validateResult: func(t *testing.T, result *StateMachineInfo) {
				assert.Nil(t, result, "Result should be nil on validation error")
			},
			description: "Wrong service ARN should fail validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockSFNClientInterface{}
			tc.mockSetup(mockClient)
			ctx := createTestContext()

			// Act
			result, err := describeStateMachineETestHelper(ctx, mockClient, tc.stateMachineArn)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, tc.description)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, tc.description)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}

			tc.validateResult(t, result)
			mockClient.AssertExpectations(t)
		})
	}
}

// Test comprehensive validation scenarios
func TestStateMachineValidationComprehensive(t *testing.T) {
	t.Run("ValidateDefinition", func(t *testing.T) {
		tests := []struct {
			name          string
			definition    string
			expectedError bool
			errorContains string
		}{
			{
				name:          "ValidMinimalDefinition",
				definition:    `{"StartAt": "Start", "States": {"Start": {"Type": "Pass", "End": true}}}`,
				expectedError: false,
			},
			{
				name:          "EmptyDefinition",
				definition:    "",
				expectedError: true,
				errorContains: "empty definition",
			},
			{
				name:          "InvalidJSON",
				definition:    `{"StartAt": "Start", "States":`,
				expectedError: true,
				errorContains: "invalid JSON",
			},
			{
				name:          "MissingStartAt",
				definition:    `{"States": {"Start": {"Type": "Pass", "End": true}}}`,
				expectedError: true,
				errorContains: "missing required field 'StartAt'",
			},
			{
				name:          "MissingStates",
				definition:    `{"StartAt": "Start"}`,
				expectedError: true,
				errorContains: "missing required field 'States'",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				err := validateStateMachineDefinition(tc.definition)
				if tc.expectedError {
					assert.Error(t, err)
					if tc.errorContains != "" {
						assert.Contains(t, err.Error(), tc.errorContains)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("ValidateStateMachineName", func(t *testing.T) {
		tests := []struct {
			name          string
			stateName     string
			expectedError bool
			errorContains string
		}{
			{
				name:          "ValidName",
				stateName:     "MyStateMachine",
				expectedError: false,
			},
			{
				name:          "ValidNameWithHyphens",
				stateName:     "my-state-machine",
				expectedError: false,
			},
			{
				name:          "ValidNameWithUnderscores",
				stateName:     "my_state_machine",
				expectedError: false,
			},
			{
				name:          "EmptyName",
				stateName:     "",
				expectedError: true,
				errorContains: "invalid state machine name",
			},
			{
				name:          "NameWithSpaces",
				stateName:     "My State Machine",
				expectedError: true,
				errorContains: "name contains invalid characters",
			},
			{
				name:          "NameTooLong",
				stateName:     generateString("A", MaxStateMachineNameLen+1),
				expectedError: true,
				errorContains: "name too long",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				err := validateStateMachineName(tc.stateName)
				if tc.expectedError {
					assert.Error(t, err)
					if tc.errorContains != "" {
						assert.Contains(t, err.Error(), tc.errorContains)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("ValidateRoleArn", func(t *testing.T) {
		tests := []struct {
			name          string
			roleArn       string
			expectedError bool
			errorContains string
		}{
			{
				name:          "ValidRoleArn",
				roleArn:       "arn:aws:iam::123456789012:role/StepFunctionsRole",
				expectedError: false,
			},
			{
				name:          "ValidRoleArnWithPath",
				roleArn:       "arn:aws:iam::123456789012:role/service-role/StepFunctionsRole",
				expectedError: false,
			},
			{
				name:          "EmptyRoleArn",
				roleArn:       "",
				expectedError: true,
				errorContains: "role ARN is required",
			},
			{
				name:          "InvalidArnPrefix",
				roleArn:       "arn:aws:lambda::123456789012:function:test",
				expectedError: true,
				errorContains: "must start with 'arn:aws:iam::'",
			},
			{
				name:          "MissingRoleResource",
				roleArn:       "arn:aws:iam::123456789012:user/TestUser",
				expectedError: true,
				errorContains: "must contain ':role/'",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				err := validateRoleArn(tc.roleArn)
				if tc.expectedError {
					assert.Error(t, err)
					if tc.errorContains != "" {
						assert.Contains(t, err.Error(), tc.errorContains)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

// Helper functions with dependency injection for testing

func updateStateMachineETestHelper(ctx *TestContext, client SFNClientInterface, stateMachineArn, definition, roleArn string, options *StateMachineOptions) (*StateMachineInfo, error) {
	start := time.Now()
	logOperation("UpdateStateMachine", map[string]interface{}{
		"arn": stateMachineArn,
	})

	// Validate input parameters
	if err := validateStateMachineUpdate(stateMachineArn, definition, roleArn, options); err != nil {
		logResult("UpdateStateMachine", false, time.Since(start), err)
		return nil, err
	}

	input := &sfn.UpdateStateMachineInput{
		StateMachineArn: aws.String(stateMachineArn),
	}

	if definition != "" {
		input.Definition = aws.String(definition)
	}
	if roleArn != "" {
		input.RoleArn = aws.String(roleArn)
	}

	if options != nil {
		if options.LoggingConfiguration != nil {
			input.LoggingConfiguration = options.LoggingConfiguration
		}
		if options.TracingConfiguration != nil {
			input.TracingConfiguration = options.TracingConfiguration
		}
	}

	_, err := client.UpdateStateMachine(context.TODO(), input)
	if err != nil {
		logResult("UpdateStateMachine", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to update state machine: %w", err)
	}

	// Get updated state machine information
	info, err := describeStateMachineETestHelper(ctx, client, stateMachineArn)
	if err != nil {
		logResult("UpdateStateMachine", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to describe updated state machine: %w", err)
	}

	logResult("UpdateStateMachine", true, time.Since(start), nil)
	return info, nil
}

func deleteStateMachineETestHelper(ctx *TestContext, client SFNClientInterface, stateMachineArn string) error {
	start := time.Now()
	logOperation("DeleteStateMachine", map[string]interface{}{
		"arn": stateMachineArn,
	})

	if err := validateStateMachineArn(stateMachineArn); err != nil {
		logResult("DeleteStateMachine", false, time.Since(start), err)
		return err
	}

	input := &sfn.DeleteStateMachineInput{
		StateMachineArn: aws.String(stateMachineArn),
	}

	_, err := client.DeleteStateMachine(context.TODO(), input)
	if err != nil {
		logResult("DeleteStateMachine", false, time.Since(start), err)
		return fmt.Errorf("failed to delete state machine: %w", err)
	}

	logResult("DeleteStateMachine", true, time.Since(start), nil)
	return nil
}

func describeStateMachineETestHelper(ctx *TestContext, client SFNClientInterface, stateMachineArn string) (*StateMachineInfo, error) {
	start := time.Now()
	logOperation("DescribeStateMachine", map[string]interface{}{
		"arn": stateMachineArn,
	})

	if err := validateStateMachineArn(stateMachineArn); err != nil {
		logResult("DescribeStateMachine", false, time.Since(start), err)
		return nil, err
	}

	input := &sfn.DescribeStateMachineInput{
		StateMachineArn: aws.String(stateMachineArn),
	}

	result, err := client.DescribeStateMachine(context.TODO(), input)
	if err != nil {
		logResult("DescribeStateMachine", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to describe state machine: %w", err)
	}

	info := &StateMachineInfo{
		StateMachineArn:      aws.ToString(result.StateMachineArn),
		Name:                 aws.ToString(result.Name),
		Status:               result.Status,
		Definition:           aws.ToString(result.Definition),
		RoleArn:              aws.ToString(result.RoleArn),
		Type:                 result.Type,
		CreationDate:         aws.ToTime(result.CreationDate),
		LoggingConfiguration: result.LoggingConfiguration,
		TracingConfiguration: result.TracingConfiguration,
	}

	logResult("DescribeStateMachine", true, time.Since(start), nil)
	return processStateMachineInfo(info), nil
}

// Helper function to generate strings of specified length
func generateString(char string, length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += char
	}
	return result
}