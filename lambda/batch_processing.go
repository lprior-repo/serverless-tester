package lambda

import (
	"fmt"
	"sync"
	"time"
)

// BatchProcessingConfig defines configuration for batch processing operations
type BatchProcessingConfig struct {
	ConcurrencyLimit int           // Maximum number of concurrent invocations
	RetryAttempts    int           // Number of retry attempts for failed invocations
	RetryDelay       time.Duration // Delay between retry attempts
	CollectMetrics   bool          // Whether to collect processing metrics
}

// BatchProcessingMetrics contains metrics collected during batch processing
type BatchProcessingMetrics struct {
	TotalProcessed         int
	SuccessfulProcessed    int
	FailedProcessed        int
	TotalProcessingTime    time.Duration
	AverageProcessingTime  time.Duration
	MaxProcessingTime      time.Duration
	MinProcessingTime      time.Duration
}

// Global metrics storage for batch processing
var batchMetrics *BatchProcessingMetrics
var metricsMutex sync.RWMutex

// DefaultBatchProcessingConfig returns a default configuration for batch processing
func DefaultBatchProcessingConfig() BatchProcessingConfig {
	return BatchProcessingConfig{
		ConcurrencyLimit: 10,
		RetryAttempts:   3,
		RetryDelay:      100 * time.Millisecond,
		CollectMetrics:  false,
	}
}

// BatchProcessSQSMessages processes multiple SQS events in batch using default configuration.
// This is the non-error returning version that follows Terratest patterns.
func BatchProcessSQSMessages(ctx *TestContext, functionName string, sqsEvents []string) []*InvokeResult {
	results, err := BatchProcessSQSMessagesE(ctx, functionName, sqsEvents)
	if err != nil {
		ctx.T.Errorf("Failed to batch process SQS messages: %v", err)
		ctx.T.FailNow()
	}
	return results
}

// BatchProcessSQSMessagesE processes multiple SQS events in batch using default configuration.
// This is the error returning version that follows Terratest patterns.
func BatchProcessSQSMessagesE(ctx *TestContext, functionName string, sqsEvents []string) ([]*InvokeResult, error) {
	return BatchProcessSQSMessagesWithConfigE(ctx, functionName, sqsEvents, DefaultBatchProcessingConfig())
}

// BatchProcessSQSMessagesWithConfig processes multiple SQS events in batch using custom configuration.
// This is the non-error returning version that follows Terratest patterns.
func BatchProcessSQSMessagesWithConfig(ctx *TestContext, functionName string, sqsEvents []string, config BatchProcessingConfig) []*InvokeResult {
	results, err := BatchProcessSQSMessagesWithConfigE(ctx, functionName, sqsEvents, config)
	if err != nil {
		ctx.T.Errorf("Failed to batch process SQS messages with config: %v", err)
		ctx.T.FailNow()
	}
	return results
}

// BatchProcessSQSMessagesWithConfigE processes multiple SQS events in batch using custom configuration.
// This is the error returning version that follows Terratest patterns.
func BatchProcessSQSMessagesWithConfigE(ctx *TestContext, functionName string, sqsEvents []string, config BatchProcessingConfig) ([]*InvokeResult, error) {
	if err := validateBatchProcessingConfig(config); err != nil {
		return nil, fmt.Errorf("invalid batch processing configuration: %w", err)
	}
	
	if err := validateFunctionName(functionName); err != nil {
		return nil, err
	}
	
	if len(sqsEvents) == 0 {
		return []*InvokeResult{}, nil
	}
	
	logOperation("batch_process_sqs", functionName, map[string]interface{}{
		"event_count":       len(sqsEvents),
		"concurrency_limit": config.ConcurrencyLimit,
		"retry_attempts":    config.RetryAttempts,
	})
	
	// Initialize metrics if collection is enabled
	if config.CollectMetrics {
		initializeBatchMetrics()
	}
	
	startTime := time.Now()
	
	// Process events in batches with concurrency control
	results, err := processBatchWithConcurrency(ctx, functionName, sqsEvents, config)
	
	duration := time.Since(startTime)
	
	if err != nil {
		logResult("batch_process_sqs", functionName, false, duration, err)
		return nil, fmt.Errorf("failed to batch process SQS messages: %w", err)
	}
	
	// Update metrics if collection is enabled
	if config.CollectMetrics {
		updateBatchMetrics(results, duration)
	}
	
	logResult("batch_process_sqs", functionName, true, duration, nil)
	
	return results, nil
}

