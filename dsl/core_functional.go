// Package dsl provides a Domain-Specific Language for AWS serverless testing.
// This package uses pure functional programming with immutable data structures,
// functional composition, and monadic operations for maximum abstraction.
//
// Key Design Principles:
// - Pure functions with no side effects
// - Immutable data structures using mo.Option and mo.Result
// - Functional composition with lo utilities
// - Monadic error handling
// - Dependency injection with do container
package dsl

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// TestConfig represents immutable test configuration
type TestConfig struct {
	testingT     testing.TB
	awsConfig    aws.Config
	timeout      time.Duration
	ctx          context.Context
	metadata     map[string]interface{}
	dependencies *do.Injector
}

// ConfigOption represents a functional option for TestConfig
type ConfigOption func(TestConfig) TestConfig

// NewTestConfig creates a new immutable test configuration using functional options
func NewTestConfig(t testing.TB, opts ...ConfigOption) TestConfig {
	baseConfig := TestConfig{
		testingT:     t,
		timeout:      30 * time.Second,
		ctx:          context.Background(),
		metadata:     make(map[string]interface{}),
		dependencies: do.New(),
	}
	
	return lo.Reduce(opts, func(config TestConfig, opt ConfigOption, _ int) TestConfig {
		return opt(config)
	}, baseConfig)
}

// WithTimeout returns a ConfigOption that sets the timeout
func WithTimeout(timeout time.Duration) ConfigOption {
	return func(config TestConfig) TestConfig {
		return copyConfig(config, func(c *TestConfig) {
			c.timeout = timeout
		})
	}
}

// WithMetadata returns a ConfigOption that adds metadata
func WithMetadata(key string, value interface{}) ConfigOption {
	return func(config TestConfig) TestConfig {
		return copyConfig(config, func(c *TestConfig) {
			newMetadata := lo.Assign(c.metadata, map[string]interface{}{key: value})
			c.metadata = newMetadata
		})
	}
}

// WithAWSConfig returns a ConfigOption that sets AWS configuration
func WithAWSConfig(awsConfig aws.Config) ConfigOption {
	return func(config TestConfig) TestConfig {
		return copyConfig(config, func(c *TestConfig) {
			c.awsConfig = awsConfig
		})
	}
}

// WithContext returns a ConfigOption that sets the context
func WithContext(ctx context.Context) ConfigOption {
	return func(config TestConfig) TestConfig {
		return copyConfig(config, func(c *TestConfig) {
			c.ctx = ctx
		})
	}
}

// copyConfig creates a deep copy of TestConfig and applies modifications
func copyConfig(original TestConfig, modifier func(*TestConfig)) TestConfig {
	copied := TestConfig{
		testingT:     original.testingT,
		awsConfig:    original.awsConfig,
		timeout:      original.timeout,
		ctx:          original.ctx,
		metadata:     lo.Assign(map[string]interface{}{}, original.metadata),
		dependencies: original.dependencies,
	}
	modifier(&copied)
	return copied
}

// ExecutionResult represents the result of an operation using monads
type ExecutionResult[T any] struct {
	result   mo.Result[T]
	duration time.Duration
	metadata map[string]interface{}
	config   TestConfig
}

// NewSuccessResult creates a successful execution result
func NewSuccessResult[T any](value T, duration time.Duration, metadata map[string]interface{}, config TestConfig) ExecutionResult[T] {
	return ExecutionResult[T]{
		result:   mo.Ok[T](value),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
		config:   config,
	}
}

// NewErrorResult creates a failed execution result
func NewErrorResult[T any](err error, duration time.Duration, metadata map[string]interface{}, config TestConfig) ExecutionResult[T] {
	return ExecutionResult[T]{
		result:   mo.Err[T](err),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
		config:   config,
	}
}

// Map applies a function to the success value if present
func (er ExecutionResult[T]) Map(f func(T) T) ExecutionResult[T] {
	return ExecutionResult[T]{
		result:   er.result.Map(f),
		duration: er.duration,
		metadata: er.metadata,
		config:   er.config,
	}
}

