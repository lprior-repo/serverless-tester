package lambda

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateEventSourceMappingWithBatchConfig_SQS tests SQS-specific batch configuration
func TestCreateEventSourceMappingWithBatchConfig_SQS(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}

	// Setup expected response
	expectedUUID := "test-uuid-123"
	mockClient.CreateEventSourceMappingResponse = &lambda.CreateEventSourceMappingOutput{
		UUID:          aws.String(expectedUUID),
		EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
		FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
		State:         aws.String("Enabled"),
		BatchSize:     aws.Int32(50),
		MaximumBatchingWindowInSeconds: aws.Int32(5),
	}

	// Test SQS batch configuration
	batchConfig := BatchConfigurationOptions{
		BatchSize:                        50,
		MaximumBatchingWindowInSeconds:   5,
	}

	result := CreateEventSourceMappingWithBatchConfig(ctx, "test-function", 
		"arn:aws:sqs:us-east-1:123456789012:test-queue", batchConfig)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, expectedUUID, result.UUID)
	assert.Equal(t, int32(50), result.BatchSize)
	assert.Equal(t, int32(5), result.MaximumBatchingWindowInSeconds)

	// Verify the call was made correctly
	require.Len(t, mockClient.CreateEventSourceMappingCalls, 1)
	call := mockClient.CreateEventSourceMappingCalls[0]
	assert.Equal(t, "test-function", call.FunctionName)
	assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:test-queue", call.EventSourceArn)
	assert.Equal(t, int32(50), *call.BatchSize)
	assert.True(t, *call.Enabled)
}

// TestCreateEventSourceMappingWithBatchConfig_Kinesis tests Kinesis-specific batch configuration
func TestCreateEventSourceMappingWithBatchConfig_Kinesis(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}

	// Setup expected response
	expectedUUID := "test-uuid-456"
	mockClient.CreateEventSourceMappingResponse = &lambda.CreateEventSourceMappingOutput{
		UUID:          aws.String(expectedUUID),
		EventSourceArn: aws.String("arn:aws:kinesis:us-east-1:123456789012:stream/test-stream"),
		FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
		State:         aws.String("Enabled"),
		BatchSize:     aws.Int32(200),
		MaximumBatchingWindowInSeconds: aws.Int32(10),
	}

	// Test Kinesis batch configuration with parallelization
	batchConfig := BatchConfigurationOptions{
		BatchSize:                        200,
		MaximumBatchingWindowInSeconds:   10,
		ParallelizationFactor:           50,
	}

	result := CreateEventSourceMappingWithBatchConfig(ctx, "test-function", 
		"arn:aws:kinesis:us-east-1:123456789012:stream/test-stream", batchConfig)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, expectedUUID, result.UUID)
	assert.Equal(t, int32(200), result.BatchSize)
	assert.Equal(t, int32(10), result.MaximumBatchingWindowInSeconds)

	// Verify the call was made correctly
	require.Len(t, mockClient.CreateEventSourceMappingCalls, 1)
	call := mockClient.CreateEventSourceMappingCalls[0]
	assert.Equal(t, "test-function", call.FunctionName)
	assert.Equal(t, "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream", call.EventSourceArn)
}

