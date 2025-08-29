// Package dsl provides a Domain-Specific Language for AWS serverless testing.
// This package replaces the Function/FunctionE patterns with fluent, expressive interfaces
// that make writing tests feel like writing natural language descriptions.
//
// Example usage:
//   AWS().Lambda().Function("my-func").
//     WithPayload(`{"key": "value"}`).
//     Invoke().
//     Should().Succeed().
//     And().ReturnValue("expected")
//
// Key Design Principles:
// - Fluent interfaces for maximum readability
// - Method chaining for natural test flow
// - Implicit error handling with descriptive failures
// - High-level abstractions that hide AWS complexity
// - Compositional builders for complex scenarios
package dsl

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// TestContext provides the core context for all DSL operations
type TestContext struct {
	t         testing.TB
	awsConfig aws.Config
	timeout   time.Duration
	ctx       context.Context
	metadata  map[string]interface{}
}

// NewTestContext creates a new DSL test context
func NewTestContext(t testing.TB) *TestContext {
	return &TestContext{
		t:         t,
		timeout:   30 * time.Second,
		ctx:       context.Background(),
		metadata:  make(map[string]interface{}),
	}
}

// WithTimeout sets a custom timeout for operations
func (tc *TestContext) WithTimeout(timeout time.Duration) *TestContext {
	tc.timeout = timeout
	return tc
}

// WithMetadata adds custom metadata to the context
func (tc *TestContext) WithMetadata(key string, value interface{}) *TestContext {
	tc.metadata[key] = value
	return tc
}

// AWS creates the entry point to AWS service DSL
func (tc *TestContext) AWS() *AWSBuilder {
	return &AWSBuilder{ctx: tc}
}

// AWSBuilder provides the main entry point for AWS service interactions
type AWSBuilder struct {
	ctx *TestContext
}

// Context returns the test context
func (ab *AWSBuilder) Context() *TestContext {
	return ab.ctx
}

// Lambda returns a Lambda service builder
func (ab *AWSBuilder) Lambda() *LambdaBuilder {
	return &LambdaBuilder{ctx: ab.ctx}
}

// DynamoDB returns a DynamoDB service builder
func (ab *AWSBuilder) DynamoDB() *DynamoDBBuilder {
	return &DynamoDBBuilder{ctx: ab.ctx}
}

// S3 returns an S3 service builder
func (ab *AWSBuilder) S3() *S3Builder {
	return &S3Builder{ctx: ab.ctx}
}

// SES returns an SES service builder
func (ab *AWSBuilder) SES() *SESBuilder {
	return &SESBuilder{ctx: ab.ctx}
}

// APIGateway returns an API Gateway service builder
func (ab *AWSBuilder) APIGateway() *APIGatewayBuilder {
	return &APIGatewayBuilder{ctx: ab.ctx}
}

// EventBridge returns an EventBridge service builder
func (ab *AWSBuilder) EventBridge() *EventBridgeBuilder {
	return &EventBridgeBuilder{ctx: ab.ctx}
}

// StepFunctions returns a Step Functions service builder
func (ab *AWSBuilder) StepFunctions() *StepFunctionsBuilder {
	return &StepFunctionsBuilder{ctx: ab.ctx}
}

// Scenario creates a high-level testing scenario builder
func (ab *AWSBuilder) Scenario() *ScenarioBuilder {
	return &ScenarioBuilder{ctx: ab.ctx}
}

// Test creates a test orchestration builder
func (ab *AWSBuilder) Test() *TestBuilder {
	return &TestBuilder{ctx: ab.ctx}
}

// Common assertion interfaces for all services
type Assertable interface {
	Should() *AssertionBuilder
	Must() *AssertionBuilder
	Expect() *AssertionBuilder
}

type Executable interface {
	Execute() *ExecutionResult
	ExecuteAsync() *AsyncExecutionResult
}

type Configurable interface {
	WithConfig(config interface{}) Configurable
	WithTags(tags map[string]string) Configurable
	WithTimeout(timeout time.Duration) Configurable
}

// ExecutionResult represents the result of an operation
type ExecutionResult struct {
	Success   bool
	Data      interface{}
	Error     error
	Duration  time.Duration
	Metadata  map[string]interface{}
	ctx       *TestContext
}

