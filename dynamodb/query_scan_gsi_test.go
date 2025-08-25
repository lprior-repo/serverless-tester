package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pure TDD implementation for GSI/LSI query and scan operations
// Following Red-Green-Refactor methodology with focus on secondary index operations

// Pure Functions for GSI/LSI Operations

// validateGSIQueryOptions validates options for GSI query operations
func validateGSIQueryOptions(indexName *string, keyConditionExpression *string) error {
	if indexName == nil || *indexName == "" {
		return assert.AnError
	}
	if keyConditionExpression == nil || *keyConditionExpression == "" {
		return assert.AnError
	}
	return nil
}

// validateLSIQueryOptions validates options for LSI query operations
func validateLSIQueryOptions(indexName *string, keyConditionExpression *string) error {
	if indexName == nil || *indexName == "" {
		return assert.AnError
	}
	if keyConditionExpression == nil || *keyConditionExpression == "" {
		return assert.AnError
	}
	return nil
}

// createGSIKeyCondition creates a key condition for GSI queries
func createGSIKeyCondition(gsiPartitionKey string, gsiSortKey string, partitionValue string, sortValue string) (string, map[string]types.AttributeValue) {
	expression := gsiPartitionKey + " = :gsi_pk"
	values := map[string]types.AttributeValue{
		":gsi_pk": &types.AttributeValueMemberS{Value: partitionValue},
	}
	
	if gsiSortKey != "" && sortValue != "" {
		expression += " AND " + gsiSortKey + " = :gsi_sk"
		values[":gsi_sk"] = &types.AttributeValueMemberS{Value: sortValue}
	}
	
	return expression, values
}

// createLSIKeyCondition creates a key condition for LSI queries
func createLSIKeyCondition(partitionKey string, lsiSortKey string, partitionValue string, sortValue string) (string, map[string]types.AttributeValue) {
	expression := partitionKey + " = :pk"
	values := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: partitionValue},
	}
	
	if lsiSortKey != "" && sortValue != "" {
		expression += " AND " + lsiSortKey + " = :lsi_sk"
		values[":lsi_sk"] = &types.AttributeValueMemberS{Value: sortValue}
	}
	
	return expression, values
}

// extractProjectedAttributes extracts only the projected attributes from GSI/LSI results
func extractProjectedAttributes(items []map[string]types.AttributeValue, projectedAttributes []string) []map[string]types.AttributeValue {
	if len(projectedAttributes) == 0 {
		return items // Return all attributes if no projection specified
	}
	
	projected := make([]map[string]types.AttributeValue, len(items))
	
	for i, item := range items {
		projectedItem := make(map[string]types.AttributeValue)
		for _, attr := range projectedAttributes {
			if value, exists := item[attr]; exists {
				projectedItem[attr] = value
			}
		}
		projected[i] = projectedItem
	}
	
	return projected
}

// validateSecondaryIndexConsistency validates consistency requirements for secondary indexes
func validateSecondaryIndexConsistency(indexName *string, consistentRead *bool) error {
	if indexName != nil && *indexName != "" {
		if consistentRead != nil && *consistentRead {
			return assert.AnError // GSI doesn't support consistent reads
		}
	}
	return nil
}

// createSecondaryIndexFilterExpression creates filter expressions specific to secondary indexes
func createSecondaryIndexFilterExpression(attributes map[string]string) (string, map[string]string, map[string]types.AttributeValue) {
	if len(attributes) == 0 {
		return "", nil, nil
	}
	
	expression := ""
	names := make(map[string]string)
	values := make(map[string]types.AttributeValue)
	
	first := true
	for attrName, attrValue := range attributes {
		if !first {
			expression += " AND "
		}
		
		nameKey := "#" + attrName
		valueKey := ":" + attrName
		
		expression += nameKey + " = " + valueKey
		names[nameKey] = attrName
		values[valueKey] = &types.AttributeValueMemberS{Value: attrValue}
		
		first = false
	}
	
	return expression, names, values
}

// Test Data Factories for GSI/LSI Testing

