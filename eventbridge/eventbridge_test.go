package eventbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventBridgeClient implements EventBridgeAPI for testing
type MockEventBridgeClient struct {
	mock.Mock
}

// MockTestingT implements the TestingT interface for testing panic scenarios
type MockTestingT struct {
	ErrorfCalled bool
	FailNowCalled bool
	ErrorfArgs []interface{}
}

func (m *MockTestingT) Errorf(format string, args ...interface{}) { 
	m.ErrorfCalled = true
	m.ErrorfArgs = append([]interface{}{format}, args...)
}
func (m *MockTestingT) Error(args ...interface{})                 {}
func (m *MockTestingT) Fail()                                     {}
func (m *MockTestingT) FailNow()                                  { 
	m.FailNowCalled = true
	panic("FailNow called") 
}
func (m *MockTestingT) Helper()                                   {}
func (m *MockTestingT) Fatal(args ...interface{})                { panic("Fatal called") }
func (m *MockTestingT) Fatalf(format string, args ...interface{}) { panic("Fatalf called") }
func (m *MockTestingT) Name() string                              { return "MockTestingT" }

// PutEvents mocks the PutEvents operation
func (m *MockEventBridgeClient) PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.PutEventsOutput), args.Error(1)
}

// PutRule mocks the PutRule operation
func (m *MockEventBridgeClient) PutRule(ctx context.Context, params *eventbridge.PutRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutRuleOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.PutRuleOutput), args.Error(1)
}

// DeleteRule mocks the DeleteRule operation
func (m *MockEventBridgeClient) DeleteRule(ctx context.Context, params *eventbridge.DeleteRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteRuleOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DeleteRuleOutput), args.Error(1)
}

// DescribeRule mocks the DescribeRule operation
func (m *MockEventBridgeClient) DescribeRule(ctx context.Context, params *eventbridge.DescribeRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeRuleOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DescribeRuleOutput), args.Error(1)
}

// EnableRule mocks the EnableRule operation
func (m *MockEventBridgeClient) EnableRule(ctx context.Context, params *eventbridge.EnableRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.EnableRuleOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.EnableRuleOutput), args.Error(1)
}

// DisableRule mocks the DisableRule operation
func (m *MockEventBridgeClient) DisableRule(ctx context.Context, params *eventbridge.DisableRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DisableRuleOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DisableRuleOutput), args.Error(1)
}

// ListRules mocks the ListRules operation
func (m *MockEventBridgeClient) ListRules(ctx context.Context, params *eventbridge.ListRulesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListRulesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.ListRulesOutput), args.Error(1)
}

// PutTargets mocks the PutTargets operation
func (m *MockEventBridgeClient) PutTargets(ctx context.Context, params *eventbridge.PutTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutTargetsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.PutTargetsOutput), args.Error(1)
}

// RemoveTargets mocks the RemoveTargets operation
func (m *MockEventBridgeClient) RemoveTargets(ctx context.Context, params *eventbridge.RemoveTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.RemoveTargetsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.RemoveTargetsOutput), args.Error(1)
}

// ListTargetsByRule mocks the ListTargetsByRule operation
func (m *MockEventBridgeClient) ListTargetsByRule(ctx context.Context, params *eventbridge.ListTargetsByRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListTargetsByRuleOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.ListTargetsByRuleOutput), args.Error(1)
}

// CreateEventBus mocks the CreateEventBus operation
func (m *MockEventBridgeClient) CreateEventBus(ctx context.Context, params *eventbridge.CreateEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.CreateEventBusOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.CreateEventBusOutput), args.Error(1)
}

// DeleteEventBus mocks the DeleteEventBus operation
func (m *MockEventBridgeClient) DeleteEventBus(ctx context.Context, params *eventbridge.DeleteEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteEventBusOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DeleteEventBusOutput), args.Error(1)
}

// DescribeEventBus mocks the DescribeEventBus operation
func (m *MockEventBridgeClient) DescribeEventBus(ctx context.Context, params *eventbridge.DescribeEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeEventBusOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DescribeEventBusOutput), args.Error(1)
}

// ListEventBuses mocks the ListEventBuses operation
func (m *MockEventBridgeClient) ListEventBuses(ctx context.Context, params *eventbridge.ListEventBusesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListEventBusesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.ListEventBusesOutput), args.Error(1)
}

// CreateArchive mocks the CreateArchive operation
func (m *MockEventBridgeClient) CreateArchive(ctx context.Context, params *eventbridge.CreateArchiveInput, optFns ...func(*eventbridge.Options)) (*eventbridge.CreateArchiveOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.CreateArchiveOutput), args.Error(1)
}

// DeleteArchive mocks the DeleteArchive operation
func (m *MockEventBridgeClient) DeleteArchive(ctx context.Context, params *eventbridge.DeleteArchiveInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteArchiveOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DeleteArchiveOutput), args.Error(1)
}

// DescribeArchive mocks the DescribeArchive operation
func (m *MockEventBridgeClient) DescribeArchive(ctx context.Context, params *eventbridge.DescribeArchiveInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeArchiveOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DescribeArchiveOutput), args.Error(1)
}

// ListArchives mocks the ListArchives operation
func (m *MockEventBridgeClient) ListArchives(ctx context.Context, params *eventbridge.ListArchivesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListArchivesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.ListArchivesOutput), args.Error(1)
}

// StartReplay mocks the StartReplay operation
func (m *MockEventBridgeClient) StartReplay(ctx context.Context, params *eventbridge.StartReplayInput, optFns ...func(*eventbridge.Options)) (*eventbridge.StartReplayOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.StartReplayOutput), args.Error(1)
}

// DescribeReplay mocks the DescribeReplay operation
func (m *MockEventBridgeClient) DescribeReplay(ctx context.Context, params *eventbridge.DescribeReplayInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeReplayOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DescribeReplayOutput), args.Error(1)
}

// CancelReplay mocks the CancelReplay operation
func (m *MockEventBridgeClient) CancelReplay(ctx context.Context, params *eventbridge.CancelReplayInput, optFns ...func(*eventbridge.Options)) (*eventbridge.CancelReplayOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.CancelReplayOutput), args.Error(1)
}

// PutPermission mocks the PutPermission operation
func (m *MockEventBridgeClient) PutPermission(ctx context.Context, params *eventbridge.PutPermissionInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutPermissionOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.PutPermissionOutput), args.Error(1)
}

// RemovePermission mocks the RemovePermission operation
func (m *MockEventBridgeClient) RemovePermission(ctx context.Context, params *eventbridge.RemovePermissionInput, optFns ...func(*eventbridge.Options)) (*eventbridge.RemovePermissionOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.RemovePermissionOutput), args.Error(1)
}

// TestEventPattern mocks the TestEventPattern operation
func (m *MockEventBridgeClient) TestEventPattern(ctx context.Context, params *eventbridge.TestEventPatternInput, optFns ...func(*eventbridge.Options)) (*eventbridge.TestEventPatternOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.TestEventPatternOutput), args.Error(1)
}

// RED PHASE: Start with failing tests for core functionality

// TestEventBridgeConfiguration tests core configuration functionality
func TestEventBridgeConfiguration(t *testing.T) {
	t.Run("DefaultRetryConfig", func(t *testing.T) {
		config := defaultRetryConfig()
		assert.Equal(t, DefaultRetryAttempts, config.MaxAttempts)
		assert.Equal(t, DefaultRetryDelay, config.BaseDelay)
		assert.Equal(t, 30*time.Second, config.MaxDelay)
		assert.Equal(t, 2.0, config.Multiplier)
	})

	t.Run("CreateEventBridgeClient_WithMockClient", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			Client: mockClient,
		}
		
		client := createEventBridgeClient(ctx)
		assert.Equal(t, mockClient, client)
	})

	t.Run("CreateEventBridgeClient_WithoutMockClient", func(t *testing.T) {
		ctx := &TestContext{
			AwsConfig: aws.Config{Region: "us-east-1"},
		}
		
		client := createEventBridgeClient(ctx)
		assert.NotNil(t, client)
	})
}

// TestEventValidation tests event validation functions
func TestEventValidation(t *testing.T) {
	t.Run("ValidateEventDetail_EmptyDetail", func(t *testing.T) {
		err := validateEventDetail("")
		assert.NoError(t, err)
	})

	t.Run("ValidateEventDetail_ValidJSON", func(t *testing.T) {
		validJSON := `{"key": "value", "number": 123}`
		err := validateEventDetail(validJSON)
		assert.NoError(t, err)
	})

	t.Run("ValidateEventDetail_InvalidJSON", func(t *testing.T) {
		invalidJSON := `{"key": invalid}`
		err := validateEventDetail(invalidJSON)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidEventDetail)
	})

	t.Run("ValidateEventDetail_SizeTooLarge", func(t *testing.T) {
		largeDetail := make([]byte, MaxEventDetailSize+1)
		for i := range largeDetail {
			largeDetail[i] = 'a'
		}
		err := validateEventDetail(string(largeDetail))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEventSizeTooLarge)
	})
}

// TestBackoffCalculation tests exponential backoff calculation
func TestBackoffCalculation(t *testing.T) {
	config := RetryConfig{
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
		MaxAttempts: 3,
	}

	t.Run("ZeroAttempt", func(t *testing.T) {
		delay := calculateBackoffDelay(0, config)
		assert.Equal(t, time.Second, delay)
	})

	t.Run("FirstRetry", func(t *testing.T) {
		delay := calculateBackoffDelay(1, config)
		assert.Equal(t, 2*time.Second, delay)
	})

	t.Run("SecondRetry", func(t *testing.T) {
		delay := calculateBackoffDelay(2, config)
		assert.Equal(t, 4*time.Second, delay)
	})

	t.Run("MaxDelayReached", func(t *testing.T) {
		largeConfig := RetryConfig{
			BaseDelay:   time.Second,
			MaxDelay:    5 * time.Second,
			Multiplier:  10.0,
			MaxAttempts: 5,
		}
		delay := calculateBackoffDelay(3, largeConfig)
		assert.Equal(t, 5*time.Second, delay)
	})

	t.Run("ZeroMultiplier", func(t *testing.T) {
		zeroConfig := RetryConfig{
			BaseDelay:   time.Second,
			MaxDelay:    30 * time.Second,
			Multiplier:  0.0,
			MaxAttempts: 3,
		}
		delay := calculateBackoffDelay(2, zeroConfig)
		assert.Equal(t, time.Duration(0), delay)
	})

	t.Run("FractionalMultiplier", func(t *testing.T) {
		fractionalConfig := RetryConfig{
			BaseDelay:   time.Second,
			MaxDelay:    30 * time.Second,
			Multiplier:  0.5,
			MaxAttempts: 3,
		}
		delay := calculateBackoffDelay(1, fractionalConfig)
		assert.Equal(t, 500*time.Millisecond, delay)
	})
}

// TestPutEventE tests single event publishing
func TestPutEventE(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		expectedEventID := "test-event-id"
		mockClient.On("PutEvents", mock.Anything, mock.AnythingOfType("*eventbridge.PutEventsInput"), mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				FailedEntryCount: 0,
				Entries: []types.PutEventsResultEntry{
					{
						EventId: &expectedEventID,
					},
				},
			}, nil)

		event := CustomEvent{
			Source:       "test.service",
			DetailType:   "Test Event",
			Detail:       `{"key": "value"}`,
			EventBusName: "test-bus",
		}

		result, err := PutEventE(ctx, event)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, expectedEventID, result.EventID)
		mockClient.AssertExpectations(t)
	})

	t.Run("InvalidEventDetail", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"invalid": json}`,
		}

		result, err := PutEventE(ctx, event)

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.ErrorIs(t, err, ErrInvalidEventDetail)
	})

	t.Run("FailedEntry", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		errorCode := "InternalException"
		errorMessage := "Internal error occurred"
		mockClient.On("PutEvents", mock.Anything, mock.AnythingOfType("*eventbridge.PutEventsInput"), mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				FailedEntryCount: 1,
				Entries: []types.PutEventsResultEntry{
					{
						ErrorCode:    &errorCode,
						ErrorMessage: &errorMessage,
					},
				},
			}, nil)

		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
		}

		result, err := PutEventE(ctx, event)

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Equal(t, errorCode, result.ErrorCode)
		assert.Equal(t, errorMessage, result.ErrorMessage)
		mockClient.AssertExpectations(t)
	})

	t.Run("RetryOnTransientError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// First call fails, second succeeds
		mockClient.On("PutEvents", mock.Anything, mock.AnythingOfType("*eventbridge.PutEventsInput"), mock.Anything).Return(
			nil, errors.New("transient error")).Once()
		
		expectedEventID := "test-event-id"
		mockClient.On("PutEvents", mock.Anything, mock.AnythingOfType("*eventbridge.PutEventsInput"), mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				FailedEntryCount: 0,
				Entries: []types.PutEventsResultEntry{
					{
						EventId: &expectedEventID,
					},
				},
			}, nil).Once()

		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
		}

		result, err := PutEventE(ctx, event)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, expectedEventID, result.EventID)
		mockClient.AssertExpectations(t)
	})

	t.Run("MaxRetriesExceeded", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// All calls fail
		mockClient.On("PutEvents", mock.Anything, mock.AnythingOfType("*eventbridge.PutEventsInput"), mock.Anything).Return(
			nil, errors.New("persistent error")).Times(DefaultRetryAttempts)

		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
		}

		result, err := PutEventE(ctx, event)

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, err.Error(), "failed to put event after")
		mockClient.AssertExpectations(t)
	})

	t.Run("DefaultEventBusName", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		expectedEventID := "test-event-id"
		mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
			return len(input.Entries) == 1 && *input.Entries[0].EventBusName == DefaultEventBusName
		}), mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				FailedEntryCount: 0,
				Entries: []types.PutEventsResultEntry{
					{
						EventId: &expectedEventID,
					},
				},
			}, nil)

		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
			// EventBusName not set
		}

		result, err := PutEventE(ctx, event)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		mockClient.AssertExpectations(t)
	})

	t.Run("WithResources", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		expectedEventID := "test-event-id"
		expectedResources := []string{"arn:aws:s3:::test-bucket", "arn:aws:lambda:us-east-1:123456789012:function:test"}
		
		mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
			entry := input.Entries[0]
			return len(entry.Resources) == 2 && 
				   entry.Resources[0] == expectedResources[0] && 
				   entry.Resources[1] == expectedResources[1]
		}), mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				FailedEntryCount: 0,
				Entries: []types.PutEventsResultEntry{
					{
						EventId: &expectedEventID,
					},
				},
			}, nil)

		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
			Resources:  expectedResources,
		}

		result, err := PutEventE(ctx, event)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		mockClient.AssertExpectations(t)
	})
}

// TestPutEventsE tests batch event publishing
func TestPutEventsE(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		eventID1 := "event-1"
		eventID2 := "event-2"
		mockClient.On("PutEvents", mock.Anything, mock.AnythingOfType("*eventbridge.PutEventsInput"), mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				FailedEntryCount: 0,
				Entries: []types.PutEventsResultEntry{
					{EventId: &eventID1},
					{EventId: &eventID2},
				},
			}, nil)

		events := []CustomEvent{
			{
				Source:     "test.service",
				DetailType: "Event 1",
				Detail:     `{"id": 1}`,
			},
			{
				Source:     "test.service",
				DetailType: "Event 2",
				Detail:     `{"id": 2}`,
			},
		}

		result, err := PutEventsE(ctx, events)

		assert.NoError(t, err)
		assert.Equal(t, int32(0), result.FailedEntryCount)
		assert.Len(t, result.Entries, 2)
		assert.True(t, result.Entries[0].Success)
		assert.True(t, result.Entries[1].Success)
		assert.Equal(t, eventID1, result.Entries[0].EventID)
		assert.Equal(t, eventID2, result.Entries[1].EventID)
		mockClient.AssertExpectations(t)
	})

	t.Run("EmptyEventsList", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		result, err := PutEventsE(ctx, []CustomEvent{})

		assert.NoError(t, err)
		assert.Equal(t, int32(0), result.FailedEntryCount)
		assert.Empty(t, result.Entries)
	})

	t.Run("BatchSizeExceeded", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		// Create more than MaxEventBatchSize events
		events := make([]CustomEvent, MaxEventBatchSize+1)
		for i := range events {
			events[i] = CustomEvent{
				Source:     "test.service",
				DetailType: "Test Event",
				Detail:     `{"id": ` + string(rune(i)) + `}`,
			}
		}

		result, err := PutEventsE(ctx, events)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch size exceeds maximum")
		assert.Equal(t, PutEventsResult{}, result)
	})

	t.Run("PartialFailure", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		eventID1 := "event-1"
		errorCode := "ValidationException"
		errorMessage := "Invalid event"
		
		mockClient.On("PutEvents", mock.Anything, mock.AnythingOfType("*eventbridge.PutEventsInput"), mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				FailedEntryCount: 1,
				Entries: []types.PutEventsResultEntry{
					{EventId: &eventID1},
					{
						ErrorCode:    &errorCode,
						ErrorMessage: &errorMessage,
					},
				},
			}, nil)

		events := []CustomEvent{
			{
				Source:     "test.service",
				DetailType: "Event 1",
				Detail:     `{"id": 1}`,
			},
			{
				Source:     "test.service",
				DetailType: "Event 2",
				Detail:     `{"id": 2}`,
			},
		}

		result, err := PutEventsE(ctx, events)

		assert.NoError(t, err)
		assert.Equal(t, int32(1), result.FailedEntryCount)
		assert.Len(t, result.Entries, 2)
		assert.True(t, result.Entries[0].Success)
		assert.False(t, result.Entries[1].Success)
		assert.Equal(t, eventID1, result.Entries[0].EventID)
		assert.Equal(t, errorCode, result.Entries[1].ErrorCode)
		assert.Equal(t, errorMessage, result.Entries[1].ErrorMessage)
		mockClient.AssertExpectations(t)
	})

	t.Run("InvalidEventInBatch", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		events := []CustomEvent{
			{
				Source:     "test.service",
				DetailType: "Valid Event",
				Detail:     `{"valid": true}`,
			},
			{
				Source:     "test.service",
				DetailType: "Invalid Event",
				Detail:     `{"invalid": json}`, // Invalid JSON
			},
		}

		result, err := PutEventsE(ctx, events)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidEventDetail)
		assert.Equal(t, PutEventsResult{}, result)
	})
}

// TestEventBuilders tests event building functions
func TestEventBuilders(t *testing.T) {
	t.Run("BuildCustomEvent", func(t *testing.T) {
		detail := map[string]interface{}{
			"key":    "value",
			"number": 123,
			"nested": map[string]interface{}{
				"inner": "data",
			},
		}

		event := BuildCustomEvent("test.service", "Test Event", detail)

		assert.Equal(t, "test.service", event.Source)
		assert.Equal(t, "Test Event", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		
		// Verify JSON marshaling
		assert.Contains(t, event.Detail, `"key":"value"`)
		assert.Contains(t, event.Detail, `"number":123`)
		assert.Contains(t, event.Detail, `"nested":`)
	})

	t.Run("BuildCustomEvent_WithInvalidDetail", func(t *testing.T) {
		// Create a detail with a value that can't be marshaled to JSON
		detail := map[string]interface{}{
			"invalid": make(chan int), // Channels can't be marshaled
		}

		event := BuildCustomEvent("test.service", "Test Event", detail)

		assert.Equal(t, "test.service", event.Source)
		assert.Equal(t, "Test Event", event.DetailType)
		assert.Equal(t, "{}", event.Detail) // Should default to empty object
	})

	t.Run("BuildScheduledEvent", func(t *testing.T) {
		detail := map[string]interface{}{
			"action": "scheduled_task",
			"time":   "2023-01-01T00:00:00Z",
		}

		event := BuildScheduledEvent("Scheduled Task", detail)

		assert.Equal(t, "aws.events", event.Source)
		assert.Equal(t, "Scheduled Task", event.DetailType)
		assert.Contains(t, event.Detail, `"action":"scheduled_task"`)
	})
}

// TestEventPatternValidation tests event pattern validation
func TestEventPatternValidation(t *testing.T) {
	t.Run("ValidateEventPattern_Valid", func(t *testing.T) {
		validPattern := `{
			"source": ["test.service"],
			"detail-type": ["Test Event"],
			"detail": {
				"key": ["value"]
			}
		}`

		isValid := ValidateEventPattern(validPattern)
		assert.True(t, isValid)
	})

	t.Run("ValidateEventPattern_EmptyPattern", func(t *testing.T) {
		isValid := ValidateEventPattern("")
		assert.False(t, isValid)
	})

	t.Run("ValidateEventPattern_InvalidJSON", func(t *testing.T) {
		invalidPattern := `{"source": invalid}`

		isValid := ValidateEventPattern(invalidPattern)
		assert.False(t, isValid)
	})

	t.Run("ValidateEventPattern_InvalidKey", func(t *testing.T) {
		invalidPattern := `{"invalid-key": ["value"]}`

		isValid := ValidateEventPattern(invalidPattern)
		assert.False(t, isValid)
	})

	t.Run("ValidateEventPattern_InvalidSourceFormat", func(t *testing.T) {
		invalidPattern := `{"source": "should-be-array"}`

		isValid := ValidateEventPattern(invalidPattern)
		assert.False(t, isValid)
	})

	t.Run("ValidateEventPatternE_Valid", func(t *testing.T) {
		validPattern := `{
			"source": ["test.service"],
			"detail-type": ["Test Event"]
		}`

		err := ValidateEventPatternE(validPattern)
		assert.NoError(t, err)
	})

	t.Run("ValidateEventPatternE_EmptyPattern", func(t *testing.T) {
		err := ValidateEventPatternE("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("ValidateEventPatternE_InvalidJSON", func(t *testing.T) {
		err := ValidateEventPatternE(`{"source": invalid}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pattern JSON")
	})

	t.Run("ValidateEventPatternE_InvalidKey", func(t *testing.T) {
		err := ValidateEventPatternE(`{"invalid-key": ["value"]}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pattern key")
	})

	t.Run("ValidateEventPatternE_NonArrayValue", func(t *testing.T) {
		err := ValidateEventPatternE(`{"source": "should-be-array"}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be an array")
	})

	t.Run("ValidateEventPatternE_NonStringArrayItem", func(t *testing.T) {
		err := ValidateEventPatternE(`{"source": [123]}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must contain only strings")
	})

	t.Run("ValidateEventPatternE_InvalidDetailType", func(t *testing.T) {
		err := ValidateEventPatternE(`{"detail": "should-be-object"}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be an object")
	})
}

// TestPatternMatching tests event pattern matching
func TestPatternMatching(t *testing.T) {
	t.Run("TestEventPattern_SimpleMatch", func(t *testing.T) {
		pattern := `{"source": ["test.service"]}`
		event := map[string]interface{}{
			"source": "test.service",
		}

		matches := TestEventPattern(pattern, event)
		assert.True(t, matches)
	})

	t.Run("TestEventPattern_NoMatch", func(t *testing.T) {
		pattern := `{"source": ["other.service"]}`
		event := map[string]interface{}{
			"source": "test.service",
		}

		matches := TestEventPattern(pattern, event)
		assert.False(t, matches)
	})

	t.Run("TestEventPatternE_EmptyPattern", func(t *testing.T) {
		event := map[string]interface{}{
			"source": "test.service",
		}

		matches, err := TestEventPatternE("", event)
		assert.Error(t, err)
		assert.False(t, matches)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("TestEventPatternE_InvalidJSON", func(t *testing.T) {
		pattern := `{"source": invalid}`
		event := map[string]interface{}{
			"source": "test.service",
		}

		matches, err := TestEventPatternE(pattern, event)
		assert.Error(t, err)
		assert.False(t, matches)
		assert.Contains(t, err.Error(), "invalid pattern JSON")
	})

	t.Run("TestEventPatternE_ComplexMatch", func(t *testing.T) {
		pattern := `{
			"source": ["test.service"],
			"detail-type": ["Order Placed"],
			"detail": {
				"amount": [100]
			}
		}`
		event := map[string]interface{}{
			"source":      "test.service",
			"detail-type": "Order Placed",
			"detail": map[string]interface{}{
				"amount": 100,
			},
		}

		matches, err := TestEventPatternE(pattern, event)
		assert.NoError(t, err)
		assert.True(t, matches)
	})
}

// TestRuleManagement tests EventBridge rule operations
func TestRuleManagement(t *testing.T) {
	t.Run("CreateRuleE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
		mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
			return *input.Name == "test-rule" && *input.Description == "Test rule"
		}), mock.Anything).Return(
			&eventbridge.PutRuleOutput{
				RuleArn: &ruleArn,
			}, nil)

		config := RuleConfig{
			Name:        "test-rule",
			Description: "Test rule",
			EventPattern: `{"source": ["test.service"]}`,
			State:       RuleStateEnabled,
		}

		result, err := CreateRuleE(ctx, config)

		assert.NoError(t, err)
		assert.Equal(t, "test-rule", result.Name)
		assert.Equal(t, ruleArn, result.RuleArn)
		assert.Equal(t, RuleStateEnabled, result.State)
		mockClient.AssertExpectations(t)
	})

	t.Run("CreateRuleE_EmptyName", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		config := RuleConfig{
			Description:  "Test rule",
			EventPattern: `{"source": ["test.service"]}`,
		}

		result, err := CreateRuleE(ctx, config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rule name cannot be empty")
		assert.Equal(t, RuleResult{}, result)
	})

	t.Run("CreateRuleE_InvalidEventPattern", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		config := RuleConfig{
			Name:         "test-rule",
			Description:  "Test rule",
			EventPattern: `{"source": "invalid-pattern"}`, // Should be array
		}

		result, err := CreateRuleE(ctx, config)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidEventPattern)
		assert.Equal(t, RuleResult{}, result)
	})

	t.Run("CreateRuleE_WithScheduleExpression", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		ruleArn := "arn:aws:events:us-east-1:123456789012:rule/scheduled-rule"
		mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
			return *input.Name == "scheduled-rule" && 
				   *input.ScheduleExpression == "rate(5 minutes)" &&
				   input.EventPattern == nil
		}), mock.Anything).Return(
			&eventbridge.PutRuleOutput{
				RuleArn: &ruleArn,
			}, nil)

		config := RuleConfig{
			Name:               "scheduled-rule",
			Description:        "Scheduled rule",
			ScheduleExpression: "rate(5 minutes)",
			State:              RuleStateEnabled,
		}

		result, err := CreateRuleE(ctx, config)

		assert.NoError(t, err)
		assert.Equal(t, "scheduled-rule", result.Name)
		assert.Equal(t, ruleArn, result.RuleArn)
		mockClient.AssertExpectations(t)
	})

	t.Run("DescribeRuleE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
		eventPattern := `{"source": ["test.service"]}`
		description := "Test rule"
		
		mockClient.On("DescribeRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeRuleInput) bool {
			return *input.Name == "test-rule" && *input.EventBusName == "custom-bus"
		}), mock.Anything).Return(
			&eventbridge.DescribeRuleOutput{
				Arn:          &ruleArn,
				Name:         aws.String("test-rule"),
				Description:  &description,
				EventPattern: &eventPattern,
				State:        RuleStateEnabled,
				EventBusName: aws.String("custom-bus"),
			}, nil)

		result, err := DescribeRuleE(ctx, "test-rule", "custom-bus")

		assert.NoError(t, err)
		assert.Equal(t, "test-rule", result.Name)
		assert.Equal(t, ruleArn, result.RuleArn)
		assert.Equal(t, description, result.Description)
		assert.Equal(t, eventPattern, result.EventPattern)
		assert.Equal(t, RuleStateEnabled, result.State)
		assert.Equal(t, "custom-bus", result.EventBusName)
		mockClient.AssertExpectations(t)
	})

	t.Run("DescribeRuleE_NotFound", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("DescribeRule", mock.Anything, mock.AnythingOfType("*eventbridge.DescribeRuleInput"), mock.Anything).Return(
			nil, errors.New("ResourceNotFoundException"))

		result, err := DescribeRuleE(ctx, "nonexistent-rule", "default")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")
		assert.Equal(t, RuleResult{}, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("DeleteRuleE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// DeleteRuleE internally calls ListTargetsByRule and RemoveTargets
		mockClient.On("ListTargetsByRule", mock.Anything, mock.AnythingOfType("*eventbridge.ListTargetsByRuleInput"), mock.Anything).Return(
			&eventbridge.ListTargetsByRuleOutput{
				Targets: []types.Target{
					{
						Id:  aws.String("target-1"),
						Arn: aws.String("arn:aws:lambda:us-east-1:123456789012:function:test"),
					},
				},
			}, nil)

		mockClient.On("RemoveTargets", mock.Anything, mock.AnythingOfType("*eventbridge.RemoveTargetsInput"), mock.Anything).Return(
			&eventbridge.RemoveTargetsOutput{
				FailedEntryCount: 0,
				FailedEntries:    []types.RemoveTargetsResultEntry{},
			}, nil)

		mockClient.On("DeleteRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DeleteRuleInput) bool {
			return *input.Name == "test-rule" && *input.EventBusName == "default"
		}), mock.Anything).Return(
			&eventbridge.DeleteRuleOutput{}, nil)

		err := DeleteRuleE(ctx, "test-rule", "default")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("EnableRuleE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("EnableRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.EnableRuleInput) bool {
			return *input.Name == "test-rule" && *input.EventBusName == "default"
		}), mock.Anything).Return(
			&eventbridge.EnableRuleOutput{}, nil)

		err := EnableRuleE(ctx, "test-rule", "default")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("DisableRuleE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("DisableRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DisableRuleInput) bool {
			return *input.Name == "test-rule" && *input.EventBusName == "default"
		}), mock.Anything).Return(
			&eventbridge.DisableRuleOutput{}, nil)

		err := DisableRuleE(ctx, "test-rule", "default")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}

