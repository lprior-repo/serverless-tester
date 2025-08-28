// Package apigateway provides comprehensive AWS API Gateway testing utilities following Terratest patterns.
//
// This package is designed to be a drop-in replacement for Terratest's API Gateway utilities
// with enhanced serverless-specific features. It follows strict functional programming
// principles and provides both Function and FunctionE variants for all operations.
//
// Key Features:
//   - REST API lifecycle management (Create/Update/Delete/Deploy)
//   - Resource and method configuration with authorization
//   - Stage management with deployment tracking
//   - Model and request validation support
//   - Domain name and certificate integration
//   - API documentation generation and export
//   - CORS configuration and options handling
//   - Throttling and quota management
//   - Request/response transformation
//   - Comprehensive API testing utilities
//
// Example usage:
//
//   func TestAPIGatewayDeployment(t *testing.T) {
//       ctx := sfx.NewTestContext(t)
//       
//       // Create REST API
//       api := CreateRestApi(ctx, RestApiConfig{
//           Name:        "test-api",
//           Description: "Test API for serverless application",
//       })
//       
//       // Create resource and method
//       resource := CreateResource(ctx, api.Id, "/users", api.RootResourceId)
//       method := PutMethod(ctx, api.Id, resource.Id, "GET", "NONE")
//       
//       // Deploy API
//       deployment := CreateDeployment(ctx, api.Id, "test", "Test deployment")
//       
//       assert.NotEmpty(t, deployment.Id)
//   }
package apigateway

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
)

// Common API Gateway configurations
const (
	DefaultTimeout        = 30 * time.Second
	DefaultRetryAttempts  = 3
	DefaultRetryDelay     = 1 * time.Second
	MaxAPINameLength      = 255
	MaxDescriptionLength  = 1024
	DefaultThrottleLimit  = 10000
	DefaultBurstLimit     = 5000
)

// Authorization types
const (
	AuthorizationTypeNone          = "NONE"
	AuthorizationTypeAWSIAM        = "AWS_IAM"
	AuthorizationTypeCUSTOM        = "CUSTOM"
	AuthorizationTypeCOGNITOUSER   = "COGNITO_USER_POOLS"
)

// Integration types
const (
	IntegrationTypeHTTP        = types.IntegrationTypeHttp
	IntegrationTypeHTTPProxy   = types.IntegrationTypeHttpProxy
	IntegrationTypeAWS         = types.IntegrationTypeAws
	IntegrationTypeAWSProxy    = types.IntegrationTypeAwsProxy
	IntegrationTypeMOCK        = types.IntegrationTypeMock
)

// Common errors
var (
	ErrAPINotFound         = errors.New("API not found")
	ErrResourceNotFound    = errors.New("resource not found")
	ErrMethodNotFound      = errors.New("method not found")
	ErrStageNotFound       = errors.New("stage not found")
	ErrDeploymentNotFound  = errors.New("deployment not found")
	ErrDomainNotFound      = errors.New("domain not found")
	ErrInvalidAPIName      = errors.New("invalid API name")
	ErrInvalidStageName    = errors.New("invalid stage name")
	ErrInvalidMethodType   = errors.New("invalid HTTP method type")
	ErrInvalidPathPart     = errors.New("invalid path part")
)

// TestingT provides interface compatibility with testing frameworks
type TestingT interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Fail()
	FailNow()
	Helper()
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Name() string
}

// TestContext represents the testing context with AWS configuration
type TestContext struct {
	T         TestingT
	AwsConfig aws.Config
	Region    string
}

// =============================================================================
// REST API MANAGEMENT
// =============================================================================

// RestApiConfig represents configuration for creating a REST API
type RestApiConfig struct {
	Name            string
	Description     string
	Version         string
	CloneFrom       string
	BinaryMediaTypes []string
	MinimumCompressionSize *int32
	ApiKeySourceType       types.ApiKeySourceType
	EndpointConfiguration  *types.EndpointConfiguration
	Policy                 string
	Tags                   map[string]string
	DisableExecuteApiEndpoint *bool
}

// RestApiResult represents the result of REST API operations
type RestApiResult struct {
	Id                        string
	Name                      string
	Description               string
	CreatedDate               *time.Time
	Version                   string
	Warnings                  []string
	BinaryMediaTypes          []string
	MinimumCompressionSize    *int32
	ApiKeySourceType          types.ApiKeySourceType
	EndpointConfiguration     *types.EndpointConfiguration
	Policy                    string
	Tags                      map[string]string
	DisableExecuteApiEndpoint *bool
	RootResourceId            string
}

// CreateRestApi creates a new REST API and fails the test if an error occurs
func CreateRestApi(ctx *TestContext, config RestApiConfig) *RestApiResult {
	result, err := CreateRestApiE(ctx, config)
	if err != nil {
		ctx.T.Fatalf("Failed to create REST API: %v", err)
	}
	return result
}

// CreateRestApiE creates a new REST API and returns any error
func CreateRestApiE(ctx *TestContext, config RestApiConfig) (*RestApiResult, error) {
	if err := validateApiName(config.Name); err != nil {
		return nil, fmt.Errorf("invalid API name: %w", err)
	}

	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.CreateRestApiInput{
		Name: aws.String(config.Name),
	}

	if config.Description != "" {
		input.Description = aws.String(config.Description)
	}
	if config.Version != "" {
		input.Version = aws.String(config.Version)
	}
	if config.CloneFrom != "" {
		input.CloneFrom = aws.String(config.CloneFrom)
	}
	if len(config.BinaryMediaTypes) > 0 {
		input.BinaryMediaTypes = config.BinaryMediaTypes
	}
	if config.MinimumCompressionSize != nil {
		input.MinimumCompressionSize = config.MinimumCompressionSize
	}
	if config.ApiKeySourceType != "" {
		input.ApiKeySourceType = config.ApiKeySourceType
	}
	if config.EndpointConfiguration != nil {
		input.EndpointConfiguration = config.EndpointConfiguration
	}
	if config.Policy != "" {
		input.Policy = aws.String(config.Policy)
	}
	if len(config.Tags) > 0 {
		input.Tags = config.Tags
	}
	if config.DisableExecuteApiEndpoint != nil {
		input.DisableExecuteApiEndpoint = config.DisableExecuteApiEndpoint
	}

	output, err := client.CreateRestApi(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST API: %w", err)
	}

	return &RestApiResult{
		Id:                        aws.ToString(output.Id),
		Name:                      aws.ToString(output.Name),
		Description:               aws.ToString(output.Description),
		CreatedDate:               output.CreatedDate,
		Version:                   aws.ToString(output.Version),
		Warnings:                  output.Warnings,
		BinaryMediaTypes:          output.BinaryMediaTypes,
		MinimumCompressionSize:    output.MinimumCompressionSize,
		ApiKeySourceType:          output.ApiKeySourceType,
		EndpointConfiguration:     output.EndpointConfiguration,
		Policy:                    aws.ToString(output.Policy),
		Tags:                      output.Tags,
		DisableExecuteApiEndpoint: output.DisableExecuteApiEndpoint,
	}, nil
}

