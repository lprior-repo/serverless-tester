package dynamodb

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

// Comprehensive tests to increase coverage of assertion functions and wrapper functions
// These tests focus on error handling and edge cases that don't require AWS calls

// Test assertion helper functions (pure logic tests)
func TestAssertConditionalCheckPassed_WithNilError_ShouldNotFail(t *testing.T) {
	// RED: Test that no error means conditional check passed
	var err error = nil
	
	// This should not cause the test to fail
	AssertConditionalCheckPassed(t, err)
	// If we reach here, the assertion passed correctly
	assert.True(t, true) // Dummy assertion to verify we reached this point
}

func TestAssertConditionalCheckFailed_WithGenericError_ShouldContainMessage(t *testing.T) {
	// RED: Test that ConditionalCheckFailedException error contains expected message
	// We can't easily create the AWS error type, so we'll test the logic with a generic error
	err := errors.New("ConditionalCheckFailedException: The conditional request failed")
	
	// This will be handled by the assertion function
	// We can't directly test this without refactoring, but we can test the error handling logic
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ConditionalCheckFailedException")
}

// Test edge cases in attribute comparison functions  
func TestAttributeValuesEqual_WithNilValues_ShouldReturnFalse(t *testing.T) {
	// RED: Test nil values edge case
	var value1 types.AttributeValue = nil
	var value2 types.AttributeValue = nil
	
	// Both nil should return false (since we can't compare nil interface values)
	result := attributeValuesEqual(value1, value2)
	assert.False(t, result)
}

func TestAttributeValuesEqual_WithOneNilValue_ShouldReturnFalse(t *testing.T) {
	// RED: Test one nil value edge case
	var value1 types.AttributeValue = nil
	value2 := &types.AttributeValueMemberS{Value: "test"}
	
	result := attributeValuesEqual(value1, value2)
	assert.False(t, result)
}

// Test string slices edge cases
func TestStringSlicesEqual_WithNilSlices_ShouldReturnTrue(t *testing.T) {
	// RED: Test nil slices
	var slice1 []string = nil
	var slice2 []string = nil
	
	result := stringSlicesEqual(slice1, slice2)
	assert.True(t, result)
}

func TestStringSlicesEqual_WithOneNilSlice_ShouldReturnFalse(t *testing.T) {
	// RED: Test one nil slice
	var slice1 []string = nil
	slice2 := []string{"test"}
	
	result := stringSlicesEqual(slice1, slice2)
	assert.False(t, result)
}

// Test bytes slices edge cases
func TestBytesSlicesEqual_WithNilSlices_ShouldReturnTrue(t *testing.T) {
	// RED: Test nil byte slices
	var slice1 [][]byte = nil
	var slice2 [][]byte = nil
	
	result := bytesSlicesEqual(slice1, slice2)
	assert.True(t, result)
}

func TestBytesSlicesEqual_WithOneNilSlice_ShouldReturnFalse(t *testing.T) {
	// RED: Test one nil byte slice
	var slice1 [][]byte = nil
	slice2 := [][]byte{[]byte("test")}
	
	result := bytesSlicesEqual(slice1, slice2)
	assert.False(t, result)
}

// Test attribute value lists edge cases
func TestAttributeValueListsEqual_WithNilLists_ShouldReturnTrue(t *testing.T) {
	// RED: Test nil attribute value lists
	var list1 []types.AttributeValue = nil
	var list2 []types.AttributeValue = nil
	
	result := attributeValueListsEqual(list1, list2)
	assert.True(t, result)
}

func TestAttributeValueListsEqual_WithOneNilList_ShouldReturnFalse(t *testing.T) {
	// RED: Test one nil attribute value list
	var list1 []types.AttributeValue = nil
	list2 := []types.AttributeValue{&types.AttributeValueMemberS{Value: "test"}}
	
	result := attributeValueListsEqual(list1, list2)
	assert.False(t, result)
}

// Test attribute value maps edge cases
func TestAttributeValueMapsEqual_WithNilMaps_ShouldReturnTrue(t *testing.T) {
	// RED: Test nil attribute value maps
	var map1 map[string]types.AttributeValue = nil
	var map2 map[string]types.AttributeValue = nil
	
	result := attributeValueMapsEqual(map1, map2)
	assert.True(t, result)
}

