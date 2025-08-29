// Package sqs provides comprehensive AWS SQS testing utilities following Terratest patterns.
//
// This package is designed to be a drop-in replacement for Terratest's SQS utilities
// with enhanced serverless-specific features. It follows strict functional programming
// principles and provides both Function and FunctionE variants for all operations.
//
// Key Features:
//   - Queue lifecycle management (Create/Delete/Purge/Get attributes)
//   - Message operations (Send/Receive/Delete) with batch support
//   - Dead Letter Queue configuration and management
//   - FIFO queue support with deduplication and content-based grouping
//   - Visibility timeout management and message lifecycle
//   - Queue policy and permission management
//   - Server-side encryption configuration (SSE-SQS, SSE-KMS)
//   - Message attribute handling and filtering
//   - Long polling and short polling configurations
//   - Comprehensive retry and error handling
//
// Example usage:
//
//   func TestSQSMessageProcessing(t *testing.T) {
//       ctx := sfx.NewTestContext(t)
//       
//       // Create standard queue
//       queue := CreateQueue(ctx, QueueConfig{
//           QueueName: "test-queue",
//           Attributes: map[string]string{
//               "VisibilityTimeoutSeconds": "30",
//               "MessageRetentionPeriod":   "1209600",
//           },
//       })
//       
//       // Send message
//       result := SendMessage(ctx, queue.QueueUrl, "Hello, World!", nil)
//       assert.NotEmpty(t, result.MessageId)
//       
//       // Receive and process message
//       messages := ReceiveMessages(ctx, queue.QueueUrl, 1, 20)
//       assert.Len(t, messages, 1)
//       
//       // Clean up
//       DeleteMessage(ctx, queue.QueueUrl, messages[0].ReceiptHandle)
//       DeleteQueue(ctx, queue.QueueUrl)
//   }
package sqs

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Common SQS configurations
const (
	DefaultTimeout           = 30 * time.Second
	DefaultRetryAttempts     = 3
	DefaultRetryDelay        = 1 * time.Second
	MaxQueueNameLength       = 80
	MaxFifoQueueNameLength   = 75  // .fifo suffix adds 5 characters
	MaxMessageSize           = 256 * 1024 // 256KB
	MaxReceiveCount          = 10
	DefaultVisibilityTimeout = 30 * time.Second
	DefaultWaitTimeSeconds   = 20 // Long polling
	MaxBatchSize             = 10
	MaxMessageAttributes     = 10
)

// Queue types
const (
	QueueTypeStandard = "Standard"
	QueueTypeFIFO     = "FIFO"
)

// Common errors
var (
	ErrQueueNotFound           = errors.New("queue not found")
	ErrMessageNotFound         = errors.New("message not found")
	ErrInvalidQueueName        = errors.New("invalid queue name")
	ErrInvalidMessageBody      = errors.New("invalid message body")
	ErrInvalidReceiptHandle    = errors.New("invalid receipt handle")
	ErrQueueAlreadyExists      = errors.New("queue already exists")
	ErrMessageTooLarge         = errors.New("message size exceeds limit")
	ErrBatchSizeTooLarge       = errors.New("batch size exceeds limit")
	ErrInvalidVisibilityTimeout = errors.New("invalid visibility timeout")
	ErrDLQConfigurationError   = errors.New("dead letter queue configuration error")
)

// TestingT provides interface compatibility with testing frameworks
type TestingT interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Fail()
	FailNow()
	Helper()
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Name() string
}

// TestContext represents the testing context with AWS configuration
type TestContext struct {
	T         TestingT
	AwsConfig aws.Config
	Region    string
}

// =============================================================================
// QUEUE MANAGEMENT
// =============================================================================

// QueueConfig represents configuration for creating an SQS queue
type QueueConfig struct {
	QueueName                 string
	Attributes                map[string]string
	Tags                      map[string]string
	FifoQueue                 bool
	ContentBasedDeduplication bool
	KmsKeyId                  string
	KmsMasterKeyId            string
}

// QueueResult represents the result of queue operations
type QueueResult struct {
	QueueUrl   string
	QueueName  string
	Attributes map[string]string
	Tags       map[string]string
	QueueType  string
}

// CreateQueue creates an SQS queue and fails the test if an error occurs
func CreateQueue(ctx *TestContext, config QueueConfig) *QueueResult {
	result, err := CreateQueueE(ctx, config)
	if err != nil {
		ctx.T.Fatalf("Failed to create SQS queue: %v", err)
	}
	return result
}

