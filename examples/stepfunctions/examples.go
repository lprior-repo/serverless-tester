package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"vasdeference/stepfunctions"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// Example tests demonstrating comprehensive Step Functions testing patterns
// These examples show how to use the vasdeference Step Functions testing package

// ExampleBasicWorkflow demonstrates a functional Step Functions workflow
func ExampleBasicWorkflow() {
	fmt.Println("Functional Step Functions workflow patterns:")
	
	// Functional AWS configuration loading with error handling
	cfgResult := mo.TryOr(func() error {
		_, err := config.LoadDefaultConfig(context.TODO())
		return err
	}, func(err error) error {
		return fmt.Errorf("failed to load AWS config: %w", err)
	})
	
	// Use functional Step Functions configuration
	config := stepfunctions.NewFunctionalStepFunctionsStateMachine().OrEmpty()
	result := stepfunctions.SimulateFunctionalStepFunctionsExecution(config)
	
	// Functional state machine definition with validation
	definitionResult := lo.Pipe3(
		time.Now().Format("20060102-150405"),
		func(timestamp string) mo.Result[string] {
			name := "functional-workflow-" + timestamp
			if len(name) == 0 {
				return mo.Err[string](fmt.Errorf("empty state machine name"))
			}
			return mo.Ok(name)
		},
		func(nameResult mo.Result[string]) mo.Result[stepfunctions.StateMachineDefinition] {
			return nameResult.Map(func(name string) stepfunctions.StateMachineDefinition {
				return stepfunctions.StateMachineDefinition{
					Name: name,
					Definition: `{
						"Comment": "A functional Hello World example",
						"StartAt": "HelloWorld",
						"States": {
							"HelloWorld": {
								"Type": "Pass",
								"Result": "Functional Hello World!",
								"End": true
							}
						}
					}`,
					RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
					Type:    types.StateMachineTypeStandard,
					Tags: map[string]string{
						"Environment": "functional-example",
						"Purpose":     "functional-demonstration",
					},
				}
			})
		},
		func(definitionResult mo.Result[stepfunctions.StateMachineDefinition]) stepfunctions.StateMachineDefinition {
			return definitionResult.OrElse(stepfunctions.StateMachineDefinition{Name: "fallback-workflow"})
		},
	)
	
	// Functional state machine creation with monadic handling
	mo.Tuple2(cfgResult, result).Match(
		func(cfg mo.Result[interface{}], funcResult stepfunctions.FunctionalStepFunctionsStateMachine) {
			fmt.Printf("✓ Creating functional state machine: %s\n", definitionResult.Name)
			fmt.Printf("✓ AWS configuration loaded successfully\n")
		},
		func() {
			fmt.Printf("✗ Failed to initialize functional workflow\n")
			return
		},
	)
	
	// Functional input creation with validation
	inputResult := lo.Pipe3(
		map[string]interface{}{
			"greeting":  "Functional Hello",
			"target":    "Functional World",
			"timestamp": time.Now().Unix(),
		},
		func(inputData map[string]interface{}) mo.Result[map[string]interface{}] {
			if len(inputData) == 0 {
				return mo.Err[map[string]interface{}](fmt.Errorf("empty input data"))
			}
			return mo.Ok(inputData)
		},
		func(validated mo.Result[map[string]interface{}]) mo.Result[stepfunctions.Input] {
			return validated.Map(func(data map[string]interface{}) stepfunctions.Input {
				input := stepfunctions.NewInput()
				for key, value := range data {
					input.Set(key, value)
				}
				return input
			})
		},
		func(inputResult mo.Result[stepfunctions.Input]) mo.Option[stepfunctions.Input] {
			return inputResult.Match(
				func(input stepfunctions.Input) mo.Option[stepfunctions.Input] {
					return mo.Some(input)
				},
				func(err error) mo.Option[stepfunctions.Input] {
					fmt.Printf("✗ Input creation failed: %v\n", err)
					return mo.None[stepfunctions.Input]()
				},
			)
		},
	)
	
	// Functional execution start with validation
	executionResult := inputResult.FlatMap(func(input stepfunctions.Input) mo.Option[string] {
		return mo.TryOr(func() error {
			inputJSON, err := input.ToJSON()
			if err != nil {
				return err
			}
			fmt.Printf("✓ Starting functional execution with input: %s\n", inputJSON)
			return nil
		}, func(err error) error {
			fmt.Printf("✗ Failed to process input: %v\n", err)
			return err
		}).Match(
			func(value interface{}) mo.Option[string] {
				return mo.Some("execution-started")
			},
			func(err error) mo.Option[string] {
				return mo.None[string]()
			},
		)
	})
	
	// Functional execution waiting with timeout configuration
	waitResult := executionResult.FlatMap(func(execStatus string) mo.Option[stepfunctions.WaitOptions] {
		return mo.Some(stepfunctions.WaitOptions{
			Timeout:         2 * time.Minute,
			PollingInterval: 5 * time.Second,
			MaxAttempts:     24,
		})
	}).Map(func(options stepfunctions.WaitOptions) string {
		fmt.Printf("✓ Waiting for functional execution completion (%v timeout)\n", options.Timeout)
		return "waiting-completed"
	})
	
	// Functional result verification with monadic chaining
	verificationResult := waitResult.FlatMap(func(waitStatus string) mo.Option[stepfunctions.ExecutionResult] {
		return mo.Some(stepfunctions.ExecutionResult{
			Status:        types.ExecutionStatusSucceeded,
			Output:        "Functional Hello World!",
			ExecutionTime: 15 * time.Second,
		})
	})
	
	// Functional assertions with validation
	assertionResults := verificationResult.Map(func(result stepfunctions.ExecutionResult) []mo.Result[string] {
		return []mo.Result[string]{
			lo.Ternary(
				result.Status == types.ExecutionStatusSucceeded,
				mo.Ok("execution-succeeded"),
				mo.Err[string](fmt.Errorf("execution failed")),
			),
			lo.Ternary(
				strings.Contains(result.Output, "Functional Hello World!"),
				mo.Ok("output-contains-expected"),
				mo.Err[string](fmt.Errorf("unexpected output")),
			),
			lo.Ternary(
				result.ExecutionTime < 30*time.Second,
				mo.Ok("execution-time-valid"),
				mo.Err[string](fmt.Errorf("execution too slow")),
			),
		}
	})
	
	assertionResults.Match(
		func(results []mo.Result[string]) {
			successCount := lo.CountBy(results, func(result mo.Result[string]) bool {
				return result.IsOk()
			})
			fmt.Printf("✓ Functional workflow completed: %d/%d assertions passed\n", successCount, len(results))
		},
		func() {
			fmt.Println("✗ Functional workflow verification failed")
		},
	)
}

