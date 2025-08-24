package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

// TestExampleBasicCRUDOperations demonstrates basic CRUD operations
func TestExampleBasicCRUDOperations(t *testing.T) {
	tableName := "users"
	userID := "user-123"
	
	// Create an item
	item := map[string]types.AttributeValue{
		"id":       &types.AttributeValueMemberS{Value: userID},
		"name":     &types.AttributeValueMemberS{Value: "John Doe"},
		"email":    &types.AttributeValueMemberS{Value: "john.doe@example.com"},
		"age":      &types.AttributeValueMemberN{Value: "30"},
		"active":   &types.AttributeValueMemberBOOL{Value: true},
	}
	
	PutItem(t, tableName, item)
	
	// Verify item exists
	AssertItemExists(t, tableName, map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: userID},
	})
	
	// Read the item
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: userID},
	}
	retrievedItem := GetItem(t, tableName, key)
	
	assert.NotNil(t, retrievedItem)
	assert.Equal(t, "John Doe", retrievedItem["name"].(*types.AttributeValueMemberS).Value)
	
	// Update the item
	updateExpression := "SET #name = :name, age = age + :increment"
	expressionAttributeNames := map[string]string{
		"#name": "name",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name":      &types.AttributeValueMemberS{Value: "John Smith"},
		":increment": &types.AttributeValueMemberN{Value: "1"},
	}
	
	updatedItem := UpdateItem(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues)
	assert.NotNil(t, updatedItem)
	
	// Verify the update
	AssertAttributeEquals(t, tableName, key, "name", &types.AttributeValueMemberS{Value: "John Smith"})
	
	// Delete the item
	DeleteItem(t, tableName, key)
	
	// Verify item was deleted
	AssertItemNotExists(t, tableName, key)
}

// TestExampleConditionalOperations demonstrates conditional operations
func TestExampleConditionalOperations(t *testing.T) {
	tableName := "users"
	userID := "user-456"
	
	// Put item with condition (should succeed)
	item := map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: userID},
		"name": &types.AttributeValueMemberS{Value: "Jane Doe"},
	}
	options := PutItemOptions{
		ConditionExpression: stringPtr("attribute_not_exists(id)"),
	}
	
	PutItemWithOptions(t, tableName, item, options)
	
	// Try to put the same item again (should fail)
	err := PutItemWithOptionsE(t, tableName, item, options)
	AssertConditionalCheckFailed(t, err)
	
	// Update with condition
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: userID},
	}
	updateOptions := UpdateItemOptions{
		ConditionExpression: stringPtr("attribute_exists(id)"),
		ReturnValues:        types.ReturnValueAllNew,
	}
	
	updatedItem := UpdateItemWithOptions(t, tableName, key, "SET age = :age", nil, map[string]types.AttributeValue{
		":age": &types.AttributeValueMemberN{Value: "25"},
	}, updateOptions)
	
	assert.NotNil(t, updatedItem)
	
	// Clean up
	DeleteItem(t, tableName, key)
}

