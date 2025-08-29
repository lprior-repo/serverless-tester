// Package stepfunctions workflow chain - Advanced workflow orchestration for Step Functions
// Following strict Terratest patterns with Function/FunctionE variants

package stepfunctions

import (
	"fmt"
	"sync"
	"time"
)

// NewWorkflowChain creates a new workflow chain
func NewWorkflowChain() *WorkflowChain {
	return &WorkflowChain{
		steps:     make([]*WorkflowStep, 0),
		stepIndex: make(map[string]int),
	}
}

// AddStep adds a step to the workflow chain
func (wc *WorkflowChain) AddStep(step *WorkflowStep) error {
	if step.Name == "" {
		return fmt.Errorf("step name is required")
	}
	
	if _, exists := wc.stepIndex[step.Name]; exists {
		return fmt.Errorf("step with name %s already exists", step.Name)
	}
	
	// Set default dependency mode if not specified
	if step.DependencyMode == "" {
		step.DependencyMode = DependencyModeAll
	}
	
	wc.stepIndex[step.Name] = len(wc.steps)
	wc.steps = append(wc.steps, step)
	
	return nil
}

// GetSteps returns all steps in the chain
func (wc *WorkflowChain) GetSteps() []*WorkflowStep {
	return wc.steps
}

// ExecuteWorkflowChain executes a workflow chain
// This is a Terratest-style function that panics on error
func ExecuteWorkflowChain(ctx *TestContext, chain *WorkflowChain, config *WorkflowChainConfig) *WorkflowChainResult {
	result, err := ExecuteWorkflowChainE(ctx, chain, config)
	if err != nil {
		ctx.T.FailNow()
	}
	return result
}

// ExecuteWorkflowChainE executes a workflow chain with error return
func ExecuteWorkflowChainE(ctx *TestContext, chain *WorkflowChain, config *WorkflowChainConfig) (*WorkflowChainResult, error) {
	start := time.Now()
	
	// Validate inputs
	if chain == nil {
		return nil, fmt.Errorf("workflow chain is required")
	}
	
	if len(chain.steps) == 0 {
		return &WorkflowChainResult{
			StepResults: []WorkflowStepResult{},
			Stats: &WorkflowChainStats{
				TotalSteps:    0,
				ExecutedSteps: 0,
				FailedSteps:   0,
				SkippedSteps:  0,
				ExecutionTime: time.Since(start),
			},
			StartedAt:   start,
			CompletedAt: time.Now(),
		}, nil
	}
	
	if config == nil {
		config = defaultWorkflowChainConfig()
	} else {
		config = mergeWorkflowChainConfig(config)
	}
	
	// Validate workflow chain
	if err := validateWorkflowChain(chain); err != nil {
		return nil, fmt.Errorf("workflow chain validation failed: %w", err)
	}
	
	logOperation("WorkflowChainExecution", map[string]interface{}{
		"total_steps":        len(chain.steps),
		"timeout":           config.Timeout.String(),
		"max_parallel_steps": config.MaxParallelSteps,
		"fail_fast":         config.FailFast,
	})
	
	// Execute the workflow chain
	result, err := executeWorkflowChain(ctx, chain, config)
	if result != nil {
		result.StartedAt = start
		result.CompletedAt = time.Now()
		if result.Stats != nil {
			result.Stats.ExecutionTime = result.CompletedAt.Sub(result.StartedAt)
		}
	}
	
	logResult("WorkflowChainExecution", err == nil, time.Since(start), err)
	return result, err
}

// Internal implementation

// executeWorkflowChain executes the workflow chain with dependency management
func executeWorkflowChain(ctx *TestContext, chain *WorkflowChain, config *WorkflowChainConfig) (*WorkflowChainResult, error) {
	// Build dependency graph and execution order
	executionOrder, err := buildExecutionOrder(chain)
	if err != nil {
		return nil, err
	}
	
	// Initialize result tracking
	stepResults := make(map[string]*WorkflowStepResult)
	resultsMutex := sync.RWMutex{}
	
	stats := &WorkflowChainStats{
		TotalSteps: len(chain.steps),
	}
	
	// Execute steps in dependency order
	for _, levelSteps := range executionOrder {
		// Execute steps at this level in parallel (up to MaxParallelSteps)
		err := executeStepsLevel(ctx, levelSteps, config, stepResults, &resultsMutex)
		if err != nil && config.FailFast {
			break
		}
	}
	
	// Collect final results
	results := make([]WorkflowStepResult, 0, len(chain.steps))
	for _, step := range chain.steps {
		if result, exists := stepResults[step.Name]; exists {
			results = append(results, *result)
			if result.Skipped {
				stats.SkippedSteps++
			} else {
				// Count attempted steps as executed, regardless of success/failure
				stats.ExecutedSteps++
				if result.Error != nil {
					stats.FailedSteps++
				}
			}
		} else {
			// Step was never executed (likely due to failed dependencies)
			results = append(results, WorkflowStepResult{
				Step:    step,
				Skipped: true,
			})
			stats.SkippedSteps++
		}
	}
	
	return &WorkflowChainResult{
		StepResults: results,
		Stats:       stats,
	}, nil
}

