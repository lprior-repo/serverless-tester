# DynamoDB Testing Package

A comprehensive AWS DynamoDB testing package for the `vasdeference` module that follows strict Terratest patterns. This package provides complete DynamoDB operations with automatic retry logic, comprehensive test assertions, and utilities for building robust integration tests.

## Features

### Core DynamoDB Operations
- **CRUD Operations**: PutItem, GetItem, UpdateItem, DeleteItem with full options support
- **Query & Scan**: With pagination, filtering, and projection expressions
- **Batch Operations**: BatchWriteItem, BatchGetItem with automatic retry and exponential backoff
- **Transactions**: TransactWriteItems, TransactGetItems for ACID compliance
- **Table Management**: CreateTable, DeleteTable, DescribeTable, UpdateTable with full lifecycle support

### Test Infrastructure
- **Terratest Patterns**: All functions follow `Function/FunctionE` pattern for proper error handling
- **Comprehensive Assertions**: Pre-built assertions for common testing scenarios
- **Automatic Retries**: Built-in exponential backoff for batch operations and table state changes
- **Test Isolation**: Utilities for backing up and restoring table data between tests

### Advanced Features
- **GSI/LSI Management**: Create, update, and manage Global and Local Secondary Indexes
- **Stream Processing**: Helpers for working with DynamoDB Streams
- **Point-in-time Recovery**: Backup and restore capabilities for test isolation
- **Conditional Operations**: Full support for condition expressions and atomic operations

## Installation

```bash
go get vasdeference/dynamodb
```

## Quick Start

```go
package main

import (
    "testing"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "vasdeference/dynamodb"
)

func TestBasicOperations(t *testing.T) {
    tableName := "users"
    
    // Create an item
    item := map[string]types.AttributeValue{
        "id":   &types.AttributeValueMemberS{Value: "user-123"},
        "name": &types.AttributeValueMemberS{Value: "John Doe"},
        "age":  &types.AttributeValueMemberN{Value: "30"},
    }
    
    dynamodb.PutItem(t, tableName, item)
    
    // Verify item exists
    key := map[string]types.AttributeValue{
        "id": &types.AttributeValueMemberS{Value: "user-123"},
    }
    dynamodb.AssertItemExists(t, tableName, key)
    
    // Read the item
    retrievedItem := dynamodb.GetItem(t, tableName, key)
    
    // Update the item
    updateExpression := "SET #name = :name, age = age + :increment"
    expressionAttributeNames := map[string]string{"#name": "name"}
    expressionAttributeValues := map[string]types.AttributeValue{
        ":name":      &types.AttributeValueMemberS{Value: "John Smith"},
        ":increment": &types.AttributeValueMemberN{Value: "1"},
    }
    
    dynamodb.UpdateItem(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues)
    
    // Verify the update
    dynamodb.AssertAttributeEquals(t, tableName, key, "name", &types.AttributeValueMemberS{Value: "John Smith"})
    
    // Delete the item
    dynamodb.DeleteItem(t, tableName, key)
    
    // Verify deletion
    dynamodb.AssertItemNotExists(t, tableName, key)
}
```

## Core Functions

### Item Operations

#### PutItem
```go
// Basic put
dynamodb.PutItem(t, tableName, item)

// With condition expression
options := dynamodb.PutItemOptions{
    ConditionExpression: stringPtr("attribute_not_exists(id)"),
}
dynamodb.PutItemWithOptions(t, tableName, item, options)

// Error handling version
err := dynamodb.PutItemE(t, tableName, item)
if err != nil {
    // Handle error
}
```

#### GetItem
```go
// Basic get
item := dynamodb.GetItem(t, tableName, key)

// With projection expression
options := dynamodb.GetItemOptions{
    ProjectionExpression: stringPtr("id, #name, age"),
    ExpressionAttributeNames: map[string]string{"#name": "name"},
}
item := dynamodb.GetItemWithOptions(t, tableName, key, options)
```

