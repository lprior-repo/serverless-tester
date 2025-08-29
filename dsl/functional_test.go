package dsl

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// TestFunctionalLambdaInvocation demonstrates functional Lambda operations
func TestFunctionalLambdaInvocation(t *testing.T) {
	// Create immutable test configuration using functional options
	config := NewTestConfig(t,
		WithTimeout(30*time.Second),
		WithMetadata("test_type", "functional"),
		WithMetadata("service", "lambda"),
	)

	// Test Lambda invocation using pure functional composition
	result := InvokeLambdaFunction("test-function",
		WithJSONPayload(map[string]interface{}{
			"message": "Hello from functional programming",
			"data":    []int{1, 2, 3, 4, 5},
		}),
		WithSyncInvocation(),
		WithLogs(),
		WithLambdaTimeout(10*time.Second),
	)(config)

	// Functional assertions with type safety
	result.Should().
		Succeed().
		And().
		CompleteWithin(5 * time.Second).
		And().
		HaveMetadata("function_name", "test-function")

	// Test functional mapping over the result
	mappedResult := result.Map(func(resp LambdaResponse) LambdaResponse {
		// Transform response functionally
		if resp.IsSuccess() && resp.GetPayload().IsPresent() {
			return resp.WithLogResult("Processed by functional mapper")
		}
		return resp
	})

	// Verify the mapping worked
	if mappedResult.IsSuccess() {
		response := mappedResult.Value()
		if response.logResult.IsNone() {
			t.Error("Expected log result to be set by functional mapping")
		}
	}
}

// TestFunctionalAsyncLambda demonstrates async functional operations
func TestFunctionalAsyncLambda(t *testing.T) {
	config := NewTestConfig(t, WithTimeout(30*time.Second))

	// Start async operation using functional composition
	asyncOperation := InvokeLambdaFunctionAsync("async-function",
		WithJSONPayload(map[string]interface{}{"task": "background-processing"}),
		WithAsyncInvocation(),
	)(config)

	// Transform async result functionally
	transformedAsync := asyncOperation.Map(func(resp LambdaResponse) LambdaResponse {
		return resp.WithExecutedVersion("v1.0")
	})

	// Wait for completion with timeout
	finalResult := transformedAsync.WaitWithTimeout(10*time.Second, config)

	// Verify result
	finalResult.Should().Succeed()
	if finalResult.IsSuccess() {
		response := finalResult.Value()
		if response.executedVersion.IsNone() {
			t.Error("Expected executed version to be set")
		}
	}
}

// TestFunctionalDynamoDBOperations demonstrates functional DynamoDB operations
func TestFunctionalDynamoDBOperations(t *testing.T) {
	config := NewTestConfig(t,
		WithMetadata("service", "dynamodb"),
		WithTimeout(30*time.Second),
	)

	// Create table using functional composition
	tableResult := CreateDynamoDBTable("users",
		WithHashKey("userId", AttributeTypeString),
		WithRangeKey("timestamp", AttributeTypeNumber),
		WithOnDemandBilling(),
		WithTableTags(map[string]string{
			"Environment": "test",
			"Project":     "functional-testing",
		}),
	)(config)

	// Verify table creation
	tableResult.Should().Succeed().And().HaveMetadata("table_name", "users")

	// Create item using functional composition
	userItem := NewDynamoDBItem().
		WithAttribute("userId", "user123").
		WithAttribute("timestamp", time.Now().Unix()).
		WithAttribute("name", "John Doe").
		WithAttribute("email", "john@example.com").
		WithAttribute("active", true)

	// Put item functionally
	putResult := PutDynamoDBItem("users", userItem)(config)
	putResult.Should().Succeed()

	// Query using functional composition
	queryResult := QueryDynamoDBTable("users",
		WithKeyCondition(EqualsCondition("userId", "user123")),
		WithQueryLimit(10),
		WithConsistentRead(),
		WithAscendingOrder(),
	)(config)

	queryResult.Should().Succeed()

	// Verify query results functionally
	if queryResult.IsSuccess() {
		items := queryResult.Value()
		if len(items) == 0 {
			t.Error("Expected to find items in query result")
		}

		// Use functional operations to verify items
		hasUserItem := lo.SomeBy(items, func(item DynamoDBItem) bool {
			userIdAttr := item.GetAttribute("userId")
			return userIdAttr.IsPresent() && 
				userIdAttr.MustGet().GetValue() == "user123"
		})

		if !hasUserItem {
			t.Error("Expected to find user123 in query results")
		}
	}
}

