package lambda

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLogsWithOptionsE(t *testing.T) {
	testCases := []struct {
		name           string
		functionName   string
		options        *LogsOptions
		mockResponse   *cloudwatchlogs.FilterLogEventsOutput
		mockError      error
		expectedError  bool
		errorContains  string
		validateResult func(t *testing.T, logs []LogEntry)
	}{
		{
			name:         "should_retrieve_logs_successfully_with_default_options",
			functionName: "test-function",
			options:      nil, // Use defaults
			mockResponse: &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp: aws.Int64(time.Now().UnixMilli()),
						Message:   aws.String("INFO: Application started"),
					},
					{
						Timestamp: aws.Int64(time.Now().UnixMilli() + 1000),
						Message:   aws.String("DEBUG: Processing request"),
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, logs []LogEntry) {
				assert.Len(t, logs, 2)
				assert.Contains(t, logs[0].Message, "Application started")
				assert.Contains(t, logs[1].Message, "Processing request")
			},
		},
		{
			name:         "should_retrieve_logs_with_custom_options",
			functionName: "test-function",
			options: &LogsOptions{
				MaxLines:     100,
				FilterPattern: "ERROR",
				ParseLogs:    true,
				Timeout:      60 * time.Second,
			},
			mockResponse: &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp: aws.Int64(time.Now().UnixMilli()),
						Message:   aws.String("ERROR: Something went wrong RequestId: 12345-67890"),
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, logs []LogEntry) {
				assert.Len(t, logs, 1)
				assert.Contains(t, logs[0].Message, "Something went wrong")
				assert.Equal(t, "ERROR", logs[0].Level)
				assert.Equal(t, "12345-67890", logs[0].RequestID)
			},
		},
		{
			name:         "should_retrieve_logs_without_parsing",
			functionName: "test-function",
			options: &LogsOptions{
				ParseLogs: false,
			},
			mockResponse: &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp: aws.Int64(time.Now().UnixMilli()),
						Message:   aws.String("Raw log message"),
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, logs []LogEntry) {
				assert.Len(t, logs, 1)
				assert.Contains(t, logs[0].Message, "Raw log message")
				assert.Empty(t, logs[0].Level)    // Not parsed
				assert.Empty(t, logs[0].RequestID) // Not parsed
			},
		},
		{
			name:          "should_return_error_for_invalid_function_name",
			functionName:  "invalid@function",
			expectedError: true,
			errorContains: "invalid character '@'",
		},
		{
			name:         "should_return_error_when_filter_log_events_fails",
			functionName: "test-function",
			mockError:    errors.New("ResourceNotFoundException: The specified log group does not exist"),
			expectedError: true,
			errorContains: "failed to retrieve lambda logs",
		},
		{
			name:         "should_handle_empty_log_result",
			functionName: "test-function",
			mockResponse: &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{},
			},
			expectedError: false,
			validateResult: func(t *testing.T, logs []LogEntry) {
				assert.Len(t, logs, 0)
			},
		},
		{
			name:         "should_handle_time_range_filtering",
			functionName: "test-function",
			options: &LogsOptions{
				StartTime: aws.Time(time.Now().Add(-1 * time.Hour)),
				EndTime:   aws.Time(time.Now()),
			},
			mockResponse: &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp: aws.Int64(time.Now().Add(-30*time.Minute).UnixMilli()),
						Message:   aws.String("Log within time range"),
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, logs []LogEntry) {
				assert.Len(t, logs, 1)
				assert.Contains(t, logs[0].Message, "within time range")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockCwClient := &MockCloudWatchLogsClient{
				FilterLogEventsResponse: tc.mockResponse,
				FilterLogEventsError:    tc.mockError,
			}
			mockFactory := &MockClientFactory{
				CloudWatchClient: mockCwClient,
			}
			originalFactory := globalClientFactory
			SetClientFactory(mockFactory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: &mockTestingT{}}

			// Act
			result, err := GetLogsWithOptionsE(ctx, tc.functionName, tc.options)

			// Assert
			if tc.expectedError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tc.validateResult != nil {
					tc.validateResult(t, result)
				}
			}
		})
	}
}

