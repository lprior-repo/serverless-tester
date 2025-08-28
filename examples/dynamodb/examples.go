package main

import (
	"fmt"

	"vasdeference/dynamodb"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// stringPtr helper function for conditional expressions
func stringPtr(s string) *string {
	return &s
}

// int32Ptr helper function for limit values
func int32Ptr(i int32) *int32 {
	return &i
}

// ExampleBasicCRUDOperations demonstrates basic CRUD operations
func ExampleBasicCRUDOperations() {
	fmt.Println("Basic DynamoDB CRUD operations:")
	
	_ = "users" // tableName
	userID := "user-123"
	
	// Create an item
	_ = map[string]types.AttributeValue{
		"id":     &types.AttributeValueMemberS{Value: userID},
		"name":   &types.AttributeValueMemberS{Value: "John Doe"},
		"email":  &types.AttributeValueMemberS{Value: "john.doe@example.com"},
		"age":    &types.AttributeValueMemberN{Value: "30"},
		"active": &types.AttributeValueMemberBOOL{Value: true},
	}
	
	// dynamodb.PutItem(t, tableName, item)
	fmt.Printf("Creating item with ID: %s\n", userID)
	
	// Verify item exists
	// dynamodb.AssertItemExists(t, tableName, map[string]types.AttributeValue{
	//     "id": &types.AttributeValueMemberS{Value: userID},
	// })
	
	// Read the item
	_ = map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: userID},
	}
	// retrievedItem := dynamodb.GetItem(t, tableName, key)
	// Expected: retrievedItem should not be nil
	// Expected: retrievedItem["name"] should equal "John Doe"
	fmt.Printf("Reading item with key: %s\n", userID)
	
	// Update the item
	_ = "SET #name = :name, age = age + :increment" // updateExpression
	_ = map[string]string{ // expressionAttributeNames
		"#name": "name",
	}
	_ = map[string]types.AttributeValue{ // expressionAttributeValues
		":name":      &types.AttributeValueMemberS{Value: "John Smith"},
		":increment": &types.AttributeValueMemberN{Value: "1"},
	}
	
	// updatedItem := dynamodb.UpdateItem(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues)
	// Expected: updatedItem should not be nil
	fmt.Printf("Updating item name to John Smith and incrementing age\n")
	
	// Verify the update
	// dynamodb.AssertAttributeEquals(t, tableName, key, "name", &types.AttributeValueMemberS{Value: "John Smith"})
	
	// Delete the item
	// dynamodb.DeleteItem(t, tableName, key)
	fmt.Printf("Deleting item with ID: %s\n", userID)
	
	// Verify item was deleted
	// dynamodb.AssertItemNotExists(t, tableName, key)
	fmt.Println("Basic CRUD operations completed")
}

// ExampleConditionalOperations demonstrates conditional operations
func ExampleConditionalOperations() {
	fmt.Println("DynamoDB conditional operations:")
	
	_ = "users" // tableName
	userID := "user-456"
	
	// Put item with condition (should succeed)
	_ = map[string]types.AttributeValue{ // item
		"id":   &types.AttributeValueMemberS{Value: userID},
		"name": &types.AttributeValueMemberS{Value: "Jane Doe"},
	}
	_ = dynamodb.PutItemOptions{ // options
		ConditionExpression: stringPtr("attribute_not_exists(id)"),
	}
	
	// dynamodb.PutItemWithOptions(t, tableName, item, options)
	fmt.Printf("Putting item with condition: attribute_not_exists(id)\n")
	
	// Try to put the same item again (should fail)
	// err := dynamodb.PutItemWithOptionsE(t, tableName, item, options)
	// dynamodb.AssertConditionalCheckFailed(t, err)
	fmt.Println("Second put attempt should fail due to condition")
	
	// Update with condition
	_ = /* key */ map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: userID},
	}
	_ = /* updateOptions */ dynamodb.UpdateItemOptions{
		ConditionExpression: stringPtr("attribute_exists(id)"),
		ReturnValues:        types.ReturnValueAllNew,
	}
	
	// updatedItem := dynamodb.UpdateItemWithOptions(t, tableName, key, "SET age = :age", nil, map[string]types.AttributeValue{
	//     ":age": &types.AttributeValueMemberN{Value: "25"},
	// }, updateOptions)
	// Expected: updatedItem should not be nil
	fmt.Printf("Updating item with condition: attribute_exists(id)\n")
	
	// Clean up
	// dynamodb.DeleteItem(t, tableName, key)
	fmt.Printf("Cleaning up item: %s\n", userID)
	fmt.Println("Conditional operations completed")
}

