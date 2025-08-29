package s3

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// FunctionalS3Config represents immutable S3 configuration using functional options
type FunctionalS3Config struct {
	bucketName        string
	region            string
	timeout           time.Duration
	maxRetries        int
	retryDelay        time.Duration
	storageClass      types.StorageClass
	serverSideEncryption mo.Option[types.ServerSideEncryption]
	kmsKeyId          mo.Option[string]
	acl               mo.Option[types.BucketCannedACL]
	versioningEnabled bool
	corsEnabled       bool
	metadata          map[string]interface{}
}

// S3ConfigOption represents a functional option for FunctionalS3Config
type S3ConfigOption func(FunctionalS3Config) FunctionalS3Config

// Constants for functional S3 operations
const (
	FunctionalS3DefaultTimeout    = 30 * time.Second
	FunctionalS3DefaultRetries    = 3
	FunctionalS3DefaultRetryDelay = 1 * time.Second
)

// NewFunctionalS3Config creates a new S3 config with functional options
func NewFunctionalS3Config(bucketName string, opts ...S3ConfigOption) FunctionalS3Config {
	base := FunctionalS3Config{
		bucketName:           bucketName,
		region:               "us-east-1", // Default region
		timeout:              FunctionalS3DefaultTimeout,
		maxRetries:           FunctionalS3DefaultRetries,
		retryDelay:           FunctionalS3DefaultRetryDelay,
		storageClass:         types.StorageClassStandard,
		serverSideEncryption: mo.None[types.ServerSideEncryption](),
		kmsKeyId:            mo.None[string](),
		acl:                 mo.None[types.BucketCannedACL](),
		versioningEnabled:   false,
		corsEnabled:         false,
		metadata:            make(map[string]interface{}),
	}

	return lo.Reduce(opts, func(config FunctionalS3Config, opt S3ConfigOption, _ int) FunctionalS3Config {
		return opt(config)
	}, base)
}

// WithS3Region sets the AWS region functionally
func WithS3Region(region string) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               region,
			timeout:              config.timeout,
			maxRetries:           config.maxRetries,
			retryDelay:           config.retryDelay,
			storageClass:         config.storageClass,
			serverSideEncryption: config.serverSideEncryption,
			kmsKeyId:            config.kmsKeyId,
			acl:                 config.acl,
			versioningEnabled:   config.versioningEnabled,
			corsEnabled:         config.corsEnabled,
			metadata:            config.metadata,
		}
	}
}

// WithS3Timeout sets the timeout functionally
func WithS3Timeout(timeout time.Duration) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               config.region,
			timeout:              timeout,
			maxRetries:           config.maxRetries,
			retryDelay:           config.retryDelay,
			storageClass:         config.storageClass,
			serverSideEncryption: config.serverSideEncryption,
			kmsKeyId:            config.kmsKeyId,
			acl:                 config.acl,
			versioningEnabled:   config.versioningEnabled,
			corsEnabled:         config.corsEnabled,
			metadata:            config.metadata,
		}
	}
}

// WithS3RetryConfig sets retry configuration functionally
func WithS3RetryConfig(maxRetries int, retryDelay time.Duration) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               config.region,
			timeout:              config.timeout,
			maxRetries:           maxRetries,
			retryDelay:           retryDelay,
			storageClass:         config.storageClass,
			serverSideEncryption: config.serverSideEncryption,
			kmsKeyId:            config.kmsKeyId,
			acl:                 config.acl,
			versioningEnabled:   config.versioningEnabled,
			corsEnabled:         config.corsEnabled,
			metadata:            config.metadata,
		}
	}
}

// WithStorageClass sets the storage class functionally
func WithStorageClass(storageClass types.StorageClass) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               config.region,
			timeout:              config.timeout,
			maxRetries:           config.maxRetries,
			retryDelay:           config.retryDelay,
			storageClass:         storageClass,
			serverSideEncryption: config.serverSideEncryption,
			kmsKeyId:            config.kmsKeyId,
			acl:                 config.acl,
			versioningEnabled:   config.versioningEnabled,
			corsEnabled:         config.corsEnabled,
			metadata:            config.metadata,
		}
	}
}

