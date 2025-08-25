// Package eventbridge provides comprehensive AWS EventBridge testing utilities following Terratest patterns.
//
// This package is designed to be a drop-in replacement for Terratest's EventBridge utilities
// with enhanced serverless-specific features. It follows strict functional programming
// principles and provides both Function and FunctionE variants for all operations.
//
// Key Features:
//   - Event publishing with retry logic and batch operations
//   - Rule lifecycle management with pattern validation
//   - Custom event bus creation and management
//   - Archive and replay functionality for event sourcing
//   - Dead letter queue integration for error handling
//   - Schema registry integration for event validation
//   - Multi-region event replication support
//   - Comprehensive event builders for AWS services
//   - Pattern matching and filtering utilities
//   - Cross-account permission management
//
// Example usage:
//
//   func TestEventBridgeWorkflow(t *testing.T) {
//       ctx := sfx.NewTestContext(t)
//       
//       // Create custom event bus
//       bus := CreateEventBus(ctx, "test-bus")
//       
//       // Create rule with pattern
//       rule := CreateRule(ctx, RuleConfig{
//           Name: "test-rule",
//           EventPattern: `{"source": ["test.service"]}`,
//           EventBusName: bus.Name,
//       })
//       
//       // Put event and verify delivery
//       event := BuildCustomEvent("test.service", "Test Event", map[string]interface{}{
//           "key": "value",
//       })
//       result := PutEvent(ctx, event)
//       
//       assert.True(t, result.Success)
//   }
package eventbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// Common EventBridge configurations
const (
	DefaultTimeout        = 30 * time.Second
	DefaultRetryAttempts  = 3
	DefaultRetryDelay     = 1 * time.Second
	MaxEventBatchSize     = 10
	MaxEventDetailSize    = 256 * 1024 // 256KB
	DefaultEventBusName   = "default"
)

// Rule states
const (
	RuleStateEnabled  = types.RuleStateEnabled
	RuleStateDisabled = types.RuleStateDisabled
)

// Archive states
const (
	ArchiveStateEnabled  = types.ArchiveStateEnabled
	ArchiveStateDisabled = types.ArchiveStateDisabled
)

// Replay states
const (
	ReplayStateStarting  = types.ReplayStateStarting
	ReplayStateRunning   = types.ReplayStateRunning
	ReplayStateCompleted = types.ReplayStateCompleted
	ReplayStateFailed    = types.ReplayStateFailed
)

// Common errors
var (
	ErrEventNotFound        = errors.New("event not found")
	ErrRuleNotFound         = errors.New("rule not found")
	ErrEventBusNotFound     = errors.New("event bus not found")
	ErrInvalidEventPattern  = errors.New("invalid event pattern")
	ErrInvalidEventDetail   = errors.New("invalid event detail")
	ErrArchiveNotFound      = errors.New("archive not found")
	ErrReplayNotFound       = errors.New("replay not found")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrQuotaExceeded        = errors.New("quota exceeded")
	ErrEventSizeTooLarge    = errors.New("event size too large")
)

