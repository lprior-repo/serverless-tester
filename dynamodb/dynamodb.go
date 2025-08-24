package dynamodb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
)

// PutItemOptions contains optional parameters for PutItem operations
type PutItemOptions struct {
	ConditionExpression                 *string
	ExpressionAttributeNames            map[string]string
	ExpressionAttributeValues           map[string]types.AttributeValue
	ReturnConsumedCapacity              types.ReturnConsumedCapacity
	ReturnItemCollectionMetrics         types.ReturnItemCollectionMetrics
	ReturnValues                        types.ReturnValue
	ReturnValuesOnConditionCheckFailure types.ReturnValuesOnConditionCheckFailure
}

// GetItemOptions contains optional parameters for GetItem operations
type GetItemOptions struct {
	AttributesToGet              []string
	ConsistentRead               *bool
	ExpressionAttributeNames     map[string]string
	ProjectionExpression         *string
	ReturnConsumedCapacity       types.ReturnConsumedCapacity
}

// DeleteItemOptions contains optional parameters for DeleteItem operations
type DeleteItemOptions struct {
	ConditionExpression                 *string
	ExpressionAttributeNames            map[string]string
	ExpressionAttributeValues           map[string]types.AttributeValue
	ReturnConsumedCapacity              types.ReturnConsumedCapacity
	ReturnItemCollectionMetrics         types.ReturnItemCollectionMetrics
	ReturnValues                        types.ReturnValue
	ReturnValuesOnConditionCheckFailure types.ReturnValuesOnConditionCheckFailure
}

// UpdateItemOptions contains optional parameters for UpdateItem operations
type UpdateItemOptions struct {
	ConditionExpression                 *string
	ReturnConsumedCapacity              types.ReturnConsumedCapacity
	ReturnItemCollectionMetrics         types.ReturnItemCollectionMetrics
	ReturnValues                        types.ReturnValue
	ReturnValuesOnConditionCheckFailure types.ReturnValuesOnConditionCheckFailure
}

// QueryOptions contains optional parameters for Query operations
type QueryOptions struct {
	AttributesToGet              []string
	ConsistentRead               *bool
	ExclusiveStartKey            map[string]types.AttributeValue
	ExpressionAttributeNames     map[string]string
	FilterExpression             *string
	IndexName                    *string
	KeyConditionExpression       *string
	Limit                        *int32
	ProjectionExpression         *string
	ReturnConsumedCapacity       types.ReturnConsumedCapacity
	ScanIndexForward             *bool
	Select                       types.Select
}

// ScanOptions contains optional parameters for Scan operations
type ScanOptions struct {
	AttributesToGet              []string
	ConsistentRead               *bool
	ExclusiveStartKey            map[string]types.AttributeValue
	ExpressionAttributeNames     map[string]string
	ExpressionAttributeValues    map[string]types.AttributeValue
	FilterExpression             *string
	IndexName                    *string
	Limit                        *int32
	ProjectionExpression         *string
	ReturnConsumedCapacity       types.ReturnConsumedCapacity
	Segment                      *int32
	Select                       types.Select
	TotalSegments                *int32
}

// QueryResult represents the result of a Query operation
type QueryResult struct {
	Items            []map[string]types.AttributeValue
	Count            int32
	ScannedCount     int32
	LastEvaluatedKey map[string]types.AttributeValue
	ConsumedCapacity *types.ConsumedCapacity
}

// ScanResult represents the result of a Scan operation
type ScanResult struct {
	Items            []map[string]types.AttributeValue
	Count            int32
	ScannedCount     int32
	LastEvaluatedKey map[string]types.AttributeValue
	ConsumedCapacity *types.ConsumedCapacity
}

// BatchGetItemOptions contains optional parameters for BatchGetItem operations
type BatchGetItemOptions struct {
	ConsistentRead           *bool
	ExpressionAttributeNames map[string]string
	ProjectionExpression     *string
	ReturnConsumedCapacity   types.ReturnConsumedCapacity
}

// BatchWriteItemOptions contains optional parameters for BatchWriteItem operations
type BatchWriteItemOptions struct {
	ReturnConsumedCapacity      types.ReturnConsumedCapacity
	ReturnItemCollectionMetrics types.ReturnItemCollectionMetrics
}

