package dsl

import (
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// SimpleFunctionalConfig represents an immutable configuration using functional options
type SimpleFunctionalConfig struct {
	name     string
	timeout  time.Duration
	metadata map[string]interface{}
	enabled  bool
}

// ConfigOption represents a functional option for SimpleFunctionalConfig
type ConfigOption func(SimpleFunctionalConfig) SimpleFunctionalConfig

// NewSimpleFunctionalConfig creates a new config with functional options
func NewSimpleFunctionalConfig(name string, opts ...ConfigOption) SimpleFunctionalConfig {
	base := SimpleFunctionalConfig{
		name:     name,
		timeout:  30 * time.Second,
		metadata: make(map[string]interface{}),
		enabled:  true,
	}

	return lo.Reduce(opts, func(config SimpleFunctionalConfig, opt ConfigOption, _ int) SimpleFunctionalConfig {
		return opt(config)
	}, base)
}

// WithSimpleTimeout sets the timeout functionally
func WithSimpleTimeout(timeout time.Duration) ConfigOption {
	return func(config SimpleFunctionalConfig) SimpleFunctionalConfig {
		return SimpleFunctionalConfig{
			name:     config.name,
			timeout:  timeout,
			metadata: config.metadata,
			enabled:  config.enabled,
		}
	}
}

// WithSimpleMetadata adds metadata functionally
func WithSimpleMetadata(key string, value interface{}) ConfigOption {
	return func(config SimpleFunctionalConfig) SimpleFunctionalConfig {
		newMetadata := lo.Assign(config.metadata, map[string]interface{}{key: value})
		return SimpleFunctionalConfig{
			name:     config.name,
			timeout:  config.timeout,
			metadata: newMetadata,
			enabled:  config.enabled,
		}
	}
}

// WithDisabled disables the configuration
func WithDisabled() ConfigOption {
	return func(config SimpleFunctionalConfig) SimpleFunctionalConfig {
		return SimpleFunctionalConfig{
			name:     config.name,
			timeout:  config.timeout,
			metadata: config.metadata,
			enabled:  false,
		}
	}
}

// SimpleResult represents a functional result using mo.Option for nullable values
type SimpleResult[T any] struct {
	value     mo.Option[T]
	error     mo.Option[error]
	duration  time.Duration
	metadata  map[string]interface{}
}

// NewSuccessSimpleResult creates a successful result
func NewSuccessSimpleResult[T any](value T, duration time.Duration, metadata map[string]interface{}) SimpleResult[T] {
	return SimpleResult[T]{
		value:    mo.Some(value),
		error:    mo.None[error](),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
	}
}

// NewErrorSimpleResult creates an error result
func NewErrorSimpleResult[T any](err error, duration time.Duration, metadata map[string]interface{}) SimpleResult[T] {
	return SimpleResult[T]{
		value:    mo.None[T](),
		error:    mo.Some(err),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
	}
}

// IsSuccess returns true if the result is successful
func (sr SimpleResult[T]) IsSuccess() bool {
	return sr.value.IsPresent() && sr.error.IsAbsent()
}

// IsError returns true if the result has an error
func (sr SimpleResult[T]) IsError() bool {
	return sr.error.IsPresent()
}

// GetValue returns the value if present
func (sr SimpleResult[T]) GetValue() mo.Option[T] {
	return sr.value
}

// GetError returns the error if present
func (sr SimpleResult[T]) GetError() mo.Option[error] {
	return sr.error
}

// MapValue applies a function to the value if present
func (sr SimpleResult[T]) MapValue(f func(T) T) SimpleResult[T] {
	if sr.IsSuccess() {
		newValue := sr.value.Map(func(value T) (T, bool) {
			return f(value), true
		})
		return SimpleResult[T]{
			value:    newValue,
			error:    sr.error,
			duration: sr.duration,
			metadata: sr.metadata,
		}
	}
	return sr
}

// FunctionalLambdaInvocation demonstrates functional Lambda operations
func FunctionalLambdaInvocation(functionName string, payload interface{}, config SimpleFunctionalConfig) SimpleResult[string] {
	startTime := time.Now()

	if !config.enabled {
		return NewErrorSimpleResult[string](
			fmt.Errorf("operation disabled"),
			time.Since(startTime),
			map[string]interface{}{"function_name": functionName},
		)
	}

	// Simulate Lambda invocation using functional composition
	result := fmt.Sprintf(`{"result": "success", "function": "%s", "payload_received": true}`, functionName)

	metadata := lo.Assign(
		config.metadata,
		map[string]interface{}{
			"function_name": functionName,
			"payload_type": fmt.Sprintf("%T", payload),
		},
	)

	return NewSuccessSimpleResult(
		result,
		time.Since(startTime),
		metadata,
	)
}

// FunctionalDynamoDBOperation demonstrates functional DynamoDB operations
func FunctionalDynamoDBOperation(tableName string, operation string, config SimpleFunctionalConfig) SimpleResult[map[string]interface{}] {
	startTime := time.Now()

	if !config.enabled {
		return NewErrorSimpleResult[map[string]interface{}](
			fmt.Errorf("operation disabled"),
			time.Since(startTime),
			map[string]interface{}{"table_name": tableName},
		)
	}

	// Simulate DynamoDB operation using functional composition
	result := map[string]interface{}{
		"table_name": tableName,
		"operation":  operation,
		"status":     "completed",
		"items":      []string{"item1", "item2", "item3"},
	}

	metadata := lo.Assign(
		config.metadata,
		map[string]interface{}{
			"table_name": tableName,
			"operation":  operation,
		},
	)

	return NewSuccessSimpleResult(
		result,
		time.Since(startTime),
		metadata,
	)
}

// FunctionalPipeline demonstrates functional pipeline composition
func FunctionalPipeline[T any](input T, operations ...func(T) T) T {
	return lo.Reduce(operations, func(acc T, op func(T) T, _ int) T {
		return op(acc)
	}, input)
}

// FunctionalFilter demonstrates functional filtering
func FunctionalFilter[T any](items []T, predicate func(T) bool) []T {
	return lo.Filter(items, func(item T, _ int) bool {
		return predicate(item)
	})
}

// FunctionalMap demonstrates functional mapping
func FunctionalMap[T, R any](items []T, mapper func(T, int) R) []R {
	return lo.Map(items, mapper)
}

// FunctionalReduce demonstrates functional reduction
func FunctionalReduce[T, R any](items []T, reducer func(R, T, int) R, initial R) R {
	return lo.Reduce(items, reducer, initial)
}

// FunctionalPartition demonstrates functional partitioning
func FunctionalPartition[T any](items []T, predicate func(T) bool) ([]T, []T) {
	// Implement partition using filter
	truthy := FunctionalFilter(items, predicate)
	falsy := FunctionalFilter(items, func(item T) bool { return !predicate(item) })
	return truthy, falsy
}

// FunctionalGroupBy demonstrates functional grouping
func FunctionalGroupBy[T any, K comparable](items []T, keyFunc func(T) K) map[K][]T {
	return lo.GroupBy(items, keyFunc)
}

// FunctionalUnique demonstrates functional deduplication
func FunctionalUnique[T comparable](items []T) []T {
	return lo.Uniq(items)
}

// FunctionalChunk demonstrates functional chunking
func FunctionalChunk[T any](items []T, size int) [][]T {
	return lo.Chunk(items, size)
}

// FunctionalFlatten demonstrates functional flattening
func FunctionalFlatten[T any](items [][]T) []T {
	return lo.Flatten(items)
}

// FunctionalFind demonstrates functional finding
func FunctionalFind[T any](items []T, predicate func(T) bool) mo.Option[T] {
	if item, found := lo.Find(items, predicate); found {
		return mo.Some(item)
	}
	return mo.None[T]()
}

// FunctionalEvery demonstrates functional every check
func FunctionalEvery[T any](items []T, predicate func(T) bool) bool {
	return lo.EveryBy(items, predicate)
}

// FunctionalSome demonstrates functional some check
func FunctionalSome[T any](items []T, predicate func(T) bool) bool {
	return lo.SomeBy(items, predicate)
}

// Demonstration of a more complex functional composition
func ComplexFunctionalOperation(data []string) SimpleResult[map[string]interface{}] {
	startTime := time.Now()

	// Functional pipeline: filter → map → group → reduce
	processed := FunctionalPipeline(data,
		func(items []string) []string {
			return FunctionalFilter(items, func(item string) bool {
				return len(item) > 3 // Filter strings longer than 3 characters
			})
		},
		func(items []string) []string {
			return FunctionalMap(items, func(item string, _ int) string {
				return "processed_" + item // Add prefix to each item
			})
		},
	)

	// Group by first character
	grouped := FunctionalGroupBy(processed, func(item string) rune {
		if len(item) > 0 {
			return rune(item[0])
		}
		return 0
	})

	// Count items in each group
	counts := make(map[string]int)
	for key, group := range grouped {
		counts[string(key)] = len(group)
	}

	result := map[string]interface{}{
		"original_count":  len(data),
		"processed_count": len(processed),
		"groups":          len(grouped),
		"group_counts":    counts,
		"unique_items":    len(FunctionalUnique(processed)),
	}

	return NewSuccessSimpleResult(
		result,
		time.Since(startTime),
		map[string]interface{}{
			"operation": "complex_functional_processing",
		},
	)
}

// FunctionalScenarioStep represents a functional scenario step
type FunctionalScenarioStep[T any] struct {
	name      string
	operation func() SimpleResult[T]
}

// NewFunctionalScenarioStep creates a new functional scenario step
func NewFunctionalScenarioStep[T any](name string, operation func() SimpleResult[T]) FunctionalScenarioStep[T] {
	return FunctionalScenarioStep[T]{
		name:      name,
		operation: operation,
	}
}

// Execute runs the scenario step
func (fss FunctionalScenarioStep[T]) Execute() SimpleResult[T] {
	return fss.operation()
}

// FunctionalScenario represents a functional scenario
type FunctionalScenario[T any] struct {
	name  string
	steps []FunctionalScenarioStep[T]
}

// NewFunctionalScenario creates a new functional scenario
func NewFunctionalScenario[T any](name string) FunctionalScenario[T] {
	return FunctionalScenario[T]{
		name:  name,
		steps: make([]FunctionalScenarioStep[T], 0),
	}
}

// AddStep adds a step to the scenario
func (fs FunctionalScenario[T]) AddStep(step FunctionalScenarioStep[T]) FunctionalScenario[T] {
	return FunctionalScenario[T]{
		name:  fs.name,
		steps: append(fs.steps, step),
	}
}

// Execute runs scenario steps functionally with early termination on error
func (fs FunctionalScenario[T]) Execute() SimpleResult[[]T] {
	startTime := time.Now()
	
	// Execute steps with early termination using functional reduce
	executeResult := lo.Reduce(fs.steps, func(acc struct {
		results []T
		shouldStop bool
	}, step FunctionalScenarioStep[T], _ int) struct {
		results []T
		shouldStop bool
	} {
		// If we already encountered an error, don't execute more steps
		if acc.shouldStop {
			return acc
		}
		
		// Execute the current step
		stepResult := step.Execute()
		
		// If step failed, stop execution
		if stepResult.IsError() {
			return struct {
				results []T
				shouldStop bool
			}{
				results: acc.results,
				shouldStop: true,
			}
		}
		
		// Add successful result
		if stepResult.GetValue().IsPresent() {
			return struct {
				results []T
				shouldStop bool
			}{
				results: append(acc.results, stepResult.GetValue().MustGet()),
				shouldStop: false,
			}
		}
		
		return acc
	}, struct {
		results []T
		shouldStop bool
	}{
		results: make([]T, 0),
		shouldStop: false,
	})
	
	// Check if execution was terminated due to error
	if executeResult.shouldStop {
		return NewErrorSimpleResult[[]T](
			fmt.Errorf("scenario '%s' failed during step execution", fs.name),
			time.Since(startTime),
			map[string]interface{}{
				"scenario": fs.name,
				"phase":    "execution",
			},
		)
	}

	return NewSuccessSimpleResult(
		executeResult.results,
		time.Since(startTime),
		map[string]interface{}{
			"scenario":    fs.name,
			"steps_count": len(executeResult.results),
		},
	)
}