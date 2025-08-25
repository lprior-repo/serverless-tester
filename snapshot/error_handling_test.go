package snapshot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Following TDD: RED - Write failing tests for error handling scenarios

func TestNewE(t *testing.T) {
	t.Run("creates snapshot successfully with valid directory", func(t *testing.T) {
		tempDir := t.TempDir()
		
		snapshot, err := NewE(t, Options{SnapshotDir: tempDir})
		
		require.NoError(t, err)
		assert.NotNil(t, snapshot)
		assert.Equal(t, tempDir, snapshot.options.SnapshotDir)
	})

	t.Run("returns error for invalid directory", func(t *testing.T) {
		invalidDir := "/root/nonexistent/readonly"
		
		_, err := NewE(t, Options{SnapshotDir: invalidDir})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create snapshot directory")
	})

	t.Run("respects UPDATE_SNAPSHOTS environment variable", func(t *testing.T) {
		os.Setenv("UPDATE_SNAPSHOTS", "true")
		defer os.Unsetenv("UPDATE_SNAPSHOTS")
		
		snapshot, err := NewE(t, Options{SnapshotDir: t.TempDir()})
		
		require.NoError(t, err)
		assert.True(t, snapshot.options.UpdateSnapshots)
	})
}

func TestMatchE(t *testing.T) {
	t.Run("creates new snapshot successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		err := snapshot.MatchE("new_test", "test data")

		assert.NoError(t, err)
		
		// Verify file was created
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		assert.Len(t, files, 1)
	})

	t.Run("returns error for data conversion failure", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir, JSONIndent: false})

		// Create invalid data that can't be marshaled
		invalidData := make(chan int)

		err := snapshot.MatchE("invalid_test", invalidData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to convert data to bytes")
	})

	t.Run("returns error for snapshot mismatch", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		// Create initial snapshot
		err := snapshot.MatchE("mismatch_test", "original data")
		require.NoError(t, err)

		// Try to match different data
		err = snapshot.MatchE("mismatch_test", "different data")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "snapshot mismatch")
		assert.Contains(t, err.Error(), "UPDATE_SNAPSHOTS=1")
	})

	t.Run("succeeds when UPDATE_SNAPSHOTS is true", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{
			SnapshotDir:     tempDir,
			UpdateSnapshots: true,
		})

		// Create initial snapshot
		err := snapshot.MatchE("update_test", "original data")
		require.NoError(t, err)

		// Update with different data - should succeed
		err = snapshot.MatchE("update_test", "updated data")
		assert.NoError(t, err)
	})
}

func TestMatchJSONE(t *testing.T) {
	t.Run("creates JSON snapshot successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		data := map[string]string{"key": "value"}
		err := snapshot.MatchJSONE("json_test", data)

		assert.NoError(t, err)
	})

	t.Run("returns error for invalid JSON data", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		// Create data that can't be marshaled to JSON
		invalidData := make(chan int)
		err := snapshot.MatchJSONE("invalid_json", invalidData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal JSON")
	})
}

func TestUpdateSnapshotE(t *testing.T) {
	t.Run("updates snapshot file successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		filename := filepath.Join(tempDir, "test.snap")
		data := []byte("test content")

		err := snapshot.updateSnapshotE(filename, data)

		assert.NoError(t, err)
		
		// Verify file was written
		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, data, content)
	})

	t.Run("returns error for read-only directory", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		// Make directory read-only
		err := os.Chmod(tempDir, 0555)
		require.NoError(t, err)
		defer os.Chmod(tempDir, 0755) // Restore permissions

		filename := filepath.Join(tempDir, "readonly_test.snap")
		data := []byte("test data")

		err = snapshot.updateSnapshotE(filename, data)

		assert.Error(t, err)
	})
}