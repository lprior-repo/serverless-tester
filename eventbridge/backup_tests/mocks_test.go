package eventbridge

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/stretchr/testify/mock"
)

// MockEventBridgeClient provides a mock implementation for testing EventBridge operations
type MockEventBridgeClient struct {
	mock.Mock
}

// PutEvents mocks the EventBridge PutEvents operation
func (m *MockEventBridgeClient) PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.PutEventsOutput), args.Error(1)
}

// PutRule mocks the EventBridge PutRule operation
func (m *MockEventBridgeClient) PutRule(ctx context.Context, params *eventbridge.PutRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutRuleOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.PutRuleOutput), args.Error(1)
}

// DeleteRule mocks the EventBridge DeleteRule operation
func (m *MockEventBridgeClient) DeleteRule(ctx context.Context, params *eventbridge.DeleteRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteRuleOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DeleteRuleOutput), args.Error(1)
}

// DescribeRule mocks the EventBridge DescribeRule operation
func (m *MockEventBridgeClient) DescribeRule(ctx context.Context, params *eventbridge.DescribeRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeRuleOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DescribeRuleOutput), args.Error(1)
}

// EnableRule mocks the EventBridge EnableRule operation
func (m *MockEventBridgeClient) EnableRule(ctx context.Context, params *eventbridge.EnableRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.EnableRuleOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.EnableRuleOutput), args.Error(1)
}

// DisableRule mocks the EventBridge DisableRule operation
func (m *MockEventBridgeClient) DisableRule(ctx context.Context, params *eventbridge.DisableRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DisableRuleOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DisableRuleOutput), args.Error(1)
}

// ListRules mocks the EventBridge ListRules operation
func (m *MockEventBridgeClient) ListRules(ctx context.Context, params *eventbridge.ListRulesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListRulesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.ListRulesOutput), args.Error(1)
}

// PutTargets mocks the EventBridge PutTargets operation
func (m *MockEventBridgeClient) PutTargets(ctx context.Context, params *eventbridge.PutTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutTargetsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.PutTargetsOutput), args.Error(1)
}

// RemoveTargets mocks the EventBridge RemoveTargets operation
func (m *MockEventBridgeClient) RemoveTargets(ctx context.Context, params *eventbridge.RemoveTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.RemoveTargetsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.RemoveTargetsOutput), args.Error(1)
}

// ListTargetsByRule mocks the EventBridge ListTargetsByRule operation
func (m *MockEventBridgeClient) ListTargetsByRule(ctx context.Context, params *eventbridge.ListTargetsByRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListTargetsByRuleOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.ListTargetsByRuleOutput), args.Error(1)
}

// CreateEventBus mocks the EventBridge CreateEventBus operation
func (m *MockEventBridgeClient) CreateEventBus(ctx context.Context, params *eventbridge.CreateEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.CreateEventBusOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.CreateEventBusOutput), args.Error(1)
}

// DeleteEventBus mocks the EventBridge DeleteEventBus operation
func (m *MockEventBridgeClient) DeleteEventBus(ctx context.Context, params *eventbridge.DeleteEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteEventBusOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DeleteEventBusOutput), args.Error(1)
}

// DescribeEventBus mocks the EventBridge DescribeEventBus operation
func (m *MockEventBridgeClient) DescribeEventBus(ctx context.Context, params *eventbridge.DescribeEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeEventBusOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DescribeEventBusOutput), args.Error(1)
}

// ListEventBuses mocks the EventBridge ListEventBuses operation
func (m *MockEventBridgeClient) ListEventBuses(ctx context.Context, params *eventbridge.ListEventBusesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListEventBusesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.ListEventBusesOutput), args.Error(1)
}

// CreateArchive mocks the EventBridge CreateArchive operation
func (m *MockEventBridgeClient) CreateArchive(ctx context.Context, params *eventbridge.CreateArchiveInput, optFns ...func(*eventbridge.Options)) (*eventbridge.CreateArchiveOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.CreateArchiveOutput), args.Error(1)
}

