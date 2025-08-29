package eventbridge

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// Default values for functional EventBridge operations
const (
	FunctionalEventBridgeDefaultTimeout        = 30 * time.Second
	FunctionalEventBridgeDefaultMaxRetries     = 3
	FunctionalEventBridgeDefaultRetryDelay     = 200 * time.Millisecond
	FunctionalEventBridgeMaxEventBatchSize     = 10
	FunctionalEventBridgeMaxEventDetailSize    = 256 * 1024 // 256KB
	FunctionalEventBridgeDefaultEventBusName   = "default"
)

// FunctionalEventBridgeConfig represents immutable EventBridge configuration using functional options
type FunctionalEventBridgeConfig struct {
	eventBusName       string
	timeout           time.Duration
	maxRetries        int
	retryDelay        time.Duration
	eventPattern      mo.Option[string]
	scheduleExpression mo.Option[string]
	ruleState         types.RuleState
	deadLetterConfig  mo.Option[types.DeadLetterConfig]
	retryPolicy       mo.Option[types.RetryPolicy]
	metadata          map[string]interface{}
}

// FunctionalEventBridgeConfigOption represents a functional option for EventBridge configuration
type FunctionalEventBridgeConfigOption func(*FunctionalEventBridgeConfig)

// NewFunctionalEventBridgeConfig creates a new immutable EventBridge configuration with functional options
func NewFunctionalEventBridgeConfig(eventBusName string, opts ...FunctionalEventBridgeConfigOption) FunctionalEventBridgeConfig {
	config := FunctionalEventBridgeConfig{
		eventBusName:       eventBusName,
		timeout:           FunctionalEventBridgeDefaultTimeout,
		maxRetries:        FunctionalEventBridgeDefaultMaxRetries,
		retryDelay:        FunctionalEventBridgeDefaultRetryDelay,
		eventPattern:      mo.None[string](),
		scheduleExpression: mo.None[string](),
		ruleState:         types.RuleStateEnabled,
		deadLetterConfig:  mo.None[types.DeadLetterConfig](),
		retryPolicy:       mo.None[types.RetryPolicy](),
		metadata:          make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(&config)
	}

	return config
}

// Functional options for EventBridge configuration
func WithEventBridgeTimeout(timeout time.Duration) FunctionalEventBridgeConfigOption {
	return func(c *FunctionalEventBridgeConfig) {
		c.timeout = timeout
	}
}

func WithEventBridgeRetryConfig(maxRetries int, retryDelay time.Duration) FunctionalEventBridgeConfigOption {
	return func(c *FunctionalEventBridgeConfig) {
		c.maxRetries = maxRetries
		c.retryDelay = retryDelay
	}
}

func WithEventPattern(pattern string) FunctionalEventBridgeConfigOption {
	return func(c *FunctionalEventBridgeConfig) {
		c.eventPattern = mo.Some(pattern)
	}
}

func WithScheduleExpression(expression string) FunctionalEventBridgeConfigOption {
	return func(c *FunctionalEventBridgeConfig) {
		c.scheduleExpression = mo.Some(expression)
	}
}

func WithRuleState(state types.RuleState) FunctionalEventBridgeConfigOption {
	return func(c *FunctionalEventBridgeConfig) {
		c.ruleState = state
	}
}

func WithDeadLetterConfig(config types.DeadLetterConfig) FunctionalEventBridgeConfigOption {
	return func(c *FunctionalEventBridgeConfig) {
		c.deadLetterConfig = mo.Some(config)
	}
}

func WithRetryPolicy(policy types.RetryPolicy) FunctionalEventBridgeConfigOption {
	return func(c *FunctionalEventBridgeConfig) {
		c.retryPolicy = mo.Some(policy)
	}
}

func WithEventBridgeMetadata(key string, value interface{}) FunctionalEventBridgeConfigOption {
	return func(c *FunctionalEventBridgeConfig) {
		c.metadata[key] = value
	}
}

// FunctionalEventBridgeEvent represents an immutable EventBridge event
type FunctionalEventBridgeEvent struct {
	source        string
	detailType    string
	detail        string
	eventBusName  string
	resources     []string
	time          time.Time
	eventId       mo.Option[string]
	version       mo.Option[string]
	region        mo.Option[string]
	account       mo.Option[string]
}

