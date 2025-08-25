package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Test UpdateFunctionCodeE with zip file using mocks
func TestUpdateFunctionCodeE_WithZipFile_UsingMocks_ShouldUpdateCode(t *testing.T) {
	// Given: Mock setup with successful code update
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	updatedFunction := &lambda.UpdateFunctionCodeOutput{
		FunctionName: aws.String("test-function"),
		CodeSha256:   aws.String("new-sha256-hash"),
		Version:      aws.String("$LATEST"),
		LastModified: aws.String("2023-01-01T12:00:00.000+0000"),
	}
	
	mockClient.UpdateFunctionCodeResponse = updatedFunction
	mockClient.UpdateFunctionCodeError = nil
	
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String("test-function"),
		ZipFile:      []byte("new-zip-content"),
	}
	
	// When: UpdateFunctionCodeE is called
	result, err := UpdateFunctionCodeE(ctx, input)
	
	// Then: Should return updated function configuration
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-function", result.FunctionName)
	assert.Equal(t, "new-sha256-hash", result.CodeSha256)
	assert.Contains(t, mockClient.UpdateFunctionCodeCalls, "test-function")
}

// RED: Test UpdateFunctionCodeE with S3 bucket using mocks
func TestUpdateFunctionCodeE_WithS3Bucket_UsingMocks_ShouldUpdateCode(t *testing.T) {
	// Given: Mock setup with S3 code update
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	updatedFunction := &lambda.UpdateFunctionCodeOutput{
		FunctionName: aws.String("s3-function"),
		CodeSha256:   aws.String("s3-sha256-hash"),
		Version:      aws.String("$LATEST"),
		LastModified: aws.String("2023-01-01T12:00:00.000+0000"),
	}
	
	mockClient.UpdateFunctionCodeResponse = updatedFunction
	mockClient.UpdateFunctionCodeError = nil
	
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String("s3-function"),
		S3Bucket:     aws.String("my-deployment-bucket"),
		S3Key:        aws.String("lambda-functions/my-function.zip"),
		S3ObjectVersion: aws.String("version-123"),
	}
	
	// When: UpdateFunctionCodeE is called
	result, err := UpdateFunctionCodeE(ctx, input)
	
	// Then: Should return updated function configuration
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "s3-function", result.FunctionName)
	assert.Equal(t, "s3-sha256-hash", result.CodeSha256)
	assert.Contains(t, mockClient.UpdateFunctionCodeCalls, "s3-function")
}

// RED: Test UpdateFunctionCodeE with container image using mocks
func TestUpdateFunctionCodeE_WithContainerImage_UsingMocks_ShouldUpdateCode(t *testing.T) {
	// Given: Mock setup with container image update
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	updatedFunction := &lambda.UpdateFunctionCodeOutput{
		FunctionName: aws.String("container-function"),
		CodeSha256:   aws.String("container-sha256-hash"),
		Version:      aws.String("$LATEST"),
		LastModified: aws.String("2023-01-01T12:00:00.000+0000"),
		PackageType:  types.PackageTypeImage,
	}
	
	mockClient.UpdateFunctionCodeResponse = updatedFunction
	mockClient.UpdateFunctionCodeError = nil
	
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String("container-function"),
		ImageUri:     aws.String("123456789012.dkr.ecr.us-east-1.amazonaws.com/my-func:latest"),
	}
	
	// When: UpdateFunctionCodeE is called
	result, err := UpdateFunctionCodeE(ctx, input)
	
	// Then: Should return updated function configuration
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "container-function", result.FunctionName)
	assert.Equal(t, "container-sha256-hash", result.CodeSha256)
	assert.Equal(t, types.PackageTypeImage, result.PackageType)
	assert.Contains(t, mockClient.UpdateFunctionCodeCalls, "container-function")
}

// RED: Test UpdateFunctionCodeE with invalid function name using mocks
func TestUpdateFunctionCodeE_WithInvalidFunctionName_UsingMocks_ShouldReturnValidationError(t *testing.T) {
	// Given: Mock setup
	ctx, _ := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(""), // Invalid empty name
		ZipFile:      []byte("zip-content"),
	}
	
	// When: UpdateFunctionCodeE is called with invalid function name
	result, err := UpdateFunctionCodeE(ctx, input)
	
	// Then: Should return validation error without making API call
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid function name")
}

// RED: Test UpdateFunctionCodeE with API error using mocks
func TestUpdateFunctionCodeE_WithAPIError_UsingMocks_ShouldReturnError(t *testing.T) {
	// Given: Mock setup with API error
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	mockClient.UpdateFunctionCodeResponse = nil
	mockClient.UpdateFunctionCodeError = createResourceNotFoundError("non-existing-function")
	
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String("non-existing-function"),
		ZipFile:      []byte("zip-content"),
	}
	
	// When: UpdateFunctionCodeE is called
	result, err := UpdateFunctionCodeE(ctx, input)
	
	// Then: Should return error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ResourceNotFoundException")
	assert.Contains(t, mockClient.UpdateFunctionCodeCalls, "non-existing-function")
}

// RED: Test UpdateFunctionCodeE with publish version using mocks
func TestUpdateFunctionCodeE_WithPublishVersion_UsingMocks_ShouldUpdateAndPublish(t *testing.T) {
	// Given: Mock setup with publish version
	ctx, mockClient := setupMockFunctionTest()
	defer teardownMockAssertionTest()
	
	updatedFunction := &lambda.UpdateFunctionCodeOutput{
		FunctionName: aws.String("publish-function"),
		CodeSha256:   aws.String("publish-sha256-hash"),
		Version:      aws.String("1"),
		LastModified: aws.String("2023-01-01T12:00:00.000+0000"),
	}
	
	mockClient.UpdateFunctionCodeResponse = updatedFunction
	mockClient.UpdateFunctionCodeError = nil
	
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String("publish-function"),
		ZipFile:      []byte("new-zip-content"),
		Publish:      true,
	}
	
	// When: UpdateFunctionCodeE is called
	result, err := UpdateFunctionCodeE(ctx, input)
	
	// Then: Should return updated function with new version
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "publish-function", result.FunctionName)
	assert.Equal(t, "1", result.Version)
	assert.Contains(t, mockClient.UpdateFunctionCodeCalls, "publish-function")
}