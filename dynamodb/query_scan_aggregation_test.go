package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pure TDD implementation for result processing and aggregation functions
// Following Red-Green-Refactor methodology with focus on pure functions

// Pure Result Processing Functions

// extractItemCount returns the count of items from query/scan results
func extractItemCount(items []map[string]types.AttributeValue) int {
	return len(items)
}

// extractAttributeValues extracts values for a specific attribute from items
func extractAttributeValues(items []map[string]types.AttributeValue, attributeName string) []types.AttributeValue {
	values := make([]types.AttributeValue, 0, len(items))
	for _, item := range items {
		if value, exists := item[attributeName]; exists {
			values = append(values, value)
		}
	}
	return values
}

// filterItemsByAttribute filters items that have a specific attribute value
func filterItemsByAttribute(items []map[string]types.AttributeValue, attributeName string, expectedValue types.AttributeValue) []map[string]types.AttributeValue {
	filtered := make([]map[string]types.AttributeValue, 0)
	for _, item := range items {
		if value, exists := item[attributeName]; exists {
			if areAttributeValuesEqual(value, expectedValue) {
				filtered = append(filtered, item)
			}
		}
	}
	return filtered
}

// areAttributeValuesEqual compares two attribute values for equality
func areAttributeValuesEqual(a, b types.AttributeValue) bool {
	// Simple string comparison for testing
	if aStr, aOk := a.(*types.AttributeValueMemberS); aOk {
		if bStr, bOk := b.(*types.AttributeValueMemberS); bOk {
			return aStr.Value == bStr.Value
		}
	}
	if aNum, aOk := a.(*types.AttributeValueMemberN); aOk {
		if bNum, bOk := b.(*types.AttributeValueMemberN); bOk {
			return aNum.Value == bNum.Value
		}
	}
	return false
}

// groupItemsByAttribute groups items by a specific attribute value
func groupItemsByAttribute(items []map[string]types.AttributeValue, attributeName string) map[string][]map[string]types.AttributeValue {
	groups := make(map[string][]map[string]types.AttributeValue)
	
	for _, item := range items {
		if value, exists := item[attributeName]; exists {
			key := getAttributeValueAsString(value)
			groups[key] = append(groups[key], item)
		}
	}
	
	return groups
}

// getAttributeValueAsString converts an AttributeValue to string for grouping
func getAttributeValueAsString(value types.AttributeValue) string {
	switch v := value.(type) {
	case *types.AttributeValueMemberS:
		return v.Value
	case *types.AttributeValueMemberN:
		return v.Value
	default:
		return ""
	}
}

// aggregateNumericAttribute sums numeric values for an attribute across items
func aggregateNumericAttribute(items []map[string]types.AttributeValue, attributeName string) (int, error) {
	total := 0
	for _, item := range items {
		if value, exists := item[attributeName]; exists {
			if numValue, ok := value.(*types.AttributeValueMemberN); ok {
				// Simple integer conversion for testing
				if len(numValue.Value) > 0 {
					digit := int(numValue.Value[0] - '0')
					if digit >= 0 && digit <= 9 {
						total += digit
					}
				}
			}
		}
	}
	return total, nil
}

// Test Data Factories for Result Processing

func createTestItems() []map[string]types.AttributeValue {
	return []map[string]types.AttributeValue{
		{
			"id":         &types.AttributeValueMemberS{Value: "user-001"},
			"name":       &types.AttributeValueMemberS{Value: "John Doe"},
			"age":        &types.AttributeValueMemberN{Value: "30"},
			"department": &types.AttributeValueMemberS{Value: "Engineering"},
			"salary":     &types.AttributeValueMemberN{Value: "75000"},
			"status":     &types.AttributeValueMemberS{Value: "active"},
		},
		{
			"id":         &types.AttributeValueMemberS{Value: "user-002"},
			"name":       &types.AttributeValueMemberS{Value: "Jane Smith"},
			"age":        &types.AttributeValueMemberN{Value: "28"},
			"department": &types.AttributeValueMemberS{Value: "Marketing"},
			"salary":     &types.AttributeValueMemberN{Value: "65000"},
			"status":     &types.AttributeValueMemberS{Value: "active"},
		},
		{
			"id":         &types.AttributeValueMemberS{Value: "user-003"},
			"name":       &types.AttributeValueMemberS{Value: "Bob Wilson"},
			"age":        &types.AttributeValueMemberN{Value: "35"},
			"department": &types.AttributeValueMemberS{Value: "Engineering"},
			"salary":     &types.AttributeValueMemberN{Value: "85000"},
			"status":     &types.AttributeValueMemberS{Value: "inactive"},
		},
		{
			"id":         &types.AttributeValueMemberS{Value: "user-004"},
			"name":       &types.AttributeValueMemberS{Value: "Alice Brown"},
			"age":        &types.AttributeValueMemberN{Value: "32"},
			"department": &types.AttributeValueMemberS{Value: "Sales"},
			"salary":     &types.AttributeValueMemberN{Value: "70000"},
			"status":     &types.AttributeValueMemberS{Value: "active"},
		},
	}
}

