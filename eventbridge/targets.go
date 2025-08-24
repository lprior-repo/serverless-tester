package eventbridge

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// PutTargets adds targets to an EventBridge rule and panics on error (Terratest pattern)
func PutTargets(ctx *TestContext, ruleName string, eventBusName string, targets []TargetConfig) {
	err := PutTargetsE(ctx, ruleName, eventBusName, targets)
	if err != nil {
		ctx.T.Errorf("PutTargets failed: %v", err)
		ctx.T.FailNow()
	}
}

// PutTargetsE adds targets to an EventBridge rule and returns error
func PutTargetsE(ctx *TestContext, ruleName string, eventBusName string, targets []TargetConfig) error {
	startTime := time.Now()
	
	logOperation("put_targets", map[string]interface{}{
		"rule_name":    ruleName,
		"event_bus":    eventBusName,
		"target_count": len(targets),
	})
	
	if ruleName == "" {
		err := fmt.Errorf("rule name cannot be empty")
		logResult("put_targets", false, time.Since(startTime), err)
		return err
	}
	
	if len(targets) == 0 {
		err := fmt.Errorf("at least one target must be specified")
		logResult("put_targets", false, time.Since(startTime), err)
		return err
	}
	
	// Set default event bus if not specified
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	client := createEventBridgeClient(ctx)
	
	// Convert targets to AWS types
	awsTargets := make([]types.Target, 0, len(targets))
	for _, target := range targets {
		if target.ID == "" || target.Arn == "" {
			err := fmt.Errorf("target ID and ARN are required")
			logResult("put_targets", false, time.Since(startTime), err)
			return err
		}
		
		awsTarget := types.Target{
			Id:  &target.ID,
			Arn: &target.Arn,
		}
		
		// Add optional fields
		if target.RoleArn != "" {
			awsTarget.RoleArn = &target.RoleArn
		}
		if target.Input != "" {
			awsTarget.Input = &target.Input
		}
		if target.InputPath != "" {
			awsTarget.InputPath = &target.InputPath
		}
		if target.InputTransformer != nil {
			awsTarget.InputTransformer = target.InputTransformer
		}
		if target.DeadLetterConfig != nil {
			awsTarget.DeadLetterConfig = target.DeadLetterConfig
		}
		if target.RetryPolicy != nil {
			awsTarget.RetryPolicy = target.RetryPolicy
		}
		
		awsTargets = append(awsTargets, awsTarget)
	}
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.PutTargetsInput{
			Rule:         &ruleName,
			EventBusName: &eventBusName,
			Targets:      awsTargets,
		}
		
		response, err := client.PutTargets(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Check for failed entries
		if response.FailedEntryCount > 0 {
			failedTargets := make([]string, 0)
			for _, entry := range response.FailedEntries {
				targetInfo := fmt.Sprintf("ID=%s, Error=%s", 
					safeString(entry.TargetId), 
					safeString(entry.ErrorMessage))
				failedTargets = append(failedTargets, targetInfo)
			}
			
			err := fmt.Errorf("failed to add %d targets: %v", response.FailedEntryCount, failedTargets)
			logResult("put_targets", false, time.Since(startTime), err)
			return err
		}
		
		// Success case
		logResult("put_targets", true, time.Since(startTime), nil)
		return nil
	}
	
	logResult("put_targets", false, time.Since(startTime), lastErr)
	return fmt.Errorf("failed to put targets after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RemoveTargets removes targets from an EventBridge rule and panics on error (Terratest pattern)
func RemoveTargets(ctx *TestContext, ruleName string, eventBusName string, targetIDs []string) {
	err := RemoveTargetsE(ctx, ruleName, eventBusName, targetIDs)
	if err != nil {
		ctx.T.Errorf("RemoveTargets failed: %v", err)
		ctx.T.FailNow()
	}
}

// RemoveTargetsE removes targets from an EventBridge rule and returns error
func RemoveTargetsE(ctx *TestContext, ruleName string, eventBusName string, targetIDs []string) error {
	startTime := time.Now()
	
	logOperation("remove_targets", map[string]interface{}{
		"rule_name":    ruleName,
		"event_bus":    eventBusName,
		"target_count": len(targetIDs),
	})
	
	if ruleName == "" {
		err := fmt.Errorf("rule name cannot be empty")
		logResult("remove_targets", false, time.Since(startTime), err)
		return err
	}
	
	if len(targetIDs) == 0 {
		err := fmt.Errorf("at least one target ID must be specified")
		logResult("remove_targets", false, time.Since(startTime), err)
		return err
	}
	
	// Set default event bus if not specified
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.RemoveTargetsInput{
			Rule:         &ruleName,
			EventBusName: &eventBusName,
			Ids:          targetIDs,
		}
		
		response, err := client.RemoveTargets(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Check for failed entries
		if response.FailedEntryCount > 0 {
			failedTargets := make([]string, 0)
			for _, entry := range response.FailedEntries {
				targetInfo := fmt.Sprintf("ID=%s, Error=%s", 
					safeString(entry.TargetId), 
					safeString(entry.ErrorMessage))
				failedTargets = append(failedTargets, targetInfo)
			}
			
			err := fmt.Errorf("failed to remove %d targets: %v", response.FailedEntryCount, failedTargets)
			logResult("remove_targets", false, time.Since(startTime), err)
			return err
		}
		
		// Success case
		logResult("remove_targets", true, time.Since(startTime), nil)
		return nil
	}
	
	logResult("remove_targets", false, time.Since(startTime), lastErr)
	return fmt.Errorf("failed to remove targets after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RemoveAllTargets removes all targets from an EventBridge rule and panics on error (Terratest pattern)
func RemoveAllTargets(ctx *TestContext, ruleName string, eventBusName string) {
	err := RemoveAllTargetsE(ctx, ruleName, eventBusName)
	if err != nil {
		ctx.T.Errorf("RemoveAllTargets failed: %v", err)
		ctx.T.FailNow()
	}
}

// RemoveAllTargetsE removes all targets from an EventBridge rule and returns error
func RemoveAllTargetsE(ctx *TestContext, ruleName string, eventBusName string) error {
	startTime := time.Now()
	
	logOperation("remove_all_targets", map[string]interface{}{
		"rule_name": ruleName,
		"event_bus": eventBusName,
	})
	
	// First, list all targets for the rule
	targets, err := ListTargetsForRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		logResult("remove_all_targets", false, time.Since(startTime), err)
		return fmt.Errorf("failed to list targets: %w", err)
	}
	
	if len(targets) == 0 {
		// No targets to remove
		logResult("remove_all_targets", true, time.Since(startTime), nil)
		return nil
	}
	
	// Extract target IDs
	targetIDs := make([]string, len(targets))
	for i, target := range targets {
		targetIDs[i] = target.ID
	}
	
	// Remove all targets
	return RemoveTargetsE(ctx, ruleName, eventBusName, targetIDs)
}

// ListTargetsForRule lists all targets for an EventBridge rule and panics on error (Terratest pattern)
func ListTargetsForRule(ctx *TestContext, ruleName string, eventBusName string) []TargetConfig {
	results, err := ListTargetsForRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		ctx.T.Errorf("ListTargetsForRule failed: %v", err)
		ctx.T.FailNow()
	}
	return results
}

// ListTargetsForRuleE lists all targets for an EventBridge rule and returns error
func ListTargetsForRuleE(ctx *TestContext, ruleName string, eventBusName string) ([]TargetConfig, error) {
	startTime := time.Now()
	
	logOperation("list_targets_for_rule", map[string]interface{}{
		"rule_name": ruleName,
		"event_bus": eventBusName,
	})
	
	if ruleName == "" {
		err := fmt.Errorf("rule name cannot be empty")
		logResult("list_targets_for_rule", false, time.Since(startTime), err)
		return nil, err
	}
	
	// Set default event bus if not specified
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	client := createEventBridgeClient(ctx)
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.ListTargetsByRuleInput{
			Rule:         &ruleName,
			EventBusName: &eventBusName,
		}
		
		response, err := client.ListTargetsByRule(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Convert results
		results := make([]TargetConfig, len(response.Targets))
		for i, target := range response.Targets {
			config := TargetConfig{
				ID:  safeString(target.Id),
				Arn: safeString(target.Arn),
			}
			
			// Add optional fields
			if target.RoleArn != nil {
				config.RoleArn = *target.RoleArn
			}
			if target.Input != nil {
				config.Input = *target.Input
			}
			if target.InputPath != nil {
				config.InputPath = *target.InputPath
			}
			if target.InputTransformer != nil {
				config.InputTransformer = target.InputTransformer
			}
			if target.DeadLetterConfig != nil {
				config.DeadLetterConfig = target.DeadLetterConfig
			}
			if target.RetryPolicy != nil {
				config.RetryPolicy = target.RetryPolicy
			}
			
			results[i] = config
		}
		
		logResult("list_targets_for_rule", true, time.Since(startTime), nil)
		return results, nil
	}
	
	logResult("list_targets_for_rule", false, time.Since(startTime), lastErr)
	return nil, fmt.Errorf("failed to list targets after %d attempts: %w", config.MaxAttempts, lastErr)
}