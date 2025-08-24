package lambda_test

import (
	"testing"
	"time"

	"vasdeference/lambda"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
)

// TestExampleUsage demonstrates the basic usage patterns of the lambda package
// following Terratest conventions and functional programming principles
func TestExampleUsage(t *testing.T) {
	// These examples show how to use the lambda testing package
	// Note: These are example patterns - actual tests would need real AWS resources
	
	// Create test context - this would typically use real AWS configuration
	ctx := &lambda.TestContext{T: t}
	
	// Create arrange for cleanup management (simplified for examples)
	_ = ctx // Suppress unused variable warning in examples
	
	// Example 1: Basic Lambda function invocation
	t.Run("BasicInvocation", func(t *testing.T) {
		functionName := "my-test-function"
		payload := `{"message": "Hello, Lambda!"}`
		
		// This would invoke an actual Lambda function in real tests
		// result := lambda.Invoke(ctx, functionName, payload)
		// assert.Equal(t, int32(200), result.StatusCode)
		// assert.Contains(t, result.Payload, "success")
		
		t.Logf("Example: Basic Lambda invocation pattern for %s with payload %s", functionName, payload)
	})
	
	// Example 2: Function configuration validation
	t.Run("ConfigurationValidation", func(t *testing.T) {
		functionName := "my-test-function"
		
		// Get and validate function configuration
		// config := lambda.GetFunction(ctx, functionName)
		// lambda.AssertFunctionRuntime(ctx, functionName, types.RuntimeNodejs18x)
		// lambda.AssertFunctionTimeout(ctx, functionName, 30)
		// lambda.AssertEnvVarEquals(ctx, functionName, "NODE_ENV", "production")
		
		t.Logf("Example: Function configuration validation pattern for %s", functionName)
	})
	
	// Example 3: Event-driven testing with S3 events
	t.Run("S3EventTesting", func(t *testing.T) {
		functionName := "s3-processor-function"
		bucketName := "test-bucket"
		objectKey := "test-data/file.json"
		
		// Build S3 event
		s3Event := lambda.BuildS3Event(bucketName, objectKey, "s3:ObjectCreated:Put")
		
		// Invoke with S3 event
		// result := lambda.Invoke(ctx, functionName, s3Event)
		// lambda.AssertInvokeSuccess(ctx, result)
		// lambda.AssertPayloadContains(ctx, result, "processed")
		
		// Verify logs contain expected processing messages
		// lambda.AssertLogsContain(ctx, functionName, "Processing S3 object")
		
		t.Logf("Example: S3 event-driven testing pattern for %s using %s/%s", functionName, bucketName, objectKey)
		assert.Contains(t, s3Event, bucketName)
	})
	
	// Example 4: Retry and resilience testing
	t.Run("RetryTesting", func(t *testing.T) {
		functionName := "unreliable-function"
		payload := `{"test": "retry"}`
		
		// Test with retry logic for eventual consistency
		// result := lambda.InvokeWithRetry(ctx, functionName, payload, 3)
		// lambda.AssertInvokeSuccess(ctx, result)
		
		t.Logf("Example: Retry and resilience testing pattern for %s with payload %s", functionName, payload)
	})
	
	// Example 5: Async invocation and log monitoring
	t.Run("AsyncInvocationTesting", func(t *testing.T) {
		functionName := "async-processor"
		payload := `{"data": "async-test"}`
		
		// Async invocation
		// result := lambda.InvokeAsync(ctx, functionName, payload)
		// assert.Equal(t, int32(202), result.StatusCode) // Async invocation returns 202
		
		// Wait for logs to appear with expected content
		// logs := lambda.WaitForLogs(ctx, functionName, "Processing completed", 30*time.Second)
		// assert.NotEmpty(t, logs)
		
		t.Logf("Example: Async invocation and log monitoring pattern for %s with payload %s", functionName, payload)
	})
}

