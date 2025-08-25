package eventbridge

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStartReplayE_Success(t *testing.T) {
	// Test successful replay start
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()
	replayStartTime := time.Now()
	
	config := ReplayConfig{
		ReplayName:     "test-replay",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
		EventStartTime: startTime,
		EventEndTime:   endTime,
		Destination: types.ReplayDestination{
			Arn:        aws.String("arn:aws:events:us-east-1:123456789012:event-bus/target-bus"),
			FilterArns: []string{"arn:aws:events:us-east-1:123456789012:rule/target-rule"},
		},
		Description: "Test replay",
	}

	expectedOutput := &eventbridge.StartReplayOutput{
		ReplayArn:       aws.String("arn:aws:events:us-east-1:123456789012:replay/test-replay"),
		State:           types.ReplayStateStarting,
		ReplayStartTime: aws.Time(replayStartTime),
	}

	mockClient.On("StartReplay", mock.Anything, mock.MatchedBy(func(input *eventbridge.StartReplayInput) bool {
		return *input.ReplayName == "test-replay" &&
			*input.EventSourceArn == "arn:aws:events:us-east-1:123456789012:archive/test-archive" &&
			input.EventStartTime.Equal(startTime) &&
			input.EventEndTime.Equal(endTime) &&
			*input.Destination.Arn == "arn:aws:events:us-east-1:123456789012:event-bus/target-bus" &&
			*input.Description == "Test replay"
	}), mock.Anything).Return(expectedOutput, nil)

	result, err := StartReplayE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "test-replay", result.ReplayName)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:replay/test-replay", result.ReplayArn)
	assert.Equal(t, types.ReplayStateStarting, result.State)
	assert.Equal(t, &startTime, result.EventStartTime)
	assert.Equal(t, &endTime, result.EventEndTime)
	assert.Equal(t, &replayStartTime, result.ReplayStartTime)

	mockClient.AssertExpectations(t)
}

func TestStartReplayE_EmptyReplayName(t *testing.T) {
	// Test validation error with empty replay name
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ReplayConfig{
		ReplayName:     "",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
		EventStartTime: time.Now().Add(-24 * time.Hour),
		EventEndTime:   time.Now(),
	}

	result, err := StartReplayE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "replay name cannot be empty")
	assert.Equal(t, ReplayResult{}, result)

	// Should not call AWS API
	mockClient.AssertExpectations(t)
}

func TestStartReplayE_EmptyEventSourceArn(t *testing.T) {
	// Test validation error with empty event source ARN
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ReplayConfig{
		ReplayName:     "test-replay",
		EventSourceArn: "",
		EventStartTime: time.Now().Add(-24 * time.Hour),
		EventEndTime:   time.Now(),
	}

	result, err := StartReplayE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "event source ARN cannot be empty")
	assert.Equal(t, ReplayResult{}, result)

	// Should not call AWS API
	mockClient.AssertExpectations(t)
}

func TestStartReplayE_RetryOnFailure(t *testing.T) {
	// Test retry mechanism on transient failure
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ReplayConfig{
		ReplayName:     "test-replay",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
		EventStartTime: time.Now().Add(-24 * time.Hour),
		EventEndTime:   time.Now(),
		Destination: types.ReplayDestination{
			Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/target-bus"),
		},
	}

	expectedOutput := &eventbridge.StartReplayOutput{
		ReplayArn:       aws.String("arn:aws:events:us-east-1:123456789012:replay/test-replay"),
		State:           types.ReplayStateStarting,
		ReplayStartTime: aws.Time(time.Now()),
	}

	// First call fails, second succeeds
	mockClient.On("StartReplay", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, assert.AnError).Once()
	mockClient.On("StartReplay", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil).Once()

	result, err := StartReplayE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "test-replay", result.ReplayName)

	mockClient.AssertExpectations(t)
}

func TestStartReplayE_MaxRetriesExceeded(t *testing.T) {
	// Test failure after max retries exceeded
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ReplayConfig{
		ReplayName:     "test-replay",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
		EventStartTime: time.Now().Add(-24 * time.Hour),
		EventEndTime:   time.Now(),
		Destination: types.ReplayDestination{
			Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/target-bus"),
		},
	}

	// All calls fail
	mockClient.On("StartReplay", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, assert.AnError).Times(DefaultRetryAttempts)

	result, err := StartReplayE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start replay after")
	assert.Equal(t, ReplayResult{}, result)

	mockClient.AssertExpectations(t)
}

func TestStartReplay_Success(t *testing.T) {
	// Test panic-on-error wrapper with success case
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ReplayConfig{
		ReplayName:     "test-replay",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
		EventStartTime: time.Now().Add(-24 * time.Hour),
		EventEndTime:   time.Now(),
		Destination: types.ReplayDestination{
			Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/target-bus"),
		},
	}

	expectedOutput := &eventbridge.StartReplayOutput{
		ReplayArn:       aws.String("arn:aws:events:us-east-1:123456789012:replay/test-replay"),
		State:           types.ReplayStateStarting,
		ReplayStartTime: aws.Time(time.Now()),
	}

	mockClient.On("StartReplay", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil)

	result := StartReplay(ctx, config)

	assert.Equal(t, "test-replay", result.ReplayName)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:replay/test-replay", result.ReplayArn)

	mockClient.AssertExpectations(t)
}