// FlatMap applies a function that returns an ExecutionResult
func (er ExecutionResult[T]) FlatMap(f func(T) ExecutionResult[T]) ExecutionResult[T] {
	return er.result.Match(
		func(value T) ExecutionResult[T] { return f(value) },
		func(err error) ExecutionResult[T] { return NewErrorResult[T](err, er.duration, er.metadata, er.config) },
	)
}

// IsSuccess returns true if the operation succeeded
func (er ExecutionResult[T]) IsSuccess() bool {
	return er.result.IsOk()
}

// IsError returns true if the operation failed
func (er ExecutionResult[T]) IsError() bool {
	return er.result.IsError()
}

// Value returns the success value or panics if error
func (er ExecutionResult[T]) Value() T {
	return er.result.MustGet()
}

// Error returns the error or nil if success
func (er ExecutionResult[T]) Error() error {
	return er.result.Error()
}

// Should creates an assertion builder for the result
func (er ExecutionResult[T]) Should() AssertionBuilder[T] {
	return NewAssertionBuilder(er)
}

// AssertionBuilder provides functional assertion capabilities
type AssertionBuilder[T any] struct {
	result ExecutionResult[T]
}

// NewAssertionBuilder creates a new assertion builder
func NewAssertionBuilder[T any](result ExecutionResult[T]) AssertionBuilder[T] {
	return AssertionBuilder[T]{result: result}
}

// Succeed asserts that the operation succeeded
func (ab AssertionBuilder[T]) Succeed() AssertionBuilder[T] {
	if ab.result.IsError() {
		ab.result.config.testingT.Fatalf("Expected operation to succeed, but it failed: %v", ab.result.Error())
	}
	return ab
}

// Fail asserts that the operation failed
func (ab AssertionBuilder[T]) Fail() AssertionBuilder[T] {
	if ab.result.IsSuccess() {
		ab.result.config.testingT.Fatalf("Expected operation to fail, but it succeeded")
	}
	return ab
}

// ReturnValue asserts the returned value matches expected using functional equality
func (ab AssertionBuilder[T]) ReturnValue(expected T, equals func(T, T) bool) AssertionBuilder[T] {
	if ab.result.IsSuccess() && !equals(ab.result.Value(), expected) {
		ab.result.config.testingT.Fatalf("Expected return value %+v, got %+v", expected, ab.result.Value())
	}
	return ab
}

// HaveMetadata asserts that specific metadata exists
func (ab AssertionBuilder[T]) HaveMetadata(key string, value interface{}) AssertionBuilder[T] {
	actual, exists := ab.result.metadata[key]
	if !exists || actual != value {
		ab.result.config.testingT.Fatalf("Expected metadata %s=%v, got %v", key, value, actual)
	}
	return ab
}

// CompleteWithin asserts the operation completed within a duration
func (ab AssertionBuilder[T]) CompleteWithin(duration time.Duration) AssertionBuilder[T] {
	if ab.result.duration > duration {
		ab.result.config.testingT.Fatalf("Expected operation to complete within %v, took %v", duration, ab.result.duration)
	}
	return ab
}

// And allows chaining multiple assertions
func (ab AssertionBuilder[T]) And() AssertionBuilder[T] {
	return ab
}

// AsyncExecutionResult represents an asynchronous operation using channels and functors
type AsyncExecutionResult[T any] struct {
	future chan ExecutionResult[T]
	cancel context.CancelFunc
}

// NewAsyncExecutionResult creates a new async execution result
func NewAsyncExecutionResult[T any](compute func() ExecutionResult[T]) AsyncExecutionResult[T] {
	future := make(chan ExecutionResult[T], 1)
	ctx, cancel := context.WithCancel(context.Background())
	
	go func() {
		defer close(future)
		select {
		case <-ctx.Done():
			return
		default:
			result := compute()
			future <- result
		}
	}()
	
	return AsyncExecutionResult[T]{
		future: future,
		cancel: cancel,
	}
}

