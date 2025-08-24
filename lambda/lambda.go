// Package lambda provides comprehensive AWS Lambda testing utilities following Terratest patterns.
//
// This package is designed to be a drop-in replacement for Terratest's Lambda utilities
// with enhanced serverless-specific features. It follows strict functional programming
// principles and provides both Function and FunctionE variants for all operations.
//
// Key Features:
//   - Synchronous and asynchronous Lambda invocations
//   - Function configuration management and retrieval
//   - CloudWatch logs integration with parsing utilities
//   - Environment variable management and validation
//   - Event source mapping support
//   - Retry mechanisms with exponential backoff
//   - JSON payload marshaling/unmarshaling helpers
//   - Common event pattern builders for AWS services
//   - Performance monitoring and timeout handling
//
// Example usage:
//
//   func TestLambdaInvocation(t *testing.T) {
//       ctx := sfx.NewTestContext(t)
//       
//       // Invoke Lambda function synchronously
//       result := Invoke(ctx, "my-function", `{"key": "value"}`)
//       assert.Contains(t, result.Payload, "success")
//       
//       // Get function configuration
//       config := GetFunction(ctx, "my-function")
//       assert.Equal(t, "nodejs18.x", config.Runtime)
//   }
package lambda

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// Common Lambda configurations
const (
	DefaultTimeout       = 30 * time.Second
	DefaultMemorySize    = 128
	MaxPayloadSize       = 6 * 1024 * 1024 // 6MB
	DefaultRetryAttempts = 3
	DefaultRetryDelay    = 1 * time.Second
)

// Invocation types
const (
	InvocationTypeRequestResponse = types.InvocationTypeRequestResponse
	InvocationTypeEvent          = types.InvocationTypeEvent
	InvocationTypeDryRun         = types.InvocationTypeDryRun
)

// Log types
const (
	LogTypeNone = types.LogTypeNone
	LogTypeTail = types.LogTypeTail
)

// Common errors
var (
	ErrFunctionNotFound    = errors.New("lambda function not found")
	ErrInvalidPayload      = errors.New("invalid payload format")
	ErrInvocationTimeout   = errors.New("lambda invocation timeout")
	ErrInvocationFailed    = errors.New("lambda invocation failed")
	ErrLogRetrievalFailed  = errors.New("failed to retrieve lambda logs")
	ErrInvalidFunctionName = errors.New("invalid function name")
)

// TestingT provides interface compatibility with testing frameworks
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

// TestContext represents the testing context with AWS configuration
type TestContext struct {
	T         TestingT
	AwsConfig aws.Config
	Region    string
}

// InvokeOptions configures Lambda invocation behavior
type InvokeOptions struct {
	InvocationType   types.InvocationType
	LogType          types.LogType
	ClientContext    string
	Qualifier        string
	Timeout          time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
	PayloadValidator func(string) error
}

// InvokeResult contains the response from a Lambda invocation
type InvokeResult struct {
	StatusCode     int32
	Payload        string
	LogResult      string
	ExecutedVersion string
	FunctionError  string
	ExecutionTime  time.Duration
}

// FunctionConfiguration represents Lambda function configuration
type FunctionConfiguration struct {
	FunctionName    string
	FunctionArn     string
	Runtime         types.Runtime
	Handler         string
	Description     string
	Timeout         int32
	MemorySize      int32
	LastModified    string
	Environment     map[string]string
	Role            string
	State           types.State
	StateReason     string
	Version         string
}

// LogEntry represents a parsed CloudWatch log entry
type LogEntry struct {
	Timestamp time.Time
	Message   string
	Level     string
	RequestID string
}

// EventSourceMappingConfig represents event source mapping configuration
type EventSourceMappingConfig struct {
	EventSourceArn               string
	FunctionName                 string
	BatchSize                    int32
	MaximumBatchingWindowInSeconds int32
	StartingPosition             types.EventSourcePosition
	Enabled                      bool
}

// UpdateFunctionConfig represents function update configuration
type UpdateFunctionConfig struct {
	FunctionName    string
	Runtime         types.Runtime
	Handler         string
	Description     string
	Timeout         int32
	MemorySize      int32
	Environment     map[string]string
	DeadLetterConfig *types.DeadLetterConfig
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

// defaultInvokeOptions provides sensible defaults for Lambda invocations
func defaultInvokeOptions() InvokeOptions {
	return InvokeOptions{
		InvocationType: InvocationTypeRequestResponse,
		LogType:        LogTypeNone,
		Timeout:        DefaultTimeout,
		MaxRetries:     DefaultRetryAttempts,
		RetryDelay:     DefaultRetryDelay,
		PayloadValidator: func(payload string) error {
			if len(payload) > MaxPayloadSize {
				return fmt.Errorf("payload size exceeds maximum: %d bytes", MaxPayloadSize)
			}
			return nil
		},
	}
}

// validateFunctionName validates that the function name follows AWS naming conventions
func validateFunctionName(functionName string) error {
	if functionName == "" {
		return ErrInvalidFunctionName
	}
	
	if len(functionName) > 64 {
		return fmt.Errorf("%w: name too long (%d characters)", ErrInvalidFunctionName, len(functionName))
	}
	
	// Function names must match pattern: [a-zA-Z0-9-_]+
	for _, char := range functionName {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return fmt.Errorf("%w: invalid character '%c'", ErrInvalidFunctionName, char)
		}
	}
	
	return nil
}

