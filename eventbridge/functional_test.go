package eventbridge

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/samber/lo"
)

// TestFunctionalEventBridgeScenario tests a complete EventBridge scenario
func TestFunctionalEventBridgeScenario(t *testing.T) {
	// Create event bus configuration
	config := NewFunctionalEventBridgeConfig("test-event-bus",
		WithEventBridgeTimeout(30*time.Second),
		WithEventBridgeRetryConfig(3, 200*time.Millisecond),
		WithEventPattern(`{"source": ["test.service"]}`),
		WithRuleState(types.RuleStateEnabled),
		WithEventBridgeMetadata("environment", "test"),
		WithEventBridgeMetadata("scenario", "full_workflow"),
	)

	// Step 1: Create custom event bus
	createBusResult := SimulateFunctionalEventBridgeCreateEventBus(config)

	if !createBusResult.IsSuccess() {
		t.Errorf("Expected event bus creation to succeed, got error: %v", createBusResult.GetError().MustGet())
	}

	// Step 2: Create rule
	rule := NewFunctionalEventBridgeRule("test-rule", "test-event-bus")
	createRuleResult := SimulateFunctionalEventBridgeCreateRule(rule, config)

	if !createRuleResult.IsSuccess() {
		t.Errorf("Expected rule creation to succeed, got error: %v", createRuleResult.GetError().MustGet())
	}

	// Step 3: Create and put single event
	event := NewFunctionalEventBridgeEvent("test.service", "Order Created", `{"orderId": "12345", "amount": 99.99}`)
	event = event.WithEventBusName("test-event-bus").WithResources([]string{"order-service"})

	putEventResult := SimulateFunctionalEventBridgePutEvent(event, config)

	if !putEventResult.IsSuccess() {
		t.Errorf("Expected put event to succeed, got error: %v", putEventResult.GetError().MustGet())
	}

	// Verify single event result
	if putEventResult.GetResult().IsPresent() {
		result, ok := putEventResult.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected put event result to be map[string]interface{}")
		} else {
			if result["source"] != "test.service" {
				t.Errorf("Expected source to be 'test.service', got %v", result["source"])
			}
			if result["success"] != true {
				t.Error("Expected event to be successful")
			}
		}
	}

	// Step 4: Put multiple events
	events := []FunctionalEventBridgeEvent{
		NewFunctionalEventBridgeEvent("test.service", "User Registered", `{"userId": "user-001", "email": "user1@example.com"}`).
			WithEventBusName("test-event-bus"),
		
		NewFunctionalEventBridgeEvent("test.service", "Payment Processed", `{"paymentId": "pay-001", "amount": 49.99}`).
			WithEventBusName("test-event-bus").
			WithResources([]string{"payment-service"}),
		
		NewFunctionalEventBridgeEvent("test.service", "Order Shipped", `{"orderId": "12345", "trackingId": "track-001"}`).
			WithEventBusName("test-event-bus").
			WithResources([]string{"shipping-service"}),
	}

	putEventsResult := SimulateFunctionalEventBridgePutEvents(events, config)

	if !putEventsResult.IsSuccess() {
		t.Errorf("Expected put events to succeed, got error: %v", putEventsResult.GetError().MustGet())
	}

	// Verify batch events result
	if putEventsResult.GetResult().IsPresent() {
		result, ok := putEventsResult.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected put events result to be map[string]interface{}")
		} else {
			if result["failed_entry_count"] != int32(0) {
				t.Errorf("Expected no failed entries, got %v", result["failed_entry_count"])
			}
			if result["total_events"] != len(events) {
				t.Errorf("Expected %d total events, got %v", len(events), result["total_events"])
			}
		}
	}

	// Step 5: Verify metadata accumulation across operations
	allResults := []FunctionalEventBridgeOperationResult{
		createBusResult,
		createRuleResult,
		putEventResult,
		putEventsResult,
	}

	for i, result := range allResults {
		metadata := result.GetMetadata()
		if metadata["environment"] != "test" {
			t.Errorf("Operation %d missing environment metadata, metadata: %+v", i, metadata)
		}
		if result.GetDuration() == 0 {
			t.Errorf("Operation %d should have recorded duration", i)
		}
	}
}

