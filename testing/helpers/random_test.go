package helpers

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// RED: Test that UniqueID generates a unique identifier
func TestUniqueID_ShouldGenerateUniqueIdentifier(t *testing.T) {
	// Act
	id1 := UniqueID()
	id2 := UniqueID()
	
	// Assert
	assert.NotEqual(t, id1, id2, "UniqueID should generate different identifiers")
	assert.NotEmpty(t, id1, "UniqueID should not return empty string")
	assert.NotEmpty(t, id2, "UniqueID should not return empty string")
}

// RED: Test that RandomString generates string of correct length
func TestRandomString_ShouldGenerateStringOfSpecifiedLength(t *testing.T) {
	// Arrange
	testCases := []struct {
		name   string
		length int
	}{
		{"short string", 5},
		{"medium string", 10},
		{"long string", 20},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := RandomString(tc.length)
			
			// Assert
			assert.Len(t, result, tc.length, "RandomString should generate string of exact length")
			assert.NotEmpty(t, result, "RandomString should not return empty string")
		})
	}
}

// RED: Test that RandomString generates different strings
func TestRandomString_ShouldGenerateDifferentStrings(t *testing.T) {
	// Arrange
	length := 10
	
	// Act
	str1 := RandomString(length)
	str2 := RandomString(length)
	
	// Assert
	assert.NotEqual(t, str1, str2, "RandomString should generate different strings")
}

// RED: Test that RandomStringLower generates lowercase strings
func TestRandomStringLower_ShouldGenerateLowercaseStrings(t *testing.T) {
	// Arrange
	length := 10
	
	// Act
	result := RandomStringLower(length)
	
	// Assert
	assert.Len(t, result, length, "RandomStringLower should generate string of exact length")
	assert.Equal(t, strings.ToLower(result), result, "RandomStringLower should generate only lowercase characters")
	assert.NotEmpty(t, result, "RandomStringLower should not return empty string")
}

// RED: Test that RandomInt generates number within range
func TestRandomInt_ShouldGenerateNumberInRange(t *testing.T) {
	// Arrange
	min := 10
	max := 20
	
	// Act
	result := RandomInt(min, max)
	
	// Assert
	assert.GreaterOrEqual(t, result, min, "RandomInt should generate number >= min")
	assert.LessOrEqual(t, result, max, "RandomInt should generate number <= max")
}

// RED: Test that RandomInt generates different numbers
func TestRandomInt_ShouldGenerateDifferentNumbers(t *testing.T) {
	// Arrange
	min := 1
	max := 1000
	results := make(map[int]bool)
	iterations := 10
	
	// Act
	for i := 0; i < iterations; i++ {
		result := RandomInt(min, max)
		results[result] = true
	}
	
	// Assert
	assert.Greater(t, len(results), 1, "RandomInt should generate different numbers across multiple calls")
}

// RED: Test that RandomAWSResourceName generates valid AWS resource name
func TestRandomAWSResourceName_ShouldGenerateValidResourceName(t *testing.T) {
	// Arrange
	prefix := "test"
	
	// Act
	result := RandomAWSResourceName(prefix)
	
	// Assert
	assert.True(t, strings.HasPrefix(result, prefix), "Resource name should start with prefix")
	assert.Contains(t, result, "-", "Resource name should contain separator")
	assert.True(t, len(result) > len(prefix), "Resource name should be longer than prefix")
	
	// Validate AWS naming conventions (alphanumeric and hyphens only)
	for _, char := range result {
		isValid := (char >= 'a' && char <= 'z') || 
				  (char >= 'A' && char <= 'Z') || 
				  (char >= '0' && char <= '9') || 
				  char == '-'
		assert.True(t, isValid, "Resource name should contain only valid AWS characters")
	}
}

// RED: Test that RandomAWSResourceName generates different names
func TestRandomAWSResourceName_ShouldGenerateDifferentNames(t *testing.T) {
	// Arrange
	prefix := "test"
	
	// Act
	name1 := RandomAWSResourceName(prefix)
	name2 := RandomAWSResourceName(prefix)
	
	// Assert
	assert.NotEqual(t, name1, name2, "RandomAWSResourceName should generate different names")
}

// RED: Test that SeedRandom sets deterministic behavior
func TestSeedRandom_ShouldProduceDeterministicResults(t *testing.T) {
	// Arrange
	seed := int64(12345)
	
	// Act
	SeedRandom(seed)
	result1 := RandomString(10)
	
	SeedRandom(seed)
	result2 := RandomString(10)
	
	// Assert
	assert.Equal(t, result1, result2, "SeedRandom should produce deterministic results")
}

// RED: Test that timestamp-based IDs include time component
func TestTimestampID_ShouldIncludeTimeComponent(t *testing.T) {
	// Act
	id := TimestampID()
	
	// Assert
	assert.NotEmpty(t, id, "TimestampID should not return empty string")
	
	// Should contain timestamp information (basic check)
	assert.Greater(t, len(id), 10, "TimestampID should be reasonably long")
}

// RED: Test that TimestampID generates unique IDs
func TestTimestampID_ShouldGenerateUniqueIDs(t *testing.T) {
	// Act
	id1 := TimestampID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := TimestampID()
	
	// Assert
	assert.NotEqual(t, id1, id2, "TimestampID should generate unique IDs")
}

// RED: Test edge cases for random generation functions
func TestRandomGeneration_ShouldHandleEdgeCases(t *testing.T) {
	t.Run("RandomString with length 0", func(t *testing.T) {
		result := RandomString(0)
		assert.Empty(t, result, "RandomString with length 0 should return empty string")
	})
	
	t.Run("RandomString with length 1", func(t *testing.T) {
		result := RandomString(1)
		assert.Len(t, result, 1, "RandomString with length 1 should return 1 character")
	})
	
	t.Run("RandomInt with equal min and max", func(t *testing.T) {
		value := 42
		result := RandomInt(value, value)
		assert.Equal(t, value, result, "RandomInt with equal min and max should return that value")
	})
}