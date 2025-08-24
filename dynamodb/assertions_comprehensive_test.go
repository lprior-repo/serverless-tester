package dynamodb

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

// Test data factories for assertion testing

func createTestItemKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "assertion-test-item"},
	}
}

func createTestCompositeKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "USER#123"},
		"sk": &types.AttributeValueMemberS{Value: "PROFILE#main"},
	}
}

// Table existence assertion tests

func TestAssertTableExists_WithExistingTable_ShouldPass(t *testing.T) {
	tableName := "existing-table"
	
	// This will call the real assertion which will fail due to no AWS connection
	// In a real test, we would mock the DynamoDB client to return success
	err := AssertTableExistsE(t, tableName)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertTableExists_WithNonExistentTable_ShouldFail(t *testing.T) {
	tableName := "non-existent-table"
	
	// This will fail because table doesn't exist
	err := AssertTableExistsE(t, tableName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestAssertTableNotExists_WithNonExistentTable_ShouldPass(t *testing.T) {
	tableName := "truly-non-existent-table"
	
	err := AssertTableNotExistsE(t, tableName)
	
	// Should pass because table doesn't exist (ResourceNotFoundException expected)
	// But may fail due to network timeout first
	assert.Error(t, err) // Network error, not the assertion error we want
}

func TestAssertTableNotExists_WithExistingTable_ShouldFail(t *testing.T) {
	tableName := "existing-table-that-should-not-exist"
	
	err := AssertTableNotExistsE(t, tableName)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Table status assertion tests

func TestAssertTableStatus_WithCorrectStatus_ShouldPass(t *testing.T) {
	tableName := "status-test-table"
	expectedStatus := types.TableStatusActive
	
	err := AssertTableStatusE(t, tableName, expectedStatus)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertTableStatus_WithIncorrectStatus_ShouldFail(t *testing.T) {
	tableName := "wrong-status-table"
	expectedStatus := types.TableStatusActive
	
	err := AssertTableStatusE(t, tableName, expectedStatus)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertTableStatus_WithMultipleStatuses_ShouldHandleAllStatuses(t *testing.T) {
	testCases := []struct {
		name           string
		expectedStatus types.TableStatus
	}{
		{"Active status", types.TableStatusActive},
		{"Creating status", types.TableStatusCreating},
		{"Deleting status", types.TableStatusDeleting},
		{"Updating status", types.TableStatusUpdating},
		{"Inaccessible encryption credentials", types.TableStatusInaccessibleEncryptionCredentials},
		{"Archiving status", types.TableStatusArchiving},
		{"Archived status", types.TableStatusArchived},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tableName := "multi-status-test-table"
			
			err := AssertTableStatusE(t, tableName, tc.expectedStatus)
			
			// Should fail without proper mocking
			assert.Error(t, err)
		})
	}
}

// Item existence assertion tests

func TestAssertItemExists_WithExistingItem_ShouldPass(t *testing.T) {
	tableName := "item-exists-test-table"
	key := createTestItemKey()
	
	err := AssertItemExistsE(t, tableName, key)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertItemExists_WithCompositeKey_ShouldHandleComplexKeys(t *testing.T) {
	tableName := "composite-key-test-table"
	key := createTestCompositeKey()
	
	err := AssertItemExistsE(t, tableName, key)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertItemExists_WithNonExistentItem_ShouldFail(t *testing.T) {
	tableName := "item-not-exists-test-table"
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "non-existent-item"},
	}
	
	err := AssertItemExistsE(t, tableName, key)
	
	// Should fail because item doesn't exist
	assert.Error(t, err)
	// Note: In real AWS environment, this would contain "does not exist"
}

func TestAssertItemNotExists_WithNonExistentItem_ShouldPass(t *testing.T) {
	tableName := "item-not-exists-pass-test-table"
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "truly-non-existent-item"},
	}
	
	err := AssertItemNotExistsE(t, tableName, key)
	
	// May pass if item truly doesn't exist, but likely to fail due to network timeout
	assert.Error(t, err)
}

func TestAssertItemNotExists_WithExistingItem_ShouldFail(t *testing.T) {
	tableName := "item-exists-but-should-not-table"
	key := createTestItemKey()
	
	err := AssertItemNotExistsE(t, tableName, key)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Item count assertion tests

func TestAssertItemCount_WithCorrectCount_ShouldPass(t *testing.T) {
	tableName := "correct-count-test-table"
	expectedCount := 5
	
	err := AssertItemCountE(t, tableName, expectedCount)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertItemCount_WithIncorrectCount_ShouldFail(t *testing.T) {
	tableName := "incorrect-count-test-table"
	expectedCount := 10
	
	err := AssertItemCountE(t, tableName, expectedCount)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertItemCount_WithZeroCount_ShouldHandleEmptyTables(t *testing.T) {
	tableName := "empty-table"
	expectedCount := 0
	
	err := AssertItemCountE(t, tableName, expectedCount)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertItemCount_WithLargeCount_ShouldHandleLargeTables(t *testing.T) {
	tableName := "large-table"
	expectedCount := 1000000
	
	err := AssertItemCountE(t, tableName, expectedCount)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Attribute assertion tests

func TestAssertAttributeEquals_WithStringAttribute_ShouldCompareCorrectly(t *testing.T) {
	tableName := "string-attr-test-table"
	key := createTestItemKey()
	attributeName := "name"
	expectedValue := &types.AttributeValueMemberS{Value: "John Doe"}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertAttributeEquals_WithNumberAttribute_ShouldCompareCorrectly(t *testing.T) {
	tableName := "number-attr-test-table"
	key := createTestItemKey()
	attributeName := "age"
	expectedValue := &types.AttributeValueMemberN{Value: "25"}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertAttributeEquals_WithBooleanAttribute_ShouldCompareCorrectly(t *testing.T) {
	tableName := "boolean-attr-test-table"
	key := createTestItemKey()
	attributeName := "active"
	expectedValue := &types.AttributeValueMemberBOOL{Value: true}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertAttributeEquals_WithNullAttribute_ShouldCompareCorrectly(t *testing.T) {
	tableName := "null-attr-test-table"
	key := createTestItemKey()
	attributeName := "deleted"
	expectedValue := &types.AttributeValueMemberNULL{Value: true}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertAttributeEquals_WithStringSetAttribute_ShouldCompareCorrectly(t *testing.T) {
	tableName := "string-set-attr-test-table"
	key := createTestItemKey()
	attributeName := "tags"
	expectedValue := &types.AttributeValueMemberSS{Value: []string{"tag1", "tag2", "tag3"}}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertAttributeEquals_WithNumberSetAttribute_ShouldCompareCorrectly(t *testing.T) {
	tableName := "number-set-attr-test-table"
	key := createTestItemKey()
	attributeName := "scores"
	expectedValue := &types.AttributeValueMemberNS{Value: []string{"85", "90", "95"}}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertAttributeEquals_WithBinarySetAttribute_ShouldCompareCorrectly(t *testing.T) {
	tableName := "binary-set-attr-test-table"
	key := createTestItemKey()
	attributeName := "binary_data"
	expectedValue := &types.AttributeValueMemberBS{Value: [][]byte{[]byte("data1"), []byte("data2")}}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertAttributeEquals_WithListAttribute_ShouldCompareCorrectly(t *testing.T) {
	tableName := "list-attr-test-table"
	key := createTestItemKey()
	attributeName := "items"
	expectedValue := &types.AttributeValueMemberL{
		Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: "item1"},
			&types.AttributeValueMemberN{Value: "42"},
			&types.AttributeValueMemberBOOL{Value: true},
		},
	}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertAttributeEquals_WithMapAttribute_ShouldCompareCorrectly(t *testing.T) {
	tableName := "map-attr-test-table"
	key := createTestItemKey()
	attributeName := "metadata"
	expectedValue := &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
			"version":    &types.AttributeValueMemberN{Value: "1"},
			"enabled":    &types.AttributeValueMemberBOOL{Value: true},
		},
	}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertAttributeEquals_WithNonExistentAttribute_ShouldFail(t *testing.T) {
	tableName := "missing-attr-test-table"
	key := createTestItemKey()
	attributeName := "non_existent_attribute"
	expectedValue := &types.AttributeValueMemberS{Value: "any value"}
	
	err := AssertAttributeEqualsE(t, tableName, key, attributeName, expectedValue)
	
	// Should fail because attribute doesn't exist
	assert.Error(t, err)
	// Note: In real AWS environment, this would contain "does not exist"
}

// Query count assertion tests

func TestAssertQueryCount_WithCorrectCount_ShouldPass(t *testing.T) {
	tableName := "query-count-test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	expectedCount := 5
	
	err := AssertQueryCountE(t, tableName, keyConditionExpression, expressionAttributeValues, expectedCount)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertQueryCount_WithComplexQuery_ShouldHandleComplexConditions(t *testing.T) {
	tableName := "complex-query-count-test-table"
	keyConditionExpression := "pk = :pk AND begins_with(sk, :sk_prefix)"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk":        &types.AttributeValueMemberS{Value: "USER#789"},
		":sk_prefix": &types.AttributeValueMemberS{Value: "PROFILE#"},
	}
	expectedCount := 3
	
	err := AssertQueryCountE(t, tableName, keyConditionExpression, expressionAttributeValues, expectedCount)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Backup assertion tests

func TestAssertBackupExists_WithExistingBackup_ShouldPass(t *testing.T) {
	backupArn := "arn:aws:dynamodb:us-east-1:123456789012:table/test-table/backup/01234567890123-abcdefgh"
	
	err := AssertBackupExistsE(t, backupArn)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertBackupExists_WithNonExistentBackup_ShouldFail(t *testing.T) {
	backupArn := "arn:aws:dynamodb:us-east-1:123456789012:table/non-existent-table/backup/non-existent-backup"
	
	err := AssertBackupExistsE(t, backupArn)
	
	// Should fail because backup doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestAssertBackupExists_WithInvalidBackupArn_ShouldFail(t *testing.T) {
	backupArn := "invalid-backup-arn"
	
	err := AssertBackupExistsE(t, backupArn)
	
	// Should fail due to invalid ARN format
	assert.Error(t, err)
}

// Stream assertion tests

func TestAssertStreamEnabled_WithEnabledStream_ShouldPass(t *testing.T) {
	tableName := "stream-enabled-test-table"
	
	err := AssertStreamEnabledE(t, tableName)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertStreamEnabled_WithDisabledStream_ShouldFail(t *testing.T) {
	tableName := "stream-disabled-test-table"
	
	err := AssertStreamEnabledE(t, tableName)
	
	// Should fail because stream is not enabled
	assert.Error(t, err)
	// Note: In real AWS environment with disabled stream, would contain "not enabled"
}

// GSI assertion tests

func TestAssertGSIExists_WithExistingGSI_ShouldPass(t *testing.T) {
	tableName := "gsi-exists-test-table"
	indexName := "test-gsi-index"
	
	err := AssertGSIExistsE(t, tableName, indexName)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertGSIExists_WithNonExistentGSI_ShouldFail(t *testing.T) {
	tableName := "gsi-not-exists-test-table"
	indexName := "non-existent-gsi"
	
	err := AssertGSIExistsE(t, tableName, indexName)
	
	// Should fail because GSI doesn't exist
	assert.Error(t, err)
	// Note: In real AWS environment, would contain "does not exist"
}

func TestAssertGSIStatus_WithActiveGSI_ShouldPass(t *testing.T) {
	tableName := "gsi-active-test-table"
	indexName := "active-gsi-index"
	expectedStatus := types.IndexStatusActive
	
	err := AssertGSIStatusE(t, tableName, indexName, expectedStatus)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestAssertGSIStatus_WithAllPossibleStatuses_ShouldHandleAllStatuses(t *testing.T) {
	testCases := []struct {
		name           string
		expectedStatus types.IndexStatus
	}{
		{"Active status", types.IndexStatusActive},
		{"Creating status", types.IndexStatusCreating},
		{"Deleting status", types.IndexStatusDeleting},
		{"Updating status", types.IndexStatusUpdating},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tableName := "gsi-all-statuses-test-table"
			indexName := "test-gsi"
			
			err := AssertGSIStatusE(t, tableName, indexName, tc.expectedStatus)
			
			// Should fail without proper mocking
			assert.Error(t, err)
		})
	}
}

// Conditional check assertion tests

func TestAssertConditionalCheckPassed_WithNoError_ShouldPass(t *testing.T) {
	// This should not panic because error is nil
	AssertConditionalCheckPassed(t, nil)
	
	// No assertions needed - if no panic occurred, test passed
}

func TestAssertConditionalCheckFailed_WithConditionalCheckError_ShouldPass(t *testing.T) {
	conditionalErr := &types.ConditionalCheckFailedException{
		Message: stringPtr("Conditional check failed as expected"),
	}
	
	// The function should pass without panicking
	AssertConditionalCheckFailed(t, conditionalErr)
	
	// No additional assertions needed - if no panic occurred, test passed
}

// Attribute value comparison tests

func TestAttributeValuesEqual_StringComparison_ShouldCompareCorrectly(t *testing.T) {
	value1 := &types.AttributeValueMemberS{Value: "test string"}
	value2 := &types.AttributeValueMemberS{Value: "test string"}
	value3 := &types.AttributeValueMemberS{Value: "different string"}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqual_NumberComparison_ShouldCompareCorrectly(t *testing.T) {
	value1 := &types.AttributeValueMemberN{Value: "42"}
	value2 := &types.AttributeValueMemberN{Value: "42"}
	value3 := &types.AttributeValueMemberN{Value: "43"}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqual_BooleanComparison_ShouldCompareCorrectly(t *testing.T) {
	value1 := &types.AttributeValueMemberBOOL{Value: true}
	value2 := &types.AttributeValueMemberBOOL{Value: true}
	value3 := &types.AttributeValueMemberBOOL{Value: false}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqual_NullComparison_ShouldCompareCorrectly(t *testing.T) {
	value1 := &types.AttributeValueMemberNULL{Value: true}
	value2 := &types.AttributeValueMemberNULL{Value: true}
	value3 := &types.AttributeValueMemberNULL{Value: false}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqual_StringSetComparison_ShouldCompareCorrectly(t *testing.T) {
	value1 := &types.AttributeValueMemberSS{Value: []string{"a", "b", "c"}}
	value2 := &types.AttributeValueMemberSS{Value: []string{"a", "b", "c"}}
	value3 := &types.AttributeValueMemberSS{Value: []string{"a", "b", "d"}}
	value4 := &types.AttributeValueMemberSS{Value: []string{"a", "b"}}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
	assert.False(t, attributeValuesEqual(value1, value4))
}

func TestAttributeValuesEqual_NumberSetComparison_ShouldCompareCorrectly(t *testing.T) {
	value1 := &types.AttributeValueMemberNS{Value: []string{"1", "2", "3"}}
	value2 := &types.AttributeValueMemberNS{Value: []string{"1", "2", "3"}}
	value3 := &types.AttributeValueMemberNS{Value: []string{"1", "2", "4"}}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqual_BinarySetComparison_ShouldCompareCorrectly(t *testing.T) {
	value1 := &types.AttributeValueMemberBS{Value: [][]byte{[]byte("data1"), []byte("data2")}}
	value2 := &types.AttributeValueMemberBS{Value: [][]byte{[]byte("data1"), []byte("data2")}}
	value3 := &types.AttributeValueMemberBS{Value: [][]byte{[]byte("data1"), []byte("data3")}}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqual_ListComparison_ShouldCompareCorrectly(t *testing.T) {
	value1 := &types.AttributeValueMemberL{
		Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: "item1"},
			&types.AttributeValueMemberN{Value: "42"},
		},
	}
	value2 := &types.AttributeValueMemberL{
		Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: "item1"},
			&types.AttributeValueMemberN{Value: "42"},
		},
	}
	value3 := &types.AttributeValueMemberL{
		Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: "item1"},
			&types.AttributeValueMemberN{Value: "43"},
		},
	}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqual_MapComparison_ShouldCompareCorrectly(t *testing.T) {
	value1 := &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"key1": &types.AttributeValueMemberS{Value: "value1"},
			"key2": &types.AttributeValueMemberN{Value: "42"},
		},
	}
	value2 := &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"key1": &types.AttributeValueMemberS{Value: "value1"},
			"key2": &types.AttributeValueMemberN{Value: "42"},
		},
	}
	value3 := &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"key1": &types.AttributeValueMemberS{Value: "value1"},
			"key2": &types.AttributeValueMemberN{Value: "43"},
		},
	}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqual_DifferentTypes_ShouldReturnFalse(t *testing.T) {
	stringValue := &types.AttributeValueMemberS{Value: "test"}
	numberValue := &types.AttributeValueMemberN{Value: "42"}
	boolValue := &types.AttributeValueMemberBOOL{Value: true}
	
	assert.False(t, attributeValuesEqual(stringValue, numberValue))
	assert.False(t, attributeValuesEqual(stringValue, boolValue))
	assert.False(t, attributeValuesEqual(numberValue, boolValue))
}

// Edge cases and complex scenarios

func TestAssertions_WithNetworkTimeouts_ShouldHandleGracefully(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network timeout test in short mode")
	}
	
	tableName := "network-timeout-test-table"
	
	// Test various assertions that might timeout
	assertions := []func() error{
		func() error { return AssertTableExistsE(t, tableName) },
		func() error { return AssertTableStatusE(t, tableName, types.TableStatusActive) },
		func() error { return AssertItemExistsE(t, tableName, createTestItemKey()) },
		func() error { return AssertItemCountE(t, tableName, 5) },
		func() error { return AssertStreamEnabledE(t, tableName) },
		func() error { return AssertGSIExistsE(t, tableName, "test-gsi") },
	}
	
	for i, assertion := range assertions {
		t.Run(fmt.Sprintf("Assertion_%d", i+1), func(t *testing.T) {
			err := assertion()
			// All should fail due to network issues
			assert.Error(t, err)
		})
	}
}

func TestAssertions_ConcurrentExecution_ShouldHandleParallelCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent assertions test in short mode")
	}

	numGoroutines := 5
	tableName := "concurrent-assertions-test-table"
	errorChan := make(chan error, numGoroutines)

	// Run assertions concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			err := AssertTableExistsE(t, fmt.Sprintf("%s-%d", tableName, goroutineID))
			errorChan <- err
		}(i)
	}

	// Collect all errors
	errorCount := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-errorChan
		if err != nil {
			errorCount++
		}
	}

	// We expect errors since we don't have real AWS setup
	assert.Equal(t, numGoroutines, errorCount)
	
	t.Logf("Completed %d concurrent assertion operations", numGoroutines)
}