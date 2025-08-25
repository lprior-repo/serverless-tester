package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// LambdaClientInterface defines the interface for Lambda client operations
// This allows dependency injection for testing
type LambdaClientInterface interface {
	GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
	ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
	Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
	CreateFunction(ctx context.Context, params *lambda.CreateFunctionInput, optFns ...func(*lambda.Options)) (*lambda.CreateFunctionOutput, error)
	UpdateFunctionConfiguration(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error)
	UpdateFunctionCode(ctx context.Context, params *lambda.UpdateFunctionCodeInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionCodeOutput, error)
	DeleteFunction(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error)
	PublishLayerVersion(ctx context.Context, params *lambda.PublishLayerVersionInput, optFns ...func(*lambda.Options)) (*lambda.PublishLayerVersionOutput, error)
	GetLayerVersion(ctx context.Context, params *lambda.GetLayerVersionInput, optFns ...func(*lambda.Options)) (*lambda.GetLayerVersionOutput, error)
	ListLayerVersions(ctx context.Context, params *lambda.ListLayerVersionsInput, optFns ...func(*lambda.Options)) (*lambda.ListLayerVersionsOutput, error)
	DeleteLayerVersion(ctx context.Context, params *lambda.DeleteLayerVersionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteLayerVersionOutput, error)
	PublishVersion(ctx context.Context, params *lambda.PublishVersionInput, optFns ...func(*lambda.Options)) (*lambda.PublishVersionOutput, error)
	CreateAlias(ctx context.Context, params *lambda.CreateAliasInput, optFns ...func(*lambda.Options)) (*lambda.CreateAliasOutput, error)
	GetAlias(ctx context.Context, params *lambda.GetAliasInput, optFns ...func(*lambda.Options)) (*lambda.GetAliasOutput, error)
	UpdateAlias(ctx context.Context, params *lambda.UpdateAliasInput, optFns ...func(*lambda.Options)) (*lambda.UpdateAliasOutput, error)
	DeleteAlias(ctx context.Context, params *lambda.DeleteAliasInput, optFns ...func(*lambda.Options)) (*lambda.DeleteAliasOutput, error)
	CreateEventSourceMapping(ctx context.Context, params *lambda.CreateEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.CreateEventSourceMappingOutput, error)
	GetEventSourceMapping(ctx context.Context, params *lambda.GetEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.GetEventSourceMappingOutput, error)
	ListEventSourceMappings(ctx context.Context, params *lambda.ListEventSourceMappingsInput, optFns ...func(*lambda.Options)) (*lambda.ListEventSourceMappingsOutput, error)
	UpdateEventSourceMapping(ctx context.Context, params *lambda.UpdateEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.UpdateEventSourceMappingOutput, error)
	DeleteEventSourceMapping(ctx context.Context, params *lambda.DeleteEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.DeleteEventSourceMappingOutput, error)
}

// CloudWatchLogsClientInterface defines the interface for CloudWatch Logs client operations
type CloudWatchLogsClientInterface interface {
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	DescribeLogStreams(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error)
	FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error)
}

