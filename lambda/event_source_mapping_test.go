package lambda

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateEventSourceMapping_Success tests successful event source mapping creation
func TestCreateEventSourceMapping_Success(t *testing.T) {
	// Red Phase - Define expected behavior
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	// Setup expected response
	expectedUUID := "test-uuid-123"
	expectedState := "Enabled"
	expectedLastModified := time.Now()
	
	mockClient.CreateEventSourceMappingResponse = &lambda.CreateEventSourceMappingOutput{
		UUID:          aws.String(expectedUUID),
		EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
		FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
		State:         aws.String(expectedState),
		LastModified:  &expectedLastModified,
		BatchSize:     aws.Int32(10),
		StartingPosition: types.EventSourcePositionLatest,
	}

	config := EventSourceMappingConfig{
		EventSourceArn:   "arn:aws:sqs:us-east-1:123456789012:test-queue",
		FunctionName:     "test-function",
		BatchSize:        10,
		Enabled:          true,
		StartingPosition: types.EventSourcePositionLatest,
	}

	// Act
	result := CreateEventSourceMapping(ctx, config)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, expectedUUID, result.UUID)
	assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:test-queue", result.EventSourceArn)
	assert.Equal(t, expectedState, result.State)
	assert.Equal(t, int32(10), result.BatchSize)

	// Verify mock was called correctly
	require.Len(t, mockClient.CreateEventSourceMappingCalls, 1)
	call := mockClient.CreateEventSourceMappingCalls[0]
	assert.Equal(t, "test-function", call.FunctionName)
	assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:test-queue", call.EventSourceArn)
	assert.Equal(t, int32(10), *call.BatchSize)
	assert.True(t, *call.Enabled)
}

