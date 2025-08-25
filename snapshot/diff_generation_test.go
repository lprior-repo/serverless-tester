package snapshot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Following TDD: RED - Write failing tests for diff generation functionality

func TestGenerateDiff(t *testing.T) {
	t.Run("generates diff for simple text changes", func(t *testing.T) {
		snapshot := New(t, Options{
			SnapshotDir:      t.TempDir(),
			DiffContextLines: 3,
		})

		expected := []byte("line1\noriginal_line\nline3")
		actual := []byte("line1\nmodified_line\nline3")

		diff := snapshot.generateDiff(expected, actual, "test.snap")

		assert.NotEmpty(t, diff)
		assert.Contains(t, diff, "-original_line")
		assert.Contains(t, diff, "+modified_line")
		assert.Contains(t, diff, "@@") // Hunk header
	})

	t.Run("respects context lines setting", func(t *testing.T) {
		testCases := []struct {
			contextLines int
			name         string
		}{
			{0, "zero_context"},
			{1, "one_context"},
			{5, "five_context"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				snapshot := New(t, Options{
					SnapshotDir:      t.TempDir(),
					DiffContextLines: tc.contextLines,
				})

				expected := []byte("line1\nline2\noriginal\nline4\nline5\nline6\nline7")
				actual := []byte("line1\nline2\nmodified\nline4\nline5\nline6\nline7")

				diff := snapshot.generateDiff(expected, actual, "context.snap")

				assert.NotEmpty(t, diff)
				assert.Contains(t, diff, "-original")
				assert.Contains(t, diff, "+modified")
			})
		}
	})

	t.Run("handles identical content", func(t *testing.T) {
		snapshot := New(t, Options{SnapshotDir: t.TempDir()})

		content := []byte("identical content")

		diff := snapshot.generateDiff(content, content, "identical.snap")

		// Diff of identical content should be empty
		assert.Empty(t, strings.TrimSpace(diff))
	})

	t.Run("handles binary data", func(t *testing.T) {
		snapshot := New(t, Options{SnapshotDir: t.TempDir()})

		expected := []byte{0x00, 0x01, 0x02, 0x03, 0xFF}
		actual := []byte{0x00, 0x01, 0x99, 0x03, 0xFF}

		diff := snapshot.generateDiff(expected, actual, "binary.snap")

		assert.NotEmpty(t, diff)
		// Should handle binary data gracefully
	})

	t.Run("includes file names in diff output", func(t *testing.T) {
		snapshot := New(t, Options{SnapshotDir: t.TempDir()})

		expected := []byte("old content")
		actual := []byte("new content")

		diff := snapshot.generateDiff(expected, actual, "named_test.snap")

		assert.Contains(t, diff, "named_test.snap")
		assert.Contains(t, diff, "actual")
	})

	t.Run("handles multiline changes", func(t *testing.T) {
		snapshot := New(t, Options{SnapshotDir: t.TempDir()})

		expected := []byte("line1\nline2\nline3\nline4\nline5")
		actual := []byte("line1\nmodified2\nmodified3\nline4\nline5")

		diff := snapshot.generateDiff(expected, actual, "multiline.snap")

		assert.NotEmpty(t, diff)
		assert.Contains(t, diff, "-line2")
		assert.Contains(t, diff, "-line3")
		assert.Contains(t, diff, "+modified2")
		assert.Contains(t, diff, "+modified3")
	})

	t.Run("handles additions and deletions", func(t *testing.T) {
		snapshot := New(t, Options{SnapshotDir: t.TempDir()})

		expected := []byte("line1\nline2")
		actual := []byte("line1\nline2\nadded_line")

		diff := snapshot.generateDiff(expected, actual, "additions.snap")

		assert.NotEmpty(t, diff)
		assert.Contains(t, diff, "+added_line")
	})

	t.Run("handles large diffs efficiently", func(t *testing.T) {
		snapshot := New(t, Options{SnapshotDir: t.TempDir()})

		// Create large content with a small change
		var expected, actual strings.Builder
		for i := 0; i < 1000; i++ {
			expected.WriteString("line")
			expected.WriteString(string(rune('0' + i%10)))
			expected.WriteString("\n")
			
			if i == 500 {
				actual.WriteString("modified_line\n")
			} else {
				actual.WriteString("line")
				actual.WriteString(string(rune('0' + i%10)))
				actual.WriteString("\n")
			}
		}

		diff := snapshot.generateDiff([]byte(expected.String()), []byte(actual.String()), "large.snap")

		assert.NotEmpty(t, diff)
		assert.Contains(t, diff, "+modified_line")
	})
}