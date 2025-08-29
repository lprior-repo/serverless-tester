package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
)

// TestWizard provides interactive test generation
type TestWizard struct {
	outputDir   string
	projectRoot string
	templates   map[string]*TestTemplate
}

// TestTemplate represents a test file template
type TestTemplate struct {
	Name        string
	Description string
	Service     string
	Type        string
	Template    string
	Variables   map[string]string
}

// TestConfig holds configuration for generating tests
type TestConfig struct {
	Service         string
	TestType        string
	PackageName     string
	FunctionName    string
	ResourceName    string
	Region          string
	Namespace       string
	IncludeSetup    bool
	IncludeTeardown bool
	IncludeExamples bool
	CustomVars      map[string]string
}

// NewTestWizard creates a new test generation wizard
func NewTestWizard(outputDir, projectRoot string) *TestWizard {
	wizard := &TestWizard{
		outputDir:   outputDir,
		projectRoot: projectRoot,
		templates:   make(map[string]*TestTemplate),
	}
	
	wizard.loadBuiltinTemplates()
	return wizard
}

// Run starts the interactive test generation wizard
func (w *TestWizard) Run(service, testType string) error {
	fmt.Println("🧙‍♂️ SFX Test Generation Wizard")
	fmt.Println("Let's create comprehensive tests for your serverless application!")
	fmt.Println()
	
	config, err := w.gatherTestConfig(service, testType)
	if err != nil {
		return fmt.Errorf("failed to gather configuration: %w", err)
	}
	
	return w.generateTests(config)
}

// gatherTestConfig collects configuration through interactive prompts
func (w *TestWizard) gatherTestConfig(service, testType string) (*TestConfig, error) {
	config := &TestConfig{
		CustomVars: make(map[string]string),
	}
	
	// Service selection
	if service == "" {
		selectedService, err := w.selectService()
		if err != nil {
			return nil, err
		}
		config.Service = selectedService
	} else {
		config.Service = service
	}
	
	// Test type selection
	if testType == "" {
		selectedType, err := w.selectTestType()
		if err != nil {
			return nil, err
		}
		config.TestType = selectedType
	} else {
		config.TestType = testType
	}
	
	// Gather service-specific configuration
	switch config.Service {
	case "lambda":
		return w.configureLambdaTests(config)
	case "dynamodb":
		return w.configureDynamoDBTests(config)
	case "eventbridge":
		return w.configureEventBridgeTests(config)
	case "stepfunctions":
		return w.configureStepFunctionsTests(config)
	default:
		return w.configureGenericTests(config)
	}
}

// selectService prompts user to select AWS service
func (w *TestWizard) selectService() (string, error) {
	services := []string{
		"🔧 Lambda - Serverless Functions",
		"🗄️  DynamoDB - NoSQL Database", 
		"📡 EventBridge - Event Bus",
		"🔀 Step Functions - Workflows",
		"🌐 API Gateway - REST APIs",
		"📊 CloudWatch - Monitoring",
		"🏗️  CloudFormation - Infrastructure",
		"🔄 Generic - Custom Service",
	}
	
	prompt := promptui.Select{
		Label: "Select AWS Service",
		Items: services,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✓ {{ . | green }}",
		},
	}
	
	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	
	serviceMap := []string{"lambda", "dynamodb", "eventbridge", "stepfunctions", "apigateway", "cloudwatch", "cloudformation", "generic"}
	return serviceMap[index], nil
}

// selectTestType prompts user to select test type
func (w *TestWizard) selectTestType() (string, error) {
	testTypes := []string{
		"🧪 Unit Tests - Test individual functions",
		"🔗 Integration Tests - Test component interactions", 
		"🌍 E2E Tests - Test complete workflows",
		"⚡ Performance Tests - Load and benchmark tests",
		"🛡️  Security Tests - Validation and compliance",
		"📸 Snapshot Tests - Data validation tests",
	}
	
	prompt := promptui.Select{
		Label: "Select Test Type",
		Items: testTypes,
	}
	
	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	
	typeMap := []string{"unit", "integration", "e2e", "performance", "security", "snapshot"}
	return typeMap[index], nil
}

