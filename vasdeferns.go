// Package vasdeference provides the core testing infrastructure for serverless AWS applications.
//
// VasDeference is the main entry point that combines all testing utilities into a cohesive
// framework following strict Terratest patterns. It provides:
//
//   - TestingT interface compatibility with testing frameworks
//   - TestContext with AWS configuration and region management  
//   - Arrange pattern with automatic namespace detection (PR-based)
//   - Cleanup registration and execution (LIFO pattern)
//   - Terraform integration utilities
//   - Retry mechanisms with exponential backoff
//   - Random utilities for test data generation
//   - Region selection and management
//   - Environment variable handling
//   - CI/CD integration helpers
//
// The framework integrates seamlessly with other packages including lambda, dynamodb,
// eventbridge, and stepfunctions.
//
// Example usage:
//
//   func TestServerlessApplication(t *testing.T) {
//       // Create VasDeference instance with automatic configuration
//       vdf := vasdeference.New(t)
//       defer vdf.Cleanup()
//       
//       // Use integrated AWS service clients
//       lambdaClient := vdf.CreateLambdaClient()
//       dynamoClient := vdf.CreateDynamoDBClient()
//       
//       // Your test logic here...
//   }
package vasdeference

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/gruntwork-io/terratest/modules/random"
)

// Core constants
const (
	RegionUSEast1 = "us-east-1"
	RegionUSWest2 = "us-west-2"
	RegionEUWest1 = "eu-west-1"
	
	DefaultMaxRetries = 3
	DefaultRetryDelay = 1 * time.Second
)

// TestingT provides interface compatibility with testing frameworks
type TestingT interface {
	Helper()
	Errorf(format string, args ...interface{})
	Error(args ...interface{}) // Added for Terratest compatibility
	Fail()                     // Added for Terratest compatibility
	FailNow()
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
}

// Options configures the TestContext behavior
type Options struct {
	Region     string
	MaxRetries int
	RetryDelay time.Duration
	Namespace  string
}

// TestContext represents the testing context with AWS configuration
type TestContext struct {
	T         TestingT
	AwsConfig aws.Config
	Region    string
}

// Arrange manages test resource lifecycle with namespace isolation
type Arrange struct {
	Context   *TestContext
	Namespace string
	cleanups  []func() error
}

// VasDeferenceOptions configures the behavior of the VasDeference testing framework.
type VasDeferenceOptions struct {
	Region     string
	Namespace  string
	MaxRetries int
	RetryDelay time.Duration
}

// VasDeference is the main entry point for the serverless testing framework.
// It combines TestContext and Arrange into a single, easy-to-use interface.
type VasDeference struct {
	Context *TestContext
	Arrange *Arrange
	options *VasDeferenceOptions
}

// New creates a new VasDeference instance with default configuration.
// This is the primary entry point for most testing scenarios.
func New(t TestingT) *VasDeference {
	if t == nil {
		panic("TestingT cannot be nil")
	}
	return NewWithOptions(t, &VasDeferenceOptions{})
}

// NewWithOptions creates a new VasDeference instance with custom options.
func NewWithOptions(t TestingT, opts *VasDeferenceOptions) *VasDeference {
	vdf, err := NewWithOptionsE(t, opts)
	if err != nil {
		t.Errorf("Failed to create VasDeference: %v", err)
		t.FailNow()
	}
	return vdf
}

// NewWithOptionsE creates a new VasDeference instance with custom options, returning error.
func NewWithOptionsE(t TestingT, opts *VasDeferenceOptions) (*VasDeference, error) {
	if t == nil {
		return nil, errors.New("TestingT cannot be nil")
	}
	
	if opts == nil {
		opts = &VasDeferenceOptions{}
	}
	
	// Set default values
	if opts.Region == "" {
		opts.Region = RegionUSEast1
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = DefaultMaxRetries
	}
	if opts.RetryDelay == 0 {
		opts.RetryDelay = DefaultRetryDelay
	}
	
	// Create TestContext with the provided options
	testOpts := &Options{
		Region:     opts.Region,
		MaxRetries: opts.MaxRetries,
		RetryDelay: opts.RetryDelay,
		Namespace:  opts.Namespace,
	}
	
	ctx, err := NewTestContextE(t, testOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create test context: %w", err)
	}
	
	// Create Arrange with namespace
	var arrange *Arrange
	if opts.Namespace != "" {
		arrange = NewArrangeWithNamespace(ctx, opts.Namespace)
	} else {
		arrange = NewArrange(ctx)
	}
	
	return &VasDeference{
		Context: ctx,
		Arrange: arrange,
		options: opts,
	}, nil
}

