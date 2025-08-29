package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// DynamoDBAttributeType represents DynamoDB attribute types
type DynamoDBAttributeType int

const (
	AttributeTypeString DynamoDBAttributeType = iota
	AttributeTypeNumber
	AttributeTypeBinary
	AttributeTypeStringSet
	AttributeTypeNumberSet
	AttributeTypeBinarySet
	AttributeTypeMap
	AttributeTypeList
	AttributeTypeNull
	AttributeTypeBool
)

// String returns the string representation of the attribute type
func (at DynamoDBAttributeType) String() string {
	switch at {
	case AttributeTypeString:
		return "S"
	case AttributeTypeNumber:
		return "N"
	case AttributeTypeBinary:
		return "B"
	case AttributeTypeStringSet:
		return "SS"
	case AttributeTypeNumberSet:
		return "NS"
	case AttributeTypeBinarySet:
		return "BS"
	case AttributeTypeMap:
		return "M"
	case AttributeTypeList:
		return "L"
	case AttributeTypeNull:
		return "NULL"
	case AttributeTypeBool:
		return "BOOL"
	default:
		return "S"
	}
}

// ParseAttributeType parses a string into an attribute type
func ParseAttributeType(s string) mo.Option[DynamoDBAttributeType] {
	switch strings.ToUpper(s) {
	case "S", "STRING":
		return mo.Some(AttributeTypeString)
	case "N", "NUMBER":
		return mo.Some(AttributeTypeNumber)
	case "B", "BINARY":
		return mo.Some(AttributeTypeBinary)
	case "SS", "STRINGSET":
		return mo.Some(AttributeTypeStringSet)
	case "NS", "NUMBERSET":
		return mo.Some(AttributeTypeNumberSet)
	case "BS", "BINARYSET":
		return mo.Some(AttributeTypeBinarySet)
	case "M", "MAP":
		return mo.Some(AttributeTypeMap)
	case "L", "LIST":
		return mo.Some(AttributeTypeList)
	case "NULL":
		return mo.Some(AttributeTypeNull)
	case "BOOL", "BOOLEAN":
		return mo.Some(AttributeTypeBool)
	default:
		return mo.None[DynamoDBAttributeType]()
	}
}

// DynamoDBAttribute represents a DynamoDB attribute value
type DynamoDBAttribute struct {
	attributeType DynamoDBAttributeType
	value         interface{}
}

// NewDynamoDBAttribute creates a new attribute with type inference
func NewDynamoDBAttribute(value interface{}) DynamoDBAttribute {
	attributeType := inferAttributeType(value)
	return DynamoDBAttribute{
		attributeType: attributeType,
		value:         value,
	}
}

// NewDynamoDBAttributeWithType creates a new attribute with explicit type
func NewDynamoDBAttributeWithType(value interface{}, attrType DynamoDBAttributeType) DynamoDBAttribute {
	return DynamoDBAttribute{
		attributeType: attrType,
		value:         value,
	}
}

// GetType returns the attribute type
func (attr DynamoDBAttribute) GetType() DynamoDBAttributeType {
	return attr.attributeType
}

// GetValue returns the attribute value
func (attr DynamoDBAttribute) GetValue() interface{} {
	return attr.value
}

// inferAttributeType infers the DynamoDB attribute type from a Go value
func inferAttributeType(value interface{}) DynamoDBAttributeType {
	switch value.(type) {
	case string:
		return AttributeTypeString
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return AttributeTypeNumber
	case bool:
		return AttributeTypeBool
	case []byte:
		return AttributeTypeBinary
	case []string:
		return AttributeTypeStringSet
	case []int, []int32, []int64, []float32, []float64:
		return AttributeTypeNumberSet
	case [][]byte:
		return AttributeTypeBinarySet
	case map[string]interface{}:
		return AttributeTypeMap
	case []interface{}:
		return AttributeTypeList
	case nil:
		return AttributeTypeNull
	default:
		return AttributeTypeString
	}
}