// TestFunctionalEventBridgeErrorHandling tests comprehensive error handling
func TestFunctionalEventBridgeErrorHandling(t *testing.T) {
	// Test invalid event bus name
	invalidConfig := NewFunctionalEventBridgeConfig("")
	
	result := SimulateFunctionalEventBridgeCreateEventBus(invalidConfig)

	if result.IsSuccess() {
		t.Error("Expected operation to fail with empty event bus name")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	// Test invalid event source
	validConfig := NewFunctionalEventBridgeConfig("valid-bus")
	invalidEvent := NewFunctionalEventBridgeEvent("", "Test Event", `{"test": true}`)

	result2 := SimulateFunctionalEventBridgePutEvent(invalidEvent, validConfig)

	if result2.IsSuccess() {
		t.Error("Expected operation to fail with empty event source")
	}

	// Test invalid JSON in event detail
	invalidJsonEvent := NewFunctionalEventBridgeEvent("test.service", "Test Event", `{"invalid": json}`)

	result3 := SimulateFunctionalEventBridgePutEvent(invalidJsonEvent, validConfig)

	if result3.IsSuccess() {
		t.Error("Expected operation to fail with invalid JSON detail")
	}

	// Test batch size exceeding maximum
	tooManyEvents := make([]FunctionalEventBridgeEvent, FunctionalEventBridgeMaxEventBatchSize+1)
	for i := range tooManyEvents {
		tooManyEvents[i] = NewFunctionalEventBridgeEvent("test.service", "Event", `{}`)
	}

	result4 := SimulateFunctionalEventBridgePutEvents(tooManyEvents, validConfig)

	if result4.IsSuccess() {
		t.Error("Expected operation to fail with too many events")
	}

	// Verify error metadata consistency
	errorResults := []FunctionalEventBridgeOperationResult{result, result2, result3, result4}
	for i, errorResult := range errorResults {
		metadata := errorResult.GetMetadata()
		if metadata["phase"] != "validation" {
			t.Errorf("Error result %d should have validation phase, got %v", i, metadata["phase"])
		}
	}
}

// TestFunctionalEventBridgeConfigurationCombinations tests various config combinations
func TestFunctionalEventBridgeConfigurationCombinations(t *testing.T) {
	// Test minimal configuration
	minimalConfig := NewFunctionalEventBridgeConfig("minimal-bus")
	
	if minimalConfig.eventBusName != "minimal-bus" {
		t.Error("Expected minimal config to have correct event bus name")
	}
	
	if minimalConfig.timeout != FunctionalEventBridgeDefaultTimeout {
		t.Error("Expected minimal config to use default timeout")
	}

	// Test maximal configuration
	maximalConfig := NewFunctionalEventBridgeConfig("maximal-bus",
		WithEventBridgeTimeout(120*time.Second),
		WithEventBridgeRetryConfig(10, 5*time.Second),
		WithEventPattern(`{"source": ["service1", "service2"], "detail-type": ["Order"]}`),
		WithScheduleExpression("rate(5 minutes)"),
		WithRuleState(types.RuleStateDisabled),
		WithDeadLetterConfig(types.DeadLetterConfig{
			Arn: lo.ToPtr("arn:aws:sqs:us-east-1:123456789012:dlq"),
		}),
		WithRetryPolicy(types.RetryPolicy{
			MaximumRetryAttempts: lo.ToPtr(int32(5)),
			MaximumEventAge: lo.ToPtr(int32(3600)),
		}),
		WithEventBridgeMetadata("service", "event-service"),
		WithEventBridgeMetadata("environment", "production"),
		WithEventBridgeMetadata("version", "v1.2.3"),
	)

	// Verify all options were applied
	if maximalConfig.eventBusName != "maximal-bus" {
		t.Error("Expected maximal config event bus name")
	}

	if maximalConfig.timeout != 120*time.Second {
		t.Error("Expected maximal config timeout")
	}

	if maximalConfig.maxRetries != 10 {
		t.Error("Expected maximal config retries")
	}

	if maximalConfig.retryDelay != 5*time.Second {
		t.Error("Expected maximal config retry delay")
	}

	if maximalConfig.eventPattern.IsAbsent() {
		t.Error("Expected maximal config event pattern")
	}

	if maximalConfig.scheduleExpression.IsAbsent() {
		t.Error("Expected maximal config schedule expression")
	}

	if maximalConfig.ruleState != types.RuleStateDisabled {
		t.Error("Expected maximal config rule state")
	}

	if maximalConfig.deadLetterConfig.IsAbsent() {
		t.Error("Expected maximal config dead letter config")
	}

	if maximalConfig.retryPolicy.IsAbsent() {
		t.Error("Expected maximal config retry policy")
	}

	if len(maximalConfig.metadata) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(maximalConfig.metadata))
	}
}