// NewFunctionalEventBridgeEvent creates a new immutable EventBridge event
func NewFunctionalEventBridgeEvent(source, detailType, detail string) FunctionalEventBridgeEvent {
	return FunctionalEventBridgeEvent{
		source:        source,
		detailType:    detailType,
		detail:        detail,
		eventBusName:  FunctionalEventBridgeDefaultEventBusName,
		resources:     []string{},
		time:          time.Now(),
		eventId:       mo.None[string](),
		version:       mo.Some("0"),
		region:        mo.None[string](),
		account:       mo.None[string](),
	}
}

// Accessor methods for FunctionalEventBridgeEvent
func (e FunctionalEventBridgeEvent) GetSource() string               { return e.source }
func (e FunctionalEventBridgeEvent) GetDetailType() string          { return e.detailType }
func (e FunctionalEventBridgeEvent) GetDetail() string              { return e.detail }
func (e FunctionalEventBridgeEvent) GetEventBusName() string        { return e.eventBusName }
func (e FunctionalEventBridgeEvent) GetResources() []string         { return e.resources }
func (e FunctionalEventBridgeEvent) GetTime() time.Time             { return e.time }
func (e FunctionalEventBridgeEvent) GetEventId() mo.Option[string]  { return e.eventId }
func (e FunctionalEventBridgeEvent) GetVersion() mo.Option[string]  { return e.version }
func (e FunctionalEventBridgeEvent) GetRegion() mo.Option[string]   { return e.region }
func (e FunctionalEventBridgeEvent) GetAccount() mo.Option[string]  { return e.account }

// Functional options for event creation
func (e FunctionalEventBridgeEvent) WithEventBusName(eventBusName string) FunctionalEventBridgeEvent {
	return FunctionalEventBridgeEvent{
		source:       e.source,
		detailType:   e.detailType,
		detail:       e.detail,
		eventBusName: eventBusName,
		resources:    e.resources,
		time:         e.time,
		eventId:      e.eventId,
		version:      e.version,
		region:       e.region,
		account:      e.account,
	}
}

func (e FunctionalEventBridgeEvent) WithResources(resources []string) FunctionalEventBridgeEvent {
	return FunctionalEventBridgeEvent{
		source:       e.source,
		detailType:   e.detailType,
		detail:       e.detail,
		eventBusName: e.eventBusName,
		resources:    resources,
		time:         e.time,
		eventId:      e.eventId,
		version:      e.version,
		region:       e.region,
		account:      e.account,
	}
}

func (e FunctionalEventBridgeEvent) WithTime(eventTime time.Time) FunctionalEventBridgeEvent {
	return FunctionalEventBridgeEvent{
		source:       e.source,
		detailType:   e.detailType,
		detail:       e.detail,
		eventBusName: e.eventBusName,
		resources:    e.resources,
		time:         eventTime,
		eventId:      e.eventId,
		version:      e.version,
		region:       e.region,
		account:      e.account,
	}
}

func (e FunctionalEventBridgeEvent) WithEventId(eventId string) FunctionalEventBridgeEvent {
	return FunctionalEventBridgeEvent{
		source:       e.source,
		detailType:   e.detailType,
		detail:       e.detail,
		eventBusName: e.eventBusName,
		resources:    e.resources,
		time:         e.time,
		eventId:      mo.Some(eventId),
		version:      e.version,
		region:       e.region,
		account:      e.account,
	}
}

func (e FunctionalEventBridgeEvent) WithRegion(region string) FunctionalEventBridgeEvent {
	return FunctionalEventBridgeEvent{
		source:       e.source,
		detailType:   e.detailType,
		detail:       e.detail,
		eventBusName: e.eventBusName,
		resources:    e.resources,
		time:         e.time,
		eventId:      e.eventId,
		version:      e.version,
		region:       mo.Some(region),
		account:      e.account,
	}
}

func (e FunctionalEventBridgeEvent) WithAccount(account string) FunctionalEventBridgeEvent {
	return FunctionalEventBridgeEvent{
		source:       e.source,
		detailType:   e.detailType,
		detail:       e.detail,
		eventBusName: e.eventBusName,
		resources:    e.resources,
		time:         e.time,
		eventId:      e.eventId,
		version:      e.version,
		region:       e.region,
		account:      mo.Some(account),
	}
}

// FunctionalEventBridgeRule represents an immutable EventBridge rule
type FunctionalEventBridgeRule struct {
	name               string
	description        mo.Option[string]
	eventPattern       mo.Option[string]
	scheduleExpression mo.Option[string]
	state              types.RuleState
	eventBusName       string
	ruleArn            mo.Option[string]
	targets            []FunctionalEventBridgeTarget
}

