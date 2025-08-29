package dsl

import (
	"encoding/json"
	"fmt"
	"time"
)

// EventBridgeBuilder provides a fluent interface for EventBridge operations
type EventBridgeBuilder struct {
	ctx *TestContext
}

// EventBus creates an event bus builder for a specific bus
func (eb *EventBridgeBuilder) EventBus(name string) *EventBusBuilder {
	return &EventBusBuilder{
		ctx:          eb.ctx,
		eventBusName: name,
		config:       &EventBusConfig{},
	}
}

// CreateEventBus starts building a new event bus
func (eb *EventBridgeBuilder) CreateEventBus() *EventBusCreateBuilder {
	return &EventBusCreateBuilder{
		ctx:    eb.ctx,
		config: &EventBusCreateConfig{},
	}
}

// Rule creates a rule builder for a specific rule
func (eb *EventBridgeBuilder) Rule(name string) *EventRuleBuilder {
	return &EventRuleBuilder{
		ctx:      eb.ctx,
		ruleName: name,
		config:   &RuleConfig{},
	}
}

// PublishEvent creates an event publisher
func (eb *EventBridgeBuilder) PublishEvent() *EventPublishBuilder {
	return &EventPublishBuilder{
		ctx:    eb.ctx,
		config: &EventPublishConfig{},
	}
}

// Archive creates an archive builder
func (eb *EventBridgeBuilder) Archive(name string) *EventArchiveBuilder {
	return &EventArchiveBuilder{
		ctx:         eb.ctx,
		archiveName: name,
		config:      &ArchiveConfig{},
	}
}

// EventBusBuilder builds operations on an existing EventBridge event bus
type EventBusBuilder struct {
	ctx          *TestContext
	eventBusName string
	config       *EventBusConfig
}

type EventBusConfig struct {
	KMSKeyId string
	Tags     map[string]string
}

// WithKMSKey sets the KMS key for encryption
func (ebb *EventBusBuilder) WithKMSKey(keyId string) *EventBusBuilder {
	ebb.config.KMSKeyId = keyId
	return ebb
}

// Delete deletes the event bus
func (ebb *EventBusBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Event bus %s deleted", ebb.eventBusName),
		ctx:     ebb.ctx,
	}
}

// Describe describes the event bus
func (ebb *EventBusBuilder) Describe() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"Name":        ebb.eventBusName,
			"Arn":         fmt.Sprintf("arn:aws:events:us-east-1:123456789012:event-bus/%s", ebb.eventBusName),
			"CreatedTime": time.Now(),
		},
		ctx: ebb.ctx,
	}
}

// EventBusCreateBuilder builds event bus creation operations
type EventBusCreateBuilder struct {
	ctx    *TestContext
	config *EventBusCreateConfig
}

type EventBusCreateConfig struct {
	Name           string
	EventSourceName string
	KMSKeyId       string
	Tags           map[string]string
}

// Named sets the event bus name
func (ebcb *EventBusCreateBuilder) Named(name string) *EventBusCreateBuilder {
	ebcb.config.Name = name
	return ebcb
}

// WithEventSource sets the event source
func (ebcb *EventBusCreateBuilder) WithEventSource(sourceName string) *EventBusCreateBuilder {
	ebcb.config.EventSourceName = sourceName
	return ebcb
}

// WithKMSKey sets the KMS key
func (ebcb *EventBusCreateBuilder) WithKMSKey(keyId string) *EventBusCreateBuilder {
	ebcb.config.KMSKeyId = keyId
	return ebcb
}

// Create creates the event bus
func (ebcb *EventBusCreateBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"EventBusArn": fmt.Sprintf("arn:aws:events:us-east-1:123456789012:event-bus/%s", ebcb.config.Name),
			"Name":        ebcb.config.Name,
		},
		ctx: ebcb.ctx,
	}
}

// EventRuleBuilder builds operations on EventBridge rules
type EventRuleBuilder struct {
	ctx      *TestContext
	ruleName string
	config   *RuleConfig
}

type RuleConfig struct {
	EventBusName    string
	EventPattern    map[string]interface{}
	ScheduleExpression string
	State           string
	Description     string
	Targets         []RuleTarget
}

