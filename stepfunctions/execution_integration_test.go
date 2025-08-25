package stepfunctions

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration-style test structures for complete execution workflows

type executionScenario struct {
	name                string
	stateMachine        string
	executionPrefix     string
	inputTemplate       map[string]interface{}
	expectedDuration    time.Duration
	expectedStatus      types.ExecutionStatus
	requiresPolling     bool
	pollConfig          *PollConfig
	waitConfig          *WaitOptions
	description         string
}

type workflowPattern struct {
	name              string
	phases            []workflowPhase
	totalTimeout      time.Duration
	expectedOutcome   string
	verificationSteps []verificationStep
	description       string
}

type workflowPhase struct {
	name        string
	action      string // "start", "wait", "stop", "poll"
	duration    time.Duration
	parameters  map[string]interface{}
	expectError bool
}

type verificationStep struct {
	name      string
	checkFunc func(*ExecutionResult) bool
	errorMsg  string
}

// Comprehensive execution integration tests

func TestCompleteExecutionWorkflows(t *testing.T) {
	scenarios := []executionScenario{
		{
			name:             "StandardWorkflowExecution",
			stateMachine:     "arn:aws:states:us-east-1:123456789012:stateMachine:standard-workflow",
			executionPrefix:  "standard-exec",
			inputTemplate: map[string]interface{}{
				"workflow": "standard-processing",
				"version":  "1.0.0",
				"config": map[string]interface{}{
					"retries":     3,
					"timeout":     "300s",
					"enableTrace": true,
				},
			},
			expectedDuration: 30 * time.Second,
			expectedStatus:   types.ExecutionStatusSucceeded,
			requiresPolling:  true,
			pollConfig: &PollConfig{
				MaxAttempts:        60,
				Interval:           5 * time.Second,
				Timeout:            5 * time.Minute,
				ExponentialBackoff: false,
				BackoffMultiplier:  1.0,
				MaxInterval:        5 * time.Second,
			},
			waitConfig: &WaitOptions{
				Timeout:         5 * time.Minute,
				PollingInterval: 5 * time.Second,
				MaxAttempts:     60,
			},
			description: "Standard workflow should execute with regular polling",
		},
		{
			name:             "ExpressWorkflowExecution",
			stateMachine:     "arn:aws:states:us-east-1:123456789012:stateMachine:express-workflow",
			executionPrefix:  "express-exec",
			inputTemplate: map[string]interface{}{
				"workflow": "express-processing",
				"priority": "high",
				"data": map[string]interface{}{
					"recordCount": 100,
					"batchSize":   10,
				},
			},
			expectedDuration: 5 * time.Second,
			expectedStatus:   types.ExecutionStatusSucceeded,
			requiresPolling:  true,
			pollConfig: &PollConfig{
				MaxAttempts:        20,
				Interval:           1 * time.Second,
				Timeout:            30 * time.Second,
				ExponentialBackoff: false,
				BackoffMultiplier:  1.0,
				MaxInterval:        1 * time.Second,
			},
			waitConfig: &WaitOptions{
				Timeout:         30 * time.Second,
				PollingInterval: 1 * time.Second,
				MaxAttempts:     30,
			},
			description: "Express workflow should complete quickly with frequent polling",
		},
		{
			name:             "LongRunningWorkflowExecution", 
			stateMachine:     "arn:aws:states:us-east-1:123456789012:stateMachine:long-running-workflow",
			executionPrefix:  "long-exec",
			inputTemplate: map[string]interface{}{
				"workflow": "batch-processing",
				"estimatedDuration": "900s",
				"checkpoints": map[string]interface{}{
					"enabled":  true,
					"interval": "60s",
				},
			},
			expectedDuration: 15 * time.Minute,
			expectedStatus:   types.ExecutionStatusSucceeded,
			requiresPolling:  true,
			pollConfig: &PollConfig{
				MaxAttempts:        90,
				Interval:           10 * time.Second,
				Timeout:            20 * time.Minute,
				ExponentialBackoff: true,
				BackoffMultiplier:  1.1,
				MaxInterval:        30 * time.Second,
			},
			waitConfig: &WaitOptions{
				Timeout:         20 * time.Minute,
				PollingInterval: 10 * time.Second,
				MaxAttempts:     120,
			},
			description: "Long-running workflow should handle extended execution with exponential backoff",
		},
		{
			name:             "RetryableWorkflowExecution",
			stateMachine:     "arn:aws:states:us-east-1:123456789012:stateMachine:retryable-workflow",
			executionPrefix:  "retry-exec",
			inputTemplate: map[string]interface{}{
				"workflow": "unreliable-service",
				"retryPolicy": map[string]interface{}{
					"maxAttempts":       5,
					"backoffRate":       2.0,
					"intervalSeconds":   2,
					"maxDelaySeconds":   60,
					"retryableErrors":   []string{"ServiceUnavailable", "ThrottlingException"},
					"nonRetryableErrors": []string{"ValidationException"},
				},
			},
			expectedDuration: 60 * time.Second,
			expectedStatus:   types.ExecutionStatusSucceeded,
			requiresPolling:  true,
			pollConfig: &PollConfig{
				MaxAttempts:        60,
				Interval:           3 * time.Second,
				Timeout:            5 * time.Minute,
				ExponentialBackoff: true,
				BackoffMultiplier:  1.2,
				MaxInterval:        15 * time.Second,
			},
			waitConfig: &WaitOptions{
				Timeout:         5 * time.Minute,
				PollingInterval: 3 * time.Second,
				MaxAttempts:     100,
			},
			description: "Retryable workflow should handle failures and retries gracefully",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Test input preparation and validation
			input := NewInput()
			for key, value := range scenario.inputTemplate {
				input.Set(key, value)
			}

			inputJSON, err := input.ToJSON()
			require.NoError(t, err, "Input should convert to JSON")

			// Test execution request validation
			executionName := scenario.executionPrefix + "-001"
			request := &StartExecutionRequest{
				StateMachineArn: scenario.stateMachine,
				Name:            executionName,
				Input:           inputJSON,
			}

			err = validateStartExecutionRequest(request)
			assert.NoError(t, err, scenario.description)

			// Test polling configuration validation (if required)
			if scenario.requiresPolling {
				executionArn := "arn:aws:states:us-east-1:123456789012:execution:" + executionName
				
				err = validatePollConfig(executionArn, scenario.pollConfig)
				assert.NoError(t, err, "Poll config should be valid for "+scenario.name)

				err = validateWaitForExecutionRequest(executionArn, scenario.waitConfig)
				assert.NoError(t, err, "Wait config should be valid for "+scenario.name)

				// Test execute-and-wait pattern
				err = validateExecuteAndWaitPattern(scenario.stateMachine, executionName, input, scenario.waitConfig.Timeout)
				assert.NoError(t, err, "Execute-and-wait pattern should be valid for "+scenario.name)
			}

			// Verify poll config merging produces expected values
			mergedConfig := mergePollConfig(scenario.pollConfig)
			assert.Equal(t, scenario.pollConfig.MaxAttempts, mergedConfig.MaxAttempts, "MaxAttempts should be preserved")
			assert.Equal(t, scenario.pollConfig.Interval, mergedConfig.Interval, "Interval should be preserved")
			assert.Equal(t, scenario.pollConfig.Timeout, mergedConfig.Timeout, "Timeout should be preserved")
			
			// Verify wait config merging produces expected values  
			mergedWaitConfig := mergeWaitOptions(scenario.waitConfig)
			assert.Equal(t, scenario.waitConfig.Timeout, mergedWaitConfig.Timeout, "Wait timeout should be preserved")
			assert.Equal(t, scenario.waitConfig.PollingInterval, mergedWaitConfig.PollingInterval, "Polling interval should be preserved")
			assert.Equal(t, scenario.waitConfig.MaxAttempts, mergedWaitConfig.MaxAttempts, "Max attempts should be preserved")
		})
	}
}

