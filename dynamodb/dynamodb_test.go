package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestPutItem_ShouldStoreItemInTable(t *testing.T) {
	tableName := "test-table"
	item := map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: "test-id"},
		"name": &types.AttributeValueMemberS{Value: "test-name"},
	}

	PutItem(t, tableName, item)
}

func TestPutItemE_ShouldReturnErrorWhenFailure(t *testing.T) {
	invalidTableName := ""
	item := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	err := PutItemE(t, invalidTableName, item)

	assert.Error(t, err)
}

func TestPutItem_WithOptions_ShouldApplyConditionExpression(t *testing.T) {
	tableName := "test-table"
	item := map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: "test-id"},
		"name": &types.AttributeValueMemberS{Value: "test-name"},
	}
	options := PutItemOptions{
		ConditionExpression: stringPtr("attribute_not_exists(id)"),
	}

	PutItemWithOptions(t, tableName, item, options)
}

func TestGetItem_ShouldRetrieveExistingItem(t *testing.T) {
	tableName := "test-table"
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	result := GetItem(t, tableName, key)

	assert.NotNil(t, result)
}

func TestGetItemE_ShouldReturnErrorWhenFailure(t *testing.T) {
	invalidTableName := ""
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	_, err := GetItemE(t, invalidTableName, key)

	assert.Error(t, err)
}

func TestGetItem_WithOptions_ShouldApplyProjectionExpression(t *testing.T) {
	tableName := "test-table"
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	options := GetItemOptions{
		ProjectionExpression: stringPtr("id, #name"),
		ExpressionAttributeNames: map[string]string{
			"#name": "name",
		},
	}

	result := GetItemWithOptions(t, tableName, key, options)

	assert.NotNil(t, result)
}

func TestDeleteItem_ShouldRemoveItemFromTable(t *testing.T) {
	tableName := "test-table"
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	DeleteItem(t, tableName, key)
}

func TestDeleteItemE_ShouldReturnErrorWhenFailure(t *testing.T) {
	invalidTableName := ""
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	err := DeleteItemE(t, invalidTableName, key)

	assert.Error(t, err)
}

func TestDeleteItem_WithOptions_ShouldApplyConditionExpression(t *testing.T) {
	tableName := "test-table"
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	options := DeleteItemOptions{
		ConditionExpression: stringPtr("attribute_exists(id)"),
	}

	DeleteItemWithOptions(t, tableName, key, options)
}

func TestUpdateItem_ShouldModifyExistingItem(t *testing.T) {
	tableName := "test-table"
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	updateExpression := "SET #name = :name"
	expressionAttributeNames := map[string]string{
		"#name": "name",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name": &types.AttributeValueMemberS{Value: "updated-name"},
	}

	result := UpdateItem(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues)

	assert.NotNil(t, result)
}

func TestUpdateItemE_ShouldReturnErrorWhenFailure(t *testing.T) {
	invalidTableName := ""
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	updateExpression := "SET #name = :name"

	_, err := UpdateItemE(t, invalidTableName, key, updateExpression, nil, nil)

	assert.Error(t, err)
}

func TestUpdateItem_WithOptions_ShouldApplyConditionExpression(t *testing.T) {
	tableName := "test-table"
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	updateExpression := "SET #name = :name"
	expressionAttributeNames := map[string]string{
		"#name": "name",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name": &types.AttributeValueMemberS{Value: "updated-name"},
	}
	options := UpdateItemOptions{
		ConditionExpression: stringPtr("attribute_exists(id)"),
		ReturnValues:        types.ReturnValueAllNew,
	}

	result := UpdateItemWithOptions(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues, options)

	assert.NotNil(t, result)
}

func stringPtr(s string) *string {
	return &s
}