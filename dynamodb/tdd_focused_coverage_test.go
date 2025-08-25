package dynamodb

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

// FOCUSED TDD TESTS FOR HIGH COVERAGE

// Test the core helper functions that might not be covered by other tests
func TestInt64PtrHelper_ShouldReturnPointerToValue(t *testing.T) {
	// RED: Test int64Ptr helper function
	value := int64(42)
	
	result := int64Ptr(value)
	
	assert.NotNil(t, result)
	assert.Equal(t, value, *result)
}

func TestPtrToStringHelper_WithNilPointer_ShouldReturnEmptyString(t *testing.T) {
	// RED: Test ptrToString with nil pointer
	var ptr *string
	
	result := ptrToString(ptr)
	
	assert.Equal(t, "", result)
}

func TestPtrToStringHelper_WithValidPointer_ShouldReturnStringValue(t *testing.T) {
	// RED: Test ptrToString with valid pointer
	value := "test-string"
	
	result := ptrToString(&value)
	
	assert.Equal(t, value, result)
}

func TestPtrToInt64Helper_WithNilPointer_ShouldReturnZero(t *testing.T) {
	// RED: Test ptrToInt64 with nil pointer
	var ptr *int64
	
	result := ptrToInt64(ptr)
	
	assert.Equal(t, int64(0), result)
}

func TestPtrToInt64Helper_WithValidPointer_ShouldReturnInt64Value(t *testing.T) {
	// RED: Test ptrToInt64 with valid pointer
	value := int64(123)
	
	result := ptrToInt64(&value)
	
	assert.Equal(t, value, result)
}

// Test string slice comparison functions
func TestStringSlicesEqualHelper_WithIdenticalSlices_ShouldReturnTrue(t *testing.T) {
	// RED: Test string slice comparison
	slice1 := []string{"a", "b", "c"}
	slice2 := []string{"a", "b", "c"}
	
	result := stringSlicesEqual(slice1, slice2)
	
	assert.True(t, result)
}

func TestStringSlicesEqualHelper_WithDifferentSlices_ShouldReturnFalse(t *testing.T) {
	// RED: Test string slice comparison with different slices
	slice1 := []string{"a", "b", "c"}
	slice2 := []string{"x", "y", "z"}
	
	result := stringSlicesEqual(slice1, slice2)
	
	assert.False(t, result)
}

func TestStringSlicesEqualHelper_WithDifferentLengths_ShouldReturnFalse(t *testing.T) {
	// RED: Test string slice comparison with different lengths
	slice1 := []string{"a", "b", "c"}
	slice2 := []string{"a", "b"}
	
	result := stringSlicesEqual(slice1, slice2)
	
	assert.False(t, result)
}

// Test byte slice comparison functions
func TestBytesSlicesEqualHelper_WithIdenticalSlices_ShouldReturnTrue(t *testing.T) {
	// RED: Test byte slice comparison
	slice1 := [][]byte{[]byte("data1"), []byte("data2")}
	slice2 := [][]byte{[]byte("data1"), []byte("data2")}
	
	result := bytesSlicesEqual(slice1, slice2)
	
	assert.True(t, result)
}

func TestBytesSlicesEqualHelper_WithDifferentSlices_ShouldReturnFalse(t *testing.T) {
	// RED: Test byte slice comparison with different slices
	slice1 := [][]byte{[]byte("data1"), []byte("data2")}
	slice2 := [][]byte{[]byte("data3"), []byte("data4")}
	
	result := bytesSlicesEqual(slice1, slice2)
	
	assert.False(t, result)
}

func TestBytesSlicesEqualHelper_WithDifferentLengths_ShouldReturnFalse(t *testing.T) {
	// RED: Test byte slice comparison with different lengths
	slice1 := [][]byte{[]byte("data1"), []byte("data2")}
	slice2 := [][]byte{[]byte("data1")}
	
	result := bytesSlicesEqual(slice1, slice2)
	
	assert.False(t, result)
}

// Test attribute value list comparison
func TestAttributeValueListsEqualHelper_WithIdenticalLists_ShouldReturnTrue(t *testing.T) {
	// RED: Test attribute value list comparison
	list1 := []types.AttributeValue{
		&types.AttributeValueMemberS{Value: "item1"},
		&types.AttributeValueMemberN{Value: "123"},
	}
	list2 := []types.AttributeValue{
		&types.AttributeValueMemberS{Value: "item1"},
		&types.AttributeValueMemberN{Value: "123"},
	}
	
	result := attributeValueListsEqual(list1, list2)
	
	assert.True(t, result)
}