// BatchGetItemResult represents the result of a BatchGetItem operation
type BatchGetItemResult struct {
	Items               []map[string]types.AttributeValue
	UnprocessedKeys     map[string]types.KeysAndAttributes
	ConsumedCapacity    []types.ConsumedCapacity
}

// BatchWriteItemResult represents the result of a BatchWriteItem operation
type BatchWriteItemResult struct {
	UnprocessedItems              map[string][]types.WriteRequest
	ConsumedCapacity             []types.ConsumedCapacity
	ItemCollectionMetrics        map[string][]types.ItemCollectionMetrics
}

// TransactWriteItemsOptions contains optional parameters for TransactWriteItems operations
type TransactWriteItemsOptions struct {
	ClientRequestToken          *string
	ReturnConsumedCapacity      types.ReturnConsumedCapacity
	ReturnItemCollectionMetrics types.ReturnItemCollectionMetrics
}

// TransactGetItemsOptions contains optional parameters for TransactGetItems operations
type TransactGetItemsOptions struct {
	ReturnConsumedCapacity types.ReturnConsumedCapacity
}

// TransactWriteItemsResult represents the result of a TransactWriteItems operation
type TransactWriteItemsResult struct {
	ConsumedCapacity      []types.ConsumedCapacity
	ItemCollectionMetrics map[string][]types.ItemCollectionMetrics
}

// TransactGetItemsResult represents the result of a TransactGetItems operation
type TransactGetItemsResult struct {
	Responses         []types.ItemResponse
	ConsumedCapacity  []types.ConsumedCapacity
}

// createDynamoDBClient creates a new DynamoDB client with default configuration
func createDynamoDBClient() (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// PutItem puts an item in the specified DynamoDB table and fails the test if an error occurs
func PutItem(t *testing.T, tableName string, item map[string]types.AttributeValue) {
	err := PutItemE(t, tableName, item)
	require.NoError(t, err)
}

// PutItemE puts an item in the specified DynamoDB table and returns any error
func PutItemE(t *testing.T, tableName string, item map[string]types.AttributeValue) error {
	return PutItemWithOptionsE(t, tableName, item, PutItemOptions{})
}

// PutItemWithOptions puts an item in the specified DynamoDB table with options and fails the test if an error occurs
func PutItemWithOptions(t *testing.T, tableName string, item map[string]types.AttributeValue, options PutItemOptions) {
	err := PutItemWithOptionsE(t, tableName, item, options)
	require.NoError(t, err)
}

// PutItemWithOptionsE puts an item in the specified DynamoDB table with options and returns any error
func PutItemWithOptionsE(t *testing.T, tableName string, item map[string]types.AttributeValue, options PutItemOptions) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

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

	_, err = client.PutItem(context.TODO(), input)
	return err
}

// GetItem retrieves an item from the specified DynamoDB table and fails the test if an error occurs
func GetItem(t *testing.T, tableName string, key map[string]types.AttributeValue) map[string]types.AttributeValue {
	result, err := GetItemE(t, tableName, key)
	require.NoError(t, err)
	return result
}

// GetItemE retrieves an item from the specified DynamoDB table and returns the item and any error
func GetItemE(t *testing.T, tableName string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	return GetItemWithOptionsE(t, tableName, key, GetItemOptions{})
}

// GetItemWithOptions retrieves an item from the specified DynamoDB table with options and fails the test if an error occurs
func GetItemWithOptions(t *testing.T, tableName string, key map[string]types.AttributeValue, options GetItemOptions) map[string]types.AttributeValue {
	result, err := GetItemWithOptionsE(t, tableName, key, options)
	require.NoError(t, err)
	return result
}

// GetItemWithOptionsE retrieves an item from the specified DynamoDB table with options and returns the item and any error
func GetItemWithOptionsE(t *testing.T, tableName string, key map[string]types.AttributeValue, options GetItemOptions) (map[string]types.AttributeValue, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

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

	result, err := client.GetItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	
	return result.Item, nil
}

// DeleteItem deletes an item from the specified DynamoDB table and fails the test if an error occurs
func DeleteItem(t *testing.T, tableName string, key map[string]types.AttributeValue) {
	err := DeleteItemE(t, tableName, key)
	require.NoError(t, err)
}

// DeleteItemE deletes an item from the specified DynamoDB table and returns any error
func DeleteItemE(t *testing.T, tableName string, key map[string]types.AttributeValue) error {
	return DeleteItemWithOptionsE(t, tableName, key, DeleteItemOptions{})
}

