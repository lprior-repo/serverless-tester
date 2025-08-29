package eventbridge

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// Replay validation constants
const (
	MaxReplayDurationHours   = 24 * 30 // 30 days maximum replay duration
	MinReplayDurationSeconds = 60      // 1 minute minimum replay duration
)

// ValidateReplayConfig validates a replay configuration and panics on error (Terratest pattern)
func ValidateReplayConfig(config ReplayConfig) {
	err := ValidateReplayConfigE(config)
	if err != nil {
		panic(fmt.Sprintf("ValidateReplayConfig failed: %v", err))
	}
}

// ValidateReplayConfigE validates a replay configuration and returns error
func ValidateReplayConfigE(config ReplayConfig) error {
	// Validate required fields
	if config.ReplayName == "" {
		return fmt.Errorf("replay name cannot be empty")
	}

	if config.EventSourceArn == "" {
		return fmt.Errorf("event source ARN cannot be empty")
	}

	// Validate time range
	if config.EventEndTime.Before(config.EventStartTime) || config.EventEndTime.Equal(config.EventStartTime) {
		return fmt.Errorf("end time must be after start time")
	}

	// Validate minimum duration
	duration := config.EventEndTime.Sub(config.EventStartTime)
	if duration.Seconds() < MinReplayDurationSeconds {
		return fmt.Errorf("replay duration must be at least %d seconds, got %.0f seconds",
			MinReplayDurationSeconds, duration.Seconds())
	}

	// Validate maximum duration
	maxDuration := time.Duration(MaxReplayDurationHours) * time.Hour
	if duration > maxDuration {
		return fmt.Errorf("replay duration cannot exceed %d hours, got %.0f hours",
			MaxReplayDurationHours, duration.Hours())
	}

	// Validate event source ARN format (should be archive ARN)
	if !arnRegex.MatchString(config.EventSourceArn) {
		return fmt.Errorf("invalid event source ARN format: %s", config.EventSourceArn)
	}

	// Validate destination
	if err := ValidateReplayDestinationE(config.Destination); err != nil {
		return fmt.Errorf("invalid replay destination: %v", err)
	}

	return nil
}

// ValidateReplayDestination validates a replay destination and panics on error (Terratest pattern)
func ValidateReplayDestination(destination types.ReplayDestination) {
	err := ValidateReplayDestinationE(destination)
	if err != nil {
		panic(fmt.Sprintf("ValidateReplayDestination failed: %v", err))
	}
}

// ValidateReplayDestinationE validates a replay destination and returns error
func ValidateReplayDestinationE(destination types.ReplayDestination) error {
	if destination.Arn == nil {
		return fmt.Errorf("destination ARN is required")
	}

	arn := *destination.Arn

	// Validate ARN format
	if !arnRegex.MatchString(arn) {
		return fmt.Errorf("invalid destination ARN format: %s", arn)
	}

	// Validate that destination is an event bus or rule
	serviceType := GetTargetServiceType(arn)
	if serviceType != "eventbridge" {
		return fmt.Errorf("replay destination must be an EventBridge event bus or rule, got service type: %s", serviceType)
	}

	// Note: FilterArn field was removed from types.ReplayDestination
	// Skip FilterArn validation for now

	return nil
}

// ValidateArchiveConfig validates an archive configuration and panics on error (Terratest pattern)
func ValidateArchiveConfig(config ArchiveConfig) {
	err := ValidateArchiveConfigE(config)
	if err != nil {
		panic(fmt.Sprintf("ValidateArchiveConfig failed: %v", err))
	}
}

// ValidateArchiveConfigE validates an archive configuration and returns error
func ValidateArchiveConfigE(config ArchiveConfig) error {
	// Validate required fields
	if config.ArchiveName == "" {
		return fmt.Errorf("archive name cannot be empty")
	}

	if config.EventSourceArn == "" {
		return fmt.Errorf("event source ARN cannot be empty")
	}

	// Validate ARN format
	if !arnRegex.MatchString(config.EventSourceArn) {
		return fmt.Errorf("invalid event source ARN format: %s", config.EventSourceArn)
	}

	// Validate retention days if specified
	if config.RetentionDays < 0 {
		return fmt.Errorf("retention days cannot be negative: %d", config.RetentionDays)
	}

	// EventBridge allows up to indefinite retention, but practical limit is around 10 years
	maxRetentionDays := int32(3650) // ~10 years
	if config.RetentionDays > maxRetentionDays {
		return fmt.Errorf("retention days exceeds practical maximum of %d days: %d", maxRetentionDays, config.RetentionDays)
	}

	// Validate event pattern if provided
	if config.EventPattern != "" {
		if !ValidateEventPattern(config.EventPattern) {
			return fmt.Errorf("invalid event pattern: %s", config.EventPattern)
		}
	}

	return nil
}