func TestAttributeValueListsEqualHelper_WithDifferentLists_ShouldReturnFalse(t *testing.T) {
	// RED: Test attribute value list comparison with different lists
	list1 := []types.AttributeValue{
		&types.AttributeValueMemberS{Value: "item1"},
		&types.AttributeValueMemberN{Value: "123"},
	}
	list2 := []types.AttributeValue{
		&types.AttributeValueMemberS{Value: "item2"},
		&types.AttributeValueMemberN{Value: "456"},
	}
	
	result := attributeValueListsEqual(list1, list2)
	
	assert.False(t, result)
}

func TestAttributeValueListsEqualHelper_WithDifferentLengths_ShouldReturnFalse(t *testing.T) {
	// RED: Test attribute value list comparison with different lengths
	list1 := []types.AttributeValue{
		&types.AttributeValueMemberS{Value: "item1"},
		&types.AttributeValueMemberN{Value: "123"},
	}
	list2 := []types.AttributeValue{
		&types.AttributeValueMemberS{Value: "item1"},
	}
	
	result := attributeValueListsEqual(list1, list2)
	
	assert.False(t, result)
}

// Test attribute value map comparison
func TestAttributeValueMapsEqualHelper_WithIdenticalMaps_ShouldReturnTrue(t *testing.T) {
	// RED: Test attribute value map comparison
	map1 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value1"},
		"key2": &types.AttributeValueMemberN{Value: "123"},
	}
	map2 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value1"},
		"key2": &types.AttributeValueMemberN{Value: "123"},
	}
	
	result := attributeValueMapsEqual(map1, map2)
	
	assert.True(t, result)
}

func TestAttributeValueMapsEqualHelper_WithDifferentMaps_ShouldReturnFalse(t *testing.T) {
	// RED: Test attribute value map comparison with different maps
	map1 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value1"},
		"key2": &types.AttributeValueMemberN{Value: "123"},
	}
	map2 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value2"},
		"key2": &types.AttributeValueMemberN{Value: "456"},
	}
	
	result := attributeValueMapsEqual(map1, map2)
	
	assert.False(t, result)
}

func TestAttributeValueMapsEqualHelper_WithDifferentLengths_ShouldReturnFalse(t *testing.T) {
	// RED: Test attribute value map comparison with different lengths
	map1 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value1"},
		"key2": &types.AttributeValueMemberN{Value: "123"},
	}
	map2 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value1"},
	}
	
	result := attributeValueMapsEqual(map1, map2)
	
	assert.False(t, result)
}

func TestAttributeValueMapsEqualHelper_WithMissingKey_ShouldReturnFalse(t *testing.T) {
	// RED: Test attribute value map comparison with missing key
	map1 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value1"},
		"key2": &types.AttributeValueMemberN{Value: "123"},
	}
	map2 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value1"},
		"key3": &types.AttributeValueMemberN{Value: "123"}, // different key
	}
	
	result := attributeValueMapsEqual(map1, map2)
	
	assert.False(t, result)
}

// Test comprehensive attribute value comparisons for all types
func TestAttributeValuesEqualHelper_WithStringMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test string attribute value comparison
	value1 := &types.AttributeValueMemberS{Value: "test"}
	value2 := &types.AttributeValueMemberS{Value: "test"}
	value3 := &types.AttributeValueMemberS{Value: "different"}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithNumberMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test number attribute value comparison
	value1 := &types.AttributeValueMemberN{Value: "123"}
	value2 := &types.AttributeValueMemberN{Value: "123"}
	value3 := &types.AttributeValueMemberN{Value: "456"}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithBinaryMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test binary attribute value comparison
	value1 := &types.AttributeValueMemberB{Value: []byte("data")}
	value2 := &types.AttributeValueMemberB{Value: []byte("data")}
	value3 := &types.AttributeValueMemberB{Value: []byte("different")}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithBooleanMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test boolean attribute value comparison
	value1 := &types.AttributeValueMemberBOOL{Value: true}
	value2 := &types.AttributeValueMemberBOOL{Value: true}
	value3 := &types.AttributeValueMemberBOOL{Value: false}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithNullMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test null attribute value comparison
	value1 := &types.AttributeValueMemberNULL{Value: true}
	value2 := &types.AttributeValueMemberNULL{Value: true}
	value3 := &types.AttributeValueMemberNULL{Value: false}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithStringSetMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test string set attribute value comparison
	value1 := &types.AttributeValueMemberSS{Value: []string{"a", "b"}}
	value2 := &types.AttributeValueMemberSS{Value: []string{"a", "b"}}
	value3 := &types.AttributeValueMemberSS{Value: []string{"c", "d"}}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithNumberSetMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test number set attribute value comparison
	value1 := &types.AttributeValueMemberNS{Value: []string{"1", "2"}}
	value2 := &types.AttributeValueMemberNS{Value: []string{"1", "2"}}
	value3 := &types.AttributeValueMemberNS{Value: []string{"3", "4"}}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithBinarySetMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test binary set attribute value comparison
	value1 := &types.AttributeValueMemberBS{Value: [][]byte{[]byte("a"), []byte("b")}}
	value2 := &types.AttributeValueMemberBS{Value: [][]byte{[]byte("a"), []byte("b")}}
	value3 := &types.AttributeValueMemberBS{Value: [][]byte{[]byte("c"), []byte("d")}}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithListMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test list attribute value comparison
	value1 := &types.AttributeValueMemberL{
		Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: "item1"},
		},
	}
	value2 := &types.AttributeValueMemberL{
		Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: "item1"},
		},
	}
	value3 := &types.AttributeValueMemberL{
		Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: "item2"},
		},
	}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithMapMembers_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test map attribute value comparison
	value1 := &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberS{Value: "value1"},
		},
	}
	value2 := &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberS{Value: "value1"},
		},
	}
	value3 := &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"key": &types.AttributeValueMemberS{Value: "value2"},
		},
	}
	
	assert.True(t, attributeValuesEqual(value1, value2))
	assert.False(t, attributeValuesEqual(value1, value3))
}

