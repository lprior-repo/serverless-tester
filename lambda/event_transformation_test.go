package lambda

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransformS3Event_WithValidEvent_ShouldTransformCorrectly tests S3 event transformation
func TestTransformS3Event_WithValidEvent_ShouldTransformCorrectly(t *testing.T) {
	// Create a sample S3 event
	s3Event := BuildS3Event("test-bucket", "test-key.json", "s3:ObjectCreated:Put")
	
	// Transform the event
	transformedEvent, err := TransformS3EventE(s3Event, S3EventTransformConfig{
		AddMetadata:     true,
		NormalizeKeys:   true,
		FilterSizeLimit: 1024 * 1024, // 1MB
		ValidateSchema:  true,
	})
	
	require.NoError(t, err)
	assert.NotEmpty(t, transformedEvent)
	
	// Verify transformation
	var event S3Event
	err = json.Unmarshal([]byte(transformedEvent), &event)
	require.NoError(t, err)
	assert.Len(t, event.Records, 1)
	
	record := event.Records[0]
	assert.Equal(t, "test-bucket", record.S3.Bucket.Name)
	assert.Equal(t, "test-key.json", record.S3.Object.Key)
	assert.Equal(t, "s3:ObjectCreated:Put", record.EventName)
}

// TestValidateS3EventStructure_WithValidEvent_ShouldPass tests S3 event validation
func TestValidateS3EventStructure_WithValidEvent_ShouldPass(t *testing.T) {
	s3Event := BuildS3Event("valid-bucket", "valid-key.json", "s3:ObjectCreated:Put")
	
	errors := ValidateEventStructure(s3Event, "s3")
	
	assert.Empty(t, errors, "Valid S3 event should pass validation")
}

// TestValidateS3EventStructure_WithInvalidEvent_ShouldFail tests S3 event validation with invalid data
func TestValidateS3EventStructure_WithInvalidEvent_ShouldFail(t *testing.T) {
	invalidS3Event := `{"Records": []}`
	
	errors := ValidateEventStructure(invalidS3Event, "s3")
	
	assert.NotEmpty(t, errors, "Invalid S3 event should fail validation")
	assert.Contains(t, errors[0], "S3 event must have at least one record")
}

// TestNormalizeEventKeys_WithS3Event_ShouldNormalizeKeys tests key normalization
func TestNormalizeEventKeys_WithS3Event_ShouldNormalizeKeys(t *testing.T) {
	s3Event := BuildS3Event("Test-Bucket", "Test Key With Spaces.json", "s3:ObjectCreated:Put")
	
	normalizedEvent, err := NormalizeEventKeysE(s3Event, "s3")
	
	require.NoError(t, err)
	assert.NotEmpty(t, normalizedEvent)
	
	// Verify normalization
	var event S3Event
	err = json.Unmarshal([]byte(normalizedEvent), &event)
	require.NoError(t, err)
	
	// Keys should be URL-decoded and normalized
	record := event.Records[0]
	assert.Equal(t, "test-bucket", record.S3.Bucket.Name)
	// Object key might be URL-encoded in real S3 events
}

// TestEnrichEventWithMetadata_WithS3Event_ShouldAddMetadata tests event enrichment
func TestEnrichEventWithMetadata_WithS3Event_ShouldAddMetadata(t *testing.T) {
	s3Event := BuildS3Event("test-bucket", "test-file.json", "s3:ObjectCreated:Put")
	
	enrichedEvent, err := EnrichEventWithMetadataE(s3Event, "s3", EnrichmentConfig{
		AddProcessingTimestamp: true,
		AddEventSize:          true,
		AddEventType:          true,
		AddSourceInformation:  true,
	})
	
	require.NoError(t, err)
	assert.NotEmpty(t, enrichedEvent)
	
	// Verify metadata was added
	var enriched map[string]interface{}
	err = json.Unmarshal([]byte(enrichedEvent), &enriched)
	require.NoError(t, err)
	
	// Check for added metadata
	assert.Contains(t, enriched, "_metadata")
	metadata := enriched["_metadata"].(map[string]interface{})
	assert.Contains(t, metadata, "processing_timestamp")
	assert.Contains(t, metadata, "event_size_bytes")
	assert.Contains(t, metadata, "event_type")
	assert.Contains(t, metadata, "source_information")
}