// DynamoDBItem represents a DynamoDB item as an immutable map
type DynamoDBItem struct {
	attributes map[string]DynamoDBAttribute
}

// NewDynamoDBItem creates a new empty item
func NewDynamoDBItem() DynamoDBItem {
	return DynamoDBItem{
		attributes: make(map[string]DynamoDBAttribute),
	}
}

// WithAttribute adds an attribute to the item (returns new item)
func (item DynamoDBItem) WithAttribute(name string, value interface{}) DynamoDBItem {
	newAttributes := lo.Assign(item.attributes, map[string]DynamoDBAttribute{
		name: NewDynamoDBAttribute(value),
	})
	return DynamoDBItem{attributes: newAttributes}
}

// WithAttributeTyped adds a typed attribute to the item
func (item DynamoDBItem) WithAttributeTyped(name string, value interface{}, attrType DynamoDBAttributeType) DynamoDBItem {
	newAttributes := lo.Assign(item.attributes, map[string]DynamoDBAttribute{
		name: NewDynamoDBAttributeWithType(value, attrType),
	})
	return DynamoDBItem{attributes: newAttributes}
}

// GetAttribute gets an attribute from the item
func (item DynamoDBItem) GetAttribute(name string) mo.Option[DynamoDBAttribute] {
	if attr, exists := item.attributes[name]; exists {
		return mo.Some(attr)
	}
	return mo.None[DynamoDBAttribute]()
}

// GetAllAttributes returns all attributes
func (item DynamoDBItem) GetAllAttributes() map[string]DynamoDBAttribute {
	return lo.Assign(map[string]DynamoDBAttribute{}, item.attributes)
}

// HasAttribute checks if an attribute exists
func (item DynamoDBItem) HasAttribute(name string) bool {
	_, exists := item.attributes[name]
	return exists
}

// Size returns the number of attributes
func (item DynamoDBItem) Size() int {
	return len(item.attributes)
}

// DynamoDBKeySchema represents a key schema element
type DynamoDBKeySchema struct {
	attributeName string
	keyType       KeyType
}

// KeyType represents the key type (HASH or RANGE)
type KeyType int

const (
	KeyTypeHash KeyType = iota
	KeyTypeRange
)

// String returns the string representation of key type
func (kt KeyType) String() string {
	switch kt {
	case KeyTypeHash:
		return "HASH"
	case KeyTypeRange:
		return "RANGE"
	default:
		return "HASH"
	}
}

// NewHashKey creates a new hash key schema
func NewHashKey(attributeName string) DynamoDBKeySchema {
	return DynamoDBKeySchema{
		attributeName: attributeName,
		keyType:       KeyTypeHash,
	}
}

// NewRangeKey creates a new range key schema
func NewRangeKey(attributeName string) DynamoDBKeySchema {
	return DynamoDBKeySchema{
		attributeName: attributeName,
		keyType:       KeyTypeRange,
	}
}

// DynamoDBAttributeDefinition represents an attribute definition
type DynamoDBAttributeDefinition struct {
	attributeName string
	attributeType DynamoDBAttributeType
}

// NewAttributeDefinition creates a new attribute definition
func NewAttributeDefinition(name string, attrType DynamoDBAttributeType) DynamoDBAttributeDefinition {
	return DynamoDBAttributeDefinition{
		attributeName: name,
		attributeType: attrType,
	}
}

// DynamoDBTableConfiguration represents table configuration
type DynamoDBTableConfiguration struct {
	tableName             string
	attributeDefinitions  []DynamoDBAttributeDefinition
	keySchema             []DynamoDBKeySchema
	billingMode           BillingMode
	provisionedThroughput mo.Option[ProvisionedThroughput]
	tags                  map[string]string
}

// BillingMode represents DynamoDB billing modes
type BillingMode int

const (
	BillingModeProvisioned BillingMode = iota
	BillingModeOnDemand
)

