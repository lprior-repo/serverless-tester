package eventbridge

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Simple tests for target functions to boost coverage

func TestPutTargetsE_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutTargets", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	targets := []TargetConfig{
		{
			ID:  "target-1",
			Arn: "arn:aws:lambda:us-east-1:123:function:test",
		},
	}

	err := PutTargetsE(ctx, "test-rule", "default", targets)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutTargets_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutTargets", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	targets := []TargetConfig{
		{
			ID:  "target-1",
			Arn: "arn:aws:lambda:us-east-1:123:function:test",
		},
	}

	PutTargets(ctx, "test-rule", "default", targets)
	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}

func TestRemoveTargetsE_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("RemoveTargets", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.RemoveTargetsResultEntry{},
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	targetIDs := []string{"target-1", "target-2"}

	err := RemoveTargetsE(ctx, "test-rule", "default", targetIDs)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRemoveTargets_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("RemoveTargets", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.RemoveTargetsResultEntry{},
	}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	targetIDs := []string{"target-1"}

	RemoveTargets(ctx, "test-rule", "default", targetIDs)
	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}

func TestRemoveAllTargets_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	// Mock ListTargetsByRule call
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("target-1")},
			{Id: aws.String("target-2")},
		},
	}, nil)
	// Mock RemoveTargets call
	mockClient.On("RemoveTargets", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.RemoveTargetsResultEntry{},
	}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	RemoveAllTargets(ctx, "test-rule", "default")
	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}

func TestListTargetsForRuleE_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{
				Id:  aws.String("target-1"),
				Arn: aws.String("arn:aws:lambda:us-east-1:123:function:test1"),
			},
			{
				Id:  aws.String("target-2"),
				Arn: aws.String("arn:aws:lambda:us-east-1:123:function:test2"),
			},
		},
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	targets, err := ListTargetsForRuleE(ctx, "test-rule", "default")

	require.NoError(t, err)
	assert.Len(t, targets, 2)
	assert.Equal(t, "target-1", targets[0].ID)
	assert.Equal(t, "arn:aws:lambda:us-east-1:123:function:test1", targets[0].Arn)
	assert.Equal(t, "target-2", targets[1].ID)
	assert.Equal(t, "arn:aws:lambda:us-east-1:123:function:test2", targets[1].Arn)
	mockClient.AssertExpectations(t)
}

func TestListTargetsForRule_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{
				Id:  aws.String("target-1"),
				Arn: aws.String("arn:aws:lambda:us-east-1:123:function:test"),
			},
		},
	}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	targets := ListTargetsForRule(ctx, "test-rule", "default")

	assert.Len(t, targets, 1)
	assert.Equal(t, "target-1", targets[0].ID)
	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}