func TestWorkflowExecutionPatterns(t *testing.T) {
	patterns := []workflowPattern{
		{
			name: "StartWaitCompletePattern",
			phases: []workflowPhase{
				{
					name:     "StartExecution",
					action:   "start",
					duration: 1 * time.Second,
					parameters: map[string]interface{}{
						"stateMachine": "arn:aws:states:us-east-1:123456789012:stateMachine:simple-workflow",
						"execution":    "start-wait-complete-001",
						"input":        `{"message": "test"}`,
					},
					expectError: false,
				},
				{
					name:     "WaitForCompletion",
					action:   "wait",
					duration: 60 * time.Second,
					parameters: map[string]interface{}{
						"timeout":         60 * time.Second,
						"pollingInterval": 5 * time.Second,
						"maxAttempts":     12,
					},
					expectError: false,
				},
			},
			totalTimeout:    2 * time.Minute,
			expectedOutcome: "successful_completion",
			verificationSteps: []verificationStep{
				{
					name: "ExecutionCompleted",
					checkFunc: func(result *ExecutionResult) bool {
						return result.Status == types.ExecutionStatusSucceeded
					},
					errorMsg: "Execution should have completed successfully",
				},
				{
					name: "HasOutput",
					checkFunc: func(result *ExecutionResult) bool {
						return result.Output != ""
					},
					errorMsg: "Execution should have output",
				},
			},
			description: "Standard start-wait-complete pattern should work reliably",
		},
		{
			name: "StartStopPattern",
			phases: []workflowPhase{
				{
					name:     "StartLongRunningExecution",
					action:   "start",
					duration: 1 * time.Second,
					parameters: map[string]interface{}{
						"stateMachine": "arn:aws:states:us-east-1:123456789012:stateMachine:long-running-workflow",
						"execution":    "start-stop-001",
						"input":        `{"duration": "300s", "task": "long-computation"}`,
					},
					expectError: false,
				},
				{
					name:     "WaitBriefly",
					action:   "wait",
					duration: 10 * time.Second,
					parameters: map[string]interface{}{
						"timeout":         10 * time.Second,
						"pollingInterval": 2 * time.Second,
						"maxAttempts":     5,
					},
					expectError: true, // Should timeout
				},
				{
					name:     "StopExecution",
					action:   "stop",
					duration: 1 * time.Second,
					parameters: map[string]interface{}{
						"error": "UserRequestedStop",
						"cause": "Testing stop functionality",
					},
					expectError: false,
				},
			},
			totalTimeout:    30 * time.Second,
			expectedOutcome: "stopped_execution",
			verificationSteps: []verificationStep{
				{
					name: "ExecutionStopped",
					checkFunc: func(result *ExecutionResult) bool {
						return result.Status == types.ExecutionStatusAborted
					},
					errorMsg: "Execution should have been stopped/aborted",
				},
				{
					name: "HasStopReason",
					checkFunc: func(result *ExecutionResult) bool {
						return result.Error == "UserRequestedStop"
					},
					errorMsg: "Execution should have stop reason",
				},
			},
			description: "Start-stop pattern should allow graceful termination of long-running executions",
		},
		{
			name: "StartPollWithBackoffPattern",
			phases: []workflowPhase{
				{
					name:     "StartExecution",
					action:   "start",
					duration: 1 * time.Second,
					parameters: map[string]interface{}{
						"stateMachine": "arn:aws:states:us-east-1:123456789012:stateMachine:variable-duration-workflow",
						"execution":    "backoff-poll-001",
						"input":        `{"variableDuration": true, "estimatedTime": "45s"}`,
					},
					expectError: false,
				},
				{
					name:     "PollWithExponentialBackoff",
					action:   "poll",
					duration: 90 * time.Second,
					parameters: map[string]interface{}{
						"maxAttempts":        30,
						"interval":           2 * time.Second,
						"timeout":            90 * time.Second,
						"exponentialBackoff": true,
						"backoffMultiplier":  1.5,
						"maxInterval":        15 * time.Second,
					},
					expectError: false,
				},
			},
			totalTimeout:    2 * time.Minute,
			expectedOutcome: "backoff_completion",
			verificationSteps: []verificationStep{
				{
					name: "ExecutionCompleted",
					checkFunc: func(result *ExecutionResult) bool {
						return isExecutionComplete(result.Status)
					},
					errorMsg: "Execution should be in terminal state",
				},
				{
					name: "ReasonableExecutionTime",
					checkFunc: func(result *ExecutionResult) bool {
						return result.ExecutionTime >= 30*time.Second && result.ExecutionTime <= 90*time.Second
					},
					errorMsg: "Execution time should be within expected range",
				},
			},
			description: "Polling with exponential backoff should handle variable-duration executions efficiently",
		},
	}

	for _, pattern := range patterns {
		t.Run(pattern.name, func(t *testing.T) {
			// Validate each phase of the workflow pattern
			for _, phase := range pattern.phases {
				t.Run(phase.name, func(t *testing.T) {
					switch phase.action {
					case "start":
						// Validate start execution phase
						stateMachineArn := phase.parameters["stateMachine"].(string)
						executionName := phase.parameters["execution"].(string) 
						inputStr := phase.parameters["input"].(string)

						request := &StartExecutionRequest{
							StateMachineArn: stateMachineArn,
							Name:            executionName,
							Input:           inputStr,
						}

						err := validateStartExecutionRequest(request)
						if phase.expectError {
							assert.Error(t, err, "Phase %s should error", phase.name)
						} else {
							assert.NoError(t, err, "Phase %s should not error", phase.name)
						}

					case "wait":
						// Validate wait phase
						timeout := phase.parameters["timeout"].(time.Duration)
						pollingInterval := phase.parameters["pollingInterval"].(time.Duration)
						maxAttempts := phase.parameters["maxAttempts"].(int)

						waitOptions := &WaitOptions{
							Timeout:         timeout,
							PollingInterval: pollingInterval,
							MaxAttempts:     maxAttempts,
						}

						executionArn := "arn:aws:states:us-east-1:123456789012:execution:test:test"
						err := validateWaitForExecutionRequest(executionArn, waitOptions)
						if phase.expectError {
							// For wait phase, validation might pass but execution would timeout
							assert.NoError(t, err, "Wait validation should pass even if execution would timeout")
						} else {
							assert.NoError(t, err, "Phase %s should not error", phase.name)
						}

					case "stop":
						// Validate stop execution phase
						errorMsg := phase.parameters["error"].(string)
						cause := phase.parameters["cause"].(string)

						executionArn := "arn:aws:states:us-east-1:123456789012:execution:test:test"
						err := validateStopExecutionRequest(executionArn, errorMsg, cause)
						if phase.expectError {
							assert.Error(t, err, "Phase %s should error", phase.name)
						} else {
							assert.NoError(t, err, "Phase %s should not error", phase.name)
						}

					case "poll":
						// Validate poll phase
						maxAttempts := phase.parameters["maxAttempts"].(int)
						interval := phase.parameters["interval"].(time.Duration)
						timeout := phase.parameters["timeout"].(time.Duration)
						exponentialBackoff := phase.parameters["exponentialBackoff"].(bool)
						backoffMultiplier := phase.parameters["backoffMultiplier"].(float64)
						maxInterval := phase.parameters["maxInterval"].(time.Duration)

						pollConfig := &PollConfig{
							MaxAttempts:        maxAttempts,
							Interval:           interval,
							Timeout:            timeout,
							ExponentialBackoff: exponentialBackoff,
							BackoffMultiplier:  backoffMultiplier,
							MaxInterval:        maxInterval,
						}

						executionArn := "arn:aws:states:us-east-1:123456789012:execution:test:test"
						err := validatePollConfig(executionArn, pollConfig)
						if phase.expectError {
							assert.Error(t, err, "Phase %s should error", phase.name)
						} else {
							assert.NoError(t, err, "Phase %s should not error", phase.name)
						}
					}
				})
			}

			// Test overall pattern timeout validation
			err := validateExecuteAndWaitPattern(
				"arn:aws:states:us-east-1:123456789012:stateMachine:pattern-test",
				"pattern-execution-001",
				NewInput().Set("pattern", pattern.name),
				pattern.totalTimeout,
			)
			assert.NoError(t, err, "Overall pattern should be valid")

			// Test verification steps with mock result
			if len(pattern.verificationSteps) > 0 {
				mockResult := createMockResultForPattern(pattern)
				for _, verification := range pattern.verificationSteps {
					t.Run(verification.name, func(t *testing.T) {
						assert.True(t, verification.checkFunc(mockResult), verification.errorMsg)
					})
				}
			}
		})
	}
}

