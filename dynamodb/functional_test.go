package dynamodb

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/samber/lo"
)

// TestFunctionalDynamoDBConfig tests functional DynamoDB configuration
func TestFunctionalDynamoDBConfig(t *testing.T) {
	config := NewFunctionalDynamoDBConfig("test-table",
		WithDynamoDBTimeout(45*time.Second),
		WithDynamoDBRetryConfig(5, 2*time.Second),
		WithConsistentRead(true),
		WithBillingMode(types.BillingModeProvisioned),
		WithProjectionExpression("id, #name, active"),
		WithFilterExpression("active = :active"),
		WithConditionExpression("attribute_not_exists(id)"),
		WithDynamoDBMetadata("environment", "test"),
		WithDynamoDBMetadata("service", "dynamodb"),
	)

	if config.tableName != "test-table" {
		t.Errorf("Expected table name 'test-table', got '%s'", config.tableName)
	}

	if config.timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", config.timeout)
	}

	if config.maxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", config.maxRetries)
	}

	if config.retryDelay != 2*time.Second {
		t.Errorf("Expected retry delay 2s, got %v", config.retryDelay)
	}

	if !config.consistentRead {
		t.Error("Expected consistent read to be true")
	}

	if config.billingMode != types.BillingModeProvisioned {
		t.Errorf("Expected billing mode Provisioned, got %v", config.billingMode)
	}

	if config.projectionExpression.IsAbsent() {
		t.Error("Expected projection expression to be present")
	}

	if config.projectionExpression.IsPresent() && config.projectionExpression.MustGet() != "id, #name, active" {
		t.Errorf("Expected projection expression 'id, #name, active', got '%s'", config.projectionExpression.MustGet())
	}

	if config.filterExpression.IsAbsent() {
		t.Error("Expected filter expression to be present")
	}

	if config.conditionExpression.IsAbsent() {
		t.Error("Expected condition expression to be present")
	}

	if config.metadata["environment"] != "test" {
		t.Errorf("Expected environment 'test', got %v", config.metadata["environment"])
	}

	if config.metadata["service"] != "dynamodb" {
		t.Errorf("Expected service 'dynamodb', got %v", config.metadata["service"])
	}
}

// TestFunctionalDynamoDBItem tests functional item creation and manipulation
func TestFunctionalDynamoDBItem(t *testing.T) {
	item := NewFunctionalDynamoDBItem().
		WithAttribute("id", "test-123").
		WithAttribute("name", "Test User").
		WithAttribute("age", 30).
		WithAttribute("active", true).
		WithAttribute("score", 95.5)

	// Test attribute existence
	if !item.HasAttribute("id") {
		t.Error("Expected item to have 'id' attribute")
	}

	if item.HasAttribute("nonexistent") {
		t.Error("Expected item not to have 'nonexistent' attribute")
	}

	// Test attribute retrieval
	idAttr := item.GetAttribute("id")
	if idAttr.IsAbsent() {
		t.Error("Expected id attribute to be present")
	}

	if idAttr.IsPresent() {
		attr := idAttr.MustGet()
		if attr.GetValue() != "test-123" {
			t.Errorf("Expected id value 'test-123', got %v", attr.GetValue())
		}
		if attr.GetType() != types.ScalarAttributeTypeS {
			t.Errorf("Expected string type for id, got %v", attr.GetType())
		}
	}

	// Test numeric attribute
	ageAttr := item.GetAttribute("age")
	if ageAttr.IsPresent() {
		attr := ageAttr.MustGet()
		if attr.GetValue() != 30 {
			t.Errorf("Expected age value 30, got %v", attr.GetValue())
		}
		if attr.GetType() != types.ScalarAttributeTypeN {
			t.Errorf("Expected number type for age, got %v", attr.GetType())
		}
	}

	// Test all attributes
	attributes := item.GetAttributes()
	if len(attributes) != 5 {
		t.Errorf("Expected 5 attributes, got %d", len(attributes))
	}
}