// TestCreateEventSourceMappingE_ValidationError tests validation errors
func TestCreateEventSourceMappingE_ValidationError(t *testing.T) {
	ctx := &TestContext{T: t}

	testCases := []struct {
		name           string
		config         EventSourceMappingConfig
		expectedError  string
	}{
		{
			name:          "empty event source ARN",
			config:        EventSourceMappingConfig{FunctionName: "test-function"},
			expectedError: "event source ARN cannot be empty",
		},
		{
			name:          "empty function name", 
			config:        EventSourceMappingConfig{EventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue"},
			expectedError: "function name cannot be empty",
		},
		{
			name: "batch size too large",
			config: EventSourceMappingConfig{
				EventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
				FunctionName:   "test-function", 
				BatchSize:      2000,
			},
			expectedError: "batch size 2000 exceeds maximum allowed (1000)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CreateEventSourceMappingE(ctx, tc.config)
			
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// TestGetEventSourceMapping_Success tests successful retrieval of event source mapping
func TestGetEventSourceMapping_Success(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	expectedUUID := "test-uuid-123"
	expectedLastModified := time.Now()
	
	mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
		UUID:          aws.String(expectedUUID),
		EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
		FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
		State:         aws.String("Enabled"),
		LastModified:  &expectedLastModified,
		BatchSize:     aws.Int32(50),
		MaximumBatchingWindowInSeconds: aws.Int32(5),
	}

	// Act
	result := GetEventSourceMapping(ctx, expectedUUID)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, expectedUUID, result.UUID)
	assert.Equal(t, "Enabled", result.State)
	assert.Equal(t, int32(50), result.BatchSize)
	assert.Equal(t, int32(5), result.MaximumBatchingWindowInSeconds)

	// Verify mock was called
	require.Len(t, mockClient.GetEventSourceMappingCalls, 1)
	assert.Equal(t, expectedUUID, mockClient.GetEventSourceMappingCalls[0])
}

// TestGetEventSourceMappingE_EmptyUUID tests validation for empty UUID
func TestGetEventSourceMappingE_EmptyUUID(t *testing.T) {
	ctx := &TestContext{T: t}

	result, err := GetEventSourceMappingE(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "event source mapping UUID cannot be empty")
}

// TestListEventSourceMappings_Success tests successful listing of event source mappings
func TestListEventSourceMappings_Success(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	functionName := "test-function"
	
	// Setup multiple mappings response
	mappings := []types.EventSourceMappingConfiguration{
		{
			UUID:          aws.String("uuid-1"),
			EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:queue-1"),
			FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
			State:         aws.String("Enabled"),
			BatchSize:     aws.Int32(10),
		},
		{
			UUID:          aws.String("uuid-2"), 
			EventSourceArn: aws.String("arn:aws:kinesis:us-east-1:123456789012:stream/test-stream"),
			FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
			State:         aws.String("Disabled"),
			BatchSize:     aws.Int32(100),
		},
	}
	
	mockClient.ListEventSourceMappingsResponse = &lambda.ListEventSourceMappingsOutput{
		EventSourceMappings: mappings,
	}

	// Act
	results := ListEventSourceMappings(ctx, functionName)

	// Assert
	require.Len(t, results, 2)
	assert.Equal(t, "uuid-1", results[0].UUID)
	assert.Equal(t, "Enabled", results[0].State)
	assert.Equal(t, "uuid-2", results[1].UUID) 
	assert.Equal(t, "Disabled", results[1].State)

	// Verify mock was called
	require.Len(t, mockClient.ListEventSourceMappingsCalls, 1)
	assert.Equal(t, functionName, mockClient.ListEventSourceMappingsCalls[0])
}

// TestUpdateEventSourceMapping_Success tests successful updating of event source mapping
func TestUpdateEventSourceMapping_Success(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	uuid := "test-uuid-123"
	newBatchSize := int32(25)
	newEnabled := false
	
	mockClient.UpdateEventSourceMappingResponse = &lambda.UpdateEventSourceMappingOutput{
		UUID:       aws.String(uuid),
		BatchSize:  aws.Int32(newBatchSize),
		State:      aws.String("Disabled"),
	}

	config := UpdateEventSourceMappingConfig{
		BatchSize: &newBatchSize,
		Enabled:   &newEnabled,
	}

	// Act
	result := UpdateEventSourceMapping(ctx, uuid, config)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, uuid, result.UUID)
	assert.Equal(t, newBatchSize, result.BatchSize)

	// Verify mock was called correctly
	require.Len(t, mockClient.UpdateEventSourceMappingCalls, 1)
	call := mockClient.UpdateEventSourceMappingCalls[0]
	assert.Equal(t, uuid, call.UUID)
	assert.Equal(t, newBatchSize, *call.BatchSize)
	assert.Equal(t, newEnabled, *call.Enabled)
}

// TestDeleteEventSourceMapping_Success tests successful deletion of event source mapping
func TestDeleteEventSourceMapping_Success(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	uuid := "test-uuid-123"
	
	mockClient.DeleteEventSourceMappingResponse = &lambda.DeleteEventSourceMappingOutput{
		UUID:  aws.String(uuid),
		State: aws.String("Deleting"),
	}

	// Act
	result := DeleteEventSourceMapping(ctx, uuid)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, uuid, result.UUID)
	assert.Equal(t, "Deleting", result.State)

	// Verify mock was called
	require.Len(t, mockClient.DeleteEventSourceMappingCalls, 1)
	assert.Equal(t, uuid, mockClient.DeleteEventSourceMappingCalls[0])
}

// TestWaitForEventSourceMappingState_Success tests waiting for state change
func TestWaitForEventSourceMappingState_Success(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	uuid := "test-uuid-123"
	expectedState := "Enabled"
	
	// First call returns "Creating", second call returns "Enabled"
	mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
		UUID:  aws.String(uuid),
		State: aws.String("Creating"), // This will change in our test logic
	}

	// We need to test the actual wait functionality, but since the implementation
	// calls GetEventSourceMappingE internally, we'll test with a short timeout
	go func() {
		time.Sleep(100 * time.Millisecond)
		mockClient.GetEventSourceMappingResponse.State = aws.String("Enabled")
	}()

	// Act - use a short timeout for testing
	err := WaitForEventSourceMappingStateE(ctx, uuid, expectedState, 1*time.Second)

	// Assert
	assert.NoError(t, err)
	// Verify GetEventSourceMapping was called at least once
	assert.GreaterOrEqual(t, len(mockClient.GetEventSourceMappingCalls), 1)
}

// TestWaitForEventSourceMappingState_Timeout tests timeout behavior
func TestWaitForEventSourceMappingState_Timeout(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	uuid := "test-uuid-123"
	expectedState := "Enabled"
	
	// Always return "Creating" state to trigger timeout
	mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
		UUID:  aws.String(uuid),
		State: aws.String("Creating"),
	}

	// Act
	err := WaitForEventSourceMappingStateE(ctx, uuid, expectedState, 100*time.Millisecond)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for event source mapping")
	assert.Contains(t, err.Error(), uuid)
	assert.Contains(t, err.Error(), expectedState)
}

