// Package stepfunctions batch operations - Batch processing for Step Functions operations
// Following strict Terratest patterns with Function/FunctionE variants

package stepfunctions

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Batch State Machine Operations

// BatchCreateStateMachines creates multiple Step Functions state machines concurrently
// This is a Terratest-style function that panics on error
func BatchCreateStateMachines(ctx *TestContext, definitions []*StateMachineDefinition, config *BatchOperationConfig) []*BatchStateMachineResult {
	results, err := BatchCreateStateMachinesE(ctx, definitions, config)
	if err != nil {
		ctx.T.FailNow()
	}
	return results
}

// BatchCreateStateMachinesE creates multiple Step Functions state machines concurrently with error return
func BatchCreateStateMachinesE(ctx *TestContext, definitions []*StateMachineDefinition, config *BatchOperationConfig) ([]*BatchStateMachineResult, error) {
	start := time.Now()
	
	// Validate inputs
	if len(definitions) == 0 {
		return []*BatchStateMachineResult{}, nil
	}
	
	if config == nil {
		config = defaultBatchOperationConfig()
	} else {
		config = mergeBatchOperationConfig(config)
	}
	
	logOperation("BatchCreateStateMachines", map[string]interface{}{
		"definition_count":  len(definitions),
		"max_concurrency":   config.MaxConcurrency,
		"fail_fast":        config.FailFast,
		"timeout":          config.Timeout.String(),
	})
	
	// Execute batch operation with concurrency control
	results, err := executeBatchStateMachineCreation(ctx, definitions, config)
	
	logResult("BatchCreateStateMachines", err == nil, time.Since(start), err)
	return results, err
}

// BatchDeleteStateMachines deletes multiple Step Functions state machines concurrently
func BatchDeleteStateMachines(ctx *TestContext, stateMachineArns []string, config *BatchOperationConfig) []*BatchStateMachineResult {
	results, err := BatchDeleteStateMachinesE(ctx, stateMachineArns, config)
	if err != nil {
		ctx.T.FailNow()
	}
	return results
}

// BatchDeleteStateMachinesE deletes multiple Step Functions state machines concurrently with error return
func BatchDeleteStateMachinesE(ctx *TestContext, stateMachineArns []string, config *BatchOperationConfig) ([]*BatchStateMachineResult, error) {
	start := time.Now()
	
	// Validate inputs
	if len(stateMachineArns) == 0 {
		return []*BatchStateMachineResult{}, nil
	}
	
	if config == nil {
		config = defaultBatchOperationConfig()
	} else {
		config = mergeBatchOperationConfig(config)
	}
	
	logOperation("BatchDeleteStateMachines", map[string]interface{}{
		"state_machine_count": len(stateMachineArns),
		"max_concurrency":     config.MaxConcurrency,
		"fail_fast":          config.FailFast,
		"timeout":            config.Timeout.String(),
	})
	
	// Execute batch operation with concurrency control
	results, err := executeBatchStateMachineDeletion(ctx, stateMachineArns, config)
	
	logResult("BatchDeleteStateMachines", err == nil, time.Since(start), err)
	return results, err
}

// BatchTagStateMachines adds tags to multiple state machines concurrently
func BatchTagStateMachines(ctx *TestContext, operations []BatchTagOperation, config *BatchOperationConfig) []*BatchTagResult {
	results, err := BatchTagStateMachinesE(ctx, operations, config)
	if err != nil {
		ctx.T.FailNow()
	}
	return results
}

// BatchTagStateMachinesE adds tags to multiple state machines concurrently with error return
func BatchTagStateMachinesE(ctx *TestContext, operations []BatchTagOperation, config *BatchOperationConfig) ([]*BatchTagResult, error) {
	start := time.Now()
	
	// Validate inputs
	if len(operations) == 0 {
		return []*BatchTagResult{}, nil
	}
	
	if config == nil {
		config = defaultBatchOperationConfig()
	} else {
		config = mergeBatchOperationConfig(config)
	}
	
	logOperation("BatchTagStateMachines", map[string]interface{}{
		"operation_count":   len(operations),
		"max_concurrency":   config.MaxConcurrency,
		"fail_fast":        config.FailFast,
		"timeout":          config.Timeout.String(),
	})
	
	// Execute batch operation with concurrency control
	results, err := executeBatchStateMachineTagging(ctx, operations, config)
	
	logResult("BatchTagStateMachines", err == nil, time.Since(start), err)
	return results, err
}