// TestFunctionalDynamoDBItemImmutability tests item immutability
func TestFunctionalDynamoDBItemImmutability(t *testing.T) {
	originalItem := NewFunctionalDynamoDBItem().
		WithAttribute("id", "original").
		WithAttribute("count", 1)

	// Adding attribute should return new item
	newItem := originalItem.WithAttribute("name", "New Attribute")

	// Original item should remain unchanged
	if originalItem.HasAttribute("name") {
		t.Error("Expected original item to not have new attribute")
	}

	// New item should have the attribute
	if !newItem.HasAttribute("name") {
		t.Error("Expected new item to have new attribute")
	}

	// Both should have original attributes
	if !originalItem.HasAttribute("id") || !newItem.HasAttribute("id") {
		t.Error("Expected both items to have original attributes")
	}
}

// TestSimulateFunctionalDynamoDBPutItem tests put item simulation
func TestSimulateFunctionalDynamoDBPutItem(t *testing.T) {
	config := NewFunctionalDynamoDBConfig("test-table",
		WithDynamoDBTimeout(30*time.Second),
		WithDynamoDBMetadata("test_type", "put_item"),
	)

	item := NewFunctionalDynamoDBItem().
		WithAttribute("id", "put-test-123").
		WithAttribute("name", "Put Test Item").
		WithAttribute("active", true)

	result := SimulateFunctionalDynamoDBPutItem(item, config)

	if !result.IsSuccess() {
		t.Errorf("Expected put item to succeed, got error: %v", result.GetError().MustGet())
	}

	if result.GetResult().IsAbsent() {
		t.Error("Expected result to be present")
	}

	// Verify result structure
	if result.IsSuccess() && result.GetResult().IsPresent() {
		resultData, ok := result.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected result to be map[string]interface{}")
		} else {
			if resultData["table_name"] != "test-table" {
				t.Errorf("Expected table_name 'test-table', got %v", resultData["table_name"])
			}

			if resultData["operation"] != "put_item" {
				t.Errorf("Expected operation 'put_item', got %v", resultData["operation"])
			}

			if resultData["attributes"] != 3 {
				t.Errorf("Expected 3 attributes, got %v", resultData["attributes"])
			}
		}
	}

	// Verify metadata
	metadata := result.GetMetadata()
	if metadata["table_name"] != "test-table" {
		t.Errorf("Expected table_name metadata, got %v", metadata["table_name"])
	}

	if metadata["phase"] != "success" {
		t.Errorf("Expected phase metadata to be 'success', got %v", metadata["phase"])
	}

	if metadata["test_type"] != "put_item" {
		t.Errorf("Expected test_type metadata to be 'put_item', got %v", metadata["test_type"])
	}
}