// TestExampleQueryOperations demonstrates query operations with pagination
func TestExampleQueryOperations(t *testing.T) {
	tableName := "orders"
	customerID := "customer-789"
	
	// Create test data
	orders := []map[string]types.AttributeValue{
		{
			"customer_id": &types.AttributeValueMemberS{Value: customerID},
			"order_id":    &types.AttributeValueMemberS{Value: "order-001"},
			"amount":      &types.AttributeValueMemberN{Value: "99.99"},
			"status":      &types.AttributeValueMemberS{Value: "completed"},
		},
		{
			"customer_id": &types.AttributeValueMemberS{Value: customerID},
			"order_id":    &types.AttributeValueMemberS{Value: "order-002"},
			"amount":      &types.AttributeValueMemberN{Value: "149.99"},
			"status":      &types.AttributeValueMemberS{Value: "pending"},
		},
		{
			"customer_id": &types.AttributeValueMemberS{Value: customerID},
			"order_id":    &types.AttributeValueMemberS{Value: "order-003"},
			"amount":      &types.AttributeValueMemberN{Value: "199.99"},
			"status":      &types.AttributeValueMemberS{Value: "completed"},
		},
	}
	
	// Put all orders
	for _, order := range orders {
		PutItem(t, tableName, order)
	}
	
	// Query all orders for customer
	keyConditionExpression := "customer_id = :customer_id"
	expressionAttributeValues := map[string]types.AttributeValue{
		":customer_id": &types.AttributeValueMemberS{Value: customerID},
	}
	
	result := Query(t, tableName, keyConditionExpression, expressionAttributeValues)
	assert.Equal(t, int32(3), result.Count)
	
	// Query with filter expression (only completed orders)
	queryOptions := QueryOptions{
		FilterExpression: stringPtr("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		KeyConditionExpression: &keyConditionExpression,
	}
	
	filteredExpressionsAttributeValues := map[string]types.AttributeValue{
		":customer_id": &types.AttributeValueMemberS{Value: customerID},
		":status":      &types.AttributeValueMemberS{Value: "completed"},
	}
	
	filteredResult := QueryWithOptions(t, tableName, keyConditionExpression, filteredExpressionsAttributeValues, queryOptions)
	assert.Equal(t, int32(2), filteredResult.Count)
	
	// Query with pagination (limit 1 item at a time)
	paginationOptions := QueryOptions{
		Limit:                  int32Ptr(1),
		KeyConditionExpression: &keyConditionExpression,
	}
	
	var allItems []map[string]types.AttributeValue
	var exclusiveStartKey map[string]types.AttributeValue
	
	for {
		paginationOptions.ExclusiveStartKey = exclusiveStartKey
		pageResult := QueryWithOptions(t, tableName, keyConditionExpression, expressionAttributeValues, paginationOptions)
		
		allItems = append(allItems, pageResult.Items...)
		
		if pageResult.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = pageResult.LastEvaluatedKey
	}
	
	assert.Equal(t, 3, len(allItems))
	
	// Use convenience method for getting all pages
	allItemsPaginated := QueryAllPages(t, tableName, keyConditionExpression, expressionAttributeValues)
	assert.Equal(t, 3, len(allItemsPaginated))
	
	// Clean up
	for _, order := range orders {
		key := map[string]types.AttributeValue{
			"customer_id": order["customer_id"],
			"order_id":    order["order_id"],
		}
		DeleteItem(t, tableName, key)
	}
}

// TestExampleBatchOperations demonstrates batch read and write operations
func TestExampleBatchOperations(t *testing.T) {
	tableName := "products"
	
	// Prepare batch write requests
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: "product-001"},
					"name":  &types.AttributeValueMemberS{Value: "Widget A"},
					"price": &types.AttributeValueMemberN{Value: "19.99"},
				},
			},
		},
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: "product-002"},
					"name":  &types.AttributeValueMemberS{Value: "Widget B"},
					"price": &types.AttributeValueMemberN{Value: "29.99"},
				},
			},
		},
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: "product-003"},
					"name":  &types.AttributeValueMemberS{Value: "Widget C"},
					"price": &types.AttributeValueMemberN{Value: "39.99"},
				},
			},
		},
	}
	
	// Write items in batch
	result := BatchWriteItem(t, tableName, writeRequests)
	assert.NotNil(t, result)
	
	// Use retry mechanism for robustness
	BatchWriteItemWithRetry(t, tableName, writeRequests, 3)
	
	// Batch read items
	keys := []map[string]types.AttributeValue{
		{"id": &types.AttributeValueMemberS{Value: "product-001"}},
		{"id": &types.AttributeValueMemberS{Value: "product-002"}},
		{"id": &types.AttributeValueMemberS{Value: "product-003"}},
	}
	
	batchResult := BatchGetItem(t, tableName, keys)
	assert.Equal(t, 3, len(batchResult.Items))
	
	// Use retry mechanism for batch reads
	allItems := BatchGetItemWithRetry(t, tableName, keys, 3)
	assert.Equal(t, 3, len(allItems))
	
	// Clean up with batch delete
	deleteRequests := []types.WriteRequest{
		{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "product-001"},
				},
			},
		},
		{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "product-002"},
				},
			},
		},
		{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "product-003"},
				},
			},
		},
	}
	
	BatchWriteItem(t, tableName, deleteRequests)
}

// TestExampleTransactionalOperations demonstrates transactional operations
func TestExampleTransactionalOperations(t *testing.T) {
	tableName := "accounts"
	
	// Setup test accounts
	account1Item := map[string]types.AttributeValue{
		"id":      &types.AttributeValueMemberS{Value: "account-001"},
		"balance": &types.AttributeValueMemberN{Value: "1000"},
		"name":    &types.AttributeValueMemberS{Value: "Alice"},
	}
	account2Item := map[string]types.AttributeValue{
		"id":      &types.AttributeValueMemberS{Value: "account-002"},
		"balance": &types.AttributeValueMemberN{Value: "500"},
		"name":    &types.AttributeValueMemberS{Value: "Bob"},
	}
	
	PutItem(t, tableName, account1Item)
	PutItem(t, tableName, account2Item)
	
	// Perform transactional transfer (Alice sends $100 to Bob)
	transferAmount := "100"
	
	transactItems := []types.TransactWriteItem{
		{
			Update: &types.Update{
				TableName: &tableName,
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "account-001"},
				},
				UpdateExpression: stringPtr("SET balance = balance - :amount"),
				ConditionExpression: stringPtr("balance >= :amount"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":amount": &types.AttributeValueMemberN{Value: transferAmount},
				},
			},
		},
		{
			Update: &types.Update{
				TableName: &tableName,
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "account-002"},
				},
				UpdateExpression: stringPtr("SET balance = balance + :amount"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":amount": &types.AttributeValueMemberN{Value: transferAmount},
				},
			},
		},
	}
	
	// Execute the transaction
	transactResult := TransactWriteItems(t, transactItems)
	assert.NotNil(t, transactResult)
	
	// Verify the balances after transaction
	key1 := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "account-001"}}
	key2 := map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "account-002"}}
	
	AssertAttributeEquals(t, tableName, key1, "balance", &types.AttributeValueMemberN{Value: "900"})
	AssertAttributeEquals(t, tableName, key2, "balance", &types.AttributeValueMemberN{Value: "600"})
	
	// Transactional read
	transactReadItems := []types.TransactGetItem{
		{
			Get: &types.Get{
				TableName: &tableName,
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "account-001"},
				},
			},
		},
		{
			Get: &types.Get{
				TableName: &tableName,
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "account-002"},
				},
			},
		},
	}
	
	readResult := TransactGetItems(t, transactReadItems)
	assert.Equal(t, 2, len(readResult.Responses))
	
	// Clean up
	DeleteItem(t, tableName, key1)
	DeleteItem(t, tableName, key2)
}

