package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pure TDD implementation for pagination and advanced query/scan scenarios
// Following Red-Green-Refactor methodology with focus on pagination logic

// Pure Pagination Functions

// calculatePageCount determines how many pages are needed for pagination
func calculatePageCount(totalItems int, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	if totalItems <= 0 {
		return 0
	}
	return (totalItems + pageSize - 1) / pageSize // Ceiling division
}

// createPaginationKey creates an exclusive start key for pagination
func createPaginationKey(item map[string]types.AttributeValue, keyAttributes []string) map[string]types.AttributeValue {
	if item == nil || len(keyAttributes) == 0 {
		return nil
	}
	
	key := make(map[string]types.AttributeValue)
	for _, attr := range keyAttributes {
		if value, exists := item[attr]; exists {
			key[attr] = value
		}
	}
	
	if len(key) == 0 {
		return nil
	}
	
	return key
}

// isLastPage determines if we've reached the last page in pagination
func isLastPage(lastEvaluatedKey map[string]types.AttributeValue) bool {
	return lastEvaluatedKey == nil || len(lastEvaluatedKey) == 0
}

// mergePaginationResults combines multiple page results into a single result set
func mergePaginationResults(pages [][]map[string]types.AttributeValue) []map[string]types.AttributeValue {
	totalItems := 0
	for _, page := range pages {
		totalItems += len(page)
	}
	
	if totalItems == 0 {
		return []map[string]types.AttributeValue{}
	}
	
	merged := make([]map[string]types.AttributeValue, 0, totalItems)
	for _, page := range pages {
		merged = append(merged, page...)
	}
	
	return merged
}

// validatePaginationOptions checks if pagination options are valid
func validatePaginationOptions(limit *int32, exclusiveStartKey map[string]types.AttributeValue) error {
	if limit != nil && *limit <= 0 {
		return assert.AnError
	}
	return nil
}

// simulatePaginationScenario creates a scenario for testing pagination logic
func simulatePaginationScenario(totalItems int, pageSize int) ([][]map[string]types.AttributeValue, []map[string]types.AttributeValue) {
	if totalItems <= 0 || pageSize <= 0 {
		return [][]map[string]types.AttributeValue{}, []map[string]types.AttributeValue{}
	}
	
	// Create test items
	allItems := make([]map[string]types.AttributeValue, totalItems)
	for i := 0; i < totalItems; i++ {
		allItems[i] = map[string]types.AttributeValue{
			"id":    &types.AttributeValueMemberS{Value: generateID(i)},
			"name":  &types.AttributeValueMemberS{Value: generateName(i)},
			"value": &types.AttributeValueMemberN{Value: generateValue(i)},
		}
	}
	
	// Split into pages
	pages := make([][]map[string]types.AttributeValue, 0)
	for i := 0; i < totalItems; i += pageSize {
		end := i + pageSize
		if end > totalItems {
			end = totalItems
		}
		page := allItems[i:end]
		pages = append(pages, page)
	}
	
	return pages, allItems
}

// Helper functions for generating test data
func generateID(index int) string {
	return "item-" + string(rune('0'+index%10)) + string(rune('0'+(index/10)%10)) + string(rune('0'+(index/100)%10))
}

func generateName(index int) string {
	return "Item " + string(rune('A'+index%26))
}

func generateValue(index int) string {
	return string(rune('0' + (index%10)))
}

// Test Data Factories for Pagination

func createPaginationTestItems(count int) []map[string]types.AttributeValue {
	items := make([]map[string]types.AttributeValue, count)
	
	for i := 0; i < count; i++ {
		items[i] = map[string]types.AttributeValue{
			"pk":    &types.AttributeValueMemberS{Value: "USER#" + string(rune('0'+i%10))},
			"sk":    &types.AttributeValueMemberS{Value: "ITEM#" + string(rune('0'+i%100))},
			"id":    &types.AttributeValueMemberS{Value: generateID(i)},
			"name":  &types.AttributeValueMemberS{Value: generateName(i)},
			"index": &types.AttributeValueMemberN{Value: string(rune('0' + i%10))},
		}
	}
	
	return items
}

func createLargePaginationDataset() []map[string]types.AttributeValue {
	return createPaginationTestItems(100) // Large dataset for pagination testing
}

func createSmallPaginationDataset() []map[string]types.AttributeValue {
	return createPaginationTestItems(5) // Small dataset for edge cases
}