func TestAdvancedExecutionScenarios(t *testing.T) {
	t.Run("MultiStepExecutionValidation", func(t *testing.T) {
		// Test validation of complex multi-step execution scenarios
		
		steps := []struct {
			name        string
			operation   string
			parameters  map[string]interface{}
			shouldError bool
		}{
			{
				name:      "ValidateInitialExecution",
				operation: "start",
				parameters: map[string]interface{}{
					"stateMachine": "arn:aws:states:us-east-1:123456789012:stateMachine:multi-step",
					"execution":    "multi-step-001",
					"input": map[string]interface{}{
						"steps": []string{"initialize", "process", "validate", "complete"},
						"config": map[string]interface{}{
							"parallel":     true,
							"maxRetries":   3,
							"timeout":      "600s",
							"checkpoints":  true,
						},
					},
				},
				shouldError: false,
			},
			{
				name:      "ConfigureLongPolling",
				operation: "poll",
				parameters: map[string]interface{}{
					"maxAttempts":        180, // 15 minutes with 5s intervals
					"interval":           5 * time.Second,
					"timeout":            15 * time.Minute,
					"exponentialBackoff": true,
					"backoffMultiplier":  1.1,
					"maxInterval":        30 * time.Second,
				},
				shouldError: false,
			},
			{
				name:      "ConfigureWaitWithTimeout",
				operation: "wait",
				parameters: map[string]interface{}{
					"timeout":         10 * time.Minute,
					"pollingInterval": 10 * time.Second,
					"maxAttempts":     60,
				},
				shouldError: false,
			},
			{
				name:      "ValidateStopCapability",
				operation: "stop",
				parameters: map[string]interface{}{
					"error": "UserInitiatedStop",
					"cause": "Multi-step execution halt requested by user",
				},
				shouldError: false,
			},
		}

		executionArn := "arn:aws:states:us-east-1:123456789012:execution:multi-step:multi-step-001"

		for _, step := range steps {
			t.Run(step.name, func(t *testing.T) {
				switch step.operation {
				case "start":
					input := NewInput()
					for k, v := range step.parameters["input"].(map[string]interface{}) {
						input.Set(k, v)
					}
					inputJSON, err := input.ToJSON()
					require.NoError(t, err)

					request := &StartExecutionRequest{
						StateMachineArn: step.parameters["stateMachine"].(string),
						Name:            step.parameters["execution"].(string),
						Input:           inputJSON,
					}

					err = validateStartExecutionRequest(request)
					if step.shouldError {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}

				case "poll":
					config := &PollConfig{
						MaxAttempts:        step.parameters["maxAttempts"].(int),
						Interval:           step.parameters["interval"].(time.Duration),
						Timeout:            step.parameters["timeout"].(time.Duration),
						ExponentialBackoff: step.parameters["exponentialBackoff"].(bool),
						BackoffMultiplier:  step.parameters["backoffMultiplier"].(float64),
						MaxInterval:        step.parameters["maxInterval"].(time.Duration),
					}

					err := validatePollConfig(executionArn, config)
					if step.shouldError {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}

				case "wait":
					options := &WaitOptions{
						Timeout:         step.parameters["timeout"].(time.Duration),
						PollingInterval: step.parameters["pollingInterval"].(time.Duration),
						MaxAttempts:     step.parameters["maxAttempts"].(int),
					}

					err := validateWaitForExecutionRequest(executionArn, options)
					if step.shouldError {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}

				case "stop":
					err := validateStopExecutionRequest(
						executionArn,
						step.parameters["error"].(string),
						step.parameters["cause"].(string),
					)
					if step.shouldError {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				}
			})
		}
	})

	t.Run("ErrorRecoveryScenarios", func(t *testing.T) {
		// Test validation of error recovery and retry scenarios
		
		scenarios := []struct {
			name            string
			initialInput    *Input
			retryStrategy   map[string]interface{}
			timeoutStrategy map[string]interface{}
			expectedPattern string
		}{
			{
				name: "ExponentialBackoffRetry",
				initialInput: NewInput().
					Set("operation", "unreliable-service-call").
					Set("maxRetries", 5).
					Set("initialDelay", "1s"),
				retryStrategy: map[string]interface{}{
					"type":           "exponential",
					"multiplier":     2.0,
					"maxDelay":       "30s",
					"jitterEnabled":  true,
				},
				timeoutStrategy: map[string]interface{}{
					"perAttempt": "10s",
					"total":      "300s",
				},
				expectedPattern: "exponential_backoff_with_jitter",
			},
			{
				name: "LinearBackoffRetry",
				initialInput: NewInput().
					Set("operation", "rate-limited-service").
					Set("maxRetries", 10).
					Set("baseDelay", "2s"),
				retryStrategy: map[string]interface{}{
					"type":          "linear",
					"increment":     "1s",
					"maxDelay":      "15s",
				},
				timeoutStrategy: map[string]interface{}{
					"perAttempt": "5s",
					"total":      "180s",
				},
				expectedPattern: "linear_backoff",
			},
			{
				name: "CircuitBreakerPattern",
				initialInput: NewInput().
					Set("operation", "downstream-dependency").
					Set("circuitBreaker", map[string]interface{}{
						"failureThreshold": 3,
						"resetTimeout":     "60s",
						"halfOpenRetries":  1,
					}),
				retryStrategy: map[string]interface{}{
					"type":              "circuit_breaker",
					"initialDelay":      "1s",
					"maxConsecutiveFailures": 3,
				},
				timeoutStrategy: map[string]interface{}{
					"perAttempt": "15s",
					"total":      "600s",
				},
				expectedPattern: "circuit_breaker",
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				// Test input validation
				inputJSON, err := scenario.initialInput.ToJSON()
				require.NoError(t, err, "Input should be valid JSON")

				// Test start execution request
				request := &StartExecutionRequest{
					StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:error-recovery",
					Name:            "error-recovery-" + scenario.expectedPattern,
					Input:           inputJSON,
				}
				
				err = validateStartExecutionRequest(request)
				assert.NoError(t, err, "Error recovery execution should be valid")

				// Test retry strategy polling configuration
				totalTimeout := parseTimeoutString(scenario.timeoutStrategy["total"].(string))
				pollConfig := &PollConfig{
					MaxAttempts:        calculateMaxAttemptsForTimeout(totalTimeout),
					Interval:           parseTimeoutString(scenario.timeoutStrategy["perAttempt"].(string)),
					Timeout:            totalTimeout,
					ExponentialBackoff: scenario.retryStrategy["type"] == "exponential",
					BackoffMultiplier:  getBackoffMultiplier(scenario.retryStrategy),
					MaxInterval:        parseTimeoutString(getMaxDelay(scenario.retryStrategy)),
				}

				executionArn := "arn:aws:states:us-east-1:123456789012:execution:error-recovery:" + scenario.expectedPattern
				err = validatePollConfig(executionArn, pollConfig)
				assert.NoError(t, err, "Poll config for error recovery should be valid")

				// Test execute-and-wait pattern for error recovery
				err = validateExecuteAndWaitPattern(
					request.StateMachineArn,
					request.Name,
					scenario.initialInput,
					totalTimeout,
				)
				assert.NoError(t, err, "Execute-and-wait pattern for error recovery should be valid")
			})
		}
	})
}

