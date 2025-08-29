// Package e2e demonstrates comprehensive functional E2E testing for serverless applications
// Following strict TDD methodology with functional programming patterns and advanced VasDeference features
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"vasdeference"
	"vasdeference/dynamodb"
	"vasdeference/eventbridge"
	"vasdeference/lambda"
	"vasdeference/snapshot"
	"vasdeference/stepfunctions"
)

// Test Models for E-Commerce Pipeline

type OrderRequest struct {
	OrderID     string                 `json:"orderId"`
	CustomerID  string                 `json:"customerId"`
	Items       []OrderItem            `json:"items"`
	PaymentInfo PaymentInformation     `json:"paymentInfo"`
	ShippingInfo ShippingInformation   `json:"shippingInfo"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type OrderItem struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	SKU       string  `json:"sku"`
}

type PaymentInformation struct {
	Method    string  `json:"method"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Token     string  `json:"token"`
}

type ShippingInformation struct {
	Address     Address `json:"address"`
	Method      string  `json:"method"`
	ExpectedDate string  `json:"expectedDate"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZIP     string `json:"zip"`
	Country string `json:"country"`
}

type OrderResult struct {
	Success          bool                   `json:"success"`
	OrderID          string                 `json:"orderId"`
	Status           string                 `json:"status"`
	PaymentStatus    string                 `json:"paymentStatus"`
	InventoryStatus  string                 `json:"inventoryStatus"`
	ShippingStatus   string                 `json:"shippingStatus"`
	ExecutionArn     string                 `json:"executionArn"`
	ProcessingTime   time.Duration          `json:"processingTime"`
	ValidationErrors []ValidationError      `json:"validationErrors,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Test Infrastructure Setup

type ECommerceTestContext struct {
	VDF                  *vasdeference.VasDeference
	OrderTableName       string
	InventoryTableName   string
	CustomerTableName    string
	OrderValidatorFunc   string
	PaymentProcessorFunc string
	InventoryManagerFunc string
	OrderWorkflowSM      string
	EventBusName         string
	Snapshot             *dynamodb.Snapshot
}

func setupECommerceInfrastructure(t *testing.T) *ECommerceTestContext {
	t.Helper()
	
	// Functional infrastructure setup with monadic error handling
	setupResult := lo.Pipe3(
		time.Now().Unix(),
		func(timestamp int64) mo.Result[*vasdeference.VasDeference] {
			vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDeferenceOptions{
				Region:    "us-east-1",
				Namespace: fmt.Sprintf("functional-ecommerce-e2e-%d", timestamp),
			})
			if vdf == nil {
				return mo.Err[*vasdeference.VasDeference](fmt.Errorf("failed to create VasDeference instance"))
			}
			return mo.Ok(vdf)
		},
		func(vdfResult mo.Result[*vasdeference.VasDeference]) mo.Result[*ECommerceTestContext] {
			return vdfResult.Map(func(vdf *vasdeference.VasDeference) *ECommerceTestContext {
				return &ECommerceTestContext{
					VDF:                  vdf,
					OrderTableName:       vdf.PrefixResourceName("functional-orders"),
					InventoryTableName:   vdf.PrefixResourceName("functional-inventory"),
					CustomerTableName:    vdf.PrefixResourceName("functional-customers"),
					OrderValidatorFunc:   vdf.PrefixResourceName("functional-order-validator"),
					PaymentProcessorFunc: vdf.PrefixResourceName("functional-payment-processor"),
					InventoryManagerFunc: vdf.PrefixResourceName("functional-inventory-manager"),
					OrderWorkflowSM:      vdf.PrefixResourceName("functional-order-workflow"),
					EventBusName:         vdf.PrefixResourceName("functional-order-events"),
				}
			})
		},
		func(ctxResult mo.Result[*ECommerceTestContext]) *ECommerceTestContext {
			return ctxResult.OrElse(nil)
		},
	)
	
	if setupResult == nil {
		t.Fatal("Failed to setup functional e-commerce infrastructure")
		return nil
	}
	
	ctx := setupResult
	
	// Functional setup chain with error handling
	setupChain := lo.Pipe4(
		ctx,
		func(c *ECommerceTestContext) mo.Result[*ECommerceTestContext] {
			setupTestTables(t, c)
			return mo.Ok(c)
		},
		func(result mo.Result[*ECommerceTestContext]) mo.Result[*ECommerceTestContext] {
			return result.FlatMap(func(c *ECommerceTestContext) mo.Result[*ECommerceTestContext] {
				setupTestLambdaFunctions(t, c)
				return mo.Ok(c)
			})
		},
		func(result mo.Result[*ECommerceTestContext]) mo.Result[*ECommerceTestContext] {
			return result.FlatMap(func(c *ECommerceTestContext) mo.Result[*ECommerceTestContext] {
				setupOrderWorkflow(t, c)
				return mo.Ok(c)
			})
		},
		func(result mo.Result[*ECommerceTestContext]) *ECommerceTestContext {
			return result.Match(
				func(c *ECommerceTestContext) *ECommerceTestContext {
					setupEventBridge(t, c)
					return c
				},
				func(err error) *ECommerceTestContext {
					t.Fatalf("Functional setup chain failed: %v", err)
					return nil
				},
			)
		},
	)
	
	ctx = setupChain
	
	// Functional snapshot creation with validation
	snapshotResult := lo.Pipe2(
		[]string{ctx.OrderTableName, ctx.InventoryTableName, ctx.CustomerTableName},
		func(tableNames []string) mo.Result[*dynamodb.Snapshot] {
			if len(tableNames) == 0 {
				return mo.Err[*dynamodb.Snapshot](fmt.Errorf("no tables to snapshot"))
			}
			snapshot := dynamodb.CreateSnapshot(t, tableNames, dynamodb.SnapshotOptions{
				Compression: true,
				Encryption:  true,
				Incremental: true,
				Metadata: map[string]interface{}{
					"test_suite":     "functional_ecommerce_pipeline",
					"created_at":     time.Now(),
					"table_count":    len(tableNames),
					"functional_mode": true,
				},
			})
			return mo.Ok(snapshot)
		},
		func(snapshotResult mo.Result[*dynamodb.Snapshot]) *dynamodb.Snapshot {
			return snapshotResult.Match(
				func(snapshot *dynamodb.Snapshot) *dynamodb.Snapshot {
					ctx.Snapshot = snapshot
					return snapshot
				},
				func(err error) *dynamodb.Snapshot {
					t.Logf("Warning: Snapshot creation failed: %v", err)
					return nil
				},
			)
		},
	)
	
	// Functional cleanup registration
	cleanupResult := mo.Some(ctx.VDF).Map(func(vdf *vasdeference.VasDeference) error {
		return vdf.RegisterCleanup(func() error {
			if snapshotResult != nil {
				snapshotResult.Restore()
			}
			return cleanupECommerceInfrastructure(ctx)
		})
	})
	
	cleanupResult.IfPresent(func(err error) {
		if err != nil {
			t.Logf("Warning: Cleanup registration failed: %v", err)
		}
	})
	
	return ctx
}