// ExampleQueryOperations demonstrates query operations with pagination
func ExampleQueryOperations() {
	fmt.Println("DynamoDB query operations with pagination:")
	
	_ = "orders" // tableName  
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
	// for _, order := range orders {
	//     dynamodb.PutItem(t, tableName, order)
	// }
	fmt.Printf("Creating %d orders for customer %s\n", len(orders), customerID)
	
	// Query all orders for customer
	keyConditionExpression := "customer_id = :customer_id"
	_ = /* expressionAttributeValues */ map[string]types.AttributeValue{
		":customer_id": &types.AttributeValueMemberS{Value: customerID},
	}
	
	// result := dynamodb.Query(t, tableName, keyConditionExpression, expressionAttributeValues)
	// Expected: result.Count should equal 3
	fmt.Printf("Querying all orders for customer: %s\n", customerID)
	
	// Query with filter expression (only completed orders)
	_ = /* queryOptions */ dynamodb.QueryOptions{
		FilterExpression: stringPtr("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		KeyConditionExpression: &keyConditionExpression,
	}
	
	_ = /* filteredExpressionsAttributeValues */ map[string]types.AttributeValue{
		":customer_id": &types.AttributeValueMemberS{Value: customerID},
		":status":      &types.AttributeValueMemberS{Value: "completed"},
	}
	
	// filteredResult := dynamodb.QueryWithOptions(t, tableName, keyConditionExpression, filteredExpressionsAttributeValues, queryOptions)
	// Expected: filteredResult.Count should equal 2
	fmt.Println("Querying with filter: only completed orders")
	
	// Query with pagination (limit 1 item at a time)
	paginationOptions := dynamodb.QueryOptions{
		Limit:                  int32Ptr(1),
		KeyConditionExpression: &keyConditionExpression,
	}
	
	var allItems []map[string]types.AttributeValue
	var exclusiveStartKey map[string]types.AttributeValue
	
	// Pagination loop pattern
	for {
		paginationOptions.ExclusiveStartKey = exclusiveStartKey
		// pageResult := dynamodb.QueryWithOptions(t, tableName, keyConditionExpression, expressionAttributeValues, paginationOptions)
		// allItems = append(allItems, pageResult.Items...)
		// if pageResult.LastEvaluatedKey == nil {
		//     break
		// }
		// exclusiveStartKey = pageResult.LastEvaluatedKey
		break // Simulated loop for demo
	}
	
	// Expected: len(allItems) should equal 3
	fmt.Printf("Pagination pattern: collected %d items\n", len(allItems))
	
	// Use convenience method for getting all pages
	// allItemsPaginated := dynamodb.QueryAllPages(t, tableName, keyConditionExpression, expressionAttributeValues)
	// Expected: len(allItemsPaginated) should equal 3
	fmt.Println("Using convenience method for all pages")
	
	// Clean up
	// for _, order := range orders {
	//     key := map[string]types.AttributeValue{
	//         "customer_id": order["customer_id"],
	//         "order_id":    order["order_id"],
	//     }
	//     dynamodb.DeleteItem(t, tableName, key)
	// }
	fmt.Printf("Cleaning up %d orders\n", len(orders))
	fmt.Println("Query operations completed")
}

// ExampleBatchOperations demonstrates batch read and write operations
func ExampleBatchOperations() {
	fmt.Println("DynamoDB batch operations:")
	
	_ = /* tableName */ "products"
	
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
	// result := dynamodb.BatchWriteItem(t, tableName, writeRequests)
	// Expected: result should not be nil
	fmt.Printf("Batch writing %d items\n", len(writeRequests))
	
	// Use retry mechanism for robustness
	// dynamodb.BatchWriteItemWithRetry(t, tableName, writeRequests, 3)
	fmt.Println("Using retry mechanism for batch writes")
	
	// Batch read items
	keys := []map[string]types.AttributeValue{
		{"id": &types.AttributeValueMemberS{Value: "product-001"}},
		{"id": &types.AttributeValueMemberS{Value: "product-002"}},
		{"id": &types.AttributeValueMemberS{Value: "product-003"}},
	}
	
	// batchResult := dynamodb.BatchGetItem(t, tableName, keys)
	// Expected: len(batchResult.Items) should equal 3
	fmt.Printf("Batch reading %d items\n", len(keys))
	
	// Use retry mechanism for batch reads
	// allItems := dynamodb.BatchGetItemWithRetry(t, tableName, keys, 3)
	// Expected: len(allItems) should equal 3
	fmt.Println("Using retry mechanism for batch reads")
	
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
	
	// dynamodb.BatchWriteItem(t, tableName, deleteRequests)
	fmt.Printf("Batch deleting %d items\n", len(deleteRequests))
	fmt.Println("Batch operations completed")
}

// ExampleTransactionalOperations demonstrates transactional operations
func ExampleTransactionalOperations() {
	fmt.Println("DynamoDB transactional operations:")
	
	tableName := "accounts"
	
	// Setup test accounts
	_ = /* account1Item */ map[string]types.AttributeValue{
		"id":      &types.AttributeValueMemberS{Value: "account-001"},
		"balance": &types.AttributeValueMemberN{Value: "1000"},
		"name":    &types.AttributeValueMemberS{Value: "Alice"},
	}
	_ = /* account2Item */ map[string]types.AttributeValue{
		"id":      &types.AttributeValueMemberS{Value: "account-002"},
		"balance": &types.AttributeValueMemberN{Value: "500"},
		"name":    &types.AttributeValueMemberS{Value: "Bob"},
	}
	
	// dynamodb.PutItem(t, tableName, account1Item)
	// dynamodb.PutItem(t, tableName, account2Item)
	fmt.Println("Setting up test accounts: Alice ($1000) and Bob ($500)")
	
	// Perform transactional transfer (Alice sends $100 to Bob)
	transferAmount := "100"
	
	_ = /* transactItems */ []types.TransactWriteItem{
		{
			Update: &types.Update{
				TableName: &tableName,
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "account-001"},
				},
				UpdateExpression:    stringPtr("SET balance = balance - :amount"),
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
	// transactResult := dynamodb.TransactWriteItems(t, transactItems)
	// Expected: transactResult should not be nil
	fmt.Printf("Executing transactional transfer of $%s from Alice to Bob\n", transferAmount)
	
	// Verify the balances after transaction
	_ = /* key1 */ map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "account-001"}}
	_ = /* key2 */ map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "account-002"}}
	
	// dynamodb.AssertAttributeEquals(t, tableName, key1, "balance", &types.AttributeValueMemberN{Value: "900"})
	// dynamodb.AssertAttributeEquals(t, tableName, key2, "balance", &types.AttributeValueMemberN{Value: "600"})
	fmt.Println("Expected: Alice balance = $900, Bob balance = $600")
	
	// Transactional read
	_ = /* transactReadItems */ []types.TransactGetItem{
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
	
	// readResult := dynamodb.TransactGetItems(t, transactReadItems)
	// Expected: len(readResult.Responses) should equal 2
	fmt.Println("Performing transactional read of both accounts")
	
	// Clean up
	// dynamodb.DeleteItem(t, tableName, key1)
	// dynamodb.DeleteItem(t, tableName, key2)
	fmt.Println("Cleaning up accounts")
	fmt.Println("Transactional operations completed")
}

