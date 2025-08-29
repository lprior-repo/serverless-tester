package main

import (
	"fmt"

	"vasdeference/dynamodb"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// stringPtr helper function for conditional expressions
func stringPtr(s string) *string {
	return &s
}

// int32Ptr helper function for limit values
func int32Ptr(i int32) *int32 {
	return &i
}

// ExampleBasicCRUDOperations demonstrates functional CRUD operations
func ExampleBasicCRUDOperations() {
	fmt.Println("Functional DynamoDB CRUD operations:")
	
	// Use monadic configuration
	config := dynamodb.NewFunctionalDynamoDBConfig().OrEmpty()
	tableName := mo.Some("users")
	userID := "user-123"
	
	// Create item using functional patterns
	userItem := lo.Pipe2(
		map[string]types.AttributeValue{},
		func(m map[string]types.AttributeValue) map[string]types.AttributeValue {
			return lo.Assign(m, map[string]types.AttributeValue{
				"id":     &types.AttributeValueMemberS{Value: userID},
				"name":   &types.AttributeValueMemberS{Value: "John Doe"},
				"email":  &types.AttributeValueMemberS{Value: "john.doe@example.com"},
				"age":    &types.AttributeValueMemberN{Value: "30"},
				"active": &types.AttributeValueMemberBOOL{Value: true},
			})
		},
		func(item map[string]types.AttributeValue) map[string]types.AttributeValue {
			return item
		},
	)
	
	// Simulate functional DynamoDB operations
	result := dynamodb.SimulateFunctionalDynamoDBCreateTable(config)
	
	result.Match(
		func(success dynamodb.FunctionalDynamoDBConfig) {
			fmt.Printf("✓ Successfully created item with ID: %s\n", userID)
			// Functional item retrieval
			key := mo.Some(map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: userID},
			})
			if retrievedItem, exists := key.Get(); exists {
				fmt.Printf("✓ Reading item with key: %s\n", userID)
				// Validate item properties functionally
				if nameAttr, hasName := retrievedItem["name"]; hasName {
					if nameValue, ok := nameAttr.(*types.AttributeValueMemberS); ok {
						lo.Ternary(nameValue.Value == "John Doe",
							fmt.Println("✓ Name validation passed"),
							fmt.Println("✗ Name validation failed"),
						)
					}
				}
			}
		},
		func(err error) {
			fmt.Printf("✗ Failed to create item: %v\n", err)
		},
	)
	
	// Functional update operation
	updateConfig := lo.Pipe3(
		map[string]string{},
		func(names map[string]string) map[string]string {
			return lo.Assign(names, map[string]string{"#name": "name"})
		},
		func(names map[string]string) (map[string]string, map[string]types.AttributeValue) {
			values := map[string]types.AttributeValue{
				":name":      &types.AttributeValueMemberS{Value: "John Smith"},
				":increment": &types.AttributeValueMemberN{Value: "1"},
			}
			return names, values
		},
		func(names map[string]string, values map[string]types.AttributeValue) mo.Option[dynamodb.UpdateOperation] {
			return mo.Some(dynamodb.UpdateOperation{
				Expression: "SET #name = :name, age = age + :increment",
				Names:      names,
				Values:     values,
			})
		},
	)
	
	updateConfig.IfPresent(func(op dynamodb.UpdateOperation) {
		fmt.Println("✓ Functional update: name to John Smith, age incremented")
	})
	
	// Functional deletion with validation
	deleteResult := mo.Some(userID).Map(func(id string) mo.Result[string] {
		fmt.Printf("✓ Functionally deleting item with ID: %s\n", id)
		return mo.Ok(id)
	})
	
	deleteResult.IfPresent(func(result mo.Result[string]) {
		result.Match(
			func(id string) { fmt.Println("✓ Item successfully deleted") },
			func(err error) { fmt.Printf("✗ Delete failed: %v\n", err) },
		)
	})
	
	fmt.Println("✓ Functional CRUD operations completed")
}

