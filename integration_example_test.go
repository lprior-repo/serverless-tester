package main

import (
	"context"
	"testing"
	"time"

	"vasdeference"
	"vasdeference/dynamodb"
	"vasdeference/eventbridge"
	"vasdeference/lambda"
	"vasdeference/parallel"
	"vasdeference/snapshot"
	"vasdeference/stepfunctions"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/stretchr/testify/assert"
)

// TestComprehensiveServerlessWorkflow demonstrates the full power of the vasdeference framework
func TestComprehensiveServerlessWorkflow(t *testing.T) {
	// Initialize vasdeference with automatic namespace detection
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	// Create snapshot tester for regression testing
	snap := snapshot.New(t, snapshot.Options{
		SnapshotDir: "testdata/integration_snapshots",
		Sanitizers: []snapshot.Sanitizer{
			snapshot.SanitizeTimestamps(),
			snapshot.SanitizeLambdaRequestID(),
			snapshot.SanitizeExecutionArn(),
		},
	})

	// Mock Terraform outputs (in real scenarios, these would come from actual Terraform)
	terraformOutputs := map[string]interface{}{
		"order_processor_function":    vdf.PrefixResourceName("order-processor"),
		"inventory_processor_function": vdf.PrefixResourceName("inventory-processor"),
		"orders_table":                vdf.PrefixResourceName("orders"),
		"inventory_table":             vdf.PrefixResourceName("inventory"),
		"order_events_bus":            vdf.PrefixResourceName("order-events"),
		"order_workflow_arn":          "arn:aws:states:us-east-1:123456789012:stateMachine:" + vdf.PrefixResourceName("order-workflow"),
	}

	// Get resource names from Terraform outputs
	orderProcessorFunction := vdf.GetTerraformOutput(terraformOutputs, "order_processor_function")
	inventoryProcessorFunction := vdf.GetTerraformOutput(terraformOutputs, "inventory_processor_function")
	ordersTable := vdf.GetTerraformOutput(terraformOutputs, "orders_table")
	inventoryTable := vdf.GetTerraformOutput(terraformOutputs, "inventory_table")
	orderEventsBus := vdf.GetTerraformOutput(terraformOutputs, "order_events_bus")
	orderWorkflowArn := vdf.GetTerraformOutput(terraformOutputs, "order_workflow_arn")

	t.Run("Complete Order Processing Workflow", func(t *testing.T) {
		// Step 1: Setup test data in DynamoDB
		t.Log("Step 1: Setting up test data...")

		// Create inventory items
		inventoryItems := []map[string]interface{}{
			{
				"PK":        "PRODUCT#WIDGET-001",
				"SK":        "INVENTORY",
				"ProductId": "WIDGET-001",
				"Quantity":  100,
				"Price":     29.99,
			},
			{
				"PK":        "PRODUCT#GADGET-001", 
				"SK":        "INVENTORY",
				"ProductId": "GADGET-001",
				"Quantity":  50,
				"Price":     49.99,
			},
		}

		for _, item := range inventoryItems {
			dynamodb.PutItem(t, vdf.GetTestContext(), inventoryTable, item)
		}

		// Verify inventory setup with assertions
		dynamodb.AssertItemExists(t, vdf.GetTestContext(), inventoryTable, 
			dynamodb.CreateKey("PRODUCT#WIDGET-001", "INVENTORY"))
		dynamodb.AssertItemCount(t, vdf.GetTestContext(), inventoryTable, 2)

		// Step 2: Test Lambda functions individually
		t.Log("Step 2: Testing Lambda functions...")

		// Test order processor function
		orderRequest := map[string]interface{}{
			"orderId": "ORDER-123",
			"customerId": "CUSTOMER-456",
			"items": []map[string]interface{}{
				{"productId": "WIDGET-001", "quantity": 2},
				{"productId": "GADGET-001", "quantity": 1},
			},
		}

		orderResult := lambda.Invoke(t, vdf.GetTestContext(), orderProcessorFunction, orderRequest)
		lambda.AssertInvokeSuccess(t, orderResult)
		
		// Parse and validate response
		var orderResponse map[string]interface{}
		lambda.ParseInvokeOutput(t, orderResult, &orderResponse)
		assert.Equal(t, "ORDER-123", orderResponse["orderId"])
		assert.Equal(t, "pending", orderResponse["status"])

		// Snapshot the order processing response
		snap.MatchLambdaResponse("order_processor_response", orderResult.Payload)

		// Step 3: Test EventBridge event publishing
		t.Log("Step 3: Testing EventBridge integration...")

		// Create order created event
		orderCreatedDetail := map[string]interface{}{
			"orderId":    "ORDER-123",
			"customerId": "CUSTOMER-456",
			"status":     "created",
			"total":      109.97,
		}

		// Publish event to custom bus
		eventbridge.PutEvent(t, vdf.GetTestContext(), orderEventsBus, 
			"order.service", "Order Created", orderCreatedDetail)

		// Test event pattern matching
		eventPattern := eventbridge.EventPattern{
			Source:     []string{"order.service"},
			DetailType: []string{"Order Created"},
			Detail: map[string]interface{}{
				"status": []string{"created", "pending"},
			},
		}
		patternJSON := eventbridge.BuildEventPattern(t, eventPattern)
		
		testEvent := map[string]interface{}{
			"source":      "order.service",
			"detail-type": "Order Created",
			"detail":      orderCreatedDetail,
		}
		
		eventbridge.AssertEventPatternMatches(t, vdf.GetTestContext(), patternJSON, testEvent)

		// Step 4: Test Step Functions workflow
		t.Log("Step 4: Testing Step Functions workflow...")

		// Create workflow input using builder
		workflowInput := stepfunctions.CreateOrderInput("ORDER-123", []map[string]interface{}{
			{"sku": "WIDGET-001", "quantity": 2, "price": 29.99},
			{"sku": "GADGET-001", "quantity": 1, "price": 49.99},
		}).
		Set("customerId", "CUSTOMER-456").
		Set("priority", "standard").
		Build()

		// Execute workflow and wait for completion
		execution := stepfunctions.ExecuteStateMachineAndWait(t, vdf.GetTestContext(),
			orderWorkflowArn, "test-execution-123", workflowInput, 3*time.Minute)

		// Assert workflow succeeded
		stepfunctions.AssertExecutionSucceeded(t, execution)
		
		// Validate workflow output
		expectedOutput := map[string]interface{}{
			"orderId":  "ORDER-123",
			"status":   "completed",
			"total":    109.97,
		}
		stepfunctions.AssertExecutionOutput(t, execution, expectedOutput)

		// Snapshot workflow execution
		snap.MatchStepFunctionExecution("order_workflow_execution", execution)

		// Step 5: Verify final state in DynamoDB
		t.Log("Step 5: Verifying final state...")

		// Check order was created
		orderItem := dynamodb.GetItem(t, vdf.GetTestContext(), ordersTable,
			dynamodb.CreateKey("ORDER-123", "ORDER"))
		dynamodb.AssertAttributeEquals(t, orderItem, "Status", "completed")
		dynamodb.AssertAttributeEquals(t, orderItem, "CustomerId", "CUSTOMER-456")

		// Check inventory was decremented
		widgetInventory := dynamodb.GetItem(t, vdf.GetTestContext(), inventoryTable,
			dynamodb.CreateKey("PRODUCT#WIDGET-001", "INVENTORY"))
		dynamodb.AssertAttributeEquals(t, widgetInventory, "Quantity", 98) // 100 - 2

		gadgetInventory := dynamodb.GetItem(t, vdf.GetTestContext(), inventoryTable,
			dynamodb.CreateKey("PRODUCT#GADGET-001", "INVENTORY"))
		dynamodb.AssertAttributeEquals(t, gadgetInventory, "Quantity", 49) // 50 - 1

		// Snapshot final database state
		finalState := []map[string]interface{}{
			orderItem,
			widgetInventory,
			gadgetInventory,
		}
		snap.MatchDynamoDBItems("final_database_state", finalState)

		t.Log("✅ Complete order processing workflow test passed!")
	})
}

