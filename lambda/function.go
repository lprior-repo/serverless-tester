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

// UpdateFunctionCode updates the code of a Lambda function.
// This is the non-error returning version that follows Terratest patterns.
func UpdateFunctionCode(ctx *TestContext, input *lambda.UpdateFunctionCodeInput) *UpdateFunctionCodeResult {
	result, err := UpdateFunctionCodeE(ctx, input)
	if err != nil {
		ctx.T.Errorf("Failed to update Lambda function code: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// UpdateFunctionCodeE updates the code of a Lambda function.
// This is the error returning version that follows Terratest patterns.
func UpdateFunctionCodeE(ctx *TestContext, input *lambda.UpdateFunctionCodeInput) (*UpdateFunctionCodeResult, error) {
	startTime := time.Now()
	
	functionName := ""
	if input.FunctionName != nil {
		functionName = *input.FunctionName
	}
	
	// Validate function name
	if err := validateFunctionName(functionName); err != nil {
		return nil, err
	}
	
	logOperation("update_function_code", functionName, map[string]interface{}{
		"has_zip_file":   input.ZipFile != nil,
		"has_s3_bucket":  input.S3Bucket != nil,
		"has_image_uri":  input.ImageUri != nil,
		"publish":        input.Publish,
	})
	
	client := createLambdaClient(ctx)
	
	output, err := client.UpdateFunctionCode(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("update_function_code", functionName, false, duration, err)
		return nil, fmt.Errorf("failed to update function code: %w", err)
	}
	
	result := convertToUpdateFunctionCodeResult(output)
	
	duration := time.Since(startTime)
	logResult("update_function_code", functionName, true, duration, nil)
	
	return result, nil
}

// convertToUpdateFunctionCodeResult converts AWS UpdateFunctionCodeOutput to our type
func convertToUpdateFunctionCodeResult(output *lambda.UpdateFunctionCodeOutput) *UpdateFunctionCodeResult {
	result := &UpdateFunctionCodeResult{}
	
	if output.FunctionName != nil {
		result.FunctionName = *output.FunctionName
	}
	
	if output.FunctionArn != nil {
		result.FunctionArn = *output.FunctionArn
	}
	
	result.Runtime = output.Runtime
	
	if output.Handler != nil {
		result.Handler = *output.Handler
	}
	
	if output.Description != nil {
		result.Description = *output.Description
	}
	
	if output.Timeout != nil {
		result.Timeout = *output.Timeout
	}
	
	if output.MemorySize != nil {
		result.MemorySize = *output.MemorySize
	}
	
	if output.LastModified != nil {
		result.LastModified = *output.LastModified
	}
	
	if output.CodeSha256 != nil {
		result.CodeSha256 = *output.CodeSha256
	}
	
	if output.Version != nil {
		result.Version = *output.Version
	}
	
	result.PackageType = output.PackageType
	
	return result
}

// PublishLayerVersionE creates a new Lambda layer version.
// This is the error returning version that follows Terratest patterns.
func PublishLayerVersionE(ctx *TestContext, input *lambda.PublishLayerVersionInput) (*LayerVersionInfo, error) {
	startTime := time.Now()
	
	layerName := ""
	if input.LayerName != nil {
		layerName = *input.LayerName
	}
	
	// Validate layer name
	if err := validateLayerName(layerName); err != nil {
		return nil, err
	}
	
	logOperation("publish_layer_version", layerName, map[string]interface{}{
		"description": aws.ToString(input.Description),
	})
	
	client := createLambdaClient(ctx)
	
	output, err := client.PublishLayerVersion(context.Background(), input)
	if err != nil {
		duration := time.Since(startTime)
		logResult("publish_layer_version", layerName, false, duration, err)
		return nil, fmt.Errorf("failed to publish layer version: %w", err)
	}
	
	result := &LayerVersionInfo{
		LayerName:          layerName,
		LayerArn:           aws.ToString(output.LayerArn),
		LayerVersionArn:    aws.ToString(output.LayerVersionArn),
		Version:            output.Version,
		Description:        aws.ToString(output.Description),
		CreatedDate:        aws.ToString(output.CreatedDate),
		CompatibleRuntimes: output.CompatibleRuntimes,
	}
	
	if output.Content != nil {
		result.CodeSha256 = aws.ToString(output.Content.CodeSha256)
		result.CodeSize = output.Content.CodeSize
	}
	
	duration := time.Since(startTime)
	logResult("publish_layer_version", layerName, true, duration, nil)
	
	return result, nil
}

// =============================================================================
// FUNCTION URL CONFIGURATION
// =============================================================================

// FunctionUrlConfig represents Lambda function URL configuration
type FunctionUrlConfig struct {
	FunctionName   string
	AuthType       types.FunctionUrlAuthType
	Cors           *types.Cors
	Qualifier      string
}

// FunctionUrlResult represents the result of function URL operations
type FunctionUrlResult struct {
	FunctionUrl      string
	FunctionArn      string
	AuthType         types.FunctionUrlAuthType
	Cors             *types.Cors
	CreationTime     string
	LastModifiedTime string
}

// CreateFunctionUrl creates a function URL configuration and fails the test if an error occurs
func CreateFunctionUrl(ctx *TestContext, config FunctionUrlConfig) *FunctionUrlResult {
	result, err := CreateFunctionUrlE(ctx, config)
	if err != nil {
		ctx.T.Fatalf("Failed to create function URL: %v", err)
	}
	return result
}

// CreateFunctionUrlE creates a function URL configuration and returns any error
func CreateFunctionUrlE(ctx *TestContext, config FunctionUrlConfig) (*FunctionUrlResult, error) {
	if err := validateFunctionName(config.FunctionName); err != nil {
		return nil, fmt.Errorf("invalid function name: %w", err)
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.CreateFunctionUrlConfigInput{
		FunctionName: aws.String(config.FunctionName),
		AuthType:     config.AuthType,
	}

	if config.Cors != nil {
		input.Cors = config.Cors
	}
	if config.Qualifier != "" {
		input.Qualifier = aws.String(config.Qualifier)
	}

	output, err := client.CreateFunctionUrlConfig(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to create function URL: %w", err)
	}

	return &FunctionUrlResult{
		FunctionUrl:      aws.ToString(output.FunctionUrl),
		FunctionArn:      aws.ToString(output.FunctionArn),
		AuthType:         output.AuthType,
		Cors:             output.Cors,
		CreationTime:     aws.ToString(output.CreationTime),
		LastModifiedTime: aws.ToString(output.LastModifiedTime),
	}, nil
}

// GetFunctionUrl gets the function URL configuration and fails the test if an error occurs
func GetFunctionUrl(ctx *TestContext, functionName string) *FunctionUrlResult {
	result, err := GetFunctionUrlE(ctx, functionName)
	if err != nil {
		ctx.T.Fatalf("Failed to get function URL: %v", err)
	}
	return result
}

// GetFunctionUrlE gets the function URL configuration and returns any error
func GetFunctionUrlE(ctx *TestContext, functionName string) (*FunctionUrlResult, error) {
	if err := validateFunctionName(functionName); err != nil {
		return nil, fmt.Errorf("invalid function name: %w", err)
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(functionName),
	}

	output, err := client.GetFunctionUrlConfig(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get function URL: %w", err)
	}

	return &FunctionUrlResult{
		FunctionUrl:      aws.ToString(output.FunctionUrl),
		FunctionArn:      aws.ToString(output.FunctionArn),
		AuthType:         output.AuthType,
		Cors:             output.Cors,
		CreationTime:     aws.ToString(output.CreationTime),
		LastModifiedTime: aws.ToString(output.LastModifiedTime),
	}, nil
}

// DeleteFunctionUrl deletes the function URL configuration and fails the test if an error occurs
func DeleteFunctionUrl(ctx *TestContext, functionName string) {
	err := DeleteFunctionUrlE(ctx, functionName)
	if err != nil {
		ctx.T.Fatalf("Failed to delete function URL: %v", err)
	}
}

// DeleteFunctionUrlE deletes the function URL configuration and returns any error
func DeleteFunctionUrlE(ctx *TestContext, functionName string) error {
	if err := validateFunctionName(functionName); err != nil {
		return fmt.Errorf("invalid function name: %w", err)
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.DeleteFunctionUrlConfigInput{
		FunctionName: aws.String(functionName),
	}

	_, err := client.DeleteFunctionUrlConfig(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to delete function URL: %w", err)
	}

	return nil
}

// UpdateFunctionUrl updates the function URL configuration and fails the test if an error occurs
func UpdateFunctionUrl(ctx *TestContext, config FunctionUrlConfig) *FunctionUrlResult {
	result, err := UpdateFunctionUrlE(ctx, config)
	if err != nil {
		ctx.T.Fatalf("Failed to update function URL: %v", err)
	}
	return result
}

// UpdateFunctionUrlE updates the function URL configuration and returns any error
func UpdateFunctionUrlE(ctx *TestContext, config FunctionUrlConfig) (*FunctionUrlResult, error) {
	if err := validateFunctionName(config.FunctionName); err != nil {
		return nil, fmt.Errorf("invalid function name: %w", err)
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.UpdateFunctionUrlConfigInput{
		FunctionName: aws.String(config.FunctionName),
		AuthType:     config.AuthType,
	}

	if config.Cors != nil {
		input.Cors = config.Cors
	}
	if config.Qualifier != "" {
		input.Qualifier = aws.String(config.Qualifier)
	}

	output, err := client.UpdateFunctionUrlConfig(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to update function URL: %w", err)
	}

	return &FunctionUrlResult{
		FunctionUrl:      aws.ToString(output.FunctionUrl),
		FunctionArn:      aws.ToString(output.FunctionArn),
		AuthType:         output.AuthType,
		Cors:             output.Cors,
		CreationTime:     aws.ToString(output.CreationTime),
		LastModifiedTime: aws.ToString(output.LastModifiedTime),
	}, nil
}

// =============================================================================
// PROVISIONED CONCURRENCY CONFIGURATION
// =============================================================================

// ProvisionedConcurrencyConfig represents provisioned concurrency configuration
type ProvisionedConcurrencyConfig struct {
	FunctionName                    string
	Qualifier                       string
	ProvisionedConcurrencyRequested int32
}

// ProvisionedConcurrencyResult represents the result of provisioned concurrency operations
type ProvisionedConcurrencyResult struct {
	FunctionArn                      string
	Qualifier                        string
	ProvisionedConcurrencyRequested  int32
	ProvisionedConcurrencyAvailable  int32
	ProvisionedConcurrencyAllocated  int32
	Status                           types.ProvisionedConcurrencyStatus
	StatusReason                     string
	LastModified                     string
}

// PutProvisionedConcurrency configures provisioned concurrency and fails the test if an error occurs
func PutProvisionedConcurrency(ctx *TestContext, config ProvisionedConcurrencyConfig) *ProvisionedConcurrencyResult {
	result, err := PutProvisionedConcurrencyE(ctx, config)
	if err != nil {
		ctx.T.Fatalf("Failed to put provisioned concurrency: %v", err)
	}
	return result
}

// PutProvisionedConcurrencyE configures provisioned concurrency and returns any error
func PutProvisionedConcurrencyE(ctx *TestContext, config ProvisionedConcurrencyConfig) (*ProvisionedConcurrencyResult, error) {
	if err := validateFunctionName(config.FunctionName); err != nil {
		return nil, fmt.Errorf("invalid function name: %w", err)
	}

	if config.ProvisionedConcurrencyRequested <= 0 {
		return nil, fmt.Errorf("provisioned concurrency must be greater than 0")
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.PutProvisionedConcurrencyConfigInput{
		FunctionName:                    aws.String(config.FunctionName),
		Qualifier:                       aws.String(config.Qualifier),
		ProvisionedConcurrencyRequested: aws.Int32(config.ProvisionedConcurrencyRequested),
	}

	output, err := client.PutProvisionedConcurrencyConfig(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to put provisioned concurrency: %w", err)
	}

	return &ProvisionedConcurrencyResult{
		FunctionArn:                      aws.ToString(output.FunctionArn),
		Qualifier:                        aws.ToString(output.Qualifier),
		ProvisionedConcurrencyRequested:  aws.ToInt32(output.ProvisionedConcurrencyRequested),
		ProvisionedConcurrencyAvailable:  aws.ToInt32(output.ProvisionedConcurrencyAvailable),
		ProvisionedConcurrencyAllocated:  aws.ToInt32(output.ProvisionedConcurrencyAllocated),
		Status:                           output.Status,
		StatusReason:                     aws.ToString(output.StatusReason),
		LastModified:                     aws.ToString(output.LastModified),
	}, nil
}

// GetProvisionedConcurrency gets the provisioned concurrency configuration and fails the test if an error occurs
func GetProvisionedConcurrency(ctx *TestContext, functionName, qualifier string) *ProvisionedConcurrencyResult {
	result, err := GetProvisionedConcurrencyE(ctx, functionName, qualifier)
	if err != nil {
		ctx.T.Fatalf("Failed to get provisioned concurrency: %v", err)
	}
	return result
}

// GetProvisionedConcurrencyE gets the provisioned concurrency configuration and returns any error
func GetProvisionedConcurrencyE(ctx *TestContext, functionName, qualifier string) (*ProvisionedConcurrencyResult, error) {
	if err := validateFunctionName(functionName); err != nil {
		return nil, fmt.Errorf("invalid function name: %w", err)
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.GetProvisionedConcurrencyConfigInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(qualifier),
	}

	output, err := client.GetProvisionedConcurrencyConfig(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get provisioned concurrency: %w", err)
	}

	return &ProvisionedConcurrencyResult{
		FunctionArn:                      aws.ToString(output.FunctionArn),
		Qualifier:                        aws.ToString(output.Qualifier),
		ProvisionedConcurrencyRequested:  aws.ToInt32(output.ProvisionedConcurrencyRequested),
		ProvisionedConcurrencyAvailable:  aws.ToInt32(output.ProvisionedConcurrencyAvailable),
		ProvisionedConcurrencyAllocated:  aws.ToInt32(output.ProvisionedConcurrencyAllocated),
		Status:                           output.Status,
		StatusReason:                     aws.ToString(output.StatusReason),
		LastModified:                     aws.ToString(output.LastModified),
	}, nil
}

// DeleteProvisionedConcurrency deletes the provisioned concurrency configuration and fails the test if an error occurs
func DeleteProvisionedConcurrency(ctx *TestContext, functionName, qualifier string) {
	err := DeleteProvisionedConcurrencyE(ctx, functionName, qualifier)
	if err != nil {
		ctx.T.Fatalf("Failed to delete provisioned concurrency: %v", err)
	}
}

// DeleteProvisionedConcurrencyE deletes the provisioned concurrency configuration and returns any error
func DeleteProvisionedConcurrencyE(ctx *TestContext, functionName, qualifier string) error {
	if err := validateFunctionName(functionName); err != nil {
		return fmt.Errorf("invalid function name: %w", err)
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.DeleteProvisionedConcurrencyConfigInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(qualifier),
	}

	_, err := client.DeleteProvisionedConcurrencyConfig(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to delete provisioned concurrency: %w", err)
	}

	return nil
}

// =============================================================================
// RESERVED CONCURRENCY CONFIGURATION
// =============================================================================

// ReservedConcurrencyConfig represents reserved concurrency configuration
type ReservedConcurrencyConfig struct {
	FunctionName           string
	ReservedConcurrency    int32
}

// ReservedConcurrencyResult represents the result of reserved concurrency operations
type ReservedConcurrencyResult struct {
	FunctionArn         string
	ReservedConcurrency int32
}

// PutReservedConcurrency sets the reserved concurrency and fails the test if an error occurs
func PutReservedConcurrency(ctx *TestContext, config ReservedConcurrencyConfig) *ReservedConcurrencyResult {
	result, err := PutReservedConcurrencyE(ctx, config)
	if err != nil {
		ctx.T.Fatalf("Failed to put reserved concurrency: %v", err)
	}
	return result
}

// PutReservedConcurrencyE sets the reserved concurrency and returns any error
func PutReservedConcurrencyE(ctx *TestContext, config ReservedConcurrencyConfig) (*ReservedConcurrencyResult, error) {
	if err := validateFunctionName(config.FunctionName); err != nil {
		return nil, fmt.Errorf("invalid function name: %w", err)
	}

	if config.ReservedConcurrency < 0 {
		return nil, fmt.Errorf("reserved concurrency cannot be negative")
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.PutReservedConcurrencyInput{
		FunctionName:        aws.String(config.FunctionName),
		ReservedConcurrency: aws.Int32(config.ReservedConcurrency),
	}

	output, err := client.PutReservedConcurrency(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to put reserved concurrency: %w", err)
	}

	return &ReservedConcurrencyResult{
		FunctionArn:         aws.ToString(output.FunctionArn),
		ReservedConcurrency: aws.ToInt32(output.ReservedConcurrency),
	}, nil
}

// GetReservedConcurrency gets the reserved concurrency and fails the test if an error occurs
func GetReservedConcurrency(ctx *TestContext, functionName string) *ReservedConcurrencyResult {
	result, err := GetReservedConcurrencyE(ctx, functionName)
	if err != nil {
		ctx.T.Fatalf("Failed to get reserved concurrency: %v", err)
	}
	return result
}

// GetReservedConcurrencyE gets the reserved concurrency and returns any error
func GetReservedConcurrencyE(ctx *TestContext, functionName string) (*ReservedConcurrencyResult, error) {
	if err := validateFunctionName(functionName); err != nil {
		return nil, fmt.Errorf("invalid function name: %w", err)
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.GetReservedConcurrencyInput{
		FunctionName: aws.String(functionName),
	}

	output, err := client.GetReservedConcurrency(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get reserved concurrency: %w", err)
	}

	return &ReservedConcurrencyResult{
		FunctionArn:         aws.ToString(output.FunctionArn),
		ReservedConcurrency: aws.ToInt32(output.ReservedConcurrency),
	}, nil
}

// DeleteReservedConcurrency removes the reserved concurrency and fails the test if an error occurs
func DeleteReservedConcurrency(ctx *TestContext, functionName string) {
	err := DeleteReservedConcurrencyE(ctx, functionName)
	if err != nil {
		ctx.T.Fatalf("Failed to delete reserved concurrency: %v", err)
	}
}

// DeleteReservedConcurrencyE removes the reserved concurrency and returns any error
func DeleteReservedConcurrencyE(ctx *TestContext, functionName string) error {
	if err := validateFunctionName(functionName); err != nil {
		return fmt.Errorf("invalid function name: %w", err)
	}

	client := createLambdaClient(ctx)
	
	input := &lambda.DeleteReservedConcurrencyInput{
		FunctionName: aws.String(functionName),
	}

	_, err := client.DeleteReservedConcurrency(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to delete reserved concurrency: %w", err)
	}

	return nil
}

// validateLayerName validates that the layer name follows AWS naming conventions
func validateLayerName(layerName string) error {
	if layerName == "" {
		return fmt.Errorf("invalid layer name: empty name")
	}
	
	if len(layerName) > 64 {
		return fmt.Errorf("invalid layer name: name too long (%d characters)", len(layerName))
	}
	
	// Layer names must match pattern: [a-zA-Z0-9-_]+
	for _, char := range layerName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return fmt.Errorf("invalid layer name: invalid character '%c'", char)
		}
	}
	
	return nil
}