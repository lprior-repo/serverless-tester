package stepfunctions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test CreateActivity functionality using TDD methodology
// Focus on validation logic that can be tested without mocking AWS SDK

func TestCreateActivityE_ValidationSuccess(t *testing.T) {
	// Test that validation passes for valid inputs
	
	// Valid activity names
	validNames := []string{
		"test-activity",
		"MyActivity123",
		"activity_with_underscores",
		"activity-with-dashes",
		"a", // minimum length
	}
	
	for _, name := range validNames {
		err := validateActivityName(name)
		assert.NoError(t, err, "Expected valid name %s to pass validation", name)
	}
}

func TestCreateActivityE_ValidationFailures(t *testing.T) {
	// Test validation failures for invalid inputs
	
	testCases := []struct {
		name          string
		activityName  string
		expectedError string
	}{
		{
			name:          "empty name",
			activityName:  "",
			expectedError: "invalid activity name",
		},
		{
			name:          "name too long",
			activityName:  "a" + string(make([]byte, 81)), // 82 characters
			expectedError: "name too long",
		},
		{
			name:          "invalid characters",
			activityName:  "activity with spaces",
			expectedError: "invalid characters",
		},
		{
			name:          "invalid special characters",
			activityName:  "activity@invalid",
			expectedError: "invalid characters",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateActivityName(tc.activityName)
			assert.Error(t, err, "Expected validation to fail for %s", tc.activityName)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestValidateActivityArn_Success(t *testing.T) {
	// Test ARN validation for valid inputs
	
	validArns := []string{
		"arn:aws:states:us-east-1:123456789012:activity:my-activity",
		"arn:aws:states:us-west-2:987654321098:activity:test_activity",
		"arn:aws:states:eu-west-1:123456789012:activity:activity-name",
	}
	
	for _, arn := range validArns {
		err := validateActivityArn(arn)
		assert.NoError(t, err, "Expected valid ARN %s to pass validation", arn)
	}
}

func TestValidateActivityArn_Failures(t *testing.T) {
	// Test ARN validation failures
	
	testCases := []struct {
		name          string
		arn           string
		expectedError string
	}{
		{
			name:          "empty ARN",
			arn:           "",
			expectedError: "activity ARN is required",
		},
		{
			name:          "invalid prefix",
			arn:           "arn:aws:lambda:us-east-1:123456789012:activity:my-activity",
			expectedError: "must start with 'arn:aws:states:'",
		},
		{
			name:          "missing activity segment",
			arn:           "arn:aws:states:us-east-1:123456789012:stateMachine:my-sm",
			expectedError: "must contain ':activity:'",
		},
		{
			name:          "malformed ARN",
			arn:           "invalid-arn",
			expectedError: "must start with 'arn:aws:states:'",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateActivityArn(tc.arn)
			assert.Error(t, err, "Expected validation to fail for %s", tc.arn)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestValidateTaskToken_Success(t *testing.T) {
	// Test task token validation for valid inputs
	
	validTokens := []string{
		"AAAAKgAAAAIAAAAAAAAAASdKaZC1WL4bjhzCOw7z6aXG6kUzCwAcPcMWL9dD6McjE/H0o2nxF2bR3TgF2g==",
		"AAAAKgAAAAIAAAAAAAAAAQ6kSGOq4dXp/6N2Jtu8B13s8dDRkmJ-B5YhKJ1N3fy5X7n2V7Yk2pLjJUQjUw==",
		"short-token-for-test", // Minimum length test
	}
	
	for _, token := range validTokens {
		err := validateTaskToken(token)
		assert.NoError(t, err, "Expected valid token to pass validation")
	}
}

func TestValidateTaskToken_Failures(t *testing.T) {
	// Test task token validation failures
	
	testCases := []struct {
		name          string
		token         string
		expectedError string
	}{
		{
			name:          "empty token",
			token:         "",
			expectedError: "task token is required",
		},
		{
			name:          "token too short",
			token:         "short",
			expectedError: "token too short",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTaskToken(tc.token)
			assert.Error(t, err, "Expected validation to fail for %s", tc.token)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestValidateMapRunArn_Success(t *testing.T) {
	// Test Map Run ARN validation for valid inputs
	
	validArns := []string{
		"arn:aws:states:us-east-1:123456789012:mapRun:my-execution:my-map-run",
		"arn:aws:states:us-west-2:987654321098:mapRun:test-exec:map-run-123",
		"arn:aws:states:eu-west-1:123456789012:mapRun:execution_name:map_run_name",
	}
	
	for _, arn := range validArns {
		err := validateMapRunArn(arn)
		assert.NoError(t, err, "Expected valid Map Run ARN %s to pass validation", arn)
	}
}

func TestValidateMapRunArn_Failures(t *testing.T) {
	// Test Map Run ARN validation failures
	
	testCases := []struct {
		name          string
		arn           string
		expectedError string
	}{
		{
			name:          "empty ARN",
			arn:           "",
			expectedError: "map run ARN is required",
		},
		{
			name:          "invalid prefix",
			arn:           "arn:aws:lambda:us-east-1:123456789012:mapRun:exec:map",
			expectedError: "must start with 'arn:aws:states:'",
		},
		{
			name:          "missing mapRun segment",
			arn:           "arn:aws:states:us-east-1:123456789012:execution:my-exec:map",
			expectedError: "must contain ':mapRun:'",
		},
		{
			name:          "malformed ARN",
			arn:           "invalid-arn",
			expectedError: "must start with 'arn:aws:states:'",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMapRunArn(tc.arn)
			assert.Error(t, err, "Expected validation to fail for %s", tc.arn)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}