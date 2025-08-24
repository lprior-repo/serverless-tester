package eventbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// PutEvent publishes a single event to EventBridge and panics on error (Terratest pattern)
func PutEvent(ctx *TestContext, event CustomEvent) PutEventResult {
	result, err := PutEventE(ctx, event)
	if err != nil {
		ctx.T.Errorf("PutEvent failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// PutEventE publishes a single event to EventBridge and returns error
func PutEventE(ctx *TestContext, event CustomEvent) (PutEventResult, error) {
	startTime := time.Now()
	
	logOperation("put_event", map[string]interface{}{
		"source":        event.Source,
		"detail_type":   event.DetailType,
		"event_bus":     event.EventBusName,
	})
	
	// Validate event detail
	if err := validateEventDetail(event.Detail); err != nil {
		logResult("put_event", false, time.Since(startTime), err)
		return PutEventResult{}, err
	}
	
	// Set default event bus if not specified
	eventBusName := event.EventBusName
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	// Set default time if not specified
	eventTime := event.Time
	if eventTime == nil {
		now := time.Now()
		eventTime = &now
	}
	
	client := createEventBridgeClient(ctx)
	
	// Create the put events request
	input := &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Source:       &event.Source,
				DetailType:   &event.DetailType,
				Detail:       &event.Detail,
				EventBusName: &eventBusName,
				Time:         eventTime,
			},
		},
	}
	
	// Add resources if specified
	if len(event.Resources) > 0 {
		input.Entries[0].Resources = event.Resources
	}
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		response, err := client.PutEvents(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Check if the event failed
		if response.FailedEntryCount > 0 {
			if len(response.Entries) > 0 {
				entry := response.Entries[0]
				result := PutEventResult{
					Success:      false,
					ErrorCode:    safeString(entry.ErrorCode),
					ErrorMessage: safeString(entry.ErrorMessage),
				}
				
				lastErr = fmt.Errorf("event failed: %s - %s", result.ErrorCode, result.ErrorMessage)
				logResult("put_event", false, time.Since(startTime), lastErr)
				return result, lastErr
			}
		}
		
		// Success case
		if len(response.Entries) > 0 {
			entry := response.Entries[0]
			result := PutEventResult{
				EventID: safeString(entry.EventId),
				Success: true,
			}
			
			logResult("put_event", true, time.Since(startTime), nil)
			return result, nil
		}
	}
	
	logResult("put_event", false, time.Since(startTime), lastErr)
	return PutEventResult{}, fmt.Errorf("failed to put event after %d attempts: %w", config.MaxAttempts, lastErr)
}

// PutEvents publishes multiple events to EventBridge and panics on error (Terratest pattern)
func PutEvents(ctx *TestContext, events []CustomEvent) PutEventsResult {
	result, err := PutEventsE(ctx, events)
	if err != nil {
		ctx.T.Errorf("PutEvents failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// PutEventsE publishes multiple events to EventBridge and returns error
func PutEventsE(ctx *TestContext, events []CustomEvent) (PutEventsResult, error) {
	startTime := time.Now()
	
	logOperation("put_events", map[string]interface{}{
		"event_count": len(events),
	})
	
	if len(events) == 0 {
		return PutEventsResult{Entries: []PutEventResult{}}, nil
	}
	
	if len(events) > MaxEventBatchSize {
		err := fmt.Errorf("batch size exceeds maximum: %d events", MaxEventBatchSize)
		logResult("put_events", false, time.Since(startTime), err)
		return PutEventsResult{}, err
	}
	
	client := createEventBridgeClient(ctx)
	
	// Build entries
	entries := make([]types.PutEventsRequestEntry, 0, len(events))
	for _, event := range events {
		// Validate event detail
		if err := validateEventDetail(event.Detail); err != nil {
			logResult("put_events", false, time.Since(startTime), err)
			return PutEventsResult{}, err
		}
		
		// Set default event bus if not specified
		eventBusName := event.EventBusName
		if eventBusName == "" {
			eventBusName = DefaultEventBusName
		}
		
		// Set default time if not specified
		eventTime := event.Time
		if eventTime == nil {
			now := time.Now()
			eventTime = &now
		}
		
		entry := types.PutEventsRequestEntry{
			Source:       &event.Source,
			DetailType:   &event.DetailType,
			Detail:       &event.Detail,
			EventBusName: &eventBusName,
			Time:         eventTime,
		}
		
		// Add resources if specified
		if len(event.Resources) > 0 {
			entry.Resources = event.Resources
		}
		
		entries = append(entries, entry)
	}
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.PutEventsInput{
			Entries: entries,
		}
		
		response, err := client.PutEvents(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Process results
		results := make([]PutEventResult, len(response.Entries))
		for i, entry := range response.Entries {
			if entry.EventId != nil {
				results[i] = PutEventResult{
					EventID: *entry.EventId,
					Success: true,
				}
			} else {
				results[i] = PutEventResult{
					Success:      false,
					ErrorCode:    safeString(entry.ErrorCode),
					ErrorMessage: safeString(entry.ErrorMessage),
				}
			}
		}
		
		result := PutEventsResult{
			FailedEntryCount: response.FailedEntryCount,
			Entries:         results,
		}
		
		logResult("put_events", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("put_events", false, time.Since(startTime), lastErr)
	return PutEventsResult{}, fmt.Errorf("failed to put events after %d attempts: %w", config.MaxAttempts, lastErr)
}

// BuildCustomEvent creates a custom EventBridge event with proper formatting
func BuildCustomEvent(source string, detailType string, detail map[string]interface{}) CustomEvent {
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		detailJSON = []byte("{}")
	}
	
	return CustomEvent{
		Source:       source,
		DetailType:   detailType,
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
	}
}

// BuildScheduledEvent creates a scheduled EventBridge event
func BuildScheduledEvent(detailType string, detail map[string]interface{}) CustomEvent {
	return BuildCustomEvent("aws.events", detailType, detail)
}

// ValidateEventPattern validates that an event pattern is properly formatted
func ValidateEventPattern(pattern string) bool {
	if pattern == "" {
		return false
	}
	
	var patternData map[string]interface{}
	if err := json.Unmarshal([]byte(pattern), &patternData); err != nil {
		return false
	}
	
	// Basic validation - pattern should be an object with valid keys
	validKeys := map[string]bool{
		"source":        true,
		"detail-type":   true,
		"detail":        true,
		"account":       true,
		"region":        true,
		"time":          true,
		"id":            true,
		"resources":     true,
		"version":       true,
	}
	
	// Check that values are in correct format (arrays for most keys)
	for key, value := range patternData {
		if !validKeys[key] {
			return false
		}
		
		// For most fields, values should be arrays
		switch key {
		case "source", "detail-type", "account", "region", "resources":
			// These should be arrays of strings
			if arr, ok := value.([]interface{}); ok {
				for _, item := range arr {
					if _, isString := item.(string); !isString {
						return false
					}
				}
			} else {
				// Single strings are not valid for patterns
				return false
			}
		case "detail":
			// Detail can be an object with nested matching rules
			if _, ok := value.(map[string]interface{}); !ok {
				return false
			}
		}
	}
	
	return true
}

// safeString safely dereferences a string pointer
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}