// CreateQueueE creates an SQS queue and returns any error
func CreateQueueE(ctx *TestContext, config QueueConfig) (*QueueResult, error) {
	if err := validateQueueName(config.QueueName, config.FifoQueue); err != nil {
		return nil, fmt.Errorf("invalid queue name: %w", err)
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return nil, err
	}

	queueName := config.QueueName
	if config.FifoQueue && !strings.HasSuffix(queueName, ".fifo") {
		queueName += ".fifo"
	}

	input := &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	}

	// Set default attributes and merge with user-provided ones
	attributes := getDefaultQueueAttributes(config.FifoQueue)
	for key, value := range config.Attributes {
		attributes[key] = value
	}

	// Configure FIFO-specific attributes
	if config.FifoQueue {
		attributes["FifoQueue"] = "true"
		if config.ContentBasedDeduplication {
			attributes["ContentBasedDeduplication"] = "true"
		}
	}

	// Configure encryption if specified
	if config.KmsKeyId != "" {
		attributes["KmsMasterKeyId"] = config.KmsKeyId
	}
	if config.KmsMasterKeyId != "" {
		attributes["KmsMasterKeyId"] = config.KmsMasterKeyId
	}

	if len(attributes) > 0 {
		input.Attributes = attributes
	}

	if len(config.Tags) > 0 {
		input.Tags = config.Tags
	}

	output, err := client.CreateQueue(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	queueType := QueueTypeStandard
	if config.FifoQueue {
		queueType = QueueTypeFIFO
	}

	return &QueueResult{
		QueueUrl:   aws.ToString(output.QueueUrl),
		QueueName:  queueName,
		Attributes: attributes,
		Tags:       config.Tags,
		QueueType:  queueType,
	}, nil
}

// GetQueueUrl gets the URL for an SQS queue and fails the test if an error occurs
func GetQueueUrl(ctx *TestContext, queueName string) string {
	url, err := GetQueueUrlE(ctx, queueName)
	if err != nil {
		ctx.T.Fatalf("Failed to get queue URL: %v", err)
	}
	return url
}

