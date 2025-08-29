package dsl

import (
	"fmt"
	"time"
)

// APIGatewayBuilder provides a fluent interface for API Gateway operations
type APIGatewayBuilder struct {
	ctx *TestContext
}

// RestAPI creates a REST API builder for a specific API
func (agb *APIGatewayBuilder) RestAPI(apiId string) *RestAPIBuilder {
	return &RestAPIBuilder{
		ctx:   agb.ctx,
		apiId: apiId,
		config: &RestAPIConfig{},
	}
}

// CreateRestAPI starts building a new REST API
func (agb *APIGatewayBuilder) CreateRestAPI() *RestAPICreateBuilder {
	return &RestAPICreateBuilder{
		ctx:    agb.ctx,
		config: &RestAPICreateConfig{},
	}
}

// HttpAPI creates an HTTP API builder for a specific API
func (agb *APIGatewayBuilder) HttpAPI(apiId string) *HttpAPIBuilder {
	return &HttpAPIBuilder{
		ctx:   agb.ctx,
		apiId: apiId,
		config: &HttpAPIConfig{},
	}
}

// DomainName creates a domain name builder
func (agb *APIGatewayBuilder) DomainName(domainName string) *DomainNameBuilder {
	return &DomainNameBuilder{
		ctx:        agb.ctx,
		domainName: domainName,
		config:     &DomainNameConfig{},
	}
}

// RestAPIBuilder builds operations on an existing REST API
type RestAPIBuilder struct {
	ctx    *TestContext
	apiId  string
	config *RestAPIConfig
}

type RestAPIConfig struct {
	Name        string
	Description string
	Version     string
	Tags        map[string]string
}

// Resource creates a resource builder for the API
func (rab *RestAPIBuilder) Resource(resourceId string) *ResourceBuilder {
	return &ResourceBuilder{
		ctx:        rab.ctx,
		apiId:      rab.apiId,
		resourceId: resourceId,
		config:     &ResourceConfig{},
	}
}

// Stage creates a stage builder for the API
func (rab *RestAPIBuilder) Stage(stageName string) *StageBuilder {
	return &StageBuilder{
		ctx:       rab.ctx,
		apiId:     rab.apiId,
		stageName: stageName,
		config:    &StageConfig{},
	}
}

// Deployment creates a deployment builder for the API
func (rab *RestAPIBuilder) Deployment() *DeploymentBuilder {
	return &DeploymentBuilder{
		ctx:    rab.ctx,
		apiId:  rab.apiId,
		config: &DeploymentConfig{},
	}
}

// Delete deletes the REST API
func (rab *RestAPIBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("REST API %s deleted", rab.apiId),
		ctx:     rab.ctx,
	}
}

// GetInfo retrieves API information
func (rab *RestAPIBuilder) GetInfo() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"id":          rab.apiId,
			"name":        "sample-api",
			"description": "Sample REST API",
			"createdDate": time.Now().Add(-24 * time.Hour),
		},
		ctx: rab.ctx,
	}
}

// TestInvoke creates a test invoke builder
func (rab *RestAPIBuilder) TestInvoke() *TestInvokeBuilder {
	return &TestInvokeBuilder{
		ctx:   rab.ctx,
		apiId: rab.apiId,
		config: &TestInvokeConfig{},
	}
}

// RestAPICreateBuilder builds REST API creation operations
type RestAPICreateBuilder struct {
	ctx    *TestContext
	config *RestAPICreateConfig
}

type RestAPICreateConfig struct {
	Name            string
	Description     string
	Version         string
	CloneFrom       string
	BinaryMediaTypes []string
	MinimumCompressionSize *int32
	Tags            map[string]string
}

// Named sets the API name
func (racb *RestAPICreateBuilder) Named(name string) *RestAPICreateBuilder {
	racb.config.Name = name
	return racb
}

// WithDescription sets the API description
func (racb *RestAPICreateBuilder) WithDescription(description string) *RestAPICreateBuilder {
	racb.config.Description = description
	return racb
}

// WithVersion sets the API version
func (racb *RestAPICreateBuilder) WithVersion(version string) *RestAPICreateBuilder {
	racb.config.Version = version
	return racb
}

// WithBinaryMediaTypes sets binary media types
func (racb *RestAPICreateBuilder) WithBinaryMediaTypes(types ...string) *RestAPICreateBuilder {
	racb.config.BinaryMediaTypes = types
	return racb
}

