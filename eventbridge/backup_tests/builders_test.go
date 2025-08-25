package eventbridge

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildS3Event_Success(t *testing.T) {
	// Test successful S3 event building
	bucketName := "test-bucket"
	objectKey := "path/to/object.jpg"
	eventName := "s3:ObjectCreated:Put"

	event := BuildS3Event(bucketName, objectKey, eventName)

	assert.Equal(t, "aws.s3", event.Source)
	assert.Equal(t, "Object Created", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Parse detail JSON to verify structure
	var detail S3EventDetail
	err := json.Unmarshal([]byte(event.Detail), &detail)
	require.NoError(t, err)

	assert.Equal(t, "aws:s3", detail.EventSource)
	assert.Equal(t, eventName, detail.EventName)
	assert.Equal(t, bucketName, detail.S3.Bucket.Name)
	assert.Equal(t, objectKey, detail.S3.Object.Key)
	assert.Contains(t, detail.S3.Bucket.Arn, bucketName)
}

func TestBuildEC2Event_Success(t *testing.T) {
	// Test successful EC2 event building
	instanceID := "i-1234567890abcdef0"
	state := "running"
	previousState := "pending"

	event := BuildEC2Event(instanceID, state, previousState)

	assert.Equal(t, "aws.ec2", event.Source)
	assert.Equal(t, "EC2 Instance State-change Notification", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Parse detail JSON to verify structure
	var detail EC2EventDetail
	err := json.Unmarshal([]byte(event.Detail), &detail)
	require.NoError(t, err)

	assert.Equal(t, instanceID, detail.InstanceID)
	assert.Equal(t, state, detail.State)
	assert.Equal(t, previousState, detail.PreviousState)
}

func TestBuildEC2Event_NoPreviousState(t *testing.T) {
	// Test EC2 event building without previous state
	instanceID := "i-1234567890abcdef0"
	state := "running"

	event := BuildEC2Event(instanceID, state, "")

	var detail EC2EventDetail
	err := json.Unmarshal([]byte(event.Detail), &detail)
	require.NoError(t, err)

	assert.Equal(t, instanceID, detail.InstanceID)
	assert.Equal(t, state, detail.State)
	assert.Empty(t, detail.PreviousState)
}

func TestBuildLambdaEvent_Success(t *testing.T) {
	// Test successful Lambda event building
	functionName := "my-function"
	functionArn := "arn:aws:lambda:us-east-1:123456789012:function:my-function"
	state := "Active"

	event := BuildLambdaEvent(functionName, functionArn, state, "")

	assert.Equal(t, "aws.lambda", event.Source)
	assert.Equal(t, "Lambda Function State Change", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Parse detail JSON to verify structure
	var detail LambdaEventDetail
	err := json.Unmarshal([]byte(event.Detail), &detail)
	require.NoError(t, err)

	assert.Equal(t, functionName, detail.FunctionName)
	assert.Equal(t, functionArn, detail.FunctionArn)
	assert.Equal(t, state, detail.State)
}

func TestBuildDynamoDBEvent_Success(t *testing.T) {
	// Test successful DynamoDB event building
	tableName := "MyTable"
	eventName := "INSERT"
	record := map[string]interface{}{
		"Keys": map[string]interface{}{
			"id": map[string]interface{}{"S": "123"},
		},
	}

	event := BuildDynamoDBEvent(tableName, eventName, record)

	assert.Equal(t, "aws.dynamodb", event.Source)
	assert.Equal(t, "DynamoDB Table Event", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Parse detail JSON to verify structure
	var detail DynamoDBEventDetail
	err := json.Unmarshal([]byte(event.Detail), &detail)
	require.NoError(t, err)

	assert.Equal(t, tableName, detail.TableName)
	assert.Equal(t, eventName, detail.EventName)
	assert.Contains(t, detail.EventSourceArn, tableName)
	assert.NotEmpty(t, detail.DynamoDBRecord)
}

func TestBuildSQSEvent_Success(t *testing.T) {
	// Test successful SQS event building
	queueUrl := "https://sqs.us-east-1.amazonaws.com/123456789012/MyQueue"
	messageId := "msg-12345"
	body := "Hello World"

	event := BuildSQSEvent(queueUrl, messageId, body)

	assert.Equal(t, "aws.sqs", event.Source)
	assert.Equal(t, "SQS Message Available", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Parse detail JSON to verify structure
	var detail SQSEventDetail
	err := json.Unmarshal([]byte(event.Detail), &detail)
	require.NoError(t, err)

	assert.Equal(t, queueUrl, detail.QueueUrl)
	assert.Equal(t, messageId, detail.MessageId)
	assert.Equal(t, body, detail.Body)
	assert.Contains(t, detail.QueueArn, "MyQueue")
}

func TestBuildCodeCommitEvent_Success(t *testing.T) {
	// Test successful CodeCommit event building
	repositoryName := "MyRepo"
	commitId := "abc123def456"

	event := BuildCodeCommitEvent(repositoryName, "branch", "main", commitId)

	assert.Equal(t, "aws.codecommit", event.Source)
	assert.Equal(t, "CodeCommit Repository State Change", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Verify detail contains expected data
	assert.Contains(t, event.Detail, repositoryName)
	assert.Contains(t, event.Detail, commitId)
}

func TestBuildCodePipelineEvent_Success(t *testing.T) {
	// Test successful CodePipeline event building
	pipelineName := "MyPipeline"
	state := "SUCCEEDED"

	event := BuildCodePipelineEvent(pipelineName, state, "exec-123")

	assert.Equal(t, "aws.codepipeline", event.Source)
	assert.Equal(t, "CodePipeline Pipeline Execution State Change", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Verify detail contains expected data
	assert.Contains(t, event.Detail, pipelineName)
	assert.Contains(t, event.Detail, state)
}

func TestBuildECSEvent_Success(t *testing.T) {
	// Test successful ECS event building
	clusterArn := "arn:aws:ecs:us-east-1:123456789012:cluster/MyCluster"
	taskArn := "arn:aws:ecs:us-east-1:123456789012:task/abc123"
	lastStatus := "RUNNING"

	event := BuildECSEvent(clusterArn, taskArn, lastStatus, "RUNNING")

	assert.Equal(t, "aws.ecs", event.Source)
	assert.Equal(t, "ECS Task State Change", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Verify detail contains expected data
	assert.Contains(t, event.Detail, clusterArn)
	assert.Contains(t, event.Detail, taskArn)
	assert.Contains(t, event.Detail, lastStatus)
}

func TestBuildCloudWatchEvent_Success(t *testing.T) {
	// Test successful CloudWatch event building
	alarmName := "CPUUtilizationHigh"
	state := "ALARM"
	reason := "Threshold Crossed"

	event := BuildCloudWatchEvent(alarmName, state, reason, "CPUUtilization")

	assert.Equal(t, "aws.cloudwatch", event.Source)
	assert.Equal(t, "CloudWatch Alarm State Change", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Parse detail JSON to verify structure
	var detail map[string]interface{}
	err := json.Unmarshal([]byte(event.Detail), &detail)
	require.NoError(t, err)

	assert.Equal(t, alarmName, detail["alarmName"])
	assert.Equal(t, state, detail["newState"])
	assert.Equal(t, reason, detail["reason"])
}

func TestBuildAPIGatewayEvent_Success(t *testing.T) {
	// Test successful API Gateway event building
	apiId := "abc123def456"
	stage := "prod"
	requestId := "req-12345"

	event := BuildAPIGatewayEvent(apiId, stage, requestId, "200")

	assert.Equal(t, "aws.apigateway", event.Source)
	assert.Equal(t, "API Gateway Execution", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Verify detail contains expected data
	assert.Contains(t, event.Detail, apiId)
	assert.Contains(t, event.Detail, stage)
	assert.Contains(t, event.Detail, requestId)
}

func TestBuildSNSEvent_Success(t *testing.T) {
	// Test successful SNS event building
	topicArn := "arn:aws:sns:us-east-1:123456789012:MyTopic"
	message := "Hello SNS"

	event := BuildSNSEvent(topicArn, "Test Subject", message)

	assert.Equal(t, "aws.sns", event.Source)
	assert.Equal(t, "SNS Message", event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Verify detail contains expected data
	assert.Contains(t, event.Detail, topicArn)
	assert.Contains(t, event.Detail, message)
}

func TestBuildScheduledEventWithCron_Success(t *testing.T) {
	// Test successful cron-based scheduled event building
	cronExpression := "0 2 * * ? *" // Daily at 2 AM

	ruleConfig := BuildScheduledEventWithCron(cronExpression, map[string]interface{}{})

	assert.Contains(t, ruleConfig.Name, "scheduled-rule")
	assert.Contains(t, ruleConfig.Description, cronExpression)
	assert.Contains(t, ruleConfig.ScheduleExpression, "cron")
	assert.Equal(t, RuleStateEnabled, ruleConfig.State)
	assert.Equal(t, DefaultEventBusName, ruleConfig.EventBusName)
}

func TestBuildScheduledEventWithRate_Success(t *testing.T) {
	// Test successful rate-based scheduled event building

	ruleConfig := BuildScheduledEventWithRate(5, "minutes")

	assert.Contains(t, ruleConfig.Name, "rate-rule")
	assert.Contains(t, ruleConfig.Description, "every 5 minutes")
	assert.Contains(t, ruleConfig.ScheduleExpression, "rate")
	assert.Equal(t, RuleStateEnabled, ruleConfig.State)
	assert.Equal(t, DefaultEventBusName, ruleConfig.EventBusName)
}

func TestAllBuilders_ProduceValidJSON(t *testing.T) {
	// Test that all builders produce valid JSON
	builders := map[string]CustomEvent{
		"S3":           BuildS3Event("bucket", "key", "s3:ObjectCreated:Put"),
		"EC2":          BuildEC2Event("i-123", "running", "pending"),
		"Lambda":       BuildLambdaEvent("func", "arn:aws:lambda:us-east-1:123:function:func", "Active", ""),
		"DynamoDB":     BuildDynamoDBEvent("table", "INSERT", map[string]interface{}{"test": "data"}),
		"SQS":          BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123/queue", "msg", "body"),
		"CodeCommit":   BuildCodeCommitEvent("repo", "branch", "main", "commit123"),
		"CodePipeline": BuildCodePipelineEvent("pipeline", "SUCCEEDED", "exec-123"),
		"ECS":          BuildECSEvent("cluster-arn", "task-arn", "RUNNING", "RUNNING"),
		"CloudWatch":   BuildCloudWatchEvent("alarm", "ALARM", "reason", "metric"),
		"APIGateway":   BuildAPIGatewayEvent("api123", "prod", "req123", "200"),
		"SNS":          BuildSNSEvent("arn:aws:sns:us-east-1:123:topic", "subject", "message"),
	}

	for name, event := range builders {
		t.Run(name, func(t *testing.T) {
			// Verify it's valid JSON
			var jsonData interface{}
			err := json.Unmarshal([]byte(event.Detail), &jsonData)
			require.NoError(t, err, "Invalid JSON produced by %s builder", name)

			// Verify basic event properties
			assert.NotEmpty(t, event.Source, "%s builder should set Source", name)
			assert.NotEmpty(t, event.DetailType, "%s builder should set DetailType", name)
			assert.Equal(t, DefaultEventBusName, event.EventBusName, "%s builder should set EventBusName", name)
			assert.NotEmpty(t, event.Detail, "%s builder should set Detail", name)

			// Verify source follows AWS convention
			assert.True(t, strings.HasPrefix(event.Source, "aws."), "%s builder source should start with 'aws.'", name)
		})
	}
}

func TestBuilders_EmptyInputs(t *testing.T) {
	// Test builders handle empty inputs gracefully
	tests := []struct {
		name    string
		builder func() CustomEvent
	}{
		{"S3_Empty", func() CustomEvent { return BuildS3Event("", "", "") }},
		{"EC2_Empty", func() CustomEvent { return BuildEC2Event("", "", "") }},
		{"Lambda_Empty", func() CustomEvent { return BuildLambdaEvent("", "", "", "") }},
		{"DynamoDB_Empty", func() CustomEvent { return BuildDynamoDBEvent("", "", nil) }},
		{"SQS_Empty", func() CustomEvent { return BuildSQSEvent("", "", "") }},
		{"CodeCommit_Empty", func() CustomEvent { return BuildCodeCommitEvent("", "", "", "") }},
		{"CodePipeline_Empty", func() CustomEvent { return BuildCodePipelineEvent("", "", "") }},
		{"ECS_Empty", func() CustomEvent { return BuildECSEvent("", "", "", "") }},
		{"CloudWatch_Empty", func() CustomEvent { return BuildCloudWatchEvent("", "", "", "") }},
		{"APIGateway_Empty", func() CustomEvent { return BuildAPIGatewayEvent("", "", "", "") }},
		{"SNS_Empty", func() CustomEvent { return BuildSNSEvent("", "", "") }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := test.builder()

			// Should still produce valid JSON even with empty inputs
			var jsonData interface{}
			err := json.Unmarshal([]byte(event.Detail), &jsonData)
			require.NoError(t, err, "Builder should produce valid JSON even with empty inputs")

			// Basic properties should still be set
			assert.NotEmpty(t, event.Source)
			assert.NotEmpty(t, event.DetailType)
			assert.Equal(t, DefaultEventBusName, event.EventBusName)
		})
	}
}

func TestBuildCustomEvent_Success(t *testing.T) {
	// Test the generic custom event builder
	source := "custom.service"
	detailType := "Custom Event"
	detail := map[string]interface{}{
		"key":    "value",
		"number": 42,
		"nested": map[string]interface{}{
			"innerKey": "innerValue",
		},
	}

	event := BuildCustomEvent(source, detailType, detail)

	assert.Equal(t, source, event.Source)
	assert.Equal(t, detailType, event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Verify JSON detail
	var parsedDetail map[string]interface{}
	err := json.Unmarshal([]byte(event.Detail), &parsedDetail)
	require.NoError(t, err)

	assert.Equal(t, "value", parsedDetail["key"])
	assert.Equal(t, float64(42), parsedDetail["number"]) // JSON numbers are float64
	
	nested, ok := parsedDetail["nested"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "innerValue", nested["innerKey"])
}

func TestBuildScheduledEvent_Success(t *testing.T) {
	// Test the scheduled event builder
	detailType := "Custom Scheduled Event"
	detail := map[string]interface{}{
		"action": "cleanup",
		"target": "old-logs",
	}

	event := BuildScheduledEvent(detailType, detail)

	assert.Equal(t, "aws.events", event.Source)
	assert.Equal(t, detailType, event.DetailType)
	assert.Equal(t, DefaultEventBusName, event.EventBusName)

	// Verify JSON detail
	var parsedDetail map[string]interface{}
	err := json.Unmarshal([]byte(event.Detail), &parsedDetail)
	require.NoError(t, err)

	assert.Equal(t, "cleanup", parsedDetail["action"])
	assert.Equal(t, "old-logs", parsedDetail["target"])
}