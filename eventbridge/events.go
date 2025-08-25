package eventbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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

// ValidateEventPatternE validates an event pattern and returns error
func ValidateEventPatternE(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("event pattern cannot be empty")
	}
	
	var patternData map[string]interface{}
	if err := json.Unmarshal([]byte(pattern), &patternData); err != nil {
		return fmt.Errorf("invalid pattern JSON: %v", err)
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
			return fmt.Errorf("invalid pattern key: %s", key)
		}
		
		// For most fields, values should be arrays
		switch key {
		case "source", "detail-type", "account", "region", "resources":
			// These should be arrays of strings
			if arr, ok := value.([]interface{}); ok {
				for _, item := range arr {
					if _, isString := item.(string); !isString {
						return fmt.Errorf("pattern field %s must contain only strings", key)
					}
				}
			} else {
				// Single strings are not valid for patterns
				return fmt.Errorf("pattern field %s must be an array", key)
			}
		case "detail":
			// Detail can be an object with nested matching rules
			if _, ok := value.(map[string]interface{}); !ok {
				return fmt.Errorf("pattern field detail must be an object")
			}
		}
	}
	
	return nil
}

// TestEventPattern tests if an event matches a pattern
func TestEventPattern(pattern string, event map[string]interface{}) bool {
	matches, err := TestEventPatternE(pattern, event)
	if err != nil {
		return false
	}
	return matches
}

// TestEventPatternE tests if an event matches a pattern and returns error
func TestEventPatternE(pattern string, event map[string]interface{}) (bool, error) {
	if pattern == "" {
		return false, fmt.Errorf("event pattern cannot be empty")
	}
	
	var patternMap map[string]interface{}
	if err := json.Unmarshal([]byte(pattern), &patternMap); err != nil {
		return false, fmt.Errorf("invalid pattern JSON: %v", err)
	}
	
	return matchesEventPattern(patternMap, event), nil
}

// BuildEventPattern creates an event pattern from a map structure
func BuildEventPattern(patternMap map[string]interface{}) string {
	bytes, _ := json.Marshal(patternMap)
	return string(bytes)
}

// BuildSourcePattern creates a pattern that matches specific sources
func BuildSourcePattern(sources ...string) string {
	sourceList := make([]interface{}, len(sources))
	for i, source := range sources {
		sourceList[i] = source
	}
	pattern := map[string]interface{}{
		"source": sourceList,
	}
	return BuildEventPattern(pattern)
}

// BuildDetailTypePattern creates a pattern that matches specific detail types
func BuildDetailTypePattern(detailTypes ...string) string {
	typeList := make([]interface{}, len(detailTypes))
	for i, detailType := range detailTypes {
		typeList[i] = detailType
	}
	pattern := map[string]interface{}{
		"detail-type": typeList,
	}
	return BuildEventPattern(pattern)
}

// BuildNumericPattern creates a pattern with numeric comparison
func BuildNumericPattern(fieldPath string, operator string, value float64) string {
	numericCondition := []interface{}{operator, value}
	
	pattern := buildNestedPattern(fieldPath, []interface{}{
		map[string]interface{}{"numeric": numericCondition},
	})
	
	return BuildEventPattern(pattern)
}

// BuildExistsPattern creates a pattern that checks for field existence
func BuildExistsPattern(fieldPath string, shouldExist bool) string {
	pattern := buildNestedPattern(fieldPath, []interface{}{
		map[string]interface{}{"exists": shouldExist},
	})
	
	return BuildEventPattern(pattern)
}

// BuildPrefixPattern creates a pattern that matches string prefixes
func BuildPrefixPattern(fieldPath string, prefix string) string {
	pattern := buildNestedPattern(fieldPath, []interface{}{
		map[string]interface{}{"prefix": prefix},
	})
	
	return BuildEventPattern(pattern)
}

// buildNestedPattern creates nested pattern structure for field paths like "detail.amount"
func buildNestedPattern(fieldPath string, condition []interface{}) map[string]interface{} {
	parts := strings.Split(fieldPath, ".")
	if len(parts) == 1 {
		return map[string]interface{}{
			parts[0]: condition,
		}
	}
	
	// Build nested structure
	result := map[string]interface{}{}
	current := result
	
	for i, part := range parts[:len(parts)-1] {
		if i == len(parts)-2 {
			// Last level - add the condition
			current[part] = map[string]interface{}{
				parts[len(parts)-1]: condition,
			}
		} else {
			// Intermediate level - create nested map
			current[part] = map[string]interface{}{}
			current = current[part].(map[string]interface{})
		}
	}
	
	return result
}

// PatternBuilder helps build complex event patterns
type PatternBuilder struct {
	pattern map[string]interface{}
}

// NewPatternBuilder creates a new pattern builder
func NewPatternBuilder() *PatternBuilder {
	return &PatternBuilder{
		pattern: make(map[string]interface{}),
	}
}

// AddSource adds sources to the pattern
func (pb *PatternBuilder) AddSource(sources ...string) *PatternBuilder {
	sourceList := make([]interface{}, len(sources))
	for i, source := range sources {
		sourceList[i] = source
	}
	pb.pattern["source"] = sourceList
	return pb
}

