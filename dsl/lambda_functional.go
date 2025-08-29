package dsl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// LambdaPayload represents different types of Lambda payloads using ADT
type LambdaPayload struct {
	value mo.Either[string, interface{}] // Either raw string or JSON-serializable data
}

// NewStringPayload creates a string payload
func NewStringPayload(payload string) LambdaPayload {
	return LambdaPayload{value: mo.Left[string, interface{}](payload)}
}

// NewJSONPayload creates a JSON payload from any serializable data
func NewJSONPayload(payload interface{}) LambdaPayload {
	return LambdaPayload{value: mo.Right[string, interface{}](payload)}
}

// ToBytes converts the payload to bytes using functional composition
func (lp LambdaPayload) ToBytes() mo.Result[[]byte] {
	return lp.value.Match(
		func(str string) mo.Result[[]byte] {
			return mo.Ok[[]byte]([]byte(str))
		},
		func(data interface{}) mo.Result[[]byte] {
			return TryWithError(func() ([]byte, error) {
				return json.Marshal(data)
			})
		},
	)
}

// LambdaConfig represents immutable Lambda configuration
type LambdaConfig struct {
	functionName  string
	payload       mo.Option[LambdaPayload]
	qualifier     mo.Option[string]
	invokeType    InvokeType
	logType       LogType
	clientContext mo.Option[string]
	timeout       time.Duration
}

// InvokeType represents Lambda invocation types as an enum
type InvokeType int

const (
	InvokeTypeSync InvokeType = iota
	InvokeTypeAsync
	InvokeTypeDryRun
)

// LogType represents Lambda log types
type LogType int

const (
	LogTypeNone LogType = iota
	LogTypeTail
)

// NewLambdaConfig creates a new immutable Lambda configuration
func NewLambdaConfig(functionName string) LambdaConfig {
	return LambdaConfig{
		functionName:  functionName,
		payload:       mo.None[LambdaPayload](),
		qualifier:     mo.None[string](),
		invokeType:    InvokeTypeSync,
		logType:       LogTypeNone,
		clientContext: mo.None[string](),
		timeout:       30 * time.Second,
	}
}

// LambdaConfigOption represents functional options for Lambda configuration
type LambdaConfigOption func(LambdaConfig) LambdaConfig

// WithPayload sets the Lambda payload
func WithPayload(payload LambdaPayload) LambdaConfigOption {
	return func(config LambdaConfig) LambdaConfig {
		return LambdaConfig{
			functionName:  config.functionName,
			payload:       mo.Some(payload),
			qualifier:     config.qualifier,
			invokeType:    config.invokeType,
			logType:       config.logType,
			clientContext: config.clientContext,
			timeout:       config.timeout,
		}
	}
}

// WithJSONPayload is a convenience function for JSON payloads
func WithJSONPayload(data interface{}) LambdaConfigOption {
	return WithPayload(NewJSONPayload(data))
}

// WithStringPayload is a convenience function for string payloads
func WithStringPayload(payload string) LambdaConfigOption {
	return WithPayload(NewStringPayload(payload))
}

// WithQualifier sets the Lambda qualifier (version or alias)
func WithQualifier(qualifier string) LambdaConfigOption {
	return func(config LambdaConfig) LambdaConfig {
		return LambdaConfig{
			functionName:  config.functionName,
			payload:       config.payload,
			qualifier:     mo.Some(qualifier),
			invokeType:    config.invokeType,
			logType:       config.logType,
			clientContext: config.clientContext,
			timeout:       config.timeout,
		}
	}
}

// WithAsyncInvocation sets the invocation to asynchronous
func WithAsyncInvocation() LambdaConfigOption {
	return func(config LambdaConfig) LambdaConfig {
		return LambdaConfig{
			functionName:  config.functionName,
			payload:       config.payload,
			qualifier:     config.qualifier,
			invokeType:    InvokeTypeAsync,
			logType:       config.logType,
			clientContext: config.clientContext,
			timeout:       config.timeout,
		}
	}
}