// WithServerSideEncryption sets encryption functionally
func WithServerSideEncryption(encryption types.ServerSideEncryption) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               config.region,
			timeout:              config.timeout,
			maxRetries:           config.maxRetries,
			retryDelay:           config.retryDelay,
			storageClass:         config.storageClass,
			serverSideEncryption: mo.Some(encryption),
			kmsKeyId:            config.kmsKeyId,
			acl:                 config.acl,
			versioningEnabled:   config.versioningEnabled,
			corsEnabled:         config.corsEnabled,
			metadata:            config.metadata,
		}
	}
}

// WithKMSKeyId sets KMS key ID functionally
func WithKMSKeyId(kmsKeyId string) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               config.region,
			timeout:              config.timeout,
			maxRetries:           config.maxRetries,
			retryDelay:           config.retryDelay,
			storageClass:         config.storageClass,
			serverSideEncryption: config.serverSideEncryption,
			kmsKeyId:            mo.Some(kmsKeyId),
			acl:                 config.acl,
			versioningEnabled:   config.versioningEnabled,
			corsEnabled:         config.corsEnabled,
			metadata:            config.metadata,
		}
	}
}

// WithACL sets the ACL functionally
func WithACL(acl types.BucketCannedACL) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               config.region,
			timeout:              config.timeout,
			maxRetries:           config.maxRetries,
			retryDelay:           config.retryDelay,
			storageClass:         config.storageClass,
			serverSideEncryption: config.serverSideEncryption,
			kmsKeyId:            config.kmsKeyId,
			acl:                 mo.Some(acl),
			versioningEnabled:   config.versioningEnabled,
			corsEnabled:         config.corsEnabled,
			metadata:            config.metadata,
		}
	}
}

// WithVersioning enables versioning functionally
func WithVersioning(enabled bool) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               config.region,
			timeout:              config.timeout,
			maxRetries:           config.maxRetries,
			retryDelay:           config.retryDelay,
			storageClass:         config.storageClass,
			serverSideEncryption: config.serverSideEncryption,
			kmsKeyId:            config.kmsKeyId,
			acl:                 config.acl,
			versioningEnabled:   enabled,
			corsEnabled:         config.corsEnabled,
			metadata:            config.metadata,
		}
	}
}

// WithCORS enables CORS functionally
func WithCORS(enabled bool) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               config.region,
			timeout:              config.timeout,
			maxRetries:           config.maxRetries,
			retryDelay:           config.retryDelay,
			storageClass:         config.storageClass,
			serverSideEncryption: config.serverSideEncryption,
			kmsKeyId:            config.kmsKeyId,
			acl:                 config.acl,
			versioningEnabled:   config.versioningEnabled,
			corsEnabled:         enabled,
			metadata:            config.metadata,
		}
	}
}

// WithS3Metadata adds metadata functionally
func WithS3Metadata(key string, value interface{}) S3ConfigOption {
	return func(config FunctionalS3Config) FunctionalS3Config {
		newMetadata := lo.Assign(config.metadata, map[string]interface{}{key: value})
		return FunctionalS3Config{
			bucketName:           config.bucketName,
			region:               config.region,
			timeout:              config.timeout,
			maxRetries:           config.maxRetries,
			retryDelay:           config.retryDelay,
			storageClass:         config.storageClass,
			serverSideEncryption: config.serverSideEncryption,
			kmsKeyId:            config.kmsKeyId,
			acl:                 config.acl,
			versioningEnabled:   config.versioningEnabled,
			corsEnabled:         config.corsEnabled,
			metadata:            newMetadata,
		}
	}
}

// FunctionalS3Object represents an immutable S3 object
type FunctionalS3Object struct {
	key         string
	content     mo.Option[string]
	contentType string
	size        int64
	etag        mo.Option[string]
	versionId   mo.Option[string]
	lastModified time.Time
	metadata    map[string]string
}

