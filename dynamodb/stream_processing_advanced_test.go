package dynamodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StreamEventProcessor simulates advanced DynamoDB streams processing
type StreamEventProcessor struct {
	MaxRetries       int
	BackoffMultiplier time.Duration
	ErrorThreshold   float64
	ProcessingStats  ProcessingStatistics
}

type ProcessingStatistics struct {
	TotalEvents     int64
	SuccessfulEvents int64
	FailedEvents     int64
	RetryAttempts    int64
	ErrorRate       float64
}

// StreamEvent represents a DynamoDB stream event with comprehensive metadata
type StreamEvent struct {
	EventID        string
	EventName      string
	EventSource    string
	EventVersion   string
	EventSourceARN string
	Timestamp      time.Time
	DynamoDB       StreamDynamoDBData
	UserIdentity   *StreamUserIdentity
	Records        []StreamEvent
}

type StreamDynamoDBData struct {
	ApproximateCreationDateTime time.Time
	Keys                       map[string]types.AttributeValue
	NewImage                   map[string]types.AttributeValue
	OldImage                   map[string]types.AttributeValue
	SequenceNumber             string
	SizeBytes                  int64
	StreamViewType             string
}

type StreamUserIdentity struct {
	Type        string
	PrincipalID string
}

// TestStreamEventProcessing_ComplexWorkflow tests comprehensive stream processing
func TestStreamEventProcessing_ComplexWorkflow(t *testing.T) {
	// Test case: Process a stream of related events representing an e-commerce workflow
	
	processor := StreamEventProcessor{
		MaxRetries:        3,
		BackoffMultiplier: 100 * time.Millisecond,
		ErrorThreshold:    0.05, // 5% error threshold
		ProcessingStats:   ProcessingStatistics{},
	}

	// Create sequence of related events
	events := createECommerceWorkflowEvents(t)
	
	// Process events and track statistics
	for _, event := range events {
		err := processor.processEvent(event)
		if err != nil {
			processor.ProcessingStats.FailedEvents++
		} else {
			processor.ProcessingStats.SuccessfulEvents++
		}
		processor.ProcessingStats.TotalEvents++
	}

	// Calculate error rate
	processor.ProcessingStats.ErrorRate = float64(processor.ProcessingStats.FailedEvents) / float64(processor.ProcessingStats.TotalEvents)

	// Validate processing results
	assert.Equal(t, int64(len(events)), processor.ProcessingStats.TotalEvents)
	assert.LessOrEqual(t, processor.ProcessingStats.ErrorRate, processor.ErrorThreshold)
	assert.Greater(t, processor.ProcessingStats.SuccessfulEvents, int64(0))
}

// TestStreamEvent_ErrorHandling tests stream processing error scenarios
func TestStreamEvent_ErrorHandling(t *testing.T) {
	// Test case: Handle various error conditions in stream processing
	
	processor := StreamEventProcessor{
		MaxRetries:        2,
		BackoffMultiplier: 50 * time.Millisecond,
	}

	// Create events that should trigger different error conditions
	errorEvents := []StreamEvent{
		// Malformed event
		{
			EventID:   "malformed-event-1",
			EventName: "INVALID",
			DynamoDB: StreamDynamoDBData{
				Keys: nil, // Missing required keys
			},
		},
		// Large event exceeding processing limits
		{
			EventID:   "large-event-1",
			EventName: "MODIFY",
			DynamoDB: StreamDynamoDBData{
				Keys: map[string]types.AttributeValue{
					"ID": &types.AttributeValueMemberS{Value: "large-item"},
				},
				SizeBytes: 1024 * 1024, // 1MB - simulating large event
				NewImage:  createLargeAttributeMap(1000), // Large attribute map
			},
		},
		// Event with processing dependencies
		{
			EventID:   "dependency-event-1",
			EventName: "INSERT",
			DynamoDB: StreamDynamoDBData{
				Keys: map[string]types.AttributeValue{
					"ID": &types.AttributeValueMemberS{Value: "dependent-item"},
					"ParentID": &types.AttributeValueMemberS{Value: "non-existent-parent"},
				},
			},
		},
	}

	var processingErrors []error
	for _, event := range errorEvents {
		err := processor.processEventWithErrorHandling(event)
		if err != nil {
			processingErrors = append(processingErrors, err)
		}
	}

	// Validate error handling
	assert.Len(t, processingErrors, len(errorEvents), "Should have errors for all problematic events")
	
	// Check error types
	assert.Contains(t, processingErrors[0].Error(), "missing required keys")
	assert.Contains(t, processingErrors[1].Error(), "event size exceeds limit")
	assert.Contains(t, processingErrors[2].Error(), "dependency not found")
}