// AddDetailType adds detail types to the pattern
func (pb *PatternBuilder) AddDetailType(detailTypes ...string) *PatternBuilder {
	typeList := make([]interface{}, len(detailTypes))
	for i, detailType := range detailTypes {
		typeList[i] = detailType
	}
	pb.pattern["detail-type"] = typeList
	return pb
}

// AddNumericCondition adds a numeric condition to the pattern
func (pb *PatternBuilder) AddNumericCondition(fieldPath string, operator string, value float64) *PatternBuilder {
	numericCondition := []interface{}{operator, value}
	pb.addDetailCondition(fieldPath, []interface{}{
		map[string]interface{}{"numeric": numericCondition},
	})
	return pb
}

// AddExistsCondition adds an exists condition to the pattern
func (pb *PatternBuilder) AddExistsCondition(fieldPath string, shouldExist bool) *PatternBuilder {
	pb.addDetailCondition(fieldPath, []interface{}{
		map[string]interface{}{"exists": shouldExist},
	})
	return pb
}

// addDetailCondition adds a condition to the detail section
func (pb *PatternBuilder) addDetailCondition(fieldPath string, condition []interface{}) {
	if pb.pattern["detail"] == nil {
		pb.pattern["detail"] = make(map[string]interface{})
	}
	
	detail := pb.pattern["detail"].(map[string]interface{})
	parts := strings.Split(fieldPath, ".")
	
	// Remove "detail" prefix if present
	if parts[0] == "detail" {
		parts = parts[1:]
	}
	
	current := detail
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - add condition
			current[part] = condition
		} else {
			// Create nested structure
			if current[part] == nil {
				current[part] = make(map[string]interface{})
			}
			current = current[part].(map[string]interface{})
		}
	}
}

// Build returns the built pattern as JSON string
func (pb *PatternBuilder) Build() string {
	return BuildEventPattern(pb.pattern)
}

// OptimizeEventSizeE optimizes event size by compacting JSON and removing unnecessary whitespace
func OptimizeEventSizeE(event CustomEvent) (CustomEvent, error) {
	// Parse and reformat JSON to remove whitespace
	if event.Detail != "" {
		var detail interface{}
		if err := json.Unmarshal([]byte(event.Detail), &detail); err != nil {
			return event, fmt.Errorf("invalid event detail JSON: %v", err)
		}
		
		// Re-marshal without indentation
		compactJSON, err := json.Marshal(detail)
		if err != nil {
			return event, fmt.Errorf("failed to compact JSON: %v", err)
		}
		
		event.Detail = string(compactJSON)
	}
	
	return event, nil
}

// CalculateEventSize calculates the approximate size of an event in bytes
func CalculateEventSize(event CustomEvent) int {
	size := 0
	size += len(event.Source)
	size += len(event.DetailType)
	size += len(event.Detail)
	size += len(event.EventBusName)
	
	// Add overhead for other fields and metadata
	size += 100 // Approximate overhead
	
	return size
}

// ValidateEventSizeE validates that an event is within size limits
func ValidateEventSizeE(event CustomEvent) error {
	const maxEventSize = 256 * 1024 // 256KB limit
	
	size := CalculateEventSize(event)
	if size > maxEventSize {
		return fmt.Errorf("event size too large: %d bytes exceeds maximum %d bytes", size, maxEventSize)
	}
	
	return nil
}

// CreateEventBatches splits events into batches of specified size
func CreateEventBatches(events []CustomEvent, batchSize int) [][]CustomEvent {
	if batchSize <= 0 {
		batchSize = 10 // Default batch size
	}
	
	var batches [][]CustomEvent
	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}
		batches = append(batches, events[i:end])
	}
	
	return batches
}

// ValidateEventBatchE validates a batch of events
func ValidateEventBatchE(events []CustomEvent) error {
	if len(events) == 0 {
		return fmt.Errorf("batch cannot be empty")
	}
	
	const maxBatchSize = 10
	if len(events) > maxBatchSize {
		return fmt.Errorf("batch size exceeds maximum: %d events", maxBatchSize)
	}
	
	// Validate each event in the batch
	for i, event := range events {
		if err := ValidateEventSizeE(event); err != nil {
			return fmt.Errorf("event %d in batch is invalid: %v", i, err)
		}
	}
	
	return nil
}

// AnalyzePatternComplexity analyzes the complexity of an event pattern
func AnalyzePatternComplexity(pattern string) string {
	var patternMap map[string]interface{}
	if err := json.Unmarshal([]byte(pattern), &patternMap); err != nil {
		return "invalid"
	}
	
	complexity := 0
	
	// Count top-level fields
	complexity += len(patternMap)
	
	// Analyze nested complexity
	for _, value := range patternMap {
		complexity += analyzeValueComplexity(value)
	}
	
	if complexity <= 3 {
		return "low"
	} else if complexity <= 8 {
		return "medium"
	} else {
		return "high"
	}
}