// MockLambdaClient provides a mock implementation for testing
type MockLambdaClient struct {
	GetFunctionResponse                    *lambda.GetFunctionOutput
	GetFunctionError                       error
	ListFunctionsResponse                  *lambda.ListFunctionsOutput
	ListFunctionsError                     error
	InvokeResponse                         *lambda.InvokeOutput
	InvokeError                            error
	CreateFunctionResponse                 *lambda.CreateFunctionOutput
	CreateFunctionError                    error
	UpdateFunctionConfigurationResponse    *lambda.UpdateFunctionConfigurationOutput
	UpdateFunctionConfigurationError       error
	UpdateFunctionCodeResponse             *lambda.UpdateFunctionCodeOutput
	UpdateFunctionCodeError                error
	DeleteFunctionResponse                 *lambda.DeleteFunctionOutput
	DeleteFunctionError                    error
	
	// Layer management responses
	PublishLayerVersionResponse            *lambda.PublishLayerVersionOutput
	PublishLayerVersionError               error
	GetLayerVersionResponse                *lambda.GetLayerVersionOutput
	GetLayerVersionError                   error
	ListLayerVersionsResponse              *lambda.ListLayerVersionsOutput
	ListLayerVersionsError                 error
	DeleteLayerVersionResponse             *lambda.DeleteLayerVersionOutput
	DeleteLayerVersionError                error
	
	// Version and Alias management responses
	PublishVersionResponse                 *lambda.PublishVersionOutput
	PublishVersionError                    error
	CreateAliasResponse                    *lambda.CreateAliasOutput
	CreateAliasError                       error
	GetAliasResponse                       *lambda.GetAliasOutput
	GetAliasError                          error
	UpdateAliasResponse                    *lambda.UpdateAliasOutput
	UpdateAliasError                       error
	DeleteAliasResponse                    *lambda.DeleteAliasOutput
	DeleteAliasError                       error
	
	// Event Source Mapping responses
	CreateEventSourceMappingResponse       *lambda.CreateEventSourceMappingOutput
	CreateEventSourceMappingError          error
	GetEventSourceMappingResponse          *lambda.GetEventSourceMappingOutput
	GetEventSourceMappingError             error
	GetEventSourceMappingResponses         []*lambda.GetEventSourceMappingOutput // For sequential responses
	GetEventSourceMappingResponseIndex     int                                    // Current response index
	ListEventSourceMappingsResponse        *lambda.ListEventSourceMappingsOutput
	ListEventSourceMappingsError           error
	UpdateEventSourceMappingResponse       *lambda.UpdateEventSourceMappingOutput
	UpdateEventSourceMappingError          error
	DeleteEventSourceMappingResponse       *lambda.DeleteEventSourceMappingOutput
	DeleteEventSourceMappingError          error
	
	// Call tracking for verification
	GetFunctionCalls                       []string
	ListFunctionsCalls                     int
	InvokeCalls                            []string
	CreateFunctionCalls                    []string
	UpdateFunctionConfigurationCalls       []string
	UpdateFunctionCodeCalls                []string
	DeleteFunctionCalls                    []string
	
	// Layer management call tracking
	PublishLayerVersionCalls               []string
	GetLayerVersionCalls                   []MockLayerVersionCall
	ListLayerVersionsCalls                 []string
	DeleteLayerVersionCalls                []MockLayerVersionCall
	
	// Version and Alias management call tracking
	PublishVersionCalls                    []string
	CreateAliasCalls                       []MockAliasCall
	GetAliasCalls                          []MockAliasCall
	UpdateAliasCalls                       []MockAliasCall
	DeleteAliasCalls                       []MockAliasCall
	
	// Event Source Mapping call tracking
	CreateEventSourceMappingCalls          []MockEventSourceMappingCall
	GetEventSourceMappingCalls             []string
	ListEventSourceMappingsCalls           []string
	UpdateEventSourceMappingCalls          []MockUpdateEventSourceMappingCall
	DeleteEventSourceMappingCalls          []string
}

// MockEventSourceMappingCall tracks parameters for CreateEventSourceMapping calls
type MockEventSourceMappingCall struct {
	FunctionName     string
	EventSourceArn   string
	BatchSize        *int32
	Enabled          *bool
	StartingPosition string
}

// MockUpdateEventSourceMappingCall tracks parameters for UpdateEventSourceMapping calls
type MockUpdateEventSourceMappingCall struct {
	UUID         string
	BatchSize    *int32
	Enabled      *bool
	FunctionName *string
}

// MockLayerVersionCall tracks parameters for layer version calls
type MockLayerVersionCall struct {
	LayerName      string
	VersionNumber  int64
}

// MockAliasCall tracks parameters for alias calls
type MockAliasCall struct {
	FunctionName string
	AliasName    string
}