// configureLambdaTests gathers Lambda-specific configuration
func (w *TestWizard) configureLambdaTests(config *TestConfig) (*TestConfig, error) {
	// Function name
	funcPrompt := promptui.Prompt{
		Label:   "Lambda Function Name",
		Default: "my-function",
	}
	functionName, err := funcPrompt.Run()
	if err != nil {
		return nil, err
	}
	config.FunctionName = functionName
	
	// Package name
	pkgPrompt := promptui.Prompt{
		Label:   "Go Package Name",
		Default: "lambda",
	}
	packageName, err := pkgPrompt.Run()
	if err != nil {
		return nil, err
	}
	config.PackageName = packageName
	
	// Event sources
	eventSources, err := w.selectLambdaEventSources()
	if err != nil {
		return nil, err
	}
	config.CustomVars["EventSources"] = strings.Join(eventSources, ",")
	
	return w.configureCommonOptions(config)
}

// selectLambdaEventSources prompts for Lambda event sources
func (w *TestWizard) selectLambdaEventSources() ([]string, error) {
	sources := []string{
		"API Gateway",
		"S3",
		"DynamoDB Streams",
		"SQS",
		"SNS", 
		"EventBridge",
		"CloudWatch Events",
		"Kinesis",
	}
	
	var selected []string
	
	for {
		prompt := promptui.SelectWithAdd{
			Label:    "Select Event Sources (or 'Done' to finish)",
			Items:    append(sources, "Done"),
			AddLabel: "Add custom event source",
		}
		
		index, result, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		
		if result == "Done" {
			break
		}
		
		if index == -1 {
			// Custom event source
			selected = append(selected, result)
		} else if index < len(sources) {
			selected = append(selected, sources[index])
		}
	}
	
	return selected, nil
}

// configureDynamoDBTests gathers DynamoDB-specific configuration
func (w *TestWizard) configureDynamoDBTests(config *TestConfig) (*TestConfig, error) {
	// Table name
	tablePrompt := promptui.Prompt{
		Label:   "DynamoDB Table Name",
		Default: "my-table",
	}
	tableName, err := tablePrompt.Run()
	if err != nil {
		return nil, err
	}
	config.ResourceName = tableName
	
	// Package name
	pkgPrompt := promptui.Prompt{
		Label:   "Go Package Name",
		Default: "dynamodb",
	}
	packageName, err := pkgPrompt.Run()
	if err != nil {
		return nil, err
	}
	config.PackageName = packageName
	
	// Table operations
	operations, err := w.selectDynamoDBOperations()
	if err != nil {
		return nil, err
	}
	config.CustomVars["Operations"] = strings.Join(operations, ",")
	
	return w.configureCommonOptions(config)
}

// selectDynamoDBOperations prompts for DynamoDB operations to test
func (w *TestWizard) selectDynamoDBOperations() ([]string, error) {
	operations := []string{
		"PutItem - Create items",
		"GetItem - Read items",
		"UpdateItem - Update items", 
		"DeleteItem - Delete items",
		"Query - Query items",
		"Scan - Scan table",
		"BatchGetItem - Batch reads",
		"BatchWriteItem - Batch writes",
		"TransactWriteItems - Transactions",
		"TransactGetItems - Transactional reads",
	}
	
	var selected []string
	
	for {
		prompt := promptui.Select{
			Label: "Select DynamoDB Operations (or Done to finish)",
			Items: append(operations, "Done"),
		}
		
		index, result, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		
		if result == "Done" {
			break
		}
		
		opName := strings.Split(operations[index], " - ")[0]
		selected = append(selected, opName)
	}
	
	return selected, nil
}

