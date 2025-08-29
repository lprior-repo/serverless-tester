// Package stepfunctions provides comprehensive AWS Step Functions testing utilities following Terratest patterns.
//
// This package is designed to provide a complete testing framework for AWS Step Functions
// with enhanced serverless-specific features. It follows strict functional programming
// principles and provides both Function and FunctionE variants for all operations.
//
// Key Features:
//   - State machine lifecycle management (Create/Update/Delete/Describe)
//   - Execution management with wait/retry logic (Start/Stop/Describe/Wait)
//   - Input builders with fluent API for complex workflows
//   - History analysis and debugging utilities
//   - Express workflow support with synchronous execution
//   - Map state and parallel execution testing
//   - Error handling and retry testing capabilities
//   - Activity task management
//   - Comprehensive execution assertions
//   - Batch execution support
//
// Example usage:
//
//   func TestStateMachineExecution(t *testing.T) {
//       ctx := sfx.NewTestContext(t)
//       
//       // Start execution with input
//       input := NewInput().Set("message", "Hello, World!")
//       execution := StartExecution(ctx, "my-state-machine", "test-execution", input)
//       
//       // Wait for completion and assert success
//       result := WaitForExecution(ctx, execution.ExecutionArn, 5*time.Minute)
//       AssertExecutionSucceeded(t, result)
//   }
package stepfunctions

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

// Common Step Functions configurations
const (
	DefaultTimeout         = 5 * time.Minute
	DefaultPollingInterval = 5 * time.Second
	MaxStateMachineNameLen = 80
	MaxExecutionNameLen    = 80
	MaxDefinitionSize      = 1024 * 1024 // 1MB
	DefaultRetryAttempts   = 3
	DefaultRetryDelay      = 1 * time.Second
)

// State machine types
const (
	StateMachineTypeStandard = types.StateMachineTypeStandard
	StateMachineTypeExpress  = types.StateMachineTypeExpress
)

// Execution statuses
const (
	ExecutionStatusRunning   = types.ExecutionStatusRunning
	ExecutionStatusSucceeded = types.ExecutionStatusSucceeded
	ExecutionStatusFailed    = types.ExecutionStatusFailed
	ExecutionStatusTimedOut  = types.ExecutionStatusTimedOut
	ExecutionStatusAborted   = types.ExecutionStatusAborted
)

// History event types (using actual AWS SDK constants)
const (
	HistoryEventTypeExecutionStarted   = types.HistoryEventTypeExecutionStarted
	HistoryEventTypeExecutionSucceeded = types.HistoryEventTypeExecutionSucceeded
	HistoryEventTypeExecutionFailed    = types.HistoryEventTypeExecutionFailed
	HistoryEventTypeExecutionTimedOut  = types.HistoryEventTypeExecutionTimedOut
	HistoryEventTypeExecutionAborted   = types.HistoryEventTypeExecutionAborted
	HistoryEventTypeTaskStateEntered   = types.HistoryEventTypeTaskStateEntered
	HistoryEventTypeTaskStateExited    = types.HistoryEventTypeTaskStateExited
)

// Common errors
var (
	ErrStateMachineNotFound    = errors.New("state machine not found")
	ErrExecutionNotFound       = errors.New("execution not found")
	ErrInvalidStateMachineName = errors.New("invalid state machine name")
	ErrInvalidExecutionName    = errors.New("invalid execution name")
	ErrInvalidDefinition       = errors.New("invalid state machine definition")
	ErrExecutionTimeout        = errors.New("execution timeout")
	ErrExecutionFailed         = errors.New("execution failed")
	ErrInvalidInput            = errors.New("invalid input")
	ErrInvalidArn              = errors.New("invalid ARN")
	ErrActivityNotFound        = errors.New("activity not found")
	ErrInvalidActivityName     = errors.New("invalid activity name")
	ErrTaskTokenRequired       = errors.New("task token is required")
	ErrInvalidTaskToken        = errors.New("invalid task token")
	ErrTaskTokenNotFound       = errors.New("task token not found")
	ErrMapRunNotFound          = errors.New("map run not found")
	ErrInvalidMapRunArn        = errors.New("invalid map run ARN")
)

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
}

// StateMachineDefinition represents a Step Functions state machine configuration
type StateMachineDefinition struct {
	Name               string
	Definition         string
	RoleArn            string
	Type               types.StateMachineType
	LoggingConfiguration *types.LoggingConfiguration
	TracingConfiguration *types.TracingConfiguration
	Tags               map[string]string
}

// StateMachineInfo contains information about an existing state machine
type StateMachineInfo struct {
	StateMachineArn      string
	Name                 string
	Status               types.StateMachineStatus
	Definition           string
	RoleArn              string
	Type                 types.StateMachineType
	CreationDate         time.Time
	LoggingConfiguration *types.LoggingConfiguration
	TracingConfiguration *types.TracingConfiguration
}

// ExecutionResult contains the response from a Step Functions execution
type ExecutionResult struct {
	ExecutionArn     string
	StateMachineArn  string
	Name             string
	Status           types.ExecutionStatus
	StartDate        time.Time
	StopDate         time.Time
	Input            string
	Output           string
	Error            string
	Cause            string
	ExecutionTime    time.Duration
	MapRunArn        string
	RedriveCount     int
	RedriveDate      time.Time
}