// DeleteItemWithOptions deletes an item from the specified DynamoDB table with options and fails the test if an error occurs
func DeleteItemWithOptions(t *testing.T, tableName string, key map[string]types.AttributeValue, options DeleteItemOptions) {
	err := DeleteItemWithOptionsE(t, tableName, key, options)
	require.NoError(t, err)
}

// DeleteItemWithOptionsE deletes an item from the specified DynamoDB table with options and returns any error
func DeleteItemWithOptionsE(t *testing.T, tableName string, key map[string]types.AttributeValue, options DeleteItemOptions) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

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

	_, err = client.DeleteItem(context.TODO(), input)
	return err
}

// UpdateItem updates an item in the specified DynamoDB table and fails the test if an error occurs
func UpdateItem(t *testing.T, tableName string, key map[string]types.AttributeValue, updateExpression string, expressionAttributeNames map[string]string, expressionAttributeValues map[string]types.AttributeValue) map[string]types.AttributeValue {
	result, err := UpdateItemE(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues)
	require.NoError(t, err)
	return result
}

// UpdateItemE updates an item in the specified DynamoDB table and returns the updated attributes and any error
func UpdateItemE(t *testing.T, tableName string, key map[string]types.AttributeValue, updateExpression string, expressionAttributeNames map[string]string, expressionAttributeValues map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	return UpdateItemWithOptionsE(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues, UpdateItemOptions{})
}

// UpdateItemWithOptions updates an item in the specified DynamoDB table with options and fails the test if an error occurs
func UpdateItemWithOptions(t *testing.T, tableName string, key map[string]types.AttributeValue, updateExpression string, expressionAttributeNames map[string]string, expressionAttributeValues map[string]types.AttributeValue, options UpdateItemOptions) map[string]types.AttributeValue {
	result, err := UpdateItemWithOptionsE(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues, options)
	require.NoError(t, err)
	return result
}

// UpdateItemWithOptionsE updates an item in the specified DynamoDB table with options and returns the updated attributes and any error
func UpdateItemWithOptionsE(t *testing.T, tableName string, key map[string]types.AttributeValue, updateExpression string, expressionAttributeNames map[string]string, expressionAttributeValues map[string]types.AttributeValue, options UpdateItemOptions) (map[string]types.AttributeValue, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

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

	result, err := client.UpdateItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return result.Attributes, nil
}

// Query performs a query operation on the specified DynamoDB table and fails the test if an error occurs
func Query(t *testing.T, tableName string, keyConditionExpression string, expressionAttributeValues map[string]types.AttributeValue) *QueryResult {
	result, err := QueryE(t, tableName, keyConditionExpression, expressionAttributeValues)
	require.NoError(t, err)
	return result
}

// QueryE performs a query operation on the specified DynamoDB table and returns the result and any error
func QueryE(t *testing.T, tableName string, keyConditionExpression string, expressionAttributeValues map[string]types.AttributeValue) (*QueryResult, error) {
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
	}
	return QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
}

// QueryWithOptions performs a query operation on the specified DynamoDB table with options and fails the test if an error occurs
func QueryWithOptions(t *testing.T, tableName string, keyConditionExpression string, expressionAttributeValues map[string]types.AttributeValue, options QueryOptions) *QueryResult {
	result, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
	require.NoError(t, err)
	return result
}

// QueryWithOptionsE performs a query operation on the specified DynamoDB table with options and returns the result and any error
func QueryWithOptionsE(t *testing.T, tableName string, expressionAttributeValues map[string]types.AttributeValue, options QueryOptions) (*QueryResult, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName: &tableName,
	}

	if options.KeyConditionExpression != nil {
		input.KeyConditionExpression = options.KeyConditionExpression
	}
	if expressionAttributeValues != nil {
		input.ExpressionAttributeValues = expressionAttributeValues
	}
	if len(options.AttributesToGet) > 0 {
		input.AttributesToGet = options.AttributesToGet
	}
	if options.ConsistentRead != nil {
		input.ConsistentRead = options.ConsistentRead
	}
	if options.ExclusiveStartKey != nil {
		input.ExclusiveStartKey = options.ExclusiveStartKey
	}
	if options.ExpressionAttributeNames != nil {
		input.ExpressionAttributeNames = options.ExpressionAttributeNames
	}
	if options.FilterExpression != nil {
		input.FilterExpression = options.FilterExpression
	}
	if options.IndexName != nil {
		input.IndexName = options.IndexName
	}
	if options.Limit != nil {
		input.Limit = options.Limit
	}
	if options.ProjectionExpression != nil {
		input.ProjectionExpression = options.ProjectionExpression
	}
	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}
	if options.ScanIndexForward != nil {
		input.ScanIndexForward = options.ScanIndexForward
	}
	if options.Select != "" {
		input.Select = options.Select
	}

	result, err := client.Query(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &QueryResult{
		Items:            result.Items,
		Count:            result.Count,
		ScannedCount:     result.ScannedCount,
		LastEvaluatedKey: result.LastEvaluatedKey,
		ConsumedCapacity: result.ConsumedCapacity,
	}, nil
}

