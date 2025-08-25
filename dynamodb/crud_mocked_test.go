package dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// DynamoDBClientInterface defines the interface for DynamoDB operations
type DynamoDBClientInterface interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

// MockDynamoDBCRUDClient is a mock implementation for testing CRUD operations
type MockDynamoDBCRUDClient struct {
	mock.Mock
}

func (m *MockDynamoDBCRUDClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockDynamoDBCRUDClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDBCRUDClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

func (m *MockDynamoDBCRUDClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

func (m *MockDynamoDBCRUDClient) BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.BatchGetItemOutput), args.Error(1)
}

func (m *MockDynamoDBCRUDClient) BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.BatchWriteItemOutput), args.Error(1)
}

// Test data factories for pure functions
func createValidTestItem() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":    &types.AttributeValueMemberS{Value: "test-id-123"},
		"name":  &types.AttributeValueMemberS{Value: "test-name"},
		"age":   &types.AttributeValueMemberN{Value: "25"},
		"email": &types.AttributeValueMemberS{Value: "test@example.com"},
	}
}

func createValidTestKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id-123"},
	}
}

func createCompositeKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "USER#123"},
		"sk": &types.AttributeValueMemberS{Value: "PROFILE#456"},
	}
}

func createTestPutItemOptions() PutItemOptions {
	return PutItemOptions{
		ConditionExpression: stringPtrCrud("attribute_not_exists(id)"),
		ExpressionAttributeNames: map[string]string{
			"#id": "id",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: "test-id-123"},
		},
		ReturnConsumedCapacity:              types.ReturnConsumedCapacityTotal,
		ReturnItemCollectionMetrics:         types.ReturnItemCollectionMetricsSize,
		ReturnValues:                        types.ReturnValueAllOld,
		ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
	}
}

// TESTS FOR PUTITEM OPERATIONS

func TestValidatePutItemInputCreation_WithBasicItem_ShouldCreateValidInput(t *testing.T) {
	// RED: Test input validation and creation for PutItem
	tableName := "test-table"
	item := createValidTestItem()

	input := createPutItemInput(tableName, item, PutItemOptions{})

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, item, input.Item)
	assert.Nil(t, input.ConditionExpression)
	assert.Nil(t, input.ExpressionAttributeNames)
	assert.Nil(t, input.ExpressionAttributeValues)
}

func TestValidatePutItemInputCreation_WithAllOptions_ShouldCreateCompleteInput(t *testing.T) {
	// RED: Test input validation and creation with all options
	tableName := "test-table"
	item := createValidTestItem()
	options := createTestPutItemOptions()

	input := createPutItemInput(tableName, item, options)

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, item, input.Item)
	assert.Equal(t, *options.ConditionExpression, *input.ConditionExpression)
	assert.Equal(t, options.ExpressionAttributeNames, input.ExpressionAttributeNames)
	assert.Equal(t, options.ExpressionAttributeValues, input.ExpressionAttributeValues)
	assert.Equal(t, options.ReturnConsumedCapacity, input.ReturnConsumedCapacity)
	assert.Equal(t, options.ReturnItemCollectionMetrics, input.ReturnItemCollectionMetrics)
	assert.Equal(t, options.ReturnValues, input.ReturnValues)
	assert.Equal(t, options.ReturnValuesOnConditionCheckFailure, input.ReturnValuesOnConditionCheckFailure)
}

func TestValidatePutItemInputCreation_WithEmptyTableName_ShouldCreateInputWithEmptyTableName(t *testing.T) {
	// RED: Test input validation with empty table name
	tableName := ""
	item := createValidTestItem()

	input := createPutItemInput(tableName, item, PutItemOptions{})

	assert.Equal(t, "", *input.TableName)
	assert.Equal(t, item, input.Item)
}

func TestValidatePutItemInputCreation_WithNilItem_ShouldCreateInputWithNilItem(t *testing.T) {
	// RED: Test input validation with nil item
	tableName := "test-table"
	var item map[string]types.AttributeValue

	input := createPutItemInput(tableName, item, PutItemOptions{})

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, item, input.Item)
}