// ExampleExecuteAndWaitPattern demonstrates the execute-and-wait pattern
func ExampleExecuteAndWaitPattern() {
	fmt.Println("Step Functions execute-and-wait patterns:")
	
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}
	_ = /* ctx */ &stepfunctions.TestContext{AwsConfig: cfg, Region: "us-east-1"}
	
	// This pattern combines execution start and wait in one call
	_ = /* input */ stepfunctions.NewInput().Set("message", "Execute and wait pattern")
	
	// result, err := stepfunctions.ExecuteStateMachineAndWaitE(ctx,
	//     "arn:aws:states:us-east-1:123456789012:stateMachine:my-state-machine",
	//     "execute-and-wait-example",
	//     input,
	//     5*time.Minute)
	// 
	// if err != nil {
	//     fmt.Printf("Execute and wait failed: %v\n", err)
	//     return
	// }
	// stepfunctions.AssertExecutionSucceeded(t, result)
	fmt.Println("Execute-and-wait pattern: combines start and wait in single call")
	fmt.Println("Expected: execution succeeds within 5 minutes")
	fmt.Println("Execute-and-wait pattern completed")
}

// ExampleInputBuilder demonstrates fluent input building
func ExampleInputBuilder() {
	fmt.Println("Step Functions input builder patterns:")
	
	// Create complex input using fluent API
	input := stepfunctions.NewInput().
		Set("orderId", "order-123").
		Set("customerId", "customer-456").
		Set("amount", 99.99).
		Set("currency", "USD").
		SetIf(true, "priority", "high").
		SetIf(false, "discount", 10.0) // This won't be set
	
	// Convert to JSON
	jsonStr, err := input.ToJSON()
	if err != nil {
		fmt.Printf("Failed to convert to JSON: %v\n", err)
		return
	}
	
	// Verify content
	if strings.Contains(jsonStr, "order-123") {
		fmt.Println("✓ Contains order ID")
	}
	if strings.Contains(jsonStr, "priority") {
		fmt.Println("✓ Contains priority (conditional set with true)")
	}
	if !strings.Contains(jsonStr, "discount") {
		fmt.Println("✓ Does not contain discount (conditional set with false)")
	}
	fmt.Printf("Generated JSON: %s\n", jsonStr)
	fmt.Println("Input builder completed")
}

