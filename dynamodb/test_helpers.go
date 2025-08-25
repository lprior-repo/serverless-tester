package dynamodb

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Shared test helper functions to avoid duplication across test files

// Pointer helper functions
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

// Batch operation test data factories
func createBatchWriteRequestsWithSize(numItems int) []types.WriteRequest {
	requests := make([]types.WriteRequest, numItems)
	
	for i := 0; i < numItems; i++ {
		if i%2 == 0 {
			// Put request
			requests[i] = types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"id":          &types.AttributeValueMemberS{Value: fmt.Sprintf("batch-put-%03d", i+1)},
						"name":        &types.AttributeValueMemberS{Value: fmt.Sprintf("Batch Item %d", i+1)},
						"type":        &types.AttributeValueMemberS{Value: "put"},
						"batch_num":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", i+1)},
						"created_at":  &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
						"data":        &types.AttributeValueMemberS{Value: fmt.Sprintf("batch data for item %d", i+1)},
					},
				},
			}
		} else {
			// Delete request
			requests[i] = types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{Value: fmt.Sprintf("batch-delete-%03d", i+1)},
					},
				},
			}
		}
	}
	
	return requests
}

func createSimpleBatchWriteRequests() []types.WriteRequest {
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

func createBatchGetKeysWithSize(numKeys int) []map[string]types.AttributeValue {
	keys := make([]map[string]types.AttributeValue, numKeys)
	
	for i := 0; i < numKeys; i++ {
		keys[i] = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: fmt.Sprintf("batch-get-%03d", i+1)},
		}
	}
	
	return keys
}

func createSimpleBatchGetKeys() []map[string]types.AttributeValue {
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

func createLargeBatchWriteRequestsWithSize(numItems int) []types.WriteRequest {
	requests := make([]types.WriteRequest, numItems)
	
	for i := 0; i < numItems; i++ {
		requests[i] = types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"id":          &types.AttributeValueMemberS{Value: fmt.Sprintf("large-batch-%05d", i+1)},
					"name":        &types.AttributeValueMemberS{Value: fmt.Sprintf("Large Batch Item %d", i+1)},
					"type":        &types.AttributeValueMemberS{Value: "large-batch"},
					"batch_size":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", numItems)},
					"item_index":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", i)},
					"created_at":  &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
					"description": &types.AttributeValueMemberS{Value: fmt.Sprintf("Generated large batch item %d of %d", i+1, numItems)},
				},
			},
		}
	}
	
	return requests
}

// Test data factories for common test items
func createTestItem(id, name, itemType string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: id},
		"name": &types.AttributeValueMemberS{Value: name},
		"type": &types.AttributeValueMemberS{Value: itemType},
	}
}

func createTestKey(id string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: id},
	}
}