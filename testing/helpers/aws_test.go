package helpers

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/stretchr/testify/assert"
)

// Mock AWS clients for testing
type MockDynamoDBClient struct {
	DescribeTableFunc func(ctx context.Context, params *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error)
}

func (m *MockDynamoDBClient) DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {
	if m.DescribeTableFunc != nil {
		return m.DescribeTableFunc(ctx, params)
	}
	return nil, errors.New("DescribeTable not mocked")
}

type MockLambdaClient struct {
	GetFunctionFunc func(ctx context.Context, params *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error)
}

func (m *MockLambdaClient) GetFunction(ctx context.Context, params *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error) {
	if m.GetFunctionFunc != nil {
		return m.GetFunctionFunc(ctx, params)
	}
	return nil, errors.New("GetFunction not mocked")
}

// RED: Test that ValidateAWSResourceName validates correct resource names
func TestValidateAWSResourceName_ShouldAcceptValidNames(t *testing.T) {
	// Arrange
	validNames := []string{
		"test-resource",
		"my-lambda-function-123",
		"table_name",
		"resource.with.dots",
		"UPPERCASE-resource",
		"resource123",
	}
	
	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			// Act
			result := ValidateAWSResourceName(name)
			
			// Assert
			assert.True(t, result, "Valid AWS resource name should be accepted: %s", name)
		})
	}
}

// RED: Test that ValidateAWSResourceName rejects invalid resource names
func TestValidateAWSResourceName_ShouldRejectInvalidNames(t *testing.T) {
	// Arrange
	invalidNames := []string{
		"",                    // Empty string
		"resource with spaces",  // Spaces
		"resource@with@symbols", // Invalid symbols
		"resource#hash",       // Hash symbol
		"resource%percent",    // Percent symbol
	}
	
	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			// Act
			result := ValidateAWSResourceName(name)
			
			// Assert
			assert.False(t, result, "Invalid AWS resource name should be rejected: %s", name)
		})
	}
}

// RED: Test that ValidateAWSArn validates correct ARN format
func TestValidateAWSArn_ShouldAcceptValidArns(t *testing.T) {
	// Arrange
	validArns := []string{
		"arn:aws:dynamodb:us-west-2:123456789012:table/my-table",
		"arn:aws:lambda:us-east-1:123456789012:function:my-function",
		"arn:aws:s3:::my-bucket/my-object",
		"arn:aws:iam::123456789012:role/my-role",
		"arn:aws:sns:us-east-1:123456789012:my-topic",
	}
	
	for _, arn := range validArns {
		t.Run(arn, func(t *testing.T) {
			// Act
			result := ValidateAWSArn(arn)
			
			// Assert
			assert.True(t, result, "Valid AWS ARN should be accepted: %s", arn)
		})
	}
}

// RED: Test that ValidateAWSArn rejects invalid ARN format
func TestValidateAWSArn_ShouldRejectInvalidArns(t *testing.T) {
	// Arrange
	invalidArns := []string{
		"",                                    // Empty string
		"not-an-arn",                         // Not ARN format
		"arn:aws:dynamodb",                   // Incomplete ARN
		"arn:aws:dynamodb:us-west-2",         // Missing parts
		"invalid:aws:dynamodb:us-west-2:123456789012:table/my-table", // Invalid partition
	}
	
	for _, arn := range invalidArns {
		t.Run(arn, func(t *testing.T) {
			// Act
			result := ValidateAWSArn(arn)
			
			// Assert
			assert.False(t, result, "Invalid AWS ARN should be rejected: %s", arn)
		})
	}
}

// RED: Test that AssertDynamoDBTableExists validates table existence
func TestAssertDynamoDBTableExists_ShouldSucceedWhenTableExists(t *testing.T) {
	// Arrange
	tableName := "test-table"
	mockClient := &MockDynamoDBClient{
		DescribeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {
			assert.Equal(t, tableName, *params.TableName, "Should request correct table name")
			return &dynamodb.DescribeTableOutput{}, nil
		},
	}
	
	// Act
	err := AssertDynamoDBTableExists(context.Background(), mockClient, tableName)
	
	// Assert
	assert.NoError(t, err, "AssertDynamoDBTableExists should succeed when table exists")
}

// RED: Test that AssertDynamoDBTableExists fails when table doesn't exist
func TestAssertDynamoDBTableExists_ShouldFailWhenTableDoesNotExist(t *testing.T) {
	// Arrange
	tableName := "nonexistent-table"
	mockClient := &MockDynamoDBClient{
		DescribeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {
			return nil, errors.New("ResourceNotFoundException")
		},
	}
	
	// Act
	err := AssertDynamoDBTableExists(context.Background(), mockClient, tableName)
	
	// Assert
	assert.Error(t, err, "AssertDynamoDBTableExists should fail when table does not exist")
	assert.Contains(t, err.Error(), "ResourceNotFoundException", "Error should contain AWS exception")
}

