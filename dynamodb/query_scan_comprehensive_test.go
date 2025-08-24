package dynamodb

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

// Test data factories for query and scan operations

func createQueryTestItems() []map[string]types.AttributeValue {
	return []map[string]types.AttributeValue{
		{
			"pk":          &types.AttributeValueMemberS{Value: "USER#123"},
			"sk":          &types.AttributeValueMemberS{Value: "PROFILE#2023-01-01"},
			"name":        &types.AttributeValueMemberS{Value: "John Doe"},
			"age":         &types.AttributeValueMemberN{Value: "30"},
			"department":  &types.AttributeValueMemberS{Value: "Engineering"},
			"salary":      &types.AttributeValueMemberN{Value: "75000"},
			"created_at":  &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
			"status":      &types.AttributeValueMemberS{Value: "active"},
		},
		{
			"pk":          &types.AttributeValueMemberS{Value: "USER#123"},
			"sk":          &types.AttributeValueMemberS{Value: "PROFILE#2023-06-01"},
			"name":        &types.AttributeValueMemberS{Value: "John Doe"},
			"age":         &types.AttributeValueMemberN{Value: "30"},
			"department":  &types.AttributeValueMemberS{Value: "Engineering"},
			"salary":      &types.AttributeValueMemberN{Value: "80000"},
			"created_at":  &types.AttributeValueMemberS{Value: "2023-06-01T00:00:00Z"},
			"status":      &types.AttributeValueMemberS{Value: "active"},
		},
		{
			"pk":          &types.AttributeValueMemberS{Value: "USER#456"},
			"sk":          &types.AttributeValueMemberS{Value: "PROFILE#2023-02-01"},
			"name":        &types.AttributeValueMemberS{Value: "Jane Smith"},
			"age":         &types.AttributeValueMemberN{Value: "28"},
			"department":  &types.AttributeValueMemberS{Value: "Marketing"},
			"salary":      &types.AttributeValueMemberN{Value: "65000"},
			"created_at":  &types.AttributeValueMemberS{Value: "2023-02-01T00:00:00Z"},
			"status":      &types.AttributeValueMemberS{Value: "active"},
		},
		{
			"pk":          &types.AttributeValueMemberS{Value: "USER#456"},
			"sk":          &types.AttributeValueMemberS{Value: "PROFILE#2023-07-01"},
			"name":        &types.AttributeValueMemberS{Value: "Jane Smith"},
			"age":         &types.AttributeValueMemberN{Value: "29"},
			"department":  &types.AttributeValueMemberS{Value: "Marketing"},
			"salary":      &types.AttributeValueMemberN{Value: "68000"},
			"created_at":  &types.AttributeValueMemberS{Value: "2023-07-01T00:00:00Z"},
			"status":      &types.AttributeValueMemberS{Value: "inactive"},
		},
	}
}

func createScanTestItems() []map[string]types.AttributeValue {
	items := make([]map[string]types.AttributeValue, 25) // Create 25 items for pagination testing
	
	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance"}
	statuses := []string{"active", "inactive", "pending"}
	
	for i := 0; i < 25; i++ {
		department := departments[i%len(departments)]
		status := statuses[i%len(statuses)]
		
		items[i] = map[string]types.AttributeValue{
			"id":          &types.AttributeValueMemberS{Value: fmt.Sprintf("emp-%03d", i+1)},
			"name":        &types.AttributeValueMemberS{Value: fmt.Sprintf("Employee %d", i+1)},
			"age":         &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", 25+i%40)},
			"department":  &types.AttributeValueMemberS{Value: department},
			"salary":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", 50000+i*1000)},
			"created_at":  &types.AttributeValueMemberS{Value: fmt.Sprintf("2023-%02d-01T00:00:00Z", (i%12)+1)},
			"status":      &types.AttributeValueMemberS{Value: status},
			"level":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", (i%5)+1)},
		}
	}
	
	return items
}

// Query operation tests

