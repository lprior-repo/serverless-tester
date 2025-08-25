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

// RED: Test CreateRuleE with valid configuration
func TestCreateRuleE_WithValidConfig_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := RuleConfig{
		Name:         "test-rule",
		Description:  "Test rule description",
		EventPattern: `{"source": ["aws.s3"]}`,
		State:        RuleStateEnabled,
		EventBusName: "custom-bus",
		Tags: map[string]string{
			"Environment": "test",
			"Purpose":     "testing",
		},
	}

	ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
	mockClient.On("PutRule", context.Background(), mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.Name == "test-rule" &&
			*input.Description == "Test rule description" &&
			*input.EventPattern == `{"source": ["aws.s3"]}` &&
			input.State == RuleStateEnabled &&
			*input.EventBusName == "custom-bus" &&
			len(input.Tags) == 2
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: &ruleArn,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateRuleE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, ruleArn, result.RuleArn)
	assert.Equal(t, RuleStateEnabled, result.State)
	assert.Equal(t, "Test rule description", result.Description)
	assert.Equal(t, "custom-bus", result.EventBusName)
	assert.Equal(t, `{"source": ["aws.s3"]}`, result.EventPattern)
	mockClient.AssertExpectations(t)
}

// RED: Test CreateRuleE with empty rule name
func TestCreateRuleE_WithEmptyRuleName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := RuleConfig{
		Name:         "", // Empty name
		Description:  "Test rule description",
		EventPattern: `{"source": ["aws.s3"]}`,
	}

	result, err := CreateRuleE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
	assert.Empty(t, result.Name)
}

// RED: Test CreateRuleE with invalid event pattern
func TestCreateRuleE_WithInvalidEventPattern_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := RuleConfig{
		Name:         "test-rule",
		Description:  "Test rule description",
		EventPattern: `{"source": "invalid-pattern"}`, // Invalid pattern (should be array)
	}

	result, err := CreateRuleE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event pattern")
	assert.Empty(t, result.Name)
}

// RED: Test CreateRuleE with default values
func TestCreateRuleE_WithMinimalConfig_UsesDefaults(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := RuleConfig{
		Name:        "test-rule",
		Description: "Test rule",
	}

	ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
	mockClient.On("PutRule", context.Background(), mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.Name == "test-rule" &&
			*input.EventBusName == DefaultEventBusName &&
			input.State == RuleStateEnabled
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: &ruleArn,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateRuleE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, DefaultEventBusName, result.EventBusName)
	assert.Equal(t, RuleStateEnabled, result.State)
	mockClient.AssertExpectations(t)
}

// RED: Test CreateRuleE with schedule expression
func TestCreateRuleE_WithScheduleExpression_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := RuleConfig{
		Name:               "scheduled-rule",
		Description:        "Scheduled rule",
		ScheduleExpression: "rate(5 minutes)",
	}

	ruleArn := "arn:aws:events:us-east-1:123456789012:rule/scheduled-rule"
	mockClient.On("PutRule", context.Background(), mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.Name == "scheduled-rule" &&
			*input.ScheduleExpression == "rate(5 minutes)"
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: &ruleArn,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateRuleE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "scheduled-rule", result.Name)
	assert.Equal(t, "rate(5 minutes)", result.ScheduleExpression)
	mockClient.AssertExpectations(t)
}

// RED: Test CreateRuleE with retry on transient error
func TestCreateRuleE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := RuleConfig{
		Name:        "test-rule",
		Description: "Test rule",
	}

	ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
	// First call fails, second succeeds
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.PutRuleOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("PutRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: &ruleArn,
	}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := CreateRuleE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, ruleArn, result.RuleArn)
	mockClient.AssertExpectations(t)
}

