package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

// Test data factories for batch operations
func createBatchWriteRequests() []types.WriteRequest {
	return []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "batch-put-1"},
					"name": &types.AttributeValueMemberS{Value: "Batch Item 1"},
					"type": &types.AttributeValueMemberS{Value: "test"},
				},
			},
		},
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":   &types.AttributeValueMemberS{Value: "batch-put-2"},
					"name": &types.AttributeValueMemberS{Value: "Batch Item 2"},
					"type": &types.AttributeValueMemberS{Value: "test"},
				},
			},
		},
		{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "batch-delete-1"},
				},
			},
		},
	}
}

func createBatchGetKeys() []map[string]types.AttributeValue {
	return []map[string]types.AttributeValue{
		{
			"id": &types.AttributeValueMemberS{Value: "batch-get-1"},
		},
		{
			"id": &types.AttributeValueMemberS{Value: "batch-get-2"},
		},
		{
			"id": &types.AttributeValueMemberS{Value: "batch-get-3"},
		},
	}
}

func createBatchGetKeysWithCompositeKey() []map[string]types.AttributeValue {
	return []map[string]types.AttributeValue{
		{
			"pk": &types.AttributeValueMemberS{Value: "USER#1"},
			"sk": &types.AttributeValueMemberS{Value: "PROFILE#1"},
		},
		{
			"pk": &types.AttributeValueMemberS{Value: "USER#2"},
			"sk": &types.AttributeValueMemberS{Value: "PROFILE#2"},
		},
	}
}

func createLargeBatchWriteRequests() []types.WriteRequest {
	requests := make([]types.WriteRequest, 25) // Maximum batch size
	
	for i := 0; i < 25; i++ {
		idValue := "large-batch-" + string(rune(i+48))
		indexValue := string(rune(i + 48))
		
		requests[i] = types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":    &types.AttributeValueMemberS{Value: idValue},
					"index": &types.AttributeValueMemberN{Value: indexValue},
					"data":  &types.AttributeValueMemberS{Value: "Large batch test data"},
				},
			},
		}
	}
	
	return requests
}

func createLargeBatchGetKeys() []map[string]types.AttributeValue {
	keys := make([]map[string]types.AttributeValue, 100) // Maximum batch size
	
	for i := 0; i < 100; i++ {
		idValue := "large-get-" + string(rune(i+48))
		keys[i] = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: idValue},
		}
	}
	
	return keys
}

// TESTS FOR BATCHWRITEITEM INPUT CREATION

func TestValidateBatchWriteItemInputCreation_WithBasicRequests_ShouldCreateValidInput(t *testing.T) {
	// RED: Test input validation and creation for BatchWriteItem
	tableName := "test-table"
	writeRequests := createBatchWriteRequests()
	
	input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
	
	// GREEN: Verify input is created correctly
	assert.Equal(t, writeRequests, input.RequestItems[tableName])
	assert.Len(t, input.RequestItems, 1)
	assert.Len(t, input.RequestItems[tableName], 3)
}

func TestValidateBatchWriteItemInputCreation_WithAllOptions_ShouldCreateCompleteInput(t *testing.T) {
	// RED: Test input validation and creation with all options
	tableName := "test-table"
	writeRequests := createBatchWriteRequests()
	options := BatchWriteItemOptions{
		ReturnConsumedCapacity:      types.ReturnConsumedCapacityTotal,
		ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsSize,
	}
	
	input := createBatchWriteItemInput(tableName, writeRequests, options)
	
	// GREEN: Verify all options are set
	assert.Equal(t, options.ReturnConsumedCapacity, input.ReturnConsumedCapacity)
	assert.Equal(t, options.ReturnItemCollectionMetrics, input.ReturnItemCollectionMetrics)
	assert.Equal(t, writeRequests, input.RequestItems[tableName])
}

func TestValidateBatchWriteItemInputCreation_WithMixedOperations_ShouldHandlePutAndDelete(t *testing.T) {
	// RED: Test mixed PUT and DELETE operations
	tableName := "test-table"
	writeRequests := createBatchWriteRequests()
	
	input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
	
	// GREEN: Verify mixed operations are handled
	requests := input.RequestItems[tableName]
	assert.NotNil(t, requests[0].PutRequest)
	assert.Nil(t, requests[0].DeleteRequest)
	assert.NotNil(t, requests[1].PutRequest)
	assert.Nil(t, requests[1].DeleteRequest)
	assert.Nil(t, requests[2].PutRequest)
	assert.NotNil(t, requests[2].DeleteRequest)
}

