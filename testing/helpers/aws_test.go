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

// Use the existing MockTestingT from assertions_test.go
// Note: The existing MockTestingT doesn't track Helper() calls, so we'll skip those assertions

// RED: Test that GetRandomRegion calls Helper and uses Terratest pattern
func TestGetRandomRegion_ShouldCallHelperAndFollowTerratestPattern(t *testing.T) {
	// This test verifies the function structure without requiring AWS credentials
	// In a real environment, it would use aws.GetRandomRegion from Terratest
	t.Skip("Skipping integration test - requires real AWS credentials for Terratest aws.GetRandomRegion")
	
	// Arrange
	mockT := &MockTestingT{}
	
	// Act (would call GetRandomRegion(mockT) in real test)
	// region := GetRandomRegion(mockT)
	
	// Assert
	// Note: We can't test Helper() calls with existing MockTestingT, but the function should call it
	// assert.NotEmpty(t, region, "Should return a valid AWS region")
}

// RED: Test that GetRandomRegionExcluding calls Helper and excludes regions
func TestGetRandomRegionExcluding_ShouldCallHelperAndExcludeRegions(t *testing.T) {
	t.Skip("Skipping integration test - requires real AWS credentials for Terratest aws.GetRandomRegion")
	
	// Arrange
	mockT := &MockTestingT{}
	excludeRegions := []string{"us-east-1", "us-west-2"}
	
	// Act (would call GetRandomRegionExcluding(mockT, excludeRegions) in real test)
	// region := GetRandomRegionExcluding(mockT, excludeRegions)
	
	// Assert
	// Note: We can't test Helper() calls with existing MockTestingT, but the function should call it
	// assert.NotContains(t, excludeRegions, region, "Should not return excluded regions")
}

// RED: Test that GetAccountId calls Helper and handles errors appropriately
func TestGetAccountId_ShouldCallHelperAndHandleErrors(t *testing.T) {
	t.Skip("Skipping integration test - requires real AWS credentials for Terratest aws.GetAccountIdE")
	
	// Arrange
	mockT := &MockTestingT{}
	
	// Act (would call GetAccountId(mockT) in real test)
	// accountId := GetAccountId(mockT)
	
	// Assert
	// Note: We can't test Helper() calls with existing MockTestingT, but the function should call it
	// If GetAccountIdE fails, it should call Errorf and FailNow
	// assert.NotEmpty(t, accountId, "Should return valid account ID")
	// assert.Len(t, accountId, 12, "AWS account ID should be 12 digits")
}

// RED: Test that GetAccountIdE follows Terratest error handling pattern
func TestGetAccountIdE_ShouldReturnErrorForInvalidCredentials(t *testing.T) {
	t.Skip("Skipping integration test - requires real AWS credentials for Terratest aws.GetAccountIdE")
	
	// Arrange
	mockT := &MockTestingT{}
	
	// Act (would call GetAccountIdE(mockT) in real test)
	// accountId, err := GetAccountIdE(mockT)
	
	// Assert
	// In a real test with invalid credentials:
	// assert.Error(t, err, "Should return error with invalid credentials")
	// assert.Empty(t, accountId, "Should return empty account ID on error")
	// Note: The function should not fail the test, just return error
}

// RED: Test that GetAvailabilityZones calls Helper and returns zones
func TestGetAvailabilityZones_ShouldCallHelperAndReturnZones(t *testing.T) {
	t.Skip("Skipping integration test - requires real AWS credentials for Terratest aws.GetAvailabilityZones")
	
	// Arrange
	mockT := &MockTestingT{}
	region := "us-east-1"
	
	// Act (would call GetAvailabilityZones(mockT, region) in real test)
	// zones := GetAvailabilityZones(mockT, region)
	
	// Assert
	// Note: We can't test Helper() calls with existing MockTestingT, but the function should call it
	// assert.NotEmpty(t, zones, "Should return availability zones")
	// assert.Greater(t, len(zones), 0, "Should have at least one availability zone")
}

// RED: Test that GetAvailabilityZonesE returns error for invalid region
func TestGetAvailabilityZonesE_ShouldReturnErrorForInvalidRegion(t *testing.T) {
	t.Skip("Skipping integration test - requires real AWS credentials for Terratest aws.GetAvailabilityZonesE")
	
	// Arrange
	mockT := &MockTestingT{}
	invalidRegion := "invalid-region"
	
	// Act (would call GetAvailabilityZonesE(mockT, invalidRegion) in real test)
	// zones, err := GetAvailabilityZonesE(mockT, invalidRegion)
	
	// Assert
	// assert.Error(t, err, "Should return error for invalid region")
	// assert.Empty(t, zones, "Should return empty zones on error")
	// assert.False(t, mockT.failCalled, "Should not fail the test, just return error")
}

// RED: Test that CreateRandomResourceName creates unique resource names
func TestCreateRandomResourceName_ShouldCreateUniqueResourceNames(t *testing.T) {
	// Arrange
	prefix := "test-resource"
	
	// Act
	name1 := CreateRandomResourceName(prefix)
	name2 := CreateRandomResourceName(prefix)
	name3 := CreateRandomResourceName(prefix)
	
	// Assert
	assert.NotEmpty(t, name1, "Should return non-empty name")
	assert.NotEmpty(t, name2, "Should return non-empty name")
	assert.NotEmpty(t, name3, "Should return non-empty name")
	
	assert.Contains(t, name1, prefix, "Should contain prefix")
	assert.Contains(t, name2, prefix, "Should contain prefix")
	assert.Contains(t, name3, prefix, "Should contain prefix")
	
	// All names should be unique
	assert.NotEqual(t, name1, name2, "Names should be unique")
	assert.NotEqual(t, name1, name3, "Names should be unique")
	assert.NotEqual(t, name2, name3, "Names should be unique")
}