// TestFilterEventsBySize_WithMixedSizes_ShouldFilterCorrectly tests size-based filtering
func TestFilterEventsBySize_WithMixedSizes_ShouldFilterCorrectly(t *testing.T) {
	// Create events of different sizes
	smallEvent := BuildS3Event("bucket1", "small.txt", "s3:ObjectCreated:Put")
	largeEvent := BuildSQSEvent("test-queue", fmt.Sprintf("large message %s", 
		make([]byte, 2000))) // Large message
	
	events := []string{smallEvent, largeEvent}
	
	filteredEvents, err := FilterEventsBySizeE(events, EventSizeFilter{
		MaxSizeBytes: 1500, // Only allow smaller events
	})
	
	require.NoError(t, err)
	assert.Len(t, filteredEvents, 1, "Should filter out large events")
	
	// Verify the remaining event is the small one
	var event S3Event
	err = json.Unmarshal([]byte(filteredEvents[0]), &event)
	require.NoError(t, err)
	assert.Equal(t, "bucket1", event.Records[0].S3.Bucket.Name)
}

// TestTransformDynamoDBEvent_WithValidEvent_ShouldTransformCorrectly tests DynamoDB event transformation
func TestTransformDynamoDBEvent_WithValidEvent_ShouldTransformCorrectly(t *testing.T) {
	dynamoEvent := BuildDynamoDBEvent("users", "INSERT", map[string]interface{}{
		"id": map[string]interface{}{"S": "user123"},
		"name": map[string]interface{}{"S": "John Doe"},
	})
	
	transformedEvent, err := TransformDynamoDBEventE(dynamoEvent, DynamoDBEventTransformConfig{
		SimplifyAttributeValues: true,
		AddTimestampMetadata:    true,
		ValidateSchema:          true,
	})
	
	require.NoError(t, err)
	assert.NotEmpty(t, transformedEvent)
	
	// Verify transformation
	var event DynamoDBEvent
	err = json.Unmarshal([]byte(transformedEvent), &event)
	require.NoError(t, err)
	assert.Len(t, event.Records, 1)
	
	record := event.Records[0]
	assert.Equal(t, "INSERT", record.EventName)
	assert.Contains(t, record.Dynamodb.Keys, "id")
}

// TestValidateEventChain_WithValidChain_ShouldPass tests event chain validation
func TestValidateEventChain_WithValidChain_ShouldPass(t *testing.T) {
	// Create a chain of related events
	events := []string{
		BuildS3Event("source-bucket", "data.json", "s3:ObjectCreated:Put"),
		BuildSQSEvent("processing-queue", "process data.json"),
		BuildDynamoDBEvent("processed-data", "INSERT", map[string]interface{}{
			"id": map[string]interface{}{"S": "processed-123"},
		}),
	}
	
	validationResult, err := ValidateEventChainE(events, EventChainValidationConfig{
		CheckSequence:    true,
		CheckReferences:  true,
		CheckTimestamps:  true,
		MaxChainLength:   5,
	})
	
	require.NoError(t, err)
	assert.True(t, validationResult.IsValid)
	assert.Empty(t, validationResult.ValidationErrors)
	assert.Equal(t, 3, validationResult.ChainLength)
}

// TestExtractEventMetadata_WithS3Event_ShouldExtractCorrectly tests metadata extraction
func TestExtractEventMetadata_WithS3Event_ShouldExtractCorrectly(t *testing.T) {
	s3Event := BuildS3Event("analytics-bucket", "logs/2024/01/15/access.log", "s3:ObjectCreated:Put")
	
	metadata, err := ExtractEventMetadataE(s3Event, "s3")
	
	require.NoError(t, err)
	assert.NotEmpty(t, metadata)
	
	// Verify extracted metadata
	assert.Equal(t, "s3", metadata["event_type"])
	assert.Equal(t, "analytics-bucket", metadata["source_bucket"])
	assert.Equal(t, "logs/2024/01/15/access.log", metadata["object_key"])
	assert.Equal(t, "s3:ObjectCreated:Put", metadata["event_name"])
	assert.Contains(t, metadata, "event_size_bytes")
	assert.Contains(t, metadata, "extraction_timestamp")
}

