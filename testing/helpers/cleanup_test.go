package helpers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// RED: Test that CleanupManager can register and execute cleanup functions
func TestCleanupManager_ShouldRegisterAndExecuteCleanupFunctions(t *testing.T) {
	// Arrange
	cleanup := NewCleanupManager()
	executed := false
	
	cleanupFunc := func() error {
		executed = true
		return nil
	}
	
	// Act
	cleanup.Register("test-cleanup", cleanupFunc)
	err := cleanup.ExecuteAll()
	
	// Assert
	assert.NoError(t, err, "ExecuteAll should succeed")
	assert.True(t, executed, "Cleanup function should be executed")
}

// RED: Test that CleanupManager executes cleanup functions in reverse order (LIFO)
func TestCleanupManager_ShouldExecuteCleanupInReverseOrder(t *testing.T) {
	// Arrange
	cleanup := NewCleanupManager()
	executed := make([]string, 0)
	
	cleanup.Register("first", func() error {
		executed = append(executed, "first")
		return nil
	})
	
	cleanup.Register("second", func() error {
		executed = append(executed, "second")
		return nil
	})
	
	cleanup.Register("third", func() error {
		executed = append(executed, "third")
		return nil
	})
	
	// Act
	err := cleanup.ExecuteAll()
	
	// Assert
	assert.NoError(t, err, "ExecuteAll should succeed")
	assert.Equal(t, []string{"third", "second", "first"}, executed, "Cleanup functions should execute in reverse order")
}

// RED: Test that CleanupManager continues executing even if one cleanup fails
func TestCleanupManager_ShouldContinueExecutionOnFailure(t *testing.T) {
	// Arrange
	cleanup := NewCleanupManager()
	executed := make([]string, 0)
	
	cleanup.Register("first", func() error {
		executed = append(executed, "first")
		return nil
	})
	
	cleanup.Register("failing", func() error {
		executed = append(executed, "failing")
		return errors.New("cleanup failed")
	})
	
	cleanup.Register("third", func() error {
		executed = append(executed, "third")
		return nil
	})
	
	// Act
	err := cleanup.ExecuteAll()
	
	// Assert
	assert.Error(t, err, "ExecuteAll should return error")
	assert.Contains(t, err.Error(), "cleanup failed", "Error should contain original failure")
	assert.Equal(t, []string{"third", "failing", "first"}, executed, "All cleanup functions should execute despite failures")
}

// RED: Test that CleanupManager can be reset
func TestCleanupManager_ShouldAllowReset(t *testing.T) {
	// Arrange
	cleanup := NewCleanupManager()
	executed := false
	
	cleanup.Register("test", func() error {
		executed = true
		return nil
	})
	
	// Act
	cleanup.Reset()
	err := cleanup.ExecuteAll()
	
	// Assert
	assert.NoError(t, err, "ExecuteAll should succeed after reset")
	assert.False(t, executed, "Cleanup function should not execute after reset")
}

// RED: Test that CleanupManager can get count of registered cleanups
func TestCleanupManager_ShouldProvideCleanupCount(t *testing.T) {
	// Arrange
	cleanup := NewCleanupManager()
	
	// Act & Assert
	assert.Equal(t, 0, cleanup.Count(), "Initial count should be zero")
	
	cleanup.Register("first", func() error { return nil })
	assert.Equal(t, 1, cleanup.Count(), "Count should be 1 after registering one cleanup")
	
	cleanup.Register("second", func() error { return nil })
	assert.Equal(t, 2, cleanup.Count(), "Count should be 2 after registering two cleanups")
	
	cleanup.Reset()
	assert.Equal(t, 0, cleanup.Count(), "Count should be zero after reset")
}

// RED: Test that CleanupManager can check if it has any cleanups registered
func TestCleanupManager_ShouldIndicateIfEmpty(t *testing.T) {
	// Arrange
	cleanup := NewCleanupManager()
	
	// Act & Assert
	assert.True(t, cleanup.IsEmpty(), "Should be empty initially")
	
	cleanup.Register("test", func() error { return nil })
	assert.False(t, cleanup.IsEmpty(), "Should not be empty after registering cleanup")
	
	cleanup.Reset()
	assert.True(t, cleanup.IsEmpty(), "Should be empty after reset")
}

// RED: Test that DeferCleanup registers cleanup with t.Cleanup when using real testing.T
func TestDeferCleanup_ShouldRegisterWithTestingT(t *testing.T) {
	// Arrange
	executed := false
	cleanupFunc := func() error {
		executed = true
		return nil
	}
	
	// Act
	DeferCleanup(t, "test-cleanup", cleanupFunc)
	
	// Since we can't directly test t.Cleanup execution in a unit test,
	// we'll test that the function doesn't panic and accepts the parameters
	// The actual cleanup execution will be tested in integration tests
	
	// Assert - just verify the function call succeeds
	assert.False(t, executed, "Cleanup function should not execute immediately")
}

// RED: Test that TerraformCleanupWrapper wraps Terraform-specific cleanup patterns
func TestTerraformCleanupWrapper_ShouldWrapTerraformCleanup(t *testing.T) {
	// Arrange
	cleanup := NewCleanupManager()
	destroyCalled := false
	
	terraformCleanup := func() {
		destroyCalled = true
	}
	
	// Act
	wrappedCleanup := TerraformCleanupWrapper("test-terraform", terraformCleanup)
	cleanup.Register("terraform-test", wrappedCleanup)
	err := cleanup.ExecuteAll()
	
	// Assert
	assert.NoError(t, err, "Wrapped Terraform cleanup should succeed")
	assert.True(t, destroyCalled, "Terraform destroy should be called")
}

// RED: Test that AWSResourceCleanupWrapper wraps AWS-specific cleanup patterns
func TestAWSResourceCleanupWrapper_ShouldWrapAWSCleanup(t *testing.T) {
	// Arrange
	cleanup := NewCleanupManager()
	resourceDeleted := false
	
	awsCleanup := func() error {
		resourceDeleted = true
		return nil
	}
	
	// Act
	wrappedCleanup := AWSResourceCleanupWrapper("test-resource", "aws-service", awsCleanup)
	cleanup.Register("aws-test", wrappedCleanup)
	err := cleanup.ExecuteAll()
	
	// Assert
	assert.NoError(t, err, "Wrapped AWS cleanup should succeed")
	assert.True(t, resourceDeleted, "AWS resource should be deleted")
}

// RED: Test that AWSResourceCleanupWrapper handles cleanup failures gracefully
func TestAWSResourceCleanupWrapper_ShouldHandleFailuresGracefully(t *testing.T) {
	// Arrange
	cleanup := NewCleanupManager()
	
	awsCleanup := func() error {
		return errors.New("resource deletion failed")
	}
	
	// Act
	wrappedCleanup := AWSResourceCleanupWrapper("test-resource", "aws-service", awsCleanup)
	cleanup.Register("aws-test", wrappedCleanup)
	err := cleanup.ExecuteAll()
	
	// Assert
	assert.Error(t, err, "Wrapped AWS cleanup should return error")
	assert.Contains(t, err.Error(), "resource deletion failed", "Error should contain original failure message")
}