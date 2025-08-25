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
	"github.com/stretchr/testify/require"
)

// TestTargetOperations_PutTargets_ComprehensiveCoverage tests all aspects of PutTargets operations
func TestTargetOperations_PutTargets_ComprehensiveCoverage(t *testing.T) {
	t.Run("PutTargets with all optional fields configured", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		// Create a target with all possible configurations
		inputTransformer := &types.InputTransformer{
			InputPathsMap: map[string]string{
				"timestamp": "$.time",
				"source":    "$.source",
				"detail":    "$.detail",
			},
			InputTemplate: aws.String(`{
				"event_time": "<timestamp>",
				"event_source": "<source>",
				"event_detail": "<detail>",
				"processed_by": "eventbridge-target"
			}`),
		}

		deadLetterConfig := &types.DeadLetterConfig{
			Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:dlq"),
		}

		retryPolicy := &types.RetryPolicy{
			MaximumRetryAttempts:      aws.Int32(5),
			MaximumEventAgeInSeconds: aws.Int32(600),
		}

		targets := []TargetConfig{
			{
				ID:               "comprehensive-target",
				Arn:              "arn:aws:lambda:us-east-1:123456789012:function:ProcessEvent",
				RoleArn:          "arn:aws:iam::123456789012:role/EventBridgeExecutionRole",
				Input:            `{"static": "data", "processed_at": "2024-01-01T00:00:00Z"}`,
				InputPath:        "$.detail.payload",
				InputTransformer: inputTransformer,
				DeadLetterConfig: deadLetterConfig,
				RetryPolicy:      retryPolicy,
			},
		}

		// Mock successful response
		mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
			target := input.Targets[0]
			return *input.Rule == "comprehensive-rule" &&
				*input.EventBusName == "custom-bus" &&
				len(input.Targets) == 1 &&
				*target.Id == "comprehensive-target" &&
				*target.Arn == "arn:aws:lambda:us-east-1:123456789012:function:ProcessEvent" &&
				*target.RoleArn == "arn:aws:iam::123456789012:role/EventBridgeExecutionRole" &&
				*target.Input == `{"static": "data", "processed_at": "2024-01-01T00:00:00Z"}` &&
				*target.InputPath == "$.detail.payload" &&
				target.InputTransformer != nil &&
				target.DeadLetterConfig != nil &&
				target.RetryPolicy != nil &&
				*target.RetryPolicy.MaximumRetryAttempts == 5 &&
				*target.RetryPolicy.MaximumEventAgeInSeconds == 600
		}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.PutTargetsResultEntry{},
		}, nil)

		err := PutTargetsE(ctx, "comprehensive-rule", "custom-bus", targets)

		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("PutTargets batch operation with mixed configurations", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-west-2",
			Client: mockClient,
		}

		targets := []TargetConfig{
			{
				ID:      "lambda-target",
				Arn:     "arn:aws:lambda:us-west-2:123456789012:function:LambdaProcessor",
				RoleArn: "arn:aws:iam::123456789012:role/LambdaRole",
				Input:   `{"type": "lambda"}`,
			},
			{
				ID:        "sqs-target",
				Arn:       "arn:aws:sqs:us-west-2:123456789012:my-queue",
				InputPath: "$.detail",
			},
			{
				ID:      "sns-target",
				Arn:     "arn:aws:sns:us-west-2:123456789012:my-topic",
				RoleArn: "arn:aws:iam::123456789012:role/SNSRole",
			},
		}

		mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
			return len(input.Targets) == 3 &&
				*input.Targets[0].Id == "lambda-target" &&
				*input.Targets[1].Id == "sqs-target" &&
				*input.Targets[2].Id == "sns-target"
		}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.PutTargetsResultEntry{},
		}, nil)

		err := PutTargetsE(ctx, "multi-target-rule", "default", targets)

		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("PutTargets with multiple partial failures", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		targets := []TargetConfig{
			{ID: "target-1", Arn: "arn:aws:lambda:us-east-1:123:function:Valid"},
			{ID: "target-2", Arn: "arn:aws:lambda:us-east-1:123:function:Invalid"},
			{ID: "target-3", Arn: "arn:aws:sqs:us-east-1:123:queue:Invalid"},
		}

		mockClient.On("PutTargets", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.PutTargetsOutput{
			FailedEntryCount: 2,
			FailedEntries: []types.PutTargetsResultEntry{
				{
					TargetId:     aws.String("target-2"),
					ErrorCode:    aws.String("ValidationException"),
					ErrorMessage: aws.String("Invalid Lambda function ARN"),
				},
				{
					TargetId:     aws.String("target-3"),
					ErrorCode:    aws.String("ResourceNotFoundException"),
					ErrorMessage: aws.String("SQS queue does not exist"),
				},
			},
		}, nil)

		err := PutTargetsE(ctx, "test-rule", "default", targets)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add 2 targets")
		assert.Contains(t, err.Error(), "ID=target-2")
		assert.Contains(t, err.Error(), "Invalid Lambda function ARN")
		assert.Contains(t, err.Error(), "ID=target-3")
		assert.Contains(t, err.Error(), "SQS queue does not exist")
		mockClient.AssertExpectations(t)
	})

	t.Run("PutTargets validation edge cases", func(t *testing.T) {
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: &MockEventBridgeClient{},
		}

		// Test empty target ID
		targets := []TargetConfig{
			{
				ID:  "", // Empty ID
				Arn: "arn:aws:lambda:us-east-1:123456789012:function:MyFunction",
			},
		}
		err := PutTargetsE(ctx, "test-rule", "default", targets)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target ID and ARN are required")

		// Test empty target ARN
		targets = []TargetConfig{
			{
				ID:  "valid-id",
				Arn: "", // Empty ARN
			},
		}
		err = PutTargetsE(ctx, "test-rule", "default", targets)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target ID and ARN are required")

		// Test nil targets slice
		err = PutTargetsE(ctx, "test-rule", "default", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one target must be specified")
	})
}