// TestExampleAssertions demonstrates the testing assertion functions
func TestExampleAssertions(t *testing.T) {
	tableName := "test-table"
	
	// Assert table exists (assuming it's been created)
	AssertTableExists(t, tableName)
	
	// Assert table status
	AssertTableStatus(t, tableName, types.TableStatusActive)
	
	// Create a test item
	item := map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: "test-item"},
		"name": &types.AttributeValueMemberS{Value: "Test Item"},
	}
	PutItem(t, tableName, item)
	
	// Test item assertions
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-item"},
	}
	
	AssertItemExists(t, tableName, key)
	AssertAttributeEquals(t, tableName, key, "name", &types.AttributeValueMemberS{Value: "Test Item"})
	
	// Test query count assertion
	keyConditionExpression := "id = :id"
	expressionAttributeValues := map[string]types.AttributeValue{
		":id": &types.AttributeValueMemberS{Value: "test-item"},
	}
	AssertQueryCount(t, tableName, keyConditionExpression, expressionAttributeValues, 1)
	
	// Test conditional operation assertions
	updateOptions := UpdateItemOptions{
		ConditionExpression: stringPtr("attribute_exists(id)"),
	}
	
	_, err := UpdateItemWithOptionsE(t, tableName, key, "SET description = :desc", nil, map[string]types.AttributeValue{
		":desc": &types.AttributeValueMemberS{Value: "Updated description"},
	}, updateOptions)
	
	AssertConditionalCheckPassed(t, err)
	
	// Clean up
	DeleteItem(t, tableName, key)
	AssertItemNotExists(t, tableName, key)
}

// TestExampleScanOperations demonstrates scan operations with filtering
func TestExampleScanOperations(t *testing.T) {
	tableName := "inventory"
	
	// Create test data
	items := []map[string]types.AttributeValue{
		{
			"id":       &types.AttributeValueMemberS{Value: "item-001"},
			"category": &types.AttributeValueMemberS{Value: "electronics"},
			"stock":    &types.AttributeValueMemberN{Value: "50"},
			"price":    &types.AttributeValueMemberN{Value: "299.99"},
		},
		{
			"id":       &types.AttributeValueMemberS{Value: "item-002"},
			"category": &types.AttributeValueMemberS{Value: "electronics"},
			"stock":    &types.AttributeValueMemberN{Value: "0"},
			"price":    &types.AttributeValueMemberN{Value: "199.99"},
		},
		{
			"id":       &types.AttributeValueMemberS{Value: "item-003"},
			"category": &types.AttributeValueMemberS{Value: "books"},
			"stock":    &types.AttributeValueMemberN{Value: "25"},
			"price":    &types.AttributeValueMemberN{Value: "24.99"},
		},
	}
	
	// Put all items
	for _, item := range items {
		PutItem(t, tableName, item)
	}
	
	// Scan all items
	result := Scan(t, tableName)
	assert.True(t, result.Count >= 3)
	
	// Scan with filter for electronics category
	scanOptions := ScanOptions{
		FilterExpression: stringPtr("category = :category"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":category": &types.AttributeValueMemberS{Value: "electronics"},
		},
	}
	
	filteredResult := ScanWithOptions(t, tableName, scanOptions)
	assert.Equal(t, int32(2), filteredResult.Count)
	
	// Scan with filter for in-stock items
	stockOptions := ScanOptions{
		FilterExpression: stringPtr("stock > :zero"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":zero": &types.AttributeValueMemberN{Value: "0"},
		},
	}
	
	inStockResult := ScanWithOptions(t, tableName, stockOptions)
	assert.Equal(t, int32(2), inStockResult.Count)
	
	// Use scan with pagination
	allItems := ScanAllPages(t, tableName)
	assert.True(t, len(allItems) >= 3)
	
	// Verify item count
	AssertItemCount(t, tableName, len(items))
	
	// Clean up
	for _, item := range items {
		key := map[string]types.AttributeValue{
			"id": item["id"],
		}
		DeleteItem(t, tableName, key)
	}
}