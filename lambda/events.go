package lambda

import (
	"encoding/json"
	"fmt"
	"time"
)

// Common AWS event patterns and builders for Lambda testing

// S3Event represents an S3 event that triggers Lambda
type S3Event struct {
	Records []S3Record `json:"Records"`
}

// S3Record represents a single S3 event record
type S3Record struct {
	EventVersion      string                 `json:"eventVersion"`
	EventSource       string                 `json:"eventSource"`
	EventName         string                 `json:"eventName"`
	EventTime         string                 `json:"eventTime"`
	UserIdentity      map[string]interface{} `json:"userIdentity"`
	RequestParameters map[string]interface{} `json:"requestParameters"`
	ResponseElements  map[string]interface{} `json:"responseElements"`
	S3                S3EventData            `json:"s3"`
}

// S3EventData represents the S3-specific data in an event
type S3EventData struct {
	SchemaVersion   string    `json:"s3SchemaVersion"`
	ConfigurationId string    `json:"configurationId"`
	Bucket          S3Bucket  `json:"bucket"`
	Object          S3Object  `json:"object"`
}

// S3Bucket represents the S3 bucket information in an event
type S3Bucket struct {
	Name          string                 `json:"name"`
	OwnerIdentity map[string]interface{} `json:"ownerIdentity"`
	Arn           string                 `json:"arn"`
}

// S3Object represents the S3 object information in an event
type S3Object struct {
	Key       string `json:"key"`
	Size      int64  `json:"size"`
	ETag      string `json:"eTag"`
	Sequencer string `json:"sequencer"`
}

// DynamoDBEvent represents a DynamoDB stream event
type DynamoDBEvent struct {
	Records []DynamoDBRecord `json:"Records"`
}

// DynamoDBRecord represents a single DynamoDB stream record
type DynamoDBRecord struct {
	EventID      string                 `json:"eventID"`
	EventName    string                 `json:"eventName"`
	EventVersion string                 `json:"eventVersion"`
	EventSource  string                 `json:"eventSource"`
	AwsRegion    string                 `json:"awsRegion"`
	Dynamodb     DynamoDBStreamRecord   `json:"dynamodb"`
	UserIdentity map[string]interface{} `json:"userIdentity,omitempty"`
}

// DynamoDBStreamRecord represents the DynamoDB stream record data
type DynamoDBStreamRecord struct {
	ApproximateCreationDateTime int64                      `json:"ApproximateCreationDateTime"`
	Keys                        map[string]interface{}     `json:"Keys"`
	NewImage                    map[string]interface{}     `json:"NewImage,omitempty"`
	OldImage                    map[string]interface{}     `json:"OldImage,omitempty"`
	SequenceNumber              string                     `json:"SequenceNumber"`
	SizeBytes                   int64                      `json:"SizeBytes"`
	StreamViewType              string                     `json:"StreamViewType"`
}

// SQSEvent represents an SQS event
type SQSEvent struct {
	Records []SQSRecord `json:"Records"`
}

// SQSRecord represents a single SQS message record
type SQSRecord struct {
	MessageId         string                 `json:"messageId"`
	ReceiptHandle     string                 `json:"receiptHandle"`
	Body              string                 `json:"body"`
	Attributes        map[string]interface{} `json:"attributes"`
	MessageAttributes map[string]interface{} `json:"messageAttributes"`
	Md5OfBody         string                 `json:"md5OfBody"`
	EventSource       string                 `json:"eventSource"`
	EventSourceARN    string                 `json:"eventSourceARN"`
	AwsRegion         string                 `json:"awsRegion"`
}

// SNSEvent represents an SNS event
type SNSEvent struct {
	Records []SNSRecord `json:"Records"`
}

// SNSRecord represents a single SNS record
type SNSRecord struct {
	EventVersion         string                 `json:"EventVersion"`
	EventSubscriptionArn string                 `json:"EventSubscriptionArn"`
	EventSource          string                 `json:"EventSource"`
	Sns                  SNSMessage             `json:"Sns"`
}

// SNSMessage represents the SNS message data
type SNSMessage struct {
	SignatureVersion  string                 `json:"SignatureVersion"`
	Timestamp         string                 `json:"Timestamp"`
	Signature         string                 `json:"Signature"`
	SigningCertUrl    string                 `json:"SigningCertUrl"`
	MessageId         string                 `json:"MessageId"`
	Message           string                 `json:"Message"`
	MessageAttributes map[string]interface{} `json:"MessageAttributes"`
	Type              string                 `json:"Type"`
	UnsubscribeUrl    string                 `json:"UnsubscribeUrl"`
	TopicArn          string                 `json:"TopicArn"`
	Subject           string                 `json:"Subject"`
}

