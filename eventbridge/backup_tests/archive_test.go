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

func TestCreateArchiveE_Success(t *testing.T) {
	// Test successful archive creation
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ArchiveConfig{
		ArchiveName:    "test-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
		Description:    "Test archive",
		RetentionDays:  30,
		EventPattern:   `{"source":["aws.ec2"]}`,
	}

	expectedOutput := &eventbridge.CreateArchiveOutput{
		ArchiveArn:   aws.String("arn:aws:events:us-east-1:123456789012:archive/test-archive"),
		State:        types.ArchiveStateEnabled,
		CreationTime: aws.Time(time.Now()),
	}

	mockClient.On("CreateArchive", mock.Anything, mock.MatchedBy(func(input *eventbridge.CreateArchiveInput) bool {
		return *input.ArchiveName == "test-archive" &&
			*input.EventSourceArn == "arn:aws:events:us-east-1:123456789012:event-bus/default" &&
			*input.Description == "Test archive" &&
			*input.RetentionDays == 30 &&
			*input.EventPattern == `{"source":["aws.ec2"]}`
	}), mock.Anything).Return(expectedOutput, nil)

	result, err := CreateArchiveE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "test-archive", result.ArchiveName)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:archive/test-archive", result.ArchiveArn)
	assert.Equal(t, types.ArchiveStateEnabled, result.State)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:event-bus/default", result.EventSourceArn)
	assert.NotNil(t, result.CreationTime)

	mockClient.AssertExpectations(t)
}

func TestCreateArchiveE_EmptyArchiveName(t *testing.T) {
	// Test validation error with empty archive name
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ArchiveConfig{
		ArchiveName:    "",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
	}

	result, err := CreateArchiveE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "archive name cannot be empty")
	assert.Equal(t, ArchiveResult{}, result)

	// Should not call AWS API
	mockClient.AssertExpectations(t)
}

func TestCreateArchiveE_EmptyEventSourceArn(t *testing.T) {
	// Test validation error with empty event source ARN
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ArchiveConfig{
		ArchiveName:    "test-archive",
		EventSourceArn: "",
	}

	result, err := CreateArchiveE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "event source ARN cannot be empty")
	assert.Equal(t, ArchiveResult{}, result)

	// Should not call AWS API
	mockClient.AssertExpectations(t)
}

func TestCreateArchiveE_InvalidEventPattern(t *testing.T) {
	// Test validation error with invalid event pattern
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ArchiveConfig{
		ArchiveName:    "test-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
		EventPattern:   "invalid-json",
	}

	result, err := CreateArchiveE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event pattern")
	assert.Equal(t, ArchiveResult{}, result)

	// Should not call AWS API
	mockClient.AssertExpectations(t)
}

func TestCreateArchiveE_WithoutOptionalFields(t *testing.T) {
	// Test successful archive creation without optional fields
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ArchiveConfig{
		ArchiveName:    "test-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
	}

	expectedOutput := &eventbridge.CreateArchiveOutput{
		ArchiveArn:   aws.String("arn:aws:events:us-east-1:123456789012:archive/test-archive"),
		State:        types.ArchiveStateEnabled,
		CreationTime: aws.Time(time.Now()),
	}

	mockClient.On("CreateArchive", mock.Anything, mock.MatchedBy(func(input *eventbridge.CreateArchiveInput) bool {
		return *input.ArchiveName == "test-archive" &&
			*input.EventSourceArn == "arn:aws:events:us-east-1:123456789012:event-bus/default" &&
			input.EventPattern == nil &&
			input.RetentionDays == nil
	}), mock.Anything).Return(expectedOutput, nil)

	result, err := CreateArchiveE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "test-archive", result.ArchiveName)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:archive/test-archive", result.ArchiveArn)

	mockClient.AssertExpectations(t)
}

func TestCreateArchiveE_RetryOnFailure(t *testing.T) {
	// Test retry mechanism on transient failure
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ArchiveConfig{
		ArchiveName:    "test-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
	}

	expectedOutput := &eventbridge.CreateArchiveOutput{
		ArchiveArn:   aws.String("arn:aws:events:us-east-1:123456789012:archive/test-archive"),
		State:        types.ArchiveStateEnabled,
		CreationTime: aws.Time(time.Now()),
	}

	// First call fails, second succeeds
	mockClient.On("CreateArchive", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, assert.AnError).Once()
	mockClient.On("CreateArchive", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil).Once()

	result, err := CreateArchiveE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "test-archive", result.ArchiveName)

	mockClient.AssertExpectations(t)
}