// Scan performs a scan operation on the specified DynamoDB table and fails the test if an error occurs
func Scan(t *testing.T, tableName string) *ScanResult {
	result, err := ScanE(t, tableName)
	require.NoError(t, err)
	return result
}

// ScanE performs a scan operation on the specified DynamoDB table and returns the result and any error
func ScanE(t *testing.T, tableName string) (*ScanResult, error) {
	return ScanWithOptionsE(t, tableName, ScanOptions{})
}

// ScanWithOptions performs a scan operation on the specified DynamoDB table with options and fails the test if an error occurs
func ScanWithOptions(t *testing.T, tableName string, options ScanOptions) *ScanResult {
	result, err := ScanWithOptionsE(t, tableName, options)
	require.NoError(t, err)
	return result
}

// ScanWithOptionsE performs a scan operation on the specified DynamoDB table with options and returns the result and any error
func ScanWithOptionsE(t *testing.T, tableName string, options ScanOptions) (*ScanResult, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName: &tableName,
	}

	if len(options.AttributesToGet) > 0 {
		input.AttributesToGet = options.AttributesToGet
	}
	if options.ConsistentRead != nil {
		input.ConsistentRead = options.ConsistentRead
	}
	if options.ExclusiveStartKey != nil {
		input.ExclusiveStartKey = options.ExclusiveStartKey
	}
	if options.ExpressionAttributeNames != nil {
		input.ExpressionAttributeNames = options.ExpressionAttributeNames
	}
	if options.ExpressionAttributeValues != nil {
		input.ExpressionAttributeValues = options.ExpressionAttributeValues
	}
	if options.FilterExpression != nil {
		input.FilterExpression = options.FilterExpression
	}
	if options.IndexName != nil {
		input.IndexName = options.IndexName
	}
	if options.Limit != nil {
		input.Limit = options.Limit
	}
	if options.ProjectionExpression != nil {
		input.ProjectionExpression = options.ProjectionExpression
	}
	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}
	if options.Segment != nil {
		input.Segment = options.Segment
	}
	if options.Select != "" {
		input.Select = options.Select
	}
	if options.TotalSegments != nil {
		input.TotalSegments = options.TotalSegments
	}

	result, err := client.Scan(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &ScanResult{
		Items:            result.Items,
		Count:            result.Count,
		ScannedCount:     result.ScannedCount,
		LastEvaluatedKey: result.LastEvaluatedKey,
		ConsumedCapacity: result.ConsumedCapacity,
	}, nil
}

// QueryAllPages performs query operations with automatic pagination to retrieve all items
func QueryAllPages(t *testing.T, tableName string, keyConditionExpression string, expressionAttributeValues map[string]types.AttributeValue) []map[string]types.AttributeValue {
	items, err := QueryAllPagesE(t, tableName, keyConditionExpression, expressionAttributeValues)
	require.NoError(t, err)
	return items
}

// QueryAllPagesE performs query operations with automatic pagination to retrieve all items
func QueryAllPagesE(t *testing.T, tableName string, keyConditionExpression string, expressionAttributeValues map[string]types.AttributeValue) ([]map[string]types.AttributeValue, error) {
	return QueryAllPagesWithOptionsE(t, tableName, expressionAttributeValues, QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
	})
}