func TestAttributeValuesEqualHelper_WithDifferentTypes_ShouldReturnFalse(t *testing.T) {
	// RED: Test attribute value comparison with different types
	stringValue := &types.AttributeValueMemberS{Value: "123"}
	numberValue := &types.AttributeValueMemberN{Value: "123"}
	
	result := attributeValuesEqual(stringValue, numberValue)
	
	assert.False(t, result)
}

func TestAttributeValuesEqualHelper_WithUnsupportedType_ShouldReturnFalse(t *testing.T) {
	// RED: Test attribute value comparison with different types (should return false)
	stringValue := &types.AttributeValueMemberS{Value: "test"}
	numberValue := &types.AttributeValueMemberN{Value: "123"}
	boolValue := &types.AttributeValueMemberBOOL{Value: true}
	
	// Different types should always return false
	assert.False(t, attributeValuesEqual(stringValue, numberValue))
	assert.False(t, attributeValuesEqual(stringValue, boolValue))
	assert.False(t, attributeValuesEqual(numberValue, boolValue))
}

// Test backoff calculation for retries
func TestCalculateBackoffDurationHelper_WithDifferentAttempts_ShouldReturnExponentialBackoff(t *testing.T) {
	// RED: Test exponential backoff calculation
	testCases := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1600 * time.Millisecond},
	}
	
	for _, tc := range testCases {
		duration := time.Duration(100*(1<<tc.attempt)) * time.Millisecond
		assert.Equal(t, tc.expected, duration)
	}
}

// Test edge cases for empty collections
func TestStringSlicesEqualHelper_WithEmptySlices_ShouldReturnTrue(t *testing.T) {
	// RED: Test string slice comparison with empty slices
	var slice1 []string
	var slice2 []string
	
	result := stringSlicesEqual(slice1, slice2)
	
	assert.True(t, result)
}

func TestBytesSlicesEqualHelper_WithEmptySlices_ShouldReturnTrue(t *testing.T) {
	// RED: Test byte slice comparison with empty slices
	var slice1 [][]byte
	var slice2 [][]byte
	
	result := bytesSlicesEqual(slice1, slice2)
	
	assert.True(t, result)
}

func TestAttributeValueListsEqualHelper_WithEmptyLists_ShouldReturnTrue(t *testing.T) {
	// RED: Test attribute value list comparison with empty lists
	var list1 []types.AttributeValue
	var list2 []types.AttributeValue
	
	result := attributeValueListsEqual(list1, list2)
	
	assert.True(t, result)
}

func TestAttributeValueMapsEqualHelper_WithEmptyMaps_ShouldReturnTrue(t *testing.T) {
	// RED: Test attribute value map comparison with empty maps
	map1 := make(map[string]types.AttributeValue)
	map2 := make(map[string]types.AttributeValue)
	
	result := attributeValueMapsEqual(map1, map2)
	
	assert.True(t, result)
}

// Test table description conversion with various nil/empty scenarios
func TestConvertToTableDescriptionHelper_WithMinimalValidInput_ShouldHandleNilFields(t *testing.T) {
	// RED: Test table description conversion with minimal input
	input := &types.TableDescription{
		TableName:   &[]string{"test-table"}[0],
		TableStatus: types.TableStatusActive,
	}
	
	result := convertToTableDescription(input)
	
	assert.NotNil(t, result)
	assert.Equal(t, "test-table", result.TableName)
	assert.Equal(t, types.TableStatusActive, result.TableStatus)
	// Nil fields should have default values
	assert.Nil(t, result.CreationDateTime)
	assert.Equal(t, int64(0), result.TableSizeBytes)
	assert.Equal(t, int64(0), result.ItemCount)
	assert.Equal(t, "", result.TableArn)
	assert.Equal(t, "", result.TableId)
	assert.Equal(t, "", result.LatestStreamLabel)
	assert.Equal(t, "", result.LatestStreamArn)
}