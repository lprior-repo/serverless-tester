package stepfunctions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// Default values for functional Step Functions operations
const (
	FunctionalStepFunctionsDefaultTimeout         = 5 * time.Minute
	FunctionalStepFunctionsDefaultPollingInterval = 5 * time.Second
	FunctionalStepFunctionsDefaultMaxRetries      = 3
	FunctionalStepFunctionsDefaultRetryDelay      = 200 * time.Millisecond
	FunctionalStepFunctionsMaxStateMachineNameLen = 80
	FunctionalStepFunctionsMaxExecutionNameLen    = 80
	FunctionalStepFunctionsMaxDefinitionSize      = 1024 * 1024 // 1MB
)

// FunctionalStepFunctionsConfig represents immutable Step Functions configuration using functional options
type FunctionalStepFunctionsConfig struct {
	timeout              time.Duration
	pollingInterval      time.Duration
	maxRetries           int
	retryDelay           time.Duration
	stateMachineType     types.StateMachineType
	loggingConfiguration mo.Option[types.LoggingConfiguration]
	tracingConfiguration mo.Option[types.TracingConfiguration]
	tags                 map[string]string
	metadata             map[string]interface{}
}

// FunctionalStepFunctionsConfigOption represents a functional option for Step Functions configuration
type FunctionalStepFunctionsConfigOption func(*FunctionalStepFunctionsConfig)

// NewFunctionalStepFunctionsConfig creates a new immutable Step Functions configuration with functional options
func NewFunctionalStepFunctionsConfig(opts ...FunctionalStepFunctionsConfigOption) FunctionalStepFunctionsConfig {
	config := FunctionalStepFunctionsConfig{
		timeout:              FunctionalStepFunctionsDefaultTimeout,
		pollingInterval:      FunctionalStepFunctionsDefaultPollingInterval,
		maxRetries:           FunctionalStepFunctionsDefaultMaxRetries,
		retryDelay:           FunctionalStepFunctionsDefaultRetryDelay,
		stateMachineType:     types.StateMachineTypeStandard,
		loggingConfiguration: mo.None[types.LoggingConfiguration](),
		tracingConfiguration: mo.None[types.TracingConfiguration](),
		tags:                 make(map[string]string),
		metadata:             make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(&config)
	}

	return config
}

// Functional options for Step Functions configuration
func WithStepFunctionsTimeout(timeout time.Duration) FunctionalStepFunctionsConfigOption {
	return func(c *FunctionalStepFunctionsConfig) {
		c.timeout = timeout
	}
}

func WithStepFunctionsPollingInterval(interval time.Duration) FunctionalStepFunctionsConfigOption {
	return func(c *FunctionalStepFunctionsConfig) {
		c.pollingInterval = interval
	}
}

func WithStepFunctionsRetryConfig(maxRetries int, retryDelay time.Duration) FunctionalStepFunctionsConfigOption {
	return func(c *FunctionalStepFunctionsConfig) {
		c.maxRetries = maxRetries
		c.retryDelay = retryDelay
	}
}

func WithStateMachineType(smType types.StateMachineType) FunctionalStepFunctionsConfigOption {
	return func(c *FunctionalStepFunctionsConfig) {
		c.stateMachineType = smType
	}
}

func WithLoggingConfiguration(config types.LoggingConfiguration) FunctionalStepFunctionsConfigOption {
	return func(c *FunctionalStepFunctionsConfig) {
		c.loggingConfiguration = mo.Some(config)
	}
}

func WithTracingConfiguration(config types.TracingConfiguration) FunctionalStepFunctionsConfigOption {
	return func(c *FunctionalStepFunctionsConfig) {
		c.tracingConfiguration = mo.Some(config)
	}
}

func WithStepFunctionsTags(tags map[string]string) FunctionalStepFunctionsConfigOption {
	return func(c *FunctionalStepFunctionsConfig) {
		c.tags = tags
	}
}

func WithStepFunctionsMetadata(key string, value interface{}) FunctionalStepFunctionsConfigOption {
	return func(c *FunctionalStepFunctionsConfig) {
		c.metadata[key] = value
	}
}