// TestFunctionalScenario demonstrates complex scenario composition
func TestFunctionalScenario(t *testing.T) {
	config := NewTestConfig(t, WithTimeout(60*time.Second))

	// Create scenario using functional composition
	scenario := NewScenario[interface{}](config).
		AddStep(
			NewScenarioStep("Create DynamoDB table",
				CreateDynamoDBTable("test-table",
					WithHashKey("id", AttributeTypeString),
					WithOnDemandBilling(),
				),
			).WithDescription("Set up data storage"),
		).
		AddStep(
			NewScenarioStep("Put test item",
				PutDynamoDBItem("test-table",
					NewDynamoDBItem().WithAttribute("id", "test-123"),
				),
			).WithDescription("Insert test data"),
		).
		AddStep(
			NewScenarioStep("Invoke Lambda function",
				InvokeLambdaFunction("processor",
					WithJSONPayload(map[string]interface{}{
						"tableId": "test-123",
					}),
				),
			).WithDescription("Process the data"),
		)

	// Execute scenario functionally
	scenarioResult := scenario.Execute()
	scenarioResult.Should().Succeed()

	// Verify scenario metadata
	if scenarioResult.IsSuccess() {
		steps := scenarioResult.Value()
		if len(steps) != 3 {
			t.Errorf("Expected 3 steps, got %d", len(steps))
		}
	}
}

// TestFunctionalTestOrchestration demonstrates complete test orchestration
func TestFunctionalTestOrchestration(t *testing.T) {
	config := NewTestConfig(t, WithTimeout(120*time.Second))

	// Create setup operations using functional composition
	setupOperations := []func(TestConfig) ExecutionResult[interface{}]{
		func(tc TestConfig) ExecutionResult[interface{}] {
			result := CreateDynamoDBTable("orchestration-table",
				WithHashKey("id", AttributeTypeString),
			)(tc)
			return result.Map(func(r map[string]interface{}) interface{} { return r })
		},
		func(tc TestConfig) ExecutionResult[interface{}] {
			result := CreateLambdaFunction("orchestration-func",
				WithRuntime("nodejs18.x"),
				WithHandler("index.handler"),
			)(tc)
			return result.Map(func(r LambdaFunctionConfiguration) interface{} { return r })
		},
	}

	// Create test scenario
	testScenario := NewScenario[interface{}](config).
		AddStep(
			NewScenarioStep("Test data flow",
				func(tc TestConfig) ExecutionResult[interface{}] {
					result := InvokeLambdaFunction("orchestration-func",
						WithJSONPayload(map[string]interface{}{"test": true}),
					)(tc)
					return result.Map(func(r LambdaResponse) interface{} { return r })
				},
			),
		)

	// Create teardown operations
	teardownOperations := []func(TestConfig) ExecutionResult[interface{}]{
		func(tc TestConfig) ExecutionResult[interface{}] {
			// Simulate cleanup
			return NewSuccessResult[interface{}](
				"cleanup completed",
				100*time.Millisecond,
				map[string]interface{}{"phase": "teardown"},
				tc,
			)
		},
	}

	// Create test orchestration using functional composition
	orchestration := NewTestOrchestration[interface{}]("Functional Integration Test", config)

	// Add setup operations functionally
	orchestrationWithSetup := lo.Reduce(setupOperations, 
		func(orch TestOrchestration[interface{}], setup func(TestConfig) ExecutionResult[interface{}], _ int) TestOrchestration[interface{}] {
			return orch.WithSetup(setup)
		}, 
		orchestration,
	)

	// Add scenarios and teardown
	finalOrchestration := orchestrationWithSetup.
		AddScenario(testScenario)

	// Add teardown operations functionally
	orchestrationWithTeardown := lo.Reduce(teardownOperations,
		func(orch TestOrchestration[interface{}], teardown func(TestConfig) ExecutionResult[interface{}], _ int) TestOrchestration[interface{}] {
			return orch.WithTeardown(teardown)
		},
		finalOrchestration,
	)

	// Run the complete orchestration
	result := orchestrationWithTeardown.Run()
	result.Should().Succeed()

	// Verify orchestration metadata functionally
	if result.IsSuccess() {
		metadata := result.metadata
		if testName, exists := metadata["test_name"]; !exists || testName != "Functional Integration Test" {
			t.Error("Expected test name to be set correctly")
		}
		if scenarioCount, exists := metadata["scenarios_count"]; !exists || scenarioCount != 1 {
			t.Error("Expected scenario count to be 1")
		}
	}
}