// ActivityInfo contains information about a Step Functions activity
type ActivityInfo struct {
	ActivityArn  string
	Name         string
	CreationDate time.Time
	Tags         map[string]string
}

// ActivityTask represents a task retrieved from an activity
type ActivityTask struct {
	TaskToken string
	Input     string
}

// TaskResult represents the result of a completed task
type TaskResult struct {
	TaskToken string
	Output    string
	Error     string
	Cause     string
}

// MapRunInfo contains information about a map run
type MapRunInfo struct {
	MapRunArn           string
	ExecutionArn        string
	Status              types.MapRunStatus
	StartDate           time.Time
	StopDate            time.Time
	MaxConcurrency      int32
	ToleratedFailurePercentage float64
	ToleratedFailureCount      int64
	ExecutionCounts     map[string]int64
}

// SyncExecutionResult contains the response from a synchronous execution
type SyncExecutionResult struct {
	ExecutionArn   string
	StateMachineArn string
	Name           string
	Status         types.SyncExecutionStatus
	StartDate      time.Time
	StopDate       time.Time
	Input          string
	Output         string
	Error          string
	Cause          string
	ExecutionTime  time.Duration
	BillingDetails *types.BillingDetails
	TraceHeader    string
}

// HistoryEvent represents a Step Functions execution history event
type HistoryEvent struct {
	Timestamp                           time.Time
	Type                                types.HistoryEventType
	ID                                  int64
	PreviousEventID                     int64
	StateEnteredEventDetails            *StateEnteredEventDetails
	StateExitedEventDetails             *StateExitedEventDetails
	TaskStateEnteredEventDetails        *TaskStateEnteredEventDetails
	TaskFailedEventDetails              *TaskFailedEventDetails
	TaskSucceededEventDetails           *TaskSucceededEventDetails
	TaskRetryEventDetails               *TaskRetryEventDetails
	ExecutionStartedEventDetails        *ExecutionStartedEventDetails
	ExecutionSucceededEventDetails      *ExecutionSucceededEventDetails
	ExecutionFailedEventDetails         *ExecutionFailedEventDetails
	ExecutionTimedOutEventDetails       *ExecutionTimedOutEventDetails
	ExecutionAbortedEventDetails        *ExecutionAbortedEventDetails
	LambdaFunctionScheduledEventDetails *LambdaFunctionScheduledEventDetails
	LambdaFunctionStartedEventDetails   *LambdaFunctionStartedEventDetails
	LambdaFunctionSucceededEventDetails *LambdaFunctionSucceededEventDetails
	LambdaFunctionFailedEventDetails    *LambdaFunctionFailedEventDetails
	LambdaFunctionTimedOutEventDetails  *LambdaFunctionTimedOutEventDetails
}

// StateEnteredEventDetails contains details for state entered events
type StateEnteredEventDetails struct {
	Name  string
	Input string
}

// StateExitedEventDetails contains details for state exited events
type StateExitedEventDetails struct {
	Name   string
	Output string
}

// TaskStateEnteredEventDetails contains details for task state entered events
type TaskStateEnteredEventDetails struct {
	Name     string
	Input    string
	Resource string
}

// ExecutionStartedEventDetails contains details for execution started events
type ExecutionStartedEventDetails struct {
	Input   string
	RoleArn string
}

// ExecutionSucceededEventDetails contains details for execution succeeded events
type ExecutionSucceededEventDetails struct {
	Output string
}

// ExecutionFailedEventDetails contains details for execution failed events
type ExecutionFailedEventDetails struct {
	Error string
	Cause string
}

// ExecutionTimedOutEventDetails contains details for execution timed out events
type ExecutionTimedOutEventDetails struct {
	Error string
	Cause string
}

// ExecutionAbortedEventDetails contains details for execution aborted events
type ExecutionAbortedEventDetails struct {
	Error string
	Cause string
}

// LambdaFunctionScheduledEventDetails contains details for Lambda function scheduled events
type LambdaFunctionScheduledEventDetails struct {
	Resource                   string
	Input                      string
	TimeoutInSeconds           int64
	TaskCredentials            string
}

// LambdaFunctionStartedEventDetails contains details for Lambda function started events
type LambdaFunctionStartedEventDetails struct {
}

// LambdaFunctionSucceededEventDetails contains details for Lambda function succeeded events
type LambdaFunctionSucceededEventDetails struct {
	Output string
}

// LambdaFunctionFailedEventDetails contains details for Lambda function failed events
type LambdaFunctionFailedEventDetails struct {
	Error string
	Cause string
}

// LambdaFunctionTimedOutEventDetails contains details for Lambda function timed out events
type LambdaFunctionTimedOutEventDetails struct {
	Error string
	Cause string
}

// TaskFailedEventDetails contains details for task failed events
type TaskFailedEventDetails struct {
	Resource     string
	ResourceType string
	Error        string
	Cause        string
}

// TaskSucceededEventDetails contains details for task succeeded events
type TaskSucceededEventDetails struct {
	Resource     string
	ResourceType string
	Output       string
}

