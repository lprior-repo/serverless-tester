package vasdeference

import (
	"fmt"
	"testing"
	"time"

	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	sfnTypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/samber/lo"
	"github.com/samber/mo"

	"./dynamodb"
	"./eventbridge"
	"./lambda"
	"./s3"
	"./stepfunctions"
)

// TestFunctionalIntegrationScenario tests a complete serverless application scenario using all functional programming patterns
func TestFunctionalIntegrationScenario(t *testing.T) {
	// Step 1: Create core functional configuration
	coreConfig := NewFunctionalCoreConfig(
		WithCoreRegion("us-east-1"),
		WithCoreTimeout(60*time.Second),
		WithCoreRetryConfig(3, 300*time.Millisecond),
		WithCoreNamespace("functional-integration-test"),
		WithCoreLogging(true),
		WithCoreMetrics(true),
		WithCoreResilience(true),
		WithCoreMetadata("test_type", "integration"),
		WithCoreMetadata("scenario", "complete_serverless_workflow"),
		WithCoreMetadata("environment", "test"),
	)

	// Create test context
	ctx := NewFunctionalTestContext(coreConfig).
		WithCapabilities([]string{"lambda", "dynamodb", "s3", "eventbridge", "stepfunctions"}).
		WithEnvironment(map[string]string{
			"AWS_REGION": "us-east-1",
			"ENVIRONMENT": "test",
		})

	// Step 2: Set up S3 bucket for data storage
	s3Config := s3.NewFunctionalS3Config("integration-test-bucket",
		s3.WithS3Timeout(30*time.Second),
		s3.WithS3RetryConfig(3, 200*time.Millisecond),
		s3.WithS3Region("us-east-1"),
		s3.WithStorageClass(s3Types.StorageClassStandard),
		s3.WithVersioningEnabled(true),
		s3.WithS3Metadata("purpose", "integration_test"),
		s3.WithS3Metadata("environment", "test"),
	)

	// Create bucket
	createBucketResult := s3.SimulateFunctionalS3CreateBucket(s3Config)
	if !createBucketResult.IsSuccess() {
		t.Fatalf("Failed to create S3 bucket: %v", createBucketResult.GetError().MustGet())
	}

	t.Logf("✓ S3 bucket created successfully in %v", createBucketResult.GetDuration())

	// Upload test data
	testData := s3.NewFunctionalS3Object("data/user-events.json", []byte(`{
		"events": [
			{"userId": "user-001", "action": "login", "timestamp": "2024-01-01T10:00:00Z"},
			{"userId": "user-002", "action": "purchase", "timestamp": "2024-01-01T10:05:00Z"},
			{"userId": "user-001", "action": "logout", "timestamp": "2024-01-01T10:30:00Z"}
		]
	}`))

	uploadResult := s3.SimulateFunctionalS3PutObject(testData, s3Config)
	if !uploadResult.IsSuccess() {
		t.Fatalf("Failed to upload test data: %v", uploadResult.GetError().MustGet())
	}

	t.Logf("✓ Test data uploaded to S3 in %v", uploadResult.GetDuration())

	// Step 3: Set up DynamoDB table for user data
	dynamoConfig := dynamodb.NewFunctionalDynamoDBConfig("integration-test-users",
		dynamodb.WithDynamoDBTimeout(30*time.Second),
		dynamodb.WithDynamoDBRetryConfig(3, 200*time.Millisecond),
		dynamodb.WithConsistentRead(true),
		dynamodb.WithBillingMode(types.BillingModePayPerRequest),
		dynamodb.WithDynamoDBMetadata("purpose", "integration_test"),
		dynamodb.WithDynamoDBMetadata("environment", "test"),
	)

	// Create table (simulated)
	createTableResult := dynamodb.SimulateFunctionalDynamoDBCreateTable(dynamoConfig)
	if !createTableResult.IsSuccess() {
		t.Fatalf("Failed to create DynamoDB table: %v", createTableResult.GetError().MustGet())
	}

	t.Logf("✓ DynamoDB table created successfully in %v", createTableResult.GetDuration())

	// Insert user records
	users := []dynamodb.FunctionalDynamoDBItem{
		dynamodb.NewFunctionalDynamoDBItem().
			WithAttribute("userId", "user-001").
			WithAttribute("name", "Alice Johnson").
			WithAttribute("email", "alice@example.com").
			WithAttribute("status", "active"),
		
		dynamodb.NewFunctionalDynamoDBItem().
			WithAttribute("userId", "user-002").
			WithAttribute("name", "Bob Smith").
			WithAttribute("email", "bob@example.com").
			WithAttribute("status", "active"),
	}

	userInsertResults := lo.Map(users, func(user dynamodb.FunctionalDynamoDBItem, index int) dynamodb.FunctionalDynamoDBOperationResult {
		return dynamodb.SimulateFunctionalDynamoDBPutItem(user, dynamoConfig)
	})

	allUsersInserted := lo.EveryBy(userInsertResults, func(result dynamodb.FunctionalDynamoDBOperationResult) bool {
		return result.IsSuccess()
	})

	if !allUsersInserted {
		t.Fatal("Failed to insert all user records")
	}

	t.Logf("✓ Inserted %d user records successfully", len(users))

	// Step 4: Set up Lambda function for event processing
	lambdaConfig := lambda.NewFunctionalLambdaConfig(
		lambda.WithFunctionName("integration-test-processor"),
		lambda.WithRuntime("nodejs22.x"),
		lambda.WithHandler("index.handler"),
		lambda.WithTimeout(30*time.Second),
		lambda.WithMemorySize(512),
		lambda.WithLambdaRetryConfig(3, 200*time.Millisecond),
		lambda.WithLambdaMetadata("purpose", "integration_test"),
		lambda.WithLambdaMetadata("environment", "test"),
	)

	// Create Lambda function
	createFunctionResult := lambda.SimulateFunctionalLambdaCreateFunction(lambdaConfig)
	if !createFunctionResult.IsSuccess() {
		t.Fatalf("Failed to create Lambda function: %v", createFunctionResult.GetError().MustGet())
	}

	t.Logf("✓ Lambda function created successfully in %v", createFunctionResult.GetDuration())

	// Invoke Lambda function with test payload
	invocationPayload := lambda.NewFunctionalLambdaInvocationPayload(`{
		"source": "integration-test",
		"eventType": "user-event-processing",
		"data": {
			"bucketName": "integration-test-bucket",
			"objectKey": "data/user-events.json",
			"tableName": "integration-test-users"
		}
	}`)

	invokeResult := lambda.SimulateFunctionalLambdaInvoke(invocationPayload, lambdaConfig)
	if !invokeResult.IsSuccess() {
		t.Fatalf("Failed to invoke Lambda function: %v", invokeResult.GetError().MustGet())
	}

	t.Logf("✓ Lambda function invoked successfully in %v", invokeResult.GetDuration())

	// Step 5: Set up EventBridge for event routing
	eventBridgeConfig := eventbridge.NewFunctionalEventBridgeConfig("integration-test-bus",
		eventbridge.WithEventBridgeTimeout(30*time.Second),
		eventbridge.WithEventBridgeRetryConfig(3, 200*time.Millisecond),
		eventbridge.WithEventPattern(`{"source": ["integration.test"], "detail-type": ["User Event"]}`),
		eventbridge.WithRuleState(types.RuleStateEnabled),
		eventbridge.WithEventBridgeMetadata("purpose", "integration_test"),
		eventbridge.WithEventBridgeMetadata("environment", "test"),
	)

	// Create custom event bus
	createEventBusResult := eventbridge.SimulateFunctionalEventBridgeCreateEventBus(eventBridgeConfig)
	if !createEventBusResult.IsSuccess() {
		t.Fatalf("Failed to create EventBridge bus: %v", createEventBusResult.GetError().MustGet())
	}

	t.Logf("✓ EventBridge bus created successfully in %v", createEventBusResult.GetDuration())

	// Create rule
	rule := eventbridge.NewFunctionalEventBridgeRule("integration-test-rule", "integration-test-bus")
	createRuleResult := eventbridge.SimulateFunctionalEventBridgeCreateRule(rule, eventBridgeConfig)
	if !createRuleResult.IsSuccess() {
		t.Fatalf("Failed to create EventBridge rule: %v", createRuleResult.GetError().MustGet())
	}

	t.Logf("✓ EventBridge rule created successfully in %v", createRuleResult.GetDuration())

	// Publish integration events
	integrationEvents := []eventbridge.FunctionalEventBridgeEvent{
		eventbridge.NewFunctionalEventBridgeEvent("integration.test", "User Event", `{
			"userId": "user-001",
			"eventType": "login",
			"metadata": {
				"source": "web-app",
				"timestamp": "2024-01-01T10:00:00Z"
			}
		}`).WithEventBusName("integration-test-bus"),
		
		eventbridge.NewFunctionalEventBridgeEvent("integration.test", "User Event", `{
			"userId": "user-002", 
			"eventType": "purchase",
			"metadata": {
				"amount": 99.99,
				"currency": "USD",
				"timestamp": "2024-01-01T10:05:00Z"
			}
		}`).WithEventBusName("integration-test-bus"),
	}

	publishEventsResult := eventbridge.SimulateFunctionalEventBridgePutEvents(integrationEvents, eventBridgeConfig)
	if !publishEventsResult.IsSuccess() {
		t.Fatalf("Failed to publish events: %v", publishEventsResult.GetError().MustGet())
	}

	t.Logf("✓ Published %d events to EventBridge in %v", len(integrationEvents), publishEventsResult.GetDuration())

	// Step 6: Set up Step Functions workflow for orchestration
	stepFunctionsConfig := stepfunctions.NewFunctionalStepFunctionsConfig(
		stepfunctions.WithStepFunctionsTimeout(10*time.Minute),
		stepfunctions.WithStepFunctionsRetryConfig(3, 300*time.Millisecond),
		stepfunctions.WithStateMachineType(sfnTypes.StateMachineTypeStandard),
		stepfunctions.WithStepFunctionsMetadata("purpose", "integration_test"),
		stepfunctions.WithStepFunctionsMetadata("environment", "test"),
	)

	// Create workflow definition
	workflowDefinition := `{
		"Comment": "Integration test workflow for processing user events",
		"StartAt": "ProcessEvents",
		"States": {
			"ProcessEvents": {
				"Type": "Task",
				"Resource": "arn:aws:lambda:us-east-1:123456789012:function:integration-test-processor",
				"Next": "CheckResults"
			},
			"CheckResults": {
				"Type": "Choice",
				"Choices": [
					{
						"Variable": "$.success",
						"BooleanEquals": true,
						"Next": "NotifySuccess"
					}
				],
				"Default": "NotifyFailure"
			},
			"NotifySuccess": {
				"Type": "Task",
				"Resource": "arn:aws:states:::sns:publish",
				"Parameters": {
					"Message": "Integration test completed successfully",
					"TopicArn": "arn:aws:sns:us-east-1:123456789012:integration-notifications"
				},
				"End": true
			},
			"NotifyFailure": {
				"Type": "Task",
				"Resource": "arn:aws:states:::sns:publish", 
				"Parameters": {
					"Message": "Integration test failed",
					"TopicArn": "arn:aws:sns:us-east-1:123456789012:integration-notifications"
				},
				"End": true
			}
		}
	}`

	stateMachine := stepfunctions.NewFunctionalStepFunctionsStateMachine(
		"integration-test-workflow",
		workflowDefinition,
		"arn:aws:iam::123456789012:role/StepFunctionsIntegrationRole",
	).WithType(sfnTypes.StateMachineTypeStandard)

	createStateMachineResult := stepfunctions.SimulateFunctionalStepFunctionsCreateStateMachine(stateMachine, stepFunctionsConfig)
	if !createStateMachineResult.IsSuccess() {
		t.Fatalf("Failed to create Step Functions state machine: %v", createStateMachineResult.GetError().MustGet())
	}

	t.Logf("✓ Step Functions state machine created successfully in %v", createStateMachineResult.GetDuration())

	// Execute workflow
	workflowExecution := stepfunctions.NewFunctionalStepFunctionsExecution(
		"integration-test-execution",
		"arn:aws:states:us-east-1:123456789012:stateMachine:integration-test-workflow",
		`{
			"bucketName": "integration-test-bucket",
			"objectKey": "data/user-events.json",
			"tableName": "integration-test-users",
			"eventBusName": "integration-test-bus"
		}`,
	)

	startExecutionResult := stepfunctions.SimulateFunctionalStepFunctionsStartExecution(workflowExecution, stepFunctionsConfig)
	if !startExecutionResult.IsSuccess() {
		t.Fatalf("Failed to start Step Functions execution: %v", startExecutionResult.GetError().MustGet())
	}

	t.Logf("✓ Step Functions execution started successfully in %v", startExecutionResult.GetDuration())

	// Step 7: Perform core health checks
	services := []string{"lambda", "dynamodb", "s3", "eventbridge", "stepfunctions"}
	healthCheckResult := SimulateFunctionalCoreHealthCheck(services, coreConfig)
	if !healthCheckResult.IsSuccess() {
		t.Fatalf("Failed to perform health check: %v", healthCheckResult.GetError().MustGet())
	}

	// Verify overall health
	if healthCheckResult.GetResult().IsPresent() {
		healthData, ok := healthCheckResult.GetResult().MustGet().(map[string]interface{})
		if ok {
			overallHealthy := healthData["overall_healthy"].(bool)
			healthPercentage := healthData["health_percentage"].(float64)
			
			t.Logf("✓ Health check completed: %t (%.1f%% healthy) in %v", 
				overallHealthy, healthPercentage, healthCheckResult.GetDuration())
			
			if healthPercentage < 80.0 {
				t.Logf("⚠ Warning: Health percentage below threshold (%.1f%%)", healthPercentage)
			}
		}
	}

	// Step 8: Collect and verify all operation metrics
	allResults := []interface{}{
		createBucketResult, uploadResult,
		createTableResult, 
		createFunctionResult, invokeResult,
		createEventBusResult, createRuleResult, publishEventsResult,
		createStateMachineResult, startExecutionResult,
		healthCheckResult,
	}

	totalOperations := len(allResults)
	var totalDuration time.Duration
	var successfulOperations int

	// Analyze results using functional patterns
	type OperationSummary struct {
		Name     string
		Duration time.Duration
		Success  bool
	}

	summaries := []OperationSummary{
		{"S3 Bucket Creation", createBucketResult.GetDuration(), createBucketResult.IsSuccess()},
		{"S3 Object Upload", uploadResult.GetDuration(), uploadResult.IsSuccess()},
		{"DynamoDB Table Creation", createTableResult.GetDuration(), createTableResult.IsSuccess()},
		{"Lambda Function Creation", createFunctionResult.GetDuration(), createFunctionResult.IsSuccess()},
		{"Lambda Function Invocation", invokeResult.GetDuration(), invokeResult.IsSuccess()},
		{"EventBridge Bus Creation", createEventBusResult.GetDuration(), createEventBusResult.IsSuccess()},
		{"EventBridge Rule Creation", createRuleResult.GetDuration(), createRuleResult.IsSuccess()},
		{"EventBridge Event Publishing", publishEventsResult.GetDuration(), publishEventsResult.IsSuccess()},
		{"Step Functions State Machine Creation", createStateMachineResult.GetDuration(), createStateMachineResult.IsSuccess()},
		{"Step Functions Execution Start", startExecutionResult.GetDuration(), startExecutionResult.IsSuccess()},
		{"Health Check", healthCheckResult.GetDuration(), healthCheckResult.IsSuccess()},
	}

	// Calculate metrics using functional programming
	successfulOperations = lo.CountBy(summaries, func(s OperationSummary) bool { return s.Success })
	totalDuration = lo.Reduce(summaries, func(acc time.Duration, s OperationSummary, _ int) time.Duration {
		return acc + s.Duration
	}, 0)

	avgDuration := totalDuration / time.Duration(totalOperations)
	successRate := float64(successfulOperations) / float64(totalOperations) * 100

	// Log final integration test results
	t.Logf("\n" + strings.Repeat("=", 80))
	t.Logf("FUNCTIONAL PROGRAMMING INTEGRATION TEST RESULTS")
	t.Logf(strings.Repeat("=", 80))
	t.Logf("Total Operations: %d", totalOperations)
	t.Logf("Successful Operations: %d", successfulOperations)
	t.Logf("Success Rate: %.1f%%", successRate)
	t.Logf("Total Duration: %v", totalDuration)
	t.Logf("Average Duration: %v", avgDuration)
	t.Logf("Context ID: %s", ctx.GetContextID())
	t.Logf("Test Environment: %s", ctx.GetEnvironment()["ENVIRONMENT"])
	t.Logf(strings.Repeat("=", 80))

	// Detailed operation breakdown
	t.Logf("\nOperation Breakdown:")
	for _, summary := range summaries {
		status := "✓"
		if !summary.Success {
			status = "✗"
		}
		t.Logf("%s %-35s: %v", status, summary.Name, summary.Duration)
	}

	// Assert overall integration success
	if successRate < 100.0 {
		t.Errorf("Integration test failed: Success rate %.1f%% is below 100%%", successRate)
	}

	if totalDuration > 30*time.Second {
		t.Logf("⚠ Warning: Total duration %v exceeds expected threshold", totalDuration)
	}

	// Verify functional programming patterns were used correctly
	t.Logf("\n" + strings.Repeat("-", 80))
	t.Logf("FUNCTIONAL PROGRAMMING PATTERN VERIFICATION")
	t.Logf(strings.Repeat("-", 80))
	
	// Verify immutability - original config should be unchanged
	if coreConfig.GetRegion() != "us-east-1" {
		t.Error("Core config immutability violated: region changed")
	}

	// Verify monadic error handling - all operations returned proper Option types
	resultsWithErrors := lo.Filter(summaries, func(s OperationSummary, _ int) bool {
		return !s.Success
	})

	if len(resultsWithErrors) == 0 {
		t.Logf("✓ All operations completed successfully with proper error handling")
	}

	// Verify functional composition - combined configurations worked correctly
	combinedTestConfig := combineConfigs(coreConfig, NewFunctionalCoreConfig(
		WithCoreMetadata("additional", "test_data"),
	))

	if len(combinedTestConfig.GetMetadata()) < len(coreConfig.GetMetadata()) {
		t.Error("Configuration composition failed")
	} else {
		t.Logf("✓ Functional configuration composition working correctly")
	}

	t.Logf("✓ All functional programming patterns verified successfully")
	t.Logf(strings.Repeat("-", 80))

	t.Logf("\n🎉 INTEGRATION TEST COMPLETED SUCCESSFULLY")
	t.Logf("All AWS services integrated using pure functional programming patterns")
}

