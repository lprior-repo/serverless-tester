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

// TestTargetValidation_ComprehensiveCoverage tests all target validation scenarios
func TestTargetValidation_ComprehensiveCoverage(t *testing.T) {
	t.Run("ValidateTargetConfiguration should validate Lambda targets", func(t *testing.T) {
		// RED: Write failing test for Lambda target validation
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
		}

		target := TargetConfig{
			ID:  "lambda-target",
			Arn: "arn:aws:lambda:us-east-1:123456789012:function:TestFunction",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		require.NoError(t, err)
	})

	t.Run("ValidateTargetConfiguration should validate SQS targets", func(t *testing.T) {
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
		}

		target := TargetConfig{
			ID:  "sqs-target",
			Arn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		require.NoError(t, err)
	})

	t.Run("ValidateTargetConfiguration should validate SNS targets", func(t *testing.T) {
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
		}

		target := TargetConfig{
			ID:      "sns-target",
			Arn:     "arn:aws:sns:us-east-1:123456789012:test-topic",
			RoleArn: "arn:aws:iam::123456789012:role/EventBridgeRole",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		require.NoError(t, err)
	})

	t.Run("ValidateTargetConfiguration should reject invalid ARN format", func(t *testing.T) {
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
		}

		target := TargetConfig{
			ID:  "invalid-target",
			Arn: "invalid-arn-format",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ARN format")
	})

	t.Run("ValidateTargetConfiguration should reject empty target ID", func(t *testing.T) {
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
		}

		target := TargetConfig{
			ID:  "", // Empty ID
			Arn: "arn:aws:lambda:us-east-1:123456789012:function:TestFunction",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target ID cannot be empty")
	})

	t.Run("ValidateTargetConfiguration should reject malformed JSON in Input field", func(t *testing.T) {
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
		}

		target := TargetConfig{
			ID:    "json-target",
			Arn:   "arn:aws:lambda:us-east-1:123456789012:function:TestFunction",
			Input: "{ invalid json",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON in Input field")
	})

	t.Run("ValidateTargetConfiguration should reject invalid InputPath JSONPath", func(t *testing.T) {
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
		}

		target := TargetConfig{
			ID:        "path-target",
			Arn:       "arn:aws:lambda:us-east-1:123456789012:function:TestFunction",
			InputPath: "invalid-jsonpath-expression",
		}

		err := ValidateTargetConfigurationE(ctx, target)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSONPath in InputPath")
	})
}

// TestTargetService_ComprehensiveCoverage tests service-specific target functionality  
func TestTargetService_ComprehensiveCoverage(t *testing.T) {
	t.Run("GetTargetServiceType should identify Lambda targets", func(t *testing.T) {
		targetArn := "arn:aws:lambda:us-east-1:123456789012:function:TestFunction"
		serviceType := GetTargetServiceType(targetArn)
		assert.Equal(t, "lambda", serviceType)
	})

	t.Run("GetTargetServiceType should identify SQS targets", func(t *testing.T) {
		targetArn := "arn:aws:sqs:us-east-1:123456789012:test-queue"
		serviceType := GetTargetServiceType(targetArn)
		assert.Equal(t, "sqs", serviceType)
	})

	t.Run("GetTargetServiceType should identify SNS targets", func(t *testing.T) {
		targetArn := "arn:aws:sns:us-east-1:123456789012:test-topic"
		serviceType := GetTargetServiceType(targetArn)
		assert.Equal(t, "sns", serviceType)
	})

	t.Run("GetTargetServiceType should identify Step Functions targets", func(t *testing.T) {
		targetArn := "arn:aws:states:us-east-1:123456789012:stateMachine:TestStateMachine"
		serviceType := GetTargetServiceType(targetArn)
		assert.Equal(t, "stepfunctions", serviceType)
	})

	t.Run("GetTargetServiceType should identify Kinesis targets", func(t *testing.T) {
		targetArn := "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream"
		serviceType := GetTargetServiceType(targetArn)
		assert.Equal(t, "kinesis", serviceType)
	})

	t.Run("GetTargetServiceType should identify ECS tasks", func(t *testing.T) {
		targetArn := "arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"
		serviceType := GetTargetServiceType(targetArn)
		assert.Equal(t, "ecs", serviceType)
	})

	t.Run("GetTargetServiceType should return unknown for unsupported services", func(t *testing.T) {
		targetArn := "arn:aws:unsupported:us-east-1:123456789012:resource/test"
		serviceType := GetTargetServiceType(targetArn)
		assert.Equal(t, "unknown", serviceType)
	})

	t.Run("GetTargetServiceType should handle malformed ARNs gracefully", func(t *testing.T) {
		targetArn := "malformed-arn"
		serviceType := GetTargetServiceType(targetArn)
		assert.Equal(t, "unknown", serviceType)
	})
}