func (m *MockLambdaClient) GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error) {
	if params != nil && params.FunctionName != nil {
		m.GetFunctionCalls = append(m.GetFunctionCalls, *params.FunctionName)
	}
	return m.GetFunctionResponse, m.GetFunctionError
}

func (m *MockLambdaClient) ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
	m.ListFunctionsCalls++
	return m.ListFunctionsResponse, m.ListFunctionsError
}

func (m *MockLambdaClient) Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	if params != nil && params.FunctionName != nil {
		m.InvokeCalls = append(m.InvokeCalls, *params.FunctionName)
	}
	return m.InvokeResponse, m.InvokeError
}

func (m *MockLambdaClient) CreateFunction(ctx context.Context, params *lambda.CreateFunctionInput, optFns ...func(*lambda.Options)) (*lambda.CreateFunctionOutput, error) {
	if params != nil && params.FunctionName != nil {
		m.CreateFunctionCalls = append(m.CreateFunctionCalls, *params.FunctionName)
	}
	return m.CreateFunctionResponse, m.CreateFunctionError
}

func (m *MockLambdaClient) UpdateFunctionConfiguration(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error) {
	if params != nil && params.FunctionName != nil {
		m.UpdateFunctionConfigurationCalls = append(m.UpdateFunctionConfigurationCalls, *params.FunctionName)
	}
	return m.UpdateFunctionConfigurationResponse, m.UpdateFunctionConfigurationError
}

func (m *MockLambdaClient) UpdateFunctionCode(ctx context.Context, params *lambda.UpdateFunctionCodeInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionCodeOutput, error) {
	if params != nil && params.FunctionName != nil {
		m.UpdateFunctionCodeCalls = append(m.UpdateFunctionCodeCalls, *params.FunctionName)
	}
	return m.UpdateFunctionCodeResponse, m.UpdateFunctionCodeError
}

func (m *MockLambdaClient) DeleteFunction(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error) {
	if params != nil && params.FunctionName != nil {
		m.DeleteFunctionCalls = append(m.DeleteFunctionCalls, *params.FunctionName)
	}
	return m.DeleteFunctionResponse, m.DeleteFunctionError
}

func (m *MockLambdaClient) PublishLayerVersion(ctx context.Context, params *lambda.PublishLayerVersionInput, optFns ...func(*lambda.Options)) (*lambda.PublishLayerVersionOutput, error) {
	if params != nil && params.LayerName != nil {
		m.PublishLayerVersionCalls = append(m.PublishLayerVersionCalls, *params.LayerName)
	}
	return m.PublishLayerVersionResponse, m.PublishLayerVersionError
}

func (m *MockLambdaClient) GetLayerVersion(ctx context.Context, params *lambda.GetLayerVersionInput, optFns ...func(*lambda.Options)) (*lambda.GetLayerVersionOutput, error) {
	if params != nil {
		call := MockLayerVersionCall{}
		if params.LayerName != nil {
			call.LayerName = *params.LayerName
		}
		if params.VersionNumber != nil {
			call.VersionNumber = *params.VersionNumber
		}
		m.GetLayerVersionCalls = append(m.GetLayerVersionCalls, call)
	}
	return m.GetLayerVersionResponse, m.GetLayerVersionError
}

func (m *MockLambdaClient) ListLayerVersions(ctx context.Context, params *lambda.ListLayerVersionsInput, optFns ...func(*lambda.Options)) (*lambda.ListLayerVersionsOutput, error) {
	if params != nil && params.LayerName != nil {
		m.ListLayerVersionsCalls = append(m.ListLayerVersionsCalls, *params.LayerName)
	}
	return m.ListLayerVersionsResponse, m.ListLayerVersionsError
}