func TestGetLogsE(t *testing.T) {
	t.Run("should_call_get_logs_with_options_e_with_nil_options", func(t *testing.T) {
		// Arrange
		mockCwClient := &MockCloudWatchLogsClient{
			FilterLogEventsResponse: &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp: aws.Int64(time.Now().UnixMilli()),
						Message:   aws.String("Test log message"),
					},
				},
			},
		}
		mockFactory := &MockClientFactory{
			CloudWatchClient: mockCwClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		result, err := GetLogsE(ctx, "test-function")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Contains(t, result[0].Message, "Test log message")
	})
}

func TestGetLogs(t *testing.T) {
	t.Run("should_call_get_logs_e_and_fail_test_on_error", func(t *testing.T) {
		// Arrange
		mockCwClient := &MockCloudWatchLogsClient{
			FilterLogEventsError: errors.New("log retrieval failed"),
		}
		mockFactory := &MockClientFactory{
			CloudWatchClient: mockCwClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		mockT := &mockTestingT{}
		ctx := &TestContext{T: mockT}

		// Act
		result := GetLogs(ctx, "test-function")

		// Assert
		assert.Nil(t, result)
		assert.True(t, mockT.failNowCalled)
		assert.True(t, mockT.errorCalled)
	})
}

func TestGetRecentLogsE(t *testing.T) {
	t.Run("should_retrieve_recent_logs_with_time_filter", func(t *testing.T) {
		// Arrange
		mockCwClient := &MockCloudWatchLogsClient{
			FilterLogEventsResponse: &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp: aws.Int64(time.Now().Add(-5*time.Minute).UnixMilli()),
						Message:   aws.String("Recent log entry"),
					},
				},
			},
		}
		mockFactory := &MockClientFactory{
			CloudWatchClient: mockCwClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		result, err := GetRecentLogsE(ctx, "test-function", 10*time.Minute)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Contains(t, result[0].Message, "Recent log entry")
	})
}

func TestWaitForLogsE(t *testing.T) {
	t.Run("should_return_logs_when_pattern_found", func(t *testing.T) {
		// Arrange
		mockCwClient := &MockCloudWatchLogsClient{
			FilterLogEventsResponse: &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp: aws.Int64(time.Now().UnixMilli()),
						Message:   aws.String("SUCCESS: Operation completed"),
					},
				},
			},
		}
		mockFactory := &MockClientFactory{
			CloudWatchClient: mockCwClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		result, err := WaitForLogsE(ctx, "test-function", "SUCCESS", 5*time.Second)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Contains(t, result[0].Message, "SUCCESS")
	})

	t.Run("should_timeout_when_pattern_not_found", func(t *testing.T) {
		// Arrange
		mockCwClient := &MockCloudWatchLogsClient{
			FilterLogEventsResponse: &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp: aws.Int64(time.Now().UnixMilli()),
						Message:   aws.String("INFO: Normal operation"),
					},
				},
			},
		}
		mockFactory := &MockClientFactory{
			CloudWatchClient: mockCwClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		start := time.Now()
		result, err := WaitForLogsE(ctx, "test-function", "SUCCESS", 100*time.Millisecond)
		duration := time.Since(start)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "timeout waiting for logs with pattern: SUCCESS")
		assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
		assert.Less(t, duration, 500*time.Millisecond)
	})

	t.Run("should_continue_when_get_logs_fails_temporarily", func(t *testing.T) {
		// Arrange - This test simulates temporary failures followed by success
		mockCwClient := &MockCloudWatchLogsClient{
			FilterLogEventsError: errors.New("temporary failure"),
		}
		mockFactory := &MockClientFactory{
			CloudWatchClient: mockCwClient,
		}
		originalFactory := globalClientFactory
		SetClientFactory(mockFactory)
		defer SetClientFactory(originalFactory)

		ctx := &TestContext{T: &mockTestingT{}}

		// Act
		result, err := WaitForLogsE(ctx, "test-function", "SUCCESS", 100*time.Millisecond)

		// Assert - Should timeout since we never return success
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "timeout waiting for logs")
	})
}