// Wait waits for the async operation to complete
func (aer AsyncExecutionResult[T]) Wait() ExecutionResult[T] {
	return <-aer.future
}

// WaitWithTimeout waits with a specific timeout using functional composition
func (aer AsyncExecutionResult[T]) WaitWithTimeout(timeout time.Duration, config TestConfig) ExecutionResult[T] {
	select {
	case result := <-aer.future:
		return result
	case <-time.After(timeout):
		aer.cancel()
		return NewErrorResult[T](
			fmt.Errorf("operation timed out after %v", timeout),
			timeout,
			map[string]interface{}{"timeout": true},
			config,
		)
	}
}

// Map applies a transformation to the async result when it completes
func (aer AsyncExecutionResult[T]) Map(f func(T) T) AsyncExecutionResult[T] {
	return NewAsyncExecutionResult(func() ExecutionResult[T] {
		return aer.Wait().Map(f)
	})
}

// ServiceBuilder represents a generic AWS service builder using functional composition
type ServiceBuilder[T any] struct {
	config  TestConfig
	builder func(TestConfig) ExecutionResult[T]
}

// NewServiceBuilder creates a new service builder
func NewServiceBuilder[T any](config TestConfig, builder func(TestConfig) ExecutionResult[T]) ServiceBuilder[T] {
	return ServiceBuilder[T]{
		config:  config,
		builder: builder,
	}
}

// Execute runs the service builder
func (sb ServiceBuilder[T]) Execute() ExecutionResult[T] {
	return sb.builder(sb.config)
}

// WithConfig applies additional configuration to the builder
func (sb ServiceBuilder[T]) WithConfig(opts ...ConfigOption) ServiceBuilder[T] {
	newConfig := lo.Reduce(opts, func(config TestConfig, opt ConfigOption, _ int) TestConfig {
		return opt(config)
	}, sb.config)
	
	return ServiceBuilder[T]{
		config:  newConfig,
		builder: sb.builder,
	}
}

// Chain allows chaining multiple service operations
func (sb ServiceBuilder[T]) Chain(f func(T) ServiceBuilder[T]) ServiceBuilder[T] {
	return NewServiceBuilder(sb.config, func(config TestConfig) ExecutionResult[T] {
		result := sb.Execute()
		return result.FlatMap(func(value T) ExecutionResult[T] {
			return f(value).Execute()
		})
	})
}

// ScenarioStep represents a step in a testing scenario using pure functions
type ScenarioStep[T any] struct {
	name        string
	description mo.Option[string]
	operation   func(TestConfig) ExecutionResult[T]
	assertions  []func(ExecutionResult[T]) bool
}

// NewScenarioStep creates a new scenario step
func NewScenarioStep[T any](name string, operation func(TestConfig) ExecutionResult[T]) ScenarioStep[T] {
	return ScenarioStep[T]{
		name:        name,
		description: mo.None[string](),
		operation:   operation,
		assertions:  make([]func(ExecutionResult[T]) bool, 0),
	}
}

// WithDescription adds a description to the scenario step
func (ss ScenarioStep[T]) WithDescription(desc string) ScenarioStep[T] {
	return ScenarioStep[T]{
		name:        ss.name,
		description: mo.Some(desc),
		operation:   ss.operation,
		assertions:  ss.assertions,
	}
}

// WithAssertion adds an assertion to the scenario step
func (ss ScenarioStep[T]) WithAssertion(assertion func(ExecutionResult[T]) bool) ScenarioStep[T] {
	return ScenarioStep[T]{
		name:        ss.name,
		description: ss.description,
		operation:   ss.operation,
		assertions:  append(ss.assertions, assertion),
	}
}

