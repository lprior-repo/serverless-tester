package eventbridge

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

// Test PutEventE with empty event detail  
func TestPutEventE_EmptyEventDetail(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutEvents", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{
				EventId: aws.String("event-123"),
			},
		},
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	event := CustomEvent{
		Source:       "test.service",
		DetailType:   "Test Event",
		Detail:       "", // Empty detail should be valid
		EventBusName: "default",
	}

	result, err := PutEventE(ctx, event)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "event-123", result.EventID)
	mockClient.AssertExpectations(t)
}

// Test PutEventE with no response entries (edge case)
func TestPutEventE_NoResponseEntries(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutEvents", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries:         []types.PutEventsResultEntry{}, // Empty entries
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	event := CustomEvent{
		Source:       "test.service",
		DetailType:   "Test Event",
		Detail:       `{"key": "value"}`,
		EventBusName: "default",
	}

	_, err := PutEventE(ctx, event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to put event after")
	mockClient.AssertExpectations(t)
}

// Test PutEventsE with events having default times set
func TestPutEventsE_DefaultTimesSet(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	mockClient.On("PutEvents", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{
				EventId: aws.String("event-123"),
			},
			{
				EventId: aws.String("event-456"),
			},
		},
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	events := []CustomEvent{
		{
			Source:       "test.service1",
			DetailType:   "Test Event 1",
			Detail:       `{"key": "value1"}`,
			EventBusName: "", // Should default to "default"
			Time:         nil, // Should be set automatically
		},
		{
			Source:       "test.service2",
			DetailType:   "Test Event 2",
			Detail:       `{"key": "value2"}`,
			EventBusName: "custom-bus",
			Time:         nil, // Should be set automatically
		},
	}

	result, err := PutEventsE(ctx, events)

	require.NoError(t, err)
	assert.Equal(t, int32(0), result.FailedEntryCount)
	assert.Len(t, result.Entries, 2)
	
	assert.True(t, result.Entries[0].Success)
	assert.Equal(t, "event-123", result.Entries[0].EventID)
	
	assert.True(t, result.Entries[1].Success)
	assert.Equal(t, "event-456", result.Entries[1].EventID)
	
	mockClient.AssertExpectations(t)
}

// Test PutEventsE with retry failure
func TestPutEventsE_RetryFailure(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	// All attempts fail
	mockClient.On("PutEvents", context.Background(), mock.Anything, mock.Anything).Return(nil, errors.New("persistent failure")).Times(DefaultRetryAttempts)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	events := []CustomEvent{
		{
			Source:       "test.service",
			DetailType:   "Test Event",
			Detail:       `{"key": "value"}`,
			EventBusName: "default",
		},
	}

	result, err := PutEventsE(ctx, events)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to put events after")
	assert.Empty(t, result.Entries)
	mockClient.AssertExpectations(t)
}

// Test BuildCustomEvent with marshal error (by providing invalid detail)
func TestBuildCustomEvent_MarshalError(t *testing.T) {
	// Using a channel in the detail map will cause json.Marshal to fail
	detail := map[string]interface{}{
		"channel": make(chan int), // Cannot be marshaled to JSON
		"valid":   "data",
	}
	
	event := BuildCustomEvent("test.service", "Test Event", detail)
	
	assert.Equal(t, "test.service", event.Source)
	assert.Equal(t, "Test Event", event.DetailType)
	assert.Equal(t, "{}", event.Detail) // Should fall back to empty JSON
	assert.Equal(t, "default", event.EventBusName)
}

// Test BuildScheduledEvent function
func TestBuildScheduledEvent_ValidDetail(t *testing.T) {
	detail := map[string]interface{}{
		"action": "scheduled-task",
		"params": map[string]interface{}{
			"timeout": 300,
			"region":  "us-east-1",
		},
	}
	
	event := BuildScheduledEvent("Scheduled Task", detail)
	
	assert.Equal(t, "aws.events", event.Source)
	assert.Equal(t, "Scheduled Task", event.DetailType)
	assert.Contains(t, event.Detail, `"action":"scheduled-task"`)
	assert.Contains(t, event.Detail, `"timeout":300`)
	assert.Contains(t, event.Detail, `"region":"us-east-1"`)
	assert.Equal(t, "default", event.EventBusName)
}

// Test ValidateEventPattern edge cases
func TestValidateEventPattern_AllValidKeys(t *testing.T) {
	pattern := `{
		"source": ["test.service"],
		"detail-type": ["Test Event"],
		"detail": {
			"key": ["value"]
		},
		"account": ["123456789012"],
		"region": ["us-east-1", "us-west-2"],
		"time": ["2023-01-01T00:00:00Z"],
		"id": ["event-id-123"],
		"resources": ["arn:aws:s3:::my-bucket"],
		"version": ["0"]
	}`
	
	assert.True(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_DetailTypeValidation(t *testing.T) {
	pattern := `{"detail-type": ["Event Type 1", "Event Type 2"]}`
	assert.True(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_AccountValidation(t *testing.T) {
	pattern := `{"account": ["123456789012", "123456789013"]}`
	assert.True(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_RegionValidation(t *testing.T) {
	pattern := `{"region": ["us-east-1", "us-west-2"]}`
	assert.True(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_ResourcesValidation(t *testing.T) {
	pattern := `{"resources": ["arn:aws:s3:::bucket1", "arn:aws:lambda:us-east-1:123:function:func1"]}`
	assert.True(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_DetailValidation(t *testing.T) {
	pattern := `{
		"detail": {
			"key1": ["value1", "value2"],
			"key2": {
				"nested": ["nested-value"]
			}
		}
	}`
	assert.True(t, ValidateEventPattern(pattern))
}

// Test error cases for ValidateEventPattern
func TestValidateEventPattern_NonStringInArray(t *testing.T) {
	pattern := `{"source": ["valid-source", 123]}`
	assert.False(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_NonArrayForSource(t *testing.T) {
	pattern := `{"source": "single-source"}`
	assert.False(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_NonArrayForDetailType(t *testing.T) {
	pattern := `{"detail-type": "single-detail-type"}`
	assert.False(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_NonArrayForAccount(t *testing.T) {
	pattern := `{"account": "123456789012"}`
	assert.False(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_NonArrayForRegion(t *testing.T) {
	pattern := `{"region": "us-east-1"}`
	assert.False(t, ValidateEventPattern(pattern))
}

func TestValidateEventPattern_NonArrayForResources(t *testing.T) {
	pattern := `{"resources": "arn:aws:s3:::bucket"}`
	assert.False(t, ValidateEventPattern(pattern))
}