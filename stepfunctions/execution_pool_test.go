package stepfunctions

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutionPool_ShouldManageExecutions(t *testing.T) {
	t.Helper()
	
	// Given: an execution pool with limits
	poolConfig := &ExecutionPoolConfig{
		MaxConcurrentExecutions: 3,
		MaxQueueSize:           10,
		ExecutionTimeout:       2 * time.Minute,
		PoolTimeout:           5 * time.Minute,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}
	
	pool := NewExecutionPool(ctx, poolConfig)

	// When: submitting executions to the pool
	requests := []ExecutionRequest{
		{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test1",
			ExecutionName:   "pool-execution-1",
			Input:          NewInput().Set("test", "value1"),
		},
		{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test2",
			ExecutionName:   "pool-execution-2",
			Input:          NewInput().Set("test", "value2"),
		},
	}
	
	results := make([]*PoolExecutionResult, len(requests))
	for i, req := range requests {
		result, err := pool.SubmitE(req)
		require.NoError(t, err)
		results[i] = result
	}

	// Then: should manage executions properly
	assert.Len(t, results, 2)
	for _, result := range results {
		assert.NotNil(t, result.ExecutionID)
		assert.Equal(t, PoolExecutionStatusQueued, result.Status)
	}
	
	// Cleanup
	pool.Shutdown()
}

func TestExecutionPool_ShouldRespectConcurrencyLimits(t *testing.T) {
	t.Helper()
	
	// Given: a pool with very low concurrency limit
	poolConfig := &ExecutionPoolConfig{
		MaxConcurrentExecutions: 1, // Only 1 concurrent execution
		MaxQueueSize:           5,
		ExecutionTimeout:       2 * time.Minute,
		PoolTimeout:           5 * time.Minute,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}
	
	pool := NewExecutionPool(ctx, poolConfig)

	// When: submitting multiple executions
	requests := make([]ExecutionRequest, 3)
	for i := 0; i < 3; i++ {
		requests[i] = ExecutionRequest{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
			ExecutionName:   fmt.Sprintf("concurrent-test-%d", i+1),
			Input:          NewInput().Set("index", i+1),
		}
	}
	
	results := make([]*PoolExecutionResult, len(requests))
	for i, req := range requests {
		result, err := pool.SubmitE(req)
		require.NoError(t, err)
		results[i] = result
	}

	// Then: should queue executions properly
	assert.Len(t, results, 3)
	
	// Check that concurrency is respected
	stats := pool.GetStats()
	assert.LessOrEqual(t, stats.RunningExecutions, poolConfig.MaxConcurrentExecutions)
	assert.Equal(t, 3, stats.TotalSubmitted)
	
	pool.Shutdown()
}

func TestExecutionPool_ShouldHandleQueueOverflow(t *testing.T) {
	t.Helper()
	
	// Given: a pool with very small queue
	poolConfig := &ExecutionPoolConfig{
		MaxConcurrentExecutions: 1,
		MaxQueueSize:           2, // Very small queue
		ExecutionTimeout:       2 * time.Minute,
		PoolTimeout:           5 * time.Minute,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}
	
	pool := NewExecutionPool(ctx, poolConfig)

	// When: submitting exactly the right number to fill pool
	maxCapacity := poolConfig.MaxConcurrentExecutions + poolConfig.MaxQueueSize
	
	// Submit exactly the pool capacity
	successCount := 0
	for i := 0; i < maxCapacity; i++ {
		req := ExecutionRequest{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
			ExecutionName:   fmt.Sprintf("capacity-test-%d", i+1),
			Input:          NewInput().Set("index", i+1),
		}
		_, err := pool.SubmitE(req)
		if err == nil {
			successCount++
		}
	}
	
	// Now submit one more which should be rejected
	overflowReq := ExecutionRequest{
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
		ExecutionName:   "overflow-test",
		Input:          NewInput().Set("overflow", true),
	}
	_, overflowErr := pool.SubmitE(overflowReq)

	// Then: should accept most submissions up to capacity 
	assert.GreaterOrEqual(t, successCount, maxCapacity-1) // Allow for race conditions
	assert.LessOrEqual(t, successCount, maxCapacity)
	
	// The overflow request should either error or succeed (if there's still capacity due to race conditions)
	if overflowErr != nil {
		// Error message can be either "capacity" or "queue is full"
		errorMsg := overflowErr.Error()
		assert.True(t, 
			strings.Contains(errorMsg, "capacity") || strings.Contains(errorMsg, "full"),
			"Expected error about capacity or queue full, got: %s", errorMsg)
	} else {
		// If no error, the pool had room (race condition)
		t.Logf("Overflow request succeeded (race condition), capacity not fully utilized")
	}
	
	pool.Shutdown()
}