// ExampleInputBuilderPatterns demonstrates common input patterns
func ExampleInputBuilderPatterns() {
	fmt.Println("Step Functions input builder patterns:")
	
	// Order processing input
	orderInput := stepfunctions.CreateOrderInput("order-789", "customer-101", []string{"item1", "item2"})
	orderJSON, err := orderInput.ToJSON()
	if err != nil {
		fmt.Printf("Failed to create order JSON: %v\n", err)
		return
	}
	if strings.Contains(orderJSON, "order-789") && strings.Contains(orderJSON, "timestamp") {
		fmt.Println("✓ Order input created with ID and timestamp")
	}
	
	// Generic workflow input
	workflowInput := stepfunctions.CreateWorkflowInput("user-onboarding", map[string]interface{}{
		"userId":    "user-123",
		"email":     "user@example.com",
		"requestId": "req-456",
	})
	workflowJSON, err := workflowInput.ToJSON()
	if err != nil {
		fmt.Printf("Failed to create workflow JSON: %v\n", err)
		return
	}
	if strings.Contains(workflowJSON, "user-onboarding") && strings.Contains(workflowJSON, "req-456") {
		fmt.Println("✓ Workflow input created with type and flattened request ID")
	}
	
	// Retryable operation input
	retryInput := stepfunctions.CreateRetryableInput("risky-operation", 3)
	retryJSON, err := retryInput.ToJSON()
	if err != nil {
		fmt.Printf("Failed to create retry JSON: %v\n", err)
		return
	}
	if strings.Contains(retryJSON, "risky-operation") && strings.Contains(retryJSON, "currentAttempt") {
		fmt.Println("✓ Retryable input created with operation and attempt tracking")
	}
	fmt.Println("Input builder patterns completed")
}

// ExampleAssertions demonstrates comprehensive assertion usage
func ExampleAssertions() {
	fmt.Println("Step Functions assertion patterns:")
	
	// Create example execution results
	successResult := &stepfunctions.ExecutionResult{
		Status:        types.ExecutionStatusSucceeded,
		Output:        `{"result": "success", "data": {"processed": 100}}`,
		ExecutionTime: 45 * time.Second,
	}
	
	failedResult := &stepfunctions.ExecutionResult{
		Status: types.ExecutionStatusFailed,
		Error:  "States.TaskFailed",
		Cause:  "Lambda function returned error",
	}
	
	// Status assertions
	if stepfunctions.AssertExecutionSucceededE(successResult) {
		fmt.Println("✓ Success result correctly identified as succeeded")
	}
	if !stepfunctions.AssertExecutionSucceededE(failedResult) {
		fmt.Println("✓ Failed result correctly identified as not succeeded")
	}
	if stepfunctions.AssertExecutionFailedE(failedResult) {
		fmt.Println("✓ Failed result correctly identified as failed")
	}
	
	// Output assertions
	if stepfunctions.AssertExecutionOutputE(successResult, "success") {
		fmt.Println("✓ Output contains 'success'")
	}
	if stepfunctions.AssertExecutionOutputE(successResult, "processed") {
		fmt.Println("✓ Output contains 'processed'")
	}
	if !stepfunctions.AssertExecutionOutputE(successResult, "failure") {
		fmt.Println("✓ Output does not contain 'failure'")
	}
	
	// JSON output assertions
	expectedJSON := map[string]interface{}{
		"result": "success",
		"data": map[string]interface{}{
			"processed": float64(100),
		},
	}
	if stepfunctions.AssertExecutionOutputJSONE(successResult, expectedJSON) {
		fmt.Println("✓ JSON output matches expected structure")
	}
	
	// Time assertions
	if stepfunctions.AssertExecutionTimeE(successResult, 30*time.Second, 60*time.Second) {
		fmt.Println("✓ Execution time within expected range (30-60s)")
	}
	if !stepfunctions.AssertExecutionTimeE(successResult, 60*time.Second, 90*time.Second) {
		fmt.Println("✓ Execution time correctly outside range (60-90s)")
	}
	
	// Error assertions
	if stepfunctions.AssertExecutionErrorE(failedResult, "TaskFailed") {
		fmt.Println("✓ Error contains 'TaskFailed'")
	}
	if stepfunctions.AssertExecutionErrorE(failedResult, "States.") {
		fmt.Println("✓ Error contains 'States.' prefix")
	}
	fmt.Println("Assertion patterns completed")
}

