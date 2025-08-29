package main

import (
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// Test data structures for mutation testing
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

// Pure functions for testing - these will be mutated
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

// Critical function for mutation testing - this has conditional logic that should be tested
func processNamesPipeline(names []string) int {
	// This function has multiple steps that can be mutated
	if len(names) == 0 {
		return 0
	}

	// Step 1: Filter out empty strings
	filtered := lo.Filter(names, func(name string, _ int) bool {
		return name != ""
	})

	if len(filtered) == 0 {
		return 0
	}

	// Step 2: Add prefix to each name
	prefixed := lo.Map(filtered, func(name string, _ int) string {
		return "processed_" + name
	})

	// Step 3: Count results
	return len(prefixed)
}

// Complex function with multiple conditional branches for mutation testing
func validateAndProcessBatch(configs []TestConfig, minSize int) (int, error) {
	if len(configs) < minSize {
		return 0, fmt.Errorf("batch too small: need at least %d configs", minSize)
	}

	validCount := 0
	processedCount := 0

	for _, config := range configs {
		if validateConfig(config) {
			validCount++
			
			result := processConfig(config)
			if result.IsSuccess() {
				processedCount++
			}
		}
	}

	if processedCount == 0 {
		return 0, fmt.Errorf("no configurations were successfully processed")
	}

	return processedCount, nil
}

func main() {
	// This main function tests all our functions
	configs := []TestConfig{
		createTestConfig("test1", 10*time.Second, true),
		createTestConfig("test2", 20*time.Second, false),
		createTestConfig("", 30*time.Second, true), // Invalid
		createTestConfig("test3", 40*time.Second, true),
	}

	validConfigs := filterValidConfigs(configs)
	results := processConfigs(validConfigs)
	successCount := countSuccessful(results)

	processed, err := validateAndProcessBatch(configs, 2)
	
	pipelineResult := processNamesPipeline([]string{"alice", "bob", "", "charlie"})

	fmt.Printf("Valid configs: %d, Success count: %d, Processed: %d, Pipeline: %d\n", 
		len(validConfigs), successCount, processed, pipelineResult)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}