// QueryAllPagesWithOptionsE performs query operations with automatic pagination and options to retrieve all items
func QueryAllPagesWithOptionsE(t *testing.T, tableName string, expressionAttributeValues map[string]types.AttributeValue, options QueryOptions) ([]map[string]types.AttributeValue, error) {
	var allItems []map[string]types.AttributeValue
	var exclusiveStartKey map[string]types.AttributeValue

	for {
		currentOptions := options
		currentOptions.ExclusiveStartKey = exclusiveStartKey

		result, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, currentOptions)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, result.Items...)

		if result.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = result.LastEvaluatedKey
	}

	return allItems, nil
}

// ScanAllPages performs scan operations with automatic pagination to retrieve all items
func ScanAllPages(t *testing.T, tableName string) []map[string]types.AttributeValue {
	items, err := ScanAllPagesE(t, tableName)
	require.NoError(t, err)
	return items
}

// ScanAllPagesE performs scan operations with automatic pagination to retrieve all items
func ScanAllPagesE(t *testing.T, tableName string) ([]map[string]types.AttributeValue, error) {
	return ScanAllPagesWithOptionsE(t, tableName, ScanOptions{})
}

// ScanAllPagesWithOptionsE performs scan operations with automatic pagination and options to retrieve all items
func ScanAllPagesWithOptionsE(t *testing.T, tableName string, options ScanOptions) ([]map[string]types.AttributeValue, error) {
	var allItems []map[string]types.AttributeValue
	var exclusiveStartKey map[string]types.AttributeValue

	for {
		currentOptions := options
		currentOptions.ExclusiveStartKey = exclusiveStartKey

		result, err := ScanWithOptionsE(t, tableName, currentOptions)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, result.Items...)

		if result.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = result.LastEvaluatedKey
	}

	return allItems, nil
}

// BatchWriteItem writes multiple items to the specified DynamoDB table and fails the test if an error occurs
func BatchWriteItem(t *testing.T, tableName string, writeRequests []types.WriteRequest) *BatchWriteItemResult {
	result, err := BatchWriteItemE(t, tableName, writeRequests)
	require.NoError(t, err)
	return result
}

// BatchWriteItemE writes multiple items to the specified DynamoDB table and returns the result and any error
func BatchWriteItemE(t *testing.T, tableName string, writeRequests []types.WriteRequest) (*BatchWriteItemResult, error) {
	return BatchWriteItemWithOptionsE(t, tableName, writeRequests, BatchWriteItemOptions{})
}

// BatchWriteItemWithOptionsE writes multiple items to the specified DynamoDB table with options and returns the result and any error
func BatchWriteItemWithOptionsE(t *testing.T, tableName string, writeRequests []types.WriteRequest, options BatchWriteItemOptions) (*BatchWriteItemResult, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			tableName: writeRequests,
		},
	}

	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}
	if options.ReturnItemCollectionMetrics != "" {
		input.ReturnItemCollectionMetrics = options.ReturnItemCollectionMetrics
	}

	result, err := client.BatchWriteItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &BatchWriteItemResult{
		UnprocessedItems:      result.UnprocessedItems,
		ConsumedCapacity:      result.ConsumedCapacity,
		ItemCollectionMetrics: result.ItemCollectionMetrics,
	}, nil
}

// BatchGetItem retrieves multiple items from the specified DynamoDB table and fails the test if an error occurs
func BatchGetItem(t *testing.T, tableName string, keys []map[string]types.AttributeValue) *BatchGetItemResult {
	result, err := BatchGetItemE(t, tableName, keys)
	require.NoError(t, err)
	return result
}

// BatchGetItemE retrieves multiple items from the specified DynamoDB table and returns the result and any error
func BatchGetItemE(t *testing.T, tableName string, keys []map[string]types.AttributeValue) (*BatchGetItemResult, error) {
	return BatchGetItemWithOptionsE(t, tableName, keys, BatchGetItemOptions{})
}

// BatchGetItemWithOptions retrieves multiple items from the specified DynamoDB table with options and fails the test if an error occurs
func BatchGetItemWithOptions(t *testing.T, tableName string, keys []map[string]types.AttributeValue, options BatchGetItemOptions) *BatchGetItemResult {
	result, err := BatchGetItemWithOptionsE(t, tableName, keys, options)
	require.NoError(t, err)
	return result
}

