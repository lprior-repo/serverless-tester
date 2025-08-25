package lambda

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEventSourceMappingE(t *testing.T) {
	testCases := []struct {
		name           string
		config         EventSourceMappingConfig
		mockResponse   *lambda.CreateEventSourceMappingOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, result *EventSourceMappingResult)
	}{
		{
			name: "should_create_event_source_mapping_successfully",
			config: EventSourceMappingConfig{
				EventSourceArn:   "arn:aws:sqs:us-east-1:123456789012:test-queue",
				FunctionName:     "test-function",
				BatchSize:        10,
				StartingPosition: types.EventSourcePositionLatest,
				Enabled:          true,
			},
			mockResponse: &lambda.CreateEventSourceMappingOutput{
				UUID:            aws.String("12345-67890-abcdef"),
				EventSourceArn:  aws.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
				FunctionArn:     aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				LastModified:    aws.Time(time.Now()),
				State:           aws.String("Creating"),
				BatchSize:       aws.Int32(10),
				StartingPosition: types.EventSourcePositionLatest,
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *EventSourceMappingResult) {
				assert.Equal(t, "12345-67890-abcdef", result.UUID)
				assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:test-queue", result.EventSourceArn)
				assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:test-function", result.FunctionArn)
				assert.Equal(t, "Creating", result.State)
				assert.Equal(t, int32(10), result.BatchSize)
				assert.Equal(t, types.EventSourcePositionLatest, result.StartingPosition)
			},
		},
		{
			name: "should_return_error_for_invalid_configuration",
			config: EventSourceMappingConfig{
				EventSourceArn: "", // Empty ARN should fail validation
				FunctionName:   "test-function",
			},
			expectedError: true,
			errorContains: "event source ARN cannot be empty",
		},
		{
			name: "should_return_error_for_invalid_function_name",
			config: EventSourceMappingConfig{
				EventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
				FunctionName:   "invalid@function",
			},
			expectedError: true,
			errorContains: "invalid character '@'",
		},
		{
			name: "should_return_error_when_create_mapping_fails",
			config: EventSourceMappingConfig{
				EventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
				FunctionName:   "test-function",
			},
			mockError:     errors.New("ResourceConflictException: The resource already exists"),
			expectedError: true,
			errorContains: "failed to create event source mapping",
		},
		{
			name: "should_handle_kinesis_event_source",
			config: EventSourceMappingConfig{
				EventSourceArn:   "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
				FunctionName:     "test-function",
				BatchSize:        100,
				StartingPosition: types.EventSourcePositionTrimHorizon,
				Enabled:          true,
			},
			mockResponse: &lambda.CreateEventSourceMappingOutput{
				UUID:            aws.String("kinesis-mapping-uuid"),
				EventSourceArn:  aws.String("arn:aws:kinesis:us-east-1:123456789012:stream/test-stream"),
				FunctionArn:     aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				State:           aws.String("Enabled"),
				BatchSize:       aws.Int32(100),
				StartingPosition: types.EventSourcePositionTrimHorizon,
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *EventSourceMappingResult) {
				assert.Equal(t, "kinesis-mapping-uuid", result.UUID)
				assert.Contains(t, result.EventSourceArn, "kinesis")
				assert.Equal(t, int32(100), result.BatchSize)
				assert.Equal(t, types.EventSourcePositionTrimHorizon, result.StartingPosition)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				CreateEventSourceMappingResponse: tc.mockResponse,
				CreateEventSourceMappingError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := CreateEventSourceMappingE(ctx, tc.config)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}

			// Verify mock interactions for successful calls
			if !tc.expectedError && tc.config.FunctionName != "" && !strings.Contains(tc.config.FunctionName, "@") {
				assert.Len(t, mockClient.CreateEventSourceMappingCalls, 1)
				call := mockClient.CreateEventSourceMappingCalls[0]
				assert.Equal(t, tc.config.FunctionName, call.FunctionName)
				assert.Equal(t, tc.config.EventSourceArn, call.EventSourceArn)
			}
		})
	}
}

