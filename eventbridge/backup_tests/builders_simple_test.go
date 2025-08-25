package eventbridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Simple tests for key builder functions to boost coverage quickly

func TestBuildS3Event_Basic(t *testing.T) {
	event := BuildS3Event("test-bucket", "test-key", "ObjectCreated:Put")

	assert.Equal(t, "aws.s3", event.Source)
	assert.NotEmpty(t, event.DetailType)
	assert.Contains(t, event.Detail, "test-bucket")
	assert.Contains(t, event.Detail, "test-key")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildEC2Event_Basic(t *testing.T) {
	event := BuildEC2Event("i-123456789", "running", "pending")

	assert.Equal(t, "aws.ec2", event.Source)
	assert.Equal(t, "EC2 Instance State-change Notification", event.DetailType)
	assert.Contains(t, event.Detail, "i-123456789")
	assert.Contains(t, event.Detail, "running")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildLambdaEvent_Basic(t *testing.T) {
	event := BuildLambdaEvent("test-function", "arn:aws:lambda:us-east-1:123:function:test", "Active", "")

	assert.Equal(t, "aws.lambda", event.Source)
	assert.Equal(t, "Lambda Function State Change", event.DetailType)
	assert.Contains(t, event.Detail, "test-function")
	assert.Contains(t, event.Detail, "Active")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildDynamoDBEvent_Basic(t *testing.T) {
	record := map[string]interface{}{
		"Keys": map[string]interface{}{
			"id": map[string]interface{}{
				"S": "123",
			},
		},
	}
	event := BuildDynamoDBEvent("test-table", "INSERT", record)

	assert.Equal(t, "aws.dynamodb", event.Source)
	assert.Equal(t, "DynamoDB Stream Record", event.DetailType)
	assert.Contains(t, event.Detail, "test-table")
	assert.Contains(t, event.Detail, "INSERT")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildSQSEvent_Basic(t *testing.T) {
	event := BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/test-queue", "msg-123", "test message")

	assert.Equal(t, "aws.sqs", event.Source)
	assert.Equal(t, "SQS Message", event.DetailType)
	assert.Contains(t, event.Detail, "test-queue")
	assert.Contains(t, event.Detail, "test message")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildCodeCommitEvent_Basic(t *testing.T) {
	event := BuildCodeCommitEvent("test-repo", "branch", "main", "abc123")

	assert.Equal(t, "aws.codecommit", event.Source)
	assert.Equal(t, "CodeCommit Repository State Change", event.DetailType)
	assert.Contains(t, event.Detail, "test-repo")
	assert.Contains(t, event.Detail, "main")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildCodePipelineEvent_Basic(t *testing.T) {
	event := BuildCodePipelineEvent("test-pipeline", "STARTED", "exec-123")

	assert.Equal(t, "aws.codepipeline", event.Source)
	assert.Equal(t, "CodePipeline Pipeline Execution State Change", event.DetailType)
	assert.Contains(t, event.Detail, "test-pipeline")
	assert.Contains(t, event.Detail, "STARTED")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildECSEvent_Basic(t *testing.T) {
	event := BuildECSEvent("cluster-arn", "task-arn", "RUNNING", "RUNNING")

	assert.Equal(t, "aws.ecs", event.Source)
	assert.Equal(t, "ECS Task State Change", event.DetailType)
	assert.Contains(t, event.Detail, "cluster-arn")
	assert.Contains(t, event.Detail, "task-arn")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildCloudWatchEvent_Basic(t *testing.T) {
	event := BuildCloudWatchEvent("test-alarm", "ALARM", "High CPU", "CPUUtilization")

	assert.Equal(t, "aws.cloudwatch", event.Source)
	assert.Equal(t, "CloudWatch Alarm State Change", event.DetailType)
	assert.Contains(t, event.Detail, "test-alarm")
	assert.Contains(t, event.Detail, "ALARM")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildAPIGatewayEvent_Basic(t *testing.T) {
	event := BuildAPIGatewayEvent("api-123", "prod", "req-456", "200")

	assert.Equal(t, "aws.apigateway", event.Source)
	assert.Equal(t, "API Gateway Execution State Change", event.DetailType)
	assert.Contains(t, event.Detail, "api-123")
	assert.Contains(t, event.Detail, "prod")
	assert.Equal(t, "default", event.EventBusName)
}

func TestBuildSNSEvent_Basic(t *testing.T) {
	event := BuildSNSEvent("arn:aws:sns:us-east-1:123:test-topic", "Test Subject", "Test message")

	assert.Equal(t, "aws.sns", event.Source)
	assert.Equal(t, "SNS Message", event.DetailType)
	assert.Contains(t, event.Detail, "test-topic")
	assert.Contains(t, event.Detail, "Test message")
	assert.Equal(t, "default", event.EventBusName)
}