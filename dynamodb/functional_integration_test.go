package dynamodb

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/samber/lo"
)

// TestFunctionalDynamoDBScenario tests a complete DynamoDB scenario
func TestFunctionalDynamoDBScenario(t *testing.T) {
	// Create table configuration
	config := NewFunctionalDynamoDBConfig("users-table",
		WithDynamoDBTimeout(30*time.Second),
		WithDynamoDBRetryConfig(3, 200*time.Millisecond),
		WithConsistentRead(true),
		WithBillingMode(types.BillingModePayPerRequest),
		WithDynamoDBMetadata("environment", "test"),
		WithDynamoDBMetadata("scenario", "full_workflow"),
	)

	// Step 1: Create and put multiple items
	users := []FunctionalDynamoDBItem{
		NewFunctionalDynamoDBItem().
			WithAttribute("id", "user-001").
			WithAttribute("name", "Alice Johnson").
			WithAttribute("email", "alice@example.com").
			WithAttribute("active", true).
			WithAttribute("role", "admin").
			WithAttribute("score", 95),
		
		NewFunctionalDynamoDBItem().
			WithAttribute("id", "user-002").
			WithAttribute("name", "Bob Smith").
			WithAttribute("email", "bob@example.com").
			WithAttribute("active", true).
			WithAttribute("role", "user").
			WithAttribute("score", 87),
		
		NewFunctionalDynamoDBItem().
			WithAttribute("id", "user-003").
			WithAttribute("name", "Charlie Brown").
			WithAttribute("email", "charlie@example.com").
			WithAttribute("active", false).
			WithAttribute("role", "user").
			WithAttribute("score", 72),
	}

	// Put each user item
	putResults := lo.Map(users, func(user FunctionalDynamoDBItem, index int) FunctionalDynamoDBOperationResult {
		return SimulateFunctionalDynamoDBPutItem(user, config)
	})

	// Verify all put operations succeeded
	allPutsSucceeded := lo.EveryBy(putResults, func(result FunctionalDynamoDBOperationResult) bool {
		return result.IsSuccess()
	})

	if !allPutsSucceeded {
		t.Error("Expected all put operations to succeed")
	}

	// Step 2: Get individual items
	getUserResults := []FunctionalDynamoDBOperationResult{}
	for i := 1; i <= 3; i++ {
		key := map[string]FunctionalDynamoDBAttribute{
			"id": NewFunctionalDynamoDBAttribute(fmt.Sprintf("user-%03d", i)),
		}
		result := SimulateFunctionalDynamoDBGetItem(key, config)
		getUserResults = append(getUserResults, result)
	}

	// Verify all get operations succeeded
	allGetsSucceeded := lo.EveryBy(getUserResults, func(result FunctionalDynamoDBOperationResult) bool {
		return result.IsSuccess()
	})

	if !allGetsSucceeded {
		t.Error("Expected all get operations to succeed")
	}

	// Step 3: Query for active users
	queryConfig := NewFunctionalDynamoDBConfig("users-table",
		WithDynamoDBTimeout(30*time.Second),
		WithProjectionExpression("id, #name, active, #role"),
		WithFilterExpression("active = :active_val"),
		WithDynamoDBMetadata("environment", "test"), // Add missing environment metadata
		WithDynamoDBMetadata("scenario", "full_workflow"), // Add scenario metadata
		WithDynamoDBMetadata("query_type", "active_users"),
	)

	keyCondition := map[string]FunctionalDynamoDBAttribute{
		"status": NewFunctionalDynamoDBAttribute("active"),
	}

	queryResult := SimulateFunctionalDynamoDBQuery(keyCondition, queryConfig)

	if !queryResult.IsSuccess() {
		t.Errorf("Expected query to succeed, got error: %v", queryResult.GetError().MustGet())
	}

	// Verify query results
	if queryResult.GetResult().IsPresent() {
		items, ok := queryResult.GetResult().MustGet().([]FunctionalDynamoDBItem)
		if !ok {
			t.Error("Expected query result to be []FunctionalDynamoDBItem")
		} else {
			if len(items) == 0 {
				t.Error("Expected query to return items")
			}

			// Verify each returned item has required attributes
			for i, item := range items {
				if !item.HasAttribute("id") {
					t.Errorf("Item %d missing 'id' attribute", i)
				}
				if !item.HasAttribute("name") {
					t.Errorf("Item %d missing 'name' attribute", i)
				}
				if !item.HasAttribute("active") {
					t.Errorf("Item %d missing 'active' attribute", i)
				}
			}
		}
	}

	// Step 4: Verify metadata accumulation across operations
	allMetadata := []map[string]interface{}{}
	for _, result := range putResults {
		allMetadata = append(allMetadata, result.GetMetadata())
	}
	for _, result := range getUserResults {
		allMetadata = append(allMetadata, result.GetMetadata())
	}
	allMetadata = append(allMetadata, queryResult.GetMetadata())

	// Check that all operations preserved environment metadata
	for i, metadata := range allMetadata {
		if metadata["environment"] != "test" {
			t.Errorf("Operation %d missing environment metadata, metadata: %+v", i, metadata)
		}
	}
}

