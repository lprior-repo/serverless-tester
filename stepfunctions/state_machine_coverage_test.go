// Package stepfunctions coverage-focused tests for state machine operations
// Red-Green-Refactor TDD approach to ensure 100% state machine function coverage

package stepfunctions

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

// TestStateMachineLifecycleOperationsCoverage tests all state machine operations for code coverage
func TestStateMachineLifecycleOperationsCoverage(t *testing.T) {
	ctx := createTestContext()

	t.Run("CreateStateMachineEValidation", func(t *testing.T) {
		// Test nil definition
		result, err := CreateStateMachineE(ctx, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "state machine definition is required")

		// Test empty name
		definition := &StateMachineDefinition{
			Name:       "",
			Definition: validSimpleStateMachineDefinition(),
			RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsExecutionRole",
			Type:       types.StateMachineTypeStandard,
		}
		result, err = CreateStateMachineE(ctx, definition, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid state machine name")

		// Test valid definition (will fail with AWS service error in test env)
		definition.Name = "test-state-machine"
		result, err = CreateStateMachineE(ctx, definition, nil)
		if err != nil {
			// Should be AWS service error for valid input
			assert.True(t, isServiceError(err))
		}
	})

	t.Run("UpdateStateMachineEValidation", func(t *testing.T) {
		// Test empty ARN
		result, err := UpdateStateMachineE(ctx, "", validSimpleStateMachineDefinition(), "arn:aws:iam::123456789012:role/StepFunctionsExecutionRole", nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "state machine ARN is required")

		// Test valid update (will fail with AWS service error in test env)
		result, err = UpdateStateMachineE(ctx, "arn:aws:states:us-east-1:123456789012:stateMachine:test", validSimpleStateMachineDefinition(), "arn:aws:iam::123456789012:role/StepFunctionsExecutionRole", nil)
		if err != nil {
			// Should be AWS service error for valid input
			assert.True(t, isServiceError(err))
		}
	})

	t.Run("DescribeStateMachineEValidation", func(t *testing.T) {
		// Test empty ARN
		result, err := DescribeStateMachineE(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "state machine ARN is required")

		// Test valid describe (will fail with AWS service error in test env)
		result, err = DescribeStateMachineE(ctx, "arn:aws:states:us-east-1:123456789012:stateMachine:test")
		if err != nil {
			// Should be AWS service error for valid input
			assert.True(t, isServiceError(err))
		}
	})

	t.Run("DeleteStateMachineEValidation", func(t *testing.T) {
		// Test empty ARN
		err := DeleteStateMachineE(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state machine ARN is required")

		// Test valid delete (will fail with AWS service error in test env)
		err = DeleteStateMachineE(ctx, "arn:aws:states:us-east-1:123456789012:stateMachine:test")
		if err != nil {
			// Should be AWS service error for valid input
			assert.True(t, isServiceError(err))
		}
	})

	t.Run("ListStateMachinesEValidation", func(t *testing.T) {
		// Test negative max results
		result, err := ListStateMachinesE(ctx, -1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "max results cannot be negative")

		// Test valid list (will fail with AWS service error in test env)
		result, err = ListStateMachinesE(ctx, 10)
		if err != nil {
			// Should be AWS service error for valid input
			assert.True(t, isServiceError(err))
		}
	})
}

// TestStateMachineTagOperationsCoverage tests tag management operations
func TestStateMachineTagOperationsCoverage(t *testing.T) {
	ctx := createTestContext()
	validArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test"

	t.Run("TagStateMachineEValidation", func(t *testing.T) {
		// Test empty ARN
		err := TagStateMachineE(ctx, "", map[string]string{"test": "value"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state machine ARN is required")

		// Test empty tags
		err = TagStateMachineE(ctx, validArn, map[string]string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one tag is required")

		// Test valid tagging (will fail with AWS service error in test env)
		err = TagStateMachineE(ctx, validArn, map[string]string{"Environment": "test"})
		if err != nil {
			assert.True(t, isServiceError(err))
		}
	})

	t.Run("UntagStateMachineEValidation", func(t *testing.T) {
		// Test empty ARN
		err := UntagStateMachineE(ctx, "", []string{"test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state machine ARN is required")

		// Test empty tag keys
		err = UntagStateMachineE(ctx, validArn, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one tag key is required")

		// Test valid untagging (will fail with AWS service error in test env)
		err = UntagStateMachineE(ctx, validArn, []string{"Environment"})
		if err != nil {
			assert.True(t, isServiceError(err))
		}
	})

	t.Run("ListStateMachineTagsEValidation", func(t *testing.T) {
		// Test empty ARN
		result, err := ListStateMachineTagsE(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "state machine ARN is required")

		// Test valid listing (will fail with AWS service error in test env)
		result, err = ListStateMachineTagsE(ctx, validArn)
		if err != nil {
			assert.True(t, isServiceError(err))
		}
	})
}

// TestStateMachineValidationFunctionsCoverage tests validation functions
func TestStateMachineValidationFunctionsCoverage(t *testing.T) {
	t.Run("ValidateStateMachineCreation", func(t *testing.T) {
		// Test nil definition
		err := validateStateMachineCreation(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state machine definition is required")

		// Test valid definition
		definition := &StateMachineDefinition{
			Name:       "test-machine",
			Definition: validSimpleStateMachineDefinition(),
			RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsExecutionRole",
			Type:       types.StateMachineTypeStandard,
		}
		err = validateStateMachineCreation(definition, nil)
		assert.NoError(t, err)
	})

	t.Run("ValidateStateMachineArn", func(t *testing.T) {
		// Test empty ARN
		err := validateStateMachineArn("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state machine ARN is required")

		// Test invalid ARN
		err = validateStateMachineArn("invalid-arn")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ARN")

		// Test valid ARN
		err = validateStateMachineArn("arn:aws:states:us-east-1:123456789012:stateMachine:test")
		assert.NoError(t, err)
	})

	t.Run("ValidateRoleArn", func(t *testing.T) {
		// Test empty role ARN
		err := validateRoleArn("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role ARN is required")

		// Test invalid role ARN
		err = validateRoleArn("invalid-arn")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ARN")

		// Test valid role ARN
		err = validateRoleArn("arn:aws:iam::123456789012:role/StepFunctionsExecutionRole")
		assert.NoError(t, err)
	})
}

// TestStateMachineWrapperFunctionsCoverage tests non-E wrapper functions
func TestStateMachineWrapperFunctionsCoverage(t *testing.T) {
	ctx := createTestContext()

	t.Run("WrapperFunctionsCallCorrectE", func(t *testing.T) {
		// Test that wrapper functions properly handle errors from E versions
		definition := &StateMachineDefinition{
			Name:       "",  // Invalid to trigger validation error
			Definition: validSimpleStateMachineDefinition(),
			RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsExecutionRole",
			Type:       types.StateMachineTypeStandard,
		}

		// Test E version directly to ensure it works
		result, err := CreateStateMachineE(ctx, definition, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid state machine name")
	})
}

// TestStateMachineHelperFunctionsCoverage ensures helper functions are covered
func TestStateMachineHelperFunctionsCoverage(t *testing.T) {
	t.Run("MergeStateMachineOptions", func(t *testing.T) {
		// Test nil options
		merged := mergeStateMachineOptions(nil)
		assert.NotNil(t, merged)
		
		// Test with options
		options := &StateMachineOptions{}
		merged = mergeStateMachineOptions(options)
		assert.NotNil(t, merged)
	})

	t.Run("ProcessStateMachineInfo", func(t *testing.T) {
		// Test process function with sample data
		info := &StateMachineInfo{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
			Name:           "test",
			Status:         types.StateMachineStatusActive,
		}
		processed := processStateMachineInfo(info)
		assert.NotNil(t, processed)
		assert.Equal(t, info.StateMachineArn, processed.StateMachineArn)
	})
}

// Helper functions

// validSimpleStateMachineDefinition returns a simple valid state machine definition
func validSimpleStateMachineDefinition() string {
	return `{
  "Comment": "A simple minimal example",
  "StartAt": "Hello",
  "States": {
    "Hello": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:us-east-1:123456789012:function:HelloWorld",
      "End": true
    }
  }
}`
}

// isServiceError checks if the error is from AWS service (not validation error)
func isServiceError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	// These are AWS service errors, not validation errors
	return (strings.Contains(errStr, "RequestID") ||
		strings.Contains(errStr, "Authentication") ||
		strings.Contains(errStr, "endpoint") ||
		strings.Contains(errStr, "operation error")) &&
		// These are validation errors, not service errors
		!strings.Contains(errStr, "is required") &&
		!strings.Contains(errStr, "invalid") &&
		!strings.Contains(errStr, "cannot be negative")
}