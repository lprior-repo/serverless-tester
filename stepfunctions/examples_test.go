package stepfunctions

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example tests demonstrating comprehensive Step Functions testing patterns
// These examples show how to use the vasdeference Step Functions testing package

// TestStepFunctionsBasicWorkflow demonstrates a complete workflow test
func TestStepFunctionsBasicWorkflow(t *testing.T) {
	// Skip in CI/automated testing - this requires real AWS resources
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Setup AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	require.NoError(t, err, "Failed to load AWS config")
	
	ctx := &TestContext{
		T:         t,
		AwsConfig: cfg,
		Region:    "us-east-1",
	}
	
	// Define a simple state machine
	definition := &StateMachineDefinition{
		Name: "test-workflow-" + time.Now().Format("20060102-150405"),
		Definition: `{
			"Comment": "A Hello World example",
			"StartAt": "HelloWorld",
			"States": {
				"HelloWorld": {
					"Type": "Pass",
					"Result": "Hello World!",
					"End": true
				}
			}
		}`,
		RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
		Type:    types.StateMachineTypeStandard,
		Tags: map[string]string{
			"Environment": "test",
			"Purpose":     "example",
		},
	}
	
	// Step 1: Create the state machine
	stateMachine, err := CreateStateMachineE(ctx, definition, nil)
	require.NoError(t, err, "Failed to create state machine")
	assert.Equal(t, definition.Name, stateMachine.Name)
	
	// Cleanup function
	defer func() {
		_ = DeleteStateMachineE(ctx, stateMachine.StateMachineArn)
	}()
	
	// Step 2: Start an execution
	input := NewInput().
		Set("greeting", "Hello").
		Set("target", "World").
		Set("timestamp", time.Now().Unix())
	
	execution, err := StartExecutionE(ctx, stateMachine.StateMachineArn, "test-execution", input)
	require.NoError(t, err, "Failed to start execution")
	assert.NotEmpty(t, execution.ExecutionArn)
	
	// Step 3: Wait for completion
	finalResult, err := WaitForExecutionE(ctx, execution.ExecutionArn, &WaitOptions{
		Timeout:         2 * time.Minute,
		PollingInterval: 5 * time.Second,
		MaxAttempts:     24,
	})
	require.NoError(t, err, "Execution failed to complete")
	
	// Step 4: Assert execution succeeded
	AssertExecutionSucceeded(t, finalResult)
	AssertExecutionOutput(t, finalResult, "Hello World!")
	AssertExecutionTime(t, finalResult, 0, 30*time.Second)
}

// TestStepFunctionsExecuteAndWaitPattern demonstrates the execute-and-wait pattern
func TestStepFunctionsExecuteAndWaitPattern(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	ctx := &TestContext{T: t, AwsConfig: cfg, Region: "us-east-1"}
	
	// This pattern combines execution start and wait in one call
	input := NewInput().Set("message", "Execute and wait pattern")
	
	result, err := ExecuteStateMachineAndWaitE(ctx,
		"arn:aws:states:us-east-1:123456789012:stateMachine:my-state-machine",
		"execute-and-wait-test",
		input,
		5*time.Minute)
	
	require.NoError(t, err)
	AssertExecutionSucceeded(t, result)
}

// TestStepFunctionsInputBuilder demonstrates fluent input building
func TestStepFunctionsInputBuilder(t *testing.T) {
	// Create complex input using fluent API
	input := NewInput().
		Set("orderId", "order-123").
		Set("customerId", "customer-456").
		Set("amount", 99.99).
		Set("currency", "USD").
		SetIf(true, "priority", "high").
		SetIf(false, "discount", 10.0) // This won't be set
	
	// Convert to JSON
	jsonStr, err := input.ToJSON()
	require.NoError(t, err)
	
	// Verify content
	assert.Contains(t, jsonStr, "order-123")
	assert.Contains(t, jsonStr, "priority")
	assert.NotContains(t, jsonStr, "discount")
}

// TestStepFunctionsInputBuilderPatterns demonstrates common input patterns
func TestStepFunctionsInputBuilderPatterns(t *testing.T) {
	// Order processing input
	orderInput := CreateOrderInput("order-789", "customer-101", []string{"item1", "item2"})
	orderJSON, err := orderInput.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, orderJSON, "order-789")
	assert.Contains(t, orderJSON, "timestamp")
	
	// Generic workflow input
	workflowInput := CreateWorkflowInput("user-onboarding", map[string]interface{}{
		"userId":      "user-123",
		"email":       "user@example.com",
		"requestId":   "req-456",
	})
	workflowJSON, err := workflowInput.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, workflowJSON, "user-onboarding")
	assert.Contains(t, workflowJSON, "req-456") // Should be flattened to top level
	
	// Retryable operation input
	retryInput := CreateRetryableInput("risky-operation", 3)
	retryJSON, err := retryInput.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, retryJSON, "risky-operation")
	assert.Contains(t, retryJSON, "currentAttempt")
}

