package lambda

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// GetFunction retrieves the configuration of a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func GetFunction(ctx *TestContext, functionName string) *FunctionConfiguration {
	config, err := GetFunctionE(ctx, functionName)
	if err != nil {
		ctx.T.Errorf("Failed to get Lambda function configuration: %v", err)
		ctx.T.FailNow()
	}
	return config
}

// GetFunctionE retrieves the configuration of a Lambda function.
// This is the error returning version that follows Terratest patterns.
func GetFunctionE(ctx *TestContext, functionName string) (*FunctionConfiguration, error) {
	startTime := time.Now()
	
	// Validate function name
	if err := validateFunctionName(functionName); err != nil {
		return nil, err
	}
	
	// Log operation start
	logOperation("get_function", functionName, map[string]interface{}{})
	
	// Create Lambda client
	client := createLambdaClient(ctx)
	
	// Execute get function operation
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	}
	
	output, err := client.GetFunction(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("get_function", functionName, false, duration, err)
		return nil, fmt.Errorf("failed to get function configuration: %w", err)
	}
	
	// Convert AWS types to our configuration type
	config := convertToFunctionConfiguration(output.Configuration)
	
	duration := time.Since(startTime)
	logResult("get_function", functionName, true, duration, nil)
	
	return config, nil
}

// FunctionExists checks if a Lambda function exists.
// This is the non-error returning version that follows Terratest patterns.
func FunctionExists(ctx *TestContext, functionName string) bool {
	exists, err := FunctionExistsE(ctx, functionName)
	if err != nil {
		ctx.T.Errorf("Failed to check if Lambda function exists: %v", err)
		ctx.T.FailNow()
	}
	return exists
}

// FunctionExistsE checks if a Lambda function exists.
// This is the error returning version that follows Terratest patterns.
func FunctionExistsE(ctx *TestContext, functionName string) (bool, error) {
	_, err := GetFunctionE(ctx, functionName)
	if err != nil {
		// Check if the error is "function not found"
		if strings.Contains(err.Error(), "ResourceNotFoundException") ||
		   strings.Contains(err.Error(), "Function not found") {
			return false, nil
		}
		// Other errors should be propagated
		return false, err
	}
	return true, nil
}

// WaitForFunctionActive waits for a Lambda function to be in Active state.
// This is the non-error returning version that follows Terratest patterns.
func WaitForFunctionActive(ctx *TestContext, functionName string, timeout time.Duration) {
	err := WaitForFunctionActiveE(ctx, functionName, timeout)
	if err != nil {
		ctx.T.Errorf("Function did not become active within timeout: %v", err)
		ctx.T.FailNow()
	}
}

// WaitForFunctionActiveE waits for a Lambda function to be in Active state.
// This is the error returning version that follows Terratest patterns.
func WaitForFunctionActiveE(ctx *TestContext, functionName string, timeout time.Duration) error {
	startTime := time.Now()
	
	logOperation("wait_for_function_active", functionName, map[string]interface{}{
		"timeout_seconds": timeout.Seconds(),
	})
	
	// Poll the function state until it's active or timeout
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	timeoutChan := time.After(timeout)
	
	for {
		select {
		case <-timeoutChan:
			duration := time.Since(startTime)
			logResult("wait_for_function_active", functionName, false, duration, 
				fmt.Errorf("timeout waiting for function to be active"))
			return fmt.Errorf("timeout waiting for function %s to be active", functionName)
			
		case <-ticker.C:
			config, err := GetFunctionE(ctx, functionName)
			if err != nil {
				duration := time.Since(startTime)
				logResult("wait_for_function_active", functionName, false, duration, err)
				return err
			}
			
			if config.State == types.StateActive {
				duration := time.Since(startTime)
				logResult("wait_for_function_active", functionName, true, duration, nil)
				return nil
			}
		}
	}
}