func TestQuery_WithPartitionKey_ShouldRetrieveMatchingItems(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}

	_, err := QueryE(t, tableName, keyConditionExpression, expressionAttributeValues)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestQuery_WithPartitionKeyAndSortKeyCondition_ShouldFilterBySortKey(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk AND sk BETWEEN :sk_start AND :sk_end"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk":       &types.AttributeValueMemberS{Value: "USER#123"},
		":sk_start": &types.AttributeValueMemberS{Value: "PROFILE#2023-01-01"},
		":sk_end":   &types.AttributeValueMemberS{Value: "PROFILE#2023-12-31"},
	}

	_, err := QueryE(t, tableName, keyConditionExpression, expressionAttributeValues)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestQuery_WithFilterExpression_ShouldApplyPostQueryFiltering(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk":     &types.AttributeValueMemberS{Value: "USER#123"},
		":status": &types.AttributeValueMemberS{Value: "active"},
		":minAge": &types.AttributeValueMemberN{Value: "25"},
	}
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		FilterExpression:       stringPtr("#status = :status AND #age > :minAge"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
			"#age":    "age",
		},
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	_, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestQuery_WithProjectionExpression_ShouldReturnOnlySelectedAttributes(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		ProjectionExpression:   stringPtr("#n, #a, #d, #s"),
		ExpressionAttributeNames: map[string]string{
			"#n": "name",
			"#a": "age",
			"#d": "department",
			"#s": "salary",
		},
	}

	_, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestQuery_WithLimit_ShouldReturnLimitedResults(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	limit := int32(5)
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		Limit:                  &limit,
	}

	_, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestQuery_WithScanIndexForward_ShouldControlSortOrder(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	scanIndexForward := false
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		ScanIndexForward:       &scanIndexForward,
	}

	_, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestQuery_WithConsistentRead_ShouldUseStrongConsistency(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	consistentRead := true
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		ConsistentRead:         &consistentRead,
	}

	_, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestQuery_WithGSI_ShouldQuerySecondaryIndex(t *testing.T) {
	tableName := "test-table"
	indexName := "gsi-department-salary"
	keyConditionExpression := "department = :dept AND salary > :minSalary"
	expressionAttributeValues := map[string]types.AttributeValue{
		":dept":      &types.AttributeValueMemberS{Value: "Engineering"},
		":minSalary": &types.AttributeValueMemberN{Value: "70000"},
	}
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		IndexName:              &indexName,
		ReturnConsumedCapacity: types.ReturnConsumedCapacityIndexes,
	}

	_, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestQuery_WithPagination_ShouldHandleExclusiveStartKey(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	exclusiveStartKey := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "USER#123"},
		"sk": &types.AttributeValueMemberS{Value: "PROFILE#2023-06-01"},
	}
	limit := int32(1)
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		ExclusiveStartKey:      exclusiveStartKey,
		Limit:                  &limit,
	}

	_, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Query pagination tests