func TestGetEventSourceMappingE(t *testing.T) {
	testCases := []struct {
		name           string
		uuid           string
		mockResponse   *lambda.GetEventSourceMappingOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, result *EventSourceMappingResult)
	}{
		{
			name: "should_get_event_source_mapping_successfully",
			uuid: "12345-67890-abcdef",
			mockResponse: &lambda.GetEventSourceMappingOutput{
				UUID:                             aws.String("12345-67890-abcdef"),
				EventSourceArn:                   aws.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
				FunctionArn:                      aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				LastModified:                     aws.Time(time.Now()),
				LastProcessingResult:             aws.String("OK"),
				State:                            aws.String("Enabled"),
				StateTransitionReason:            aws.String("User action"),
				BatchSize:                        aws.Int32(10),
				MaximumBatchingWindowInSeconds:   aws.Int32(5),
				StartingPosition:                 types.EventSourcePositionLatest,
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *EventSourceMappingResult) {
				assert.Equal(t, "12345-67890-abcdef", result.UUID)
				assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:test-queue", result.EventSourceArn)
				assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:test-function", result.FunctionArn)
				assert.Equal(t, "OK", result.LastProcessingResult)
				assert.Equal(t, "Enabled", result.State)
				assert.Equal(t, "User action", result.StateTransitionReason)
				assert.Equal(t, int32(10), result.BatchSize)
				assert.Equal(t, int32(5), result.MaximumBatchingWindowInSeconds)
				assert.Equal(t, types.EventSourcePositionLatest, result.StartingPosition)
			},
		},
		{
			name:          "should_return_error_for_empty_uuid",
			uuid:          "",
			expectedError: true,
			errorContains: "event source mapping UUID cannot be empty",
		},
		{
			name:          "should_return_error_when_get_mapping_fails",
			uuid:          "12345-67890-abcdef",
			mockError:     errors.New("ResourceNotFoundException: The resource you requested does not exist"),
			expectedError: true,
			errorContains: "failed to get event source mapping",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetEventSourceMappingResponse: tc.mockResponse,
				GetEventSourceMappingError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := GetEventSourceMappingE(ctx, tc.uuid)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}

			// Verify mock interactions for successful calls
			if !tc.expectedError && tc.uuid != "" {
				assert.Contains(t, mockClient.GetEventSourceMappingCalls, tc.uuid)
			}
		})
	}
}

func TestListEventSourceMappingsE(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		mockResponse   *lambda.ListEventSourceMappingsOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, results []EventSourceMappingResult)
	}{
		{
			name:         "should_list_event_source_mappings_successfully",
			functionName: "test-function",
			mockResponse: &lambda.ListEventSourceMappingsOutput{
				EventSourceMappings: []types.EventSourceMappingConfiguration{
					{
						UUID:           aws.String("mapping-1"),
						EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:queue-1"),
						FunctionArn:    aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
						State:          aws.String("Enabled"),
						BatchSize:      aws.Int32(10),
					},
					{
						UUID:           aws.String("mapping-2"),
						EventSourceArn: aws.String("arn:aws:kinesis:us-east-1:123456789012:stream/stream-1"),
						FunctionArn:    aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
						State:          aws.String("Disabled"),
						BatchSize:      aws.Int32(100),
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, results []EventSourceMappingResult) {
				assert.Len(t, results, 2)
				assert.Equal(t, "mapping-1", results[0].UUID)
				assert.Contains(t, results[0].EventSourceArn, "sqs")
				assert.Equal(t, "Enabled", results[0].State)
				assert.Equal(t, int32(10), results[0].BatchSize)
				
				assert.Equal(t, "mapping-2", results[1].UUID)
				assert.Contains(t, results[1].EventSourceArn, "kinesis")
				assert.Equal(t, "Disabled", results[1].State)
				assert.Equal(t, int32(100), results[1].BatchSize)
			},
		},
		{
			name:         "should_return_empty_list_when_no_mappings",
			functionName: "test-function",
			mockResponse: &lambda.ListEventSourceMappingsOutput{
				EventSourceMappings: []types.EventSourceMappingConfiguration{},
			},
			expectedError: false,
			validateResult: func(t *testing.T, results []EventSourceMappingResult) {
				assert.Len(t, results, 0)
			},
		},
		{
			name:          "should_return_error_for_invalid_function_name",
			functionName:  "invalid@function",
			expectedError: true,
			errorContains: "invalid character '@'",
		},
		{
			name:          "should_return_error_when_list_mappings_fails",
			functionName:  "test-function",
			mockError:     errors.New("AccessDeniedException: User is not authorized"),
			expectedError: true,
			errorContains: "failed to list event source mappings",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				ListEventSourceMappingsResponse: tc.mockResponse,
				ListEventSourceMappingsError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := ListEventSourceMappingsE(ctx, tc.functionName)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}

			// Verify mock interactions for successful calls
			if !tc.expectedError && tc.functionName != "" && !strings.Contains(tc.functionName, "@") {
				assert.Contains(t, mockClient.ListEventSourceMappingsCalls, tc.functionName)
			}
		})
	}
}