// TestStreamEvent_BatchProcessing tests batch processing of stream events
func TestStreamEvent_BatchProcessing(t *testing.T) {
	// Test case: Process multiple events in batches with proper ordering and dependencies
	
	batchSize := 10
	totalEvents := 50
	
	events := make([]StreamEvent, totalEvents)
	for i := 0; i < totalEvents; i++ {
		events[i] = StreamEvent{
			EventID:        fmt.Sprintf("batch-event-%d", i),
			EventName:      "MODIFY",
			EventSource:    "aws:dynamodb",
			SequenceNumber: fmt.Sprintf("%010d", i),
			Timestamp:      time.Now().Add(time.Duration(i) * time.Millisecond),
			DynamoDB: StreamDynamoDBData{
				Keys: map[string]types.AttributeValue{
					"ID": &types.AttributeValueMemberS{Value: fmt.Sprintf("item-%d", i)},
				},
				SequenceNumber: fmt.Sprintf("%010d", i),
				SizeBytes:     int64(128 + i*10),
			},
		}
	}

	// Process in batches
	batches := createEventBatches(events, batchSize)
	
	assert.Equal(t, 5, len(batches), "Should have 5 batches")
	
	for i, batch := range batches {
		if i < len(batches)-1 {
			assert.Len(t, batch, batchSize, "All batches except last should be full size")
		} else {
			assert.Len(t, batch, totalEvents%batchSize, "Last batch should have remainder")
		}
		
		// Validate sequence ordering within batch
		for j := 1; j < len(batch); j++ {
			prevSeq, _ := strconv.Atoi(batch[j-1].DynamoDB.SequenceNumber)
			currSeq, _ := strconv.Atoi(batch[j].DynamoDB.SequenceNumber)
			assert.Less(t, prevSeq, currSeq, "Events should be ordered by sequence number")
		}
	}
}

// TestStreamEvent_FilteringAndTransformation tests event filtering and transformation
func TestStreamEvent_FilteringAndTransformation(t *testing.T) {
	// Test case: Filter and transform stream events based on business rules
	
	events := createMixedStreamEvents(t)
	
	// Define filtering rules
	filterRules := StreamFilterRules{
		EventTypes: []string{"INSERT", "MODIFY"},
		TablePatterns: []string{"prod-*", "user-*"},
		AttributeFilters: map[string]interface{}{
			"Status": []string{"Active", "Pending"},
			"Amount": map[string]float64{"min": 0, "max": 10000},
		},
		SizeLimits: map[string]int64{
			"max": 1024 * 100, // 100KB limit
		},
	}

	// Apply filters
	filteredEvents := applyStreamFilters(events, filterRules)
	
	// Validate filtering results
	assert.Less(t, len(filteredEvents), len(events), "Should filter out some events")
	
	for _, event := range filteredEvents {
		// Check event type filter
		assert.Contains(t, filterRules.EventTypes, event.EventName)
		
		// Check size limits
		assert.LessOrEqual(t, event.DynamoDB.SizeBytes, filterRules.SizeLimits["max"])
	}

	// Transform events
	transformedEvents := transformStreamEvents(filteredEvents)
	
	// Validate transformations
	assert.Len(t, transformedEvents, len(filteredEvents))
	
	for _, transformed := range transformedEvents {
		// Should have additional metadata
		assert.NotNil(t, transformed.ProcessingMetadata)
		assert.NotEmpty(t, transformed.ProcessingMetadata.ProcessedAt)
		assert.NotEmpty(t, transformed.ProcessingMetadata.TransformationType)
	}
}

