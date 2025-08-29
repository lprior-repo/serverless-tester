package testutils

import (
	"fmt"
	"reflect"
	"strings"
)

// ContractTest provides contract testing capabilities for microservices
type ContractTest struct {
	contracts map[string]*Contract
	pacts     map[string]*Pact
}

// Contract represents a service contract
type Contract struct {
	name     string
	version  string
	provider *Provider
	consumer *Consumer
}

// Provider represents the service providing an API
type Provider struct {
	name      string
	serviceType string
	endpoints map[string]*Endpoint
	events    map[string]*EventSchema
}

// Consumer represents the service consuming an API
type Consumer struct {
	name        string
	serviceType string
	expectations map[string]*Expectation
	eventExpectations map[string]*EventExpectation
}

// Endpoint represents an API endpoint
type Endpoint struct {
	method   string
	path     string
	request  map[string]interface{}
	response map[string]interface{}
}

// EventSchema represents an event structure
type EventSchema struct {
	name   string
	schema map[string]interface{}
}

// Expectation represents what a consumer expects
type Expectation struct {
	operation string
	input     map[string]interface{}
	output    map[string]interface{}
}

// EventExpectation represents expected event structure
type EventExpectation struct {
	eventType string
	schema    map[string]interface{}
}

// ContractVerification represents the result of contract verification
type ContractVerification struct {
	IsValid               bool
	Violations            []string
	TestsRun              int
	AWSSpecificViolations []string
}

// BackwardCompatibility represents compatibility check results
type BackwardCompatibility struct {
	IsCompatible     bool
	Changes          []string
	BreakingChanges  []string
}

// Pact represents a consumer-driven contract (Pact-style)
type Pact struct {
	consumer     string
	provider     string
	interactions []*Interaction
}

// Interaction represents a single interaction in a pact
type Interaction struct {
	description string
	given       string
	request     *PactRequest
	response    *PactResponse
}

// PactRequest represents a request in a pact
type PactRequest struct {
	method  string
	path    string
	headers map[string]string
	body    map[string]interface{}
}

// PactResponse represents a response in a pact
type PactResponse struct {
	status  int
	headers map[string]string
	body    map[string]interface{}
}

// PactVerification represents pact verification results
type PactVerification struct {
	IsValid       bool
	Violations    []string
	TestsRun      int
}

// NewContractTest creates a new contract testing instance
func NewContractTest() *ContractTest {
	return &ContractTest{
		contracts: make(map[string]*Contract),
		pacts:     make(map[string]*Pact),
	}
}

// CreateContract creates a new contract
func (ct *ContractTest) CreateContract(name, version string) *Contract {
	contract := &Contract{
		name:    name,
		version: version,
	}
	ct.contracts[name] = contract
	return contract
}

// DefineProvider defines the provider side of a contract
func (c *Contract) DefineProvider(name, serviceType string) *Provider {
	provider := &Provider{
		name:        name,
		serviceType: serviceType,
		endpoints:   make(map[string]*Endpoint),
		events:      make(map[string]*EventSchema),
	}
	c.provider = provider
	return provider
}

// DefineConsumer defines the consumer side of a contract
func (c *Contract) DefineConsumer(name, serviceType string) *Consumer {
	consumer := &Consumer{
		name:              name,
		serviceType:       serviceType,
		expectations:      make(map[string]*Expectation),
		eventExpectations: make(map[string]*EventExpectation),
	}
	c.consumer = consumer
	return consumer
}

// AddEndpoint adds an endpoint to the provider
func (p *Provider) AddEndpoint(method, path string, request, response map[string]interface{}) {
	endpoint := &Endpoint{
		method:   method,
		path:     path,
		request:  request,
		response: response,
	}
	key := fmt.Sprintf("%s:%s", method, path)
	p.endpoints[key] = endpoint
}

// AddEventPublication adds an event publication to the provider
func (p *Provider) AddEventPublication(eventName string, schema map[string]interface{}) {
	event := &EventSchema{
		name:   eventName,
		schema: schema,
	}
	p.events[eventName] = event
}

// ExpectOperation adds an operation expectation to the consumer
func (c *Consumer) ExpectOperation(operation string, input map[string]interface{}) {
	expectation := &Expectation{
		operation: operation,
		input:     input,
	}
	c.expectations[operation] = expectation
}

// ExpectResponse adds a response expectation to the consumer
func (c *Consumer) ExpectResponse(status int, response map[string]interface{}) {
	// Store as a special expectation
	expectation := &Expectation{
		operation: fmt.Sprintf("HTTP_%d", status),
		output:    response,
	}
	c.expectations[fmt.Sprintf("response_%d", status)] = expectation
}

// ExpectEvent adds an event expectation to the consumer
func (c *Consumer) ExpectEvent(eventType string, schema map[string]interface{}) {
	eventExp := &EventExpectation{
		eventType: eventType,
		schema:    schema,
	}
	c.eventExpectations[eventType] = eventExp
}

// Contract getter methods
func (c *Contract) GetName() string        { return c.name }
func (c *Contract) GetVersion() string     { return c.version }
func (c *Contract) GetProvider() *Provider { return c.provider }
func (c *Contract) GetConsumer() *Consumer { return c.consumer }