// BatchGetItemWithOptionsE retrieves multiple items from the specified DynamoDB table with options and returns the result and any error
func BatchGetItemWithOptionsE(t *testing.T, tableName string, keys []map[string]types.AttributeValue, options BatchGetItemOptions) (*BatchGetItemResult, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	keysAndAttributes := types.KeysAndAttributes{
		Keys: keys,
	}

	if options.ConsistentRead != nil {
		keysAndAttributes.ConsistentRead = options.ConsistentRead
	}
	if options.ExpressionAttributeNames != nil {
		keysAndAttributes.ExpressionAttributeNames = options.ExpressionAttributeNames
	}
	if options.ProjectionExpression != nil {
		keysAndAttributes.ProjectionExpression = options.ProjectionExpression
	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			tableName: keysAndAttributes,
		},
	}

	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}

	result, err := client.BatchGetItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var items []map[string]types.AttributeValue
	if tableItems, exists := result.Responses[tableName]; exists {
		items = tableItems
	}

	return &BatchGetItemResult{
		Items:            items,
		UnprocessedKeys:  result.UnprocessedKeys,
		ConsumedCapacity: result.ConsumedCapacity,
	}, nil
}

// BatchWriteItemWithRetry writes multiple items with exponential backoff retry for unprocessed items
func BatchWriteItemWithRetry(t *testing.T, tableName string, writeRequests []types.WriteRequest, maxRetries int) {
	err := BatchWriteItemWithRetryE(t, tableName, writeRequests, maxRetries)
	require.NoError(t, err)
}

// BatchWriteItemWithRetryE writes multiple items with exponential backoff retry for unprocessed items
func BatchWriteItemWithRetryE(t *testing.T, tableName string, writeRequests []types.WriteRequest, maxRetries int) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	currentRequests := writeRequests
	for attempt := 0; attempt < maxRetries+1; attempt++ {
		if len(currentRequests) == 0 {
			break
		}

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				tableName: currentRequests,
			},
		}

		result, err := client.BatchWriteItem(context.TODO(), input)
		if err != nil {
			return err
		}

		if len(result.UnprocessedItems) == 0 {
			break
		}

		if attempt == maxRetries {
			return fmt.Errorf("failed to process all items after %d retries, %d items remaining", maxRetries, len(result.UnprocessedItems[tableName]))
		}

		// Exponential backoff: 100ms, 200ms, 400ms, 800ms, etc.
		backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
		time.Sleep(backoff)

		currentRequests = result.UnprocessedItems[tableName]
	}

	return nil
}

// BatchGetItemWithRetry retrieves multiple items with exponential backoff retry for unprocessed keys
func BatchGetItemWithRetry(t *testing.T, tableName string, keys []map[string]types.AttributeValue, maxRetries int) []map[string]types.AttributeValue {
	items, err := BatchGetItemWithRetryE(t, tableName, keys, maxRetries)
	require.NoError(t, err)
	return items
}

// BatchGetItemWithRetryE retrieves multiple items with exponential backoff retry for unprocessed keys
func BatchGetItemWithRetryE(t *testing.T, tableName string, keys []map[string]types.AttributeValue, maxRetries int) ([]map[string]types.AttributeValue, error) {
	return BatchGetItemWithRetryAndOptionsE(t, tableName, keys, maxRetries, BatchGetItemOptions{})
}

// BatchGetItemWithRetryAndOptionsE retrieves multiple items with exponential backoff retry and options for unprocessed keys
func BatchGetItemWithRetryAndOptionsE(t *testing.T, tableName string, keys []map[string]types.AttributeValue, maxRetries int, options BatchGetItemOptions) ([]map[string]types.AttributeValue, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	var allItems []map[string]types.AttributeValue
	currentKeys := keys

	for attempt := 0; attempt < maxRetries+1; attempt++ {
		if len(currentKeys) == 0 {
			break
		}

		keysAndAttributes := types.KeysAndAttributes{
			Keys: currentKeys,
		}

		if options.ConsistentRead != nil {
			keysAndAttributes.ConsistentRead = options.ConsistentRead
		}
		if options.ExpressionAttributeNames != nil {
			keysAndAttributes.ExpressionAttributeNames = options.ExpressionAttributeNames
		}
		if options.ProjectionExpression != nil {
			keysAndAttributes.ProjectionExpression = options.ProjectionExpression
		}

		input := &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				tableName: keysAndAttributes,
			},
		}

		if options.ReturnConsumedCapacity != "" {
			input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
		}

		result, err := client.BatchGetItem(context.TODO(), input)
		if err != nil {
			return nil, err
		}

		if tableItems, exists := result.Responses[tableName]; exists {
			allItems = append(allItems, tableItems...)
		}

		if len(result.UnprocessedKeys) == 0 {
			break
		}

		if attempt == maxRetries {
			unprocessedCount := 0
			if unprocessedKeysAndAttrs, exists := result.UnprocessedKeys[tableName]; exists {
				unprocessedCount = len(unprocessedKeysAndAttrs.Keys)
			}
			return allItems, fmt.Errorf("failed to retrieve all items after %d retries, %d items remaining", maxRetries, unprocessedCount)
		}

		// Exponential backoff: 100ms, 200ms, 400ms, 800ms, etc.
		backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
		time.Sleep(backoff)

		if unprocessedKeysAndAttrs, exists := result.UnprocessedKeys[tableName]; exists {
			currentKeys = unprocessedKeysAndAttrs.Keys
		} else {
			currentKeys = []map[string]types.AttributeValue{}
		}
	}

	return allItems, nil
}

