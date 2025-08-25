package lambda

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// AssertFunctionExists asserts that a Lambda function exists.
// This follows Terratest assertion patterns and fails the test if assertion fails.
func AssertFunctionExists(ctx *TestContext, functionName string) {
	exists, err := FunctionExistsE(ctx, functionName)
	if err != nil {
		ctx.T.Errorf("Failed to check if function exists: %v", err)
		ctx.T.FailNow()
	}
	
	if !exists {
		ctx.T.Errorf("Expected Lambda function '%s' to exist, but it does not", functionName)
		ctx.T.FailNow()
	}
}

// AssertFunctionDoesNotExist asserts that a Lambda function does not exist.
// This follows Terratest assertion patterns and fails the test if assertion fails.
func AssertFunctionDoesNotExist(ctx *TestContext, functionName string) {
	exists, err := FunctionExistsE(ctx, functionName)
	if err != nil {
		// If we get an error other than "not found", that's a problem
		if !strings.Contains(err.Error(), "ResourceNotFoundException") &&
		   !strings.Contains(err.Error(), "Function not found") {
			ctx.T.Errorf("Failed to check if function exists: %v", err)
			ctx.T.FailNow()
		}
		// If we get a "not found" error, that's what we expect
		return
	}
	
	if exists {
		ctx.T.Errorf("Expected Lambda function '%s' not to exist, but it does", functionName)
		ctx.T.FailNow()
	}
}

// AssertFunctionRuntime asserts that a Lambda function has the expected runtime.
func AssertFunctionRuntime(ctx *TestContext, functionName string, expectedRuntime types.Runtime) {
	config := GetFunction(ctx, functionName)
	
	if config.Runtime != expectedRuntime {
		ctx.T.Errorf("Expected function '%s' to have runtime '%s', but got '%s'", 
			functionName, string(expectedRuntime), string(config.Runtime))
		ctx.T.FailNow()
	}
}

// AssertFunctionHandler asserts that a Lambda function has the expected handler.
func AssertFunctionHandler(ctx *TestContext, functionName string, expectedHandler string) {
	config := GetFunction(ctx, functionName)
	
	if config.Handler != expectedHandler {
		ctx.T.Errorf("Expected function '%s' to have handler '%s', but got '%s'", 
			functionName, expectedHandler, config.Handler)
		ctx.T.FailNow()
	}
}

// AssertFunctionTimeout asserts that a Lambda function has the expected timeout.
func AssertFunctionTimeout(ctx *TestContext, functionName string, expectedTimeout int32) {
	config := GetFunction(ctx, functionName)
	
	if config.Timeout != expectedTimeout {
		ctx.T.Errorf("Expected function '%s' to have timeout %d seconds, but got %d", 
			functionName, expectedTimeout, config.Timeout)
		ctx.T.FailNow()
	}
}

// AssertFunctionMemorySize asserts that a Lambda function has the expected memory size.
func AssertFunctionMemorySize(ctx *TestContext, functionName string, expectedMemorySize int32) {
	config := GetFunction(ctx, functionName)
	
	if config.MemorySize != expectedMemorySize {
		ctx.T.Errorf("Expected function '%s' to have memory size %d MB, but got %d", 
			functionName, expectedMemorySize, config.MemorySize)
		ctx.T.FailNow()
	}
}

// AssertFunctionState asserts that a Lambda function is in the expected state.
func AssertFunctionState(ctx *TestContext, functionName string, expectedState types.State) {
	config := GetFunction(ctx, functionName)
	
	if config.State != expectedState {
		ctx.T.Errorf("Expected function '%s' to be in state '%s', but got '%s'", 
			functionName, string(expectedState), string(config.State))
		ctx.T.FailNow()
	}
}

// AssertEnvVarEquals asserts that a Lambda function's environment variable has the expected value.
func AssertEnvVarEquals(ctx *TestContext, functionName string, varName string, expectedValue string) {
	actualValue := GetEnvVar(ctx, functionName, varName)
	
	if actualValue != expectedValue {
		ctx.T.Errorf("Expected environment variable '%s' in function '%s' to be '%s', but got '%s'", 
			varName, functionName, expectedValue, actualValue)
		ctx.T.FailNow()
	}
}

// AssertEnvVarExists asserts that a Lambda function has the specified environment variable.
func AssertEnvVarExists(ctx *TestContext, functionName string, varName string) {
	_, err := GetEnvVarE(ctx, functionName, varName)
	if err != nil {
		ctx.T.Errorf("Expected environment variable '%s' to exist in function '%s', but it does not: %v", 
			varName, functionName, err)
		ctx.T.FailNow()
	}
}