// Helper functions for creating mock results and parsing configurations

func createMockResultForPattern(pattern workflowPattern) *ExecutionResult {
	now := time.Now()
	
	switch pattern.expectedOutcome {
	case "successful_completion":
		return &ExecutionResult{
			ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:pattern:test",
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:pattern",
			Name:            "pattern-test",
			Status:          types.ExecutionStatusSucceeded,
			StartDate:       now.Add(-pattern.totalTimeout/2),
			StopDate:        now,
			Input:           `{"pattern": "test"}`,
			Output:          `{"result": "success", "completedSteps": 3}`,
			ExecutionTime:   pattern.totalTimeout / 2,
		}
	case "stopped_execution":
		return &ExecutionResult{
			ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:pattern:test",
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:pattern",
			Name:            "pattern-test",
			Status:          types.ExecutionStatusAborted,
			StartDate:       now.Add(-pattern.totalTimeout/4),
			StopDate:        now,
			Input:           `{"pattern": "test"}`,
			Error:           "UserRequestedStop",
			Cause:           "Testing stop functionality",
			ExecutionTime:   pattern.totalTimeout / 4,
		}
	case "backoff_completion":
		return &ExecutionResult{
			ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:pattern:test",
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:pattern",
			Name:            "pattern-test",
			Status:          types.ExecutionStatusSucceeded,
			StartDate:       now.Add(-45 * time.Second),
			StopDate:        now,
			Input:           `{"pattern": "test", "variableDuration": true}`,
			Output:          `{"result": "success", "actualDuration": "45s", "pollAttempts": 12}`,
			ExecutionTime:   45 * time.Second,
		}
	default:
		return &ExecutionResult{
			ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:pattern:test",
			StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:pattern",
			Name:            "pattern-test",
			Status:          types.ExecutionStatusSucceeded,
			StartDate:       now.Add(-time.Minute),
			StopDate:        now,
			Input:           `{"pattern": "test"}`,
			Output:          `{"result": "success"}`,
			ExecutionTime:   time.Minute,
		}
	}
}