// TestConfigureEventSourceMappingForThroughput tests throughput optimization
func TestConfigureEventSourceMappingForThroughput(t *testing.T) {
	testCases := []struct {
		name                    string
		eventSourceArn          string
		expectedBatchSize       int32
		expectedBatchingWindow  int32
	}{
		{
			name:                   "SQS throughput optimization",
			eventSourceArn:         "arn:aws:sqs:us-east-1:123456789012:high-throughput-queue",
			expectedBatchSize:      1000, // Max batch size for SQS
			expectedBatchingWindow: 5,    // Buffer for full batches
		},
		{
			name:                   "Kinesis throughput optimization",
			eventSourceArn:         "arn:aws:kinesis:us-east-1:123456789012:stream/high-volume-stream",
			expectedBatchSize:      10000, // Max batch size for Kinesis
			expectedBatchingWindow: 0,     // No buffering for max throughput
		},
		{
			name:                   "DynamoDB throughput optimization",
			eventSourceArn:         "arn:aws:dynamodb:us-east-1:123456789012:table/Users/stream/2023-01-01T00:00:00.000",
			expectedBatchSize:      1000, // Max batch size for DynamoDB
			expectedBatchingWindow: 0,    // No buffering
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockLambdaClient{}
			factory := &MockClientFactory{LambdaClient: mockClient}
			originalFactory := globalClientFactory
			SetClientFactory(factory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: t}

			// Setup expected response with throughput-optimized configuration
			mockClient.CreateEventSourceMappingResponse = &lambda.CreateEventSourceMappingOutput{
				UUID:          aws.String("throughput-uuid-123"),
				EventSourceArn: aws.String(tc.eventSourceArn),
				FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:throughput-function"),
				State:         aws.String("Enabled"),
				BatchSize:     aws.Int32(tc.expectedBatchSize),
				MaximumBatchingWindowInSeconds: aws.Int32(tc.expectedBatchingWindow),
			}

			result := ConfigureEventSourceMappingForThroughput(ctx, "throughput-function", tc.eventSourceArn)

			// Assert throughput optimization worked
			require.NotNil(t, result)
			assert.Equal(t, "throughput-uuid-123", result.UUID)
			assert.Equal(t, tc.expectedBatchSize, result.BatchSize)
			assert.Equal(t, tc.expectedBatchingWindow, result.MaximumBatchingWindowInSeconds)

			// Verify correct batch configuration was sent
			require.Len(t, mockClient.CreateEventSourceMappingCalls, 1)
			call := mockClient.CreateEventSourceMappingCalls[0]
			assert.Equal(t, "throughput-function", call.FunctionName)
			assert.Equal(t, tc.eventSourceArn, call.EventSourceArn)
			assert.Equal(t, tc.expectedBatchSize, *call.BatchSize)
		})
	}
}

// TestConfigureEventSourceMappingForLatency tests latency optimization
func TestConfigureEventSourceMappingForLatency(t *testing.T) {
	testCases := []struct {
		name                    string
		eventSourceArn          string
		expectedBatchSize       int32
		expectedBatchingWindow  int32
	}{
		{
			name:                   "SQS latency optimization",
			eventSourceArn:         "arn:aws:sqs:us-east-1:123456789012:low-latency-queue",
			expectedBatchSize:      1, // Min batch size for lowest latency
			expectedBatchingWindow: 0, // No buffering
		},
		{
			name:                   "Kinesis latency optimization",
			eventSourceArn:         "arn:aws:kinesis:us-east-1:123456789012:stream/real-time-stream",
			expectedBatchSize:      1, // Min batch size
			expectedBatchingWindow: 0, // No buffering
		},
		{
			name:                   "DynamoDB latency optimization",
			eventSourceArn:         "arn:aws:dynamodb:us-east-1:123456789012:table/RealTimeData/stream/2023-01-01T00:00:00.000",
			expectedBatchSize:      1, // Min batch size
			expectedBatchingWindow: 0, // No buffering
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockLambdaClient{}
			factory := &MockClientFactory{LambdaClient: mockClient}
			originalFactory := globalClientFactory
			SetClientFactory(factory)
			defer SetClientFactory(originalFactory)

			ctx := &TestContext{T: t}

			// Setup expected response with latency-optimized configuration
			mockClient.CreateEventSourceMappingResponse = &lambda.CreateEventSourceMappingOutput{
				UUID:          aws.String("latency-uuid-123"),
				EventSourceArn: aws.String(tc.eventSourceArn),
				FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:latency-function"),
				State:         aws.String("Enabled"),
				BatchSize:     aws.Int32(tc.expectedBatchSize),
				MaximumBatchingWindowInSeconds: aws.Int32(tc.expectedBatchingWindow),
			}

			result := ConfigureEventSourceMappingForLatency(ctx, "latency-function", tc.eventSourceArn)

			// Assert latency optimization worked
			require.NotNil(t, result)
			assert.Equal(t, "latency-uuid-123", result.UUID)
			assert.Equal(t, tc.expectedBatchSize, result.BatchSize)
			assert.Equal(t, tc.expectedBatchingWindow, result.MaximumBatchingWindowInSeconds)

			// Verify correct batch configuration was sent
			require.Len(t, mockClient.CreateEventSourceMappingCalls, 1)
			call := mockClient.CreateEventSourceMappingCalls[0]
			assert.Equal(t, "latency-function", call.FunctionName)
			assert.Equal(t, tc.eventSourceArn, call.EventSourceArn)
			assert.Equal(t, tc.expectedBatchSize, *call.BatchSize)
		})
	}
}