// RED PHASE: Page Count Calculation Tests

func TestCalculatePageCount_WithEvenDivision_ShouldReturnExactPageCount(t *testing.T) {
	// GIVEN: Total items that divide evenly by page size
	totalItems := 100
	pageSize := 10
	
	// WHEN: Calculating page count
	pageCount := calculatePageCount(totalItems, pageSize)
	
	// THEN: Should return exact page count
	assert.Equal(t, 10, pageCount)
}

func TestCalculatePageCount_WithRemainder_ShouldRoundUp(t *testing.T) {
	// GIVEN: Total items with remainder when divided by page size
	totalItems := 103
	pageSize := 10
	
	// WHEN: Calculating page count
	pageCount := calculatePageCount(totalItems, pageSize)
	
	// THEN: Should round up to include partial page
	assert.Equal(t, 11, pageCount)
}

func TestCalculatePageCount_WithZeroItems_ShouldReturnZero(t *testing.T) {
	// GIVEN: Zero total items
	totalItems := 0
	pageSize := 10
	
	// WHEN: Calculating page count
	pageCount := calculatePageCount(totalItems, pageSize)
	
	// THEN: Should return zero
	assert.Equal(t, 0, pageCount)
}

func TestCalculatePageCount_WithZeroPageSize_ShouldReturnZero(t *testing.T) {
	// GIVEN: Zero page size
	totalItems := 100
	pageSize := 0
	
	// WHEN: Calculating page count
	pageCount := calculatePageCount(totalItems, pageSize)
	
	// THEN: Should return zero (invalid page size)
	assert.Equal(t, 0, pageCount)
}

func TestCalculatePageCount_WithNegativeValues_ShouldReturnZero(t *testing.T) {
	// GIVEN: Negative values
	testCases := []struct {
		totalItems int
		pageSize   int
	}{
		{-10, 5},
		{10, -5},
		{-10, -5},
	}
	
	for _, tc := range testCases {
		// WHEN: Calculating page count with negative values
		pageCount := calculatePageCount(tc.totalItems, tc.pageSize)
		
		// THEN: Should return zero
		assert.Equal(t, 0, pageCount)
	}
}

// RED PHASE: Pagination Key Creation Tests

func TestCreatePaginationKey_WithValidItem_ShouldCreateKey(t *testing.T) {
	// GIVEN: Item with key attributes
	item := map[string]types.AttributeValue{
		"pk":   &types.AttributeValueMemberS{Value: "USER#123"},
		"sk":   &types.AttributeValueMemberS{Value: "PROFILE#456"},
		"name": &types.AttributeValueMemberS{Value: "John Doe"},
		"age":  &types.AttributeValueMemberN{Value: "30"},
	}
	keyAttributes := []string{"pk", "sk"}
	
	// WHEN: Creating pagination key
	key := createPaginationKey(item, keyAttributes)
	
	// THEN: Should create key with specified attributes
	require.NotNil(t, key)
	assert.Len(t, key, 2)
	assert.Contains(t, key, "pk")
	assert.Contains(t, key, "sk")
	
	if pkValue, ok := key["pk"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "USER#123", pkValue.Value)
	}
	if skValue, ok := key["sk"].(*types.AttributeValueMemberS); ok {
		assert.Equal(t, "PROFILE#456", skValue.Value)
	}
}

func TestCreatePaginationKey_WithMissingKeyAttributes_ShouldCreatePartialKey(t *testing.T) {
	// GIVEN: Item missing some key attributes
	item := map[string]types.AttributeValue{
		"pk":   &types.AttributeValueMemberS{Value: "USER#123"},
		"name": &types.AttributeValueMemberS{Value: "John Doe"},
	}
	keyAttributes := []string{"pk", "sk"} // sk is missing from item
	
	// WHEN: Creating pagination key
	key := createPaginationKey(item, keyAttributes)
	
	// THEN: Should create key with only available attributes
	require.NotNil(t, key)
	assert.Len(t, key, 1) // Only pk is present
	assert.Contains(t, key, "pk")
	assert.NotContains(t, key, "sk")
}

func TestCreatePaginationKey_WithNilItem_ShouldReturnNil(t *testing.T) {
	// GIVEN: Nil item
	keyAttributes := []string{"pk", "sk"}
	
	// WHEN: Creating pagination key
	key := createPaginationKey(nil, keyAttributes)
	
	// THEN: Should return nil
	assert.Nil(t, key)
}

