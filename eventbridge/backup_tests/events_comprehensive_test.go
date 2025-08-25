package eventbridge

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// RED: Test PutEventE with successful event publication
func TestPutEventE_WithValidEvent_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	testEvent := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"key": "value"}`,
	}

	eventID := "test-event-id"
	mockClient.On("PutEvents", context.Background(), mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		return len(input.Entries) == 1 &&
			*input.Entries[0].Source == "test.service" &&
			*input.Entries[0].DetailType == "Test Event" &&
			*input.Entries[0].Detail == `{"key": "value"}`
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{EventId: &eventID},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := PutEventE(ctx, testEvent)

	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, eventID, result.EventID)
	mockClient.AssertExpectations(t)
}

// RED: Test PutEventE with invalid event detail
func TestPutEventE_WithInvalidEventDetail_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	testEvent := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"invalid": json}`, // Invalid JSON
	}

	result, err := PutEventE(ctx, testEvent)

	assert.Error(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, err.Error(), "invalid event detail")
}

// RED: Test PutEventE with default event bus
func TestPutEventE_WithoutEventBus_UsesDefaultEventBus(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	testEvent := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"key": "value"}`,
		// EventBusName not set
	}

	eventID := "test-event-id"
	mockClient.On("PutEvents", context.Background(), mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		return *input.Entries[0].EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{EventId: &eventID},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := PutEventE(ctx, testEvent)

	assert.NoError(t, err)
	assert.True(t, result.Success)
	mockClient.AssertExpectations(t)
}

// RED: Test PutEventE with failed entry
func TestPutEventE_WithFailedEntry_ReturnsError(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	testEvent := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"key": "value"}`,
	}

	errorCode := "ValidationException"
	errorMessage := "Event validation failed"
	mockClient.On("PutEvents", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 1,
		Entries: []types.PutEventsResultEntry{
			{
				ErrorCode:    &errorCode,
				ErrorMessage: &errorMessage,
			},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := PutEventE(ctx, testEvent)

	assert.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, errorCode, result.ErrorCode)
	assert.Equal(t, errorMessage, result.ErrorMessage)
	assert.Contains(t, err.Error(), "event failed")
}

// RED: Test PutEventE with retry on transient error
func TestPutEventE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	testEvent := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"key": "value"}`,
	}

	// First call fails, second succeeds
	eventID := "test-event-id"
	mockClient.On("PutEvents", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.PutEventsOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("PutEvents", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{EventId: &eventID},
		},
	}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := PutEventE(ctx, testEvent)

	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, eventID, result.EventID)
	mockClient.AssertExpectations(t)
}

// RED: Test PutEventsE with valid events
func TestPutEventsE_WithValidEvents_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	testEvents := []CustomEvent{
		{
			Source:     "test.service",
			DetailType: "Test Event 1",
			Detail:     `{"key": "value1"}`,
		},
		{
			Source:     "test.service",
			DetailType: "Test Event 2",
			Detail:     `{"key": "value2"}`,
		},
	}

	eventID1 := "test-event-id-1"
	eventID2 := "test-event-id-2"
	mockClient.On("PutEvents", context.Background(), mock.MatchedBy(func(input *eventbridge.PutEventsInput) bool {
		return len(input.Entries) == 2
	}), mock.Anything).Return(&eventbridge.PutEventsOutput{
		FailedEntryCount: 0,
		Entries: []types.PutEventsResultEntry{
			{EventId: &eventID1},
			{EventId: &eventID2},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := PutEventsE(ctx, testEvents)

	assert.NoError(t, err)
	assert.Equal(t, int32(0), result.FailedEntryCount)
	assert.Len(t, result.Entries, 2)
	assert.True(t, result.Entries[0].Success)
	assert.True(t, result.Entries[1].Success)
	assert.Equal(t, eventID1, result.Entries[0].EventID)
	assert.Equal(t, eventID2, result.Entries[1].EventID)
	mockClient.AssertExpectations(t)
}

// RED: Test PutEventsE with empty events slice
func TestPutEventsE_WithEmptyEvents_ReturnsEmptyResult(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	testEvents := []CustomEvent{}

	result, err := PutEventsE(ctx, testEvents)

	assert.NoError(t, err)
	assert.Len(t, result.Entries, 0)
}

// RED: Test PutEventsE with batch size exceeded
func TestPutEventsE_WithExcessiveBatchSize_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	
	// Create more events than MaxEventBatchSize
	testEvents := make([]CustomEvent, MaxEventBatchSize+1)
	for i := range testEvents {
		testEvents[i] = CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"index": ` + string(rune(i)) + `}`,
		}
	}

	result, err := PutEventsE(ctx, testEvents)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch size exceeds maximum")
	assert.Empty(t, result.Entries)
}

