package lambda

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
)

// Test FilterLogs utility function
func TestFilterLogs_WithPattern(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{Message: "ERROR: Something went wrong"},
		{Message: "INFO: Processing request"},
		{Message: "ERROR: Another error"},
		{Message: "DEBUG: Debug info"},
	}
	
	// Act
	filtered := FilterLogs(logs, "ERROR")
	
	// Assert
	assert.Len(t, filtered, 2)
}

func TestFilterLogs_EmptyPattern(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{Message: "Test message"},
		{Message: "Another message"},
	}
	
	// Act
	filtered := FilterLogs(logs, "")
	
	// Assert
	assert.Len(t, filtered, 2) // Empty pattern matches all
}

func TestFilterLogs_EmptyLogs(t *testing.T) {
	// Act
	filtered := FilterLogs([]LogEntry{}, "ERROR")
	
	// Assert
	assert.Len(t, filtered, 0)
}

// Test FilterLogsByLevel utility function
func TestFilterLogsByLevel_MatchingLevel(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{Level: "ERROR"},
		{Level: "INFO"},
		{Level: "ERROR"},
		{Level: "DEBUG"},
	}
	
	// Act
	filtered := FilterLogsByLevel(logs, "ERROR")
	
	// Assert
	assert.Len(t, filtered, 2)
}

func TestFilterLogsByLevel_CaseInsensitive(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{Level: "ERROR"},
		{Level: "INFO"},
	}
	
	// Act
	filtered := FilterLogsByLevel(logs, "error")
	
	// Assert
	assert.Len(t, filtered, 1)
}

// Test GetLogStats utility function
func TestGetLogStats_ValidLogs(t *testing.T) {
	// Arrange
	baseTime := time.Now()
	logs := []LogEntry{
		{Timestamp: baseTime, Level: "INFO"},
		{Timestamp: baseTime.Add(time.Minute), Level: "ERROR"},
		{Timestamp: baseTime.Add(2 * time.Minute), Level: "INFO"},
	}
	
	// Act
	stats := GetLogStats(logs)
	
	// Assert
	assert.Equal(t, 3, stats.TotalEntries)
	assert.Equal(t, 2, stats.LevelCounts["INFO"])
	assert.Equal(t, 1, stats.LevelCounts["ERROR"])
	assert.Equal(t, 2*time.Minute, stats.TimeSpan)
}

func TestGetLogStats_EmptyLogs(t *testing.T) {
	// Act
	stats := GetLogStats([]LogEntry{})
	
	// Assert
	assert.Equal(t, 0, stats.TotalEntries)
	assert.Equal(t, 0, len(stats.LevelCounts))
}

func TestGetLogStats_SingleLog(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{Timestamp: time.Now(), Level: "INFO"},
	}
	
	// Act
	stats := GetLogStats(logs)
	
	// Assert
	assert.Equal(t, 1, stats.TotalEntries)
	assert.Equal(t, time.Duration(0), stats.TimeSpan)
}

// Test LogsContain utility function
func TestLogsContain_MatchingText(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{Message: "Processing request"},
		{Message: "Error occurred"},
	}
	
	// Act & Assert
	assert.True(t, LogsContain(logs, "Processing"))
	assert.True(t, LogsContain(logs, "Error"))
	assert.False(t, LogsContain(logs, "Missing"))
}

func TestLogsContain_EmptyText(t *testing.T) {
	// Arrange
	logs := []LogEntry{{Message: "Test"}}
	
	// Act & Assert
	assert.True(t, LogsContain(logs, ""))
}

func TestLogsContain_EmptyLogs(t *testing.T) {
	// Act & Assert
	assert.False(t, LogsContain([]LogEntry{}, "test"))
}

// Test LogsContainLevel utility function
func TestLogsContainLevel_MatchingLevel(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{Level: "ERROR"},
		{Level: "INFO"},
	}
	
	// Act & Assert
	assert.True(t, LogsContainLevel(logs, "ERROR"))
	assert.True(t, LogsContainLevel(logs, "error")) // Case insensitive
	assert.False(t, LogsContainLevel(logs, "DEBUG"))
}

