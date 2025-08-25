package snapshot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Following TDD: RED - Write failing tests for missing sanitizer functions

func TestSanitizeCustom(t *testing.T) {
	t.Run("creates custom regex-based sanitizer", func(t *testing.T) {
		sanitizer := SanitizeCustom(`secret-\d+`, "<SECRET>")
		
		input := []byte(`{"token": "secret-12345", "data": "public-info"}`)
		expected := []byte(`{"token": "<SECRET>", "data": "public-info"}`)
		
		result := sanitizer(input)
		assert.Equal(t, expected, result)
	})

	t.Run("handles multiple matches", func(t *testing.T) {
		sanitizer := SanitizeCustom(`\bkey\d+\b`, "<KEY>")
		
		input := []byte("key123 and key456 but not keyword")
		expected := []byte("<KEY> and <KEY> but not keyword")
		
		result := sanitizer(input)
		assert.Equal(t, expected, result)
	})

	t.Run("handles no matches gracefully", func(t *testing.T) {
		sanitizer := SanitizeCustom(`nomatch`, "<REPLACE>")
		
		input := []byte("this text has nothing to replace")
		expected := input // Should be unchanged
		
		result := sanitizer(input)
		assert.Equal(t, expected, result)
	})
}

func TestSanitizeField(t *testing.T) {
	t.Run("replaces specific field values", func(t *testing.T) {
		sanitizer := SanitizeField("password", "<REDACTED>")
		
		input := []byte(`{"username": "alice", "password": "secret123", "email": "alice@example.com"}`)
		expected := []byte(`{"username": "alice", "password": "<REDACTED>", "email": "alice@example.com"}`)
		
		result := sanitizer(input)
		assert.Equal(t, expected, result)
	})

	t.Run("handles multiple instances of same field", func(t *testing.T) {
		sanitizer := SanitizeField("token", "<TOKEN>")
		
		input := []byte(`{"token": "abc123", "refresh_token": "def456"}`)
		// Should only replace exact field name "token", not "refresh_token"
		expected := []byte(`{"token": "<TOKEN>", "refresh_token": "def456"}`)
		
		result := sanitizer(input)
		assert.Equal(t, expected, result)
	})

	t.Run("handles field not present", func(t *testing.T) {
		sanitizer := SanitizeField("nonexistent", "<VALUE>")
		
		input := []byte(`{"username": "alice", "email": "alice@example.com"}`)
		expected := input // Should be unchanged
		
		result := sanitizer(input)
		assert.Equal(t, expected, result)
	})

	t.Run("handles special characters in field name", func(t *testing.T) {
		sanitizer := SanitizeField("api-key", "<API_KEY>")
		
		input := []byte(`{"api-key": "secret-value", "other": "data"}`)
		expected := []byte(`{"api-key": "<API_KEY>", "other": "data"}`)
		
		result := sanitizer(input)
		assert.Equal(t, expected, result)
	})
}