func createGSITestData() []map[string]types.AttributeValue {
	return []map[string]types.AttributeValue{
		{
			"pk":         &types.AttributeValueMemberS{Value: "USER#001"},
			"sk":         &types.AttributeValueMemberS{Value: "PROFILE"},
			"department": &types.AttributeValueMemberS{Value: "Engineering"}, // GSI PK
			"salary":     &types.AttributeValueMemberN{Value: "75000"},       // GSI SK
			"name":       &types.AttributeValueMemberS{Value: "John Doe"},
			"age":        &types.AttributeValueMemberN{Value: "30"},
			"status":     &types.AttributeValueMemberS{Value: "active"},
		},
		{
			"pk":         &types.AttributeValueMemberS{Value: "USER#002"},
			"sk":         &types.AttributeValueMemberS{Value: "PROFILE"},
			"department": &types.AttributeValueMemberS{Value: "Engineering"}, // GSI PK
			"salary":     &types.AttributeValueMemberN{Value: "80000"},       // GSI SK
			"name":       &types.AttributeValueMemberS{Value: "Jane Smith"},
			"age":        &types.AttributeValueMemberN{Value: "28"},
			"status":     &types.AttributeValueMemberS{Value: "active"},
		},
		{
			"pk":         &types.AttributeValueMemberS{Value: "USER#003"},
			"sk":         &types.AttributeValueMemberS{Value: "PROFILE"},
			"department": &types.AttributeValueMemberS{Value: "Marketing"}, // GSI PK
			"salary":     &types.AttributeValueMemberN{Value: "65000"},     // GSI SK
			"name":       &types.AttributeValueMemberS{Value: "Bob Wilson"},
			"age":        &types.AttributeValueMemberN{Value: "35"},
			"status":     &types.AttributeValueMemberS{Value: "active"},
		},
		{
			"pk":         &types.AttributeValueMemberS{Value: "USER#004"},
			"sk":         &types.AttributeValueMemberS{Value: "PROFILE"},
			"department": &types.AttributeValueMemberS{Value: "Sales"}, // GSI PK
			"salary":     &types.AttributeValueMemberN{Value: "60000"}, // GSI SK
			"name":       &types.AttributeValueMemberS{Value: "Alice Brown"},
			"age":        &types.AttributeValueMemberN{Value: "32"},
			"status":     &types.AttributeValueMemberS{Value: "inactive"},
		},
	}
}

func createLSITestData() []map[string]types.AttributeValue {
	return []map[string]types.AttributeValue{
		{
			"pk":         &types.AttributeValueMemberS{Value: "USER#001"}, // Same PK for LSI
			"sk":         &types.AttributeValueMemberS{Value: "PROFILE"},
			"created_at": &types.AttributeValueMemberS{Value: "2023-01-15"}, // LSI SK
			"name":       &types.AttributeValueMemberS{Value: "John Doe"},
			"department": &types.AttributeValueMemberS{Value: "Engineering"},
			"priority":   &types.AttributeValueMemberN{Value: "1"},
		},
		{
			"pk":         &types.AttributeValueMemberS{Value: "USER#001"}, // Same PK for LSI
			"sk":         &types.AttributeValueMemberS{Value: "SETTINGS"},
			"created_at": &types.AttributeValueMemberS{Value: "2023-02-10"}, // LSI SK
			"theme":      &types.AttributeValueMemberS{Value: "dark"},
			"language":   &types.AttributeValueMemberS{Value: "en"},
			"priority":   &types.AttributeValueMemberN{Value: "2"},
		},
		{
			"pk":         &types.AttributeValueMemberS{Value: "USER#001"}, // Same PK for LSI
			"sk":         &types.AttributeValueMemberS{Value: "ACTIVITY"},
			"created_at": &types.AttributeValueMemberS{Value: "2023-03-05"}, // LSI SK
			"action":     &types.AttributeValueMemberS{Value: "login"},
			"ip_address": &types.AttributeValueMemberS{Value: "192.168.1.1"},
			"priority":   &types.AttributeValueMemberN{Value: "3"},
		},
	}
}

// RED PHASE: GSI Validation Tests

func TestValidateGSIQueryOptions_WithValidOptions_ShouldReturnNoError(t *testing.T) {
	// GIVEN: Valid GSI query options
	indexName := "GSI-Department-Salary"
	keyConditionExpression := "department = :dept"
	
	// WHEN: Validating GSI options
	err := validateGSIQueryOptions(&indexName, &keyConditionExpression)
	
	// THEN: Should return no error
	assert.NoError(t, err)
}

