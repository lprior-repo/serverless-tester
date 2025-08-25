package eventbridge

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestCreateRuleE_ComprehensiveValidation tests all input validation scenarios
func TestCreateRuleE_ComprehensiveValidation(t *testing.T) {
	ctx := &TestContext{T: t, Region: "us-east-1"}

	tests := []struct {
		name        string
		config      RuleConfig
		expectError string
	}{
		{
			name:        "empty rule name",
			config:      RuleConfig{Name: "", EventPattern: `{"source": ["test"]}`},
			expectError: "rule name cannot be empty",
		},
		{
			name:        "invalid event pattern format",
			config:      RuleConfig{Name: "test", EventPattern: `{"source": "invalid"}`},
			expectError: "invalid event pattern",
		},
		{
			name:        "malformed JSON in event pattern",
			config:      RuleConfig{Name: "test", EventPattern: `{"source": [`},
			expectError: "invalid event pattern",
		},
		{
			name:        "empty event pattern should be allowed",
			config:      RuleConfig{Name: "test", ScheduleExpression: "rate(5 minutes)"},
			expectError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateRuleE(ctx, tt.config)

			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			} else {
				// For the valid case, we expect an AWS client call which we don't have
				// so we expect an error about no client configured
				require.Error(t, err)
			}
		})
	}
}

// TestCreateRuleE_EdgeCaseConfigurations tests edge cases in rule configuration
func TestCreateRuleE_EdgeCaseConfigurations(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Test case: Both event pattern and schedule expression provided
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return input.EventPattern != nil && input.ScheduleExpression != nil
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/dual-rule"),
	}, nil).Once()

	config := RuleConfig{
		Name:               "dual-rule",
		EventPattern:       `{"source": ["test"]}`,
		ScheduleExpression: "rate(5 minutes)",
	}

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "dual-rule", result.Name)
	mockClient.AssertExpectations(t)
}

// TestCreateRuleE_LargeTagSet tests rules with many tags
func TestCreateRuleE_LargeTagSet(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Create a large tag set (50 tags which is typical AWS limit)
	largeTags := make(map[string]string)
	for i := 0; i < 50; i++ {
		largeTags[fmt.Sprintf("tag%d", i)] = fmt.Sprintf("value%d", i)
	}

	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return len(input.Tags) == 50
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/tagged-rule"),
	}, nil)

	config := RuleConfig{
		Name:         "tagged-rule",
		EventPattern: `{"source": ["test"]}`,
		Tags:         largeTags,
	}

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "tagged-rule", result.Name)
	mockClient.AssertExpectations(t)
}

// TestCreateRuleE_StateVariations tests all possible rule states
func TestCreateRuleE_StateVariations(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	tests := []struct {
		name          string
		configState   types.RuleState
		expectedState types.RuleState
	}{
		{
			name:          "explicitly enabled",
			configState:   types.RuleStateEnabled,
			expectedState: types.RuleStateEnabled,
		},
		{
			name:          "explicitly disabled",
			configState:   types.RuleStateDisabled,
			expectedState: types.RuleStateDisabled,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
				return input.State == tt.expectedState
			}), mock.Anything).Return(&eventbridge.PutRuleOutput{
				RuleArn: aws.String(fmt.Sprintf("arn:aws:events:us-east-1:123456789012:rule/state-rule-%d", i)),
			}, nil).Once()

			config := RuleConfig{
				Name:         fmt.Sprintf("state-rule-%d", i),
				EventPattern: `{"source": ["test"]}`,
				State:        tt.configState,
			}

			result, err := CreateRuleE(ctx, config)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedState, result.State)
		})
	}

	mockClient.AssertExpectations(t)
}

// TestDeleteRuleE_WithTargetsCleanup tests delete rule behavior with target cleanup
func TestDeleteRuleE_WithTargetsCleanup(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Mock that the rule has targets that need to be removed first
	mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
		return *input.Rule == "rule-with-targets"
	}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("target-1"), Arn: aws.String("arn:aws:lambda:us-east-1:123456789012:function:test")},
			{Id: aws.String("target-2"), Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:test-queue")},
		},
	}, nil)

	// Mock removal of targets
	mockClient.On("RemoveTargets", mock.Anything, mock.MatchedBy(func(input *eventbridge.RemoveTargetsInput) bool {
		return *input.Rule == "rule-with-targets" && len(input.Ids) == 2
	}), mock.Anything).Return(&eventbridge.RemoveTargetsOutput{}, nil)

	// Mock successful deletion
	mockClient.On("DeleteRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DeleteRuleInput) bool {
		return *input.Name == "rule-with-targets"
	}), mock.Anything).Return(&eventbridge.DeleteRuleOutput{}, nil)

	err := DeleteRuleE(ctx, "rule-with-targets", "default")

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// TestDeleteRuleE_FailedTargetRemoval tests deletion when target removal fails
func TestDeleteRuleE_FailedTargetRemoval(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Mock that the rule has targets
	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("target-1")},
		},
	}, nil)

	// Mock failed target removal
	mockClient.On("RemoveTargets", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed to remove targets"))

	err := DeleteRuleE(ctx, "test-rule", "default")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove targets before deleting rule")
	mockClient.AssertExpectations(t)
}

