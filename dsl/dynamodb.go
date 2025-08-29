package dsl

import (
	"fmt"
	"time"
)

// DynamoDBBuilder provides a fluent interface for DynamoDB operations
type DynamoDBBuilder struct {
	ctx *TestContext
}

// Table creates a DynamoDB table builder for a specific table
func (db *DynamoDBBuilder) Table(name string) *DynamoTableBuilder {
	return &DynamoTableBuilder{
		ctx:       db.ctx,
		tableName: name,
		config:    &TableConfig{},
	}
}

// CreateTable starts building a new DynamoDB table
func (db *DynamoDBBuilder) CreateTable() *DynamoTableCreateBuilder {
	return &DynamoTableCreateBuilder{
		ctx:    db.ctx,
		config: &TableCreateConfig{},
	}
}

// DynamoTableBuilder builds operations on an existing DynamoDB table
type DynamoTableBuilder struct {
	ctx       *TestContext
	tableName string
	config    *TableConfig
}

type TableConfig struct {
	ConsistentRead bool
	Timeout        time.Duration
	RetryAttempts  int
}

// WithConsistentRead enables strongly consistent reads
func (dtb *DynamoTableBuilder) WithConsistentRead() *DynamoTableBuilder {
	dtb.config.ConsistentRead = true
	return dtb
}

// PutItem creates an item builder for putting items
func (dtb *DynamoTableBuilder) PutItem() *DynamoPutItemBuilder {
	return &DynamoPutItemBuilder{
		ctx:       dtb.ctx,
		tableName: dtb.tableName,
		item:      make(map[string]interface{}),
		config:    &PutItemConfig{},
	}
}

// GetItem creates an item builder for getting items
func (dtb *DynamoTableBuilder) GetItem() *DynamoGetItemBuilder {
	return &DynamoGetItemBuilder{
		ctx:       dtb.ctx,
		tableName: dtb.tableName,
		key:       make(map[string]interface{}),
		config:    &GetItemConfig{},
	}
}

// Query creates a query builder
func (dtb *DynamoTableBuilder) Query() *DynamoQueryBuilder {
	return &DynamoQueryBuilder{
		ctx:       dtb.ctx,
		tableName: dtb.tableName,
		config:    &QueryConfig{},
	}
}

// Scan creates a scan builder
func (dtb *DynamoTableBuilder) Scan() *DynamoScanBuilder {
	return &DynamoScanBuilder{
		ctx:       dtb.ctx,
		tableName: dtb.tableName,
		config:    &ScanConfig{},
	}
}

// Delete deletes the table
func (dtb *DynamoTableBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Table %s deleted", dtb.tableName),
		ctx:     dtb.ctx,
	}
}

// Exists checks if the table exists
func (dtb *DynamoTableBuilder) Exists() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    true, // Simulated existence check
		ctx:     dtb.ctx,
	}
}

// WaitUntilActive waits until the table is active
func (dtb *DynamoTableBuilder) WaitUntilActive() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    "Table is now active",
		ctx:     dtb.ctx,
	}
}

// DynamoPutItemBuilder builds put item operations
type DynamoPutItemBuilder struct {
	ctx       *TestContext
	tableName string
	item      map[string]interface{}
	config    *PutItemConfig
}

type PutItemConfig struct {
	ConditionExpression string
	ReturnValues        string
}

// WithAttribute adds an attribute to the item
func (dpib *DynamoPutItemBuilder) WithAttribute(name string, value interface{}) *DynamoPutItemBuilder {
	dpib.item[name] = value
	return dpib
}

// WithItem sets the entire item
func (dpib *DynamoPutItemBuilder) WithItem(item map[string]interface{}) *DynamoPutItemBuilder {
	dpib.item = item
	return dpib
}

// OnlyIfNotExists adds a condition that the item doesn't exist
func (dpib *DynamoPutItemBuilder) OnlyIfNotExists() *DynamoPutItemBuilder {
	dpib.config.ConditionExpression = "attribute_not_exists(#pk)"
	return dpib
}

// WithCondition adds a custom condition expression
func (dpib *DynamoPutItemBuilder) WithCondition(expression string) *DynamoPutItemBuilder {
	dpib.config.ConditionExpression = expression
	return dpib
}

// Execute puts the item
func (dpib *DynamoPutItemBuilder) Execute() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    dpib.item,
		Metadata: map[string]interface{}{
			"table_name":  dpib.tableName,
			"item_count":  len(dpib.item),
			"conditional": dpib.config.ConditionExpression != "",
		},
		ctx: dpib.ctx,
	}
}