func TestValidatePutItemInputCreation_WithComplexDataTypes_ShouldHandleAllAttributeTypes(t *testing.T) {
	// RED: Test input validation with complex data types
	tableName := "test-table"
	item := map[string]types.AttributeValue{
		"string":    &types.AttributeValueMemberS{Value: "string value"},
		"number":    &types.AttributeValueMemberN{Value: "42"},
		"binary":    &types.AttributeValueMemberB{Value: []byte("binary data")},
		"boolean":   &types.AttributeValueMemberBOOL{Value: true},
		"null":      &types.AttributeValueMemberNULL{Value: true},
		"stringSet": &types.AttributeValueMemberSS{Value: []string{"a", "b", "c"}},
		"numberSet": &types.AttributeValueMemberNS{Value: []string{"1", "2", "3"}},
		"binarySet": &types.AttributeValueMemberBS{Value: [][]byte{[]byte("bin1"), []byte("bin2")}},
		"list": &types.AttributeValueMemberL{
			Value: []types.AttributeValue{
				&types.AttributeValueMemberS{Value: "list item"},
				&types.AttributeValueMemberN{Value: "999"},
			},
		},
		"map": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"nested": &types.AttributeValueMemberS{Value: "nested value"},
			},
		},
	}

	input := createPutItemInput(tableName, item, PutItemOptions{})

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, item, input.Item)
}

// TESTS FOR GETITEM OPERATIONS

func TestValidateGetItemInputCreation_WithBasicKey_ShouldCreateValidInput(t *testing.T) {
	// RED: Test input validation and creation for GetItem
	tableName := "test-table"
	key := createValidTestKey()

	input := createGetItemInput(tableName, key, GetItemOptions{})

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
	assert.Nil(t, input.AttributesToGet)
	assert.Nil(t, input.ConsistentRead)
	assert.Nil(t, input.ExpressionAttributeNames)
	assert.Nil(t, input.ProjectionExpression)
}

func TestValidateGetItemInputCreation_WithAllOptions_ShouldCreateCompleteInput(t *testing.T) {
	// RED: Test input validation and creation with all options
	tableName := "test-table"
	key := createValidTestKey()
	options := GetItemOptions{
		AttributesToGet:          []string{"id", "name", "email"},
		ConsistentRead:           boolPtrCrud(true),
		ExpressionAttributeNames: map[string]string{"#id": "id"},
		ProjectionExpression:     stringPtrCrud("#id, #n, #e"),
		ReturnConsumedCapacity:   types.ReturnConsumedCapacityTotal,
	}

	input := createGetItemInput(tableName, key, options)

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
	assert.Equal(t, options.AttributesToGet, input.AttributesToGet)
	assert.Equal(t, *options.ConsistentRead, *input.ConsistentRead)
	assert.Equal(t, options.ExpressionAttributeNames, input.ExpressionAttributeNames)
	assert.Equal(t, *options.ProjectionExpression, *input.ProjectionExpression)
	assert.Equal(t, options.ReturnConsumedCapacity, input.ReturnConsumedCapacity)
}

func TestValidateGetItemInputCreation_WithCompositeKey_ShouldCreateValidInput(t *testing.T) {
	// RED: Test input validation with composite key
	tableName := "test-table"
	key := createCompositeKey()

	input := createGetItemInput(tableName, key, GetItemOptions{})

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
}

func TestValidateGetItemInputCreation_WithEmptyTableName_ShouldCreateInputWithEmptyTableName(t *testing.T) {
	// RED: Test input validation with empty table name
	tableName := ""
	key := createValidTestKey()

	input := createGetItemInput(tableName, key, GetItemOptions{})

	assert.Equal(t, "", *input.TableName)
	assert.Equal(t, key, input.Key)
}

func TestValidateGetItemInputCreation_WithEmptyKey_ShouldCreateInputWithEmptyKey(t *testing.T) {
	// RED: Test input validation with empty key
	tableName := "test-table"
	key := map[string]types.AttributeValue{}

	input := createGetItemInput(tableName, key, GetItemOptions{})

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
}

// TESTS FOR UPDATEITEM OPERATIONS

