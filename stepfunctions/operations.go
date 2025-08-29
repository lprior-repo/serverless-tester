// Package stepfunctions operations - AWS Step Functions lifecycle management and execution operations
// Following strict Terratest patterns with Function/FunctionE variants

package stepfunctions

import (
	"context"
	"encoding/json"
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
	
	// Validate input parameters first
	if err := validateStateMachineCreation(definition, options); err != nil {
		logResult("CreateStateMachine", false, time.Since(start), err)
		return nil, err
	}
	
	logOperation("CreateStateMachine", map[string]interface{}{
		"name": definition.Name,
		"type": definition.Type,
	})
	
	// Continue with AWS client operations
	
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
	
	// Convert results to our standard format - initialize empty slice instead of nil
	infos := make([]*StateMachineInfo, 0, len(result.StateMachines))
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

// Tag Management for State Machines

// TagStateMachine adds tags to a state machine
func TagStateMachine(ctx *TestContext, stateMachineArn string, tags map[string]string) {
	err := TagStateMachineE(ctx, stateMachineArn, tags)
	if err != nil {
		ctx.T.FailNow()
	}
}

// TagStateMachineE adds tags to a state machine with error return
func TagStateMachineE(ctx *TestContext, stateMachineArn string, tags map[string]string) error {
	start := time.Now()
	logOperation("TagStateMachine", map[string]interface{}{
		"state_machine": stateMachineArn,
		"tag_count":     len(tags),
	})

	// Validate input parameters
	if err := validateStateMachineArn(stateMachineArn); err != nil {
		logResult("TagStateMachine", false, time.Since(start), err)
		return err
	}

	if len(tags) == 0 {
		err := fmt.Errorf("at least one tag is required")
		logResult("TagStateMachine", false, time.Since(start), err)
		return err
	}

	client := createStepFunctionsClient(ctx)

	// Convert tags to AWS SDK format
	var awsTags []types.Tag
	for key, value := range tags {
		awsTags = append(awsTags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	input := &sfn.TagResourceInput{
		ResourceArn: aws.String(stateMachineArn),
		Tags:        awsTags,
	}

	_, err := client.TagResource(context.TODO(), input)
	if err != nil {
		logResult("TagStateMachine", false, time.Since(start), err)
		return fmt.Errorf("failed to tag state machine: %w", err)
	}

	logResult("TagStateMachine", true, time.Since(start), nil)
	return nil
}

// UntagStateMachine removes tags from a state machine
func UntagStateMachine(ctx *TestContext, stateMachineArn string, tagKeys []string) {
	err := UntagStateMachineE(ctx, stateMachineArn, tagKeys)
	if err != nil {
		ctx.T.FailNow()
	}
}

// UntagStateMachineE removes tags from a state machine with error return
func UntagStateMachineE(ctx *TestContext, stateMachineArn string, tagKeys []string) error {
	start := time.Now()
	logOperation("UntagStateMachine", map[string]interface{}{
		"state_machine": stateMachineArn,
		"key_count":     len(tagKeys),
	})

	// Validate input parameters
	if err := validateStateMachineArn(stateMachineArn); err != nil {
		logResult("UntagStateMachine", false, time.Since(start), err)
		return err
	}

	if len(tagKeys) == 0 {
		err := fmt.Errorf("at least one tag key is required")
		logResult("UntagStateMachine", false, time.Since(start), err)
		return err
	}

	client := createStepFunctionsClient(ctx)

	input := &sfn.UntagResourceInput{
		ResourceArn: aws.String(stateMachineArn),
		TagKeys:     tagKeys,
	}

	_, err := client.UntagResource(context.TODO(), input)
	if err != nil {
		logResult("UntagStateMachine", false, time.Since(start), err)
		return fmt.Errorf("failed to untag state machine: %w", err)
	}

	logResult("UntagStateMachine", true, time.Since(start), nil)
	return nil
}

// ListStateMachineTags lists all tags for a state machine
func ListStateMachineTags(ctx *TestContext, stateMachineArn string) map[string]string {
	tags, err := ListStateMachineTagsE(ctx, stateMachineArn)
	if err != nil {
		ctx.T.FailNow()
	}
	return tags
}

// ListStateMachineTagsE lists all tags for a state machine with error return
func ListStateMachineTagsE(ctx *TestContext, stateMachineArn string) (map[string]string, error) {
	start := time.Now()
	logOperation("ListStateMachineTags", map[string]interface{}{
		"state_machine": stateMachineArn,
	})

	// Validate input parameters
	if err := validateStateMachineArn(stateMachineArn); err != nil {
		logResult("ListStateMachineTags", false, time.Since(start), err)
		return nil, err
	}

	client := createStepFunctionsClient(ctx)

	input := &sfn.ListTagsForResourceInput{
		ResourceArn: aws.String(stateMachineArn),
	}

	result, err := client.ListTagsForResource(context.TODO(), input)
	if err != nil {
		logResult("ListStateMachineTags", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to list state machine tags: %w", err)
	}

	// Convert AWS tags to map
	tags := make(map[string]string)
	for _, tag := range result.Tags {
		if tag.Key != nil && tag.Value != nil {
			tags[*tag.Key] = *tag.Value
		}
	}

	logResult("ListStateMachineTags", true, time.Since(start), nil)
	return tags, nil
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
	return execResult, nil
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

// Activity Management Operations

// CreateActivity creates a new Step Functions activity
func CreateActivity(ctx *TestContext, activityName string, tags map[string]string) *ActivityInfo {
	info, err := CreateActivityE(ctx, activityName, tags)
	if err != nil {
		ctx.T.FailNow()
	}
	return info
}

// CreateActivityE creates a new Step Functions activity with error return
func CreateActivityE(ctx *TestContext, activityName string, tags map[string]string) (*ActivityInfo, error) {
	start := time.Now()
	
	// Validate input parameters first
	if err := validateActivityName(activityName); err != nil {
		logResult("CreateActivity", false, time.Since(start), err)
		return nil, err
	}
	
	logOperation("CreateActivity", map[string]interface{}{
		"name":      activityName,
		"tag_count": len(tags),
	})
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.CreateActivityInput{
		Name: aws.String(activityName),
	}
	
	// Add tags if provided
	if len(tags) > 0 {
		input.Tags = convertTagsToSFN(tags)
	}
	
	result, err := client.CreateActivity(context.TODO(), input)
	if err != nil {
		logResult("CreateActivity", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}
	
	// Convert result to our standard format
	info := &ActivityInfo{
		ActivityArn:  aws.ToString(result.ActivityArn),
		Name:         activityName,
		CreationDate: aws.ToTime(result.CreationDate),
		Tags:         tags,
	}
	
	logResult("CreateActivity", true, time.Since(start), nil)
	return info, nil
}

// DeleteActivity deletes a Step Functions activity
func DeleteActivity(ctx *TestContext, activityArn string) {
	err := DeleteActivityE(ctx, activityArn)
	if err != nil {
		ctx.T.FailNow()
	}
}

// DeleteActivityE deletes a Step Functions activity with error return
func DeleteActivityE(ctx *TestContext, activityArn string) error {
	start := time.Now()
	logOperation("DeleteActivity", map[string]interface{}{
		"arn": activityArn,
	})
	
	// Validate input parameters
	if err := validateActivityArn(activityArn); err != nil {
		logResult("DeleteActivity", false, time.Since(start), err)
		return err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.DeleteActivityInput{
		ActivityArn: aws.String(activityArn),
	}
	
	_, err := client.DeleteActivity(context.TODO(), input)
	if err != nil {
		logResult("DeleteActivity", false, time.Since(start), err)
		return fmt.Errorf("failed to delete activity: %w", err)
	}
	
	logResult("DeleteActivity", true, time.Since(start), nil)
	return nil
}

// DescribeActivity gets information about a Step Functions activity
func DescribeActivity(ctx *TestContext, activityArn string) *ActivityInfo {
	info, err := DescribeActivityE(ctx, activityArn)
	if err != nil {
		ctx.T.FailNow()
	}
	return info
}

// DescribeActivityE gets information about a Step Functions activity with error return
func DescribeActivityE(ctx *TestContext, activityArn string) (*ActivityInfo, error) {
	start := time.Now()
	logOperation("DescribeActivity", map[string]interface{}{
		"arn": activityArn,
	})
	
	// Validate input parameters
	if err := validateActivityArn(activityArn); err != nil {
		logResult("DescribeActivity", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.DescribeActivityInput{
		ActivityArn: aws.String(activityArn),
	}
	
	result, err := client.DescribeActivity(context.TODO(), input)
	if err != nil {
		logResult("DescribeActivity", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to describe activity: %w", err)
	}
	
	// Convert result to our standard format
	info := &ActivityInfo{
		ActivityArn:  aws.ToString(result.ActivityArn),
		Name:         aws.ToString(result.Name),
		CreationDate: aws.ToTime(result.CreationDate),
	}
	
	logResult("DescribeActivity", true, time.Since(start), nil)
	return info, nil
}

// ListActivities lists Step Functions activities
func ListActivities(ctx *TestContext, maxResults int32) []*ActivityInfo {
	infos, err := ListActivitiesE(ctx, maxResults)
	if err != nil {
		ctx.T.FailNow()
	}
	return infos
}

// ListActivitiesE lists Step Functions activities with error return
func ListActivitiesE(ctx *TestContext, maxResults int32) ([]*ActivityInfo, error) {
	start := time.Now()
	logOperation("ListActivities", map[string]interface{}{
		"max_results": maxResults,
	})
	
	// Validate input parameters
	if err := validateMaxResults(maxResults); err != nil {
		logResult("ListActivities", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.ListActivitiesInput{}
	if maxResults > 0 {
		input.MaxResults = maxResults
	}
	
	result, err := client.ListActivities(context.TODO(), input)
	if err != nil {
		logResult("ListActivities", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}
	
	// Convert results to our standard format - initialize empty slice instead of nil
	infos := make([]*ActivityInfo, 0, len(result.Activities))
	for _, activity := range result.Activities {
		info := &ActivityInfo{
			ActivityArn:  aws.ToString(activity.ActivityArn),
			Name:         aws.ToString(activity.Name),
			CreationDate: aws.ToTime(activity.CreationDate),
		}
		infos = append(infos, info)
	}
	
	logResult("ListActivities", true, time.Since(start), nil)
	return infos, nil
}

// Task Communication Operations

// GetActivityTask retrieves a task from an activity
func GetActivityTask(ctx *TestContext, activityArn, workerName string) *ActivityTask {
	task, err := GetActivityTaskE(ctx, activityArn, workerName)
	if err != nil {
		ctx.T.FailNow()
	}
	return task
}

// GetActivityTaskE retrieves a task from an activity with error return
func GetActivityTaskE(ctx *TestContext, activityArn, workerName string) (*ActivityTask, error) {
	start := time.Now()
	logOperation("GetActivityTask", map[string]interface{}{
		"activity":    activityArn,
		"worker_name": workerName,
	})
	
	// Validate input parameters
	if err := validateActivityArn(activityArn); err != nil {
		logResult("GetActivityTask", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.GetActivityTaskInput{
		ActivityArn: aws.String(activityArn),
	}
	
	// Worker name is optional
	if workerName != "" {
		input.WorkerName = aws.String(workerName)
	}
	
	result, err := client.GetActivityTask(context.TODO(), input)
	if err != nil {
		logResult("GetActivityTask", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to get activity task: %w", err)
	}
	
	// Convert result to our standard format
	task := &ActivityTask{
		TaskToken: aws.ToString(result.TaskToken),
		Input:     aws.ToString(result.Input),
	}
	
	logResult("GetActivityTask", true, time.Since(start), nil)
	return task, nil
}

// SendTaskSuccess sends a success result for a task
func SendTaskSuccess(ctx *TestContext, taskToken, output string) {
	err := SendTaskSuccessE(ctx, taskToken, output)
	if err != nil {
		ctx.T.FailNow()
	}
}

// SendTaskSuccessE sends a success result for a task with error return
func SendTaskSuccessE(ctx *TestContext, taskToken, output string) error {
	start := time.Now()
	logOperation("SendTaskSuccess", map[string]interface{}{
		"task_token": taskToken,
	})
	
	// Validate input parameters
	if err := validateTaskToken(taskToken); err != nil {
		logResult("SendTaskSuccess", false, time.Since(start), err)
		return err
	}
	
	// Output is optional but must be valid JSON if provided
	if output != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
			logResult("SendTaskSuccess", false, time.Since(start), err)
			return fmt.Errorf("invalid JSON output: %w", err)
		}
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.SendTaskSuccessInput{
		TaskToken: aws.String(taskToken),
		Output:    aws.String(output),
	}
	
	_, err := client.SendTaskSuccess(context.TODO(), input)
	if err != nil {
		logResult("SendTaskSuccess", false, time.Since(start), err)
		return fmt.Errorf("failed to send task success: %w", err)
	}
	
	logResult("SendTaskSuccess", true, time.Since(start), nil)
	return nil
}

// SendTaskFailure sends a failure result for a task
func SendTaskFailure(ctx *TestContext, taskToken, error, cause string) {
	err := SendTaskFailureE(ctx, taskToken, error, cause)
	if err != nil {
		ctx.T.FailNow()
	}
}

// SendTaskFailureE sends a failure result for a task with error return
func SendTaskFailureE(ctx *TestContext, taskToken, error, cause string) error {
	start := time.Now()
	logOperation("SendTaskFailure", map[string]interface{}{
		"task_token": taskToken,
		"error":      error,
	})
	
	// Validate input parameters
	if err := validateTaskToken(taskToken); err != nil {
		logResult("SendTaskFailure", false, time.Since(start), err)
		return err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.SendTaskFailureInput{
		TaskToken: aws.String(taskToken),
	}
	
	// Error and cause are optional
	if error != "" {
		input.Error = aws.String(error)
	}
	if cause != "" {
		input.Cause = aws.String(cause)
	}
	
	_, err := client.SendTaskFailure(context.TODO(), input)
	if err != nil {
		logResult("SendTaskFailure", false, time.Since(start), err)
		return fmt.Errorf("failed to send task failure: %w", err)
	}
	
	logResult("SendTaskFailure", true, time.Since(start), nil)
	return nil
}

// SendTaskHeartbeat sends a heartbeat for a task
func SendTaskHeartbeat(ctx *TestContext, taskToken string) {
	err := SendTaskHeartbeatE(ctx, taskToken)
	if err != nil {
		ctx.T.FailNow()
	}
}

// SendTaskHeartbeatE sends a heartbeat for a task with error return
func SendTaskHeartbeatE(ctx *TestContext, taskToken string) error {
	start := time.Now()
	logOperation("SendTaskHeartbeat", map[string]interface{}{
		"task_token": taskToken,
	})
	
	// Validate input parameters
	if err := validateTaskToken(taskToken); err != nil {
		logResult("SendTaskHeartbeat", false, time.Since(start), err)
		return err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.SendTaskHeartbeatInput{
		TaskToken: aws.String(taskToken),
	}
	
	_, err := client.SendTaskHeartbeat(context.TODO(), input)
	if err != nil {
		logResult("SendTaskHeartbeat", false, time.Since(start), err)
		return fmt.Errorf("failed to send task heartbeat: %w", err)
	}
	
	logResult("SendTaskHeartbeat", true, time.Since(start), nil)
	return nil
}

// Synchronous Execution Operations

// StartSyncExecution starts a synchronous Step Functions execution
func StartSyncExecution(ctx *TestContext, stateMachineArn, name string, input *Input) *SyncExecutionResult {
	result, err := StartSyncExecutionE(ctx, stateMachineArn, name, input)
	if err != nil {
		ctx.T.FailNow()
	}
	return result
}

// StartSyncExecutionE starts a synchronous Step Functions execution with error return
func StartSyncExecutionE(ctx *TestContext, stateMachineArn, name string, input *Input) (*SyncExecutionResult, error) {
	start := time.Now()
	logOperation("StartSyncExecution", map[string]interface{}{
		"state_machine": stateMachineArn,
		"name":         name,
	})
	
	// Build input JSON
	var inputJSON string
	var err error
	if input != nil && !input.isEmpty() {
		inputJSON, err = input.ToJSON()
		if err != nil {
			logResult("StartSyncExecution", false, time.Since(start), err)
			return nil, fmt.Errorf("failed to convert input to JSON: %w", err)
		}
	}
	
	// Create request and validate
	request := &StartExecutionRequest{
		StateMachineArn: stateMachineArn,
		Name:            name,
		Input:           inputJSON,
	}
	
	if err := validateStartExecutionRequest(request); err != nil {
		logResult("StartSyncExecution", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	sfnInput := &sfn.StartSyncExecutionInput{
		StateMachineArn: aws.String(request.StateMachineArn),
		Name:            aws.String(request.Name),
	}
	
	if request.Input != "" {
		sfnInput.Input = aws.String(request.Input)
	}
	if request.TraceHeader != "" {
		sfnInput.TraceHeader = aws.String(request.TraceHeader)
	}
	
	result, err := client.StartSyncExecution(context.TODO(), sfnInput)
	if err != nil {
		logResult("StartSyncExecution", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to start synchronous execution: %w", err)
	}
	
	// Convert result to our standard format
	syncResult := &SyncExecutionResult{
		ExecutionArn:   aws.ToString(result.ExecutionArn),
		StateMachineArn: stateMachineArn,
		Name:           name,
		Status:         result.Status,
		StartDate:      aws.ToTime(result.StartDate),
		StopDate:       aws.ToTime(result.StopDate),
		Input:          inputJSON,
		Output:         aws.ToString(result.Output),
		Error:          aws.ToString(result.Error),
		Cause:          aws.ToString(result.Cause),
		BillingDetails: result.BillingDetails,
		TraceHeader:    request.TraceHeader,
	}
	
	// Calculate execution time
	if !syncResult.StartDate.IsZero() && !syncResult.StopDate.IsZero() {
		syncResult.ExecutionTime = syncResult.StopDate.Sub(syncResult.StartDate)
	}
	
	logResult("StartSyncExecution", true, time.Since(start), nil)
	return syncResult, nil
}

// DescribeStateMachineForExecution gets state machine information for a specific execution
func DescribeStateMachineForExecution(ctx *TestContext, executionArn string) *StateMachineInfo {
	info, err := DescribeStateMachineForExecutionE(ctx, executionArn)
	if err != nil {
		ctx.T.FailNow()
	}
	return info
}

// DescribeStateMachineForExecutionE gets state machine information for a specific execution with error return
func DescribeStateMachineForExecutionE(ctx *TestContext, executionArn string) (*StateMachineInfo, error) {
	start := time.Now()
	logOperation("DescribeStateMachineForExecution", map[string]interface{}{
		"execution": executionArn,
	})
	
	// Validate input parameters
	if err := validateExecutionArn(executionArn); err != nil {
		logResult("DescribeStateMachineForExecution", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.DescribeStateMachineForExecutionInput{
		ExecutionArn: aws.String(executionArn),
	}
	
	result, err := client.DescribeStateMachineForExecution(context.TODO(), input)
	if err != nil {
		logResult("DescribeStateMachineForExecution", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to describe state machine for execution: %w", err)
	}
	
	// Convert result to our standard format
	info := &StateMachineInfo{
		StateMachineArn:      aws.ToString(result.StateMachineArn),
		Name:                 aws.ToString(result.Name),
		Definition:           aws.ToString(result.Definition),
		RoleArn:              aws.ToString(result.RoleArn),
		LoggingConfiguration: result.LoggingConfiguration,
		TracingConfiguration: result.TracingConfiguration,
	}
	
	logResult("DescribeStateMachineForExecution", true, time.Since(start), nil)
	return processStateMachineInfo(info), nil
}

// Map Run Operations

// DescribeMapRun gets information about a Step Functions map run
func DescribeMapRun(ctx *TestContext, mapRunArn string) *MapRunInfo {
	info, err := DescribeMapRunE(ctx, mapRunArn)
	if err != nil {
		ctx.T.FailNow()
	}
	return info
}

// DescribeMapRunE gets information about a Step Functions map run with error return
func DescribeMapRunE(ctx *TestContext, mapRunArn string) (*MapRunInfo, error) {
	start := time.Now()
	logOperation("DescribeMapRun", map[string]interface{}{
		"map_run": mapRunArn,
	})
	
	// Validate input parameters
	if err := validateMapRunArn(mapRunArn); err != nil {
		logResult("DescribeMapRun", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.DescribeMapRunInput{
		MapRunArn: aws.String(mapRunArn),
	}
	
	result, err := client.DescribeMapRun(context.TODO(), input)
	if err != nil {
		logResult("DescribeMapRun", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to describe map run: %w", err)
	}
	
	// Convert result to our standard format
	info := &MapRunInfo{
		MapRunArn:           aws.ToString(result.MapRunArn),
		ExecutionArn:        aws.ToString(result.ExecutionArn),
		Status:              result.Status,
		StartDate:           aws.ToTime(result.StartDate),
		StopDate:            aws.ToTime(result.StopDate),
		MaxConcurrency:      result.MaxConcurrency,
		ToleratedFailurePercentage: float64(result.ToleratedFailurePercentage),
		ToleratedFailureCount:      result.ToleratedFailureCount,
	}
	
	// Convert execution counts
	if result.ExecutionCounts != nil {
		info.ExecutionCounts = make(map[string]int64)
		info.ExecutionCounts["succeeded"] = result.ExecutionCounts.Succeeded
		info.ExecutionCounts["failed"] = result.ExecutionCounts.Failed
		info.ExecutionCounts["running"] = result.ExecutionCounts.Running
		info.ExecutionCounts["aborted"] = result.ExecutionCounts.Aborted
		info.ExecutionCounts["timedOut"] = result.ExecutionCounts.TimedOut
		info.ExecutionCounts["total"] = result.ExecutionCounts.Total
	}
	
	logResult("DescribeMapRun", true, time.Since(start), nil)
	return info, nil
}

// ListMapRuns lists map runs for a specific execution
func ListMapRuns(ctx *TestContext, executionArn string, maxResults int32) []*MapRunInfo {
	infos, err := ListMapRunsE(ctx, executionArn, maxResults)
	if err != nil {
		ctx.T.FailNow()
	}
	return infos
}

// ListMapRunsE lists map runs for a specific execution with error return
func ListMapRunsE(ctx *TestContext, executionArn string, maxResults int32) ([]*MapRunInfo, error) {
	start := time.Now()
	logOperation("ListMapRuns", map[string]interface{}{
		"execution":   executionArn,
		"max_results": maxResults,
	})
	
	// Validate input parameters
	if err := validateExecutionArn(executionArn); err != nil {
		logResult("ListMapRuns", false, time.Since(start), err)
		return nil, err
	}
	
	if err := validateMaxResults(maxResults); err != nil {
		logResult("ListMapRuns", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.ListMapRunsInput{
		ExecutionArn: aws.String(executionArn),
	}
	
	if maxResults > 0 {
		input.MaxResults = maxResults
	}
	
	result, err := client.ListMapRuns(context.TODO(), input)
	if err != nil {
		logResult("ListMapRuns", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to list map runs: %w", err)
	}
	
	// Convert results to our standard format - initialize empty slice instead of nil
	infos := make([]*MapRunInfo, 0, len(result.MapRuns))
	for _, mapRun := range result.MapRuns {
		info := &MapRunInfo{
			MapRunArn:    aws.ToString(mapRun.MapRunArn),
			ExecutionArn: aws.ToString(mapRun.ExecutionArn),
			StartDate:    aws.ToTime(mapRun.StartDate),
			StopDate:     aws.ToTime(mapRun.StopDate),
		}
		infos = append(infos, info)
	}
	
	logResult("ListMapRuns", true, time.Since(start), nil)
	return infos, nil
}

// UpdateMapRun updates a map run status
func UpdateMapRun(ctx *TestContext, mapRunArn string, maxConcurrency int32, toleratedFailurePercentage float64) {
	err := UpdateMapRunE(ctx, mapRunArn, maxConcurrency, toleratedFailurePercentage)
	if err != nil {
		ctx.T.FailNow()
	}
}

// UpdateMapRunE updates a map run status with error return
func UpdateMapRunE(ctx *TestContext, mapRunArn string, maxConcurrency int32, toleratedFailurePercentage float64) error {
	start := time.Now()
	logOperation("UpdateMapRun", map[string]interface{}{
		"map_run":                      mapRunArn,
		"max_concurrency":              maxConcurrency,
		"tolerated_failure_percentage": toleratedFailurePercentage,
	})
	
	// Validate input parameters
	if err := validateMapRunArn(mapRunArn); err != nil {
		logResult("UpdateMapRun", false, time.Since(start), err)
		return err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.UpdateMapRunInput{
		MapRunArn: aws.String(mapRunArn),
	}
	
	// Add optional parameters
	if maxConcurrency > 0 {
		input.MaxConcurrency = &maxConcurrency
	}
	if toleratedFailurePercentage >= 0 {
		floatVal := float32(toleratedFailurePercentage)
		input.ToleratedFailurePercentage = &floatVal
	}
	
	_, err := client.UpdateMapRun(context.TODO(), input)
	if err != nil {
		logResult("UpdateMapRun", false, time.Since(start), err)
		return fmt.Errorf("failed to update map run: %w", err)
	}
	
	logResult("UpdateMapRun", true, time.Since(start), nil)
	return nil
}

// Additional Execution Operations

// ListExecutions lists executions for a state machine
func ListExecutions(ctx *TestContext, stateMachineArn string, statusFilter types.ExecutionStatus, maxResults int32) []*ExecutionResult {
	results, err := ListExecutionsE(ctx, stateMachineArn, statusFilter, maxResults)
	if err != nil {
		ctx.T.FailNow()
	}
	return results
}

// ListExecutionsE lists executions for a state machine with error return
// This replaces the stub implementation in assertions.go
func ListExecutionsE(ctx *TestContext, stateMachineArn string, statusFilter types.ExecutionStatus, maxResults int32) ([]*ExecutionResult, error) {
	start := time.Now()
	logOperation("ListExecutions", map[string]interface{}{
		"state_machine": stateMachineArn,
		"status_filter": statusFilter,
		"max_results":   maxResults,
	})
	
	// Validate input parameters
	if err := validateListExecutionsRequest(stateMachineArn, statusFilter, maxResults); err != nil {
		logResult("ListExecutions", false, time.Since(start), err)
		return nil, err
	}
	
	client := createStepFunctionsClient(ctx)
	
	input := &sfn.ListExecutionsInput{
		StateMachineArn: aws.String(stateMachineArn),
	}
	
	// Add optional parameters
	if maxResults > 0 {
		input.MaxResults = maxResults
	}
	
	// Status filter handling
	if string(statusFilter) != "" {
		input.StatusFilter = statusFilter
	}
	
	result, err := client.ListExecutions(context.TODO(), input)
	if err != nil {
		logResult("ListExecutions", false, time.Since(start), err)
		return nil, fmt.Errorf("failed to list executions: %w", err)
	}
	
	// Convert results to our standard format - initialize empty slice instead of nil
	executions := make([]*ExecutionResult, 0, len(result.Executions))
	for _, execution := range result.Executions {
		execResult := &ExecutionResult{
			ExecutionArn:    aws.ToString(execution.ExecutionArn),
			StateMachineArn: aws.ToString(execution.StateMachineArn),
			Name:            aws.ToString(execution.Name),
			Status:          execution.Status,
			StartDate:       aws.ToTime(execution.StartDate),
			StopDate:        aws.ToTime(execution.StopDate),
			MapRunArn:       aws.ToString(execution.MapRunArn),
			RedriveCount:    int(aws.ToInt32(execution.RedriveCount)),
			RedriveDate:     aws.ToTime(execution.RedriveDate),
		}
		
		// Calculate execution time if both dates are available
		if !execResult.StartDate.IsZero() && !execResult.StopDate.IsZero() {
			execResult.ExecutionTime = execResult.StopDate.Sub(execResult.StartDate)
		}
		
		executions = append(executions, execResult)
	}
	
	logResult("ListExecutions", true, time.Since(start), nil)
	return executions, nil
}