#### UpdateItem
```go
updateExpression := "SET #name = :name, age = age + :increment"
expressionAttributeNames := map[string]string{"#name": "name"}
expressionAttributeValues := map[string]types.AttributeValue{
    ":name":      &types.AttributeValueMemberS{Value: "Updated Name"},
    ":increment": &types.AttributeValueMemberN{Value: "1"},
}

result := dynamodb.UpdateItem(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues)
```

#### DeleteItem
```go
// Basic delete
dynamodb.DeleteItem(t, tableName, key)

// With condition
options := dynamodb.DeleteItemOptions{
    ConditionExpression: stringPtr("attribute_exists(id)"),
}
dynamodb.DeleteItemWithOptions(t, tableName, key, options)
```

### Query & Scan Operations

#### Query
```go
keyConditionExpression := "customer_id = :customer_id"
expressionAttributeValues := map[string]types.AttributeValue{
    ":customer_id": &types.AttributeValueMemberS{Value: "customer-123"},
}

result := dynamodb.Query(t, tableName, keyConditionExpression, expressionAttributeValues)

// With filtering
options := dynamodb.QueryOptions{
    FilterExpression: stringPtr("#status = :status"),
    ExpressionAttributeNames: map[string]string{"#status": "status"},
    Limit: int32Ptr(10),
}
result := dynamodb.QueryWithOptions(t, tableName, keyConditionExpression, expressionAttributeValues, options)

// Get all pages automatically
allItems := dynamodb.QueryAllPages(t, tableName, keyConditionExpression, expressionAttributeValues)
```

#### Scan
```go
// Basic scan
result := dynamodb.Scan(t, tableName)

// With filter
options := dynamodb.ScanOptions{
    FilterExpression: stringPtr("age > :minAge"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":minAge": &types.AttributeValueMemberN{Value: "18"},
    },
}
result := dynamodb.ScanWithOptions(t, tableName, options)

// Get all pages
allItems := dynamodb.ScanAllPages(t, tableName)
```

### Batch Operations

#### Batch Write
```go
writeRequests := []types.WriteRequest{
    {
        PutRequest: &types.PutRequest{
            Item: map[string]types.AttributeValue{
                "id": &types.AttributeValueMemberS{Value: "item-1"},
                "name": &types.AttributeValueMemberS{Value: "Item 1"},
            },
        },
    },
    {
        DeleteRequest: &types.DeleteRequest{
            Key: map[string]types.AttributeValue{
                "id": &types.AttributeValueMemberS{Value: "item-2"},
            },
        },
    },
}

// Basic batch write
result := dynamodb.BatchWriteItem(t, tableName, writeRequests)

// With automatic retry for unprocessed items
dynamodb.BatchWriteItemWithRetry(t, tableName, writeRequests, 3)
```

#### Batch Get
```go
keys := []map[string]types.AttributeValue{
    {"id": &types.AttributeValueMemberS{Value: "item-1"}},
    {"id": &types.AttributeValueMemberS{Value: "item-2"}},
    {"id": &types.AttributeValueMemberS{Value: "item-3"}},
}

// Basic batch get
result := dynamodb.BatchGetItem(t, tableName, keys)

// With retry for unprocessed keys
items := dynamodb.BatchGetItemWithRetry(t, tableName, keys, 3)
```

### Transaction Operations

#### Transaction Write
```go
transactItems := []types.TransactWriteItem{
    {
        Put: &types.Put{
            TableName: stringPtr(tableName),
            Item: map[string]types.AttributeValue{
                "id": &types.AttributeValueMemberS{Value: "item-1"},
                "balance": &types.AttributeValueMemberN{Value: "1000"},
            },
            ConditionExpression: stringPtr("attribute_not_exists(id)"),
        },
    },
    {
        Update: &types.Update{
            TableName: stringPtr(tableName),
            Key: map[string]types.AttributeValue{
                "id": &types.AttributeValueMemberS{Value: "item-2"},
            },
            UpdateExpression: stringPtr("SET balance = balance - :amount"),
            ConditionExpression: stringPtr("balance >= :amount"),
            ExpressionAttributeValues: map[string]types.AttributeValue{
                ":amount": &types.AttributeValueMemberN{Value: "100"},
            },
        },
    },
}

result := dynamodb.TransactWriteItems(t, transactItems)
```

