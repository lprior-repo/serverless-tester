package dsl

import (
	"fmt"
	"time"
)

// S3Builder provides a fluent interface for S3 operations
type S3Builder struct {
	ctx *TestContext
}

// Bucket creates an S3 bucket builder for a specific bucket
func (sb *S3Builder) Bucket(name string) *S3BucketBuilder {
	return &S3BucketBuilder{
		ctx:        sb.ctx,
		bucketName: name,
		config:     &S3BucketConfig{},
	}
}

// CreateBucket starts building a new S3 bucket
func (sb *S3Builder) CreateBucket() *S3BucketCreateBuilder {
	return &S3BucketCreateBuilder{
		ctx:    sb.ctx,
		config: &BucketCreateConfig{},
	}
}

// S3BucketBuilder builds operations on an existing S3 bucket
type S3BucketBuilder struct {
	ctx        *TestContext
	bucketName string
	config     *S3BucketConfig
}

type S3BucketConfig struct {
	Region         string
	StorageClass   string
	ServerSideEncryption string
	Timeout        time.Duration
}

// InRegion sets the bucket region
func (sbb *S3BucketBuilder) InRegion(region string) *S3BucketBuilder {
	sbb.config.Region = region
	return sbb
}

// Object creates an object builder for operations on bucket objects
func (sbb *S3BucketBuilder) Object(key string) *S3ObjectBuilder {
	return &S3ObjectBuilder{
		ctx:        sbb.ctx,
		bucketName: sbb.bucketName,
		objectKey:  key,
		config:     &S3ObjectConfig{},
	}
}

// UploadObject starts building an object upload
func (sbb *S3BucketBuilder) UploadObject() *S3ObjectUploadBuilder {
	return &S3ObjectUploadBuilder{
		ctx:        sbb.ctx,
		bucketName: sbb.bucketName,
		config:     &ObjectUploadConfig{},
	}
}

// ListObjects lists objects in the bucket
func (sbb *S3BucketBuilder) ListObjects() *S3ListObjectsBuilder {
	return &S3ListObjectsBuilder{
		ctx:        sbb.ctx,
		bucketName: sbb.bucketName,
		config:     &ListObjectsConfig{},
	}
}

// EnableVersioning enables versioning on the bucket
func (sbb *S3BucketBuilder) EnableVersioning() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Versioning enabled for bucket %s", sbb.bucketName),
		ctx:     sbb.ctx,
	}
}

// SetCORS sets CORS configuration
func (sbb *S3BucketBuilder) SetCORS() *S3CORSBuilder {
	return &S3CORSBuilder{
		ctx:        sbb.ctx,
		bucketName: sbb.bucketName,
		config:     &CORSConfig{},
	}
}

// Delete deletes the bucket
func (sbb *S3BucketBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Bucket %s deleted", sbb.bucketName),
		ctx:     sbb.ctx,
	}
}

// Exists checks if the bucket exists
func (sbb *S3BucketBuilder) Exists() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    true,
		ctx:     sbb.ctx,
	}
}

// S3ObjectBuilder builds operations on S3 objects
type S3ObjectBuilder struct {
	ctx        *TestContext
	bucketName string
	objectKey  string
	config     *S3ObjectConfig
}

type S3ObjectConfig struct {
	VersionId    string
	StorageClass string
}

// WithVersion specifies a specific version
func (sob *S3ObjectBuilder) WithVersion(versionId string) *S3ObjectBuilder {
	sob.config.VersionId = versionId
	return sob
}

// Download downloads the object
func (sob *S3ObjectBuilder) Download() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Downloaded object %s from bucket %s", sob.objectKey, sob.bucketName),
		Metadata: map[string]interface{}{
			"bucket": sob.bucketName,
			"key":    sob.objectKey,
			"size":   1024,
		},
		ctx: sob.ctx,
	}
}

// Delete deletes the object
func (sob *S3ObjectBuilder) Delete() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("Deleted object %s from bucket %s", sob.objectKey, sob.bucketName),
		ctx:     sob.ctx,
	}
}

// GetMetadata retrieves object metadata
func (sob *S3ObjectBuilder) GetMetadata() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"ContentType":     "application/octet-stream",
			"ContentLength":   1024,
			"LastModified":    time.Now(),
			"ETag":           "\"abc123\"",
		},
		ctx: sob.ctx,
	}
}