// TestParallelServerlessOperations demonstrates parallel test execution with resource isolation
func TestParallelServerlessOperations(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	// Create parallel runner with resource isolation
	runner, err := parallel.NewRunner(t, parallel.Options{
		PoolSize:          4, // 4 concurrent workers
		ResourceIsolation: true,
		TestTimeout:       2 * time.Minute,
	})
	assert.NoError(t, err)

	// Define test cases that run in parallel
	testCases := []parallel.TestCase{
		{
			Name: "ParallelLambdaInvocation",
			Func: func(ctx context.Context, tctx *parallel.TestContext) error {
				functionName := tctx.GetFunctionName("data-processor")
				
				input := map[string]interface{}{
					"testId": vasdeference.UniqueId(),
					"data":   []int{1, 2, 3, 4, 5},
				}

				result, err := lambda.InvokeE(tctx, vdf.GetTestContext(), functionName, input)
				if err != nil {
					return err
				}

				lambda.AssertInvokeSuccess(tctx, result)
				return nil
			},
		},
		{
			Name: "ParallelDynamoDBOperations", 
			Func: func(ctx context.Context, tctx *parallel.TestContext) error {
				tableName := tctx.GetTableName("test-data")
				
				testItem := map[string]interface{}{
					"PK": vasdeference.UniqueId(),
					"SK": "DATA",
					"TestValue": "parallel-test",
				}

				err := dynamodb.PutItemE(tctx, vdf.GetTestContext(), tableName, testItem)
				if err != nil {
					return err
				}

				// Verify item exists
				item, err := dynamodb.GetItemE(tctx, vdf.GetTestContext(), tableName,
					dynamodb.CreateKey(testItem["PK"].(string), "DATA"))
				if err != nil {
					return err
				}

				dynamodb.AssertItemExists(tctx, item)
				return nil
			},
		},
		{
			Name: "ParallelEventBridgePublishing",
			Func: func(ctx context.Context, tctx *parallel.TestContext) error {
				eventBusName := tctx.GetTableName("test-events") // Using table name pattern for event bus

				eventDetail := map[string]interface{}{
					"testId":    vasdeference.UniqueId(),
					"timestamp": time.Now().Unix(),
					"message":   "parallel test event",
				}

				return eventbridge.PutEventE(tctx, vdf.GetTestContext(), eventBusName,
					"test.service", "Parallel Test Event", eventDetail)
			},
		},
		{
			Name: "ParallelStepFunctionExecution",
			Func: func(ctx context.Context, tctx *parallel.TestContext) error {
				// Mock state machine ARN for parallel test
				stateMachineArn := "arn:aws:states:us-east-1:123456789012:stateMachine:" + 
					tctx.GetFunctionName("test-workflow")

				input := stepfunctions.NewInput().
					Set("testId", vasdeference.UniqueId()).
					Set("operation", "parallel-test").
					Build()

				executionName := "parallel-exec-" + vasdeference.RandomString(8)
				
				execution, err := stepfunctions.ExecuteStateMachineAndWaitE(tctx, vdf.GetTestContext(),
					stateMachineArn, executionName, input, 1*time.Minute)
				if err != nil {
					return err
				}

				stepfunctions.AssertExecutionSucceeded(tctx, execution)
				return nil
			},
		},
	}

	// Run all test cases in parallel
	runner.Run(testCases)
}