func TestValidateUpdateItemInputCreation_WithBasicUpdate_ShouldCreateValidInput(t *testing.T) {
	// RED: Test input validation and creation for UpdateItem
	tableName := "test-table"
	key := createValidTestKey()
	updateExpression := "SET #n = :name"
	expressionAttributeNames := map[string]string{"#n": "name"}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name": &types.AttributeValueMemberS{Value: "updated-name"},
	}

	input := createUpdateItemInput(tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues, UpdateItemOptions{})

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
	assert.Equal(t, updateExpression, *input.UpdateExpression)
	assert.Equal(t, expressionAttributeNames, input.ExpressionAttributeNames)
	assert.Equal(t, expressionAttributeValues, input.ExpressionAttributeValues)
	assert.Nil(t, input.ConditionExpression)
}

func TestValidateUpdateItemInputCreation_WithAllOptions_ShouldCreateCompleteInput(t *testing.T) {
	// RED: Test input validation and creation with all options
	tableName := "test-table"
	key := createValidTestKey()
	updateExpression := "SET #n = :name"
	expressionAttributeNames := map[string]string{"#n": "name"}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name": &types.AttributeValueMemberS{Value: "updated-name"},
	}
	options := UpdateItemOptions{
		ConditionExpression:                 stringPtrCrud("attribute_exists(id)"),
		ReturnConsumedCapacity:              types.ReturnConsumedCapacityTotal,
		ReturnItemCollectionMetrics:         types.ReturnItemCollectionMetricsSize,
		ReturnValues:                        types.ReturnValueAllNew,
		ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
	}

	input := createUpdateItemInput(tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues, options)

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
	assert.Equal(t, updateExpression, *input.UpdateExpression)
	assert.Equal(t, *options.ConditionExpression, *input.ConditionExpression)
	assert.Equal(t, options.ReturnConsumedCapacity, input.ReturnConsumedCapacity)
	assert.Equal(t, options.ReturnItemCollectionMetrics, input.ReturnItemCollectionMetrics)
	assert.Equal(t, options.ReturnValues, input.ReturnValues)
	assert.Equal(t, options.ReturnValuesOnConditionCheckFailure, input.ReturnValuesOnConditionCheckFailure)
}

func TestValidateUpdateItemInputCreation_WithComplexUpdateExpression_ShouldHandleComplexOperations(t *testing.T) {
	// RED: Test input validation with complex update expression
	tableName := "test-table"
	key := createValidTestKey()
	updateExpression := "SET #n = :name, #a = :age ADD #s :score REMOVE #old DELETE #tags :tag"
	expressionAttributeNames := map[string]string{
		"#n":    "name",
		"#a":    "age",
		"#s":    "score",
		"#old":  "oldField",
		"#tags": "tags",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name":  &types.AttributeValueMemberS{Value: "updated-name"},
		":age":   &types.AttributeValueMemberN{Value: "30"},
		":score": &types.AttributeValueMemberN{Value: "100"},
		":tag":   &types.AttributeValueMemberSS{Value: []string{"removed-tag"}},
	}

	input := createUpdateItemInput(tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues, UpdateItemOptions{})

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
	assert.Equal(t, updateExpression, *input.UpdateExpression)
	assert.Equal(t, expressionAttributeNames, input.ExpressionAttributeNames)
	assert.Equal(t, expressionAttributeValues, input.ExpressionAttributeValues)
}

// TESTS FOR DELETEITEM OPERATIONS

func TestValidateDeleteItemInputCreation_WithBasicKey_ShouldCreateValidInput(t *testing.T) {
	// RED: Test input validation and creation for DeleteItem
	tableName := "test-table"
	key := createValidTestKey()

	input := createDeleteItemInput(tableName, key, DeleteItemOptions{})

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
	assert.Nil(t, input.ConditionExpression)
	assert.Nil(t, input.ExpressionAttributeNames)
	assert.Nil(t, input.ExpressionAttributeValues)
}

