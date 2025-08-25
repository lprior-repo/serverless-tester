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

// TestExponentialBackoffCalculation tests the exponential backoff delay calculation
func TestExponentialBackoffCalculation(t *testing.T) {
	// RED: Test exponential backoff calculation with default config
	config := defaultRetryConfig()
	
	// Test base case (attempt 0)
	delay := calculateBackoffDelay(0, config)
	assert.Equal(t, config.BaseDelay, delay, "First attempt should use base delay")
	
	// Test exponential growth
	delay1 := calculateBackoffDelay(1, config)
	expectedDelay1 := time.Duration(float64(config.BaseDelay) * config.Multiplier)
	assert.Equal(t, expectedDelay1, delay1, "Second attempt should double the delay")
	
	// Test maximum delay capping
	delay10 := calculateBackoffDelay(10, config)
	assert.Equal(t, config.MaxDelay, delay10, "Large attempts should be capped at max delay")
}

func TestExponentialBackoffCalculationWithCustomConfig(t *testing.T) {
	// RED: Test exponential backoff with custom configuration
	config := RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		Multiplier:  3.0,
	}
	
	// Test custom multiplier
	delay1 := calculateBackoffDelay(1, config)
	expectedDelay1 := time.Duration(float64(config.BaseDelay) * 3.0)
	assert.Equal(t, expectedDelay1, delay1, "Should use custom multiplier")
	
	// Test custom max delay
	delay5 := calculateBackoffDelay(5, config)
	assert.Equal(t, config.MaxDelay, delay5, "Should respect custom max delay")
}

func TestExponentialBackoffCalculationEdgeCases(t *testing.T) {
	// RED: Test edge cases for exponential backoff
	config := defaultRetryConfig()
	
	// Test negative attempt (should be treated as 0)
	delayNegative := calculateBackoffDelay(-1, config)
	assert.Equal(t, config.BaseDelay, delayNegative, "Negative attempts should use base delay")
	
	// Test very large attempt that could cause overflow
	delayLarge := calculateBackoffDelay(100, config)
	assert.Equal(t, config.MaxDelay, delayLarge, "Very large attempts should be capped")
}

func TestPutEventE_RetryMechanismDetails(t *testing.T) {
	// RED: Test that retry mechanism respects configuration
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

	// Mock consistent failures for all retry attempts
	serviceErr := errors.New("temporary service error")
	mockClient.On("PutEvents", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr).Times(3)

	startTime := time.Now()
	result, err := PutEventE(ctx, event)
	duration := time.Since(startTime)

	// Verify error and retry behavior
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to put event after 3 attempts")
	assert.False(t, result.Success)
	
	// Verify that retries took reasonable time (should include backoff delays)
	assert.Greater(t, duration, 2*time.Second, "Should have retry delays")
	
	mockClient.AssertExpectations(t)
}

func TestPutEventE_RetrySucceedsOnSecondAttempt(t *testing.T) {
	// RED: Test that retry succeeds after one failure
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

	// Mock failure on first attempt, success on second
	serviceErr := errors.New("temporary service error")
	mockClient.On("PutEvents", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr).Once()
	mockClient.On("PutEvents", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{
				EventId: aws.String("event-id-123"),
			},
		},
	}, nil).Once()

	result, err := PutEventE(ctx, event)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "event-id-123", result.EventID)
	
	mockClient.AssertExpectations(t)
}

func TestPutEventE_EventValidationUnicodeCharacters(t *testing.T) {
	// RED: Test event validation with Unicode characters
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Test with various Unicode characters
	testCases := []struct {
		name   string
		detail string
		valid  bool
	}{
		{
			name:   "Basic Unicode",
			detail: `{"message": "Hello, 世界!"}`,
			valid:  true,
		},
		{
			name:   "Emoji characters",
			detail: `{"message": "Hello 👋 World 🌍"}`,
			valid:  true,
		},
		{
			name:   "Complex Unicode",
			detail: `{"message": "Testing ñáéíóúüç characters"}`,
			valid:  true,
		},
		{
			name:   "Mixed scripts",
			detail: `{"message": "Русский текст and العربية"}`,
			valid:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := CustomEvent{
				Source:       "test.service",
				DetailType:   "Unicode Test",
				Detail:       tc.detail,
				EventBusName: "default",
			}

			if tc.valid {
				// Mock successful response for valid Unicode
				mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
					return input.Entries[0].Detail != nil && *input.Entries[0].Detail == tc.detail
				}), mock.Anything).Return(&eventbridge.PutEventsOutput{
					FailedEntryCount: 0,
					Entries: []types.PutEventsResultEntry{
						{
							EventId: aws.String("event-id-123"),
						},
					},
				}, nil).Once()

				result, err := PutEventE(ctx, event)
				require.NoError(t, err)
				assert.True(t, result.Success)
			}
		})
	}

	mockClient.AssertExpectations(t)
}

