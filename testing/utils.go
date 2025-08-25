package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	awstest "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
)

// Use TestingT interface from mock.go to avoid duplication

// TestUtils provides comprehensive testing utilities
type TestUtils struct {
	rng *rand.Rand
}

// NewTestUtils creates a new test utilities instance
func NewTestUtils() *TestUtils {
	return &TestUtils{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AWS Client Mocking

// MockLambdaClient represents a mock Lambda client for testing
type MockLambdaClient struct {
	InvokeCalls []MockInvokeCall
}

// MockInvokeCall represents a Lambda invoke call
type MockInvokeCall struct {
	FunctionName string
	Payload      []byte
}

// CreateMockLambdaClient creates a mock Lambda client for testing
func (u *TestUtils) CreateMockLambdaClient() *MockLambdaClient {
	return &MockLambdaClient{
		InvokeCalls: make([]MockInvokeCall, 0),
	}
}

// MockDynamoDBClient represents a mock DynamoDB client for testing
type MockDynamoDBClient struct {
	PutItemCalls []MockPutItemCall
	GetItemCalls []MockGetItemCall
}

// MockPutItemCall represents a DynamoDB PutItem call
type MockPutItemCall struct {
	TableName string
	Item      map[string]interface{}
}

// MockGetItemCall represents a DynamoDB GetItem call
type MockGetItemCall struct {
	TableName string
	Key       map[string]interface{}
}

// CreateMockDynamoDBClient creates a mock DynamoDB client for testing
func (u *TestUtils) CreateMockDynamoDBClient() *MockDynamoDBClient {
	return &MockDynamoDBClient{
		PutItemCalls: make([]MockPutItemCall, 0),
		GetItemCalls: make([]MockGetItemCall, 0),
	}
}

// Context Creation Helpers

// CreateBackgroundContext creates a background context
func (u *TestUtils) CreateBackgroundContext() context.Context {
	return context.Background()
}

// CreateTimeoutContext creates a context with timeout
func (u *TestUtils) CreateTimeoutContext(timeout time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return ctx
}

// CreateCancelableContext creates a cancelable context
func (u *TestUtils) CreateCancelableContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// Test Data Generation

// GenerateRandomString generates a random string of specified length
func (u *TestUtils) GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[u.rng.Intn(len(charset))]
	}
	return string(result)
}

// GenerateRandomNumber generates a random number between min and max (inclusive)
func (u *TestUtils) GenerateRandomNumber(min, max int) int {
	return u.rng.Intn(max-min+1) + min
}

// GenerateRandomId generates a unique ID using Terratest's random module
// This is the preferred method for ID generation following Terratest patterns
func (u *TestUtils) GenerateRandomId() string {
	return random.UniqueId()
}

// GenerateUUID generates a new UUID (legacy support)
// GenerateRandomId() is preferred for consistency with Terratest patterns
func (u *TestUtils) GenerateUUID() string {
	return uuid.New().String()
}

// TerratestAdapter adapts our TestingT to Terratest's TestingT interface
type TerratestAdapter struct {
	t TestingT
}

func (ta *TerratestAdapter) Error(args ...interface{}) {
	ta.t.Error(args...)
}

func (ta *TerratestAdapter) Errorf(format string, args ...interface{}) {
	ta.t.Errorf(format, args...)
}

func (ta *TerratestAdapter) Fail() {
	ta.t.Fail()
}

func (ta *TerratestAdapter) FailNow() {
	ta.t.FailNow()
}

func (ta *TerratestAdapter) Fatal(args ...interface{}) {
	ta.t.Error(args...)
	ta.t.FailNow()
}

func (ta *TerratestAdapter) Fatalf(format string, args ...interface{}) {
	ta.t.Errorf(format, args...)
	ta.t.FailNow()
}

func (ta *TerratestAdapter) Name() string {
	return ta.t.Name()
}

// GetRandomRegion gets a random AWS region using Terratest's aws module
func (u *TestUtils) GetRandomRegion(t TestingT) string {
	adapter := &TerratestAdapter{t: t}
	return awstest.GetRandomRegion(adapter, nil, nil)
}

// GetAccountId gets the current AWS account ID using Terratest's aws module
func (u *TestUtils) GetAccountId(t TestingT, region string) string {
	adapter := &TerratestAdapter{t: t}
	accountId, err := awstest.GetAccountIdE(adapter)
	if err != nil {
		t.Errorf("Failed to get AWS account ID: %v", err)
		t.FailNow()
	}
	return accountId
}