// validatePayload validates that the payload is valid JSON
func validatePayload(payload string) error {
	if payload == "" {
		return nil // Empty payload is valid
	}
	
	var jsonData interface{}
	if err := json.Unmarshal([]byte(payload), &jsonData); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}
	
	return nil
}

// createLambdaClient creates a Lambda client from the test context
func createLambdaClient(ctx *TestContext) *lambda.Client {
	return lambda.NewFromConfig(ctx.AwsConfig)
}

// createCloudWatchLogsClient creates a CloudWatch Logs client from the test context
func createCloudWatchLogsClient(ctx *TestContext) *cloudwatchlogs.Client {
	return cloudwatchlogs.NewFromConfig(ctx.AwsConfig)
}

// logOperation logs the start of an operation for observability
func logOperation(operation string, functionName string, details map[string]interface{}) {
	logData := map[string]interface{}{
		"operation":     operation,
		"function_name": functionName,
	}
	
	// Merge details into log data
	for key, value := range details {
		logData[key] = value
	}
	
	slog.Info("Lambda operation started", slog.Any("details", logData))
}

// logResult logs the result of an operation for observability
func logResult(operation string, functionName string, success bool, duration time.Duration, err error) {
	logData := map[string]interface{}{
		"operation":     operation,
		"function_name": functionName,
		"success":       success,
		"duration_ms":   duration.Milliseconds(),
	}
	
	if err != nil {
		logData["error"] = err.Error()
		slog.Error("Lambda operation failed", slog.Any("details", logData))
	} else {
		slog.Info("Lambda operation completed", slog.Any("details", logData))
	}
}

// mergeInvokeOptions merges user options with defaults
func mergeInvokeOptions(userOpts *InvokeOptions) InvokeOptions {
	defaults := defaultInvokeOptions()
	
	if userOpts == nil {
		return defaults
	}
	
	// Use user values where provided, defaults otherwise
	if userOpts.InvocationType == "" {
		userOpts.InvocationType = defaults.InvocationType
	}
	if userOpts.LogType == "" {
		userOpts.LogType = defaults.LogType
	}
	if userOpts.Timeout == 0 {
		userOpts.Timeout = defaults.Timeout
	}
	if userOpts.MaxRetries == 0 {
		userOpts.MaxRetries = defaults.MaxRetries
	}
	if userOpts.RetryDelay == 0 {
		userOpts.RetryDelay = defaults.RetryDelay
	}
	if userOpts.PayloadValidator == nil {
		userOpts.PayloadValidator = defaults.PayloadValidator
	}
	
	return *userOpts
}

// calculateBackoffDelay calculates exponential backoff delay with jitter
func calculateBackoffDelay(attempt int, config RetryConfig) time.Duration {
	if attempt <= 0 {
		return config.BaseDelay
	}
	
	// Calculate exponential delay: BaseDelay * (Multiplier ^ attempt)
	// Use safer approach to avoid overflow
	multiplier := 1.0
	for i := 0; i < attempt; i++ {
		newMultiplier := multiplier * config.Multiplier
		// Check for overflow by comparing with MaxDelay threshold
		delay := time.Duration(float64(config.BaseDelay) * newMultiplier)
		if delay < 0 || delay > config.MaxDelay || newMultiplier > float64(config.MaxDelay)/float64(config.BaseDelay) {
			return config.MaxDelay
		}
		multiplier = newMultiplier
	}
	
	delay := time.Duration(float64(config.BaseDelay) * multiplier)
	
	// Cap at maximum delay
	if delay > config.MaxDelay || delay < 0 {
		delay = config.MaxDelay
	}
	
	return delay
}

// sanitizeLogResult removes sensitive information from log results
func sanitizeLogResult(logResult string) string {
	// Remove AWS request IDs and other sensitive info for cleaner logs
	lines := strings.Split(logResult, "\n")
	var sanitized []string
	
	for _, line := range lines {
		// Skip lines that look like AWS internal logs
		if strings.Contains(line, "RequestId:") || 
		   strings.Contains(line, "Duration:") ||
		   strings.Contains(line, "Billed Duration:") {
			continue
		}
		
		sanitized = append(sanitized, line)
	}
	
	return strings.Join(sanitized, "\n")
}