// EventBridgeAPI defines the interface for EventBridge operations to enable mocking
type EventBridgeAPI interface {
	PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
	PutRule(ctx context.Context, params *eventbridge.PutRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutRuleOutput, error)
	DeleteRule(ctx context.Context, params *eventbridge.DeleteRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteRuleOutput, error)
	DescribeRule(ctx context.Context, params *eventbridge.DescribeRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeRuleOutput, error)
	EnableRule(ctx context.Context, params *eventbridge.EnableRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.EnableRuleOutput, error)
	DisableRule(ctx context.Context, params *eventbridge.DisableRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DisableRuleOutput, error)
	ListRules(ctx context.Context, params *eventbridge.ListRulesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListRulesOutput, error)
	PutTargets(ctx context.Context, params *eventbridge.PutTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutTargetsOutput, error)
	RemoveTargets(ctx context.Context, params *eventbridge.RemoveTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.RemoveTargetsOutput, error)
	ListTargetsByRule(ctx context.Context, params *eventbridge.ListTargetsByRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListTargetsByRuleOutput, error)
	CreateEventBus(ctx context.Context, params *eventbridge.CreateEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.CreateEventBusOutput, error)
	DeleteEventBus(ctx context.Context, params *eventbridge.DeleteEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteEventBusOutput, error)
	DescribeEventBus(ctx context.Context, params *eventbridge.DescribeEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeEventBusOutput, error)
	ListEventBuses(ctx context.Context, params *eventbridge.ListEventBusesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListEventBusesOutput, error)
	CreateArchive(ctx context.Context, params *eventbridge.CreateArchiveInput, optFns ...func(*eventbridge.Options)) (*eventbridge.CreateArchiveOutput, error)
	DeleteArchive(ctx context.Context, params *eventbridge.DeleteArchiveInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteArchiveOutput, error)
	DescribeArchive(ctx context.Context, params *eventbridge.DescribeArchiveInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeArchiveOutput, error)
	ListArchives(ctx context.Context, params *eventbridge.ListArchivesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListArchivesOutput, error)
	StartReplay(ctx context.Context, params *eventbridge.StartReplayInput, optFns ...func(*eventbridge.Options)) (*eventbridge.StartReplayOutput, error)
	DescribeReplay(ctx context.Context, params *eventbridge.DescribeReplayInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeReplayOutput, error)
	CancelReplay(ctx context.Context, params *eventbridge.CancelReplayInput, optFns ...func(*eventbridge.Options)) (*eventbridge.CancelReplayOutput, error)
	PutPermission(ctx context.Context, params *eventbridge.PutPermissionInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutPermissionOutput, error)
	RemovePermission(ctx context.Context, params *eventbridge.RemovePermissionInput, optFns ...func(*eventbridge.Options)) (*eventbridge.RemovePermissionOutput, error)
	TestEventPattern(ctx context.Context, params *eventbridge.TestEventPatternInput, optFns ...func(*eventbridge.Options)) (*eventbridge.TestEventPatternOutput, error)
}

// TestingT provides interface compatibility with testing frameworks
// This interface is compatible with both testing.T and Terratest's testing interface
type TestingT interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{}) // Added for Terratest compatibility
	Fail()                     // Added for Terratest compatibility
	FailNow()
	Helper()
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Name() string
}

// TestContext represents the testing context with AWS configuration
type TestContext struct {
	T         TestingT
	AwsConfig aws.Config
	Region    string
	Client    EventBridgeAPI // Optional: allows injection of mock clients for testing
}

// CustomEvent represents a custom EventBridge event
type CustomEvent struct {
	Source       string
	DetailType   string
	Detail       string
	EventBusName string
	Resources    []string
	Time         *time.Time
}

// PutEventResult contains the response from putting an event
type PutEventResult struct {
	EventID   string
	Success   bool
	ErrorCode string
	ErrorMessage string
}

// PutEventsResult contains the response from putting multiple events
type PutEventsResult struct {
	FailedEntryCount int32
	Entries         []PutEventResult
}

// RuleConfig represents EventBridge rule configuration
type RuleConfig struct {
	Name               string
	Description        string
	EventPattern       string
	ScheduleExpression string
	State              types.RuleState
	EventBusName       string
	Tags               map[string]string
}

// RuleResult contains the response from rule operations
type RuleResult struct {
	Name               string
	RuleArn            string
	State              types.RuleState
	Description        string
	EventBusName       string
	EventPattern       string
	ScheduleExpression string
}

// TargetConfig represents EventBridge target configuration
type TargetConfig struct {
	ID            string
	Arn           string
	RoleArn       string
	Input         string
	InputPath     string
	InputTransformer *types.InputTransformer
	DeadLetterConfig *types.DeadLetterConfig
	RetryPolicy      *types.RetryPolicy
}

// EventBusConfig represents EventBridge custom bus configuration
type EventBusConfig struct {
	Name              string
	EventSourceName   string
	Tags              map[string]string
	DeadLetterConfig  *types.DeadLetterConfig
}

// EventBusResult contains the response from event bus operations
type EventBusResult struct {
	Name              string
	Arn               string
	EventSourceName   string
	CreationTime      *time.Time
}

// ArchiveConfig represents EventBridge archive configuration
type ArchiveConfig struct {
	ArchiveName      string
	EventSourceArn   string
	EventPattern     string
	Description      string
	RetentionDays    int32
}

