package eventbridge

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestEventBusOperations_ComprehensiveCoverage tests all event bus operations
func TestEventBusOperations_ComprehensiveCoverage(t *testing.T) {
	t.Run("CreateEventBus with comprehensive configuration", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		config := EventBusConfig{
			Name:            "comprehensive-event-bus",
			Tags: map[string]string{
				"Environment": "production",
				"Team":        "orders",
				"Purpose":     "order-processing",
			},
			DeadLetterConfig: &types.DeadLetterConfig{
				Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:eventbus-dlq"),
			},
		}

		mockClient.On("CreateEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.CreateEventBusInput) bool {
			return *input.Name == "comprehensive-event-bus" &&
				*input.EventSourceName == "myapp.orders" &&
				len(input.Tags) == 3 &&
				input.Tags[0].Key != nil && // Verify tags are present
				input.DeadLetterConfig != nil &&
				*input.DeadLetterConfig.Arn == "arn:aws:sqs:us-east-1:123456789012:eventbus-dlq"
		}), mock.Anything).Return(&eventbridge.CreateEventBusOutput{
			EventBusArn: aws.String("arn:aws:events:us-east-1:123456789012:event-bus/comprehensive-event-bus"),
		}, nil)

		result, err := CreateEventBusE(ctx, config)

		require.NoError(t, err)
		assert.Equal(t, "comprehensive-event-bus", result.Name)
		assert.Equal(t, "arn:aws:events:us-east-1:123456789012:event-bus/comprehensive-event-bus", result.Arn)
		assert.Equal(t, "myapp.orders", result.EventSourceName)
		mockClient.AssertExpectations(t)
	})

	t.Run("CreateEventBus with minimal configuration", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-west-2",
			Client: mockClient,
		}

		config := EventBusConfig{
			Name: "minimal-event-bus",
		}

		mockClient.On("CreateEventBus", mock.Anything, mock.MatchedBy(func(input *eventbridge.CreateEventBusInput) bool {
			return *input.Name == "minimal-event-bus" &&
				input.EventSourceName == nil && // No event source
				len(input.Tags) == 0 &&         // No tags
				input.DeadLetterConfig == nil    // No DLQ
		}), mock.Anything).Return(&eventbridge.CreateEventBusOutput{
			EventBusArn: aws.String("arn:aws:events:us-west-2:123456789012:event-bus/minimal-event-bus"),
		}, nil)

		result, err := CreateEventBusE(ctx, config)

		require.NoError(t, err)
		assert.Equal(t, "minimal-event-bus", result.Name)
		assert.Equal(t, "arn:aws:events:us-west-2:123456789012:event-bus/minimal-event-bus", result.Arn)
		assert.Empty(t, result.EventSourceName)
		mockClient.AssertExpectations(t)
	})

	t.Run("ListEventBuses with multiple buses", func(t *testing.T) {
		mockClient := &MockEventBridgeClient{}
		ctx := &TestContext{
			T:      t,
			Region: "us-east-1",
			Client: mockClient,
		}

		creationTime1 := time.Now().Add(-48 * time.Hour)
		creationTime2 := time.Now().Add(-24 * time.Hour)

		mockClient.On("ListEventBuses", mock.Anything, mock.Anything, mock.Anything).Return(&eventbridge.ListEventBusesOutput{
			EventBuses: []types.EventBus{
				{
					Name:         aws.String("default"),
					Arn:          aws.String("arn:aws:events:us-east-1:123456789012:event-bus/default"),
					Description:  aws.String("Default event bus"),
					CreationTime: &creationTime1,
				},
				{
					Name:            aws.String("custom-bus-1"),
					Arn:             aws.String("arn:aws:events:us-east-1:123456789012:event-bus/custom-bus-1"),
					Description:     aws.String("First custom event bus"),
					CreationTime:    &creationTime2,
				},
				{
					Name:            aws.String("custom-bus-2"),
					Arn:             aws.String("arn:aws:events:us-east-1:123456789012:event-bus/custom-bus-2"),
					Description:     aws.String("Second custom event bus"),
					CreationTime:    &creationTime2,
				},
			},
		}, nil)

		results, err := ListEventBusesE(ctx)

		require.NoError(t, err)
		require.Len(t, results, 3)

		// Verify default bus
		defaultBus := results[0]
		assert.Equal(t, "default", defaultBus.Name)
		assert.Equal(t, "arn:aws:events:us-east-1:123456789012:event-bus/default", defaultBus.Arn)

		// Verify custom buses
		customBus1 := results[1]
		assert.Equal(t, "custom-bus-1", customBus1.Name)

		customBus2 := results[2]
		assert.Equal(t, "custom-bus-2", customBus2.Name)

		mockClient.AssertExpectations(t)
	})
}

// TestEventBusValidation_ComprehensiveCoverage tests validation logic
func TestEventBusValidation_ComprehensiveCoverage(t *testing.T) {
	t.Run("ValidateEventBusConfig should accept valid configuration", func(t *testing.T) {
		config := EventBusConfig{
			Name:            "valid-event-bus",
			Tags: map[string]string{
				"Environment": "production",
				"Team":        "backend",
			},
			DeadLetterConfig: &types.DeadLetterConfig{
				Arn: aws.String("arn:aws:sqs:us-east-1:123456789012:dlq"),
			},
		}

		err := ValidateEventBusConfigE(config)
		require.NoError(t, err)
	})

	t.Run("ValidateEventBusConfig should reject empty name", func(t *testing.T) {
		config := EventBusConfig{
			Name: "",
		}

		err := ValidateEventBusConfigE(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event bus name cannot be empty")
	})

	t.Run("ValidateEventBusName should accept valid names", func(t *testing.T) {
		validNames := []string{
			"valid-bus",
			"ValidBus123",
			"my.event.bus",
			"event_bus_name",
			"a", // Single character
			"a-very-long-event-bus-name-that-is-still-within-limits",
		}

		for _, name := range validNames {
			err := ValidateEventBusNameE(name)
			assert.NoError(t, err, "Name should be valid: %s", name)
		}
	})

	t.Run("ValidateEventBusName should reject invalid names", func(t *testing.T) {
		invalidNames := []string{
			"",                    // Empty
			"invalid@name",        // Invalid character @
			"invalid#name",        // Invalid character #
			"invalid name",        // Space not allowed
			"name-with-",          // Trailing dash
			"-name-with-leading",  // Leading dash
			"name.with.",          // Trailing dot
			".name.with.leading",  // Leading dot
		}

		for _, name := range invalidNames {
			err := ValidateEventBusNameE(name)
			assert.Error(t, err, "Name should be invalid: %s", name)
		}
	})
}