// TestEventSourceMappingBatchConfigurationManagement tests higher-level batch management
func TestEventSourceMappingBatchConfigurationManagement(t *testing.T) {

	t.Run("OptimizeBatchSizeForThroughput", func(t *testing.T) {
		// This test will validate that we can configure batch sizes for optimal throughput
		// across different event source types
		testCases := []struct {
			eventSourceType string
			expectedBatchSize int32
		}{
			{"sqs", 10},      // SQS optimal batch size
			{"kinesis", 100}, // Kinesis optimal batch size  
			{"dynamodb", 50}, // DynamoDB streams optimal batch size
		}

		for _, tc := range testCases {
			t.Run(tc.eventSourceType, func(t *testing.T) {
				batchSize := optimizeBatchSizeForEventSource(tc.eventSourceType)
				assert.Equal(t, tc.expectedBatchSize, batchSize)
			})
		}
	})

	t.Run("BatchSizeValidationByEventSource", func(t *testing.T) {
		// Test that batch sizes are properly validated for different event sources
		testCases := []struct {
			eventSource string
			batchSize   int32
			shouldPass  bool
		}{
			{"arn:aws:sqs:us-east-1:123456789012:queue", 1, true},    // SQS min
			{"arn:aws:sqs:us-east-1:123456789012:queue", 10, true},   // SQS normal
			{"arn:aws:sqs:us-east-1:123456789012:queue", 1000, true}, // SQS max
			{"arn:aws:sqs:us-east-1:123456789012:queue", 1001, false}, // SQS over max
			{"arn:aws:kinesis:us-east-1:123456789012:stream/test", 1, true},   // Kinesis min
			{"arn:aws:kinesis:us-east-1:123456789012:stream/test", 100, true}, // Kinesis normal
			{"arn:aws:kinesis:us-east-1:123456789012:stream/test", 10000, true}, // Kinesis max
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s_batch_%d", tc.eventSource, tc.batchSize), func(t *testing.T) {
				config := EventSourceMappingConfig{
					EventSourceArn: tc.eventSource,
					FunctionName:   "test-function",
					BatchSize:      tc.batchSize,
				}

				err := validateEventSourceMappingConfig(&config)
				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}

// TestEventSourceMappingParallelizationConfiguration tests parallelization settings
func TestEventSourceMappingParallelizationConfiguration(t *testing.T) {
	t.Run("ParallelizationFactorValidation", func(t *testing.T) {
		// Test that parallelization factor is properly validated
		testCases := []struct {
			parallelizationFactor int32
			shouldPass            bool
		}{
			{1, true},   // Min parallelization
			{10, true},  // Normal parallelization
			{100, true}, // Max parallelization
			{101, false}, // Over max
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("parallelization_%d", tc.parallelizationFactor), func(t *testing.T) {
				err := validateParallelizationFactor(tc.parallelizationFactor)
				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}

// TestEventSourceMappingFilterConfiguration tests filtering configurations
func TestEventSourceMappingFilterConfiguration(t *testing.T) {
	t.Run("FilterCriteriaValidation", func(t *testing.T) {
		// Test that filter criteria are properly validated
		validFilter := map[string]interface{}{
			"eventType": []string{"OrderPlaced", "OrderCancelled"},
		}
		
		invalidFilter := map[string]interface{}{
			"invalidKey": "invalidValue",
		}

		err := validateFilterCriteria(validFilter)
		assert.NoError(t, err)

		err = validateFilterCriteria(invalidFilter) 
		assert.Error(t, err)
	})
}

// optimizeBatchSizeForEventSource returns optimal batch size for different event sources
func optimizeBatchSizeForEventSource(eventSourceType string) int32 {
	switch eventSourceType {
	case "sqs":
		return 10
	case "kinesis":
		return 100
	case "dynamodb":
		return 50
	default:
		return 10 // Default batch size
	}
}

// validateParallelizationFactor validates parallelization factor
func validateParallelizationFactor(factor int32) error {
	if factor < 1 || factor > 100 {
		return fmt.Errorf("parallelization factor must be between 1 and 100, got %d", factor)
	}
	return nil
}