// executeStepsLevel executes all steps at a given dependency level
func executeStepsLevel(ctx *TestContext, steps []*WorkflowStep, config *WorkflowChainConfig, stepResults map[string]*WorkflowStepResult, mutex *sync.RWMutex) error {
	if len(steps) == 0 {
		return nil
	}
	
	// Execute steps in parallel with concurrency control
	semaphore := make(chan struct{}, config.MaxParallelSteps)
	var wg sync.WaitGroup
	errorChan := make(chan error, len(steps))
	
	for _, step := range steps {
		wg.Add(1)
		go func(s *WorkflowStep) {
			defer wg.Done()
			
			// Acquire concurrency slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Execute step
			result := executeWorkflowStep(ctx, s, stepResults, mutex, config)
			
			// Store result
			mutex.Lock()
			stepResults[s.Name] = result
			mutex.Unlock()
			
			// Report errors
			if result.Error != nil {
				errorChan <- result.Error
			}
		}(step)
	}
	
	wg.Wait()
	close(errorChan)
	
	// Check for errors
	var firstError error
	for err := range errorChan {
		if firstError == nil {
			firstError = err
		}
	}
	
	return firstError
}

// executeWorkflowStep executes a single workflow step with retry logic
func executeWorkflowStep(ctx *TestContext, step *WorkflowStep, stepResults map[string]*WorkflowStepResult, mutex *sync.RWMutex, config *WorkflowChainConfig) *WorkflowStepResult {
	result := &WorkflowStepResult{
		Step:      step,
		StartedAt: time.Now(),
	}
	
	// Check if dependencies are satisfied
	if !areDependenciesSatisfied(step, stepResults, mutex) {
		result.Skipped = true
		result.CompletedAt = time.Now()
		return result
	}
	
	// Evaluate condition if present
	if step.Condition != nil && !evaluateCondition(step.Condition, stepResults, mutex) {
		result.Skipped = true
		result.CompletedAt = time.Now()
		return result
	}
	
	// Execute step with retry logic
	var lastError error
	maxAttempts := 1
	if step.RetryConfig != nil {
		maxAttempts = step.RetryConfig.MaxAttempts
	}
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.AttemptCount = attempt
		
		// Process input with template substitution
		processedInput, err := processStepInput(step.Input, stepResults, mutex)
		if err != nil {
			lastError = fmt.Errorf("input processing failed: %w", err)
			continue
		}
		
		// Execute the Step Function
		execResult, err := StartExecutionE(ctx, step.StateMachineArn, 
			fmt.Sprintf("%s-attempt-%d", step.Name, attempt), processedInput)
		
		if err == nil {
			result.Result = execResult
			result.CompletedAt = time.Now()
			return result
		}
		
		lastError = err
		
		// Wait before retry (if not last attempt)
		if attempt < maxAttempts && step.RetryConfig != nil {
			interval := calculateRetryInterval(step.RetryConfig, attempt)
			time.Sleep(interval)
		}
	}
	
	result.Error = lastError
	result.CompletedAt = time.Now()
	return result
}

// Dependency and validation functions

// buildExecutionOrder creates a topologically sorted execution order
func buildExecutionOrder(chain *WorkflowChain) ([][]*WorkflowStep, error) {
	// Simple implementation: group steps by dependency depth
	levels := make([][]*WorkflowStep, 0)
	processed := make(map[string]bool)
	
	for len(processed) < len(chain.steps) {
		currentLevel := make([]*WorkflowStep, 0)
		
		for _, step := range chain.steps {
			if processed[step.Name] {
				continue
			}
			
			// Check if all dependencies are processed
			canExecute := true
			for _, dep := range step.Dependencies {
				if !processed[dep] {
					canExecute = false
					break
				}
			}
			
			if canExecute {
				currentLevel = append(currentLevel, step)
			}
		}
		
		if len(currentLevel) == 0 {
			// No progress made - likely circular dependency
			return nil, fmt.Errorf("circular dependency detected in workflow chain")
		}
		
		// Mark steps in current level as processed
		for _, step := range currentLevel {
			processed[step.Name] = true
		}
		
		levels = append(levels, currentLevel)
	}
	
	return levels, nil
}