// TaskRetryEventDetails contains details for task retry events
type TaskRetryEventDetails struct {
	Resource     string
	ResourceType string
	Error        string
	Cause        string
}

// Input provides a fluent API for building Step Functions input
type Input struct {
	data map[string]interface{}
}

// StartExecutionRequest represents a request to start a Step Functions execution
type StartExecutionRequest struct {
	StateMachineArn string
	Name            string
	Input           string
	TraceHeader     string
}

// ExecutionOptions configures Step Functions execution behavior
type ExecutionOptions struct {
	TraceHeader string
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
}

// WaitOptions configures waiting behavior for executions
type WaitOptions struct {
	Timeout         time.Duration
	PollingInterval time.Duration
	MaxAttempts     int
}

// PollConfig configures polling behavior with exponential backoff
type PollConfig struct {
	MaxAttempts        int
	Interval           time.Duration
	Timeout            time.Duration
	ExponentialBackoff bool
	BackoffMultiplier  float64
	MaxInterval        time.Duration
}

// StateMachineOptions configures state machine creation behavior
type StateMachineOptions struct {
	LoggingConfiguration *types.LoggingConfiguration
	TracingConfiguration *types.TracingConfiguration
	Tags                 map[string]string
}

// RetryConfig configures retry behavior for operations
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// ExecutionStats provides statistics about execution history
type ExecutionStats struct {
	TotalEvents         int
	ExecutionDuration   time.Duration
	StateTransitions    int
	LambdaInvocations   int
	ErrorCount          int
	RetryAttempts       int
	EventTypeCounts     map[types.HistoryEventType]int
	StateExecutionTimes map[string]time.Duration
}

// ExecutionAnalysis provides comprehensive analysis of an execution's history
type ExecutionAnalysis struct {
	TotalSteps      int
	CompletedSteps  int
	FailedSteps     int
	RetryAttempts   int
	TotalDuration   time.Duration
	ExecutionStatus types.ExecutionStatus
	FailureReason   string
	FailureCause    string
	KeyEvents       []HistoryEvent
	StepTimings     map[string]time.Duration
	ResourceUsage   map[string]int
}

// FailedStep represents information about a failed step in execution
type FailedStep struct {
	StepName      string
	EventID       int64
	FailureTime   time.Time
	Error         string
	Cause         string
	ResourceType  string
	ResourceArn   string
	AttemptNumber int
}

// RetryAttempt represents a retry attempt for a failed step
type RetryAttempt struct {
	StepName      string
	AttemptNumber int
	RetryTime     time.Time
	Error         string
	Cause         string
	ResourceType  string
	ResourceArn   string
}

// TimelineEvent represents an event in the execution timeline
type TimelineEvent struct {
	EventID     int64
	EventType   types.HistoryEventType
	Timestamp   time.Time
	StepName    string
	Duration    time.Duration
	Description string
	Details     interface{}
}

// Event detail structures for comprehensive history analysis are defined above

// NewInput creates a new input builder
func NewInput() *Input {
	return &Input{
		data: make(map[string]interface{}),
	}
}

// Set adds a key-value pair to the input
func (i *Input) Set(key string, value interface{}) *Input {
	i.data[key] = value
	return i
}

// SetIf conditionally adds a key-value pair to the input
func (i *Input) SetIf(condition bool, key string, value interface{}) *Input {
	if condition {
		i.data[key] = value
	}
	return i
}

// Merge combines this input with another input, with the other input taking precedence
func (i *Input) Merge(other *Input) *Input {
	if other == nil {
		return i
	}
	
	for key, value := range other.data {
		i.data[key] = value
	}
	return i
}

