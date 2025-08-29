// Package stepfunctions parallel execution - Advanced orchestration for concurrent Step Functions executions
// Following strict Terratest patterns with Function/FunctionE variants

package stepfunctions

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ParallelExecutions executes multiple Step Functions concurrently
// This is a Terratest-style function that panics on error
func ParallelExecutions(ctx *TestContext, executions []ParallelExecution, config *ParallelExecutionConfig) []*ExecutionResult {
	results, err := ParallelExecutionsE(ctx, executions, config)
	if err != nil {
		ctx.T.FailNow()
	}
	return results
}

// ParallelExecutionsE executes multiple Step Functions concurrently with error return
func ParallelExecutionsE(ctx *TestContext, executions []ParallelExecution, config *ParallelExecutionConfig) ([]*ExecutionResult, error) {
	start := time.Now()
	
	// Validate inputs
	if len(executions) == 0 {
		return []*ExecutionResult{}, nil
	}
	
	if config == nil {
		config = defaultParallelExecutionConfig()
	} else {
		config = mergeParallelExecutionConfig(config)
	}
	
	logOperation("ParallelExecutions", map[string]interface{}{
		"execution_count":     len(executions),
		"max_concurrency":     config.MaxConcurrency,
		"wait_for_completion": config.WaitForCompletion,
		"fail_fast":          config.FailFast,
		"timeout":            config.Timeout.String(),
	})
	
	// Execute in parallel with concurrency control
	results, err := executeParallelWithConcurrencyControl(ctx, executions, config)
	
	logResult("ParallelExecutions", err == nil, time.Since(start), err)
	return results, err
}

// executeParallelWithConcurrencyControl manages parallel execution with concurrency limits
func executeParallelWithConcurrencyControl(ctx *TestContext, executions []ParallelExecution, config *ParallelExecutionConfig) ([]*ExecutionResult, error) {
	// Create channels for coordination
	semaphore := make(chan struct{}, config.MaxConcurrency)
	results := make([]*ExecutionResult, len(executions))
	errorChan := make(chan error, len(executions))
	
	var wg sync.WaitGroup
	
	// Create context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	
	// Launch goroutines for each execution
	for i, execution := range executions {
		wg.Add(1)
		go func(index int, exec ParallelExecution) {
			defer wg.Done()
			
			// Acquire semaphore slot
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-execCtx.Done():
				errorChan <- fmt.Errorf("timeout waiting for concurrency slot")
				return
			}
			
			// Execute the step function
			result, err := executeSingleStepFunction(ctx, exec, config, execCtx)
			if err != nil {
				if config.FailFast {
					errorChan <- fmt.Errorf("parallel execution failed: %w", err)
					return
				}
				// Create error result but don't fail fast
				result = &ExecutionResult{
					ExecutionArn:    fmt.Sprintf("failed-execution-%d", index),
					StateMachineArn: exec.StateMachineArn,
					Name:           exec.ExecutionName,
					Status:         ExecutionStatusFailed,
					StartDate:      time.Now(),
					StopDate:       time.Now(),
					Error:          err.Error(),
				}
			}
			
			results[index] = result
		}(i, execution)
	}
	
	// Wait for completion or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// All executions completed
		break
	case <-execCtx.Done():
		return results, fmt.Errorf("parallel executions timeout after %v", config.Timeout)
	case err := <-errorChan:
		if config.FailFast {
			cancel() // Cancel remaining operations
			return results, err
		}
	}
	
	// Filter out nil results
	validResults := make([]*ExecutionResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			validResults = append(validResults, result)
		}
	}
	
	return validResults, nil
}