// String returns the string representation of billing mode
func (bm BillingMode) String() string {
	switch bm {
	case BillingModeProvisioned:
		return "PROVISIONED"
	case BillingModeOnDemand:
		return "PAY_PER_REQUEST"
	default:
		return "PAY_PER_REQUEST"
	}
}

// ProvisionedThroughput represents provisioned throughput settings
type ProvisionedThroughput struct {
	readCapacityUnits  int64
	writeCapacityUnits int64
}

// NewProvisionedThroughput creates new provisioned throughput
func NewProvisionedThroughput(readUnits, writeUnits int64) ProvisionedThroughput {
	return ProvisionedThroughput{
		readCapacityUnits:  readUnits,
		writeCapacityUnits: writeUnits,
	}
}

// NewDynamoDBTableConfiguration creates a new table configuration
func NewDynamoDBTableConfiguration(tableName string) DynamoDBTableConfiguration {
	return DynamoDBTableConfiguration{
		tableName:             tableName,
		attributeDefinitions:  make([]DynamoDBAttributeDefinition, 0),
		keySchema:             make([]DynamoDBKeySchema, 0),
		billingMode:           BillingModeOnDemand,
		provisionedThroughput: mo.None[ProvisionedThroughput](),
		tags:                  make(map[string]string),
	}
}

// DynamoDBTableOption represents functional options for table configuration
type DynamoDBTableOption func(DynamoDBTableConfiguration) DynamoDBTableConfiguration

// WithHashKey adds a hash key to the table
func WithHashKey(attributeName string, attributeType DynamoDBAttributeType) DynamoDBTableOption {
	return func(config DynamoDBTableConfiguration) DynamoDBTableConfiguration {
		return DynamoDBTableConfiguration{
			tableName: config.tableName,
			attributeDefinitions: append(config.attributeDefinitions, 
				NewAttributeDefinition(attributeName, attributeType)),
			keySchema: append(config.keySchema, NewHashKey(attributeName)),
			billingMode:           config.billingMode,
			provisionedThroughput: config.provisionedThroughput,
			tags:                  config.tags,
		}
	}
}

// WithRangeKey adds a range key to the table
func WithRangeKey(attributeName string, attributeType DynamoDBAttributeType) DynamoDBTableOption {
	return func(config DynamoDBTableConfiguration) DynamoDBTableConfiguration {
		return DynamoDBTableConfiguration{
			tableName: config.tableName,
			attributeDefinitions: append(config.attributeDefinitions, 
				NewAttributeDefinition(attributeName, attributeType)),
			keySchema: append(config.keySchema, NewRangeKey(attributeName)),
			billingMode:           config.billingMode,
			provisionedThroughput: config.provisionedThroughput,
			tags:                  config.tags,
		}
	}
}

// WithOnDemandBilling sets billing mode to on-demand
func WithOnDemandBilling() DynamoDBTableOption {
	return func(config DynamoDBTableConfiguration) DynamoDBTableConfiguration {
		return DynamoDBTableConfiguration{
			tableName:             config.tableName,
			attributeDefinitions:  config.attributeDefinitions,
			keySchema:             config.keySchema,
			billingMode:           BillingModeOnDemand,
			provisionedThroughput: mo.None[ProvisionedThroughput](),
			tags:                  config.tags,
		}
	}
}

// WithProvisionedBilling sets billing mode to provisioned
func WithProvisionedBilling(readUnits, writeUnits int64) DynamoDBTableOption {
	return func(config DynamoDBTableConfiguration) DynamoDBTableConfiguration {
		return DynamoDBTableConfiguration{
			tableName:             config.tableName,
			attributeDefinitions:  config.attributeDefinitions,
			keySchema:             config.keySchema,
			billingMode:           BillingModeProvisioned,
			provisionedThroughput: mo.Some(NewProvisionedThroughput(readUnits, writeUnits)),
			tags:                  config.tags,
		}
	}
}

