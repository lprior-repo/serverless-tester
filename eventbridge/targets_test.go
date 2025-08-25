package eventbridge

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPutTargetsE_Success(t *testing.T) {
	// Test successful target addition
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	targets := []TargetConfig{
		{
			ID:      "1",
			Arn:     "arn:aws:lambda:us-east-1:123456789012:function:MyFunction",
			RoleArn: "arn:aws:iam::123456789012:role/EventBridgeRole",
			Input:   `{"key": "value"}`,
		},
		{
			ID:        "2",
			Arn:       "arn:aws:sqs:us-east-1:123456789012:MyQueue",
			InputPath: "$.detail",
		},
	}

	// Mock successful response
	mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
		return *input.Rule == "test-rule" &&
			*input.EventBusName == "default" &&
			len(input.Targets) == 2 &&
			*input.Targets[0].Id == "1" &&
			*input.Targets[0].RoleArn == "arn:aws:iam::123456789012:role/EventBridgeRole"
	}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil)

	err := PutTargetsE(ctx, "test-rule", "default", targets)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutTargetsE_EmptyRuleName(t *testing.T) {
	// Test validation for empty rule name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	targets := []TargetConfig{
		{ID: "1", Arn: "arn:aws:lambda:us-east-1:123456789012:function:MyFunction"},
	}

	err := PutTargetsE(ctx, "", "default", targets)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
}

func TestPutTargetsE_EmptyTargets(t *testing.T) {
	// Test validation for empty targets
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := PutTargetsE(ctx, "test-rule", "default", []TargetConfig{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one target must be specified")
}

func TestPutTargetsE_InvalidTarget(t *testing.T) {
	// Test validation for target missing ID or ARN
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	targets := []TargetConfig{
		{
			ID: "1",
			// Missing ARN
		},
	}

	err := PutTargetsE(ctx, "test-rule", "default", targets)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target ID and ARN are required")
}

func TestPutTargetsE_DefaultEventBus(t *testing.T) {
	// Test with default event bus
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	targets := []TargetConfig{
		{ID: "1", Arn: "arn:aws:lambda:us-east-1:123456789012:function:MyFunction"},
	}

	// Mock successful response
	mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
		return *input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil)

	err := PutTargetsE(ctx, "test-rule", "", targets) // Empty event bus should default

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutTargetsE_WithInputTransformer(t *testing.T) {
	// Test target with input transformer
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	inputTemplate := `{"event_time": "<timestamp>", "event_source": "<source>"}`
	inputTransformer := &types.InputTransformer{
		InputPathsMap: map[string]string{
			"timestamp": "$.time",
			"source":    "$.source",
		},
		InputTemplate: &inputTemplate,
	}

	targets := []TargetConfig{
		{
			ID:               "1",
			Arn:              "arn:aws:lambda:us-east-1:123456789012:function:MyFunction",
			InputTransformer: inputTransformer,
		},
	}

	// Mock successful response
	mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
		return input.Targets[0].InputTransformer != nil &&
			input.Targets[0].InputTransformer.InputTemplate != nil &&
			*input.Targets[0].InputTransformer.InputTemplate == inputTemplate
	}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil)

	err := PutTargetsE(ctx, "test-rule", "default", targets)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutTargetsE_WithDeadLetterConfig(t *testing.T) {
	// Test target with dead letter queue config
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	deadLetterConfig := &types.DeadLetterConfig{
		Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:DeadLetterQueue"),
	}

	targets := []TargetConfig{
		{
			ID:               "1",
			Arn:              "arn:aws:lambda:us-east-1:123456789012:function:MyFunction",
			DeadLetterConfig: deadLetterConfig,
		},
	}

	// Mock successful response
	mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
		return input.Targets[0].DeadLetterConfig != nil &&
			*input.Targets[0].DeadLetterConfig.Arn == "arn:aws:sqs:us-east-1:123456789012:DeadLetterQueue"
	}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil)

	err := PutTargetsE(ctx, "test-rule", "default", targets)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutTargetsE_WithRetryPolicy(t *testing.T) {
	// Test target with retry policy
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	retryPolicy := &types.RetryPolicy{
		MaximumRetryAttempts:      aws.Int32(3),
		MaximumEventAgeInSeconds: aws.Int32(300),
	}

	targets := []TargetConfig{
		{
			ID:          "1",
			Arn:         "arn:aws:lambda:us-east-1:123456789012:function:MyFunction",
			RetryPolicy: retryPolicy,
		},
	}

	// Mock successful response
	mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
		return input.Targets[0].RetryPolicy != nil &&
			*input.Targets[0].RetryPolicy.MaximumRetryAttempts == 3
	}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.PutTargetsResultEntry{},
	}, nil)

	err := PutTargetsE(ctx, "test-rule", "default", targets)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutTargetsE_PartialFailure(t *testing.T) {
	// Test partial failure scenario
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	targets := []TargetConfig{
		{ID: "1", Arn: "arn:aws:lambda:us-east-1:123456789012:function:MyFunction"},
		{ID: "2", Arn: "arn:aws:sqs:us-east-1:123456789012:MyQueue"},
	}

	// Mock partial failure response
	mockClient.On("PutTargets", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.PutTargetsOutput{
		FailedEntryCount: 1,
		FailedEntries: []types.PutTargetsResultEntry{
			{
				TargetId:     aws.String("2"),
				ErrorCode:    aws.String("ValidationException"),
				ErrorMessage: aws.String("Invalid target ARN"),
			},
		},
	}, nil)

	err := PutTargetsE(ctx, "test-rule", "default", targets)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add 1 targets")
	assert.Contains(t, err.Error(), "ID=2")
	assert.Contains(t, err.Error(), "Invalid target ARN")
	mockClient.AssertExpectations(t)
}