// TestSimulateFunctionalDynamoDBPutItemValidationError tests put item validation errors
func TestSimulateFunctionalDynamoDBPutItemValidationError(t *testing.T) {
	// Test with empty table name
	config := NewFunctionalDynamoDBConfig("")
	emptyItem := NewFunctionalDynamoDBItem()

	result := SimulateFunctionalDynamoDBPutItem(emptyItem, config)

	if result.IsSuccess() {
		t.Error("Expected put item to fail due to validation errors")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	// Verify error metadata
	metadata := result.GetMetadata()
	if metadata["phase"] != "validation" {
		t.Errorf("Expected phase metadata to be 'validation', got %v", metadata["phase"])
	}
}

// TestSimulateFunctionalDynamoDBGetItem tests get item simulation
func TestSimulateFunctionalDynamoDBGetItem(t *testing.T) {
	config := NewFunctionalDynamoDBConfig("users-table",
		WithConsistentRead(true),
		WithDynamoDBMetadata("test_type", "get_item"),
	)

	key := map[string]FunctionalDynamoDBAttribute{
		"id": NewFunctionalDynamoDBAttribute("user-123"),
	}

	result := SimulateFunctionalDynamoDBGetItem(key, config)

	if !result.IsSuccess() {
		t.Errorf("Expected get item to succeed, got error: %v", result.GetError().MustGet())
	}

	if result.GetResult().IsAbsent() {
		t.Error("Expected result to be present")
	}

	// Verify result structure
	if result.IsSuccess() && result.GetResult().IsPresent() {
		item, ok := result.GetResult().MustGet().(FunctionalDynamoDBItem)
		if !ok {
			t.Error("Expected result to be FunctionalDynamoDBItem")
		} else {
			if !item.HasAttribute("id") {
				t.Error("Expected returned item to have 'id' attribute")
			}

			if !item.HasAttribute("name") {
				t.Error("Expected returned item to have 'name' attribute")
			}

			// Verify specific attribute values
			idAttr := item.GetAttribute("id")
			if idAttr.IsPresent() && idAttr.MustGet().GetValue() != "test-id-123" {
				t.Errorf("Expected id 'test-id-123', got %v", idAttr.MustGet().GetValue())
			}
		}
	}

	// Verify metadata
	metadata := result.GetMetadata()
	if metadata["operation"] != "get_item" {
		t.Errorf("Expected operation 'get_item', got %v", metadata["operation"])
	}
}

// TestSimulateFunctionalDynamoDBQuery tests query simulation
func TestSimulateFunctionalDynamoDBQuery(t *testing.T) {
	config := NewFunctionalDynamoDBConfig("items-table",
		WithProjectionExpression("id, #name, active"),
		WithFilterExpression("active = :active"),
		WithDynamoDBMetadata("test_type", "query"),
	)

	keyCondition := map[string]FunctionalDynamoDBAttribute{
		"partition_key": NewFunctionalDynamoDBAttribute("test-partition"),
	}

	result := SimulateFunctionalDynamoDBQuery(keyCondition, config)

	if !result.IsSuccess() {
		t.Errorf("Expected query to succeed, got error: %v", result.GetError().MustGet())
	}

	if result.GetResult().IsAbsent() {
		t.Error("Expected result to be present")
	}

	// Verify result structure
	if result.IsSuccess() && result.GetResult().IsPresent() {
		items, ok := result.GetResult().MustGet().([]FunctionalDynamoDBItem)
		if !ok {
			t.Error("Expected result to be []FunctionalDynamoDBItem")
		} else {
			if len(items) == 0 {
				t.Error("Expected query to return items")
			}

			// Verify first item
			if len(items) > 0 {
				firstItem := items[0]
				if !firstItem.HasAttribute("id") {
					t.Error("Expected first item to have 'id' attribute")
				}

				if !firstItem.HasAttribute("name") {
					t.Error("Expected first item to have 'name' attribute")
				}
			}
		}
	}

	// Verify metadata
	metadata := result.GetMetadata()
	if metadata["operation"] != "query" {
		t.Errorf("Expected operation 'query', got %v", metadata["operation"])
	}

	if metadata["item_count"] == nil {
		t.Error("Expected item_count metadata to be present")
	}
}

// TestFunctionalDynamoDBOperationResultMapping tests result transformation
func TestFunctionalDynamoDBOperationResultMapping(t *testing.T) {
	config := NewFunctionalDynamoDBConfig("mapping-table")
	item := NewFunctionalDynamoDBItem().WithAttribute("id", "mapping-test")

	result := SimulateFunctionalDynamoDBPutItem(item, config)

	// Test mapping successful result
	mappedResult := result.MapResult(func(r interface{}) interface{} {
		if resultMap, ok := r.(map[string]interface{}); ok {
			// Add transformation marker
			resultMap["transformed"] = true
			resultMap["transformation_time"] = time.Now().Unix()
			return resultMap
		}
		return r
	})

	if !mappedResult.IsSuccess() {
		t.Error("Expected mapped result to be successful")
	}

	if mappedResult.GetResult().IsAbsent() {
		t.Error("Expected mapped result to be present")
	}

	// Verify transformation
	if mappedResult.GetResult().IsPresent() {
		transformedResult, ok := mappedResult.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected transformed result to be map[string]interface{}")
		} else {
			if transformedResult["transformed"] != true {
				t.Error("Expected result to be transformed")
			}

			if transformedResult["transformation_time"] == nil {
				t.Error("Expected transformation_time to be present")
			}
		}
	}
}