// Assertion Helpers

// AssertStringNotEmpty asserts that a string is not empty
func (u *TestUtils) AssertStringNotEmpty(t TestingT, value, message string) {
	if value == "" {
		t.Errorf("String is empty: %s", message)
	}
}

// AssertJSONValid asserts that a string is valid JSON
func (u *TestUtils) AssertJSONValid(t TestingT, jsonStr, message string) {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
		t.Errorf("Invalid JSON: %s - %v", message, err)
	}
}

// AssertAWSArn asserts that a string is a valid AWS ARN
func (u *TestUtils) AssertAWSArn(t TestingT, arn, message string) {
	// AWS ARN format: arn:partition:service:region:account-id:resource
	arnPattern := regexp.MustCompile(`^arn:[^:]+:[^:]+:[^:]*:[^:]*:[^:]+.*$`)
	if !arnPattern.MatchString(arn) {
		t.Errorf("Invalid AWS ARN format: %s - %s", message, arn)
	}
}

// Environment Variable Management

// EnvironmentManager manages environment variables for testing
type EnvironmentManager struct {
	originalValues map[string]string
}

// CreateEnvironmentManager creates a new environment variable manager
func (u *TestUtils) CreateEnvironmentManager() *EnvironmentManager {
	return &EnvironmentManager{
		originalValues: make(map[string]string),
	}
}

// SetEnv sets an environment variable, storing the original value
func (em *EnvironmentManager) SetEnv(key, value string) {
	if _, exists := em.originalValues[key]; !exists {
		em.originalValues[key] = os.Getenv(key)
	}
	os.Setenv(key, value)
}

// GetEnv gets an environment variable with default
func (em *EnvironmentManager) GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RestoreEnv restores an environment variable to its original value
func (em *EnvironmentManager) RestoreEnv(key, originalValue string) {
	if originalValue == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, originalValue)
	}
}

// RestoreAll restores all environment variables to their original values
func (em *EnvironmentManager) RestoreAll() {
	for key, originalValue := range em.originalValues {
		em.RestoreEnv(key, originalValue)
	}
	em.originalValues = make(map[string]string)
}

// Performance Benchmarking

// Benchmark tracks execution time for performance testing
type Benchmark struct {
	name      string
	startTime time.Time
}

// CreateBenchmark creates a new benchmark tracker
func (u *TestUtils) CreateBenchmark(name string) *Benchmark {
	return &Benchmark{
		name:      name,
		startTime: time.Now(),
	}
}

// Stop stops the benchmark and returns the duration
func (b *Benchmark) Stop() time.Duration {
	return time.Since(b.startTime)
}

// GetName returns the benchmark name
func (b *Benchmark) GetName() string {
	return b.name
}

// Retry Utilities

// RetryOperation retries an operation with exponential backoff
func (u *TestUtils) RetryOperation(operation func() error, maxAttempts int, baseDelay time.Duration) error {
	var lastErr error
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = operation()
		if lastErr == nil {
			return nil
		}
		
		if attempt < maxAttempts {
			delay := time.Duration(attempt) * baseDelay
			time.Sleep(delay)
		}
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", maxAttempts, lastErr)
}

// Validation Helpers

// ValidateAWSResourceName validates AWS resource naming conventions
func (u *TestUtils) ValidateAWSResourceName(name string) bool {
	// Basic AWS resource name validation (alphanumeric, hyphens, underscores)
	pattern := regexp.MustCompile(`^[a-zA-Z0-9-_\.]+$`)
	return pattern.MatchString(name) && len(name) >= 1 && len(name) <= 255
}

// ValidateEmail validates basic email format
func (u *TestUtils) ValidateEmail(email string) bool {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return pattern.MatchString(email)
}

// String Helpers

// SanitizeForAWS sanitizes a string for AWS resource naming
func (u *TestUtils) SanitizeForAWS(input string) string {
	// Convert to lowercase and replace invalid characters
	sanitized := strings.ToLower(input)
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
	
	// Remove leading/trailing hyphens and ensure not empty
	result = strings.Trim(result, "-")
	if result == "" {
		result = "default"
	}
	
	return result
}

// TruncateString truncates a string to specified length with ellipsis
func (u *TestUtils) TruncateString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	if maxLength <= 3 {
		return "..."
	}
	return input[:maxLength-3] + "..."
}