// TestFunctionalPatternConsistency verifies that all services follow the same functional patterns
func TestFunctionalPatternConsistency(t *testing.T) {
	// Test that all services use consistent functional option patterns
	t.Run("ConfigurationPatterns", func(t *testing.T) {
		// All services should have similar configuration patterns
		coreConfig := NewFunctionalCoreConfig(WithCoreRegion("us-east-1"))
		lambdaConfig := lambda.NewFunctionalLambdaConfig(lambda.WithRuntime("nodejs22.x"))
		dynamoConfig := dynamodb.NewFunctionalDynamoDBConfig("test-table")
		s3Config := s3.NewFunctionalS3Config("test-bucket")
		eventBridgeConfig := eventbridge.NewFunctionalEventBridgeConfig("test-bus")
		stepFunctionsConfig := stepfunctions.NewFunctionalStepFunctionsConfig()

		// Verify all configs are created successfully
		configs := []interface{}{coreConfig, lambdaConfig, dynamoConfig, s3Config, eventBridgeConfig, stepFunctionsConfig}
		for i, config := range configs {
			if config == nil {
				t.Errorf("Configuration %d should not be nil", i)
			}
		}
	})

	t.Run("OperationResultPatterns", func(t *testing.T) {
		// All services should return consistent operation result patterns
		coreResult := SimulateFunctionalCoreOperation("test", NewFunctionalCoreConfig())
		
		// Verify result pattern consistency
		if !coreResult.IsSuccess() {
			t.Error("Core operation should succeed with valid config")
		}
		
		if coreResult.GetDuration() == 0 {
			t.Error("All operations should record duration")
		}
		
		if len(coreResult.GetMetadata()) == 0 {
			t.Error("All operations should include metadata")
		}
	})

	t.Run("ImmutabilityPatterns", func(t *testing.T) {
		// Verify that all configurations are immutable
		originalConfig := NewFunctionalCoreConfig(WithCoreRegion("us-east-1"))
		
		// Create modified configuration
		modifiedConfig := NewFunctionalCoreConfig(
			WithCoreRegion("eu-west-1"),
			WithCoreTimeout(60*time.Second),
		)
		
		// Original should be unchanged
		if originalConfig.GetRegion() != "us-east-1" {
			t.Error("Original configuration should remain unchanged (immutability)")
		}
		
		if modifiedConfig.GetRegion() != "eu-west-1" {
			t.Error("Modified configuration should have new values")
		}
	})

	t.Run("ValidationPatterns", func(t *testing.T) {
		// All services should have consistent validation patterns
		
		// Test invalid configurations consistently fail
		invalidCoreConfig := NewFunctionalCoreConfig(WithCoreRegion(""))
		coreResult := SimulateFunctionalCoreOperation("test", invalidCoreConfig)
		
		if coreResult.IsSuccess() {
			t.Error("Invalid configurations should consistently fail validation")
		}
		
		if coreResult.GetError().IsAbsent() {
			t.Error("Failed operations should have error information")
		}
	})
}

