package main

import (
	"fmt"
	"time"

	"vasdeference/lambda"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// ExampleBasicUsage demonstrates the basic usage patterns of the lambda package
// following functional programming principles
func ExampleBasicUsage() {
	// These examples show how to use the lambda testing package
	// Note: These are example patterns - actual implementations would need real AWS resources
	
	// Create test context - this would typically use real AWS configuration
	// ctx := &lambda.TestContext{T: t}
	fmt.Println("Basic Lambda usage patterns:")
	
	// Example 1: Basic Lambda function invocation
	_ = /* functionName */ "my-test-function"
	_ = /* payload */ `{"message": "Hello, Lambda!"}`
	
	// This would invoke an actual Lambda function in real implementations
	// result := lambda.Invoke(ctx, functionName, payload)
	// Expected: result.StatusCode == 200 and result.Payload contains "success"
	
	fmt.Printf("Basic Lambda invocation pattern for %s with payload %s\n", functionName, payload)
	
	// Example 2: Function configuration validation
	functionName = "my-test-function"
	
	// Get and validate function configuration
	// config := lambda.GetFunction(ctx, functionName)
	// lambda.AssertFunctionRuntime(ctx, functionName, types.RuntimeNodejs18x)
	// lambda.AssertFunctionTimeout(ctx, functionName, 30)
	// lambda.AssertEnvVarEquals(ctx, functionName, "NODE_ENV", "production")
	
	fmt.Printf("Function configuration validation pattern for %s\n", functionName)
	
	// Example 3: Event-driven processing with S3 events
	functionName = "s3-processor-function"
	_ = /* bucketName */ "test-bucket"
	_ = /* objectKey */ "test-data/file.json"
	
	// Build S3 event
	_ = /* s3Event */ lambda.BuildS3Event(bucketName, objectKey, "s3:ObjectCreated:Put")
	
	// Invoke with S3 event
	// result := lambda.Invoke(ctx, functionName, s3Event)
	// lambda.AssertInvokeSuccess(ctx, result)
	// lambda.AssertPayloadContains(ctx, result, "processed")
	
	// Verify logs contain expected processing messages
	// lambda.AssertLogsContain(ctx, functionName, "Processing S3 object")
	
	fmt.Printf("S3 event-driven pattern for %s using %s/%s\n", functionName, bucketName, objectKey)
	if len(s3Event) > 0 && bucketName != "" {
		fmt.Printf("S3 event generated successfully\n")
	}
	
	// Example 4: Retry and resilience patterns
	functionName = "unreliable-function"
	payload = `{"test": "retry"}`
	
	// Implement retry logic for eventual consistency
	// result := lambda.InvokeWithRetry(ctx, functionName, payload, 3)
	// lambda.AssertInvokeSuccess(ctx, result)
	
	fmt.Printf("Retry and resilience pattern for %s with payload %s\n", functionName, payload)
	
	// Example 5: Async invocation and log monitoring
	functionName = "async-processor"
	payload = `{"data": "async-test"}`
	
	// Async invocation
	// result := lambda.InvokeAsync(ctx, functionName, payload)
	// Expected: result.StatusCode == 202 (Async invocation returns 202)
	
	// Wait for logs to appear with expected content
	// logs := lambda.WaitForLogs(ctx, functionName, "Processing completed", 30*time.Second)
	// Expected: logs should not be empty
	
	fmt.Printf("Async invocation and log monitoring pattern for %s with payload %s\n", functionName, payload)
}

// ExampleTableDrivenValidation demonstrates comprehensive validation patterns
func ExampleTableDrivenValidation() {
	fmt.Println("Table-driven validation patterns:")
	
	// Table-driven scenarios for multiple payload validation
	_ = /* scenarios */ []struct {
		name              string
		functionName      string
		payload           string
		expectSuccess     bool
		expectedInPayload string
		description       string
	}{
		{
			name:              "ValidPayload",
			functionName:      "validator-function",
			payload:           `{"name": "John", "age": 30}`,
			expectSuccess:     true,
			expectedInPayload: "valid",
			description:       "Valid payload should be processed successfully",
		},
		{
			name:              "InvalidPayload",
			functionName:      "validator-function",
			payload:           `{"name": ""}`,
			expectSuccess:     false,
			expectedInPayload: "error",
			description:       "Invalid payload should return validation error",
		},
		{
			name:              "EmptyPayload",
			functionName:      "validator-function",
			payload:           `{}`,
			expectSuccess:     false,
			expectedInPayload: "missing_required_fields",
			description:       "Empty payload should return missing fields error",
		},
	}
	
	for _, scenario := range scenarios {
		fmt.Printf("Scenario %s: %s\n", scenario.name, scenario.description)
		
		// In real implementations, this would invoke actual Lambda functions
		// result := lambda.Invoke(ctx, scenario.functionName, scenario.payload)
		// 
		// if scenario.expectSuccess {
		//     lambda.AssertInvokeSuccess(ctx, result)
		// } else {
		//     lambda.AssertInvokeError(ctx, result)
		// }
		// 
		// lambda.AssertPayloadContains(ctx, result, scenario.expectedInPayload)
	}
}

// ExampleAdvancedEventProcessing shows advanced event processing patterns
func ExampleAdvancedEventProcessing() {
	fmt.Println("Advanced event processing patterns:")
	
	_ = /* functionName */ "multi-trigger-function"
		
	// S3 event processing
	_ = /* s3Event */ lambda.BuildS3Event("data-bucket", "input/data.json", "s3:ObjectCreated:Put")
	
	// result := lambda.Invoke(ctx, functionName, s3Event)
	// lambda.AssertInvokeSuccess(ctx, result)
	// lambda.AssertPayloadContains(ctx, result, "s3_processed")
	
	fmt.Println("S3 event processing pattern completed")
		
	// DynamoDB event processing
	_ = /* keys */ map[string]interface{}{
		"id": map[string]interface{}{"S": "item-123"},
	}
	_ = /* dynamoEvent */ lambda.BuildDynamoDBEvent("users-table", "INSERT", keys)
	
	// result := lambda.Invoke(ctx, functionName, dynamoEvent)
	// lambda.AssertInvokeSuccess(ctx, result)
	// lambda.AssertPayloadContains(ctx, result, "dynamodb_processed")
	
	fmt.Println("DynamoDB event processing pattern completed")
		
	// SQS event processing
	_ = /* queueUrl */ "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
	_ = /* messageBody */ `{"order_id": "order-456", "status": "pending"}`
	_ = /* sqsEvent */ lambda.BuildSQSEvent(queueUrl, messageBody)
	
	// result := lambda.Invoke(ctx, functionName, sqsEvent)
	// lambda.AssertInvokeSuccess(ctx, result)
	// lambda.AssertPayloadContains(ctx, result, "sqs_processed")
	
	fmt.Println("SQS event processing pattern completed")
		
	// API Gateway event processing
	_ = /* apiEvent */ lambda.BuildAPIGatewayEvent("POST", "/api/orders", `{"item": "widget", "quantity": 5}`)
	
	// result := lambda.Invoke(ctx, functionName, apiEvent)
	// lambda.AssertInvokeSuccess(ctx, result)
	// 
	// // Parse API Gateway response
	// var apiResponse map[string]interface{}
	// lambda.ParseInvokeOutput(result, &apiResponse)
	// Expected: apiResponse["statusCode"] == 200
	
	fmt.Println("API Gateway event processing pattern completed")
}

// ExampleEventSourceMapping demonstrates event source mapping patterns
func ExampleEventSourceMapping() {
	fmt.Println("SQS event source mapping patterns:")
	
	_ = /* functionName */ "sqs-consumer-function"
	_ = /* queueArn */ "arn:aws:sqs:us-east-1:123456789012:test-queue"
	
	// Create event source mapping configuration
	_ = /* mappingConfig */ lambda.EventSourceMappingConfig{
		EventSourceArn:   queueArn,
		FunctionName:     functionName,
		BatchSize:        10,
		Enabled:          true,
		StartingPosition: types.EventSourcePositionLatest,
	}
	
	// In real implementations:
	// mapping := lambda.CreateEventSourceMapping(ctx, mappingConfig)
	// arrange.RegisterCleanup(func() error {
	//     return lambda.DeleteEventSourceMappingE(ctx, mapping.UUID)
	// })
	// 
	// // Wait for mapping to be active
	// lambda.WaitForEventSourceMappingState(ctx, mapping.UUID, "Enabled", 2*time.Minute)
	// 
	// // Verify mapping configuration
	// Expected: mapping.EventSourceArn == queueArn
	// Expected: mapping.BatchSize == 10
	
	fmt.Printf("SQS event source mapping configured for function %s with batch size %d\n", 
		mappingConfig.FunctionName, mappingConfig.BatchSize)
}

// ExamplePerformancePatterns shows performance and load patterns
func ExamplePerformancePatterns() {
	fmt.Println("Performance and load patterns:")
	
	_ = /* functionName */ "high-load-function"
	_ = /* payload */ `{"load_test": true}`
	
	// Concurrent invocations pattern
	_ = /* concurrency */ 10
	_ = /* results */ make(chan *lambda.InvokeResult, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func() {
			// result := lambda.Invoke(ctx, functionName, payload)
			// results <- result
			results <- &lambda.InvokeResult{StatusCode: 200, Payload: `{"success": true}`}
		}()
	}
	
	// Collect results
	_ = /* successCount */ 0
	for i := 0; i < concurrency; i++ {
	_ = /* result */ <-results
		if result.StatusCode == 200 {
			successCount++
		}
	}
	
	// Expected: all invocations should succeed
	fmt.Printf("Concurrent invocations completed: %d/%d successful\n", successCount, concurrency)
	
	// Timeout configuration patterns
	functionName = "timeout-test-function"
	
	// Function timeout configuration
	// config := lambda.GetFunction(ctx, functionName)
	// lambda.AssertFunctionTimeout(ctx, functionName, 30)
	
	// Invocation with custom timeout
	_ = /* opts */ &lambda.InvokeOptions{
		Timeout: 5 * time.Second,
	}
	
	// result := lambda.InvokeWithOptions(ctx, functionName, `{"delay": 10}`, opts)
	// lambda.AssertExecutionTimeLessThan(ctx, result, 5*time.Second)
	
	fmt.Printf("Timeout pattern configured for %s with %v timeout\n", functionName, opts.Timeout)
}

// ExampleErrorHandlingAndRecovery demonstrates error scenarios and recovery patterns
func ExampleErrorHandlingAndRecovery() {
	fmt.Println("Error handling and recovery patterns:")
	
	_ = /* functionName */ "error-test-function"
	
	// Function error handling
	_ = /* errorPayload */ `{"trigger_error": "division_by_zero"}`
	
	// In real implementations:
	// result := lambda.Invoke(ctx, functionName, errorPayload)
	// lambda.AssertInvokeError(ctx, result)
	// Expected: result.FunctionError contains "Handled"
	
	// Verify error logs
	// lambda.AssertLogsContainLevel(ctx, functionName, "ERROR")
	// lambda.AssertLogsContain(ctx, functionName, "division_by_zero")
	
	fmt.Printf("Error scenario pattern for %s with payload %s\n", functionName, errorPayload)
	
	// Recovery after transient errors
	functionName = "recovery-test-function"
	
	// Recovery from transient errors
	_ = /* recoverablePayload */ `{"simulate_transient_error": true}`
	
	// Use retry logic to handle transient errors
	// result := lambda.InvokeWithRetry(ctx, functionName, recoverablePayload, 3)
	// lambda.AssertInvokeSuccess(ctx, result)
	
	// Verify recovery logs
	// lambda.AssertLogsContain(ctx, functionName, "recovered_after_retry")
	
	fmt.Printf("Recovery pattern for %s with payload %s\n", functionName, recoverablePayload)
}

// ExampleComprehensiveIntegration shows a full end-to-end workflow
func ExampleComprehensiveIntegration() {
	fmt.Println("Comprehensive serverless workflow patterns:")
	
	// This example shows a complete serverless workflow
	// Note: Would require actual AWS Lambda functions and resources in practice
		
	// Step 1: API Gateway receives order
	_ = /* orderData */ `{
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
	
	_ = /* apiEvent */ lambda.BuildAPIGatewayEvent("POST", "/api/orders", orderData)
		
	// Order validation function
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
	
	fmt.Println("Comprehensive order processing workflow pattern completed")
}

// ExampleDataFactories shows how to create reusable data factories
func ExampleDataFactories() {
	fmt.Println("Data factory patterns:")
	
	// Data factory functions following functional programming principles
	_ = /* createUserPayload */ func(name string, age int) string {
		return lambda.MarshalPayload(map[string]interface{}{
			"name":       name,
			"age":        age,
			"created_at": time.Now().UTC().Format(time.RFC3339),
		})
	}
	
	_ = /* createOrderPayload */ func(customerID string, amount float64) string {
		return lambda.MarshalPayload(map[string]interface{}{
			"customer_id": customerID,
			"amount":      amount,
			"currency":    "USD",
			"status":      "pending",
			"created_at":  time.Now().UTC().Format(time.RFC3339),
		})
	}
	
	// Example usage of data factories
	_ = /* userPayload */ createUserPayload("John Doe", 30)
	
	// result := lambda.Invoke(ctx, "user-processor", userPayload)
	// lambda.AssertInvokeSuccess(ctx, result)
	// lambda.AssertPayloadContains(ctx, result, "user_created")
	
	fmt.Printf("User payload created: contains John Doe and age 30\n")
	
	_ = /* orderPayload */ createOrderPayload("cust-123", 99.99)
	
	// result := lambda.Invoke(ctx, "order-processor", orderPayload)
	// lambda.AssertInvokeSuccess(ctx, result)
	// lambda.AssertPayloadContains(ctx, result, "order_processed")
	
	fmt.Printf("Order payload created: contains cust-123 and amount 99.99\n")
}

// ExampleValidationHelpers shows comprehensive validation patterns
func ExampleValidationHelpers() {
	fmt.Println("Configuration validation patterns:")
	
	// Create expected configuration for validation
	_ = /* expectedConfig */ &lambda.FunctionConfiguration{
		FunctionName: "production-function",
		Runtime:      types.RuntimeNodejs18x,
		Handler:      "index.handler",
		Timeout:      30,
		MemorySize:   256,
		State:        types.StateActive,
		Environment: map[string]string{
			"NODE_ENV":  "production",
			"LOG_LEVEL": "info",
		},
	}
	
	// Example function configuration (in real implementations, this would come from GetFunction)
	_ = /* actualConfig */ &lambda.FunctionConfiguration{
		FunctionName: "production-function",
		Runtime:      types.RuntimeNodejs18x,
		Handler:      "index.handler",
		Timeout:      30,
		MemorySize:   256,
		State:        types.StateActive,
		Environment: map[string]string{
			"NODE_ENV":  "production",
			"LOG_LEVEL": "info",
			"DEBUG":     "false", // Extra environment variable
		},
	}
	
	// Validate configuration
	_ = /* errors */ lambda.ValidateFunctionConfiguration(actualConfig, expectedConfig)
	
	// In this case, no errors should be found since actual matches or exceeds expected
	if len(errors) == 0 {
		fmt.Println("Configuration validation passed")
	}
	
	// Example with validation errors
	_ = /* invalidConfig */ &lambda.FunctionConfiguration{
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
	if len(errors) > 0 {
		fmt.Printf("Invalid configuration has %d validation errors\n", len(errors))
		// Print validation errors
		for _, err := range errors {
			fmt.Printf("Validation error: %s\n", err)
		}
	}
}

// ExampleFluentAPIPatterns demonstrates fluent API patterns
func ExampleFluentAPIPatterns() {
	fmt.Println("Fluent API patterns:")
	
	// Current functional syntax
	_ = /* functionName */ "test-function"
	_ = /* payload */ `{"test": true}`
	
	// Standard function calls
	// result := lambda.Invoke(ctx, functionName, payload)
	// lambda.AssertInvokeSuccess(ctx, result)
	// lambda.AssertPayloadContains(ctx, result, "success")
	// lambda.AssertExecutionTimeLessThan(ctx, result, 5*time.Second)
	
	fmt.Printf("Current functional syntax for %s with %s\n", functionName, payload)
	
	// Potential fluent syntax (would require additional wrapper implementation)
	// Example of what fluent syntax might look like:
	// lambda.Function(ctx, "test-function").
	//     WithPayload(`{"test": true}`).
	//     Invoke().
	//     ShouldSucceed().
	//     ShouldContainInPayload("success").
	//     ShouldExecuteWithin(5*time.Second)
	
	fmt.Println("Fluent syntax pattern (conceptual)")
}

// Helper functions for data processing and event creation

// extractPayloadFromResult demonstrates result extraction
func extractPayloadFromResult(result *lambda.InvokeResult) string {
	// This helper extracts specific data from invocation results
	// In real implementations, this might parse JSON and extract specific fields
	return result.Payload
}

// buildCustomEvent demonstrates custom event creation
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
			"event_type": eventType,
			"order_data": data,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
	default:
		return lambda.MarshalPayload(map[string]interface{}{
			"event_type": eventType,
			"data":       data,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// ExampleOrganizationPatterns shows recommended organization patterns
func ExampleOrganizationPatterns() {
	fmt.Println("Lambda function organization patterns:")
	
	// Group 1: Unit-level patterns for individual functions
	fmt.Println("Unit patterns:")
	fmt.Println("  - Payload validation logic")
	fmt.Println("  - Event parsing utilities")
	fmt.Println("  - Configuration validation")
	
	// Group 2: Integration patterns for Lambda functions
	fmt.Println("Integration patterns:")
	fmt.Println("  - Individual Lambda function invocation")
	fmt.Println("  - Chain of Lambda function calls")
	fmt.Println("  - Event-driven function execution")
	
	// Group 3: End-to-end patterns for complete workflows
	fmt.Println("End-to-end patterns:")
	fmt.Println("  - Complete order processing workflow")
	fmt.Println("  - Complete user registration flow")
	fmt.Println("  - Complete data processing pipeline")
	
	// Group 4: Performance and load patterns
	fmt.Println("Performance patterns:")
	fmt.Println("  - Function under load scenarios")
	fmt.Println("  - Concurrent invocations")
	fmt.Println("  - Timeout scenarios")
}

// runAllExamples demonstrates running all Lambda examples
func runAllExamples() {
	fmt.Println("Running all Lambda package examples:")
	
	ExampleBasicUsage()
	fmt.Println("")
	
	ExampleTableDrivenValidation()
	fmt.Println("")
	
	ExampleAdvancedEventProcessing()
	fmt.Println("")
	
	ExampleEventSourceMapping()
	fmt.Println("")
	
	ExamplePerformancePatterns()
	fmt.Println("")
	
	ExampleErrorHandlingAndRecovery()
	fmt.Println("")
	
	ExampleComprehensiveIntegration()
	fmt.Println("")
	
	ExampleDataFactories()
	fmt.Println("")
	
	ExampleValidationHelpers()
	fmt.Println("")
	
	ExampleFluentAPIPatterns()
	fmt.Println("")
	
	ExampleOrganizationPatterns()
	fmt.Println("All Lambda examples completed")
}

func main() {
	// This file demonstrates Lambda usage patterns with the vasdeference framework.
	// Run examples with: go run ./examples/lambda/examples.go
	
	runAllExamples()
}