// NewFunctionalS3Object creates a new S3 object
func NewFunctionalS3Object(key string) FunctionalS3Object {
	return FunctionalS3Object{
		key:         key,
		content:     mo.None[string](),
		contentType: "binary/octet-stream",
		size:        0,
		etag:        mo.None[string](),
		versionId:   mo.None[string](),
		lastModified: time.Now(),
		metadata:    make(map[string]string),
	}
}

// WithContent sets the object content
func (fso FunctionalS3Object) WithContent(content string) FunctionalS3Object {
	return FunctionalS3Object{
		key:         fso.key,
		content:     mo.Some(content),
		contentType: fso.contentType,
		size:        int64(len(content)),
		etag:        fso.etag,
		versionId:   fso.versionId,
		lastModified: fso.lastModified,
		metadata:    fso.metadata,
	}
}

// WithContentType sets the content type
func (fso FunctionalS3Object) WithContentType(contentType string) FunctionalS3Object {
	return FunctionalS3Object{
		key:         fso.key,
		content:     fso.content,
		contentType: contentType,
		size:        fso.size,
		etag:        fso.etag,
		versionId:   fso.versionId,
		lastModified: fso.lastModified,
		metadata:    fso.metadata,
	}
}

// WithETag sets the ETag
func (fso FunctionalS3Object) WithETag(etag string) FunctionalS3Object {
	return FunctionalS3Object{
		key:         fso.key,
		content:     fso.content,
		contentType: fso.contentType,
		size:        fso.size,
		etag:        mo.Some(etag),
		versionId:   fso.versionId,
		lastModified: fso.lastModified,
		metadata:    fso.metadata,
	}
}

// WithVersionId sets the version ID
func (fso FunctionalS3Object) WithVersionId(versionId string) FunctionalS3Object {
	return FunctionalS3Object{
		key:         fso.key,
		content:     fso.content,
		contentType: fso.contentType,
		size:        fso.size,
		etag:        fso.etag,
		versionId:   mo.Some(versionId),
		lastModified: fso.lastModified,
		metadata:    fso.metadata,
	}
}

// WithObjectMetadata adds metadata to the object
func (fso FunctionalS3Object) WithObjectMetadata(key, value string) FunctionalS3Object {
	newMetadata := lo.Assign(fso.metadata, map[string]string{key: value})
	return FunctionalS3Object{
		key:         fso.key,
		content:     fso.content,
		contentType: fso.contentType,
		size:        fso.size,
		etag:        fso.etag,
		versionId:   fso.versionId,
		lastModified: fso.lastModified,
		metadata:    newMetadata,
	}
}

// GetKey returns the object key
func (fso FunctionalS3Object) GetKey() string {
	return fso.key
}

// GetContent returns the object content if present
func (fso FunctionalS3Object) GetContent() mo.Option[string] {
	return fso.content
}

// GetContentType returns the content type
func (fso FunctionalS3Object) GetContentType() string {
	return fso.contentType
}

// GetSize returns the object size
func (fso FunctionalS3Object) GetSize() int64 {
	return fso.size
}

// GetETag returns the ETag if present
func (fso FunctionalS3Object) GetETag() mo.Option[string] {
	return fso.etag
}

// GetVersionId returns the version ID if present
func (fso FunctionalS3Object) GetVersionId() mo.Option[string] {
	return fso.versionId
}

// GetLastModified returns the last modified time
func (fso FunctionalS3Object) GetLastModified() time.Time {
	return fso.lastModified
}

// GetObjectMetadata returns the object metadata
func (fso FunctionalS3Object) GetObjectMetadata() map[string]string {
	return fso.metadata
}

// FunctionalS3OperationResult represents the result of an S3 operation
type FunctionalS3OperationResult struct {
	result   mo.Option[interface{}]
	error    mo.Option[error]
	duration time.Duration
	metadata map[string]interface{}
}

// NewSuccessS3Result creates a successful operation result
func NewSuccessS3Result(result interface{}, duration time.Duration, metadata map[string]interface{}) FunctionalS3OperationResult {
	return FunctionalS3OperationResult{
		result:   mo.Some(result),
		error:    mo.None[error](),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
	}
}

