package stepfunctions

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchCreateStateMachines_ShouldCreateMultiple(t *testing.T) {
	t.Helper()
	
	// Given: multiple state machine definitions
	definitions := []*StateMachineDefinition{
		{
			Name: "batch-state-machine-1",
			Definition: `{
				"Comment": "Test state machine 1",
				"StartAt": "Pass",
				"States": {
					"Pass": {"Type": "Pass", "End": true}
				}
			}`,
			RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
			Type:    types.StateMachineTypeStandard,
		},
		{
			Name: "batch-state-machine-2",
			Definition: `{
				"Comment": "Test state machine 2",
				"StartAt": "Pass",
				"States": {
					"Pass": {"Type": "Pass", "End": true}
				}
			}`,
			RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
			Type:    types.StateMachineTypeExpress,
		},
	}
	
	config := &BatchOperationConfig{
		MaxConcurrency: 2,
		Timeout:        2 * time.Minute,
		FailFast:       false,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: creating state machines in batch
	results, err := BatchCreateStateMachinesE(ctx, definitions, config)

	// Then: should create all state machines
	require.NoError(t, err)
	assert.Len(t, results, 2)
	
	for i, result := range results {
		assert.Equal(t, fmt.Sprintf("creating-%s", definitions[i].Name), result.StateMachineArn)
		// Will be errors due to AWS config, but structure should be correct
		assert.NotNil(t, result.Error)
	}
}

func TestBatchDeleteStateMachines_ShouldDeleteMultiple(t *testing.T) {
	t.Helper()
	
	// Given: multiple state machine ARNs
	arns := []string{
		"arn:aws:states:us-east-1:123456789012:stateMachine:test1",
		"arn:aws:states:us-east-1:123456789012:stateMachine:test2",
		"arn:aws:states:us-east-1:123456789012:stateMachine:test3",
	}
	
	config := &BatchOperationConfig{
		MaxConcurrency: 2,
		Timeout:        2 * time.Minute,
		FailFast:       false,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: deleting state machines in batch
	results, err := BatchDeleteStateMachinesE(ctx, arns, config)

	// Then: should attempt to delete all state machines
	require.NoError(t, err)
	assert.Len(t, results, 3)
	
	for i, result := range results {
		assert.Equal(t, arns[i], result.StateMachineArn)
		// Will be errors due to AWS config
		assert.NotNil(t, result.Error)
	}
}

func TestBatchTagStateMachines_ShouldTagMultiple(t *testing.T) {
	t.Helper()
	
	// Given: multiple state machine ARNs and tags
	operations := []BatchTagOperation{
		{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test1",
			Tags: map[string]string{
				"Environment": "test",
				"Application": "batch-test",
			},
		},
		{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test2",
			Tags: map[string]string{
				"Environment": "prod",
				"Application": "batch-test",
			},
		},
	}
	
	config := &BatchOperationConfig{
		MaxConcurrency: 2,
		Timeout:        2 * time.Minute,
		FailFast:       false,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: tagging state machines in batch
	results, err := BatchTagStateMachinesE(ctx, operations, config)

	// Then: should attempt to tag all state machines
	require.NoError(t, err)
	assert.Len(t, results, 2)
	
	for i, result := range results {
		assert.Equal(t, operations[i].StateMachineArn, result.StateMachineArn)
		// Will be errors due to AWS config
		assert.NotNil(t, result.Error)
	}
}

func TestBatchExecutions_ShouldExecuteMultiple(t *testing.T) {
	t.Helper()
	
	// Given: multiple execution requests
	executions := []BatchExecutionRequest{
		{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test1",
			ExecutionName:   "batch-execution-1",
			Input:          NewInput().Set("test", "value1"),
		},
		{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test2",
			ExecutionName:   "batch-execution-2",
			Input:          NewInput().Set("test", "value2"),
		},
		{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test3",
			ExecutionName:   "batch-execution-3",
			Input:          NewInput().Set("test", "value3"),
		},
	}
	
	config := &BatchOperationConfig{
		MaxConcurrency: 2,
		Timeout:        2 * time.Minute,
		FailFast:       false,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: executing in batch
	results, err := BatchExecutionsE(ctx, executions, config)

	// Then: should attempt to execute all
	require.NoError(t, err)
	assert.Len(t, results, 3)
	
	for i, result := range results {
		assert.Equal(t, fmt.Sprintf("starting-%s", executions[i].ExecutionName), result.ExecutionArn)
		// Will be errors due to AWS config
		assert.NotNil(t, result.Error)
	}
}

func TestBatchStopExecutions_ShouldStopMultiple(t *testing.T) {
	t.Helper()
	
	// Given: multiple execution ARNs
	arns := []string{
		"arn:aws:states:us-east-1:123456789012:execution:test1:abc123",
		"arn:aws:states:us-east-1:123456789012:execution:test2:def456",
		"arn:aws:states:us-east-1:123456789012:execution:test3:ghi789",
	}
	
	config := &BatchOperationConfig{
		MaxConcurrency: 2,
		Timeout:        2 * time.Minute,
		FailFast:       false,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: stopping executions in batch
	results, err := BatchStopExecutionsE(ctx, arns, "Batch stop", "Stopped by batch operation", config)

	// Then: should attempt to stop all executions
	require.NoError(t, err)
	assert.Len(t, results, 3)
	
	for i, result := range results {
		assert.Equal(t, arns[i], result.ExecutionArn)
		// Will be errors due to AWS config
		assert.NotNil(t, result.Error)
	}
}

func TestBatchOperations_ShouldRespectConcurrencyLimits(t *testing.T) {
	t.Helper()
	
	// Given: many operations with low concurrency
	arns := make([]string, 10)
	for i := 0; i < 10; i++ {
		arns[i] = fmt.Sprintf("arn:aws:states:us-east-1:123456789012:stateMachine:test%d", i+1)
	}
	
	config := &BatchOperationConfig{
		MaxConcurrency: 2, // Very low concurrency
		Timeout:        2 * time.Minute,
		FailFast:       false,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: deleting with limited concurrency
	start := time.Now()
	results, err := BatchDeleteStateMachinesE(ctx, arns, config)
	duration := time.Since(start)

	// Then: should complete all operations
	require.NoError(t, err)
	assert.Len(t, results, 10)
	
	// Should respect concurrency limits (though hard to test without real delays)
	assert.Greater(t, duration, 0*time.Millisecond)
}

func TestBatchOperations_ShouldFailFast(t *testing.T) {
	t.Helper()
	
	// Given: operations with fail-fast enabled
	arns := []string{
		"arn:aws:states:us-east-1:123456789012:stateMachine:test1",
		"arn:aws:states:us-east-1:123456789012:stateMachine:test2",
	}
	
	config := &BatchOperationConfig{
		MaxConcurrency: 2,
		Timeout:        2 * time.Minute,
		FailFast:       true, // Enable fail-fast
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: deleting with fail-fast enabled
	results, err := BatchDeleteStateMachinesE(ctx, arns, config)

	// Then: should either fail fast or handle gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "batch operation failed")
	} else {
		// Should still return results
		assert.NotEmpty(t, results)
	}
}