// TestTargetOperations_RemoveTargets_ComprehensiveCoverage tests all aspects of RemoveTargets operations
func TestTargetOperations_RemoveTargets_ComprehensiveCoverage(t *testing.T) {
	t.Run("RemoveTargets batch operation success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		targetIDs := []string{"target-1", "target-2", "target-3", "target-4", "target-5"}

		mockClient.On("RemoveTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
			return *input.Rule == "batch-rule" &&
				*input.EventBusName == "custom-bus" &&
				len(input.Ids) == 5 &&
				input.Ids[0] == "target-1" &&
				input.Ids[4] == "target-5"
		}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.RemoveTargetsResultEntry{},
		}, nil)

		err := RemoveTargetsE(ctx, "batch-rule", "custom-bus", targetIDs)

		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveTargets with partial failures and detailed error messages", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		targetIDs := []string{"existing-target", "missing-target-1", "missing-target-2"}

		mockClient.On("RemoveTargets", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
			FailedEntryCount: 2,
			FailedEntries: []types.RemoveTargetsResultEntry{
				{
					TargetId:     aws.String("missing-target-1"),
					ErrorCode:    aws.String("ResourceNotFoundException"),
					ErrorMessage: aws.String("Target not found for rule"),
				},
				{
					TargetId:     aws.String("missing-target-2"),
					ErrorCode:    aws.String("InvalidParameterException"),
					ErrorMessage: aws.String("Invalid target ID format"),
				},
			},
		}, nil)

		err := RemoveTargetsE(ctx, "test-rule", "default", targetIDs)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove 2 targets")
		assert.Contains(t, err.Error(), "ID=missing-target-1")
		assert.Contains(t, err.Error(), "Target not found for rule")
		assert.Contains(t, err.Error(), "ID=missing-target-2")
		assert.Contains(t, err.Error(), "Invalid target ID format")
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveTargets validation edge cases", func(t *testing.T) {
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: &MockEventBridgeClient{},
		}

		// Test with nil target IDs slice
		err := RemoveTargetsE(ctx, "test-rule", "default", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one target ID must be specified")

		// Test with empty target IDs slice  
		err = RemoveTargetsE(ctx, "test-rule", "default", []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one target ID must be specified")

		// Test with nil rule name (Go doesn't allow nil strings, so test empty string)
		err = RemoveTargetsE(ctx, "", "default", []string{"target-1"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rule name cannot be empty")
	})
}

// TestTargetOperations_ListTargets_ComprehensiveCoverage tests all aspects of ListTargets operations
func TestTargetOperations_ListTargets_ComprehensiveCoverage(t *testing.T) {
	t.Run("ListTargetsForRule with complex target configurations", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		inputTransformer := &types.InputTransformer{
			InputPathsMap: map[string]string{
				"timestamp": "$.time",
				"detail":    "$.detail",
			},
			InputTemplate: aws.String(`{"time": "<timestamp>", "data": "<detail>"}`),
		}

		deadLetterConfig := &types.DeadLetterConfig{
			Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:dlq"),
		}

		retryPolicy := &types.RetryPolicy{
			MaximumRetryAttempts:      aws.Int32(3),
			MaximumEventAgeInSeconds: aws.Int32(300),
		}

		expectedResponse := &eventbridge.ListTargetsByRuleOutput{
			Targets: []types.Target{
				{
					Id:               aws.String("lambda-target"),
					Arn:              aws.String("arn:aws:lambda:us-east-1:123456789012:function:ProcessEvent"),
					RoleArn:          aws.String("arn:aws:iam::123456789012:role/EventBridgeRole"),
					Input:            aws.String(`{"type": "lambda", "version": "1.0"}`),
					InputPath:        aws.String("$.detail.payload"),
					InputTransformer: inputTransformer,
					DeadLetterConfig: deadLetterConfig,
					RetryPolicy:      retryPolicy,
				},
				{
					Id:  aws.String("sqs-target"),
					Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:ProcessQueue"),
				},
				{
					Id:      aws.String("sns-target"),
					Arn:     aws.String("arn:aws:sns:us-east-1:123456789012:NotifyTopic"),
					RoleArn: aws.String("arn:aws:iam::123456789012:role/SNSPublishRole"),
				},
			},
		}

		mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
			return *input.Rule == "complex-rule" && *input.EventBusName == "production-bus"
		}), mock.Anything).Return(expectedResponse, nil)

		results, err := ListTargetsForRuleE(ctx, "complex-rule", "production-bus")

		require.NoError(t, err)
		require.Len(t, results, 3)

		// Validate first target (Lambda with all configurations)
		lambdaTarget := results[0]
		assert.Equal(t, "lambda-target", lambdaTarget.ID)
		assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:ProcessEvent", lambdaTarget.Arn)
		assert.Equal(t, "arn:aws:iam::123456789012:role/EventBridgeRole", lambdaTarget.RoleArn)
		assert.Equal(t, `{"type": "lambda", "version": "1.0"}`, lambdaTarget.Input)
		assert.Equal(t, "$.detail.payload", lambdaTarget.InputPath)
		require.NotNil(t, lambdaTarget.InputTransformer)
		assert.Equal(t, `{"time": "<timestamp>", "data": "<detail>"}`, *lambdaTarget.InputTransformer.InputTemplate)
		require.NotNil(t, lambdaTarget.DeadLetterConfig)
		assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:dlq", *lambdaTarget.DeadLetterConfig.Arn)
		require.NotNil(t, lambdaTarget.RetryPolicy)
		assert.Equal(t, int32(3), *lambdaTarget.RetryPolicy.MaximumRetryAttempts)
		assert.Equal(t, int32(300), *lambdaTarget.RetryPolicy.MaximumEventAgeInSeconds)

		// Validate second target (SQS with minimal config)
		sqsTarget := results[1]
		assert.Equal(t, "sqs-target", sqsTarget.ID)
		assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:ProcessQueue", sqsTarget.Arn)
		assert.Empty(t, sqsTarget.RoleArn)
		assert.Empty(t, sqsTarget.Input)
		assert.Empty(t, sqsTarget.InputPath)
		assert.Nil(t, sqsTarget.InputTransformer)
		assert.Nil(t, sqsTarget.DeadLetterConfig)
		assert.Nil(t, sqsTarget.RetryPolicy)

		// Validate third target (SNS with role)
		snsTarget := results[2]
		assert.Equal(t, "sns-target", snsTarget.ID)
		assert.Equal(t, "arn:aws:sns:us-east-1:123456789012:NotifyTopic", snsTarget.Arn)
		assert.Equal(t, "arn:aws:iam::123456789012:role/SNSPublishRole", snsTarget.RoleArn)

		mockClient.AssertExpectations(t)
	})

	t.Run("ListTargetsForRule empty response", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
			return *input.Rule == "empty-rule"
		}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
			Targets: []types.Target{},
		}, nil)

		results, err := ListTargetsForRuleE(ctx, "empty-rule", "default")

		require.NoError(t, err)
		assert.Empty(t, results)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListTargetsForRule service error with retry", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		serviceErr := errors.New("service temporarily unavailable")
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr).Times(3)

		results, err := ListTargetsForRuleE(ctx, "test-rule", "default")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list targets after 3 attempts")
		assert.Contains(t, err.Error(), "service temporarily unavailable")
		assert.Nil(t, results)
		mockClient.AssertExpectations(t)
	})
}