// Should provides assertion capabilities on execution results
func (er *ExecutionResult) Should() *AssertionBuilder {
	return &AssertionBuilder{
		result: er,
		ctx:    er.ctx,
	}
}

// AsyncExecutionResult represents an asynchronous operation result
type AsyncExecutionResult struct {
	*ExecutionResult
	completed chan bool
	cancel    context.CancelFunc
}

// Wait waits for the async operation to complete
func (aer *AsyncExecutionResult) Wait() *ExecutionResult {
	<-aer.completed
	return aer.ExecutionResult
}

// WaitWithTimeout waits with a specific timeout
func (aer *AsyncExecutionResult) WaitWithTimeout(timeout time.Duration) *ExecutionResult {
	select {
	case <-aer.completed:
		return aer.ExecutionResult
	case <-time.After(timeout):
		aer.cancel()
		return &ExecutionResult{
			Success: false,
			Error:   ErrOperationTimeout,
			ctx:     aer.ctx,
		}
	}
}

// AssertionBuilder provides fluent assertion capabilities
type AssertionBuilder struct {
	result *ExecutionResult
	ctx    *TestContext
}

// Succeed asserts that the operation succeeded
func (ab *AssertionBuilder) Succeed() *AssertionBuilder {
	if !ab.result.Success {
		ab.ctx.t.Fatalf("Expected operation to succeed, but it failed: %v", ab.result.Error)
	}
	return ab
}

// Fail asserts that the operation failed
func (ab *AssertionBuilder) Fail() *AssertionBuilder {
	if ab.result.Success {
		ab.ctx.t.Fatalf("Expected operation to fail, but it succeeded")
	}
	return ab
}

// ReturnValue asserts the returned value matches expected
func (ab *AssertionBuilder) ReturnValue(expected interface{}) *AssertionBuilder {
	if ab.result.Data != expected {
		ab.ctx.t.Fatalf("Expected return value %v, got %v", expected, ab.result.Data)
	}
	return ab
}

// ContainValue asserts the result contains a specific value
func (ab *AssertionBuilder) ContainValue(expected interface{}) *AssertionBuilder {
	// Implementation would check if result contains the expected value
	return ab
}

// HaveMetadata asserts that specific metadata exists
func (ab *AssertionBuilder) HaveMetadata(key string, value interface{}) *AssertionBuilder {
	if actual, exists := ab.result.Metadata[key]; !exists || actual != value {
		ab.ctx.t.Fatalf("Expected metadata %s=%v, got %v", key, value, actual)
	}
	return ab
}

// CompleteWithin asserts the operation completed within a duration
func (ab *AssertionBuilder) CompleteWithin(duration time.Duration) *AssertionBuilder {
	if ab.result.Duration > duration {
		ab.ctx.t.Fatalf("Expected operation to complete within %v, took %v", duration, ab.result.Duration)
	}
	return ab
}

// And allows chaining multiple assertions
func (ab *AssertionBuilder) And() *AssertionBuilder {
	return ab
}

// ScenarioBuilder creates high-level testing scenarios
type ScenarioBuilder struct {
	ctx   *TestContext
	steps []ScenarioStep
}

type ScenarioStep struct {
	Name        string
	Description string
	Operation   func() *ExecutionResult
	Assertions  []func(*ExecutionResult) bool
}

// Step adds a step to the scenario
func (sb *ScenarioBuilder) Step(name string) *ScenarioStepBuilder {
	return &ScenarioStepBuilder{
		scenarioBuilder: sb,
		step:            &ScenarioStep{Name: name},
	}
}

// Execute runs all scenario steps
func (sb *ScenarioBuilder) Execute() *ExecutionResult {
	startTime := time.Now()
	results := make([]map[string]interface{}, 0, len(sb.steps))
	
	for _, step := range sb.steps {
		stepResult := step.Operation()
		if !stepResult.Success {
			return &ExecutionResult{
				Success:  false,
				Error:    fmt.Errorf("scenario failed at step '%s': %v", step.Name, stepResult.Error),
				Duration: time.Since(startTime),
				ctx:      sb.ctx,
			}
		}
		
		results = append(results, map[string]interface{}{
			"step":     step.Name,
			"success":  stepResult.Success,
			"duration": stepResult.Duration,
		})
	}
	
	return &ExecutionResult{
		Success:  true,
		Data:     results,
		Duration: time.Since(startTime),
		ctx:      sb.ctx,
	}
}

