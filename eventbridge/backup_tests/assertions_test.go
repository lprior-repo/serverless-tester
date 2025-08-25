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

func TestAssertRuleExistsE_Success(t *testing.T) {
	// Test successful rule existence assertion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeRuleInput) bool {
		return *input.Name == "test-rule"
	}), mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name: aws.String("test-rule"),
		Arn:  aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
	}, nil)

	err := AssertRuleExistsE(ctx, "test-rule", "default")
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAssertRuleExistsE_RuleNotFound(t *testing.T) {
	// Test rule not found
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("rule not found"))

	err := AssertRuleExistsE(ctx, "nonexistent-rule", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule 'nonexistent-rule' does not exist")
	mockClient.AssertExpectations(t)
}

func TestAssertRuleEnabledE_Success(t *testing.T) {
	// Test successful enabled rule assertion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:  aws.String("test-rule"),
		State: types.RuleStateEnabled,
	}, nil)

	err := AssertRuleEnabledE(ctx, "test-rule", "default")
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAssertRuleEnabledE_RuleDisabled(t *testing.T) {
	// Test rule is disabled when expecting enabled
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:  aws.String("test-rule"),
		State: types.RuleStateDisabled,
	}, nil)

	err := AssertRuleEnabledE(ctx, "test-rule", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not enabled")
	assert.Contains(t, err.Error(), "DISABLED")
	mockClient.AssertExpectations(t)
}

func TestAssertRuleDisabledE_Success(t *testing.T) {
	// Test successful disabled rule assertion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:  aws.String("test-rule"),
		State: types.RuleStateDisabled,
	}, nil)

	err := AssertRuleDisabledE(ctx, "test-rule", "default")
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAssertRuleDisabledE_RuleEnabled(t *testing.T) {
	// Test rule is enabled when expecting disabled
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:  aws.String("test-rule"),
		State: types.RuleStateEnabled,
	}, nil)

	err := AssertRuleDisabledE(ctx, "test-rule", "default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not disabled")
	assert.Contains(t, err.Error(), "ENABLED")
	mockClient.AssertExpectations(t)
}

func TestAssertTargetExistsE_Success(t *testing.T) {
	// Test successful target existence assertion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("target-1"), Arn: aws.String("arn:aws:lambda:us-east-1:123:function:func1")},
			{Id: aws.String("target-2"), Arn: aws.String("arn:aws:lambda:us-east-1:123:function:func2")},
		},
	}, nil)

	err := AssertTargetExistsE(ctx, "test-rule", "default", "target-1")
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAssertTargetExistsE_TargetNotFound(t *testing.T) {
	// Test target not found
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("target-1"), Arn: aws.String("arn:aws:lambda:us-east-1:123:function:func1")},
		},
	}, nil)

	err := AssertTargetExistsE(ctx, "test-rule", "default", "target-nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target 'target-nonexistent' not found")
	mockClient.AssertExpectations(t)
}

func TestAssertTargetCountE_Success(t *testing.T) {
	// Test successful target count assertion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("target-1")},
			{Id: aws.String("target-2")},
		},
	}, nil)

	err := AssertTargetCountE(ctx, "test-rule", "default", 2)
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAssertTargetCountE_WrongCount(t *testing.T) {
	// Test wrong target count
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("target-1")},
		},
	}, nil)

	err := AssertTargetCountE(ctx, "test-rule", "default", 3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule 'test-rule' has 1 targets, expected 3")
	mockClient.AssertExpectations(t)
}

func TestAssertEventBusExistsE_Success(t *testing.T) {
	// Test successful event bus existence assertion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeEventBusInput) bool {
		return *input.Name == "custom-bus"
	}), mock.Anything).Return(&eventbridge.DescribeEventBusOutput{
		Name: aws.String("custom-bus"),
		Arn:  aws.String("arn:aws:events:us-east-1:123456789012:event-bus/custom-bus"),
	}, nil)

	err := AssertEventBusExistsE(ctx, "custom-bus")
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAssertEventBusExistsE_BusNotFound(t *testing.T) {
	// Test event bus not found
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeEventBus", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("event bus not found"))

	err := AssertEventBusExistsE(ctx, "nonexistent-bus")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event bus 'nonexistent-bus' does not exist")
	mockClient.AssertExpectations(t)
}