// TestFunctionalDynamoDBValidation tests validation utilities
func TestFunctionalDynamoDBValidation(t *testing.T) {
	validator := FunctionalDynamoDBValidation()

	// Test table name validation
	err := validator.ValidateTableName("")
	if err == nil {
		t.Error("Expected empty table name to be invalid")
	}

	err = validator.ValidateTableName("valid-table-name")
	if err != nil {
		t.Errorf("Expected valid table name to pass validation, got: %v", err)
	}

	longName := string(make([]byte, 256)) // Longer than 255 characters
	err = validator.ValidateTableName(longName)
	if err == nil {
		t.Error("Expected long table name to be invalid")
	}

	// Test item validation
	emptyItem := NewFunctionalDynamoDBItem()
	err = validator.ValidateItem(emptyItem)
	if err == nil {
		t.Error("Expected empty item to be invalid")
	}

	validItem := NewFunctionalDynamoDBItem().WithAttribute("id", "test")
	err = validator.ValidateItem(validItem)
	if err != nil {
		t.Errorf("Expected valid item to pass validation, got: %v", err)
	}

	// Test key schema validation
	emptyKey := map[string]FunctionalDynamoDBAttribute{}
	err = validator.ValidateKeySchema(emptyKey)
	if err == nil {
		t.Error("Expected empty key schema to be invalid")
	}

	validKey := map[string]FunctionalDynamoDBAttribute{
		"id": NewFunctionalDynamoDBAttribute("test-id"),
	}
	err = validator.ValidateKeySchema(validKey)
	if err != nil {
		t.Errorf("Expected valid key schema to pass validation, got: %v", err)
	}
}

// TestFunctionalDynamoDBRetryWithBackoff tests retry functionality
func TestFunctionalDynamoDBRetryWithBackoff(t *testing.T) {
	// Test successful operation on first try
	attempt := 0
	successResult := FunctionalDynamoDBRetryWithBackoff(
		func() (string, error) {
			attempt++
			return "success", nil
		},
		3,
		10*time.Millisecond,
	)

	if successResult.IsError() {
		t.Error("Expected retry to succeed")
	}

	if successResult.GetResult().IsAbsent() {
		t.Error("Expected result to be present")
	}

	if successResult.GetResult().MustGet() != "success" {
		t.Errorf("Expected result 'success', got '%s'", successResult.GetResult().MustGet())
	}

	if attempt != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempt)
	}

	// Test operation that succeeds after retries
	attempt = 0
	retryResult := FunctionalDynamoDBRetryWithBackoff(
		func() (string, error) {
			attempt++
			if attempt < 3 {
				return "", fmt.Errorf("attempt %d failed", attempt)
			}
			return "success after retries", nil
		},
		5,
		5*time.Millisecond,
	)

	if retryResult.IsError() {
		t.Error("Expected retry to eventually succeed")
	}

	if retryResult.GetResult().MustGet() != "success after retries" {
		t.Errorf("Expected result 'success after retries', got '%s'", retryResult.GetResult().MustGet())
	}

	if attempt != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempt)
	}
}