// ToJSON converts the input to JSON string
func (i *Input) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(i.data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal input to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// Get retrieves a value from the input
func (i *Input) Get(key string) (interface{}, bool) {
	value, exists := i.data[key]
	return value, exists
}

// GetString retrieves a string value from the input
func (i *Input) GetString(key string) (string, bool) {
	if value, exists := i.data[key]; exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetInt retrieves an integer value from the input
func (i *Input) GetInt(key string) (int, bool) {
	if value, exists := i.data[key]; exists {
		if num, ok := value.(int); ok {
			return num, true
		}
	}
	return 0, false
}

// GetBool retrieves a boolean value from the input
func (i *Input) GetBool(key string) (bool, bool) {
	if value, exists := i.data[key]; exists {
		if boolean, ok := value.(bool); ok {
			return boolean, true
		}
	}
	return false, false
}

// isEmpty checks if the input has no data
func (i *Input) isEmpty() bool {
	return len(i.data) == 0
}

// Common input patterns for typical workflows

// CreateOrderInput creates input for order processing workflows
func CreateOrderInput(orderID, customerID string, items []string) *Input {
	return NewInput().
		Set("orderId", orderID).
		Set("customerId", customerID).
		Set("items", items).
		Set("timestamp", time.Now().Format(time.RFC3339))
}

// CreateWorkflowInput creates input for generic workflow execution
func CreateWorkflowInput(workflowName string, parameters map[string]interface{}) *Input {
	input := NewInput().
		Set("workflowName", workflowName).
		Set("parameters", parameters)
	
	if parameters != nil {
		// Flatten common parameters to top level for easier access
		if correlationID, exists := parameters["correlationId"]; exists {
			input.Set("correlationId", correlationID)
		}
		if requestID, exists := parameters["requestId"]; exists {
			input.Set("requestId", requestID)
		}
	}
	
	return input
}

// CreateRetryableInput creates input for workflows with retry logic
func CreateRetryableInput(operation string, maxRetries int) *Input {
	return NewInput().
		Set("operation", operation).
		Set("maxRetries", maxRetries).
		Set("currentAttempt", 0).
		Set("retryEnabled", true)
}

// Validation functions

// validateStateMachineName validates that the state machine name follows AWS naming conventions
func validateStateMachineName(stateMachineName string) error {
	if stateMachineName == "" {
		return ErrInvalidStateMachineName
	}
	
	if len(stateMachineName) > MaxStateMachineNameLen {
		return fmt.Errorf("%w: name too long (%d characters)", ErrInvalidStateMachineName, len(stateMachineName))
	}
	
	// State machine names must match pattern: [a-zA-Z0-9-_]+
	matched, err := regexp.MatchString(`^[a-zA-Z0-9-_]+$`, stateMachineName)
	if err != nil {
		return fmt.Errorf("%w: regex validation failed: %v", ErrInvalidStateMachineName, err)
	}
	
	if !matched {
		return fmt.Errorf("%w: name contains invalid characters", ErrInvalidStateMachineName)
	}
	
	return nil
}

// validateExecutionName validates that the execution name follows AWS naming conventions
func validateExecutionName(executionName string) error {
	if executionName == "" {
		return ErrInvalidExecutionName
	}
	
	if len(executionName) > MaxExecutionNameLen {
		return fmt.Errorf("%w: name too long (%d characters)", ErrInvalidExecutionName, len(executionName))
	}
	
	// Execution names must match pattern: [a-zA-Z0-9-_]+
	matched, err := regexp.MatchString(`^[a-zA-Z0-9-_]+$`, executionName)
	if err != nil {
		return fmt.Errorf("%w: regex validation failed: %v", ErrInvalidExecutionName, err)
	}
	
	if !matched {
		return fmt.Errorf("%w: name contains invalid characters", ErrInvalidExecutionName)
	}
	
	return nil
}

// validateStateMachineDefinition validates that the state machine definition is valid JSON
func validateStateMachineDefinition(definition string) error {
	if definition == "" {
		return fmt.Errorf("%w: empty definition", ErrInvalidDefinition)
	}
	
	if len(definition) > MaxDefinitionSize {
		return fmt.Errorf("%w: definition too large (%d bytes)", ErrInvalidDefinition, len(definition))
	}
	
	// Parse as JSON to validate structure
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(definition), &parsed); err != nil {
		return fmt.Errorf("%w: invalid JSON: %v", ErrInvalidDefinition, err)
	}
	
	// Check for required fields
	if _, exists := parsed["StartAt"]; !exists {
		return fmt.Errorf("%w: missing required field 'StartAt'", ErrInvalidDefinition)
	}
	
	if _, exists := parsed["States"]; !exists {
		return fmt.Errorf("%w: missing required field 'States'", ErrInvalidDefinition)
	}
	
	return nil
}

// validateExecutionResult validates execution result against expectations
func validateExecutionResult(result *ExecutionResult, expectSuccess bool) []string {
	var errors []string
	
	if result == nil {
		errors = append(errors, "execution result is nil")
		return errors
	}
	
	if expectSuccess {
		if result.Status != types.ExecutionStatusSucceeded {
			errors = append(errors, fmt.Sprintf("expected execution to succeed but got status: %s", result.Status))
		}
	} else {
		if result.Status == types.ExecutionStatusSucceeded {
			errors = append(errors, "expected execution to fail but it succeeded")
		}
	}
	
	return errors
}


// Utility functions for AWS client operations

// createStepFunctionsClient creates a Step Functions client from the test context
func createStepFunctionsClient(ctx *TestContext) *sfn.Client {
	return sfn.NewFromConfig(ctx.AwsConfig)
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
	
	slog.Info("Step Functions operation started", slog.Any("details", logData))
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
		slog.Error("Step Functions operation failed", slog.Any("details", logData))
	} else {
		slog.Info("Step Functions operation completed", slog.Any("details", logData))
	}
}

// defaultExecutionOptions provides sensible defaults for executions
func defaultExecutionOptions() ExecutionOptions {
	return ExecutionOptions{
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultRetryAttempts,
		RetryDelay: DefaultRetryDelay,
	}
}

