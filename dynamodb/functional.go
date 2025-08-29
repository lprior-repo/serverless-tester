package dynamodb

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// FunctionalDynamoDBConfig represents immutable DynamoDB configuration using functional options
type FunctionalDynamoDBConfig struct {
	tableName           string
	timeout             time.Duration
	maxRetries          int
	retryDelay          time.Duration
	consistentRead      bool
	billingMode         types.BillingMode
	projectionExpression mo.Option[string]
	filterExpression    mo.Option[string]
	conditionExpression mo.Option[string]
	metadata           map[string]interface{}
}

// DynamoDBConfigOption represents a functional option for FunctionalDynamoDBConfig
type DynamoDBConfigOption func(FunctionalDynamoDBConfig) FunctionalDynamoDBConfig

// Constants for functional DynamoDB operations
const (
	FunctionalDynamoDBDefaultTimeout    = 30 * time.Second
	FunctionalDynamoDBDefaultRetries    = 3
	FunctionalDynamoDBDefaultRetryDelay = 1 * time.Second
)

// NewFunctionalDynamoDBConfig creates a new DynamoDB config with functional options
func NewFunctionalDynamoDBConfig(tableName string, opts ...DynamoDBConfigOption) FunctionalDynamoDBConfig {
	base := FunctionalDynamoDBConfig{
		tableName:           tableName,
		timeout:             FunctionalDynamoDBDefaultTimeout,
		maxRetries:          FunctionalDynamoDBDefaultRetries,
		retryDelay:          FunctionalDynamoDBDefaultRetryDelay,
		consistentRead:      false,
		billingMode:         types.BillingModePayPerRequest,
		projectionExpression: mo.None[string](),
		filterExpression:    mo.None[string](),
		conditionExpression: mo.None[string](),
		metadata:           make(map[string]interface{}),
	}

	return lo.Reduce(opts, func(config FunctionalDynamoDBConfig, opt DynamoDBConfigOption, _ int) FunctionalDynamoDBConfig {
		return opt(config)
	}, base)
}

// WithDynamoDBTimeout sets the timeout functionally
func WithDynamoDBTimeout(timeout time.Duration) DynamoDBConfigOption {
	return func(config FunctionalDynamoDBConfig) FunctionalDynamoDBConfig {
		return FunctionalDynamoDBConfig{
			tableName:           config.tableName,
			timeout:             timeout,
			maxRetries:          config.maxRetries,
			retryDelay:          config.retryDelay,
			consistentRead:      config.consistentRead,
			billingMode:         config.billingMode,
			projectionExpression: config.projectionExpression,
			filterExpression:    config.filterExpression,
			conditionExpression: config.conditionExpression,
			metadata:           config.metadata,
		}
	}
}

// WithDynamoDBRetryConfig sets retry configuration functionally
func WithDynamoDBRetryConfig(maxRetries int, retryDelay time.Duration) DynamoDBConfigOption {
	return func(config FunctionalDynamoDBConfig) FunctionalDynamoDBConfig {
		return FunctionalDynamoDBConfig{
			tableName:           config.tableName,
			timeout:             config.timeout,
			maxRetries:          maxRetries,
			retryDelay:          retryDelay,
			consistentRead:      config.consistentRead,
			billingMode:         config.billingMode,
			projectionExpression: config.projectionExpression,
			filterExpression:    config.filterExpression,
			conditionExpression: config.conditionExpression,
			metadata:           config.metadata,
		}
	}
}

// WithConsistentRead enables consistent read functionally
func WithConsistentRead(consistentRead bool) DynamoDBConfigOption {
	return func(config FunctionalDynamoDBConfig) FunctionalDynamoDBConfig {
		return FunctionalDynamoDBConfig{
			tableName:           config.tableName,
			timeout:             config.timeout,
			maxRetries:          config.maxRetries,
			retryDelay:          config.retryDelay,
			consistentRead:      consistentRead,
			billingMode:         config.billingMode,
			projectionExpression: config.projectionExpression,
			filterExpression:    config.filterExpression,
			conditionExpression: config.conditionExpression,
			metadata:           config.metadata,
		}
	}
}