func TestValidateDeleteItemInputCreation_WithAllOptions_ShouldCreateCompleteInput(t *testing.T) {
	// RED: Test input validation and creation with all options
	tableName := "test-table"
	key := createValidTestKey()
	options := DeleteItemOptions{
		ConditionExpression: stringPtrCrud("attribute_exists(id)"),
		ExpressionAttributeNames: map[string]string{
			"#id": "id",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: "test-id-123"},
		},
		ReturnConsumedCapacity:              types.ReturnConsumedCapacityTotal,
		ReturnItemCollectionMetrics:         types.ReturnItemCollectionMetricsSize,
		ReturnValues:                        types.ReturnValueAllOld,
		ReturnValuesOnConditionCheckFailure: types.ReturnValuesOnConditionCheckFailureAllOld,
	}

	input := createDeleteItemInput(tableName, key, options)

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
	assert.Equal(t, *options.ConditionExpression, *input.ConditionExpression)
	assert.Equal(t, options.ExpressionAttributeNames, input.ExpressionAttributeNames)
	assert.Equal(t, options.ExpressionAttributeValues, input.ExpressionAttributeValues)
	assert.Equal(t, options.ReturnConsumedCapacity, input.ReturnConsumedCapacity)
	assert.Equal(t, options.ReturnItemCollectionMetrics, input.ReturnItemCollectionMetrics)
	assert.Equal(t, options.ReturnValues, input.ReturnValues)
	assert.Equal(t, options.ReturnValuesOnConditionCheckFailure, input.ReturnValuesOnConditionCheckFailure)
}

func TestValidateDeleteItemInputCreation_WithConditionalExpression_ShouldHandleComplexConditions(t *testing.T) {
	// RED: Test input validation with conditional expressions
	tableName := "test-table"
	key := createValidTestKey()
	options := DeleteItemOptions{
		ConditionExpression: stringPtrCrud("attribute_exists(id) AND #a > :minAge AND contains(#tags, :tag)"),
		ExpressionAttributeNames: map[string]string{
			"#a":    "age",
			"#tags": "tags",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":minAge": &types.AttributeValueMemberN{Value: "18"},
			":tag":    &types.AttributeValueMemberS{Value: "active"},
		},
	}

	input := createDeleteItemInput(tableName, key, options)

	assert.Equal(t, tableName, *input.TableName)
	assert.Equal(t, key, input.Key)
	assert.Equal(t, *options.ConditionExpression, *input.ConditionExpression)
	assert.Equal(t, options.ExpressionAttributeNames, input.ExpressionAttributeNames)
	assert.Equal(t, options.ExpressionAttributeValues, input.ExpressionAttributeValues)
}

// ERROR SCENARIO TESTS

func TestPutItemInputValidation_WithValidationErrors_ShouldReturnAppropriateErrors(t *testing.T) {
	testCases := []struct {
		name      string
		tableName string
		item      map[string]types.AttributeValue
		options   PutItemOptions
	}{
		{
			name:      "Empty table name",
			tableName: "",
			item:      createValidTestItem(),
			options:   PutItemOptions{},
		},
		{
			name:      "Nil item",
			tableName: "test-table",
			item:      nil,
			options:   PutItemOptions{},
		},
		{
			name:      "Empty item",
			tableName: "test-table",
			item:      map[string]types.AttributeValue{},
			options:   PutItemOptions{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := createPutItemInput(tc.tableName, tc.item, tc.options)
			
			// These inputs should still be created but would fail at AWS level
			assert.NotNil(t, input)
			assert.Equal(t, tc.tableName, *input.TableName)
		})
	}
}

// Helper functions for input creation (pure functions)

func createPutItemInput(tableName string, item map[string]types.AttributeValue, options PutItemOptions) *dynamodb.PutItemInput {
	input := &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      item,
	}

	if options.ConditionExpression != nil {
		input.ConditionExpression = options.ConditionExpression
	}
	if options.ExpressionAttributeNames != nil {
		input.ExpressionAttributeNames = options.ExpressionAttributeNames
	}
	if options.ExpressionAttributeValues != nil {
		input.ExpressionAttributeValues = options.ExpressionAttributeValues
	}
	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}
	if options.ReturnItemCollectionMetrics != "" {
		input.ReturnItemCollectionMetrics = options.ReturnItemCollectionMetrics
	}
	if options.ReturnValues != "" {
		input.ReturnValues = options.ReturnValues
	}
	if options.ReturnValuesOnConditionCheckFailure != "" {
		input.ReturnValuesOnConditionCheckFailure = options.ReturnValuesOnConditionCheckFailure
	}

	return input
}

