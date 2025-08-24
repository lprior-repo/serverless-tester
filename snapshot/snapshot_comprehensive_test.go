package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"vasdeference"
)

// Mock TestingT implementation for comprehensive testing
type mockTestingT struct {
	errors []string
	failed bool
	mutex  sync.Mutex
	name   string
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *mockTestingT) FailNow() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failed = true
	panic("FailNow called")
}

func (m *mockTestingT) Helper() {}

func (m *mockTestingT) Name() string {
	return m.name
}

func (m *mockTestingT) getErrors() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]string{}, m.errors...)
}

func (m *mockTestingT) isFailed() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.failed
}

func newMockTestingT(name string) *mockTestingT {
	return &mockTestingT{name: name}
}

func TestSnapshotCreation(t *testing.T) {
	t.Run("New creates snapshot with default options", func(t *testing.T) {
		snapshot := New(t)

		assert.NotNil(t, snapshot)
		assert.Equal(t, "testdata/snapshots", snapshot.options.SnapshotDir)
		assert.Equal(t, 3, snapshot.options.DiffContextLines)
		assert.True(t, snapshot.options.JSONIndent)
		assert.False(t, snapshot.options.UpdateSnapshots)
		assert.Len(t, snapshot.options.Sanitizers, 0)
	})

	t.Run("New creates snapshot with custom options", func(t *testing.T) {
		customSanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("test"), []byte("mock"))
		}

		opts := Options{
			SnapshotDir:      "custom/snapshots",
			UpdateSnapshots:  true,
			Sanitizers:       []Sanitizer{customSanitizer},
			DiffContextLines: 10,
			JSONIndent:       false,
		}

		snapshot := New(t, opts)

		assert.Equal(t, "custom/snapshots", snapshot.options.SnapshotDir)
		assert.True(t, snapshot.options.UpdateSnapshots)
		assert.Len(t, snapshot.options.Sanitizers, 1)
		assert.Equal(t, 10, snapshot.options.DiffContextLines)
		assert.False(t, snapshot.options.JSONIndent)
	})

	t.Run("New respects UPDATE_SNAPSHOTS environment variable", func(t *testing.T) {
		testCases := []struct {
			envValue string
			expected bool
		}{
			{"true", true},
			{"1", true},
			{"false", false},
			{"0", false},
			{"", false},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("env_%s", tc.envValue), func(t *testing.T) {
				if tc.envValue != "" {
					os.Setenv("UPDATE_SNAPSHOTS", tc.envValue)
					defer os.Unsetenv("UPDATE_SNAPSHOTS")
				}

				snapshot := New(t)
				assert.Equal(t, tc.expected, snapshot.options.UpdateSnapshots)
			})
		}
	})

	t.Run("New creates snapshot directory if it doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "nested", "snapshots")

		snapshot := New(t, Options{SnapshotDir: snapshotDir})

		assert.NotNil(t, snapshot)
		assert.DirExists(t, snapshotDir)
	})

	t.Run("New handles invalid snapshot directory gracefully", func(t *testing.T) {
		mockT := newMockTestingT("test_invalid_dir")
		
		// Try to create directory in a read-only location (will fail)
		invalidDir := "/root/readonly/snapshots"

		defer func() {
			recover() // Expected panic from FailNow
		}()

		New(mockT, Options{SnapshotDir: invalidDir})

		assert.True(t, mockT.isFailed())
	})

	t.Run("NewE returns error when directory creation fails", func(t *testing.T) {
		invalidDir := "/root/readonly/snapshots"

		_, err := NewE(t, Options{SnapshotDir: invalidDir})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create snapshot directory")
	})

	t.Run("NewE creates snapshot successfully with valid options", func(t *testing.T) {
		tempDir := t.TempDir()

		snapshot, err := NewE(t, Options{SnapshotDir: tempDir})
		require.NoError(t, err)
		assert.NotNil(t, snapshot)
	})
}

