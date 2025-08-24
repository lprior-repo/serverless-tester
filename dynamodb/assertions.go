package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertTableExists verifies that a DynamoDB table exists and fails the test if it does not
func AssertTableExists(t *testing.T, tableName string) {
	err := AssertTableExistsE(t, tableName)
	require.NoError(t, err)
}

// AssertTableExistsE verifies that a DynamoDB table exists and returns any error
func AssertTableExistsE(t *testing.T, tableName string) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	_, err = client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	
	if err != nil {
		return fmt.Errorf("table %s does not exist: %w", tableName, err)
	}
	
	return nil
}

// AssertTableNotExists verifies that a DynamoDB table does not exist and fails the test if it does
func AssertTableNotExists(t *testing.T, tableName string) {
	err := AssertTableNotExistsE(t, tableName)
	require.NoError(t, err)
}

// AssertTableNotExistsE verifies that a DynamoDB table does not exist and returns any error
func AssertTableNotExistsE(t *testing.T, tableName string) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	_, err = client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	
	if err == nil {
		return fmt.Errorf("table %s exists but should not", tableName)
	}
	
	// Check if it's a ResourceNotFoundException (which is expected)
	var notFoundErr *types.ResourceNotFoundException
	if !errors.As(err, &notFoundErr) {
		return fmt.Errorf("unexpected error when checking table %s: %w", tableName, err)
	}
	
	return nil
}

// AssertTableStatus verifies that a DynamoDB table has the expected status
func AssertTableStatus(t *testing.T, tableName string, expectedStatus types.TableStatus) {
	err := AssertTableStatusE(t, tableName, expectedStatus)
	require.NoError(t, err)
}

// AssertTableStatusE verifies that a DynamoDB table has the expected status and returns any error
func AssertTableStatusE(t *testing.T, tableName string, expectedStatus types.TableStatus) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	result, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	if err != nil {
		return fmt.Errorf("failed to describe table %s: %w", tableName, err)
	}

	if result.Table.TableStatus != expectedStatus {
		return fmt.Errorf("table %s has status %s but expected %s", tableName, result.Table.TableStatus, expectedStatus)
	}
	
	return nil
}

// AssertItemExists verifies that an item exists in the DynamoDB table and fails the test if it does not
func AssertItemExists(t *testing.T, tableName string, key map[string]types.AttributeValue) {
	err := AssertItemExistsE(t, tableName, key)
	require.NoError(t, err)
}

// AssertItemExistsE verifies that an item exists in the DynamoDB table and returns any error
func AssertItemExistsE(t *testing.T, tableName string, key map[string]types.AttributeValue) error {
	item, err := GetItemE(t, tableName, key)
	if err != nil {
		return fmt.Errorf("failed to get item from table %s: %w", tableName, err)
	}
	
	if item == nil || len(item) == 0 {
		return fmt.Errorf("item with key %v does not exist in table %s", key, tableName)
	}
	
	return nil
}

// AssertItemNotExists verifies that an item does not exist in the DynamoDB table and fails the test if it does
func AssertItemNotExists(t *testing.T, tableName string, key map[string]types.AttributeValue) {
	err := AssertItemNotExistsE(t, tableName, key)
	require.NoError(t, err)
}

// AssertItemNotExistsE verifies that an item does not exist in the DynamoDB table and returns any error
func AssertItemNotExistsE(t *testing.T, tableName string, key map[string]types.AttributeValue) error {
	item, err := GetItemE(t, tableName, key)
	if err != nil {
		return fmt.Errorf("failed to get item from table %s: %w", tableName, err)
	}
	
	if item != nil && len(item) > 0 {
		return fmt.Errorf("item with key %v exists in table %s but should not", key, tableName)
	}
	
	return nil
}

// AssertItemCount verifies that a table contains the expected number of items
func AssertItemCount(t *testing.T, tableName string, expectedCount int) {
	err := AssertItemCountE(t, tableName, expectedCount)
	require.NoError(t, err)
}

// AssertItemCountE verifies that a table contains the expected number of items and returns any error
func AssertItemCountE(t *testing.T, tableName string, expectedCount int) error {
	result, err := ScanE(t, tableName)
	if err != nil {
		return fmt.Errorf("failed to scan table %s: %w", tableName, err)
	}
	
	actualCount := int(result.Count)
	if actualCount != expectedCount {
		return fmt.Errorf("table %s contains %d items but expected %d", tableName, actualCount, expectedCount)
	}
	
	return nil
}