// BatchProcessDynamoDBEvents processes multiple DynamoDB stream events in batch using default configuration.
// This is the non-error returning version that follows Terratest patterns.
func BatchProcessDynamoDBEvents(ctx *TestContext, functionName string, dynamodbEvents []string) []*InvokeResult {
	results, err := BatchProcessDynamoDBEventsE(ctx, functionName, dynamodbEvents)
	if err != nil {
		ctx.T.Errorf("Failed to batch process DynamoDB events: %v", err)
		ctx.T.FailNow()
	}
	return results
}

// BatchProcessDynamoDBEventsE processes multiple DynamoDB stream events in batch using default configuration.
// This is the error returning version that follows Terratest patterns.
func BatchProcessDynamoDBEventsE(ctx *TestContext, functionName string, dynamodbEvents []string) ([]*InvokeResult, error) {
	return BatchProcessDynamoDBEventsWithConfigE(ctx, functionName, dynamodbEvents, DefaultBatchProcessingConfig())
}

// BatchProcessDynamoDBEventsWithConfigE processes multiple DynamoDB stream events in batch using custom configuration.
// This is the error returning version that follows Terratest patterns.
func BatchProcessDynamoDBEventsWithConfigE(ctx *TestContext, functionName string, dynamodbEvents []string, config BatchProcessingConfig) ([]*InvokeResult, error) {
	if err := validateBatchProcessingConfig(config); err != nil {
		return nil, fmt.Errorf("invalid batch processing configuration: %w", err)
	}
	
	if err := validateFunctionName(functionName); err != nil {
		return nil, err
	}
	
	if len(dynamodbEvents) == 0 {
		return []*InvokeResult{}, nil
	}
	
	logOperation("batch_process_dynamodb", functionName, map[string]interface{}{
		"event_count":       len(dynamodbEvents),
		"concurrency_limit": config.ConcurrencyLimit,
		"retry_attempts":    config.RetryAttempts,
	})
	
	// Initialize metrics if collection is enabled
	if config.CollectMetrics {
		initializeBatchMetrics()
	}
	
	startTime := time.Now()
	
	// Process events in batches with concurrency control
	results, err := processBatchWithConcurrency(ctx, functionName, dynamodbEvents, config)
	
	duration := time.Since(startTime)
	
	if err != nil {
		logResult("batch_process_dynamodb", functionName, false, duration, err)
		return nil, fmt.Errorf("failed to batch process DynamoDB events: %w", err)
	}
	
	// Update metrics if collection is enabled
	if config.CollectMetrics {
		updateBatchMetrics(results, duration)
	}
	
	logResult("batch_process_dynamodb", functionName, true, duration, nil)
	
	return results, nil
}

// BatchProcessKinesisEventsWithConfigE processes multiple Kinesis stream events in batch using custom configuration.
// This is the error returning version that follows Terratest patterns.
func BatchProcessKinesisEventsWithConfigE(ctx *TestContext, functionName string, kinesisEvents []string, config BatchProcessingConfig) ([]*InvokeResult, error) {
	if err := validateBatchProcessingConfig(config); err != nil {
		return nil, fmt.Errorf("invalid batch processing configuration: %w", err)
	}
	
	if err := validateFunctionName(functionName); err != nil {
		return nil, err
	}
	
	if len(kinesisEvents) == 0 {
		return []*InvokeResult{}, nil
	}
	
	logOperation("batch_process_kinesis", functionName, map[string]interface{}{
		"event_count":       len(kinesisEvents),
		"concurrency_limit": config.ConcurrencyLimit,
		"retry_attempts":    config.RetryAttempts,
	})
	
	// Initialize metrics if collection is enabled
	if config.CollectMetrics {
		initializeBatchMetrics()
	}
	
	startTime := time.Now()
	
	// Process events in batches with concurrency control
	results, err := processBatchWithConcurrency(ctx, functionName, kinesisEvents, config)
	
	duration := time.Since(startTime)
	
	if err != nil {
		logResult("batch_process_kinesis", functionName, false, duration, err)
		return nil, fmt.Errorf("failed to batch process Kinesis events: %w", err)
	}
	
	// Update metrics if collection is enabled
	if config.CollectMetrics {
		updateBatchMetrics(results, duration)
	}
	
	logResult("batch_process_kinesis", functionName, true, duration, nil)
	
	return results, nil
}

