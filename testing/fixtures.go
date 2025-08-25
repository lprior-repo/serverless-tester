package testutils

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/random"
)

// TestFixtures provides common test data and fixtures for testing
type TestFixtures struct {
	namespace string
	timestamp string
}

// NewTestFixtures creates a new test fixtures instance with default namespace
func NewTestFixtures() *TestFixtures {
	return NewTestFixturesWithNamespace("test")
}

// NewTestFixturesWithNamespace creates a new test fixtures instance with custom namespace
func NewTestFixturesWithNamespace(namespace string) *TestFixtures {
	return &TestFixtures{
		namespace: namespace,
		timestamp: random.UniqueId(),
	}
}

// getShortId returns a safe substring of timestamp for use in names
func (f *TestFixtures) getShortId() string {
	if len(f.timestamp) >= 10 {
		return f.timestamp[:10]
	}
	return f.timestamp
}

// GetLambdaFunctionNames returns sample Lambda function names
func (f *TestFixtures) GetLambdaFunctionNames() []string {
	shortId := f.getShortId()
	return []string{
		fmt.Sprintf("%s-handler-function-%s", f.namespace, shortId),
		fmt.Sprintf("%s-processor-function-%s", f.namespace, shortId),
		fmt.Sprintf("%s-webhook-function-%s", f.namespace, shortId),
	}
}

// GetDynamoDBTableNames returns sample DynamoDB table names
func (f *TestFixtures) GetDynamoDBTableNames() []string {
	shortId := f.getShortId()
	return []string{
		fmt.Sprintf("%s-users-table-%s", f.namespace, shortId),
		fmt.Sprintf("%s-orders-table-%s", f.namespace, shortId),
		fmt.Sprintf("%s-sessions-table-%s", f.namespace, shortId),
	}
}

// GetEventBridgeRuleNames returns sample EventBridge rule names
func (f *TestFixtures) GetEventBridgeRuleNames() []string {
	shortId := f.getShortId()
	return []string{
		fmt.Sprintf("%s-user-created-rule-%s", f.namespace, shortId),
		fmt.Sprintf("%s-order-processed-rule-%s", f.namespace, shortId),
		fmt.Sprintf("%s-payment-received-rule-%s", f.namespace, shortId),
	}
}

// GetStepFunctionNames returns sample Step Function names
func (f *TestFixtures) GetStepFunctionNames() []string {
	shortId := f.getShortId()
	return []string{
		fmt.Sprintf("%s-order-workflow-%s", f.namespace, shortId),
		fmt.Sprintf("%s-user-onboarding-%s", f.namespace, shortId),
		fmt.Sprintf("%s-data-pipeline-%s", f.namespace, shortId),
	}
}

// GetSampleS3Event returns a sample S3 event payload
func (f *TestFixtures) GetSampleS3Event() string {
	return `{
	"Records": [
		{
			"eventVersion": "2.0",
			"eventSource": "aws:s3",
			"awsRegion": "us-east-1",
			"eventTime": "2023-01-01T12:00:00.000Z",
			"eventName": "ObjectCreated:Put",
			"userIdentity": {
				"principalId": "AWS:AIDACKCEVSQ6C2EXAMPLE"
			},
			"requestParameters": {
				"sourceIPAddress": "127.0.0.1"
			},
			"responseElements": {
				"x-amz-request-id": "C3D13FE58DE4C810",
				"x-amz-id-2": "FMyUVURIY8/IgAtTv8xRjskZQpcIZ9KG4V5Wp6S7S/JRWeUWerMUE5JgHvANOjpD"
			},
			"s3": {
				"s3SchemaVersion": "1.0",
				"configurationId": "testConfigRule",
				"bucket": {
					"name": "` + f.namespace + `-test-bucket",
					"ownerIdentity": {
						"principalId": "A3NL1KOZZKExample"
					},
					"arn": "arn:aws:s3:::` + f.namespace + `-test-bucket"
				},
				"object": {
					"key": "test-object.jpg",
					"size": 1024,
					"eTag": "d41d8cd98f00b204e9800998ecf8427e",
					"sequencer": "0A1B2C3D4E5F678901"
				}
			}
		}
	]
}`
}

// GetSampleSQSEvent returns a sample SQS event payload
func (f *TestFixtures) GetSampleSQSEvent() string {
	return `{
	"Records": [
		{
			"messageId": "059f36b4-87a3-44ab-83d2-661975830a7d",
			"receiptHandle": "AQEBwJnKyrHigUMZj6rYigCgxlaS3SLy0a...",
			"body": "Test message body",
			"attributes": {
				"ApproximateReceiveCount": "1",
				"SentTimestamp": "1545082649183",
				"SenderId": "AIDAIENQZJOLO23YVJ4VO",
				"ApproximateFirstReceiveTimestamp": "1545082649185"
			},
			"messageAttributes": {},
			"md5OfBody": "098f6bcd4621d373cade4e832627b4f6",
			"eventSource": "aws:sqs",
			"eventSourceARN": "arn:aws:sqs:us-east-1:123456789012:` + f.namespace + `-test-queue",
			"awsRegion": "us-east-1"
		}
	]
}`
}