func TestPutTargetsE_ServiceError(t *testing.T) {
	// Test AWS service error
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	targets := []TargetConfig{
		{ID: "1", Arn: "arn:aws:lambda:us-east-1:123456789012:function:MyFunction"},
	}

	serviceErr := errors.New("service error")
	mockClient.On("PutTargets", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	err := PutTargetsE(ctx, "test-rule", "default", targets)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to put targets after 3 attempts")
	mockClient.AssertExpectations(t)
}

func TestRemoveTargetsE_Success(t *testing.T) {
	// Test successful target removal
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	targetIDs := []string{"1", "2"}

	// Mock successful response
	mockClient.On("RemoveTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
		return *input.Rule == "test-rule" &&
			*input.EventBusName == "default" &&
			len(input.Ids) == 2 &&
			input.Ids[0] == "1" &&
			input.Ids[1] == "2"
	}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.RemoveTargetsResultEntry{},
	}, nil)

	err := RemoveTargetsE(ctx, "test-rule", "default", targetIDs)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRemoveTargetsE_EmptyRuleName(t *testing.T) {
	// Test validation for empty rule name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := RemoveTargetsE(ctx, "", "default", []string{"1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
}

func TestRemoveTargetsE_EmptyTargetIDs(t *testing.T) {
	// Test validation for empty target IDs
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := RemoveTargetsE(ctx, "test-rule", "default", []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one target ID must be specified")
}

func TestRemoveTargetsE_PartialFailure(t *testing.T) {
	// Test partial failure scenario
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	targetIDs := []string{"1", "2"}

	// Mock partial failure response
	mockClient.On("RemoveTargets", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
		FailedEntryCount: 1,
		FailedEntries: []types.RemoveTargetsResultEntry{
			{
				TargetId:     aws.String("1"),
				ErrorCode:    aws.String("ResourceNotFoundException"),
				ErrorMessage: aws.String("Target not found"),
			},
		},
	}, nil)

	err := RemoveTargetsE(ctx, "test-rule", "default", targetIDs)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove 1 targets")
	assert.Contains(t, err.Error(), "ID=1")
	assert.Contains(t, err.Error(), "Target not found")
	mockClient.AssertExpectations(t)
}