type RuleTarget struct {
	Id      string
	Arn     string
	RoleArn string
	Input   string
}

// OnEventBus sets the event bus for the rule
func (erb *EventRuleBuilder) OnEventBus(eventBusName string) *EventRuleBuilder {
	erb.config.EventBusName = eventBusName
	return erb
}

// WithEventPattern sets the event pattern for the rule
func (erb *EventRuleBuilder) WithEventPattern(pattern map[string]interface{}) *EventRuleBuilder {
	erb.config.EventPattern = pattern
	return erb
}

// WithScheduleExpression sets a schedule expression
func (erb *EventRuleBuilder) WithScheduleExpression(expression string) *EventRuleBuilder {
	erb.config.ScheduleExpression = expression
	return erb
}

// WithDescription sets the rule description
func (erb *EventRuleBuilder) WithDescription(description string) *EventRuleBuilder {
	erb.config.Description = description
	return erb
}

// Enabled sets the rule state to enabled
func (erb *EventRuleBuilder) Enabled() *EventRuleBuilder {
	erb.config.State = "ENABLED"
	return erb
}

// Disabled sets the rule state to disabled
func (erb *EventRuleBuilder) Disabled() *EventRuleBuilder {
	erb.config.State = "DISABLED"
	return erb
}

// AddTarget adds a target to the rule
func (erb *EventRuleBuilder) AddTarget() *RuleTargetBuilder {
	return &RuleTargetBuilder{
		ruleBuilder: erb,
		target:      &RuleTarget{},
	}
}

// Create creates the rule
func (erb *EventRuleBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"RuleArn": fmt.Sprintf("arn:aws:events:us-east-1:123456789012:rule/%s", erb.ruleName),
			"Name":    erb.ruleName,
		},
		ctx: erb.ctx,
	}
}

// Delete deletes the rule
func (erb *EventRuleBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Rule %s deleted", erb.ruleName),
		ctx:     erb.ctx,
	}
}

// Describe describes the rule
func (erb *EventRuleBuilder) Describe() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"Name":        erb.ruleName,
			"State":       erb.config.State,
			"Description": erb.config.Description,
		},
		ctx: erb.ctx,
	}
}

// RuleTargetBuilder builds rule targets
type RuleTargetBuilder struct {
	ruleBuilder *EventRuleBuilder
	target      *RuleTarget
}

// WithId sets the target ID
func (rtb *RuleTargetBuilder) WithId(id string) *RuleTargetBuilder {
	rtb.target.Id = id
	return rtb
}

// WithArn sets the target ARN
func (rtb *RuleTargetBuilder) WithArn(arn string) *RuleTargetBuilder {
	rtb.target.Arn = arn
	return rtb
}

// WithRole sets the IAM role for the target
func (rtb *RuleTargetBuilder) WithRole(roleArn string) *RuleTargetBuilder {
	rtb.target.RoleArn = roleArn
	return rtb
}

// WithInput sets the input for the target
func (rtb *RuleTargetBuilder) WithInput(input string) *RuleTargetBuilder {
	rtb.target.Input = input
	return rtb
}

// Done completes the target and returns to rule builder
func (rtb *RuleTargetBuilder) Done() *EventRuleBuilder {
	rtb.ruleBuilder.config.Targets = append(rtb.ruleBuilder.config.Targets, *rtb.target)
	return rtb.ruleBuilder
}

// EventPublishBuilder builds event publishing operations
type EventPublishBuilder struct {
	ctx    *TestContext
	config *EventPublishConfig
}

type EventPublishConfig struct {
	EventBusName string
	Source       string
	DetailType   string
	Detail       map[string]interface{}
	Resources    []string
	Time         *time.Time
}

// ToEventBus sets the target event bus
func (epb *EventPublishBuilder) ToEventBus(eventBusName string) *EventPublishBuilder {
	epb.config.EventBusName = eventBusName
	return epb
}

// WithSource sets the event source
func (epb *EventPublishBuilder) WithSource(source string) *EventPublishBuilder {
	epb.config.Source = source
	return epb
}

// WithDetailType sets the detail type
func (epb *EventPublishBuilder) WithDetailType(detailType string) *EventPublishBuilder {
	epb.config.DetailType = detailType
	return epb
}

