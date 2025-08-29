package stepfunctions

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/samber/lo"
)

// TestFunctionalStepFunctionsScenario tests a complete Step Functions scenario
func TestFunctionalStepFunctionsScenario(t *testing.T) {
	// Create Step Functions configuration
	config := NewFunctionalStepFunctionsConfig(
		WithStepFunctionsTimeout(10*time.Minute),
		WithStepFunctionsRetryConfig(3, 200*time.Millisecond),
		WithStepFunctionsPollingInterval(2*time.Second),
		WithStateMachineType(types.StateMachineTypeStandard),
		WithStepFunctionsMetadata("environment", "test"),
		WithStepFunctionsMetadata("scenario", "full_workflow"),
	)

	// Step 1: Create state machine
	definition := `{
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

	stateMachine := NewFunctionalStepFunctionsStateMachine(
		"test-state-machine", 
		definition, 
		"arn:aws:iam::123456789012:role/StepFunctionsRole",
	).WithType(types.StateMachineTypeStandard).WithTags(map[string]string{
		"Environment": "test",
		"Project":     "functional-testing",
	})

	createSMResult := SimulateFunctionalStepFunctionsCreateStateMachine(stateMachine, config)

	if !createSMResult.IsSuccess() {
		t.Errorf("Expected state machine creation to succeed, got error: %v", createSMResult.GetError().MustGet())
	}

	// Verify state machine result
	if createSMResult.GetResult().IsPresent() {
		result, ok := createSMResult.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected state machine result to be map[string]interface{}")
		} else {
			if result["name"] != "test-state-machine" {
				t.Errorf("Expected state machine name to be 'test-state-machine', got %v", result["name"])
			}
			if result["created"] != true {
				t.Error("Expected state machine to be created")
			}
		}
	}

	// Step 2: Start execution
	execution := NewFunctionalStepFunctionsExecution(
		"test-execution",
		"arn:aws:states:us-east-1:123456789012:stateMachine:test-state-machine",
		`{"message": "Hello, World!", "count": 5}`,
	)

	startExecResult := SimulateFunctionalStepFunctionsStartExecution(execution, config)

	if !startExecResult.IsSuccess() {
		t.Errorf("Expected execution start to succeed, got error: %v", startExecResult.GetError().MustGet())
	}

	// Verify execution result
	var executionArn string
	if startExecResult.GetResult().IsPresent() {
		result, ok := startExecResult.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected execution result to be map[string]interface{}")
		} else {
			if result["name"] != "test-execution" {
				t.Errorf("Expected execution name to be 'test-execution', got %v", result["name"])
			}
			if result["started"] != true {
				t.Error("Expected execution to be started")
			}
			executionArn, _ = result["execution_arn"].(string)
		}
	}

	// Step 3: Describe execution
	describeExecResult := SimulateFunctionalStepFunctionsDescribeExecution(executionArn, config)

	if !describeExecResult.IsSuccess() {
		t.Errorf("Expected execution describe to succeed, got error: %v", describeExecResult.GetError().MustGet())
	}

	// Verify describe result
	if describeExecResult.GetResult().IsPresent() {
		result, ok := describeExecResult.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected describe result to be map[string]interface{}")
		} else {
			if result["execution_arn"] != executionArn {
				t.Errorf("Expected execution ARN to match: %s vs %v", executionArn, result["execution_arn"])
			}
			if result["status"] == "" {
				t.Error("Expected execution to have status")
			}
		}
	}

	// Step 4: Create activity
	activity := NewFunctionalStepFunctionsActivity("test-activity")
	createActivityResult := SimulateFunctionalStepFunctionsCreateActivity(activity, config)

	if !createActivityResult.IsSuccess() {
		t.Errorf("Expected activity creation to succeed, got error: %v", createActivityResult.GetError().MustGet())
	}

	// Step 5: Verify metadata accumulation across operations
	allResults := []FunctionalStepFunctionsOperationResult{
		createSMResult,
		startExecResult,
		describeExecResult,
		createActivityResult,
	}

	for i, result := range allResults {
		metadata := result.GetMetadata()
		if metadata["environment"] != "test" {
			t.Errorf("Operation %d missing environment metadata, metadata: %+v", i, metadata)
		}
		if result.GetDuration() == 0 {
			t.Errorf("Operation %d should have recorded duration", i)
		}
	}
}

// TestFunctionalStepFunctionsErrorHandling tests comprehensive error handling
func TestFunctionalStepFunctionsErrorHandling(t *testing.T) {
	config := NewFunctionalStepFunctionsConfig()

	// Test invalid state machine name
	invalidSM := NewFunctionalStepFunctionsStateMachine("", "valid-definition", "arn:aws:iam::123456789012:role/Role")
	
	result := SimulateFunctionalStepFunctionsCreateStateMachine(invalidSM, config)

	if result.IsSuccess() {
		t.Error("Expected operation to fail with empty state machine name")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	// Test invalid definition (empty)
	invalidDefSM := NewFunctionalStepFunctionsStateMachine("valid-name", "", "arn:aws:iam::123456789012:role/Role")
	
	result2 := SimulateFunctionalStepFunctionsCreateStateMachine(invalidDefSM, config)

	if result2.IsSuccess() {
		t.Error("Expected operation to fail with empty definition")
	}

	// Test invalid role ARN
	invalidRoleSM := NewFunctionalStepFunctionsStateMachine("valid-name", `{"StartAt":"S1","States":{"S1":{"Type":"Pass","End":true}}}`, "invalid-role-arn")
	
	result3 := SimulateFunctionalStepFunctionsCreateStateMachine(invalidRoleSM, config)

	if result3.IsSuccess() {
		t.Error("Expected operation to fail with invalid role ARN")
	}

	// Test invalid execution name
	invalidExecution := NewFunctionalStepFunctionsExecution("", "arn:aws:states:us-east-1:123456789012:stateMachine:test", "{}")
	
	result4 := SimulateFunctionalStepFunctionsStartExecution(invalidExecution, config)

	if result4.IsSuccess() {
		t.Error("Expected operation to fail with empty execution name")
	}

	// Test invalid execution input (malformed JSON)
	invalidInputExecution := NewFunctionalStepFunctionsExecution("valid-name", "arn:aws:states:us-east-1:123456789012:stateMachine:test", `{"invalid": json}`)
	
	result5 := SimulateFunctionalStepFunctionsStartExecution(invalidInputExecution, config)

	if result5.IsSuccess() {
		t.Error("Expected operation to fail with invalid JSON input")
	}

	// Test empty execution ARN for describe
	result6 := SimulateFunctionalStepFunctionsDescribeExecution("", config)

	if result6.IsSuccess() {
		t.Error("Expected operation to fail with empty execution ARN")
	}

	// Verify error metadata consistency
	errorResults := []FunctionalStepFunctionsOperationResult{result, result2, result3, result4, result5, result6}
	for i, errorResult := range errorResults {
		metadata := errorResult.GetMetadata()
		if metadata["phase"] != "validation" {
			t.Errorf("Error result %d should have validation phase, got %v", i, metadata["phase"])
		}
	}
}

// TestFunctionalStepFunctionsConfigurationCombinations tests various config combinations
func TestFunctionalStepFunctionsConfigurationCombinations(t *testing.T) {
	// Test minimal configuration
	minimalConfig := NewFunctionalStepFunctionsConfig()
	
	if minimalConfig.timeout != FunctionalStepFunctionsDefaultTimeout {
		t.Error("Expected minimal config to use default timeout")
	}
	
	if minimalConfig.pollingInterval != FunctionalStepFunctionsDefaultPollingInterval {
		t.Error("Expected minimal config to use default polling interval")
	}

	// Test maximal configuration
	maximalConfig := NewFunctionalStepFunctionsConfig(
		WithStepFunctionsTimeout(30*time.Minute),
		WithStepFunctionsPollingInterval(10*time.Second),
		WithStepFunctionsRetryConfig(10, 5*time.Second),
		WithStateMachineType(types.StateMachineTypeExpress),
		WithLoggingConfiguration(types.LoggingConfiguration{
			Level:                types.LogLevelAll,
			IncludeExecutionData: true,
			Destinations: []types.LogDestination{
				{
					CloudWatchLogsLogGroup: &types.CloudWatchLogsLogGroup{
						LogGroupArn: lo.ToPtr("arn:aws:logs:us-east-1:123456789012:log-group:test-log-group"),
					},
				},
			},
		}),
		WithTracingConfiguration(types.TracingConfiguration{
			Enabled: lo.ToPtr(true),
		}),
		WithStepFunctionsTags(map[string]string{
			"Environment": "production",
			"Team":        "backend",
		}),
		WithStepFunctionsMetadata("service", "step-functions-service"),
		WithStepFunctionsMetadata("environment", "production"),
		WithStepFunctionsMetadata("version", "v2.0.0"),
	)

	// Verify all options were applied
	if maximalConfig.timeout != 30*time.Minute {
		t.Error("Expected maximal config timeout")
	}

	if maximalConfig.pollingInterval != 10*time.Second {
		t.Error("Expected maximal config polling interval")
	}

	if maximalConfig.maxRetries != 10 {
		t.Error("Expected maximal config retries")
	}

	if maximalConfig.retryDelay != 5*time.Second {
		t.Error("Expected maximal config retry delay")
	}

	if maximalConfig.stateMachineType != types.StateMachineTypeExpress {
		t.Error("Expected maximal config state machine type")
	}

	if maximalConfig.loggingConfiguration.IsAbsent() {
		t.Error("Expected maximal config logging configuration")
	}

	if maximalConfig.tracingConfiguration.IsAbsent() {
		t.Error("Expected maximal config tracing configuration")
	}

	if len(maximalConfig.tags) != 2 {
		t.Errorf("Expected 2 tag entries, got %d", len(maximalConfig.tags))
	}

	if len(maximalConfig.metadata) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(maximalConfig.metadata))
	}
}

// TestFunctionalStepFunctionsStateMachineManipulation tests advanced state machine operations
func TestFunctionalStepFunctionsStateMachineManipulation(t *testing.T) {
	config := NewFunctionalStepFunctionsConfig(
		WithStepFunctionsMetadata("test_type", "state_machine_manipulation"),
	)

	// Create different types of state machines
	simpleDefinition := `{
		"Comment": "A simple pass state",
		"StartAt": "PassState",
		"States": {
			"PassState": {
				"Type": "Pass",
				"Result": "Hello World!",
				"End": true
			}
		}
	}`

	complexDefinition := `{
		"Comment": "A complex state machine with multiple state types",
		"StartAt": "ProcessInput",
		"States": {
			"ProcessInput": {
				"Type": "Task",
				"Resource": "arn:aws:lambda:us-east-1:123456789012:function:ProcessInput",
				"Next": "CheckCondition"
			},
			"CheckCondition": {
				"Type": "Choice",
				"Choices": [
					{
						"Variable": "$.shouldContinue",
						"BooleanEquals": true,
						"Next": "ContinueProcessing"
					}
				],
				"Default": "EndProcessing"
			},
			"ContinueProcessing": {
				"Type": "Parallel",
				"Branches": [
					{
						"StartAt": "ProcessA",
						"States": {
							"ProcessA": {
								"Type": "Task",
								"Resource": "arn:aws:lambda:us-east-1:123456789012:function:ProcessA",
								"End": true
							}
						}
					},
					{
						"StartAt": "ProcessB",
						"States": {
							"ProcessB": {
								"Type": "Task",
								"Resource": "arn:aws:lambda:us-east-1:123456789012:function:ProcessB",
								"End": true
							}
						}
					}
				],
				"End": true
			},
			"EndProcessing": {
				"Type": "Succeed"
			}
		}
	}`

	stateMachines := []FunctionalStepFunctionsStateMachine{
		NewFunctionalStepFunctionsStateMachine("simple-sm", simpleDefinition, "arn:aws:iam::123456789012:role/SimpleRole").
			WithType(types.StateMachineTypeExpress),
		
		NewFunctionalStepFunctionsStateMachine("complex-sm", complexDefinition, "arn:aws:iam::123456789012:role/ComplexRole").
			WithType(types.StateMachineTypeStandard).
			WithTags(map[string]string{"Complexity": "High", "Purpose": "DataProcessing"}),
	}

	// Test state machine properties
	for i, sm := range stateMachines {
		if sm.GetName() == "" {
			t.Errorf("State machine %d should have non-empty name", i)
		}

		if sm.GetDefinition() == "" {
			t.Errorf("State machine %d should have non-empty definition", i)
		}

		if sm.GetRoleArn() == "" {
			t.Errorf("State machine %d should have non-empty role ARN", i)
		}

		if sm.GetCreationDate().IsZero() {
			t.Errorf("State machine %d should have valid creation date", i)
		}
	}

	// Test state machine operations
	for i, sm := range stateMachines {
		result := SimulateFunctionalStepFunctionsCreateStateMachine(sm, config)

		if !result.IsSuccess() {
			t.Errorf("Expected state machine %d creation to succeed, got error: %v", i, result.GetError().MustGet())
		}

		// Verify result includes state machine information
		if result.GetResult().IsPresent() {
			resultMap, ok := result.GetResult().MustGet().(map[string]interface{})
			if ok {
				if resultMap["name"] != sm.GetName() {
					t.Errorf("Expected state machine %d name in result to match", i)
				}
				if resultMap["type"] != string(sm.GetType()) {
					t.Errorf("Expected state machine %d type in result to match", i)
				}
			}
		}
	}
}

// TestFunctionalStepFunctionsExecutionLifecycle tests execution lifecycle management
func TestFunctionalStepFunctionsExecutionLifecycle(t *testing.T) {
	config := NewFunctionalStepFunctionsConfig(
		WithStepFunctionsMetadata("test_type", "execution_lifecycle"),
	)

	stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test-lifecycle-sm"

	// Create different types of executions
	executions := []FunctionalStepFunctionsExecution{
		NewFunctionalStepFunctionsExecution("simple-exec", stateMachineArn, `{}`),
		
		NewFunctionalStepFunctionsExecution("complex-exec", stateMachineArn, `{
			"inputData": {
				"userId": "user-123",
				"requestId": "req-456",
				"timestamp": "2024-01-01T12:00:00Z"
			},
			"options": {
				"timeout": 300,
				"retries": 3
			}
		}`),
		
		NewFunctionalStepFunctionsExecution("batch-exec", stateMachineArn, `{
			"items": [
				{"id": 1, "value": "item1"},
				{"id": 2, "value": "item2"},
				{"id": 3, "value": "item3"}
			],
			"batchSize": 10
		}`),
	}

	// Test execution lifecycle
	for i, execution := range executions {
		// Start execution
		startResult := SimulateFunctionalStepFunctionsStartExecution(execution, config)

		if !startResult.IsSuccess() {
			t.Errorf("Expected execution %d start to succeed, got error: %v", i, startResult.GetError().MustGet())
		}

		// Get execution ARN from result
		var executionArn string
		if startResult.GetResult().IsPresent() {
			resultMap, ok := startResult.GetResult().MustGet().(map[string]interface{})
			if ok {
				executionArn, _ = resultMap["execution_arn"].(string)
			}
		}

		if executionArn == "" {
			t.Errorf("Expected execution %d to have ARN", i)
			continue
		}

		// Describe execution
		describeResult := SimulateFunctionalStepFunctionsDescribeExecution(executionArn, config)

		if !describeResult.IsSuccess() {
			t.Errorf("Expected execution %d describe to succeed, got error: %v", i, describeResult.GetError().MustGet())
		}

		// Verify execution status
		if describeResult.GetResult().IsPresent() {
			resultMap, ok := describeResult.GetResult().MustGet().(map[string]interface{})
			if ok {
				status, _ := resultMap["status"].(string)
				validStatuses := []string{
					string(types.ExecutionStatusRunning),
					string(types.ExecutionStatusSucceeded),
					string(types.ExecutionStatusFailed),
				}
				
				if !lo.Contains(validStatuses, status) {
					t.Errorf("Expected execution %d to have valid status, got %s", i, status)
				}
			}
		}
	}
}

// TestFunctionalStepFunctionsValidationRules tests Step Functions-specific validation
func TestFunctionalStepFunctionsValidationRules(t *testing.T) {
	config := NewFunctionalStepFunctionsConfig()

	// Test state machine name validation
	invalidNames := []string{
		"",                    // Empty
		"invalid@name",        // Contains invalid character
		"name with spaces",    // Contains spaces
		strings.Repeat("a", FunctionalStepFunctionsMaxStateMachineNameLen+1), // Too long
	}

	validDefinition := `{"StartAt":"S1","States":{"S1":{"Type":"Pass","End":true}}}`
	validRoleArn := "arn:aws:iam::123456789012:role/TestRole"

	for _, name := range invalidNames {
		sm := NewFunctionalStepFunctionsStateMachine(name, validDefinition, validRoleArn)
		result := SimulateFunctionalStepFunctionsCreateStateMachine(sm, config)
		
		if result.IsSuccess() {
			t.Errorf("Expected state machine creation to fail for invalid name: %s", name)
		}
	}

	// Test valid state machine names
	validNames := []string{
		"valid-name",
		"ValidName123",
		"test_state_machine",
		"SM-2024-01",
	}

	for _, name := range validNames {
		sm := NewFunctionalStepFunctionsStateMachine(name, validDefinition, validRoleArn)
		result := SimulateFunctionalStepFunctionsCreateStateMachine(sm, config)
		
		if !result.IsSuccess() {
			t.Errorf("Expected state machine creation to succeed for valid name: %s", name)
		}
	}

	// Test definition validation
	invalidDefinitions := []string{
		"",                    // Empty
		"invalid json",        // Invalid JSON
		`{}`,                  // Missing required fields
		`{"StartAt": "S1"}`,   // Missing States
		`{"States": {}}`,      // Missing StartAt
	}

	for _, definition := range invalidDefinitions {
		sm := NewFunctionalStepFunctionsStateMachine("valid-name", definition, validRoleArn)
		result := SimulateFunctionalStepFunctionsCreateStateMachine(sm, config)
		
		if result.IsSuccess() {
			t.Errorf("Expected state machine creation to fail for invalid definition: %s", definition)
		}
	}

	// Test role ARN validation
	invalidRoleArns := []string{
		"",                           // Empty
		"invalid-arn",                // Not an ARN
		"arn:aws:s3:::bucket/object", // Wrong service
		"arn:aws:iam::invalid:role/TestRole", // Invalid account ID
	}

	for _, roleArn := range invalidRoleArns {
		sm := NewFunctionalStepFunctionsStateMachine("valid-name", validDefinition, roleArn)
		result := SimulateFunctionalStepFunctionsCreateStateMachine(sm, config)
		
		if result.IsSuccess() {
			t.Errorf("Expected state machine creation to fail for invalid role ARN: %s", roleArn)
		}
	}

	// Test execution input size validation
	largeInput := fmt.Sprintf(`{"data": "%s"}`, strings.Repeat("x", 33*1024)) // > 32KB
	stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:test"
	
	largeInputExecution := NewFunctionalStepFunctionsExecution("test-exec", stateMachineArn, largeInput)
	result := SimulateFunctionalStepFunctionsStartExecution(largeInputExecution, config)
	
	if result.IsSuccess() {
		t.Error("Expected execution start to fail for large input")
	}
}

// TestFunctionalStepFunctionsActivityManagement tests activity management
func TestFunctionalStepFunctionsActivityManagement(t *testing.T) {
	config := NewFunctionalStepFunctionsConfig(
		WithStepFunctionsMetadata("test_type", "activity_management"),
	)

	// Create different activities
	activities := []FunctionalStepFunctionsActivity{
		NewFunctionalStepFunctionsActivity("simple-activity"),
		NewFunctionalStepFunctionsActivity("complex-activity-2024"),
		NewFunctionalStepFunctionsActivity("data_processing_activity"),
	}

	// Test activity creation
	for i, activity := range activities {
		result := SimulateFunctionalStepFunctionsCreateActivity(activity, config)

		if !result.IsSuccess() {
			t.Errorf("Expected activity %d creation to succeed, got error: %v", i, result.GetError().MustGet())
		}

		// Verify result includes activity information
		if result.GetResult().IsPresent() {
			resultMap, ok := result.GetResult().MustGet().(map[string]interface{})
			if ok {
				if resultMap["name"] != activity.GetName() {
					t.Errorf("Expected activity %d name in result to match", i)
				}
				if resultMap["created"] != true {
					t.Errorf("Expected activity %d to be created", i)
				}
				if resultMap["activity_arn"] == "" {
					t.Errorf("Expected activity %d to have ARN", i)
				}
			}
		}
	}

	// Test invalid activity names
	invalidActivityNames := []string{
		"",
		"invalid@activity",
		"activity with spaces",
		strings.Repeat("a", FunctionalStepFunctionsMaxStateMachineNameLen+1),
	}

	for _, name := range invalidActivityNames {
		activity := NewFunctionalStepFunctionsActivity(name)
		result := SimulateFunctionalStepFunctionsCreateActivity(activity, config)
		
		if result.IsSuccess() {
			t.Errorf("Expected activity creation to fail for invalid name: %s", name)
		}
	}
}

// TestFunctionalStepFunctionsPerformanceMetrics tests performance tracking
func TestFunctionalStepFunctionsPerformanceMetrics(t *testing.T) {
	config := NewFunctionalStepFunctionsConfig(
		WithStepFunctionsTimeout(1*time.Second),
		WithStepFunctionsRetryConfig(1, 50*time.Millisecond), // Faster for testing
		WithStepFunctionsMetadata("test_type", "performance"),
	)

	definition := `{"StartAt":"S1","States":{"S1":{"Type":"Pass","End":true}}}`
	stateMachine := NewFunctionalStepFunctionsStateMachine("perf-sm", definition, "arn:aws:iam::123456789012:role/Role")

	startTime := time.Now()
	result := SimulateFunctionalStepFunctionsCreateStateMachine(stateMachine, config)
	operationTime := time.Since(startTime)

	if !result.IsSuccess() {
		t.Errorf("Expected performance test to succeed, got error: %v", result.GetError().MustGet())
	}

	// Verify duration tracking
	recordedDuration := result.GetDuration()
	if recordedDuration == 0 {
		t.Error("Expected duration to be recorded")
	}

	// Duration should be reasonable (less than total operation time)
	if recordedDuration > operationTime {
		t.Error("Recorded duration should not exceed total operation time")
	}

	// Test multiple operations for performance consistency
	results := []FunctionalStepFunctionsOperationResult{}
	for i := 0; i < 5; i++ {
		testSM := NewFunctionalStepFunctionsStateMachine(
			fmt.Sprintf("perf-sm-%d", i), 
			definition, 
			"arn:aws:iam::123456789012:role/Role",
		)
		
		result := SimulateFunctionalStepFunctionsCreateStateMachine(testSM, config)
		results = append(results, result)
	}

	// All operations should succeed
	allSucceeded := lo.EveryBy(results, func(r FunctionalStepFunctionsOperationResult) bool {
		return r.IsSuccess()
	})

	if !allSucceeded {
		t.Error("Expected all performance test operations to succeed")
	}

	// All operations should have duration recorded
	allHaveDuration := lo.EveryBy(results, func(r FunctionalStepFunctionsOperationResult) bool {
		return r.GetDuration() > 0
	})

	if !allHaveDuration {
		t.Error("Expected all operations to have duration recorded")
	}
}

// BenchmarkFunctionalStepFunctionsScenario benchmarks a complete scenario
func BenchmarkFunctionalStepFunctionsScenario(b *testing.B) {
	config := NewFunctionalStepFunctionsConfig(
		WithStepFunctionsTimeout(5*time.Second),
		WithStepFunctionsRetryConfig(1, 10*time.Millisecond),
	)

	definition := `{"StartAt":"S1","States":{"S1":{"Type":"Pass","End":true}}}`
	stateMachine := NewFunctionalStepFunctionsStateMachine("benchmark-sm", definition, "arn:aws:iam::123456789012:role/Role")
	execution := NewFunctionalStepFunctionsExecution("benchmark-exec", "arn:aws:states:us-east-1:123456789012:stateMachine:benchmark-sm", `{"test": true}`)
	activity := NewFunctionalStepFunctionsActivity("benchmark-activity")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate full Step Functions workflow
		createSMResult := SimulateFunctionalStepFunctionsCreateStateMachine(stateMachine, config)
		if !createSMResult.IsSuccess() {
			b.Errorf("Create state machine operation failed: %v", createSMResult.GetError().MustGet())
		}

		startExecResult := SimulateFunctionalStepFunctionsStartExecution(execution, config)
		if !startExecResult.IsSuccess() {
			b.Errorf("Start execution operation failed: %v", startExecResult.GetError().MustGet())
		}

		// Get execution ARN for describe
		var executionArn string
		if startExecResult.GetResult().IsPresent() {
			if resultMap, ok := startExecResult.GetResult().MustGet().(map[string]interface{}); ok {
				executionArn, _ = resultMap["execution_arn"].(string)
			}
		}

		if executionArn != "" {
			describeExecResult := SimulateFunctionalStepFunctionsDescribeExecution(executionArn, config)
			if !describeExecResult.IsSuccess() {
				b.Errorf("Describe execution operation failed: %v", describeExecResult.GetError().MustGet())
			}
		}

		createActivityResult := SimulateFunctionalStepFunctionsCreateActivity(activity, config)
		if !createActivityResult.IsSuccess() {
			b.Errorf("Create activity operation failed: %v", createActivityResult.GetError().MustGet())
		}
	}
}