// areDependenciesSatisfied checks if step dependencies are satisfied
func areDependenciesSatisfied(step *WorkflowStep, stepResults map[string]*WorkflowStepResult, mutex *sync.RWMutex) bool {
	if len(step.Dependencies) == 0 {
		return true
	}
	
	mutex.RLock()
	defer mutex.RUnlock()
	
	satisfiedCount := 0
	for _, dep := range step.Dependencies {
		if result, exists := stepResults[dep]; exists && !result.Skipped && result.Error == nil {
			satisfiedCount++
		}
	}
	
	switch step.DependencyMode {
	case DependencyModeAll:
		return satisfiedCount == len(step.Dependencies)
	case DependencyModeAny:
		return satisfiedCount > 0
	default:
		return satisfiedCount == len(step.Dependencies)
	}
}

// evaluateCondition evaluates a workflow condition (simplified implementation)
func evaluateCondition(condition *WorkflowCondition, stepResults map[string]*WorkflowStepResult, mutex *sync.RWMutex) bool {
	// For simplicity, just return true (real implementation would parse and evaluate the expression)
	// In a production system, this would use a template engine like Go templates or a simple expression evaluator
	return true
}

// processStepInput processes step input with template substitution (simplified implementation)
func processStepInput(input *Input, stepResults map[string]*WorkflowStepResult, mutex *sync.RWMutex) (*Input, error) {
	if input == nil {
		return NewInput(), nil
	}
	
	// For simplicity, return the input as-is
	// Real implementation would perform template substitution using step results
	return input, nil
}

// calculateRetryInterval calculates the retry interval with exponential backoff
func calculateRetryInterval(config *WorkflowRetryConfig, attempt int) time.Duration {
	if config.BackoffMultiplier <= 1.0 {
		return config.InitialInterval
	}
	
	interval := time.Duration(float64(config.InitialInterval) * 
		(config.BackoffMultiplier * float64(attempt-1)))
	
	if config.MaxInterval > 0 && interval > config.MaxInterval {
		interval = config.MaxInterval
	}
	
	return interval
}

// validateWorkflowChain validates the workflow chain for common issues
func validateWorkflowChain(chain *WorkflowChain) error {
	if len(chain.steps) == 0 {
		return nil
	}
	
	// Check for duplicate step names
	stepNames := make(map[string]bool)
	for _, step := range chain.steps {
		if step.Name == "" {
			return fmt.Errorf("step name is required")
		}
		
		if stepNames[step.Name] {
			return fmt.Errorf("duplicate step name: %s", step.Name)
		}
		stepNames[step.Name] = true
		
		// Validate dependencies exist
		for _, dep := range step.Dependencies {
			if !stepNames[dep] && !contains(chain.steps, dep) {
				return fmt.Errorf("step %s depends on non-existent step: %s", step.Name, dep)
			}
		}
	}
	
	// Check for circular dependencies using simple cycle detection
	_, err := buildExecutionOrder(chain)
	return err
}

// Utility functions

// contains checks if a step with the given name exists in the chain
func contains(steps []*WorkflowStep, name string) bool {
	for _, step := range steps {
		if step.Name == name {
			return true
		}
	}
	return false
}

// Configuration functions

// defaultWorkflowChainConfig provides sensible defaults
func defaultWorkflowChainConfig() *WorkflowChainConfig {
	return &WorkflowChainConfig{
		Timeout:          DefaultTimeout,
		MaxParallelSteps: 3,
		FailFast:         true,
	}
}

// mergeWorkflowChainConfig merges user config with defaults
func mergeWorkflowChainConfig(userConfig *WorkflowChainConfig) *WorkflowChainConfig {
	defaults := defaultWorkflowChainConfig()
	
	if userConfig == nil {
		return defaults
	}
	
	merged := *userConfig
	
	if merged.Timeout == 0 {
		merged.Timeout = defaults.Timeout
	}
	if merged.MaxParallelSteps == 0 {
		merged.MaxParallelSteps = defaults.MaxParallelSteps
	}
	
	return &merged
}