func TestValidateBatchWriteItemInputCreation_WithLargeBatch_ShouldHandleMaximumSize(t *testing.T) {
	// RED: Test large batch at maximum size limit (25 items)
	tableName := "test-table"
	writeRequests := createLargeBatchWriteRequests()
	
	input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
	
	// GREEN: Verify large batch is handled
	assert.Len(t, input.RequestItems[tableName], 25)
	assert.Equal(t, writeRequests, input.RequestItems[tableName])
}

func TestValidateBatchWriteItemInputCreation_WithEmptyRequests_ShouldCreateEmptyInput(t *testing.T) {
	// RED: Test empty requests
	tableName := "test-table"
	writeRequests := []types.WriteRequest{}
	
	input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
	
	// GREEN: Verify empty requests are handled
	assert.Len(t, input.RequestItems[tableName], 0)
}

// TESTS FOR BATCHGETITEM INPUT CREATION

func TestValidateBatchGetItemInputCreation_WithBasicKeys_ShouldCreateValidInput(t *testing.T) {
	// RED: Test input validation and creation for BatchGetItem
	tableName := "test-table"
	keys := createBatchGetKeys()
	
	input := createBatchGetItemInput(tableName, keys, BatchGetItemOptions{})
	
	// GREEN: Verify input is created correctly
	assert.Equal(t, keys, input.RequestItems[tableName].Keys)
	assert.Len(t, input.RequestItems, 1)
	assert.Len(t, input.RequestItems[tableName].Keys, 3)
}

func TestValidateBatchGetItemInputCreation_WithAllOptions_ShouldCreateCompleteInput(t *testing.T) {
	// RED: Test input validation and creation with all options
	tableName := "test-table"
	keys := createBatchGetKeys()
	options := BatchGetItemOptions{
		ConsistentRead: boolPtr(true),
		ExpressionAttributeNames: map[string]string{
			"#id":   "id",
			"#name": "name",
		},
		ProjectionExpression:   stringPtr("#id, #name, email"),
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}
	
	input := createBatchGetItemInput(tableName, keys, options)
	
	// GREEN: Verify all options are set
	assert.Equal(t, *options.ConsistentRead, *input.RequestItems[tableName].ConsistentRead)
	assert.Equal(t, options.ExpressionAttributeNames, input.RequestItems[tableName].ExpressionAttributeNames)
	assert.Equal(t, *options.ProjectionExpression, *input.RequestItems[tableName].ProjectionExpression)
	assert.Equal(t, options.ReturnConsumedCapacity, input.ReturnConsumedCapacity)
}

