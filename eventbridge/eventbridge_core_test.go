package eventbridge

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// RED: Test defaultRetryConfig returns proper default values
func TestDefaultRetryConfig_ReturnsProperDefaults(t *testing.T) {
	config := defaultRetryConfig()

	assert.Equal(t, DefaultRetryAttempts, config.MaxAttempts)
	assert.Equal(t, DefaultRetryDelay, config.BaseDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

// RED: Test createEventBridgeClient with mock client
func TestCreateEventBridgeClient_WithMockClient_ReturnsMockClient(t *testing.T) {
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		Client: mockClient,
	}

	client := createEventBridgeClient(ctx)

	assert.Equal(t, mockClient, client)
}

// RED: Test createEventBridgeClient without mock client creates real client
func TestCreateEventBridgeClient_WithoutMockClient_CreatesRealClient(t *testing.T) {
	ctx := &TestContext{
		AwsConfig: aws.Config{
			Region: "us-east-1",
		},
		Client: nil,
	}

	client := createEventBridgeClient(ctx)

	assert.NotNil(t, client)
	// Real client should be of type *eventbridge.Client
	_, ok := client.(*eventbridge.Client)
	assert.True(t, ok)
}

// RED: Test validateEventDetail with valid JSON
func TestValidateEventDetail_WithValidJSON_ReturnsNil(t *testing.T) {
	validJSON := `{"key": "value", "number": 123}`

	err := validateEventDetail(validJSON)

	assert.NoError(t, err)
}

// RED: Test validateEventDetail with empty detail
func TestValidateEventDetail_WithEmptyDetail_ReturnsNil(t *testing.T) {
	err := validateEventDetail("")

	assert.NoError(t, err)
}

// RED: Test validateEventDetail with invalid JSON
func TestValidateEventDetail_WithInvalidJSON_ReturnsError(t *testing.T) {
	invalidJSON := `{"key": "value", "invalid"`

	err := validateEventDetail(invalidJSON)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event detail")
}

// RED: Test validateEventDetail with oversized detail
func TestValidateEventDetail_WithOversizedDetail_ReturnsError(t *testing.T) {
	// Create a detail larger than MaxEventDetailSize (256KB)
	largeDetail := `{"data": "` + string(make([]byte, MaxEventDetailSize+1)) + `"}`

	err := validateEventDetail(largeDetail)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "detail size exceeds maximum")
}

// RED: Test calculateBackoffDelay with attempt 0
func TestCalculateBackoffDelay_WithAttemptZero_ReturnsBaseDelay(t *testing.T) {
	config := RetryConfig{
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}

	delay := calculateBackoffDelay(0, config)

	assert.Equal(t, config.BaseDelay, delay)
}

// RED: Test calculateBackoffDelay with negative attempt
func TestCalculateBackoffDelay_WithNegativeAttempt_ReturnsBaseDelay(t *testing.T) {
	config := RetryConfig{
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}

	delay := calculateBackoffDelay(-1, config)

	assert.Equal(t, config.BaseDelay, delay)
}

// RED: Test calculateBackoffDelay with exponential growth
func TestCalculateBackoffDelay_WithValidAttempts_ReturnsExponentialDelay(t *testing.T) {
	config := RetryConfig{
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 2 * time.Second},  // BaseDelay * (Multiplier ^ 1)
		{2, 4 * time.Second},  // BaseDelay * (Multiplier ^ 2)
		{3, 8 * time.Second},  // BaseDelay * (Multiplier ^ 3)
	}

	for _, test := range tests {
		delay := calculateBackoffDelay(test.attempt, config)
		assert.Equal(t, test.expected, delay)
	}
}

// RED: Test calculateBackoffDelay caps at max delay
func TestCalculateBackoffDelay_WithLargeAttempt_CapsAtMaxDelay(t *testing.T) {
	config := RetryConfig{
		BaseDelay:   time.Second,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
	}

	// Large attempt should be capped at MaxDelay
	delay := calculateBackoffDelay(10, config)

	assert.Equal(t, config.MaxDelay, delay)
}