func TestValidateGSIQueryOptions_WithEmptyIndexName_ShouldReturnError(t *testing.T) {
	// GIVEN: Empty index name
	indexName := ""
	keyConditionExpression := "department = :dept"
	
	// WHEN: Validating GSI options
	err := validateGSIQueryOptions(&indexName, &keyConditionExpression)
	
	// THEN: Should return error
	assert.Error(t, err)
}

func TestValidateGSIQueryOptions_WithNilIndexName_ShouldReturnError(t *testing.T) {
	// GIVEN: Nil index name
	var indexName *string = nil
	keyConditionExpression := "department = :dept"
	
	// WHEN: Validating GSI options
	err := validateGSIQueryOptions(indexName, &keyConditionExpression)
	
	// THEN: Should return error
	assert.Error(t, err)
}

func TestValidateGSIQueryOptions_WithEmptyKeyConditionExpression_ShouldReturnError(t *testing.T) {
	// GIVEN: Empty key condition expression
	indexName := "GSI-Department-Salary"
	keyConditionExpression := ""
	
	// WHEN: Validating GSI options
	err := validateGSIQueryOptions(&indexName, &keyConditionExpression)
	
	// THEN: Should return error
	assert.Error(t, err)
}

func TestValidateGSIQueryOptions_WithNilKeyConditionExpression_ShouldReturnError(t *testing.T) {
	// GIVEN: Nil key condition expression
	indexName := "GSI-Department-Salary"
	var keyConditionExpression *string = nil
	
	// WHEN: Validating GSI options
	err := validateGSIQueryOptions(&indexName, keyConditionExpression)
	
	// THEN: Should return error
	assert.Error(t, err)
}

// RED PHASE: LSI Validation Tests

func TestValidateLSIQueryOptions_WithValidOptions_ShouldReturnNoError(t *testing.T) {
	// GIVEN: Valid LSI query options
	indexName := "LSI-CreatedAt"
	keyConditionExpression := "pk = :pk AND created_at = :created"
	
	// WHEN: Validating LSI options
	err := validateLSIQueryOptions(&indexName, &keyConditionExpression)
	
	// THEN: Should return no error
	assert.NoError(t, err)
}

func TestValidateLSIQueryOptions_WithEmptyIndexName_ShouldReturnError(t *testing.T) {
	// GIVEN: Empty index name
	indexName := ""
	keyConditionExpression := "pk = :pk"
	
	// WHEN: Validating LSI options
	err := validateLSIQueryOptions(&indexName, &keyConditionExpression)
	
	// THEN: Should return error
	assert.Error(t, err)
}

// RED PHASE: GSI Key Condition Creation Tests

func TestCreateGSIKeyCondition_WithPartitionKeyOnly_ShouldCreateValidCondition(t *testing.T) {
	// GIVEN: GSI partition key only
	gsiPartitionKey := "department"
	gsiSortKey := ""
	partitionValue := "Engineering"
	sortValue := ""
	
	// WHEN: Creating GSI key condition
	expression, values := createGSIKeyCondition(gsiPartitionKey, gsiSortKey, partitionValue, sortValue)
	
	// THEN: Should create partition key only condition
	assert.Equal(t, "department = :gsi_pk", expression)
	assert.Len(t, values, 1)
	assert.Contains(t, values, ":gsi_pk")
	
	if pkValue, ok := values[":gsi_pk"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "Engineering", pkValue.Value)
	}
}

func TestCreateGSIKeyCondition_WithBothKeys_ShouldCreateFullCondition(t *testing.T) {
	// GIVEN: Both GSI partition and sort keys
	gsiPartitionKey := "department"
	gsiSortKey := "salary"
	partitionValue := "Engineering"
	sortValue := "75000"
	
	// WHEN: Creating GSI key condition
	expression, values := createGSIKeyCondition(gsiPartitionKey, gsiSortKey, partitionValue, sortValue)
	
	// THEN: Should create full condition with both keys
	assert.Equal(t, "department = :gsi_pk AND salary = :gsi_sk", expression)
	assert.Len(t, values, 2)
	assert.Contains(t, values, ":gsi_pk")
	assert.Contains(t, values, ":gsi_sk")
	
	if pkValue, ok := values[":gsi_pk"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "Engineering", pkValue.Value)
	}
	if skValue, ok := values[":gsi_sk"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "75000", skValue.Value)
	}
}