// processBatchWithConcurrency processes events with controlled concurrency and retry logic
func processBatchWithConcurrency(ctx *TestContext, functionName string, events []string, config BatchProcessingConfig) ([]*InvokeResult, error) {
	eventCount := len(events)
	results := make([]*InvokeResult, eventCount)
	
	// Create semaphore to control concurrency
	semaphore := make(chan struct{}, config.ConcurrencyLimit)
	
	// Create error group for coordinated processing
	var wg sync.WaitGroup
	var mu sync.Mutex
	var processingErrors []error
	
	for i, event := range events {
		wg.Add(1)
		
		go func(index int, eventPayload string) {
			defer wg.Done()
			
			// Acquire semaphore to limit concurrency
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Process single event with retry logic
			result, err := processEventWithRetry(ctx, functionName, eventPayload, config)
			
			mu.Lock()
			if err != nil {
				processingErrors = append(processingErrors, fmt.Errorf("event %d failed: %w", index, err))
				// Create error result
				results[index] = &InvokeResult{
					StatusCode: 500,
					Payload:    fmt.Sprintf(`{"error": "processing failed: %s"}`, err.Error()),
					ExecutionTime: 0,
				}
			} else {
				results[index] = result
			}
			mu.Unlock()
			
		}(i, event)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	
	// Return results even if some events failed (partial success)
	// This allows callers to handle failures individually
	return results, nil
}

// processEventWithRetry processes a single event with retry logic
func processEventWithRetry(ctx *TestContext, functionName, eventPayload string, config BatchProcessingConfig) (*InvokeResult, error) {
	var lastError error
	
	for attempt := 0; attempt <= config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Apply delay before retry
			time.Sleep(config.RetryDelay)
		}
		
		startTime := time.Now()
		result, err := InvokeE(ctx, functionName, eventPayload)
		duration := time.Since(startTime)
		
		if err != nil {
			lastError = err
			continue
		}
		
		// Set execution duration
		result.ExecutionTime = duration
		
		// Check if function execution was successful
		if result.FunctionError == "" {
			return result, nil
		}
		
		// Function error occurred, prepare for retry
		lastError = fmt.Errorf("function error: %s", result.FunctionError)
		
		// For the last attempt, still return the result with error
		if attempt == config.RetryAttempts {
			return result, nil
		}
	}
	
	return nil, lastError
}

// validateBatchProcessingConfig validates the batch processing configuration
func validateBatchProcessingConfig(config BatchProcessingConfig) error {
	if config.ConcurrencyLimit <= 0 {
		return fmt.Errorf("concurrency limit must be greater than 0, got %d", config.ConcurrencyLimit)
	}
	
	if config.ConcurrencyLimit > 1000 {
		return fmt.Errorf("concurrency limit exceeds maximum allowed (1000), got %d", config.ConcurrencyLimit)
	}
	
	if config.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts must be non-negative, got %d", config.RetryAttempts)
	}
	
	if config.RetryDelay < 0 {
		return fmt.Errorf("retry delay must be non-negative, got %v", config.RetryDelay)
	}
	
	return nil
}

// calculateOptimalBatchSize calculates optimal batch size for given parameters
func calculateOptimalBatchSize(totalItems, concurrencyLimit int) (batchSize int, batchCount int) {
	if totalItems <= concurrencyLimit {
		return 1, totalItems
	}
	
	// Calculate batch size to evenly distribute work
	batchSize = (totalItems + concurrencyLimit - 1) / concurrencyLimit
	batchCount = concurrencyLimit
	
	return batchSize, batchCount
}

// initializeBatchMetrics initializes the global batch processing metrics
func initializeBatchMetrics() {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	
	batchMetrics = &BatchProcessingMetrics{
		TotalProcessed:      0,
		SuccessfulProcessed: 0,
		FailedProcessed:     0,
		TotalProcessingTime: 0,
		MinProcessingTime:   time.Hour, // Set to high value initially
		MaxProcessingTime:   0,
	}
}

// updateBatchMetrics updates the global batch processing metrics
func updateBatchMetrics(results []*InvokeResult, totalDuration time.Duration) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	
	if batchMetrics == nil {
		return
	}
	
	batchMetrics.TotalProcessed = len(results)
	batchMetrics.TotalProcessingTime = totalDuration
	
	// Calculate success/failure counts and individual processing times
	var successfulCount, failedCount int
	var minTime, maxTime time.Duration = time.Hour, 0
	
	for _, result := range results {
		if result.FunctionError == "" {
			successfulCount++
		} else {
			failedCount++
		}
		
		// Track individual processing times if available
		if result.ExecutionTime > 0 {
			if result.ExecutionTime < minTime {
				minTime = result.ExecutionTime
			}
			if result.ExecutionTime > maxTime {
				maxTime = result.ExecutionTime
			}
		}
	}
	
	batchMetrics.SuccessfulProcessed = successfulCount
	batchMetrics.FailedProcessed = failedCount
	
	if len(results) > 0 {
		batchMetrics.AverageProcessingTime = totalDuration / time.Duration(len(results))
	}
	
	if minTime < time.Hour {
		batchMetrics.MinProcessingTime = minTime
	}
	batchMetrics.MaxProcessingTime = maxTime
}