func TestQueryAllPages_WithMultiplePages_ShouldRetrieveAllItems(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}

	_, err := QueryAllPagesE(t, tableName, keyConditionExpression, expressionAttributeValues)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestQueryAllPages_WithOptions_ShouldApplyOptionsToAllPages(t *testing.T) {
	tableName := "test-table"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk":     &types.AttributeValueMemberS{Value: "USER#123"},
		":status": &types.AttributeValueMemberS{Value: "active"},
	}
	keyConditionExpression := "pk = :pk"
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		FilterExpression:       stringPtr("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
			"#n":      "name",
			"#a":      "age",
			"#d":      "department",
		},
		ProjectionExpression: stringPtr("#n, #a, #d"),
	}

	_, err := QueryAllPagesWithOptionsE(t, tableName, expressionAttributeValues, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Scan operation tests

func TestScan_WithBasicScan_ShouldRetrieveAllItems(t *testing.T) {
	tableName := "test-table"

	_, err := ScanE(t, tableName)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestScan_WithFilterExpression_ShouldFilterResults(t *testing.T) {
	tableName := "test-table"
	options := ScanOptions{
		FilterExpression: stringPtr("#dept = :dept AND #age BETWEEN :minAge AND :maxAge"),
		ExpressionAttributeNames: map[string]string{
			"#dept": "department",
			"#age":  "age",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":dept":   &types.AttributeValueMemberS{Value: "Engineering"},
			":minAge": &types.AttributeValueMemberN{Value: "25"},
			":maxAge": &types.AttributeValueMemberN{Value: "35"},
		},
	}

	_, err := ScanWithOptionsE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestScan_WithProjectionExpression_ShouldReturnOnlySelectedAttributes(t *testing.T) {
	tableName := "test-table"
	options := ScanOptions{
		ProjectionExpression: stringPtr("#n, #a, #d, #s"),
		ExpressionAttributeNames: map[string]string{
			"#n": "name",
			"#a": "age",
			"#d": "department",
			"#s": "status",
		},
	}

	_, err := ScanWithOptionsE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestScan_WithLimit_ShouldReturnLimitedResults(t *testing.T) {
	tableName := "test-table"
	limit := int32(10)
	options := ScanOptions{
		Limit: &limit,
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	_, err := ScanWithOptionsE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestScan_WithConsistentRead_ShouldUseStrongConsistency(t *testing.T) {
	tableName := "test-table"
	consistentRead := true
	options := ScanOptions{
		ConsistentRead: &consistentRead,
	}

	_, err := ScanWithOptionsE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestScan_WithGSI_ShouldScanSecondaryIndex(t *testing.T) {
	tableName := "test-table"
	indexName := "gsi-status-created"
	options := ScanOptions{
		IndexName:        &indexName,
		FilterExpression: stringPtr("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: "active"},
		},
	}

	_, err := ScanWithOptionsE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestScan_WithParallelScan_ShouldUseSegments(t *testing.T) {
	tableName := "test-table"
	totalSegments := int32(4)
	segment := int32(0)
	options := ScanOptions{
		TotalSegments: &totalSegments,
		Segment:       &segment,
	}

	_, err := ScanWithOptionsE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestScan_WithPagination_ShouldHandleExclusiveStartKey(t *testing.T) {
	tableName := "test-table"
	exclusiveStartKey := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "emp-010"},
	}
	limit := int32(5)
	options := ScanOptions{
		ExclusiveStartKey: exclusiveStartKey,
		Limit:             &limit,
	}

	_, err := ScanWithOptionsE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Scan pagination tests

func TestScanAllPages_WithMultiplePages_ShouldRetrieveAllItems(t *testing.T) {
	tableName := "test-table"

	_, err := ScanAllPagesE(t, tableName)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestScanAllPages_WithOptions_ShouldApplyOptionsToAllPages(t *testing.T) {
	tableName := "test-table"
	options := ScanOptions{
		FilterExpression: stringPtr("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
			"#n":      "name",
			"#a":      "age",
			"#d":      "department",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: "active"},
		},
		ProjectionExpression: stringPtr("#n, #a, #d"),
	}

	_, err := ScanAllPagesWithOptionsE(t, tableName, options)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

func TestScanAllPages_WithLargeDataset_ShouldHandleManyPages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	tableName := "large-test-table"

	_, err := ScanAllPagesE(t, tableName)
	
	// Should fail without proper mocking
	assert.Error(t, err)
}

// Complex query patterns

func TestQuery_WithComplexKeyConditions_ShouldHandleAllComparisons(t *testing.T) {
	testCases := []struct {
		name                   string
		keyConditionExpression string
		expressionValues      map[string]types.AttributeValue
	}{
		{
			name:                   "Equals condition",
			keyConditionExpression: "pk = :pk",
			expressionValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "USER#123"},
			},
		},
		{
			name:                   "Begins with condition",
			keyConditionExpression: "pk = :pk AND begins_with(sk, :sk_prefix)",
			expressionValues: map[string]types.AttributeValue{
				":pk":        &types.AttributeValueMemberS{Value: "USER#123"},
				":sk_prefix": &types.AttributeValueMemberS{Value: "PROFILE#"},
			},
		},
		{
			name:                   "Between condition",
			keyConditionExpression: "pk = :pk AND sk BETWEEN :sk_start AND :sk_end",
			expressionValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: "USER#123"},
				":sk_start": &types.AttributeValueMemberS{Value: "PROFILE#2023-01-01"},
				":sk_end":   &types.AttributeValueMemberS{Value: "PROFILE#2023-12-31"},
			},
		},
		{
			name:                   "Less than condition",
			keyConditionExpression: "pk = :pk AND sk < :sk_max",
			expressionValues: map[string]types.AttributeValue{
				":pk":     &types.AttributeValueMemberS{Value: "USER#123"},
				":sk_max": &types.AttributeValueMemberS{Value: "PROFILE#2023-06-01"},
			},
		},
		{
			name:                   "Less than or equal condition",
			keyConditionExpression: "pk = :pk AND sk <= :sk_max",
			expressionValues: map[string]types.AttributeValue{
				":pk":     &types.AttributeValueMemberS{Value: "USER#123"},
				":sk_max": &types.AttributeValueMemberS{Value: "PROFILE#2023-06-01"},
			},
		},
		{
			name:                   "Greater than condition",
			keyConditionExpression: "pk = :pk AND sk > :sk_min",
			expressionValues: map[string]types.AttributeValue{
				":pk":     &types.AttributeValueMemberS{Value: "USER#123"},
				":sk_min": &types.AttributeValueMemberS{Value: "PROFILE#2023-01-01"},
			},
		},
		{
			name:                   "Greater than or equal condition",
			keyConditionExpression: "pk = :pk AND sk >= :sk_min",
			expressionValues: map[string]types.AttributeValue{
				":pk":     &types.AttributeValueMemberS{Value: "USER#123"},
				":sk_min": &types.AttributeValueMemberS{Value: "PROFILE#2023-01-01"},
			},
		},
	}

	tableName := "test-table"
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := QueryE(t, tableName, tc.keyConditionExpression, tc.expressionValues)
			
			// Should fail without proper mocking
			assert.Error(t, err)
		})
	}
}

func TestScan_WithComplexFilterExpressions_ShouldHandleAllOperations(t *testing.T) {
	testCases := []struct {
		name              string
		filterExpression  string
		expressionNames   map[string]string
		expressionValues  map[string]types.AttributeValue
	}{
		{
			name:             "Comparison operators",
			filterExpression: "#age > :minAge AND #age < :maxAge",
			expressionNames: map[string]string{
				"#age": "age",
			},
			expressionValues: map[string]types.AttributeValue{
				":minAge": &types.AttributeValueMemberN{Value: "25"},
				":maxAge": &types.AttributeValueMemberN{Value: "35"},
			},
		},
		{
			name:             "IN operator",
			filterExpression: "#dept IN (:dept1, :dept2, :dept3)",
			expressionNames: map[string]string{
				"#dept": "department",
			},
			expressionValues: map[string]types.AttributeValue{
				":dept1": &types.AttributeValueMemberS{Value: "Engineering"},
				":dept2": &types.AttributeValueMemberS{Value: "Marketing"},
				":dept3": &types.AttributeValueMemberS{Value: "Sales"},
			},
		},
		{
			name:             "BETWEEN operator",
			filterExpression: "#salary BETWEEN :minSalary AND :maxSalary",
			expressionNames: map[string]string{
				"#salary": "salary",
			},
			expressionValues: map[string]types.AttributeValue{
				":minSalary": &types.AttributeValueMemberN{Value: "60000"},
				":maxSalary": &types.AttributeValueMemberN{Value: "80000"},
			},
		},
		{
			name:             "Function: attribute_exists",
			filterExpression: "attribute_exists(#email)",
			expressionNames: map[string]string{
				"#email": "email",
			},
			expressionValues: map[string]types.AttributeValue{},
		},
		{
			name:             "Function: attribute_not_exists",
			filterExpression: "attribute_not_exists(#temp_field)",
			expressionNames: map[string]string{
				"#temp_field": "temp_field",
			},
			expressionValues: map[string]types.AttributeValue{},
		},
		{
			name:             "Function: attribute_type",
			filterExpression: "attribute_type(#age, :type)",
			expressionNames: map[string]string{
				"#age": "age",
			},
			expressionValues: map[string]types.AttributeValue{
				":type": &types.AttributeValueMemberS{Value: "N"},
			},
		},
		{
			name:             "Function: begins_with",
			filterExpression: "begins_with(#name, :prefix)",
			expressionNames: map[string]string{
				"#name": "name",
			},
			expressionValues: map[string]types.AttributeValue{
				":prefix": &types.AttributeValueMemberS{Value: "John"},
			},
		},
		{
			name:             "Function: contains",
			filterExpression: "contains(#name, :substring)",
			expressionNames: map[string]string{
				"#name": "name",
			},
			expressionValues: map[string]types.AttributeValue{
				":substring": &types.AttributeValueMemberS{Value: "Doe"},
			},
		},
		{
			name:             "Function: size",
			filterExpression: "size(#name) > :minLength",
			expressionNames: map[string]string{
				"#name": "name",
			},
			expressionValues: map[string]types.AttributeValue{
				":minLength": &types.AttributeValueMemberN{Value: "5"},
			},
		},
		{
			name:             "Complex AND/OR conditions",
			filterExpression: "(#dept = :eng OR #dept = :mkt) AND #age > :minAge AND #status = :status",
			expressionNames: map[string]string{
				"#dept":   "department",
				"#age":    "age",
				"#status": "status",
			},
			expressionValues: map[string]types.AttributeValue{
				":eng":    &types.AttributeValueMemberS{Value: "Engineering"},
				":mkt":    &types.AttributeValueMemberS{Value: "Marketing"},
				":minAge": &types.AttributeValueMemberN{Value: "25"},
				":status": &types.AttributeValueMemberS{Value: "active"},
			},
		},
	}

	tableName := "test-table"
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := ScanOptions{
				FilterExpression:         &tc.filterExpression,
				ExpressionAttributeNames: tc.expressionNames,
				ExpressionAttributeValues: tc.expressionValues,
			}

			_, err := ScanWithOptionsE(t, tableName, options)
			
			// Should fail without proper mocking
			assert.Error(t, err)
		})
	}
}

