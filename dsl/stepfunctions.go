package dsl

import (
	"encoding/json"
	"fmt"
	"time"
)

// StepFunctionsBuilder provides a fluent interface for Step Functions operations
type StepFunctionsBuilder struct {
	ctx *TestContext
}

// StateMachine creates a state machine builder for a specific state machine
func (sfb *StepFunctionsBuilder) StateMachine(name string) *StateMachineBuilder {
	return &StateMachineBuilder{
		ctx:              sfb.ctx,
		stateMachineName: name,
		config:           &StateMachineConfig{},
	}
}

// CreateStateMachine starts building a new state machine
func (sfb *StepFunctionsBuilder) CreateStateMachine() *StateMachineCreateBuilder {
	return &StateMachineCreateBuilder{
		ctx:    sfb.ctx,
		config: &StateMachineCreateConfig{},
	}
}

// Execution creates an execution builder for a specific execution
func (sfb *StepFunctionsBuilder) Execution(executionArn string) *ExecutionBuilder {
	return &ExecutionBuilder{
		ctx:          sfb.ctx,
		executionArn: executionArn,
		config:       &ExecutionConfig{},
	}
}

// Activity creates an activity builder for a specific activity
func (sfb *StepFunctionsBuilder) Activity(name string) *ActivityBuilder {
	return &ActivityBuilder{
		ctx:          sfb.ctx,
		activityName: name,
		config:       &ActivityConfig{},
	}
}

// StateMachineBuilder builds operations on an existing Step Functions state machine
type StateMachineBuilder struct {
	ctx              *TestContext
	stateMachineName string
	config           *StateMachineConfig
}

type StateMachineConfig struct {
	RoleArn    string
	Definition string
	Type       string
	Tags       map[string]string
}

// WithRole sets the IAM role for the state machine
func (smb *StateMachineBuilder) WithRole(roleArn string) *StateMachineBuilder {
	smb.config.RoleArn = roleArn
	return smb
}

// StartExecution starts an execution of the state machine
func (smb *StateMachineBuilder) StartExecution() *ExecutionStartBuilder {
	return &ExecutionStartBuilder{
		ctx:              smb.ctx,
		stateMachineName: smb.stateMachineName,
		config:           &ExecutionStartConfig{},
	}
}

// Update updates the state machine
func (smb *StateMachineBuilder) Update() *StateMachineUpdateBuilder {
	return &StateMachineUpdateBuilder{
		ctx:              smb.ctx,
		stateMachineName: smb.stateMachineName,
		config:           &StateMachineUpdateConfig{},
	}
}

// Delete deletes the state machine
func (smb *StateMachineBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("State machine %s deleted", smb.stateMachineName),
		ctx:     smb.ctx,
	}
}

// Describe describes the state machine
func (smb *StateMachineBuilder) Describe() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"stateMachineArn": fmt.Sprintf("arn:aws:states:us-east-1:123456789012:stateMachine:%s", smb.stateMachineName),
			"name":            smb.stateMachineName,
			"status":          "ACTIVE",
			"type":            "STANDARD",
			"creationDate":    time.Now(),
		},
		ctx: smb.ctx,
	}
}

// ListExecutions lists executions of the state machine
func (smb *StateMachineBuilder) ListExecutions() *ExecutionListBuilder {
	return &ExecutionListBuilder{
		ctx:              smb.ctx,
		stateMachineName: smb.stateMachineName,
		config:           &ExecutionListConfig{},
	}
}

// StateMachineCreateBuilder builds state machine creation operations
type StateMachineCreateBuilder struct {
	ctx    *TestContext
	config *StateMachineCreateConfig
}

type StateMachineCreateConfig struct {
	Name       string
	Definition string
	RoleArn    string
	Type       string
	Tags       map[string]string
}

// Named sets the state machine name
func (smcb *StateMachineCreateBuilder) Named(name string) *StateMachineCreateBuilder {
	smcb.config.Name = name
	return smcb
}

// WithDefinition sets the state machine definition
func (smcb *StateMachineCreateBuilder) WithDefinition(definition string) *StateMachineCreateBuilder {
	smcb.config.Definition = definition
	return smcb
}

// WithJSONDefinition sets the definition from a struct
func (smcb *StateMachineCreateBuilder) WithJSONDefinition(definition interface{}) *StateMachineCreateBuilder {
	if jsonDef, err := json.Marshal(definition); err == nil {
		smcb.config.Definition = string(jsonDef)
	}
	return smcb
}

// WithRole sets the IAM role
func (smcb *StateMachineCreateBuilder) WithRole(roleArn string) *StateMachineCreateBuilder {
	smcb.config.RoleArn = roleArn
	return smcb
}