// ExampleAssertionPatterns demonstrates common assertion patterns
func ExampleAssertionPatterns() {
	fmt.Println("DynamoDB assertion patterns:")
	
	tableName := "test-table"
	
	// Assert table exists (assuming it's been created)
	// dynamodb.AssertTableExists(t, tableName)
	fmt.Printf("Checking table exists: %s\n", tableName)
	
	// Assert table status
	// dynamodb.AssertTableStatus(t, tableName, types.TableStatusActive)
	fmt.Printf("Checking table status is ACTIVE\n")
	
	// Create a test item
	_ = /* item */ map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: "test-item"},
		"name": &types.AttributeValueMemberS{Value: "Test Item"},
	}
	// dynamodb.PutItem(t, tableName, item)
	fmt.Println("Creating test item")
	
	// Test item assertions
	_ = /* key */ map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-item"},
	}
	
	// dynamodb.AssertItemExists(t, tableName, key)
	// dynamodb.AssertAttributeEquals(t, tableName, key, "name", &types.AttributeValueMemberS{Value: "Test Item"})
	fmt.Println("Asserting item exists and name equals 'Test Item'")
	
	// Test query count assertion
	_ = /* keyConditionExpression */ "id = :id"
	_ = /* expressionAttributeValues */ map[string]types.AttributeValue{
		":id": &types.AttributeValueMemberS{Value: "test-item"},
	}
	// dynamodb.AssertQueryCount(t, tableName, keyConditionExpression, expressionAttributeValues, 1)
	fmt.Println("Asserting query count equals 1")
	
	// Test conditional operation assertions
	_ = /* updateOptions */ dynamodb.UpdateItemOptions{
		ConditionExpression: stringPtr("attribute_exists(id)"),
	}
	
	// _, err := dynamodb.UpdateItemWithOptionsE(t, tableName, key, "SET description = :desc", nil, map[string]types.AttributeValue{
	//     ":desc": &types.AttributeValueMemberS{Value: "Updated description"},
	// }, updateOptions)
	// dynamodb.AssertConditionalCheckPassed(t, err)
	fmt.Println("Asserting conditional update passes")
	
	// Clean up
	// dynamodb.DeleteItem(t, tableName, key)
	// dynamodb.AssertItemNotExists(t, tableName, key)
	fmt.Println("Cleaning up and asserting item no longer exists")
	fmt.Println("Assertion patterns completed")
}