func (m *MockLambdaClient) DeleteLayerVersion(ctx context.Context, params *lambda.DeleteLayerVersionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteLayerVersionOutput, error) {
	if params != nil {
		call := MockLayerVersionCall{}
		if params.LayerName != nil {
			call.LayerName = *params.LayerName
		}
		if params.VersionNumber != nil {
			call.VersionNumber = *params.VersionNumber
		}
		m.DeleteLayerVersionCalls = append(m.DeleteLayerVersionCalls, call)
	}
	return m.DeleteLayerVersionResponse, m.DeleteLayerVersionError
}



func (m *MockLambdaClient) CreateEventSourceMapping(ctx context.Context, params *lambda.CreateEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.CreateEventSourceMappingOutput, error) {
	if params != nil {
		call := MockEventSourceMappingCall{}
		if params.FunctionName != nil {
			call.FunctionName = *params.FunctionName
		}
		if params.EventSourceArn != nil {
			call.EventSourceArn = *params.EventSourceArn
		}
		call.BatchSize = params.BatchSize
		call.Enabled = params.Enabled
		if params.StartingPosition != "" {
			call.StartingPosition = string(params.StartingPosition)
		}
		m.CreateEventSourceMappingCalls = append(m.CreateEventSourceMappingCalls, call)
	}
	return m.CreateEventSourceMappingResponse, m.CreateEventSourceMappingError
}

func (m *MockLambdaClient) GetEventSourceMapping(ctx context.Context, params *lambda.GetEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.GetEventSourceMappingOutput, error) {
	if params != nil && params.UUID != nil {
		m.GetEventSourceMappingCalls = append(m.GetEventSourceMappingCalls, *params.UUID)
	}
	
	// If we have sequential responses, use those instead of the single response
	if len(m.GetEventSourceMappingResponses) > 0 {
		if m.GetEventSourceMappingResponseIndex < len(m.GetEventSourceMappingResponses) {
			response := m.GetEventSourceMappingResponses[m.GetEventSourceMappingResponseIndex]
			m.GetEventSourceMappingResponseIndex++
			return response, m.GetEventSourceMappingError
		}
		// If we've exhausted sequential responses, return the last one
		return m.GetEventSourceMappingResponses[len(m.GetEventSourceMappingResponses)-1], m.GetEventSourceMappingError
	}
	
	return m.GetEventSourceMappingResponse, m.GetEventSourceMappingError
}

func (m *MockLambdaClient) ListEventSourceMappings(ctx context.Context, params *lambda.ListEventSourceMappingsInput, optFns ...func(*lambda.Options)) (*lambda.ListEventSourceMappingsOutput, error) {
	if params != nil && params.FunctionName != nil {
		m.ListEventSourceMappingsCalls = append(m.ListEventSourceMappingsCalls, *params.FunctionName)
	}
	return m.ListEventSourceMappingsResponse, m.ListEventSourceMappingsError
}

func (m *MockLambdaClient) UpdateEventSourceMapping(ctx context.Context, params *lambda.UpdateEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.UpdateEventSourceMappingOutput, error) {
	if params != nil {
		call := MockUpdateEventSourceMappingCall{}
		if params.UUID != nil {
			call.UUID = *params.UUID
		}
		call.BatchSize = params.BatchSize
		call.Enabled = params.Enabled
		call.FunctionName = params.FunctionName
		m.UpdateEventSourceMappingCalls = append(m.UpdateEventSourceMappingCalls, call)
	}
	return m.UpdateEventSourceMappingResponse, m.UpdateEventSourceMappingError
}

func (m *MockLambdaClient) DeleteEventSourceMapping(ctx context.Context, params *lambda.DeleteEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.DeleteEventSourceMappingOutput, error) {
	if params != nil && params.UUID != nil {
		m.DeleteEventSourceMappingCalls = append(m.DeleteEventSourceMappingCalls, *params.UUID)
	}
	return m.DeleteEventSourceMappingResponse, m.DeleteEventSourceMappingError
}

