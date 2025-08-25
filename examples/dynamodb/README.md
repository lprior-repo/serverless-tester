# DynamoDB

CRUD operations, transactions, and query patterns for DynamoDB testing.

## Quick Start

```bash
# Run DynamoDB examples
go test -v ./examples/dynamodb/

# View patterns
go run ./examples/dynamodb/examples.go
```

## Basic Operations

### Put Item
```go
func TestPutItem(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("users")
    
    err := dynamodb.PutItem(vdf.Context, tableName, map[string]interface{}{
        "id":    "user-123",
        "name":  "John Doe",
        "email": "john@example.com",
    })
    assert.NoError(t, err)
}
```

### Get Item
```go
func TestGetItem(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("users")
    
    // Put item first
    dynamodb.PutItem(vdf.Context, tableName, map[string]interface{}{
        "id": "user-123", "name": "John Doe",
    })
    
    // Get item back
    key := map[string]interface{}{"id": "user-123"}
    item := dynamodb.GetItem(vdf.Context, tableName, key)
    
    assert.Equal(t, "John Doe", item["name"])
}
```

### Update Item
```go
func TestUpdateItem(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("users")
    key := map[string]interface{}{"id": "user-123"}
    
    // Update with condition
    err := dynamodb.UpdateItem(vdf.Context, tableName, key,
        "SET #name = :name, updated_at = :timestamp",
        map[string]string{"#name": "name"},
        map[string]interface{}{
            ":name":      "Jane Doe",
            ":timestamp": time.Now().Unix(),
        },
    )
    assert.NoError(t, err)
}
```

### Delete Item
```go
func TestDeleteItem(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("users")
    key := map[string]interface{}{"id": "user-123"}
    
    // Delete with condition
    err := dynamodb.DeleteItemWithCondition(vdf.Context, tableName, key,
        "attribute_exists(id)",
        nil, nil,
    )
    assert.NoError(t, err)
}
```

## Query Operations

### Simple Query
```go
func TestQuery(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("orders")
    
    // Query by partition key
    items := dynamodb.Query(vdf.Context, tableName,
        "customer_id = :customer_id",
        map[string]interface{}{
            ":customer_id": "cust-123",
        },
    )
    
    assert.NotEmpty(t, items)
}
```

### Query with Filter
```go
func TestQueryWithFilter(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("orders")
    
    items := dynamodb.QueryWithFilter(vdf.Context, tableName,
        "customer_id = :customer_id",
        "order_status = :status",
        map[string]interface{}{
            ":customer_id": "cust-123",
            ":status":      "completed",
        },
    )
    
    for _, item := range items {
        assert.Equal(t, "completed", item["order_status"])
    }
}
```

### Query with Pagination
```go
func TestQueryPagination(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("orders")
    
    // Get all pages
    allItems := dynamodb.QueryAllPages(vdf.Context, tableName,
        "customer_id = :customer_id",
        map[string]interface{}{
            ":customer_id": "cust-123",
        },
    )
    
    // Process in chunks of 25
    pageSize := 25
    for i := 0; i < len(allItems); i += pageSize {
        end := i + pageSize
        if end > len(allItems) {
            end = len(allItems)
        }
        
        chunk := allItems[i:end]
        // Process chunk...
    }
}
```

## Scan Operations

### Full Table Scan
```go
func TestScan(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("users")
    
    items := dynamodb.Scan(vdf.Context, tableName)
    
    assert.NotEmpty(t, items)
}
```

### Scan with Filter
```go
func TestScanWithFilter(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("users")
    
    items := dynamodb.ScanWithFilter(vdf.Context, tableName,
        "account_type = :type",
        map[string]interface{}{
            ":type": "premium",
        },
    )
    
    for _, item := range items {
        assert.Equal(t, "premium", item["account_type"])
    }
}
```

## Batch Operations