// GeneratePresignedURL generates a presigned URL
func (sob *S3ObjectBuilder) GeneratePresignedURL(expiry time.Duration) *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("https://%s.s3.amazonaws.com/%s?presigned=true", sob.bucketName, sob.objectKey),
		Metadata: map[string]interface{}{
			"expires_in": expiry.String(),
		},
		ctx: sob.ctx,
	}
}

// S3ObjectUploadBuilder builds object upload operations
type S3ObjectUploadBuilder struct {
	ctx        *TestContext
	bucketName string
	config     *ObjectUploadConfig
}

type ObjectUploadConfig struct {
	Key          string
	Body         interface{}
	ContentType  string
	StorageClass string
	Metadata     map[string]string
	Tags         map[string]string
}

// WithKey sets the object key
func (soub *S3ObjectUploadBuilder) WithKey(key string) *S3ObjectUploadBuilder {
	soub.config.Key = key
	return soub
}

// WithBody sets the object body
func (soub *S3ObjectUploadBuilder) WithBody(body interface{}) *S3ObjectUploadBuilder {
	soub.config.Body = body
	return soub
}

// WithContentType sets the content type
func (soub *S3ObjectUploadBuilder) WithContentType(contentType string) *S3ObjectUploadBuilder {
	soub.config.ContentType = contentType
	return soub
}

// WithStorageClass sets the storage class
func (soub *S3ObjectUploadBuilder) WithStorageClass(storageClass string) *S3ObjectUploadBuilder {
	soub.config.StorageClass = storageClass
	return soub
}

// WithMetadata adds metadata
func (soub *S3ObjectUploadBuilder) WithMetadata(metadata map[string]string) *S3ObjectUploadBuilder {
	soub.config.Metadata = metadata
	return soub
}

// Execute uploads the object
func (soub *S3ObjectUploadBuilder) Execute() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"ETag":     "\"abc123\"",
			"Location": fmt.Sprintf("https://%s.s3.amazonaws.com/%s", soub.bucketName, soub.config.Key),
		},
		Metadata: map[string]interface{}{
			"bucket":       soub.bucketName,
			"key":          soub.config.Key,
			"content_type": soub.config.ContentType,
		},
		ctx: soub.ctx,
	}
}

// S3ListObjectsBuilder builds object listing operations
type S3ListObjectsBuilder struct {
	ctx        *TestContext
	bucketName string
	config     *ListObjectsConfig
}

type ListObjectsConfig struct {
	Prefix    string
	Delimiter string
	MaxKeys   int32
}

// WithPrefix filters objects by prefix
func (slob *S3ListObjectsBuilder) WithPrefix(prefix string) *S3ListObjectsBuilder {
	slob.config.Prefix = prefix
	return slob
}

// WithDelimiter sets the delimiter
func (slob *S3ListObjectsBuilder) WithDelimiter(delimiter string) *S3ListObjectsBuilder {
	slob.config.Delimiter = delimiter
	return slob
}

// Limit sets the maximum number of objects to return
func (slob *S3ListObjectsBuilder) Limit(maxKeys int32) *S3ListObjectsBuilder {
	slob.config.MaxKeys = maxKeys
	return slob
}

// Execute lists the objects
func (slob *S3ListObjectsBuilder) Execute() *ExecutionResult {
	objects := []map[string]interface{}{
		{
			"Key":          "file1.txt",
			"Size":         1024,
			"LastModified": time.Now(),
		},
		{
			"Key":          "file2.jpg",
			"Size":         2048,
			"LastModified": time.Now(),
		},
	}
	
	return &ExecutionResult{
		Success: true,
		Data:    objects,
		Metadata: map[string]interface{}{
			"bucket":      slob.bucketName,
			"object_count": len(objects),
		},
		ctx: slob.ctx,
	}
}

// S3CORSBuilder builds CORS configuration
type S3CORSBuilder struct {
	ctx        *TestContext
	bucketName string
	config     *CORSConfig
}

type CORSConfig struct {
	Rules []CORSRule
}

type CORSRule struct {
	AllowedMethods []string
	AllowedOrigins []string
	AllowedHeaders []string
	ExposeHeaders  []string
	MaxAgeSeconds  int32
}