// FunctionalStepFunctionsStateMachine represents an immutable Step Functions state machine
type FunctionalStepFunctionsStateMachine struct {
	name                 string
	definition           string
	roleArn              string
	stateMachineType     types.StateMachineType
	stateMachineArn      mo.Option[string]
	status               mo.Option[types.StateMachineStatus]
	creationDate         time.Time
	loggingConfiguration mo.Option[types.LoggingConfiguration]
	tracingConfiguration mo.Option[types.TracingConfiguration]
	tags                 map[string]string
}

// NewFunctionalStepFunctionsStateMachine creates a new immutable Step Functions state machine
func NewFunctionalStepFunctionsStateMachine(name, definition, roleArn string) FunctionalStepFunctionsStateMachine {
	return FunctionalStepFunctionsStateMachine{
		name:                 name,
		definition:           definition,
		roleArn:              roleArn,
		stateMachineType:     types.StateMachineTypeStandard,
		stateMachineArn:      mo.None[string](),
		status:               mo.None[types.StateMachineStatus](),
		creationDate:         time.Now(),
		loggingConfiguration: mo.None[types.LoggingConfiguration](),
		tracingConfiguration: mo.None[types.TracingConfiguration](),
		tags:                 make(map[string]string),
	}
}

// Accessor methods for FunctionalStepFunctionsStateMachine
func (sm FunctionalStepFunctionsStateMachine) GetName() string { return sm.name }
func (sm FunctionalStepFunctionsStateMachine) GetDefinition() string { return sm.definition }
func (sm FunctionalStepFunctionsStateMachine) GetRoleArn() string { return sm.roleArn }
func (sm FunctionalStepFunctionsStateMachine) GetType() types.StateMachineType { return sm.stateMachineType }
func (sm FunctionalStepFunctionsStateMachine) GetArn() mo.Option[string] { return sm.stateMachineArn }
func (sm FunctionalStepFunctionsStateMachine) GetStatus() mo.Option[types.StateMachineStatus] { return sm.status }
func (sm FunctionalStepFunctionsStateMachine) GetCreationDate() time.Time { return sm.creationDate }
func (sm FunctionalStepFunctionsStateMachine) GetLoggingConfiguration() mo.Option[types.LoggingConfiguration] { return sm.loggingConfiguration }
func (sm FunctionalStepFunctionsStateMachine) GetTracingConfiguration() mo.Option[types.TracingConfiguration] { return sm.tracingConfiguration }
func (sm FunctionalStepFunctionsStateMachine) GetTags() map[string]string { return sm.tags }

// Functional options for state machine creation
func (sm FunctionalStepFunctionsStateMachine) WithType(stateMachineType types.StateMachineType) FunctionalStepFunctionsStateMachine {
	return FunctionalStepFunctionsStateMachine{
		name:                 sm.name,
		definition:           sm.definition,
		roleArn:              sm.roleArn,
		stateMachineType:     stateMachineType,
		stateMachineArn:      sm.stateMachineArn,
		status:               sm.status,
		creationDate:         sm.creationDate,
		loggingConfiguration: sm.loggingConfiguration,
		tracingConfiguration: sm.tracingConfiguration,
		tags:                 sm.tags,
	}
}

func (sm FunctionalStepFunctionsStateMachine) WithArn(arn string) FunctionalStepFunctionsStateMachine {
	return FunctionalStepFunctionsStateMachine{
		name:                 sm.name,
		definition:           sm.definition,
		roleArn:              sm.roleArn,
		stateMachineType:     sm.stateMachineType,
		stateMachineArn:      mo.Some(arn),
		status:               sm.status,
		creationDate:         sm.creationDate,
		loggingConfiguration: sm.loggingConfiguration,
		tracingConfiguration: sm.tracingConfiguration,
		tags:                 sm.tags,
	}
}

func (sm FunctionalStepFunctionsStateMachine) WithStatus(status types.StateMachineStatus) FunctionalStepFunctionsStateMachine {
	return FunctionalStepFunctionsStateMachine{
		name:                 sm.name,
		definition:           sm.definition,
		roleArn:              sm.roleArn,
		stateMachineType:     sm.stateMachineType,
		stateMachineArn:      sm.stateMachineArn,
		status:               mo.Some(status),
		creationDate:         sm.creationDate,
		loggingConfiguration: sm.loggingConfiguration,
		tracingConfiguration: sm.tracingConfiguration,
		tags:                 sm.tags,
	}
}