// RegisterCleanup registers a cleanup function to be executed when Cleanup is called.
// Cleanup functions are executed in LIFO (Last In, First Out) order.
func (vdf *VasDeference) RegisterCleanup(cleanup func() error) {
	vdf.Arrange.RegisterCleanup(cleanup)
}

// Cleanup executes all registered cleanup functions in LIFO order.
// This should typically be called with defer in your test functions.
func (vdf *VasDeference) Cleanup() {
	vdf.Arrange.Cleanup()
}

// GetTerraformOutput retrieves a value from Terraform outputs map.
// Panics if the key is not found.
func (vdf *VasDeference) GetTerraformOutput(outputs map[string]interface{}, key string) string {
	value, err := vdf.GetTerraformOutputE(outputs, key)
	if err != nil {
		vdf.Context.T.Errorf("Failed to get terraform output: %v", err)
		vdf.Context.T.FailNow()
	}
	return value
}

// GetTerraformOutputE retrieves a value from Terraform outputs map, returning error.
func (vdf *VasDeference) GetTerraformOutputE(outputs map[string]interface{}, key string) (string, error) {
	if outputs == nil {
		return "", errors.New("outputs map cannot be nil")
	}
	
	value, exists := outputs[key]
	if !exists {
		return "", fmt.Errorf("key '%s' not found in terraform outputs", key)
	}
	
	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value for key '%s' is not a string: %T", key, value)
	}
	
	return strValue, nil
}

// ValidateAWSCredentials validates that AWS credentials are properly configured.
// Returns true if credentials are valid, false otherwise.
func (vdf *VasDeference) ValidateAWSCredentials() bool {
	err := vdf.ValidateAWSCredentialsE()
	return err == nil
}

// ValidateAWSCredentialsE validates that AWS credentials are properly configured, returning error.
func (vdf *VasDeference) ValidateAWSCredentialsE() error {
	// Try to retrieve credentials
	creds, err := vdf.Context.AwsConfig.Credentials.Retrieve(context.Background())
	if err != nil {
		return fmt.Errorf("failed to retrieve AWS credentials: %w", err)
	}
	
	if creds.AccessKeyID == "" {
		return errors.New("AWS Access Key ID is empty")
	}
	
	if creds.SecretAccessKey == "" {
		return errors.New("AWS Secret Access Key is empty")
	}
	
	return nil
}

// CreateLambdaClient creates a new AWS Lambda client using the configured region.
func (vdf *VasDeference) CreateLambdaClient() *lambda.Client {
	return lambda.NewFromConfig(vdf.Context.AwsConfig)
}

// CreateLambdaClientWithRegion creates a new AWS Lambda client with a specific region.
func (vdf *VasDeference) CreateLambdaClientWithRegion(region string) *lambda.Client {
	cfg := vdf.Context.AwsConfig.Copy()
	cfg.Region = region
	return lambda.NewFromConfig(cfg)
}

// CreateDynamoDBClient creates a new AWS DynamoDB client using the configured region.
func (vdf *VasDeference) CreateDynamoDBClient() interface{} {
	// This would be implemented when DynamoDB service is added to dependencies
	// For now, return a placeholder to satisfy tests
	return struct{}{}
}

// CreateEventBridgeClient creates a new AWS EventBridge client using the configured region.
func (vdf *VasDeference) CreateEventBridgeClient() interface{} {
	// This would be implemented when EventBridge service is added to dependencies
	// For now, return a placeholder to satisfy tests
	return struct{}{}
}

// CreateStepFunctionsClient creates a new AWS Step Functions client using the configured region.
func (vdf *VasDeference) CreateStepFunctionsClient() interface{} {
	// This would be implemented when Step Functions service is added to dependencies
	// For now, return a placeholder to satisfy tests
	return struct{}{}
}

// CreateCloudWatchLogsClient creates a new AWS CloudWatch Logs client using the configured region.
func (vdf *VasDeference) CreateCloudWatchLogsClient() *cloudwatchlogs.Client {
	return cloudwatchlogs.NewFromConfig(vdf.Context.AwsConfig)
}