// DynamoGetItemBuilder builds get item operations
type DynamoGetItemBuilder struct {
	ctx       *TestContext
	tableName string
	key       map[string]interface{}
	config    *GetItemConfig
}

type GetItemConfig struct {
	ProjectionExpression string
	ConsistentRead       bool
}

// WithKey sets a key attribute
func (dgib *DynamoGetItemBuilder) WithKey(name string, value interface{}) *DynamoGetItemBuilder {
	dgib.key[name] = value
	return dgib
}

// WithKeys sets all key attributes
func (dgib *DynamoGetItemBuilder) WithKeys(keys map[string]interface{}) *DynamoGetItemBuilder {
	dgib.key = keys
	return dgib
}

// ProjectOnly specifies which attributes to retrieve
func (dgib *DynamoGetItemBuilder) ProjectOnly(attributes ...string) *DynamoGetItemBuilder {
	dgib.config.ProjectionExpression = fmt.Sprintf("#%s", attributes[0])
	for i := 1; i < len(attributes); i++ {
		dgib.config.ProjectionExpression += fmt.Sprintf(", #%s", attributes[i])
	}
	return dgib
}

// WithConsistentRead enables consistent read
func (dgib *DynamoGetItemBuilder) WithConsistentRead() *DynamoGetItemBuilder {
	dgib.config.ConsistentRead = true
	return dgib
}

// Execute gets the item
func (dgib *DynamoGetItemBuilder) Execute() *ExecutionResult {
	// Simulate getting an item
	item := map[string]interface{}{
		"id":   dgib.key["id"],
		"name": "Sample Item",
		"data": "Some data",
	}
	
	return &ExecutionResult{
		Success: true,
		Data:    item,
		ctx:     dgib.ctx,
	}
}

// DynamoQueryBuilder builds query operations
type DynamoQueryBuilder struct {
	ctx       *TestContext
	tableName string
	config    *QueryConfig
}

type QueryConfig struct {
	KeyConditionExpression string
	FilterExpression       string
	IndexName              string
	ScanIndexForward       *bool
	Limit                  *int32
	ProjectionExpression   string
}

// WhereKey sets the key condition
func (dqb *DynamoQueryBuilder) WhereKey(keyName string) *DynamoKeyConditionBuilder {
	return &DynamoKeyConditionBuilder{
		queryBuilder: dqb,
		keyName:      keyName,
	}
}

// OnIndex specifies the index to query
func (dqb *DynamoQueryBuilder) OnIndex(indexName string) *DynamoQueryBuilder {
	dqb.config.IndexName = indexName
	return dqb
}

// Ascending sets the sort order to ascending
func (dqb *DynamoQueryBuilder) Ascending() *DynamoQueryBuilder {
	forward := true
	dqb.config.ScanIndexForward = &forward
	return dqb
}

// Descending sets the sort order to descending
func (dqb *DynamoQueryBuilder) Descending() *DynamoQueryBuilder {
	forward := false
	dqb.config.ScanIndexForward = &forward
	return dqb
}

// Limit sets the maximum number of items to return
func (dqb *DynamoQueryBuilder) Limit(limit int32) *DynamoQueryBuilder {
	dqb.config.Limit = &limit
	return dqb
}

// Execute executes the query
func (dqb *DynamoQueryBuilder) Execute() *ExecutionResult {
	// Simulate query results
	items := []map[string]interface{}{
		{"id": "1", "name": "Item 1"},
		{"id": "2", "name": "Item 2"},
	}
	
	return &ExecutionResult{
		Success: true,
		Data:    items,
		Metadata: map[string]interface{}{
			"count":      len(items),
			"table_name": dqb.tableName,
		},
		ctx: dqb.ctx,
	}
}

// DynamoKeyConditionBuilder builds key conditions for queries
type DynamoKeyConditionBuilder struct {
	queryBuilder *DynamoQueryBuilder
	keyName      string
}

// Equals sets an equality condition
func (dkcb *DynamoKeyConditionBuilder) Equals(value interface{}) *DynamoQueryBuilder {
	dkcb.queryBuilder.config.KeyConditionExpression = fmt.Sprintf("#%s = :val", dkcb.keyName)
	return dkcb.queryBuilder
}

// BeginsWith sets a begins_with condition
func (dkcb *DynamoKeyConditionBuilder) BeginsWith(prefix string) *DynamoQueryBuilder {
	dkcb.queryBuilder.config.KeyConditionExpression = fmt.Sprintf("begins_with(#%s, :prefix)", dkcb.keyName)
	return dkcb.queryBuilder
}

// Between sets a between condition
func (dkcb *DynamoKeyConditionBuilder) Between(low, high interface{}) *DynamoQueryBuilder {
	dkcb.queryBuilder.config.KeyConditionExpression = fmt.Sprintf("#%s BETWEEN :low AND :high", dkcb.keyName)
	return dkcb.queryBuilder
}

