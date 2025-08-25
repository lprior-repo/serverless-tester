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

// RED: Test CreateArchiveE with valid configuration
func TestCreateArchiveE_WithValidConfig_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := ArchiveConfig{
		ArchiveName:      "test-archive",
		EventSourceArn:   "arn:aws:events:us-east-1:123456789012:event-bus/default",
		Description:      "Test archive",
		EventPattern:     `{"source": ["aws.s3"]}`,
		RetentionDays:    30,
	}

	archiveArn := "arn:aws:events:us-east-1:123456789012:archive/test-archive"
	creationTime := time.Now()
	mockClient.On("CreateArchive", context.Background(), mock.MatchedBy(func(input *eventbridge.CreateArchiveInput) bool {
		return *input.ArchiveName == "test-archive" &&
			*input.EventSourceArn == "arn:aws:events:us-east-1:123456789012:event-bus/default" &&
			*input.Description == "Test archive" &&
			*input.EventPattern == `{"source": ["aws.s3"]}` &&
			*input.RetentionDays == 30
	}), mock.Anything).Return(&eventbridge.CreateArchiveOutput{
		ArchiveArn:   &archiveArn,
		State:        ArchiveStateEnabled,
		CreationTime: &creationTime,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateArchiveE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "test-archive", result.ArchiveName)
	assert.Equal(t, archiveArn, result.ArchiveArn)
	assert.Equal(t, ArchiveStateEnabled, result.State)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:event-bus/default", result.EventSourceArn)
	assert.Equal(t, &creationTime, result.CreationTime)
	mockClient.AssertExpectations(t)
}

// RED: Test CreateArchiveE with empty archive name
func TestCreateArchiveE_WithEmptyArchiveName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := ArchiveConfig{
		ArchiveName:    "", // Empty name
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
	}

	result, err := CreateArchiveE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "archive name cannot be empty")
	assert.Empty(t, result.ArchiveName)
}

// RED: Test CreateArchiveE with empty event source ARN
func TestCreateArchiveE_WithEmptyEventSourceArn_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := ArchiveConfig{
		ArchiveName:    "test-archive",
		EventSourceArn: "", // Empty ARN
	}

	result, err := CreateArchiveE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event source ARN cannot be empty")
	assert.Empty(t, result.ArchiveName)
}

// RED: Test CreateArchiveE with invalid event pattern
func TestCreateArchiveE_WithInvalidEventPattern_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := ArchiveConfig{
		ArchiveName:    "test-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
		EventPattern:   `{"source": "invalid-pattern"}`, // Invalid pattern (should be array)
	}

	result, err := CreateArchiveE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event pattern")
	assert.Empty(t, result.ArchiveName)
}

// RED: Test CreateArchiveE with minimal configuration
func TestCreateArchiveE_WithMinimalConfig_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := ArchiveConfig{
		ArchiveName:    "simple-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
	}

	archiveArn := "arn:aws:events:us-east-1:123456789012:archive/simple-archive"
	mockClient.On("CreateArchive", context.Background(), mock.MatchedBy(func(input *eventbridge.CreateArchiveInput) bool {
		return *input.ArchiveName == "simple-archive" &&
			*input.EventSourceArn == "arn:aws:events:us-east-1:123456789012:event-bus/default" &&
			input.EventPattern == nil &&
			input.RetentionDays == nil // Should be nil for indefinite retention
	}), mock.Anything).Return(&eventbridge.CreateArchiveOutput{
		ArchiveArn: &archiveArn,
		State:      ArchiveStateEnabled,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateArchiveE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "simple-archive", result.ArchiveName)
	assert.Equal(t, archiveArn, result.ArchiveArn)
	mockClient.AssertExpectations(t)
}

// RED: Test CreateArchiveE with retry on transient error
func TestCreateArchiveE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := ArchiveConfig{
		ArchiveName:    "retry-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
	}

	archiveArn := "arn:aws:events:us-east-1:123456789012:archive/retry-archive"
	// First call fails, second succeeds
	mockClient.On("CreateArchive", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.CreateArchiveOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("CreateArchive", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.CreateArchiveOutput{
		ArchiveArn: &archiveArn,
		State:      ArchiveStateEnabled,
	}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateArchiveE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "retry-archive", result.ArchiveName)
	assert.Equal(t, archiveArn, result.ArchiveArn)
	mockClient.AssertExpectations(t)
}

