package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// FunctionalLambdaConfig represents immutable Lambda configuration using functional options
type FunctionalLambdaConfig struct {
	functionName      string
	invocationType   types.InvocationType
	logType          types.LogType
	timeout          time.Duration
	maxRetries       int
	retryDelay       time.Duration
	clientContext    string
	qualifier        string
	payloadValidator mo.Option[func(string) error]
	metadata         map[string]interface{}
}

// LambdaConfigOption represents a functional option for FunctionalLambdaConfig
type LambdaConfigOption func(FunctionalLambdaConfig) FunctionalLambdaConfig

// NewFunctionalLambdaConfig creates a new Lambda config with functional options
func NewFunctionalLambdaConfig(functionName string, opts ...LambdaConfigOption) FunctionalLambdaConfig {
	base := FunctionalLambdaConfig{
		functionName:     functionName,
		invocationType:   types.InvocationTypeRequestResponse,
		logType:          types.LogTypeNone,
		timeout:          DefaultTimeout,
		maxRetries:       DefaultRetryAttempts,
		retryDelay:       DefaultRetryDelay,
		clientContext:    "",
		qualifier:        "",
		payloadValidator: mo.None[func(string) error](),
		metadata:         make(map[string]interface{}),
	}

	return lo.Reduce(opts, func(config FunctionalLambdaConfig, opt LambdaConfigOption, _ int) FunctionalLambdaConfig {
		return opt(config)
	}, base)
}

// WithInvocationType sets the invocation type functionally
func WithInvocationType(invocationType types.InvocationType) LambdaConfigOption {
	return func(config FunctionalLambdaConfig) FunctionalLambdaConfig {
		return FunctionalLambdaConfig{
			functionName:     config.functionName,
			invocationType:   invocationType,
			logType:          config.logType,
			timeout:          config.timeout,
			maxRetries:       config.maxRetries,
			retryDelay:       config.retryDelay,
			clientContext:    config.clientContext,
			qualifier:        config.qualifier,
			payloadValidator: config.payloadValidator,
			metadata:         config.metadata,
		}
	}
}

// WithLogType sets the log type functionally
func WithLogType(logType types.LogType) LambdaConfigOption {
	return func(config FunctionalLambdaConfig) FunctionalLambdaConfig {
		return FunctionalLambdaConfig{
			functionName:     config.functionName,
			invocationType:   config.invocationType,
			logType:          logType,
			timeout:          config.timeout,
			maxRetries:       config.maxRetries,
			retryDelay:       config.retryDelay,
			clientContext:    config.clientContext,
			qualifier:        config.qualifier,
			payloadValidator: config.payloadValidator,
			metadata:         config.metadata,
		}
	}
}

// WithTimeout sets the timeout functionally
func WithTimeout(timeout time.Duration) LambdaConfigOption {
	return func(config FunctionalLambdaConfig) FunctionalLambdaConfig {
		return FunctionalLambdaConfig{
			functionName:     config.functionName,
			invocationType:   config.invocationType,
			logType:          config.logType,
			timeout:          timeout,
			maxRetries:       config.maxRetries,
			retryDelay:       config.retryDelay,
			clientContext:    config.clientContext,
			qualifier:        config.qualifier,
			payloadValidator: config.payloadValidator,
			metadata:         config.metadata,
		}
	}
}

// WithRetryConfig sets retry configuration functionally
func WithRetryConfig(maxRetries int, retryDelay time.Duration) LambdaConfigOption {
	return func(config FunctionalLambdaConfig) FunctionalLambdaConfig {
		return FunctionalLambdaConfig{
			functionName:     config.functionName,
			invocationType:   config.invocationType,
			logType:          config.logType,
			timeout:          config.timeout,
			maxRetries:       maxRetries,
			retryDelay:       retryDelay,
			clientContext:    config.clientContext,
			qualifier:        config.qualifier,
			payloadValidator: config.payloadValidator,
			metadata:         config.metadata,
		}
	}
}

// WithQualifier sets the function qualifier functionally
func WithQualifier(qualifier string) LambdaConfigOption {
	return func(config FunctionalLambdaConfig) FunctionalLambdaConfig {
		return FunctionalLambdaConfig{
			functionName:     config.functionName,
			invocationType:   config.invocationType,
			logType:          config.logType,
			timeout:          config.timeout,
			maxRetries:       config.maxRetries,
			retryDelay:       config.retryDelay,
			clientContext:    config.clientContext,
			qualifier:        qualifier,
			payloadValidator: config.payloadValidator,
			metadata:         config.metadata,
		}
	}
}