// NewFunctionalEventBridgeRule creates a new immutable EventBridge rule
func NewFunctionalEventBridgeRule(name, eventBusName string) FunctionalEventBridgeRule {
	return FunctionalEventBridgeRule{
		name:               name,
		description:        mo.None[string](),
		eventPattern:       mo.None[string](),
		scheduleExpression: mo.None[string](),
		state:              types.RuleStateEnabled,
		eventBusName:       eventBusName,
		ruleArn:            mo.None[string](),
		targets:            []FunctionalEventBridgeTarget{},
	}
}

// Accessor methods for FunctionalEventBridgeRule
func (r FunctionalEventBridgeRule) GetName() string                           { return r.name }
func (r FunctionalEventBridgeRule) GetDescription() mo.Option[string]        { return r.description }
func (r FunctionalEventBridgeRule) GetEventPattern() mo.Option[string]       { return r.eventPattern }
func (r FunctionalEventBridgeRule) GetScheduleExpression() mo.Option[string] { return r.scheduleExpression }
func (r FunctionalEventBridgeRule) GetState() types.RuleState                { return r.state }
func (r FunctionalEventBridgeRule) GetEventBusName() string                  { return r.eventBusName }
func (r FunctionalEventBridgeRule) GetRuleArn() mo.Option[string]            { return r.ruleArn }
func (r FunctionalEventBridgeRule) GetTargets() []FunctionalEventBridgeTarget { return r.targets }

// FunctionalEventBridgeTarget represents an immutable EventBridge target
type FunctionalEventBridgeTarget struct {
	id               string
	arn              string
	roleArn          mo.Option[string]
	input            mo.Option[string]
	inputPath        mo.Option[string]
	inputTransformer mo.Option[types.InputTransformer]
	deadLetterConfig mo.Option[types.DeadLetterConfig]
	retryPolicy      mo.Option[types.RetryPolicy]
}

// NewFunctionalEventBridgeTarget creates a new immutable EventBridge target
func NewFunctionalEventBridgeTarget(id, arn string) FunctionalEventBridgeTarget {
	return FunctionalEventBridgeTarget{
		id:               id,
		arn:              arn,
		roleArn:          mo.None[string](),
		input:            mo.None[string](),
		inputPath:        mo.None[string](),
		inputTransformer: mo.None[types.InputTransformer](),
		deadLetterConfig: mo.None[types.DeadLetterConfig](),
		retryPolicy:      mo.None[types.RetryPolicy](),
	}
}

// Accessor methods for FunctionalEventBridgeTarget
func (t FunctionalEventBridgeTarget) GetId() string                                 { return t.id }
func (t FunctionalEventBridgeTarget) GetArn() string                               { return t.arn }
func (t FunctionalEventBridgeTarget) GetRoleArn() mo.Option[string]                { return t.roleArn }
func (t FunctionalEventBridgeTarget) GetInput() mo.Option[string]                  { return t.input }
func (t FunctionalEventBridgeTarget) GetInputPath() mo.Option[string]              { return t.inputPath }
func (t FunctionalEventBridgeTarget) GetInputTransformer() mo.Option[types.InputTransformer] { return t.inputTransformer }
func (t FunctionalEventBridgeTarget) GetDeadLetterConfig() mo.Option[types.DeadLetterConfig] { return t.deadLetterConfig }
func (t FunctionalEventBridgeTarget) GetRetryPolicy() mo.Option[types.RetryPolicy] { return t.retryPolicy }

// FunctionalEventBridgeOperationResult represents the result of an EventBridge operation using monads
type FunctionalEventBridgeOperationResult struct {
	result   mo.Option[interface{}]
	err      mo.Option[error]
	duration time.Duration
	metadata map[string]interface{}
}

// NewFunctionalEventBridgeOperationResult creates a new operation result
func NewFunctionalEventBridgeOperationResult(
	result mo.Option[interface{}],
	err mo.Option[error],
	duration time.Duration,
	metadata map[string]interface{},
) FunctionalEventBridgeOperationResult {
	return FunctionalEventBridgeOperationResult{
		result:   result,
		err:      err,
		duration: duration,
		metadata: metadata,
	}
}

// IsSuccess returns true if the operation was successful
func (r FunctionalEventBridgeOperationResult) IsSuccess() bool {
	return r.err.IsAbsent()
}

// GetResult returns the operation result
func (r FunctionalEventBridgeOperationResult) GetResult() mo.Option[interface{}] {
	return r.result
}