// RED: Test ValidateEventPattern with valid pattern
func TestValidateEventPattern_WithValidPattern_ReturnsTrue(t *testing.T) {
	validPattern := `{"source": ["aws.s3"]}`

	isValid := ValidateEventPattern(validPattern)

	assert.True(t, isValid)
}

// RED: Test ValidateEventPattern with empty pattern
func TestValidateEventPattern_WithEmptyPattern_ReturnsFalse(t *testing.T) {
	isValid := ValidateEventPattern("")

	assert.False(t, isValid)
}

// RED: Test ValidateEventPattern with invalid JSON
func TestValidateEventPattern_WithInvalidJSON_ReturnsFalse(t *testing.T) {
	invalidPattern := `{"source": ["aws.s3"]` // Missing closing brace

	isValid := ValidateEventPattern(invalidPattern)

	assert.False(t, isValid)
}

// RED: Test ValidateEventPattern with invalid key
func TestValidateEventPattern_WithInvalidKey_ReturnsFalse(t *testing.T) {
	invalidPattern := `{"invalid-key": ["value"]}`

	isValid := ValidateEventPattern(invalidPattern)

	assert.False(t, isValid)
}

// RED: Test ValidateEventPattern with non-array source
func TestValidateEventPattern_WithNonArraySource_ReturnsFalse(t *testing.T) {
	invalidPattern := `{"source": "aws.s3"}` // Should be array

	isValid := ValidateEventPattern(invalidPattern)

	assert.False(t, isValid)
}

// RED: Test ValidateEventPatternE with valid pattern
func TestValidateEventPatternE_WithValidPattern_ReturnsNil(t *testing.T) {
	validPattern := `{"source": ["aws.s3"], "detail-type": ["Object Created"]}`

	err := ValidateEventPatternE(validPattern)

	assert.NoError(t, err)
}

// RED: Test ValidateEventPatternE with empty pattern
func TestValidateEventPatternE_WithEmptyPattern_ReturnsError(t *testing.T) {
	err := ValidateEventPatternE("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event pattern cannot be empty")
}

// RED: Test ValidateEventPatternE with invalid JSON
func TestValidateEventPatternE_WithInvalidJSON_ReturnsError(t *testing.T) {
	invalidPattern := `{"source": ["aws.s3"]` // Missing closing brace

	err := ValidateEventPatternE(invalidPattern)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pattern JSON")
}

// RED: Test ValidateEventPatternE with invalid key
func TestValidateEventPatternE_WithInvalidKey_ReturnsError(t *testing.T) {
	invalidPattern := `{"invalid-key": ["value"]}`

	err := ValidateEventPatternE(invalidPattern)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pattern key")
}

// RED: Test ValidateEventPatternE with non-array field
func TestValidateEventPatternE_WithNonArrayField_ReturnsError(t *testing.T) {
	invalidPattern := `{"source": "aws.s3"}` // Should be array

	err := ValidateEventPatternE(invalidPattern)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an array")
}