// TestBatchTargetOperations_ComprehensiveCoverage tests batch operations for targets
func TestBatchTargetOperations_ComprehensiveCoverage(t *testing.T) {
	t.Run("BatchPutTargets should handle maximum batch size", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		// Create maximum number of targets (EventBridge supports 100 targets per rule)
		const maxTargets = 100
		targets := make([]TargetConfig, maxTargets)
		for i := 0; i < maxTargets; i++ {
			targets[i] = TargetConfig{
				ID:  fmt.Sprintf("target-%03d", i+1),
				Arn: fmt.Sprintf("arn:aws:lambda:us-east-1:123456789012:function:Function%03d", i+1),
			}
		}

		mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
			return len(input.Targets) == maxTargets
		}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.PutTargetsResultEntry{},
		}, nil)

		err := BatchPutTargetsE(ctx, "batch-rule", "default", targets, 50)

		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("BatchPutTargets should handle chunked operations", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		targets := make([]TargetConfig, 25)
		for i := 0; i < 25; i++ {
			targets[i] = TargetConfig{
				ID:  fmt.Sprintf("chunk-target-%d", i+1),
				Arn: fmt.Sprintf("arn:aws:lambda:us-east-1:123456789012:function:ChunkFunction%d", i+1),
			}
		}

		// Expect 3 calls: 10, 10, 5 targets
		mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
			return len(input.Targets) == 10
		}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.PutTargetsResultEntry{},
		}, nil).Times(2)

		mockClient.On("PutTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutTargetsInput) bool {
			return len(input.Targets) == 5
		}), mock.Anything).Return(&eventbridge.PutTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.PutTargetsResultEntry{},
		}, nil).Once()

		err := BatchPutTargetsE(ctx, "chunked-rule", "default", targets, 10)

		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("BatchRemoveTargets should handle large batch removals", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		targetIDs := make([]string, 150)
		for i := 0; i < 150; i++ {
			targetIDs[i] = fmt.Sprintf("remove-target-%03d", i+1)
		}

		// EventBridge limits RemoveTargets to 100 per call
		// Expect 2 calls: 100, 50 targets
		mockClient.On("RemoveTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
			return len(input.Ids) == 100
		}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.RemoveTargetsResultEntry{},
		}, nil).Once()

		mockClient.On("RemoveTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
			return len(input.Ids) == 50
		}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.RemoveTargetsResultEntry{},
		}, nil).Once()

		err := BatchRemoveTargetsE(ctx, "large-batch-rule", "default", targetIDs, 100)

		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}