// APIGatewayEvent represents an API Gateway event
type APIGatewayEvent struct {
	Resource                        string                         `json:"resource"`
	Path                            string                         `json:"path"`
	HttpMethod                      string                         `json:"httpMethod"`
	Headers                         map[string]string              `json:"headers"`
	MultiValueHeaders               map[string][]string            `json:"multiValueHeaders"`
	QueryStringParameters           map[string]string              `json:"queryStringParameters"`
	MultiValueQueryStringParameters map[string][]string            `json:"multiValueQueryStringParameters"`
	PathParameters                  map[string]string              `json:"pathParameters"`
	StageVariables                  map[string]string              `json:"stageVariables"`
	RequestContext                  APIGatewayRequestContext       `json:"requestContext"`
	Body                            string                         `json:"body"`
	IsBase64Encoded                 bool                           `json:"isBase64Encoded"`
}

// APIGatewayRequestContext represents the API Gateway request context
type APIGatewayRequestContext struct {
	AccountId    string                        `json:"accountId"`
	ApiId        string                        `json:"apiId"`
	HttpMethod   string                        `json:"httpMethod"`
	RequestId    string                        `json:"requestId"`
	ResourceId   string                        `json:"resourceId"`
	ResourcePath string                        `json:"resourcePath"`
	Stage        string                        `json:"stage"`
	Identity     APIGatewayRequestIdentity     `json:"identity"`
}

// APIGatewayRequestIdentity represents the API Gateway request identity
type APIGatewayRequestIdentity struct {
	SourceIp  string `json:"sourceIp"`
	UserAgent string `json:"userAgent"`
}

// Event builder functions

// BuildS3Event creates an S3 event for Lambda testing.
// This is the non-error returning version that follows Terratest patterns.
func BuildS3Event(bucketName string, objectKey string, eventName string) string {
	event, err := BuildS3EventE(bucketName, objectKey, eventName)
	if err != nil {
		panic(fmt.Sprintf("Failed to build S3 event: %v", err))
	}
	return event
}

// BuildS3EventE creates an S3 event for Lambda testing.
// This is the error returning version that follows Terratest patterns.
func BuildS3EventE(bucketName string, objectKey string, eventName string) (string, error) {
	if bucketName == "" {
		return "", fmt.Errorf("bucket name cannot be empty")
	}
	if objectKey == "" {
		return "", fmt.Errorf("object key cannot be empty")
	}
	if eventName == "" {
		eventName = "s3:ObjectCreated:Put"
	}
	
	event := S3Event{
		Records: []S3Record{
			{
				EventVersion: "2.1",
				EventSource:  "aws:s3",
				EventName:    eventName,
				EventTime:    time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
				UserIdentity: map[string]interface{}{
					"principalId": "AWS:AIDACKCEVSQ6C2EXAMPLE",
				},
				RequestParameters: map[string]interface{}{
					"sourceIPAddress": "203.0.113.1",
				},
				ResponseElements: map[string]interface{}{
					"x-amz-request-id": "C3D13FE58DE4C810",
					"x-amz-id-2":       "FMyUVURIY8/IgAtTv8xRjskZQpcIZ9KG4V5Wp6S7S/JRWeUWerMUE5JgHvANOjpD",
				},
				S3: S3EventData{
					SchemaVersion:   "1.0",
					ConfigurationId: "testConfigRule",
					Bucket: S3Bucket{
						Name: bucketName,
						OwnerIdentity: map[string]interface{}{
							"principalId": "A3NL1KOZZKExample",
						},
						Arn: fmt.Sprintf("arn:aws:s3:::%s", bucketName),
					},
					Object: S3Object{
						Key:       objectKey,
						Size:      1024,
						ETag:      "0123456789abcdef0123456789abcdef",
						Sequencer: "0A1B2C3D4E5F678901",
					},
				},
			},
		},
	}
	
	return MarshalPayloadE(event)
}

// BuildDynamoDBEvent creates a DynamoDB stream event for Lambda testing.
// This is the non-error returning version that follows Terratest patterns.
func BuildDynamoDBEvent(tableName string, eventName string, keys map[string]interface{}) string {
	event, err := BuildDynamoDBEventE(tableName, eventName, keys)
	if err != nil {
		panic(fmt.Sprintf("Failed to build DynamoDB event: %v", err))
	}
	return event
}

