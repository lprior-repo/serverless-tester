package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestBatchWriteItem_ShouldWriteMultipleItems(t *testing.T) {
	tableName := "test-table"
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "test-id-1"},
					"name": &types.AttributeValueMemberS{Value: "test-name-1"},
				},
			},
		},
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "test-id-2"},
					"name": &types.AttributeValueMemberS{Value: "test-name-2"},
				},
			},
		},
		{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id-3"},
				},
			},
		},
	}

	result := BatchWriteItem(t, tableName, writeRequests)

	assert.NotNil(t, result)
}

func TestBatchWriteItemE_ShouldReturnErrorWhenFailure(t *testing.T) {
	invalidTableName := ""
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id"},
				},
			},
		},
	}

	_, err := BatchWriteItemE(t, invalidTableName, writeRequests)

	assert.Error(t, err)
}

func TestBatchGetItem_ShouldRetrieveMultipleItems(t *testing.T) {
	tableName := "test-table"
	keys := []map[string]types.AttributeValue{
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-1"},
		},
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-2"},
		},
	}

	result := BatchGetItem(t, tableName, keys)

	assert.NotNil(t, result)
}

func TestBatchGetItemE_ShouldReturnErrorWhenFailure(t *testing.T) {
	invalidTableName := ""
	keys := []map[string]types.AttributeValue{
		{
			"id": &types.AttributeValueMemberS{Value: "test-id"},
		},
	}

	_, err := BatchGetItemE(t, invalidTableName, keys)

	assert.Error(t, err)
}

func TestBatchGetItem_WithOptions_ShouldApplyProjectionExpression(t *testing.T) {
	tableName := "test-table"
	keys := []map[string]types.AttributeValue{
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-1"},
		},
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-2"},
		},
	}
	options := BatchGetItemOptions{
		ProjectionExpression: stringPtr("id, #name"),
		ExpressionAttributeNames: map[string]string{
			"#name": "name",
		},
		ConsistentRead: boolPtr(true),
	}

	result := BatchGetItemWithOptions(t, tableName, keys, options)

	assert.NotNil(t, result)
}

func TestBatchWriteItemWithRetry_ShouldHandleUnprocessedItems(t *testing.T) {
	tableName := "test-table"
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "test-id-1"},
					"name": &types.AttributeValueMemberS{Value: "test-name-1"},
				},
			},
		},
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "test-id-2"},
					"name": &types.AttributeValueMemberS{Value: "test-name-2"},
				},
			},
		},
	}

	BatchWriteItemWithRetry(t, tableName, writeRequests, 3)
}

func TestBatchGetItemWithRetry_ShouldHandleUnprocessedKeys(t *testing.T) {
	tableName := "test-table"
	keys := []map[string]types.AttributeValue{
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-1"},
		},
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-2"},
		},
		{
			"id": &types.AttributeValueMemberS{Value: "test-id-3"},
		},
	}

	result := BatchGetItemWithRetry(t, tableName, keys, 3)

	assert.NotNil(t, result)
}