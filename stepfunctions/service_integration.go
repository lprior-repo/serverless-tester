// Package stepfunctions service integration - AWS service integration utilities
// Following strict functional programming principles and TDD methodology

package stepfunctions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

// Service integration types and structures

// ServiceIntegrationMetrics tracks metrics for AWS service integrations
type ServiceIntegrationMetrics struct {
	LambdaInvocations   int
	DynamoDBOperations  int
	SNSPublications     int
	SQSMessages         int
	S3Operations        int
	TotalDuration       time.Duration
	AverageStepDuration time.Duration
	FailedInvocations   int
	RetryAttempts       int
}

// ResourceDependencies maps state names to their dependencies
type ResourceDependencies map[string][]string

// ServiceResourceInfo contains extracted service resource information
type ServiceResourceInfo struct {
	LambdaFunctions  []string
	DynamoDBTables   []string
	SNSTopics        []string
	SQSQueues        []string
	S3Buckets        []string
}

// Integration validation functions

// ValidateLambdaIntegration validates Lambda service integration
func ValidateLambdaIntegration(stateMachine *StateMachineDefinition, input *Input) {
	err := ValidateLambdaIntegrationE(stateMachine, input)
	if err != nil {
		panic(fmt.Sprintf("Lambda integration validation failed: %v", err))
	}
}

// ValidateLambdaIntegrationE validates Lambda service integration with error return
func ValidateLambdaIntegrationE(stateMachine *StateMachineDefinition, input *Input) error {
	if stateMachine == nil {
		return fmt.Errorf("state machine definition is required")
	}

	if err := validateStateMachineDefinition(stateMachine.Definition); err != nil {
		return fmt.Errorf("invalid state machine definition: %w", err)
	}

	// Extract Lambda resources from definition
	lambdaFunctions := ExtractLambdaFunctions(stateMachine.Definition)
	if len(lambdaFunctions) == 0 {
		return fmt.Errorf("no Lambda functions found in state machine definition")
	}

	// Validate each Lambda function ARN
	for _, functionArn := range lambdaFunctions {
		if functionArn == "" {
			return fmt.Errorf("empty resource ARN")
		}
		
		if err := validateLambdaFunctionArn(functionArn); err != nil {
			return fmt.Errorf("invalid lambda resource ARN: %w", err)
		}
	}

	// Validate input compatibility
	if input != nil {
		if err := validateLambdaInput(input); err != nil {
			return fmt.Errorf("invalid Lambda input: %w", err)
		}
	}

	return nil
}

// ValidateDynamoDBIntegration validates DynamoDB service integration
func ValidateDynamoDBIntegration(stateMachine *StateMachineDefinition, input *Input) {
	err := ValidateDynamoDBIntegrationE(stateMachine, input)
	if err != nil {
		panic(fmt.Sprintf("DynamoDB integration validation failed: %v", err))
	}
}

// ValidateDynamoDBIntegrationE validates DynamoDB service integration with error return
func ValidateDynamoDBIntegrationE(stateMachine *StateMachineDefinition, input *Input) error {
	if stateMachine == nil {
		return fmt.Errorf("state machine definition is required")
	}

	if err := validateStateMachineDefinition(stateMachine.Definition); err != nil {
		return fmt.Errorf("invalid state machine definition: %w", err)
	}

	// Parse state machine definition
	var definition map[string]interface{}
	if err := json.Unmarshal([]byte(stateMachine.Definition), &definition); err != nil {
		return fmt.Errorf("failed to parse state machine definition: %w", err)
	}

	// Validate DynamoDB operations
	if err := validateDynamoDBOperations(definition); err != nil {
		return err
	}

	// Validate DynamoDB tables exist
	tables := ExtractDynamoDBTables(stateMachine.Definition)
	if len(tables) == 0 {
		return fmt.Errorf("no DynamoDB tables found in state machine definition")
	}

	// Validate input contains required fields for DynamoDB operations
	if input != nil {
		if err := validateDynamoDBInput(input, definition); err != nil {
			return fmt.Errorf("invalid DynamoDB input: %w", err)
		}
	}

	return nil
}

// ValidateSNSIntegration validates SNS service integration
func ValidateSNSIntegration(stateMachine *StateMachineDefinition, input *Input) {
	err := ValidateSNSIntegrationE(stateMachine, input)
	if err != nil {
		panic(fmt.Sprintf("SNS integration validation failed: %v", err))
	}
}