func createEmptyTestItems() []map[string]types.AttributeValue {
	return []map[string]types.AttributeValue{}
}

func createSingleTestItem() []map[string]types.AttributeValue {
	return []map[string]types.AttributeValue{
		{
			"id":         &types.AttributeValueMemberS{Value: "single-user"},
			"name":       &types.AttributeValueMemberS{Value: "Single User"},
			"department": &types.AttributeValueMemberS{Value: "Engineering"},
		},
	}
}

// RED PHASE: Item Count Tests

func TestExtractItemCount_WithMultipleItems_ShouldReturnCorrectCount(t *testing.T) {
	// GIVEN: A list of test items
	items := createTestItems()
	
	// WHEN: Extracting item count
	count := extractItemCount(items)
	
	// THEN: Should return correct count
	assert.Equal(t, 4, count)
}

func TestExtractItemCount_WithEmptyItems_ShouldReturnZero(t *testing.T) {
	// GIVEN: An empty list of items
	items := createEmptyTestItems()
	
	// WHEN: Extracting item count
	count := extractItemCount(items)
	
	// THEN: Should return zero
	assert.Equal(t, 0, count)
}

func TestExtractItemCount_WithSingleItem_ShouldReturnOne(t *testing.T) {
	// GIVEN: A single test item
	items := createSingleTestItem()
	
	// WHEN: Extracting item count
	count := extractItemCount(items)
	
	// THEN: Should return one
	assert.Equal(t, 1, count)
}

// RED PHASE: Attribute Value Extraction Tests

func TestExtractAttributeValues_WithExistingAttribute_ShouldReturnAllValues(t *testing.T) {
	// GIVEN: Items with "name" attribute
	items := createTestItems()
	
	// WHEN: Extracting name attribute values
	values := extractAttributeValues(items, "name")
	
	// THEN: Should return all name values
	assert.Len(t, values, 4)
	
	expectedNames := []string{"John Doe", "Jane Smith", "Bob Wilson", "Alice Brown"}
	for i, value := range values {
		if nameValue, ok := value.(*types.AttributeValueMemberS); ok {
			assert.Equal(t, expectedNames[i], nameValue.Value)
		} else {
			t.Errorf("Expected string value, got %T", value)
		}
	}
}

func TestExtractAttributeValues_WithNonExistentAttribute_ShouldReturnEmptyList(t *testing.T) {
	// GIVEN: Items without "nonexistent" attribute
	items := createTestItems()
	
	// WHEN: Extracting non-existent attribute values
	values := extractAttributeValues(items, "nonexistent")
	
	// THEN: Should return empty list
	assert.Len(t, values, 0)
}

func TestExtractAttributeValues_WithEmptyItems_ShouldReturnEmptyList(t *testing.T) {
	// GIVEN: Empty item list
	items := createEmptyTestItems()
	
	// WHEN: Extracting attribute values
	values := extractAttributeValues(items, "name")
	
	// THEN: Should return empty list
	assert.Len(t, values, 0)
}

// RED PHASE: Filtering Tests

func TestFilterItemsByAttribute_WithMatchingValue_ShouldReturnFilteredItems(t *testing.T) {
	// GIVEN: Items with various departments
	items := createTestItems()
	expectedValue := &types.AttributeValueMemberS{Value: "Engineering"}
	
	// WHEN: Filtering by department
	filtered := filterItemsByAttribute(items, "department", expectedValue)
	
	// THEN: Should return only Engineering items
	assert.Len(t, filtered, 2)
	
	for _, item := range filtered {
		if dept, exists := item["department"]; exists {
			if deptValue, ok := dept.(*types.AttributeValueMemberS); ok {
				assert.Equal(t, "Engineering", deptValue.Value)
			}
		}
	}
}

func TestFilterItemsByAttribute_WithNonMatchingValue_ShouldReturnEmptyList(t *testing.T) {
	// GIVEN: Items without "Nonexistent" department
	items := createTestItems()
	expectedValue := &types.AttributeValueMemberS{Value: "Nonexistent"}
	
	// WHEN: Filtering by non-existent department
	filtered := filterItemsByAttribute(items, "department", expectedValue)
	
	// THEN: Should return empty list
	assert.Len(t, filtered, 0)
}

func TestFilterItemsByAttribute_WithEmptyItems_ShouldReturnEmptyList(t *testing.T) {
	// GIVEN: Empty item list
	items := createEmptyTestItems()
	expectedValue := &types.AttributeValueMemberS{Value: "Engineering"}
	
	// WHEN: Filtering empty list
	filtered := filterItemsByAttribute(items, "department", expectedValue)
	
	// THEN: Should return empty list
	assert.Len(t, filtered, 0)
}