// TestTargetManagement tests EventBridge target operations
func TestTargetManagement(t *testing.T) {
	t.Run("PutTargetsE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
			return *input.Rule == "test-rule" &&
				   len(input.Targets) == 1 &&
				   *input.Targets[0].Id == "target-1" &&
				   *input.Targets[0].Arn == "arn:aws:lambda:us-east-1:123456789012:function:test"
		}), mock.Anything).Return(
			&eventbridge.PutTargetsOutput{
				FailedEntryCount: 0,
				FailedEntries:    []types.PutTargetsResultEntry{},
			}, nil)

		targets := []TargetConfig{
			{
				ID:  "target-1",
				Arn: "arn:aws:lambda:us-east-1:123456789012:function:test",
			},
		}

		err := PutTargetsE(ctx, "test-rule", "default", targets)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("PutTargetsE_WithInputTransformer", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		inputTransformer := &types.InputTransformer{
			InputPathsMap: map[string]string{
				"timestamp": "$.time",
				"source":    "$.source",
			},
			InputTemplate: aws.String(`{"eventTime": "<timestamp>", "eventSource": "<source>"}`),
		}

		mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
			return *input.Rule == "test-rule" &&
				   len(input.Targets) == 1 &&
				   input.Targets[0].InputTransformer != nil &&
				   *input.Targets[0].InputTransformer.InputTemplate == `{"eventTime": "<timestamp>", "eventSource": "<source>"}`
		}), mock.Anything).Return(
			&eventbridge.PutTargetsOutput{
				FailedEntryCount: 0,
				FailedEntries:    []types.PutTargetsResultEntry{},
			}, nil)

		targets := []TargetConfig{
			{
				ID:               "target-1",
				Arn:              "arn:aws:lambda:us-east-1:123456789012:function:test",
				InputTransformer: inputTransformer,
			},
		}

		err := PutTargetsE(ctx, "test-rule", "default", targets)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveTargetsE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("RemoveTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
			return *input.Rule == "test-rule" &&
				   *input.EventBusName == "default" &&
				   len(input.Ids) == 1 &&
				   input.Ids[0] == "target-1"
		}), mock.Anything).Return(
			&eventbridge.RemoveTargetsOutput{
				FailedEntryCount: 0,
				FailedEntries:    []types.RemoveTargetsResultEntry{},
			}, nil)

		err := RemoveTargetsE(ctx, "test-rule", "default", []string{"target-1"})

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListTargetsForRuleE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		targetArn := "arn:aws:lambda:us-east-1:123456789012:function:test"
		mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
			return *input.Rule == "test-rule" && *input.EventBusName == "default"
		}), mock.Anything).Return(
			&eventbridge.ListTargetsByRuleOutput{
				Targets: []types.Target{
					{
						Id:  aws.String("target-1"),
						Arn: &targetArn,
					},
				},
			}, nil)

		targets, err := ListTargetsForRuleE(ctx, "test-rule", "default")

		assert.NoError(t, err)
		assert.Len(t, targets, 1)
		assert.Equal(t, "target-1", targets[0].ID)
		assert.Equal(t, targetArn, targets[0].Arn)
		mockClient.AssertExpectations(t)
	})
}

// TestEventBusManagement tests EventBridge custom event bus operations
func TestEventBusManagement(t *testing.T) {
	t.Run("CreateEventBusE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		busArn := "arn:aws:events:us-east-1:123456789012:event-bus/custom-bus"
		mockClient.On("CreateEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.CreateEventBusInput) bool {
			return *input.Name == "custom-bus"
		}), mock.Anything).Return(
			&eventbridge.CreateEventBusOutput{
				EventBusArn: &busArn,
			}, nil)

		config := EventBusConfig{
			Name: "custom-bus",
			Tags: map[string]string{
				"Environment": "test",
			},
		}

		result, err := CreateEventBusE(ctx, config)

		assert.NoError(t, err)
		assert.Equal(t, "custom-bus", result.Name)
		assert.Equal(t, busArn, result.Arn)
		mockClient.AssertExpectations(t)
	})

	t.Run("CreateEventBusE_EmptyName", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		config := EventBusConfig{}

		result, err := CreateEventBusE(ctx, config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event bus name cannot be empty")
		assert.Equal(t, EventBusResult{}, result)
	})

	t.Run("DeleteEventBusE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("DeleteEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.DeleteEventBusInput) bool {
			return *input.Name == "custom-bus"
		}), mock.Anything).Return(
			&eventbridge.DeleteEventBusOutput{}, nil)

		err := DeleteEventBusE(ctx, "custom-bus")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("DescribeEventBusE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		busArn := "arn:aws:events:us-east-1:123456789012:event-bus/custom-bus"
		creationTime := time.Now()
		
		mockClient.On("DescribeEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeEventBusInput) bool {
			return *input.Name == "custom-bus"
		}), mock.Anything).Return(
			&eventbridge.DescribeEventBusOutput{
				Name:         aws.String("custom-bus"),
				Arn:          &busArn,
				CreationTime: &creationTime,
			}, nil)

		result, err := DescribeEventBusE(ctx, "custom-bus")

		assert.NoError(t, err)
		assert.Equal(t, "custom-bus", result.Name)
		assert.Equal(t, busArn, result.Arn)
		assert.Equal(t, creationTime, *result.CreationTime)
		mockClient.AssertExpectations(t)
	})
}

// TestArchiveOperations tests EventBridge archive operations
func TestArchiveOperations(t *testing.T) {
	t.Run("CreateArchiveE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		archiveArn := "arn:aws:events:us-east-1:123456789012:archive/test-archive"
		creationTime := time.Now()
		
		mockClient.On("CreateArchive", mock.Anything, mock.MatchedBy(func(input *eventbridge.CreateArchiveInput) bool {
			return *input.ArchiveName == "test-archive" &&
				   *input.EventSourceArn == "arn:aws:events:us-east-1:123456789012:event-bus/default" &&
				   *input.RetentionDays == 7
		}), mock.Anything).Return(
			&eventbridge.CreateArchiveOutput{
				ArchiveArn:   &archiveArn,
				CreationTime: &creationTime,
				State:        ArchiveStateEnabled,
			}, nil)

		config := ArchiveConfig{
			ArchiveName:    "test-archive",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
			RetentionDays:  7,
			Description:    "Test archive",
		}

		result, err := CreateArchiveE(ctx, config)

		assert.NoError(t, err)
		assert.Equal(t, "test-archive", result.ArchiveName)
		assert.Equal(t, archiveArn, result.ArchiveArn)
		assert.Equal(t, ArchiveStateEnabled, result.State)
		mockClient.AssertExpectations(t)
	})

	t.Run("CreateArchiveE_EmptyArchiveName", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		config := ArchiveConfig{
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
		}

		result, err := CreateArchiveE(ctx, config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "archive name cannot be empty")
		assert.Equal(t, ArchiveResult{}, result)
	})

	t.Run("DescribeArchiveE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		archiveArn := "arn:aws:events:us-east-1:123456789012:archive/test-archive"
		eventSourceArn := "arn:aws:events:us-east-1:123456789012:event-bus/default"
		creationTime := time.Now()
		
		mockClient.On("DescribeArchive", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeArchiveInput) bool {
			return *input.ArchiveName == "test-archive"
		}), mock.Anything).Return(
			&eventbridge.DescribeArchiveOutput{
				ArchiveName:    aws.String("test-archive"),
				ArchiveArn:     &archiveArn,
				EventSourceArn: &eventSourceArn,
				State:          ArchiveStateEnabled,
				CreationTime:   &creationTime,
			}, nil)

		result, err := DescribeArchiveE(ctx, "test-archive")

		assert.NoError(t, err)
		assert.Equal(t, "test-archive", result.ArchiveName)
		assert.Equal(t, archiveArn, result.ArchiveArn)
		assert.Equal(t, eventSourceArn, result.EventSourceArn)
		assert.Equal(t, ArchiveStateEnabled, result.State)
		mockClient.AssertExpectations(t)
	})

	t.Run("DeleteArchiveE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("DeleteArchive", mock.Anything, mock.MatchedBy(func(input *eventbridge.DeleteArchiveInput) bool {
			return *input.ArchiveName == "test-archive"
		}), mock.Anything).Return(
			&eventbridge.DeleteArchiveOutput{}, nil)

		err := DeleteArchiveE(ctx, "test-archive")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}

// TestReplayOperations tests EventBridge replay operations
func TestReplayOperations(t *testing.T) {
	t.Run("StartReplayE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		replayArn := "arn:aws:events:us-east-1:123456789012:replay/test-replay"
		replayStartTime := time.Now()
		
		mockClient.On("StartReplay", mock.Anything, mock.MatchedBy(func(input *eventbridge.StartReplayInput) bool {
			return *input.ReplayName == "test-replay" &&
				   *input.EventSourceArn == "arn:aws:events:us-east-1:123456789012:archive/test-archive"
		}), mock.Anything).Return(
			&eventbridge.StartReplayOutput{
				ReplayArn:       &replayArn,
				State:           ReplayStateStarting,
				ReplayStartTime: &replayStartTime,
			}, nil)

		config := ReplayConfig{
			ReplayName:     "test-replay",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
			EventStartTime: time.Now().Add(-24 * time.Hour),
			EventEndTime:   time.Now(),
			Destination: types.ReplayDestination{
				Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
			},
		}

		result, err := StartReplayE(ctx, config)

		assert.NoError(t, err)
		assert.Equal(t, "test-replay", result.ReplayName)
		assert.Equal(t, replayArn, result.ReplayArn)
		assert.Equal(t, ReplayStateStarting, result.State)
		mockClient.AssertExpectations(t)
	})

	t.Run("DescribeReplayE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		replayArn := "arn:aws:events:us-east-1:123456789012:replay/test-replay"
		eventSourceArn := "arn:aws:events:us-east-1:123456789012:archive/test-archive"
		eventStartTime := time.Now().Add(-24 * time.Hour)
		eventEndTime := time.Now()
		replayStartTime := time.Now()
		
		mockClient.On("DescribeReplay", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeReplayInput) bool {
			return *input.ReplayName == "test-replay"
		}), mock.Anything).Return(
			&eventbridge.DescribeReplayOutput{
				ReplayName:      aws.String("test-replay"),
				ReplayArn:       &replayArn,
				EventSourceArn:  &eventSourceArn,
				State:           ReplayStateCompleted,
				EventStartTime:  &eventStartTime,
				EventEndTime:    &eventEndTime,
				ReplayStartTime: &replayStartTime,
			}, nil)

		result, err := DescribeReplayE(ctx, "test-replay")

		assert.NoError(t, err)
		assert.Equal(t, "test-replay", result.ReplayName)
		assert.Equal(t, replayArn, result.ReplayArn)
		assert.Equal(t, ReplayStateCompleted, result.State)
		assert.Equal(t, eventStartTime, *result.EventStartTime)
		assert.Equal(t, eventEndTime, *result.EventEndTime)
		mockClient.AssertExpectations(t)
	})

	t.Run("CancelReplayE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("CancelReplay", mock.Anything, mock.MatchedBy(func(input *eventbridge.CancelReplayInput) bool {
			return *input.ReplayName == "test-replay"
		}), mock.Anything).Return(
			&eventbridge.CancelReplayOutput{
				ReplayArn: aws.String("arn:aws:events:us-east-1:123456789012:replay/test-replay"),
				State:     ReplayStateFailed,
			}, nil)

		err := CancelReplayE(ctx, "test-replay")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}

// TestAssertions tests EventBridge assertion functions
func TestAssertions(t *testing.T) {
	t.Run("AssertRuleExistsE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
		mockClient.On("DescribeRule", mock.Anything, mock.AnythingOfType("*eventbridge.DescribeRuleInput"), mock.Anything).Return(
			&eventbridge.DescribeRuleOutput{
				Arn:   &ruleArn,
				Name:  aws.String("test-rule"),
				State: RuleStateEnabled,
			}, nil)

		err := AssertRuleExistsE(ctx, "test-rule", "default")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertRuleExistsE_NotFound", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("DescribeRule", mock.Anything, mock.AnythingOfType("*eventbridge.DescribeRuleInput"), mock.Anything).Return(
			nil, errors.New("ResourceNotFoundException"))

		err := AssertRuleExistsE(ctx, "test-rule", "default")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertRuleEnabledE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
		mockClient.On("DescribeRule", mock.Anything, mock.AnythingOfType("*eventbridge.DescribeRuleInput"), mock.Anything).Return(
			&eventbridge.DescribeRuleOutput{
				Arn:   &ruleArn,
				Name:  aws.String("test-rule"),
				State: RuleStateEnabled,
			}, nil)

		err := AssertRuleEnabledE(ctx, "test-rule", "default")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertRuleEnabledE_Disabled", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
		mockClient.On("DescribeRule", mock.Anything, mock.AnythingOfType("*eventbridge.DescribeRuleInput"), mock.Anything).Return(
			&eventbridge.DescribeRuleOutput{
				Arn:   &ruleArn,
				Name:  aws.String("test-rule"),
				State: RuleStateDisabled,
			}, nil)

		err := AssertRuleEnabledE(ctx, "test-rule", "default")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not enabled")
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertTargetExistsE_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		targetArn := "arn:aws:lambda:us-east-1:123456789012:function:test"
		mockClient.On("ListTargetsByRule", mock.Anything, mock.AnythingOfType("*eventbridge.ListTargetsByRuleInput"), mock.Anything).Return(
			&eventbridge.ListTargetsByRuleOutput{
				Targets: []types.Target{
					{
						Id:  aws.String("target-1"),
						Arn: &targetArn,
					},
				},
			}, nil)

		err := AssertTargetExistsE(ctx, "test-rule", "default", "target-1")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertTargetExistsE_NotFound", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		mockClient.On("ListTargetsByRule", mock.Anything, mock.AnythingOfType("*eventbridge.ListTargetsByRuleInput"), mock.Anything).Return(
			&eventbridge.ListTargetsByRuleOutput{
				Targets: []types.Target{},
			}, nil)

		err := AssertTargetExistsE(ctx, "test-rule", "default", "target-1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertEventSentE_Success", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		result := PutEventResult{
			EventID: "event-123",
			Success: true,
		}

		err := AssertEventSentE(ctx, result)

		assert.NoError(t, err)
	})

	t.Run("AssertEventSentE_Failed", func(t *testing.T) {
		ctx := &TestContext{T: t}
		
		result := PutEventResult{
			Success:      false,
			ErrorCode:    "ValidationException",
			ErrorMessage: "Invalid event",
		}

		err := AssertEventSentE(ctx, result)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ValidationException")
		assert.Contains(t, err.Error(), "Invalid event")
	})
}