func TestFilterLogs(t *testing.T) {
	testCases := []struct {
		name     string
		logs     []LogEntry
		pattern  string
		expected []LogEntry
	}{
		{
			name: "should_filter_logs_by_pattern",
			logs: []LogEntry{
				{Message: "INFO: Application started"},
				{Message: "ERROR: Something went wrong"},
				{Message: "INFO: Processing complete"},
				{Message: "ERROR: Another error occurred"},
			},
			pattern: "ERROR",
			expected: []LogEntry{
				{Message: "ERROR: Something went wrong"},
				{Message: "ERROR: Another error occurred"},
			},
		},
		{
			name: "should_return_empty_when_no_matches",
			logs: []LogEntry{
				{Message: "INFO: Application started"},
				{Message: "DEBUG: Processing request"},
			},
			pattern: "ERROR",
			expected: []LogEntry{},
		},
		{
			name:     "should_handle_empty_log_list",
			logs:     []LogEntry{},
			pattern:  "ERROR",
			expected: []LogEntry{},
		},
		{
			name: "should_match_partial_strings",
			logs: []LogEntry{
				{Message: "Request processed successfully"},
				{Message: "Failed to process request"},
				{Message: "Processing in progress"},
			},
			pattern: "process",
			expected: []LogEntry{
				{Message: "Request processed successfully"},
				{Message: "Failed to process request"},
				{Message: "Processing in progress"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := FilterLogs(tc.logs, tc.pattern)

			// Assert
			assert.Len(t, result, len(tc.expected))
			for i, expectedLog := range tc.expected {
				assert.Equal(t, expectedLog.Message, result[i].Message)
			}
		})
	}
}

func TestFilterLogsByLevel(t *testing.T) {
	testCases := []struct {
		name     string
		logs     []LogEntry
		level    string
		expected []LogEntry
	}{
		{
			name: "should_filter_logs_by_level",
			logs: []LogEntry{
				{Message: "Info message", Level: "INFO"},
				{Message: "Error message", Level: "ERROR"},
				{Message: "Debug message", Level: "DEBUG"},
				{Message: "Another error", Level: "ERROR"},
			},
			level: "ERROR",
			expected: []LogEntry{
				{Message: "Error message", Level: "ERROR"},
				{Message: "Another error", Level: "ERROR"},
			},
		},
		{
			name: "should_handle_case_insensitive_matching",
			logs: []LogEntry{
				{Message: "Info message", Level: "info"},
				{Message: "Error message", Level: "ERROR"},
				{Message: "Another info", Level: "Info"},
			},
			level: "INFO",
			expected: []LogEntry{
				{Message: "Info message", Level: "info"},
				{Message: "Another info", Level: "Info"},
			},
		},
		{
			name: "should_return_empty_when_no_matches",
			logs: []LogEntry{
				{Message: "Info message", Level: "INFO"},
				{Message: "Debug message", Level: "DEBUG"},
			},
			level: "ERROR",
			expected: []LogEntry{},
		},
		{
			name: "should_skip_logs_with_empty_level",
			logs: []LogEntry{
				{Message: "Message without level", Level: ""},
				{Message: "Error message", Level: "ERROR"},
				{Message: "Another message without level"},
			},
			level: "ERROR",
			expected: []LogEntry{
				{Message: "Error message", Level: "ERROR"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := FilterLogsByLevel(tc.logs, tc.level)

			// Assert
			assert.Len(t, result, len(tc.expected))
			for i, expectedLog := range tc.expected {
				assert.Equal(t, expectedLog.Message, result[i].Message)
				assert.Equal(t, expectedLog.Level, result[i].Level)
			}
		})
	}
}

func TestGetLogStats(t *testing.T) {
	testCases := []struct {
		name           string
		logs           []LogEntry
		validateStats  func(t *testing.T, stats LogStats)
	}{
		{
			name: "should_calculate_stats_for_multiple_logs",
			logs: []LogEntry{
				{
					Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Message:   "First log",
					Level:     "INFO",
				},
				{
					Timestamp: time.Date(2023, 1, 1, 12, 1, 0, 0, time.UTC),
					Message:   "Second log",
					Level:     "ERROR",
				},
				{
					Timestamp: time.Date(2023, 1, 1, 12, 2, 0, 0, time.UTC),
					Message:   "Third log",
					Level:     "INFO",
				},
			},
			validateStats: func(t *testing.T, stats LogStats) {
				assert.Equal(t, 3, stats.TotalEntries)
				assert.Equal(t, time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), stats.EarliestEntry)
				assert.Equal(t, time.Date(2023, 1, 1, 12, 2, 0, 0, time.UTC), stats.LatestEntry)
				assert.Equal(t, 2*time.Minute, stats.TimeSpan)
				assert.Equal(t, 2, stats.LevelCounts["INFO"])
				assert.Equal(t, 1, stats.LevelCounts["ERROR"])
			},
		},
		{
			name: "should_handle_single_log_entry",
			logs: []LogEntry{
				{
					Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Message:   "Single log",
					Level:     "DEBUG",
				},
			},
			validateStats: func(t *testing.T, stats LogStats) {
				assert.Equal(t, 1, stats.TotalEntries)
				assert.Equal(t, time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), stats.EarliestEntry)
				assert.Equal(t, time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), stats.LatestEntry)
				assert.Equal(t, time.Duration(0), stats.TimeSpan)
				assert.Equal(t, 1, stats.LevelCounts["DEBUG"])
			},
		},
		{
			name: "should_handle_empty_log_list",
			logs: []LogEntry{},
			validateStats: func(t *testing.T, stats LogStats) {
				assert.Equal(t, 0, stats.TotalEntries)
				assert.Equal(t, time.Time{}, stats.EarliestEntry)
				assert.Equal(t, time.Time{}, stats.LatestEntry)
				assert.Equal(t, time.Duration(0), stats.TimeSpan)
				assert.Empty(t, stats.LevelCounts)
			},
		},
		{
			name: "should_handle_logs_without_levels",
			logs: []LogEntry{
				{
					Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Message:   "Log without level",
					Level:     "",
				},
				{
					Timestamp: time.Date(2023, 1, 1, 12, 1, 0, 0, time.UTC),
					Message:   "Another log without level",
				},
			},
			validateStats: func(t *testing.T, stats LogStats) {
				assert.Equal(t, 2, stats.TotalEntries)
				assert.Empty(t, stats.LevelCounts) // No levels to count
			},
		},
		{
			name: "should_sort_logs_by_timestamp_correctly",
			logs: []LogEntry{
				{
					Timestamp: time.Date(2023, 1, 1, 12, 2, 0, 0, time.UTC), // Latest
					Message:   "Third chronologically",
					Level:     "INFO",
				},
				{
					Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), // Earliest
					Message:   "First chronologically",
					Level:     "DEBUG",
				},
				{
					Timestamp: time.Date(2023, 1, 1, 12, 1, 0, 0, time.UTC), // Middle
					Message:   "Second chronologically",
					Level:     "WARN",
				},
			},
			validateStats: func(t *testing.T, stats LogStats) {
				assert.Equal(t, 3, stats.TotalEntries)
				assert.Equal(t, time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), stats.EarliestEntry)
				assert.Equal(t, time.Date(2023, 1, 1, 12, 2, 0, 0, time.UTC), stats.LatestEntry)
				assert.Equal(t, 2*time.Minute, stats.TimeSpan)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			stats := GetLogStats(tc.logs)

			// Assert
			tc.validateStats(t, stats)
		})
	}
}