func TestSnapshotMatching(t *testing.T) {
	t.Run("Match creates new snapshot when none exists", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		data := "test snapshot data"
		snapshot.Match("new_test", data)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		assert.Equal(t, data, string(content))
	})

	t.Run("Match succeeds when data matches existing snapshot", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		data := "consistent test data"
		
		// Create initial snapshot
		snapshot.Match("consistent_test", data)
		
		// Verify it matches on subsequent calls
		snapshot.Match("consistent_test", data)
	})

	t.Run("Match fails when data doesn't match existing snapshot", func(t *testing.T) {
		tempDir := t.TempDir()
		mockT := newMockTestingT("mismatch_test")
		snapshot := New(mockT, Options{SnapshotDir: tempDir})

		// Create initial snapshot
		snapshot.Match("mismatch_test", "original data")

		// Try to match with different data
		snapshot.Match("mismatch_test", "modified data")

		errors := mockT.getErrors()
		assert.Len(t, errors, 1)
		assert.Contains(t, errors[0], "Snapshot mismatch")
		assert.Contains(t, errors[0], "UPDATE_SNAPSHOTS=1")
	})

	t.Run("Match updates snapshot when UPDATE_SNAPSHOTS is true", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{
			SnapshotDir:     tempDir,
			UpdateSnapshots: true,
		})

		// Create initial snapshot
		snapshot.Match("update_test", "original data")

		// Update with new data
		snapshot.Match("update_test", "updated data")

		// Verify file was updated
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		assert.Equal(t, "updated data", string(content))
	})

	t.Run("Match applies sanitizers before comparison", func(t *testing.T) {
		tempDir := t.TempDir()
		
		timestampSanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("2023-01-01"), []byte("<DATE>"))
		}

		snapshot := New(t, Options{
			SnapshotDir: tempDir,
			Sanitizers:  []Sanitizer{timestampSanitizer},
		})

		// Create snapshot with dynamic data
		snapshot.Match("sanitized_test", "Event occurred on 2023-01-01")

		// Verify sanitized data was stored
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		assert.Equal(t, "Event occurred on <DATE>", string(content))

		// New data with different date should match due to sanitization
		snapshot.Match("sanitized_test", "Event occurred on 2023-12-31")
	})

	t.Run("MatchE returns error instead of failing test", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		// Create initial snapshot
		err := snapshot.MatchE("error_test", "original")
		require.NoError(t, err)

		// Try mismatched data
		err = snapshot.MatchE("error_test", "modified")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "snapshot mismatch")
	})
}

func TestJSONMatching(t *testing.T) {
	t.Run("MatchJSON formats data as indented JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		data := map[string]interface{}{
			"name":    "test",
			"value":   42,
			"active":  true,
			"items":   []string{"a", "b", "c"},
			"nested":  map[string]string{"key": "value"},
		}

		snapshot.MatchJSON("json_test", data)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)

		// Verify it's valid JSON
		var parsed map[string]interface{}
		err = json.Unmarshal(content, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "test", parsed["name"])
		assert.Equal(t, float64(42), parsed["value"]) // JSON numbers are float64
		assert.Equal(t, true, parsed["active"])

		// Verify indentation
		assert.Contains(t, string(content), "  \"name\"")
	})

	t.Run("MatchJSON handles complex nested structures", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		complexData := map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": 1, "name": "Alice", "roles": []string{"admin", "user"}},
				{"id": 2, "name": "Bob", "roles": []string{"user"}},
			},
			"metadata": map[string]interface{}{
				"version": "1.0",
				"created": "2023-01-01T00:00:00Z",
				"config": map[string]interface{}{
					"debug": true,
					"timeout": 30,
				},
			},
		}

		snapshot.MatchJSON("complex_json_test", complexData)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(content, &parsed)
		require.NoError(t, err)

		users := parsed["users"].([]interface{})
		assert.Len(t, users, 2)
	})

	t.Run("MatchJSONE returns error for invalid JSON", func(t *testing.T) {
		snapshot := New(t, Options{SnapshotDir: t.TempDir()})

		// Try to match data that can't be marshaled to JSON
		invalidData := make(chan int) // Channels can't be marshaled

		err := snapshot.MatchJSONE("invalid_json_test", invalidData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal JSON")
	})

	t.Run("MatchJSONE succeeds for valid data", func(t *testing.T) {
		snapshot := New(t, Options{SnapshotDir: t.TempDir()})

		data := map[string]string{"key": "value"}
		err := snapshot.MatchJSONE("valid_json_test", data)
		assert.NoError(t, err)
	})
}