// WithPayloadValidator sets payload validation functionally
func WithPayloadValidator(validator func(string) error) LambdaConfigOption {
	return func(config FunctionalLambdaConfig) FunctionalLambdaConfig {
		return FunctionalLambdaConfig{
			functionName:     config.functionName,
			invocationType:   config.invocationType,
			logType:          config.logType,
			timeout:          config.timeout,
			maxRetries:       config.maxRetries,
			retryDelay:       config.retryDelay,
			clientContext:    config.clientContext,
			qualifier:        config.qualifier,
			payloadValidator: mo.Some(validator),
			metadata:         config.metadata,
		}
	}
}

// WithLambdaMetadata adds metadata functionally
func WithLambdaMetadata(key string, value interface{}) LambdaConfigOption {
	return func(config FunctionalLambdaConfig) FunctionalLambdaConfig {
		newMetadata := lo.Assign(config.metadata, map[string]interface{}{key: value})
		return FunctionalLambdaConfig{
			functionName:     config.functionName,
			invocationType:   config.invocationType,
			logType:          config.logType,
			timeout:          config.timeout,
			maxRetries:       config.maxRetries,
			retryDelay:       config.retryDelay,
			clientContext:    config.clientContext,
			qualifier:        config.qualifier,
			payloadValidator: config.payloadValidator,
			metadata:         newMetadata,
		}
	}
}

// GetMetadata returns the metadata for the FunctionalLambdaConfig
func (flc FunctionalLambdaConfig) GetMetadata() map[string]interface{} {
	return flc.metadata
}

// FunctionalInvokeResult represents Lambda invocation result with functional programming principles
type FunctionalInvokeResult struct {
	result   mo.Option[InvokeResult]
	error    mo.Option[error]
	duration time.Duration
	metadata map[string]interface{}
}

// NewSuccessFunctionalInvokeResult creates a successful invoke result
func NewSuccessFunctionalInvokeResult(result InvokeResult, duration time.Duration, metadata map[string]interface{}) FunctionalInvokeResult {
	return FunctionalInvokeResult{
		result:   mo.Some(result),
		error:    mo.None[error](),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
	}
}

// NewErrorFunctionalInvokeResult creates an error invoke result
func NewErrorFunctionalInvokeResult(err error, duration time.Duration, metadata map[string]interface{}) FunctionalInvokeResult {
	return FunctionalInvokeResult{
		result:   mo.None[InvokeResult](),
		error:    mo.Some(err),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
	}
}

// IsSuccess returns true if the invocation was successful
func (fir FunctionalInvokeResult) IsSuccess() bool {
	return fir.result.IsPresent() && fir.error.IsAbsent()
}

// IsError returns true if the invocation has an error
func (fir FunctionalInvokeResult) IsError() bool {
	return fir.error.IsPresent()
}

// GetResult returns the invoke result if present
func (fir FunctionalInvokeResult) GetResult() mo.Option[InvokeResult] {
	return fir.result
}

// GetError returns the error if present
func (fir FunctionalInvokeResult) GetError() mo.Option[error] {
	return fir.error
}

// GetDuration returns the execution duration
func (fir FunctionalInvokeResult) GetDuration() time.Duration {
	return fir.duration
}

// GetMetadata returns the metadata
func (fir FunctionalInvokeResult) GetMetadata() map[string]interface{} {
	return fir.metadata
}

// MapResult applies a function to the result if present
func (fir FunctionalInvokeResult) MapResult(f func(InvokeResult) InvokeResult) FunctionalInvokeResult {
	if fir.IsSuccess() {
		newResult := fir.result.Map(func(result InvokeResult) (InvokeResult, bool) {
			return f(result), true
		})
		return FunctionalInvokeResult{
			result:   newResult,
			error:    fir.error,
			duration: fir.duration,
			metadata: fir.metadata,
		}
	}
	return fir
}