// AssertAttributeEquals verifies that an item's attribute has the expected value
func AssertAttributeEquals(t *testing.T, tableName string, key map[string]types.AttributeValue, attributeName string, expectedValue types.AttributeValue) {
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	require.NoError(t, err)
}

// AssertAttributeEqualsE verifies that an item's attribute has the expected value and returns any error
func AssertAttributeEqualsE(t *testing.T, tableName string, key map[string]types.AttributeValue, attributeName string, expectedValue types.AttributeValue) error {
	item, err := GetItemE(t, tableName, key)
	if err != nil {
		return fmt.Errorf("failed to get item from table %s: %w", tableName, err)
	}
	
	if item == nil || len(item) == 0 {
		return fmt.Errorf("item with key %v does not exist in table %s", key, tableName)
	}
	
	actualValue, exists := item[attributeName]
	if !exists {
		return fmt.Errorf("attribute %s does not exist in item", attributeName)
	}
	
	if !attributeValuesEqual(actualValue, expectedValue) {
		return fmt.Errorf("attribute %s has value %v but expected %v", attributeName, actualValue, expectedValue)
	}
	
	return nil
}

// AssertQueryCount verifies that a query returns the expected number of items
func AssertQueryCount(t *testing.T, tableName string, keyConditionExpression string, expressionAttributeValues map[string]types.AttributeValue, expectedCount int) {
	err := AssertQueryCountE(t, tableName, keyConditionExpression, expressionAttributeValues, expectedCount)
	require.NoError(t, err)
}

// AssertQueryCountE verifies that a query returns the expected number of items and returns any error
func AssertQueryCountE(t *testing.T, tableName string, keyConditionExpression string, expressionAttributeValues map[string]types.AttributeValue, expectedCount int) error {
	result, err := QueryE(t, tableName, keyConditionExpression, expressionAttributeValues)
	if err != nil {
		return fmt.Errorf("failed to query table %s: %w", tableName, err)
	}
	
	actualCount := int(result.Count)
	if actualCount != expectedCount {
		return fmt.Errorf("query returned %d items but expected %d", actualCount, expectedCount)
	}
	
	return nil
}

// AssertBackupExists verifies that a backup exists for the specified table
func AssertBackupExists(t *testing.T, backupArn string) {
	err := AssertBackupExistsE(t, backupArn)
	require.NoError(t, err)
}

// AssertBackupExistsE verifies that a backup exists for the specified table and returns any error
func AssertBackupExistsE(t *testing.T, backupArn string) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	_, err = client.DescribeBackup(context.TODO(), &dynamodb.DescribeBackupInput{
		BackupArn: &backupArn,
	})
	
	if err != nil {
		return fmt.Errorf("backup %s does not exist: %w", backupArn, err)
	}
	
	return nil
}

// AssertStreamEnabled verifies that DynamoDB Streams is enabled on the table
func AssertStreamEnabled(t *testing.T, tableName string) {
	err := AssertStreamEnabledE(t, tableName)
	require.NoError(t, err)
}

// AssertStreamEnabledE verifies that DynamoDB Streams is enabled on the table and returns any error
func AssertStreamEnabledE(t *testing.T, tableName string) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	result, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	if err != nil {
		return fmt.Errorf("failed to describe table %s: %w", tableName, err)
	}

	if result.Table.StreamSpecification == nil {
		return fmt.Errorf("table %s does not have a stream specification", tableName)
	}

	if result.Table.StreamSpecification.StreamEnabled == nil || !*result.Table.StreamSpecification.StreamEnabled {
		return fmt.Errorf("streams are not enabled on table %s", tableName)
	}
	
	return nil
}

// AssertGSIExists verifies that a Global Secondary Index exists on the table
func AssertGSIExists(t *testing.T, tableName string, indexName string) {
	err := AssertGSIExistsE(t, tableName, indexName)
	require.NoError(t, err)
}

// AssertGSIExistsE verifies that a Global Secondary Index exists on the table and returns any error
func AssertGSIExistsE(t *testing.T, tableName string, indexName string) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	result, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	if err != nil {
		return fmt.Errorf("failed to describe table %s: %w", tableName, err)
	}

	for _, gsi := range result.Table.GlobalSecondaryIndexes {
		if gsi.IndexName != nil && *gsi.IndexName == indexName {
			return nil
		}
	}
	
	return fmt.Errorf("GSI %s does not exist on table %s", indexName, tableName)
}