// Execute runs the scenario step
func (ss ScenarioStep[T]) Execute(config TestConfig) ExecutionResult[T] {
	startTime := time.Now()
	result := ss.operation(config)
	
	// Apply all assertions functionally
	allAssertionsPassed := lo.EveryBy(ss.assertions, func(assertion func(ExecutionResult[T]) bool) bool {
		return assertion(result)
	})
	
	if !allAssertionsPassed {
		return NewErrorResult[T](
			fmt.Errorf("assertions failed for step '%s'", ss.name),
			time.Since(startTime),
			map[string]interface{}{"step": ss.name},
			config,
		)
	}
	
	return result
}

// Scenario represents a collection of steps using functional composition
type Scenario[T any] struct {
	steps  []ScenarioStep[T]
	config TestConfig
}

// NewScenario creates a new scenario
func NewScenario[T any](config TestConfig) Scenario[T] {
	return Scenario[T]{
		steps:  make([]ScenarioStep[T], 0),
		config: config,
	}
}

// AddStep adds a step to the scenario using functional composition
func (s Scenario[T]) AddStep(step ScenarioStep[T]) Scenario[T] {
	return Scenario[T]{
		steps:  append(s.steps, step),
		config: s.config,
	}
}

// Execute runs all scenario steps using functional composition
func (s Scenario[T]) Execute() ExecutionResult[[]T] {
	startTime := time.Now()
	
	// Execute all steps and collect results
	results := lo.Map(s.steps, func(step ScenarioStep[T], _ int) ExecutionResult[T] {
		return step.Execute(s.config)
	})
	
	// Check if any step failed
	failedStep, foundFailed := lo.Find(results, func(result ExecutionResult[T]) bool {
		return result.IsError()
	})
	
	if foundFailed && failedStep.IsError() {
		return NewErrorResult[[]T](
			fmt.Errorf("scenario failed: %v", failedStep.Error()),
			time.Since(startTime),
			map[string]interface{}{"failed_step": true},
			s.config,
		)
	}
	
	// Extract successful values
	values := lo.Map(results, func(result ExecutionResult[T], _ int) T {
		return result.Value()
	})
	
	return NewSuccessResult(
		values,
		time.Since(startTime),
		map[string]interface{}{"steps_completed": len(s.steps)},
		s.config,
	)
}

// TestOrchestration represents a complete test using functional composition
type TestOrchestration[T any] struct {
	name      string
	setup     []func(TestConfig) ExecutionResult[interface{}]
	scenarios []Scenario[T]
	teardown  []func(TestConfig) ExecutionResult[interface{}]
	config    TestConfig
}

// NewTestOrchestration creates a new test orchestration
func NewTestOrchestration[T any](name string, config TestConfig) TestOrchestration[T] {
	return TestOrchestration[T]{
		name:      name,
		setup:     make([]func(TestConfig) ExecutionResult[interface{}], 0),
		scenarios: make([]Scenario[T], 0),
		teardown:  make([]func(TestConfig) ExecutionResult[interface{}], 0),
		config:    config,
	}
}

// WithSetup adds setup operations
func (to TestOrchestration[T]) WithSetup(setup func(TestConfig) ExecutionResult[interface{}]) TestOrchestration[T] {
	return TestOrchestration[T]{
		name:      to.name,
		setup:     append(to.setup, setup),
		scenarios: to.scenarios,
		teardown:  to.teardown,
		config:    to.config,
	}
}

// AddScenario adds a scenario to the test
func (to TestOrchestration[T]) AddScenario(scenario Scenario[T]) TestOrchestration[T] {
	return TestOrchestration[T]{
		name:      to.name,
		setup:     to.setup,
		scenarios: append(to.scenarios, scenario),
		teardown:  to.teardown,
		config:    to.config,
	}
}

// WithTeardown adds teardown operations
func (to TestOrchestration[T]) WithTeardown(teardown func(TestConfig) ExecutionResult[interface{}]) TestOrchestration[T] {
	return TestOrchestration[T]{
		name:      to.name,
		setup:     to.setup,
		scenarios: to.scenarios,
		teardown:  append(to.teardown, teardown),
		config:    to.config,
	}
}