// RED: Test that AssertLambdaFunctionExists validates function existence
func TestAssertLambdaFunctionExists_ShouldSucceedWhenFunctionExists(t *testing.T) {
	// Arrange
	functionName := "test-function"
	mockClient := &MockLambdaClient{
		GetFunctionFunc: func(ctx context.Context, params *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error) {
			assert.Equal(t, functionName, *params.FunctionName, "Should request correct function name")
			return &lambda.GetFunctionOutput{}, nil
		},
	}
	
	// Act
	err := AssertLambdaFunctionExists(context.Background(), mockClient, functionName)
	
	// Assert
	assert.NoError(t, err, "AssertLambdaFunctionExists should succeed when function exists")
}

// RED: Test that AssertLambdaFunctionExists fails when function doesn't exist
func TestAssertLambdaFunctionExists_ShouldFailWhenFunctionDoesNotExist(t *testing.T) {
	// Arrange
	functionName := "nonexistent-function"
	mockClient := &MockLambdaClient{
		GetFunctionFunc: func(ctx context.Context, params *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error) {
			return nil, errors.New("ResourceNotFoundException")
		},
	}
	
	// Act
	err := AssertLambdaFunctionExists(context.Background(), mockClient, functionName)
	
	// Assert
	assert.Error(t, err, "AssertLambdaFunctionExists should fail when function does not exist")
	assert.Contains(t, err.Error(), "ResourceNotFoundException", "Error should contain AWS exception")
}

// RED: Test that AssertDynamoDBTableExistsE validates table existence with retries
func TestAssertDynamoDBTableExistsE_ShouldRetryOnTransientFailures(t *testing.T) {
	// Arrange
	tableName := "test-table"
	callCount := 0
	mockClient := &MockDynamoDBClient{
		DescribeTableFunc: func(ctx context.Context, params *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {
			callCount++
			if callCount < 3 {
				return nil, errors.New("ThrottlingException")
			}
			return &dynamodb.DescribeTableOutput{}, nil
		},
	}
	
	// Act
	err := AssertDynamoDBTableExistsE(context.Background(), mockClient, tableName, 3)
	
	// Assert
	assert.NoError(t, err, "AssertDynamoDBTableExistsE should succeed after retries")
	assert.Equal(t, 3, callCount, "Should make 3 attempts")
}

// RED: Test that AssertLambdaFunctionExistsE validates function existence with retries
func TestAssertLambdaFunctionExistsE_ShouldRetryOnTransientFailures(t *testing.T) {
	// Arrange
	functionName := "test-function"
	callCount := 0
	mockClient := &MockLambdaClient{
		GetFunctionFunc: func(ctx context.Context, params *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error) {
			callCount++
			if callCount < 2 {
				return nil, errors.New("TooManyRequestsException")
			}
			return &lambda.GetFunctionOutput{}, nil
		},
	}
	
	// Act
	err := AssertLambdaFunctionExistsE(context.Background(), mockClient, functionName, 3)
	
	// Assert
	assert.NoError(t, err, "AssertLambdaFunctionExistsE should succeed after retries")
	assert.Equal(t, 2, callCount, "Should make 2 attempts")
}

// RED: Test that GetAWSResourceType extracts resource type from ARN
func TestGetAWSResourceType_ShouldExtractResourceTypeFromArn(t *testing.T) {
	// Arrange
	testCases := []struct {
		arn          string
		expectedType string
	}{
		{"arn:aws:dynamodb:us-west-2:123456789012:table/my-table", "table"},
		{"arn:aws:lambda:us-east-1:123456789012:function:my-function", "function"},
		{"arn:aws:s3:::my-bucket", ""},
		{"arn:aws:iam::123456789012:role/my-role", "role"},
		{"arn:aws:sns:us-east-1:123456789012:my-topic", ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.arn, func(t *testing.T) {
			// Act
			result := GetAWSResourceType(tc.arn)
			
			// Assert
			assert.Equal(t, tc.expectedType, result, "Should extract correct resource type from ARN")
		})
	}
}

// RED: Test that GetAWSResourceName extracts resource name from ARN
func TestGetAWSResourceName_ShouldExtractResourceNameFromArn(t *testing.T) {
	// Arrange
	testCases := []struct {
		arn          string
		expectedName string
	}{
		{"arn:aws:dynamodb:us-west-2:123456789012:table/my-table", "my-table"},
		{"arn:aws:lambda:us-east-1:123456789012:function:my-function", "my-function"},
		{"arn:aws:s3:::my-bucket", "my-bucket"},
		{"arn:aws:iam::123456789012:role/my-role", "my-role"},
		{"arn:aws:sns:us-east-1:123456789012:my-topic", "my-topic"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.arn, func(t *testing.T) {
			// Act
			result := GetAWSResourceName(tc.arn)
			
			// Assert
			assert.Equal(t, tc.expectedName, result, "Should extract correct resource name from ARN")
		})
	}
}