// configureEventBridgeTests gathers EventBridge-specific configuration
func (w *TestWizard) configureEventBridgeTests(config *TestConfig) (*TestConfig, error) {
	// Event bus name
	busPrompt := promptui.Prompt{
		Label:   "EventBridge Event Bus Name",
		Default: "default",
	}
	busName, err := busPrompt.Run()
	if err != nil {
		return nil, err
	}
	config.ResourceName = busName
	
	// Package name
	pkgPrompt := promptui.Prompt{
		Label:   "Go Package Name", 
		Default: "eventbridge",
	}
	packageName, err := pkgPrompt.Run()
	if err != nil {
		return nil, err
	}
	config.PackageName = packageName
	
	// Event patterns
	patterns, err := w.gatherEventPatterns()
	if err != nil {
		return nil, err
	}
	config.CustomVars["EventPatterns"] = strings.Join(patterns, ",")
	
	return w.configureCommonOptions(config)
}

// gatherEventPatterns prompts for event patterns
func (w *TestWizard) gatherEventPatterns() ([]string, error) {
	var patterns []string
	
	for {
		prompt := promptui.Prompt{
			Label: "Enter Event Pattern Source (e.g., 'aws.s3', or 'done' to finish)",
		}
		
		pattern, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		
		if strings.ToLower(pattern) == "done" {
			break
		}
		
		patterns = append(patterns, pattern)
	}
	
	return patterns, nil
}

// configureStepFunctionsTests gathers Step Functions configuration
func (w *TestWizard) configureStepFunctionsTests(config *TestConfig) (*TestConfig, error) {
	// State machine name
	smPrompt := promptui.Prompt{
		Label:   "State Machine Name",
		Default: "my-state-machine",
	}
	stateMachine, err := smPrompt.Run()
	if err != nil {
		return nil, err
	}
	config.ResourceName = stateMachine
	
	// Package name
	pkgPrompt := promptui.Prompt{
		Label:   "Go Package Name",
		Default: "stepfunctions",
	}
	packageName, err := pkgPrompt.Run()
	if err != nil {
		return nil, err
	}
	config.PackageName = packageName
	
	return w.configureCommonOptions(config)
}

// configureGenericTests gathers generic test configuration
func (w *TestWizard) configureGenericTests(config *TestConfig) (*TestConfig, error) {
	// Resource name
	resourcePrompt := promptui.Prompt{
		Label: "Resource/Service Name",
	}
	resourceName, err := resourcePrompt.Run()
	if err != nil {
		return nil, err
	}
	config.ResourceName = resourceName
	
	// Package name
	pkgPrompt := promptui.Prompt{
		Label:   "Go Package Name",
		Default: strings.ToLower(resourceName),
	}
	packageName, err := pkgPrompt.Run()
	if err != nil {
		return nil, err
	}
	config.PackageName = packageName
	
	return w.configureCommonOptions(config)
}

// configureCommonOptions gathers common configuration options
func (w *TestWizard) configureCommonOptions(config *TestConfig) (*TestConfig, error) {
	// Region
	regionPrompt := promptui.Prompt{
		Label:   "AWS Region",
		Default: "us-east-1",
	}
	region, err := regionPrompt.Run()
	if err != nil {
		return nil, err
	}
	config.Region = region
	
	// Namespace
	namespacePrompt := promptui.Prompt{
		Label:   "Test Namespace",
		Default: "test",
	}
	namespace, err := namespacePrompt.Run()
	if err != nil {
		return nil, err
	}
	config.Namespace = namespace
	
	// Options
	setupPrompt := promptui.Confirm{
		Label:   "Include setup/teardown functions",
		Default: true,
	}
	config.IncludeSetup, err = setupPrompt.Run()
	if err != nil {
		return nil, err
	}
	
	examplesPrompt := promptui.Confirm{
		Label:   "Include example tests",
		Default: true,
	}
	config.IncludeExamples, err = examplesPrompt.Run()
	if err != nil {
		return nil, err
	}
	
	return config, nil
}