// FunctionalInvokeLambda invokes a Lambda function using functional programming principles
func FunctionalInvokeLambda(ctx *TestContext, payload string, config FunctionalLambdaConfig) FunctionalInvokeResult {
	startTime := time.Now()
	
	// Validate inputs using functional composition
	validationResult := ExecuteFunctionalPipeline[string](payload,
		func(p string) (string, error) {
			if err := validateFunctionName(config.functionName); err != nil {
				return p, err
			}
			return p, nil
		},
		func(p string) (string, error) {
			if err := validatePayload(p); err != nil {
				return p, err
			}
			return p, nil
		},
		func(p string) (string, error) {
			if config.payloadValidator.IsPresent() {
				validator := config.payloadValidator.MustGet()
				if err := validator(p); err != nil {
					return p, err
				}
			}
			return p, nil
		},
	)
	
	if validationResult.IsError() {
		return NewErrorFunctionalInvokeResult(
			validationResult.GetError().MustGet(),
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"function_name": config.functionName,
				"phase":         "validation",
			}),
		)
	}

	// Create Lambda client
	client := createLambdaClient(ctx)
	
	// Build invoke input using functional composition
	invokeInput := &lambda.InvokeInput{
		FunctionName:   aws.String(config.functionName),
		InvocationType: config.invocationType,
		LogType:        config.logType,
		Payload:        []byte(payload),
	}
	
	// Add optional parameters functionally
	if config.qualifier != "" {
		invokeInput.Qualifier = aws.String(config.qualifier)
	}
	if config.clientContext != "" {
		invokeInput.ClientContext = aws.String(config.clientContext)
	}

	// Execute invocation with retry logic
	retryResult := FunctionalRetryWithBackoff(
		func() (InvokeResult, error) {
			response, err := client.Invoke(context.Background(), invokeInput)
			if err != nil {
				return InvokeResult{}, err
			}
			
			// Convert response to InvokeResult
			result := InvokeResult{
				StatusCode:      response.StatusCode,
				Payload:         string(response.Payload),
				LogResult:       aws.ToString(response.LogResult),
				ExecutedVersion: aws.ToString(response.ExecutedVersion),
				FunctionError:   aws.ToString(response.FunctionError),
				ExecutionTime:   time.Since(startTime),
			}
			
			return result, nil
		},
		config.maxRetries,
		config.retryDelay,
	)
	
	totalDuration := time.Since(startTime)
	
	if retryResult.IsError() {
		return NewErrorFunctionalInvokeResult(
			retryResult.GetError().MustGet(),
			totalDuration,
			lo.Assign(config.metadata, map[string]interface{}{
				"function_name": config.functionName,
				"phase":         "execution",
				"retries":       config.maxRetries,
			}),
		)
	}
	
	return NewSuccessFunctionalInvokeResult(
		retryResult.GetResult().MustGet(),
		totalDuration,
		lo.Assign(config.metadata, map[string]interface{}{
			"function_name": config.functionName,
			"phase":         "success",
		}),
	)
}

// FunctionalPipeline represents a validation pipeline
type FunctionalPipeline[T any] struct {
	value T
	error mo.Option[error]
}

// ExecuteFunctionalPipeline executes a series of validation functions
func ExecuteFunctionalPipeline[T any](input T, operations ...func(T) (T, error)) FunctionalPipeline[T] {
	return lo.Reduce(operations, func(acc FunctionalPipeline[T], op func(T) (T, error), _ int) FunctionalPipeline[T] {
		if acc.error.IsPresent() {
			return acc
		}
		
		result, err := op(acc.value)
		if err != nil {
			return FunctionalPipeline[T]{
				value: acc.value,
				error: mo.Some(err),
			}
		}
		
		return FunctionalPipeline[T]{
			value: result,
			error: mo.None[error](),
		}
	}, FunctionalPipeline[T]{
		value: input,
		error: mo.None[error](),
	})
}

// IsError returns true if the pipeline has an error
func (fp FunctionalPipeline[T]) IsError() bool {
	return fp.error.IsPresent()
}

// GetError returns the error if present
func (fp FunctionalPipeline[T]) GetError() mo.Option[error] {
	return fp.error
}

// GetValue returns the value
func (fp FunctionalPipeline[T]) GetValue() T {
	return fp.value
}

// FunctionalRetryResult represents the result of a retry operation
type FunctionalRetryResult[T any] struct {
	result mo.Option[T]
	error  mo.Option[error]
}

// IsError returns true if the retry failed
func (frr FunctionalRetryResult[T]) IsError() bool {
	return frr.error.IsPresent()
}

// GetResult returns the result if present
func (frr FunctionalRetryResult[T]) GetResult() mo.Option[T] {
	return frr.result
}

// GetError returns the error if present
func (frr FunctionalRetryResult[T]) GetError() mo.Option[error] {
	return frr.error
}

// FunctionalRetryWithBackoff retries an operation with exponential backoff
func FunctionalRetryWithBackoff[T any](operation func() (T, error), maxRetries int, delay time.Duration) FunctionalRetryResult[T] {
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := operation()
		if err == nil {
			return FunctionalRetryResult[T]{
				result: mo.Some(result),
				error:  mo.None[error](),
			}
		}
		
		// Wait before retry (except on last attempt)
		if attempt < maxRetries-1 {
			backoffDelay := time.Duration(attempt+1) * delay
			time.Sleep(backoffDelay)
		}
	}
	
	// Final attempt to capture the last error
	_, err := operation()
	return FunctionalRetryResult[T]{
		result: mo.None[T](),
		error:  mo.Some(err),
	}
}