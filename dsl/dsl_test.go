package dsl

import (
	"testing"
	"time"
)

// TestDSLBasicFunctionality tests the core DSL functionality
func TestDSLBasicFunctionality(t *testing.T) {
	// Test basic AWS builder creation
	aws := NewAWS(t)
	if aws == nil {
		t.Fatal("Failed to create AWS builder")
	}

	// Test Lambda builder creation
	lambdaBuilder := aws.Lambda()
	if lambdaBuilder == nil {
		t.Fatal("Failed to create Lambda builder")
	}

	// Test function invocation
	result := aws.Lambda().
		Function("test-function").
		WithJSONPayload(map[string]interface{}{"test": "data"}).
		Invoke()

	if result == nil {
		t.Fatal("Lambda invocation returned nil result")
	}

	// Test assertions
	result.Should().Succeed()
}

// TestDSLDynamoDBOperations tests DynamoDB DSL operations
func TestDSLDynamoDBOperations(t *testing.T) {
	aws := NewAWS(t)

	// Test table creation
	createResult := aws.DynamoDB().
		CreateTable().
		Named("test-table").
		WithHashKey("id", "S").
		WithOnDemandBilling().
		Create()

	createResult.Should().Succeed()

	// Test put item operation
	putResult := aws.DynamoDB().
		Table("test-table").
		PutItem().
		WithAttribute("id", "test-id").
		WithAttribute("name", "Test User").
		Execute()

	putResult.Should().Succeed()

	// Test query operation
	queryResult := aws.DynamoDB().
		Table("test-table").
		Query().
		WhereKey("id").
		Equals("test-id").
		Execute()

	queryResult.Should().Succeed()
}

// TestDSLS3Operations tests S3 DSL operations
func TestDSLS3Operations(t *testing.T) {
	aws := NewAWS(t)

	// Test bucket creation
	bucketResult := aws.S3().
		CreateBucket().
		Named("test-bucket").
		InRegion("us-east-1").
		WithVersioning().
		Create()

	bucketResult.Should().Succeed()

	// Test object upload
	uploadResult := aws.S3().
		Bucket("test-bucket").
		UploadObject().
		WithKey("test/object.txt").
		WithBody("test content").
		WithContentType("text/plain").
		Execute()

	uploadResult.Should().Succeed()

	// Test object download
	downloadResult := aws.S3().
		Bucket("test-bucket").
		Object("test/object.txt").
		Download()

	downloadResult.Should().Succeed()
}

// TestDSLSESOperations tests SES DSL operations
func TestDSLSESOperations(t *testing.T) {
	aws := NewAWS(t)

	// Test identity verification
	verifyResult := aws.SES().
		Identity("test@example.com").
		Verify()

	verifyResult.Should().Succeed()

	// Test email sending
	emailResult := aws.SES().
		Email().
		From("test@example.com").
		To("recipient@example.com").
		WithSubject("Test Email").
		WithTextBody("This is a test email").
		Send()

	emailResult.Should().Succeed()
}

// TestDSLEventBridgeOperations tests EventBridge DSL operations
func TestDSLEventBridgeOperations(t *testing.T) {
	aws := NewAWS(t)

	// Test event bus creation
	busResult := aws.EventBridge().
		CreateEventBus().
		Named("test-bus").
		Create()

	busResult.Should().Succeed()

	// Test rule creation
	ruleResult := aws.EventBridge().
		Rule("test-rule").
		OnEventBus("test-bus").
		WithEventPattern(map[string]interface{}{
			"source": []string{"test.app"},
		}).
		Enabled().
		Create()

	ruleResult.Should().Succeed()

	// Test event publishing
	publishResult := aws.EventBridge().
		PublishEvent().
		ToEventBus("test-bus").
		WithSource("test.app").
		WithDetailType("Test Event").
		WithDetail(map[string]interface{}{
			"message": "Hello World",
		}).
		Publish()

	publishResult.Should().Succeed()
}