// WithBillingMode sets billing mode functionally
func WithBillingMode(billingMode types.BillingMode) DynamoDBConfigOption {
	return func(config FunctionalDynamoDBConfig) FunctionalDynamoDBConfig {
		return FunctionalDynamoDBConfig{
			tableName:           config.tableName,
			timeout:             config.timeout,
			maxRetries:          config.maxRetries,
			retryDelay:          config.retryDelay,
			consistentRead:      config.consistentRead,
			billingMode:         billingMode,
			projectionExpression: config.projectionExpression,
			filterExpression:    config.filterExpression,
			conditionExpression: config.conditionExpression,
			metadata:           config.metadata,
		}
	}
}

// WithProjectionExpression sets projection expression functionally
func WithProjectionExpression(expression string) DynamoDBConfigOption {
	return func(config FunctionalDynamoDBConfig) FunctionalDynamoDBConfig {
		return FunctionalDynamoDBConfig{
			tableName:           config.tableName,
			timeout:             config.timeout,
			maxRetries:          config.maxRetries,
			retryDelay:          config.retryDelay,
			consistentRead:      config.consistentRead,
			billingMode:         config.billingMode,
			projectionExpression: mo.Some(expression),
			filterExpression:    config.filterExpression,
			conditionExpression: config.conditionExpression,
			metadata:           config.metadata,
		}
	}
}

// WithFilterExpression sets filter expression functionally
func WithFilterExpression(expression string) DynamoDBConfigOption {
	return func(config FunctionalDynamoDBConfig) FunctionalDynamoDBConfig {
		return FunctionalDynamoDBConfig{
			tableName:           config.tableName,
			timeout:             config.timeout,
			maxRetries:          config.maxRetries,
			retryDelay:          config.retryDelay,
			consistentRead:      config.consistentRead,
			billingMode:         config.billingMode,
			projectionExpression: config.projectionExpression,
			filterExpression:    mo.Some(expression),
			conditionExpression: config.conditionExpression,
			metadata:           config.metadata,
		}
	}
}

// WithConditionExpression sets condition expression functionally
func WithConditionExpression(expression string) DynamoDBConfigOption {
	return func(config FunctionalDynamoDBConfig) FunctionalDynamoDBConfig {
		return FunctionalDynamoDBConfig{
			tableName:           config.tableName,
			timeout:             config.timeout,
			maxRetries:          config.maxRetries,
			retryDelay:          config.retryDelay,
			consistentRead:      config.consistentRead,
			billingMode:         config.billingMode,
			projectionExpression: config.projectionExpression,
			filterExpression:    config.filterExpression,
			conditionExpression: mo.Some(expression),
			metadata:           config.metadata,
		}
	}
}

// WithDynamoDBMetadata adds metadata functionally
func WithDynamoDBMetadata(key string, value interface{}) DynamoDBConfigOption {
	return func(config FunctionalDynamoDBConfig) FunctionalDynamoDBConfig {
		newMetadata := lo.Assign(config.metadata, map[string]interface{}{key: value})
		return FunctionalDynamoDBConfig{
			tableName:           config.tableName,
			timeout:             config.timeout,
			maxRetries:          config.maxRetries,
			retryDelay:          config.retryDelay,
			consistentRead:      config.consistentRead,
			billingMode:         config.billingMode,
			projectionExpression: config.projectionExpression,
			filterExpression:    config.filterExpression,
			conditionExpression: config.conditionExpression,
			metadata:           newMetadata,
		}
	}
}

// FunctionalDynamoDBItem represents an immutable DynamoDB item
type FunctionalDynamoDBItem struct {
	attributes map[string]FunctionalDynamoDBAttribute
}