// WithSyncInvocation sets the invocation to synchronous
func WithSyncInvocation() LambdaConfigOption {
	return func(config LambdaConfig) LambdaConfig {
		return LambdaConfig{
			functionName:  config.functionName,
			payload:       config.payload,
			qualifier:     config.qualifier,
			invokeType:    InvokeTypeSync,
			logType:       config.logType,
			clientContext: config.clientContext,
			timeout:       config.timeout,
		}
	}
}

// WithDryRun sets the invocation to dry run
func WithDryRun() LambdaConfigOption {
	return func(config LambdaConfig) LambdaConfig {
		return LambdaConfig{
			functionName:  config.functionName,
			payload:       config.payload,
			qualifier:     config.qualifier,
			invokeType:    InvokeTypeDryRun,
			logType:       config.logType,
			clientContext: config.clientContext,
			timeout:       config.timeout,
		}
	}
}

// WithLogs enables log retrieval
func WithLogs() LambdaConfigOption {
	return func(config LambdaConfig) LambdaConfig {
		return LambdaConfig{
			functionName:  config.functionName,
			payload:       config.payload,
			qualifier:     config.qualifier,
			invokeType:    config.invokeType,
			logType:       LogTypeTail,
			clientContext: config.clientContext,
			timeout:       config.timeout,
		}
	}
}

// WithLambdaTimeout sets a custom timeout for Lambda invocation
func WithLambdaTimeout(timeout time.Duration) LambdaConfigOption {
	return func(config LambdaConfig) LambdaConfig {
		return LambdaConfig{
			functionName:  config.functionName,
			payload:       config.payload,
			qualifier:     config.qualifier,
			invokeType:    config.invokeType,
			logType:       config.logType,
			clientContext: config.clientContext,
			timeout:       timeout,
		}
	}
}

// LambdaResponse represents the response from Lambda invocation
type LambdaResponse struct {
	statusCode   int
	payload      mo.Option[string]
	errorMessage mo.Option[string]
	logResult    mo.Option[string]
	executedVersion mo.Option[string]
}

// NewLambdaResponse creates a successful Lambda response
func NewLambdaResponse(statusCode int, payload string) LambdaResponse {
	return LambdaResponse{
		statusCode:      statusCode,
		payload:         mo.Some(payload),
		errorMessage:    mo.None[string](),
		logResult:       mo.None[string](),
		executedVersion: mo.None[string](),
	}
}

// NewLambdaErrorResponse creates an error Lambda response
func NewLambdaErrorResponse(statusCode int, errorMessage string) LambdaResponse {
	return LambdaResponse{
		statusCode:   statusCode,
		payload:      mo.None[string](),
		errorMessage: mo.Some(errorMessage),
		logResult:    mo.None[string](),
		executedVersion: mo.None[string](),
	}
}

// WithLogResult adds log result to the response
func (lr LambdaResponse) WithLogResult(logResult string) LambdaResponse {
	return LambdaResponse{
		statusCode:      lr.statusCode,
		payload:         lr.payload,
		errorMessage:    lr.errorMessage,
		logResult:       mo.Some(logResult),
		executedVersion: lr.executedVersion,
	}
}

// WithExecutedVersion adds executed version to the response
func (lr LambdaResponse) WithExecutedVersion(version string) LambdaResponse {
	return LambdaResponse{
		statusCode:      lr.statusCode,
		payload:         lr.payload,
		errorMessage:    lr.errorMessage,
		logResult:       lr.logResult,
		executedVersion: mo.Some(version),
	}
}

// IsSuccess checks if the Lambda invocation was successful
func (lr LambdaResponse) IsSuccess() bool {
	return lr.statusCode >= 200 && lr.statusCode < 300
}

// GetPayload returns the payload if successful
func (lr LambdaResponse) GetPayload() mo.Option[string] {
	return lr.payload
}

// GetError returns the error message if failed
func (lr LambdaResponse) GetError() mo.Option[string] {
	return lr.errorMessage
}

// LambdaService represents pure functional Lambda operations
type LambdaService struct {
	// In a real implementation, this would contain AWS SDK clients
	// For now, we simulate Lambda operations
}