func TestMergeLogsOptions(t *testing.T) {
	testCases := []struct {
		name         string
		userOpts     *LogsOptions
		validateOpts func(t *testing.T, opts *LogsOptions)
	}{
		{
			name:     "should_return_defaults_for_nil_options",
			userOpts: nil,
			validateOpts: func(t *testing.T, opts *LogsOptions) {
				assert.Equal(t, int32(1000), opts.MaxLines)
				assert.True(t, opts.ParseLogs)
				assert.Equal(t, 30*time.Second, opts.Timeout)
			},
		},
		{
			name: "should_preserve_user_values_and_use_defaults_for_unset",
			userOpts: &LogsOptions{
				MaxLines:     500,
				FilterPattern: "ERROR",
				// ParseLogs and Timeout not set - should use defaults
			},
			validateOpts: func(t *testing.T, opts *LogsOptions) {
				assert.Equal(t, int32(500), opts.MaxLines)
				assert.Equal(t, "ERROR", opts.FilterPattern)
				assert.True(t, opts.ParseLogs)     // Default
				assert.Equal(t, 30*time.Second, opts.Timeout) // Default
			},
		},
		{
			name: "should_preserve_all_user_values",
			userOpts: &LogsOptions{
				StartTime:     aws.Time(time.Now().Add(-1 * time.Hour)),
				EndTime:       aws.Time(time.Now()),
				MaxLines:      200,
				FilterPattern: "DEBUG",
				ParseLogs:     false,
				Timeout:       60 * time.Second,
			},
			validateOpts: func(t *testing.T, opts *LogsOptions) {
				assert.Equal(t, int32(200), opts.MaxLines)
				assert.Equal(t, "DEBUG", opts.FilterPattern)
				assert.False(t, opts.ParseLogs)
				assert.Equal(t, 60*time.Second, opts.Timeout)
				assert.NotNil(t, opts.StartTime)
				assert.NotNil(t, opts.EndTime)
			},
		},
		{
			name: "should_use_defaults_for_zero_values",
			userOpts: &LogsOptions{
				MaxLines: 0,        // Should use default
				Timeout:  0,        // Should use default
				ParseLogs: false,   // Explicit false should be preserved
			},
			validateOpts: func(t *testing.T, opts *LogsOptions) {
				assert.Equal(t, int32(1000), opts.MaxLines)     // Default
				assert.False(t, opts.ParseLogs)                 // Preserved user value
				assert.Equal(t, 30*time.Second, opts.Timeout)   // Default
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := mergeLogsOptions(tc.userOpts)

			// Assert
			tc.validateOpts(t, result)
		})
	}
}

