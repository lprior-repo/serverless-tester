package lambda

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Setup helper for batch processing tests
func setupMockBatchProcessingTest() (*TestContext, *MockLambdaClient) {
	mockLambdaClient := &MockLambdaClient{}
	mockCloudWatchClient := &MockCloudWatchLogsClient{}
	
	mockFactory := &MockClientFactory{
		LambdaClient:     mockLambdaClient,
		CloudWatchClient: mockCloudWatchClient,
	}
	
	SetClientFactory(mockFactory)
	
	mockT := &mockAssertionTestingT{}
	ctx := &TestContext{T: mockT}
	
	return ctx, mockLambdaClient
}

// TestBatchProcessSQSMessages_WithValidMessages_ShouldProcessAll tests batch processing of SQS messages
func TestBatchProcessSQSMessages_WithValidMessages_ShouldProcessAll(t *testing.T) {
	// Create test context
	ctx, mockClient := setupMockBatchProcessingTest()
	defer teardownMockAssertionTest()
	
	// Create multiple SQS events
	messages := []string{
		"Message 1",
		"Message 2", 
		"Message 3",
	}
	
	sqsEvents := make([]string, len(messages))
	for i, msg := range messages {
		sqsEvents[i] = BuildSQSEvent("test-queue", msg)
	}
	
	// Set up mock response for successful processing
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode:      200,
		Payload:         []byte(`{"status": "processed"}`),
		ExecutedVersion: aws.String("$LATEST"),
		FunctionError:   nil,
	}
	mockClient.InvokeError = nil
	
	// Process messages in batch
	results, err := BatchProcessSQSMessagesE(ctx, "batch-processor", sqsEvents)
	
	// Assertions
	require.NoError(t, err)
	assert.Len(t, results, len(messages))
	
	// Verify all messages were processed successfully
	for _, result := range results {
		assert.Equal(t, int32(200), result.StatusCode)
		assert.NotEmpty(t, result.Payload)
		assert.Empty(t, result.FunctionError)
	}
}