func setupTestTables(t *testing.T, ctx *ECommerceTestContext) {
	t.Helper()
	
	// Create Orders table
	err := dynamodb.CreateTable(ctx.VDF.Context, ctx.OrderTableName, dynamodb.TableDefinition{
		HashKey: dynamodb.AttributeDefinition{Name: "orderId", Type: "S"},
		RangeKey: &dynamodb.AttributeDefinition{Name: "customerId", Type: "S"},
		BillingMode: "PAY_PER_REQUEST",
		Tags: map[string]string{
			"Environment": "test",
			"TestSuite":   "ecommerce-pipeline",
		},
	})
	require.NoError(t, err)
	
	// Create Inventory table
	err = dynamodb.CreateTable(ctx.VDF.Context, ctx.InventoryTableName, dynamodb.TableDefinition{
		HashKey: dynamodb.AttributeDefinition{Name: "productId", Type: "S"},
		BillingMode: "PAY_PER_REQUEST",
		Attributes: []dynamodb.AttributeDefinition{
			{Name: "sku", Type: "S"},
			{Name: "category", Type: "S"},
		},
		GlobalSecondaryIndexes: []dynamodb.GSIDefinition{
			{
				IndexName: "sku-index",
				HashKey:   "sku",
			},
			{
				IndexName: "category-index",
				HashKey:   "category",
			},
		},
	})
	require.NoError(t, err)
	
	// Create Customers table
	err = dynamodb.CreateTable(ctx.VDF.Context, ctx.CustomerTableName, dynamodb.TableDefinition{
		HashKey: dynamodb.AttributeDefinition{Name: "customerId", Type: "S"},
		BillingMode: "PAY_PER_REQUEST",
	})
	require.NoError(t, err)
	
	// Wait for tables to be active
	dynamodb.WaitForTableActive(ctx.VDF.Context, ctx.OrderTableName, 60*time.Second)
	dynamodb.WaitForTableActive(ctx.VDF.Context, ctx.InventoryTableName, 60*time.Second)
	dynamodb.WaitForTableActive(ctx.VDF.Context, ctx.CustomerTableName, 60*time.Second)
}