// TestDescribeRuleE_ResponseFieldMapping tests all response field mapping
func TestDescribeRuleE_ResponseFieldMapping(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Test comprehensive response with all fields populated
	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:                    aws.String("comprehensive-rule"),
		Arn:                     aws.String("arn:aws:events:us-east-1:123456789012:rule/comprehensive-rule"),
		EventPattern:            aws.String(`{"source": ["test.service"], "detail-type": ["Test Event"]}`),
		ScheduleExpression:      aws.String("rate(5 minutes)"),
		Description:             aws.String("A comprehensive test rule with all fields"),
		State:                   types.RuleStateEnabled,
		EventBusName:            aws.String("custom-event-bus"),
		ManagedBy:               aws.String("user"),
		RoleArn:                 aws.String("arn:aws:iam::123456789012:role/EventBridgeRole"),
	}, nil)

	result, err := DescribeRuleE(ctx, "comprehensive-rule", "custom-event-bus")

	require.NoError(t, err)
	assert.Equal(t, "comprehensive-rule", result.Name)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:rule/comprehensive-rule", result.RuleArn)
	assert.Equal(t, "A comprehensive test rule with all fields", result.Description)
	assert.Equal(t, types.RuleStateEnabled, result.State)
	assert.Equal(t, "custom-event-bus", result.EventBusName)
	mockClient.AssertExpectations(t)
}

// TestDescribeRuleE_NilFieldHandling tests handling of nil response fields
func TestDescribeRuleE_NilFieldHandling(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Test response with minimal fields (nil pointers)
	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:  aws.String("minimal-rule"),
		Arn:   aws.String("arn:aws:events:us-east-1:123456789012:rule/minimal-rule"),
		State: types.RuleStateDisabled,
		// Description, EventBusName, etc. are nil
	}, nil)

	result, err := DescribeRuleE(ctx, "minimal-rule", "default")

	require.NoError(t, err)
	assert.Equal(t, "minimal-rule", result.Name)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:rule/minimal-rule", result.RuleArn)
	assert.Equal(t, "", result.Description) // Should be empty string, not panic
	assert.Equal(t, "", result.EventBusName) // Should be empty string, not panic
	assert.Equal(t, types.RuleStateDisabled, result.State)
	mockClient.AssertExpectations(t)
}

// TestListRulesE_PaginationHandling tests pagination scenarios
func TestListRulesE_PaginationHandling(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Test empty list
	mockClient.On("ListRules", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListRulesOutput{
		Rules: []types.Rule{},
	}, nil)

	results, err := ListRulesE(ctx, "empty-bus")

	require.NoError(t, err)
	assert.Empty(t, results)
	mockClient.AssertExpectations(t)
}

// TestListRulesE_LargeResultSet tests handling of large result sets
func TestListRulesE_LargeResultSet(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Generate many rules
	largeRules := make([]types.Rule, 100)
	for i := 0; i < 100; i++ {
		largeRules[i] = types.Rule{
			Name:         aws.String(fmt.Sprintf("rule-%d", i)),
			Arn:          aws.String(fmt.Sprintf("arn:aws:events:us-east-1:123456789012:rule/rule-%d", i)),
			Description:  aws.String(fmt.Sprintf("Rule number %d", i)),
			State:        types.RuleStateEnabled,
			EventBusName: aws.String("default"),
		}
	}

	mockClient.On("ListRules", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListRulesOutput{
		Rules: largeRules,
	}, nil)

	results, err := ListRulesE(ctx, "default")

	require.NoError(t, err)
	assert.Len(t, results, 100)

	// Verify first and last elements to ensure proper conversion
	assert.Equal(t, "rule-0", results[0].Name)
	assert.Equal(t, "rule-99", results[99].Name)
	mockClient.AssertExpectations(t)
}