// BuildDynamoDBEventE creates a DynamoDB stream event for Lambda testing.
// This is the error returning version that follows Terratest patterns.
func BuildDynamoDBEventE(tableName string, eventName string, keys map[string]interface{}) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}
	if eventName == "" {
		eventName = "INSERT"
	}
	if keys == nil {
		keys = map[string]interface{}{
			"id": map[string]interface{}{"S": "example-id"},
		}
	}
	
	event := DynamoDBEvent{
		Records: []DynamoDBRecord{
			{
				EventID:      "c4ca4238a0b923820dcc509a6f75849b",
				EventName:    eventName,
				EventVersion: "1.1",
				EventSource:  "aws:dynamodb",
				AwsRegion:    "us-east-1",
				Dynamodb: DynamoDBStreamRecord{
					ApproximateCreationDateTime: time.Now().Unix(),
					Keys:                        keys,
					SequenceNumber:              "111",
					SizeBytes:                   26,
					StreamViewType:              "KEYS_ONLY",
				},
			},
		},
	}
	
	return MarshalPayloadE(event)
}

// BuildSQSEvent creates an SQS event for Lambda testing.
// This is the non-error returning version that follows Terratest patterns.
func BuildSQSEvent(queueUrl string, messageBody string) string {
	event, err := BuildSQSEventE(queueUrl, messageBody)
	if err != nil {
		panic(fmt.Sprintf("Failed to build SQS event: %v", err))
	}
	return event
}

// BuildSQSEventE creates an SQS event for Lambda testing.
// This is the error returning version that follows Terratest patterns.
func BuildSQSEventE(queueUrl string, messageBody string) (string, error) {
	if queueUrl == "" {
		return "", fmt.Errorf("queue URL cannot be empty")
	}
	if messageBody == "" {
		messageBody = "Hello from SQS!"
	}
	
	event := SQSEvent{
		Records: []SQSRecord{
			{
				MessageId:     "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
				ReceiptHandle: "MessageReceiptHandle",
				Body:          messageBody,
				Attributes: map[string]interface{}{
					"ApproximateReceiveCount":          "1",
					"SentTimestamp":                    "1523232000000",
					"SenderId":                         "123456789012",
					"ApproximateFirstReceiveTimestamp": "1523232000001",
				},
				MessageAttributes: make(map[string]interface{}),
				Md5OfBody:         "7b270e59b47ff90a553787216d55d909",
				EventSource:       "aws:sqs",
				EventSourceARN:    queueUrl,
				AwsRegion:         "us-east-1",
			},
		},
	}
	
	return MarshalPayloadE(event)
}

// BuildSNSEvent creates an SNS event for Lambda testing.
// This is the non-error returning version that follows Terratest patterns.
func BuildSNSEvent(topicArn string, message string) string {
	event, err := BuildSNSEventE(topicArn, message)
	if err != nil {
		panic(fmt.Sprintf("Failed to build SNS event: %v", err))
	}
	return event
}

// BuildSNSEventE creates an SNS event for Lambda testing.
// This is the error returning version that follows Terratest patterns.
func BuildSNSEventE(topicArn string, message string) (string, error) {
	if topicArn == "" {
		return "", fmt.Errorf("topic ARN cannot be empty")
	}
	if message == "" {
		message = "Hello from SNS!"
	}
	
	event := SNSEvent{
		Records: []SNSRecord{
			{
				EventVersion:         "1.0",
				EventSubscriptionArn: topicArn + ":c9bd9c72-7a5c-11e7-916d-3b9c1e4b9c4d",
				EventSource:          "aws:sns",
				Sns: SNSMessage{
					SignatureVersion: "1",
					Timestamp:        time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
					Signature:        "EXAMPLE",
					SigningCertUrl:   "EXAMPLE",
					MessageId:        "95df01b4-ee98-5cb9-9903-4c221d41eb5e",
					Message:          message,
					MessageAttributes: make(map[string]interface{}),
					Type:             "Notification",
					UnsubscribeUrl:   "EXAMPLE",
					TopicArn:         topicArn,
					Subject:          "TestInvoke",
				},
			},
		},
	}
	
	return MarshalPayloadE(event)
}

// BuildAPIGatewayEvent creates an API Gateway event for Lambda testing.
// This is the non-error returning version that follows Terratest patterns.
func BuildAPIGatewayEvent(httpMethod string, path string, body string) string {
	event, err := BuildAPIGatewayEventE(httpMethod, path, body)
	if err != nil {
		panic(fmt.Sprintf("Failed to build API Gateway event: %v", err))
	}
	return event
}