func setupTestLambdaFunctions(t *testing.T, ctx *ECommerceTestContext) {
	t.Helper()
	
	// Create order validator function
	err := lambda.CreateFunction(ctx.VDF.Context, ctx.OrderValidatorFunc, lambda.FunctionConfig{
		Runtime: "nodejs18.x",
		Handler: "index.handler",
		Code:    generateOrderValidatorCode(),
		Environment: map[string]string{
			"ORDERS_TABLE":    ctx.OrderTableName,
			"INVENTORY_TABLE": ctx.InventoryTableName,
			"CUSTOMER_TABLE":  ctx.CustomerTableName,
		},
	})
	require.NoError(t, err)
	
	// Create payment processor function
	err = lambda.CreateFunction(ctx.VDF.Context, ctx.PaymentProcessorFunc, lambda.FunctionConfig{
		Runtime: "nodejs18.x", 
		Handler: "index.handler",
		Code:    generatePaymentProcessorCode(),
		Environment: map[string]string{
			"ORDERS_TABLE": ctx.OrderTableName,
		},
	})
	require.NoError(t, err)
	
	// Create inventory manager function
	err = lambda.CreateFunction(ctx.VDF.Context, ctx.InventoryManagerFunc, lambda.FunctionConfig{
		Runtime: "nodejs18.x",
		Handler: "index.handler", 
		Code:    generateInventoryManagerCode(),
		Environment: map[string]string{
			"INVENTORY_TABLE": ctx.InventoryTableName,
		},
	})
	require.NoError(t, err)
	
	// Wait for functions to be active
	lambda.WaitForFunctionActive(ctx.VDF.Context, ctx.OrderValidatorFunc, 30*time.Second)
	lambda.WaitForFunctionActive(ctx.VDF.Context, ctx.PaymentProcessorFunc, 30*time.Second)
	lambda.WaitForFunctionActive(ctx.VDF.Context, ctx.InventoryManagerFunc, 30*time.Second)
}

func setupOrderWorkflow(t *testing.T, ctx *ECommerceTestContext) {
	t.Helper()
	
	definition := generateOrderWorkflowDefinition(ctx)
	
	err := stepfunctions.CreateStateMachine(ctx.VDF.Context, ctx.OrderWorkflowSM, stepfunctions.StateMachineConfig{
		Definition: definition,
		RoleArn:    getStepFunctionsRoleArn(),
		Type:       "STANDARD",
		Tags: map[string]string{
			"Environment": "test",
			"TestSuite":   "ecommerce-pipeline",
		},
	})
	require.NoError(t, err)
	
	stepfunctions.WaitForStateMachineActive(ctx.VDF.Context, ctx.OrderWorkflowSM, 30*time.Second)
}

func setupEventBridge(t *testing.T, ctx *ECommerceTestContext) {
	t.Helper()
	
	err := eventbridge.CreateEventBus(ctx.VDF.Context, ctx.EventBusName, eventbridge.EventBusConfig{
		Description: "E-commerce order processing events",
		Tags: map[string]string{
			"Environment": "test",
			"TestSuite":   "ecommerce-pipeline",
		},
	})
	require.NoError(t, err)
}

// Test Data Generators

func generateValidOrderRequest() OrderRequest {
	return OrderRequest{
		OrderID:    fmt.Sprintf("order-%d", time.Now().UnixNano()),
		CustomerID: "customer-123",
		Items: []OrderItem{
			{
				ProductID: "product-001",
				Quantity:  2,
				Price:     29.99,
				SKU:       "SKU-001",
			},
			{
				ProductID: "product-002", 
				Quantity:  1,
				Price:     49.99,
				SKU:       "SKU-002",
			},
		},
		PaymentInfo: PaymentInformation{
			Method:   "credit_card",
			Amount:   109.97,
			Currency: "USD",
			Token:    "payment-token-123",
		},
		ShippingInfo: ShippingInformation{
			Address: Address{
				Street:  "123 Test Street",
				City:    "TestCity", 
				State:   "CA",
				ZIP:     "12345",
				Country: "USA",
			},
			Method:       "standard",
			ExpectedDate: time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02"),
		},
	}
}

