package snapshot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTestingTFail implements vasdeference.TestingT but tracks calls without panicking
type MockTestingTFail struct {
	errorCalled bool
	failCalled  bool
	testName    string
}

func (m *MockTestingTFail) Helper()                                        {}
func (m *MockTestingTFail) Errorf(format string, args ...interface{})     { m.errorCalled = true }
func (m *MockTestingTFail) Error(args ...interface{})                     { m.errorCalled = true }
func (m *MockTestingTFail) Fail()                                          { m.failCalled = true }
func (m *MockTestingTFail) FailNow()                                       { m.failCalled = true }
func (m *MockTestingTFail) Fatal(args ...interface{})                     { m.failCalled = true }
func (m *MockTestingTFail) Fatalf(format string, args ...interface{})     { m.failCalled = true }
func (m *MockTestingTFail) Logf(format string, args ...interface{})       {}
func (m *MockTestingTFail) Name() string                                   { return m.testName }

// TestMatchToBytesErrorCoverage specifically tests lines 86-89 in Match method
func TestMatchToBytesErrorCoverage(t *testing.T) {
	tempDir := t.TempDir()
	
	mockT := &MockTestingTFail{testName: "TestMatchToBytesErrorCoverage"}
	opts := Options{SnapshotDir: tempDir}
	s := New(mockT, opts)
	
	// Use ErrorReader to trigger toBytes error
	errorReader := &ErrorReader{}
	
	// This should hit lines 86-89 (toBytes error path in Match)
	s.Match("error_test", errorReader)
	
	// Verify the error path was taken
	assert.True(t, mockT.errorCalled, "Expected Errorf to be called")
	assert.True(t, mockT.failCalled, "Expected FailNow to be called")
}

// TestMatchSnapshotMismatchCoverage specifically tests lines 118-121 in Match method
func TestMatchSnapshotMismatchCoverage(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create existing snapshot with different content
	snapshotFile := filepath.Join(tempDir, "TestMatchSnapshotMismatchCoverage_mismatch_test.snap")
	err := os.WriteFile(snapshotFile, []byte("original content"), 0644)
	require.NoError(t, err)
	
	mockT := &MockTestingTFail{testName: "TestMatchSnapshotMismatchCoverage"}
	opts := Options{SnapshotDir: tempDir}
	s := New(mockT, opts)
	
	// This should hit lines 118-121 (mismatch detection in Match)
	s.Match("mismatch_test", "different content")
	
	// Verify the mismatch error path was taken
	assert.True(t, mockT.errorCalled, "Expected mismatch error to be reported")
}

// TestMatchESanitizerCoverage specifically tests lines 391-393 in MatchE method
func TestMatchESanitizerCoverage(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a sanitizer to ensure the loop is executed
	sanitizer := func(data []byte) []byte {
		return append(data, []byte(" - sanitized")...)
	}
	
	mockT := &MockTestingTFail{testName: "TestMatchESanitizerCoverage"}
	opts := Options{
		SnapshotDir: tempDir,
		Sanitizers:  []Sanitizer{sanitizer},
	}
	s := New(mockT, opts)
	
	// This should hit lines 391-393 (sanitizer loop in MatchE)
	err := s.MatchE("sanitizer_test", "test content")
	assert.NoError(t, err, "MatchE should succeed with sanitizer")
	
	// Verify the sanitizer was applied by reading the file
	content, err := os.ReadFile(filepath.Join(tempDir, "TestMatchESanitizerCoverage_sanitizer_test.snap"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "sanitized")
}

// TestMatchEReadErrorCoverage specifically tests lines 409-411 in MatchE method
func TestMatchEReadErrorCoverage(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create file with permission error
	snapshotFile := filepath.Join(tempDir, "TestMatchEReadErrorCoverage_error_test.snap")
	err := os.WriteFile(snapshotFile, []byte("test content"), 0644)
	require.NoError(t, err)
	
	// Make file unreadable
	err = os.Chmod(snapshotFile, 0000)
	require.NoError(t, err)
	defer os.Chmod(snapshotFile, 0644)
	
	mockT := &MockTestingTFail{testName: "TestMatchEReadErrorCoverage"}
	opts := Options{
		SnapshotDir:     tempDir,
		UpdateSnapshots: false, // Force read attempt
	}
	s := New(mockT, opts)
	
	// This should hit lines 409-411 (read error in MatchE)
	err = s.MatchE("error_test", "test content")
	assert.Error(t, err, "Expected read error")
	assert.Contains(t, err.Error(), "failed to read snapshot")
}

// TestAdditionalCoverageBoosts tests additional methods to push coverage over 98%
func TestAdditionalCoverageBoosts(t *testing.T) {
	tempDir := t.TempDir()
	opts := Options{SnapshotDir: tempDir}
	s := New(t, opts)
	
	// Test MatchJSON to increase coverage
	data := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}
	s.MatchJSON("json_test", data)
	
	// Test MatchLambdaResponse with actual JSON bytes
	responseBytes, _ := os.ReadFile(filepath.Join(tempDir, "TestAdditionalCoverageBoosts_json_test.snap"))
	if responseBytes != nil {
		s.MatchLambdaResponse("lambda_test", responseBytes)
	}
	
	// Test MatchDynamoDBItems
	items := []map[string]interface{}{
		{"id": "1", "name": "item1"},
		{"id": "2", "name": "item2"},
	}
	s.MatchDynamoDBItems("dynamo_test", items)
	
	// Test MatchStepFunctionExecution
	execution := map[string]interface{}{
		"executionArn": "arn:aws:states:us-east-1:123456789012:execution:MyStateMachine:exec-1",
		"status":       "SUCCEEDED",
		"startDate":    "2023-01-01T12:00:00.000Z",
	}
	s.MatchStepFunctionExecution("step_func_test", execution)
	
	// Test generateDiff method by creating a mismatch scenario with mock
	mockT := &MockTestingTFail{testName: "TestAdditionalCoverageBoosts"}
	s2 := New(mockT, opts)
	
	// Create existing snapshot
	existingFile := filepath.Join(tempDir, "TestAdditionalCoverageBoosts_diff_test.snap")
	err := os.WriteFile(existingFile, []byte("original content"), 0644)
	require.NoError(t, err)
	
	s2.Match("diff_test", "modified content")
	assert.True(t, mockT.errorCalled, "Expected diff error")
}

// TestUpdateSnapshotErrorPath tests error handling in updateSnapshot
func TestUpdateSnapshotErrorPath(t *testing.T) {
	// Create read-only directory to cause write error
	tempDir := t.TempDir()
	err := os.Chmod(tempDir, 0555)
	require.NoError(t, err)
	defer os.Chmod(tempDir, 0755)
	
	mockT := &MockTestingTFail{testName: "TestUpdateSnapshotErrorPath"}
	opts := Options{
		SnapshotDir:     tempDir,
		UpdateSnapshots: true, // Force update attempt
	}
	s := New(mockT, opts)
	
	// This should trigger updateSnapshot error path
	s.Match("write_error_test", "test content")
	
	assert.True(t, mockT.errorCalled, "Expected write error to be reported")
	assert.True(t, mockT.failCalled, "Expected FailNow to be called")
}