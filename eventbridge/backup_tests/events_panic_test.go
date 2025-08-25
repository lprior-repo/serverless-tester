package eventbridge

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/stretchr/testify/mock"
)

// RED: Test PutEvent panic version with error calls FailNow
func TestPutEvent_WithError_CallsFailNow(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutEvents", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.PutEventsOutput)(nil), errors.New("test error"))

	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT, Client: mockClient}
	testEvent := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"key": "value"}`,
	}

	// This should not panic but should call FailNow
	PutEvent(ctx, testEvent)

	mockT.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// RED: Test PutEvent panic version with invalid detail calls FailNow
func TestPutEvent_WithInvalidDetail_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}
	testEvent := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"invalid": json}`, // Invalid JSON
	}

	// This should not panic but should call FailNow
	PutEvent(ctx, testEvent)

	mockT.AssertExpectations(t)
}

// RED: Test PutEvents panic version with error calls FailNow
func TestPutEvents_WithError_CallsFailNow(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutEvents", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.PutEventsOutput)(nil), errors.New("test error"))

	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT, Client: mockClient}
	testEvents := []CustomEvent{
		{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
		},
	}

	// This should not panic but should call FailNow
	PutEvents(ctx, testEvents)

	mockT.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// RED: Test PutEvents panic version with oversized batch calls FailNow
func TestPutEvents_WithOversizedBatch_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}
	
	// Create more events than MaxEventBatchSize
	testEvents := make([]CustomEvent, MaxEventBatchSize+1)
	for i := range testEvents {
		testEvents[i] = CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"index": 1}`,
		}
	}

	// This should not panic but should call FailNow
	PutEvents(ctx, testEvents)

	mockT.AssertExpectations(t)
}