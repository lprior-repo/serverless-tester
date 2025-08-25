// Package vasdeference provides Terratest integration helpers for comprehensive serverless testing.
//
// This file contains utilities that directly leverage Terratest modules to provide
// enhanced testing capabilities for AWS serverless applications.
package vasdeference

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

// TerraformOptions creates Terraform options for testing infrastructure
func (vdf *VasDeference) TerraformOptions(terraformDir string, vars map[string]interface{}) *terraform.Options {
	return &terraform.Options{
		TerraformDir: terraformDir,
		Vars:         vars,
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": vdf.Context.Region,
		},
		NoColor: true,
	}
}

// InitAndApplyTerraform initializes and applies Terraform configuration
func (vdf *VasDeference) InitAndApplyTerraform(options *terraform.Options) {
	err := vdf.InitAndApplyTerraformE(options)
	if err != nil {
		vdf.Context.T.Errorf("Failed to initialize and apply Terraform: %v", err)
		vdf.Context.T.FailNow()
	}
}

// InitAndApplyTerraformE initializes and applies Terraform configuration, returning error
func (vdf *VasDeference) InitAndApplyTerraformE(options *terraform.Options) error {
	terraform.InitAndApply(vdf.Context.T, options)
	
	// Register cleanup to destroy resources
	vdf.RegisterCleanup(func() error {
		terraform.Destroy(vdf.Context.T, options)
		return nil
	})
	
	return nil
}

// GetTerraformOutputAsString retrieves a Terraform output value as string
func (vdf *VasDeference) GetTerraformOutputAsString(options *terraform.Options, key string) string {
	return terraform.Output(vdf.Context.T, options, key)
}

// GetTerraformOutputAsMap retrieves a Terraform output value as map
func (vdf *VasDeference) GetTerraformOutputAsMap(options *terraform.Options, key string) map[string]interface{} {
	outputMap := terraform.OutputMap(vdf.Context.T, options, key)
	result := make(map[string]interface{})
	for k, v := range outputMap {
		result[k] = v
	}
	return result
}

// GetTerraformOutputAsList retrieves a Terraform output value as list
func (vdf *VasDeference) GetTerraformOutputAsList(options *terraform.Options, key string) []interface{} {
	outputList := terraform.OutputList(vdf.Context.T, options, key)
	result := make([]interface{}, len(outputList))
	for i, v := range outputList {
		result[i] = v
	}
	return result
}

// RetryableAction represents an action that can be retried
type RetryableAction struct {
	Description   string
	MaxRetries    int
	TimeBetween   time.Duration
	Action        func() (string, error)
}

// DoWithRetry executes an action with Terratest's retry logic
func (vdf *VasDeference) DoWithRetry(action RetryableAction) string {
	result, err := vdf.DoWithRetryE(action)
	if err != nil {
		vdf.Context.T.Errorf("Action failed after retries: %v", err)
		vdf.Context.T.FailNow()
	}
	return result
}

// DoWithRetryE executes an action with Terratest's retry logic, returning error
func (vdf *VasDeference) DoWithRetryE(action RetryableAction) (string, error) {
	var result string
	var lastErr error
	
	retry.DoWithRetry(
		vdf.Context.T,
		action.Description,
		action.MaxRetries,
		action.TimeBetween,
		func() (string, error) {
			res, err := action.Action()
			if err != nil {
				lastErr = err
				return "", err
			}
			result = res
			lastErr = nil
			return res, nil
		},
	)
	
	return result, lastErr
}

// WaitUntil waits until a condition is met using Terratest's retry logic
func (vdf *VasDeference) WaitUntil(description string, maxRetries int, timeBetween time.Duration, condition func() (bool, error)) {
	err := vdf.WaitUntilE(description, maxRetries, timeBetween, condition)
	if err != nil {
		vdf.Context.T.Errorf("Wait condition failed: %v", err)
		vdf.Context.T.FailNow()
	}
}