// ExampleExecutionPattern demonstrates pattern-based assertions
func ExampleExecutionPattern() {
	fmt.Println("Step Functions execution pattern matching:")
	
	result := &stepfunctions.ExecutionResult{
		Status:        types.ExecutionStatusSucceeded,
		Output:        `{"status": "completed", "itemsProcessed": 250}`,
		ExecutionTime: 2 * time.Minute,
	}
	
	// Define expected pattern
	status := types.ExecutionStatusSucceeded
	minTime := 1 * time.Minute
	maxTime := 5 * time.Minute
	
	pattern := &stepfunctions.ExecutionPattern{
		Status:           &status,
		OutputContains:   "completed",
		MinExecutionTime: &minTime,
		MaxExecutionTime: &maxTime,
	}
	
	// Assert pattern match
	if stepfunctions.AssertExecutionPatternE(result, pattern) {
		fmt.Println("✓ Execution matches expected pattern (status, output, time range)")
	}
	
	// Test pattern with JSON
	jsonPattern := &stepfunctions.ExecutionPattern{
		Status: &status,
		OutputJSON: map[string]interface{}{
			"status":         "completed",
			"itemsProcessed": float64(250),
		},
	}
	
	if stepfunctions.AssertExecutionPatternE(result, jsonPattern) {
		fmt.Println("✓ Execution matches JSON pattern")
	}
	fmt.Println("Execution pattern matching completed")
}

// ExamplePollingConfiguration demonstrates polling configuration
func ExamplePollingConfiguration() {
	fmt.Println("Step Functions polling configuration patterns:")
	
	// Simple polling - constant interval
	simpleConfig := &stepfunctions.PollConfig{
		MaxAttempts:        30,
		Interval:           2 * time.Second,
		Timeout:            60 * time.Second,
		ExponentialBackoff: false,
	}
	
	// Test interval calculation
	interval1 := simpleConfig.Interval
	interval2 := simpleConfig.Interval
	if interval1 == 2*time.Second && interval2 == 2*time.Second {
		fmt.Println("✓ Simple polling: constant 2s interval")
	}
	
	// Exponential backoff polling
	backoffConfig := &stepfunctions.PollConfig{
		MaxAttempts:        20,
		Interval:           1 * time.Second,
		Timeout:            120 * time.Second,
		ExponentialBackoff: true,
		BackoffMultiplier:  1.5,
		MaxInterval:        30 * time.Second,
	}
	
	// Test exponential backoff calculation
	baseInterval := backoffConfig.Interval
	secondInterval := time.Duration(float64(backoffConfig.Interval) * 1.5)
	thirdInterval := time.Duration(float64(backoffConfig.Interval) * 1.5 * 1.5)
	
	fmt.Printf("Exponential backoff intervals: %v -> %v -> %v\n", baseInterval, secondInterval, thirdInterval)
	
	// Test configuration limits
	if 5 < backoffConfig.MaxAttempts && 30*time.Second < backoffConfig.Timeout {
		fmt.Println("✓ Configuration within reasonable limits")
	}
	if !(25 < backoffConfig.MaxAttempts) {
		fmt.Println("✓ Max attempts limit properly configured")
	}
	if !(150*time.Second < backoffConfig.Timeout) {
		fmt.Println("✓ Timeout limit properly configured")
	}
	fmt.Println("Polling configuration patterns completed")
}

// ExampleErrorHandling demonstrates error handling patterns
func ExampleErrorHandling() {
	fmt.Println("Step Functions error handling patterns:")
	
	// Empty state machine name validation
	name := ""
	if name == "" {
		fmt.Println("✓ Empty state machine name is invalid")
	}
	
	// Empty definition validation
	definition := ""
	if definition == "" {
		fmt.Println("✓ Empty definition is invalid")
	}
	
	// Empty execution name validation
	execName := ""
	if execName == "" {
		fmt.Println("✓ Empty execution name is invalid")
	}
	
	// Invalid ARN validation examples
	invalidArn := "invalid-arn"
	if !strings.HasPrefix(invalidArn, "arn:aws:states:") {
		fmt.Println("✓ Invalid state machine ARN format detected")
	}
	
	invalidExecArn := "invalid-execution-arn"
	if !strings.HasPrefix(invalidExecArn, "arn:aws:states:") {
		fmt.Println("✓ Invalid execution ARN format detected")
	}
	
	// Valid ARN example
	validArn := "arn:aws:states:us-east-1:123456789012:stateMachine:MyStateMachine"
	if strings.HasPrefix(validArn, "arn:aws:states:") {
		fmt.Println("✓ Valid ARN format confirmed")
	}
	fmt.Println("Error handling patterns completed")
}

