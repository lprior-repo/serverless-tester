package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

// Test CRUD operations with proper TDD approach and mocking
// Since the existing functions make direct AWS calls, we'll focus on testing error handling
// and input validation that we can verify without actual AWS calls

func TestPutItemE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for empty table name
	tableName := ""
	item := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	err := PutItemE(t, tableName, item)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestPutItemWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	tableName := ""
	item := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	options := PutItemOptions{
		ConditionExpression: stringPtr("attribute_not_exists(id)"),
	}

	err := PutItemWithOptionsE(t, tableName, item, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestGetItemE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for GetItem
	tableName := ""
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	_, err := GetItemE(t, tableName, key)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestGetItemWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	tableName := ""
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	options := GetItemOptions{
		ConsistentRead: boolPtrForCrud(true),
	}

	_, err := GetItemWithOptionsE(t, tableName, key, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestDeleteItemE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for DeleteItem
	tableName := ""
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	err := DeleteItemE(t, tableName, key)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestDeleteItemWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	tableName := ""
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	options := DeleteItemOptions{
		ConditionExpression: stringPtr("attribute_exists(id)"),
	}

	err := DeleteItemWithOptionsE(t, tableName, key, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestUpdateItemE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for UpdateItem
	tableName := ""
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	updateExpression := "SET #name = :name"
	expressionAttributeNames := map[string]string{
		"#name": "name",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name": &types.AttributeValueMemberS{Value: "updated-name"},
	}

	_, err := UpdateItemE(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestUpdateItemWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	tableName := ""
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	updateExpression := "SET #name = :name"
	expressionAttributeNames := map[string]string{
		"#name": "name",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name": &types.AttributeValueMemberS{Value: "updated-name"},
	}
	options := UpdateItemOptions{
		ConditionExpression: stringPtr("attribute_exists(id)"),
	}

	_, err := UpdateItemWithOptionsE(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestQueryE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for Query
	tableName := ""
	keyConditionExpression := "id = :id"
	expressionAttributeValues := map[string]types.AttributeValue{
		":id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	_, err := QueryE(t, tableName, keyConditionExpression, expressionAttributeValues)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestQueryWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	tableName := ""
	expressionAttributeValues := map[string]types.AttributeValue{
		":id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	options := QueryOptions{
		KeyConditionExpression: stringPtr("id = :id"),
		ConsistentRead:         boolPtrForCrud(true),
	}

	_, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestScanE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for Scan
	tableName := ""

	_, err := ScanE(t, tableName)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestScanWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	tableName := ""
	options := ScanOptions{
		FilterExpression: stringPtr("attribute_exists(active)"),
	}

	_, err := ScanWithOptionsE(t, tableName, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestQueryAllPagesE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for paginated query
	tableName := ""
	keyConditionExpression := "id = :id"
	expressionAttributeValues := map[string]types.AttributeValue{
		":id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	_, err := QueryAllPagesE(t, tableName, keyConditionExpression, expressionAttributeValues)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestQueryAllPagesWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for paginated query with options
	tableName := ""
	expressionAttributeValues := map[string]types.AttributeValue{
		":id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	options := QueryOptions{
		KeyConditionExpression: stringPtr("id = :id"),
	}

	_, err := QueryAllPagesWithOptionsE(t, tableName, expressionAttributeValues, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestScanAllPagesE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for paginated scan
	tableName := ""

	_, err := ScanAllPagesE(t, tableName)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestScanAllPagesWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for paginated scan with options
	tableName := ""
	options := ScanOptions{
		FilterExpression: stringPtr("attribute_exists(active)"),
	}

	_, err := ScanAllPagesWithOptionsE(t, tableName, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

// Test batch operations
func TestBatchWriteItemE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for batch write
	tableName := ""
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id"},
				},
			},
		},
	}

	_, err := BatchWriteItemE(t, tableName, writeRequests)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestBatchWriteItemWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	tableName := ""
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id"},
				},
			},
		},
	}
	options := BatchWriteItemOptions{
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	_, err := BatchWriteItemWithOptionsE(t, tableName, writeRequests, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestBatchGetItemE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for batch get
	tableName := ""
	keys := []map[string]types.AttributeValue{
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-1"},
		},
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-2"},
		},
	}

	_, err := BatchGetItemE(t, tableName, keys)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestBatchGetItemWithOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	tableName := ""
	keys := []map[string]types.AttributeValue{
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-1"},
		},
	}
	options := BatchGetItemOptions{
		ConsistentRead: boolPtrForCrud(true),
	}

	_, err := BatchGetItemWithOptionsE(t, tableName, keys, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

// Test batch operations with retry
func TestBatchWriteItemWithRetryE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for batch write with retry
	tableName := ""
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id"},
				},
			},
		},
	}
	maxRetries := 3

	err := BatchWriteItemWithRetryE(t, tableName, writeRequests, maxRetries)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestBatchGetItemWithRetryE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for batch get with retry
	tableName := ""
	keys := []map[string]types.AttributeValue{
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-1"},
		},
	}
	maxRetries := 3

	_, err := BatchGetItemWithRetryE(t, tableName, keys, maxRetries)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

func TestBatchGetItemWithRetryAndOptionsE_WithEmptyTableName_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for batch get with retry and options
	tableName := ""
	keys := []map[string]types.AttributeValue{
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-1"},
		},
	}
	maxRetries := 3
	options := BatchGetItemOptions{
		ConsistentRead: boolPtrForCrud(true),
	}

	_, err := BatchGetItemWithRetryAndOptionsE(t, tableName, keys, maxRetries, options)

	// Should return an error due to empty table name
	assert.Error(t, err)
}

// Test transaction operations
func TestTransactWriteItemsE_WithEmptyTransactItems_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for transaction write
	var transactItems []types.TransactWriteItem

	_, err := TransactWriteItemsE(t, transactItems)

	// Should return an error due to empty transaction items
	assert.Error(t, err)
}

func TestTransactWriteItemsWithOptionsE_WithEmptyTransactItems_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	var transactItems []types.TransactWriteItem
	options := TransactWriteItemsOptions{
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	_, err := TransactWriteItemsWithOptionsE(t, transactItems, options)

	// Should return an error due to empty transaction items
	assert.Error(t, err)
}

func TestTransactGetItemsE_WithEmptyTransactItems_ShouldReturnError(t *testing.T) {
	// RED: Test input validation for transaction get
	var transactItems []types.TransactGetItem

	_, err := TransactGetItemsE(t, transactItems)

	// Should return an error due to empty transaction items
	assert.Error(t, err)
}

func TestTransactGetItemsWithOptionsE_WithEmptyTransactItems_ShouldReturnError(t *testing.T) {
	// RED: Test input validation with options
	var transactItems []types.TransactGetItem
	options := TransactGetItemsOptions{
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	_, err := TransactGetItemsWithOptionsE(t, transactItems, options)

	// Should return an error due to empty transaction items
	assert.Error(t, err)
}

// Helper functions for tests
func boolPtrForCrud(b bool) *bool {
	return &b
}