func TestExtractRequestID(t *testing.T) {
	testCases := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "should_extract_request_id_with_RequestId_colon_space",
			message:  "START RequestId: 12345-67890-abcdef Version: $LATEST",
			expected: "12345-67890-abcdef",
		},
		{
			name:     "should_extract_request_id_with_REQUEST_ID_colon",
			message:  "Function executed REQUEST_ID:abcdef-12345-67890",
			expected: "abcdef-12345-67890",
		},
		{
			name:     "should_extract_request_id_with_requestId_colon_space",
			message:  "Processing complete requestId: xyz-123-abc Duration: 100ms",
			expected: "xyz-123-abc",
		},
		{
			name:     "should_return_empty_when_no_request_id_found",
			message:  "Simple log message without request ID",
			expected: "",
		},
		{
			name:     "should_handle_request_id_at_end_of_message",
			message:  "Log message RequestId: final-request-id",
			expected: "final-request-id",
		},
		{
			name:     "should_extract_first_request_id_when_multiple_patterns",
			message:  "RequestId: first-id and REQUEST_ID:second-id",
			expected: "first-id",
		},
		{
			name:     "should_handle_empty_message",
			message:  "",
			expected: "",
		},
		{
			name:     "should_handle_request_id_with_tabs_and_newlines",
			message:  "START RequestId: 12345-67890-abcdef\tVersion: $LATEST\n",
			expected: "12345-67890-abcdef",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := extractRequestID(tc.message)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractLogLevel(t *testing.T) {
	testCases := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "should_extract_error_level",
			message:  "ERROR: Something went wrong",
			expected: "ERROR",
		},
		{
			name:     "should_extract_warn_level",
			message:  "WARN: This is a warning message",
			expected: "WARN",
		},
		{
			name:     "should_extract_info_level",
			message:  "INFO: Application started successfully",
			expected: "INFO",
		},
		{
			name:     "should_extract_debug_level",
			message:  "DEBUG: Processing request with parameters",
			expected: "DEBUG",
		},
		{
			name:     "should_extract_trace_level",
			message:  "TRACE: Detailed execution trace",
			expected: "TRACE",
		},
		{
			name:     "should_handle_case_insensitive_matching",
			message:  "error: lowercase error message",
			expected: "ERROR",
		},
		{
			name:     "should_extract_level_from_middle_of_message",
			message:  "Application encountered an ERROR during processing",
			expected: "ERROR",
		},
		{
			name:     "should_return_info_for_lambda_report_line",
			message:  "REPORT RequestId: 12345 Duration: 100ms Billed Duration: 100ms Memory Size: 128 MB",
			expected: "INFO",
		},
		{
			name:     "should_return_info_for_lambda_start_line",
			message:  "START RequestId: 12345 Version: $LATEST",
			expected: "INFO",
		},
		{
			name:     "should_return_info_for_lambda_end_line",
			message:  "END RequestId: 12345",
			expected: "INFO",
		},
		{
			name:     "should_return_empty_for_no_recognizable_level",
			message:  "Simple log message without level",
			expected: "",
		},
		{
			name:     "should_return_first_found_level",
			message:  "ERROR occurred during INFO processing",
			expected: "ERROR", // ERROR appears first
		},
		{
			name:     "should_handle_empty_message",
			message:  "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := extractLogLevel(tc.message)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestLogsContain(t *testing.T) {
	testCases := []struct {
		name     string
		logs     []LogEntry
		text     string
		expected bool
	}{
		{
			name: "should_return_true_when_text_found",
			logs: []LogEntry{
				{Message: "INFO: Application started"},
				{Message: "ERROR: Something went wrong"},
				{Message: "INFO: Processing complete"},
			},
			text:     "Something went wrong",
			expected: true,
		},
		{
			name: "should_return_false_when_text_not_found",
			logs: []LogEntry{
				{Message: "INFO: Application started"},
				{Message: "DEBUG: Processing request"},
			},
			text:     "ERROR",
			expected: false,
		},
		{
			name:     "should_return_false_for_empty_log_list",
			logs:     []LogEntry{},
			text:     "test",
			expected: false,
		},
		{
			name: "should_match_partial_text",
			logs: []LogEntry{
				{Message: "Processing request successfully"},
			},
			text:     "success",
			expected: true,
		},
		{
			name: "should_be_case_sensitive",
			logs: []LogEntry{
				{Message: "INFO: Application started"},
			},
			text:     "application", // lowercase, should not match "Application"
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := LogsContain(tc.logs, tc.text)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestLogsContainLevel(t *testing.T) {
	testCases := []struct {
		name     string
		logs     []LogEntry
		level    string
		expected bool
	}{
		{
			name: "should_return_true_when_level_found",
			logs: []LogEntry{
				{Message: "Info message", Level: "INFO"},
				{Message: "Error message", Level: "ERROR"},
			},
			level:    "ERROR",
			expected: true,
		},
		{
			name: "should_return_false_when_level_not_found",
			logs: []LogEntry{
				{Message: "Info message", Level: "INFO"},
				{Message: "Debug message", Level: "DEBUG"},
			},
			level:    "ERROR",
			expected: false,
		},
		{
			name:     "should_return_false_for_empty_log_list",
			logs:     []LogEntry{},
			level:    "ERROR",
			expected: false,
		},
		{
			name: "should_handle_case_insensitive_matching",
			logs: []LogEntry{
				{Message: "Error message", Level: "error"},
			},
			level:    "ERROR",
			expected: true,
		},
		{
			name: "should_return_false_for_empty_level_in_logs",
			logs: []LogEntry{
				{Message: "Message without level", Level: ""},
				{Message: "Another message"},
			},
			level:    "INFO",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := LogsContainLevel(tc.logs, tc.level)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetLogsByRequestID(t *testing.T) {
	testCases := []struct {
		name      string
		logs      []LogEntry
		requestID string
		expected  []LogEntry
	}{
		{
			name: "should_return_logs_with_matching_request_id",
			logs: []LogEntry{
				{Message: "Start", RequestID: "req-123"},
				{Message: "Processing", RequestID: "req-123"},
				{Message: "Different request", RequestID: "req-456"},
				{Message: "End", RequestID: "req-123"},
			},
			requestID: "req-123",
			expected: []LogEntry{
				{Message: "Start", RequestID: "req-123"},
				{Message: "Processing", RequestID: "req-123"},
				{Message: "End", RequestID: "req-123"},
			},
		},
		{
			name: "should_return_empty_when_no_matches",
			logs: []LogEntry{
				{Message: "Log 1", RequestID: "req-123"},
				{Message: "Log 2", RequestID: "req-456"},
			},
			requestID: "req-789",
			expected:  []LogEntry{},
		},
		{
			name:      "should_handle_empty_log_list",
			logs:      []LogEntry{},
			requestID: "req-123",
			expected:  []LogEntry{},
		},
		{
			name: "should_handle_logs_without_request_id",
			logs: []LogEntry{
				{Message: "Log without request ID", RequestID: ""},
				{Message: "Log with request ID", RequestID: "req-123"},
			},
			requestID: "req-123",
			expected: []LogEntry{
				{Message: "Log with request ID", RequestID: "req-123"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := GetLogsByRequestID(tc.logs, tc.requestID)

			// Assert
			assert.Len(t, result, len(tc.expected))
			for i, expectedLog := range tc.expected {
				assert.Equal(t, expectedLog.Message, result[i].Message)
				assert.Equal(t, expectedLog.RequestID, result[i].RequestID)
			}
		})
	}
}

func TestConvertLogEvents(t *testing.T) {
	now := time.Now()
	
	testCases := []struct {
		name           string
		events         []types.FilteredLogEvent
		validateResult func(t *testing.T, logs []LogEntry)
	}{
		{
			name: "should_convert_aws_log_events_to_log_entries",
			events: []types.FilteredLogEvent{
				{
					Timestamp: aws.Int64(now.UnixMilli()),
					Message:   aws.String("First log message"),
				},
				{
					Timestamp: aws.Int64(now.Add(time.Second).UnixMilli()),
					Message:   aws.String("Second log message"),
				},
			},
			validateResult: func(t *testing.T, logs []LogEntry) {
				assert.Len(t, logs, 2)
				assert.Equal(t, "First log message", logs[0].Message)
				assert.Equal(t, "Second log message", logs[1].Message)
				assert.WithinDuration(t, now, logs[0].Timestamp, time.Millisecond)
				assert.WithinDuration(t, now.Add(time.Second), logs[1].Timestamp, time.Millisecond)
			},
		},
		{
			name:   "should_handle_empty_events_list",
			events: []types.FilteredLogEvent{},
			validateResult: func(t *testing.T, logs []LogEntry) {
				assert.Len(t, logs, 0)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := convertLogEvents(tc.events)

			// Assert
			tc.validateResult(t, result)
		})
	}
}

func TestParseLogEvents(t *testing.T) {
	now := time.Now()
	
	testCases := []struct {
		name           string
		events         []types.FilteredLogEvent
		validateResult func(t *testing.T, logs []LogEntry)
	}{
		{
			name: "should_parse_log_events_with_request_id_and_level",
			events: []types.FilteredLogEvent{
				{
					Timestamp: aws.Int64(now.UnixMilli()),
					Message:   aws.String("ERROR: Something went wrong RequestId: 12345-67890-abcdef"),
				},
				{
					Timestamp: aws.Int64(now.Add(time.Second).UnixMilli()),
					Message:   aws.String("INFO: Processing complete requestId: xyz-123-abc"),
				},
			},
			validateResult: func(t *testing.T, logs []LogEntry) {
				assert.Len(t, logs, 2)
				
				// First log
				assert.Equal(t, "ERROR: Something went wrong RequestId: 12345-67890-abcdef", logs[0].Message)
				assert.Equal(t, "ERROR", logs[0].Level)
				assert.Equal(t, "12345-67890-abcdef", logs[0].RequestID)
				assert.WithinDuration(t, now, logs[0].Timestamp, time.Millisecond)
				
				// Second log
				assert.Equal(t, "INFO: Processing complete requestId: xyz-123-abc", logs[1].Message)
				assert.Equal(t, "INFO", logs[1].Level)
				assert.Equal(t, "xyz-123-abc", logs[1].RequestID)
			},
		},
		{
			name: "should_handle_logs_without_parsing_information",
			events: []types.FilteredLogEvent{
				{
					Timestamp: aws.Int64(now.UnixMilli()),
					Message:   aws.String("Simple log message without level or request ID"),
				},
			},
			validateResult: func(t *testing.T, logs []LogEntry) {
				assert.Len(t, logs, 1)
				assert.Equal(t, "Simple log message without level or request ID", logs[0].Message)
				assert.Empty(t, logs[0].Level)
				assert.Empty(t, logs[0].RequestID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := parseLogEvents(tc.events)

			// Assert
			tc.validateResult(t, result)
		})
	}
}