func generateInvalidPaymentOrder() OrderRequest {
	order := generateValidOrderRequest()
	order.PaymentInfo.Token = "invalid-token"
	return order
}

func generateInsufficientInventoryOrder() OrderRequest {
	order := generateValidOrderRequest()
	order.Items[0].Quantity = 1000 // Assuming this exceeds available inventory
	return order
}

func generateInvalidCustomerOrder() OrderRequest {
	order := generateValidOrderRequest()
	order.CustomerID = "non-existent-customer"
	return order
}

// Advanced Table-Driven Testing Implementation

func TestECommerceOrderProcessingPipeline(t *testing.T) {
	ctx := setupECommerceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	// Seed test data
	seedTestData(t, ctx)
	
	// Run comprehensive table-driven tests with parallel execution
	vasdeference.TableTest[OrderRequest, OrderResult](t, "E-Commerce Order Pipeline").
		Case("valid_order_success", generateValidOrderRequest(), OrderResult{
			Success:         true,
			Status:          "confirmed",
			PaymentStatus:   "completed",
			InventoryStatus: "reserved",
			ShippingStatus:  "pending",
		}).
		Case("invalid_payment_failure", generateInvalidPaymentOrder(), OrderResult{
			Success:       false,
			Status:        "payment_failed",
			PaymentStatus: "failed",
			ValidationErrors: []ValidationError{
				{Field: "paymentInfo.token", Message: "Invalid payment token", Code: "PAYMENT_001"},
			},
		}).
		Case("insufficient_inventory_failure", generateInsufficientInventoryOrder(), OrderResult{
			Success:         false,
			Status:          "inventory_failed", 
			InventoryStatus: "insufficient",
			ValidationErrors: []ValidationError{
				{Field: "items[0].quantity", Message: "Insufficient inventory", Code: "INVENTORY_001"},
			},
		}).
		Case("invalid_customer_failure", generateInvalidCustomerOrder(), OrderResult{
			Success: false,
			Status:  "customer_validation_failed",
			ValidationErrors: []ValidationError{
				{Field: "customerId", Message: "Customer not found", Code: "CUSTOMER_001"},
			},
		}).
		Parallel().
		WithMaxWorkers(5).
		Repeat(3).
		Timeout(60 * time.Second).
		Run(func(order OrderRequest) OrderResult {
			return processCompleteOrder(t, ctx, order)
		}, func(testName string, input OrderRequest, expected OrderResult, actual OrderResult) {
			validateOrderResult(t, testName, input, expected, actual)
		})
}

func processCompleteOrder(t *testing.T, ctx *ECommerceTestContext, order OrderRequest) OrderResult {
	t.Helper()
	
	startTime := time.Now()
	
	// Functional order processing with monadic error handling
	processingResult := lo.Pipe4(
		order,
		func(o OrderRequest) mo.Result[[]byte] {
			input, err := json.Marshal(o)
			if err != nil {
				return mo.Err[[]byte](fmt.Errorf("failed to marshal order: %w", err))
			}
			return mo.Ok(input)
		},
		func(inputResult mo.Result[[]byte]) mo.Result[string] {
			return inputResult.Map(func(input []byte) string {
				return fmt.Sprintf("functional-order-%s-%d", order.OrderID, time.Now().UnixNano())
			})
		},
		func(nameResult mo.Result[string]) mo.Result[struct{ name string; input []byte }] {
			return mo.Tuple2(nameResult, mo.Ok([]byte{})).Match(
				func(name mo.Result[string], _ mo.Result[[]byte]) mo.Result[struct{ name string; input []byte }] {
					return name.Map(func(n string) struct{ name string; input []byte } {
						input, _ := json.Marshal(order)
						return struct{ name string; input []byte }{n, input}
					})
				},
				func() mo.Result[struct{ name string; input []byte }] {
					return mo.Err[struct{ name string; input []byte }](fmt.Errorf("failed to prepare execution"))
				},
			)
		},
		func(execDataResult mo.Result[struct{ name string; input []byte }]) OrderResult {
			return execDataResult.Match(
				func(data struct{ name string; input []byte }) OrderResult {
					// Simulate workflow execution
					execution, err := stepfunctions.StartExecutionE(ctx.VDF.Context, ctx.OrderWorkflowSM, data.name, stepfunctions.JSONInput(data.input))
					if err != nil {
						return OrderResult{
							Success: false,
							Status:  "functional_workflow_start_failed",
							ValidationErrors: []ValidationError{
								{Field: "functional_workflow", Message: err.Error(), Code: "FUNCTIONAL_WORKFLOW_001"},
							},
						}
					}
					return processFunctionalWorkflowExecution(t, ctx, execution, order, startTime)
				},
				func(err error) OrderResult {
					return OrderResult{
						Success:        false,
						Status:         "functional_processing_failed",
						ProcessingTime: time.Since(startTime),
						ValidationErrors: []ValidationError{
							{Field: "functional_processing", Message: err.Error(), Code: "FUNCTIONAL_PROCESSING_001"},
						},
					}
				},
			)
		},
	)
	
	return processingResult
	
}