// Batch Execution Operations

// BatchExecutions starts multiple Step Functions executions concurrently
func BatchExecutions(ctx *TestContext, executions []BatchExecutionRequest, config *BatchOperationConfig) []*BatchExecutionResult {
	results, err := BatchExecutionsE(ctx, executions, config)
	if err != nil {
		ctx.T.FailNow()
	}
	return results
}

// BatchExecutionsE starts multiple Step Functions executions concurrently with error return
func BatchExecutionsE(ctx *TestContext, executions []BatchExecutionRequest, config *BatchOperationConfig) ([]*BatchExecutionResult, error) {
	start := time.Now()
	
	// Validate inputs
	if len(executions) == 0 {
		return []*BatchExecutionResult{}, nil
	}
	
	if config == nil {
		config = defaultBatchOperationConfig()
	} else {
		config = mergeBatchOperationConfig(config)
	}
	
	logOperation("BatchExecutions", map[string]interface{}{
		"execution_count":   len(executions),
		"max_concurrency":   config.MaxConcurrency,
		"fail_fast":        config.FailFast,
		"timeout":          config.Timeout.String(),
	})
	
	// Execute batch operation with concurrency control
	results, err := executeBatchExecutions(ctx, executions, config)
	
	logResult("BatchExecutions", err == nil, time.Since(start), err)
	return results, err
}

// BatchStopExecutions stops multiple Step Functions executions concurrently
func BatchStopExecutions(ctx *TestContext, executionArns []string, errorMsg, cause string, config *BatchOperationConfig) []*BatchExecutionResult {
	results, err := BatchStopExecutionsE(ctx, executionArns, errorMsg, cause, config)
	if err != nil {
		ctx.T.FailNow()
	}
	return results
}

// BatchStopExecutionsE stops multiple Step Functions executions concurrently with error return
func BatchStopExecutionsE(ctx *TestContext, executionArns []string, errorMsg, cause string, config *BatchOperationConfig) ([]*BatchExecutionResult, error) {
	start := time.Now()
	
	// Validate inputs
	if len(executionArns) == 0 {
		return []*BatchExecutionResult{}, nil
	}
	
	if config == nil {
		config = defaultBatchOperationConfig()
	} else {
		config = mergeBatchOperationConfig(config)
	}
	
	logOperation("BatchStopExecutions", map[string]interface{}{
		"execution_count":   len(executionArns),
		"max_concurrency":   config.MaxConcurrency,
		"fail_fast":        config.FailFast,
		"timeout":          config.Timeout.String(),
		"error":            errorMsg,
		"cause":            cause,
	})
	
	// Execute batch operation with concurrency control
	results, err := executeBatchExecutionStop(ctx, executionArns, errorMsg, cause, config)
	
	logResult("BatchStopExecutions", err == nil, time.Since(start), err)
	return results, err
}

// Internal Implementation Functions

// executeBatchStateMachineCreation handles concurrent state machine creation
func executeBatchStateMachineCreation(ctx *TestContext, definitions []*StateMachineDefinition, config *BatchOperationConfig) ([]*BatchStateMachineResult, error) {
	// Create channels for coordination
	semaphore := make(chan struct{}, config.MaxConcurrency)
	results := make([]*BatchStateMachineResult, len(definitions))
	errorChan := make(chan error, len(definitions))
	
	var wg sync.WaitGroup
	
	// Create context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	
	// Launch goroutines for each creation
	for i, definition := range definitions {
		wg.Add(1)
		go func(index int, def *StateMachineDefinition) {
			defer wg.Done()
			
			// Acquire semaphore slot
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-execCtx.Done():
				errorChan <- fmt.Errorf("timeout waiting for concurrency slot")
				return
			}
			
			// Execute the state machine creation
			result := &BatchStateMachineResult{
				StateMachineArn: fmt.Sprintf("creating-%s", def.Name),
			}
			
			stateMachine, err := CreateStateMachineE(ctx, def, nil)
			if err != nil {
				if config.FailFast {
					errorChan <- fmt.Errorf("batch operation failed: %w", err)
					return
				}
				result.Error = err
			} else {
				result.StateMachine = stateMachine
				result.StateMachineArn = stateMachine.StateMachineArn
			}
			
			results[index] = result
		}(i, definition)
	}
	
	// Wait for completion or timeout
	return waitForBatchCompletion(results, &wg, execCtx, errorChan, config)
}