// defaultWaitOptions provides sensible defaults for waiting
func defaultWaitOptions() WaitOptions {
	return WaitOptions{
		Timeout:         DefaultTimeout,
		PollingInterval: DefaultPollingInterval,
		MaxAttempts:     int(DefaultTimeout / DefaultPollingInterval),
	}
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

// mergeExecutionOptions merges user options with defaults
func mergeExecutionOptions(userOpts *ExecutionOptions) ExecutionOptions {
	defaults := defaultExecutionOptions()
	
	if userOpts == nil {
		return defaults
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
	
	return *userOpts
}

// mergeWaitOptions merges user wait options with defaults
func mergeWaitOptions(userOpts *WaitOptions) WaitOptions {
	defaults := defaultWaitOptions()
	
	if userOpts == nil {
		return defaults
	}
	
	if userOpts.Timeout == 0 {
		userOpts.Timeout = defaults.Timeout
	}
	if userOpts.PollingInterval == 0 {
		userOpts.PollingInterval = defaults.PollingInterval
	}
	if userOpts.MaxAttempts == 0 {
		userOpts.MaxAttempts = defaults.MaxAttempts
	}
	
	return *userOpts
}

// calculateBackoffDelay calculates exponential backoff delay with jitter
func calculateBackoffDelay(attempt int, config RetryConfig) time.Duration {
	if attempt <= 0 {
		return config.BaseDelay
	}
	
	// Calculate exponential delay: BaseDelay * (Multiplier ^ attempt)
	multiplier := 1.0
	for i := 0; i < attempt; i++ {
		multiplier *= config.Multiplier
	}
	delay := time.Duration(float64(config.BaseDelay) * multiplier)
	
	// Cap at maximum delay
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	
	return delay
}

// State machine validation functions

// validateStateMachineCreation validates parameters for creating a state machine
func validateStateMachineCreation(definition *StateMachineDefinition, options *StateMachineOptions) error {
	if definition == nil {
		return fmt.Errorf("state machine definition is required")
	}
	
	if err := validateStateMachineName(definition.Name); err != nil {
		return err
	}
	
	if err := validateStateMachineDefinition(definition.Definition); err != nil {
		return err
	}
	
	if definition.RoleArn == "" {
		return fmt.Errorf("role ARN is required")
	}
	
	if err := validateRoleArn(definition.RoleArn); err != nil {
		return err
	}
	
	return nil
}

// validateStateMachineUpdate validates parameters for updating a state machine
func validateStateMachineUpdate(stateMachineArn, definition, roleArn string, options *StateMachineOptions) error {
	if stateMachineArn == "" {
		return fmt.Errorf("state machine ARN is required")
	}
	
	if err := validateStateMachineArn(stateMachineArn); err != nil {
		return err
	}
	
	if definition != "" {
		if err := validateStateMachineDefinition(definition); err != nil {
			return err
		}
	}
	
	if roleArn == "" {
		return fmt.Errorf("role ARN is required")
	}
	
	if err := validateRoleArn(roleArn); err != nil {
		return err
	}
	
	return nil
}

// validateStateMachineArn validates that the state machine ARN is valid
func validateStateMachineArn(stateMachineArn string) error {
	if stateMachineArn == "" {
		return fmt.Errorf("state machine ARN is required")
	}
	
	// Basic ARN format validation
	if !strings.HasPrefix(stateMachineArn, "arn:aws:states:") {
		return fmt.Errorf("%w: must start with 'arn:aws:states:'", ErrInvalidArn)
	}
	
	if !strings.Contains(stateMachineArn, ":stateMachine:") {
		return fmt.Errorf("%w: must contain ':stateMachine:'", ErrInvalidArn)
	}
	
	return nil
}

// validateRoleArn validates that the IAM role ARN is valid
func validateRoleArn(roleArn string) error {
	if roleArn == "" {
		return fmt.Errorf("role ARN is required")
	}
	
	// Basic ARN format validation
	if !strings.HasPrefix(roleArn, "arn:aws:iam::") {
		return fmt.Errorf("%w: must start with 'arn:aws:iam::'", ErrInvalidArn)
	}
	
	if !strings.Contains(roleArn, ":role/") {
		return fmt.Errorf("%w: must contain ':role/'", ErrInvalidArn)
	}
	
	return nil
}

// validateMaxResults validates the max results parameter
func validateMaxResults(maxResults int32) error {
	if maxResults < 0 {
		return fmt.Errorf("max results cannot be negative")
	}
	
	if maxResults > 1000 {
		return fmt.Errorf("max results cannot exceed 1000")
	}
	
	return nil
}

// processStateMachineInfo processes state machine information for consistent output
func processStateMachineInfo(info *StateMachineInfo) *StateMachineInfo {
	if info == nil {
		return nil
	}
	
	// Return a copy to prevent mutation
	processed := *info
	
	// Ensure definition is not empty
	if processed.Definition == "" {
		processed.Definition = "{}"
	}
	
	return &processed
}

// validateStateMachineStatus checks if a status matches the expected status
func validateStateMachineStatus(actual, expected types.StateMachineStatus) bool {
	return actual == expected
}

// Execution validation functions

// validateStartExecutionRequest validates parameters for starting an execution
func validateStartExecutionRequest(request *StartExecutionRequest) error {
	if request == nil {
		return fmt.Errorf("start execution request is required")
	}
	
	if request.StateMachineArn == "" {
		return fmt.Errorf("state machine ARN is required")
	}
	
	if err := validateStateMachineArn(request.StateMachineArn); err != nil {
		return err
	}
	
	if request.Name == "" {
		return fmt.Errorf("execution name is required")
	}
	
	if err := validateExecutionName(request.Name); err != nil {
		return err
	}
	
	// Validate input JSON if provided
	if request.Input != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(request.Input), &jsonData); err != nil {
			return fmt.Errorf("%w: invalid JSON: %v", ErrInvalidInput, err)
		}
	}
	
	return nil
}

