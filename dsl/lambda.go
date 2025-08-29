package dsl

import (
	"encoding/json"
	"fmt"
	"time"
)

// LambdaBuilder provides a fluent interface for Lambda operations
type LambdaBuilder struct {
	ctx *TestContext
}

// Function creates a Lambda function builder for a specific function
func (lb *LambdaBuilder) Function(name string) *LambdaFunctionBuilder {
	return &LambdaFunctionBuilder{
		ctx:          lb.ctx,
		functionName: name,
		config:       &LambdaFunctionConfig{},
	}
}

// CreateFunction starts building a new Lambda function
func (lb *LambdaBuilder) CreateFunction() *LambdaFunctionCreateBuilder {
	return &LambdaFunctionCreateBuilder{
		ctx:    lb.ctx,
		config: &LambdaCreateConfig{},
	}
}

// ListFunctions returns a builder to list Lambda functions
func (lb *LambdaBuilder) ListFunctions() *LambdaListBuilder {
	return &LambdaListBuilder{ctx: lb.ctx}
}

// LambdaFunctionBuilder builds operations on an existing Lambda function
type LambdaFunctionBuilder struct {
	ctx          *TestContext
	functionName string
	config       *LambdaFunctionConfig
}

// LambdaFunctionConfig holds configuration for Lambda function operations
type LambdaFunctionConfig struct {
	Payload      interface{}
	Qualifier    string
	InvokeType   string
	LogType      string
	ClientContext string
	Timeout      time.Duration
}

// WithPayload sets the payload for function invocation
func (lfb *LambdaFunctionBuilder) WithPayload(payload interface{}) *LambdaFunctionBuilder {
	lfb.config.Payload = payload
	return lfb
}

// WithJSONPayload sets a JSON payload from a struct
func (lfb *LambdaFunctionBuilder) WithJSONPayload(data interface{}) *LambdaFunctionBuilder {
	lfb.config.Payload = data
	return lfb
}

// WithQualifier sets the function version or alias
func (lfb *LambdaFunctionBuilder) WithQualifier(qualifier string) *LambdaFunctionBuilder {
	lfb.config.Qualifier = qualifier
	return lfb
}

// Async sets the invocation to be asynchronous
func (lfb *LambdaFunctionBuilder) Async() *LambdaFunctionBuilder {
	lfb.config.InvokeType = "Event"
	return lfb
}

// Sync sets the invocation to be synchronous (default)
func (lfb *LambdaFunctionBuilder) Sync() *LambdaFunctionBuilder {
	lfb.config.InvokeType = "RequestResponse"
	return lfb
}

// DryRun performs a dry run validation
func (lfb *LambdaFunctionBuilder) DryRun() *LambdaFunctionBuilder {
	lfb.config.InvokeType = "DryRun"
	return lfb
}

// WithLogs includes execution logs in the response
func (lfb *LambdaFunctionBuilder) WithLogs() *LambdaFunctionBuilder {
	lfb.config.LogType = "Tail"
	return lfb
}

// WithTimeout sets a custom timeout for the invocation
func (lfb *LambdaFunctionBuilder) WithTimeout(timeout time.Duration) *LambdaFunctionBuilder {
	lfb.config.Timeout = timeout
	return lfb
}

// Invoke executes the Lambda function with configured parameters
func (lfb *LambdaFunctionBuilder) Invoke() *ExecutionResult {
	startTime := time.Now()
	
	// Convert payload to JSON if needed
	var payloadBytes []byte
	var err error
	
	if lfb.config.Payload != nil {
		switch p := lfb.config.Payload.(type) {
		case string:
			payloadBytes = []byte(p)
		case []byte:
			payloadBytes = p
		default:
			payloadBytes, err = json.Marshal(p)
			if err != nil {
				return &ExecutionResult{
					Success:  false,
					Error:    fmt.Errorf("failed to marshal payload: %w", err),
					Duration: time.Since(startTime),
					ctx:      lfb.ctx,
				}
			}
		}
	}
	
	// Here we would call the actual AWS Lambda service
	// For now, we'll simulate the operation
	result := &ExecutionResult{
		Success:  true,
		Data:     string(payloadBytes), // Simulated response
		Duration: time.Since(startTime),
		Metadata: map[string]interface{}{
			"function_name": lfb.functionName,
			"invoke_type":   lfb.config.InvokeType,
			"payload_size":  len(payloadBytes),
		},
		ctx: lfb.ctx,
	}
	
	return result
}

