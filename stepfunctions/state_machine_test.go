package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

// Test data factories for state machine operations

func createTestStateMachineOptions() *StateMachineOptions {
	return &StateMachineOptions{
		LoggingConfiguration: &types.LoggingConfiguration{
			IncludeExecutionData: true,
			Level:                types.LogLevelOff,
			Destinations: []types.LogDestination{
				{
					CloudWatchLogsLogGroup: &types.CloudWatchLogsLogGroup{
						LogGroupArn: aws.String("arn:aws:logs:us-east-1:123456789012:log-group:stepfunctions-logs:*"),
					},
				},
			},
		},
		TracingConfiguration: &types.TracingConfiguration{
			Enabled: true,
		},
		Tags: map[string]string{
			"Environment": "test",
			"Purpose":     "testing",
		},
	}
}

func createTestStateMachineInfo() *StateMachineInfo {
	return &StateMachineInfo{
		StateMachineArn:      "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
		Name:                 "test-state-machine",
		Status:               types.StateMachineStatusActive,
		Definition:           `{"Comment": "Test state machine", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "Result": "Hello World!", "End": true}}}`,
		RoleArn:              "arn:aws:iam::123456789012:role/StepFunctionsRole",
		Type:                 types.StateMachineTypeStandard,
		CreationDate:         time.Now().Add(-24 * time.Hour),
		LoggingConfiguration: nil,
		TracingConfiguration: nil,
	}
}

// Table-driven test structures

type stateMachineCreationTestCase struct {
	name                 string
	definition           *StateMachineDefinition
	options              *StateMachineOptions
	shouldError          bool
	expectedErrorMessage string
	description          string
}

type stateMachineUpdateTestCase struct {
	name                 string
	stateMachineArn      string
	definition           string
	roleArn              string
	options              *StateMachineOptions
	shouldError          bool
	expectedErrorMessage string
	description          string
}

// Unit tests for state machine operations