// generateTests generates test files based on configuration
func (w *TestWizard) generateTests(config *TestConfig) error {
	fmt.Printf("\n🏗️  Generating %s tests for %s...\n", config.TestType, config.Service)
	
	// Create output directory
	if err := os.MkdirAll(w.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Generate main test file
	if err := w.generateMainTestFile(config); err != nil {
		return fmt.Errorf("failed to generate main test file: %w", err)
	}
	
	// Generate example tests if requested
	if config.IncludeExamples {
		if err := w.generateExampleTests(config); err != nil {
			return fmt.Errorf("failed to generate example tests: %w", err)
		}
	}
	
	// Generate helper files
	if err := w.generateHelperFiles(config); err != nil {
		return fmt.Errorf("failed to generate helper files: %w", err)
	}
	
	// Generate README
	if err := w.generateReadme(config); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}
	
	fmt.Println("\n✅ Test generation completed!")
	fmt.Printf("📁 Files generated in: %s\n", w.outputDir)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Review the generated test files")
	fmt.Println("2. Update test data and assertions as needed")
	fmt.Println("3. Run tests with: go test -v")
	fmt.Println("4. Use 'sfx test run' for interactive test execution")
	
	return nil
}

// generateMainTestFile creates the main test file
func (w *TestWizard) generateMainTestFile(config *TestConfig) error {
	templateKey := fmt.Sprintf("%s_%s", config.Service, config.TestType)
	template, exists := w.templates[templateKey]
	if !exists {
		template = w.templates["generic_integration"] // fallback
	}
	
	filename := fmt.Sprintf("%s_%s_test.go", config.PackageName, config.TestType)
	filepath := filepath.Join(w.outputDir, filename)
	
	return w.renderTemplate(template, config, filepath)
}

// generateExampleTests creates example test files
func (w *TestWizard) generateExampleTests(config *TestConfig) error {
	templateKey := fmt.Sprintf("%s_examples", config.Service)
	template, exists := w.templates[templateKey]
	if !exists {
		return nil // Skip if no examples template
	}
	
	filename := fmt.Sprintf("%s_examples_test.go", config.PackageName)
	filepath := filepath.Join(w.outputDir, filename)
	
	return w.renderTemplate(template, config, filepath)
}

// generateHelperFiles creates test helper files
func (w *TestWizard) generateHelperFiles(config *TestConfig) error {
	if !config.IncludeSetup {
		return nil
	}
	
	template := w.templates["helpers"]
	filename := "helpers_test.go"
	filepath := filepath.Join(w.outputDir, filename)
	
	return w.renderTemplate(template, config, filepath)
}

// generateReadme creates a README file for the tests
func (w *TestWizard) generateReadme(config *TestConfig) error {
	template := w.templates["readme"]
	filename := "README.md"
	filepath := filepath.Join(w.outputDir, filename)
	
	return w.renderTemplate(template, config, filepath)
}