// DynamoScanBuilder builds scan operations
type DynamoScanBuilder struct {
	ctx       *TestContext
	tableName string
	config    *ScanConfig
}

type ScanConfig struct {
	FilterExpression     string
	ProjectionExpression string
	Limit                *int32
	Segment              *int32
	TotalSegments        *int32
}

// WithFilter adds a filter expression
func (dsb *DynamoScanBuilder) WithFilter(expression string) *DynamoScanBuilder {
	dsb.config.FilterExpression = expression
	return dsb
}

// Limit sets the maximum number of items to return
func (dsb *DynamoScanBuilder) Limit(limit int32) *DynamoScanBuilder {
	dsb.config.Limit = &limit
	return dsb
}

// Parallel enables parallel scanning
func (dsb *DynamoScanBuilder) Parallel(segment, totalSegments int32) *DynamoScanBuilder {
	dsb.config.Segment = &segment
	dsb.config.TotalSegments = &totalSegments
	return dsb
}

// Execute executes the scan
func (dsb *DynamoScanBuilder) Execute() *ExecutionResult {
	// Simulate scan results
	items := []map[string]interface{}{
		{"id": "1", "name": "Scanned Item 1"},
		{"id": "2", "name": "Scanned Item 2"},
		{"id": "3", "name": "Scanned Item 3"},
	}
	
	return &ExecutionResult{
		Success: true,
		Data:    items,
		Metadata: map[string]interface{}{
			"scanned_count": len(items),
			"table_name":    dsb.tableName,
		},
		ctx: dsb.ctx,
	}
}

// DynamoTableCreateBuilder builds table creation operations
type DynamoTableCreateBuilder struct {
	ctx    *TestContext
	config *TableCreateConfig
}

type TableCreateConfig struct {
	TableName             string
	AttributeDefinitions  []AttributeDefinition
	KeySchema             []KeySchemaElement
	BillingMode           string
	ProvisionedThroughput *ProvisionedThroughput
	Tags                  map[string]string
}

type AttributeDefinition struct {
	AttributeName string
	AttributeType string
}

type KeySchemaElement struct {
	AttributeName string
	KeyType       string
}

type ProvisionedThroughput struct {
	ReadCapacityUnits  int64
	WriteCapacityUnits int64
}

// Named sets the table name
func (dtcb *DynamoTableCreateBuilder) Named(name string) *DynamoTableCreateBuilder {
	dtcb.config.TableName = name
	return dtcb
}

// WithHashKey sets the hash key
func (dtcb *DynamoTableCreateBuilder) WithHashKey(name, attrType string) *DynamoTableCreateBuilder {
	dtcb.config.AttributeDefinitions = append(dtcb.config.AttributeDefinitions, AttributeDefinition{
		AttributeName: name,
		AttributeType: attrType,
	})
	dtcb.config.KeySchema = append(dtcb.config.KeySchema, KeySchemaElement{
		AttributeName: name,
		KeyType:       "HASH",
	})
	return dtcb
}

// WithRangeKey sets the range key
func (dtcb *DynamoTableCreateBuilder) WithRangeKey(name, attrType string) *DynamoTableCreateBuilder {
	dtcb.config.AttributeDefinitions = append(dtcb.config.AttributeDefinitions, AttributeDefinition{
		AttributeName: name,
		AttributeType: attrType,
	})
	dtcb.config.KeySchema = append(dtcb.config.KeySchema, KeySchemaElement{
		AttributeName: name,
		KeyType:       "RANGE",
	})
	return dtcb
}

// WithOnDemandBilling sets billing mode to on-demand
func (dtcb *DynamoTableCreateBuilder) WithOnDemandBilling() *DynamoTableCreateBuilder {
	dtcb.config.BillingMode = "PAY_PER_REQUEST"
	return dtcb
}

// WithProvisionedBilling sets billing mode to provisioned
func (dtcb *DynamoTableCreateBuilder) WithProvisionedBilling(readUnits, writeUnits int64) *DynamoTableCreateBuilder {
	dtcb.config.BillingMode = "PROVISIONED"
	dtcb.config.ProvisionedThroughput = &ProvisionedThroughput{
		ReadCapacityUnits:  readUnits,
		WriteCapacityUnits: writeUnits,
	}
	return dtcb
}

// Create creates the table
func (dtcb *DynamoTableCreateBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"TableName":   dtcb.config.TableName,
			"TableStatus": "CREATING",
			"BillingMode": dtcb.config.BillingMode,
		},
		ctx: dtcb.ctx,
	}
}