// BuildReplayConfig creates a ReplayConfig with the given parameters
func BuildReplayConfig(replayName string, archiveArn string, startTime time.Time, endTime time.Time, destinationArn string) ReplayConfig {
	return ReplayConfig{
		ReplayName:     replayName,
		EventSourceArn: archiveArn,
		EventStartTime: startTime,
		EventEndTime:   endTime,
		Destination: types.ReplayDestination{
			Arn: &destinationArn,
		},
	}
}

// BuildArchiveConfig creates an ArchiveConfig with the given parameters
func BuildArchiveConfig(archiveName string, eventSourceArn string, retentionDays int32) ArchiveConfig {
	return ArchiveConfig{
		ArchiveName:    archiveName,
		EventSourceArn: eventSourceArn,
		RetentionDays:  retentionDays,
	}
}

// BuildArchiveConfigWithPattern creates an ArchiveConfig with event pattern filtering
func BuildArchiveConfigWithPattern(archiveName string, eventSourceArn string, eventPattern string, retentionDays int32) ArchiveConfig {
	return ArchiveConfig{
		ArchiveName:    archiveName,
		EventSourceArn: eventSourceArn,
		EventPattern:   eventPattern,
		RetentionDays:  retentionDays,
	}
}

// isEventBridgeRule checks if an ARN represents an EventBridge rule
func isEventBridgeRule(arn string) bool {
	// EventBridge rule ARN format: arn:aws:events:region:account:rule/rule-name
	return arnRegex.MatchString(arn) &&
		GetTargetServiceType(arn) == "eventbridge" &&
		(containsPattern(arn, ":rule/") || containsPattern(arn, ":event-bus/"))
}

// containsPattern checks if a string contains a specific pattern
func containsPattern(s string, pattern string) bool {
	return len(s) >= len(pattern) &&
		findSubstring(s, pattern) >= 1
}

// findSubstring finds the index of a substring in a string (-1 if not found)
func findSubstring(s string, substr string) int {
	if len(substr) == 0 {
		return 0
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}

// CalculateReplayDuration calculates the duration of a replay
func CalculateReplayDuration(config ReplayConfig) time.Duration {
	return config.EventEndTime.Sub(config.EventStartTime)
}

// EstimateReplayTime estimates how long a replay will take based on event volume
func EstimateReplayTime(eventCount int64, eventsPerSecond int32) time.Duration {
	if eventsPerSecond <= 0 {
		eventsPerSecond = 100 // Default processing rate
	}

	processingSeconds := eventCount / int64(eventsPerSecond)
	return time.Duration(processingSeconds) * time.Second
}

// ValidateReplayTimeRange validates that a replay time range is within acceptable bounds
func ValidateReplayTimeRange(startTime time.Time, endTime time.Time) error {
	now := time.Now()

	// Check that times are not in the future
	if startTime.After(now) {
		return fmt.Errorf("replay start time cannot be in the future")
	}

	if endTime.After(now) {
		return fmt.Errorf("replay end time cannot be in the future")
	}

	// Validate the time range manually without creating a full ReplayConfig
	// This avoids other validation errors from required fields
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return fmt.Errorf("end time must be after start time")
	}

	// Validate minimum duration
	duration := endTime.Sub(startTime)
	if duration.Seconds() < MinReplayDurationSeconds {
		return fmt.Errorf("replay duration must be at least %d seconds, got %.0f seconds",
			MinReplayDurationSeconds, duration.Seconds())
	}

	// Validate maximum duration
	maxDuration := time.Duration(MaxReplayDurationHours) * time.Hour
	if duration > maxDuration {
		return fmt.Errorf("replay duration cannot exceed %d hours, got %.0f hours",
			MaxReplayDurationHours, duration.Hours())
	}

	return nil
}