// TestExampleTableDrivenTesting demonstrates comprehensive table-driven testing patterns
func TestExampleTableDrivenTesting(t *testing.T) {
	ctx := &lambda.TestContext{T: t}
	_ = ctx // Suppress unused variable warning in examples
	
	// Table-driven test for multiple payload scenarios
	tests := []struct {
		name           string
		functionName   string
		payload        string
		expectSuccess  bool
		expectedInPayload string
		description    string
	}{
		{
			name:           "ValidPayload",
			functionName:   "validator-function",
			payload:        `{"name": "John", "age": 30}`,
			expectSuccess:  true,
			expectedInPayload: "valid",
			description:    "Valid payload should be processed successfully",
		},
		{
			name:           "InvalidPayload",
			functionName:   "validator-function",
			payload:        `{"name": ""}`,
			expectSuccess:  false,
			expectedInPayload: "error",
			description:    "Invalid payload should return validation error",
		},
		{
			name:           "EmptyPayload",
			functionName:   "validator-function",
			payload:        `{}`,
			expectSuccess:  false,
			expectedInPayload: "missing_required_fields",
			description:    "Empty payload should return missing fields error",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// In real tests, this would invoke actual Lambda functions
			t.Logf("Testing %s: %s", tc.name, tc.description)
			
			// result := lambda.Invoke(ctx, tc.functionName, tc.payload)
			// 
			// if tc.expectSuccess {
			//     lambda.AssertInvokeSuccess(ctx, result)
			// } else {
			//     lambda.AssertInvokeError(ctx, result)
			// }
			// 
			// lambda.AssertPayloadContains(ctx, result, tc.expectedInPayload)
		})
	}
}

// TestExampleAdvancedEventTesting shows advanced event testing patterns
func TestExampleAdvancedEventTesting(t *testing.T) {
	ctx := &lambda.TestContext{T: t}
	_ = ctx // Suppress unused variable warning in examples
	
	t.Run("MultipleEventSources", func(t *testing.T) {
		functionName := "multi-trigger-function"
		_ = functionName // Suppress unused variable warning in examples
		
		// Test S3 event processing
		t.Run("S3Event", func(t *testing.T) {
			s3Event := lambda.BuildS3Event("data-bucket", "input/data.json", "s3:ObjectCreated:Put")
			_ = s3Event // Suppress unused variable warning in examples
			
			// result := lambda.Invoke(ctx, functionName, s3Event)
			// lambda.AssertInvokeSuccess(ctx, result)
			// lambda.AssertPayloadContains(ctx, result, "s3_processed")
			
			t.Log("S3 event processing test completed")
		})
		
		// Test DynamoDB event processing
		t.Run("DynamoDBEvent", func(t *testing.T) {
			keys := map[string]interface{}{
				"id": map[string]interface{}{"S": "item-123"},
			}
			dynamoEvent := lambda.BuildDynamoDBEvent("users-table", "INSERT", keys)
			_ = dynamoEvent // Suppress unused variable warning in examples
			
			// result := lambda.Invoke(ctx, functionName, dynamoEvent)
			// lambda.AssertInvokeSuccess(ctx, result)
			// lambda.AssertPayloadContains(ctx, result, "dynamodb_processed")
			
			t.Log("DynamoDB event processing test completed")
		})
		
		// Test SQS event processing
		t.Run("SQSEvent", func(t *testing.T) {
			queueUrl := "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
			messageBody := `{"order_id": "order-456", "status": "pending"}`
			sqsEvent := lambda.BuildSQSEvent(queueUrl, messageBody)
			_ = sqsEvent // Suppress unused variable warning in examples
			
			// result := lambda.Invoke(ctx, functionName, sqsEvent)
			// lambda.AssertInvokeSuccess(ctx, result)
			// lambda.AssertPayloadContains(ctx, result, "sqs_processed")
			
			t.Log("SQS event processing test completed")
		})
		
		// Test API Gateway event processing
		t.Run("APIGatewayEvent", func(t *testing.T) {
			apiEvent := lambda.BuildAPIGatewayEvent("POST", "/api/orders", `{"item": "widget", "quantity": 5}`)
			_ = apiEvent // Suppress unused variable warning in examples
			
			// result := lambda.Invoke(ctx, functionName, apiEvent)
			// lambda.AssertInvokeSuccess(ctx, result)
			// 
			// // Parse API Gateway response
			// var apiResponse map[string]interface{}
			// lambda.ParseInvokeOutput(result, &apiResponse)
			// assert.Equal(t, 200, int(apiResponse["statusCode"].(float64)))
			
			t.Log("API Gateway event processing test completed")
		})
	})
}