// Stub implementations for interface compliance - not used in event source mapping tests
func (m *MockLambdaClient) PublishVersion(ctx context.Context, params *lambda.PublishVersionInput, optFns ...func(*lambda.Options)) (*lambda.PublishVersionOutput, error) {
	return nil, nil
}

func (m *MockLambdaClient) CreateAlias(ctx context.Context, params *lambda.CreateAliasInput, optFns ...func(*lambda.Options)) (*lambda.CreateAliasOutput, error) {
	return nil, nil
}

func (m *MockLambdaClient) GetAlias(ctx context.Context, params *lambda.GetAliasInput, optFns ...func(*lambda.Options)) (*lambda.GetAliasOutput, error) {
	return nil, nil
}

func (m *MockLambdaClient) UpdateAlias(ctx context.Context, params *lambda.UpdateAliasInput, optFns ...func(*lambda.Options)) (*lambda.UpdateAliasOutput, error) {
	return nil, nil
}

func (m *MockLambdaClient) DeleteAlias(ctx context.Context, params *lambda.DeleteAliasInput, optFns ...func(*lambda.Options)) (*lambda.DeleteAliasOutput, error) {
	return nil, nil
}

func (m *MockLambdaClient) ListAliases(ctx context.Context, params *lambda.ListAliasesInput, optFns ...func(*lambda.Options)) (*lambda.ListAliasesOutput, error) {
	return nil, nil
}


// Reset clears all call tracking and responses for reuse in tests
func (m *MockLambdaClient) Reset() {
	m.GetFunctionResponse = nil
	m.GetFunctionError = nil
	m.ListFunctionsResponse = nil
	m.ListFunctionsError = nil
	m.InvokeResponse = nil
	m.InvokeError = nil
	m.CreateFunctionResponse = nil
	m.CreateFunctionError = nil
	m.UpdateFunctionConfigurationResponse = nil
	m.UpdateFunctionConfigurationError = nil
	m.UpdateFunctionCodeResponse = nil
	m.UpdateFunctionCodeError = nil
	m.DeleteFunctionResponse = nil
	m.DeleteFunctionError = nil
	
	// Reset Layer management responses
	m.PublishLayerVersionResponse = nil
	m.PublishLayerVersionError = nil
	m.GetLayerVersionResponse = nil
	m.GetLayerVersionError = nil
	m.ListLayerVersionsResponse = nil
	m.ListLayerVersionsError = nil
	m.DeleteLayerVersionResponse = nil
	m.DeleteLayerVersionError = nil
	
	// Reset Event Source Mapping responses
	m.CreateEventSourceMappingResponse = nil
	m.CreateEventSourceMappingError = nil
	m.GetEventSourceMappingResponse = nil
	m.GetEventSourceMappingError = nil
	m.GetEventSourceMappingResponses = nil
	m.GetEventSourceMappingResponseIndex = 0
	m.ListEventSourceMappingsResponse = nil
	m.ListEventSourceMappingsError = nil
	m.UpdateEventSourceMappingResponse = nil
	m.UpdateEventSourceMappingError = nil
	m.DeleteEventSourceMappingResponse = nil
	m.DeleteEventSourceMappingError = nil
	
	m.GetFunctionCalls = nil
	m.ListFunctionsCalls = 0
	m.InvokeCalls = nil
	m.CreateFunctionCalls = nil
	m.UpdateFunctionConfigurationCalls = nil
	m.UpdateFunctionCodeCalls = nil
	m.DeleteFunctionCalls = nil
	
	// Reset Layer management call tracking
	m.PublishLayerVersionCalls = nil
	m.GetLayerVersionCalls = nil
	m.ListLayerVersionsCalls = nil
	m.DeleteLayerVersionCalls = nil
	
	// Reset Event Source Mapping call tracking
	m.CreateEventSourceMappingCalls = nil
	m.GetEventSourceMappingCalls = nil
	m.ListEventSourceMappingsCalls = nil
	m.UpdateEventSourceMappingCalls = nil
	m.DeleteEventSourceMappingCalls = nil
}

