package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test that test data fixtures can be created
func TestFixtures_ShouldProvideCommonTestData(t *testing.T) {
	// Act
	fixtures := NewTestFixtures()
	
	// Assert
	require.NotNil(t, fixtures)
	assert.NotEmpty(t, fixtures.GetLambdaFunctionNames())
	assert.NotEmpty(t, fixtures.GetDynamoDBTableNames())
	assert.NotEmpty(t, fixtures.GetEventBridgeRuleNames())
	assert.NotEmpty(t, fixtures.GetStepFunctionNames())
}

// RED: Test that fixtures provide valid AWS resource names
func TestFixtures_ShouldProvideValidAWSResourceNames(t *testing.T) {
	// Arrange
	fixtures := NewTestFixtures()
	
	// Act
	lambdaName := fixtures.GetLambdaFunctionNames()[0]
	tableName := fixtures.GetDynamoDBTableNames()[0]
	ruleName := fixtures.GetEventBridgeRuleNames()[0]
	
	// Assert - All names should follow AWS naming conventions
	assert.Regexp(t, "^[a-zA-Z0-9-_]+$", lambdaName)
	assert.Regexp(t, "^[a-zA-Z0-9-_\\.]+$", tableName)
	assert.Regexp(t, "^[a-zA-Z0-9-_\\.]+$", ruleName)
}

// RED: Test that fixtures provide sample AWS events
func TestFixtures_ShouldProvideSampleAWSEvents(t *testing.T) {
	// Arrange
	fixtures := NewTestFixtures()
	
	// Act
	s3Event := fixtures.GetSampleS3Event()
	sqsEvent := fixtures.GetSampleSQSEvent()
	apiGatewayEvent := fixtures.GetSampleAPIGatewayEvent()
	
	// Assert
	assert.Contains(t, s3Event, "eventName")
	assert.Contains(t, s3Event, "s3")
	
	assert.Contains(t, sqsEvent, "Records")
	assert.Contains(t, sqsEvent, "messageId")
	
	assert.Contains(t, apiGatewayEvent, "httpMethod")
	assert.Contains(t, apiGatewayEvent, "path")
}

// RED: Test that fixtures provide valid Terraform outputs structure
func TestFixtures_ShouldProvideTerraformOutputs(t *testing.T) {
	// Arrange
	fixtures := NewTestFixtures()
	
	// Act
	outputs := fixtures.GetSampleTerraformOutputs()
	
	// Assert
	require.NotNil(t, outputs)
	assert.IsType(t, map[string]interface{}{}, outputs)
	
	// Should contain common outputs
	_, hasLambdaArn := outputs["lambda_function_arn"]
	_, hasApiUrl := outputs["api_gateway_url"]
	_, hasTableName := outputs["dynamodb_table_name"]
	
	assert.True(t, hasLambdaArn)
	assert.True(t, hasApiUrl)
	assert.True(t, hasTableName)
}

// RED: Test that fixtures can generate data with custom namespace
func TestFixtures_ShouldSupportCustomNamespace(t *testing.T) {
	// Arrange
	namespace := "custom-test-ns"
	
	// Act
	fixtures := NewTestFixturesWithNamespace(namespace)
	lambdaNames := fixtures.GetLambdaFunctionNames()
	
	// Assert
	require.NotEmpty(t, lambdaNames)
	// Names should be prefixed with namespace
	assert.Contains(t, lambdaNames[0], namespace)
}

// RED: Test that GetSampleEnvironmentVariables provides expected environment variables
func TestFixtures_ShouldProvideExpectedEnvironmentVariables(t *testing.T) {
	// Arrange
	fixtures := NewTestFixturesWithNamespace("test-env")
	
	// Act
	envVars := fixtures.GetSampleEnvironmentVariables()
	
	// Assert
	assert.NotEmpty(t, envVars, "Environment variables should not be empty")
	assert.Equal(t, "us-east-1", envVars["AWS_REGION"], "Should have correct AWS region")
	assert.Equal(t, "test", envVars["ENVIRONMENT"], "Should have correct environment")
	assert.Equal(t, "DEBUG", envVars["LOG_LEVEL"], "Should have correct log level")
	assert.Contains(t, envVars["DYNAMODB_TABLE_NAME"], "test-env-test-table", "Should have namespaced table name")
	assert.Contains(t, envVars["API_BASE_URL"], "test-env-api", "Should have namespaced API URL")
	assert.Contains(t, envVars["QUEUE_URL"], "test-env-queue", "Should have namespaced queue URL")
	assert.Contains(t, envVars["BUCKET_NAME"], "test-env-test-bucket", "Should have namespaced bucket name")
}

// RED: Test that GetSampleEnvironmentVariables uses namespace correctly
func TestFixtures_ShouldNamespaceEnvironmentVariables(t *testing.T) {
	// Arrange
	namespace := "custom-namespace"
	fixtures := NewTestFixturesWithNamespace(namespace)
	
	// Act
	envVars := fixtures.GetSampleEnvironmentVariables()
	
	// Assert
	assert.Contains(t, envVars["DYNAMODB_TABLE_NAME"], namespace, "Table name should include namespace")
	assert.Contains(t, envVars["API_BASE_URL"], namespace, "API URL should include namespace")
	assert.Contains(t, envVars["QUEUE_URL"], namespace, "Queue URL should include namespace")
	assert.Contains(t, envVars["BUCKET_NAME"], namespace, "Bucket name should include namespace")
}

// RED: Test that GetSampleLambdaContext provides expected Lambda context data
func TestFixtures_ShouldProvideExpectedLambdaContext(t *testing.T) {
	// Arrange
	fixtures := NewTestFixturesWithNamespace("test-lambda")
	
	// Act
	context := fixtures.GetSampleLambdaContext()
	
	// Assert
	assert.NotEmpty(t, context, "Lambda context should not be empty")
	assert.Contains(t, context["functionName"], "test-lambda-handler", "Should have namespaced function name")
	assert.Equal(t, "$LATEST", context["functionVersion"], "Should have correct function version")
	assert.Contains(t, context["invokedFunctionArn"], "test-lambda-handler", "Should have correct ARN")
	assert.Equal(t, "128", context["memoryLimitInMB"], "Should have correct memory limit")
	assert.Equal(t, 30000, context["remainingTimeInMillis"], "Should have correct remaining time")
	assert.Contains(t, context["logGroupName"], "test-lambda-handler", "Should have correct log group")
	assert.NotEmpty(t, context["logStreamName"], "Should have log stream name")
	assert.NotEmpty(t, context["awsRequestId"], "Should have request ID")
}

// RED: Test that GetSampleLambdaContext uses namespace correctly
func TestFixtures_ShouldNamespaceLambdaContext(t *testing.T) {
	// Arrange
	namespace := "custom-lambda"
	fixtures := NewTestFixturesWithNamespace(namespace)
	
	// Act
	context := fixtures.GetSampleLambdaContext()
	
	// Assert
	functionName := context["functionName"].(string)
	invokedFunctionArn := context["invokedFunctionArn"].(string)
	logGroupName := context["logGroupName"].(string)
	
	assert.Contains(t, functionName, namespace, "Function name should include namespace")
	assert.Contains(t, invokedFunctionArn, namespace, "Function ARN should include namespace")
	assert.Contains(t, logGroupName, namespace, "Log group name should include namespace")
}