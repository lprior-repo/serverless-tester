// Package stepfunctions operations - AWS Step Functions lifecycle management and execution operations
// Following strict Terratest patterns with Function/FunctionE variants

package stepfunctions

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

// State Machine Lifecycle Management

// CreateStateMachine creates a new Step Functions state machine
// This is a Terratest-style function that panics on error
func CreateStateMachine(ctx *TestContext, definition *StateMachineDefinition, options *StateMachineOptions) *StateMachineInfo {
	info, err := CreateStateMachineE(ctx, definition, options)
	if err != nil {
		ctx.T.FailNow()
	}
	return info
}

// CreateStateMachineE creates a new Step Functions state machine with error return
func CreateStateMachineE(ctx *TestContext, definition *StateMachineDefinition, options *StateMachineOptions) (*StateMachineInfo, error) {
	start := time.Now()
	logOperation("CreateStateMachine", map[string]interface{}{
		"name": definition.Name,
		"type": definition.Type,
	})
	
	// Validate input parameters
	if err := validateStateMachineCreation(definition, options); err != nil {
		logResult("CreateStateMachine", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	// Merge options with definition
	mergedOptions := mergeStateMachineOptions(options)
	
	input := &sfn.CreateStateMachineInput{
		Name:       aws.String(definition.Name),
		Definition: aws.String(definition.Definition),
		RoleArn:    aws.String(definition.RoleArn),
		Type:       definition.Type,
		Tags:       convertTagsToSFN(definition.Tags),
	}
	
	// Add optional configurations
	if mergedOptions.LoggingConfiguration != nil {
		input.LoggingConfiguration = mergedOptions.LoggingConfiguration
	}
	if mergedOptions.TracingConfiguration != nil {
		input.TracingConfiguration = mergedOptions.TracingConfiguration
	}
	
	result, err := client.CreateStateMachine(context.TODO(), input)
	if err != nil {
		logResult("CreateStateMachine", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to create state machine: %w", err)
	}
	
	// Convert result to our standard format
	info := &StateMachineInfo{
		StateMachineArn:      aws.ToString(result.StateMachineArn),
		Name:                 definition.Name,
		Status:               types.StateMachineStatusActive, // New state machines are active
		Definition:           definition.Definition,
		RoleArn:              definition.RoleArn,
		Type:                 definition.Type,
		CreationDate:         aws.ToTime(result.CreationDate),
		LoggingConfiguration: mergedOptions.LoggingConfiguration,
		TracingConfiguration: mergedOptions.TracingConfiguration,
	}
	
	logResult("CreateStateMachine", true, time.Since(start), nil)
	return info, nil
}

// UpdateStateMachine updates an existing Step Functions state machine
func UpdateStateMachine(ctx *TestContext, stateMachineArn, definition, roleArn string, options *StateMachineOptions) *StateMachineInfo {
	info, err := UpdateStateMachineE(ctx, stateMachineArn, definition, roleArn, options)
	if err != nil {
		ctx.T.FailNow()
	}
	return info
}

// UpdateStateMachineE updates an existing Step Functions state machine with error return
func UpdateStateMachineE(ctx *TestContext, stateMachineArn, definition, roleArn string, options *StateMachineOptions) (*StateMachineInfo, error) {
	start := time.Now()
	logOperation("UpdateStateMachine", map[string]interface{}{
		"arn": stateMachineArn,
	})
	
	// Validate input parameters
	if err := validateStateMachineUpdate(stateMachineArn, definition, roleArn, options); err != nil {
		logResult("UpdateStateMachine", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.UpdateStateMachineInput{
		StateMachineArn: aws.String(stateMachineArn),
	}
	
	// Only set fields that are provided
	if definition != "" {
		input.Definition = aws.String(definition)
	}
	if roleArn != "" {
		input.RoleArn = aws.String(roleArn)
	}
	
	// Add optional configurations
	if options != nil {
		if options.LoggingConfiguration != nil {
			input.LoggingConfiguration = options.LoggingConfiguration
		}
		if options.TracingConfiguration != nil {
			input.TracingConfiguration = options.TracingConfiguration
		}
	}
	
	_, err := client.UpdateStateMachine(context.TODO(), input)
	if err != nil {
		logResult("UpdateStateMachine", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to update state machine: %w", err)
	}
	
	// Get updated state machine information
	info, err := DescribeStateMachineE(ctx, stateMachineArn)
	if err != nil {
		logResult("UpdateStateMachine", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to describe updated state machine: %w", err)
	}
	
	logResult("UpdateStateMachine", true, time.Since(start), nil)
	return info, nil
}

// DescribeStateMachine gets information about a Step Functions state machine
func DescribeStateMachine(ctx *TestContext, stateMachineArn string) *StateMachineInfo {
	info, err := DescribeStateMachineE(ctx, stateMachineArn)
	if err != nil {
		ctx.T.FailNow()
	}
	return info
}

// DescribeStateMachineE gets information about a Step Functions state machine with error return
func DescribeStateMachineE(ctx *TestContext, stateMachineArn string) (*StateMachineInfo, error) {
	start := time.Now()
	logOperation("DescribeStateMachine", map[string]interface{}{
		"arn": stateMachineArn,
	})
	
	// Validate input parameters
	if err := validateStateMachineArn(stateMachineArn); err != nil {
		logResult("DescribeStateMachine", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.DescribeStateMachineInput{
		StateMachineArn: aws.String(stateMachineArn),
	}
	
	result, err := client.DescribeStateMachine(context.TODO(), input)
	if err != nil {
		logResult("DescribeStateMachine", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to describe state machine: %w", err)
	}
	
	// Convert result to our standard format
	info := &StateMachineInfo{
		StateMachineArn:      aws.ToString(result.StateMachineArn),
		Name:                 aws.ToString(result.Name),
		Status:               result.Status,
		Definition:           aws.ToString(result.Definition),
		RoleArn:              aws.ToString(result.RoleArn),
		Type:                 result.Type,
		CreationDate:         aws.ToTime(result.CreationDate),
		LoggingConfiguration: result.LoggingConfiguration,
		TracingConfiguration: result.TracingConfiguration,
	}
	
	logResult("DescribeStateMachine", true, time.Since(start), nil)
	return processStateMachineInfo(info), nil
}

// DeleteStateMachine deletes a Step Functions state machine
func DeleteStateMachine(ctx *TestContext, stateMachineArn string) {
	err := DeleteStateMachineE(ctx, stateMachineArn)
	if err != nil {
		ctx.T.FailNow()
	}
}

// DeleteStateMachineE deletes a Step Functions state machine with error return
func DeleteStateMachineE(ctx *TestContext, stateMachineArn string) error {
	start := time.Now()
	logOperation("DeleteStateMachine", map[string]interface{}{
		"arn": stateMachineArn,
	})
	
	// Validate input parameters
	if err := validateStateMachineArn(stateMachineArn); err != nil {
		logResult("DeleteStateMachine", false, time.Since(start), err)
		return err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.DeleteStateMachineInput{
		StateMachineArn: aws.String(stateMachineArn),
	}
	
	_, err := client.DeleteStateMachine(context.TODO(), input)
	if err != nil {
		logResult("DeleteStateMachine", false, time.Since(start), err)
		return fmt.Errorf("failed to delete state machine: %w", err)
	}
	
	logResult("DeleteStateMachine", true, time.Since(start), nil)
	return nil
}

// ListStateMachines lists Step Functions state machines
func ListStateMachines(ctx *TestContext, maxResults int32) []*StateMachineInfo {
	infos, err := ListStateMachinesE(ctx, maxResults)
	if err != nil {
		ctx.T.FailNow()
	}
	return infos
}

// ListStateMachinesE lists Step Functions state machines with error return
func ListStateMachinesE(ctx *TestContext, maxResults int32) ([]*StateMachineInfo, error) {
	start := time.Now()
	logOperation("ListStateMachines", map[string]interface{}{
		"max_results": maxResults,
	})
	
	// Validate input parameters
	if err := validateMaxResults(maxResults); err != nil {
		logResult("ListStateMachines", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.ListStateMachinesInput{}
	if maxResults > 0 {
		input.MaxResults = maxResults
	}
	
	result, err := client.ListStateMachines(context.TODO(), input)
	if err != nil {
		logResult("ListStateMachines", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to list state machines: %w", err)
	}
	
	// Convert results to our standard format
	var infos []*StateMachineInfo
	for _, sm := range result.StateMachines {
		info := &StateMachineInfo{
			StateMachineArn: aws.ToString(sm.StateMachineArn),
			Name:            aws.ToString(sm.Name),
			Type:            sm.Type,
			CreationDate:    aws.ToTime(sm.CreationDate),
		}
		infos = append(infos, processStateMachineInfo(info))
	}
	
	logResult("ListStateMachines", true, time.Since(start), nil)
	return infos, nil
}

// Execution Management

// StartExecution starts a new Step Functions execution
func StartExecution(ctx *TestContext, stateMachineArn, executionName string, input *Input) *ExecutionResult {
	result, err := StartExecutionE(ctx, stateMachineArn, executionName, input)
	if err != nil {
		ctx.T.FailNow()
	}
	return result
}

// StartExecutionE starts a new Step Functions execution with error return
func StartExecutionE(ctx *TestContext, stateMachineArn, executionName string, input *Input) (*ExecutionResult, error) {
	start := time.Now()
	logOperation("StartExecution", map[string]interface{}{
		"state_machine": stateMachineArn,
		"name":         executionName,
	})
	
	// Build input JSON
	var inputJSON string
	var err error
	if input != nil && !input.isEmpty() {
		inputJSON, err = input.ToJSON()
		if err != nil {
			logResult("StartExecution", false, time.Since(start), err)
			return nil, fmt.Errorf("failed to convert input to JSON: %w", err)
		}
	}
	
	// Create request and validate
	request := &StartExecutionRequest{
		StateMachineArn: stateMachineArn,
		Name:            executionName,
		Input:           inputJSON,
	}
	
	if err := validateStartExecutionRequest(request); err != nil {
		logResult("StartExecution", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	sfnInput := &sfn.StartExecutionInput{
		StateMachineArn: aws.String(request.StateMachineArn),
		Name:            aws.String(request.Name),
	}
	
	if request.Input != "" {
		sfnInput.Input = aws.String(request.Input)
	}
	if request.TraceHeader != "" {
		sfnInput.TraceHeader = aws.String(request.TraceHeader)
	}
	
	result, err := client.StartExecution(context.TODO(), sfnInput)
	if err != nil {
		logResult("StartExecution", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to start execution: %w", err)
	}
	
	// Convert result to our standard format
	execResult := &ExecutionResult{
		ExecutionArn:    aws.ToString(result.ExecutionArn),
		StateMachineArn: stateMachineArn,
		Name:            executionName,
		Status:          types.ExecutionStatusRunning, // New executions start as running
		StartDate:       aws.ToTime(result.StartDate),
		Input:           inputJSON,
	}
	
	logResult("StartExecution", true, time.Since(start), nil)
	return processExecutionResult(execResult), nil
}

// StopExecution stops a running Step Functions execution
func StopExecution(ctx *TestContext, executionArn, error, cause string) *ExecutionResult {
	result, err := StopExecutionE(ctx, executionArn, error, cause)
	if err != nil {
		ctx.T.FailNow()
	}
	return result
}

// StopExecutionE stops a running Step Functions execution with error return
func StopExecutionE(ctx *TestContext, executionArn, error, cause string) (*ExecutionResult, error) {
	start := time.Now()
	logOperation("StopExecution", map[string]interface{}{
		"execution": executionArn,
		"error":     error,
	})
	
	// Validate input parameters
	if err := validateStopExecutionRequest(executionArn, error, cause); err != nil {
		logResult("StopExecution", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.StopExecutionInput{
		ExecutionArn: aws.String(executionArn),
	}
	
	if error != "" {
		input.Error = aws.String(error)
	}
	if cause != "" {
		input.Cause = aws.String(cause)
	}
	
	_, err := client.StopExecution(context.TODO(), input)
	if err != nil {
		logResult("StopExecution", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to stop execution: %w", err)
	}
	
	// Get updated execution information
	execResult, err := DescribeExecutionE(ctx, executionArn)
	if err != nil {
		logResult("StopExecution", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to describe stopped execution: %w", err)
	}
	
	logResult("StopExecution", true, time.Since(start), nil)
	return execResult, nil
}

// DescribeExecution gets information about a Step Functions execution
func DescribeExecution(ctx *TestContext, executionArn string) *ExecutionResult {
	result, err := DescribeExecutionE(ctx, executionArn)
	if err != nil {
		ctx.T.FailNow()
	}
	return result
}

// DescribeExecutionE gets information about a Step Functions execution with error return
func DescribeExecutionE(ctx *TestContext, executionArn string) (*ExecutionResult, error) {
	start := time.Now()
	logOperation("DescribeExecution", map[string]interface{}{
		"execution": executionArn,
	})
	
	// Validate input parameters
	if err := validateExecutionArn(executionArn); err != nil {
		logResult("DescribeExecution", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.DescribeExecutionInput{
		ExecutionArn: aws.String(executionArn),
	}
	
	result, err := client.DescribeExecution(context.TODO(), input)
	if err != nil {
		logResult("DescribeExecution", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to describe execution: %w", err)
	}
	
	// Convert result to our standard format
	execResult := &ExecutionResult{
		ExecutionArn:    aws.ToString(result.ExecutionArn),
		StateMachineArn: aws.ToString(result.StateMachineArn),
		Name:            aws.ToString(result.Name),
		Status:          result.Status,
		StartDate:       aws.ToTime(result.StartDate),
		StopDate:        aws.ToTime(result.StopDate),
		Input:           aws.ToString(result.Input),
		Output:          aws.ToString(result.Output),
		Error:           aws.ToString(result.Error),
		Cause:           aws.ToString(result.Cause),
		MapRunArn:       aws.ToString(result.MapRunArn),
		RedriveCount:    int(aws.ToInt32(result.RedriveCount)),
		RedriveDate:     aws.ToTime(result.RedriveDate),
	}
	
	// Calculate execution time
	if !execResult.StartDate.IsZero() && !execResult.StopDate.IsZero() {
		execResult.ExecutionTime = execResult.StopDate.Sub(execResult.StartDate)
	}
	
	logResult("DescribeExecution", true, time.Since(start), nil)
	return processExecutionResult(execResult), nil
}

// WaitForExecution waits for a Step Functions execution to complete
func WaitForExecution(ctx *TestContext, executionArn string, options *WaitOptions) *ExecutionResult {
	result, err := WaitForExecutionE(ctx, executionArn, options)
	if err != nil {
		ctx.T.FailNow()
	}
	return result
}

// WaitForExecutionE waits for a Step Functions execution to complete with error return
func WaitForExecutionE(ctx *TestContext, executionArn string, options *WaitOptions) (*ExecutionResult, error) {
	start := time.Now()
	logOperation("WaitForExecution", map[string]interface{}{
		"execution": executionArn,
	})
	
	// Validate input parameters
	if err := validateWaitForExecutionRequest(executionArn, options); err != nil {
		logResult("WaitForExecution", false, time.Since(start), err)
		return nil, err
	}
	
	// Merge with defaults
	mergedOptions := mergeWaitOptions(options)
	
	// Convert to PollConfig
	pollConfig := &PollConfig{
		MaxAttempts:        mergedOptions.MaxAttempts,
		Interval:           mergedOptions.PollingInterval,
		Timeout:            mergedOptions.Timeout,
		ExponentialBackoff: false, // Keep simple for basic wait
		BackoffMultiplier:  1.0,
		MaxInterval:        mergedOptions.PollingInterval,
	}
	
	result, err := pollUntilComplete(ctx, executionArn, pollConfig)
	if err != nil {
		logResult("WaitForExecution", false, time.Since(start), err)
		return nil, err
	}
	
	logResult("WaitForExecution", true, time.Since(start), nil)
	return result, nil
}

// ExecuteStateMachineAndWait executes a state machine and waits for completion
func ExecuteStateMachineAndWait(ctx *TestContext, stateMachineArn, executionName string, input *Input, timeout time.Duration) *ExecutionResult {
	result, err := ExecuteStateMachineAndWaitE(ctx, stateMachineArn, executionName, input, timeout)
	if err != nil {
		ctx.T.FailNow()
	}
	return result
}

// ExecuteStateMachineAndWaitE executes a state machine and waits for completion with error return
func ExecuteStateMachineAndWaitE(ctx *TestContext, stateMachineArn, executionName string, input *Input, timeout time.Duration) (*ExecutionResult, error) {
	start := time.Now()
	logOperation("ExecuteStateMachineAndWait", map[string]interface{}{
		"state_machine": stateMachineArn,
		"execution":     executionName,
		"timeout":       timeout.String(),
	})
	
	// Validate parameters
	if err := validateExecuteAndWaitPattern(stateMachineArn, executionName, input, timeout); err != nil {
		logResult("ExecuteStateMachineAndWait", false, time.Since(start), err)
		return nil, err
	}
	
	// Start execution
	execResult, err := StartExecutionE(ctx, stateMachineArn, executionName, input)
	if err != nil {
		logResult("ExecuteStateMachineAndWait", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to start execution: %w", err)
	}
	
	// Wait for completion
	waitOptions := &WaitOptions{
		Timeout:         timeout,
		PollingInterval: DefaultPollingInterval,
		MaxAttempts:     int(timeout / DefaultPollingInterval),
	}
	
	finalResult, err := WaitForExecutionE(ctx, execResult.ExecutionArn, waitOptions)
	if err != nil {
		logResult("ExecuteStateMachineAndWait", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to wait for execution: %w", err)
	}
	
	logResult("ExecuteStateMachineAndWait", true, time.Since(start), nil)
	return finalResult, nil
}

// PollUntilComplete polls an execution until it reaches a terminal state
func PollUntilComplete(ctx *TestContext, executionArn string, config *PollConfig) *ExecutionResult {
	result, err := PollUntilCompleteE(ctx, executionArn, config)
	if err != nil {
		ctx.T.FailNow()
	}
	return result
}

// PollUntilCompleteE polls an execution until it reaches a terminal state with error return
func PollUntilCompleteE(ctx *TestContext, executionArn string, config *PollConfig) (*ExecutionResult, error) {
	start := time.Now()
	logOperation("PollUntilComplete", map[string]interface{}{
		"execution": executionArn,
	})
	
	// Validate parameters
	if err := validatePollConfig(executionArn, config); err != nil {
		logResult("PollUntilComplete", false, time.Since(start), err)
		return nil, err
	}
	
	result, err := pollUntilComplete(ctx, executionArn, config)
	if err != nil {
		logResult("PollUntilComplete", false, time.Since(start), err)
		return nil, err
	}
	
	logResult("PollUntilComplete", true, time.Since(start), nil)
	return result, nil
}

// pollUntilComplete is the internal polling implementation
func pollUntilComplete(ctx *TestContext, executionArn string, config *PollConfig) (*ExecutionResult, error) {
	mergedConfig := mergePollConfig(config)
	startTime := time.Now()
	
	for attempt := 1; ; attempt++ {
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
		if !shouldContinuePolling(attempt, elapsed, mergedConfig) {
			return nil, fmt.Errorf("%w: execution did not complete after %d attempts in %v", 
				ErrExecutionTimeout, attempt-1, elapsed)
		}
		
		// Calculate next interval and sleep
		interval := calculateNextPollInterval(attempt, mergedConfig)
		time.Sleep(interval)
	}
}

// Utility functions for operations

// mergeStateMachineOptions merges user options with defaults
func mergeStateMachineOptions(userOpts *StateMachineOptions) *StateMachineOptions {
	if userOpts == nil {
		return &StateMachineOptions{}
	}
	
	// Return a copy to prevent mutation
	merged := *userOpts
	return &merged
}

// convertTagsToSFN converts string map tags to Step Functions tag format
func convertTagsToSFN(tags map[string]string) []types.Tag {
	if tags == nil {
		return nil
	}
	
	var sfnTags []types.Tag
	for key, value := range tags {
		sfnTags = append(sfnTags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	
	return sfnTags
}