package testutils

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

// ========== TDD: RED PHASE - Contract Testing Tests ==========

// RED: Test contract creation and validation
func TestContractTesting_ShouldCreateAndValidateContracts(t *testing.T) {
	contractTest := NewContractTest()
	
	// Create a contract for Lambda -> DynamoDB interaction
	contract := contractTest.CreateContract("LambdaToDynamoDB", "1.0.0")
	
	// Define provider (Lambda function)
	provider := contract.DefineProvider("UserService", "lambda")
	provider.AddEndpoint("POST", "/users", map[string]interface{}{
		"name":  "string",
		"email": "string",
	}, map[string]interface{}{
		"id":     "string",
		"status": "created",
	})
	
	// Define consumer (DynamoDB interaction)
	consumer := contract.DefineConsumer("UserDatabase", "dynamodb")
	consumer.ExpectOperation("PutItem", map[string]interface{}{
		"TableName": "Users",
		"Item": map[string]interface{}{
			"id":    "string",
			"name":  "string",
			"email": "string",
		},
	})
	
	assert.Equal(t, "LambdaToDynamoDB", contract.GetName(), "Contract name should match")
	assert.Equal(t, "1.0.0", contract.GetVersion(), "Contract version should match")
	assert.NotNil(t, contract.GetProvider(), "Provider should be defined")
	assert.NotNil(t, contract.GetConsumer(), "Consumer should be defined")
}

// RED: Test contract verification between services
func TestContractTesting_ShouldVerifyServiceContracts(t *testing.T) {
	contractTest := NewContractTest()
	contract := contractTest.CreateContract("APIGatewayToLambda", "1.0.0")
	
	// Setup API Gateway -> Lambda contract
	provider := contract.DefineProvider("APIGateway", "apigateway")
	provider.AddEndpoint("GET", "/health", nil, map[string]interface{}{
		"status": "string",
		"timestamp": "number",
	})
	
	consumer := contract.DefineConsumer("HealthCheckLambda", "lambda")
	consumer.ExpectResponse(200, map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
	})
	
	// Verify contract
	verification := contractTest.VerifyContract(contract)
	
	assert.True(t, verification.IsValid, "Contract should be valid")
	assert.Empty(t, verification.Violations, "Should have no violations")
	assert.NotEmpty(t, verification.TestsRun, "Should have run tests")
}

// RED: Test contract violation detection
func TestContractTesting_ShouldDetectContractViolations(t *testing.T) {
	contractTest := NewContractTest()
	contract := contractTest.CreateContract("InvalidContract", "1.0.0")
	
	// Create mismatched contract
	provider := contract.DefineProvider("Service1", "lambda")
	provider.AddEndpoint("POST", "/data", map[string]interface{}{
		"name": "string",
	}, map[string]interface{}{
		"id": "string",
		"status": "created",
	})
	
	consumer := contract.DefineConsumer("Service2", "lambda")
	// Consumer expects different response structure
	consumer.ExpectResponse(200, map[string]interface{}{
		"user_id": "string",  // Different field name
		"result":  "success", // Different field
	})
	
	verification := contractTest.VerifyContract(contract)
	
	assert.False(t, verification.IsValid, "Contract should be invalid due to mismatch")
	assert.NotEmpty(t, verification.Violations, "Should have violations")
}

// RED: Test pact-style consumer-driven contracts
func TestContractTesting_ShouldSupportConsumerDrivenContracts(t *testing.T) {
	contractTest := NewContractTest()
	
	// Consumer defines what it expects from provider
	pact := contractTest.CreatePact("UserServiceConsumer", "UserServiceProvider")
	
	interaction := pact.AddInteraction("Get user by ID")
	interaction.Given("user exists with ID 123")
	interaction.WhenReceiving("GET request to /users/123")
	interaction.WithHeaders(map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	})
	interaction.WillRespondWith(200, map[string]interface{}{
		"id":    "123",
		"name":  "John Doe",
		"email": "john@example.com",
	})
	
	// Verify pact
	pactVerification := contractTest.VerifyPact(pact)
	
	assert.True(t, pactVerification.IsValid, "Pact should be valid")
	assert.Equal(t, "UserServiceConsumer", pact.GetConsumer(), "Consumer should match")
	assert.Equal(t, "UserServiceProvider", pact.GetProvider(), "Provider should match")
}

// RED: Test contract evolution and backward compatibility
func TestContractTesting_ShouldHandleContractEvolution(t *testing.T) {
	contractTest := NewContractTest()
	
	// Original contract v1.0.0
	originalContract := contractTest.CreateContract("UserAPI", "1.0.0")
	provider1 := originalContract.DefineProvider("UserService", "lambda")
	provider1.AddEndpoint("GET", "/user/{id}", nil, map[string]interface{}{
		"id":   "string",
		"name": "string",
	})
	
	// Evolved contract v1.1.0 with additional field
	evolvedContract := contractTest.CreateContract("UserAPI", "1.1.0")
	provider2 := evolvedContract.DefineProvider("UserService", "lambda")
	provider2.AddEndpoint("GET", "/user/{id}", nil, map[string]interface{}{
		"id":    "string",
		"name":  "string",
		"email": "string", // New field
	})
	
	// Check backward compatibility
	compatibility := contractTest.CheckBackwardCompatibility(originalContract, evolvedContract)
	
	assert.True(t, compatibility.IsCompatible, "Should be backward compatible")
	found := false
	for _, change := range compatibility.Changes {
		if strings.Contains(change, "Added field: email") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should detect new field")
	assert.Empty(t, compatibility.BreakingChanges, "Should have no breaking changes")
}

// RED: Test integration with AWS services contract validation
func TestContractTesting_ShouldValidateAWSServiceContracts(t *testing.T) {
	contractTest := NewContractTest()
	
	// Contract for Lambda -> EventBridge interaction
	contract := contractTest.CreateContract("LambdaToEventBridge", "1.0.0")
	
	provider := contract.DefineProvider("OrderProcessingLambda", "lambda")
	provider.AddEventPublication("order.created", map[string]interface{}{
		"source":      "order.service",
		"detail-type": "Order Created",
		"detail": map[string]interface{}{
			"orderId":     "string",
			"customerId":  "string",
			"amount":      "number",
			"timestamp":   "string",
			"status":      "string",
		},
	})
	
	consumer := contract.DefineConsumer("EventBridge", "eventbridge")
	
	consumer.ExpectEvent("order.created", map[string]interface{}{
		"source":      "order.service",
		"detail-type": "Order Created",
		"detail": map[string]interface{}{
			"orderId":    "string",
			"customerId": "string",
			"amount":     "number",
		},
	})
	
	verification := contractTest.VerifyAWSContract(contract)
	
	assert.True(t, verification.IsValid, "AWS contract should be valid")
	assert.Empty(t, verification.AWSSpecificViolations, "Should have no AWS-specific violations")
}