func TestCreateArchiveE_MaxRetriesExceeded(t *testing.T) {
	// Test failure after max retries exceeded
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ArchiveConfig{
		ArchiveName:    "test-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
	}

	// All calls fail
	mockClient.On("CreateArchive", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, assert.AnError).Times(DefaultRetryAttempts)

	result, err := CreateArchiveE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create archive after")
	assert.Equal(t, ArchiveResult{}, result)

	mockClient.AssertExpectations(t)
}

func TestCreateArchive_Success(t *testing.T) {
	// Test panic-on-error wrapper with success case
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := ArchiveConfig{
		ArchiveName:    "test-archive",
		EventSourceArn: "arn:aws:events:us-east-1:123456789012:event-bus/default",
	}

	expectedOutput := &eventbridge.CreateArchiveOutput{
		ArchiveArn:   aws.String("arn:aws:events:us-east-1:123456789012:archive/test-archive"),
		State:        types.ArchiveStateEnabled,
		CreationTime: aws.Time(time.Now()),
	}

	mockClient.On("CreateArchive", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil)

	result := CreateArchive(ctx, config)

	assert.Equal(t, "test-archive", result.ArchiveName)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:archive/test-archive", result.ArchiveArn)

	mockClient.AssertExpectations(t)
}

func TestDeleteArchiveE_Success(t *testing.T) {
	// Test successful archive deletion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	archiveName := "test-archive"
	expectedOutput := &eventbridge.DeleteArchiveOutput{}

	mockClient.On("DeleteArchive", mock.Anything, mock.MatchedBy(func(input *eventbridge.DeleteArchiveInput) bool {
		return *input.ArchiveName == "test-archive"
	}), mock.Anything).Return(expectedOutput, nil)

	err := DeleteArchiveE(ctx, archiveName)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDeleteArchiveE_EmptyArchiveName(t *testing.T) {
	// Test validation error with empty archive name
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	err := DeleteArchiveE(ctx, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "archive name cannot be empty")

	// Should not call AWS API
	mockClient.AssertExpectations(t)
}

func TestDeleteArchiveE_RetryOnFailure(t *testing.T) {
	// Test retry mechanism on transient failure
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	archiveName := "test-archive"
	expectedOutput := &eventbridge.DeleteArchiveOutput{}

	// First call fails, second succeeds
	mockClient.On("DeleteArchive", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, assert.AnError).Once()
	mockClient.On("DeleteArchive", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil).Once()

	err := DeleteArchiveE(ctx, archiveName)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDeleteArchive_Success(t *testing.T) {
	// Test panic-on-error wrapper with success case
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	archiveName := "test-archive"
	expectedOutput := &eventbridge.DeleteArchiveOutput{}

	mockClient.On("DeleteArchive", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil)

	DeleteArchive(ctx, archiveName)

	mockClient.AssertExpectations(t)
}

func TestDescribeArchiveE_Success(t *testing.T) {
	// Test successful archive description
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	archiveName := "test-archive"
	creationTime := time.Now()
	expectedOutput := &eventbridge.DescribeArchiveOutput{
		ArchiveName:    aws.String("test-archive"),
		ArchiveArn:     aws.String("arn:aws:events:us-east-1:123456789012:archive/test-archive"),
		State:          types.ArchiveStateEnabled,
		EventSourceArn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
		CreationTime:   aws.Time(creationTime),
	}

	mockClient.On("DescribeArchive", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeArchiveInput) bool {
		return *input.ArchiveName == "test-archive"
	}), mock.Anything).Return(expectedOutput, nil)

	result, err := DescribeArchiveE(ctx, archiveName)

	require.NoError(t, err)
	assert.Equal(t, "test-archive", result.ArchiveName)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:archive/test-archive", result.ArchiveArn)
	assert.Equal(t, types.ArchiveStateEnabled, result.State)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:event-bus/default", result.EventSourceArn)
	assert.Equal(t, &creationTime, result.CreationTime)

	mockClient.AssertExpectations(t)
}

