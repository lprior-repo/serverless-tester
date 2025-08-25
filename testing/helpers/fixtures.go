package helpers

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
)

// TerraformOptions represents Terraform options for testing
// Following Terratest patterns for Terraform configuration
type TerraformOptions struct {
	TerraformDir  string
	Vars          map[string]interface{}
	BackendConfig map[string]interface{}
	Id            string
}

// AWSOptions represents AWS-specific options for testing
// Following Terratest patterns for AWS configuration
type AWSOptions struct {
	Region   string
	Profile  string
	TestName string
}

// TestConfig represents test configuration following Terratest conventions
type TestConfig struct {
	AWSRegion      string
	AWSProfile     string
	Environment    string
	ResourcePrefix string
	Tags           map[string]string
}

// TestFixture represents a test fixture following Terratest patterns
type TestFixture struct {
	Name            string
	Description     string
	TerraformDir    string
	Variables       map[string]interface{}
	ExpectedOutputs []string
	Cleanup         func() error
}

// CreateTerraformOptions creates Terraform options for testing
// Following Terratest patterns for Terraform configuration
func CreateTerraformOptions(terraformDir string, vars map[string]interface{}) TerraformOptions {
	uniqueId := UniqueID()
	
	slog.Info("Creating Terraform options",
		"terraformDir", terraformDir,
		"uniqueId", uniqueId,
		"varsCount", len(vars),
	)
	
	return TerraformOptions{
		TerraformDir: terraformDir,
		Vars:         vars,
		Id:           uniqueId,
	}
}

// CreateTerraformOptionsWithBackend creates Terraform options with backend configuration
// Following Terratest patterns for Terraform with remote state
func CreateTerraformOptionsWithBackend(terraformDir string, vars map[string]interface{}, backendConfig map[string]interface{}) TerraformOptions {
	uniqueId := UniqueID()
	
	slog.Info("Creating Terraform options with backend",
		"terraformDir", terraformDir,
		"uniqueId", uniqueId,
		"varsCount", len(vars),
		"backendConfigCount", len(backendConfig),
	)
	
	return TerraformOptions{
		TerraformDir:  terraformDir,
		Vars:          vars,
		BackendConfig: backendConfig,
		Id:            uniqueId,
	}
}

// AWSTestOptions creates AWS options for testing
// Following Terratest patterns for AWS configuration
func AWSTestOptions(region, profile string) AWSOptions {
	testName := fmt.Sprintf("terratest-%s", RandomString(8))
	
	slog.Info("Creating AWS test options",
		"region", region,
		"profile", profile,
		"testName", testName,
	)
	
	return AWSOptions{
		Region:   region,
		Profile:  profile,
		TestName: testName,
	}
}

// GenerateTestResourceName generates a unique resource name for testing
// Following Terratest patterns for resource naming
func GenerateTestResourceName(prefix, resourceType string) string {
	timestamp := time.Now().Format("20060102-150405")
	random := RandomStringLower(4)
	
	name := fmt.Sprintf("%s-%s-%s-%s", prefix, resourceType, timestamp, random)
	
	slog.Info("Generated test resource name",
		"prefix", prefix,
		"resourceType", resourceType,
		"name", name,
	)
	
	return name
}

// TestDataPath creates a path to test data files
// Following Terratest patterns for test data organization
func TestDataPath(subPaths ...string) string {
	paths := append([]string{"testdata"}, subPaths...)
	path := filepath.Join(paths...)
	
	slog.Debug("Created test data path",
		"subPaths", subPaths,
		"fullPath", path,
	)
	
	return path
}

// LoadTestConfig loads test configuration
// Following Terratest patterns for configuration management
func LoadTestConfig(config TestConfig) TestConfig {
	slog.Info("Loading test configuration",
		"awsRegion", config.AWSRegion,
		"environment", config.Environment,
		"resourcePrefix", config.ResourcePrefix,
	)
	
	return config
}

// ValidateTestConfig validates test configuration
func ValidateTestConfig(config TestConfig) error {
	var errors []string
	
	if config.AWSRegion == "" {
		errors = append(errors, "AWSRegion is required")
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
	}
	
	slog.Info("Test configuration validated successfully")
	return nil
}

// ApplyTestConfigDefaults applies default values to test configuration
func ApplyTestConfigDefaults(config TestConfig) TestConfig {
	if config.AWSProfile == "" {
		config.AWSProfile = "default"
	}
	
	if config.Environment == "" {
		config.Environment = "test"
	}
	
	if config.ResourcePrefix == "" {
		config.ResourcePrefix = "terratest"
	}
	
	if config.Tags == nil {
		config.Tags = make(map[string]string)
	}
	
	// Add default tags
	if config.Tags["Environment"] == "" {
		config.Tags["Environment"] = config.Environment
	}
	
	if config.Tags["CreatedBy"] == "" {
		config.Tags["CreatedBy"] = "terratest"
	}
	
	slog.Info("Applied test configuration defaults",
		"awsProfile", config.AWSProfile,
		"environment", config.Environment,
		"resourcePrefix", config.ResourcePrefix,
	)
	
	return config
}

// CreateTestTags creates standard test tags
// Following Terratest patterns for resource tagging
func CreateTestTags(testName, environment string) map[string]string {
	tags := map[string]string{
		"TestName":    testName,
		"Environment": environment,
		"CreatedBy":   "terratest",
		"Timestamp":   time.Now().Format(time.RFC3339),
	}
	
	slog.Info("Created test tags",
		"testName", testName,
		"environment", environment,
		"tagCount", len(tags),
	)
	
	return tags
}

// MergeTestTags merges multiple tag maps, with later maps taking precedence
func MergeTestTags(tagMaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	
	for _, tagMap := range tagMaps {
		for key, value := range tagMap {
			result[key] = value
		}
	}
	
	slog.Debug("Merged test tags",
		"inputMaps", len(tagMaps),
		"resultTagCount", len(result),
	)
	
	return result
}

// ValidateTestFixture validates a test fixture structure
func ValidateTestFixture(fixture TestFixture) bool {
	if fixture.Name == "" {
		slog.Error("Test fixture validation failed: name is required")
		return false
	}
	
	if fixture.TerraformDir == "" {
		slog.Error("Test fixture validation failed: TerraformDir is required")
		return false
	}
	
	slog.Info("Test fixture validation passed",
		"name", fixture.Name,
		"terraformDir", fixture.TerraformDir,
	)
	
	return true
}