func TestDescribeReplayE_Success(t *testing.T) {
	// Test successful replay description
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	replayName := "test-replay"
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()
	replayStartTime := time.Now().Add(-1 * time.Hour)
	replayEndTime := time.Now().Add(-30 * time.Minute)
	
	expectedOutput := &eventbridge.DescribeReplayOutput{
		ReplayName:      aws.String("test-replay"),
		ReplayArn:       aws.String("arn:aws:events:us-east-1:123456789012:replay/test-replay"),
		State:           types.ReplayStateCompleted,
		StateReason:     aws.String("Replay completed successfully"),
		EventStartTime:  aws.Time(startTime),
		EventEndTime:    aws.Time(endTime),
		ReplayStartTime: aws.Time(replayStartTime),
		ReplayEndTime:   aws.Time(replayEndTime),
	}

	mockClient.On("DescribeReplay", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeReplayInput) bool {
		return *input.ReplayName == "test-replay"
	}), mock.Anything).Return(expectedOutput, nil)

	result, err := DescribeReplayE(ctx, replayName)

	require.NoError(t, err)
	assert.Equal(t, "test-replay", result.ReplayName)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:replay/test-replay", result.ReplayArn)
	assert.Equal(t, types.ReplayStateCompleted, result.State)
	assert.Equal(t, "Replay completed successfully", result.StateReason)
	assert.Equal(t, &startTime, result.EventStartTime)
	assert.Equal(t, &endTime, result.EventEndTime)
	assert.Equal(t, &replayStartTime, result.ReplayStartTime)
	assert.Equal(t, &replayEndTime, result.ReplayEndTime)

	mockClient.AssertExpectations(t)
}

func TestDescribeReplayE_EmptyReplayName(t *testing.T) {
	// Test validation error with empty replay name
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	result, err := DescribeReplayE(ctx, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "replay name cannot be empty")
	assert.Equal(t, ReplayResult{}, result)

	// Should not call AWS API
	mockClient.AssertExpectations(t)
}

func TestDescribeReplay_Success(t *testing.T) {
	// Test panic-on-error wrapper with success case
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	replayName := "test-replay"
	startTime := time.Now().Add(-24 * time.Hour)
	
	expectedOutput := &eventbridge.DescribeReplayOutput{
		ReplayName:      aws.String("test-replay"),
		ReplayArn:       aws.String("arn:aws:events:us-east-1:123456789012:replay/test-replay"),
		State:           types.ReplayStateCompleted,
		EventStartTime:  aws.Time(startTime),
		EventEndTime:    aws.Time(time.Now()),
		ReplayStartTime: aws.Time(time.Now().Add(-1 * time.Hour)),
		ReplayEndTime:   aws.Time(time.Now().Add(-30 * time.Minute)),
	}

	mockClient.On("DescribeReplay", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil)

	result := DescribeReplay(ctx, replayName)

	assert.Equal(t, "test-replay", result.ReplayName)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:replay/test-replay", result.ReplayArn)

	mockClient.AssertExpectations(t)
}

func TestCancelReplayE_Success(t *testing.T) {
	// Test successful replay cancellation
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	replayName := "test-replay"
	expectedOutput := &eventbridge.CancelReplayOutput{}

	mockClient.On("CancelReplay", mock.Anything, mock.MatchedBy(func(input *eventbridge.CancelReplayInput) bool {
		return *input.ReplayName == "test-replay"
	}), mock.Anything).Return(expectedOutput, nil)

	err := CancelReplayE(ctx, replayName)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCancelReplayE_EmptyReplayName(t *testing.T) {
	// Test validation error with empty replay name
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	err := CancelReplayE(ctx, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "replay name cannot be empty")

	// Should not call AWS API
	mockClient.AssertExpectations(t)
}

func TestCancelReplayE_RetryOnFailure(t *testing.T) {
	// Test retry mechanism on transient failure
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	replayName := "test-replay"
	expectedOutput := &eventbridge.CancelReplayOutput{}

	// First call fails, second succeeds
	mockClient.On("CancelReplay", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, assert.AnError).Once()
	mockClient.On("CancelReplay", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil).Once()

	err := CancelReplayE(ctx, replayName)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCancelReplay_Success(t *testing.T) {
	// Test panic-on-error wrapper with success case
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	replayName := "test-replay"
	expectedOutput := &eventbridge.CancelReplayOutput{}

	mockClient.On("CancelReplay", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil)

	CancelReplay(ctx, replayName)

	mockClient.AssertExpectations(t)
}