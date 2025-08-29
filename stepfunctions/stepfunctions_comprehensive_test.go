package stepfunctions

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === TDD Test Suite for Step Functions Comprehensive Coverage ===
// Following strict Red-Green-Refactor methodology for 100% coverage

// Test State Machine Operations Coverage

func TestCreateStateMachine_ComprehensiveValidation(t *testing.T) {
	t.Parallel()
	
	// Red: Test missing definition
	t.Run("FailsWithNilDefinition", func(t *testing.T) {
		ctx := createComprehensiveTestContext(t)
		_, err := CreateStateMachineE(ctx, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state machine definition is required")
	})
	
	// Red: Test invalid role ARN 
	t.Run("FailsWithInvalidRoleArn", func(t *testing.T) {
		ctx := createComprehensiveTestContext(t)
		definition := &StateMachineDefinition{
			Name:       "test-machine",
			Definition: `{"StartAt":"Pass","States":{"Pass":{"Type":"Pass","End":true}}}`,
			RoleArn:    "invalid-arn",
			Type:       StateMachineTypeStandard,
		}
		_, err := CreateStateMachineE(ctx, definition, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ARN")
	})
	
	// Green: Test valid definition with proper structure
	t.Run("ValidatesProperDefinitionStructure", func(t *testing.T) {
		ctx := createComprehensiveTestContext(t)
		definition := &StateMachineDefinition{
			Name:       "test-machine",
			Definition: `{"StartAt":"Pass","States":{"Pass":{"Type":"Pass","End":true}}}`,
			RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
			Type:       StateMachineTypeStandard,
			Tags:       map[string]string{"Environment": "test"},
		}
		
		// This will fail with AWS auth error, but validates input processing
		_, err := CreateStateMachineE(ctx, definition, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create state machine")
	})
}

// Test Input Builder Comprehensive Coverage

func TestInputBuilder_ComprehensiveOperations(t *testing.T) {
	t.Parallel()
	
	// Green: Test complex input building
	t.Run("BuildsComplexInput", func(t *testing.T) {
		input := NewInput().
			Set("string_field", "test_value").
			Set("int_field", 42).
			Set("bool_field", true).
			Set("array_field", []string{"item1", "item2"}).
			Set("object_field", map[string]interface{}{"nested": "value"})
		
		jsonStr, err := input.ToJSON()
		require.NoError(t, err)
		assert.NotEmpty(t, jsonStr)
		
		// Validate JSON structure
		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(jsonStr), &parsed)
		require.NoError(t, err)
		assert.Equal(t, "test_value", parsed["string_field"])
		assert.Equal(t, float64(42), parsed["int_field"]) // JSON numbers are float64
		assert.Equal(t, true, parsed["bool_field"])
	})
	
	// Green: Test conditional input building
	t.Run("BuildsConditionalInput", func(t *testing.T) {
		input := NewInput().
			SetIf(true, "included_field", "included_value").
			SetIf(false, "excluded_field", "excluded_value")
		
		includedValue, exists := input.Get("included_field")
		assert.True(t, exists)
		assert.Equal(t, "included_value", includedValue)
		
		_, exists = input.Get("excluded_field")
		assert.False(t, exists)
	})
	
	// Green: Test input merging with precedence
	t.Run("MergesInputWithPrecedence", func(t *testing.T) {
		input1 := NewInput().
			Set("shared_key", "original_value").
			Set("unique_key1", "value1")
		
		input2 := NewInput().
			Set("shared_key", "overridden_value").
			Set("unique_key2", "value2")
		
		merged := input1.Merge(input2)
		
		sharedValue, _ := merged.Get("shared_key")
		assert.Equal(t, "overridden_value", sharedValue, "Merged input should take precedence")
		
		_, exists1 := merged.Get("unique_key1")
		assert.True(t, exists1)
		
		_, exists2 := merged.Get("unique_key2")  
		assert.True(t, exists2)
	})
	
	// Green: Test type-safe getters
	t.Run("HandlesTypeSafeGetters", func(t *testing.T) {
		input := NewInput().
			Set("string_val", "test").
			Set("int_val", 42).
			Set("bool_val", true).
			Set("wrong_type", "not_an_int")
		
		// Test successful type conversions
		strVal, exists := input.GetString("string_val")
		assert.True(t, exists)
		assert.Equal(t, "test", strVal)
		
		intVal, exists := input.GetInt("int_val")
		assert.True(t, exists)
		assert.Equal(t, 42, intVal)
		
		boolVal, exists := input.GetBool("bool_val")
		assert.True(t, exists)
		assert.Equal(t, true, boolVal)
		
		// Test type mismatch
		_, exists = input.GetInt("wrong_type")
		assert.False(t, exists, "Should fail type conversion")
		
		// Test non-existent key
		_, exists = input.GetString("missing_key")
		assert.False(t, exists)
	})
}

// Test History Analysis Comprehensive Coverage

func TestHistoryAnalysis_ComprehensiveScenarios(t *testing.T) {
	t.Parallel()
	
	// Green: Test empty history analysis
	t.Run("HandlesEmptyHistory", func(t *testing.T) {
		analysis, err := AnalyzeExecutionHistoryE([]HistoryEvent{})
		require.NoError(t, err)
		assert.NotNil(t, analysis)
		assert.Equal(t, 0, analysis.TotalSteps)
		assert.Equal(t, 0, analysis.CompletedSteps)
		assert.Equal(t, 0, analysis.FailedSteps)
		assert.Equal(t, types.ExecutionStatusRunning, analysis.ExecutionStatus)
	})
	
	// Green: Test successful execution analysis
	t.Run("AnalyzesSuccessfulExecution", func(t *testing.T) {
		baseTime := time.Now()
		history := []HistoryEvent{
			{
				ID:        1,
				Type:      types.HistoryEventTypeExecutionStarted,
				Timestamp: baseTime,
				ExecutionStartedEventDetails: &ExecutionStartedEventDetails{
					Input:   `{"test": "value"}`,
					RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
				},
			},
			{
				ID:        2,
				Type:      types.HistoryEventTypeTaskStateEntered,
				Timestamp: baseTime.Add(1 * time.Second),
				StateEnteredEventDetails: &StateEnteredEventDetails{
					Name:  "ProcessTask",
					Input: `{"test": "value"}`,
				},
			},
			{
				ID:        3,
				Type:      types.HistoryEventTypeTaskSucceeded,
				Timestamp: baseTime.Add(2 * time.Second),
				TaskSucceededEventDetails: &TaskSucceededEventDetails{
					Resource:     "arn:aws:lambda:us-east-1:123456789012:function:test-function",
					ResourceType: "lambda",
					Output:       `{"result": "success"}`,
				},
			},
			{
				ID:        4,
				Type:      types.HistoryEventTypeExecutionSucceeded,
				Timestamp: baseTime.Add(3 * time.Second),
				ExecutionSucceededEventDetails: &ExecutionSucceededEventDetails{
					Output: `{"result": "success"}`,
				},
			},
		}
		
		analysis, err := AnalyzeExecutionHistoryE(history)
		require.NoError(t, err)
		assert.Equal(t, 4, analysis.TotalSteps)
		assert.Equal(t, 1, analysis.CompletedSteps) // Only TaskSucceeded counts as completed step
		assert.Equal(t, 0, analysis.FailedSteps)
		assert.Equal(t, 0, analysis.RetryAttempts)
		assert.Equal(t, types.ExecutionStatusSucceeded, analysis.ExecutionStatus)
		assert.True(t, analysis.TotalDuration > 0)
	})
}

// Test utility function coverage

func TestUtilityFunctions_ComprehensiveValidation(t *testing.T) {
	t.Parallel()
	
	// Green: Test ARN parsing functions
	t.Run("ParsesExecutionArnComponents", func(t *testing.T) {
		validArn := "arn:aws:states:us-east-1:123456789012:execution:test-machine:test-execution"
		stateMachine, execution, err := parseExecutionArn(validArn)
		require.NoError(t, err)
		assert.Equal(t, "test-machine", stateMachine)
		assert.Equal(t, "test-execution", execution)
	})
	
	// Red: Test invalid ARN parsing
	t.Run("FailsForInvalidArnParsing", func(t *testing.T) {
		invalidArn := "invalid:arn:format"
		_, _, err := parseExecutionArn(invalidArn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid execution ARN format")
	})
	
	// Green: Test execution completion checking
	t.Run("ChecksExecutionCompletionStates", func(t *testing.T) {
		completedStates := []types.ExecutionStatus{
			types.ExecutionStatusSucceeded,
			types.ExecutionStatusFailed,
			types.ExecutionStatusTimedOut,
			types.ExecutionStatusAborted,
		}
		
		for _, status := range completedStates {
			assert.True(t, isExecutionComplete(status), "Status %s should be complete", status)
		}
		
		assert.False(t, isExecutionComplete(types.ExecutionStatusRunning), "Running should not be complete")
	})
	
	// Green: Test backoff delay calculation
	t.Run("CalculatesExponentialBackoffCorrectly", func(t *testing.T) {
		config := RetryConfig{
			BaseDelay:  1 * time.Second,
			MaxDelay:   30 * time.Second,
			Multiplier: 2.0,
		}
		
		// Test progression
		delay1 := calculateBackoffDelay(0, config)
		delay2 := calculateBackoffDelay(1, config)
		delay3 := calculateBackoffDelay(2, config)
		
		assert.Equal(t, 1*time.Second, delay1)
		assert.Equal(t, 2*time.Second, delay2)
		assert.Equal(t, 4*time.Second, delay3)
		
		// Test max cap
		delay10 := calculateBackoffDelay(10, config)
		assert.Equal(t, 30*time.Second, delay10)
	})
}

// Helper function to create comprehensive test context 
func createComprehensiveTestContext(t TestingT) *TestContext {
	// Create a test context directly since we have the interface  
	return &TestContext{
		T:         t,
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
}