// TestUpdateEventSourceMappingBatchConfiguration tests updating batch configuration
func TestUpdateEventSourceMappingBatchConfiguration(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}

	uuid := "update-test-uuid"
	
	// Setup current mapping response (for the GET call)
	mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
		UUID:          aws.String(uuid),
		EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:update-queue"),
		FunctionArn:   aws.String("arn:aws:lambda:us-east-1:123456789012:function:update-function"),
		State:         aws.String("Enabled"),
		BatchSize:     aws.Int32(10),
		MaximumBatchingWindowInSeconds: aws.Int32(1),
	}

	// Setup update response
	newBatchSize := int32(100)
	newBatchingWindow := int32(5)
	mockClient.UpdateEventSourceMappingResponse = &lambda.UpdateEventSourceMappingOutput{
		UUID:          aws.String(uuid),
		EventSourceArn: aws.String("arn:aws:sqs:us-east-1:123456789012:update-queue"),
		BatchSize:     aws.Int32(newBatchSize),
		MaximumBatchingWindowInSeconds: aws.Int32(newBatchingWindow),
		State:         aws.String("Enabled"),
	}

	// New batch configuration
	batchConfig := BatchConfigurationOptions{
		BatchSize:                        newBatchSize,
		MaximumBatchingWindowInSeconds:   newBatchingWindow,
	}

	result := UpdateEventSourceMappingBatchConfiguration(ctx, uuid, batchConfig)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, uuid, result.UUID)
	assert.Equal(t, newBatchSize, result.BatchSize)
	assert.Equal(t, newBatchingWindow, result.MaximumBatchingWindowInSeconds)

	// Verify both GET and UPDATE calls were made
	require.Len(t, mockClient.GetEventSourceMappingCalls, 1)
	assert.Equal(t, uuid, mockClient.GetEventSourceMappingCalls[0])
	
	require.Len(t, mockClient.UpdateEventSourceMappingCalls, 1)
	updateCall := mockClient.UpdateEventSourceMappingCalls[0]
	assert.Equal(t, uuid, updateCall.UUID)
	assert.Equal(t, newBatchSize, *updateCall.BatchSize)
}

// TestBatchConfigurationValidation tests validation for different event source types
func TestBatchConfigurationValidation(t *testing.T) {
	testCases := []struct {
		name                string
		eventSourceArn      string
		batchConfig         BatchConfigurationOptions
		shouldError         bool
		expectedErrorMsg    string
	}{
		{
			name:           "Valid SQS configuration",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:valid-queue",
			batchConfig: BatchConfigurationOptions{
				BatchSize:                        10,
				MaximumBatchingWindowInSeconds:   5,
			},
			shouldError: false,
		},
		{
			name:           "Invalid SQS batch size too large",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:invalid-queue",
			batchConfig: BatchConfigurationOptions{
				BatchSize: 2000, // Too large for SQS
			},
			shouldError:      true,
			expectedErrorMsg: "SQS batch size must be between 1 and 1000",
		},
		{
			name:           "Invalid SQS batching window too large",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:invalid-queue",
			batchConfig: BatchConfigurationOptions{
				BatchSize:                        10,
				MaximumBatchingWindowInSeconds:   500, // Too large for SQS
			},
			shouldError:      true,
			expectedErrorMsg: "SQS maximum batching window must be between 0 and 300",
		},
		{
			name:           "Valid Kinesis configuration",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/valid-stream",
			batchConfig: BatchConfigurationOptions{
				BatchSize:                        100,
				ParallelizationFactor:           10,
				MaximumBatchingWindowInSeconds:   5,
			},
			shouldError: false,
		},
		{
			name:           "Invalid Kinesis batch size too large",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/invalid-stream",
			batchConfig: BatchConfigurationOptions{
				BatchSize: 20000, // Too large for Kinesis
			},
			shouldError:      true,
			expectedErrorMsg: "Kinesis batch size must be between 1 and 10000",
		},
		{
			name:           "Invalid Kinesis parallelization factor",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/invalid-stream",
			batchConfig: BatchConfigurationOptions{
				BatchSize:                        100,
				ParallelizationFactor:           200, // Too large
			},
			shouldError:      true,
			expectedErrorMsg: "Kinesis parallelization factor must be between 1 and 100",
		},
		{
			name:           "Valid DynamoDB configuration",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/TestTable/stream/2023-01-01T00:00:00.000",
			batchConfig: BatchConfigurationOptions{
				BatchSize:                        50,
				ParallelizationFactor:           5,
			},
			shouldError: false,
		},
		{
			name:           "Invalid DynamoDB parallelization factor",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/TestTable/stream/2023-01-01T00:00:00.000",
			batchConfig: BatchConfigurationOptions{
				BatchSize:                        50,
				ParallelizationFactor:           150, // Too large
			},
			shouldError:      true,
			expectedErrorMsg: "DynamoDB parallelization factor must be between 1 and 100",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &TestContext{T: t}
			
			result, err := CreateEventSourceMappingWithBatchConfigE(ctx, "test-function", 
				tc.eventSourceArn, tc.batchConfig)

			if tc.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.expectedErrorMsg)
			} else {
				// For valid configurations, we'll get an error from the mock not being set up
				// but the validation should pass, so the error should be about the actual API call
				if err != nil {
					assert.NotContains(t, err.Error(), "batch size")
					assert.NotContains(t, err.Error(), "parallelization factor")
					assert.NotContains(t, err.Error(), "batching window")
				}
			}
		})
	}
}