// NewLambdaService creates a new Lambda service
func NewLambdaService() LambdaService {
	return LambdaService{}
}

// InvokeLambda invokes a Lambda function using pure functional composition
func (ls LambdaService) InvokeLambda(config LambdaConfig, testConfig TestConfig) ExecutionResult[LambdaResponse] {
	return ls.executeInvocation(config, testConfig)
}

// executeInvocation simulates Lambda invocation with functional error handling
func (ls LambdaService) executeInvocation(config LambdaConfig, testConfig TestConfig) ExecutionResult[LambdaResponse] {
	startTime := time.Now()
	
	// Simulate payload processing
	payloadBytes := config.payload.Match(
		func(payload LambdaPayload) mo.Result[[]byte] {
			return payload.ToBytes()
		},
		func() mo.Result[[]byte] {
			return mo.Ok[[]byte]([]byte("{}"))
		},
	)
	
	// Check for payload conversion errors
	if payloadBytes.IsError() {
		return NewErrorResult[LambdaResponse](
			fmt.Errorf("failed to process payload: %w", payloadBytes.Error()),
			time.Since(startTime),
			map[string]interface{}{
				"function_name": config.functionName,
				"error_type":   "payload_error",
			},
			testConfig,
		)
	}
	
	// Simulate Lambda invocation based on invoke type
	response := config.invokeType.Match(
		func() LambdaResponse { // Sync
			return NewLambdaResponse(200, fmt.Sprintf(`{"result": "success", "input_size": %d}`, len(payloadBytes.MustGet())))
		},
		func() LambdaResponse { // Async
			return LambdaResponse{
				statusCode:   202,
				payload:      mo.None[string](),
				errorMessage: mo.None[string](),
				logResult:    mo.None[string](),
				executedVersion: mo.Some("$LATEST"),
			}
		},
		func() LambdaResponse { // DryRun
			return NewLambdaResponse(204, "")
		},
	)
	
	// Add logs if requested
	if config.logType == LogTypeTail {
		response = response.WithLogResult("START RequestId: 123-456-789\nTest log entry\nEND RequestId: 123-456-789")
	}
	
	// Add executed version
	response = response.WithExecutedVersion("$LATEST")
	
	// Create metadata using functional composition
	metadata := lo.Assign(
		map[string]interface{}{
			"function_name": config.functionName,
			"invoke_type":   invokeTypeToString(config.invokeType),
			"payload_size":  len(payloadBytes.MustGet()),
		},
		lo.Ternary(config.qualifier.IsPresent(), 
			map[string]interface{}{"qualifier": config.qualifier.MustGet()}, 
			map[string]interface{}{}),
	)
	
	return NewSuccessResult(
		response,
		time.Since(startTime),
		metadata,
		testConfig,
	)
}

// InvokeLambdaAsync invokes a Lambda function asynchronously
func (ls LambdaService) InvokeLambdaAsync(config LambdaConfig, testConfig TestConfig) AsyncExecutionResult[LambdaResponse] {
	return NewAsyncExecutionResult(func() ExecutionResult[LambdaResponse] {
		// Simulate async processing delay
		time.Sleep(100 * time.Millisecond)
		return ls.executeInvocation(config, testConfig)
	})
}

// GetLambdaConfiguration retrieves Lambda function configuration
func (ls LambdaService) GetLambdaConfiguration(functionName string, testConfig TestConfig) ExecutionResult[LambdaFunctionConfiguration] {
	startTime := time.Now()
	
	config := LambdaFunctionConfiguration{
		functionName: functionName,
		runtime:      "nodejs18.x",
		handler:      "index.handler",
		codeSize:     1024,
		timeout:      30,
		memorySize:   128,
		lastModified: time.Now().Add(-24 * time.Hour),
		version:      "$LATEST",
		environment:  map[string]string{},
	}
	
	return NewSuccessResult(
		config,
		time.Since(startTime),
		map[string]interface{}{
			"function_name": functionName,
			"operation":     "get_configuration",
		},
		testConfig,
	)
}

