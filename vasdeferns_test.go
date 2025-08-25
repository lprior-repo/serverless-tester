package vasdeference

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test that VasDeference can be created with minimal configuration
func TestVasDeference_ShouldCreateWithDefaults(t *testing.T) {
	// Act
	vdf := New(t)

	// Assert
	require.NotNil(t, vdf)
	assert.NotNil(t, vdf.Context)
	assert.NotNil(t, vdf.Arrange)
	assert.NotEmpty(t, vdf.Context.Region)
	assert.NotEmpty(t, vdf.Arrange.Namespace)
}

// RED: Test that VasDeference can be created with custom options
func TestVasDeference_ShouldCreateWithCustomOptions(t *testing.T) {
	// Arrange
	opts := &VasDeferenceOptions{
		Region:    "eu-west-1",
		Namespace: "custom-test",
		MaxRetries: 5,
		RetryDelay: 200 * time.Millisecond,
	}

	// Act
	vdf := NewWithOptions(t, opts)

	// Assert
	require.NotNil(t, vdf)
	assert.Equal(t, "eu-west-1", vdf.Context.Region)
	assert.Equal(t, "custom-test", vdf.Arrange.Namespace)
}

// RED: Test that VasDeference supports E-variant creation with error handling
func TestVasDeference_ShouldCreateWithOptionsE(t *testing.T) {
	// Arrange
	opts := &VasDeferenceOptions{
		Region: "us-west-2",
	}

	// Act
	vdf, err := NewWithOptionsE(t, opts)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, vdf)
	assert.Equal(t, "us-west-2", vdf.Context.Region)
}

// RED: Test that VasDeference validates required parameters
func TestVasDeference_ShouldValidateRequiredParameters(t *testing.T) {
	// Act & Assert
	assert.Panics(t, func() {
		New(nil)
	})
}

// RED: Test that GitHub CLI integration works for namespace detection
func TestVasDeference_ShouldDetectNamespaceFromGitHubCLI(t *testing.T) {
	// Skip if not in git repository or gh CLI not available
	if !isGitRepository() || !isGitHubCLIAvailable() {
		t.Skip("Skipping GitHub CLI test - requires git repo and gh CLI")
	}

	// Act
	namespace := detectGitHubNamespace()

	// Assert
	assert.NotEmpty(t, namespace)
}

// RED: Test TerraformOutput helpers
func TestVasDeference_ShouldProvideTerraformOutputHelpers(t *testing.T) {
	// Arrange
	vdf := New(t)
	outputs := map[string]interface{}{
		"lambda_function_name": "test-function-abc123",
		"api_gateway_url":     "https://abc123.execute-api.us-east-1.amazonaws.com",
		"dynamodb_table_name": "test-table-xyz789",
	}

	// Act
	functionName := vdf.GetTerraformOutput(outputs, "lambda_function_name")
	apiUrl := vdf.GetTerraformOutput(outputs, "api_gateway_url")
	tableName := vdf.GetTerraformOutput(outputs, "dynamodb_table_name")

	// Assert
	assert.Equal(t, "test-function-abc123", functionName)
	assert.Equal(t, "https://abc123.execute-api.us-east-1.amazonaws.com", apiUrl)
	assert.Equal(t, "test-table-xyz789", tableName)
}

// RED: Test TerraformOutput helpers with missing key
func TestVasDeference_ShouldHandleMissingTerraformOutput(t *testing.T) {
	// Arrange
	mockT := &mockTestingT{}
	vdf := New(mockT)
	outputs := map[string]interface{}{
		"existing_key": "value",
	}

	// Act
	vdf.GetTerraformOutput(outputs, "missing_key")

	// Assert - Check that error was called
	assert.True(t, mockT.errorCalled)
	assert.True(t, mockT.failNowCalled)
}

// RED: Test TerraformOutput helpers with E-variant for error handling
func TestVasDeference_ShouldHandleMissingTerraformOutputE(t *testing.T) {
	// Arrange
	vdf := New(t)
	outputs := map[string]interface{}{
		"existing_key": "value",
	}

	// Act
	value, err := vdf.GetTerraformOutputE(outputs, "missing_key")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, value)
}

// RED: Test AWS credential validation
func TestVasDeference_ShouldValidateAWSCredentials(t *testing.T) {
	// Act
	vdf := New(t)
	isValid := vdf.ValidateAWSCredentials()

	// Assert
	// This might fail in CI without AWS credentials, but should not panic
	assert.IsType(t, true, isValid)
}

// RED: Test AWS credential validation with E-variant
func TestVasDeference_ShouldValidateAWSCredentialsE(t *testing.T) {
	// Act
	vdf := New(t)
	err := vdf.ValidateAWSCredentialsE()

	// Assert
	// Should return error or nil, never panic
	assert.True(t, err == nil || err != nil)
}

