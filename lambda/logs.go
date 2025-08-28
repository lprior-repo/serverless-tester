package lambda

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// LogsOptions configures log retrieval behavior
type LogsOptions struct {
	StartTime    *time.Time
	EndTime      *time.Time
	MaxLines     int32
	FilterPattern string
	ParseLogs    bool
	Timeout      time.Duration
}

// GetLogs retrieves CloudWatch logs for a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func GetLogs(ctx *TestContext, functionName string) []LogEntry {
	logs, err := GetLogsE(ctx, functionName)
	if err != nil {
		ctx.T.Errorf("Failed to get Lambda function logs: %v", err)
		ctx.T.FailNow()
	}
	return logs
}

// GetLogsE retrieves CloudWatch logs for a Lambda function.
// This is the error returning version that follows Terratest patterns.
func GetLogsE(ctx *TestContext, functionName string) ([]LogEntry, error) {
	return GetLogsWithOptionsE(ctx, functionName, nil)
}

// GetLogsWithOptions retrieves CloudWatch logs for a Lambda function with custom options.
// This is the non-error returning version that follows Terratest patterns.
func GetLogsWithOptions(ctx *TestContext, functionName string, opts *LogsOptions) []LogEntry {
	logs, err := GetLogsWithOptionsE(ctx, functionName, opts)
	if err != nil {
		ctx.T.Errorf("Failed to get Lambda function logs with options: %v", err)
		ctx.T.FailNow()
	}
	return logs
}

// GetLogsWithOptionsE retrieves CloudWatch logs for a Lambda function with custom options.
// This is the error returning version that follows Terratest patterns.
func GetLogsWithOptionsE(ctx *TestContext, functionName string, opts *LogsOptions) ([]LogEntry, error) {
	startTime := time.Now()
	
	// Validate function name
	if err := validateFunctionName(functionName); err != nil {
		return nil, err
	}
	
	// Set default options
	if opts == nil {
		opts = &LogsOptions{}
	}
	opts = mergeLogsOptions(opts)
	
	logOperation("get_logs", functionName, map[string]interface{}{
		"max_lines":      opts.MaxLines,
		"filter_pattern": opts.FilterPattern,
		"parse_logs":     opts.ParseLogs,
	})
	
	// Get log group name for the function
	logGroupName := fmt.Sprintf("/aws/lambda/%s", functionName)
	
	// Retrieve logs from CloudWatch
	logEvents, err := retrieveCloudWatchLogs(ctx, logGroupName, opts)
	if err != nil {
		duration := time.Since(startTime)
		logResult("get_logs", functionName, false, duration, err)
		return nil, err
	}
	
	// Parse logs if requested
	var parsedLogs []LogEntry
	if opts.ParseLogs {
		parsedLogs = parseLogEvents(logEvents)
	} else {
		parsedLogs = convertLogEvents(logEvents)
	}
	
	duration := time.Since(startTime)
	logResult("get_logs", functionName, true, duration, nil)
	
	return parsedLogs, nil
}

// GetRecentLogs retrieves recent CloudWatch logs for a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func GetRecentLogs(ctx *TestContext, functionName string, duration time.Duration) []LogEntry {
	logs, err := GetRecentLogsE(ctx, functionName, duration)
	if err != nil {
		ctx.T.Errorf("Failed to get recent Lambda function logs: %v", err)
		ctx.T.FailNow()
	}
	return logs
}

// GetRecentLogsE retrieves recent CloudWatch logs for a Lambda function.
// This is the error returning version that follows Terratest patterns.
func GetRecentLogsE(ctx *TestContext, functionName string, duration time.Duration) ([]LogEntry, error) {
	endTime := time.Now()
	startTime := endTime.Add(-duration)
	
	opts := &LogsOptions{
		StartTime: &startTime,
		EndTime:   &endTime,
		ParseLogs: true,
	}
	
	return GetLogsWithOptionsE(ctx, functionName, opts)
}

// WaitForLogs waits for log entries to appear in CloudWatch logs.
// This is the non-error returning version that follows Terratest patterns.
func WaitForLogs(ctx *TestContext, functionName string, expectedPattern string, timeout time.Duration) []LogEntry {
	logs, err := WaitForLogsE(ctx, functionName, expectedPattern, timeout)
	if err != nil {
		ctx.T.Errorf("Timed out waiting for logs: %v", err)
		ctx.T.FailNow()
	}
	return logs
}

