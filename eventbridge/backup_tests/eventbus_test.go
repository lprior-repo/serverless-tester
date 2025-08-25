package eventbridge

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateEventBusE_Success(t *testing.T) {
	// Test successful event bus creation
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	config := EventBusConfig{
		Name:            "test-custom-bus",
		EventSourceName: "test.source",
		Tags: map[string]string{
			"Environment": "test",
			"Project":     "eventbridge-tests",
		},
		DeadLetterConfig: &types.DeadLetterConfig{
			Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:DeadLetterQueue"),
		},
	}

	expectedBusArn := "arn:aws:events:us-east-1:123456789012:event-bus/test-custom-bus"

	// Mock successful response
	mockClient.On("CreateEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.CreateEventBusInput) bool {
		return *input.Name == "test-custom-bus" &&
			*input.EventSourceName == "test.source" &&
			len(input.Tags) == 2 &&
			input.DeadLetterConfig != nil &&
			*input.DeadLetterConfig.Arn == "arn:aws:sqs:us-east-1:123456789012:DeadLetterQueue"
	}), mock.Anything).Return(&eventbridge.CreateEventBusOutput{
		EventBusArn: aws.String(expectedBusArn),
	}, nil)

	result, err := CreateEventBusE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "test-custom-bus", result.Name)
	assert.Equal(t, expectedBusArn, result.Arn)
	assert.Equal(t, "test.source", result.EventSourceName)
	mockClient.AssertExpectations(t)
}

func TestCreateEventBusE_EmptyName(t *testing.T) {
	// Test validation for empty event bus name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	config := EventBusConfig{
		Name: "", // Empty name should cause validation error
	}

	result, err := CreateEventBusE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "event bus name cannot be empty")
	assert.Equal(t, EventBusResult{}, result)
}

func TestCreateEventBusE_DefaultBusName(t *testing.T) {
	// Test validation for default event bus name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	config := EventBusConfig{
		Name: DefaultEventBusName, // Cannot create default bus
	}

	result, err := CreateEventBusE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot create default event bus")
	assert.Equal(t, EventBusResult{}, result)
}

func TestCreateEventBusE_MinimalConfig(t *testing.T) {
	// Test with minimal configuration
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	config := EventBusConfig{
		Name: "minimal-bus",
		// No event source name, tags, or dead letter config
	}

	// Mock successful response
	mockClient.On("CreateEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.CreateEventBusInput) bool {
		return *input.Name == "minimal-bus" &&
			input.EventSourceName == nil &&
			len(input.Tags) == 0 &&
			input.DeadLetterConfig == nil
	}), mock.Anything).Return(&eventbridge.CreateEventBusOutput{
		EventBusArn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/minimal-bus"),
	}, nil)

	result, err := CreateEventBusE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "minimal-bus", result.Name)
	mockClient.AssertExpectations(t)
}

func TestCreateEventBusE_ServiceError(t *testing.T) {
	// Test AWS service error
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	config := EventBusConfig{
		Name: "test-bus",
	}

	serviceErr := errors.New("service error")
	mockClient.On("CreateEventBus", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	result, err := CreateEventBusE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create event bus after 3 attempts")
	assert.Equal(t, EventBusResult{}, result)
	mockClient.AssertExpectations(t)
}

func TestDeleteEventBusE_Success(t *testing.T) {
	// Test successful event bus deletion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock successful response
	mockClient.On("DeleteEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.DeleteEventBusInput) bool {
		return *input.Name == "test-custom-bus"
	}), mock.Anything).Return(&eventbridge.DeleteEventBusOutput{}, nil)

	err := DeleteEventBusE(ctx, "test-custom-bus")

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDeleteEventBusE_EmptyName(t *testing.T) {
	// Test validation for empty event bus name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := DeleteEventBusE(ctx, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "event bus name cannot be empty")
}

func TestDeleteEventBusE_DefaultBusName(t *testing.T) {
	// Test validation for default event bus name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := DeleteEventBusE(ctx, DefaultEventBusName)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete default event bus")
}

