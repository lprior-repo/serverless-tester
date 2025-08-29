package s3

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/samber/lo"
)

// TestFunctionalS3Scenario tests a complete S3 scenario
func TestFunctionalS3Scenario(t *testing.T) {
	// Create bucket configuration
	config := NewFunctionalS3Config("test-bucket",
		WithS3Timeout(30*time.Second),
		WithS3RetryConfig(3, 200*time.Millisecond),
		WithS3Region("us-east-1"),
		WithStorageClass(types.StorageClassStandard),
		WithVersioningEnabled(true),
		WithCORSEnabled(true),
		WithS3Metadata("environment", "test"),
		WithS3Metadata("scenario", "full_workflow"),
	)

	// Step 1: Create bucket
	createResult := SimulateFunctionalS3CreateBucket(config)

	if !createResult.IsSuccess() {
		t.Errorf("Expected bucket creation to succeed, got error: %v", createResult.GetError().MustGet())
	}

	// Step 2: Put multiple objects
	objects := []FunctionalS3Object{
		NewFunctionalS3Object("documents/readme.txt", []byte("This is a test readme file")),
		NewFunctionalS3Object("images/logo.png", []byte("fake-png-data")),
		NewFunctionalS3Object("data/users.json", []byte(`{"users": [{"id": 1, "name": "Alice"}]}`)),
	}

	// Put each object
	putResults := lo.Map(objects, func(obj FunctionalS3Object, index int) FunctionalS3OperationResult {
		return SimulateFunctionalS3PutObject(obj, config)
	})

	// Verify all put operations succeeded
	allPutsSucceeded := lo.EveryBy(putResults, func(result FunctionalS3OperationResult) bool {
		return result.IsSuccess()
	})

	if !allPutsSucceeded {
		t.Error("Expected all put operations to succeed")
	}

	// Step 3: Get individual objects
	getResults := []FunctionalS3OperationResult{}
	for _, obj := range objects {
		result := SimulateFunctionalS3GetObject(obj.GetKey(), config)
		getResults = append(getResults, result)
	}

	// Verify all get operations succeeded
	allGetsSucceeded := lo.EveryBy(getResults, func(result FunctionalS3OperationResult) bool {
		return result.IsSuccess()
	})

	if !allGetsSucceeded {
		t.Error("Expected all get operations to succeed")
	}

	// Step 4: List objects with prefix
	listConfig := NewFunctionalS3Config("test-bucket",
		WithS3Timeout(30*time.Second),
		WithS3Metadata("environment", "test"),
		WithS3Metadata("scenario", "full_workflow"),
		WithS3Metadata("operation", "list_objects"),
	)

	listResult := SimulateFunctionalS3ListObjects("documents/", listConfig)

	if !listResult.IsSuccess() {
		t.Errorf("Expected list operation to succeed, got error: %v", listResult.GetError().MustGet())
	}

	// Verify list results
	if listResult.GetResult().IsPresent() {
		objects, ok := listResult.GetResult().MustGet().([]FunctionalS3Object)
		if !ok {
			t.Error("Expected list result to be []FunctionalS3Object")
		} else {
			if len(objects) == 0 {
				t.Error("Expected list operation to return objects")
			}

			// Verify each returned object has required attributes
			for i, obj := range objects {
				if obj.GetKey() == "" {
					t.Errorf("Object %d missing key", i)
				}
				if len(obj.GetContent()) == 0 {
					t.Errorf("Object %d missing content", i)
				}
			}
		}
	}

	// Step 5: Verify metadata accumulation across operations
	allMetadata := []map[string]interface{}{}
	allMetadata = append(allMetadata, createResult.GetMetadata())
	for _, result := range putResults {
		allMetadata = append(allMetadata, result.GetMetadata())
	}
	for _, result := range getResults {
		allMetadata = append(allMetadata, result.GetMetadata())
	}
	allMetadata = append(allMetadata, listResult.GetMetadata())

	// Check that all operations preserved environment metadata
	for i, metadata := range allMetadata {
		if metadata["environment"] != "test" {
			t.Errorf("Operation %d missing environment metadata, metadata: %+v", i, metadata)
		}
	}
}

// TestFunctionalS3ErrorHandling tests comprehensive error handling
func TestFunctionalS3ErrorHandling(t *testing.T) {
	// Test invalid bucket name
	invalidConfig := NewFunctionalS3Config("")
	
	result := SimulateFunctionalS3CreateBucket(invalidConfig)

	if result.IsSuccess() {
		t.Error("Expected operation to fail with invalid bucket name")
	}

	if result.GetError().IsAbsent() {
		t.Error("Expected error to be present")
	}

	// Test invalid object key
	validConfig := NewFunctionalS3Config("valid-bucket")
	invalidObject := NewFunctionalS3Object("", []byte("test content"))

	result2 := SimulateFunctionalS3PutObject(invalidObject, validConfig)

	if result2.IsSuccess() {
		t.Error("Expected operation to fail with invalid object key")
	}

	// Test invalid region
	invalidRegionConfig := NewFunctionalS3Config("valid-bucket", WithS3Region("invalid-region"))
	validObject := NewFunctionalS3Object("test-key", []byte("test content"))

	result3 := SimulateFunctionalS3PutObject(validObject, invalidRegionConfig)

	if result3.IsSuccess() {
		t.Error("Expected operation to fail with invalid region")
	}

	// Verify error metadata consistency
	errorResults := []FunctionalS3OperationResult{result, result2, result3}
	for i, errorResult := range errorResults {
		metadata := errorResult.GetMetadata()
		if metadata["phase"] != "validation" {
			t.Errorf("Error result %d should have validation phase, got %v", i, metadata["phase"])
		}
	}
}