// ArchiveResult contains the response from archive operations
type ArchiveResult struct {
	ArchiveName      string
	ArchiveArn       string
	State            types.ArchiveState
	EventSourceArn   string
	CreationTime     *time.Time
}

// ReplayConfig represents EventBridge replay configuration
type ReplayConfig struct {
	ReplayName         string
	EventSourceArn     string
	EventStartTime     time.Time
	EventEndTime       time.Time
	Destination        types.ReplayDestination
	Description        string
}

// ReplayResult contains the response from replay operations
type ReplayResult struct {
	ReplayName         string
	ReplayArn          string
	State              types.ReplayState
	StateReason        string
	EventStartTime     *time.Time
	EventEndTime       *time.Time
	ReplayStartTime    *time.Time
	ReplayEndTime      *time.Time
}

// RetryConfig configures retry behavior for operations
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// defaultRetryConfig provides sensible defaults for retry operations
func defaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: DefaultRetryAttempts,
		BaseDelay:   DefaultRetryDelay,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}
}

// createEventBridgeClient creates an EventBridge client from the test context
func createEventBridgeClient(ctx *TestContext) EventBridgeAPI {
	// If a mock client is provided, use it for testing
	if ctx.Client != nil {
		return ctx.Client
	}
	// Otherwise, create a real AWS client
	return eventbridge.NewFromConfig(ctx.AwsConfig)
}

// logOperation logs the start of an operation for observability
func logOperation(operation string, details map[string]interface{}) {
	logData := map[string]interface{}{
		"operation": operation,
	}
	
	// Merge details into log data
	for key, value := range details {
		logData[key] = value
	}
	
	slog.Info("EventBridge operation started", slog.Any("details", logData))
}

// logResult logs the result of an operation for observability
func logResult(operation string, success bool, duration time.Duration, err error) {
	logData := map[string]interface{}{
		"operation":   operation,
		"success":     success,
		"duration_ms": duration.Milliseconds(),
	}
	
	if err != nil {
		logData["error"] = err.Error()
		slog.Error("EventBridge operation failed", slog.Any("details", logData))
	} else {
		slog.Info("EventBridge operation completed", slog.Any("details", logData))
	}
}

// validateEventDetail validates that the event detail is valid JSON
func validateEventDetail(detail string) error {
	if detail == "" {
		return nil // Empty detail is valid
	}
	
	if len(detail) > MaxEventDetailSize {
		return fmt.Errorf("%w: detail size exceeds maximum: %d bytes", ErrEventSizeTooLarge, len(detail))
	}
	
	var jsonData interface{}
	if err := json.Unmarshal([]byte(detail), &jsonData); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEventDetail, err)
	}
	
	return nil
}

// calculateBackoffDelay calculates exponential backoff delay
func calculateBackoffDelay(attempt int, config RetryConfig) time.Duration {
	if attempt <= 0 {
		return config.BaseDelay
	}
	
	// Calculate exponential delay: BaseDelay * (Multiplier ^ attempt)
	// Handle overflow by capping at MaxDelay early
	multiplier := 1.0
	baseDelayFloat := float64(config.BaseDelay)
	maxDelayFloat := float64(config.MaxDelay)
	
	for i := 0; i < attempt; i++ {
		nextMultiplier := multiplier * config.Multiplier
		
		// Handle special cases for multipliers
		if config.Multiplier == 0.0 {
			// Zero multiplier should result in zero delay after first iteration
			multiplier = 0.0
			break
		}
		
		// Check for overflow (but allow fractional multipliers that decrease)
		if nextMultiplier > maxDelayFloat/baseDelayFloat {
			return config.MaxDelay
		}
		
		// Only treat decreasing multiplier as overflow if it's not intentional (fractional)
		if config.Multiplier > 1.0 && nextMultiplier < multiplier {
			return config.MaxDelay
		}
		
		multiplier = nextMultiplier
	}
	
	delay := time.Duration(baseDelayFloat * multiplier)
	
	// Cap at maximum delay as final safeguard
	if delay > config.MaxDelay || delay < 0 { // Also check for negative overflow
		delay = config.MaxDelay
	}
	
	return delay
}