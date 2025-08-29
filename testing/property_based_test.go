package testutils

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// ========== TDD: RED PHASE - Property-Based Testing Tests ==========

// RED: Test property-based testing framework exists and works
func TestPropertyBasedTesting_ShouldGenerateRandomData(t *testing.T) {
	propTest := NewPropertyBasedTest()
	
	// Test string generator
	stringGen := propTest.StringGenerator(1, 10)
	value := stringGen.Generate()
	assert.NotEmpty(t, value, "Generated string should not be empty")
	assert.True(t, len(value) >= 1 && len(value) <= 10, "Generated string should be within bounds")
	
	// Test integer generator
	intGen := propTest.IntGenerator(-100, 100)
	intValue := intGen.Generate()
	assert.True(t, intValue >= -100 && intValue <= 100, "Generated int should be within bounds")
}

// RED: Test property validation with sophisticated generators
func TestPropertyBasedTesting_ShouldValidateProperties(t *testing.T) {
	propTest := NewPropertyBasedTest()
	
	// Property: String length should equal actual length
	stringLengthProperty := func(str string) bool {
		return len(str) == len([]rune(str))
	}
	
	result := propTest.CheckProperty(
		"String length property",
		propTest.StringGenerator(1, 100),
		stringLengthProperty,
		100, // iterations
	)
	
	assert.True(t, result.Passed, "String length property should pass")
	assert.Equal(t, 100, result.TestsRun, "Should run all iterations")
}

// RED: Test AWS-specific property generators
func TestPropertyBasedTesting_ShouldGenerateAWSResources(t *testing.T) {
	propTest := NewPropertyBasedTest()
	
	// Test ARN generator
	arnGen := propTest.ArnGenerator()
	arn := arnGen.Generate()
	assert.Contains(t, arn, "arn:aws:", "Generated ARN should start with AWS prefix")
	assert.True(t, len(arn) > 20, "ARN should have reasonable length")
	
	// Test Lambda function name generator
	lambdaGen := propTest.LambdaFunctionNameGenerator()
	name := lambdaGen.Generate()
	assert.True(t, len(name) >= 1 && len(name) <= 64, "Lambda name should be within AWS limits")
	assert.NotContains(t, name, " ", "Lambda name should not contain spaces")
}

// RED: Test composite generators for complex scenarios
func TestPropertyBasedTesting_ShouldHandleCompositeGenerators(t *testing.T) {
	propTest := NewPropertyBasedTest()
	
	// Test event generator that combines multiple properties
	eventGen := propTest.EventGenerator()
	event := eventGen.Generate()
	assert.NotEmpty(t, event.EventID, "Event should have ID")
	assert.NotEmpty(t, event.Source, "Event should have source")
	assert.True(t, event.Timestamp > 0, "Event should have timestamp")
}

// RED: Test shrinking functionality when property fails
func TestPropertyBasedTesting_ShouldShrinkFailingCases(t *testing.T) {
	propTest := NewPropertyBasedTest()
	
	// Property that fails for strings longer than 5 characters
	shortStringProperty := func(str string) bool {
		return len(str) <= 5
	}
	
	result := propTest.CheckProperty(
		"Short string property",
		propTest.StringGenerator(1, 20),
		shortStringProperty,
		50,
	)
	
	assert.False(t, result.Passed, "Property should fail for long strings")
	assert.NotEmpty(t, result.CounterExample, "Should provide counter example")
	assert.NotEmpty(t, result.ShrunkExample, "Should provide shrunk example")
}