// TestBatchProcessingConfig_Validation tests configuration validation
func TestBatchProcessingConfig_Validation_ShouldValidateCorrectly(t *testing.T) {
	tests := []struct {
		name        string
		config      BatchProcessingConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidConfig",
			config: BatchProcessingConfig{
				ConcurrencyLimit: 5,
				RetryAttempts:   3,
				RetryDelay:      100 * time.Millisecond,
			},
			expectError: false,
		},
		{
			name: "InvalidConcurrencyLimit",
			config: BatchProcessingConfig{
				ConcurrencyLimit: 0,
				RetryAttempts:   3,
				RetryDelay:      100 * time.Millisecond,
			},
			expectError: true,
			errorMsg:    "concurrency limit must be greater than 0",
		},
		{
			name: "InvalidRetryAttempts",
			config: BatchProcessingConfig{
				ConcurrencyLimit: 5,
				RetryAttempts:   -1,
				RetryDelay:      100 * time.Millisecond,
			},
			expectError: true,
			errorMsg:    "retry attempts must be non-negative",
		},
		{
			name: "InvalidRetryDelay",
			config: BatchProcessingConfig{
				ConcurrencyLimit: 5,
				RetryAttempts:   3,
				RetryDelay:      -1 * time.Millisecond,
			},
			expectError: true,
			errorMsg:    "retry delay must be non-negative",
		},
		{
			name: "ExcessiveConcurrency",
			config: BatchProcessingConfig{
				ConcurrencyLimit: 1001,
				RetryAttempts:   3,
				RetryDelay:      100 * time.Millisecond,
			},
			expectError: true,
			errorMsg:    "concurrency limit exceeds maximum allowed (1000)",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBatchProcessingConfig(tc.config)
			
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestDefaultBatchProcessingConfig_ShouldReturnValidDefaults tests default configuration
func TestDefaultBatchProcessingConfig_ShouldReturnValidDefaults(t *testing.T) {
	config := DefaultBatchProcessingConfig()
	
	assert.Equal(t, 10, config.ConcurrencyLimit)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, 100*time.Millisecond, config.RetryDelay)
	assert.False(t, config.CollectMetrics)
	
	// Verify configuration is valid
	err := validateBatchProcessingConfig(config)
	require.NoError(t, err)
}

// TestBatchProcessDynamoDBEvents_WithStreamRecords_ShouldProcessAll tests DynamoDB stream batch processing
func TestBatchProcessDynamoDBEvents_WithStreamRecords_ShouldProcessAll(t *testing.T) {
	ctx, mockClient := setupMockBatchProcessingTest()
	defer teardownMockAssertionTest()
	
	// Create DynamoDB events with different operation types
	events := []string{
		BuildDynamoDBEvent("users", "INSERT", map[string]interface{}{
			"id": map[string]interface{}{"S": "user1"},
		}),
		BuildDynamoDBEvent("users", "MODIFY", map[string]interface{}{
			"id": map[string]interface{}{"S": "user2"},
		}),
	}
	
	// Set up mock response
	mockClient.InvokeResponse = &lambda.InvokeOutput{
		StatusCode:      200,
		Payload:         []byte(`{"processed_record": 1}`),
		ExecutedVersion: aws.String("$LATEST"),
		FunctionError:   nil,
	}
	mockClient.InvokeError = nil
	
	// Process DynamoDB events in batch
	results, err := BatchProcessDynamoDBEventsE(ctx, "dynamodb-processor", events)
	
	require.NoError(t, err)
	assert.Len(t, results, 2)
	
	// Verify all events were processed
	for _, result := range results {
		assert.Equal(t, int32(200), result.StatusCode)
		assert.NotEmpty(t, result.Payload)
	}
}

// TestOptimalBatchSize_Calculation tests optimal batch size calculation
func TestOptimalBatchSize_Calculation_ShouldCalculateCorrectly(t *testing.T) {
	tests := []struct {
		name               string
		totalItems         int
		concurrencyLimit   int
		expectedBatchSize  int
		expectedBatchCount int
	}{
		{
			name:               "SmallBatch",
			totalItems:         10,
			concurrencyLimit:   3,
			expectedBatchSize:  4, // ceil(10/3) = 4
			expectedBatchCount: 3,
		},
		{
			name:               "LargeBatch",
			totalItems:         1000,
			concurrencyLimit:   10,
			expectedBatchSize:  100, // 1000/10 = 100
			expectedBatchCount: 10,
		},
		{
			name:               "ExactDivision",
			totalItems:         50,
			concurrencyLimit:   5,
			expectedBatchSize:  10, // 50/5 = 10
			expectedBatchCount: 5,
		},
		{
			name:               "SingleItem",
			totalItems:         1,
			concurrencyLimit:   5,
			expectedBatchSize:  1,
			expectedBatchCount: 1,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			batchSize, batchCount := calculateOptimalBatchSize(tc.totalItems, tc.concurrencyLimit)
			
			assert.Equal(t, tc.expectedBatchSize, batchSize)
			assert.Equal(t, tc.expectedBatchCount, batchCount)
			
			// Verify that batch size * batch count covers all items
			totalCovered := batchSize * batchCount
			assert.GreaterOrEqual(t, totalCovered, tc.totalItems)
		})
	}
}

// TestBatchProcessing_WithEmptyEventList_ShouldReturnEmptyResults tests handling of empty event lists
func TestBatchProcessing_WithEmptyEventList_ShouldReturnEmptyResults(t *testing.T) {
	ctx, _ := setupMockBatchProcessingTest()
	defer teardownMockAssertionTest()
	
	// Test with empty SQS events
	results, err := BatchProcessSQSMessagesE(ctx, "test-function", []string{})
	require.NoError(t, err)
	assert.Empty(t, results)
	
	// Test with empty DynamoDB events
	results2, err := BatchProcessDynamoDBEventsE(ctx, "test-function", []string{})
	require.NoError(t, err)
	assert.Empty(t, results2)
}