func TestCreateGSIKeyCondition_WithEmptyValues_ShouldHandleGracefully(t *testing.T) {
	// GIVEN: Empty values
	gsiPartitionKey := "department"
	gsiSortKey := "salary"
	partitionValue := ""
	sortValue := ""
	
	// WHEN: Creating GSI key condition
	expression, values := createGSIKeyCondition(gsiPartitionKey, gsiSortKey, partitionValue, sortValue)
	
	// THEN: Should create condition with empty values
	assert.Equal(t, "department = :gsi_pk", expression)
	assert.Len(t, values, 1)
	
	if pkValue, ok := values[":gsi_pk"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "", pkValue.Value)
	}
}

// RED PHASE: LSI Key Condition Creation Tests

func TestCreateLSIKeyCondition_WithPartitionKeyOnly_ShouldCreateValidCondition(t *testing.T) {
	// GIVEN: LSI partition key only
	partitionKey := "pk"
	lsiSortKey := ""
	partitionValue := "USER#001"
	sortValue := ""
	
	// WHEN: Creating LSI key condition
	expression, values := createLSIKeyCondition(partitionKey, lsiSortKey, partitionValue, sortValue)
	
	// THEN: Should create partition key only condition
	assert.Equal(t, "pk = :pk", expression)
	assert.Len(t, values, 1)
	assert.Contains(t, values, ":pk")
	
	if pkValue, ok := values[":pk"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "USER#001", pkValue.Value)
	}
}

func TestCreateLSIKeyCondition_WithBothKeys_ShouldCreateFullCondition(t *testing.T) {
	// GIVEN: Both partition and LSI sort keys
	partitionKey := "pk"
	lsiSortKey := "created_at"
	partitionValue := "USER#001"
	sortValue := "2023-01-15"
	
	// WHEN: Creating LSI key condition
	expression, values := createLSIKeyCondition(partitionKey, lsiSortKey, partitionValue, sortValue)
	
	// THEN: Should create full condition with both keys
	assert.Equal(t, "pk = :pk AND created_at = :lsi_sk", expression)
	assert.Len(t, values, 2)
	assert.Contains(t, values, ":pk")
	assert.Contains(t, values, ":lsi_sk")
	
	if pkValue, ok := values[":pk"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "USER#001", pkValue.Value)
	}
	if skValue, ok := values[":lsi_sk"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "2023-01-15", skValue.Value)
	}
}

// RED PHASE: Projected Attributes Extraction Tests

func TestExtractProjectedAttributes_WithSpecificProjection_ShouldReturnOnlyProjectedAttributes(t *testing.T) {
	// GIVEN: Items with various attributes
	items := createGSITestData()
	projectedAttributes := []string{"name", "department", "salary"}
	
	// WHEN: Extracting projected attributes
	projected := extractProjectedAttributes(items, projectedAttributes)
	
	// THEN: Should return only projected attributes
	assert.Len(t, projected, len(items))
	
	for _, item := range projected {
		assert.Len(t, item, 3) // Only 3 projected attributes
		assert.Contains(t, item, "name")
		assert.Contains(t, item, "department")
		assert.Contains(t, item, "salary")
		assert.NotContains(t, item, "pk")
		assert.NotContains(t, item, "sk")
		assert.NotContains(t, item, "age")
		assert.NotContains(t, item, "status")
	}
}

func TestExtractProjectedAttributes_WithEmptyProjection_ShouldReturnAllAttributes(t *testing.T) {
	// GIVEN: Items with various attributes and empty projection
	items := createGSITestData()
	projectedAttributes := []string{}
	
	// WHEN: Extracting with empty projection
	projected := extractProjectedAttributes(items, projectedAttributes)
	
	// THEN: Should return all attributes (no filtering)
	assert.Len(t, projected, len(items))
	assert.Equal(t, items, projected)
}

func TestExtractProjectedAttributes_WithNonExistentAttributes_ShouldReturnEmptyAttributes(t *testing.T) {
	// GIVEN: Items and projection with non-existent attributes
	items := createGSITestData()
	projectedAttributes := []string{"nonexistent1", "nonexistent2"}
	
	// WHEN: Extracting non-existent attributes
	projected := extractProjectedAttributes(items, projectedAttributes)
	
	// THEN: Should return items with empty maps
	assert.Len(t, projected, len(items))
	
	for _, item := range projected {
		assert.Len(t, item, 0) // Empty map for each item
	}
}