// TestEventSourceTypeExtraction tests the getEventSourceType function
func TestEventSourceTypeExtraction(t *testing.T) {
	testCases := []struct {
		name           string
		eventSourceArn string
		expectedType   string
	}{
		{
			name:           "SQS queue",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:test-queue",
			expectedType:   "sqs",
		},
		{
			name:           "Kinesis stream",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/test-stream",
			expectedType:   "kinesis",
		},
		{
			name:           "DynamoDB stream",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/TestTable/stream/2023-01-01T00:00:00.000",
			expectedType:   "dynamodb",
		},
		{
			name:           "Unknown event source",
			eventSourceArn: "arn:aws:s3:::test-bucket",
			expectedType:   "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getEventSourceType(tc.eventSourceArn)
			assert.Equal(t, tc.expectedType, result)
		})
	}
}

// TestBatchConfigurationOptimization tests the optimization logic
func TestBatchConfigurationOptimization(t *testing.T) {
	testCases := []struct {
		name           string
		eventSourceArn string
		inputConfig    BatchConfigurationOptions
		expectedConfig BatchConfigurationOptions
	}{
		{
			name:           "SQS optimization with zero values",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:optimize-queue",
			inputConfig:    BatchConfigurationOptions{}, // All zeros
			expectedConfig: BatchConfigurationOptions{
				BatchSize:                        10, // SQS optimal default
				MaximumBatchingWindowInSeconds:   1,  // Low latency default
			},
		},
		{
			name:           "Kinesis optimization with zero values",
			eventSourceArn: "arn:aws:kinesis:us-east-1:123456789012:stream/optimize-stream",
			inputConfig:    BatchConfigurationOptions{}, // All zeros
			expectedConfig: BatchConfigurationOptions{
				BatchSize:                        100, // Kinesis optimal default
				ParallelizationFactor:           10,  // Reasonable parallelization
				MaximumBatchingWindowInSeconds:   5,   // Buffer for efficiency
			},
		},
		{
			name:           "DynamoDB optimization with zero values",
			eventSourceArn: "arn:aws:dynamodb:us-east-1:123456789012:table/OptimizeTable/stream/2023-01-01T00:00:00.000",
			inputConfig:    BatchConfigurationOptions{}, // All zeros
			expectedConfig: BatchConfigurationOptions{
				BatchSize:                        50, // DynamoDB optimal default
				ParallelizationFactor:           1,  // Conservative for consistency
			},
		},
		{
			name:           "Preserve non-zero values",
			eventSourceArn: "arn:aws:sqs:us-east-1:123456789012:preserve-queue",
			inputConfig: BatchConfigurationOptions{
				BatchSize:                        25, // User specified
				MaximumBatchingWindowInSeconds:   3,  // User specified
			},
			expectedConfig: BatchConfigurationOptions{
				BatchSize:                        25, // Preserved
				MaximumBatchingWindowInSeconds:   3,  // Preserved
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := optimizeBatchConfigurationForEventSource(tc.eventSourceArn, tc.inputConfig)
			
			assert.Equal(t, tc.expectedConfig.BatchSize, result.BatchSize)
			assert.Equal(t, tc.expectedConfig.ParallelizationFactor, result.ParallelizationFactor)
			assert.Equal(t, tc.expectedConfig.MaximumBatchingWindowInSeconds, result.MaximumBatchingWindowInSeconds)
		})
	}
}