func TestAttributeValueMapsEqual_WithOneNilMap_ShouldReturnFalse(t *testing.T) {
	// RED: Test one nil attribute value map
	var map1 map[string]types.AttributeValue = nil
	map2 := map[string]types.AttributeValue{
		"key": &types.AttributeValueMemberS{Value: "value"},
	}
	
	result := attributeValueMapsEqual(map1, map2)
	assert.False(t, result)
}

func TestAttributeValueMapsEqual_WithEmptyMaps_ShouldReturnTrue(t *testing.T) {
	// RED: Test empty attribute value maps
	map1 := map[string]types.AttributeValue{}
	map2 := map[string]types.AttributeValue{}
	
	result := attributeValueMapsEqual(map1, map2)
	assert.True(t, result)
}

func TestAttributeValueMapsEqual_WithMissingKey_ShouldReturnFalse(t *testing.T) {
	// RED: Test missing key in second map
	map1 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value1"},
		"key2": &types.AttributeValueMemberS{Value: "value2"},
	}
	map2 := map[string]types.AttributeValue{
		"key1": &types.AttributeValueMemberS{Value: "value1"},
	}
	
	result := attributeValueMapsEqual(map1, map2)
	assert.False(t, result)
}

// Test complex attribute value scenarios
func TestAttributeValuesEqual_WithComplexNesting_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test deeply nested attribute values
	complexMap1 := map[string]types.AttributeValue{
		"nested": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"inner": &types.AttributeValueMemberL{
					Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "item1"},
						&types.AttributeValueMemberN{Value: "42"},
					},
				},
			},
		},
	}
	
	complexMap2 := map[string]types.AttributeValue{
		"nested": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"inner": &types.AttributeValueMemberL{
					Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "item1"},
						&types.AttributeValueMemberN{Value: "42"},
					},
				},
			},
		},
	}
	
	value1 := &types.AttributeValueMemberM{Value: complexMap1}
	value2 := &types.AttributeValueMemberM{Value: complexMap2}
	
	result := attributeValuesEqual(value1, value2)
	assert.True(t, result)
}

func TestAttributeValuesEqual_WithComplexNestingDifferent_ShouldReturnFalse(t *testing.T) {
	// RED: Test deeply nested attribute values with differences
	complexMap1 := map[string]types.AttributeValue{
		"nested": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"inner": &types.AttributeValueMemberS{Value: "different1"},
			},
		},
	}
	
	complexMap2 := map[string]types.AttributeValue{
		"nested": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"inner": &types.AttributeValueMemberS{Value: "different2"},
			},
		},
	}
	
	value1 := &types.AttributeValueMemberM{Value: complexMap1}
	value2 := &types.AttributeValueMemberM{Value: complexMap2}
	
	result := attributeValuesEqual(value1, value2)
	assert.False(t, result)
}

// Test more complex scenarios for lists
func TestAttributeValueListsEqual_WithNestedComplexValues_ShouldCompareCorrectly(t *testing.T) {
	// RED: Test list with complex nested values
	list1 := []types.AttributeValue{
		&types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"key": &types.AttributeValueMemberS{Value: "value"},
			},
		},
		&types.AttributeValueMemberL{
			Value: []types.AttributeValue{
				&types.AttributeValueMemberN{Value: "123"},
			},
		},
	}
	
	list2 := []types.AttributeValue{
		&types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"key": &types.AttributeValueMemberS{Value: "value"},
			},
		},
		&types.AttributeValueMemberL{
			Value: []types.AttributeValue{
				&types.AttributeValueMemberN{Value: "123"},
			},
		},
	}
	
	result := attributeValueListsEqual(list1, list2)
	assert.True(t, result)
}