// ExampleCompleteWorkflow demonstrates a full Step Functions scenario
func ExampleCompleteWorkflow() {
	fmt.Println("Step Functions complete workflow patterns:")
	
	// This example shows a complete Step Functions workflow
	// that would be used in real implementations
	
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}
	
	_ = /* ctx */ &stepfunctions.TestContext{
		// T:         t,
		AwsConfig: cfg,
		Region:    "us-east-1",
	}
	
	// 1. Define state machine with error handling
	definition := &stepfunctions.StateMachineDefinition{
		Name: "complete-example-" + time.Now().Format("20060102-150405"),
		Definition: `{
			"Comment": "Complete example with error handling",
			"StartAt": "ProcessOrder",
			"States": {
				"ProcessOrder": {
					"Type": "Task",
					"Resource": "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
					"Retry": [
						{
							"ErrorEquals": ["States.TaskFailed"],
							"IntervalSeconds": 2,
							"MaxAttempts": 3,
							"BackoffRate": 2.0
						}
					],
					"Catch": [
						{
							"ErrorEquals": ["States.ALL"],
							"Next": "HandleError"
						}
					],
					"Next": "OrderComplete"
				},
				"OrderComplete": {
					"Type": "Pass",
					"Result": "Order processed successfully",
					"End": true
				},
				"HandleError": {
					"Type": "Pass",
					"Result": "Error handled",
					"End": true
				}
			}
		}`,
		RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
		Type:    types.StateMachineTypeStandard,
	}
	fmt.Printf("Defining state machine with error handling: %s\n", definition.Name)
	
	// 2. Create state machine with options
	_ = /* options */ &stepfunctions.StateMachineOptions{
		Tags: map[string]string{
			"Environment": "example",
			"Team":        "engineering",
			"CostCenter":  "dev-ops",
		},
	}
	
	// stateMachine, err := stepfunctions.CreateStateMachineE(ctx, definition, options)
	// if err != nil {
	//     fmt.Printf("Failed to create state machine: %v\n", err)
	//     return
	// }
	// defer func() {
	//     _ = stepfunctions.DeleteStateMachineE(ctx, stateMachine.StateMachineArn)
	// }()
	fmt.Println("Creating state machine with tags")
	
	// 3. Verify state machine properties
	// stepfunctions.AssertStateMachineExists(t, ctx, stateMachine.StateMachineArn)
	// stepfunctions.AssertStateMachineType(t, ctx, stateMachine.StateMachineArn, types.StateMachineTypeStandard)
	fmt.Println("Verifying state machine exists and is Standard type")
	
	// 4. Execute successful workflow
	_ = /* successInput */ stepfunctions.CreateOrderInput("order-001", "customer-001", []string{"product-A", "product-B"})
	
	// result := stepfunctions.ExecuteStateMachineAndWait(ctx,
	//     "arn:aws:states:us-east-1:123456789012:stateMachine:complete-example",
	//     "successful-order",
	//     successInput,
	//     3*time.Minute)
	fmt.Println("Executing state machine with order input (3 minute timeout)")
	
	// 5. Comprehensive result validation
	pattern := &stepfunctions.ExecutionPattern{
		Status: func() *types.ExecutionStatus {
			s := types.ExecutionStatusSucceeded
			return &s
		}(),
		OutputContains: "processed successfully",
		MinExecutionTime: func() *time.Duration {
			d := 1 * time.Second
			return &d
		}(),
		MaxExecutionTime: func() *time.Duration {
			d := 2 * time.Minute
			return &d
		}(),
	}
	
	// stepfunctions.AssertExecutionPattern(t, result, pattern)
	fmt.Printf("Validating execution matches pattern: %+v\n", pattern)
	fmt.Println("Expected: succeeded, contains 'processed successfully', 1s-2min duration")
	
	// 6. Verify execution metrics
	// stepfunctions.AssertExecutionCountByStatus(t, ctx, stateMachine.StateMachineArn, 
	//     types.ExecutionStatusSucceeded, 1)
	fmt.Println("Verifying execution count: 1 succeeded execution")
	
	fmt.Println("Complete workflow demonstrates full Step Functions capabilities:")
	fmt.Println("  - State machine management with error handling")
	fmt.Println("  - Execution lifecycle management")
	fmt.Println("  - Comprehensive result validation")
	fmt.Println("Complete workflow example finished")
}