// RED: Test ValidateEventPatternE with invalid detail structure
func TestValidateEventPatternE_WithInvalidDetailStructure_ReturnsError(t *testing.T) {
	invalidPattern := `{"detail": ["not-an-object"]}`

	err := ValidateEventPatternE(invalidPattern)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "detail must be an object")
}

// RED: Test TestEventPattern with matching pattern
func TestTestEventPattern_WithMatchingEvent_ReturnsTrue(t *testing.T) {
	pattern := `{"source": ["aws.s3"]}`
	event := map[string]interface{}{
		"source": "aws.s3",
		"detail-type": "Object Created",
	}

	matches := TestEventPattern(pattern, event)

	assert.True(t, matches)
}

// RED: Test TestEventPattern with non-matching pattern
func TestTestEventPattern_WithNonMatchingEvent_ReturnsFalse(t *testing.T) {
	pattern := `{"source": ["aws.s3"]}`
	event := map[string]interface{}{
		"source": "aws.ec2",
		"detail-type": "Instance State Change",
	}

	matches := TestEventPattern(pattern, event)

	assert.False(t, matches)
}

// RED: Test TestEventPatternE with matching pattern
func TestTestEventPatternE_WithMatchingEvent_ReturnsTrue(t *testing.T) {
	pattern := `{"source": ["aws.s3"], "detail-type": ["Object Created"]}`
	event := map[string]interface{}{
		"source": "aws.s3",
		"detail-type": "Object Created",
	}

	matches, err := TestEventPatternE(pattern, event)

	assert.NoError(t, err)
	assert.True(t, matches)
}

// RED: Test TestEventPatternE with empty pattern
func TestTestEventPatternE_WithEmptyPattern_ReturnsError(t *testing.T) {
	event := map[string]interface{}{
		"source": "aws.s3",
	}

	matches, err := TestEventPatternE("", event)

	assert.Error(t, err)
	assert.False(t, matches)
	assert.Contains(t, err.Error(), "event pattern cannot be empty")
}

// RED: Test BuildCustomEvent with valid data
func TestBuildCustomEvent_WithValidData_ReturnsEvent(t *testing.T) {
	detail := map[string]interface{}{
		"key":    "value",
		"number": 123,
	}

	event := BuildCustomEvent("test.service", "Test Event", detail)

	assert.Equal(t, "test.service", event.Source)
	assert.Equal(t, "Test Event", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)
	
	// Verify detail is valid JSON
	var parsedDetail map[string]interface{}
	err := json.Unmarshal([]byte(event.Detail), &parsedDetail)
	assert.NoError(t, err)
	assert.Equal(t, "value", parsedDetail["key"])
	assert.Equal(t, float64(123), parsedDetail["number"])
}

// RED: Test BuildCustomEvent with nil detail
func TestBuildCustomEvent_WithNilDetail_ReturnsEventWithEmptyJSON(t *testing.T) {
	event := BuildCustomEvent("test.service", "Test Event", nil)

	assert.Equal(t, "test.service", event.Source)
	assert.Equal(t, "Test Event", event.DetailType)
	assert.Equal(t, "{}", event.Detail)
}

// RED: Test BuildScheduledEvent
func TestBuildScheduledEvent_WithValidData_ReturnsScheduledEvent(t *testing.T) {
	detail := map[string]interface{}{
		"schedule": "rate(5 minutes)",
	}

	event := BuildScheduledEvent("Scheduled Event", detail)

	assert.Equal(t, "aws.events", event.Source)
	assert.Equal(t, "Scheduled Event", event.DetailType)
	assert.Contains(t, event.Detail, "schedule")
}

// RED: Test BuildEventPattern
func TestBuildEventPattern_WithValidMap_ReturnsJSONString(t *testing.T) {
	patternMap := map[string]interface{}{
		"source": []interface{}{"aws.s3"},
		"detail-type": []interface{}{"Object Created"},
	}

	pattern := BuildEventPattern(patternMap)

	var parsedPattern map[string]interface{}
	err := json.Unmarshal([]byte(pattern), &parsedPattern)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"aws.s3"}, parsedPattern["source"])
	assert.Equal(t, []interface{}{"Object Created"}, parsedPattern["detail-type"])
}