// TestBatchProcessing_WithInvalidFunctionName_ShouldReturnError tests validation of function name
func TestBatchProcessing_WithInvalidFunctionName_ShouldReturnError(t *testing.T) {
	ctx, _ := setupMockBatchProcessingTest()
	defer teardownMockAssertionTest()
	
	sqsEvent := BuildSQSEvent("test-queue", "test message")
	
	// Test with empty function name
	_, err := BatchProcessSQSMessagesE(ctx, "", []string{sqsEvent})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "function name cannot be empty")
	
	// Test with invalid function name format
	_, err = BatchProcessSQSMessagesE(ctx, "invalid name with spaces", []string{sqsEvent})
	require.Error(t, err)
}

// TestBatchMetrics_InitializationAndReset tests metrics initialization and reset
func TestBatchMetrics_InitializationAndReset_ShouldWorkCorrectly(t *testing.T) {
	// Initially no metrics
	metrics := GetBatchProcessingMetrics()
	assert.Nil(t, metrics)
	
	// Initialize metrics
	initializeBatchMetrics()
	metrics = GetBatchProcessingMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.TotalProcessed)
	assert.Equal(t, 0, metrics.SuccessfulProcessed)
	assert.Equal(t, 0, metrics.FailedProcessed)
	
	// Reset metrics
	ResetBatchProcessingMetrics()
	metrics = GetBatchProcessingMetrics()
	assert.Nil(t, metrics)
}

// TestCreateOptimizedBatchConfig_ShouldCreateValidConfiguration tests optimized configuration creation
func TestCreateOptimizedBatchConfig_ShouldCreateValidConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		eventCount     int
		targetLatency  time.Duration
		maxConcurrency int
	}{
		{
			name:           "SmallBatch",
			eventCount:     10,
			targetLatency:  1 * time.Second,
			maxConcurrency: 5,
		},
		{
			name:           "LargeBatch",
			eventCount:     100,
			targetLatency:  5 * time.Second,
			maxConcurrency: 20,
		},
		{
			name:           "SingleEvent",
			eventCount:     1,
			targetLatency:  500 * time.Millisecond,
			maxConcurrency: 10,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := CreateOptimizedBatchConfig(tc.eventCount, tc.targetLatency, tc.maxConcurrency)
			
			// Verify configuration is valid
			err := validateBatchProcessingConfig(config)
			require.NoError(t, err)
			
			// Verify constraints
			assert.LessOrEqual(t, config.ConcurrencyLimit, tc.maxConcurrency)
			assert.GreaterOrEqual(t, config.ConcurrencyLimit, 1)
			assert.True(t, config.CollectMetrics)
		})
	}
}

// TestAnalyzeBatchPerformance_WithValidMetrics_ShouldProvideAnalysis tests performance analysis
func TestAnalyzeBatchPerformance_WithValidMetrics_ShouldProvideAnalysis(t *testing.T) {
	// Create sample metrics
	metrics := &BatchProcessingMetrics{
		TotalProcessed:        100,
		SuccessfulProcessed:   95,
		FailedProcessed:       5,
		TotalProcessingTime:   10 * time.Second,
		AverageProcessingTime: 100 * time.Millisecond,
	}
	
	analysis := AnalyzeBatchPerformance(metrics, 100)
	
	require.NotNil(t, analysis)
	assert.Equal(t, 100, analysis["total_events"])
	assert.Equal(t, 0.95, analysis["success_rate"])
	assert.Equal(t, 0.05, analysis["failure_rate"])
	assert.Contains(t, analysis, "performance_grade")
	assert.Contains(t, analysis, "recommendation")
	assert.Contains(t, analysis, "throughput_events_per_second")
}

// TestAnalyzeBatchPerformance_WithNilMetrics_ShouldReturnError tests analysis with nil metrics
func TestAnalyzeBatchPerformance_WithNilMetrics_ShouldReturnError(t *testing.T) {
	analysis := AnalyzeBatchPerformance(nil, 100)
	
	require.NotNil(t, analysis)
	assert.Contains(t, analysis, "error")
	assert.Equal(t, "no metrics available for analysis", analysis["error"])
}