// BuildAPIGatewayEventE creates an API Gateway event for Lambda testing.
// This is the error returning version that follows Terratest patterns.
func BuildAPIGatewayEventE(httpMethod string, path string, body string) (string, error) {
	if httpMethod == "" {
		httpMethod = "GET"
	}
	if path == "" {
		path = "/"
	}
	
	event := APIGatewayEvent{
		Resource:   "/{proxy+}",
		Path:       path,
		HttpMethod: httpMethod,
		Headers: map[string]string{
			"Accept":                          "*/*",
			"Accept-Encoding":                 "gzip, deflate",
			"Cache-Control":                   "no-cache",
			"CloudFront-Forwarded-Proto":      "https",
			"CloudFront-Is-Desktop-Viewer":    "true",
			"CloudFront-Is-Mobile-Viewer":     "false",
			"CloudFront-Is-SmartTV-Viewer":    "false",
			"CloudFront-Is-Tablet-Viewer":     "false",
			"CloudFront-Viewer-Country":       "US",
			"Host":                            "test.execute-api.us-east-1.amazonaws.com",
			"Postman-Token":                   "9f583ef0-ed83-4a38-aef3-eb9ce3f7a57f",
			"User-Agent":                      "PostmanRuntime/2.4.5",
			"Via":                             "1.1 d98420743a69852491bbdea73f7680bd.cloudfront.net (CloudFront)",
			"X-Amz-Cf-Id":                     "pn-PWmJc6thYnZm5P0NMgOUglL1DYtl0gdeJky8tqsg8iS_sgsKD1A==",
			"X-Forwarded-For":                 "54.240.196.186, 54.182.214.83",
			"X-Forwarded-Port":                "443",
			"X-Forwarded-Proto":               "https",
		},
		MultiValueHeaders:               make(map[string][]string),
		QueryStringParameters:           make(map[string]string),
		MultiValueQueryStringParameters: make(map[string][]string),
		PathParameters:                  make(map[string]string),
		StageVariables:                  make(map[string]string),
		RequestContext: APIGatewayRequestContext{
			AccountId:    "123456789012",
			ApiId:        "1234567890",
			HttpMethod:   httpMethod,
			RequestId:    "c6af9ac6-7b61-11e6-9a41-93e8deadbeef",
			ResourceId:   "123456",
			ResourcePath: "/{proxy+}",
			Stage:        "prod",
			Identity: APIGatewayRequestIdentity{
				SourceIp:  "203.0.113.1",
				UserAgent: "PostmanRuntime/2.4.5",
			},
		},
		Body:            body,
		IsBase64Encoded: false,
	}
	
	return MarshalPayloadE(event)
}

// Utility functions for event manipulation

// AddS3EventRecord adds another S3 event record to an existing S3 event.
func AddS3EventRecord(existingEventJSON string, bucketName string, objectKey string, eventName string) string {
	result, err := AddS3EventRecordE(existingEventJSON, bucketName, objectKey, eventName)
	if err != nil {
		panic(fmt.Sprintf("Failed to add S3 event record: %v", err))
	}
	return result
}

// AddS3EventRecordE adds another S3 event record to an existing S3 event.
func AddS3EventRecordE(existingEventJSON string, bucketName string, objectKey string, eventName string) (string, error) {
	var event S3Event
	if err := json.Unmarshal([]byte(existingEventJSON), &event); err != nil {
		return "", fmt.Errorf("failed to unmarshal existing S3 event: %w", err)
	}
	
	// Create new record
	newRecordJSON, err := BuildS3EventE(bucketName, objectKey, eventName)
	if err != nil {
		return "", err
	}
	
	var newEvent S3Event
	if err := json.Unmarshal([]byte(newRecordJSON), &newEvent); err != nil {
		return "", fmt.Errorf("failed to unmarshal new S3 event: %w", err)
	}
	
	// Add the new record to the existing event
	event.Records = append(event.Records, newEvent.Records...)
	
	return MarshalPayloadE(event)
}

// ParseS3Event parses an S3 event from JSON.
func ParseS3Event(eventJSON string) S3Event {
	event, err := ParseS3EventE(eventJSON)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse S3 event: %v", err))
	}
	return event
}

// ParseS3EventE parses an S3 event from JSON.
func ParseS3EventE(eventJSON string) (S3Event, error) {
	var event S3Event
	err := json.Unmarshal([]byte(eventJSON), &event)
	if err != nil {
		return S3Event{}, fmt.Errorf("failed to parse S3 event: %w", err)
	}
	return event, nil
}