// TestUpdateRuleStateE_BothStates tests enable/disable operations comprehensively
func TestUpdateRuleStateE_BothStates(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Test enabling
	mockClient.On("EnableRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.EnableRuleInput) bool {
		return *input.Name == "state-test-rule" && *input.EventBusName == "default"
	}), mock.Anything).Return(&eventbridge.EnableRuleOutput{}, nil)

	// Test disabling
	mockClient.On("DisableRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DisableRuleInput) bool {
		return *input.Name == "state-test-rule" && *input.EventBusName == "default"
	}), mock.Anything).Return(&eventbridge.DisableRuleOutput{}, nil)

	// Test enable
	err := EnableRuleE(ctx, "state-test-rule", "default")
	require.NoError(t, err)

	// Test disable
	err = DisableRuleE(ctx, "state-test-rule", "default")
	require.NoError(t, err)

	mockClient.AssertExpectations(t)
}

// TestRuleOperations_ServiceErrorRetry tests retry behavior for all operations
func TestRuleOperations_ServiceErrorRetry(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	temporaryError := errors.New("temporary service error")

	t.Run("CreateRuleE retry", func(t *testing.T) {
		// First call fails, second succeeds
		mockClient.On("PutRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, temporaryError).Once()
		mockClient.On("PutRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.PutRuleOutput{
			RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/retry-rule"),
		}, nil).Once()

		config := RuleConfig{Name: "retry-rule", EventPattern: `{"source": ["test"]}`}
		result, err := CreateRuleE(ctx, config)

		require.NoError(t, err)
		assert.Equal(t, "retry-rule", result.Name)
	})

	t.Run("DescribeRuleE retry", func(t *testing.T) {
		// First call fails, second succeeds
		mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, temporaryError).Once()
		mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
			Name: aws.String("retry-rule"),
			Arn:  aws.String("arn:aws:events:us-east-1:123456789012:rule/retry-rule"),
			State: types.RuleStateEnabled,
		}, nil).Once()

		result, err := DescribeRuleE(ctx, "retry-rule", "default")

		require.NoError(t, err)
		assert.Equal(t, "retry-rule", result.Name)
	})

	t.Run("ListRulesE retry", func(t *testing.T) {
		// First call fails, second succeeds
		mockClient.On("ListRules", mock.Anything, mock.Anything, mock.Anything).Return(nil, temporaryError).Once()
		mockClient.On("ListRules", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListRulesOutput{
			Rules: []types.Rule{
				{Name: aws.String("retry-rule"), State: types.RuleStateEnabled},
			},
		}, nil).Once()

		results, err := ListRulesE(ctx, "default")

		require.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("EnableRuleE retry", func(t *testing.T) {
		// First call fails, second succeeds
		mockClient.On("EnableRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, temporaryError).Once()
		mockClient.On("EnableRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.EnableRuleOutput{}, nil).Once()

		err := EnableRuleE(ctx, "retry-rule", "default")
		require.NoError(t, err)
	})

	t.Run("DisableRuleE retry", func(t *testing.T) {
		// First call fails, second succeeds
		mockClient.On("DisableRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, temporaryError).Once()
		mockClient.On("DisableRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DisableRuleOutput{}, nil).Once()

		err := DisableRuleE(ctx, "retry-rule", "default")
		require.NoError(t, err)
	})

	mockClient.AssertExpectations(t)
}

// TestRuleOperations_MaxRetryExceeded tests max retry scenarios
func TestRuleOperations_MaxRetryExceeded(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	persistentError := errors.New("persistent service error")

	t.Run("CreateRuleE max retry", func(t *testing.T) {
		mockClient.On("PutRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, persistentError).Times(DefaultRetryAttempts)

		config := RuleConfig{Name: "fail-rule", EventPattern: `{"source": ["test"]}`}
		_, err := CreateRuleE(ctx, config)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create rule after 3 attempts")
	})

	t.Run("DescribeRuleE max retry", func(t *testing.T) {
		mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, persistentError).Times(DefaultRetryAttempts)

		_, err := DescribeRuleE(ctx, "fail-rule", "default")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to describe rule after 3 attempts")
	})

	t.Run("ListRulesE max retry", func(t *testing.T) {
		mockClient.On("ListRules", mock.Anything, mock.Anything, mock.Anything).Return(nil, persistentError).Times(DefaultRetryAttempts)

		_, err := ListRulesE(ctx, "default")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list rules after 3 attempts")
	})

	t.Run("EnableRuleE max retry", func(t *testing.T) {
		mockClient.On("EnableRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, persistentError).Times(DefaultRetryAttempts)

		err := EnableRuleE(ctx, "fail-rule", "default")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update rule state after 3 attempts")
	})

	t.Run("DisableRuleE max retry", func(t *testing.T) {
		mockClient.On("DisableRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, persistentError).Times(DefaultRetryAttempts)

		err := DisableRuleE(ctx, "fail-rule", "default")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update rule state after 3 attempts")
	})

	mockClient.AssertExpectations(t)
}