// validateStopExecutionRequest validates parameters for stopping an execution
func validateStopExecutionRequest(executionArn, error, cause string) error {
	if executionArn == "" {
		return fmt.Errorf("execution ARN is required")
	}
	
	if err := validateExecutionArn(executionArn); err != nil {
		return err
	}
	
	// Error and cause are optional parameters
	return nil
}

// validateWaitForExecutionRequest validates parameters for waiting for an execution
func validateWaitForExecutionRequest(executionArn string, options *WaitOptions) error {
	if executionArn == "" {
		return fmt.Errorf("execution ARN is required")
	}
	
	if err := validateExecutionArn(executionArn); err != nil {
		return err
	}
	
	// Options are optional - defaults will be used
	return nil
}

// validateExecutionArn validates that the execution ARN is valid
func validateExecutionArn(executionArn string) error {
	if executionArn == "" {
		return fmt.Errorf("execution ARN is required")
	}
	
	// Basic ARN format validation
	if !strings.HasPrefix(executionArn, "arn:aws:states:") {
		return fmt.Errorf("%w: invalid execution ARN format", ErrInvalidArn)
	}
	
	if !strings.Contains(executionArn, ":execution:") {
		return fmt.Errorf("%w: invalid execution ARN format", ErrInvalidArn)
	}
	
	return nil
}

// validateListExecutionsRequest validates parameters for listing executions
func validateListExecutionsRequest(stateMachineArn string, statusFilter types.ExecutionStatus, maxResults int32) error {
	if stateMachineArn == "" {
		return fmt.Errorf("state machine ARN is required")
	}
	
	if err := validateStateMachineArn(stateMachineArn); err != nil {
		return err
	}
	
	if err := validateMaxResults(maxResults); err != nil {
		return err
	}
	
	// Status filter is optional
	return nil
}


// isExecutionComplete checks if an execution has reached a terminal state
func isExecutionComplete(status types.ExecutionStatus) bool {
	switch status {
	case types.ExecutionStatusSucceeded, types.ExecutionStatusFailed, 
		 types.ExecutionStatusTimedOut, types.ExecutionStatusAborted:
		return true
	default:
		return false
	}
}

// parseExecutionArn extracts state machine and execution names from an execution ARN
func parseExecutionArn(executionArn string) (string, string, error) {
	if executionArn == "" {
		return "", "", fmt.Errorf("execution ARN is required")
	}
	
	// Expected format: arn:aws:states:region:account:execution:stateMachine:executionName
	parts := strings.Split(executionArn, ":")
	if len(parts) < 7 {
		return "", "", fmt.Errorf("%w: invalid execution ARN format", ErrInvalidArn)
	}
	
	if parts[0] != "arn" || parts[1] != "aws" || parts[2] != "states" || parts[5] != "execution" {
		return "", "", fmt.Errorf("%w: invalid execution ARN format", ErrInvalidArn)
	}
	
	stateMachineName := parts[6]
	executionName := parts[7]
	
	return stateMachineName, executionName, nil
}

// Waiting and polling functions

// validatePollConfig validates polling configuration parameters
func validatePollConfig(executionArn string, config *PollConfig) error {
	if err := validateExecutionArn(executionArn); err != nil {
		return err
	}
	
	// Config is optional - defaults will be used
	if config != nil {
		if config.MaxAttempts < 0 {
			return fmt.Errorf("max attempts cannot be negative")
		}
		if config.Interval < 0 {
			return fmt.Errorf("interval cannot be negative")
		}
		if config.Timeout < 0 {
			return fmt.Errorf("timeout cannot be negative")
		}
		if config.BackoffMultiplier < 1.0 {
			return fmt.Errorf("backoff multiplier must be >= 1.0")
		}
	}
	
	return nil
}

// calculateNextPollInterval calculates the next polling interval with exponential backoff
func calculateNextPollInterval(currentAttempt int, config *PollConfig) time.Duration {
	if config == nil || !config.ExponentialBackoff {
		if config != nil {
			return config.Interval
		}
		return DefaultPollingInterval
	}
	
	// Calculate exponential backoff
	baseInterval := config.Interval
	if baseInterval == 0 {
		baseInterval = DefaultPollingInterval
	}
	
	// Calculate exponential delay
	multiplier := 1.0
	for i := 1; i < currentAttempt; i++ {
		multiplier *= config.BackoffMultiplier
	}
	
	interval := time.Duration(float64(baseInterval) * multiplier)
	
	// Cap at max interval
	if config.MaxInterval > 0 && interval > config.MaxInterval {
		interval = config.MaxInterval
	}
	
	return interval
}

// shouldContinuePolling checks if polling should continue based on limits
func shouldContinuePolling(attempt int, elapsed time.Duration, config *PollConfig) bool {
	if config == nil {
		// Use default limits
		return attempt < 60 && elapsed < DefaultTimeout
	}
	
	if config.MaxAttempts > 0 && attempt > config.MaxAttempts {
		return false
	}
	
	if config.Timeout > 0 && elapsed > config.Timeout {
		return false
	}
	
	return true
}