func TestAWSServiceIntegration(t *testing.T) {
	t.Run("MatchLambdaResponse handles JSON payload with sanitization", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		lambdaResponse := map[string]interface{}{
			"statusCode": 200,
			"body":       `{"message": "success"}`,
			"headers": map[string]string{
				"Content-Type": "application/json",
			},
			"requestId": "12345-abcde-67890",
			"timestamp": "2023-01-01T12:00:00.000Z",
		}

		payload, _ := json.Marshal(lambdaResponse)
		snapshot.MatchLambdaResponse("lambda_test", payload)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)

		// Verify sanitization occurred
		contentStr := string(content)
		assert.Contains(t, contentStr, "<REQUEST_ID>")
		assert.Contains(t, contentStr, "<TIMESTAMP>")
		assert.NotContains(t, contentStr, "12345-abcde-67890")
		assert.NotContains(t, contentStr, "2023-01-01T12:00:00.000Z")
	})

	t.Run("MatchLambdaResponse handles non-JSON payload", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		payload := []byte("plain text response")
		snapshot.MatchLambdaResponse("lambda_text_test", payload)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		assert.Equal(t, "plain text response", string(content))
	})

	t.Run("MatchLambdaResponse applies custom sanitizers", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		customSanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("secret-value"), []byte("<SECRET>"))
		}

		payload := []byte(`{"data": "secret-value", "public": "visible"}`)
		snapshot.MatchLambdaResponse("lambda_custom_test", payload, customSanitizer)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		assert.Contains(t, string(content), "<SECRET>")
		assert.NotContains(t, string(content), "secret-value")
	})

	t.Run("MatchDynamoDBItems formats items correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		items := []map[string]interface{}{
			{
				"id":   "1",
				"name": "Alice",
				"age":  30,
				"active": true,
			},
			{
				"id":   "2", 
				"name": "Bob",
				"age":  25,
				"active": false,
			},
		}

		snapshot.MatchDynamoDBItems("dynamodb_test", items)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)

		var parsed []map[string]interface{}
		err = json.Unmarshal(content, &parsed)
		require.NoError(t, err)
		assert.Len(t, parsed, 2)
	})

	t.Run("MatchStepFunctionExecution sanitizes execution-specific data", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		executionOutput := map[string]interface{}{
			"executionArn": "arn:aws:states:us-east-1:123456789012:execution:MyStateMachine:abc123",
			"status":       "SUCCEEDED",
			"startDate":    "2023-01-01T10:00:00Z",
			"endDate":      "2023-01-01T10:05:00Z",
			"output":       `{"result": "success"}`,
			"stateMachineArn": "arn:aws:states:us-east-1:123456789012:stateMachine:MyStateMachine",
			"name": "abc123",
		}

		snapshot.MatchStepFunctionExecution("stepfunctions_test", executionOutput)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "<EXECUTION_ARN>")
		assert.Contains(t, contentStr, "<TIMESTAMP>")
		assert.Contains(t, contentStr, "<UUID>")
	})
}