// InvokeAsync performs an asynchronous invocation
func (lfb *LambdaFunctionBuilder) InvokeAsync() *AsyncExecutionResult {
	completed := make(chan bool, 1)
	
	// Simulate async execution
	go func() {
		time.Sleep(100 * time.Millisecond) // Simulate processing
		completed <- true
	}()
	
	return &AsyncExecutionResult{
		ExecutionResult: &ExecutionResult{
			Success:  true,
			Duration: 0, // Will be set when completed
			ctx:      lfb.ctx,
		},
		completed: completed,
	}
}

// GetConfiguration retrieves the function configuration
func (lfb *LambdaFunctionBuilder) GetConfiguration() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"FunctionName": lfb.functionName,
			"Runtime":      "nodejs18.x",
			"Handler":      "index.handler",
			"CodeSize":     1024,
			"Timeout":      30,
		},
		ctx: lfb.ctx,
	}
}

// UpdateConfiguration updates the function configuration
func (lfb *LambdaFunctionBuilder) UpdateConfiguration() *LambdaUpdateConfigBuilder {
	return &LambdaUpdateConfigBuilder{
		ctx:          lfb.ctx,
		functionName: lfb.functionName,
		updates:      make(map[string]interface{}),
	}
}

// Delete deletes the Lambda function
func (lfb *LambdaFunctionBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Function %s deleted", lfb.functionName),
		ctx:     lfb.ctx,
	}
}

// GetLogs retrieves CloudWatch logs for the function
func (lfb *LambdaFunctionBuilder) GetLogs() *LambdaLogsBuilder {
	return &LambdaLogsBuilder{
		ctx:          lfb.ctx,
		functionName: lfb.functionName,
		config:       &LogsConfig{},
	}
}

// Exists checks if the function exists
func (lfb *LambdaFunctionBuilder) Exists() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    true, // Simulated existence check
		ctx:     lfb.ctx,
	}
}

// LambdaUpdateConfigBuilder builds function configuration updates
type LambdaUpdateConfigBuilder struct {
	ctx          *TestContext
	functionName string
	updates      map[string]interface{}
}

// WithTimeout updates the function timeout
func (lucb *LambdaUpdateConfigBuilder) WithTimeout(timeout int) *LambdaUpdateConfigBuilder {
	lucb.updates["Timeout"] = timeout
	return lucb
}

// WithMemory updates the function memory
func (lucb *LambdaUpdateConfigBuilder) WithMemory(memory int) *LambdaUpdateConfigBuilder {
	lucb.updates["MemorySize"] = memory
	return lucb
}

// WithEnvironment updates environment variables
func (lucb *LambdaUpdateConfigBuilder) WithEnvironment(vars map[string]string) *LambdaUpdateConfigBuilder {
	lucb.updates["Environment"] = vars
	return lucb
}

// Apply applies the configuration updates
func (lucb *LambdaUpdateConfigBuilder) Apply() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    lucb.updates,
		ctx:     lucb.ctx,
	}
}

// LambdaLogsBuilder builds log retrieval operations
type LambdaLogsBuilder struct {
	ctx          *TestContext
	functionName string
	config       *LogsConfig
}

type LogsConfig struct {
	StartTime     *time.Time
	EndTime       *time.Time
	FilterPattern string
	MaxLines      int
}

// Since sets the start time for log retrieval
func (llb *LambdaLogsBuilder) Since(duration time.Duration) *LambdaLogsBuilder {
	startTime := time.Now().Add(-duration)
	llb.config.StartTime = &startTime
	return llb
}