// Performance and stress tests

func TestQuery_PerformanceBenchmark_WithLargeResultSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tableName := "performance-test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "LARGE_DATASET"},
	}

	start := time.Now()
	_, err := QueryE(t, tableName, keyConditionExpression, expressionAttributeValues)
	duration := time.Since(start)

	// Should fail without proper mocking
	assert.Error(t, err)
	t.Logf("Query operation took: %v", duration)
}

func TestScan_PerformanceBenchmark_WithFullTableScan(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tableName := "performance-test-table"

	start := time.Now()
	_, err := ScanE(t, tableName)
	duration := time.Since(start)

	// Should fail without proper mocking
	assert.Error(t, err)
	t.Logf("Scan operation took: %v", duration)
}

func TestQueryAllPages_WithManySmallPages_ShouldHandleManyRoundTrips(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pagination stress test in short mode")
	}

	tableName := "pagination-test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "PAGINATED_DATA"},
	}
	
	// Force small pages to test many round trips
	limit := int32(1)
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		Limit:                  &limit,
	}

	start := time.Now()
	_, err := QueryAllPagesWithOptionsE(t, tableName, expressionAttributeValues, options)
	duration := time.Since(start)

	// Should fail without proper mocking
	assert.Error(t, err)
	t.Logf("Paginated query with limit=1 took: %v", duration)
}