// executeSingleStepFunction executes a single step function within the parallel context
func executeSingleStepFunction(ctx *TestContext, execution ParallelExecution, config *ParallelExecutionConfig, execCtx context.Context) (*ExecutionResult, error) {
	// Validate individual execution
	if err := validateParallelExecution(execution); err != nil {
		return nil, err
	}
	
	// Start the execution
	result, err := StartExecutionE(ctx, execution.StateMachineArn, execution.ExecutionName, execution.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to start execution %s: %w", execution.ExecutionName, err)
	}
	
	// If not waiting for completion, return immediately
	if !config.WaitForCompletion {
		return result, nil
	}
	
	// Wait for completion with timeout
	waitConfig := &PollConfig{
		Timeout:            config.Timeout,
		Interval:           DefaultPollingInterval,
		MaxAttempts:        int(config.Timeout / DefaultPollingInterval),
		ExponentialBackoff: false,
		BackoffMultiplier:  1.0,
	}
	
	finalResult, err := pollUntilCompleteWithContext(ctx, result.ExecutionArn, waitConfig, execCtx)
	if err != nil {
		return nil, fmt.Errorf("execution %s failed or timed out: %w", execution.ExecutionName, err)
	}
	
	return finalResult, nil
}

// pollUntilCompleteWithContext polls with context cancellation support
func pollUntilCompleteWithContext(ctx *TestContext, executionArn string, config *PollConfig, execCtx context.Context) (*ExecutionResult, error) {
	startTime := time.Now()
	
	for attempt := 1; ; attempt++ {
		select {
		case <-execCtx.Done():
			return nil, fmt.Errorf("execution polling cancelled or timed out")
		default:
		}
		
		// Get current execution status
		result, err := DescribeExecutionE(ctx, executionArn)
		if err != nil {
			return nil, fmt.Errorf("failed to describe execution (attempt %d): %w", attempt, err)
		}
		
		// Check if execution is complete
		if isExecutionComplete(result.Status) {
			return result, nil
		}
		
		elapsed := time.Since(startTime)
		
		// Check if we should continue polling
		if !shouldContinuePolling(attempt, elapsed, config) {
			return nil, fmt.Errorf("execution did not complete after %d attempts in %v", attempt-1, elapsed)
		}
		
		// Calculate next interval and sleep with context awareness
		interval := calculateNextPollInterval(attempt, config)
		timer := time.NewTimer(interval)
		
		select {
		case <-timer.C:
			// Continue polling
		case <-execCtx.Done():
			timer.Stop()
			return nil, fmt.Errorf("execution polling cancelled or timed out")
		}
	}
}

// Validation functions

// validateParallelExecution validates a single parallel execution request
func validateParallelExecution(execution ParallelExecution) error {
	if execution.StateMachineArn == "" {
		return fmt.Errorf("state machine ARN is required")
	}
	
	if err := validateStateMachineArn(execution.StateMachineArn); err != nil {
		return fmt.Errorf("invalid state machine ARN: %w", err)
	}
	
	if execution.ExecutionName == "" {
		return fmt.Errorf("execution name is required")
	}
	
	if err := validateExecutionName(execution.ExecutionName); err != nil {
		return fmt.Errorf("invalid execution name: %w", err)
	}
	
	// Input validation (if provided)
	if execution.Input != nil && !execution.Input.isEmpty() {
		if _, err := execution.Input.ToJSON(); err != nil {
			return fmt.Errorf("invalid input JSON: %w", err)
		}
	}
	
	return nil
}

// Configuration functions

// defaultParallelExecutionConfig provides sensible defaults
func defaultParallelExecutionConfig() *ParallelExecutionConfig {
	return &ParallelExecutionConfig{
		MaxConcurrency:    3,
		Timeout:          DefaultTimeout,
		WaitForCompletion: true,
		FailFast:         false,
	}
}

// mergeParallelExecutionConfig merges user config with defaults
func mergeParallelExecutionConfig(userConfig *ParallelExecutionConfig) *ParallelExecutionConfig {
	defaults := defaultParallelExecutionConfig()
	
	if userConfig == nil {
		return defaults
	}
	
	merged := *userConfig
	
	if merged.MaxConcurrency == 0 {
		merged.MaxConcurrency = defaults.MaxConcurrency
	}
	if merged.Timeout == 0 {
		merged.Timeout = defaults.Timeout
	}
	
	return &merged
}