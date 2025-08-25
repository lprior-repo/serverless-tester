package eventbridge

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestCreateRuleE_EventPatternValidation tests comprehensive event pattern validation
func TestCreateRuleE_EventPatternValidation(t *testing.T) {
	ctx := &TestContext{T: t, Region: "us-east-1"}

	tests := []struct {
		name           string
		eventPattern   string
		shouldBeValid  bool
		errorContains  string
	}{
		{
			name:          "valid simple pattern",
			eventPattern:  `{"source": ["test.service"]}`,
			shouldBeValid: true,
		},
		{
			name:          "valid complex pattern",
			eventPattern:  `{"source": ["test.service"], "detail-type": ["Test Event"], "detail": {"state": ["created"]}}`,
			shouldBeValid: true,
		},
		{
			name:          "valid pattern with nested fields",
			eventPattern:  `{"detail": {"user": {"id": [1, 2, 3]}, "action": ["login", "logout"]}}`,
			shouldBeValid: true,
		},
		{
			name:          "invalid pattern - source as string",
			eventPattern:  `{"source": "test.service"}`,
			shouldBeValid: false,
			errorContains: "invalid event pattern",
		},
		{
			name:          "invalid pattern - malformed JSON",
			eventPattern:  `{"source": ["test.service"`,
			shouldBeValid: false,
			errorContains: "invalid event pattern",
		},
		{
			name:          "valid pattern - empty JSON (matches all)",
			eventPattern:  `{}`,
			shouldBeValid: true,
		},
		{
			name:          "invalid pattern - non-JSON",
			eventPattern:  `this is not json`,
			shouldBeValid: false,
			errorContains: "invalid event pattern",
		},
		{
			name:          "invalid pattern - null values",
			eventPattern:  `{"source": null}`,
			shouldBeValid: false,
			errorContains: "invalid event pattern",
		},
		{
			name:          "valid pattern - empty array values (matches any)",
			eventPattern:  `{"source": []}`,
			shouldBeValid: true,
		},
		{
			name:          "valid pattern - account field",
			eventPattern:  `{"account": ["123456789012"], "source": ["test.service"]}`,
			shouldBeValid: true,
		},
		{
			name:          "valid pattern - region field",
			eventPattern:  `{"region": ["us-east-1"], "source": ["test.service"]}`,
			shouldBeValid: true,
		},
		{
			name:          "valid pattern - time field",
			eventPattern:  `{"time": ["2023-01-01T00:00:00Z"], "source": ["test.service"]}`,
			shouldBeValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := RuleConfig{
				Name:         "test-rule",
				EventPattern: tt.eventPattern,
			}

			_, err := CreateRuleE(ctx, config)

			if tt.shouldBeValid {
				// For valid patterns, we expect a different error (no AWS client)
				// but NOT a pattern validation error
				require.Error(t, err)
				assert.NotContains(t, err.Error(), "invalid event pattern")
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			}
		})
	}
}

// TestCreateRuleE_ScheduleExpressionValidation tests schedule expression validation
func TestCreateRuleE_ScheduleExpressionValidation(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	tests := []struct {
		name               string
		scheduleExpression string
		shouldSucceed      bool
	}{
		{
			name:               "valid rate expression",
			scheduleExpression: "rate(5 minutes)",
			shouldSucceed:      true,
		},
		{
			name:               "valid cron expression",
			scheduleExpression: "cron(0 10 * * ? *)",
			shouldSucceed:      true,
		},
		{
			name:               "valid rate with hours",
			scheduleExpression: "rate(2 hours)",
			shouldSucceed:      true,
		},
		{
			name:               "valid rate with days",
			scheduleExpression: "rate(1 day)",
			shouldSucceed:      true,
		},
		{
			name:               "complex cron expression",
			scheduleExpression: "cron(15 10 ? * 6L 2022-2030)",
			shouldSucceed:      true,
		},
		{
			name:               "empty schedule expression",
			scheduleExpression: "",
			shouldSucceed:      true,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
				if tt.scheduleExpression == "" {
					return input.ScheduleExpression == nil
				}
				return input.ScheduleExpression != nil && *input.ScheduleExpression == tt.scheduleExpression
			}), mock.Anything).Return(&eventbridge.PutRuleOutput{
				RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/schedule-rule-" + string(rune(i))),
			}, nil).Once()

			config := RuleConfig{
				Name:               "schedule-rule-" + string(rune(i)),
				ScheduleExpression: tt.scheduleExpression,
			}

			result, err := CreateRuleE(ctx, config)

			if tt.shouldSucceed {
				require.NoError(t, err)
				assert.Contains(t, result.Name, "schedule-rule")
			} else {
				require.Error(t, err)
			}
		})
	}

	mockClient.AssertExpectations(t)
}