// TestTableDrivenServerlessScenarios demonstrates table-driven testing for various AWS services
func TestTableDrivenServerlessScenarios(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	t.Run("LambdaTableDriven", func(t *testing.T) {
		functionName := vdf.PrefixResourceName("validator")

		testCases := []struct {
			name           string
			input          map[string]interface{}
			expectSuccess  bool
			expectedOutput map[string]interface{}
		}{
			{
				name: "ValidInput",
				input: map[string]interface{}{
					"email": "user@example.com",
					"age":   25,
				},
				expectSuccess: true,
				expectedOutput: map[string]interface{}{
					"valid": true,
					"message": "validation_passed",
				},
			},
			{
				name: "InvalidEmail",
				input: map[string]interface{}{
					"email": "invalid-email",
					"age":   25,
				},
				expectSuccess: true,
				expectedOutput: map[string]interface{}{
					"valid": false,
					"errors": []string{"invalid_email_format"},
				},
			},
			{
				name: "MissingFields",
				input: map[string]interface{}{
					"email": "user@example.com",
				},
				expectSuccess: true,
				expectedOutput: map[string]interface{}{
					"valid": false,
					"errors": []string{"missing_age"},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := lambda.Invoke(t, vdf.GetTestContext(), functionName, tc.input)
				
				if tc.expectSuccess {
					lambda.AssertInvokeSuccess(t, result)
					
					var actualOutput map[string]interface{}
					lambda.ParseInvokeOutput(t, result, &actualOutput)
					assert.Equal(t, tc.expectedOutput, actualOutput)
				} else {
					lambda.AssertInvokeError(t, result)
				}
			})
		}
	})

	t.Run("StepFunctionsTableDriven", func(t *testing.T) {
		stateMachineArn := vdf.GetTerraformOutput(map[string]interface{}{
			"validation_workflow": "arn:aws:states:us-east-1:123456789012:stateMachine:" + 
				vdf.PrefixResourceName("validation-workflow"),
		}, "validation_workflow")

		testCases := []stepfunctions.TestCase{
			{
				Name: "SuccessfulValidation",
				Input: stepfunctions.NewInput().
					Set("requestId", vasdeference.UniqueId()).
					Set("data", map[string]interface{}{
						"type":  "order",
						"value": 100.50,
					}).
					Build(),
				ExpectedStatus: types.ExecutionStatusSucceeded,
				ExpectedOutput: map[string]interface{}{
					"status":   "validated",
					"approved": true,
				},
			},
			{
				Name: "ValidationFailure",
				Input: stepfunctions.NewInput().
					Set("requestId", vasdeference.UniqueId()).
					Set("data", map[string]interface{}{
						"type":  "invalid",
						"value": -50,
					}).
					Build(),
				ExpectedStatus: types.ExecutionStatusFailed,
				ExpectedError:  "ValidationError",
			},
			{
				Name: "RetryScenario",
				Input: stepfunctions.CreateRetryableInput(true, 3).
					Set("requestId", vasdeference.UniqueId()).
					Build(),
				ExpectedStatus: types.ExecutionStatusSucceeded,
				PostTest: func(t vasdeference.TestingT, ctx *vasdeference.TestContext, execution *stepfunctions.DescribeExecutionOutput) {
					// Verify retries occurred by checking execution history
					history := stepfunctions.GetExecutionHistory(t, ctx, execution.ExecutionArn)
					
					retryEvents := 0
					for _, event := range history {
						if event.Type == types.HistoryEventTypeTaskFailed {
							retryEvents++
						}
					}
					assert.GreaterOrEqual(t, retryEvents, 1, "Should have at least one retry attempt")
				},
			},
		}

		// Use table-driven test runner for Step Functions
		stepfunctions.RunTableDrivenTests(t, vdf.GetTestContext(), stateMachineArn, testCases)
	})
}

