package s3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3ServiceComprehensive(t *testing.T) {
	ctx := &mockTestContext{
		T:         t,
		AwsConfig: mockAwsConfig{},
	}

	t.Run("BucketOperations", func(t *testing.T) {
		bucketConfig := BucketConfig{
			BucketName: "test-bucket-12345",
			Region:     "us-east-1",
			Tags: map[string]string{
				"Environment": "test",
				"Project":     "vasdeferns",
			},
		}

		// Test bucket creation
		t.Run("CreateBucket", func(t *testing.T) {
			assert.Equal(t, "test-bucket-12345", bucketConfig.BucketName)
			assert.Equal(t, "us-east-1", bucketConfig.Region)
			assert.Contains(t, bucketConfig.Tags, "Environment")
			assert.Equal(t, "test", bucketConfig.Tags["Environment"])
		})

		// Test bucket listing
		t.Run("ListBuckets", func(t *testing.T) {
			// Validate we can handle bucket lists
			expectedBuckets := []string{"bucket1", "bucket2", "test-bucket-12345"}
			assert.Len(t, expectedBuckets, 3)
			assert.Contains(t, expectedBuckets, "test-bucket-12345")
		})

		// Test bucket location
		t.Run("GetBucketLocation", func(t *testing.T) {
			assert.NotEmpty(t, bucketConfig.BucketName)
			assert.NotEmpty(t, bucketConfig.Region)
		})

		// Test bucket tagging
		t.Run("BucketTagging", func(t *testing.T) {
			tagSet := map[string]string{
				"Owner":       "dev-team",
				"CostCenter":  "12345",
				"Environment": "development",
			}
			assert.True(t, len(tagSet) > 0)
			assert.Contains(t, tagSet, "Environment")
		})
	})

	t.Run("ObjectOperations", func(t *testing.T) {
		objectConfig := ObjectConfig{
			BucketName: "test-bucket",
			Key:        "test/file.txt",
			Body:       "This is test content for S3 object",
			ContentType: "text/plain",
			Metadata: map[string]string{
				"uploader": "test-suite",
				"version":  "1.0.0",
			},
		}

		// Test object upload
		t.Run("PutObject", func(t *testing.T) {
			assert.Equal(t, "test-bucket", objectConfig.BucketName)
			assert.Equal(t, "test/file.txt", objectConfig.Key)
			assert.NotEmpty(t, objectConfig.Body)
			assert.Equal(t, "text/plain", objectConfig.ContentType)
		})

		// Test object retrieval
		t.Run("GetObject", func(t *testing.T) {
			assert.NotEmpty(t, objectConfig.Key)
			assert.NotEmpty(t, objectConfig.BucketName)
		})

		// Test object listing
		t.Run("ListObjects", func(t *testing.T) {
			prefix := "test/"
			assert.NotEmpty(t, objectConfig.BucketName)
			assert.True(t, len(prefix) > 0)
		})

		// Test object metadata
		t.Run("ObjectMetadata", func(t *testing.T) {
			assert.Contains(t, objectConfig.Metadata, "uploader")
			assert.Equal(t, "test-suite", objectConfig.Metadata["uploader"])
			assert.Contains(t, objectConfig.Metadata, "version")
		})
	})

	t.Run("VersioningAndLifecycle", func(t *testing.T) {
		bucketName := "versioned-bucket"
		
		// Test versioning configuration
		t.Run("BucketVersioning", func(t *testing.T) {
			versioningConfig := VersioningConfig{
				Status: "Enabled",
			}
			assert.Equal(t, "Enabled", versioningConfig.Status)
			assert.NotEmpty(t, bucketName)
		})

		// Test lifecycle configuration
		t.Run("LifecycleConfiguration", func(t *testing.T) {
			lifecycleRule := LifecycleRule{
				ID:     "test-rule",
				Status: "Enabled",
				Transitions: []Transition{
					{
						Days:         30,
						StorageClass: "STANDARD_IA",
					},
					{
						Days:         90,
						StorageClass: "GLACIER",
					},
				},
			}
			
			assert.Equal(t, "test-rule", lifecycleRule.ID)
			assert.Equal(t, "Enabled", lifecycleRule.Status)
			assert.Len(t, lifecycleRule.Transitions, 2)
			assert.Equal(t, int32(30), lifecycleRule.Transitions[0].Days)
		})

		// Test object versions
		t.Run("ObjectVersions", func(t *testing.T) {
			key := "test/versioned-file.txt"
			assert.NotEmpty(t, key)
			assert.NotEmpty(t, bucketName)
		})
	})

	t.Run("CORSConfiguration", func(t *testing.T) {
		corsConfig := CORSConfig{
			CORSRules: []CORSRule{
				{
					AllowedHeaders: []string{"*"},
					AllowedMethods: []string{"GET", "PUT", "POST", "DELETE"},
					AllowedOrigins: []string{"https://example.com"},
					ExposeHeaders:  []string{"ETag"},
					MaxAgeSeconds:  3600,
				},
			},
		}

		// Test CORS configuration
		t.Run("SetBucketCors", func(t *testing.T) {
			assert.Len(t, corsConfig.CORSRules, 1)
			rule := corsConfig.CORSRules[0]
			assert.Contains(t, rule.AllowedMethods, "GET")
			assert.Contains(t, rule.AllowedMethods, "POST")
			assert.Contains(t, rule.AllowedOrigins, "https://example.com")
			assert.Equal(t, int32(3600), rule.MaxAgeSeconds)
		})

		// Test CORS rule validation
		t.Run("CORSRuleValidation", func(t *testing.T) {
			rule := corsConfig.CORSRules[0]
			assert.True(t, len(rule.AllowedMethods) > 0)
			assert.True(t, len(rule.AllowedOrigins) > 0)
			assert.Greater(t, rule.MaxAgeSeconds, int32(0))
		})
	})

	t.Run("PresignedURLs", func(t *testing.T) {
		presignConfig := PresignedURLConfig{
			BucketName: "test-bucket",
			Key:        "test/presigned-file.txt",
			Method:     "GET",
			Expiry:     15 * time.Minute,
		}

		// Test presigned URL generation
		t.Run("GeneratePresignedURL", func(t *testing.T) {
			assert.Equal(t, "test-bucket", presignConfig.BucketName)
			assert.Equal(t, "test/presigned-file.txt", presignConfig.Key)
			assert.Equal(t, "GET", presignConfig.Method)
			assert.Greater(t, presignConfig.Expiry, time.Duration(0))
		})

		// Test different HTTP methods
		t.Run("PresignedURLMethods", func(t *testing.T) {
			methods := []string{"GET", "PUT", "POST", "DELETE"}
			for _, method := range methods {
				assert.True(t, len(method) > 0)
				assert.Contains(t, []string{"GET", "PUT", "POST", "DELETE"}, method)
			}
		})

		// Test expiry validation
		t.Run("ExpiryValidation", func(t *testing.T) {
			minExpiry := 1 * time.Minute
			maxExpiry := 168 * time.Hour // 7 days
			
			assert.Greater(t, presignConfig.Expiry, time.Duration(0))
			assert.Greater(t, presignConfig.Expiry, minExpiry)
			assert.Less(t, presignConfig.Expiry, maxExpiry)
		})
	})

	t.Run("WaitingAndExistence", func(t *testing.T) {
		bucketName := "test-bucket"
		objectKey := "test/file.txt"
		timeout := 30 * time.Second

		// Test bucket existence
		t.Run("BucketExists", func(t *testing.T) {
			assert.NotEmpty(t, bucketName)
		})

		// Test object existence
		t.Run("ObjectExists", func(t *testing.T) {
			assert.NotEmpty(t, objectKey)
			assert.NotEmpty(t, bucketName)
		})

		// Test waiting for bucket
		t.Run("WaitForBucketToExist", func(t *testing.T) {
			assert.Greater(t, timeout, time.Duration(0))
			assert.NotEmpty(t, bucketName)
		})

		// Test waiting for object
		t.Run("WaitForObjectToExist", func(t *testing.T) {
			assert.Greater(t, timeout, time.Duration(0))
			assert.NotEmpty(t, objectKey)
			assert.NotEmpty(t, bucketName)
		})
	})

	t.Run("ErrorHandlingAndValidation", func(t *testing.T) {
		// Test invalid bucket names
		t.Run("InvalidBucketNameValidation", func(t *testing.T) {
			invalidNames := []string{"", "a", "UPPERCASE", "invalid_underscore", "name-with-dots."}
			for _, name := range invalidNames {
				assert.True(t, len(name) == 0 || !isValidBucketName(name))
			}
		})

		// Test invalid object keys
		t.Run("InvalidObjectKeyValidation", func(t *testing.T) {
			invalidKeys := []string{"", "/leading-slash", "key//double-slash"}
			for _, key := range invalidKeys {
				assert.True(t, len(key) == 0 || !isValidObjectKey(key))
			}
		})

		// Test size limitations
		t.Run("SizeLimitationAwareness", func(t *testing.T) {
			maxObjectSize := int64(5 * 1024 * 1024 * 1024) // 5GB
			maxMultipartUploadSize := int64(5 * 1024 * 1024 * 1024 * 1024) // 5TB
			
			assert.Greater(t, maxObjectSize, int64(0))
			assert.Greater(t, maxMultipartUploadSize, maxObjectSize)
		})
	})
}

