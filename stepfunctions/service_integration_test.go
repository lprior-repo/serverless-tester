// Package stepfunctions service integration tests - AWS service integration testing
// Following strict TDD methodology with comprehensive testing patterns

package stepfunctions

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data for service integration
func createTestLambdaIntegrationDefinition() string {
	return `{
		"Comment": "Lambda integration test",
		"StartAt": "InvokeLambda",
		"States": {
			"InvokeLambda": {
				"Type": "Task",
				"Resource": "arn:aws:lambda:us-east-1:123456789012:function:TestFunction",
				"End": true
			}
		}
	}`
}

func createTestDynamoDBIntegrationDefinition() string {
	return `{
		"Comment": "DynamoDB integration test",
		"StartAt": "PutItem",
		"States": {
			"PutItem": {
				"Type": "Task",
				"Resource": "arn:aws:states:::dynamodb:putItem",
				"Parameters": {
					"TableName": "TestTable",
					"Item": {
						"id": {"S.$": "$.id"},
						"data": {"S.$": "$.data"}
					}
				},
				"End": true
			}
		}
	}`
}

func createTestSNSIntegrationDefinition() string {
	return `{
		"Comment": "SNS integration test",
		"StartAt": "PublishMessage",
		"States": {
			"PublishMessage": {
				"Type": "Task",
				"Resource": "arn:aws:states:::sns:publish",
				"Parameters": {
					"TopicArn": "arn:aws:sns:us-east-1:123456789012:TestTopic",
					"Message.$": "$.message"
				},
				"End": true
			}
		}
	}`
}

func createTestComplexWorkflowDefinition() string {
	return `{
		"Comment": "Complex workflow with multiple service integrations",
		"StartAt": "ProcessData",
		"States": {
			"ProcessData": {
				"Type": "Task",
				"Resource": "arn:aws:lambda:us-east-1:123456789012:function:DataProcessor",
				"Next": "StoreToDB"
			},
			"StoreToDB": {
				"Type": "Task",
				"Resource": "arn:aws:states:::dynamodb:putItem",
				"Parameters": {
					"TableName": "ProcessedData",
					"Item": {
						"id": {"S.$": "$.id"},
						"result": {"S.$": "$.result"},
						"timestamp": {"S.$": "$$.State.EnteredTime"}
					}
				},
				"Next": "SendNotification"
			},
			"SendNotification": {
				"Type": "Task",
				"Resource": "arn:aws:states:::sns:publish",
				"Parameters": {
					"TopicArn": "arn:aws:sns:us-east-1:123456789012:ProcessedTopic",
					"Message.$": "$.result"
				},
				"End": true
			}
		}
	}`
}