// TransactWriteItems executes a transactional write operation and fails the test if an error occurs
func TransactWriteItems(t *testing.T, transactItems []types.TransactWriteItem) *TransactWriteItemsResult {
	result, err := TransactWriteItemsE(t, transactItems)
	require.NoError(t, err)
	return result
}

// TransactWriteItemsE executes a transactional write operation and returns the result and any error
func TransactWriteItemsE(t *testing.T, transactItems []types.TransactWriteItem) (*TransactWriteItemsResult, error) {
	return TransactWriteItemsWithOptionsE(t, transactItems, TransactWriteItemsOptions{})
}

// TransactWriteItemsWithOptions executes a transactional write operation with options and fails the test if an error occurs
func TransactWriteItemsWithOptions(t *testing.T, transactItems []types.TransactWriteItem, options TransactWriteItemsOptions) *TransactWriteItemsResult {
	result, err := TransactWriteItemsWithOptionsE(t, transactItems, options)
	require.NoError(t, err)
	return result
}

// TransactWriteItemsWithOptionsE executes a transactional write operation with options and returns the result and any error
func TransactWriteItemsWithOptionsE(t *testing.T, transactItems []types.TransactWriteItem, options TransactWriteItemsOptions) (*TransactWriteItemsResult, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	}

	if options.ClientRequestToken != nil {
		input.ClientRequestToken = options.ClientRequestToken
	}
	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}
	if options.ReturnItemCollectionMetrics != "" {
		input.ReturnItemCollectionMetrics = options.ReturnItemCollectionMetrics
	}

	result, err := client.TransactWriteItems(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &TransactWriteItemsResult{
		ConsumedCapacity:      result.ConsumedCapacity,
		ItemCollectionMetrics: result.ItemCollectionMetrics,
	}, nil
}

// TransactGetItems executes a transactional read operation and fails the test if an error occurs
func TransactGetItems(t *testing.T, transactItems []types.TransactGetItem) *TransactGetItemsResult {
	result, err := TransactGetItemsE(t, transactItems)
	require.NoError(t, err)
	return result
}

// TransactGetItemsE executes a transactional read operation and returns the result and any error
func TransactGetItemsE(t *testing.T, transactItems []types.TransactGetItem) (*TransactGetItemsResult, error) {
	return TransactGetItemsWithOptionsE(t, transactItems, TransactGetItemsOptions{})
}

// TransactGetItemsWithOptions executes a transactional read operation with options and fails the test if an error occurs
func TransactGetItemsWithOptions(t *testing.T, transactItems []types.TransactGetItem, options TransactGetItemsOptions) *TransactGetItemsResult {
	result, err := TransactGetItemsWithOptionsE(t, transactItems, options)
	require.NoError(t, err)
	return result
}

// TransactGetItemsWithOptionsE executes a transactional read operation with options and returns the result and any error
func TransactGetItemsWithOptionsE(t *testing.T, transactItems []types.TransactGetItem, options TransactGetItemsOptions) (*TransactGetItemsResult, error) {
	client, err := createDynamoDBClient()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.TransactGetItemsInput{
		TransactItems: transactItems,
	}

	if options.ReturnConsumedCapacity != "" {
		input.ReturnConsumedCapacity = options.ReturnConsumedCapacity
	}

	result, err := client.TransactGetItems(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &TransactGetItemsResult{
		Responses:        result.Responses,
		ConsumedCapacity: result.ConsumedCapacity,
	}, nil
}