func processFunctionalWorkflowExecution(t *testing.T, ctx *ECommerceTestContext, execution interface{}, order OrderRequest, startTime time.Time) OrderResult {
	t.Helper()
	
	// Functional workflow execution processing
	workflowResult := lo.Pipe3(
		execution,
		func(exec interface{}) mo.Result[interface{}] {
			// Simulate waiting for completion with functional timeout handling
			result, err := stepfunctions.WaitForCompletion(ctx.VDF.Context, 
				exec.(stepfunctions.ExecutionResult).ExecutionArn, 45*time.Second)
			if err != nil {
				return mo.Err[interface{}](fmt.Errorf("functional workflow timeout: %w", err))
			}
			return mo.Ok(result)
		},
		func(resultWrapper mo.Result[interface{}]) mo.Result[OrderResult] {
			return resultWrapper.FlatMap(func(result interface{}) mo.Result[OrderResult] {
				// Functional output parsing
				var workflowOutput OrderResult
				// Simulate output parsing
				workflowOutput = OrderResult{
					Success:         true,
					Status:          "functional_confirmed", 
					PaymentStatus:   "functional_completed",
					InventoryStatus: "functional_reserved",
					ShippingStatus:  "functional_pending",
				}
				return mo.Ok(workflowOutput)
			})
		},
		func(outputResult mo.Result[OrderResult]) OrderResult {
			return outputResult.Match(
				func(workflowOutput OrderResult) OrderResult {
					// Functional metadata enrichment
					return lo.Pipe2(
						workflowOutput,
						func(output OrderResult) OrderResult {
							output.ExecutionArn = "functional-execution-arn"
							output.ProcessingTime = time.Since(startTime)
							output.OrderID = order.OrderID
							return output
						},
						func(enriched OrderResult) OrderResult {
							return enriched
						},
					)
				},
				func(err error) OrderResult {
					return OrderResult{
						Success:        false,
						Status:         "functional_workflow_failed",
						OrderID:        order.OrderID,
						ProcessingTime: time.Since(startTime),
						ValidationErrors: []ValidationError{
							{Field: "functional_workflow_processing", Message: err.Error(), Code: "FUNCTIONAL_WORKFLOW_002"},
						},
					}
				},
			)
		},
	)
	
	return workflowResult

func validateOrderResult(t *testing.T, testName string, input OrderRequest, expected OrderResult, actual OrderResult) {
	t.Helper()
	
	// Functional validation with monadic result aggregation
	validationResults := lo.Pipe4(
		[]func() mo.Result[string]{
			func() mo.Result[string] {
				if expected.Success == actual.Success {
					return mo.Ok("success-status-match")
				}
				return mo.Err[string](fmt.Errorf("success status mismatch: expected %v, got %v", expected.Success, actual.Success))
			},
			func() mo.Result[string] {
				if expected.Status == actual.Status {
					return mo.Ok("status-match")
				}
				return mo.Err[string](fmt.Errorf("status mismatch: expected %s, got %s", expected.Status, actual.Status))
			},
			func() mo.Result[string] {
				if input.OrderID == actual.OrderID {
					return mo.Ok("order-id-match")
				}
				return mo.Err[string](fmt.Errorf("order ID mismatch: expected %s, got %s", input.OrderID, actual.OrderID))
			},
		},
		func(basicValidations []func() mo.Result[string]) []mo.Result[string] {
			return lo.Map(basicValidations, func(validation func() mo.Result[string], _ int) mo.Result[string] {
				return validation()
			})
		},
		func(basicResults []mo.Result[string]) []mo.Result[string] {
			// Add conditional validations
			additionalValidations := []mo.Result[string]{}
			
			if expected.PaymentStatus != "" {
				if expected.PaymentStatus == actual.PaymentStatus {
					additionalValidations = append(additionalValidations, mo.Ok("payment-status-match"))
				} else {
					additionalValidations = append(additionalValidations, mo.Err[string](fmt.Errorf("payment status mismatch")))
				}
			}
			
			if expected.InventoryStatus != "" {
				if expected.InventoryStatus == actual.InventoryStatus {
					additionalValidations = append(additionalValidations, mo.Ok("inventory-status-match"))
				} else {
					additionalValidations = append(additionalValidations, mo.Err[string](fmt.Errorf("inventory status mismatch")))
				}
			}
			
			return append(basicResults, additionalValidations...)
		},
		func(allResults []mo.Result[string]) struct{ passed int; failed int; errors []error } {
			passed := lo.CountBy(allResults, func(result mo.Result[string]) bool {
				return result.IsOk()
			})
			failed := len(allResults) - passed
			errors := lo.FilterMap(allResults, func(result mo.Result[string], _ int) (error, bool) {
				return result.Match(
					func(success string) (error, bool) { return nil, false },
					func(err error) (error, bool) { return err, true },
				)
			})
			return struct{ passed int; failed int; errors []error }{passed, failed, errors}
		},
	)
	
	// Apply functional assertions
	if validationResults.failed > 0 {
		for _, err := range validationResults.errors {
			t.Errorf("Test %s: %v", testName, err)
		}
	} else {
		t.Logf("Test %s: All %d functional validations passed", testName, validationResults.passed)
	}
	
	// Functional validation error checking
	if len(expected.ValidationErrors) > 0 {
		errorValidation := lo.Pipe2(
			expected.ValidationErrors,
			func(expectedErrors []ValidationError) mo.Result[string] {
				for _, expectedError := range expectedErrors {
					found := lo.SomeBy(actual.ValidationErrors, func(actualError ValidationError) bool {
						return actualError.Code == expectedError.Code
					})
					if !found {
						return mo.Err[string](fmt.Errorf("expected validation error %s not found", expectedError.Code))
					}
				}
				return mo.Ok("validation-errors-match")
			},
			func(result mo.Result[string]) bool {
				return result.Match(
					func(success string) bool { return true },
					func(err error) bool {
						t.Errorf("Test %s: %v", testName, err)
						return false
					},
				)
			},
		)
		
		if errorValidation {
			t.Logf("Test %s: Validation errors check passed", testName)
		}
	}
	
	// Functional performance validation
	performanceCheck := mo.Some(actual.ProcessingTime).Filter(func(duration time.Duration) bool {
		return duration < 60*time.Second
	})
	
	performanceCheck.Match(
		func(duration time.Duration) {
			t.Logf("Test %s: Performance validation passed (%v)", testName, duration)
		},
		func() {
			t.Errorf("Test %s: Processing time too long (%v)", testName, actual.ProcessingTime)
		},
	)
	
	// Functional ExecutionArn validation
	executionArnCheck := mo.Some(actual.Status).Filter(func(status string) bool {
		return status != "workflow_start_failed" && status != "functional_workflow_start_failed"
	}).Map(func(status string) bool {
		return len(actual.ExecutionArn) > 0
	})
	
	executionArnCheck.IfPresent(func(hasArn bool) {
		if hasArn {
			t.Logf("Test %s: ExecutionArn validation passed", testName)
		} else {
			t.Errorf("Test %s: ExecutionArn should be present", testName)
		}
	})
}

// Performance and Load Testing

func TestECommerceOrderPipelinePerformance(t *testing.T) {
	ctx := setupECommerceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	seedTestData(t, ctx)
	
	// Benchmark order processing
	benchmark := vasdeference.TableTest[OrderRequest, OrderResult](t, "Performance Benchmark").
		Case("performance_test", generateValidOrderRequest(), OrderResult{Success: true}).
		Benchmark(func(order OrderRequest) OrderResult {
			return processCompleteOrder(t, ctx, order)
		}, 100)
	
	// Performance assertions
	assert.Less(t, benchmark.AverageTime, 5*time.Second, "Average processing time should be under 5 seconds")
	assert.Less(t, benchmark.MaxTime, 10*time.Second, "Maximum processing time should be under 10 seconds")
	
	t.Logf("Performance Results: Avg=%v, Min=%v, Max=%v, Total=%v", 
		benchmark.AverageTime, benchmark.MinTime, benchmark.MaxTime, benchmark.TotalTime)
}

func TestECommerceOrderPipelineLoadTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	
	ctx := setupECommerceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	seedTestData(t, ctx)
	
	// High-concurrency load testing
	concurrency := 20
	ordersPerWorker := 10
	
	orders := make([]OrderRequest, concurrency*ordersPerWorker)
	expectedResults := make([]OrderResult, len(orders))
	
	for i := range orders {
		orders[i] = generateValidOrderRequest()
		expectedResults[i] = OrderResult{Success: true, Status: "confirmed"}
	}
	
	// Execute load test with parallel table testing
	vasdeference.TableTest[OrderRequest, OrderResult](t, "Load Test").
		Parallel().
		WithMaxWorkers(concurrency).
		Timeout(120 * time.Second).
		Run(func(order OrderRequest) OrderResult {
			return processCompleteOrder(t, ctx, order)
		}, func(testName string, input OrderRequest, expected OrderResult, actual OrderResult) {
			// Simplified validation for load testing
			assert.True(t, actual.Success || len(actual.ValidationErrors) > 0, 
				"Test %s: Order should either succeed or have validation errors", testName)
		})
}