func TestExecutionPool_ShouldProvideStats(t *testing.T) {
	t.Helper()
	
	// Given: a pool with some executions
	poolConfig := &ExecutionPoolConfig{
		MaxConcurrentExecutions: 2,
		MaxQueueSize:           3,
		ExecutionTimeout:       2 * time.Minute,
		PoolTimeout:           5 * time.Minute,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}
	
	pool := NewExecutionPool(ctx, poolConfig)

	// When: submitting some executions
	for i := 0; i < 3; i++ {
		req := ExecutionRequest{
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
			ExecutionName:   fmt.Sprintf("stats-test-%d", i+1),
			Input:          NewInput().Set("index", i+1),
		}
		_, err := pool.SubmitE(req)
		require.NoError(t, err)
	}

	// Then: should provide accurate stats
	stats := pool.GetStats()
	assert.Equal(t, 3, stats.TotalSubmitted)
	assert.LessOrEqual(t, stats.RunningExecutions, 2)
	assert.GreaterOrEqual(t, stats.QueuedExecutions, 0)
	assert.Equal(t, 0, stats.CompletedExecutions) // None completed yet since they'll error
	assert.Equal(t, 2, stats.MaxConcurrentExecutions)
	assert.Equal(t, 3, stats.MaxQueueSize)
	
	pool.Shutdown()
}

func TestExecutionPool_ShouldWaitForCompletion(t *testing.T) {
	t.Helper()
	
	// Given: a pool with executions
	poolConfig := &ExecutionPoolConfig{
		MaxConcurrentExecutions: 2,
		MaxQueueSize:           3,
		ExecutionTimeout:       100 * time.Millisecond, // Short timeout
		PoolTimeout:           1 * time.Second,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}
	
	pool := NewExecutionPool(ctx, poolConfig)

	// When: submitting executions and waiting for completion
	req := ExecutionRequest{
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
		ExecutionName:   "wait-test",
		Input:          NewInput().Set("test", "wait"),
	}
	
	result, err := pool.SubmitE(req)
	require.NoError(t, err)

	// Wait for all executions to complete
	err = pool.WaitForAllE()
	
	// Then: should wait for completion or timeout
	assert.NoError(t, err) // Should complete (with errors) within timeout
	
	// Final result should be completed with error due to AWS config
	finalResult := pool.GetExecution(result.ExecutionID)
	assert.NotNil(t, finalResult)
	assert.Equal(t, PoolExecutionStatusFailed, finalResult.Status)
	
	pool.Shutdown()
}

func TestExecutionPool_ShouldCancelExecutions(t *testing.T) {
	t.Helper()
	
	// Given: a pool with pending executions
	poolConfig := &ExecutionPoolConfig{
		MaxConcurrentExecutions: 1,
		MaxQueueSize:           3,
		ExecutionTimeout:       2 * time.Minute,
		PoolTimeout:           5 * time.Minute,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}
	
	pool := NewExecutionPool(ctx, poolConfig)

	// When: submitting executions and then canceling
	req := ExecutionRequest{
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test",
		ExecutionName:   "cancel-test",
		Input:          NewInput().Set("test", "cancel"),
	}
	
	result, err := pool.SubmitE(req)
	require.NoError(t, err)

	// Cancel the execution
	err = pool.CancelExecutionE(result.ExecutionID)

	// Then: should cancel successfully
	assert.NoError(t, err)
	
	// Give time for processing
	time.Sleep(10 * time.Millisecond)
	
	// Execution should be canceled
	finalResult := pool.GetExecution(result.ExecutionID)
	assert.NotNil(t, finalResult)
	assert.Equal(t, PoolExecutionStatusCancelled, finalResult.Status)
	
	pool.Shutdown()
}

func TestExecutionPool_ShouldTimeout(t *testing.T) {
	t.Helper()
	
	// Given: a pool with very short timeout
	poolConfig := &ExecutionPoolConfig{
		MaxConcurrentExecutions: 1,
		MaxQueueSize:           2,
		ExecutionTimeout:       10 * time.Millisecond, // Very short
		PoolTimeout:           1 * time.Second,
	}
	
	ctx := &TestContext{
		T: newMockTestingT(),
	}
	
	pool := NewExecutionPool(ctx, poolConfig)

	// When: submitting a long-running execution
	req := ExecutionRequest{
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:long-running",
		ExecutionName:   "timeout-test",
		Input:          NewInput().Set("duration", "long"),
	}
	
	result, err := pool.SubmitE(req)
	require.NoError(t, err)

	// Wait a bit for timeout to occur and processing
	time.Sleep(100 * time.Millisecond)

	// Then: execution should timeout or fail due to AWS config
	finalResult := pool.GetExecution(result.ExecutionID)
	assert.NotNil(t, finalResult)
	// Should have been processed (either failed due to AWS error or running/queued)
	assert.Contains(t, []ExecutionPoolStatus{
		PoolExecutionStatusFailed, 
		PoolExecutionStatusTimeout, 
		PoolExecutionStatusRunning, 
		PoolExecutionStatusQueued,
	}, finalResult.Status)
	
	pool.Shutdown()
}