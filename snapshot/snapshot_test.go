package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSnapshot(t *testing.T) {
	t.Run("creates snapshot with default options", func(t *testing.T) {
		snapshot := New(t)
		
		assert.NotNil(t, snapshot)
		assert.Equal(t, "testdata/snapshots", snapshot.options.SnapshotDir)
		assert.Equal(t, 3, snapshot.options.DiffContextLines)
		assert.True(t, snapshot.options.JSONIndent)
	})

	t.Run("creates snapshot with custom options", func(t *testing.T) {
		opts := Options{
			SnapshotDir:      "custom/snapshots",
			UpdateSnapshots:  true,
			DiffContextLines: 5,
			JSONIndent:       false,
		}
		
		snapshot := New(t, opts)
		
		assert.Equal(t, "custom/snapshots", snapshot.options.SnapshotDir)
		assert.True(t, snapshot.options.UpdateSnapshots)
		assert.Equal(t, 5, snapshot.options.DiffContextLines)
		assert.False(t, snapshot.options.JSONIndent)
	})

	t.Run("respects UPDATE_SNAPSHOTS environment variable", func(t *testing.T) {
		os.Setenv("UPDATE_SNAPSHOTS", "true")
		defer os.Unsetenv("UPDATE_SNAPSHOTS")
		
		snapshot := New(t)
		
		assert.True(t, snapshot.options.UpdateSnapshots)
	})
}

func TestMatch(t *testing.T) {
	t.Run("creates new snapshot when none exists", func(t *testing.T) {
		tempDir := t.TempDir()
		
		snapshot := New(t, Options{SnapshotDir: tempDir})
		data := "test data"
		
		snapshot.Match("test", data)
		
		// Verify snapshot file was created
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1, "Expected exactly one snapshot file to be created")
		
		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		assert.Equal(t, data, string(content))
	})

	t.Run("matches existing snapshot", func(t *testing.T) {
		tempDir := t.TempDir()
		
		snapshot := New(t, Options{SnapshotDir: tempDir})
		
		// First create the snapshot
		snapshot.Match("test", "test data")
		
		// Then verify it matches on subsequent runs
		snapshot.Match("test", "test data")
	})

	t.Run("applies sanitizers before comparison", func(t *testing.T) {
		tempDir := t.TempDir()
		
		sanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("dynamic"), []byte("static"))
		}
		
		snapshot := New(t, Options{
			SnapshotDir: tempDir,
			Sanitizers:  []Sanitizer{sanitizer},
		})
		
		snapshot.Match("test", "dynamic data")
		
		// Verify sanitized data was stored
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1, "Expected exactly one snapshot file to be created")
		
		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		assert.Equal(t, "static data", string(content))
	})
}

func TestMatchJSON(t *testing.T) {
	t.Run("formats JSON with indentation", func(t *testing.T) {
		tempDir := t.TempDir()
		
		snapshot := New(t, Options{SnapshotDir: tempDir})
		data := map[string]interface{}{
			"name": "test",
			"value": 42,
		}
		
		snapshot.MatchJSON("test", data)
		
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1, "Expected exactly one snapshot file to be created")
		
		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		
		var parsed map[string]interface{}
		err = json.Unmarshal(content, &parsed)
		require.NoError(t, err)
		
		// Check that the keys exist and values are approximately equal
		assert.Equal(t, "test", parsed["name"])
		assert.Equal(t, float64(42), parsed["value"]) // JSON unmarshals numbers as float64
	})
}

func TestMatchLambdaResponse(t *testing.T) {
	t.Run("handles JSON Lambda response with sanitization", func(t *testing.T) {
		tempDir := t.TempDir()
		
		snapshot := New(t, Options{SnapshotDir: tempDir})
		
		response := map[string]interface{}{
			"statusCode": 200,
			"body":       "success",
			"requestId":  "abc-123-def",
			"timestamp":  "2023-01-01T10:00:00Z",
		}
		
		payload, _ := json.Marshal(response)
		snapshot.MatchLambdaResponse("test", payload)
		
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1, "Expected exactly one snapshot file to be created")
		
		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		
		// Verify request ID and timestamp were sanitized
		assert.Contains(t, string(content), "<REQUEST_ID>")
		assert.Contains(t, string(content), "<TIMESTAMP>")
	})
}