// WithDetail sets the event detail
func (epb *EventPublishBuilder) WithDetail(detail map[string]interface{}) *EventPublishBuilder {
	epb.config.Detail = detail
	return epb
}

// WithResources adds resource ARNs
func (epb *EventPublishBuilder) WithResources(resources ...string) *EventPublishBuilder {
	epb.config.Resources = append(epb.config.Resources, resources...)
	return epb
}

// At sets the event time
func (epb *EventPublishBuilder) At(eventTime time.Time) *EventPublishBuilder {
	epb.config.Time = &eventTime
	return epb
}

// Publish publishes the event
func (epb *EventPublishBuilder) Publish() *ExecutionResult {
	detailJson, _ := json.Marshal(epb.config.Detail)
	
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"Entries": []map[string]interface{}{
				{
					"EventId":    "12345678-1234-1234-1234-123456789012",
					"Source":     epb.config.Source,
					"DetailType": epb.config.DetailType,
					"Detail":     string(detailJson),
				},
			},
			"FailedEntryCount": 0,
		},
		Metadata: map[string]interface{}{
			"event_bus": epb.config.EventBusName,
			"source":    epb.config.Source,
		},
		ctx: epb.ctx,
	}
}

// EventArchiveBuilder builds archive operations
type EventArchiveBuilder struct {
	ctx         *TestContext
	archiveName string
	config      *ArchiveConfig
}

type ArchiveConfig struct {
	EventSourceArn   string
	Description      string
	EventPattern     map[string]interface{}
	RetentionDays    int32
}

// FromEventBus sets the source event bus
func (eab *EventArchiveBuilder) FromEventBus(eventBusArn string) *EventArchiveBuilder {
	eab.config.EventSourceArn = eventBusArn
	return eab
}

// WithDescription sets the archive description
func (eab *EventArchiveBuilder) WithDescription(description string) *EventArchiveBuilder {
	eab.config.Description = description
	return eab
}

// WithEventPattern sets the event pattern for filtering
func (eab *EventArchiveBuilder) WithEventPattern(pattern map[string]interface{}) *EventArchiveBuilder {
	eab.config.EventPattern = pattern
	return eab
}

// WithRetention sets the retention period
func (eab *EventArchiveBuilder) WithRetention(days int32) *EventArchiveBuilder {
	eab.config.RetentionDays = days
	return eab
}

// Create creates the archive
func (eab *EventArchiveBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"ArchiveArn": fmt.Sprintf("arn:aws:events:us-east-1:123456789012:archive/%s", eab.archiveName),
			"State":      "ENABLED",
		},
		ctx: eab.ctx,
	}
}

// Delete deletes the archive
func (eab *EventArchiveBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Archive %s deleted", eab.archiveName),
		ctx:     eab.ctx,
	}
}

// StartReplay starts a replay from the archive
func (eab *EventArchiveBuilder) StartReplay() *ReplayBuilder {
	return &ReplayBuilder{
		ctx:         eab.ctx,
		archiveName: eab.archiveName,
		config:      &ReplayConfig{},
	}
}

// ReplayBuilder builds replay operations
type ReplayBuilder struct {
	ctx         *TestContext
	archiveName string
	config      *ReplayConfig
}

type ReplayConfig struct {
	ReplayName        string
	Destination       string
	EventStartTime    *time.Time
	EventEndTime      *time.Time
	EventSourceArn    string
}

// Named sets the replay name
func (rb *ReplayBuilder) Named(name string) *ReplayBuilder {
	rb.config.ReplayName = name
	return rb
}

// ToDestination sets the replay destination
func (rb *ReplayBuilder) ToDestination(destination string) *ReplayBuilder {
	rb.config.Destination = destination
	return rb
}

// Between sets the time range for replay
func (rb *ReplayBuilder) Between(startTime, endTime time.Time) *ReplayBuilder {
	rb.config.EventStartTime = &startTime
	rb.config.EventEndTime = &endTime
	return rb
}

// Start starts the replay
func (rb *ReplayBuilder) Start() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"ReplayArn": fmt.Sprintf("arn:aws:events:us-east-1:123456789012:replay/%s", rb.config.ReplayName),
			"State":     "STARTING",
		},
		ctx: rb.ctx,
	}
}