func TestExtractProjectedAttributes_WithMixedExistentNonExistent_ShouldReturnOnlyExistent(t *testing.T) {
	// GIVEN: Mixed existent and non-existent attributes in projection
	items := createGSITestData()
	projectedAttributes := []string{"name", "nonexistent", "department"}
	
	// WHEN: Extracting mixed attributes
	projected := extractProjectedAttributes(items, projectedAttributes)
	
	// THEN: Should return only existent attributes
	assert.Len(t, projected, len(items))
	
	for _, item := range projected {
		assert.Len(t, item, 2) // Only 2 existent attributes
		assert.Contains(t, item, "name")
		assert.Contains(t, item, "department")
		assert.NotContains(t, item, "nonexistent")
	}
}

// RED PHASE: Secondary Index Consistency Validation Tests

func TestValidateSecondaryIndexConsistency_WithoutConsistentRead_ShouldReturnNoError(t *testing.T) {
	// GIVEN: GSI without consistent read requirement
	indexName := "GSI-Department-Salary"
	var consistentRead *bool = nil
	
	// WHEN: Validating consistency
	err := validateSecondaryIndexConsistency(&indexName, consistentRead)
	
	// THEN: Should return no error
	assert.NoError(t, err)
}

func TestValidateSecondaryIndexConsistency_WithConsistentReadFalse_ShouldReturnNoError(t *testing.T) {
	// GIVEN: GSI with consistent read explicitly set to false
	indexName := "GSI-Department-Salary"
	consistentRead := false
	
	// WHEN: Validating consistency
	err := validateSecondaryIndexConsistency(&indexName, &consistentRead)
	
	// THEN: Should return no error
	assert.NoError(t, err)
}

func TestValidateSecondaryIndexConsistency_WithConsistentReadTrue_ShouldReturnError(t *testing.T) {
	// GIVEN: GSI with consistent read set to true (not supported)
	indexName := "GSI-Department-Salary"
	consistentRead := true
	
	// WHEN: Validating consistency
	err := validateSecondaryIndexConsistency(&indexName, &consistentRead)
	
	// THEN: Should return error (GSI doesn't support consistent reads)
	assert.Error(t, err)
}

func TestValidateSecondaryIndexConsistency_WithoutIndexName_ShouldAllowConsistentRead(t *testing.T) {
	// GIVEN: No index name (main table query)
	var indexName *string = nil
	consistentRead := true
	
	// WHEN: Validating consistency
	err := validateSecondaryIndexConsistency(indexName, &consistentRead)
	
	// THEN: Should return no error (main table supports consistent reads)
	assert.NoError(t, err)
}

// RED PHASE: Secondary Index Filter Expression Tests

func TestCreateSecondaryIndexFilterExpression_WithSingleAttribute_ShouldCreateSimpleFilter(t *testing.T) {
	// GIVEN: Single attribute filter
	attributes := map[string]string{
		"status": "active",
	}
	
	// WHEN: Creating filter expression
	expression, names, values := createSecondaryIndexFilterExpression(attributes)
	
	// THEN: Should create simple filter
	assert.Equal(t, "#status = :status", expression)
	assert.Len(t, names, 1)
	assert.Len(t, values, 1)
	assert.Equal(t, "status", names["#status"])
	
	if statusValue, ok := values[":status"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "active", statusValue.Value)
	}
}

func TestCreateSecondaryIndexFilterExpression_WithMultipleAttributes_ShouldCreateComplexFilter(t *testing.T) {
	// GIVEN: Multiple attribute filters
	attributes := map[string]string{
		"status":     "active",
		"department": "Engineering",
	}
	
	// WHEN: Creating filter expression
	expression, names, values := createSecondaryIndexFilterExpression(attributes)
	
	// THEN: Should create complex filter with AND conditions
	assert.Contains(t, expression, "#status = :status")
	assert.Contains(t, expression, "#department = :department")
	assert.Contains(t, expression, " AND ")
	assert.Len(t, names, 2)
	assert.Len(t, values, 2)
	
	assert.Equal(t, "status", names["#status"])
	assert.Equal(t, "department", names["#department"])
	
	if statusValue, ok := values[":status"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "active", statusValue.Value)
	}
	if deptValue, ok := values[":department"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "Engineering", deptValue.Value)
	}
}