// validateExecuteAndWaitPattern validates parameters for the execute-and-wait pattern
func validateExecuteAndWaitPattern(stateMachineArn, executionName string, input *Input, timeout time.Duration) error {
	if stateMachineArn == "" {
		return fmt.Errorf("state machine ARN is required")
	}
	
	if err := validateStateMachineArn(stateMachineArn); err != nil {
		return err
	}
	
	if executionName == "" {
		return fmt.Errorf("execution name is required")
	}
	
	if err := validateExecutionName(executionName); err != nil {
		return err
	}
	
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	// Input validation (if provided)
	if input != nil && !input.isEmpty() {
		if _, err := input.ToJSON(); err != nil {
			return fmt.Errorf("invalid input JSON: %w", err)
		}
	}
	
	return nil
}

// defaultPollConfig provides sensible defaults for polling operations
func defaultPollConfig() *PollConfig {
	return &PollConfig{
		MaxAttempts:        int(DefaultTimeout / DefaultPollingInterval),
		Interval:           DefaultPollingInterval,
		Timeout:            DefaultTimeout,
		ExponentialBackoff: false,
		BackoffMultiplier:  1.5,
		MaxInterval:        30 * time.Second,
	}
}

// Activity validation functions

// validateActivityName validates that the activity name follows AWS naming conventions
func validateActivityName(activityName string) error {
	if activityName == "" {
		return ErrInvalidActivityName
	}
	
	if len(activityName) > 80 {
		return fmt.Errorf("%w: name too long (%d characters)", ErrInvalidActivityName, len(activityName))
	}
	
	// Activity names must match pattern: [a-zA-Z0-9-_]+
	matched, err := regexp.MatchString(`^[a-zA-Z0-9-_]+$`, activityName)
	if err != nil {
		return fmt.Errorf("%w: regex validation failed: %v", ErrInvalidActivityName, err)
	}
	
	if !matched {
		return fmt.Errorf("%w: name contains invalid characters", ErrInvalidActivityName)
	}
	
	return nil
}

// validateActivityArn validates that the activity ARN is valid
func validateActivityArn(activityArn string) error {
	if activityArn == "" {
		return fmt.Errorf("activity ARN is required")
	}
	
	// Basic ARN format validation
	if !strings.HasPrefix(activityArn, "arn:aws:states:") {
		return fmt.Errorf("%w: must start with 'arn:aws:states:'", ErrInvalidArn)
	}
	
	if !strings.Contains(activityArn, ":activity:") {
		return fmt.Errorf("%w: must contain ':activity:'", ErrInvalidArn)
	}
	
	return nil
}

// validateTaskToken validates that the task token is valid
func validateTaskToken(taskToken string) error {
	if taskToken == "" {
		return ErrTaskTokenRequired
	}
	
	// Task tokens are typically UUID-like strings with specific format
	if len(taskToken) < 10 {
		return fmt.Errorf("%w: token too short", ErrInvalidTaskToken)
	}
	
	return nil
}

// validateMapRunArn validates that the map run ARN is valid
func validateMapRunArn(mapRunArn string) error {
	if mapRunArn == "" {
		return fmt.Errorf("map run ARN is required")
	}
	
	// Basic ARN format validation
	if !strings.HasPrefix(mapRunArn, "arn:aws:states:") {
		return fmt.Errorf("%w: must start with 'arn:aws:states:'", ErrInvalidMapRunArn)
	}
	
	if !strings.Contains(mapRunArn, ":mapRun:") {
		return fmt.Errorf("%w: must contain ':mapRun:'", ErrInvalidMapRunArn)
	}
	
	return nil
}


// mergePollConfig merges user config with defaults
func mergePollConfig(userConfig *PollConfig) *PollConfig {
	defaults := defaultPollConfig()
	
	if userConfig == nil {
		return defaults
	}
	
	merged := *userConfig
	
	// Apply defaults for zero values
	if merged.MaxAttempts == 0 {
		merged.MaxAttempts = defaults.MaxAttempts
	}
	if merged.Interval == 0 {
		merged.Interval = defaults.Interval
	}
	if merged.Timeout == 0 {
		merged.Timeout = defaults.Timeout
	}
	if merged.BackoffMultiplier == 0 {
		merged.BackoffMultiplier = defaults.BackoffMultiplier
	}
	if merged.MaxInterval == 0 {
		merged.MaxInterval = defaults.MaxInterval
	}
	
	return &merged
}

// ParallelExecution represents a single execution request for parallel processing
type ParallelExecution struct {
	StateMachineArn string
	ExecutionName   string
	Input          *Input
	TraceHeader    string
}

// ParallelExecutionConfig configures parallel execution behavior
type ParallelExecutionConfig struct {
	MaxConcurrency    int
	Timeout          time.Duration
	WaitForCompletion bool
	FailFast         bool
}

// ParallelExecutionResult contains results from parallel execution
type ParallelExecutionResult struct {
	Execution *ParallelExecution
	Result    *ExecutionResult
	Error     error
}