// GetSampleAPIGatewayEvent returns a sample API Gateway event payload
func (f *TestFixtures) GetSampleAPIGatewayEvent() string {
	return `{
	"resource": "/users/{id}",
	"path": "/users/123",
	"httpMethod": "GET",
	"headers": {
		"Accept": "application/json",
		"CloudFront-Forwarded-Proto": "https",
		"CloudFront-Is-Desktop-Viewer": "true",
		"CloudFront-Is-Mobile-Viewer": "false",
		"CloudFront-Is-SmartTV-Viewer": "false",
		"CloudFront-Is-Tablet-Viewer": "false",
		"CloudFront-Viewer-Country": "US",
		"Content-Type": "application/json",
		"Host": "` + f.namespace + `-api.execute-api.us-east-1.amazonaws.com",
		"User-Agent": "Custom User Agent String",
		"Via": "1.1 08f323deadbeefa7af34d5feb414ce27.cloudfront.net (CloudFront)",
		"X-Amz-Cf-Id": "cDehVQoZnx43VYQb9j2-nvCh-9z396Uhbp027Y2JvkCPNLmGJHqlaA==",
		"X-Amzn-Trace-Id": "Root=1-5bdb8ca0-70a2a5371636894bfc"
	},
	"pathParameters": {
		"id": "123"
	},
	"queryStringParameters": {
		"filter": "active",
		"limit": "10"
	},
	"body": null,
	"isBase64Encoded": false,
	"stageVariables": {
		"environment": "test"
	},
	"requestContext": {
		"path": "/test/users/{id}",
		"accountId": "123456789012",
		"resourceId": "123456",
		"stage": "test",
		"requestId": "c6af9ac6-7b61-11e6-9a41-93e8deadbeef",
		"identity": {
			"cognitoIdentityPoolId": null,
			"accountId": null,
			"cognitoIdentityId": null,
			"caller": null,
			"apiKey": null,
			"sourceIp": "127.0.0.1",
			"cognitoAuthenticationType": null,
			"cognitoAuthenticationProvider": null,
			"userArn": null,
			"userAgent": "Custom User Agent String",
			"user": null
		},
		"resourcePath": "/users/{id}",
		"httpMethod": "GET",
		"apiId": "1234567890"
	}
}`
}

// GetSampleTerraformOutputs returns sample Terraform outputs structure
func (f *TestFixtures) GetSampleTerraformOutputs() map[string]interface{} {
	shortId := f.getShortId()
	return map[string]interface{}{
		"lambda_function_arn": fmt.Sprintf("arn:aws:lambda:us-east-1:123456789012:function:%s-handler-%s",
			f.namespace, shortId),
		"lambda_function_name": fmt.Sprintf("%s-handler-%s", f.namespace, shortId),
		"api_gateway_url": fmt.Sprintf("https://%s-api.execute-api.us-east-1.amazonaws.com/prod",
			shortId),
		"dynamodb_table_name": fmt.Sprintf("%s-users-%s", f.namespace, shortId),
		"dynamodb_table_arn": fmt.Sprintf("arn:aws:dynamodb:us-east-1:123456789012:table/%s-users-%s",
			f.namespace, shortId),
		"eventbridge_rule_name": fmt.Sprintf("%s-user-events-%s", f.namespace, shortId),
		"step_function_arn": fmt.Sprintf("arn:aws:states:us-east-1:123456789012:stateMachine:%s-workflow-%s",
			f.namespace, shortId),
	}
}

// GetSampleEnvironmentVariables returns common environment variables for testing
func (f *TestFixtures) GetSampleEnvironmentVariables() map[string]string {
	return map[string]string{
		"AWS_REGION":           "us-east-1",
		"ENVIRONMENT":          "test",
		"LOG_LEVEL":           "DEBUG",
		"DYNAMODB_TABLE_NAME": fmt.Sprintf("%s-test-table", f.namespace),
		"API_BASE_URL":        fmt.Sprintf("https://%s-api.example.com", f.namespace),
		"QUEUE_URL":           fmt.Sprintf("https://sqs.us-east-1.amazonaws.com/123456789012/%s-queue", f.namespace),
		"BUCKET_NAME":         fmt.Sprintf("%s-test-bucket", f.namespace),
	}
}

// GetSampleLambdaContext returns sample Lambda context data
func (f *TestFixtures) GetSampleLambdaContext() map[string]interface{} {
	return map[string]interface{}{
		"functionName":    fmt.Sprintf("%s-handler", f.namespace),
		"functionVersion": "$LATEST",
		"invokedFunctionArn": fmt.Sprintf("arn:aws:lambda:us-east-1:123456789012:function:%s-handler",
			f.namespace),
		"memoryLimitInMB": "128",
		"remainingTimeInMillis": 30000,
		"logGroupName":  fmt.Sprintf("/aws/lambda/%s-handler", f.namespace),
		"logStreamName": fmt.Sprintf("2023/01/01/[$LATEST]%s", f.getShortId()),
		"awsRequestId":  "c6af9ac6-7b61-11e6-9a41-93e8deadbeef",
	}
}