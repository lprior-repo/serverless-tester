package dsl

import (
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// TestSimpleFunctionalConfig demonstrates functional configuration
func TestSimpleFunctionalConfig(t *testing.T) {
	// Create config using functional options
	config := NewSimpleFunctionalConfig("test-function",
		WithSimpleTimeout(45*time.Second),
		WithSimpleMetadata("environment", "test"),
		WithSimpleMetadata("version", "1.0"),
	)

	if config.name != "test-function" {
		t.Errorf("Expected name 'test-function', got '%s'", config.name)
	}

	if config.timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", config.timeout)
	}

	if config.metadata["environment"] != "test" {
		t.Errorf("Expected environment 'test', got %v", config.metadata["environment"])
	}

	if config.metadata["version"] != "1.0" {
		t.Errorf("Expected version '1.0', got %v", config.metadata["version"])
	}
}

// TestFunctionalLambdaInvocation demonstrates functional Lambda operations
func TestFunctionalLambdaInvocation(t *testing.T) {
	config := NewSimpleFunctionalConfig("my-function",
		WithSimpleMetadata("test_type", "functional"),
	)

	payload := map[string]interface{}{
		"message": "Hello from functional programming",
		"numbers": []int{1, 2, 3, 4, 5},
	}

	result := FunctionalLambdaInvocation("my-function", payload, config)

	if !result.IsSuccess() {
		t.Errorf("Expected successful result, got error: %v", result.GetError())
	}

	if result.GetValue().IsAbsent() {
		t.Error("Expected value to be present")
	}

	// Test functional mapping
	mappedResult := result.MapValue(func(response string) string {
		return "transformed: " + response
	})

	if !mappedResult.IsSuccess() {
		t.Error("Expected mapped result to be successful")
	}

	if mappedResult.GetValue().IsAbsent() {
		t.Error("Expected mapped value to be present")
	}
}

// TestFunctionalLambdaDisabled demonstrates error handling with disabled operations
func TestFunctionalLambdaDisabled(t *testing.T) {
	config := NewSimpleFunctionalConfig("disabled-function",
		WithDisabled(),
	)

	result := FunctionalLambdaInvocation("disabled-function", nil, config)

	if result.IsSuccess() {
		t.Error("Expected disabled operation to fail")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present for disabled operation")
	}
}

// TestFunctionalDynamoDBOperation demonstrates functional DynamoDB operations
func TestFunctionalDynamoDBOperation(t *testing.T) {
	config := NewSimpleFunctionalConfig("users-table",
		WithSimpleMetadata("operation_type", "query"),
	)

	result := FunctionalDynamoDBOperation("users", "query", config)

	if !result.IsSuccess() {
		t.Errorf("Expected successful result, got error: %v", result.GetError())
	}

	if result.GetValue().IsAbsent() {
		t.Error("Expected value to be present")
	}

	// Verify result structure
	if result.IsSuccess() && result.GetValue().IsPresent() {
		data := result.GetValue().MustGet()
		if data["table_name"] != "users" {
			t.Errorf("Expected table_name 'users', got %v", data["table_name"])
		}
		if data["operation"] != "query" {
			t.Errorf("Expected operation 'query', got %v", data["operation"])
		}
	}
}

// TestFunctionalOperations demonstrates various functional programming utilities
func TestFunctionalOperations(t *testing.T) {
	data := []string{"apple", "banana", "cat", "dog", "elephant", "fox"}

	// Test functional filter
	filtered := FunctionalFilter(data, func(item string) bool {
		return len(item) > 4
	})

	expected := []string{"apple", "banana", "elephant"}
	if len(filtered) != len(expected) {
		t.Errorf("Expected %d filtered items, got %d", len(expected), len(filtered))
	}

	// Test functional map
	mapped := FunctionalMap(data, func(item string, index int) string {
		return fmt.Sprintf("%d:%s", index, item)
	})

	if len(mapped) != len(data) {
		t.Errorf("Expected %d mapped items, got %d", len(data), len(mapped))
	}

	if mapped[0] != "0:apple" {
		t.Errorf("Expected '0:apple', got '%s'", mapped[0])
	}

	// Test functional reduce
	totalLength := FunctionalReduce(data, func(acc int, item string, _ int) int {
		return acc + len(item)
	}, 0)

	expectedTotal := len("apple") + len("banana") + len("cat") + len("dog") + len("elephant") + len("fox")
	if totalLength != expectedTotal {
		t.Errorf("Expected total length %d, got %d", expectedTotal, totalLength)
	}

	// Test functional partition
	long, short := FunctionalPartition(data, func(item string) bool {
		return len(item) > 4
	})

	if len(long) != 3 {
		t.Errorf("Expected 3 long items, got %d", len(long))
	}

	if len(short) != 3 {
		t.Errorf("Expected 3 short items, got %d", len(short))
	}

	// Test functional group by
	groups := FunctionalGroupBy(data, func(item string) int {
		return len(item)
	})

	if len(groups) == 0 {
		t.Error("Expected grouped items")
	}

	// Test functional unique
	duplicateData := append(data, "apple", "banana")
	unique := FunctionalUnique(duplicateData)

	if len(unique) != len(data) {
		t.Errorf("Expected %d unique items, got %d", len(data), len(unique))
	}

	// Test functional find
	found := FunctionalFind(data, func(item string) bool {
		return item == "elephant"
	})

	if found.IsAbsent() {
		t.Error("Expected to find 'elephant'")
	}

	if found.IsPresent() && found.MustGet() != "elephant" {
		t.Errorf("Expected 'elephant', got '%s'", found.MustGet())
	}

	// Test functional every
	allStrings := FunctionalEvery(data, func(item string) bool {
		return len(item) > 0
	})

	if !allStrings {
		t.Error("Expected all items to be non-empty strings")
	}

	// Test functional some
	hasLongString := FunctionalSome(data, func(item string) bool {
		return len(item) > 6
	})

	if !hasLongString {
		t.Error("Expected some items to be longer than 6 characters")
	}
}

