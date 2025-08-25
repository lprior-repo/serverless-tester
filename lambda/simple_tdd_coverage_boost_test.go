package lambda

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple TDD tests to boost coverage without complex retry logic

func TestSimpleCoverage_TDD_EventBuilders(t *testing.T) {
	t.Run("S3Event_Basic", func(t *testing.T) {
		event, err := BuildS3EventE("bucket", "key.txt", "s3:ObjectCreated:Put")
		require.NoError(t, err)
		assert.Contains(t, event, "bucket")
		assert.Contains(t, event, "key.txt")
	})
	
	t.Run("DynamoDBEvent_Basic", func(t *testing.T) {
		keys := map[string]interface{}{"id": "123"}
		event, err := BuildDynamoDBEventE("table", "INSERT", keys)
		require.NoError(t, err)
		assert.Contains(t, event, "table")
		assert.Contains(t, event, "INSERT")
	})
	
	t.Run("SNSEvent_Basic", func(t *testing.T) {
		event, err := BuildSNSEventE("topic", "message")
		require.NoError(t, err)
		assert.Contains(t, event, "topic")
		assert.Contains(t, event, "message")
	})
	
	t.Run("SQSEvent_Basic", func(t *testing.T) {
		event, err := BuildSQSEventE("queue", "body")
		require.NoError(t, err)
		assert.Contains(t, event, "queue")
		assert.Contains(t, event, "body")
	})
	
	t.Run("APIGatewayEvent_Basic", func(t *testing.T) {
		event, err := BuildAPIGatewayEventE("GET", "/path", "body")
		require.NoError(t, err)
		assert.Contains(t, event, "GET")
		assert.Contains(t, event, "/path")
	})
	
	t.Run("CloudWatchEvent_Basic", func(t *testing.T) {
		detail := map[string]interface{}{"key": "value"}
		event, err := BuildCloudWatchEventE("source", "type", detail)
		require.NoError(t, err)
		assert.Contains(t, event, "source")
		assert.Contains(t, event, "type")
	})
	
	t.Run("KinesisEvent_Basic", func(t *testing.T) {
		event, err := BuildKinesisEventE("stream", "partition", "data")
		require.NoError(t, err)
		assert.Contains(t, event, "stream")
		assert.Contains(t, event, "data")
	})
}

func TestSimpleCoverage_TDD_UtilityFunctions(t *testing.T) {
	t.Run("MarshalPayloadE_SimpleTypes", func(t *testing.T) {
		result, err := MarshalPayloadE("hello")
		require.NoError(t, err)
		assert.Equal(t, `"hello"`, result)
		
		result, err = MarshalPayloadE(42)
		require.NoError(t, err)
		assert.Equal(t, "42", result)
		
		result, err = MarshalPayloadE(nil)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
	
	t.Run("ParseInvokeOutputE_BasicCases", func(t *testing.T) {
		result := &InvokeResult{Payload: `{"key": "value"}`}
		var target map[string]interface{}
		err := ParseInvokeOutputE(result, &target)
		require.NoError(t, err)
		assert.Equal(t, "value", target["key"])
	})
	
	t.Run("ValidationFunctions_BasicCases", func(t *testing.T) {
		err := validateFunctionName("valid-name")
		assert.NoError(t, err)
		
		err = validateFunctionName("")
		assert.Error(t, err)
		
		err = validatePayload(`{"valid": "json"}`)
		assert.NoError(t, err)
		
		err = validatePayload(`{invalid json}`)
		assert.Error(t, err)
	})
}

func TestSimpleCoverage_TDD_ConfigurationHelpers(t *testing.T) {
	t.Run("DefaultConfigurations", func(t *testing.T) {
		retryConfig := defaultRetryConfig()
		assert.Equal(t, DefaultRetryAttempts, retryConfig.MaxAttempts)
		
		invokeOptions := defaultInvokeOptions()
		assert.Equal(t, DefaultTimeout, invokeOptions.Timeout)
		assert.NotNil(t, invokeOptions.PayloadValidator)
		
		merged := mergeInvokeOptions(nil)
		assert.Equal(t, InvocationTypeRequestResponse, merged.InvocationType)
	})
}

func TestSimpleCoverage_TDD_StringUtilities(t *testing.T) {
	t.Run("StringOperations", func(t *testing.T) {
		assert.True(t, contains("hello world", "world"))
		assert.False(t, contains("hello", "xyz"))
		
		assert.Equal(t, 6, indexOf("hello world", "world"))
		assert.Equal(t, -1, indexOf("hello", "xyz"))
	})
}

func TestSimpleCoverage_TDD_BackoffCalculation(t *testing.T) {
	t.Run("BackoffCalculation", func(t *testing.T) {
		config := RetryConfig{
			BaseDelay:  100 * time.Millisecond,
			Multiplier: 2.0,
			MaxDelay:   5 * time.Second,
		}
		
		delay := calculateBackoffDelay(0, config)
		assert.Equal(t, 100*time.Millisecond, delay)
		
		delay = calculateBackoffDelay(1, config)
		assert.Equal(t, 200*time.Millisecond, delay)
		
		delay = calculateBackoffDelay(10, config)
		assert.Equal(t, 5*time.Second, delay)
	})
}

func TestSimpleCoverage_TDD_LogSanitization(t *testing.T) {
	t.Run("LogSanitization", func(t *testing.T) {
		input := `App log 1
RequestId: abc-123
App log 2
Duration: 100ms`
		
		result := sanitizeLogResult(input)
		assert.NotContains(t, result, "RequestId:")
		assert.NotContains(t, result, "Duration:")
		assert.Contains(t, result, "App log 1")
		assert.Contains(t, result, "App log 2")
	})
}

func TestSimpleCoverage_TDD_InvokeSimpleSuccess(t *testing.T) {
	t.Run("InvokeSuccessPath", func(t *testing.T) {
		ctx := &TestContext{T: t, Region: "us-east-1"}
		
		// Setup mock
		oldFactory := globalClientFactory
		defer func() { globalClientFactory = oldFactory }()
		
		mockFactory := &MockClientFactory{}
		globalClientFactory = mockFactory
		mockClient := &MockLambdaClient{}
		mockFactory.LambdaClient = mockClient
		
		mockClient.InvokeResponse = &lambda.InvokeOutput{
			StatusCode: 200,
			Payload:    []byte(`{"success": true}`),
		}
		
		result, err := InvokeWithOptionsE(ctx, "test-function", `{}`, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int32(200), result.StatusCode)
		assert.True(t, result.ExecutionTime > 0)
	})
}