// renderTemplate renders a template with configuration and saves to file
func (w *TestWizard) renderTemplate(tmpl *TestTemplate, config *TestConfig, filepath string) error {
	t, err := template.New(tmpl.Name).Parse(tmpl.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	data := struct {
		*TestConfig
		Template *TestTemplate
	}{
		TestConfig: config,
		Template:   tmpl,
	}
	
	if err := t.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	
	fmt.Printf("✓ Generated: %s\n", filepath)
	return nil
}

// loadBuiltinTemplates loads built-in test templates
func (w *TestWizard) loadBuiltinTemplates() {
	// Lambda integration test template
	w.templates["lambda_integration"] = &TestTemplate{
		Name:        "lambda_integration",
		Description: "Lambda function integration tests",
		Service:     "lambda",
		Type:        "integration",
		Template: `package {{.PackageName}}_test

import (
	"testing"
	
	"vasdeference"
	"vasdeference/lambda"
)

func Test{{.FunctionName}}Integration(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()
	
	functionName := vdf.PrefixResourceName("{{.FunctionName}}")
	
	// Test function invocation
	payload := ` + "`{\"test\": \"data\"}`" + `
	result := lambda.InvokeFunction(t, functionName, payload)
	
	// Assertions
	lambda.AssertFunctionInvocationSuccess(t, result)
	lambda.AssertResponseContains(t, result, "expected_value")
}

func Test{{.FunctionName}}ErrorHandling(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()
	
	functionName := vdf.PrefixResourceName("{{.FunctionName}}")
	
	// Test error scenarios
	invalidPayload := ` + "`{\"invalid\": \"data\"}`" + `
	result := lambda.InvokeFunctionE(t, functionName, invalidPayload)
	
	// Assert error handling
	lambda.AssertFunctionError(t, result)
}

{{if .IncludeSetup}}
func TestMain(m *testing.M) {
	// Setup test environment
	setupTestEnvironment()
	
	// Run tests
	exitCode := m.Run()
	
	// Cleanup
	cleanupTestEnvironment()
	
	os.Exit(exitCode)
}

func setupTestEnvironment() {
	// Initialize test resources
}

func cleanupTestEnvironment() {
	// Cleanup test resources
}
{{end}}
`,
	}
	
	// DynamoDB integration test template
	w.templates["dynamodb_integration"] = &TestTemplate{
		Name:        "dynamodb_integration", 
		Description: "DynamoDB table integration tests",
		Service:     "dynamodb",
		Type:        "integration",
		Template: `package {{.PackageName}}_test

import (
	"testing"
	
	"vasdeference"
	"vasdeference/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func Test{{.ResourceName}}CRUD(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()
	
	tableName := vdf.PrefixResourceName("{{.ResourceName}}")
	
	// Test item creation
	item := map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: "test-id"},
		"name": &types.AttributeValueMemberS{Value: "Test Item"},
	}
	
	dynamodb.PutItem(t, tableName, item)
	
	// Test item retrieval
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "test-id"},
	}
	
	retrievedItem := dynamodb.GetItem(t, tableName, key)
	dynamodb.AssertItemExists(t, tableName, key)
	
	// Test item update
	updateExpression := "SET #name = :name"
	expressionAttributeNames := map[string]string{"#name": "name"}
	expressionAttributeValues := map[string]types.AttributeValue{
		":name": &types.AttributeValueMemberS{Value: "Updated Item"},
	}
	
	dynamodb.UpdateItem(t, tableName, key, updateExpression, expressionAttributeNames, expressionAttributeValues)
	
	// Test item deletion
	dynamodb.DeleteItem(t, tableName, key)
	dynamodb.AssertItemNotExists(t, tableName, key)
}

{{range $op := split .CustomVars.Operations ","}}
func Test{{$.ResourceName}}{{$op}}(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()
	
	tableName := vdf.PrefixResourceName("{{$.ResourceName}}")
	
	// Test {{$op}} operation
	// TODO: Implement {{$op}} test logic
}
{{end}}
`,
	}
	
	// EventBridge integration test template  
	w.templates["eventbridge_integration"] = &TestTemplate{
		Name:        "eventbridge_integration",
		Description: "EventBridge event bus integration tests", 
		Service:     "eventbridge",
		Type:        "integration",
		Template: `package {{.PackageName}}_test

import (
	"testing"
	
	"vasdeference" 
	"vasdeference/eventbridge"
)

func Test{{.ResourceName}}EventPublishing(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()
	
	eventBusName := "{{.ResourceName}}"
	
	// Test event publishing
	event := eventbridge.NewEvent().
		WithSource("test.application").
		WithDetailType("Test Event").
		WithDetail(map[string]interface{}{
			"test": "data",
		})
	
	eventbridge.PublishEvent(t, eventBusName, event)
	
	// Verify event was published
	eventbridge.AssertEventPublished(t, eventBusName, "test.application")
}

func Test{{.ResourceName}}RuleProcessing(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()
	
	eventBusName := "{{.ResourceName}}"
	ruleName := vdf.PrefixResourceName("test-rule")
	
	// Create test rule
	eventPattern := map[string]interface{}{
		"source": []string{"test.application"},
	}
	
	eventbridge.CreateRule(t, eventBusName, ruleName, eventPattern)
	vdf.RegisterCleanup(func() error {
		return eventbridge.DeleteRuleE(t, eventBusName, ruleName)
	})
	
	// Test rule matching
	event := eventbridge.NewEvent().
		WithSource("test.application").
		WithDetailType("Matching Event")
	
	eventbridge.PublishEvent(t, eventBusName, event)
	eventbridge.AssertRuleMatched(t, ruleName)
}
`,
	}
	
	// Generic test template
	w.templates["generic_integration"] = &TestTemplate{
		Name:        "generic_integration",
		Description: "Generic integration test template",
		Service:     "generic",
		Type:        "integration",
		Template: `package {{.PackageName}}_test

import (
	"testing"
	
	"vasdeference"
)

func Test{{.ResourceName}}Integration(t *testing.T) {
	vdf := vasdeference.New(t)
	defer vdf.Cleanup()
	
	// Test your {{.Service}} resource
	resourceName := vdf.PrefixResourceName("{{.ResourceName}}")
	
	// TODO: Implement your integration tests
	t.Logf("Testing resource: %s", resourceName)
}

{{if .IncludeSetup}}
func TestMain(m *testing.M) {
	// Setup test environment
	// TODO: Add setup logic
	
	// Run tests
	exitCode := m.Run()
	
	// Cleanup
	// TODO: Add cleanup logic
	
	os.Exit(exitCode)
}
{{end}}
`,
	}
	
	// Test helpers template
	w.templates["helpers"] = &TestTemplate{
		Name: "helpers",
		Template: `package {{.PackageName}}_test

import (
	"testing"
	
	"vasdeference"
)

// Helper functions for {{.Service}} tests

// setup{{.ResourceName}} initializes test resources
func setup{{.ResourceName}}(t *testing.T) *vasdeference.VasDeference {
	vdf := vasdeference.New(t)
	
	// TODO: Add setup logic
	
	return vdf
}

// cleanup{{.ResourceName}} cleans up test resources
func cleanup{{.ResourceName}}(vdf *vasdeference.VasDeference) {
	vdf.Cleanup()
}

// generate{{.ResourceName}}TestData creates test data
func generate{{.ResourceName}}TestData() interface{} {
	// TODO: Generate appropriate test data
	return nil
}

// assert{{.ResourceName}}State validates resource state
func assert{{.ResourceName}}State(t *testing.T, expected, actual interface{}) {
	// TODO: Add custom assertions
}
`,
	}
	
	// README template
	w.templates["readme"] = &TestTemplate{
		Name: "readme",
		Template: `# {{.Service}} Tests

Generated test suite for {{.Service}} {{.TestType}} tests.

## Overview

This test suite provides comprehensive testing for:
- {{.Service}} service operations
- {{.TestType}} test scenarios
- Resource: {{.ResourceName}}

## Prerequisites

- Go 1.24+ 
- AWS credentials configured
- SFX testing framework

## Running Tests

### All Tests
` + "```bash" + `
go test -v
` + "```" + `

### Specific Tests
` + "```bash" + `
go test -v -run TestSpecificFunction
` + "```" + `

### With Coverage
` + "```bash" + `
go test -v -cover
` + "```" + `

### Using SFX CLI
` + "```bash" + `
sfx test run --filter {{.PackageName}}
` + "```" + `

## Test Structure

- Main integration tests: ` + "`{{.PackageName}}_{{.TestType}}_test.go`" + `
{{if .IncludeExamples}}- Example tests: ` + "`{{.PackageName}}_examples_test.go`" + `{{end}}
{{if .IncludeSetup}}- Helper functions: ` + "`helpers_test.go`" + `{{end}}

## Configuration

- Region: {{.Region}}
- Namespace: {{.Namespace}}
- Resource: {{.ResourceName}}

## Customization

1. Update test data in helper functions
2. Add service-specific assertions
3. Configure additional test scenarios
4. Add performance benchmarks if needed

## Best Practices

- Use descriptive test names
- Include setup/teardown in test functions
- Use VasDeference for resource management
- Add proper error handling
- Validate all assertions

For more information, see the [SFX Documentation](https://github.com/your-org/sfx).
`,
	}
}