// TestTargetOperations_RemoveAllTargets_ComprehensiveCoverage tests RemoveAllTargets operations
func TestTargetOperations_RemoveAllTargets_ComprehensiveCoverage(t *testing.T) {
	t.Run("RemoveAllTargets with large number of targets", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		// Create 10 targets
		targets := make([]types.Target, 10)
		expectedIDs := make([]string, 10)
		for i := 0; i < 10; i++ {
			targetID := fmt.Sprintf("target-%d", i+1)
			targets[i] = types.Target{
				Id:  aws.String(targetID),
				Arn: aws.String(fmt.Sprintf("arn:aws:lambda:us-east-1:123456789012:function:Function%d", i+1)),
			}
			expectedIDs[i] = targetID
		}

		// Mock ListTargetsByRule
		mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
			return *input.Rule == "large-rule"
		}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
			Targets: targets,
		}, nil)

		// Mock RemoveTargets
		mockClient.On("RemoveTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
			return len(input.Ids) == 10 &&
				input.Ids[0] == "target-1" &&
				input.Ids[9] == "target-10"
		}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.RemoveTargetsResultEntry{},
		}, nil)

		err := RemoveAllTargetsE(ctx, "large-rule", "default")

		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveAllTargets failure during list phase", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		listErr := errors.New("access denied to list targets")
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, listErr)

		err := RemoveAllTargetsE(ctx, "test-rule", "default")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list targets")
		assert.Contains(t, err.Error(), "access denied to list targets")
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveAllTargets failure during removal phase", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		// Mock successful list
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
			Targets: []types.Target{
				{Id: aws.String("target-1"), Arn: aws.String("arn:aws:lambda:us-east-1:123:function:Test1")},
				{Id: aws.String("target-2"), Arn: aws.String("arn:aws:lambda:us-east-1:123:function:Test2")},
			},
		}, nil)

		// Mock failed removal
		removeErr := errors.New("insufficient permissions to remove targets")
		mockClient.On("RemoveTargets", mock.Anything, mock.Anything, mock.Anything).Return(nil, removeErr).Times(3)

		err := RemoveAllTargetsE(ctx, "test-rule", "default")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove targets after 3 attempts")
		assert.Contains(t, err.Error(), "insufficient permissions to remove targets")
		mockClient.AssertExpectations(t)
	})
}