// RED: Test calculateBackoffDelay with overflow protection
func TestCalculateBackoffDelay_WithOverflowCondition_CapsAtMaxDelay(t *testing.T) {
	config := RetryConfig{
		BaseDelay:   time.Hour,
		MaxDelay:    2 * time.Hour,
		Multiplier:  1000.0, // Large multiplier to cause overflow
	}

	delay := calculateBackoffDelay(5, config)

	assert.Equal(t, config.MaxDelay, delay)
}

// Note: safeString is not exported, so we don't test it directly
// Using existing MockEventBridgeClient from mocks_test.go

// TestingTMock for testing
type TestingTMock struct {
	mock.Mock
}

func (m *TestingTMock) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *TestingTMock) Error(args ...interface{}) {
	m.Called(args)
}

func (m *TestingTMock) Fail() {
	m.Called()
}

func (m *TestingTMock) FailNow() {
	m.Called()
}

func (m *TestingTMock) Helper() {
	m.Called()
}

func (m *TestingTMock) Fatal(args ...interface{}) {
	m.Called(args)
}

func (m *TestingTMock) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *TestingTMock) Name() string {
	args := m.Called()
	return args.String(0)
}

// RED: Test constants have expected values
func TestConstants_HaveExpectedValues(t *testing.T) {
	assert.Equal(t, 30*time.Second, DefaultTimeout)
	assert.Equal(t, 3, DefaultRetryAttempts)
	assert.Equal(t, 1*time.Second, DefaultRetryDelay)
	assert.Equal(t, 10, MaxEventBatchSize)
	assert.Equal(t, 256*1024, MaxEventDetailSize)
	assert.Equal(t, "default", DefaultEventBusName)
}

// RED: Test rule state constants
func TestRuleStateConstants_MatchAWSTypes(t *testing.T) {
	assert.Equal(t, types.RuleStateEnabled, RuleStateEnabled)
	assert.Equal(t, types.RuleStateDisabled, RuleStateDisabled)
}

// RED: Test archive state constants
func TestArchiveStateConstants_MatchAWSTypes(t *testing.T) {
	assert.Equal(t, types.ArchiveStateEnabled, ArchiveStateEnabled)
	assert.Equal(t, types.ArchiveStateDisabled, ArchiveStateDisabled)
}

// RED: Test replay state constants
func TestReplayStateConstants_MatchAWSTypes(t *testing.T) {
	assert.Equal(t, types.ReplayStateStarting, ReplayStateStarting)
	assert.Equal(t, types.ReplayStateRunning, ReplayStateRunning)
	assert.Equal(t, types.ReplayStateCompleted, ReplayStateCompleted)
	assert.Equal(t, types.ReplayStateFailed, ReplayStateFailed)
}

// RED: Test error variables are defined
func TestErrorVariables_AreDefined(t *testing.T) {
	assert.NotNil(t, ErrEventNotFound)
	assert.NotNil(t, ErrRuleNotFound)
	assert.NotNil(t, ErrEventBusNotFound)
	assert.NotNil(t, ErrInvalidEventPattern)
	assert.NotNil(t, ErrInvalidEventDetail)
	assert.NotNil(t, ErrArchiveNotFound)
	assert.NotNil(t, ErrReplayNotFound)
	assert.NotNil(t, ErrPermissionDenied)
	assert.NotNil(t, ErrQuotaExceeded)
	assert.NotNil(t, ErrEventSizeTooLarge)

	assert.Contains(t, ErrEventNotFound.Error(), "event not found")
	assert.Contains(t, ErrRuleNotFound.Error(), "rule not found")
	assert.Contains(t, ErrEventBusNotFound.Error(), "event bus not found")
	assert.Contains(t, ErrInvalidEventPattern.Error(), "invalid event pattern")
	assert.Contains(t, ErrInvalidEventDetail.Error(), "invalid event detail")
	assert.Contains(t, ErrArchiveNotFound.Error(), "archive not found")
	assert.Contains(t, ErrReplayNotFound.Error(), "replay not found")
	assert.Contains(t, ErrPermissionDenied.Error(), "permission denied")
	assert.Contains(t, ErrQuotaExceeded.Error(), "quota exceeded")
	assert.Contains(t, ErrEventSizeTooLarge.Error(), "event size too large")
}