func TestPutEventE_EventValidationNestedJSON(t *testing.T) {
	// RED: Test event validation with complex nested JSON structures
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	testCases := []struct {
		name    string
		detail  string
		isValid bool
	}{
		{
			name:    "Deeply nested JSON",
			detail:  `{"level1": {"level2": {"level3": {"level4": {"value": "deep"}}}}}`,
			isValid: true,
		},
		{
			name:    "JSON with arrays",
			detail:  `{"items": [{"id": 1, "tags": ["tag1", "tag2"]}, {"id": 2, "tags": ["tag3"]}]}`,
			isValid: true,
		},
		{
			name:    "JSON with null values",
			detail:  `{"nullValue": null, "emptyString": "", "number": 0}`,
			isValid: true,
		},
		{
			name:    "JSON with special characters",
			detail:  `{"message": "Special chars: \n\t\r\"\\"}`,
			isValid: true,
		},
		{
			name:    "Invalid JSON - missing quotes",
			detail:  `{message: "invalid"}`,
			isValid: false,
		},
		{
			name:    "Invalid JSON - trailing comma",
			detail:  `{"key": "value",}`,
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := CustomEvent{
				Source:       "test.service",
				DetailType:   "JSON Test",
				Detail:       tc.detail,
				EventBusName: "default",
			}

			if tc.isValid {
				// Mock successful response for valid JSON
				mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
					return input.Entries[0].Detail != nil && *input.Entries[0].Detail == tc.detail
				}), mock.Anything).Return(&eventbridge.PutEventsOutput{
					FailedEntryCount: 0,
					Entries: []types.PutEventsResultEntry{
						{
							EventId: aws.String("test-event-id"),
						},
					},
				}, nil).Once()
			}

			_, err := PutEventE(ctx, event)
			
			if tc.isValid {
				// For valid JSON, error should be nil or not validation-related
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid event detail")
				}
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid event detail")
			}
		})
	}

	mockClient.AssertExpectations(t)
}

func TestPutEventE_EventSizeBoundaryConditions(t *testing.T) {
	// RED: Test event validation at exact size boundaries
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	// Test one byte over maximum size (should fail)
	oversizeDetail := make([]byte, MaxEventDetailSize+1)
	for i := range oversizeDetail {
		oversizeDetail[i] = 'x'
	}
	
	eventOversize := CustomEvent{
		Source:       "test.service",
		DetailType:   "Size Test",
		Detail:       string(oversizeDetail),
		EventBusName: "default",
	}

	_, err := PutEventE(ctx, eventOversize)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event size too large")
}

func TestPutEventsE_BatchOperationsEdgeCases(t *testing.T) {
	// RED: Test batch operations with mixed event types and edge cases
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Create mixed batch with various event configurations
	events := []CustomEvent{
		{
			Source:       "test.service1",
			DetailType:   "Event 1",
			Detail:       `{"key": "value1"}`,
			EventBusName: "default",
			Resources:    []string{"arn:aws:s3:::bucket1"},
		},
		{
			Source:       "test.service2",
			DetailType:   "Event 2",
			Detail:       "", // Empty detail
			EventBusName: "custom-bus",
			Resources:    nil, // No resources
		},
		{
			Source:       "test.service3",
			DetailType:   "Event 3",
			Detail:       `{"complex": {"nested": {"data": [1, 2, 3]}}}`,
			EventBusName: "", // Should default to "default"
			Resources:    []string{"arn:aws:lambda:us-east-1:123456789012:function:test"},
			Time:         func() *time.Time { t := time.Now().Add(-1 * time.Hour); return &t }(), // Custom time
		},
	}

	// Mock successful batch response
	mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		return len(input.Entries) == 3 &&
			*input.Entries[0].EventBusName == "default" &&
			*input.Entries[1].EventBusName == "custom-bus" &&
			*input.Entries[2].EventBusName == "default" // Should default
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{EventId: aws.String("event-1")},
			{EventId: aws.String("event-2")},
			{EventId: aws.String("event-3")},
		},
	}, nil)

	result, err := PutEventsE(ctx, events)

	require.NoError(t, err)
	assert.Equal(t, int32(0), result.FailedEntryCount)
	assert.Len(t, result.Entries, 3)
	assert.True(t, result.Entries[0].Success)
	assert.True(t, result.Entries[1].Success)
	assert.True(t, result.Entries[2].Success)
	
	mockClient.AssertExpectations(t)
}