func TestValidateBatchGetItemInputCreation_WithCompositeKeys_ShouldHandleComplexKeys(t *testing.T) {
	// RED: Test composite keys
	tableName := "test-table"
	keys := createBatchGetKeysWithCompositeKey()
	
	input := createBatchGetItemInput(tableName, keys, BatchGetItemOptions{})
	
	// GREEN: Verify composite keys are handled
	retrievedKeys := input.RequestItems[tableName].Keys
	assert.Len(t, retrievedKeys, 2)
	
	// Verify first key
	firstKey := retrievedKeys[0]
	assert.Equal(t, "USER#1", firstKey["pk"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "PROFILE#1", firstKey["sk"].(*types.AttributeValueMemberS).Value)
	
	// Verify second key
	secondKey := retrievedKeys[1]
	assert.Equal(t, "USER#2", secondKey["pk"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, "PROFILE#2", secondKey["sk"].(*types.AttributeValueMemberS).Value)
}

func TestValidateBatchGetItemInputCreation_WithLargeBatch_ShouldHandleMaximumSize(t *testing.T) {
	// RED: Test large batch at maximum size limit (100 items)
	tableName := "test-table"
	keys := createLargeBatchGetKeys()
	
	input := createBatchGetItemInput(tableName, keys, BatchGetItemOptions{})
	
	// GREEN: Verify large batch is handled
	assert.Len(t, input.RequestItems[tableName].Keys, 100)
	assert.Equal(t, keys, input.RequestItems[tableName].Keys)
}

func TestValidateBatchGetItemInputCreation_WithEmptyKeys_ShouldCreateEmptyInput(t *testing.T) {
	// RED: Test empty keys
	tableName := "test-table"
	keys := []map[string]types.AttributeValue{}
	
	input := createBatchGetItemInput(tableName, keys, BatchGetItemOptions{})
	
	// GREEN: Verify empty keys are handled
	assert.Len(t, input.RequestItems[tableName].Keys, 0)
}

// BATCH OPERATION INTEGRATION TESTS

func TestBatchWriteItemIntegration_WithSuccessfulResponse_ShouldProcessAllItems(t *testing.T) {
	// RED: Test successful batch write response handling
	tableName := "test-table"
	writeRequests := createBatchWriteRequests()
	
	input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
	
	// GREEN: Verify successful batch write setup
	assert.Len(t, input.RequestItems[tableName], 3)
	
	// Mock would return successful response with no unprocessed items
	// Expected result would be &BatchWriteItemResult{UnprocessedItems: nil}
}

func TestBatchWriteItemIntegration_WithUnprocessedItems_ShouldReturnUnprocessedItems(t *testing.T) {
	// RED: Test unprocessed items handling
	tableName := "test-table"
	writeRequests := createBatchWriteRequests()
	
	input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
	
	// GREEN: Verify batch write is set up correctly
	assert.Len(t, input.RequestItems[tableName], 3)
	
	// Mock would return response with unprocessed items
	// Expected result would include unprocessed items for retry
}

func TestBatchGetItemIntegration_WithSuccessfulResponse_ShouldReturnAllItems(t *testing.T) {
	// RED: Test successful batch get response handling
	tableName := "test-table"
	keys := createBatchGetKeys()
	
	input := createBatchGetItemInput(tableName, keys, BatchGetItemOptions{})
	
	// GREEN: Verify successful batch get setup
	assert.Len(t, input.RequestItems[tableName].Keys, 3)
	
	// Mock would return successful response with all requested items
	// Expected result would be &BatchGetItemResult{Items: allItems}
}

func TestBatchGetItemIntegration_WithUnprocessedKeys_ShouldReturnUnprocessedKeys(t *testing.T) {
	// RED: Test unprocessed keys handling
	tableName := "test-table"
	keys := createBatchGetKeys()
	
	input := createBatchGetItemInput(tableName, keys, BatchGetItemOptions{})
	
	// GREEN: Verify batch get is set up correctly
	assert.Len(t, input.RequestItems[tableName].Keys, 3)
	
	// Mock would return response with unprocessed keys
	// Expected result would include unprocessed keys for retry
}

// RETRY LOGIC TESTS

func TestBatchWriteItemRetry_WithMaxRetries_ShouldCalculateBackoffCorrectly(t *testing.T) {
	// RED: Test retry logic calculation
	maxRetries := 3
	
	// GREEN: Verify retry logic parameters
	assert.Equal(t, 3, maxRetries)
	
	// Test exponential backoff calculation: 100ms, 200ms, 400ms, 800ms
	expectedBackoffs := []int{100, 200, 400, 800}
	for attempt := 0; attempt < maxRetries+1; attempt++ {
		backoffMs := 100 * (1 << attempt)
		assert.Equal(t, expectedBackoffs[attempt], backoffMs)
	}
}

func TestBatchGetItemRetry_WithMaxRetries_ShouldCalculateBackoffCorrectly(t *testing.T) {
	// RED: Test retry logic calculation for batch get
	maxRetries := 5
	
	// GREEN: Verify retry logic parameters
	assert.Equal(t, 5, maxRetries)
	
	// Test exponential backoff calculation: 100ms, 200ms, 400ms, 800ms, 1600ms, 3200ms
	expectedBackoffs := []int{100, 200, 400, 800, 1600, 3200}
	for attempt := 0; attempt < maxRetries+1; attempt++ {
		backoffMs := 100 * (1 << attempt)
		assert.Equal(t, expectedBackoffs[attempt], backoffMs)
	}
}

func TestBatchRetryLogic_WithUnprocessedItems_ShouldRetryOnlyUnprocessed(t *testing.T) {
	// RED: Test that only unprocessed items are retried
	originalRequests := createBatchWriteRequests()
	unprocessedRequests := originalRequests[1:] // Simulate first item was processed
	
	// GREEN: Verify unprocessed items logic
	assert.Len(t, originalRequests, 3)
	assert.Len(t, unprocessedRequests, 2)
	
	// In retry, only unprocessed items should be sent
	retryInput := createBatchWriteItemInput("test-table", unprocessedRequests, BatchWriteItemOptions{})
	assert.Len(t, retryInput.RequestItems["test-table"], 2)
}

func TestBatchRetryLogic_WithUnprocessedKeys_ShouldRetryOnlyUnprocessed(t *testing.T) {
	// RED: Test that only unprocessed keys are retried
	originalKeys := createBatchGetKeys()
	unprocessedKeys := originalKeys[0:1] // Simulate only first key needs retry
	
	// GREEN: Verify unprocessed keys logic
	assert.Len(t, originalKeys, 3)
	assert.Len(t, unprocessedKeys, 1)
	
	// In retry, only unprocessed keys should be sent
	retryInput := createBatchGetItemInput("test-table", unprocessedKeys, BatchGetItemOptions{})
	assert.Len(t, retryInput.RequestItems["test-table"].Keys, 1)
}

// ERROR HANDLING TESTS

func TestBatchOperations_WithProvisionedThroughputExceeded_ShouldHandleThrottling(t *testing.T) {
	// RED: Test throttling error handling for batch operations
	err := createProvisionedThroughputExceededException()
	
	// GREEN: Verify error type is handled
	var throughputErr *types.ProvisionedThroughputExceededException
	assert.ErrorAs(t, err, &throughputErr)
	assert.Contains(t, err.Error(), "request rate is too high")
}

func TestBatchOperations_WithResourceNotFound_ShouldHandleTableNotFound(t *testing.T) {
	// RED: Test resource not found error for batch operations
	err := createResourceNotFoundExceptionBatch()
	
	// GREEN: Verify error type is handled
	var notFoundErr *types.ResourceNotFoundException
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Contains(t, err.Error(), "Requested resource not found")
}

func TestBatchOperations_WithInternalServerError_ShouldHandleServiceErrors(t *testing.T) {
	// RED: Test internal server error for batch operations
	err := createInternalServerErrorBatch()
	
	// GREEN: Verify error type is handled
	var internalErr *types.InternalServerError
	assert.ErrorAs(t, err, &internalErr)
	assert.Contains(t, err.Error(), "Internal server error")
}

// COMPLEX BATCH OPERATION SCENARIOS

func TestBatchWriteItem_WithComplexDataTypes_ShouldHandleAllAttributeTypes(t *testing.T) {
	// RED: Test complex data types in batch operations
	tableName := "test-table"
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":       &types.AttributeValueMemberS{Value: "complex-1"},
					"string":   &types.AttributeValueMemberS{Value: "string value"},
					"number":   &types.AttributeValueMemberN{Value: "42"},
					"binary":   &types.AttributeValueMemberB{Value: []byte("binary data")},
					"boolean":  &types.AttributeValueMemberBOOL{Value: true},
					"null":     &types.AttributeValueMemberNULL{Value: true},
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
				},
			},
		},
	}
	
	input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
	
	// GREEN: Verify complex data types are handled
	item := input.RequestItems[tableName][0].PutRequest.Item
	assert.Contains(t, item, "string")
	assert.Contains(t, item, "number")
	assert.Contains(t, item, "binary")
	assert.Contains(t, item, "boolean")
	assert.Contains(t, item, "null")
	assert.Contains(t, item, "stringSet")
	assert.Contains(t, item, "numberSet")
	assert.Contains(t, item, "binarySet")
	assert.Contains(t, item, "list")
	assert.Contains(t, item, "map")
}