// WithTableTags adds tags to the table
func WithTableTags(tags map[string]string) DynamoDBTableOption {
	return func(config DynamoDBTableConfiguration) DynamoDBTableConfiguration {
		return DynamoDBTableConfiguration{
			tableName:             config.tableName,
			attributeDefinitions:  config.attributeDefinitions,
			keySchema:             config.keySchema,
			billingMode:           config.billingMode,
			provisionedThroughput: config.provisionedThroughput,
			tags:                  lo.Assign(config.tags, tags),
		}
	}
}

// DynamoDBQueryCondition represents query conditions using functional composition
type DynamoDBQueryCondition struct {
	keyName     string
	operator    QueryOperator
	values      []interface{}
	expression  mo.Option[string]
}

// QueryOperator represents query operators
type QueryOperator int

const (
	QueryOperatorEquals QueryOperator = iota
	QueryOperatorLessThan
	QueryOperatorLessThanOrEqual
	QueryOperatorGreaterThan
	QueryOperatorGreaterThanOrEqual
	QueryOperatorBetween
	QueryOperatorBeginsWith
)

// String returns the string representation of the operator
func (qo QueryOperator) String() string {
	switch qo {
	case QueryOperatorEquals:
		return "="
	case QueryOperatorLessThan:
		return "<"
	case QueryOperatorLessThanOrEqual:
		return "<="
	case QueryOperatorGreaterThan:
		return ">"
	case QueryOperatorGreaterThanOrEqual:
		return ">="
	case QueryOperatorBetween:
		return "BETWEEN"
	case QueryOperatorBeginsWith:
		return "begins_with"
	default:
		return "="
	}
}

// NewQueryCondition creates a new query condition
func NewQueryCondition(keyName string, operator QueryOperator, values ...interface{}) DynamoDBQueryCondition {
	return DynamoDBQueryCondition{
		keyName:     keyName,
		operator:    operator,
		values:      values,
		expression:  mo.None[string](),
	}
}

// ToExpression converts the condition to an expression string
func (qc DynamoDBQueryCondition) ToExpression() string {
	if qc.expression.IsPresent() {
		return qc.expression.MustGet()
	}
	
	switch qc.operator {
	case QueryOperatorEquals:
		return fmt.Sprintf("#%s = :%s", qc.keyName, qc.keyName)
	case QueryOperatorBetween:
		return fmt.Sprintf("#%s BETWEEN :%s_low AND :%s_high", qc.keyName, qc.keyName, qc.keyName)
	case QueryOperatorBeginsWith:
		return fmt.Sprintf("begins_with(#%s, :%s_prefix)", qc.keyName, qc.keyName)
	default:
		return fmt.Sprintf("#%s %s :%s", qc.keyName, qc.operator.String(), qc.keyName)
	}
}

// DynamoDBQueryConfiguration represents query configuration
type DynamoDBQueryConfiguration struct {
	tableName               string
	keyCondition            mo.Option[DynamoDBQueryCondition]
	filterExpression        mo.Option[string]
	projectionExpression    mo.Option[string]
	indexName               mo.Option[string]
	scanIndexForward        mo.Option[bool]
	limit                   mo.Option[int32]
	consistentRead          bool
}

// NewDynamoDBQueryConfiguration creates a new query configuration
func NewDynamoDBQueryConfiguration(tableName string) DynamoDBQueryConfiguration {
	return DynamoDBQueryConfiguration{
		tableName:               tableName,
		keyCondition:            mo.None[DynamoDBQueryCondition](),
		filterExpression:        mo.None[string](),
		projectionExpression:    mo.None[string](),
		indexName:               mo.None[string](),
		scanIndexForward:        mo.None[bool](),
		limit:                   mo.None[int32](),
		consistentRead:          false,
	}
}

// DynamoDBQueryOption represents functional options for queries
type DynamoDBQueryOption func(DynamoDBQueryConfiguration) DynamoDBQueryConfiguration

