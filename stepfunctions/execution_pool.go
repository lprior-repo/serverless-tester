// Package stepfunctions execution pool - Advanced execution pool management for Step Functions
// Following strict Terratest patterns with Function/FunctionE variants

package stepfunctions

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ExecutionPool manages a pool of Step Functions executions with limits and scheduling
type ExecutionPool struct {
	ctx          *TestContext
	config       *ExecutionPoolConfig
	
	// Internal state
	mu           sync.RWMutex
	executions   map[string]*PoolExecutionResult
	queue        []*PoolExecutionResult
	running      map[string]*PoolExecutionResult
	
	// Channels for coordination
	submitChan   chan *PoolExecutionResult
	cancelChan   chan string
	shutdownChan chan struct{}
	doneChan     chan struct{}
	
	// Worker coordination
	workerWg     sync.WaitGroup
	shutdown     bool
	
	// Statistics
	stats        ExecutionPoolStats
}

// NewExecutionPool creates a new execution pool with the specified configuration
func NewExecutionPool(ctx *TestContext, config *ExecutionPoolConfig) *ExecutionPool {
	if config == nil {
		config = defaultExecutionPoolConfig()
	} else {
		config = mergeExecutionPoolConfig(config)
	}
	
	pool := &ExecutionPool{
		ctx:          ctx,
		config:       config,
		executions:   make(map[string]*PoolExecutionResult),
		queue:        make([]*PoolExecutionResult, 0),
		running:      make(map[string]*PoolExecutionResult),
		submitChan:   make(chan *PoolExecutionResult, config.MaxQueueSize),
		cancelChan:   make(chan string, 100),
		shutdownChan: make(chan struct{}),
		doneChan:     make(chan struct{}),
		stats: ExecutionPoolStats{
			MaxConcurrentExecutions: config.MaxConcurrentExecutions,
			MaxQueueSize:          config.MaxQueueSize,
		},
	}
	
	// Start the pool manager goroutine
	go pool.managePool()
	
	logOperation("ExecutionPoolCreated", map[string]interface{}{
		"max_concurrent": config.MaxConcurrentExecutions,
		"max_queue_size": config.MaxQueueSize,
		"timeout":       config.ExecutionTimeout.String(),
	})
	
	return pool
}

// Public Interface

// Submit submits an execution request to the pool
// This is a Terratest-style function that panics on error
func (p *ExecutionPool) Submit(request ExecutionRequest) *PoolExecutionResult {
	result, err := p.SubmitE(request)
	if err != nil {
		p.ctx.T.FailNow()
	}
	return result
}

// SubmitE submits an execution request to the pool with error return
func (p *ExecutionPool) SubmitE(request ExecutionRequest) (*PoolExecutionResult, error) {
	// Validate request
	if err := p.validateExecutionRequest(request); err != nil {
		return nil, err
	}
	
	p.mu.Lock()
	if p.shutdown {
		p.mu.Unlock()
		return nil, fmt.Errorf("execution pool is shutdown")
	}
	
	// Check if we can accept more executions
	totalInPool := len(p.queue) + len(p.running)
	if totalInPool >= p.config.MaxConcurrentExecutions+p.config.MaxQueueSize {
		p.mu.Unlock()
		return nil, fmt.Errorf("execution pool is at capacity")
	}
	
	// Create execution result
	result := &PoolExecutionResult{
		ExecutionID: generateExecutionID(),
		Status:      PoolExecutionStatusQueued,
		SubmittedAt: time.Now(),
	}
	
	p.executions[result.ExecutionID] = result
	p.stats.TotalSubmitted++
	p.mu.Unlock()
	
	// Store request for later processing
	result.request = request
	
	// Submit to manager
	select {
	case p.submitChan <- result:
		return result, nil
	default:
		// Channel full, remove from tracking
		p.mu.Lock()
		delete(p.executions, result.ExecutionID)
		p.stats.TotalSubmitted--
		p.mu.Unlock()
		return nil, fmt.Errorf("execution pool submit queue is full")
	}
}

// GetExecution returns the current state of an execution
func (p *ExecutionPool) GetExecution(executionID string) *PoolExecutionResult {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if result, exists := p.executions[executionID]; exists {
		// Return a copy to prevent race conditions
		resultCopy := *result
		return &resultCopy
	}
	return nil
}

