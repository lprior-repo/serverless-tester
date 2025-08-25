package eventbridge

import (
	"encoding/json"
	"fmt"
	"time"
)

// S3EventDetail represents the detail structure for S3 events
type S3EventDetail struct {
	EventVersion      string `json:"eventVersion"`
	EventSource       string `json:"eventSource"`
	EventName         string `json:"eventName"`
	EventTime         string `json:"eventTime"`
	AwsRegion         string `json:"awsRegion"`
	SourceIPAddress   string `json:"sourceIPAddress"`
	UserAgent         string `json:"userAgent"`
	RequestParameters struct {
		SourceIPAddress string `json:"sourceIPAddress"`
	} `json:"requestParameters"`
	ResponseElements struct {
		XAmzRequestID string `json:"x-amz-request-id"`
		XAmzID2       string `json:"x-amz-id-2"`
	} `json:"responseElements"`
	S3 struct {
		S3SchemaVersion string `json:"s3SchemaVersion"`
		ConfigurationID string `json:"configurationId"`
		Bucket          struct {
			Name          string `json:"name"`
			OwnerIdentity struct {
				PrincipalID string `json:"principalId"`
			} `json:"ownerIdentity"`
			Arn string `json:"arn"`
		} `json:"bucket"`
		Object struct {
			Key       string `json:"key"`
			Size      int64  `json:"size"`
			ETag      string `json:"eTag"`
			Sequencer string `json:"sequencer"`
		} `json:"object"`
	} `json:"s3"`
}

// EC2EventDetail represents the detail structure for EC2 events
type EC2EventDetail struct {
	InstanceID    string `json:"instance-id"`
	State         string `json:"state"`
	PreviousState string `json:"previous-state,omitempty"`
}

// LambdaEventDetail represents the detail structure for Lambda events
type LambdaEventDetail struct {
	FunctionName    string `json:"functionName"`
	FunctionArn     string `json:"functionArn"`
	Runtime         string `json:"runtime"`
	State           string `json:"state"`
	PreviousState   string `json:"previousState,omitempty"`
	StateReason     string `json:"stateReason,omitempty"`
	StateReasonCode string `json:"stateReasonCode,omitempty"`
}

// DynamoDBEventDetail represents the detail structure for DynamoDB events
type DynamoDBEventDetail struct {
	TableName       string                 `json:"tableName"`
	EventName       string                 `json:"eventName"`
	EventSourceArn  string                 `json:"eventSourceArn"`
	EventSource     string                 `json:"eventSource"`
	AwsRegion       string                 `json:"awsRegion"`
	DynamoDBRecord  map[string]interface{} `json:"dynamodb"`
}

// SQSEventDetail represents the detail structure for SQS events
type SQSEventDetail struct {
	QueueUrl      string                 `json:"queueUrl"`
	QueueArn      string                 `json:"queueArn"`
	MessageId     string                 `json:"messageId"`
	ReceiptHandle string                 `json:"receiptHandle"`
	Body          string                 `json:"body"`
	Attributes    map[string]interface{} `json:"attributes"`
}

// BuildS3Event creates an S3 EventBridge event
func BuildS3Event(bucketName string, objectKey string, eventName string) CustomEvent {
	detail := S3EventDetail{
		EventVersion: "2.1",
		EventSource:  "aws:s3",
		EventName:    eventName,
		EventTime:    time.Now().UTC().Format(time.RFC3339),
		AwsRegion:    "us-east-1",
		S3: struct {
			S3SchemaVersion string `json:"s3SchemaVersion"`
			ConfigurationID string `json:"configurationId"`
			Bucket          struct {
				Name          string `json:"name"`
				OwnerIdentity struct {
					PrincipalID string `json:"principalId"`
				} `json:"ownerIdentity"`
				Arn string `json:"arn"`
			} `json:"bucket"`
			Object struct {
				Key       string `json:"key"`
				Size      int64  `json:"size"`
				ETag      string `json:"eTag"`
				Sequencer string `json:"sequencer"`
			} `json:"object"`
		}{
			S3SchemaVersion: "1.0",
			Bucket: struct {
				Name          string `json:"name"`
				OwnerIdentity struct {
					PrincipalID string `json:"principalId"`
				} `json:"ownerIdentity"`
				Arn string `json:"arn"`
			}{
				Name: bucketName,
				Arn:  fmt.Sprintf("arn:aws:s3:::%s", bucketName),
			},
			Object: struct {
				Key       string `json:"key"`
				Size      int64  `json:"size"`
				ETag      string `json:"eTag"`
				Sequencer string `json:"sequencer"`
			}{
				Key:  objectKey,
				Size: 1024,
				ETag: "d41d8cd98f00b204e9800998ecf8427e",
			},
		},
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.s3",
		DetailType:   "Object Created",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{fmt.Sprintf("arn:aws:s3:::%s/%s", bucketName, objectKey)},
	}
}