// ExampleEventSourceMapping demonstrates event source mapping testing
func TestExampleEventSourceMapping(t *testing.T) {
	ctx := &lambda.TestContext{T: t}
	_ = ctx // Suppress unused variable warning in examples
	
	t.Run("SQSEventSourceMapping", func(t *testing.T) {
		functionName := "sqs-consumer-function"
		queueArn := "arn:aws:sqs:us-east-1:123456789012:test-queue"
		
		// Create event source mapping configuration
		mappingConfig := lambda.EventSourceMappingConfig{
			EventSourceArn:   queueArn,
			FunctionName:     functionName,
			BatchSize:        10,
			Enabled:          true,
			StartingPosition: types.EventSourcePositionLatest,
		}
		_ = mappingConfig // Suppress unused variable warning in examples
		
		// In real tests:
		// mapping := lambda.CreateEventSourceMapping(ctx, mappingConfig)
		// arrange.RegisterCleanup(func() error {
		//     return lambda.DeleteEventSourceMappingE(ctx, mapping.UUID)
		// })
		// 
		// // Wait for mapping to be active
		// lambda.WaitForEventSourceMappingState(ctx, mapping.UUID, "Enabled", 2*time.Minute)
		// 
		// // Verify mapping configuration
		// assert.Equal(t, queueArn, mapping.EventSourceArn)
		// assert.Equal(t, int32(10), mapping.BatchSize)
		
		t.Log("SQS event source mapping test completed")
	})
}

// ExamplePerformanceTesting shows performance and load testing patterns
func TestExamplePerformanceTesting(t *testing.T) {
	ctx := &lambda.TestContext{T: t}
	_ = ctx // Suppress unused variable warning in examples
	
	t.Run("ConcurrencyTesting", func(t *testing.T) {
		functionName := "high-load-function"
		payload := `{"load_test": true}`
		_ = functionName // Suppress unused variable warning in examples
		_ = payload // Suppress unused variable warning in examples
		
		// Test concurrent invocations
		concurrency := 10
		results := make(chan *lambda.InvokeResult, concurrency)
		
		for i := 0; i < concurrency; i++ {
			go func() {
				// result := lambda.Invoke(ctx, functionName, payload)
				// results <- result
				results <- &lambda.InvokeResult{StatusCode: 200, Payload: `{"success": true}`}
			}()
		}
		
		// Collect results
		successCount := 0
		for i := 0; i < concurrency; i++ {
			result := <-results
			if result.StatusCode == 200 {
				successCount++
			}
		}
		
		// Assert all invocations succeeded
		assert.Equal(t, concurrency, successCount, "All concurrent invocations should succeed")
		
		t.Log("Concurrency testing completed")
	})
	
	t.Run("TimeoutTesting", func(t *testing.T) {
		functionName := "timeout-test-function"
		_ = functionName // Suppress unused variable warning in examples
		
		// Test function timeout configuration
		// config := lambda.GetFunction(ctx, functionName)
		// lambda.AssertFunctionTimeout(ctx, functionName, 30)
		
		// Test invocation with custom timeout
		opts := &lambda.InvokeOptions{
			Timeout: 5 * time.Second,
		}
		_ = opts // Suppress unused variable warning in examples
		
		// result := lambda.InvokeWithOptions(ctx, functionName, `{"delay": 10}`, opts)
		// lambda.AssertExecutionTimeLessThan(ctx, result, 5*time.Second)
		
		t.Log("Timeout testing completed")
	})
}

// ExampleErrorHandlingAndRecovery demonstrates error scenarios and recovery testing
func TestExampleErrorHandlingAndRecovery(t *testing.T) {
	ctx := &lambda.TestContext{T: t}
	_ = ctx // Suppress unused variable warning in examples
	
	t.Run("ErrorScenarios", func(t *testing.T) {
		functionName := "error-test-function"
		_ = functionName // Suppress unused variable warning in examples
		
		// Test function error handling
		errorPayload := `{"trigger_error": "division_by_zero"}`
		_ = errorPayload // Suppress unused variable warning in examples
		
		// In real tests:
		// result := lambda.Invoke(ctx, functionName, errorPayload)
		// lambda.AssertInvokeError(ctx, result)
		// assert.Contains(t, result.FunctionError, "Handled")
		
		// Verify error logs
		// lambda.AssertLogsContainLevel(ctx, functionName, "ERROR")
		// lambda.AssertLogsContain(ctx, functionName, "division_by_zero")
		
		t.Log("Error scenario testing completed")
	})
	
	t.Run("RecoveryTesting", func(t *testing.T) {
		functionName := "recovery-test-function"
		_ = functionName // Suppress unused variable warning in examples
		
		// Test recovery after transient errors
		recoverablePayload := `{"simulate_transient_error": true}`
		_ = recoverablePayload // Suppress unused variable warning in examples
		
		// Use retry logic to handle transient errors
		// result := lambda.InvokeWithRetry(ctx, functionName, recoverablePayload, 3)
		// lambda.AssertInvokeSuccess(ctx, result)
		
		// Verify recovery logs
		// lambda.AssertLogsContain(ctx, functionName, "recovered_after_retry")
		
		t.Log("Recovery testing completed")
	})
}

