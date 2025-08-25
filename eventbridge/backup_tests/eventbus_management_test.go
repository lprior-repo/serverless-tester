package eventbridge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// RED: Test CreateEventBusE with valid configuration
func TestCreateEventBusE_WithValidConfig_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := EventBusConfig{
		Name:            "test-custom-bus",
		EventSourceName: "aws.partner/example.com",
		Tags: map[string]string{
			"Environment": "test",
			"Purpose":     "testing",
		},
		DeadLetterConfig: &types.DeadLetterConfig{
			Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:dlq"),
		},
	}

	busArn := "arn:aws:events:us-east-1:123456789012:event-bus/test-custom-bus"
	mockClient.On("CreateEventBus", context.Background(), mock.MatchedBy(func(input *eventbridge.CreateEventBusInput) bool {
		return *input.Name == "test-custom-bus" &&
			*input.EventSourceName == "aws.partner/example.com" &&
			len(input.Tags) == 2 &&
			input.DeadLetterConfig != nil &&
			*input.DeadLetterConfig.Arn == "arn:aws:sqs:us-east-1:123456789012:dlq"
	}), mock.Anything).Return(&eventbridge.CreateEventBusOutput{
		EventBusArn: &busArn,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateEventBusE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "test-custom-bus", result.Name)
	assert.Equal(t, busArn, result.Arn)
	assert.Equal(t, "aws.partner/example.com", result.EventSourceName)
	mockClient.AssertExpectations(t)
}

// RED: Test CreateEventBusE with empty bus name
func TestCreateEventBusE_WithEmptyBusName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := EventBusConfig{
		Name: "", // Empty name
	}

	result, err := CreateEventBusE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event bus name cannot be empty")
	assert.Empty(t, result.Name)
}

// RED: Test CreateEventBusE with default bus name
func TestCreateEventBusE_WithDefaultBusName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := EventBusConfig{
		Name: DefaultEventBusName, // Cannot create default bus
	}

	result, err := CreateEventBusE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot create default event bus")
	assert.Empty(t, result.Name)
}

// RED: Test CreateEventBusE with minimal configuration
func TestCreateEventBusE_WithMinimalConfig_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := EventBusConfig{
		Name: "simple-bus",
	}

	busArn := "arn:aws:events:us-east-1:123456789012:event-bus/simple-bus"
	mockClient.On("CreateEventBus", context.Background(), mock.MatchedBy(func(input *eventbridge.CreateEventBusInput) bool {
		return *input.Name == "simple-bus" &&
			input.EventSourceName == nil &&
			len(input.Tags) == 0 &&
			input.DeadLetterConfig == nil
	}), mock.Anything).Return(&eventbridge.CreateEventBusOutput{
		EventBusArn: &busArn,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateEventBusE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "simple-bus", result.Name)
	assert.Equal(t, busArn, result.Arn)
	assert.Empty(t, result.EventSourceName)
	mockClient.AssertExpectations(t)
}

// RED: Test CreateEventBusE with retry on transient error
func TestCreateEventBusE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := EventBusConfig{
		Name: "retry-bus",
	}

	busArn := "arn:aws:events:us-east-1:123456789012:event-bus/retry-bus"
	// First call fails, second succeeds
	mockClient.On("CreateEventBus", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.CreateEventBusOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("CreateEventBus", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.CreateEventBusOutput{
		EventBusArn: &busArn,
	}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateEventBusE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "retry-bus", result.Name)
	assert.Equal(t, busArn, result.Arn)
	mockClient.AssertExpectations(t)
}