// RED: Test BuildSourcePattern
func TestBuildSourcePattern_WithValidSources_ReturnsPattern(t *testing.T) {
	pattern := BuildSourcePattern("aws.s3", "aws.ec2")

	var parsedPattern map[string]interface{}
	err := json.Unmarshal([]byte(pattern), &parsedPattern)
	assert.NoError(t, err)
	
	sources, ok := parsedPattern["source"].([]interface{})
	assert.True(t, ok)
	assert.Contains(t, sources, "aws.s3")
	assert.Contains(t, sources, "aws.ec2")
}

// RED: Test BuildDetailTypePattern
func TestBuildDetailTypePattern_WithValidTypes_ReturnsPattern(t *testing.T) {
	pattern := BuildDetailTypePattern("Object Created", "Object Deleted")

	var parsedPattern map[string]interface{}
	err := json.Unmarshal([]byte(pattern), &parsedPattern)
	assert.NoError(t, err)
	
	types, ok := parsedPattern["detail-type"].([]interface{})
	assert.True(t, ok)
	assert.Contains(t, types, "Object Created")
	assert.Contains(t, types, "Object Deleted")
}

// RED: Test BuildNumericPattern
func TestBuildNumericPattern_WithValidCondition_ReturnsPattern(t *testing.T) {
	pattern := BuildNumericPattern("detail.amount", ">", 100.0)

	var parsedPattern map[string]interface{}
	err := json.Unmarshal([]byte(pattern), &parsedPattern)
	assert.NoError(t, err)
	
	detail, ok := parsedPattern["detail"].(map[string]interface{})
	assert.True(t, ok)
	
	amount, ok := detail["amount"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, amount, 1)
	
	condition, ok := amount[0].(map[string]interface{})
	assert.True(t, ok)
	
	numeric, ok := condition["numeric"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, ">", numeric[0])
	assert.Equal(t, 100.0, numeric[1])
}

// RED: Test BuildExistsPattern
func TestBuildExistsPattern_WithValidCondition_ReturnsPattern(t *testing.T) {
	pattern := BuildExistsPattern("detail.field", true)

	var parsedPattern map[string]interface{}
	err := json.Unmarshal([]byte(pattern), &parsedPattern)
	assert.NoError(t, err)
	
	detail, ok := parsedPattern["detail"].(map[string]interface{})
	assert.True(t, ok)
	
	field, ok := detail["field"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, field, 1)
	
	condition, ok := field[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, condition["exists"])
}

// RED: Test BuildPrefixPattern
func TestBuildPrefixPattern_WithValidPrefix_ReturnsPattern(t *testing.T) {
	pattern := BuildPrefixPattern("detail.message", "ERROR:")

	var parsedPattern map[string]interface{}
	err := json.Unmarshal([]byte(pattern), &parsedPattern)
	assert.NoError(t, err)
	
	detail, ok := parsedPattern["detail"].(map[string]interface{})
	assert.True(t, ok)
	
	message, ok := detail["message"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, message, 1)
	
	condition, ok := message[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "ERROR:", condition["prefix"])
}

