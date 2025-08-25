package eventbridge

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// Event bus validation constants
const (
	MaxEventBusNameLength = 256
	MinEventBusNameLength = 1
)

// Event bus name validation regex
var (
	eventBusNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
)

// ValidateEventBusConfig validates an event bus configuration and panics on error (Terratest pattern)
func ValidateEventBusConfig(config EventBusConfig) {
	err := ValidateEventBusConfigE(config)
	if err != nil {
		panic(fmt.Sprintf("ValidateEventBusConfig failed: %v", err))
	}
}

// ValidateEventBusConfigE validates an event bus configuration and returns error
func ValidateEventBusConfigE(config EventBusConfig) error {
	// Validate event bus name
	if err := ValidateEventBusNameE(config.Name); err != nil {
		return fmt.Errorf("invalid event bus name: %v", err)
	}

	// Validate event source name if provided
	if config.EventSourceName != "" {
		if err := ValidateEventSourceNameE(config.EventSourceName); err != nil {
			return fmt.Errorf("invalid event source name: %v", err)
		}
	}

	// Validate dead letter config if provided
	if config.DeadLetterConfig != nil {
		if err := ValidateDeadLetterConfigE(config.DeadLetterConfig); err != nil {
			return fmt.Errorf("invalid dead letter config: %v", err)
		}
	}

	// Validate tags
	if err := ValidateEventBusTagsE(config.Tags); err != nil {
		return fmt.Errorf("invalid tags: %v", err)
	}

	return nil
}

// ValidateEventBusName validates an event bus name and panics on error (Terratest pattern)
func ValidateEventBusName(name string) {
	err := ValidateEventBusNameE(name)
	if err != nil {
		panic(fmt.Sprintf("ValidateEventBusName failed: %v", err))
	}
}

// ValidateEventBusNameE validates an event bus name and returns error
func ValidateEventBusNameE(name string) error {
	// Check if name is empty
	if name == "" {
		return fmt.Errorf("event bus name cannot be empty")
	}

	// Check length constraints
	if len(name) < MinEventBusNameLength {
		return fmt.Errorf("event bus name too short: minimum length is %d, got %d", MinEventBusNameLength, len(name))
	}

	if len(name) > MaxEventBusNameLength {
		return fmt.Errorf("event bus name too long: maximum length is %d, got %d", MaxEventBusNameLength, len(name))
	}

	// Check for invalid characters
	if !eventBusNameRegex.MatchString(name) {
		return fmt.Errorf("invalid characters in event bus name: %s (only alphanumeric, dots, hyphens, and underscores are allowed)", name)
	}

	// Check for leading/trailing special characters
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") || strings.HasPrefix(name, "_") {
		return fmt.Errorf("event bus name cannot start with special characters: %s", name)
	}

	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, "-") || strings.HasSuffix(name, "_") {
		return fmt.Errorf("event bus name cannot end with special characters: %s", name)
	}

	// Reserved names (AWS reserved names)
	reservedNames := []string{
		"default",
		"aws.events",
		"aws.s3",
		"aws.ec2",
		"aws.lambda",
		"aws.ecs",
	}

	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			// Allow "default" as it's the standard name
			if reserved == "default" {
				continue
			}
			return fmt.Errorf("event bus name cannot be a reserved name: %s", name)
		}
	}

	return nil
}

// ValidateEventSourceName validates an event source name and returns error
func ValidateEventSourceNameE(eventSourceName string) error {
	if eventSourceName == "" {
		return fmt.Errorf("event source name cannot be empty")
	}

	// Event source names should follow reverse DNS notation
	// e.g., com.company.service, myapp.orders, etc.
	if len(eventSourceName) > 256 {
		return fmt.Errorf("event source name too long: maximum length is 256, got %d", len(eventSourceName))
	}

	// Basic validation for event source name format
	eventSourceRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !eventSourceRegex.MatchString(eventSourceName) {
		return fmt.Errorf("invalid characters in event source name: %s", eventSourceName)
	}

	// Should not start or end with special characters
	if strings.HasPrefix(eventSourceName, ".") || strings.HasPrefix(eventSourceName, "-") {
		return fmt.Errorf("event source name cannot start with special characters: %s", eventSourceName)
	}

	if strings.HasSuffix(eventSourceName, ".") || strings.HasSuffix(eventSourceName, "-") {
		return fmt.Errorf("event source name cannot end with special characters: %s", eventSourceName)
	}

	return nil
}

