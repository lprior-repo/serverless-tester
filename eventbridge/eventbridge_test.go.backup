package eventbridge

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPutEvent tests basic event publishing functionality
func TestPutEvent(t *testing.T) {
	// RED: This test will fail because PutEvent doesn't exist yet
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}

	event := CustomEvent{
		Source:      "test.service",
		DetailType:  "Test Event",
		Detail:      `{"key": "value"}`,
		EventBusName: "default",
	}

	// This should not panic and should return a valid result
	result := PutEvent(ctx, event)
	
	require.NotNil(t, result)
	assert.NotEmpty(t, result.EventID)
	assert.True(t, result.Success)
}

// TestPutEventE tests error handling variant
func TestPutEventE(t *testing.T) {
	// RED: This test will fail because PutEventE doesn't exist yet
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}

	event := CustomEvent{
		Source:      "test.service",
		DetailType:  "Test Event",
		Detail:      `{"key": "value"}`,
		EventBusName: "default",
	}

	result, err := PutEventE(ctx, event)
	
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.EventID)
	assert.True(t, result.Success)
}

// TestCreateRule tests EventBridge rule creation
func TestCreateRule(t *testing.T) {
	// RED: This test will fail because CreateRule doesn't exist yet
	ctx := &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}

	ruleConfig := RuleConfig{
		Name:         "test-rule",
		Description:  "Test rule for EventBridge",
		EventPattern: `{"source": ["test.service"]}`,
		State:        types.RuleStateEnabled,
		EventBusName: "default",
	}

	result := CreateRule(ctx, ruleConfig)
	
	require.NotNil(t, result)
	assert.Equal(t, "test-rule", result.Name)
	assert.NotEmpty(t, result.RuleArn)
	assert.Equal(t, types.RuleStateEnabled, result.State)
}

// TestBuildCustomEvent tests event builder functionality
func TestBuildCustomEvent(t *testing.T) {
	// RED: This test will fail because BuildCustomEvent doesn't exist yet
	event := BuildCustomEvent("test.service", "Test Event", map[string]interface{}{
		"key":   "value",
		"count": 42,
	})

	assert.Equal(t, "test.service", event.Source)
	assert.Equal(t, "Test Event", event.DetailType)
	assert.Contains(t, event.Detail, `"key":"value"`)
	assert.Contains(t, event.Detail, `"count":42`)
	assert.Equal(t, "default", event.EventBusName)
}

// TestValidateEventPattern tests pattern validation
func TestValidateEventPattern(t *testing.T) {
	// RED: This test will fail because ValidateEventPattern doesn't exist yet
	validPattern := `{"source": ["test.service"], "detail-type": ["Test Event"]}`
	invalidPattern := `{"source": "invalid-format"}`

	assert.True(t, ValidateEventPattern(validPattern))
	assert.False(t, ValidateEventPattern(invalidPattern))
}