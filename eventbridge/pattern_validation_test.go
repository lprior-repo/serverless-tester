package eventbridge

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventPatternMatching(t *testing.T) {
	t.Run("MatchBasicPattern", func(t *testing.T) {
		pattern := `{"source": ["myapp.orders"]}`
		event := map[string]interface{}{
			"source": "myapp.orders",
			"detail-type": "Order Placed",
		}
		
		matches, err := TestEventPatternE(pattern, event)
		require.NoError(t, err)
		assert.True(t, matches, "Event should match pattern")
	})

	t.Run("MatchArrayPattern", func(t *testing.T) {
		pattern := `{"source": ["myapp.orders", "myapp.customers"]}`
		event := map[string]interface{}{
			"source": "myapp.customers",
			"detail-type": "Customer Created",
		}
		
		matches, err := TestEventPatternE(pattern, event)
		require.NoError(t, err)
		assert.True(t, matches, "Event should match one of the array values")
	})

	t.Run("MatchNestedPattern", func(t *testing.T) {
		pattern := `{"detail": {"status": ["completed", "failed"]}}`
		event := map[string]interface{}{
			"source": "myapp.orders",
			"detail": map[string]interface{}{
				"orderId": "12345",
				"status": "completed",
			},
		}
		
		matches, err := TestEventPatternE(pattern, event)
		require.NoError(t, err)
		assert.True(t, matches, "Event should match nested pattern")
	})

	t.Run("NoMatchDifferentSource", func(t *testing.T) {
		pattern := `{"source": ["myapp.orders"]}`
		event := map[string]interface{}{
			"source": "myapp.inventory",
			"detail-type": "Stock Updated",
		}
		
		matches, err := TestEventPatternE(pattern, event)
		require.NoError(t, err)
		assert.False(t, matches, "Event should not match pattern")
	})

	t.Run("NoMatchMissingField", func(t *testing.T) {
		pattern := `{"detail": {"status": ["completed"]}}`
		event := map[string]interface{}{
			"source": "myapp.orders",
			"detail": map[string]interface{}{
				"orderId": "12345",
			},
		}
		
		matches, err := TestEventPatternE(pattern, event)
		require.NoError(t, err)
		assert.False(t, matches, "Event should not match pattern when field is missing")
	})

	t.Run("InvalidPatternJSON", func(t *testing.T) {
		pattern := `{"invalid": json}`
		event := map[string]interface{}{}
		
		_, err := TestEventPatternE(pattern, event)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pattern JSON")
	})
}

func TestEventPatternValidationExtended(t *testing.T) {
	t.Run("ValidPatterns", func(t *testing.T) {
		validPatterns := []string{
			`{"source": ["myapp"]}`,
			`{"detail-type": ["Order Placed"]}`,
			`{"detail": {"status": ["completed", "pending"]}}`,
			`{"source": ["myapp"], "detail-type": ["Order Placed"]}`,
			`{"detail": {"order": {"amount": [{"numeric": [">", 100]}]}}}`,
			`{"detail": {"timestamp": [{"exists": true}]}}`,
		}

		for _, pattern := range validPatterns {
			t.Run(pattern, func(t *testing.T) {
				err := ValidateEventPatternE(pattern)
				assert.NoError(t, err, "Pattern should be valid: %s", pattern)
			})
		}
	})

	t.Run("InvalidPatterns", func(t *testing.T) {
		invalidPatterns := []struct {
			pattern     string
			description string
		}{
			{"", "empty pattern"},
			{`invalid json`, "invalid JSON"},
			{`{"source": "string"}`, "non-array value"},
			{`{"detail": "not-object"}`, "non-object detail"},
		}

		for _, tc := range invalidPatterns {
			t.Run(tc.description, func(t *testing.T) {
				err := ValidateEventPatternE(tc.pattern)
				require.Error(t, err, "Pattern should be invalid: %s", tc.pattern)
			})
		}
	})
}

func TestBuildEventPattern(t *testing.T) {
	t.Run("SimpleSourcePattern", func(t *testing.T) {
		pattern := BuildEventPattern(map[string]interface{}{
			"source": []string{"myapp.orders"},
		})
		
		expected := `{"source":["myapp.orders"]}`
		assert.JSONEq(t, expected, pattern)
	})

	t.Run("MultiFieldPattern", func(t *testing.T) {
		pattern := BuildEventPattern(map[string]interface{}{
			"source": []string{"myapp.orders"},
			"detail-type": []string{"Order Placed", "Order Updated"},
		})
		
		var result map[string]interface{}
		err := json.Unmarshal([]byte(pattern), &result)
		require.NoError(t, err)
		
		assert.Equal(t, []interface{}{"myapp.orders"}, result["source"])
		assert.Equal(t, []interface{}{"Order Placed", "Order Updated"}, result["detail-type"])
	})

	t.Run("NestedPattern", func(t *testing.T) {
		pattern := BuildEventPattern(map[string]interface{}{
			"detail": map[string]interface{}{
				"status": []string{"completed", "failed"},
			},
		})
		
		var result map[string]interface{}
		err := json.Unmarshal([]byte(pattern), &result)
		require.NoError(t, err)
		
		detail := result["detail"].(map[string]interface{})
		assert.Equal(t, []interface{}{"completed", "failed"}, detail["status"])
	})
}

func TestAdvancedEventPatterns(t *testing.T) {
	t.Run("NumericPattern", func(t *testing.T) {
		pattern := `{"detail": {"amount": [{"numeric": [">", 100]}]}}`
		event := map[string]interface{}{
			"detail": map[string]interface{}{
				"amount": 150,
			},
		}
		
		matches, err := TestEventPatternE(pattern, event)
		require.NoError(t, err)
		assert.True(t, matches, "Event should match numeric pattern")
	})

	t.Run("ExistsPattern", func(t *testing.T) {
		pattern := `{"detail": {"timestamp": [{"exists": true}]}}`
		event := map[string]interface{}{
			"detail": map[string]interface{}{
				"timestamp": "2023-01-01T00:00:00Z",
				"orderId": "12345",
			},
		}
		
		matches, err := TestEventPatternE(pattern, event)
		require.NoError(t, err)
		assert.True(t, matches, "Event should match exists pattern")
	})

	t.Run("PrefixPattern", func(t *testing.T) {
		pattern := `{"detail": {"orderId": [{"prefix": "ORDER-"}]}}`
		event := map[string]interface{}{
			"detail": map[string]interface{}{
				"orderId": "ORDER-12345",
			},
		}
		
		matches, err := TestEventPatternE(pattern, event)
		require.NoError(t, err)
		assert.True(t, matches, "Event should match prefix pattern")
	})
}