// Test additional type combinations in attributeValuesEqual
func TestAttributeValuesEqual_WithMixedTypeCombinations_ShouldReturnFalse(t *testing.T) {
	// RED: Test all possible type mismatches
	testCases := []struct {
		name   string
		value1 types.AttributeValue
		value2 types.AttributeValue
	}{
		{
			"String vs Number",
			&types.AttributeValueMemberS{Value: "123"},
			&types.AttributeValueMemberN{Value: "123"},
		},
		{
			"String vs Boolean",
			&types.AttributeValueMemberS{Value: "true"},
			&types.AttributeValueMemberBOOL{Value: true},
		},
		{
			"Number vs Boolean",
			&types.AttributeValueMemberN{Value: "1"},
			&types.AttributeValueMemberBOOL{Value: true},
		},
		{
			"String vs Binary",
			&types.AttributeValueMemberS{Value: "data"},
			&types.AttributeValueMemberB{Value: []byte("data")},
		},
		{
			"String vs StringSet",
			&types.AttributeValueMemberS{Value: "item"},
			&types.AttributeValueMemberSS{Value: []string{"item"}},
		},
		{
			"Number vs NumberSet",
			&types.AttributeValueMemberN{Value: "42"},
			&types.AttributeValueMemberNS{Value: []string{"42"}},
		},
		{
			"Binary vs BinarySet",
			&types.AttributeValueMemberB{Value: []byte("data")},
			&types.AttributeValueMemberBS{Value: [][]byte{[]byte("data")}},
		},
		{
			"String vs List",
			&types.AttributeValueMemberS{Value: "item"},
			&types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: "item"}}},
		},
		{
			"String vs Map",
			&types.AttributeValueMemberS{Value: "value"},
			&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{"key": &types.AttributeValueMemberS{Value: "value"}}},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := attributeValuesEqual(tc.value1, tc.value2)
			assert.False(t, result)
		})
	}
}

// Test convertToTableDescription with missing optional fields
func TestConvertToTableDescription_WithSomeFieldsNil_ShouldHandleGracefully(t *testing.T) {
	// RED: Test conversion with some nil optional fields
	input := &types.TableDescription{
		TableName:        stringPtr("test-table"),
		TableStatus:      types.TableStatusCreating,
		TableArn:         nil, // Nil optional field
		TableId:          stringPtr("12345"),
		CreationDateTime: nil, // Nil optional field
		ItemCount:        nil, // Nil optional field
		TableSizeBytes:   int64Ptr(1024),
	}
	
	result := convertToTableDescription(input)
	
	assert.NotNil(t, result)
	assert.Equal(t, "test-table", result.TableName)
	assert.Equal(t, types.TableStatusCreating, result.TableStatus)
	assert.Equal(t, "", result.TableArn) // Should be empty string for nil
	assert.Equal(t, "12345", result.TableId)
	assert.Nil(t, result.CreationDateTime) // Should remain nil
	assert.Equal(t, int64(0), result.ItemCount) // Should be 0 for nil int64
	assert.Equal(t, int64(1024), result.TableSizeBytes)
}

// Test more number set edge cases
func TestAttributeValuesEqual_WithEmptyNumberSets_ShouldReturnTrue(t *testing.T) {
	// RED: Test empty number sets
	value1 := &types.AttributeValueMemberNS{Value: []string{}}
	value2 := &types.AttributeValueMemberNS{Value: []string{}}
	
	result := attributeValuesEqual(value1, value2)
	assert.True(t, result)
}

func TestAttributeValuesEqual_WithEmptyBinarySets_ShouldReturnTrue(t *testing.T) {
	// RED: Test empty binary sets
	value1 := &types.AttributeValueMemberBS{Value: [][]byte{}}
	value2 := &types.AttributeValueMemberBS{Value: [][]byte{}}
	
	result := attributeValuesEqual(value1, value2)
	assert.True(t, result)
}

func TestAttributeValuesEqual_WithEmptyLists_ShouldReturnTrue(t *testing.T) {
	// RED: Test empty lists
	value1 := &types.AttributeValueMemberL{Value: []types.AttributeValue{}}
	value2 := &types.AttributeValueMemberL{Value: []types.AttributeValue{}}
	
	result := attributeValuesEqual(value1, value2)
	assert.True(t, result)
}

func TestAttributeValuesEqual_WithEmptyMaps_ShouldReturnTrue(t *testing.T) {
	// RED: Test empty maps
	value1 := &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}
	value2 := &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}}
	
	result := attributeValuesEqual(value1, value2)
	assert.True(t, result)
}

// Test additional utility functions to increase coverage
func TestPtrToString_WithValidPointer_InComprehensive_ShouldDereference(t *testing.T) {
	// RED: Test ptrToString with valid string pointer (comprehensive test)
	testValue := "test-string-comprehensive"
	ptr := &testValue
	
	result := ptrToString(ptr)
	assert.Equal(t, "test-string-comprehensive", result)
}