func (sm FunctionalStepFunctionsStateMachine) WithTags(tags map[string]string) FunctionalStepFunctionsStateMachine {
	return FunctionalStepFunctionsStateMachine{
		name:                 sm.name,
		definition:           sm.definition,
		roleArn:              sm.roleArn,
		stateMachineType:     sm.stateMachineType,
		stateMachineArn:      sm.stateMachineArn,
		status:               sm.status,
		creationDate:         sm.creationDate,
		loggingConfiguration: sm.loggingConfiguration,
		tracingConfiguration: sm.tracingConfiguration,
		tags:                 tags,
	}
}

// FunctionalStepFunctionsExecution represents an immutable Step Functions execution
type FunctionalStepFunctionsExecution struct {
	name             string
	stateMachineArn  string
	input            string
	executionArn     mo.Option[string]
	status           mo.Option[types.ExecutionStatus]
	startDate        time.Time
	stopDate         mo.Option[time.Time]
	output           mo.Option[string]
	error            mo.Option[string]
	cause            mo.Option[string]
	mapRunArn        mo.Option[string]
	redriveCount     int
}

// NewFunctionalStepFunctionsExecution creates a new immutable Step Functions execution
func NewFunctionalStepFunctionsExecution(name, stateMachineArn, input string) FunctionalStepFunctionsExecution {
	return FunctionalStepFunctionsExecution{
		name:            name,
		stateMachineArn: stateMachineArn,
		input:           input,
		executionArn:    mo.None[string](),
		status:          mo.None[types.ExecutionStatus](),
		startDate:       time.Now(),
		stopDate:        mo.None[time.Time](),
		output:          mo.None[string](),
		error:           mo.None[string](),
		cause:           mo.None[string](),
		mapRunArn:       mo.None[string](),
		redriveCount:    0,
	}
}

// Accessor methods for FunctionalStepFunctionsExecution
func (ex FunctionalStepFunctionsExecution) GetName() string { return ex.name }
func (ex FunctionalStepFunctionsExecution) GetStateMachineArn() string { return ex.stateMachineArn }
func (ex FunctionalStepFunctionsExecution) GetInput() string { return ex.input }
func (ex FunctionalStepFunctionsExecution) GetExecutionArn() mo.Option[string] { return ex.executionArn }
func (ex FunctionalStepFunctionsExecution) GetStatus() mo.Option[types.ExecutionStatus] { return ex.status }
func (ex FunctionalStepFunctionsExecution) GetStartDate() time.Time { return ex.startDate }
func (ex FunctionalStepFunctionsExecution) GetStopDate() mo.Option[time.Time] { return ex.stopDate }
func (ex FunctionalStepFunctionsExecution) GetOutput() mo.Option[string] { return ex.output }
func (ex FunctionalStepFunctionsExecution) GetError() mo.Option[string] { return ex.error }
func (ex FunctionalStepFunctionsExecution) GetCause() mo.Option[string] { return ex.cause }
func (ex FunctionalStepFunctionsExecution) GetMapRunArn() mo.Option[string] { return ex.mapRunArn }
func (ex FunctionalStepFunctionsExecution) GetRedriveCount() int { return ex.redriveCount }

// FunctionalStepFunctionsActivity represents an immutable Step Functions activity
type FunctionalStepFunctionsActivity struct {
	name         string
	activityArn  mo.Option[string]
	creationDate time.Time
	tags         map[string]string
}

// NewFunctionalStepFunctionsActivity creates a new immutable Step Functions activity
func NewFunctionalStepFunctionsActivity(name string) FunctionalStepFunctionsActivity {
	return FunctionalStepFunctionsActivity{
		name:         name,
		activityArn:  mo.None[string](),
		creationDate: time.Now(),
		tags:         make(map[string]string),
	}
}

