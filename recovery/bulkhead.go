package recovery

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// BulkheadConfig defines the configuration for a bulkhead
type BulkheadConfig struct {
	Name         string
	MaxConcurrent int
	QueueSize    int
	Timeout      time.Duration
}

// BulkheadMetrics provides metrics about bulkhead state
type BulkheadMetrics struct {
	Name            string
	MaxConcurrent   int
	QueueSize       int
	ActiveRequests  int32
	QueuedRequests  int32
	TotalRequests   int64
	TotalSuccesses  int64
	TotalFailures   int64
	TotalRejections int64
	TotalTimeouts   int64
	AverageWaitTime time.Duration
}

// Bulkhead implements the bulkhead pattern for isolating different types of operations
type Bulkhead struct {
	config           BulkheadConfig
	semaphore        chan struct{}
	queue            chan *bulkheadRequest
	activeRequests   int32
	queuedRequests   int32
	totalRequests    int64
	totalSuccesses   int64
	totalFailures    int64
	totalRejections  int64
	totalTimeouts    int64
	mutex            sync.RWMutex
}

// bulkheadRequest represents a queued request
type bulkheadRequest struct {
	operation func() (interface{}, error)
	result    chan bulkheadResult
	startTime time.Time
}

// bulkheadResult represents the result of a bulkhead operation
type bulkheadResult struct {
	value interface{}
	err   error
}

// BulkheadManager manages multiple bulkheads for different service types
type BulkheadManager struct {
	bulkheads map[string]*Bulkhead
	mutex     sync.RWMutex
}

// NewBulkhead creates a new bulkhead with the given configuration
func NewBulkhead(config BulkheadConfig) *Bulkhead {
	// Set default values if not specified
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = 10
	}
	if config.QueueSize == 0 {
		config.QueueSize = 100
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	bulkhead := &Bulkhead{
		config:    config,
		semaphore: make(chan struct{}, config.MaxConcurrent),
		queue:     make(chan *bulkheadRequest, config.QueueSize),
	}

	// Start worker goroutines to process queued requests
	for i := 0; i < config.MaxConcurrent; i++ {
		go bulkhead.worker()
	}

	slog.Info("Bulkhead created",
		"name", config.Name,
		"maxConcurrent", config.MaxConcurrent,
		"queueSize", config.QueueSize,
		"timeout", config.Timeout,
	)

	return bulkhead
}

// Name returns the name of the bulkhead
func (b *Bulkhead) Name() string {
	return b.config.Name
}

// MaxConcurrent returns the maximum concurrent operations allowed
func (b *Bulkhead) MaxConcurrent() int {
	return b.config.MaxConcurrent
}

// QueueSize returns the size of the operation queue
func (b *Bulkhead) QueueSize() int {
	return b.config.QueueSize
}

// worker processes queued requests
func (b *Bulkhead) worker() {
	for request := range b.queue {
		atomic.AddInt32(&b.queuedRequests, -1)
		atomic.AddInt32(&b.activeRequests, 1)

		// Execute the operation
		result, err := request.operation()
		
		atomic.AddInt32(&b.activeRequests, -1)

		// Send result back
		select {
		case request.result <- bulkheadResult{value: result, err: err}:
		default:
			// Channel was closed due to timeout
		}

		close(request.result)
	}
}

// Execute executes an operation through the bulkhead
func (b *Bulkhead) Execute(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	atomic.AddInt64(&b.totalRequests, 1)

	request := &bulkheadRequest{
		operation: operation,
		result:    make(chan bulkheadResult, 1),
		startTime: time.Now(),
	}

	// Try to queue the request
	select {
	case b.queue <- request:
		atomic.AddInt32(&b.queuedRequests, 1)
		slog.Debug("Request queued",
			"bulkhead", b.config.Name,
			"queueSize", atomic.LoadInt32(&b.queuedRequests),
		)
	default:
		// Queue is full
		atomic.AddInt64(&b.totalRejections, 1)
		slog.Warn("Request rejected - bulkhead is full",
			"bulkhead", b.config.Name,
			"maxConcurrent", b.config.MaxConcurrent,
			"queueSize", b.config.QueueSize,
		)
		return nil, errors.New("bulkhead is full - request rejected")
	}

	// Wait for result or timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, b.config.Timeout)
	defer cancel()

	select {
	case result := <-request.result:
		if result.err != nil {
			atomic.AddInt64(&b.totalFailures, 1)
		} else {
			atomic.AddInt64(&b.totalSuccesses, 1)
		}
		return result.value, result.err

	case <-timeoutCtx.Done():
		atomic.AddInt64(&b.totalTimeouts, 1)
		slog.Warn("Request timed out",
			"bulkhead", b.config.Name,
			"timeout", b.config.Timeout,
		)
		return nil, errors.New("bulkhead operation timeout")
	}
}