func TestLogsContainLevel_EmptyLogs(t *testing.T) {
	// Act & Assert
	assert.False(t, LogsContainLevel([]LogEntry{}, "ERROR"))
}

// Test GetLogsByRequestID utility function
func TestGetLogsByRequestID_MatchingID(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{RequestID: "12345"},
		{RequestID: "67890"},
		{RequestID: "12345"},
	}
	
	// Act
	filtered := GetLogsByRequestID(logs, "12345")
	
	// Assert
	assert.Len(t, filtered, 2)
}

func TestGetLogsByRequestID_NonMatchingID(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{RequestID: "12345"},
		{RequestID: "67890"},
	}
	
	// Act
	filtered := GetLogsByRequestID(logs, "99999")
	
	// Assert
	assert.Len(t, filtered, 0)
}

// Test ValidateFunctionConfiguration utility function
func TestValidateFunctionConfiguration_ValidConfig(t *testing.T) {
	// Arrange
	config := &FunctionConfiguration{
		FunctionName: "test-func",
		Runtime:      types.RuntimeNodejs18x,
		Handler:      "index.handler",
	}
	expected := &FunctionConfiguration{
		FunctionName: "test-func",
		Runtime:      types.RuntimeNodejs18x,
	}
	
	// Act
	errors := ValidateFunctionConfiguration(config, expected)
	
	// Assert
	assert.Len(t, errors, 0)
}

func TestValidateFunctionConfiguration_NilConfig(t *testing.T) {
	// Arrange
	expected := &FunctionConfiguration{FunctionName: "test"}
	
	// Act
	errors := ValidateFunctionConfiguration(nil, expected)
	
	// Assert
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "function configuration is nil")
}

func TestValidateFunctionConfiguration_NilExpected(t *testing.T) {
	// Arrange
	config := &FunctionConfiguration{FunctionName: "test"}
	
	// Act
	errors := ValidateFunctionConfiguration(config, nil)
	
	// Assert
	assert.Len(t, errors, 0)
}

// Test ValidateInvokeResult utility function
func TestValidateInvokeResult_ValidSuccess(t *testing.T) {
	// Arrange
	result := &InvokeResult{
		StatusCode:    200,
		Payload:       `{"success": true}`,
		FunctionError: "",
	}
	
	// Act
	errors := ValidateInvokeResult(result, true, "success")
	
	// Assert
	assert.Len(t, errors, 0)
}

func TestValidateInvokeResult_NilResult(t *testing.T) {
	// Act
	errors := ValidateInvokeResult(nil, true, "")
	
	// Assert
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "invoke result is nil")
}

func TestValidateInvokeResult_UnexpectedSuccess(t *testing.T) {
	// Arrange
	result := &InvokeResult{
		StatusCode:    200,
		FunctionError: "",
	}
	
	// Act
	errors := ValidateInvokeResult(result, false, "")
	
	// Assert
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected failure but invocation succeeded")
}

// Test ValidateLogEntries utility function
func TestValidateLogEntries_Valid(t *testing.T) {
	// Arrange
	logs := []LogEntry{
		{Message: "Processing", Level: "INFO"},
		{Message: "Error occurred", Level: "ERROR"},
	}
	
	// Act
	errors := ValidateLogEntries(logs, 2, "Processing", "ERROR")
	
	// Assert
	assert.Len(t, errors, 0)
}

func TestValidateLogEntries_WrongCount(t *testing.T) {
	// Arrange
	logs := []LogEntry{{Message: "Test"}}
	
	// Act
	errors := ValidateLogEntries(logs, 3, "", "")
	
	// Assert
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected 3 log entries, got 1")
}

func TestValidateLogEntries_MissingContent(t *testing.T) {
	// Arrange
	logs := []LogEntry{{Message: "Test"}}
	
	// Act
	errors := ValidateLogEntries(logs, -1, "Missing", "")
	
	// Assert
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected logs to contain 'Missing'")
}

func TestValidateLogEntries_MissingLevel(t *testing.T) {
	// Arrange
	logs := []LogEntry{{Level: "INFO"}}
	
	// Act
	errors := ValidateLogEntries(logs, -1, "", "ERROR")
	
	// Assert
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "expected logs to contain level 'ERROR'")
}