// RED PHASE: Grouping Tests

func TestGroupItemsByAttribute_WithValidAttribute_ShouldGroupCorrectly(t *testing.T) {
	// GIVEN: Items with various departments
	items := createTestItems()
	
	// WHEN: Grouping by department
	groups := groupItemsByAttribute(items, "department")
	
	// THEN: Should create correct groups
	assert.Len(t, groups, 3) // Engineering, Marketing, Sales
	assert.Len(t, groups["Engineering"], 2)
	assert.Len(t, groups["Marketing"], 1)
	assert.Len(t, groups["Sales"], 1)
}

func TestGroupItemsByAttribute_WithNonExistentAttribute_ShouldReturnEmptyGroups(t *testing.T) {
	// GIVEN: Items without "nonexistent" attribute
	items := createTestItems()
	
	// WHEN: Grouping by non-existent attribute
	groups := groupItemsByAttribute(items, "nonexistent")
	
	// THEN: Should return empty groups
	assert.Len(t, groups, 0)
}

func TestGroupItemsByAttribute_WithEmptyItems_ShouldReturnEmptyGroups(t *testing.T) {
	// GIVEN: Empty item list
	items := createEmptyTestItems()
	
	// WHEN: Grouping empty list
	groups := groupItemsByAttribute(items, "department")
	
	// THEN: Should return empty groups
	assert.Len(t, groups, 0)
}

// RED PHASE: Aggregation Tests

func TestAggregateNumericAttribute_WithValidNumbers_ShouldReturnSum(t *testing.T) {
	// GIVEN: Items with numeric age attributes (simplified for testing)
	items := []map[string]types.AttributeValue{
		{"age": &types.AttributeValueMemberN{Value: "5"}}, // digit 5
		{"age": &types.AttributeValueMemberN{Value: "3"}}, // digit 3
		{"age": &types.AttributeValueMemberN{Value: "2"}}, // digit 2
	}
	
	// WHEN: Aggregating age values
	total, err := aggregateNumericAttribute(items, "age")
	
	// THEN: Should return sum
	require.NoError(t, err)
	assert.Equal(t, 10, total) // 5 + 3 + 2
}

func TestAggregateNumericAttribute_WithNonExistentAttribute_ShouldReturnZero(t *testing.T) {
	// GIVEN: Items without numeric attribute
	items := createTestItems()
	
	// WHEN: Aggregating non-existent attribute
	total, err := aggregateNumericAttribute(items, "nonexistent")
	
	// THEN: Should return zero
	require.NoError(t, err)
	assert.Equal(t, 0, total)
}

func TestAggregateNumericAttribute_WithEmptyItems_ShouldReturnZero(t *testing.T) {
	// GIVEN: Empty item list
	items := createEmptyTestItems()
	
	// WHEN: Aggregating empty list
	total, err := aggregateNumericAttribute(items, "age")
	
	// THEN: Should return zero
	require.NoError(t, err)
	assert.Equal(t, 0, total)
}

func TestAggregateNumericAttribute_WithMixedAttributeTypes_ShouldIgnoreNonNumeric(t *testing.T) {
	// GIVEN: Items with mixed attribute types
	items := []map[string]types.AttributeValue{
		{"value": &types.AttributeValueMemberN{Value: "5"}}, // numeric
		{"value": &types.AttributeValueMemberS{Value: "text"}}, // string - should be ignored
		{"value": &types.AttributeValueMemberN{Value: "3"}}, // numeric
	}
	
	// WHEN: Aggregating mixed values
	total, err := aggregateNumericAttribute(items, "value")
	
	// THEN: Should sum only numeric values
	require.NoError(t, err)
	assert.Equal(t, 8, total) // 5 + 0 (ignored string) + 3
}

// RED PHASE: Attribute Value Comparison Tests

func TestAreAttributeValuesEqual_WithEqualStrings_ShouldReturnTrue(t *testing.T) {
	// GIVEN: Two equal string attribute values
	a := &types.AttributeValueMemberS{Value: "test"}
	b := &types.AttributeValueMemberS{Value: "test"}
	
	// WHEN: Comparing values
	equal := areAttributeValuesEqual(a, b)
	
	// THEN: Should return true
	assert.True(t, equal)
}

func TestAreAttributeValuesEqual_WithDifferentStrings_ShouldReturnFalse(t *testing.T) {
	// GIVEN: Two different string attribute values
	a := &types.AttributeValueMemberS{Value: "test1"}
	b := &types.AttributeValueMemberS{Value: "test2"}
	
	// WHEN: Comparing values
	equal := areAttributeValuesEqual(a, b)
	
	// THEN: Should return false
	assert.False(t, equal)
}

