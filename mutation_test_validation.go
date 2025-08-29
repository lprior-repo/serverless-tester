package main

import (
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// FunctionalConfigExample demonstrates functional programming patterns
type FunctionalConfigExample struct {
	name     string
	timeout  time.Duration
	enabled  bool
	metadata map[string]interface{}
}

// NewFunctionalConfigExample creates immutable configuration
func NewFunctionalConfigExample(options ...func(*FunctionalConfigExample)) *FunctionalConfigExample {
	config := &FunctionalConfigExample{
		name:     "default",
		timeout:  30 * time.Second,
		enabled:  true,
		metadata: make(map[string]interface{}),
	}

	for _, option := range options {
		option(config)
	}

	return config
}

// Functional options
func WithName(name string) func(*FunctionalConfigExample) {
	return func(c *FunctionalConfigExample) {
		c.name = name
	}
}

func WithTimeout(timeout time.Duration) func(*FunctionalConfigExample) {
	return func(c *FunctionalConfigExample) {
		c.timeout = timeout
	}
}

func WithEnabled(enabled bool) func(*FunctionalConfigExample) {
	return func(c *FunctionalConfigExample) {
		c.enabled = enabled
	}
}

func WithMetadata(key string, value interface{}) func(*FunctionalConfigExample) {
	return func(c *FunctionalConfigExample) {
		c.metadata[key] = value
	}
}

// Getters for immutable access
func (c *FunctionalConfigExample) GetName() string {
	return c.name
}

func (c *FunctionalConfigExample) GetTimeout() time.Duration {
	return c.timeout
}

func (c *FunctionalConfigExample) IsEnabled() bool {
	return c.enabled
}

func (c *FunctionalConfigExample) GetMetadata() map[string]interface{} {
	return c.metadata
}

// FunctionalResult represents a monadic result
type FunctionalResult struct {
	result   mo.Option[interface{}]
	error    mo.Option[error]
	duration time.Duration
}

func (r *FunctionalResult) IsSuccess() bool {
	return r.result.IsPresent() && r.error.IsAbsent()
}

func (r *FunctionalResult) GetResult() mo.Option[interface{}] {
	return r.result
}

func (r *FunctionalResult) GetError() mo.Option[error] {
	return r.error
}

func (r *FunctionalResult) GetDuration() time.Duration {
	return r.duration
}

// Pure functions for validation and processing
func validateConfigName(name string) mo.Option[error] {
	if name == "" {
		return mo.Some[error](fmt.Errorf("name cannot be empty"))
	}
	if len(name) > 100 {
		return mo.Some[error](fmt.Errorf("name too long"))
	}
	return mo.None[error]()
}

func validateTimeout(timeout time.Duration) mo.Option[error] {
	if timeout <= 0 {
		return mo.Some[error](fmt.Errorf("timeout must be positive"))
	}
	if timeout > time.Hour {
		return mo.Some[error](fmt.Errorf("timeout too long"))
	}
	return mo.None[error]()
}

func validateEnabled(enabled bool) mo.Option[error] {
	// All boolean values are valid
	return mo.None[error]()
}

// Functional validation pipeline
func ValidateConfiguration(config *FunctionalConfigExample) mo.Option[error] {
	validations := []func() mo.Option[error]{
		func() mo.Option[error] { return validateConfigName(config.GetName()) },
		func() mo.Option[error] { return validateTimeout(config.GetTimeout()) },
		func() mo.Option[error] { return validateEnabled(config.IsEnabled()) },
	}

	for _, validate := range validations {
		if err := validate(); err.IsPresent() {
			return err
		}
	}

	return mo.None[error]()
}

// Functional processing with monadic operations
func ProcessConfiguration(config *FunctionalConfigExample) *FunctionalResult {
	startTime := time.Now()

	// Validation pipeline
	validationError := ValidateConfiguration(config)
	if validationError.IsPresent() {
		return &FunctionalResult{
			result:   mo.None[interface{}](),
			error:    validationError,
			duration: time.Since(startTime),
		}
	}

	// Processing pipeline
	processedData := map[string]interface{}{
		"name":         config.GetName(),
		"timeout":      config.GetTimeout().Seconds(),
		"enabled":      config.IsEnabled(),
		"metadata":     config.GetMetadata(),
		"processed_at": time.Now().Unix(),
	}

	return &FunctionalResult{
		result:   mo.Some[interface{}](processedData),
		error:    mo.None[error](),
		duration: time.Since(startTime),
	}
}

// Function composition examples
func createConfigurations(names []string) []FunctionalConfigExample {
	return lo.Map(names, func(name string, _ int) FunctionalConfigExample {
		return *NewFunctionalConfigExample(
			WithName(name),
			WithTimeout(30*time.Second),
			WithEnabled(true),
		)
	})
}

func filterValidConfigurations(configs []FunctionalConfigExample) []FunctionalConfigExample {
	return lo.Filter(configs, func(config FunctionalConfigExample, _ int) bool {
		return ValidateConfiguration(&config).IsAbsent()
	})
}

func processConfigurations(configs []FunctionalConfigExample) []*FunctionalResult {
	return lo.Map(configs, func(config FunctionalConfigExample, _ int) *FunctionalResult {
		return ProcessConfiguration(&config)
	})
}

// Pipeline composition - manual composition without lo.Pipe3
func ProcessConfigurationPipeline(names []string) []*FunctionalResult {
	step1 := createConfigurations(names)
	step2 := filterValidConfigurations(step1)
	step3 := processConfigurations(step2)
	return step3
}

// Monadic operations example
func ExtractSuccessfulResults(results []*FunctionalResult) []interface{} {
	return lo.FilterMap(results, func(result *FunctionalResult, _ int) (interface{}, bool) {
		if result.IsSuccess() && result.GetResult().IsPresent() {
			return result.GetResult().MustGet(), true
		}
		return nil, false
	})
}

// Pure function for counting successes
func CountSuccessfulResults(results []*FunctionalResult) int {
	return lo.CountBy(results, func(result *FunctionalResult) bool {
		return result.IsSuccess()
	})
}

func main() {
	fmt.Println("🧪 Functional Programming Mutation Testing Validation")

	// Test functional configuration creation
	config := NewFunctionalConfigExample(
		WithName("test-config"),
		WithTimeout(45*time.Second),
		WithEnabled(true),
		WithMetadata("environment", "test"),
	)

	fmt.Printf("✅ Created configuration: %s\n", config.GetName())

	// Test validation
	validationResult := ValidateConfiguration(config)
	if validationResult.IsPresent() {
		fmt.Printf("❌ Validation error: %v\n", validationResult.MustGet())
	} else {
		fmt.Println("✅ Configuration validation passed")
	}

	// Test processing
	result := ProcessConfiguration(config)
	if result.IsSuccess() {
		fmt.Printf("✅ Processing succeeded in %v\n", result.GetDuration())
		if data := result.GetResult(); data.IsPresent() {
			fmt.Printf("✅ Result data: %+v\n", data.MustGet())
		}
	}

	// Test pipeline processing
	names := []string{"config1", "config2", "", "config3", "very-long-configuration-name-that-exceeds-the-maximum-allowed-length-limit"}
	results := ProcessConfigurationPipeline(names)

	successCount := CountSuccessfulResults(results)
	fmt.Printf("✅ Pipeline processed %d configurations, %d successful\n", len(results), successCount)

	// Test monadic extraction
	successfulData := ExtractSuccessfulResults(results)
	fmt.Printf("✅ Extracted %d successful results\n", len(successfulData))

	// Function composition validation - manual composition
	step1Pipeline := lo.Filter(names, func(name string, _ int) bool {
		return name != ""
	})
	step2Pipeline := lo.Map(step1Pipeline, func(name string, _ int) string {
		if len(name) > 50 {
			return name[:50]
		}
		return name
	})
	fmt.Printf("✅ Function composition result: %d valid names\n", len(step2Pipeline))

	fmt.Println("🎉 Functional programming validation completed successfully!")
}