// TestTargetOperations_RetryLogic_ComprehensiveCoverage tests retry mechanisms
func TestTargetOperations_RetryLogic_ComprehensiveCoverage(t *testing.T) {
	t.Run("PutTargets retry logic with eventual success", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		targets := []TargetConfig{
			{ID: "retry-target", Arn: "arn:aws:lambda:us-east-1:123456789012:function:RetryTest"},
		}

		// First two calls fail, third succeeds
		retryErr := errors.New("temporary service error")
		mockClient.On("PutTargets", mock.Anything, mock.Anything, mock.Anything).Return(nil, retryErr).Times(2)
		mockClient.On("PutTargets", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.PutTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.PutTargetsResultEntry{},
		}, nil).Once()

		start := time.Now()
		err := PutTargetsE(ctx, "retry-rule", "default", targets)
		duration := time.Since(start)

		require.NoError(t, err)
		// Should have some delay due to retries (at least 2 seconds for backoff)
		assert.True(t, duration >= 2*time.Second, "Expected retry delays")
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveTargets exhaustive retry attempts", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		targetIDs := []string{"persistent-failure"}

		persistentErr := errors.New("consistent service failure")
		mockClient.On("RemoveTargets", mock.Anything, mock.Anything, mock.Anything).Return(nil, persistentErr).Times(3)

		start := time.Now()
		err := RemoveTargetsE(ctx, "failing-rule", "default", targetIDs)
		duration := time.Since(start)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove targets after 3 attempts")
		assert.Contains(t, err.Error(), "consistent service failure")
		// Should have delay from multiple retries
		assert.True(t, duration >= 3*time.Second, "Expected multiple retry delays")
		mockClient.AssertExpectations(t)
	})

	t.Run("ListTargetsForRule retry with intermittent failures", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		intermittentErr := errors.New("network timeout")
		
		// First call fails, second succeeds
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, intermittentErr).Once()
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
			Targets: []types.Target{
				{Id: aws.String("recovered-target"), Arn: aws.String("arn:aws:lambda:us-east-1:123:function:Test")},
			},
		}, nil).Once()

		results, err := ListTargetsForRuleE(ctx, "intermittent-rule", "default")

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "recovered-target", results[0].ID)
		mockClient.AssertExpectations(t)
	})
}

