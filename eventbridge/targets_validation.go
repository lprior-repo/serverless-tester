package eventbridge

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// Target validation constants
const (
	MaxTargetRetryAttempts    = 185  // EventBridge maximum retry attempts
	MaxEventAgeSeconds       = 86400 // EventBridge maximum event age (24 hours)
	MaxTargetsPerRule        = 100   // EventBridge maximum targets per rule
	MaxInputTransformerSize  = 8192  // EventBridge maximum input transformer size
)

// ARN validation regex patterns
var (
	arnRegex     = regexp.MustCompile(`^arn:aws:[\w-]+:[\w-]*:\d{12}:[\w\/-]+$`)
	jsonPathRegex = regexp.MustCompile(`^\$[.\[\]'":\w\-\*]*$`)
)

// ValidateTargetConfiguration validates a target configuration and panics on error (Terratest pattern)
func ValidateTargetConfiguration(ctx *TestContext, target TargetConfig) {
	err := ValidateTargetConfigurationE(ctx, target)
	if err != nil {
		ctx.T.Errorf("ValidateTargetConfiguration failed: %v", err)
		ctx.T.FailNow()
	}
}

// ValidateTargetConfigurationE validates a target configuration and returns error
func ValidateTargetConfigurationE(ctx *TestContext, target TargetConfig) error {
	// Validate required fields
	if target.ID == "" {
		return fmt.Errorf("target ID cannot be empty")
	}

	if target.Arn == "" {
		return fmt.Errorf("target ARN cannot be empty")
	}

	// Validate ARN format
	if !arnRegex.MatchString(target.Arn) {
		return fmt.Errorf("invalid ARN format: %s", target.Arn)
	}

	// Validate JSON in Input field if provided
	if target.Input != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(target.Input), &jsonData); err != nil {
			return fmt.Errorf("invalid JSON in Input field: %v", err)
		}
	}

	// Validate JSONPath in InputPath field if provided
	if target.InputPath != "" {
		if !jsonPathRegex.MatchString(target.InputPath) {
			return fmt.Errorf("invalid JSONPath in InputPath: %s", target.InputPath)
		}
	}

	// Validate InputTransformer if provided
	if target.InputTransformer != nil {
		if err := ValidateInputTransformerE(target.InputTransformer); err != nil {
			return fmt.Errorf("invalid InputTransformer: %v", err)
		}
	}

	// Validate DeadLetterConfig if provided
	if target.DeadLetterConfig != nil {
		if err := ValidateDeadLetterConfigE(target.DeadLetterConfig); err != nil {
			return fmt.Errorf("invalid DeadLetterConfig: %v", err)
		}
	}

	// Validate RetryPolicy if provided
	if target.RetryPolicy != nil {
		if err := ValidateRetryPolicyE(target.RetryPolicy); err != nil {
			return fmt.Errorf("invalid RetryPolicy: %v", err)
		}
	}

	return nil
}

// GetTargetServiceType extracts the AWS service type from a target ARN
func GetTargetServiceType(targetArn string) string {
	if !arnRegex.MatchString(targetArn) {
		return "unknown"
	}

	// ARN format: arn:aws:service:region:account:resource-type/resource-id
	parts := strings.Split(targetArn, ":")
	if len(parts) < 6 {
		return "unknown"
	}

	service := parts[2]
	
	// Map AWS services to normalized names
	serviceMap := map[string]string{
		"lambda":       "lambda",
		"sqs":          "sqs", 
		"sns":          "sns",
		"states":       "stepfunctions",
		"kinesis":      "kinesis",
		"ecs":          "ecs",
		"events":       "eventbridge",
		"logs":         "cloudwatch",
		"firehose":     "kinesis-firehose",
		"pipes":        "eventbridge-pipes",
	}

	if normalizedService, exists := serviceMap[service]; exists {
		return normalizedService
	}

	return "unknown"
}