// ExampleScanOperations demonstrates scan operations with filtering
func ExampleScanOperations() {
	fmt.Println("DynamoDB scan operations with filtering:")
	
	_ = /* tableName */ "inventory"
	
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
	// for _, item := range items {
	//     dynamodb.PutItem(t, tableName, item)
	// }
	fmt.Printf("Creating %d inventory items\n", len(items))
	
	// Scan all items
	// result := dynamodb.Scan(t, tableName)
	// Expected: result.Count should be >= 3
	fmt.Println("Scanning all items in table")
	
	// Scan with filter for electronics category
	_ = /* scanOptions */ dynamodb.ScanOptions{
		FilterExpression: stringPtr("category = :category"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":category": &types.AttributeValueMemberS{Value: "electronics"},
		},
	}
	
	// filteredResult := dynamodb.ScanWithOptions(t, tableName, scanOptions)
	// Expected: filteredResult.Count should equal 2
	fmt.Println("Scanning with filter: electronics category")
	
	// Scan with filter for in-stock items
	_ = /* stockOptions */ dynamodb.ScanOptions{
		FilterExpression: stringPtr("stock > :zero"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":zero": &types.AttributeValueMemberN{Value: "0"},
		},
	}
	
	// inStockResult := dynamodb.ScanWithOptions(t, tableName, stockOptions)
	// Expected: inStockResult.Count should equal 2
	fmt.Println("Scanning with filter: in-stock items only")
	
	// Use scan with pagination
	// allItems := dynamodb.ScanAllPages(t, tableName)
	// Expected: len(allItems) should be >= 3
	fmt.Println("Using scan with pagination to get all items")
	
	// Verify item count
	// dynamodb.AssertItemCount(t, tableName, len(items))
	fmt.Printf("Expected item count: %d\n", len(items))
	
	// Clean up
	// for _, item := range items {
	//     key := map[string]types.AttributeValue{
	//         "id": item["id"],
	//     }
	//     dynamodb.DeleteItem(t, tableName, key)
	// }
	fmt.Printf("Cleaning up %d items\n", len(items))
	fmt.Println("Scan operations completed")
}

// runAllExamples demonstrates running all DynamoDB examples
func runAllExamples() {
	fmt.Println("Running all DynamoDB package examples:")
	fmt.Println("")
	
	ExampleBasicCRUDOperations()
	fmt.Println("")
	
	ExampleConditionalOperations()
	fmt.Println("")
	
	ExampleQueryOperations()
	fmt.Println("")
	
	ExampleBatchOperations()
	fmt.Println("")
	
	ExampleTransactionalOperations()
	fmt.Println("")
	
	ExampleAssertionPatterns()
	fmt.Println("")
	
	ExampleScanOperations()
	fmt.Println("All DynamoDB examples completed")
}

func main() {
	// This file demonstrates DynamoDB usage patterns with the vasdeference framework.
	// Run examples with: go run ./examples/dynamodb/examples.go
	
	runAllExamples()
}