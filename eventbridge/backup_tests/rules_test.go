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

func TestCreateRuleE_Success(t *testing.T) {
	// Test successful rule creation
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "test-rule",
		Description:  "Test rule description",
		EventPattern: `{"source": ["test.service"]}`,
		State:        types.RuleStateEnabled,
		EventBusName: "default",
		Tags: map[string]string{
			"Environment": "test",
			"Project":     "eventbridge-tests",
		},
	}

	expectedRuleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"

	// Mock successful response
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.Name == "test-rule" &&
			*input.Description == "Test rule description" &&
			*input.EventPattern == `{"source": ["test.service"]}` &&
			*input.EventBusName == "default" &&
			input.State == types.RuleStateEnabled &&
			len(input.Tags) == 2
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String(expectedRuleArn),
	}, nil)

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, expectedRuleArn, result.RuleArn)
	assert.Equal(t, types.RuleStateEnabled, result.State)
	assert.Equal(t, "Test rule description", result.Description)
	assert.Equal(t, "default", result.EventBusName)
	mockClient.AssertExpectations(t)
}

func TestCreateRuleE_WithScheduleExpression(t *testing.T) {
	// Test rule creation with schedule expression
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	config := RuleConfig{
		Name:               "scheduled-rule",
		Description:        "Scheduled rule",
		ScheduleExpression: "rate(5 minutes)",
		State:              types.RuleStateEnabled,
	}

	// Mock successful response
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.ScheduleExpression == "rate(5 minutes)" &&
			input.EventPattern == nil
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/scheduled-rule"),
	}, nil)

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "scheduled-rule", result.Name)
	mockClient.AssertExpectations(t)
}

func TestCreateRuleE_DefaultValues(t *testing.T) {
	// Test rule creation with default values
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "test-rule",
		EventPattern: `{"source": ["test.service"]}`,
		// EventBusName not set - should default to "default"
		// State not set - should default to enabled
	}

	// Mock successful response
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.EventBusName == DefaultEventBusName &&
			input.State == RuleStateEnabled
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
	}, nil)

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, DefaultEventBusName, result.EventBusName)
	assert.Equal(t, types.RuleStateEnabled, result.State)
	mockClient.AssertExpectations(t)
}

func TestCreateRuleE_EmptyRuleName(t *testing.T) {
	// Test validation for empty rule name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	config := RuleConfig{
		Name:         "",
		EventPattern: `{"source": ["test.service"]}`,
	}

	result, err := CreateRuleE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
	assert.Equal(t, RuleResult{}, result)
}

func TestCreateRuleE_InvalidEventPattern(t *testing.T) {
	// Test validation for invalid event pattern
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	config := RuleConfig{
		Name:         "test-rule",
		EventPattern: `{"source": "invalid-format"}`, // Should be array
	}

	result, err := CreateRuleE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event pattern")
	assert.Equal(t, RuleResult{}, result)
}

func TestCreateRuleE_ServiceError(t *testing.T) {
	// Test AWS service error
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "test-rule",
		EventPattern: `{"source": ["test.service"]}`,
	}

	serviceErr := errors.New("service error")
	mockClient.On("PutRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	result, err := CreateRuleE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create rule after 3 attempts")
	assert.Equal(t, RuleResult{}, result)
	mockClient.AssertExpectations(t)
}

func TestDeleteRuleE_Success(t *testing.T) {
	// Test successful rule deletion
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock RemoveTargets (called first)
	mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
		return *input.Rule == "test-rule" && *input.EventBusName == "default"
	}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{}, // No targets to remove
	}, nil)

	// Mock successful DeleteRule
	mockClient.On("DeleteRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DeleteRuleInput) bool {
		return *input.Name == "test-rule" && *input.EventBusName == "default"
	}), mock.Anything).Return(&eventbridge.DeleteRuleOutput{}, nil)

	err := DeleteRuleE(ctx, "test-rule", "default")

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDeleteRuleE_EmptyRuleName(t *testing.T) {
	// Test validation for empty rule name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := DeleteRuleE(ctx, "", "default")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
}