// TestStepFunctionsAssertions demonstrates comprehensive assertion usage
func TestStepFunctionsAssertions(t *testing.T) {
	// Create test execution results
	successResult := &ExecutionResult{
		Status:        types.ExecutionStatusSucceeded,
		Output:        `{"result": "success", "data": {"processed": 100}}`,
		ExecutionTime: 45 * time.Second,
	}
	
	failedResult := &ExecutionResult{
		Status: types.ExecutionStatusFailed,
		Error:  "States.TaskFailed",
		Cause:  "Lambda function returned error",
	}
	
	// Status assertions
	assert.True(t, AssertExecutionSucceededE(successResult))
	assert.False(t, AssertExecutionSucceededE(failedResult))
	assert.True(t, AssertExecutionFailedE(failedResult))
	
	// Output assertions
	assert.True(t, AssertExecutionOutputE(successResult, "success"))
	assert.True(t, AssertExecutionOutputE(successResult, "processed"))
	assert.False(t, AssertExecutionOutputE(successResult, "failure"))
	
	// JSON output assertions
	expectedJSON := map[string]interface{}{
		"result": "success",
		"data": map[string]interface{}{
			"processed": float64(100),
		},
	}
	assert.True(t, AssertExecutionOutputJSONE(successResult, expectedJSON))
	
	// Time assertions
	assert.True(t, AssertExecutionTimeE(successResult, 30*time.Second, 60*time.Second))
	assert.False(t, AssertExecutionTimeE(successResult, 60*time.Second, 90*time.Second))
	
	// Error assertions
	assert.True(t, AssertExecutionErrorE(failedResult, "TaskFailed"))
	assert.True(t, AssertExecutionErrorE(failedResult, "States."))
}

// TestStepFunctionsExecutionPattern demonstrates pattern-based assertions
func TestStepFunctionsExecutionPattern(t *testing.T) {
	result := &ExecutionResult{
		Status:        types.ExecutionStatusSucceeded,
		Output:        `{"status": "completed", "itemsProcessed": 250}`,
		ExecutionTime: 2 * time.Minute,
	}
	
	// Define expected pattern
	status := types.ExecutionStatusSucceeded
	minTime := 1 * time.Minute
	maxTime := 5 * time.Minute
	
	pattern := &ExecutionPattern{
		Status:           &status,
		OutputContains:   "completed",
		MinExecutionTime: &minTime,
		MaxExecutionTime: &maxTime,
	}
	
	// Assert pattern match
	assert.True(t, AssertExecutionPatternE(result, pattern))
	
	// Test pattern with JSON
	jsonPattern := &ExecutionPattern{
		Status: &status,
		OutputJSON: map[string]interface{}{
			"status":         "completed",
			"itemsProcessed": float64(250),
		},
	}
	
	assert.True(t, AssertExecutionPatternE(result, jsonPattern))
}

// TestStepFunctionsPollingConfiguration demonstrates polling configuration
func TestStepFunctionsPollingConfiguration(t *testing.T) {
	// Test different polling strategies
	
	// Simple polling - constant interval
	simpleConfig := &PollConfig{
		MaxAttempts:        30,
		Interval:           2 * time.Second,
		Timeout:            60 * time.Second,
		ExponentialBackoff: false,
	}
	
	// Test interval calculation
	interval1 := calculateNextPollInterval(1, simpleConfig)
	interval2 := calculateNextPollInterval(5, simpleConfig)
	assert.Equal(t, 2*time.Second, interval1)
	assert.Equal(t, 2*time.Second, interval2) // Should be same without backoff
	
	// Exponential backoff polling
	backoffConfig := &PollConfig{
		MaxAttempts:        20,
		Interval:           1 * time.Second,
		Timeout:            120 * time.Second,
		ExponentialBackoff: true,
		BackoffMultiplier:  1.5,
		MaxInterval:        30 * time.Second,
	}
	
	// Test exponential backoff
	interval1 = calculateNextPollInterval(1, backoffConfig)
	interval2 = calculateNextPollInterval(2, backoffConfig)
	interval3 := calculateNextPollInterval(3, backoffConfig)
	
	assert.Equal(t, 1*time.Second, interval1)                    // First attempt
	assert.Equal(t, 1*time.Second*150/100, interval2)           // 1 * 1.5
	assert.Equal(t, 1*time.Second*150*150/10000, interval3)     // 1 * 1.5^2
	
	// Test continuation logic
	assert.True(t, shouldContinuePolling(5, 30*time.Second, backoffConfig))
	assert.False(t, shouldContinuePolling(25, 30*time.Second, backoffConfig)) // Exceeded max attempts
	assert.False(t, shouldContinuePolling(5, 150*time.Second, backoffConfig)) // Exceeded timeout
}