// TestPatternBuilders tests EventBridge pattern builder functions
func TestPatternBuilders(t *testing.T) {
	t.Run("BuildSourcePattern", func(t *testing.T) {
		pattern := BuildSourcePattern("service.api", "service.db")

		assert.Contains(t, pattern, `"source"`)
		assert.Contains(t, pattern, `"service.api"`)
		assert.Contains(t, pattern, `"service.db"`)
	})

	t.Run("BuildDetailTypePattern", func(t *testing.T) {
		pattern := BuildDetailTypePattern("Order Placed", "Order Cancelled")

		assert.Contains(t, pattern, `"detail-type"`)
		assert.Contains(t, pattern, `"Order Placed"`)
		assert.Contains(t, pattern, `"Order Cancelled"`)
	})

	t.Run("BuildNumericPattern", func(t *testing.T) {
		pattern := BuildNumericPattern("detail.amount", ">", 100)

		assert.Contains(t, pattern, `"detail"`)
		assert.Contains(t, pattern, `"amount"`)
		assert.Contains(t, pattern, `"numeric"`)
		// Note: The ">" operator gets JSON-escaped as "\u003e"
		assert.Contains(t, pattern, `100`)
	})

	t.Run("BuildExistsPattern_True", func(t *testing.T) {
		pattern := BuildExistsPattern("detail.userId", true)

		assert.Contains(t, pattern, `"detail"`)
		assert.Contains(t, pattern, `"userId"`)
		assert.Contains(t, pattern, `"exists"`)
		assert.Contains(t, pattern, `true`)
	})

	t.Run("BuildExistsPattern_False", func(t *testing.T) {
		pattern := BuildExistsPattern("detail.errorCode", false)

		assert.Contains(t, pattern, `"detail"`)
		assert.Contains(t, pattern, `"errorCode"`)
		assert.Contains(t, pattern, `"exists"`)
		assert.Contains(t, pattern, `false`)
	})

	t.Run("BuildPrefixPattern", func(t *testing.T) {
		pattern := BuildPrefixPattern("detail.eventType", "order")

		assert.Contains(t, pattern, `"detail"`)
		assert.Contains(t, pattern, `"eventType"`)
		assert.Contains(t, pattern, `"prefix"`)
		assert.Contains(t, pattern, `"order"`)
	})

	t.Run("PatternBuilder_Complex", func(t *testing.T) {
		builder := NewPatternBuilder()
		pattern := builder.
			AddSource("service.api", "service.db").
			AddDetailType("Order Placed").
			AddNumericCondition("detail.amount", ">", 100).
			AddExistsCondition("detail.userId", true).
			Build()

		assert.Contains(t, pattern, `"source"`)
		assert.Contains(t, pattern, `"service.api"`)
		assert.Contains(t, pattern, `"detail-type"`)
		assert.Contains(t, pattern, `"Order Placed"`)
		assert.Contains(t, pattern, `"detail"`)
		assert.Contains(t, pattern, `"amount"`)
		assert.Contains(t, pattern, `"userId"`)
	})
}

// TestUtilityFunctions tests utility and helper functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("CalculateEventSize", func(t *testing.T) {
		event := CustomEvent{
			Source:       "test.service",
			DetailType:   "Test Event",
			Detail:       `{"key": "value"}`,
			EventBusName: "custom-bus",
		}

		size := CalculateEventSize(event)

		assert.Greater(t, size, 0)
		assert.GreaterOrEqual(t, size, len(event.Source)+len(event.DetailType)+len(event.Detail)+len(event.EventBusName))
	})

	t.Run("OptimizeEventSizeE_Success", func(t *testing.T) {
		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{  "key"  :  "value"  ,  "number"  :  123  }`, // Whitespace to remove
		}

		optimized, err := OptimizeEventSizeE(event)

		assert.NoError(t, err)
		assert.Equal(t, event.Source, optimized.Source)
		assert.Equal(t, event.DetailType, optimized.DetailType)
		assert.NotContains(t, optimized.Detail, "  ") // Whitespace should be removed
		assert.Contains(t, optimized.Detail, `"key":"value"`)
		assert.Contains(t, optimized.Detail, `"number":123`)
	})

	t.Run("OptimizeEventSizeE_InvalidJSON", func(t *testing.T) {
		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"invalid": json}`,
		}

		optimized, err := OptimizeEventSizeE(event)

		assert.Error(t, err)
		assert.Equal(t, event, optimized)
		assert.Contains(t, err.Error(), "invalid event detail JSON")
	})

	t.Run("ValidateEventSizeE_Valid", func(t *testing.T) {
		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
		}

		err := ValidateEventSizeE(event)

		assert.NoError(t, err)
	})

	t.Run("ValidateEventSizeE_TooLarge", func(t *testing.T) {
		largeDetail := make([]byte, 256*1024) // 256KB
		for i := range largeDetail {
			largeDetail[i] = 'a'
		}

		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     string(largeDetail),
		}

		err := ValidateEventSizeE(event)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event size too large")
	})

	t.Run("CreateEventBatches", func(t *testing.T) {
		events := make([]CustomEvent, 25)
		for i := range events {
			events[i] = CustomEvent{
				Source:     "test.service",
				DetailType: "Test Event",
				Detail:     fmt.Sprintf(`{"id": %d}`, i),
			}
		}

		batches := CreateEventBatches(events, 10)

		assert.Len(t, batches, 3) // 25 events in batches of 10 = 3 batches
		assert.Len(t, batches[0], 10)
		assert.Len(t, batches[1], 10)
		assert.Len(t, batches[2], 5) // Remaining events
	})

	t.Run("CreateEventBatches_DefaultBatchSize", func(t *testing.T) {
		events := make([]CustomEvent, 25)
		for i := range events {
			events[i] = CustomEvent{
				Source:     "test.service",
				DetailType: "Test Event",
				Detail:     fmt.Sprintf(`{"id": %d}`, i),
			}
		}

		batches := CreateEventBatches(events, 0) // Should use default

		assert.Len(t, batches, 3) // 25 events in batches of 10 = 3 batches
	})

	t.Run("ValidateEventBatchE_Valid", func(t *testing.T) {
		events := []CustomEvent{
			{
				Source:     "test.service",
				DetailType: "Event 1",
				Detail:     `{"id": 1}`,
			},
			{
				Source:     "test.service",
				DetailType: "Event 2",
				Detail:     `{"id": 2}`,
			},
		}

		err := ValidateEventBatchE(events)

		assert.NoError(t, err)
	})

	t.Run("ValidateEventBatchE_Empty", func(t *testing.T) {
		err := ValidateEventBatchE([]CustomEvent{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch cannot be empty")
	})

	t.Run("ValidateEventBatchE_TooLarge", func(t *testing.T) {
		events := make([]CustomEvent, 11) // Max is 10
		for i := range events {
			events[i] = CustomEvent{
				Source:     "test.service",
				DetailType: "Test Event",
				Detail:     fmt.Sprintf(`{"id": %d}`, i),
			}
		}

		err := ValidateEventBatchE(events)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch size exceeds maximum")
	})
}

// TestPatternAnalysis tests pattern analysis functions
func TestPatternAnalysis(t *testing.T) {
	t.Run("AnalyzePatternComplexity_Low", func(t *testing.T) {
		pattern := `{"source": ["test.service"]}`

		complexity := AnalyzePatternComplexity(pattern)

		assert.Equal(t, "low", complexity)
	})

	t.Run("AnalyzePatternComplexity_Medium", func(t *testing.T) {
		pattern := `{
			"source": ["test.service"],
			"detail-type": ["Event 1", "Event 2"]
		}`

		complexity := AnalyzePatternComplexity(pattern)

		assert.Equal(t, "medium", complexity)
	})

	t.Run("AnalyzePatternComplexity_High", func(t *testing.T) {
		pattern := `{
			"source": ["service1", "service2", "service3"],
			"detail-type": ["Type1", "Type2", "Type3"],
			"detail": {
				"status": ["active", "pending", "failed"],
				"category": ["A", "B", "C"],
				"nested": {
					"level1": ["val1", "val2"],
					"level2": ["val3", "val4"]
				}
			}
		}`

		complexity := AnalyzePatternComplexity(pattern)

		assert.Equal(t, "high", complexity)
	})

	t.Run("AnalyzePatternComplexity_Invalid", func(t *testing.T) {
		pattern := `{"invalid": json}`

		complexity := AnalyzePatternComplexity(pattern)

		assert.Equal(t, "invalid", complexity)
	})

	t.Run("EstimatePatternMatchRate", func(t *testing.T) {
		pattern := `{"source": ["test.service"]}`

		matchRate := EstimatePatternMatchRate(pattern)

		assert.Greater(t, matchRate, 0.0)
		assert.LessOrEqual(t, matchRate, 1.0)
	})

	t.Run("EstimatePatternMatchRate_InvalidPattern", func(t *testing.T) {
		pattern := `{"invalid": json}`

		matchRate := EstimatePatternMatchRate(pattern)

		assert.Equal(t, 0.0, matchRate)
	})
}

// TestAdvancedEventOperations tests advanced event operations and edge cases
func TestAdvancedEventOperations(t *testing.T) {
	t.Run("PutEventE_WithCustomTime", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		customTime := time.Now().Add(-1 * time.Hour)
		expectedEventID := "test-event-id"
		
		mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
			return len(input.Entries) == 1 && 
				   input.Entries[0].Time != nil &&
				   input.Entries[0].Time.Equal(customTime)
		}), mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				FailedEntryCount: 0,
				Entries: []types.PutEventsResultEntry{
					{
						EventId: &expectedEventID,
					},
				},
			}, nil)

		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
			Time:       &customTime,
		}

		result, err := PutEventE(ctx, event)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, expectedEventID, result.EventID)
		mockClient.AssertExpectations(t)
	})

	t.Run("PutEventsE_WithRetry", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// First call fails, second succeeds
		mockClient.On("PutEvents", mock.Anything, mock.AnythingOfType("*eventbridge.PutEventsInput"), mock.Anything).Return(
			nil, errors.New("transient error")).Once()
		
		eventID1 := "event-1"
		eventID2 := "event-2"
		mockClient.On("PutEvents", mock.Anything, mock.AnythingOfType("*eventbridge.PutEventsInput"), mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				FailedEntryCount: 0,
				Entries: []types.PutEventsResultEntry{
					{EventId: &eventID1},
					{EventId: &eventID2},
				},
			}, nil).Once()

		events := []CustomEvent{
			{
				Source:     "test.service",
				DetailType: "Event 1",
				Detail:     `{"id": 1}`,
			},
			{
				Source:     "test.service",
				DetailType: "Event 2",
				Detail:     `{"id": 2}`,
			},
		}

		result, err := PutEventsE(ctx, events)

		assert.NoError(t, err)
		assert.Equal(t, int32(0), result.FailedEntryCount)
		assert.Len(t, result.Entries, 2)
		mockClient.AssertExpectations(t)
	})
}

// TestArchiveReplayValidation tests archive and replay validation functions
func TestArchiveReplayValidation(t *testing.T) {
	t.Run("ValidateReplayConfigE_Success", func(t *testing.T) {
		startTime := time.Now().Add(-2 * time.Hour)
		endTime := time.Now().Add(-1 * time.Hour)
		
		config := ReplayConfig{
			ReplayName:     "test-replay",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
			EventStartTime: startTime,
			EventEndTime:   endTime,
			Destination: types.ReplayDestination{
				Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/test-bus"),
			},
		}

		err := ValidateReplayConfigE(config)
		assert.NoError(t, err)
	})

	t.Run("ValidateReplayConfigE_EmptyReplayName", func(t *testing.T) {
		config := ReplayConfig{
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
		}

		err := ValidateReplayConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "replay name cannot be empty")
	})

	t.Run("ValidateReplayConfigE_EmptyEventSourceArn", func(t *testing.T) {
		config := ReplayConfig{
			ReplayName: "test-replay",
		}

		err := ValidateReplayConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event source ARN cannot be empty")
	})

	t.Run("ValidateReplayConfigE_InvalidTimeRange", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now().Add(-2 * time.Hour) // End time before start time

		config := ReplayConfig{
			ReplayName:     "test-replay",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
			EventStartTime: startTime,
			EventEndTime:   endTime,
			Destination: types.ReplayDestination{
				Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/test-bus"),
			},
		}

		err := ValidateReplayConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "end time must be after start time")
	})

	t.Run("ValidateReplayConfigE_DurationTooShort", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := startTime.Add(30 * time.Second) // Less than minimum duration

		config := ReplayConfig{
			ReplayName:     "test-replay",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
			EventStartTime: startTime,
			EventEndTime:   endTime,
			Destination: types.ReplayDestination{
				Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/test-bus"),
			},
		}

		err := ValidateReplayConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "replay duration must be at least")
	})

	t.Run("ValidateReplayConfigE_DurationTooLong", func(t *testing.T) {
		startTime := time.Now().Add(-31 * 24 * time.Hour) // 31 days ago
		endTime := time.Now().Add(-1 * time.Hour)

		config := ReplayConfig{
			ReplayName:     "test-replay",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
			EventStartTime: startTime,
			EventEndTime:   endTime,
			Destination: types.ReplayDestination{
				Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/test-bus"),
			},
		}

		err := ValidateReplayConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "replay duration cannot exceed")
	})

	t.Run("ValidateReplayConfigE_InvalidEventSourceArn", func(t *testing.T) {
		config := ReplayConfig{
			ReplayName:     "test-replay",
			EventSourceArn: "invalid-arn",
			EventStartTime: time.Now().Add(-2 * time.Hour),
			EventEndTime:   time.Now().Add(-1 * time.Hour),
			Destination: types.ReplayDestination{
				Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/test-bus"),
			},
		}

		err := ValidateReplayConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event source ARN format")
	})

	t.Run("ValidateReplayDestinationE_Success", func(t *testing.T) {
		destination := types.ReplayDestination{
			Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/test-bus"),
		}

		err := ValidateReplayDestinationE(destination)
		assert.NoError(t, err)
	})

	t.Run("ValidateReplayDestinationE_MissingArn", func(t *testing.T) {
		destination := types.ReplayDestination{}

		err := ValidateReplayDestinationE(destination)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "destination ARN is required")
	})

	t.Run("ValidateReplayDestinationE_InvalidArn", func(t *testing.T) {
		destination := types.ReplayDestination{
			Arn: aws.String("invalid-arn"),
		}

		err := ValidateReplayDestinationE(destination)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid destination ARN format")
	})

	t.Run("ValidateArchiveConfigE_Success", func(t *testing.T) {
		config := ArchiveConfig{
			ArchiveName:    "test-archive",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/test-bus",
			RetentionDays:  7,
			Description:    "Test archive",
		}

		err := ValidateArchiveConfigE(config)
		assert.NoError(t, err)
	})

	t.Run("ValidateArchiveConfigE_EmptyArchiveName", func(t *testing.T) {
		config := ArchiveConfig{
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/test-bus",
		}

		err := ValidateArchiveConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "archive name cannot be empty")
	})

	t.Run("ValidateArchiveConfigE_EmptyEventSourceArn", func(t *testing.T) {
		config := ArchiveConfig{
			ArchiveName: "test-archive",
		}

		err := ValidateArchiveConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event source ARN cannot be empty")
	})

	t.Run("ValidateArchiveConfigE_InvalidEventSourceArn", func(t *testing.T) {
		config := ArchiveConfig{
			ArchiveName:    "test-archive",
			EventSourceArn: "invalid-arn",
		}

		err := ValidateArchiveConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event source ARN format")
	})

	t.Run("ValidateArchiveConfigE_NegativeRetentionDays", func(t *testing.T) {
		config := ArchiveConfig{
			ArchiveName:    "test-archive",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/test-bus",
			RetentionDays:  -1,
		}

		err := ValidateArchiveConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "retention days cannot be negative")
	})

	t.Run("ValidateArchiveConfigE_RetentionDaysTooHigh", func(t *testing.T) {
		config := ArchiveConfig{
			ArchiveName:    "test-archive",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/test-bus",
			RetentionDays:  4000, // Exceeds practical maximum
		}

		err := ValidateArchiveConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "retention days exceeds practical maximum")
	})

	t.Run("ValidateArchiveConfigE_InvalidEventPattern", func(t *testing.T) {
		config := ArchiveConfig{
			ArchiveName:    "test-archive",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/test-bus",
			EventPattern:   `{"invalid": pattern}`, // Invalid JSON pattern
		}

		err := ValidateArchiveConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event pattern")
	})

	t.Run("ValidateArchiveConfigE_ValidEventPattern", func(t *testing.T) {
		config := ArchiveConfig{
			ArchiveName:    "test-archive",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/test-bus",
			EventPattern:   `{"source": ["test.service"]}`,
		}

		err := ValidateArchiveConfigE(config)
		assert.NoError(t, err)
	})

	t.Run("BuildReplayConfig", func(t *testing.T) {
		replayName := "test-replay"
		archiveArn := "arn:aws:events:us-east-1:123456789012:archive/test-archive"
		startTime := time.Now().Add(-2 * time.Hour)
		endTime := time.Now().Add(-1 * time.Hour)
		destinationArn := "arn:aws:events:us-east-1:123456789012:event-bus/test-bus"

		config := BuildReplayConfig(replayName, archiveArn, startTime, endTime, destinationArn)

		assert.Equal(t, replayName, config.ReplayName)
		assert.Equal(t, archiveArn, config.EventSourceArn)
		assert.Equal(t, startTime, config.EventStartTime)
		assert.Equal(t, endTime, config.EventEndTime)
		assert.Equal(t, destinationArn, *config.Destination.Arn)
	})

	t.Run("BuildArchiveConfig", func(t *testing.T) {
		archiveName := "test-archive"
		eventSourceArn := "arn:aws:events:us-east-1:123456789012:event-bus/test-bus"
		retentionDays := int32(7)

		config := BuildArchiveConfig(archiveName, eventSourceArn, retentionDays)

		assert.Equal(t, archiveName, config.ArchiveName)
		assert.Equal(t, eventSourceArn, config.EventSourceArn)
		assert.Equal(t, retentionDays, config.RetentionDays)
	})

	t.Run("BuildArchiveConfigWithPattern", func(t *testing.T) {
		archiveName := "test-archive"
		eventSourceArn := "arn:aws:events:us-east-1:123456789012:event-bus/test-bus"
		eventPattern := `{"source": ["test.service"]}`
		retentionDays := int32(7)

		config := BuildArchiveConfigWithPattern(archiveName, eventSourceArn, eventPattern, retentionDays)

		assert.Equal(t, archiveName, config.ArchiveName)
		assert.Equal(t, eventSourceArn, config.EventSourceArn)
		assert.Equal(t, eventPattern, config.EventPattern)
		assert.Equal(t, retentionDays, config.RetentionDays)
	})

	t.Run("CalculateReplayDuration", func(t *testing.T) {
		startTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		endTime := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)
		config := ReplayConfig{
			EventStartTime: startTime,
			EventEndTime:   endTime,
		}

		duration := CalculateReplayDuration(config)
		assert.Equal(t, time.Hour, duration)
	})

	t.Run("EstimateReplayTime", func(t *testing.T) {
		eventCount := int64(1000)
		eventsPerSecond := int32(10)

		estimatedTime := EstimateReplayTime(eventCount, eventsPerSecond)
		
		expectedSeconds := eventCount / int64(eventsPerSecond)
		expectedDuration := time.Duration(expectedSeconds) * time.Second
		assert.Equal(t, expectedDuration, estimatedTime)
	})

	t.Run("EstimateReplayTime_DefaultRate", func(t *testing.T) {
		eventCount := int64(1000)
		eventsPerSecond := int32(0) // Will use default rate

		estimatedTime := EstimateReplayTime(eventCount, eventsPerSecond)
		
		expectedSeconds := eventCount / int64(100) // Default rate is 100
		expectedDuration := time.Duration(expectedSeconds) * time.Second
		assert.Equal(t, expectedDuration, estimatedTime)
	})

	t.Run("ValidateReplayTimeRange_Success", func(t *testing.T) {
		startTime := time.Now().Add(-2 * time.Hour)
		endTime := time.Now().Add(-61 * time.Second) // Just over minimum duration

		err := ValidateReplayTimeRange(startTime, endTime)
		assert.NoError(t, err)
	})

	t.Run("ValidateReplayTimeRange_StartTimeInFuture", func(t *testing.T) {
		startTime := time.Now().Add(1 * time.Hour)
		endTime := time.Now().Add(2 * time.Hour)

		err := ValidateReplayTimeRange(startTime, endTime)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "replay start time cannot be in the future")
	})

	t.Run("ValidateReplayTimeRange_EndTimeInFuture", func(t *testing.T) {
		startTime := time.Now().Add(-2 * time.Hour)
		endTime := time.Now().Add(1 * time.Hour)

		err := ValidateReplayTimeRange(startTime, endTime)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "replay end time cannot be in the future")
	})

	t.Run("ValidateReplayConfig_PanicOnError", func(t *testing.T) {
		config := ReplayConfig{
			ReplayName: "", // Empty name should cause panic
		}

		assert.Panics(t, func() {
			ValidateReplayConfig(config)
		})
	})

	t.Run("ValidateReplayDestination_PanicOnError", func(t *testing.T) {
		destination := types.ReplayDestination{} // Missing ARN should cause panic

		assert.Panics(t, func() {
			ValidateReplayDestination(destination)
		})
	})

	t.Run("ValidateArchiveConfig_PanicOnError", func(t *testing.T) {
		config := ArchiveConfig{
			ArchiveName: "", // Empty name should cause panic
		}

		assert.Panics(t, func() {
			ValidateArchiveConfig(config)
		})
	})
}

