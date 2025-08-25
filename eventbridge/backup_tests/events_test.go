package eventbridge

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPutEventE_Success(t *testing.T) {
	// RED: Test for successful event publishing
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	event := CustomEvent{
		Source:       "test.service",
		DetailType:   "Test Event",
		Detail:       `{"key": "value"}`,
		EventBusName: "default",
	}

	// Mock successful response
	mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		return len(input.Entries) == 1 &&
			*input.Entries[0].Source == "test.service" &&
			*input.Entries[0].DetailType == "Test Event" &&
			*input.Entries[0].EventBusName == "default"
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{
				EventId: aws.String("event-id-123"),
			},
		},
	}, nil)

	// Execute
	result, err := PutEventE(ctx, event)

	// Assert
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "event-id-123", result.EventID)
	assert.Empty(t, result.ErrorCode)
	assert.Empty(t, result.ErrorMessage)
	mockClient.AssertExpectations(t)
}

func TestPutEventE_ValidationError(t *testing.T) {
	// Test event detail validation
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	// Test invalid JSON detail
	event := CustomEvent{
		Source:       "test.service",
		DetailType:   "Test Event",
		Detail:       `{"invalid": json}`,
		EventBusName: "default",
	}

	result, err := PutEventE(ctx, event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event detail")
	assert.False(t, result.Success)
}

func TestPutEventE_EventTooLarge(t *testing.T) {
	// Test event size validation
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	// Create a large detail string that exceeds MaxEventDetailSize
	largeDetail := make([]byte, MaxEventDetailSize+1)
	for i := range largeDetail {
		largeDetail[i] = 'a'
	}

	event := CustomEvent{
		Source:       "test.service",
		DetailType:   "Test Event",
		Detail:       string(largeDetail),
		EventBusName: "default",
	}

	result, err := PutEventE(ctx, event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "event size too large")
	assert.False(t, result.Success)
}

func TestPutEventE_DefaultValues(t *testing.T) {
	// Test default values for event bus and time
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	event := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"key": "value"}`,
		// EventBusName not set - should default to "default"
		// Time not set - should default to current time
	}

	// Mock successful response
	mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		return len(input.Entries) == 1 &&
			*input.Entries[0].EventBusName == DefaultEventBusName &&
			input.Entries[0].Time != nil
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{
				EventId: aws.String("event-id-123"),
			},
		},
	}, nil)

	result, err := PutEventE(ctx, event)

	require.NoError(t, err)
	assert.True(t, result.Success)
	mockClient.AssertExpectations(t)
}

func TestPutEventE_WithResources(t *testing.T) {
	// Test event with resources
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	event := CustomEvent{
		Source:       "test.service",
		DetailType:   "Test Event",
		Detail:       `{"key": "value"}`,
		EventBusName: "default",
		Resources:    []string{"arn:aws:s3:::my-bucket", "arn:aws:lambda:us-east-1:123456789012:function:my-function"},
	}

	// Mock successful response
	mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		return len(input.Entries) == 1 &&
			len(input.Entries[0].Resources) == 2 &&
			input.Entries[0].Resources[0] == "arn:aws:s3:::my-bucket"
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{
				EventId: aws.String("event-id-123"),
			},
		},
	}, nil)

	result, err := PutEventE(ctx, event)

	require.NoError(t, err)
	assert.True(t, result.Success)
	mockClient.AssertExpectations(t)
}

func TestPutEventE_ServiceError(t *testing.T) {
	// Test AWS service error
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	event := CustomEvent{
		Source:       "test.service",
		DetailType:   "Test Event",
		Detail:       `{"key": "value"}`,
		EventBusName: "default",
	}

	serviceErr := errors.New("service error")
	mockClient.On("PutEvents", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	result, err := PutEventE(ctx, event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to put event after 3 attempts")
	assert.False(t, result.Success)
	mockClient.AssertExpectations(t)
}

func TestPutEventE_EventFailed(t *testing.T) {
	// Test event failure in response
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	event := CustomEvent{
		Source:       "test.service",
		DetailType:   "Test Event",
		Detail:       `{"key": "value"}`,
		EventBusName: "default",
	}

	// Mock failed response
	mockClient.On("PutEvents", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 1,
		Entries: []types.PutEventsResultEntry{
			{
				ErrorCode:    aws.String("ValidationException"),
				ErrorMessage: aws.String("Event pattern is not valid"),
			},
		},
	}, nil)

	result, err := PutEventE(ctx, event)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "ValidationException", result.ErrorCode)
	assert.Equal(t, "Event pattern is not valid", result.ErrorMessage)
	mockClient.AssertExpectations(t)
}

func TestPutEventsE_Success(t *testing.T) {
	// Test successful batch event publishing
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	events := []CustomEvent{
		{
			Source:       "test.service",
			DetailType:   "Test Event 1",
			Detail:       `{"key": "value1"}`,
			EventBusName: "default",
		},
		{
			Source:       "test.service",
			DetailType:   "Test Event 2",
			Detail:       `{"key": "value2"}`,
			EventBusName: "default",
		},
	}

	// Mock successful response
	mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		return len(input.Entries) == 2
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{
				EventId: aws.String("event-id-1"),
			},
			{
				EventId: aws.String("event-id-2"),
			},
		},
	}, nil)

	result, err := PutEventsE(ctx, events)

	require.NoError(t, err)
	assert.Equal(t, int32(0), result.FailedEntryCount)
	assert.Len(t, result.Entries, 2)
	assert.True(t, result.Entries[0].Success)
	assert.Equal(t, "event-id-1", result.Entries[0].EventID)
	assert.True(t, result.Entries[1].Success)
	assert.Equal(t, "event-id-2", result.Entries[1].EventID)
	mockClient.AssertExpectations(t)
}

func TestPutEventsE_EmptyEvents(t *testing.T) {
	// Test empty events slice
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	result, err := PutEventsE(ctx, []CustomEvent{})

	require.NoError(t, err)
	assert.Equal(t, int32(0), result.FailedEntryCount)
	assert.Empty(t, result.Entries)
}

func TestPutEventsE_TooManyEvents(t *testing.T) {
	// Test batch size limit
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	// Create more events than allowed in batch
	events := make([]CustomEvent, MaxEventBatchSize+1)
	for i := range events {
		events[i] = CustomEvent{
			Source:     "test.service",
			DetailType: fmt.Sprintf("Test Event %d", i),
			Detail:     `{"key": "value"}`,
		}
	}

	result, err := PutEventsE(ctx, events)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "batch size exceeds maximum")
	assert.Equal(t, PutEventsResult{}, result)
}

func TestPutEventsE_PartialFailure(t *testing.T) {
	// Test partial failure scenario
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	events := []CustomEvent{
		{
			Source:       "test.service",
			DetailType:   "Test Event 1",
			Detail:       `{"key": "value1"}`,
			EventBusName: "default",
		},
		{
			Source:       "test.service",
			DetailType:   "Test Event 2",
			Detail:       `{"key": "value2"}`,
			EventBusName: "default",
		},
	}

	// Mock mixed response - one success, one failure
	mockClient.On("PutEvents", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 1,
		Entries: []types.PutEventsResultEntry{
			{
				EventId: aws.String("event-id-1"),
			},
			{
				ErrorCode:    aws.String("ValidationException"),
				ErrorMessage: aws.String("Invalid event"),
			},
		},
	}, nil)

	result, err := PutEventsE(ctx, events)

	require.NoError(t, err)
	assert.Equal(t, int32(1), result.FailedEntryCount)
	assert.Len(t, result.Entries, 2)
	assert.True(t, result.Entries[0].Success)
	assert.Equal(t, "event-id-1", result.Entries[0].EventID)
	assert.False(t, result.Entries[1].Success)
	assert.Equal(t, "ValidationException", result.Entries[1].ErrorCode)
	mockClient.AssertExpectations(t)
}

func TestPutEvent_PanicOnError(t *testing.T) {
	// Test that PutEvent panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: mockClient,
	}

	event := CustomEvent{
		Source:       "test.service",
		DetailType:   "Test Event",
		Detail:       `{"invalid": json}`,
		EventBusName: "default",
	}

	// This should call FailNow due to validation error
	result := PutEvent(ctx, event)

	// Should return empty result since FailNow was called
	assert.Equal(t, PutEventResult{}, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestPutEvents_PanicOnError(t *testing.T) {
	// Test that PutEvents panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	// Create events that will exceed batch size
	events := make([]CustomEvent, MaxEventBatchSize+1)
	for i := range events {
		events[i] = CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
		}
	}

	result := PutEvents(ctx, events)

	assert.Equal(t, PutEventsResult{}, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}


func TestValidateEventDetail_Success(t *testing.T) {
	// Test valid JSON detail
	err := validateEventDetail(`{"key": "value", "count": 42}`)
	assert.NoError(t, err)
}

func TestValidateEventDetail_EmptyDetail(t *testing.T) {
	// Test empty detail (should be valid)
	err := validateEventDetail("")
	assert.NoError(t, err)
}

func TestValidateEventDetail_InvalidJSON(t *testing.T) {
	// Test invalid JSON
	err := validateEventDetail(`{"invalid": json}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event detail")
}