// WithKeyCondition sets the key condition for the query
func WithKeyCondition(condition DynamoDBQueryCondition) DynamoDBQueryOption {
	return func(config DynamoDBQueryConfiguration) DynamoDBQueryConfiguration {
		return DynamoDBQueryConfiguration{
			tableName:               config.tableName,
			keyCondition:            mo.Some(condition),
			filterExpression:        config.filterExpression,
			projectionExpression:    config.projectionExpression,
			indexName:               config.indexName,
			scanIndexForward:        config.scanIndexForward,
			limit:                   config.limit,
			consistentRead:          config.consistentRead,
		}
	}
}

// WithQueryLimit sets the limit for the query
func WithQueryLimit(limit int32) DynamoDBQueryOption {
	return func(config DynamoDBQueryConfiguration) DynamoDBQueryConfiguration {
		return DynamoDBQueryConfiguration{
			tableName:               config.tableName,
			keyCondition:            config.keyCondition,
			filterExpression:        config.filterExpression,
			projectionExpression:    config.projectionExpression,
			indexName:               config.indexName,
			scanIndexForward:        config.scanIndexForward,
			limit:                   mo.Some(limit),
			consistentRead:          config.consistentRead,
		}
	}
}

// WithConsistentRead enables consistent read
func WithConsistentRead() DynamoDBQueryOption {
	return func(config DynamoDBQueryConfiguration) DynamoDBQueryConfiguration {
		return DynamoDBQueryConfiguration{
			tableName:               config.tableName,
			keyCondition:            config.keyCondition,
			filterExpression:        config.filterExpression,
			projectionExpression:    config.projectionExpression,
			indexName:               config.indexName,
			scanIndexForward:        config.scanIndexForward,
			limit:                   config.limit,
			consistentRead:          true,
		}
	}
}

// WithIndex sets the index name for the query
func WithIndex(indexName string) DynamoDBQueryOption {
	return func(config DynamoDBQueryConfiguration) DynamoDBQueryConfiguration {
		return DynamoDBQueryConfiguration{
			tableName:               config.tableName,
			keyCondition:            config.keyCondition,
			filterExpression:        config.filterExpression,
			projectionExpression:    config.projectionExpression,
			indexName:               mo.Some(indexName),
			scanIndexForward:        config.scanIndexForward,
			limit:                   config.limit,
			consistentRead:          config.consistentRead,
		}
	}
}

// WithAscendingOrder sets the scan order to ascending
func WithAscendingOrder() DynamoDBQueryOption {
	return func(config DynamoDBQueryConfiguration) DynamoDBQueryConfiguration {
		return DynamoDBQueryConfiguration{
			tableName:               config.tableName,
			keyCondition:            config.keyCondition,
			filterExpression:        config.filterExpression,
			projectionExpression:    config.projectionExpression,
			indexName:               config.indexName,
			scanIndexForward:        mo.Some(true),
			limit:                   config.limit,
			consistentRead:          config.consistentRead,
		}
	}
}

// WithDescendingOrder sets the scan order to descending
func WithDescendingOrder() DynamoDBQueryOption {
	return func(config DynamoDBQueryConfiguration) DynamoDBQueryConfiguration {
		return DynamoDBQueryConfiguration{
			tableName:               config.tableName,
			keyCondition:            config.keyCondition,
			filterExpression:        config.filterExpression,
			projectionExpression:    config.projectionExpression,
			indexName:               config.indexName,
			scanIndexForward:        mo.Some(false),
			limit:                   config.limit,
			consistentRead:          config.consistentRead,
		}
	}
}

// DynamoDBService represents pure functional DynamoDB operations
type DynamoDBService struct {
	// In a real implementation, this would contain AWS SDK clients
}

// NewDynamoDBService creates a new DynamoDB service
func NewDynamoDBService() DynamoDBService {
	return DynamoDBService{}
}