func TestDescribeArchiveE_EmptyArchiveName(t *testing.T) {
	// Test validation error with empty archive name
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	result, err := DescribeArchiveE(ctx, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "archive name cannot be empty")
	assert.Equal(t, ArchiveResult{}, result)

	// Should not call AWS API
	mockClient.AssertExpectations(t)
}

func TestDescribeArchive_Success(t *testing.T) {
	// Test panic-on-error wrapper with success case
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	archiveName := "test-archive"
	creationTime := time.Now()
	expectedOutput := &eventbridge.DescribeArchiveOutput{
		ArchiveName:    aws.String("test-archive"),
		ArchiveArn:     aws.String("arn:aws:events:us-east-1:123456789012:archive/test-archive"),
		State:          types.ArchiveStateEnabled,
		EventSourceArn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
		CreationTime:   aws.Time(creationTime),
	}

	mockClient.On("DescribeArchive", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil)

	result := DescribeArchive(ctx, archiveName)

	assert.Equal(t, "test-archive", result.ArchiveName)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:archive/test-archive", result.ArchiveArn)

	mockClient.AssertExpectations(t)
}

func TestListArchivesE_Success(t *testing.T) {
	// Test successful archive listing
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	eventSourceArn := "arn:aws:events:us-east-1:123456789012:event-bus/default"
	creationTime := time.Now()
	
	expectedOutput := &eventbridge.ListArchivesOutput{
		Archives: []types.Archive{
			{
				ArchiveName:    aws.String("archive-1"),
				State:          types.ArchiveStateEnabled,
				EventSourceArn: aws.String(eventSourceArn),
				CreationTime:   aws.Time(creationTime),
			},
			{
				ArchiveName:    aws.String("archive-2"),
				State:          types.ArchiveStateDisabled,
				EventSourceArn: aws.String(eventSourceArn),
				CreationTime:   aws.Time(creationTime.Add(time.Hour)),
			},
		},
	}

	mockClient.On("ListArchives", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListArchivesInput) bool {
		return *input.EventSourceArn == eventSourceArn
	}), mock.Anything).Return(expectedOutput, nil)

	results, err := ListArchivesE(ctx, eventSourceArn)

	require.NoError(t, err)
	require.Len(t, results, 2)
	
	assert.Equal(t, "archive-1", results[0].ArchiveName)
	assert.Equal(t, types.ArchiveStateEnabled, results[0].State)
	assert.Equal(t, eventSourceArn, results[0].EventSourceArn)
	
	assert.Equal(t, "archive-2", results[1].ArchiveName)
	assert.Equal(t, types.ArchiveStateDisabled, results[1].State)

	mockClient.AssertExpectations(t)
}

func TestListArchivesE_EmptyEventSourceArn(t *testing.T) {
	// Test listing archives without event source ARN filter
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	expectedOutput := &eventbridge.ListArchivesOutput{
		Archives: []types.Archive{
			{
				ArchiveName:    aws.String("archive-1"),
				State:          types.ArchiveStateEnabled,
				EventSourceArn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
				CreationTime:   aws.Time(time.Now()),
			},
		},
	}

	mockClient.On("ListArchives", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListArchivesInput) bool {
		return input.EventSourceArn == nil
	}), mock.Anything).Return(expectedOutput, nil)

	results, err := ListArchivesE(ctx, "")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "archive-1", results[0].ArchiveName)

	mockClient.AssertExpectations(t)
}

func TestListArchives_Success(t *testing.T) {
	// Test panic-on-error wrapper with success case
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	eventSourceArn := "arn:aws:events:us-east-1:123456789012:event-bus/default"
	
	expectedOutput := &eventbridge.ListArchivesOutput{
		Archives: []types.Archive{
			{
				ArchiveName:    aws.String("archive-1"),
				State:          types.ArchiveStateEnabled,
				EventSourceArn: aws.String(eventSourceArn),
				CreationTime:   aws.Time(time.Now()),
			},
		},
	}

	mockClient.On("ListArchives", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOutput, nil)

	results := ListArchives(ctx, eventSourceArn)

	require.Len(t, results, 1)
	assert.Equal(t, "archive-1", results[0].ArchiveName)

	mockClient.AssertExpectations(t)
}