// FunctionalDynamoDBAttribute represents an attribute value with type safety
type FunctionalDynamoDBAttribute struct {
	attributeType types.ScalarAttributeType
	value         interface{}
}

// NewFunctionalDynamoDBAttribute creates a new attribute with type inference
func NewFunctionalDynamoDBAttribute(value interface{}) FunctionalDynamoDBAttribute {
	var attrType types.ScalarAttributeType
	
	switch value.(type) {
	case string:
		attrType = types.ScalarAttributeTypeS
	case int, int32, int64, float32, float64:
		attrType = types.ScalarAttributeTypeN
	case []byte:
		attrType = types.ScalarAttributeTypeB
	default:
		attrType = types.ScalarAttributeTypeS // Default to string
	}
	
	return FunctionalDynamoDBAttribute{
		attributeType: attrType,
		value:         value,
	}
}

// GetValue returns the attribute value
func (fda FunctionalDynamoDBAttribute) GetValue() interface{} {
	return fda.value
}

// GetType returns the attribute type
func (fda FunctionalDynamoDBAttribute) GetType() types.ScalarAttributeType {
	return fda.attributeType
}

// NewFunctionalDynamoDBItem creates a new immutable item
func NewFunctionalDynamoDBItem() FunctionalDynamoDBItem {
	return FunctionalDynamoDBItem{
		attributes: make(map[string]FunctionalDynamoDBAttribute),
	}
}

// WithAttribute adds an attribute to the item functionally
func (fdi FunctionalDynamoDBItem) WithAttribute(name string, value interface{}) FunctionalDynamoDBItem {
	newAttributes := lo.Assign(fdi.attributes, map[string]FunctionalDynamoDBAttribute{
		name: NewFunctionalDynamoDBAttribute(value),
	})
	
	return FunctionalDynamoDBItem{
		attributes: newAttributes,
	}
}

// GetAttribute returns an attribute if present
func (fdi FunctionalDynamoDBItem) GetAttribute(name string) mo.Option[FunctionalDynamoDBAttribute] {
	if attr, exists := fdi.attributes[name]; exists {
		return mo.Some(attr)
	}
	return mo.None[FunctionalDynamoDBAttribute]()
}

// GetAttributes returns all attributes
func (fdi FunctionalDynamoDBItem) GetAttributes() map[string]FunctionalDynamoDBAttribute {
	return fdi.attributes
}

// HasAttribute checks if an attribute exists
func (fdi FunctionalDynamoDBItem) HasAttribute(name string) bool {
	_, exists := fdi.attributes[name]
	return exists
}

// FunctionalDynamoDBOperationResult represents the result of a DynamoDB operation
type FunctionalDynamoDBOperationResult struct {
	result   mo.Option[interface{}]
	error    mo.Option[error]
	duration time.Duration
	metadata map[string]interface{}
}

// NewSuccessDynamoDBResult creates a successful operation result
func NewSuccessDynamoDBResult(result interface{}, duration time.Duration, metadata map[string]interface{}) FunctionalDynamoDBOperationResult {
	return FunctionalDynamoDBOperationResult{
		result:   mo.Some(result),
		error:    mo.None[error](),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
	}
}

// NewErrorDynamoDBResult creates an error operation result
func NewErrorDynamoDBResult(err error, duration time.Duration, metadata map[string]interface{}) FunctionalDynamoDBOperationResult {
	return FunctionalDynamoDBOperationResult{
		result:   mo.None[interface{}](),
		error:    mo.Some(err),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
	}
}

// IsSuccess returns true if the operation was successful
func (fdor FunctionalDynamoDBOperationResult) IsSuccess() bool {
	return fdor.result.IsPresent() && fdor.error.IsAbsent()
}

// IsError returns true if the operation has an error
func (fdor FunctionalDynamoDBOperationResult) IsError() bool {
	return fdor.error.IsPresent()
}