func TestDeleteRuleE_DefaultEventBus(t *testing.T) {
	// Test rule deletion with default event bus
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	// Mock RemoveTargets
	mockClient.On("ListTargetsByRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
		return *input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)

	// Mock successful DeleteRule
	mockClient.On("DeleteRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DeleteRuleInput) bool {
		return *input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.DeleteRuleOutput{}, nil)

	err := DeleteRuleE(ctx, "test-rule", "") // Empty event bus should default

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDescribeRuleE_Success(t *testing.T) {
	// Test successful rule description
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	expectedResponse := &eventbridge.DescribeRuleOutput{
		Name:         aws.String("test-rule"),
		Arn:          aws.String("arn:aws:events:us-east-1:123456789012:rule/test-rule"),
		EventPattern: aws.String(`{"source": ["test.service"]}`),
		Description:  aws.String("Test rule description"),
		State:        types.RuleStateEnabled,
		EventBusName: aws.String("default"),
	}

	mockClient.On("DescribeRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DescribeRuleInput) bool {
		return *input.Name == "test-rule" && *input.EventBusName == "default"
	}), mock.Anything).Return(expectedResponse, nil)

	result, err := DescribeRuleE(ctx, "test-rule", "default")

	require.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, "arn:aws:events:us-east-1:123456789012:rule/test-rule", result.RuleArn)
	assert.Equal(t, "Test rule description", result.Description)
	assert.Equal(t, types.RuleStateEnabled, result.State)
	assert.Equal(t, "default", result.EventBusName)
	mockClient.AssertExpectations(t)
}

func TestDescribeRuleE_EmptyRuleName(t *testing.T) {
	// Test validation for empty rule name
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	result, err := DescribeRuleE(ctx, "", "default")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
	assert.Equal(t, RuleResult{}, result)
}

func TestEnableRuleE_Success(t *testing.T) {
	// Test successful rule enable
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	mockClient.On("EnableRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.EnableRuleInput) bool {
		return *input.Name == "test-rule" && *input.EventBusName == "default"
	}), mock.Anything).Return(&eventbridge.EnableRuleOutput{}, nil)

	err := EnableRuleE(ctx, "test-rule", "default")

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDisableRuleE_Success(t *testing.T) {
	// Test successful rule disable
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	mockClient.On("DisableRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.DisableRuleInput) bool {
		return *input.Name == "test-rule" && *input.EventBusName == "default"
	}), mock.Anything).Return(&eventbridge.DisableRuleOutput{}, nil)

	err := DisableRuleE(ctx, "test-rule", "default")

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestListRulesE_Success(t *testing.T) {
	// Test successful rule listing
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	expectedResponse := &eventbridge.ListRulesOutput{
		Rules: []types.Rule{
			{
				Name:         aws.String("rule-1"),
				Arn:          aws.String("arn:aws:events:us-east-1:123456789012:rule/rule-1"),
				Description:  aws.String("First rule"),
				State:        types.RuleStateEnabled,
				EventBusName: aws.String("default"),
			},
			{
				Name:         aws.String("rule-2"),
				Arn:          aws.String("arn:aws:events:us-east-1:123456789012:rule/rule-2"),
				Description:  aws.String("Second rule"),
				State:        types.RuleStateDisabled,
				EventBusName: aws.String("default"),
			},
		},
	}

	mockClient.On("ListRules", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListRulesInput) bool {
		return *input.EventBusName == "default"
	}), mock.Anything).Return(expectedResponse, nil)

	results, err := ListRulesE(ctx, "default")

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "rule-1", results[0].Name)
	assert.Equal(t, types.RuleStateEnabled, results[0].State)
	assert.Equal(t, "rule-2", results[1].Name)
	assert.Equal(t, types.RuleStateDisabled, results[1].State)
	mockClient.AssertExpectations(t)
}

