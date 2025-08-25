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
	"github.com/stretchr/testify/require"
)

// Test CreateRuleE success case
func TestCreateRuleE_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "test-rule",
		Description:  "Test rule",
		EventPattern: `{"source": ["test.service"]}`,
		State:        types.RuleStateEnabled,
		EventBusName: "default",
	}

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:rule/test-rule", result.RuleArn)
	assert.Equal(t, types.RuleStateEnabled, result.State)
	assert.Equal(t, "Test rule", result.Description)
	assert.Equal(t, "default", result.EventBusName)
	mockClient.AssertExpectations(t)
}

// Test CreateRuleE with schedule expression
func TestCreateRuleE_MockedWithSchedule(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/scheduled-rule"),
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := RuleConfig{
		Name:               "scheduled-rule",
		Description:        "Scheduled rule",
		ScheduleExpression: "rate(5 minutes)",
		State:              types.RuleStateEnabled,
		EventBusName:       "default",
	}

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "scheduled-rule", result.Name)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:rule/scheduled-rule", result.RuleArn)
	mockClient.AssertExpectations(t)
}

// Test CreateRuleE with tags
func TestCreateRuleE_MockedWithTags(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/tagged-rule"),
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "tagged-rule",
		Description:  "Rule with tags",
		EventPattern: `{"source": ["test.service"]}`,
		State:        types.RuleStateEnabled,
		EventBusName: "default",
		Tags: map[string]string{
			"Environment": "test",
			"Team":        "backend",
		},
	}

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "tagged-rule", result.Name)
	mockClient.AssertExpectations(t)
}

// Test CreateRuleE with default values
func TestCreateRuleE_MockedDefaults(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/default-rule"),
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "default-rule",
		EventPattern: `{"source": ["test.service"]}`,
		// EventBusName and State not specified - should use defaults
	}

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "default-rule", result.Name)
	assert.Equal(t, "default", result.EventBusName) // Should use default
	assert.Equal(t, types.RuleStateEnabled, result.State) // Should use default
	mockClient.AssertExpectations(t)
}

// Test CreateRuleE retry success
func TestCreateRuleE_MockedRetrySuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	// First call fails
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(nil, errors.New("temporary failure")).Once()
	// Second call succeeds
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/retry-rule"),
	}, nil).Once()

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "retry-rule",
		EventPattern: `{"source": ["test.service"]}`,
	}

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "retry-rule", result.Name)
	mockClient.AssertExpectations(t)
}

// Test CreateRuleE max retries exceeded
func TestCreateRuleE_MockedMaxRetriesExceeded(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(nil, errors.New("persistent failure")).Times(DefaultRetryAttempts)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "failed-rule",
		EventPattern: `{"source": ["test.service"]}`,
	}

	result, err := CreateRuleE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create rule after")
	assert.Empty(t, result.Name)
	mockClient.AssertExpectations(t)
}

// Test CreateRule panic variant
func TestCreateRule_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
	}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "test-rule",
		EventPattern: `{"source": ["test.service"]}`,
	}

	result := CreateRule(ctx, config)

	assert.Equal(t, "test-rule", result.Name)
	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}

// Test CreateRule panic on error
func TestCreateRule_MockedPanicOnError(t *testing.T) {
	mockT := &MockTestingT{}
	mockT.On("Errorf", "CreateRule failed: %v", mock.Anything).Once()
	mockT.On("FailNow").Once()

	ctx := &TestContext{T: mockT}

	config := RuleConfig{
		Name: "", // Empty name should cause error
	}

	CreateRule(ctx, config)

	mockT.AssertExpectations(t)
}

