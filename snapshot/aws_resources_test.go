package snapshot

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Following TDD: RED - Write failing tests first

func TestMatchDynamoDBItems(t *testing.T) {
	t.Run("formats DynamoDB items correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		items := []map[string]interface{}{
			{
				"id":     "item1",
				"name":   "Test Item 1",
				"active": true,
				"count":  42,
			},
			{
				"id":     "item2", 
				"name":   "Test Item 2",
				"active": false,
				"count":  0,
			},
		}

		snapshot.MatchDynamoDBItems("dynamodb_items", items)

		// Verify file was created
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		// Verify content is valid JSON
		content, err := os.ReadFile(tempDir + "/" + files[0].Name())
		require.NoError(t, err)

		var parsed []map[string]interface{}
		err = json.Unmarshal(content, &parsed)
		require.NoError(t, err)
		assert.Len(t, parsed, 2)
		assert.Equal(t, "item1", parsed[0]["id"])
	})

	t.Run("applies DynamoDB-specific sanitizers", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		items := []map[string]interface{}{
			{
				"id":        "item1",
				"timestamp": "2023-01-01T12:00:00Z",
			},
		}

		snapshot.MatchDynamoDBItems("sanitized_items", items)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		content, err := os.ReadFile(tempDir + "/" + files[0].Name())
		require.NoError(t, err)

		// Should contain sanitized timestamp
		assert.Contains(t, string(content), "<TIMESTAMP>")
	})
}

func TestMatchStepFunctionExecution(t *testing.T) {
	t.Run("sanitizes Step Function execution data", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshot := New(t, Options{SnapshotDir: tempDir})

		execution := map[string]interface{}{
			"executionArn": "arn:aws:states:us-east-1:123456789:execution:MyStateMachine:abc-123",
			"status":       "SUCCEEDED", 
			"startDate":    "2023-01-01T10:00:00Z",
			"name":         "550e8400-e29b-41d4-a716-446655440000",
		}

		snapshot.MatchStepFunctionExecution("stepfunction_execution", execution)

		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		content, err := os.ReadFile(tempDir + "/" + files[0].Name())
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "<EXECUTION_ARN>")
		assert.Contains(t, contentStr, "<TIMESTAMP>")
		assert.Contains(t, contentStr, "<UUID>")
	})
}