// AsStandard sets the type to STANDARD
func (smcb *StateMachineCreateBuilder) AsStandard() *StateMachineCreateBuilder {
	smcb.config.Type = "STANDARD"
	return smcb
}

// AsExpress sets the type to EXPRESS
func (smcb *StateMachineCreateBuilder) AsExpress() *StateMachineCreateBuilder {
	smcb.config.Type = "EXPRESS"
	return smcb
}

// WithTags adds tags to the state machine
func (smcb *StateMachineCreateBuilder) WithTags(tags map[string]string) *StateMachineCreateBuilder {
	smcb.config.Tags = tags
	return smcb
}

// Create creates the state machine
func (smcb *StateMachineCreateBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"stateMachineArn": fmt.Sprintf("arn:aws:states:us-east-1:123456789012:stateMachine:%s", smcb.config.Name),
			"creationDate":    time.Now(),
		},
		ctx: smcb.ctx,
	}
}

// ExecutionStartBuilder builds execution start operations
type ExecutionStartBuilder struct {
	ctx              *TestContext
	stateMachineName string
	config           *ExecutionStartConfig
}

type ExecutionStartConfig struct {
	Name            string
	Input           string
	TraceHeader     string
}

// WithName sets the execution name
func (esb *ExecutionStartBuilder) WithName(name string) *ExecutionStartBuilder {
	esb.config.Name = name
	return esb
}

// WithInput sets the execution input
func (esb *ExecutionStartBuilder) WithInput(input string) *ExecutionStartBuilder {
	esb.config.Input = input
	return esb
}

// WithJSONInput sets the input from a struct
func (esb *ExecutionStartBuilder) WithJSONInput(input interface{}) *ExecutionStartBuilder {
	if jsonInput, err := json.Marshal(input); err == nil {
		esb.config.Input = string(jsonInput)
	}
	return esb
}

// Execute starts the execution
func (esb *ExecutionStartBuilder) Execute() *ExecutionResult {
	executionArn := fmt.Sprintf("arn:aws:states:us-east-1:123456789012:execution:%s:%s", esb.stateMachineName, esb.config.Name)
	
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"executionArn": executionArn,
			"startDate":    time.Now(),
		},
		Metadata: map[string]interface{}{
			"state_machine": esb.stateMachineName,
			"execution_name": esb.config.Name,
		},
		ctx: esb.ctx,
	}
}

// ExecutionBuilder builds operations on Step Functions executions
type ExecutionBuilder struct {
	ctx          *TestContext
	executionArn string
	config       *ExecutionConfig
}

type ExecutionConfig struct {
	Timeout time.Duration
}

// WithTimeout sets a timeout for waiting operations
func (eb *ExecutionBuilder) WithTimeout(timeout time.Duration) *ExecutionBuilder {
	eb.config.Timeout = timeout
	return eb
}

// Stop stops the execution
func (eb *ExecutionBuilder) Stop() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"stopDate": time.Now(),
		},
		ctx: eb.ctx,
	}
}

// Describe describes the execution
func (eb *ExecutionBuilder) Describe() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"executionArn": eb.executionArn,
			"status":       "SUCCEEDED",
			"startDate":    time.Now().Add(-5 * time.Minute),
			"stopDate":     time.Now(),
			"input":        "{}",
			"output":       "{\"result\": \"success\"}",
		},
		ctx: eb.ctx,
	}
}

// GetHistory gets the execution history
func (eb *ExecutionBuilder) GetHistory() *ExecutionHistoryBuilder {
	return &ExecutionHistoryBuilder{
		ctx:          eb.ctx,
		executionArn: eb.executionArn,
		config:       &ExecutionHistoryConfig{},
	}
}

// WaitUntilCompleted waits until the execution completes
func (eb *ExecutionBuilder) WaitUntilCompleted() *ExecutionResult {
	// Simulate waiting for completion
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"status":   "SUCCEEDED",
			"stopDate": time.Now(),
			"output":   "{\"result\": \"success\"}",
		},
		Duration: 2 * time.Second, // Simulated execution time
		ctx:      eb.ctx,
	}
}

// ExecutionHistoryBuilder builds execution history operations
type ExecutionHistoryBuilder struct {
	ctx          *TestContext
	executionArn string
	config       *ExecutionHistoryConfig
}

type ExecutionHistoryConfig struct {
	MaxResults      int32
	IncludeDetails  bool
	ReverseOrder    bool
}

// WithMaxResults limits the number of events returned
func (ehb *ExecutionHistoryBuilder) WithMaxResults(maxResults int32) *ExecutionHistoryBuilder {
	ehb.config.MaxResults = maxResults
	return ehb
}

