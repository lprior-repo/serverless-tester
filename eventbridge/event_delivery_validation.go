package eventbridge

import (
	"fmt"
	"time"
)

// DeliveryMetrics represents event delivery metrics for monitoring
type DeliveryMetrics struct {
	TargetID           string
	DeliveredCount     int64
	FailedCount        int64
	DeliveryLatencyMs  int64
	LastDeliveryTime   time.Time
	ErrorRate          float64
}

// AssertEventDeliverySuccess asserts that an event was successfully delivered and panics if not (Terratest pattern)
func AssertEventDeliverySuccess(ctx *TestContext, result PutEventResult) {
	err := AssertEventDeliverySuccessE(ctx, result)
	if err != nil {
		ctx.T.Errorf("AssertEventDeliverySuccess failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertEventDeliverySuccessE asserts that an event was successfully delivered and returns error
func AssertEventDeliverySuccessE(ctx *TestContext, result PutEventResult) error {
	if !result.Success {
		return fmt.Errorf("event delivery failed: %s - %s", result.ErrorCode, result.ErrorMessage)
	}
	
	if result.EventID == "" {
		return fmt.Errorf("event ID is empty, indicating delivery failure")
	}
	
	return nil
}

// AssertBatchDeliverySuccess asserts that all events in a batch were successfully delivered and panics if not (Terratest pattern)
func AssertBatchDeliverySuccess(ctx *TestContext, result PutEventsResult) {
	err := AssertBatchDeliverySuccessE(ctx, result)
	if err != nil {
		ctx.T.Errorf("AssertBatchDeliverySuccess failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertBatchDeliverySuccessE asserts that all events in a batch were successfully delivered and returns error
func AssertBatchDeliverySuccessE(ctx *TestContext, result PutEventsResult) error {
	if result.FailedEntryCount > 0 {
		var failedDetails []string
		for i, entry := range result.Entries {
			if !entry.Success {
				failedDetails = append(failedDetails, fmt.Sprintf("Entry %d: %s - %s", i, entry.ErrorCode, entry.ErrorMessage))
			}
		}
		return fmt.Errorf("batch delivery failed: %d out of %d events failed:\n%s", 
			result.FailedEntryCount, len(result.Entries), joinStrings(failedDetails, "\n"))
	}
	
	for i, entry := range result.Entries {
		if entry.EventID == "" {
			return fmt.Errorf("entry %d has empty event ID, indicating delivery failure", i)
		}
	}
	
	return nil
}

// ValidateDeliveryMetrics validates delivery metrics and panics on error (Terratest pattern)
func ValidateDeliveryMetrics(metrics DeliveryMetrics) {
	err := ValidateDeliveryMetricsE(metrics)
	if err != nil {
		panic(fmt.Sprintf("ValidateDeliveryMetrics failed: %v", err))
	}
}

// ValidateDeliveryMetricsE validates delivery metrics and returns error
func ValidateDeliveryMetricsE(metrics DeliveryMetrics) error {
	if metrics.TargetID == "" {
		return fmt.Errorf("target ID cannot be empty")
	}
	
	if metrics.DeliveredCount < 0 {
		return fmt.Errorf("delivered count cannot be negative: %d", metrics.DeliveredCount)
	}
	
	if metrics.FailedCount < 0 {
		return fmt.Errorf("failed count cannot be negative: %d", metrics.FailedCount)
	}
	
	if metrics.DeliveryLatencyMs < 0 {
		return fmt.Errorf("delivery latency cannot be negative: %d ms", metrics.DeliveryLatencyMs)
	}
	
	if metrics.ErrorRate < 0.0 || metrics.ErrorRate > 1.0 {
		return fmt.Errorf("error rate must be between 0.0 and 1.0: %f", metrics.ErrorRate)
	}
	
	// Validate that error rate matches counts if both are provided
	totalCount := metrics.DeliveredCount + metrics.FailedCount
	if totalCount > 0 {
		expectedErrorRate := float64(metrics.FailedCount) / float64(totalCount)
		tolerance := 0.001 // Allow small floating point differences
		if abs(metrics.ErrorRate-expectedErrorRate) > tolerance {
			return fmt.Errorf("error rate (%f) does not match calculated rate from counts (%f)", 
				metrics.ErrorRate, expectedErrorRate)
		}
	}
	
	return nil
}

// WaitForEventDelivery waits for an event to be delivered and panics on timeout (Terratest pattern)
func WaitForEventDelivery(ctx *TestContext, eventID string, timeout time.Duration) {
	err := WaitForEventDeliveryE(ctx, eventID, timeout)
	if err != nil {
		ctx.T.Errorf("WaitForEventDelivery failed: %v", err)
		ctx.T.FailNow()
	}
}

// WaitForEventDeliveryE waits for an event to be delivered and returns error on timeout
func WaitForEventDeliveryE(ctx *TestContext, eventID string, timeout time.Duration) error {
	startTime := time.Now()
	checkInterval := 100 * time.Millisecond
	
	for time.Since(startTime) < timeout {
		// In a real implementation, this would check CloudWatch metrics or logs
		// For testing purposes, we'll simulate a timeout
		time.Sleep(checkInterval)
		
		// Simulate checking delivery status
		// In practice, this would query CloudWatch metrics for the specific event
		if isEventDelivered(eventID) {
			return nil
		}
	}
	
	return fmt.Errorf("timeout waiting for event '%s' to be delivered after %v", eventID, timeout)
}

// AssertDeliveryLatency asserts that delivery latency is within acceptable bounds and panics if not (Terratest pattern)
func AssertDeliveryLatency(ctx *TestContext, metrics DeliveryMetrics, maxLatencyMs int64) {
	err := AssertDeliveryLatencyE(ctx, metrics, maxLatencyMs)
	if err != nil {
		ctx.T.Errorf("AssertDeliveryLatency failed: %v", err)
		ctx.T.FailNow()
	}
}

// AssertDeliveryLatencyE asserts that delivery latency is within acceptable bounds and returns error
func AssertDeliveryLatencyE(ctx *TestContext, metrics DeliveryMetrics, maxLatencyMs int64) error {
	if metrics.DeliveryLatencyMs > maxLatencyMs {
		return fmt.Errorf("delivery latency (%d ms) exceeds maximum allowed (%d ms) for target %s", 
			metrics.DeliveryLatencyMs, maxLatencyMs, metrics.TargetID)
	}
	
	if metrics.DeliveryLatencyMs < 0 {
		return fmt.Errorf("delivery latency cannot be negative: %d ms", metrics.DeliveryLatencyMs)
	}
	
	return nil
}

// MonitorEventDelivery monitors event delivery for a rule over a specified duration and panics on error (Terratest pattern)
func MonitorEventDelivery(ctx *TestContext, ruleName string, eventBusName string, duration time.Duration) DeliveryMetrics {
	metrics, err := MonitorEventDeliveryE(ctx, ruleName, eventBusName, duration)
	if err != nil {
		ctx.T.Errorf("MonitorEventDelivery failed: %v", err)
		ctx.T.FailNow()
	}
	return metrics
}

// MonitorEventDeliveryE monitors event delivery for a rule over a specified duration and returns error
func MonitorEventDeliveryE(ctx *TestContext, ruleName string, eventBusName string, duration time.Duration) (DeliveryMetrics, error) {
	if ruleName == "" {
		return DeliveryMetrics{}, fmt.Errorf("rule name cannot be empty")
	}
	
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	if duration <= 0 {
		return DeliveryMetrics{}, fmt.Errorf("monitoring duration must be positive: %v", duration)
	}
	
	// In a real implementation, this would collect CloudWatch metrics
	// For testing, return simulated metrics
	return DeliveryMetrics{
		TargetID:           ruleName + "-target",
		DeliveredCount:     100,
		FailedCount:        2,
		DeliveryLatencyMs:  45,
		LastDeliveryTime:   time.Now(),
		ErrorRate:          0.02,
	}, nil
}

// Helper functions for implementation

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

// abs returns the absolute value of a float64 (pure function)
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// isEventDelivered simulates checking if an event has been delivered (mock for testing)
func isEventDelivered(eventID string) bool {
	// In a real implementation, this would check CloudWatch logs or metrics
	// For testing purposes, always return false to trigger timeout
	return false
}