// Create creates the REST API
func (racb *RestAPICreateBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"id":          "abcd123456",
			"name":        racb.config.Name,
			"description": racb.config.Description,
			"createdDate": time.Now(),
		},
		ctx: racb.ctx,
	}
}

// ResourceBuilder builds operations on API resources
type ResourceBuilder struct {
	ctx        *TestContext
	apiId      string
	resourceId string
	config     *ResourceConfig
}

type ResourceConfig struct {
	PathPart string
	ParentId string
}

// Method creates a method builder for the resource
func (rb *ResourceBuilder) Method(httpMethod string) *MethodBuilder {
	return &MethodBuilder{
		ctx:        rb.ctx,
		apiId:      rb.apiId,
		resourceId: rb.resourceId,
		httpMethod: httpMethod,
		config:     &MethodConfig{},
	}
}

// CreateChild creates a child resource
func (rb *ResourceBuilder) CreateChild() *ResourceCreateBuilder {
	return &ResourceCreateBuilder{
		ctx:      rb.ctx,
		apiId:    rb.apiId,
		parentId: rb.resourceId,
		config:   &ResourceCreateConfig{},
	}
}

// Delete deletes the resource
func (rb *ResourceBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Resource %s deleted from API %s", rb.resourceId, rb.apiId),
		ctx:     rb.ctx,
	}
}

// GetInfo retrieves resource information
func (rb *ResourceBuilder) GetInfo() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"id":       rb.resourceId,
			"pathPart": "users",
			"path":     "/users",
		},
		ctx: rb.ctx,
	}
}

// ResourceCreateBuilder builds resource creation operations
type ResourceCreateBuilder struct {
	ctx      *TestContext
	apiId    string
	parentId string
	config   *ResourceCreateConfig
}

type ResourceCreateConfig struct {
	PathPart string
}

// WithPathPart sets the path part for the resource
func (rcb *ResourceCreateBuilder) WithPathPart(pathPart string) *ResourceCreateBuilder {
	rcb.config.PathPart = pathPart
	return rcb
}

// Create creates the resource
func (rcb *ResourceCreateBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"id":       "xyz789",
			"pathPart": rcb.config.PathPart,
			"parentId": rcb.parentId,
		},
		ctx: rcb.ctx,
	}
}

// MethodBuilder builds operations on API methods
type MethodBuilder struct {
	ctx        *TestContext
	apiId      string
	resourceId string
	httpMethod string
	config     *MethodConfig
}

type MethodConfig struct {
	AuthorizationType   string
	AuthorizerId        string
	ApiKeyRequired      bool
	RequestParameters   map[string]bool
	RequestModels       map[string]string
	RequestValidatorId  string
}

// WithAuthorization sets the authorization type
func (mb *MethodBuilder) WithAuthorization(authType string) *MethodBuilder {
	mb.config.AuthorizationType = authType
	return mb
}

// WithAuthorizer sets a custom authorizer
func (mb *MethodBuilder) WithAuthorizer(authorizerId string) *MethodBuilder {
	mb.config.AuthorizerId = authorizerId
	return mb
}

// RequireAPIKey requires an API key for this method
func (mb *MethodBuilder) RequireAPIKey() *MethodBuilder {
	mb.config.ApiKeyRequired = true
	return mb
}

// WithRequestParameter adds a request parameter
func (mb *MethodBuilder) WithRequestParameter(name string, required bool) *MethodBuilder {
	if mb.config.RequestParameters == nil {
		mb.config.RequestParameters = make(map[string]bool)
	}
	mb.config.RequestParameters[name] = required
	return mb
}

// Create creates the method
func (mb *MethodBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"httpMethod":        mb.httpMethod,
			"authorizationType": mb.config.AuthorizationType,
			"apiKeyRequired":    mb.config.ApiKeyRequired,
		},
		ctx: mb.ctx,
	}
}

// Integration creates an integration builder for the method
func (mb *MethodBuilder) Integration() *IntegrationBuilder {
	return &IntegrationBuilder{
		ctx:        mb.ctx,
		apiId:      mb.apiId,
		resourceId: mb.resourceId,
		httpMethod: mb.httpMethod,
		config:     &IntegrationConfig{},
	}
}

// Delete deletes the method
func (mb *MethodBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Method %s deleted from resource %s", mb.httpMethod, mb.resourceId),
		ctx:     mb.ctx,
	}
}

// IntegrationBuilder builds integration operations
type IntegrationBuilder struct {
	ctx        *TestContext
	apiId      string
	resourceId string
	httpMethod string
	config     *IntegrationConfig
}