// VerifyContract verifies that a contract is internally consistent
func (ct *ContractTest) VerifyContract(contract *Contract) ContractVerification {
	violations := make([]string, 0)
	testsRun := 0
	
	if contract.provider == nil {
		violations = append(violations, "Contract has no provider defined")
		return ContractVerification{
			IsValid:    false,
			Violations: violations,
			TestsRun:   testsRun,
		}
	}
	
	if contract.consumer == nil {
		violations = append(violations, "Contract has no consumer defined")
		return ContractVerification{
			IsValid:    false,
			Violations: violations,
			TestsRun:   testsRun,
		}
	}
	
	// Check endpoint compatibility
	for endpointKey, endpoint := range contract.provider.endpoints {
		testsRun++
		
		// Look for matching consumer expectations
		found := false
		for _, expectation := range contract.consumer.expectations {
			if ct.isCompatibleEndpoint(endpoint, expectation) {
				found = true
				break
			}
		}
		
		if !found {
			violations = append(violations, fmt.Sprintf("No consumer expectation for endpoint %s", endpointKey))
		}
	}
	
	// Check event compatibility
	for eventName, event := range contract.provider.events {
		testsRun++
		
		if eventExp, exists := contract.consumer.eventExpectations[eventName]; exists {
			if !ct.isCompatibleEventSchema(event.schema, eventExp.schema) {
				violations = append(violations, fmt.Sprintf("Event schema mismatch for %s", eventName))
			}
		} else {
			violations = append(violations, fmt.Sprintf("No consumer expectation for event %s", eventName))
		}
	}
	
	return ContractVerification{
		IsValid:    len(violations) == 0,
		Violations: violations,
		TestsRun:   testsRun,
	}
}

// CreatePact creates a new pact (consumer-driven contract)
func (ct *ContractTest) CreatePact(consumer, provider string) *Pact {
	pact := &Pact{
		consumer:     consumer,
		provider:     provider,
		interactions: make([]*Interaction, 0),
	}
	key := fmt.Sprintf("%s-%s", consumer, provider)
	ct.pacts[key] = pact
	return pact
}

// AddInteraction adds an interaction to a pact
func (p *Pact) AddInteraction(description string) *Interaction {
	interaction := &Interaction{
		description: description,
		request:     &PactRequest{headers: make(map[string]string)},
		response:    &PactResponse{headers: make(map[string]string)},
	}
	p.interactions = append(p.interactions, interaction)
	return interaction
}

// Pact interaction builder methods
func (i *Interaction) Given(state string) *Interaction {
	i.given = state
	return i
}

func (i *Interaction) WhenReceiving(method string) *Interaction {
	parts := strings.SplitN(method, " ", 3)
	if len(parts) >= 2 {
		i.request.method = parts[0]
		if len(parts) >= 4 && parts[2] == "to" {
			i.request.path = parts[3]
		}
	}
	return i
}

func (i *Interaction) WithHeaders(headers map[string]string) *Interaction {
	for k, v := range headers {
		i.request.headers[k] = v
	}
	return i
}

func (i *Interaction) WillRespondWith(status int, body map[string]interface{}) *Interaction {
	i.response.status = status
	i.response.body = body
	return i
}

// Pact getter methods
func (p *Pact) GetConsumer() string { return p.consumer }
func (p *Pact) GetProvider() string { return p.provider }

// VerifyPact verifies a pact contract
func (ct *ContractTest) VerifyPact(pact *Pact) PactVerification {
	violations := make([]string, 0)
	testsRun := len(pact.interactions)
	
	for _, interaction := range pact.interactions {
		// Basic validation of interaction structure
		if interaction.description == "" {
			violations = append(violations, "Interaction missing description")
		}
		
		if interaction.request == nil {
			violations = append(violations, "Interaction missing request")
		} else {
			if interaction.request.method == "" {
				violations = append(violations, "Request missing method")
			}
		}
		
		if interaction.response == nil {
			violations = append(violations, "Interaction missing response")
		} else {
			if interaction.response.status == 0 {
				violations = append(violations, "Response missing status")
			}
		}
	}
	
	return PactVerification{
		IsValid:    len(violations) == 0,
		Violations: violations,
		TestsRun:   testsRun,
	}
}