// CreateTable creates a DynamoDB table using functional composition
func (ds DynamoDBService) CreateTable(config DynamoDBTableConfiguration, testConfig TestConfig) ExecutionResult[map[string]interface{}] {
	startTime := time.Now()
	
	// Validate configuration
	if len(config.keySchema) == 0 {
		return NewErrorResult[map[string]interface{}](
			fmt.Errorf("table must have at least one key"),
			time.Since(startTime),
			map[string]interface{}{
				"table_name": config.tableName,
				"error_type": "validation_error",
			},
			testConfig,
		)
	}
	
	// Simulate table creation
	result := map[string]interface{}{
		"TableName":   config.tableName,
		"TableStatus": "CREATING",
		"BillingMode": config.billingMode.String(),
		"KeySchema":   lo.Map(config.keySchema, func(ks DynamoDBKeySchema, _ int) map[string]string {
			return map[string]string{
				"AttributeName": ks.attributeName,
				"KeyType":       ks.keyType.String(),
			}
		}),
		"AttributeDefinitions": lo.Map(config.attributeDefinitions, func(ad DynamoDBAttributeDefinition, _ int) map[string]string {
			return map[string]string{
				"AttributeName": ad.attributeName,
				"AttributeType": ad.attributeType.String(),
			}
		}),
	}
	
	metadata := map[string]interface{}{
		"table_name":    config.tableName,
		"billing_mode":  config.billingMode.String(),
		"key_count":     len(config.keySchema),
		"operation":     "create_table",
	}
	
	return NewSuccessResult(
		result,
		time.Since(startTime),
		metadata,
		testConfig,
	)
}

// PutItem puts an item into a DynamoDB table
func (ds DynamoDBService) PutItem(tableName string, item DynamoDBItem, testConfig TestConfig) ExecutionResult[DynamoDBItem] {
	startTime := time.Now()
	
	// Validate item
	if item.Size() == 0 {
		return NewErrorResult[DynamoDBItem](
			fmt.Errorf("item cannot be empty"),
			time.Since(startTime),
			map[string]interface{}{
				"table_name": tableName,
				"error_type": "validation_error",
			},
			testConfig,
		)
	}
	
	// Simulate put item operation
	metadata := map[string]interface{}{
		"table_name":  tableName,
		"item_count":  item.Size(),
		"operation":   "put_item",
		"conditional": false,
	}
	
	return NewSuccessResult(
		item,
		time.Since(startTime),
		metadata,
		testConfig,
	)
}

// GetItem gets an item from a DynamoDB table
func (ds DynamoDBService) GetItem(tableName string, key DynamoDBItem, testConfig TestConfig) ExecutionResult[mo.Option[DynamoDBItem]] {
	startTime := time.Now()
	
	// Simulate get item operation - return a mock item
	mockItem := NewDynamoDBItem().
		WithAttribute("id", "test-id").
		WithAttribute("name", "Test Item").
		WithAttribute("timestamp", time.Now().Unix())
	
	metadata := map[string]interface{}{
		"table_name": tableName,
		"operation":  "get_item",
		"found":      true,
	}
	
	return NewSuccessResult(
		mo.Some(mockItem),
		time.Since(startTime),
		metadata,
		testConfig,
	)
}

// Query executes a query on a DynamoDB table
func (ds DynamoDBService) Query(config DynamoDBQueryConfiguration, testConfig TestConfig) ExecutionResult[[]DynamoDBItem] {
	startTime := time.Now()
	
	// Simulate query results
	items := lo.Times(3, func(i int) DynamoDBItem {
		return NewDynamoDBItem().
			WithAttribute("id", fmt.Sprintf("item-%d", i+1)).
			WithAttribute("name", fmt.Sprintf("Item %d", i+1)).
			WithAttribute("timestamp", time.Now().Unix()+int64(i))
	})
	
	metadata := map[string]interface{}{
		"table_name":  config.tableName,
		"count":       len(items),
		"operation":   "query",
		"index_used":  config.indexName.IsPresent(),
	}
	
	return NewSuccessResult(
		items,
		time.Since(startTime),
		metadata,
		testConfig,
	)
}