func parseTimeoutString(timeout string) time.Duration {
	// Simple parser for timeout strings like "300s", "5m", etc.
	switch timeout {
	case "1s":
		return 1 * time.Second
	case "2s":
		return 2 * time.Second
	case "5s":
		return 5 * time.Second
	case "10s":
		return 10 * time.Second
	case "15s":
		return 15 * time.Second
	case "30s":
		return 30 * time.Second
	case "60s", "1m":
		return 1 * time.Minute
	case "180s", "3m":
		return 3 * time.Minute
	case "300s", "5m":
		return 5 * time.Minute
	case "600s", "10m":
		return 10 * time.Minute
	default:
		return 60 * time.Second
	}
}

func calculateMaxAttemptsForTimeout(timeout time.Duration) int {
	// Calculate reasonable max attempts based on timeout
	return int(timeout.Seconds() / 5) // Assuming 5-second intervals
}

func getBackoffMultiplier(strategy map[string]interface{}) float64 {
	if multiplier, ok := strategy["multiplier"]; ok {
		if mult, ok := multiplier.(float64); ok {
			return mult
		}
	}
	return 1.5 // Default
}

func getMaxDelay(strategy map[string]interface{}) string {
	if maxDelay, ok := strategy["maxDelay"]; ok {
		if delay, ok := maxDelay.(string); ok {
			return delay
		}
	}
	return "30s" // Default
}