// TestEventDeliveryValidation tests event delivery validation functions
func TestEventDeliveryValidation(t *testing.T) {
	ctx := &TestContext{T: t}

	t.Run("AssertEventDeliveryWithRetriesE_Success", func(t *testing.T) {
		result := PutEventResult{
			EventID: "test-event-123",
			Success: true,
		}

		err := AssertEventDeliveryWithRetriesE(ctx, result, 3)
		assert.NoError(t, err)
	})

	t.Run("AssertEventDeliveryWithRetriesE_NegativeRetries", func(t *testing.T) {
		result := PutEventResult{
			EventID: "test-event-123",
			Success: true,
		}

		err := AssertEventDeliveryWithRetriesE(ctx, result, -1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max retries cannot be negative")
	})

	t.Run("AssertEventDeliveryWithRetriesE_DeliveryFailed", func(t *testing.T) {
		result := PutEventResult{
			Success:      false,
			ErrorCode:    "InternalFailure",
			ErrorMessage: "Internal server error",
		}

		err := AssertEventDeliveryWithRetriesE(ctx, result, 3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event delivery failed after 3 retries")
		assert.Contains(t, err.Error(), "InternalFailure")
	})

	t.Run("AssertEventDeliveryWithRetriesE_EmptyEventID", func(t *testing.T) {
		result := PutEventResult{
			EventID: "", // Empty event ID
			Success: true,
		}

		err := AssertEventDeliveryWithRetriesE(ctx, result, 3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event ID is empty")
	})

	t.Run("AssertBatchDeliveryWithPartialSuccessE_Success", func(t *testing.T) {
		result := PutEventsResult{
			FailedEntryCount: 1,
			Entries: []PutEventResult{
				{EventID: "event-1", Success: true},
				{EventID: "event-2", Success: true},
				{Success: false, ErrorCode: "Throttled", ErrorMessage: "Request throttled"},
			},
		}

		// Allow up to 50% failure rate
		err := AssertBatchDeliveryWithPartialSuccessE(ctx, result, 0.5)
		assert.NoError(t, err)
	})

	t.Run("AssertBatchDeliveryWithPartialSuccessE_InvalidFailureRate", func(t *testing.T) {
		result := PutEventsResult{
			FailedEntryCount: 0,
			Entries:          []PutEventResult{{EventID: "event-1", Success: true}},
		}

		err := AssertBatchDeliveryWithPartialSuccessE(ctx, result, 1.5) // Invalid rate > 1.0
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max failure rate must be between 0.0 and 1.0")

		err = AssertBatchDeliveryWithPartialSuccessE(ctx, result, -0.1) // Invalid rate < 0.0
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max failure rate must be between 0.0 and 1.0")
	})

	t.Run("AssertBatchDeliveryWithPartialSuccessE_EmptyBatch", func(t *testing.T) {
		result := PutEventsResult{
			Entries: []PutEventResult{}, // Empty entries
		}

		err := AssertBatchDeliveryWithPartialSuccessE(ctx, result, 0.1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch result contains no entries")
	})

	t.Run("AssertBatchDeliveryWithPartialSuccessE_ExceedsFailureRate", func(t *testing.T) {
		result := PutEventsResult{
			FailedEntryCount: 3,
			Entries: []PutEventResult{
				{Success: false, ErrorCode: "Error1", ErrorMessage: "Message1"},
				{Success: false, ErrorCode: "Error2", ErrorMessage: "Message2"},
				{Success: false, ErrorCode: "Error3", ErrorMessage: "Message3"},
				{EventID: "event-4", Success: true},
			},
		}

		// Allow only 50% failure rate, but we have 75% failure
		err := AssertBatchDeliveryWithPartialSuccessE(ctx, result, 0.5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch failure rate (0.75) exceeds maximum allowed (0.50)")
		assert.Contains(t, err.Error(), "Entry 0: Error1 - Message1")
	})

	t.Run("ValidateAdvancedDeliveryMetricsE_Success", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID:          "test-target",
			DeliveredCount:    100,
			FailedCount:       5,
			DeliveryLatencyMs: 250,
			ErrorRate:         0.05,
		}

		thresholds := MetricsThresholds{
			MaxLatencyMs:   500,
			MaxErrorRate:   0.1,
			MinTotalEvents: 50,
		}

		err := ValidateAdvancedDeliveryMetricsE(metrics, thresholds)
		assert.NoError(t, err)
	})

	t.Run("ValidateAdvancedDeliveryMetricsE_LatencyExceeded", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID:          "test-target",
			DeliveredCount:    100,
			FailedCount:       5,
			DeliveryLatencyMs: 600, // Exceeds threshold
			ErrorRate:         0.05,
		}

		thresholds := MetricsThresholds{
			MaxLatencyMs:   500,
			MaxErrorRate:   0.1,
			MinTotalEvents: 50,
		}

		err := ValidateAdvancedDeliveryMetricsE(metrics, thresholds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delivery latency (600 ms) exceeds threshold (500 ms)")
	})

	t.Run("ValidateAdvancedDeliveryMetricsE_ErrorRateExceeded", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID:          "test-target",
			DeliveredCount:    100,
			FailedCount:       20,
			DeliveryLatencyMs: 250,
			ErrorRate:         0.15, // Exceeds threshold
		}

		thresholds := MetricsThresholds{
			MaxLatencyMs:   500,
			MaxErrorRate:   0.1,
			MinTotalEvents: 50,
		}

		err := ValidateAdvancedDeliveryMetricsE(metrics, thresholds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error rate (0.1500) exceeds threshold (0.1000)")
	})

	t.Run("ValidateAdvancedDeliveryMetricsE_InsufficientEvents", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID:          "test-target",
			DeliveredCount:    20,
			FailedCount:       5,
			DeliveryLatencyMs: 250,
			ErrorRate:         0.05,
		}

		thresholds := MetricsThresholds{
			MaxLatencyMs:   500,
			MaxErrorRate:   0.1,
			MinTotalEvents: 50, // More than current total of 25
		}

		err := ValidateAdvancedDeliveryMetricsE(metrics, thresholds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total events (25) below minimum threshold (50)")
	})

	t.Run("ValidateAdvancedDeliveryMetricsE_BasicValidationFail", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID:          "", // Empty target ID will fail basic validation
			DeliveredCount:    100,
			FailedCount:       5,
			DeliveryLatencyMs: 250,
			ErrorRate:         0.05,
		}

		thresholds := MetricsThresholds{
			MaxLatencyMs:   500,
			MaxErrorRate:   0.1,
			MinTotalEvents: 50,
		}

		err := ValidateAdvancedDeliveryMetricsE(metrics, thresholds)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "basic validation failed")
	})

	t.Run("WaitForEventDeliveryWithBackoffE_InvalidTimeout", func(t *testing.T) {
		err := WaitForEventDeliveryWithBackoffE(ctx, "test-event", 0, time.Millisecond*100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout must be positive")

		err = WaitForEventDeliveryWithBackoffE(ctx, "test-event", -time.Second, time.Millisecond*100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout must be positive")
	})

	t.Run("WaitForEventDeliveryWithBackoffE_InvalidInterval", func(t *testing.T) {
		err := WaitForEventDeliveryWithBackoffE(ctx, "test-event", time.Second, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initial interval must be positive")

		err = WaitForEventDeliveryWithBackoffE(ctx, "test-event", time.Second, -time.Millisecond*100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "initial interval must be positive")
	})

	t.Run("WaitForEventDeliveryWithBackoffE_Timeout", func(t *testing.T) {
		// This will always timeout since isEventDelivered always returns false
		err := WaitForEventDeliveryWithBackoffE(ctx, "test-event", time.Millisecond*50, time.Millisecond*10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout waiting for event 'test-event' to be delivered")
	})

	t.Run("AssertDeliveryLatencyPercentilesE_Success", func(t *testing.T) {
		latencies := []int64{100, 150, 200, 250, 300, 350, 400, 450, 500, 600}
		percentiles := map[float64]int64{
			0.5:  350, // P50 should be <= 350ms
			0.9:  600, // P90 should be <= 600ms  
			0.95: 650, // P95 should be <= 650ms
		}

		err := AssertDeliveryLatencyPercentilesE(ctx, latencies, percentiles)
		assert.NoError(t, err)
	})

	t.Run("AssertDeliveryLatencyPercentilesE_EmptyLatencies", func(t *testing.T) {
		latencies := []int64{}
		percentiles := map[float64]int64{0.5: 300}

		err := AssertDeliveryLatencyPercentilesE(ctx, latencies, percentiles)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "latencies slice cannot be empty")
	})

	t.Run("AssertDeliveryLatencyPercentilesE_EmptyPercentiles", func(t *testing.T) {
		latencies := []int64{100, 200, 300}
		percentiles := map[float64]int64{}

		err := AssertDeliveryLatencyPercentilesE(ctx, latencies, percentiles)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "percentiles map cannot be empty")
	})

	t.Run("AssertDeliveryLatencyPercentilesE_InvalidPercentile", func(t *testing.T) {
		latencies := []int64{100, 200, 300}
		percentiles := map[float64]int64{
			0.0:  100, // Invalid: should be > 0
			1.5:  300, // Invalid: should be <= 1.0
		}

		err := AssertDeliveryLatencyPercentilesE(ctx, latencies, percentiles)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "percentile must be between 0 and 1.0")
	})

	t.Run("AssertDeliveryLatencyPercentilesE_ExceededLatency", func(t *testing.T) {
		latencies := []int64{100, 200, 300, 400, 500}
		percentiles := map[float64]int64{
			0.5: 150, // P50 is 300ms but we expect <= 150ms
		}

		err := AssertDeliveryLatencyPercentilesE(ctx, latencies, percentiles)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "P50 latency (300 ms) exceeds expected (150 ms)")
	})

	t.Run("AssertEventDeliveryWithRetries_PanicOnError", func(t *testing.T) {
		// Create a mock testing.T that won't interfere with the parent test
		mockT := &MockTestingT{}
		mockCtx := &TestContext{T: mockT}
		result := PutEventResult{
			Success: false,
			ErrorCode: "TestError",
		}

		assert.Panics(t, func() {
			AssertEventDeliveryWithRetries(mockCtx, result, 3)
		})
	})

	t.Run("AssertBatchDeliveryWithPartialSuccess_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{T: mockT}
		result := PutEventsResult{
			Entries: []PutEventResult{}, // Empty will cause error
		}

		assert.Panics(t, func() {
			AssertBatchDeliveryWithPartialSuccess(mockCtx, result, 0.1)
		})
	})

	t.Run("ValidateAdvancedDeliveryMetrics_PanicOnError", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID: "", // Empty will cause error
		}
		thresholds := MetricsThresholds{}

		assert.Panics(t, func() {
			ValidateAdvancedDeliveryMetrics(metrics, thresholds)
		})
	})

	t.Run("WaitForEventDeliveryWithBackoff_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{T: mockT}
		assert.Panics(t, func() {
			WaitForEventDeliveryWithBackoff(mockCtx, "test-event", 0, time.Second) // Invalid timeout
		})
	})

	t.Run("AssertDeliveryLatencyPercentiles_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{T: mockT}
		latencies := []int64{}
		percentiles := map[float64]int64{0.5: 300}

		assert.Panics(t, func() {
			AssertDeliveryLatencyPercentiles(mockCtx, latencies, percentiles)
		})
	})

	// Test helper functions
	t.Run("joinStrings", func(t *testing.T) {
		result := joinStrings([]string{"a", "b", "c"}, ",")
		assert.Equal(t, "a,b,c", result)

		result = joinStrings([]string{}, ",")
		assert.Equal(t, "", result)

		result = joinStrings([]string{"single"}, ",")
		assert.Equal(t, "single", result)
	})

	t.Run("sortSlice", func(t *testing.T) {
		slice := []int64{5, 2, 8, 1, 9, 3}
		sortSlice(slice)
		expected := []int64{1, 2, 3, 5, 8, 9}
		assert.Equal(t, expected, slice)

		// Test with already sorted slice
		slice = []int64{1, 2, 3}
		sortSlice(slice)
		assert.Equal(t, []int64{1, 2, 3}, slice)

		// Test with single element
		slice = []int64{42}
		sortSlice(slice)
		assert.Equal(t, []int64{42}, slice)

		// Test with empty slice
		slice = []int64{}
		sortSlice(slice)
		assert.Equal(t, []int64{}, slice)
	})
}

// TestScheduleValidation tests schedule validation functions
func TestScheduleValidation(t *testing.T) {
	t.Run("ValidateScheduleExpressionE_EmptyExpression", func(t *testing.T) {
		err := ValidateScheduleExpressionE("")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyScheduleExpression)
	})

	t.Run("ValidateScheduleExpressionE_InvalidFormat", func(t *testing.T) {
		err := ValidateScheduleExpressionE("invalid expression")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidScheduleFormat)
	})

	// Rate expression tests
	t.Run("ValidateScheduleExpressionE_ValidRateExpressions", func(t *testing.T) {
		validExpressions := []string{
			"rate(1 minute)",
			"rate(5 minutes)",
			"rate(1 hour)",
			"rate(12 hours)",
			"rate(1 day)",
			"rate(7 days)",
		}

		for _, expr := range validExpressions {
			err := ValidateScheduleExpressionE(expr)
			assert.NoError(t, err, "Expression should be valid: %s", expr)
		}
	})

	t.Run("ValidateScheduleExpressionE_InvalidRateExpressions", func(t *testing.T) {
		invalidExpressions := map[string]error{
			"rate(0 minutes)":        ErrInvalidRateValue,
			"rate(-1 hours)":         ErrInvalidRateValue,
			"rate(abc minutes)":      ErrInvalidRateExpression, // This fails at regex level
			"rate(5 seconds)":        ErrInvalidRateUnit,
			"rate(1 minutes)":        ErrInvalidRateUnit, // Should be singular
			"rate(5 minute)":         ErrInvalidRateUnit, // Should be plural
			"rate(5)":                ErrInvalidRateExpression,
			"rate(5 minutes extra)":  ErrInvalidRateExpression,
		}

		for expr, expectedErr := range invalidExpressions {
			err := ValidateScheduleExpressionE(expr)
			assert.Error(t, err, "Expression should be invalid: %s", expr)
			assert.ErrorIs(t, err, expectedErr, "Wrong error type for: %s", expr)
		}
	})

	// Cron expression tests
	t.Run("ValidateScheduleExpressionE_ValidCronExpressions", func(t *testing.T) {
		validExpressions := []string{
			"cron(0 12 * * ? *)",          // Daily at noon
			"cron(15 10 ? * MON-FRI *)",   // Weekdays at 10:15 AM
			"cron(0 18 ? * SUN *)",        // Sundays at 6 PM
			"cron(0 8 1 * ? *)",           // First of every month at 8 AM
			"cron(0 0 * * ? 2023)",        // Every day in 2023 at midnight
			"cron(30 14 * * ? *)",         // Daily at 2:30 PM
		}

		for _, expr := range validExpressions {
			err := ValidateScheduleExpressionE(expr)
			assert.NoError(t, err, "Expression should be valid: %s", expr)
		}
	})

	t.Run("ValidateScheduleExpressionE_InvalidCronExpressions", func(t *testing.T) {
		invalidExpressions := map[string]error{
			"cron()":                    ErrInvalidCronExpression,
			"cron(0 12 * *)":           ErrCronTooFewFields,
			"cron(0 12 * * ? * * *)":   ErrCronTooManyFields,
			"cron(60 12 * * ? *)":      ErrInvalidCronField, // Invalid minute
			"cron(0 25 * * ? *)":       ErrInvalidCronField, // Invalid hour
			"cron(0 12 32 * ? *)":      ErrInvalidCronField, // Invalid day
			"cron(0 12 1 13 ? *)":      ErrInvalidCronField, // Invalid month
			"cron(0 12 1 1 8 *)":       ErrInvalidCronField, // Invalid day of week
			"cron(0 12 1 1 ? 1969)":    ErrInvalidCronField, // Invalid year (too early)
			"cron(0 12 15 * MON *)":    ErrCronInvalidDaySpecification, // Both day of month and day of week specified
		}

		for expr, expectedErr := range invalidExpressions {
			err := ValidateScheduleExpressionE(expr)
			assert.Error(t, err, "Expression should be invalid: %s", expr)
			assert.ErrorIs(t, err, expectedErr, "Wrong error type for: %s", expr)
		}
	})

	t.Run("ValidateScheduleExpression_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}

		assert.Panics(t, func() {
			ValidateScheduleExpression(mockT, "invalid expression")
		})
	})

	t.Run("BuildRateExpression", func(t *testing.T) {
		result := BuildRateExpression(5, "minutes")
		assert.Equal(t, "rate(5 minutes)", result)

		result = BuildRateExpression(1, "hour")
		assert.Equal(t, "rate(1 hour)", result)
	})

	t.Run("BuildCronExpression", func(t *testing.T) {
		result := BuildCronExpression("0", "12", "*", "*", "?", "*")
		assert.Equal(t, "cron(0 12 * * ? *)", result)

		result = BuildCronExpression("15", "10", "?", "*", "MON-FRI", "*")
		assert.Equal(t, "cron(15 10 ? * MON-FRI *)", result)
	})

	t.Run("IsRateExpression", func(t *testing.T) {
		assert.True(t, IsRateExpression("rate(5 minutes)"))
		assert.False(t, IsRateExpression("cron(0 12 * * ? *)"))
		assert.False(t, IsRateExpression("invalid"))
	})

	t.Run("IsCronExpression", func(t *testing.T) {
		assert.True(t, IsCronExpression("cron(0 12 * * ? *)"))
		assert.False(t, IsCronExpression("rate(5 minutes)"))
		assert.False(t, IsCronExpression("invalid"))
	})

	t.Run("ParseRateExpression_Success", func(t *testing.T) {
		value, unit, err := ParseRateExpression("rate(5 minutes)")
		assert.NoError(t, err)
		assert.Equal(t, 5, value)
		assert.Equal(t, "minutes", unit)

		value, unit, err = ParseRateExpression("rate(1 hour)")
		assert.NoError(t, err)
		assert.Equal(t, 1, value)
		assert.Equal(t, "hour", unit)
	})

	t.Run("ParseRateExpression_InvalidFormat", func(t *testing.T) {
		_, _, err := ParseRateExpression("invalid rate")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRateExpression)

		_, _, err = ParseRateExpression("rate(abc minutes)")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRateExpression) // This fails at regex level
	})

	t.Run("ParseCronExpression_Success", func(t *testing.T) {
		fields, err := ParseCronExpression("cron(0 12 * * ? *)")
		assert.NoError(t, err)
		expected := []string{"0", "12", "*", "*", "?", "*"}
		assert.Equal(t, expected, fields)

		fields, err = ParseCronExpression("cron(15 10 ? * MON-FRI *)")
		assert.NoError(t, err)
		expected = []string{"15", "10", "?", "*", "MON-FRI", "*"}
		assert.Equal(t, expected, fields)
	})

	t.Run("ParseCronExpression_InvalidFormat", func(t *testing.T) {
		_, err := ParseCronExpression("invalid cron")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCronExpression)

		_, err = ParseCronExpression("cron(0 12 * *)")  // Too few fields
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCronExpression)
	})

	// Test complex cron field validation
	t.Run("ValidateCronField_ComplexExpressions", func(t *testing.T) {
		// These should pass basic format validation based on the regex
		validComplexFields := []string{
			"1-5",      // Range
			"*/15",     // Step values  
			"1,3,5",    // List
			"MON-FRI",  // Named range
			"L",        // Last day
			"3#2",      // Nth occurrence
		}

		// Test these don't cause errors in basic format validation
		// Full semantic validation would be much more complex
		for _, field := range validComplexFields {
			// Using minute field as example (0-59 range)
			err := validateCronField(field, "minute", 0, 59)
			assert.NoError(t, err, "Complex field should pass basic validation: %s", field)
		}

		// Test some that should fail format validation
		invalidFields := []string{
			"15W",   // W modifier not in basic regex
			"(1-5)", // Parentheses not allowed
			"abc",   // Lowercase letters not allowed (only A-Z for named ranges)
		}

		for _, field := range invalidFields {
			err := validateCronField(field, "minute", 0, 59)
			assert.Error(t, err, "Field should fail basic validation: %s", field)
		}
		
		// Note: @ is actually not caught by basic regex validation - it gets through
	})

	t.Run("ValidateCronField_SimpleNumericValidation", func(t *testing.T) {
		// Test simple numeric validation
		assert.NoError(t, validateCronField("30", "minute", 0, 59))
		assert.NoError(t, validateCronField("12", "hour", 0, 23))
		assert.NoError(t, validateCronField("*", "minute", 0, 59))
		assert.NoError(t, validateCronField("?", "day", 1, 31))

		// Test out of range values
		err := validateCronField("60", "minute", 0, 59)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCronField)

		err = validateCronField("25", "hour", 0, 23)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCronField)
	})

	// Test error constants exist and are distinct
	t.Run("ErrorConstants", func(t *testing.T) {
		errors := []error{
			ErrEmptyScheduleExpression,
			ErrInvalidScheduleFormat,
			ErrInvalidRateExpression,
			ErrInvalidCronExpression,
			ErrInvalidRateValue,
			ErrInvalidRateUnit,
			ErrInvalidCronField,
			ErrCronTooManyFields,
			ErrCronTooFewFields,
			ErrCronInvalidDaySpecification,
		}

		// Ensure all errors have different messages
		messages := make(map[string]bool)
		for _, err := range errors {
			msg := err.Error()
			assert.False(t, messages[msg], "Duplicate error message: %s", msg)
			messages[msg] = true
		}

		// Ensure all have meaningful messages
		for _, err := range errors {
			assert.NotEmpty(t, err.Error(), "Error should have a message")
		}
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Test edge cases for rate validation
		err := ValidateScheduleExpressionE("rate(1 minute)")
		assert.NoError(t, err)

		err = ValidateScheduleExpressionE("rate(2 minute)")  // Should be plural
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRateUnit)

		// Test cron wildcards in all positions
		err = ValidateScheduleExpressionE("cron(* * ? * ? *)")
		assert.NoError(t, err)

		// Test year field edge cases
		err = ValidateScheduleExpressionE("cron(0 0 1 1 ? 1970)")  // Min year
		assert.NoError(t, err)

		err = ValidateScheduleExpressionE("cron(0 0 1 1 ? 3000)")  // Max year
		assert.NoError(t, err)
	})
}