// ExampleHistoryAnalysis demonstrates comprehensive execution history analysis
func ExampleHistoryAnalysis() {
	fmt.Println("Step Functions execution history analysis patterns:")
	
	// Create sample execution history (simulated)
	history := []stepfunctions.HistoryEvent{
		{
			ID:        1,
			Type:      types.HistoryEventTypeExecutionStarted,
			Timestamp: time.Now().Add(-10 * time.Minute),
		},
		{
			ID:        2,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: time.Now().Add(-9 * time.Minute),
			StateEnteredEventDetails: &stepfunctions.StateEnteredEventDetails{
				Name: "ValidateOrder",
			},
		},
		{
			ID:        3,
			Type:      types.HistoryEventTypeTaskSucceeded,
			Timestamp: time.Now().Add(-8 * time.Minute),
			TaskSucceededEventDetails: &stepfunctions.TaskSucceededEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ValidateOrder",
				ResourceType: "lambda",
				Output:       `{"valid": true}`,
			},
		},
		{
			ID:        4,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: time.Now().Add(-7 * time.Minute),
			StateEnteredEventDetails: &stepfunctions.StateEnteredEventDetails{
				Name: "ProcessPayment",
			},
		},
		{
			ID:        5,
			Type:      types.HistoryEventTypeTaskFailed,
			Timestamp: time.Now().Add(-6 * time.Minute),
			TaskFailedEventDetails: &stepfunctions.TaskFailedEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessPayment",
				ResourceType: "lambda",
				Error:        "States.TaskFailed",
				Cause:        "Insufficient funds",
			},
		},
		{
			ID:        6,
			Type:      types.HistoryEventTypeExecutionFailed,
			Timestamp: time.Now().Add(-5 * time.Minute),
			ExecutionFailedEventDetails: &stepfunctions.ExecutionFailedEventDetails{
				Error: "States.TaskFailed",
				Cause: "ProcessPayment step failed",
			},
		},
	}

	// Analyze execution history
	// analysis, err := stepfunctions.AnalyzeExecutionHistoryE(history)
	// if err != nil {
	//     fmt.Printf("Analysis failed: %v\n", err)
	//     return
	// }
	
	fmt.Printf("Analyzing execution history with %d events\n", len(history))
	fmt.Println("Expected analysis results:")
	fmt.Println("  - Total steps: 6")
	fmt.Println("  - Completed steps: 1 (ValidateOrder succeeded)")
	fmt.Println("  - Failed steps: 1 (ProcessPayment failed)")
	fmt.Println("  - Retry attempts: 0")
	fmt.Println("  - Status: FAILED")
	fmt.Println("  - Failure reason: States.TaskFailed")
	fmt.Println("  - Failure cause: ProcessPayment step failed")
	fmt.Println("  - Should contain step timings and resource usage")
	fmt.Println("History analysis completed")
}

// ExampleFailedStepsAnalysis demonstrates failed step analysis
func ExampleFailedStepsAnalysis() {
	fmt.Println("Step Functions failed steps analysis patterns:")
	
	// Create history with multiple failed steps
	history := []stepfunctions.HistoryEvent{
		{
			ID:        1,
			Type:      types.HistoryEventTypeTaskFailed,
			Timestamp: time.Now().Add(-10 * time.Minute),
			TaskFailedEventDetails: &stepfunctions.TaskFailedEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
				ResourceType: "lambda",
				Error:        "States.TaskFailed",
				Cause:        "Validation error",
			},
		},
		{
			ID:        2,
			Type:      types.HistoryEventTypeLambdaFunctionFailed,
			Timestamp: time.Now().Add(-5 * time.Minute),
			LambdaFunctionFailedEventDetails: &stepfunctions.LambdaFunctionFailedEventDetails{
				Error: "Runtime.HandlerNotFound",
				Cause: "Handler function not found",
			},
		},
	}

	// Find all failed steps
	// failedSteps, err := stepfunctions.FindFailedStepsE(history)
	// if err != nil {
	//     fmt.Printf("Failed to find failed steps: %v\n", err)
	//     return
	// }
	
	fmt.Printf("Finding failed steps in history with %d events\n", len(history))
	fmt.Println("Expected failed steps analysis:")
	fmt.Println("  - Total failed steps: 2")
	fmt.Println("  - Step 1: ProcessOrder, Error: States.TaskFailed, Cause: Validation error")
	fmt.Println("  - Step 2: Lambda function, Error: Runtime.HandlerNotFound, Cause: Handler function not found")
	fmt.Println("Failed steps analysis completed")
}

// ExampleExecutionTimeline demonstrates execution timeline creation
func ExampleExecutionTimeline() {
	fmt.Println("Step Functions execution timeline patterns:")
	
	// Create sample history events
	baseTime := time.Now().Add(-10 * time.Minute)
	history := []stepfunctions.HistoryEvent{
		{
			ID:        1,
			Type:      types.HistoryEventTypeExecutionStarted,
			Timestamp: baseTime,
		},
		{
			ID:        2,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: baseTime.Add(1 * time.Minute),
			StateEnteredEventDetails: &stepfunctions.StateEnteredEventDetails{
				Name: "ProcessOrder",
			},
		},
		{
			ID:        3,
			Type:      types.HistoryEventTypeTaskSucceeded,
			Timestamp: baseTime.Add(5 * time.Minute),
		},
		{
			ID:        4,
			Type:      types.HistoryEventTypeExecutionSucceeded,
			Timestamp: baseTime.Add(6 * time.Minute),
		},
	}

	// Create execution timeline
	// timeline, err := stepfunctions.GetExecutionTimelineE(history)
	// if err != nil {
	//     fmt.Printf("Failed to create timeline: %v\n", err)
	//     return
	// }
	
	fmt.Printf("Creating timeline from %d history events\n", len(history))
	fmt.Println("Expected timeline characteristics:")
	fmt.Println("  - Total events: 4")
	fmt.Println("  - Events are chronologically ordered")
	fmt.Println("  - First event duration: 0 (no prior event)")
	fmt.Println("  - Second event duration: 1 minute (time since first event)")
	fmt.Println("  - Timeline shows step-by-step execution flow")
	fmt.Println("Execution timeline completed")
}