// RED: Test DeleteEventBusE with valid bus name
func TestDeleteEventBusE_WithValidBusName_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	mockClient.On("DeleteEventBus", context.Background(), mock.MatchedBy(func(input *eventbridge.DeleteEventBusInput) bool {
		return *input.Name == "test-bus"
	}), mock.Anything).Return(&eventbridge.DeleteEventBusOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := DeleteEventBusE(ctx, "test-bus")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test DeleteEventBusE with empty bus name
func TestDeleteEventBusE_WithEmptyBusName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	err := DeleteEventBusE(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event bus name cannot be empty")
}

// RED: Test DeleteEventBusE with default bus name
func TestDeleteEventBusE_WithDefaultBusName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	err := DeleteEventBusE(ctx, DefaultEventBusName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete default event bus")
}

// RED: Test DeleteEventBusE with retry on transient error
func TestDeleteEventBusE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	// First call fails, second succeeds
	mockClient.On("DeleteEventBus", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.DeleteEventBusOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("DeleteEventBus", context.Background(), mock.Anything, mock.Anything).Return(
		&eventbridge.DeleteEventBusOutput{}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := DeleteEventBusE(ctx, "test-bus")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test DescribeEventBusE with valid bus name
func TestDescribeEventBusE_WithValidBusName_ReturnsDetails(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	busName := "test-bus"
	busArn := "arn:aws:events:us-east-1:123456789012:event-bus/test-bus"
	creationTime := time.Now()

	mockClient.On("DescribeEventBus", context.Background(), mock.MatchedBy(func(input *eventbridge.DescribeEventBusInput) bool {
		return *input.Name == "test-bus"
	}), mock.Anything).Return(&eventbridge.DescribeEventBusOutput{
		Name:         &busName,
		Arn:          &busArn,
		CreationTime: &creationTime,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := DescribeEventBusE(ctx, "test-bus")

	assert.NoError(t, err)
	assert.Equal(t, "test-bus", result.Name)
	assert.Equal(t, busArn, result.Arn)
	assert.Equal(t, &creationTime, result.CreationTime)
	mockClient.AssertExpectations(t)
}

// RED: Test DescribeEventBusE with default bus name
func TestDescribeEventBusE_WithDefaultBusName_ReturnsDefaultBus(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	busName := DefaultEventBusName
	busArn := "arn:aws:events:us-east-1:123456789012:event-bus/default"

	mockClient.On("DescribeEventBus", context.Background(), mock.MatchedBy(func(input *eventbridge.DescribeEventBusInput) bool {
		return *input.Name == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.DescribeEventBusOutput{
		Name: &busName,
		Arn:  &busArn,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := DescribeEventBusE(ctx, "")

	assert.NoError(t, err)
	assert.Equal(t, DefaultEventBusName, result.Name)
	assert.Equal(t, busArn, result.Arn)
	mockClient.AssertExpectations(t)
}

// RED: Test DescribeEventBusE with retry on transient error
func TestDescribeEventBusE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	busName := "test-bus"
	busArn := "arn:aws:events:us-east-1:123456789012:event-bus/test-bus"

	// First call fails, second succeeds
	mockClient.On("DescribeEventBus", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.DescribeEventBusOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("DescribeEventBus", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.DescribeEventBusOutput{
		Name: &busName,
		Arn:  &busArn,
	}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := DescribeEventBusE(ctx, "test-bus")

	assert.NoError(t, err)
	assert.Equal(t, "test-bus", result.Name)
	assert.Equal(t, busArn, result.Arn)
	mockClient.AssertExpectations(t)
}

// RED: Test ListEventBusesE with valid response
func TestListEventBusesE_WithValidResponse_ReturnsAllBuses(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	busName1 := DefaultEventBusName
	busArn1 := "arn:aws:events:us-east-1:123456789012:event-bus/default"
	creationTime1 := time.Now()

	busName2 := "custom-bus"
	busArn2 := "arn:aws:events:us-east-1:123456789012:event-bus/custom-bus"
	creationTime2 := time.Now().Add(-time.Hour)

	mockClient.On("ListEventBuses", context.Background(), mock.MatchedBy(func(input *eventbridge.ListEventBusesInput) bool {
		return true // No specific input validation needed for list operation
	}), mock.Anything).Return(&eventbridge.ListEventBusesOutput{
		EventBuses: []types.EventBus{
			{
				Name:         &busName1,
				Arn:          &busArn1,
				CreationTime: &creationTime1,
			},
			{
				Name:         &busName2,
				Arn:          &busArn2,
				CreationTime: &creationTime2,
			},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	results, err := ListEventBusesE(ctx)

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Equal(t, DefaultEventBusName, results[0].Name)
	assert.Equal(t, busArn1, results[0].Arn)
	assert.Equal(t, &creationTime1, results[0].CreationTime)

	assert.Equal(t, "custom-bus", results[1].Name)
	assert.Equal(t, busArn2, results[1].Arn)
	assert.Equal(t, &creationTime2, results[1].CreationTime)

	mockClient.AssertExpectations(t)
}

// RED: Test ListEventBusesE with empty response
func TestListEventBusesE_WithEmptyResponse_ReturnsEmptySlice(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	mockClient.On("ListEventBuses", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListEventBusesOutput{
		EventBuses: []types.EventBus{},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	results, err := ListEventBusesE(ctx)

	assert.NoError(t, err)
	assert.Len(t, results, 0)
	mockClient.AssertExpectations(t)
}

// RED: Test PutPermissionE with valid parameters
func TestPutPermissionE_WithValidParameters_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	mockClient.On("PutPermission", context.Background(), mock.MatchedBy(func(input *eventbridge.PutPermissionInput) bool {
		return *input.StatementId == "test-statement" &&
			*input.Principal == "123456789012" &&
			*input.Action == "events:PutEvents" &&
			*input.EventBusName == "custom-bus"
	}), mock.Anything).Return(&eventbridge.PutPermissionOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := PutPermissionE(ctx, "custom-bus", "test-statement", "123456789012", "events:PutEvents")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test PutPermissionE with empty statement ID
func TestPutPermissionE_WithEmptyStatementId_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	err := PutPermissionE(ctx, "custom-bus", "", "123456789012", "events:PutEvents")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "statement ID cannot be empty")
}

// RED: Test PutPermissionE with empty principal
func TestPutPermissionE_WithEmptyPrincipal_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	err := PutPermissionE(ctx, "custom-bus", "test-statement", "", "events:PutEvents")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "principal cannot be empty")
}

// RED: Test PutPermissionE with default action
func TestPutPermissionE_WithDefaultAction_UsesDefaultAction(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	mockClient.On("PutPermission", context.Background(), mock.MatchedBy(func(input *eventbridge.PutPermissionInput) bool {
		return *input.StatementId == "test-statement" &&
			*input.Principal == "123456789012" &&
			*input.Action == "events:PutEvents" &&
			*input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.PutPermissionOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := PutPermissionE(ctx, "", "test-statement", "123456789012", "") // Empty action should default

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test PutPermissionE with retry on transient error
func TestPutPermissionE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	// First call fails, second succeeds
	mockClient.On("PutPermission", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.PutPermissionOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("PutPermission", context.Background(), mock.Anything, mock.Anything).Return(
		&eventbridge.PutPermissionOutput{}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := PutPermissionE(ctx, "custom-bus", "test-statement", "123456789012", "events:PutEvents")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test RemovePermissionE with valid parameters
func TestRemovePermissionE_WithValidParameters_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	mockClient.On("RemovePermission", context.Background(), mock.MatchedBy(func(input *eventbridge.RemovePermissionInput) bool {
		return *input.StatementId == "test-statement" &&
			*input.EventBusName == "custom-bus"
	}), mock.Anything).Return(&eventbridge.RemovePermissionOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := RemovePermissionE(ctx, "custom-bus", "test-statement")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test RemovePermissionE with empty statement ID
func TestRemovePermissionE_WithEmptyStatementId_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	err := RemovePermissionE(ctx, "custom-bus", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "statement ID cannot be empty")
}

// RED: Test RemovePermissionE with default event bus
func TestRemovePermissionE_WithDefaultEventBus_UsesDefault(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	mockClient.On("RemovePermission", context.Background(), mock.MatchedBy(func(input *eventbridge.RemovePermissionInput) bool {
		return *input.StatementId == "test-statement" &&
			*input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.RemovePermissionOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := RemovePermissionE(ctx, "", "test-statement")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test RemovePermissionE with retry on transient error
func TestRemovePermissionE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	// First call fails, second succeeds
	mockClient.On("RemovePermission", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.RemovePermissionOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("RemovePermission", context.Background(), mock.Anything, mock.Anything).Return(
		&eventbridge.RemovePermissionOutput{}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := RemovePermissionE(ctx, "custom-bus", "test-statement")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test panic versions of functions call FailNow on error
func TestCreateEventBus_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}
	config := EventBusConfig{
		Name: "", // Empty name to cause error
	}

	// This should not panic but should call FailNow
	CreateEventBus(ctx, config)

	mockT.AssertExpectations(t)
}

// RED: Test DeleteEventBus panic version
func TestDeleteEventBus_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	DeleteEventBus(ctx, "") // Empty name to cause error

	mockT.AssertExpectations(t)
}

// RED: Test DescribeEventBus panic version
func TestDescribeEventBus_WithError_CallsFailNow(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("DescribeEventBus", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.DescribeEventBusOutput)(nil), errors.New("test error"))

	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT, Client: mockClient}

	// This should not panic but should call FailNow
	DescribeEventBus(ctx, "test-bus")

	mockT.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// RED: Test ListEventBuses panic version
func TestListEventBuses_WithError_CallsFailNow(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListEventBuses", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.ListEventBusesOutput)(nil), errors.New("test error"))

	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT, Client: mockClient}

	// This should not panic but should call FailNow
	ListEventBuses(ctx)

	mockT.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// RED: Test PutPermission panic version
func TestPutPermission_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	PutPermission(ctx, "custom-bus", "", "123456789012", "events:PutEvents") // Empty statement ID to cause error

	mockT.AssertExpectations(t)
}

// RED: Test RemovePermission panic version
func TestRemovePermission_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	RemovePermission(ctx, "custom-bus", "") // Empty statement ID to cause error

	mockT.AssertExpectations(t)
}