// BuildEC2Event creates an EC2 EventBridge event
func BuildEC2Event(instanceID string, state string, previousState string) CustomEvent {
	detail := EC2EventDetail{
		InstanceID: instanceID,
		State:      state,
	}
	
	if previousState != "" {
		detail.PreviousState = previousState
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	detailType := "EC2 Instance State-change Notification"
	
	return CustomEvent{
		Source:       "aws.ec2",
		DetailType:   detailType,
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{fmt.Sprintf("arn:aws:ec2:us-east-1:123456789012:instance/%s", instanceID)},
	}
}

// BuildLambdaEvent creates a Lambda EventBridge event
func BuildLambdaEvent(functionName string, functionArn string, state string, stateReason string) CustomEvent {
	detail := LambdaEventDetail{
		FunctionName: functionName,
		FunctionArn:  functionArn,
		Runtime:      "python3.9",
		State:        state,
		StateReason:  stateReason,
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.lambda",
		DetailType:   "Lambda Function State Change",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{functionArn},
	}
}

// BuildDynamoDBEvent creates a DynamoDB EventBridge event
func BuildDynamoDBEvent(tableName string, eventName string, record map[string]interface{}) CustomEvent {
	detail := DynamoDBEventDetail{
		TableName:      tableName,
		EventName:      eventName,
		EventSource:    "aws:dynamodb",
		EventSourceArn: fmt.Sprintf("arn:aws:dynamodb:us-east-1:123456789012:table/%s", tableName),
		AwsRegion:      "us-east-1",
		DynamoDBRecord: record,
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.dynamodb",
		DetailType:   "DynamoDB Table Event",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{fmt.Sprintf("arn:aws:dynamodb:us-east-1:123456789012:table/%s", tableName)},
	}
}

// BuildSQSEvent creates an SQS EventBridge event
func BuildSQSEvent(queueUrl string, messageID string, messageBody string) CustomEvent {
	detail := SQSEventDetail{
		QueueUrl:      queueUrl,
		QueueArn:      fmt.Sprintf("arn:aws:sqs:us-east-1:123456789012:%s", queueUrl),
		MessageId:     messageID,
		Body:          messageBody,
		Attributes:    make(map[string]interface{}),
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.sqs",
		DetailType:   "SQS Message Available",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{detail.QueueArn},
	}
}

// BuildCodeCommitEvent creates a CodeCommit EventBridge event
func BuildCodeCommitEvent(repositoryName string, referenceType string, referenceName string, commitID string) CustomEvent {
	detail := map[string]interface{}{
		"repositoryName": repositoryName,
		"referenceType":  referenceType,
		"referenceName":  referenceName,
		"commitId":       commitID,
		"event":          "referenceCreated",
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.codecommit",
		DetailType:   "CodeCommit Repository State Change",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{fmt.Sprintf("arn:aws:codecommit:us-east-1:123456789012:%s", repositoryName)},
	}
}

// BuildCodePipelineEvent creates a CodePipeline EventBridge event
func BuildCodePipelineEvent(pipelineName string, state string, executionID string) CustomEvent {
	detail := map[string]interface{}{
		"pipeline":         pipelineName,
		"state":            state,
		"execution-id":     executionID,
		"pipeline-arn":     fmt.Sprintf("arn:aws:codepipeline:us-east-1:123456789012:pipeline/%s", pipelineName),
		"execution-result": map[string]interface{}{},
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.codepipeline",
		DetailType:   "CodePipeline Pipeline Execution State Change",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{fmt.Sprintf("arn:aws:codepipeline:us-east-1:123456789012:pipeline/%s", pipelineName)},
	}
}

// BuildECSEvent creates an ECS EventBridge event
func BuildECSEvent(clusterArn string, taskArn string, lastStatus string, desiredStatus string) CustomEvent {
	detail := map[string]interface{}{
		"clusterArn":    clusterArn,
		"taskArn":       taskArn,
		"lastStatus":    lastStatus,
		"desiredStatus": desiredStatus,
		"containers":    []interface{}{},
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.ecs",
		DetailType:   "ECS Task State Change",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{taskArn},
	}
}

// BuildCloudWatchEvent creates a CloudWatch EventBridge event
func BuildCloudWatchEvent(alarmName string, newState string, reason string, metricName string) CustomEvent {
	detail := map[string]interface{}{
		"alarmName": alarmName,
		"newState":  newState,
		"reason":    reason,
		"state": map[string]interface{}{
			"value":     newState,
			"reason":    reason,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
		"previousState": map[string]interface{}{
			"value":     "OK",
			"reason":    "Threshold Crossed",
			"timestamp": time.Now().Add(-5*time.Minute).UTC().Format(time.RFC3339),
		},
		"configuration": map[string]interface{}{
			"metricName": metricName,
			"namespace":  "AWS/EC2",
		},
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.cloudwatch",
		DetailType:   "CloudWatch Alarm State Change",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{fmt.Sprintf("arn:aws:cloudwatch:us-east-1:123456789012:alarm:%s", alarmName)},
	}
}

// BuildAPIGatewayEvent creates an API Gateway EventBridge event
func BuildAPIGatewayEvent(apiID string, stage string, requestID string, status string) CustomEvent {
	detail := map[string]interface{}{
		"apiId":     apiID,
		"stage":     stage,
		"requestId": requestID,
		"status":    status,
		"responseLength": 1024,
		"requestTime": time.Now().UTC().Format(time.RFC3339),
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.apigateway",
		DetailType:   "API Gateway Execution",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{fmt.Sprintf("arn:aws:apigateway:us-east-1:123456789012:restapi/%s", apiID)},
	}
}

// BuildSNSEvent creates an SNS EventBridge event
func BuildSNSEvent(topicArn string, subject string, message string) CustomEvent {
	detail := map[string]interface{}{
		"TopicArn": topicArn,
		"Subject":  subject,
		"Message":  message,
		"MessageId": "12345678-1234-1234-1234-123456789012",
		"Timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	detailJSON, _ := json.Marshal(detail)
	
	return CustomEvent{
		Source:       "aws.sns",
		DetailType:   "SNS Message",
		Detail:       string(detailJSON),
		EventBusName: DefaultEventBusName,
		Resources:    []string{topicArn},
	}
}

// BuildScheduledEventWithCron creates a scheduled EventBridge event with cron expression
func BuildScheduledEventWithCron(cronExpression string, input map[string]interface{}) RuleConfig {
	return RuleConfig{
		Name:               fmt.Sprintf("scheduled-rule-%d", time.Now().Unix()),
		Description:        fmt.Sprintf("Scheduled rule with cron: %s", cronExpression),
		ScheduleExpression: fmt.Sprintf("cron(%s)", cronExpression),
		State:              RuleStateEnabled,
		EventBusName:       DefaultEventBusName,
	}
}

// BuildScheduledEventWithRate creates a scheduled EventBridge event with rate expression
func BuildScheduledEventWithRate(rateValue int, rateUnit string) RuleConfig {
	return RuleConfig{
		Name:               fmt.Sprintf("rate-rule-%d", time.Now().Unix()),
		Description:        fmt.Sprintf("Rate rule: every %d %s", rateValue, rateUnit),
		ScheduleExpression: fmt.Sprintf("rate(%d %s)", rateValue, rateUnit),
		State:              RuleStateEnabled,
		EventBusName:       DefaultEventBusName,
	}
}