// RED: Test PatternBuilder functionality
func TestPatternBuilder_WithAllMethods_BuildsComplexPattern(t *testing.T) {
	builder := NewPatternBuilder()
	
	pattern := builder.
		AddSource("aws.s3", "aws.ec2").
		AddDetailType("Object Created", "Instance State Change").
		AddNumericCondition("detail.amount", ">", 100.0).
		AddExistsCondition("detail.user", true).
		Build()

	var parsedPattern map[string]interface{}
	err := json.Unmarshal([]byte(pattern), &parsedPattern)
	assert.NoError(t, err)
	
	// Verify source
	sources, ok := parsedPattern["source"].([]interface{})
	assert.True(t, ok)
	assert.Contains(t, sources, "aws.s3")
	assert.Contains(t, sources, "aws.ec2")
	
	// Verify detail-type
	detailTypes, ok := parsedPattern["detail-type"].([]interface{})
	assert.True(t, ok)
	assert.Contains(t, detailTypes, "Object Created")
	assert.Contains(t, detailTypes, "Instance State Change")
	
	// Verify detail conditions exist
	detail, ok := parsedPattern["detail"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, detail, "amount")
	assert.Contains(t, detail, "user")
}

// RED: Test OptimizeEventSizeE with valid event
func TestOptimizeEventSizeE_WithValidEvent_OptimizesJSON(t *testing.T) {
	event := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `  { "key" : "value" , "nested" : { "field" : 123 } }  `,
	}

	optimized, err := OptimizeEventSizeE(event)

	assert.NoError(t, err)
	assert.Equal(t, event.Source, optimized.Source)
	assert.Equal(t, event.DetailType, optimized.DetailType)
	
	// JSON should be compacted (no extra spaces)
	assert.NotContains(t, optimized.Detail, "  ")
	
	// Verify it's still valid JSON
	var parsedDetail map[string]interface{}
	err = json.Unmarshal([]byte(optimized.Detail), &parsedDetail)
	assert.NoError(t, err)
	assert.Equal(t, "value", parsedDetail["key"])
}

// RED: Test OptimizeEventSizeE with invalid JSON
func TestOptimizeEventSizeE_WithInvalidJSON_ReturnsError(t *testing.T) {
	event := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"invalid": json}`,
	}

	_, err := OptimizeEventSizeE(event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event detail JSON")
}

// RED: Test CalculateEventSize
func TestCalculateEventSize_WithEvent_ReturnsApproximateSize(t *testing.T) {
	event := CustomEvent{
		Source:       "test.service",     // 12 chars
		DetailType:   "Test Event",       // 10 chars
		Detail:       `{"key":"value"}`,  // 15 chars
		EventBusName: "custom-bus",       // 10 chars
	}

	size := CalculateEventSize(event)

	// Should be at least the sum of string lengths plus overhead
	expectedMinSize := 12 + 10 + 15 + 10 + 100 // 100 is the overhead
	assert.GreaterOrEqual(t, size, expectedMinSize)
}

// RED: Test ValidateEventSizeE with valid size
func TestValidateEventSizeE_WithValidSize_ReturnsNil(t *testing.T) {
	event := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"key": "value"}`,
	}

	err := ValidateEventSizeE(event)

	assert.NoError(t, err)
}

// RED: Test ValidateEventSizeE with oversized event
func TestValidateEventSizeE_WithOversizedEvent_ReturnsError(t *testing.T) {
	// Create a large detail to exceed the limit
	largeData := make([]byte, 300*1024) // 300KB
	for i := range largeData {
		largeData[i] = 'a'
	}
	
	event := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"data": "` + string(largeData) + `"}`,
	}

	err := ValidateEventSizeE(event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event size too large")
}

// RED: Test CreateEventBatches with valid input
func TestCreateEventBatches_WithValidInput_ReturnsBatches(t *testing.T) {
	events := make([]CustomEvent, 25)
	for i := range events {
		events[i] = CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{"index": ` + string(rune(i)) + `}`,
		}
	}

	batches := CreateEventBatches(events, 10)

	assert.Len(t, batches, 3) // 25 events in batches of 10 = 3 batches
	assert.Len(t, batches[0], 10)
	assert.Len(t, batches[1], 10)
	assert.Len(t, batches[2], 5)
}