func TestUpdateEventSourceMappingE(t *testing.T) {
	testCases := []struct {
		name           string
		uuid           string
		config         UpdateEventSourceMappingConfig
		mockResponse   *lambda.UpdateEventSourceMappingOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, result *EventSourceMappingResult)
	}{
		{
			name: "should_update_event_source_mapping_successfully",
			uuid: "12345-67890-abcdef",
			config: UpdateEventSourceMappingConfig{
				BatchSize: aws.Int32(20),
				Enabled:   aws.Bool(false),
			},
			mockResponse: &lambda.UpdateEventSourceMappingOutput{
				UUID:       aws.String("12345-67890-abcdef"),
				State:      aws.String("Disabled"),
				BatchSize:  aws.Int32(20),
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *EventSourceMappingResult) {
				assert.Equal(t, "12345-67890-abcdef", result.UUID)
				assert.Equal(t, "Disabled", result.State)
				assert.Equal(t, int32(20), result.BatchSize)
			},
		},
		{
			name:   "should_return_error_for_empty_uuid",
			uuid:   "",
			config: UpdateEventSourceMappingConfig{},
			expectedError: true,
			errorContains: "event source mapping UUID cannot be empty",
		},
		{
			name: "should_return_error_when_update_mapping_fails",
			uuid: "12345-67890-abcdef",
			config: UpdateEventSourceMappingConfig{
				BatchSize: aws.Int32(20),
			},
			mockError:     errors.New("ResourceNotFoundException: The resource you requested does not exist"),
			expectedError: true,
			errorContains: "failed to update event source mapping",
		},
		{
			name: "should_update_multiple_fields",
			uuid: "12345-67890-abcdef",
			config: UpdateEventSourceMappingConfig{
				BatchSize:                      aws.Int32(50),
				MaximumBatchingWindowInSeconds: aws.Int32(10),
				Enabled:                        aws.Bool(true),
				FunctionName:                   aws.String("updated-function"),
			},
			mockResponse: &lambda.UpdateEventSourceMappingOutput{
				UUID:                           aws.String("12345-67890-abcdef"),
				State:                          aws.String("Enabled"),
				BatchSize:                      aws.Int32(50),
				MaximumBatchingWindowInSeconds: aws.Int32(10),
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *EventSourceMappingResult) {
				assert.Equal(t, "12345-67890-abcdef", result.UUID)
				assert.Equal(t, "Enabled", result.State)
				assert.Equal(t, int32(50), result.BatchSize)
				assert.Equal(t, int32(10), result.MaximumBatchingWindowInSeconds)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				UpdateEventSourceMappingResponse: tc.mockResponse,
				UpdateEventSourceMappingError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := UpdateEventSourceMappingE(ctx, tc.uuid, tc.config)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}

			// Verify mock interactions for successful calls
			if !tc.expectedError && tc.uuid != "" {
				assert.Len(t, mockClient.UpdateEventSourceMappingCalls, 1)
				call := mockClient.UpdateEventSourceMappingCalls[0]
				assert.Equal(t, tc.uuid, call.UUID)
			}
		})
	}
}