type IntegrationConfig struct {
	Type                    string
	IntegrationHttpMethod   string
	Uri                     string
	Credentials             string
	RequestParameters       map[string]string
	RequestTemplates        map[string]string
	TimeoutInMillis         int32
	ConnectionType          string
	ConnectionId            string
}

// WithType sets the integration type
func (ib *IntegrationBuilder) WithType(integrationType string) *IntegrationBuilder {
	ib.config.Type = integrationType
	return ib
}

// WithLambdaFunction integrates with a Lambda function
func (ib *IntegrationBuilder) WithLambdaFunction(functionArn string) *IntegrationBuilder {
	ib.config.Type = "AWS_PROXY"
	ib.config.IntegrationHttpMethod = "POST"
	ib.config.Uri = functionArn
	return ib
}

// WithHTTPEndpoint integrates with an HTTP endpoint
func (ib *IntegrationBuilder) WithHTTPEndpoint(uri string) *IntegrationBuilder {
	ib.config.Type = "HTTP_PROXY"
	ib.config.Uri = uri
	return ib
}

// WithCredentials sets the IAM credentials
func (ib *IntegrationBuilder) WithCredentials(credentials string) *IntegrationBuilder {
	ib.config.Credentials = credentials
	return ib
}

// Create creates the integration
func (ib *IntegrationBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"type":        ib.config.Type,
			"httpMethod":  ib.config.IntegrationHttpMethod,
			"uri":         ib.config.Uri,
		},
		ctx: ib.ctx,
	}
}

// Update updates the integration
func (ib *IntegrationBuilder) Update() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    "Integration updated successfully",
		ctx:     ib.ctx,
	}
}

// StageBuilder builds operations on API stages
type StageBuilder struct {
	ctx       *TestContext
	apiId     string
	stageName string
	config    *StageConfig
}

type StageConfig struct {
	DeploymentId        string
	Description         string
	CacheClusterEnabled bool
	CacheClusterSize    string
	Variables           map[string]string
	TracingConfig       *TracingConfig
}

type TracingConfig struct {
	TracingEnabled bool
}

// WithDeployment sets the deployment for the stage
func (sb *StageBuilder) WithDeployment(deploymentId string) *StageBuilder {
	sb.config.DeploymentId = deploymentId
	return sb
}

// WithDescription sets the stage description
func (sb *StageBuilder) WithDescription(description string) *StageBuilder {
	sb.config.Description = description
	return sb
}

// WithCaching enables caching for the stage
func (sb *StageBuilder) WithCaching(size string) *StageBuilder {
	sb.config.CacheClusterEnabled = true
	sb.config.CacheClusterSize = size
	return sb
}

// WithVariables sets stage variables
func (sb *StageBuilder) WithVariables(variables map[string]string) *StageBuilder {
	sb.config.Variables = variables
	return sb
}

// WithTracing enables X-Ray tracing
func (sb *StageBuilder) WithTracing() *StageBuilder {
	sb.config.TracingConfig = &TracingConfig{TracingEnabled: true}
	return sb
}

// Create creates the stage
func (sb *StageBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"stageName":    sb.stageName,
			"deploymentId": sb.config.DeploymentId,
			"createdDate":  time.Now(),
		},
		ctx: sb.ctx,
	}
}

// Update updates the stage
func (sb *StageBuilder) Update() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Stage %s updated", sb.stageName),
		ctx:     sb.ctx,
	}
}

// Delete deletes the stage
func (sb *StageBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Stage %s deleted", sb.stageName),
		ctx:     sb.ctx,
	}
}

// DeploymentBuilder builds deployment operations
type DeploymentBuilder struct {
	ctx    *TestContext
	apiId  string
	config *DeploymentConfig
}

type DeploymentConfig struct {
	StageName   string
	Description string
	Variables   map[string]string
}

// ToStage sets the target stage for deployment
func (db *DeploymentBuilder) ToStage(stageName string) *DeploymentBuilder {
	db.config.StageName = stageName
	return db
}

// WithDescription sets the deployment description
func (db *DeploymentBuilder) WithDescription(description string) *DeploymentBuilder {
	db.config.Description = description
	return db
}

// Create creates the deployment
func (db *DeploymentBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"id":          "deploy123",
			"description": db.config.Description,
			"createdDate": time.Now(),
		},
		ctx: db.ctx,
	}
}

// TestInvokeBuilder builds test invoke operations
type TestInvokeBuilder struct {
	ctx    *TestContext
	apiId  string
	config *TestInvokeConfig
}