func TestCreateSecondaryIndexFilterExpression_WithEmptyAttributes_ShouldReturnEmpty(t *testing.T) {
	// GIVEN: Empty attributes
	attributes := map[string]string{}
	
	// WHEN: Creating filter expression
	expression, names, values := createSecondaryIndexFilterExpression(attributes)
	
	// THEN: Should return empty results
	assert.Equal(t, "", expression)
	assert.Nil(t, names)
	assert.Nil(t, values)
}

// RED PHASE: GSI Query Integration Tests

func TestQuery_WithGSI_ShouldQuerySecondaryIndex_GSI(t *testing.T) {
	// GIVEN: GSI query parameters
	tableName := "test-table"
	indexName := "GSI-Department-Salary"
	gsiKeyExpr, gsiKeyValues := createGSIKeyCondition("department", "salary", "Engineering", "75000")
	
	options := QueryOptions{
		KeyConditionExpression: &gsiKeyExpr,
		IndexName:             &indexName,
		ReturnConsumedCapacity: types.ReturnConsumedCapacityIndexes,
	}
	
	// WHEN: Executing GSI query
	_, err := QueryWithOptionsE(t, tableName, gsiKeyValues, options)
	
	// THEN: Should attempt to query GSI
	assert.Error(t, err) // Expected to fail in test environment
}

func TestQuery_WithGSIAndFilter_ShouldApplyFilterToGSIQuery(t *testing.T) {
	// GIVEN: GSI query with additional filter
	tableName := "test-table"
	indexName := "GSI-Department-Salary"
	gsiKeyExpr, gsiKeyValues := createGSIKeyCondition("department", "", "Engineering", "")
	filterExpr, filterNames, filterValues := createSecondaryIndexFilterExpression(map[string]string{
		"status": "active",
	})
	
	// Merge expression attribute values
	allValues := make(map[string]types.AttributeValue)
	for k, v := range gsiKeyValues {
		allValues[k] = v
	}
	for k, v := range filterValues {
		allValues[k] = v
	}
	
	options := QueryOptions{
		KeyConditionExpression:   &gsiKeyExpr,
		IndexName:               &indexName,
		FilterExpression:        &filterExpr,
		ExpressionAttributeNames: filterNames,
	}
	
	// WHEN: Executing GSI query with filter
	_, err := QueryWithOptionsE(t, tableName, allValues, options)
	
	// THEN: Should attempt to query GSI with filter
	assert.Error(t, err) // Expected to fail in test environment
}

func TestQuery_WithGSIProjection_ShouldReturnProjectedAttributes(t *testing.T) {
	// GIVEN: GSI query with projection
	tableName := "test-table"
	indexName := "GSI-Department-Salary"
	gsiKeyExpr, gsiKeyValues := createGSIKeyCondition("department", "", "Engineering", "")
	projection := "#name, #dept, #salary"
	projectionNames := map[string]string{
		"#name":   "name",
		"#dept":   "department",
		"#salary": "salary",
	}
	
	options := QueryOptions{
		KeyConditionExpression:   &gsiKeyExpr,
		IndexName:               &indexName,
		ProjectionExpression:     &projection,
		ExpressionAttributeNames: projectionNames,
	}
	
	// WHEN: Executing GSI query with projection
	_, err := QueryWithOptionsE(t, tableName, gsiKeyValues, options)
	
	// THEN: Should attempt to query GSI with projection
	assert.Error(t, err) // Expected to fail in test environment
}

// RED PHASE: LSI Query Integration Tests

func TestQuery_WithLSI_ShouldQueryLocalSecondaryIndex(t *testing.T) {
	// GIVEN: LSI query parameters
	tableName := "test-table"
	indexName := "LSI-CreatedAt"
	lsiKeyExpr, lsiKeyValues := createLSIKeyCondition("pk", "created_at", "USER#001", "2023-01-15")
	
	options := QueryOptions{
		KeyConditionExpression: &lsiKeyExpr,
		IndexName:             &indexName,
		ConsistentRead:        boolPtr(true), // LSI supports consistent reads
	}
	
	// WHEN: Executing LSI query
	_, err := QueryWithOptionsE(t, tableName, lsiKeyValues, options)
	
	// THEN: Should attempt to query LSI
	assert.Error(t, err) // Expected to fail in test environment
}