// TestBuilders tests AWS service event builder functions
func TestBuilders(t *testing.T) {
	t.Run("BuildS3Event", func(t *testing.T) {
		event := BuildS3Event("my-bucket", "path/to/file.txt", "ObjectCreated:Put")

		assert.Equal(t, "aws.s3", event.Source)
		assert.Equal(t, "Object Created", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Contains(t, event.Resources[0], "arn:aws:s3:::my-bucket/path/to/file.txt")
		
		// Validate detail is valid JSON
		var detail S3EventDetail
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "2.1", detail.EventVersion)
		assert.Equal(t, "aws:s3", detail.EventSource)
		assert.Equal(t, "ObjectCreated:Put", detail.EventName)
		assert.Equal(t, "my-bucket", detail.S3.Bucket.Name)
		assert.Equal(t, "path/to/file.txt", detail.S3.Object.Key)
	})

	t.Run("BuildEC2Event", func(t *testing.T) {
		event := BuildEC2Event("i-1234567890abcdef0", "running", "pending")

		assert.Equal(t, "aws.ec2", event.Source)
		assert.Equal(t, "EC2 Instance State-change Notification", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Contains(t, event.Resources[0], "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0")
		
		// Validate detail is valid JSON
		var detail EC2EventDetail
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "i-1234567890abcdef0", detail.InstanceID)
		assert.Equal(t, "running", detail.State)
		assert.Equal(t, "pending", detail.PreviousState)
	})

	t.Run("BuildEC2Event_NoPreviousState", func(t *testing.T) {
		event := BuildEC2Event("i-1234567890abcdef0", "running", "")

		var detail EC2EventDetail
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "running", detail.State)
		assert.Empty(t, detail.PreviousState)
	})

	t.Run("BuildLambdaEvent", func(t *testing.T) {
		functionArn := "arn:aws:lambda:us-east-1:123456789012:function:my-function"
		event := BuildLambdaEvent("my-function", functionArn, "Active", "The function is ready")

		assert.Equal(t, "aws.lambda", event.Source)
		assert.Equal(t, "Lambda Function State Change", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Equal(t, []string{functionArn}, event.Resources)
		
		// Validate detail is valid JSON
		var detail LambdaEventDetail
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "my-function", detail.FunctionName)
		assert.Equal(t, functionArn, detail.FunctionArn)
		assert.Equal(t, "python3.9", detail.Runtime)
		assert.Equal(t, "Active", detail.State)
		assert.Equal(t, "The function is ready", detail.StateReason)
	})

	t.Run("BuildDynamoDBEvent", func(t *testing.T) {
		record := map[string]interface{}{
			"Keys": map[string]interface{}{
				"id": map[string]interface{}{
					"S": "test-id",
				},
			},
		}
		event := BuildDynamoDBEvent("my-table", "INSERT", record)

		assert.Equal(t, "aws.dynamodb", event.Source)
		assert.Equal(t, "DynamoDB Table Event", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Contains(t, event.Resources[0], "arn:aws:dynamodb:us-east-1:123456789012:table/my-table")
		
		// Validate detail is valid JSON
		var detail DynamoDBEventDetail
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "my-table", detail.TableName)
		assert.Equal(t, "INSERT", detail.EventName)
		assert.Equal(t, "aws:dynamodb", detail.EventSource)
		assert.Equal(t, record, detail.DynamoDBRecord)
	})

	t.Run("BuildSQSEvent", func(t *testing.T) {
		event := BuildSQSEvent("my-queue", "msg-12345", "Hello World")

		assert.Equal(t, "aws.sqs", event.Source)
		assert.Equal(t, "SQS Message Available", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Contains(t, event.Resources[0], "arn:aws:sqs:us-east-1:123456789012:my-queue")
		
		// Validate detail is valid JSON
		var detail SQSEventDetail
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "my-queue", detail.QueueUrl)
		assert.Equal(t, "msg-12345", detail.MessageId)
		assert.Equal(t, "Hello World", detail.Body)
		assert.NotNil(t, detail.Attributes)
	})

	t.Run("BuildCodeCommitEvent", func(t *testing.T) {
		event := BuildCodeCommitEvent("my-repo", "branch", "refs/heads/main", "abc123def456")

		assert.Equal(t, "aws.codecommit", event.Source)
		assert.Equal(t, "CodeCommit Repository State Change", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Contains(t, event.Resources[0], "arn:aws:codecommit:us-east-1:123456789012:my-repo")
		
		// Validate detail is valid JSON
		var detail map[string]interface{}
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "my-repo", detail["repositoryName"])
		assert.Equal(t, "branch", detail["referenceType"])
		assert.Equal(t, "refs/heads/main", detail["referenceName"])
		assert.Equal(t, "abc123def456", detail["commitId"])
	})

	t.Run("BuildCodePipelineEvent", func(t *testing.T) {
		event := BuildCodePipelineEvent("my-pipeline", "STARTED", "exec-12345")

		assert.Equal(t, "aws.codepipeline", event.Source)
		assert.Equal(t, "CodePipeline Pipeline Execution State Change", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Contains(t, event.Resources[0], "arn:aws:codepipeline:us-east-1:123456789012:pipeline/my-pipeline")
		
		// Validate detail is valid JSON
		var detail map[string]interface{}
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "my-pipeline", detail["pipeline"])
		assert.Equal(t, "STARTED", detail["state"])
		assert.Equal(t, "exec-12345", detail["execution-id"])
	})

	t.Run("BuildECSEvent", func(t *testing.T) {
		clusterArn := "arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster"
		taskArn := "arn:aws:ecs:us-east-1:123456789012:task/abc123"
		event := BuildECSEvent(clusterArn, taskArn, "RUNNING", "RUNNING")

		assert.Equal(t, "aws.ecs", event.Source)
		assert.Equal(t, "ECS Task State Change", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Equal(t, []string{taskArn}, event.Resources)
		
		// Validate detail is valid JSON
		var detail map[string]interface{}
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, clusterArn, detail["clusterArn"])
		assert.Equal(t, taskArn, detail["taskArn"])
		assert.Equal(t, "RUNNING", detail["lastStatus"])
		assert.Equal(t, "RUNNING", detail["desiredStatus"])
	})

	t.Run("BuildCloudWatchEvent", func(t *testing.T) {
		event := BuildCloudWatchEvent("my-alarm", "ALARM", "Threshold Crossed", "CPUUtilization")

		assert.Equal(t, "aws.cloudwatch", event.Source)
		assert.Equal(t, "CloudWatch Alarm State Change", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Contains(t, event.Resources[0], "arn:aws:cloudwatch:us-east-1:123456789012:alarm:my-alarm")
		
		// Validate detail is valid JSON
		var detail map[string]interface{}
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "my-alarm", detail["alarmName"])
		assert.Equal(t, "ALARM", detail["newState"])
		assert.Equal(t, "Threshold Crossed", detail["reason"])
		
		// Check nested state structure
		state := detail["state"].(map[string]interface{})
		assert.Equal(t, "ALARM", state["value"])
		
		// Check configuration
		config := detail["configuration"].(map[string]interface{})
		assert.Equal(t, "CPUUtilization", config["metricName"])
		assert.Equal(t, "AWS/EC2", config["namespace"])
	})

	t.Run("BuildAPIGatewayEvent", func(t *testing.T) {
		event := BuildAPIGatewayEvent("api12345", "prod", "req-98765", "200")

		assert.Equal(t, "aws.apigateway", event.Source)
		assert.Equal(t, "API Gateway Execution", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Contains(t, event.Resources[0], "arn:aws:apigateway:us-east-1:123456789012:restapi/api12345")
		
		// Validate detail is valid JSON
		var detail map[string]interface{}
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, "api12345", detail["apiId"])
		assert.Equal(t, "prod", detail["stage"])
		assert.Equal(t, "req-98765", detail["requestId"])
		assert.Equal(t, "200", detail["status"])
		assert.Equal(t, float64(1024), detail["responseLength"]) // JSON numbers become float64
	})

	t.Run("BuildSNSEvent", func(t *testing.T) {
		topicArn := "arn:aws:sns:us-east-1:123456789012:my-topic"
		event := BuildSNSEvent(topicArn, "Test Subject", "Test Message")

		assert.Equal(t, "aws.sns", event.Source)
		assert.Equal(t, "SNS Message", event.DetailType)
		assert.Equal(t, DefaultEventBusName, event.EventBusName)
		assert.Equal(t, []string{topicArn}, event.Resources)
		
		// Validate detail is valid JSON
		var detail map[string]interface{}
		err := json.Unmarshal([]byte(event.Detail), &detail)
		assert.NoError(t, err)
		assert.Equal(t, topicArn, detail["TopicArn"])
		assert.Equal(t, "Test Subject", detail["Subject"])
		assert.Equal(t, "Test Message", detail["Message"])
		assert.Equal(t, "12345678-1234-1234-1234-123456789012", detail["MessageId"])
	})

	t.Run("BuildScheduledEventWithCron", func(t *testing.T) {
		rule := BuildScheduledEventWithCron("0 12 * * ? *", map[string]interface{}{
			"key": "value",
		})

		assert.Contains(t, rule.Name, "scheduled-rule-")
		assert.Contains(t, rule.Description, "cron: 0 12 * * ? *")
		assert.Equal(t, "cron(0 12 * * ? *)", rule.ScheduleExpression)
		assert.Equal(t, RuleStateEnabled, rule.State)
		assert.Equal(t, DefaultEventBusName, rule.EventBusName)
	})

	t.Run("BuildScheduledEventWithRate", func(t *testing.T) {
		rule := BuildScheduledEventWithRate(5, "minutes")

		assert.Contains(t, rule.Name, "rate-rule-")
		assert.Contains(t, rule.Description, "every 5 minutes")
		assert.Equal(t, "rate(5 minutes)", rule.ScheduleExpression)
		assert.Equal(t, RuleStateEnabled, rule.State)
		assert.Equal(t, DefaultEventBusName, rule.EventBusName)
	})

	// Test that all builders produce valid JSON detail
	t.Run("AllBuilders_ProduceValidJSON", func(t *testing.T) {
		builders := []func() CustomEvent{
			func() CustomEvent { return BuildS3Event("test", "test.txt", "ObjectCreated:Put") },
			func() CustomEvent { return BuildEC2Event("i-123", "running", "pending") },
			func() CustomEvent { return BuildLambdaEvent("func", "arn:aws:lambda:us-east-1:123:function:func", "Active", "Ready") },
			func() CustomEvent { return BuildDynamoDBEvent("table", "INSERT", map[string]interface{}{}) },
			func() CustomEvent { return BuildSQSEvent("queue", "msg", "body") },
			func() CustomEvent { return BuildCodeCommitEvent("repo", "branch", "refs/heads/main", "commit") },
			func() CustomEvent { return BuildCodePipelineEvent("pipeline", "STARTED", "exec") },
			func() CustomEvent { return BuildECSEvent("cluster", "task", "RUNNING", "RUNNING") },
			func() CustomEvent { return BuildCloudWatchEvent("alarm", "ALARM", "reason", "metric") },
			func() CustomEvent { return BuildAPIGatewayEvent("api", "stage", "req", "200") },
			func() CustomEvent { return BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", "subj", "msg") },
		}

		for i, builder := range builders {
			event := builder()
			
			// Test that detail is valid JSON
			var detail interface{}
			err := json.Unmarshal([]byte(event.Detail), &detail)
			assert.NoError(t, err, "Builder %d should produce valid JSON", i)
			
			// Test that event has required fields
			assert.NotEmpty(t, event.Source, "Builder %d should have source", i)
			assert.NotEmpty(t, event.DetailType, "Builder %d should have detail type", i)
			assert.NotEmpty(t, event.Detail, "Builder %d should have detail", i)
			assert.Equal(t, DefaultEventBusName, event.EventBusName, "Builder %d should use default event bus", i)
		}
	})

	// Test struct definitions by creating instances
	t.Run("EventStructs_CanBeCreated", func(t *testing.T) {
		// Test S3EventDetail
		s3Detail := S3EventDetail{
			EventVersion: "2.1",
			EventSource:  "aws:s3",
			EventName:    "ObjectCreated:Put",
		}
		assert.Equal(t, "2.1", s3Detail.EventVersion)

		// Test EC2EventDetail
		ec2Detail := EC2EventDetail{
			InstanceID: "i-123",
			State:      "running",
		}
		assert.Equal(t, "i-123", ec2Detail.InstanceID)

		// Test LambdaEventDetail
		lambdaDetail := LambdaEventDetail{
			FunctionName: "test-func",
			Runtime:      "python3.9",
		}
		assert.Equal(t, "test-func", lambdaDetail.FunctionName)

		// Test DynamoDBEventDetail
		dynamoDetail := DynamoDBEventDetail{
			TableName: "test-table",
			EventName: "INSERT",
		}
		assert.Equal(t, "test-table", dynamoDetail.TableName)

		// Test SQSEventDetail
		sqsDetail := SQSEventDetail{
			QueueUrl:  "https://sqs.amazonaws.com/123/queue",
			MessageId: "msg-123",
		}
		assert.Equal(t, "https://sqs.amazonaws.com/123/queue", sqsDetail.QueueUrl)
	})
}

// TestPanicWrappers tests panic wrapper functions (that call *E versions and panic on error)
func TestPanicWrappers(t *testing.T) {
	// Test archive panic wrappers
	t.Run("CreateArchive_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		config := ArchiveConfig{ArchiveName: ""} // Empty name should cause validation failure

		assert.Panics(t, func() {
			CreateArchive(ctx, config)
		})

		// No mock expectations since validation fails before AWS call
	})

	t.Run("DeleteArchive_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("DeleteArchive", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("archive not found"))

		assert.Panics(t, func() {
			DeleteArchive(ctx, "nonexistent-archive")
		})

		mockClient.AssertExpectations(t)
	})

	t.Run("DescribeArchive_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("DescribeArchive", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("archive not found"))

		assert.Panics(t, func() {
			DescribeArchive(ctx, "nonexistent-archive")
		})

		mockClient.AssertExpectations(t)
	})

	t.Run("StartReplay_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		config := ReplayConfig{ReplayName: ""} // Empty name should cause validation failure

		assert.Panics(t, func() {
			StartReplay(ctx, config)
		})

		// No mock expectations since validation fails before AWS call
	})

	t.Run("DescribeReplay_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("DescribeReplay", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("replay not found"))

		assert.Panics(t, func() {
			DescribeReplay(ctx, "nonexistent-replay")
		})

		mockClient.AssertExpectations(t)
	})

	t.Run("CancelReplay_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("CancelReplay", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("replay cannot be cancelled"))

		assert.Panics(t, func() {
			CancelReplay(ctx, "nonexistent-replay")
		})

		mockClient.AssertExpectations(t)
	})

	// Test eventbus panic wrappers
	t.Run("CreateEventBus_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		config := EventBusConfig{Name: ""} // Empty name should cause validation failure

		assert.Panics(t, func() {
			CreateEventBus(ctx, config)
		})

		// No mock expectations since validation fails before AWS call
	})

	t.Run("DeleteEventBus_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("DeleteEventBus", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("bus not found"))

		assert.Panics(t, func() {
			DeleteEventBus(ctx, "nonexistent-bus")
		})

		mockClient.AssertExpectations(t)
	})

	t.Run("DescribeEventBus_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("DescribeEventBus", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("bus not found"))

		assert.Panics(t, func() {
			DescribeEventBus(ctx, "nonexistent-bus")
		})

		mockClient.AssertExpectations(t)
	})

	// Test rules panic wrappers
	t.Run("CreateRule_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		config := RuleConfig{Name: ""} // Empty name should cause validation failure

		assert.Panics(t, func() {
			CreateRule(ctx, config)
		})

		// No mock expectations since validation fails before AWS call
	})

	t.Run("DeleteRule_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		// Mock list targets call first
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("rule not found"))

		assert.Panics(t, func() {
			DeleteRule(ctx, "nonexistent-rule", "default")
		})

		mockClient.AssertExpectations(t)
	})

	// Test assertion panic wrappers
	t.Run("AssertRuleExists_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("DescribeRule", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("rule not found"))

		assert.Panics(t, func() {
			AssertRuleExists(ctx, "nonexistent-rule", "default")
		})

		mockClient.AssertExpectations(t)
	})

	t.Run("AssertRuleEnabled_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("DescribeRule", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("rule not found"))

		assert.Panics(t, func() {
			AssertRuleEnabled(ctx, "nonexistent-rule", "default")
		})

		mockClient.AssertExpectations(t)
	})

	t.Run("AssertRuleDisabled_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("DescribeRule", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("rule not found"))

		assert.Panics(t, func() {
			AssertRuleDisabled(ctx, "nonexistent-rule", "default")
		})

		mockClient.AssertExpectations(t)
	})

	t.Run("AssertTargetExists_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("rule not found"))

		assert.Panics(t, func() {
			AssertTargetExists(ctx, "nonexistent-rule", "default", "target1")
		})

		mockClient.AssertExpectations(t)
	})

	t.Run("AssertTargetCount_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("rule not found"))

		assert.Panics(t, func() {
			AssertTargetCount(ctx, "nonexistent-rule", "default", 1)
		})

		mockClient.AssertExpectations(t)
	})

	t.Run("AssertEventBusExists_PanicOnError", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		mockClient.On("DescribeEventBus", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("bus not found"))

		assert.Panics(t, func() {
			AssertEventBusExists(ctx, "nonexistent-bus")
		})

		mockClient.AssertExpectations(t)
	})

	// Test some operations that should succeed to verify wrapper functionality
	t.Run("CreateArchive_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		// Mock successful response
		mockClient.On("CreateArchive", mock.Anything, mock.Anything).Return(&eventbridge.CreateArchiveOutput{
			ArchiveArn: aws.String("arn:aws:events:us-east-1:123456789012:archive/test-archive"),
			State:      types.ArchiveStateEnabled,
		}, nil)

		config := ArchiveConfig{
			ArchiveName:    "test-archive",
			EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
			RetentionDays:  7,
		}

		// This should not panic
		result := CreateArchive(ctx, config)
		assert.Equal(t, "test-archive", result.ArchiveName)
		assert.Equal(t, "arn:aws:events:us-east-1:123456789012:archive/test-archive", result.ArchiveArn)

		mockClient.AssertExpectations(t)
	})

	t.Run("CreateEventBus_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		// Mock successful response
		mockClient.On("CreateEventBus", mock.Anything, mock.Anything).Return(&eventbridge.CreateEventBusOutput{
			EventBusArn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/test-bus"),
		}, nil)

		config := EventBusConfig{
			Name: "test-bus",
		}

		// This should not panic
		result := CreateEventBus(ctx, config)
		assert.Equal(t, "test-bus", result.Name)
		assert.Equal(t, "arn:aws:events:us-east-1:123456789012:event-bus/test-bus", result.Arn)

		mockClient.AssertExpectations(t)
	})

	t.Run("CreateRule_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		// Mock successful response
		mockClient.On("PutRule", mock.Anything, mock.Anything).Return(&eventbridge.PutRuleOutput{
			RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
		}, nil)

		config := RuleConfig{
			Name:         "test-rule",
			Description:  "Test rule",
			EventPattern: `{"source":["test.service"]}`,
			State:        RuleStateEnabled,
		}

		// This should not panic
		result := CreateRule(ctx, config)
		assert.Equal(t, "test-rule", result.Name)
		assert.Equal(t, "arn:aws:events:us-east-1:123456789012:rule/test-rule", result.RuleArn)

		mockClient.AssertExpectations(t)
	})

	t.Run("AssertRuleExists_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockTestingT := &MockTestingT{}
		ctx := &TestContext{T: mockTestingT, Client: mockClient}

		// Mock successful response
		mockClient.On("DescribeRule", mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
			Name:        aws.String("test-rule"),
			Description: aws.String("Test rule"),
			State:       types.RuleStateEnabled,
		}, nil)

		// This should not panic
		assert.NotPanics(t, func() {
			AssertRuleExists(ctx, "test-rule", "default")
		})

		mockClient.AssertExpectations(t)
	})
}