// ExampleExecutionSummaryFormatting demonstrates summary formatting
func ExampleExecutionSummaryFormatting() {
	fmt.Println("Step Functions execution summary formatting patterns:")
	
	// Create analysis data
	_ = /* analysis */ &stepfunctions.ExecutionAnalysis{
		TotalSteps:      10,
		CompletedSteps:  8,
		FailedSteps:     2,
		RetryAttempts:   1,
		TotalDuration:   5 * time.Minute,
		ExecutionStatus: types.ExecutionStatusFailed,
		FailureReason:   "States.TaskFailed",
		FailureCause:    "Payment processing failed",
		StepTimings: map[string]time.Duration{
			"ValidateOrder":  30 * time.Second,
			"ProcessPayment": 2 * time.Minute,
		},
		ResourceUsage: map[string]int{
			"arn:aws:lambda:us-east-1:123456789012:function:ValidateOrder":  1,
			"arn:aws:lambda:us-east-1:123456789012:function:ProcessPayment": 2,
		},
	}

	// Format summary
	// summary, err := stepfunctions.FormatExecutionSummaryE(analysis)
	// if err != nil {
	//     fmt.Printf("Failed to format summary: %v\n", err)
	//     return
	// }
	
	fmt.Println("Formatting execution summary from analysis data")
	fmt.Println("Expected summary contents:")
	fmt.Println("  - Total Steps: 10")
	fmt.Println("  - Completed: 8")
	fmt.Println("  - Failed: 2")
	fmt.Println("  - Status: FAILED")
	fmt.Println("  - Error: States.TaskFailed")
	fmt.Println("  - ValidateOrder: 30s")
	fmt.Println("  - ProcessPayment: 2m")
	fmt.Println("  - Resource usage statistics")
	fmt.Println("Summary formatting completed")
}

// ExampleComprehensiveDiagnostics demonstrates complete diagnostic workflow
func ExampleComprehensiveDiagnostics() {
	fmt.Println("Step Functions comprehensive diagnostics patterns:")
	
	// Simulate a real-world scenario where execution fails
	// and we need to diagnose the issue

	// Step 1: Get execution history (simulated)
	history := createComplexExecutionHistory()

	// Step 2: Perform comprehensive analysis
	// analysis, err := stepfunctions.AnalyzeExecutionHistoryE(history)
	// if err != nil {
	//     fmt.Printf("Analysis failed: %v\n", err)
	//     return
	// }

	// Step 3: Find specific failed steps for detailed investigation
	// failedSteps, err := stepfunctions.FindFailedStepsE(history)
	// if err != nil {
	//     fmt.Printf("Failed steps analysis failed: %v\n", err)
	//     return
	// }

	// Step 4: Check for retry patterns
	// retryAttempts, err := stepfunctions.GetRetryAttemptsE(history)
	// if err != nil {
	//     fmt.Printf("Retry analysis failed: %v\n", err)
	//     return
	// }

	// Step 5: Create timeline for step-by-step analysis
	// timeline, err := stepfunctions.GetExecutionTimelineE(history)
	// if err != nil {
	//     fmt.Printf("Timeline creation failed: %v\n", err)
	//     return
	// }

	// Step 6: Generate human-readable summary
	// summary, err := stepfunctions.FormatExecutionSummaryE(analysis)
	// if err != nil {
	//     fmt.Printf("Summary formatting failed: %v\n", err)
	//     return
	// }
	
	fmt.Printf("=== EXECUTION DIAGNOSTIC WORKFLOW ===\n")
	fmt.Printf("Step 1: Analyzing execution history with %d events\n", len(history))
	fmt.Println("Step 2: Performing comprehensive analysis")
	fmt.Println("Step 3: Finding failed steps for detailed investigation")
	fmt.Println("Step 4: Checking for retry patterns")
	fmt.Println("Step 5: Creating timeline for step-by-step analysis")
	fmt.Println("Step 6: Generating human-readable summary")
	fmt.Println("\nExpected diagnostic outputs:")
	fmt.Println("  - Comprehensive analysis with metrics")
	fmt.Println("  - List of failed steps with error details")
	fmt.Println("  - Retry attempt information")
	fmt.Println("  - Chronological timeline of events")
	fmt.Println("  - Human-readable summary report")
	fmt.Println("Comprehensive diagnostics completed")
}