// TestErrorScenarios demonstrates comprehensive error handling across all services
func TestErrorScenarios(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()

	t.Run("LambdaErrorHandling", func(t *testing.T) {
		// Test non-existent function
		nonExistentFunction := vdf.PrefixResourceName("non-existent")
		
		_, err := lambda.InvokeE(t, vdf.GetTestContext(), nonExistentFunction, map[string]interface{}{})
		assert.Error(t, err, "Should fail when invoking non-existent function")
		
		// Assert function doesn't exist
		lambda.AssertFunctionDoesNotExist(t, vdf.GetTestContext(), nonExistentFunction)
	})

	t.Run("DynamoDBErrorHandling", func(t *testing.T) {
		// Test non-existent table
		nonExistentTable := vdf.PrefixResourceName("non-existent")
		
		_, err := dynamodb.GetItemE(t, vdf.GetTestContext(), nonExistentTable,
			dynamodb.CreateKey("TEST", "ITEM"))
		assert.Error(t, err, "Should fail when accessing non-existent table")
		
		// Assert table doesn't exist
		dynamodb.AssertTableDoesNotExist(t, vdf.GetTestContext(), nonExistentTable)
	})

	t.Run("EventBridgeErrorHandling", func(t *testing.T) {
		// Test invalid event pattern
		invalidPattern := `{"invalid": json}`
		testEvent := `{"source": "test"}`
		
		_, err := eventbridge.TestEventPatternE(t, vdf.GetTestContext(), invalidPattern, testEvent)
		assert.Error(t, err, "Should fail with invalid event pattern")
	})

	t.Run("StepFunctionsErrorHandling", func(t *testing.T) {
		// Test non-existent state machine
		nonExistentArn := "arn:aws:states:us-east-1:123456789012:stateMachine:non-existent"
		
		_, err := stepfunctions.StartExecutionE(t, vdf.GetTestContext(), nonExistentArn, "test", map[string]interface{}{})
		assert.Error(t, err, "Should fail when starting execution on non-existent state machine")
		
		// Assert state machine doesn't exist
		stepfunctions.AssertStateMachineDoesNotExist(t, vdf.GetTestContext(), nonExistentArn)
	})
}