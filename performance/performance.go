// Package performance provides comprehensive benchmarking and optimization for the VasDeference framework
// Implements performance monitoring, memory pooling, load testing, and regression detection
package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryPool provides object pooling to reduce garbage collection overhead
type MemoryPool[T any] struct {
	pool    sync.Pool
	factory func() T
}

// NewMemoryPool creates a new memory pool with the given factory function
func NewMemoryPool[T any](factory func() T) *MemoryPool[T] {
	return &MemoryPool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return factory()
			},
		},
		factory: factory,
	}
}

// Get retrieves an object from the pool or creates a new one
func (p *MemoryPool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns an object to the pool for reuse
func (p *MemoryPool[T]) Put(obj T) {
	p.pool.Put(obj)
}

// PerformanceMonitor tracks real-time performance metrics
type PerformanceMonitor struct {
	mu           sync.RWMutex
	metrics      map[string]*MetricData
	running      bool
	stopChan     chan struct{}
	errors       map[string]int64
	startTime    time.Time
}

// MetricData stores performance data for a specific operation
type MetricData struct {
	Count          int64
	TotalLatency   int64 // nanoseconds
	MinLatency     int64
	MaxLatency     int64
	Errors         int64
	AverageLatency time.Duration
	lastUpdate     time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:   make(map[string]*MetricData),
		errors:    make(map[string]int64),
		stopChan:  make(chan struct{}),
		startTime: time.Now(),
	}
}

// Start begins performance monitoring
func (pm *PerformanceMonitor) Start() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if pm.running {
		return
	}
	
	pm.running = true
	pm.startTime = time.Now()
	
	go pm.monitorLoop()
}

// Stop ends performance monitoring
func (pm *PerformanceMonitor) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if !pm.running {
		return
	}
	
	pm.running = false
	close(pm.stopChan)
}

// RecordLatency records a latency measurement for an operation
func (pm *PerformanceMonitor) RecordLatency(operation string, latency time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	metric, exists := pm.metrics[operation]
	if !exists {
		metric = &MetricData{
			MinLatency: latency.Nanoseconds(),
			MaxLatency: latency.Nanoseconds(),
		}
		pm.metrics[operation] = metric
	}
	
	// Update counters
	atomic.AddInt64(&metric.Count, 1)
	atomic.AddInt64(&metric.TotalLatency, latency.Nanoseconds())
	
	// Update min/max (not atomic for simplicity in this implementation)
	latencyNs := latency.Nanoseconds()
	if latencyNs < metric.MinLatency {
		metric.MinLatency = latencyNs
	}
	if latencyNs > metric.MaxLatency {
		metric.MaxLatency = latencyNs
	}
	
	// Calculate average
	count := atomic.LoadInt64(&metric.Count)
	totalLatency := atomic.LoadInt64(&metric.TotalLatency)
	metric.AverageLatency = time.Duration(totalLatency / count)
	metric.lastUpdate = time.Now()
}

// RecordError records an error for an operation
func (pm *PerformanceMonitor) RecordError(operation string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.errors[operation] = pm.errors[operation] + 1
	
	// Also update the metric data if it exists, or create it
	metric, exists := pm.metrics[operation]
	if !exists {
		metric = &MetricData{}
		pm.metrics[operation] = metric
	}
	atomic.AddInt64(&metric.Errors, 1)
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() map[string]*MetricData {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	// Return a copy to avoid concurrent access issues
	result := make(map[string]*MetricData)
	for k, v := range pm.metrics {
		result[k] = &MetricData{
			Count:          atomic.LoadInt64(&v.Count),
			TotalLatency:   atomic.LoadInt64(&v.TotalLatency),
			MinLatency:     v.MinLatency,
			MaxLatency:     v.MaxLatency,
			Errors:         atomic.LoadInt64(&v.Errors),
			AverageLatency: v.AverageLatency,
			lastUpdate:     v.lastUpdate,
		}
	}
	
	return result
}

// monitorLoop runs background monitoring tasks
func (pm *PerformanceMonitor) monitorLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			pm.updateStatistics()
		case <-pm.stopChan:
			return
		}
	}
}

// updateStatistics periodically updates derived statistics
func (pm *PerformanceMonitor) updateStatistics() {
	// Could implement additional background calculations here
	// For now, statistics are calculated on-demand
}

// LoadTestConfig configures a load test execution
type LoadTestConfig struct {
	TestName       string
	Duration       time.Duration
	RequestsPerSec int
	Concurrency    int
	TestFunction   func() error
}

// LoadTestResults contains the results of a load test
type LoadTestResults struct {
	TestName        string
	TotalRequests   int64
	SuccessfulReqs  int64
	FailedReqs      int64
	ActualRPS       float64
	ErrorRate       float64
	AverageLatency  time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	Duration        time.Duration
}