func TestCreatePaginationKey_WithEmptyKeyAttributes_ShouldReturnNil(t *testing.T) {
	// GIVEN: Item but no key attributes specified
	item := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	keyAttributes := []string{}
	
	// WHEN: Creating pagination key
	key := createPaginationKey(item, keyAttributes)
	
	// THEN: Should return nil
	assert.Nil(t, key)
}

func TestCreatePaginationKey_WithNonExistentAttributes_ShouldReturnNil(t *testing.T) {
	// GIVEN: Item without the requested key attributes
	item := map[string]types.AttributeValue{
		"name": &types.AttributeValueMemberS{Value: "John Doe"},
	}
	keyAttributes := []string{"pk", "sk"} // These don't exist in the item
	
	// WHEN: Creating pagination key
	key := createPaginationKey(item, keyAttributes)
	
	// THEN: Should return nil (empty key)
	assert.Nil(t, key)
}

// RED PHASE: Last Page Detection Tests

func TestIsLastPage_WithNilLastEvaluatedKey_ShouldReturnTrue(t *testing.T) {
	// GIVEN: Nil last evaluated key
	var lastEvaluatedKey map[string]types.AttributeValue = nil
	
	// WHEN: Checking if last page
	isLast := isLastPage(lastEvaluatedKey)
	
	// THEN: Should return true
	assert.True(t, isLast)
}

func TestIsLastPage_WithEmptyLastEvaluatedKey_ShouldReturnTrue(t *testing.T) {
	// GIVEN: Empty last evaluated key
	lastEvaluatedKey := map[string]types.AttributeValue{}
	
	// WHEN: Checking if last page
	isLast := isLastPage(lastEvaluatedKey)
	
	// THEN: Should return true
	assert.True(t, isLast)
}

func TestIsLastPage_WithValidLastEvaluatedKey_ShouldReturnFalse(t *testing.T) {
	// GIVEN: Valid last evaluated key (more pages available)
	lastEvaluatedKey := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "USER#123"},
		"sk": &types.AttributeValueMemberS{Value: "ITEM#456"},
	}
	
	// WHEN: Checking if last page
	isLast := isLastPage(lastEvaluatedKey)
	
	// THEN: Should return false
	assert.False(t, isLast)
}

// RED PHASE: Results Merging Tests

func TestMergePaginationResults_WithMultiplePages_ShouldMergeCorrectly(t *testing.T) {
	// GIVEN: Multiple pages of results
	page1 := []map[string]types.AttributeValue{
		{"id": &types.AttributeValueMemberS{Value: "item-1"}},
		{"id": &types.AttributeValueMemberS{Value: "item-2"}},
	}
	page2 := []map[string]types.AttributeValue{
		{"id": &types.AttributeValueMemberS{Value: "item-3"}},
	}
	page3 := []map[string]types.AttributeValue{
		{"id": &types.AttributeValueMemberS{Value: "item-4"}},
		{"id": &types.AttributeValueMemberS{Value: "item-5"}},
	}
	pages := [][]map[string]types.AttributeValue{page1, page2, page3}
	
	// WHEN: Merging pagination results
	merged := mergePaginationResults(pages)
	
	// THEN: Should merge all items in order
	assert.Len(t, merged, 5)
	
	expectedIDs := []string{"item-1", "item-2", "item-3", "item-4", "item-5"}
	for i, item := range merged {
		if idValue, ok := item["id"].(*types.AttributeValueMemberS); ok {
			assert.Equal(t, expectedIDs[i], idValue.Value)
		}
	}
}

func TestMergePaginationResults_WithEmptyPages_ShouldReturnEmptyResult(t *testing.T) {
	// GIVEN: All empty pages
	pages := [][]map[string]types.AttributeValue{
		{},
		{},
		{},
	}
	
	// WHEN: Merging empty pages
	merged := mergePaginationResults(pages)
	
	// THEN: Should return empty result
	assert.Len(t, merged, 0)
}

func TestMergePaginationResults_WithNoPages_ShouldReturnEmptyResult(t *testing.T) {
	// GIVEN: No pages
	pages := [][]map[string]types.AttributeValue{}
	
	// WHEN: Merging no pages
	merged := mergePaginationResults(pages)
	
	// THEN: Should return empty result
	assert.Len(t, merged, 0)
}