// NewErrorS3Result creates an error operation result
func NewErrorS3Result(err error, duration time.Duration, metadata map[string]interface{}) FunctionalS3OperationResult {
	return FunctionalS3OperationResult{
		result:   mo.None[interface{}](),
		error:    mo.Some(err),
		duration: duration,
		metadata: lo.Assign(map[string]interface{}{}, metadata),
	}
}

// IsSuccess returns true if the operation was successful
func (fsor FunctionalS3OperationResult) IsSuccess() bool {
	return fsor.result.IsPresent() && fsor.error.IsAbsent()
}

// IsError returns true if the operation has an error
func (fsor FunctionalS3OperationResult) IsError() bool {
	return fsor.error.IsPresent()
}

// GetResult returns the operation result if present
func (fsor FunctionalS3OperationResult) GetResult() mo.Option[interface{}] {
	return fsor.result
}

// GetError returns the error if present
func (fsor FunctionalS3OperationResult) GetError() mo.Option[error] {
	return fsor.error
}

// GetDuration returns the execution duration
func (fsor FunctionalS3OperationResult) GetDuration() time.Duration {
	return fsor.duration
}

// GetMetadata returns the metadata
func (fsor FunctionalS3OperationResult) GetMetadata() map[string]interface{} {
	return fsor.metadata
}

// MapResult applies a function to the result if present
func (fsor FunctionalS3OperationResult) MapResult(f func(interface{}) interface{}) FunctionalS3OperationResult {
	if fsor.IsSuccess() {
		newResult := fsor.result.Map(func(result interface{}) (interface{}, bool) {
			return f(result), true
		})
		return FunctionalS3OperationResult{
			result:   newResult,
			error:    fsor.error,
			duration: fsor.duration,
			metadata: fsor.metadata,
		}
	}
	return fsor
}

// FunctionalS3Validation provides functional validation utilities
func FunctionalS3Validation() struct {
	ValidateBucketName func(string) error
	ValidateObjectKey  func(string) error
	ValidateRegion     func(string) error
} {
	return struct {
		ValidateBucketName func(string) error
		ValidateObjectKey  func(string) error
		ValidateRegion     func(string) error
	}{
		ValidateBucketName: func(bucketName string) error {
			if len(bucketName) == 0 {
				return fmt.Errorf("bucket name cannot be empty")
			}
			if len(bucketName) < 3 {
				return fmt.Errorf("bucket name must be at least 3 characters long")
			}
			if len(bucketName) > 63 {
				return fmt.Errorf("bucket name cannot be longer than 63 characters")
			}
			if strings.Contains(bucketName, "_") {
				return fmt.Errorf("bucket name cannot contain underscores")
			}
			if strings.Contains(bucketName, " ") {
				return fmt.Errorf("bucket name cannot contain spaces")
			}
			return nil
		},
		ValidateObjectKey: func(objectKey string) error {
			if len(objectKey) == 0 {
				return fmt.Errorf("object key cannot be empty")
			}
			if len(objectKey) > 1024 {
				return fmt.Errorf("object key cannot be longer than 1024 characters")
			}
			return nil
		},
		ValidateRegion: func(region string) error {
			if len(region) == 0 {
				return fmt.Errorf("region cannot be empty")
			}
			validRegions := []string{
				"us-east-1", "us-east-2", "us-west-1", "us-west-2",
				"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
				"ap-northeast-1", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2",
				"ap-south-1", "sa-east-1", "ca-central-1",
			}
			if !lo.Contains(validRegions, region) {
				return fmt.Errorf("invalid region: %s", region)
			}
			return nil
		},
	}
}

// FunctionalRetryResult represents the result of a retry operation for S3
type FunctionalS3RetryResult[T any] struct {
	result mo.Option[T]
	error  mo.Option[error]
}

// IsError returns true if the retry failed
func (fsrr FunctionalS3RetryResult[T]) IsError() bool {
	return fsrr.error.IsPresent()
}

// GetResult returns the result if present
func (fsrr FunctionalS3RetryResult[T]) GetResult() mo.Option[T] {
	return fsrr.result
}

// GetError returns the error if present
func (fsrr FunctionalS3RetryResult[T]) GetError() mo.Option[error] {
	return fsrr.error
}