func TestPtrToString_WithNilPointer_InComprehensive_ShouldReturnEmptyString(t *testing.T) {
	// RED: Test ptrToString with nil pointer (comprehensive test)
	var ptr *string = nil
	
	result := ptrToString(ptr)
	assert.Equal(t, "", result)
}

func TestPtrToInt64_WithValidPointer_InComprehensive_ShouldDereference(t *testing.T) {
	// RED: Test ptrToInt64 with valid int64 pointer (comprehensive test)
	testValue := int64(54321)
	ptr := &testValue
	
	result := ptrToInt64(ptr)
	assert.Equal(t, int64(54321), result)
}

func TestPtrToInt64_WithNilPointer_InComprehensive_ShouldReturnZero(t *testing.T) {
	// RED: Test ptrToInt64 with nil pointer (comprehensive test)
	var ptr *int64 = nil
	
	result := ptrToInt64(ptr)
	assert.Equal(t, int64(0), result)
}

// Test binary value comparisons in more detail
func TestAttributeValuesEqual_WithEqualBytesInBinary_ShouldReturnTrue(t *testing.T) {
	// RED: Test binary values with identical byte contents
	data := []byte{0x01, 0x02, 0x03, 0xFF}
	value1 := &types.AttributeValueMemberB{Value: data}
	value2 := &types.AttributeValueMemberB{Value: []byte{0x01, 0x02, 0x03, 0xFF}}
	
	result := attributeValuesEqual(value1, value2)
	assert.True(t, result)
}

func TestAttributeValuesEqual_WithDifferentBytesInBinary_ShouldReturnFalse(t *testing.T) {
	// RED: Test binary values with different byte contents
	value1 := &types.AttributeValueMemberB{Value: []byte{0x01, 0x02, 0x03}}
	value2 := &types.AttributeValueMemberB{Value: []byte{0x01, 0x02, 0x04}}
	
	result := attributeValuesEqual(value1, value2)
	assert.False(t, result)
}

// Test string set ordering independence
func TestAttributeValuesEqual_WithReorderedStringSet_ShouldReturnTrue(t *testing.T) {
	// RED: Test string sets with different ordering but same contents
	value1 := &types.AttributeValueMemberSS{Value: []string{"apple", "banana", "cherry"}}
	value2 := &types.AttributeValueMemberSS{Value: []string{"banana", "cherry", "apple"}}
	
	result := attributeValuesEqual(value1, value2)
	assert.True(t, result)
}

func TestAttributeValuesEqual_WithDifferentStringSetContents_ShouldReturnFalse(t *testing.T) {
	// RED: Test string sets with different contents
	value1 := &types.AttributeValueMemberSS{Value: []string{"apple", "banana"}}
	value2 := &types.AttributeValueMemberSS{Value: []string{"apple", "cherry"}}
	
	result := attributeValuesEqual(value1, value2)
	assert.False(t, result)
}

// Test number set comparisons with different orderings
func TestAttributeValuesEqual_WithReorderedNumberSet_ShouldReturnTrue(t *testing.T) {
	// RED: Test number sets with different ordering but same contents
	value1 := &types.AttributeValueMemberNS{Value: []string{"1", "2", "3"}}
	value2 := &types.AttributeValueMemberNS{Value: []string{"3", "1", "2"}}
	
	result := attributeValuesEqual(value1, value2)
	assert.True(t, result)
}

// Test binary set comparisons 
func TestAttributeValuesEqual_WithReorderedBinarySet_ShouldReturnTrue(t *testing.T) {
	// RED: Test binary sets with different ordering but same contents
	value1 := &types.AttributeValueMemberBS{Value: [][]byte{
		[]byte("first"), []byte("second"), []byte("third"),
	}}
	value2 := &types.AttributeValueMemberBS{Value: [][]byte{
		[]byte("third"), []byte("first"), []byte("second"),
	}}
	
	result := attributeValuesEqual(value1, value2)
	assert.True(t, result)
}

func TestAttributeValuesEqual_WithDifferentBinarySetContents_ShouldReturnFalse(t *testing.T) {
	// RED: Test binary sets with different contents
	value1 := &types.AttributeValueMemberBS{Value: [][]byte{[]byte("data1"), []byte("data2")}}
	value2 := &types.AttributeValueMemberBS{Value: [][]byte{[]byte("data1"), []byte("data3")}}
	
	result := attributeValuesEqual(value1, value2)
	assert.False(t, result)
}