// LambdaFunctionConfiguration represents Lambda function configuration
type LambdaFunctionConfiguration struct {
	functionName string
	runtime      string
	handler      string
	codeSize     int64
	timeout      int
	memorySize   int
	lastModified time.Time
	version      string
	environment  map[string]string
}

// GetFunctionName returns the function name
func (lfc LambdaFunctionConfiguration) GetFunctionName() string {
	return lfc.functionName
}

// GetRuntime returns the runtime
func (lfc LambdaFunctionConfiguration) GetRuntime() string {
	return lfc.runtime
}

// GetHandler returns the handler
func (lfc LambdaFunctionConfiguration) GetHandler() string {
	return lfc.handler
}

// GetCodeSize returns the code size
func (lfc LambdaFunctionConfiguration) GetCodeSize() int64 {
	return lfc.codeSize
}

// GetTimeout returns the timeout
func (lfc LambdaFunctionConfiguration) GetTimeout() int {
	return lfc.timeout
}

// GetMemorySize returns the memory size
func (lfc LambdaFunctionConfiguration) GetMemorySize() int {
	return lfc.memorySize
}

// GetVersion returns the version
func (lfc LambdaFunctionConfiguration) GetVersion() string {
	return lfc.version
}

// GetEnvironment returns the environment variables
func (lfc LambdaFunctionConfiguration) GetEnvironment() map[string]string {
	return lo.Assign(map[string]string{}, lfc.environment)
}

// Utility functions for InvokeType using pattern matching
func (it InvokeType) Match(syncFn func() LambdaResponse, asyncFn func() LambdaResponse, dryRunFn func() LambdaResponse) LambdaResponse {
	switch it {
	case InvokeTypeSync:
		return syncFn()
	case InvokeTypeAsync:
		return asyncFn()
	case InvokeTypeDryRun:
		return dryRunFn()
	default:
		return syncFn()
	}
}

// invokeTypeToString converts InvokeType to string
func invokeTypeToString(it InvokeType) string {
	switch it {
	case InvokeTypeSync:
		return "RequestResponse"
	case InvokeTypeAsync:
		return "Event"
	case InvokeTypeDryRun:
		return "DryRun"
	default:
		return "RequestResponse"
	}
}

// High-level functional Lambda operations

// CreateLambdaFunction creates a Lambda function using functional composition
func CreateLambdaFunction(name string, opts ...LambdaCreateOption) func(TestConfig) ExecutionResult[LambdaFunctionConfiguration] {
	return func(testConfig TestConfig) ExecutionResult[LambdaFunctionConfiguration] {
		startTime := time.Now()
		
		config := lo.Reduce(opts, func(config LambdaCreateConfig, opt LambdaCreateOption, _ int) LambdaCreateConfig {
			return opt(config)
		}, NewLambdaCreateConfig(name))
		
		// Simulate function creation
		functionConfig := LambdaFunctionConfiguration{
			functionName: config.functionName,
			runtime:      config.runtime.OrElse("nodejs18.x"),
			handler:      config.handler.OrElse("index.handler"),
			codeSize:     1024,
			timeout:      30,
			memorySize:   128,
			lastModified: time.Now(),
			version:      "$LATEST",
			environment:  config.environment,
		}
		
		return NewSuccessResult(
			functionConfig,
			time.Since(startTime),
			map[string]interface{}{
				"function_name": config.functionName,
				"operation":     "create_function",
			},
			testConfig,
		)
	}
}

// LambdaCreateConfig represents configuration for creating Lambda functions
type LambdaCreateConfig struct {
	functionName string
	runtime      mo.Option[string]
	handler      mo.Option[string]
	code         mo.Option[interface{}]
	role         mo.Option[string]
	environment  map[string]string
	timeout      mo.Option[int]
	memorySize   mo.Option[int]
	tags         map[string]string
}

// NewLambdaCreateConfig creates a new Lambda creation configuration
func NewLambdaCreateConfig(functionName string) LambdaCreateConfig {
	return LambdaCreateConfig{
		functionName: functionName,
		runtime:      mo.None[string](),
		handler:      mo.None[string](),
		code:         mo.None[interface{}](),
		role:         mo.None[string](),
		environment:  make(map[string]string),
		timeout:      mo.None[int](),
		memorySize:   mo.None[int](),
		tags:         make(map[string]string),
	}
}