// Until sets the end time for log retrieval
func (llb *LambdaLogsBuilder) Until(endTime time.Time) *LambdaLogsBuilder {
	llb.config.EndTime = &endTime
	return llb
}

// WithFilter sets a filter pattern for logs
func (llb *LambdaLogsBuilder) WithFilter(pattern string) *LambdaLogsBuilder {
	llb.config.FilterPattern = pattern
	return llb
}

// Limit sets the maximum number of log lines
func (llb *LambdaLogsBuilder) Limit(maxLines int) *LambdaLogsBuilder {
	llb.config.MaxLines = maxLines
	return llb
}

// Fetch retrieves the logs
func (llb *LambdaLogsBuilder) Fetch() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: []string{
			"START RequestId: 123-456-789 Version: $LATEST",
			"Test log entry",
			"END RequestId: 123-456-789",
		},
		ctx: llb.ctx,
	}
}

// LambdaFunctionCreateBuilder builds new Lambda function creation
type LambdaFunctionCreateBuilder struct {
	ctx    *TestContext
	config *LambdaCreateConfig
}

type LambdaCreateConfig struct {
	FunctionName string
	Runtime      string
	Handler      string
	Code         interface{}
	Role         string
	Environment  map[string]string
	Timeout      int
	MemorySize   int
	Tags         map[string]string
}

// Named sets the function name
func (lfcb *LambdaFunctionCreateBuilder) Named(name string) *LambdaFunctionCreateBuilder {
	lfcb.config.FunctionName = name
	return lfcb
}

// WithRuntime sets the runtime
func (lfcb *LambdaFunctionCreateBuilder) WithRuntime(runtime string) *LambdaFunctionCreateBuilder {
	lfcb.config.Runtime = runtime
	return lfcb
}

// WithHandler sets the handler
func (lfcb *LambdaFunctionCreateBuilder) WithHandler(handler string) *LambdaFunctionCreateBuilder {
	lfcb.config.Handler = handler
	return lfcb
}

// WithZipFile sets the code from a zip file
func (lfcb *LambdaFunctionCreateBuilder) WithZipFile(zipData []byte) *LambdaFunctionCreateBuilder {
	lfcb.config.Code = zipData
	return lfcb
}

// WithS3Code sets the code from S3
func (lfcb *LambdaFunctionCreateBuilder) WithS3Code(bucket, key string) *LambdaFunctionCreateBuilder {
	lfcb.config.Code = map[string]string{
		"S3Bucket": bucket,
		"S3Key":    key,
	}
	return lfcb
}

// WithRole sets the execution role
func (lfcb *LambdaFunctionCreateBuilder) WithRole(roleArn string) *LambdaFunctionCreateBuilder {
	lfcb.config.Role = roleArn
	return lfcb
}

// Create creates the Lambda function
func (lfcb *LambdaFunctionCreateBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"FunctionName": lfcb.config.FunctionName,
			"FunctionArn":  fmt.Sprintf("arn:aws:lambda:us-east-1:123456789012:function:%s", lfcb.config.FunctionName),
		},
		ctx: lfcb.ctx,
	}
}

// LambdaListBuilder builds function listing operations
type LambdaListBuilder struct {
	ctx    *TestContext
	filter map[string]interface{}
}

// WithTag filters functions by tag
func (llb *LambdaListBuilder) WithTag(key, value string) *LambdaListBuilder {
	if llb.filter == nil {
		llb.filter = make(map[string]interface{})
	}
	llb.filter[key] = value
	return llb
}

// List retrieves the function list
func (llb *LambdaListBuilder) List() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: []map[string]interface{}{
			{
				"FunctionName": "function-1",
				"Runtime":      "nodejs18.x",
			},
			{
				"FunctionName": "function-2",
				"Runtime":      "python3.9",
			},
		},
		ctx: llb.ctx,
	}
}