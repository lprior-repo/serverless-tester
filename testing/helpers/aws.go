package helpers

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
)

// DynamoDBClient defines the interface for DynamoDB operations
type DynamoDBClient interface {
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error)
}

// LambdaClient defines the interface for Lambda operations
type LambdaClient interface {
	GetFunction(ctx context.Context, params *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error)
}

// ValidateAWSResourceName validates AWS resource naming conventions
// Follows AWS naming rules: alphanumeric, hyphens, underscores, dots
func ValidateAWSResourceName(name string) bool {
	if name == "" {
		return false
	}
	
	// AWS resource name pattern: alphanumeric, hyphens, underscores, dots
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	return pattern.MatchString(name) && len(name) <= 255
}

// ValidateAWSArn validates AWS ARN format
// ARN format: arn:partition:service:region:account-id:resource
func ValidateAWSArn(arn string) bool {
	if arn == "" {
		return false
	}
	
	// Basic ARN pattern validation
	pattern := regexp.MustCompile(`^arn:[^:]+:[^:]+:[^:]*:[^:]*:.+`)
	return pattern.MatchString(arn)
}

// AssertDynamoDBTableExists validates that a DynamoDB table exists
// Following Terratest patterns for AWS resource assertions
func AssertDynamoDBTableExists(ctx context.Context, client DynamoDBClient, tableName string) error {
	slog.Info("Checking DynamoDB table existence",
		"tableName", tableName,
	)
	
	input := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	
	_, err := client.DescribeTable(ctx, input)
	if err != nil {
		slog.Error("DynamoDB table does not exist",
			"tableName", tableName,
			"error", err.Error(),
		)
		return fmt.Errorf("DynamoDB table %s does not exist: %w", tableName, err)
	}
	
	slog.Info("DynamoDB table exists",
		"tableName", tableName,
	)
	return nil
}

// AssertDynamoDBTableExistsE validates that a DynamoDB table exists with retry logic
// Following Terratest patterns for AWS resource assertions with retries
func AssertDynamoDBTableExistsE(ctx context.Context, client DynamoDBClient, tableName string, maxRetries int) error {
	return DoWithRetry(
		fmt.Sprintf("assert DynamoDB table %s exists", tableName),
		maxRetries,
		2*time.Second,
		func() error {
			return AssertDynamoDBTableExists(ctx, client, tableName)
		},
	)
}

// AssertLambdaFunctionExists validates that a Lambda function exists
// Following Terratest patterns for AWS resource assertions
func AssertLambdaFunctionExists(ctx context.Context, client LambdaClient, functionName string) error {
	slog.Info("Checking Lambda function existence",
		"functionName", functionName,
	)
	
	input := &lambda.GetFunctionInput{
		FunctionName: &functionName,
	}
	
	_, err := client.GetFunction(ctx, input)
	if err != nil {
		slog.Error("Lambda function does not exist",
			"functionName", functionName,
			"error", err.Error(),
		)
		return fmt.Errorf("Lambda function %s does not exist: %w", functionName, err)
	}
	
	slog.Info("Lambda function exists",
		"functionName", functionName,
	)
	return nil
}

// AssertLambdaFunctionExistsE validates that a Lambda function exists with retry logic
// Following Terratest patterns for AWS resource assertions with retries
func AssertLambdaFunctionExistsE(ctx context.Context, client LambdaClient, functionName string, maxRetries int) error {
	return DoWithRetry(
		fmt.Sprintf("assert Lambda function %s exists", functionName),
		maxRetries,
		2*time.Second,
		func() error {
			return AssertLambdaFunctionExists(ctx, client, functionName)
		},
	)
}

// GetAWSResourceType extracts the resource type from an AWS ARN
func GetAWSResourceType(arn string) string {
	if !ValidateAWSArn(arn) {
		return ""
	}
	
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return ""
	}
	
	// For Lambda functions, the resource type is in parts[5] and resource name in parts[6]
	if len(parts) >= 7 && parts[2] == "lambda" {
		return parts[5] // "function"
	}
	
	resource := parts[5]
	
	// Handle resource types with separators (e.g., "table/my-table" -> "table")
	if strings.Contains(resource, "/") {
		resourceParts := strings.SplitN(resource, "/", 2)
		return resourceParts[0]
	}
	
	// For resources without type separators (like S3 buckets), return empty
	return ""
}

// GetAWSResourceName extracts the resource name from an AWS ARN
func GetAWSResourceName(arn string) string {
	if !ValidateAWSArn(arn) {
		return ""
	}
	
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return ""
	}
	
	// For Lambda functions, the resource name is in parts[6], optionally with version in parts[7]
	if len(parts) >= 7 && parts[2] == "lambda" {
		if len(parts) >= 8 {
			return parts[6] + ":" + parts[7] // function name with version
		}
		return parts[6] // function name without version
	}
	
	resource := parts[5]
	
	// Handle resource names with separators (e.g., "table/my-table" -> "my-table")
	if strings.Contains(resource, "/") {
		resourceParts := strings.SplitN(resource, "/", 2)
		if len(resourceParts) > 1 {
			return resourceParts[1]
		}
	}
	
	// For resources without separators (like S3 buckets), return the whole resource part
	return resource
}

