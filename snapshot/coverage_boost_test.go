package snapshot

import (
	"encoding/json"
	"os"
	"path/filepath" 
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Following TDD: Focus on the remaining coverage gaps to reach 90%

func TestNewEdgeCases(t *testing.T) {
	t.Run("handles directory creation failure path", func(t *testing.T) {
		// Create a mock that panics on directory creation error
		mockT := &testFailMock{name: "FailureTest"}
		
		defer func() {
			r := recover()
			assert.NotNil(t, r, "Expected panic from FailNow")
		}()

		// Try to create in invalid path
		_ = New(mockT, Options{SnapshotDir: "/nonexistent/readonly/path"})
	})

	t.Run("extracts test name when available", func(t *testing.T) {
		mockWithName := &testWithName{name: "MyTestName"}
		tempDir := t.TempDir()
		
		snapshot := New(mockWithName, Options{SnapshotDir: tempDir})
		
		assert.Contains(t, snapshot.testName, "MyTestName")
	})
}

func TestMatchEdgeCases(t *testing.T) {
	t.Run("handles snapshot read error", func(t *testing.T) {
		tempDir := t.TempDir()
		mockT := &testFailMock{name: "ReadErrorTest"}
		snapshot := New(mockT, Options{SnapshotDir: tempDir})
		
		// Create a snapshot file first
		snapshot.Match("read_error_test", "original data")
		
		// Make the file unreadable
		snapshotPath := filepath.Join(tempDir, "ReadErrorTest_read_error_test.snap")
		err := os.Chmod(snapshotPath, 0000)
		require.NoError(t, err)
		defer os.Chmod(snapshotPath, 0644) // Restore permissions
		
		defer func() {
			r := recover()
			assert.NotNil(t, r, "Expected panic from FailNow")
		}()
		
		// Try to read it - should fail
		snapshot.Match("read_error_test", "new data")
	})

	t.Run("handles updateSnapshot write error", func(t *testing.T) {
		tempDir := t.TempDir()
		mockT := &testFailMock{name: "WriteErrorTest"}
		
		// Make directory read-only after creation
		err := os.Chmod(tempDir, 0555)
		require.NoError(t, err)
		defer os.Chmod(tempDir, 0755)
		
		snapshot := New(mockT, Options{
			SnapshotDir:     tempDir,
			UpdateSnapshots: true,
		})
		
		defer func() {
			r := recover()
			assert.NotNil(t, r, "Expected panic from FailNow")
		}()
		
		snapshot.Match("write_error_test", "test data")
	})
}

func TestMatchJSONEdgeCases(t *testing.T) {
	t.Run("handles JSON marshal error", func(t *testing.T) {
		tempDir := t.TempDir()
		mockT := &testFailMock{name: "JSONErrorTest"}
		snapshot := New(mockT, Options{SnapshotDir: tempDir})
		
		defer func() {
			r := recover()
			assert.NotNil(t, r, "Expected panic from FailNow")
		}()
		
		// Try to marshal invalid data
		invalidData := make(chan int) // Channels can't be marshaled
		snapshot.MatchJSON("invalid_json", invalidData)
	})
}

func TestMatchLambdaResponseEdgeCases(t *testing.T) {
	t.Run("handles non-JSON payload", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})
		
		// Test with plain text (non-JSON)
		payload := []byte("plain text response")
		snapshot.MatchLambdaResponse("lambda_text", payload)
		
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)
		
		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		assert.Equal(t, "plain text response", string(content))
	})

	t.Run("applies custom sanitizers correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})
		
		customSanitizer := SanitizeField("secret", "<HIDDEN>")
		
		response := map[string]interface{}{
			"statusCode": 200,
			"secret":     "secret-value",
			"requestId":  "req-123",
		}
		
		payload, _ := json.Marshal(response)
		snapshot.MatchLambdaResponse("lambda_custom", payload, customSanitizer)
		
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
		require.NoError(t, err)
		
		contentStr := string(content)
		assert.Contains(t, contentStr, "<HIDDEN>")
		assert.Contains(t, contentStr, "<REQUEST_ID>")
		assert.NotContains(t, contentStr, "secret-value")
		assert.NotContains(t, contentStr, "req-123")
	})
}

func TestMatchEEdgeCases(t *testing.T) {
	t.Run("handles read file not found error", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})
		
		// Try to read non-existent snapshot - should create new one
		err := snapshot.MatchE("nonexistent", "new data")
		assert.NoError(t, err)
		
		// Verify file was created
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		assert.Len(t, files, 1)
	})

	t.Run("handles updateSnapshotE failure in update mode", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Make directory read-only
		err := os.Chmod(tempDir, 0555)
		require.NoError(t, err)
		defer os.Chmod(tempDir, 0755)
		
		snapshot := New(t, Options{
			SnapshotDir:     tempDir,
			UpdateSnapshots: true,
		})
		
		err = snapshot.MatchE("write_fail", "data")
		assert.Error(t, err)
	})

	t.Run("handles updateSnapshotE failure for new snapshot", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Make directory read-only after creation
		err := os.Chmod(tempDir, 0555)
		require.NoError(t, err)
		defer os.Chmod(tempDir, 0755)
		
		snapshot := New(t, Options{SnapshotDir: tempDir})
		
		err = snapshot.MatchE("new_fail", "data")
		assert.Error(t, err)
	})
}

// Mock implementations for testing edge cases

type testFailMock struct {
	name string
}

func (m *testFailMock) Helper()                               {}
func (m *testFailMock) Errorf(format string, args ...interface{}) {}
func (m *testFailMock) Error(args ...interface{})            {}
func (m *testFailMock) Fail()                                {}
func (m *testFailMock) FailNow()                             { panic("FailNow called") }
func (m *testFailMock) Fatal(args ...interface{})            { panic("Fatal called") }
func (m *testFailMock) Fatalf(format string, args ...interface{}) { panic("Fatalf called") }
func (m *testFailMock) Logf(format string, args ...interface{}) {}
func (m *testFailMock) Name() string                         { return m.name }

type testWithName struct {
	name string
}

func (m *testWithName) Helper()                               {}
func (m *testWithName) Errorf(format string, args ...interface{}) {}
func (m *testWithName) Error(args ...interface{})            {}
func (m *testWithName) Fail()                                {}
func (m *testWithName) FailNow()                             {}
func (m *testWithName) Fatal(args ...interface{})            {}
func (m *testWithName) Fatalf(format string, args ...interface{}) {}
func (m *testWithName) Logf(format string, args ...interface{}) {}
func (m *testWithName) Name() string                         { return m.name }