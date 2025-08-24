package lambda

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	
	// Validate batch size is within AWS limits
	if config.BatchSize > 0 {
		// Different services have different limits, but 1000 is generally safe
		if config.BatchSize > 1000 {
			return fmt.Errorf("batch size %d exceeds maximum allowed (1000)", config.BatchSize)
		}
	}
	
	return nil
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