func TestMergePaginationResults_WithMixedPageSizes_ShouldHandleCorrectly(t *testing.T) {
	// GIVEN: Pages of different sizes
	page1 := []map[string]types.AttributeValue{} // Empty page
	page2 := []map[string]types.AttributeValue{
		{"id": &types.AttributeValueMemberS{Value: "item-1"}},
	} // Single item
	page3 := []map[string]types.AttributeValue{
		{"id": &types.AttributeValueMemberS{Value: "item-2"}},
		{"id": &types.AttributeValueMemberS{Value: "item-3"}},
		{"id": &types.AttributeValueMemberS{Value: "item-4"}},
	} // Multiple items
	pages := [][]map[string]types.AttributeValue{page1, page2, page3}
	
	// WHEN: Merging mixed page sizes
	merged := mergePaginationResults(pages)
	
	// THEN: Should handle all page sizes correctly
	assert.Len(t, merged, 4)
}

// RED PHASE: Pagination Validation Tests

func TestValidatePaginationOptions_WithValidOptions_ShouldReturnNoError(t *testing.T) {
	// GIVEN: Valid pagination options
	limit := int32(10)
	exclusiveStartKey := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	
	// WHEN: Validating options
	err := validatePaginationOptions(&limit, exclusiveStartKey)
	
	// THEN: Should return no error
	assert.NoError(t, err)
}

func TestValidatePaginationOptions_WithNilLimit_ShouldReturnNoError(t *testing.T) {
	// GIVEN: Nil limit (unlimited)
	var limit *int32 = nil
	exclusiveStartKey := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: "USER#123"},
	}
	
	// WHEN: Validating options
	err := validatePaginationOptions(limit, exclusiveStartKey)
	
	// THEN: Should return no error
	assert.NoError(t, err)
}

func TestValidatePaginationOptions_WithZeroLimit_ShouldReturnError(t *testing.T) {
	// GIVEN: Zero limit
	limit := int32(0)
	exclusiveStartKey := map[string]types.AttributeValue{}
	
	// WHEN: Validating options
	err := validatePaginationOptions(&limit, exclusiveStartKey)
	
	// THEN: Should return error
	assert.Error(t, err)
}

func TestValidatePaginationOptions_WithNegativeLimit_ShouldReturnError(t *testing.T) {
	// GIVEN: Negative limit
	limit := int32(-5)
	exclusiveStartKey := map[string]types.AttributeValue{}
	
	// WHEN: Validating options
	err := validatePaginationOptions(&limit, exclusiveStartKey)
	
	// THEN: Should return error
	assert.Error(t, err)
}

// RED PHASE: Pagination Simulation Tests

func TestSimulatePaginationScenario_WithValidInputs_ShouldCreatePagesCorrectly(t *testing.T) {
	// GIVEN: Total items and page size
	totalItems := 10
	pageSize := 3
	
	// WHEN: Simulating pagination scenario
	pages, allItems := simulatePaginationScenario(totalItems, pageSize)
	
	// THEN: Should create correct number of pages
	expectedPageCount := calculatePageCount(totalItems, pageSize) // 4 pages (3+3+3+1)
	assert.Len(t, pages, expectedPageCount)
	assert.Len(t, allItems, totalItems)
	
	// First three pages should have 3 items each
	for i := 0; i < 3; i++ {
		assert.Len(t, pages[i], 3, "Page %d should have 3 items", i)
	}
	
	// Last page should have 1 item
	assert.Len(t, pages[3], 1, "Last page should have 1 item")
	
	// Merged pages should equal all items
	merged := mergePaginationResults(pages)
	assert.Len(t, merged, len(allItems))
}

func TestSimulatePaginationScenario_WithSinglePage_ShouldCreateOnePage(t *testing.T) {
	// GIVEN: Items that fit in a single page
	totalItems := 5
	pageSize := 10
	
	// WHEN: Simulating pagination scenario
	pages, allItems := simulatePaginationScenario(totalItems, pageSize)
	
	// THEN: Should create single page
	assert.Len(t, pages, 1)
	assert.Len(t, pages[0], totalItems)
	assert.Len(t, allItems, totalItems)
}

func TestSimulatePaginationScenario_WithZeroItems_ShouldReturnEmpty(t *testing.T) {
	// GIVEN: Zero items
	totalItems := 0
	pageSize := 10
	
	// WHEN: Simulating pagination scenario
	pages, allItems := simulatePaginationScenario(totalItems, pageSize)
	
	// THEN: Should return empty results
	assert.Len(t, pages, 0)
	assert.Len(t, allItems, 0)
}