// WaitUntilE waits until a condition is met using Terratest's retry logic, returning error
func (vdf *VasDeference) WaitUntilE(description string, maxRetries int, timeBetween time.Duration, condition func() (bool, error)) error {
	var lastErr error
	retry.DoWithRetry(
		vdf.Context.T,
		description,
		maxRetries,
		timeBetween,
		func() (string, error) {
			success, condErr := condition()
			if condErr != nil {
				lastErr = condErr
				return "", condErr
			}
			if !success {
				lastErr = fmt.Errorf("condition not met")
				return "", lastErr
			}
			lastErr = nil
			return "success", nil
		},
	)
	return lastErr
}

// GetAWSSession creates an AWS session using Terratest's aws module
func (vdf *VasDeference) GetAWSSession() interface{} {
	// Note: This would typically return *session.Session from aws-sdk-go v1
	// For v2, we use the config directly from vdf.Context.AwsConfig
	return vdf.Context.AwsConfig
}

// GetAllAvailableRegions returns all available AWS regions
func (vdf *VasDeference) GetAllAvailableRegions() []string {
	return aws.GetAllAwsRegions(vdf.Context.T)
}

// GetAvailabilityZones returns availability zones for the current region
func (vdf *VasDeference) GetAvailabilityZones() []string {
	return aws.GetAvailabilityZones(vdf.Context.T, vdf.Context.Region)
}

// CreateRandomResourceName creates a resource name with random suffix
func (vdf *VasDeference) CreateRandomResourceName(prefix string, maxLength int) string {
	suffix := random.UniqueId()
	name := fmt.Sprintf("%s-%s", prefix, suffix)
	
	// Truncate if necessary to meet AWS naming requirements
	if len(name) > maxLength {
		availableLength := maxLength - len(suffix) - 1 // -1 for hyphen
		if availableLength > 0 {
			name = fmt.Sprintf("%s-%s", prefix[:availableLength], suffix)
		} else {
			name = suffix[:maxLength]
		}
	}
	
	return strings.ToLower(name)
}

// CreateTestIdentifier creates a unique test identifier combining namespace and random ID
func (vdf *VasDeference) CreateTestIdentifier() string {
	return fmt.Sprintf("%s-%s", vdf.Arrange.Namespace, random.UniqueId())
}

// GetRandomRegionExcluding returns a random AWS region excluding specified regions
func (vdf *VasDeference) GetRandomRegionExcluding(excludeRegions []string) string {
	return aws.GetRandomRegion(vdf.Context.T, nil, excludeRegions)
}

// ValidateAWSResourceExists validates that an AWS resource exists
func (vdf *VasDeference) ValidateAWSResourceExists(resourceType, resourceId string, validator func() (bool, error)) {
	err := vdf.ValidateAWSResourceExistsE(resourceType, resourceId, validator)
	if err != nil {
		vdf.Context.T.Errorf("AWS resource validation failed: %v", err)
		vdf.Context.T.FailNow()
	}
}

// ValidateAWSResourceExistsE validates that an AWS resource exists, returning error
func (vdf *VasDeference) ValidateAWSResourceExistsE(resourceType, resourceId string, validator func() (bool, error)) error {
	var lastErr error
	retry.DoWithRetry(
		vdf.Context.T,
		fmt.Sprintf("Validating %s %s exists", resourceType, resourceId),
		3,
		2*time.Second,
		func() (string, error) {
			exists, validErr := validator()
			if validErr != nil {
				lastErr = validErr
				return "", validErr
			}
			if !exists {
				lastErr = fmt.Errorf("%s %s does not exist", resourceType, resourceId)
				return "", lastErr
			}
			lastErr = nil
			return "exists", nil
		},
	)
	return lastErr
}

// CreateResourceNameWithTimestamp creates a resource name with timestamp for uniqueness
func (vdf *VasDeference) CreateResourceNameWithTimestamp(prefix string) string {
	return fmt.Sprintf("%s-%s-%s", 
		vdf.Arrange.Namespace, 
		prefix, 
		strings.ReplaceAll(random.UniqueId(), "-", ""))
}

// LogTerraformOutput logs all Terraform outputs for debugging
func (vdf *VasDeference) LogTerraformOutput(options *terraform.Options) {
	allOutputs := terraform.OutputAll(vdf.Context.T, options)
	vdf.Context.T.Logf("Terraform outputs:")
	for key, value := range allOutputs {
		vdf.Context.T.Logf("  %s: %v", key, value)
	}
}