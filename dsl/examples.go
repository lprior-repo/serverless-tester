package dsl

import (
	"testing"
	"time"
)

// ExampleBasicLambdaInvocation demonstrates a simple Lambda function test
func ExampleBasicLambdaInvocation(t *testing.T) {
	// Create test context
	aws := NewAWS(t)
	
	// Simple Lambda invocation with assertions
	result := aws.Lambda().
		Function("my-function").
		WithJSONPayload(map[string]interface{}{"key": "value"}).
		Sync().
		WithLogs().
		Invoke()
	
	// Chain assertions fluently
	result.Should().
		Succeed().
		And().
		CompleteWithin(5 * time.Second).
		And().
		HaveMetadata("function_name", "my-function")
}

// ExampleDynamoDBOperations demonstrates DynamoDB testing with the DSL
func ExampleDynamoDBOperations(t *testing.T) {
	aws := NewAWS(t)
	
	// Create and configure table
	aws.DynamoDB().
		CreateTable().
		Named("users").
		WithHashKey("userId", "S").
		WithRangeKey("timestamp", "N").
		WithOnDemandBilling().
		Create().
		Should().
		Succeed()
	
	// Put item
	aws.DynamoDB().
		Table("users").
		PutItem().
		WithAttribute("userId", "user123").
		WithAttribute("timestamp", 1234567890).
		WithAttribute("name", "John Doe").
		Execute().
		Should().
		Succeed()
	
	// Query items
	aws.DynamoDB().
		Table("users").
		Query().
		WhereKey("userId").
		Equals("user123").
		Limit(10).
		Execute().
		Should().
		Succeed().
		And().
		ContainValue("user123")
}

// ExampleS3BucketManagement demonstrates S3 operations using the DSL
func ExampleS3BucketManagement(t *testing.T) {
	aws := NewAWS(t)
	
	// Create bucket with configuration
	aws.S3().
		CreateBucket().
		Named("test-bucket-12345").
		InRegion("us-west-2").
		WithVersioning().
		WithEncryption().
		Create().
		Should().
		Succeed()
	
	// Upload object
	aws.S3().
		Bucket("test-bucket-12345").
		UploadObject().
		WithKey("test/file.json").
		WithBody(`{"message": "Hello World"}`).
		WithContentType("application/json").
		Execute().
		Should().
		Succeed().
		And().
		HaveMetadata("content_type", "application/json")
	
	// Download and verify
	aws.S3().
		Bucket("test-bucket-12345").
		Object("test/file.json").
		Download().
		Should().
		Succeed()
}

// ExampleSESEmailWorkflow demonstrates email sending with SES
func ExampleSESEmailWorkflow(t *testing.T) {
	aws := NewAWS(t)
	
	// Verify sender identity
	aws.SES().
		Identity("sender@example.com").
		Verify().
		Should().
		Succeed()
	
	// Create email template
	aws.SES().
		Template("welcome-email").
		WithSubject("Welcome {{name}}!").
		WithHTMLPart("<h1>Hello {{name}}</h1>").
		WithTextPart("Hello {{name}}").
		Create().
		Should().
		Succeed()
	
	// Send templated email
	aws.SES().
		Email().
		From("sender@example.com").
		To("recipient@example.com").
		WithTemplate("welcome-email", map[string]interface{}{
			"name": "John",
		}).
		Send().
		Should().
		Succeed().
		And().
		HaveMetadata("to_count", 1)
}

// ExampleEventBridgeWorkflow demonstrates EventBridge event handling
func ExampleEventBridgeWorkflow(t *testing.T) {
	aws := NewAWS(t)
	
	// Create custom event bus
	aws.EventBridge().
		CreateEventBus().
		Named("test-bus").
		Create().
		Should().
		Succeed()
	
	// Create rule with event pattern
	aws.EventBridge().
		Rule("user-events").
		OnEventBus("test-bus").
		WithEventPattern(map[string]interface{}{
			"source":      []string{"myapp.users"},
			"detail-type": []string{"User Created"},
		}).
		Enabled().
		AddTarget().
		WithId("lambda-target").
		WithArn("arn:aws:lambda:us-east-1:123456789012:function:process-user").
		Done().
		Create().
		Should().
		Succeed()
	
	// Publish event
	aws.EventBridge().
		PublishEvent().
		ToEventBus("test-bus").
		WithSource("myapp.users").
		WithDetailType("User Created").
		WithDetail(map[string]interface{}{
			"userId": "user123",
			"name":   "John Doe",
		}).
		Publish().
		Should().
		Succeed().
		And().
		HaveMetadata("event_bus", "test-bus")
}

// ExampleStepFunctionsWorkflow demonstrates Step Functions state machine testing
func ExampleStepFunctionsWorkflow(t *testing.T) {
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
	
	// Create state machine
	aws.StepFunctions().
		CreateStateMachine().
		Named("HelloWorldStateMachine").
		WithDefinition(definition).
		WithRole("arn:aws:iam::123456789012:role/StepFunctionsRole").
		AsStandard().
		Create().
		Should().
		Succeed()
	
	// Start execution
	execution := aws.StepFunctions().
		StateMachine("HelloWorldStateMachine").
		StartExecution().
		WithName("test-execution").
		WithJSONInput(map[string]interface{}{"input": "test"}).
		Execute()
	
	execution.Should().Succeed()
	
	// Wait for completion and verify result
	aws.StepFunctions().
		Execution(execution.Data.(map[string]interface{})["executionArn"].(string)).
		WaitUntilCompleted().
		Should().
		Succeed().
		And().
		HaveMetadata("status", "SUCCEEDED")
}