// UpdateFunction updates the configuration of a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func UpdateFunction(ctx *TestContext, config UpdateFunctionConfig) *FunctionConfiguration {
	result, err := UpdateFunctionE(ctx, config)
	if err != nil {
		ctx.T.Errorf("Failed to update Lambda function: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// UpdateFunctionE updates the configuration of a Lambda function.
// This is the error returning version that follows Terratest patterns.
func UpdateFunctionE(ctx *TestContext, config UpdateFunctionConfig) (*FunctionConfiguration, error) {
	startTime := time.Now()
	
	// Validate function name
	if err := validateFunctionName(config.FunctionName); err != nil {
		return nil, err
	}
	
	// Log operation start
	logOperation("update_function", config.FunctionName, map[string]interface{}{
		"runtime":     string(config.Runtime),
		"handler":     config.Handler,
		"timeout":     config.Timeout,
		"memory_size": config.MemorySize,
	})
	
	// Create Lambda client
	client := createLambdaClient(ctx)
	
	// Update function configuration
	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(config.FunctionName),
	}
	
	// Set optional fields
	if config.Runtime != "" {
		input.Runtime = config.Runtime
	}
	
	if config.Handler != "" {
		input.Handler = aws.String(config.Handler)
	}
	
	if config.Description != "" {
		input.Description = aws.String(config.Description)
	}
	
	if config.Timeout > 0 {
		input.Timeout = aws.Int32(config.Timeout)
	}
	
	if config.MemorySize > 0 {
		input.MemorySize = aws.Int32(config.MemorySize)
	}
	
	if len(config.Environment) > 0 {
		input.Environment = &types.Environment{
			Variables: config.Environment,
		}
	}
	
	if config.DeadLetterConfig != nil {
		input.DeadLetterConfig = config.DeadLetterConfig
	}
	
	// Execute update
	output, err := client.UpdateFunctionConfiguration(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("update_function", config.FunctionName, false, duration, err)
		return nil, fmt.Errorf("failed to update function configuration: %w", err)
	}
	
	// Convert response to our type
	result := convertToFunctionConfiguration(output)
	
	duration := time.Since(startTime)
	logResult("update_function", config.FunctionName, true, duration, nil)
	
	return result, nil
}

// GetEnvVar retrieves an environment variable from a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func GetEnvVar(ctx *TestContext, functionName string, varName string) string {
	value, err := GetEnvVarE(ctx, functionName, varName)
	if err != nil {
		ctx.T.Errorf("Failed to get environment variable: %v", err)
		ctx.T.FailNow()
	}
	return value
}

// GetEnvVarE retrieves an environment variable from a Lambda function.
// This is the error returning version that follows Terratest patterns.
func GetEnvVarE(ctx *TestContext, functionName string, varName string) (string, error) {
	config, err := GetFunctionE(ctx, functionName)
	if err != nil {
		return "", err
	}
	
	if config.Environment == nil {
		return "", fmt.Errorf("function %s has no environment variables", functionName)
	}
	
	value, exists := config.Environment[varName]
	if !exists {
		return "", fmt.Errorf("environment variable %s not found in function %s", varName, functionName)
	}
	
	return value, nil
}

// ListFunctions lists all Lambda functions in the current region.
// This is the non-error returning version that follows Terratest patterns.
func ListFunctions(ctx *TestContext) []FunctionConfiguration {
	functions, err := ListFunctionsE(ctx)
	if err != nil {
		ctx.T.Errorf("Failed to list Lambda functions: %v", err)
		ctx.T.FailNow()
	}
	return functions
}

// ListFunctionsE lists all Lambda functions in the current region.
// This is the error returning version that follows Terratest patterns.
func ListFunctionsE(ctx *TestContext) ([]FunctionConfiguration, error) {
	startTime := time.Now()
	
	logOperation("list_functions", "all", map[string]interface{}{})
	
	client := createLambdaClient(ctx)
	
	var allFunctions []FunctionConfiguration
	var nextMarker *string
	
	for {
		input := &lambda.ListFunctionsInput{
			MaxItems: aws.Int32(100), // AWS maximum
		}
		
		if nextMarker != nil {
			input.Marker = nextMarker
		}
		
		output, err := client.ListFunctions(context.Background(), input)
		if err != nil {
			duration := time.Since(startTime)
			logResult("list_functions", "all", false, duration, err)
			return nil, fmt.Errorf("failed to list functions: %w", err)
		}
		
		// Convert AWS types to our configuration types
		for _, awsFunction := range output.Functions {
			config := convertToFunctionConfiguration(&awsFunction)
			allFunctions = append(allFunctions, *config)
		}
		
		// Check if there are more functions to retrieve
		if output.NextMarker == nil {
			break
		}
		nextMarker = output.NextMarker
	}
	
	duration := time.Since(startTime)
	logResult("list_functions", "all", true, duration, nil)
	
	return allFunctions, nil
}