// LambdaCreateOption represents functional options for Lambda creation
type LambdaCreateOption func(LambdaCreateConfig) LambdaCreateConfig

// WithRuntime sets the Lambda runtime
func WithRuntime(runtime string) LambdaCreateOption {
	return func(config LambdaCreateConfig) LambdaCreateConfig {
		return LambdaCreateConfig{
			functionName: config.functionName,
			runtime:      mo.Some(runtime),
			handler:      config.handler,
			code:         config.code,
			role:         config.role,
			environment:  config.environment,
			timeout:      config.timeout,
			memorySize:   config.memorySize,
			tags:         config.tags,
		}
	}
}

// WithHandler sets the Lambda handler
func WithHandler(handler string) LambdaCreateOption {
	return func(config LambdaCreateConfig) LambdaCreateConfig {
		return LambdaCreateConfig{
			functionName: config.functionName,
			runtime:      config.runtime,
			handler:      mo.Some(handler),
			code:         config.code,
			role:         config.role,
			environment:  config.environment,
			timeout:      config.timeout,
			memorySize:   config.memorySize,
			tags:         config.tags,
		}
	}
}

// WithRole sets the Lambda execution role
func WithRole(role string) LambdaCreateOption {
	return func(config LambdaCreateConfig) LambdaCreateConfig {
		return LambdaCreateConfig{
			functionName: config.functionName,
			runtime:      config.runtime,
			handler:      config.handler,
			code:         config.code,
			role:         mo.Some(role),
			environment:  config.environment,
			timeout:      config.timeout,
			memorySize:   config.memorySize,
			tags:         config.tags,
		}
	}
}

// WithEnvironmentVariables sets environment variables
func WithEnvironmentVariables(env map[string]string) LambdaCreateOption {
	return func(config LambdaCreateConfig) LambdaCreateConfig {
		return LambdaCreateConfig{
			functionName: config.functionName,
			runtime:      config.runtime,
			handler:      config.handler,
			code:         config.code,
			role:         config.role,
			environment:  lo.Assign(config.environment, env),
			timeout:      config.timeout,
			memorySize:   config.memorySize,
			tags:         config.tags,
		}
	}
}

// Functional Lambda DSL Entry Points

// InvokeLambdaFunction creates a Lambda invocation operation
func InvokeLambdaFunction(functionName string, opts ...LambdaConfigOption) func(TestConfig) ExecutionResult[LambdaResponse] {
	return func(testConfig TestConfig) ExecutionResult[LambdaResponse] {
		config := lo.Reduce(opts, func(config LambdaConfig, opt LambdaConfigOption, _ int) LambdaConfig {
			return opt(config)
		}, NewLambdaConfig(functionName))
		
		service := NewLambdaService()
		return service.InvokeLambda(config, testConfig)
	}
}

// InvokeLambdaFunctionAsync creates an async Lambda invocation operation
func InvokeLambdaFunctionAsync(functionName string, opts ...LambdaConfigOption) func(TestConfig) AsyncExecutionResult[LambdaResponse] {
	return func(testConfig TestConfig) AsyncExecutionResult[LambdaResponse] {
		config := lo.Reduce(opts, func(config LambdaConfig, opt LambdaConfigOption, _ int) LambdaConfig {
			return opt(config)
		}, NewLambdaConfig(functionName))
		
		service := NewLambdaService()
		return service.InvokeLambdaAsync(config, testConfig)
	}
}

// GetLambdaFunctionConfiguration creates a configuration retrieval operation
func GetLambdaFunctionConfiguration(functionName string) func(TestConfig) ExecutionResult[LambdaFunctionConfiguration] {
	return func(testConfig TestConfig) ExecutionResult[LambdaFunctionConfiguration] {
		service := NewLambdaService()
		return service.GetLambdaConfiguration(functionName, testConfig)
	}
}