// executeBatchStateMachineDeletion handles concurrent state machine deletion
func executeBatchStateMachineDeletion(ctx *TestContext, stateMachineArns []string, config *BatchOperationConfig) ([]*BatchStateMachineResult, error) {
	// Create channels for coordination
	semaphore := make(chan struct{}, config.MaxConcurrency)
	results := make([]*BatchStateMachineResult, len(stateMachineArns))
	errorChan := make(chan error, len(stateMachineArns))
	
	var wg sync.WaitGroup
	
	// Create context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	
	// Launch goroutines for each deletion
	for i, arn := range stateMachineArns {
		wg.Add(1)
		go func(index int, stateMachineArn string) {
			defer wg.Done()
			
			// Acquire semaphore slot
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-execCtx.Done():
				errorChan <- fmt.Errorf("timeout waiting for concurrency slot")
				return
			}
			
			// Execute the state machine deletion
			result := &BatchStateMachineResult{
				StateMachineArn: stateMachineArn,
			}
			
			err := DeleteStateMachineE(ctx, stateMachineArn)
			if err != nil {
				if config.FailFast {
					errorChan <- fmt.Errorf("batch operation failed: %w", err)
					return
				}
				result.Error = err
			}
			
			results[index] = result
		}(i, arn)
	}
	
	// Wait for completion or timeout
	return waitForBatchCompletion(results, &wg, execCtx, errorChan, config)
}

// executeBatchStateMachineTagging handles concurrent state machine tagging
func executeBatchStateMachineTagging(ctx *TestContext, operations []BatchTagOperation, config *BatchOperationConfig) ([]*BatchTagResult, error) {
	// Create channels for coordination
	semaphore := make(chan struct{}, config.MaxConcurrency)
	results := make([]*BatchTagResult, len(operations))
	errorChan := make(chan error, len(operations))
	
	var wg sync.WaitGroup
	
	// Create context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	
	// Launch goroutines for each tagging operation
	for i, operation := range operations {
		wg.Add(1)
		go func(index int, op BatchTagOperation) {
			defer wg.Done()
			
			// Acquire semaphore slot
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-execCtx.Done():
				errorChan <- fmt.Errorf("timeout waiting for concurrency slot")
				return
			}
			
			// Execute the tagging operation
			result := &BatchTagResult{
				StateMachineArn: op.StateMachineArn,
				Tags:           op.Tags,
			}
			
			err := TagStateMachineE(ctx, op.StateMachineArn, op.Tags)
			if err != nil {
				if config.FailFast {
					errorChan <- fmt.Errorf("batch operation failed: %w", err)
					return
				}
				result.Error = err
			}
			
			results[index] = result
		}(i, operation)
	}
	
	// Wait for completion or timeout
	return waitForBatchTagCompletion(results, &wg, execCtx, errorChan, config)
}

// executeBatchExecutions handles concurrent execution starting
func executeBatchExecutions(ctx *TestContext, executions []BatchExecutionRequest, config *BatchOperationConfig) ([]*BatchExecutionResult, error) {
	// Create channels for coordination
	semaphore := make(chan struct{}, config.MaxConcurrency)
	results := make([]*BatchExecutionResult, len(executions))
	errorChan := make(chan error, len(executions))
	
	var wg sync.WaitGroup
	
	// Create context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	
	// Launch goroutines for each execution
	for i, execution := range executions {
		wg.Add(1)
		go func(index int, exec BatchExecutionRequest) {
			defer wg.Done()
			
			// Acquire semaphore slot
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-execCtx.Done():
				errorChan <- fmt.Errorf("timeout waiting for concurrency slot")
				return
			}
			
			// Execute the execution start
			result := &BatchExecutionResult{
				ExecutionArn: fmt.Sprintf("starting-%s", exec.ExecutionName),
			}
			
			execution, err := StartExecutionE(ctx, exec.StateMachineArn, exec.ExecutionName, exec.Input)
			if err != nil {
				if config.FailFast {
					errorChan <- fmt.Errorf("batch operation failed: %w", err)
					return
				}
				result.Error = err
			} else {
				result.Execution = execution
				result.ExecutionArn = execution.ExecutionArn
			}
			
			results[index] = result
		}(i, execution)
	}
	
	// Wait for completion or timeout
	return waitForBatchExecutionCompletion(results, &wg, execCtx, errorChan, config)
}

