package stepfunctions

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// SFNClientInterface defines the interface for Step Functions client operations
type SFNClientInterface interface {
	CreateStateMachine(ctx context.Context, input *sfn.CreateStateMachineInput, opts ...func(*sfn.Options)) (*sfn.CreateStateMachineOutput, error)
	UpdateStateMachine(ctx context.Context, input *sfn.UpdateStateMachineInput, opts ...func(*sfn.Options)) (*sfn.UpdateStateMachineOutput, error)
	DescribeStateMachine(ctx context.Context, input *sfn.DescribeStateMachineInput, opts ...func(*sfn.Options)) (*sfn.DescribeStateMachineOutput, error)
	DeleteStateMachine(ctx context.Context, input *sfn.DeleteStateMachineInput, opts ...func(*sfn.Options)) (*sfn.DeleteStateMachineOutput, error)
	ListStateMachines(ctx context.Context, input *sfn.ListStateMachinesInput, opts ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error)
	StartExecution(ctx context.Context, input *sfn.StartExecutionInput, opts ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error)
	StopExecution(ctx context.Context, input *sfn.StopExecutionInput, opts ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error)
	DescribeExecution(ctx context.Context, input *sfn.DescribeExecutionInput, opts ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error)
}

// MockSFNClientInterface is a mock implementation
type MockSFNClientInterface struct {
	mock.Mock
}

// MockStepFunctionsClient provides mock functions for testing
type MockStepFunctionsClient struct {
	GetExecutionHistoryFunc func(ctx context.Context, input *sfn.GetExecutionHistoryInput) (*sfn.GetExecutionHistoryOutput, error)
}

func (m *MockSFNClientInterface) CreateStateMachine(ctx context.Context, input *sfn.CreateStateMachineInput, opts ...func(*sfn.Options)) (*sfn.CreateStateMachineOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sfn.CreateStateMachineOutput), args.Error(1)
}

func (m *MockSFNClientInterface) UpdateStateMachine(ctx context.Context, input *sfn.UpdateStateMachineInput, opts ...func(*sfn.Options)) (*sfn.UpdateStateMachineOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sfn.UpdateStateMachineOutput), args.Error(1)
}

func (m *MockSFNClientInterface) DescribeStateMachine(ctx context.Context, input *sfn.DescribeStateMachineInput, opts ...func(*sfn.Options)) (*sfn.DescribeStateMachineOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sfn.DescribeStateMachineOutput), args.Error(1)
}

func (m *MockSFNClientInterface) DeleteStateMachine(ctx context.Context, input *sfn.DeleteStateMachineInput, opts ...func(*sfn.Options)) (*sfn.DeleteStateMachineOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sfn.DeleteStateMachineOutput), args.Error(1)
}

func (m *MockSFNClientInterface) ListStateMachines(ctx context.Context, input *sfn.ListStateMachinesInput, opts ...func(*sfn.Options)) (*sfn.ListStateMachinesOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sfn.ListStateMachinesOutput), args.Error(1)
}

func (m *MockSFNClientInterface) StartExecution(ctx context.Context, input *sfn.StartExecutionInput, opts ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sfn.StartExecutionOutput), args.Error(1)
}

func (m *MockSFNClientInterface) StopExecution(ctx context.Context, input *sfn.StopExecutionInput, opts ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sfn.StopExecutionOutput), args.Error(1)
}

func (m *MockSFNClientInterface) DescribeExecution(ctx context.Context, input *sfn.DescribeExecutionInput, opts ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sfn.DescribeExecutionOutput), args.Error(1)
}