func TestDeleteEventBusE_ServiceError(t *testing.T) {
	// Test AWS service error
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	serviceErr := errors.New("service error")
	mockClient.On("DeleteEventBus", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	err := DeleteEventBusE(ctx, "test-bus")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete event bus after 3 attempts")
	mockClient.AssertExpectations(t)
}

func TestDescribeEventBusE_Success(t *testing.T) {
	// Test successful event bus description
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	creationTime := time.Now()
	expectedResponse := &eventbridge.DescribeEventBusOutput{
		Name:         aws.String("test-custom-bus"),
		Arn:          aws.String("arn:aws:events:us-east-1:123456789012:event-bus/test-custom-bus"),
		CreationTime: &creationTime,
	}

	mockClient.On("DescribeEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeEventBusInput) bool {
		return *input.Name == "test-custom-bus"
	}), mock.Anything).Return(expectedResponse, nil)

	result, err := DescribeEventBusE(ctx, "test-custom-bus")

	require.NoError(t, err)
	assert.Equal(t, "test-custom-bus", result.Name)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:event-bus/test-custom-bus", result.Arn)
	assert.Equal(t, &creationTime, result.CreationTime)
	mockClient.AssertExpectations(t)
}

func TestDescribeEventBusE_DefaultBus(t *testing.T) {
	// Test describing default event bus
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	expectedResponse := &eventbridge.DescribeEventBusOutput{
		Name: aws.String("default"),
		Arn:  aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
	}

	mockClient.On("DescribeEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeEventBusInput) bool {
		return *input.Name == DefaultEventBusName
	}), mock.Anything).Return(expectedResponse, nil)

	result, err := DescribeEventBusE(ctx, "") // Empty should default

	require.NoError(t, err)
	assert.Equal(t, "default", result.Name)
	mockClient.AssertExpectations(t)
}

func TestDescribeEventBusE_ServiceError(t *testing.T) {
	// Test AWS service error
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	serviceErr := errors.New("service error")
	mockClient.On("DescribeEventBus", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	result, err := DescribeEventBusE(ctx, "test-bus")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to describe event bus after 3 attempts")
	assert.Equal(t, EventBusResult{}, result)
	mockClient.AssertExpectations(t)
}

func TestListEventBusesE_Success(t *testing.T) {
	// Test successful event bus listing
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	creationTime := time.Now()
	expectedResponse := &eventbridge.ListEventBusesOutput{
		EventBuses: []types.EventBus{
			{
				Name:         aws.String("default"),
				Arn:          aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
				CreationTime: &creationTime,
			},
			{
				Name:         aws.String("custom-bus"),
				Arn:          aws.String("arn:aws:events:us-east-1:123456789012:event-bus/custom-bus"),
				CreationTime: &creationTime,
			},
		},
	}

	mockClient.On("ListEventBuses", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	results, err := ListEventBusesE(ctx)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "default", results[0].Name)
	assert.Equal(t, "custom-bus", results[1].Name)
	assert.Equal(t, &creationTime, results[0].CreationTime)
	mockClient.AssertExpectations(t)
}

func TestListEventBusesE_Empty(t *testing.T) {
	// Test empty event bus list
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	mockClient.On("ListEventBuses", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListEventBusesOutput{
		EventBuses: []types.EventBus{},
	}, nil)

	results, err := ListEventBusesE(ctx)

	require.NoError(t, err)
	assert.Empty(t, results)
	mockClient.AssertExpectations(t)
}

func TestListEventBusesE_ServiceError(t *testing.T) {
	// Test AWS service error
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	serviceErr := errors.New("service error")
	mockClient.On("ListEventBuses", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	results, err := ListEventBusesE(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list event buses after 3 attempts")
	assert.Nil(t, results)
	mockClient.AssertExpectations(t)
}

func TestPutPermissionE_Success(t *testing.T) {
	// Test successful permission addition
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock successful response
	mockClient.On("PutPermission", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutPermissionInput) bool {
		return *input.StatementId == "allow-cross-account" &&
			*input.Principal == "123456789012" &&
			*input.Action == "events:PutEvents" &&
			*input.EventBusName == "default"
	}), mock.Anything).Return(&eventbridge.PutPermissionOutput{}, nil)

	err := PutPermissionE(ctx, "default", "allow-cross-account", "123456789012", "events:PutEvents")

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutPermissionE_DefaultAction(t *testing.T) {
	// Test with default action
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock successful response
	mockClient.On("PutPermission", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutPermissionInput) bool {
		return *input.Action == "events:PutEvents" // Should default to this
	}), mock.Anything).Return(&eventbridge.PutPermissionOutput{}, nil)

	err := PutPermissionE(ctx, "default", "allow-cross-account", "123456789012", "") // Empty action should default

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutPermissionE_DefaultEventBus(t *testing.T) {
	// Test with default event bus
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock successful response
	mockClient.On("PutPermission", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutPermissionInput) bool {
		return *input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.PutPermissionOutput{}, nil)

	err := PutPermissionE(ctx, "", "allow-cross-account", "123456789012", "events:PutEvents") // Empty event bus should default

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutPermissionE_EmptyStatementID(t *testing.T) {
	// Test validation for empty statement ID
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := PutPermissionE(ctx, "default", "", "123456789012", "events:PutEvents")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "statement ID cannot be empty")
}

func TestPutPermissionE_EmptyPrincipal(t *testing.T) {
	// Test validation for empty principal
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := PutPermissionE(ctx, "default", "allow-cross-account", "", "events:PutEvents")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "principal cannot be empty")
}