// convertToFunctionConfiguration converts AWS Lambda configuration to our type
func convertToFunctionConfiguration(awsConfig interface{}) *FunctionConfiguration {
	config := &FunctionConfiguration{}
	
	// Handle different AWS SDK output types
	switch v := awsConfig.(type) {
	case *lambda.GetFunctionOutput:
		return convertFunctionConfig(v.Configuration)
	case *lambda.CreateFunctionOutput:
		return convertCreateFunctionOutput(v)
	case *lambda.UpdateFunctionConfigurationOutput:
		return convertUpdateFunctionOutput(v)
	case *types.FunctionConfiguration:
		return convertFunctionConfig(v)
	default:
		return config
	}
}

// convertFunctionConfig converts types.FunctionConfiguration to our type
func convertFunctionConfig(awsConfig *types.FunctionConfiguration) *FunctionConfiguration {
	config := &FunctionConfiguration{}
	
	if awsConfig.FunctionName != nil {
		config.FunctionName = *awsConfig.FunctionName
	}
	
	if awsConfig.FunctionArn != nil {
		config.FunctionArn = *awsConfig.FunctionArn
	}
	
	config.Runtime = awsConfig.Runtime
	
	if awsConfig.Handler != nil {
		config.Handler = *awsConfig.Handler
	}
	
	if awsConfig.Description != nil {
		config.Description = *awsConfig.Description
	}
	
	if awsConfig.Timeout != nil {
		config.Timeout = *awsConfig.Timeout
	}
	
	if awsConfig.MemorySize != nil {
		config.MemorySize = *awsConfig.MemorySize
	}
	
	if awsConfig.LastModified != nil {
		config.LastModified = *awsConfig.LastModified
	}
	
	if awsConfig.Role != nil {
		config.Role = *awsConfig.Role
	}
	
	config.State = awsConfig.State
	
	if awsConfig.StateReason != nil {
		config.StateReason = *awsConfig.StateReason
	}
	
	if awsConfig.Version != nil {
		config.Version = *awsConfig.Version
	}
	
	// Convert environment variables
	if awsConfig.Environment != nil && awsConfig.Environment.Variables != nil {
		config.Environment = make(map[string]string)
		for key, value := range awsConfig.Environment.Variables {
			config.Environment[key] = value
		}
	}
	
	return config
}

// convertCreateFunctionOutput converts CreateFunctionOutput to our type
func convertCreateFunctionOutput(output *lambda.CreateFunctionOutput) *FunctionConfiguration {
	config := &FunctionConfiguration{}
	
	if output.FunctionName != nil {
		config.FunctionName = *output.FunctionName
	}
	
	if output.FunctionArn != nil {
		config.FunctionArn = *output.FunctionArn
	}
	
	config.Runtime = output.Runtime
	
	if output.Handler != nil {
		config.Handler = *output.Handler
	}
	
	if output.Description != nil {
		config.Description = *output.Description
	}
	
	if output.Timeout != nil {
		config.Timeout = *output.Timeout
	}
	
	if output.MemorySize != nil {
		config.MemorySize = *output.MemorySize
	}
	
	if output.LastModified != nil {
		config.LastModified = *output.LastModified
	}
	
	if output.Role != nil {
		config.Role = *output.Role
	}
	
	config.State = output.State
	
	if output.StateReason != nil {
		config.StateReason = *output.StateReason
	}
	
	if output.Version != nil {
		config.Version = *output.Version
	}
	
	// Convert environment variables
	if output.Environment != nil && output.Environment.Variables != nil {
		config.Environment = make(map[string]string)
		for key, value := range output.Environment.Variables {
			config.Environment[key] = value
		}
	}
	
	return config
}

