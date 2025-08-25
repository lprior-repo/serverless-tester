package eventbridge

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// CreateRule creates an EventBridge rule and panics on error (Terratest pattern)
func CreateRule(ctx *TestContext, config RuleConfig) RuleResult {
	result, err := CreateRuleE(ctx, config)
	if err != nil {
		ctx.T.Errorf("CreateRule failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// CreateRuleE creates an EventBridge rule and returns error
func CreateRuleE(ctx *TestContext, config RuleConfig) (RuleResult, error) {
	startTime := time.Now()
	
	logOperation("create_rule", map[string]interface{}{
		"rule_name":    config.Name,
		"event_bus":    config.EventBusName,
		"description":  config.Description,
	})
	
	// Validate rule configuration
	if config.Name == "" {
		err := fmt.Errorf("rule name cannot be empty")
		logResult("create_rule", false, time.Since(startTime), err)
		return RuleResult{}, err
	}
	
	// Validate event pattern if provided
	if config.EventPattern != "" && !ValidateEventPattern(config.EventPattern) {
		err := fmt.Errorf("%w: %s", ErrInvalidEventPattern, config.EventPattern)
		logResult("create_rule", false, time.Since(startTime), err)
		return RuleResult{}, err
	}
	
	// Set default event bus if not specified
	eventBusName := config.EventBusName
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	// Set default state if not specified
	state := config.State
	if state == "" {
		state = RuleStateEnabled
	}
	
	client := createEventBridgeClient(ctx)
	
	// Create the rule
	input := &eventbridge.PutRuleInput{
		Name:         &config.Name,
		Description:  &config.Description,
		EventBusName: &eventBusName,
		State:        state,
	}
	
	// Set either event pattern or schedule expression
	if config.EventPattern != "" {
		input.EventPattern = &config.EventPattern
	}
	if config.ScheduleExpression != "" {
		input.ScheduleExpression = &config.ScheduleExpression
	}
	
	// Add tags if specified
	if len(config.Tags) > 0 {
		tags := make([]types.Tag, 0, len(config.Tags))
		for key, value := range config.Tags {
			tags = append(tags, types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}
		input.Tags = tags
	}
	
	// Execute with retry logic
	retryConfig := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < retryConfig.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, retryConfig)
			time.Sleep(delay)
		}
		
		response, err := client.PutRule(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		result := RuleResult{
			Name:               config.Name,
			RuleArn:            safeString(response.RuleArn),
			State:        state,
			Description:        config.Description,
			EventBusName:       eventBusName,
			EventPattern:       config.EventPattern,
			ScheduleExpression: config.ScheduleExpression,
		}
		
		logResult("create_rule", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("create_rule", false, time.Since(startTime), lastErr)
	return RuleResult{}, fmt.Errorf("failed to create rule after %d attempts: %w", retryConfig.MaxAttempts, lastErr)
}

// DeleteRule deletes an EventBridge rule and panics on error (Terratest pattern)
func DeleteRule(ctx *TestContext, ruleName string, eventBusName string) {
	err := DeleteRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		ctx.T.Errorf("DeleteRule failed: %v", err)
		ctx.T.FailNow()
	}
}

// DeleteRuleE deletes an EventBridge rule and returns error
func DeleteRuleE(ctx *TestContext, ruleName string, eventBusName string) error {
	startTime := time.Now()
	
	logOperation("delete_rule", map[string]interface{}{
		"rule_name":  ruleName,
		"event_bus":  eventBusName,
	})
	
	if ruleName == "" {
		err := fmt.Errorf("rule name cannot be empty")
		logResult("delete_rule", false, time.Since(startTime), err)
		return err
	}
	
	// Set default event bus if not specified
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	client := createEventBridgeClient(ctx)
	
	// First, remove all targets from the rule
	if err := RemoveAllTargetsE(ctx, ruleName, eventBusName); err != nil {
		logResult("delete_rule", false, time.Since(startTime), err)
		return fmt.Errorf("failed to remove targets before deleting rule: %w", err)
	}
	
	// Execute with retry logic
	config := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, config)
			time.Sleep(delay)
		}
		
		input := &eventbridge.DeleteRuleInput{
			Name:         &ruleName,
			EventBusName: &eventBusName,
		}
		
		_, err := client.DeleteRule(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		logResult("delete_rule", true, time.Since(startTime), nil)
		return nil
	}
	
	logResult("delete_rule", false, time.Since(startTime), lastErr)
	return fmt.Errorf("failed to delete rule after %d attempts: %w", config.MaxAttempts, lastErr)
}

// DescribeRule describes an EventBridge rule and panics on error (Terratest pattern)
func DescribeRule(ctx *TestContext, ruleName string, eventBusName string) RuleResult {
	result, err := DescribeRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		ctx.T.Errorf("DescribeRule failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// DescribeRuleE describes an EventBridge rule and returns error
func DescribeRuleE(ctx *TestContext, ruleName string, eventBusName string) (RuleResult, error) {
	startTime := time.Now()
	
	logOperation("describe_rule", map[string]interface{}{
		"rule_name":  ruleName,
		"event_bus":  eventBusName,
	})
	
	if ruleName == "" {
		err := fmt.Errorf("rule name cannot be empty")
		logResult("describe_rule", false, time.Since(startTime), err)
		return RuleResult{}, err
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
		
		input := &eventbridge.DescribeRuleInput{
			Name:         &ruleName,
			EventBusName: &eventBusName,
		}
		
		response, err := client.DescribeRule(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		result := RuleResult{
			Name:               safeString(response.Name),
			RuleArn:            safeString(response.Arn),
			State:              response.State,
			Description:        safeString(response.Description),
			EventBusName:       safeString(response.EventBusName),
			EventPattern:       safeString(response.EventPattern),
			ScheduleExpression: safeString(response.ScheduleExpression),
		}
		
		logResult("describe_rule", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("describe_rule", false, time.Since(startTime), lastErr)
	return RuleResult{}, fmt.Errorf("failed to describe rule after %d attempts: %w", config.MaxAttempts, lastErr)
}

// EnableRule enables an EventBridge rule and panics on error (Terratest pattern)
func EnableRule(ctx *TestContext, ruleName string, eventBusName string) {
	err := EnableRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		ctx.T.Errorf("EnableRule failed: %v", err)
		ctx.T.FailNow()
	}
}

// EnableRuleE enables an EventBridge rule and returns error
func EnableRuleE(ctx *TestContext, ruleName string, eventBusName string) error {
	return updateRuleStateE(ctx, ruleName, eventBusName, RuleStateEnabled)
}

// DisableRule disables an EventBridge rule and panics on error (Terratest pattern)
func DisableRule(ctx *TestContext, ruleName string, eventBusName string) {
	err := DisableRuleE(ctx, ruleName, eventBusName)
	if err != nil {
		ctx.T.Errorf("DisableRule failed: %v", err)
		ctx.T.FailNow()
	}
}

// DisableRuleE disables an EventBridge rule and returns error
func DisableRuleE(ctx *TestContext, ruleName string, eventBusName string) error {
	return updateRuleStateE(ctx, ruleName, eventBusName, RuleStateDisabled)
}

// ListRules lists EventBridge rules and panics on error (Terratest pattern)
func ListRules(ctx *TestContext, eventBusName string) []RuleResult {
	results, err := ListRulesE(ctx, eventBusName)
	if err != nil {
		ctx.T.Errorf("ListRules failed: %v", err)
		ctx.T.FailNow()
	}
	return results
}

// ListRulesE lists EventBridge rules and returns error
func ListRulesE(ctx *TestContext, eventBusName string) ([]RuleResult, error) {
	startTime := time.Now()
	
	logOperation("list_rules", map[string]interface{}{
		"event_bus": eventBusName,
	})
	
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
		
		input := &eventbridge.ListRulesInput{
			EventBusName: &eventBusName,
		}
		
		response, err := client.ListRules(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Convert results
		results := make([]RuleResult, len(response.Rules))
		for i, rule := range response.Rules {
			results[i] = RuleResult{
				Name:               safeString(rule.Name),
				RuleArn:            safeString(rule.Arn),
				State:        rule.State,
				Description:        safeString(rule.Description),
				EventBusName:       safeString(rule.EventBusName),
				EventPattern:       safeString(rule.EventPattern),
				ScheduleExpression: safeString(rule.ScheduleExpression),
			}
		}
		
		logResult("list_rules", true, time.Since(startTime), nil)
		return results, nil
	}
	
	logResult("list_rules", false, time.Since(startTime), lastErr)
	return nil, fmt.Errorf("failed to list rules after %d attempts: %w", config.MaxAttempts, lastErr)
}

// updateRuleStateE updates the state of an EventBridge rule
func updateRuleStateE(ctx *TestContext, ruleName string, eventBusName string, state types.RuleState) error {
	startTime := time.Now()
	operation := "enable_rule"
	if state == RuleStateDisabled {
		operation = "disable_rule"
	}
	
	logOperation(operation, map[string]interface{}{
		"rule_name":  ruleName,
		"event_bus":  eventBusName,
		"state":      string(state),
	})
	
	if ruleName == "" {
		err := fmt.Errorf("rule name cannot be empty")
		logResult(operation, false, time.Since(startTime), err)
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
		
		var err error
		if state == RuleStateEnabled {
			input := &eventbridge.EnableRuleInput{
				Name:         &ruleName,
				EventBusName: &eventBusName,
			}
			_, err = client.EnableRule(context.Background(), input)
		} else {
			input := &eventbridge.DisableRuleInput{
				Name:         &ruleName,
				EventBusName: &eventBusName,
			}
			_, err = client.DisableRule(context.Background(), input)
		}
		
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		logResult(operation, true, time.Since(startTime), nil)
		return nil
	}
	
	logResult(operation, false, time.Since(startTime), lastErr)
	return fmt.Errorf("failed to update rule state after %d attempts: %w", config.MaxAttempts, lastErr)
}

// UpdateRule updates an EventBridge rule and panics on error (Terratest pattern)
func UpdateRule(ctx *TestContext, config RuleConfig) RuleResult {
	result, err := UpdateRuleE(ctx, config)
	if err != nil {
		ctx.T.Errorf("UpdateRule failed: %v", err)
		ctx.T.FailNow()
	}
	return result
}

// UpdateRuleE updates an EventBridge rule and returns error
func UpdateRuleE(ctx *TestContext, config RuleConfig) (RuleResult, error) {
	startTime := time.Now()
	
	logOperation("update_rule", map[string]interface{}{
		"rule_name":    config.Name,
		"event_bus":    config.EventBusName,
		"description":  config.Description,
	})
	
	// Validate rule configuration
	if config.Name == "" {
		err := fmt.Errorf("rule name cannot be empty")
		logResult("update_rule", false, time.Since(startTime), err)
		return RuleResult{}, err
	}
	
	// Validate event pattern if provided
	if config.EventPattern != "" && !ValidateEventPattern(config.EventPattern) {
		err := fmt.Errorf("%w: %s", ErrInvalidEventPattern, config.EventPattern)
		logResult("update_rule", false, time.Since(startTime), err)
		return RuleResult{}, err
	}
	
	// Set default event bus if not specified
	eventBusName := config.EventBusName
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	
	// Set default state if not specified
	state := config.State
	if state == "" {
		state = RuleStateEnabled
	}
	
	client := createEventBridgeClient(ctx)
	
	// Update the rule using PutRule (same as create, but for existing rule)
	input := &eventbridge.PutRuleInput{
		Name:         &config.Name,
		Description:  &config.Description,
		EventBusName: &eventBusName,
		State:        state,
	}
	
	// Set either event pattern or schedule expression
	if config.EventPattern != "" {
		input.EventPattern = &config.EventPattern
	}
	if config.ScheduleExpression != "" {
		input.ScheduleExpression = &config.ScheduleExpression
	}
	
	// Add tags if specified
	if len(config.Tags) > 0 {
		tags := make([]types.Tag, 0, len(config.Tags))
		for key, value := range config.Tags {
			tags = append(tags, types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}
		input.Tags = tags
	}
	
	// Execute with retry logic
	retryConfig := defaultRetryConfig()
	var lastErr error
	
	for attempt := 0; attempt < retryConfig.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateBackoffDelay(attempt-1, retryConfig)
			time.Sleep(delay)
		}
		
		response, err := client.PutRule(context.Background(), input)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Success case
		result := RuleResult{
			Name:               config.Name,
			RuleArn:            safeString(response.RuleArn),
			State:        state,
			Description:        config.Description,
			EventBusName:       eventBusName,
			EventPattern:       config.EventPattern,
			ScheduleExpression: config.ScheduleExpression,
		}
		
		logResult("update_rule", true, time.Since(startTime), nil)
		return result, nil
	}
	
	logResult("update_rule", false, time.Since(startTime), lastErr)
	return RuleResult{}, fmt.Errorf("failed to update rule after %d attempts: %w", retryConfig.MaxAttempts, lastErr)
}