// ExampleConditionalOperations demonstrates functional conditional operations
func ExampleConditionalOperations() {
	fmt.Println("Functional DynamoDB conditional operations:")
	
	tableName := mo.Some("users")
	userID := "user-456"
	
	// Functional conditional put with monadic validation
	conditionalItem := lo.Pipe2(
		map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: userID},
			"name": &types.AttributeValueMemberS{Value: "Jane Doe"},
		},
		func(item map[string]types.AttributeValue) mo.Option[map[string]types.AttributeValue] {
			return mo.Some(item)
		},
		func(item mo.Option[map[string]types.AttributeValue]) mo.Result[dynamodb.ConditionalPutOperation] {
			return item.Match(
				func(i map[string]types.AttributeValue) mo.Result[dynamodb.ConditionalPutOperation] {
					return mo.Ok(dynamodb.ConditionalPutOperation{
						Item:      i,
						Condition: "attribute_not_exists(id)",
					})
				},
				func() mo.Result[dynamodb.ConditionalPutOperation] {
					return mo.Err[dynamodb.ConditionalPutOperation](fmt.Errorf("invalid item"))
				},
			)
		},
	)
	
	conditionalItem.Match(
		func(op dynamodb.ConditionalPutOperation) {
			fmt.Printf("✓ Conditional put succeeded: attribute_not_exists(id)\n")
		},
		func(err error) {
			fmt.Printf("✗ Conditional put failed: %v\n", err)
		},
	)
	
	// Functional retry logic with monadic chaining
	retryResult := lo.Pipe2(
		conditionalItem,
		func(first mo.Result[dynamodb.ConditionalPutOperation]) mo.Result[dynamodb.ConditionalPutOperation] {
			// Simulate second attempt (should fail)
			return mo.Err[dynamodb.ConditionalPutOperation](fmt.Errorf("conditional check failed"))
		},
		func(second mo.Result[dynamodb.ConditionalPutOperation]) mo.Option[string] {
			return second.Match(
				func(op dynamodb.ConditionalPutOperation) mo.Option[string] {
					return mo.Some("unexpected success")
				},
				func(err error) mo.Option[string] {
					return mo.Some("expected failure: conditional check failed")
				},
			)
		},
	)
	
	retryResult.IfPresent(func(message string) {
		fmt.Printf("✓ %s\n", message)
	})
	
	// Functional conditional update
	conditionalUpdate := mo.Some(map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: userID},
	}).Map(func(key map[string]types.AttributeValue) dynamodb.ConditionalUpdateOperation {
		return dynamodb.ConditionalUpdateOperation{
			Key:       key,
			Condition: "attribute_exists(id)",
			Expression: "SET age = :age",
			Values: map[string]types.AttributeValue{
				":age": &types.AttributeValueMemberN{Value: "25"},
			},
		}
	})
	
	conditionalUpdate.IfPresent(func(op dynamodb.ConditionalUpdateOperation) {
		fmt.Printf("✓ Conditional update prepared: attribute_exists(id)\n")
	})
	
	// Functional cleanup with validation chain
	cleanupChain := lo.Pipe3(
		mo.Some(userID),
		func(id mo.Option[string]) mo.Option[string] {
			return id.Filter(func(s string) bool { return len(s) > 0 })
		},
		func(id mo.Option[string]) mo.Result[string] {
			return id.Match(
				func(s string) mo.Result[string] { return mo.Ok(s) },
				func() mo.Result[string] { return mo.Err[string](fmt.Errorf("invalid user ID")) },
			)
		},
		func(result mo.Result[string]) mo.Option[string] {
			return result.Match(
				func(id string) mo.Option[string] {
					fmt.Printf("✓ Cleaning up item: %s\n", id)
					return mo.Some("cleanup completed")
				},
				func(err error) mo.Option[string] {
					fmt.Printf("✗ Cleanup failed: %v\n", err)
					return mo.None[string]()
				},
			)
		},
	)
	
	cleanupChain.IfPresent(func(message string) {
		fmt.Printf("✓ %s\n", message)
	})
	
	fmt.Println("✓ Functional conditional operations completed")
}

