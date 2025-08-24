package sfx

import (
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVasDeference_ExampleBasicUsage demonstrates the basic usage of VasDeference
func TestVasDeference_ExampleBasicUsage(t *testing.T) {
	// Create VasDeference instance with automatic configuration
	vdf := New(t)
	defer vdf.Cleanup()
	
	// Use the configured AWS clients
	lambdaClient := vdf.CreateLambdaClient()
	require.NotNil(t, lambdaClient)
	
	// Use namespace for resource naming
	resourceName := vdf.PrefixResourceName("test-function")
	assert.Contains(t, resourceName, "test-function")
	
	// Register cleanup for resources
	vdf.RegisterCleanup(func() error {
		// Clean up your resources here
		return nil
	})
}

// TestVasDeference_ExampleWithCustomOptions demonstrates VasDeference with custom configuration
func TestVasDeference_ExampleWithCustomOptions(t *testing.T) {
	// Create with custom options
	vdf := NewWithOptions(t, &VasDefernOptions{
		Region:     "eu-west-1",
		Namespace:  "integration-test",
		MaxRetries: 5,
		RetryDelay: 200 * time.Millisecond,
	})
	defer vdf.Cleanup()
	
	assert.Equal(t, "eu-west-1", vdf.GetDefaultRegion())
	assert.Equal(t, "integration-test", vdf.Arrange.Namespace)
}

// TestVasDeference_ExampleTerraformIntegration demonstrates Terraform output integration
func TestVasDeference_ExampleTerraformIntegration(t *testing.T) {
	vdf := New(t)
	defer vdf.Cleanup()
	
	// Simulate Terraform outputs
	outputs := map[string]interface{}{
		"lambda_function_name": "my-function-abc123",
		"api_gateway_url":     "https://abc123.execute-api.us-east-1.amazonaws.com",
		"dynamodb_table_name": "my-table-xyz789",
	}
	
	// Extract values safely
	functionName := vdf.GetTerraformOutput(outputs, "lambda_function_name")
	apiUrl := vdf.GetTerraformOutput(outputs, "api_gateway_url")
	tableName := vdf.GetTerraformOutput(outputs, "dynamodb_table_name")
	
	// Use extracted values in tests
	assert.Equal(t, "my-function-abc123", functionName)
	assert.Equal(t, "https://abc123.execute-api.us-east-1.amazonaws.com", apiUrl)
	assert.Equal(t, "my-table-xyz789", tableName)
}

// TestVasDeference_ExampleLambdaIntegration demonstrates integration with lambda package
func TestVasDeference_ExampleLambdaIntegration(t *testing.T) {
	vdf := New(t)
	defer vdf.Cleanup()
	
	// Create Lambda client
	lambdaClient := vdf.CreateLambdaClient()
	require.NotNil(t, lambdaClient)
	
	functionName := vdf.PrefixResourceName("test-function")
	
	// This is how you would integrate with the lambda package:
	// (Commented out as implementation details are in lambda package)
	
	// function := lambda.NewFunction(vdf.Context, &lambda.FunctionConfig{
	//     Name: functionName,
	//     Runtime: "nodejs18.x",
	//     Handler: "index.handler",
	// })
	
	// vdf.RegisterCleanup(func() error {
	//     return function.Delete()
	// })
	
	assert.NotEmpty(t, functionName)
}

// TestVasDeference_ExampleEventBridgeIntegration demonstrates integration with eventbridge package
func TestVasDeference_ExampleEventBridgeIntegration(t *testing.T) {
	vdf := New(t)
	defer vdf.Cleanup()
	
	// Create EventBridge client
	eventBridgeClient := vdf.CreateEventBridgeClient()
	require.NotNil(t, eventBridgeClient)
	
	ruleName := vdf.PrefixResourceName("test-rule")
	
	// This is how you would integrate with the eventbridge package:
	// (Commented out as implementation details are in eventbridge package)
	
	// rule := eventbridge.NewRule(vdf.Context, &eventbridge.RuleConfig{
	//     Name: ruleName,
	//     EventPattern: map[string]interface{}{
	//         "source": []string{"myapp"},
	//     },
	// })
	
	// vdf.RegisterCleanup(func() error {
	//     return rule.Delete()
	// })
	
	assert.NotEmpty(t, ruleName)
}

// TestVasDeference_ExampleDynamoDBIntegration demonstrates integration with dynamodb package
func TestVasDeference_ExampleDynamoDBIntegration(t *testing.T) {
	vdf := New(t)
	defer vdf.Cleanup()
	
	// Create DynamoDB client
	dynamoClient := vdf.CreateDynamoDBClient()
	require.NotNil(t, dynamoClient)
	
	tableName := vdf.PrefixResourceName("test-table")
	
	// This is how you would integrate with the dynamodb package:
	// (Commented out as implementation details are in dynamodb package)
	
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
	
	// vdf.RegisterCleanup(func() error {
	//     return table.Delete()
	// })
	
	assert.NotEmpty(t, tableName)
}

// TestVasDeference_ExampleStepFunctionsIntegration demonstrates integration with stepfunctions package
func TestVasDeference_ExampleStepFunctionsIntegration(t *testing.T) {
	vdf := New(t)
	defer vdf.Cleanup()
	
	// Create Step Functions client
	sfnClient := vdf.CreateStepFunctionsClient()
	require.NotNil(t, sfnClient)
	
	stateMachineName := vdf.PrefixResourceName("test-state-machine")
	
	// This is how you would integrate with the stepfunctions package:
	// (Commented out as implementation details are in stepfunctions package)
	
	// definition := `{
	//   "Comment": "A Hello World example",
	//   "StartAt": "HelloWorld",
	//   "States": {
	//     "HelloWorld": {
	//       "Type": "Pass",
	//       "Result": "Hello World!",
	//       "End": true
	//     }
	//   }
	// }`
	
	// stateMachine := stepfunctions.NewStateMachine(vdf.Context, &stepfunctions.StateMachineConfig{
	//     Name: stateMachineName,
	//     Definition: definition,
	//     RoleArn: "arn:aws:iam::123456789012:role/StepFunctionsRole",
	// })
	
	// vdf.RegisterCleanup(func() error {
	//     return stateMachine.Delete()
	// })
	
	assert.NotEmpty(t, stateMachineName)
}

// TestVasDeference_ExampleRetryMechanism demonstrates using the retry functionality
func TestVasDeference_ExampleRetryMechanism(t *testing.T) {
	vdf := New(t)
	defer vdf.Cleanup()
	
	attempts := 0
	
	// Use built-in retry mechanism
	err := Retry(func() error {
		attempts++
		if attempts < 3 {
			return assert.AnError
		}
		return nil
	}, 5)
	
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

// TestVasDeference_ExampleEnvironmentVariables demonstrates environment variable handling
func TestVasDeference_ExampleEnvironmentVariables(t *testing.T) {
	vdf := New(t)
	defer vdf.Cleanup()
	
	// Get environment variables with defaults
	region := vdf.GetEnvVarWithDefault("AWS_REGION", "us-east-1")
	profile := vdf.GetEnvVarWithDefault("AWS_PROFILE", "default")
	
	assert.NotEmpty(t, region)
	assert.NotEmpty(t, profile)
	
	// Check if running in CI
	isCI := vdf.IsCIEnvironment()
	assert.IsType(t, true, isCI)
}

// TestVasDeference_ExampleFullIntegration demonstrates a complete test scenario
func TestVasDeference_ExampleFullIntegration(t *testing.T) {
	// Create VasDeference with production-like configuration
	vdf := NewWithOptions(t, &VasDefernOptions{
		Region:     SelectRegion(), // Random region for testing
		Namespace:  "integration-" + UniqueId(), // Unique namespace
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
	})
	defer vdf.Cleanup()
	
	// Validate AWS credentials - skip if not configured
	if !vdf.ValidateAWSCredentials() {
		t.Skip("Skipping test - AWS credentials not configured")
	}
	
	// Create clients for all services
	lambdaClient := vdf.CreateLambdaClient()
	dynamoClient := vdf.CreateDynamoDBClient()
	eventBridgeClient := vdf.CreateEventBridgeClient()
	sfnClient := vdf.CreateStepFunctionsClient()
	
	require.NotNil(t, lambdaClient)
	require.NotNil(t, dynamoClient)
	require.NotNil(t, eventBridgeClient)
	require.NotNil(t, sfnClient)
	
	// Generate unique resource names
	functionName := vdf.PrefixResourceName("processor")
	tableName := vdf.PrefixResourceName("data")
	ruleName := vdf.PrefixResourceName("trigger")
	stateMachineName := vdf.PrefixResourceName("workflow")
	
	// All names should be unique and sanitized
	assert.NotEmpty(t, functionName)
	assert.NotEmpty(t, tableName)
	assert.NotEmpty(t, ruleName)
	assert.NotEmpty(t, stateMachineName)
	
	// All names should contain the namespace
	assert.Contains(t, functionName, vdf.Arrange.Namespace)
	assert.Contains(t, tableName, vdf.Arrange.Namespace)
	assert.Contains(t, ruleName, vdf.Arrange.Namespace)
	assert.Contains(t, stateMachineName, vdf.Arrange.Namespace)
	
	// Register cleanup for all resources
	vdf.RegisterCleanup(func() error {
		// In a real test, this would clean up all created resources
		return nil
	})
}