// TestFunctionalPipeline demonstrates functional pipeline composition
func TestFunctionalPipeline(t *testing.T) {
	input := "hello"

	// Create pipeline operations
	addPrefix := func(s string) string { return "prefix_" + s }
	addSuffix := func(s string) string { return s + "_suffix" }
	toUpper := func(s string) string { return fmt.Sprintf("UPPER(%s)", s) }

	// Apply pipeline
	result := FunctionalPipeline(input, addPrefix, addSuffix, toUpper)

	expected := "UPPER(prefix_hello_suffix)"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestComplexFunctionalOperation demonstrates complex functional composition
func TestComplexFunctionalOperation(t *testing.T) {
	data := []string{"apple", "banana", "cat", "dog", "elephant", "fox", "grape", "house"}

	result := ComplexFunctionalOperation(data)

	if !result.IsSuccess() {
		t.Errorf("Expected successful result, got error: %v", result.GetError())
	}

	if result.GetValue().IsAbsent() {
		t.Error("Expected value to be present")
	}

	// Verify processing results
	if result.IsSuccess() && result.GetValue().IsPresent() {
		data := result.GetValue().MustGet()

		originalCount, ok := data["original_count"].(int)
		if !ok || originalCount != 8 {
			t.Errorf("Expected original_count 8, got %v", data["original_count"])
		}

		// Verify that items longer than 3 characters were processed
		processedCount, ok := data["processed_count"].(int)
		if !ok || processedCount == 0 {
			t.Errorf("Expected processed items, got %v", data["processed_count"])
		}
	}
}

// TestFunctionalScenario demonstrates functional scenario composition
func TestFunctionalScenario(t *testing.T) {
	// Create scenario with functional steps
	scenario := NewFunctionalScenario[string]("Functional Test Scenario").
		AddStep(
			NewFunctionalScenarioStep("Step 1: Lambda invocation",
				func() SimpleResult[string] {
					config := NewSimpleFunctionalConfig("step1-function")
					result := FunctionalLambdaInvocation("step1-function", "test-data", config)
					return result
				},
			),
		).
		AddStep(
			NewFunctionalScenarioStep("Step 2: Data processing",
				func() SimpleResult[string] {
					return NewSuccessSimpleResult(
						"processed data",
						100*time.Millisecond,
						map[string]interface{}{"step": 2},
					)
				},
			),
		).
		AddStep(
			NewFunctionalScenarioStep("Step 3: Final validation",
				func() SimpleResult[string] {
					return NewSuccessSimpleResult(
						"validation passed",
						50*time.Millisecond,
						map[string]interface{}{"step": 3},
					)
				},
			),
		)

	// Execute scenario
	result := scenario.Execute()

	if !result.IsSuccess() {
		t.Errorf("Expected successful scenario, got error: %v", result.GetError())
	}

	if result.GetValue().IsAbsent() {
		t.Error("Expected scenario value to be present")
	}

	// Verify all steps executed
	if result.IsSuccess() && result.GetValue().IsPresent() {
		values := result.GetValue().MustGet()
		if len(values) != 3 {
			t.Errorf("Expected 3 step results, got %d", len(values))
		}
	}
}

// TestFunctionalScenarioWithError demonstrates error handling in scenarios
func TestFunctionalScenarioWithError(t *testing.T) {
	scenario := NewFunctionalScenario[string]("Error Scenario").
		AddStep(
			NewFunctionalScenarioStep("Success step",
				func() SimpleResult[string] {
					return NewSuccessSimpleResult(
						"success",
						10*time.Millisecond,
						map[string]interface{}{},
					)
				},
			),
		).
		AddStep(
			NewFunctionalScenarioStep("Error step",
				func() SimpleResult[string] {
					return NewErrorSimpleResult[string](
						fmt.Errorf("intentional error"),
						20*time.Millisecond,
						map[string]interface{}{},
					)
				},
			),
		).
		AddStep(
			NewFunctionalScenarioStep("This should not execute",
				func() SimpleResult[string] {
					t.Error("This step should not execute due to previous error")
					return NewSuccessSimpleResult("should not reach", 0, map[string]interface{}{})
				},
			),
		)

	result := scenario.Execute()

	if result.IsSuccess() {
		t.Error("Expected scenario to fail due to error step")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}
}

// TestOptionTypeUsage demonstrates mo.Option usage
func TestOptionTypeUsage(t *testing.T) {
	// Test Some value
	someValue := mo.Some("hello")
	if someValue.IsAbsent() {
		t.Error("Expected Some value to be present")
	}

	if someValue.MustGet() != "hello" {
		t.Errorf("Expected 'hello', got '%s'", someValue.MustGet())
	}

	// Test None value
	noneValue := mo.None[string]()
	if noneValue.IsPresent() {
		t.Error("Expected None value to be absent")
	}

	// Test Option mapping
	mappedSome := someValue.Map(func(s string) (string, bool) {
		return s + " world", true
	})

	if mappedSome.IsAbsent() {
		t.Error("Expected mapped Some to be present")
	}

	if mappedSome.MustGet() != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", mappedSome.MustGet())
	}

	// Test Option flat mapping
	flatMapped := someValue.FlatMap(func(s string) mo.Option[string] {
		if len(s) > 3 {
			return mo.Some(s + "!")
		}
		return mo.None[string]()
	})

	if flatMapped.IsAbsent() {
		t.Error("Expected flat mapped value to be present")
	}

	if flatMapped.MustGet() != "hello!" {
		t.Errorf("Expected 'hello!', got '%s'", flatMapped.MustGet())
	}
}