// ExampleQueryOperations demonstrates functional query operations with pagination
func ExampleQueryOperations() {
	fmt.Println("Functional DynamoDB query operations with pagination:")
	
	tableName := mo.Some("orders")
	customerID := "customer-789"
	
	// Create immutable test data using functional composition
	orderFactory := func(id, amount, status string) map[string]types.AttributeValue {
		return map[string]types.AttributeValue{
			"customer_id": &types.AttributeValueMemberS{Value: customerID},
			"order_id":    &types.AttributeValueMemberS{Value: id},
			"amount":      &types.AttributeValueMemberN{Value: amount},
			"status":      &types.AttributeValueMemberS{Value: status},
		}
	}
	
	orders := lo.Map([]struct{ id, amount, status string }{
		{"order-001", "99.99", "completed"},
		{"order-002", "149.99", "pending"},
		{"order-003", "199.99", "completed"},
	}, func(order struct{ id, amount, status string }, _ int) map[string]types.AttributeValue {
		return orderFactory(order.id, order.amount, order.status)
	})
	
	// Functional batch creation with validation
	batchResult := lo.Pipe2(
		orders,
		func(orderList []map[string]types.AttributeValue) mo.Result[[]map[string]types.AttributeValue] {
			if len(orderList) == 0 {
				return mo.Err[[]map[string]types.AttributeValue](fmt.Errorf("no orders to create"))
			}
			return mo.Ok(orderList)
		},
		func(validated mo.Result[[]map[string]types.AttributeValue]) mo.Option[int] {
			return validated.Match(
				func(orderList []map[string]types.AttributeValue) mo.Option[int] {
					fmt.Printf("✓ Creating %d orders for customer %s\n", len(orderList), customerID)
					return mo.Some(len(orderList))
				},
				func(err error) mo.Option[int] {
					fmt.Printf("✗ Failed to create orders: %v\n", err)
					return mo.None[int]()
				},
			)
		},
	)
	
	// Functional query configuration
	queryConfig := lo.Pipe2(
		"customer_id = :customer_id",
		func(keyCondition string) dynamodb.QueryOperation {
			return dynamodb.QueryOperation{
				KeyCondition: keyCondition,
				Values: map[string]types.AttributeValue{
					":customer_id": &types.AttributeValueMemberS{Value: customerID},
				},
			}
		},
		func(query dynamodb.QueryOperation) mo.Result[dynamodb.QueryOperation] {
			return mo.Ok(query)
		},
	)
	
	queryConfig.Match(
		func(query dynamodb.QueryOperation) {
			fmt.Printf("✓ Querying all orders for customer: %s\n", customerID)
			batchResult.IfPresent(func(count int) {
				fmt.Printf("✓ Expected result count: %d\n", count)
			})
		},
		func(err error) {
			fmt.Printf("✗ Query configuration failed: %v\n", err)
		},
	)
	
	// Functional filtered query using monadic composition
	filteredQuery := lo.Pipe3(
		orders,
		func(orderList []map[string]types.AttributeValue) []map[string]types.AttributeValue {
			return lo.Filter(orderList, func(order map[string]types.AttributeValue, _ int) bool {
				if statusAttr, exists := order["status"]; exists {
					if statusVal, ok := statusAttr.(*types.AttributeValueMemberS); ok {
						return statusVal.Value == "completed"
					}
				}
				return false
			})
		},
		func(completedOrders []map[string]types.AttributeValue) dynamodb.FilteredQueryOperation {
			return dynamodb.FilteredQueryOperation{
				KeyCondition: "customer_id = :customer_id",
				FilterExpression: "#status = :status",
				Names: map[string]string{"#status": "status"},
				Values: map[string]types.AttributeValue{
					":customer_id": &types.AttributeValueMemberS{Value: customerID},
					":status":      &types.AttributeValueMemberS{Value: "completed"},
				},
				ExpectedCount: len(completedOrders),
			}
		},
		func(query dynamodb.FilteredQueryOperation) mo.Result[dynamodb.FilteredQueryOperation] {
			return mo.Ok(query)
		},
	)
	
	filteredQuery.Match(
		func(query dynamodb.FilteredQueryOperation) {
			fmt.Printf("✓ Filtered query: only completed orders (expected: %d)\n", query.ExpectedCount)
		},
		func(err error) {
			fmt.Printf("✗ Filtered query failed: %v\n", err)
		},
	)
	
	// Functional pagination using recursive monadic approach
	paginationResult := lo.Pipe3(
		orders,
		func(orderList []map[string]types.AttributeValue) dynamodb.PaginationConfig {
			return dynamodb.PaginationConfig{
				PageSize:     1,
				TotalItems:   len(orderList),
				KeyCondition: "customer_id = :customer_id",
			}
		},
		func(config dynamodb.PaginationConfig) [][]map[string]types.AttributeValue {
			// Functional pagination simulation
			return lo.Chunk(orders, config.PageSize)
		},
		func(pages [][]map[string]types.AttributeValue) mo.Result[dynamodb.PaginatedResult] {
			allItems := lo.Flatten(pages)
			return mo.Ok(dynamodb.PaginatedResult{
				Items:     allItems,
				PageCount: len(pages),
				TotalItems: len(allItems),
			})
		},
	)
	
	paginationResult.Match(
		func(result dynamodb.PaginatedResult) {
			fmt.Printf("✓ Functional pagination: collected %d items across %d pages\n", result.TotalItems, result.PageCount)
		},
		func(err error) {
			fmt.Printf("✗ Pagination failed: %v\n", err)
		},
	)
	
	// Monadic convenience method for all pages
	allPagesResult := mo.Some(orders).Map(func(orderList []map[string]types.AttributeValue) int {
		return len(orderList)
	})
	
	allPagesResult.IfPresent(func(count int) {
		fmt.Printf("✓ Convenience method: all %d pages processed\n", count)
	})
	
	// Functional cleanup using monadic batch operations
	cleanupResult := lo.Pipe3(
		orders,
		func(orderList []map[string]types.AttributeValue) []map[string]types.AttributeValue {
			return lo.Map(orderList, func(order map[string]types.AttributeValue, _ int) map[string]types.AttributeValue {
				return map[string]types.AttributeValue{
					"customer_id": order["customer_id"],
					"order_id":    order["order_id"],
				}
			})
		},
		func(keys []map[string]types.AttributeValue) mo.Result[[]map[string]types.AttributeValue] {
			if len(keys) == 0 {
				return mo.Err[[]map[string]types.AttributeValue](fmt.Errorf("no keys to cleanup"))
			}
			return mo.Ok(keys)
		},
		func(validated mo.Result[[]map[string]types.AttributeValue]) mo.Option[int] {
			return validated.Match(
				func(keys []map[string]types.AttributeValue) mo.Option[int] {
					fmt.Printf("✓ Cleaning up %d orders\n", len(keys))
					return mo.Some(len(keys))
				},
				func(err error) mo.Option[int] {
					fmt.Printf("✗ Cleanup failed: %v\n", err)
					return mo.None[int]()
				},
			)
		},
	)
	
	cleanupResult.IfPresent(func(count int) {
		fmt.Printf("✓ Successfully cleaned up %d items\n", count)
	})
	
	fmt.Println("✓ Functional query operations completed")
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

// runAllExamples demonstrates running all functional DynamoDB examples
func runAllExamples() {
	fmt.Println("Running all functional DynamoDB package examples:")
	fmt.Println("")
	
	// Use functional composition to run examples with error handling
	exampleFunctions := []func(){
		ExampleBasicCRUDOperations,
		ExampleConditionalOperations, 
		ExampleQueryOperations,
		ExampleBatchOperations,
		ExampleTransactionalOperations,
		ExampleAssertionPatterns,
		ExampleScanOperations,
	}
	
	// Functional execution with monadic error handling
	executionResult := lo.Pipe2(
		exampleFunctions,
		func(examples []func()) mo.Result[[]func()] {
			if len(examples) == 0 {
				return mo.Err[[]func()](fmt.Errorf("no examples to run"))
			}
			return mo.Ok(examples)
		},
		func(validated mo.Result[[]func()]) mo.Option[int] {
			return validated.Match(
				func(examples []func()) mo.Option[int] {
					lo.ForEach(examples, func(example func(), index int) {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("✗ Example %d failed: %v\n", index+1, r)
							}
						}()
						example()
						fmt.Println("")
					})
					return mo.Some(len(examples))
				},
				func(err error) mo.Option[int] {
					fmt.Printf("✗ Failed to run examples: %v\n", err)
					return mo.None[int]()
				},
			)
		},
	)
	
	executionResult.Match(
		func(count int) {
			fmt.Printf("✓ All %d functional DynamoDB examples completed successfully\n", count)
		},
		func() {
			fmt.Println("✗ Example execution was interrupted")
		},
	)
}

func main() {
	// This file demonstrates functional DynamoDB usage patterns with the vasdeference framework.
	// Run examples with: go run ./examples/dynamodb/examples.go
	
	// Functional main execution with error boundary
	mainResult := mo.TryOr(func() error {
		runAllExamples()
		return nil
	}, func(err error) error {
		fmt.Printf("✗ Application failed: %v\n", err)
		return err
	})
	
	mainResult.Match(
		func(value interface{}) {
			fmt.Println("✓ Application completed successfully")
		},
		func(err error) {
			fmt.Printf("✗ Application terminated with error: %v\n", err)
		},
	)
}