// TestStepFunctionsErrorHandling demonstrates error handling patterns
func TestStepFunctionsErrorHandling(t *testing.T) {
	// Test validation errors
	err := validateStateMachineName("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid state machine name")
	
	err = validateStateMachineDefinition("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty definition")
	
	err = validateExecutionName("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid execution name")
	
	// Test ARN validation
	err = validateStateMachineArn("invalid-arn")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ARN")
	
	err = validateExecutionArn("invalid-arn")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ARN")
}

// TestStepFunctionsCompleteExample demonstrates a full testing scenario
func TestStepFunctionsCompleteExample(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive integration test in short mode")
	}
	
	// This example shows a complete Step Functions testing workflow
	// that would be used in real integration tests
	
	cfg, err := config.LoadDefaultConfig(context.TODO())
	require.NoError(t, err)
	
	ctx := &TestContext{
		T:         t,
		AwsConfig: cfg,
		Region:    "us-east-1",
	}
	
	// 1. Define state machine with error handling
	definition := &StateMachineDefinition{
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
	
	// 2. Create state machine with options
	options := &StateMachineOptions{
		Tags: map[string]string{
			"Environment": "test",
			"Team":        "engineering",
			"CostCenter":  "dev-ops",
		},
	}
	
	stateMachine, err := CreateStateMachineE(ctx, definition, options)
	require.NoError(t, err)
	defer func() {
		_ = DeleteStateMachineE(ctx, stateMachine.StateMachineArn)
	}()
	
	// 3. Verify state machine properties
	AssertStateMachineExists(t, ctx, stateMachine.StateMachineArn)
	AssertStateMachineType(t, ctx, stateMachine.StateMachineArn, types.StateMachineTypeStandard)
	
	// 4. Test successful execution
	successInput := CreateOrderInput("order-001", "customer-001", []string{"product-A", "product-B"})
	
	result := ExecuteStateMachineAndWait(ctx,
		stateMachine.StateMachineArn,
		"successful-order",
		successInput,
		3*time.Minute)
	
	// 5. Comprehensive result validation
	pattern := &ExecutionPattern{
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
	
	AssertExecutionPattern(t, result, pattern)
	
	// 6. Test execution count
	AssertExecutionCountByStatus(t, ctx, stateMachine.StateMachineArn, 
		types.ExecutionStatusSucceeded, 1)
	
	// This example demonstrates the full power of the Step Functions testing package
	// combining state machine management, execution testing, and comprehensive assertions
}

// BenchmarkStepFunctionsOperations benchmarks core operations
func BenchmarkStepFunctionsOperations(b *testing.B) {
	// Benchmark input creation
	b.Run("InputCreation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			input := NewInput().
				Set("orderId", "order-123").
				Set("customerId", "customer-456").
				Set("items", []string{"item1", "item2"})
			_, _ = input.ToJSON()
		}
	})
	
	// Benchmark validation functions
	b.Run("Validation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateStateMachineName("my-state-machine")
			_ = validateExecutionName("my-execution")
			_ = validateStateMachineArn("arn:aws:states:us-east-1:123456789012:stateMachine:my-machine")
		}
	})
	
	// Benchmark assertions
	b.Run("Assertions", func(b *testing.B) {
		result := &ExecutionResult{
			Status:        types.ExecutionStatusSucceeded,
			Output:        `{"result": "success"}`,
			ExecutionTime: 30 * time.Second,
		}
		
		for i := 0; i < b.N; i++ {
			_ = AssertExecutionSucceededE(result)
			_ = AssertExecutionOutputE(result, "success")
			_ = AssertExecutionTimeE(result, 10*time.Second, 60*time.Second)
		}
	})
}