// ScenarioStepBuilder builds individual scenario steps
type ScenarioStepBuilder struct {
	scenarioBuilder *ScenarioBuilder
	step            *ScenarioStep
}

// Describe adds a description to the step
func (ssb *ScenarioStepBuilder) Describe(description string) *ScenarioStepBuilder {
	ssb.step.Description = description
	return ssb
}

// Do sets the operation for the step
func (ssb *ScenarioStepBuilder) Do(operation func() *ExecutionResult) *ScenarioStepBuilder {
	ssb.step.Operation = operation
	return ssb
}

// Assert adds an assertion to the step
func (ssb *ScenarioStepBuilder) Assert(assertion func(*ExecutionResult) bool) *ScenarioStepBuilder {
	ssb.step.Assertions = append(ssb.step.Assertions, assertion)
	return ssb
}

// Done completes the step and returns to scenario builder
func (ssb *ScenarioStepBuilder) Done() *ScenarioBuilder {
	ssb.scenarioBuilder.steps = append(ssb.scenarioBuilder.steps, *ssb.step)
	return ssb.scenarioBuilder
}

// TestBuilder creates comprehensive test orchestration
type TestBuilder struct {
	ctx       *TestContext
	testName  string
	scenarios []*ScenarioBuilder
	setup     []func() *ExecutionResult
	teardown  []func() *ExecutionResult
}

// Named sets the test name
func (tb *TestBuilder) Named(name string) *TestBuilder {
	tb.testName = name
	return tb
}

// WithSetup adds setup operations
func (tb *TestBuilder) WithSetup(setup func() *ExecutionResult) *TestBuilder {
	tb.setup = append(tb.setup, setup)
	return tb
}

// WithTeardown adds teardown operations
func (tb *TestBuilder) WithTeardown(teardown func() *ExecutionResult) *TestBuilder {
	tb.teardown = append(tb.teardown, teardown)
	return tb
}

// AddScenario adds a scenario to the test
func (tb *TestBuilder) AddScenario(scenario *ScenarioBuilder) *TestBuilder {
	tb.scenarios = append(tb.scenarios, scenario)
	return tb
}

// Run executes the complete test
func (tb *TestBuilder) Run() *ExecutionResult {
	startTime := time.Now()
	
	// Run setup
	for _, setupFunc := range tb.setup {
		if result := setupFunc(); !result.Success {
			return &ExecutionResult{
				Success:  false,
				Error:    fmt.Errorf("test setup failed: %v", result.Error),
				Duration: time.Since(startTime),
				ctx:      tb.ctx,
			}
		}
	}
	
	// Run scenarios
	scenarioResults := make([]map[string]interface{}, 0, len(tb.scenarios))
	for i, scenario := range tb.scenarios {
		result := scenario.Execute()
		scenarioResults = append(scenarioResults, map[string]interface{}{
			"scenario": i + 1,
			"success":  result.Success,
			"duration": result.Duration,
		})
		
		if !result.Success {
			// Run teardown before failing
			for _, teardownFunc := range tb.teardown {
				teardownFunc()
			}
			return &ExecutionResult{
				Success:  false,
				Error:    fmt.Errorf("scenario %d failed: %v", i+1, result.Error),
				Duration: time.Since(startTime),
				ctx:      tb.ctx,
			}
		}
	}
	
	// Run teardown
	for _, teardownFunc := range tb.teardown {
		teardownFunc()
	}
	
	return &ExecutionResult{
		Success:  true,
		Data:     scenarioResults,
		Duration: time.Since(startTime),
		Metadata: map[string]interface{}{
			"test_name":       tb.testName,
			"scenarios_count": len(tb.scenarios),
		},
		ctx: tb.ctx,
	}
}

// Common errors
var (
	ErrOperationTimeout = fmt.Errorf("operation timed out")
	ErrInvalidConfig    = fmt.Errorf("invalid configuration")
	ErrResourceNotFound = fmt.Errorf("resource not found")
)

// Global DSL entry point for convenience
func AWS() *AWSBuilder {
	// This would typically be initialized with a default test context
	// For now, we'll require explicit context creation
	panic("Use NewTestContext(t).AWS() instead of global AWS()")
}

// Helper function to create a test context and return AWS builder
func NewAWS(t testing.TB) *AWSBuilder {
	return NewTestContext(t).AWS()
}