// FunctionalS3RetryWithBackoff retries an S3 operation with exponential backoff
func FunctionalS3RetryWithBackoff[T any](operation func() (T, error), maxRetries int, delay time.Duration) FunctionalS3RetryResult[T] {
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := operation()
		if err == nil {
			return FunctionalS3RetryResult[T]{
				result: mo.Some(result),
				error:  mo.None[error](),
			}
		}
		
		// Wait before retry (except on last attempt)
		if attempt < maxRetries-1 {
			backoffDelay := time.Duration(attempt+1) * delay
			time.Sleep(backoffDelay)
		}
	}
	
	// Final attempt to capture the last error
	_, err := operation()
	return FunctionalS3RetryResult[T]{
		result: mo.None[T](),
		error:  mo.Some(err),
	}
}

// FunctionalValidationPipeline represents a validation pipeline for S3
type FunctionalS3ValidationPipeline[T any] struct {
	value T
	error mo.Option[error]
}

// ExecuteFunctionalS3ValidationPipeline executes a series of validation functions
func ExecuteFunctionalS3ValidationPipeline[T any](input T, operations ...func(T) (T, error)) FunctionalS3ValidationPipeline[T] {
	return lo.Reduce(operations, func(acc FunctionalS3ValidationPipeline[T], op func(T) (T, error), _ int) FunctionalS3ValidationPipeline[T] {
		if acc.error.IsPresent() {
			return acc
		}
		
		result, err := op(acc.value)
		if err != nil {
			return FunctionalS3ValidationPipeline[T]{
				value: acc.value,
				error: mo.Some(err),
			}
		}
		
		return FunctionalS3ValidationPipeline[T]{
			value: result,
			error: mo.None[error](),
		}
	}, FunctionalS3ValidationPipeline[T]{
		value: input,
		error: mo.None[error](),
	})
}

// IsError returns true if the pipeline has an error
func (fsvp FunctionalS3ValidationPipeline[T]) IsError() bool {
	return fsvp.error.IsPresent()
}

// GetError returns the error if present
func (fsvp FunctionalS3ValidationPipeline[T]) GetError() mo.Option[error] {
	return fsvp.error
}

// GetValue returns the value
func (fsvp FunctionalS3ValidationPipeline[T]) GetValue() T {
	return fsvp.value
}

// SimulateFunctionalS3CreateBucket simulates creating an S3 bucket
func SimulateFunctionalS3CreateBucket(config FunctionalS3Config) FunctionalS3OperationResult {
	startTime := time.Now()
	
	validator := FunctionalS3Validation()
	
	// Validate inputs using functional composition
	validationPipeline := ExecuteFunctionalS3ValidationPipeline[string](config.bucketName,
		func(bucketName string) (string, error) {
			if err := validator.ValidateBucketName(bucketName); err != nil {
				return bucketName, err
			}
			return bucketName, nil
		},
		func(bucketName string) (string, error) {
			if err := validator.ValidateRegion(config.region); err != nil {
				return bucketName, err
			}
			return bucketName, nil
		},
	)
	
	if validationPipeline.IsError() {
		return NewErrorS3Result(
			validationPipeline.GetError().MustGet(),
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"bucket_name": config.bucketName,
				"operation":   "create_bucket",
				"phase":       "validation",
			}),
		)
	}

	// Simulate create bucket operation with retry logic
	retryResult := FunctionalS3RetryWithBackoff(
		func() (map[string]interface{}, error) {
			// Simulate successful bucket creation
			result := map[string]interface{}{
				"bucket_name":     config.bucketName,
				"region":          config.region,
				"operation":       "create_bucket",
				"versioning":      config.versioningEnabled,
				"cors_enabled":    config.corsEnabled,
				"storage_class":   config.storageClass,
				"created_at":      time.Now().Format(time.RFC3339),
			}
			
			if config.acl.IsPresent() {
				result["acl"] = config.acl.MustGet()
			}
			
			if config.serverSideEncryption.IsPresent() {
				result["encryption"] = config.serverSideEncryption.MustGet()
				if config.kmsKeyId.IsPresent() {
					result["kms_key_id"] = config.kmsKeyId.MustGet()
				}
			}
			
			return result, nil
		},
		config.maxRetries,
		config.retryDelay,
	)
	
	totalDuration := time.Since(startTime)
	
	if retryResult.IsError() {
		return NewErrorS3Result(
			retryResult.GetError().MustGet(),
			totalDuration,
			lo.Assign(config.metadata, map[string]interface{}{
				"bucket_name": config.bucketName,
				"operation":   "create_bucket",
				"phase":       "execution",
				"retries":     config.maxRetries,
			}),
		)
	}
	
	return NewSuccessS3Result(
		retryResult.GetResult().MustGet(),
		totalDuration,
		lo.Assign(config.metadata, map[string]interface{}{
			"bucket_name": config.bucketName,
			"operation":   "create_bucket",
			"phase":       "success",
		}),
	)
}