// TestLoUtilities demonstrates various lo utilities
func TestLoUtilities(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Test lo.Chunk
	chunks := lo.Chunk(numbers, 3)
	expectedChunks := 4 // [1,2,3], [4,5,6], [7,8,9], [10]
	if len(chunks) != expectedChunks {
		t.Errorf("Expected %d chunks, got %d", expectedChunks, len(chunks))
	}

	// Test lo.Reverse
	reversed := lo.Reverse(numbers)
	if reversed[0] != 10 || reversed[len(reversed)-1] != 1 {
		t.Error("Expected reversed array to start with 10 and end with 1")
	}

	// Test lo.Range
	rangeResult := lo.Range(5)
	expected := []int{0, 1, 2, 3, 4}
	if len(rangeResult) != len(expected) {
		t.Errorf("Expected range of length %d, got %d", len(expected), len(rangeResult))
	}

	// Test lo.Times
	timesResult := lo.Times(3, func(i int) string {
		return fmt.Sprintf("item-%d", i)
	})

	if len(timesResult) != 3 {
		t.Errorf("Expected 3 items, got %d", len(timesResult))
	}

	if timesResult[1] != "item-1" {
		t.Errorf("Expected 'item-1', got '%s'", timesResult[1])
	}

	// Test lo.Ternary
	ternaryResult := lo.Ternary(true, "yes", "no")
	if ternaryResult != "yes" {
		t.Errorf("Expected 'yes', got '%s'", ternaryResult)
	}

	ternaryResult2 := lo.Ternary(false, "yes", "no")
	if ternaryResult2 != "no" {
		t.Errorf("Expected 'no', got '%s'", ternaryResult2)
	}

	// Test lo.Must
	mustResult := lo.Must(func() (string, error) {
		return "success", nil
	}())

	if mustResult != "success" {
		t.Errorf("Expected 'success', got '%s'", mustResult)
	}
}

// BenchmarkFunctionalOperations benchmarks functional operations
func BenchmarkFunctionalOperations(b *testing.B) {
	data := lo.Range(1000)

	b.Run("FunctionalFilter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FunctionalFilter(data, func(item int) bool {
				return item%2 == 0
			})
		}
	})

	b.Run("FunctionalMap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FunctionalMap(data, func(item int, _ int) int {
				return item * 2
			})
		}
	})

	b.Run("FunctionalReduce", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FunctionalReduce(data, func(acc int, item int, _ int) int {
				return acc + item
			}, 0)
		}
	})
}