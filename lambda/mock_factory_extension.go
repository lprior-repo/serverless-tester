package lambda

import "github.com/stretchr/testify/mock"

// TestifyMockClientFactory extends MockClientFactory to work with testify/mock
type TestifyMockClientFactory struct {
	mock.Mock
}

func (f *TestifyMockClientFactory) CreateLambdaClient(ctx *TestContext) LambdaClientInterface {
	args := f.Called(ctx)
	return args.Get(0).(LambdaClientInterface)
}

func (f *TestifyMockClientFactory) CreateCloudWatchLogsClient(ctx *TestContext) CloudWatchLogsClientInterface {
	args := f.Called(ctx)
	if args.Get(0) == nil {
		return &MockCloudWatchLogsClient{}
	}
	return args.Get(0).(CloudWatchLogsClientInterface)
}