// GetResult returns the operation result if present
func (fdor FunctionalDynamoDBOperationResult) GetResult() mo.Option[interface{}] {
	return fdor.result
}

// GetError returns the error if present
func (fdor FunctionalDynamoDBOperationResult) GetError() mo.Option[error] {
	return fdor.error
}

// GetDuration returns the execution duration
func (fdor FunctionalDynamoDBOperationResult) GetDuration() time.Duration {
	return fdor.duration
}

// GetMetadata returns the metadata
func (fdor FunctionalDynamoDBOperationResult) GetMetadata() map[string]interface{} {
	return fdor.metadata
}

// MapResult applies a function to the result if present
func (fdor FunctionalDynamoDBOperationResult) MapResult(f func(interface{}) interface{}) FunctionalDynamoDBOperationResult {
	if fdor.IsSuccess() {
		newResult := fdor.result.Map(func(result interface{}) (interface{}, bool) {
			return f(result), true
		})
		return FunctionalDynamoDBOperationResult{
			result:   newResult,
			error:    fdor.error,
			duration: fdor.duration,
			metadata: fdor.metadata,
		}
	}
	return fdor
}

// FunctionalDynamoDBValidation provides functional validation utilities
func FunctionalDynamoDBValidation() struct {
	ValidateTableName func(string) error
	ValidateItem      func(FunctionalDynamoDBItem) error
	ValidateKeySchema func(map[string]FunctionalDynamoDBAttribute) error
} {
	return struct {
		ValidateTableName func(string) error
		ValidateItem      func(FunctionalDynamoDBItem) error
		ValidateKeySchema func(map[string]FunctionalDynamoDBAttribute) error
	}{
		ValidateTableName: func(tableName string) error {
			if len(tableName) == 0 {
				return fmt.Errorf("table name cannot be empty")
			}
			if len(tableName) > 255 {
				return fmt.Errorf("table name cannot be longer than 255 characters")
			}
			return nil
		},
		ValidateItem: func(item FunctionalDynamoDBItem) error {
			if len(item.attributes) == 0 {
				return fmt.Errorf("item cannot be empty")
			}
			return nil
		},
		ValidateKeySchema: func(keySchema map[string]FunctionalDynamoDBAttribute) error {
			if len(keySchema) == 0 {
				return fmt.Errorf("key schema cannot be empty")
			}
			return nil
		},
	}
}

// FunctionalRetryWithBackoff retries an operation with exponential backoff
func FunctionalDynamoDBRetryWithBackoff[T any](operation func() (T, error), maxRetries int, delay time.Duration) FunctionalRetryResult[T] {
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := operation()
		if err == nil {
			return FunctionalRetryResult[T]{
				result: mo.Some(result),
				error:  mo.None[error](),
			}
		}
		
		// Wait before retry (except on last attempt)
		if attempt < maxRetries-1 {
			backoffDelay := time.Duration(attempt+1) * delay
			time.Sleep(backoffDelay)
		}
	}
	
	// Final attempt to capture the last error
	_, err := operation()
	return FunctionalRetryResult[T]{
		result: mo.None[T](),
		error:  mo.Some(err),
	}
}

// FunctionalRetryResult represents the result of a retry operation
type FunctionalRetryResult[T any] struct {
	result mo.Option[T]
	error  mo.Option[error]
}

// IsError returns true if the retry failed
func (frr FunctionalRetryResult[T]) IsError() bool {
	return frr.error.IsPresent()
}

// GetResult returns the result if present
func (frr FunctionalRetryResult[T]) GetResult() mo.Option[T] {
	return frr.result
}

// GetError returns the error if present
func (frr FunctionalRetryResult[T]) GetError() mo.Option[error] {
	return frr.error
}