// WaitForLogsE waits for log entries to appear in CloudWatch logs.
// This is the error returning version that follows Terratest patterns.
func WaitForLogsE(ctx *TestContext, functionName string, expectedPattern string, timeout time.Duration) ([]LogEntry, error) {
	startTime := time.Now()
	
	logOperation("wait_for_logs", functionName, map[string]interface{}{
		"expected_pattern": expectedPattern,
		"timeout_seconds":  timeout.Seconds(),
	})
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	timeoutChan := time.After(timeout)
	
	for {
		select {
		case <-timeoutChan:
			duration := time.Since(startTime)
			logResult("wait_for_logs", functionName, false, duration, 
				fmt.Errorf("timeout waiting for logs with pattern: %s", expectedPattern))
			return nil, fmt.Errorf("timeout waiting for logs with pattern: %s", expectedPattern)
			
		case <-ticker.C:
			// Get recent logs
			logs, err := GetRecentLogsE(ctx, functionName, 5*time.Minute)
			if err != nil {
				continue // Keep trying
			}
			
			// Check if any log entry matches the expected pattern
			for _, logEntry := range logs {
				if strings.Contains(logEntry.Message, expectedPattern) {
					duration := time.Since(startTime)
					logResult("wait_for_logs", functionName, true, duration, nil)
					return logs, nil
				}
			}
		}
	}
}

// FilterLogs filters log entries based on a pattern.
// This is a utility function that doesn't follow the Terratest E pattern.
func FilterLogs(logs []LogEntry, pattern string) []LogEntry {
	var filtered []LogEntry
	
	for _, logEntry := range logs {
		if strings.Contains(logEntry.Message, pattern) {
			filtered = append(filtered, logEntry)
		}
	}
	
	return filtered
}

// FilterLogsByLevel filters log entries by log level.
// This is a utility function that doesn't follow the Terratest E pattern.
func FilterLogsByLevel(logs []LogEntry, level string) []LogEntry {
	var filtered []LogEntry
	
	for _, logEntry := range logs {
		if strings.EqualFold(logEntry.Level, level) {
			filtered = append(filtered, logEntry)
		}
	}
	
	return filtered
}

// GetLogStats calculates statistics from log entries.
// This is a utility function that returns structured information.
func GetLogStats(logs []LogEntry) LogStats {
	stats := LogStats{
		TotalEntries: len(logs),
		LevelCounts:  make(map[string]int),
	}
	
	if len(logs) == 0 {
		return stats
	}
	
	// Sort logs by timestamp to get earliest/latest
	sortedLogs := make([]LogEntry, len(logs))
	copy(sortedLogs, logs)
	sort.Slice(sortedLogs, func(i, j int) bool {
		return sortedLogs[i].Timestamp.Before(sortedLogs[j].Timestamp)
	})
	
	stats.EarliestEntry = sortedLogs[0].Timestamp
	stats.LatestEntry = sortedLogs[len(sortedLogs)-1].Timestamp
	stats.TimeSpan = stats.LatestEntry.Sub(stats.EarliestEntry)
	
	// Count entries by level
	for _, logEntry := range logs {
		if logEntry.Level != "" {
			stats.LevelCounts[logEntry.Level]++
		}
	}
	
	return stats
}

// LogStats provides statistics about log entries
type LogStats struct {
	TotalEntries  int
	EarliestEntry time.Time
	LatestEntry   time.Time
	TimeSpan      time.Duration
	LevelCounts   map[string]int
}

// mergeLogsOptions merges user options with defaults
func mergeLogsOptions(userOpts *LogsOptions) *LogsOptions {
	defaults := &LogsOptions{
		MaxLines:  1000,
		ParseLogs: true,
		Timeout:   30 * time.Second,
	}
	
	if userOpts == nil {
		return defaults
	}
	
	// Create a new merged options struct without modifying the input
	merged := &LogsOptions{
		StartTime:     userOpts.StartTime,
		EndTime:       userOpts.EndTime,
		FilterPattern: userOpts.FilterPattern,
		MaxLines:      userOpts.MaxLines,
		ParseLogs:     userOpts.ParseLogs,
		Timeout:       userOpts.Timeout,
	}
	
	// Apply defaults for zero values
	if merged.MaxLines == 0 {
		merged.MaxLines = defaults.MaxLines
	}
	if merged.Timeout == 0 {
		merged.Timeout = defaults.Timeout
	}
	
	// For ParseLogs, default to true (our default behavior)
	// Since Go bool zero value is false, we assume users want parsing by default
	// unless they explicitly set ParseLogs to false in their options
	if !merged.ParseLogs {
		merged.ParseLogs = defaults.ParseLogs
	}
	
	return merged
}

// retrieveCloudWatchLogs fetches log events from CloudWatch
func retrieveCloudWatchLogs(ctx *TestContext, logGroupName string, opts *LogsOptions) ([]types.FilteredLogEvent, error) {
	client := createCloudWatchLogsClient(ctx)
	
	// Create context with timeout
	logCtx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	
	// Build filter logs input
	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(logGroupName),
		Limit:        aws.Int32(opts.MaxLines),
	}
	
	if opts.StartTime != nil {
		input.StartTime = aws.Int64(opts.StartTime.UnixMilli())
	}
	
	if opts.EndTime != nil {
		input.EndTime = aws.Int64(opts.EndTime.UnixMilli())
	}
	
	if opts.FilterPattern != "" {
		input.FilterPattern = aws.String(opts.FilterPattern)
	}
	
	// Execute the request
	var allEvents []types.FilteredLogEvent
	var nextToken *string
	
	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}
		
		output, err := client.FilterLogEvents(logCtx, input)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrLogRetrievalFailed, err)
		}
		
		allEvents = append(allEvents, output.Events...)
		
		// Check if there are more events
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}
	
	return allEvents, nil
}