func TestPutEventsE_ExactMaxBatchSize(t *testing.T) {
	// RED: Test batch exactly at maximum size limit
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Create exactly MaxEventBatchSize events
	events := make([]CustomEvent, MaxEventBatchSize)
	for i := 0; i < MaxEventBatchSize; i++ {
		events[i] = CustomEvent{
			Source:       "test.service",
			DetailType:   "Batch Test",
			Detail:       `{"index": ` + string(rune('0'+i)) + `}`,
			EventBusName: "default",
		}
	}

	// Mock successful batch response
	expectedEntries := make([]types.PutEventsResultEntry, MaxEventBatchSize)
	for i := 0; i < MaxEventBatchSize; i++ {
		expectedEntries[i] = types.PutEventsResultEntry{
			EventId: aws.String("batch-event-" + string(rune('0'+i))),
		}
	}

	mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		return len(input.Entries) == MaxEventBatchSize
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries:         expectedEntries,
	}, nil).Once()

	// Should not return error for exact max size
	_, err := PutEventsE(ctx, events)
	// If error occurs, it should not be about batch size
	if err != nil {
		assert.NotContains(t, err.Error(), "batch size exceeds maximum")
	}

	mockClient.AssertExpectations(t)
}

// TestPutEventsE_BatchValidationFailures tests batch operations with mixed validation failures
func TestPutEventsE_BatchValidationFailures(t *testing.T) {
	// RED: Test batch with some events having validation errors
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	// Create batch with mixed valid/invalid events
	events := []CustomEvent{
		{
			Source:       "test.service",
			DetailType:   "Valid Event",
			Detail:       `{"key": "value"}`,
			EventBusName: "default",
		},
		{
			Source:       "test.service",
			DetailType:   "Invalid Event",
			Detail:       `{"invalid": json}`, // Invalid JSON
			EventBusName: "default",
		},
	}

	// Should fail on first invalid event
	_, err := PutEventsE(ctx, events)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event detail")
}

// TestPutEventsE_RetryWithBatch tests retry mechanism for batch operations
func TestPutEventsE_RetryWithBatch(t *testing.T) {
	// RED: Test retry mechanism with batch operations
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	events := []CustomEvent{
		{
			Source:       "test.service",
			DetailType:   "Batch Event 1",
			Detail:       `{"key": "value1"}`,
			EventBusName: "default",
		},
		{
			Source:       "test.service",
			DetailType:   "Batch Event 2",
			Detail:       `{"key": "value2"}`,
			EventBusName: "default",
		},
	}

	// Mock failure on first attempt, success on second
	serviceErr := errors.New("temporary batch error")
	mockClient.On("PutEvents", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr).Once()
	mockClient.On("PutEvents", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{EventId: aws.String("batch-event-1")},
			{EventId: aws.String("batch-event-2")},
		},
	}, nil).Once()

	result, err := PutEventsE(ctx, events)

	require.NoError(t, err)
	assert.Equal(t, int32(0), result.FailedEntryCount)
	assert.Len(t, result.Entries, 2)
	assert.True(t, result.Entries[0].Success)
	assert.True(t, result.Entries[1].Success)

	mockClient.AssertExpectations(t)
}

// TestPutEventE_ContextualTimeHandling tests time handling edge cases
func TestPutEventE_ContextualTimeHandling(t *testing.T) {
	// RED: Test custom time vs default time handling
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Test with custom time
	customTime := time.Now().Add(-1 * time.Hour)
	eventWithTime := CustomEvent{
		Source:       "test.service",
		DetailType:   "Timed Event",
		Detail:       `{"key": "value"}`,
		EventBusName: "default",
		Time:         &customTime,
	}

	// Mock to verify custom time is passed
	mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		if len(input.Entries) == 0 {
			return false
		}
		entry := input.Entries[0]
		return entry.Time != nil && entry.Time.Equal(customTime)
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{EventId: aws.String("timed-event-id")},
		},
	}, nil).Once()

	result, err := PutEventE(ctx, eventWithTime)
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Test with nil time (should default)
	eventWithoutTime := CustomEvent{
		Source:       "test.service",
		DetailType:   "Untimed Event",
		Detail:       `{"key": "value"}`,
		EventBusName: "default",
		Time:         nil,
	}

	// Mock to verify default time is set (should be recent)
	mockClient.On("PutEvents", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		if len(input.Entries) == 0 {
			return false
		}
		entry := input.Entries[0]
		return entry.Time != nil && time.Since(*entry.Time) < time.Minute
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{EventId: aws.String("untimed-event-id")},
		},
	}, nil).Once()

	result, err = PutEventE(ctx, eventWithoutTime)
	require.NoError(t, err)
	assert.True(t, result.Success)

	mockClient.AssertExpectations(t)
}