type TestInvokeConfig struct {
	ResourceId      string
	HttpMethod      string
	PathWithQueryString string
	Body            string
	Headers         map[string]string
	StageVariables  map[string]string
}

// OnResource sets the resource to test
func (tib *TestInvokeBuilder) OnResource(resourceId string) *TestInvokeBuilder {
	tib.config.ResourceId = resourceId
	return tib
}

// WithMethod sets the HTTP method
func (tib *TestInvokeBuilder) WithMethod(httpMethod string) *TestInvokeBuilder {
	tib.config.HttpMethod = httpMethod
	return tib
}

// WithPath sets the path with query string
func (tib *TestInvokeBuilder) WithPath(path string) *TestInvokeBuilder {
	tib.config.PathWithQueryString = path
	return tib
}

// WithBody sets the request body
func (tib *TestInvokeBuilder) WithBody(body string) *TestInvokeBuilder {
	tib.config.Body = body
	return tib
}

// WithHeaders sets the request headers
func (tib *TestInvokeBuilder) WithHeaders(headers map[string]string) *TestInvokeBuilder {
	tib.config.Headers = headers
	return tib
}

// Execute executes the test invoke
func (tib *TestInvokeBuilder) Execute() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"status":  200,
			"body":    "{\"message\": \"Hello World\"}",
			"headers": map[string]string{"Content-Type": "application/json"},
			"latency": 123,
		},
		ctx: tib.ctx,
	}
}

// HttpAPIBuilder builds operations on HTTP APIs (API Gateway v2)
type HttpAPIBuilder struct {
	ctx    *TestContext
	apiId  string
	config *HttpAPIConfig
}

type HttpAPIConfig struct {
	Name        string
	Description string
	Version     string
}

// Route creates a route builder for the HTTP API
func (hab *HttpAPIBuilder) Route(routeKey string) *RouteBuilder {
	return &RouteBuilder{
		ctx:      hab.ctx,
		apiId:    hab.apiId,
		routeKey: routeKey,
		config:   &RouteConfig{},
	}
}

// Delete deletes the HTTP API
func (hab *HttpAPIBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("HTTP API %s deleted", hab.apiId),
		ctx:     hab.ctx,
	}
}

// RouteBuilder builds operations on HTTP API routes
type RouteBuilder struct {
	ctx      *TestContext
	apiId    string
	routeKey string
	config   *RouteConfig
}

type RouteConfig struct {
	Target              string
	AuthorizationType   string
	AuthorizerId        string
	ApiKeyRequired      bool
	RequestParameters   map[string]bool
}

// WithTarget sets the route target
func (rb *RouteBuilder) WithTarget(target string) *RouteBuilder {
	rb.config.Target = target
	return rb
}

// Create creates the route
func (rb *RouteBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"routeId":  "route123",
			"routeKey": rb.routeKey,
			"target":   rb.config.Target,
		},
		ctx: rb.ctx,
	}
}

// DomainNameBuilder builds operations on custom domain names
type DomainNameBuilder struct {
	ctx        *TestContext
	domainName string
	config     *DomainNameConfig
}

type DomainNameConfig struct {
	CertificateArn      string
	SecurityPolicy      string
	EndpointType        string
	ValidationMethod    string
}

// WithCertificate sets the SSL certificate
func (dnb *DomainNameBuilder) WithCertificate(certificateArn string) *DomainNameBuilder {
	dnb.config.CertificateArn = certificateArn
	return dnb
}

// WithSecurityPolicy sets the security policy
func (dnb *DomainNameBuilder) WithSecurityPolicy(policy string) *DomainNameBuilder {
	dnb.config.SecurityPolicy = policy
	return dnb
}

// AsEdgeOptimized sets the endpoint type to edge-optimized
func (dnb *DomainNameBuilder) AsEdgeOptimized() *DomainNameBuilder {
	dnb.config.EndpointType = "EDGE"
	return dnb
}

// AsRegional sets the endpoint type to regional
func (dnb *DomainNameBuilder) AsRegional() *DomainNameBuilder {
	dnb.config.EndpointType = "REGIONAL"
	return dnb
}

// Create creates the custom domain name
func (dnb *DomainNameBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"domainName":               dnb.domainName,
			"certificateArn":          dnb.config.CertificateArn,
			"distributionDomainName":   "d123456789.cloudfront.net",
			"domainNameStatus":        "AVAILABLE",
		},
		ctx: dnb.ctx,
	}
}

// Delete deletes the custom domain name
func (dnb *DomainNameBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Domain name %s deleted", dnb.domainName),
		ctx:     dnb.ctx,
	}
}