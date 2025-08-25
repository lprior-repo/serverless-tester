package eventbridge

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test validateEventDetail function
func TestValidateEventDetail_ValidJSON(t *testing.T) {
	// RED: Test valid JSON detail
	detail := `{"key": "value", "number": 42}`
	err := validateEventDetail(detail)
	
	assert.NoError(t, err)
}

func TestValidateEventDetail_EmptyString(t *testing.T) {
	// RED: Test empty detail (should be valid)
	detail := ""
	err := validateEventDetail(detail)
	
	assert.NoError(t, err)
}

func TestValidateEventDetail_MalformedJSON(t *testing.T) {
	// RED: Test invalid JSON
	detail := `{"key": "value", "invalid": }`
	err := validateEventDetail(detail)
	
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidEventDetail)
}

func TestValidateEventDetail_SizeTooLarge(t *testing.T) {
	// RED: Test detail size exceeding maximum
	largeDetail := strings.Repeat("a", MaxEventDetailSize+1)
	err := validateEventDetail(largeDetail)
	
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrEventSizeTooLarge)
}

func TestValidateEventDetail_MaxSizeValid(t *testing.T) {
	// RED: Test detail at exactly max size with valid JSON
	baseJSON := `{"data":"`
	suffix := `"}`
	padding := strings.Repeat("x", MaxEventDetailSize-len(baseJSON)-len(suffix))
	maxSizeDetail := baseJSON + padding + suffix
	
	err := validateEventDetail(maxSizeDetail)
	
	assert.NoError(t, err)
}

// Test calculateBackoffDelay function
func TestCalculateBackoffDelay_FirstAttempt(t *testing.T) {
	// RED: Test first attempt (attempt = 0)
	config := RetryConfig{
		BaseDelay:  1 * time.Second,
		Multiplier: 2.0,
		MaxDelay:   30 * time.Second,
	}
	
	delay := calculateBackoffDelay(0, config)
	
	assert.Equal(t, 1*time.Second, delay)
}

func TestCalculateBackoffDelay_SecondAttempt(t *testing.T) {
	// RED: Test second attempt (attempt = 1)
	config := RetryConfig{
		BaseDelay:  1 * time.Second,
		Multiplier: 2.0,
		MaxDelay:   30 * time.Second,
	}
	
	delay := calculateBackoffDelay(1, config)
	
	assert.Equal(t, 2*time.Second, delay)
}

func TestCalculateBackoffDelay_ThirdAttempt(t *testing.T) {
	// RED: Test third attempt (attempt = 2)
	config := RetryConfig{
		BaseDelay:  1 * time.Second,
		Multiplier: 2.0,
		MaxDelay:   30 * time.Second,
	}
	
	delay := calculateBackoffDelay(2, config)
	
	assert.Equal(t, 4*time.Second, delay)
}

func TestCalculateBackoffDelay_MaxDelayReached(t *testing.T) {
	// RED: Test that delay is capped at maximum
	config := RetryConfig{
		BaseDelay:  1 * time.Second,
		Multiplier: 2.0,
		MaxDelay:   5 * time.Second,
	}
	
	delay := calculateBackoffDelay(10, config) // Should exceed max delay
	
	assert.Equal(t, 5*time.Second, delay)
}

func TestCalculateBackoffDelay_NegativeAttempt(t *testing.T) {
	// RED: Test negative attempt returns base delay
	config := RetryConfig{
		BaseDelay:  1 * time.Second,
		Multiplier: 2.0,
		MaxDelay:   30 * time.Second,
	}
	
	delay := calculateBackoffDelay(-1, config)
	
	assert.Equal(t, 1*time.Second, delay)
}

func TestCalculateBackoffDelay_ZeroMultiplier(t *testing.T) {
	// RED: Test zero multiplier
	config := RetryConfig{
		BaseDelay:  2 * time.Second,
		Multiplier: 0.0,
		MaxDelay:   30 * time.Second,
	}
	
	delay := calculateBackoffDelay(3, config)
	
	assert.Equal(t, time.Duration(0), delay)
}

func TestCalculateBackoffDelay_FractionalMultiplier(t *testing.T) {
	// RED: Test fractional multiplier (exponential decay)
	config := RetryConfig{
		BaseDelay:  4 * time.Second,
		Multiplier: 0.5,
		MaxDelay:   30 * time.Second,
	}
	
	delay := calculateBackoffDelay(2, config) // 4 * (0.5^2) = 1 second
	
	assert.Equal(t, 1*time.Second, delay)
}

// Test safeString function
func TestSafeString_ValidString(t *testing.T) {
	// RED: Test valid string pointer
	str := "test string"
	result := safeString(&str)
	
	assert.Equal(t, "test string", result)
}

func TestSafeString_NilPointer(t *testing.T) {
	// RED: Test nil pointer returns empty string
	result := safeString(nil)
	
	assert.Equal(t, "", result)
}

func TestSafeString_EmptyString(t *testing.T) {
	// RED: Test empty string pointer
	str := ""
	result := safeString(&str)
	
	assert.Equal(t, "", result)
}

// Test defaultRetryConfig function
func TestDefaultRetryConfig(t *testing.T) {
	// RED: Test default retry configuration values
	config := defaultRetryConfig()
	
	assert.Equal(t, DefaultRetryAttempts, config.MaxAttempts)
	assert.Equal(t, DefaultRetryDelay, config.BaseDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

// Test createEventBridgeClient function with mock client
func TestCreateEventBridgeClient_MockClientReturned(t *testing.T) {
	// RED: Test that provided mock client is returned
	mockClient := &MockEventBridgeClient{}
	ctx := &TestContext{
		T:      t,
		Client: mockClient,
	}
	
	client := createEventBridgeClient(ctx)
	
	assert.Equal(t, mockClient, client)
}

func TestCreateEventBridgeClient_RealClientCreated(t *testing.T) {
	// RED: Test that real client is created when no mock provided
	ctx := &TestContext{
		T:      t,
		Client: nil, // No mock client
	}
	
	// This will create a real client, which should not be nil
	client := createEventBridgeClient(ctx)
	
	assert.NotNil(t, client)
}