// GetError returns the operation error
func (r FunctionalEventBridgeOperationResult) GetError() mo.Option[error] {
	return r.err
}

// GetDuration returns the operation duration
func (r FunctionalEventBridgeOperationResult) GetDuration() time.Duration {
	return r.duration
}

// GetMetadata returns the operation metadata
func (r FunctionalEventBridgeOperationResult) GetMetadata() map[string]interface{} {
	return r.metadata
}

// Validation functions using functional programming patterns
type FunctionalEventBridgeValidationRule[T any] func(T) mo.Option[error]

// validateEventBusName validates EventBridge event bus name
func validateEventBusName(eventBusName string) mo.Option[error] {
	if eventBusName == "" {
		return mo.Some[error](fmt.Errorf("event bus name cannot be empty"))
	}

	if len(eventBusName) > 256 {
		return mo.Some[error](fmt.Errorf("event bus name cannot exceed 256 characters"))
	}

	// EventBridge event bus names must follow specific patterns
	if eventBusName != "default" && !isValidEventBusName(eventBusName) {
		return mo.Some[error](fmt.Errorf("invalid event bus name format: %s", eventBusName))
	}

	return mo.None[error]()
}

// isValidEventBusName checks if event bus name follows EventBridge naming rules
func isValidEventBusName(name string) bool {
	// EventBridge custom event bus names can contain alphanumeric characters, dots, hyphens, and underscores
	// but cannot start or end with a dot or hyphen
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") || 
	   strings.HasSuffix(name, ".") || strings.HasSuffix(name, "-") {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '.' || char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// validateEventSource validates EventBridge event source
func validateEventSource(source string) mo.Option[error] {
	if source == "" {
		return mo.Some[error](fmt.Errorf("event source cannot be empty"))
	}

	if len(source) > 256 {
		return mo.Some[error](fmt.Errorf("event source cannot exceed 256 characters"))
	}

	// EventBridge sources typically use reverse domain notation
	if !isValidEventSource(source) {
		return mo.Some[error](fmt.Errorf("invalid event source format: %s", source))
	}

	return mo.None[error]()
}

// isValidEventSource checks if event source follows EventBridge naming conventions
func isValidEventSource(source string) bool {
	// EventBridge sources can contain alphanumeric characters, dots, hyphens
	// Common patterns: myapp.orders, aws.s3, custom.service
	for _, char := range source {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '.' || char == '-') {
			return false
		}
	}

	return true
}

// validateEventDetail validates EventBridge event detail JSON
func validateEventDetail(detail string) mo.Option[error] {
	if detail == "" {
		return mo.None[error]() // Empty detail is valid
	}

	if len(detail) > FunctionalEventBridgeMaxEventDetailSize {
		return mo.Some[error](fmt.Errorf("event detail exceeds maximum size of %d bytes", 
			FunctionalEventBridgeMaxEventDetailSize))
	}

	var jsonData interface{}
	if err := json.Unmarshal([]byte(detail), &jsonData); err != nil {
		return mo.Some[error](fmt.Errorf("invalid JSON in event detail: %v", err))
	}

	return mo.None[error]()
}

// validateEventPattern validates EventBridge event pattern JSON
func validateEventPattern(pattern string) mo.Option[error] {
	if pattern == "" {
		return mo.None[error]() // Empty pattern is valid
	}

	var patternData interface{}
	if err := json.Unmarshal([]byte(pattern), &patternData); err != nil {
		return mo.Some[error](fmt.Errorf("invalid JSON in event pattern: %v", err))
	}

	return mo.None[error]()
}

// validateScheduleExpression validates EventBridge schedule expression
func validateScheduleExpression(expression string) mo.Option[error] {
	if expression == "" {
		return mo.None[error]() // Empty expression is valid
	}

	// Basic validation for rate() and cron() expressions
	if !strings.HasPrefix(expression, "rate(") && !strings.HasPrefix(expression, "cron(") {
		return mo.Some[error](fmt.Errorf("schedule expression must start with 'rate(' or 'cron('"))
	}

	if !strings.HasSuffix(expression, ")") {
		return mo.Some[error](fmt.Errorf("schedule expression must end with ')'"))
	}

	return mo.None[error]()
}

// Functional validation pipeline
func validateEventBridgeConfig(config FunctionalEventBridgeConfig) mo.Option[error] {
	validators := []func() mo.Option[error]{
		func() mo.Option[error] { return validateEventBusName(config.eventBusName) },
		func() mo.Option[error] {
			return config.eventPattern.
				Map(validateEventPattern).
				OrElse(mo.None[error]())
		},
		func() mo.Option[error] {
			return config.scheduleExpression.
				Map(validateScheduleExpression).
				OrElse(mo.None[error]())
		},
	}

	for _, validator := range validators {
		if err := validator(); err.IsPresent() {
			return err
		}
	}

	return mo.None[error]()
}

// validateEventBridgeEvent validates an EventBridge event
func validateEventBridgeEvent(event FunctionalEventBridgeEvent) mo.Option[error] {
	validators := []func() mo.Option[error]{
		func() mo.Option[error] { return validateEventBusName(event.eventBusName) },
		func() mo.Option[error] { return validateEventSource(event.source) },
		func() mo.Option[error] { return validateEventDetail(event.detail) },
	}

	for _, validator := range validators {
		if err := validator(); err.IsPresent() {
			return err
		}
	}

	return mo.None[error]()
}

// Retry logic with exponential backoff using generics
func withEventBridgeRetry[T any](operation func() (T, error), maxRetries int, retryDelay time.Duration) (T, error) {
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
		if isTransientError(lastErr) {
			continue
		}

		// Don't retry for permanent errors
		break
	}

	return result, lastErr
}