// ExtractS3BucketName extracts the bucket name from an S3 event JSON.
func ExtractS3BucketName(eventJSON string) string {
	name, err := ExtractS3BucketNameE(eventJSON)
	if err != nil {
		panic(fmt.Sprintf("Failed to extract S3 bucket name: %v", err))
	}
	return name
}

// ExtractS3BucketNameE extracts the bucket name from an S3 event JSON.
func ExtractS3BucketNameE(eventJSON string) (string, error) {
	event, err := ParseS3EventE(eventJSON)
	if err != nil {
		return "", err
	}
	
	if len(event.Records) == 0 {
		return "", fmt.Errorf("no S3 records found in event")
	}
	
	return event.Records[0].S3.Bucket.Name, nil
}

// ExtractS3ObjectKey extracts the object key from an S3 event JSON.
func ExtractS3ObjectKey(eventJSON string) string {
	key, err := ExtractS3ObjectKeyE(eventJSON)
	if err != nil {
		panic(fmt.Sprintf("Failed to extract S3 object key: %v", err))
	}
	return key
}

// ExtractS3ObjectKeyE extracts the object key from an S3 event JSON.
func ExtractS3ObjectKeyE(eventJSON string) (string, error) {
	event, err := ParseS3EventE(eventJSON)
	if err != nil {
		return "", err
	}
	
	if len(event.Records) == 0 {
		return "", fmt.Errorf("no S3 records found in event")
	}
	
	return event.Records[0].S3.Object.Key, nil
}

// ValidateEventStructure validates that an event JSON has the expected structure.
func ValidateEventStructure(eventJSON string, eventType string) []string {
	var errors []string
	
	switch eventType {
	case "s3":
		event, err := ParseS3EventE(eventJSON)
		if err != nil {
			errors = append(errors, fmt.Sprintf("invalid S3 event structure: %v", err))
		} else {
			// Additional validation for S3 structure
			if len(event.Records) == 0 {
				errors = append(errors, "S3 event must have at least one record")
			} else {
				for i, record := range event.Records {
					if record.EventSource != "aws:s3" {
						errors = append(errors, fmt.Sprintf("S3 record %d has wrong event source: %s", i, record.EventSource))
						break
					}
				}
			}
		}
	case "dynamodb":
		var event DynamoDBEvent
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			errors = append(errors, fmt.Sprintf("invalid DynamoDB event structure: %v", err))
		} else {
			// Additional validation for DynamoDB structure
			if len(event.Records) == 0 {
				errors = append(errors, "DynamoDB event must have at least one record")
			} else {
				for i, record := range event.Records {
					if record.EventSource != "aws:dynamodb" {
						errors = append(errors, fmt.Sprintf("DynamoDB record %d has wrong event source: %s", i, record.EventSource))
						break
					}
				}
			}
		}
	case "sqs":
		var event SQSEvent
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			errors = append(errors, fmt.Sprintf("invalid SQS event structure: %v", err))
		} else {
			// Additional validation for SQS structure
			if len(event.Records) == 0 {
				errors = append(errors, "SQS event must have at least one record")
			} else {
				for i, record := range event.Records {
					if record.EventSource != "aws:sqs" {
						errors = append(errors, fmt.Sprintf("SQS record %d has wrong event source: %s", i, record.EventSource))
						break
					}
				}
			}
		}
	case "sns":
		var event SNSEvent
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			errors = append(errors, fmt.Sprintf("invalid SNS event structure: %v", err))
		} else {
			// Additional validation for SNS structure
			if len(event.Records) == 0 {
				errors = append(errors, "SNS event must have at least one record")
			} else {
				for i, record := range event.Records {
					if record.EventSource != "aws:sns" {
						errors = append(errors, fmt.Sprintf("SNS record %d has wrong event source: %s", i, record.EventSource))
						break
					}
				}
			}
		}
	case "apigateway":
		var event APIGatewayEvent
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			errors = append(errors, fmt.Sprintf("invalid API Gateway event structure: %v", err))
		} else {
			// Additional validation for API Gateway structure
			if event.HttpMethod == "" {
				errors = append(errors, "API Gateway event must have an HTTP method")
			}
			if event.Path == "" {
				errors = append(errors, "API Gateway event must have a path")
			}
		}
	default:
		errors = append(errors, fmt.Sprintf("unknown event type: %s", eventType))
	}
	
	return errors
}