// BatchPutTargets puts targets in batches and panics on error (Terratest pattern)
func BatchPutTargets(ctx *TestContext, ruleName string, eventBusName string, targets []TargetConfig, batchSize int) {
	err := BatchPutTargetsE(ctx, ruleName, eventBusName, targets, batchSize)
	if err != nil {
		ctx.T.Errorf("BatchPutTargets failed: %v", err)
		ctx.T.FailNow()
	}
}

// BatchPutTargetsE puts targets in batches and returns error
func BatchPutTargetsE(ctx *TestContext, ruleName string, eventBusName string, targets []TargetConfig, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}

	if batchSize > MaxTargetsPerRule {
		batchSize = MaxTargetsPerRule
	}

	// Process targets in batches
	for i := 0; i < len(targets); i += batchSize {
		end := i + batchSize
		if end > len(targets) {
			end = len(targets)
		}

		batch := targets[i:end]
		if err := PutTargetsE(ctx, ruleName, eventBusName, batch); err != nil {
			return fmt.Errorf("failed to put batch starting at index %d: %v", i, err)
		}
	}

	return nil
}

// BatchRemoveTargets removes targets in batches and panics on error (Terratest pattern)
func BatchRemoveTargets(ctx *TestContext, ruleName string, eventBusName string, targetIDs []string, batchSize int) {
	err := BatchRemoveTargetsE(ctx, ruleName, eventBusName, targetIDs, batchSize)
	if err != nil {
		ctx.T.Errorf("BatchRemoveTargets failed: %v", err)
		ctx.T.FailNow()
	}
}

// BatchRemoveTargetsE removes targets in batches and returns error
func BatchRemoveTargetsE(ctx *TestContext, ruleName string, eventBusName string, targetIDs []string, batchSize int) error {
	if batchSize <= 0 {
		batchSize = MaxTargetsPerRule // Default to maximum batch size
	}

	if batchSize > MaxTargetsPerRule {
		batchSize = MaxTargetsPerRule
	}

	// Process target IDs in batches
	for i := 0; i < len(targetIDs); i += batchSize {
		end := i + batchSize
		if end > len(targetIDs) {
			end = len(targetIDs)
		}

		batch := targetIDs[i:end]
		if err := RemoveTargetsE(ctx, ruleName, eventBusName, batch); err != nil {
			return fmt.Errorf("failed to remove batch starting at index %d: %v", i, err)
		}
	}

	return nil
}

// BuildInputTransformer creates an InputTransformer with the given path mappings and template
func BuildInputTransformer(pathMap map[string]string, template string) *types.InputTransformer {
	return &types.InputTransformer{
		InputPathsMap: pathMap,
		InputTemplate: aws.String(template),
	}
}

// ValidateInputTransformer validates an InputTransformer and panics on error (Terratest pattern)  
func ValidateInputTransformer(transformer *types.InputTransformer) {
	err := ValidateInputTransformerE(transformer)
	if err != nil {
		panic(fmt.Sprintf("ValidateInputTransformer failed: %v", err))
	}
}

// ValidateInputTransformerE validates an InputTransformer and returns error
func ValidateInputTransformerE(transformer *types.InputTransformer) error {
	if transformer == nil {
		return fmt.Errorf("InputTransformer cannot be nil")
	}

	if transformer.InputTemplate == nil {
		return fmt.Errorf("InputTemplate is required")
	}

	template := *transformer.InputTemplate

	// Validate that template is valid JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(template), &jsonData); err != nil {
		return fmt.Errorf("invalid JSON in InputTemplate: %v", err)
	}

	// Check template size
	if len(template) > MaxInputTransformerSize {
		return fmt.Errorf("InputTemplate exceeds maximum size of %d bytes", MaxInputTransformerSize)
	}

	// Validate that all placeholders in template have corresponding mappings
	if transformer.InputPathsMap != nil {
		placeholders := extractPlaceholders(template)
		for _, placeholder := range placeholders {
			if _, exists := transformer.InputPathsMap[placeholder]; !exists {
				return fmt.Errorf("unmapped placeholder in template: %s", placeholder)
			}
		}
	}

	return nil
}