// ExecuteLoadTest runs a load test with the given configuration
func ExecuteLoadTest(config LoadTestConfig) LoadTestResults {
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()
	
	var (
		totalRequests   int64
		successfulReqs  int64
		failedReqs      int64
		latencies       []time.Duration
		latenciesMutex  sync.Mutex
		wg              sync.WaitGroup
	)
	
	startTime := time.Now()
	
	// Rate limiter
	requestInterval := time.Second / time.Duration(config.RequestsPerSec)
	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()
	
	// Worker semaphore
	semaphore := make(chan struct{}, config.Concurrency)
	
	// Request generator loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				wg.Add(1)
				go func() {
					defer wg.Done()
					
					// Acquire worker slot
					semaphore <- struct{}{}
					defer func() { <-semaphore }()
					
					// Execute request
					reqStart := time.Now()
					err := config.TestFunction()
					latency := time.Since(reqStart)
					
					// Record results
					atomic.AddInt64(&totalRequests, 1)
					if err != nil {
						atomic.AddInt64(&failedReqs, 1)
					} else {
						atomic.AddInt64(&successfulReqs, 1)
					}
					
					// Record latency
					latenciesMutex.Lock()
					latencies = append(latencies, latency)
					latenciesMutex.Unlock()
				}()
			}
		}
	}()
	
	// Wait for test duration
	<-ctx.Done()
	wg.Wait()
	
	actualDuration := time.Since(startTime)
	actualRPS := float64(totalRequests) / actualDuration.Seconds()
	errorRate := float64(failedReqs) / float64(totalRequests)
	
	// Calculate latency percentiles
	p50, p95, p99, min, max := calculateLatencyPercentiles(latencies)
	avgLatency := calculateAverageLatency(latencies)
	
	return LoadTestResults{
		TestName:       config.TestName,
		TotalRequests:  totalRequests,
		SuccessfulReqs: successfulReqs,
		FailedReqs:     failedReqs,
		ActualRPS:      actualRPS,
		ErrorRate:      errorRate,
		AverageLatency: avgLatency,
		P50Latency:     p50,
		P95Latency:     p95,
		P99Latency:     p99,
		MinLatency:     min,
		MaxLatency:     max,
		Duration:       actualDuration,
	}
}

// RegressionDetector detects performance regressions
type RegressionDetector struct {
	baselines map[string]PerformanceMetrics
	mu        sync.RWMutex
}

// RegressionResult contains regression analysis results
type RegressionResult struct {
	TestName      string
	HasRegression bool
	Alerts        []string
	Severity      string
}

// NewRegressionDetector creates a new regression detector
func NewRegressionDetector() *RegressionDetector {
	return &RegressionDetector{
		baselines: make(map[string]PerformanceMetrics),
	}
}

// SetBaseline sets the performance baseline for a test
func (rd *RegressionDetector) SetBaseline(testName string, baseline PerformanceMetrics) {
	rd.mu.Lock()
	defer rd.mu.Unlock()
	
	rd.baselines[testName] = baseline
}

// CheckRegression checks if current metrics show regression compared to baseline
func (rd *RegressionDetector) CheckRegression(testName string, current PerformanceMetrics) RegressionResult {
	rd.mu.RLock()
	baseline, exists := rd.baselines[testName]
	rd.mu.RUnlock()
	
	if !exists {
		return RegressionResult{
			TestName:      testName,
			HasRegression: false,
			Alerts:        []string{"No baseline available for comparison"},
		}
	}
	
	var alerts []string
	severity := "none"
	
	// Check latency regression (>20% increase)
	if current.AverageLatency > baseline.AverageLatency*120/100 {
		increase := float64(current.AverageLatency-baseline.AverageLatency) / float64(baseline.AverageLatency) * 100
		alerts = append(alerts, fmt.Sprintf("Average latency increased by %.1f%% (%v → %v)", 
			increase, baseline.AverageLatency, current.AverageLatency))
		severity = "medium"
	}
	
	// Check P95 latency regression (>30% increase)
	if current.P95Latency > baseline.P95Latency*130/100 {
		increase := float64(current.P95Latency-baseline.P95Latency) / float64(baseline.P95Latency) * 100
		alerts = append(alerts, fmt.Sprintf("P95 latency increased by %.1f%% (%v → %v)", 
			increase, baseline.P95Latency, current.P95Latency))
		if severity == "none" {
			severity = "medium"
		}
	}
	
	// Check throughput regression (>20% decrease)
	if current.Throughput < baseline.Throughput*80/100 {
		decrease := float64(baseline.Throughput-current.Throughput) / float64(baseline.Throughput) * 100
		alerts = append(alerts, fmt.Sprintf("Throughput decreased by %.1f%% (%.2f → %.2f RPS)", 
			decrease, baseline.Throughput, current.Throughput))
		severity = "high"
	}
	
	// Check error rate regression (>2x increase)
	if current.ErrorRate > baseline.ErrorRate*2 && baseline.ErrorRate > 0 {
		multiplier := current.ErrorRate / baseline.ErrorRate
		alerts = append(alerts, fmt.Sprintf("Error rate increased by %.1fx (%.2f%% → %.2f%%)", 
			multiplier, baseline.ErrorRate*100, current.ErrorRate*100))
		severity = "high"
	}
	
	// Check memory usage regression (>50% increase)
	if baseline.MemoryUsed > 0 && current.MemoryUsed > baseline.MemoryUsed*150/100 {
		increase := float64(current.MemoryUsed-baseline.MemoryUsed) / float64(baseline.MemoryUsed) * 100
		alerts = append(alerts, fmt.Sprintf("Memory usage increased by %.1f%% (%d → %d bytes)", 
			increase, baseline.MemoryUsed, current.MemoryUsed))
		if severity == "none" {
			severity = "medium"
		}
	}
	
	return RegressionResult{
		TestName:      testName,
		HasRegression: len(alerts) > 0,
		Alerts:        alerts,
		Severity:      severity,
	}
}