// MockCloudWatchLogsClient provides a mock implementation for testing
type MockCloudWatchLogsClient struct {
	DescribeLogGroupsResponse   *cloudwatchlogs.DescribeLogGroupsOutput
	DescribeLogGroupsError      error
	DescribeLogStreamsResponse  *cloudwatchlogs.DescribeLogStreamsOutput
	DescribeLogStreamsError     error
	FilterLogEventsResponse     *cloudwatchlogs.FilterLogEventsOutput
	FilterLogEventsError        error
	
	// Call tracking for verification
	DescribeLogGroupsCalls      int
	DescribeLogStreamsCalls     int
	FilterLogEventsCalls        []string
}

func (m *MockCloudWatchLogsClient) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	m.DescribeLogGroupsCalls++
	return m.DescribeLogGroupsResponse, m.DescribeLogGroupsError
}

func (m *MockCloudWatchLogsClient) DescribeLogStreams(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
	m.DescribeLogStreamsCalls++
	return m.DescribeLogStreamsResponse, m.DescribeLogStreamsError
}

func (m *MockCloudWatchLogsClient) FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
	if params != nil && params.LogGroupName != nil {
		m.FilterLogEventsCalls = append(m.FilterLogEventsCalls, *params.LogGroupName)
	}
	return m.FilterLogEventsResponse, m.FilterLogEventsError
}

// Reset clears all call tracking and responses for reuse in tests
func (m *MockCloudWatchLogsClient) Reset() {
	m.DescribeLogGroupsResponse = nil
	m.DescribeLogGroupsError = nil
	m.DescribeLogStreamsResponse = nil
	m.DescribeLogStreamsError = nil
	m.FilterLogEventsResponse = nil
	m.FilterLogEventsError = nil
	
	m.DescribeLogGroupsCalls = 0
	m.DescribeLogStreamsCalls = 0
	m.FilterLogEventsCalls = nil
}

// ClientFactory provides dependency injection for creating AWS clients
type ClientFactory interface {
	CreateLambdaClient(ctx *TestContext) LambdaClientInterface
	CreateCloudWatchLogsClient(ctx *TestContext) CloudWatchLogsClientInterface
}

// DefaultClientFactory creates real AWS clients
type DefaultClientFactory struct{}

func (f *DefaultClientFactory) CreateLambdaClient(ctx *TestContext) LambdaClientInterface {
	return lambda.NewFromConfig(ctx.AwsConfig)
}

func (f *DefaultClientFactory) CreateCloudWatchLogsClient(ctx *TestContext) CloudWatchLogsClientInterface {
	return cloudwatchlogs.NewFromConfig(ctx.AwsConfig)
}

// MockClientFactory creates mock clients for testing
type MockClientFactory struct {
	LambdaClient       *MockLambdaClient
	CloudWatchClient   *MockCloudWatchLogsClient
}

func (f *MockClientFactory) CreateLambdaClient(ctx *TestContext) LambdaClientInterface {
	return f.LambdaClient
}

func (f *MockClientFactory) CreateCloudWatchLogsClient(ctx *TestContext) CloudWatchLogsClientInterface {
	return f.CloudWatchClient
}

// Global client factory - can be overridden for testing
var globalClientFactory ClientFactory = &DefaultClientFactory{}

// SetClientFactory allows overriding the client factory for testing
func SetClientFactory(factory ClientFactory) {
	globalClientFactory = factory
}

// createResourceNotFoundError creates a mock ResourceNotFoundException error
func createResourceNotFoundError(functionName string) error {
	return fmt.Errorf("ResourceNotFoundException: Function not found: arn:aws:lambda:us-east-1:123456789012:function:%s", functionName)
}

// createAccessDeniedError creates a mock AccessDeniedException error
func createAccessDeniedError(operation string) error {
	return fmt.Errorf("AccessDeniedException: User is not authorized to perform: lambda:%s", operation)
}