// Test DeleteRuleE success
func TestDeleteRuleE_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	// Mock RemoveAllTargetsE calls
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)
	// Mock DeleteRule call
	mockClient.On("DeleteRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.DeleteRuleOutput{}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	err := DeleteRuleE(ctx, "test-rule", "default")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// Test DeleteRuleE with default event bus
func TestDeleteRuleE_MockedDefaultEventBus(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)
	mockClient.On("DeleteRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.DeleteRuleOutput{}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	err := DeleteRuleE(ctx, "test-rule", "") // Empty event bus should use default

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// Test DeleteRule panic variant
func TestDeleteRule_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)
	mockClient.On("DeleteRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.DeleteRuleOutput{}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	DeleteRule(ctx, "test-rule", "default")

	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}

// Test DescribeRuleE success
func TestDescribeRuleE_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("DescribeRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:         aws.String("test-rule"),
		Arn:          aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
		Description:  aws.String("Test rule description"),
		State:        types.RuleStateEnabled,
		EventBusName: aws.String("default"),
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	result, err := DescribeRuleE(ctx, "test-rule", "default")

	require.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:rule/test-rule", result.RuleArn)
	assert.Equal(t, "Test rule description", result.Description)
	assert.Equal(t, types.RuleStateEnabled, result.State)
	assert.Equal(t, "default", result.EventBusName)
	mockClient.AssertExpectations(t)
}

// Test DescribeRule panic variant
func TestDescribeRule_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("DescribeRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name: aws.String("test-rule"),
		Arn:  aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
	}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	result := DescribeRule(ctx, "test-rule", "default")

	assert.Equal(t, "test-rule", result.Name)
	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}

// Test EnableRuleE success  
func TestEnableRuleE_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("EnableRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.EnableRuleOutput{}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	err := EnableRuleE(ctx, "test-rule", "default")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// Test EnableRule panic variant
func TestEnableRule_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("EnableRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.EnableRuleOutput{}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	EnableRule(ctx, "test-rule", "default")

	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}

// Test DisableRuleE success
func TestDisableRuleE_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("DisableRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.DisableRuleOutput{}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	err := DisableRuleE(ctx, "test-rule", "default")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// Test DisableRule panic variant
func TestDisableRule_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("DisableRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.DisableRuleOutput{}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	DisableRule(ctx, "test-rule", "default")

	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}

// Test ListRulesE success
func TestListRulesE_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListRules", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListRulesOutput{
		Rules: []types.Rule{
			{
				Name:        aws.String("rule-1"),
				Arn:         aws.String("arn:aws:events:us-east-1:123456789012:rule/rule-1"),
				Description: aws.String("First rule"),
				State:       types.RuleStateEnabled,
			},
			{
				Name:        aws.String("rule-2"),
				Arn:         aws.String("arn:aws:events:us-east-1:123456789012:rule/rule-2"),
				Description: aws.String("Second rule"),
				State:       types.RuleStateDisabled,
			},
		},
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	results, err := ListRulesE(ctx, "default")

	require.NoError(t, err)
	assert.Len(t, results, 2)
	
	assert.Equal(t, "rule-1", results[0].Name)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:rule/rule-1", results[0].RuleArn)
	assert.Equal(t, "First rule", results[0].Description)
	assert.Equal(t, types.RuleStateEnabled, results[0].State)
	
	assert.Equal(t, "rule-2", results[1].Name)
	assert.Equal(t, types.RuleStateDisabled, results[1].State)
	
	mockClient.AssertExpectations(t)
}

// Test ListRulesE with default event bus
func TestListRulesE_MockedDefaultEventBus(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListRules", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListRulesOutput{
		Rules: []types.Rule{},
	}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	results, err := ListRulesE(ctx, "") // Empty event bus should use default

	require.NoError(t, err)
	assert.Empty(t, results)
	mockClient.AssertExpectations(t)
}

// Test ListRules panic variant
func TestListRules_MockedSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListRules", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListRulesOutput{
		Rules: []types.Rule{
			{
				Name: aws.String("rule-1"),
			},
		},
	}, nil)

	mockT := &MockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Client: mockClient,
	}

	results := ListRules(ctx, "default")

	assert.Len(t, results, 1)
	assert.Equal(t, "rule-1", results[0].Name)
	mockClient.AssertExpectations(t)
	mockT.AssertNotCalled(t, "Errorf")
	mockT.AssertNotCalled(t, "FailNow")
}

// Test updateRuleStateE helper function (covered via EnableRuleE/DisableRuleE above, but let's test edge cases)
func TestUpdateRuleStateE_MockedEmptyRuleName(t *testing.T) {
	ctx := &TestContext{T: t}

	err := EnableRuleE(ctx, "", "default") // Empty rule name

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
}

func TestUpdateRuleStateE_MockedDefaultEventBus(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("EnableRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.EnableRuleOutput{}, nil)

	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	err := EnableRuleE(ctx, "test-rule", "") // Empty event bus should use default

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}