// SimulateFunctionalDynamoDBPutItem simulates putting an item to DynamoDB
func SimulateFunctionalDynamoDBPutItem(item FunctionalDynamoDBItem, config FunctionalDynamoDBConfig) FunctionalDynamoDBOperationResult {
	startTime := time.Now()
	
	validator := FunctionalDynamoDBValidation()
	
	// Validate inputs using functional composition
	validationPipeline := ExecuteFunctionalValidationPipeline[FunctionalDynamoDBItem](item,
		func(i FunctionalDynamoDBItem) (FunctionalDynamoDBItem, error) {
			if err := validator.ValidateTableName(config.tableName); err != nil {
				return i, err
			}
			return i, nil
		},
		func(i FunctionalDynamoDBItem) (FunctionalDynamoDBItem, error) {
			if err := validator.ValidateItem(i); err != nil {
				return i, err
			}
			return i, nil
		},
	)
	
	if validationPipeline.IsError() {
		return NewErrorDynamoDBResult(
			validationPipeline.GetError().MustGet(),
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"table_name": config.tableName,
				"operation":  "put_item",
				"phase":      "validation",
			}),
		)
	}

	// Simulate put item operation with retry logic
	retryResult := FunctionalDynamoDBRetryWithBackoff(
		func() (map[string]interface{}, error) {
			// Simulate successful put item operation
			result := map[string]interface{}{
				"table_name":     config.tableName,
				"operation":      "put_item",
				"attributes":     len(item.attributes),
				"consumed_capacity": map[string]interface{}{
					"capacity_units": 1.0,
				},
			}
			
			return result, nil
		},
		config.maxRetries,
		config.retryDelay,
	)
	
	totalDuration := time.Since(startTime)
	
	if retryResult.IsError() {
		return NewErrorDynamoDBResult(
			retryResult.GetError().MustGet(),
			totalDuration,
			lo.Assign(config.metadata, map[string]interface{}{
				"table_name": config.tableName,
				"operation":  "put_item",
				"phase":      "execution",
				"retries":    config.maxRetries,
			}),
		)
	}
	
	return NewSuccessDynamoDBResult(
		retryResult.GetResult().MustGet(),
		totalDuration,
		lo.Assign(config.metadata, map[string]interface{}{
			"table_name": config.tableName,
			"operation":  "put_item",
			"phase":      "success",
		}),
	)
}

// SimulateFunctionalDynamoDBGetItem simulates getting an item from DynamoDB
func SimulateFunctionalDynamoDBGetItem(key map[string]FunctionalDynamoDBAttribute, config FunctionalDynamoDBConfig) FunctionalDynamoDBOperationResult {
	startTime := time.Now()
	
	validator := FunctionalDynamoDBValidation()
	
	// Validate inputs
	if err := validator.ValidateTableName(config.tableName); err != nil {
		return NewErrorDynamoDBResult(
			err,
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"table_name": config.tableName,
				"operation":  "get_item",
				"phase":      "validation",
			}),
		)
	}
	
	if err := validator.ValidateKeySchema(key); err != nil {
		return NewErrorDynamoDBResult(
			err,
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"table_name": config.tableName,
				"operation":  "get_item",
				"phase":      "validation",
			}),
		)
	}

	// Simulate get item operation with retry logic
	retryResult := FunctionalDynamoDBRetryWithBackoff(
		func() (FunctionalDynamoDBItem, error) {
			// Simulate successful get item operation
			result := NewFunctionalDynamoDBItem().
				WithAttribute("id", "test-id-123").
				WithAttribute("name", "Test Item").
				WithAttribute("active", true).
				WithAttribute("score", 95)
			
			return result, nil
		},
		config.maxRetries,
		config.retryDelay,
	)
	
	totalDuration := time.Since(startTime)
	
	if retryResult.IsError() {
		return NewErrorDynamoDBResult(
			retryResult.GetError().MustGet(),
			totalDuration,
			lo.Assign(config.metadata, map[string]interface{}{
				"table_name": config.tableName,
				"operation":  "get_item",
				"phase":      "execution",
				"retries":    config.maxRetries,
			}),
		)
	}
	
	return NewSuccessDynamoDBResult(
		retryResult.GetResult().MustGet(),
		totalDuration,
		lo.Assign(config.metadata, map[string]interface{}{
			"table_name": config.tableName,
			"operation":  "get_item",
			"phase":      "success",
		}),
	)
}