// AssertEnvVarDoesNotExist asserts that a Lambda function does not have the specified environment variable.
func AssertEnvVarDoesNotExist(ctx *TestContext, functionName string, varName string) {
	_, err := GetEnvVarE(ctx, functionName, varName)
	if err == nil {
		ctx.T.Errorf("Expected environment variable '%s' not to exist in function '%s', but it does", 
			varName, functionName)
		ctx.T.FailNow()
	}
}

// AssertInvokeSuccess asserts that a Lambda invocation was successful.
func AssertInvokeSuccess(ctx *TestContext, result *InvokeResult) {
	if result == nil {
		ctx.T.Errorf("Expected invocation result to be non-nil")
		ctx.T.FailNow()
		return
	}
	
	if result.FunctionError != "" {
		ctx.T.Errorf("Expected successful invocation, but got function error: %s", result.FunctionError)
		ctx.T.FailNow()
		return
	}
	
	if result.StatusCode < 200 || result.StatusCode >= 300 {
		ctx.T.Errorf("Expected successful status code (2xx), but got %d", result.StatusCode)
		ctx.T.FailNow()
		return
	}
}

// AssertInvokeError asserts that a Lambda invocation resulted in an error.
func AssertInvokeError(ctx *TestContext, result *InvokeResult) {
	if result == nil {
		ctx.T.Errorf("Expected invocation result to be non-nil")
		ctx.T.FailNow()
		return
	}
	
	if result.FunctionError == "" {
		ctx.T.Errorf("Expected invocation to result in function error, but got none")
		ctx.T.FailNow()
		return
	}
}

// AssertPayloadContains asserts that the invocation payload contains the expected text.
func AssertPayloadContains(ctx *TestContext, result *InvokeResult, expectedText string) {
	if result == nil {
		ctx.T.Errorf("Expected invocation result to be non-nil")
		ctx.T.FailNow()
		return
	}
	
	if !strings.Contains(result.Payload, expectedText) {
		ctx.T.Errorf("Expected payload to contain '%s', but payload was: %s", 
			expectedText, result.Payload)
		ctx.T.FailNow()
		return
	}
}

// AssertPayloadEquals asserts that the invocation payload equals the expected value.
func AssertPayloadEquals(ctx *TestContext, result *InvokeResult, expectedPayload string) {
	if result == nil {
		ctx.T.Errorf("Expected invocation result to be non-nil")
		ctx.T.FailNow()
		return
	}
	
	if result.Payload != expectedPayload {
		ctx.T.Errorf("Expected payload to be '%s', but got '%s'", 
			expectedPayload, result.Payload)
		ctx.T.FailNow()
		return
	}
}

// AssertExecutionTimeLessThan asserts that the Lambda execution time is less than the expected duration.
func AssertExecutionTimeLessThan(ctx *TestContext, result *InvokeResult, maxDuration time.Duration) {
	if result == nil {
		ctx.T.Errorf("Expected invocation result to be non-nil")
		ctx.T.FailNow()
		return
	}
	
	if result.ExecutionTime >= maxDuration {
		ctx.T.Errorf("Expected execution time to be less than %v, but got %v", 
			maxDuration, result.ExecutionTime)
		ctx.T.FailNow()
		return
	}
}

// AssertLogsContain asserts that the function logs contain the expected text.
func AssertLogsContain(ctx *TestContext, functionName string, expectedText string) {
	logs := GetRecentLogs(ctx, functionName, 5*time.Minute)
	
	if !LogsContain(logs, expectedText) {
		// Create a summary of log messages for better error reporting
		var logMessages []string
		for _, log := range logs {
			logMessages = append(logMessages, log.Message)
		}
		
		ctx.T.Errorf("Expected logs for function '%s' to contain '%s', but got logs: %v", 
			functionName, expectedText, logMessages)
		ctx.T.FailNow()
	}
}

// AssertLogsContainLevel asserts that the function logs contain entries at the specified level.
func AssertLogsContainLevel(ctx *TestContext, functionName string, expectedLevel string) {
	logs := GetRecentLogs(ctx, functionName, 5*time.Minute)
	
	if !LogsContainLevel(logs, expectedLevel) {
		// Create a summary of log levels for better error reporting
		levels := make(map[string]int)
		for _, log := range logs {
			if log.Level != "" {
				levels[log.Level]++
			}
		}
		
		ctx.T.Errorf("Expected logs for function '%s' to contain level '%s', but found levels: %v", 
			functionName, expectedLevel, levels)
		ctx.T.FailNow()
	}
}

// AssertLogsDoNotContain asserts that the function logs do not contain the specified text.
func AssertLogsDoNotContain(ctx *TestContext, functionName string, forbiddenText string) {
	logs := GetRecentLogs(ctx, functionName, 5*time.Minute)
	
	if LogsContain(logs, forbiddenText) {
		ctx.T.Errorf("Expected logs for function '%s' not to contain '%s', but they do", 
			functionName, forbiddenText)
		ctx.T.FailNow()
	}
}