### Batch Write
```go
func TestBatchWrite(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("users")
    
    items := []map[string]interface{}{
        {"id": "user-001", "name": "Alice"},
        {"id": "user-002", "name": "Bob"},
        {"id": "user-003", "name": "Charlie"},
    }
    
    err := dynamodb.BatchWriteItems(vdf.Context, tableName, items)
    assert.NoError(t, err)
}
```

### Batch Read
```go
func TestBatchRead(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("users")
    
    keys := []map[string]interface{}{
        {"id": "user-001"},
        {"id": "user-002"},
        {"id": "user-003"},
    }
    
    items := dynamodb.BatchGetItems(vdf.Context, tableName, keys)
    assert.Len(t, items, 3)
}
```

## Transactions

### Simple Transaction
```go
func TestTransaction(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("accounts")
    
    // Transfer money between accounts
    err := dynamodb.TransactWrite(vdf.Context, []dynamodb.TransactItem{
        {
            TableName: tableName,
            Key:       map[string]interface{}{"id": "account-001"},
            Update:    "SET balance = balance - :amount",
            Condition: "balance >= :amount",
            Values: map[string]interface{}{
                ":amount": 100,
            },
        },
        {
            TableName: tableName,
            Key:       map[string]interface{}{"id": "account-002"},
            Update:    "SET balance = balance + :amount",
            Values: map[string]interface{}{
                ":amount": 100,
            },
        },
    })
    
    assert.NoError(t, err)
}
```

### Transaction with Condition Checks
```go
func TestTransactionWithConditions(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("inventory")
    
    err := dynamodb.TransactWrite(vdf.Context, []dynamodb.TransactItem{
        {
            TableName: tableName,
            Key:       map[string]interface{}{"product_id": "item-123"},
            Update:    "SET quantity = quantity - :qty",
            Condition: "quantity >= :qty",
            Values: map[string]interface{}{
                ":qty": 5,
            },
        },
        {
            TableName: tableName,
            Key:       map[string]interface{}{"product_id": "item-456"},
            Put: map[string]interface{}{
                "product_id": "item-456",
                "order_id":   "order-789",
                "quantity":   5,
                "timestamp":  time.Now().Unix(),
            },
            Condition: "attribute_not_exists(product_id)",
        },
    })
    
    assert.NoError(t, err)
}
```

## Advanced Patterns

### Conditional Updates
```go
func TestConditionalUpdate(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("counters")
    key := map[string]interface{}{"id": "page-views"}
    
    // Optimistic locking with version
    err := dynamodb.UpdateItemWithCondition(vdf.Context, tableName, key,
        "SET #count = #count + :inc, #version = #version + :one",
        "#version = :current_version",
        map[string]string{
            "#count":   "count",
            "#version": "version",
        },
        map[string]interface{}{
            ":inc":             1,
            ":one":             1,
            ":current_version": 5, // Expected current version
        },
    )
    
    if err != nil {
        // Handle optimistic locking failure
        assert.Contains(t, err.Error(), "ConditionalCheckFailed")
    }
}
```

### TTL (Time To Live)
```go
func TestTTLItems(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("sessions")
    
    // Create item that expires in 1 hour
    expiryTime := time.Now().Add(1 * time.Hour).Unix()
    
    err := dynamodb.PutItem(vdf.Context, tableName, map[string]interface{}{
        "session_id": "sess-123",
        "user_id":    "user-456",
        "data":       map[string]interface{}{"key": "value"},
        "ttl":        expiryTime, // TTL attribute
    })
    
    assert.NoError(t, err)
}
```

## Table Management

### Create Table
```go
func TestCreateTable(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("test-table")
    
    table := dynamodb.CreateTable(vdf.Context, &dynamodb.TableConfig{
        TableName:   tableName,
        BillingMode: "PAY_PER_REQUEST",
        AttributeDefinitions: []dynamodb.AttributeDefinition{
            {AttributeName: "id", AttributeType: "S"},
            {AttributeName: "sort_key", AttributeType: "S"},
        },
        KeySchema: []dynamodb.KeySchemaElement{
            {AttributeName: "id", KeyType: "HASH"},
            {AttributeName: "sort_key", KeyType: "RANGE"},
        },
    })
    
    vdf.RegisterCleanup(func() error {
        return table.Delete()
    })
    
    dynamodb.AssertTableExists(vdf.Context, tableName)
}
```