// TestStreamEvent_DeadLetterQueueHandling tests DLQ processing for failed events
func TestStreamEvent_DeadLetterQueueHandling(t *testing.T) {
	// Test case: Handle events that fail processing after max retries
	
	dlqProcessor := DeadLetterQueueProcessor{
		MaxRetries:     3,
		RetryDelay:     100 * time.Millisecond,
		DLQThreshold:   5, // After 5 failures, send to DLQ
		FailureReasons: make(map[string]int),
	}

	// Create events that will fail processing
	failingEvents := []StreamEvent{
		createFailingEvent("validation-failure", "Validation failed"),
		createFailingEvent("dependency-failure", "Missing dependency"),
		createFailingEvent("processing-timeout", "Processing timeout"),
		createFailingEvent("resource-unavailable", "Resource unavailable"),
		createFailingEvent("permission-denied", "Permission denied"),
		createFailingEvent("invalid-state", "Invalid state transition"),
	}

	// Process failing events
	for _, event := range failingEvents {
		result := dlqProcessor.processWithDLQ(event)
		assert.NotNil(t, result)
	}

	// Validate DLQ processing
	assert.Equal(t, len(failingEvents), len(dlqProcessor.FailureReasons))
	assert.Greater(t, dlqProcessor.TotalDLQSent, 0)
}

// TestStreamEvent_MetricsAndMonitoring tests comprehensive metrics collection
func TestStreamEvent_MetricsAndMonitoring(t *testing.T) {
	// Test case: Collect detailed metrics during stream processing
	
	metricsCollector := StreamMetricsCollector{
		StartTime:         time.Now(),
		EventMetrics:      make(map[string]*EventTypeMetrics),
		TableMetrics:      make(map[string]*TableMetrics),
		ProcessingMetrics: &ProcessingMetrics{},
	}

	events := createMetricsTestEvents(t, 100)
	
	// Process events and collect metrics
	for _, event := range events {
		startTime := time.Now()
		err := processEventWithMetrics(event, &metricsCollector)
		processingTime := time.Since(startTime)
		
		metricsCollector.recordEventProcessing(event, processingTime, err)
	}

	// Generate metrics summary
	summary := metricsCollector.generateSummary()
	
	// Validate metrics collection
	assert.Equal(t, int64(len(events)), summary.TotalEvents)
	assert.Greater(t, summary.AverageProcessingTime.Milliseconds(), int64(0))
	assert.NotEmpty(t, summary.EventTypeBreakdown)
	assert.NotEmpty(t, summary.TableBreakdown)
	
	// Validate specific metrics
	assert.Contains(t, summary.EventTypeBreakdown, "INSERT")
	assert.Contains(t, summary.EventTypeBreakdown, "MODIFY")
	assert.Contains(t, summary.EventTypeBreakdown, "REMOVE")
	
	for eventType, metrics := range summary.EventTypeBreakdown {
		assert.Greater(t, metrics.Count, int64(0), fmt.Sprintf("Event type %s should have count > 0", eventType))
		assert.GreaterOrEqual(t, metrics.SuccessRate, 0.0, "Success rate should be >= 0")
		assert.LessOrEqual(t, metrics.SuccessRate, 1.0, "Success rate should be <= 1")
	}
}

// Supporting types and functions for stream processing tests

type StreamFilterRules struct {
	EventTypes       []string
	TablePatterns    []string
	AttributeFilters map[string]interface{}
	SizeLimits      map[string]int64
}

type TransformedStreamEvent struct {
	StreamEvent
	ProcessingMetadata *ProcessingMetadata
}

type ProcessingMetadata struct {
	ProcessedAt        time.Time
	TransformationType string
	EnrichmentData     map[string]interface{}
	ValidationResults  []ValidationResult
}

type ValidationResult struct {
	Field   string
	Valid   bool
	Message string
}

type DeadLetterQueueProcessor struct {
	MaxRetries     int
	RetryDelay     time.Duration
	DLQThreshold   int
	FailureReasons map[string]int
	TotalDLQSent   int
}

type ProcessingResult struct {
	Success    bool
	Error      error
	Retries    int
	SentToDLQ  bool
	ProcessingTime time.Duration
}

type StreamMetricsCollector struct {
	StartTime         time.Time
	EventMetrics      map[string]*EventTypeMetrics
	TableMetrics      map[string]*TableMetrics
	ProcessingMetrics *ProcessingMetrics
}

type EventTypeMetrics struct {
	Count               int64
	TotalProcessingTime time.Duration
	AverageProcessingTime time.Duration
	SuccessCount        int64
	FailureCount        int64
	SuccessRate         float64
}

type TableMetrics struct {
	EventCount    int64
	TotalSizeBytes int64
	AverageSize   float64
	LastProcessed time.Time
}