// convertUpdateFunctionOutput converts UpdateFunctionConfigurationOutput to our type
func convertUpdateFunctionOutput(output *lambda.UpdateFunctionConfigurationOutput) *FunctionConfiguration {
	config := &FunctionConfiguration{}
	
	if output.FunctionName != nil {
		config.FunctionName = *output.FunctionName
	}
	
	if output.FunctionArn != nil {
		config.FunctionArn = *output.FunctionArn
	}
	
	config.Runtime = output.Runtime
	
	if output.Handler != nil {
		config.Handler = *output.Handler
	}
	
	if output.Description != nil {
		config.Description = *output.Description
	}
	
	if output.Timeout != nil {
		config.Timeout = *output.Timeout
	}
	
	if output.MemorySize != nil {
		config.MemorySize = *output.MemorySize
	}
	
	if output.LastModified != nil {
		config.LastModified = *output.LastModified
	}
	
	if output.Role != nil {
		config.Role = *output.Role
	}
	
	config.State = output.State
	
	if output.StateReason != nil {
		config.StateReason = *output.StateReason
	}
	
	if output.Version != nil {
		config.Version = *output.Version
	}
	
	// Convert environment variables
	if output.Environment != nil && output.Environment.Variables != nil {
		config.Environment = make(map[string]string)
		for key, value := range output.Environment.Variables {
			config.Environment[key] = value
		}
	}
	
	return config
}

// CreateFunction creates a new Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func CreateFunction(ctx *TestContext, input *lambda.CreateFunctionInput) *FunctionConfiguration {
	config, err := CreateFunctionE(ctx, input)
	if err != nil {
		ctx.T.Errorf("Failed to create Lambda function: %v", err)
		ctx.T.FailNow()
	}
	return config
}

// CreateFunctionE creates a new Lambda function.
// This is the error returning version that follows Terratest patterns.
func CreateFunctionE(ctx *TestContext, input *lambda.CreateFunctionInput) (*FunctionConfiguration, error) {
	startTime := time.Now()
	
	functionName := ""
	if input.FunctionName != nil {
		functionName = *input.FunctionName
	}
	
	// Validate function name
	if err := validateFunctionName(functionName); err != nil {
		return nil, err
	}
	
	logOperation("create_function", functionName, map[string]interface{}{
		"runtime": string(input.Runtime),
		"handler": aws.ToString(input.Handler),
	})
	
	client := createLambdaClient(ctx)
	
	output, err := client.CreateFunction(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("create_function", functionName, false, duration, err)
		return nil, fmt.Errorf("failed to create function: %w", err)
	}
	
	config := convertToFunctionConfiguration(output)
	
	duration := time.Since(startTime)
	logResult("create_function", functionName, true, duration, nil)
	
	return config, nil
}

// DeleteFunction deletes a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func DeleteFunction(ctx *TestContext, functionName string) {
	err := DeleteFunctionE(ctx, functionName)
	if err != nil {
		ctx.T.Errorf("Failed to delete Lambda function: %v", err)
		ctx.T.FailNow()
	}
}

// DeleteFunctionE deletes a Lambda function.
// This is the error returning version that follows Terratest patterns.
func DeleteFunctionE(ctx *TestContext, functionName string) error {
	startTime := time.Now()
	
	// Validate function name
	if err := validateFunctionName(functionName); err != nil {
		return err
	}
	
	logOperation("delete_function", functionName, map[string]interface{}{})
	
	client := createLambdaClient(ctx)
	
	input := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	}
	
	_, err := client.DeleteFunction(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("delete_function", functionName, false, duration, err)
		return fmt.Errorf("failed to delete function: %w", err)
	}
	
	duration := time.Since(startTime)
	logResult("delete_function", functionName, true, duration, nil)
	
	return nil
}