// WithDetails includes event details
func (ehb *ExecutionHistoryBuilder) WithDetails() *ExecutionHistoryBuilder {
	ehb.config.IncludeDetails = true
	return ehb
}

// InReverseOrder returns events in reverse chronological order
func (ehb *ExecutionHistoryBuilder) InReverseOrder() *ExecutionHistoryBuilder {
	ehb.config.ReverseOrder = true
	return ehb
}

// Fetch retrieves the execution history
func (ehb *ExecutionHistoryBuilder) Fetch() *ExecutionResult {
	events := []map[string]interface{}{
		{
			"timestamp":                 time.Now().Add(-2 * time.Minute),
			"type":                      "ExecutionStarted",
			"id":                        1,
			"executionStartedEventDetails": map[string]interface{}{
				"input": "{}",
			},
		},
		{
			"timestamp": time.Now(),
			"type":      "ExecutionSucceeded",
			"id":        2,
			"executionSucceededEventDetails": map[string]interface{}{
				"output": "{\"result\": \"success\"}",
			},
		},
	}
	
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"events": events,
		},
		ctx: ehb.ctx,
	}
}

// ExecutionListBuilder builds execution listing operations
type ExecutionListBuilder struct {
	ctx              *TestContext
	stateMachineName string
	config           *ExecutionListConfig
}

type ExecutionListConfig struct {
	StatusFilter string
	MaxResults   int32
}

// WithStatus filters by execution status
func (elb *ExecutionListBuilder) WithStatus(status string) *ExecutionListBuilder {
	elb.config.StatusFilter = status
	return elb
}

// Limit limits the number of results
func (elb *ExecutionListBuilder) Limit(maxResults int32) *ExecutionListBuilder {
	elb.config.MaxResults = maxResults
	return elb
}

// List retrieves the execution list
func (elb *ExecutionListBuilder) List() *ExecutionResult {
	executions := []map[string]interface{}{
		{
			"executionArn": fmt.Sprintf("arn:aws:states:us-east-1:123456789012:execution:%s:exec-1", elb.stateMachineName),
			"name":         "exec-1",
			"status":       "SUCCEEDED",
			"startDate":    time.Now().Add(-time.Hour),
			"stopDate":     time.Now().Add(-30 * time.Minute),
		},
		{
			"executionArn": fmt.Sprintf("arn:aws:states:us-east-1:123456789012:execution:%s:exec-2", elb.stateMachineName),
			"name":         "exec-2",
			"status":       "RUNNING",
			"startDate":    time.Now().Add(-10 * time.Minute),
		},
	}
	
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"executions": executions,
		},
		ctx: elb.ctx,
	}
}

// StateMachineUpdateBuilder builds state machine update operations
type StateMachineUpdateBuilder struct {
	ctx              *TestContext
	stateMachineName string
	config           *StateMachineUpdateConfig
}

type StateMachineUpdateConfig struct {
	Definition string
	RoleArn    string
}

// WithDefinition updates the definition
func (smub *StateMachineUpdateBuilder) WithDefinition(definition string) *StateMachineUpdateBuilder {
	smub.config.Definition = definition
	return smub
}

// WithRole updates the IAM role
func (smub *StateMachineUpdateBuilder) WithRole(roleArn string) *StateMachineUpdateBuilder {
	smub.config.RoleArn = roleArn
	return smub
}

// Apply applies the updates
func (smub *StateMachineUpdateBuilder) Apply() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"updateDate": time.Now(),
		},
		ctx: smub.ctx,
	}
}

// ActivityBuilder builds operations on Step Functions activities
type ActivityBuilder struct {
	ctx          *TestContext
	activityName string
	config       *ActivityConfig
}

type ActivityConfig struct {
	Tags map[string]string
}

// Create creates the activity
func (ab *ActivityBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"activityArn":  fmt.Sprintf("arn:aws:states:us-east-1:123456789012:activity:%s", ab.activityName),
			"creationDate": time.Now(),
		},
		ctx: ab.ctx,
	}
}

// Delete deletes the activity
func (ab *ActivityBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Activity %s deleted", ab.activityName),
		ctx:     ab.ctx,
	}
}

// GetTask polls for a task from the activity
func (ab *ActivityBuilder) GetTask() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"taskToken": "AAAAKgAAAAIAAAAAAAAAAVnYLJtdT9y-s2-6ND3OVNkuVZLvpOAiDkQF6vzkBCwmhRJu",
			"input":     "{}",
		},
		ctx: ab.ctx,
	}
}