// SimulateFunctionalS3PutObject simulates putting an object to S3
func SimulateFunctionalS3PutObject(object FunctionalS3Object, config FunctionalS3Config) FunctionalS3OperationResult {
	startTime := time.Now()
	
	validator := FunctionalS3Validation()
	
	// Validate inputs using functional composition
	validationPipeline := ExecuteFunctionalS3ValidationPipeline[FunctionalS3Object](object,
		func(obj FunctionalS3Object) (FunctionalS3Object, error) {
			if err := validator.ValidateBucketName(config.bucketName); err != nil {
				return obj, err
			}
			return obj, nil
		},
		func(obj FunctionalS3Object) (FunctionalS3Object, error) {
			if err := validator.ValidateObjectKey(obj.key); err != nil {
				return obj, err
			}
			return obj, nil
		},
		func(obj FunctionalS3Object) (FunctionalS3Object, error) {
			if obj.content.IsAbsent() {
				return obj, fmt.Errorf("object content cannot be empty")
			}
			return obj, nil
		},
	)
	
	if validationPipeline.IsError() {
		return NewErrorS3Result(
			validationPipeline.GetError().MustGet(),
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"bucket_name": config.bucketName,
				"object_key":  object.key,
				"operation":   "put_object",
				"phase":       "validation",
			}),
		)
	}

	// Simulate put object operation with retry logic
	retryResult := FunctionalS3RetryWithBackoff(
		func() (FunctionalS3Object, error) {
			// Simulate successful object creation with ETag and version ID
			resultObject := object.
				WithETag("\"d41d8cd98f00b204e9800998ecf8427e\"").
				WithVersionId("3/L4kqtJlcpXroDTDmJ+rmSpXd3dIbrHY+MTRCxf3vjVBH40Nr8X8gdRQBpUMLUo")
			
			return resultObject, nil
		},
		config.maxRetries,
		config.retryDelay,
	)
	
	totalDuration := time.Since(startTime)
	
	if retryResult.IsError() {
		return NewErrorS3Result(
			retryResult.GetError().MustGet(),
			totalDuration,
			lo.Assign(config.metadata, map[string]interface{}{
				"bucket_name": config.bucketName,
				"object_key":  object.key,
				"operation":   "put_object",
				"phase":       "execution",
				"retries":     config.maxRetries,
			}),
		)
	}
	
	return NewSuccessS3Result(
		retryResult.GetResult().MustGet(),
		totalDuration,
		lo.Assign(config.metadata, map[string]interface{}{
			"bucket_name": config.bucketName,
			"object_key":  object.key,
			"operation":   "put_object",
			"phase":       "success",
			"object_size": object.size,
		}),
	)
}