// CheckBackwardCompatibility checks if a contract evolution is backward compatible
func (ct *ContractTest) CheckBackwardCompatibility(oldContract, newContract *Contract) BackwardCompatibility {
	changes := make([]string, 0)
	breakingChanges := make([]string, 0)
	
	// Compare provider endpoints
	if oldContract.provider != nil && newContract.provider != nil {
		for oldEndpointKey, oldEndpoint := range oldContract.provider.endpoints {
			if newEndpoint, exists := newContract.provider.endpoints[oldEndpointKey]; exists {
				// Check for response schema changes
				oldFields := ct.getFieldNames(oldEndpoint.response)
				newFields := ct.getFieldNames(newEndpoint.response)
				
				// Check for removed fields (breaking change)
				for _, oldField := range oldFields {
					if !ct.containsString(newFields, oldField) {
						breakingChanges = append(breakingChanges, fmt.Sprintf("Removed field: %s from %s", oldField, oldEndpointKey))
					}
				}
				
				// Check for added fields (compatible change)
				for _, newField := range newFields {
					if !ct.containsString(oldFields, newField) {
						changes = append(changes, fmt.Sprintf("Added field: %s to %s", newField, oldEndpointKey))
					}
				}
			} else {
				breakingChanges = append(breakingChanges, fmt.Sprintf("Removed endpoint: %s", oldEndpointKey))
			}
		}
		
		// Check for new endpoints (compatible change)
		for newEndpointKey := range newContract.provider.endpoints {
			if _, exists := oldContract.provider.endpoints[newEndpointKey]; !exists {
				changes = append(changes, fmt.Sprintf("Added endpoint: %s", newEndpointKey))
			}
		}
	}
	
	return BackwardCompatibility{
		IsCompatible:    len(breakingChanges) == 0,
		Changes:         changes,
		BreakingChanges: breakingChanges,
	}
}

// VerifyAWSContract verifies contracts with AWS service-specific validation
func (ct *ContractTest) VerifyAWSContract(contract *Contract) ContractVerification {
	baseVerification := ct.VerifyContract(contract)
	awsViolations := make([]string, 0)
	
	// AWS-specific validations
	if contract.provider != nil {
		// Check for AWS service naming conventions
		if contract.provider.serviceType == "lambda" {
			for _, endpoint := range contract.provider.endpoints {
				if endpoint.response != nil {
					// Lambda responses should have consistent structure
					if _, hasStatusCode := endpoint.response["statusCode"]; !hasStatusCode {
						awsViolations = append(awsViolations, "Lambda response missing statusCode")
					}
				}
			}
		}
		
		// Check EventBridge event structure
		if contract.provider.serviceType == "eventbridge" || 
		   (contract.consumer != nil && contract.consumer.serviceType == "eventbridge") {
			for _, event := range contract.provider.events {
				if event.schema != nil {
					requiredFields := []string{"source", "detail-type", "detail"}
					for _, field := range requiredFields {
						if _, hasField := event.schema[field]; !hasField {
							awsViolations = append(awsViolations, fmt.Sprintf("EventBridge event missing required field: %s", field))
						}
					}
				}
			}
		}
	}
	
	return ContractVerification{
		IsValid:               baseVerification.IsValid && len(awsViolations) == 0,
		Violations:            baseVerification.Violations,
		TestsRun:              baseVerification.TestsRun,
		AWSSpecificViolations: awsViolations,
	}
}

// Helper functions

func (ct *ContractTest) isCompatibleEndpoint(endpoint *Endpoint, expectation *Expectation) bool {
	if endpoint == nil || expectation == nil {
		return false
	}
	
	// Check if expectation is for HTTP response and endpoint provides compatible response
	if strings.HasPrefix(expectation.operation, "HTTP_") && endpoint.response != nil {
		// For demonstration, check if response structures are compatible
		return ct.areStructuresCompatible(endpoint.response, expectation.output)
	}
	
	return false
}

func (ct *ContractTest) areStructuresCompatible(provided, expected map[string]interface{}) bool {
	if provided == nil || expected == nil {
		return false
	}
	
	// Check if expected fields are present in provided structure
	for expectedKey, expectedValue := range expected {
		providedValue, exists := provided[expectedKey]
		if !exists {
			return false
		}
		
		// Simple type compatibility check
		expectedType := fmt.Sprintf("%T", expectedValue)
		providedType := fmt.Sprintf("%T", providedValue)
		
		// Allow some flexibility in type matching
		if !ct.areTypesCompatible(expectedType, providedType) {
			return false
		}
	}
	
	return true
}

func (ct *ContractTest) areTypesCompatible(expected, provided string) bool {
	if expected == provided {
		return true
	}
	
	// Allow some basic type compatibility
	numericTypes := []string{"int", "int64", "float64", "number"}
	if ct.containsString(numericTypes, expected) && ct.containsString(numericTypes, provided) {
		return true
	}
	
	return false
}

func (ct *ContractTest) isCompatibleEventSchema(providerSchema, consumerSchema map[string]interface{}) bool {
	// Check if consumer schema is a subset of provider schema
	return ct.isSchemaSubset(consumerSchema, providerSchema)
}

func (ct *ContractTest) isSchemaSubset(subset, superset map[string]interface{}) bool {
	for key, subValue := range subset {
		superValue, exists := superset[key]
		if !exists {
			return false
		}
		
		// Simple type checking - in real implementation would be more comprehensive
		if reflect.TypeOf(subValue) != reflect.TypeOf(superValue) {
			return false
		}
	}
	return true
}

func (ct *ContractTest) getFieldNames(schema map[string]interface{}) []string {
	if schema == nil {
		return []string{}
	}
	
	fields := make([]string, 0, len(schema))
	for field := range schema {
		fields = append(fields, field)
	}
	return fields
}

func (ct *ContractTest) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}