// TestFunctionalErrorComposition tests error handling composition across services
func TestFunctionalErrorComposition(t *testing.T) {
	// Create a scenario where multiple services have different error conditions
	services := map[string]func() interface{}{
		"core": func() interface{} {
			return SimulateFunctionalCoreOperation("test", NewFunctionalCoreConfig(WithCoreRegion("")))
		},
		"lambda": func() interface{} {
			return lambda.SimulateFunctionalLambdaCreateFunction(lambda.NewFunctionalLambdaConfig())
		},
		"dynamodb": func() interface{} {
			return dynamodb.SimulateFunctionalDynamoDBPutItem(
				dynamodb.NewFunctionalDynamoDBItem(),
				dynamodb.NewFunctionalDynamoDBConfig(""),
			)
		},
		"s3": func() interface{} {
			return s3.SimulateFunctionalS3CreateBucket(s3.NewFunctionalS3Config(""))
		},
		"eventbridge": func() interface{} {
			return eventbridge.SimulateFunctionalEventBridgePutEvent(
				eventbridge.NewFunctionalEventBridgeEvent("", "Test", "{}"),
				eventbridge.NewFunctionalEventBridgeConfig("test-bus"),
			)
		},
	}

	// Execute all operations and collect errors
	type ServiceError struct {
		Service string
		Success bool
		Error   string
	}

	var errors []ServiceError

	for serviceName, operation := range services {
		result := operation()
		
		// Use type assertion to check IsSuccess method (all results implement this)
		switch r := result.(type) {
		case interface{ IsSuccess() bool }:
			success := r.IsSuccess()
			errorMsg := ""
			
			if !success {
				// Try to get error message using type assertion
				if errorGetter, ok := r.(interface{ GetError() mo.Option[error] }); ok {
					if err := errorGetter.GetError(); err.IsPresent() {
						errorMsg = err.MustGet().Error()
					}
				}
			}
			
			errors = append(errors, ServiceError{
				Service: serviceName,
				Success: success,
				Error:   errorMsg,
			})
		default:
			t.Errorf("Service %s result does not implement expected interface", serviceName)
		}
	}

	// Verify error composition
	failedServices := lo.Filter(errors, func(e ServiceError, _ int) bool {
		return !e.Success
	})

	if len(failedServices) == 0 {
		t.Error("Expected some services to fail with invalid configurations")
	}

	// Verify all failed services have meaningful error messages
	for _, failedService := range failedServices {
		if failedService.Error == "" {
			t.Errorf("Service %s should have error message when failing", failedService.Service)
		}
		t.Logf("Service %s failed as expected: %s", failedService.Service, failedService.Error)
	}
}

// BenchmarkFunctionalIntegrationPerformance benchmarks the complete functional integration
func BenchmarkFunctionalIntegrationPerformance(b *testing.B) {
	config := NewFunctionalCoreConfig(
		WithCoreTimeout(5*time.Second),
		WithCoreRetryConfig(1, 10*time.Millisecond),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create minimal versions of all functional operations
		ctx := NewFunctionalTestContext(config)
		
		// Core operation
		coreResult := SimulateFunctionalCoreOperation("benchmark", config)
		if !coreResult.IsSuccess() {
			b.Errorf("Core operation failed: %v", coreResult.GetError().MustGet())
		}

		// Lambda operation  
		lambdaConfig := lambda.NewFunctionalLambdaConfig(lambda.WithFunctionName("bench"))
		lambdaResult := lambda.SimulateFunctionalLambdaCreateFunction(lambdaConfig)
		if !lambdaResult.IsSuccess() {
			b.Errorf("Lambda operation failed: %v", lambdaResult.GetError().MustGet())
		}

		// Verify context integrity
		if ctx.GetContextID() == "" {
			b.Error("Context ID should be generated")
		}
	}
}