func TestListRulesE_DefaultEventBus(t *testing.T) {
	// Test listing rules with default event bus
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	mockClient.On("ListRules", mock.Anything, mock.MatchedBy(func(input *eventbridge.ListRulesInput) bool {
		return *input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.ListRulesOutput{
		Rules: []types.Rule{},
	}, nil)

	results, err := ListRulesE(ctx, "") // Empty should default

	require.NoError(t, err)
	assert.Empty(t, results)
	mockClient.AssertExpectations(t)
}

func TestUpdateRuleStateE_EmptyRuleName(t *testing.T) {
	// Test validation for empty rule name in state update
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	err := updateRuleStateE(ctx, "", "default", types.RuleStateEnabled)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
}

func TestCreateRule_PanicOnError(t *testing.T) {
	// Test that CreateRule panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	config := RuleConfig{
		Name:         "", // Empty name should cause validation error
		EventPattern: `{"source": ["test.service"]}`,
	}

	result := CreateRule(ctx, config)

	assert.Equal(t, RuleResult{}, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestDeleteRule_PanicOnError(t *testing.T) {
	// Test that DeleteRule panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	DeleteRule(ctx, "", "default") // Empty name should cause error

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestDescribeRule_PanicOnError(t *testing.T) {
	// Test that DescribeRule panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	result := DescribeRule(ctx, "", "default") // Empty name should cause error

	assert.Equal(t, RuleResult{}, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestEnableRule_PanicOnError(t *testing.T) {
	// Test that EnableRule panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	EnableRule(ctx, "", "default") // Empty name should cause error

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestDisableRule_PanicOnError(t *testing.T) {
	// Test that DisableRule panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	DisableRule(ctx, "", "default") // Empty name should cause error

	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

func TestListRules_PanicOnError(t *testing.T) {
	// Test that ListRules panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: mockClient,
	}

	serviceErr := errors.New("service error")
	mockClient.On("ListRules", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	result := ListRules(ctx, "default")

	assert.Nil(t, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
	mockClient.AssertExpectations(t)
}

func TestUpdateRuleE_Success(t *testing.T) {
	// RED: Test for successful rule update
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "test-rule",
		Description:  "Updated rule description",
		EventPattern: `{"source": ["updated.service"]}`,
		State:        types.RuleStateDisabled,
		EventBusName: "default",
		Tags: map[string]string{
			"Environment": "production",
			"Version":     "2.0",
		},
	}

	expectedRuleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"

	// Mock successful response for update (PutRule)
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.Name == "test-rule" &&
			*input.Description == "Updated rule description" &&
			*input.EventPattern == `{"source": ["updated.service"]}` &&
			*input.EventBusName == "default" &&
			input.State == types.RuleStateDisabled &&
			len(input.Tags) == 2
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String(expectedRuleArn),
	}, nil)

	result, err := UpdateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, expectedRuleArn, result.RuleArn)
	assert.Equal(t, types.RuleStateDisabled, result.State)
	assert.Equal(t, "Updated rule description", result.Description)
	assert.Equal(t, "default", result.EventBusName)
	mockClient.AssertExpectations(t)
}

func TestUpdateRuleE_EmptyRuleName(t *testing.T) {
	// Test validation for empty rule name in update
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	config := RuleConfig{
		Name:         "",
		Description:  "Updated description",
		EventPattern: `{"source": ["test.service"]}`,
	}

	result, err := UpdateRuleE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
	assert.Equal(t, RuleResult{}, result)
}

func TestUpdateRuleE_InvalidEventPattern(t *testing.T) {
	// Test validation for invalid event pattern in update
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	config := RuleConfig{
		Name:         "test-rule",
		Description:  "Updated description",
		EventPattern: `{"source": "invalid-format"}`, // Should be array
	}

	result, err := UpdateRuleE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event pattern")
	assert.Equal(t, RuleResult{}, result)
}

func TestUpdateRuleE_ServiceError(t *testing.T) {
	// Test AWS service error during update
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Region: "us-east-1",
		Client: mockClient,
	}

	config := RuleConfig{
		Name:         "test-rule",
		Description:  "Updated description",
		EventPattern: `{"source": ["test.service"]}`,
	}

	serviceErr := errors.New("service error")
	mockClient.On("PutRule", mock.Anything, mock.Anything, mock.Anything).Return(nil, serviceErr)

	result, err := UpdateRuleE(ctx, config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update rule after 3 attempts")
	assert.Equal(t, RuleResult{}, result)
	mockClient.AssertExpectations(t)
}

func TestUpdateRule_PanicOnError(t *testing.T) {
	// Test that UpdateRule panics on error (Terratest pattern)
	mockT := &mockTestingT{}
	ctx := &TestContext{
		T:      mockT,
		Region: "us-east-1",
		Client: &MockEventBridgeClient{},
	}

	config := RuleConfig{
		Name:         "", // Empty name should cause validation error
		Description:  "Updated description",
		EventPattern: `{"source": ["test.service"]}`,
	}

	result := UpdateRule(ctx, config)

	assert.Equal(t, RuleResult{}, result)
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}