// GetRestApi gets information about a REST API and fails the test if an error occurs
func GetRestApi(ctx *TestContext, apiId string) *RestApiResult {
	result, err := GetRestApiE(ctx, apiId)
	if err != nil {
		ctx.T.Fatalf("Failed to get REST API: %v", err)
	}
	return result
}

// GetRestApiE gets information about a REST API and returns any error
func GetRestApiE(ctx *TestContext, apiId string) (*RestApiResult, error) {
	if apiId == "" {
		return nil, fmt.Errorf("API ID cannot be empty")
	}

	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	output, err := client.GetRestApi(context.Background(), &apigateway.GetRestApiInput{
		RestApiId: aws.String(apiId),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get REST API: %w", err)
	}

	// Get root resource ID
	resourcesOutput, err := client.GetResources(context.Background(), &apigateway.GetResourcesInput{
		RestApiId: aws.String(apiId),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get API resources: %w", err)
	}

	rootResourceId := ""
	for _, resource := range resourcesOutput.Items {
		if aws.ToString(resource.PathPart) == "" && aws.ToString(resource.Path) == "/" {
			rootResourceId = aws.ToString(resource.Id)
			break
		}
	}

	return &RestApiResult{
		Id:                        aws.ToString(output.Id),
		Name:                      aws.ToString(output.Name),
		Description:               aws.ToString(output.Description),
		CreatedDate:               output.CreatedDate,
		Version:                   aws.ToString(output.Version),
		Warnings:                  output.Warnings,
		BinaryMediaTypes:          output.BinaryMediaTypes,
		MinimumCompressionSize:    output.MinimumCompressionSize,
		ApiKeySourceType:          output.ApiKeySourceType,
		EndpointConfiguration:     output.EndpointConfiguration,
		Policy:                    aws.ToString(output.Policy),
		Tags:                      output.Tags,
		DisableExecuteApiEndpoint: output.DisableExecuteApiEndpoint,
		RootResourceId:            rootResourceId,
	}, nil
}

// DeleteRestApi deletes a REST API and fails the test if an error occurs
func DeleteRestApi(ctx *TestContext, apiId string) {
	err := DeleteRestApiE(ctx, apiId)
	if err != nil {
		ctx.T.Fatalf("Failed to delete REST API: %v", err)
	}
}

// DeleteRestApiE deletes a REST API and returns any error
func DeleteRestApiE(ctx *TestContext, apiId string) error {
	if apiId == "" {
		return fmt.Errorf("API ID cannot be empty")
	}

	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.DeleteRestApi(context.Background(), &apigateway.DeleteRestApiInput{
		RestApiId: aws.String(apiId),
	})
	if err != nil {
		return fmt.Errorf("failed to delete REST API: %w", err)
	}

	return nil
}

// =============================================================================
// RESOURCE MANAGEMENT
// =============================================================================

// ResourceResult represents the result of resource operations
type ResourceResult struct {
	Id            string
	ParentId      string
	PathPart      string
	Path          string
	ResourceMethods map[string]types.Method
}

// CreateResource creates a new API resource and fails the test if an error occurs
func CreateResource(ctx *TestContext, apiId string, pathPart string, parentId string) *ResourceResult {
	result, err := CreateResourceE(ctx, apiId, pathPart, parentId)
	if err != nil {
		ctx.T.Fatalf("Failed to create resource: %v", err)
	}
	return result
}

// CreateResourceE creates a new API resource and returns any error
func CreateResourceE(ctx *TestContext, apiId string, pathPart string, parentId string) (*ResourceResult, error) {
	if apiId == "" {
		return nil, fmt.Errorf("API ID cannot be empty")
	}
	if pathPart == "" {
		return nil, fmt.Errorf("path part cannot be empty")
	}
	if parentId == "" {
		return nil, fmt.Errorf("parent ID cannot be empty")
	}

	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	output, err := client.CreateResource(context.Background(), &apigateway.CreateResourceInput{
		RestApiId: aws.String(apiId),
		ParentId:  aws.String(parentId),
		PathPart:  aws.String(pathPart),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return &ResourceResult{
		Id:              aws.ToString(output.Id),
		ParentId:        aws.ToString(output.ParentId),
		PathPart:        aws.ToString(output.PathPart),
		Path:            aws.ToString(output.Path),
		ResourceMethods: output.ResourceMethods,
	}, nil
}

// =============================================================================
// METHOD MANAGEMENT
// =============================================================================

// MethodResult represents the result of method operations
type MethodResult struct {
	HttpMethod         string
	AuthorizationType  string
	AuthorizerId       string
	ApiKeyRequired     bool
	RequestValidatorId string
	RequestModels      map[string]string
	RequestParameters  map[string]bool
	MethodIntegration  *types.Integration
	MethodResponses    map[string]types.MethodResponse
}

// PutMethod creates or updates a method and fails the test if an error occurs
func PutMethod(ctx *TestContext, apiId string, resourceId string, httpMethod string, authorizationType string) *MethodResult {
	result, err := PutMethodE(ctx, apiId, resourceId, httpMethod, authorizationType)
	if err != nil {
		ctx.T.Fatalf("Failed to put method: %v", err)
	}
	return result
}

// PutMethodE creates or updates a method and returns any error
func PutMethodE(ctx *TestContext, apiId string, resourceId string, httpMethod string, authorizationType string) (*MethodResult, error) {
	if apiId == "" {
		return nil, fmt.Errorf("API ID cannot be empty")
	}
	if resourceId == "" {
		return nil, fmt.Errorf("resource ID cannot be empty")
	}
	if httpMethod == "" {
		return nil, fmt.Errorf("HTTP method cannot be empty")
	}
	if authorizationType == "" {
		return nil, fmt.Errorf("authorization type cannot be empty")
	}

	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	output, err := client.PutMethod(context.Background(), &apigateway.PutMethodInput{
		RestApiId:         aws.String(apiId),
		ResourceId:        aws.String(resourceId),
		HttpMethod:        aws.String(httpMethod),
		AuthorizationType: aws.String(authorizationType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to put method: %w", err)
	}

	return &MethodResult{
		HttpMethod:         aws.ToString(output.HttpMethod),
		AuthorizationType:  aws.ToString(output.AuthorizationType),
		AuthorizerId:       aws.ToString(output.AuthorizerId),
		ApiKeyRequired:     aws.ToBool(output.ApiKeyRequired),
		RequestValidatorId: aws.ToString(output.RequestValidatorId),
		RequestModels:      output.RequestModels,
		RequestParameters:  output.RequestParameters,
		MethodIntegration:  output.MethodIntegration,
		MethodResponses:    output.MethodResponses,
	}, nil
}

// =============================================================================
// DEPLOYMENT MANAGEMENT
// =============================================================================

// DeploymentResult represents the result of deployment operations
type DeploymentResult struct {
	Id           string
	Description  string
	CreatedDate  *time.Time
	ApiSummary   map[string]map[string]types.MethodSnapshot
}

// CreateDeployment creates a new deployment and fails the test if an error occurs
func CreateDeployment(ctx *TestContext, apiId string, stageName string, description string) *DeploymentResult {
	result, err := CreateDeploymentE(ctx, apiId, stageName, description)
	if err != nil {
		ctx.T.Fatalf("Failed to create deployment: %v", err)
	}
	return result
}

// CreateDeploymentE creates a new deployment and returns any error
func CreateDeploymentE(ctx *TestContext, apiId string, stageName string, description string) (*DeploymentResult, error) {
	if apiId == "" {
		return nil, fmt.Errorf("API ID cannot be empty")
	}
	if stageName == "" {
		return nil, fmt.Errorf("stage name cannot be empty")
	}

	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.CreateDeploymentInput{
		RestApiId: aws.String(apiId),
		StageName: aws.String(stageName),
	}

	if description != "" {
		input.Description = aws.String(description)
	}

	output, err := client.CreateDeployment(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	return &DeploymentResult{
		Id:          aws.ToString(output.Id),
		Description: aws.ToString(output.Description),
		CreatedDate: output.CreatedDate,
		ApiSummary:  output.ApiSummary,
	}, nil
}

// =============================================================================
// STAGE MANAGEMENT
// =============================================================================

// StageConfig represents configuration for creating a stage
type StageConfig struct {
	StageName            string
	DeploymentId         string
	Description          string
	CacheClusterEnabled  *bool
	CacheClusterSize     types.CacheClusterSize
	Variables            map[string]string
	DocumentationVersion string
	CanarySettings       *types.CanarySettings
	TracingConfig        *types.TracingConfig
	Tags                 map[string]string
}

// StageResult represents the result of stage operations
type StageResult struct {
	DeploymentId         string
	StageName            string
	Description          string
	CacheClusterEnabled  bool
	CacheClusterSize     types.CacheClusterSize
	CacheClusterStatus   types.CacheClusterStatus
	MethodSettings       map[string]types.MethodSetting
	Variables            map[string]string
	DocumentationVersion string
	AccessLogSettings    *types.AccessLogSettings
	CanarySettings       *types.CanarySettings
	TracingConfig        *types.TracingConfig
	WebAclArn            string
	Tags                 map[string]string
	CreatedDate          *time.Time
	LastUpdatedDate      *time.Time
}

// CreateStage creates a new stage and fails the test if an error occurs
func CreateStage(ctx *TestContext, apiId string, config StageConfig) *StageResult {
	result, err := CreateStageE(ctx, apiId, config)
	if err != nil {
		ctx.T.Fatalf("Failed to create stage: %v", err)
	}
	return result
}

// CreateStageE creates a new stage and returns any error
func CreateStageE(ctx *TestContext, apiId string, config StageConfig) (*StageResult, error) {
	if apiId == "" {
		return nil, fmt.Errorf("API ID cannot be empty")
	}
	if config.StageName == "" {
		return nil, fmt.Errorf("stage name cannot be empty")
	}
	if config.DeploymentId == "" {
		return nil, fmt.Errorf("deployment ID cannot be empty")
	}

	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.CreateStageInput{
		RestApiId:    aws.String(apiId),
		StageName:    aws.String(config.StageName),
		DeploymentId: aws.String(config.DeploymentId),
	}

	if config.Description != "" {
		input.Description = aws.String(config.Description)
	}
	if config.CacheClusterEnabled != nil {
		input.CacheClusterEnabled = config.CacheClusterEnabled
	}
	if config.CacheClusterSize != "" {
		input.CacheClusterSize = config.CacheClusterSize
	}
	if len(config.Variables) > 0 {
		input.Variables = config.Variables
	}
	if config.DocumentationVersion != "" {
		input.DocumentationVersion = aws.String(config.DocumentationVersion)
	}
	if config.CanarySettings != nil {
		input.CanarySettings = config.CanarySettings
	}
	if config.TracingConfig != nil {
		input.TracingConfig = config.TracingConfig
	}
	if len(config.Tags) > 0 {
		input.Tags = config.Tags
	}

	output, err := client.CreateStage(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to create stage: %w", err)
	}

	return convertToStageResult(output), nil
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// createApiGatewayClient creates a new API Gateway client
func createApiGatewayClient(ctx *TestContext) (*apigateway.Client, error) {
	if ctx.AwsConfig.Region == "" {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		ctx.AwsConfig = cfg
	}
	return apigateway.NewFromConfig(ctx.AwsConfig), nil
}

// validateApiName validates that the API name follows AWS naming conventions
func validateApiName(name string) error {
	if name == "" {
		return fmt.Errorf("API name cannot be empty")
	}
	if len(name) > MaxAPINameLength {
		return fmt.Errorf("API name cannot exceed %d characters", MaxAPINameLength)
	}
	return nil
}

// GetStage gets an existing stage and fails the test if an error occurs
func GetStage(ctx *TestContext, apiId, stageName string) *StageResult {
	result, err := GetStageE(ctx, apiId, stageName)
	if err != nil {
		ctx.T.Fatalf("Failed to get stage: %v", err)
	}
	return result
}

// GetStageE gets an existing stage and returns any error
func GetStageE(ctx *TestContext, apiId, stageName string) (*StageResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.GetStageInput{
		RestApiId: aws.String(apiId),
		StageName: aws.String(stageName),
	}

	output, err := client.GetStage(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertGetStageToStageResult(output), nil
}

// UpdateStage updates an existing stage and fails the test if an error occurs
func UpdateStage(ctx *TestContext, apiId, stageName string, patchOps []types.PatchOp) *StageResult {
	result, err := UpdateStageE(ctx, apiId, stageName, patchOps)
	if err != nil {
		ctx.T.Fatalf("Failed to update stage: %v", err)
	}
	return result
}

// UpdateStageE updates an existing stage and returns any error
func UpdateStageE(ctx *TestContext, apiId, stageName string, patchOps []types.PatchOp) (*StageResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.UpdateStageInput{
		RestApiId:    aws.String(apiId),
		StageName:    aws.String(stageName),
		PatchOps:     patchOps,
	}

	output, err := client.UpdateStage(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertUpdateStageToStageResult(output), nil
}

// DeleteStage deletes an existing stage and fails the test if an error occurs
func DeleteStage(ctx *TestContext, apiId, stageName string) {
	err := DeleteStageE(ctx, apiId, stageName)
	if err != nil {
		ctx.T.Fatalf("Failed to delete stage: %v", err)
	}
}

// DeleteStageE deletes an existing stage and returns any error
func DeleteStageE(ctx *TestContext, apiId, stageName string) error {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return err
	}

	input := &apigateway.DeleteStageInput{
		RestApiId: aws.String(apiId),
		StageName: aws.String(stageName),
	}

	_, err = client.DeleteStage(context.TODO(), input)
	return err
}

// =============================================================================
// METHOD MANAGEMENT
// =============================================================================

// GetMethod gets an existing method and fails the test if an error occurs
func GetMethod(ctx *TestContext, apiId, resourceId, httpMethod string) *MethodResult {
	result, err := GetMethodE(ctx, apiId, resourceId, httpMethod)
	if err != nil {
		ctx.T.Fatalf("Failed to get method: %v", err)
	}
	return result
}

// GetMethodE gets an existing method and returns any error
func GetMethodE(ctx *TestContext, apiId, resourceId, httpMethod string) (*MethodResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.GetMethodInput{
		RestApiId:  aws.String(apiId),
		ResourceId: aws.String(resourceId),
		HttpMethod: aws.String(httpMethod),
	}

	output, err := client.GetMethod(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertGetMethodToMethodResult(output), nil
}

// UpdateMethod updates an existing method and fails the test if an error occurs
func UpdateMethod(ctx *TestContext, apiId, resourceId, httpMethod string, patchOps []types.PatchOp) *MethodResult {
	result, err := UpdateMethodE(ctx, apiId, resourceId, httpMethod, patchOps)
	if err != nil {
		ctx.T.Fatalf("Failed to update method: %v", err)
	}
	return result
}

// UpdateMethodE updates an existing method and returns any error
func UpdateMethodE(ctx *TestContext, apiId, resourceId, httpMethod string, patchOps []types.PatchOp) (*MethodResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.UpdateMethodInput{
		RestApiId:  aws.String(apiId),
		ResourceId: aws.String(resourceId),
		HttpMethod: aws.String(httpMethod),
		PatchOps:   patchOps,
	}

	output, err := client.UpdateMethod(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertUpdateMethodToMethodResult(output), nil
}

// DeleteMethod deletes an existing method and fails the test if an error occurs
func DeleteMethod(ctx *TestContext, apiId, resourceId, httpMethod string) {
	err := DeleteMethodE(ctx, apiId, resourceId, httpMethod)
	if err != nil {
		ctx.T.Fatalf("Failed to delete method: %v", err)
	}
}

// DeleteMethodE deletes an existing method and returns any error
func DeleteMethodE(ctx *TestContext, apiId, resourceId, httpMethod string) error {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return err
	}

	input := &apigateway.DeleteMethodInput{
		RestApiId:  aws.String(apiId),
		ResourceId: aws.String(resourceId),
		HttpMethod: aws.String(httpMethod),
	}

	_, err = client.DeleteMethod(context.TODO(), input)
	return err
}

// =============================================================================
// INTEGRATION MANAGEMENT
// =============================================================================

// IntegrationConfig represents configuration for method integration
type IntegrationConfig struct {
	Type                    types.IntegrationType     `json:"type"`
	HttpMethod              string                    `json:"httpMethod,omitempty"`
	Uri                     string                    `json:"uri,omitempty"`
	ConnectionType          types.ConnectionType      `json:"connectionType,omitempty"`
	ConnectionId            string                    `json:"connectionId,omitempty"`
	Credentials             string                    `json:"credentials,omitempty"`
	RequestParameters       map[string]string         `json:"requestParameters,omitempty"`
	RequestTemplates        map[string]string         `json:"requestTemplates,omitempty"`
	PassthroughBehavior     string                    `json:"passthroughBehavior,omitempty"`
	CacheNamespace          string                    `json:"cacheNamespace,omitempty"`
	CacheKeyParameters      []string                  `json:"cacheKeyParameters,omitempty"`
	ContentHandling         types.ContentHandlingStrategy `json:"contentHandling,omitempty"`
	TimeoutInMillis         int32                     `json:"timeoutInMillis,omitempty"`
	TlsConfig               *types.TlsConfig          `json:"tlsConfig,omitempty"`
}

// IntegrationResult represents the result of integration operations
type IntegrationResult struct {
	Type                    types.IntegrationType         `json:"type"`
	HttpMethod              string                        `json:"httpMethod,omitempty"`
	Uri                     string                        `json:"uri,omitempty"`
	ConnectionType          types.ConnectionType          `json:"connectionType,omitempty"`
	ConnectionId            string                        `json:"connectionId,omitempty"`
	Credentials             string                        `json:"credentials,omitempty"`
	RequestParameters       map[string]string             `json:"requestParameters,omitempty"`
	RequestTemplates        map[string]string             `json:"requestTemplates,omitempty"`
	PassthroughBehavior     string                        `json:"passthroughBehavior,omitempty"`
	CacheNamespace          string                        `json:"cacheNamespace,omitempty"`
	CacheKeyParameters      []string                      `json:"cacheKeyParameters,omitempty"`
	ContentHandling         types.ContentHandlingStrategy `json:"contentHandling,omitempty"`
	TimeoutInMillis         int32                         `json:"timeoutInMillis,omitempty"`
	IntegrationResponses    map[string]types.IntegrationResponse `json:"integrationResponses,omitempty"`
	TlsConfig               *types.TlsConfig              `json:"tlsConfig,omitempty"`
}

// PutIntegration creates or updates a method integration and fails the test if an error occurs
func PutIntegration(ctx *TestContext, apiId, resourceId, httpMethod string, config IntegrationConfig) *IntegrationResult {
	result, err := PutIntegrationE(ctx, apiId, resourceId, httpMethod, config)
	if err != nil {
		ctx.T.Fatalf("Failed to put integration: %v", err)
	}
	return result
}

// PutIntegrationE creates or updates a method integration and returns any error
func PutIntegrationE(ctx *TestContext, apiId, resourceId, httpMethod string, config IntegrationConfig) (*IntegrationResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.PutIntegrationInput{
		RestApiId:  aws.String(apiId),
		ResourceId: aws.String(resourceId),
		HttpMethod: aws.String(httpMethod),
		Type:       config.Type,
	}

	if config.HttpMethod != "" {
		input.IntegrationHttpMethod = aws.String(config.HttpMethod)
	}
	if config.Uri != "" {
		input.Uri = aws.String(config.Uri)
	}
	if config.ConnectionType != "" {
		input.ConnectionType = config.ConnectionType
	}
	if config.ConnectionId != "" {
		input.ConnectionId = aws.String(config.ConnectionId)
	}
	if config.Credentials != "" {
		input.Credentials = aws.String(config.Credentials)
	}
	if len(config.RequestParameters) > 0 {
		input.RequestParameters = config.RequestParameters
	}
	if len(config.RequestTemplates) > 0 {
		input.RequestTemplates = config.RequestTemplates
	}
	if config.PassthroughBehavior != "" {
		input.PassthroughBehavior = aws.String(config.PassthroughBehavior)
	}
	if config.CacheNamespace != "" {
		input.CacheNamespace = aws.String(config.CacheNamespace)
	}
	if len(config.CacheKeyParameters) > 0 {
		input.CacheKeyParameters = config.CacheKeyParameters
	}
	if config.ContentHandling != "" {
		input.ContentHandling = config.ContentHandling
	}
	if config.TimeoutInMillis > 0 {
		input.TimeoutInMillis = aws.Int32(config.TimeoutInMillis)
	}
	if config.TlsConfig != nil {
		input.TlsConfig = config.TlsConfig
	}

	output, err := client.PutIntegration(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertToIntegrationResult(output), nil
}

// GetIntegration gets a method integration and fails the test if an error occurs
func GetIntegration(ctx *TestContext, apiId, resourceId, httpMethod string) *IntegrationResult {
	result, err := GetIntegrationE(ctx, apiId, resourceId, httpMethod)
	if err != nil {
		ctx.T.Fatalf("Failed to get integration: %v", err)
	}
	return result
}

// GetIntegrationE gets a method integration and returns any error
func GetIntegrationE(ctx *TestContext, apiId, resourceId, httpMethod string) (*IntegrationResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.GetIntegrationInput{
		RestApiId:  aws.String(apiId),
		ResourceId: aws.String(resourceId),
		HttpMethod: aws.String(httpMethod),
	}

	output, err := client.GetIntegration(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertGetIntegrationToResult(output), nil
}

// DeleteIntegration deletes a method integration and fails the test if an error occurs
func DeleteIntegration(ctx *TestContext, apiId, resourceId, httpMethod string) {
	err := DeleteIntegrationE(ctx, apiId, resourceId, httpMethod)
	if err != nil {
		ctx.T.Fatalf("Failed to delete integration: %v", err)
	}
}

// DeleteIntegrationE deletes a method integration and returns any error
func DeleteIntegrationE(ctx *TestContext, apiId, resourceId, httpMethod string) error {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return err
	}

	input := &apigateway.DeleteIntegrationInput{
		RestApiId:  aws.String(apiId),
		ResourceId: aws.String(resourceId),
		HttpMethod: aws.String(httpMethod),
	}

	_, err = client.DeleteIntegration(context.TODO(), input)
	return err
}

// =============================================================================
// RESOURCE MANAGEMENT
// =============================================================================

// GetResource gets an existing resource and fails the test if an error occurs
func GetResource(ctx *TestContext, apiId, resourceId string) *ResourceResult {
	result, err := GetResourceE(ctx, apiId, resourceId)
	if err != nil {
		ctx.T.Fatalf("Failed to get resource: %v", err)
	}
	return result
}

// GetResourceE gets an existing resource and returns any error
func GetResourceE(ctx *TestContext, apiId, resourceId string) (*ResourceResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.GetResourceInput{
		RestApiId:  aws.String(apiId),
		ResourceId: aws.String(resourceId),
	}

	output, err := client.GetResource(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertGetResourceToResult(output), nil
}

// GetResources lists all resources for an API and fails the test if an error occurs
func GetResources(ctx *TestContext, apiId string) []ResourceResult {
	results, err := GetResourcesE(ctx, apiId)
	if err != nil {
		ctx.T.Fatalf("Failed to get resources: %v", err)
	}
	return results
}

// GetResourcesE lists all resources for an API and returns any error
func GetResourcesE(ctx *TestContext, apiId string) ([]ResourceResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(apiId),
	}

	output, err := client.GetResources(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var results []ResourceResult
	for _, resource := range output.Items {
		results = append(results, *convertResourceToResult(&resource))
	}

	return results, nil
}

// DeleteResource deletes an existing resource and fails the test if an error occurs
func DeleteResource(ctx *TestContext, apiId, resourceId string) {
	err := DeleteResourceE(ctx, apiId, resourceId)
	if err != nil {
		ctx.T.Fatalf("Failed to delete resource: %v", err)
	}
}

// DeleteResourceE deletes an existing resource and returns any error
func DeleteResourceE(ctx *TestContext, apiId, resourceId string) error {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return err
	}

	input := &apigateway.DeleteResourceInput{
		RestApiId:  aws.String(apiId),
		ResourceId: aws.String(resourceId),
	}

	_, err = client.DeleteResource(context.TODO(), input)
	return err
}

// =============================================================================
// DEPLOYMENT MANAGEMENT
// =============================================================================

// GetDeployment gets an existing deployment and fails the test if an error occurs
func GetDeployment(ctx *TestContext, apiId, deploymentId string) *DeploymentResult {
	result, err := GetDeploymentE(ctx, apiId, deploymentId)
	if err != nil {
		ctx.T.Fatalf("Failed to get deployment: %v", err)
	}
	return result
}

// GetDeploymentE gets an existing deployment and returns any error
func GetDeploymentE(ctx *TestContext, apiId, deploymentId string) (*DeploymentResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.GetDeploymentInput{
		RestApiId:    aws.String(apiId),
		DeploymentId: aws.String(deploymentId),
	}

	output, err := client.GetDeployment(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertGetDeploymentToResult(output), nil
}

// GetDeployments lists all deployments for an API and fails the test if an error occurs
func GetDeployments(ctx *TestContext, apiId string) []DeploymentResult {
	results, err := GetDeploymentsE(ctx, apiId)
	if err != nil {
		ctx.T.Fatalf("Failed to get deployments: %v", err)
	}
	return results
}

// GetDeploymentsE lists all deployments for an API and returns any error
func GetDeploymentsE(ctx *TestContext, apiId string) ([]DeploymentResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.GetDeploymentsInput{
		RestApiId: aws.String(apiId),
	}

	output, err := client.GetDeployments(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var results []DeploymentResult
	for _, deployment := range output.Items {
		results = append(results, *convertDeploymentToResult(&deployment))
	}

	return results, nil
}

// DeleteDeployment deletes an existing deployment and fails the test if an error occurs
func DeleteDeployment(ctx *TestContext, apiId, deploymentId string) {
	err := DeleteDeploymentE(ctx, apiId, deploymentId)
	if err != nil {
		ctx.T.Fatalf("Failed to delete deployment: %v", err)
	}
}

// DeleteDeploymentE deletes an existing deployment and returns any error
func DeleteDeploymentE(ctx *TestContext, apiId, deploymentId string) error {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return err
	}

	input := &apigateway.DeleteDeploymentInput{
		RestApiId:    aws.String(apiId),
		DeploymentId: aws.String(deploymentId),
	}

	_, err = client.DeleteDeployment(context.TODO(), input)
	return err
}

// =============================================================================
// API KEY MANAGEMENT
// =============================================================================

// APIKeyConfig represents configuration for creating an API key
type APIKeyConfig struct {
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Enabled      bool              `json:"enabled"`
	GenerateDistinctId bool        `json:"generateDistinctId,omitempty"`
	Value        string            `json:"value,omitempty"`
	StageKeys    []types.StageKey  `json:"stageKeys,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// APIKeyResult represents the result of API key operations
type APIKeyResult struct {
	Id           string            `json:"id"`
	Value        string            `json:"value"`
	Name         string            `json:"name"`
	CustomerId   string            `json:"customerId,omitempty"`
	Description  string            `json:"description,omitempty"`
	Enabled      bool              `json:"enabled"`
	CreatedDate  *time.Time        `json:"createdDate,omitempty"`
	LastUpdatedDate *time.Time     `json:"lastUpdatedDate,omitempty"`
	StageKeys    []string          `json:"stageKeys,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// CreateAPIKey creates a new API key and fails the test if an error occurs
func CreateAPIKey(ctx *TestContext, config APIKeyConfig) *APIKeyResult {
	result, err := CreateAPIKeyE(ctx, config)
	if err != nil {
		ctx.T.Fatalf("Failed to create API key: %v", err)
	}
	return result
}

// CreateAPIKeyE creates a new API key and returns any error
func CreateAPIKeyE(ctx *TestContext, config APIKeyConfig) (*APIKeyResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.CreateApiKeyInput{
		Name:    aws.String(config.Name),
		Enabled: aws.Bool(config.Enabled),
	}

	if config.Description != "" {
		input.Description = aws.String(config.Description)
	}
	if config.GenerateDistinctId {
		input.GenerateDistinctId = aws.Bool(config.GenerateDistinctId)
	}
	if config.Value != "" {
		input.Value = aws.String(config.Value)
	}
	if len(config.StageKeys) > 0 {
		input.StageKeys = config.StageKeys
	}
	if len(config.Tags) > 0 {
		input.Tags = config.Tags
	}

	output, err := client.CreateApiKey(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertToAPIKeyResult(output), nil
}

// GetAPIKey gets an existing API key and fails the test if an error occurs
func GetAPIKey(ctx *TestContext, apiKeyId string) *APIKeyResult {
	result, err := GetAPIKeyE(ctx, apiKeyId)
	if err != nil {
		ctx.T.Fatalf("Failed to get API key: %v", err)
	}
	return result
}

// GetAPIKeyE gets an existing API key and returns any error
func GetAPIKeyE(ctx *TestContext, apiKeyId string) (*APIKeyResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.GetApiKeyInput{
		ApiKey: aws.String(apiKeyId),
	}

	output, err := client.GetApiKey(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertGetAPIKeyToResult(output), nil
}

// UpdateAPIKey updates an existing API key and fails the test if an error occurs
func UpdateAPIKey(ctx *TestContext, apiKeyId string, patchOps []types.PatchOp) *APIKeyResult {
	result, err := UpdateAPIKeyE(ctx, apiKeyId, patchOps)
	if err != nil {
		ctx.T.Fatalf("Failed to update API key: %v", err)
	}
	return result
}

// UpdateAPIKeyE updates an existing API key and returns any error
func UpdateAPIKeyE(ctx *TestContext, apiKeyId string, patchOps []types.PatchOp) (*APIKeyResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.UpdateApiKeyInput{
		ApiKey:   aws.String(apiKeyId),
		PatchOps: patchOps,
	}

	output, err := client.UpdateApiKey(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertUpdateAPIKeyToResult(output), nil
}

// DeleteAPIKey deletes an existing API key and fails the test if an error occurs
func DeleteAPIKey(ctx *TestContext, apiKeyId string) {
	err := DeleteAPIKeyE(ctx, apiKeyId)
	if err != nil {
		ctx.T.Fatalf("Failed to delete API key: %v", err)
	}
}

// DeleteAPIKeyE deletes an existing API key and returns any error
func DeleteAPIKeyE(ctx *TestContext, apiKeyId string) error {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return err
	}

	input := &apigateway.DeleteApiKeyInput{
		ApiKey: aws.String(apiKeyId),
	}

	_, err = client.DeleteApiKey(context.TODO(), input)
	return err
}

// =============================================================================
// USAGE PLAN MANAGEMENT  
// =============================================================================

// UsagePlanConfig represents configuration for creating a usage plan
type UsagePlanConfig struct {
	Name         string                       `json:"name"`
	Description  string                       `json:"description,omitempty"`
	ApiStages    []types.ApiStage             `json:"apiStages,omitempty"`
	Throttle     *types.ThrottleSettings      `json:"throttle,omitempty"`
	Quota        *types.QuotaSettings         `json:"quota,omitempty"`
	ProductCode  string                       `json:"productCode,omitempty"`
	Tags         map[string]string            `json:"tags,omitempty"`
}

// UsagePlanResult represents the result of usage plan operations
type UsagePlanResult struct {
	Id           string                       `json:"id"`
	Name         string                       `json:"name"`
	Description  string                       `json:"description,omitempty"`
	ApiStages    []types.ApiStage             `json:"apiStages,omitempty"`
	Throttle     *types.ThrottleSettings      `json:"throttle,omitempty"`
	Quota        *types.QuotaSettings         `json:"quota,omitempty"`
	ProductCode  string                       `json:"productCode,omitempty"`
	Tags         map[string]string            `json:"tags,omitempty"`
}

// CreateUsagePlan creates a new usage plan and fails the test if an error occurs
func CreateUsagePlan(ctx *TestContext, config UsagePlanConfig) *UsagePlanResult {
	result, err := CreateUsagePlanE(ctx, config)
	if err != nil {
		ctx.T.Fatalf("Failed to create usage plan: %v", err)
	}
	return result
}

// CreateUsagePlanE creates a new usage plan and returns any error
func CreateUsagePlanE(ctx *TestContext, config UsagePlanConfig) (*UsagePlanResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.CreateUsagePlanInput{
		Name: aws.String(config.Name),
	}

	if config.Description != "" {
		input.Description = aws.String(config.Description)
	}
	if len(config.ApiStages) > 0 {
		input.ApiStages = config.ApiStages
	}
	if config.Throttle != nil {
		input.Throttle = config.Throttle
	}
	if config.Quota != nil {
		input.Quota = config.Quota
	}
	if config.ProductCode != "" {
		input.ProductCode = aws.String(config.ProductCode)
	}
	if len(config.Tags) > 0 {
		input.Tags = config.Tags
	}

	output, err := client.CreateUsagePlan(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertToUsagePlanResult(output), nil
}

// GetUsagePlan gets an existing usage plan and fails the test if an error occurs
func GetUsagePlan(ctx *TestContext, usagePlanId string) *UsagePlanResult {
	result, err := GetUsagePlanE(ctx, usagePlanId)
	if err != nil {
		ctx.T.Fatalf("Failed to get usage plan: %v", err)
	}
	return result
}

// GetUsagePlanE gets an existing usage plan and returns any error
func GetUsagePlanE(ctx *TestContext, usagePlanId string) (*UsagePlanResult, error) {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &apigateway.GetUsagePlanInput{
		UsagePlanId: aws.String(usagePlanId),
	}

	output, err := client.GetUsagePlan(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return convertGetUsagePlanToResult(output), nil
}

// DeleteUsagePlan deletes an existing usage plan and fails the test if an error occurs
func DeleteUsagePlan(ctx *TestContext, usagePlanId string) {
	err := DeleteUsagePlanE(ctx, usagePlanId)
	if err != nil {
		ctx.T.Fatalf("Failed to delete usage plan: %v", err)
	}
}

// DeleteUsagePlanE deletes an existing usage plan and returns any error
func DeleteUsagePlanE(ctx *TestContext, usagePlanId string) error {
	client, err := createApiGatewayClient(ctx)
	if err != nil {
		return err
	}

	input := &apigateway.DeleteUsagePlanInput{
		UsagePlanId: aws.String(usagePlanId),
	}

	_, err = client.DeleteUsagePlan(context.TODO(), input)
	return err
}

// =============================================================================
// UTILITY AND CONVERSION FUNCTIONS
// =============================================================================

// convertToStageResult converts AWS SDK stage output to our custom type
func convertToStageResult(stage *apigateway.CreateStageOutput) *StageResult {
	return &StageResult{
		DeploymentId:         aws.ToString(stage.DeploymentId),
		StageName:            aws.ToString(stage.StageName),
		Description:          aws.ToString(stage.Description),
		CacheClusterEnabled:  aws.ToBool(stage.CacheClusterEnabled),
		CacheClusterSize:     stage.CacheClusterSize,
		CacheClusterStatus:   stage.CacheClusterStatus,
		MethodSettings:       stage.MethodSettings,
		Variables:            stage.Variables,
		DocumentationVersion: aws.ToString(stage.DocumentationVersion),
		AccessLogSettings:    stage.AccessLogSettings,
		CanarySettings:       stage.CanarySettings,
		TracingConfig:        stage.TracingConfig,
		WebAclArn:            aws.ToString(stage.WebAclArn),
		Tags:                 stage.Tags,
		CreatedDate:          stage.CreatedDate,
		LastUpdatedDate:      stage.LastUpdatedDate,
	}
}

// convertGetStageToStageResult converts GetStage output to StageResult
func convertGetStageToStageResult(stage *apigateway.GetStageOutput) *StageResult {
	return &StageResult{
		DeploymentId:         aws.ToString(stage.DeploymentId),
		StageName:            aws.ToString(stage.StageName),
		Description:          aws.ToString(stage.Description),
		CacheClusterEnabled:  aws.ToBool(stage.CacheClusterEnabled),
		CacheClusterSize:     stage.CacheClusterSize,
		CacheClusterStatus:   stage.CacheClusterStatus,
		MethodSettings:       stage.MethodSettings,
		Variables:            stage.Variables,
		DocumentationVersion: aws.ToString(stage.DocumentationVersion),
		AccessLogSettings:    stage.AccessLogSettings,
		CanarySettings:       stage.CanarySettings,
		TracingConfig:        stage.TracingConfig,
		WebAclArn:            aws.ToString(stage.WebAclArn),
		Tags:                 stage.Tags,
		CreatedDate:          stage.CreatedDate,
		LastUpdatedDate:      stage.LastUpdatedDate,
	}
}

// convertUpdateStageToStageResult converts UpdateStage output to StageResult  
func convertUpdateStageToStageResult(stage *apigateway.UpdateStageOutput) *StageResult {
	return &StageResult{
		DeploymentId:         aws.ToString(stage.DeploymentId),
		StageName:            aws.ToString(stage.StageName),
		Description:          aws.ToString(stage.Description),
		CacheClusterEnabled:  aws.ToBool(stage.CacheClusterEnabled),
		CacheClusterSize:     stage.CacheClusterSize,
		CacheClusterStatus:   stage.CacheClusterStatus,
		MethodSettings:       stage.MethodSettings,
		Variables:            stage.Variables,
		DocumentationVersion: aws.ToString(stage.DocumentationVersion),
		AccessLogSettings:    stage.AccessLogSettings,
		CanarySettings:       stage.CanarySettings,
		TracingConfig:        stage.TracingConfig,
		WebAclArn:            aws.ToString(stage.WebAclArn),
		Tags:                 stage.Tags,
		CreatedDate:          stage.CreatedDate,
		LastUpdatedDate:      stage.LastUpdatedDate,
	}
}

// convertGetMethodToMethodResult converts GetMethod output to MethodResult
func convertGetMethodToMethodResult(method *apigateway.GetMethodOutput) *MethodResult {
	return &MethodResult{
		HttpMethod:             aws.ToString(method.HttpMethod),
		AuthorizationType:      aws.ToString(method.AuthorizationType),
		AuthorizerId:           aws.ToString(method.AuthorizerId),
		ApiKeyRequired:         aws.ToBool(method.ApiKeyRequired),
		RequestValidatorId:     aws.ToString(method.RequestValidatorId),
		OperationName:          aws.ToString(method.OperationName),
		RequestParameters:      method.RequestParameters,
		RequestModels:          method.RequestModels,
		MethodResponses:        method.MethodResponses,
		MethodIntegration:      method.MethodIntegration,
		AuthorizationScopes:    method.AuthorizationScopes,
	}
}

// convertUpdateMethodToMethodResult converts UpdateMethod output to MethodResult
func convertUpdateMethodToMethodResult(method *apigateway.UpdateMethodOutput) *MethodResult {
	return &MethodResult{
		HttpMethod:             aws.ToString(method.HttpMethod),
		AuthorizationType:      aws.ToString(method.AuthorizationType),
		AuthorizerId:           aws.ToString(method.AuthorizerId),
		ApiKeyRequired:         aws.ToBool(method.ApiKeyRequired),
		RequestValidatorId:     aws.ToString(method.RequestValidatorId),
		OperationName:          aws.ToString(method.OperationName),
		RequestParameters:      method.RequestParameters,
		RequestModels:          method.RequestModels,
		MethodResponses:        method.MethodResponses,
		MethodIntegration:      method.MethodIntegration,
		AuthorizationScopes:    method.AuthorizationScopes,
	}
}

// convertToIntegrationResult converts PutIntegration output to IntegrationResult
func convertToIntegrationResult(integration *apigateway.PutIntegrationOutput) *IntegrationResult {
	return &IntegrationResult{
		Type:                    integration.Type,
		HttpMethod:              aws.ToString(integration.HttpMethod),
		Uri:                     aws.ToString(integration.Uri),
		ConnectionType:          integration.ConnectionType,
		ConnectionId:            aws.ToString(integration.ConnectionId),
		Credentials:             aws.ToString(integration.Credentials),
		RequestParameters:       integration.RequestParameters,
		RequestTemplates:        integration.RequestTemplates,
		PassthroughBehavior:     aws.ToString(integration.PassthroughBehavior),
		CacheNamespace:          aws.ToString(integration.CacheNamespace),
		CacheKeyParameters:      integration.CacheKeyParameters,
		ContentHandling:         integration.ContentHandling,
		TimeoutInMillis:         aws.ToInt32(integration.TimeoutInMillis),
		IntegrationResponses:    integration.IntegrationResponses,
		TlsConfig:               integration.TlsConfig,
	}
}

// convertGetIntegrationToResult converts GetIntegration output to IntegrationResult
func convertGetIntegrationToResult(integration *apigateway.GetIntegrationOutput) *IntegrationResult {
	return &IntegrationResult{
		Type:                    integration.Type,
		HttpMethod:              aws.ToString(integration.HttpMethod),
		Uri:                     aws.ToString(integration.Uri),
		ConnectionType:          integration.ConnectionType,
		ConnectionId:            aws.ToString(integration.ConnectionId),
		Credentials:             aws.ToString(integration.Credentials),
		RequestParameters:       integration.RequestParameters,
		RequestTemplates:        integration.RequestTemplates,
		PassthroughBehavior:     aws.ToString(integration.PassthroughBehavior),
		CacheNamespace:          aws.ToString(integration.CacheNamespace),
		CacheKeyParameters:      integration.CacheKeyParameters,
		ContentHandling:         integration.ContentHandling,
		TimeoutInMillis:         aws.ToInt32(integration.TimeoutInMillis),
		IntegrationResponses:    integration.IntegrationResponses,
		TlsConfig:               integration.TlsConfig,
	}
}

// convertGetResourceToResult converts GetResource output to ResourceResult
func convertGetResourceToResult(resource *apigateway.GetResourceOutput) *ResourceResult {
	return &ResourceResult{
		Id:              aws.ToString(resource.Id),
		ParentId:        aws.ToString(resource.ParentId),
		PathPart:        aws.ToString(resource.PathPart),
		Path:            aws.ToString(resource.Path),
		ResourceMethods: resource.ResourceMethods,
	}
}

// convertResourceToResult converts Resource to ResourceResult
func convertResourceToResult(resource *types.Resource) *ResourceResult {
	return &ResourceResult{
		Id:              aws.ToString(resource.Id),
		ParentId:        aws.ToString(resource.ParentId),
		PathPart:        aws.ToString(resource.PathPart),
		Path:            aws.ToString(resource.Path),
		ResourceMethods: resource.ResourceMethods,
	}
}

// convertGetDeploymentToResult converts GetDeployment output to DeploymentResult
func convertGetDeploymentToResult(deployment *apigateway.GetDeploymentOutput) *DeploymentResult {
	return &DeploymentResult{
		Id:              aws.ToString(deployment.Id),
		Description:     aws.ToString(deployment.Description),
		CreatedDate:     deployment.CreatedDate,
		ApiSummary:      deployment.ApiSummary,
	}
}

// convertDeploymentToResult converts Deployment to DeploymentResult
func convertDeploymentToResult(deployment *types.Deployment) *DeploymentResult {
	return &DeploymentResult{
		Id:              aws.ToString(deployment.Id),
		Description:     aws.ToString(deployment.Description),
		CreatedDate:     deployment.CreatedDate,
		ApiSummary:      deployment.ApiSummary,
	}
}

// convertToAPIKeyResult converts CreateApiKey output to APIKeyResult
func convertToAPIKeyResult(apiKey *apigateway.CreateApiKeyOutput) *APIKeyResult {
	return &APIKeyResult{
		Id:              aws.ToString(apiKey.Id),
		Value:           aws.ToString(apiKey.Value),
		Name:            aws.ToString(apiKey.Name),
		CustomerId:      aws.ToString(apiKey.CustomerId),
		Description:     aws.ToString(apiKey.Description),
		Enabled:         aws.ToBool(apiKey.Enabled),
		CreatedDate:     apiKey.CreatedDate,
		LastUpdatedDate: apiKey.LastUpdatedDate,
		StageKeys:       apiKey.StageKeys,
		Tags:            apiKey.Tags,
	}
}

// convertGetAPIKeyToResult converts GetApiKey output to APIKeyResult
func convertGetAPIKeyToResult(apiKey *apigateway.GetApiKeyOutput) *APIKeyResult {
	return &APIKeyResult{
		Id:              aws.ToString(apiKey.Id),
		Value:           aws.ToString(apiKey.Value),
		Name:            aws.ToString(apiKey.Name),
		CustomerId:      aws.ToString(apiKey.CustomerId),
		Description:     aws.ToString(apiKey.Description),
		Enabled:         aws.ToBool(apiKey.Enabled),
		CreatedDate:     apiKey.CreatedDate,
		LastUpdatedDate: apiKey.LastUpdatedDate,
		StageKeys:       apiKey.StageKeys,
		Tags:            apiKey.Tags,
	}
}

// convertUpdateAPIKeyToResult converts UpdateApiKey output to APIKeyResult
func convertUpdateAPIKeyToResult(apiKey *apigateway.UpdateApiKeyOutput) *APIKeyResult {
	return &APIKeyResult{
		Id:              aws.ToString(apiKey.Id),
		Value:           aws.ToString(apiKey.Value),
		Name:            aws.ToString(apiKey.Name),
		CustomerId:      aws.ToString(apiKey.CustomerId),
		Description:     aws.ToString(apiKey.Description),
		Enabled:         aws.ToBool(apiKey.Enabled),
		CreatedDate:     apiKey.CreatedDate,
		LastUpdatedDate: apiKey.LastUpdatedDate,
		StageKeys:       apiKey.StageKeys,
		Tags:            apiKey.Tags,
	}
}

// convertToUsagePlanResult converts CreateUsagePlan output to UsagePlanResult
func convertToUsagePlanResult(usagePlan *apigateway.CreateUsagePlanOutput) *UsagePlanResult {
	return &UsagePlanResult{
		Id:          aws.ToString(usagePlan.Id),
		Name:        aws.ToString(usagePlan.Name),
		Description: aws.ToString(usagePlan.Description),
		ApiStages:   usagePlan.ApiStages,
		Throttle:    usagePlan.Throttle,
		Quota:       usagePlan.Quota,
		ProductCode: aws.ToString(usagePlan.ProductCode),
		Tags:        usagePlan.Tags,
	}
}

// convertGetUsagePlanToResult converts GetUsagePlan output to UsagePlanResult
func convertGetUsagePlanToResult(usagePlan *apigateway.GetUsagePlanOutput) *UsagePlanResult {
	return &UsagePlanResult{
		Id:          aws.ToString(usagePlan.Id),
		Name:        aws.ToString(usagePlan.Name),
		Description: aws.ToString(usagePlan.Description),
		ApiStages:   usagePlan.ApiStages,
		Throttle:    usagePlan.Throttle,
		Quota:       usagePlan.Quota,
		ProductCode: aws.ToString(usagePlan.ProductCode),
		Tags:        usagePlan.Tags,
	}
}