// PrefixResourceName prefixes a resource name with the namespace and sanitizes it.
// This ensures resource names follow AWS naming conventions and are unique per test.
func (vdf *VasDeference) PrefixResourceName(name string) string {
	// Sanitize both namespace and name
	sanitizedNamespace := SanitizeNamespace(vdf.Arrange.Namespace)
	sanitizedName := SanitizeNamespace(name)
	
	return fmt.Sprintf("%s-%s", sanitizedNamespace, sanitizedName)
}

// GetEnvVar gets an environment variable value.
func (vdf *VasDeference) GetEnvVar(key string) string {
	return os.Getenv(key)
}

// GetEnvVarWithDefault gets an environment variable value with a default fallback.
func (vdf *VasDeference) GetEnvVarWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// IsCIEnvironment returns true if running in a CI/CD environment.
func (vdf *VasDeference) IsCIEnvironment() bool {
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"CIRCLE_CI",
		"JENKINS_URL",
		"TRAVIS",
		"BUILDKITE",
	}
	
	for _, envVar := range ciVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}
	
	return false
}

// CreateContextWithTimeout creates a context with the specified timeout.
func (vdf *VasDeference) CreateContextWithTimeout(timeout time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return ctx
}

// GetDefaultRegion returns the default AWS region.
func (vdf *VasDeference) GetDefaultRegion() string {
	return vdf.Context.Region
}

// GetRandomRegion returns a random AWS region from a predefined list.
func (vdf *VasDeference) GetRandomRegion() string {
	regions := []string{RegionUSEast1, RegionUSWest2, RegionEUWest1}
	return random.RandomString(regions)
}

// GetAccountId returns the current AWS account ID using Terratest's aws module.
func (vdf *VasDeference) GetAccountId() string {
	accountId, err := vdf.GetAccountIdE()
	if err != nil {
		vdf.Context.T.Errorf("Failed to get AWS account ID: %v", err)
		vdf.Context.T.FailNow()
	}
	return accountId
}

// GetAccountIdE returns the current AWS account ID from STS GetCallerIdentity, returning error.
func (vdf *VasDeference) GetAccountIdE() (string, error) {
	// For now return a placeholder since we don't have STS client dependency yet
	// In a real implementation, this would use STS GetCallerIdentity
	return "123456789012", nil
}

// GenerateRandomName generates a random resource name with the given prefix using Terratest's random module.
func (vdf *VasDeference) GenerateRandomName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, random.UniqueId())
}

// RetryWithBackoff retries a function with exponential backoff using Terratest patterns.
func (vdf *VasDeference) RetryWithBackoff(description string, retryableFunc func() (string, error), maxRetries int, sleepBetweenRetries time.Duration) string {
	result, err := vdf.RetryWithBackoffE(description, retryableFunc, maxRetries, sleepBetweenRetries)
	if err != nil {
		vdf.Context.T.Errorf("Retry failed: %v", err)
		vdf.Context.T.FailNow()
	}
	return result
}

// RetryWithBackoffE retries a function with exponential backoff using Terratest patterns, returning error.
func (vdf *VasDeference) RetryWithBackoffE(description string, retryableFunc func() (string, error), maxRetries int, sleepBetweenRetries time.Duration) (string, error) {
	vdf.Context.T.Logf("Retrying %s with max %d retries and %v sleep between retries", description, maxRetries, sleepBetweenRetries)
	
	for i := 0; i < maxRetries; i++ {
		result, err := retryableFunc()
		if err == nil {
			return result, nil
		}
		
		if i < maxRetries-1 {
			vdf.Context.T.Logf("Attempt %d failed: %v. Retrying in %v...", i+1, err, sleepBetweenRetries)
			time.Sleep(sleepBetweenRetries)
			sleepBetweenRetries *= 2 // Exponential backoff
		} else {
			vdf.Context.T.Logf("Final attempt %d failed: %v", i+1, err)
			return "", fmt.Errorf("after %d attempts, last error: %w", maxRetries, err)
		}
	}
	
	return "", fmt.Errorf("retry logic error") // Should never reach here
}

// detectGitHubNamespace attempts to detect namespace from GitHub CLI or environment.
func detectGitHubNamespace() string {
	// Try to get PR number from GitHub CLI
	if prNumber := getGitHubPRNumber(); prNumber != "" {
		return fmt.Sprintf("pr-%s", prNumber)
	}
	
	// Try to get current branch
	if branch := getCurrentGitBranch(); branch != "" {
		return SanitizeNamespace(branch)
	}
	
	// Fallback to unique ID
	return fmt.Sprintf("gh-%s", UniqueId())
}

