package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// RED: Test that CreateTerraformOptions creates proper Terraform options
func TestCreateTerraformOptions_ShouldCreateValidOptions(t *testing.T) {
	// Arrange
	terraformDir := "/path/to/terraform"
	vars := map[string]interface{}{
		"region": "us-west-2",
		"name":   "test-resource",
	}
	
	// Act
	options := CreateTerraformOptions(terraformDir, vars)
	
	// Assert
	assert.Equal(t, terraformDir, options.TerraformDir, "Should set correct Terraform directory")
	assert.Equal(t, vars, options.Vars, "Should set correct variables")
	assert.NotEmpty(t, options.Id, "Should generate unique ID")
}

// RED: Test that CreateTerraformOptionsWithBackend creates options with backend config
func TestCreateTerraformOptionsWithBackend_ShouldCreateOptionsWithBackend(t *testing.T) {
	// Arrange
	terraformDir := "/path/to/terraform"
	vars := map[string]interface{}{
		"region": "us-west-2",
	}
	backendConfig := map[string]interface{}{
		"bucket": "my-terraform-state",
		"key":    "path/to/state.tfstate",
		"region": "us-west-2",
	}
	
	// Act
	options := CreateTerraformOptionsWithBackend(terraformDir, vars, backendConfig)
	
	// Assert
	assert.Equal(t, terraformDir, options.TerraformDir, "Should set correct Terraform directory")
	assert.Equal(t, vars, options.Vars, "Should set correct variables")
	assert.Equal(t, backendConfig, options.BackendConfig, "Should set correct backend config")
	assert.NotEmpty(t, options.Id, "Should generate unique ID")
}

// RED: Test that AWSTestOptions creates proper AWS test options
func TestAWSTestOptions_ShouldCreateValidOptions(t *testing.T) {
	// Arrange
	region := "us-east-1"
	profile := "test-profile"
	
	// Act
	options := AWSTestOptions(region, profile)
	
	// Assert
	assert.Equal(t, region, options.Region, "Should set correct region")
	assert.Equal(t, profile, options.Profile, "Should set correct profile")
	assert.NotEmpty(t, options.TestName, "Should generate test name")
}

// RED: Test that GenerateTestResourceName creates valid resource names
func TestGenerateTestResourceName_ShouldCreateValidResourceNames(t *testing.T) {
	// Arrange
	testCases := []struct {
		prefix     string
		resourceType string
	}{
		{"test", "lambda"},
		{"integration", "dynamodb"},
		{"e2e", "s3"},
		{"unit", "iam"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.prefix+"-"+tc.resourceType, func(t *testing.T) {
			// Act
			name := GenerateTestResourceName(tc.prefix, tc.resourceType)
			
			// Assert
			assert.Contains(t, name, tc.prefix, "Should contain prefix")
			assert.Contains(t, name, tc.resourceType, "Should contain resource type")
			assert.True(t, ValidateAWSResourceName(name), "Should be valid AWS resource name")
			assert.LessOrEqual(t, len(name), 64, "Should be reasonable length for AWS resources")
		})
	}
}

// RED: Test that GenerateTestResourceName creates unique names
func TestGenerateTestResourceName_ShouldCreateUniqueNames(t *testing.T) {
	// Arrange
	prefix := "test"
	resourceType := "lambda"
	
	// Act
	name1 := GenerateTestResourceName(prefix, resourceType)
	name2 := GenerateTestResourceName(prefix, resourceType)
	
	// Assert
	assert.NotEqual(t, name1, name2, "Should generate unique names")
}

// RED: Test that TestDataPath creates valid data paths
func TestTestDataPath_ShouldCreateValidPaths(t *testing.T) {
	// Arrange
	subPaths := []string{"terraform", "fixtures", "test.tf"}
	
	// Act
	path := TestDataPath(subPaths...)
	
	// Assert
	assert.Contains(t, path, "testdata", "Should contain testdata directory")
	assert.Contains(t, path, "terraform", "Should contain first subpath")
	assert.Contains(t, path, "fixtures", "Should contain second subpath")
	assert.Contains(t, path, "test.tf", "Should contain third subpath")
}

// RED: Test that LoadTestConfig loads and validates configuration
func TestLoadTestConfig_ShouldLoadValidConfiguration(t *testing.T) {
	// Arrange - create a temporary config for testing
	configData := TestConfig{
		AWSRegion:  "us-west-2",
		AWSProfile: "default",
		Environment: "test",
		ResourcePrefix: "terratest",
		Tags: map[string]string{
			"Environment": "test",
			"Project":     "terratest",
		},
	}
	
	// Act
	config := LoadTestConfig(configData)
	
	// Assert
	assert.Equal(t, "us-west-2", config.AWSRegion, "Should load correct AWS region")
	assert.Equal(t, "default", config.AWSProfile, "Should load correct AWS profile")
	assert.Equal(t, "test", config.Environment, "Should load correct environment")
	assert.Equal(t, "terratest", config.ResourcePrefix, "Should load correct resource prefix")
	assert.NotEmpty(t, config.Tags, "Should load tags")
}

