// Package s3 provides comprehensive AWS S3 testing utilities following Terratest patterns.
//
// This package is designed to be a drop-in replacement for Terratest's S3 utilities
// with enhanced serverless-specific features. It follows strict functional programming
// principles and provides both Function and FunctionE variants for all operations.
//
// Key Features:
//   - Bucket lifecycle management (Create/Delete/Update/List)
//   - Object operations (Put/Get/Delete/Copy/List) with metadata support
//   - Versioning and lifecycle policy management
//   - CORS configuration and access control
//   - Server-side encryption with KMS integration
//   - Presigned URL generation for secure access
//   - Multipart upload support for large files
//   - Event notification configuration
//   - Cross-region replication setup
//   - Comprehensive S3 testing utilities
//
// Example usage:
//
//   func TestS3Operations(t *testing.T) {
//       ctx := sfx.NewTestContext(t)
//       
//       // Create bucket
//       bucket := CreateBucket(ctx, BucketConfig{
//           BucketName: "test-bucket",
//           Region:     "us-east-1",
//       })
//       
//       // Upload object
//       object := PutObject(ctx, bucket.Name, "test.txt", "Hello World!")
//       
//       // Get object
//       content := GetObject(ctx, bucket.Name, "test.txt")
//       
//       // Enable versioning
//       EnableVersioning(ctx, bucket.Name)
//       
//       assert.NotEmpty(t, object.ETag)
//       assert.Equal(t, "Hello World!", content)
//   }
package s3

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"
)

// Common S3 configurations
const (
	DefaultTimeout       = 30 * time.Second
	DefaultRetryAttempts = 3
	DefaultRetryDelay    = 1 * time.Second
	DefaultPartSize      = 5 * 1024 * 1024 // 5MB
	DefaultMaxRetries    = 3
	DefaultMaxParts      = 10000
)

// Storage classes
const (
	StorageClassStandard           = types.StorageClassStandard
	StorageClassReducedRedundancy  = types.StorageClassReducedRedundancy
	StorageClassStandardIa         = types.StorageClassStandardIa
	StorageClassOnezoneIa          = types.StorageClassOnezoneIa
	StorageClassIntelligentTiering = types.StorageClassIntelligentTiering
	StorageClassGlacier            = types.StorageClassGlacier
	StorageClassDeepArchive        = types.StorageClassDeepArchive
)

// Server-side encryption
const (
	ServerSideEncryptionAES256 = types.ServerSideEncryptionAes256
	ServerSideEncryptionKMS    = types.ServerSideEncryptionAwsKms
)

// Common errors
var (
	ErrBucketNotFound      = errors.New("bucket not found")
	ErrObjectNotFound      = errors.New("object not found")
	ErrBucketExists        = errors.New("bucket already exists")
	ErrInvalidBucketName   = errors.New("invalid bucket name")
	ErrInvalidObjectKey    = errors.New("invalid object key")
	ErrAccessDenied        = errors.New("access denied")
	ErrInvalidEncryption   = errors.New("invalid encryption configuration")
	ErrVersioningDisabled  = errors.New("versioning not enabled")
	ErrReplicationFailed   = errors.New("replication configuration failed")
)

// TestContext represents the test context with AWS configuration
type TestContext interface {
	AwsConfig() aws.Config
	T() interface{ 
		Errorf(format string, args ...interface{})
		Error(args ...interface{})
		Fail()
	}
}