func TestSanitizerFunctionality(t *testing.T) {
	t.Run("SanitizeTimestamps handles various timestamp formats", func(t *testing.T) {
		sanitizer := SanitizeTimestamps()
		
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "ISO 8601 with Z",
				input:    `{"time": "2023-01-01T12:00:00Z"}`,
				expected: `{"time": "<TIMESTAMP>"}`,
			},
			{
				name:     "ISO 8601 with milliseconds",
				input:    `{"time": "2023-01-01T12:00:00.123Z"}`,
				expected: `{"time": "<TIMESTAMP>"}`,
			},
			{
				name:     "Unix timestamp field",
				input:    `{"timestamp": 1672574400}`,
				expected: `{"timestamp": "<TIMESTAMP>"}`,
			},
			{
				name:     "Multiple timestamp fields",
				input:    `{"createdAt": 1672574400, "updatedAt": 1672574500}`,
				expected: `{"createdAt": "<TIMESTAMP>", "updatedAt": "<TIMESTAMP>"}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := sanitizer([]byte(tc.input))
				assert.Equal(t, tc.expected, string(result))
			})
		}
	})

	t.Run("SanitizeUUIDs replaces various UUID formats", func(t *testing.T) {
		sanitizer := SanitizeUUIDs()
		
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "lowercase UUID",
				input:    `{"id": "550e8400-e29b-41d4-a716-446655440000"}`,
				expected: `{"id": "<UUID>"}`,
			},
			{
				name:     "uppercase UUID",
				input:    `{"id": "550E8400-E29B-41D4-A716-446655440000"}`,
				expected: `{"id": "<UUID>"}`,
			},
			{
				name:     "mixed case UUID",
				input:    `{"id": "550e8400-E29B-41d4-A716-446655440000"}`,
				expected: `{"id": "<UUID>"}`,
			},
			{
				name:     "multiple UUIDs",
				input:    `{"user": "550e8400-e29b-41d4-a716-446655440000", "session": "123e4567-e89b-12d3-a456-426614174000"}`,
				expected: `{"user": "<UUID>", "session": "<UUID>"}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := sanitizer([]byte(tc.input))
				assert.Equal(t, tc.expected, string(result))
			})
		}
	})

	t.Run("SanitizeLambdaRequestID handles request ID variations", func(t *testing.T) {
		sanitizer := SanitizeLambdaRequestID()
		
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "requestId field",
				input:    `{"requestId": "12345-abcde-67890"}`,
				expected: `{"requestId": "<REQUEST_ID>"}`,
			},
			{
				name:     "RequestId field",
				input:    `{"RequestId": "12345-abcde-67890"}`,
				expected: `{"RequestId": "<REQUEST_ID>"}`,
			},
			{
				name:     "multiple request ID fields",
				input:    `{"requestId": "abc123", "parentRequestId": "def456"}`,
				expected: `{"requestId": "<REQUEST_ID>", "parentRequestId": "def456"}`, // Only exact field names match
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := sanitizer([]byte(tc.input))
				assert.Equal(t, tc.expected, string(result))
			})
		}
	})

	t.Run("SanitizeExecutionArn replaces Step Functions ARNs", func(t *testing.T) {
		sanitizer := SanitizeExecutionArn()
		
		input := `{"executionArn": "arn:aws:states:us-east-1:123456789012:execution:MyStateMachine:test-execution"}`
		expected := `{"executionArn": "<EXECUTION_ARN>"}`
		
		result := sanitizer([]byte(input))
		assert.Equal(t, expected, string(result))
	})

	t.Run("SanitizeDynamoDBMetadata removes metadata fields", func(t *testing.T) {
		sanitizer := SanitizeDynamoDBMetadata()
		
		input := `{"Items": [{"id": "1"}], "ConsumedCapacity": 1.5, "ScannedCount": 10, "Count": 1}`
		result := sanitizer([]byte(input))
		
		resultStr := string(result)
		assert.NotContains(t, resultStr, "ConsumedCapacity")
		assert.NotContains(t, resultStr, "ScannedCount")
		assert.NotContains(t, resultStr, "Count")
		assert.Contains(t, resultStr, "Items")
	})

	t.Run("SanitizeCustom creates custom regex-based sanitizer", func(t *testing.T) {
		sanitizer := SanitizeCustom(`token-[a-zA-Z0-9]+`, "<TOKEN>")
		
		input := `{"auth": "token-abc123xyz", "data": "public"}`
		expected := `{"auth": "<TOKEN>", "data": "public"}`
		
		result := sanitizer([]byte(input))
		assert.Equal(t, expected, string(result))
	})

	t.Run("SanitizeField replaces specific field values", func(t *testing.T) {
		sanitizer := SanitizeField("password", "<REDACTED>")
		
		input := `{"username": "alice", "password": "secret123", "email": "alice@example.com"}`
		expected := `{"username": "alice", "password": "<REDACTED>", "email": "alice@example.com"}`
		
		result := sanitizer([]byte(input))
		assert.Equal(t, expected, string(result))
	})

	t.Run("chained sanitizers work together", func(t *testing.T) {
		snapshot := New(t, Options{
			SnapshotDir: t.TempDir(),
			Sanitizers: []Sanitizer{
				SanitizeTimestamps(),
				SanitizeUUIDs(),
				SanitizeLambdaRequestID(),
			},
		})

		complexData := map[string]interface{}{
			"id":        "550e8400-e29b-41d4-a716-446655440000",
			"timestamp": "2023-01-01T12:00:00Z",
			"requestId": "abc-123-def",
			"data":      "normal data",
		}

		payload, _ := json.Marshal(complexData)
		snapshot.Match("chained_test", payload)

		files, err := os.ReadDir(snapshot.options.SnapshotDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(snapshot.options.SnapshotDir, files[0].Name()))
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "<UUID>")
		assert.Contains(t, contentStr, "<TIMESTAMP>")
		assert.Contains(t, contentStr, "<REQUEST_ID>")
		assert.Contains(t, contentStr, "normal data")
	})
}

