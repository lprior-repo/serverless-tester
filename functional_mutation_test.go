package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
)

// Simple functional testing example that demonstrates mutation testing validation
func TestFunctionalValidation(t *testing.T) {
	t.Run("validates configuration creation", func(t *testing.T) {
		config := createTestConfig("test-function", 30*time.Second, true)
		assert.Equal(t, "test-function", config.name)
		assert.Equal(t, 30*time.Second, config.timeout)
		assert.True(t, config.enabled)
	})

	t.Run("validates immutable operations", func(t *testing.T) {
		configs := []TestConfig{
			createTestConfig("config1", 10*time.Second, true),
			createTestConfig("config2", 20*time.Second, false),
			createTestConfig("", 30*time.Second, true), // Invalid
		}

		validConfigs := filterValidConfigs(configs)
		assert.Len(t, validConfigs, 2)

		processedResults := processConfigs(validConfigs)
		successCount := countSuccessful(processedResults)
		assert.Equal(t, 2, successCount)
	})

	t.Run("validates monadic operations", func(t *testing.T) {
		result := createResult(true, "success", nil)
		assert.True(t, result.IsSuccess())

		data := result.GetData()
		assert.True(t, data.IsPresent())
		assert.Equal(t, "success", data.MustGet())

		errorResult := createResult(false, nil, fmt.Errorf("test error"))
		assert.False(t, errorResult.IsSuccess())
		assert.True(t, errorResult.GetError().IsPresent())
	})

	t.Run("validates function composition", func(t *testing.T) {
		names := []string{"a", "bb", "ccc", "", "dddd"}
		processed := processNamesPipeline(names)

		// Should filter out empty, then add prefixes, then count
		assert.Equal(t, 4, processed)
	})
}

// Test data structures
type TestConfig struct {
	name    string
	timeout time.Duration
	enabled bool
}

type TestResult struct {
	data  mo.Option[interface{}]
	error mo.Option[error]
}

func (r *TestResult) IsSuccess() bool {
	return r.data.IsPresent() && r.error.IsAbsent()
}

func (r *TestResult) GetData() mo.Option[interface{}] {
	return r.data
}

func (r *TestResult) GetError() mo.Option[error] {
	return r.error
}

// Pure functions for testing
func createTestConfig(name string, timeout time.Duration, enabled bool) TestConfig {
	return TestConfig{
		name:    name,
		timeout: timeout,
		enabled: enabled,
	}
}

func validateConfig(config TestConfig) bool {
	return config.name != "" && config.timeout > 0
}

func filterValidConfigs(configs []TestConfig) []TestConfig {
	return lo.Filter(configs, func(config TestConfig, _ int) bool {
		return validateConfig(config)
	})
}

func processConfig(config TestConfig) *TestResult {
	if !validateConfig(config) {
		return &TestResult{
			data:  mo.None[interface{}](),
			error: mo.Some[error](fmt.Errorf("invalid config")),
		}
	}

	processedData := map[string]interface{}{
		"name":         config.name,
		"timeout_sec":  config.timeout.Seconds(),
		"enabled":      config.enabled,
		"processed_at": time.Now().Unix(),
	}

	return &TestResult{
		data:  mo.Some[interface{}](processedData),
		error: mo.None[error](),
	}
}

func processConfigs(configs []TestConfig) []*TestResult {
	return lo.Map(configs, func(config TestConfig, _ int) *TestResult {
		return processConfig(config)
	})
}

func countSuccessful(results []*TestResult) int {
	return lo.CountBy(results, func(result *TestResult) bool {
		return result.IsSuccess()
	})
}

func createResult(success bool, data interface{}, err error) *TestResult {
	if success {
		return &TestResult{
			data:  mo.Some[interface{}](data),
			error: mo.None[error](),
		}
	}
	return &TestResult{
		data:  mo.None[interface{}](),
		error: mo.Some[error](err),
	}
}

// Function composition example
func processNamesPipeline(names []string) int {
	// Step 1: Filter out empty strings
	filtered := lo.Filter(names, func(name string, _ int) bool {
		return name != ""
	})

	// Step 2: Add prefix to each name
	prefixed := lo.Map(filtered, func(name string, _ int) string {
		return "processed_" + name
	})

	// Step 3: Count results
	return len(prefixed)
}

func main() {
	// Run a quick validation
	fmt.Println("🧪 Functional Mutation Test Validation")

	// Create test configurations
	configs := []TestConfig{
		createTestConfig("func1", 10*time.Second, true),
		createTestConfig("func2", 20*time.Second, false),
		createTestConfig("", 30*time.Second, true), // Invalid
		createTestConfig("func3", 40*time.Second, true),
	}

	fmt.Printf("✅ Created %d test configurations\n", len(configs))

	// Filter valid configurations
	validConfigs := filterValidConfigs(configs)
	fmt.Printf("✅ Filtered to %d valid configurations\n", len(validConfigs))

	// Process configurations
	results := processConfigs(validConfigs)
	successCount := countSuccessful(results)
	fmt.Printf("✅ Processed configurations: %d successful\n", successCount)

	// Test monadic operations
	successResult := createResult(true, "test data", nil)
	if successResult.IsSuccess() {
		fmt.Println("✅ Monadic success result validation passed")
	}

	errorResult := createResult(false, nil, fmt.Errorf("test error"))
	if !errorResult.IsSuccess() {
		fmt.Println("✅ Monadic error result validation passed")
	}

	// Test function composition
	testNames := []string{"alice", "bob", "", "charlie"}
	processed := processNamesPipeline(testNames)
	fmt.Printf("✅ Function composition result: %d processed names\n", processed)

	fmt.Println("🎉 Functional mutation test validation completed successfully!")
}