// RED: Test that ValidateTestConfig validates configuration properly
func TestValidateTestConfig_ShouldValidateConfiguration(t *testing.T) {
	// Arrange
	validConfig := TestConfig{
		AWSRegion:  "us-west-2",
		AWSProfile: "default",
		Environment: "test",
		ResourcePrefix: "terratest",
		Tags: map[string]string{
			"Environment": "test",
		},
	}
	
	invalidConfig := TestConfig{
		AWSRegion:  "", // Missing required field
		AWSProfile: "default",
		Environment: "test",
	}
	
	// Act & Assert
	err := ValidateTestConfig(validConfig)
	assert.NoError(t, err, "Valid config should pass validation")
	
	err = ValidateTestConfig(invalidConfig)
	assert.Error(t, err, "Invalid config should fail validation")
	assert.Contains(t, err.Error(), "AWSRegion", "Error should mention missing AWSRegion")
}

// RED: Test that ApplyTestConfigDefaults sets reasonable defaults
func TestApplyTestConfigDefaults_ShouldSetReasonableDefaults(t *testing.T) {
	// Arrange
	config := TestConfig{
		AWSRegion: "us-west-2", // Required field set
	}
	
	// Act
	result := ApplyTestConfigDefaults(config)
	
	// Assert
	assert.Equal(t, "us-west-2", result.AWSRegion, "Should preserve set values")
	assert.Equal(t, "default", result.AWSProfile, "Should set default profile")
	assert.Equal(t, "test", result.Environment, "Should set default environment")
	assert.Equal(t, "terratest", result.ResourcePrefix, "Should set default resource prefix")
	assert.NotEmpty(t, result.Tags, "Should set default tags")
}

// RED: Test that CreateTestTags generates proper test tags
func TestCreateTestTags_ShouldGenerateProperTestTags(t *testing.T) {
	// Arrange
	testName := "TestMyFunction"
	environment := "integration"
	
	// Act
	tags := CreateTestTags(testName, environment)
	
	// Assert
	assert.Equal(t, testName, tags["TestName"], "Should set test name tag")
	assert.Equal(t, environment, tags["Environment"], "Should set environment tag")
	assert.Equal(t, "terratest", tags["CreatedBy"], "Should set created by tag")
	assert.NotEmpty(t, tags["Timestamp"], "Should set timestamp tag")
}

// RED: Test that MergeTestTags properly merges tag maps
func TestMergeTestTags_ShouldMergeTagsProperly(t *testing.T) {
	// Arrange
	baseTags := map[string]string{
		"Environment": "test",
		"Project":     "myproject",
	}
	
	additionalTags := map[string]string{
		"Environment": "staging", // Override existing
		"Owner":       "team-a",   // New tag
	}
	
	// Act
	result := MergeTestTags(baseTags, additionalTags)
	
	// Assert
	assert.Equal(t, "staging", result["Environment"], "Should override existing tags")
	assert.Equal(t, "myproject", result["Project"], "Should preserve base tags")
	assert.Equal(t, "team-a", result["Owner"], "Should add new tags")
	assert.Len(t, result, 3, "Should have correct number of tags")
}

// RED: Test that test fixture structure validation works
func TestTestFixture_ShouldValidateStructure(t *testing.T) {
	// Arrange
	fixture := TestFixture{
		Name:        "dynamodb-table",
		Description: "DynamoDB table for testing",
		TerraformDir: "/path/to/terraform",
		Variables: map[string]interface{}{
			"table_name": "test-table",
		},
		ExpectedOutputs: []string{"table_arn", "table_name"},
		Cleanup: func() error {
			return nil
		},
	}
	
	// Act
	isValid := ValidateTestFixture(fixture)
	
	// Assert
	assert.True(t, isValid, "Valid fixture should pass validation")
}

// RED: Test that test fixture validation catches missing required fields
func TestTestFixture_ShouldCatchMissingRequiredFields(t *testing.T) {
	// Arrange
	fixture := TestFixture{
		Name: "", // Missing required name
		TerraformDir: "/path/to/terraform",
	}
	
	// Act
	isValid := ValidateTestFixture(fixture)
	
	// Assert
	assert.False(t, isValid, "Invalid fixture should fail validation")
}