// GetStats returns current pool statistics
func (p *ExecutionPool) GetStats() ExecutionPoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	// Calculate current stats
	stats := p.stats
	stats.RunningExecutions = len(p.running)
	stats.QueuedExecutions = len(p.queue)
	
	// Calculate average execution time
	if stats.CompletedExecutions > 0 {
		totalDuration := time.Duration(0)
		count := 0
		
		for _, result := range p.executions {
			if result.Status == PoolExecutionStatusCompleted || result.Status == PoolExecutionStatusFailed {
				if !result.StartedAt.IsZero() && !result.CompletedAt.IsZero() {
					totalDuration += result.CompletedAt.Sub(result.StartedAt)
					count++
				}
			}
		}
		
		if count > 0 {
			stats.AverageExecutionTime = totalDuration / time.Duration(count)
		}
	}
	
	return stats
}

// CancelExecution cancels a queued or running execution
func (p *ExecutionPool) CancelExecution(executionID string) {
	err := p.CancelExecutionE(executionID)
	if err != nil {
		p.ctx.T.FailNow()
	}
}

// CancelExecutionE cancels a queued or running execution with error return
func (p *ExecutionPool) CancelExecutionE(executionID string) error {
	p.mu.RLock()
	result, exists := p.executions[executionID]
	p.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}
	
	if result.Status == PoolExecutionStatusCompleted || result.Status == PoolExecutionStatusFailed || result.Status == PoolExecutionStatusCancelled {
		return fmt.Errorf("execution %s cannot be cancelled (status: %s)", executionID, result.Status)
	}
	
	// Send cancel request to manager
	select {
	case p.cancelChan <- executionID:
		return nil
	default:
		return fmt.Errorf("cancel request queue is full")
	}
}

// WaitForAll waits for all submitted executions to complete
func (p *ExecutionPool) WaitForAll() {
	err := p.WaitForAllE()
	if err != nil {
		p.ctx.T.FailNow()
	}
}

// WaitForAllE waits for all submitted executions to complete with error return
func (p *ExecutionPool) WaitForAllE() error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	timeout := time.After(p.config.PoolTimeout)
	
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for executions to complete")
		case <-ticker.C:
			p.mu.RLock()
			running := len(p.running)
			queued := len(p.queue)
			p.mu.RUnlock()
			
			if running == 0 && queued == 0 {
				return nil
			}
		}
	}
}

// Shutdown gracefully shuts down the execution pool
func (p *ExecutionPool) Shutdown() {
	p.mu.Lock()
	if p.shutdown {
		p.mu.Unlock()
		return
	}
	p.shutdown = true
	p.mu.Unlock()
	
	close(p.shutdownChan)
	
	// Wait for manager to finish
	<-p.doneChan
	
	logOperation("ExecutionPoolShutdown", map[string]interface{}{
		"total_submitted": p.stats.TotalSubmitted,
		"completed":       p.stats.CompletedExecutions,
		"failed":         p.stats.FailedExecutions,
	})
}

// Internal Implementation

// managePool is the main pool management goroutine
func (p *ExecutionPool) managePool() {
	defer close(p.doneChan)
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-p.shutdownChan:
			// Cancel all running executions
			p.cancelAllRunning()
			return
			
		case result := <-p.submitChan:
			p.handleSubmission(result)
			
		case executionID := <-p.cancelChan:
			p.handleCancellation(executionID)
			
		case <-ticker.C:
			p.processQueue()
			p.checkTimeouts()
		}
	}
}

// handleSubmission processes a new execution submission
func (p *ExecutionPool) handleSubmission(result *PoolExecutionResult) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Add to queue with priority sorting
	p.queue = append(p.queue, result)
	p.sortQueue()
	p.stats.QueuedExecutions = len(p.queue)
}

// processQueue processes the execution queue
func (p *ExecutionPool) processQueue() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Start executions if we have capacity
	for len(p.queue) > 0 && len(p.running) < p.config.MaxConcurrentExecutions {
		result := p.queue[0]
		p.queue = p.queue[1:]
		
		// Move to running
		p.running[result.ExecutionID] = result
		
		// Start execution in background
		p.workerWg.Add(1)
		go p.executeInBackground(result)
	}
}

// executeInBackground runs a Step Functions execution in the background
func (p *ExecutionPool) executeInBackground(result *PoolExecutionResult) {
	defer p.workerWg.Done()
	
	// Update status to running
	p.mu.Lock()
	result.Status = PoolExecutionStatusRunning
	result.StartedAt = time.Now()
	p.mu.Unlock()
	
	// Execute the Step Function
	execResult, err := StartExecutionE(p.ctx, result.request.StateMachineArn, result.request.ExecutionName, result.request.Input)
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Remove from running
	delete(p.running, result.ExecutionID)
	
	// Update result
	result.CompletedAt = time.Now()
	if err != nil {
		result.Status = PoolExecutionStatusFailed
		result.Error = err
		p.stats.FailedExecutions++
	} else {
		result.Status = PoolExecutionStatusCompleted
		result.Result = execResult
		result.ExecutionArn = execResult.ExecutionArn
		p.stats.CompletedExecutions++
	}
}