func TestAssertPatternMatchesE_Success(t *testing.T) {
	// Test successful pattern matching assertion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	pattern := `{"source": ["test.service"]}`
	event := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"key": "value"}`,
	}

	mockClient.On("TestEventPattern", mock.Anything, mock.MatchedBy(func(input *eventbridge.TestEventPatternInput) bool {
		return *input.EventPattern == pattern && *input.Event == event.Detail
	}), mock.Anything).Return(&eventbridge.TestEventPatternOutput{
		Result: true,
	}, nil)

	err := AssertPatternMatchesE(ctx, event, pattern)
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAssertPatternMatchesE_NoMatch(t *testing.T) {
	// Test pattern doesn't match event
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	pattern := `{"source": ["different.service"]}`
	event := CustomEvent{
		Source:     "test.service",
		DetailType: "Test Event",
		Detail:     `{"key": "value"}`,
	}

	mockClient.On("TestEventPattern", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.TestEventPatternOutput{
		Result: false,
	}, nil)

	err := AssertPatternMatchesE(ctx, event, pattern)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event does not match pattern")
	mockClient.AssertExpectations(t)
}

func TestAssertEventSentE_Success(t *testing.T) {
	// Test successful event sent assertion
	result := PutEventResult{
		EventID: "event-123",
		Success: true,
	}

	err := AssertEventSentE(nil, result)
	require.NoError(t, err)
}

func TestAssertEventSentE_EventFailed(t *testing.T) {
	// Test event failed assertion
	result := PutEventResult{
		Success:      false,
		ErrorCode:    "ValidationException",
		ErrorMessage: "Invalid event",
	}

	err := AssertEventSentE(nil, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event was not successfully sent")
	assert.Contains(t, err.Error(), "ValidationException")
	assert.Contains(t, err.Error(), "Invalid event")
}

func TestAssertEventSentE_EmptyEventID(t *testing.T) {
	// Test event with empty ID
	result := PutEventResult{
		EventID: "",
		Success: true,
	}

	err := AssertEventSentE(nil, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event ID is empty")
}

func TestAssertEventBatchSentE_Success(t *testing.T) {
	// Test successful event batch sent assertion
	result := PutEventsResult{
		FailedEntryCount: 0,
		Entries: []PutEventResult{
			{EventID: "event-1", Success: true},
			{EventID: "event-2", Success: true},
		},
	}

	err := AssertEventBatchSentE(nil, result)
	require.NoError(t, err)
}

func TestAssertEventBatchSentE_SomeFailures(t *testing.T) {
	// Test batch with some failures
	result := PutEventsResult{
		FailedEntryCount: 1,
		Entries: []PutEventResult{
			{EventID: "event-1", Success: true},
			{Success: false, ErrorCode: "ValidationException", ErrorMessage: "Invalid"},
		},
	}

	err := AssertEventBatchSentE(nil, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "1 out of 2 events failed")
	assert.Contains(t, err.Error(), "ValidationException")
}

func TestAssertEventBatchSentE_AllFailures(t *testing.T) {
	// Test batch with all failures
	result := PutEventsResult{
		FailedEntryCount: 2,
		Entries: []PutEventResult{
			{Success: false, ErrorCode: "ValidationException"},
			{Success: false, ErrorCode: "ThrottlingException"},
		},
	}

	err := AssertEventBatchSentE(nil, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "2 out of 2 events failed")
}

func TestAssertRuleHasPatternE_Success(t *testing.T) {
	// Test successful rule pattern assertion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	expectedPattern := `{"source": ["test.service"]}`
	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:         aws.String("test-rule"),
		EventPattern: aws.String(expectedPattern),
	}, nil)

	err := AssertRuleHasPatternE(ctx, "test-rule", "default", expectedPattern)
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAssertRuleHasPatternE_DifferentPattern(t *testing.T) {
	// Test rule has different pattern
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	actualPattern := `{"source": ["other.service"]}`
	expectedPattern := `{"source": ["test.service"]}`

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:         aws.String("test-rule"),
		EventPattern: aws.String(actualPattern),
	}, nil)

	err := AssertRuleHasPatternE(ctx, "test-rule", "default", expectedPattern)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule has different pattern")
	assert.Contains(t, err.Error(), actualPattern)
	assert.Contains(t, err.Error(), expectedPattern)
	mockClient.AssertExpectations(t)
}

func TestAssertRuleHasPatternE_NoPattern(t *testing.T) {
	// Test rule has no pattern (scheduled rule)
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	expectedPattern := `{"source": ["test.service"]}`
	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:               aws.String("test-rule"),
		ScheduleExpression: aws.String("rate(5 minutes)"),
		// No EventPattern field
	}, nil)

	err := AssertRuleHasPatternE(ctx, "test-rule", "default", expectedPattern)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule has no event pattern")
	mockClient.AssertExpectations(t)
}

func TestWaitForReplayCompletionE_AlreadyCompleted(t *testing.T) {
	// Test replay already completed
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeReplay", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeReplayOutput{
		ReplayName: aws.String("test-replay"),
		State:      types.ReplayStateCompleted,
	}, nil)

	err := WaitForReplayCompletionE(ctx, "test-replay", 1*time.Second)
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestWaitForArchiveStateE_AlreadyCorrectState(t *testing.T) {
	// Test archive already in desired state
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	mockClient.On("DescribeArchive", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeArchiveOutput{
		ArchiveName: aws.String("test-archive"),
		State:       types.ArchiveStateEnabled,
	}, nil)

	err := WaitForArchiveStateE(ctx, "test-archive", types.ArchiveStateEnabled, 1*time.Second)
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// Test panic-on-error functions (Terratest pattern)
func TestAssertRuleExists_PanicOnError(t *testing.T) {
	// Test that AssertRuleExists panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("rule not found"))

	AssertRuleExists(ctx, "nonexistent-rule", "default")

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestAssertRuleEnabled_PanicOnError(t *testing.T) {
	// Test that AssertRuleEnabled panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		State: types.RuleStateDisabled,
	}, nil)

	AssertRuleEnabled(ctx, "test-rule", "default")

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestAssertRuleDisabled_PanicOnError(t *testing.T) {
	// Test that AssertRuleDisabled panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		State: types.RuleStateEnabled,
	}, nil)

	AssertRuleDisabled(ctx, "test-rule", "default")

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestAssertTargetExists_PanicOnError(t *testing.T) {
	// Test that AssertTargetExists panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)

	AssertTargetExists(ctx, "test-rule", "default", "nonexistent-target")

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestAssertTargetCount_PanicOnError(t *testing.T) {
	// Test that AssertTargetCount panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("target-1")},
		},
	}, nil)

	AssertTargetCount(ctx, "test-rule", "default", 2) // Expect 2 but only 1 exists

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestAssertEventBusExists_PanicOnError(t *testing.T) {
	// Test that AssertEventBusExists panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	mockClient.On("DescribeEventBus", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("event bus not found"))

	AssertEventBusExists(ctx, "nonexistent-bus")

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestAssertPatternMatches_PanicOnError(t *testing.T) {
	// Test that AssertPatternMatches panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	mockClient.On("TestEventPattern", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.TestEventPatternOutput{
		Result: false,
	}, nil)

	AssertPatternMatches(ctx, CustomEvent{}, `{"source": ["test.service"]}`)

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestAssertEventSent_PanicOnError(t *testing.T) {
	// Test that AssertEventSent panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{T: mockT}

	result := PutEventResult{Success: false, ErrorCode: "ValidationException"}

	AssertEventSent(ctx, result)

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestAssertEventBatchSent_PanicOnError(t *testing.T) {
	// Test that AssertEventBatchSent panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{T: mockT}

	result := PutEventsResult{
		FailedEntryCount: 1,
		Entries: []PutEventResult{
			{Success: false, ErrorCode: "ValidationException"},
		},
	}

	AssertEventBatchSent(ctx, result)

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestAssertRuleHasPattern_PanicOnError(t *testing.T) {
	// Test that AssertRuleHasPattern panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name: aws.String("test-rule"),
		// No EventPattern - should cause error
	}, nil)

	AssertRuleHasPattern(ctx, "test-rule", "default", `{"source": ["test.service"]}`)

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}