## Assertions

```go
func TestAssertions(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("test-table")
    
    // Table assertions
    dynamodb.AssertTableExists(vdf.Context, tableName)
    dynamodb.AssertTableStatus(vdf.Context, tableName, "ACTIVE")
    
    // Item assertions
    key := map[string]interface{}{"id": "test-item"}
    dynamodb.AssertItemExists(vdf.Context, tableName, key)
    dynamodb.AssertItemNotExists(vdf.Context, tableName, 
        map[string]interface{}{"id": "missing-item"})
    
    // Attribute assertions
    dynamodb.AssertAttributeEquals(vdf.Context, tableName, key, "name", "John")
    dynamodb.AssertAttributeExists(vdf.Context, tableName, key, "created_at")
    
    // Count assertions
    count := dynamodb.GetItemCount(vdf.Context, tableName)
    assert.Greater(t, count, 0)
}
```

## Error Handling

```go
func TestErrorScenarios(t *testing.T) {
    vdf := vasdeference.New(t)
    defer vdf.Cleanup()
    
    tableName := vdf.PrefixResourceName("test-table")
    key := map[string]interface{}{"id": "test-item"}
    
    // Test conditional check failure
    err := dynamodb.UpdateItemWithCondition(vdf.Context, tableName, key,
        "SET #name = :name",
        "attribute_not_exists(id)", // This will fail if item exists
        map[string]string{"#name": "name"},
        map[string]interface{}{":name": "New Name"},
    )
    
    assert.Error(t, err)
    dynamodb.AssertConditionalCheckFailed(vdf.Context, err)
    
    // Test resource not found
    err = dynamodb.GetItem(vdf.Context, "non-existent-table", key)
    assert.Error(t, err)
    dynamodb.AssertResourceNotFound(vdf.Context, err)
}
```

## Run Examples

```bash
# View all patterns
go run ./examples/dynamodb/examples.go

# Run as tests
go test -v ./examples/dynamodb/

# Test specific operations
go test -v ./examples/dynamodb/ -run TestPutItem
go test -v ./examples/dynamodb/ -run TestTransaction
```

## API Reference

### Basic Operations
- `dynamodb.PutItem(ctx, tableName, item)` - Create/update item
- `dynamodb.GetItem(ctx, tableName, key)` - Retrieve item
- `dynamodb.UpdateItem(ctx, tableName, key, expression, names, values)` - Update item
- `dynamodb.DeleteItem(ctx, tableName, key)` - Delete item

### Query/Scan Operations
- `dynamodb.Query(ctx, tableName, keyCondition, values)` - Query items
- `dynamodb.QueryWithFilter(ctx, tableName, keyCondition, filter, values)` - Query with filter
- `dynamodb.Scan(ctx, tableName)` - Scan table
- `dynamodb.ScanWithFilter(ctx, tableName, filter, values)` - Scan with filter

### Batch Operations
- `dynamodb.BatchWriteItems(ctx, tableName, items)` - Batch write
- `dynamodb.BatchGetItems(ctx, tableName, keys)` - Batch read

### Transactions
- `dynamodb.TransactWrite(ctx, items)` - Transaction write
- `dynamodb.TransactRead(ctx, items)` - Transaction read

### Assertions
- `dynamodb.AssertItemExists(ctx, tableName, key)` - Assert item exists
- `dynamodb.AssertAttributeEquals(ctx, tableName, key, attr, value)` - Assert attribute value

---

**Next:** [EventBridge Examples](../eventbridge/) | [Step Functions Examples](../stepfunctions/)