// Accessor methods for FunctionalStepFunctionsActivity
func (a FunctionalStepFunctionsActivity) GetName() string { return a.name }
func (a FunctionalStepFunctionsActivity) GetActivityArn() mo.Option[string] { return a.activityArn }
func (a FunctionalStepFunctionsActivity) GetCreationDate() time.Time { return a.creationDate }
func (a FunctionalStepFunctionsActivity) GetTags() map[string]string { return a.tags }

// FunctionalStepFunctionsOperationResult represents the result of a Step Functions operation using monads
type FunctionalStepFunctionsOperationResult struct {
	result   mo.Option[interface{}]
	err      mo.Option[error]
	duration time.Duration
	metadata map[string]interface{}
}

// NewFunctionalStepFunctionsOperationResult creates a new operation result
func NewFunctionalStepFunctionsOperationResult(
	result mo.Option[interface{}],
	err mo.Option[error],
	duration time.Duration,
	metadata map[string]interface{},
) FunctionalStepFunctionsOperationResult {
	return FunctionalStepFunctionsOperationResult{
		result:   result,
		err:      err,
		duration: duration,
		metadata: metadata,
	}
}

// IsSuccess returns true if the operation was successful
func (r FunctionalStepFunctionsOperationResult) IsSuccess() bool {
	return r.err.IsAbsent()
}

// GetResult returns the operation result
func (r FunctionalStepFunctionsOperationResult) GetResult() mo.Option[interface{}] {
	return r.result
}

// GetError returns the operation error
func (r FunctionalStepFunctionsOperationResult) GetError() mo.Option[error] {
	return r.err
}

// GetDuration returns the operation duration
func (r FunctionalStepFunctionsOperationResult) GetDuration() time.Duration {
	return r.duration
}

// GetMetadata returns the operation metadata
func (r FunctionalStepFunctionsOperationResult) GetMetadata() map[string]interface{} {
	return r.metadata
}

// Validation functions using functional programming patterns
type FunctionalStepFunctionsValidationRule[T any] func(T) mo.Option[error]

// validateStateMachineName validates Step Functions state machine name
func validateStateMachineName(name string) mo.Option[error] {
	if name == "" {
		return mo.Some[error](fmt.Errorf("state machine name cannot be empty"))
	}

	if len(name) > FunctionalStepFunctionsMaxStateMachineNameLen {
		return mo.Some[error](fmt.Errorf("state machine name cannot exceed %d characters", 
			FunctionalStepFunctionsMaxStateMachineNameLen))
	}

	// Step Functions names can contain letters, numbers, hyphens, and underscores
	if !isValidStepFunctionsName(name) {
		return mo.Some[error](fmt.Errorf("invalid state machine name format: %s", name))
	}

	return mo.None[error]()
}

// validateExecutionName validates Step Functions execution name
func validateExecutionName(name string) mo.Option[error] {
	if name == "" {
		return mo.Some[error](fmt.Errorf("execution name cannot be empty"))
	}

	if len(name) > FunctionalStepFunctionsMaxExecutionNameLen {
		return mo.Some[error](fmt.Errorf("execution name cannot exceed %d characters", 
			FunctionalStepFunctionsMaxExecutionNameLen))
	}

	if !isValidStepFunctionsName(name) {
		return mo.Some[error](fmt.Errorf("invalid execution name format: %s", name))
	}

	return mo.None[error]()
}

// isValidStepFunctionsName checks if name follows Step Functions naming rules
func isValidStepFunctionsName(name string) bool {
	// Step Functions names can contain ASCII letters, numbers, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

// validateStateMachineDefinition validates Step Functions state machine definition
func validateStateMachineDefinition(definition string) mo.Option[error] {
	if definition == "" {
		return mo.Some[error](fmt.Errorf("state machine definition cannot be empty"))
	}

	if len(definition) > FunctionalStepFunctionsMaxDefinitionSize {
		return mo.Some[error](fmt.Errorf("state machine definition exceeds maximum size of %d bytes", 
			FunctionalStepFunctionsMaxDefinitionSize))
	}

	// Validate that definition is valid JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(definition), &jsonData); err != nil {
		return mo.Some[error](fmt.Errorf("invalid JSON in state machine definition: %v", err))
	}

	// Basic ASL validation - check for required fields
	definitionMap, ok := jsonData.(map[string]interface{})
	if !ok {
		return mo.Some[error](fmt.Errorf("state machine definition must be a JSON object"))
	}

	if _, hasStartAt := definitionMap["StartAt"]; !hasStartAt {
		return mo.Some[error](fmt.Errorf("state machine definition must have 'StartAt' field"))
	}

	if _, hasStates := definitionMap["States"]; !hasStates {
		return mo.Some[error](fmt.Errorf("state machine definition must have 'States' field"))
	}

	return mo.None[error]()
}

