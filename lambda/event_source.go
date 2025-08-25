package lambda

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// CreateEventSourceMapping creates a new event source mapping for a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func CreateEventSourceMapping(ctx *TestContext, config EventSourceMappingConfig) *EventSourceMappingResult {
	result, err := CreateEventSourceMappingE(ctx, config)
	if err != nil {
		ctx.T.Errorf("Failed to create event source mapping: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// CreateEventSourceMappingE creates a new event source mapping for a Lambda function.
// This is the error returning version that follows Terratest patterns.
func CreateEventSourceMappingE(ctx *TestContext, config EventSourceMappingConfig) (*EventSourceMappingResult, error) {
	startTime := time.Now()
	
	// Validate configuration
	if err := validateEventSourceMappingConfig(&config); err != nil {
		return nil, err
	}
	
	logOperation("create_event_source_mapping", config.FunctionName, map[string]interface{}{
		"event_source_arn": config.EventSourceArn,
		"batch_size":       config.BatchSize,
		"enabled":          config.Enabled,
	})
	
	client := createLambdaClient(ctx)
	
	input := &lambda.CreateEventSourceMappingInput{
		EventSourceArn:   aws.String(config.EventSourceArn),
		FunctionName:     aws.String(config.FunctionName),
		StartingPosition: config.StartingPosition,
	}
	
	// Set optional fields
	if config.BatchSize > 0 {
		input.BatchSize = aws.Int32(config.BatchSize)
	}
	
	if config.MaximumBatchingWindowInSeconds > 0 {
		input.MaximumBatchingWindowInSeconds = aws.Int32(config.MaximumBatchingWindowInSeconds)
	}
	
	// Enabled defaults to true if not specified
	input.Enabled = aws.Bool(config.Enabled)
	
	output, err := client.CreateEventSourceMapping(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("create_event_source_mapping", config.FunctionName, false, duration, err)
		return nil, fmt.Errorf("failed to create event source mapping: %w", err)
	}
	
	result := convertToEventSourceMappingResult(output)
	
	duration := time.Since(startTime)
	logResult("create_event_source_mapping", config.FunctionName, true, duration, nil)
	
	return result, nil
}

// GetEventSourceMapping retrieves an event source mapping by UUID.
// This is the non-error returning version that follows Terratest patterns.
func GetEventSourceMapping(ctx *TestContext, uuid string) *EventSourceMappingResult {
	result, err := GetEventSourceMappingE(ctx, uuid)
	if err != nil {
		ctx.T.Errorf("Failed to get event source mapping: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// GetEventSourceMappingE retrieves an event source mapping by UUID.
// This is the error returning version that follows Terratest patterns.
func GetEventSourceMappingE(ctx *TestContext, uuid string) (*EventSourceMappingResult, error) {
	startTime := time.Now()
	
	if uuid == "" {
		return nil, fmt.Errorf("event source mapping UUID cannot be empty")
	}
	
	logOperation("get_event_source_mapping", uuid, map[string]interface{}{
		"uuid": uuid,
	})
	
	client := createLambdaClient(ctx)
	
	input := &lambda.GetEventSourceMappingInput{
		UUID: aws.String(uuid),
	}
	
	output, err := client.GetEventSourceMapping(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("get_event_source_mapping", uuid, false, duration, err)
		return nil, fmt.Errorf("failed to get event source mapping: %w", err)
	}
	
	result := convertToEventSourceMappingResult(output)
	
	duration := time.Since(startTime)
	logResult("get_event_source_mapping", uuid, true, duration, nil)
	
	return result, nil
}

// ListEventSourceMappings lists all event source mappings for a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func ListEventSourceMappings(ctx *TestContext, functionName string) []EventSourceMappingResult {
	results, err := ListEventSourceMappingsE(ctx, functionName)
	if err != nil {
		ctx.T.Errorf("Failed to list event source mappings: %v", err)
		ctx.T.FailNow()
	}
	return results
}

// ListEventSourceMappingsE lists all event source mappings for a Lambda function.
// This is the error returning version that follows Terratest patterns.
func ListEventSourceMappingsE(ctx *TestContext, functionName string) ([]EventSourceMappingResult, error) {
	startTime := time.Now()
	
	// Validate function name
	if err := validateFunctionName(functionName); err != nil {
		return nil, err
	}
	
	logOperation("list_event_source_mappings", functionName, map[string]interface{}{})
	
	client := createLambdaClient(ctx)
	
	input := &lambda.ListEventSourceMappingsInput{
		FunctionName: aws.String(functionName),
		MaxItems:     aws.Int32(100), // AWS maximum
	}
	
	var allMappings []EventSourceMappingResult
	var nextMarker *string
	
	for {
		if nextMarker != nil {
			input.Marker = nextMarker
		}
		
		output, err := client.ListEventSourceMappings(context.Background(), input)
		if err != nil {
			duration := time.Since(startTime)
			logResult("list_event_source_mappings", functionName, false, duration, err)
			return nil, fmt.Errorf("failed to list event source mappings: %w", err)
		}
		
		// Convert AWS types to our result types
		for _, awsMapping := range output.EventSourceMappings {
			result := convertToEventSourceMappingResult(&awsMapping)
			allMappings = append(allMappings, *result)
		}
		
		// Check if there are more mappings to retrieve
		if output.NextMarker == nil {
			break
		}
		nextMarker = output.NextMarker
	}
	
	duration := time.Since(startTime)
	logResult("list_event_source_mappings", functionName, true, duration, nil)
	
	return allMappings, nil
}

// UpdateEventSourceMapping updates an existing event source mapping.
// This is the non-error returning version that follows Terratest patterns.
func UpdateEventSourceMapping(ctx *TestContext, uuid string, config UpdateEventSourceMappingConfig) *EventSourceMappingResult {
	result, err := UpdateEventSourceMappingE(ctx, uuid, config)
	if err != nil {
		ctx.T.Errorf("Failed to update event source mapping: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// UpdateEventSourceMappingE updates an existing event source mapping.
// This is the error returning version that follows Terratest patterns.
func UpdateEventSourceMappingE(ctx *TestContext, uuid string, config UpdateEventSourceMappingConfig) (*EventSourceMappingResult, error) {
	startTime := time.Now()
	
	if uuid == "" {
		return nil, fmt.Errorf("event source mapping UUID cannot be empty")
	}
	
	logOperation("update_event_source_mapping", uuid, map[string]interface{}{
		"uuid":       uuid,
		"batch_size": config.BatchSize,
		"enabled":    config.Enabled,
	})
	
	client := createLambdaClient(ctx)
	
	input := &lambda.UpdateEventSourceMappingInput{
		UUID: aws.String(uuid),
	}
	
	// Set optional fields that should be updated
	if config.BatchSize != nil {
		input.BatchSize = config.BatchSize
	}
	
	if config.MaximumBatchingWindowInSeconds != nil {
		input.MaximumBatchingWindowInSeconds = config.MaximumBatchingWindowInSeconds
	}
	
	if config.Enabled != nil {
		input.Enabled = config.Enabled
	}
	
	if config.FunctionName != nil {
		input.FunctionName = config.FunctionName
	}
	
	output, err := client.UpdateEventSourceMapping(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("update_event_source_mapping", uuid, false, duration, err)
		return nil, fmt.Errorf("failed to update event source mapping: %w", err)
	}
	
	result := convertToEventSourceMappingResult(output)
	
	duration := time.Since(startTime)
	logResult("update_event_source_mapping", uuid, true, duration, nil)
	
	return result, nil
}

// DeleteEventSourceMapping deletes an event source mapping.
// This is the non-error returning version that follows Terratest patterns.
func DeleteEventSourceMapping(ctx *TestContext, uuid string) *EventSourceMappingResult {
	result, err := DeleteEventSourceMappingE(ctx, uuid)
	if err != nil {
		ctx.T.Errorf("Failed to delete event source mapping: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// DeleteEventSourceMappingE deletes an event source mapping.
// This is the error returning version that follows Terratest patterns.
func DeleteEventSourceMappingE(ctx *TestContext, uuid string) (*EventSourceMappingResult, error) {
	startTime := time.Now()
	
	if uuid == "" {
		return nil, fmt.Errorf("event source mapping UUID cannot be empty")
	}
	
	logOperation("delete_event_source_mapping", uuid, map[string]interface{}{
		"uuid": uuid,
	})
	
	client := createLambdaClient(ctx)
	
	input := &lambda.DeleteEventSourceMappingInput{
		UUID: aws.String(uuid),
	}
	
	output, err := client.DeleteEventSourceMapping(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("delete_event_source_mapping", uuid, false, duration, err)
		return nil, fmt.Errorf("failed to delete event source mapping: %w", err)
	}
	
	result := convertToEventSourceMappingResult(output)
	
	duration := time.Since(startTime)
	logResult("delete_event_source_mapping", uuid, true, duration, nil)
	
	return result, nil
}

// WaitForEventSourceMappingState waits for an event source mapping to reach the desired state.
// This is the non-error returning version that follows Terratest patterns.
func WaitForEventSourceMappingState(ctx *TestContext, uuid string, expectedState string, timeout time.Duration) {
	err := WaitForEventSourceMappingStateE(ctx, uuid, expectedState, timeout)
	if err != nil {
		ctx.T.Errorf("Event source mapping did not reach expected state within timeout: %v", err)
		ctx.T.FailNow()
	}
}

// WaitForEventSourceMappingStateE waits for an event source mapping to reach the desired state.
// This is the error returning version that follows Terratest patterns.
func WaitForEventSourceMappingStateE(ctx *TestContext, uuid string, expectedState string, timeout time.Duration) error {
	startTime := time.Now()
	
	logOperation("wait_for_event_source_mapping_state", uuid, map[string]interface{}{
		"expected_state":  expectedState,
		"timeout_seconds": timeout.Seconds(),
	})
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	timeoutChan := time.After(timeout)
	
	for {
		select {
		case <-timeoutChan:
			duration := time.Since(startTime)
			logResult("wait_for_event_source_mapping_state", uuid, false, duration, 
				fmt.Errorf("timeout waiting for state %s", expectedState))
			return fmt.Errorf("timeout waiting for event source mapping %s to reach state %s", uuid, expectedState)
			
		case <-ticker.C:
			mapping, err := GetEventSourceMappingE(ctx, uuid)
			if err != nil {
				duration := time.Since(startTime)
				logResult("wait_for_event_source_mapping_state", uuid, false, duration, err)
				return err
			}
			
			if mapping.State == expectedState {
				duration := time.Since(startTime)
				logResult("wait_for_event_source_mapping_state", uuid, true, duration, nil)
				return nil
			}
		}
	}
}

// EventSourceMappingResult represents the result of an event source mapping operation
type EventSourceMappingResult struct {
	UUID                             string
	EventSourceArn                   string
	FunctionArn                      string
	FunctionName                     string
	LastModified                     time.Time
	LastProcessingResult             string
	State                            string
	StateTransitionReason            string
	BatchSize                        int32
	MaximumBatchingWindowInSeconds   int32
	StartingPosition                 types.EventSourcePosition
	StartingPositionTimestamp        *time.Time
}

// UpdateEventSourceMappingConfig represents the configuration for updating an event source mapping
type UpdateEventSourceMappingConfig struct {
	BatchSize                        *int32
	MaximumBatchingWindowInSeconds   *int32
	Enabled                          *bool
	FunctionName                     *string
}

// validateEventSourceMappingConfig validates the event source mapping configuration
func validateEventSourceMappingConfig(config *EventSourceMappingConfig) error {
	if config == nil {
		return fmt.Errorf("event source mapping configuration cannot be nil")
	}
	
	if config.EventSourceArn == "" {
		return fmt.Errorf("event source ARN cannot be empty")
	}
	
	if config.FunctionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	
	if err := validateFunctionName(config.FunctionName); err != nil {
		return err
	}
	
	// Validate batch size based on event source type
	if config.BatchSize > 0 {
		eventSourceType := getEventSourceType(config.EventSourceArn)
		switch eventSourceType {
		case "sqs":
			if config.BatchSize > 1000 {
				return fmt.Errorf("SQS batch size %d exceeds maximum allowed (1000)", config.BatchSize)
			}
		case "kinesis":
			if config.BatchSize > 10000 {
				return fmt.Errorf("Kinesis batch size %d exceeds maximum allowed (10000)", config.BatchSize)
			}
		case "dynamodb":
			if config.BatchSize > 1000 {
				return fmt.Errorf("DynamoDB batch size %d exceeds maximum allowed (1000)", config.BatchSize)
			}
		default:
			// For unknown event sources, use conservative limit
			if config.BatchSize > 1000 {
				return fmt.Errorf("batch size %d exceeds maximum allowed (1000)", config.BatchSize)
			}
		}
	}
	
	return nil
}

// BatchConfigurationOptions represents advanced batch configuration settings
type BatchConfigurationOptions struct {
	BatchSize                        int32
	MaximumBatchingWindowInSeconds   int32
	ParallelizationFactor           int32
	MaximumRecordAgeInSeconds       int32
	BisectBatchOnFunctionError      bool
	MaximumRetryAttempts            int32
	TumblingWindowInSeconds         int32
}

// FilterCriteria represents event filtering configuration
type FilterCriteria struct {
	Filters []EventFilter `json:"filters"`
}

// EventFilter represents an individual filter rule
type EventFilter struct {
	Pattern map[string]interface{} `json:"pattern"`
}

// CreateEventSourceMappingWithBatchConfig creates an event source mapping with optimized batch configuration
func CreateEventSourceMappingWithBatchConfig(ctx *TestContext, functionName, eventSourceArn string, batchConfig BatchConfigurationOptions) *EventSourceMappingResult {
	result, err := CreateEventSourceMappingWithBatchConfigE(ctx, functionName, eventSourceArn, batchConfig)
	if err != nil {
		ctx.T.Errorf("Failed to create event source mapping with batch config: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// CreateEventSourceMappingWithBatchConfigE creates an event source mapping with optimized batch configuration (error-returning version)
func CreateEventSourceMappingWithBatchConfigE(ctx *TestContext, functionName, eventSourceArn string, batchConfig BatchConfigurationOptions) (*EventSourceMappingResult, error) {
	// Validate batch configuration for the specific event source type
	if err := validateBatchConfigurationForEventSource(eventSourceArn, batchConfig); err != nil {
		return nil, err
	}

	// Optimize batch configuration based on event source type
	optimizedConfig := optimizeBatchConfigurationForEventSource(eventSourceArn, batchConfig)

	config := EventSourceMappingConfig{
		EventSourceArn:                 eventSourceArn,
		FunctionName:                   functionName,
		BatchSize:                      optimizedConfig.BatchSize,
		MaximumBatchingWindowInSeconds: optimizedConfig.MaximumBatchingWindowInSeconds,
		StartingPosition:               types.EventSourcePositionLatest,
		Enabled:                        true,
	}

	return CreateEventSourceMappingE(ctx, config)
}

// UpdateEventSourceMappingBatchConfiguration updates batch configuration for an existing mapping
func UpdateEventSourceMappingBatchConfiguration(ctx *TestContext, uuid string, batchConfig BatchConfigurationOptions) *EventSourceMappingResult {
	result, err := UpdateEventSourceMappingBatchConfigurationE(ctx, uuid, batchConfig)
	if err != nil {
		ctx.T.Errorf("Failed to update event source mapping batch configuration: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// UpdateEventSourceMappingBatchConfigurationE updates batch configuration for an existing mapping (error-returning version)
func UpdateEventSourceMappingBatchConfigurationE(ctx *TestContext, uuid string, batchConfig BatchConfigurationOptions) (*EventSourceMappingResult, error) {
	// Get current mapping to validate event source type
	currentMapping, err := GetEventSourceMappingE(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get current mapping: %w", err)
	}

	// Validate and optimize the new batch configuration
	if err := validateBatchConfigurationForEventSource(currentMapping.EventSourceArn, batchConfig); err != nil {
		return nil, err
	}

	optimizedConfig := optimizeBatchConfigurationForEventSource(currentMapping.EventSourceArn, batchConfig)

	updateConfig := UpdateEventSourceMappingConfig{
		BatchSize:                        &optimizedConfig.BatchSize,
		MaximumBatchingWindowInSeconds:   &optimizedConfig.MaximumBatchingWindowInSeconds,
	}

	return UpdateEventSourceMappingE(ctx, uuid, updateConfig)
}

// ConfigureEventSourceMappingForThroughput configures an event source mapping for optimal throughput
func ConfigureEventSourceMappingForThroughput(ctx *TestContext, functionName, eventSourceArn string) *EventSourceMappingResult {
	result, err := ConfigureEventSourceMappingForThroughputE(ctx, functionName, eventSourceArn)
	if err != nil {
		ctx.T.Errorf("Failed to configure event source mapping for throughput: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// ConfigureEventSourceMappingForThroughputE configures an event source mapping for optimal throughput (error-returning version)
func ConfigureEventSourceMappingForThroughputE(ctx *TestContext, functionName, eventSourceArn string) (*EventSourceMappingResult, error) {
	// Get optimal configuration for throughput based on event source type
	optimalConfig := getOptimalThroughputConfiguration(eventSourceArn)

	return CreateEventSourceMappingWithBatchConfigE(ctx, functionName, eventSourceArn, optimalConfig)
}

// ConfigureEventSourceMappingForLatency configures an event source mapping for optimal latency
func ConfigureEventSourceMappingForLatency(ctx *TestContext, functionName, eventSourceArn string) *EventSourceMappingResult {
	result, err := ConfigureEventSourceMappingForLatencyE(ctx, functionName, eventSourceArn)
	if err != nil {
		ctx.T.Errorf("Failed to configure event source mapping for latency: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// ConfigureEventSourceMappingForLatencyE configures an event source mapping for optimal latency (error-returning version)
func ConfigureEventSourceMappingForLatencyE(ctx *TestContext, functionName, eventSourceArn string) (*EventSourceMappingResult, error) {
	// Get optimal configuration for latency based on event source type
	optimalConfig := getOptimalLatencyConfiguration(eventSourceArn)

	return CreateEventSourceMappingWithBatchConfigE(ctx, functionName, eventSourceArn, optimalConfig)
}

// validateBatchConfigurationForEventSource validates batch configuration against event source limits
func validateBatchConfigurationForEventSource(eventSourceArn string, config BatchConfigurationOptions) error {
	eventSourceType := getEventSourceType(eventSourceArn)
	
	switch eventSourceType {
	case "sqs":
		return validateSQSBatchConfiguration(config)
	case "kinesis":
		return validateKinesisBatchConfiguration(config)
	case "dynamodb":
		return validateDynamoDBBatchConfiguration(config)
	default:
		return validateDefaultBatchConfiguration(config)
	}
}

// validateSQSBatchConfiguration validates SQS-specific batch configuration
func validateSQSBatchConfiguration(config BatchConfigurationOptions) error {
	if config.BatchSize < 1 || config.BatchSize > 1000 {
		return fmt.Errorf("SQS batch size must be between 1 and 1000, got %d", config.BatchSize)
	}
	
	if config.MaximumBatchingWindowInSeconds < 0 || config.MaximumBatchingWindowInSeconds > 300 {
		return fmt.Errorf("SQS maximum batching window must be between 0 and 300 seconds, got %d", config.MaximumBatchingWindowInSeconds)
	}
	
	return nil
}

// validateKinesisBatchConfiguration validates Kinesis-specific batch configuration
func validateKinesisBatchConfiguration(config BatchConfigurationOptions) error {
	if config.BatchSize < 1 || config.BatchSize > 10000 {
		return fmt.Errorf("Kinesis batch size must be between 1 and 10000, got %d", config.BatchSize)
	}
	
	if config.ParallelizationFactor < 1 || config.ParallelizationFactor > 100 {
		return fmt.Errorf("Kinesis parallelization factor must be between 1 and 100, got %d", config.ParallelizationFactor)
	}
	
	if config.MaximumBatchingWindowInSeconds < 0 || config.MaximumBatchingWindowInSeconds > 300 {
		return fmt.Errorf("Kinesis maximum batching window must be between 0 and 300 seconds, got %d", config.MaximumBatchingWindowInSeconds)
	}
	
	return nil
}

// validateDynamoDBBatchConfiguration validates DynamoDB-specific batch configuration
func validateDynamoDBBatchConfiguration(config BatchConfigurationOptions) error {
	if config.BatchSize < 1 || config.BatchSize > 1000 {
		return fmt.Errorf("DynamoDB batch size must be between 1 and 1000, got %d", config.BatchSize)
	}
	
	if config.ParallelizationFactor < 1 || config.ParallelizationFactor > 100 {
		return fmt.Errorf("DynamoDB parallelization factor must be between 1 and 100, got %d", config.ParallelizationFactor)
	}
	
	return nil
}

// validateDefaultBatchConfiguration validates default batch configuration
func validateDefaultBatchConfiguration(config BatchConfigurationOptions) error {
	if config.BatchSize < 1 || config.BatchSize > 1000 {
		return fmt.Errorf("batch size must be between 1 and 1000, got %d", config.BatchSize)
	}
	return nil
}

// optimizeBatchConfigurationForEventSource optimizes batch configuration based on event source
func optimizeBatchConfigurationForEventSource(eventSourceArn string, config BatchConfigurationOptions) BatchConfigurationOptions {
	eventSourceType := getEventSourceType(eventSourceArn)
	
	optimized := config
	
	switch eventSourceType {
	case "sqs":
		// SQS optimization: balance between throughput and cost
		if optimized.BatchSize == 0 {
			optimized.BatchSize = 10 // Optimal for most SQS use cases
		}
		if optimized.MaximumBatchingWindowInSeconds == 0 {
			optimized.MaximumBatchingWindowInSeconds = 1 // Low latency
		}
		
	case "kinesis":
		// Kinesis optimization: maximize parallelization for high throughput
		if optimized.BatchSize == 0 {
			optimized.BatchSize = 100 // Good balance for Kinesis
		}
		if optimized.ParallelizationFactor == 0 {
			optimized.ParallelizationFactor = 10 // Reasonable parallelization
		}
		if optimized.MaximumBatchingWindowInSeconds == 0 {
			optimized.MaximumBatchingWindowInSeconds = 5 // Buffer for efficiency
		}
		
	case "dynamodb":
		// DynamoDB streams optimization: balance between throughput and consistency
		if optimized.BatchSize == 0 {
			optimized.BatchSize = 50 // Good for DynamoDB streams
		}
		if optimized.ParallelizationFactor == 0 {
			optimized.ParallelizationFactor = 1 // Conservative for consistency
		}
	}
	
	return optimized
}

// getOptimalThroughputConfiguration returns optimal configuration for maximum throughput
func getOptimalThroughputConfiguration(eventSourceArn string) BatchConfigurationOptions {
	eventSourceType := getEventSourceType(eventSourceArn)
	
	switch eventSourceType {
	case "sqs":
		return BatchConfigurationOptions{
			BatchSize:                        1000, // Max batch size for SQS
			MaximumBatchingWindowInSeconds:   5,    // Buffer for full batches
		}
	case "kinesis":
		return BatchConfigurationOptions{
			BatchSize:                        10000, // Max batch size for Kinesis
			ParallelizationFactor:           100,   // Max parallelization
			MaximumBatchingWindowInSeconds:   0,    // No buffering for max throughput
		}
	case "dynamodb":
		return BatchConfigurationOptions{
			BatchSize:                        1000, // Max batch size for DynamoDB
			ParallelizationFactor:           100,   // Max parallelization
		}
	default:
		return BatchConfigurationOptions{
			BatchSize:                        100,
			MaximumBatchingWindowInSeconds:   1,
		}
	}
}

// getOptimalLatencyConfiguration returns optimal configuration for minimum latency
func getOptimalLatencyConfiguration(eventSourceArn string) BatchConfigurationOptions {
	eventSourceType := getEventSourceType(eventSourceArn)
	
	switch eventSourceType {
	case "sqs":
		return BatchConfigurationOptions{
			BatchSize:                        1, // Min batch size for lowest latency
			MaximumBatchingWindowInSeconds:   0, // No buffering
		}
	case "kinesis":
		return BatchConfigurationOptions{
			BatchSize:                        1,   // Min batch size
			ParallelizationFactor:           100, // Max parallelization for fast processing
			MaximumBatchingWindowInSeconds:   0,   // No buffering
		}
	case "dynamodb":
		return BatchConfigurationOptions{
			BatchSize:                        1,   // Min batch size
			ParallelizationFactor:           1,   // Sequential processing for consistency
		}
	default:
		return BatchConfigurationOptions{
			BatchSize:                        1,
			MaximumBatchingWindowInSeconds:   0,
		}
	}
}

// getEventSourceType extracts event source type from ARN
func getEventSourceType(eventSourceArn string) string {
	if strings.Contains(eventSourceArn, ":sqs:") {
		return "sqs"
	}
	if strings.Contains(eventSourceArn, ":kinesis:") {
		return "kinesis"
	}
	if strings.Contains(eventSourceArn, ":dynamodb:") {
		return "dynamodb"
	}
	return "unknown"
}

// convertToEventSourceMappingResult converts AWS event source mapping to our result type
func convertToEventSourceMappingResult(awsMapping interface{}) *EventSourceMappingResult {
	result := &EventSourceMappingResult{}
	
	// Handle different AWS SDK response types
	switch mapping := awsMapping.(type) {
	case *lambda.CreateEventSourceMappingOutput:
		if mapping.UUID != nil {
			result.UUID = *mapping.UUID
		}
		if mapping.EventSourceArn != nil {
			result.EventSourceArn = *mapping.EventSourceArn
		}
		if mapping.FunctionArn != nil {
			result.FunctionArn = *mapping.FunctionArn
		}
		if mapping.LastModified != nil {
			result.LastModified = *mapping.LastModified
		}
		if mapping.LastProcessingResult != nil {
			result.LastProcessingResult = *mapping.LastProcessingResult
		}
		if mapping.State != nil {
			result.State = *mapping.State
		}
		if mapping.StateTransitionReason != nil {
			result.StateTransitionReason = *mapping.StateTransitionReason
		}
		if mapping.BatchSize != nil {
			result.BatchSize = *mapping.BatchSize
		}
		if mapping.MaximumBatchingWindowInSeconds != nil {
			result.MaximumBatchingWindowInSeconds = *mapping.MaximumBatchingWindowInSeconds
		}
		result.StartingPosition = mapping.StartingPosition
		if mapping.StartingPositionTimestamp != nil {
			result.StartingPositionTimestamp = mapping.StartingPositionTimestamp
		}
		
	case *lambda.GetEventSourceMappingOutput:
		if mapping.UUID != nil {
			result.UUID = *mapping.UUID
		}
		if mapping.EventSourceArn != nil {
			result.EventSourceArn = *mapping.EventSourceArn
		}
		if mapping.FunctionArn != nil {
			result.FunctionArn = *mapping.FunctionArn
		}
		if mapping.LastModified != nil {
			result.LastModified = *mapping.LastModified
		}
		if mapping.LastProcessingResult != nil {
			result.LastProcessingResult = *mapping.LastProcessingResult
		}
		if mapping.State != nil {
			result.State = *mapping.State
		}
		if mapping.StateTransitionReason != nil {
			result.StateTransitionReason = *mapping.StateTransitionReason
		}
		if mapping.BatchSize != nil {
			result.BatchSize = *mapping.BatchSize
		}
		if mapping.MaximumBatchingWindowInSeconds != nil {
			result.MaximumBatchingWindowInSeconds = *mapping.MaximumBatchingWindowInSeconds
		}
		result.StartingPosition = mapping.StartingPosition
		if mapping.StartingPositionTimestamp != nil {
			result.StartingPositionTimestamp = mapping.StartingPositionTimestamp
		}
		
	case *lambda.UpdateEventSourceMappingOutput:
		if mapping.UUID != nil {
			result.UUID = *mapping.UUID
		}
		if mapping.EventSourceArn != nil {
			result.EventSourceArn = *mapping.EventSourceArn
		}
		if mapping.FunctionArn != nil {
			result.FunctionArn = *mapping.FunctionArn
		}
		if mapping.LastModified != nil {
			result.LastModified = *mapping.LastModified
		}
		if mapping.LastProcessingResult != nil {
			result.LastProcessingResult = *mapping.LastProcessingResult
		}
		if mapping.State != nil {
			result.State = *mapping.State
		}
		if mapping.StateTransitionReason != nil {
			result.StateTransitionReason = *mapping.StateTransitionReason
		}
		if mapping.BatchSize != nil {
			result.BatchSize = *mapping.BatchSize
		}
		if mapping.MaximumBatchingWindowInSeconds != nil {
			result.MaximumBatchingWindowInSeconds = *mapping.MaximumBatchingWindowInSeconds
		}
		result.StartingPosition = mapping.StartingPosition
		if mapping.StartingPositionTimestamp != nil {
			result.StartingPositionTimestamp = mapping.StartingPositionTimestamp
		}
		
	case *lambda.DeleteEventSourceMappingOutput:
		if mapping.UUID != nil {
			result.UUID = *mapping.UUID
		}
		if mapping.EventSourceArn != nil {
			result.EventSourceArn = *mapping.EventSourceArn
		}
		if mapping.FunctionArn != nil {
			result.FunctionArn = *mapping.FunctionArn
		}
		if mapping.LastModified != nil {
			result.LastModified = *mapping.LastModified
		}
		if mapping.LastProcessingResult != nil {
			result.LastProcessingResult = *mapping.LastProcessingResult
		}
		if mapping.State != nil {
			result.State = *mapping.State
		}
		if mapping.StateTransitionReason != nil {
			result.StateTransitionReason = *mapping.StateTransitionReason
		}
		if mapping.BatchSize != nil {
			result.BatchSize = *mapping.BatchSize
		}
		if mapping.MaximumBatchingWindowInSeconds != nil {
			result.MaximumBatchingWindowInSeconds = *mapping.MaximumBatchingWindowInSeconds
		}
		result.StartingPosition = mapping.StartingPosition
		if mapping.StartingPositionTimestamp != nil {
			result.StartingPositionTimestamp = mapping.StartingPositionTimestamp
		}
		
	case *types.EventSourceMappingConfiguration:
		if mapping.UUID != nil {
			result.UUID = *mapping.UUID
		}
		if mapping.EventSourceArn != nil {
			result.EventSourceArn = *mapping.EventSourceArn
		}
		if mapping.FunctionArn != nil {
			result.FunctionArn = *mapping.FunctionArn
		}
		if mapping.LastModified != nil {
			result.LastModified = *mapping.LastModified
		}
		if mapping.LastProcessingResult != nil {
			result.LastProcessingResult = *mapping.LastProcessingResult
		}
		if mapping.State != nil {
			result.State = *mapping.State
		}
		if mapping.StateTransitionReason != nil {
			result.StateTransitionReason = *mapping.StateTransitionReason
		}
		if mapping.BatchSize != nil {
			result.BatchSize = *mapping.BatchSize
		}
		if mapping.MaximumBatchingWindowInSeconds != nil {
			result.MaximumBatchingWindowInSeconds = *mapping.MaximumBatchingWindowInSeconds
		}
		result.StartingPosition = mapping.StartingPosition
		if mapping.StartingPositionTimestamp != nil {
			result.StartingPositionTimestamp = mapping.StartingPositionTimestamp
		}
	}
	
	return result
}

// EventSourceMappingExistsE checks if an event source mapping exists by UUID.
func EventSourceMappingExistsE(ctx *TestContext, uuid string) (bool, error) {
	_, err := GetEventSourceMappingE(ctx, uuid)
	if err != nil {
		// Check if the error is "not found"
		if strings.Contains(err.Error(), "ResourceNotFoundException") ||
		   strings.Contains(err.Error(), "does not exist") {
			return false, nil
		}
		// Other errors should be propagated
		return false, err
	}
	return true, nil
}

// LogGroupExistsE checks if a CloudWatch log group exists for the function.
func LogGroupExistsE(ctx *TestContext, functionName string) (bool, error) {
	logGroupName := fmt.Sprintf("/aws/lambda/%s", functionName)
	client := createCloudWatchLogsClient(ctx)
	
	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(logGroupName),
		Limit:              aws.Int32(1),
	}
	
	output, err := client.DescribeLogGroups(context.Background(), input)
	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			return false, nil
		}
		return false, err
	}
	
	// Check if we found the exact log group
	for _, logGroup := range output.LogGroups {
		if aws.ToString(logGroup.LogGroupName) == logGroupName {
			return true, nil
		}
	}
	
	return false, nil
}

// LogStreamExistsE checks if a CloudWatch log stream exists for the function.
func LogStreamExistsE(ctx *TestContext, functionName string, logStreamName string) (bool, error) {
	logGroupName := fmt.Sprintf("/aws/lambda/%s", functionName)
	client := createCloudWatchLogsClient(ctx)
	
	input := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroupName),
		LogStreamNamePrefix: aws.String(logStreamName),
		Limit:               aws.Int32(1),
	}
	
	output, err := client.DescribeLogStreams(context.Background(), input)
	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			return false, nil
		}
		return false, err
	}
	
	// Check if we found the exact log stream
	for _, logStream := range output.LogStreams {
		if aws.ToString(logStream.LogStreamName) == logStreamName {
			return true, nil
		}
	}
	
	return false, nil
}