#### Transaction Read
```go
transactItems := []types.TransactGetItem{
    {
        Get: &types.Get{
            TableName: stringPtr(tableName),
            Key: map[string]types.AttributeValue{
                "id": &types.AttributeValueMemberS{Value: "item-1"},
            },
        },
    },
    {
        Get: &types.Get{
            TableName: stringPtr(tableName),
            Key: map[string]types.AttributeValue{
                "id": &types.AttributeValueMemberS{Value: "item-2"},
            },
        },
    },
}

result := dynamodb.TransactGetItems(t, transactItems)
```

### Table Lifecycle Management

#### Create Table
```go
keySchema := []types.KeySchemaElement{
    {
        AttributeName: stringPtr("id"),
        KeyType:       types.KeyTypeHash,
    },
}

attributeDefinitions := []types.AttributeDefinition{
    {
        AttributeName: stringPtr("id"),
        AttributeType: types.ScalarAttributeTypeS,
    },
}

// Basic table creation
table := dynamodb.CreateTable(t, tableName, keySchema, attributeDefinitions)

// With options
options := dynamodb.CreateTableOptions{
    BillingMode: types.BillingModePayPerRequest,
    StreamSpecification: &types.StreamSpecification{
        StreamEnabled:  boolPtr(true),
        StreamViewType: types.StreamViewTypeNewAndOldImages,
    },
}
table := dynamodb.CreateTableWithOptions(t, tableName, keySchema, attributeDefinitions, options)

// Wait for table to become active
dynamodb.WaitForTableActive(t, tableName, 5*time.Minute)
```

#### Delete Table
```go
dynamodb.DeleteTable(t, tableName)
dynamodb.WaitForTableDeleted(t, tableName, 5*time.Minute)
```

## Test Assertions

The package provides comprehensive assertions for validating DynamoDB operations:

### Table Assertions
```go
// Table existence
dynamodb.AssertTableExists(t, tableName)
dynamodb.AssertTableNotExists(t, tableName)

// Table status
dynamodb.AssertTableStatus(t, tableName, types.TableStatusActive)

// Stream configuration
dynamodb.AssertStreamEnabled(t, tableName)

// GSI assertions
dynamodb.AssertGSIExists(t, tableName, indexName)
dynamodb.AssertGSIStatus(t, tableName, indexName, types.IndexStatusActive)
```

### Item Assertions
```go
// Item existence
dynamodb.AssertItemExists(t, tableName, key)
dynamodb.AssertItemNotExists(t, tableName, key)

// Item count
dynamodb.AssertItemCount(t, tableName, expectedCount)

// Attribute values
dynamodb.AssertAttributeEquals(t, tableName, key, "name", &types.AttributeValueMemberS{Value: "John Doe"})

// Query results
dynamodb.AssertQueryCount(t, tableName, keyConditionExpression, expressionAttributeValues, 5)
```

### Conditional Operation Assertions
```go
// Assert conditional check passed
err := dynamodb.UpdateItemE(t, tableName, key, updateExpression, nil, values)
dynamodb.AssertConditionalCheckPassed(t, err)

// Assert conditional check failed as expected
err = dynamodb.PutItemWithOptionsE(t, tableName, item, options)
dynamodb.AssertConditionalCheckFailed(t, err)
```

### Backup Assertions
```go
dynamodb.AssertBackupExists(t, backupArn)
```

## Error Handling Patterns

The package follows Terratest's function/functionE pattern:

- **Function**: Fails the test immediately if an error occurs (uses `require.NoError`)
- **FunctionE**: Returns the error for custom handling

```go
// Fails test on error
item := dynamodb.GetItem(t, tableName, key)

// Returns error for custom handling
item, err := dynamodb.GetItemE(t, tableName, key)
if err != nil {
    // Custom error handling
    t.Logf("Failed to get item: %v", err)
    return
}
```

## Advanced Usage