// ExampleAPIGatewaySetup demonstrates API Gateway configuration
func ExampleAPIGatewaySetup(t *testing.T) {
	aws := NewAWS(t)
	
	// Create REST API
	api := aws.APIGateway().
		CreateRestAPI().
		Named("test-api").
		WithDescription("Test API for demonstration").
		Create()
	
	api.Should().Succeed()
	
	apiId := api.Data.(map[string]interface{})["id"].(string)
	
	// Create resource
	resource := aws.APIGateway().
		RestAPI(apiId).
		Resource("root").
		CreateChild().
		WithPathPart("users").
		Create()
	
	resource.Should().Succeed()
	
	resourceId := resource.Data.(map[string]interface{})["id"].(string)
	
	// Create method with Lambda integration
	aws.APIGateway().
		RestAPI(apiId).
		Resource(resourceId).
		Method("GET").
		WithAuthorization("NONE").
		Create().
		Should().
		Succeed()
	
	// Test invoke the method
	aws.APIGateway().
		RestAPI(apiId).
		TestInvoke().
		OnResource(resourceId).
		WithMethod("GET").
		WithPath("/users").
		Execute().
		Should().
		Succeed().
		And().
		HaveMetadata("status", 200)
}

// ExampleComplexScenario demonstrates a high-level testing scenario
func ExampleComplexScenario(t *testing.T) {
	aws := NewAWS(t)
	
	scenario := aws.Scenario().
		Step("Create S3 bucket").
		Describe("Set up storage for user uploads").
		Do(func() *ExecutionResult {
			return aws.S3().
				CreateBucket().
				Named("user-uploads-bucket").
				WithVersioning().
				Create()
		}).
		Done().
		Step("Create DynamoDB table").
		Describe("Set up user metadata storage").
		Do(func() *ExecutionResult {
			return aws.DynamoDB().
				CreateTable().
				Named("user-metadata").
				WithHashKey("userId", "S").
				WithOnDemandBilling().
				Create()
		}).
		Done().
		Step("Deploy Lambda function").
		Describe("Deploy user processing function").
		Do(func() *ExecutionResult {
			return aws.Lambda().
				CreateFunction().
				Named("process-user-upload").
				WithRuntime("nodejs18.x").
				WithHandler("index.handler").
				WithRole("arn:aws:iam::123456789012:role/LambdaRole").
				Create()
		}).
		Done().
		Step("Test end-to-end workflow").
		Describe("Simulate user upload and processing").
		Do(func() *ExecutionResult {
			// Upload file
			uploadResult := aws.S3().
				Bucket("user-uploads-bucket").
				UploadObject().
				WithKey("user123/photo.jpg").
				WithBody("fake-image-data").
				Execute()
			
			if !uploadResult.Success {
				return uploadResult
			}
			
			// Process with Lambda
			return aws.Lambda().
				Function("process-user-upload").
				WithJSONPayload(map[string]interface{}{
					"bucket": "user-uploads-bucket",
					"key":    "user123/photo.jpg",
				}).
				Invoke()
		}).
		Done()
	
	// Execute the complete scenario
	scenario.Execute().
		Should().
		Succeed().
		And().
		CompleteWithin(30 * time.Second)
}

// ExampleFullTestOrchestration demonstrates complete test orchestration
func ExampleFullTestOrchestration(t *testing.T) {
	aws := NewAWS(t)
	
	// Create scenarios
	setupScenario := aws.Scenario().
		Step("Create infrastructure").
		Do(func() *ExecutionResult {
			// Setup logic here
			return &ExecutionResult{Success: true, ctx: aws.Context()}
		}).
		Done()
	
	testScenario := aws.Scenario().
		Step("Run tests").
		Do(func() *ExecutionResult {
			// Test logic here
			return &ExecutionResult{Success: true, ctx: aws.Context()}
		}).
		Done()
	
	// Orchestrate complete test
	aws.Test().
		Named("Complete Integration Test").
		WithSetup(func() *ExecutionResult {
			return setupScenario.Execute()
		}).
		AddScenario(testScenario).
		WithTeardown(func() *ExecutionResult {
			// Cleanup logic
			return &ExecutionResult{Success: true, ctx: aws.Context()}
		}).
		Run().
		Should().
		Succeed().
		And().
		HaveMetadata("test_name", "Complete Integration Test")
}

// ExampleAsyncOperations demonstrates asynchronous operations
func ExampleAsyncOperations(t *testing.T) {
	aws := NewAWS(t)
	
	// Start async Lambda invocation
	asyncResult := aws.Lambda().
		Function("long-running-task").
		WithJSONPayload(map[string]interface{}{"task": "process-data"}).
		Async().
		InvokeAsync()
	
	// Do other work while Lambda runs
	aws.S3().
		Bucket("temp-bucket").
		Object("status.json").
		Download().
		Should().
		Succeed()
	
	// Wait for async operation to complete
	asyncResult.
		WaitWithTimeout(10 * time.Second).
		Should().
		Succeed()
}