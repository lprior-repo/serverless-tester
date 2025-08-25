package eventbridge

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// RED: Test PutTargetsE with valid targets
func TestPutTargetsE_WithValidTargets_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	targets := []TargetConfig{
		{
			ID:      "target-1",
			Arn:     "arn:aws:lambda:us-east-1:123456789012:function:test1",
			RoleArn: "arn:aws:iam::123456789012:role/service-role/MyRole",
			Input:   `{"key": "value1"}`,
		},
		{
			ID:        "target-2",
			Arn:       "arn:aws:lambda:us-east-1:123456789012:function:test2",
			InputPath: "$.detail",
		},
	}

	mockClient.On("PutTargets", context.Background(), mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
		return *input.Rule == "test-rule" &&
			*input.EventBusName == DefaultEventBusName &&
			len(input.Targets) == 2 &&
			*input.Targets[0].Id == "target-1" &&
			*input.Targets[0].Arn == "arn:aws:lambda:us-east-1:123456789012:function:test1" &&
			*input.Targets[0].RoleArn == "arn:aws:iam::123456789012:role/service-role/MyRole" &&
			*input.Targets[0].Input == `{"key": "value1"}` &&
			*input.Targets[1].Id == "target-2" &&
			*input.Targets[1].Arn == "arn:aws:lambda:us-east-1:123456789012:function:test2" &&
			*input.Targets[1].InputPath == "$.detail"
	}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := PutTargetsE(ctx, "test-rule", "", targets)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test PutTargetsE with empty rule name
func TestPutTargetsE_WithEmptyRuleName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	targets := []TargetConfig{
		{ID: "target-1", Arn: "arn:aws:lambda:us-east-1:123456789012:function:test"},
	}

	err := PutTargetsE(ctx, "", DefaultEventBusName, targets)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
}

// RED: Test PutTargetsE with empty targets
func TestPutTargetsE_WithEmptyTargets_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	targets := []TargetConfig{}

	err := PutTargetsE(ctx, "test-rule", DefaultEventBusName, targets)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one target must be specified")
}

// RED: Test PutTargetsE with invalid target (missing ID)
func TestPutTargetsE_WithInvalidTarget_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	targets := []TargetConfig{
		{
			// ID missing
			Arn: "arn:aws:lambda:us-east-1:123456789012:function:test",
		},
	}

	err := PutTargetsE(ctx, "test-rule", DefaultEventBusName, targets)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target ID and ARN are required")
}

// RED: Test PutTargetsE with invalid target (missing ARN)
func TestPutTargetsE_WithInvalidTargetMissingArn_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	targets := []TargetConfig{
		{
			ID: "target-1",
			// Arn missing
		},
	}

	err := PutTargetsE(ctx, "test-rule", DefaultEventBusName, targets)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target ID and ARN are required")
}

