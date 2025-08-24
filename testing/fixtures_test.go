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