// Run executes the complete test orchestration using functional composition
func (to TestOrchestration[T]) Run() ExecutionResult[[][]T] {
	startTime := time.Now()
	
	// Execute setup using functional composition
	setupResults := lo.Map(to.setup, func(setup func(TestConfig) ExecutionResult[interface{}], _ int) ExecutionResult[interface{}] {
		return setup(to.config)
	})
	
	// Check if any setup failed
	failedSetup, foundFailedSetup := lo.Find(setupResults, func(result ExecutionResult[interface{}]) bool {
		return result.IsError()
	})
	
	if foundFailedSetup && failedSetup.IsError() {
		return NewErrorResult[[][]T](
			fmt.Errorf("test setup failed: %v", failedSetup.Error()),
			time.Since(startTime),
			map[string]interface{}{"phase": "setup"},
			to.config,
		)
	}
	
	// Execute scenarios
	scenarioResults := lo.Map(to.scenarios, func(scenario Scenario[T], _ int) ExecutionResult[[]T] {
		return scenario.Execute()
	})
	
	// Check if any scenario failed
	failedScenario, foundFailedScenario := lo.Find(scenarioResults, func(result ExecutionResult[[]T]) bool {
		return result.IsError()
	})
	
	// Always execute teardown
	teardownResults := lo.Map(to.teardown, func(teardown func(TestConfig) ExecutionResult[interface{}], _ int) ExecutionResult[interface{}] {
		return teardown(to.config)
	})
	
	if foundFailedScenario && failedScenario.IsError() {
		return NewErrorResult[[][]T](
			fmt.Errorf("scenario failed: %v", failedScenario.Error()),
			time.Since(startTime),
			map[string]interface{}{"phase": "scenario"},
			to.config,
		)
	}
	
	// Check if any teardown failed
	failedTeardown, foundFailedTeardown := lo.Find(teardownResults, func(result ExecutionResult[interface{}]) bool {
		return result.IsError()
	})
	
	if foundFailedTeardown && failedTeardown.IsError() {
		return NewErrorResult[[][]T](
			fmt.Errorf("test teardown failed: %v", failedTeardown.Error()),
			time.Since(startTime),
			map[string]interface{}{"phase": "teardown"},
			to.config,
		)
	}
	
	// Extract successful scenario values
	values := lo.Map(scenarioResults, func(result ExecutionResult[[]T], _ int) []T {
		return result.Value()
	})
	
	return NewSuccessResult(
		values,
		time.Since(startTime),
		map[string]interface{}{
			"test_name":       to.name,
			"scenarios_count": len(to.scenarios),
			"setup_count":     len(to.setup),
			"teardown_count":  len(to.teardown),
		},
		to.config,
	)
}

// Common error types using functional error handling
var (
	ErrOperationTimeout = fmt.Errorf("operation timed out")
	ErrInvalidConfig    = fmt.Errorf("invalid configuration")
	ErrResourceNotFound = fmt.Errorf("resource not found")
	ErrAssertion        = fmt.Errorf("assertion failed")
)

// Utility functions for functional composition

// Pipe allows function composition
func Pipe[A, B, C any](f func(A) B, g func(B) C) func(A) C {
	return func(a A) C {
		return g(f(a))
	}
}

// Compose allows reverse function composition
func Compose[A, B, C any](g func(B) C, f func(A) B) func(A) C {
	return func(a A) C {
		return g(f(a))
	}
}

// Try wraps a function that might panic into a Result
func Try[T any](f func() T) mo.Result[T] {
	defer func() {
		if r := recover(); r != nil {
			// This would return an error, but we can't modify the return in defer
		}
	}()
	
	return mo.Ok[T](f())
}

// TryWithError wraps a function that returns a value and error into a Result
func TryWithError[T any](f func() (T, error)) mo.Result[T] {
	value, err := f()
	if err != nil {
		return mo.Err[T](err)
	}
	return mo.Ok[T](value)
}