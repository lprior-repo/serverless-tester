package lambda

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests for log utility functions that have 0% coverage

func TestFilterLogs_WithMatchingPattern_UsingMocks_ShouldReturnFilteredLogs(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "INFO This is an info message",
			Level:     "INFO",
			RequestID: "abc-123",
		},
		{
			Timestamp: time.Now().Add(1 * time.Second),
			Message:   "ERROR Something went wrong",
			Level:     "ERROR",
			RequestID: "abc-123",
		},
		{
			Timestamp: time.Now().Add(2 * time.Second),
			Message:   "DEBUG Debug information",
			Level:     "DEBUG",
			RequestID: "def-456",
		},
	}
	
	// Act
	filtered := FilterLogs(logs, "ERROR")
	
	// Assert
	assert.Len(t, filtered, 1, "Should return 1 log entry containing ERROR")
	assert.Contains(t, filtered[0].Message, "ERROR")
}

func TestFilterLogs_WithNonMatchingPattern_UsingMocks_ShouldReturnEmpty(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "INFO This is an info message",
			Level:     "INFO",
			RequestID: "abc-123",
		},
	}
	
	// Act
	filtered := FilterLogs(logs, "NONEXISTENT")
	
	// Assert
	assert.Len(t, filtered, 0, "Should return empty slice for non-matching pattern")
}

func TestFilterLogsByLevel_WithMatchingLevel_UsingMocks_ShouldReturnFilteredLogs(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "INFO This is an info message",
			Level:     "INFO",
			RequestID: "abc-123",
		},
		{
			Timestamp: time.Now().Add(1 * time.Second),
			Message:   "ERROR Something went wrong",
			Level:     "ERROR",
			RequestID: "abc-123",
		},
	}
	
	// Act
	filtered := FilterLogsByLevel(logs, "INFO")
	
	// Assert
	assert.Len(t, filtered, 1, "Should return 1 log entry with INFO level")
	assert.Equal(t, "INFO", filtered[0].Level)
}

func TestFilterLogsByLevel_WithNonMatchingLevel_UsingMocks_ShouldReturnEmpty(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "INFO This is an info message",
			Level:     "INFO",
			RequestID: "abc-123",
		},
	}
	
	// Act
	filtered := FilterLogsByLevel(logs, "WARN")
	
	// Assert
	assert.Len(t, filtered, 0, "Should return empty slice for non-matching level")
}

func TestGetLogStats_WithMultipleLogs_UsingMocks_ShouldReturnCorrectStats(t *testing.T) {
	// Arrange
	baseTime := time.Now()
	logs := []LogEntry{
		{
			Timestamp: baseTime,
			Message:   "INFO First message",
			Level:     "INFO",
			RequestID: "abc-123",
		},
		{
			Timestamp: baseTime.Add(1 * time.Second),
			Message:   "ERROR Something went wrong",
			Level:     "ERROR",
			RequestID: "abc-123",
		},
		{
			Timestamp: baseTime.Add(2 * time.Second),
			Message:   "DEBUG Debug information",
			Level:     "DEBUG",
			RequestID: "def-456",
		},
	}
	
	// Act
	stats := GetLogStats(logs)
	
	// Assert
	assert.Equal(t, 3, stats.TotalEntries, "Should count all log entries")
	assert.Equal(t, 1, stats.LevelCounts["INFO"], "Should count INFO logs")
	assert.Equal(t, 1, stats.LevelCounts["ERROR"], "Should count ERROR logs")
	assert.Equal(t, 1, stats.LevelCounts["DEBUG"], "Should count DEBUG logs")
	assert.True(t, stats.TimeSpan > 0, "Should calculate time span between earliest and latest")
	assert.Equal(t, baseTime, stats.EarliestEntry, "Should identify earliest entry")
	assert.Equal(t, baseTime.Add(2*time.Second), stats.LatestEntry, "Should identify latest entry")
}

func TestGetLogStats_WithEmptyLogs_UsingMocks_ShouldReturnEmptyStats(t *testing.T) {
	// Arrange
	logs := []LogEntry{}
	
	// Act
	stats := GetLogStats(logs)
	
	// Assert
	assert.Equal(t, 0, stats.TotalEntries, "Should have zero entries")
	assert.Empty(t, stats.LevelCounts, "Should have empty level counts")
	assert.True(t, stats.EarliestEntry.IsZero(), "Should have zero earliest entry")
	assert.True(t, stats.LatestEntry.IsZero(), "Should have zero latest entry")
	assert.Equal(t, time.Duration(0), stats.TimeSpan, "Should have zero time span")
}

func TestLogsContain_WithMatchingText_UsingMocks_ShouldReturnTrue(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "ERROR Something went wrong",
			Level:     "ERROR",
			RequestID: "abc-123",
		},
	}
	
	// Act
	result := LogsContain(logs, "went wrong")
	
	// Assert
	assert.True(t, result, "Should return true when logs contain the text")
}

func TestLogsContain_WithNonMatchingText_UsingMocks_ShouldReturnFalse(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "INFO Everything is fine",
			Level:     "INFO",
			RequestID: "abc-123",
		},
	}
	
	// Act
	result := LogsContain(logs, "nonexistent text")
	
	// Assert
	assert.False(t, result, "Should return false when logs don't contain the text")
}

func TestLogsContainLevel_WithMatchingLevel_UsingMocks_ShouldReturnTrue(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "ERROR Something went wrong",
			Level:     "ERROR",
			RequestID: "abc-123",
		},
	}
	
	// Act
	result := LogsContainLevel(logs, "ERROR")
	
	// Assert
	assert.True(t, result, "Should return true when logs contain the level")
}

func TestLogsContainLevel_WithNonMatchingLevel_UsingMocks_ShouldReturnFalse(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "INFO Everything is fine",
			Level:     "INFO",
			RequestID: "abc-123",
		},
	}
	
	// Act
	result := LogsContainLevel(logs, "WARN")
	
	// Assert
	assert.False(t, result, "Should return false when logs don't contain the level")
}

func TestGetLogsByRequestID_WithMatchingRequestID_UsingMocks_ShouldReturnFilteredLogs(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "INFO First message",
			Level:     "INFO",
			RequestID: "abc-123",
		},
		{
			Timestamp: time.Now().Add(1 * time.Second),
			Message:   "ERROR Something went wrong",
			Level:     "ERROR",
			RequestID: "abc-123",
		},
		{
			Timestamp: time.Now().Add(2 * time.Second),
			Message:   "DEBUG Debug information",
			Level:     "DEBUG",
			RequestID: "def-456",
		},
	}
	
	// Act
	filtered := GetLogsByRequestID(logs, "abc-123")
	
	// Assert
	assert.Len(t, filtered, 2, "Should return 2 log entries with matching request ID")
	for _, log := range filtered {
		assert.Equal(t, "abc-123", log.RequestID, "All returned logs should have the matching request ID")
	}
}

func TestGetLogsByRequestID_WithNonMatchingRequestID_UsingMocks_ShouldReturnEmpty(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{
			Timestamp: time.Now(),
			Message:   "INFO Everything is fine",
			Level:     "INFO",
			RequestID: "abc-123",
		},
	}
	
	// Act
	filtered := GetLogsByRequestID(logs, "nonexistent-id")
	
	// Assert
	assert.Len(t, filtered, 0, "Should return empty slice for non-matching request ID")
}