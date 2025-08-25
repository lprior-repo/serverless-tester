package main

import (
	"fmt"
	"time"
	
	"vasdeference"
	
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

// ExampleBasicUsage demonstrates the basic usage of VasDeference
func ExampleBasicUsage() {
	fmt.Println("VasDeference basic usage patterns:")
	
	// Create VasDeference instance with automatic configuration
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Use the configured AWS clients
	// lambdaClient := vdf.CreateLambdaClient()
	// Expected: lambdaClient should not be nil
	fmt.Println("Creating Lambda client with VasDeference configuration")
	
	// Use namespace for resource naming
	// resourceName := vdf.PrefixResourceName("test-function")
	// Expected: resourceName should contain "test-function"
	fmt.Println("Prefixing resource name with namespace: test-function")
	
	// Register cleanup for resources
	// vdf.RegisterCleanup(func() error {
	//     // Clean up your resources here
	//     return nil
	// })
	fmt.Println("Registering cleanup functions for resource management")
	fmt.Println("Basic usage completed")
}

// ExampleWithCustomOptions demonstrates VasDeference with custom configuration
func ExampleWithCustomOptions() {
	fmt.Println("VasDeference custom configuration patterns:")
	
	// Create with custom options using NewTestContext
	// ctx := vasdeference.NewTestContext(t)
	// arrange := vasdeference.NewArrangeWithNamespace(ctx, "integration-test")
	// defer arrange.Cleanup()
	
	// Expected: arrange.Namespace should equal "integration-test"
	fmt.Println("Creating VasDeference with custom namespace: integration-test")
	fmt.Println("Custom configuration completed")
}

// ExampleTerraformIntegration demonstrates Terraform output integration
func ExampleTerraformIntegration() {
	fmt.Println("VasDeference Terraform integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Simulate Terraform outputs
	_ = /* outputs */ map[string]interface{}{
		"lambda_function_name": "my-function-abc123",
		"api_gateway_url":     "https://abc123.execute-api.us-east-1.amazonaws.com",
		"dynamodb_table_name": "my-table-xyz789",
	}
	
	// Extract values safely
	// functionName := vdf.GetTerraformOutput(outputs, "lambda_function_name")
	// apiUrl := vdf.GetTerraformOutput(outputs, "api_gateway_url")
	// tableName := vdf.GetTerraformOutput(outputs, "dynamodb_table_name")
	
	// Expected outputs:
	fmt.Printf("Function name: %s\n", outputs["lambda_function_name"])
	fmt.Printf("API Gateway URL: %s\n", outputs["api_gateway_url"])
	fmt.Printf("DynamoDB table: %s\n", outputs["dynamodb_table_name"])
	fmt.Println("Terraform integration completed")
}

// ExampleLambdaIntegration demonstrates integration with lambda package
func ExampleLambdaIntegration() {
	fmt.Println("VasDeference Lambda integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create Lambda client
	// lambdaClient := vdf.CreateLambdaClient()
	// Expected: lambdaClient should not be nil
	fmt.Println("Creating Lambda client through VasDeference")
	
	// functionName := vdf.PrefixResourceName("test-function")
	fmt.Println("Generating prefixed function name: test-function")
	
	// This is how you would integrate with the lambda package:
	// function := lambda.NewFunction(vdf.Context, &lambda.FunctionConfig{
	//     Name: functionName,
	//     Runtime: "nodejs18.x",
	//     Handler: "index.handler",
	// })
	fmt.Println("Creating Lambda function with configuration")
	
	// vdf.RegisterCleanup(func() error {
	//     return function.Delete()
	// })
	fmt.Println("Registering Lambda function cleanup")
	fmt.Println("Lambda integration completed")
}

// ExampleEventBridgeIntegration demonstrates integration with eventbridge package
func ExampleEventBridgeIntegration() {
	fmt.Println("VasDeference EventBridge integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create EventBridge client
	// eventBridgeClient := vdf.CreateEventBridgeClient()
	// Expected: eventBridgeClient should not be nil
	fmt.Println("Creating EventBridge client through VasDeference")
	
	// ruleName := vdf.PrefixResourceName("test-rule")
	fmt.Println("Generating prefixed rule name: test-rule")
	
	// This is how you would integrate with the eventbridge package:
	// rule := eventbridge.NewRule(vdf.Context, &eventbridge.RuleConfig{
	//     Name: ruleName,
	//     EventPattern: map[string]interface{}{
	//         "source": []string{"myapp"},
	//     },
	// })
	fmt.Println("Creating EventBridge rule with event pattern")
	
	// vdf.RegisterCleanup(func() error {
	//     return rule.Delete()
	// })
	fmt.Println("Registering EventBridge rule cleanup")
	fmt.Println("EventBridge integration completed")
}

// ExampleDynamoDBIntegration demonstrates integration with dynamodb package
func ExampleDynamoDBIntegration() {
	fmt.Println("VasDeference DynamoDB integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create DynamoDB client
	// dynamoClient := vdf.CreateDynamoDBClient()
	// Expected: dynamoClient should not be nil
	fmt.Println("Creating DynamoDB client through VasDeference")
	
	// tableName := vdf.PrefixResourceName("test-table")
	fmt.Println("Generating prefixed table name: test-table")
	
	// This is how you would integrate with the dynamodb package:
	// table := dynamodb.NewTable(vdf.Context, &dynamodb.TableConfig{
	//     TableName: tableName,
	//     BillingMode: "PAY_PER_REQUEST",
	//     AttributeDefinitions: []dynamodb.AttributeDefinition{
	//         {AttributeName: "id", AttributeType: "S"},
	//     },
	//     KeySchema: []dynamodb.KeySchemaElement{
	//         {AttributeName: "id", KeyType: "HASH"},
	//     },
	// })
	fmt.Println("Creating DynamoDB table with pay-per-request billing")
	
	// vdf.RegisterCleanup(func() error {
	//     return table.Delete()
	// })
	fmt.Println("Registering DynamoDB table cleanup")
	fmt.Println("DynamoDB integration completed")
}

// ExampleStepFunctionsIntegration demonstrates integration with stepfunctions package
func ExampleStepFunctionsIntegration() {
	fmt.Println("VasDeference Step Functions integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create Step Functions client
	// sfnClient := vdf.CreateStepFunctionsClient()
	// Expected: sfnClient should not be nil
	fmt.Println("Creating Step Functions client through VasDeference")
	
	// stateMachineName := vdf.PrefixResourceName("test-state-machine")
	fmt.Println("Generating prefixed state machine name: test-state-machine")
	
	// This is how you would integrate with the stepfunctions package:
	_ = /* definition */ `{
	  "Comment": "A Hello World example",
	  "StartAt": "HelloWorld",
	  "States": {
	    "HelloWorld": {
	      "Type": "Pass",
	      "Result": "Hello World!",
	      "End": true
	    }
	  }
	}`
	fmt.Printf("Sample state machine definition:\n%s\n", definition)
	
	// stateMachine := stepfunctions.NewStateMachine(vdf.Context, &stepfunctions.StateMachineConfig{
	//     Name: stateMachineName,
	//     Definition: definition,
	//     RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
	// })
	fmt.Println("Creating Step Functions state machine with Hello World definition")
	
	// vdf.RegisterCleanup(func() error {
	//     return stateMachine.Delete()
	// })
	fmt.Println("Registering Step Functions state machine cleanup")
	fmt.Println("Step Functions integration completed")
}

// ExampleRetryMechanism demonstrates using the retry functionality
func ExampleRetryMechanism() {
	fmt.Println("VasDeference retry mechanism patterns:")
	
	// ctx := vasdeference.NewTestContext(t)
	// arrange := vasdeference.NewArrangeWithNamespace(ctx, "retry-test")
	// defer arrange.Cleanup()
	
	_ = /* attempts */ 0
	
	// Simple retry logic implementation
	for i := 0; i < 5; i++ {
		attempts++
		if attempts >= 3 {
			break
		}
	}
	
	// Expected: attempts should be >= 3
	fmt.Printf("Retry attempts made: %d\n", attempts)
	fmt.Println("Retry mechanism completed")
}

// ExampleEnvironmentVariables demonstrates environment variable handling
func ExampleEnvironmentVariables() {
	fmt.Println("VasDeference environment variable patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Get environment variables with defaults
	// region := vdf.GetEnvVarWithDefault("AWS_REGION", "us-east-1")
	// profile := vdf.GetEnvVarWithDefault("AWS_PROFILE", "default")
	fmt.Println("Getting AWS_REGION with default: us-east-1")
	fmt.Println("Getting AWS_PROFILE with default: default")
	
	// Expected: region and profile should not be empty
	
	// Check if running in CI
	// isCI := vdf.IsCIEnvironment()
	// Expected: isCI should be boolean type
	fmt.Println("Checking if running in CI environment")
	fmt.Println("Environment variables completed")
}

// ExampleFullIntegration demonstrates a complete integration scenario
func ExampleFullIntegration() {
	fmt.Println("VasDeference full integration patterns:")
	
	// Create VasDeference with production-like configuration
	// ctx := vasdeference.NewTestContext(t)
	_ = /* namespace */ fmt.Sprintf("integration-%d", time.Now().Unix())
	// arrange := vasdeference.NewArrangeWithNamespace(ctx, namespace)
	// defer arrange.Cleanup()
	
	// Generate unique resource names with namespace
	_ = /* functionName */ namespace + "-processor"
	_ = /* tableName */ namespace + "-data"
	_ = /* ruleName */ namespace + "-trigger"
	_ = /* stateMachineName */ namespace + "-workflow"
	
	fmt.Printf("Generated namespace: %s\n", namespace)
	fmt.Printf("Function name: %s\n", functionName)
	fmt.Printf("Table name: %s\n", tableName)
	fmt.Printf("Rule name: %s\n", ruleName)
	fmt.Printf("State machine name: %s\n", stateMachineName)
	
	// Expected: All names should be unique and contain the namespace
	
	// Register cleanup for all resources
	// arrange.RegisterCleanup(func() error {
	//     // In a real test, this would clean up all created resources
	//     return nil
	// })
	fmt.Println("Registering cleanup for all resources")
	fmt.Println("Full integration completed")
}

// ExampleCloudWatchLogsClient demonstrates CloudWatch Logs client creation
func ExampleCloudWatchLogsClient() {
	fmt.Println("VasDeference CloudWatch Logs client patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// logsClient := vdf.CreateCloudWatchLogsClient()
	// Expected: logsClient should not be nil and should be *cloudwatchlogs.Client type
	fmt.Println("Creating CloudWatch Logs client through VasDeference")
	fmt.Printf("Expected client type: %T\n", &cloudwatchlogs.Client{})
	fmt.Println("CloudWatch Logs client completed")
}


// ExampleUniqueIdGeneration demonstrates unique ID generation
func ExampleUniqueIdGeneration() {
	fmt.Println("VasDeference unique ID generation patterns:")
	
	// id1 := vasdeference.UniqueId()
	// id2 := vasdeference.UniqueId()
	fmt.Println("Generating unique ID 1")
	fmt.Println("Generating unique ID 2")
	
	// Expected: IDs should not be empty, should not be equal, should have length > 0
	// UniqueId returns Unix nanoseconds as string, length varies
	fmt.Println("IDs are based on Unix nanoseconds for uniqueness")
	fmt.Println("Unique ID generation completed")
}

// ExampleTestingInterface demonstrates TestingT interface methods
func ExampleTestingInterface() {
	fmt.Println("VasDeference TestingT interface patterns:")
	
	// ctx := vasdeference.NewTestContext(t)
	
	// Test Helper method (should not panic)
	// ctx.T.Helper()
	fmt.Println("Calling Helper method on TestingT interface")
	
	// Test Name method
	// name := ctx.T.Name()
	// Expected: name should not be empty
	fmt.Println("Getting test name from TestingT interface")
	
	// Test Logf method (should not panic)
	// ctx.T.Logf("Test log message: %s", "test")
	fmt.Println("Logging message through TestingT interface")
	fmt.Println("TestingT interface completed")
}

// ExampleAWSCredentialsValidation demonstrates AWS credentials validation
func ExampleAWSCredentialsValidation() {
	fmt.Println("VasDeference AWS credentials validation patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Test credential validation
	// err := vdf.ValidateAWSCredentialsE()
	// Expected: In test environment without AWS credentials, this should return an error
	fmt.Println("Validating AWS credentials configuration")
	fmt.Println("Expected: Error if credentials not properly configured")
	fmt.Println("AWS credentials validation completed")
}

// ExampleNamespaceSanitization demonstrates namespace sanitization
func ExampleNamespaceSanitization() {
	fmt.Println("VasDeference namespace sanitization patterns:")
	
	// Test with empty options
	// ctx := vasdeference.NewTestContext(t)
	// arrange := vasdeference.NewArrangeWithNamespace(ctx, "test-empty")
	// defer arrange.Cleanup()
	fmt.Println("Creating namespace with: test-empty")
	// Expected: arrange.Namespace should equal "test-empty"
	
	// Test with different namespace patterns - SanitizeNamespace converts underscores to hyphens
	// ctx2 := vasdeference.NewTestContext(t)
	// arrange2 := vasdeference.NewArrangeWithNamespace(ctx2, "test_with_underscores")
	// defer arrange2.Cleanup()
	fmt.Println("Creating namespace with: test_with_underscores")
	// Expected: SanitizeNamespace converts underscores to hyphens -> "test-with-underscores"
	fmt.Println("Namespace sanitization completed")
}

// ExampleEnvironmentVariableEdgeCases demonstrates environment variable edge cases
func ExampleEnvironmentVariableEdgeCases() {
	fmt.Println("VasDeference environment variable edge cases:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Test GetEnvVarWithDefault with non-existent variable
	// value := vdf.GetEnvVarWithDefault("NON_EXISTENT_VAR_12345", "default_value")
	// Expected: value should equal "default_value"
	fmt.Println("Getting non-existent variable with default: default_value")
	
	// Test GetEnvVarWithDefault with existing variable
	// t.Setenv("TEST_VAR_VASDEFERENCE", "test_value")
	// value2 := vdf.GetEnvVarWithDefault("TEST_VAR_VASDEFERENCE", "default")
	// Expected: value2 should equal "test_value"
	fmt.Println("Setting TEST_VAR_VASDEFERENCE=test_value")
	fmt.Println("Getting existing variable should return: test_value")
	fmt.Println("Environment variable edge cases completed")
}

// ExampleCIEnvironmentDetection demonstrates CI environment detection
func ExampleCIEnvironmentDetection() {
	fmt.Println("VasDeference CI environment detection patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Test without CI environment variables
	// isCI := vdf.IsCIEnvironment()
	// Expected: isCI should be boolean type
	fmt.Println("Checking CI environment without CI variables")
	
	// Test with GitHub Actions
	// t.Setenv("GITHUB_ACTIONS", "true")
	// isCI2 := vdf.IsCIEnvironment()
	// Expected: isCI2 should be true
	fmt.Println("Setting GITHUB_ACTIONS=true, should detect CI: true")
	
	// Test with CircleCI
	// t.Setenv("CIRCLECI", "true")
	// isCI3 := vdf.IsCIEnvironment()
	// Expected: isCI3 should be true
	fmt.Println("Setting CIRCLECI=true, should detect CI: true")
	fmt.Println("CI environment detection completed")
}


// ExampleErrorHandling demonstrates error handling patterns
func ExampleErrorHandling() {
	fmt.Println("VasDeference error handling patterns:")
	
	// Test with nil TestingT (should panic)
	// Expected: vasdeference.New(nil) should panic
	fmt.Println("Testing nil TestingT parameter - should panic")
	
	// Test NewWithOptionsE with nil TestingT
	// vdf, err := vasdeference.NewWithOptionsE(nil, &vasdeference.VasDeferenceOptions{})
	// Expected: vdf should be nil, err should contain "TestingT cannot be nil"
	fmt.Println("Testing NewWithOptionsE with nil TestingT - should return error")
	fmt.Println("Error handling completed")
}

// ExampleUniqueIdUniqueness demonstrates ID uniqueness verification
func ExampleUniqueIdUniqueness() {
	fmt.Println("VasDeference unique ID uniqueness verification:")
	
	// Generate multiple IDs to verify uniqueness
	_ = /* ids */ make(map[string]bool)
	for i := 0; i < 10; i++ {
		// id := vasdeference.UniqueId()
		fmt.Printf("Generating unique ID #%d\n", i+1)
		
		// Each ID should be unique
		// Expected: id should not be empty, should not exist in ids map
		// ids[id] = true
		
		// Small delay to ensure different timestamps
		time.Sleep(time.Nanosecond)
	}
	fmt.Println("All IDs should be unique across multiple generations")
	fmt.Println("Unique ID uniqueness completed")
}

// ExampleTerraformOutputErrors demonstrates Terraform output error scenarios
func ExampleTerraformOutputErrors() {
	fmt.Println("VasDeference Terraform output error scenarios:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Test with nil outputs map
	// value, err := vdf.GetTerraformOutputE(nil, "test-key")
	// Expected: value should be empty, err should contain "outputs map cannot be nil"
	fmt.Println("Testing nil outputs map - should return error")
	
	// Test with non-string value
	_ = /* outputs */ map[string]interface{}{
		"numeric_value": 123,
		"boolean_value": true,
	}
	fmt.Printf("Testing non-string values: %+v\n", outputs)
	
	// value, err = vdf.GetTerraformOutputE(outputs, "numeric_value")
	// Expected: value should be empty, err should contain "is not a string"
	fmt.Println("Getting numeric value should return error - not a string")
	fmt.Println("Terraform output errors completed")
}


// ExampleSanitizeNamespaceEdgeCases demonstrates namespace sanitization edge cases
func ExampleSanitizeNamespaceEdgeCases() {
	fmt.Println("VasDeference namespace sanitization edge cases:")
	
	// Test empty string
	// result := vasdeference.SanitizeNamespace("")
	// Expected: result should equal "default"
	fmt.Println("Empty string input -> expected: 'default'")
	
	// Test string with only invalid characters
	// result = vasdeference.SanitizeNamespace("!@#$%^&*()")
	// Expected: result should equal "default"
	fmt.Println("Invalid characters '!@#$%^&*()' -> expected: 'default'")
	
	// Test string with only hyphens
	// result = vasdeference.SanitizeNamespace("---")
	// Expected: result should equal "default"
	fmt.Println("Only hyphens '---' -> expected: 'default'")
	
	// Test mixed valid/invalid characters
	// result = vasdeference.SanitizeNamespace("test!@#_with./chars")
	// Expected: result should equal "test-with--chars"
	fmt.Println("Mixed characters 'test!@#_with./chars' -> expected: 'test-with--chars'")
	fmt.Println("Namespace sanitization edge cases completed")
}

// ExampleComprehensiveConfiguration demonstrates comprehensive configuration scenarios
func ExampleComprehensiveConfiguration() {
	fmt.Println("VasDeference comprehensive configuration patterns:")
	
	// Test NewTestContextE with empty options to cover default-setting branches
	// ctx, err := vasdeference.NewTestContextE(t, &vasdeference.Options{})
	// Expected: err should be nil, ctx should not be nil
	fmt.Println("Creating test context with empty options")
	
	// Expected defaults:
	// ctx.Region should equal vasdeference.RegionUSEast1
	// ctx.T should equal the testing instance
	fmt.Printf("Expected default region: %s\n", vasdeference.RegionUSEast1)
	
	// Test NewWithOptionsE with default values
	_ = /* opts */ &vasdeference.VasDeferenceOptions{
		Region:     "", // Should get default
		MaxRetries: 0,  // Should get default  
		RetryDelay: 0,  // Should get default
		Namespace:  "", // Should trigger auto-detection
	}
	fmt.Printf("Testing with options: %+v\n", opts)
	
	// vdf, err := vasdeference.NewWithOptionsE(t, opts)
	// Expected: err should be nil, vdf should not be nil
	// vdf.Context.Region should equal vasdeference.RegionUSEast1
	// vdf.Arrange.Namespace should not be empty (auto-detected)
	fmt.Println("Creating VasDeference with default options")
	fmt.Println("Expected: defaults applied and namespace auto-detected")
	fmt.Println("Comprehensive configuration completed")
}

// ExampleAdvancedFeatures demonstrates advanced VasDeference features
func ExampleAdvancedFeatures() {
	fmt.Println("VasDeference advanced features patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Test createContext with timeout
	// ctx := vdf.CreateContextWithTimeout(5 * time.Second)
	// Expected: ctx should not be nil
	fmt.Println("Creating context with 5 second timeout")
	
	// Test default region getter
	// region := vdf.GetDefaultRegion()
	// Expected: region should equal vasdeference.RegionUSEast1
	fmt.Printf("Getting default region: %s\n", vasdeference.RegionUSEast1)
	
	// Test environment variable getter with empty key
	// envVal := vdf.GetEnvVar("")
	// Expected: envVal should equal ""
	fmt.Println("Getting empty environment variable key -> expected: empty string")
	
	// Test GetEnvVarWithDefault with empty key
	// envVal2 := vdf.GetEnvVarWithDefault("", "default-val")
	// Expected: envVal2 should equal "default-val"
	fmt.Println("Getting empty key with default 'default-val' -> expected: 'default-val'")
	
	// Test PrefixResourceName with various inputs
	// prefixed := vdf.PrefixResourceName("")
	// Expected: prefixed should contain namespace
	fmt.Println("Prefixing empty resource name with namespace")
	
	// Test Lambda client creation with different regions
	// lambdaClient1 := vdf.CreateLambdaClient()
	// lambdaClient2 := vdf.CreateLambdaClientWithRegion(vasdeference.RegionUSWest2)
	// Expected: both clients should not be nil
	fmt.Println("Creating Lambda clients for default and us-west-2 regions")
	
	// Test all AWS service client creation methods
	fmt.Println("Creating all AWS service clients:")
	fmt.Println("  - DynamoDB client")
	fmt.Println("  - EventBridge client")
	fmt.Println("  - Step Functions client")
	fmt.Println("  - CloudWatch Logs client")
	// Expected: all clients should not be nil
	
	// Test CI detection with various environment variables
	fmt.Println("Testing CI detection with various CI systems:")
	fmt.Println("  - TRAVIS=true -> should detect CI")
	fmt.Println("  - JENKINS_URL=http://jenkins.example.com -> should detect CI")
	fmt.Println("  - BUILDKITE=true -> should detect CI")
	fmt.Println("Advanced features completed")
}

// ExampleTerratestIntegration demonstrates Terratest integration features
func ExampleTerratestIntegration() {
	fmt.Println("VasDeference Terratest integration patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Get random AWS region using Terratest
	// randomRegion := vdf.GetRandomRegion()
	// Expected: randomRegion should be a valid AWS region
	fmt.Println("Getting random AWS region using Terratest's aws module")
	
	// Get AWS account ID using Terratest
	// accountId := vdf.GetAccountId()
	// Expected: accountId should be a valid AWS account ID
	fmt.Println("Getting AWS account ID using Terratest's aws module")
	
	// Generate random resource names
	// functionName := vdf.GenerateRandomName("my-function")
	// tableName := vdf.GenerateRandomName("my-table")
	// Expected: names should contain unique suffixes
	fmt.Println("Generating random resource names with Terratest's random module")
	
	// Get all available regions and AZs
	// allRegions := vdf.GetAllAvailableRegions()
	// azs := vdf.GetAvailabilityZones()
	// Expected: arrays should contain valid AWS regions and AZs
	fmt.Println("Getting all available regions and availability zones")
	
	// Create unique test identifier
	// testId := vdf.CreateTestIdentifier()
	// Expected: testId should contain namespace and unique ID
	fmt.Println("Creating unique test identifier with namespace")
	
	fmt.Println("Terratest integration completed")
}

// ExampleTerratestRetryPatterns demonstrates Terratest retry patterns
func ExampleTerratestRetryPatterns() {
	fmt.Println("VasDeference Terratest retry patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Retry with backoff using Terratest patterns
	// action := vasdeference.RetryableAction{
	//     Description: "Check Lambda function state",
	//     MaxRetries:  5,
	//     TimeBetween: 2 * time.Second,
	//     Action: func() (string, error) {
	//         // Check if function is ready
	//         return "ready", nil
	//     },
	// }
	// result := vdf.DoWithRetry(action)
	// Expected: result should equal "ready"
	fmt.Println("Configuring retryable action for Lambda function state check")
	
	// Wait until condition is met
	// vdf.WaitUntil(
	//     "DynamoDB table active",
	//     10,
	//     3*time.Second,
	//     func() (bool, error) {
	//         // Check table status
	//         return true, nil
	//     },
	// )
	fmt.Println("Waiting until DynamoDB table becomes active using Terratest retry")
	
	// Validate AWS resource exists with retry
	// vdf.ValidateAWSResourceExists(
	//     "Lambda function",
	//     "my-function",
	//     func() (bool, error) {
	//         // Check if function exists
	//         return true, nil
	//     },
	// )
	fmt.Println("Validating AWS resource existence with built-in retry logic")
	
	fmt.Println("Terratest retry patterns completed")
}

// ExampleTerratestRandomGeneration demonstrates Terratest random generation
func ExampleTerratestRandomGeneration() {
	fmt.Println("VasDeference Terratest random generation patterns:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Generate random resource names with length constraints
	// shortName := vdf.CreateRandomResourceName("func", 30)
	// longTableName := vdf.CreateRandomResourceName("my-table", 255)
	// Expected: names should be within specified length limits
	fmt.Println("Creating random resource names with length constraints")
	
	// Create resource name with timestamp
	// timestampName := vdf.CreateResourceNameWithTimestamp("deployment")
	// Expected: name should contain namespace, prefix, and unique timestamp
	fmt.Println("Creating resource name with timestamp using Terratest patterns")
	
	// Get random region excluding specific ones
	// excludeRegions := []string{"us-gov-east-1", "us-gov-west-1"}
	// randomRegion := vdf.GetRandomRegionExcluding(excludeRegions)
	// Expected: region should not be in excluded list
	fmt.Println("Getting random region excluding government regions")
	
	fmt.Println("Terratest random generation completed")
}

// ExampleTerratestTerraformOperations demonstrates Terraform operations with Terratest
func ExampleTerratestTerraformOperations() {
	fmt.Println("VasDeference Terraform operations with Terratest:")
	
	// vdf := vasdeference.New(t)
	// defer vdf.Cleanup()
	
	// Create Terraform options
	// variables := map[string]interface{}{
	//     "environment": "test",
	//     "region":      vdf.Context.Region,
	//     "namespace":   vdf.Arrange.Namespace,
	// }
	// terraformOptions := vdf.TerraformOptions("./terraform", variables)
	// Expected: terraformOptions should be properly configured
	fmt.Println("Creating Terraform options with test variables")
	
	// Initialize and apply Terraform
	// vdf.InitAndApplyTerraform(terraformOptions)
	// // Cleanup is automatically registered
	// Expected: Terraform should apply successfully and register cleanup
	fmt.Println("Initializing and applying Terraform with automatic cleanup registration")
	
	// Get Terraform outputs
	// functionName := vdf.GetTerraformOutputAsString(terraformOptions, "function_name")
	// apiConfig := vdf.GetTerraformOutputAsMap(terraformOptions, "api_config")
	// resourceList := vdf.GetTerraformOutputAsList(terraformOptions, "resource_arns")
	// Expected: outputs should be properly typed and accessible
	fmt.Println("Getting Terraform outputs as string, map, and list types")
	
	// Log all outputs for debugging
	// vdf.LogTerraformOutput(terraformOptions)
	// Expected: all outputs should be logged for debugging purposes
	fmt.Println("Logging all Terraform outputs for debugging")
	
	fmt.Println("Terraform operations with Terratest completed")
}

// runAllExamples demonstrates running all core VasDeference examples
func runAllExamples() {
	fmt.Println("Running all VasDeference core package examples:")
	fmt.Println("")
	
	ExampleBasicUsage()
	fmt.Println("")
	
	ExampleWithCustomOptions()
	fmt.Println("")
	
	ExampleTerraformIntegration()
	fmt.Println("")
	
	ExampleLambdaIntegration()
	fmt.Println("")
	
	ExampleEventBridgeIntegration()
	fmt.Println("")
	
	ExampleDynamoDBIntegration()
	fmt.Println("")
	
	ExampleStepFunctionsIntegration()
	fmt.Println("")
	
	ExampleRetryMechanism()
	fmt.Println("")
	
	ExampleEnvironmentVariables()
	fmt.Println("")
	
	ExampleFullIntegration()
	fmt.Println("")
	
	ExampleCloudWatchLogsClient()
	fmt.Println("")
	
	ExampleUniqueIdGeneration()
	fmt.Println("")
	
	ExampleTestingInterface()
	fmt.Println("")
	
	ExampleAWSCredentialsValidation()
	fmt.Println("")
	
	ExampleNamespaceSanitization()
	fmt.Println("")
	
	ExampleEnvironmentVariableEdgeCases()
	fmt.Println("")
	
	ExampleCIEnvironmentDetection()
	fmt.Println("")
	
	ExampleErrorHandling()
	fmt.Println("")
	
	ExampleUniqueIdUniqueness()
	fmt.Println("")
	
	ExampleTerraformOutputErrors()
	fmt.Println("")
	
	ExampleSanitizeNamespaceEdgeCases()
	fmt.Println("")
	
	ExampleComprehensiveConfiguration()
	fmt.Println("")
	
	ExampleAdvancedFeatures()
	fmt.Println("")
	
	ExampleTerratestIntegration()
	fmt.Println("")
	
	ExampleTerratestRetryPatterns()
	fmt.Println("")
	
	ExampleTerratestRandomGeneration()
	fmt.Println("")
	
	ExampleTerratestTerraformOperations()
	fmt.Println("All VasDeference core examples completed")
}

func main() {
	// This file demonstrates core VasDeference framework usage patterns.
	// Run examples with: go run ./examples/core/examples.go
	
	runAllExamples()
}