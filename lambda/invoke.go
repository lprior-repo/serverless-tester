package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// Invoke synchronously invokes a Lambda function with the given payload.
// This is the non-error returning version that follows Terratest patterns.
func Invoke(ctx *TestContext, functionName string, payload string) *InvokeResult {
	result, err := InvokeE(ctx, functionName, payload)
	if err != nil {
		ctx.T.Errorf("Lambda invocation failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// InvokeE synchronously invokes a Lambda function with the given payload.
// This is the error returning version that follows Terratest patterns.
func InvokeE(ctx *TestContext, functionName string, payload string) (*InvokeResult, error) {
	return InvokeWithOptionsE(ctx, functionName, payload, nil)
}

// InvokeWithOptions synchronously invokes a Lambda function with custom options.
// This is the non-error returning version that follows Terratest patterns.
func InvokeWithOptions(ctx *TestContext, functionName string, payload string, opts *InvokeOptions) *InvokeResult {
	result, err := InvokeWithOptionsE(ctx, functionName, payload, opts)
	if err != nil {
		ctx.T.Errorf("Lambda invocation with options failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// InvokeWithOptionsE synchronously invokes a Lambda function with custom options.
// This is the error returning version that follows Terratest patterns.
func InvokeWithOptionsE(ctx *TestContext, functionName string, payload string, opts *InvokeOptions) (*InvokeResult, error) {
	startTime := time.Now()
	
	// Validate inputs
	if err := validateFunctionName(functionName); err != nil {
		return nil, err
	}
	
	if err := validatePayload(payload); err != nil {
		return nil, err
	}
	
	// Merge options with defaults
	options := mergeInvokeOptions(opts)
	
	// Validate payload with custom validator if provided
	if err := options.PayloadValidator(payload); err != nil {
		return nil, err
	}
	
	// Log operation start
	logOperation("invoke", functionName, map[string]interface{}{
		"payload_size":    len(payload),
		"invocation_type": string(options.InvocationType),
		"log_type":        string(options.LogType),
	})
	
	// Create Lambda client
	client := createLambdaClient(ctx)
	
	// Execute invocation with retry logic
	result, err := executeInvokeWithRetry(client, functionName, payload, options)
	
	duration := time.Since(startTime)
	logResult("invoke", functionName, err == nil, duration, err)
	
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvocationFailed, err)
	}
	
	result.ExecutionTime = duration
	return result, nil
}

// InvokeAsync asynchronously invokes a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func InvokeAsync(ctx *TestContext, functionName string, payload string) *InvokeResult {
	result, err := InvokeAsyncE(ctx, functionName, payload)
	if err != nil {
		ctx.T.Errorf("Lambda async invocation failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// InvokeAsyncE asynchronously invokes a Lambda function.
// This is the error returning version that follows Terratest patterns.
func InvokeAsyncE(ctx *TestContext, functionName string, payload string) (*InvokeResult, error) {
	opts := &InvokeOptions{
		InvocationType: types.InvocationTypeEvent,
		LogType:        LogTypeNone,
	}
	return InvokeWithOptionsE(ctx, functionName, payload, opts)
}

// InvokeWithRetry invokes a Lambda function with built-in retry logic for eventual consistency.
// This is the non-error returning version that follows Terratest patterns.
func InvokeWithRetry(ctx *TestContext, functionName string, payload string, maxRetries int) *InvokeResult {
	result, err := InvokeWithRetryE(ctx, functionName, payload, maxRetries)
	if err != nil {
		ctx.T.Errorf("Lambda invocation with retry failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// InvokeWithRetryE invokes a Lambda function with built-in retry logic for eventual consistency.
// This is the error returning version that follows Terratest patterns.
func InvokeWithRetryE(ctx *TestContext, functionName string, payload string, maxRetries int) (*InvokeResult, error) {
	opts := &InvokeOptions{
		MaxRetries: maxRetries,
	}
	return InvokeWithOptionsE(ctx, functionName, payload, opts)
}

// DryRunInvoke performs a dry run invocation to validate the function and payload.
// This is the non-error returning version that follows Terratest patterns.
func DryRunInvoke(ctx *TestContext, functionName string, payload string) *InvokeResult {
	result, err := DryRunInvokeE(ctx, functionName, payload)
	if err != nil {
		ctx.T.Errorf("Lambda dry run invocation failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// DryRunInvokeE performs a dry run invocation to validate the function and payload.
// This is the error returning version that follows Terratest patterns.
func DryRunInvokeE(ctx *TestContext, functionName string, payload string) (*InvokeResult, error) {
	opts := &InvokeOptions{
		InvocationType: types.InvocationTypeDryRun,
	}
	return InvokeWithOptionsE(ctx, functionName, payload, opts)
}

// executeInvokeWithRetry handles the actual invocation with retry logic
func executeInvokeWithRetry(client *lambda.Client, functionName string, payload string, opts InvokeOptions) (*InvokeResult, error) {
	retryConfig := defaultRetryConfig()
	retryConfig.MaxAttempts = opts.MaxRetries
	retryConfig.BaseDelay = opts.RetryDelay
	
	var lastErr error
	
	for attempt := 1; attempt <= retryConfig.MaxAttempts; attempt++ {
		result, err := executeInvoke(client, functionName, payload, opts)
		if err == nil {
			return result, nil
		}
		
		lastErr = err
		
		// Don't retry for certain errors
		if isNonRetryableError(err) {
			return nil, err
		}
		
		// If this was the last attempt, return the error
		if attempt == retryConfig.MaxAttempts {
			break
		}
		
		// Wait before retrying with exponential backoff
		delay := calculateBackoffDelay(attempt, retryConfig)
		time.Sleep(delay)
	}
	
	return nil, lastErr
}

// executeInvoke performs the actual Lambda invocation
func executeInvoke(client *lambda.Client, functionName string, payload string, opts InvokeOptions) (*InvokeResult, error) {
	// Create invocation context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	
	// Prepare invocation input
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: opts.InvocationType,
		LogType:        opts.LogType,
		Payload:        []byte(payload),
	}
	
	// Add optional parameters
	if opts.ClientContext != "" {
		input.ClientContext = aws.String(opts.ClientContext)
	}
	
	if opts.Qualifier != "" {
		input.Qualifier = aws.String(opts.Qualifier)
	}
	
	// Execute the invocation
	output, err := client.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("lambda invoke API call failed: %w", err)
	}
	
	// Parse the response
	result := &InvokeResult{
		StatusCode: output.StatusCode,
		Payload:    string(output.Payload),
	}
	
	if output.LogResult != nil {
		result.LogResult = sanitizeLogResult(*output.LogResult)
	}
	
	if output.ExecutedVersion != nil {
		result.ExecutedVersion = *output.ExecutedVersion
	}
	
	if output.FunctionError != nil {
		result.FunctionError = *output.FunctionError
	}
	
	// Check for function errors
	if result.FunctionError != "" {
		return result, fmt.Errorf("lambda function error: %s", result.FunctionError)
	}
	
	// Validate status code
	if result.StatusCode < 200 || result.StatusCode >= 300 {
		return result, fmt.Errorf("lambda invocation failed with status code: %d", result.StatusCode)
	}
	
	return result, nil
}

// isNonRetryableError determines if an error should not be retried
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	
	// Don't retry for these types of errors
	nonRetryablePatterns := []string{
		"InvalidParameterValueException",
		"ResourceNotFoundException", 
		"InvalidRequestContentException",
		"RequestTooLargeException",
		"UnsupportedMediaTypeException",
	}
	
	for _, pattern := range nonRetryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}
	
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[0:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     indexOf(s, substr) >= 0))
}

// indexOf returns the index of the first occurrence of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ParseInvokeOutput parses the JSON payload from a Lambda invocation result.
// This is the non-error returning version that follows Terratest patterns.
func ParseInvokeOutput(result *InvokeResult, target interface{}) {
	err := ParseInvokeOutputE(result, target)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse Lambda invocation output: %v", err))
	}
}

// ParseInvokeOutputE parses the JSON payload from a Lambda invocation result.
// This is the error returning version that follows Terratest patterns.
func ParseInvokeOutputE(result *InvokeResult, target interface{}) error {
	if result == nil {
		return fmt.Errorf("invoke result is nil")
	}
	
	if result.Payload == "" {
		return fmt.Errorf("payload is empty")
	}
	
	err := json.Unmarshal([]byte(result.Payload), target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	
	return nil
}

// MarshalPayload marshals an object to JSON for use as Lambda payload.
// This is the non-error returning version that follows Terratest patterns.
func MarshalPayload(payload interface{}) string {
	result, err := MarshalPayloadE(payload)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal Lambda payload: %v", err))
	}
	return result
}

// MarshalPayloadE marshals an object to JSON for use as Lambda payload.
// This is the error returning version that follows Terratest patterns.
func MarshalPayloadE(payload interface{}) (string, error) {
	if payload == nil {
		return "", nil
	}
	
	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	return string(bytes), nil
}