// isTransientError determines if an error is transient and should be retried
func isTransientError(err error) bool {
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
	}

	return lo.Some(transientErrors, func(transientErr string) bool {
		return strings.Contains(errorMessage, transientErr)
	})
}

// Simulated EventBridge operations for testing
func SimulateFunctionalEventBridgePutEvent(event FunctionalEventBridgeEvent, config FunctionalEventBridgeConfig) FunctionalEventBridgeOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":     "put_event",
		"source":        event.source,
		"detail_type":   event.detailType,
		"event_bus":     event.eventBusName,
		"phase":         "validation",
		"timestamp":     startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate configuration
	if configErr := validateEventBridgeConfig(config); configErr.IsPresent() {
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	// Validate event
	if eventErr := validateEventBridgeEvent(event); eventErr.IsPresent() {
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			eventErr,
			time.Since(startTime),
			metadata,
		)
	}

	metadata["phase"] = "execution"

	// Simulate putting event with retry logic
	result, err := withEventBridgeRetry(func() (map[string]interface{}, error) {
		// Simulate EventBridge put event operation
		time.Sleep(time.Duration(lo.RandomInt(10, 100)) * time.Millisecond)
		
		return map[string]interface{}{
			"event_id":      fmt.Sprintf("event-%d", time.Now().Unix()),
			"source":        event.source,
			"detail_type":   event.detailType,
			"detail_size":   len(event.detail),
			"event_bus":     event.eventBusName,
			"resources":     len(event.resources),
			"success":       true,
		}, nil
	}, config.maxRetries, config.retryDelay)

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()

	if err != nil {
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			duration,
			metadata,
		)
	}

	return NewFunctionalEventBridgeOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	)
}

// SimulateFunctionalEventBridgePutEvents simulates putting multiple events
func SimulateFunctionalEventBridgePutEvents(events []FunctionalEventBridgeEvent, config FunctionalEventBridgeConfig) FunctionalEventBridgeOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":    "put_events",
		"event_count":  len(events),
		"event_bus":    config.eventBusName,
		"phase":        "validation",
		"timestamp":    startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate batch size
	if len(events) > FunctionalEventBridgeMaxEventBatchSize {
		err := fmt.Errorf("batch size exceeds maximum: %d events", len(events))
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			time.Since(startTime),
			metadata,
		)
	}

	// Validate configuration
	if configErr := validateEventBridgeConfig(config); configErr.IsPresent() {
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	// Validate all events
	for i, event := range events {
		if eventErr := validateEventBridgeEvent(event); eventErr.IsPresent() {
			metadata["failed_event_index"] = i
			return NewFunctionalEventBridgeOperationResult(
				mo.None[interface{}](),
				eventErr,
				time.Since(startTime),
				metadata,
			)
		}
	}

	metadata["phase"] = "execution"

	// Simulate putting events with retry logic
	result, err := withEventBridgeRetry(func() (map[string]interface{}, error) {
		// Simulate EventBridge put events operation
		time.Sleep(time.Duration(lo.RandomInt(50, 200)) * time.Millisecond)
		
		// Process each event
		results := lo.Map(events, func(event FunctionalEventBridgeEvent, index int) map[string]interface{} {
			return map[string]interface{}{
				"event_id":      fmt.Sprintf("event-%d-%d", time.Now().Unix(), index),
				"source":        event.source,
				"detail_type":   event.detailType,
				"success":       true,
			}
		})

		return map[string]interface{}{
			"failed_entry_count": 0,
			"entries":           results,
			"total_events":      len(events),
		}, nil
	}, config.maxRetries, config.retryDelay)

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()

	if err != nil {
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			duration,
			metadata,
		)
	}

	return NewFunctionalEventBridgeOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	)
}