// SimulateFunctionalDynamoDBQuery simulates querying DynamoDB
func SimulateFunctionalDynamoDBQuery(keyCondition map[string]FunctionalDynamoDBAttribute, config FunctionalDynamoDBConfig) FunctionalDynamoDBOperationResult {
	startTime := time.Now()
	
	validator := FunctionalDynamoDBValidation()
	
	// Validate inputs
	if err := validator.ValidateTableName(config.tableName); err != nil {
		return NewErrorDynamoDBResult(
			err,
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"table_name": config.tableName,
				"operation":  "query",
				"phase":      "validation",
			}),
		)
	}

	// Simulate query operation with retry logic
	retryResult := FunctionalDynamoDBRetryWithBackoff(
		func() ([]FunctionalDynamoDBItem, error) {
			// Simulate successful query operation returning multiple items
			items := []FunctionalDynamoDBItem{
				NewFunctionalDynamoDBItem().
					WithAttribute("id", "item-1").
					WithAttribute("name", "First Item").
					WithAttribute("active", true),
				NewFunctionalDynamoDBItem().
					WithAttribute("id", "item-2").
					WithAttribute("name", "Second Item").
					WithAttribute("active", false),
				NewFunctionalDynamoDBItem().
					WithAttribute("id", "item-3").
					WithAttribute("name", "Third Item").
					WithAttribute("active", true),
			}
			
			return items, nil
		},
		config.maxRetries,
		config.retryDelay,
	)
	
	totalDuration := time.Since(startTime)
	
	if retryResult.IsError() {
		return NewErrorDynamoDBResult(
			retryResult.GetError().MustGet(),
			totalDuration,
			lo.Assign(config.metadata, map[string]interface{}{
				"table_name": config.tableName,
				"operation":  "query",
				"phase":      "execution",
				"retries":    config.maxRetries,
			}),
		)
	}
	
	return NewSuccessDynamoDBResult(
		retryResult.GetResult().MustGet(),
		totalDuration,
		lo.Assign(config.metadata, map[string]interface{}{
			"table_name": config.tableName,
			"operation":  "query",
			"phase":      "success",
			"item_count": len(retryResult.GetResult().MustGet()),
		}),
	)
}

// FunctionalValidationPipeline represents a validation pipeline for DynamoDB
type FunctionalValidationPipeline[T any] struct {
	value T
	error mo.Option[error]
}

// ExecuteFunctionalValidationPipeline executes a series of validation functions
func ExecuteFunctionalValidationPipeline[T any](input T, operations ...func(T) (T, error)) FunctionalValidationPipeline[T] {
	return lo.Reduce(operations, func(acc FunctionalValidationPipeline[T], op func(T) (T, error), _ int) FunctionalValidationPipeline[T] {
		if acc.error.IsPresent() {
			return acc
		}
		
		result, err := op(acc.value)
		if err != nil {
			return FunctionalValidationPipeline[T]{
				value: acc.value,
				error: mo.Some(err),
			}
		}
		
		return FunctionalValidationPipeline[T]{
			value: result,
			error: mo.None[error](),
		}
	}, FunctionalValidationPipeline[T]{
		value: input,
		error: mo.None[error](),
	})
}

// IsError returns true if the pipeline has an error
func (fvp FunctionalValidationPipeline[T]) IsError() bool {
	return fvp.error.IsPresent()
}

// GetError returns the error if present
func (fvp FunctionalValidationPipeline[T]) GetError() mo.Option[error] {
	return fvp.error
}

// GetValue returns the value
func (fvp FunctionalValidationPipeline[T]) GetValue() T {
	return fvp.value
}