func TestWithSanitizers(t *testing.T) {
	t.Run("WithSanitizers creates new snapshot with combined sanitizers", func(t *testing.T) {
		existingSanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("old"), []byte("existing"))
		}

		originalSnapshot := New(t, Options{
			SnapshotDir: t.TempDir(),
			Sanitizers:  []Sanitizer{existingSanitizer},
		})

		newSanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("new"), []byte("added"))
		}

		combinedSnapshot := originalSnapshot.WithSanitizers(newSanitizer)

		// Verify original is unchanged
		assert.Len(t, originalSnapshot.options.Sanitizers, 1)
		
		// Verify combined has both sanitizers
		assert.Len(t, combinedSnapshot.options.Sanitizers, 2)
		
		// Verify they're different objects
		assert.NotEqual(t, &originalSnapshot.options.Sanitizers, &combinedSnapshot.options.Sanitizers)
	})

	t.Run("WithSanitizers works with multiple new sanitizers", func(t *testing.T) {
		originalSnapshot := New(t, Options{SnapshotDir: t.TempDir()})

		sanitizer1 := func(data []byte) []byte { return data }
		sanitizer2 := func(data []byte) []byte { return data }
		sanitizer3 := func(data []byte) []byte { return data }

		combinedSnapshot := originalSnapshot.WithSanitizers(sanitizer1, sanitizer2, sanitizer3)

		assert.Len(t, originalSnapshot.options.Sanitizers, 0)
		assert.Len(t, combinedSnapshot.options.Sanitizers, 3)
	})
}

func TestDiffGenerationAndComparison(t *testing.T) {
	t.Run("generateDiff creates readable diff output", func(t *testing.T) {
		snapshot := New(t, Options{
			SnapshotDir:      t.TempDir(),
			DiffContextLines: 2,
		})

		expected := []byte("line1\nline2\nline3\nline4\nline5")
		actual := []byte("line1\nmodified\nline3\nline4\nline5")

		diff := snapshot.generateDiff(expected, actual, "test.snap")

		assert.Contains(t, diff, "line2")
		assert.Contains(t, diff, "modified")
		assert.Contains(t, diff, "@@")
		assert.Contains(t, diff, "+modified")
		assert.Contains(t, diff, "-line2")
	})

	t.Run("generateDiff handles context lines configuration", func(t *testing.T) {
		testCases := []int{0, 1, 3, 5}

		for _, contextLines := range testCases {
			t.Run(fmt.Sprintf("context_%d", contextLines), func(t *testing.T) {
				snapshot := New(t, Options{
					SnapshotDir:      t.TempDir(),
					DiffContextLines: contextLines,
				})

				expected := []byte("line1\nline2\nline3\nline4\nline5\nline6\nline7")
				actual := []byte("line1\nline2\nmodified\nline4\nline5\nline6\nline7")

				diff := snapshot.generateDiff(expected, actual, "test.snap")
				assert.NotEmpty(t, diff)
			})
		}
	})

	t.Run("diff handles binary data gracefully", func(t *testing.T) {
		snapshot := New(t, Options{SnapshotDir: t.TempDir()})

		expected := []byte{0x00, 0x01, 0x02, 0xFF}
		actual := []byte{0x00, 0x01, 0x03, 0xFF}

		diff := snapshot.generateDiff(expected, actual, "binary.snap")
		assert.NotEmpty(t, diff)
	})
}

func TestFileSystemOperations(t *testing.T) {
	t.Run("snapshot directory creation with permissions", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "custom", "nested", "snapshots")

		snapshot := New(t, Options{SnapshotDir: snapshotDir})

		assert.DirExists(t, snapshotDir)
		
		// Verify directory has correct permissions
		info, err := os.Stat(snapshotDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
	})

	t.Run("snapshot file naming and sanitization", func(t *testing.T) {
		tempDir := t.TempDir()
		mockT := newMockTestingT("Test_Complex_Name/with/slashes and spaces")
		snapshot := New(mockT, Options{SnapshotDir: tempDir})

		snapshot.Match("test@name#with$special%chars!", "data")

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		filename := files[0].Name()
		assert.Contains(t, filename, "Complex_Name_with_slashes_and_spaces")
		assert.Contains(t, filename, "test_name_with_special_chars_")
		assert.HasSuffix(t, filename, ".snap")
		
		// Verify no invalid filename characters
		assert.NotContains(t, filename, "@")
		assert.NotContains(t, filename, "#")
		assert.NotContains(t, filename, "$")
		assert.NotContains(t, filename, "%")
		assert.NotContains(t, filename, "!")
		assert.NotContains(t, filename, "/")
	})

	t.Run("concurrent snapshot operations are safe", func(t *testing.T) {
		tempDir := t.TempDir()
		
		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				mockT := newMockTestingT(fmt.Sprintf("ConcurrentTest_%d", id))
				snapshot := New(mockT, Options{SnapshotDir: tempDir})
				
				data := fmt.Sprintf("data for goroutine %d", id)
				testName := fmt.Sprintf("concurrent_test_%d", id)
				
				snapshot.Match(testName, data)
			}(i)
		}

		wg.Wait()

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		assert.Len(t, files, numGoroutines)
	})

	t.Run("snapshot file permissions are correct", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		snapshot.Match("permission_test", "test data")

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		filepath := filepath.Join(tempDir, files[0].Name())
		info, err := os.Stat(filepath)
		require.NoError(t, err)
		
		assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
	})
}