// TestFunctionalS3ConfigurationCombinations tests various config combinations
func TestFunctionalS3ConfigurationCombinations(t *testing.T) {
	// Test minimal configuration
	minimalConfig := NewFunctionalS3Config("minimal-bucket")
	
	if minimalConfig.bucketName != "minimal-bucket" {
		t.Error("Expected minimal config to have correct bucket name")
	}
	
	if minimalConfig.timeout != FunctionalS3DefaultTimeout {
		t.Error("Expected minimal config to use default timeout")
	}

	// Test maximal configuration
	maximalConfig := NewFunctionalS3Config("maximal-bucket",
		WithS3Timeout(120*time.Second),
		WithS3RetryConfig(10, 5*time.Second),
		WithS3Region("eu-west-1"),
		WithStorageClass(types.StorageClassGlacier),
		WithServerSideEncryption(types.ServerSideEncryptionAes256),
		WithKMSKeyId("arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"),
		WithACL(types.BucketCannedACLPrivate),
		WithVersioningEnabled(true),
		WithCORSEnabled(true),
		WithS3Metadata("service", "data-service"),
		WithS3Metadata("environment", "production"),
		WithS3Metadata("version", "v3.2.1"),
	)

	// Verify all options were applied
	if maximalConfig.bucketName != "maximal-bucket" {
		t.Error("Expected maximal config bucket name")
	}

	if maximalConfig.timeout != 120*time.Second {
		t.Error("Expected maximal config timeout")
	}

	if maximalConfig.maxRetries != 10 {
		t.Error("Expected maximal config retries")
	}

	if maximalConfig.retryDelay != 5*time.Second {
		t.Error("Expected maximal config retry delay")
	}

	if maximalConfig.region != "eu-west-1" {
		t.Error("Expected maximal config region")
	}

	if maximalConfig.storageClass != types.StorageClassGlacier {
		t.Error("Expected maximal config storage class")
	}

	if maximalConfig.serverSideEncryption.IsAbsent() {
		t.Error("Expected maximal config server side encryption")
	}

	if maximalConfig.kmsKeyId.IsAbsent() {
		t.Error("Expected maximal config KMS key ID")
	}

	if maximalConfig.acl.IsAbsent() {
		t.Error("Expected maximal config ACL")
	}

	if !maximalConfig.versioningEnabled {
		t.Error("Expected maximal config versioning enabled")
	}

	if !maximalConfig.corsEnabled {
		t.Error("Expected maximal config CORS enabled")
	}

	if len(maximalConfig.metadata) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(maximalConfig.metadata))
	}
}

// TestFunctionalS3ObjectManipulation tests advanced object operations
func TestFunctionalS3ObjectManipulation(t *testing.T) {
	// Create objects with different content types
	textObject := NewFunctionalS3Object("documents/readme.txt", []byte("This is a text file"))
	jsonObject := NewFunctionalS3Object("data/config.json", []byte(`{"version": "1.0", "debug": true}`))
	binaryObject := NewFunctionalS3Object("images/photo.jpg", make([]byte, 1024)) // 1KB of binary data

	objects := []FunctionalS3Object{textObject, jsonObject, binaryObject}

	// Test object properties
	for i, obj := range objects {
		if obj.GetKey() == "" {
			t.Errorf("Object %d should have non-empty key", i)
		}

		if len(obj.GetContent()) == 0 {
			t.Errorf("Object %d should have non-empty content", i)
		}

		if obj.GetSize() != int64(len(obj.GetContent())) {
			t.Errorf("Object %d size mismatch: expected %d, got %d", 
				i, len(obj.GetContent()), obj.GetSize())
		}

		if obj.GetLastModified().IsZero() {
			t.Errorf("Object %d should have last modified time", i)
		}
	}

	// Test object operations with various objects
	config := NewFunctionalS3Config("object-test-bucket",
		WithS3Metadata("test_type", "object_manipulation"),
	)

	for i, obj := range objects {
		result := SimulateFunctionalS3PutObject(obj, config)

		if !result.IsSuccess() {
			t.Errorf("Expected object %d put to succeed, got error: %v", i, result.GetError().MustGet())
		}

		// Verify result includes object information
		if result.GetResult().IsPresent() {
			resultMap, ok := result.GetResult().MustGet().(map[string]interface{})
			if ok {
				if resultMap["size"] != obj.GetSize() {
					t.Errorf("Expected object %d size in result to match", i)
				}
				if resultMap["key"] != obj.GetKey() {
					t.Errorf("Expected object %d key in result to match", i)
				}
			}
		}
	}
}