// TestEventBatchTransformation_WithMixedEvents_ShouldTransformAll tests batch transformation
func TestEventBatchTransformation_WithMixedEvents_ShouldTransformAll(t *testing.T) {
	events := []TypedEvent{
		{Type: "s3", Payload: BuildS3Event("bucket1", "file1.json", "s3:ObjectCreated:Put")},
		{Type: "sqs", Payload: BuildSQSEvent("queue1", "message1")},
		{Type: "dynamodb", Payload: BuildDynamoDBEvent("table1", "INSERT", map[string]interface{}{
			"id": map[string]interface{}{"S": "item1"},
		})},
	}
	
	transformedEvents, err := BatchTransformEventsE(events, BatchTransformConfig{
		NormalizeKeys:    true,
		AddMetadata:      true,
		ValidateSchemas:  true,
		FilterInvalid:    true,
	})
	
	require.NoError(t, err)
	assert.Len(t, transformedEvents, 3, "All events should be transformed")
	
	// Verify each transformation
	for i, transformed := range transformedEvents {
		assert.NotEmpty(t, transformed.Payload)
		assert.Equal(t, events[i].Type, transformed.Type)
		
		// Verify metadata was added
		var eventData map[string]interface{}
		err := json.Unmarshal([]byte(transformed.Payload), &eventData)
		require.NoError(t, err)
		assert.Contains(t, eventData, "_metadata")
	}
}

// TestEventValidationRules_CustomRules_ShouldApplyCorrectly tests custom validation rules
func TestEventValidationRules_CustomRules_ShouldApplyCorrectly(t *testing.T) {
	s3Event := BuildS3Event("test-bucket", "test-file.txt", "s3:ObjectCreated:Put")
	
	// Define custom validation rules
	rules := []EventValidationRule{
		{
			Name: "bucket_name_format",
			Rule: func(event map[string]interface{}) bool {
				// Check if bucket name follows naming convention
				records, ok := event["Records"].([]interface{})
				if !ok || len(records) == 0 {
					return false
				}
				
				record := records[0].(map[string]interface{})
				s3Data := record["s3"].(map[string]interface{})
				bucket := s3Data["bucket"].(map[string]interface{})
				bucketName := bucket["name"].(string)
				
				return len(bucketName) >= 3 && len(bucketName) <= 63
			},
			ErrorMessage: "Bucket name must be between 3 and 63 characters",
		},
		{
			Name: "object_key_extension",
			Rule: func(event map[string]interface{}) bool {
				records := event["Records"].([]interface{})
				record := records[0].(map[string]interface{})
				s3Data := record["s3"].(map[string]interface{})
				object := s3Data["object"].(map[string]interface{})
				objectKey := object["key"].(string)
				
				return len(objectKey) > 4 && objectKey[len(objectKey)-4:] == ".txt"
			},
			ErrorMessage: "Object key must have .txt extension",
		},
	}
	
	validationResult, err := ValidateEventWithCustomRulesE(s3Event, rules)
	
	require.NoError(t, err)
	assert.True(t, validationResult.IsValid)
	assert.Empty(t, validationResult.FailedRules)
}

// TestEventTransformation_WithInvalidInput_ShouldReturnError tests error handling
func TestEventTransformation_WithInvalidInput_ShouldReturnError(t *testing.T) {
	invalidJSON := `{"invalid": json structure`
	
	_, err := TransformS3EventE(invalidJSON, S3EventTransformConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse S3 event")
}

// TestEventSizeCalculation_WithDifferentEvents_ShouldCalculateCorrectly tests size calculation
func TestEventSizeCalculation_WithDifferentEvents_ShouldCalculateCorrectly(t *testing.T) {
	s3Event := BuildS3Event("test-bucket", "test.json", "s3:ObjectCreated:Put")
	sqsEvent := BuildSQSEvent("test-queue", "test message")
	
	s3Size := CalculateEventSize(s3Event)
	sqsSize := CalculateEventSize(sqsEvent)
	
	assert.Greater(t, s3Size, 0)
	assert.Greater(t, sqsSize, 0)
	assert.NotEqual(t, s3Size, sqsSize, "Different events should have different sizes")
}

// TestEventTimestampExtraction_WithDifferentEventTypes_ShouldExtractCorrectly tests timestamp extraction
func TestEventTimestampExtraction_WithDifferentEventTypes_ShouldExtractCorrectly(t *testing.T) {
	tests := []struct {
		name      string
		eventJSON string
		eventType string
	}{
		{
			name:      "S3Event",
			eventJSON: BuildS3Event("bucket", "key", "s3:ObjectCreated:Put"),
			eventType: "s3",
		},
		{
			name:      "SQSEvent",
			eventJSON: BuildSQSEvent("queue", "message"),
			eventType: "sqs",
		},
		{
			name:      "DynamoDBEvent",
			eventJSON: BuildDynamoDBEvent("table", "INSERT", nil),
			eventType: "dynamodb",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			timestamp, err := ExtractEventTimestampE(tc.eventJSON, tc.eventType)
			
			require.NoError(t, err)
			assert.WithinDuration(t, time.Now(), timestamp, 5*time.Minute,
				"Event timestamp should be recent")
		})
	}
}