// validateRoleArn validates AWS IAM role ARN format
func validateRoleArn(roleArn string) mo.Option[error] {
	if roleArn == "" {
		return mo.Some[error](fmt.Errorf("role ARN cannot be empty"))
	}

	// Basic ARN validation
	arnPattern := `^arn:aws:iam::[0-9]{12}:role/.*$`
	matched, err := regexp.MatchString(arnPattern, roleArn)
	if err != nil {
		return mo.Some[error](fmt.Errorf("error validating role ARN pattern: %v", err))
	}

	if !matched {
		return mo.Some[error](fmt.Errorf("invalid role ARN format: %s", roleArn))
	}

	return mo.None[error]()
}

// validateExecutionInput validates Step Functions execution input
func validateExecutionInput(input string) mo.Option[error] {
	if input == "" {
		return mo.None[error]() // Empty input is valid
	}

	// Maximum input size for Step Functions is 32KB
	maxInputSize := 32 * 1024
	if len(input) > maxInputSize {
		return mo.Some[error](fmt.Errorf("execution input exceeds maximum size of %d bytes", maxInputSize))
	}

	// Validate that input is valid JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(input), &jsonData); err != nil {
		return mo.Some[error](fmt.Errorf("invalid JSON in execution input: %v", err))
	}

	return mo.None[error]()
}

// validateActivityName validates Step Functions activity name
func validateActivityName(name string) mo.Option[error] {
	if name == "" {
		return mo.Some[error](fmt.Errorf("activity name cannot be empty"))
	}

	if len(name) > FunctionalStepFunctionsMaxStateMachineNameLen {
		return mo.Some[error](fmt.Errorf("activity name cannot exceed %d characters", 
			FunctionalStepFunctionsMaxStateMachineNameLen))
	}

	if !isValidStepFunctionsName(name) {
		return mo.Some[error](fmt.Errorf("invalid activity name format: %s", name))
	}

	return mo.None[error]()
}

// Functional validation pipeline
func validateStepFunctionsConfig(config FunctionalStepFunctionsConfig) mo.Option[error] {
	// Config validation is primarily about checking timeout and polling values
	if config.timeout <= 0 {
		return mo.Some[error](fmt.Errorf("timeout must be positive"))
	}

	if config.pollingInterval <= 0 {
		return mo.Some[error](fmt.Errorf("polling interval must be positive"))
	}

	if config.maxRetries < 0 {
		return mo.Some[error](fmt.Errorf("max retries cannot be negative"))
	}

	return mo.None[error]()
}

// validateStepFunctionsStateMachine validates a Step Functions state machine
func validateStepFunctionsStateMachine(stateMachine FunctionalStepFunctionsStateMachine) mo.Option[error] {
	validators := []func() mo.Option[error]{
		func() mo.Option[error] { return validateStateMachineName(stateMachine.name) },
		func() mo.Option[error] { return validateStateMachineDefinition(stateMachine.definition) },
		func() mo.Option[error] { return validateRoleArn(stateMachine.roleArn) },
	}

	for _, validator := range validators {
		if err := validator(); err.IsPresent() {
			return err
		}
	}

	return mo.None[error]()
}

// validateStepFunctionsExecution validates a Step Functions execution
func validateStepFunctionsExecution(execution FunctionalStepFunctionsExecution) mo.Option[error] {
	validators := []func() mo.Option[error]{
		func() mo.Option[error] { return validateExecutionName(execution.name) },
		func() mo.Option[error] { return validateExecutionInput(execution.input) },
	}

	for _, validator := range validators {
		if err := validator(); err.IsPresent() {
			return err
		}
	}

	return mo.None[error]()
}