// DeleteArchive mocks the EventBridge DeleteArchive operation
func (m *MockEventBridgeClient) DeleteArchive(ctx context.Context, params *eventbridge.DeleteArchiveInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteArchiveOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DeleteArchiveOutput), args.Error(1)
}

// DescribeArchive mocks the EventBridge DescribeArchive operation
func (m *MockEventBridgeClient) DescribeArchive(ctx context.Context, params *eventbridge.DescribeArchiveInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeArchiveOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DescribeArchiveOutput), args.Error(1)
}

// ListArchives mocks the EventBridge ListArchives operation
func (m *MockEventBridgeClient) ListArchives(ctx context.Context, params *eventbridge.ListArchivesInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListArchivesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.ListArchivesOutput), args.Error(1)
}

// StartReplay mocks the EventBridge StartReplay operation
func (m *MockEventBridgeClient) StartReplay(ctx context.Context, params *eventbridge.StartReplayInput, optFns ...func(*eventbridge.Options)) (*eventbridge.StartReplayOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.StartReplayOutput), args.Error(1)
}

// DescribeReplay mocks the EventBridge DescribeReplay operation
func (m *MockEventBridgeClient) DescribeReplay(ctx context.Context, params *eventbridge.DescribeReplayInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeReplayOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.DescribeReplayOutput), args.Error(1)
}

// CancelReplay mocks the EventBridge CancelReplay operation
func (m *MockEventBridgeClient) CancelReplay(ctx context.Context, params *eventbridge.CancelReplayInput, optFns ...func(*eventbridge.Options)) (*eventbridge.CancelReplayOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.CancelReplayOutput), args.Error(1)
}

// PutPermission mocks the EventBridge PutPermission operation
func (m *MockEventBridgeClient) PutPermission(ctx context.Context, params *eventbridge.PutPermissionInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutPermissionOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.PutPermissionOutput), args.Error(1)
}

// RemovePermission mocks the EventBridge RemovePermission operation
func (m *MockEventBridgeClient) RemovePermission(ctx context.Context, params *eventbridge.RemovePermissionInput, optFns ...func(*eventbridge.Options)) (*eventbridge.RemovePermissionOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.RemovePermissionOutput), args.Error(1)
}

// TestEventPattern mocks the EventBridge TestEventPattern operation
func (m *MockEventBridgeClient) TestEventPattern(ctx context.Context, params *eventbridge.TestEventPatternInput, optFns ...func(*eventbridge.Options)) (*eventbridge.TestEventPatternOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventbridge.TestEventPatternOutput), args.Error(1)
}

// AnyTime returns a mock matcher for any time pointer
func (m *MockEventBridgeClient) AnyTime() interface{} {
	return mock.AnythingOfType("*time.Time")
}

// MockTestingT provides a mock implementation of TestingT interface for testing
type MockTestingT struct {
	mock.Mock
	ErrorfCalled  bool
	FailNowCalled bool
	ErrorCalled   bool
	FatalCalled   bool
}

// Errorf mocks the Errorf method
func (m *MockTestingT) Errorf(format string, args ...interface{}) {
	m.ErrorfCalled = true
	m.Called(format, args)
}

// FailNow mocks the FailNow method
func (m *MockTestingT) FailNow() {
	m.FailNowCalled = true
	m.Called()
}

// Error mocks the Error method
func (m *MockTestingT) Error(args ...interface{}) {
	m.ErrorCalled = true
	m.Called(args)
}

// Fail mocks the Fail method
func (m *MockTestingT) Fail() {
	m.Called()
}

// Helper mocks the Helper method
func (m *MockTestingT) Helper() {
	m.Called()
}

// Fatal mocks the Fatal method
func (m *MockTestingT) Fatal(args ...interface{}) {
	m.FatalCalled = true
	m.Called(args)
}

// Fatalf mocks the Fatalf method
func (m *MockTestingT) Fatalf(format string, args ...interface{}) {
	m.FatalCalled = true
	m.Called(format, args)
}

// Name mocks the Name method
func (m *MockTestingT) Name() string {
	args := m.Called()
	return args.String(0)
}