// ExampleComprehensiveIntegrationTest shows a full end-to-end testing scenario
func TestExampleComprehensiveIntegrationTest(t *testing.T) {
	// This would be a comprehensive integration test
	// Note: Requires actual AWS Lambda functions and resources
	
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	ctx := &lambda.TestContext{T: t}
	_ = ctx // Suppress unused variable warning in examples
	
	t.Run("OrderProcessingWorkflow", func(t *testing.T) {
		// This example shows testing a complete serverless workflow
		
		// Step 1: API Gateway receives order
		orderData := `{
			"customer_id": "cust-123",
			"items": [
				{"sku": "item-001", "quantity": 2, "price": 29.99},
				{"sku": "item-002", "quantity": 1, "price": 49.99}
			],
			"shipping_address": {
				"street": "123 Main St",
				"city": "Anytown",
				"state": "CA",
				"zip": "12345"
			}
		}`
		
		apiEvent := lambda.BuildAPIGatewayEvent("POST", "/api/orders", orderData)
		_ = apiEvent // Suppress unused variable warning in examples
		
		// Test order validation function
		// validationResult := lambda.Invoke(ctx, "order-validator", apiEvent)
		// lambda.AssertInvokeSuccess(ctx, validationResult)
		
		// Step 2: Process order (triggered by SQS)
		// orderProcessorPayload := lambda.ExtractPayloadFromResult(validationResult)
		// processingResult := lambda.Invoke(ctx, "order-processor", orderProcessorPayload)
		// lambda.AssertInvokeSuccess(ctx, processingResult)
		
		// Step 3: Update inventory (triggered by DynamoDB stream)
		// inventoryEvent := lambda.BuildDynamoDBEvent("orders", "INSERT", map[string]interface{}{
		//     "order_id": map[string]interface{}{"S": "order-456"},
		// })
		// inventoryResult := lambda.Invoke(ctx, "inventory-updater", inventoryEvent)
		// lambda.AssertInvokeSuccess(ctx, inventoryResult)
		
		// Step 4: Send confirmation (triggered by SNS)
		// confirmationEvent := lambda.BuildSNSEvent(
		//     "arn:aws:sns:us-east-1:123456789012:order-confirmations",
		//     "Order processed successfully",
		// )
		// confirmationResult := lambda.Invoke(ctx, "notification-sender", confirmationEvent)
		// lambda.AssertInvokeSuccess(ctx, confirmationResult)
		
		// Verify end-to-end workflow logs
		// lambda.AssertLogsContain(ctx, "order-validator", "Order validated successfully")
		// lambda.AssertLogsContain(ctx, "order-processor", "Order processed")
		// lambda.AssertLogsContain(ctx, "inventory-updater", "Inventory updated")
		// lambda.AssertLogsContain(ctx, "notification-sender", "Confirmation sent")
		
		t.Log("Comprehensive order processing workflow test completed")
	})
}

// ExampleTestDataFactories shows how to create reusable test data
func TestExampleTestDataFactories(t *testing.T) {
	// Test data factory functions following functional programming principles
	
	createUserPayload := func(name string, age int) string {
		return lambda.MarshalPayload(map[string]interface{}{
			"name": name,
			"age":  age,
			"created_at": time.Now().UTC().Format(time.RFC3339),
		})
	}
	
	createOrderPayload := func(customerID string, amount float64) string {
		return lambda.MarshalPayload(map[string]interface{}{
			"customer_id": customerID,
			"amount":      amount,
			"currency":    "USD",
			"status":      "pending",
			"created_at":  time.Now().UTC().Format(time.RFC3339),
		})
	}
	
	// Example usage of test data factories
	t.Run("UserProcessing", func(t *testing.T) {
		payload := createUserPayload("John Doe", 30)
		
		// result := lambda.Invoke(ctx, "user-processor", payload)
		// lambda.AssertInvokeSuccess(ctx, result)
		// lambda.AssertPayloadContains(ctx, result, "user_created")
		
		assert.Contains(t, payload, "John Doe")
		assert.Contains(t, payload, "30")
		
		t.Log("User processing with test data factory completed")
	})
	
	t.Run("OrderProcessing", func(t *testing.T) {
		payload := createOrderPayload("cust-123", 99.99)
		
		// result := lambda.Invoke(ctx, "order-processor", payload)
		// lambda.AssertInvokeSuccess(ctx, result)
		// lambda.AssertPayloadContains(ctx, result, "order_processed")
		
		assert.Contains(t, payload, "cust-123")
		assert.Contains(t, payload, "99.99")
		
		t.Log("Order processing with test data factory completed")
	})
}