// TestFunctionalDynamoDBErrorHandling tests comprehensive error handling
func TestFunctionalDynamoDBErrorHandling(t *testing.T) {
	// Test invalid table name
	invalidConfig := NewFunctionalDynamoDBConfig("")
	validItem := NewFunctionalDynamoDBItem().WithAttribute("id", "test")

	result := SimulateFunctionalDynamoDBPutItem(validItem, invalidConfig)

	if result.IsSuccess() {
		t.Error("Expected operation to fail with invalid table name")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	// Test empty item
	validConfig := NewFunctionalDynamoDBConfig("test-table")
	emptyItem := NewFunctionalDynamoDBItem()

	result2 := SimulateFunctionalDynamoDBPutItem(emptyItem, validConfig)

	if result2.IsSuccess() {
		t.Error("Expected operation to fail with empty item")
	}

	// Test invalid key for get operation
	emptyKey := map[string]FunctionalDynamoDBAttribute{}
	result3 := SimulateFunctionalDynamoDBGetItem(emptyKey, validConfig)

	if result3.IsSuccess() {
		t.Error("Expected get operation to fail with empty key")
	}

	// Verify error metadata consistency
	errorResults := []FunctionalDynamoDBOperationResult{result, result2, result3}
	for i, errorResult := range errorResults {
		metadata := errorResult.GetMetadata()
		if metadata["phase"] != "validation" {
			t.Errorf("Error result %d should have validation phase, got %v", i, metadata["phase"])
		}
	}
}

// TestFunctionalDynamoDBConfigurationCombinations tests various config combinations
func TestFunctionalDynamoDBConfigurationCombinations(t *testing.T) {
	// Test minimal configuration
	minimalConfig := NewFunctionalDynamoDBConfig("minimal-table")
	
	if minimalConfig.tableName != "minimal-table" {
		t.Error("Expected minimal config to have correct table name")
	}
	
	if minimalConfig.timeout != FunctionalDynamoDBDefaultTimeout {
		t.Error("Expected minimal config to use default timeout")
	}

	// Test maximal configuration
	maximalConfig := NewFunctionalDynamoDBConfig("maximal-table",
		WithDynamoDBTimeout(120*time.Second),
		WithDynamoDBRetryConfig(10, 5*time.Second),
		WithConsistentRead(true),
		WithBillingMode(types.BillingModeProvisioned),
		WithProjectionExpression("id, #name, email, active"),
		WithFilterExpression("active = :active AND #role = :role"),
		WithConditionExpression("attribute_not_exists(id)"),
		WithDynamoDBMetadata("service", "user-service"),
		WithDynamoDBMetadata("environment", "production"),
		WithDynamoDBMetadata("version", "v2.1.0"),
	)

	// Verify all options were applied
	if maximalConfig.tableName != "maximal-table" {
		t.Error("Expected maximal config table name")
	}

	if maximalConfig.timeout != 120*time.Second {
		t.Error("Expected maximal config timeout")
	}

	if maximalConfig.maxRetries != 10 {
		t.Error("Expected maximal config retries")
	}

	if maximalConfig.retryDelay != 5*time.Second {
		t.Error("Expected maximal config retry delay")
	}

	if !maximalConfig.consistentRead {
		t.Error("Expected maximal config consistent read")
	}

	if maximalConfig.billingMode != types.BillingModeProvisioned {
		t.Error("Expected maximal config billing mode")
	}

	if maximalConfig.projectionExpression.IsAbsent() {
		t.Error("Expected maximal config projection expression")
	}

	if maximalConfig.filterExpression.IsAbsent() {
		t.Error("Expected maximal config filter expression")
	}

	if maximalConfig.conditionExpression.IsAbsent() {
		t.Error("Expected maximal config condition expression")
	}

	if len(maximalConfig.metadata) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(maximalConfig.metadata))
	}
}