// RED: Test that CreateRandomResourceName handles different prefix patterns
func TestCreateRandomResourceName_ShouldHandleDifferentPrefixes(t *testing.T) {
	// Arrange
	testCases := []struct {
		name   string
		prefix string
	}{
		{"empty prefix", ""},
		{"simple prefix", "test"},
		{"hyphenated prefix", "test-resource"},
		{"underscored prefix", "test_resource"},
		{"numeric prefix", "test123"},
		{"mixed case", "TestResource"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			resourceName := CreateRandomResourceName(tc.prefix)
			
			// Assert
			assert.NotEmpty(t, resourceName, "Should return non-empty name")
			if tc.prefix != "" {
				assert.Contains(t, resourceName, tc.prefix, "Should contain prefix")
			}
			assert.Greater(t, len(resourceName), len(tc.prefix), "Should be longer than prefix")
		})
	}
}

// RED: Test that CreateRandomAWSResourceName creates valid AWS names
func TestCreateRandomAWSResourceName_ShouldCreateValidAWSNames(t *testing.T) {
	// Arrange
	prefix := "aws-test"
	
	// Act
	name1 := CreateRandomAWSResourceName(prefix)
	name2 := CreateRandomAWSResourceName(prefix)
	
	// Assert
	assert.NotEmpty(t, name1, "Should return non-empty name")
	assert.NotEmpty(t, name2, "Should return non-empty name")
	assert.NotEqual(t, name1, name2, "Should create unique names")
	
	assert.Contains(t, name1, prefix, "Should contain prefix")
	assert.Contains(t, name2, prefix, "Should contain prefix")
	
	// Should follow AWS naming patterns
	assert.Regexp(t, "^[a-zA-Z0-9-_]+$", name1, "Should match AWS naming pattern")
	assert.Regexp(t, "^[a-zA-Z0-9-_]+$", name2, "Should match AWS naming pattern")
	
	// Should pass AWS validation
	assert.True(t, ValidateAWSResourceName(name1), "Should pass AWS validation")
	assert.True(t, ValidateAWSResourceName(name2), "Should pass AWS validation")
}

// RED: Test that CreateRandomAWSResourceName handles empty prefix
func TestCreateRandomAWSResourceName_ShouldHandleEmptyPrefix(t *testing.T) {
	// Arrange
	prefix := ""
	
	// Act
	resourceName := CreateRandomAWSResourceName(prefix)
	
	// Assert
	assert.NotEmpty(t, resourceName, "Should return non-empty name even with empty prefix")
	assert.Contains(t, resourceName, "test", "Should use default prefix when empty")
	assert.True(t, ValidateAWSResourceName(resourceName), "Should create valid AWS name")
}

// RED: Test that CreateRandomAWSResourceName validates generated names
func TestCreateRandomAWSResourceName_ShouldValidateGeneratedNames(t *testing.T) {
	// Arrange
	prefixes := []string{"test", "aws-resource", "lambda-func", "ddb-table"}
	
	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			// Act
			resourceName := CreateRandomAWSResourceName(prefix)
			
			// Assert
			assert.NotEmpty(t, resourceName, "Should return non-empty name")
			assert.True(t, ValidateAWSResourceName(resourceName), "Generated name should be valid")
			assert.Contains(t, resourceName, prefix, "Should contain prefix")
		})
	}
}

// RED: Test resource name generation follows consistent patterns  
func TestResourceNameGeneration_ShouldFollowConsistentPatterns(t *testing.T) {
	// This test verifies both functions follow consistent patterns without AWS calls
	
	t.Run("should generate predictable length ranges", func(t *testing.T) {
		prefix := "consistent-test"
		
		// Generate multiple names
		randomNames := make([]string, 10)
		awsNames := make([]string, 10)
		
		for i := 0; i < 10; i++ {
			randomNames[i] = CreateRandomResourceName(prefix)
			awsNames[i] = CreateRandomAWSResourceName(prefix)
		}
		
		// All names should be in reasonable length range
		for _, name := range randomNames {
			assert.GreaterOrEqual(t, len(name), len(prefix)+1, "Random name should be longer than prefix")
			assert.LessOrEqual(t, len(name), len(prefix)+20, "Random name should not be excessively long")
		}
		
		for _, name := range awsNames {
			assert.GreaterOrEqual(t, len(name), len(prefix)+1, "AWS name should be longer than prefix")
			assert.LessOrEqual(t, len(name), len(prefix)+20, "AWS name should not be excessively long")
		}
	})
	
	t.Run("should use different random suffixes", func(t *testing.T) {
		prefix := "uniqueness-test"
		names := make([]string, 20)
		
		// Generate many names to test uniqueness
		for i := 0; i < 20; i++ {
			names[i] = CreateRandomResourceName(prefix)
		}
		
		// Check all names are unique
		nameSet := make(map[string]bool)
		for _, name := range names {
			assert.False(t, nameSet[name], "Name should be unique: %s", name)
			nameSet[name] = true
		}
	})
}