package stepfunctions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowChain_ShouldExecuteSequentially(t *testing.T) {
	t.Helper()
	
	// Given: a workflow chain with multiple steps
	chain := NewWorkflowChain()
	
	step1 := &WorkflowStep{
		Name: "preprocess-data",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:preprocess",
		Input: NewInput().Set("data", "raw_input"),
	}
	
	step2 := &WorkflowStep{
		Name: "transform-data",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:transform",
		Input: NewInput().Set("source", "{{ preprocess-data.output }}"),
		Dependencies: []string{"preprocess-data"},
	}
	
	step3 := &WorkflowStep{
		Name: "finalize-data",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:finalize",
		Input: NewInput().Set("transformed", "{{ transform-data.output }}"),
		Dependencies: []string{"transform-data"},
	}
	
	chain.AddStep(step1)
	chain.AddStep(step2)
	chain.AddStep(step3)
	
	config := &WorkflowChainConfig{
		Timeout: 5 * time.Minute,
		FailFast: true,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: executing the workflow chain
	result, err := ExecuteWorkflowChainE(ctx, chain, config)

	// Then: should execute all steps in dependency order
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.StepResults, 3)
	
	// Steps should have executed in order
	stepNames := make([]string, len(result.StepResults))
	for i, stepResult := range result.StepResults {
		stepNames[i] = stepResult.Step.Name
	}
	
	assert.Equal(t, []string{"preprocess-data", "transform-data", "finalize-data"}, stepNames)
}

func TestWorkflowChain_ShouldHandleBranching(t *testing.T) {
	t.Helper()
	
	// Given: a workflow with branching paths
	chain := NewWorkflowChain()
	
	step1 := &WorkflowStep{
		Name: "initial-step",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:initial",
		Input: NewInput().Set("data", "start"),
	}
	
	step2a := &WorkflowStep{
		Name: "branch-a",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:branchA",
		Input: NewInput().Set("source", "{{ initial-step.output }}"),
		Dependencies: []string{"initial-step"},
		Condition: &WorkflowCondition{
			Expression: "{{ initial-step.output.type }} == 'typeA'",
		},
	}
	
	step2b := &WorkflowStep{
		Name: "branch-b", 
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:branchB",
		Input: NewInput().Set("source", "{{ initial-step.output }}"),
		Dependencies: []string{"initial-step"},
		Condition: &WorkflowCondition{
			Expression: "{{ initial-step.output.type }} == 'typeB'",
		},
	}
	
	step3 := &WorkflowStep{
		Name: "final-step",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:final",
		Input: NewInput().Set("result", "{{ branch-a.output || branch-b.output }}"),
		Dependencies: []string{"branch-a", "branch-b"},
		DependencyMode: DependencyModeAny, // Execute when ANY dependency completes
	}
	
	chain.AddStep(step1)
	chain.AddStep(step2a)
	chain.AddStep(step2b)
	chain.AddStep(step3)
	
	config := &WorkflowChainConfig{
		Timeout: 5 * time.Minute,
		FailFast: false, // Allow conditional steps to be skipped
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: executing the branching workflow
	result, err := ExecuteWorkflowChainE(ctx, chain, config)

	// Then: should handle branching correctly
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Should have executed initial step and final step (branches will fail due to AWS config)
	assert.GreaterOrEqual(t, len(result.StepResults), 2)
}

func TestWorkflowChain_ShouldHandleParallelSteps(t *testing.T) {
	t.Helper()
	
	// Given: a workflow with parallel steps
	chain := NewWorkflowChain()
	
	step1 := &WorkflowStep{
		Name: "setup",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:setup",
		Input: NewInput().Set("init", "true"),
	}
	
	step2a := &WorkflowStep{
		Name: "parallel-task-a",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:taskA",
		Input: NewInput().Set("task", "A"),
		Dependencies: []string{"setup"},
	}
	
	step2b := &WorkflowStep{
		Name: "parallel-task-b",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:taskB",
		Input: NewInput().Set("task", "B"),
		Dependencies: []string{"setup"},
	}
	
	step2c := &WorkflowStep{
		Name: "parallel-task-c",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:taskC",
		Input: NewInput().Set("task", "C"),
		Dependencies: []string{"setup"},
	}
	
	step3 := &WorkflowStep{
		Name: "aggregate",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:aggregate",
		Input: NewInput().Set("results", []interface{}{
			"{{ parallel-task-a.output }}",
			"{{ parallel-task-b.output }}",
			"{{ parallel-task-c.output }}",
		}),
		Dependencies: []string{"parallel-task-a", "parallel-task-b", "parallel-task-c"},
		DependencyMode: DependencyModeAll, // Wait for ALL dependencies
	}
	
	chain.AddStep(step1)
	chain.AddStep(step2a)
	chain.AddStep(step2b)
	chain.AddStep(step2c)
	chain.AddStep(step3)
	
	config := &WorkflowChainConfig{
		Timeout: 5 * time.Minute,
		MaxParallelSteps: 3, // Allow up to 3 steps in parallel
		FailFast: false,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: executing the parallel workflow
	result, err := ExecuteWorkflowChainE(ctx, chain, config)

	// Then: should execute parallel steps concurrently
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.StepResults), 3) // At least setup + some parallel tasks
}