// RED: Test CreateEventBatches with zero batch size
func TestCreateEventBatches_WithZeroBatchSize_UsesDefaultSize(t *testing.T) {
	events := make([]CustomEvent, 15)
	for i := range events {
		events[i] = CustomEvent{
			Source: "test.service",
		}
	}

	batches := CreateEventBatches(events, 0)

	assert.Len(t, batches, 2) // 15 events in batches of 10 = 2 batches
	assert.Len(t, batches[0], 10)
	assert.Len(t, batches[1], 5)
}

// RED: Test ValidateEventBatchE with valid batch
func TestValidateEventBatchE_WithValidBatch_ReturnsNil(t *testing.T) {
	events := []CustomEvent{
		{
			Source:     "test.service",
			DetailType: "Test Event 1",
			Detail:     `{"key": "value1"}`,
		},
		{
			Source:     "test.service",
			DetailType: "Test Event 2",
			Detail:     `{"key": "value2"}`,
		},
	}

	err := ValidateEventBatchE(events)

	assert.NoError(t, err)
}

// RED: Test ValidateEventBatchE with empty batch
func TestValidateEventBatchE_WithEmptyBatch_ReturnsError(t *testing.T) {
	events := []CustomEvent{}

	err := ValidateEventBatchE(events)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch cannot be empty")
}

// RED: Test ValidateEventBatchE with oversized batch
func TestValidateEventBatchE_WithOversizedBatch_ReturnsError(t *testing.T) {
	events := make([]CustomEvent, 15) // More than max batch size of 10
	for i := range events {
		events[i] = CustomEvent{
			Source:     "test.service",
			DetailType: "Test Event",
			Detail:     `{}`,
		}
	}

	err := ValidateEventBatchE(events)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch size exceeds maximum")
}

// RED: Test AnalyzePatternComplexity with simple pattern
func TestAnalyzePatternComplexity_WithSimplePattern_ReturnsLow(t *testing.T) {
	simplePattern := `{"source": ["aws.s3"]}`

	complexity := AnalyzePatternComplexity(simplePattern)

	assert.Equal(t, "low", complexity)
}

// RED: Test AnalyzePatternComplexity with complex pattern
func TestAnalyzePatternComplexity_WithComplexPattern_ReturnsHigh(t *testing.T) {
	complexPattern := `{
		"source": ["aws.s3", "aws.ec2", "aws.lambda"],
		"detail-type": ["Object Created", "Object Deleted", "Instance State Change"],
		"detail": {
			"field1": ["value1", "value2"],
			"field2": ["value3", "value4"],
			"nested": {
				"field3": ["value5"]
			}
		}
	}`

	complexity := AnalyzePatternComplexity(complexPattern)

	assert.Equal(t, "high", complexity)
}

// RED: Test AnalyzePatternComplexity with invalid pattern
func TestAnalyzePatternComplexity_WithInvalidPattern_ReturnsInvalid(t *testing.T) {
	invalidPattern := `{"source": ["aws.s3"]` // Missing closing brace

	complexity := AnalyzePatternComplexity(invalidPattern)

	assert.Equal(t, "invalid", complexity)
}

// RED: Test EstimatePatternMatchRate with specific pattern
func TestEstimatePatternMatchRate_WithSpecificPattern_ReturnsLowRate(t *testing.T) {
	specificPattern := `{
		"source": ["aws.s3"],
		"detail-type": ["Object Created"],
		"detail": {
			"bucket": ["specific-bucket"],
			"key": ["specific-key"]
		}
	}`

	rate := EstimatePatternMatchRate(specificPattern)

	assert.Greater(t, rate, 0.0)
	assert.Less(t, rate, 1.0)
	assert.Less(t, rate, 0.5) // Should be relatively low due to specificity
}

// RED: Test EstimatePatternMatchRate with broad pattern
func TestEstimatePatternMatchRate_WithBroadPattern_ReturnsHigherRate(t *testing.T) {
	broadPattern := `{"source": ["aws.s3"]}`

	rate := EstimatePatternMatchRate(broadPattern)

	assert.Greater(t, rate, 0.0)
	assert.Less(t, rate, 1.0)
	assert.Greater(t, rate, 0.3) // Should be relatively higher
}