// TestCreateStateMachineEWithMock tests CreateStateMachineE with proper mocking
func TestCreateStateMachineEWithMock(t *testing.T) {
	tests := []struct {
		name           string
		definition     *StateMachineDefinition
		options        *StateMachineOptions
		mockSetup      func(*MockSFNClientInterface)
		expectedError  bool
		errorContains  string
		validateResult func(*testing.T, *StateMachineInfo)
		description    string
	}{
		{
			name:       "SuccessfulCreation",
			definition: createValidStateMachineDefinition(),
			options:    createValidStateMachineOptions(),
			mockSetup: func(mockClient *MockSFNClientInterface) {
				expectedOutput := &sfn.CreateStateMachineOutput{
					StateMachineArn: aws.String("arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine"),
					CreationDate:    aws.Time(time.Now()),
				}
				mockClient.On("CreateStateMachine", mock.Anything, mock.MatchedBy(func(input *sfn.CreateStateMachineInput) bool {
					return aws.ToString(input.Name) == "test-state-machine" &&
						aws.ToString(input.RoleArn) == "arn:aws:iam::123456789012:role/StepFunctionsRole"
				})).Return(expectedOutput, nil)
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *StateMachineInfo) {
				assert.NotNil(t, result, "Result should not be nil")
				assert.Equal(t, "test-state-machine", result.Name, "Name should match")
				assert.Equal(t, types.StateMachineTypeStandard, result.Type, "Type should match")
				assert.Contains(t, result.StateMachineArn, "test-state-machine", "ARN should contain state machine name")
			},
			description: "Valid state machine creation should succeed",
		},
		{
			name: "InvalidStateMachineName",
			definition: &StateMachineDefinition{
				Name:       "", // Invalid empty name
				Definition: `{"StartAt": "Test", "States": {"Test": {"Type": "Pass", "End": true}}}`,
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			options:       nil,
			mockSetup:     func(mockClient *MockSFNClientInterface) {},
			expectedError: true,
			errorContains: "invalid state machine name",
			validateResult: func(t *testing.T, result *StateMachineInfo) {
				assert.Nil(t, result, "Result should be nil on validation error")
			},
			description: "Invalid state machine name should fail validation",
		},
		{
			name:       "AWSServiceError",
			definition: createValidStateMachineDefinition(),
			options:    nil,
			mockSetup: func(mockClient *MockSFNClientInterface) {
				mockClient.On("CreateStateMachine", mock.Anything, mock.Anything).Return(nil, errors.New("AWS service unavailable"))
			},
			expectedError: true,
			errorContains: "failed to create state machine",
			validateResult: func(t *testing.T, result *StateMachineInfo) {
				assert.Nil(t, result, "Result should be nil on AWS error")
			},
			description: "AWS service error should be properly handled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockSFNClientInterface{}
			tc.mockSetup(mockClient)

			// Create a test implementation that uses our mock
			createStateMachineEWithClient := func(ctx *TestContext, client SFNClientInterface, definition *StateMachineDefinition, options *StateMachineOptions) (*StateMachineInfo, error) {
				start := time.Now()
				logOperation("CreateStateMachine", map[string]interface{}{
					"name": definition.Name,
					"type": definition.Type,
				})

				// Validate input parameters
				if err := validateStateMachineCreation(definition, options); err != nil {
					logResult("CreateStateMachine", false, time.Since(start), err)
					return nil, err
				}

				// Merge options with definition
				mergedOptions := mergeStateMachineOptions(options)

				input := &sfn.CreateStateMachineInput{
					Name:       aws.String(definition.Name),
					Definition: aws.String(definition.Definition),
					RoleArn:    aws.String(definition.RoleArn),
					Type:       definition.Type,
					Tags:       convertTagsToSFN(definition.Tags),
				}

				// Add optional configurations
				if mergedOptions.LoggingConfiguration != nil {
					input.LoggingConfiguration = mergedOptions.LoggingConfiguration
				}
				if mergedOptions.TracingConfiguration != nil {
					input.TracingConfiguration = mergedOptions.TracingConfiguration
				}

				result, err := client.CreateStateMachine(context.TODO(), input)
				if err != nil {
					logResult("CreateStateMachine", false, time.Since(start), err)
					return nil, fmt.Errorf("failed to create state machine: %w", err)
				}

				// Convert result to our standard format
				info := &StateMachineInfo{
					StateMachineArn:      aws.ToString(result.StateMachineArn),
					Name:                 definition.Name,
					Status:               types.StateMachineStatusActive,
					Definition:           definition.Definition,
					RoleArn:              definition.RoleArn,
					Type:                 definition.Type,
					CreationDate:         aws.ToTime(result.CreationDate),
					LoggingConfiguration: mergedOptions.LoggingConfiguration,
					TracingConfiguration: mergedOptions.TracingConfiguration,
				}

				logResult("CreateStateMachine", true, time.Since(start), nil)
				return info, nil
			}

			ctx := createTestContext()

			// Act
			result, err := createStateMachineEWithClient(ctx, mockClient, tc.definition, tc.options)

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

// TestStartExecutionEWithMock tests StartExecutionE with proper mocking
func TestStartExecutionEWithMock(t *testing.T) {
	tests := []struct {
		name             string
		stateMachineArn  string
		executionName    string
		input            *Input
		mockSetup        func(*MockSFNClientInterface)
		expectedError    bool
		errorContains    string
		validateResult   func(*testing.T, *ExecutionResult)
		description      string
	}{
		{
			name:            "SuccessfulStart",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			executionName:   "test-execution",
			input:           NewInput().Set("message", "Hello, World!"),
			mockSetup: func(mockClient *MockSFNClientInterface) {
				expectedOutput := &sfn.StartExecutionOutput{
					ExecutionArn: aws.String("arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution"),
					StartDate:    aws.Time(time.Now()),
				}
				mockClient.On("StartExecution", mock.Anything, mock.MatchedBy(func(input *sfn.StartExecutionInput) bool {
					return aws.ToString(input.StateMachineArn) == "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine" &&
						aws.ToString(input.Name) == "test-execution"
				})).Return(expectedOutput, nil)
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *ExecutionResult) {
				assert.NotNil(t, result, "Result should not be nil")
				assert.Equal(t, "test-execution", result.Name, "Name should match")
				assert.Equal(t, types.ExecutionStatusRunning, result.Status, "Status should be running")
				assert.Contains(t, result.ExecutionArn, "test-execution", "ARN should contain execution name")
			},
			description: "Valid execution start should succeed",
		},
		{
			name:            "EmptyStateMachineArn",
			stateMachineArn: "", // Invalid empty ARN
			executionName:   "test-execution",
			input:           NewInput(),
			mockSetup:       func(mockClient *MockSFNClientInterface) {},
			expectedError:   true,
			errorContains:   "state machine ARN is required",
			validateResult: func(t *testing.T, result *ExecutionResult) {
				assert.Nil(t, result, "Result should be nil on validation error")
			},
			description: "Empty state machine ARN should fail validation",
		},
		{
			name:            "InvalidExecutionName",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			executionName:   "invalid@name", // Invalid characters
			input:           NewInput(),
			mockSetup:       func(mockClient *MockSFNClientInterface) {},
			expectedError:   true,
			errorContains:   "name contains invalid characters",
			validateResult: func(t *testing.T, result *ExecutionResult) {
				assert.Nil(t, result, "Result should be nil on validation error")
			},
			description: "Invalid execution name should fail validation",
		},
		{
			name:            "AWSServiceError",
			stateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-machine",
			executionName:   "test-execution",
			input:           NewInput(),
			mockSetup: func(mockClient *MockSFNClientInterface) {
				mockClient.On("StartExecution", mock.Anything, mock.Anything).Return(nil, errors.New("execution limit exceeded"))
			},
			expectedError: true,
			errorContains: "failed to start execution",
			validateResult: func(t *testing.T, result *ExecutionResult) {
				assert.Nil(t, result, "Result should be nil on AWS error")
			},
			description: "AWS service error should be properly handled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockSFNClientInterface{}
			tc.mockSetup(mockClient)

			// Create a test implementation that uses our mock
			startExecutionEWithClient := func(ctx *TestContext, client SFNClientInterface, stateMachineArn, executionName string, input *Input) (*ExecutionResult, error) {
				start := time.Now()
				logOperation("StartExecution", map[string]interface{}{
					"state_machine": stateMachineArn,
					"name":          executionName,
				})

				// Build input JSON
				var inputJSON string
				var err error
				if input != nil && !input.isEmpty() {
					inputJSON, err = input.ToJSON()
					if err != nil {
						logResult("StartExecution", false, time.Since(start), err)
						return nil, fmt.Errorf("failed to convert input to JSON: %w", err)
					}
				}

				// Create request and validate
				request := &StartExecutionRequest{
					StateMachineArn: stateMachineArn,
					Name:            executionName,
					Input:           inputJSON,
				}

				if err := validateStartExecutionRequest(request); err != nil {
					logResult("StartExecution", false, time.Since(start), err)
					return nil, err
				}

				sfnInput := &sfn.StartExecutionInput{
					StateMachineArn: aws.String(request.StateMachineArn),
					Name:            aws.String(request.Name),
				}

				if request.Input != "" {
					sfnInput.Input = aws.String(request.Input)
				}

				result, err := client.StartExecution(context.TODO(), sfnInput)
				if err != nil {
					logResult("StartExecution", false, time.Since(start), err)
					return nil, fmt.Errorf("failed to start execution: %w", err)
				}

				// Convert result to our standard format
				execResult := &ExecutionResult{
					ExecutionArn:    aws.ToString(result.ExecutionArn),
					StateMachineArn: stateMachineArn,
					Name:            executionName,
					Status:          types.ExecutionStatusRunning,
					StartDate:       aws.ToTime(result.StartDate),
					Input:           inputJSON,
				}

				logResult("StartExecution", true, time.Since(start), nil)
				return processExecutionResult(execResult), nil
			}

			ctx := createTestContext()

			// Act
			result, err := startExecutionEWithClient(ctx, mockClient, tc.stateMachineArn, tc.executionName, tc.input)

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

// TestDescribeExecutionEWithMock tests DescribeExecutionE with proper mocking
func TestDescribeExecutionEWithMock(t *testing.T) {
	tests := []struct {
		name           string
		executionArn   string
		mockSetup      func(*MockSFNClientInterface)
		expectedError  bool
		errorContains  string
		validateResult func(*testing.T, *ExecutionResult)
		description    string
	}{
		{
			name:         "SuccessfulDescribe",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			mockSetup: func(mockClient *MockSFNClientInterface) {
				startTime := time.Now().Add(-5 * time.Minute)
				stopTime := time.Now()
				expectedOutput := &sfn.DescribeExecutionOutput{
					ExecutionArn:    aws.String("arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution"),
					StateMachineArn: aws.String("arn:aws:states:us-east-1:123456789012:stateMachine:test-machine"),
					Name:            aws.String("test-execution"),
					Status:          types.ExecutionStatusSucceeded,
					StartDate:       aws.Time(startTime),
					StopDate:        aws.Time(stopTime),
					Input:           aws.String(`{"message": "Hello, World!"}`),
					Output:          aws.String(`{"result": "success"}`),
				}
				mockClient.On("DescribeExecution", mock.Anything, mock.MatchedBy(func(input *sfn.DescribeExecutionInput) bool {
					return aws.ToString(input.ExecutionArn) == "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution"
				})).Return(expectedOutput, nil)
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *ExecutionResult) {
				assert.NotNil(t, result, "Result should not be nil")
				assert.Equal(t, "test-execution", result.Name, "Name should match")
				assert.Equal(t, types.ExecutionStatusSucceeded, result.Status, "Status should be succeeded")
				assert.Contains(t, result.Input, "Hello, World!", "Input should contain expected content")
				assert.Contains(t, result.Output, "success", "Output should contain expected content")
				assert.True(t, result.ExecutionTime > 0, "ExecutionTime should be calculated")
			},
			description: "Valid describe execution should succeed",
		},
		{
			name:          "EmptyExecutionArn",
			executionArn:  "", // Invalid empty ARN
			mockSetup:     func(mockClient *MockSFNClientInterface) {},
			expectedError: true,
			errorContains: "execution ARN is required",
			validateResult: func(t *testing.T, result *ExecutionResult) {
				assert.Nil(t, result, "Result should be nil on validation error")
			},
			description: "Empty execution ARN should fail validation",
		},
		{
			name:         "AWSServiceError",
			executionArn: "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution",
			mockSetup: func(mockClient *MockSFNClientInterface) {
				mockClient.On("DescribeExecution", mock.Anything, mock.Anything).Return(nil, errors.New("execution not found"))
			},
			expectedError: true,
			errorContains: "failed to describe execution",
			validateResult: func(t *testing.T, result *ExecutionResult) {
				assert.Nil(t, result, "Result should be nil on AWS error")
			},
			description: "AWS service error should be properly handled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockSFNClientInterface{}
			tc.mockSetup(mockClient)

			// Create a test implementation that uses our mock
			describeExecutionEWithClient := func(ctx *TestContext, client SFNClientInterface, executionArn string) (*ExecutionResult, error) {
				start := time.Now()
				logOperation("DescribeExecution", map[string]interface{}{
					"execution": executionArn,
				})

				// Validate input parameters
				if err := validateExecutionArn(executionArn); err != nil {
					logResult("DescribeExecution", false, time.Since(start), err)
					return nil, err
				}

				input := &sfn.DescribeExecutionInput{
					ExecutionArn: aws.String(executionArn),
				}

				result, err := client.DescribeExecution(context.TODO(), input)
				if err != nil {
					logResult("DescribeExecution", false, time.Since(start), err)
					return nil, fmt.Errorf("failed to describe execution: %w", err)
				}

				// Convert result to our standard format
				execResult := &ExecutionResult{
					ExecutionArn:    aws.ToString(result.ExecutionArn),
					StateMachineArn: aws.ToString(result.StateMachineArn),
					Name:            aws.ToString(result.Name),
					Status:          result.Status,
					StartDate:       aws.ToTime(result.StartDate),
					StopDate:        aws.ToTime(result.StopDate),
					Input:           aws.ToString(result.Input),
					Output:          aws.ToString(result.Output),
					Error:           aws.ToString(result.Error),
					Cause:           aws.ToString(result.Cause),
				}

				// Calculate execution time
				if !execResult.StartDate.IsZero() && !execResult.StopDate.IsZero() {
					execResult.ExecutionTime = execResult.StopDate.Sub(execResult.StartDate)
				}

				logResult("DescribeExecution", true, time.Since(start), nil)
				return processExecutionResult(execResult), nil
			}

			ctx := createTestContext()

			// Act
			result, err := describeExecutionEWithClient(ctx, mockClient, tc.executionArn)

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

// TestListStateMachinesEWithMock tests ListStateMachinesE with proper mocking  
func TestListStateMachinesEWithMock(t *testing.T) {
	tests := []struct {
		name           string
		maxResults     int32
		mockSetup      func(*MockSFNClientInterface)
		expectedError  bool
		errorContains  string
		validateResult func(*testing.T, []*StateMachineInfo)
		description    string
	}{
		{
			name:       "SuccessfulList",
			maxResults: 10,
			mockSetup: func(mockClient *MockSFNClientInterface) {
				expectedOutput := &sfn.ListStateMachinesOutput{
					StateMachines: []types.StateMachineListItem{
						{
							StateMachineArn: aws.String("arn:aws:states:us-east-1:123456789012:stateMachine:test-machine-1"),
							Name:            aws.String("test-machine-1"),
							Type:            types.StateMachineTypeStandard,
							CreationDate:    aws.Time(time.Now().Add(-24 * time.Hour)),
						},
						{
							StateMachineArn: aws.String("arn:aws:states:us-east-1:123456789012:stateMachine:test-machine-2"),
							Name:            aws.String("test-machine-2"),
							Type:            types.StateMachineTypeExpress,
							CreationDate:    aws.Time(time.Now().Add(-12 * time.Hour)),
						},
					},
				}
				mockClient.On("ListStateMachines", mock.Anything, mock.MatchedBy(func(input *sfn.ListStateMachinesInput) bool {
					return input.MaxResults == 10
				})).Return(expectedOutput, nil)
			},
			expectedError: false,
			validateResult: func(t *testing.T, result []*StateMachineInfo) {
				assert.NotNil(t, result, "Result should not be nil")
				assert.Len(t, result, 2, "Should return 2 state machines")
				assert.Equal(t, "test-machine-1", result[0].Name, "First machine name should match")
				assert.Equal(t, "test-machine-2", result[1].Name, "Second machine name should match")
				assert.Equal(t, types.StateMachineTypeStandard, result[0].Type, "First machine type should be standard")
				assert.Equal(t, types.StateMachineTypeExpress, result[1].Type, "Second machine type should be express")
			},
			description: "Valid list request should succeed",
		},
		{
			name:       "EmptyList",
			maxResults: 0, // Default max results
			mockSetup: func(mockClient *MockSFNClientInterface) {
				expectedOutput := &sfn.ListStateMachinesOutput{
					StateMachines: []types.StateMachineListItem{},
				}
				mockClient.On("ListStateMachines", mock.Anything, mock.AnythingOfType("*sfn.ListStateMachinesInput")).Return(expectedOutput, nil)
			},
			expectedError: false,
			validateResult: func(t *testing.T, result []*StateMachineInfo) {
				assert.NotNil(t, result, "Result should not be nil")
				assert.Len(t, result, 0, "Should return empty list")
			},
			description: "Empty list should be handled correctly",
		},
		{
			name:          "NegativeMaxResults",
			maxResults:    -1, // Invalid negative value
			mockSetup:     func(mockClient *MockSFNClientInterface) {},
			expectedError: true,
			errorContains: "max results cannot be negative",
			validateResult: func(t *testing.T, result []*StateMachineInfo) {
				assert.Nil(t, result, "Result should be nil on validation error")
			},
			description: "Negative max results should fail validation",
		},
		{
			name:       "AWSServiceError",
			maxResults: 10,
			mockSetup: func(mockClient *MockSFNClientInterface) {
				mockClient.On("ListStateMachines", mock.Anything, mock.Anything).Return(nil, errors.New("access denied"))
			},
			expectedError: true,
			errorContains: "failed to list state machines",
			validateResult: func(t *testing.T, result []*StateMachineInfo) {
				assert.Nil(t, result, "Result should be nil on AWS error")
			},
			description: "AWS service error should be properly handled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockSFNClientInterface{}
			tc.mockSetup(mockClient)

			// Create a test implementation that uses our mock
			listStateMachinesEWithClient := func(ctx *TestContext, client SFNClientInterface, maxResults int32) ([]*StateMachineInfo, error) {
				start := time.Now()
				logOperation("ListStateMachines", map[string]interface{}{
					"max_results": maxResults,
				})

				// Validate input parameters
				if err := validateMaxResults(maxResults); err != nil {
					logResult("ListStateMachines", false, time.Since(start), err)
					return nil, err
				}

				input := &sfn.ListStateMachinesInput{}
				if maxResults > 0 {
					input.MaxResults = maxResults
				}

				result, err := client.ListStateMachines(context.TODO(), input)
				if err != nil {
					logResult("ListStateMachines", false, time.Since(start), err)
					return nil, fmt.Errorf("failed to list state machines: %w", err)
				}

				// Convert results to our standard format - initialize empty slice instead of nil
				infos := make([]*StateMachineInfo, 0, len(result.StateMachines))
				for _, sm := range result.StateMachines {
					info := &StateMachineInfo{
						StateMachineArn: aws.ToString(sm.StateMachineArn),
						Name:            aws.ToString(sm.Name),
						Type:            sm.Type,
						CreationDate:    aws.ToTime(sm.CreationDate),
					}
					infos = append(infos, processStateMachineInfo(info))
				}

				logResult("ListStateMachines", true, time.Since(start), nil)
				return infos, nil
			}

			ctx := createTestContext()

			// Act
			result, err := listStateMachinesEWithClient(ctx, mockClient, tc.maxResults)

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