// TestFunctionalEventBridgeEventManipulation tests advanced event operations
func TestFunctionalEventBridgeEventManipulation(t *testing.T) {
	// Create events with different configurations
	basicEvent := NewFunctionalEventBridgeEvent("myapp.orders", "Order Created", `{"orderId": "123", "total": 99.99}`)
	
	complexEvent := basicEvent.
		WithEventBusName("orders-bus").
		WithResources([]string{"arn:aws:lambda:us-east-1:123456789012:function:ProcessOrder"}).
		WithRegion("us-east-1").
		WithAccount("123456789012").
		WithEventId("evt-12345")

	events := []FunctionalEventBridgeEvent{basicEvent, complexEvent}

	// Test event properties
	for i, event := range events {
		if event.GetSource() == "" {
			t.Errorf("Event %d should have non-empty source", i)
		}

		if event.GetDetailType() == "" {
			t.Errorf("Event %d should have non-empty detail type", i)
		}

		if event.GetDetail() == "" {
			t.Errorf("Event %d should have non-empty detail", i)
		}

		if event.GetTime().IsZero() {
			t.Errorf("Event %d should have valid time", i)
		}
	}

	// Test complex event specific properties
	if complexEvent.GetEventBusName() != "orders-bus" {
		t.Error("Expected complex event to have orders-bus")
	}

	if len(complexEvent.GetResources()) != 1 {
		t.Error("Expected complex event to have one resource")
	}

	if complexEvent.GetRegion().IsAbsent() {
		t.Error("Expected complex event to have region")
	}

	if complexEvent.GetAccount().IsAbsent() {
		t.Error("Expected complex event to have account")
	}

	if complexEvent.GetEventId().IsAbsent() {
		t.Error("Expected complex event to have event ID")
	}

	// Test event operations
	config := NewFunctionalEventBridgeConfig("event-test-bus",
		WithEventBridgeMetadata("test_type", "event_manipulation"),
	)

	for i, event := range events {
		result := SimulateFunctionalEventBridgePutEvent(event, config)

		if !result.IsSuccess() {
			t.Errorf("Expected event %d put to succeed, got error: %v", i, result.GetError().MustGet())
		}

		// Verify result includes event information
		if result.GetResult().IsPresent() {
			resultMap, ok := result.GetResult().MustGet().(map[string]interface{})
			if ok {
				if resultMap["source"] != event.GetSource() {
					t.Errorf("Expected event %d source in result to match", i)
				}
				if resultMap["detail_type"] != event.GetDetailType() {
					t.Errorf("Expected event %d detail type in result to match", i)
				}
			}
		}
	}
}

// TestFunctionalEventBridgeRuleCreation tests EventBridge rule creation
func TestFunctionalEventBridgeRuleCreation(t *testing.T) {
	config := NewFunctionalEventBridgeConfig("rules-bus",
		WithEventPattern(`{"source": ["test.service"]}`),
		WithEventBridgeMetadata("test_type", "rule_creation"),
	)

	// Test basic rule creation
	basicRule := NewFunctionalEventBridgeRule("basic-rule", "rules-bus")

	result := SimulateFunctionalEventBridgeCreateRule(basicRule, config)

	if !result.IsSuccess() {
		t.Errorf("Expected rule creation to succeed, got error: %v", result.GetError().MustGet())
	}

	// Verify rule result
	if result.GetResult().IsPresent() {
		ruleResult, ok := result.GetResult().MustGet().(map[string]interface{})
		if !ok {
			t.Error("Expected rule result to be map[string]interface{}")
		} else {
			if ruleResult["rule_name"] != "basic-rule" {
				t.Errorf("Expected rule name to be 'basic-rule', got %v", ruleResult["rule_name"])
			}
			if ruleResult["event_bus"] != "rules-bus" {
				t.Errorf("Expected event bus to be 'rules-bus', got %v", ruleResult["event_bus"])
			}
			if ruleResult["created"] != true {
				t.Error("Expected rule to be created")
			}
		}
	}

	// Test rule with invalid name
	invalidRule := NewFunctionalEventBridgeRule("", "rules-bus")
	invalidResult := SimulateFunctionalEventBridgeCreateRule(invalidRule, config)

	if invalidResult.IsSuccess() {
		t.Error("Expected rule creation to fail with empty name")
	}
}