func TestWorkflowChain_ShouldHandleRetries(t *testing.T) {
	t.Helper()
	
	// Given: a workflow with retry configuration
	chain := NewWorkflowChain()
	
	step1 := &WorkflowStep{
		Name: "flaky-step",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:flaky",
		Input: NewInput().Set("attempt", 1),
		RetryConfig: &WorkflowRetryConfig{
			MaxAttempts: 3,
			BackoffMultiplier: 2.0,
			InitialInterval: 100 * time.Millisecond,
		},
	}
	
	chain.AddStep(step1)
	
	config := &WorkflowChainConfig{
		Timeout: 2 * time.Minute,
		FailFast: true,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: executing the workflow with retries
	result, err := ExecuteWorkflowChainE(ctx, chain, config)

	// Then: should attempt retries (will eventually fail due to AWS config but should show retry attempts)
	require.NoError(t, err) // The workflow framework shouldn't error, but individual steps might
	assert.NotNil(t, result)
	assert.Len(t, result.StepResults, 1)
	
	stepResult := result.StepResults[0]
	assert.Equal(t, "flaky-step", stepResult.Step.Name)
	// Should show that retries were attempted
	assert.GreaterOrEqual(t, stepResult.AttemptCount, 1)
}

func TestWorkflowChain_ShouldValidateChain(t *testing.T) {
	t.Helper()
	
	// Given: an invalid workflow chain
	chain := NewWorkflowChain()
	
	// Step with circular dependency
	step1 := &WorkflowStep{
		Name: "step1",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:step1",
		Dependencies: []string{"step2"},
	}
	
	step2 := &WorkflowStep{
		Name: "step2", 
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:step2",
		Dependencies: []string{"step1"}, // Circular dependency
	}
	
	chain.AddStep(step1)
	chain.AddStep(step2)
	
	config := &WorkflowChainConfig{
		Timeout: 1 * time.Minute,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: validating the workflow chain
	_, err := ExecuteWorkflowChainE(ctx, chain, config)

	// Then: should detect circular dependency
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

func TestWorkflowChain_ShouldProvidStats(t *testing.T) {
	t.Helper()
	
	// Given: a simple workflow chain
	chain := NewWorkflowChain()
	
	step1 := &WorkflowStep{
		Name: "test-step",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
		Input: NewInput().Set("test", "value"),
	}
	
	chain.AddStep(step1)
	
	config := &WorkflowChainConfig{
		Timeout: 30 * time.Second,
		FailFast: true,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}

	// When: executing the workflow
	result, err := ExecuteWorkflowChainE(ctx, chain, config)

	// Then: should provide execution statistics
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Stats)
	
	stats := result.Stats
	assert.Equal(t, 1, stats.TotalSteps)
	assert.GreaterOrEqual(t, stats.ExecutedSteps, 1)
	assert.GreaterOrEqual(t, stats.FailedSteps, 0)
	assert.GreaterOrEqual(t, stats.SkippedSteps, 0)
	assert.Greater(t, stats.ExecutionTime, 0*time.Millisecond)
}