// TestTargetInputTransformation_ComprehensiveCoverage tests input transformation scenarios
func TestTargetInputTransformation_ComprehensiveCoverage(t *testing.T) {
	t.Run("BuildInputTransformer should create valid transformation", func(t *testing.T) {
		pathMap := map[string]string{
			"timestamp": "$.time",
			"source":    "$.source",
			"account":   "$.account",
			"region":    "$.region",
		}

		template := `{
			"event_time": "<timestamp>",
			"event_source": "<source>",
			"aws_account": "<account>",
			"aws_region": "<region>",
			"processed_at": "2024-01-01T00:00:00Z"
		}`

		transformer := BuildInputTransformer(pathMap, template)

		require.NotNil(t, transformer)
		assert.Equal(t, 4, len(transformer.InputPathsMap))
		assert.Contains(t, *transformer.InputTemplate, "event_time")
		assert.Contains(t, *transformer.InputTemplate, "processed_at")
	})

	t.Run("ValidateInputTransformer should accept valid configurations", func(t *testing.T) {
		transformer := &types.InputTransformer{
			InputPathsMap: map[string]string{
				"detail": "$.detail",
				"source": "$.source",
			},
			InputTemplate: aws.String(`{"data": "<detail>", "origin": "<source>"}`),
		}

		err := ValidateInputTransformerE(transformer)
		require.NoError(t, err)
	})

	t.Run("ValidateInputTransformer should reject invalid template JSON", func(t *testing.T) {
		transformer := &types.InputTransformer{
			InputPathsMap: map[string]string{
				"detail": "$.detail",
			},
			InputTemplate: aws.String(`{ invalid json template`),
		}

		err := ValidateInputTransformerE(transformer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON in InputTemplate")
	})

	t.Run("ValidateInputTransformer should reject template with unmapped placeholders", func(t *testing.T) {
		transformer := &types.InputTransformer{
			InputPathsMap: map[string]string{
				"detail": "$.detail",
			},
			InputTemplate: aws.String(`{"data": "<detail>", "missing": "<unmapped>"}`),
		}

		err := ValidateInputTransformerE(transformer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unmapped placeholder")
	})
}

// TestTargetDeadLetterQueue_ComprehensiveCoverage tests DLQ configuration
func TestTargetDeadLetterQueue_ComprehensiveCoverage(t *testing.T) {
	t.Run("ValidateDeadLetterConfig should accept valid SQS ARN", func(t *testing.T) {
		dlqConfig := &types.DeadLetterConfig{
			Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:dlq"),
		}

		err := ValidateDeadLetterConfigE(dlqConfig)
		require.NoError(t, err)
	})

	t.Run("ValidateDeadLetterConfig should reject non-SQS ARN", func(t *testing.T) {
		dlqConfig := &types.DeadLetterConfig{
			Arn: aws.String("arn:aws:sns:us-east-1:123456789012:topic"),
		}

		err := ValidateDeadLetterConfigE(dlqConfig)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dead letter queue must be an SQS queue")
	})

	t.Run("ValidateDeadLetterConfig should reject malformed ARN", func(t *testing.T) {
		dlqConfig := &types.DeadLetterConfig{
			Arn: aws.String("invalid-dlq-arn"),
		}

		err := ValidateDeadLetterConfigE(dlqConfig)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid dead letter queue ARN")
	})

	t.Run("BuildDeadLetterConfig should create valid configuration", func(t *testing.T) {
		queueArn := "arn:aws:sqs:us-east-1:123456789012:my-dlq"
		dlqConfig := BuildDeadLetterConfig(queueArn)

		require.NotNil(t, dlqConfig)
		assert.Equal(t, queueArn, *dlqConfig.Arn)
	})
}

// TestTargetRetryPolicy_ComprehensiveCoverage tests retry policy configuration
func TestTargetRetryPolicy_ComprehensiveCoverage(t *testing.T) {
	t.Run("ValidateRetryPolicy should accept valid configuration", func(t *testing.T) {
		retryPolicy := &types.RetryPolicy{
			MaximumRetryAttempts:      aws.Int32(5),
			MaximumEventAgeInSeconds: aws.Int32(600),
		}

		err := ValidateRetryPolicyE(retryPolicy)
		require.NoError(t, err)
	})

	t.Run("ValidateRetryPolicy should reject excessive retry attempts", func(t *testing.T) {
		retryPolicy := &types.RetryPolicy{
			MaximumRetryAttempts:      aws.Int32(200), // EventBridge max is 185
			MaximumEventAgeInSeconds: aws.Int32(600),
		}

		err := ValidateRetryPolicyE(retryPolicy)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum retry attempts exceeds limit")
	})

	t.Run("ValidateRetryPolicy should reject excessive event age", func(t *testing.T) {
		retryPolicy := &types.RetryPolicy{
			MaximumRetryAttempts:      aws.Int32(3),
			MaximumEventAgeInSeconds: aws.Int32(86500), // EventBridge max is 86400 (24 hours)
		}

		err := ValidateRetryPolicyE(retryPolicy)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum event age exceeds limit")
	})

	t.Run("BuildRetryPolicy should create valid configuration", func(t *testing.T) {
		retryPolicy := BuildRetryPolicy(3, 300)

		require.NotNil(t, retryPolicy)
		assert.Equal(t, int32(3), *retryPolicy.MaximumRetryAttempts)
		assert.Equal(t, int32(300), *retryPolicy.MaximumEventAgeInSeconds)
	})
}

// TestTargetErrorScenarios_ComprehensiveCoverage tests comprehensive error scenarios
func TestTargetErrorScenarios_ComprehensiveCoverage(t *testing.T) {
	t.Run("PutTargets should handle mixed success and failure responses", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		targets := []TargetConfig{
			{ID: "success-1", Arn: "arn:aws:lambda:us-east-1:123:function:Valid1"},
			{ID: "failure-1", Arn: "arn:aws:lambda:us-east-1:123:function:Invalid1"},
			{ID: "success-2", Arn: "arn:aws:sqs:us-east-1:123:queue:Valid"},
			{ID: "failure-2", Arn: "arn:aws:sns:us-east-1:123:topic:Invalid"},
		}

		mockClient.On("PutTargets", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.PutTargetsOutput{
			FailedEntryCount: 2,
			FailedEntries: []types.PutTargetsResultEntry{
				{
					TargetId:     aws.String("failure-1"),
					ErrorCode:    aws.String("ValidationException"),
					ErrorMessage: aws.String("Function not found"),
				},
				{
					TargetId:     aws.String("failure-2"),
					ErrorCode:    aws.String("AccessDeniedException"),
					ErrorMessage: aws.String("Insufficient permissions for topic"),
				},
			},
		}, nil)

		err := PutTargetsE(ctx, "mixed-result-rule", "default", targets)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add 2 targets")
		assert.Contains(t, err.Error(), "failure-1")
		assert.Contains(t, err.Error(), "Function not found")
		assert.Contains(t, err.Error(), "failure-2")
		assert.Contains(t, err.Error(), "Insufficient permissions for topic")
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveTargets should handle service throttling with exponential backoff", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		targetIDs := []string{"throttled-target"}

		throttleErr := errors.New("ThrottlingException: request rate exceeded")
		mockClient.On("RemoveTargets", mock.Anything, mock.Anything, mock.Anything).Return(nil, throttleErr).Times(2)
		mockClient.On("RemoveTargets", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.RemoveTargetsOutput{
			FailedEntryCount: 0,
			FailedEntries:    []types.RemoveTargetsResultEntry{},
		}, nil).Once()

		start := time.Now()
		err := RemoveTargetsE(ctx, "throttled-rule", "default", targetIDs)
		duration := time.Since(start)

		require.NoError(t, err)
		// Should have delays from retries (exponential backoff: ~1s + ~2s)
		assert.True(t, duration >= 3*time.Second, "Expected exponential backoff delays")
		mockClient.AssertExpectations(t)
	})

	t.Run("ListTargetsForRule should handle empty rule scenario", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
			return *input.Rule == "rule-with-no-targets"
		}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
			Targets: []types.Target{}, // Empty targets list
		}, nil)

		results, err := ListTargetsForRuleE(ctx, "rule-with-no-targets", "default")

		require.NoError(t, err)
		assert.Empty(t, results)
		mockClient.AssertExpectations(t)
	})

	t.Run("RemoveAllTargets should handle no targets to remove", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		// Mock empty targets list
		mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
			Targets: []types.Target{},
		}, nil)

		// RemoveTargets should NOT be called
		// This tests the early return optimization in RemoveAllTargetsE

		err := RemoveAllTargetsE(ctx, "empty-rule", "default")

		require.NoError(t, err)
		mockClient.AssertExpectations(t) // Should only see ListTargetsByRule call
	})
}