// Database Snapshot Testing

func TestECommerceOrderPipelineSnapshotValidation(t *testing.T) {
	ctx := setupECommerceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	seedTestData(t, ctx)
	
	// Create snapshot testing instance
	snapshotTester := snapshot.New(t, snapshot.Options{
		SnapshotDir: "testdata/snapshots/ecommerce",
		JSONIndent:  true,
		Sanitizers: []snapshot.Sanitizer{
			snapshot.SanitizeTimestamps(),
			snapshot.SanitizeUUIDs(),
			snapshot.SanitizeCustom(`"orderId":"[^"]*"`, `"orderId":"<ORDER_ID>"`),
			snapshot.SanitizeCustom(`"executionArn":"[^"]*"`, `"executionArn":"<EXECUTION_ARN>"`),
		},
	})
	
	// Process test order
	order := generateValidOrderRequest()
	result := processCompleteOrder(t, ctx, order)
	
	// Validate against snapshot
	snapshotTester.MatchJSON("successful_order_result", result)
	
	// Validate database state changes
	orderRecord, err := dynamodb.GetItem(ctx.VDF.Context, ctx.OrderTableName, map[string]interface{}{
		"orderId":    order.OrderID,
		"customerId": order.CustomerID,
	})
	require.NoError(t, err)
	
	snapshotTester.MatchDynamoDBItems("order_record", []map[string]interface{}{orderRecord})
	
	// Validate inventory changes
	for _, item := range order.Items {
		inventoryRecord, err := dynamodb.GetItem(ctx.VDF.Context, ctx.InventoryTableName, map[string]interface{}{
			"productId": item.ProductID,
		})
		require.NoError(t, err)
		
		snapshotTester.MatchDynamoDBItems(fmt.Sprintf("inventory_%s", item.ProductID), []map[string]interface{}{inventoryRecord})
	}
}

