package eventbridge

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestCreateRuleE_UpdateExistingRule tests updating a rule by calling CreateRuleE with same name
func TestCreateRuleE_UpdateExistingRule(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// First, create a rule
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.Name == "updatable-rule" && 
			   *input.Description == "Original description" &&
			   input.EventPattern != nil && *input.EventPattern == `{"source": ["original.service"]}`
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/updatable-rule"),
	}, nil).Once()

	originalConfig := RuleConfig{
		Name:         "updatable-rule",
		Description:  "Original description",
		EventPattern: `{"source": ["original.service"]}`,
		State:        types.RuleStateEnabled,
	}

	result, err := CreateRuleE(ctx, originalConfig)
	require.NoError(t, err)
	assert.Equal(t, "updatable-rule", result.Name)

	// Now update the same rule with different parameters
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.Name == "updatable-rule" && 
			   *input.Description == "Updated description" &&
			   input.EventPattern != nil && *input.EventPattern == `{"source": ["updated.service"]}`
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/updatable-rule"),
	}, nil).Once()

	updatedConfig := RuleConfig{
		Name:         "updatable-rule",  // Same name
		Description:  "Updated description",
		EventPattern: `{"source": ["updated.service"]}`,
		State:        types.RuleStateDisabled,  // Changed state
	}

	updatedResult, err := CreateRuleE(ctx, updatedConfig)
	require.NoError(t, err)
	assert.Equal(t, "updatable-rule", updatedResult.Name)
	assert.Equal(t, types.RuleStateDisabled, updatedResult.State)

	mockClient.AssertExpectations(t)
}

// TestCreateRuleE_UpdateFromEventPatternToSchedule tests changing rule type
func TestCreateRuleE_UpdateFromEventPatternToSchedule(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// First, create a rule with event pattern
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return input.EventPattern != nil && input.ScheduleExpression == nil
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/changing-rule"),
	}, nil).Once()

	originalConfig := RuleConfig{
		Name:         "changing-rule",
		EventPattern: `{"source": ["test.service"]}`,
	}

	_, err := CreateRuleE(ctx, originalConfig)
	require.NoError(t, err)

	// Update to use schedule expression instead
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return input.ScheduleExpression != nil && *input.ScheduleExpression == "rate(5 minutes)" && input.EventPattern == nil
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/changing-rule"),
	}, nil).Once()

	updatedConfig := RuleConfig{
		Name:               "changing-rule",
		ScheduleExpression: "rate(5 minutes)",
		// EventPattern removed
	}

	_, err = CreateRuleE(ctx, updatedConfig)
	require.NoError(t, err)

	mockClient.AssertExpectations(t)
}

// TestCreateRuleE_UpdateTags tests updating rule tags
func TestCreateRuleE_UpdateTags(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Create rule with initial tags
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return len(input.Tags) == 2 && 
			   containsTag(input.Tags, "Environment", "dev") &&
			   containsTag(input.Tags, "Team", "backend")
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/tagged-rule"),
	}, nil).Once()

	originalConfig := RuleConfig{
		Name:         "tagged-rule",
		EventPattern: `{"source": ["test"]}`,
		Tags: map[string]string{
			"Environment": "dev",
			"Team":        "backend",
		},
	}

	_, err := CreateRuleE(ctx, originalConfig)
	require.NoError(t, err)

	// Update tags
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return len(input.Tags) == 3 && 
			   containsTag(input.Tags, "Environment", "prod") &&
			   containsTag(input.Tags, "Team", "backend") &&
			   containsTag(input.Tags, "Version", "2.0")
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/tagged-rule"),
	}, nil).Once()

	updatedConfig := RuleConfig{
		Name:         "tagged-rule",
		EventPattern: `{"source": ["test"]}`,
		Tags: map[string]string{
			"Environment": "prod",    // Changed
			"Team":        "backend", // Same
			"Version":     "2.0",     // New
		},
	}

	_, err = CreateRuleE(ctx, updatedConfig)
	require.NoError(t, err)

	mockClient.AssertExpectations(t)
}