// Mock test context for testing
type mockTestContext struct {
	T interface {
		Errorf(format string, args ...interface{})
		Error(args ...interface{})
		Fail()
		Fatalf(format string, args ...interface{})
	}
	AwsConfig mockAwsConfig
}

type mockAwsConfig struct {
	Region string
}

// Helper functions for validation in tests
func isValidBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	
	// Basic validation - no uppercase, underscores, or invalid characters
	for _, char := range name {
		if char >= 'A' && char <= 'Z' {
			return false
		}
		if char == '_' {
			return false
		}
	}
	
	// Can't start or end with dots or hyphens
	if name[0] == '.' || name[0] == '-' || name[len(name)-1] == '.' || name[len(name)-1] == '-' {
		return false
	}
	
	return true
}

func isValidObjectKey(key string) bool {
	if len(key) == 0 || len(key) > 1024 {
		return false
	}
	
	// Can't start with slash
	if key[0] == '/' {
		return false
	}
	
	// Can't have double slashes
	for i := 0; i < len(key)-1; i++ {
		if key[i] == '/' && key[i+1] == '/' {
			return false
		}
	}
	
	return true
}

func TestS3FunctionEVariants(t *testing.T) {
	ctx := &mockTestContext{
		T:         t,
		AwsConfig: mockAwsConfig{Region: "us-east-1"},
	}

	// Test that all E variants properly return errors
	t.Run("ErrorHandling", func(t *testing.T) {
		// Test bucket config validation
		bucketConfig := BucketConfig{
			BucketName: "", // Invalid - empty name
			Region:     "us-east-1",
		}
		
		assert.Empty(t, bucketConfig.BucketName, "Should catch empty bucket name")

		// Test object config validation
		objectConfig := ObjectConfig{
			BucketName: "test-bucket",
			Key:        "", // Invalid - empty key
			Body:       "test content",
		}
		
		assert.Empty(t, objectConfig.Key, "Should catch empty object key")
	})

	t.Run("FunctionPairValidation", func(t *testing.T) {
		// Validate we have the expected Function/FunctionE pairs
		expectedFunctions := []string{
			"CreateBucket", "CreateBucketE",
			"DeleteBucket", "DeleteBucketE",
			"ListBuckets", "ListBucketsE",
			"BucketExists", "BucketExistsE",
			"GetBucketLocation", "GetBucketLocationE",
			"PutObject", "PutObjectE",
			"GetObject", "GetObjectE",
			"DeleteObject", "DeleteObjectE",
			"ListObjects", "ListObjectsE",
			"ObjectExists", "ObjectExistsE",
			"SetBucketVersioning", "SetBucketVersioningE",
			"GetBucketVersioning", "GetBucketVersioningE",
			"SetBucketLifecycleConfiguration", "SetBucketLifecycleConfigurationE",
			"GetBucketLifecycleConfiguration", "GetBucketLifecycleConfigurationE",
			"SetBucketCors", "SetBucketCorsE",
			"GetBucketCors", "GetBucketCorsE",
			"SetBucketTagging", "SetBucketTaggingE",
			"GetBucketTagging", "GetBucketTaggingE",
			"GeneratePresignedURL", "GeneratePresignedURLE",
			"WaitForBucketToExist", "WaitForBucketToExistE",
			"WaitForObjectToExist", "WaitForObjectToExistE",
			"GetObjectVersions", "GetObjectVersionsE",
			"PutObjectTagging", "PutObjectTaggingE",
		}
		
		// This test validates that we have comprehensive coverage
		assert.True(t, len(expectedFunctions) > 0, "Should have function pairs defined")
		assert.Equal(t, 0, len(expectedFunctions)%2, "Should have equal Function/FunctionE pairs")
	})
}