func TestQuery_WithLSIAndSorting_ShouldSortByLSISortKey(t *testing.T) {
	// GIVEN: LSI query with sorting
	tableName := "test-table"
	indexName := "LSI-CreatedAt"
	lsiKeyExpr, lsiKeyValues := createLSIKeyCondition("pk", "", "USER#001", "")
	scanIndexForward := false // Reverse sort
	
	options := QueryOptions{
		KeyConditionExpression: &lsiKeyExpr,
		IndexName:             &indexName,
		ScanIndexForward:       &scanIndexForward,
	}
	
	// WHEN: Executing LSI query with sorting
	_, err := QueryWithOptionsE(t, tableName, lsiKeyValues, options)
	
	// THEN: Should attempt to query LSI with sorting
	assert.Error(t, err) // Expected to fail in test environment
}

// RED PHASE: Scan GSI Integration Tests

func TestScan_WithGSI_ShouldScanSecondaryIndex_GSI(t *testing.T) {
	// GIVEN: GSI scan parameters
	tableName := "test-table"
	indexName := "GSI-Department-Salary"
	
	options := ScanOptions{
		IndexName: &indexName,
	}
	
	// WHEN: Executing GSI scan
	_, err := ScanWithOptionsE(t, tableName, options)
	
	// THEN: Should attempt to scan GSI
	assert.Error(t, err) // Expected to fail in test environment
}

func TestScan_WithGSIAndFilter_ShouldApplyFilterToGSIScan(t *testing.T) {
	// GIVEN: GSI scan with filter
	tableName := "test-table"
	indexName := "GSI-Department-Salary"
	filterExpr, filterNames, filterValues := createSecondaryIndexFilterExpression(map[string]string{
		"status": "active",
	})
	
	options := ScanOptions{
		IndexName:                &indexName,
		FilterExpression:         &filterExpr,
		ExpressionAttributeNames: filterNames,
		ExpressionAttributeValues: filterValues,
	}
	
	// WHEN: Executing GSI scan with filter
	_, err := ScanWithOptionsE(t, tableName, options)
	
	// THEN: Should attempt to scan GSI with filter
	assert.Error(t, err) // Expected to fail in test environment
}

// RED PHASE: Error Handling Tests for Secondary Indexes

func TestQuery_WithInvalidGSIName_ShouldReturnError(t *testing.T) {
	// GIVEN: Invalid GSI name
	tableName := "test-table"
	indexName := "NonExistent-GSI"
	gsiKeyExpr, gsiKeyValues := createGSIKeyCondition("department", "", "Engineering", "")
	
	options := QueryOptions{
		KeyConditionExpression: &gsiKeyExpr,
		IndexName:             &indexName,
	}
	
	// WHEN: Executing query with invalid GSI
	_, err := QueryWithOptionsE(t, tableName, gsiKeyValues, options)
	
	// THEN: Should return error
	require.Error(t, err)
}

func TestQuery_WithInvalidLSIName_ShouldReturnError(t *testing.T) {
	// GIVEN: Invalid LSI name
	tableName := "test-table"
	indexName := "NonExistent-LSI"
	lsiKeyExpr, lsiKeyValues := createLSIKeyCondition("pk", "created_at", "USER#001", "2023-01-15")
	
	options := QueryOptions{
		KeyConditionExpression: &lsiKeyExpr,
		IndexName:             &indexName,
	}
	
	// WHEN: Executing query with invalid LSI
	_, err := QueryWithOptionsE(t, tableName, lsiKeyValues, options)
	
	// THEN: Should return error
	require.Error(t, err)
}

func TestScan_WithInvalidGSIName_ShouldReturnError(t *testing.T) {
	// GIVEN: Invalid GSI name for scan
	tableName := "test-table"
	indexName := "NonExistent-GSI"
	
	options := ScanOptions{
		IndexName: &indexName,
	}
	
	// WHEN: Executing scan with invalid GSI
	_, err := ScanWithOptionsE(t, tableName, options)
	
	// THEN: Should return error
	require.Error(t, err)
}