// TestCreateRuleE_EventPatternAndScheduleExpression tests rules with both patterns
func TestCreateRuleE_EventPatternAndScheduleExpression(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// AWS allows rules with both event pattern AND schedule expression
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return input.EventPattern != nil && 
			   input.ScheduleExpression != nil &&
			   *input.EventPattern == `{"source": ["test.service"]}` &&
			   *input.ScheduleExpression == "rate(5 minutes)"
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/hybrid-rule"),
	}, nil)

	config := RuleConfig{
		Name:               "hybrid-rule",
		EventPattern:       `{"source": ["test.service"]}`,
		ScheduleExpression: "rate(5 minutes)",
	}

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "hybrid-rule", result.Name)
	mockClient.AssertExpectations(t)
}

// TestCreateRuleE_NoPatternNoSchedule tests rules without pattern or schedule
func TestCreateRuleE_NoPatternNoSchedule(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	// Rule without pattern or schedule should still be allowed (matches all events)
	mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
		return input.EventPattern == nil && input.ScheduleExpression == nil
	}), mock.Anything).Return(&eventbridge.PutRuleOutput{
		RuleArn: aws.String("arn:aws:events:us-east-1:123456789012:rule/catch-all-rule"),
	}, nil)

	config := RuleConfig{
		Name:        "catch-all-rule",
		Description: "Rule that matches all events",
	}

	result, err := CreateRuleE(ctx, config)

	require.NoError(t, err)
	assert.Equal(t, "catch-all-rule", result.Name)
	mockClient.AssertExpectations(t)
}

// TestValidateEventPattern_Comprehensive tests the ValidateEventPattern function directly
func TestValidateEventPattern_Comprehensive(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		expectValid bool
	}{
		{
			name:        "valid simple pattern",
			pattern:     `{"source": ["test"]}`,
			expectValid: true,
		},
		{
			name:        "valid complex pattern",
			pattern:     `{"source": ["test"], "detail-type": ["event"], "detail": {"key": ["value"]}}`,
			expectValid: true,
		},
		{
			name:        "invalid - source as string",
			pattern:     `{"source": "test"}`,
			expectValid: false,
		},
		{
			name:        "valid - empty object (matches all events)",
			pattern:     `{}`,
			expectValid: true,
		},
		{
			name:        "invalid - malformed JSON",
			pattern:     `{"source": ["test"`,
			expectValid: false,
		},
		{
			name:        "invalid - null values",
			pattern:     `{"source": null}`,
			expectValid: false,
		},
		{
			name:        "valid - empty arrays (matches any value)",
			pattern:     `{"source": []}`,
			expectValid: true,
		},
		{
			name:        "valid - numeric values in arrays",
			pattern:     `{"detail": {"count": [1, 2, 3]}}`,
			expectValid: true,
		},
		{
			name:        "valid - boolean values in arrays",
			pattern:     `{"detail": {"active": [true, false]}}`,
			expectValid: true,
		},
		{
			name:        "valid - AWS standard fields",
			pattern:     `{"account": ["123456789012"], "region": ["us-east-1"], "source": ["test"]}`,
			expectValid: true,
		},
		{
			name:        "empty string pattern",
			pattern:     "",
			expectValid: false, // Empty string patterns are not allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEventPattern(tt.pattern)
			assert.Equal(t, tt.expectValid, result, "Pattern: %s", tt.pattern)
		})
	}
}

// TestCreateRuleE_EventBusVariations tests rules on different event buses
func TestCreateRuleE_EventBusVariations(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{T: t, Region: "us-east-1", Client: mockClient}

	tests := []struct {
		name            string
		eventBusName    string
		expectedBusName string
	}{
		{
			name:            "default event bus explicitly set",
			eventBusName:    "default",
			expectedBusName: "default",
		},
		{
			name:            "custom event bus",
			eventBusName:    "my-custom-bus",
			expectedBusName: "my-custom-bus",
		},
		{
			name:            "empty event bus defaults to default",
			eventBusName:    "",
			expectedBusName: "default",
		},
		{
			name:            "partner event bus",
			eventBusName:    "aws.partner/example.com/123456789012/my-bus",
			expectedBusName: "aws.partner/example.com/123456789012/my-bus",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.On("PutRule", mock.Anything, mock.MatchedBy(func(input *eventbridge.PutRuleInput) bool {
				return *input.EventBusName == tt.expectedBusName
			}), mock.Anything).Return(&eventbridge.PutRuleOutput{
				RuleArn: aws.String(fmt.Sprintf("arn:aws:events:us-east-1:123456789012:rule/bus-rule-%d", i)),
			}, nil).Once()

			config := RuleConfig{
				Name:         fmt.Sprintf("bus-rule-%d", i),
				EventPattern: `{"source": ["test"]}`,
				EventBusName: tt.eventBusName,
			}

			result, err := CreateRuleE(ctx, config)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedBusName, result.EventBusName)
		})
	}

	mockClient.AssertExpectations(t)
}