// RED: Test DeleteRuleE with valid rule
func TestDeleteRuleE_WithValidRule_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	// Mock ListTargetsByRule to return no targets
	mockClient.On("ListTargetsByRule", context.Background(), mock.MatchedBy(func(input *eventbridge.ListTargetsByRuleInput) bool {
		return *input.Rule == "test-rule" &&
			*input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{},
	}, nil)

	// Mock DeleteRule
	mockClient.On("DeleteRule", context.Background(), mock.MatchedBy(func(input *eventbridge.DeleteRuleInput) bool {
		return *input.Name == "test-rule" &&
			*input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.DeleteRuleOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := DeleteRuleE(ctx, "test-rule", "")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test DeleteRuleE with empty rule name
func TestDeleteRuleE_WithEmptyRuleName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	err := DeleteRuleE(ctx, "", DefaultEventBusName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
}

// RED: Test DeleteRuleE with targets removal failure
func TestDeleteRuleE_WithTargetsRemovalFailure_ReturnsError(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	// Mock ListTargetsByRule to return targets
	mockClient.On("ListTargetsByRule", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListTargetsByRuleOutput{
		Targets: []types.Target{
			{
				Id:  aws.String("target-1"),
				Arn: aws.String("arn:aws:lambda:us-east-1:123456789012:function:test"),
			},
		},
	}, nil)

	// Mock RemoveTargets to fail
	mockClient.On("RemoveTargets", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.RemoveTargetsOutput)(nil), errors.New("remove targets failed"))

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := DeleteRuleE(ctx, "test-rule", DefaultEventBusName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove targets")
	mockClient.AssertExpectations(t)
}

// RED: Test DescribeRuleE with valid rule
func TestDescribeRuleE_WithValidRule_ReturnsRuleDetails(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	ruleName := "test-rule"
	ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
	description := "Test rule description"
	eventPattern := `{"source": ["aws.s3"]}`
	
	mockClient.On("DescribeRule", context.Background(), mock.MatchedBy(func(input *eventbridge.DescribeRuleInput) bool {
		return *input.Name == "test-rule" &&
			*input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.DescribeRuleOutput{
		Name:         &ruleName,
		Arn:          &ruleArn,
		Description:  &description,
		EventPattern: &eventPattern,
		State:        RuleStateEnabled,
		EventBusName: aws.String(DefaultEventBusName),
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := DescribeRuleE(ctx, "test-rule", "")

	assert.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, ruleArn, result.RuleArn)
	assert.Equal(t, description, result.Description)
	assert.Equal(t, eventPattern, result.EventPattern)
	assert.Equal(t, RuleStateEnabled, result.State)
	assert.Equal(t, DefaultEventBusName, result.EventBusName)
	mockClient.AssertExpectations(t)
}

// RED: Test DescribeRuleE with empty rule name
func TestDescribeRuleE_WithEmptyRuleName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	result, err := DescribeRuleE(ctx, "", DefaultEventBusName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
	assert.Empty(t, result.Name)
}

// RED: Test EnableRuleE with valid rule
func TestEnableRuleE_WithValidRule_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	mockClient.On("EnableRule", context.Background(), mock.MatchedBy(func(input *eventbridge.EnableRuleInput) bool {
		return *input.Name == "test-rule" &&
			*input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.EnableRuleOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := EnableRuleE(ctx, "test-rule", "")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test DisableRuleE with valid rule
func TestDisableRuleE_WithValidRule_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	mockClient.On("DisableRule", context.Background(), mock.MatchedBy(func(input *eventbridge.DisableRuleInput) bool {
		return *input.Name == "test-rule" &&
			*input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.DisableRuleOutput{}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := DisableRuleE(ctx, "test-rule", "")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test updateRuleStateE with empty rule name
func TestUpdateRuleStateE_WithEmptyRuleName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}

	err := updateRuleStateE(ctx, "", DefaultEventBusName, RuleStateEnabled)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
}

// RED: Test updateRuleStateE with retry on failure
func TestUpdateRuleStateE_WithTransientError_RetriesAndSucceeds(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	// First call fails, second succeeds
	mockClient.On("EnableRule", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.EnableRuleOutput)(nil), errors.New("transient error")).Once()
	mockClient.On("EnableRule", context.Background(), mock.Anything, mock.Anything).Return(
		&eventbridge.EnableRuleOutput{}, nil).Once()

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	err := updateRuleStateE(ctx, "test-rule", DefaultEventBusName, RuleStateEnabled)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// RED: Test ListRulesE with valid event bus
func TestListRulesE_WithValidEventBus_ReturnsRules(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	ruleName1 := "rule-1"
	ruleName2 := "rule-2"
	ruleArn1 := "arn:aws:events:us-east-1:123456789012:rule/rule-1"
	ruleArn2 := "arn:aws:events:us-east-1:123456789012:rule/rule-2"
	
	mockClient.On("ListRules", context.Background(), mock.MatchedBy(func(input *eventbridge.ListRulesInput) bool {
		return *input.EventBusName == DefaultEventBusName
	}), mock.Anything).Return(&eventbridge.ListRulesOutput{
		Rules: []types.Rule{
			{
				Name:         &ruleName1,
				Arn:          &ruleArn1,
				State:        RuleStateEnabled,
				EventBusName: aws.String(DefaultEventBusName),
			},
			{
				Name:         &ruleName2,
				Arn:          &ruleArn2,
				State:        RuleStateDisabled,
				EventBusName: aws.String(DefaultEventBusName),
			},
		},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	results, err := ListRulesE(ctx, "")

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	
	assert.Equal(t, "rule-1", results[0].Name)
	assert.Equal(t, ruleArn1, results[0].RuleArn)
	assert.Equal(t, RuleStateEnabled, results[0].State)
	
	assert.Equal(t, "rule-2", results[1].Name)
	assert.Equal(t, ruleArn2, results[1].RuleArn)
	assert.Equal(t, RuleStateDisabled, results[1].State)
	
	mockClient.AssertExpectations(t)
}

// RED: Test ListRulesE with empty result
func TestListRulesE_WithEmptyResult_ReturnsEmptySlice(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	
	mockClient.On("ListRules", context.Background(), mock.Anything, mock.Anything).Return(&eventbridge.ListRulesOutput{
		Rules: []types.Rule{},
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	results, err := ListRulesE(ctx, DefaultEventBusName)

	assert.NoError(t, err)
	assert.Len(t, results, 0)
	mockClient.AssertExpectations(t)
}

// RED: Test UpdateRuleE with valid configuration
func TestUpdateRuleE_WithValidConfig_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := RuleConfig{
		Name:         "test-rule",
		Description:  "Updated rule description",
		EventPattern: `{"source": ["aws.ec2"]}`,
		State:        RuleStateDisabled,
		EventBusName: "custom-bus",
		Tags: map[string]string{
			"Environment": "production",
		},
	}

	ruleArn := "arn:aws:events:us-east-1:123456789012:rule/test-rule"
	mockClient.On("PutRule", context.Background(), mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.Name == "test-rule" &&
			*input.Description == "Updated rule description" &&
			*input.EventPattern == `{"source": ["aws.ec2"]}` &&
			input.State == RuleStateDisabled &&
			*input.EventBusName == "custom-bus" &&
			len(input.Tags) == 1
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: &ruleArn,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := UpdateRuleE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "test-rule", result.Name)
	assert.Equal(t, ruleArn, result.RuleArn)
	assert.Equal(t, RuleStateDisabled, result.State)
	assert.Equal(t, "Updated rule description", result.Description)
	assert.Equal(t, "custom-bus", result.EventBusName)
	assert.Equal(t, `{"source": ["aws.ec2"]}`, result.EventPattern)
	mockClient.AssertExpectations(t)
}

// RED: Test UpdateRuleE with empty rule name
func TestUpdateRuleE_WithEmptyRuleName_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := RuleConfig{
		Name:        "", // Empty name
		Description: "Updated rule",
	}

	result, err := UpdateRuleE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule name cannot be empty")
	assert.Empty(t, result.Name)
}

// RED: Test UpdateRuleE with invalid event pattern
func TestUpdateRuleE_WithInvalidEventPattern_ReturnsError(t *testing.T) {
	ctx := &TestContext{T: &TestingTMock{}}
	config := RuleConfig{
		Name:         "test-rule",
		Description:  "Updated rule",
		EventPattern: `{"source": "invalid-pattern"}`, // Invalid pattern
	}

	result, err := UpdateRuleE(ctx, config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event pattern")
	assert.Empty(t, result.Name)
}

// RED: Test UpdateRuleE with schedule expression
func TestUpdateRuleE_WithScheduleExpression_ReturnsSuccess(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	config := RuleConfig{
		Name:               "scheduled-rule",
		Description:        "Updated scheduled rule",
		ScheduleExpression: "rate(10 minutes)",
	}

	ruleArn := "arn:aws:events:us-east-1:123456789012:rule/scheduled-rule"
	mockClient.On("PutRule", context.Background(), mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return *input.Name == "scheduled-rule" &&
			*input.ScheduleExpression == "rate(10 minutes)" &&
			input.EventPattern == nil // Should be nil when schedule is used
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: &ruleArn,
	}, nil)

	ctx := &TestContext{Client: mockClient, T: &TestingTMock{}}
	result, err := UpdateRuleE(ctx, config)

	assert.NoError(t, err)
	assert.Equal(t, "scheduled-rule", result.Name)
	assert.Equal(t, "rate(10 minutes)", result.ScheduleExpression)
	mockClient.AssertExpectations(t)
}

// RED: Test panic versions of functions call FailNow on error
func TestCreateRule_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}
	config := RuleConfig{
		Name: "", // Empty name to cause error
	}

	// This should not panic but should call FailNow
	CreateRule(ctx, config)

	mockT.AssertExpectations(t)
}

// RED: Test DeleteRule panic version
func TestDeleteRule_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	DeleteRule(ctx, "", DefaultEventBusName) // Empty name to cause error

	mockT.AssertExpectations(t)
}

// RED: Test DescribeRule panic version
func TestDescribeRule_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}

	// This should not panic but should call FailNow
	DescribeRule(ctx, "", DefaultEventBusName) // Empty name to cause error

	mockT.AssertExpectations(t)
}

// RED: Test EnableRule panic version
func TestEnableRule_WithError_CallsFailNow(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("EnableRule", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.EnableRuleOutput)(nil), errors.New("test error"))

	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT, Client: mockClient}

	// This should not panic but should call FailNow
	EnableRule(ctx, "test-rule", DefaultEventBusName)

	mockT.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// RED: Test DisableRule panic version
func TestDisableRule_WithError_CallsFailNow(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("DisableRule", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.DisableRuleOutput)(nil), errors.New("test error"))

	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT, Client: mockClient}

	// This should not panic but should call FailNow
	DisableRule(ctx, "test-rule", DefaultEventBusName)

	mockT.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// RED: Test ListRules panic version
func TestListRules_WithError_CallsFailNow(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	mockClient.On("ListRules", context.Background(), mock.Anything, mock.Anything).Return(
		(*eventbridge.ListRulesOutput)(nil), errors.New("test error"))

	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT, Client: mockClient}

	// This should not panic but should call FailNow
	ListRules(ctx, DefaultEventBusName)

	mockT.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

// RED: Test UpdateRule panic version
func TestUpdateRule_WithError_CallsFailNow(t *testing.T) {
	mockT := &TestingTMock{}
	mockT.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return()
	mockT.On("FailNow").Return()

	ctx := &TestContext{T: mockT}
	config := RuleConfig{
		Name: "", // Empty name to cause error
	}

	// This should not panic but should call FailNow
	UpdateRule(ctx, config)

	mockT.AssertExpectations(t)
}