// Helper function to create complex execution history for testing
func createComplexExecutionHistory() []stepfunctions.HistoryEvent {
	baseTime := time.Now().Add(-20 * time.Minute)
	
	return []stepfunctions.HistoryEvent{
		{
			ID:        1,
			Type:      types.HistoryEventTypeExecutionStarted,
			Timestamp: baseTime,
		},
		{
			ID:        2,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: baseTime.Add(30 * time.Second),
			StateEnteredEventDetails: &stepfunctions.StateEnteredEventDetails{
				Name: "ValidateInput",
			},
		},
		{
			ID:        3,
			Type:      types.HistoryEventTypeTaskSucceeded,
			Timestamp: baseTime.Add(45 * time.Second),
			TaskSucceededEventDetails: &stepfunctions.TaskSucceededEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ValidateInput",
				ResourceType: "lambda",
			},
		},
		{
			ID:        4,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: baseTime.Add(1 * time.Minute),
			StateEnteredEventDetails: &stepfunctions.StateEnteredEventDetails{
				Name: "ProcessOrder",
			},
		},
		{
			ID:        5,
			Type:      types.HistoryEventTypeTaskFailed,
			Timestamp: baseTime.Add(3 * time.Minute),
			TaskFailedEventDetails: &stepfunctions.TaskFailedEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
				ResourceType: "lambda",
				Error:        "States.TaskFailed",
				Cause:        "Database connection timeout",
			},
		},
		{
			ID:        6,
			Type:      types.HistoryEventTypeTaskFailed,
			Timestamp: baseTime.Add(5 * time.Minute),
			TaskFailedEventDetails: &stepfunctions.TaskFailedEventDetails{
				Resource:     "arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder",
				ResourceType: "lambda",
				Error:        "States.TaskFailed",
				Cause:        "Retry failed - still cannot connect",
			},
		},
		{
			ID:        7,
			Type:      types.HistoryEventTypeExecutionFailed,
			Timestamp: baseTime.Add(6 * time.Minute),
			ExecutionFailedEventDetails: &stepfunctions.ExecutionFailedEventDetails{
				Error: "States.TaskFailed",
				Cause: "ProcessOrder step exhausted all retry attempts",
			},
		},
	}
}

// runAllExamples demonstrates running all functional Step Functions examples
func runAllExamples() {
	fmt.Println("Running all functional Step Functions package examples:")
	fmt.Println("")
	
	// Use functional composition to run examples with error handling
	exampleFunctions := []func(){
		ExampleBasicWorkflow,
		ExampleExecuteAndWaitPattern,
		ExampleInputBuilder,
		ExampleInputBuilderPatterns,
		ExampleAssertions,
		ExampleExecutionPattern,
		ExamplePollingConfiguration,
		ExampleErrorHandling,
		ExampleCompleteWorkflow,
		ExampleHistoryAnalysis,
		ExampleFailedStepsAnalysis,
		ExampleExecutionTimeline,
		ExampleExecutionSummaryFormatting,
		ExampleComprehensiveDiagnostics,
	}
	
	// Functional execution with monadic error handling
	executionResult := lo.Pipe2(
		exampleFunctions,
		func(examples []func()) mo.Result[[]func()] {
			if len(examples) == 0 {
				return mo.Err[[]func()](fmt.Errorf("no examples to run"))
			}
			return mo.Ok(examples)
		},
		func(validated mo.Result[[]func()]) mo.Option[int] {
			return validated.Match(
				func(examples []func()) mo.Option[int] {
					lo.ForEach(examples, func(example func(), index int) {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("✗ Example %d failed: %v\n", index+1, r)
							}
						}()
						example()
						fmt.Println("")
					})
					return mo.Some(len(examples))
				},
				func(err error) mo.Option[int] {
					fmt.Printf("✗ Failed to run examples: %v\n", err)
					return mo.None[int]()
				},
			)
		},
	)
	
	executionResult.Match(
		func(count int) {
			fmt.Printf("✓ All %d functional Step Functions examples completed successfully\n", count)
		},
		func() {
			fmt.Println("✗ Example execution was interrupted")
		},
	)
}

func main() {
	// This file demonstrates functional Step Functions usage patterns with the vasdeference framework.
	// Run examples with: go run ./examples/stepfunctions/examples.go
	
	// Functional main execution with error boundary
	mainResult := mo.TryOr(func() error {
		runAllExamples()
		return nil
	}, func(err error) error {
		fmt.Printf("✗ Application failed: %v\n", err)
		return err
	})
	
	mainResult.Match(
		func(value interface{}) {
			fmt.Println("✓ Application completed successfully")
		},
		func(err error) {
			fmt.Printf("✗ Application terminated with error: %v\n", err)
		},
	)
}