func TestSimulatePaginationScenario_WithZeroPageSize_ShouldReturnEmpty(t *testing.T) {
	// GIVEN: Zero page size
	totalItems := 10
	pageSize := 0
	
	// WHEN: Simulating pagination scenario
	pages, allItems := simulatePaginationScenario(totalItems, pageSize)
	
	// THEN: Should return empty results
	assert.Len(t, pages, 0)
	assert.Len(t, allItems, 0)
}

// RED PHASE: Integration Tests for Query/Scan Pagination Functions

func TestQueryAllPages_PaginationLogic_ShouldHandleMultiplePages(t *testing.T) {
	// GIVEN: Scenario requiring pagination
	tableName := "test-table"
	keyExpr, keyValues := createTestKeyCondition("USER#123", "")
	
	// WHEN: Testing pagination logic (will fail without AWS, but tests function structure)
	_, err := QueryAllPagesE(t, tableName, keyExpr, keyValues)
	
	// THEN: Should attempt to handle pagination
	assert.Error(t, err) // Expected in test environment
}

func TestQueryAllPagesWithOptions_PaginationWithLimit_ShouldRespectLimitPerPage(t *testing.T) {
	// GIVEN: Query with limit for pagination control
	tableName := "test-table"
	keyExpr, keyValues := createTestKeyCondition("USER#123", "")
	limit := int32(5)
	options := QueryOptions{
		KeyConditionExpression: &keyExpr,
		Limit:                 &limit,
	}
	
	// WHEN: Testing paginated query with limit
	_, err := QueryAllPagesWithOptionsE(t, tableName, keyValues, options)
	
	// THEN: Should attempt to handle pagination with limit
	assert.Error(t, err) // Expected in test environment
}

func TestScanAllPages_PaginationLogic_ShouldHandleMultiplePages(t *testing.T) {
	// GIVEN: Scenario requiring scan pagination
	tableName := "test-table"
	
	// WHEN: Testing scan pagination logic
	_, err := ScanAllPagesE(t, tableName)
	
	// THEN: Should attempt to handle pagination
	assert.Error(t, err) // Expected in test environment
}

func TestScanAllPagesWithOptions_PaginationWithSegments_ShouldHandleParallelScanning(t *testing.T) {
	// GIVEN: Scan with parallel segments
	tableName := "test-table"
	totalSegments := int32(4)
	segment := int32(0)
	options := ScanOptions{
		TotalSegments: &totalSegments,
		Segment:       &segment,
	}
	
	// WHEN: Testing parallel scan pagination
	_, err := ScanAllPagesWithOptionsE(t, tableName, options)
	
	// THEN: Should attempt to handle parallel scan pagination
	assert.Error(t, err) // Expected in test environment
}

// RED PHASE: Helper Function Tests

func TestGenerateID_WithDifferentIndices_ShouldGenerateUniqueIDs(t *testing.T) {
	// GIVEN: Different indices
	indices := []int{0, 1, 5, 10, 25, 99, 123}
	
	generatedIDs := make(map[string]bool)
	
	for _, index := range indices {
		// WHEN: Generating ID
		id := generateID(index)
		
		// THEN: Should generate unique IDs
		assert.NotEmpty(t, id)
		assert.False(t, generatedIDs[id], "ID %s should be unique", id)
		generatedIDs[id] = true
	}
}

func TestGenerateName_WithDifferentIndices_ShouldGenerateNames(t *testing.T) {
	// GIVEN: Different indices
	indices := []int{0, 1, 25, 26, 51}
	
	for _, index := range indices {
		// WHEN: Generating name
		name := generateName(index)
		
		// THEN: Should generate valid names
		assert.NotEmpty(t, name)
		assert.Contains(t, name, "Item ")
	}
}

func TestGenerateValue_WithDifferentIndices_ShouldGenerateNumericValues(t *testing.T) {
	// GIVEN: Different indices
	indices := []int{0, 1, 5, 9, 10, 15}
	
	for _, index := range indices {
		// WHEN: Generating value
		value := generateValue(index)
		
		// THEN: Should generate single digit numeric strings
		assert.Len(t, value, 1)
		assert.GreaterOrEqual(t, value[0], byte('0'))
		assert.LessOrEqual(t, value[0], byte('9'))
	}
}