// TestTargetOperations_TerratestPattern_ComprehensiveCoverage tests panic vs error patterns
func TestTargetOperations_TerratestPattern_ComprehensiveCoverage(t *testing.T) {
	t.Run("PutTargets panic pattern validation", func(t *testing.T) {
		mockT := &mockTestingT{}
		ctx := &TestContext{
			T:      mockT,
			Region: "us-east-1",
			Client: &MockEventBridgeClient{},
		}

		// This should cause validation error and trigger panic pattern
		PutTargets(ctx, "test-rule", "default", []TargetConfig{})

		assert.True(t, mockT.errorCalled)
		assert.True(t, mockT.failNowCalled)
	})

	t.Run("RemoveTargets panic pattern validation", func(t *testing.T) {
		mockT := &mockTestingT{}
		ctx := &TestContext{
			T:      mockT,
			Region: "us-east-1",
			Client: &MockEventBridgeClient{},
		}

		// This should cause validation error and trigger panic pattern
		RemoveTargets(ctx, "", "default", []string{"target-1"})

		assert.True(t, mockT.errorCalled)
		assert.True(t, mockT.failNowCalled)
	})

	t.Run("ListTargetsForRule panic pattern validation", func(t *testing.T) {
		mockT := &mockTestingT{}
		ctx := &TestContext{
			T:      mockT,
			Region: "us-east-1",
			Client: &MockEventBridgeClient{},
		}

		// This should cause validation error and trigger panic pattern
		result := ListTargetsForRule(ctx, "", "default")

		assert.Nil(t, result)
		assert.True(t, mockT.errorCalled)
		assert.True(t, mockT.failNowCalled)
	})

	t.Run("RemoveAllTargets panic pattern on service error", func(t *testing.T) {
		mockT := &mockTestingT{}
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      mockT,
			Region: "us-east-1",
			Client: mockClient,
		}

		serviceErr := errors.New("critical service error")
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr).Times(3)

		RemoveAllTargets(ctx, "test-rule", "default")

		assert.True(t, mockT.errorCalled)
		assert.True(t, mockT.failNowCalled)
		mockClient.AssertExpectations(t)
	})
}