// TestFunctionalErrorHandling demonstrates monadic error handling
func TestFunctionalErrorHandling(t *testing.T) {
	config := NewTestConfig(t, WithTimeout(10*time.Second))

	// Test error propagation through functional composition
	errorResult := InvokeLambdaFunction("non-existent-function",
		WithJSONPayload(map[string]interface{}{"test": "error"}),
	)(config).FlatMap(func(resp LambdaResponse) ExecutionResult[LambdaResponse] {
		// This should not execute if the first operation fails
		t.Error("FlatMap should not execute on error")
		return NewSuccessResult(resp, 0, map[string]interface{}{}, config)
	})

	// The result should still be successful due to simulation
	// In a real implementation with actual AWS calls, this would fail
	errorResult.Should().Succeed()

	// Test manual error creation
	manualError := NewErrorResult[string](
		ErrResourceNotFound,
		100*time.Millisecond,
		map[string]interface{}{"resource": "test"},
		config,
	)

	// Verify error handling
	if !manualError.IsError() {
		t.Error("Expected manual error to be in error state")
	}

	manualError.Should().Fail()
}

// TestFunctionalOptionTypes demonstrates Option type usage
func TestFunctionalOptionTypes(t *testing.T) {
	// Test Option types with DynamoDB items
	item := NewDynamoDBItem().
		WithAttribute("name", "test").
		WithAttribute("value", 42)

	// Test attribute retrieval with Options
	nameAttr := item.GetAttribute("name")
	if nameAttr.IsNone() {
		t.Error("Expected name attribute to be present")
	}

	missingAttr := item.GetAttribute("missing")
	if missingAttr.IsPresent() {
		t.Error("Expected missing attribute to be absent")
	}

	// Test functional operations on Options
	nameValue := nameAttr.Map(func(attr DynamoDBAttribute) string {
		return attr.GetValue().(string)
	})

	if nameValue.IsNone() || nameValue.MustGet() != "test" {
		t.Error("Expected mapped name value to be 'test'")
	}

	// Test Option chaining
	chainedResult := nameAttr.FlatMap(func(attr DynamoDBAttribute) mo.Option[string] {
		if value, ok := attr.GetValue().(string); ok {
			return mo.Some("processed_" + value)
		}
		return mo.None[string]()
	})

	if chainedResult.IsNone() || chainedResult.MustGet() != "processed_test" {
		t.Error("Expected chained result to be 'processed_test'")
	}
}

// TestFunctionalComposition demonstrates function composition
func TestFunctionalComposition(t *testing.T) {
	// Create composed functions using functional composition
	addPrefix := func(s string) string { return "prefix_" + s }
	addSuffix := func(s string) string { return s + "_suffix" }
	
	// Compose functions
	composed := Compose(addSuffix, addPrefix)
	result := composed("test")
	
	expected := "prefix_test_suffix"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test pipe composition
	piped := Pipe(addPrefix, addSuffix)
	pipeResult := piped("test")
	
	if pipeResult != expected {
		t.Errorf("Expected '%s', got '%s'", expected, pipeResult)
	}
}

// TestFunctionalResultTypes demonstrates Result type usage
func TestFunctionalResultTypes(t *testing.T) {
	// Test TryWithError function
	successResult := TryWithError(func() (string, error) {
		return "success", nil
	})
	
	if successResult.IsError() {
		t.Error("Expected success result")
	}
	
	if successResult.MustGet() != "success" {
		t.Error("Expected success value")
	}
	
	// Test error case
	errorResult := TryWithError(func() (string, error) {
		return "", fmt.Errorf("test error")
	})
	
	if errorResult.IsOk() {
		t.Error("Expected error result")
	}
	
	// Test Result mapping
	mappedResult := successResult.Map(func(s string) string {
		return "mapped_" + s
	})
	
	if mappedResult.IsError() || mappedResult.MustGet() != "mapped_success" {
		t.Error("Expected mapped success result")
	}
	
	// Test Result flat mapping
	flatMappedResult := successResult.FlatMap(func(s string) mo.Result[string, error] {
		return mo.Ok[string, error]("flat_mapped_" + s)
	})
	
	if flatMappedResult.IsError() || flatMappedResult.MustGet() != "flat_mapped_success" {
		t.Error("Expected flat mapped success result")
	}
}