// GetMetrics returns current metrics for the bulkhead
func (b *Bulkhead) GetMetrics() BulkheadMetrics {
	return BulkheadMetrics{
		Name:            b.config.Name,
		MaxConcurrent:   b.config.MaxConcurrent,
		QueueSize:       b.config.QueueSize,
		ActiveRequests:  atomic.LoadInt32(&b.activeRequests),
		QueuedRequests:  atomic.LoadInt32(&b.queuedRequests),
		TotalRequests:   atomic.LoadInt64(&b.totalRequests),
		TotalSuccesses:  atomic.LoadInt64(&b.totalSuccesses),
		TotalFailures:   atomic.LoadInt64(&b.totalFailures),
		TotalRejections: atomic.LoadInt64(&b.totalRejections),
		TotalTimeouts:   atomic.LoadInt64(&b.totalTimeouts),
	}
}

// NewBulkheadManager creates a new bulkhead manager
func NewBulkheadManager() *BulkheadManager {
	return &BulkheadManager{
		bulkheads: make(map[string]*Bulkhead),
	}
}

// AddBulkhead adds a bulkhead for a specific service type
func (bm *BulkheadManager) AddBulkhead(serviceType string, config BulkheadConfig) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	bulkhead := NewBulkhead(config)
	bm.bulkheads[serviceType] = bulkhead

	slog.Info("Bulkhead added to manager",
		"serviceType", serviceType,
		"bulkheadName", config.Name,
	)
}

// GetBulkhead returns the bulkhead for a specific service type
func (bm *BulkheadManager) GetBulkhead(serviceType string) *Bulkhead {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	return bm.bulkheads[serviceType]
}

// ExecuteWithServiceBulkhead executes an operation using the bulkhead for the specified service type
func (bm *BulkheadManager) ExecuteWithServiceBulkhead(ctx context.Context, serviceType string, operation func() (interface{}, error)) (interface{}, error) {
	bulkhead := bm.GetBulkhead(serviceType)
	if bulkhead == nil {
		return nil, errors.New("no bulkhead configured for service type: " + serviceType)
	}

	return bulkhead.Execute(ctx, operation)
}

// GetAllMetrics returns metrics for all managed bulkheads
func (bm *BulkheadManager) GetAllMetrics() map[string]BulkheadMetrics {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	metrics := make(map[string]BulkheadMetrics)
	for serviceType, bulkhead := range bm.bulkheads {
		metrics[serviceType] = bulkhead.GetMetrics()
	}

	return metrics
}

// ExecuteWithBulkhead executes an operation with bulkhead protection
func ExecuteWithBulkhead[T any](ctx context.Context, bulkhead *Bulkhead, operation func() (T, error)) (T, error) {
	var zeroValue T

	result, err := bulkhead.Execute(ctx, func() (interface{}, error) {
		return operation()
	})

	if err != nil {
		return zeroValue, err
	}

	// Type assertion
	typedResult, ok := result.(T)
	if !ok {
		return zeroValue, errors.New("unexpected result type from bulkhead operation")
	}

	return typedResult, nil
}

// ExecuteWithServiceBulkhead is a typed version of the bulkhead manager execution
func ExecuteWithServiceBulkhead[T any](ctx context.Context, manager *BulkheadManager, serviceType string, operation func() (T, error)) (T, error) {
	var zeroValue T

	result, err := manager.ExecuteWithServiceBulkhead(ctx, serviceType, func() (interface{}, error) {
		return operation()
	})

	if err != nil {
		return zeroValue, err
	}

	// Type assertion
	typedResult, ok := result.(T)
	if !ok {
		return zeroValue, errors.New("unexpected result type from service bulkhead operation")
	}

	return typedResult, nil
}