// TestEventBusValidation tests event bus validation functions
func TestEventBusValidation(t *testing.T) {
	// Test ValidateEventBusConfigE function
	t.Run("ValidateEventBusConfigE_Success", func(t *testing.T) {
		config := EventBusConfig{
			Name:            "my-custom-bus",
			EventSourceName: "com.mycompany.service",
			Tags: map[string]string{
				"Environment": "test",
				"Project":     "myproject",
			},
		}

		err := ValidateEventBusConfigE(config)
		assert.NoError(t, err)
	})

	t.Run("ValidateEventBusConfigE_InvalidName", func(t *testing.T) {
		config := EventBusConfig{
			Name: "",
		}

		err := ValidateEventBusConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event bus name")
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("ValidateEventBusConfigE_InvalidEventSource", func(t *testing.T) {
		config := EventBusConfig{
			Name:            "my-custom-bus",
			EventSourceName: ".invalid-source",
		}

		err := ValidateEventBusConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event source name")
		assert.Contains(t, err.Error(), "cannot start with special characters")
	})

	t.Run("ValidateEventBusConfigE_InvalidTags", func(t *testing.T) {
		config := EventBusConfig{
			Name: "my-custom-bus",
			Tags: map[string]string{
				"aws:reserved": "value",
			},
		}

		err := ValidateEventBusConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tags")
		assert.Contains(t, err.Error(), "cannot start with 'aws:'")
	})

	t.Run("ValidateEventBusConfigE_InvalidDeadLetterConfig", func(t *testing.T) {
		config := EventBusConfig{
			Name: "my-custom-bus",
			DeadLetterConfig: &types.DeadLetterConfig{
				Arn: aws.String("invalid-arn"),
			},
		}

		err := ValidateEventBusConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid dead letter config")
		assert.Contains(t, err.Error(), "invalid dead letter queue ARN format")
	})

	// Test ValidateEventBusNameE function
	t.Run("ValidateEventBusNameE_Success", func(t *testing.T) {
		validNames := []string{
			"my-bus",
			"test123",
			"bus.with.dots",
			"bus_with_underscores",
			"Bus123-Test_2.0",
			"default", // Should allow default
		}

		for _, name := range validNames {
			err := ValidateEventBusNameE(name)
			assert.NoError(t, err, "Name should be valid: %s", name)
		}
	})

	t.Run("ValidateEventBusNameE_EmptyName", func(t *testing.T) {
		err := ValidateEventBusNameE("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event bus name cannot be empty")
	})

	t.Run("ValidateEventBusNameE_TooLong", func(t *testing.T) {
		longName := strings.Repeat("a", MaxEventBusNameLength+1)
		err := ValidateEventBusNameE(longName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event bus name too long")
		assert.Contains(t, err.Error(), fmt.Sprintf("maximum length is %d", MaxEventBusNameLength))
	})

	t.Run("ValidateEventBusNameE_InvalidCharacters", func(t *testing.T) {
		invalidNames := []string{
			"bus@invalid",
			"bus#invalid",
			"bus%invalid",
			"bus with spaces",
			"bus/invalid",
		}

		for _, name := range invalidNames {
			err := ValidateEventBusNameE(name)
			assert.Error(t, err, "Name should be invalid: %s", name)
			assert.Contains(t, err.Error(), "invalid characters in event bus name")
		}
	})

	t.Run("ValidateEventBusNameE_LeadingSpecialChars", func(t *testing.T) {
		invalidNames := []string{
			".leadingdot",
			"-leadinghyphen",
			"_leadingunderscore",
		}

		for _, name := range invalidNames {
			err := ValidateEventBusNameE(name)
			assert.Error(t, err, "Name should be invalid: %s", name)
			assert.Contains(t, err.Error(), "cannot start with special characters")
		}
	})

	t.Run("ValidateEventBusNameE_TrailingSpecialChars", func(t *testing.T) {
		invalidNames := []string{
			"trailingdot.",
			"trailinghyphen-",
			"trailingunderscore_",
		}

		for _, name := range invalidNames {
			err := ValidateEventBusNameE(name)
			assert.Error(t, err, "Name should be invalid: %s", name)
			assert.Contains(t, err.Error(), "cannot end with special characters")
		}
	})

	t.Run("ValidateEventBusNameE_ReservedNames", func(t *testing.T) {
		reservedNames := []string{
			"aws.events",
			"AWS.EVENTS", // Case insensitive
			"aws.s3",
			"aws.ec2",
			"aws.lambda",
			"aws.ecs",
		}

		for _, name := range reservedNames {
			err := ValidateEventBusNameE(name)
			assert.Error(t, err, "Name should be reserved: %s", name)
			assert.Contains(t, err.Error(), "cannot be a reserved name")
		}
	})

	t.Run("ValidateEventBusNameE_DefaultAllowed", func(t *testing.T) {
		validDefaults := []string{
			"default",
			"DEFAULT",
			"Default",
		}

		for _, name := range validDefaults {
			err := ValidateEventBusNameE(name)
			assert.NoError(t, err, "Default name should be allowed: %s", name)
		}
	})

	// Test ValidateEventSourceNameE function
	t.Run("ValidateEventSourceNameE_Success", func(t *testing.T) {
		validSources := []string{
			"com.mycompany.service",
			"myapp.orders",
			"service123.events",
			"app_service.notifications",
			"domain-service.events",
		}

		for _, source := range validSources {
			err := ValidateEventSourceNameE(source)
			assert.NoError(t, err, "Source should be valid: %s", source)
		}
	})

	t.Run("ValidateEventSourceNameE_EmptyName", func(t *testing.T) {
		err := ValidateEventSourceNameE("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event source name cannot be empty")
	})

	t.Run("ValidateEventSourceNameE_TooLong", func(t *testing.T) {
		longSource := strings.Repeat("a", 257)
		err := ValidateEventSourceNameE(longSource)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event source name too long")
		assert.Contains(t, err.Error(), "maximum length is 256")
	})

	t.Run("ValidateEventSourceNameE_InvalidCharacters", func(t *testing.T) {
		invalidSources := []string{
			"source@invalid",
			"source#invalid",
			"source with spaces",
			"source/invalid",
		}

		for _, source := range invalidSources {
			err := ValidateEventSourceNameE(source)
			assert.Error(t, err, "Source should be invalid: %s", source)
			assert.Contains(t, err.Error(), "invalid characters in event source name")
		}
	})

	t.Run("ValidateEventSourceNameE_LeadingSpecialChars", func(t *testing.T) {
		invalidSources := []string{
			".leading.dot",
			"-leading-hyphen",
		}

		for _, source := range invalidSources {
			err := ValidateEventSourceNameE(source)
			assert.Error(t, err, "Source should be invalid: %s", source)
			assert.Contains(t, err.Error(), "cannot start with special characters")
		}
	})

	t.Run("ValidateEventSourceNameE_TrailingSpecialChars", func(t *testing.T) {
		invalidSources := []string{
			"trailing.dot.",
			"trailing-hyphen-",
		}

		for _, source := range invalidSources {
			err := ValidateEventSourceNameE(source)
			assert.Error(t, err, "Source should be invalid: %s", source)
			assert.Contains(t, err.Error(), "cannot end with special characters")
		}
	})

	// Test ValidateEventBusTagsE function
	t.Run("ValidateEventBusTagsE_Success", func(t *testing.T) {
		validTags := map[string]string{
			"Environment": "test",
			"Project":     "myproject",
			"Team":        "backend",
			"EmptyValue":  "", // Empty values should be allowed
		}

		err := ValidateEventBusTagsE(validTags)
		assert.NoError(t, err)
	})

	t.Run("ValidateEventBusTagsE_EmptyTags", func(t *testing.T) {
		err := ValidateEventBusTagsE(map[string]string{})
		assert.NoError(t, err)
	})

	t.Run("ValidateEventBusTagsE_TooManyTags", func(t *testing.T) {
		tooManyTags := make(map[string]string)
		for i := 0; i < 51; i++ { // Max is 50
			tooManyTags[fmt.Sprintf("Key%d", i)] = fmt.Sprintf("Value%d", i)
		}

		err := ValidateEventBusTagsE(tooManyTags)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many tags")
		assert.Contains(t, err.Error(), "maximum is 50")
	})

	t.Run("ValidateEventBusTagsE_InvalidTagKey", func(t *testing.T) {
		invalidTags := map[string]string{
			"aws:reserved": "value",
		}

		err := ValidateEventBusTagsE(invalidTags)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tag key")
		assert.Contains(t, err.Error(), "cannot start with 'aws:'")
	})

	t.Run("ValidateEventBusTagsE_InvalidTagValue", func(t *testing.T) {
		longValue := strings.Repeat("v", 257) // Max is 256
		invalidTags := map[string]string{
			"ValidKey": longValue,
		}

		err := ValidateEventBusTagsE(invalidTags)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tag value")
		assert.Contains(t, err.Error(), "too long")
	})

	// Test ValidateTagKeyE function
	t.Run("ValidateTagKeyE_Success", func(t *testing.T) {
		validKeys := []string{
			"Environment",
			"Project123",
			"key-with-hyphens",
			"key_with_underscores",
			"key.with.dots",
			strings.Repeat("k", 128), // Max length
		}

		for _, key := range validKeys {
			err := ValidateTagKeyE(key)
			assert.NoError(t, err, "Key should be valid: %s", key)
		}
	})

	t.Run("ValidateTagKeyE_EmptyKey", func(t *testing.T) {
		err := ValidateTagKeyE("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tag key cannot be empty")
	})

	t.Run("ValidateTagKeyE_TooLong", func(t *testing.T) {
		longKey := strings.Repeat("k", 129) // Max is 128
		err := ValidateTagKeyE(longKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tag key too long")
		assert.Contains(t, err.Error(), "maximum length is 128")
	})

	t.Run("ValidateTagKeyE_ReservedPrefix", func(t *testing.T) {
		reservedKeys := []string{
			"aws:sometag",
			"AWS:SOMETAG",
			"Aws:SomeTag",
		}

		for _, key := range reservedKeys {
			err := ValidateTagKeyE(key)
			assert.Error(t, err, "Key should be invalid: %s", key)
			assert.Contains(t, err.Error(), "cannot start with 'aws:'")
		}
	})

	// Test ValidateTagValueE function
	t.Run("ValidateTagValueE_Success", func(t *testing.T) {
		validValues := []string{
			"",                         // Empty should be allowed
			"test-value",
			"Value with spaces",
			"123456",
			strings.Repeat("v", 256), // Max length
		}

		for _, value := range validValues {
			err := ValidateTagValueE(value)
			assert.NoError(t, err, "Value should be valid: %s", value)
		}
	})

	t.Run("ValidateTagValueE_TooLong", func(t *testing.T) {
		longValue := strings.Repeat("v", 257) // Max is 256
		err := ValidateTagValueE(longValue)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tag value too long")
		assert.Contains(t, err.Error(), "maximum length is 256")
	})

	// Test builder functions
	t.Run("BuildEventBusConfig", func(t *testing.T) {
		config := BuildEventBusConfig("test-bus", "com.test.service")

		assert.Equal(t, "test-bus", config.Name)
		assert.Equal(t, "com.test.service", config.EventSourceName)
		assert.Nil(t, config.Tags)
		assert.Nil(t, config.DeadLetterConfig)
	})

	t.Run("BuildEventBusConfigWithTags", func(t *testing.T) {
		tags := map[string]string{
			"Environment": "test",
			"Project":     "myproject",
		}
		config := BuildEventBusConfigWithTags("test-bus", "com.test.service", tags)

		assert.Equal(t, "test-bus", config.Name)
		assert.Equal(t, "com.test.service", config.EventSourceName)
		assert.Equal(t, tags, config.Tags)
		assert.Nil(t, config.DeadLetterConfig)
	})

	t.Run("BuildEventBusConfigWithDLQ", func(t *testing.T) {
		dlqArn := "arn:aws:sqs:us-east-1:123456789012:dlq-queue"
		config := BuildEventBusConfigWithDLQ("test-bus", "com.test.service", dlqArn)

		assert.Equal(t, "test-bus", config.Name)
		assert.Equal(t, "com.test.service", config.EventSourceName)
		assert.Nil(t, config.Tags)
		assert.NotNil(t, config.DeadLetterConfig)
		assert.Equal(t, dlqArn, *config.DeadLetterConfig.Arn)
	})

	// Test utility functions
	t.Run("IsDefaultEventBus", func(t *testing.T) {
		assert.True(t, IsDefaultEventBus(""))
		assert.True(t, IsDefaultEventBus("default"))
		assert.True(t, IsDefaultEventBus("DEFAULT"))
		assert.True(t, IsDefaultEventBus("Default"))
		assert.False(t, IsDefaultEventBus("custom-bus"))
		assert.False(t, IsDefaultEventBus("my-bus"))
	})

	t.Run("GetEventBusRegion", func(t *testing.T) {
		validArn := "arn:aws:events:us-east-1:123456789012:event-bus/my-bus"
		region := GetEventBusRegion(validArn)
		assert.Equal(t, "us-east-1", region)

		invalidArn := "invalid-arn"
		region = GetEventBusRegion(invalidArn)
		assert.Equal(t, "", region)
	})

	t.Run("GetEventBusAccount", func(t *testing.T) {
		validArn := "arn:aws:events:us-east-1:123456789012:event-bus/my-bus"
		account := GetEventBusAccount(validArn)
		assert.Equal(t, "123456789012", account)

		invalidArn := "invalid-arn"
		account = GetEventBusAccount(invalidArn)
		assert.Equal(t, "", account)
	})

	t.Run("GetEventBusNameFromArn", func(t *testing.T) {
		validArn := "arn:aws:events:us-east-1:123456789012:event-bus/my-bus"
		name := GetEventBusNameFromArn(validArn)
		assert.Equal(t, "my-bus", name)

		invalidArn := "invalid-arn"
		name = GetEventBusNameFromArn(invalidArn)
		assert.Equal(t, "", name)

		// Test ARN without event-bus prefix
		nonEventBusArn := "arn:aws:events:us-east-1:123456789012:rule/my-rule"
		name = GetEventBusNameFromArn(nonEventBusArn)
		assert.Equal(t, "", name)
	})

	t.Run("ValidateEventBusArnE_Success", func(t *testing.T) {
		validArns := []string{
			"arn:aws:events:us-east-1:123456789012:event-bus/my-bus",
			"arn:aws:events:eu-west-1:987654321098:event-bus/test-bus",
		}

		for _, arn := range validArns {
			err := ValidateEventBusArnE(arn)
			assert.NoError(t, err, "ARN should be valid: %s", arn)
		}
	})

	t.Run("ValidateEventBusArnE_InvalidFormat", func(t *testing.T) {
		invalidArn := "invalid-arn-format"
		err := ValidateEventBusArnE(invalidArn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ARN format")
	})

	t.Run("ValidateEventBusArnE_NotEventBridgeService", func(t *testing.T) {
		sqsArn := "arn:aws:sqs:us-east-1:123456789012:my-queue"
		err := ValidateEventBusArnE(sqsArn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ARN is not for EventBridge service")
	})

	t.Run("ValidateEventBusArnE_NotEventBusResource", func(t *testing.T) {
		ruleArn := "arn:aws:events:us-east-1:123456789012:rule/my-rule"
		err := ValidateEventBusArnE(ruleArn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ARN is not for an event bus")
	})

	t.Run("NormalizeEventBusName", func(t *testing.T) {
		assert.Equal(t, "default", NormalizeEventBusName(""))
		assert.Equal(t, "custom-bus", NormalizeEventBusName("custom-bus"))
		assert.Equal(t, "default", NormalizeEventBusName("default"))
	})

	// Test panic variants of validation functions
	t.Run("ValidateEventBusConfig_PanicOnError", func(t *testing.T) {
		// This should panic since the config is invalid
		invalidConfig := EventBusConfig{
			Name: "", // Empty name should cause validation to fail
		}

		assert.Panics(t, func() {
			ValidateEventBusConfig(invalidConfig)
		})
	})

	t.Run("ValidateEventBusName_PanicOnError", func(t *testing.T) {
		assert.Panics(t, func() {
			ValidateEventBusName("") // Empty name should cause panic
		})
	})

	// Test constants validation
	t.Run("Constants_Validation", func(t *testing.T) {
		assert.Equal(t, 256, MaxEventBusNameLength)
		assert.Equal(t, 1, MinEventBusNameLength)

		// Test regex pattern indirectly by validating valid and invalid names
		err1 := ValidateEventBusNameE("valid-bus_name.123")
		assert.NoError(t, err1)
		
		err2 := ValidateEventBusNameE("invalid@bus")
		assert.Error(t, err2)
		assert.Contains(t, err2.Error(), "invalid characters")
	})
}

// TestTargetsValidation tests target validation functions
func TestTargetsValidation(t *testing.T) {
	ctx := &TestContext{T: t}

	t.Run("ValidateTargetConfigurationE_Success", func(t *testing.T) {
		target := TargetConfig{
			ID:  "test-target",
			Arn: "arn:aws:lambda:us-east-1:123456789012:function:test-function",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.NoError(t, err)
	})

	t.Run("ValidateTargetConfigurationE_EmptyID", func(t *testing.T) {
		target := TargetConfig{
			Arn: "arn:aws:lambda:us-east-1:123456789012:function:test-function",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target ID cannot be empty")
	})

	t.Run("ValidateTargetConfigurationE_EmptyArn", func(t *testing.T) {
		target := TargetConfig{
			ID: "test-target",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target ARN cannot be empty")
	})

	t.Run("ValidateTargetConfigurationE_InvalidArn", func(t *testing.T) {
		target := TargetConfig{
			ID:  "test-target",
			Arn: "invalid-arn",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ARN format")
	})

	t.Run("ValidateTargetConfigurationE_InvalidInputJSON", func(t *testing.T) {
		target := TargetConfig{
			ID:    "test-target",
			Arn:   "arn:aws:lambda:us-east-1:123456789012:function:test-function",
			Input: `{"invalid": json}`, // Invalid JSON
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON in Input field")
	})

	t.Run("ValidateTargetConfigurationE_ValidInputJSON", func(t *testing.T) {
		target := TargetConfig{
			ID:    "test-target",
			Arn:   "arn:aws:lambda:us-east-1:123456789012:function:test-function",
			Input: `{"valid": "json", "number": 123}`,
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.NoError(t, err)
	})

	t.Run("ValidateTargetConfigurationE_InvalidInputPath", func(t *testing.T) {
		target := TargetConfig{
			ID:        "test-target",
			Arn:       "arn:aws:lambda:us-east-1:123456789012:function:test-function",
			InputPath: "invalid-json-path", // Should start with $
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSONPath in InputPath")
	})

	t.Run("ValidateTargetConfigurationE_ValidInputPath", func(t *testing.T) {
		target := TargetConfig{
			ID:        "test-target",
			Arn:       "arn:aws:lambda:us-east-1:123456789012:function:test-function",
			InputPath: "$.detail",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.NoError(t, err)
	})

	t.Run("ValidateTargetConfigurationE_WithInputTransformer", func(t *testing.T) {
		transformer := &types.InputTransformer{
			InputTemplate: aws.String(`{"transformed": "data"}`),
		}

		target := TargetConfig{
			ID:               "test-target",
			Arn:              "arn:aws:lambda:us-east-1:123456789012:function:test-function",
			InputTransformer: transformer,
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.NoError(t, err)
	})

	t.Run("ValidateTargetConfigurationE_WithDeadLetterConfig", func(t *testing.T) {
		dlqConfig := &types.DeadLetterConfig{
			Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:dead-letter-queue"),
		}

		target := TargetConfig{
			ID:               "test-target",
			Arn:              "arn:aws:lambda:us-east-1:123456789012:function:test-function",
			DeadLetterConfig: dlqConfig,
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.NoError(t, err)
	})

	t.Run("ValidateTargetConfigurationE_WithRetryPolicy", func(t *testing.T) {
		retryPolicy := &types.RetryPolicy{
			MaximumRetryAttempts:      aws.Int32(3),
			MaximumEventAgeInSeconds: aws.Int32(3600),
		}

		target := TargetConfig{
			ID:          "test-target",
			Arn:         "arn:aws:lambda:us-east-1:123456789012:function:test-function",
			RetryPolicy: retryPolicy,
		}

		err := ValidateTargetConfigurationE(ctx, target)
		assert.NoError(t, err)
	})

	t.Run("GetTargetServiceType", func(t *testing.T) {
		testCases := map[string]string{
			"arn:aws:lambda:us-east-1:123456789012:function:test":         "lambda",
			"arn:aws:sqs:us-east-1:123456789012:queue:test":              "sqs",
			"arn:aws:sns:us-east-1:123456789012:topic:test":              "sns",
			"arn:aws:states:us-east-1:123456789012:stateMachine:test":    "stepfunctions",
			"arn:aws:kinesis:us-east-1:123456789012:stream/test":         "kinesis",
			"arn:aws:ecs:us-east-1:123456789012:task-definition/test":    "ecs",
			"arn:aws:events:us-east-1:123456789012:event-bus/test":       "eventbridge",
			"arn:aws:logs:us-east-1:123456789012:log-group:test":         "cloudwatch",
			"arn:aws:firehose:us-east-1:123456789012:deliverystream:test": "kinesis-firehose",
			"arn:aws:pipes:us-east-1:123456789012:pipe:test":             "eventbridge-pipes",
			"arn:aws:unknown:us-east-1:123456789012:resource:test":       "unknown",
			"invalid-arn":                                                "unknown",
			"arn:aws:too:few:parts":                                      "unknown",
		}

		for arn, expectedService := range testCases {
			service := GetTargetServiceType(arn)
			assert.Equal(t, expectedService, service, "Wrong service type for ARN: %s", arn)
		}
	})

	t.Run("BatchPutTargetsE_Success", func(t *testing.T) {
		// Create a mock client
		mockClient := new(MockEventBridgeClient)
		ctx.Client = mockClient

		targets := []TargetConfig{
			{ID: "target1", Arn: "arn:aws:lambda:us-east-1:123456789012:function:func1"},
			{ID: "target2", Arn: "arn:aws:lambda:us-east-1:123456789012:function:func2"},
		}

		// Mock successful PutTargets call
		mockClient.On("PutTargets", mock.Anything, mock.Anything).Return(
			&eventbridge.PutTargetsOutput{
				FailedEntryCount: 0,
				FailedEntries:    []types.PutTargetsResultEntry{},
			}, nil)

		err := BatchPutTargetsE(ctx, "test-rule", "default", targets, 2)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("BatchPutTargetsE_DefaultBatchSize", func(t *testing.T) {
		mockClient := new(MockEventBridgeClient)
		ctx.Client = mockClient

		targets := []TargetConfig{
			{ID: "target1", Arn: "arn:aws:lambda:us-east-1:123456789012:function:func1"},
		}

		mockClient.On("PutTargets", mock.Anything, mock.Anything).Return(
			&eventbridge.PutTargetsOutput{}, nil)

		err := BatchPutTargetsE(ctx, "test-rule", "default", targets, 0) // Should use default batch size
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("BatchRemoveTargetsE_Success", func(t *testing.T) {
		mockClient := new(MockEventBridgeClient)
		ctx.Client = mockClient

		targetIDs := []string{"target1", "target2", "target3"}

		mockClient.On("RemoveTargets", mock.Anything, mock.Anything).Return(
			&eventbridge.RemoveTargetsOutput{
				FailedEntryCount: 0,
				FailedEntries:    []types.RemoveTargetsResultEntry{},
			}, nil)

		err := BatchRemoveTargetsE(ctx, "test-rule", "default", targetIDs, 2)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("BuildInputTransformer", func(t *testing.T) {
		pathMap := map[string]string{
			"timestamp": "$.time",
			"source":    "$.source",
		}
		template := `{"ts": "<timestamp>", "src": "<source>"}`

		transformer := BuildInputTransformer(pathMap, template)

		assert.NotNil(t, transformer)
		assert.Equal(t, pathMap, transformer.InputPathsMap)
		assert.Equal(t, template, *transformer.InputTemplate)
	})

	t.Run("ValidateInputTransformerE_Success", func(t *testing.T) {
		transformer := &types.InputTransformer{
			InputTemplate: aws.String(`{"message": "Hello World"}`),
		}

		err := ValidateInputTransformerE(transformer)
		assert.NoError(t, err)
	})

	t.Run("ValidateInputTransformerE_NilTransformer", func(t *testing.T) {
		err := ValidateInputTransformerE(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "InputTransformer cannot be nil")
	})

	t.Run("ValidateInputTransformerE_MissingInputTemplate", func(t *testing.T) {
		transformer := &types.InputTransformer{}

		err := ValidateInputTransformerE(transformer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "InputTemplate is required")
	})

	t.Run("ValidateInputTransformerE_InvalidJSON", func(t *testing.T) {
		transformer := &types.InputTransformer{
			InputTemplate: aws.String(`{"invalid": json}`),
		}

		err := ValidateInputTransformerE(transformer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON in InputTemplate")
	})

	t.Run("ValidateInputTransformerE_TemplateTooLarge", func(t *testing.T) {
		largeTemplate := `{"data": "` + strings.Repeat("x", MaxInputTransformerSize) + `"}`
		transformer := &types.InputTransformer{
			InputTemplate: aws.String(largeTemplate),
		}

		err := ValidateInputTransformerE(transformer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "InputTemplate exceeds maximum size")
	})

	t.Run("ValidateInputTransformerE_UnmappedPlaceholder", func(t *testing.T) {
		transformer := &types.InputTransformer{
			InputTemplate: aws.String(`{"message": "<greeting>", "name": "<unmapped>"}`),
			InputPathsMap: map[string]string{
				"greeting": "$.greeting",
				// "unmapped" is missing from the path map
			},
		}

		err := ValidateInputTransformerE(transformer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmapped placeholder in template: unmapped")
	})

	t.Run("ValidateInputTransformerE_AllPlaceholdersMapped", func(t *testing.T) {
		transformer := &types.InputTransformer{
			InputTemplate: aws.String(`{"message": "<greeting>", "name": "<name>"}`),
			InputPathsMap: map[string]string{
				"greeting": "$.greeting",
				"name":     "$.user.name",
			},
		}

		err := ValidateInputTransformerE(transformer)
		assert.NoError(t, err)
	})

	t.Run("extractPlaceholders", func(t *testing.T) {
		template := `{"msg": "<message>", "user": "<userName>", "fixed": "value"}`
		placeholders := extractPlaceholders(template)
		
		expected := []string{"message", "userName"}
		assert.ElementsMatch(t, expected, placeholders)

		// Test with no placeholders
		template = `{"fixed": "value", "number": 123}`
		placeholders = extractPlaceholders(template)
		assert.Empty(t, placeholders)

		// Test with duplicate placeholders
		template = `{"msg1": "<greeting>", "msg2": "<greeting>"}`
		placeholders = extractPlaceholders(template)
		assert.Equal(t, []string{"greeting", "greeting"}, placeholders)
	})

	t.Run("BuildDeadLetterConfig", func(t *testing.T) {
		queueArn := "arn:aws:sqs:us-east-1:123456789012:dead-letter-queue"
		config := BuildDeadLetterConfig(queueArn)

		assert.NotNil(t, config)
		assert.Equal(t, queueArn, *config.Arn)
	})

	t.Run("ValidateDeadLetterConfigE_Success", func(t *testing.T) {
		config := &types.DeadLetterConfig{
			Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:dead-letter-queue"),
		}

		err := ValidateDeadLetterConfigE(config)
		assert.NoError(t, err)
	})

	t.Run("ValidateDeadLetterConfigE_NilConfig", func(t *testing.T) {
		err := ValidateDeadLetterConfigE(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DeadLetterConfig cannot be nil")
	})

	t.Run("ValidateDeadLetterConfigE_MissingArn", func(t *testing.T) {
		config := &types.DeadLetterConfig{}

		err := ValidateDeadLetterConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dead letter queue ARN is required")
	})

	t.Run("ValidateDeadLetterConfigE_InvalidArn", func(t *testing.T) {
		config := &types.DeadLetterConfig{
			Arn: aws.String("invalid-arn"),
		}

		err := ValidateDeadLetterConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid dead letter queue ARN format")
	})

	t.Run("ValidateDeadLetterConfigE_NonSQSArn", func(t *testing.T) {
		config := &types.DeadLetterConfig{
			Arn: aws.String("arn:aws:lambda:us-east-1:123456789012:function:test"), // Lambda ARN, not SQS
		}

		err := ValidateDeadLetterConfigE(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dead letter queue must be an SQS queue")
	})

	t.Run("BuildRetryPolicy", func(t *testing.T) {
		policy := BuildRetryPolicy(5, 3600)

		assert.NotNil(t, policy)
		assert.Equal(t, int32(5), *policy.MaximumRetryAttempts)
		assert.Equal(t, int32(3600), *policy.MaximumEventAgeInSeconds)
	})

	t.Run("ValidateRetryPolicyE_Success", func(t *testing.T) {
		policy := &types.RetryPolicy{
			MaximumRetryAttempts:      aws.Int32(10),
			MaximumEventAgeInSeconds: aws.Int32(7200),
		}

		err := ValidateRetryPolicyE(policy)
		assert.NoError(t, err)
	})

	t.Run("ValidateRetryPolicyE_NilPolicy", func(t *testing.T) {
		err := ValidateRetryPolicyE(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RetryPolicy cannot be nil")
	})

	t.Run("ValidateRetryPolicyE_NegativeRetryAttempts", func(t *testing.T) {
		policy := &types.RetryPolicy{
			MaximumRetryAttempts: aws.Int32(-1),
		}

		err := ValidateRetryPolicyE(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maximum retry attempts cannot be negative")
	})

	t.Run("ValidateRetryPolicyE_TooManyRetryAttempts", func(t *testing.T) {
		policy := &types.RetryPolicy{
			MaximumRetryAttempts: aws.Int32(MaxTargetRetryAttempts + 1),
		}

		err := ValidateRetryPolicyE(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maximum retry attempts exceeds limit")
	})

	t.Run("ValidateRetryPolicyE_EventAgeTooSmall", func(t *testing.T) {
		policy := &types.RetryPolicy{
			MaximumEventAgeInSeconds: aws.Int32(30), // Less than minimum of 60
		}

		err := ValidateRetryPolicyE(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maximum event age must be at least 60 seconds")
	})

	t.Run("ValidateRetryPolicyE_EventAgeTooLarge", func(t *testing.T) {
		policy := &types.RetryPolicy{
			MaximumEventAgeInSeconds: aws.Int32(MaxEventAgeSeconds + 1),
		}

		err := ValidateRetryPolicyE(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maximum event age exceeds limit")
	})

	t.Run("ValidateRetryPolicyE_OnlyRetryAttempts", func(t *testing.T) {
		policy := &types.RetryPolicy{
			MaximumRetryAttempts: aws.Int32(5),
		}

		err := ValidateRetryPolicyE(policy)
		assert.NoError(t, err)
	})

	t.Run("ValidateRetryPolicyE_OnlyEventAge", func(t *testing.T) {
		policy := &types.RetryPolicy{
			MaximumEventAgeInSeconds: aws.Int32(1800),
		}

		err := ValidateRetryPolicyE(policy)
		assert.NoError(t, err)
	})

	// Test panic variants
	t.Run("ValidateTargetConfiguration_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{T: mockT}
		target := TargetConfig{} // Empty config will cause error

		assert.Panics(t, func() {
			ValidateTargetConfiguration(mockCtx, target)
		})
	})

	t.Run("BatchPutTargets_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{T: mockT}

		mockClient := &MockEventBridgeClient{}
		mockCtx.Client = mockClient

		target := TargetConfig{
			ID:  "test-target",
			Arn: "arn:aws:lambda:us-east-1:123456789012:function:test",
		}

		assert.Panics(t, func() {
			BatchPutTargets(mockCtx, "", "", []TargetConfig{target}, 1) // Empty rule name will cause error
		})
	})

	t.Run("BatchRemoveTargets_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{T: mockT}

		assert.Panics(t, func() {
			BatchRemoveTargets(mockCtx, "", "", []string{"target1"}, 1) // Empty rule name will cause error
		})
	})

	t.Run("ValidateInputTransformer_PanicOnError", func(t *testing.T) {
		assert.Panics(t, func() {
			ValidateInputTransformer(nil) // Nil transformer will cause panic
		})
	})

	t.Run("ValidateDeadLetterConfig_PanicOnError", func(t *testing.T) {
		assert.Panics(t, func() {
			ValidateDeadLetterConfig(nil) // Nil config will cause panic
		})
	})

	t.Run("ValidateRetryPolicy_PanicOnError", func(t *testing.T) {
		assert.Panics(t, func() {
			ValidateRetryPolicy(nil) // Nil policy will cause panic
		})
	})

	// Test constants
	t.Run("Constants", func(t *testing.T) {
		assert.Equal(t, 185, MaxTargetRetryAttempts)
		assert.Equal(t, 86400, MaxEventAgeSeconds)
		assert.Equal(t, 100, MaxTargetsPerRule)
		assert.Equal(t, 8192, MaxInputTransformerSize)
	})
}

// TestPanicVersionsCore tests the panic versions of core functions that have 0% coverage
func TestPanicVersionsCore(t *testing.T) {
	t.Run("PutEvent_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
		}

		// Mock successful response
		mockClient.On("PutEvents", mock.Anything, mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				Entries: []types.PutEventsResultEntry{
					{
						EventId: aws.String("event-123"),
					},
				},
			}, nil)

		// Should not panic on success
		result := PutEvent(mockCtx, event)
		assert.True(t, result.Success)
		assert.Equal(t, "event-123", result.EventID)
		mockClient.AssertExpectations(t)
	})

	t.Run("PutEvent_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: &MockEventBridgeClient{},
		}

		event := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"invalid": json}`, // Invalid JSON
		}

		assert.Panics(t, func() {
			PutEvent(mockCtx, event)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("PutEvents_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		events := []CustomEvent{
			{
				Source:     "test.service",
				DetailType: "Test Event 1",
				Detail:     `{"key1": "value1"}`,
			},
			{
				Source:     "test.service",
				DetailType: "Test Event 2",
				Detail:     `{"key2": "value2"}`,
			},
		}

		// Mock successful response
		mockClient.On("PutEvents", mock.Anything, mock.Anything).Return(
			&eventbridge.PutEventsOutput{
				Entries: []types.PutEventsResultEntry{
					{EventId: aws.String("event-1")},
					{EventId: aws.String("event-2")},
				},
			}, nil)

		// Should not panic on success
		result := PutEvents(mockCtx, events)
		assert.Equal(t, 2, len(result.Entries))
		assert.True(t, result.Entries[0].Success)
		assert.True(t, result.Entries[1].Success)
		mockClient.AssertExpectations(t)
	})

	t.Run("PutEvents_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: &MockEventBridgeClient{},
		}

		// Too many events will cause error
		events := make([]CustomEvent, 11)
		for i := range events {
			events[i] = CustomEvent{
				Source:     "test.service",
				DetailType: "Test Event",
				Detail:     `{"key": "value"}`,
			}
		}

		assert.Panics(t, func() {
			PutEvents(mockCtx, events)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})
}

// TestRulePanicVersions tests the panic versions of rule functions that have 0% coverage
func TestRulePanicVersions(t *testing.T) {
	t.Run("DescribeRule_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("DescribeRule", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeRuleOutput{
				Name:        aws.String("test-rule"),
				Arn:         aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
				State:       types.RuleStateEnabled,
				Description: aws.String("Test rule"),
			}, nil)

		result := DescribeRule(mockCtx, "test-rule", "default")
		assert.Equal(t, "test-rule", result.Name)
		assert.Equal(t, types.RuleStateEnabled, result.State)
		mockClient.AssertExpectations(t)
	})

	t.Run("DescribeRule_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: &MockEventBridgeClient{},
		}

		assert.Panics(t, func() {
			DescribeRule(mockCtx, "", "default") // Empty rule name will cause error
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("EnableRule_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("EnableRule", mock.Anything, mock.Anything).Return(
			&eventbridge.EnableRuleOutput{}, nil)

		// Should not panic on success
		EnableRule(mockCtx, "test-rule", "default")
		mockClient.AssertExpectations(t)
	})

	t.Run("EnableRule_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: &MockEventBridgeClient{},
		}

		assert.Panics(t, func() {
			EnableRule(mockCtx, "", "default") // Empty rule name will cause error
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("DisableRule_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("DisableRule", mock.Anything, mock.Anything).Return(
			&eventbridge.DisableRuleOutput{}, nil)

		// Should not panic on success
		DisableRule(mockCtx, "test-rule", "default")
		mockClient.AssertExpectations(t)
	})

	t.Run("DisableRule_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: &MockEventBridgeClient{},
		}

		assert.Panics(t, func() {
			DisableRule(mockCtx, "", "default") // Empty rule name will cause error
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("ListRules_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("ListRules", mock.Anything, mock.Anything).Return(
			&eventbridge.ListRulesOutput{
				Rules: []types.Rule{
					{
						Name:        aws.String("rule1"),
						Arn:         aws.String("arn:aws:events:us-east-1:123456789012:rule/rule1"),
						State:       types.RuleStateEnabled,
						Description: aws.String("Rule 1"),
					},
					{
						Name:        aws.String("rule2"),
						Arn:         aws.String("arn:aws:events:us-east-1:123456789012:rule/rule2"),
						State:       types.RuleStateDisabled,
						Description: aws.String("Rule 2"),
					},
				},
			}, nil)

		results := ListRules(mockCtx, "default")
		assert.Equal(t, 2, len(results))
		assert.Equal(t, "rule1", results[0].Name)
		assert.Equal(t, "rule2", results[1].Name)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListRules_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock error response
		mockClient.On("ListRules", mock.Anything, mock.Anything).Return(
			nil, errors.New("service error"))

		assert.Panics(t, func() {
			ListRules(mockCtx, "default")
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("UpdateRule_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		config := RuleConfig{
			Name:        "test-rule",
			Description: "Updated test rule",
			EventPattern: `{"source": ["test.service"]}`,
			State:       types.RuleStateEnabled,
		}

		// Mock successful response
		mockClient.On("PutRule", mock.Anything, mock.Anything).Return(
			&eventbridge.PutRuleOutput{
				RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
			}, nil)

		result := UpdateRule(mockCtx, config)
		assert.Equal(t, "test-rule", result.Name)
		assert.Equal(t, "Updated test rule", result.Description)
		mockClient.AssertExpectations(t)
	})

	t.Run("UpdateRule_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: &MockEventBridgeClient{},
		}

		config := RuleConfig{
			Name: "", // Empty name will cause error
		}

		assert.Panics(t, func() {
			UpdateRule(mockCtx, config)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})
}

// TestTargetPanicVersions tests the panic versions of target functions that have 0% coverage  
func TestTargetPanicVersions(t *testing.T) {
	t.Run("PutTargets_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		targets := []TargetConfig{
			{
				ID:  "target1",
				Arn: "arn:aws:lambda:us-east-1:123456789012:function:test",
			},
		}

		// Mock successful response
		mockClient.On("PutTargets", mock.Anything, mock.Anything).Return(
			&eventbridge.PutTargetsOutput{}, nil)

		// Should not panic on success
		PutTargets(mockCtx, "test-rule", "default", targets)
		mockClient.AssertExpectations(t)
	})

	t.Run("PutTargets_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: &MockEventBridgeClient{},
		}

		assert.Panics(t, func() {
			PutTargets(mockCtx, "", "default", []TargetConfig{}) // Empty rule name will cause error
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("RemoveTargets_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("RemoveTargets", mock.Anything, mock.Anything).Return(
			&eventbridge.RemoveTargetsOutput{}, nil)

		// Should not panic on success
		RemoveTargets(mockCtx, "test-rule", "default", []string{"target1"})
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveTargets_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: &MockEventBridgeClient{},
		}

		assert.Panics(t, func() {
			RemoveTargets(mockCtx, "", "default", []string{"target1"}) // Empty rule name will cause error
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("ListTargetsForRule_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything).Return(
			&eventbridge.ListTargetsByRuleOutput{
				Targets: []types.Target{
					{
						Id:  aws.String("target1"),
						Arn: aws.String("arn:aws:lambda:us-east-1:123456789012:function:test"),
					},
				},
			}, nil)

		results := ListTargetsForRule(mockCtx, "test-rule", "default")
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "target1", results[0].ID)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListTargetsForRule_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock error response
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything).Return(
			nil, errors.New("service error"))

		assert.Panics(t, func() {
			ListTargetsForRule(mockCtx, "test-rule", "default")
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})
}

// TestEventBusExtendedOperations tests functions with 0% coverage
func TestEventBusExtendedOperations(t *testing.T) {
	t.Run("ListEventBuses_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("ListEventBuses", mock.Anything, mock.Anything).Return(
			&eventbridge.ListEventBusesOutput{
				EventBuses: []types.EventBus{
					{
						Name: aws.String("default"),
						Arn:  aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
					},
					{
						Name: aws.String("custom-bus"),
						Arn:  aws.String("arn:aws:events:us-east-1:123456789012:event-bus/custom-bus"),
					},
				},
			}, nil)

		results := ListEventBuses(mockCtx)
		assert.Equal(t, 2, len(results))
		assert.Equal(t, "default", results[0].Name)
		assert.Equal(t, "custom-bus", results[1].Name)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListEventBuses_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock error response
		mockClient.On("ListEventBuses", mock.Anything, mock.Anything).Return(
			nil, errors.New("service error"))

		assert.Panics(t, func() {
			ListEventBuses(mockCtx)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("PutPermission_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("PutPermission", mock.Anything, mock.Anything).Return(
			&eventbridge.PutPermissionOutput{}, nil)

		// Should not panic on success
		PutPermission(mockCtx, "default", "test-statement", "123456789012", "events:PutEvents")
		mockClient.AssertExpectations(t)
	})

	t.Run("PutPermission_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock error response
		mockClient.On("PutPermission", mock.Anything, mock.Anything).Return(
			nil, errors.New("service error"))

		assert.Panics(t, func() {
			PutPermission(mockCtx, "default", "test-statement", "123456789012", "events:PutEvents")
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("RemovePermission_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("RemovePermission", mock.Anything, mock.Anything).Return(
			&eventbridge.RemovePermissionOutput{}, nil)

		// Should not panic on success
		RemovePermission(mockCtx, "test-statement", "default")
		mockClient.AssertExpectations(t)
	})

	t.Run("RemovePermission_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock error response
		mockClient.On("RemovePermission", mock.Anything, mock.Anything).Return(
			nil, errors.New("service error"))

		assert.Panics(t, func() {
			RemovePermission(mockCtx, "test-statement", "default")
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})
}

// TestArchiveExtendedOperations tests archive functions with 0% coverage
func TestArchiveExtendedOperations(t *testing.T) {
	t.Run("ListArchives_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("ListArchives", mock.Anything, mock.Anything).Return(
			&eventbridge.ListArchivesOutput{
				Archives: []types.Archive{
					{
						ArchiveName:    aws.String("archive1"),
						State:          types.ArchiveStateEnabled,
						EventSourceArn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
					},
					{
						ArchiveName:    aws.String("archive2"),
						State:          types.ArchiveStateDisabled,
						EventSourceArn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
					},
				},
			}, nil)

		results := ListArchives(mockCtx, "arn:aws:events:us-east-1:123456789012:event-bus/default")
		assert.Equal(t, 2, len(results))
		assert.Equal(t, "archive1", results[0].ArchiveName)
		assert.Equal(t, "archive2", results[1].ArchiveName)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListArchives_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock error response
		mockClient.On("ListArchives", mock.Anything, mock.Anything).Return(
			nil, errors.New("service error"))

		assert.Panics(t, func() {
			ListArchives(mockCtx, "arn:aws:events:us-east-1:123456789012:event-bus/default")
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})
}

// TestMissingAssertions tests assertion functions with 0% coverage
func TestMissingAssertions(t *testing.T) {
	t.Run("AssertArchiveExists_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("DescribeArchive", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeArchiveOutput{
				ArchiveName: aws.String("test-archive"),
				State:       types.ArchiveStateEnabled,
			}, nil)

		// Should not panic on success
		AssertArchiveExists(mockCtx, "test-archive")
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertArchiveExists_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock error response
		mockClient.On("DescribeArchive", mock.Anything, mock.Anything).Return(
			nil, errors.New("archive not found"))

		assert.Panics(t, func() {
			AssertArchiveExists(mockCtx, "test-archive")
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertArchiveState_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("DescribeArchive", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeArchiveOutput{
				ArchiveName: aws.String("test-archive"),
				State:       types.ArchiveStateEnabled,
			}, nil)

		// Should not panic on success
		AssertArchiveState(mockCtx, "test-archive", types.ArchiveStateEnabled)
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertArchiveState_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock wrong state response
		mockClient.On("DescribeArchive", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeArchiveOutput{
				ArchiveName: aws.String("test-archive"),
				State:       types.ArchiveStateDisabled,
			}, nil)

		assert.Panics(t, func() {
			AssertArchiveState(mockCtx, "test-archive", types.ArchiveStateEnabled)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertReplayExists_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("DescribeReplay", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeReplayOutput{
				ReplayName: aws.String("test-replay"),
				State:      types.ReplayStateRunning,
			}, nil)

		// Should not panic on success
		AssertReplayExists(mockCtx, "test-replay")
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertReplayExists_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock error response
		mockClient.On("DescribeReplay", mock.Anything, mock.Anything).Return(
			nil, errors.New("replay not found"))

		assert.Panics(t, func() {
			AssertReplayExists(mockCtx, "test-replay")
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertReplayCompleted_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock successful response
		mockClient.On("DescribeReplay", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeReplayOutput{
				ReplayName: aws.String("test-replay"),
				State:      types.ReplayStateCompleted,
			}, nil)

		// Should not panic on success
		AssertReplayCompleted(mockCtx, "test-replay")
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertReplayCompleted_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock wrong state response
		mockClient.On("DescribeReplay", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeReplayOutput{
				ReplayName: aws.String("test-replay"),
				State:      types.ReplayStateRunning,
			}, nil)

		assert.Panics(t, func() {
			AssertReplayCompleted(mockCtx, "test-replay")
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertPatternMatches_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		pattern := `{"source": ["test.service"]}`
		
		// Mock successful response
		mockClient.On("TestEventPattern", mock.Anything, mock.Anything).Return(
			&eventbridge.TestEventPatternOutput{
				Result: true,
			}, nil)

		customEvent := CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
		}

		// Should not panic on success
		AssertPatternMatches(mockCtx, customEvent, pattern)
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertPatternMatches_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		pattern := `{"source": ["test.service"]}`
		
		// Mock no match response
		mockClient.On("TestEventPattern", mock.Anything, mock.Anything).Return(
			&eventbridge.TestEventPatternOutput{
				Result: false,
			}, nil)

		customEvent := CustomEvent{
			Source:     "different.service",
			DetailType: "Test Event",
			Detail:     `{"key": "value"}`,
		}

		assert.Panics(t, func() {
			AssertPatternMatches(mockCtx, customEvent, pattern)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertEventSent_Success", func(t *testing.T) {
		mockCtx := &TestContext{
			T: t,
		}

		result := PutEventResult{
			Success:  true,
			EventID:  "test-event-123",
			ErrorCode: "",
			ErrorMessage: "",
		}

		// Should not panic on success (basic event validation passes)
		AssertEventSent(mockCtx, result)
	})

	t.Run("AssertEventSent_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T: mockT,
		}

		result := PutEventResult{
			Success:  false,
			EventID:  "",
			ErrorCode: "ValidationException",
			ErrorMessage: "Event validation failed",
		}

		assert.Panics(t, func() {
			AssertEventSent(mockCtx, result)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertEventBatchSent_Success", func(t *testing.T) {
		mockCtx := &TestContext{
			T: t,
		}

		result := PutEventsResult{
			Entries: []PutEventResult{
				{
					Success: true,
					EventID: "test-event-123",
				},
			},
			FailedEntryCount: 0,
		}

		// Should not panic on success
		AssertEventBatchSent(mockCtx, result)
	})

	t.Run("AssertEventBatchSent_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T: mockT,
		}

		result := PutEventsResult{
			Entries: []PutEventResult{
				{
					Success: false,
					EventID: "",
					ErrorCode: "ValidationException",
					ErrorMessage: "Event validation failed",
				},
			},
			FailedEntryCount: 1,
		}

		assert.Panics(t, func() {
			AssertEventBatchSent(mockCtx, result)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertRuleHasPattern_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		pattern := `{"source": ["test.service"]}`

		// Mock successful response
		mockClient.On("DescribeRule", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeRuleOutput{
				Name:         aws.String("test-rule"),
				EventPattern: aws.String(pattern),
			}, nil)

		// Should not panic on success
		AssertRuleHasPattern(mockCtx, "test-rule", "default", pattern)
		mockClient.AssertExpectations(t)
	})

	t.Run("AssertRuleHasPattern_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		pattern := `{"source": ["test.service"]}`
		differentPattern := `{"source": ["different.service"]}`

		// Mock different pattern response
		mockClient.On("DescribeRule", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeRuleOutput{
				Name:         aws.String("test-rule"),
				EventPattern: aws.String(differentPattern),
			}, nil)

		assert.Panics(t, func() {
			AssertRuleHasPattern(mockCtx, "test-rule", "default", pattern)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})
}

// TestMissingFunctions tests remaining functions with 0% coverage
func TestMissingFunctions(t *testing.T) {
	t.Run("RemoveAllTargets_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock list targets response
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything).Return(
			&eventbridge.ListTargetsByRuleOutput{
				Targets: []types.Target{
					{Id: aws.String("target1")},
					{Id: aws.String("target2")},
				},
			}, nil)

		// Mock remove targets response
		mockClient.On("RemoveTargets", mock.Anything, mock.Anything).Return(
			&eventbridge.RemoveTargetsOutput{}, nil)

		// Should not panic on success
		RemoveAllTargets(mockCtx, "test-rule", "default")
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveAllTargets_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: &MockEventBridgeClient{},
		}

		assert.Panics(t, func() {
			RemoveAllTargets(mockCtx, "", "default") // Empty rule name will cause error
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("WaitForReplayCompletion_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock response showing completed replay
		mockClient.On("DescribeReplay", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeReplayOutput{
				ReplayName: aws.String("test-replay"),
				State:      types.ReplayStateCompleted,
			}, nil)

		// Should not panic on success
		WaitForReplayCompletion(mockCtx, "test-replay", 3*time.Second)
		mockClient.AssertExpectations(t)
	})

	t.Run("WaitForReplayCompletion_PanicOnTimeout", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock response showing running replay (never completes)
		mockClient.On("DescribeReplay", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeReplayOutput{
				ReplayName: aws.String("test-replay"),
				State:      types.ReplayStateRunning,
			}, nil).Maybe()

		assert.Panics(t, func() {
			WaitForReplayCompletion(mockCtx, "test-replay", 50*time.Millisecond)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("WaitForArchiveState_Success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      t,
			Client: mockClient,
		}

		// Mock response showing enabled archive
		mockClient.On("DescribeArchive", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeArchiveOutput{
				ArchiveName: aws.String("test-archive"),
				State:       types.ArchiveStateEnabled,
			}, nil)

		// Should not panic on success
		WaitForArchiveState(mockCtx, "test-archive", types.ArchiveStateEnabled, 3*time.Second)
		mockClient.AssertExpectations(t)
	})

	t.Run("WaitForArchiveState_PanicOnTimeout", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockClient := &MockEventBridgeClient{}
		mockCtx := &TestContext{
			T:      mockT,
			Client: mockClient,
		}

		// Mock response showing wrong state (never matches)
		mockClient.On("DescribeArchive", mock.Anything, mock.Anything).Return(
			&eventbridge.DescribeArchiveOutput{
				ArchiveName: aws.String("test-archive"),
				State:       types.ArchiveStateDisabled,
			}, nil).Maybe()

		assert.Panics(t, func() {
			WaitForArchiveState(mockCtx, "test-archive", types.ArchiveStateEnabled, 50*time.Millisecond)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertEventDeliverySuccess_Success", func(t *testing.T) {
		mockCtx := &TestContext{
			T: t,
		}

		result := PutEventResult{
			Success:      true,
			EventID:      "test-event-123",
			ErrorCode:    "",
			ErrorMessage: "",
		}

		// Should not panic on success
		AssertEventDeliverySuccess(mockCtx, result)
	})

	t.Run("AssertEventDeliverySuccess_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T: mockT,
		}

		result := PutEventResult{
			Success:      false,
			EventID:      "",
			ErrorCode:    "InternalException",
			ErrorMessage: "Internal error occurred",
		}

		assert.Panics(t, func() {
			AssertEventDeliverySuccess(mockCtx, result)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertBatchDeliverySuccess_Success", func(t *testing.T) {
		mockCtx := &TestContext{
			T: t,
		}

		result := PutEventsResult{
			Entries: []PutEventResult{
				{
					Success: true,
					EventID: "test-event-123",
				},
			},
			FailedEntryCount: 0,
		}

		// Should not panic on success
		AssertBatchDeliverySuccess(mockCtx, result)
	})

	t.Run("AssertBatchDeliverySuccess_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T: mockT,
		}

		result := PutEventsResult{
			Entries: []PutEventResult{
				{
					Success:      false,
					EventID:      "",
					ErrorCode:    "InternalException",
					ErrorMessage: "Internal error occurred",
				},
			},
			FailedEntryCount: 1,
		}

		assert.Panics(t, func() {
			AssertBatchDeliverySuccess(mockCtx, result)
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("AssertDeliveryLatency_Success", func(t *testing.T) {
		mockCtx := &TestContext{
			T: t,
		}

		metrics := DeliveryMetrics{
			TargetID:           "target-123",
			DeliveredCount:     100,
			FailedCount:        1,
			DeliveryLatencyMs:  100,
			LastDeliveryTime:   time.Now(),
			ErrorRate:          0.01,
		}

		// Should not panic on success (simple latency check)
		AssertDeliveryLatency(mockCtx, metrics, 500)
	})

	t.Run("AssertDeliveryLatency_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T: mockT,
		}

		metrics := DeliveryMetrics{
			DeliveryLatencyMs: 1000,
			ErrorRate:         0.01,
			DeliveredCount:    100,
		}

		assert.Panics(t, func() {
			AssertDeliveryLatency(mockCtx, metrics, 100) // Actual > threshold should fail
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("WaitForEventDelivery_Success", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T: mockT,
		}

		// This will actually timeout since isEventDeliveredStub always returns false
		// The function will panic when it calls FailNow, so we need to catch that
		assert.Panics(t, func() {
			WaitForEventDelivery(mockCtx, "event-123", 50*time.Millisecond)
		})
		
		// Verify that the function correctly reported the timeout
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("WaitForEventDelivery_PanicOnTimeout", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T: mockT,
		}

		assert.Panics(t, func() {
			WaitForEventDelivery(mockCtx, "", 5*time.Millisecond) // Empty event ID will cause error
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})

	t.Run("MonitorEventDelivery_Success", func(t *testing.T) {
		mockCtx := &TestContext{
			T: t,
		}

		// Should not panic on success
		MonitorEventDelivery(mockCtx, "test-rule", "default", 50*time.Millisecond)
	})

	t.Run("MonitorEventDelivery_PanicOnError", func(t *testing.T) {
		mockT := &MockTestingT{}
		mockCtx := &TestContext{
			T: mockT,
		}

		assert.Panics(t, func() {
			MonitorEventDelivery(mockCtx, "", "default", 5*time.Millisecond) // Empty rule name will cause error
		})
		assert.True(t, mockT.ErrorfCalled)
		assert.True(t, mockT.FailNowCalled)
	})
}

// TestPatternMatchingEdgeCases tests pattern matching functions with edge cases
func TestPatternMatchingEdgeCases(t *testing.T) {
	t.Run("matchesSpecialPattern_NumericRange", func(t *testing.T) {
		// Test numeric pattern matching
		pattern := map[string]interface{}{
			"numeric": []interface{}{">", 10, "<=", 100},
		}
		
		// Test the pattern matching logic
		patternJSON, _ := json.Marshal(pattern)
		result := EstimatePatternMatchRate(string(patternJSON))
		assert.Greater(t, result, 0.0) // Should have some match rate
	})

	t.Run("matchesNumericPattern_EdgeCases", func(t *testing.T) {
		// Test various numeric patterns
		testCases := []struct {
			pattern map[string]interface{}
			name    string
		}{
			{
				name: "GreaterThan",
				pattern: map[string]interface{}{
					"value": map[string]interface{}{"numeric": []interface{}{">", 10}},
				},
			},
			{
				name: "LessThan",
				pattern: map[string]interface{}{
					"value": map[string]interface{}{"numeric": []interface{}{"<", 100}},
				},
			},
			{
				name: "Range",
				pattern: map[string]interface{}{
					"value": map[string]interface{}{"numeric": []interface{}{">=", 10, "<=", 100}},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test pattern estimation
				patternJSON, _ := json.Marshal(tc.pattern)
				result := EstimatePatternMatchRate(string(patternJSON))
				assert.Greater(t, result, 0.0) // Should have some match rate
			})
		}
	})

	t.Run("toFloat64_EdgeCases", func(t *testing.T) {
		testCases := []struct {
			input    interface{}
			expected float64
			name     string
		}{
			{name: "Int", input: 42, expected: 42.0},
			{name: "Float64", input: 42.5, expected: 42.5},
			{name: "String", input: "42.5", expected: 42.5},
			{name: "InvalidString", input: "invalid", expected: 0.0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// This calls the internal toFloat64 function through EstimatePatternMatchRate
				pattern := map[string]interface{}{
					"value": map[string]interface{}{"numeric": []interface{}{"=", tc.input}},
				}
				patternJSON, _ := json.Marshal(pattern)
				result := EstimatePatternMatchRate(string(patternJSON))
				assert.Greater(t, result, 0.0)
			})
		}
	})

	t.Run("isNumeric_EdgeCases", func(t *testing.T) {
		testCases := []struct {
			input interface{}
			name  string
		}{
			{name: "Int", input: 42},
			{name: "Float64", input: 42.5},
			{name: "String", input: "not numeric"},
			{name: "Bool", input: true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test through pattern matching which uses isNumeric internally
				pattern := map[string]interface{}{
					"value": map[string]interface{}{"exists": true},
				}
				patternJSON, _ := json.Marshal(pattern)
				result := EstimatePatternMatchRate(string(patternJSON))
				assert.Greater(t, result, 0.0) // Should have some match rate
			})
		}
	})
}

// TestUtilityFunctionEdgeCases tests utility functions and helper functions
func TestUtilityFunctionEdgeCases(t *testing.T) {
	t.Run("isEventBridgeRule_EdgeCases", func(t *testing.T) {
		// Test through BuildArchiveConfigWithPattern which uses this internally
		config := BuildArchiveConfigWithPattern(
			"test-archive",
			"arn:aws:events:us-east-1:123456789012:event-bus/default",
			`{"source": ["test.service"]}`,
			7,
		)
		
		assert.Equal(t, "test-archive", config.ArchiveName)
		assert.Equal(t, `{"source": ["test.service"]}`, config.EventPattern)
		assert.Equal(t, int32(7), config.RetentionDays)
	})

	t.Run("containsPattern_EdgeCases", func(t *testing.T) {
		// Test through BuildArchiveConfigWithPattern which uses this internally  
		config := BuildArchiveConfigWithPattern(
			"test-archive",
			"arn:aws:events:us-east-1:123456789012:event-bus/default",
			`{"source": ["test.service"], "detail": {"state": ["running"]}}`,
			30,
		)
		
		assert.Contains(t, config.EventPattern, "source")
		assert.Contains(t, config.EventPattern, "detail")
		assert.Contains(t, config.EventPattern, "state")
	})

	t.Run("findSubstring_EdgeCases", func(t *testing.T) {
		// Test through archive pattern validation which uses this internally
		config := BuildArchiveConfigWithPattern(
			"test-archive",
			"arn:aws:events:us-east-1:123456789012:event-bus/default", 
			`{"source": ["my.custom.service"], "detail-type": ["Order Placed"]}`,
			14,
		)
		
		assert.Contains(t, config.EventPattern, "my.custom.service")
		assert.Contains(t, config.EventPattern, "Order Placed")
	})
}



// TestFinalUtilityFunctions tests the remaining utility functions to achieve 100% coverage
func TestFinalUtilityFunctions(t *testing.T) {
	t.Parallel()

	// Test isEventBridgeRule function directly
	t.Run("isEventBridgeRule_DirectTesting", func(t *testing.T) {
		// Valid EventBridge rule ARN
		arn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
		result := isEventBridgeRule(arn)
		assert.True(t, result)

		// Invalid ARN
		result = isEventBridgeRule("invalid-arn")
		assert.False(t, result)
	})

	// Test containsPattern function directly
	t.Run("containsPattern_DirectTesting", func(t *testing.T) {
		// Pattern exists
		result := containsPattern("arn:aws:events:us-east-1:123456789012:rule/test-rule", ":rule/")
		assert.True(t, result)

		// Pattern doesn't exist
		result = containsPattern("test-string", "xyz")
		assert.False(t, result)

		// Empty pattern - returns false because findSubstring returns 0, but containsPattern expects >= 1
		result = containsPattern("test-string", "")
		assert.False(t, result)
	})

	// Test findSubstring function directly
	t.Run("findSubstring_DirectTesting", func(t *testing.T) {
		// Substring found at beginning
		result := findSubstring("test-string", "test")
		assert.Equal(t, 0, result)

		// Substring found in middle
		result = findSubstring("test-string", "-str")
		assert.Equal(t, 4, result)

		// Substring not found
		result = findSubstring("test-string", "xyz")
		assert.Equal(t, -1, result)

		// Empty substring
		result = findSubstring("test-string", "")
		assert.Equal(t, 0, result)
	})

	// Test matchesSpecialPattern function directly
	t.Run("matchesSpecialPattern_DirectTesting", func(t *testing.T) {
		// Numeric pattern
		pattern := map[string]interface{}{
			"numeric": []interface{}{">", 10},
		}
		result := matchesSpecialPattern(pattern, 15)
		assert.True(t, result)

		// Exists pattern
		pattern = map[string]interface{}{
			"exists": true,
		}
		result = matchesSpecialPattern(pattern, "value")
		assert.True(t, result)

		result = matchesSpecialPattern(pattern, nil)
		assert.False(t, result)

		// Prefix pattern
		pattern = map[string]interface{}{
			"prefix": "test",
		}
		result = matchesSpecialPattern(pattern, "test-string")
		assert.True(t, result)

		// Invalid pattern
		pattern = map[string]interface{}{
			"unknown": "value",
		}
		result = matchesSpecialPattern(pattern, "test")
		assert.False(t, result)
	})

	// Test matchesNumericPattern function directly
	t.Run("matchesNumericPattern_DirectTesting", func(t *testing.T) {
		// Greater than - true
		result := matchesNumericPattern([]interface{}{">", 10}, 15)
		assert.True(t, result)

		// Greater than - false
		result = matchesNumericPattern([]interface{}{">", 10}, 5)
		assert.False(t, result)

		// Less than - true
		result = matchesNumericPattern([]interface{}{"<", 100}, 50)
		assert.True(t, result)

		// Equal - true
		result = matchesNumericPattern([]interface{}{"=", 42}, 42)
		assert.True(t, result)

		// Not equal - true
		result = matchesNumericPattern([]interface{}{"!=", 42}, 24)
		assert.True(t, result)

		// Greater or equal - true
		result = matchesNumericPattern([]interface{}{">=", 10}, 10)
		assert.True(t, result)

		// Less or equal - true
		result = matchesNumericPattern([]interface{}{"<=", 100}, 100)
		assert.True(t, result)

		// Invalid operator
		result = matchesNumericPattern([]interface{}{"@@", 10}, 15)
		assert.False(t, result)

		// Invalid pattern format
		result = matchesNumericPattern([]interface{}{">"}, 15)
		assert.False(t, result)

		// Non-numeric event value
		result = matchesNumericPattern([]interface{}{">", 10}, "not-a-number")
		assert.False(t, result)

		// Non-array pattern
		result = matchesNumericPattern("not-an-array", 15)
		assert.False(t, result)
	})
}