func TestCreateStateMachineE(t *testing.T) {
	tests := []stateMachineCreationTestCase{
		{
			name: "ValidStateMachineCreation",
			definition: &StateMachineDefinition{
				Name:       "test-state-machine",
				Definition: `{"Comment": "Test", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			options:     nil,
			shouldError: false,
			description: "Valid state machine should be created successfully",
		},
		{
			name: "InvalidStateMachineName",
			definition: &StateMachineDefinition{
				Name:       "",
				Definition: `{"Comment": "Test", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			options:              nil,
			shouldError:          true,
			expectedErrorMessage: "invalid state machine name",
			description:          "Empty state machine name should fail",
		},
		{
			name: "InvalidDefinition",
			definition: &StateMachineDefinition{
				Name:       "test-state-machine",
				Definition: `{"invalid": json}`,
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			options:              nil,
			shouldError:          true,
			expectedErrorMessage: "invalid state machine definition",
			description:          "Invalid JSON definition should fail",
		},
		{
			name: "EmptyRoleArn",
			definition: &StateMachineDefinition{
				Name:       "test-state-machine",
				Definition: `{"Comment": "Test", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
				RoleArn:    "",
				Type:       types.StateMachineTypeStandard,
			},
			options:              nil,
			shouldError:          true,
			expectedErrorMessage: "role ARN is required",
			description:          "Empty role ARN should fail",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStateMachineCreation(tc.definition, tc.options)
			
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

func TestUpdateStateMachineE(t *testing.T) {
	tests := []stateMachineUpdateTestCase{
		{
			name:            "ValidStateMachineUpdate",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			definition:      `{"Comment": "Updated", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
			roleArn:         "arn:aws:iam::123456789012:role/StepFunctionsRole",
			options:         nil,
			shouldError:     false,
			description:     "Valid state machine update should succeed",
		},
		{
			name:                 "InvalidStateMachineArn",
			stateMachineArn:      "",
			definition:           `{"Comment": "Updated", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
			roleArn:              "arn:aws:iam::123456789012:role/StepFunctionsRole",
			options:              nil,
			shouldError:          true,
			expectedErrorMessage: "state machine ARN is required",
			description:          "Empty state machine ARN should fail",
		},
		{
			name:                 "InvalidDefinition",
			stateMachineArn:      "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			definition:           `{"invalid": json}`,
			roleArn:              "arn:aws:iam::123456789012:role/StepFunctionsRole",
			options:              nil,
			shouldError:          true,
			expectedErrorMessage: "invalid state machine definition",
			description:          "Invalid JSON definition should fail",
		},
		{
			name:                 "EmptyRoleArn",
			stateMachineArn:      "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			definition:           `{"Comment": "Updated", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
			roleArn:              "",
			options:              nil,
			shouldError:          true,
			expectedErrorMessage: "role ARN is required",
			description:          "Empty role ARN should fail",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStateMachineUpdate(tc.stateMachineArn, tc.definition, tc.roleArn, tc.options)
			
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

func TestDescribeStateMachineE(t *testing.T) {
	tests := []struct {
		name                 string
		stateMachineArn      string
		shouldError          bool
		expectedErrorMessage string
		description          string
	}{
		{
			name:            "ValidStateMachineArn",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			shouldError:     false,
			description:     "Valid state machine ARN should pass validation",
		},
		{
			name:                 "EmptyStateMachineArn",
			stateMachineArn:      "",
			shouldError:          true,
			expectedErrorMessage: "state machine ARN is required",
			description:          "Empty state machine ARN should fail",
		},
		{
			name:                 "InvalidStateMachineArn",
			stateMachineArn:      "invalid-arn",
			shouldError:          true,
			expectedErrorMessage: "invalid ARN",
			description:          "Invalid ARN format should fail",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStateMachineArn(tc.stateMachineArn)
			
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

func TestDeleteStateMachineE(t *testing.T) {
	tests := []struct {
		name                 string
		stateMachineArn      string
		shouldError          bool
		expectedErrorMessage string
		description          string
	}{
		{
			name:            "ValidDeletion",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			shouldError:     false,
			description:     "Valid state machine ARN should pass validation for deletion",
		},
		{
			name:                 "EmptyStateMachineArn",
			stateMachineArn:      "",
			shouldError:          true,
			expectedErrorMessage: "state machine ARN is required",
			description:          "Empty state machine ARN should fail deletion",
		},
		{
			name:                 "InvalidStateMachineArn",
			stateMachineArn:      "not-an-arn",
			shouldError:          true,
			expectedErrorMessage: "invalid ARN",
			description:          "Invalid ARN format should fail deletion",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateStateMachineArn(tc.stateMachineArn)
			
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

func TestListStateMachinesE(t *testing.T) {
	tests := []struct {
		name        string
		maxResults  int32
		shouldError bool
		description string
	}{
		{
			name:        "ValidMaxResults",
			maxResults:  100,
			shouldError: false,
			description: "Valid max results should pass",
		},
		{
			name:        "DefaultMaxResults",
			maxResults:  0,
			shouldError: false,
			description: "Default max results (0) should pass",
		},
		{
			name:        "NegativeMaxResults",
			maxResults:  -1,
			shouldError: true,
			description: "Negative max results should fail",
		},
		{
			name:        "TooLargeMaxResults",
			maxResults:  1001,
			shouldError: true,
			description: "Max results exceeding limit should fail",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateMaxResults(tc.maxResults)
			
			// Assert
			if tc.shouldError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// Tests for state machine info processing

func TestProcessStateMachineInfo(t *testing.T) {
	// Arrange
	info := createTestStateMachineInfo()
	
	// Act
	processed := processStateMachineInfo(info)
	
	// Assert
	assert.Equal(t, info.StateMachineArn, processed.StateMachineArn, "ARN should be preserved")
	assert.Equal(t, info.Name, processed.Name, "Name should be preserved")
	assert.Equal(t, info.Status, processed.Status, "Status should be preserved")
	assert.Equal(t, info.Type, processed.Type, "Type should be preserved")
	assert.NotEmpty(t, processed.Definition, "Definition should not be empty")
}

func TestValidateStateMachineStatus(t *testing.T) {
	tests := []struct {
		name           string
		status         types.StateMachineStatus
		expectedStatus types.StateMachineStatus
		shouldMatch    bool
		description    string
	}{
		{
			name:           "ActiveStatusMatch",
			status:         types.StateMachineStatusActive,
			expectedStatus: types.StateMachineStatusActive,
			shouldMatch:    true,
			description:    "Active status should match expected active status",
		},
		{
			name:           "DeletingStatusMatch",
			status:         types.StateMachineStatusDeleting,
			expectedStatus: types.StateMachineStatusDeleting,
			shouldMatch:    true,
			description:    "Deleting status should match expected deleting status",
		},
		{
			name:           "StatusMismatch",
			status:         types.StateMachineStatusActive,
			expectedStatus: types.StateMachineStatusDeleting,
			shouldMatch:    false,
			description:    "Active status should not match expected deleting status",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			matches := validateStateMachineStatus(tc.status, tc.expectedStatus)
			
			// Assert
			assert.Equal(t, tc.shouldMatch, matches, tc.description)
		})
	}
}

// Tests for non-E variants (panic on error functions)

func TestCreateStateMachine(t *testing.T) {
	tests := []struct {
		name              string
		definition        *StateMachineDefinition
		options           *StateMachineOptions
		shouldPanic       bool
		expectedArn       string
		description       string
	}{
		{
			name: "ValidCreationShouldSucceed",
			definition: &StateMachineDefinition{
				Name:       "test-state-machine",
				Definition: `{"Comment": "Test", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			options:     createTestStateMachineOptions(),
			shouldPanic: false,
			expectedArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			description: "Valid state machine creation should succeed without panic",
		},
		{
			name: "InvalidNameShouldPanic",
			definition: &StateMachineDefinition{
				Name:       "", // Invalid name
				Definition: `{"Comment": "Test", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "End": true}}}`,
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			options:     nil,
			shouldPanic: true,
			description: "Invalid state machine name should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				// Test that the function panics for invalid input
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected CreateStateMachine to panic for %s, but it did not", tc.description)
					}
				}()
				
				// This should panic due to validation failure, we can't test with actual AWS call
				testCtx := &TestContext{T: &MockT{errorMessages: make([]string, 0)}}
				CreateStateMachine(testCtx, tc.definition, tc.options)
			} else {
				// For valid cases, we can't test the actual AWS call without mocking
				// but we can test that validation passes
				err := validateStateMachineCreation(tc.definition, tc.options)
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestDescribeStateMachine(t *testing.T) {
	tests := []struct {
		name            string
		stateMachineArn string
		shouldPanic     bool
		description     string
	}{
		{
			name:            "ValidArnShouldSucceed",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			shouldPanic:     false,
			description:     "Valid state machine ARN should succeed without panic",
		},
		{
			name:            "InvalidArnShouldPanic",
			stateMachineArn: "", // Invalid ARN
			shouldPanic:     true,
			description:     "Invalid state machine ARN should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				// Test that the function panics for invalid input
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected DescribeStateMachine to panic for %s, but it did not", tc.description)
					}
				}()
				
				testCtx := &TestContext{T: &MockT{errorMessages: make([]string, 0)}}
				DescribeStateMachine(testCtx, tc.stateMachineArn)
			} else {
				// For valid cases, we can only test that validation passes
				err := validateStateMachineArn(tc.stateMachineArn)
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestDeleteStateMachine(t *testing.T) {
	tests := []struct {
		name            string
		stateMachineArn string
		shouldPanic     bool
		description     string
	}{
		{
			name:            "ValidArnShouldSucceed",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
			shouldPanic:     false,
			description:     "Valid state machine ARN should succeed without panic",
		},
		{
			name:            "InvalidArnShouldPanic",
			stateMachineArn: "invalid-arn",
			shouldPanic:     true,
			description:     "Invalid state machine ARN should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				// Test that the function panics for invalid input
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected DeleteStateMachine to panic for %s, but it did not", tc.description)
					}
				}()
				
				testCtx := &TestContext{T: &MockT{errorMessages: make([]string, 0)}}
				DeleteStateMachine(testCtx, tc.stateMachineArn)
			} else {
				// For valid cases, we can only test that validation passes
				err := validateStateMachineArn(tc.stateMachineArn)
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func TestListStateMachines(t *testing.T) {
	tests := []struct {
		name        string
		maxResults  int32
		shouldPanic bool
		description string
	}{
		{
			name:        "ValidMaxResultsShouldSucceed",
			maxResults:  100,
			shouldPanic: false,
			description: "Valid max results should succeed without panic",
		},
		{
			name:        "InvalidMaxResultsShouldPanic",
			maxResults:  -1,
			shouldPanic: true,
			description: "Invalid max results should cause panic",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				// Test that the function panics for invalid input
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected ListStateMachines to panic for %s, but it did not", tc.description)
					}
				}()
				
				testCtx := &TestContext{T: &MockT{errorMessages: make([]string, 0)}}
				ListStateMachines(testCtx, tc.maxResults)
			} else {
				// For valid cases, we can only test that validation passes
				err := validateMaxResults(tc.maxResults)
				assert.NoError(t, err, tc.description)
			}
		})
	}
}


// Benchmark tests for state machine operations

func BenchmarkValidateStateMachineNameForStateMachines(b *testing.B) {
	stateMachineName := "my-test-state-machine"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateStateMachineName(stateMachineName)
	}
}

func BenchmarkValidateStateMachineDefinition(b *testing.B) {
	definition := `{"Comment": "Test state machine", "StartAt": "HelloWorld", "States": {"HelloWorld": {"Type": "Pass", "Result": "Hello World!", "End": true}}}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateStateMachineDefinition(definition)
	}
}

func BenchmarkProcessStateMachineInfo(b *testing.B) {
	info := createTestStateMachineInfo()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processStateMachineInfo(info)
	}
}