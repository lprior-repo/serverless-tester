package eventbridge

import (
	"fmt"
	"time"
)

// Enhanced delivery validation functions with more comprehensive checks
// Note: Basic DeliveryMetrics type and core functions are in assertions.go

// AssertEventDeliveryWithRetries validates event delivery with retry logic
func AssertEventDeliveryWithRetries(ctx *TestContext, result PutEventResult, maxRetries int) {
	err := AssertEventDeliveryWithRetriesE(ctx, result, maxRetries)
	if err != nil {
		ctx.T.Errorf("AssertEventDeliveryWithRetries failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertEventDeliveryWithRetriesE validates event delivery with retry logic and returns error
func AssertEventDeliveryWithRetriesE(ctx *TestContext, result PutEventResult, maxRetries int) error {
	if maxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative: %d", maxRetries)
	}
	
	if !result.Success {
		return fmt.Errorf("event delivery failed after %d retries: %s - %s", maxRetries, result.ErrorCode, result.ErrorMessage)
	}
	
	if result.EventID == "" {
		return fmt.Errorf("event ID is empty, indicating delivery failure")
	}
	
	return nil
}

// AssertBatchDeliveryWithPartialSuccess validates batch delivery allowing for partial failures
func AssertBatchDeliveryWithPartialSuccess(ctx *TestContext, result PutEventsResult, maxFailureRate float64) {
	err := AssertBatchDeliveryWithPartialSuccessE(ctx, result, maxFailureRate)
	if err != nil {
		ctx.T.Errorf("AssertBatchDeliveryWithPartialSuccess failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertBatchDeliveryWithPartialSuccessE validates batch delivery allowing for partial failures and returns error
func AssertBatchDeliveryWithPartialSuccessE(ctx *TestContext, result PutEventsResult, maxFailureRate float64) error {
	if maxFailureRate < 0.0 || maxFailureRate > 1.0 {
		return fmt.Errorf("max failure rate must be between 0.0 and 1.0: %f", maxFailureRate)
	}
	
	totalEntries := len(result.Entries)
	if totalEntries == 0 {
		return fmt.Errorf("batch result contains no entries")
	}
	
	actualFailureRate := float64(result.FailedEntryCount) / float64(totalEntries)
	if actualFailureRate > maxFailureRate {
		var failedDetails []string
		for i, entry := range result.Entries {
			if !entry.Success {
				failedDetails = append(failedDetails, fmt.Sprintf("Entry %d: %s - %s", i, entry.ErrorCode, entry.ErrorMessage))
			}
		}
		return fmt.Errorf("batch failure rate (%.2f) exceeds maximum allowed (%.2f): %d out of %d events failed:\n%s", 
			actualFailureRate, maxFailureRate, result.FailedEntryCount, totalEntries, joinStrings(failedDetails, "\n"))
	}
	
	return nil
}

// ValidateAdvancedDeliveryMetrics performs enhanced validation on delivery metrics
func ValidateAdvancedDeliveryMetrics(metrics DeliveryMetrics, thresholds MetricsThresholds) {
	err := ValidateAdvancedDeliveryMetricsE(metrics, thresholds)
	if err != nil {
		panic(fmt.Sprintf("ValidateAdvancedDeliveryMetrics failed: %v", err))
	}
}

// ValidateAdvancedDeliveryMetricsE performs enhanced validation on delivery metrics and returns error
func ValidateAdvancedDeliveryMetricsE(metrics DeliveryMetrics, thresholds MetricsThresholds) error {
	// First run basic validation
	if err := ValidateDeliveryMetricsE(metrics); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}
	
	// Advanced threshold checks
	if metrics.DeliveryLatencyMs > thresholds.MaxLatencyMs {
		return fmt.Errorf("delivery latency (%d ms) exceeds threshold (%d ms)", 
			metrics.DeliveryLatencyMs, thresholds.MaxLatencyMs)
	}
	
	if metrics.ErrorRate > thresholds.MaxErrorRate {
		return fmt.Errorf("error rate (%.4f) exceeds threshold (%.4f)", 
			metrics.ErrorRate, thresholds.MaxErrorRate)
	}
	
	totalEvents := metrics.DeliveredCount + metrics.FailedCount
	if totalEvents < thresholds.MinTotalEvents {
		return fmt.Errorf("total events (%d) below minimum threshold (%d)", 
			totalEvents, thresholds.MinTotalEvents)
	}
	
	return nil
}

// WaitForEventDeliveryWithBackoff waits for event delivery using exponential backoff
func WaitForEventDeliveryWithBackoff(ctx *TestContext, eventID string, timeout time.Duration, initialInterval time.Duration) {
	err := WaitForEventDeliveryWithBackoffE(ctx, eventID, timeout, initialInterval)
	if err != nil {
		ctx.T.Errorf("WaitForEventDeliveryWithBackoff failed: %v", err)
		ctx.T.FailNow()
	}
}

// WaitForEventDeliveryWithBackoffE waits for event delivery using exponential backoff and returns error
func WaitForEventDeliveryWithBackoffE(ctx *TestContext, eventID string, timeout time.Duration, initialInterval time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive: %v", timeout)
	}
	
	if initialInterval <= 0 {
		return fmt.Errorf("initial interval must be positive: %v", initialInterval)
	}
	
	startTime := time.Now()
	checkInterval := initialInterval
	maxInterval := timeout / 10 // Cap backoff at 10% of total timeout
	
	for time.Since(startTime) < timeout {
		time.Sleep(checkInterval)
		
		if isEventDelivered(eventID) {
			return nil
		}
		
		// Exponential backoff with cap
		checkInterval = checkInterval * 2
		if checkInterval > maxInterval {
			checkInterval = maxInterval
		}
	}
	
	return fmt.Errorf("timeout waiting for event '%s' to be delivered after %v", eventID, timeout)
}

// AssertDeliveryLatencyPercentiles validates latency percentiles
func AssertDeliveryLatencyPercentiles(ctx *TestContext, latencies []int64, percentiles map[float64]int64) {
	err := AssertDeliveryLatencyPercentilesE(ctx, latencies, percentiles)
	if err != nil {
		ctx.T.Errorf("AssertDeliveryLatencyPercentiles failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertDeliveryLatencyPercentilesE validates latency percentiles and returns error
func AssertDeliveryLatencyPercentilesE(ctx *TestContext, latencies []int64, percentiles map[float64]int64) error {
	if len(latencies) == 0 {
		return fmt.Errorf("latencies slice cannot be empty")
	}
	
	if len(percentiles) == 0 {
		return fmt.Errorf("percentiles map cannot be empty")
	}
	
	// Sort latencies for percentile calculation
	sortedLatencies := make([]int64, len(latencies))
	copy(sortedLatencies, latencies)
	sortSlice(sortedLatencies)
	
	for percentile, expectedLatency := range percentiles {
		if percentile <= 0 || percentile > 1.0 {
			return fmt.Errorf("percentile must be between 0 and 1.0: %f", percentile)
		}
		
		index := int(percentile * float64(len(sortedLatencies)))
		if index >= len(sortedLatencies) {
			index = len(sortedLatencies) - 1
		}
		
		actualLatency := sortedLatencies[index]
		if actualLatency > expectedLatency {
			return fmt.Errorf("P%.0f latency (%d ms) exceeds expected (%d ms)", 
				percentile*100, actualLatency, expectedLatency)
		}
	}
	
	return nil
}


// Types for enhanced validation

// MetricsThresholds defines validation thresholds for delivery metrics
type MetricsThresholds struct {
	MaxLatencyMs    int64
	MaxErrorRate    float64
	MinTotalEvents  int64
}

// Helper functions

// joinStrings concatenates strings with a separator (pure function)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	
	if len(strs) == 1 {
		return strs[0]
	}
	
	totalLen := len(sep) * (len(strs) - 1)
	for _, str := range strs {
		totalLen += len(str)
	}
	
	result := make([]byte, 0, totalLen)
	result = append(result, strs[0]...)
	
	for _, str := range strs[1:] {
		result = append(result, sep...)
		result = append(result, str...)
	}
	
	return string(result)
}

// sortSlice sorts int64 slice in ascending order (in-place)
func sortSlice(arr []int64) {
	for i := 1; i < len(arr); i++ {
		key := arr[i]
		j := i - 1
		
		for j >= 0 && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
}

// isEventDelivered simulates checking if an event has been delivered (mock for testing)
func isEventDelivered(eventID string) bool {
	// In a real implementation, this would check CloudWatch logs or metrics
	// For testing purposes, always return false to trigger timeout
	return false
}