// RED: Test DeleteArchiveE with valid archive name
func TestDeleteArchiveE_WithValidArchiveName_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	mockClient.On("DeleteArchive", context.Background(), mock.MatchedBy(func(input *eventbridge.DeleteArchiveInput) bool {
		return *input.ArchiveName == "test-archive"
	}), mock.Anything).Return(&eventbridge.DeleteArchiveOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := DeleteArchiveE(ctx, "test-archive")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test DeleteArchiveE with empty archive name
func TestDeleteArchiveE_WithEmptyArchiveName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	err := DeleteArchiveE(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "archive name cannot be empty")
}

// RED: Test DeleteArchiveE with retry on transient error
func TestDeleteArchiveE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	// First call fails, second succeeds
	mockClient.On("DeleteArchive", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.DeleteArchiveOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("DeleteArchive", context.Background(), mock.Anything, mock.Anything).Return(
		&eventbridge.DeleteArchiveOutput{}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := DeleteArchiveE(ctx, "test-archive")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test DescribeArchiveE with valid archive name
func TestDescribeArchiveE_WithValidArchiveName_ReturnsDetails(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	archiveName := "test-archive"
	archiveArn := "arn:aws:events:us-east-1:123456789012:archive/test-archive"
	eventSourceArn := "arn:aws:events:us-east-1:123456789012:event-bus/default"
	creationTime := time.Now()

	mockClient.On("DescribeArchive", context.Background(), mock.MatchedBy(func(input *eventbridge.DescribeArchiveInput) bool {
		return *input.ArchiveName == "test-archive"
	}), mock.Anything).Return(&eventbridge.DescribeArchiveOutput{
		ArchiveName:    &archiveName,
		ArchiveArn:     &archiveArn,
		State:          ArchiveStateEnabled,
		EventSourceArn: &eventSourceArn,
		CreationTime:   &creationTime,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := DescribeArchiveE(ctx, "test-archive")

	assert.NoError(t, err)
	assert.Equal(t, "test-archive", result.ArchiveName)
	assert.Equal(t, archiveArn, result.ArchiveArn)
	assert.Equal(t, ArchiveStateEnabled, result.State)
	assert.Equal(t, eventSourceArn, result.EventSourceArn)
	assert.Equal(t, &creationTime, result.CreationTime)
	mockClient.AssertExpectations(t)
}

// RED: Test DescribeArchiveE with empty archive name
func TestDescribeArchiveE_WithEmptyArchiveName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	result, err := DescribeArchiveE(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "archive name cannot be empty")
	assert.Empty(t, result.ArchiveName)
}

// RED: Test ListArchivesE with valid event source ARN
func TestListArchivesE_WithValidEventSourceArn_ReturnsArchives(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	archiveName1 := "archive-1"
	archiveName2 := "archive-2"
	eventSourceArn := "arn:aws:events:us-east-1:123456789012:event-bus/default"
	creationTime1 := time.Now()
	creationTime2 := time.Now().Add(-time.Hour)

	mockClient.On("ListArchives", context.Background(), mock.MatchedBy(func(input *eventbridge.ListArchivesInput) bool {
		return *input.EventSourceArn == eventSourceArn
	}), mock.Anything).Return(&eventbridge.ListArchivesOutput{
		Archives: []types.Archive{
			{
				ArchiveName:    &archiveName1,
				State:          ArchiveStateEnabled,
				EventSourceArn: &eventSourceArn,
				CreationTime:   &creationTime1,
			},
			{
				ArchiveName:    &archiveName2,
				State:          ArchiveStateDisabled,
				EventSourceArn: &eventSourceArn,
				CreationTime:   &creationTime2,
			},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	results, err := ListArchivesE(ctx, eventSourceArn)

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Equal(t, "archive-1", results[0].ArchiveName)
	assert.Equal(t, ArchiveStateEnabled, results[0].State)
	assert.Equal(t, eventSourceArn, results[0].EventSourceArn)
	assert.Equal(t, &creationTime1, results[0].CreationTime)

	assert.Equal(t, "archive-2", results[1].ArchiveName)
	assert.Equal(t, ArchiveStateDisabled, results[1].State)
	assert.Equal(t, eventSourceArn, results[1].EventSourceArn)
	assert.Equal(t, &creationTime2, results[1].CreationTime)

	mockClient.AssertExpectations(t)
}

// RED: Test ListArchivesE with empty event source ARN
func TestListArchivesE_WithEmptyEventSourceArn_ReturnsAllArchives(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	archiveName := "archive-1"
	eventSourceArn := "arn:aws:events:us-east-1:123456789012:event-bus/default"

	mockClient.On("ListArchives", context.Background(), mock.MatchedBy(func(input *eventbridge.ListArchivesInput) bool {
		return input.EventSourceArn == nil // Should be nil when empty string is passed
	}), mock.Anything).Return(&eventbridge.ListArchivesOutput{
		Archives: []types.Archive{
			{
				ArchiveName:    &archiveName,
				State:          ArchiveStateEnabled,
				EventSourceArn: &eventSourceArn,
			},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	results, err := ListArchivesE(ctx, "")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "archive-1", results[0].ArchiveName)
	mockClient.AssertExpectations(t)
}

// RED: Test ListArchivesE with empty response
func TestListArchivesE_WithEmptyResponse_ReturnsEmptySlice(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	mockClient.On("ListArchives", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListArchivesOutput{
		Archives: []types.Archive{},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	results, err := ListArchivesE(ctx, "arn:aws:events:us-east-1:123456789012:event-bus/default")

	assert.NoError(t, err)
	assert.Len(t, results, 0)
	mockClient.AssertExpectations(t)
}

// RED: Test StartReplayE with valid configuration
func TestStartReplayE_WithValidConfig_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	eventStartTime := time.Now().Add(-24 * time.Hour)
	eventEndTime := time.Now()
	replayStartTime := time.Now()

	config := ReplayConfig{
		ReplayName:     "test-replay",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
		EventStartTime: eventStartTime,
		EventEndTime:   eventEndTime,
		Destination: types.ReplayDestination{
			Arn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/replay-bus"),
		},
		Description: "Test replay",
	}

	replayArn := "arn:aws:events:us-east-1:123456789012:replay/test-replay"
	mockClient.On("StartReplay", context.Background(), mock.MatchedBy(func(input *eventbridge.StartReplayInput) bool {
		return *input.ReplayName == "test-replay" &&
			*input.EventSourceArn == "arn:aws:events:us-east-1:123456789012:archive/test-archive" &&
			input.EventStartTime.Equal(eventStartTime) &&
			input.EventEndTime.Equal(eventEndTime) &&
			*input.Description == "Test replay"
	}), mock.Anything).Return(&eventbridge.StartReplayOutput{
		ReplayArn:       &replayArn,
		State:           ReplayStateStarting,
		ReplayStartTime: &replayStartTime,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := StartReplayE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "test-replay", result.ReplayName)
	assert.Equal(t, replayArn, result.ReplayArn)
	assert.Equal(t, ReplayStateStarting, result.State)
	assert.Equal(t, &eventStartTime, result.EventStartTime)
	assert.Equal(t, &eventEndTime, result.EventEndTime)
	assert.Equal(t, &replayStartTime, result.ReplayStartTime)
	mockClient.AssertExpectations(t)
}

// RED: Test StartReplayE with empty replay name
func TestStartReplayE_WithEmptyReplayName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := ReplayConfig{
		ReplayName:     "", // Empty name
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:archive/test-archive",
		EventStartTime: time.Now().Add(-24 * time.Hour),
		EventEndTime:   time.Now(),
	}

	result, err := StartReplayE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replay name cannot be empty")
	assert.Empty(t, result.ReplayName)
}

// RED: Test StartReplayE with empty event source ARN
func TestStartReplayE_WithEmptyEventSourceArn_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := ReplayConfig{
		ReplayName:     "test-replay",
		EventSourceArn: "", // Empty ARN
		EventStartTime: time.Now().Add(-24 * time.Hour),
		EventEndTime:   time.Now(),
	}

	result, err := StartReplayE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event source ARN cannot be empty")
	assert.Empty(t, result.ReplayName)
}

// RED: Test DescribeReplayE with valid replay name
func TestDescribeReplayE_WithValidReplayName_ReturnsDetails(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	replayName := "test-replay"
	replayArn := "arn:aws:events:us-east-1:123456789012:replay/test-replay"
	stateReason := "Replay completed successfully"
	eventStartTime := time.Now().Add(-24 * time.Hour)
	eventEndTime := time.Now()
	replayStartTime := time.Now().Add(-time.Hour)
	replayEndTime := time.Now()

	mockClient.On("DescribeReplay", context.Background(), mock.MatchedBy(func(input *eventbridge.DescribeReplayInput) bool {
		return *input.ReplayName == "test-replay"
	}), mock.Anything).Return(&eventbridge.DescribeReplayOutput{
		ReplayName:      &replayName,
		ReplayArn:       &replayArn,
		State:           ReplayStateCompleted,
		StateReason:     &stateReason,
		EventStartTime:  &eventStartTime,
		EventEndTime:    &eventEndTime,
		ReplayStartTime: &replayStartTime,
		ReplayEndTime:   &replayEndTime,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := DescribeReplayE(ctx, "test-replay")

	assert.NoError(t, err)
	assert.Equal(t, "test-replay", result.ReplayName)
	assert.Equal(t, replayArn, result.ReplayArn)
	assert.Equal(t, ReplayStateCompleted, result.State)
	assert.Equal(t, stateReason, result.StateReason)
	assert.Equal(t, &eventStartTime, result.EventStartTime)
	assert.Equal(t, &eventEndTime, result.EventEndTime)
	assert.Equal(t, &replayStartTime, result.ReplayStartTime)
	assert.Equal(t, &replayEndTime, result.ReplayEndTime)
	mockClient.AssertExpectations(t)
}

// RED: Test DescribeReplayE with empty replay name
func TestDescribeReplayE_WithEmptyReplayName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	result, err := DescribeReplayE(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replay name cannot be empty")
	assert.Empty(t, result.ReplayName)
}

// RED: Test CancelReplayE with valid replay name
func TestCancelReplayE_WithValidReplayName_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	mockClient.On("CancelReplay", context.Background(), mock.MatchedBy(func(input *eventbridge.CancelReplayInput) bool {
		return *input.ReplayName == "test-replay"
	}), mock.Anything).Return(&eventbridge.CancelReplayOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := CancelReplayE(ctx, "test-replay")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test CancelReplayE with empty replay name
func TestCancelReplayE_WithEmptyReplayName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	err := CancelReplayE(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replay name cannot be empty")
}

// RED: Test CancelReplayE with retry on transient error
func TestCancelReplayE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}

	// First call fails, second succeeds
	mockClient.On("CancelReplay", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.CancelReplayOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("CancelReplay", context.Background(), mock.Anything, mock.Anything).Return(
		&eventbridge.CancelReplayOutput{}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := CancelReplayE(ctx, "test-replay")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test panic versions of functions call FailNow on error
func TestCreateArchive_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}
	config := ArchiveConfig{
		ArchiveName: "", // Empty name to cause error
	}

	// This should not panic but should call FailNow
	CreateArchive(ctx, config)

	mockT.AssertExpectations(t)
}

// RED: Test DeleteArchive panic version
func TestDeleteArchive_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	DeleteArchive(ctx, "") // Empty name to cause error

	mockT.AssertExpectations(t)
}

// RED: Test DescribeArchive panic version
func TestDescribeArchive_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	DescribeArchive(ctx, "") // Empty name to cause error

	mockT.AssertExpectations(t)
}

// RED: Test ListArchives panic version
func TestListArchives_WithError_CallsFailNow(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListArchives", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.ListArchivesOutput)(nil), errors.New("test error"))

	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT, Client: mockClient}

	// This should not panic but should call FailNow
	ListArchives(ctx, "arn:aws:events:us-east-1:123456789012:event-bus/default")

	mockT.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// RED: Test StartReplay panic version
func TestStartReplay_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}
	config := ReplayConfig{
		ReplayName: "", // Empty name to cause error
	}

	// This should not panic but should call FailNow
	StartReplay(ctx, config)

	mockT.AssertExpectations(t)
}

// RED: Test DescribeReplay panic version
func TestDescribeReplay_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	DescribeReplay(ctx, "") // Empty name to cause error

	mockT.AssertExpectations(t)
}

// RED: Test CancelReplay panic version
func TestCancelReplay_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	CancelReplay(ctx, "") // Empty name to cause error

	mockT.AssertExpectations(t)
}