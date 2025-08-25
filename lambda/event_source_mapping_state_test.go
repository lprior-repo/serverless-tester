package lambda

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWaitForEventSourceMappingState_SimpleSuccess tests basic wait functionality with immediate success
func TestWaitForEventSourceMappingState_SimpleSuccess(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	uuid := "immediate-success-uuid"
	expectedState := "Enabled"
	
	// Set up mock to return the expected state immediately
	mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
		UUID:  aws.String(uuid),
		State: aws.String(expectedState),
	}

	// Act - this should succeed immediately since the mock returns the target state
	err := WaitForEventSourceMappingStateE(ctx, uuid, expectedState, 1*time.Second)

	// Assert
	assert.NoError(t, err)
	
	// Verify GetEventSourceMapping was called at least once
	assert.GreaterOrEqual(t, len(mockClient.GetEventSourceMappingCalls), 1)
	assert.Equal(t, uuid, mockClient.GetEventSourceMappingCalls[0])
}

// TestWaitForEventSourceMappingState_ImmediateTimeout tests timeout with a very short duration
func TestWaitForEventSourceMappingState_ImmediateTimeout(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	uuid := "timeout-test-uuid"
	expectedState := "Enabled"
	wrongState := "Creating"
	
	// Set up mock to return the wrong state (should cause timeout)
	mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
		UUID:  aws.String(uuid),
		State: aws.String(wrongState),
	}

	// Act - use a very short timeout to trigger timeout quickly
	err := WaitForEventSourceMappingStateE(ctx, uuid, expectedState, 50*time.Millisecond)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for event source mapping")
	assert.Contains(t, err.Error(), uuid)
	assert.Contains(t, err.Error(), expectedState)
}

// TestWaitForEventSourceMappingState_MultipleStates tests sequential state changes using the enhanced mock
func TestWaitForEventSourceMappingState_MultipleStates(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	uuid := "multi-state-uuid"
	targetState := "Enabled"
	
	// Set up sequential responses: Creating -> Creating -> Enabled
	mockClient.GetEventSourceMappingResponses = []*lambda.GetEventSourceMappingOutput{
		{
			UUID:  aws.String(uuid),
			State: aws.String("Creating"),
		},
		{
			UUID:  aws.String(uuid),
			State: aws.String("Creating"),
		},
		{
			UUID:  aws.String(uuid),
			State: aws.String("Enabled"),
		},
	}

	// Act
	err := WaitForEventSourceMappingStateE(ctx, uuid, targetState, 1*time.Second)

	// Assert
	assert.NoError(t, err)
	
	// Verify GetEventSourceMapping was called multiple times
	assert.GreaterOrEqual(t, len(mockClient.GetEventSourceMappingCalls), 2)
	// Should have made at least 3 calls (Creating, Creating, Enabled)
	assert.GreaterOrEqual(t, len(mockClient.GetEventSourceMappingCalls), 3)
}

// TestGetEventSourceMapping_DirectCall tests the GetEventSourceMapping function directly
func TestGetEventSourceMapping_DirectCall(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	uuid := "direct-call-uuid"
	
	mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
		UUID:          aws.String(uuid),
		EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
		FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
		State:         aws.String("Enabled"),
		BatchSize:     aws.Int32(10),
	}

	// Act
	result := GetEventSourceMapping(ctx, uuid)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, uuid, result.UUID)
	assert.Equal(t, "Enabled", result.State)
	assert.Equal(t, int32(10), result.BatchSize)
	
	// Verify the mock was called
	require.Len(t, mockClient.GetEventSourceMappingCalls, 1)
	assert.Equal(t, uuid, mockClient.GetEventSourceMappingCalls[0])
}

// TestUpdateEventSourceMappingBatchConfiguration_Complete tests the complete update workflow  
func TestUpdateEventSourceMappingBatchConfiguration_Complete(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}

	uuid := "update-batch-config-uuid"
	
	// Setup current mapping response (for the GET call)
	mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
		UUID:          aws.String(uuid),
		EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:update-queue"),
		FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:update-function"),
		State:         aws.String("Enabled"),
		BatchSize:     aws.Int32(10),
		MaximumBatchingWindowInSeconds: aws.Int32(1),
	}

	// Setup update response
	newBatchSize := int32(100)
	newBatchingWindow := int32(5)
	mockClient.UpdateEventSourceMappingResponse = &lambda.UpdateEventSourceMappingOutput{
		UUID:          aws.String(uuid),
		EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:update-queue"),
		BatchSize:     aws.Int32(newBatchSize),
		MaximumBatchingWindowInSeconds: aws.Int32(newBatchingWindow),
		State:         aws.String("Enabled"),
	}

	// New batch configuration
	batchConfig := BatchConfigurationOptions{
		BatchSize:                        newBatchSize,
		MaximumBatchingWindowInSeconds:   newBatchingWindow,
	}

	// Act
	result := UpdateEventSourceMappingBatchConfiguration(ctx, uuid, batchConfig)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, uuid, result.UUID)
	assert.Equal(t, newBatchSize, result.BatchSize)
	assert.Equal(t, newBatchingWindow, result.MaximumBatchingWindowInSeconds)

	// Verify both GET and UPDATE calls were made
	require.Len(t, mockClient.GetEventSourceMappingCalls, 1)
	assert.Equal(t, uuid, mockClient.GetEventSourceMappingCalls[0])
	
	require.Len(t, mockClient.UpdateEventSourceMappingCalls, 1)
	updateCall := mockClient.UpdateEventSourceMappingCalls[0]
	assert.Equal(t, uuid, updateCall.UUID)
	assert.Equal(t, newBatchSize, *updateCall.BatchSize)
}