// Use the TestingT interface from retry.go to avoid duplication

// GetRandomRegion returns a random AWS region using Terratest's aws module
// This provides consistent region selection for distributed testing
func GetRandomRegion(t TestingT) string {
	t.Helper()
	return aws.GetRandomRegion(t, nil, nil)
}

// GetRandomRegionExcluding returns a random AWS region excluding specified regions
// This is useful when testing multi-region scenarios or avoiding specific regions
func GetRandomRegionExcluding(t TestingT, excludeRegions []string) string {
	t.Helper()
	return aws.GetRandomRegion(t, nil, excludeRegions)
}

// GetAccountId gets the current AWS account ID using Terratest's aws module
// This is useful for constructing ARNs and validating account-specific resources
func GetAccountId(t TestingT) string {
	t.Helper()
	accountId, err := aws.GetAccountIdE(t)
	if err != nil {
		t.Errorf("Failed to get AWS account ID: %v", err)
		t.FailNow()
	}
	return accountId
}

// GetAccountIdE gets the current AWS account ID using Terratest's aws module with error handling
func GetAccountIdE(t TestingT) (string, error) {
	return aws.GetAccountIdE(t)
}

// GetAvailabilityZones returns availability zones for the specified region
func GetAvailabilityZones(t TestingT, region string) []string {
	t.Helper()
	return aws.GetAvailabilityZones(t, region)
}

// GetAvailabilityZonesE returns availability zones for the specified region with error handling
func GetAvailabilityZonesE(t TestingT, region string) ([]string, error) {
	return aws.GetAvailabilityZonesE(t, region)
}

// CreateRandomResourceName creates a resource name with random suffix using Terratest patterns
// This ensures unique resource names across test runs
func CreateRandomResourceName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, random.UniqueId())
}

// CreateRandomAWSResourceName creates a resource name that follows AWS naming conventions
// This uses Terratest's random ID generation and validates AWS naming rules
func CreateRandomAWSResourceName(prefix string) string {
	if prefix == "" {
		prefix = "test"
	}
	
	// Create name with random suffix
	name := fmt.Sprintf("%s-%s", prefix, random.UniqueId())
	
	// Ensure it follows AWS naming conventions
	if !ValidateAWSResourceName(name) {
		// If validation fails, create a simpler name
		name = fmt.Sprintf("%s-%s", strings.ToLower(prefix), strings.ReplaceAll(random.UniqueId(), "_", "-"))
	}
	
	return name
}

// BuildARN constructs an AWS ARN from components using standard format
func BuildARN(partition, service, region, accountID, resource string) string {
	return fmt.Sprintf("arn:%s:%s:%s:%s:%s", partition, service, region, accountID, resource)
}

// BuildLambdaFunctionARN constructs a Lambda function ARN
func BuildLambdaFunctionARN(region, accountID, functionName string) string {
	return BuildARN("aws", "lambda", region, accountID, fmt.Sprintf("function:%s", functionName))
}

// BuildDynamoDBTableARN constructs a DynamoDB table ARN
func BuildDynamoDBTableARN(region, accountID, tableName string) string {
	return BuildARN("aws", "dynamodb", region, accountID, fmt.Sprintf("table/%s", tableName))
}

// BuildEventBridgeRuleARN constructs an EventBridge rule ARN
func BuildEventBridgeRuleARN(region, accountID, ruleName string) string {
	return BuildARN("aws", "events", region, accountID, fmt.Sprintf("rule/%s", ruleName))
}

// BuildStepFunctionsStateMachineARN constructs a Step Functions state machine ARN
func BuildStepFunctionsStateMachineARN(region, accountID, stateMachineName string) string {
	return BuildARN("aws", "states", region, accountID, fmt.Sprintf("stateMachine:%s", stateMachineName))
}

// ValidateRegion checks if a region string is a valid AWS region format
func ValidateRegion(region string) bool {
	if region == "" {
		return false
	}
	
	// AWS region format: us-west-2, eu-central-1, ap-southeast-1, etc.
	pattern := regexp.MustCompile(`^[a-z]{2}-[a-z]+-\d+$`)
	return pattern.MatchString(region)
}

// IsAWSService checks if a service name is a valid AWS service identifier
func IsAWSService(service string) bool {
	// List of common AWS services (this could be expanded)
	commonServices := map[string]bool{
		"lambda":      true,
		"dynamodb":    true,
		"s3":          true,
		"events":      true,
		"states":      true,
		"iam":         true,
		"ec2":         true,
		"rds":         true,
		"sqs":         true,
		"sns":         true,
		"cloudwatch":  true,
		"logs":        true,
		"apigateway":  true,
		"cognito":     true,
		"kinesis":     true,
	}
	
	return commonServices[service]
}

// SanitizeResourceNameForAWS sanitizes a string for AWS resource naming
// This removes invalid characters and ensures the name follows AWS conventions
func SanitizeResourceNameForAWS(input string) string {
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