// executeBatchExecutionStop handles concurrent execution stopping
func executeBatchExecutionStop(ctx *TestContext, executionArns []string, errorMsg, cause string, config *BatchOperationConfig) ([]*BatchExecutionResult, error) {
	// Create channels for coordination
	semaphore := make(chan struct{}, config.MaxConcurrency)
	results := make([]*BatchExecutionResult, len(executionArns))
	errorChan := make(chan error, len(executionArns))
	
	var wg sync.WaitGroup
	
	// Create context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	
	// Launch goroutines for each execution stop
	for i, arn := range executionArns {
		wg.Add(1)
		go func(index int, executionArn string) {
			defer wg.Done()
			
			// Acquire semaphore slot
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-execCtx.Done():
				errorChan <- fmt.Errorf("timeout waiting for concurrency slot")
				return
			}
			
			// Execute the execution stop
			result := &BatchExecutionResult{
				ExecutionArn: executionArn,
			}
			
			execution, err := StopExecutionE(ctx, executionArn, errorMsg, cause)
			if err != nil {
				if config.FailFast {
					errorChan <- fmt.Errorf("batch operation failed: %w", err)
					return
				}
				result.Error = err
			} else {
				result.Execution = execution
			}
			
			results[index] = result
		}(i, arn)
	}
	
	// Wait for completion or timeout
	return waitForBatchExecutionCompletion(results, &wg, execCtx, errorChan, config)
}

// Utility functions for batch operations

// waitForBatchCompletion waits for batch state machine operations to complete
func waitForBatchCompletion(results []*BatchStateMachineResult, wg *sync.WaitGroup, execCtx context.Context, errorChan <-chan error, config *BatchOperationConfig) ([]*BatchStateMachineResult, error) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// All operations completed
		break
	case <-execCtx.Done():
		return filterValidStateMachineResults(results), fmt.Errorf("batch operations timeout after %v", config.Timeout)
	case err := <-errorChan:
		if config.FailFast {
			return filterValidStateMachineResults(results), err
		}
	}
	
	return filterValidStateMachineResults(results), nil
}

// waitForBatchTagCompletion waits for batch tag operations to complete
func waitForBatchTagCompletion(results []*BatchTagResult, wg *sync.WaitGroup, execCtx context.Context, errorChan <-chan error, config *BatchOperationConfig) ([]*BatchTagResult, error) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// All operations completed
		break
	case <-execCtx.Done():
		return filterValidTagResults(results), fmt.Errorf("batch operations timeout after %v", config.Timeout)
	case err := <-errorChan:
		if config.FailFast {
			return filterValidTagResults(results), err
		}
	}
	
	return filterValidTagResults(results), nil
}

// waitForBatchExecutionCompletion waits for batch execution operations to complete
func waitForBatchExecutionCompletion(results []*BatchExecutionResult, wg *sync.WaitGroup, execCtx context.Context, errorChan <-chan error, config *BatchOperationConfig) ([]*BatchExecutionResult, error) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// All operations completed
		break
	case <-execCtx.Done():
		return filterValidExecutionResults(results), fmt.Errorf("batch operations timeout after %v", config.Timeout)
	case err := <-errorChan:
		if config.FailFast {
			return filterValidExecutionResults(results), err
		}
	}
	
	return filterValidExecutionResults(results), nil
}

// Filter functions to remove nil results

func filterValidStateMachineResults(results []*BatchStateMachineResult) []*BatchStateMachineResult {
	validResults := make([]*BatchStateMachineResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			validResults = append(validResults, result)
		}
	}
	return validResults
}

func filterValidTagResults(results []*BatchTagResult) []*BatchTagResult {
	validResults := make([]*BatchTagResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			validResults = append(validResults, result)
		}
	}
	return validResults
}

func filterValidExecutionResults(results []*BatchExecutionResult) []*BatchExecutionResult {
	validResults := make([]*BatchExecutionResult, 0, len(results))
	for _, result := range results {
		if result != nil {
			validResults = append(validResults, result)
		}
	}
	return validResults
}

// Configuration functions

// defaultBatchOperationConfig provides sensible defaults for batch operations
func defaultBatchOperationConfig() *BatchOperationConfig {
	return &BatchOperationConfig{
		MaxConcurrency: 3,
		Timeout:        DefaultTimeout,
		FailFast:       false,
	}
}

// mergeBatchOperationConfig merges user config with defaults
func mergeBatchOperationConfig(userConfig *BatchOperationConfig) *BatchOperationConfig {
	defaults := defaultBatchOperationConfig()
	
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