// Scan executes a scan on a DynamoDB table
func (ds DynamoDBService) Scan(tableName string, testConfig TestConfig) ExecutionResult[[]DynamoDBItem] {
	startTime := time.Now()
	
	// Simulate scan results
	items := lo.Times(5, func(i int) DynamoDBItem {
		return NewDynamoDBItem().
			WithAttribute("id", fmt.Sprintf("scan-item-%d", i+1)).
			WithAttribute("data", fmt.Sprintf("Scanned Data %d", i+1))
	})
	
	metadata := map[string]interface{}{
		"table_name":     tableName,
		"scanned_count":  len(items),
		"operation":      "scan",
	}
	
	return NewSuccessResult(
		items,
		time.Since(startTime),
		metadata,
		testConfig,
	)
}

// High-level functional DynamoDB operations

// CreateDynamoDBTable creates a DynamoDB table using functional composition
func CreateDynamoDBTable(tableName string, opts ...DynamoDBTableOption) func(TestConfig) ExecutionResult[map[string]interface{}] {
	return func(testConfig TestConfig) ExecutionResult[map[string]interface{}] {
		config := lo.Reduce(opts, func(config DynamoDBTableConfiguration, opt DynamoDBTableOption, _ int) DynamoDBTableConfiguration {
			return opt(config)
		}, NewDynamoDBTableConfiguration(tableName))
		
		service := NewDynamoDBService()
		return service.CreateTable(config, testConfig)
	}
}

// PutDynamoDBItem puts an item into a DynamoDB table
func PutDynamoDBItem(tableName string, item DynamoDBItem) func(TestConfig) ExecutionResult[DynamoDBItem] {
	return func(testConfig TestConfig) ExecutionResult[DynamoDBItem] {
		service := NewDynamoDBService()
		return service.PutItem(tableName, item, testConfig)
	}
}

// GetDynamoDBItem gets an item from a DynamoDB table
func GetDynamoDBItem(tableName string, key DynamoDBItem) func(TestConfig) ExecutionResult[mo.Option[DynamoDBItem]] {
	return func(testConfig TestConfig) ExecutionResult[mo.Option[DynamoDBItem]] {
		service := NewDynamoDBService()
		return service.GetItem(tableName, key, testConfig)
	}
}

// QueryDynamoDBTable executes a query on a DynamoDB table
func QueryDynamoDBTable(tableName string, opts ...DynamoDBQueryOption) func(TestConfig) ExecutionResult[[]DynamoDBItem] {
	return func(testConfig TestConfig) ExecutionResult[[]DynamoDBItem] {
		config := lo.Reduce(opts, func(config DynamoDBQueryConfiguration, opt DynamoDBQueryOption, _ int) DynamoDBQueryConfiguration {
			return opt(config)
		}, NewDynamoDBQueryConfiguration(tableName))
		
		service := NewDynamoDBService()
		return service.Query(config, testConfig)
	}
}

// ScanDynamoDBTable executes a scan on a DynamoDB table
func ScanDynamoDBTable(tableName string) func(TestConfig) ExecutionResult[[]DynamoDBItem] {
	return func(testConfig TestConfig) ExecutionResult[[]DynamoDBItem] {
		service := NewDynamoDBService()
		return service.Scan(tableName, testConfig)
	}
}

// Convenience functions for creating query conditions

// EqualsCondition creates an equals condition
func EqualsCondition(keyName string, value interface{}) DynamoDBQueryCondition {
	return NewQueryCondition(keyName, QueryOperatorEquals, value)
}

// BetweenCondition creates a between condition
func BetweenCondition(keyName string, lowValue, highValue interface{}) DynamoDBQueryCondition {
	return NewQueryCondition(keyName, QueryOperatorBetween, lowValue, highValue)
}

// BeginsWithCondition creates a begins_with condition
func BeginsWithCondition(keyName string, prefix string) DynamoDBQueryCondition {
	return NewQueryCondition(keyName, QueryOperatorBeginsWith, prefix)
}