// TestFunctionalDynamoDBUtilities tests functional utilities used in DynamoDB operations
func TestFunctionalDynamoDBUtilities(t *testing.T) {
	// Test lo.Reduce for options application
	options := []DynamoDBConfigOption{
		WithDynamoDBTimeout(30 * time.Second),
		WithDynamoDBRetryConfig(5, 2*time.Second),
		WithDynamoDBMetadata("test", "value"),
	}

	baseConfig := NewFunctionalDynamoDBConfig("test-table")
	
	finalConfig := lo.Reduce(options, func(config FunctionalDynamoDBConfig, opt DynamoDBConfigOption, _ int) FunctionalDynamoDBConfig {
		return opt(config)
	}, baseConfig)

	if finalConfig.timeout != 30*time.Second {
		t.Error("Expected timeout to be applied via reduce")
	}

	if finalConfig.maxRetries != 5 {
		t.Error("Expected max retries to be applied via reduce")
	}

	if finalConfig.retryDelay != 2*time.Second {
		t.Error("Expected retry delay to be applied via reduce")
	}

	if finalConfig.metadata["test"] != "value" {
		t.Error("Expected metadata to be applied via reduce")
	}

	// Test lo.Assign for metadata merging
	originalMetadata := map[string]interface{}{"key1": "value1"}
	newMetadata := map[string]interface{}{"key2": "value2"}
	
	mergedMetadata := lo.Assign(originalMetadata, newMetadata)
	
	if mergedMetadata["key1"] != "value1" {
		t.Error("Expected original metadata to be preserved")
	}
	
	if mergedMetadata["key2"] != "value2" {
		t.Error("Expected new metadata to be merged")
	}
}

// TestFunctionalValidationPipeline tests the validation pipeline
func TestFunctionalValidationPipeline(t *testing.T) {
	// Test successful pipeline
	pipeline := ExecuteFunctionalValidationPipeline[string]("test-table",
		func(tableName string) (string, error) {
			if len(tableName) == 0 {
				return tableName, fmt.Errorf("table name cannot be empty")
			}
			return tableName, nil
		},
		func(tableName string) (string, error) {
			if len(tableName) > 255 {
				return tableName, fmt.Errorf("table name too long")
			}
			return tableName, nil
		},
	)

	if pipeline.IsError() {
		t.Errorf("Expected pipeline to succeed, got error: %v", pipeline.GetError().MustGet())
	}

	if pipeline.GetValue() != "test-table" {
		t.Errorf("Expected value 'test-table', got '%s'", pipeline.GetValue())
	}

	// Test failing pipeline
	failingPipeline := ExecuteFunctionalValidationPipeline[string]("",
		func(tableName string) (string, error) {
			if len(tableName) == 0 {
				return tableName, fmt.Errorf("table name cannot be empty")
			}
			return tableName, nil
		},
		func(tableName string) (string, error) {
			t.Error("This function should not be called after error")
			return tableName, nil
		},
	)

	if !failingPipeline.IsError() {
		t.Error("Expected pipeline to fail")
	}

	if failingPipeline.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}
}

// BenchmarkFunctionalDynamoDBPutItem benchmarks put item simulation
func BenchmarkFunctionalDynamoDBPutItem(b *testing.B) {
	config := NewFunctionalDynamoDBConfig("benchmark-table",
		WithDynamoDBTimeout(5*time.Second),
		WithDynamoDBRetryConfig(1, 10*time.Millisecond),
	)

	item := NewFunctionalDynamoDBItem().
		WithAttribute("id", "benchmark-123").
		WithAttribute("name", "Benchmark Item").
		WithAttribute("active", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := SimulateFunctionalDynamoDBPutItem(item, config)
		if !result.IsSuccess() {
			b.Errorf("Benchmark iteration failed: %v", result.GetError().MustGet())
		}
	}
}

// BenchmarkFunctionalDynamoDBConfig benchmarks configuration creation
func BenchmarkFunctionalDynamoDBConfig(b *testing.B) {
	options := []DynamoDBConfigOption{
		WithDynamoDBTimeout(30 * time.Second),
		WithDynamoDBRetryConfig(3, 100*time.Millisecond),
		WithConsistentRead(true),
		WithProjectionExpression("id, #name"),
		WithDynamoDBMetadata("environment", "benchmark"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewFunctionalDynamoDBConfig("benchmark-table", options...)
	}
}