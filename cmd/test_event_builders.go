package main

import (
	"encoding/json"
	"fmt"
	"log"
	"vasdeference/lambda"
)

func main() {
	fmt.Println("Testing Event Builders...")
	
	// Test S3 Event Builder
	fmt.Println("\n=== Testing S3 Event Builder ===")
	s3Event := lambda.BuildS3Event("test-bucket", "test-key.txt", "")
	var s3 lambda.S3Event
	if err := json.Unmarshal([]byte(s3Event), &s3); err != nil {
		log.Printf("S3 Event parsing failed: %v", err)
	} else {
		fmt.Printf("✓ S3 Event created successfully - Bucket: %s, Key: %s\n", 
			s3.Records[0].S3.Bucket.Name, s3.Records[0].S3.Object.Key)
	}
	
	// Test DynamoDB Event Builder
	fmt.Println("\n=== Testing DynamoDB Event Builder ===")
	keys := map[string]interface{}{
		"id": map[string]interface{}{"S": "test-id-123"},
	}
	dynamoEvent := lambda.BuildDynamoDBEvent("test-table", "INSERT", keys)
	var dynamo lambda.DynamoDBEvent
	if err := json.Unmarshal([]byte(dynamoEvent), &dynamo); err != nil {
		log.Printf("DynamoDB Event parsing failed: %v", err)
	} else {
		fmt.Printf("✓ DynamoDB Event created successfully - Event: %s\n", 
			dynamo.Records[0].EventName)
	}
	
	// Test SQS Event Builder
	fmt.Println("\n=== Testing SQS Event Builder ===")
	sqsEvent := lambda.BuildSQSEvent("https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", "Test message")
	var sqs lambda.SQSEvent
	if err := json.Unmarshal([]byte(sqsEvent), &sqs); err != nil {
		log.Printf("SQS Event parsing failed: %v", err)
	} else {
		fmt.Printf("✓ SQS Event created successfully - Message: %s\n", 
			sqs.Records[0].Body)
	}
	
	// Test SNS Event Builder
	fmt.Println("\n=== Testing SNS Event Builder ===")
	snsEvent := lambda.BuildSNSEvent("arn:aws:sns:us-east-1:123456789012:test-topic", "Test SNS message")
	var sns lambda.SNSEvent
	if err := json.Unmarshal([]byte(snsEvent), &sns); err != nil {
		log.Printf("SNS Event parsing failed: %v", err)
	} else {
		fmt.Printf("✓ SNS Event created successfully - Message: %s\n", 
			sns.Records[0].Sns.Message)
	}
	
	// Test API Gateway Event Builder
	fmt.Println("\n=== Testing API Gateway Event Builder ===")
	apiEvent := lambda.BuildAPIGatewayEvent("POST", "/api/test", `{"test": true}`)
	var api lambda.APIGatewayEvent
	if err := json.Unmarshal([]byte(apiEvent), &api); err != nil {
		log.Printf("API Gateway Event parsing failed: %v", err)
	} else {
		fmt.Printf("✓ API Gateway Event created successfully - Method: %s, Path: %s\n", 
			api.HttpMethod, api.Path)
	}
	
	// Test CloudWatch Event Builder (NEW)
	fmt.Println("\n=== Testing CloudWatch Event Builder ===")
	detail := map[string]interface{}{
		"state": "ALARM",
		"description": "Test CloudWatch alarm",
	}
	cwEvent := lambda.BuildCloudWatchEvent("aws.cloudwatch", "CloudWatch Alarm State Change", detail)
	var cw lambda.CloudWatchEvent
	if err := json.Unmarshal([]byte(cwEvent), &cw); err != nil {
		log.Printf("CloudWatch Event parsing failed: %v", err)
	} else {
		fmt.Printf("✓ CloudWatch Event created successfully - Source: %s, DetailType: %s\n", 
			cw.Source, cw.DetailType)
	}
	
	// Test Kinesis Event Builder (NEW)
	fmt.Println("\n=== Testing Kinesis Event Builder ===")
	kinesisEvent := lambda.BuildKinesisEvent("test-stream", "partition-key-1", "Test Kinesis data")
	var kinesis lambda.KinesisEvent
	if err := json.Unmarshal([]byte(kinesisEvent), &kinesis); err != nil {
		log.Printf("Kinesis Event parsing failed: %v", err)
	} else {
		fmt.Printf("✓ Kinesis Event created successfully - Stream: %s, PartitionKey: %s\n", 
			kinesis.Records[0].EventSourceARN, kinesis.Records[0].Kinesis.PartitionKey)
	}
	
	fmt.Println("\n=== All Event Builders Test Completed ===")
}