package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestTransactWriteItems_ShouldExecuteTransactionalWrites(t *testing.T) {
	transactItems := []types.TransactWriteItem{
		{
			Put: &types.Put{
				TableName: stringPtr("test-table-1"),
				Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "test-id-1"},
					"name": &types.AttributeValueMemberS{Value: "test-name-1"},
				},
				ConditionExpression: stringPtr("attribute_not_exists(id)"),
			},
		},
		{
			Update: &types.Update{
				TableName: stringPtr("test-table-2"),
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id-2"},
				},
				UpdateExpression: stringPtr("SET #name = :name"),
				ExpressionAttributeNames: map[string]string{
					"#name": "name",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":name": &types.AttributeValueMemberS{Value: "updated-name"},
				},
			},
		},
		{
			Delete: &types.Delete{
				TableName: stringPtr("test-table-3"),
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id-3"},
				},
				ConditionExpression: stringPtr("attribute_exists(id)"),
			},
		},
	}

	result := TransactWriteItems(t, transactItems)

	assert.NotNil(t, result)
}

func TestTransactWriteItemsE_ShouldReturnErrorWhenFailure(t *testing.T) {
	transactItems := []types.TransactWriteItem{
		{
			Put: &types.Put{
				TableName: stringPtr(""),
				Item: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id"},
				},
			},
		},
	}

	_, err := TransactWriteItemsE(t, transactItems)

	assert.Error(t, err)
}

func TestTransactWriteItems_WithOptions_ShouldApplyClientRequestToken(t *testing.T) {
	transactItems := []types.TransactWriteItem{
		{
			Put: &types.Put{
				TableName: stringPtr("test-table"),
				Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "test-id"},
					"name": &types.AttributeValueMemberS{Value: "test-name"},
				},
			},
		},
	}
	options := TransactWriteItemsOptions{
		ClientRequestToken:     stringPtr("unique-request-token"),
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
		ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsSize,
	}

	result := TransactWriteItemsWithOptions(t, transactItems, options)

	assert.NotNil(t, result)
}

func TestTransactGetItems_ShouldRetrieveItemsTransactionally(t *testing.T) {
	transactItems := []types.TransactGetItem{
		{
			Get: &types.Get{
				TableName: stringPtr("test-table-1"),
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id-1"},
				},
			},
		},
		{
			Get: &types.Get{
				TableName: stringPtr("test-table-2"),
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id-2"},
				},
				ProjectionExpression: stringPtr("id, #name"),
				ExpressionAttributeNames: map[string]string{
					"#name": "name",
				},
			},
		},
	}

	result := TransactGetItems(t, transactItems)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Responses)
}

func TestTransactGetItemsE_ShouldReturnErrorWhenFailure(t *testing.T) {
	transactItems := []types.TransactGetItem{
		{
			Get: &types.Get{
				TableName: stringPtr(""),
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id"},
				},
			},
		},
	}

	_, err := TransactGetItemsE(t, transactItems)

	assert.Error(t, err)
}

func TestTransactGetItems_WithOptions_ShouldApplyReturnConsumedCapacity(t *testing.T) {
	transactItems := []types.TransactGetItem{
		{
			Get: &types.Get{
				TableName: stringPtr("test-table"),
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "test-id"},
				},
			},
		},
	}
	options := TransactGetItemsOptions{
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	result := TransactGetItemsWithOptions(t, transactItems, options)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Responses)
}