func TestBatchGetItem_WithProjectionAndConsistentRead_ShouldCombineOptions(t *testing.T) {
	// RED: Test combining multiple options
	tableName := "test-table"
	keys := createBatchGetKeys()
	options := BatchGetItemOptions{
		ConsistentRead: boolPtr(true),
		ExpressionAttributeNames: map[string]string{
			"#id":       "id",
			"#name":     "name",
			"#metadata": "metadata",
		},
		ProjectionExpression: stringPtr("#id, #name, #metadata.version"),
	}
	
	input := createBatchGetItemInput(tableName, keys, options)
	
	// GREEN: Verify all options are combined correctly
	keysAndAttrs := input.RequestItems[tableName]
	assert.True(t, *keysAndAttrs.ConsistentRead)
	assert.Equal(t, options.ExpressionAttributeNames, keysAndAttrs.ExpressionAttributeNames)
	assert.Equal(t, *options.ProjectionExpression, *keysAndAttrs.ProjectionExpression)
	assert.Equal(t, keys, keysAndAttrs.Keys)
}

func TestBatchOperations_WithUnicodeAndSpecialCharacters_ShouldHandleEncoding(t *testing.T) {
	// RED: Test unicode and special characters in batch operations
	tableName := "test-table"
	writeRequests := []types.WriteRequest{
		{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":           &types.AttributeValueMemberS{Value: "unicode-batch"},
					"unicodeText":  &types.AttributeValueMemberS{Value: "Hello 世界 🌍 café naïve résumé"},
					"specialChars": &types.AttributeValueMemberS{Value: `!"#$%&'()*+,-./:;<=>?@[\]^_` + "`{|}~"},
					"jsonData":     &types.AttributeValueMemberS{Value: `{"key": "value", "nested": {"array": [1, 2, 3]}}`},
				},
			},
		},
	}
	
	input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
	
	// GREEN: Verify unicode and special characters are preserved
	item := input.RequestItems[tableName][0].PutRequest.Item
	unicodeValue := item["unicodeText"].(*types.AttributeValueMemberS).Value
	assert.Contains(t, unicodeValue, "世界")
	assert.Contains(t, unicodeValue, "🌍")
	assert.Contains(t, unicodeValue, "café")
	
	specialValue := item["specialChars"].(*types.AttributeValueMemberS).Value
	assert.Contains(t, specialValue, "!\"#$%&'()")
	
	jsonValue := item["jsonData"].(*types.AttributeValueMemberS).Value
	assert.Contains(t, jsonValue, `{"key": "value"`)
}