type ProcessingMetrics struct {
	TotalThroughput  float64
	PeakThroughput   float64
	ErrorRate        float64
	AverageLatency   time.Duration
}

type MetricsSummary struct {
	TotalEvents           int64
	ProcessingDuration    time.Duration
	AverageProcessingTime time.Duration
	TotalThroughput       float64
	ErrorRate            float64
	EventTypeBreakdown   map[string]*EventTypeMetrics
	TableBreakdown       map[string]*TableMetrics
}

// Helper functions for stream processing tests

func (p *StreamEventProcessor) processEvent(event StreamEvent) error {
	// Simulate processing logic with potential failures
	if event.EventName == "INVALID" || event.DynamoDB.Keys == nil {
		return errors.New("invalid event structure")
	}
	
	if event.DynamoDB.SizeBytes > 500*1024 { // 500KB limit
		return errors.New("event size exceeds processing limit")
	}
	
	// Simulate processing time
	time.Sleep(1 * time.Millisecond)
	
	return nil
}

func (p *StreamEventProcessor) processEventWithErrorHandling(event StreamEvent) error {
	if event.DynamoDB.Keys == nil {
		return errors.New("missing required keys")
	}
	
	if event.DynamoDB.SizeBytes > 500*1024 {
		return errors.New("event size exceeds limit")
	}
	
	if event.DynamoDB.Keys["ParentID"] != nil {
		parentID := event.DynamoDB.Keys["ParentID"].(*types.AttributeValueMemberS).Value
		if parentID == "non-existent-parent" {
			return errors.New("dependency not found")
		}
	}
	
	return nil
}

func createECommerceWorkflowEvents(t *testing.T) []StreamEvent {
	baseTime := time.Now()
	
	return []StreamEvent{
		// Order creation
		{
			EventID:   "order-created-1",
			EventName: "INSERT",
			EventSource: "aws:dynamodb",
			Timestamp: baseTime,
			DynamoDB: StreamDynamoDBData{
				Keys: map[string]types.AttributeValue{
					"OrderID": &types.AttributeValueMemberS{Value: "order-12345"},
				},
				NewImage: map[string]types.AttributeValue{
					"OrderID": &types.AttributeValueMemberS{Value: "order-12345"},
					"CustomerID": &types.AttributeValueMemberS{Value: "customer-67890"},
					"Status": &types.AttributeValueMemberS{Value: "Pending"},
					"Amount": &types.AttributeValueMemberN{Value: "299.99"},
				},
				SequenceNumber: "0000001",
				SizeBytes: 256,
			},
		},
		// Inventory update
		{
			EventID:   "inventory-updated-1",
			EventName: "MODIFY",
			EventSource: "aws:dynamodb",
			Timestamp: baseTime.Add(1 * time.Second),
			DynamoDB: StreamDynamoDBData{
				Keys: map[string]types.AttributeValue{
					"ProductID": &types.AttributeValueMemberS{Value: "product-abc123"},
				},
				OldImage: map[string]types.AttributeValue{
					"StockCount": &types.AttributeValueMemberN{Value: "10"},
				},
				NewImage: map[string]types.AttributeValue{
					"StockCount": &types.AttributeValueMemberN{Value: "8"},
				},
				SequenceNumber: "0000002",
				SizeBytes: 128,
			},
		},
		// Payment processing
		{
			EventID:   "payment-processed-1",
			EventName: "INSERT",
			EventSource: "aws:dynamodb",
			Timestamp: baseTime.Add(2 * time.Second),
			DynamoDB: StreamDynamoDBData{
				Keys: map[string]types.AttributeValue{
					"PaymentID": &types.AttributeValueMemberS{Value: "payment-xyz789"},
				},
				NewImage: map[string]types.AttributeValue{
					"PaymentID": &types.AttributeValueMemberS{Value: "payment-xyz789"},
					"OrderID": &types.AttributeValueMemberS{Value: "order-12345"},
					"Amount": &types.AttributeValueMemberN{Value: "299.99"},
					"Status": &types.AttributeValueMemberS{Value: "Completed"},
				},
				SequenceNumber: "0000003",
				SizeBytes: 192,
			},
		},
		// Order status update
		{
			EventID:   "order-confirmed-1",
			EventName: "MODIFY",
			EventSource: "aws:dynamodb",
			Timestamp: baseTime.Add(3 * time.Second),
			DynamoDB: StreamDynamoDBData{
				Keys: map[string]types.AttributeValue{
					"OrderID": &types.AttributeValueMemberS{Value: "order-12345"},
				},
				OldImage: map[string]types.AttributeValue{
					"Status": &types.AttributeValueMemberS{Value: "Pending"},
				},
				NewImage: map[string]types.AttributeValue{
					"Status": &types.AttributeValueMemberS{Value: "Confirmed"},
					"ConfirmedAt": &types.AttributeValueMemberS{Value: baseTime.Add(3*time.Second).Format(time.RFC3339)},
				},
				SequenceNumber: "0000004",
				SizeBytes: 164,
			},
		},
	}
}