// TestTargetOperations_EdgeCases_ComprehensiveCoverage tests unusual but valid scenarios
func TestTargetOperations_EdgeCases_ComprehensiveCoverage(t *testing.T) {
	t.Run("PutTargets with very long input transformer template", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		// Create a very long input template
		longTemplate := `{
			"event_id": "<event_id>",
			"timestamp": "<timestamp>",
			"source": "<source>",
			"detail_type": "<detail_type>",
			"account": "<account>",
			"region": "<region>",
			"resources": "<resources>",
			"detail": "<detail>",
			"metadata": {
				"processed_at": "` + time.Now().Format(time.RFC3339) + `",
				"processing_version": "2.0",
				"environment": "production",
				"correlation_id": "<correlation_id>",
				"request_id": "<request_id>",
				"extra_large_field": "` + string(make([]byte, 1000)) + `"
			}
		}`

		inputTransformer := &types.InputTransformer{
			InputPathsMap: map[string]string{
				"event_id":       "$.id",
				"timestamp":      "$.time",
				"source":         "$.source",
				"detail_type":    "$.detail-type",
				"account":        "$.account",
				"region":         "$.region",
				"resources":      "$.resources",
				"detail":         "$.detail",
				"correlation_id": "$.detail.correlation_id",
				"request_id":     "$.detail.request_id",
			},
			InputTemplate: aws.String(longTemplate),
		}

		targets := []TargetConfig{
			{
				ID:               "long-template-target",
				Arn:              "arn:aws:lambda:us-east-1:123456789012:function:ProcessComplexEvent",
				InputTransformer: inputTransformer,
			},
		}

		mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
			return input.Targets[0].InputTransformer != nil &&
				len(*input.Targets[0].InputTransformer.InputTemplate) > 1000
		}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.PutTargetsResultEntry{},
		}, nil)

		err := PutTargetsE(ctx, "complex-rule", "default", targets)

		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveTargets with targets containing special characters", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		specialTargetIDs := []string{
			"target-with-dashes",
			"target_with_underscores",
			"target.with.dots",
			"target123withNumbers",
			"TARGET-WITH-CAPS",
		}

		mockClient.On("RemoveTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
			return len(input.Ids) == 5 &&
				input.Ids[0] == "target-with-dashes" &&
				input.Ids[1] == "target_with_underscores" &&
				input.Ids[2] == "target.with.dots" &&
				input.Ids[3] == "target123withNumbers" &&
				input.Ids[4] == "TARGET-WITH-CAPS"
		}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.RemoveTargetsResultEntry{},
		}, nil)

		err := RemoveTargetsE(ctx, "special-chars-rule", "default", specialTargetIDs)

		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListTargetsForRule with nil pointer handling in response", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		// Response with some nil pointers that safeString should handle
		expectedResponse := &eventbridge.ListTargetsByRuleOutput{
			Targets: []types.Target{
				{
					Id:            aws.String("target-with-nils"),
					Arn:           aws.String("arn:aws:lambda:us-east-1:123456789012:function:Test"),
					RoleArn:       nil, // This will be nil
					Input:         nil, // This will be nil
					InputPath:     nil, // This will be nil
				},
			},
		}

		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

		results, err := ListTargetsForRuleE(ctx, "nil-fields-rule", "default")

		require.NoError(t, err)
		require.Len(t, results, 1)
		
		target := results[0]
		assert.Equal(t, "target-with-nils", target.ID)
		assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:Test", target.Arn)
		assert.Empty(t, target.RoleArn)    // safeString should return empty string
		assert.Empty(t, target.Input)      // safeString should return empty string
		assert.Empty(t, target.InputPath)  // safeString should return empty string
		
		mockClient.AssertExpectations(t)
	})
}