// ValidateSNSIntegrationE validates SNS service integration with error return
func ValidateSNSIntegrationE(stateMachine *StateMachineDefinition, input *Input) error {
	if stateMachine == nil {
		return fmt.Errorf("state machine definition is required")
	}

	if err := validateStateMachineDefinition(stateMachine.Definition); err != nil {
		return fmt.Errorf("invalid state machine definition: %w", err)
	}

	// Parse state machine definition
	var definition map[string]interface{}
	if err := json.Unmarshal([]byte(stateMachine.Definition), &definition); err != nil {
		return fmt.Errorf("failed to parse state machine definition: %w", err)
	}

	// Validate SNS operations
	if err := validateSNSOperations(definition); err != nil {
		return err
	}

	// Validate SNS topics
	topics := ExtractSNSTopics(stateMachine.Definition)
	if len(topics) == 0 {
		return fmt.Errorf("no SNS topics found in state machine definition")
	}

	for _, topicArn := range topics {
		if err := validateSNSTopicArn(topicArn); err != nil {
			return fmt.Errorf("invalid SNS topic ARN: %w", err)
		}
	}

	// Validate input for SNS operations
	if input != nil {
		if err := validateSNSInput(input, definition); err != nil {
			return fmt.Errorf("invalid SNS input: %w", err)
		}
	}

	return nil
}

// ValidateComplexWorkflow validates complex multi-service workflow
func ValidateComplexWorkflow(stateMachine *StateMachineDefinition, input *Input) {
	err := ValidateComplexWorkflowE(stateMachine, input)
	if err != nil {
		panic(fmt.Sprintf("Complex workflow validation failed: %v", err))
	}
}

// ValidateComplexWorkflowE validates complex multi-service workflow with error return
func ValidateComplexWorkflowE(stateMachine *StateMachineDefinition, input *Input) error {
	if stateMachine == nil {
		return fmt.Errorf("state machine definition is required")
	}

	if err := validateStateMachineDefinition(stateMachine.Definition); err != nil {
		return fmt.Errorf("invalid state machine definition: %w", err)
	}

	// Validate all service integrations present in the workflow
	if err := ValidateLambdaIntegrationE(stateMachine, input); err != nil {
		// Lambda validation is optional for complex workflows
		if !strings.Contains(err.Error(), "no Lambda functions found") {
			return fmt.Errorf("Lambda integration validation failed: %w", err)
		}
	}

	if err := ValidateDynamoDBIntegrationE(stateMachine, input); err != nil {
		// DynamoDB validation is optional for complex workflows
		if !strings.Contains(err.Error(), "no DynamoDB tables found") {
			return fmt.Errorf("DynamoDB integration validation failed: %w", err)
		}
	}

	if err := ValidateSNSIntegrationE(stateMachine, input); err != nil {
		// SNS validation is optional for complex workflows
		if !strings.Contains(err.Error(), "no SNS topics found") {
			return fmt.Errorf("SNS integration validation failed: %w", err)
		}
	}

	// Validate required input fields for complex workflow
	if input != nil {
		if err := validateComplexWorkflowInput(input); err != nil {
			return err
		}
	}

	// Validate resource dependencies
	if _, err := AnalyzeResourceDependenciesE(stateMachine.Definition); err != nil {
		return fmt.Errorf("failed to analyze resource dependencies: %w", err)
	}

	return nil
}

// Resource extraction functions

// ExtractLambdaFunctions extracts Lambda function ARNs from state machine definition
func ExtractLambdaFunctions(definition string) []string {
	var functions []string
	
	// Pattern to match Lambda function ARNs
	lambdaPattern := regexp.MustCompile(`arn:aws:lambda:[^"]+`)
	matches := lambdaPattern.FindAllString(definition, -1)
	
	// Remove duplicates
	seen := make(map[string]bool)
	for _, match := range matches {
		if !seen[match] {
			functions = append(functions, match)
			seen[match] = true
		}
	}
	
	return functions
}

// ExtractDynamoDBTables extracts DynamoDB table names from state machine definition
func ExtractDynamoDBTables(definition string) []string {
	var tables []string
	
	// Parse the definition to find DynamoDB operations
	var definitionMap map[string]interface{}
	if err := json.Unmarshal([]byte(definition), &definitionMap); err != nil {
		return tables
	}
	
	// Extract table names from Parameters.TableName fields
	tableNames := extractTableNamesFromDefinition(definitionMap)
	
	return tableNames
}