// TestFunctionalEventBridgeValidationRules tests EventBridge-specific validation
func TestFunctionalEventBridgeValidationRules(t *testing.T) {
	// Test event bus name validation
	invalidBusNames := []string{
		"",                    // Empty
		".starts-with-dot",    // Starts with dot
		"ends-with-dot.",      // Ends with dot
		"-starts-with-dash",   // Starts with dash
		"ends-with-dash-",     // Ends with dash
		"invalid@character",   // Contains invalid character
	}

	for _, busName := range invalidBusNames {
		config := NewFunctionalEventBridgeConfig(busName)
		result := SimulateFunctionalEventBridgeCreateEventBus(config)
		
		if result.IsSuccess() {
			t.Errorf("Expected event bus creation to fail for invalid name: %s", busName)
		}
	}

	// Test valid event bus names
	validBusNames := []string{
		"valid-bus",
		"test.bus.com",
		"bus123",
		"my_awesome_bus_2024",
		"default",
	}

	for _, busName := range validBusNames {
		config := NewFunctionalEventBridgeConfig(busName)
		result := SimulateFunctionalEventBridgeCreateEventBus(config)
		
		if !result.IsSuccess() {
			t.Errorf("Expected event bus creation to succeed for valid name: %s", busName)
		}
	}

	// Test event source validation
	invalidSources := []string{
		"",                    // Empty
		"invalid@source",      // Contains invalid character
		"source_with_underscore", // Contains underscore (not typical)
	}

	config := NewFunctionalEventBridgeConfig("valid-bus")
	
	for _, source := range invalidSources {
		event := NewFunctionalEventBridgeEvent(source, "Test Event", `{}`)
		result := SimulateFunctionalEventBridgePutEvent(event, config)
		
		if result.IsSuccess() {
			t.Errorf("Expected put event to fail for invalid source: %s", source)
		}
	}

	// Test event detail size validation
	largeDetail := strings.Repeat(`{"key": "value"}`, 100000) // Large JSON
	largeEvent := NewFunctionalEventBridgeEvent("test.service", "Large Event", largeDetail)
	
	result := SimulateFunctionalEventBridgePutEvent(largeEvent, config)
	
	if result.IsSuccess() {
		t.Error("Expected put event to fail for large detail")
	}
}

// TestFunctionalEventBridgeScheduleExpressions tests schedule expression validation
func TestFunctionalEventBridgeScheduleExpressions(t *testing.T) {
	validExpressions := []string{
		"rate(5 minutes)",
		"rate(1 hour)",
		"cron(0 12 * * ? *)",
		"cron(15 10 ? * 6L 2002-2005)",
	}

	invalidExpressions := []string{
		"invalid expression",
		"rate(5 minutes", // Missing closing parenthesis
		"cron0 12 * * ? *)", // Missing opening parenthesis
		"schedule(5 minutes)", // Wrong prefix
	}

	eventBusName := "schedule-test-bus"

	// Test valid expressions
	for _, expr := range validExpressions {
		config := NewFunctionalEventBridgeConfig(eventBusName,
			WithScheduleExpression(expr),
		)

		result := SimulateFunctionalEventBridgeCreateEventBus(config)
		
		if !result.IsSuccess() {
			t.Errorf("Expected success for valid schedule expression: %s", expr)
		}
	}

	// Test invalid expressions
	for _, expr := range invalidExpressions {
		config := NewFunctionalEventBridgeConfig(eventBusName,
			WithScheduleExpression(expr),
		)

		result := SimulateFunctionalEventBridgeCreateEventBus(config)
		
		if result.IsSuccess() {
			t.Errorf("Expected failure for invalid schedule expression: %s", expr)
		}
	}
}