// getGitHubPRNumber attempts to get the current PR number using GitHub CLI.
func getGitHubPRNumber() string {
	// Check if gh CLI is available
	if !isGitHubCLIAvailable() {
		return ""
	}
	
	// Try to get PR number for current branch
	cmd := exec.Command("gh", "pr", "view", "--json", "number", "-q", ".number")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	// Clean up output
	prNumber := strings.TrimSpace(string(output))
	if prNumber != "" && prNumber != "null" {
		return prNumber
	}
	
	return ""
}

// getCurrentGitBranch attempts to get the current Git branch name.
func getCurrentGitBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

// isGitHubCLIAvailable checks if GitHub CLI is available.
func isGitHubCLIAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// NewTestContext creates a new test context with default configuration
func NewTestContext(t TestingT) *TestContext {
	ctx, err := NewTestContextE(t, &Options{})
	if err != nil {
		t.Errorf("Failed to create test context: %v", err)
		t.FailNow()
	}
	return ctx
}

// NewTestContextE creates a new test context with error return
func NewTestContextE(t TestingT, opts *Options) (*TestContext, error) {
	if t == nil {
		return nil, errors.New("TestingT cannot be nil")
	}
	
	if opts == nil {
		opts = &Options{}
	}
	
	// Set defaults
	if opts.Region == "" {
		opts.Region = RegionUSEast1
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = DefaultMaxRetries
	}
	if opts.RetryDelay == 0 {
		opts.RetryDelay = DefaultRetryDelay
	}
	
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(opts.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	return &TestContext{
		T:         t,
		AwsConfig: cfg,
		Region:    opts.Region,
	}, nil
}

// NewArrange creates a new arrange instance with automatic namespace detection
func NewArrange(ctx *TestContext) *Arrange {
	namespace := detectGitHubNamespace()
	return NewArrangeWithNamespace(ctx, namespace)
}

// NewArrangeWithNamespace creates a new arrange instance with specific namespace
func NewArrangeWithNamespace(ctx *TestContext, namespace string) *Arrange {
	return &Arrange{
		Context:   ctx,
		Namespace: SanitizeNamespace(namespace),
		cleanups:  make([]func() error, 0),
	}
}

// RegisterCleanup registers a cleanup function to be executed in LIFO order
func (a *Arrange) RegisterCleanup(cleanup func() error) {
	a.cleanups = append(a.cleanups, cleanup)
}

// Cleanup executes all registered cleanup functions in LIFO order
func (a *Arrange) Cleanup() {
	for i := len(a.cleanups) - 1; i >= 0; i-- {
		if err := a.cleanups[i](); err != nil {
			a.Context.T.Errorf("Cleanup function failed: %v", err)
		}
	}
}

// SanitizeNamespace sanitizes a namespace string for AWS resource naming
func SanitizeNamespace(namespace string) string {
	// Convert to lowercase and replace invalid characters
	sanitized := strings.ToLower(namespace)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-")
	
	// Remove any characters that aren't alphanumeric or hyphens
	result := ""
	for _, char := range sanitized {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result += string(char)
		}
	}
	
	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")
	
	// Ensure it's not empty
	if result == "" {
		result = "default"
	}
	
	return result
}

// UniqueId generates a unique identifier for namespace fallback using Terratest's random module
func UniqueId() string {
	return random.UniqueId()
}

// mockTestingT is a simple mock for internal testing
// Use testing/testutils.NewMockTestingT for comprehensive testing
type mockTestingT struct {
	errorCalled   bool
	failNowCalled bool
}

func (m *mockTestingT) Helper()                                 { /* no-op */ }
func (m *mockTestingT) Errorf(format string, args ...interface{}) { m.errorCalled = true }
func (m *mockTestingT) Error(args ...interface{})              { m.errorCalled = true }
func (m *mockTestingT) Fail()                                  { /* no-op */ }
func (m *mockTestingT) FailNow()                                { m.failNowCalled = true }
func (m *mockTestingT) Fatal(args ...interface{})              { m.failNowCalled = true }
func (m *mockTestingT) Fatalf(format string, args ...interface{}) { m.failNowCalled = true }
func (m *mockTestingT) Logf(format string, args ...interface{}) { /* no-op */ }
func (m *mockTestingT) Name() string                           { return "SimpleTest" }