// Batch Operations Types

// BatchOperationConfig configures batch operation behavior
type BatchOperationConfig struct {
	MaxConcurrency int
	Timeout        time.Duration
	FailFast       bool
}

// BatchExecutionRequest represents a single execution request for batch processing
type BatchExecutionRequest struct {
	StateMachineArn string
	ExecutionName   string
	Input          *Input
	TraceHeader    string
}

// BatchTagOperation represents a single tag operation for batch processing
type BatchTagOperation struct {
	StateMachineArn string
	Tags           map[string]string
}

// BatchStateMachineResult contains results from batch state machine operations
type BatchStateMachineResult struct {
	StateMachine    *StateMachineInfo
	StateMachineArn string
	Error          error
}

// BatchExecutionResult contains results from batch execution operations
type BatchExecutionResult struct {
	Execution    *ExecutionResult
	ExecutionArn string
	Error       error
}

// BatchTagResult contains results from batch tag operations
type BatchTagResult struct {
	StateMachineArn string
	Tags           map[string]string
	Error          error
}

// Execution Pool Types

// ExecutionPoolStatus represents the status of a pool execution
type ExecutionPoolStatus string

const (
	PoolExecutionStatusQueued    ExecutionPoolStatus = "QUEUED"
	PoolExecutionStatusRunning   ExecutionPoolStatus = "RUNNING"
	PoolExecutionStatusCompleted ExecutionPoolStatus = "COMPLETED"
	PoolExecutionStatusFailed    ExecutionPoolStatus = "FAILED"
	PoolExecutionStatusTimeout   ExecutionPoolStatus = "TIMEOUT"
	PoolExecutionStatusCancelled ExecutionPoolStatus = "CANCELLED"
)

// ExecutionPoolConfig configures execution pool behavior
type ExecutionPoolConfig struct {
	MaxConcurrentExecutions int
	MaxQueueSize           int
	ExecutionTimeout       time.Duration
	PoolTimeout           time.Duration
}

// ExecutionRequest represents a request to execute a Step Function
type ExecutionRequest struct {
	StateMachineArn string
	ExecutionName   string
	Input          *Input
	TraceHeader    string
	Priority       int // Higher numbers = higher priority
}

// PoolExecutionResult contains the result of submitting an execution to the pool
type PoolExecutionResult struct {
	ExecutionID   string
	Status        ExecutionPoolStatus
	SubmittedAt   time.Time
	StartedAt     time.Time
	CompletedAt   time.Time
	ExecutionArn  string
	Error         error
	Result        *ExecutionResult
	
	// Private field for internal use
	request       ExecutionRequest
}

// ExecutionPoolStats provides statistics about the execution pool
type ExecutionPoolStats struct {
	TotalSubmitted         int
	RunningExecutions      int
	QueuedExecutions       int
	CompletedExecutions    int
	FailedExecutions       int
	CancelledExecutions    int
	MaxConcurrentExecutions int
	MaxQueueSize          int
	AverageExecutionTime  time.Duration
}

// Workflow Chain Types

// DependencyMode defines how dependencies are evaluated
type DependencyMode string

const (
	DependencyModeAll DependencyMode = "ALL" // Wait for all dependencies to complete
	DependencyModeAny DependencyMode = "ANY" // Wait for any dependency to complete
)

// WorkflowChainConfig configures workflow chain execution behavior
type WorkflowChainConfig struct {
	Timeout          time.Duration
	MaxParallelSteps int
	FailFast         bool
}

// WorkflowCondition defines conditional execution logic
type WorkflowCondition struct {
	Expression string // Template expression that evaluates to true/false
}

// WorkflowRetryConfig configures retry behavior for workflow steps
type WorkflowRetryConfig struct {
	MaxAttempts       int
	InitialInterval   time.Duration
	BackoffMultiplier float64
	MaxInterval       time.Duration
}

// WorkflowStep represents a single step in a workflow chain
type WorkflowStep struct {
	Name            string
	StateMachineArn string
	Input           *Input
	TraceHeader     string
	Dependencies    []string
	DependencyMode  DependencyMode
	Condition       *WorkflowCondition
	RetryConfig     *WorkflowRetryConfig
}

// WorkflowStepResult contains the result of executing a workflow step
type WorkflowStepResult struct {
	Step         *WorkflowStep
	Result       *ExecutionResult
	Error        error
	StartedAt    time.Time
	CompletedAt  time.Time
	AttemptCount int
	Skipped      bool
}

// WorkflowChainResult contains the result of executing a workflow chain
type WorkflowChainResult struct {
	StepResults []WorkflowStepResult
	Stats       *WorkflowChainStats
	StartedAt   time.Time
	CompletedAt time.Time
}

// WorkflowChainStats provides statistics about workflow chain execution
type WorkflowChainStats struct {
	TotalSteps     int
	ExecutedSteps  int
	FailedSteps    int
	SkippedSteps   int
	ExecutionTime  time.Duration
}

// WorkflowChain represents a collection of workflow steps with dependencies
type WorkflowChain struct {
	steps     []*WorkflowStep
	stepIndex map[string]int
}