// GetBatchProcessingMetrics returns the current batch processing metrics
func GetBatchProcessingMetrics() *BatchProcessingMetrics {
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()
	
	if batchMetrics == nil {
		return nil
	}
	
	// Return a copy to prevent external modification
	return &BatchProcessingMetrics{
		TotalProcessed:        batchMetrics.TotalProcessed,
		SuccessfulProcessed:   batchMetrics.SuccessfulProcessed,
		FailedProcessed:       batchMetrics.FailedProcessed,
		TotalProcessingTime:   batchMetrics.TotalProcessingTime,
		AverageProcessingTime: batchMetrics.AverageProcessingTime,
		MaxProcessingTime:     batchMetrics.MaxProcessingTime,
		MinProcessingTime:     batchMetrics.MinProcessingTime,
	}
}

// ResetBatchProcessingMetrics resets the global batch processing metrics
func ResetBatchProcessingMetrics() {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	
	batchMetrics = nil
}

// CreateOptimizedBatchConfig creates an optimized batch processing configuration based on requirements
func CreateOptimizedBatchConfig(eventCount int, targetLatency time.Duration, maxConcurrency int) BatchProcessingConfig {
	// Calculate optimal concurrency based on event count and target latency
	concurrency := eventCount / 10 // Start with 1 worker per 10 events
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > maxConcurrency {
		concurrency = maxConcurrency
	}
	
	// Calculate retry attempts based on target latency
	retryAttempts := 3
	retryDelay := targetLatency / time.Duration(retryAttempts+1)
	
	// Ensure reasonable bounds
	if retryDelay > 5*time.Second {
		retryDelay = 5 * time.Second
	}
	if retryDelay < 10*time.Millisecond {
		retryDelay = 10 * time.Millisecond
	}
	
	return BatchProcessingConfig{
		ConcurrencyLimit: concurrency,
		RetryAttempts:   retryAttempts,
		RetryDelay:      retryDelay,
		CollectMetrics:  true,
	}
}

// AnalyzeBatchPerformance analyzes batch processing performance and provides optimization recommendations
func AnalyzeBatchPerformance(metrics *BatchProcessingMetrics, eventCount int) map[string]interface{} {
	if metrics == nil {
		return map[string]interface{}{
			"error": "no metrics available for analysis",
		}
	}
	
	analysis := map[string]interface{}{
		"total_events":         eventCount,
		"success_rate":         float64(metrics.SuccessfulProcessed) / float64(metrics.TotalProcessed),
		"failure_rate":         float64(metrics.FailedProcessed) / float64(metrics.TotalProcessed),
		"average_latency_ms":   metrics.AverageProcessingTime.Milliseconds(),
		"total_duration_ms":    metrics.TotalProcessingTime.Milliseconds(),
	}
	
	// Performance assessment
	successRate := float64(metrics.SuccessfulProcessed) / float64(metrics.TotalProcessed)
	if successRate >= 0.99 {
		analysis["performance_grade"] = "A"
		analysis["recommendation"] = "Excellent performance. Consider increasing batch size for better throughput."
	} else if successRate >= 0.95 {
		analysis["performance_grade"] = "B"
		analysis["recommendation"] = "Good performance. Monitor error patterns for improvement opportunities."
	} else if successRate >= 0.90 {
		analysis["performance_grade"] = "C"
		analysis["recommendation"] = "Moderate performance. Consider increasing retry attempts or investigating error causes."
	} else {
		analysis["performance_grade"] = "D"
		analysis["recommendation"] = "Poor performance. Investigate error causes and consider reducing concurrency."
	}
	
	// Throughput analysis
	throughputEventsPerSecond := float64(metrics.TotalProcessed) / metrics.TotalProcessingTime.Seconds()
	analysis["throughput_events_per_second"] = throughputEventsPerSecond
	
	if throughputEventsPerSecond > 100 {
		analysis["throughput_grade"] = "High"
	} else if throughputEventsPerSecond > 10 {
		analysis["throughput_grade"] = "Medium"
	} else {
		analysis["throughput_grade"] = "Low"
	}
	
	return analysis
}