// ExtractSNSTopics extracts SNS topic ARNs from state machine definition
func ExtractSNSTopics(definition string) []string {
	var topics []string
	
	// Parse the definition to find SNS operations
	var definitionMap map[string]interface{}
	if err := json.Unmarshal([]byte(definition), &definitionMap); err != nil {
		return topics
	}
	
	// Extract topic ARNs from Parameters.TopicArn fields
	topicArns := extractTopicArnsFromDefinition(definitionMap)
	
	return topicArns
}

// AnalyzeResourceDependencies analyzes resource dependencies in workflow
func AnalyzeResourceDependencies(definition string) ResourceDependencies {
	deps, err := AnalyzeResourceDependenciesE(definition)
	if err != nil {
		return ResourceDependencies{}
	}
	return deps
}

// AnalyzeResourceDependenciesE analyzes resource dependencies with error return
func AnalyzeResourceDependenciesE(definition string) (ResourceDependencies, error) {
	dependencies := make(ResourceDependencies)
	
	// Parse state machine definition
	var definitionMap map[string]interface{}
	if err := json.Unmarshal([]byte(definition), &definitionMap); err != nil {
		return nil, fmt.Errorf("failed to parse state machine definition: %w", err)
	}
	
	// Extract states
	states, exists := definitionMap["States"]
	if !exists {
		return dependencies, nil
	}
	
	statesMap, ok := states.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid States format")
	}
	
	// Build dependency graph
	for stateName := range statesMap {
		dependencies[stateName] = findStateDependencies(stateName, statesMap)
	}
	
	return dependencies, nil
}

// CollectServiceIntegrationMetrics collects metrics from execution history
func CollectServiceIntegrationMetrics(history []HistoryEvent) ServiceIntegrationMetrics {
	metrics := ServiceIntegrationMetrics{}
	
	if len(history) == 0 {
		return metrics
	}
	
	// Calculate total duration
	var startTime, endTime time.Time
	for _, event := range history {
		switch event.Type {
		case types.HistoryEventTypeExecutionStarted:
			if startTime.IsZero() || event.Timestamp.Before(startTime) {
				startTime = event.Timestamp
			}
		case types.HistoryEventTypeExecutionSucceeded,
			 types.HistoryEventTypeExecutionFailed,
			 types.HistoryEventTypeExecutionTimedOut,
			 types.HistoryEventTypeExecutionAborted:
			if endTime.IsZero() || event.Timestamp.After(endTime) {
				endTime = event.Timestamp
			}
		}
	}
	
	if !startTime.IsZero() && !endTime.IsZero() {
		metrics.TotalDuration = endTime.Sub(startTime)
	}
	
	// Count service invocations
	var totalSteps int
	var totalStepDuration time.Duration
	
	for _, event := range history {
		switch event.Type {
		case types.HistoryEventTypeTaskSucceeded:
			totalSteps++
			if event.TaskSucceededEventDetails != nil {
				resource := event.TaskSucceededEventDetails.Resource
				if strings.Contains(resource, "lambda") {
					metrics.LambdaInvocations++
				} else if strings.Contains(resource, "dynamodb") {
					metrics.DynamoDBOperations++
				} else if strings.Contains(resource, "sns") {
					metrics.SNSPublications++
				}
			}
		case types.HistoryEventTypeTaskFailed:
			metrics.FailedInvocations++
			if event.TaskFailedEventDetails != nil {
				resource := event.TaskFailedEventDetails.Resource
				if strings.Contains(resource, "lambda") {
					metrics.LambdaInvocations++
				}
			}
		case types.HistoryEventTypeLambdaFunctionSucceeded:
			metrics.LambdaInvocations++
			totalSteps++
		case types.HistoryEventTypeLambdaFunctionFailed:
			metrics.LambdaInvocations++
			metrics.FailedInvocations++
		}
	}
	
	// Calculate average step duration
	if totalSteps > 0 {
		metrics.AverageStepDuration = totalStepDuration / time.Duration(totalSteps)
	}
	
	// Count retry attempts
	metrics.RetryAttempts = countRetryAttempts(history)
	
	return metrics
}

// Internal validation helper functions