func TestSanitizers(t *testing.T) {
	t.Run("SanitizeTimestamps replaces ISO timestamps", func(t *testing.T) {
		sanitizer := SanitizeTimestamps()
		input := []byte(`{"timestamp": "2023-01-01T10:00:00Z"}`)
		result := sanitizer(input)
		
		assert.Contains(t, string(result), "<TIMESTAMP>")
	})

	t.Run("SanitizeUUIDs replaces UUID patterns", func(t *testing.T) {
		sanitizer := SanitizeUUIDs()
		input := []byte(`{"id": "550e8400-e29b-41d4-a716-446655440000"}`)
		result := sanitizer(input)
		
		assert.Contains(t, string(result), "<UUID>")
	})

	t.Run("SanitizeLambdaRequestID replaces request IDs", func(t *testing.T) {
		sanitizer := SanitizeLambdaRequestID()
		input := []byte(`{"requestId": "abc-123-def"}`)
		result := sanitizer(input)
		
		assert.Contains(t, string(result), "<REQUEST_ID>")
	})

	t.Run("SanitizeExecutionArn replaces Step Functions execution ARNs", func(t *testing.T) {
		sanitizer := SanitizeExecutionArn()
		input := []byte(`arn:aws:states:us-east-1:123456789012:execution:MyStateMachine:test-execution`)
		result := sanitizer(input)
		
		assert.Contains(t, string(result), "<EXECUTION_ARN>")
	})

	t.Run("SanitizeDynamoDBMetadata removes metadata fields", func(t *testing.T) {
		sanitizer := SanitizeDynamoDBMetadata()
		input := []byte(`{"Items": [], "ConsumedCapacity": 1.0, "ScannedCount": 5}`)
		result := sanitizer(input)
		
		assert.NotContains(t, string(result), "ConsumedCapacity")
		assert.NotContains(t, string(result), "ScannedCount")
	})
}

func TestWithSanitizers(t *testing.T) {
	t.Run("combines existing and new sanitizers", func(t *testing.T) {
		existingSanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("old"), []byte("existing"))
		}
		
		newSanitizer := func(data []byte) []byte {
			return bytes.ReplaceAll(data, []byte("new"), []byte("added"))
		}
		
		snapshot := New(t, Options{
			SnapshotDir: t.TempDir(),
			Sanitizers:  []Sanitizer{existingSanitizer},
		})
		
		combined := snapshot.WithSanitizers(newSanitizer)
		
		assert.Len(t, combined.options.Sanitizers, 2)
		assert.NotEqual(t, &snapshot.options.Sanitizers, &combined.options.Sanitizers)
	})
}

func TestGlobalFunctions(t *testing.T) {
	t.Run("Match creates default snapshot", func(t *testing.T) {
		// Override default snapshot directory for test
		originalSnapshot := defaultSnapshot
		defer func() { defaultSnapshot = originalSnapshot }()
		
		Match(t, "test", "data")
		
		// Verify snapshot was created
		assert.NotNil(t, defaultSnapshot)
	})

	t.Run("MatchJSON creates default snapshot for JSON", func(t *testing.T) {
		originalSnapshot := defaultSnapshot
		defer func() { defaultSnapshot = originalSnapshot }()
		
		data := map[string]string{"key": "value"}
		MatchJSON(t, "test", data)
		
		assert.NotNil(t, defaultSnapshot)
	})
}

// Test helper that implements TestingT interface for testing
type testHelper struct {
	failed bool
	errors []string
}

func (th *testHelper) Errorf(format string, args ...interface{}) {
	th.failed = true
	th.errors = append(th.errors, fmt.Sprintf(format, args...))
}

func (th *testHelper) FailNow() {
	th.failed = true
	panic("test failed")
}

func TestToBytes(t *testing.T) {
	snapshot := New(t)

	t.Run("converts byte slice", func(t *testing.T) {
		input := []byte("test data")
		result, err := snapshot.toBytes(input)
		
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("converts string", func(t *testing.T) {
		input := "test data"
		result, err := snapshot.toBytes(input)
		
		require.NoError(t, err)
		assert.Equal(t, []byte(input), result)
	})

	t.Run("converts strings.Reader", func(t *testing.T) {
		input := strings.NewReader("test data")
		result, err := snapshot.toBytes(input)
		
		require.NoError(t, err)
		assert.Equal(t, []byte("test data"), result)
	})

	t.Run("converts struct to JSON with indent", func(t *testing.T) {
		input := map[string]string{"key": "value"}
		result, err := snapshot.toBytes(input)
		
		require.NoError(t, err)
		assert.True(t, json.Valid(result))
		assert.Contains(t, string(result), "  ") // Check for indentation
	})
}

func TestSanitizeNames(t *testing.T) {
	t.Run("sanitizeTestName removes Test prefix and replaces invalid chars", func(t *testing.T) {
		result := sanitizeTestName("TestMyFunction/WithSlash AndSpace")
		expected := "MyFunction_WithSlash_AndSpace"
		assert.Equal(t, expected, result)
	})

	t.Run("sanitizeSnapshotName replaces non-alphanumeric characters", func(t *testing.T) {
		result := sanitizeSnapshotName("test@name#with$special%chars")
		expected := "test_name_with_special_chars"
		assert.Equal(t, expected, result)
	})
}