// TestLambdaIntegrationValidation tests Lambda service integration validation
func TestLambdaIntegrationValidation(t *testing.T) {
	tests := []struct {
		name           string
		stateMachine   *StateMachineDefinition
		input          *Input
		expectedError  string
		shouldValidate bool
	}{
		{
			name: "ValidLambdaIntegration",
			stateMachine: &StateMachineDefinition{
				Name:       "test-lambda-integration",
				Definition: createTestLambdaIntegrationDefinition(),
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			input:          NewInput().Set("payload", "test data"),
			shouldValidate: true,
		},
		{
			name: "InvalidLambdaArn",
			stateMachine: &StateMachineDefinition{
				Name: "test-invalid-lambda",
				Definition: `{
					"Comment": "Invalid Lambda ARN",
					"StartAt": "InvokeLambda",
					"States": {
						"InvokeLambda": {
							"Type": "Task",
							"Resource": "invalid-lambda-arn",
							"End": true
						}
					}
				}`,
				RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:    types.StateMachineTypeStandard,
			},
			input:         NewInput().Set("payload", "test data"),
			expectedError: "invalid lambda resource ARN",
		},
		{
			name: "MissingLambdaFunction",
			stateMachine: &StateMachineDefinition{
				Name: "test-missing-lambda",
				Definition: `{
					"Comment": "Missing Lambda function",
					"StartAt": "InvokeLambda",
					"States": {
						"InvokeLambda": {
							"Type": "Task",
							"Resource": "",
							"End": true
						}
					}
				}`,
				RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:    types.StateMachineTypeStandard,
			},
			input:         NewInput().Set("payload", "test data"),
			expectedError: "empty resource ARN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			err := ValidateLambdaIntegrationE(tt.stateMachine, tt.input)

			if tt.shouldValidate {
				assert.NoError(t, err, "Lambda integration should be valid")
			} else {
				require.Error(t, err, "Lambda integration should be invalid")
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

// TestDynamoDBIntegrationValidation tests DynamoDB service integration validation
func TestDynamoDBIntegrationValidation(t *testing.T) {
	tests := []struct {
		name           string
		stateMachine   *StateMachineDefinition
		input          *Input
		expectedError  string
		shouldValidate bool
	}{
		{
			name: "ValidDynamoDBIntegration",
			stateMachine: &StateMachineDefinition{
				Name:       "test-dynamodb-integration",
				Definition: createTestDynamoDBIntegrationDefinition(),
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			input:          NewInput().Set("id", "test-id").Set("data", "test-data"),
			shouldValidate: true,
		},
		{
			name: "MissingTableName",
			stateMachine: &StateMachineDefinition{
				Name: "test-missing-table",
				Definition: `{
					"Comment": "Missing table name",
					"StartAt": "PutItem",
					"States": {
						"PutItem": {
							"Type": "Task",
							"Resource": "arn:aws:states:::dynamodb:putItem",
							"Parameters": {
								"Item": {
									"id": {"S.$": "$.id"}
								}
							},
							"End": true
						}
					}
				}`,
				RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:    types.StateMachineTypeStandard,
			},
			input:         NewInput().Set("id", "test-id"),
			expectedError: "missing TableName parameter",
		},
		{
			name: "InvalidDynamoDBOperation",
			stateMachine: &StateMachineDefinition{
				Name: "test-invalid-operation",
				Definition: `{
					"Comment": "Invalid DynamoDB operation",
					"StartAt": "InvalidOp",
					"States": {
						"InvalidOp": {
							"Type": "Task",
							"Resource": "arn:aws:states:::dynamodb:invalidOperation",
							"End": true
						}
					}
				}`,
				RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:    types.StateMachineTypeStandard,
			},
			input:         NewInput().Set("id", "test-id"),
			expectedError: "unsupported DynamoDB operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			err := ValidateDynamoDBIntegrationE(tt.stateMachine, tt.input)

			if tt.shouldValidate {
				assert.NoError(t, err, "DynamoDB integration should be valid")
			} else {
				require.Error(t, err, "DynamoDB integration should be invalid")
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

// TestSNSIntegrationValidation tests SNS service integration validation
func TestSNSIntegrationValidation(t *testing.T) {
	tests := []struct {
		name           string
		stateMachine   *StateMachineDefinition
		input          *Input
		expectedError  string
		shouldValidate bool
	}{
		{
			name: "ValidSNSIntegration",
			stateMachine: &StateMachineDefinition{
				Name:       "test-sns-integration",
				Definition: createTestSNSIntegrationDefinition(),
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			input:          NewInput().Set("message", "test notification"),
			shouldValidate: true,
		},
		{
			name: "MissingTopicArn",
			stateMachine: &StateMachineDefinition{
				Name: "test-missing-topic",
				Definition: `{
					"Comment": "Missing topic ARN",
					"StartAt": "PublishMessage",
					"States": {
						"PublishMessage": {
							"Type": "Task",
							"Resource": "arn:aws:states:::sns:publish",
							"Parameters": {
								"Message.$": "$.message"
							},
							"End": true
						}
					}
				}`,
				RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:    types.StateMachineTypeStandard,
			},
			input:         NewInput().Set("message", "test"),
			expectedError: "missing TopicArn parameter",
		},
		{
			name: "InvalidTopicArn",
			stateMachine: &StateMachineDefinition{
				Name: "test-invalid-topic",
				Definition: `{
					"Comment": "Invalid topic ARN",
					"StartAt": "PublishMessage",
					"States": {
						"PublishMessage": {
							"Type": "Task",
							"Resource": "arn:aws:states:::sns:publish",
							"Parameters": {
								"TopicArn": "invalid-topic-arn",
								"Message.$": "$.message"
							},
							"End": true
						}
					}
				}`,
				RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:    types.StateMachineTypeStandard,
			},
			input:         NewInput().Set("message", "test"),
			expectedError: "invalid SNS topic ARN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			err := ValidateSNSIntegrationE(tt.stateMachine, tt.input)

			if tt.shouldValidate {
				assert.NoError(t, err, "SNS integration should be valid")
			} else {
				require.Error(t, err, "SNS integration should be invalid")
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

// TestComplexWorkflowValidation tests complex multi-service workflow validation
func TestComplexWorkflowValidation(t *testing.T) {
	tests := []struct {
		name           string
		stateMachine   *StateMachineDefinition
		input          *Input
		expectedError  string
		shouldValidate bool
	}{
		{
			name: "ValidComplexWorkflow",
			stateMachine: &StateMachineDefinition{
				Name:       "test-complex-workflow",
				Definition: createTestComplexWorkflowDefinition(),
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			input:          NewInput().Set("id", "workflow-1").Set("data", "initial data"),
			shouldValidate: true,
		},
		{
			name: "MissingInputData",
			stateMachine: &StateMachineDefinition{
				Name:       "test-missing-input",
				Definition: createTestComplexWorkflowDefinition(),
				RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
				Type:       types.StateMachineTypeStandard,
			},
			input:         NewInput(), // Empty input
			expectedError: "missing required input field 'id'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			err := ValidateComplexWorkflowE(tt.stateMachine, tt.input)

			if tt.shouldValidate {
				assert.NoError(t, err, "Complex workflow should be valid")
			} else {
				require.Error(t, err, "Complex workflow should be invalid")
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

// TestServiceIntegrationResourceExtraction tests extracting service resources from definitions
func TestServiceIntegrationResourceExtraction(t *testing.T) {
	tests := []struct {
		name               string
		definition         string
		expectedLambdas    []string
		expectedDynamoTables []string
		expectedSNSTopics  []string
	}{
		{
			name:            "LambdaOnlyWorkflow",
			definition:      createTestLambdaIntegrationDefinition(),
			expectedLambdas: []string{"arn:aws:lambda:us-east-1:123456789012:function:TestFunction"},
		},
		{
			name:                 "DynamoDBOnlyWorkflow", 
			definition:           createTestDynamoDBIntegrationDefinition(),
			expectedDynamoTables: []string{"TestTable"},
		},
		{
			name:              "SNSOnlyWorkflow",
			definition:        createTestSNSIntegrationDefinition(),
			expectedSNSTopics: []string{"arn:aws:sns:us-east-1:123456789012:TestTopic"},
		},
		{
			name:                 "ComplexWorkflow",
			definition:           createTestComplexWorkflowDefinition(),
			expectedLambdas:      []string{"arn:aws:lambda:us-east-1:123456789012:function:DataProcessor"},
			expectedDynamoTables: []string{"ProcessedData"},
			expectedSNSTopics:    []string{"arn:aws:sns:us-east-1:123456789012:ProcessedTopic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			lambdas := ExtractLambdaFunctions(tt.definition)
			dynamoTables := ExtractDynamoDBTables(tt.definition)
			snsTopics := ExtractSNSTopics(tt.definition)

			assert.Equal(t, tt.expectedLambdas, lambdas, "Lambda functions should match")
			assert.Equal(t, tt.expectedDynamoTables, dynamoTables, "DynamoDB tables should match")
			assert.Equal(t, tt.expectedSNSTopics, snsTopics, "SNS topics should match")
		})
	}
}

// TestWorkflowResourceDependencies tests analyzing resource dependencies in workflows
func TestWorkflowResourceDependencies(t *testing.T) {
	tests := []struct {
		name         string
		definition   string
		expectedDeps map[string][]string
	}{
		{
			name:       "SimpleLinearDependencies",
			definition: createTestComplexWorkflowDefinition(),
			expectedDeps: map[string][]string{
				"ProcessData":      {},
				"StoreToDB":        {"ProcessData"},
				"SendNotification": {"StoreToDB"},
			},
		},
		{
			name: "ParallelDependencies",
			definition: `{
				"StartAt": "Parallel",
				"States": {
					"Parallel": {
						"Type": "Parallel",
						"Branches": [
							{
								"StartAt": "Lambda1",
								"States": {
									"Lambda1": {
										"Type": "Task",
										"Resource": "arn:aws:lambda:us-east-1:123456789012:function:Func1",
										"End": true
									}
								}
							},
							{
								"StartAt": "Lambda2",
								"States": {
									"Lambda2": {
										"Type": "Task",
										"Resource": "arn:aws:lambda:us-east-1:123456789012:function:Func2",
										"End": true
									}
								}
							}
						],
						"End": true
					}
				}
			}`,
			expectedDeps: map[string][]string{
				"Parallel": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			deps, err := AnalyzeResourceDependenciesE(tt.definition)
			require.NoError(t, err, "Should analyze dependencies without error")

			for state, expectedDeps := range tt.expectedDeps {
				actualDeps, exists := deps[state]
				assert.True(t, exists, "State %s should exist in dependencies", state)
				assert.ElementsMatch(t, expectedDeps, actualDeps, "Dependencies for %s should match", state)
			}
		})
	}
}

// TestServiceIntegrationMetrics tests metrics collection for service integrations
func TestServiceIntegrationMetrics(t *testing.T) {
	history := []HistoryEvent{
		{
			ID:        1,
			Type:      types.HistoryEventTypeExecutionStarted,
			Timestamp: time.Now().Add(-10 * time.Minute),
		},
		{
			ID:        2,
			Type:      types.HistoryEventTypeTaskStateEntered,
			Timestamp: time.Now().Add(-9 * time.Minute),
			StateEnteredEventDetails: &StateEnteredEventDetails{
				Name: "InvokeLambda",
			},
		},
		{
			ID:        3,
			Type:      types.HistoryEventTypeTaskSucceeded,
			Timestamp: time.Now().Add(-8 * time.Minute),
			TaskSucceededEventDetails: &TaskSucceededEventDetails{
				Resource: "arn:aws:lambda:us-east-1:123456789012:function:TestFunction",
				Output:   `{"result": "success"}`,
			},
		},
		{
			ID:        4,
			Type:      types.HistoryEventTypeExecutionSucceeded,
			Timestamp: time.Now().Add(-7 * time.Minute),
		},
	}

	tests := []struct {
		name                 string
		history              []HistoryEvent
		expectedLambdaCalls  int
		expectedDynamoCalls  int
		expectedSNSCalls     int
		expectedTotalDuration time.Duration
	}{
		{
			name:                "LambdaInvocation",
			history:             history,
			expectedLambdaCalls: 1,
			expectedDynamoCalls: 0,
			expectedSNSCalls:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			metrics := CollectServiceIntegrationMetrics(tt.history)

			assert.Equal(t, tt.expectedLambdaCalls, metrics.LambdaInvocations, "Lambda invocations should match")
			assert.Equal(t, tt.expectedDynamoCalls, metrics.DynamoDBOperations, "DynamoDB operations should match")
			assert.Equal(t, tt.expectedSNSCalls, metrics.SNSPublications, "SNS publications should match")
			
			if tt.expectedTotalDuration > 0 {
				assert.InDelta(t, tt.expectedTotalDuration.Minutes(), metrics.TotalDuration.Minutes(), 1.0, "Duration should be approximately correct")
			}
		})
	}
}

// Test helper functions for service integration scenarios
func createTestServiceIntegrationContext(t *testing.T) *TestContext {
	t.Helper()
	
	return &TestContext{
		T:         &MockT{TestingT: t},
		AwsConfig: aws.Config{Region: "us-east-1"},
		Region:    "us-east-1",
	}
}

func createTestServiceIntegrationInput(serviceType string) *Input {
	switch serviceType {
	case "lambda":
		return NewInput().
			Set("payload", "test data").
			Set("requestId", "test-request-123")
	case "dynamodb":
		return NewInput().
			Set("id", "test-id-123").
			Set("data", "test data").
			Set("timestamp", time.Now().Format(time.RFC3339))
	case "sns":
		return NewInput().
			Set("message", "Test notification").
			Set("subject", "Test Subject")
	default:
		return NewInput()
	}
}

func createTestExecutionResultWithService(service string, status types.ExecutionStatus) *ExecutionResult {
	return &ExecutionResult{
		ExecutionArn:    "arn:aws:states:us-east-1:123456789012:execution:test-" + service + ":execution-123",
		StateMachineArn: "arn:aws:states:us-east-1:123456789012:stateMachine:test-" + service,
		Name:            "test-execution-" + service,
		Status:          status,
		StartDate:       time.Now().Add(-5 * time.Minute),
		StopDate:        time.Now(),
		Input:           `{"test": "data"}`,
		Output:          `{"result": "success"}`,
		ExecutionTime:   5 * time.Minute,
	}
}

// Benchmark tests for service integration performance
func BenchmarkServiceIntegrationValidation(b *testing.B) {
	stateMachine := &StateMachineDefinition{
		Name:       "benchmark-test",
		Definition: createTestComplexWorkflowDefinition(),
		RoleArn:    "arn:aws:iam::123456789012:role/StepFunctionsRole",
		Type:       types.StateMachineTypeStandard,
	}
	input := NewInput().Set("id", "benchmark-id").Set("data", "benchmark data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateComplexWorkflowE(stateMachine, input)
	}
}

func BenchmarkResourceExtraction(b *testing.B) {
	definition := createTestComplexWorkflowDefinition()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtractLambdaFunctions(definition)
		_ = ExtractDynamoDBTables(definition)
		_ = ExtractSNSTopics(definition)
	}
}

func BenchmarkServiceIntegrationMetrics(b *testing.B) {
	history := make([]HistoryEvent, 100)
	for i := 0; i < 100; i++ {
		history[i] = HistoryEvent{
			ID:        int64(i + 1),
			Type:      types.HistoryEventTypeTaskSucceeded,
			Timestamp: time.Now().Add(-time.Duration(100-i) * time.Minute),
			TaskSucceededEventDetails: &TaskSucceededEventDetails{
				Resource: "arn:aws:lambda:us-east-1:123456789012:function:TestFunction",
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CollectServiceIntegrationMetrics(history)
	}
}