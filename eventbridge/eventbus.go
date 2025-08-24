package eventbridge

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// CreateEventBus creates a custom EventBridge event bus and panics on error (Terratest pattern)
func CreateEventBus(ctx *TestContext, config EventBusConfig) EventBusResult {
	result, err := CreateEventBusE(ctx, config)
	if err != nil {
		ctx.T.Errorf("CreateEventBus failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// CreateEventBusE creates a custom EventBridge event bus and returns error
func CreateEventBusE(ctx *TestContext, config EventBusConfig) (EventBusResult, error) {
	startTime := time.Now()
	
	logOperation("create_event_bus", map[string]interface{}{
		"bus_name":         config.Name,
		"event_source":     config.EventSourceName,
	})
	
	if config.Name == "" {
		err := fmt.Errorf("event bus name cannot be empty")
		logResult("create_event_bus", false, time.Since(startTime), err)
		return EventBusResult{}, err
	}
	
	if config.Name == DefaultEventBusName {
		err := fmt.Errorf("cannot create default event bus")
		logResult("create_event_bus", false, time.Since(startTime), err)
		return EventBusResult{}, err
	}
	
	client := createEventBridgeClient(ctx)
	
	// Create the event bus
	input := &eventbridge.CreateEventBusInput{
		Name: &config.Name,
	}
	
	// Add event source name if specified (for partner event buses)
	if config.EventSourceName != "" {
		input.EventSourceName = &config.EventSourceName
	}
	
	// Add tags if specified
	if len(config.Tags) > 0 {
		tags := make([]types.Tag, 0, len(config.Tags))
		for key, value := range config.Tags {
			tags = append(tags, types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}
		input.Tags = tags
	}
	
	// Add dead letter config if specified
	if config.DeadLetterConfig != nil {
		input.DeadLetterConfig = config.DeadLetterConfig
	}
	
	// Execute with retry logic
	retryConfig := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < retryConfig.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, retryConfig)
			time.Sleep(delay)
		}
		
		response, err := client.CreateEventBus(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		result := EventBusResult{
			Name:            config.Name,
			Arn:             safeString(response.EventBusArn),
			EventSourceName: config.EventSourceName,
		}
		
		logResult("create_event_bus", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("create_event_bus", false, time.Since(startTime), lastErr)
	return EventBusResult{}, fmt.Errorf("failed to create event bus after %d attempts: %w", retryConfig.MaxAttempts, lastErr)
}

// DeleteEventBus deletes a custom EventBridge event bus and panics on error (Terratest pattern)
func DeleteEventBus(ctx *TestContext, eventBusName string) {
	err := DeleteEventBusE(ctx, eventBusName)
	if err != nil {
		ctx.T.Errorf("DeleteEventBus failed: %v", err)
		ctx.T.FailNow()
	}
}

// DeleteEventBusE deletes a custom EventBridge event bus and returns error
func DeleteEventBusE(ctx *TestContext, eventBusName string) error {
	startTime := time.Now()
	
	logOperation("delete_event_bus", map[string]interface{}{
		"bus_name": eventBusName,
	})
	
	if eventBusName == "" {
		err := fmt.Errorf("event bus name cannot be empty")
		logResult("delete_event_bus", false, time.Since(startTime), err)
		return err
	}
	
	if eventBusName == DefaultEventBusName {
		err := fmt.Errorf("cannot delete default event bus")
		logResult("delete_event_bus", false, time.Since(startTime), err)
		return err
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.DeleteEventBusInput{
			Name: &eventBusName,
		}
		
		_, err := client.DeleteEventBus(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		logResult("delete_event_bus", true, time.Since(startTime), nil)
		return nil
	}
	
	logResult("delete_event_bus", false, time.Since(startTime), lastErr)
	return fmt.Errorf("failed to delete event bus after %d attempts: %w", config.MaxAttempts, lastErr)
}

// DescribeEventBus describes an EventBridge event bus and panics on error (Terratest pattern)
func DescribeEventBus(ctx *TestContext, eventBusName string) EventBusResult {
	result, err := DescribeEventBusE(ctx, eventBusName)
	if err != nil {
		ctx.T.Errorf("DescribeEventBus failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// DescribeEventBusE describes an EventBridge event bus and returns error
func DescribeEventBusE(ctx *TestContext, eventBusName string) (EventBusResult, error) {
	startTime := time.Now()
	
	logOperation("describe_event_bus", map[string]interface{}{
		"bus_name": eventBusName,
	})
	
	// Set default event bus if not specified
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.DescribeEventBusInput{
			Name: &eventBusName,
		}
		
		response, err := client.DescribeEventBus(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		result := EventBusResult{
			Name:         safeString(response.Name),
			Arn:          safeString(response.Arn),
			CreationTime: response.CreationTime,
		}
		
		logResult("describe_event_bus", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("describe_event_bus", false, time.Since(startTime), lastErr)
	return EventBusResult{}, fmt.Errorf("failed to describe event bus after %d attempts: %w", config.MaxAttempts, lastErr)
}

// ListEventBuses lists EventBridge event buses and panics on error (Terratest pattern)
func ListEventBuses(ctx *TestContext) []EventBusResult {
	results, err := ListEventBusesE(ctx)
	if err != nil {
		ctx.T.Errorf("ListEventBuses failed: %v", err)
		ctx.T.FailNow()
	}
	return results
}

// ListEventBusesE lists EventBridge event buses and returns error
func ListEventBusesE(ctx *TestContext) ([]EventBusResult, error) {
	startTime := time.Now()
	
	logOperation("list_event_buses", map[string]interface{}{})
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.ListEventBusesInput{}
		
		response, err := client.ListEventBuses(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Convert results
		results := make([]EventBusResult, len(response.EventBuses))
		for i, bus := range response.EventBuses {
			results[i] = EventBusResult{
				Name:         safeString(bus.Name),
				Arn:          safeString(bus.Arn),
				CreationTime: bus.CreationTime,
			}
		}
		
		logResult("list_event_buses", true, time.Since(startTime), nil)
		return results, nil
	}
	
	logResult("list_event_buses", false, time.Since(startTime), lastErr)
	return nil, fmt.Errorf("failed to list event buses after %d attempts: %w", config.MaxAttempts, lastErr)
}

// PutPermission adds permission for cross-account access and panics on error (Terratest pattern)
func PutPermission(ctx *TestContext, eventBusName string, statementID string, principal string, action string) {
	err := PutPermissionE(ctx, eventBusName, statementID, principal, action)
	if err != nil {
		ctx.T.Errorf("PutPermission failed: %v", err)
		ctx.T.FailNow()
	}
}

// PutPermissionE adds permission for cross-account access and returns error
func PutPermissionE(ctx *TestContext, eventBusName string, statementID string, principal string, action string) error {
	startTime := time.Now()
	
	logOperation("put_permission", map[string]interface{}{
		"bus_name":     eventBusName,
		"statement_id": statementID,
		"principal":    principal,
		"action":       action,
	})
	
	if statementID == "" {
		err := fmt.Errorf("statement ID cannot be empty")
		logResult("put_permission", false, time.Since(startTime), err)
		return err
	}
	
	if principal == "" {
		err := fmt.Errorf("principal cannot be empty")
		logResult("put_permission", false, time.Since(startTime), err)
		return err
	}
	
	if action == "" {
		action = "events:PutEvents"
	}
	
	// Set default event bus if not specified
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.PutPermissionInput{
			StatementId:  &statementID,
			Principal:    &principal,
			Action:       &action,
			EventBusName: &eventBusName,
		}
		
		_, err := client.PutPermission(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		logResult("put_permission", true, time.Since(startTime), nil)
		return nil
	}
	
	logResult("put_permission", false, time.Since(startTime), lastErr)
	return fmt.Errorf("failed to put permission after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RemovePermission removes permission for cross-account access and panics on error (Terratest pattern)
func RemovePermission(ctx *TestContext, eventBusName string, statementID string) {
	err := RemovePermissionE(ctx, eventBusName, statementID)
	if err != nil {
		ctx.T.Errorf("RemovePermission failed: %v", err)
		ctx.T.FailNow()
	}
}

// RemovePermissionE removes permission for cross-account access and returns error
func RemovePermissionE(ctx *TestContext, eventBusName string, statementID string) error {
	startTime := time.Now()
	
	logOperation("remove_permission", map[string]interface{}{
		"bus_name":     eventBusName,
		"statement_id": statementID,
	})
	
	if statementID == "" {
		err := fmt.Errorf("statement ID cannot be empty")
		logResult("remove_permission", false, time.Since(startTime), err)
		return err
	}
	
	// Set default event bus if not specified
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.RemovePermissionInput{
			StatementId:  &statementID,
			EventBusName: &eventBusName,
		}
		
		_, err := client.RemovePermission(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		logResult("remove_permission", true, time.Since(startTime), nil)
		return nil
	}
	
	logResult("remove_permission", false, time.Since(startTime), lastErr)
	return fmt.Errorf("failed to remove permission after %d attempts: %w", config.MaxAttempts, lastErr)
}