// RED: Test EstimatePatternMatchRate with invalid pattern
func TestEstimatePatternMatchRate_WithInvalidPattern_ReturnsZero(t *testing.T) {
	invalidPattern := `{"source": ["aws.s3"]` // Missing closing brace

	rate := EstimatePatternMatchRate(invalidPattern)

	assert.Equal(t, 0.0, rate)
}

// RED: Test toFloat64 with various numeric types
func TestToFloat64_WithVariousTypes_ConvertsCorrectly(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
		hasError bool
	}{
		{float64(123.45), 123.45, false},
		{float32(123.45), float64(float32(123.45)), false},
		{int(123), 123.0, false},
		{int32(123), 123.0, false},
		{int64(123), 123.0, false},
		{"123.45", 123.45, false},
		{"invalid", 0.0, true},
		{true, 0.0, true},
	}

	for _, test := range tests {
		result, err := toFloat64(test.input)
		
		if test.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	}
}

// Additional helper function tests for pattern matching

// RED: Test matchesEventPattern with matching event
func TestMatchesEventPattern_WithMatchingEvent_ReturnsTrue(t *testing.T) {
	pattern := map[string]interface{}{
		"source": []interface{}{"aws.s3"},
	}
	event := map[string]interface{}{
		"source": "aws.s3",
	}

	matches := matchesEventPattern(pattern, event)

	assert.True(t, matches)
}

// RED: Test matchesEventPattern with non-matching event
func TestMatchesEventPattern_WithNonMatchingEvent_ReturnsFalse(t *testing.T) {
	pattern := map[string]interface{}{
		"source": []interface{}{"aws.s3"},
	}
	event := map[string]interface{}{
		"source": "aws.ec2",
	}

	matches := matchesEventPattern(pattern, event)

	assert.False(t, matches)
}

// RED: Test matchesNumericPattern with greater than comparison
func TestMatchesNumericPattern_WithGreaterThan_ReturnsCorrectResult(t *testing.T) {
	numericPattern := []interface{}{">", 100.0}
	
	assert.True(t, matchesNumericPattern(numericPattern, 150.0))
	assert.False(t, matchesNumericPattern(numericPattern, 50.0))
	assert.False(t, matchesNumericPattern(numericPattern, 100.0))
}

// RED: Test matchesNumericPattern with equals comparison
func TestMatchesNumericPattern_WithEquals_ReturnsCorrectResult(t *testing.T) {
	numericPattern := []interface{}{"=", 100.0}
	
	assert.True(t, matchesNumericPattern(numericPattern, 100.0))
	assert.False(t, matchesNumericPattern(numericPattern, 150.0))
}

// RED: Test matchesSpecialPattern with exists condition
func TestMatchesSpecialPattern_WithExistsCondition_ReturnsCorrectResult(t *testing.T) {
	existsPattern := map[string]interface{}{
		"exists": true,
	}
	
	assert.True(t, matchesSpecialPattern(existsPattern, "some value"))
	assert.False(t, matchesSpecialPattern(existsPattern, nil))
	
	existsPattern["exists"] = false
	assert.False(t, matchesSpecialPattern(existsPattern, "some value"))
	assert.True(t, matchesSpecialPattern(existsPattern, nil))
}

// RED: Test matchesSpecialPattern with prefix condition
func TestMatchesSpecialPattern_WithPrefixCondition_ReturnsCorrectResult(t *testing.T) {
	prefixPattern := map[string]interface{}{
		"prefix": "ERROR:",
	}
	
	assert.True(t, matchesSpecialPattern(prefixPattern, "ERROR: Something went wrong"))
	assert.False(t, matchesSpecialPattern(prefixPattern, "INFO: Everything is fine"))
	assert.False(t, matchesSpecialPattern(prefixPattern, 123)) // Non-string value
}