func createLargeAttributeMap(numAttributes int) map[string]types.AttributeValue {
	attrs := make(map[string]types.AttributeValue)
	
	for i := 0; i < numAttributes; i++ {
		key := fmt.Sprintf("attr_%d", i)
		value := fmt.Sprintf("value_%d_with_lots_of_data_to_make_it_large", i)
		attrs[key] = &types.AttributeValueMemberS{Value: value}
	}
	
	return attrs
}

func createEventBatches(events []StreamEvent, batchSize int) [][]StreamEvent {
	var batches [][]StreamEvent
	
	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}
		batches = append(batches, events[i:end])
	}
	
	return batches
}

func createMixedStreamEvents(t *testing.T) []StreamEvent {
	events := make([]StreamEvent, 20)
	
	eventTypes := []string{"INSERT", "MODIFY", "REMOVE"}
	statuses := []string{"Active", "Inactive", "Pending", "Expired"}
	
	for i := 0; i < 20; i++ {
		eventType := eventTypes[i%len(eventTypes)]
		status := statuses[i%len(statuses)]
		
		events[i] = StreamEvent{
			EventID:   fmt.Sprintf("mixed-event-%d", i),
			EventName: eventType,
			EventSource: "aws:dynamodb",
			DynamoDB: StreamDynamoDBData{
				Keys: map[string]types.AttributeValue{
					"ID": &types.AttributeValueMemberS{Value: fmt.Sprintf("item-%d", i)},
				},
				NewImage: map[string]types.AttributeValue{
					"Status": &types.AttributeValueMemberS{Value: status},
					"Amount": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", float64(i*100+50))},
				},
				SizeBytes: int64(128 + i*50), // Varying sizes
			},
		}
	}
	
	return events
}

func applyStreamFilters(events []StreamEvent, rules StreamFilterRules) []StreamEvent {
	var filtered []StreamEvent
	
	for _, event := range events {
		// Check event type filter
		eventTypeMatch := false
		for _, allowedType := range rules.EventTypes {
			if event.EventName == allowedType {
				eventTypeMatch = true
				break
			}
		}
		if !eventTypeMatch {
			continue
		}
		
		// Check size limits
		if maxSize, exists := rules.SizeLimits["max"]; exists {
			if event.DynamoDB.SizeBytes > maxSize {
				continue
			}
		}
		
		// Check attribute filters
		if event.DynamoDB.NewImage != nil {
			if statusFilter, exists := rules.AttributeFilters["Status"]; exists {
				allowedStatuses := statusFilter.([]string)
				if statusAttr, hasStatus := event.DynamoDB.NewImage["Status"]; hasStatus {
					status := statusAttr.(*types.AttributeValueMemberS).Value
					statusMatch := false
					for _, allowedStatus := range allowedStatuses {
						if status == allowedStatus {
							statusMatch = true
							break
						}
					}
					if !statusMatch {
						continue
					}
				}
			}
		}
		
		filtered = append(filtered, event)
	}
	
	return filtered
}

func transformStreamEvents(events []StreamEvent) []TransformedStreamEvent {
	transformed := make([]TransformedStreamEvent, len(events))
	
	for i, event := range events {
		transformed[i] = TransformedStreamEvent{
			StreamEvent: event,
			ProcessingMetadata: &ProcessingMetadata{
				ProcessedAt:        time.Now(),
				TransformationType: "standard",
				EnrichmentData:     make(map[string]interface{}),
				ValidationResults: []ValidationResult{
					{Field: "EventID", Valid: true, Message: "Valid"},
					{Field: "EventName", Valid: true, Message: "Valid"},
				},
			},
		}
	}
	
	return transformed
}