func TestDataConversion(t *testing.T) {
	t.Run("toBytes handles different data types", func(t *testing.T) {
		snapshot := New(t, Options{JSONIndent: true})

		testCases := []struct {
			name     string
			input    interface{}
			validate func(t *testing.T, result []byte)
		}{
			{
				name:  "byte slice",
				input: []byte("hello world"),
				validate: func(t *testing.T, result []byte) {
					assert.Equal(t, []byte("hello world"), result)
				},
			},
			{
				name:  "string",
				input: "hello world",
				validate: func(t *testing.T, result []byte) {
					assert.Equal(t, []byte("hello world"), result)
				},
			},
			{
				name:  "strings.Reader",
				input: strings.NewReader("hello world"),
				validate: func(t *testing.T, result []byte) {
					assert.Equal(t, []byte("hello world"), result)
				},
			},
			{
				name:  "struct with indentation",
				input: map[string]string{"key": "value"},
				validate: func(t *testing.T, result []byte) {
					assert.True(t, json.Valid(result))
					assert.Contains(t, string(result), "  \"key\"")
				},
			},
			{
				name:  "complex struct",
				input: map[string]interface{}{
					"string": "value",
					"number": 42,
					"bool":   true,
					"array":  []int{1, 2, 3},
					"object": map[string]string{"nested": "value"},
				},
				validate: func(t *testing.T, result []byte) {
					assert.True(t, json.Valid(result))
					
					var parsed map[string]interface{}
					err := json.Unmarshal(result, &parsed)
					require.NoError(t, err)
					
					assert.Equal(t, "value", parsed["string"])
					assert.Equal(t, float64(42), parsed["number"])
					assert.Equal(t, true, parsed["bool"])
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := snapshot.toBytes(tc.input)
				require.NoError(t, err)
				tc.validate(t, result)
			})
		}
	})

	t.Run("toBytes with JSONIndent false", func(t *testing.T) {
		snapshot := New(t, Options{JSONIndent: false})

		data := map[string]string{"key": "value"}
		result, err := snapshot.toBytes(data)
		require.NoError(t, err)

		assert.True(t, json.Valid(result))
		assert.NotContains(t, string(result), "  \"key\"") // No indentation
	})

	t.Run("toBytes handles io.Reader correctly", func(t *testing.T) {
		snapshot := New(t)

		readers := []io.Reader{
			strings.NewReader("test data"),
			bytes.NewBufferString("buffer data"),
			bytes.NewReader([]byte("byte reader data")),
		}

		expectedData := []string{
			"test data",
			"buffer data", 
			"byte reader data",
		}

		for i, reader := range readers {
			result, err := snapshot.toBytes(reader)
			require.NoError(t, err)
			assert.Equal(t, []byte(expectedData[i]), result)
		}
	})
}

