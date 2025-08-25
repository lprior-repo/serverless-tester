package snapshot

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveCoverage tests all remaining methods to achieve 98%+ coverage
func TestComprehensiveCoverage(t *testing.T) {
	tempDir := t.TempDir()
	opts := Options{SnapshotDir: tempDir}
	s := New(t, opts)
	
	// Test all major methods to increase coverage
	
	// 1. Test MatchJSON with error path (try to marshal invalid data)
	t.Run("MatchJSON_marshal_error", func(t *testing.T) {
		// Create circular reference to cause marshal error
		type Node struct {
			Value string
			Next  *Node
		}
		node1 := &Node{Value: "node1"}
		node2 := &Node{Value: "node2"} 
		node1.Next = node2
		node2.Next = node1
		
		// This should cause JSON marshal error and hit error path
		defer func() {
			recover() // Ignore panic from test framework
		}()
		s.MatchJSON("json_error", node1)
	})
	
	// 2. Test MatchLambdaResponse with JSON payload
	t.Run("MatchLambdaResponse_json", func(t *testing.T) {
		response := map[string]interface{}{
			"statusCode": 200,
			"requestId":  "test-123",
			"body":       "success",
		}
		
		jsonBytes, _ := json.Marshal(response)
		s.MatchLambdaResponse("lambda_response", jsonBytes)
		
		// Verify sanitization worked
		content, err := os.ReadFile(filepath.Join(tempDir, "TestComprehensiveCoverage_lambda_response.snap"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "<REQUEST_ID>")
	})
	
	// 3. Test MatchLambdaResponse with non-JSON payload
	t.Run("MatchLambdaResponse_non_json", func(t *testing.T) {
		nonJsonPayload := []byte("plain text response")
		s.MatchLambdaResponse("lambda_plain", nonJsonPayload)
		
		content, err := os.ReadFile(filepath.Join(tempDir, "TestComprehensiveCoverage_lambda_plain.snap"))
		require.NoError(t, err)
		assert.Equal(t, "plain text response", string(content))
	})
	
	// 4. Test MatchDynamoDBItems
	t.Run("MatchDynamoDBItems", func(t *testing.T) {
		items := []map[string]interface{}{
			{"id": "1", "name": "test1", "timestamp": "2023-01-01T00:00:00.000Z"},
			{"id": "2", "name": "test2"},
		}
		
		s.MatchDynamoDBItems("dynamo_items", items)
		
		content, err := os.ReadFile(filepath.Join(tempDir, "TestComprehensiveCoverage_dynamo_items.snap"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "<TIMESTAMP>")
	})
	
	// 5. Test MatchStepFunctionExecution
	t.Run("MatchStepFunctionExecution", func(t *testing.T) {
		execution := map[string]interface{}{
			"executionArn": "arn:aws:states:us-east-1:123456789012:execution:MyStateMachine:exec-123",
			"status":       "SUCCEEDED",
			"startDate":    "2023-01-01T12:00:00.000Z",
			"endDate":      "2023-01-01T12:05:00.000Z",
		}
		
		s.MatchStepFunctionExecution("step_execution", execution)
		
		content, err := os.ReadFile(filepath.Join(tempDir, "TestComprehensiveCoverage_step_execution.snap"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "<EXECUTION_ARN>")
		assert.Contains(t, string(content), "<TIMESTAMP>")
	})
	
	// 6. Test WithSanitizers functionality
	t.Run("WithSanitizers", func(t *testing.T) {
		baseSanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("old"), []byte("OLD"))
		}
		
		baseOpts := Options{
			SnapshotDir: tempDir,
			Sanitizers:  []Sanitizer{baseSanitizer},
		}
		baseSnapshot := New(t, baseOpts)
		
		additionalSanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("new"), []byte("NEW"))
		}
		
		combined := baseSnapshot.WithSanitizers(additionalSanitizer)
		combined.Match("with_sanitizers", "old and new content")
		
		content, err := os.ReadFile(filepath.Join(tempDir, "TestComprehensiveCoverage_with_sanitizers.snap"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "OLD")
		assert.Contains(t, string(content), "NEW")
	})
	
	// 7. Test all toBytes variations
	t.Run("toBytes_variations", func(t *testing.T) {
		// Test string conversion
		result, err := s.toBytes("test string")
		assert.NoError(t, err)
		assert.Equal(t, []byte("test string"), result)
		
		// Test byte slice
		result, err = s.toBytes([]byte("test bytes"))
		assert.NoError(t, err)
		assert.Equal(t, []byte("test bytes"), result)
		
		// Test struct with JSON indent
		data := map[string]interface{}{"key": "value", "number": 42}
		result, err = s.toBytes(data)
		assert.NoError(t, err)
		assert.Contains(t, string(result), "{\n  \"key\"")
		
		// Test with JSONIndent false
		s.options.JSONIndent = false
		result, err = s.toBytes(data)
		assert.NoError(t, err)
		assert.NotContains(t, string(result), "\n  ")
	})
}