func TestListTargetsForRuleE_Success(t *testing.T) {
	// Test successful target listing
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	expectedResponse := &eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{
				Id:        aws.String("1"),
				Arn:       aws.String("arn:aws:lambda:us-east-1:123456789012:function:MyFunction"),
				RoleArn:   aws.String("arn:aws:iam::123456789012:role/EventBridgeRole"),
				Input:     aws.String(`{"key": "value"}`),
				InputPath: aws.String("$.detail"),
			},
			{
				Id:  aws.String("2"),
				Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:MyQueue"),
			},
		},
	}

	mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
		return *input.Rule == "test-rule" && *input.EventBusName == "default"
	}), mock.Anything).Return(expectedResponse, nil)

	results, err := ListTargetsForRuleE(ctx, "test-rule", "default")

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "1", results[0].ID)
	assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:MyFunction", results[0].Arn)
	assert.Equal(t, "arn:aws:iam::123456789012:role/EventBridgeRole", results[0].RoleArn)
	assert.Equal(t, `{"key": "value"}`, results[0].Input)
	assert.Equal(t, "$.detail", results[0].InputPath)
	assert.Equal(t, "2", results[1].ID)
	assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:MyQueue", results[1].Arn)
	mockClient.AssertExpectations(t)
}

func TestListTargetsForRuleE_EmptyRuleName(t *testing.T) {
	// Test validation for empty rule name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	results, err := ListTargetsForRuleE(ctx, "", "default")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
	assert.Nil(t, results)
}

func TestListTargetsForRuleE_DefaultEventBus(t *testing.T) {
	// Test with default event bus
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
		return *input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)

	results, err := ListTargetsForRuleE(ctx, "test-rule", "") // Empty should default

	require.NoError(t, err)
	assert.Empty(t, results)
	mockClient.AssertExpectations(t)
}

func TestRemoveAllTargetsE_Success(t *testing.T) {
	// Test successful removal of all targets
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock ListTargetsByRule
	mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
		return *input.Rule == "test-rule"
	}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("1"), Arn: aws.String("arn:aws:lambda:us-east-1:123456789012:function:MyFunction")},
			{Id: aws.String("2"), Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:MyQueue")},
		},
	}, nil)

	// Mock RemoveTargets
	mockClient.On("RemoveTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
		return len(input.Ids) == 2 && input.Ids[0] == "1" && input.Ids[1] == "2"
	}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
		FailedEntryCount: 0,
		FailedEntries:    []types.RemoveTargetsResultEntry{},
	}, nil)

	err := RemoveAllTargetsE(ctx, "test-rule", "default")

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRemoveAllTargetsE_NoTargets(t *testing.T) {
	// Test with no existing targets
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock ListTargetsByRule returning empty
	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)

	err := RemoveAllTargetsE(ctx, "test-rule", "default")

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutTargets_PanicOnError(t *testing.T) {
	// Test that PutTargets panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	targets := []TargetConfig{} // Empty targets should cause error

	PutTargets(ctx, "test-rule", "default", targets)

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestRemoveTargets_PanicOnError(t *testing.T) {
	// Test that RemoveTargets panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	RemoveTargets(ctx, "", "default", []string{"1"}) // Empty rule name should cause error

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestRemoveAllTargets_PanicOnError(t *testing.T) {
	// Test that RemoveAllTargets panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: mockClient,
	}

	serviceErr := errors.New("service error")
	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	RemoveAllTargets(ctx, "test-rule", "default")

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestListTargetsForRule_PanicOnError(t *testing.T) {
	// Test that ListTargetsForRule panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	result := ListTargetsForRule(ctx, "", "default") // Empty rule name should cause error

	assert.Nil(t, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}