// TestFunctionalEventBridgePerformanceMetrics tests performance tracking
func TestFunctionalEventBridgePerformanceMetrics(t *testing.T) {
	config := NewFunctionalEventBridgeConfig("performance-bus",
		WithEventBridgeTimeout(1*time.Second),
		WithEventBridgeRetryConfig(1, 50*time.Millisecond), // Faster for testing
		WithEventBridgeMetadata("test_type", "performance"),
	)

	event := NewFunctionalEventBridgeEvent("perf.service", "Performance Test", `{"test": "data"}`)

	startTime := time.Now()
	result := SimulateFunctionalEventBridgePutEvent(event, config)
	operationTime := time.Since(startTime)

	if !result.IsSuccess() {
		t.Errorf("Expected performance test to succeed, got error: %v", result.GetError().MustGet())
	}

	// Verify duration tracking
	recordedDuration := result.GetDuration()
	if recordedDuration == 0 {
		t.Error("Expected duration to be recorded")
	}

	// Duration should be reasonable (less than total operation time)
	if recordedDuration > operationTime {
		t.Error("Recorded duration should not exceed total operation time")
	}

	// Test multiple operations for performance consistency
	results := []FunctionalEventBridgeOperationResult{}
	for i := 0; i < 5; i++ {
		testEvent := NewFunctionalEventBridgeEvent("perf.service", "Batch Test", 
			fmt.Sprintf(`{"index": %d, "timestamp": "%s"}`, i, time.Now().Format(time.RFC3339)))
		
		result := SimulateFunctionalEventBridgePutEvent(testEvent, config)
		results = append(results, result)
	}

	// All operations should succeed
	allSucceeded := lo.EveryBy(results, func(r FunctionalEventBridgeOperationResult) bool {
		return r.IsSuccess()
	})

	if !allSucceeded {
		t.Error("Expected all performance test operations to succeed")
	}

	// All operations should have duration recorded
	allHaveDuration := lo.EveryBy(results, func(r FunctionalEventBridgeOperationResult) bool {
		return r.GetDuration() > 0
	})

	if !allHaveDuration {
		t.Error("Expected all operations to have duration recorded")
	}
}

// BenchmarkFunctionalEventBridgeScenario benchmarks a complete scenario
func BenchmarkFunctionalEventBridgeScenario(b *testing.B) {
	config := NewFunctionalEventBridgeConfig("benchmark-bus",
		WithEventBridgeTimeout(5*time.Second),
		WithEventBridgeRetryConfig(1, 10*time.Millisecond),
	)

	event := NewFunctionalEventBridgeEvent("bench.service", "Benchmark Event", `{"benchmark": true}`)
	rule := NewFunctionalEventBridgeRule("benchmark-rule", "benchmark-bus")

	events := []FunctionalEventBridgeEvent{
		event,
		event.WithEventId("batch-1"),
		event.WithEventId("batch-2"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate full EventBridge workflow
		createBusResult := SimulateFunctionalEventBridgeCreateEventBus(config)
		if !createBusResult.IsSuccess() {
			b.Errorf("Create event bus operation failed: %v", createBusResult.GetError().MustGet())
		}

		createRuleResult := SimulateFunctionalEventBridgeCreateRule(rule, config)
		if !createRuleResult.IsSuccess() {
			b.Errorf("Create rule operation failed: %v", createRuleResult.GetError().MustGet())
		}

		putEventResult := SimulateFunctionalEventBridgePutEvent(event, config)
		if !putEventResult.IsSuccess() {
			b.Errorf("Put event operation failed: %v", putEventResult.GetError().MustGet())
		}

		putEventsResult := SimulateFunctionalEventBridgePutEvents(events, config)
		if !putEventsResult.IsSuccess() {
			b.Errorf("Put events operation failed: %v", putEventsResult.GetError().MustGet())
		}
	}
}