// AssertGSIStatus verifies that a Global Secondary Index has the expected status
func AssertGSIStatus(t *testing.T, tableName string, indexName string, expectedStatus types.IndexStatus) {
	err := AssertGSIStatusE(t, tableName, indexName, expectedStatus)
	require.NoError(t, err)
}

// AssertGSIStatusE verifies that a Global Secondary Index has the expected status and returns any error
func AssertGSIStatusE(t *testing.T, tableName string, indexName string, expectedStatus types.IndexStatus) error {
	client, err := createDynamoDBClient()
	if err != nil {
		return err
	}

	result, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	if err != nil {
		return fmt.Errorf("failed to describe table %s: %w", tableName, err)
	}

	for _, gsi := range result.Table.GlobalSecondaryIndexes {
		if gsi.IndexName != nil && *gsi.IndexName == indexName {
			if gsi.IndexStatus != expectedStatus {
				return fmt.Errorf("GSI %s has status %s but expected %s", indexName, gsi.IndexStatus, expectedStatus)
			}
			return nil
		}
	}
	
	return fmt.Errorf("GSI %s does not exist on table %s", indexName, tableName)
}

// AssertConditionalCheckPassed is a helper to assert that a conditional operation succeeded
func AssertConditionalCheckPassed(t *testing.T, err error) {
	if err != nil {
		var conditionalErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalErr) {
			assert.Fail(t, "Conditional check failed when it should have passed", conditionalErr.Error())
		} else {
			assert.Fail(t, "Operation failed with unexpected error", err.Error())
		}
	}
}

// AssertConditionalCheckFailed is a helper to assert that a conditional operation failed as expected
func AssertConditionalCheckFailed(t *testing.T, err error) {
	require.Error(t, err, "Expected conditional check to fail but operation succeeded")
	
	var conditionalErr *types.ConditionalCheckFailedException
	assert.True(t, errors.As(err, &conditionalErr), "Expected ConditionalCheckFailedException but got: %T", err)
}

// attributeValuesEqual compares two DynamoDB attribute values for equality
func attributeValuesEqual(a, b types.AttributeValue) bool {
	switch av := a.(type) {
	case *types.AttributeValueMemberS:
		if bv, ok := b.(*types.AttributeValueMemberS); ok {
			return av.Value == bv.Value
		}
	case *types.AttributeValueMemberN:
		if bv, ok := b.(*types.AttributeValueMemberN); ok {
			return av.Value == bv.Value
		}
	case *types.AttributeValueMemberB:
		if bv, ok := b.(*types.AttributeValueMemberB); ok {
			return string(av.Value) == string(bv.Value)
		}
	case *types.AttributeValueMemberBOOL:
		if bv, ok := b.(*types.AttributeValueMemberBOOL); ok {
			return av.Value == bv.Value
		}
	case *types.AttributeValueMemberNULL:
		if bv, ok := b.(*types.AttributeValueMemberNULL); ok {
			return av.Value == bv.Value
		}
	case *types.AttributeValueMemberSS:
		if bv, ok := b.(*types.AttributeValueMemberSS); ok {
			return stringSlicesEqual(av.Value, bv.Value)
		}
	case *types.AttributeValueMemberNS:
		if bv, ok := b.(*types.AttributeValueMemberNS); ok {
			return stringSlicesEqual(av.Value, bv.Value)
		}
	case *types.AttributeValueMemberBS:
		if bv, ok := b.(*types.AttributeValueMemberBS); ok {
			return bytesSlicesEqual(av.Value, bv.Value)
		}
	case *types.AttributeValueMemberL:
		if bv, ok := b.(*types.AttributeValueMemberL); ok {
			return attributeValueListsEqual(av.Value, bv.Value)
		}
	case *types.AttributeValueMemberM:
		if bv, ok := b.(*types.AttributeValueMemberM); ok {
			return attributeValueMapsEqual(av.Value, bv.Value)
		}
	}
	return false
}

// stringSlicesEqual compares two string slices for equality
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// bytesSlicesEqual compares two byte slices for equality
func bytesSlicesEqual(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if string(v) != string(b[i]) {
			return false
		}
	}
	return true
}

// attributeValueListsEqual compares two attribute value lists for equality
func attributeValueListsEqual(a, b []types.AttributeValue) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if !attributeValuesEqual(v, b[i]) {
			return false
		}
	}
	return true
}

// attributeValueMapsEqual compares two attribute value maps for equality
func attributeValueMapsEqual(a, b map[string]types.AttributeValue) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, exists := b[k]; !exists || !attributeValuesEqual(v, bv) {
			return false
		}
	}
	return true
}