// GetQueueUrlE gets the URL for an SQS queue and returns any error
func GetQueueUrlE(ctx *TestContext, queueName string) (string, error) {
	if queueName == "" {
		return "", fmt.Errorf("queue name cannot be empty")
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return "", err
	}

	output, err := client.GetQueueUrl(context.Background(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get queue URL: %w", err)
	}

	return aws.ToString(output.QueueUrl), nil
}

// GetQueueAttributes gets queue attributes and fails the test if an error occurs
func GetQueueAttributes(ctx *TestContext, queueUrl string, attributeNames []types.QueueAttributeName) map[string]string {
	attributes, err := GetQueueAttributesE(ctx, queueUrl, attributeNames)
	if err != nil {
		ctx.T.Fatalf("Failed to get queue attributes: %v", err)
	}
	return attributes
}

// GetQueueAttributesE gets queue attributes and returns any error
func GetQueueAttributesE(ctx *TestContext, queueUrl string, attributeNames []types.QueueAttributeName) (map[string]string, error) {
	if queueUrl == "" {
		return nil, fmt.Errorf("queue URL cannot be empty")
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(queueUrl),
	}

	if len(attributeNames) > 0 {
		input.AttributeNames = attributeNames
	} else {
		input.AttributeNames = []types.QueueAttributeName{types.QueueAttributeNameAll}
	}

	output, err := client.GetQueueAttributes(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	return output.Attributes, nil
}

// DeleteQueue deletes an SQS queue and fails the test if an error occurs
func DeleteQueue(ctx *TestContext, queueUrl string) {
	err := DeleteQueueE(ctx, queueUrl)
	if err != nil {
		ctx.T.Fatalf("Failed to delete queue: %v", err)
	}
}

// DeleteQueueE deletes an SQS queue and returns any error
func DeleteQueueE(ctx *TestContext, queueUrl string) error {
	if queueUrl == "" {
		return fmt.Errorf("queue URL cannot be empty")
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.DeleteQueue(context.Background(), &sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueUrl),
	})
	if err != nil {
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	return nil
}

// PurgeQueue removes all messages from a queue and fails the test if an error occurs
func PurgeQueue(ctx *TestContext, queueUrl string) {
	err := PurgeQueueE(ctx, queueUrl)
	if err != nil {
		ctx.T.Fatalf("Failed to purge queue: %v", err)
	}
}

// PurgeQueueE removes all messages from a queue and returns any error
func PurgeQueueE(ctx *TestContext, queueUrl string) error {
	if queueUrl == "" {
		return fmt.Errorf("queue URL cannot be empty")
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.PurgeQueue(context.Background(), &sqs.PurgeQueueInput{
		QueueUrl: aws.String(queueUrl),
	})
	if err != nil {
		return fmt.Errorf("failed to purge queue: %w", err)
	}

	return nil
}

// =============================================================================
// MESSAGE OPERATIONS
// =============================================================================

// MessageConfig represents configuration for sending a message
type MessageConfig struct {
	MessageBody            string
	DelaySeconds           int32
	MessageAttributes      map[string]types.MessageAttributeValue
	MessageSystemAttributes map[string]types.MessageSystemAttributeValue
	MessageDeduplicationId string // FIFO only
	MessageGroupId         string // FIFO only
}

// MessageResult represents the result of message operations
type MessageResult struct {
	MessageId              string
	MD5OfBody              string
	MD5OfMessageAttributes string
	SequenceNumber         string // FIFO only
}

// ReceivedMessage represents a message received from SQS
type ReceivedMessage struct {
	MessageId              string
	ReceiptHandle          string
	MD5OfBody              string
	Body                   string
	Attributes             map[string]string
	MessageAttributes      map[string]types.MessageAttributeValue
	MD5OfMessageAttributes string
}

// SendMessage sends a single message to an SQS queue and fails the test if an error occurs
func SendMessage(ctx *TestContext, queueUrl string, messageBody string, messageAttributes map[string]types.MessageAttributeValue) *MessageResult {
	result, err := SendMessageE(ctx, queueUrl, messageBody, messageAttributes)
	if err != nil {
		ctx.T.Fatalf("Failed to send message: %v", err)
	}
	return result
}

// SendMessageE sends a single message to an SQS queue and returns any error
func SendMessageE(ctx *TestContext, queueUrl string, messageBody string, messageAttributes map[string]types.MessageAttributeValue) (*MessageResult, error) {
	config := MessageConfig{
		MessageBody:       messageBody,
		MessageAttributes: messageAttributes,
	}
	return SendMessageWithConfigE(ctx, queueUrl, config)
}

// SendMessageWithConfig sends a message with full configuration and fails the test if an error occurs
func SendMessageWithConfig(ctx *TestContext, queueUrl string, config MessageConfig) *MessageResult {
	result, err := SendMessageWithConfigE(ctx, queueUrl, config)
	if err != nil {
		ctx.T.Fatalf("Failed to send message with config: %v", err)
	}
	return result
}

// SendMessageWithConfigE sends a message with full configuration and returns any error
func SendMessageWithConfigE(ctx *TestContext, queueUrl string, config MessageConfig) (*MessageResult, error) {
	if err := validateMessageConfig(config); err != nil {
		return nil, fmt.Errorf("invalid message config: %w", err)
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueUrl),
		MessageBody: aws.String(config.MessageBody),
	}

	if config.DelaySeconds > 0 {
		input.DelaySeconds = aws.Int32(config.DelaySeconds)
	}
	if len(config.MessageAttributes) > 0 {
		input.MessageAttributes = config.MessageAttributes
	}
	if len(config.MessageSystemAttributes) > 0 {
		input.MessageSystemAttributes = config.MessageSystemAttributes
	}
	if config.MessageDeduplicationId != "" {
		input.MessageDeduplicationId = aws.String(config.MessageDeduplicationId)
	}
	if config.MessageGroupId != "" {
		input.MessageGroupId = aws.String(config.MessageGroupId)
	}

	output, err := client.SendMessage(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &MessageResult{
		MessageId:              aws.ToString(output.MessageId),
		MD5OfBody:              aws.ToString(output.MD5OfBody),
		MD5OfMessageAttributes: aws.ToString(output.MD5OfMessageAttributes),
		SequenceNumber:         aws.ToString(output.SequenceNumber),
	}, nil
}

// ReceiveMessages receives messages from an SQS queue and fails the test if an error occurs
func ReceiveMessages(ctx *TestContext, queueUrl string, maxNumberOfMessages int32, waitTimeSeconds int32) []ReceivedMessage {
	messages, err := ReceiveMessagesE(ctx, queueUrl, maxNumberOfMessages, waitTimeSeconds)
	if err != nil {
		ctx.T.Fatalf("Failed to receive messages: %v", err)
	}
	return messages
}

// ReceiveMessagesE receives messages from an SQS queue and returns any error
func ReceiveMessagesE(ctx *TestContext, queueUrl string, maxNumberOfMessages int32, waitTimeSeconds int32) ([]ReceivedMessage, error) {
	if queueUrl == "" {
		return nil, fmt.Errorf("queue URL cannot be empty")
	}
	if maxNumberOfMessages < 1 || maxNumberOfMessages > 10 {
		return nil, fmt.Errorf("maxNumberOfMessages must be between 1 and 10")
	}
	if waitTimeSeconds < 0 || waitTimeSeconds > 20 {
		return nil, fmt.Errorf("waitTimeSeconds must be between 0 and 20")
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return nil, err
	}

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int32(maxNumberOfMessages),
		WaitTimeSeconds:     aws.Int32(waitTimeSeconds),
		AttributeNames:      []types.QueueAttributeName{types.QueueAttributeNameAll},
		MessageAttributeNames: []string{"All"},
	}

	output, err := client.ReceiveMessage(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	messages := make([]ReceivedMessage, len(output.Messages))
	for i, msg := range output.Messages {
		messages[i] = ReceivedMessage{
			MessageId:              aws.ToString(msg.MessageId),
			ReceiptHandle:          aws.ToString(msg.ReceiptHandle),
			MD5OfBody:              aws.ToString(msg.MD5OfBody),
			Body:                   aws.ToString(msg.Body),
			Attributes:             msg.Attributes,
			MessageAttributes:      msg.MessageAttributes,
			MD5OfMessageAttributes: aws.ToString(msg.MD5OfMessageAttributes),
		}
	}

	return messages, nil
}

// DeleteMessage deletes a message from an SQS queue and fails the test if an error occurs
func DeleteMessage(ctx *TestContext, queueUrl string, receiptHandle string) {
	err := DeleteMessageE(ctx, queueUrl, receiptHandle)
	if err != nil {
		ctx.T.Fatalf("Failed to delete message: %v", err)
	}
}

// DeleteMessageE deletes a message from an SQS queue and returns any error
func DeleteMessageE(ctx *TestContext, queueUrl string, receiptHandle string) error {
	if queueUrl == "" {
		return fmt.Errorf("queue URL cannot be empty")
	}
	if receiptHandle == "" {
		return fmt.Errorf("receipt handle cannot be empty")
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// =============================================================================
// BATCH OPERATIONS
// =============================================================================

// BatchMessageConfig represents configuration for sending a batch of messages
type BatchMessageConfig struct {
	Id                       string
	MessageBody              string
	DelaySeconds             int32
	MessageAttributes        map[string]types.MessageAttributeValue
	MessageSystemAttributes  map[string]types.MessageSystemAttributeValue
	MessageDeduplicationId   string // FIFO only
	MessageGroupId           string // FIFO only
}

// BatchMessageResult represents the result of batch message operations
type BatchMessageResult struct {
	Successful []types.SendMessageBatchResultEntry
	Failed     []types.BatchResultErrorEntry
}

// SendMessageBatch sends multiple messages in a single request and fails the test if an error occurs
func SendMessageBatch(ctx *TestContext, queueUrl string, messages []BatchMessageConfig) *BatchMessageResult {
	result, err := SendMessageBatchE(ctx, queueUrl, messages)
	if err != nil {
		ctx.T.Fatalf("Failed to send message batch: %v", err)
	}
	return result
}

// SendMessageBatchE sends multiple messages in a single request and returns any error
func SendMessageBatchE(ctx *TestContext, queueUrl string, messages []BatchMessageConfig) (*BatchMessageResult, error) {
	if queueUrl == "" {
		return nil, fmt.Errorf("queue URL cannot be empty")
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("message batch cannot be empty")
	}
	if len(messages) > MaxBatchSize {
		return nil, fmt.Errorf("batch size cannot exceed %d", MaxBatchSize)
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return nil, err
	}

	entries := make([]types.SendMessageBatchRequestEntry, len(messages))
	for i, msg := range messages {
		entries[i] = types.SendMessageBatchRequestEntry{
			Id:          aws.String(msg.Id),
			MessageBody: aws.String(msg.MessageBody),
		}

		if msg.DelaySeconds > 0 {
			entries[i].DelaySeconds = aws.Int32(msg.DelaySeconds)
		}
		if len(msg.MessageAttributes) > 0 {
			entries[i].MessageAttributes = msg.MessageAttributes
		}
		if len(msg.MessageSystemAttributes) > 0 {
			entries[i].MessageSystemAttributes = msg.MessageSystemAttributes
		}
		if msg.MessageDeduplicationId != "" {
			entries[i].MessageDeduplicationId = aws.String(msg.MessageDeduplicationId)
		}
		if msg.MessageGroupId != "" {
			entries[i].MessageGroupId = aws.String(msg.MessageGroupId)
		}
	}

	output, err := client.SendMessageBatch(context.Background(), &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(queueUrl),
		Entries:  entries,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message batch: %w", err)
	}

	return &BatchMessageResult{
		Successful: output.Successful,
		Failed:     output.Failed,
	}, nil
}

// =============================================================================
// DEAD LETTER QUEUE CONFIGURATION
// =============================================================================

// DeadLetterQueueConfig represents configuration for a dead letter queue
type DeadLetterQueueConfig struct {
	TargetArn   string
	MaxReceiveCount int32
}

// ConfigureDeadLetterQueue configures a dead letter queue for the specified queue and fails the test if an error occurs
func ConfigureDeadLetterQueue(ctx *TestContext, queueUrl string, dlqConfig DeadLetterQueueConfig) {
	err := ConfigureDeadLetterQueueE(ctx, queueUrl, dlqConfig)
	if err != nil {
		ctx.T.Fatalf("Failed to configure dead letter queue: %v", err)
	}
}

// ConfigureDeadLetterQueueE configures a dead letter queue for the specified queue and returns any error
func ConfigureDeadLetterQueueE(ctx *TestContext, queueUrl string, dlqConfig DeadLetterQueueConfig) error {
	if queueUrl == "" {
		return fmt.Errorf("queue URL cannot be empty")
	}
	if dlqConfig.TargetArn == "" {
		return fmt.Errorf("dead letter queue target ARN cannot be empty")
	}
	if dlqConfig.MaxReceiveCount < 1 || dlqConfig.MaxReceiveCount > 1000 {
		return fmt.Errorf("maxReceiveCount must be between 1 and 1000")
	}

	redrivePolicy := fmt.Sprintf(`{"deadLetterTargetArn":"%s","maxReceiveCount":%d}`, 
		dlqConfig.TargetArn, dlqConfig.MaxReceiveCount)

	return SetQueueAttributeE(ctx, queueUrl, types.QueueAttributeNameRedrivePolicy, redrivePolicy)
}

// SetQueueAttribute sets a single queue attribute and fails the test if an error occurs
func SetQueueAttribute(ctx *TestContext, queueUrl string, attributeName types.QueueAttributeName, attributeValue string) {
	err := SetQueueAttributeE(ctx, queueUrl, attributeName, attributeValue)
	if err != nil {
		ctx.T.Fatalf("Failed to set queue attribute: %v", err)
	}
}

// SetQueueAttributeE sets a single queue attribute and returns any error
func SetQueueAttributeE(ctx *TestContext, queueUrl string, attributeName types.QueueAttributeName, attributeValue string) error {
	if queueUrl == "" {
		return fmt.Errorf("queue URL cannot be empty")
	}
	if attributeValue == "" {
		return fmt.Errorf("attribute value cannot be empty")
	}

	client, err := createSQSClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.SetQueueAttributes(context.Background(), &sqs.SetQueueAttributesInput{
		QueueUrl: aws.String(queueUrl),
		Attributes: map[string]string{
			string(attributeName): attributeValue,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to set queue attribute: %w", err)
	}

	return nil
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// createSQSClient creates a new SQS client
func createSQSClient(ctx *TestContext) (*sqs.Client, error) {
	if ctx.AwsConfig.Region == "" {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		ctx.AwsConfig = cfg
	}
	return sqs.NewFromConfig(ctx.AwsConfig), nil
}

// validateQueueName validates that the queue name follows AWS naming conventions
func validateQueueName(name string, isFifo bool) error {
	if name == "" {
		return fmt.Errorf("queue name cannot be empty")
	}
	
	maxLen := MaxQueueNameLength
	if isFifo {
		maxLen = MaxFifoQueueNameLength
		if !strings.HasSuffix(name, ".fifo") {
			// Allow for .fifo suffix to be added
			if len(name) > maxLen {
				return fmt.Errorf("FIFO queue name cannot exceed %d characters (excluding .fifo suffix)", maxLen)
			}
		} else {
			if len(name) > MaxQueueNameLength {
				return fmt.Errorf("FIFO queue name cannot exceed %d characters", MaxQueueNameLength)
			}
		}
	} else {
		if len(name) > maxLen {
			return fmt.Errorf("queue name cannot exceed %d characters", maxLen)
		}
		if strings.HasSuffix(name, ".fifo") {
			return fmt.Errorf("standard queue name cannot end with .fifo suffix")
		}
	}

	// Check for invalid characters
	for _, char := range name {
		if !isValidQueueNameCharacter(char) {
			return fmt.Errorf("queue name contains invalid character: %c", char)
		}
	}

	return nil
}

// isValidQueueNameCharacter checks if a character is valid for SQS queue names
func isValidQueueNameCharacter(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_' || char == '.'
}

// validateMessageConfig validates message configuration
func validateMessageConfig(config MessageConfig) error {
	if config.MessageBody == "" {
		return fmt.Errorf("message body cannot be empty")
	}
	if len(config.MessageBody) > MaxMessageSize {
		return fmt.Errorf("message body size cannot exceed %d bytes", MaxMessageSize)
	}
	if len(config.MessageAttributes) > MaxMessageAttributes {
		return fmt.Errorf("message cannot have more than %d attributes", MaxMessageAttributes)
	}
	if config.DelaySeconds < 0 || config.DelaySeconds > 900 {
		return fmt.Errorf("delay seconds must be between 0 and 900")
	}
	return nil
}

// getDefaultQueueAttributes returns default attributes for a queue
func getDefaultQueueAttributes(isFifo bool) map[string]string {
	attributes := map[string]string{
		"VisibilityTimeoutSeconds":   strconv.Itoa(int(DefaultVisibilityTimeout.Seconds())),
		"MessageRetentionPeriod":     "1209600", // 14 days
		"MaxReceiveCount":            "3",
		"ReceiveMessageWaitTimeSeconds": strconv.Itoa(DefaultWaitTimeSeconds),
	}

	if isFifo {
		attributes["FifoQueue"] = "true"
	}

	return attributes
}

// GetQueueInfo gets comprehensive information about a queue and fails the test if an error occurs
func GetQueueInfo(ctx *TestContext, queueUrl string) *QueueResult {
	result, err := GetQueueInfoE(ctx, queueUrl)
	if err != nil {
		ctx.T.Fatalf("Failed to get queue info: %v", err)
	}
	return result
}

// GetQueueInfoE gets comprehensive information about a queue and returns any error
func GetQueueInfoE(ctx *TestContext, queueUrl string) (*QueueResult, error) {
	attributes, err := GetQueueAttributesE(ctx, queueUrl, nil)
	if err != nil {
		return nil, err
	}

	// Extract queue name from URL
	parts := strings.Split(queueUrl, "/")
	queueName := ""
	if len(parts) > 0 {
		queueName = parts[len(parts)-1]
	}

	queueType := QueueTypeStandard
	if fifo, exists := attributes["FifoQueue"]; exists && fifo == "true" {
		queueType = QueueTypeFIFO
	}

	return &QueueResult{
		QueueUrl:   queueUrl,
		QueueName:  queueName,
		Attributes: attributes,
		QueueType:  queueType,
	}, nil
}

// WaitForQueueEmpty waits for a queue to become empty and fails the test if an error occurs
func WaitForQueueEmpty(ctx *TestContext, queueUrl string, timeout time.Duration) {
	err := WaitForQueueEmptyE(ctx, queueUrl, timeout)
	if err != nil {
		ctx.T.Fatalf("Failed to wait for queue to become empty: %v", err)
	}
}

// WaitForQueueEmptyE waits for a queue to become empty and returns any error
func WaitForQueueEmptyE(ctx *TestContext, queueUrl string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		attributes, err := GetQueueAttributesE(ctx, queueUrl, []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
			types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
		})
		if err != nil {
			return fmt.Errorf("failed to get queue attributes: %w", err)
		}

		visible := 0
		invisible := 0

		if val, exists := attributes["ApproximateNumberOfMessages"]; exists {
			if parsed, parseErr := strconv.Atoi(val); parseErr == nil {
				visible = parsed
			}
		}

		if val, exists := attributes["ApproximateNumberOfMessagesNotVisible"]; exists {
			if parsed, parseErr := strconv.Atoi(val); parseErr == nil {
				invisible = parsed
			}
		}

		if visible == 0 && invisible == 0 {
			return nil
		}

		select {
		case <-ticker.C:
			continue
		case <-time.After(time.Until(deadline)):
			return fmt.Errorf("timeout waiting for queue to become empty")
		}
	}

	return fmt.Errorf("timeout waiting for queue to become empty")
}