// RED: Test PutTargetsE with failed entries
func TestPutTargetsE_WithFailedEntries_ReturnsError(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	targets := []TargetConfig{
		{ID: "target-1", Arn: "arn:aws:lambda:us-east-1:123456789012:function:test"},
	}

	targetId := "target-1"
	errorMessage := "Target validation failed"
	mockClient.On("PutTargets", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 1,
		FailedEntries: []types.PutTargetsResultEntry{
			{
				TargetId:     &targetId,
				ErrorMessage: &errorMessage,
			},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := PutTargetsE(ctx, "test-rule", DefaultEventBusName, targets)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add 1 targets")
	assert.Contains(t, err.Error(), "ID=target-1")
	assert.Contains(t, err.Error(), "Error=Target validation failed")
	mockClient.AssertExpectations(t)
}

// RED: Test PutTargetsE with retry on transient error
func TestPutTargetsE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	targets := []TargetConfig{
		{ID: "target-1", Arn: "arn:aws:lambda:us-east-1:123456789012:function:test"},
	}

	// First call fails, second succeeds
	mockClient.On("PutTargets", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.PutTargetsOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("PutTargets", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := PutTargetsE(ctx, "test-rule", DefaultEventBusName, targets)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test PutTargetsE with complex target configuration
func TestPutTargetsE_WithComplexTargetConfig_ConfiguresAllFields(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	inputTransformer := &types.InputTransformer{
		InputPathsMap: map[string]string{
			"timestamp": "$.time",
		},
		InputTemplate: aws.String(`{"event_time": "<timestamp>"}`),
	}
	
	deadLetterConfig := &types.DeadLetterConfig{
		Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:dlq"),
	}
	
	retryPolicy := &types.RetryPolicy{
		MaximumRetryAttempts:     aws.Int32(3),
		MaximumEventAgeInSeconds: aws.Int32(3600),
	}

	targets := []TargetConfig{
		{
			ID:               "complex-target",
			Arn:              "arn:aws:lambda:us-east-1:123456789012:function:test",
			RoleArn:          "arn:aws:iam::123456789012:role/EventBridgeRole",
			InputTransformer: inputTransformer,
			DeadLetterConfig: deadLetterConfig,
			RetryPolicy:      retryPolicy,
		},
	}

	mockClient.On("PutTargets", context.Background(), mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
		target := input.Targets[0]
		return *target.Id == "complex-target" &&
			*target.Arn == "arn:aws:lambda:us-east-1:123456789012:function:test" &&
			*target.RoleArn == "arn:aws:iam::123456789012:role/EventBridgeRole" &&
			target.InputTransformer != nil &&
			target.DeadLetterConfig != nil &&
			target.RetryPolicy != nil
	}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := PutTargetsE(ctx, "test-rule", DefaultEventBusName, targets)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test RemoveTargetsE with valid target IDs
func TestRemoveTargetsE_WithValidTargetIds_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	targetIDs := []string{"target-1", "target-2"}

	mockClient.On("RemoveTargets", context.Background(), mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
		return *input.Rule == "test-rule" &&
			*input.EventBusName == DefaultEventBusName &&
			len(input.Ids) == 2 &&
			input.Ids[0] == "target-1" &&
			input.Ids[1] == "target-2"
	}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.RemoveTargetsResultEntry{},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := RemoveTargetsE(ctx, "test-rule", "", targetIDs)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test RemoveTargetsE with empty rule name
func TestRemoveTargetsE_WithEmptyRuleName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	targetIDs := []string{"target-1"}

	err := RemoveTargetsE(ctx, "", DefaultEventBusName, targetIDs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
}

// RED: Test RemoveTargetsE with empty target IDs
func TestRemoveTargetsE_WithEmptyTargetIds_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	targetIDs := []string{}

	err := RemoveTargetsE(ctx, "test-rule", DefaultEventBusName, targetIDs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one target ID must be specified")
}

// RED: Test RemoveTargetsE with failed entries
func TestRemoveTargetsE_WithFailedEntries_ReturnsError(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	targetIDs := []string{"target-1"}

	targetId := "target-1"
	errorMessage := "Target not found"
	mockClient.On("RemoveTargets", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
		FailedEntryCount: 1,
		FailedEntries: []types.RemoveTargetsResultEntry{
			{
				TargetId:     &targetId,
				ErrorMessage: &errorMessage,
			},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := RemoveTargetsE(ctx, "test-rule", DefaultEventBusName, targetIDs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove 1 targets")
	assert.Contains(t, err.Error(), "ID=target-1")
	assert.Contains(t, err.Error(), "Error=Target not found")
	mockClient.AssertExpectations(t)
}

// RED: Test ListTargetsForRuleE with valid rule
func TestListTargetsForRuleE_WithValidRule_ReturnsTargets(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	targetId1 := "target-1"
	targetArn1 := "arn:aws:lambda:us-east-1:123456789012:function:test1"
	roleArn := "arn:aws:iam::123456789012:role/EventBridgeRole"
	input := `{"key": "value"}`
	inputPath := "$.detail"
	
	targetId2 := "target-2"
	targetArn2 := "arn:aws:sqs:us-east-1:123456789012:queue"

	mockClient.On("ListTargetsByRule", context.Background(), mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
		return *input.Rule == "test-rule" &&
			*input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{
				Id:        &targetId1,
				Arn:       &targetArn1,
				RoleArn:   &roleArn,
				Input:     &input,
				InputPath: &inputPath,
			},
			{
				Id:  &targetId2,
				Arn: &targetArn2,
			},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	results, err := ListTargetsForRuleE(ctx, "test-rule", "")

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	
	assert.Equal(t, "target-1", results[0].ID)
	assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:test1", results[0].Arn)
	assert.Equal(t, "arn:aws:iam::123456789012:role/EventBridgeRole", results[0].RoleArn)
	assert.Equal(t, `{"key": "value"}`, results[0].Input)
	assert.Equal(t, "$.detail", results[0].InputPath)
	
	assert.Equal(t, "target-2", results[1].ID)
	assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:queue", results[1].Arn)
	assert.Empty(t, results[1].RoleArn) // Should be empty when not set
	
	mockClient.AssertExpectations(t)
}

// RED: Test ListTargetsForRuleE with empty rule name
func TestListTargetsForRuleE_WithEmptyRuleName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	results, err := ListTargetsForRuleE(ctx, "", DefaultEventBusName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
	assert.Nil(t, results)
}

// RED: Test ListTargetsForRuleE with empty result
func TestListTargetsForRuleE_WithEmptyResult_ReturnsEmptySlice(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	results, err := ListTargetsForRuleE(ctx, "test-rule", DefaultEventBusName)

	assert.NoError(t, err)
	assert.Len(t, results, 0)
	mockClient.AssertExpectations(t)
}

// RED: Test RemoveAllTargetsE with existing targets
func TestRemoveAllTargetsE_WithExistingTargets_RemovesAllTargets(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	targetId1 := "target-1"
	targetArn1 := "arn:aws:lambda:us-east-1:123456789012:function:test1"
	targetId2 := "target-2"
	targetArn2 := "arn:aws:sqs:us-east-1:123456789012:queue"

	// Mock ListTargetsByRule
	mockClient.On("ListTargetsByRule", context.Background(), mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
		return *input.Rule == "test-rule" &&
			*input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: &targetId1, Arn: &targetArn1},
			{Id: &targetId2, Arn: &targetArn2},
		},
	}, nil)

	// Mock RemoveTargets
	mockClient.On("RemoveTargets", context.Background(), mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
		return *input.Rule == "test-rule" &&
			*input.EventBusName == DefaultEventBusName &&
			len(input.Ids) == 2 &&
			((input.Ids[0] == "target-1" && input.Ids[1] == "target-2") ||
				(input.Ids[0] == "target-2" && input.Ids[1] == "target-1"))
	}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.RemoveTargetsResultEntry{},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := RemoveAllTargetsE(ctx, "test-rule", "")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test RemoveAllTargetsE with no targets
func TestRemoveAllTargetsE_WithNoTargets_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	// Mock ListTargetsByRule to return empty result
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := RemoveAllTargetsE(ctx, "test-rule", DefaultEventBusName)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test RemoveAllTargetsE with list targets failure
func TestRemoveAllTargetsE_WithListTargetsFailure_ReturnsError(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	// Mock ListTargetsByRule to fail
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.ListTargetsByRuleOutput)(nil), errors.New("list targets failed"))

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := RemoveAllTargetsE(ctx, "test-rule", DefaultEventBusName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list targets")
	mockClient.AssertExpectations(t)
}

// RED: Test panic versions of functions call FailNow on error
func TestPutTargets_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}
	targets := []TargetConfig{} // Empty targets to cause error

	// This should not panic but should call FailNow
	PutTargets(ctx, "test-rule", DefaultEventBusName, targets)

	mockT.AssertExpectations(t)
}

// RED: Test RemoveTargets panic version
func TestRemoveTargets_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}
	targetIDs := []string{} // Empty IDs to cause error

	// This should not panic but should call FailNow
	RemoveTargets(ctx, "test-rule", DefaultEventBusName, targetIDs)

	mockT.AssertExpectations(t)
}

// RED: Test RemoveAllTargets panic version
func TestRemoveAllTargets_WithError_CallsFailNow(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.ListTargetsByRuleOutput)(nil), errors.New("test error"))

	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT, Client: mockClient}

	// This should not panic but should call FailNow
	RemoveAllTargets(ctx, "test-rule", DefaultEventBusName)

	mockT.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// RED: Test ListTargetsForRule panic version
func TestListTargetsForRule_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	ListTargetsForRule(ctx, "", DefaultEventBusName) // Empty rule name to cause error

	mockT.AssertExpectations(t)
}