func TestGlobalFunctions(t *testing.T) {
	t.Run("Match function uses global instance", func(t *testing.T) {
		// Reset global snapshot
		originalSnapshot := defaultSnapshot
		defer func() { defaultSnapshot = originalSnapshot }()
		defaultSnapshot = nil

		mockT := newMockTestingT("GlobalTest")
		Match(mockT, "global_test", "test data")

		assert.NotNil(t, defaultSnapshot)
		assert.Equal(t, mockT, defaultSnapshot.t)
	})

	t.Run("MatchJSON function uses global instance", func(t *testing.T) {
		originalSnapshot := defaultSnapshot
		defer func() { defaultSnapshot = originalSnapshot }()
		defaultSnapshot = nil

		mockT := newMockTestingT("GlobalJSONTest") 
		data := map[string]string{"key": "value"}
		MatchJSON(mockT, "global_json_test", data)

		assert.NotNil(t, defaultSnapshot)
		assert.Equal(t, mockT, defaultSnapshot.t)
	})

	t.Run("global instance is reused for same TestingT", func(t *testing.T) {
		originalSnapshot := defaultSnapshot
		defer func() { defaultSnapshot = originalSnapshot }()
		defaultSnapshot = nil

		mockT := newMockTestingT("ReuseTest")
		
		Match(mockT, "test1", "data1")
		firstSnapshot := defaultSnapshot

		Match(mockT, "test2", "data2") 
		secondSnapshot := defaultSnapshot

		assert.Same(t, firstSnapshot, secondSnapshot)
	})

	t.Run("global instance is recreated for different TestingT", func(t *testing.T) {
		originalSnapshot := defaultSnapshot
		defer func() { defaultSnapshot = originalSnapshot }()
		defaultSnapshot = nil

		mockT1 := newMockTestingT("Test1")
		mockT2 := newMockTestingT("Test2")

		Match(mockT1, "test1", "data1")
		firstSnapshot := defaultSnapshot

		Match(mockT2, "test2", "data2")
		secondSnapshot := defaultSnapshot

		assert.NotSame(t, firstSnapshot, secondSnapshot)
		assert.Equal(t, mockT2, secondSnapshot.t)
	})
}

func TestNameSanitization(t *testing.T) {
	t.Run("sanitizeTestName processes test names correctly", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"TestSimple", "Simple"},
			{"TestWithSpaces AndMore", "WithSpaces_AndMore"},
			{"TestWith/Slashes/AndMore", "With_Slashes_AndMore"},
			{"Test_Already_Underscored", "_Already_Underscored"},
			{"TestMixed/Slash AndSpace", "Mixed_Slash_AndSpace"},
			{"NoTestPrefix", "NoTestPrefix"},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				result := sanitizeTestName(tc.input)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("sanitizeSnapshotName handles special characters", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"simple_test", "simple_test"},
			{"test@name#with$special%chars", "test_name_with_special_chars"},
			{"test-with-dashes", "test-with-dashes"},
			{"test_with_underscores", "test_with_underscores"},
			{"MixedCase123", "MixedCase123"},
			{"test!@#$%^&*()", "test__________"},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				result := sanitizeSnapshotName(tc.input)
				assert.Equal(t, tc.expected, result)
			})
		}
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("handles read errors gracefully", func(t *testing.T) {
		tempDir := t.TempDir()
		mockT := newMockTestingT("ReadErrorTest")
		snapshot := New(mockT, Options{SnapshotDir: tempDir})

		// Create a snapshot file
		snapshot.Match("error_test", "original data")

		// Make the file unreadable
		snapshotPath := filepath.Join(tempDir, "ReadErrorTest_error_test.snap")
		err := os.Chmod(snapshotPath, 0000)
		require.NoError(t, err)

		// Try to read it back (should fail)
		defer func() {
			recover() // Expected panic from FailNow
		}()

		snapshot.Match("error_test", "new data")

		assert.True(t, mockT.isFailed())
	})

	t.Run("handles write errors gracefully", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Make directory read-only
		err := os.Chmod(tempDir, 0555)
		require.NoError(t, err)
		defer os.Chmod(tempDir, 0755) // Restore for cleanup

		mockT := newMockTestingT("WriteErrorTest")

		defer func() {
			recover() // Expected panic from FailNow
		}()

		snapshot := New(mockT, Options{
			SnapshotDir:     tempDir,
			UpdateSnapshots: true,
		})
		snapshot.Match("write_error_test", "test data")

		assert.True(t, mockT.isFailed())
	})

	t.Run("updateSnapshotE returns error instead of panicking", func(t *testing.T) {
		tempDir := t.TempDir()
		err := os.Chmod(tempDir, 0555) // Read-only
		require.NoError(t, err)
		defer os.Chmod(tempDir, 0755)

		snapshot := New(t, Options{SnapshotDir: tempDir})
		
		filename := filepath.Join(tempDir, "test.snap")
		err = snapshot.updateSnapshotE(filename, []byte("data"))
		assert.Error(t, err)
	})
}