// ValidateEventBusTags validates event bus tags and returns error
func ValidateEventBusTagsE(tags map[string]string) error {
	// EventBridge supports up to 50 tags per resource
	const maxTags = 50
	if len(tags) > maxTags {
		return fmt.Errorf("too many tags: maximum is %d, got %d", maxTags, len(tags))
	}

	// Validate each tag
	for key, value := range tags {
		if err := ValidateTagKeyE(key); err != nil {
			return fmt.Errorf("invalid tag key '%s': %v", key, err)
		}

		if err := ValidateTagValueE(value); err != nil {
			return fmt.Errorf("invalid tag value for key '%s': %v", key, err)
		}
	}

	return nil
}

// ValidateTagKey validates a tag key and returns error
func ValidateTagKeyE(key string) error {
	if key == "" {
		return fmt.Errorf("tag key cannot be empty")
	}

	if len(key) > 128 {
		return fmt.Errorf("tag key too long: maximum length is 128, got %d", len(key))
	}

	// Tag keys cannot start with "aws:" (reserved for AWS use)
	if strings.HasPrefix(strings.ToLower(key), "aws:") {
		return fmt.Errorf("tag key cannot start with 'aws:' (reserved prefix): %s", key)
	}

	return nil
}

// ValidateTagValue validates a tag value and returns error
func ValidateTagValueE(value string) error {
	// Tag values can be empty (unlike keys)
	if len(value) > 256 {
		return fmt.Errorf("tag value too long: maximum length is 256, got %d", len(value))
	}

	return nil
}

// BuildEventBusConfig creates an EventBusConfig with the given parameters
func BuildEventBusConfig(name string, eventSourceName string) EventBusConfig {
	return EventBusConfig{
		Name:            name,
		EventSourceName: eventSourceName,
	}
}

// BuildEventBusConfigWithTags creates an EventBusConfig with tags
func BuildEventBusConfigWithTags(name string, eventSourceName string, tags map[string]string) EventBusConfig {
	return EventBusConfig{
		Name:            name,
		EventSourceName: eventSourceName,
		Tags:            tags,
	}
}

// BuildEventBusConfigWithDLQ creates an EventBusConfig with dead letter queue
func BuildEventBusConfigWithDLQ(name string, eventSourceName string, dlqArn string) EventBusConfig {
	return EventBusConfig{
		Name:            name,
		EventSourceName: eventSourceName,
		DeadLetterConfig: &types.DeadLetterConfig{
			Arn: &dlqArn,
		},
	}
}

// IsDefaultEventBus checks if a given event bus name is the default event bus
func IsDefaultEventBus(name string) bool {
	return name == "" || strings.ToLower(name) == "default"
}

// GetEventBusRegion extracts the region from an event bus ARN
func GetEventBusRegion(eventBusArn string) string {
	if !arnRegex.MatchString(eventBusArn) {
		return ""
	}

	// ARN format: arn:aws:events:region:account:event-bus/name
	parts := strings.Split(eventBusArn, ":")
	if len(parts) >= 4 {
		return parts[3]
	}

	return ""
}

// GetEventBusAccount extracts the account ID from an event bus ARN
func GetEventBusAccount(eventBusArn string) string {
	if !arnRegex.MatchString(eventBusArn) {
		return ""
	}

	// ARN format: arn:aws:events:region:account:event-bus/name
	parts := strings.Split(eventBusArn, ":")
	if len(parts) >= 5 {
		return parts[4]
	}

	return ""
}

// GetEventBusNameFromArn extracts the event bus name from an ARN
func GetEventBusNameFromArn(eventBusArn string) string {
	if !arnRegex.MatchString(eventBusArn) {
		return ""
	}

	// ARN format: arn:aws:events:region:account:event-bus/name
	parts := strings.Split(eventBusArn, ":")
	if len(parts) >= 6 {
		resource := parts[5]
		// Extract name from "event-bus/name"
		if strings.HasPrefix(resource, "event-bus/") {
			return resource[10:] // Remove "event-bus/" prefix
		}
	}

	return ""
}

// ValidateEventBusArn validates an event bus ARN format
func ValidateEventBusArnE(arn string) error {
	if !arnRegex.MatchString(arn) {
		return fmt.Errorf("invalid ARN format: %s", arn)
	}

	// Check if it's specifically an event bus ARN
	if !strings.Contains(arn, ":events:") {
		return fmt.Errorf("ARN is not for EventBridge service: %s", arn)
	}

	if !strings.Contains(arn, ":event-bus/") {
		return fmt.Errorf("ARN is not for an event bus: %s", arn)
	}

	return nil
}

// NormalizeEventBusName normalizes an event bus name (handles empty string as "default")
func NormalizeEventBusName(name string) string {
	if name == "" {
		return "default"
	}
	return name
}