// extractPlaceholders extracts placeholder names from InputTransformer template
func extractPlaceholders(template string) []string {
	placeholderRegex := regexp.MustCompile(`<(\w+)>`)
	matches := placeholderRegex.FindAllStringSubmatch(template, -1)
	
	placeholders := make([]string, len(matches))
	for i, match := range matches {
		if len(match) > 1 {
			placeholders[i] = match[1]
		}
	}
	
	return placeholders
}

// BuildDeadLetterConfig creates a DeadLetterConfig with the given SQS queue ARN
func BuildDeadLetterConfig(queueArn string) *types.DeadLetterConfig {
	return &types.DeadLetterConfig{
		Arn: aws.String(queueArn),
	}
}

// ValidateDeadLetterConfig validates a DeadLetterConfig and panics on error (Terratest pattern)
func ValidateDeadLetterConfig(config *types.DeadLetterConfig) {
	err := ValidateDeadLetterConfigE(config)
	if err != nil {
		panic(fmt.Sprintf("ValidateDeadLetterConfig failed: %v", err))
	}
}

// ValidateDeadLetterConfigE validates a DeadLetterConfig and returns error
func ValidateDeadLetterConfigE(config *types.DeadLetterConfig) error {
	if config == nil {
		return fmt.Errorf("DeadLetterConfig cannot be nil")
	}

	if config.Arn == nil {
		return fmt.Errorf("dead letter queue ARN is required")
	}

	arn := *config.Arn

	// Validate ARN format
	if !arnRegex.MatchString(arn) {
		return fmt.Errorf("invalid dead letter queue ARN format: %s", arn)
	}

	// Validate that ARN is for SQS (dead letter queues must be SQS)
	if !strings.Contains(arn, ":sqs:") {
		return fmt.Errorf("dead letter queue must be an SQS queue, got: %s", arn)
	}

	return nil
}

// BuildRetryPolicy creates a RetryPolicy with the given retry attempts and maximum event age
func BuildRetryPolicy(maxRetryAttempts int32, maxEventAgeSeconds int32) *types.RetryPolicy {
	return &types.RetryPolicy{
		MaximumRetryAttempts:      aws.Int32(maxRetryAttempts),
		MaximumEventAgeInSeconds: aws.Int32(maxEventAgeSeconds),
	}
}

// ValidateRetryPolicy validates a RetryPolicy and panics on error (Terratest pattern)
func ValidateRetryPolicy(policy *types.RetryPolicy) {
	err := ValidateRetryPolicyE(policy)
	if err != nil {
		panic(fmt.Sprintf("ValidateRetryPolicy failed: %v", err))
	}
}

// ValidateRetryPolicyE validates a RetryPolicy and returns error
func ValidateRetryPolicyE(policy *types.RetryPolicy) error {
	if policy == nil {
		return fmt.Errorf("RetryPolicy cannot be nil")
	}

	// Validate retry attempts
	if policy.MaximumRetryAttempts != nil {
		attempts := *policy.MaximumRetryAttempts
		if attempts < 0 {
			return fmt.Errorf("maximum retry attempts cannot be negative: %d", attempts)
		}
		if attempts > MaxTargetRetryAttempts {
			return fmt.Errorf("maximum retry attempts exceeds limit of %d: %d", MaxTargetRetryAttempts, attempts)
		}
	}

	// Validate event age
	if policy.MaximumEventAgeInSeconds != nil {
		age := *policy.MaximumEventAgeInSeconds
		if age < 60 {  // Minimum is 1 minute
			return fmt.Errorf("maximum event age must be at least 60 seconds: %d", age)
		}
		if age > MaxEventAgeSeconds {
			return fmt.Errorf("maximum event age exceeds limit of %d seconds: %d", MaxEventAgeSeconds, age)
		}
	}

	return nil
}