func TestDeleteEventSourceMappingE(t *testing.T) {
	testCases := []struct {
		name           string
		uuid           string
		mockResponse   *lambda.DeleteEventSourceMappingOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, result *EventSourceMappingResult)
	}{
		{
			name: "should_delete_event_source_mapping_successfully",
			uuid: "12345-67890-abcdef",
			mockResponse: &lambda.DeleteEventSourceMappingOutput{
				UUID:  aws.String("12345-67890-abcdef"),
				State: aws.String("Deleting"),
			},
			expectedError: false,
			validateResult: func(t *testing.T, result *EventSourceMappingResult) {
				assert.Equal(t, "12345-67890-abcdef", result.UUID)
				assert.Equal(t, "Deleting", result.State)
			},
		},
		{
			name:          "should_return_error_for_empty_uuid",
			uuid:          "",
			expectedError: true,
			errorContains: "event source mapping UUID cannot be empty",
		},
		{
			name:          "should_return_error_when_delete_mapping_fails",
			uuid:          "12345-67890-abcdef",
			mockError:     errors.New("ResourceNotFoundException: The resource you requested does not exist"),
			expectedError: true,
			errorContains: "failed to delete event source mapping",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				DeleteEventSourceMappingResponse: tc.mockResponse,
				DeleteEventSourceMappingError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := DeleteEventSourceMappingE(ctx, tc.uuid)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}

			// Verify mock interactions for successful calls
			if !tc.expectedError && tc.uuid != "" {
				assert.Contains(t, mockClient.DeleteEventSourceMappingCalls, tc.uuid)
			}
		})
	}
}

func TestWaitForEventSourceMappingStateE(t *testing.T) {
	t.Run("should_return_immediately_when_mapping_is_in_expected_state", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			GetEventSourceMappingResponse: &lambda.GetEventSourceMappingOutput{
				UUID:  aws.String("12345-67890-abcdef"),
				State: aws.String("Enabled"),
			},
		}
		mockFactory := &MockClientFactory{
			LambdaClient: mockClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		err := WaitForEventSourceMappingStateE(ctx, "12345-67890-abcdef", "Enabled", 5*time.Second)

		// Assert
		assert.NoError(t, err)
		assert.Contains(t, mockClient.GetEventSourceMappingCalls, "12345-67890-abcdef")
	})

	t.Run("should_timeout_when_mapping_never_reaches_expected_state", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			GetEventSourceMappingResponse: &lambda.GetEventSourceMappingOutput{
				UUID:  aws.String("12345-67890-abcdef"),
				State: aws.String("Creating"),
			},
		}
		mockFactory := &MockClientFactory{
			LambdaClient: mockClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		start := time.Now()
		err := WaitForEventSourceMappingStateE(ctx, "12345-67890-abcdef", "Enabled", 100*time.Millisecond)
		duration := time.Since(start)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout waiting for event source mapping")
		assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
		assert.Less(t, duration, 500*time.Millisecond) // Should not take too long
	})

	t.Run("should_return_error_when_get_mapping_fails", func(t *testing.T) {
		// Arrange
		mockClient := &MockLambdaClient{
			GetEventSourceMappingError: errors.New("AccessDeniedException: User is not authorized"),
		}
		mockFactory := &MockClientFactory{
			LambdaClient: mockClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		err := WaitForEventSourceMappingStateE(ctx, "12345-67890-abcdef", "Enabled", 5*time.Second)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AccessDeniedException")
	})
}