// SimulateFunctionalS3GetObject simulates getting an object from S3
func SimulateFunctionalS3GetObject(objectKey string, config FunctionalS3Config) FunctionalS3OperationResult {
	startTime := time.Now()
	
	validator := FunctionalS3Validation()
	
	// Validate inputs
	if err := validator.ValidateBucketName(config.bucketName); err != nil {
		return NewErrorS3Result(
			err,
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"bucket_name": config.bucketName,
				"object_key":  objectKey,
				"operation":   "get_object",
				"phase":       "validation",
			}),
		)
	}
	
	if err := validator.ValidateObjectKey(objectKey); err != nil {
		return NewErrorS3Result(
			err,
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"bucket_name": config.bucketName,
				"object_key":  objectKey,
				"operation":   "get_object",
				"phase":       "validation",
			}),
		)
	}

	// Simulate get object operation with retry logic
	retryResult := FunctionalS3RetryWithBackoff(
		func() (FunctionalS3Object, error) {
			// Simulate successful object retrieval
			result := NewFunctionalS3Object(objectKey).
				WithContent("This is the content of the S3 object").
				WithContentType("text/plain").
				WithETag("\"d41d8cd98f00b204e9800998ecf8427e\"").
				WithVersionId("null").
				WithObjectMetadata("x-amz-meta-author", "functional-s3").
				WithObjectMetadata("x-amz-meta-created", time.Now().Format(time.RFC3339))
			
			return result, nil
		},
		config.maxRetries,
		config.retryDelay,
	)
	
	totalDuration := time.Since(startTime)
	
	if retryResult.IsError() {
		return NewErrorS3Result(
			retryResult.GetError().MustGet(),
			totalDuration,
			lo.Assign(config.metadata, map[string]interface{}{
				"bucket_name": config.bucketName,
				"object_key":  objectKey,
				"operation":   "get_object",
				"phase":       "execution",
				"retries":     config.maxRetries,
			}),
		)
	}
	
	return NewSuccessS3Result(
		retryResult.GetResult().MustGet(),
		totalDuration,
		lo.Assign(config.metadata, map[string]interface{}{
			"bucket_name": config.bucketName,
			"object_key":  objectKey,
			"operation":   "get_object",
			"phase":       "success",
		}),
	)
}

// SimulateFunctionalS3ListObjects simulates listing objects in an S3 bucket
func SimulateFunctionalS3ListObjects(prefix string, config FunctionalS3Config) FunctionalS3OperationResult {
	startTime := time.Now()
	
	validator := FunctionalS3Validation()
	
	// Validate inputs
	if err := validator.ValidateBucketName(config.bucketName); err != nil {
		return NewErrorS3Result(
			err,
			time.Since(startTime),
			lo.Assign(config.metadata, map[string]interface{}{
				"bucket_name": config.bucketName,
				"operation":   "list_objects",
				"phase":       "validation",
			}),
		)
	}

	// Simulate list objects operation with retry logic
	retryResult := FunctionalS3RetryWithBackoff(
		func() ([]FunctionalS3Object, error) {
			// Simulate successful object listing
			objects := []FunctionalS3Object{
				NewFunctionalS3Object(prefix + "document1.txt").
					WithContentType("text/plain").
					WithETag("\"abc123def456\"").
					WithObjectMetadata("x-amz-meta-type", "document"),
				NewFunctionalS3Object(prefix + "image1.jpg").
					WithContentType("image/jpeg").
					WithETag("\"def456ghi789\"").
					WithObjectMetadata("x-amz-meta-type", "image"),
				NewFunctionalS3Object(prefix + "data.json").
					WithContentType("application/json").
					WithETag("\"ghi789jkl012\"").
					WithObjectMetadata("x-amz-meta-type", "data"),
			}
			
			return objects, nil
		},
		config.maxRetries,
		config.retryDelay,
	)
	
	totalDuration := time.Since(startTime)
	
	if retryResult.IsError() {
		return NewErrorS3Result(
			retryResult.GetError().MustGet(),
			totalDuration,
			lo.Assign(config.metadata, map[string]interface{}{
				"bucket_name": config.bucketName,
				"operation":   "list_objects",
				"phase":       "execution",
				"retries":     config.maxRetries,
			}),
		)
	}
	
	return NewSuccessS3Result(
		retryResult.GetResult().MustGet(),
		totalDuration,
		lo.Assign(config.metadata, map[string]interface{}{
			"bucket_name": config.bucketName,
			"operation":   "list_objects",
			"phase":       "success",
			"object_count": len(retryResult.GetResult().MustGet()),
			"prefix":       prefix,
		}),
	)
}