// TestFunctionalS3ValidationRules tests S3-specific validation
func TestFunctionalS3ValidationRules(t *testing.T) {
	// Test bucket name validation
	invalidBucketNames := []string{
		"",                    // Empty
		"a",                   // Too short
		"UPPERCASE",           // Contains uppercase
		"bucket_with_underscores", // Contains underscores
		"bucket..double.dots", // Double dots
		"bucket-",             // Ends with hyphen
		"192.168.1.1",        // IP address format
	}

	for _, bucketName := range invalidBucketNames {
		config := NewFunctionalS3Config(bucketName)
		result := SimulateFunctionalS3CreateBucket(config)
		
		if result.IsSuccess() {
			t.Errorf("Expected bucket creation to fail for invalid name: %s", bucketName)
		}
	}

	// Test valid bucket names
	validBucketNames := []string{
		"valid-bucket",
		"test.bucket.com",
		"bucket123",
		"my-awesome-bucket-2024",
	}

	for _, bucketName := range validBucketNames {
		config := NewFunctionalS3Config(bucketName)
		result := SimulateFunctionalS3CreateBucket(config)
		
		if !result.IsSuccess() {
			t.Errorf("Expected bucket creation to succeed for valid name: %s", bucketName)
		}
	}

	// Test object key validation
	invalidObjectKeys := []string{
		"",                // Empty
		"key//double/slash", // Double slash
		"key/with/trailing/slash/", // Trailing slash
	}

	config := NewFunctionalS3Config("valid-bucket")
	
	for _, objectKey := range invalidObjectKeys {
		obj := NewFunctionalS3Object(objectKey, []byte("test content"))
		result := SimulateFunctionalS3PutObject(obj, config)
		
		if result.IsSuccess() {
			t.Errorf("Expected put operation to fail for invalid key: %s", objectKey)
		}
	}
}

// TestFunctionalS3PerformanceMetrics tests performance tracking
func TestFunctionalS3PerformanceMetrics(t *testing.T) {
	config := NewFunctionalS3Config("performance-bucket",
		WithS3Timeout(1*time.Second),
		WithS3RetryConfig(1, 50*time.Millisecond), // Faster for testing
		WithS3Metadata("test_type", "performance"),
	)

	obj := NewFunctionalS3Object("performance/test-object.txt", []byte("performance test data"))

	startTime := time.Now()
	result := SimulateFunctionalS3PutObject(obj, config)
	operationTime := time.Since(startTime)

	if !result.IsSuccess() {
		t.Errorf("Expected performance test to succeed, got error: %v", result.GetError().MustGet())
	}

	// Verify duration tracking
	recordedDuration := result.GetDuration()
	if recordedDuration == 0 {
		t.Error("Expected duration to be recorded")
	}

	// Duration should be reasonable (less than total operation time)
	if recordedDuration > operationTime {
		t.Error("Recorded duration should not exceed total operation time")
	}

	// Test multiple operations for performance consistency
	results := []FunctionalS3OperationResult{}
	for i := 0; i < 5; i++ {
		testObj := NewFunctionalS3Object(fmt.Sprintf("perf/object-%03d.txt", i), 
			[]byte(fmt.Sprintf("test data %d", i)))
		
		result := SimulateFunctionalS3PutObject(testObj, config)
		results = append(results, result)
	}

	// All operations should succeed
	allSucceeded := lo.EveryBy(results, func(r FunctionalS3OperationResult) bool {
		return r.IsSuccess()
	})

	if !allSucceeded {
		t.Error("Expected all performance test operations to succeed")
	}

	// All operations should have duration recorded
	allHaveDuration := lo.EveryBy(results, func(r FunctionalS3OperationResult) bool {
		return r.GetDuration() > 0
	})

	if !allHaveDuration {
		t.Error("Expected all operations to have duration recorded")
	}
}

// BenchmarkFunctionalS3Scenario benchmarks a complete scenario
func BenchmarkFunctionalS3Scenario(b *testing.B) {
	config := NewFunctionalS3Config("benchmark-bucket",
		WithS3Timeout(5*time.Second),
		WithS3RetryConfig(1, 10*time.Millisecond),
	)

	obj := NewFunctionalS3Object("benchmark/test-object.txt", []byte("Benchmark test data"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate full S3 workflow
		createResult := SimulateFunctionalS3CreateBucket(config)
		if !createResult.IsSuccess() {
			b.Errorf("Create bucket operation failed: %v", createResult.GetError().MustGet())
		}

		putResult := SimulateFunctionalS3PutObject(obj, config)
		if !putResult.IsSuccess() {
			b.Errorf("Put operation failed: %v", putResult.GetError().MustGet())
		}

		getResult := SimulateFunctionalS3GetObject(obj.GetKey(), config)
		if !getResult.IsSuccess() {
			b.Errorf("Get operation failed: %v", getResult.GetError().MustGet())
		}

		listResult := SimulateFunctionalS3ListObjects("benchmark/", config)
		if !listResult.IsSuccess() {
			b.Errorf("List operation failed: %v", listResult.GetError().MustGet())
		}
	}
}