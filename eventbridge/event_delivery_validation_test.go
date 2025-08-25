package eventbridge

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestAssertEventDeliverySuccess tests event delivery success assertion
func TestAssertEventDeliverySuccess(t *testing.T) {
	ctx := &TestContext{T: t}

	t.Run("should pass for successful delivery", func(t *testing.T) {
		result := PutEventResult{
			EventID:      "event-123",
			Success:      true,
			ErrorCode:    "",
			ErrorMessage: "",
		}

		// Should not panic
		assert.NotPanics(t, func() {
			AssertEventDeliverySuccess(ctx, result)
		})
	})

	t.Run("should fail for unsuccessful delivery", func(t *testing.T) {
		result := PutEventResult{
			EventID:      "",
			Success:      false,
			ErrorCode:    "ValidationException",
			ErrorMessage: "Event detail is invalid",
		}

		mockT := &MockTestingT{}
		// Use mock.AnythingOfType to match any error message format
		mockT.On("Errorf", "AssertEventDeliverySuccess failed: %v", mock.AnythingOfType("[]interface {}")).Once()
		mockT.On("FailNow").Once()
		
		ctx.T = mockT

		// Should panic and record error
		assert.Panics(t, func() {
			AssertEventDeliverySuccess(ctx, result)
		})
		
		mockT.AssertExpectations(t)
	})
}

// TestAssertBatchDeliverySuccess tests batch event delivery success assertion
func TestAssertBatchDeliverySuccess(t *testing.T) {
	ctx := &TestContext{T: t}

	t.Run("should pass for all successful deliveries", func(t *testing.T) {
		result := PutEventsResult{
			FailedEntryCount: 0,
			Entries: []PutEventResult{
				{EventID: "event-1", Success: true},
				{EventID: "event-2", Success: true},
			},
		}

		assert.NotPanics(t, func() {
			AssertBatchDeliverySuccess(ctx, result)
		})
	})

	t.Run("should fail with partial failures", func(t *testing.T) {
		result := PutEventsResult{
			FailedEntryCount: 1,
			Entries: []PutEventResult{
				{EventID: "event-1", Success: true},
				{EventID: "", Success: false, ErrorCode: "ValidationException"},
			},
		}

		mockT := &MockTestingT{}
		mockT.On("Errorf", "AssertBatchDeliverySuccess failed: %v", []interface{}{"batch delivery failed: 1 out of 2 events failed:\nEntry 1: ValidationException - "}).Once()
		mockT.On("FailNow").Once()
		
		ctx.T = mockT

		assert.Panics(t, func() {
			AssertBatchDeliverySuccess(ctx, result)
		})
		
		mockT.AssertExpectations(t)
	})
}

// TestAssertTargetDeliveryMonitoring tests target delivery monitoring assertions
func TestAssertTargetDeliveryMonitoring(t *testing.T) {
	t.Run("should validate delivery metrics structure", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID:           "target-1",
			DeliveredCount:     100,
			FailedCount:        2,
			DeliveryLatencyMs:  50,
			LastDeliveryTime:   time.Now(),
			ErrorRate:          0.02,
		}

		err := ValidateDeliveryMetricsE(metrics)
		assert.NoError(t, err)
	})

	t.Run("should reject invalid metrics", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID:           "",
			DeliveredCount:     -1,
			FailedCount:        -1,
			DeliveryLatencyMs:  -10,
			ErrorRate:          1.5, // Invalid: > 1.0
		}

		err := ValidateDeliveryMetricsE(metrics)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target ID cannot be empty")
	})
}

// TestMonitorEventDelivery tests event delivery monitoring functionality
func TestMonitorEventDelivery(t *testing.T) {
	ctx := &TestContext{T: t}
	
	t.Run("should track delivery success rate", func(t *testing.T) {
		ruleName := "test-rule"
		eventBusName := "test-bus"
		duration := 5 * time.Minute

		// Should not panic when monitoring is properly configured
		assert.NotPanics(t, func() {
			metrics := MonitorEventDelivery(ctx, ruleName, eventBusName, duration)
			assert.NotEmpty(t, metrics.TargetID)
			assert.GreaterOrEqual(t, metrics.DeliveredCount, int64(0))
		})
	})
}

// TestWaitForEventDelivery tests waiting for event delivery completion
func TestWaitForEventDelivery(t *testing.T) {
	ctx := &TestContext{T: t}

	t.Run("should timeout appropriately", func(t *testing.T) {
		eventID := "event-123"
		timeout := 1 * time.Second

		start := time.Now()
		
		mockT := &MockTestingT{}
		mockT.On("Errorf", "WaitForEventDelivery failed: %v", []interface{}{"timeout waiting for event 'event-123' to be delivered after 1s"}).Once()
		mockT.On("FailNow").Once()
		
		ctx.T = mockT

		// This should timeout quickly for testing
		assert.Panics(t, func() {
			WaitForEventDelivery(ctx, eventID, timeout)
		})

		elapsed := time.Since(start)
		assert.True(t, elapsed >= timeout, "Should wait at least the timeout duration")
		assert.True(t, elapsed < timeout*2, "Should not wait much longer than timeout")
		
		mockT.AssertExpectations(t)
	})
}

// TestAssertDeliveryLatency tests delivery latency assertions
func TestAssertDeliveryLatency(t *testing.T) {
	ctx := &TestContext{T: t}

	t.Run("should pass for acceptable latency", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID:          "target-1",
			DeliveryLatencyMs: 100, // 100ms
		}
		maxLatencyMs := int64(500) // 500ms limit

		assert.NotPanics(t, func() {
			AssertDeliveryLatency(ctx, metrics, maxLatencyMs)
		})
	})

	t.Run("should fail for excessive latency", func(t *testing.T) {
		metrics := DeliveryMetrics{
			TargetID:          "target-1",
			DeliveryLatencyMs: 1000, // 1000ms
		}
		maxLatencyMs := int64(500) // 500ms limit

		mockT := &MockTestingT{}
		mockT.On("Errorf", "AssertDeliveryLatency failed: %v", []interface{}{"delivery latency (1000 ms) exceeds maximum allowed (500 ms) for target target-1"}).Once()
		mockT.On("FailNow").Once()
		
		ctx.T = mockT

		assert.Panics(t, func() {
			AssertDeliveryLatency(ctx, metrics, maxLatencyMs)
		})
		
		mockT.AssertExpectations(t)
	})
}

// Use existing MockTestingT from mocks_test.go