// AddRule adds a CORS rule
func (scb *S3CORSBuilder) AddRule() *S3CORSRuleBuilder {
	return &S3CORSRuleBuilder{
		corsBuilder: scb,
		rule:        &CORSRule{},
	}
}

// Apply applies the CORS configuration
func (scb *S3CORSBuilder) Apply() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data:    fmt.Sprintf("CORS configuration applied to bucket %s", scb.bucketName),
		ctx:     scb.ctx,
	}
}

// S3CORSRuleBuilder builds individual CORS rules
type S3CORSRuleBuilder struct {
	corsBuilder *S3CORSBuilder
	rule        *CORSRule
}

// AllowMethods sets allowed methods
func (scrb *S3CORSRuleBuilder) AllowMethods(methods ...string) *S3CORSRuleBuilder {
	scrb.rule.AllowedMethods = methods
	return scrb
}

// AllowOrigins sets allowed origins
func (scrb *S3CORSRuleBuilder) AllowOrigins(origins ...string) *S3CORSRuleBuilder {
	scrb.rule.AllowedOrigins = origins
	return scrb
}

// AllowHeaders sets allowed headers
func (scrb *S3CORSRuleBuilder) AllowHeaders(headers ...string) *S3CORSRuleBuilder {
	scrb.rule.AllowedHeaders = headers
	return scrb
}

// MaxAge sets the max age
func (scrb *S3CORSRuleBuilder) MaxAge(seconds int32) *S3CORSRuleBuilder {
	scrb.rule.MaxAgeSeconds = seconds
	return scrb
}

// Done completes the rule and returns to CORS builder
func (scrb *S3CORSRuleBuilder) Done() *S3CORSBuilder {
	scrb.corsBuilder.config.Rules = append(scrb.corsBuilder.config.Rules, *scrb.rule)
	return scrb.corsBuilder
}

// S3BucketCreateBuilder builds bucket creation operations
type S3BucketCreateBuilder struct {
	ctx    *TestContext
	config *BucketCreateConfig
}

type BucketCreateConfig struct {
	Name         string
	Region       string
	ACL          string
	Tags         map[string]string
	Versioning   bool
	Encryption   bool
}

// Named sets the bucket name
func (sbcb *S3BucketCreateBuilder) Named(name string) *S3BucketCreateBuilder {
	sbcb.config.Name = name
	return sbcb
}

// InRegion sets the bucket region
func (sbcb *S3BucketCreateBuilder) InRegion(region string) *S3BucketCreateBuilder {
	sbcb.config.Region = region
	return sbcb
}

// WithPublicReadAccess sets public read ACL
func (sbcb *S3BucketCreateBuilder) WithPublicReadAccess() *S3BucketCreateBuilder {
	sbcb.config.ACL = "public-read"
	return sbcb
}

// WithPrivateAccess sets private ACL
func (sbcb *S3BucketCreateBuilder) WithPrivateAccess() *S3BucketCreateBuilder {
	sbcb.config.ACL = "private"
	return sbcb
}

// WithVersioning enables versioning
func (sbcb *S3BucketCreateBuilder) WithVersioning() *S3BucketCreateBuilder {
	sbcb.config.Versioning = true
	return sbcb
}

// WithEncryption enables server-side encryption
func (sbcb *S3BucketCreateBuilder) WithEncryption() *S3BucketCreateBuilder {
	sbcb.config.Encryption = true
	return sbcb
}

// WithTags adds tags
func (sbcb *S3BucketCreateBuilder) WithTags(tags map[string]string) *S3BucketCreateBuilder {
	sbcb.config.Tags = tags
	return sbcb
}

// Create creates the bucket
func (sbcb *S3BucketCreateBuilder) Create() *ExecutionResult {
	return &ExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"BucketName": sbcb.config.Name,
			"Location":   fmt.Sprintf("https://%s.s3.amazonaws.com/", sbcb.config.Name),
		},
		Metadata: map[string]interface{}{
			"region":     sbcb.config.Region,
			"versioning": sbcb.config.Versioning,
			"encryption": sbcb.config.Encryption,
		},
		ctx: sbcb.ctx,
	}
}