// Retry logic with exponential backoff using generics
func withStepFunctionsRetry[T any](operation func() (T, error), maxRetries int, retryDelay time.Duration) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoffDelay := time.Duration(float64(retryDelay) * (1.5 * float64(attempt)))
			time.Sleep(backoffDelay)
		}

		result, lastErr = operation()
		if lastErr == nil {
			return result, nil
		}

		// Continue retrying for transient errors
		if isStepFunctionsTransientError(lastErr) {
			continue
		}

		// Don't retry for permanent errors
		break
	}

	return result, lastErr
}

// isStepFunctionsTransientError determines if an error is transient and should be retried
func isStepFunctionsTransientError(err error) bool {
	if err == nil {
		return false
	}

	errorMessage := strings.ToLower(err.Error())
	transientErrors := []string{
		"throttling",
		"timeout",
		"service unavailable",
		"internal server error",
		"rate exceeded",
		"execution limit exceeded",
		"concurrent executions limit exceeded",
	}

	return lo.Some(transientErrors, func(transientErr string) bool {
		return strings.Contains(errorMessage, transientErr)
	})
}

// Simulated Step Functions operations for testing
func SimulateFunctionalStepFunctionsCreateStateMachine(stateMachine FunctionalStepFunctionsStateMachine, config FunctionalStepFunctionsConfig) FunctionalStepFunctionsOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":          "create_state_machine",
		"state_machine_name": stateMachine.name,
		"state_machine_type": string(stateMachine.stateMachineType),
		"phase":              "validation",
		"timestamp":          startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate configuration
	if configErr := validateStepFunctionsConfig(config); configErr.IsPresent() {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	// Validate state machine
	if smErr := validateStepFunctionsStateMachine(stateMachine); smErr.IsPresent() {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			smErr,
			time.Since(startTime),
			metadata,
		)
	}

	metadata["phase"] = "execution"

	// Simulate state machine creation with retry logic
	result, err := withStepFunctionsRetry(func() (map[string]interface{}, error) {
		// Simulate Step Functions create state machine operation
		time.Sleep(time.Duration(lo.RandomInt(50, 300)) * time.Millisecond)
		
		return map[string]interface{}{
			"state_machine_arn": fmt.Sprintf("arn:aws:states:us-east-1:123456789012:stateMachine:%s", stateMachine.name),
			"creation_date":     stateMachine.creationDate.Format(time.RFC3339),
			"name":              stateMachine.name,
			"type":              string(stateMachine.stateMachineType),
			"status":            string(types.StateMachineStatusActive),
			"role_arn":          stateMachine.roleArn,
			"created":           true,
		}, nil
	}, config.maxRetries, config.retryDelay)

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()

	if err != nil {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			duration,
			metadata,
		)
	}

	return NewFunctionalStepFunctionsOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	)
}

// SimulateFunctionalStepFunctionsStartExecution simulates starting a Step Functions execution
func SimulateFunctionalStepFunctionsStartExecution(execution FunctionalStepFunctionsExecution, config FunctionalStepFunctionsConfig) FunctionalStepFunctionsOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":         "start_execution",
		"execution_name":    execution.name,
		"state_machine_arn": execution.stateMachineArn,
		"input_size":        len(execution.input),
		"phase":             "validation",
		"timestamp":         startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate configuration
	if configErr := validateStepFunctionsConfig(config); configErr.IsPresent() {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	// Validate execution
	if execErr := validateStepFunctionsExecution(execution); execErr.IsPresent() {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			execErr,
			time.Since(startTime),
			metadata,
		)
	}

	metadata["phase"] = "execution"

	// Simulate execution start with retry logic
	result, err := withStepFunctionsRetry(func() (map[string]interface{}, error) {
		// Simulate Step Functions start execution operation
		time.Sleep(time.Duration(lo.RandomInt(30, 200)) * time.Millisecond)
		
		return map[string]interface{}{
			"execution_arn":     fmt.Sprintf("%s:execution:%s:%s", execution.stateMachineArn, execution.name, fmt.Sprint(time.Now().Unix())),
			"start_date":        execution.startDate.Format(time.RFC3339),
			"name":              execution.name,
			"state_machine_arn": execution.stateMachineArn,
			"status":            string(types.ExecutionStatusRunning),
			"input":             execution.input,
			"started":           true,
		}, nil
	}, config.maxRetries, config.retryDelay)

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()

	if err != nil {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			duration,
			metadata,
		)
	}

	return NewFunctionalStepFunctionsOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	)
}