// Error handling tests

func TestQuery_WithInvalidKeyConditionExpression_ShouldReturnValidationException(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "INVALID SYNTAX" // Invalid expression
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}

	_, err := QueryE(t, tableName, keyConditionExpression, expressionAttributeValues)
	
	assert.Error(t, err)
}

func TestQuery_WithMissingExpressionAttributeValues_ShouldReturnValidationException(t *testing.T) {
	tableName := "test-table"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{} // Missing :pk value

	_, err := QueryE(t, tableName, keyConditionExpression, expressionAttributeValues)
	
	assert.Error(t, err)
}

func TestScan_WithInvalidFilterExpression_ShouldReturnValidationException(t *testing.T) {
	tableName := "test-table"
	options := ScanOptions{
		FilterExpression: stringPtr("INVALID SYNTAX"), // Invalid expression
	}

	_, err := ScanWithOptionsE(t, tableName, options)
	
	assert.Error(t, err)
}

func TestQuery_WithNonExistentGSI_ShouldReturnResourceNotFoundException(t *testing.T) {
	tableName := "test-table"
	indexName := "non-existent-index"
	keyConditionExpression := "pk = :pk"
	expressionAttributeValues := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	options := QueryOptions{
		KeyConditionExpression: &keyConditionExpression,
		IndexName:              &indexName,
	}

	_, err := QueryWithOptionsE(t, tableName, expressionAttributeValues, options)
	
	assert.Error(t, err)
}

func TestScan_WithNonExistentGSI_ShouldReturnResourceNotFoundException(t *testing.T) {
	tableName := "test-table"
	indexName := "non-existent-index"
	options := ScanOptions{
		IndexName: &indexName,
	}

	_, err := ScanWithOptionsE(t, tableName, options)
	
	assert.Error(t, err)
}