func TestPutPermissionE_ServiceError(t *testing.T) {
	// Test AWS service error
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	serviceErr := errors.New("service error")
	mockClient.On("PutPermission", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	err := PutPermissionE(ctx, "default", "allow-cross-account", "123456789012", "events:PutEvents")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to put permission after 3 attempts")
	mockClient.AssertExpectations(t)
}

func TestRemovePermissionE_Success(t *testing.T) {
	// Test successful permission removal
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock successful response
	mockClient.On("RemovePermission", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemovePermissionInput) bool {
		return *input.StatementId == "allow-cross-account" &&
			*input.EventBusName == "default"
	}), mock.Anything).Return(&eventbridge.RemovePermissionOutput{}, nil)

	err := RemovePermissionE(ctx, "default", "allow-cross-account")

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRemovePermissionE_EmptyStatementID(t *testing.T) {
	// Test validation for empty statement ID
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := RemovePermissionE(ctx, "default", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "statement ID cannot be empty")
}

func TestRemovePermissionE_DefaultEventBus(t *testing.T) {
	// Test with default event bus
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock successful response
	mockClient.On("RemovePermission", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemovePermissionInput) bool {
		return *input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.RemovePermissionOutput{}, nil)

	err := RemovePermissionE(ctx, "", "allow-cross-account") // Empty event bus should default

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRemovePermissionE_ServiceError(t *testing.T) {
	// Test AWS service error
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	serviceErr := errors.New("service error")
	mockClient.On("RemovePermission", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	err := RemovePermissionE(ctx, "default", "allow-cross-account")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove permission after 3 attempts")
	mockClient.AssertExpectations(t)
}

func TestCreateEventBus_PanicOnError(t *testing.T) {
	// Test that CreateEventBus panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	config := EventBusConfig{
		Name: "", // Empty name should cause validation error
	}

	result := CreateEventBus(ctx, config)

	assert.Equal(t, EventBusResult{}, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestDeleteEventBus_PanicOnError(t *testing.T) {
	// Test that DeleteEventBus panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	DeleteEventBus(ctx, "") // Empty name should cause error

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestDescribeEventBus_PanicOnError(t *testing.T) {
	// Test that DescribeEventBus panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: mockClient,
	}

	serviceErr := errors.New("service error")
	mockClient.On("DescribeEventBus", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	result := DescribeEventBus(ctx, "test-bus")

	assert.Equal(t, EventBusResult{}, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestListEventBuses_PanicOnError(t *testing.T) {
	// Test that ListEventBuses panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: mockClient,
	}

	serviceErr := errors.New("service error")
	mockClient.On("ListEventBuses", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	result := ListEventBuses(ctx)

	assert.Nil(t, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestPutPermission_PanicOnError(t *testing.T) {
	// Test that PutPermission panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	PutPermission(ctx, "default", "", "123456789012", "events:PutEvents") // Empty statement ID should cause error

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestRemovePermission_PanicOnError(t *testing.T) {
	// Test that RemovePermission panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	RemovePermission(ctx, "default", "") // Empty statement ID should cause error

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}