// TestUpdateSnapshotMode tests update functionality 
func TestUpdateSnapshotMode(t *testing.T) {
	tempDir := t.TempDir()
	
	opts := Options{
		SnapshotDir:     tempDir,
		UpdateSnapshots: true,
	}
	s := New(t, opts)
	
	// Test that update mode creates/updates snapshots
	s.Match("update_test", "initial content")
	
	content, err := os.ReadFile(filepath.Join(tempDir, "TestUpdateSnapshotMode_update_test.snap"))
	require.NoError(t, err)
	assert.Equal(t, "initial content", string(content))
	
	// Update with different content
	s.Match("update_test", "updated content")
	
	content, err = os.ReadFile(filepath.Join(tempDir, "TestUpdateSnapshotMode_update_test.snap"))
	require.NoError(t, err)
	assert.Equal(t, "updated content", string(content))
}

// TestEnvironmentVariables tests environment variable handling
func TestEnvironmentVariables(t *testing.T) {
	// Test UPDATE_SNAPSHOTS="1"
	t.Run("UPDATE_SNAPSHOTS_1", func(t *testing.T) {
		os.Setenv("UPDATE_SNAPSHOTS", "1")
		defer os.Unsetenv("UPDATE_SNAPSHOTS")
		
		s := New(t)
		assert.True(t, s.options.UpdateSnapshots)
	})
	
	// Test UPDATE_SNAPSHOTS="true"
	t.Run("UPDATE_SNAPSHOTS_true", func(t *testing.T) {
		os.Setenv("UPDATE_SNAPSHOTS", "true") 
		defer os.Unsetenv("UPDATE_SNAPSHOTS")
		
		s := New(t)
		assert.True(t, s.options.UpdateSnapshots)
	})
}

// TestSanitizerFunctions tests all sanitizer functions comprehensively
func TestSanitizerFunctions(t *testing.T) {
	// Test SanitizeTimestamps with various formats
	t.Run("SanitizeTimestamps_comprehensive", func(t *testing.T) {
		sanitizer := SanitizeTimestamps()
		
		// Test ISO timestamp
		input := []byte(`{"time":"2023-01-01T12:00:00.123Z"}`)
		result := sanitizer(input)
		assert.Contains(t, string(result), "<TIMESTAMP>")
		
		// Test Unix timestamp
		input = []byte(`{"timestamp": 1640995200000}`)
		result = sanitizer(input)
		assert.Contains(t, string(result), "<TIMESTAMP>")
		
		// Test createdAt field
		input = []byte(`{"createdAt": 1640995200}`)
		result = sanitizer(input)
		assert.Contains(t, string(result), "<TIMESTAMP>")
	})
	
	// Test all other sanitizers
	t.Run("all_sanitizers", func(t *testing.T) {
		// UUID sanitizer
		uuidSanitizer := SanitizeUUIDs()
		result := uuidSanitizer([]byte(`{"id":"550e8400-e29b-41d4-a716-446655440000"}`))
		assert.Contains(t, string(result), "<UUID>")
		
		// Lambda Request ID
		reqSanitizer := SanitizeLambdaRequestID()
		result = reqSanitizer([]byte(`{"requestId":"abc-123","RequestId":"def-456"}`))
		assert.Contains(t, string(result), "<REQUEST_ID>")
		
		// Execution ARN
		arnSanitizer := SanitizeExecutionArn()
		result = arnSanitizer([]byte(`{"arn":"arn:aws:states:us-east-1:123:execution:test:exec1"}`))
		assert.Contains(t, string(result), "<EXECUTION_ARN>")
		
		// DynamoDB Metadata
		dbSanitizer := SanitizeDynamoDBMetadata()
		result = dbSanitizer([]byte(`{"Items":[],"Count":5,"ScannedCount":10,"ConsumedCapacity":{"units":1}}`))
		assert.NotContains(t, string(result), "Count")
		assert.NotContains(t, string(result), "ScannedCount")
		assert.NotContains(t, string(result), "ConsumedCapacity")
		
		// Custom sanitizer
		customSanitizer := SanitizeCustom(`"password":"[^"]*"`, `"password":"REDACTED"`)
		result = customSanitizer([]byte(`{"password":"secret123"}`))
		assert.Contains(t, string(result), "REDACTED")
		
		// Field sanitizer
		fieldSanitizer := SanitizeField("apiKey", "MASKED")
		result = fieldSanitizer([]byte(`{"apiKey":"key123","other":"value"}`))
		assert.Contains(t, string(result), "MASKED")
		assert.Contains(t, string(result), "value")
	})
}

// TestNameSanitization tests name sanitization functions
func TestNameSanitization(t *testing.T) {
	// Test sanitizeTestName
	tests := []struct {
		input    string
		expected string
	}{
		{"TestExample", "Example"},
		{"TestExample/SubTest", "Example_SubTest"},
		{"TestExample With Spaces", "Example_With_Spaces"}, 
		{"NoTestPrefix", "NoTestPrefix"},
		{"TestMultiple/Levels/Deep", "Multiple_Levels_Deep"},
	}
	
	for _, test := range tests {
		result := sanitizeTestName(test.input)
		assert.Equal(t, test.expected, result, "sanitizeTestName(%q)", test.input)
	}
	
	// Test sanitizeSnapshotName
	snapshotTests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with spaces", "with_spaces"},
		{"with!special@chars#", "with_special_chars_"},
		{"test-with_underscores", "test-with_underscores"},
		{"complex/path?query=1&test=2", "complex_path_query_1_test_2"},
	}
	
	for _, test := range snapshotTests {
		result := sanitizeSnapshotName(test.input)
		assert.Equal(t, test.expected, result, "sanitizeSnapshotName(%q)", test.input)
	}
}