// SimulateFunctionalEventBridgeCreateRule simulates creating an EventBridge rule
func SimulateFunctionalEventBridgeCreateRule(rule FunctionalEventBridgeRule, config FunctionalEventBridgeConfig) FunctionalEventBridgeOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":    "create_rule",
		"rule_name":    rule.name,
		"event_bus":    rule.eventBusName,
		"rule_state":   string(rule.state),
		"phase":        "validation",
		"timestamp":    startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate configuration
	if configErr := validateEventBridgeConfig(config); configErr.IsPresent() {
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	// Validate rule
	if rule.name == "" {
		err := fmt.Errorf("rule name cannot be empty")
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			time.Since(startTime),
			metadata,
		)
	}

	// Validate event pattern if present
	if rule.eventPattern.IsPresent() {
		if patternErr := validateEventPattern(rule.eventPattern.MustGet()); patternErr.IsPresent() {
			return NewFunctionalEventBridgeOperationResult(
				mo.None[interface{}](),
				patternErr,
				time.Since(startTime),
				metadata,
			)
		}
	}

	// Validate schedule expression if present
	if rule.scheduleExpression.IsPresent() {
		if scheduleErr := validateScheduleExpression(rule.scheduleExpression.MustGet()); scheduleErr.IsPresent() {
			return NewFunctionalEventBridgeOperationResult(
				mo.None[interface{}](),
				scheduleErr,
				time.Since(startTime),
				metadata,
			)
		}
	}

	metadata["phase"] = "execution"

	// Simulate rule creation with retry logic
	result, err := withEventBridgeRetry(func() (map[string]interface{}, error) {
		// Simulate EventBridge create rule operation
		time.Sleep(time.Duration(lo.RandomInt(20, 150)) * time.Millisecond)
		
		return map[string]interface{}{
			"rule_arn":     fmt.Sprintf("arn:aws:events:us-east-1:123456789012:rule/%s/%s", rule.eventBusName, rule.name),
			"rule_name":    rule.name,
			"state":        string(rule.state),
			"event_bus":    rule.eventBusName,
			"created":      true,
		}, nil
	}, config.maxRetries, config.retryDelay)

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()

	if err != nil {
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			duration,
			metadata,
		)
	}

	return NewFunctionalEventBridgeOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	)
}

// SimulateFunctionalEventBridgeCreateEventBus simulates creating an EventBridge custom event bus
func SimulateFunctionalEventBridgeCreateEventBus(config FunctionalEventBridgeConfig) FunctionalEventBridgeOperationResult {
	startTime := time.Now()
	metadata := map[string]interface{}{
		"operation":   "create_event_bus",
		"event_bus":   config.eventBusName,
		"phase":       "validation",
		"timestamp":   startTime,
	}

	// Add config metadata
	for k, v := range config.metadata {
		metadata[k] = v
	}

	// Validate configuration
	if configErr := validateEventBridgeConfig(config); configErr.IsPresent() {
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			configErr,
			time.Since(startTime),
			metadata,
		)
	}

	metadata["phase"] = "execution"

	// Simulate event bus creation with retry logic
	result, err := withEventBridgeRetry(func() (map[string]interface{}, error) {
		// Simulate EventBridge create event bus operation
		time.Sleep(time.Duration(lo.RandomInt(30, 200)) * time.Millisecond)
		
		return map[string]interface{}{
			"event_bus_arn":  fmt.Sprintf("arn:aws:events:us-east-1:123456789012:event-bus/%s", config.eventBusName),
			"event_bus_name": config.eventBusName,
			"created":        true,
		}, nil
	}, config.maxRetries, config.retryDelay)

	duration := time.Since(startTime)
	metadata["phase"] = "completed"
	metadata["duration_ms"] = duration.Milliseconds()

	if err != nil {
		return NewFunctionalEventBridgeOperationResult(
			mo.None[interface{}](),
			mo.Some[error](err),
			duration,
			metadata,
		)
	}

	return NewFunctionalEventBridgeOperationResult(
		mo.Some[interface{}](result),
		mo.None[error](),
		duration,
		metadata,
	)
}