// AssertLogCount asserts that the function has the expected number of log entries.
func AssertLogCount(ctx *TestContext, functionName string, expectedCount int, duration time.Duration) {
	logs := GetRecentLogs(ctx, functionName, duration)
	
	actualCount := len(logs)
	if actualCount != expectedCount {
		ctx.T.Errorf("Expected function '%s' to have %d log entries, but got %d", 
			functionName, expectedCount, actualCount)
		ctx.T.FailNow()
	}
}

// Validation functions that return boolean values for use in conditional logic

// ValidateFunctionConfiguration validates all aspects of a function configuration.
func ValidateFunctionConfiguration(config *FunctionConfiguration, expected *FunctionConfiguration) []string {
	var errors []string
	
	if config == nil {
		errors = append(errors, "function configuration is nil")
		return errors
	}
	
	if expected == nil {
		return errors // Nothing to validate against
	}
	
	if expected.FunctionName != "" && config.FunctionName != expected.FunctionName {
		errors = append(errors, fmt.Sprintf("expected function name '%s', got '%s'", 
			expected.FunctionName, config.FunctionName))
	}
	
	if expected.Runtime != "" && config.Runtime != expected.Runtime {
		errors = append(errors, fmt.Sprintf("expected runtime '%s', got '%s'", 
			string(expected.Runtime), string(config.Runtime)))
	}
	
	if expected.Handler != "" && config.Handler != expected.Handler {
		errors = append(errors, fmt.Sprintf("expected handler '%s', got '%s'", 
			expected.Handler, config.Handler))
	}
	
	if expected.Timeout > 0 && config.Timeout != expected.Timeout {
		errors = append(errors, fmt.Sprintf("expected timeout %d, got %d", 
			expected.Timeout, config.Timeout))
	}
	
	if expected.MemorySize > 0 && config.MemorySize != expected.MemorySize {
		errors = append(errors, fmt.Sprintf("expected memory size %d, got %d", 
			expected.MemorySize, config.MemorySize))
	}
	
	if expected.State != "" && config.State != expected.State {
		errors = append(errors, fmt.Sprintf("expected state '%s', got '%s'", 
			string(expected.State), string(config.State)))
	}
	
	// Validate environment variables
	if len(expected.Environment) > 0 {
		for key, expectedValue := range expected.Environment {
			if actualValue, exists := config.Environment[key]; !exists {
				errors = append(errors, fmt.Sprintf("expected environment variable '%s' not found", key))
			} else if actualValue != expectedValue {
				errors = append(errors, fmt.Sprintf("expected env var '%s'='%s', got '%s'", 
					key, expectedValue, actualValue))
			}
		}
	}
	
	return errors
}

// ValidateInvokeResult validates an invocation result against expected criteria.
func ValidateInvokeResult(result *InvokeResult, expectSuccess bool, expectedPayloadContains string) []string {
	var errors []string
	
	if result == nil {
		errors = append(errors, "invoke result is nil")
		return errors
	}
	
	// Validate success/failure expectation
	isSuccess := result.FunctionError == "" && result.StatusCode >= 200 && result.StatusCode < 300
	if expectSuccess && !isSuccess {
		if result.FunctionError != "" {
			errors = append(errors, fmt.Sprintf("expected success but got function error: %s", result.FunctionError))
		} else {
			errors = append(errors, fmt.Sprintf("expected success but got status code: %d", result.StatusCode))
		}
	}
	
	if !expectSuccess && isSuccess {
		errors = append(errors, "expected failure but invocation succeeded")
	}
	
	// Validate payload content
	if expectedPayloadContains != "" && !strings.Contains(result.Payload, expectedPayloadContains) {
		errors = append(errors, fmt.Sprintf("expected payload to contain '%s', but payload was: %s", 
			expectedPayloadContains, result.Payload))
	}
	
	return errors
}

// ValidateLogEntries validates log entries against expected criteria.
func ValidateLogEntries(logs []LogEntry, expectedCount int, expectedContains string, expectedLevel string) []string {
	var errors []string
	
	// Validate count
	if expectedCount >= 0 && len(logs) != expectedCount {
		errors = append(errors, fmt.Sprintf("expected %d log entries, got %d", expectedCount, len(logs)))
	}
	
	// Validate content
	if expectedContains != "" && !LogsContain(logs, expectedContains) {
		errors = append(errors, fmt.Sprintf("expected logs to contain '%s'", expectedContains))
	}
	
	// Validate level
	if expectedLevel != "" && !LogsContainLevel(logs, expectedLevel) {
		errors = append(errors, fmt.Sprintf("expected logs to contain level '%s'", expectedLevel))
	}
	
	return errors
}