// SimulateFunctionalStepFunctionsDescribeExecution simulates describing a Step Functions execution
func SimulateFunctionalStepFunctionsDescribeExecution(executionArn string, config FunctionalStepFunctionsConfig) FunctionalStepFunctionsOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":      "describe_execution",
		"execution_arn":  executionArn,
		"phase":          "validation",
		"timestamp":      startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate configuration
	if configErr := validateStepFunctionsConfig(config); configErr.IsPresent() {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	// Basic ARN validation
	if executionArn == "" {
		err := fmt.Errorf("execution ARN cannot be empty")
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			time.Since(startTime),
			metadata,
		)
	}

	metadata["phase"] = "execution"

	// Simulate execution description with retry logic
	result, err := withStepFunctionsRetry(func() (map[string]interface{}, error) {
		// Simulate Step Functions describe execution operation
		time.Sleep(time.Duration(lo.RandomInt(20, 150)) * time.Millisecond)
		
		// Simulate different execution statuses
		statuses := []types.ExecutionStatus{
			types.ExecutionStatusRunning,
			types.ExecutionStatusSucceeded,
			types.ExecutionStatusFailed,
		}
		
		status := statuses[lo.RandomInt(0, len(statuses))]
		
		executionResult := map[string]interface{}{
			"execution_arn":     executionArn,
			"status":            string(status),
			"start_date":        time.Now().Add(-time.Minute).Format(time.RFC3339),
			"name":              "test-execution",
			"state_machine_arn": "arn:aws:states:us-east-1:123456789012:stateMachine:test-sm",
			"input":             `{"message": "hello"}`,
		}

		// Add completion data for finished executions
		if status != types.ExecutionStatusRunning {
			executionResult["stop_date"] = time.Now().Format(time.RFC3339)
			if status == types.ExecutionStatusSucceeded {
				executionResult["output"] = `{"result": "success"}`
			} else {
				executionResult["error"] = "States.TaskFailed"
				executionResult["cause"] = "Task execution failed"
			}
		}

		return executionResult, nil
	}, config.maxRetries, config.retryDelay)

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()

	if err != nil {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			duration,
			metadata,
		)
	}

	return NewFunctionalStepFunctionsOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	)
}

// SimulateFunctionalStepFunctionsCreateActivity simulates creating a Step Functions activity
func SimulateFunctionalStepFunctionsCreateActivity(activity FunctionalStepFunctionsActivity, config FunctionalStepFunctionsConfig) FunctionalStepFunctionsOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":     "create_activity",
		"activity_name": activity.name,
		"phase":         "validation",
		"timestamp":     startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate configuration
	if configErr := validateStepFunctionsConfig(config); configErr.IsPresent() {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	// Validate activity name
	if actErr := validateActivityName(activity.name); actErr.IsPresent() {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			actErr,
			time.Since(startTime),
			metadata,
		)
	}

	metadata["phase"] = "execution"

	// Simulate activity creation with retry logic
	result, err := withStepFunctionsRetry(func() (map[string]interface{}, error) {
		// Simulate Step Functions create activity operation
		time.Sleep(time.Duration(lo.RandomInt(30, 200)) * time.Millisecond)
		
		return map[string]interface{}{
			"activity_arn":   fmt.Sprintf("arn:aws:states:us-east-1:123456789012:activity:%s", activity.name),
			"creation_date":  activity.creationDate.Format(time.RFC3339),
			"name":           activity.name,
			"created":        true,
		}, nil
	}, config.maxRetries, config.retryDelay)

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()

	if err != nil {
		return NewFunctionalStepFunctionsOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			duration,
			metadata,
		)
	}

	return NewFunctionalStepFunctionsOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	)
}