// analyzeValueComplexity recursively analyzes value complexity
func analyzeValueComplexity(value interface{}) int {
	switch v := value.(type) {
	case []interface{}:
		complexity := len(v)
		for _, item := range v {
			complexity += analyzeValueComplexity(item)
		}
		return complexity
	case map[string]interface{}:
		complexity := len(v)
		for _, item := range v {
			complexity += analyzeValueComplexity(item)
		}
		return complexity
	default:
		return 1
	}
}

// EstimatePatternMatchRate estimates how often a pattern would match events
func EstimatePatternMatchRate(pattern string) float64 {
	var patternMap map[string]interface{}
	if err := json.Unmarshal([]byte(pattern), &patternMap); err != nil {
		return 0.0
	}
	
	// Simple heuristic: more specific patterns have lower match rates
	specificity := 0.0
	
	// Each field reduces match rate
	specificity += float64(len(patternMap)) * 0.1
	
	// Analyze field specificity
	for key, value := range patternMap {
		switch key {
		case "source":
			if arr, ok := value.([]interface{}); ok {
				specificity += float64(len(arr)) * 0.1
			}
		case "detail-type":
			if arr, ok := value.([]interface{}); ok {
				specificity += float64(len(arr)) * 0.1
			}
		case "detail":
			if obj, ok := value.(map[string]interface{}); ok {
				specificity += float64(len(obj)) * 0.2
			}
		}
	}
	
	// Convert specificity to match rate (inverse relationship)
	matchRate := 1.0 / (1.0 + specificity)
	
	// Ensure reasonable bounds
	if matchRate > 0.9 {
		matchRate = 0.9
	}
	if matchRate < 0.1 {
		matchRate = 0.1
	}
	
	return matchRate
}

// matchesEventPattern checks if an event matches a pattern recursively
func matchesEventPattern(pattern map[string]interface{}, event map[string]interface{}) bool {
	for key, patternValue := range pattern {
		eventValue, exists := event[key]
		if !exists {
			return false
		}
		
		if !matchesPatternValue(patternValue, eventValue) {
			return false
		}
	}
	return true
}

// matchesPatternValue checks if an event value matches a pattern value
func matchesPatternValue(patternValue, eventValue interface{}) bool {
	// Pattern values can be arrays or nested objects
	switch pv := patternValue.(type) {
	case []interface{}:
		return matchesPatternArray(pv, eventValue)
	case map[string]interface{}:
		if ev, ok := eventValue.(map[string]interface{}); ok {
			return matchesEventPattern(pv, ev)
		}
		return false
	default:
		// Direct value comparison
		return reflect.DeepEqual(patternValue, eventValue)
	}
}

// matchesPatternArray checks if an event value matches any item in a pattern array
func matchesPatternArray(patternArray []interface{}, eventValue interface{}) bool {
	for _, patternItem := range patternArray {
		if matchesPatternItem(patternItem, eventValue) {
			return true
		}
	}
	return false
}

// matchesPatternItem checks if an event value matches a specific pattern item
func matchesPatternItem(patternItem, eventValue interface{}) bool {
	// Check for special pattern operators
	if patternMap, ok := patternItem.(map[string]interface{}); ok {
		return matchesSpecialPattern(patternMap, eventValue)
	}
	
	// Direct value comparison
	return reflect.DeepEqual(patternItem, eventValue)
}

// matchesSpecialPattern handles special EventBridge pattern operators
func matchesSpecialPattern(patternMap map[string]interface{}, eventValue interface{}) bool {
	// Handle "numeric" patterns
	if numeric, exists := patternMap["numeric"]; exists {
		return matchesNumericPattern(numeric, eventValue)
	}
	
	// Handle "exists" patterns
	if exists, existsOk := patternMap["exists"]; existsOk {
		if shouldExist, ok := exists.(bool); ok {
			return shouldExist == (eventValue != nil)
		}
	}
	
	// Handle "prefix" patterns
	if prefix, exists := patternMap["prefix"]; exists {
		if prefixStr, ok := prefix.(string); ok {
			if eventStr, ok := eventValue.(string); ok {
				return strings.HasPrefix(eventStr, prefixStr)
			}
		}
	}
	
	return false
}

// matchesNumericPattern handles numeric comparison patterns
func matchesNumericPattern(numericPattern, eventValue interface{}) bool {
	numArray, ok := numericPattern.([]interface{})
	if !ok || len(numArray) != 2 {
		return false
	}
	
	operator, ok := numArray[0].(string)
	if !ok {
		return false
	}
	
	// Convert event value to float64 for comparison
	eventFloat, err := toFloat64(eventValue)
	if err != nil {
		return false
	}
	
	// Convert pattern value to float64
	patternFloat, err := toFloat64(numArray[1])
	if err != nil {
		return false
	}
	
	switch operator {
	case ">":
		return eventFloat > patternFloat
	case ">=":
		return eventFloat >= patternFloat
	case "<":
		return eventFloat < patternFloat
	case "<=":
		return eventFloat <= patternFloat
	case "=":
		return eventFloat == patternFloat
	case "!=":
		return eventFloat != patternFloat
	default:
		return false
	}
}

// toFloat64 converts various numeric types to float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// safeString safely dereferences a string pointer
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}