func TestValidateEventSourceMappingConfig(t *testing.T) {
	testCases := []struct {
		name          string
		config        *EventSourceMappingConfig
		expectedError bool
		errorContains string
	}{
		{
			name: "should_accept_valid_sqs_configuration",
			config: &EventSourceMappingConfig{
				EventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
				FunctionName:   "test-function",
				BatchSize:      10,
				Enabled:        true,
			},
			expectedError: false,
		},
		{
			name: "should_accept_valid_kinesis_configuration",
			config: &EventSourceMappingConfig{
				EventSourceArn:   "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
				FunctionName:     "test-function",
				BatchSize:        100,
				StartingPosition: types.EventSourcePositionTrimHorizon,
				Enabled:          true,
			},
			expectedError: false,
		},
		{
			name: "should_accept_valid_dynamodb_configuration",
			config: &EventSourceMappingConfig{
				EventSourceArn:   "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2023-01-01T00:00:00.000",
				FunctionName:     "test-function",
				BatchSize:        50,
				StartingPosition: types.EventSourcePositionLatest,
				Enabled:          true,
			},
			expectedError: false,
		},
		{
			name:          "should_return_error_for_nil_configuration",
			config:        nil,
			expectedError: true,
			errorContains: "event source mapping configuration cannot be nil",
		},
		{
			name: "should_return_error_for_empty_event_source_arn",
			config: &EventSourceMappingConfig{
				EventSourceArn: "",
				FunctionName:   "test-function",
			},
			expectedError: true,
			errorContains: "event source ARN cannot be empty",
		},
		{
			name: "should_return_error_for_empty_function_name",
			config: &EventSourceMappingConfig{
				EventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
				FunctionName:   "",
			},
			expectedError: true,
			errorContains: "function name cannot be empty",
		},
		{
			name: "should_return_error_for_invalid_function_name",
			config: &EventSourceMappingConfig{
				EventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
				FunctionName:   "invalid@function",
			},
			expectedError: true,
			errorContains: "invalid character '@'",
		},
		{
			name: "should_return_error_for_excessive_sqs_batch_size",
			config: &EventSourceMappingConfig{
				EventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
				FunctionName:   "test-function",
				BatchSize:      1001, // Exceeds SQS limit
			},
			expectedError: true,
			errorContains: "SQS batch size 1001 exceeds maximum allowed (1000)",
		},
		{
			name: "should_return_error_for_excessive_kinesis_batch_size",
			config: &EventSourceMappingConfig{
				EventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
				FunctionName:   "test-function",
				BatchSize:      10001, // Exceeds Kinesis limit
			},
			expectedError: true,
			errorContains: "Kinesis batch size 10001 exceeds maximum allowed (10000)",
		},
		{
			name: "should_return_error_for_excessive_dynamodb_batch_size",
			config: &EventSourceMappingConfig{
				EventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2023-01-01T00:00:00.000",
				FunctionName:   "test-function",
				BatchSize:      1001, // Exceeds DynamoDB limit
			},
			expectedError: true,
			errorContains: "DynamoDB batch size 1001 exceeds maximum allowed (1000)",
		},
		{
			name: "should_return_error_for_excessive_unknown_batch_size",
			config: &EventSourceMappingConfig{
				EventSourceArn: "arn:aws:unknown:us-east-1:123456789012:resource/test",
				FunctionName:   "test-function",
				BatchSize:      1001, // Exceeds default limit
			},
			expectedError: true,
			errorContains: "batch size 1001 exceeds maximum allowed (1000)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateEventSourceMappingConfig(tc.config)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEventSourceType(t *testing.T) {
	testCases := []struct {
		name           string
		eventSourceArn string
		expected       string
	}{
		{
			name:           "should_identify_sqs_event_source",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			expected:       "sqs",
		},
		{
			name:           "should_identify_kinesis_event_source",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
			expected:       "kinesis",
		},
		{
			name:           "should_identify_dynamodb_event_source",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2023-01-01T00:00:00.000",
			expected:       "dynamodb",
		},
		{
			name:           "should_return_unknown_for_unrecognized_event_source",
			eventSourceArn: "arn:aws:s3:::test-bucket",
			expected:       "unknown",
		},
		{
			name:           "should_handle_empty_arn",
			eventSourceArn: "",
			expected:       "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := getEventSourceType(tc.eventSourceArn)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateBatchConfigurationForEventSource(t *testing.T) {
	testCases := []struct {
		name           string
		eventSourceArn string
		config         BatchConfigurationOptions
		expectedError  bool
		errorContains  string
	}{
		{
			name:           "should_accept_valid_sqs_batch_configuration",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			config: BatchConfigurationOptions{
				BatchSize:                        10,
				MaximumBatchingWindowInSeconds:   5,
			},
			expectedError: false,
		},
		{
			name:           "should_accept_valid_kinesis_batch_configuration",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
			config: BatchConfigurationOptions{
				BatchSize:                        100,
				ParallelizationFactor:           10,
				MaximumBatchingWindowInSeconds:   10,
			},
			expectedError: false,
		},
		{
			name:           "should_accept_valid_dynamodb_batch_configuration",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2023-01-01T00:00:00.000",
			config: BatchConfigurationOptions{
				BatchSize:             50,
				ParallelizationFactor: 5,
			},
			expectedError: false,
		},
		{
			name:           "should_reject_invalid_sqs_batch_size",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			config: BatchConfigurationOptions{
				BatchSize: 1001, // Exceeds SQS limit
			},
			expectedError: true,
			errorContains: "SQS batch size must be between 1 and 1000",
		},
		{
			name:           "should_reject_invalid_sqs_batching_window",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			config: BatchConfigurationOptions{
				BatchSize:                        10,
				MaximumBatchingWindowInSeconds:   301, // Exceeds SQS limit
			},
			expectedError: true,
			errorContains: "SQS maximum batching window must be between 0 and 300 seconds",
		},
		{
			name:           "should_reject_invalid_kinesis_parallelization_factor",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
			config: BatchConfigurationOptions{
				BatchSize:             100,
				ParallelizationFactor: 101, // Exceeds Kinesis limit
			},
			expectedError: true,
			errorContains: "Kinesis parallelization factor must be between 1 and 100",
		},
		{
			name:           "should_reject_invalid_dynamodb_parallelization_factor",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2023-01-01T00:00:00.000",
			config: BatchConfigurationOptions{
				BatchSize:             50,
				ParallelizationFactor: 101, // Exceeds DynamoDB limit
			},
			expectedError: true,
			errorContains: "DynamoDB parallelization factor must be between 1 and 100",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := validateBatchConfigurationForEventSource(tc.eventSourceArn, tc.config)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOptimizeBatchConfigurationForEventSource(t *testing.T) {
	testCases := []struct {
		name           string
		eventSourceArn string
		input          BatchConfigurationOptions
		validateOutput func(t *testing.T, output BatchConfigurationOptions)
	}{
		{
			name:           "should_optimize_sqs_configuration",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			input:          BatchConfigurationOptions{}, // All zero values
			validateOutput: func(t *testing.T, output BatchConfigurationOptions) {
				assert.Equal(t, int32(10), output.BatchSize) // SQS optimal default
				assert.Equal(t, int32(1), output.MaximumBatchingWindowInSeconds)
			},
		},
		{
			name:           "should_optimize_kinesis_configuration",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
			input:          BatchConfigurationOptions{}, // All zero values
			validateOutput: func(t *testing.T, output BatchConfigurationOptions) {
				assert.Equal(t, int32(100), output.BatchSize) // Kinesis optimal default
				assert.Equal(t, int32(10), output.ParallelizationFactor)
				assert.Equal(t, int32(5), output.MaximumBatchingWindowInSeconds)
			},
		},
		{
			name:           "should_optimize_dynamodb_configuration",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2023-01-01T00:00:00.000",
			input:          BatchConfigurationOptions{}, // All zero values
			validateOutput: func(t *testing.T, output BatchConfigurationOptions) {
				assert.Equal(t, int32(50), output.BatchSize) // DynamoDB optimal default
				assert.Equal(t, int32(1), output.ParallelizationFactor) // Conservative for consistency
			},
		},
		{
			name:           "should_preserve_non_zero_user_values",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			input: BatchConfigurationOptions{
				BatchSize:                        25, // User-specified value
				MaximumBatchingWindowInSeconds:   0,  // Should be optimized
			},
			validateOutput: func(t *testing.T, output BatchConfigurationOptions) {
				assert.Equal(t, int32(25), output.BatchSize) // Preserved user value
				assert.Equal(t, int32(1), output.MaximumBatchingWindowInSeconds) // Optimized
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := optimizeBatchConfigurationForEventSource(tc.eventSourceArn, tc.input)

			// Assert
			tc.validateOutput(t, result)
		})
	}
}

func TestGetOptimalThroughputConfiguration(t *testing.T) {
	testCases := []struct {
		name           string
		eventSourceArn string
		validateConfig func(t *testing.T, config BatchConfigurationOptions)
	}{
		{
			name:           "should_return_optimal_sqs_throughput_configuration",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			validateConfig: func(t *testing.T, config BatchConfigurationOptions) {
				assert.Equal(t, int32(1000), config.BatchSize) // Max SQS batch size
				assert.Equal(t, int32(5), config.MaximumBatchingWindowInSeconds) // Buffer for full batches
			},
		},
		{
			name:           "should_return_optimal_kinesis_throughput_configuration",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
			validateConfig: func(t *testing.T, config BatchConfigurationOptions) {
				assert.Equal(t, int32(10000), config.BatchSize) // Max Kinesis batch size
				assert.Equal(t, int32(100), config.ParallelizationFactor) // Max parallelization
				assert.Equal(t, int32(0), config.MaximumBatchingWindowInSeconds) // No buffering for max throughput
			},
		},
		{
			name:           "should_return_optimal_dynamodb_throughput_configuration",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2023-01-01T00:00:00.000",
			validateConfig: func(t *testing.T, config BatchConfigurationOptions) {
				assert.Equal(t, int32(1000), config.BatchSize) // Max DynamoDB batch size
				assert.Equal(t, int32(100), config.ParallelizationFactor) // Max parallelization
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			config := getOptimalThroughputConfiguration(tc.eventSourceArn)

			// Assert
			tc.validateConfig(t, config)
		})
	}
}

func TestGetOptimalLatencyConfiguration(t *testing.T) {
	testCases := []struct {
		name           string
		eventSourceArn string
		validateConfig func(t *testing.T, config BatchConfigurationOptions)
	}{
		{
			name:           "should_return_optimal_sqs_latency_configuration",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			validateConfig: func(t *testing.T, config BatchConfigurationOptions) {
				assert.Equal(t, int32(1), config.BatchSize) // Min batch size for lowest latency
				assert.Equal(t, int32(0), config.MaximumBatchingWindowInSeconds) // No buffering
			},
		},
		{
			name:           "should_return_optimal_kinesis_latency_configuration",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
			validateConfig: func(t *testing.T, config BatchConfigurationOptions) {
				assert.Equal(t, int32(1), config.BatchSize) // Min batch size
				assert.Equal(t, int32(100), config.ParallelizationFactor) // Max parallelization for fast processing
				assert.Equal(t, int32(0), config.MaximumBatchingWindowInSeconds) // No buffering
			},
		},
		{
			name:           "should_return_optimal_dynamodb_latency_configuration",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/stream/2023-01-01T00:00:00.000",
			validateConfig: func(t *testing.T, config BatchConfigurationOptions) {
				assert.Equal(t, int32(1), config.BatchSize) // Min batch size
				assert.Equal(t, int32(1), config.ParallelizationFactor) // Sequential processing for consistency
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			config := getOptimalLatencyConfiguration(tc.eventSourceArn)

			// Assert
			tc.validateConfig(t, config)
		})
	}
}

func TestEventSourceMappingExistsE(t *testing.T) {
	testCases := []struct {
		name         string
		uuid         string
		mockResponse *lambda.GetEventSourceMappingOutput
		mockError    error
		expected     bool
		expectError  bool
	}{
		{
			name: "should_return_true_when_mapping_exists",
			uuid: "existing-mapping",
			mockResponse: &lambda.GetEventSourceMappingOutput{
				UUID: aws.String("existing-mapping"),
			},
			expected: true,
		},
		{
			name:      "should_return_false_when_mapping_not_found",
			uuid:      "non-existent-mapping",
			mockError: errors.New("ResourceNotFoundException: The resource you requested does not exist"),
			expected:  false,
		},
		{
			name:      "should_return_false_for_does_not_exist_error",
			uuid:      "non-existent-mapping",
			mockError: errors.New("Event source mapping does not exist"),
			expected:  false,
		},
		{
			name:        "should_return_error_for_other_errors",
			uuid:        "test-mapping",
			mockError:   errors.New("AccessDeniedException: User is not authorized"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := &MockLambdaClient{
				GetEventSourceMappingResponse: tc.mockResponse,
				GetEventSourceMappingError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				LambdaClient: mockClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := EventSourceMappingExistsE(ctx, tc.uuid)

			// Assert
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}