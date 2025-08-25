package eventbridge

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateBackoffDelay_ZeroAttempt(t *testing.T) {
	// Test with zero attempt should return base delay
	config := RetryConfig{
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}

	delay := calculateBackoffDelay(0, config)
	assert.Equal(t, 1*time.Second, delay)
}

func TestCalculateBackoffDelay_ExponentialGrowth(t *testing.T) {
	// Test exponential growth
	config := RetryConfig{
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}

	// First attempt: 1 * 2^1 = 2 seconds
	delay1 := calculateBackoffDelay(1, config)
	assert.Equal(t, 2*time.Second, delay1)

	// Second attempt: 1 * 2^2 = 4 seconds
	delay2 := calculateBackoffDelay(2, config)
	assert.Equal(t, 4*time.Second, delay2)

	// Third attempt: 1 * 2^3 = 8 seconds
	delay3 := calculateBackoffDelay(3, config)
	assert.Equal(t, 8*time.Second, delay3)
}

func TestCalculateBackoffDelay_MaxDelayCap(t *testing.T) {
	// Test that delay is capped at MaxDelay
	config := RetryConfig{
		BaseDelay:   1 * time.Second,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
	}

	// This would normally be 1 * 2^10 = 1024 seconds, but should be capped at 10
	delay := calculateBackoffDelay(10, config)
	assert.Equal(t, 10*time.Second, delay)
}

func TestCalculateBackoffDelay_DifferentMultiplier(t *testing.T) {
	// Test with different multiplier
	config := RetryConfig{
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  1.5,
	}

	// First attempt: 1 * 1.5^1 = 1.5 seconds
	delay := calculateBackoffDelay(1, config)
	assert.Equal(t, 1500*time.Millisecond, delay)
}

func TestDefaultRetryConfig_Values(t *testing.T) {
	// Test default retry configuration values
	config := defaultRetryConfig()

	assert.Equal(t, DefaultRetryAttempts, config.MaxAttempts)
	assert.Equal(t, DefaultRetryDelay, config.BaseDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

func TestValidateEventDetail_AtSizeLimit(t *testing.T) {
	// Test detail at exactly the size limit (should be valid)
	detailAtLimit := make([]byte, MaxEventDetailSize)
	for i := range detailAtLimit {
		detailAtLimit[i] = 'a'
	}
	// Make it valid JSON
	detailJSON := `"` + string(detailAtLimit[:MaxEventDetailSize-2]) + `"`
	
	err := validateEventDetail(detailJSON)
	assert.NoError(t, err)
}

func TestSafeString_WithEmptyString(t *testing.T) {
	// Test safeString with pointer to empty string
	value := ""
	result := safeString(&value)
	assert.Equal(t, "", result)
}

func TestConstants_Values(t *testing.T) {
	// Test that constants have expected values
	assert.Equal(t, 30*time.Second, DefaultTimeout)
	assert.Equal(t, 3, DefaultRetryAttempts)
	assert.Equal(t, 1*time.Second, DefaultRetryDelay)
	assert.Equal(t, 10, MaxEventBatchSize)
	assert.Equal(t, 256*1024, MaxEventDetailSize)
	assert.Equal(t, "default", DefaultEventBusName)
}

func TestConstants_RuleStates(t *testing.T) {
	// Test rule state constants
	assert.Equal(t, "ENABLED", string(RuleStateEnabled))
	assert.Equal(t, "DISABLED", string(RuleStateDisabled))
}

func TestConstants_ArchiveStates(t *testing.T) {
	// Test archive state constants
	assert.Equal(t, "ENABLED", string(ArchiveStateEnabled))
	assert.Equal(t, "DISABLED", string(ArchiveStateDisabled))
}

func TestConstants_ReplayStates(t *testing.T) {
	// Test replay state constants
	assert.Equal(t, "STARTING", string(ReplayStateStarting))
	assert.Equal(t, "RUNNING", string(ReplayStateRunning))
	assert.Equal(t, "COMPLETED", string(ReplayStateCompleted))
	assert.Equal(t, "FAILED", string(ReplayStateFailed))
}

func TestErrors_Defined(t *testing.T) {
	// Test that all expected errors are defined
	require.NotNil(t, ErrEventNotFound)
	require.NotNil(t, ErrRuleNotFound)
	require.NotNil(t, ErrEventBusNotFound)
	require.NotNil(t, ErrInvalidEventPattern)
	require.NotNil(t, ErrInvalidEventDetail)
	require.NotNil(t, ErrArchiveNotFound)
	require.NotNil(t, ErrReplayNotFound)
	require.NotNil(t, ErrPermissionDenied)
	require.NotNil(t, ErrQuotaExceeded)
	require.NotNil(t, ErrEventSizeTooLarge)

	// Test error messages
	assert.Equal(t, "event not found", ErrEventNotFound.Error())
	assert.Equal(t, "rule not found", ErrRuleNotFound.Error())
	assert.Equal(t, "event bus not found", ErrEventBusNotFound.Error())
	assert.Equal(t, "invalid event pattern", ErrInvalidEventPattern.Error())
	assert.Equal(t, "invalid event detail", ErrInvalidEventDetail.Error())
	assert.Equal(t, "archive not found", ErrArchiveNotFound.Error())
	assert.Equal(t, "replay not found", ErrReplayNotFound.Error())
	assert.Equal(t, "permission denied", ErrPermissionDenied.Error())
	assert.Equal(t, "quota exceeded", ErrQuotaExceeded.Error())
	assert.Equal(t, "event size too large", ErrEventSizeTooLarge.Error())
}

func TestCreateEventBridgeClient_WithMockClient(t *testing.T) {
	// Test that mock client is returned when provided
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}

	client := createEventBridgeClient(ctx)
	assert.Equal(t, mockClient, client)
}

func TestCreateEventBridgeClient_WithoutMockClient(t *testing.T) {
	// Test that real client is created when no mock provided
	ctx := &TestContext{
		T:      t,
		Client: nil,
	}

	client := createEventBridgeClient(ctx)
	assert.NotNil(t, client)
	// Should not be our mock type
	_, isMock := client.(*MockEventBridgeClient)
	assert.False(t, isMock)
}