// handleCancellation processes an execution cancellation request
func (p *ExecutionPool) handleCancellation(executionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	result, exists := p.executions[executionID]
	if !exists {
		return
	}
	
	switch result.Status {
	case PoolExecutionStatusQueued:
		// Remove from queue
		for i, queuedResult := range p.queue {
			if queuedResult.ExecutionID == executionID {
				p.queue = append(p.queue[:i], p.queue[i+1:]...)
				break
			}
		}
		result.Status = PoolExecutionStatusCancelled
		result.CompletedAt = time.Now()
		p.stats.CancelledExecutions++
		
	case PoolExecutionStatusRunning:
		// For running executions, we'd need to call StopExecution
		// For now, just mark as cancelled (real implementation would stop the AWS execution)
		result.Status = PoolExecutionStatusCancelled
		result.CompletedAt = time.Now()
		p.stats.CancelledExecutions++
		delete(p.running, executionID)
	}
}

// checkTimeouts checks for execution timeouts
func (p *ExecutionPool) checkTimeouts() {
	if p.config.ExecutionTimeout == 0 {
		return
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	now := time.Now()
	for executionID, result := range p.running {
		if result.Status == PoolExecutionStatusRunning && !result.StartedAt.IsZero() {
			if now.Sub(result.StartedAt) > p.config.ExecutionTimeout {
				result.Status = PoolExecutionStatusTimeout
				result.CompletedAt = now
				result.Error = fmt.Errorf("execution timeout after %v", p.config.ExecutionTimeout)
				delete(p.running, executionID)
				p.stats.FailedExecutions++
			}
		}
	}
}

// cancelAllRunning cancels all running executions during shutdown
func (p *ExecutionPool) cancelAllRunning() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for executionID, result := range p.running {
		result.Status = PoolExecutionStatusCancelled
		result.CompletedAt = time.Now()
		p.stats.CancelledExecutions++
		delete(p.running, executionID)
	}
	
	// Also cancel queued executions
	for _, result := range p.queue {
		result.Status = PoolExecutionStatusCancelled
		result.CompletedAt = time.Now()
		p.stats.CancelledExecutions++
	}
	p.queue = nil
}

// sortQueue sorts the execution queue by priority (higher priority first)
func (p *ExecutionPool) sortQueue() {
	sort.Slice(p.queue, func(i, j int) bool {
		return p.queue[i].request.Priority > p.queue[j].request.Priority
	})
}

// Validation functions

// validateExecutionRequest validates an execution request
func (p *ExecutionPool) validateExecutionRequest(request ExecutionRequest) error {
	if request.StateMachineArn == "" {
		return fmt.Errorf("state machine ARN is required")
	}
	
	if err := validateStateMachineArn(request.StateMachineArn); err != nil {
		return fmt.Errorf("invalid state machine ARN: %w", err)
	}
	
	if request.ExecutionName == "" {
		return fmt.Errorf("execution name is required")
	}
	
	if err := validateExecutionName(request.ExecutionName); err != nil {
		return fmt.Errorf("invalid execution name: %w", err)
	}
	
	// Input validation (if provided)
	if request.Input != nil && !request.Input.isEmpty() {
		if _, err := request.Input.ToJSON(); err != nil {
			return fmt.Errorf("invalid input JSON: %w", err)
		}
	}
	
	return nil
}

// Configuration functions

// defaultExecutionPoolConfig provides sensible defaults
func defaultExecutionPoolConfig() *ExecutionPoolConfig {
	return &ExecutionPoolConfig{
		MaxConcurrentExecutions: 3,
		MaxQueueSize:           10,
		ExecutionTimeout:       DefaultTimeout,
		PoolTimeout:           10 * time.Minute,
	}
}

// mergeExecutionPoolConfig merges user config with defaults
func mergeExecutionPoolConfig(userConfig *ExecutionPoolConfig) *ExecutionPoolConfig {
	defaults := defaultExecutionPoolConfig()
	
	if userConfig == nil {
		return defaults
	}
	
	merged := *userConfig
	
	if merged.MaxConcurrentExecutions == 0 {
		merged.MaxConcurrentExecutions = defaults.MaxConcurrentExecutions
	}
	if merged.MaxQueueSize == 0 {
		merged.MaxQueueSize = defaults.MaxQueueSize
	}
	if merged.ExecutionTimeout == 0 {
		merged.ExecutionTimeout = defaults.ExecutionTimeout
	}
	if merged.PoolTimeout == 0 {
		merged.PoolTimeout = defaults.PoolTimeout
	}
	
	return &merged
}

// Utility functions

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	return fmt.Sprintf("pool-exec-%s", uuid.New().String()[:8])
}