// TestDSLStepFunctionsOperations tests Step Functions DSL operations
func TestDSLStepFunctionsOperations(t *testing.T) {
	aws := NewAWS(t)

	definition := `{
		"Comment": "A Hello World example",
		"StartAt": "HelloWorld",
		"States": {
			"HelloWorld": {
				"Type": "Pass",
				"Result": "Hello World!",
				"End": true
			}
		}
	}`

	// Test state machine creation
	smResult := aws.StepFunctions().
		CreateStateMachine().
		Named("test-state-machine").
		WithDefinition(definition).
		WithRole("arn:aws:iam::123456789012:role/StepFunctionsRole").
		AsStandard().
		Create()

	smResult.Should().Succeed()

	// Test execution start
	execResult := aws.StepFunctions().
		StateMachine("test-state-machine").
		StartExecution().
		WithName("test-execution").
		WithJSONInput(map[string]interface{}{"input": "test"}).
		Execute()

	execResult.Should().Succeed()
}

// TestDSLAPIGatewayOperations tests API Gateway DSL operations
func TestDSLAPIGatewayOperations(t *testing.T) {
	aws := NewAWS(t)

	// Test REST API creation
	apiResult := aws.APIGateway().
		CreateRestAPI().
		Named("test-api").
		WithDescription("Test API").
		Create()

	apiResult.Should().Succeed()

	// Test resource creation
	resourceResult := aws.APIGateway().
		RestAPI("test-api-id").
		Resource("root").
		CreateChild().
		WithPathPart("users").
		Create()

	resourceResult.Should().Succeed()

	// Test method creation
	methodResult := aws.APIGateway().
		RestAPI("test-api-id").
		Resource("resource-id").
		Method("GET").
		WithAuthorization("NONE").
		Create()

	methodResult.Should().Succeed()
}

// TestDSLScenarios tests scenario building and execution
func TestDSLScenarios(t *testing.T) {
	aws := NewAWS(t)

	scenario := aws.Scenario().
		Step("Create bucket").
		Describe("Create S3 bucket for testing").
		Do(func() *ExecutionResult {
			return aws.S3().
				CreateBucket().
				Named("scenario-test-bucket").
				Create()
		}).
		Done().
		Step("Upload file").
		Describe("Upload test file to bucket").
		Do(func() *ExecutionResult {
			return aws.S3().
				Bucket("scenario-test-bucket").
				UploadObject().
				WithKey("test.txt").
				WithBody("test content").
				Execute()
		}).
		Done()

	result := scenario.Execute()
	result.Should().Succeed()
}

// TestDSLTestOrchestration tests complete test orchestration
func TestDSLTestOrchestration(t *testing.T) {
	aws := NewAWS(t)

	setupScenario := aws.Scenario().
		Step("Setup infrastructure").
		Do(func() *ExecutionResult {
			return &ExecutionResult{Success: true, ctx: aws.Context()}
		}).
		Done()

	testScenario := aws.Scenario().
		Step("Run main test").
		Do(func() *ExecutionResult {
			return aws.Lambda().
				Function("test-function").
				WithJSONPayload(map[string]interface{}{"test": true}).
				Invoke()
		}).
		Done()

	testResult := aws.Test().
		Named("Full Integration Test").
		WithSetup(func() *ExecutionResult {
			return setupScenario.Execute()
		}).
		AddScenario(testScenario).
		WithTeardown(func() *ExecutionResult {
			return &ExecutionResult{Success: true, ctx: aws.Context()}
		}).
		Run()

	testResult.Should().
		Succeed().
		And().
		CompleteWithin(30 * time.Second).
		And().
		HaveMetadata("test_name", "Full Integration Test")
}

// TestDSLAssertions tests various assertion capabilities
func TestDSLAssertions(t *testing.T) {
	aws := NewAWS(t)

	result := aws.Lambda().
		Function("test-function").
		WithTimeout(10 * time.Second).
		Invoke()

	// Test chained assertions
	result.Should().
		Succeed().
		And().
		CompleteWithin(5 * time.Second).
		And().
		HaveMetadata("function_name", "test-function")
}

// TestDSLAsyncOperations tests asynchronous operation handling
func TestDSLAsyncOperations(t *testing.T) {
	aws := NewAWS(t)

	// Test async Lambda invocation
	asyncResult := aws.Lambda().
		Function("long-running-function").
		WithJSONPayload(map[string]interface{}{"task": "process"}).
		Async().
		InvokeAsync()

	// Test waiting for completion
	completedResult := asyncResult.WaitWithTimeout(30 * time.Second)
	completedResult.Should().Succeed()
}