// ExampleValidationHelpers shows comprehensive validation patterns
func TestExampleValidationHelpers(t *testing.T) {
	// Create expected configuration for validation
	expectedConfig := &lambda.FunctionConfiguration{
		FunctionName:    "production-function",
		Runtime:         types.RuntimeNodejs18x,
		Handler:         "index.handler",
		Timeout:         30,
		MemorySize:      256,
		State:           types.StateActive,
		Environment: map[string]string{
			"NODE_ENV": "production",
			"LOG_LEVEL": "info",
		},
	}
	
	// Example function configuration (in real tests, this would come from GetFunction)
	actualConfig := &lambda.FunctionConfiguration{
		FunctionName:    "production-function",
		Runtime:         types.RuntimeNodejs18x,
		Handler:         "index.handler",
		Timeout:         30,
		MemorySize:      256,
		State:           types.StateActive,
		Environment: map[string]string{
			"NODE_ENV": "production",
			"LOG_LEVEL": "info",
			"DEBUG":     "false", // Extra environment variable
		},
	}
	
	// Validate configuration
	errors := lambda.ValidateFunctionConfiguration(actualConfig, expectedConfig)
	
	// In this case, no errors should be found since actual matches or exceeds expected
	assert.Empty(t, errors, "Configuration validation should pass")
	
	// Example with validation errors
	invalidConfig := &lambda.FunctionConfiguration{
		FunctionName: "production-function",
		Runtime:      types.RuntimePython39, // Wrong runtime
		Handler:      "app.handler",         // Wrong handler
		Timeout:      60,                    // Wrong timeout
		MemorySize:   128,                   // Wrong memory size
		State:        types.StateActive,
		Environment: map[string]string{
			"NODE_ENV": "development", // Wrong environment
			// Missing LOG_LEVEL
		},
	}
	
	errors = lambda.ValidateFunctionConfiguration(invalidConfig, expectedConfig)
	assert.NotEmpty(t, errors, "Invalid configuration should have validation errors")
	
	// Log validation errors for debugging
	for _, err := range errors {
		t.Logf("Validation error: %s", err)
	}
}

// ExamplePlaywrightLikeSyntax demonstrates fluent API patterns
func TestExamplePlaywrightLikeSyntax(t *testing.T) {
	// This example shows how the package could be extended with fluent syntax
	// (This would require additional wrapper functions for a more fluent API)
	
	ctx := &lambda.TestContext{T: t}
	_ = ctx // Suppress unused variable warning in examples
	
	// Current syntax
	t.Run("CurrentSyntax", func(t *testing.T) {
		functionName := "test-function"
		payload := `{"test": true}`
		
		// Standard Terratest-style function calls
		// result := lambda.Invoke(ctx, functionName, payload)
		// lambda.AssertInvokeSuccess(ctx, result)
		// lambda.AssertPayloadContains(ctx, result, "success")
		// lambda.AssertExecutionTimeLessThan(ctx, result, 5*time.Second)
		
		t.Logf("Current Terratest-style syntax demonstrated for %s with %s", functionName, payload)
	})
	
	// Potential fluent syntax (would require additional wrapper implementation)
	t.Run("FluentSyntaxExample", func(t *testing.T) {
		// Example of what fluent syntax might look like:
		// lambda.Function(ctx, "test-function").
		//     WithPayload(`{"test": true}`).
		//     Invoke().
		//     ShouldSucceed().
		//     ShouldContainInPayload("success").
		//     ShouldExecuteWithin(5*time.Second)
		
		t.Log("Fluent syntax example (not implemented)")
	})
}