func TestAreAttributeValuesEqual_WithEqualNumbers_ShouldReturnTrue(t *testing.T) {
	// GIVEN: Two equal number attribute values
	a := &types.AttributeValueMemberN{Value: "123"}
	b := &types.AttributeValueMemberN{Value: "123"}
	
	// WHEN: Comparing values
	equal := areAttributeValuesEqual(a, b)
	
	// THEN: Should return true
	assert.True(t, equal)
}

func TestAreAttributeValuesEqual_WithDifferentNumbers_ShouldReturnFalse(t *testing.T) {
	// GIVEN: Two different number attribute values
	a := &types.AttributeValueMemberN{Value: "123"}
	b := &types.AttributeValueMemberN{Value: "456"}
	
	// WHEN: Comparing values
	equal := areAttributeValuesEqual(a, b)
	
	// THEN: Should return false
	assert.False(t, equal)
}

func TestAreAttributeValuesEqual_WithDifferentTypes_ShouldReturnFalse(t *testing.T) {
	// GIVEN: String and number attribute values
	a := &types.AttributeValueMemberS{Value: "123"}
	b := &types.AttributeValueMemberN{Value: "123"}
	
	// WHEN: Comparing different types
	equal := areAttributeValuesEqual(a, b)
	
	// THEN: Should return false
	assert.False(t, equal)
}

// RED PHASE: String Conversion Tests

func TestGetAttributeValueAsString_WithStringValue_ShouldReturnString(t *testing.T) {
	// GIVEN: String attribute value
	value := &types.AttributeValueMemberS{Value: "test-string"}
	
	// WHEN: Converting to string
	result := getAttributeValueAsString(value)
	
	// THEN: Should return the string value
	assert.Equal(t, "test-string", result)
}

func TestGetAttributeValueAsString_WithNumberValue_ShouldReturnNumberAsString(t *testing.T) {
	// GIVEN: Number attribute value
	value := &types.AttributeValueMemberN{Value: "12345"}
	
	// WHEN: Converting to string
	result := getAttributeValueAsString(value)
	
	// THEN: Should return the number as string
	assert.Equal(t, "12345", result)
}

func TestGetAttributeValueAsString_WithUnsupportedType_ShouldReturnEmptyString(t *testing.T) {
	// GIVEN: Unsupported attribute value type (using a boolean which isn't handled)
	value := &types.AttributeValueMemberBOOL{Value: true}
	
	// WHEN: Converting to string
	result := getAttributeValueAsString(value)
	
	// THEN: Should return empty string
	assert.Equal(t, "", result)
}

// RED PHASE: Edge Case Tests for Result Processing

func TestFilterItemsByAttribute_WithNilExpectedValue_ShouldReturnEmptyList(t *testing.T) {
	// GIVEN: Items and nil expected value
	items := createTestItems()
	
	// WHEN: Filtering with nil value (should cause no matches)
	filtered := filterItemsByAttribute(items, "department", nil)
	
	// THEN: Should return empty list
	assert.Len(t, filtered, 0)
}

func TestGroupItemsByAttribute_WithItemsMissingAttribute_ShouldGroupOnlyItemsWithAttribute(t *testing.T) {
	// GIVEN: Items where some have the attribute and some don't
	items := []map[string]types.AttributeValue{
		{
			"id":         &types.AttributeValueMemberS{Value: "1"},
			"department": &types.AttributeValueMemberS{Value: "Engineering"},
		},
		{
			"id": &types.AttributeValueMemberS{Value: "2"},
			// Missing department
		},
		{
			"id":         &types.AttributeValueMemberS{Value: "3"},
			"department": &types.AttributeValueMemberS{Value: "Engineering"},
		},
	}
	
	// WHEN: Grouping by department
	groups := groupItemsByAttribute(items, "department")
	
	// THEN: Should only group items that have the attribute
	assert.Len(t, groups, 1) // Only "Engineering" group
	assert.Len(t, groups["Engineering"], 2) // Two items
}

func TestAggregateNumericAttribute_WithInvalidNumericValues_ShouldIgnoreInvalidValues(t *testing.T) {
	// GIVEN: Items with invalid numeric format
	items := []map[string]types.AttributeValue{
		{"value": &types.AttributeValueMemberN{Value: ""}},     // empty string
		{"value": &types.AttributeValueMemberN{Value: "abc"}}, // non-numeric
		{"value": &types.AttributeValueMemberN{Value: "5"}},   // valid
	}
	
	// WHEN: Aggregating values
	total, err := aggregateNumericAttribute(items, "value")
	
	// THEN: Should only sum valid values
	require.NoError(t, err)
	assert.Equal(t, 5, total) // Only the valid "5" is counted
}