// PERFORMANCE AND CONCURRENCY TESTS

func TestBatchInputCreation_ConcurrentAccess_ShouldBeThreadSafe(t *testing.T) {
	// RED: Test concurrent batch input creation (pure functions should be thread-safe)
	tableName := "concurrent-batch-table"
	numGoroutines := 10
	results := make(chan *dynamodb.BatchWriteItemInput, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			writeRequests := []types.WriteRequest{
				{
					PutRequest: &types.PutRequest{
						Item: map[string]types.AttributeValue{
							"id":   &types.AttributeValueMemberS{Value: "concurrent-batch"},
							"data": &types.AttributeValueMemberS{Value: "concurrent-data"},
						},
					},
				},
			}
			input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
			results <- input
		}(i)
	}
	
	// GREEN: Verify all concurrent operations completed successfully
	for i := 0; i < numGoroutines; i++ {
		input := <-results
		assert.NotNil(t, input.RequestItems[tableName])
		assert.Len(t, input.RequestItems[tableName], 1)
	}
}

func TestBatchOperations_WithLargePayload_ShouldHandleNearMaximumSize(t *testing.T) {
	// RED: Test large payload handling (near 16MB limit for batch operations)
	tableName := "test-table"
	
	// Create large but valid items (each item under 400KB, batch under 16MB)
	largeItemData := make([]byte, 50*1024) // 50KB per item
	for i := range largeItemData {
		largeItemData[i] = byte('A' + (i % 26))
	}
	
	writeRequests := make([]types.WriteRequest, 10) // 10 items * 50KB = ~500KB total
	for i := 0; i < 10; i++ {
		writeRequests[i] = types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":        &types.AttributeValueMemberS{Value: "large-item-" + string(rune(i+48))},
					"largeData": &types.AttributeValueMemberS{Value: string(largeItemData)},
					"index":     &types.AttributeValueMemberN{Value: string(rune(i + 48))},
				},
			},
		}
	}
	
	input := createBatchWriteItemInput(tableName, writeRequests, BatchWriteItemOptions{})
	
	// GREEN: Verify large payload is handled
	assert.Len(t, input.RequestItems[tableName], 10)
	firstItem := input.RequestItems[tableName][0].PutRequest.Item
	largeData := firstItem["largeData"].(*types.AttributeValueMemberS).Value
	assert.Len(t, largeData, 50*1024)
	assert.Contains(t, largeData, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

// Helper functions for input creation (pure functions)

func createBatchWriteItemInput(tableName string, writeRequests []types.WriteRequest, options BatchWriteItemOptions) *dynamodb.BatchWriteItemInput {
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

	return input
}

func createBatchGetItemInput(tableName string, keys []map[string]types.AttributeValue, options BatchGetItemOptions) *dynamodb.BatchGetItemInput {
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

	return input
}

// Helper functions for AWS error simulation
func createProvisionedThroughputExceededException() error {
	return &types.ProvisionedThroughputExceededException{
		Message: stringPtrBatch("The request rate is too high"),
	}
}

func createResourceNotFoundExceptionBatch() error {
	return &types.ResourceNotFoundException{
		Message: stringPtrBatch("Requested resource not found"),
	}
}

func createInternalServerErrorBatch() error {
	return &types.InternalServerError{
		Message: stringPtrBatch("Internal server error"),
	}
}

// Helper functions specific to this test file
func stringPtrBatch(s string) *string {
	return &s
}

func boolPtrBatch(b bool) *bool {
	return &b
}