// convertLogEvents converts AWS log events to our LogEntry type
func convertLogEvents(events []types.FilteredLogEvent) []LogEntry {
	var logEntries []LogEntry
	
	for _, event := range events {
		entry := LogEntry{
			Timestamp: time.UnixMilli(*event.Timestamp),
			Message:   *event.Message,
		}
		
		logEntries = append(logEntries, entry)
	}
	
	return logEntries
}

// parseLogEvents parses AWS log events into structured LogEntry objects
func parseLogEvents(events []types.FilteredLogEvent) []LogEntry {
	var logEntries []LogEntry
	
	for _, event := range events {
		entry := LogEntry{
			Timestamp: time.UnixMilli(*event.Timestamp),
			Message:   *event.Message,
		}
		
		// Parse log level and request ID from Lambda log format
		message := *event.Message
		
		// Extract request ID from Lambda logs
		if requestID := extractRequestID(message); requestID != "" {
			entry.RequestID = requestID
		}
		
		// Extract log level
		if level := extractLogLevel(message); level != "" {
			entry.Level = level
		}
		
		logEntries = append(logEntries, entry)
	}
	
	return logEntries
}

// extractRequestID extracts AWS request ID from log message
func extractRequestID(message string) string {
	// Lambda logs often contain request IDs in specific formats
	patterns := []string{
		"RequestId: ",
		"REQUEST_ID:",
		"requestId: ",
	}
	
	for _, pattern := range patterns {
		if idx := strings.Index(message, pattern); idx != -1 {
			start := idx + len(pattern)
			end := strings.IndexAny(message[start:], " \t\n\r")
			if end == -1 {
				end = len(message) - start
			}
			return strings.TrimSpace(message[start : start+end])
		}
	}
	
	// Also check for Lambda log format: timestamp\trequest-id\tlevel\tmessage
	parts := strings.Split(message, "\t")
	if len(parts) >= 2 {
		// Second part might be a request ID (after timestamp)
		requestIDCandidate := strings.TrimSpace(parts[1])
		// Simple heuristic: request IDs are typically alphanumeric with hyphens and reasonable length
		if len(requestIDCandidate) > 5 && len(requestIDCandidate) < 50 {
			// Basic validation for request ID pattern (contains hyphens, alphanumeric)
			if strings.Contains(requestIDCandidate, "-") && isAlphanumericWithHyphens(requestIDCandidate) {
				return requestIDCandidate
			}
		}
	}
	
	return ""
}

// isAlphanumericWithHyphens checks if string contains only alphanumeric characters and hyphens
func isAlphanumericWithHyphens(s string) bool {
	for _, char := range s {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}
	return true
}

// extractLogLevel extracts log level from message
func extractLogLevel(message string) string {
	message = strings.ToUpper(message)
	
	// Check for Lambda-specific log patterns first (these are always INFO level)
	if strings.Contains(message, "REPORT REQUESTID:") {
		return "INFO"
	}
	
	if strings.Contains(message, "START REQUESTID:") {
		return "INFO"
	}
	
	if strings.Contains(message, "END REQUESTID:") {
		return "INFO"
	}
	
	// Check for standard log levels (order matters: more specific first)
	levels := []string{"ERROR", "WARN", "DEBUG", "TRACE", "INFO"}
	
	for _, level := range levels {
		if strings.Contains(message, level) {
			return level
		}
	}
	
	return ""
}

// LogsContain checks if any log entry contains the specified text.
// This is a utility function for test assertions.
func LogsContain(logs []LogEntry, text string) bool {
	for _, logEntry := range logs {
		if strings.Contains(logEntry.Message, text) {
			return true
		}
	}
	return false
}

// LogsContainLevel checks if logs contain entries at the specified level.
// This is a utility function for test assertions.
func LogsContainLevel(logs []LogEntry, level string) bool {
	for _, logEntry := range logs {
		if strings.EqualFold(logEntry.Level, level) {
			return true
		}
	}
	return false
}

// GetLogsByRequestID retrieves all log entries for a specific request ID.
// This is a utility function for debugging specific invocations.
func GetLogsByRequestID(logs []LogEntry, requestID string) []LogEntry {
	var filtered []LogEntry
	
	for _, logEntry := range logs {
		if logEntry.RequestID == requestID {
			filtered = append(filtered, logEntry)
		}
	}
	
	return filtered
}