func TestValidateEventDetail_TooLarge(t *testing.T) {
	// Test detail size validation
	largeDetail := make([]byte, MaxEventDetailSize+1)
	for i := range largeDetail {
		largeDetail[i] = 'a'
	}

	err := validateEventDetail(string(largeDetail))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event size too large")
}

func TestSafeString_WithValue(t *testing.T) {
	// Test safeString with non-nil pointer
	value := "test-value"
	result := safeString(&value)
	assert.Equal(t, "test-value", result)
}

func TestSafeString_WithNil(t *testing.T) {
	// Test safeString with nil pointer
	result := safeString(nil)
	assert.Equal(t, "", result)
}

// mockTestingT implements TestingT interface for testing panic behavior
type mockTestingT struct {
	errorCalled   bool
	failNowCalled bool
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errorCalled = true
}

func (m *mockTestingT) Error(args ...interface{}) {
	m.errorCalled = true
}

func (m *mockTestingT) Fail() {
	// Set error called to true to track failure
	m.errorCalled = true
}

func (m *mockTestingT) FailNow() {
	m.failNowCalled = true
}

func (m *mockTestingT) Helper() {
	// No-op for mock
}

func (m *mockTestingT) Fatal(args ...interface{}) {
	m.failNowCalled = true
	panic("test fatal")
}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.failNowCalled = true
	panic("test fatal")
}

func (m *mockTestingT) Name() string {
	return "mockTestingT"
}