// Bucket configuration for creating buckets
type BucketConfig struct {
	BucketName string            `json:"bucketName"`
	Region     string            `json:"region"`
	ACL        types.BucketCannedACL `json:"acl,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// Object configuration for uploading objects
type ObjectConfig struct {
	BucketName      string                        `json:"bucketName"`
	Key             string                        `json:"key"`
	Body            io.Reader                     `json:"-"`
	ContentType     string                        `json:"contentType,omitempty"`
	StorageClass    types.StorageClass            `json:"storageClass,omitempty"`
	ServerSideEncryption types.ServerSideEncryption `json:"serverSideEncryption,omitempty"`
	KMSKeyID        string                        `json:"kmsKeyId,omitempty"`
	Metadata        map[string]string             `json:"metadata,omitempty"`
	Tags            map[string]string             `json:"tags,omitempty"`
	CacheControl    string                        `json:"cacheControl,omitempty"`
	ContentDisposition string                     `json:"contentDisposition,omitempty"`
}

// CORS configuration
type CORSConfig struct {
	BucketName   string       `json:"bucketName"`
	CORSRules    []CORSRule   `json:"corsRules"`
}

type CORSRule struct {
	AllowedHeaders []string `json:"allowedHeaders,omitempty"`
	AllowedMethods []string `json:"allowedMethods"`
	AllowedOrigins []string `json:"allowedOrigins"`
	ExposeHeaders  []string `json:"exposeHeaders,omitempty"`
	MaxAgeSeconds  int32    `json:"maxAgeSeconds,omitempty"`
}

// Lifecycle configuration
type LifecycleConfig struct {
	BucketName string           `json:"bucketName"`
	Rules      []LifecycleRule  `json:"rules"`
}

type LifecycleRule struct {
	ID                    string                    `json:"id"`
	Status                types.ExpirationStatus   `json:"status"`
	Filter                *LifecycleRuleFilter     `json:"filter,omitempty"`
	Expiration            *LifecycleExpiration     `json:"expiration,omitempty"`
	Transitions           []LifecycleTransition    `json:"transitions,omitempty"`
	NoncurrentVersionExpiration *NoncurrentVersionExpiration `json:"noncurrentVersionExpiration,omitempty"`
}

type LifecycleRuleFilter struct {
	Prefix string            `json:"prefix,omitempty"`
	Tags   map[string]string `json:"tags,omitempty"`
}

type LifecycleExpiration struct {
	Days                      int32 `json:"days,omitempty"`
	ExpiredObjectDeleteMarker bool  `json:"expiredObjectDeleteMarker,omitempty"`
}

type LifecycleTransition struct {
	Days         int32              `json:"days,omitempty"`
	Date         *time.Time         `json:"date,omitempty"`
	StorageClass types.StorageClass `json:"storageClass"`
}

type NoncurrentVersionExpiration struct {
	NoncurrentDays int32 `json:"noncurrentDays"`
}

// Results structures
type BucketResult struct {
	Name      string            `json:"name"`
	Region    string            `json:"region"`
	CreatedAt time.Time         `json:"createdAt"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type ObjectResult struct {
	BucketName   string            `json:"bucketName"`
	Key          string            `json:"key"`
	ETag         string            `json:"etag"`
	Size         int64             `json:"size"`
	LastModified time.Time         `json:"lastModified"`
	StorageClass types.StorageClass `json:"storageClass,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	VersionId    string            `json:"versionId,omitempty"`
}

type PresignedURLResult struct {
	URL       string        `json:"url"`
	ExpiresAt time.Time     `json:"expiresAt"`
	Method    string        `json:"method"`
	Duration  time.Duration `json:"duration"`
}

type BucketVersioningResult struct {
	BucketName string                     `json:"bucketName"`
	Status     types.BucketVersioningStatus `json:"status"`
}

// createS3Client creates an S3 client using shared AWS config from TestContext
func createS3Client(ctx TestContext) *s3.Client {
	return s3.NewFromConfig(ctx.AwsConfig)
}

// CreateBucket creates an S3 bucket
func CreateBucket(ctx TestContext, config BucketConfig) *BucketResult {
	result, err := CreateBucketE(ctx, config)
	require.NoError(ctx.T, err)
	return result
}

// CreateBucketE creates an S3 bucket
func CreateBucketE(ctx TestContext, config BucketConfig) (*BucketResult, error) {
	client := createS3Client(ctx)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(config.BucketName),
	}

	if config.ACL != "" {
		input.ACL = config.ACL
	}

	// Add region constraint if not us-east-1
	if config.Region != "" && config.Region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(config.Region),
		}
	}

	_, err := client.CreateBucket(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	// Add tags if provided
	if len(config.Tags) > 0 {
		err = PutBucketTaggingE(ctx, config.BucketName, config.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to add tags to bucket: %w", err)
		}
	}

	return &BucketResult{
		Name:      config.BucketName,
		Region:    config.Region,
		CreatedAt: time.Now(),
		Tags:      config.Tags,
	}, nil
}

// DeleteBucket deletes an S3 bucket
func DeleteBucket(ctx TestContext, bucketName string) {
	err := DeleteBucketE(ctx, bucketName)
	require.NoError(ctx.T, err)
}

// DeleteBucketE deletes an S3 bucket
func DeleteBucketE(ctx TestContext, bucketName string) error {
	client := createS3Client(ctx)

	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}

	_, err := client.DeleteBucket(context.TODO(), input)
	return err
}

// BucketExists checks if a bucket exists
func BucketExists(ctx TestContext, bucketName string) bool {
	exists, err := BucketExistsE(ctx, bucketName)
	require.NoError(ctx.T, err)
	return exists
}

// BucketExistsE checks if a bucket exists
func BucketExistsE(ctx TestContext, bucketName string) (bool, error) {
	client := createS3Client(ctx)

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}

	_, err := client.HeadBucket(context.TODO(), input)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// ListBuckets lists all buckets
func ListBuckets(ctx TestContext) []BucketResult {
	buckets, err := ListBucketsE(ctx)
	require.NoError(ctx.T, err)
	return buckets
}

// ListBucketsE lists all buckets
func ListBucketsE(ctx TestContext) ([]BucketResult, error) {
	client := createS3Client(ctx)

	input := &s3.ListBucketsInput{}

	output, err := client.ListBuckets(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var buckets []BucketResult
	for _, bucket := range output.Buckets {
		buckets = append(buckets, BucketResult{
			Name:      aws.ToString(bucket.Name),
			CreatedAt: aws.ToTime(bucket.CreationDate),
		})
	}

	return buckets, nil
}

// PutObject uploads an object to S3
func PutObject(ctx TestContext, bucketName, key, body string) *ObjectResult {
	result, err := PutObjectE(ctx, bucketName, key, body)
	require.NoError(ctx.T, err)
	return result
}

// PutObjectE uploads an object to S3
func PutObjectE(ctx TestContext, bucketName, key, body string) (*ObjectResult, error) {
	config := ObjectConfig{
		BucketName: bucketName,
		Key:        key,
		Body:       strings.NewReader(body),
	}
	return PutObjectWithConfigE(ctx, config)
}

// PutObjectWithConfig uploads an object to S3 with configuration
func PutObjectWithConfig(ctx TestContext, config ObjectConfig) *ObjectResult {
	result, err := PutObjectWithConfigE(ctx, config)
	require.NoError(ctx.T, err)
	return result
}

// PutObjectWithConfigE uploads an object to S3 with configuration
func PutObjectWithConfigE(ctx TestContext, config ObjectConfig) (*ObjectResult, error) {
	client := createS3Client(ctx)

	input := &s3.PutObjectInput{
		Bucket: aws.String(config.BucketName),
		Key:    aws.String(config.Key),
		Body:   config.Body,
	}

	if config.ContentType != "" {
		input.ContentType = aws.String(config.ContentType)
	}

	if config.StorageClass != "" {
		input.StorageClass = config.StorageClass
	}

	if config.ServerSideEncryption != "" {
		input.ServerSideEncryption = config.ServerSideEncryption
	}

	if config.KMSKeyID != "" {
		input.SSEKMSKeyId = aws.String(config.KMSKeyID)
	}

	if len(config.Metadata) > 0 {
		input.Metadata = config.Metadata
	}

	if config.CacheControl != "" {
		input.CacheControl = aws.String(config.CacheControl)
	}

	if config.ContentDisposition != "" {
		input.ContentDisposition = aws.String(config.ContentDisposition)
	}

	output, err := client.PutObject(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	// Add tags if provided
	if len(config.Tags) > 0 {
		err = PutObjectTaggingE(ctx, config.BucketName, config.Key, config.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to add tags to object: %w", err)
		}
	}

	return &ObjectResult{
		BucketName: config.BucketName,
		Key:        config.Key,
		ETag:       aws.ToString(output.ETag),
		VersionId:  aws.ToString(output.VersionId),
	}, nil
}

// GetObject downloads an object from S3
func GetObject(ctx TestContext, bucketName, key string) string {
	content, err := GetObjectE(ctx, bucketName, key)
	require.NoError(ctx.T, err)
	return content
}

// GetObjectE downloads an object from S3
func GetObjectE(ctx TestContext, bucketName, key string) (string, error) {
	client := createS3Client(ctx)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	output, err := client.GetObject(context.TODO(), input)
	if err != nil {
		return "", err
	}
	defer output.Body.Close()

	body, err := io.ReadAll(output.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetObjectWithVersion downloads a specific version of an object from S3
func GetObjectWithVersion(ctx TestContext, bucketName, key, versionId string) string {
	content, err := GetObjectWithVersionE(ctx, bucketName, key, versionId)
	require.NoError(ctx.T, err)
	return content
}

// GetObjectWithVersionE downloads a specific version of an object from S3
func GetObjectWithVersionE(ctx TestContext, bucketName, key, versionId string) (string, error) {
	client := createS3Client(ctx)

	input := &s3.GetObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(key),
		VersionId: aws.String(versionId),
	}

	output, err := client.GetObject(context.TODO(), input)
	if err != nil {
		return "", err
	}
	defer output.Body.Close()

	body, err := io.ReadAll(output.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// DeleteObject deletes an object from S3
func DeleteObject(ctx TestContext, bucketName, key string) {
	err := DeleteObjectE(ctx, bucketName, key)
	require.NoError(ctx.T, err)
}

// DeleteObjectE deletes an object from S3
func DeleteObjectE(ctx TestContext, bucketName, key string) error {
	client := createS3Client(ctx)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	_, err := client.DeleteObject(context.TODO(), input)
	return err
}

// ObjectExists checks if an object exists
func ObjectExists(ctx TestContext, bucketName, key string) bool {
	exists, err := ObjectExistsE(ctx, bucketName, key)
	require.NoError(ctx.T, err)
	return exists
}

// ObjectExistsE checks if an object exists
func ObjectExistsE(ctx TestContext, bucketName, key string) (bool, error) {
	client := createS3Client(ctx)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	_, err := client.HeadObject(context.TODO(), input)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// ListObjects lists objects in a bucket
func ListObjects(ctx TestContext, bucketName string) []ObjectResult {
	objects, err := ListObjectsE(ctx, bucketName)
	require.NoError(ctx.T, err)
	return objects
}

// ListObjectsE lists objects in a bucket
func ListObjectsE(ctx TestContext, bucketName string) ([]ObjectResult, error) {
	return ListObjectsWithPrefixE(ctx, bucketName, "")
}

// ListObjectsWithPrefix lists objects in a bucket with a prefix
func ListObjectsWithPrefix(ctx TestContext, bucketName, prefix string) []ObjectResult {
	objects, err := ListObjectsWithPrefixE(ctx, bucketName, prefix)
	require.NoError(ctx.T, err)
	return objects
}

// ListObjectsWithPrefixE lists objects in a bucket with a prefix
func ListObjectsWithPrefixE(ctx TestContext, bucketName, prefix string) ([]ObjectResult, error) {
	client := createS3Client(ctx)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	var allObjects []ObjectResult

	for {
		output, err := client.ListObjectsV2(context.TODO(), input)
		if err != nil {
			return nil, err
		}

		for _, object := range output.Contents {
			allObjects = append(allObjects, ObjectResult{
				BucketName:   bucketName,
				Key:          aws.ToString(object.Key),
				ETag:         aws.ToString(object.ETag),
				Size:         aws.ToInt64(object.Size),
				LastModified: aws.ToTime(object.LastModified),
				StorageClass: object.StorageClass,
			})
		}

		if !output.IsTruncated {
			break
		}

		input.ContinuationToken = output.NextContinuationToken
	}

	return allObjects, nil
}

// CopyObject copies an object within S3
func CopyObject(ctx TestContext, sourceBucket, sourceKey, destBucket, destKey string) *ObjectResult {
	result, err := CopyObjectE(ctx, sourceBucket, sourceKey, destBucket, destKey)
	require.NoError(ctx.T, err)
	return result
}

// CopyObjectE copies an object within S3
func CopyObjectE(ctx TestContext, sourceBucket, sourceKey, destBucket, destKey string) (*ObjectResult, error) {
	client := createS3Client(ctx)

	copySource := fmt.Sprintf("%s/%s", sourceBucket, sourceKey)

	input := &s3.CopyObjectInput{
		Bucket:     aws.String(destBucket),
		Key:        aws.String(destKey),
		CopySource: aws.String(copySource),
	}

	output, err := client.CopyObject(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &ObjectResult{
		BucketName: destBucket,
		Key:        destKey,
		ETag:       aws.ToString(output.CopyObjectResult.ETag),
		VersionId:  aws.ToString(output.VersionId),
	}, nil
}

// EnableVersioning enables versioning on a bucket
func EnableVersioning(ctx TestContext, bucketName string) *BucketVersioningResult {
	result, err := EnableVersioningE(ctx, bucketName)
	require.NoError(ctx.T, err)
	return result
}

// EnableVersioningE enables versioning on a bucket
func EnableVersioningE(ctx TestContext, bucketName string) (*BucketVersioningResult, error) {
	client := createS3Client(ctx)

	input := &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	}

	_, err := client.PutBucketVersioning(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &BucketVersioningResult{
		BucketName: bucketName,
		Status:     types.BucketVersioningStatusEnabled,
	}, nil
}

// GetBucketVersioning gets the versioning status of a bucket
func GetBucketVersioning(ctx TestContext, bucketName string) *BucketVersioningResult {
	result, err := GetBucketVersioningE(ctx, bucketName)
	require.NoError(ctx.T, err)
	return result
}

// GetBucketVersioningE gets the versioning status of a bucket
func GetBucketVersioningE(ctx TestContext, bucketName string) (*BucketVersioningResult, error) {
	client := createS3Client(ctx)

	input := &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	}

	output, err := client.GetBucketVersioning(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &BucketVersioningResult{
		BucketName: bucketName,
		Status:     output.Status,
	}, nil
}

// PutBucketCORS sets CORS configuration for a bucket
func PutBucketCORS(ctx TestContext, config CORSConfig) {
	err := PutBucketCORSE(ctx, config)
	require.NoError(ctx.T, err)
}

// PutBucketCORSE sets CORS configuration for a bucket
func PutBucketCORSE(ctx TestContext, config CORSConfig) error {
	client := createS3Client(ctx)

	var corsRules []types.CORSRule
	for _, rule := range config.CORSRules {
		corsRule := types.CORSRule{
			AllowedMethods: rule.AllowedMethods,
			AllowedOrigins: rule.AllowedOrigins,
		}

		if len(rule.AllowedHeaders) > 0 {
			corsRule.AllowedHeaders = rule.AllowedHeaders
		}

		if len(rule.ExposeHeaders) > 0 {
			corsRule.ExposeHeaders = rule.ExposeHeaders
		}

		if rule.MaxAgeSeconds > 0 {
			corsRule.MaxAgeSeconds = aws.Int32(rule.MaxAgeSeconds)
		}

		corsRules = append(corsRules, corsRule)
	}

	input := &s3.PutBucketCorsInput{
		Bucket: aws.String(config.BucketName),
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: corsRules,
		},
	}

	_, err := client.PutBucketCors(context.TODO(), input)
	return err
}

// PutBucketTagging sets tags for a bucket
func PutBucketTagging(ctx TestContext, bucketName string, tags map[string]string) {
	err := PutBucketTaggingE(ctx, bucketName, tags)
	require.NoError(ctx.T, err)
}

// PutBucketTaggingE sets tags for a bucket
func PutBucketTaggingE(ctx TestContext, bucketName string, tags map[string]string) error {
	client := createS3Client(ctx)

	var tagSet []types.Tag
	for key, value := range tags {
		tagSet = append(tagSet, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	input := &s3.PutBucketTaggingInput{
		Bucket: aws.String(bucketName),
		Tagging: &types.Tagging{
			TagSet: tagSet,
		},
	}

	_, err := client.PutBucketTagging(context.TODO(), input)
	return err
}

// PutObjectTagging sets tags for an object
func PutObjectTagging(ctx TestContext, bucketName, key string, tags map[string]string) {
	err := PutObjectTaggingE(ctx, bucketName, key, tags)
	require.NoError(ctx.T, err)
}

// PutObjectTaggingE sets tags for an object
func PutObjectTaggingE(ctx TestContext, bucketName, key string, tags map[string]string) error {
	client := createS3Client(ctx)

	var tagSet []types.Tag
	for tagKey, tagValue := range tags {
		tagSet = append(tagSet, types.Tag{
			Key:   aws.String(tagKey),
			Value: aws.String(tagValue),
		})
	}

	input := &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Tagging: &types.Tagging{
			TagSet: tagSet,
		},
	}

	_, err := client.PutObjectTagging(context.TODO(), input)
	return err
}

// CreatePresignedURL creates a presigned URL for an object
func CreatePresignedURL(ctx TestContext, bucketName, key string, duration time.Duration) *PresignedURLResult {
	result, err := CreatePresignedURLE(ctx, bucketName, key, duration)
	require.NoError(ctx.T, err)
	return result
}

// CreatePresignedURLE creates a presigned URL for an object
func CreatePresignedURLE(ctx TestContext, bucketName, key string, duration time.Duration) (*PresignedURLResult, error) {
	client := createS3Client(ctx)

	presigner := s3.NewPresignClient(client)

	request, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return nil, err
	}

	return &PresignedURLResult{
		URL:       request.URL,
		ExpiresAt: time.Now().Add(duration),
		Method:    request.Method,
		Duration:  duration,
	}, nil
}

// CreatePresignedUploadURL creates a presigned URL for uploading an object
func CreatePresignedUploadURL(ctx TestContext, bucketName, key string, duration time.Duration) *PresignedURLResult {
	result, err := CreatePresignedUploadURLE(ctx, bucketName, key, duration)
	require.NoError(ctx.T, err)
	return result
}

// CreatePresignedUploadURLE creates a presigned URL for uploading an object
func CreatePresignedUploadURLE(ctx TestContext, bucketName, key string, duration time.Duration) (*PresignedURLResult, error) {
	client := createS3Client(ctx)

	presigner := s3.NewPresignClient(client)

	request, err := presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return nil, err
	}

	return &PresignedURLResult{
		URL:       request.URL,
		ExpiresAt: time.Now().Add(duration),
		Method:    request.Method,
		Duration:  duration,
	}, nil
}

// UploadLargeObject uploads a large object using multipart upload
func UploadLargeObject(ctx TestContext, bucketName, key string, body io.Reader) *ObjectResult {
	result, err := UploadLargeObjectE(ctx, bucketName, key, body)
	require.NoError(ctx.T, err)
	return result
}

// UploadLargeObjectE uploads a large object using multipart upload
func UploadLargeObjectE(ctx TestContext, bucketName, key string, body io.Reader) (*ObjectResult, error) {
	client := createS3Client(ctx)

	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = DefaultPartSize
		u.Concurrency = 5
	})

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   body,
	}

	result, err := uploader.Upload(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	return &ObjectResult{
		BucketName: bucketName,
		Key:        key,
		ETag:       aws.ToString(result.ETag),
		VersionId:  aws.ToString(result.VersionID),
	}, nil
}

// WaitForBucketExists waits for a bucket to exist
func WaitForBucketExists(ctx TestContext, bucketName string, timeout time.Duration) {
	err := WaitForBucketExistsE(ctx, bucketName, timeout)
	require.NoError(ctx.T, err)
}

// WaitForBucketExistsE waits for a bucket to exist
func WaitForBucketExistsE(ctx TestContext, bucketName string, timeout time.Duration) error {
	startTime := time.Now()

	for time.Since(startTime) < timeout {
		exists, err := BucketExistsE(ctx, bucketName)
		if err != nil {
			return err
		}

		if exists {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for bucket to exist: %s", bucketName)
}

// WaitForObjectExists waits for an object to exist
func WaitForObjectExists(ctx TestContext, bucketName, key string, timeout time.Duration) {
	err := WaitForObjectExistsE(ctx, bucketName, key, timeout)
	require.NoError(ctx.T, err)
}

// WaitForObjectExistsE waits for an object to exist
func WaitForObjectExistsE(ctx TestContext, bucketName, key string, timeout time.Duration) error {
	startTime := time.Now()

	for time.Since(startTime) < timeout {
		exists, err := ObjectExistsE(ctx, bucketName, key)
		if err != nil {
			return err
		}

		if exists {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for object to exist: %s/%s", bucketName, key)
}

// EmptyBucket deletes all objects in a bucket
func EmptyBucket(ctx TestContext, bucketName string) {
	err := EmptyBucketE(ctx, bucketName)
	require.NoError(ctx.T, err)
}

// EmptyBucketE deletes all objects in a bucket
func EmptyBucketE(ctx TestContext, bucketName string) error {
	objects, err := ListObjectsE(ctx, bucketName)
	if err != nil {
		return err
	}

	for _, object := range objects {
		err := DeleteObjectE(ctx, bucketName, object.Key)
		if err != nil {
			return err
		}
	}

	return nil
}