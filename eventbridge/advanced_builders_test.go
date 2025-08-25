package eventbridge

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvancedEventBuilders(t *testing.T) {
	t.Run("BuildSourcePattern", func(t *testing.T) {
		pattern := BuildSourcePattern("myapp.orders", "myapp.inventory")
		expected := `{"source":["myapp.orders","myapp.inventory"]}`
		assert.JSONEq(t, expected, pattern)
	})

	t.Run("BuildDetailTypePattern", func(t *testing.T) {
		pattern := BuildDetailTypePattern("Order Placed", "Order Updated")
		expected := `{"detail-type":["Order Placed","Order Updated"]}`
		assert.JSONEq(t, expected, pattern)
	})

	t.Run("BuildNumericPattern", func(t *testing.T) {
		pattern := BuildNumericPattern("detail.amount", ">", 100)
		expected := `{"detail":{"amount":[{"numeric":[">",100]}]}}`
		assert.JSONEq(t, expected, pattern)
	})

	t.Run("BuildExistsPattern", func(t *testing.T) {
		pattern := BuildExistsPattern("detail.orderId", true)
		expected := `{"detail":{"orderId":[{"exists":true}]}}`
		assert.JSONEq(t, expected, pattern)
	})

	t.Run("BuildPrefixPattern", func(t *testing.T) {
		pattern := BuildPrefixPattern("detail.orderId", "ORDER-")
		expected := `{"detail":{"orderId":[{"prefix":"ORDER-"}]}}`
		assert.JSONEq(t, expected, pattern)
	})

	t.Run("BuildComplexPattern", func(t *testing.T) {
		builder := NewPatternBuilder()
		builder.AddSource("myapp.orders")
		builder.AddDetailType("Order Placed", "Order Updated")
		builder.AddNumericCondition("detail.amount", ">", 100)
		builder.AddExistsCondition("detail.customerId", true)
		
		pattern := builder.Build()
		
		var result map[string]interface{}
		err := json.Unmarshal([]byte(pattern), &result)
		require.NoError(t, err)
		
		assert.Equal(t, []interface{}{"myapp.orders"}, result["source"])
		assert.Equal(t, []interface{}{"Order Placed", "Order Updated"}, result["detail-type"])
		
		detail := result["detail"].(map[string]interface{})
		assert.Contains(t, detail, "amount")
		assert.Contains(t, detail, "customerId")
	})
}

func TestEventSizeOptimization(t *testing.T) {
	t.Run("OptimizeEventSize", func(t *testing.T) {
		// Event with extra whitespace and formatting
		event := CustomEvent{
			Source:     "myapp.orders",
			DetailType: "Order Placed",
			Detail:     `{  "orderId":   "12345",  "customerName": "John Doe",   "description":  "A detailed description",  "metadata":   {  "createdAt": "2023-01-01T00:00:00Z", "processedBy":  "automated-system"  }  }`,
		}
		
		optimized, err := OptimizeEventSizeE(event)
		require.NoError(t, err)
		
		originalSize := len(event.Detail)
		optimizedSize := len(optimized.Detail)
		
		assert.LessOrEqual(t, optimizedSize, originalSize, "Optimized event should be same or smaller size")
		
		// Ensure it's still valid JSON
		var detail map[string]interface{}
		err = json.Unmarshal([]byte(optimized.Detail), &detail)
		require.NoError(t, err)
		
		// Verify no extra whitespace
		assert.NotContains(t, optimized.Detail, "  ")
	})

	t.Run("CalculateEventSize", func(t *testing.T) {
		event := CustomEvent{
			Source:     "myapp.orders",
			DetailType: "Order Placed",
			Detail:     `{"orderId":"12345","amount":99.99}`,
		}
		
		size := CalculateEventSize(event)
		assert.Greater(t, size, 0)
		assert.Less(t, size, 1000) // Reasonable size
	})

	t.Run("ValidateEventSize", func(t *testing.T) {
		// Small event - should pass
		smallEvent := CustomEvent{
			Source:     "test",
			DetailType: "Test",
			Detail:     `{"id":"123"}`,
		}
		
		err := ValidateEventSizeE(smallEvent)
		assert.NoError(t, err)
		
		// Large event - should fail
		largeDetail := make([]byte, 300000) // 300KB
		for i := range largeDetail {
			largeDetail[i] = 'a'
		}
		
		largeEvent := CustomEvent{
			Source:     "test",
			DetailType: "Test",
			Detail:     string(largeDetail),
		}
		
		err = ValidateEventSizeE(largeEvent)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event size too large")
	})
}

func TestBatchEventOperations(t *testing.T) {
	t.Run("CreateEventBatch", func(t *testing.T) {
		events := []CustomEvent{
			{Source: "myapp", DetailType: "Event1", Detail: `{"id":"1"}`},
			{Source: "myapp", DetailType: "Event2", Detail: `{"id":"2"}`},
			{Source: "myapp", DetailType: "Event3", Detail: `{"id":"3"}`},
		}
		
		batches := CreateEventBatches(events, 2)
		require.Len(t, batches, 2)
		
		assert.Len(t, batches[0], 2)
		assert.Len(t, batches[1], 1)
	})

	t.Run("ValidateEventBatch", func(t *testing.T) {
		// Valid batch
		validBatch := []CustomEvent{
			{Source: "test", DetailType: "Test", Detail: `{"id":"1"}`},
			{Source: "test", DetailType: "Test", Detail: `{"id":"2"}`},
		}
		
		err := ValidateEventBatchE(validBatch)
		assert.NoError(t, err)
		
		// Empty batch
		err = ValidateEventBatchE([]CustomEvent{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "batch cannot be empty")
		
		// Batch too large
		largeBatch := make([]CustomEvent, 15)
		for i := range largeBatch {
			largeBatch[i] = CustomEvent{Source: "test", DetailType: "Test", Detail: `{"id":"` + string(rune(i)) + `"}`}
		}
		
		err = ValidateEventBatchE(largeBatch)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "batch size exceeds maximum")
	})
}

func TestEventPatternAnalyzer(t *testing.T) {
	t.Run("AnalyzePatternComplexity", func(t *testing.T) {
		simplePattern := `{"source":["myapp"]}`
		complexity := AnalyzePatternComplexity(simplePattern)
		assert.Equal(t, "low", complexity)
		
		complexPattern := `{"source":["app1","app2"],"detail":{"amount":[{"numeric":[">",100]}],"status":["pending","processing"]}}`
		complexity = AnalyzePatternComplexity(complexPattern)
		assert.Equal(t, "high", complexity)
	})

	t.Run("EstimatePatternMatchRate", func(t *testing.T) {
		broadPattern := `{"source":["myapp"]}`
		rate := EstimatePatternMatchRate(broadPattern)
		assert.Greater(t, rate, 0.7) // Broad patterns should have high match rate
		
		narrowPattern := `{"source":["myapp.orders"],"detail":{"status":["completed"],"amount":[{"numeric":[">",1000]}]}}`
		rate = EstimatePatternMatchRate(narrowPattern)
		assert.Less(t, rate, 0.7) // Narrow patterns should have lower match rate
	})
}