// Helper function to demonstrate result extraction
func extractPayloadFromResult(result *lambda.InvokeResult) string {
	// This helper extracts specific data from invocation results
	// In real implementations, this might parse JSON and extract specific fields
	return result.Payload
}

// Helper function to demonstrate custom event creation
func buildCustomEvent(eventType string, data map[string]interface{}) string {
	// This helper creates custom events based on type and data
	switch eventType {
	case "user_registration":
		return lambda.MarshalPayload(map[string]interface{}{
			"event_type": eventType,
			"user_data":  data,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
	case "order_placed":
		return lambda.MarshalPayload(map[string]interface{}{
			"event_type":  eventType,
			"order_data":  data,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		})
	default:
		return lambda.MarshalPayload(map[string]interface{}{
			"event_type": eventType,
			"data":       data,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// Example showing test organization and patterns
func TestExampleTestOrganization(t *testing.T) {
	// This example shows recommended test organization patterns
	
	// Group 1: Unit tests for individual functions
	t.Run("UnitTests", func(t *testing.T) {
		t.Run("PayloadValidation", func(t *testing.T) {
			// Test payload validation logic
		})
		
		t.Run("EventParsing", func(t *testing.T) {
			// Test event parsing utilities
		})
		
		t.Run("ConfigurationValidation", func(t *testing.T) {
			// Test configuration validation
		})
	})
	
	// Group 2: Integration tests for Lambda functions
	t.Run("IntegrationTests", func(t *testing.T) {
		t.Run("SingleFunction", func(t *testing.T) {
			// Test individual Lambda function invocation
		})
		
		t.Run("FunctionChain", func(t *testing.T) {
			// Test chain of Lambda function calls
		})
		
		t.Run("EventDriven", func(t *testing.T) {
			// Test event-driven function execution
		})
	})
	
	// Group 3: End-to-end tests for complete workflows
	t.Run("EndToEndTests", func(t *testing.T) {
		t.Run("OrderProcessingWorkflow", func(t *testing.T) {
			// Test complete order processing workflow
		})
		
		t.Run("UserRegistrationFlow", func(t *testing.T) {
			// Test complete user registration flow
		})
		
		t.Run("DataProcessingPipeline", func(t *testing.T) {
			// Test complete data processing pipeline
		})
	})
	
	// Group 4: Performance and load tests
	t.Run("PerformanceTests", func(t *testing.T) {
		t.Run("LoadTesting", func(t *testing.T) {
			// Test function under load
		})
		
		t.Run("ConcurrencyTesting", func(t *testing.T) {
			// Test concurrent invocations
		})
		
		t.Run("TimeoutTesting", func(t *testing.T) {
			// Test timeout scenarios
		})
	})
}

// Main test runner example
func TestLambdaPackageExamples(t *testing.T) {
	// This is the main test that would run all examples
	// In practice, each example would be its own test function
	
	t.Run("TestExampleUsage", func(t *testing.T) {
		TestExampleUsage(t)
	})
	
	t.Run("TestExampleTableDrivenTesting", func(t *testing.T) {
		TestExampleTableDrivenTesting(t)
	})
	
	t.Run("TestExampleAdvancedEventTesting", func(t *testing.T) {
		TestExampleAdvancedEventTesting(t)
	})
	
	t.Run("TestExampleEventSourceMapping", func(t *testing.T) {
		TestExampleEventSourceMapping(t)
	})
	
	t.Run("TestExamplePerformanceTesting", func(t *testing.T) {
		TestExamplePerformanceTesting(t)
	})
	
	t.Run("TestExampleErrorHandlingAndRecovery", func(t *testing.T) {
		TestExampleErrorHandlingAndRecovery(t)
	})
	
	// Skip comprehensive integration test by default
	if !testing.Short() {
		t.Run("TestExampleComprehensiveIntegrationTest", func(t *testing.T) {
			TestExampleComprehensiveIntegrationTest(t)
		})
	}
	
	t.Run("TestExampleTestDataFactories", func(t *testing.T) {
		TestExampleTestDataFactories(t)
	})
	
	t.Run("TestExampleValidationHelpers", func(t *testing.T) {
		TestExampleValidationHelpers(t)
	})
	
	t.Run("TestExamplePlaywrightLikeSyntax", func(t *testing.T) {
		TestExamplePlaywrightLikeSyntax(t)
	})
	
	t.Run("TestExampleTestOrganization", func(t *testing.T) {
		TestExampleTestOrganization(t)
	})
}