// ExecutePerformanceBenchmark executes a performance benchmark
func ExecutePerformanceBenchmark(config PerformanceBenchmarkConfig) PerformanceMetrics {
	var (
		latencies      []time.Duration
		latenciesMutex sync.Mutex
		errorCount     int64
		wg             sync.WaitGroup
	)
	
	// Start memory profiling if requested
	var memStats1, memStats2 runtime.MemStats
	if config.MemoryProfile {
		runtime.GC()
		runtime.ReadMemStats(&memStats1)
	}
	
	startTime := time.Now()
	
	// Execute benchmark with controlled concurrency
	semaphore := make(chan struct{}, config.Concurrency)
	
	for i := 0; i < config.Iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Acquire worker slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Execute test function
			opStart := time.Now()
			err := config.TestFunction()
			latency := time.Since(opStart)
			
			// Record results
			latenciesMutex.Lock()
			latencies = append(latencies, latency)
			latenciesMutex.Unlock()
			
			if err != nil {
				atomic.AddInt64(&errorCount, 1)
			}
		}()
	}
	
	wg.Wait()
	duration := time.Since(startTime)
	
	// Collect memory stats if requested
	var memoryAllocated, memoryUsed int64
	if config.MemoryProfile {
		runtime.GC()
		runtime.ReadMemStats(&memStats2)
		memoryAllocated = int64(memStats2.TotalAlloc - memStats1.TotalAlloc)
		memoryUsed = int64(memStats2.Alloc)
	}
	
	// Calculate metrics
	p50, p95, p99, min, max := calculateLatencyPercentiles(latencies)
	avgLatency := calculateAverageLatency(latencies)
	throughput := float64(config.Iterations) / duration.Seconds()
	errorRate := float64(errorCount) / float64(config.Iterations)
	
	return PerformanceMetrics{
		TestName:        config.Name,
		AverageLatency:  avgLatency,
		P50Latency:     p50,
		P95Latency:     p95,
		P99Latency:     p99,
		MinLatency:     min,
		MaxLatency:     max,
		Throughput:     throughput,
		ErrorRate:      errorRate,
		MemoryAllocated: memoryAllocated,
		MemoryUsed:     memoryUsed,
		Timestamp:      time.Now(),
	}
}

// Utility functions for latency calculations

func calculateLatencyPercentiles(latencies []time.Duration) (p50, p95, p99, min, max time.Duration) {
	if len(latencies) == 0 {
		return 0, 0, 0, 0, 0
	}
	
	// Simple sorting for percentile calculation
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)
	
	// Basic bubble sort (good enough for testing)
	n := len(sortedLatencies)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if sortedLatencies[j] > sortedLatencies[j+1] {
				sortedLatencies[j], sortedLatencies[j+1] = sortedLatencies[j+1], sortedLatencies[j]
			}
		}
	}
	
	min = sortedLatencies[0]
	max = sortedLatencies[n-1]
	p50 = sortedLatencies[n*50/100]
	p95 = sortedLatencies[n*95/100]
	p99 = sortedLatencies[n*99/100]
	
	return p50, p95, p99, min, max
}

func calculateAverageLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}
	
	return total / time.Duration(len(latencies))
}