### Conditional Operations
```go
// Put with condition
options := dynamodb.PutItemOptions{
    ConditionExpression: stringPtr("attribute_not_exists(id)"),
}
err := dynamodb.PutItemWithOptionsE(t, tableName, item, options)
dynamodb.AssertConditionalCheckFailed(t, err)

// Update with condition and return values
updateOptions := dynamodb.UpdateItemOptions{
    ConditionExpression: stringPtr("balance >= :amount"),
    ReturnValues: types.ReturnValueAllNew,
}
result := dynamodb.UpdateItemWithOptions(t, tableName, key, updateExpression, names, values, updateOptions)
```

### Complex Queries
```go
// Query with multiple conditions and pagination
options := dynamodb.QueryOptions{
    KeyConditionExpression: stringPtr("customer_id = :customer_id AND order_date BETWEEN :start_date AND :end_date"),
    FilterExpression: stringPtr("#status = :status AND amount > :min_amount"),
    ExpressionAttributeNames: map[string]string{
        "#status": "status",
    },
    ProjectionExpression: stringPtr("order_id, amount, #status"),
    Limit: int32Ptr(100),
    ScanIndexForward: boolPtr(false), // Descending order
}

result := dynamodb.QueryWithOptions(t, tableName, "", expressionAttributeValues, options)
```

### Stream Processing
```go
// Enable streams on table creation
streamSpec := &types.StreamSpecification{
    StreamEnabled:  boolPtr(true),
    StreamViewType: types.StreamViewTypeNewAndOldImages,
}
options := dynamodb.CreateTableOptions{
    StreamSpecification: streamSpec,
}
table := dynamodb.CreateTableWithOptions(t, tableName, keySchema, attributeDefinitions, options)

// Verify stream is enabled
dynamodb.AssertStreamEnabled(t, tableName)
```

## Testing Best Practices

### Test Isolation
```go
func TestWithCleanup(t *testing.T) {
    tableName := "test-table"
    
    // Setup
    table := setupTestTable(t, tableName)
    defer func() {
        // Cleanup
        dynamodb.DeleteTable(t, tableName)
        dynamodb.WaitForTableDeleted(t, tableName, 2*time.Minute)
    }()
    
    // Test operations
    // ...
}
```

### Parallel Testing
```go
func TestParallelOperations(t *testing.T) {
    t.Parallel()
    
    tableName := fmt.Sprintf("test-table-%d", time.Now().UnixNano())
    
    // Each test gets its own table
    // ...
}
```

### Retry and Backoff
```go
// Use built-in retry mechanisms for robust testing
dynamodb.BatchWriteItemWithRetry(t, tableName, writeRequests, 5)
dynamodb.BatchGetItemWithRetry(t, tableName, keys, 5)

// Wait for table state changes
dynamodb.WaitForTableActive(t, tableName, 10*time.Minute)
```

## Environment Setup

### AWS Credentials
Set up AWS credentials through any of the standard methods:
- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- AWS credentials file
- IAM roles (for EC2/ECS)
- AWS SSO

### LocalStack Support
The package works with LocalStack for local testing:
```bash
# Start LocalStack
docker run -d -p 4566:4566 localstack/localstack

# Set endpoint for local testing
export AWS_ENDPOINT_URL=http://localhost:4566
```

### Required Permissions
Ensure your AWS credentials have the following DynamoDB permissions:
- `dynamodb:CreateTable`
- `dynamodb:DeleteTable`
- `dynamodb:DescribeTable`
- `dynamodb:UpdateTable`
- `dynamodb:PutItem`
- `dynamodb:GetItem`
- `dynamodb:UpdateItem`
- `dynamodb:DeleteItem`
- `dynamodb:Query`
- `dynamodb:Scan`
- `dynamodb:BatchWriteItem`
- `dynamodb:BatchGetItem`
- `dynamodb:TransactWriteItems`
- `dynamodb:TransactGetItems`
- `dynamodb:ListTables`
- `dynamodb:CreateBackup`
- `dynamodb:DescribeBackup`
- `dynamodb:RestoreTableFromBackup`

## Contributing

This package is designed to be extended. When adding new functions:

1. Follow the `Function/FunctionE` pattern
2. Add comprehensive options structs for complex operations
3. Include corresponding test assertions
4. Add examples in the examples_test.go file
5. Update this README with usage patterns

## License

This package is part of the vasdeference module and follows the same licensing terms.