// TestEventSourceMappingLifecycle tests the complete lifecycle of an event source mapping
func TestEventSourceMappingLifecycle(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}

	functionName := "lifecycle-test-function"
	eventSourceArn := "arn:aws:sqs:us-east-1:123456789012:lifecycle-queue"
	uuid := "lifecycle-uuid"

	t.Run("Create", func(t *testing.T) {
		mockClient.Reset()
		
		mockClient.CreateEventSourceMappingResponse = &lambda.CreateEventSourceMappingOutput{
			UUID:          aws.String(uuid),
			EventSourceArn: aws.String(eventSourceArn),
			FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:" + functionName),
			State:         aws.String("Creating"),
			BatchSize:     aws.Int32(10),
		}

		config := EventSourceMappingConfig{
			EventSourceArn:   eventSourceArn,
			FunctionName:     functionName,
			BatchSize:        10,
			Enabled:          true,
		}

		result := CreateEventSourceMapping(ctx, config)
		
		require.NotNil(t, result)
		assert.Equal(t, uuid, result.UUID)
		assert.Equal(t, "Creating", result.State)
	})

	t.Run("Get", func(t *testing.T) {
		mockClient.Reset()
		
		mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
			UUID:          aws.String(uuid),
			EventSourceArn: aws.String(eventSourceArn),
			State:         aws.String("Enabled"),
			BatchSize:     aws.Int32(10),
		}

		result := GetEventSourceMapping(ctx, uuid)
		
		require.NotNil(t, result)
		assert.Equal(t, uuid, result.UUID)
		assert.Equal(t, "Enabled", result.State)
	})

	t.Run("Update", func(t *testing.T) {
		mockClient.Reset()
		
		newBatchSize := int32(25)
		mockClient.UpdateEventSourceMappingResponse = &lambda.UpdateEventSourceMappingOutput{
			UUID:      aws.String(uuid),
			BatchSize: aws.Int32(newBatchSize),
			State:     aws.String("Updating"),
		}

		updateConfig := UpdateEventSourceMappingConfig{
			BatchSize: &newBatchSize,
		}

		result := UpdateEventSourceMapping(ctx, uuid, updateConfig)
		
		require.NotNil(t, result)
		assert.Equal(t, uuid, result.UUID)
		assert.Equal(t, newBatchSize, result.BatchSize)
		assert.Equal(t, "Updating", result.State)
	})

	t.Run("Delete", func(t *testing.T) {
		mockClient.Reset()
		
		mockClient.DeleteEventSourceMappingResponse = &lambda.DeleteEventSourceMappingOutput{
			UUID:  aws.String(uuid),
			State: aws.String("Deleting"),
		}

		result := DeleteEventSourceMapping(ctx, uuid)
		
		require.NotNil(t, result)
		assert.Equal(t, uuid, result.UUID)
		assert.Equal(t, "Deleting", result.State)
	})
}

// TestListEventSourceMappings_Comprehensive tests listing with pagination and multiple mappings
func TestListEventSourceMappings_Comprehensive(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	functionName := "comprehensive-list-function"
	
	// Setup multiple mappings response
	mappings := []types.EventSourceMappingConfiguration{
		{
			UUID:          aws.String("uuid-1"),
			EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:queue-1"),
			FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:comprehensive-list-function"),
			State:         aws.String("Enabled"),
			BatchSize:     aws.Int32(10),
		},
		{
			UUID:          aws.String("uuid-2"), 
			EventSourceArn: aws.String("arn:aws:kinesis:us-east-1:123456789012:stream/test-stream"),
			FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:comprehensive-list-function"),
			State:         aws.String("Disabled"),
			BatchSize:     aws.Int32(100),
		},
		{
			UUID:          aws.String("uuid-3"),
			EventSourceArn: aws.String("arn:aws:dynamodb:us-east-1:123456789012:table/TestTable/stream/2023-01-01T00:00:00.000"),
			FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:comprehensive-list-function"),
			State:         aws.String("Creating"),
			BatchSize:     aws.Int32(50),
		},
	}
	
	mockClient.ListEventSourceMappingsResponse = &lambda.ListEventSourceMappingsOutput{
		EventSourceMappings: mappings,
	}

	// Act
	results := ListEventSourceMappings(ctx, functionName)

	// Assert
	require.Len(t, results, 3)
	
	// Verify each mapping
	assert.Equal(t, "uuid-1", results[0].UUID)
	assert.Equal(t, "Enabled", results[0].State)
	assert.Equal(t, int32(10), results[0].BatchSize)
	
	assert.Equal(t, "uuid-2", results[1].UUID) 
	assert.Equal(t, "Disabled", results[1].State)
	assert.Equal(t, int32(100), results[1].BatchSize)
	
	assert.Equal(t, "uuid-3", results[2].UUID)
	assert.Equal(t, "Creating", results[2].State)
	assert.Equal(t, int32(50), results[2].BatchSize)

	// Verify mock was called
	require.Len(t, mockClient.ListEventSourceMappingsCalls, 1)
	assert.Equal(t, functionName, mockClient.ListEventSourceMappingsCalls[0])
}