func createFailingEvent(eventID, reason string) StreamEvent {
	return StreamEvent{
		EventID:   eventID,
		EventName: "INSERT",
		EventSource: "aws:dynamodb",
		DynamoDB: StreamDynamoDBData{
			Keys: map[string]types.AttributeValue{
				"ID": &types.AttributeValueMemberS{Value: eventID},
			},
			NewImage: map[string]types.AttributeValue{
				"FailureReason": &types.AttributeValueMemberS{Value: reason},
			},
		},
	}
}

func (dlq *DeadLetterQueueProcessor) processWithDLQ(event StreamEvent) *ProcessingResult {
	result := &ProcessingResult{}
	
	// Simulate processing that always fails for test events
	if event.DynamoDB.NewImage != nil {
		if reason, exists := event.DynamoDB.NewImage["FailureReason"]; exists {
			reasonStr := reason.(*types.AttributeValueMemberS).Value
			dlq.FailureReasons[reasonStr]++
			result.Error = errors.New(reasonStr)
			result.SentToDLQ = true
			dlq.TotalDLQSent++
		}
	}
	
	return result
}

func createMetricsTestEvents(t *testing.T, count int) []StreamEvent {
	events := make([]StreamEvent, count)
	eventTypes := []string{"INSERT", "MODIFY", "REMOVE"}
	tables := []string{"orders", "users", "products", "payments"}
	
	for i := 0; i < count; i++ {
		eventType := eventTypes[i%len(eventTypes)]
		tableName := tables[i%len(tables)]
		
		events[i] = StreamEvent{
			EventID:     fmt.Sprintf("metrics-event-%d", i),
			EventName:   eventType,
			EventSource: "aws:dynamodb",
			EventSourceARN: fmt.Sprintf("arn:aws:dynamodb:us-east-1:123456789012:table/%s/stream", tableName),
			Timestamp:   time.Now().Add(time.Duration(i) * time.Millisecond),
			DynamoDB: StreamDynamoDBData{
				Keys: map[string]types.AttributeValue{
					"ID": &types.AttributeValueMemberS{Value: fmt.Sprintf("item-%d", i)},
				},
				SequenceNumber: fmt.Sprintf("%010d", i),
				SizeBytes:     int64(64 + i%512), // Varying sizes
			},
		}
	}
	
	return events
}

func processEventWithMetrics(event StreamEvent, collector *StreamMetricsCollector) error {
	// Simulate processing with occasional failures
	if event.EventID[len(event.EventID)-1] == '7' { // Fail events ending in 7
		return errors.New("simulated processing failure")
	}
	
	// Simulate processing delay
	time.Sleep(time.Duration(1+event.DynamoDB.SizeBytes/1000) * time.Microsecond)
	
	return nil
}

func (collector *StreamMetricsCollector) recordEventProcessing(event StreamEvent, processingTime time.Duration, err error) {
	// Record event type metrics
	if collector.EventMetrics[event.EventName] == nil {
		collector.EventMetrics[event.EventName] = &EventTypeMetrics{}
	}
	
	eventMetrics := collector.EventMetrics[event.EventName]
	eventMetrics.Count++
	eventMetrics.TotalProcessingTime += processingTime
	
	if err == nil {
		eventMetrics.SuccessCount++
	} else {
		eventMetrics.FailureCount++
	}
	
	eventMetrics.AverageProcessingTime = eventMetrics.TotalProcessingTime / time.Duration(eventMetrics.Count)
	eventMetrics.SuccessRate = float64(eventMetrics.SuccessCount) / float64(eventMetrics.Count)
}

func (collector *StreamMetricsCollector) generateSummary() *MetricsSummary {
	totalEvents := int64(0)
	totalProcessingTime := time.Duration(0)
	
	for _, metrics := range collector.EventMetrics {
		totalEvents += metrics.Count
		totalProcessingTime += metrics.TotalProcessingTime
	}
	
	averageProcessingTime := time.Duration(0)
	if totalEvents > 0 {
		averageProcessingTime = totalProcessingTime / time.Duration(totalEvents)
	}
	
	return &MetricsSummary{
		TotalEvents:           totalEvents,
		ProcessingDuration:    time.Since(collector.StartTime),
		AverageProcessingTime: averageProcessingTime,
		TotalThroughput:      float64(totalEvents) / time.Since(collector.StartTime).Seconds(),
		EventTypeBreakdown:   collector.EventMetrics,
		TableBreakdown:       collector.TableMetrics,
	}
}