// TestEventSourceMappingStateTransitions tests state transition handling
func TestEventSourceMappingStateTransitions(t *testing.T) {
	mockClient := &MockLambdaClient{}
	factory := &MockClientFactory{LambdaClient: mockClient}
	originalFactory := globalClientFactory
	SetClientFactory(factory)
	defer SetClientFactory(originalFactory)

	ctx := &TestContext{T: t}
	
	uuid := "state-transition-uuid"
	
	testCases := []struct {
		name             string
		initialState     string
		targetState      string
		transitionStates []string
		shouldTimeout    bool
	}{
		{
			name:             "Creating to Enabled transition",
			initialState:     "Creating",
			targetState:      "Enabled",
			transitionStates: []string{"Creating", "Creating", "Enabled"},
			shouldTimeout:    false,
		},
		{
			name:             "Updating to Disabled transition",
			initialState:     "Updating",
			targetState:      "Disabled",
			transitionStates: []string{"Updating", "Updating", "Disabled"},
			shouldTimeout:    false,
		},
		{
			name:             "Timeout scenario",
			initialState:     "Creating",
			targetState:      "Enabled",
			transitionStates: []string{"Creating"}, // Never transitions
			shouldTimeout:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock client
			mockClient.Reset()
			
			// Set up sequential state transitions using the new mock capability
			if !tc.shouldTimeout {
				// Create sequential responses for state transitions
				var responses []*lambda.GetEventSourceMappingOutput
				for _, state := range tc.transitionStates {
					responses = append(responses, &lambda.GetEventSourceMappingOutput{
						UUID:  aws.String(uuid),
						State: aws.String(state),
					})
				}
				mockClient.GetEventSourceMappingResponses = responses
			} else {
				// For timeout tests, always return the same state
				mockClient.GetEventSourceMappingResponse = &lambda.GetEventSourceMappingOutput{
					UUID:  aws.String(uuid),
					State: aws.String(tc.initialState),
				}
			}

			// Test the wait function
			timeout := 200 * time.Millisecond
			if tc.shouldTimeout {
				timeout = 100 * time.Millisecond // Short timeout for timeout test
			}
			
			err := WaitForEventSourceMappingStateE(ctx, uuid, tc.targetState, timeout)

			if tc.shouldTimeout {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout waiting for event source mapping")
			} else {
				assert.NoError(t, err)
			}

			// Verify GetEventSourceMapping was called
			assert.GreaterOrEqual(t, len(mockClient.GetEventSourceMappingCalls), 1)
		})
	}
}

// TestAdvancedFilteringAndParallelization tests advanced filtering and parallelization features
func TestAdvancedFilteringAndParallelization(t *testing.T) {
	t.Run("ParallelizationFactorValidation", func(t *testing.T) {
		// Test that parallelization factor is properly validated
		testCases := []struct {
			parallelizationFactor int32
			shouldPass            bool
		}{
			{1, true},   // Min parallelization
			{10, true},  // Normal parallelization
			{100, true}, // Max parallelization
			{101, false}, // Over max
			{0, false},   // Under min
			{-1, false},  // Negative
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("parallelization_%d", tc.parallelizationFactor), func(t *testing.T) {
				config := BatchConfigurationOptions{
					BatchSize:                        10,
					ParallelizationFactor:           tc.parallelizationFactor,
				}
				
				err := validateKinesisBatchConfiguration(config)
				if tc.shouldPass {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "parallelization factor")
				}
			})
		}
	})

	t.Run("FilterCriteriaValidation", func(t *testing.T) {
		// Test that filter criteria are properly validated
		validFilter := FilterCriteria{
			Filters: []EventFilter{
				{
					Pattern: map[string]interface{}{
						"eventType": []string{"OrderPlaced", "OrderCancelled"},
					},
				},
			},
		}
		
		invalidFilter := FilterCriteria{
			Filters: []EventFilter{
				{
					Pattern: map[string]interface{}{
						"invalidKey": "invalidValue",
					},
				},
			},
		}

		err := validateFilterCriteria(validFilter.Filters[0].Pattern)
		assert.NoError(t, err)

		err = validateFilterCriteria(invalidFilter.Filters[0].Pattern) 
		assert.Error(t, err)
	})
}


// validateFilterCriteria validates event filtering criteria  
func validateFilterCriteria(criteria map[string]interface{}) error {
	// Simple validation - in reality this would be more complex
	allowedKeys := map[string]bool{
		"eventType":   true,
		"source":      true,
		"detail-type": true,
		"detail":      true,
		"account":     true,
		"region":      true,
		"time":        true,
		"id":          true,
		"version":     true,
	}
	
	for key := range criteria {
		if !allowedKeys[key] {
			return fmt.Errorf("invalid filter key: %s", key)
		}
	}
	return nil
}