// validateLambdaFunctionArn validates Lambda function ARN format
func validateLambdaFunctionArn(functionArn string) error {
	if functionArn == "" {
		return fmt.Errorf("Lambda function ARN cannot be empty")
	}
	
	if !strings.HasPrefix(functionArn, "arn:aws:lambda:") {
		return fmt.Errorf("invalid Lambda function ARN format")
	}
	
	if !strings.Contains(functionArn, ":function:") {
		return fmt.Errorf("Lambda ARN must contain ':function:'")
	}
	
	return nil
}

// validateLambdaInput validates input for Lambda integration
func validateLambdaInput(input *Input) error {
	if input == nil || input.isEmpty() {
		return nil // Input is optional for Lambda
	}
	
	// Validate that input can be serialized to JSON
	if _, err := input.ToJSON(); err != nil {
		return fmt.Errorf("input cannot be serialized to JSON: %w", err)
	}
	
	return nil
}

// validateDynamoDBOperations validates DynamoDB operations in definition
func validateDynamoDBOperations(definition map[string]interface{}) error {
	states, exists := definition["States"]
	if !exists {
		return fmt.Errorf("no States found in definition")
	}
	
	statesMap, ok := states.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid States format")
	}
	
	for stateName, state := range statesMap {
		stateMap, ok := state.(map[string]interface{})
		if !ok {
			continue
		}
		
		resource, exists := stateMap["Resource"]
		if !exists {
			continue
		}
		
		resourceStr, ok := resource.(string)
		if !ok {
			continue
		}
		
		if strings.Contains(resourceStr, "dynamodb:") {
			if err := validateDynamoDBOperation(stateName, stateMap); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// validateDynamoDBOperation validates a single DynamoDB operation
func validateDynamoDBOperation(stateName string, stateMap map[string]interface{}) error {
	resource := stateMap["Resource"].(string)
	
	// Validate supported operations
	supportedOps := []string{"putItem", "getItem", "updateItem", "deleteItem", "scan", "query"}
	isSupported := false
	for _, op := range supportedOps {
		if strings.Contains(resource, op) {
			isSupported = true
			break
		}
	}
	
	if !isSupported {
		return fmt.Errorf("unsupported DynamoDB operation in state %s", stateName)
	}
	
	// Check for required Parameters
	parameters, exists := stateMap["Parameters"]
	if !exists {
		return fmt.Errorf("missing Parameters for DynamoDB operation in state %s", stateName)
	}
	
	paramsMap, ok := parameters.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid Parameters format in state %s", stateName)
	}
	
	// Check for TableName
	if _, exists := paramsMap["TableName"]; !exists {
		return fmt.Errorf("missing TableName parameter in state %s", stateName)
	}
	
	return nil
}

// validateDynamoDBInput validates input for DynamoDB operations
func validateDynamoDBInput(input *Input, definition map[string]interface{}) error {
	if input == nil || input.isEmpty() {
		return fmt.Errorf("input is required for DynamoDB operations")
	}
	
	// Basic validation - input should be valid JSON
	if _, err := input.ToJSON(); err != nil {
		return fmt.Errorf("input cannot be serialized to JSON: %w", err)
	}
	
	return nil
}

// validateSNSOperations validates SNS operations in definition
func validateSNSOperations(definition map[string]interface{}) error {
	states, exists := definition["States"]
	if !exists {
		return fmt.Errorf("no States found in definition")
	}
	
	statesMap, ok := states.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid States format")
	}
	
	for stateName, state := range statesMap {
		stateMap, ok := state.(map[string]interface{})
		if !ok {
			continue
		}
		
		resource, exists := stateMap["Resource"]
		if !exists {
			continue
		}
		
		resourceStr, ok := resource.(string)
		if !ok {
			continue
		}
		
		if strings.Contains(resourceStr, "sns:") {
			if err := validateSNSOperation(stateName, stateMap); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// validateSNSOperation validates a single SNS operation
func validateSNSOperation(stateName string, stateMap map[string]interface{}) error {
	// Check for required Parameters
	parameters, exists := stateMap["Parameters"]
	if !exists {
		return fmt.Errorf("missing Parameters for SNS operation in state %s", stateName)
	}
	
	paramsMap, ok := parameters.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid Parameters format in state %s", stateName)
	}
	
	// Check for TopicArn
	if _, exists := paramsMap["TopicArn"]; !exists {
		return fmt.Errorf("missing TopicArn parameter in state %s", stateName)
	}
	
	return nil
}

// validateSNSTopicArn validates SNS topic ARN format
func validateSNSTopicArn(topicArn string) error {
	if topicArn == "" {
		return fmt.Errorf("SNS topic ARN cannot be empty")
	}
	
	if !strings.HasPrefix(topicArn, "arn:aws:sns:") {
		return fmt.Errorf("invalid SNS topic ARN format")
	}
	
	return nil
}

// validateSNSInput validates input for SNS operations
func validateSNSInput(input *Input, definition map[string]interface{}) error {
	if input == nil || input.isEmpty() {
		return fmt.Errorf("input is required for SNS operations")
	}
	
	// Validate that input can be serialized to JSON
	if _, err := input.ToJSON(); err != nil {
		return fmt.Errorf("input cannot be serialized to JSON: %w", err)
	}
	
	return nil
}

// validateComplexWorkflowInput validates input for complex workflows
func validateComplexWorkflowInput(input *Input) error {
	if input == nil || input.isEmpty() {
		return fmt.Errorf("input is required for complex workflows")
	}
	
	// Check for required fields
	if _, exists := input.Get("id"); !exists {
		return fmt.Errorf("missing required input field 'id'")
	}
	
	return nil
}

// Helper functions for resource extraction

// extractTableNamesFromDefinition recursively extracts DynamoDB table names
func extractTableNamesFromDefinition(obj interface{}) []string {
	var tables []string
	seen := make(map[string]bool)
	
	switch v := obj.(type) {
	case map[string]interface{}:
		// Check if this is a TableName field
		if tableName, exists := v["TableName"]; exists {
			if tableStr, ok := tableName.(string); ok && !seen[tableStr] {
				tables = append(tables, tableStr)
				seen[tableStr] = true
			}
		}
		
		// Recursively check all values
		for _, value := range v {
			subTables := extractTableNamesFromDefinition(value)
			for _, table := range subTables {
				if !seen[table] {
					tables = append(tables, table)
					seen[table] = true
				}
			}
		}
	case []interface{}:
		// Recursively check array elements
		for _, value := range v {
			subTables := extractTableNamesFromDefinition(value)
			for _, table := range subTables {
				if !seen[table] {
					tables = append(tables, table)
					seen[table] = true
				}
			}
		}
	}
	
	return tables
}

// extractTopicArnsFromDefinition recursively extracts SNS topic ARNs
func extractTopicArnsFromDefinition(obj interface{}) []string {
	var topics []string
	seen := make(map[string]bool)
	
	switch v := obj.(type) {
	case map[string]interface{}:
		// Check if this is a TopicArn field
		if topicArn, exists := v["TopicArn"]; exists {
			if topicStr, ok := topicArn.(string); ok && !seen[topicStr] {
				topics = append(topics, topicStr)
				seen[topicStr] = true
			}
		}
		
		// Recursively check all values
		for _, value := range v {
			subTopics := extractTopicArnsFromDefinition(value)
			for _, topic := range subTopics {
				if !seen[topic] {
					topics = append(topics, topic)
					seen[topic] = true
				}
			}
		}
	case []interface{}:
		// Recursively check array elements
		for _, value := range v {
			subTopics := extractTopicArnsFromDefinition(value)
			for _, topic := range subTopics {
				if !seen[topic] {
					topics = append(topics, topic)
					seen[topic] = true
				}
			}
		}
	}
	
	return topics
}

// findStateDependencies finds dependencies for a specific state
func findStateDependencies(stateName string, statesMap map[string]interface{}) []string {
	var dependencies []string
	
	// Find all states that point to this state via "Next"
	for name, state := range statesMap {
		if name == stateName {
			continue
		}
		
		stateMap, ok := state.(map[string]interface{})
		if !ok {
			continue
		}
		
		if next, exists := stateMap["Next"]; exists {
			if nextStr, ok := next.(string); ok && nextStr == stateName {
				dependencies = append(dependencies, name)
			}
		}
		
		// Also check choice states and parallel branches
		if stateType, exists := stateMap["Type"]; exists {
			if typeStr, ok := stateType.(string); ok && typeStr == "Choice" {
				if choices, exists := stateMap["Choices"]; exists {
					if choicesArray, ok := choices.([]interface{}); ok {
						for _, choice := range choicesArray {
							if choiceMap, ok := choice.(map[string]interface{}); ok {
								if next, exists := choiceMap["Next"]; exists {
									if nextStr, ok := next.(string); ok && nextStr == stateName {
										dependencies = append(dependencies, name)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	return dependencies
}