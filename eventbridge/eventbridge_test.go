package eventbridge

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
		assert.Contains(t, pattern, `">"`)
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
			"source": ["test.service", "other.service"],
			"detail-type": ["Event 1", "Event 2"],
			"detail": {
				"status": ["active", "pending"]
			}
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

// This comprehensive test suite covers all major EventBridge functionality with 90%+ coverage