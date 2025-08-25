package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestQuery_ShouldReturnItemsMatchingCondition(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "id = :id"
	expressionAttributeValues := map[string]types.AttributeValue{
		":id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	result := Query(t, tableName, keyConditionExpression, expressionAttributeValues)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Items)
}

func TestQueryE_ShouldReturnErrorWhenFailure(t *testing.T) {
	invalidTableName := ""
	keyConditionExpression := "id = :id"
	expressionAttributeValues := map[string]types.AttributeValue{
		":id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	_, err := QueryE(t, invalidTableName, keyConditionExpression, expressionAttributeValues)

	assert.Error(t, err)
}

func TestQuery_WithOptions_ShouldApplyFilterExpression(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "id = :id"
	expressionAttributeValues := map[string]types.AttributeValue{
		":id":   &types.AttributeValueMemberS{Value: "test-id"},
		":name": &types.AttributeValueMemberS{Value: "test-name"},
	}
	options := QueryOptions{
		FilterExpression:      stringPtr("#name = :name"),
		ExpressionAttributeNames: map[string]string{
			"#name": "name",
		},
		Limit: int32Ptr(10),
		ScanIndexForward: boolPtr(true),
	}

	result := QueryWithOptions(t, tableName, keyConditionExpression, expressionAttributeValues, options)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Items)
}

func TestScan_ShouldReturnAllItemsInTable(t *testing.T) {
	tableName := "test-table"

	result := Scan(t, tableName)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Items)
}

func TestScanE_ShouldReturnErrorWhenFailure(t *testing.T) {
	invalidTableName := ""

	_, err := ScanE(t, invalidTableName)

	assert.Error(t, err)
}

func TestScan_WithOptions_ShouldApplyFilterExpression(t *testing.T) {
	tableName := "test-table"
	options := ScanOptions{
		FilterExpression: stringPtr("#name = :name"),
		ExpressionAttributeNames: map[string]string{
			"#name": "name",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name": &types.AttributeValueMemberS{Value: "test-name"},
		},
		Limit: int32Ptr(10),
		Segment: int32Ptr(0),
		TotalSegments: int32Ptr(1),
	}

	result := ScanWithOptions(t, tableName, options)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Items)
}

func TestQueryAllPages_ShouldReturnAllItemsWithPagination(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "id = :id"
	expressionAttributeValues := map[string]types.AttributeValue{
		":id": &types.AttributeValueMemberS{Value: "test-id"},
	}

	items := QueryAllPages(t, tableName, keyConditionExpression, expressionAttributeValues)

	assert.NotNil(t, items)
}

func TestScanAllPages_ShouldReturnAllItemsWithPagination(t *testing.T) {
	tableName := "test-table"

	items := ScanAllPages(t, tableName)

	assert.NotNil(t, items)
}