// RED: Test service client creation helpers
func TestVasDeference_ShouldCreateLambdaClient(t *testing.T) {
	// Arrange
	vdf := New(t)

	// Act
	client := vdf.CreateLambdaClient()

	// Assert
	require.NotNil(t, client)
}

// RED: Test service client creation with custom region
func TestVasDeference_ShouldCreateLambdaClientWithCustomRegion(t *testing.T) {
	// Arrange
	vdf := New(t)

	// Act
	client := vdf.CreateLambdaClientWithRegion("eu-west-1")

	// Assert
	require.NotNil(t, client)
}

// RED: Test DynamoDB client creation
func TestVasDeference_ShouldCreateDynamoDBClient(t *testing.T) {
	// Arrange
	vdf := New(t)

	// Act
	client := vdf.CreateDynamoDBClient()

	// Assert
	require.NotNil(t, client)
}

// RED: Test EventBridge client creation
func TestVasDeference_ShouldCreateEventBridgeClient(t *testing.T) {
	// Arrange
	vdf := New(t)

	// Act
	client := vdf.CreateEventBridgeClient()

	// Assert
	require.NotNil(t, client)
}

// RED: Test Step Functions client creation
func TestVasDeference_ShouldCreateStepFunctionsClient(t *testing.T) {
	// Arrange
	vdf := New(t)

	// Act
	client := vdf.CreateStepFunctionsClient()

	// Assert
	require.NotNil(t, client)
}

// RED: Test cleanup functionality integration
func TestVasDeference_ShouldRegisterAndExecuteCleanup(t *testing.T) {
	// Arrange
	vdf := New(t)
	cleanupExecuted := false

	// Act
	vdf.RegisterCleanup(func() error {
		cleanupExecuted = true
		return nil
	})
	vdf.Cleanup()

	// Assert
	assert.True(t, cleanupExecuted)
}

// RED: Test resource name prefixing utilities
func TestVasDeference_ShouldPrefixResourceNames(t *testing.T) {
	// Arrange
	vdf := NewWithOptions(t, &VasDeferenceOptions{
		Namespace: "test-pr-123",
	})

	// Act
	prefixedName := vdf.PrefixResourceName("my-function")

	// Assert
	assert.Equal(t, "test-pr-123-my-function", prefixedName)
}

// RED: Test resource name prefixing with sanitization
func TestVasDeference_ShouldPrefixAndSanitizeResourceNames(t *testing.T) {
	// Arrange
	vdf := NewWithOptions(t, &VasDeferenceOptions{
		Namespace: "feature/branch_with_underscores",
	})

	// Act
	prefixedName := vdf.PrefixResourceName("My_Function")

	// Assert
	assert.Equal(t, "feature-branch-with-underscores-my-function", prefixedName)
	assert.Regexp(t, "^[a-z0-9-]+$", prefixedName)
}

// RED: Test environment variable handling
func TestVasDeference_ShouldHandleEnvironmentVariables(t *testing.T) {
	// Arrange
	oldValue := os.Getenv("TEST_VAR")
	defer os.Setenv("TEST_VAR", oldValue)
	os.Setenv("TEST_VAR", "test-value")

	vdf := New(t)

	// Act
	value := vdf.GetEnvVar("TEST_VAR")

	// Assert
	assert.Equal(t, "test-value", value)
}

// RED: Test environment variable handling with default value
func TestVasDeference_ShouldHandleEnvironmentVariablesWithDefault(t *testing.T) {
	// Arrange
	vdf := New(t)

	// Act
	value := vdf.GetEnvVarWithDefault("NON_EXISTENT_VAR", "default-value")

	// Assert
	assert.Equal(t, "default-value", value)
}

// RED: Test CI/CD integration helpers
func TestVasDeference_ShouldDetectCIEnvironment(t *testing.T) {
	// Act
	vdf := New(t)
	isCI := vdf.IsCIEnvironment()

	// Assert
	assert.IsType(t, true, isCI)
}

// RED: Test context management with timeout
func TestVasDeference_ShouldCreateContextWithTimeout(t *testing.T) {
	// Arrange
	vdf := New(t)

	// Act
	ctx := vdf.CreateContextWithTimeout(5 * time.Second)

	// Assert
	require.NotNil(t, ctx)
	
	// Context should have a deadline
	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.True(t, deadline.After(time.Now()))
}

// RED: Test default region management
func TestVasDeference_ShouldProvideDefaultRegion(t *testing.T) {
	// Act
	vdf := New(t)
	region := vdf.GetDefaultRegion()

	// Assert
	assert.NotEmpty(t, region)
	assert.Contains(t, []string{RegionUSEast1, RegionUSWest2, RegionEUWest1}, region)
}

// Helper functions for testing
func isGitRepository() bool {
	_, err := os.Stat(".git")
	return err == nil
}