// TestFunctionalDynamoDBItemManipulation tests advanced item operations
func TestFunctionalDynamoDBItemManipulation(t *testing.T) {
	// Create a complex item with various data types
	complexItem := NewFunctionalDynamoDBItem().
		WithAttribute("id", "complex-001").
		WithAttribute("name", "Complex Test Item").
		WithAttribute("count", 42).
		WithAttribute("score", 98.7).
		WithAttribute("active", true).
		WithAttribute("tags", "important,test,complex").
		WithAttribute("created_at", time.Now().Unix())

	// Test attribute type inference
	attributes := complexItem.GetAttributes()
	
	// String attribute
	idAttr := attributes["id"]
	if idAttr.GetType() != types.ScalarAttributeTypeS {
		t.Errorf("Expected string type for id, got %v", idAttr.GetType())
	}

	// Numeric attributes
	countAttr := attributes["count"]
	if countAttr.GetType() != types.ScalarAttributeTypeN {
		t.Errorf("Expected number type for count, got %v", countAttr.GetType())
	}

	scoreAttr := attributes["score"]
	if scoreAttr.GetType() != types.ScalarAttributeTypeN {
		t.Errorf("Expected number type for score, got %v", scoreAttr.GetType())
	}

	// Test item operations with this complex item
	config := NewFunctionalDynamoDBConfig("complex-table",
		WithDynamoDBMetadata("item_type", "complex"),
	)

	result := SimulateFunctionalDynamoDBPutItem(complexItem, config)

	if !result.IsSuccess() {
		t.Errorf("Expected complex item put to succeed, got error: %v", result.GetError().MustGet())
	}

	// Verify result includes attribute count
	if result.GetResult().IsPresent() {
		resultMap, ok := result.GetResult().MustGet().(map[string]interface{})
		if ok && resultMap["attributes"] != 7 {
			t.Errorf("Expected 7 attributes in result, got %v", resultMap["attributes"])
		}
	}
}

// TestFunctionalDynamoDBPerformanceMetrics tests performance tracking
func TestFunctionalDynamoDBPerformanceMetrics(t *testing.T) {
	config := NewFunctionalDynamoDBConfig("performance-table",
		WithDynamoDBTimeout(1*time.Second),
		WithDynamoDBRetryConfig(1, 50*time.Millisecond), // Faster for testing
		WithDynamoDBMetadata("test_type", "performance"),
	)

	item := NewFunctionalDynamoDBItem().
		WithAttribute("id", "perf-001").
		WithAttribute("data", "performance test data")

	startTime := time.Now()
	result := SimulateFunctionalDynamoDBPutItem(item, config)
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
	results := []FunctionalDynamoDBOperationResult{}
	for i := 0; i < 5; i++ {
		testItem := NewFunctionalDynamoDBItem().
			WithAttribute("id", fmt.Sprintf("perf-%03d", i)).
			WithAttribute("index", i)
		
		result := SimulateFunctionalDynamoDBPutItem(testItem, config)
		results = append(results, result)
	}

	// All operations should succeed
	allSucceeded := lo.EveryBy(results, func(r FunctionalDynamoDBOperationResult) bool {
		return r.IsSuccess()
	})

	if !allSucceeded {
		t.Error("Expected all performance test operations to succeed")
	}

	// All operations should have duration recorded
	allHaveDuration := lo.EveryBy(results, func(r FunctionalDynamoDBOperationResult) bool {
		return r.GetDuration() > 0
	})

	if !allHaveDuration {
		t.Error("Expected all operations to have duration recorded")
	}
}

// BenchmarkFunctionalDynamoDBScenario benchmarks a complete scenario
func BenchmarkFunctionalDynamoDBScenario(b *testing.B) {
	config := NewFunctionalDynamoDBConfig("benchmark-table",
		WithDynamoDBTimeout(5*time.Second),
		WithDynamoDBRetryConfig(1, 10*time.Millisecond),
	)

	item := NewFunctionalDynamoDBItem().
		WithAttribute("id", "bench-item").
		WithAttribute("name", "Benchmark Item").
		WithAttribute("active", true)

	key := map[string]FunctionalDynamoDBAttribute{
		"id": NewFunctionalDynamoDBAttribute("bench-item"),
	}

	queryKey := map[string]FunctionalDynamoDBAttribute{
		"status": NewFunctionalDynamoDBAttribute("active"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate full CRUD cycle
		putResult := SimulateFunctionalDynamoDBPutItem(item, config)
		if !putResult.IsSuccess() {
			b.Errorf("Put operation failed: %v", putResult.GetError().MustGet())
		}

		getResult := SimulateFunctionalDynamoDBGetItem(key, config)
		if !getResult.IsSuccess() {
			b.Errorf("Get operation failed: %v", getResult.GetError().MustGet())
		}

		queryResult := SimulateFunctionalDynamoDBQuery(queryKey, config)
		if !queryResult.IsSuccess() {
			b.Errorf("Query operation failed: %v", queryResult.GetError().MustGet())
		}
	}
}