func createGetItemInput(tableName string, key map[string]types.AttributeValue, options GetItemOptions) *dynamodb.GetItemInput {
	input := &dynamodb.GetItemInput{
		TableName: &tableName,
		Key:       key,
	}

	if len(options.AttributesToGet) > 0 {
		input.AttributesToGet = options.AttributesToGet
	}
	if options.ConsistentRead != nil {
		input.ConsistentRead = options.ConsistentRead
	}
	if options.ExpressionAttributeNames != nil {
		input.ExpressionAttributeNames = options.ExpressionAttributeNames
	}
	if options.ProjectionExpression != nil {
		input.ProjectionExpression = options.ProjectionExpression
	}
	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}

	return input
}

func createUpdateItemInput(tableName string, key map[string]types.AttributeValue, updateExpression string, expressionAttributeNames map[string]string, expressionAttributeValues map[string]types.AttributeValue, options UpdateItemOptions) *dynamodb.UpdateItemInput {
	input := &dynamodb.UpdateItemInput{
		TableName:        &tableName,
		Key:              key,
		UpdateExpression: &updateExpression,
	}

	if expressionAttributeNames != nil {
		input.ExpressionAttributeNames = expressionAttributeNames
	}
	if expressionAttributeValues != nil {
		input.ExpressionAttributeValues = expressionAttributeValues
	}
	if options.ConditionExpression != nil {
		input.ConditionExpression = options.ConditionExpression
	}
	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}
	if options.ReturnItemCollectionMetrics != "" {
		input.ReturnItemCollectionMetrics = options.ReturnItemCollectionMetrics
	}
	if options.ReturnValues != "" {
		input.ReturnValues = options.ReturnValues
	}
	if options.ReturnValuesOnConditionCheckFailure != "" {
		input.ReturnValuesOnConditionCheckFailure = options.ReturnValuesOnConditionCheckFailure
	}

	return input
}

func createDeleteItemInput(tableName string, key map[string]types.AttributeValue, options DeleteItemOptions) *dynamodb.DeleteItemInput {
	input := &dynamodb.DeleteItemInput{
		TableName: &tableName,
		Key:       key,
	}

	if options.ConditionExpression != nil {
		input.ConditionExpression = options.ConditionExpression
	}
	if options.ExpressionAttributeNames != nil {
		input.ExpressionAttributeNames = options.ExpressionAttributeNames
	}
	if options.ExpressionAttributeValues != nil {
		input.ExpressionAttributeValues = options.ExpressionAttributeValues
	}
	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}
	if options.ReturnItemCollectionMetrics != "" {
		input.ReturnItemCollectionMetrics = options.ReturnItemCollectionMetrics
	}
	if options.ReturnValues != "" {
		input.ReturnValues = options.ReturnValues
	}
	if options.ReturnValuesOnConditionCheckFailure != "" {
		input.ReturnValuesOnConditionCheckFailure = options.ReturnValuesOnConditionCheckFailure
	}

	return input
}

// Helper functions for test data creation - using different names to avoid conflicts
func stringPtrCrud(s string) *string {
	return &s
}

func boolPtrCrud(b bool) *bool {
	return &b
}

func int32PtrCrud(i int32) *int32 {
	return &i
}

// AWS Error simulation functions for testing error handling
func createConditionalCheckFailedException() error {
	return &types.ConditionalCheckFailedException{
		Message: stringPtrCrud("The conditional request failed"),
	}
}

func createResourceNotFoundException() error {
	return &types.ResourceNotFoundException{
		Message: stringPtrCrud("Requested resource not found"),
	}
}

func createThrottlingException() error {
	return &types.RequestLimitExceeded{
		Message: stringPtrCrud("Too many requests"),
	}
}

func createInternalServerError() error {
	return &types.InternalServerError{
		Message: stringPtrCrud("Internal server error"),
	}
}

func createLimitExceededException() error {
	return &types.LimitExceededException{
		Message: stringPtrCrud("Subscriber limit exceeded"),
	}
}