// TestCreateRuleE_RemoveAllTags tests removing all tags from a rule
func TestCreateRuleE_RemoveAllTags(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Create rule with tags
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return len(input.Tags) == 1
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/no-tags-rule"),
	}, nil).Once()

	originalConfig := RuleConfig{
		Name:         "no-tags-rule",
		EventPattern: `{"source": ["test"]}`,
		Tags: map[string]string{
			"Environment": "dev",
		},
	}

	_, err := CreateRuleE(ctx, originalConfig)
	require.NoError(t, err)

	// Remove all tags (empty map)
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return len(input.Tags) == 0
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/no-tags-rule"),
	}, nil).Once()

	updatedConfig := RuleConfig{
		Name:         "no-tags-rule",
		EventPattern: `{"source": ["test"]}`,
		Tags:         map[string]string{}, // Empty tags
	}

	_, err = CreateRuleE(ctx, updatedConfig)
	require.NoError(t, err)

	mockClient.AssertExpectations(t)
}

// TestDeleteRuleE_RetryOnTargetRemovalFail tests retry behavior during target removal
func TestDeleteRuleE_RetryOnTargetRemovalFail(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Mock that rule has targets
	mockClient.On("ListTargetsByRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{Id: aws.String("target-1")},
		},
	}, nil)

	// First attempt to remove targets fails
	mockClient.On("RemoveTargets", mock.Anything, mock.Anything, mock.Anything).Return(nil, 
		&types.ConcurrentModificationException{Message: aws.String("concurrent modification")}).Once()

	// Second attempt succeeds
	mockClient.On("RemoveTargets", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.RemoveTargetsOutput{}, nil).Once()

	// Then delete rule succeeds
	mockClient.On("DeleteRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DeleteRuleOutput{}, nil)

	err := DeleteRuleE(ctx, "retry-rule", "default")
	require.NoError(t, err)

	mockClient.AssertExpectations(t)
}

// TestListRulesE_EmptyListHandling tests handling of empty rule lists
func TestListRulesE_EmptyListHandling(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Return empty rules list
	mockClient.On("ListRules", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListRulesInput) bool {
		return *input.EventBusName == "empty-bus"
	}), mock.Anything).Return(&eventbridge.ListRulesOutput{
		Rules: []types.Rule{}, // Empty list
	}, nil)

	results, err := ListRulesE(ctx, "empty-bus")

	require.NoError(t, err)
	assert.Empty(t, results)
	assert.NotNil(t, results) // Should be empty slice, not nil
	mockClient.AssertExpectations(t)
}

// TestDescribeRuleE_MissingOptionalFields tests handling when optional fields are missing
func TestDescribeRuleE_MissingOptionalFields(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Return response with only required fields
	mockClient.On("DescribeRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:  aws.String("minimal-rule"),
		State: types.RuleStateEnabled,
		// Arn, Description, EventPattern, ScheduleExpression, etc. are nil
	}, nil)

	result, err := DescribeRuleE(ctx, "minimal-rule", "default")

	require.NoError(t, err)
	assert.Equal(t, "minimal-rule", result.Name)
	assert.Equal(t, "", result.RuleArn)      // safeString should handle nil
	assert.Equal(t, "", result.Description)  // safeString should handle nil
	assert.Equal(t, "", result.EventBusName) // safeString should handle nil
	assert.Equal(t, types.RuleStateEnabled, result.State)
	mockClient.AssertExpectations(t)
}

// TestEnableDisableRule_StateToggling tests toggling rule state multiple times
func TestEnableDisableRule_StateToggling(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Enable -> Disable -> Enable sequence
	mockClient.On("EnableRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.EnableRuleOutput{}, nil).Twice()
	mockClient.On("DisableRule", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.DisableRuleOutput{}, nil).Once()

	// Enable rule
	err := EnableRuleE(ctx, "toggle-rule", "default")
	require.NoError(t, err)

	// Disable rule  
	err = DisableRuleE(ctx, "toggle-rule", "default")
	require.NoError(t, err)

	// Enable again
	err = EnableRuleE(ctx, "toggle-rule", "default")
	require.NoError(t, err)

	mockClient.AssertExpectations(t)
}

// Helper function to check if tags contain a specific key-value pair
func containsTag(tags []types.Tag, key, value string) bool {
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil && *tag.Key == key && *tag.Value == value {
			return true
		}
	}
	return false
}