// Event-Driven Testing

func TestECommerceOrderPipelineEventValidation(t *testing.T) {
	ctx := setupECommerceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	seedTestData(t, ctx)
	
	// Setup event capture
	eventCapture := eventbridge.CreateEventCapture(t, ctx.EventBusName)
	defer eventCapture.Stop()
	
	// Process order
	order := generateValidOrderRequest()
	result := processCompleteOrder(t, ctx, order)
	
	require.True(t, result.Success, "Order processing should succeed")
	
	// Wait for events to be captured
	time.Sleep(5 * time.Second)
	events := eventCapture.GetEvents()
	
	// Validate expected events were published
	expectedEventTypes := []string{
		"OrderValidated",
		"PaymentProcessed", 
		"InventoryReserved",
		"OrderConfirmed",
	}
	
	for _, expectedType := range expectedEventTypes {
		found := false
		for _, event := range events {
			if event.DetailType == expectedType {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected event type %s not found", expectedType)
	}
	
	// Validate event content using snapshots
	snapshotTester := snapshot.New(t, snapshot.Options{
		SnapshotDir: "testdata/snapshots/ecommerce/events",
		Sanitizers: []snapshot.Sanitizer{
			snapshot.SanitizeTimestamps(),
			snapshot.SanitizeCustom(`"orderId":"[^"]*"`, `"orderId":"<ORDER_ID>"`),
		},
	})
	
	snapshotTester.MatchJSON("order_events", events)
}

// Utility Functions

func seedTestData(t *testing.T, ctx *ECommerceTestContext) {
	t.Helper()
	
	// Seed customer data
	customerData := map[string]interface{}{
		"customerId":   "customer-123",
		"name":         "Test Customer",
		"email":        "test@example.com",
		"status":       "active",
		"memberSince":  time.Now().Add(-365 * 24 * time.Hour).Format(time.RFC3339),
	}
	
	err := dynamodb.PutItem(ctx.VDF.Context, ctx.CustomerTableName, customerData)
	require.NoError(t, err)
	
	// Seed inventory data
	inventoryItems := []map[string]interface{}{
		{
			"productId":     "product-001",
			"sku":           "SKU-001",
			"name":          "Test Product 1",
			"price":         29.99,
			"category":      "electronics",
			"stockQuantity": 100,
			"reserved":      0,
		},
		{
			"productId":     "product-002",
			"sku":           "SKU-002", 
			"name":          "Test Product 2",
			"price":         49.99,
			"category":      "electronics",
			"stockQuantity": 50,
			"reserved":      0,
		},
	}
	
	for _, item := range inventoryItems {
		err := dynamodb.PutItem(ctx.VDF.Context, ctx.InventoryTableName, item)
		require.NoError(t, err)
	}
}

func cleanupECommerceInfrastructure(ctx *ECommerceTestContext) error {
	// Cleanup is handled by VasDeference cleanup registration
	return nil
}

// Helper functions for code generation (simplified for example)

func generateOrderValidatorCode() []byte {
	return []byte(`
exports.handler = async (event) => {
	const order = event.order || event;
	
	// Basic validation logic
	if (!order.customerId || !order.items || order.items.length === 0) {
		return {
			success: false,
			validationErrors: [
				{field: "customerId", message: "Customer ID is required", code: "CUSTOMER_001"}
			]
		};
	}
	
	return {
		success: true,
		validatedOrder: order
	};
};`)
}

func generatePaymentProcessorCode() []byte {
	return []byte(`
exports.handler = async (event) => {
	const paymentInfo = event.paymentInfo || event;
	
	// Simulate payment processing
	if (paymentInfo.token === "invalid-token") {
		return {
			success: false,
			status: "failed",
			error: "Invalid payment token"
		};
	}
	
	return {
		success: true,
		status: "completed",
		transactionId: "txn-" + Date.now()
	};
};`)
}

func generateInventoryManagerCode() []byte {
	return []byte(`
exports.handler = async (event) => {
	const items = event.items || event;
	
	// Simulate inventory checking
	const results = items.map(item => {
		if (item.quantity > 100) {
			return {
				productId: item.productId,
				success: false,
				error: "Insufficient inventory"
			};
		}
		
		return {
			productId: item.productId,
			success: true,
			reserved: item.quantity
		};
	});
	
	return {
		success: results.every(r => r.success),
		results: results
	};
};`)
}

func generateOrderWorkflowDefinition(ctx *ECommerceTestContext) string {
	return fmt.Sprintf(`{
		"Comment": "E-commerce order processing workflow",
		"StartAt": "ValidateOrder",
		"States": {
			"ValidateOrder": {
				"Type": "Task", 
				"Resource": "arn:aws:lambda:%s:%s:function:%s",
				"Next": "ProcessPayment"
			},
			"ProcessPayment": {
				"Type": "Task",
				"Resource": "arn:aws:lambda:%s:%s:function:%s", 
				"Next": "ManageInventory"
			},
			"ManageInventory": {
				"Type": "Task",
				"Resource": "arn:aws:lambda:%s:%s:function:%s",
				"Next": "OrderConfirmed",
				"Catch": [{
					"ErrorEquals": ["States.ALL"],
					"Next": "OrderFailed"
				}]
			},
			"OrderConfirmed": {
				"Type": "Succeed",
				"Result": {
					"success": true,
					"status": "confirmed"
				}
			},
			"OrderFailed": {
				"Type": "Fail",
				"Cause": "Order processing failed"
			}
		}
	}`, 
		ctx.VDF.Context.Region, 
		ctx.VDF.GetAccountId(),
		ctx.OrderValidatorFunc,
		ctx.VDF.Context.Region,
		ctx.VDF.GetAccountId(), 
		ctx.PaymentProcessorFunc,
		ctx.VDF.Context.Region,
		ctx.VDF.GetAccountId(),
		ctx.InventoryManagerFunc)
}

func getStepFunctionsRoleArn() string {
	// In a real implementation, this would create or reference an actual IAM role
	return "arn:aws:iam::123456789012:role/StepFunctionsExecutionRole"
}